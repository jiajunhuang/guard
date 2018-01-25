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
	n.addRoute([]byte("/user/hello"))

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

func TestQuery(t *testing.T) {
	n := &node{}
	n.addRoute([]byte("/user"))

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

func TestRefreshStatus(t *testing.T) {
	n := &node{}
	n.addRoute([]byte("/user/hello"), GET)

	status := n.status
	now := RightNow()
	n.status.key = now - 3*statusStep

	n.refreshStatus(now)
	if n.status == status {
		t.Errorf("n.status should refresh to %d, but it's %+v", now, n.status)
	}

	status = n.status
	if status.key != now {
		t.Errorf("brand new status's key should be %d, but status is: %+v, status.prev is: %+v, status.next is: %+v", now, status, status.prev, status.next)
	}
	if status.OK != 0 || status.TooManyRequests != 0 || status.InternalError != 0 || status.BadGateway != 0 {
		t.Errorf("brand new status's property should be reset, but it not: %+v", status)
	}
}

func TestRefreshStatusShouldNotRefresh(t *testing.T) {
	n := &node{}
	n.addRoute([]byte("/user/hello"), GET)

	now := RightNow()
	status := n.status
	status.key = now
	n.refreshStatus(now)
	if n.status != status {
		t.Errorf("n.status should not be refreshed, should be %p, but n is: %+v", status, n)
	}
}

// benchmark
func BenchmarkIncr(b *testing.B) {
	n := &node{}
	n.addRoute([]byte("/user/hello"))

	for i := 0; i < b.N; i++ {
		n.incr(http.StatusBadGateway)
	}
}

func BenchmarkQuery(b *testing.B) {
	n := &node{}
	n.addRoute([]byte("/user/hello"))

	for i := 0; i < 100; i++ {
		n.incr(http.StatusBadGateway)
	}

	for i := 0; i < b.N; i++ {
		n.query()
	}
}

func BenchmarkRightNow(b *testing.B) {
	for i := 0; i < b.N; i++ {
		RightNow()
	}
}
