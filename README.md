# guard

[![Build Status](https://travis-ci.org/jiajunhuang/guard.svg?branch=master)](https://travis-ci.org/jiajunhuang/guard)

guard is a generic high performance circuit breaker written in Go. It has four major component:

- radix tree & mux: which stores registed URLs(it's a customized version [httprouter](https://github.com/julienschmidt/httprouter))
- timeline bucket: which records response results
- load balancer: which distributes requests
- circuit breaker: which make sure your backend services will not breakdown by a large quantity of requests
