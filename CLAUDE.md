# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Code Structure

The root directory is organized into **modules** and **special folders**.

**Special folders** (not modules):
- `bin/` — compiled binary output
- `cmd/` — application entrypoints (main packages)
- `docs/` — documentation
- `migration/` — database migration files
- `scripts/` — utility scripts (e.g., seed data)

Everything else at the root level is a **module** (e.g., `auth/`, `example/`). Each module represents a business domain and follows this structure:

```
<module>/
├── interfaces.go        # All interfaces: Reader, Writer, UseCaseX (source for go generate)
├── <module>.go          # Glue/registration code: wires handlers, use cases, and repositories
├── store/
│   └── postgres/        # PostgreSQL implementation of Reader/Writer interfaces
├── usecase/<submodule>/ # Business logic — one package and one struct per use case
├── http/                # HTTP handlers (other protocols get their own folder: grpc/, kafka/, etc.)
└── mock/                # MockGen-generated mocks (do not edit manually)
```

**Module conventions:**
- `<module>.go` is the registration entrypoint. Each module exposes a method (e.g., `RegisterHTTPHandlers(mux)`) called from `cmd/`, keeping `main.go` clean.
- `interfaces.go` defines all use case and repository interfaces. The `//go:generate` directive here produces all mocks into `mock/`.
- `usecase/<submodule>/` — one struct, one use case. No struct handles multiple use cases. A module may have multiple submodules, each with their own usecase package.
- `store/` is grouped by technology (e.g., `postgres/`, and potentially `kafka/`, `slack/`, `pubsub/` in the future).
- `domain/` is a special shared module containing structs and utilities used across all modules. It holds no inter-module business logic; any logic here must be truly domain-level (e.g., `xerror`). Domain may have subdomains.

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

HTTP status mapping from `xerror.Code` is handled by a dedicated error mapper (WIP — will live in `domain/errormapper`).

## Database

- **Driver:** PostgreSQL via [`jmoiron/sqlx`](https://github.com/jmoiron/sqlx)
- **Migrations:** [`pressly/goose`](https://github.com/pressly/goose) format in `migration/` — filenames prefixed with `00001_`, `00002_`, etc.

## Module Documentation

Each module maintains its own `README.md` describing its purpose, design decisions, and usage.
