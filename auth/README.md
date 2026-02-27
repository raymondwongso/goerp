# Authx - Authentication and Authorization Module

## Authentication

### Login() return (User)
Login (or register if not yet registered) using 3rd party OAuth provider. Return basic user info, session token and refresh token. Session and refresh token are returned in httpOnly cookie.

Supported OAuth provider:
1. Google

### Google/Callback()
Receives callback after login with Google OAuth flow.

### RefreshToken(refresh_token) return (User)
Refresh token when session token expired. Return basic user info, new session and refresh token. Session and refresh token are returned in httpOnly cookie.

## Authorization

### HasPermission(user_id, resource, action) return (bool)
Checks if `user_id` has permission to do `action` unto `resource`

### HasPermissions(user_id, resource, actions) return (map[string]bool)
Checks if `user_id` has permissions to do `actions` unto `resource`. Return map with `action` as key. Typically useful for bulk checking.
