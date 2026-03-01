generate:
	go generate ./...

test:
	go test ./... -coverpkg=./... -coverprofile=coverage.txt
	go tool cover -func=coverage.txt

build:
	go build -o=bin/api ./cmd/api

run-api:
	go build -o=bin/api ./cmd/api
	./bin/api