package main

import (
	"net"
	"net/http"
	"testing"
	"time"
)

func getGracefulListener(t *testing.T) net.Listener {
	ln, err := net.Listen("tcp", ":45678")
	if err != nil {
		t.Logf("error while listen at %s", err)
	}
	return newGracefulListener(ln, time.Second*10)
}

func TestGracefulListener(t *testing.T) {
	gln := getGracefulListener(t)
	gln.Close()

	gln.Close()
}

func TestGracefulListenerAddr(t *testing.T) {
	gln := getGracefulListener(t)
	defer gln.Close()

	gln.Addr()
}

func TestGracefulListen(t *testing.T) {
	gln := getGracefulListener(t)
	defer gln.Close()
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })
	go http.Serve(gln, handler)
	time.Sleep(time.Second)
}
