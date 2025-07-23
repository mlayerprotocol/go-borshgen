package templates

// Complete template with all necessary functions
const MainTemplate = `

package {{.Package}}

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"sync"
	{{if and .Options.ZeroCopy (not .Options.SafeMode)}}"unsafe"{{end}}
	{{ range .Packages }}"{{ .Package }}"
	{{end}}
)
{{ range .Packages }}var _ {{ .CustomType }}
{{end}}
var _ bytes.Buffer
var _  sync.Pool
var _ = fmt.Print
var _ = errors.New("")
var _ = binary.MaxVarintLen16
var _  json.RawMessage
var _  =  math.Pi
var _ = fmt.Print
{{range .Structs}}
{{$options := .Options}}
{{$structName := .Name}}

{{if $options.ZeroCopy}}
// ZeroCopyView provides zero-copy access to {{.Name}} data
type {{.Name}}View struct {
	data   []byte
	offset int
}

// New{{.Name}}View creates a zero-copy view of {{.Name}} from binary data
func New{{.Name}}View(data []byte) (*{{.Name}}View, error) {
	if len(data) < 4 {
		return nil, errors.New("data too short for header")
	}
	return &{{.Name}}View{data: data, offset: 0}, nil
}

{{range .Fields}}
{{if and (not .ShouldIgnore) .CanZeroCopy}}
{{if eq .TypeName "string"}}
// {{.Name}} returns the {{.BinaryTag}} field as a zero-copy string
func (v *{{$structName}}View) {{.Name}}() (string, error) {
	offset := v.calculateFieldOffset("{{.Name}}")
	if offset < 0 {
		return "", fmt.Errorf("cannot calculate offset for {{.Name}}")
	}
	if offset+2 > len(v.data) {
		return "", fmt.Errorf("buffer too short for {{.Name}} length")
	}
	length := binary.LittleEndian.Uint16(v.data[offset:offset+2])
	if offset+2+int(length) > len(v.data) {
		return "", fmt.Errorf("buffer too short for {{.Name}} data")
	}
	{{if $options.SafeMode}}
	return string(v.data[offset+2:offset+2+int(length)]), nil
	{{else}}
	return bytesToStringUnsafe(v.data[offset+2:offset+2+int(length)]), nil
	{{end}}
}
{{else if eq .TypeName "[]byte"}}
// {{.Name}} returns the {{.BinaryTag}} field as a zero-copy byte slice
func (v *{{$structName}}View) {{.Name}}() ([]byte, error) {
	offset := v.calculateFieldOffset("{{.Name}}")
	if offset < 0 {
		return nil, fmt.Errorf("cannot calculate offset for {{.Name}}")
	}
	if offset+2 > len(v.data) {
		return nil, fmt.Errorf("buffer too short for {{.Name}} length")
	}
	length := binary.LittleEndian.Uint16(v.data[offset:offset+2])
	if offset+2+int(length) > len(v.data) {
		return nil, fmt.Errorf("buffer too short for {{.Name}} data")
	}
	return v.data[offset+2:offset+2+int(length)], nil
}
{{else if eq .TypeName "uint64"}}
// {{.Name}} returns the {{.BinaryTag}} field
func (v *{{$structName}}View) {{.Name}}() (uint64, error) {
	offset := v.calculateFieldOffset("{{.Name}}")
	if offset < 0 {
		return 0, fmt.Errorf("cannot calculate offset for {{.Name}}")
	}
	if offset+8 > len(v.data) {
		return 0, fmt.Errorf("buffer too short for {{.Name}}")
	}
	return binary.LittleEndian.Uint64(v.data[offset:]), nil
}
{{else if eq .TypeName "uint32"}}
// {{.Name}} returns the {{.BinaryTag}} field
func (v *{{$structName}}View) {{.Name}}() (uint32, error) {
	offset := v.calculateFieldOffset("{{.Name}}")
	if offset < 0 {
		return 0, fmt.Errorf("cannot calculate offset for {{.Name}}")
	}
	if offset+4 > len(v.data) {
		return 0, fmt.Errorf("buffer too short for {{.Name}}")
	}
	return binary.LittleEndian.Uint32(v.data[offset:]), nil
}
{{else if eq .TypeName "int64"}}
// {{.Name}} returns the {{.BinaryTag}} field
func (v *{{$structName}}View) {{.Name}}() (int64, error) {
	offset := v.calculateFieldOffset("{{.Name}}")
	if offset < 0 {
		return 0, fmt.Errorf("cannot calculate offset for {{.Name}}")
	}
	if offset+8 > len(v.data) {
		return 0, fmt.Errorf("buffer too short for {{.Name}}")
	}
	return int64(binary.LittleEndian.Uint64(v.data[offset:])), nil
}
{{else if eq .TypeName "int32"}}
// {{.Name}} returns the {{.BinaryTag}} field
func (v *{{$structName}}View) {{.Name}}() (int32, error) {
	offset := v.calculateFieldOffset("{{.Name}}")
	if offset < 0 {
		return 0, fmt.Errorf("cannot calculate offset for {{.Name}}")
	}
	if offset+4 > len(v.data) {
		return 0, fmt.Errorf("buffer too short for {{.Name}}")
	}
	return int32(binary.LittleEndian.Uint32(v.data[offset:])), nil
}
{{else if eq .TypeName "int"}}
// {{.Name}} returns the {{.BinaryTag}} field
func (v *{{$structName}}View) {{.Name}}() (int, error) {
	offset := v.calculateFieldOffset("{{.Name}}")
	if offset < 0 {
		return 0, fmt.Errorf("cannot calculate offset for {{.Name}}")
	}
	if offset+4 > len(v.data) {
		return 0, fmt.Errorf("buffer too short for {{.Name}}")
	}
	return int(binary.LittleEndian.Uint32(v.data[offset:])), nil
}
{{else if eq .TypeName "float32"}}
// {{.Name}} returns the {{.BinaryTag}} field
func (v *{{$structName}}View) {{.Name}}() (float32, error) {
	offset := v.calculateFieldOffset("{{.Name}}")
	if offset < 0 {
		return 0, fmt.Errorf("cannot calculate offset for {{.Name}}")
	}
	if offset+4 > len(v.data) {
		return 0, fmt.Errorf("buffer too short for {{.Name}}")
	}
	return math.Float32frombits(binary.LittleEndian.Uint32(v.data[offset:])), nil
}
{{else if eq .TypeName "float64"}}
// {{.Name}} returns the {{.BinaryTag}} field
func (v *{{$structName}}View) {{.Name}}() (float64, error) {
	offset := v.calculateFieldOffset("{{.Name}}")
	if offset < 0 {
		return 0, fmt.Errorf("cannot calculate offset for {{.Name}}")
	}
	if offset+8 > len(v.data) {
		return 0, fmt.Errorf("buffer too short for {{.Name}}")
	}
	return math.Float64frombits(binary.LittleEndian.Uint64(v.data[offset:])), nil
}
{{else if eq .TypeName "bool"}}
// {{.Name}} returns the {{.BinaryTag}} field
func (v *{{$structName}}View) {{.Name}}() (bool, error) {
	offset := v.calculateFieldOffset("{{.Name}}")
	if offset < 0 {
		return false, fmt.Errorf("cannot calculate offset for {{.Name}}")
	}
	if offset >= len(v.data) {
		return false, fmt.Errorf("buffer too short for {{.Name}}")
	}
	return v.data[offset] != 0, nil
}
{{end}}
{{end}}
{{end}}


// calculateFieldOffset calculates the byte offset for a specific field
func (v *{{.Name}}View) calculateFieldOffset(fieldName string) int {
	offset := 0
	length := uint16(0)
	_ = length
	{{range .Fields}}
		{{if not .ShouldIgnore}}
			if fieldName == "{{.Name}}" {
				return offset
			}
			// Skip {{.Name}} field
			{{if eq .TypeName "string"}}
				if offset+2 > len(v.data) {
					return -1
				}
				length = binary.LittleEndian.Uint16(v.data[offset:offset+2])
				offset += 2 + int(length)
			{{else if eq .TypeName "[]byte"}}
				if offset+2 > len(v.data) {
					return -1
				}
				length = binary.LittleEndian.Uint16(v.data[offset:offset+2])
			{{else if or (eq .TypeName "uint64") (eq .TypeName "int64") (eq .TypeName "float64")}}
				offset += 8

			{{else if or (eq .TypeName "uint32") (eq .TypeName "int32") (eq .TypeName "int") (eq .TypeName "float32") (eq .TypeName "rune")}}
				offset += 4

			{{else if or (eq .TypeName "uint16") (eq .TypeName "int16")}}
				offset += 2

			{{else if or (eq .TypeName "uint8") (eq .TypeName "int8") (eq .TypeName "byte") (eq .TypeName "bool")}}
				offset += 1

			{{else if and .IsStruct (not .IsSlice)}}
				if offset+2 > len(v.data) {
					return -1
				}
				length = binary.LittleEndian.Uint16(v.data[offset:offset+2])
				offset += 2 + int(length)
			{{else if and .IsPointer (not .IsPointerSlice)}}
				offset += 1 // non-nil marker
				if offset > 0 && offset-1 < len(v.data) && v.data[offset-1] != 0 {
					if offset+2 > len(v.data) {
						return -1
					}
					length := binary.LittleEndian.Uint16(v.data[offset:offset+2])
					offset += 2 + int(length)
				}
			{{else if or  .IsSlice .IsPointerSlice }}
				if offset+2 > len(v.data) {
					return -1
				}
				length = binary.LittleEndian.Uint16(v.data[offset:offset+2])
				offset += 2
				{{if eq .TypeName "[]byte"}}
					offset += int(length)
				{{else}}
					// Simplified: assumes fixed-size elements or nested structs
					for i := 0; i < int(length); i++ {
						if offset+2 > len(v.data) {
							return -1
						}
						elemLen := binary.LittleEndian.Uint16(v.data[offset:offset+2])
						offset += 2 + int(elemLen)
					}
				{{end}}
			{{else}}
				offset += 8 // conservative estimate for unknown types
			{{end}}
		{{end}}
	{{end}}
	return offset
}

// ToStruct converts the view to a regular struct (performs copying)
func (v *{{.Name}}View) ToStruct() (*{{.Name}}, error) {
	var s {{.Name}}
	err := s.UnmarshalBorsh(v.data)
	return &s, err
}
	{{end}}


	// Encode creates a deterministic encoding of fields with "enc" tag
func (s {{.Name}}) EncodeFields() (tags []string, encTypes []string, values []any) {
	len := {{sortedEncFieldsLen .Fields}}
	if len > 0 {
		tags = make([]string, len)
		encTypes = make([]string, len)
		values = make([]any, len)
		i := 0
		
		{{range sortedEncFields .Fields}}
			tags[i] = "{{.BinaryTag}}"
			encTypes[i] = "{{.EncType}}"
			values[i] = s.{{.Name}}
			i++
		{{end}}
		_ = i
	}
	
	return tags, encTypes, values
}

	// Encode creates a deterministic encoding of fields with "enc" tag
func (s {{.Name}}) Encode() ([]byte, error) {
	var buf  = &bytes.Buffer{}
	
	{{range sortedEncFields .Fields}}
	// Field: {{.Name}} (tag: {{.BinaryTag}})
	// IsBasicType: {{.IsBasicType}}
	// CustomeFieldEncoder: {{.IsCustomFieldEncoder}}
	// CustomeElementncoder: {{.TypeName}}
	// Field: {{.}}
	
	{

		
		{{ if or .IsPointer .IsPointerSlice }}
			if s.{{.Name}} == nil {
				goto SKIP{{.Name}}
			}
		{{end}}

		{
		
		{{if .IsCustomFieldEncoder}}
			data, err := {{.CustomFieldEncoder}}.Encode(({{.PointerDeref}}(s.{{.Name}})), s)
			if err != nil {
				return nil, fmt.Errorf("failed to encode {{.Name}}: %v", err)
			}
			buf.Write(data)
		{{else if .IsCustomElementEncoder}}
			data, err := {{.CustomElementEncoder}}.Encode(({{.PointerDeref}}(s.{{.Name}})), s)
			if err != nil {
				return nil, fmt.Errorf("failed to encode {{.Name}}: %v", err)
			}
			buf.Write(data)
			
		{{ else if and .Element .Element.IsSlice  }}
				// {{.Name}} ({{.BinaryTag}}) - slice
				// ElementType: {{.Element.TypeName}}
				// Type: {{ .Element.TypeName }}
				// CustomType: {{ .Element.CustomTypeName }}
				// IsCustomEncoder: {{ .IsCustomElementEncoder }}

				
				{{template "encodeSlice" .Element }}

		{{ else if or .IsPointer .IsPointerSlice }}
					// {{.Name}} ({{.BinaryTag}}) - Pointer
					// ElementType: {{.Element.ElementType}}
					// Type: {{ .Element.TypeName }}
					// ActualType: {{ .ActualType }}
					// BasicType: {{ .Element.IsBasicType }}
			
					// ElementType: {{ .Element.TypeName }}


					{{template "encodeScalarElement"  dict
					"Var" (printf "s.%s" .Name)
					"FieldName" .Name
					"ElementType" .ElementType
					"TypeName" .TypeName
					"IsPointer" .IsPointer
					"PointerDeref" .PointerDeref
					"PointerRef" .PointerRef
					"IsCustomElementEncoder" .IsCustomElementEncoder
					"CustomElementEncoder" .CustomElementEncoder
					"IsStruct" .IsStruct
					"IsBasicType" .Element.IsBasicType
					"Element" .Element
					"Field" .
					}}
			
	
		{{else}}
		// IsCustomElementEncoder: {{.IsCustomElementEncoder}}
		// IsCustomElementEncoder: {{.Element.IsCustomElementEncoder}}
					{{template "encodeScalarElement" dict
					"Var" (printf "s.%s" .Name)
					"FieldName" .Name
					"IsSlice" .IsSlice
					"ElementType" .Element.ElementType
					"IsPointer" .Element.IsPointer
					"PointerRef" .Element.PointerRef
					"PointerDeref" .Element.PointerDeref
					"IsCustomElementEncoder" .Element.IsCustomElementEncoder
					"CustomElementEncoder" .Element.CustomElementEncoder
					"IsStruct" .Element.IsStruct
					"IsBasicType" .Element.IsBasicType
					"Element" .Element.Element
					"Field" .Element.Field
				}}
			{{end}}

		}
		{{ if or .Element.IsPointer .Element.IsPointerSlice  }}
			 SKIP{{.Name}}:
		{{end}}

	}
	{{end}}

	return buf.Bytes(), nil
}

{{if $options.ZeroCopy}}
// ToView converts the struct to a zero-copy view
func (s *{{.Name}}) ToView() (*{{.Name}}View, error) {
	data, err := s.MarshalBorsh()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal struct to binary: %v", err)
	}
	return New{{.Name}}View(data)
}
{{end}}

{{template "marshalBinary" .}}

{{template "unmarshalBinary" .}}

{{template "binarySize" .}}

{{if $options.ZeroCopy}}
// UnmarshalBorshZeroCopy creates a zero-copy view
func (s *{{.Name}}) UnmarshalBorshZeroCopy(data []byte) (*{{.Name}}View, error) {
	return New{{.Name}}View(data)
}
{{end}}










{{end}}
`