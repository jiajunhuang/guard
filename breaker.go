package main

/*
circuit breaker, response for handle requests, decide reject it or not, record response
status.
*/

import (
	"net/http"
)

// Breaker is circuit breaker, it's a collection of:
// - routers, which is `Host` -> `URL pattern` mapper
// - timeline, who response for record response status, `Host` -> `Timeline` mapper
// - balancer, who response for choose which backend should we proxy, `Host` -> `Balancer` mapper
type Breaker struct {
	routers   map[string]*Router
	timelines map[string]*Timeline
	balancers map[string]Balancer
}

// NewBreaker return a brand new circuit breaker, with nothing in mapper
func NewBreaker() *Breaker {
	return &Breaker{
		make(map[string]*Router),
		make(map[string]*Timeline),
		make(map[string]Balancer),
	}
}

// UpdateAPP insert or overwrite a existing app in router, timeline, balancer
// NOTE: Not concurrency safe!
func (b *Breaker) UpdateAPP(app string) {
	b.routers[app] = NewRouter()
	b.timelines[app] = NewTimeline()
	b.balancers[app] = NewWRR()
}

func (b *Breaker) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	app := r.Host
	// if app exist?
	var exist bool
	router, exist := b.routers[app]
	if !exist {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("app " + app + " not exist, please contact admin to update the configuration"))
		return
	}

	// if circuit breaker open?
	timeline, exist := b.timelines[app]
	if !exist {
		panic("timeline of app " + app + "does not exist, it should been create in registration")
	}
	path := r.URL.Path
	_, url, tsr := router.GenericURL(r.Method, path)
	if tsr && router.RedirectTrailingSlash {
		code := 301 // Permanent redirect, request with GET method
		if r.Method != "GET" {
			// Temporary redirect, request with same method
			// As of Go 1.3, Go does not support status code 308.
			code = 307
		}
		if len(path) > 1 && path[len(path)-1] == '/' {
			r.URL.Path = path[:len(path)-1]
		} else {
			r.URL.Path = path + "/"
		}
		http.Redirect(w, r, r.URL.String(), code)
		return
	}
	_, _, _, _, ratio := timeline.QueryStatus(url)
	if ratio > 0.3 {
		w.WriteHeader(http.StatusTooManyRequests)
		return
	}

	// if circuit breaker not open, proxy
	balancer, exist := b.balancers[app]
	if !exist {
		panic("balancer of app " + app + "does not exist, it should been create in registration")
	}
	responseWriter := NewResponseWriter(w)
	Proxy(balancer, *responseWriter, r)

	// record the response
	timeline.Incr(url, responseWriter.Status())
}
