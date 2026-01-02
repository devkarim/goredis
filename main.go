package main

import (
	"log"
)

func main() {
	s := NewServer(Config{})
	log.Fatal(s.Start())
}

