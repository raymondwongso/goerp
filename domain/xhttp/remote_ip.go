package xhttp

import "net/http"

// RemoteIP returns the client IP address from the request.
// It checks X-Forwarded-For and X-Real-IP headers first (set by reverse proxies),
// and returns an empty string if neither header is present.
func RemoteIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return xff
	}
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}
	return ""
}
