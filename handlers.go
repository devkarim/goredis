package main

import (
	"github.com/devkarim/goredis/resp"
	"sync"
)

type RedisObjectType string

const (
	RedisObjectString RedisObjectType = "string"
	RedisObjectHash   RedisObjectType = "hash"
)

type RedisObject struct {
	Type RedisObjectType
	Str  string
	Hash map[string]string
}

var store = map[string]*RedisObject{}
var storeMutex = sync.RWMutex{}

var Handlers = map[string]func([]resp.Value) resp.Value{
	"PING":    ping,
	"SET":     set,
	"GET":     get,
	"HSET":    hset,
	"HGET":    hget,
	"HGETALL": hgetall,
}

func ping(args []resp.Value) resp.Value {
	if len(args) != 0 {
		return resp.Value{Type: resp.RespString, Str: args[0].Str}
	}

	return resp.Value{Type: resp.RespString, Str: "PONG"}
}

func set(args []resp.Value) resp.Value {
	if len(args) != 2 {
		return resp.Value{Type: resp.RespError, Str: "ERR wrong number of arguments for 'set' command"}
	}

	key := args[0].Str
	val := args[1].Str

	storeMutex.Lock()
	store[key] = &RedisObject{Type: RedisObjectString, Str: val}
	defer storeMutex.Unlock()

	return resp.Value{Type: resp.RespString, Str: "OK"}
}

func get(args []resp.Value) resp.Value {
	if len(args) != 1 {
		return resp.Value{Type: resp.RespError, Str: "ERR wrong number of arguments for 'get' command"}
	}

	key := args[0].Str

	storeMutex.RLock()
	obj, ok := store[key]
	defer storeMutex.RUnlock()

	if !ok {
		return resp.Value{Type: resp.RespNil}
	}

	if obj.Type != RedisObjectString {
		return resp.Value{Type: resp.RespError, Str: "WRONGTYPE Operation against a key holding the wrong kind of value"}
	}

	return resp.Value{Type: resp.RespBulk, Str: obj.Str}
}

func hset(args []resp.Value) resp.Value {
	if len(args) != 3 {
		return resp.Value{Type: resp.RespError, Str: "ERR wrong number of arguments for 'hset' command"}
	}

	hash := args[0].Str
	key := args[1].Str
	val := args[2].Str

	storeMutex.Lock()
	if _, ok := store[hash]; !ok {
		store[hash] = &RedisObject{Type: RedisObjectHash}
		store[hash].Hash = map[string]string{}
	}
	store[hash].Hash[key] = val
	defer storeMutex.Unlock()

	return resp.Value{Type: resp.RespString, Str: "OK"}
}

func hget(args []resp.Value) resp.Value {
	if len(args) != 2 {
		return resp.Value{Type: resp.RespError, Str: "ERR wrong number of arguments for 'hget' command"}
	}

	hash := args[0].Str
	key := args[1].Str

	storeMutex.RLock()
	obj, ok := store[hash]
	defer storeMutex.RUnlock()

	if !ok {
		return resp.Value{Type: resp.RespNil}
	}

	if obj.Type != RedisObjectHash {
		return resp.Value{Type: resp.RespError, Str: "WRONGTYPE Operation against a key holding the wrong kind of value"}
	}

	return resp.Value{Type: resp.RespBulk, Str: store[hash].Hash[key]}
}

func hgetall(args []resp.Value) resp.Value {
	if len(args) != 1 {
		return resp.Value{Type: resp.RespError, Str: "ERR wrong number of arguments for 'hgetall' command"}
	}

	hash := args[0].Str

	storeMutex.RLock()
	obj, ok := store[hash]
	defer storeMutex.RUnlock()

	if !ok {
		return resp.Value{Type: resp.RespNil}
	}

	if obj.Type != RedisObjectHash {
		return resp.Value{Type: resp.RespError, Str: "WRONGTYPE Operation against a key holding the wrong kind of value"}
	}

	idx := 0
	arr := make([]resp.Value, len(obj.Hash)*2)

	for key, value := range obj.Hash {
		arr[idx] = resp.Value{Type: resp.RespBulk, Str: key}
		arr[idx+1] = resp.Value{Type: resp.RespBulk, Str: value}
		idx += 2
	}

	return resp.Value{Type: resp.RespArray, Array: arr}
}
