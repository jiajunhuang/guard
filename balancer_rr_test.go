package main

import (
	"testing"
)

func TestRRBalancer(t *testing.T) {
	h1 := "192.168.1.1"
	h2 := "192.168.1.2"
	h3 := "192.168.1.3"
	b1 := Backend{h1, 80, 5}
	b2 := Backend{h2, 80, 1}
	b3 := Backend{h3, 80, 1}

	// no backends
	balancer := NewRR()
	_, found := balancer.Select()
	if found {
		t.Error("no backend should found!")
	}

	balancer = NewRR(b1, b2, b3)
	_, found = balancer.Select()
	if !found {
		t.Error("one backend should found!")
	}
}

func BenchmarkRRSelect(b *testing.B) {
	h1 := "192.168.1.1"
	h2 := "192.168.1.2"
	h3 := "192.168.1.3"
	b1 := Backend{h1, 80, 5}
	b2 := Backend{h2, 80, 1}
	b3 := Backend{h3, 80, 1}

	// no backends
	balancer := NewRR(b1, b2, b3)

	for i := 0; i < b.N; i++ {
		balancer.Select()
	}
}
