package main

import (
	"testing"

	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttputil"
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

	a := NewApp(NewRdm(), true)
	a.root = nil

	ctx := &fasthttp.RequestCtx{}
	ctx.Request.SetRequestURI("/")

	a.ServeHTTP(ctx)
}

func TestApplicationNilBalancer(t *testing.T) {
	defer shouldPanic()

	a := NewApp(nil, true)

	ctx := &fasthttp.RequestCtx{}
	ctx.Request.SetRequestURI("/")

	a.ServeHTTP(ctx)
}

func TestApplicationRedirect(t *testing.T) {
	ln := fasthttputil.NewInmemoryListener()
	defer ln.Close()

	go fasthttp.Serve(ln, fakeHandler)

	setFakeBackend(ln.Addr().String(), 1)

	fb := fakeBalancer{}
	a := NewApp(fb, true)
	a.AddRoute("/user/jhon", "POST")
	a.AddRoute("/user/jhon/card/", "POST")

	//redirect
	ctx := &fasthttp.RequestCtx{}
	ctx.Request.SetRequestURI("/user/jhon/")
	ctx.Request.Header.SetMethod("POST")
	a.ServeHTTP(ctx)
	if code := ctx.Response.StatusCode(); code != fasthttp.StatusTemporaryRedirect {
		t.Errorf("response code should be %d but got: %d", fasthttp.StatusTemporaryRedirect, code)
	}

	// redirect
	ctx = &fasthttp.RequestCtx{}
	ctx.Request.SetRequestURI("/user/jhon/card")
	ctx.Request.Header.SetMethod("POST")
	a.ServeHTTP(ctx)
	if code := ctx.Response.StatusCode(); code != fasthttp.StatusTemporaryRedirect {
		t.Errorf("response code should be %d but got: %d", fasthttp.StatusTemporaryRedirect, code)
	}

	// not found
	ctx = &fasthttp.RequestCtx{}
	ctx.Request.SetRequestURI("/user/what/card")
	ctx.Request.Header.SetMethod("POST")
	a.ServeHTTP(ctx)
	if code := ctx.Response.StatusCode(); code != fasthttp.StatusNotFound {
		t.Errorf("response code should be %d but got: %d", fasthttp.StatusNotFound, code)
	}

	// method not allowed
	ctx = &fasthttp.RequestCtx{}
	ctx.Request.SetRequestURI("/user/jhon")
	ctx.Request.Header.SetMethod("GET")
	a.ServeHTTP(ctx)
	if code := ctx.Response.StatusCode(); code != fasthttp.StatusMethodNotAllowed {
		t.Errorf("response code should be %d but got: %d", fasthttp.StatusMethodNotAllowed, code)
	}
}

func TestApplicationCircuit(t *testing.T) {
	ln := fasthttputil.NewInmemoryListener()
	defer ln.Close()

	go fasthttp.Serve(ln, fakeHandler)

	setFakeBackend(ln.Addr().String(), 1)

	fb := fakeBalancer{}
	a := NewApp(fb, true)
	a.AddRoute("/user/jhon", "POST")

	//redirect
	ctx := &fasthttp.RequestCtx{}
	ctx.Request.SetRequestURI("/user/jhon/")
	ctx.Request.Header.SetMethod("POST")
	a.ServeHTTP(ctx)
	if code := ctx.Response.StatusCode(); code != fasthttp.StatusTemporaryRedirect {
		t.Errorf("response code should be %d but got: %d", fasthttp.StatusTemporaryRedirect, code)
	}

	// circuit is not on
	ctx = &fasthttp.RequestCtx{}
	ctx.Request.SetRequestURI("/user/jhon")
	ctx.Request.Header.SetMethod("POST")
	a.ServeHTTP(ctx) // do not check what `Proxy` returns

	// circuit is on
	for i := 0; i < 100; i++ {
		a.root.incr(fasthttp.StatusBadGateway)
	}
	a.ServeHTTP(ctx)
	if code := ctx.Response.StatusCode(); code != fasthttp.StatusTooManyRequests {
		t.Errorf("response code should be %d but got: %d", fasthttp.StatusTooManyRequests, code)
	}
}
