package main

import (
	"testing"
)

func TestRandomBalancer(t *testing.T) {
	h1 := "192.168.1.1"
	h2 := "192.168.1.2"
	h3 := "192.168.1.3"
	b1 := NewBackend(h1, "80", 5)
	b2 := NewBackend(h2, "80", 1)
	b3 := NewBackend(h3, "80", 1)

	// no backends
	balancer := NewRdm()
	_, found := balancer.Select()
	if found {
		t.Error("no backend should found!")
	}

	balancer = NewRdm(b1, b2, b3)
	_, found = balancer.Select()
	if !found {
		t.Error("one backend should found!")
	}
}

func BenchmarkRdmSelect(b *testing.B) {
	h1 := "192.168.1.1"
	h2 := "192.168.1.2"
	h3 := "192.168.1.3"
	b1 := NewBackend(h1, "80", 5)
	b2 := NewBackend(h2, "80", 1)
	b3 := NewBackend(h3, "80", 1)

	// no backends
	balancer := NewRdm(b1, b2, b3)

	for i := 0; i < b.N; i++ {
		balancer.Select()
	}
}
