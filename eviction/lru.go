package eviction

import (
	"container/list"
	"sync"
)

type LRU struct {
	mu sync.Mutex
	list *list.List
	dict map[string]*list.Element
}

func NewLRU() *LRU {
	return &LRU{list: list.New(), dict: make(map[string]*list.Element)}
}

func (lru *LRU) Access(key string) {
	lru.mu.Lock()
	defer lru.mu.Unlock()

	el, ok := lru.dict[key]
	if ok {
		lru.list.MoveToFront(el)
		return
	}

	lru.dict[key] = lru.list.PushFront(key)
}

func (lru *LRU) Remove(key string) {
	lru.mu.Lock()
	defer lru.mu.Unlock()

	el, ok := lru.dict[key]
	if ok {
		lru.list.Remove(el)
		delete(lru.dict, key)
	}
}

func (lru *LRU) SelectVictim() (string, bool) {
	lru.mu.Lock()
	defer lru.mu.Unlock()

	if lru.list.Len() == 0 {
		return "", false
	}
	victim := lru.list.Back()
	key := victim.Value.(string)
	lru.list.Remove(victim)
	delete(lru.dict, key)
	return key, true
}

