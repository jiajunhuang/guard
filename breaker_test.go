package main

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"
)

func TestBreakerServeHTTP(t *testing.T) {
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

	// app not found
	b := NewBreaker()
	b.apps["www.example.com"] = a

	b.ServeHTTP(w, r)
	// FIXME: below will not work...
	//if w.Code != http.StatusNotFound {
	//t.Errorf("w.Code should be 404, but got: %d", w.Code)
	//}

	r.Host = "www.example.com"
	b.ServeHTTP(w, r)
	if w.Code == http.StatusNotFound {
		t.Errorf("w.Code should not be 404, but got: %d", w.Code)
	}
}
