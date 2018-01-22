package main

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestProxy(t *testing.T) {
}

// borrowed from Go:
// https://github.com/golang/go/blob/64ccd4589e657a380836d87e8dd801bf53c0d475/src/net/http/httputil/reverseproxy_test.go#L675-L681
type staticTransport struct {
	res *http.Response
}

func (s *staticTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	return s.res, nil
}

type fakeBalancer struct{}

var fakeBackend = Backend{"127.0.0.1", 10089, 1}

func (b fakeBalancer) Select() (*Backend, bool) {
	return &fakeBackend, true
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

	for i := 0; i < b.N; i++ {
		Proxy(balancer, w, r)
	}
}
