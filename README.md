# go-borshgen



**go-borshgen** is an go-gen implementation of the [Borsh] binary serialization format for Go
projects for performance critical applications. It avoids the overhead added by Go reflections.

Borsh stands for _Binary Object Representation Serializer for Hashing_. It is
meant to be used in security-critical projects as it prioritizes consistency,
safety, speed, and comes with a strict specification.

## Features

- Go generator for supported types and custom encoders for complex types



## Usage

See sample usage in in **tests/testhelper.go**

- Add ``` //go:generate borshgen -tag=msg -fallback=json ``` comment over all structs that require code generation
- Add the relevant tags
- Attach custom Parsers for unsupported types (see table of supported tags below)
- Run generator ``` borshgen -<input file or directory> ```
- 

### Examples/How to Test
1. Run the generator tests in **borshgen_test.go** file within the root directory. This will
generate the helper methods within **tests** directory.
2. Run all the tests within **tests/e2e_test.go**




## Type Mappings

Borsh                 | Go           |  Description
--------------------- | -------------- |--------
`bool`		      | `bool`	       |
`u8` integer          | `uint8`        |
`u16` integer         | `uint16`       |
`u32` integer         | `uint32`       |
`u64` integer         | `uint64`       |
`u128` integer        | `big.Int`  |
`i8` integer          | `int8`        |
`i16` integer         | `int16`       |
`i32` integer         | `int32`       |
`i64` integer         | `int64`       |
`i128` integer        |            |  Not supported yet
`f32` float           | `float32`      |
`f64` float           | `float64`      |
fixed-size array      | `[size]type`   |  **Not supported yet**
dynamic-size array    |  `[]type`      |  go slice
string                | `string`       |
option                |  `*type`         |   go pointer
map                   |   `map`          |
set                   |   `map[type]struct{}`  | **Not supported yet**
structs               |   `struct`      |
enum                  |   `borsh.Enum`  |    **Not supported yet**
