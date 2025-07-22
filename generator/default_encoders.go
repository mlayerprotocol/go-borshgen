package generator

import (
	"encoding/json"
	"fmt"
)

type BinaryMarshaler interface {
	MarshalBorsh() ([]byte, error)
	BinarySize() (int, error)
}

type BinaryUnmarshaler interface {
	UnmarshalBorsh(data []byte) error
}
type BinaryEncoder interface {
Encode() ([]byte, error)
}

type BorshEncoder interface {
	BinaryEncoder
	BinaryMarshaler
	BinaryUnmarshaler
}


type CustomElementEncoder interface {
	MarshalBorsh(field any) ([]byte, error)
	UnmarshalBorsh(data []byte) (any, error)
	BinarySize(field any) (int, error)
	Encode(field any) ([]byte, error)
	
}
var _DefaultJsonRawMessageEncoder = DefaultJsonRawMessageEncoder{}
var _DefaultByteArrayEncoder = DefaultByteArrayEncoder{}

type DefaultJsonRawMessageEncoder struct {}

func (c DefaultJsonRawMessageEncoder) MarshalBorsh(field any, parentStruct any) ([]byte, error) {
	if field == nil {
			return []byte{}, nil
		}
	if f, ok := field.(json.RawMessage); !ok {
		return []byte{}, fmt.Errorf("expected json.RawMessage, got %T", field)
	} else {
		return []byte(f), nil
	}
}
func (c DefaultJsonRawMessageEncoder) UnmarshalBorsh(data []byte) (any, error) {
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

func (c DefaultByteArrayEncoder) MarshalBorsh(field any, parentStruct any) ([]byte, error) {
	if field == nil {
			return []byte{}, nil
		}
	if f, ok := field.([]byte); !ok {
		return []byte{}, fmt.Errorf("expected []byte, got %T", field)
	} else {
		return f, nil
	}
}
func (c DefaultByteArrayEncoder) UnmarshalBorsh(data []byte) (any, error) {
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

