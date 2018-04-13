SHELL = /bin/bash -o pipefail

.PHONY: install test clean release

MEGACHECK := $(GOPATH)/bin/megacheck
BUMP_VERSION := $(GOPATH)/bin/bump_version

test: vet
	go list ./... | grep -v vendor | xargs go test

install:
	go get -v github.com/kevinburke/goose/...

$(MEGACHECK):
	go get -u honnef.co/go/tools/cmd/megacheck

vet: $(MEGACHECK)
	go vet ./cmd/... ./lib/...
	$(MEGACHECK) --ignore='github.com/kevinburke/goose/lib/goose/*.go:S1002' ./cmd/... ./lib/...

race-test: vet
	go list ./... | grep -v vendor | xargs go test -v -race

$(BUMP_VERSION):
	go get -u github.com/kevinburke/bump_version

release: race-test | $(BUMP_VERSION)
	$(BUMP_VERSION) minor cmd/goose/main.go
