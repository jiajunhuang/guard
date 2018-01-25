package main

import (
	"log"

	"github.com/valyala/fasthttp"
)

// guard is a high performance circuit breaker written in Go.

var (
	breaker = NewBreaker()
)

func main() {
	go configManager()
	if err := fasthttp.ListenAndServe(":23456", breaker.ServeHTTP); err != nil {
		log.Fatalf("error in fasthttp server: %s", err)
	}
}
