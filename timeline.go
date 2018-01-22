package main

import (
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

/*
timeline is a singly linked list wrapper, whose node is a bucket that stores response
status of each URL. we don't care what HTTP Method extracly, we just care about URL.

                    +----------+
                    | timeline |
                    +----------+
                   /            \
              +--------+        +--------+
              |  head  |->....->|  tail  |
              +--------+        +--------+

and each bucket is a map, the key is {GenericURL}@{HTTP Response Status Code}, like:
`/user/:/hello@200`, `/user/*@500`, `/user/:@502`, etc.

each bucket collect all the response status in recent 10 seconds. the key of bucket
is timestamp%10
*/

const (
	bucketStep = 10 // each bucket stores response code in last `bucketStep` seconds
	maxBuckets = 60 // we store `maxBuckets` buckets
)

// RightNow return latest bucket key
func RightNow() int64 {
	ts := time.Now().Unix()
	return ts - ts%bucketStep
}

// Status is an collection of status: 200, 429, 500, 502
type Status struct {
	OK              uint32 `json:"ok"`
	TooManyRequests uint32 `json:"too_many_requests"`
	InternalError   uint32 `json:"internal_error"`
	BadGateway      uint32 `json:"bad_gateway"`
}

// Counter counts value of key
type Counter map[string]*Status

// Bucket is bucket in timeline
type Bucket struct {
	key     int64
	counter Counter
	next    *Bucket
}

// NewBucket return a brand new bucket with given key
func NewBucket(key int64) *Bucket {
	return &Bucket{key: key, counter: make(Counter)}
}

// Timeline is a singly linked list wrapper
type Timeline struct {
	lock sync.RWMutex
	head *Bucket
	tail *Bucket
}

// NewTimeline return a brand new timeline
func NewTimeline() *Timeline {
	b := Bucket{key: RightNow(), counter: make(Counter)}
	t := Timeline{}
	t.head = &b
	t.tail = &b

	return &t
}

// Incr increase by 1 on the given genericURL and status code, return value after incr
func (t *Timeline) Incr(url string, code int) uint32 {
	now := RightNow()
	tail := t.tail

	if tail == nil {
		panic("timelist should always has at least one bucket, but now tail is pointer to nil")
	}

	// lock...
	t.lock.Lock()

	if tail.key != now {
		b := NewBucket(now)
		tail.next = b
		t.tail = b
		tail = b
	}

	// check if head is outdated
	oldest := now - bucketStep*maxBuckets
	for {
		if t.head.key < oldest {
			t.head = t.head.next
		} else {
			break
		}
	}

	var status *Status
	status = tail.counter[url]
	if status == nil {
		status = &Status{}
		tail.counter[url] = status
	}

	// defer is too slow
	t.lock.Unlock()

	switch code {
	case 200:
		return atomic.AddUint32(&status.OK, 1)
	case 429:
		return atomic.AddUint32(&status.TooManyRequests, 1)
	case 500:
		return atomic.AddUint32(&status.InternalError, 1)
	case 502:
		return atomic.AddUint32(&status.BadGateway, 1)
	default:
		panic("bad status code" + strconv.Itoa(code))
	}
}

// QueryStatus return counts of status 200, 429, 500, 502, and failure ratio
func (t *Timeline) QueryStatus(url string) (uint32, uint32, uint32, uint32, float64) {
	tail := t.tail
	if tail == nil {
		panic("t.tail should never be nil")
	}

	var status *Status

	// lock
	t.lock.RLock()

	status = tail.counter[url]
	if status == nil {
		// it will be created while execute Incr
		t.lock.RUnlock()
		return 0, 0, 0, 0, 0.0
	}

	ok, too, internal, bad := status.OK, status.TooManyRequests, status.InternalError, status.BadGateway

	// defer is too slow
	t.lock.RUnlock()

	ratio := float64(
		too+internal+bad,
	) / float64(
		1+ok+too+internal+bad,
	)

	return ok, too, internal, bad, ratio
}
