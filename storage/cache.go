package storage

import (
	"errors"
	"hash/fnv"
	"sync"

	"github.com/devkarim/goredis/eviction"
)

var ErrWrongType = errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")

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

func (r *RedisObject) Size() int {
	if r.Type == RedisObjectString {
		return len(r.Str)
	}

	size := 0
	for k, v := range r.Hash {
		size += len(k) + len(v)
	}
	return size
}

type Shard struct {
	Mu            sync.RWMutex
	Store         map[string]*RedisObject
	Policy        eviction.Policy
	CurrentMemory int // in bytes
	MaxMemory     int // in bytes
}

var shards []*Shard

func Setup(policy eviction.Policy, maxMemory int) {
	shards = make([]*Shard, 256)

	for i := 0; i < len(shards); i++ {
		shards[i] = &Shard{}
		shards[i].Mu = sync.RWMutex{}
		shards[i].Store = map[string]*RedisObject{}
		shards[i].Policy = policy
		shards[i].MaxMemory = maxMemory
	}
}

func GetShard(key string) *Shard {
	h := fnv.New32a()
	h.Write([]byte(key))
	return shards[h.Sum32()%uint32(len(shards))]
}

func (s *Shard) evict(neededSize int) {
	for s.CurrentMemory+neededSize > s.MaxMemory {
		victim, ok := s.Policy.SelectVictim()
		if ok {
			victimVal := s.Store[victim]
			delete(s.Store, victim)
			if victimVal != nil {
				s.CurrentMemory -= victimVal.Size()
			}
		}
		if !ok {
			break
		}
	}
}

func (s *Shard) SetString(key string, val string) error {
	s.Mu.Lock()
	defer s.Mu.Unlock()

	oldVal, ok := s.Store[key]

	if ok && s.Store[key].Type != RedisObjectString {
		return ErrWrongType
	}

	if ok {
		s.CurrentMemory -= oldVal.Size()
	}

	newObj := &RedisObject{Type: RedisObjectString, Str: val}

	s.evict(newObj.Size())
	s.Policy.Access(key)

	s.Store[key] = newObj
	s.CurrentMemory += newObj.Size()

	return nil
}

func (s *Shard) GetString(key string) (string, bool, error) {
	s.Mu.RLock()
	defer s.Mu.RUnlock()

	obj, ok := s.Store[key]
	if !ok {
		return "", false, nil
	}

	if obj.Type != RedisObjectString {
		return "", false, ErrWrongType
	}

	s.Policy.Access(key)

	return obj.Str, true, nil
}

func (s *Shard) HSet(hash, key, val string) error {
	s.Mu.Lock()
	defer s.Mu.Unlock()

	obj, ok := s.Store[hash]

	if ok && s.Store[hash].Type != RedisObjectHash {
		return ErrWrongType
	}

	if ok {
		s.CurrentMemory -= obj.Size()
	}

	neededMemory := len(key) + len(val)
	if ok {
		neededMemory += obj.Size()
	}

	s.evict(neededMemory)
	s.Policy.Access(hash)

	if !ok {
		obj = &RedisObject{Type: RedisObjectHash, Hash: map[string]string{}}
		s.Store[hash] = obj
	}

	obj.Hash[key] = val
	s.CurrentMemory += obj.Size()

	return nil
}

func (s *Shard) HGet(hash, key string) (string, bool, error) {
	s.Mu.RLock()
	defer s.Mu.RUnlock()

	obj, ok := s.Store[hash]
	if !ok {
		return "", false, nil
	}

	if obj.Type != RedisObjectHash {
		return "", false, ErrWrongType
	}

	s.Policy.Access(hash)

	val, ok := obj.Hash[key]
	if !ok {
		return "", false, nil
	}

	return val, true, nil
}

func (s *Shard) HGetAll(hash string) ([]string, bool, error) {
	s.Mu.RLock()
	defer s.Mu.RUnlock()

	obj, ok := s.Store[hash]

	if !ok {
		return []string{}, false, nil
	}

	if obj.Type != RedisObjectHash {
		return []string{}, false, ErrWrongType
	}

	arr := make([]string, len(obj.Hash)*2)

	s.Policy.Access(hash)

	idx := 0

	for key, value := range obj.Hash {
		arr[idx] = key
		arr[idx+1] = value
		idx += 2
	}

	return arr, true, nil
}
