package main

import (
	"sync/atomic"
)

// RR is struct for naive round robin balance algorithm
type RR struct {
	upstream []Backend
	index    uint64
}

// NewRR return a brand new naive round robin balancer
func NewRR(backends ...Backend) *RR {
	return &RR{backends, 0}
}

// Select return a backend randomly
func (r *RR) Select() (b *Backend, found bool) {
	length := uint64(len(r.upstream))
	if length == 0 {
		return nil, false
	} else if length == 1 {
		return &r.upstream[0], true
	}

	// TODO: shuold we check for overflow?
	return &(r.upstream[atomic.AddUint64(&r.index, 1)%length]), true
}
