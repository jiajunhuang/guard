package main

import (
	"testing"

	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttputil"
)

type fakeBalancer struct{}

var fakeBackend = Backend{}

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

	fakeBackend.url = ln.Addr().String()
	fakeBackend.Weight = 0

	fb := fakeBalancer{}

	ctx := &fasthttp.RequestCtx{}
	ctx.Request.SetRequestURI("/")
	Proxy(fb, ctx)

	if code := ctx.Response.StatusCode(); code != fasthttp.StatusForbidden {
		t.Errorf("response code should be %d but got: %d", fasthttp.StatusForbidden, code)
	}
}

func TestProxyWorks(t *testing.T) {
	ln := fasthttputil.NewInmemoryListener()
	defer ln.Close()

	go fasthttp.Serve(ln, fakeHandler)

	fakeBackend.url = ln.Addr().String()
	fakeBackend.Weight = 1
	fakeBackend.client = &fasthttp.HostClient{Addr: fakeBackend.url, MaxConns: fasthttp.DefaultMaxConnsPerHost}

	fb := fakeBalancer{}

	ctx := &fasthttp.RequestCtx{}
	ctx.Request.SetRequestURI("/")
	Proxy(fb, ctx)

	if code := ctx.Response.StatusCode(); code != fasthttp.StatusOK {
		t.Errorf("response code should be %d but got: %d", fasthttp.StatusOK, code)
	}
}
