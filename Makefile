.PHONY: install test clean release

install:
	go get github.com/kevinburke/goose/cmd/goose

test:
	go test ./...

race-test:
	go test -race ./...

release:
	go get -u github.com/Shyp/bump_version
	bump_version minor cmd/goose/main.go
