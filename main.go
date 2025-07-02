package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/mlayerprotocol/go-borshgen/generator"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: borshgen <input.go>") 
		fmt.Println("Options:")
		fmt.Println("  //go:generate borshgen -tag=msg -fallback=json -encode-tag=enc")
		fmt.Println("  //go:generate borshgen -tag=binary -fallback=msg")
		fmt.Println("  //go:generate borshgen -ignore=- -max-string=32767")
		fmt.Println("  //go:generate borshgen -zero-copy -unsafe")
		os.Exit(1)
	}

	inputFile := os.Args[1]
	
	// Default values
	primaryTag := "msg"
	fallbackTag := "json"
	ignoreTag := "-"
	usePooling := true
	maxString := 32767
	// zeroCopy := false
	// safeMode := true
	encodeTag := "enc"
	var err error
	// Parse additional flags
	for i := 2; i < len(os.Args); i++ {
		arg := os.Args[i]
		
		if strings.HasPrefix(arg, "-tag=") {
			primaryTag = strings.TrimPrefix(arg, "-tag=")
		} else if strings.HasPrefix(arg, "-fallback=") {
			fallbackTag = strings.TrimPrefix(arg, "-fallback=")
		} else if strings.HasPrefix(arg, "-ignore=") {
			ignoreTag = strings.TrimPrefix(arg, "-ignore=")
		} else if arg == "-no-pool" {
			usePooling = false
		// } else if arg == "-zero-copy" {
		// 	zeroCopy = false // TODO: not yet tested
		// } else if arg == "-unsafe" {
		// 	safeMode = false
		} else if strings.HasPrefix(arg, "-max-string=") {
			maxString, err = strconv.Atoi(strings.TrimPrefix(arg, "-max-string="))

		} else if strings.HasPrefix(arg, "-encodeTag=") {
			encodeTag = strings.TrimPrefix(arg, "-encode-tag=")

		}
		if err != nil {
			fmt.Printf("Invalid max-string value: %v\n", err)
			os.Exit(1)
		}
	}

	// var err error
	// if zeroCopy {
	// 	err = GenerateWithZeroCopy(inputFile, primaryTag, fallbackTag, ignoreTag, usePooling, zeroCopy, safeMode, maxString)
	// } else {
	// 	err = Generate(inputFile,  "", primaryTag, fallbackTag, encodeTag, ignoreTag,   usePooling, maxString)
	// }
	err = generator.Generate(inputFile,  "", primaryTag, fallbackTag, encodeTag, ignoreTag,  usePooling, maxString)
	
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}