package eviction

import (
	"container/list"
)

type FIFO struct {
	list *list.List
	dict map[string]*list.Element
}

func NewFIFO() *FIFO {
	return &FIFO{list: list.New(), dict: make(map[string]*list.Element)}
}

func (f *FIFO) Access(key string) {
	_, ok := f.dict[key]
	if ok {
		return
	}
	f.dict[key] = f.list.PushFront(key)
}

func (f *FIFO) Remove(key string) {
	el, ok := f.dict[key]
	if !ok {
		return
	}
	f.list.Remove(el)
	delete(f.dict, key)
}

func (f *FIFO) SelectVictim() (string, bool) {
	if f.list.Len() == 0 {
		return "", false
	}
	victim := f.list.Back()
	key := victim.Value.(string)
	f.list.Remove(victim)
	delete(f.dict, key)
	return key, true
}
