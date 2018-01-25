package main

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttputil"
)

type fakeBalancer struct{}

var fakeBackend = Backend{}

func setFakeBackend(url string, weight int) {
	fakeBackend.Weight = weight
	fakeBackend.URL = url
	fakeBackend.client = &fasthttp.HostClient{Addr: url, MaxConns: fasthttp.DefaultMaxConnsPerHost}
}

func (b fakeBalancer) Select() (*Backend, bool) {
	if fakeBackend.Weight == 0 {
		return nil, false
	}
	return &fakeBackend, true
}

func fakeHandler(ctx *fasthttp.RequestCtx) {
	ctx.WriteString("hoho!")
}

func TestProxyBackendNotFound(t *testing.T) {
	ln := fasthttputil.NewInmemoryListener()
	defer ln.Close()

	go fasthttp.Serve(ln, fakeHandler)

	setFakeBackend(ln.Addr().String(), 0)

	fb := fakeBalancer{}

	ctx := &fasthttp.RequestCtx{}
	ctx.Request.SetRequestURI("/")
	Proxy(fb, ctx)

	if code := ctx.Response.StatusCode(); code != fasthttp.StatusForbidden {
		t.Errorf("response code should be %d but got: %d", fasthttp.StatusForbidden, code)
	}
}

func TestProxyWorks(t *testing.T) {
	fakeServer := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}),
	)
	defer fakeServer.Close()

	u, _ := url.ParseRequestURI(fakeServer.URL)

	setFakeBackend(u.Host, 1)

	fb := fakeBalancer{}

	ctx := &fasthttp.RequestCtx{}
	// RequestURI is used first if it's not empty: https://github.com/valyala/fasthttp/issues/114
	ctx.Request.SetRequestURI("http://" + u.Host + "/")
	Proxy(fb, ctx)

	if code := ctx.Response.StatusCode(); code != fasthttp.StatusOK {
		t.Errorf("response code should be %d but got: %d", fasthttp.StatusOK, code)
	}
}
