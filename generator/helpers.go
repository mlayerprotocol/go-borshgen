package generator

import (
	"encoding/binary"
	"errors"
	"fmt"
	"strings"
)

func MarshalBasicTypeFieldTemplate(f FieldInfo) string {
	name := f.Name
	
	ctype := f.CustomTypeName
	if len(ctype) == 0 {
		ctype = f.TypeName
	}
	if f.IsSlice {
		ctype = f.CustomElementTypeName
	}
	ctype = strings.ReplaceAll(ctype, "*", "")
	t := f.TypeName
	if (f.IsSlice || f.IsPointer) && f.ElementType != "" {
		t = f.ElementType
	}

	prefix := ""
	if f.IsPointer {
		prefix = "&"
		
		
	}

	switch t {
	case "string":
		return fmt.Sprintf(`
		// Basictype Unmarshalling
	if offset+2 > len(data) {
		return  nil, fmt.Errorf("buffer too short for %s length")
	}
	length := binary.LittleEndian.Uint16(data[offset:offset+2])
	offset += 2
	if offset+int(length) > len(data) {
		return nil, fmt.Errorf("buffer too short for %s data")
	}
	val := %s(string(data[offset:offset+int(length)]))
	offset += int(length)
	s.%s = %sval
`, name, name, ctype, name, prefix)

	case "[]byte":
		return fmt.Sprintf(`
	if offset+2 > len(data) {
		return fmt.Errorf("buffer too short for %s length")
	}
	length := binary.LittleEndian.Uint16(data[offset:offset+2])
	offset += 2
	if offset+int(length) > len(data) {
		return nil, fmt.Errorf("buffer too short for %s data")
	}
	val := %s(data[offset:offset+int(length)]) // Zero-copy assignment
	offset += int(length)
	s.%s = %sval
`, name, name, ctype, name, prefix)

	case "uint64", "int64", "int":
		return fmt.Sprintf(`
	if offset+8 > len(data) {
		return  nil, fmt.Errorf("buffer too short for %s")
	}
	val := %s(binary.LittleEndian.Uint64(data[offset:offset+8]))
	offset += 8
	s.%s = %sval
`, name, ctype, name, prefix)

	case "uint32", "int32":
		return fmt.Sprintf(`
	if offset+4 > len(data) {
		return fmt.Errorf("buffer too short for %s")
	}
	val := %s(binary.LittleEndian.Uint32(data[offset:offset+4]))
	offset += 4
	s.%s = %sval
`, name, ctype, name, prefix)

	case "uint16", "int16":
		return fmt.Sprintf(`
	if offset+2 > len(data) {
		return fmt.Errorf("buffer too short for %s")
	}
	val := %s(binary.LittleEndian.Uint16(data[offset:offset+2]))
	offset += 2
	s.%s = %sval
`, name, ctype, name, prefix)

	case "uint8", "int8", "byte":
		return fmt.Sprintf(`
	if offset+1 > len(data) {
		return fmt.Errorf("buffer too short for %s")
	}
	val := %s(data[offset])
	offset++
	s.%s = %sval
`, name, ctype, name, prefix)
	case "float32":
		return fmt.Sprintf(`
	if offset+4 > len(data) {
		return fmt.Errorf("buffer too short for %s")
	}
	val := %s(math.Float32frombits(binary.LittleEndian.Uint32(data[offset:offset+4])))
	offset += 4
	s.%s = %sval
`, name, ctype, name, prefix)

	case "float64":
		return fmt.Sprintf(`
	if offset+8 > len(data) {
		return fmt.Errorf("buffer too short for %s")
	}
	val := %s(math.Float64frombits(binary.LittleEndian.Uint64(data[offset:offset+8])))
	offset += 8
	s.%s = %sval
`, name, ctype, name, prefix)

	case "bool":
		return fmt.Sprintf(`
	if offset >= len(data) {
		return fmt.Errorf("buffer too short for %s")
	}
	val := %s(data[offset] != 0)
	offset++
	s.%s = %sval
`, name, ctype, name, prefix)

	default:
		return fmt.Sprintf("// unsupported type: %s\n", f.TypeName)
	}
}


func UnmarshalBasicTypeFieldTemplate(f map[string]interface {}) string {
	name := ""
	if f["FieldName"] != nil {
		if b, ok := f["FieldName"].(string); ok && b != "" {
			name = b
		}
	}
	
	typeName := ""
	if f["TypeName"] != nil {
		if  b, ok := f["TypeName"].(string); ok && b != "" {
			typeName = b
		}
	}
	elementType := ""
	if f["ElementType"] != nil {
		if b, ok := f["ElementType"].(string); ok && b != "" {
			elementType = b
		}
	}
	varName := ""
	if f["Var"] != nil {
		if b, ok := f["Var"].(string); ok && b != "" {
			varName = b
		}
	}
	isPointer := false
	if f["IsPointer"] != nil {
		if b, ok := f["IsPointer"].(bool); ok && b {
			isPointer = true
		}
	}
	isSlice := false
	if f["IsSlice"] != nil {
		if b, ok := f["IsSlice"].(bool); ok && b {
			isSlice = true
		}
	}
	ctype := "";
	ctype = typeName
	
	if len(ctype) == 0 {
			ctype = elementType
		
	}
	
	ctype = strings.ReplaceAll(ctype, "*", "")
	t := elementType
	fmt.Println("IsPonter", f)
	if (len(elementType) == 0 ) || ((isSlice || isPointer) && elementType != "") {
		t = typeName
	}

	prefix := ""
	if isPointer {
		prefix = "&"
	}

	switch t {
	case "string":
		return fmt.Sprintf(`
		// Basictype Unmarshalling
	if offset+2 > len(data) {
		return fmt.Errorf("buffer too short for %s length")
	}
	length := binary.LittleEndian.Uint16(data[offset:offset+2])
	offset += 2
	if offset+int(length) > len(data) {
		return fmt.Errorf("buffer too short for %s data")
	}
	__m := %s(string(data[offset:offset+int(length)]))
	%s = %s(__m)
	offset += int(length)
`, name, name,  ctype, varName, prefix)

	case "[]byte":
		return fmt.Sprintf(`
	if offset+2 > len(data) {
		return fmt.Errorf("buffer too short for %s length")
	}
	length := binary.LittleEndian.Uint16(data[offset:offset+2])
	offset += 2
	if offset+int(length) > len(data) {
		return fmt.Errorf("buffer too short for %s data")
	}
	__m := %s(data[offset:offset+int(length)]) // Zero-copy assignment
	%s = %s(__m)
	offset += int(length)

`, name, name, ctype, varName, prefix)

	case "uint64", "int64", "int":
		return fmt.Sprintf(`
	if offset+8 > len(data) {
		return fmt.Errorf("buffer too short for %s")
	}
	__m := %s(binary.LittleEndian.Uint64(data[offset:offset+8]))
	%s = %s(__m)
	offset += 8
	
`, name,  ctype, varName, prefix)

	case "uint32", "int32":
		return fmt.Sprintf(`
	if offset+4 > len(data) {
		return fmt.Errorf("buffer too short for %s")
	}
	__m := %s(binary.LittleEndian.Uint32(data[offset:offset+4]))
	%s = %s(__m)
	offset += 4
`, name,   ctype, varName, prefix)

	case "uint16", "int16":
		return fmt.Sprintf(`
	if offset+2 > len(data) {
		return fmt.Errorf("buffer too short for %s")
	}
	__m := %s(binary.LittleEndian.Uint16(data[offset:offset+2]))
	%s = %s(__m)
	offset += 2
	
`, name,  ctype, varName, prefix)

	case "uint8", "int8", "byte":
		return fmt.Sprintf(`
	if offset+1 > len(data) {
		return fmt.Errorf("buffer too short for %s")
	}
	__m := %s(data[offset])
	%s = %s(__m)
	offset++
	
`, name,  ctype, varName, prefix)
	case "float32":
		return fmt.Sprintf(`
	if offset+4 > len(data) {
		return fmt.Errorf("buffer too short for %s")
	}
	__m := %s(math.Float32frombits(binary.LittleEndian.Uint32(data[offset:offset+4])))
	%s = %s(__m)
	offset += 4
`, name,  ctype, varName,prefix)

	case "float64":
		return fmt.Sprintf(`
	if offset+8 > len(data) {
		return fmt.Errorf("buffer too short for %s")
	}
	__m := %s(math.Float64frombits(binary.LittleEndian.Uint64(data[offset:offset+8])))
	%s = %s(__m)
	offset += 8
	
`, name,  ctype, varName, prefix )

	case "bool":
		return fmt.Sprintf(`
	if offset >= len(data) {
		return fmt.Errorf("buffer too short for %s")
	}
	__m := %s(data[offset] != 0)
	%s = %s(__m)
	offset++
`, name,  ctype, varName, prefix)

	default:
		return fmt.Sprintf("// unsupported type: %s-%s\n", typeName, elementType)
	}
}



// Helper functions for appending data
func AppendUint16(buf []byte, v uint16) []byte {
	return append(buf, byte(v), byte(v>>8))
}

func AppendUint32(buf []byte, v uint32) []byte {
	return append(buf, byte(v), byte(v>>8), byte(v>>16), byte(v>>24))
}

func AppendUint64(buf []byte, v uint64) []byte {
	return append(buf,
		byte(v), byte(v>>8), byte(v>>16), byte(v>>24),
		byte(v>>32), byte(v>>40), byte(v>>48), byte(v>>56))
}

func AppendBytes(buf, data []byte) []byte {
	buf = AppendUint16(buf, uint16(len(data)))
	return append(buf, data...)
}

func GetBytes(data []byte, offset int) ([]byte, int, error) {
	if offset+2 > len(data) {
		return nil, offset, errors.New("buffer too short for length")
	}
	length := binary.LittleEndian.Uint16(data[offset:offset+2])
	offset += 2
	if offset+int(length) > len(data) {
		return nil, offset, errors.New("buffer too short for data")
	}
	return data[offset:offset+int(length)], offset+int(length), nil
}
