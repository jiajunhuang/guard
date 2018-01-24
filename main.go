package main

import (
	"net/http"
)

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
	backend1 := Backend{"127.0.0.1", 80, 5}
	backend2 := Backend{"127.0.0.1", 80, 1}
	backend3 := Backend{"127.0.0.1", 80, 1}
	appName := "www.example.com"

	breaker := NewBreaker()
	breaker.apps[appName] = NewApp(
		NewWRR(backend1, backend2, backend3), true,
	)

	breaker.apps[appName].AddRoute("/", "GET")

	http.ListenAndServe(":23456", breaker)
}
