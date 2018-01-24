package main

/*
guard is a high performance circuit breaker written in Go.

workflow:

1. register URL patterns to router
2. find if router exist by HTTP `Host` field, if not found, return 404
3. request -> query router
            \
             -> (handler not exist?) -> return 404
             -> (handler exist but method not allowed?) -> return 405
             -> (handler exist)
                                \
                                 -> query timeline, circuit breaker not open yet? -> proxy and return, then save the response status
                                 -> circuot breaker is open? return 429 too many requests
*/

func main() {
	n := &node{}

	n.addRoute("/user/:name/hello", GET, DELETE)
	n.addRoute("/use/:this/that", GET, DELETE)

	n.byPath("/user/jhon")
}
