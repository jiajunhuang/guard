package main

/*
proxy server
*/

import (
	"net/http"
	"net/http/httputil"
)

// Proxy use httputil.ReverseProxy
func Proxy(wrr *WRR, w ResponseWriter, r *http.Request) {
	backend, found := wrr.Select()
	if !found {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	director := func(req *http.Request) {
		req.URL.Scheme = "http"
		req.URL.Host = backend.ToURL()
		req.Header.Add("X-Real-IP", r.Host)
		req.Header.Add("X-Forwarded-By", "Guard")
	}
	proxy := &httputil.ReverseProxy{Director: director}
	proxy.ServeHTTP(w, r)
}
