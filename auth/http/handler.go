package http

import (
	"encoding/json"
	"net"
	"net/http"

	domainauth "github.com/raymondwongso/goerp/domain/auth"
	"github.com/raymondwongso/goerp/domain/xhttp"
)

const sessionCookieMaxAge = 86400 * 30 // 30 days

// HandlerParam holds the dependencies for the HTTP handler
type HandlerParam struct {
	GoogleLogin    domainauth.GoogleLogin
	GoogleCallback domainauth.GoogleCallback
}

// Handler holds the HTTP handlers for the auth module
type Handler struct {
	googleLogin    domainauth.GoogleLogin
	googleCallback domainauth.GoogleCallback
}

// NewHandler creates a new Handler. Panics if any dependency is nil.
func NewHandler(param HandlerParam) *Handler {
	if param.GoogleLogin == nil {
		panic("GoogleLogin is empty")
	}
	if param.GoogleCallback == nil {
		panic("GoogleCallback is empty")
	}
	return &Handler{
		googleLogin:    param.GoogleLogin,
		googleCallback: param.GoogleCallback,
	}
}

// GoogleLogin handles PUT /auth/google/login
// It returns a JSON response with the Google OAuth redirect URL.
func (h *Handler) GoogleLogin(w http.ResponseWriter, req *http.Request) {
	redirectTo := req.URL.Query().Get("redirect_to")

	res, err := h.googleLogin.Invoke(req.Context(), domainauth.GoogleLoginRequest{
		RedirectTo: redirectTo,
		IPAddress:  remoteIP(req.RemoteAddr),
	})
	if err != nil {
		writeError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"redirect_url": res.RedirectURL})
}

// GoogleCallback handles POST /auth/google/callback
// It processes the OAuth callback, sets an HttpOnly session cookie, and returns the post-login redirect destination.
func (h *Handler) GoogleCallback(w http.ResponseWriter, req *http.Request) {
	code := req.URL.Query().Get("code")
	state := req.URL.Query().Get("state")

	res, err := h.googleCallback.Invoke(req.Context(), domainauth.GoogleCallbackRequest{
		Code:      code,
		State:     state,
		UserAgent: req.UserAgent(),
	})
	if err != nil {
		writeError(w, err)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    res.SessionID,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
		MaxAge:   sessionCookieMaxAge,
	})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"redirect_to": res.RedirectTo})
}

// remoteIP extracts the client IP from a "host:port" remote address.
// NOTE: This uses the TCP peer address directly. If the API is deployed behind a
// reverse proxy (Nginx, ALB, Cloudflare, etc.), RemoteAddr will be the proxy IP.
// Add X-Forwarded-For / X-Real-IP processing when a trusted proxy is introduced.
func remoteIP(remoteAddr string) string {
	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		// remoteAddr may be a Unix socket path or otherwise malformed; do not store it as an IP.
		return ""
	}
	return host
}

func writeError(w http.ResponseWriter, err error) {
	code := xhttp.MapError(err)
	if code == 0 {
		code = http.StatusInternalServerError
	}
	msg := err.Error()
	if code >= http.StatusInternalServerError {
		msg = "internal server error"
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
