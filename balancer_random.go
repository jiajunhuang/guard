package main

import (
	"math/rand"
)

// Rdm is struct for random balance algorithm
type Rdm struct {
	upstream []Backend
}

// NewRdm return a brand new random balancer
func NewRdm(backends ...Backend) *Rdm {
	return &Rdm{backends}
}

// Select return a backend randomly
func (r *Rdm) Select() (*Backend, bool) {
	length := len(r.upstream)
	if length == 0 {
		return nil, false
	} else if length == 1 {
		return &r.upstream[0], true
	}

	return &(r.upstream[rand.Int()%length]), true
}
