package main

// load balancer: return which backend should we proxy to

// Backend is the backend server, usally a app server like: gunicorn+flask
type Backend struct {
	Host   string `json:"host"`
	Port   string `json:"port"`
	Weight int    `json:"weight"`
	url    string // cache the result
}

// NewBackend return a new backend
func NewBackend(host string, port string, weight int) Backend {
	return Backend{host, port, weight, host + ":" + port}
}

// Balancer should have a method `Select`, which return the backend we should
// proxy.
type Balancer interface {
	Select() (*Backend, bool)
}
