package xhttp

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCORSMiddleware(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// --- wildcard mode ---

	t.Run("wildcard: sets ACAO=* on all methods", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Origin", "https://any.example.com")
		w := httptest.NewRecorder()

		CORSMiddleware("*", next).ServeHTTP(w, req)

		assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
		assert.Equal(t, "GET, POST, PUT, PATCH, DELETE, OPTIONS", w.Header().Get("Access-Control-Allow-Methods"))
		assert.Equal(t, "Content-Type, Authorization", w.Header().Get("Access-Control-Allow-Headers"))
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("wildcard: does not set Allow-Credentials (browser forbids * + credentials)", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Origin", "https://attacker.com")
		w := httptest.NewRecorder()

		CORSMiddleware("*", next).ServeHTTP(w, req)

		assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
		assert.Empty(t, w.Header().Get("Access-Control-Allow-Credentials"))
	})

	t.Run("wildcard: OPTIONS preflight returns 204 and does not call next", func(t *testing.T) {
		t.Parallel()

		called := false
		req := httptest.NewRequest(http.MethodOptions, "/", nil)
		w := httptest.NewRecorder()

		CORSMiddleware("*", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			called = true
		})).ServeHTTP(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code)
		assert.False(t, called)
	})

	// --- specific-origin mode ---

	t.Run("specific origin: matching request origin is reflected with credentials and Vary", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Origin", "https://app.example.com")
		w := httptest.NewRecorder()

		CORSMiddleware("https://app.example.com", next).ServeHTTP(w, req)

		assert.Equal(t, "https://app.example.com", w.Header().Get("Access-Control-Allow-Origin"))
		assert.Equal(t, "true", w.Header().Get("Access-Control-Allow-Credentials"))
		assert.Equal(t, "Origin", w.Header().Get("Vary"))
	})

	t.Run("specific origin: non-matching origin gets no ACAO header", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Origin", "https://evil.com")
		w := httptest.NewRecorder()

		CORSMiddleware("https://app.example.com", next).ServeHTTP(w, req)

		assert.Empty(t, w.Header().Get("Access-Control-Allow-Origin"))
		assert.Empty(t, w.Header().Get("Access-Control-Allow-Credentials"))
	})

	t.Run("specific origin: one of multiple allowed origins matches", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Origin", "https://admin.example.com")
		w := httptest.NewRecorder()

		CORSMiddleware("https://app.example.com,https://admin.example.com", next).ServeHTTP(w, req)

		assert.Equal(t, "https://admin.example.com", w.Header().Get("Access-Control-Allow-Origin"))
		assert.Equal(t, "true", w.Header().Get("Access-Control-Allow-Credentials"))
	})

	t.Run("passes non-OPTIONS request to next handler", func(t *testing.T) {
		t.Parallel()

		called := false
		req := httptest.NewRequest(http.MethodPost, "/", nil)
		w := httptest.NewRecorder()

		CORSMiddleware("*", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			called = true
			w.WriteHeader(http.StatusCreated)
		})).ServeHTTP(w, req)

		assert.True(t, called)
		assert.Equal(t, http.StatusCreated, w.Code)
	})
}
