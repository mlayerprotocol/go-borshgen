package tests

import (
	"bytes"
	"encoding/json"
	"math"
	"reflect"
	"testing"

	"github.com/mlayerprotocol/go-borshgen/tests/configs"
)
      

func TestBinaryEncoding(t *testing.T) {
	// Create test data with various data types
	testID := ID(32232)
	 parentIDs := []ID{100, 200, 300}
	 optionalCounter := int32(42)
	optionalFlag := true
	optionalScore := 99.5
	 rawMessage := json.RawMessage(`{"key": "value"}`)
	b1 := [32]byte{
		1, 2, 3, 4, 5, 6, 7, 8,
		9, 10, 11, 12, 13, 14, 15, 16,
		17, 18, 19, 20, 21, 22, 23, 24,
		25, 26, 27, 28, 29, 30, 31, 32,
	}
	b2 := [32]byte{
		21, 22, 23, 24, 5, 6, 7, 8,
		9, 10, 111, 212, 13, 14, 15, 16,
		17, 38, 19, 210, 1, 2, 23, 24,
		5, 4, 4, 1, 29, 10, 31, 32,
	}
	chainId1 := configs.ChainId("c1")
	chainId2 := configs.ChainId("c2")
	// chainId3 := configs.ChainId("c3")
	
	original := Event{
		ID:        testID,
		 Parent:    &parentIDs,
		FixedSlice: [][32]byte{b1, b2 },
		Timestamp: math.MaxInt64,
		Path: &EventPath{
			ID:        456,
			Timestamp: 9876543210,
		},
		 Data:      []byte("test data content"),
		 Counter:   -123,
		 Flag:      true,
		Score:     3.14159,
		Rating                                          :    2.5,
		Systems:   []System{"auth", "logging", "metrics"},
		Tags:      []string{"urgent", "production", "critical"},
	
		
		Paths: []EventPath{
			{ID: 111, Timestamp: 1111},
			{ID: 222, Timestamp: 2222},
			{ID: 333, Timestamp: 3333},
		},
		EID:             []int{1, 2, 3, 4, 5},
		Versions:        []int32{10, 20, 30},
		Sizes:           []uint64{1000, 2000, 3000},
		OptionalCounter: &optionalCounter,
		OptionalFlag:    &optionalFlag,
		OptionalScore:   &optionalScore,
		Ignored:         "this should be ignored",
		JsonData: json.RawMessage(`{"key": "value"}`),
		 JsonPtrData: &rawMessage,
		JsonSliceData: []json.RawMessage{
			json.RawMessage(`{"slice_key1": "slice_value1"}`),
			json.RawMessage(`{"slice_key2": "slice_value2"}`),
		},
		JsonPointerSliceData: &[]json.RawMessage{
			json.RawMessage(`{"pointer_slice_key1": "pointer_slice_value1"}`),
			json.RawMessage(`{"pointer_slice_key2": "pointer_slice_value2"}`),
		},	 
		Chain: [][][]configs.ChainId{
	{
		{chainId1, chainId2},    // s[0][0]
		{"a3", "a4"},    // s[0][1]
	},
	{
		{"b1", "b2"},    // s[1][0]
		{"b3", "b4"},    // s[1][1]
	},
	{
		{"c1", "c2"},    // s[2][0]
	},
},
}


	t.Run("BinarySize", func(t *testing.T) {
		size, _ := original.BinarySize()
		if size <= 0 {
			t.Errorf("BinarySize() returned %d, expected positive value", size)
		}

		t.Logf("Binary size: %d bytes", size)
	})



	t.Run("MarshalBinary", func(t *testing.T) {
		data, err := original.MarshalBinary()
		if err != nil {
			t.Fatalf("MarshalBinary() failed: %v", err)
		}
	
		if len(data) == 0 {
			t.Error("MarshalBinary() returned empty data")
		}
		
		t.Logf("Marshaled data length: %d bytes", len(data))
		
		// Verify size estimation is reasonable
		estimatedSize, _ := original.BinarySize()
		if len(data) > estimatedSize*2 {
			t.Errorf("Marshaled data size %d is much larger than estimated size %d", len(data), estimatedSize)
		}
	})




	t.Run("UnmarshalBinary", func(t *testing.T) {
		// Marshal original
		data, err := original.MarshalBinary()
		if err != nil {
			t.Fatalf("MarshalBinary() failed: %v", err)
		}

		// Unmarshal into new struct
		var restored Event
		err = restored.UnmarshalBinary(data)
		if err != nil {
			t.Fatalf("UnmarshalBinary() failed: %v", err)
		}

		// Compare basic fields
		if restored.Timestamp != original.Timestamp {
			t.Errorf("Timestamp mismatch: got %d, want %d", restored.Timestamp, original.Timestamp)
		}
		
		if restored.Counter != original.Counter {
			t.Errorf("Counter mismatch: got %d, want %d", restored.Counter, original.Counter)
		}
		
		if restored.Flag != original.Flag {
			t.Errorf("Flag mismatch: got %t, want %t", restored.Flag, original.Flag)
		}
		
		if math.Abs(restored.Score-original.Score) > 1e-10 {
			t.Errorf("Score mismatch: got %f, want %f", restored.Score, original.Score)
		}
		
		if math.Abs(float64(restored.Rating-original.Rating)) > 1e-6 {
			t.Errorf("Rating mismatch: got %f, want %f", restored.Rating, original.Rating)
		}

		// Compare pointer fields
		if original.ID != 0 && restored.ID != 0 {
			if restored.ID != original.ID {
				t.Errorf("ID mismatch: got %d, want %d", restored.ID, original.ID)
			}
		} else if original.ID != restored.ID {
			t.Errorf("ID pointer mismatch: got %v, want %v", restored.ID, original.ID)
		}

		// Compare slice fields
		if (original.Data != nil && !bytes.Equal(restored.Data, original.Data)) || len(restored.Data) != len(original.Data) {
			t.Errorf("Data mismatch: got %v, want %v", restored.Data, original.Data)
		}
		
		if (original.Tags != nil && !reflect.DeepEqual(restored.Tags, original.Tags)) || len(restored.Tags) != len(original.Tags) {
			t.Errorf("Tags mismatch: got %+v==%v, want %+v==%v", restored.Tags, restored.Tags==nil, original.Tags, original.Tags==nil)
		}
		
		if (original.EID != nil && !reflect.DeepEqual(restored.EID, original.EID)) || len(restored.EID) != len(original.EID) {
			t.Errorf("EID mismatch: got %v==%v, want %v==%v", restored.EID, restored.EID == nil, original.EID, original.EID==nil)
		}

		if (original.Chain != nil && !reflect.DeepEqual(restored.Chain, original.Chain)) || len(restored.Chain) != len(original.Chain) {
			t.Errorf("Chain mismatch: got %+v==%v, want %+v==%v", restored.Chain, restored.Chain==nil, original.Chain, original.Chain==nil)
		}
		
		
		if (original.Versions != nil && !reflect.DeepEqual(restored.Versions, original.Versions))  || len(restored.Versions) != len(original.Versions) {
			t.Errorf("Versions mismatch: got %v, want %v", restored.Versions, original.Versions)
		}
		
		if len(restored.Sizes) != len(original.Sizes) || (original.Sizes != nil && !reflect.DeepEqual(restored.Sizes, original.Sizes)) {
			t.Errorf("Sizes mismatch: got %v, want %v", restored.Sizes, original.Sizes)
		}

		//Compare struct fields
		if restored.Path.ID != original.Path.ID {
			t.Errorf("Path.ID mismatch: got %d, want %d", restored.Path.ID, original.Path.ID)
		}
		
		if restored.Path.Timestamp != original.Path.Timestamp {
			t.Errorf("Path.Timestamp mismatch: got %d, want %d", restored.Path.Timestamp, original.Path.Timestamp)
		}
		
		if len(restored.Paths) != len(original.Paths) {
			t.Errorf("Paths length mismatch: got %d, want %d", len(restored.Paths), len(original.Paths))
		} else {
			for i, path := range restored.Paths {
				if path.ID != original.Paths[i].ID {
					t.Errorf("Paths[%d].ID mismatch: got %d, want %d", i, path.ID, original.Paths[i].ID)
				}
				if path.Timestamp != original.Paths[i].Timestamp {
					t.Errorf("Paths[%d].Timestamp mismatch: got %d, want %d", i, path.Timestamp, original.Paths[i].Timestamp)
				}
			}
		}

		// Compare optional pointer fields
		if original.OptionalCounter != nil && restored.OptionalCounter != nil {
			if *restored.OptionalCounter != *original.OptionalCounter {
				t.Errorf("OptionalCounter mismatch: got %d, want %d", *restored.OptionalCounter, *original.OptionalCounter)
			}
		} else if original.OptionalCounter != restored.OptionalCounter {
			t.Errorf("OptionalCounter pointer mismatch")
		}
		
		if original.OptionalFlag != nil && restored.OptionalFlag != nil {
			if *restored.OptionalFlag != *original.OptionalFlag {
				t.Errorf("OptionalFlag mismatch: got %t, want %t", *restored.OptionalFlag, *original.OptionalFlag)
			}
		} else if original.OptionalFlag != restored.OptionalFlag {
			t.Errorf("OptionalFlag pointer mismatch")
		}
		
		if original.OptionalScore != nil && restored.OptionalScore != nil {
			if math.Abs(*restored.OptionalScore-*original.OptionalScore) > 1e-10 {
				t.Errorf("OptionalScore mismatch: got %f, want %f", *restored.OptionalScore, *original.OptionalScore)
			}
		} else if original.OptionalScore != restored.OptionalScore {
			t.Errorf("OptionalScore pointer mismatch")
		}

		// Verify ignored field is not affected
		if restored.Ignored != "" {
			t.Errorf("Ignored field should be empty, got: %s", restored.Ignored)
		}
		// Verify ignored field is not affected
		if bytes.Equal(restored.JsonData, original.JsonData) == false {
			t.Errorf("Ignored field should be empty, got: %s", restored.Ignored)
		}

	})
	
	t.Run("Encode", func(t *testing.T) {
		encoded, err := original.Encode()
		if err != nil {
			t.Fatalf("Encode() failed: %v", err)
		}
		
		if len(encoded) == 0 {
			t.Error("Encode() returned empty data")
		}
		
		t.Logf("Encoded data length: %d bytes", len(encoded))
		
		// Encode should be deterministic - test multiple calls
		encoded2, err := original.Encode()
		if err != nil {
			t.Fatalf("Second Encode() failed: %v", err)
		}
		
		if !bytes.Equal(encoded, encoded2) {
			t.Error("Encode() is not deterministic - multiple calls returned different results")
		}
	})


	t.Run("RoundTrip", func(t *testing.T) {
		// Full round trip test
		data, err := original.MarshalBinary()
		if err != nil {
			t.Fatalf("MarshalBinary() failed: %v", err)
		}

		var restored Event
		err = restored.UnmarshalBinary(data)
		if err != nil {
			t.Fatalf("UnmarshalBinary() failed: %v", err)
		}

		// Marshal the restored struct
		data2, err := restored.MarshalBinary()
		if err != nil {
			t.Fatalf("Second MarshalBinary() failed: %v", err)
		}

		// Data should be identical after round trip
		if !bytes.Equal(data, data2) {
			t.Error("Round trip produced different binary data")
		}
	})
}


func TestBinaryEncodingEdgeCases(t *testing.T) {
	t.Run("EmptyStruct", func(t *testing.T) {
		var empty Event
		
		data, err := empty.MarshalBinary()
		if err != nil {
			t.Fatalf("MarshalBinary() on empty struct failed: %v", err)
		}
		t.Logf("MARSHALED %+v", data)
		var restored Event
		err = restored.UnmarshalBinary(data)
		if err != nil {
			t.Fatalf("UnmarshalBinary() on empty struct failed: %v", err)
		}
		
		// Verify basic equality for zero values
		if restored.Timestamp != empty.Timestamp {
			t.Errorf("Timestamp mismatch after empty struct round trip")
		}
	})


	t.Run("NilPointers", func(t *testing.T) {
		event := Event{
			Timestamp: 12345,
			Counter:   -999,
			Flag:      false,
			// All pointers nil
			ID:              0,
			Parent:          nil,
			OptionalCounter: nil,
			OptionalFlag:    nil,
			OptionalScore:   nil,
		}
		
		data, err := event.MarshalBinary()
		if err != nil {
			t.Fatalf("MarshalBinary() with nil pointers failed: %v", err)
		}
		
		var restored Event
		err = restored.UnmarshalBinary(data)
		if err != nil {
			t.Fatalf("UnmarshalBinary() with nil pointers failed: %v", err)
		}
		
		if restored.ID != 0 {
			t.Error("ID should be nil after round trip")
		}
		if restored.Parent != nil {
			t.Error("Parent should be nil after round trip")
		}
		if restored.OptionalCounter != nil {
			t.Error("OptionalCounter should be nil after round trip")
		}
		if restored.OptionalFlag != nil {
			t.Error("OptionalFlag should be nil after round trip")
		}
		if restored.OptionalScore != nil {
			t.Error("OptionalScore should be nil after round trip")
		}
	})


	t.Run("EmptySlices", func(t *testing.T) {
		event := Event{
			Timestamp: 12345,
			Systems:   []System{}, // Empty slice
			Tags:      []string{}, // Empty slice
			EID:       []int{},    // Empty slice
			Paths:     []EventPath{}, // Empty slice
		}
		
		data, err := event.MarshalBinary()
		if err != nil {
			t.Fatalf("MarshalBinary() with empty slices failed: %v", err)
		}
		
		var restored Event
		err = restored.UnmarshalBinary(data)
		if err != nil {
			t.Fatalf("UnmarshalBinary() with empty slices failed: %v", err)
		}
		
		if len(restored.Systems) != 0 {
			t.Errorf("Systems should be empty, got length %d", len(restored.Systems))
		}
		if len(restored.Tags) != 0 {
			t.Errorf("Tags should be empty, got length %d", len(restored.Tags))
		}
		if len(restored.EID) != 0 {
			t.Errorf("EID should be empty, got length %d", len(restored.EID))
		}
		if len(restored.Paths) != 0 {
			t.Errorf("Paths should be empty, got length %d", len(restored.Paths))
		}
	})
	t.Run("ExtremValues", func(t *testing.T) {
		event := Event{
			Timestamp: math.MaxUint64,
			Counter:   math.MinInt32,
			Score:     math.MaxFloat64,
			Rating:    math.MaxFloat32,
		}
		
		data, err := event.MarshalBinary()
		if err != nil {
			t.Fatalf("MarshalBinary() with extreme values failed: %v", err)
		}
		
		var restored Event
		err = (&restored).UnmarshalBinary(data)
		if err != nil {
			t.Fatalf("UnmarshalBinary() with extreme values failed: %v", err)
		}
		
		t.Logf("%+v", restored)
		if restored.Timestamp != math.MaxUint64 {
			t.Fatalf("Timestamp extreme value mismatch: got %d, want %d", restored.Timestamp, event.Timestamp)
		}
		if restored.Counter != math.MinInt32 {
			t.Errorf("Counter extreme value mismatch: got %d, want %d", restored.Counter, int32(math.MinInt32))
		}
		if restored.Score != math.MaxFloat64 {
			t.Errorf("Score extreme value mismatch: got %f, want %f", restored.Score, math.MaxFloat64)
		}
		if restored.Rating != math.MaxFloat32 {
			t.Errorf("Rating extreme value mismatch: got %f, want %f", restored.Rating, math.MaxFloat32)
		}
	})
}

func BenchmarkBinaryEncoding(b *testing.B) {
	// Create a reasonably sized test event
	testID := ID(239039203)
	parentIDs := []ID{100, 200, 300, 400, 500}
	
	event := Event{
		ID:        testID,
		Parent:    &parentIDs,
		Timestamp: 1234567890,
		Data:      make([]byte, 1024), // 1KB of data
		Counter:   42,
		Flag:      true,
		Score:     3.14159,
		Rating:    2.5,
		Systems:   []System{"auth", "logging", "metrics", "tracing", "monitoring"},
		Tags:      []string{"prod", "critical", "urgent", "high-priority", "monitored"},
		Path: &EventPath{
			ID:        456,
			Timestamp: 9876543210,
		},
		Paths: make([]EventPath, 10),
		EID:   make([]int, 100),
	}
	
	// Fill in some data
	for i := range event.Paths {
		event.Paths[i] = EventPath{ID: ID(i), Timestamp: uint64(i * 1000)}
	}
	for i := range event.EID {
		event.EID[i] = i
	}

	b.Run("BinarySize", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = event.BinarySize()
		}
	})

	b.Run("MarshalBinary", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := event.MarshalBinary()
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("UnmarshalBinary", func(b *testing.B) {
		data, err := event.MarshalBinary()
		if err != nil {
			b.Fatal(err)
		}
		
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
		restored := Event{}
			err := restored.UnmarshalBinary(data)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("Encode", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := event.Encode()
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("RoundTrip", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			data, err := event.MarshalBinary()
			if err != nil {
				b.Fatal(err)
			}
			var restored Event
			err = restored.UnmarshalBinary(data)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}