package generator

import (
	"encoding/json"
	"fmt"
)
type CustomElementEncoder interface {
	MarshalBinary(field any) ([]byte, error)
	UnmarshalBinary(data []byte) (any, error)
	BinarySize(field any) (int, error)
	Encode(field any) ([]byte, error)
	
}
var _CustomJsonRawMessageEncoder = CustomJsonRawMessageEncoder{}
var _CustomByteArrayEncoder = CustomByteArrayEncoder{}

type CustomJsonRawMessageEncoder struct {}

func (c CustomJsonRawMessageEncoder) MarshalBinary(field any) ([]byte, error) {
	if f, ok := field.(json.RawMessage); !ok {
		return []byte{}, fmt.Errorf("expected json.RawMessage, got %T", field)
	} else {
		return []byte(f), nil
	}
}
func (c CustomJsonRawMessageEncoder) UnmarshalBinary(data []byte) (any, error) {
	m := json.RawMessage(data)
	return m, nil
}

func (c CustomJsonRawMessageEncoder) BinarySize(field any) (int, error) {
	if f, ok := field.(json.RawMessage); !ok {
		return 0, fmt.Errorf("expected json.RawMessage, got %T", field)
	} else {
		return len(f), nil
	}
}
func (c CustomJsonRawMessageEncoder) Encode(field any) ([]byte, error) {
	if f, ok := field.(json.RawMessage); !ok {
		return nil, fmt.Errorf("expected json.RawMessage, got %T", field)
	} else {
		return f,nil
	}
}

type CustomByteArrayEncoder struct {}

func (c CustomByteArrayEncoder) MarshalBinary(field any) ([]byte, error) {
	if f, ok := field.([]byte); !ok {
		return []byte{}, fmt.Errorf("expected []byte, got %T", field)
	} else {
		return f, nil
	}
}
func (c CustomByteArrayEncoder) UnmarshalBinary(data []byte) (any, error) {
	m := data
	return m, nil
}

func (c CustomByteArrayEncoder) BinarySize(field any) (int, error) {
	if f, ok := field.([]byte); !ok {
		return 0, fmt.Errorf("expected json.RawMessage, got %T", field)
	} else {
		return len(f), nil
	}
}
func (c CustomByteArrayEncoder) Encode(field any) ([]byte, error) {
	if f, ok := field.([]byte); !ok {
		return nil, fmt.Errorf("expected []byte, got %T", field)
	} else {
		return f,nil
	}
}

