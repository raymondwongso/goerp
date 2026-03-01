---
description: Senior Go Engineer skill. Apply when writing or reviewing Go code. Covers concurrency safety, code quality gates, testing discipline, naming conventions, SOLID principles, security, and error handling.
---

## Concurrency Safety
- Use `sync.WaitGroup` or `golang.org/x/sync/errgroup` for goroutine lifecycle management — never launch goroutines without a mechanism to wait for completion and collect errors.
- Protect shared mutable state with `sync.Mutex` or `sync.RWMutex`; prefer channels for communication between goroutines.
- Run `go test -race ./...` to detect race conditions before finalizing any implementation.

## Code Quality Gates
Run these before finalizing any Go file:
1. `gofmt -w <file>` — enforce canonical formatting
2. `go vet ./...` — catch common static errors
3. `go test -race ./...` — detect race conditions

## Testing Discipline
- **Table-driven tests**: use when the test logic is simple and multiple input/output pairs differ only in data.
- **Normal unit tests with mocks**: use when heavy mock setup is required and table-driven structure adds ceremony without clarity.
- Every use case must have tests covering: happy path, all store error branches, and input validation errors.
- HTTP handlers: use `httptest.NewRecorder` + `httptest.NewRequest`; cover success and all error responses.
- Store layer: use `go-sqlmock` with `sqlx.NewDb(db, "sqlmock")`.
- Use `gomock.Any()` for non-deterministic values (UUIDs generated at runtime, random state, PKCE verifiers).
- Mocks must be generated via mockgen and live in `mock/` — never hand-write mocks.

## Naming Conventions
- Package names: lowercase, short, no underscores, no stutter (not `auth.AuthService` → `auth.Service`).
- Store methods: `Get`, `GetByXxx`, `List`, `Create`, `Insert`, `Update`, `Upsert`, `DeleteByXxx`.
- Single-method interfaces: end in `-er` where idiomatic (e.g., `Storer`, `Writer`, `Reader`).
- Use case structs and methods: reflect business intent (e.g., `GoogleLogin`, `GoogleCallback`).
- Exported names: clear, idiomatic, consistent with the rest of the codebase.

## SOLID Compliance
- **Single Responsibility**: each struct and function has one clear responsibility.
- **Open/Closed**: extend behavior via interfaces, not by modifying concrete types.
- **Liskov Substitution**: interface implementations fully satisfy the contract.
- **Interface Segregation**: narrow, focused interfaces — no methods implementors don't need.
- **Dependency Inversion**: handlers and use cases depend on interfaces, never concrete implementations; inject all dependencies.
- Validate layering: `HTTP Handler → Use Case (interface) → Store (interface) → PostgreSQL`.

## Configuration via ENV
- All server/app configuration (timeouts, addresses, credentials) must come from `os.Getenv`, never hardcoded.
- Use a local `envDuration(key string, defaultVal time.Duration) time.Duration` helper to parse `time.Duration` from env with a sane default and a log warning on parse failure.
- ENV variable naming convention: `HTTP_READ_HEADER_TIMEOUT`, `HTTP_READ_TIMEOUT`, `HTTP_WRITE_TIMEOUT`, `HTTP_IDLE_TIMEOUT`, `API_ADDR`.

## Security
- SQL: parameterized queries only — no string-concatenated SQL.
- HTTP handlers: validate and sanitize all input; never expose internal error details to clients.
- Cookies: set `HttpOnly` and `Secure` flags on session cookies.
- Type assertions: always use the two-value form (`v, ok := x.(T)`).
- No hardcoded secrets, API keys, or connection strings in source code.
- `os.Getenv` values only through config structs — never log or return them in responses.
- Check that test files do not contain real credentials or tokens.
- UUIDs use UUIDv7 as per project convention.

## Error Handling
- `xerror.New(code, msg)` for typed domain errors.
- `xerror.NewWithCause(code, msg, err)` when wrapping a cause.
- `xerror.AddDetail(err, field, reason)` for field-level validation errors.
- HTTP handlers: use `xhttp.MapError(err)` and always fallback to 500 for unknown codes.
- Never use raw `errors.New` or `fmt.Errorf` for domain errors.
- Never swallow errors silently.
