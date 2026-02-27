# Authentication & Authorization System Design

> **Stack:** Google OAuth 2.0 · Cookie Sessions · RBAC · PostgreSQL · Go  
> **Constraints:** No username/password. Delegate all identity to Google. Cookie-based sessions. PostgreSQL as single source of truth.

---

## Table of Contents

1. [System Architecture](#1-system-architecture)
2. [Google Cloud Project Setup](#2-google-cloud-project-setup)
3. [Google OAuth 2.0 Flow](#3-google-oauth-20-flow)
4. [Database Schema](#4-database-schema)
5. [Cookie Design](#5-cookie-design)
6. [RBAC Design](#6-rbac-design)
7. [Permission Check Query & EXPLAIN Analysis](#7-permission-check-query--explain-analysis)
8. [Go Code: Middleware & Handlers](#8-go-code-middleware--handlers)
9. [Security Checklist](#9-security-checklist)

---

## 1. System Architecture

### 1.1 Component Overview

```
┌─────────────────────────────────────────────────────────────────────┐
│                            Browser (SPA/SSR)                        │
│              Stores HttpOnly session cookie only                     │
└───────────────────────────┬─────────────────────────────────────────┘
                            │ HTTPS
                            ▼
┌─────────────────────────────────────────────────────────────────────┐
│                  API Gateway / Reverse Proxy (nginx / Caddy)        │
│            TLS termination · Rate limiting · Cookie forwarding       │
└──────────────┬──────────────────────────────┬────────────────────────┘
               │                              │
               ▼                              ▼
┌──────────────────────────┐    ┌─────────────────────────────────────┐
│      Auth Service (Go)   │    │     Application Service (Go)        │
│                          │    │                                     │
│  · /auth/google/login    │    │  · Business logic handlers          │
│  · /auth/google/callback │    │  · Session validation middleware    │
│  · /auth/logout          │    │  · RBAC enforcement middleware      │
│  · /auth/me              │    │                                     │
└──────────┬───────────────┘    └──────────────┬──────────────────────┘
           │                                   │
           └──────────────┬────────────────────┘
                          │
                          ▼
          ┌───────────────────────────────┐
          │         PostgreSQL            │
          │                               │
          │  users · oauth_accounts       │
          │  oauth_states · sessions      │
          │  roles · permissions          │
          │  role_permissions · user_roles│
          └───────────────────────────────┘

                          ▲
                          │ Token exchange (server-to-server only)
                          ▼
          ┌───────────────────────────────┐
          │       Google OAuth 2.0        │
          │   accounts.google.com         │
          └───────────────────────────────┘
```

### 1.2 Component Responsibilities

| Component | Responsibility | Technology |
|-----------|---------------|------------|
| Browser | Stores HttpOnly session cookie, initiates OAuth redirect | Any SPA / SSR |
| API Gateway | TLS termination, rate limiting, does NOT strip cookies | nginx / Caddy |
| Auth Service | OAuth callback, session create/destroy, token rotation | Go |
| Application Service | Business logic, RBAC enforcement on every handler | Go |
| PostgreSQL | Source of truth: users, sessions, roles, permissions | PostgreSQL 15+ |
| Google OAuth 2.0 | Delegates identity verification entirely | Google Cloud |

### 1.3 Design Principles

- **Zero passwords stored.** Identity is fully delegated to Google.
- **Access tokens never reach the browser.** Google's access token is exchanged server-side and immediately discarded (we only need the `id_token` claims).
- **Stateful sessions.** The cookie holds an opaque UUID. All session state lives in PostgreSQL. Enables instant revocation.
- **RBAC is enforced in the Application Service**, not the Auth Service. Auth proves identity; the app enforces access.

---

## 2. Google Cloud Project Setup

This section walks through creating a Google Cloud project and OAuth 2.0 credentials from scratch.

### 2.1 Create a Google Cloud Project

1. Go to [https://console.cloud.google.com](https://console.cloud.google.com)
2. Click the project dropdown at the top → **"New Project"**
3. Fill in:
   - **Project name:** `my-app-prod`
   - **Organization:** Select your org if applicable
4. Click **"Create"** and wait ~30 seconds for provisioning
5. Select the new project from the dropdown

### 2.2 Enable the Required APIs

Navigate to **APIs & Services → Library** and enable:

- Search for **"Google Identity"** → Enable **"Identity and Access Management (IAM) API"** — not strictly required but good practice

> The OAuth 2.0 endpoint (`accounts.google.com`) does not require explicit API enablement. Only APIs like Google Drive, Calendar, etc. need to be enabled separately.

### 2.3 Configure the OAuth Consent Screen

**APIs & Services → OAuth consent screen**

1. **User Type:** Choose **"External"** (for public apps) or **"Internal"** (for Google Workspace org users only)
2. Click **"Create"**
3. Fill in App Information:
   - **App name:** `My App`
   - **User support email:** your email
   - **Application home page:** `https://yourdomain.com`
   - **Privacy policy URL:** `https://yourdomain.com/privacy`
   - **Authorized domains:** `yourdomain.com`
4. Click **"Save and Continue"**

**Scopes (Step 2):** Click **"Add or Remove Scopes"** and add:

| Scope | Purpose |
|-------|---------|
| `openid` | Required for OIDC / id_token |
| `email` | Get user's email address |
| `profile` | Get display name and avatar URL |

> **Do not request more scopes than you need.** `openid`, `email`, `profile` are non-sensitive and require no Google verification review.

### 2.4 Create OAuth 2.0 Client ID

**APIs & Services → Credentials → "Create Credentials" → "OAuth client ID"**

1. **Application type:** `Web application`
2. **Name:** `my-app-web`
3. **Authorized JavaScript origins:**
   - `https://yourdomain.com`
   - `http://localhost:3000` (local dev)
4. **Authorized redirect URIs:** ← Critical — must match exactly
   - `https://yourdomain.com/auth/google/callback`
   - `http://localhost:8080/auth/google/callback` (local dev)

Click **"Create"**. You will receive:

- **Client ID** — format: `123456789-abc123def456.apps.googleusercontent.com`
- **Client Secret** — treat like a password, never commit to source control

### 2.5 Store Credentials Securely

```bash
# .env — never commit this file, add to .gitignore
GOOGLE_CLIENT_ID=123456789-abc123def456.apps.googleusercontent.com
GOOGLE_CLIENT_SECRET=GOCSPX-xxxxxxxxxxxxxxxxxxxxxxxx
GOOGLE_REDIRECT_URI=https://yourdomain.com/auth/google/callback
```

In Go:

```go
oauthConfig := &oauth2.Config{
    ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
    ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
    RedirectURL:  os.Getenv("GOOGLE_REDIRECT_URI"),
    Scopes:       []string{"openid", "email", "profile"},
    Endpoint:     google.Endpoint,
}
```

### 2.6 Local Dev vs Production

| Setting | Local Dev | Production |
|---------|-----------|-----------|
| Redirect URI | `http://localhost:8080/...` | `https://yourdomain.com/...` |
| Consent Screen | Testing mode (100 test users max) | Published |
| OAuth Client | Separate dev credentials | Separate prod credentials |
| Cookie `Secure` flag | `false` (HTTP) | `true` (HTTPS only) |

> **Best practice:** Create two separate OAuth Client IDs — one for dev, one for prod. Never share credentials between environments.

### 2.7 Publishing the Consent Screen

While in "Testing" mode, only explicitly added test users can log in. To open to all Google accounts:

**APIs & Services → OAuth consent screen → "Publish App" → Confirm**

For apps using only `openid`, `email`, `profile`, Google does **not** require a formal verification review. The app goes live immediately.

---

## 3. Google OAuth 2.0 Flow

### 3.1 Full Login Sequence

```
Browser          Auth Service          PostgreSQL          Google
   │                   │                    │                  │
   │  GET /auth/login  │                    │                  │
   │──────────────────►│                    │                  │
   │                   │ Generate:          │                  │
   │                   │  state (32B random)│                  │
   │                   │  code_verifier     │                  │
   │                   │  code_challenge    │                  │
   │                   │                    │                  │
   │                   │ INSERT oauth_states│                  │
   │                   │───────────────────►│                  │
   │                   │                    │                  │
   │  302 → accounts.google.com/o/oauth2/v2/auth              │
   │  ?client_id=...&state=...&code_challenge=...             │
   │◄──────────────────│                    │                  │
   │                   │                    │                  │
   │  User authenticates at Google                            │
   │──────────────────────────────────────────────────────────►│
   │                   │                    │                  │
   │  302 → /auth/google/callback?code=...&state=...          │
   │◄──────────────────────────────────────────────────────────│
   │                   │                    │                  │
   │  GET /callback    │                    │                  │
   │──────────────────►│                    │                  │
   │                   │ DELETE oauth_states│                  │
   │                   │ WHERE state=$1     │                  │
   │                   │ RETURNING code_verifier               │
   │                   │───────────────────►│                  │
   │                   │                    │                  │
   │                   │ POST /token (code + code_verifier)    │
   │                   │──────────────────────────────────────►│
   │                   │ ◄── id_token + access_token ─────────│
   │                   │                    │                  │
   │                   │ Verify id_token sig/claims            │
   │                   │                    │                  │
   │                   │ UPSERT users + oauth_accounts         │
   │                   │───────────────────►│                  │
   │                   │                    │                  │
   │                   │ INSERT sessions    │                  │
   │                   │───────────────────►│                  │
   │                   │                    │                  │
   │  302 / + Set-Cookie: session_id=...    │                  │
   │◄──────────────────│                    │                  │
```

### 3.2 Step-by-Step Breakdown

| Step | Actor | Action | Security Note |
|------|-------|--------|---------------|
| 1 | Auth Service | Generate `state` (32 random bytes) and PKCE `code_verifier`/`code_challenge` | State prevents CSRF. PKCE prevents code interception. |
| 2 | Auth Service | Store state + code_verifier in `oauth_states` with 5-min expiry | DB-backed state is safer than cookie-based |
| 3 | Browser | Redirected to Google with `state`, `code_challenge`, `code_challenge_method=S256` | |
| 4 | Google | User authenticates, redirected back to callback URI | |
| 5 | Auth Service | `DELETE FROM oauth_states WHERE state=$1 AND expires_at > now() RETURNING code_verifier` | Single-use. Atomic consume. If not found → reject. |
| 6 | Auth Service | Exchange `code` + `code_verifier` for tokens via server-to-server POST | Never expose to browser |
| 7 | Auth Service | Verify `id_token`: signature (JWKS), `iss`, `aud`, `exp` | Use `coreos/go-oidc` — don't roll your own |
| 8 | Auth Service | UPSERT user record, log `oauth_accounts` entry | `provider_sub` is the stable identity, not email |
| 9 | Auth Service | Create session row, set HttpOnly cookie | |
| 10 | Browser | Receives cookie, redirected to app | Access token is discarded, never touches browser |

---

## 4. Database Schema

### 4.1 users

The canonical identity record. Never stores passwords or tokens.

```sql
CREATE TABLE users (
  id            UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  email         TEXT        NOT NULL UNIQUE,
  display_name  TEXT,
  avatar_url    TEXT,
  is_active     BOOLEAN     NOT NULL DEFAULT true,
  created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_users_email ON users(email);
```

> **Why is `provider_sub` not in `users`?** An email can change at the provider side. The stable identity is the provider's `sub` (Google's internal user ID, never changes). Keeping it in `oauth_accounts` separates concerns and supports multiple providers later.

### 4.2 oauth_accounts

One user can link multiple OAuth providers. Stores the provider subject identifier — not tokens.

```sql
CREATE TABLE oauth_accounts (
  id           UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id      UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  provider     TEXT        NOT NULL,   -- 'google', 'github', etc.
  provider_sub TEXT        NOT NULL,   -- Google's stable subject ID (never changes)
  email        TEXT        NOT NULL,   -- provider email at time of login (can change)
  last_login   TIMESTAMPTZ NOT NULL DEFAULT now(),
  created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE(provider, provider_sub)       -- one record per provider account
);

CREATE INDEX idx_oauth_accounts_user ON oauth_accounts(user_id);
```

### 4.3 oauth_states

Short-lived CSRF state tokens. Single-use — deleted on consumption.

```sql
CREATE TABLE oauth_states (
  state         TEXT        PRIMARY KEY,  -- 32-byte random hex
  code_verifier TEXT        NOT NULL,     -- PKCE verifier (plain; hashed to challenge)
  redirect_to   TEXT,                     -- post-login destination URL
  created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
  expires_at    TIMESTAMPTZ NOT NULL DEFAULT now() + INTERVAL '5 minutes'
);

-- Cleanup (run via pg_cron or a background goroutine)
-- DELETE FROM oauth_states WHERE expires_at < now();
```

### 4.4 sessions

Stateful server-side sessions. The cookie holds only the `id`. All session data lives here.

```sql
CREATE TABLE sessions (
  id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id         UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  ip_address      INET,
  user_agent      TEXT,
  is_revoked      BOOLEAN     NOT NULL DEFAULT false,
  absolute_expiry TIMESTAMPTZ NOT NULL DEFAULT now() + INTERVAL '30 days',
  created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  last_seen_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_sessions_user_id ON sessions(user_id);
-- Partial index: only active sessions (used by the per-request validation query)
CREATE INDEX idx_sessions_active ON sessions(id, absolute_expiry)
  WHERE NOT is_revoked;
```

### 4.5 roles

```sql
CREATE TABLE roles (
  id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  name        TEXT        NOT NULL UNIQUE,  -- 'admin', 'editor', 'viewer'
  description TEXT,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
```

### 4.6 permissions

A permission is a `(resource, action)` pair.

```sql
CREATE TABLE permissions (
  id       UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  resource TEXT NOT NULL,  -- 'article', 'user', 'report', 'billing'
  action   TEXT NOT NULL,  -- 'create', 'read', 'update', 'delete', 'publish'
  UNIQUE(resource, action)
);

-- Seed data example
INSERT INTO permissions (resource, action) VALUES
  ('article', 'create'),
  ('article', 'read'),
  ('article', 'update'),
  ('article', 'delete'),
  ('article', 'publish'),
  ('user',    'read'),
  ('user',    'update'),
  ('user',    'delete');
```

### 4.7 role_permissions

```sql
CREATE TABLE role_permissions (
  role_id       UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
  permission_id UUID NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
  PRIMARY KEY (role_id, permission_id)
);
```

### 4.8 user_roles

Assigns roles to users. Optionally scoped to a specific resource instance.

```sql
CREATE TABLE user_roles (
  user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  role_id    UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
  -- NULL = global (role applies everywhere)
--   scope_type TEXT,   -- e.g. 'organization', 'workspace'
--   scope_id   UUID,   -- FK to the specific resource instance
  granted_by UUID    REFERENCES users(id),
  granted_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (
    user_id,
    role_id
    -- COALESCE(scope_type, ''),
    -- COALESCE(scope_id, '00000000-0000-0000-0000-000000000000')
  )
);

CREATE INDEX idx_user_roles_user ON user_roles(user_id);
```

### 4.9 Entity Relationship Summary

```
users ──< oauth_accounts       (one user, many provider accounts)
users ──< sessions             (one user, many active sessions)
users ──< user_roles >── roles (many-to-many with optional scope)
roles ──< role_permissions >── permissions (many-to-many)
```

---

## 5. Cookie Design

### 5.1 Cookie Attributes

```
Set-Cookie: session_id=<uuid-v4>;
            HttpOnly;
            Secure;
            SameSite=Lax;
            Path=/;
            Max-Age=2592000;
            Domain=yourdomain.com
```

| Attribute | Value | Reason |
|-----------|-------|--------|
| `HttpOnly` | `true` | JavaScript cannot read the cookie. Prevents XSS-based session theft. |
| `Secure` | `true` | Only transmitted over HTTPS. |
| `SameSite` | `Lax` | Sent on top-level navigations (the OAuth redirect back from Google). Blocked on cross-site sub-requests. `Strict` would break the OAuth redirect. |
| `Path` | `/` | Available to all routes. |
| `Max-Age` | `2592000` (30d) | Matches `absolute_expiry` in DB. Persistent across browser restarts. |
| `Domain` | `yourdomain.com` | Explicit. Do NOT set if subdomains should not share the cookie. |

### 5.2 Why Stateful Sessions Instead of JWTs

| Concern | JWT | Stateful Session |
|---------|-----|-----------------|
| Instant revocation | Requires a blocklist (you've rebuilt statefulness) | `is_revoked = true` — immediate |
| Payload exposure | Claims are readable if not encrypted (JWE) | Cookie is an opaque UUID |
| Key management | Signing key rotation is a separate operational burden | No keys to manage |
| Audit trail | Not built-in | `sessions` table = full login history with IP + UA |
| Performance | No DB lookup (but blocklist negates this) | One indexed lookup per request |

### 5.3 Session Lifecycle

```
Login         → INSERT session row          → Set cookie
Request       → SELECT + validate expiry    → Inject user into context
Logout        → UPDATE is_revoked = true    → Set-Cookie: Max-Age=-1
Role change   → Revoke all user sessions    → Force re-authentication
Expiry        → absolute_expiry < now()     → Reject, clear cookie
```

### 5.4 Optional Sliding Expiry (Idle Timeout)

The schema uses absolute expiry (hard stop at 30 days). To also enforce an idle timeout:

```go
// Touch last_seen_at on each validated request
db.Exec(ctx, `UPDATE sessions SET last_seen_at = now() WHERE id = $1`, sessionID)

// In middleware: reject if idle > 24h
if time.Since(lastSeenAt) > 24*time.Hour {
    revokeSession(ctx, db, sessionID)
    http.Error(w, "Session expired due to inactivity", 401)
    return
}
```

---

## 6. RBAC Design

### 6.1 Model

```
User
 └──< UserRole (with optional scope) >── Role
                                           └──< RolePermission >── Permission(resource, action)
```

A permission check answers: **"Does user X have permission to perform action Y on resource Z?"**

### 6.2 Example Role Setup

```sql
INSERT INTO roles (name) VALUES ('admin'), ('editor'), ('viewer');

-- admin gets all article permissions
INSERT INTO role_permissions (role_id, permission_id)
  SELECT r.id, p.id FROM roles r, permissions p
  WHERE r.name = 'admin';

-- editor can create, read, update, publish
INSERT INTO role_permissions (role_id, permission_id)
  SELECT r.id, p.id FROM roles r, permissions p
  WHERE r.name = 'editor' AND p.action IN ('create', 'read', 'update', 'publish');

-- viewer can only read
INSERT INTO role_permissions (role_id, permission_id)
  SELECT r.id, p.id FROM roles r, permissions p
  WHERE r.name = 'viewer' AND p.action = 'read';
```

---

## 7. Permission Check Query & EXPLAIN Analysis

### 7.1 The Core Query

```sql
-- $1 = user_id    $2 = resource    $3 = action
-- $4 = scope_type (NULL for global)    $5 = scope_id (NULL for global)
SELECT EXISTS (
  SELECT 1
  FROM   user_roles ur
  JOIN   role_permissions rp ON rp.role_id = ur.role_id
  JOIN   permissions p       ON p.id = rp.permission_id
  WHERE  ur.user_id  = $1
  AND    p.resource  = $2
  AND    p.action    = $3
--   AND   (ur.scope_type IS NULL
--          OR (ur.scope_type = $4 AND ur.scope_id = $5))
) AS has_permission;
```

### 7.2 Running EXPLAIN ANALYZE

```sql
EXPLAIN (ANALYZE, BUFFERS, FORMAT TEXT)
SELECT EXISTS (
  SELECT 1
  FROM   user_roles ur
  JOIN   role_permissions rp ON rp.role_id = ur.role_id
  JOIN   permissions p       ON p.id = rp.permission_id
  WHERE  ur.user_id  = '018e1234-0000-7000-0000-000000000001'
  AND    p.resource  = 'article'
  AND    p.action    = 'delete'
  AND   (ur.scope_type IS NULL OR ur.scope_type = 'organization')
);
```

### 7.3 Expected Query Plan (With Indexes)

```
Result  (cost=8.31..8.32 rows=1 width=1) (actual time=0.04..0.04 rows=1)
  InitPlan 1 (returns $0)
    ->  Nested Loop  (cost=4.57..8.31 rows=1 width=0)
          ->  Nested Loop  (cost=4.14..7.87 rows=1 width=16)
                ->  Index Scan using idx_user_roles_user on user_roles ur
                      Index Cond: (user_id = '018e1234-...')
                      Filter: (scope_type IS NULL OR scope_type = 'organization')
                      Rows Removed by Filter: 0
                ->  Index Scan using role_permissions_pkey on role_permissions rp
                      Index Cond: (role_id = ur.role_id)
          ->  Index Scan using idx_permissions_resource_action on permissions p
                Index Cond: ((resource = 'article') AND (action = 'delete'))
Planning Time: 0.3 ms
Execution Time: 0.05 ms
```

### 7.4 Node-by-Node Analysis

| Plan Node | What PostgreSQL Does | What to Watch For |
|-----------|---------------------|-------------------|
| `Index Scan on user_roles` | Locates all roles for this `user_id` via `idx_user_roles_user` | Should return 1–5 rows for most users. If you see a Seq Scan here, the index is missing or not being used. |
| `Filter on scope_type` | Evaluates the `IS NULL OR scope_type = $4` predicate on each matched row | Low cost when the user has few roles. For scoped-heavy setups, add the composite index below. |
| `Nested Loop → role_permissions` | For each matched role, look up allowed permissions via the PK `(role_id, permission_id)` | PK is always indexed. Number of rows = total permissions across all user roles. |
| `Index Scan on permissions` | Resolves the permission row and filters by `resource` + `action` | **Must** use `idx_permissions_resource_action`. Without it, this becomes a Seq Scan on the full permissions table. |
| `EXISTS` short-circuit | Stops scanning after the **first matching row** | This is the most important optimization in the query. `EXISTS` never over-reads. Never rewrite this as `COUNT(*) > 0`. |

### 7.5 Required Indexes

```sql
-- 1. Primary access pattern: look up roles by user
CREATE INDEX idx_user_roles_user ON user_roles(user_id);

-- 2. Filter permissions by (resource, action) — covered by UNIQUE constraint
CREATE UNIQUE INDEX idx_permissions_resource_action ON permissions(resource, action);

-- 3. role_permissions PK already covers (role_id, permission_id) — no extra index needed

-- 4. Optional: composite index for scoped role lookups
CREATE INDEX idx_user_roles_user_scope ON user_roles(user_id, scope_type, scope_id)
  WHERE scope_type IS NOT NULL;
```

### 7.6 What Breaks the Plan

| Mistake | Effect |
|---------|--------|
| Missing `idx_user_roles_user` | Seq Scan on `user_roles` — catastrophic at scale |
| Missing `idx_permissions_resource_action` | Seq Scan on `permissions` on every request |
| Using `COUNT(*) > 0` instead of `EXISTS` | Scans all matching rows before returning |
| No `COALESCE` on PK for `user_roles` | NULL in PK columns causes duplicate role assignments |
| Calling this query per-resource in a list endpoint | N+1 — batch or cache the full permission set per user |

### 7.7 Performance Expectations

| Scenario | Rows in user_roles | Expected Latency |
|----------|--------------------|-----------------|
| User has 1–3 global roles | 1–3 | < 1ms |
| User has roles across 10 orgs | 10–30 | < 2ms |
| Admin with 50 permissions per role | Stops at 1st match (EXISTS) | < 1ms |
| **No indexes** | Full table scans | 50–500ms — unacceptable |

### 7.8 Caching Strategy

For APIs where every request triggers a permission check, cache the result:

```
Request
  └── in-memory cache hit? (key: user_id:resource:action, TTL: 60s)
        ├── HIT  → return cached bool, no DB
        └── MISS → query PostgreSQL → store in cache → return
```

**Invalidate cache when:**
- A `user_roles` row is inserted or deleted for this user
- `role_permissions` changes for any role this user holds

Simple Go implementation using `sync.Map` with expiry, or Redis keyed as `perm:{user_id}:{resource}:{action}`.

---

## 8. Go Code: Middleware & Handlers

### 8.1 Session Validation Middleware

```go
type contextKey string
const ctxKeyUserID contextKey = "userID"

func SessionMiddleware(db *pgxpool.Pool) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            cookie, err := r.Cookie("session_id")
            if err != nil {
                http.Error(w, "Unauthorized", http.StatusUnauthorized)
                return
            }

            var userID string
            var expiry time.Time
            var revoked bool

            err = db.QueryRow(r.Context(), `
                SELECT user_id, absolute_expiry, is_revoked
                FROM   sessions
                WHERE  id = $1
            `, cookie.Value).Scan(&userID, &expiry, &revoked)

            if err != nil || revoked || expiry.Before(time.Now()) {
                http.SetCookie(w, &http.Cookie{
                    Name: "session_id", Value: "", MaxAge: -1,
                })
                http.Error(w, "Unauthorized", http.StatusUnauthorized)
                return
            }

            ctx := context.WithValue(r.Context(), ctxKeyUserID, userID)
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}
```

### 8.2 RBAC Middleware

```go
const permQuery = `
    SELECT EXISTS (
        SELECT 1
        FROM   user_roles ur
        JOIN   role_permissions rp ON rp.role_id = ur.role_id
        JOIN   permissions p       ON p.id = rp.permission_id
        WHERE  ur.user_id = $1
        AND    p.resource = $2
        AND    p.action   = $3
        AND   (ur.scope_type IS NULL
               OR (ur.scope_type = $4 AND ur.scope_id = $5))
    )
`

func RequirePermission(db *pgxpool.Pool, resource, action string) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            userID, ok := r.Context().Value(ctxKeyUserID).(string)
            if !ok {
                http.Error(w, "Unauthorized", http.StatusUnauthorized)
                return
            }

            var allowed bool
            db.QueryRow(r.Context(), permQuery,
                userID, resource, action, nil, nil,
            ).Scan(&allowed)

            if !allowed {
                http.Error(w, "Forbidden", http.StatusForbidden)
                return
            }
            next.ServeHTTP(w, r)
        })
    }
}

// Router usage:
// mux.Handle("DELETE /articles/{id}",
//     SessionMiddleware(db)(
//         RequirePermission(db, "article", "delete")(deleteArticleHandler),
//     ),
// )
```

### 8.3 OAuth Login Handler (with PKCE)

```go
func generateState() (string, error) {
    b := make([]byte, 32)
    _, err := rand.Read(b)
    return hex.EncodeToString(b), err
}

func generatePKCE() (verifier, challenge string, err error) {
    b := make([]byte, 64)
    if _, err = rand.Read(b); err != nil {
        return
    }
    verifier = base64.RawURLEncoding.EncodeToString(b)
    sum := sha256.Sum256([]byte(verifier))
    challenge = base64.RawURLEncoding.EncodeToString(sum[:])
    return
}

func (h *AuthHandler) GoogleLogin(w http.ResponseWriter, r *http.Request) {
    state, err := generateState()
    if err != nil { http.Error(w, "Internal error", 500); return }

    verifier, challenge, err := generatePKCE()
    if err != nil { http.Error(w, "Internal error", 500); return }

    _, err = h.db.Exec(r.Context(), `
        INSERT INTO oauth_states (state, code_verifier, redirect_to)
        VALUES ($1, $2, $3)
    `, state, verifier, r.URL.Query().Get("redirect_to"))
    if err != nil { http.Error(w, "Internal error", 500); return }

    url := h.oauthConfig.AuthCodeURL(state,
        oauth2.SetAuthURLParam("code_challenge", challenge),
        oauth2.SetAuthURLParam("code_challenge_method", "S256"),
    )
    http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}
```

### 8.4 OAuth Callback Handler

```go
func (h *AuthHandler) GoogleCallback(w http.ResponseWriter, r *http.Request) {
    code  := r.URL.Query().Get("code")
    state := r.URL.Query().Get("state")

    // 1. Atomically consume the state token (single-use)
    var codeVerifier, redirectTo string
    err := h.db.QueryRow(r.Context(), `
        DELETE FROM oauth_states
        WHERE  state = $1 AND expires_at > now()
        RETURNING code_verifier, COALESCE(redirect_to, '/')
    `, state).Scan(&codeVerifier, &redirectTo)
    if err != nil {
        http.Error(w, "Invalid or expired state", http.StatusBadRequest)
        return
    }

    // 2. Exchange code → tokens (server-to-server, never exposed to browser)
    token, err := h.oauthConfig.Exchange(r.Context(), code,
        oauth2.SetAuthURLParam("code_verifier", codeVerifier),
    )
    if err != nil { http.Error(w, "Token exchange failed", 500); return }

    // 3. Verify id_token
    rawIDToken, ok := token.Extra("id_token").(string)
    if !ok { http.Error(w, "Missing id_token", 400); return }

    idToken, err := h.oidcVerifier.Verify(r.Context(), rawIDToken)
    if err != nil { http.Error(w, "Token verification failed", 400); return }

    var claims struct {
        Sub, Email, Name, Picture string
    }
    idToken.Claims(&claims)

    // 4. Upsert user + oauth_account in a transaction
    userID, err := h.upsertUser(r.Context(), claims)
    if err != nil { http.Error(w, "Failed to create user", 500); return }

    // 5. Create session
    sessionID := uuid.New().String()
    h.db.Exec(r.Context(), `
        INSERT INTO sessions (id, user_id, ip_address, user_agent)
        VALUES ($1, $2, $3, $4)
    `, sessionID, userID, r.RemoteAddr, r.UserAgent())

    // 6. Set HttpOnly cookie — access_token is discarded here, never sent to browser
    http.SetCookie(w, &http.Cookie{
        Name:     "session_id",
        Value:    sessionID,
        HttpOnly: true,
        Secure:   true,
        SameSite: http.SameSiteLaxMode,
        Path:     "/",
        MaxAge:   86400 * 30,
    })

    http.Redirect(w, r, redirectTo, http.StatusFound)
}

func (h *AuthHandler) upsertUser(ctx context.Context, claims struct{ Sub, Email, Name, Picture string }) (string, error) {
    var userID string
    err := pgx.BeginFunc(ctx, h.db, func(tx pgx.Tx) error {
        err := tx.QueryRow(ctx, `
            INSERT INTO users (email, display_name, avatar_url)
            VALUES ($1, $2, $3)
            ON CONFLICT (email) DO UPDATE
              SET display_name = EXCLUDED.display_name,
                  avatar_url   = EXCLUDED.avatar_url,
                  updated_at   = now()
            RETURNING id
        `, claims.Email, claims.Name, claims.Picture).Scan(&userID)
        if err != nil { return err }

        _, err = tx.Exec(ctx, `
            INSERT INTO oauth_accounts (user_id, provider, provider_sub, email)
            VALUES ($1, 'google', $2, $3)
            ON CONFLICT (provider, provider_sub) DO UPDATE
              SET email      = EXCLUDED.email,
                  last_login = now()
        `, userID, claims.Sub, claims.Email)
        return err
    })
    return userID, err
}
```

### 8.5 Logout Handler

```go
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
    cookie, err := r.Cookie("session_id")
    if err == nil {
        // Server-side revocation — this is the authoritative invalidation
        h.db.Exec(r.Context(),
            `UPDATE sessions SET is_revoked = true WHERE id = $1`,
            cookie.Value,
        )
    }

    // Clear cookie client-side
    http.SetCookie(w, &http.Cookie{
        Name: "session_id", Value: "",
        HttpOnly: true, Secure: true,
        SameSite: http.SameSiteLaxMode,
        Path: "/", MaxAge: -1,
    })
    http.Redirect(w, r, "/", http.StatusFound)
}
```

---

## 9. Security Checklist

| Area | Requirement | Notes |
|------|-------------|-------|
| **CSRF** | `SameSite=Lax` + DB-backed `state` parameter | Single-use, 5-min TTL |
| **XSS** | `HttpOnly` cookie; access token never reaches browser | JS cannot steal session |
| **PKCE** | `code_verifier`/`code_challenge` on every login | Prevents authorization code interception |
| **State replay** | `oauth_states` deleted on consumption | Atomic `DELETE ... RETURNING` |
| **Session fixation** | New `session_id` on every login | Never reuse IDs |
| **Logout** | Server-side `is_revoked = true` + clear cookie | Server-side is authoritative |
| **Token storage** | Google `access_token` discarded; only `sub` + `email` stored | Reduces breach surface area |
| **TLS** | `Secure` cookie flag; HSTS at gateway | Enforce at nginx/Caddy level |
| **Rate limiting** | Limit `/auth/google/login` | Prevents `oauth_states` table flooding |
| **Audit log** | `sessions` table = full login history (IP, UA, timestamps) | Query by `user_id` |
| **Cleanup jobs** | Delete expired `oauth_states` and old `sessions` | `pg_cron` or background goroutine |
| **Privilege change** | Revoke all active sessions on role grant/revoke | Forces re-auth with new privileges |
| **`id_token` verification** | Check signature, `iss`, `aud`, `exp` | Use `coreos/go-oidc` — never roll your own |

### 9.1 Required Go Dependencies

```go
// go.mod
require (
    golang.org/x/oauth2         latest  // OAuth2 client
    github.com/coreos/go-oidc/v3 latest  // OIDC id_token verification
    github.com/jackc/pgx/v5     latest  // PostgreSQL driver
    github.com/google/uuid      latest  // UUID v4 generation
)
```

---

*All schema, cookie, and code patterns reflect production-grade practices. Adapt session TTLs, scope model, and caching strategy to your specific requirements.*
