.PHONY: build clean fmt vet test bench cover profiling

COVERPROFILE=
DEBUG=
ARGS=-args -configPath=/tmp/guard.test.json -configAddr=":34567" -proxyAddr=":45678"

build: clean fmt vet test
	go build

clean:
	rm -f guard

fmt:
	go fmt ./...

vet:
	go tool vet -v .

test:
	go test -cover $(COVERPROFILE) -race $(DEBUG) $(ARGS)

bench:
	go test -bench=. -benchmem $(ARGS)

cover:
	$(eval COVERPROFILE += -coverprofile=coverage.out)
	go test -cover $(COVERPROFILE) -race $(ARGS) $(DEBUG)
	go tool cover -html=coverage.out
	rm -f coverage.out

profiling:
	go test -bench=. -cpuprofile cpu.out -memprofile mem.out $(ARGS)

release: clean fmt vet test
	GOOS=linux GOARCH=amd64 go build -o guard.linux_amd64
