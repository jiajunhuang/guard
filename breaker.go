package main

import (
	"github.com/valyala/fasthttp"
)

/*
circuit breaker, response for handle requests, decide reject it or not, record response
status.
*/

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

func (b *Breaker) ServeHTTP(ctx *fasthttp.RequestCtx) {
	appName := string(ctx.Host())
	var app *Application
	var exist bool
	if app, exist = b.apps[appName]; !exist {
		ctx.WriteString("app " + appName + " not exist")
		ctx.SetStatusCode(fasthttp.StatusNotFound)
		return
	}

	app.ServeHTTP(ctx)
}
