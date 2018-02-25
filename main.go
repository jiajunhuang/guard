package main

import (
	"flag"
	"log"
	"net"
	"os"
	"os/signal"
	"time"

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
	log.Printf("running with pid: %d\n", os.Getpid())

	// config manager
	go configManager()

	// proxy listener
	ln, err := net.Listen("tcp", *proxyAddr)
	if err != nil {
		log.Fatalf("error while listen at %s: %s", *proxyAddr, err)
	}
	gln := newGracefulListener(ln, time.Second*10)

	// singal handler
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for {
			<-c
			log.Printf("graceful shutdown...")
			gln.Close()
		}
	}()

	// proxy server
	if err := fasthttp.Serve(gln, breaker.ServeHTTP); err != nil {
		log.Fatalf("error in fasthttp server: %s", err)
	}
}
