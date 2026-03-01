---
description: Senior DevOps Engineer skill. Apply when writing Dockerfiles, Docker Compose files, Makefile targets, CI/CD workflows, or any infrastructure configuration. Covers lean container builds, local dev setup, and security/vulnerability workflows.
---

## Container Builds — Lean and Secure
- Use **multi-stage Docker builds**: one stage to compile, one minimal final stage (e.g., `gcr.io/distroless/static` or `alpine`).
- Strip debug symbols when building Go binaries: `CGO_ENABLED=0 go build -ldflags="-s -w" -o bin/api ./cmd/api`.
- Do not include source code, build tools, or package managers in the final image.
- Run the process as a non-root user in the final image.
- Pin base image versions to specific tags or digests for reproducibility.

## Local Development Setup
Makefile must cover local dev needs:
- `make run-api` — build and run the API server
- `make test` — run all tests with coverage (`go test ./... -coverpkg=./... -coverprofile=coverage.out`)
- `make build` — compile to `bin/api`
- `make compose-up` / `make compose-down` — start/stop Docker Compose services
- `make migrate-up` / `make migrate-down` — run goose migrations
- `make generate` — run `go generate ./...` for mock generation

Docker Compose (`docker-compose.yml` or `compose.yaml`) must provide:
- PostgreSQL with version matching the project (PostgreSQL 18)
- Health checks on dependent services
- Named volumes for persistent data
- Environment variables loaded from `.env` (never hardcoded in compose file)
- A `.env.example` with all required variable names but no real values

## Security and Vulnerability Workflows

### Secret / Credential Scanning
- Use `trufflesecurity/trufflehog` or `gitleaks` to scan for leaked secrets in commits and staged files.
- Run on every PR and push to main.

### Static Security Analysis (gosec)
```bash
go install github.com/securego/gosec/v2/cmd/gosec@latest
gosec -fmt=sarif -out=gosec.sarif ./...
```
- Run as a CI step; upload SARIF results to GitHub Code Scanning if available.
- Treat `HIGH` severity findings as blocking; `MEDIUM` as warnings.

### Dependency Vulnerability Check
```bash
go install golang.org/x/vuln/cmd/govulncheck@latest
govulncheck ./...
```
- Run on every CI build; fail the build on known vulnerabilities affecting called functions.

## CI/CD Pipeline Structure (GitHub Actions)
A well-structured workflow includes these jobs in order:
1. `lint` — `gofmt`, `go vet`, `golangci-lint`
2. `security` — `gosec`, `govulncheck`, secret scanning
3. `test` — `go test -race ./... -coverpkg=./...`
4. `build` — compile binary, verify it produces a valid artifact
5. `docker` — build and push image (on main branch only)

Each job should:
- Cache Go module downloads (`actions/cache` on `~/go/pkg/mod`)
- Use pinned action versions (e.g., `actions/checkout@v4`)
- Fail fast on errors
