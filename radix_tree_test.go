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

	if n.leaf != e.isLeaf {
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

	n.AddRoute("/user/hello", GET, POST)
	checkNodeValid(
		t, n,
		nodeExpceted{"/user/hello", root, GET | POST, false, true, 0, true, false},
	)

	n.AddRoute("/user/world", DELETE)
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

	n.AddRoute("/user/:name/hello", GET)
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

	n.AddRoute("/user/:name/hello", GET)
	checkNodeValid(
		t, n,
		nodeExpceted{"/user/", root, NONE, true, true, 1, false, true},
	)

	n.AddRoute("/user/:name/hello/:card", GET)
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
	n.AddRoute("/user/:name/hello/world")
	n.AddRoute("/user/*whoever")
}

func TestAddRouteMultiIndices(t *testing.T) {
	n := &node{}
	n.AddRoute("/user/:name/hello/world")
	n.AddRoute("/use/this")
	n.AddRoute("/usea/this")
	n.AddRoute("/useb/that")
	n.AddRoute("/usea/that")
}
