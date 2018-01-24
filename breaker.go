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
	apps map[string]*Application
}

// NewBreaker return a brand new circuit breaker, with nothing in mapper
func NewBreaker() *Breaker {
	return &Breaker{
		make(map[string]*Application),
	}
}

func (b *Breaker) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	appName := r.Host
	var app *Application
	var exist bool
	if app, exist = b.apps[appName]; !exist {
		w.Write([]byte("app " + appName + " not exist"))
		w.WriteHeader(http.StatusNotFound)
		return
	}

	app.ServeHTTP(w, r)
}
