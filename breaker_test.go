package main

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
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

// it's declared in proxy.go, here is just for hint
//// borrowed from Go:
//// https://github.com/golang/go/blob/64ccd4589e657a380836d87e8dd801bf53c0d475/src/net/http/httputil/reverseproxy_test.go#L675-L681
//type staticTransport struct {
//res *http.Response
//}

//func (s *staticTransport) RoundTrip(r *http.Request) (*http.Response, error) {
//return s.res, nil
//}

//type fakeBalancer struct{}

//var fakeBackend = Backend{"127.0.0.1", 9090, 1}

//func (b fakeBalancer) Select() (*Backend, bool) {
//return &fakeBackend, true
//}

func BenchmarkServeHTTP(b *testing.B) {
	// replace the global variable transport in proxy.go
	var app = "www.example.com"
	res := &http.Response{
		StatusCode: 200,
		Body:       ioutil.NopCloser(strings.NewReader("")),
	}
	transport = &staticTransport{res}

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)
	r.Host = app
	r.URL.Path = "/src/this"

	appName := "www.example.com"
	balancer := fakeBalancer{}

	breaker := NewBreaker()
	breaker.apps[appName] = NewApp(
		balancer, true,
	)

	fakeBackend.Host = "127.0.0.1"
	fakeBackend.Port = 10989
	fakeBackend.Weight = 1

	breaker.apps[appName].AddRoute("/", "GET")

	for i := 0; i < b.N; i++ {
		breaker.ServeHTTP(w, r)
	}
}
