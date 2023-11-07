SHELL = /bin/bash -o pipefail

.PHONY: install test clean release

STATICCHECK := $(GOPATH)/bin/staticcheck
BUMP_VERSION := $(GOPATH)/bin/bump_version

test: lint
	go test ./...

install:
	go install -v github.com/kevinburke/goose/...@latest

$(STATICCHECK):
	go install honnef.co/go/tools/cmd/staticcheck@latest

lint: $(STATICCHECK)
	go vet ./...
	$(STATICCHECK) ./...

race-test: lint
	go test -v -race ./...

$(BUMP_VERSION):
	go install github.com/kevinburke/bump_version@latest

release: race-test | $(BUMP_VERSION)
	$(BUMP_VERSION) minor cmd/goose/main.go
