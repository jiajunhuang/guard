package main

import (
	"sync"
)

// WRR is weighted round robin algorithm, it's borrowed from Nginx:
// https://github.com/nginx/nginx/commit/52327e0627f49dbda1e8db695e63a4b0af4448b1
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
	length := uint64(len(w.upstream))
	if length == 0 {
		return nil, false
	} else if length == 1 {
		return &w.upstream[0], true
	}

	w.lock.Lock()

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

		// defer is too slow...
		w.lock.Unlock()

		return &w.upstream[biggest], true
	}

	// defer is too slow...
	w.lock.Unlock()

	return nil, false
}
