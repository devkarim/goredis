package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
)

var ErrInvalidPolicy = errors.New("invalid policy, must be one of: lru, fifo")

type AllowedPolicy string

func (p *AllowedPolicy) String() string {
	return string(*p)
}

func (p *AllowedPolicy) Set(value string) error {
	switch value {
	case "fifo", "lru":
		*p = AllowedPolicy(value)
		return nil
	default:
		return ErrInvalidPolicy
	}
}

func main() {
	var policy AllowedPolicy = "lru"
	port := flag.String("port", "6379", "Port to listen on")
	aofPath := flag.String("aof", "database.aof", "Path of the AOF file")
	maxMemory := flag.Int("maxmemory", 1e+8, "Max memory in bytes")
	flag.Var(&policy, "policy", "Eviction policy: lru, fifo")

	flag.Parse()

	s := NewServer(Config{
		ListenAddr: fmt.Sprintf(":%s", *port),
		AofPath:    *aofPath,
		MaxMemory:  *maxMemory,
		Policy:     policy.String(),
	})
	log.Fatal(s.Start())
}
