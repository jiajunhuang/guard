package main

import (
	"testing"

	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttputil"
)

func TestBreakerServeHTTP(t *testing.T) {
	ln := fasthttputil.NewInmemoryListener()
	defer ln.Close()

	go fasthttp.Serve(ln, fakeHandler)

	setFakeBackend(ln.Addr().String(), 0)

	fb := fakeBalancer{}
	a := NewApp(fb, true)
	a.AddRoute("/", "GET")

	breaker := NewBreaker()
	appName := "www.example.com"

	// app not found
	ctx := &fasthttp.RequestCtx{}
	ctx.Request.SetRequestURI("/")
	ctx.Request.SetHost(appName)
	breaker.ServeHTTP(ctx)
	if code := ctx.Response.StatusCode(); code != fasthttp.StatusNotFound {
		t.Errorf("response code should be %d but got: %d", fasthttp.StatusNotFound, code)
	}

	// app found, but circuit is open, so return 403 forbidden
	ctx = &fasthttp.RequestCtx{}
	ctx.Request.SetRequestURI("/")
	ctx.Request.SetHost(appName)
	breaker.apps[appName] = a
	breaker.ServeHTTP(ctx)
	if code := ctx.Response.StatusCode(); code != fasthttp.StatusForbidden {
		t.Errorf("response code should be %d but got: %d", fasthttp.StatusForbidden, code)
	}
}
