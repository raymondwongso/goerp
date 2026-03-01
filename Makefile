-include .env

GOOSE_DSN := host=localhost port=$(POSTGRES_PORT) user=$(POSTGRES_USER) password=$(POSTGRES_PASSWORD) dbname=$(POSTGRES_DB) sslmode=disable

generate:
	go generate ./...

install-goose:
	go install github.com/pressly/goose/v3/cmd/goose@v3.27.0

install-gosec:
	go install github.com/securego/gosec/v2/cmd/gosec@v2.24.6

install-govulncheck:
	go install golang.org/x/vuln/cmd/govulncheck@v1.1.4

install-mockgen:
	go install go.uber.org/mock/mockgen@latest

install-tools: install-goose install-gosec install-govulncheck install-mockgen

COVER_PKGS := $(shell go list ./... | grep -Ev 'github\.com/raymondwongso/goerp/(cmd/api|auth)$$' | tr '\n' ',' | sed 's/,$$//')

test:
	go test ./... -coverpkg=$(COVER_PKGS) -coverprofile=coverage.txt
	go tool cover -func=coverage.txt

build:
	CGO_ENABLED=0 go build -ldflags="-s -w" -o bin/api ./cmd/api

run-api:
	CGO_ENABLED=0 go build -ldflags="-s -w" -o bin/api ./cmd/api
	./bin/api

compose-up:
	docker compose up -d

compose-down:
	docker compose down

goose-up:
	goose -dir migration postgres "$(GOOSE_DSN)" up

goose-down:
	goose -dir migration postgres "$(GOOSE_DSN)" down

gosec:
	gosec ./...

govulncheck:
	govulncheck ./...
