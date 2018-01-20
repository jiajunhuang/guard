package main

import (
	"testing"
)

func TestRightNow(t *testing.T) {
	if RightNow() != RightNow() {
		t.Fatalf("right now should return the same result")
	}
}

func TestTimelineIncr(t *testing.T) {
	tl := NewTimeline()
	now := RightNow()
	oldest := now - (maxBuckets+1)*bucketStep

	tl.head.key = oldest
	originHead := tl.head
	// create buckets manually
	for i := maxBuckets; i > 0; i-- {
		tl.head.next = NewBucket(now - i*bucketStep)
		tl.head = tl.head.next
	}
	tl.head = originHead

	tl.Incr("/user/:/hello", 200)
}
