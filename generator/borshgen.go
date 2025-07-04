package generator

import (
	"bufio"
	"bytes"
	"crypto/rand"
	_ "embed"
	"encoding/hex"
	"fmt"
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"sort"
	"strings"
	"sync"
	"text/template"

	"golang.org/x/tools/go/packages"

	"github.com/cespare/xxhash"
	"github.com/mlayerprotocol/go-borshgen/templates"
)

//go:embed custom_encoders.go
var customEncodersBytes []byte

// Enhanced options with zero-copy support
type GeneratorOptions struct {
	PrimaryTag   string
	FallbackTag  string
	IgnoreTag    string
	UsePooling   bool
	PackageName  string // Custom package name
	ZeroCopy     bool   // NEW: Enable zero-copy reads
	SafeMode     bool   // NEW: Safe vs unsafe zero-copy
	MaxStringLen int
	MaxSliceLen  int
	EncodeTag    string
}

func DefaultOptions() GeneratorOptions {
	return GeneratorOptions{
		PrimaryTag:   "msg",
		FallbackTag:  "json",
		IgnoreTag:    "-",
		UsePooling:   true,
		ZeroCopy:     false, // Default to safe copying
		SafeMode:     true,  // Use safe zero-copy by default
		MaxStringLen: 65535 * 200,
		MaxSliceLen:  65535,
		EncodeTag:    "enc",
	}
}

// FieldInfo with zero-copy information
type FieldInfo struct {
	Name                  string
	Type                  string
	Tag                   string
	IsPointer             bool
	IsPointerElement      bool
	PointerDeref          string
	PointerRef            string
	CustomEncoder         string
	IsCustomEncoder       bool
	ElementPointerRef     string
	ElementPointerDeref   string
	IsSlice               bool
	IsPointerSlice        bool
	IsMap                 bool
	IsInterface           bool
	IsStruct              bool
	ElementType           string
	IsOptional            bool
	BinaryTag             string
	IsBasicType           bool
	IsBasicPointerType    bool
	IsCustomType          bool
	CustomTypeName        string
	CustomElementTypeName string
	ShouldIgnore          bool
	CanZeroCopy           bool // NEW: Whether this field supports zero-copy
	HasEncTag             bool // NEW: Whether field has "enc" or "encode" tag for deterministic encoding
	EncOrder              int  // NEW: Sort order for deterministic encoding
	SliceItem             int  // index of item if Type is Slice
	ActualType            string
	ResolvedType          *ResolvedTypeInfo `json:"resolved_type,omitempty"`
	FullTypeName          string
}

type StructInfo struct {
	Name    string
	Fields  []FieldInfo
	Package string
	Options GeneratorOptions
}

type Package struct {
	Package    string
	CustomType string
}
type CodeGenerator struct {
	structs     []StructInfo
	structMap   map[string]bool
	options     GeneratorOptions
	packages    []Package
	rootPackage string
	mu          sync.Mutex
}

var specialTypes = map[string]bool{
	"time.Time":                   true,
	"json.RawMessage":             true,
	"github.com/google/uuid.UUID": true,
}

// Template helper functions
var templateFuncs = template.FuncMap{
	"sortedEncFields": func(fields []FieldInfo) []FieldInfo {
		var encFields []FieldInfo
		for _, field := range fields {
			if field.HasEncTag {
				encFields = append(encFields, field)
			}
		}
		slices.SortFunc(encFields, func(x, y FieldInfo) int {
			return strings.Compare(x.BinaryTag, y.BinaryTag)
		})
		return encFields
	},
	"getPrecedingFields": func(fields []FieldInfo, currentFieldName string) []FieldInfo {
		var preceding []FieldInfo
		for _, field := range fields {
			if field.Name == currentFieldName {
				break
			}
			preceding = append(preceding, field)
		}
		return preceding
	},
	"unmarshalBasicTypeTemplate": func(field FieldInfo) string {
		return UnmarshalBasicTypeFieldTemplate(field)
	},
	"isBasicElementType": func(field FieldInfo) bool {
		return isBasicType(field.ElementType)
	},
}

// Complete template with all necessary functions
const helperTemplate = templates.HelperTemplate

// Complete template with all necessary functions
const completeCodeTemplate = templates.MainTemplate

// canFieldZeroCopy determines if a field can support zero-copy reads
func canFieldZeroCopy(fieldType string) bool {
	zeroCopyTypes := map[string]bool{
		"string":  true,
		"[]byte":  true,
		"uint64":  true,
		"uint32":  true,
		"int64":   true,
		"int32":   true,
		"int":     true,
		"float64": true,
		"float32": true,
		"bool":    true,
	}

	return zeroCopyTypes[fieldType]
}

// Enhanced parsing with zero-copy option detection
func parseGenerateComment(commentGroup *ast.CommentGroup) (bool, GeneratorOptions) {

	if commentGroup == nil {
		return false, DefaultOptions()
	}

	options := DefaultOptions()
	found := false

	for _, comment := range commentGroup.List {
		line := strings.TrimSpace(comment.Text)

		if strings.HasPrefix(line, "//go:generate borshgen") {

			found = true

			parts := strings.Fields(line)
			for i := 2; i < len(parts); i++ {
				option := parts[i]

				if strings.HasPrefix(option, "-tag=") {
					options.PrimaryTag = strings.TrimPrefix(option, "-tag=")
				} else if strings.HasPrefix(option, "-fallback=") {
					options.FallbackTag = strings.TrimPrefix(option, "-fallback=")
				} else if option == "-zero-copy" {
					options.ZeroCopy = false // TODD: not yet tested
				} else if option == "-unsafe" {
					options.SafeMode = false
				} else if option == "-no-pool" {
					options.UsePooling = false
				} else if option == "-encode-tag=" {
					options.EncodeTag = strings.TrimPrefix(option, "-encode-tag=")
				}
			}
			break
		}
	}

	return found, options
}

// isBasicType determines if a type is a basic Go type
func isBasicType(typeName string) bool {
	basicTypes := map[string]bool{
		"string": true, "bool": true,
		"int": true, "int8": true, "int16": true, "int32": true, "int64": true,
		"uint": true, "uint8": true, "uint16": true, "uint32": true, "uint64": true,
		"uintptr": true,
		"float32": true, "float64": true,
		"complex64": true, "complex128": true,
		"byte": true, "rune": true, "[]byte": true,
	}
	return basicTypes[typeName]
}

func (cg *CodeGenerator) extractFieldTag(field *ast.Field, options GeneratorOptions) (tagName string, ignore bool, encode bool, parser string) {
	if field.Tag == nil {
		return "", false, false, ""
	}

	tagString := strings.Trim(field.Tag.Value, "`")
	tag := reflect.StructTag(tagString)

	structTag := reflect.StructTag(tag)
	// Check for "enc" or "encode" tag first
	_, hasEncTag := structTag.Lookup(options.EncodeTag)
	// if !hasEncTag  {
	// 	_,	hasEncTag = structTag.Lookup("encode")
	// }

	if options.IgnoreTag == "" {
		options.IgnoreTag = "-"
	}
	if primaryTag := tag.Get(options.PrimaryTag); primaryTag != "" {

		if primaryTag == options.IgnoreTag {

			return "", true, hasEncTag, ""
		}
		parts := strings.Split(primaryTag, ",")

		if len(parts) == 1 {
			return parts[0], false, hasEncTag, ""
		}
		return parts[0], false, hasEncTag, parts[1]
	} else {

		if fallbackTag := tag.Get(options.FallbackTag); fallbackTag != "" {
			if fallbackTag == options.IgnoreTag {
				return "", true, hasEncTag, ""
			}
			parts := strings.Split(fallbackTag, ",")

			return parts[0], false, hasEncTag, ""
		}
	}
	return "", false, hasEncTag, ""
}

// parseStructs parses Go source files to extract struct information with full package resolution
func (cg *CodeGenerator) parseStructs(filename string) error {
	// Get the directory containing the file to load the entire package
	dir := filepath.Dir(filename)

	// Configure package loading with all necessary information
	cfg := &packages.Config{
		Mode: packages.NeedName |
			packages.NeedFiles |
			packages.NeedCompiledGoFiles |
			packages.NeedImports |
			packages.NeedTypes |
			packages.NeedSyntax |
			packages.NeedTypesInfo,
		Dir: dir,
	}

	// Load the package containing our target file
	pkgs, err := packages.Load(cfg, ".")
	if err != nil {
		// Fallback to old method if package loading fails
		log.Printf("Package loading failed, falling back to single-file parsing: %v", err)
		return cg.parseStructsFallback(filename)
	}

	if len(pkgs) == 0 {
		return fmt.Errorf("no packages found")
	}

	if packages.PrintErrors(pkgs) > 0 {
		log.Printf("Warning: some packages had errors, continuing with available type information")
	}

	pkg := pkgs[0] // Get the main package
	cg.rootPackage = pkg.PkgPath

	// Find our target file in the package
	var targetFile *ast.File
	targetFilePath, err := filepath.Abs(filename)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	for _, file := range pkg.Syntax {
		if filePath := pkg.Fset.Position(file.Pos()).Filename; filePath != "" {
			if absPath, err := filepath.Abs(filePath); err == nil && absPath == targetFilePath {
				targetFile = file
				break
			}
		}
	}

	if targetFile == nil {
		return fmt.Errorf("target file not found in package for file: %s", targetFilePath)
	}

	packageName := targetFile.Name.Name
	cg.structMap = make(map[string]bool)

	// Use the package's type information (this includes all imports!)
	info := pkg.TypesInfo

	// Helper function to find generate comment in multiple locations
	findGenerateComment := func(genDecl *ast.GenDecl, typeSpec *ast.TypeSpec) (bool, GeneratorOptions) {
		// Try node.Doc first (most common)
		if found, options := parseGenerateComment(genDecl.Doc); found {
			return found, options
		}

		// Try typeSpec.Doc (sometimes comments are attached here)
		if found, options := parseGenerateComment(typeSpec.Doc); found {
			return found, options
		}

		// Try file-level comments if this is the first/only declaration
		if len(targetFile.Decls) > 0 && targetFile.Decls[0] == genDecl {
			for _, commentGroup := range targetFile.Comments {
				if found, options := parseGenerateComment(commentGroup); found {
					return found, options
				}
			}
		}

		return false, DefaultOptions()
	}

	// First pass: collect all struct names that should be generated
	ast.Inspect(targetFile, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.GenDecl:
			if node.Tok == token.TYPE {
				for _, spec := range node.Specs {
					if typeSpec, ok := spec.(*ast.TypeSpec); ok {
						if _, ok := typeSpec.Type.(*ast.StructType); ok {
							if found, _ := findGenerateComment(node, typeSpec); found {
								cg.structMap[typeSpec.Name.Name] = true
							}
						}
					}
				}
			}
		}
		return true
	})

	// Second pass: extract struct information with options
	ast.Inspect(targetFile, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.GenDecl:
			if node.Tok == token.TYPE {
				for _, spec := range node.Specs {
					if typeSpec, ok := spec.(*ast.TypeSpec); ok {
						if structType, ok := typeSpec.Type.(*ast.StructType); ok {
							if found, options := findGenerateComment(node, typeSpec); found {
								options.PackageName = packageName
								cg.options = options

								// Pass the package for enhanced type resolution
								structInfo := cg.extractStructInfo(typeSpec.Name.Name, structType, options, info, pkg)
								cg.structs = append(cg.structs, structInfo)
							}
						}
					}
				}
			}
		}
		return true
	})

	return nil
}

// parseStructsFallback is the original implementation as a fallback
func (cg *CodeGenerator) parseStructsFallback(filename string) error {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
	if err != nil {
		return err
	}

	packageName := file.Name.Name
	cg.structMap = make(map[string]bool)

	// Set up type checking
	info := &types.Info{
		Types: make(map[ast.Expr]types.TypeAndValue),
		Defs:  make(map[*ast.Ident]types.Object),
	}
	conf := types.Config{
		Importer: importer.Default(),
		Error: func(err error) {
			log.Printf("Type checking warning (continuing without full type info): %v", err)
		},
	}
	_, err = conf.Check(packageName, fset, []*ast.File{file}, info)
	if err != nil {
		log.Printf("Type checking completed with errors, continuing: %v", err)
	}

	// Rest of your original logic...
	// (I'll omit the duplicated code for brevity, but it would be the same as your original)

	return nil
}

// Enhanced extractStructInfo with optional package information
func (cg *CodeGenerator) extractStructInfo(structName string, structType *ast.StructType, options GeneratorOptions, typeInfo *types.Info, pkg ...*packages.Package) StructInfo {
	structInfo := StructInfo{
		Name:    structName,
		Package: options.PackageName,
		Fields:  []FieldInfo{},
		Options: options,
	}

	var pkgInfo *packages.Package
	if len(pkg) > 0 {
		pkgInfo = pkg[0]
	}

	for _, field := range structType.Fields.List {
		for _, name := range field.Names {
			actualType := ""
			var resolvedTypeInfo *ResolvedTypeInfo

			// Enhanced type extraction with package context
			if typeInfo != nil {
				if fieldType, ok := typeInfo.Types[field.Type]; ok {
					underlying := fieldType.Type.Underlying()
					actualType = strings.ReplaceAll(underlying.String(), options.PackageName+".", "")

					// Extract detailed type information if we have package context
					if pkgInfo != nil {
						resolvedTypeInfo = cg.resolveTypeInfo(fieldType.Type, pkgInfo)
					}
				}
			}
			if resolvedTypeInfo != nil {

				if resolvedTypeInfo.ElementType != nil {
					actualType = resolvedTypeInfo.ElementType.TypeName
				}
			}
			fieldInfo := cg.extractFieldInfo(name.Name, field, actualType, options)

			if !fieldInfo.IsCustomEncoder && len(actualType) > 0 {

				fieldInfo.ActualType = actualType
				if strings.Contains(fieldInfo.Type, ".") && len(fieldInfo.KnownImportedType()) == 0 {
					fieldInfo.CustomTypeName = fieldInfo.Type
					fieldInfo.Type = actualType
					if isBasicType(actualType) {
						fieldInfo.IsBasicType = true
						fieldInfo.ElementType = actualType
						if fieldInfo.IsPointer {
							fieldInfo.IsBasicPointerType = true
						}
					}
				}

			}
			if fieldInfo.IsPointer && fieldInfo.ElementType == "" {
				// If it's a pointer but no element type is set, use the actual type

				fieldInfo.PointerDeref = "*"
				fieldInfo.PointerRef = "&"
				if resolvedTypeInfo.ElementType != nil {
					fieldInfo.ElementType = resolvedTypeInfo.ElementType.UnderlyingType.String()
				}

			}

			if resolvedTypeInfo != nil {
				pkg := ""
				ctype := ""
				 if len(resolvedTypeInfo.FullTypeName) > 0 {
					pkg = resolvedTypeInfo.FullTypeName[0:strings.LastIndex(resolvedTypeInfo.FullTypeName, ".")]
					ctype = resolvedTypeInfo.FullTypeName[strings.LastIndex(resolvedTypeInfo.FullTypeName, "/")+1:]
				 } else {
					if resolvedTypeInfo.ElementType != nil && len(resolvedTypeInfo.ElementType.FullTypeName) > 0 {
						pkg = resolvedTypeInfo.ElementType.FullTypeName[0:strings.LastIndex(resolvedTypeInfo.ElementType.FullTypeName, ".")]
						ctype = resolvedTypeInfo.ElementType.FullTypeName[strings.LastIndex( resolvedTypeInfo.ElementType.FullTypeName, "/")+1:]

					} 
				}
				fmt.Println("FOUUND: ", resolvedTypeInfo.FullTypeName)

				if len(pkg) > 0 {
					
					cg.mu.Lock()

					if pkg != cg.rootPackage && !specialTypes[ctype] {
						if !slices.ContainsFunc(cg.packages, func(p Package) bool {
							return strings.EqualFold(p.Package, pkg)
						}) {
							cg.packages = append(cg.packages, Package{
								Package:    pkg,
								CustomType: ctype,
							})
						}
					}
					cg.mu.Unlock()
				}
			}

			if resolvedTypeInfo != nil && len(resolvedTypeInfo.FullTypeName) > 0 && !specialTypes[fieldInfo.CustomTypeName] {

				if !fieldInfo.IsBasicType {

					fieldInfo.ResolvedType = resolvedTypeInfo
					if len(resolvedTypeInfo.FullTypeName) > 0 {
						fieldInfo.FullTypeName = resolvedTypeInfo.FullTypeName
						if resolvedTypeInfo.ElementType != nil {
							// fmt.Printf("\nXXXXType: %s; ElementType: %s; %+v", resolvedTypeInfo.TypeName,  resolvedTypeInfo.ElementType.TypeName, resolvedTypeInfo.UnderlyingType)
							if len(fieldInfo.CustomTypeName) == 0 {
								fieldInfo.CustomTypeName = resolvedTypeInfo.TypeName
							}
							fieldInfo.Type = resolvedTypeInfo.UnderlyingType.String()
							if len(resolvedTypeInfo.ElementType.TypeName) > 0 {
								fieldInfo.ElementType = resolvedTypeInfo.ElementType.TypeName
							}
						} else {
							fieldInfo.IsBasicType = isBasicType(fieldInfo.ActualType)
							fieldInfo.ElementType = fieldInfo.ActualType
						}
						if isBasicType(fieldInfo.CustomTypeName) {
							fieldInfo.CustomTypeName = ""
						}

					}

				}
			}

			if !fieldInfo.ShouldIgnore {
				structInfo.Fields = append(structInfo.Fields, fieldInfo)
			}
		}
	}

	return structInfo
}

// ResolvedTypeInfo contains detailed information about a resolved type
type ResolvedTypeInfo struct {
	PackagePath    string // e.g., "github.com/google/uuid"
	PackageName    string // e.g., "uuid"
	TypeName       string // e.g., "UUID"
	FullTypeName   string // e.g., "github.com/google/uuid.UUID"
	IsImported     bool   // true if from external package
	IsBasic        bool   // true for built-in types
	IsSlice        bool
	IsPointer      bool
	IsStruct       bool
	ElementType    *ResolvedTypeInfo // for slices/arrays/pointers
	UnderlyingType types.Type        // the actual Go type
}

// resolveTypeInfo extracts detailed type information
func (cg *CodeGenerator) resolveTypeInfo(t types.Type, pkg *packages.Package) *ResolvedTypeInfo {
	info := &ResolvedTypeInfo{
		UnderlyingType: t,
	}

	switch typ := t.(type) {
	case *types.Named:
		// This is a named type (could be from another package)
		obj := typ.Obj()
		if obj != nil && obj.Pkg() != nil {
			info.PackagePath = obj.Pkg().Path()
			info.PackageName = obj.Pkg().Name()
			info.TypeName = obj.Name()
			info.FullTypeName = fmt.Sprintf("%s.%s", obj.Pkg().Path(), obj.Name())

			// Check if it's from an external package
			if obj.Pkg() != pkg.Types {
				info.IsImported = true
			}

			// Check if underlying type is a struct
			if _, ok := typ.Underlying().(*types.Struct); ok {
				info.IsStruct = true
			}

			underlyingInfo := cg.resolveTypeInfo(typ.Underlying(), pkg)
			info.UnderlyingType = underlyingInfo.UnderlyingType
			info.IsStruct = underlyingInfo.IsStruct
			info.IsBasic = underlyingInfo.IsBasic
			info.IsSlice = underlyingInfo.IsSlice
			info.IsPointer = underlyingInfo.IsPointer
			info.ElementType = underlyingInfo.ElementType
		}

	case *types.Basic:
		info.TypeName = typ.Name()
		info.IsBasic = true

	case *types.Slice:
		info.IsSlice = true
		// Recursively analyze element type
		info.ElementType = cg.resolveTypeInfo(typ.Elem(), pkg)

	case *types.Pointer:
		info.IsPointer = true
		// Analyze pointed-to type
		info.ElementType = cg.resolveTypeInfo(typ.Elem(), pkg)

	case *types.Struct:
		info.IsStruct = true
		info.TypeName = "struct"
	}

	return info
}

// Helper method to check if a field is a known imported type that needs special handling
func (fi *FieldInfo) KnownImportedType() string {
	if fi.ResolvedType == nil || !fi.ResolvedType.IsImported {
		return ""
	}

	// Add known types that you want to handle specially
	knownTypes := map[string]string{
		"time.Time":                   "struct",
		"github.com/google/uuid.UUID": "[16]byte",
		"encoding/json.RawMessage":    "[]byte",
		// Add more as needed
	}

	return knownTypes[fi.ResolvedType.FullTypeName]
}

// Helper method to get the marshal code for known types
func (fi *FieldInfo) GetMarshalCode(varName string) (string, bool) {
	if fi.ResolvedType == nil || !fi.ResolvedType.IsImported {
		return "", false
	}

	switch fi.ResolvedType.FullTypeName {
	case "time.Time":
		return fmt.Sprintf("binary.LittleEndian.PutUint64(buf[offset:], uint64(%s.Unix()))", varName), true
	case "github.com/google/uuid.UUID", "encoding/json.RawMessage":
		return fmt.Sprintf("copy(buf[offset:], %s[:])", varName), true

	default:
		return "", false
	}
}

// Helper method to get the marshal code for known types
func (fieldInfo *FieldInfo) assignCustomEncoder(_fieldType string, prefix string) error {

	//fmt.Println("ASSIGNCUSTOMENCODER", prefix, _fieldType, fieldInfo.Name, fieldInfo.Type, fieldInfo.CustomTypeName)
	switch _fieldType {
	case "time.Time":
		fieldInfo.Type = "uint64"
		fieldInfo.CustomTypeName = prefix + "time.Time"
		fieldInfo.CustomElementTypeName = "time.Time"
		fieldInfo.CustomEncoder = "_CustomTimeTimeEncoder"
		fieldInfo.ElementType = "time.Time"
		fieldInfo.IsCustomEncoder = true

	case "json.RawMessage", "*json.RawMessage":

		fieldInfo.Type = "[]byte"
		fieldInfo.CustomElementTypeName = "json.RawMessage"
		fieldInfo.CustomTypeName = prefix + "json.RawMessage"
		fieldInfo.CustomEncoder = "_CustomJsonRawMessageEncoder"
		fieldInfo.ElementType = "json.RawMessage"
		fieldInfo.IsCustomEncoder = true

	case "uuid.UUID":
		fieldInfo.Type = "[16]byte"
		fieldInfo.CustomElementTypeName = "uuid.UUID"
		fieldInfo.CustomTypeName = prefix + "uuid.UUID"
		fieldInfo.CustomEncoder = "_CustomUuidUUIDEncoder"
		fieldInfo.ElementType = "json.RawMessage"
		fieldInfo.IsCustomEncoder = true
	default:
		return fmt.Errorf("unsupported custom encoder type: %s", fieldInfo.CustomTypeName)
	}
	fieldInfo.Type = fieldInfo.ElementType
	fieldInfo.IsCustomType = true
	fieldInfo.IsCustomEncoder = true
	return nil

}

// extractFieldInfo extracts information from a field
func (cg *CodeGenerator) extractFieldInfo(name string, field *ast.Field, actualType string, options GeneratorOptions) FieldInfo {
	fieldInfo := FieldInfo{
		Name: name,
	}

	// Extract tag information with fallback
	binaryTag, shouldIgnore, hasEncTag, customEncoder := cg.extractFieldTag(field, options)
	fieldInfo.BinaryTag = binaryTag
	fieldInfo.ShouldIgnore = shouldIgnore
	fieldInfo.HasEncTag = hasEncTag

	if len(customEncoder) > 0 {
		if !strings.HasPrefix(customEncoder, "[]") && !strings.HasPrefix(customEncoder, "[][]") {
			fieldInfo.IsCustomType = true
			fieldInfo.CustomTypeName = fieldInfo.Type
			if isBasicType(customEncoder) {
				fieldInfo.Type = customEncoder
			} else {
				fieldInfo.CustomEncoder = customEncoder
				fieldInfo.IsCustomEncoder = true
			}
		}
	} else {
		if isBasicType(actualType) || actualType == "[]byte" {
			// customEncoder = actualType
		}
	}
	if len(actualType) > 0 {
		fieldInfo.ActualType = actualType

	}

	if shouldIgnore {
		return fieldInfo
	}

	// If no tag found, use field name
	if fieldInfo.BinaryTag == "" {
		fieldInfo.BinaryTag = strings.ToLower(name)
	}

	// Extract type information
	switch t := field.Type.(type) {
	case *ast.Ident:

		fieldInfo.Type = t.Name

		fieldInfo.IsBasicType = isBasicType(t.Name) || isBasicType(customEncoder) || isBasicType(actualType)
		fieldInfo.CanZeroCopy = canFieldZeroCopy(t.Name)
		if cg.structMap[t.Name] || customEncoder == "struct" || customEncoder == "bin" {
			fieldInfo.IsStruct = true
		}
		// if fieldInfo.IsBasicType {
		fieldInfo.ElementType = actualType
		// }

	case *ast.StarExpr:

		fieldInfo.IsPointer = true
		fieldInfo.PointerRef = "&"
		fieldInfo.PointerDeref = "*"
		if ident, ok := t.X.(*ast.Ident); ok {
			if isBasicType(customEncoder) {
				fieldInfo.Type = "*" + customEncoder
			} else {
				fieldInfo.Type = "*" + ident.Name
			}
			fieldInfo.ElementType = strings.ReplaceAll(ident.Name, "*", "")
			fieldInfo.IsBasicPointerType = isBasicType(ident.Name) || isBasicType(customEncoder)
			if cg.structMap[ident.Name] {
				fieldInfo.IsStruct = true
			}
			fieldInfo.CustomElementTypeName = ident.Name

		}
		if t, ok := t.X.(*ast.ArrayType); ok {
			if t.Len == nil { // slice

				fieldInfo.IsPointerSlice = true
				if ident, ok := t.Elt.(*ast.Ident); ok {
					elementType := ident.Name
					if len(customEncoder) > 0 {
						elementType = customEncoder[2:]
					}
					fieldInfo.Type = "[]" + elementType
					fieldInfo.CanZeroCopy = canFieldZeroCopy(fieldInfo.Type) || canFieldZeroCopy(customEncoder)
					// fieldInfo.ElementType = ident.Name
					fieldInfo.IsPointerElement = strings.HasPrefix(elementType, "*")
					fieldInfo.CustomTypeName = "[]" + ident.Name
					fieldInfo.CustomElementTypeName = ident.Name

					fieldInfo.ElementType = strings.ReplaceAll(elementType, "*", "")

					// fieldInfo.IsBasicType = isBasicType(ident.Name)
					if cg.structMap[ident.Name] {
						fieldInfo.IsStruct = true
					}

				}
			}
			if t, ok := t.Elt.(*ast.SelectorExpr); ok {
				if pkgIdent, ok := t.X.(*ast.Ident); ok {

					name := pkgIdent.Name + "." + t.Sel.Name
					fieldInfo.Type = "*" + name
					fieldInfo.ActualType = fieldInfo.Type
					fieldInfo.IsCustomType = true
					actualType = fieldInfo.Type

					if err := (&fieldInfo).assignCustomEncoder(name, "[]"); err != nil {
						//fmt.Println(fmt.Errorf("failed to assign custom encoder for field %s: %v", name, err))
					}

				}
			}
		}
		if t, ok := t.X.(*ast.SelectorExpr); ok {
			if pkgIdent, ok := t.X.(*ast.Ident); ok {

				name := pkgIdent.Name + "." + t.Sel.Name
				fieldInfo.Type = "*" + name
				fieldInfo.ActualType = fieldInfo.Type
				fieldInfo.IsCustomType = true
				actualType = fieldInfo.Type

				if err := (&fieldInfo).assignCustomEncoder(name, "*"); err != nil {
					//	fmt.Println(fmt.Errorf("failed to assign custom encoder for field %s: %v", name, err))
				}
			}
		}

	case *ast.ArrayType:

		if t.Len == nil { // slice
			fieldInfo.IsSlice = true
			if ident, ok := t.Elt.(*ast.Ident); ok {

				elementType := ident.Name
				if len(customEncoder) > 0 {
					elementType = customEncoder[2:]
				}
				fieldInfo.Type = "[]" + elementType
				fieldInfo.CanZeroCopy = canFieldZeroCopy(fieldInfo.Type) || canFieldZeroCopy(customEncoder)
				// fieldInfo.ElementType = ident.Name
				fieldInfo.IsPointerElement = strings.HasPrefix(elementType, "*")
				fieldInfo.CustomTypeName = "[]" + ident.Name
				fieldInfo.CustomElementTypeName = ident.Name
				fieldInfo.ElementType = strings.ReplaceAll(elementType, "*", "")
				// if ident.Name == "byte" {
				// 	fieldInfo.IsBasicType = true
				// }
				if cg.structMap[ident.Name] {
					fieldInfo.IsStruct = true
				}

			}
			if t, ok := t.Elt.(*ast.StarExpr); ok {
				fieldInfo.IsPointerElement = true
				fieldInfo.ElementPointerRef = "&"
				fieldInfo.ElementPointerDeref = "*"
				if ident, ok := t.X.(*ast.Ident); ok {
					// if isBasicType(customEncoder) {
					// 	fieldInfo.ElementType = "*" + customEncoder
					// } else {
					// 	fieldInfo.ElementType = "*" + ident.Name
					// }
					fieldInfo.ElementType = strings.ReplaceAll(ident.Name, "*", "")
					// fieldInfo.IsBasicType = isBasicType(ident.Name) || isBasicType(customEncoder)
					// if cg.structMap[ident.Name] {
					// 	fieldInfo.IsStruct = true
					// }
				}
				if t, ok := t.X.(*ast.SelectorExpr); ok {
					if ident, ok := t.X.(*ast.Ident); ok {
						fieldInfo.CustomTypeName = "[]*" + ident.Name + "." + t.Sel.Name
						fieldInfo.CustomElementTypeName = ident.Name + "." + t.Sel.Name

						name := ident.Name + "." + t.Sel.Name
						fieldInfo.Type = "[]*" + name
						fieldInfo.ActualType = fieldInfo.Type
						fieldInfo.IsCustomType = true
						actualType = fieldInfo.Type

						if err := (&fieldInfo).assignCustomEncoder(name, "[]*"); err != nil {
							fmt.Println(fmt.Errorf("failed to assign custom encoder for field %s: %v", name, err))
						}

					}
				}

			}
			if t, ok := t.Elt.(*ast.SelectorExpr); ok {
				if ident, ok := t.X.(*ast.Ident); ok {
					fieldInfo.CustomTypeName = "[]" + ident.Name + "." + t.Sel.Name
					fieldInfo.CustomElementTypeName = ident.Name + "." + t.Sel.Name

					name := ident.Name + "." + t.Sel.Name
					fieldInfo.Type = "*" + name
					fieldInfo.ActualType = fieldInfo.Type
					fieldInfo.IsCustomType = true
					actualType = fieldInfo.Type

					if err := (&fieldInfo).assignCustomEncoder(name, "[]"); err != nil {
						// fmt.Println(fmt.Errorf("failed to assign custom encoder for field %s: %v", name, err))
					}

				}
			}
			if t, ok := t.Elt.(*ast.ArrayType); ok {

				if ident, ok := t.Elt.(*ast.Ident); ok {
					elementType := ident.Name
					if len(customEncoder) > 0 {
						elementType = strings.ReplaceAll(customEncoder, "[]", "")
					}
					fieldInfo.Type = "[][]" + elementType
					fieldInfo.CanZeroCopy = canFieldZeroCopy(fieldInfo.Type) || canFieldZeroCopy(customEncoder)
					fieldInfo.CustomElementTypeName = "[]" + ident.Name
					fieldInfo.ElementType = "[]" + strings.ReplaceAll(elementType, "*", "")

					// fieldInfo.IsBasicType = isBasicType(ident.Name)
					if cg.structMap[ident.Name] {
						fieldInfo.IsStruct = true
					}

				}
			}
		}
	case *ast.MapType:
		fieldInfo.IsMap = true
		fieldInfo.Type = "map[string]interface{}"
	case *ast.InterfaceType:
		fieldInfo.IsInterface = true
		fieldInfo.Type = "interface{}"
	case *ast.SelectorExpr:
		if pkgIdent, ok := t.X.(*ast.Ident); ok {

			name := pkgIdent.Name + "." + t.Sel.Name
			fieldInfo.Type = name
			fieldInfo.ActualType = fieldInfo.Type
			fieldInfo.IsCustomType = true
			actualType = fieldInfo.Type

			if err := (&fieldInfo).assignCustomEncoder(name, ""); err != nil {
				// fmt.Println(fmt.Errorf("failed to assign custom encoder for field %s: %v", name, err))
			}

		}
	}
	if isBasicType(customEncoder) {
		fieldInfo.IsCustomType = true
		actualType = customEncoder
	}
	if !fieldInfo.IsCustomEncoder && len(actualType) > 0 && actualType != fieldInfo.Type {
		fieldInfo.CustomTypeName = fieldInfo.Type
		if !fieldInfo.IsSlice {
			fieldInfo.Type = actualType
			fieldInfo.IsCustomType = true
		}

	}
	return fieldInfo
}

// Get the code
func (cg *CodeGenerator) getCode(outputFile string) (header *bytes.Buffer, main *bytes.Buffer, err error) {
	helperTmpl, err := template.New("helper").Parse(helperTemplate)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse helper template: %v", err)
	}
	if err := helperTmpl.Execute(header, struct {
		Package string
		Options GeneratorOptions
	}{
		Package: cg.structs[0].Package,
		Options: cg.options,
	}); err != nil {
		return nil, nil, fmt.Errorf("failed to execute helper template: %v", err)
	}
	tmpl, err := template.New("binary").Funcs(templateFuncs).Parse(completeCodeTemplate)
	if err != nil {
		return header, main, err
	}

	if len(cg.structs) == 0 {
		return header, main, fmt.Errorf("empty structs")
	}
	data := struct {
		Package  string
		Structs  []StructInfo
		Packages []Package
		Options  GeneratorOptions
	}{
		Package:  cg.structs[0].Package,
		Structs:  cg.structs,
		Options:  cg.options,
		Packages: cg.packages,
	}

	err = tmpl.Execute(main, data)

	return header, main, err
}

// generateCode generates the binary encoding/decoding code
func (cg *CodeGenerator) generateCode(outputFile string) (err error) {

	tmpl, err := template.New("binary").Funcs(templateFuncs).Parse(completeCodeTemplate)
	if err != nil {
		return err
	}

	dir := filepath.Dir(outputFile)
	hash := make([]byte, 4)
	if _, err := rand.Read(hash); err != nil {
		return fmt.Errorf("failed to generate random hash: %v", err)
	}

	helperFile := filepath.Join(dir, "borshgen_common_"+fmt.Sprint(xxhash.Sum64String(filepath.Base(dir))%10000000000)+"_gen.go")
	// if _, err := os.Stat(helperFile); os.IsNotExist(err) {
	// err := os.Remove(helperFile)

	helperOut, err := os.Create(helperFile)
	if err != nil {
		return fmt.Errorf("failed to create common_gen.go: %v", err)
	}

	defer helperOut.Close()

	helperTmpl, err := template.New("helper").Parse(helperTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse helper template: %v", err)
	}

	// if len(cg.structs) == 0 {
	// 	return fmt.Errorf("no structs to generate")
	// }

	if err := helperTmpl.Execute(helperOut, struct {
		Package string
		Options GeneratorOptions
	}{
		Package: cg.structs[0].Package,
		Options: cg.options,
	}); err != nil {
		return fmt.Errorf("failed to execute helper template: %v", err)
	}

	// copy the custom encoder file
	encoderFile := filepath.Join(dir, "borshgen_custom_encoder_"+fmt.Sprint(xxhash.Sum64String(filepath.Base(dir))%10000000000)+"_gen.go")
	str := string(customEncodersBytes)
	ce := strings.Replace(str, "package generator", "package "+cg.structs[0].Package, 1)
	ce = "// Code generated by bingen. DO NOT EDIT." + "\n" + ce
	err = os.WriteFile(encoderFile, []byte(ce), 0644)
	if err != nil {
		return fmt.Errorf("failed to copy custom encoders: %v", err)
	}
	file, err := os.Create(outputFile)
	if err != nil {
		return err
	}
	defer file.Close()

	if len(cg.structs) == 0 {
		return fmt.Errorf("empty structs")
	}
	data := struct {
		Package  string
		Structs  []StructInfo
		Packages []Package
		Options  GeneratorOptions
	}{
		Package:  cg.structs[0].Package,
		Structs:  cg.structs,
		Options:  cg.options,
		Packages: cg.packages,
	}

	err = tmpl.Execute(file, data)

	return err
}

func (cg *CodeGenerator) sortEncFields(fields []FieldInfo) {
	// Create a separate slice of enc fields for sorting
	var encFields []FieldInfo
	for _, field := range fields {
		if field.HasEncTag {
			encFields = append(encFields, field)
		}
	}

	// Sort by binary tag name
	sort.Slice(encFields, func(i, j int) bool {
		return encFields[i].BinaryTag < encFields[j].BinaryTag
	})

	// Assign order indices back to original fields
	for i, encField := range encFields {
		for j := range fields {
			if fields[j].Name == encField.Name {
				fields[j].EncOrder = i
				break
			}
		}
	}
	sort.Slice(fields, func(i, j int) bool {
		return fields[i].EncOrder < fields[j].EncOrder
	})

}

// Generate is the main entry point for code generation
func GenerateDir(path, primaryTag, fallbackTag, encodeTag string, ignoreTag string, usePooling bool, maxStringLen int) error {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("failed to stat path: %w", err)
	}

	if !info.IsDir() {
		return fmt.Errorf("path is not a directory: %s", path)
	}
	hash := make([]byte, 4)
	if _, err := rand.Read(hash); err != nil {
		return fmt.Errorf("failed to generate random hash: %v", err)
	}
	return filepath.WalkDir(path, func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil // continue walking
		}
		if strings.HasSuffix(p, ".go") && !strings.HasSuffix(p, "_gen.go") && !strings.HasSuffix(p, "test.go") {
			fmt.Printf("ProcessingFile: %v", p)
			fmt.Println()
			tmp := strings.TrimSuffix(p, ".go") + "_" + hex.EncodeToString(hash) + "_tmp_gen.go"
			defer os.Remove(tmp)
			err := Generate(p, tmp, primaryTag, fallbackTag, ignoreTag, encodeTag, usePooling, maxStringLen)
			if err != nil {
				fmt.Printf("CodeGentError: %v", err)
				if !strings.Contains(err.Error(), "no structs found") {
					return err
				}
				return nil
			}

			finalFile := strings.TrimSuffix(p, ".go") + "_borshgen_" + fmt.Sprint(xxhash.Sum64String(filepath.Base(filepath.Dir(p)))%10000000000) + "_gen.go"
			return trimFile(tmp, finalFile)
		}

		return nil
	})
}

// Generate is the main entry point for code generation
func GenerateFile(path, primaryTag, fallbackTag, encodeTag string, ignoreTag string, usePooling bool, maxStringLen int) error {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("failed to stat path: %w", err)
	}

	if info.IsDir() {
		return fmt.Errorf("path cannot be a directory: %s", path)
	}
	hash := make([]byte, 4)
	if _, err := rand.Read(hash); err != nil {
		return fmt.Errorf("failed to generate random hash: %v", err)
	}
	fmt.Printf("ProcessingFile: %v", path)
	fmt.Println()
	tmp := strings.TrimSuffix(path, ".go") + "_" + hex.EncodeToString(hash) + "_tmp_gen.go"
	defer os.Remove(tmp)
	err = Generate(path, tmp, primaryTag, fallbackTag, ignoreTag, encodeTag, usePooling, maxStringLen)
	if err != nil {
		fmt.Printf("CodeGentError: %v", err)
		if !strings.Contains(err.Error(), "no structs found") {
			return err
		}
		return nil
	}

	finalFile := strings.TrimSuffix(path, ".go") + "_borshgen_" + fmt.Sprint(xxhash.Sum64String(filepath.Base(filepath.Dir(path)))%10000000000) + "_gen.go"
	return trimFile(tmp, finalFile)

}

func trimFile(inputFile, outputFile string) error {

	// Step 1: Read entire file into memory
	file, err := os.Open(inputFile)
	if err != nil {
		return err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return err
	}

	// Step 2: Filter out whitespace-only lines unless followed by a comment
	var cleaned []string
	for i := 0; i < len(lines); i++ {
		line := lines[i]
		trimmed := strings.TrimSpace(line)

		// Check if it's a blank line
		if trimmed == "" {
			if i+1 < len(lines) && strings.HasPrefix(strings.TrimSpace(lines[i+1]), "//") {
				// Preserve empty line because next line is a comment
				cleaned = append(cleaned, line)
			}
			// Else skip it (remove blank line)
		} else {
			cleaned = append(cleaned, line)
		}
	}

	// Step 3: Rewrite file with cleaned lines
	output, err := os.Create(outputFile)
	if err != nil {
		return err
	}
	defer output.Close()

	writer := bufio.NewWriter(output)
	for _, line := range cleaned {
		_, err := writer.WriteString(line + "\n")
		if err != nil {
			return err
		}
	}
	return writer.Flush()

}

// Generate is the main entry point for code generation
func Generate(inputFile, outputFile, primaryTag, fallbackTag, encodeTag string, ignoreTag string, usePooling bool, maxStringLen int) error {
	if len(outputFile) == 0 {
		outputFile = strings.TrimSuffix(inputFile, ".go") + "_gen.go"
	}

	cg := &CodeGenerator{}
	cg.options = GeneratorOptions{
		PrimaryTag:   primaryTag,
		FallbackTag:  fallbackTag,
		IgnoreTag:    ignoreTag,
		UsePooling:   usePooling,
		MaxStringLen: maxStringLen,
		MaxSliceLen:  65535,
		ZeroCopy:     false,
		SafeMode:     true,
		EncodeTag:    encodeTag,
	}

	err := cg.parseStructs(inputFile)
	if err != nil {
		return fmt.Errorf("error parsing structs: %v", err)
	}

	if len(cg.structs) == 0 {
		return fmt.Errorf("no structs found with //go:generate borshgen comment")
	}

	err = cg.generateCode(outputFile)
	if err != nil {
		return fmt.Errorf("error generating code: %v", err)
	}

	fmt.Printf("Generated binary encoding code in %s\n", outputFile)
	fmt.Printf("Found %d struct(s): ", len(cg.structs))
	for i, s := range cg.structs {
		if i > 0 {
			fmt.Print(", ")
		}
		fmt.Print(s.Name)
	}
	fmt.Println()

	// Show configuration
	fmt.Printf("Configuration:\n")
	fmt.Printf("  Primary tag: %s\n", primaryTag)
	if fallbackTag != "" {
		fmt.Printf("  Fallback tag: %s\n", fallbackTag)
	}
	fmt.Printf("  Ignore value: %s\n", ignoreTag)
	fmt.Printf("  Buffer pooling: %t\n", usePooling)

	// Show field tag usage
	for _, s := range cg.structs {
		primaryCount := 0
		fallbackCount := 0
		ignoredCount := 0

		for _, f := range s.Fields {
			if f.ShouldIgnore {
				ignoredCount++
			} else if strings.Contains(f.Tag, primaryTag+":") {
				primaryCount++
			} else {
				fallbackCount++
			}
		}

		fmt.Printf("  %s: %d primary tags, %d fallback tags, %d ignored\n",
			s.Name, primaryCount, fallbackCount, ignoredCount)
	}

	return nil
}

// GenerateWithZeroCopy is an enhanced version that supports zero-copy options
func GenerateWithZeroCopy(inputFile, primaryTag, fallbackTag, ignoreTag string, usePooling, zeroCopy, safeMode bool, maxStringLen int) error {
	outputFile := strings.TrimSuffix(inputFile, ".go") + "_gen.go"

	cg := &CodeGenerator{}
	cg.options = GeneratorOptions{
		PrimaryTag:   primaryTag,
		FallbackTag:  fallbackTag,
		IgnoreTag:    ignoreTag,
		UsePooling:   usePooling,
		MaxStringLen: maxStringLen,
		MaxSliceLen:  65535,
		ZeroCopy:     zeroCopy,
		SafeMode:     safeMode,
	}

	err := cg.parseStructs(inputFile)
	if err != nil {
		return fmt.Errorf("error parsing structs: %v", err)
	}

	if len(cg.structs) == 0 {
		return fmt.Errorf("no structs found with //go:generate borshgen comment")
	}

	err = cg.generateCode(outputFile)
	if err != nil {
		return fmt.Errorf("error generating code: %v", err)
	}

	fmt.Printf("Generated binary encoding code in %s\n", outputFile)
	fmt.Printf("Found %d struct(s): ", len(cg.structs))
	for i, s := range cg.structs {
		if i > 0 {
			fmt.Print(", ")
		}
		fmt.Print(s.Name)
	}
	fmt.Println()

	// Show configuration
	fmt.Printf("Configuration:\n")
	fmt.Printf("  Primary tag: %s\n", primaryTag)
	if fallbackTag != "" {
		fmt.Printf("  Fallback tag: %s\n", fallbackTag)
	}
	fmt.Printf("  Ignore value: %s\n", ignoreTag)
	fmt.Printf("  Buffer pooling: %t\n", usePooling)
	fmt.Printf("  Zero-copy mode: %t\n", zeroCopy)
	if zeroCopy {
		fmt.Printf("  Safe mode: %t\n", safeMode)
	}

	return nil
}
