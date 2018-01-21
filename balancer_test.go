package main

import (
	"reflect"
	"testing"
)

func TestToURL(t *testing.T) {
	h1 := "192.168.1.1"
	b1 := Backend{h1, 80, 5}

	if b1.ToURL() != "192.168.1.1:80" {
		t.Errorf("b1.ToURL should return `192.168.1.1:80` but got %s", b1.ToURL())
	}
}

func TestWRRFound(t *testing.T) {
	// weights of [5, 1, 1] should generate sequence of index: [1, 1, 2, 1, 3, 1, 1]
	h1 := "192.168.1.1"
	h2 := "192.168.1.2"
	h3 := "192.168.1.3"
	b1 := Backend{h1, 80, 5}
	b2 := Backend{h2, 80, 1}
	b3 := Backend{h3, 80, 1}

	wrr := NewWRR(b1, b2, b3)

	var r *Backend
	var f bool
	type expectResult struct {
		found    bool
		host     string
		shouldBe []int
	}
	expectResultList := []expectResult{
		expectResult{true, h1, []int{-2, 1, 1}},
		expectResult{true, h1, []int{-4, 2, 2}},
		expectResult{true, h2, []int{1, -4, 3}},
		expectResult{true, h1, []int{-1, -3, 4}},
		expectResult{true, h3, []int{4, -2, -2}},
		expectResult{true, h1, []int{2, -1, -1}},
		expectResult{true, h1, []int{0, 0, 0}},
	}

	for i, e := range expectResultList {
		r, f = wrr.Select()
		if f != e.found || r.Host != e.host {
			t.Errorf("the %dth select should found the %s, but got: %+v, %t", i, e.host, r.Host, f)
		}
		wrr.lock.Lock()
		if !reflect.DeepEqual(wrr.weights, e.shouldBe) {
			t.Errorf("wrr's weights should be %+v, but got: %+v", e.shouldBe, wrr.weights)
		}
		wrr.lock.Unlock()
	}
}

func TestWRRNotFound(t *testing.T) {
	h1 := "192.168.1.1"
	h2 := "192.168.1.2"
	h3 := "192.168.1.3"
	b1 := Backend{h1, 80, 0}
	b2 := Backend{h2, 80, 0}
	b3 := Backend{h3, 80, 0}

	wrr := NewWRR(b1, b2, b3)

	var r *Backend
	var f bool
	type expectResult struct {
		found    bool
		backend  *Backend
		shouldBe []int
	}
	expectResultList := []expectResult{
		expectResult{false, nil, []int{0, 0, 0}},
		expectResult{false, nil, []int{0, 0, 0}},
		expectResult{false, nil, []int{0, 0, 0}},
		expectResult{false, nil, []int{0, 0, 0}},
		expectResult{false, nil, []int{0, 0, 0}},
		expectResult{false, nil, []int{0, 0, 0}},
		expectResult{false, nil, []int{0, 0, 0}},
	}

	for i, e := range expectResultList {
		r, f = wrr.Select()
		if f != e.found || r != e.backend {
			t.Errorf("the %dth select should found the %+v, but got: %+v, %t", i, e.backend, r.Host, f)
		}
		wrr.lock.Lock()
		if !reflect.DeepEqual(wrr.weights, e.shouldBe) {
			t.Errorf("wrr's weights should be %+v, but got: %+v", e.shouldBe, wrr.weights)
		}
		wrr.lock.Unlock()
	}
}

func BenchmarkWRRSelect(b *testing.B) {
	h1 := "192.168.1.1"
	h2 := "192.168.1.2"
	h3 := "192.168.1.3"

	wrr := NewWRR(
		Backend{h1, 80, 5},
		Backend{h2, 80, 1},
		Backend{h3, 80, 1},
	)

	for i := 0; i < b.N; i++ {
		wrr.Select()
	}
}
