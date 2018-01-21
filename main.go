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

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

var (
	// global variable
	breaker = NewBreaker()
)

// APP registration infomation
type APP struct {
	Name     string    `json:"name"`
	URLs     []string  `json:"urls"`
	Methods  []string  `json:"methods"`
	Backends []Backend `json:"backends"`
}

func fakeProxyHandler(w http.ResponseWriter, r *http.Request, _ Params) {}

func overrideAPP(breaker *Breaker, app APP) {
	breaker.UpdateAPP(app.Name)
	router := breaker.routers[app.Name]

	for i, url := range app.URLs {
		router.Handle(strings.ToUpper(app.Methods[i]), url, fakeProxyHandler)
	}
	breaker.balancers[app.Name] = NewWRR(app.Backends...)
}

func createAPPHandler(w http.ResponseWriter, r *http.Request, _ Params) {
	var app APP
	var err error

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("failed" + err.Error()))
		return
	}
	if err = json.Unmarshal(body, &app); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("failed" + err.Error()))
		return
	}
	if len(app.Methods) != len(app.URLs) {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "failed: methods and urls should have same length and 1:1")
		return
	}

	log.Printf("gonna insert or over write app %s's configuration", app.Name)
	overrideAPP(breaker, app)

	fmt.Fprintf(w, "success!")
}

func proxy() {
	log.Fatal(http.ListenAndServe(":23456", breaker))
}

func main() {
	router := NewRouter()
	router.POST("/app", createAPPHandler)

	go proxy()
	log.Fatal(http.ListenAndServe(":12345", router))
}
