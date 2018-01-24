package main

import (
	"testing"
)

type nodeExpceted struct {
	path           string
	nType          nodeType
	methods        HTTPMethod
	wildChild      bool
	indicesIsEmpty bool
	childrenNum    int
	isLeaf         bool
	statusIsNil    bool
}

func checkNodeValid(t *testing.T, n *node, e nodeExpceted) {
	if n == nil {
		t.Fatalf("checkNodeValid: both node and expect should not be nil")
	}

	if n.path != e.path {
		t.Errorf("n.path should be %s, but n is: %+v", e.path, n)
	}

	if n.nType != e.nType {
		t.Errorf("n.nType should be %d, but n is: %+v", e.nType, n)
	}

	if n.methods != e.methods {
		t.Errorf("n.methods should be %x, but n is: %+v", e.methods, n)
	}

	if n.wildChild != e.wildChild {
		t.Errorf("n.wildChild should be %t, but n is: %+v", e.wildChild, n)
	}

	if !(len(n.indices) > 0 && !e.indicesIsEmpty || len(n.indices) == 0 && e.indicesIsEmpty) {
		t.Errorf("n.indices should be empty? %t, but n is: %+v", e.indicesIsEmpty, n)
	}

	if len(n.children) != e.childrenNum {
		t.Errorf("n should have %d childrens, but n is: %+v", e.childrenNum, n)
	}

	if n.isLeaf != e.isLeaf {
		t.Errorf("n.leaf should be %t, but n is: %+v", e.isLeaf, n)
	}

	if !(n.status == nil && e.statusIsNil || n.status != nil && !e.statusIsNil) {

		t.Errorf("n.status should be nil? %t, but n is: %+v", e.statusIsNil, n)
	}
}

func TestMin(t *testing.T) {
	if 1 != min(1, 2) {
		t.Error("min(1, 2) should return 1")
	}

	if 1 != min(2, 1) {
		t.Error("min(2, 1) should return 1")
	}
}

func TestSetMethods(t *testing.T) {
	n := &node{}

	if n.methods != NONE {
		t.Errorf("n.methods should be NONE, but got: %x", n.methods)
	}

	methods := []HTTPMethod{GET, POST, PUT, DELETE, HEAD, OPTIONS, CONNECT, TRACE, PATCH}
	for _, m := range methods {
		n.setMethods(m)

		if !n.hasMethod(m) {
			t.Errorf("n should have HTTP method %x, but got: %x", m, n.methods)
		}
	}
}

func TestInsertLeaf(t *testing.T) {
	n := &node{}

	n.insertChild("this", "/use/this")

	checkNodeValid(
		t, n,
		nodeExpceted{"this", static, NONE, false, true, 0, true, false},
	)

	n.insertChild("this", "/use/this", GET)

	if !n.hasMethod(GET) {
		t.Error("n should have HTTP method `GET` been set, but not")
	}
}

func TestInsertChild(t *testing.T) {
	n := &node{}

	n.insertChild("/:name", "/user/:name", GET)

	checkNodeValid(
		t, n,
		nodeExpceted{"/", static, NONE, true, true, 1, false, true},
	)

	// check it's child, then
	n = n.children[0]

	checkNodeValid(
		t, n,
		nodeExpceted{":name", param, GET, false, true, 0, true, false},
	)
}

func shouldPanic() {
	if err := recover(); err == nil {
		panic("should panic but not")
	}
}

func TestInsertBadParamDualWildchard(t *testing.T) {
	defer shouldPanic()

	n := &node{}

	n.insertChild("/:name:this", "/user/:name:this/there")
}

func TestInsertBadParamNoParamName(t *testing.T) {
	defer shouldPanic()

	n := &node{}

	n.insertChild("/:", "/user/:/there")
}

func TestInsertBadParamConflict(t *testing.T) {
	defer shouldPanic()

	n := &node{}

	n.insertChild("/:name", "/user/:name/there")
	n.insertChild("/:name", "/user/:name/there")
}

func TestInsertDualParam(t *testing.T) {
	n := &node{}

	n.insertChild("/:name/:card", "/user/:name/:card", POST)

	slash1 := n
	name := n.children[0]
	slash2 := name.children[0]
	card := slash2.children[0]

	checkNodeValid(
		t, slash1,
		nodeExpceted{"/", static, NONE, true, true, 1, false, true},
	)
	checkNodeValid(
		t, name,
		nodeExpceted{":name", param, NONE, false, true, 1, false, true},
	)
	checkNodeValid(
		t, slash2,
		nodeExpceted{"/", static, NONE, true, true, 1, false, true},
	)
	checkNodeValid(
		t, card,
		nodeExpceted{":card", param, POST, false, true, 0, true, false},
	)
}

func TestInsertChildCatchAll(t *testing.T) {
	n := &node{}

	n.insertChild("/user/*name", "/user/*name", POST)

	// first, check n itself
	checkNodeValid(
		t, n,
		nodeExpceted{"/user/", static, NONE, true, true, 1, false, true},
	)

	// last, the child
	n = n.children[0]
	checkNodeValid(
		t, n,
		nodeExpceted{"*name", catchAll, POST, false, true, 0, true, false},
	)
}

func TestInsertCatchAllMultiTimes(t *testing.T) {
	defer shouldPanic()

	n := &node{}
	n.insertChild("/*name/:haha", "/*name/:haha")
}

func TestInsertCatchAllNoSlash(t *testing.T) {
	defer shouldPanic()

	n := &node{}
	n.insertChild("/user*name", "/user*name")
}

func TestAddRoute(t *testing.T) {
	n := &node{}

	n.addRoute("/user/hello", GET, POST)
	checkNodeValid(
		t, n,
		nodeExpceted{"/user/hello", root, GET | POST, false, true, 0, true, false},
	)

	n.addRoute("/user/world", DELETE)
	checkNodeValid(
		t, n,
		nodeExpceted{"/user/", root, NONE, false, false, 2, false, true},
	)

	hello := n.children[0]
	world := n.children[1]
	if hello.path == "world" {
		hello, world = world, hello
	}

	checkNodeValid(
		t, hello,
		nodeExpceted{"hello", static, GET | POST, false, true, 0, true, false},
	)
	checkNodeValid(
		t, world,
		nodeExpceted{"world", static, DELETE, false, true, 0, true, false},
	)
}

func TestAddRouteWildChild(t *testing.T) {
	n := &node{}

	n.addRoute("/user/:name/hello", GET)
	checkNodeValid(
		t, n,
		nodeExpceted{"/user/", root, NONE, true, true, 1, false, true},
	)

	name := n.children[0]
	checkNodeValid(
		t, name,
		nodeExpceted{":name", param, NONE, false, true, 1, false, true},
	)

	hello := name.children[0]
	checkNodeValid(
		t, hello,
		nodeExpceted{"/hello", static, GET, false, true, 0, true, false},
	)
}

func TestAddRouteDualWildChild(t *testing.T) {
	n := &node{}

	n.addRoute("/user/:name/hello", GET)
	checkNodeValid(
		t, n,
		nodeExpceted{"/user/", root, NONE, true, true, 1, false, true},
	)

	n.addRoute("/user/:name/hello/:card", GET)
	checkNodeValid(
		t, n,
		nodeExpceted{"/user/", root, NONE, true, true, 1, false, true},
	)

	name := n.children[0]
	checkNodeValid(
		t, name,
		nodeExpceted{":name", param, NONE, false, true, 1, false, true},
	)

	slashHello := name.children[0]
	checkNodeValid(
		t, slashHello,
		nodeExpceted{"/hello", static, GET, false, false, 1, true, false},
	)

	slash := slashHello.children[0]
	checkNodeValid(
		t, slash,
		nodeExpceted{"/", static, NONE, true, true, 1, false, true},
	)

	card := slash.children[0]
	checkNodeValid(
		t, card,
		nodeExpceted{":card", param, GET, false, true, 0, true, false},
	)
}

func TestAddRouteWildParamConflict(t *testing.T) {
	defer shouldPanic()

	n := &node{}
	n.addRoute("/user/:name/hello/world")
	n.addRoute("/user/*whoever")
}

func TestAddRouteMultiIndices(t *testing.T) {
	n := &node{}
	n.addRoute("/user/:name/hello/world")
	n.addRoute("/use/this")
	n.addRoute("/usea/this")
	n.addRoute("/useb/that")
	n.addRoute("/usea/that")
}

func TestAddRouteSamePath(t *testing.T) {
	n := &node{}

	n.addRoute("/user/hello", GET, POST)
	checkNodeValid(
		t, n,
		nodeExpceted{"/user/hello", root, GET | POST, false, true, 0, true, false},
	)

	n.addRoute("/user/hello", DELETE)
	checkNodeValid(
		t, n,
		nodeExpceted{"/user/hello", root, GET | POST | DELETE, false, true, 0, true, false},
	)
}

type byPathExpected struct {
	node  *node
	tsr   bool
	found bool
}

func checkByPath(t *testing.T, n *node, tsr bool, found bool, e byPathExpected) {
	if n != e.node {
		t.Errorf("node should be %+v, but got: %+v", e.node, n)
	}

	if tsr != e.tsr {
		t.Errorf("tsr should be %t, but got: %t", e.tsr, tsr)
	}

	if found != e.found {
		t.Errorf("found should be %t, but got: %t", e.found, found)
	}
}

func TestByPath(t *testing.T) {
	n := &node{}
	n.addRoute("/user", GET, DELETE)

	checkNodeValid(
		t, n,
		nodeExpceted{"/user", root, GET | DELETE, false, true, 0, true, false},
	)

	nd, tsr, found := n.byPath("/user")
	checkByPath(t, nd, tsr, found, byPathExpected{n, false, true})

	nd, tsr, found = n.byPath("/user/")
	checkByPath(t, nd, tsr, found, byPathExpected{nil, true, false})

	nd, tsr, found = n.byPath("/what???")
	checkByPath(t, nd, tsr, found, byPathExpected{nil, false, false})

	n = &node{}
	n.addRoute("/user/", GET, DELETE)
	n.addRoute("/usera", GET, DELETE)

	checkNodeValid(
		t, n,
		nodeExpceted{"/user", root, NONE, false, false, 2, false, true},
	)
	nd, tsr, found = n.byPath("/user")
	checkByPath(t, nd, tsr, found, byPathExpected{nil, true, false})
}

func TestByPathWithWildchild(t *testing.T) {
	n := &node{}
	n.addRoute("/user/:name/hello", GET, DELETE)
	n.addRoute("/use/:this/that", GET, DELETE)

	checkNodeValid(
		t, n,
		nodeExpceted{"/use", root, NONE, false, false, 2, false, true},
	)

	nd, tsr, found := n.byPath("/user/jhon")
	checkByPath(t, nd, tsr, found, byPathExpected{nil, false, false})

	nd, tsr, found = n.byPath("/user/jhon/hello/")
	checkByPath(t, nd, tsr, found, byPathExpected{nil, true, false})
}

func TestByPathParamAndCatchAll(t *testing.T) {
	n := &node{}
	n.addRoute("/user/:name", GET, DELETE)

	nd, tsr, found := n.byPath("/user/jhon")
	checkByPath(t, nd, tsr, found, byPathExpected{n.children[0], false, true})
	nd, tsr, found = n.byPath("/user/jhon/")
	checkByPath(t, nd, tsr, found, byPathExpected{nil, true, false})

	n = &node{}
	n.addRoute("/user/*name", GET, DELETE)

	nd, tsr, found = n.byPath("/user/jhon")
	checkByPath(t, nd, tsr, found, byPathExpected{n.children[0], false, true})
}

func TestByPathBadNode(t *testing.T) {
	defer shouldPanic()

	n := &node{}
	n.addRoute("/user/:name", GET, DELETE)
	child := n.children[0]
	child.nType = static

	n.byPath("/user/jhon")
}
