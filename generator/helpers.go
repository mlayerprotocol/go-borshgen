package generator

import (
	"fmt"
	"strings"
)


func UnmarshalBasicTypeFieldTemplate(f FieldInfo) string {
	name := f.Name
	
	ctype := f.CustomTypeName
	if len(ctype) == 0 {
		ctype = f.Type
	}
	if f.IsSlice {
		ctype = f.CustomElementTypeName
	}
	ctype = strings.ReplaceAll(ctype, "*", "")
	t := f.Type
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
		return fmt.Errorf("buffer too short for %s length")
	}
	length := binary.LittleEndian.Uint16(data[offset:offset+2])
	offset += 2
	if offset+int(length) > len(data) {
		return fmt.Errorf("buffer too short for %s data")
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
		return fmt.Errorf("buffer too short for %s data")
	}
	val := %s(data[offset:offset+int(length)]) // Zero-copy assignment
	offset += int(length)
	s.%s = %sval
`, name, name, ctype, name, prefix)

	case "uint64", "int64", "int":
		return fmt.Sprintf(`
	if offset+8 > len(data) {
		return fmt.Errorf("buffer too short for %s")
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
		return fmt.Sprintf("// unsupported type: %s\n", f.Type)
	}
}

