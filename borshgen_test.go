package main

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/mlayerprotocol/go-borshgen/generator"
)
                    

 func TestGenerator(t *testing.T) {

	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		panic("could not get caller in            fo")
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
		// t.Error(err)
 }
