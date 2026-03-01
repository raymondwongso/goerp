-include .env
export

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

test:
	go test ./... -coverpkg=./... -coverprofile=coverage.txt.tmp
	grep -v -f .coverignore coverage.txt.tmp > coverage.txt
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

docs-up:
	docker run -d --name swagger-ui -p 8081:8080 \
		-e SWAGGER_JSON=/docs/openapi.yaml \
		-v $(CURDIR)/docs:/docs \
		swaggerapi/swagger-ui

docs-down:
	docker rm -f swagger-ui

goose-up:
	goose -dir migration postgres "$(GOOSE_DSN)" up

goose-down:
	goose -dir migration postgres "$(GOOSE_DSN)" down

gosec:
	gosec ./...

govulncheck:
	govulncheck ./...
