SHELL = /bin/bash -o pipefail

.PHONY: install test clean release lint race-test diff-vendor

GOBIN := $(shell go env GOBIN)
ifeq ($(GOBIN),)
GOBIN := $(shell go env GOPATH)/bin
endif

STATICCHECK := $(GOBIN)/staticcheck
BUMP_VERSION := $(GOBIN)/bump_version
DIFFER := $(GOBIN)/differ

test: lint
	go test ./...

install:
	go install -v github.com/kevinburke/goose/...@latest

$(STATICCHECK):
	go install honnef.co/go/tools/cmd/staticcheck@latest

$(DIFFER):
	go install github.com/kevinburke/differ@latest

lint: $(STATICCHECK)
	go vet ./...
	$(STATICCHECK) ./...

# Verify go.mod/go.sum and vendor/ are in sync with the source tree.
# Fails CI if "go mod tidy" or "go mod vendor" would produce a diff.
diff-vendor: | $(DIFFER)
	$(DIFFER) go mod tidy
	$(DIFFER) go mod vendor

race-test: lint diff-vendor
	go test -v -race ./...

$(BUMP_VERSION):
	go install github.com/kevinburke/bump_version@latest

release: race-test | $(BUMP_VERSION)
	$(BUMP_VERSION) minor cmd/goose/main.go
