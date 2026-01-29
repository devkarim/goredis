package storage

import (
	"errors"
	"hash/fnv"
	"log/slog"
	"sync"
	"time"

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
	Id            int
	Mu            sync.RWMutex
	Store         map[string]*RedisObject
	Expire        map[string]time.Time
	Policy        eviction.Policy
	CurrentMemory int // in bytes
	MaxMemory     int // in bytes
}

var shards []*Shard

func Setup(policy eviction.Policy, maxMemory int) {
	shards = make([]*Shard, 8)

	for i := 0; i < len(shards); i++ {
		shards[i] = &Shard{
			Id: i,
			MaxMemory: maxMemory,
			Policy: policy,
			Mu: sync.RWMutex{},
			Store: make(map[string]*RedisObject),
			Expire: make(map[string]time.Time),
		}
	}

	startJanitor()
}

func startJanitor() {
	go func() {
		for {
			time.Sleep(100 * time.Millisecond)
			for _, shard := range shards {
				shard.cleanExpiredKeys()
			}
		}
	}()
}

func GetShard(key string) *Shard {
	h := fnv.New32a()
	h.Write([]byte(key))
	return shards[h.Sum32()%uint32(len(shards))]
}

func (s *Shard) cleanExpiredKeys() {
	slog.Debug("Cleaning expired keys...", )

	s.Mu.Lock()
	defer s.Mu.Unlock()

	for {
		deleted := 0
		sampleSize := 20

		for key, expiry := range s.Expire {
			// is expired?
			if time.Now().After(expiry) {
				s.Delete(key)
				deleted++

				slog.Info("Janitor", "deleted", key, "expiry", expiry, "now", time.Now(), "total", deleted)
			}

			sampleSize--
			if sampleSize <= 0 {
				break
			}
		}

		if deleted < 5 {
			break
		}
	}
}

func (s *Shard) evict(neededSize int) {
	slog.Debug("Evicting policy", "shard", s.Id, "currentMemory", s.CurrentMemory, "neededSize", neededSize, "maxMemory", s.MaxMemory)
	for s.CurrentMemory+neededSize > s.MaxMemory {
		victim, ok := s.Policy.SelectVictim()
		if ok {
			slog.Info("Evicting policy", "evicted", victim)
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
		s.Policy.Remove(key)
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

	if s.isExpired(key) {
		return "", false, nil
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

	s.Policy.Access(hash)
	s.evict(neededMemory)

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

	if s.isExpired(hash) {
		return "", false, nil
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

	if s.isExpired(hash) {
		return []string{}, false, nil
	}

	s.Policy.Access(hash)

	arr := make([]string, len(obj.Hash)*2)
	idx := 0
	for key, value := range obj.Hash {
		arr[idx] = key
		arr[idx+1] = value
		idx += 2
	}

	return arr, true, nil
}

func (s *Shard) isExpired(key string) bool {
	s.Mu.Lock()
	defer s.Mu.Unlock()

	expiry, exists := s.Expire[key]
	if !exists || time.Now().Before(expiry) {
		return false
	}
	// The key is expired
	s.Delete(key)

	return true
}

func (s *Shard) Delete(key string) bool {
	delete(s.Store, key)
	delete(s.Expire, key)

	s.Policy.Remove(key)

	return true
}

func (s *Shard) SetExpire(key string, deadline time.Time) bool {
	s.Mu.Lock()
	defer s.Mu.Unlock()

	if _, exists := s.Store[key]; !exists {
		return false
	}

	s.Expire[key] = deadline
	return true
}

