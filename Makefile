.PHONY: install test clean

install:
	go get -u github.com/kevinburke/goose/cmd/goose

test:
	go test ./...
