package main

/*
proxy server
*/

import (
	"log"

	"github.com/valyala/fasthttp"
)

// Proxy use fasthttp: https://github.com/valyala/fasthttp/issues/64
func Proxy(balancer Balancer, ctx *fasthttp.RequestCtx) int {
	backend, found := balancer.Select()
	if !found {
		ctx.SetStatusCode(fasthttp.StatusForbidden)
		return fasthttp.StatusForbidden
	}

	client := backend.client
	req := &ctx.Request
	resp := &ctx.Response

	// prepare
	req.Header.Del("Connection")

	// proxy
	if err := client.Do(req, resp); err != nil {
		log.Printf("failed to proxy: %s", err)
		return fasthttp.StatusBadGateway
	}

	// after
	resp.Header.Del("Connection")

	return resp.StatusCode()
}
