# guard

[![Build Status](https://travis-ci.org/jiajunhuang/guard.svg?branch=master)](https://travis-ci.org/jiajunhuang/guard)
[![codecov](https://codecov.io/gh/jiajunhuang/guard/branch/master/graph/badge.svg)](https://codecov.io/gh/jiajunhuang/guard)

guard is a generic high performance circuit breaker & proxy written in Go. It has four major component:

- radix tree: which stores registed URLs(it's a customized version [httprouter](https://github.com/julienschmidt/httprouter))
- timeline bucket: which records response results
- load balancer: which distributes requests
- circuit breaker: which make sure your backend services will not breakdown by a large quantity of requests

## workflow

![workflow diagram](./workflow.png)

## benchmark

I've made a simple benchmark in my laptop(i5-3210M CPU @ 2.50GHz with 4 cores):

```bash
jiajun@idea ~: wrk --latency -H "Host: www.example.com" -c 1024 -d 30 -t 2 http://127.0.0.1:9999  # nginx
Running 30s test @ http://127.0.0.1:9999
  2 threads and 1024 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency   110.66ms  167.36ms   1.19s    94.54%
    Req/Sec     6.59k     1.02k    9.21k    62.71%
  Latency Distribution
     50%   78.99ms
     75%  153.75ms
     90%  177.76ms
     99%  986.89ms
  392348 requests in 30.06s, 318.03MB read
  Socket errors: connect 5, read 0, write 0, timeout 0
Requests/sec:  13053.00
Transfer/sec:     10.58MB
jiajun@idea ~: wrk --latency -H "Host: www.example.com" -c 1024 -d 30 -t 2 http://127.0.0.1:23456  # guard
Running 30s test @ http://127.0.0.1:23456
  2 threads and 1024 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency   142.97ms   45.58ms 390.70ms   73.94%
    Req/Sec     3.57k   559.23     4.80k    70.23%
  Latency Distribution
     50%  154.74ms
     75%  170.04ms
     90%  183.32ms
     99%  242.98ms
  212676 requests in 30.04s, 163.48MB read
  Socket errors: connect 5, read 0, write 0, timeout 0
Requests/sec:   7080.37
Transfer/sec:      5.44MB
```

for now, guard's proxy performance is about 54% of Nginx, but I'm stilling work on it! don't worry, it will
become better and better!

## TODO

- [x] radix tree(thanks @[httprouter](https://github.com/julienschmidt/httprouter))
- [x] timeline buckets for statistics
- [x] load balancer algorithm with weighted round robin
- [ ] load balancer algorithm with round robin
- [x] load balancer algorithm with random
- [x] circuit breaker
- [x] proxy server(thanks @[golang](https://golang.org/))
- [ ] dynamic configuration load & save
- [ ] graceful restart
- [ ] URL-level mutex(remove the bucket-level mutex to gain a better performance)
- [ ] fallback option while circuit breaker works(maybe serve a static html page, or return some words.)
- [ ] more test cases & benchmarks

## set it up

for now, it's a little bit inconvenient to setup, but here is the guide:

1. clone the source code using `go get -u`:

```bash
$ go get -u github.com/jiajunhuang/guard
```

2. compile it

```bash
$ $GOPATH/src/github.com/jiajunhuang/guard && make
```

3. start it

```bash
$ ./guard
```

4. now you need to register an application by send a POST request to `http://127.0.0.1:12345/app` with json like this:

```json
{
    "name": "www.example.com",
    "urls": ["/"],
    "methods": ["GET"],
    "backends": [
        {
            "host": "127.0.0.1",
            "port": 80,
            "weight": 5
        }
    ]
}
```

I'm doing it like this:

```bash
$ http POST :12345/app < backends.json 
HTTP/1.1 200 OK
Content-Length: 8
Content-Type: text/plain; charset=utf-8
Date: Sun, 21 Jan 2018 08:51:16 GMT

success!
```

5. and now, it works! whoops! try it:

```bash
$ http :23456 'Host: www.example.com'
HTTP/1.1 200 OK
...
```
