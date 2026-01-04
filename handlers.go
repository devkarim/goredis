package main

import "github.com/devkarim/goredis/resp"

var Handlers = map[string]func([]resp.Value) resp.Value {
	"ping": ping,
}

func ping([]resp.Value) resp.Value {
	res := resp.Value{Type: resp.RespString, Str: "PONG"}

	return res
}
