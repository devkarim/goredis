package main

import (
	"log"

	"github.com/devkarim/goredis/core"
)

func main() {
	cfg, err := core.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}
	s := NewServer(cfg)
	log.Fatal(s.Start())
}
