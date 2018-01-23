package main

const (
	statusStep   = 10
	maxStatusLen = 12
)

// Status is for counting http status code
// uint32 can be at most 4294967296, it's enough for proxy server, because this
// means in the past second, you've received 4294967296 requests, 429496729/second.
type Status struct {
	prev            *Status
	next            *Status
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
