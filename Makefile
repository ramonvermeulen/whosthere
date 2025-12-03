default: fmt lint install test

build:
	go build

install: build
	go install -v ./...

lint:
	golangci-lint run

fmt:
	gofmt -s -w -e .

test:
	go test -v -cover -race -timeout=120s -parallel=10 ./...


.PHONY: fmt lint test build install
