package templates

// Complete template with all necessary functions
const MarshalBinaryTemplate = `// Code generated by bingen. DO NOT EDIT.


// MarshalBinary marshals {{.Name}} to binary format
{{define "marshalBinary"}}
func (s {{.Name}}) MarshalBinary() ([]byte, error) {
	{{if .Options.UsePooling}}
		buf := binaryBufPool.Get().([]byte)
		buf = buf[:0]
		defer binaryBufPool.Put(buf)
		{{else}}
		var buf []byte
	{{end}}

	// Calculate size first
	size, err := s.BinarySize()
	if err != nil {
		return nil, err
	}
	{{if .Options.UsePooling}}
	if cap(buf) < size {
		buf = make([]byte, 0, size)
	}
	{{else}}
		buf = make([]byte, 0, size)
	{{end}}

	{{range .Fields}}
		{{if not .ShouldIgnore}}
		
		

		{{if or .IsPointer .IsPointerSlice}}
		{
			if s.{{.Name}} == nil {
				buf = append(buf, 0) // nil marker
				goto SKIP{{.Name}}
			} else {
				buf = append(buf, 1) // non-nil marker
			}
		}
		{{end}}
		{

		{{if .IsCustomFieldEncoder}}
			data, err := {{.CustomFieldEncoder}}.MarshalBinary(({{.PointerDeref}}(s.{{.Name}})), s)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal {{.Name}}: %v", err)
			}
			buf = appendBytes(buf, data)
		{{else if .IsCustomElementEncoder}}
			data, err := {{.CustomElementEncoder}}.MarshalBinary(({{.PointerDeref}}(s.{{.Name}})), s)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal {{.Name}}: %v", err)
			}
			buf = appendBytes(buf, data)
		{{ else if or .IsSlice  .Element.IsSlice  }}
				// {{.Name}} ({{.BinaryTag}}) - slice
				// ElementType: {{.Element.TypeName}}
				// Type: {{ .Element.TypeName }}
				// CustomType: {{ .Element.CustomTypeName }}
				// IsCustomEncoder: {{ .IsCustomElementEncoder }}

				
				{{template "marshalSlice"  .Element }}

		{{ else if or .IsPointer .IsPointerSlice }}
					// {{.Name}} ({{.BinaryTag}}) - Pointer
					// ElementType: {{.Element.ElementType}}
					// Type: {{ .Element.TypeName }}
					// ActualType: {{ .ActualType }}
					// BasicType: {{ .Element.IsBasicType }}
			
					// ElementType: {{ .Element.TypeName }}


					{{template "marshalScalarElement"  dict
					"Var" (printf "s.%s" .Name)
					"FieldName" .Name
					"TypeName" .Element.TypeName
					"ElementType" .Element.ElementType
					"IsPointer" .Element.IsPointer
					"PointerDeref" .Element.PointerDeref
					"PointerRef" .Element.PointerRef
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
					{{template "marshalScalarElement" dict
					"Var" (printf "s.%s" .Name)
					"FieldName" .Name
					"ElementType" .ElementType
					"TypeName" .TypeName
					"IsSlice" .IsSlice
					"IsPointer" .IsPointer
					"PointerDeref" .PointerDeref
					"PointerRef" .PointerRef
					"IsCustomElementEncoder" .IsCustomElementEncoder
					"CustomElementEncoder" .CustomElementEncoder
					"IsStruct" .IsStruct
					"IsBasicType" .IsBasicType
					"Element" .Element
					"Field" .Field
				}}
			{{end}}
				
			
		}
			{{if .IsPointer}}
					SKIP{{.Name}}:
				{{end}}
		{{end}}
	{{end}}

	{{if .Options.UsePooling}}
		// Copy result before returning buffer to pool
		result := make([]byte, len(buf))
		copy(result, buf)
		return result, nil
		{{else}}
		return buf, nil
	{{end}}
}
{{end}}
`