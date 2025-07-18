package generator

import (
	"encoding/json"
	"fmt"
)

type BinaryMarshaler interface {
	MarshalBinary() ([]byte, error)
	BinarySize() (int, error)
}

type BinaryUnMarshaler interface {
	UnmarshalBinary(data []byte) error
}
type BinaryEncoder interface {
Encode() ([]byte, error)
}

type BorshEncoder interface {
	BinaryEncoder
	BinaryMarshaler
	BinaryUnMarshaler
}


type CustomElementEncoder interface {
	MarshalBinary(field any) ([]byte, error)
	UnmarshalBinary(data []byte) (any, error)
	BinarySize(field any) (int, error)
	Encode(field any) ([]byte, error)
	
}
var _DefaultJsonRawMessageEncoder = DefaultJsonRawMessageEncoder{}
var _DefaultByteArrayEncoder = DefaultByteArrayEncoder{}

type DefaultJsonRawMessageEncoder struct {}

func (c DefaultJsonRawMessageEncoder) MarshalBinary(field any, parentStruct any) ([]byte, error) {
	if field == nil {
			return []byte{}, nil
		}
	if f, ok := field.(json.RawMessage); !ok {
		return []byte{}, fmt.Errorf("expected json.RawMessage, got %T", field)
	} else {
		return []byte(f), nil
	}
}
func (c DefaultJsonRawMessageEncoder) UnmarshalBinary(data []byte) (any, error) {
	m := json.RawMessage(data)
	return m, nil
}

func (c DefaultJsonRawMessageEncoder) BinarySize(field any, parentStruct any) (int, error) {
if field == nil {
			return 0, nil
		}
		return len( field.(json.RawMessage)), nil
	
}
func (c DefaultJsonRawMessageEncoder) Encode(field any, parent any) ([]byte, error) {
	if field == nil {
			return []byte{}, nil
		}
	if f, ok := field.(json.RawMessage); !ok {
		return nil, fmt.Errorf("expected json.RawMessage, got %T", field)
	} else {
		return f,nil
	}
}

type DefaultByteArrayEncoder struct {}

func (c DefaultByteArrayEncoder) MarshalBinary(field any, parentStruct any) ([]byte, error) {
	if field == nil {
			return []byte{}, nil
		}
	if f, ok := field.([]byte); !ok {
		return []byte{}, fmt.Errorf("expected []byte, got %T", field)
	} else {
		return f, nil
	}
}
func (c DefaultByteArrayEncoder) UnmarshalBinary(data []byte) (any, error) {
	m := data
	return m, nil
}

func (c DefaultByteArrayEncoder) BinarySize(field any, parentStruct any) (int, error) {
	if field == nil {
			return 0, nil
		}
		return len(field.([]byte)), nil
}
func (c DefaultByteArrayEncoder) Encode(field any, parentStruct any) ([]byte, error) {
	if field == nil {
			return []byte{}, nil
		}
	if f, ok := field.([]byte); !ok {
		return nil, fmt.Errorf("expected []byte, got %T", field)
	} else {
		return f,nil
	}
}

