BINARY := awm
VERSION := $(shell cat VERSION)
MODULE := $(shell sed -n '1s/^module //p' go.mod)
LDFLAGS := -s -w -X $(MODULE)/internal/awm.Version=$(VERSION)

.PHONY: fmt fmt-check sync-assets test test-go test-python vet build release clean validate

fmt:
	gofmt -w ./cmd ./internal

fmt-check:
	@files="$$(gofmt -l ./cmd ./internal)"; if [ -n "$$files" ]; then echo "gofmt required:"; echo "$$files"; exit 1; fi

sync-assets:
	./scripts/sync-assets.sh

test: test-go test-python

test-go:
	go test ./...

test-python:
	python3 -m unittest discover -s tests/python -v

vet:
	go vet ./...

build:
	mkdir -p dist
	CGO_ENABLED=0 go build -trimpath -ldflags "$(LDFLAGS)" -o dist/$(BINARY) ./cmd/awm

release:
	mkdir -p dist
	rm -f dist/$(BINARY)
	go run ./cmd/awm-release -dist dist -version "$(VERSION)"

validate: fmt-check test vet build

clean:
	rm -rf dist .build coverage.out
