package main

import (
	"bytes"
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
	noCopy noCopy
	path   []byte   // common prefix of childs
	nType  nodeType // node type, static, root, param, or catchAll
	// supported HTTP methods, for decide raise a `405 Method Not Allowd` or not,
	// if a method is support, the correspoding bit is set
	methods   HTTPMethod
	wildChild bool    // child type is param, or catchAll
	indices   []byte  // first letter of childs, it's index for binary search.
	children  []*node // childrens
	isLeaf    bool    // if it's a leaf
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
func (n *node) addRoute(path []byte, methods ...HTTPMethod) {
	fullPath := path

	/* tree is empty */
	if len(n.path) == 0 && len(n.children) == 0 {
		n.nType = root
		// reset these properties
		n.isLeaf = false
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
				isLeaf:    n.isLeaf,
				status:    n.status,
			}

			n.methods = NONE
			n.isLeaf = false
			n.status = nil
			n.children = []*node{&child}
			n.indices = []byte{n.path[i]}
			n.path = path[:i]
			n.wildChild = false
		}

		// path is shorter or equal than n.path, so quit
		if i == len(path) {
			n.setMethods(methods...)
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
			if lenPath >= lenNode && bytes.Equal(n.path, path[:lenNode]) &&
				// Check for longer wildcard, e.g. :name and :names
				(lenNode >= lenPath || path[lenNode] == '/') {
				continue walk
			} else {
				log.Panicf("%s in %s conflict with node %s", path, fullPath, n.path)

			}
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
			n.indices = append(n.indices, c)
			child := &node{}
			n.children = append(n.children, child)
			n = child
		}
		n.insertChild(path, fullPath, methods...)
	}
}

func (n *node) insertChild(path []byte, fullPath []byte, methods ...HTTPMethod) {
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
			child := &node{path: path[i:], nType: catchAll, isLeaf: true, status: StatusRing()}
			child.setMethods(methods...)
			n.children = []*node{child}

			// all done
			return
		}
	}

	// insert the remaining part of path
	n.path = path[offset:]
	n.setMethods(methods...)
	n.isLeaf = true
	n.status = StatusRing()
}

// byPath return a node with the given path
func (n *node) byPath(path []byte) (nd *node, tsr bool, found bool) {
walk:
	for {
		if len(path) > len(n.path) {
			if bytes.Equal(path[:len(n.path)], n.path) {
				path = path[len(n.path):]
				// if this node does not have a wildcard(param or catchAll) child, we can just look up
				// the next child node and continue to walk down the tree
				if !n.wildChild {
					c := path[0]

					for i := 0; i < len(n.indices); i++ {
						if c == n.indices[i] {
							n = n.children[i]
							continue walk
						}
					}

					// nothing found
					// we can recommend to redirect to the same URL without a trailing slash if a leaf
					// exists for that path
					tsr = (bytes.Equal(path, []byte("/")) && n.isLeaf)
					return nil, tsr, false
				}

				// handle wildcard child
				n = n.children[0]
				switch n.nType {
				case param:
					end := 0
					for end < len(path) && path[end] != '/' {
						end++
					}

					// we need to go deeper, because we've not visit all bytes in path
					if end < len(path) {
						if len(n.children) > 0 {
							path = path[end:]
							n = n.children[0]
							continue walk
						}

						// oh, no, we can't go deeper
						// if URL is `/user/:name/`, redirect it to `/user/:name`
						tsr = (len(path) == end+1 && path[len(path)-1] == '/')
						return nil, tsr, false
					}

					// else, n is the node we want if it's a leaf
					if n.isLeaf {
						return n, false, true
					}

					tsr = len(n.children) == 1 && n.children[0].isLeaf && bytes.Equal(n.children[0].path, []byte("/"))
					return nil, tsr, false
				case catchAll:
					return n, false, true
				default:
					log.Panicf("invalid node type: %+v", n)
				}
			}
		} else if bytes.Equal(path, n.path) {
			if n.isLeaf {
				return n, false, true
			}

			// it seems that the case in below(comment) will never hapeen...
			//if path == "/" && n.wildChild && n.nType != root {
			//return nil, true, false
			//}

			// nothing found, check if a child with this path + a trailing slash exists
			for i := 0; i < len(n.indices); i++ {
				if n.indices[i] == '/' {
					n = n.children[i]
					tsr = len(n.path) == 1 && n.isLeaf
				}

				return nil, tsr, false
			}
		}

		// nothing found, e.g. URL is `/user/jhon/card/`, but request `/user/jhon/card`
		tsr = (bytes.Equal(path, []byte("/"))) ||
			(len(n.path) == len(path)+1 && n.path[len(path)] == '/' &&
				bytes.Equal(path, n.path[:len(n.path)-1]) && n.isLeaf)
		return nil, tsr, false
	}
}
