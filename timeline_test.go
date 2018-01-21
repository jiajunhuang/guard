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
		tl.head.next = NewBucket(now - int64(i*bucketStep))
		tl.head = tl.head.next
	}
	tl.head = originHead

	tl.Incr("/user/:/hello", 200)
}

func TestBucketKey(t *testing.T) {
	r := BucketKey("/user/:/hello", 200)
	if r != "/user/:/hello@200" {
		t.Errorf("`BucketKey` should return `/user/:/hello@200`, but got: %s", r)
	}
}

func TestQueryStatus(t *testing.T) {
	tl := NewTimeline()
	url := "/user/:/hello"
	tl.Incr(url, 200)
	tl.Incr(url, 429)
	tl.Incr(url, 500)
	tl.Incr(url, 502)

	c200, c429, c500, c502, _ := tl.QueryStatus(url)
	if c200 != 1 || c429 != 1 || c500 != 1 || c502 != 1 {
		t.Errorf("timebucket incr error, status of 200, 429, 500, 502 should be one")
	}
}

func TestIncrWithNilTail(t *testing.T) {
	defer func() {
		err := recover()
		if err == nil {
			t.Error("timeline.Incr should panic because it's tail is nil, but it not")
		}
	}()

	tl := NewTimeline()
	tl.tail = nil
	url := "/user/:/hello"
	code := 200

	tl.Incr(url, code)
}

func TestQueryStatusWithNilTail(t *testing.T) {
	defer func() {
		err := recover()
		if err == nil {
			t.Error("timeline.QueryStatus should panic because it's tail is nil, but it not")
		}
	}()

	tl := NewTimeline()
	tl.tail = nil
	url := "/user/:/hello"

	tl.QueryStatus(url)
}

func BenchmarkIncr(b *testing.B) {
	tl := NewTimeline()

	for i := 0; i < b.N; i++ {
		tl.Incr("/user/:/hello", 200)
	}
}

func BenchmarkQueryStatus(b *testing.B) {
	tl := NewTimeline()

	tl.Incr("/user/:/hello", 200)
	tl.Incr("/user/:/hello", 429)
	tl.Incr("/user/:/hello", 500)
	tl.Incr("/user/:/hello", 502)

	for i := 0; i < b.N; i++ {
		tl.QueryStatus("/user/:/hello")
	}
}

func BenchmarkBucketKey(b *testing.B) {
	for i := 0; i < b.N; i++ {
		BucketKey("/user/:/hello", 200)
	}
}
