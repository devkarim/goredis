package storage

import (
	"testing"
)

func TestStringSize(t *testing.T) {
	obj := &RedisObject{Str: "Hello", Type: RedisObjectString}
	result := obj.Size()
	expected := 5

	if  result != expected {
		t.Errorf("obj.Size() = %d; want %d", result, expected)
	}
}

func TestHashSize(t *testing.T) {
	hash := make(map[string]string)
	hash["u1"] = "karim"
	hash["u2"] = "ahmed"

	obj := &RedisObject{Hash: hash, Type: RedisObjectHash}
	result := obj.Size()
	expected := 10

	if  result != expected {
		t.Errorf("obj.Size() = %d; want %d", result, expected)
	}
}
