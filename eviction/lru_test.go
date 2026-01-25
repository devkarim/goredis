package eviction

import "testing"

func TestLRU_SelectVictim_Order(t *testing.T) {
	lru := NewLRU()
	lru.Access("a")
	lru.Access("b")
	lru.Access("c")
	lru.Access("a")

	victim, ok := lru.SelectVictim()
	expected := "b"

	if !ok || victim != expected {
		t.Errorf("expected %s, got %s", expected, victim)
	}

	victim, ok = lru.SelectVictim()
	expected = "c"

	if !ok || victim != expected {
		t.Errorf("expected %s, got %s", expected, victim)
	}
}

func TestLRU_SelectVictim_Empty(t *testing.T) {
	lru := NewLRU()

	_, ok := lru.SelectVictim()
	expected := false

	if ok != expected {
		t.Errorf("expected ok to be %t, got %t", expected, ok)
	}
}

func TestLRU_Remove(t *testing.T) {
	lru := NewLRU()
	lru.Access("a")
	lru.Access("b")
	lru.Access("c")
	lru.Remove("a")

	victim, ok := lru.SelectVictim()
	expected := "b"

	if !ok || victim != expected {
		t.Errorf("expected %s, got %s", expected, victim)
	}
}

