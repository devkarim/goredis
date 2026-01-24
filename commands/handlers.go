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
	err := shard.SetString(key, val)

	if err != nil {
		return resp.Value{Type: resp.RespError, Str: err.Error()}
	}

	return resp.Value{Type: resp.RespString, Str: "OK"}
}

func get(args []resp.Value) resp.Value {
	if len(args) != 1 {
		return resp.Value{Type: resp.RespError, Str: "ERR wrong number of arguments for 'get' command"}
	}

	key := args[0].Str

	shard := storage.GetShard(key)
	val, found, err := shard.GetString(key)

	if err != nil {
		return resp.Value{Type: resp.RespError, Str: err.Error()}
	}

	if !found {
		return resp.Value{Type: resp.RespNil}
	}

	return resp.Value{Type: resp.RespBulk, Str: val}
}

func hset(args []resp.Value) resp.Value {
	if len(args) != 3 {
		return resp.Value{Type: resp.RespError, Str: "ERR wrong number of arguments for 'hset' command"}
	}

	hash := args[0].Str
	key := args[1].Str
	val := args[2].Str

	shard := storage.GetShard(hash)
	err := shard.HSet(hash, key, val)

	if err != nil {
		return resp.Value{Type: resp.RespError, Str: err.Error()}
	}

	return resp.Value{Type: resp.RespString, Str: "OK"}
}

func hget(args []resp.Value) resp.Value {
	if len(args) != 2 {
		return resp.Value{Type: resp.RespError, Str: "ERR wrong number of arguments for 'hget' command"}
	}

	hash := args[0].Str
	key := args[1].Str

	shard := storage.GetShard(hash)
	val, found, err := shard.HGet(hash, key)

	if err != nil {
		return resp.Value{Type: resp.RespError, Str: err.Error()}
	}

	if !found {
		return resp.Value{Type: resp.RespNil}
	}

	return resp.Value{Type: resp.RespBulk, Str: val}
}

func hgetall(args []resp.Value) resp.Value {
	if len(args) != 1 {
		return resp.Value{Type: resp.RespError, Str: "ERR wrong number of arguments for 'hgetall' command"}
	}

	hash := args[0].Str

	shard := storage.GetShard(hash)
	arr, found, err := shard.HGetAll(hash)

	if err != nil {
		return resp.Value{Type: resp.RespError, Str: err.Error()}
	}

	if !found {
		return resp.Value{Type: resp.RespArray, Array: make([]resp.Value, 0)}
	}

	respArray := make([]resp.Value, len(arr))

	for idx, value := range arr {
		respArray[idx] = resp.Value{Type: resp.RespBulk, Str: value}
	}

	return resp.Value{Type: resp.RespArray, Array: respArray}
}
