.PHONY: build clean fmt vet test bench cover profiling

COVERPROFILE=
DEBUG=
ARGS=

build: clean fmt vet test
	go build

clean:
	rm -f guard

fmt:
	go fmt ./...

vet:
	go tool vet -v .

test:
	go test -cover $(COVERPROFILE) -race $(DEBUG)

bench:
	go test -bench=. -benchmem $(ARGS)

cover:
	$(eval COVERPROFILE += -coverprofile=coverage.out)
	go test -cover $(COVERPROFILE) -race $(ARGS) $(DEBUG)
	go tool cover -html=coverage.out
	rm -f coverage.out

profiling:
	go test -bench=. -cpuprofile cpu.out -memprofile mem.out $(ARGS)
