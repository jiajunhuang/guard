package main

/*
proxy server
*/

import (
	"net"
	"net/http"
	"net/http/httputil"
	"time"
)

var (
	transport http.RoundTripper = &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
			DualStack: true,
		}).DialContext,
		MaxIdleConns:          2048,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		MaxIdleConnsPerHost:   2048,
	}
)

// Proxy use httputil.ReverseProxy
func Proxy(balancer Balancer, w http.ResponseWriter, r *http.Request) {
	backend, found := balancer.Select()
	if !found {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	director := func(req *http.Request) {
		req.URL.Scheme = "http"
		req.URL.Host = backend.url
		req.Header.Add("X-Real-IP", r.Host)
		req.Header.Add("X-Forwarded-By", "Guard")
	}
	proxy := &httputil.ReverseProxy{
		Director:  director,
		Transport: transport,
	}
	proxy.ServeHTTP(w, r)
}
