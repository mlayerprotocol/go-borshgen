package tests

import (
	"encoding/json"
	"fmt"

	"github.com/mlayerprotocol/go-borshgen/tests/configs"
	"github.com/mlayerprotocol/go-borshgen/tests/constants"
)


var _FixedSliceEncoder = CustomFixedSliceEncoder{}
type CustomFixedSliceEncoder struct {
	CustomElementEncoder
}

func (c CustomFixedSliceEncoder) MarshalBinary(field any) ([]byte, error) {
	if v , ok := field.([][32]byte); ok {
		out := make([]byte,  len(v)*32)
		for _, item := range v {
			out = append(out, item[:]...)
		}
		return out, nil
	}
	if v , ok := field.([][64]byte); ok {
		out := make([]byte, len(v)*64)
		for _, item := range v {
			out = append(out, item[:]...)
		}
		return out, nil
	}
		return nil, fmt.Errorf("unsupported type: %T", field)
	
}

func (c CustomFixedSliceEncoder) UnmarshalBinary(data []byte) (any, error) {
	if len(data)%32 == 0 {
		n := len(data) / 32
		out := make([][32]byte, n)
		for i := 0; i < n; i++ {
			copy(out[i][:], data[i*32:(i+1)*32])
		}
		return out, nil
	} else if len(data)%64 == 0 {
		n := len(data) / 64
		out := make([][64]byte, n)
		for i := 0; i < n; i++ {
			copy(out[i][:], data[i*64:(i+1)*64])
		}
		return out, nil
	}
	return nil, fmt.Errorf("invalid input length: %d (not multiple of 32 or 64)", len(data))
}

func (c CustomFixedSliceEncoder) BinarySize(field any) (int, error) {
	
	if v, ok :=  field.([][32]uint8); ok {
		
		return len(v) * 32, nil
	}
	if v, ok :=  field.([][64]byte); ok {
		
		return len(v) * 64, nil
	}
	fmt.Println("Not a valid fixed slice type")
		return 0, fmt.Errorf("unsupported type: %T", field)
}

func (c CustomFixedSliceEncoder) Encode(field any) ([]byte, error) {
	bz, err := c.MarshalBinary(field)
	if err != nil {
			return nil, err
	}
	return bz, nil
}

type ID int64
type System string

//go:generate borshgen -tag=msg -fallback=json
type EventPath struct {
	ID        ID     `msg:"id,int64" enc:""`
	Timestamp uint64 `msg:"ts" enc:""`
}

//go:generate borshgen -tag=msg -fallback=json
type Event struct {
	// Basic types
	ID        ID `msg:"id" enc:""`
	EventType constants.EventType `msg:"type" enc:""`
	FixedSliceCustom [][32]byte  `msg:"fsc,_FixedSliceEncoder" enc:""`
	FixedSlice [][32]byte  `msg:"fs" enc:""`
	Chain  [][][]configs.ChainId `msg:"typep" enc:""`
	//EventTypePtr *constants.EventType `msg:"typep" enc:""`
	Parent    *[]ID   `msg:"parent" enc:"f"`
	Timestamp uint64  `msg:"ts" enc:""`
	Data      []byte  `msg:"data"`
	
	// Additional basic types for comprehensive testing
	Counter    int32   `msg:"counter" enc:""`
	Flag       bool    `msg:"flag" enc:""`
	Score      float64 `msg:"score" enc:""`
	Rating     float32 `msg:"rating" enc:""`
	
	// String and byte slices
	Systems   []System    `msg:"sys,[]string" enc:""`
	Tags      []string    `msg:"tags" enc:""`
	Checksums [][]byte    `msg:"checksums" enc:""`
	
	// Struct types
	Path  *EventPath   `msg:"path" enc:""`
	Paths []EventPath `msg:"paths" enc:""`
	
	// Integer slices
	EID      []int     `msg:"eid" enc:""`
	Versions []int32   `msg:"versions" enc:""`
	Sizes    []uint64  `msg:"sizes" enc:""`
	
	// Pointer types
	OptionalCounter *int32   `msg:"opt_counter" enc:""`
	OptionalFlag    *bool    `msg:"opt_flag" enc:""`
	OptionalScore   *float64 `msg:"opt_score" enc:""`
	JsonData   json.RawMessage `msg:"jsd" enc:""`
	JsonPtrData   *json.RawMessage `msg:"jsp" enc:""`
	JsonSliceData   []json.RawMessage `msg:"jss" enc:""`
	JsonPointerSliceData   *[]json.RawMessage `msg:"jps" enc:""`
	
	
	// Ignored field
	Ignored string `msg:"-"`
}