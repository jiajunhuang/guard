package main

import (
	"sync"
)

/*
load balancer: for now, I've just implemented weighted round robin algorithms.
it's borrowed from Nginx:
https://github.com/nginx/nginx/commit/52327e0627f49dbda1e8db695e63a4b0af4448b1
*/

// Backend is the backend server, usally a app server like: gunicorn+flask
type Backend struct {
	Host   string
	Port   int
	Weight int
}

// Balancer should have a method `Select`, which return the backend we should
// proxy.
type Balancer interface {
	Select() (*Backend, bool)
}

// WRR is weighted round robin algorithm
type WRR struct {
	lock sync.Mutex

	upstream    []Backend
	totalWeight int
	weights     []int
}

// NewWRR return a instance with initialized weights & totalWeight
func NewWRR(backends ...Backend) *WRR {
	totalWeight := 0
	weights := make([]int, len(backends))

	for _, b := range backends {
		totalWeight += b.Weight
	}

	return &WRR{upstream: backends, totalWeight: totalWeight, weights: weights}
}

// Select return the backend we should proxy
// for example, weights of [5, 1, 1] should generate sequence of index:
// [1, 1, 2, 1, 3, 1, 1]
func (w *WRR) Select() (b *Backend, found bool) {
	w.lock.Lock()
	defer w.lock.Unlock()

	totalWeight := w.totalWeight
	upstream := w.upstream
	weights := w.weights
	biggest := -1
	biggestWeight := 0

	for i := range weights {
		weights[i] += upstream[i].Weight

		if weights[i] > biggestWeight {
			biggestWeight = weights[i]
			biggest = i
		}
	}

	if biggest >= 0 && biggest < len(weights) {
		weights[biggest] -= totalWeight

		return &w.upstream[biggest], true
	}

	return nil, false
}
