package main

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

// borrowed from Go:
// https://github.com/golang/go/blob/64ccd4589e657a380836d87e8dd801bf53c0d475/src/net/http/httputil/reverseproxy_test.go#L675-L681
type staticTransport struct {
	res *http.Response
}

func (s *staticTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	return s.res, nil
}

type fakeBalancer struct{}

var fakeBackend = Backend{}

func (b fakeBalancer) Select() (*Backend, bool) {
	if fakeBackend.Weight == 0 {
		return nil, false
	}
	return &fakeBackend, true
}

func fakeHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("hoho!"))
}

func TestProxy(t *testing.T) {
	fakeServer := httptest.NewServer(http.HandlerFunc(fakeHandler))
	defer fakeServer.Close()
	u, err := url.ParseRequestURI(fakeServer.URL)
	if err != nil {
		t.Errorf("failed to parse fakeServer address: %s", fakeServer.URL)
	}

	fakeBackend.Host = u.Host
	fakeBackend.Port = u.Port()
	fakeBackend.Weight = 1

	fb := fakeBalancer{}
	r, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	Proxy(fb, w, r)
}

func TestProxyNotFound(t *testing.T) {
	fakeServer := httptest.NewServer(http.HandlerFunc(fakeHandler))
	defer fakeServer.Close()
	u, err := url.ParseRequestURI(fakeServer.URL)
	if err != nil {
		t.Errorf("failed to parse fakeServer address: %s", fakeServer.URL)
	}

	fakeBackend.Host = u.Host
	fakeBackend.Port = u.Port()
	fakeBackend.Weight = 0

	fb := fakeBalancer{}
	r, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	Proxy(fb, w, r)
}

func BenchmarkProxy(b *testing.B) {
	// replace the global variable transport in proxy.go
	res := &http.Response{
		StatusCode: 200,
		Body:       ioutil.NopCloser(strings.NewReader("")),
	}
	transport = &staticTransport{res}
	balancer := fakeBalancer{}

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)

	fakeBackend.Host = "127.0.0.1"
	fakeBackend.Port = "10989"
	fakeBackend.Weight = 1

	for i := 0; i < b.N; i++ {
		Proxy(balancer, w, r)
	}
}
