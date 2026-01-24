package eviction

import "slices"

type FIFO struct {
	queue []string
}

func NewFIFO() *FIFO {
	return &FIFO{queue: make([]string, 0)}
}

func (f *FIFO) Access(key string) {
	f.queue = append(f.queue, key)
}

func (f *FIFO) Remove(key string) {
	// TODO: O(N) to remove -> too slow, replace with more efficient algorithm
	for idx, k := range f.queue {
		if k == key {
			f.queue = slices.Delete(f.queue, idx, idx+1)
			return
		}
	}
}

func (f *FIFO) SelectVictim() (string, bool) {
	if len(f.queue) == 0 {
		return "", false
	}
	victim := f.queue[0]
	f.queue = f.queue[1:]
	return victim, true
}
