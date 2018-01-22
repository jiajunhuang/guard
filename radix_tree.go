package main

// radix tree, combine with timeline

//               +-----------+
//               |  root, /  |            here is radix tree
//               +-----------+
//              /             \
//    +--------+              +----------+
//    | ...    |              |  user    |
//    +--------+              +----------+
//                                  |
//-----------------------------------------------------------------------------
//                                  |     below is ring of status
//                              +---------+
//                              | status1 |
//                              +---------+
//                             /          \
//                         +---------+     +---------+
//                         | status2 |     | status3 |
//                         +---------+     +---------+
//                             \          /
//                              +---------+
//                              | status4 |
//                              +---------+
//
// it's a radix tree, all the leaf has a ring of status.
// each time there is a request, we try to find find it through the radix tree,
// and then calculate the ratio, decide step down or not.

// Status is for counting http status code, it's a ring.
// uint32 can be at most 4294967296, it's enough for proxy server, because this
// means in the past second, you've received 4294967296 requests, 429496729/second.
type Status struct {
	prev            *StatusCode
	next            *StatusCode
	OK              uint32
	TooManyRequests uint32
	InternalError   uint32
	BadGateway      uint32
}

type nodeType uint8

const (
	static   nodeType = iota // default, static string
	root                     // root node
	leaf                     // it's a leaf
	param                    // like `:name` in `/user/:name/hello`
	catchAll                 // like `*filepath` in `/share/*filepath`
)

// HTTPMethod is HTTP method
type HTTPMethod uint16

// HTTP Methods: https://developer.mozilla.org/en-US/docs/Web/HTTP/Methods
const (
	GET HTTPMethod = 1 << iota
	POST
	PUT
	DELETE
	HEAD
	OPTIONS
	CONNECT
	TRACE
	PATCH
)

// radix tree is "read only" after constructed. which means, it's not read only
// but I assume it is...
type node struct {
	path  string   // common prefix of childs
	nType nodeType // node type, static, root, param, or catchAll
	// supported HTTP methods, for decide raise a `405 Method Not Allowd` or not,
	// if a method is support, the correspoding bit is set
	methods   uint16
	wildChild bool    // child type is param, or catchAll
	indices   string  // first letter of childs, it's index for binary search.
	children  []*node // childrens
	status    *Status // if it's a leaf, it should have a ring of `Status` struct
}

// addRoute adds a node with given path, handle all the resource with it.
// if it's a leaf, it should have a ring of `Status`.
func (n *node) addRoute(path string) {

}

// byPath return a node with the given path
func (n *node) byPath(path string) *node {

}
