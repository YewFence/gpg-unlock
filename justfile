# gpg-unlock

run *ARGS:
    go run . {{ARGS}}

build:
    go build -o gpg-unlock .

build-all:
    docker compose run --rm build

dev:
    docker compose run --rm dev bash

fmt:
    go fmt ./...

vet:
    go vet ./...
