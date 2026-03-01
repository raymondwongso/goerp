# domain — Domain Objects

`domain` contains shared objects used across modules: structs, interfaces, and utilities.

`domain` must not import any non-domain modules.

## Structure

### Root files (`domain/*.go`)

Root-level files define **shared objects** that are not specific to any one module — things like core business entities and their repository interfaces:

- `user.go` — `User` struct + `UserWriter` interface
- `oauth.go` — `OAuthState`, `OAuthAccount` structs + writer interfaces
- `session.go` — `Session` struct + `SessionWriter` interface
- `role.go`, `permission.go` — RBAC structs + interfaces
- `example.go` — Example struct (reference implementation)

### Subdomain packages (`domain/<subdomain>/`)

Subdomain packages hold objects that are **specific to a particular external domain or vendor**. They follow the same rule of not importing non-domain modules:

- `domain/auth/` — Request/result DTOs and use case interfaces for the authentication flow (`GoogleLoginRequest`, `GoogleLogin`, etc.)
- `domain/google/` — Google-specific types and the `TokenProvider` interface (OAuth2 + OIDC abstraction)

## Utilities

### xerror

`xerror` provides typed error handling with error codes (`CodeNotFound`, `CodeUnauthorized`, etc.) and helpers for wrapping, inspecting, and adding field-level detail to errors.

### xhttp

`xhttp` maps `xerror.Code` values to HTTP status codes via `MapError(err)`.
