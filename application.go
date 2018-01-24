package main

import (
	"log"

	"github.com/valyala/fasthttp"
)

// Application is an abstraction of radix-tree, timeline, balancer, and configurations...
type Application struct {
	// redirect if tsr is true?
	TSRRedirect bool

	balancer Balancer
	root     *node
}

// NewApp return a brand new Application
func NewApp(b Balancer, tsr bool) *Application {
	return &Application{tsr, b, &node{}}
}

func convertMethod(methods ...string) HTTPMethod {
	httpMethods := NONE

	if len(methods) == 0 {
		log.Panicf("at least one method is required")
	}

	for _, m := range methods {
		switch m {
		case "GET":
			httpMethods |= GET
		case "POST":
			httpMethods |= POST
		case "PUT":
			httpMethods |= PUT
		case "DELETE":
			httpMethods |= DELETE
		case "HEAD":
			httpMethods |= HEAD
		case "OPTIONS":
			httpMethods |= OPTIONS
		case "CONNECT":
			httpMethods |= CONNECT
		case "TRACE":
			httpMethods |= TRACE
		case "PATCH":
			httpMethods |= PATCH
		default:
			log.Panicf("bad http method: %s", m)
		}
	}

	return httpMethods
}

// AddRoute add a route to itself
func (a *Application) AddRoute(path string, methods ...string) {
	a.root.addRoute(path, convertMethod(methods...))
}

func (a *Application) ServeHTTP(ctx *fasthttp.RequestCtx) {
	if a.root == nil {
		log.Panic("application should bind a URL-tree")
	}
	if a.balancer == nil {
		log.Panic("application should bind a load balancer")
	}

	path := string(ctx.Path())
	n, tsr, found := a.root.byPath(path)

	// redirect?
	if tsr && a.TSRRedirect {
		code := fasthttp.StatusMovedPermanently
		if string(ctx.Method()) != "GET" {
			code = fasthttp.StatusTemporaryRedirect
		}

		var redirectTo string
		if len(path) > 1 && path[len(path)-1] == '/' {
			redirectTo = path[:len(path)-1]
		} else {
			redirectTo = path + "/"
		}
		log.Printf("redirect to %s", redirectTo)
		ctx.Redirect(redirectTo, code)
		return
	}

	// not found
	if !found {
		ctx.NotFound()
		return
	}

	// method allowed?
	if !n.hasMethod(convertMethod(string(ctx.Method()))) {
		ctx.SetStatusCode(fasthttp.StatusMethodNotAllowed)
		return
	}

	// circuit breaker is open?
	_, _, _, _, ratio := n.query()
	if ratio > 0.3 {
		log.Printf("too many requests, ratio is %f", ratio)
		ctx.SetStatusCode(fasthttp.StatusTooManyRequests)
		return
	}

	// proxy!
	code := Proxy(a.balancer, ctx)

	// feedback the result
	n.incr(code)
}
