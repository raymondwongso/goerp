package google

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/oauth2"
)

func TestProvider_GetAuthURL(t *testing.T) {
	t.Run("includes state, code_challenge, and code_challenge_method=S256", func(t *testing.T) {
		p := &provider{
			oauthConfig: &oauth2.Config{
				ClientID:    "client-id",
				RedirectURL: "http://localhost/callback",
				Endpoint: oauth2.Endpoint{
					AuthURL:  "https://example.com/auth",
					TokenURL: "https://example.com/token",
				},
			},
		}

		authURL := p.GetAuthURL("test-state", "test-challenge")

		parsed, err := url.Parse(authURL)
		assert.NoError(t, err)

		q := parsed.Query()
		assert.Equal(t, "test-state", q.Get("state"))
		assert.Equal(t, "test-challenge", q.Get("code_challenge"))
		assert.Equal(t, "S256", q.Get("code_challenge_method"))
	})
}

func TestProvider_Exchange(t *testing.T) {
	t.Run("error — token exchange HTTP failure", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "bad request", http.StatusBadRequest)
		}))
		defer server.Close()

		p := &provider{
			oauthConfig: &oauth2.Config{
				ClientID:     "client-id",
				ClientSecret: "client-secret",
				Endpoint: oauth2.Endpoint{
					AuthURL:   server.URL + "/auth",
					TokenURL:  server.URL + "/token",
					AuthStyle: oauth2.AuthStyleInParams,
				},
			},
		}

		_, err := p.Exchange(context.Background(), "auth-code", "code-verifier")
		assert.Error(t, err)
	})

	t.Run("error — missing id_token in response", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"access_token": "access-token",
				"token_type":   "Bearer",
				"expires_in":   3600,
			})
		}))
		defer server.Close()

		p := &provider{
			oauthConfig: &oauth2.Config{
				ClientID:     "client-id",
				ClientSecret: "client-secret",
				Endpoint: oauth2.Endpoint{
					AuthURL:   server.URL + "/auth",
					TokenURL:  server.URL + "/token",
					AuthStyle: oauth2.AuthStyleInParams,
				},
			},
		}

		_, err := p.Exchange(context.Background(), "auth-code", "code-verifier")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "missing id_token")
	})
}
