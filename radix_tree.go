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

// ring of Status is for counting http status code.

type nodeType uint8

const (
	static   nodeType = iota // default, static string
	root                     // root node
	param                    // like `:name` in `/user/:name/hello`
	catchAll                 // like `*filepath` in `/share/*filepath`
)

// HTTPMethod is HTTP method
type HTTPMethod uint16

// HTTP Methods: https://developer.mozilla.org/en-US/docs/Web/HTTP/Methods
const (
	NONE HTTPMethod = 0 // means no method had set
	GET  HTTPMethod = 1 << iota
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
	methods   HTTPMethod
	wildChild bool    // child type is param, or catchAll
	indices   string  // first letter of childs, it's index for binary search.
	children  []*node // childrens
	leaf      bool    // if it's a leaf
	status    *Status // if it's a leaf, it should have a ring of `Status` struct
}

func min(a, b int) int {
	if a <= b {
		return a
	}

	return b
}

func (n *node) setMethods(methods ...HTTPMethod) {
	for _, method := range methods {
		// set corresponding bit
		n.methods |= method
	}
}

func (n *node) hasMethod(method HTTPMethod) bool {
	return method == (method & n.methods)
}

// addRoute adds a node with given path, handle all the resource with it.
// if it's a leaf, it should have a ring of `Status`.
func (n *node) addRoute(path string, methods ...HTTPMethod) {
	fullPath := path

	/* tree is empty */
	if n.path == "" && len(n.children) == 0 {
		n.nType = root
		// reset these properties
		n.leaf = false
		n.methods = NONE
		n.status = nil

		// insert
		n.insertChild(path, fullPath, methods...)
		return
	}

	/* tree is not empty */
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
				path:      n.path[i:],
				nType:     static,
				methods:   n.methods,
				wildChild: n.wildChild,
				indices:   n.indices,
				children:  n.children,
				leaf:      n.leaf,
				status:    n.status,
			}

			n.path = path[:i]
			n.methods = NONE
			n.leaf = false
			n.status = nil
			n.children = []*node{&child}
			n.indices = string([]byte{n.path[i]})
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
		n.insertChild(path, fullPath, methods...)
	}
}

func (n *node) insertChild(path string, fullPath string, methods ...HTTPMethod) {
	var offset int // bytes in the path have already handled
	var numParams uint8
	var maxLen = len(path)

	for i := 0; i < len(path); i++ {
		if path[i] == ':' || path[i] == '*' {
			numParams++
		}
	}

	var i = 0
	var c byte
	for ; numParams > 0; numParams-- {
		// first step, find the first wildcard(beginning with ':' or '*') of the current path
		for i = offset; i < len(path); i++ {
			c = path[i]
			if c == ':' || c == '*' {
				break
			}
		}

		// second step, find wildcard name, wildcard name cannot contain ':' and '*'
		// stops when meet '/' or the end
		end := i + 1
		for end < maxLen && path[end] != '/' {
			switch path[end] {
			case ':', '*':
				log.Panicf("wildcards ':' or '*' are not allowed in param names: %s in %s", path, fullPath)
			default:
				end++
			}
		}

		// node whose type is param or catchAll are conflict, check it
		if len(n.children) > 0 {
			log.Panicf("wildcard route %s conflict with existing children in path %s", path[i:end], fullPath)
		}

		// check if the wildcard has a name
		if end-i < 2 {
			log.Panicf("wildcards must be named with a non-empty name in path %s", fullPath)
		}

		if c == ':' { // param
			// split path at the beginning of the wildcard
			if i > 0 {
				n.path = path[offset:i]
				offset = i
			}

			child := &node{nType: param}
			n.children = []*node{child}
			n.wildChild = true
			n = child

			// if the path doesn't end with the wildcard, then there will be another non-wildcard subpath
			// starting with '/'
			if end < maxLen {
				n.path = path[offset:end]
				offset = end

				child := &node{}
				n.children = []*node{child}
				n = child
			}
		} else { //catchAll
			if end != maxLen || numParams > 1 {
				log.Panicf("catchAll routers are only allowed once at the end of the path: %s", fullPath)
			}

			if path[i-1] != '/' {
				log.Panicf("no / before catchAll in path %s", fullPath)
			}

			// this node holding path 'xxx/'
			n.path = path[offset:i]
			n.wildChild = true

			// child node holding the variable, '*xxxx'
			child := &node{path: path[i:], nType: catchAll, leaf: true, status: StatusRing()}
			child.setMethods(methods...)
			n.children = []*node{child}

			// all done
			return
		}
	}

	// insert the remaining part of path
	n.path = path[offset:]
	n.setMethods(methods...)
	n.leaf = true
	n.status = StatusRing()
}

// byPath return a node with the given path
func (n *node) byPath(path string) (nd *node, found bool) {
	return nil, false
}
