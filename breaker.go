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
// - balancer, who response for choose which backend should we proxy, `Host` -> `Balancer` mapper
type Breaker struct {
	routers   map[string]*Router
	balancers map[string]Balancer
}

// NewBreaker return a brand new circuit breaker, with nothing in mapper
func NewBreaker() *Breaker {
	return &Breaker{
		make(map[string]*Router),
		make(map[string]Balancer),
	}
}

// UpdateAPP insert or overwrite a existing app in router, balancer
// NOTE: Not concurrency safe!
func (b *Breaker) UpdateAPP(app string) {
	b.routers[app] = NewRouter()
	b.balancers[app] = NewWRR()
}

func (b *Breaker) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	panic("not implemented")
}
