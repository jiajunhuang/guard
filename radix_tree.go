package main

import (
	"log"
)

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
	prev            *Status
	next            *Status
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

func min(a, b int) int {
	if a <= b {
		return a
	}

	return b
}

// addRoute adds a node with given path, handle all the resource with it.
// if it's a leaf, it should have a ring of `Status`.
func (n *node) addRoute(path string) {
	/* tree is empty */
	if n.path == "" && len(n.children) == 0 {
		n.nType = root
		n.insertChild(path)
		return
	}

	/* tree is not empty */

	fullPath := path
walk:
	for {
		maxLen := min(len(path), len(n.path))

		// find max common prefix
		i := 0
		for i < maxLen && path[i] == n.path[i] {
			i++
		}

		// if max common prefix is shorter than n.path, split n
		if i < len(n.path) {
			child := node{
				path:  n.path[i:],
				nType: static,
				// methods
				wildChild: n.wildChild,
				indices:   n.indices,
				children:  n.children,
				// status
			}

			n.children = []*node{&child}
			n.indices = string([]byte{n.path[i]})
			n.path = path[:i]
			n.wildChild = false
		}

		// path is shorter or equal than n.path, so quit
		if i == len(path) {
			return
		}

		// path is longer than n.path, so insert it!
		path = path[i:]

		// only one wildChild is permit
		if n.wildChild {
			n = n.children[0]

			// check if wildcard matchs.
			// for example, if n.path is `:name`, path should be `:name` or `:name/`
			// only thses two cases are permit, panic if not
			lenNode := len(n.path)
			lenPath := len(path)
			if !((n.path == path) || (lenNode == lenPath-1 && n.path == path[:lenNode])) {
				log.Panicf("%s in %s conflict with node %s", path, fullPath, n.path)
			}

			// everything works fine
			continue walk
		}

		c := path[0]

		// check if n is slash after param, e.g. path is `/jhon`, n.path is `:name`, and n.children is `/`
		if n.nType == param && c == '/' && len(n.children) == 1 {
			n = n.children[0]
			continue walk
		}

		// check if a child with next path bytes exists
		// TODO: use a binary search to search index. but for now, we just loop over it, because for the most cases
		// children will not be too much
		for i := 0; i < len(n.indices); i++ {
			if c == n.indices[i] {
				n = n.children[i]
				continue walk
			}
		}

		// insert it!
		if c != ':' && c != '*' {
			n.indices += string([]byte{c})
			child := &node{}
			n.children = append(n.children, child)
			n = child
		}
		n.insertChild(path)
	}
}

func (n *node) insertChild(path string) {
	panic("not implemented")
}

// byPath return a node with the given path
func (n *node) byPath(path string) (nd *node, tsr bool, found bool) {
	return nil, false, false
}
