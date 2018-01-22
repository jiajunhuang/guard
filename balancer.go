package main

import (
	"strconv"
)

// load balancer: return which backend should we proxy to

// Backend is the backend server, usally a app server like: gunicorn+flask
type Backend struct {
	Host   string `json:"host"`
	Port   int    `json:"port"`
	Weight int    `json:"weight"`
}

// ToURL return a string that is host:port style
func (b Backend) ToURL() string {
	return b.Host + ":" + strconv.Itoa(b.Port)
}

// Balancer should have a method `Select`, which return the backend we should
// proxy.
type Balancer interface {
	Select() (*Backend, bool)
}
