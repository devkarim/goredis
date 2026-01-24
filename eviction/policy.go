package eviction

type Policy interface {
	Access(key string)

	Remove(key string)

	SelectVictim() (string, bool)
}
