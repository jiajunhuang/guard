package main

import (
	"github.com/valyala/fasthttp"
)

// load balancer: return which backend should we proxy to
const (
	LBMWRR    = "wrr"
	LBMRR     = "rr"
	LBMRandom = "random"
)

// Backend is the backend server, usually a app server like: gunicorn+flask
type Backend struct {
	Weight int
	URL    string // cache the result
	client *fasthttp.HostClient
}

// NewBackend return a new backend
func NewBackend(url string, weight int) Backend {
	return Backend{
		weight, url,
		&fasthttp.HostClient{Addr: url, MaxConns: fasthttp.DefaultMaxConnsPerHost * 4},
	}
}

// Balancer should have a method `Select`, which return the backend we should
// proxy.
type Balancer interface {
	Select() (*Backend, bool)
}
