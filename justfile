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

release-snapshot:
    docker compose run --rm release-snapshot

dev:
    docker compose run --rm dev bash

fmt:
    go fmt ./...

vet:
    go vet ./...
