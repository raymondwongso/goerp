# Code Reviewer Agent Memory

## Codebase Patterns (confirmed)

### Interface Location
- Module-level use case interfaces live in `domain/auth/auth.go` (package `auth`) — NOT in the module's own `interfaces.go`
- Repository interfaces live in the same domain file as their struct (e.g., `domain/user.go`)

### Usecase Interface Naming
- Interfaces named after their use case: `GoogleLogin`, `GoogleCallback` (no `-er` suffix for use cases)
- Method name is always `Invoke`

### Module Wiring File (`auth/auth.go`)
- `RegisterHTTPHandlers` is the standard wiring function signature for HTTP server. Naming depends on the application (e.g: `RegisterGRPCHandlers` for grpc)
- Accepts: `ctx`, `*http.ServeMux`, `*sqlx.DB`, `trace.Tracer`, `cfg` and whatever dependecies needed, return error

### HTTP Handler Pattern
- `writeError` is a package-level helper in each handler package
- Auth handler uses `xhttp.MapError(err)` + fallback to 500 (correct pattern)

### Cookie Security
- `session_id` cookie: `HttpOnly: true`, `Secure: true`, `SameSite: Lax`, `Path: "/"`, `MaxAge: 30 days`
- `Secure: true` must NOT be commented out — even in development, prefer an env-flag to toggle it
- Tests must assert `Secure: true` on cookie

### Environment Variable Handling
- `.env` is gitignored; `.env.example` contains no real secrets — safe
- `CORS_ALLOWED_ORIGINS=*` in `.env.example` is acceptable as a dev default, but must never be `*` when `Access-Control-Allow-Credentials: true` is set simultaneously

### CORS Security Rule (critical)
- `Access-Control-Allow-Credentials: true` + `Access-Control-Allow-Origin: *` is forbidden by the browser CORS spec and is a security misconfiguration
- Either reflect the specific `Origin` header value or omit credentials header when wildcard is used
- Middleware must validate the allowedOrigins value and handle the wildcard case specially

### IP Address Trust
- `req.RemoteAddr` is used directly for IP — no X-Forwarded-For / X-Real-IP processing
- This is safe when the server is not behind a reverse proxy; document the assumption
- If a load balancer is ever added, RemoteAddr will always be the proxy IP — needs revisiting

### Middleware Pattern (xhttp)
- `responseWriter` wrapper in `logging.go` only overrides `WriteHeader`; does not override `Write`
- This causes status to remain 200 when handler writes body without calling `WriteHeader` explicitly (Go's default), which is actually correct behavior since `http.ResponseWriter.Write` calls `WriteHeader(200)` implicitly — but the wrapper won't capture it
- Missing `Write` override means status stays at wrapper's default 200 even if the underlying recorder gets 200 via implicit WriteHeader — functionally correct but incomplete wrapper

### Store Layer Error Wrapping
- Store-layer raw DB errors (non-ErrNoRows) are returned unwrapped (no xerror wrapping) — this is intentional; the use case layer wraps them with xerror context

### Test Patterns
- Handler test suite pattern: `handlerTestSuite` struct + `newHandlerTestSuite(t)` + `newHandler()` method
- gomock.Any() used for non-deterministic request fields (state, PKCE, IP)
- `login_test.go` does NOT test the IPAddress field being set in the Insert call — uses gomock.Any() which masks a potential regression

## Recurring Anti-Patterns Found
1. `Secure: true` commented out in handler (`auth/http/handler.go` line 79) while test asserts `Secure: true` — handler and test are out of sync
2. `Access-Control-Allow-Credentials: true` combined with wildcard origin — critical CORS misconfiguration
3. `responseWriter` wrapper missing `Write()` override — status capture is incomplete for handlers that don't call WriteHeader
