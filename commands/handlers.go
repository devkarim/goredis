package commands

import (
	"github.com/devkarim/goredis/resp"
	"github.com/devkarim/goredis/storage"
)

type Command struct {
	Handler func([]resp.Value) resp.Value
	IsWrite bool
}

var Registry = map[string]Command{
	"PING":    {Handler: ping, IsWrite: false},
	"SET":     {Handler: set, IsWrite: true},
	"GET":     {Handler: get, IsWrite: false},
	"HSET":    {Handler: hset, IsWrite: true},
	"HGET":    {Handler: hget, IsWrite: false},
	"HGETALL": {Handler: hgetall, IsWrite: false},
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

	shard := storage.GetShard(key)

	shard.Mu.Lock()
	shard.Store[key] = &storage.RedisObject{Type: storage.RedisObjectString, Str: val}
	shard.Mu.Unlock()

	return resp.Value{Type: resp.RespString, Str: "OK"}
}

func get(args []resp.Value) resp.Value {
	if len(args) != 1 {
		return resp.Value{Type: resp.RespError, Str: "ERR wrong number of arguments for 'get' command"}
	}

	key := args[0].Str

	shard := storage.GetShard(key)

	shard.Mu.RLock()
	obj, ok := shard.Store[key]
	shard.Mu.RUnlock()

	if !ok {
		return resp.Value{Type: resp.RespNil}
	}

	if obj.Type != storage.RedisObjectString {
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

	shard := storage.GetShard(hash)

	shard.Mu.Lock()
	defer shard.Mu.Unlock()

	if _, ok := shard.Store[hash]; !ok {
		shard.Store[hash] = &storage.RedisObject{Type: storage.RedisObjectHash}
		shard.Store[hash].Hash = map[string]string{}
	}
	if shard.Store[hash].Type != storage.RedisObjectHash {
		return resp.Value{Type: resp.RespError, Str: "WRONGTYPE Operation against a key holding the wrong kind of value"}
	}
	shard.Store[hash].Hash[key] = val

	return resp.Value{Type: resp.RespString, Str: "OK"}
}

func hget(args []resp.Value) resp.Value {
	if len(args) != 2 {
		return resp.Value{Type: resp.RespError, Str: "ERR wrong number of arguments for 'hget' command"}
	}

	hash := args[0].Str
	key := args[1].Str

	shard := storage.GetShard(hash)

	shard.Mu.RLock()
	defer shard.Mu.RUnlock()

	obj, ok := shard.Store[hash]

	if !ok {
		return resp.Value{Type: resp.RespNil}
	}

	if obj.Type != storage.RedisObjectHash {
		return resp.Value{Type: resp.RespError, Str: "WRONGTYPE Operation against a key holding the wrong kind of value"}
	}

	return resp.Value{Type: resp.RespBulk, Str: shard.Store[hash].Hash[key]}
}

func hgetall(args []resp.Value) resp.Value {
	if len(args) != 1 {
		return resp.Value{Type: resp.RespError, Str: "ERR wrong number of arguments for 'hgetall' command"}
	}

	hash := args[0].Str

	shard := storage.GetShard(hash)

	shard.Mu.RLock()
	defer shard.Mu.RUnlock()

	obj, ok := shard.Store[hash]

	if !ok {
		return resp.Value{Type: resp.RespNil}
	}

	if obj.Type != storage.RedisObjectHash {
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
