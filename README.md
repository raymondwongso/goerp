# goerp

[![codecov](https://codecov.io/gh/raymondwongso/goerp/branch/main/graph/badge.svg)](https://codecov.io/gh/raymondwongso/goerp)

An ERP (Enterprise Resource Planning) backend written in Go. goerp provides core business modules — authentication, authorization, and more — built on a clean hexagonal architecture with PostgreSQL as the primary data store.

## Requirements

- Go 1.25+
- PostgreSQL 18

## Running Locally

**1. Clone the repository:**

```bash
git clone https://github.com/raymondwongso/goerp.git
cd goerp
```

**2. Install dependencies:**

```bash
go mod download
```

**3. Set up the database:**

Run migrations using [goose](https://github.com/pressly/goose):

```bash
goose -dir migration postgres "host=localhost port=5432 user=postgres password=postgres dbname=goerp sslmode=disable" up
```

**4. Run the API server:**

```bash
make run-api
```

## Running Tests

Run all tests with coverage:

```bash
make test
```

Run a specific package or test:

```bash
go test ./domain/xhttp/... -v
go test ./example/usecase/submodulea/... -run TestCreate
```

Coverage output is written to `coverage.txt`. To view it in the browser:

```bash
go tool cover -html=coverage.txt
```

## Contributing

### Prerequisites

Install [mockgen](https://github.com/uber-go/mock) for generating mocks:

```bash
go install go.uber.org/mock/mockgen@latest
```

### Project Structure

```
goerp/
├── cmd/api/         # Application entrypoint
├── domain/          # Shared domain structs and interfaces
├── migration/       # goose migration files
├── <module>/        # Business module (e.g. auth, example)
│   ├── interfaces.go
│   ├── <module>.go
│   ├── http/
│   ├── store/postgres/
│   ├── usecase/<submodule>/
│   └── mock/
└── scripts/         # Utility scripts
```

Each business module follows hexagonal architecture: HTTP handlers depend on use case interfaces, use cases depend on store interfaces — never on concrete implementations.

### Adding a New Module

1. Create the module directory following the structure above.
2. Define all interfaces in `<module>/interfaces.go`.
3. Implement use cases in `<module>/usecase/<submodule>/`.
4. Implement the store layer in `<module>/store/postgres/`.
5. Implement HTTP handlers in `<module>/http/`.
6. Generate mocks from the module root: `go generate ./<module>/...`
7. Wire everything in `<module>/<module>.go` and register in `cmd/api`.

### Guidelines

- **No foreign keys** — referential integrity is enforced in the application layer.
- **UUIDv7** for all primary keys.
- **No cascade deletes** — deletions must be handled explicitly.
- All errors must use `domain/xerror`. Do not return raw errors across layer boundaries.
- Every use case and store method must have a corresponding unit test.
- Mocks are generated via `go generate` — do not edit them manually.

### Submitting Changes

1. Fork the repository and create a branch from `main`.
2. Make your changes and ensure all tests pass: `make test`
3. Ensure the build succeeds: `make build`
4. Open a pull request against `main` with a clear description of the change.
