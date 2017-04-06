SHELL = /bin/bash

.PHONY: install test clean release

STATICCHECK := $(shell command -v staticcheck)
BUMP_VERSION := $(shell command -v bump_version)

test: vet
	go test ./...

install:
	go get -v github.com/kevinburke/goose/...

vet:
ifndef STATICCHECK
	go get -u github.com/dominikh/go-tools/cmd/staticcheck
endif
	go vet ./cmd/... ./lib/...
	staticcheck ./cmd/... ./lib/...

race-test: vet
	go test -v -race ./...

release: race-test
ifndef BUMP_VERSION
	go get -u github.com/Shyp/bump_version
endif
	bump_version minor cmd/goose/main.go
