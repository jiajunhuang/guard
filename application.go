package main

import (
	"log"
	"net/http"
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

func (a *Application) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if a.root == nil {
		log.Panic("application should bind a URL-tree")
	}
	if a.balancer == nil {
		log.Panic("application should bind a load balancer")
	}

	path := r.URL.Path
	n, tsr, found := a.root.byPath(path)

	// redirect?
	if tsr && a.TSRRedirect {
		code := http.StatusMovedPermanently
		if r.Method != "GET" {
			code = http.StatusTemporaryRedirect
		}

		if len(path) > 1 && path[len(path)-1] == '/' {
			r.URL.Path = path[:len(path)-1]
		} else {
			r.URL.Path = path + "/"
		}
		log.Printf("redirect to %s", r.URL.String())
		http.Redirect(w, r, r.URL.String(), code)
		return
	}

	// not found
	if !found {
		http.NotFound(w, r)
		return
	}

	// method allowed?
	if !n.hasMethod(convertMethod(r.Method)) {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// circuit breaker is open?
	_, _, _, _, ratio := n.query()
	if ratio > 0.3 {
		log.Printf("too many requests, ratio is %f", ratio)
		w.WriteHeader(http.StatusTooManyRequests)
		return
	}

	// proxy!
	responseWriter := NewResponseWriter(w)
	Proxy(a.balancer, responseWriter, r)

	// feedback the result
	n.incr(responseWriter.Status())
}
