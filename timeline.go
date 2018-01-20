package main

import (
	"fmt"
	"strconv"
	"sync"
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

// BucketKey generate bucket key
func BucketKey(url string, code int) string {
	return fmt.Sprintf("%s@%d", url, code)
}

// RightNow return latest bucket key
func RightNow() int {
	t := time.Now()
	ts, _ := strconv.Atoi(t.Format("20060102150405"))
	return ts - ts%bucketStep
}

// Counter counts value of key
type Counter map[string]uint32

// Bucket is bucket in timeline
type Bucket struct {
	key     int
	counter Counter
	next    *Bucket
}

// NewBucket return a brand new bucket with given key
func NewBucket(key int) *Bucket {
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
func (t *Timeline) Incr(genericURL string, code int) uint32 {
	key := fmt.Sprintf("%s@%d", genericURL, code)
	now := RightNow()

	// lock...
	t.lock.Lock()
	defer t.lock.Unlock()

	if t.tail == nil {
		panic("timelist should always has at least one bucket, but now tail is pointer to nil")
	}

	if t.tail.key != now {
		b := NewBucket(now)
		t.tail.next = b
		t.tail = b
	}

	t.tail.counter[key]++

	// check if head is outdated
	oldest := RightNow() - bucketStep*maxBuckets
	for {
		if t.head.key < oldest {
			t.head = t.head.next
		} else {
			break
		}
	}

	return t.tail.counter[key]
}

// QueryStatus return counts of status 200, 429, 500, 502, and failure ratio
func (t *Timeline) QueryStatus(url string) (uint32, uint32, uint32, uint32, float64) {
	t.lock.RLock()
	defer t.lock.RUnlock()

	tail := t.tail

	count200 := tail.counter[BucketKey(url, 200)]
	count429 := tail.counter[BucketKey(url, 429)]
	count500 := tail.counter[BucketKey(url, 500)]
	count502 := tail.counter[BucketKey(url, 502)]
	ratio := 1 - float64(count200)/float64(count200+count429+count500+count502)

	return count200, count429, count500, count502, ratio
}
