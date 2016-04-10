.PHONY: install test clean release

install:
	go get github.com/kevinburke/goose/cmd/goose

test:
	go test ./...

release:
	go get github.com/Shyp/bump_version/bump_version
	bump_version minor cmd/goose/main.go
