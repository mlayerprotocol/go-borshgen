package main

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/mlayerprotocol/go-borshgen/generator"
)

// Test data structures
const testStructSource = `
package test

//go:generate bingen -tag=msg -fallback=json -zero-copy

// Simple struct with basic types
//go:generate bingen -tag=msg -fallback=json
type SimpleStruct struct {
	Name    string  ` + "`msg:\"name\" json:\"name\"`" + `
	Age     int32   ` + "`msg:\"age\" json:\"age\"`" + `
	Height  float32 ` + "`msg:\"height\" json:\"height\"`" + `
	Active  bool    ` + "`msg:\"active\" json:\"active\"`" + `
	Data    []byte  ` + "`msg:\"data\" json:\"data\"`" + `
	Ignored string  ` + "`msg:\"-\" json:\"-\"`" + `
}

// Complex struct with enc tags and nested structs
//go:generate bingen -tag=msg -fallback=json -zero-copy
type ComplexStruct struct {
	ID        string       ` + "`msg:\"id\" json:\"id\" enc:\"func\"`" + `
	Timestamp uint64       ` + "`msg:\"ts\" json:\"timestamp\" enc:\"int\"`" + `
	Amount    uint64       ` + "`msg:\"amount\" json:\"amount\" enc:\"int\"`" + `
	Nested    NestedStruct ` + "`msg:\"nested\" json:\"nested\" enc:\"func\"`" + `
	Optional  *bool        ` + "`msg:\"opt\" json:\"optional,omitempty\"`" + `
	Metadata  string       ` + "`json:\"meta\"`" + ` // Uses fallback tag
	Private   string       ` + "`msg:\"-\" json:\"-\"`" + ` // Ignored
}

// Nested struct
//go:generate bingen -tag=msg -fallback=json -zero-copy  
type NestedStruct struct {
	Key   string ` + "`msg:\"key\" json:\"key\" enc:\"func\"`" + `
	Value string ` + "`msg:\"val\" json:\"value\" enc:\"func\"`" + `
	Count int32  ` + "`msg:\"count\" json:\"count\" enc:\"int\"`" + `
}

// Struct without generate comment (should be ignored)
type IgnoredStruct struct {
	Field string ` + "`msg:\"field\"`" + `
}

// Struct with different options
//go:generate borshgen -tag=binary -fallback=msg -no-pool -unsafe
type OptionsStruct struct {
	Data []byte ` + "`binary:\"data\" msg:\"data\"`" + `
}
`
                        

 func TestGenerator(t *testing.T) {

	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		panic("could not get caller info")
	}
	// return filepath.Dir(filename)
	dir :=  filepath.Dir(filename)
	//	tmpFile, err := f, err := os.Create(fileName)
		// if err != nil {
		// 	t.Fatalf("Failed to create temp file: %v", err)
		// }
		
	//	defer os.Remove(t mpFile)
		// dir := filepath.Dir(tmpFile)
		err := generator.GenerateDir(filepath.Join(dir, "tests"), "msg", "json", "enc", "-", true, 1024 * 100 )
		if err != nil {
			t.Log("Successfully Generate files")
		}
		t.Error(err)
 }
