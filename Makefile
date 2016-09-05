.PHONY: install test clean release

install:
	go get github.com/kevinburke/goose/cmd/goose

vet:
	go vet ./cmd/... ./lib/...

test:
	go test ./...

race-test: vet
	go test -race ./...

release: race-test
	go get -u github.com/Shyp/bump_version
	bump_version minor cmd/goose/main.go
