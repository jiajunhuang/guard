package main

import (
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
