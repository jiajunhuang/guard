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
$ wrk --latency -H "Host: www.example.com" -c 2048 -d 30 -t 2 http://127.0.0.1:23456
Running 30s test @ http://127.0.0.1:23456
  2 threads and 2048 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency   129.72ms   47.00ms 357.08ms   73.62%
    Req/Sec     3.73k   573.26     5.10k    73.75%
  Latency Distribution
     50%  145.10ms
     75%  162.08ms
     90%  173.55ms
     99%  234.31ms
  221712 requests in 30.03s, 170.42MB read
  Socket errors: connect 1029, read 0, write 0, timeout 0
Requests/sec:   7382.58
Transfer/sec:      5.67MB
$ wrk --latency -H "Host: www.example.com" -c 2048 -d 30 -t 2 http://127.0.0.1:9999
Running 30s test @ http://127.0.0.1:9999
  2 threads and 2048 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency    98.71ms  160.47ms   1.19s    95.27%
    Req/Sec     6.68k     1.34k    9.53k    55.20%
  Latency Distribution
     50%   52.35ms
     75%  139.05ms
     90%  183.88ms
     99%  965.43ms
  396909 requests in 30.05s, 321.73MB read
  Socket errors: connect 1029, read 0, write 0, timeout 13
Requests/sec:  13208.58
Transfer/sec:     10.71MB
```

for now, guard's proxy performance is about 55% of Nginx, but I'm stilling work on it! don't worry, it will
become better and better!

by the way, thanks the [suggestion](https://github.com/jiajunhuang/guard/issues/15) from [@dongzerun](https://github.com/dongzerun),
by configure the `GOGC` in environment, guard's proxy performance is about 70% of Nginx.

```bash
$ wrk --latency -H "Host: www.example.com" -c 2048 -d 30 -t 2 http://127.0.0.1:23456  # guard, by setting environment variable `GOGC=1024`
Running 30s test @ http://127.0.0.1:23456
  2 threads and 2048 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency   106.76ms   37.07ms 355.80ms   72.19%
    Req/Sec     4.70k   806.66     6.26k    63.93%
  Latency Distribution
     50%  117.45ms
     75%  128.71ms
     90%  141.26ms
     99%  187.43ms
  278789 requests in 30.03s, 214.29MB read
  Socket errors: connect 1029, read 0, write 0, timeout 0
Requests/sec:   9283.34
Transfer/sec:      7.14MB
```

## TODO

- [x] radix tree(thanks @[httprouter](https://github.com/julienschmidt/httprouter))
- [x] timeline buckets for statistics
- [x] load balancer algorithm with weighted round robin
- [x] load balancer algorithm with round robin
- [x] load balancer algorithm with random
- [x] circuit breaker
- [x] proxy server(thanks @[golang](https://golang.org/))
- [ ] dynamic configuration load & save
- [ ] graceful restart
- [x] ~~URL-level mutex(remove the bucket-level mutex to gain a better performance)~~ the statistics module is lock-free now
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
