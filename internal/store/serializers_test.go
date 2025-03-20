package store

import (
	"testing"
)

type TestStruct struct {
	Field1 string `json:"field1"`
	Field2 int    `json:"field2"`
}

func TestSerialize(t *testing.T) {
	value := TestStruct{
		Field1: "hello",
		Field2: 42,
	}

	serialized, err := Serialize(value)
	if err != nil {
		t.Fatalf("Serialize failed: %v", err)
	}

	if len(serialized) == 0 {
		t.Fatalf("Serialized data is empty")
	}
}

func TestDeserialize(t *testing.T) {
	value := TestStruct{
		Field1: "hello",
		Field2: 42,
	}

	serialized, err := Serialize(value)
	if err != nil {
		t.Fatalf("Serialize failed: %v", err)
	}

	var deserialized TestStruct
	err = Deserialize(serialized, &deserialized)
	if err != nil {
		t.Fatalf("Deserialize failed: %v", err)
	}

	if deserialized.Field1 != value.Field1 || deserialized.Field2 != value.Field2 {
		t.Fatalf("Deserialized data does not match original data")
	}
}

func TestSerializeDeserializeRoundTrip(t *testing.T) {
	value := TestStruct{
		Field1: "hello",
		Field2: 42,
	}

	serialized, err := Serialize(value)
	if err != nil {
		t.Fatalf("Serialize failed: %v", err)
	}

	var deserialized TestStruct
	err = Deserialize(serialized, &deserialized)
	if err != nil {
		t.Fatalf("Deserialize failed: %v", err)
	}

	if deserialized != value {
		t.Fatalf("Round trip failed: deserialized data does not match original data")
	}
}
