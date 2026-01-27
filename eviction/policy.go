package eviction

import "errors"

var ErrInvalidPolicyType = errors.New("invalid policy, must be one of: lru, fifo")

type PolicyType string

const (
	PolicyLRU  PolicyType = "lru"
	PolicyFIFO PolicyType = "fifo"
)

func (p PolicyType) String() string {
	return string(p)
}

func (p *PolicyType) Set(value string) error {
	switch value {
	case string(PolicyLRU), string(PolicyFIFO):
		*p = PolicyType(value)
		return nil
	default:
		return ErrInvalidPolicyType
	}
}

func (p PolicyType) NewPolicy() Policy {
	switch p {
	case PolicyLRU:
		return NewLRU()
	case PolicyFIFO:
		return NewFIFO()
	}
	return nil
}

type Policy interface {
	Access(key string)

	Remove(key string)

	SelectVictim() (string, bool)
}
