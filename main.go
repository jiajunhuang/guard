package main

import (
	"flag"
	"log"

	"github.com/valyala/fasthttp"
)

// guard is a high performance circuit breaker written in Go.

var (
	proxyAddr  = flag.String("proxyAddr", ":23456", "proxy server listen at")
	configAddr = flag.String("configAddr", ":12345", "config server listen at")
	configPath = flag.String("configPath", "/tmp/guard.json", "configuration sync path")

	breaker = NewBreaker()
)

func main() {
	flag.Parse()

	go configManager()
	if err := fasthttp.ListenAndServe(*proxyAddr, breaker.ServeHTTP); err != nil {
		log.Fatalf("error in fasthttp server: %s", err)
	}
}
