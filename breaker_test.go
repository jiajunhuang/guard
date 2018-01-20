package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"
)

func ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "backend!")
}

func TestBreakerServeHTTP(t *testing.T) {
	var app = "www.example.com"
	fakeServer := httptest.NewServer(http.HandlerFunc(ServeHTTP))
	defer fakeServer.Close()
	u, err := url.ParseRequestURI(fakeServer.URL)
	if err != nil {
		t.Errorf("failed to parse fakeServer address: %s", fakeServer.URL)
	}
	p, _ := strconv.Atoi(u.Port())
	var upstreamConfig = APP{app, []string{"/"}, []string{"GET"}, []Backend{Backend{u.Hostname(), p, 1}}}

	// test app not exist
	b := NewBreaker()

	req, _ := http.NewRequest("GET", "/", nil)
	req.Host = app
	w := httptest.NewRecorder()
	b.ServeHTTP(w, req)
	if w.Code != 404 {
		t.Errorf("Response code should be 404, was: %d with body: %s", w.Code, w.Body.String())
	}

	// app exist. timeline exist, but weight not set
	b = NewBreaker()
	overrideAPP(b, upstreamConfig)
	b.balancers[app] = NewWRR([]Backend{}...)

	req, _ = http.NewRequest("GET", "/", nil)
	req.Host = app
	w = httptest.NewRecorder()
	b.ServeHTTP(w, req)
	if w.Code != 403 {
		t.Errorf("Response code should be 403, was: %d with body: %s", w.Code, w.Body.String())
	}

	// app exist. timeline exist, weight settled
	b = NewBreaker()
	overrideAPP(b, upstreamConfig)
	for i := 0; i < 100; i++ {
		b.timelines[app].Incr("/", 200)
	}

	req, _ = http.NewRequest("GET", "/", nil)
	req.Host = app
	w = httptest.NewRecorder()
	b.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Errorf("Response code should be 200, was: %d with body: %s", w.Code, w.Body.String())
	}

	// app exist. timeline exist, weight settled. and circuit open
	b = NewBreaker()
	overrideAPP(b, upstreamConfig)
	for i := 0; i < 100; i++ {
		b.timelines[app].Incr("/", 200)
		b.timelines[app].Incr("/", 429)
		b.timelines[app].Incr("/", 500)
		b.timelines[app].Incr("/", 502)
	}

	req, _ = http.NewRequest("GET", "/", nil)
	req.Host = app
	w = httptest.NewRecorder()
	b.ServeHTTP(w, req)
	if w.Code != 429 {
		t.Errorf("Response code should be 429, was: %d with body: %s", w.Code, w.Body.String())
	}
}

func TestTimelineNotExist(t *testing.T) {
	defer func() {
		if err := recover(); err == nil {
			t.Errorf("should panic, but haven't")
		}
	}()

	var app = "www.example.com"
	fakeServer := httptest.NewServer(http.HandlerFunc(ServeHTTP))
	defer fakeServer.Close()
	u, err := url.ParseRequestURI(fakeServer.URL)
	if err != nil {
		t.Errorf("failed to parse fakeServer address: %s", fakeServer.URL)
	}
	p, _ := strconv.Atoi(u.Port())
	var upstreamConfig = APP{app, []string{"/"}, []string{"GET"}, []Backend{Backend{u.Hostname(), p, 1}}}

	// test app not exist
	b := NewBreaker()
	overrideAPP(b, upstreamConfig)
	//delete timeline
	delete(b.timelines, app)

	req, _ := http.NewRequest("GET", "/", nil)
	req.Host = app
	w := httptest.NewRecorder()
	b.ServeHTTP(w, req)
}

func TestBalancerNotExist(t *testing.T) {
	defer func() {
		if err := recover(); err == nil {
			t.Errorf("should panic, but haven't")
		}
	}()

	var app = "www.example.com"
	fakeServer := httptest.NewServer(http.HandlerFunc(ServeHTTP))
	defer fakeServer.Close()
	u, err := url.ParseRequestURI(fakeServer.URL)
	if err != nil {
		t.Errorf("failed to parse fakeServer address: %s", fakeServer.URL)
	}
	p, _ := strconv.Atoi(u.Port())
	var upstreamConfig = APP{app, []string{"/"}, []string{"GET"}, []Backend{Backend{u.Hostname(), p, 1}}}

	// test app not exist
	b := NewBreaker()
	overrideAPP(b, upstreamConfig)
	//delete timeline
	delete(b.balancers, app)

	req, _ := http.NewRequest("GET", "/", nil)
	req.Host = app
	w := httptest.NewRecorder()
	b.ServeHTTP(w, req)
}

func TestTSR(t *testing.T) {
	var app = "www.example.com"
	fakeServer := httptest.NewServer(http.HandlerFunc(ServeHTTP))
	defer fakeServer.Close()
	u, err := url.ParseRequestURI(fakeServer.URL)
	if err != nil {
		t.Errorf("failed to parse fakeServer address: %s", fakeServer.URL)
	}
	p, _ := strconv.Atoi(u.Port())
	var upstreamConfig = APP{
		app,
		[]string{"/hello", "/src/:world", "/file/*sys"},
		[]string{"GET", "POST", "POST"},
		[]Backend{Backend{u.Hostname(), p, 1}},
	}

	// test app not exist
	b := NewBreaker()
	overrideAPP(b, upstreamConfig)

	// GET
	req, _ := http.NewRequest("GET", "/hello/", nil)
	req.Host = app
	w := httptest.NewRecorder()
	b.ServeHTTP(w, req)

	// POST
	req, _ = http.NewRequest("POST", "/src/this/", nil)
	req.Host = app
	w = httptest.NewRecorder()
	b.ServeHTTP(w, req)

	// POST
	req, _ = http.NewRequest("POST", "/file", nil)
	req.Host = app
	w = httptest.NewRecorder()
	b.ServeHTTP(w, req)
}
