package main

import (
	"reflect"
	"testing"
)

func TestWRRFound(t *testing.T) {
	// weights of [5, 1, 1] should generate sequence of index: [1, 1, 2, 1, 3, 1, 1]
	h1 := "192.168.1.1:80"
	h2 := "192.168.1.2:80"
	h3 := "192.168.1.3:80"
	b1 := NewBackend(h1, 5)
	b2 := NewBackend(h2, 1)
	b3 := NewBackend(h3, 1)

	wrr := NewWRR(b1)
	_, found := wrr.Select()
	if !found {
		t.Errorf("one backend should found")
	}

	wrr = NewWRR(b1, b2, b3)

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
		if f != e.found || r.URL != e.host {
			t.Errorf("the %dth select should found the %s, but got: %+v, %t", i, e.host, r.URL, f)
		}
		wrr.lock.Lock()
		if !reflect.DeepEqual(wrr.weights, e.shouldBe) {
			t.Errorf("wrr's weights should be %+v, but got: %+v", e.shouldBe, wrr.weights)
		}
		wrr.lock.Unlock()
	}
}

func TestWRRNotFound(t *testing.T) {
	h1 := "192.168.1.1:80"
	h2 := "192.168.1.2:80"
	h3 := "192.168.1.3:80"
	b1 := NewBackend(h1, 0)
	b2 := NewBackend(h2, 0)
	b3 := NewBackend(h3, 0)

	// no backends
	wrr := NewWRR()
	_, found := wrr.Select()
	if found {
		t.Errorf("no backend should found")
	}

	// backends with weight 0
	wrr = NewWRR(b1, b2, b3)

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
			t.Errorf("the %dth select should found the %+v, but got: %+v, %t", i, e.backend, r.URL, f)
		}
		wrr.lock.Lock()
		if !reflect.DeepEqual(wrr.weights, e.shouldBe) {
			t.Errorf("wrr's weights should be %+v, but got: %+v", e.shouldBe, wrr.weights)
		}
		wrr.lock.Unlock()
	}
}

func BenchmarkWRRSelect(b *testing.B) {
	h1 := "192.168.1.1:80"
	h2 := "192.168.1.2:80"
	h3 := "192.168.1.3:80"
	b1 := NewBackend(h1, 5)
	b2 := NewBackend(h2, 1)
	b3 := NewBackend(h3, 1)

	wrr := NewWRR(b1, b2, b3)

	for i := 0; i < b.N; i++ {
		wrr.Select()
	}
}
