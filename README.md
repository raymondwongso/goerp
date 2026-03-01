# goerp

[![codecov](https://codecov.io/gh/raymondwongso/goerp/branch/main/graph/badge.svg)](https://codecov.io/gh/raymondwongso/goerp)

An ERP (Enterprise Resource Planning) backend written in Go. goerp provides core business modules — authentication, authorization, and more — built on a clean hexagonal architecture with PostgreSQL as the primary data store.

## Requirements

- Go 1.26+
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
make install-tools
```

**3. Configure environment:**

Copy `.env.example` to `.env` and fill in the values:

```bash
cp .env.example .env
```

| Variable | Description |
|---|---|
| `POSTGRES_USER` | PostgreSQL username |
| `POSTGRES_PASSWORD` | PostgreSQL password |
| `POSTGRES_DB` | PostgreSQL database name |
| `POSTGRES_PORT` | PostgreSQL port (default `5432`) |
| `GOOGLE_CLIENT_ID` | Google OAuth2 client ID |
| `GOOGLE_CLIENT_SECRET` | Google OAuth2 client secret |
| `GOOGLE_REDIRECT_URL` | OAuth2 redirect URI registered in Google Cloud Console |
| `API_PORT` | HTTP server listen address (default `:8080`) |

**4. Start Dependencies:**

```bash
make compose-up
```

Stop with:

```bash
make compose-down
```

**5. Run migrations:**

```bash
make goose-up
```

To roll back one step:

```bash
make goose-down
```

**6. Run the API server:**

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

## API Documentation

The OpenAPI 3.0 spec lives at [`docs/openapi.yaml`](docs/openapi.yaml).

Start a local Swagger UI server (requires Docker):

```bash
make docs-up    # serves at http://localhost:8081
make docs-down  # stop and remove the container
```

## Contributing

### Prerequisites

Install all required tools (goose, gosec, govulncheck, mockgen) with a single command:

```bash
make install-tools
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

### Security Checks

Run static security analysis locally:

```bash
make gosec
```

Run dependency vulnerability check locally:

```bash
make govulncheck
```

### Submitting Changes

1. Fork the repository and create a branch from `main`.
2. Make your changes and ensure all tests pass: `make test`
3. Ensure the build succeeds: `make build`
4. Run security checks: `make gosec` and `make govulncheck`
5. Open a pull request against `main` with a clear description of the change.
