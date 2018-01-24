package main

/*
circuit breaker, response for handle requests, decide reject it or not, record response
status.
*/

import (
	"net/http"
)

// Breaker is circuit breaker, it's a collection of Application
type Breaker struct {
	balancers map[string]*Application
}

// NewBreaker return a brand new circuit breaker, with nothing in mapper
func NewBreaker() *Breaker {
	return &Breaker{
		make(map[string]*Application),
	}
}

func (b *Breaker) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	panic("not implemented")
}
