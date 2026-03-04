package xhttp

import (
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRemoteIP(t *testing.T) {
	t.Parallel()

	t.Run("returns X-Forwarded-For when present", func(t *testing.T) {
		t.Parallel()
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("X-Forwarded-For", "1.2.3.4")
		assert.Equal(t, "1.2.3.4", RemoteIP(req))
	})

	t.Run("returns X-Real-IP when X-Forwarded-For is absent", func(t *testing.T) {
		t.Parallel()
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("X-Real-IP", "5.6.7.8")
		assert.Equal(t, "5.6.7.8", RemoteIP(req))
	})

	t.Run("prefers X-Forwarded-For over X-Real-IP", func(t *testing.T) {
		t.Parallel()
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("X-Forwarded-For", "1.2.3.4")
		req.Header.Set("X-Real-IP", "5.6.7.8")
		assert.Equal(t, "1.2.3.4", RemoteIP(req))
	})

	t.Run("returns empty string when neither header is present", func(t *testing.T) {
		t.Parallel()
		req := httptest.NewRequest("GET", "/", nil)
		assert.Equal(t, "", RemoteIP(req))
	})
}
