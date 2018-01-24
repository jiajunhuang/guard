package main

import (
	"github.com/valyala/fasthttp"
)

// load balancer: return which backend should we proxy to

// Backend is the backend server, usally a app server like: gunicorn+flask
type Backend struct {
	Host   string `json:"host"`
	Port   string `json:"port"`
	Weight int    `json:"weight"`
	url    string // cache the result
	client *fasthttp.HostClient
}

// NewBackend return a new backend
func NewBackend(host string, port string, weight int) Backend {
	url := host + ":" + port
	return Backend{
		host, port, weight, url,
		&fasthttp.HostClient{Addr: url, MaxConns: fasthttp.DefaultMaxConnsPerHost * 4},
	}
}

// Balancer should have a method `Select`, which return the backend we should
// proxy.
type Balancer interface {
	Select() (*Backend, bool)
}
