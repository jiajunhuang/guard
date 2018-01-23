package main

// Router is a fake struct for now, just for compiling
type Router struct {
	MethodNotAllowed      bool  // raise 405 if not corresponding method not found?
	RedirectTrailingSlash bool  // redirect if url is not right
	tree                  *node // root of tree
}

func NewRouter() *Router {
	return &Router{}
}

func (r *Router) ByPath(method string, url string) (*node, bool, bool) {
	return nil, false, false
}

func (r *Router) GET(path string)    {}
func (r *Router) POST(path string)   {}
func (r *Router) DELETE(path string) {}
func (r *Router) PUT(path string)    {}
