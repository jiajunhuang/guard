package main

import (
	"log"
	"net/http"
	"sync/atomic"
	"unsafe"
)

const (
	statusStep   int64 = 10
	maxStatusLen       = 12
)

// RightNow return status key
func RightNow() int64 {
	t := CoarseTimeNow().Unix()
	return t - t%statusStep
}

// Status is for counting http status code
// uint32 can be at most 4294967296, it's enough for proxy server, because this
// means in the past second, you've received 4294967296 requests, 429496729/second.
type Status struct {
	prev            *Status
	next            *Status
	key             int64 // for now, key is time
	OK              uint32
	TooManyRequests uint32
	InternalError   uint32
	BadGateway      uint32
}

// StatusRing return a ring of status
func StatusRing() *Status {
	head := &Status{}
	cursor := head

	for i := 0; i < maxStatusLen-1; i++ {
		cursor.next = &Status{}
		cursor.next.prev = cursor
		cursor = cursor.next
	}

	head.prev = cursor
	cursor.next = head

	return head
}

// refreshStatus refresh the current status if it's outdate, and return the latest one
func (n *node) refreshStatus(now int64) *Status {
	status := n.status
	if status.key != now {
		if atomic.CompareAndSwapPointer(
			// first, get address of n.status, means, address of `status field in n`, get it's address,
			// cast it to `*unsafe.Pointer`
			(*unsafe.Pointer)(unsafe.Pointer(&(n.status))), unsafe.Pointer(status), unsafe.Pointer(status.next),
		) {
			// clean old data, though it may cause some dirty reads
			atomic.StoreInt64(&n.status.key, now)
			atomic.StoreUint32(&n.status.OK, 0)
			atomic.StoreUint32(&n.status.TooManyRequests, 0)
			atomic.StoreUint32(&n.status.InternalError, 0)
			atomic.StoreUint32(&n.status.BadGateway, 0)
		}
	}

	return n.status
}

// incr increase by 1 on the given genericURL and status code, return value after incr
func (n *node) incr(code int) uint32 {
	if n.status == nil {
		log.Panicf("status of node %+v is nil", n)
	}

	status := n.refreshStatus(RightNow())
	switch code {
	case http.StatusOK:
		return atomic.AddUint32(&status.OK, 1)
	case http.StatusTooManyRequests:
		return atomic.AddUint32(&status.TooManyRequests, 1)
	case http.StatusInternalServerError:
		return atomic.AddUint32(&status.InternalError, 1)
	case http.StatusBadGateway:
		return atomic.AddUint32(&status.BadGateway, 1)
	default:
		log.Printf("ignore status code %d", code)
		return 0 // just for go lint, code here should never been execute
	}
}

func (n *node) query() (uint32, uint32, uint32, uint32, float64) {
	if n.status == nil {
		log.Panicf("status of node %+v is nil", n)
	}

	status := n.refreshStatus(RightNow())
	ok, too, internal, bad := status.OK, status.TooManyRequests, status.InternalError, status.BadGateway

	ratio := float64(
		too+internal+bad,
	) / float64(
		1+ok+too+internal+bad,
	)

	return ok, too, internal, bad, ratio
}
