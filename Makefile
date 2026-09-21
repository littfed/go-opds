.PHONY: all test test-coverage bench lint lint-fix vet vuln build clean

all: test lint

test:
	go test -v -race ./...

bench:
	go test -bench=. -benchmem ./...

test-coverage:
	go test -v -race -coverprofile=coverage.out -covermode=atomic ./...

lint:
	golangci-lint run ./...

lint-fix:
	golangci-lint run --fix ./...

vet:
	go vet ./...

vuln:
	govulncheck ./...

build:
	go build -o bin/basic ./examples/basic
	go build -o bin/client ./examples/client

clean:
	go clean -cache -testcache

