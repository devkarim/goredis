package eviction

type Policy interface {
	Access(key string)

	Remove(key string)

	SelectVictim() (string, bool)
}

func NewPolicy(policy string) Policy {
	switch policy {
	case "lru":
		return NewLRU()
	case "fifo":
		return NewFIFO()
	}
	return nil
}
