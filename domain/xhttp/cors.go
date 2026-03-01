package xhttp

import (
	"net/http"
	"strings"
)

// CORSMiddleware adds CORS headers to every response.
//
// allowedOrigins is either "*" (wildcard) or a comma-separated list of allowed
// origins (e.g. "https://app.example.com,https://admin.example.com").
//
// Wildcard mode ("*"): sets Access-Control-Allow-Origin: * but does NOT set
// Access-Control-Allow-Credentials — browsers forbid the combination.
// For authenticated flows that use session cookies (credentials: "include"),
// set CORS_ALLOWED_ORIGINS to the exact frontend origin instead.
//
// Specific-origin mode: reflects the request Origin header when it matches the
// allowlist, sets Access-Control-Allow-Credentials: true, and adds Vary: Origin
// so downstream caches store separate responses per origin.
func CORSMiddleware(allowedOrigins string, next http.Handler) http.Handler {
	wildcard := allowedOrigins == "*"
	origins := parseOrigins(allowedOrigins)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if wildcard {
			w.Header().Set("Access-Control-Allow-Origin", "*")
		} else {
			requestOrigin := r.Header.Get("Origin")
			if originAllowed(requestOrigin, origins) {
				w.Header().Set("Access-Control-Allow-Origin", requestOrigin)
				w.Header().Set("Access-Control-Allow-Credentials", "true")
				w.Header().Add("Vary", "Origin")
			}
		}

		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func parseOrigins(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if t := strings.TrimSpace(p); t != "" {
			out = append(out, t)
		}
	}
	return out
}

func originAllowed(origin string, allowed []string) bool {
	for _, a := range allowed {
		if a == origin {
			return true
		}
	}
	return false
}
