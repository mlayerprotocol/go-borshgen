package templates

// Complete template with all necessary functions
const MainTemplateBackup = `

package {{.Package}}

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	{{if and .Options.ZeroCopy (not .Options.SafeMode)}}"unsafe"{{end}}
	{{ range .Packages }}"{{ .Package }}"
	{{end}}
)
{{ range .Packages }}var _ {{ .CustomType }}
{{end}}
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
	err := s.UnmarshalBinary(v.data)
	return &s, err
}
	{{end}}


	// Encode creates a deterministic encoding of fields with "enc" tag
func (s *{{.Name}}) Encode() ([]byte, error) {
	var buf []byte
	
	{{range sortedEncFields .Fields}}
	// Field: {{.Name}} (tag: {{.BinaryTag}})
	// IsBasicType: {{.Element.IsBasicType}}
	// CustomeFieldEncoder: {{.IsCustomFieldEncoder}}
	// CustomeElementncoder: {{.Element.TypeName}}
	
	{

		
		{{ if or .Element.IsPointer .Element.IsPointerSlice }}
			if s.{{.Name}} == nil {
				goto SKIP{{.Name}}
			}
		{{end}}

		{
		
		{{if .IsCustomFieldEncoder}}
			data, err := {{.CustomFieldEncoder}}.Encode(({{.PointerDeref}}(s.{{.Name}})))
			if err != nil {
				return nil, fmt.Errorf("failed to encode {{.Name}}: %v", err)
			}
			buf = append(buf, data...)
		{{else if .IsCustomElementEncoder}}
			data, err := {{.CustomElementEncoder}}.Encode(({{.PointerDeref}}(s.{{.Name}})))
			if err != nil {
				return nil, fmt.Errorf("failed to encode {{.Name}}: %v", err)
			}
			buf = append(buf, data...)
		{{ else if .Element.IsSlice  }}
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
					"ElementType" .Element.ElementType
					"IsPointer" .Element.IsPointer
					"PointerDeref" .Element.PointerDeref
					"IsCustomElementEncoder" .Element.IsCustomElementEncoder
					"CustomElementEncoder" .Element.CustomElementEncoder
					"IsStruct" .Element.IsStruct
					"IsBasicType" .Element.IsBasicType
					"Element" .Element.Element
					"Field" .
					}}
			
	
		{{else}}
		// IsCustomElementEncoder: {{.IsCustomElementEncoder}}
		// IsCustomElementEncoder: {{.Element.IsCustomElementEncoder}}
					{{template "encodeScalarElement" dict
					"Var" (printf "s.%s" .Name)
					"FieldName" .Name
					"ElementType" .Element.ElementType
					"IsPointer" .Element.IsPointer
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

	return buf, nil
}

{{if $options.ZeroCopy}}
// ToView converts the struct to a zero-copy view
func (s *{{.Name}}) ToView() (*{{.Name}}View, error) {
	data, err := s.MarshalBinary()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal struct to binary: %v", err)
	}
	return New{{.Name}}View(data)
}
{{end}}

{{template "marshalBinary" .}}

// UnmarshalBinary decodes binary data to {{.Name}}
func (s *{{.Name}}) UnmarshalBinary(data []byte) error {
    offset := 0
    var err error
    {{range .Fields}}
    {{if not .ShouldIgnore}}
	
	
    {
	{{ if  or .IsPointer .IsPointerSlice}}
			
			ptr := data[offset]
			if int(ptr) == 0 {
				s.{{.Name }} = nil
				offset++
				goto SKIP{{.Name}}
			} 
				offset++
				
	{{ end }}
		 {
        // {{.Name}} ({{.BinaryTag}})
		// Type - {{.TypeName}}
		// IsPointer - {{.IsPointer}}
		 // ElementType - {{.ElementType}}
		
		{{if .Element.IsCustomElementEncoder }}
			// Custom encoder
			// IsCustomElementEncoder: {{.Element.IsCustomElementEncoder}}
			// ElementType: {{.ElementType}}
			// PointerToSlice: {{.IsPointerSlice}}
		
				
				if offset+2 > len(data) {
					return fmt.Errorf("buffer too short for {{.Name}} length")
				}
					var itemData []byte
				itemData, offset, err = getBytes(data, offset)
				if _v, err := {{.Element.CustomElementEncoder}}.UnmarshalBinary(itemData); err != nil {
					return fmt.Errorf("failed to unmarshal custom encoder slice {{.Name}}]: %v", err)
				} else {
				 	_m := (_v).({{ .CustomTypeName}})
					s.{{.Name}} = {{.PointerRef}}_m
				}
					




        {{else if  .IsBasicType  }}

			
          		 {{ unmarshalBasicTypeTemplate . }}

        {{else if  and .IsStruct (not .IsSlice)}}
            // Nested struct
            var nestedData []byte
            nestedData, offset, err = getBytes(data, offset)
            if err != nil {
                return fmt.Errorf("failed to get bytes for nested struct {{.Name}}: %v", err)
            }
            if err = s.{{.Name}}.UnmarshalBinary(nestedData); err != nil {
                return fmt.Errorf("failed to unmarshal nested struct {{.Name}}: %v", err)
            }
        {{else if and .IsPointer (not .IsPointerSlice)}}
            // Pointer
            if offset >= len(data) {
                return fmt.Errorf("buffer too short for {{.Name}} nil marker")
            }
            if int(data[offset]) == 0 {
                s.{{.Name}} = nil
                offset++
            } else {
                offset++
                {{if isBasicElementType .  }}
					{{ unmarshalBasicTypeTemplate . }}
                {{else if .IsStruct}}
					var nestedData []byte
					nestedData, offset, err = getBytes(data, offset)
					if err != nil {
						return fmt.Errorf("failed to get bytes for nested struct pointer {{.Name}}: %v", err)
					}
					s.{{.Name}} = &{{.ElementType}}{}
					if err = s.{{.Name}}.UnmarshalBinary(nestedData); err != nil {
						return fmt.Errorf("failed to unmarshal nested struct pointer {{.Name}}: %v", err)
					}
                {{else}}
					var val {{.ElementType}}
					var valueData []byte
					valueData, offset, err = getBytes(data, offset)
					if err != nil {
						return fmt.Errorf("failed to get bytes for pointer {{.Name}}: %v", err)
					}
					if err = unmarshalValue(valueData, {{.ElementPointerRef}}val); err != nil {
						return fmt.Errorf("failed to unmarshal pointer {{.Name}}: %v", err)
					}
					s.{{.Name}} = {{.ElementPointerRef}}val
                {{end}}
            }
        {{else if  or .IsSlice (.IsPointerSlice)}}
            // Slice
			// IsPointerSlice {{.IsPointerSlice}}
			// IsBasicType {{.IsBasicType}}
			// CustomeElementTypeNameicType {{.CustomElementTypeName}}
			// ElementType {{.ElementType}}
			
			if offset+2 > len(data) {
				return fmt.Errorf("buffer too short for {{.Name}} length")
			}

			{{ if not .IsFixedArray }}
				length := binary.LittleEndian.Uint16(data[offset : offset+2])
				offset += 2
				p := make([]{{.CustomElementTypeName}}, length)
			{{ else }}
			 	var p = [{{.FixedArrayLength}}]{{.CustomElementTypeName}}
				length := {{.FixedArrayLength}}
			{{ end }}
			
			
			s.{{.Name}} = {{.PointerRef}}p

			{{if or (eq .ElementType "byte") (eq .ElementType "int8") (eq .ElementType "uint8")}}
				if offset+int(length) > len(data) {
					return fmt.Errorf("buffer too short for {{.Name}} data")
				}
				v := {{.CustomTypeName}}(data[offset:offset+int(length)]) // Zero-copy assignment
				{{ if not .IsFixedArray }}
					(s.{{.Name}}) = {{.PointerRef}}v
				{{else}}
					var _m = [{{.FixedArrayLength}}]{{.CustomElementTypeName}}{}
					copy(v, _m)
					(s.{{.Name}}) = {{.PointerRef}}_m
				{{end}}
				offset += int(length)

			{{else if eq .ElementType "string"}}
				
				
				{{ if not .IsFixedArray }}
				 	var tmp = []{{.CustomElementTypeName}}{}
					for i := 0; i < int(length); i++ {
						var strData []byte
						strData, offset, err = getBytes(data, offset)
						if err != nil {
							return fmt.Errorf("failed to decode string for {{.Name}}[%d]: %v", i, err)
						}
						tmp[i] = {{.CustomElementTypeName}}(string(strData))
					}
					((s.{{.Name}})) = {{.PointerRef}}tmp
				{{ else }}
				 var tmp = [{{.FixedArrayLength}}]{{.CustomElementTypeName}}{}
					dataLen := length*32 // fix array strings are handles as fixed array [n][32]byte
					if len(data[offset:]) < end {
						return fmt.Errorf("failed to decode [%d]string for {{.Name}}[%d]: %v", length, i, err)
					}
					n := dataLen/{{.FixedArrayLength}}
					for i := 0; i < n; i++ {
						var strData []byte
						strData, offset = getFixedBytes(data, offset, 32)
						if err != nil {
							return fmt.Errorf("failed to decode string for {{.Name}}[%d]: %v", i, err)
						}
						tmp[i] = {{.CustomElementTypeName}}(string(strData))
					}
				((s.{{.Name}})) = {{.PointerRef}}tmp
				{{end}}

			{{else if eq .ElementType "bool"}}
				tmp := make([]{{.CustomElementTypeName}}, length)
				for i := 0; i < int(length); i++ {
					if offset >= len(data) {
						return fmt.Errorf("buffer too short for {{.Name}}[%d] 1-byte value", i)
					}
					tmp[i] = {{.CustomElementTypeName}}(data[offset])
					offset++
				}
				((s.{{.Name}})) = {{.PointerRef}}tmp

			{{else if or (eq .ElementType "int16") (eq .ElementType "uint16")}}
				tmp := make([]{{.CustomElementTypeName}}, length)
				for i := 0; i < int(length); i++ {
					if offset+2 > len(data) {
						return fmt.Errorf("buffer too short for {{.Name}}[%d] 2-byte value", i)
					}
					tmp[i] = {{.CustomElementTypeName}}(binary.LittleEndian.Uint16(data[offset:offset+2]))
					offset += 2
				}
				(s.{{.Name}}) = {{.PointerRef}}tmp

			{{else if or (eq .ElementType "int32") (eq .ElementType "uint32")}}
				tmp := make([]{{.CustomElementTypeName}}, length)
				for i := 0; i < int(length); i++ {
					if offset+4 > len(data) {
						return fmt.Errorf("buffer too short for {{.Name}}[%d] 4-byte value", i)
					}
					tmp[i] = {{.CustomElementTypeName}}(binary.LittleEndian.Uint32(data[offset:offset+4]))
					offset += 4
				}
				((s.{{.Name}})) = {{.PointerRef}}tmp

			{{else if or (eq .ElementType "int64") (eq .ElementType "uint64")  (eq .ElementType "int") (eq .ElementType "uint")}}
				tmp := make([]{{.CustomElementTypeName}}, length)
				for i := 0; i < int(length); i++ {
					if offset+8 > len(data) {
						return fmt.Errorf("buffer too short for {{.Name}}[%d] 8-byte value", i)
					}
					tmp[i] = {{.CustomElementTypeName}}(binary.LittleEndian.Uint64(data[offset:offset+8]))
					offset += 8
				}
				(s.{{.Name}}) = {{.PointerRef}}tmp

			{{else if eq .ElementType "float32"}}
				tmp := make([]{{.CustomElementTypeName}}, length)
				for i := 0; i < int(length); i++ {
					if offset+4 > len(data) {
						return fmt.Errorf("buffer too short for {{.Name}}[%d] float32", i)
					}
					tmp[i] = {{.CustomElementTypeName}}(math.Float32frombits(binary.LittleEndian.Uint32(data[offset:offset+4])))
					offset += 4
				}
				(s.{{.Name}}) = {{.PointerRef}}tmp

			{{else if eq .ElementType "float64"}}
				tmp := make([]{{.CustomElementTypeName}}, length)
				for i := 0; i < int(length); i++ {
					if offset+8 > len(data) {
						return fmt.Errorf("buffer too short for {{.Name}}[%d] float64", i)
					}
					tmp[i] = {{.CustomElementTypeName}}(math.Float64frombits(binary.LittleEndian.Uint64(data[offset:offset+8])))
					offset += 8
				}
				(s.{{.Name}}) = {{.PointerRef}}tmp

			

			{{else if .IsCustomType}}
				tmp := make([]{{.CustomElementTypeName}}, length)
				for i := 0; i < int(length); i++ {
					item := new({{.ElementType}})
					if err := (item).UnmarshalBinary(data); err != nil {
						return fmt.Errorf("failed to decode custom type for {{.Name}}[%d]: %v", i, err)
					}
					tmp[i] = {{.CustomElementTypeName}}(*item)
					offset += 2 + len(data)
				}
				(s.{{.Name}}) = {{.PointerRef}}tmp
            {{else if eq .TypeName "[]byte"}}
				tmp := make([]{{.CustomElementTypeName}}, length)
				for i := 0; i < int(length); i++ {
					var strData []byte
					strData, offset, err = getBytes(data, offset)
					if err != nil {
						return fmt.Errorf("failed to decode string for {{.Name}}[%d]: %v", i, err)
					}
					tmp[i] = {{.CustomElementTypeName}}(strData)
				}
				((s.{{.Name}})) = {{.PointerRef}}tmp
            {{else if .IsStruct}} // TODO
				for i := 0; i < int(length); i++ {
					var itemData []byte
					itemData, offset, err = getBytes(data, offset)
					if err != nil {
						return fmt.Errorf("failed to get bytes for struct slice element {{.Name}}[%d]: %v", i, err)
					}
					if err = ({{.PointerDeref}}(s.{{.Name}}))[i].UnmarshalBinary(itemData); err != nil {
						return fmt.Errorf("failed to unmarshal struct slice element {{.Name}}[%d]: %v", i, err)
					}
				}
            // ({{.PointerDeref}}(s.{{.Name}})) = {{.CustomTypeName}}(vv)
            {{else}}
            for i := 0; i < int(length); i++ {
               	{{if eq .ElementType "string"}}
					var itemData []byte
					itemData, offset, err = getBytes(data, offset)
					if err != nil {
						return fmt.Errorf("failed to get bytes for string slice element {{.Name}}[%d]: %v", i, err)
					}
					val := {{.CustomElementTypeName}}(string(itemData))
					({{.PointerDeref}}(s.{{.Name}}))[i] = {{.ElementPointerRef}}val

				{{else if eq .ElementType "[]byte"}}
					var itemData []byte
					itemData, offset, err = getBytes(data, offset)
					if err != nil {
						return fmt.Errorf("failed to get bytes for []byte slice element {{.Name}}[%d]: %v", i, err)
					}
					val := {{.CustomElementTypeName}}(itemData)
					({{.PointerDeref}}(s.{{.Name}}))[i] = {{.ElementPointerRef}}val

				{{else if isBasicElementType . }}
					{{ unmarshalBasicTypeTemplate . }}
					
				{{else}}
					var itemData []byte
					itemData, offset, err = getBytes(data, offset)
					if err != nil {
						return fmt.Errorf("failed to get bytes for slice element {{.Name}}[%d]: %v", i, err)
					}
						if err = unmarshalValue(itemData, ({{.PointerDeref}}(s.{{.Name}}))[i]); err != nil {
							return fmt.Errorf("failed to unmarshal slice element {{.Name}}[%d]: %v", i, err)
						}
				
					
				{{end}}
				
				
            }
				
            // s.{{.Name}} = {{.CustomTypeName}}(vv)
            {{end}}			
        {{else}}
            // Custom type
            var valueData []byte
            valueData, offset, err = getBytes(data, offset)
            if err != nil {
                return fmt.Errorf("failed to get bytes for custom type {{.Name}}: %v", err)
            }
            if v, ok := interface{}(s.{{.Name}}).(BinaryUnMarshaler); ok {
                if err = v.UnmarshalBinary(valueData); err != nil {
                    return fmt.Errorf("failed to unmarshal custom type {{.Name}}: %v", err)
                }
            } else if err = unmarshalValue(valueData, &s.{{.Name}}); err != nil {
                return fmt.Errorf("failed to unmarshal custom type {{.Name}}: %v", err)
            }
        {{end}}
		}
		{{ if  or .IsPointer .IsPointerSlice}}
		SKIP{{.Name}}:
		{{end}}
    }
    {{end}}
    {{end}}
    return err
}


{{if $options.ZeroCopy}}
// UnmarshalBinaryZeroCopy creates a zero-copy view
func (s *{{.Name}}) UnmarshalBinaryZeroCopy(data []byte) (*{{.Name}}View, error) {
	return New{{.Name}}View(data)
}
{{end}}

// just a placeholder for the math package
func (s *{{.Name}}) mathPlaceHolder()  bool {
	return math.Ceil(0) == 0
}







// binarySize calculates the size needed for binary encoding
func (s *{{.Name}}) BinarySize() int {
	
	size := 0
	{{range .Fields}}
	{{if not .ShouldIgnore}}
	{
		// {{ .Name }}
		// {{.CustomTypeName}}
		// {{.TypeName}}
		// {{.ActualType}}
		// {{.CustomElementTypeName}}
		// {{.ElementType}}
			{{ if or .IsPointer .IsPointerSlice}}
			size++
			if s.{{.Name}} == nil {
				goto SKIP{{.Name}}
			}
			{{end}}

		{
		// {{.Name}} ({{.BinaryTag}})
		// Type: {{.TypeName}}
		// IsPointer: {{.IsPointer}}
		// ElementType: {{.ElementType}}


		{{if .Element.IsCustomElementEncoder }}
	
			s, err := {{.Element.CustomElementEncoder}}.BinarySize({{.PointerDeref}}s.{{.Name}})
				if err != nil {
					panic(fmt.Sprintf("failed to calculate binary size for custom encoder {{.Name}}: %v", err))
				}
				size += 2 + s
		
		
		  
           

		{{else if  .IsBasicType }}
			{{if eq .TypeName "string"}}
				size += 2 + len({{.PointerDeref}}s.{{.Name}}) // string: 2-byte length + content
			{{else if eq .TypeName "[]byte"}}
				size += 2 + len({{.PointerDeref}}s.{{.Name}}) // []byte: 2-byte length + content
			{{else if or (eq .TypeName "uint64") (eq .TypeName "int64") (eq .TypeName "float64")}}
				size += 8 // 64-bit: uint64, int64, float64
			{{else if or (eq .TypeName "uint32") (eq .TypeName "int32") (eq .TypeName "int") (eq .TypeName "float32") (eq .TypeName "rune")}}
				size += 4 // 32-bit: uint32, int32, int, float32, rune
			{{else if or (eq .TypeName "uint16") (eq .TypeName "int16")}}
				size += 2 // 16-bit: uint16, int16
			{{else if or (eq .TypeName "uint8") (eq .TypeName "int8") (eq .TypeName "byte")}}
				size += 1 // 8-bit: uint8, int8, byte
			{{else if eq .TypeName "bool"}}
				size += 1 // bool: 1 byte
			{{else}}
				size += 2 + s.{{.Name}}.BinarySize()
			{{end}}
		{{else if and .IsPointer (not .IsPointerSlice) }}
			
			if s.{{.Name}} != nil {
				{{if .IsStruct}}
				size += 2 + s.{{.Name}}.BinarySize()

			{{else if eq .ElementType "string"}}
				ptr := *s.{{.Name}}
				size += 2 + len(ptr)

			{{else if eq .ElementType "[]byte"}}
				ptr := *s.{{.Name}}
				size += 2 + len(ptr)

			{{else if or (eq .ElementType "uint64") (eq .ElementType "int64") (eq .ElementType "float64")}}
				size += 8

			{{else if or (eq .ElementType "uint32") (eq .ElementType "int32") (eq .ElementType "int") (eq .ElementType "float32") (eq .ElementType "rune")}}
				size += 4

			{{else if or (eq .ElementType "uint16") (eq .ElementType "int16")}}
				size += 2

			{{else if or (eq .ElementType "uint8") (eq .ElementType "int8") (eq .ElementType "byte")}}
				size += 1

			{{else if eq .ElementType "bool"}}
				size += 1

			{{else}}
				// Fallback estimate for custom or unknown pointer types
				size += 2 + 64
			{{end}}

			}
		{{else if or .IsSlice (.IsPointerSlice)}}
			{{ if .IsPointerSlice}} 
				if s.{{.Name}} == nil {
					return size
				}
			{{end}}
			j := ({{.PointerDeref}}s.{{.Name}})
			_ = j
			size += 2 // for length prefix
			{{if or (eq .TypeName "[]byte") (eq .TypeName "*[]byte")}}
				size += len(j)

			{{else if .IsStruct}}
				for i := range j {
					size += 2 + j[i].BinarySize()
				}

			{{else if or (eq .ElementType "string") (eq .ElementType "[]byte]")}}
				for _, item := range j {
					size += 2 + len(item)
				}

			{{else if eq .ElementType "[]byte"}}
				for _, item := range j {
					size += 2 + len(item)
				}

			{{else if or (eq .ElementType "uint64") (eq .ElementType "int64") (eq .ElementType "float64")}}
				size += len(j) * 8

			{{else if or (eq .ElementType "uint32") (eq .ElementType "int32") (eq .ElementType "int") (eq .ElementType "float32") (eq .ElementType "rune")}}
				size += len(j) * 4

			{{else if or (eq .ElementType "uint16") (eq .ElementType "int16")}}
				size += len(j) * 2

			{{else if or (eq .ElementType "uint8") (eq .ElementType "int8") (eq .ElementType "byte") (eq .ElementType "bool")}}
				size += len(j)

			{{else}}
				for _, item := range ({{.PointerDeref}}s.{{.Name}}) {
					// Fallback: try calling item.BinarySize() if available
					if bs, ok := interface{}(item).(interface{ BinarySize() int }); ok {
						size += 2 + bs.BinarySize()
					} else {
						size += 2 + 64 // conservative default
					}
				}
			{{end}}

		{{else}}
			// Fallback for custom types
			if bs, ok := interface{}(s.{{.Name}}).(interface{ BinarySize() int }); ok {
				size += 2 + bs.BinarySize()
			} else {
				size += 2 + 64
			}
		{{end}}
		}
		{{ if or .IsPointer .IsPointerSlice}}
			 SKIP{{.Name}}:
			{{end}}		
	}
	{{end}}
	{{end}}
	return size
}


{{end}}
`