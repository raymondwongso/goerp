test:
	go test ./... -coverpkg=./... -coverprofile=coverage.out
	go tool cover -func=coverage.out

build:
	go build -o=bin/api ./cmd/api

run-api:
	go build -o=bin/api ./cmd/api
	./bin/api