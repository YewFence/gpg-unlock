# gpg-unlock

run *ARGS:
    go run . {{ARGS}}

build:
    go build -o gpg-unlock .

install:
    go install .

uninstall:
    go run . reset
    go clean -i .

build-all:
    docker compose run --rm build

dev:
    docker compose run --rm dev bash

fmt:
    go fmt ./...

vet:
    go vet ./...
