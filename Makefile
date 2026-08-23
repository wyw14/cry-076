.PHONY: build run migrate test race vet web-install web-test web-build verify

build:
	go build ./...

run:
	go run ./cmd/server

migrate:
	go run ./cmd/migrate ./migrations

test:
	go test ./...

race:
	go test -race ./...

vet:
	go vet ./...

web-install:
	cd web && npm ci

web-test:
	cd web && npm test -- --run

web-build:
	cd web && npm run build

verify: build test race vet web-test web-build
