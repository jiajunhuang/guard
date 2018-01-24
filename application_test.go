package main

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"
)

func TestAddRouteWorks(t *testing.T) {
	a := NewApp(NewRdm(), true)

	a.AddRoute("/user/:name", "GET", "POST", "PUT", "DELETE", "HEAD", "OPTIONS", "CONNECT", "TRACE", "PATCH")
}

func TestAddRouteBadMethod(t *testing.T) {
	defer shouldPanic()

	a := NewApp(NewRdm(), true)

	a.AddRoute("/user/:name", "BABY")
}

func TestAddRouteNoMethod(t *testing.T) {
	defer shouldPanic()

	a := NewApp(NewRdm(), true)

	a.AddRoute("/user/:name")
}

func TestApplicationNilTree(t *testing.T) {
	defer shouldPanic()

	fakeServer := httptest.NewServer(http.HandlerFunc(fakeHandler))
	defer fakeServer.Close()
	u, err := url.ParseRequestURI(fakeServer.URL)
	if err != nil {
		t.Errorf("failed to parse fakeServer address: %s", fakeServer.URL)
	}
	p, _ := strconv.Atoi(u.Port())

	fakeBackend.Host = u.Host
	fakeBackend.Port = p
	fakeBackend.Weight = 0

	fb := fakeBalancer{}
	r, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	a := NewApp(fb, true)
	a.root = nil

	a.ServeHTTP(w, r)
}

func TestApplicationNilBalancer(t *testing.T) {
	defer shouldPanic()

	fakeServer := httptest.NewServer(http.HandlerFunc(fakeHandler))
	defer fakeServer.Close()
	u, err := url.ParseRequestURI(fakeServer.URL)
	if err != nil {
		t.Errorf("failed to parse fakeServer address: %s", fakeServer.URL)
	}
	p, _ := strconv.Atoi(u.Port())

	fakeBackend.Host = u.Host
	fakeBackend.Port = p
	fakeBackend.Weight = 0

	r, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	a := NewApp(nil, true)

	a.ServeHTTP(w, r)
}

func TestApplicationRedirect(t *testing.T) {
	fakeServer := httptest.NewServer(http.HandlerFunc(fakeHandler))
	defer fakeServer.Close()
	u, err := url.ParseRequestURI(fakeServer.URL)
	if err != nil {
		t.Errorf("failed to parse fakeServer address: %s", fakeServer.URL)
	}
	p, _ := strconv.Atoi(u.Port())

	fakeBackend.Host = u.Host
	fakeBackend.Port = p
	fakeBackend.Weight = 1

	fb := fakeBalancer{}
	w := httptest.NewRecorder()

	a := NewApp(fb, true)
	r, _ := http.NewRequest("POST", "/user/jhon/", nil)
	a.AddRoute("/user/jhon", "POST")
	a.AddRoute("/user/jhon/card/", "POST")

	a.ServeHTTP(w, r)

	// redirect
	if w.Code != http.StatusTemporaryRedirect {
		t.Errorf("w.Code should be %d but got: %d", http.StatusTemporaryRedirect, w.Code)
	}
	r, _ = http.NewRequest("POST", "/user/jhon/card", nil)
	if w.Code != http.StatusTemporaryRedirect {
		t.Errorf("w.Code should be %d but got: %d", http.StatusTemporaryRedirect, w.Code)
	}

	// test not found
	r, _ = http.NewRequest("POST", "/user/what/", nil)
	w = httptest.NewRecorder()

	a.ServeHTTP(w, r)

	if w.Code != http.StatusNotFound {
		t.Errorf("w.Code should be %d but got: %d", http.StatusNotFound, w.Code)
	}

	// test method not allowed
	r, _ = http.NewRequest("GET", "/user/jhon", nil)
	w = httptest.NewRecorder()

	a.ServeHTTP(w, r)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("w.Code should be %d but got: %d", http.StatusMethodNotAllowed, w.Code)
	}
}

func TestApplicationCircuit(t *testing.T) {
	fakeServer := httptest.NewServer(http.HandlerFunc(fakeHandler))
	defer fakeServer.Close()
	u, err := url.ParseRequestURI(fakeServer.URL)
	if err != nil {
		t.Errorf("failed to parse fakeServer address: %s", fakeServer.URL)
	}
	p, _ := strconv.Atoi(u.Port())

	fakeBackend.Host = u.Host
	fakeBackend.Port = p
	fakeBackend.Weight = 0

	fb := fakeBalancer{}
	w := httptest.NewRecorder()

	a := NewApp(fb, true)
	r, _ := http.NewRequest("POST", "/user/jhon", nil)
	a.AddRoute("/user/jhon", "POST")

	// circuit is not on
	a.ServeHTTP(w, r)

	// circuit is on
	for i := 0; i < 100; i++ {
		a.root.incr(http.StatusBadGateway)
	}
	a.ServeHTTP(w, r)
}
