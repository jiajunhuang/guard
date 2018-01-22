package main

import (
	"testing"
)

func TestToURL(t *testing.T) {
	h1 := "192.168.1.1"
	b1 := Backend{h1, 80, 5}

	if b1.ToURL() != "192.168.1.1:80" {
		t.Errorf("b1.ToURL should return `192.168.1.1:80` but got %s", b1.ToURL())
	}
}
