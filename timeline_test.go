package main

import (
	"net/http"
	"testing"
)

func TestStatusRing(t *testing.T) {
	ring := StatusRing()
	cursor := ring

	for i := 0; i < maxStatusLen; i++ {
		cursor = cursor.next
	}

	if cursor != ring {
		t.Errorf("after loop, cursor should equal to ring(%+v), but got: %+v", ring, cursor)
	}
}

func TestIncr(t *testing.T) {
	n := &node{}
	n.addRoute("/user/hello")

	n.incr(http.StatusOK)
	n.incr(http.StatusTooManyRequests)
	n.incr(http.StatusInternalServerError)
	n.incr(http.StatusBadGateway)
}

func TestBadIncrStatusIsNil(t *testing.T) {
	defer shouldPanic()

	n := &node{}
	n.incr(http.StatusOK)
}

func TestIncrBadStatus(t *testing.T) {
	defer shouldPanic()

	n := &node{}
	n.addRoute("/user/hello")

	n.incr(http.StatusBadRequest)
}

func TestQuery(t *testing.T) {
	n := &node{}
	n.addRoute("/user")

	ok, too, internal, bad, _ := n.query()

	if ok != 0 {
		t.Errorf("ok should be 0, but it is %d", ok)
	}
	if too != 0 {
		t.Errorf("too should be 0, but it is %d", too)
	}
	if internal != 0 {
		t.Errorf("internal should be 0, but it is %d", internal)
	}
	if bad != 0 {
		t.Errorf("bad should be 0, but it is %d", bad)
	}

	for i := 0; i < 100; i++ {
		n.incr(http.StatusInternalServerError)
		n.incr(http.StatusBadGateway)
		n.incr(http.StatusTooManyRequests)
	}

	ok, too, internal, bad, ratio := n.query()

	if ok != 0 {
		t.Errorf("ok should be 0, but it is %d", ok)
	}
	if too != 100 {
		t.Errorf("too should be 100, but it is %d", too)
	}
	if internal != 100 {
		t.Errorf("internal should be 100, but it is %d", internal)
	}
	if bad != 100 {
		t.Errorf("bad should be 100, but it is %d", bad)
	}

	if ratio < 0.75 {
		t.Errorf("ratio should at least 0.75, but it is %f", ratio)
	}
}

func TestQueryNilStatus(t *testing.T) {
	defer shouldPanic()

	n := &node{}

	n.query()
}

// benchmark
func BenchmarkIncr(b *testing.B) {
	n := &node{}
	n.addRoute("/user/hello")

	for i := 0; i < b.N; i++ {
		n.incr(http.StatusBadGateway)
	}
}

func BenchmarkQuery(b *testing.B) {
	n := &node{}
	n.addRoute("/user/hello")

	for i := 0; i < 100; i++ {
		n.incr(http.StatusBadGateway)
	}

	for i := 0; i < b.N; i++ {
		n.query()
	}
}
