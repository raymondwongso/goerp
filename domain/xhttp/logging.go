package xhttp

import (
	"log"
	"net/http"
	"time"

	"github.com/raymondwongso/goerp/domain/xsanitize"
)

// responseWriter wraps http.ResponseWriter to capture the status code written by the handler.
type responseWriter struct {
	http.ResponseWriter
	status int // 0 means WriteHeader has not been called yet
}

func (rw *responseWriter) WriteHeader(status int) {
	rw.status = status
	rw.ResponseWriter.WriteHeader(status)
}

// Write ensures WriteHeader is recorded through the wrapper before the first write,
// matching net/http's implicit WriteHeader(200) behaviour.
func (rw *responseWriter) Write(b []byte) (int, error) {
	if rw.status == 0 {
		rw.WriteHeader(http.StatusOK)
	}
	return rw.ResponseWriter.Write(b)
}

// LoggingMiddleware logs each request's method, path, status code, and elapsed time.
// Requests that result in a 4xx or 5xx are logged as errors.
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := &responseWriter{ResponseWriter: w}

		next.ServeHTTP(rw, r)

		status := rw.status
		if status == 0 {
			status = http.StatusOK
		}

		elapsed := time.Since(start)
		method := xsanitize.SanitizeEscapeCharacters(r.Method)
		path := xsanitize.SanitizeEscapeCharacters(r.URL.Path)
		if status >= http.StatusBadRequest {
			log.Printf("ERROR %s %s status=%d duration=%s", method, path, status, elapsed) // #nosec G706 -- method and path are sanitized by SanitizeEscapeCharacters
			return
		}
		log.Printf("%s %s status=%d duration=%s", method, path, status, elapsed) // #nosec G706 -- method and path are sanitized by SanitizeEscapeCharacters
	})
}
