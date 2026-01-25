package eviction

import "testing"

func TestFIFO_SelectVictim_Order(t *testing.T) {
	fifo := NewFIFO()
	fifo.Access("a")
	fifo.Access("b")
	fifo.Access("c")
	fifo.Access("a")

	result, ok := fifo.SelectVictim()
	expected := "a"

	if !ok || result != expected {
		t.Errorf("expected %s, got %s", expected, result)
	}

	result, ok = fifo.SelectVictim()
	expected = "b"

	if !ok || result != expected {
		t.Errorf("expected %s, got %s", expected, result)
	}
}

func TestFIFO_SelectVictim_Empty(t *testing.T) {
	fifo := NewFIFO()

	_, ok := fifo.SelectVictim()
	expected := false

	if expected != ok {
		t.Errorf("expected ok to be %t, got %t", expected, ok)
	}
}

func TestFIFO_Remove(t *testing.T) {
	fifo := NewFIFO()
	fifo.Access("a")
	fifo.Access("b")
	fifo.Remove("a")

	result, ok := fifo.SelectVictim()
	expected := "b"

	if !ok || expected != result {
		t.Errorf("expected %s, got %s", expected, result)
	}
}

