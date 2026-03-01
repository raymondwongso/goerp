# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Prompt Response Rules
1. Avoid overly described things that you have done. Only recap as necessary.

## Code Structure

The root directory is organized into **modules** and **special folders**.

**modules**:
- `auth` — Authentication and authorization
- `domain` — Domain objects such as struct. May have subdomain, e.g: `google` subdomain contains google related struct. Contains interfaces too. Domain may not import non-domain modules
- `example` — Example module as base references

**Special folders** (not modules):
- `bin/` — compiled binary output
- `cmd/` — application entrypoints (main packages), with each subfolders representing new application.
- `migration/` — database migration files
- `scripts/` — utility scripts (e.g., seed data)

Each module follow below convention:

```
<module>/
├── interfaces.go        # Interfaces for usecases in the module
├── <module>.go          # Glue/registration code: wires handlers, use cases, and repositories
├── store/
│   └── postgres/        # PostgreSQL implementation of repository layer
├── usecase/<submodule>/ # Business logic — one package and one struct per use case
├── http/                # HTTP handlers (other protocols get their own folder: grpc/, kafka/, etc.)
└── mock/                # MockGen-generated mocks (do not edit manually)
```

## Architecture

This project follows **Hexagonal (Clean) Architecture**. The `example/` module is the canonical reference implementation for all new modules.

Each layer depends only on the interfaces exposed by the layer below — never on concrete implementations:

```
HTTP Handler → Use Case (via interface) → Store (via interface) → PostgreSQL
```

All interfaces are defined in the module root `interfaces.go`, not in subpackages.

## Commands

```bash
# Build
make build        # Compiles to bin/api

# Run
make run-api      # Build and run the API server

# Test (all packages with coverage)
make test         # Equivalent to: go test ./... -coverpkg=./... -coverprofile=coverage.out

# Run a single test
go test ./example/usecase/submodulea/... -run TestCreate
go test ./domain/xerror/... -run TestGetCode

# Generate mocks (run from within the module directory)
go generate ./example/...
```

> This repo may contain multiple applications. Each application has its own `build` and `run` targets in the Makefile.

## Testing

All tests use:
- **[testify](https://github.com/stretchr/testify)** — assertions and test suites
- **[gomock](https://github.com/uber-go/mock)** (uber-go fork) — mock expectations
- **[mockgen](https://github.com/uber-go/mock/tree/main/mockgen)** — generates mock implementations via `go generate`

Patterns by layer:
- **Use cases:** mock store interfaces using generated mocks in `mock/`
- **Store layer:** use `go-sqlmock` to mock `*sql.DB`; pass `sqlx.NewDb(mockDB, "sqlmock")` as the queryer
- **HTTP handlers:** mock use case interfaces; use `httptest.NewRecorder` and `httptest.NewRequest`
- Test files live alongside source files (`*_test.go` in the same package)

## Error Handling

All errors use `domain/xerror`:

```go
// Return a typed error
return xerror.New(xerror.CodeNotFound, "example not found")

// Wrap a cause
return xerror.NewWithCause(xerror.CodeInternal, "db failure", err)

// Add field-level validation details
err = xerror.AddDetail(err.(xerror.Error), "field_name", "is required")

// Check code anywhere in the call chain
xerror.GetCode(err) // returns CodeUnknown if no xerror in chain
```

Mapping `xerror.Code` to http status code can be done in `domain/xhttp/mapper.go`. Create new mapper in each usecase when you need a custom mapper.

## Database

PostgreSQL version: 18

- **Driver:** PostgreSQL via [`jmoiron/sqlx`](https://github.com/jmoiron/sqlx)
- **Migrations:** [`pressly/goose`](https://github.com/pressly/goose) format in `migration/` — filenames prefixed with `00001_`, `00002_`, etc.
- **Foreign Key:** Do not use foreign key. Data integration is handled in application layer
- **UUID:** Use UUIDv7
- **On Cascade Delete:** Do not use cascade delete

## Domain Conventions

### Struct Location
- All structs belong in the `domain` package (e.g., `domain/user.go`, `domain/session.go`)
- Vendor-specific structs (Google, Slack, etc.) go in a subdomain package: `domain/google/`, `domain/slack/`, etc.

### Repository Interface Location
- Store/repository interfaces are defined in the same file as the model they operate on
- Example: `UserWriter` interface lives in `domain/user.go`, `OAuthStateWriter` in `domain/oauth.go`

### Store Layer Naming
- Use straightforward, conventional method names: `Get`, `List`, `Create`, `Insert`, `Update`, `Upsert`, `DeleteByXxx`, `GetByXxx`

## Module Documentation

Each module maintains its own `README.md` describing its purpose, design decisions, and usage.