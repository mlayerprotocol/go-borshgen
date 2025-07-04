package tests

import (
	"encoding/json"

	"github.com/mlayerprotocol/go-borshgen/tests/constants"
)

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
	//EventType constants.EventType `msg:"type" enc:""`
	EventTypePtr *constants.EventType `msg:"typep" enc:""`
	Parent    *[]ID   `msg:"parent,[]int64" enc:"f"`
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
	Path  EventPath   `msg:"path" enc:""`
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