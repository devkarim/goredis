package storage

import (
	"hash/fnv"
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

type Shard struct {
	Mu    sync.RWMutex
	Store map[string]*RedisObject
}

var shards []*Shard

func init() {
	shards = make([]*Shard, 256)

	for i := 0; i < len(shards); i++ {
		shards[i] = &Shard{}
		shards[i].Mu = sync.RWMutex{}
		shards[i].Store = map[string]*RedisObject{}
	}
}

func GetShard(key string) *Shard {
	h := fnv.New32a()
	h.Write([]byte(key))
	return shards[h.Sum32()%uint32(len(shards))]
}
