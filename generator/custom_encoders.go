package generator

import (
	"encoding/json"
	"fmt"
)
type CustomEncoder interface {
	MarshalBinary(field any) ([]byte, error)
	UnmarshalBinary(data []byte, field any) error
	BinarySize(field any) []byte
	Encode(field any) []byte
	
}
var _CustomJsonRawMessageEncoder = CustomJsonRawMessageEncoder{}

type CustomJsonRawMessageEncoder struct {}

func (c CustomJsonRawMessageEncoder) MarshalBinary(field any) ([]byte, error) {
	if f, ok := field.(json.RawMessage); !ok {
		return []byte{}, fmt.Errorf("expected json.RawMessage, got %T", field)
	} else {
		return f, nil
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