package storage

import (
	"testing"

	"github.com/devkarim/goredis/eviction"
)

func TestRedisObject_StringSize(t *testing.T) {
	obj := &RedisObject{Str: "Hello", Type: RedisObjectString}
	result := obj.Size()
	expected := 5

	if result != expected {
		t.Errorf("obj.Size() = %d; want %d", result, expected)
	}
}

func TestRedisObject_HashSize(t *testing.T) {
	hash := make(map[string]string)
	hash["u1"] = "karim"
	hash["u2"] = "ahmed"

	obj := &RedisObject{Hash: hash, Type: RedisObjectHash}
	result := obj.Size()
	expected := 14

	if result != expected {
		t.Errorf("obj.Size() = %d; want %d", result, expected)
	}
}

func TestEviction_FIFO(t *testing.T) {
	policy := eviction.NewFIFO()
	shard := &Shard{Store: make(map[string]*RedisObject), Policy: policy, MaxMemory: 10}

	shard.SetString("a", "12345")
	shard.SetString("b", "12345")
	shard.SetString("c", "12345")

	if _, ok := shard.Store["a"]; ok {
		t.Errorf("'a' should have been evicted")
	}

	if _, ok := shard.Store["c"]; !ok {
		t.Errorf("'c' should exist")
	}
}
