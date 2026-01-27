APP_NAME := whosthere

GIT_TAG    := $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo unknown)
BUILD_DATE := $(shell date -u +"%Y-%m-%d")

# Local dev build ldflags: mimic GoReleaser's defaults and also set internal/version.
LDFLAGS := -s -w \
	-X main.versionStr=$(GIT_TAG) \
	-X main.commitStr=$(GIT_COMMIT) \
	-X main.dateStr=$(BUILD_DATE)

all: dev-deps fmt lint test build release-clean

# TODO: add cross-platform deps support
dev-deps:
	go mod tidy
	pipx install mdformat
	pipx inject mdformat mdformat-gfm
	go install github.com/shurcooL/markdownfmt@latest
	brew install goreleaser
	brew install golangci-lint
	brew upgrade golangci-lint

build:
	CGO_ENABLED=0 go build -ldflags '$(LDFLAGS)' -o $(APP_NAME) .

install: build
	CGO_ENABLED=0 go install -v ./...

lint:
	golangci-lint run
	mdformat --check .

fmt:
	gofmt -s -w -e .
	mdformat .

test:
	go test -v -cover -race -timeout=120s -parallel=10 ./...

# to test a goreleaser release locally without pushing anything
release-clean:
	goreleaser release --snapshot --clean

.PHONY: fmt lint test build install deps release-clean