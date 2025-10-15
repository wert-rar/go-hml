.PHONY: build run test

build:
	go build -o bin/go-hml ./cmd/go-hml.

run: build
	go run ./cmd/go-hml --path=. --ext=.go,.py --exclude=.git,vendor

test:
	go test ./...