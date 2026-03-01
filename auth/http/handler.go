package http

import (
	"encoding/json"
	"net/http"

	"github.com/raymondwongso/goerp/auth"
	domainauth "github.com/raymondwongso/goerp/domain/auth"
	"github.com/raymondwongso/goerp/domain/xhttp"
)

const sessionCookieMaxAge = 86400 * 30 // 30 days

// HandlerParam holds the dependencies for the HTTP handler
type HandlerParam struct {
	GoogleLogin    auth.GoogleLogin
	GoogleCallback auth.GoogleCallback
}

// Handler holds the HTTP handlers for the auth module
type Handler struct {
	googleLogin    auth.GoogleLogin
	googleCallback auth.GoogleCallback
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
	})
	if err != nil {
		writeError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"redirect_url": res.RedirectURL})
}

// GoogleCallback handles POST /auth/google/callback
// It processes the OAuth callback, sets an HttpOnly session cookie, and returns the post-login redirect destination.
func (h *Handler) GoogleCallback(w http.ResponseWriter, req *http.Request) {
	code := req.URL.Query().Get("code")
	state := req.URL.Query().Get("state")

	res, err := h.googleCallback.Invoke(req.Context(), domainauth.GoogleCallbackRequest{
		Code:      code,
		State:     state,
		IPAddress: req.RemoteAddr,
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
	json.NewEncoder(w).Encode(map[string]string{"redirect_to": res.RedirectTo})
}

func writeError(w http.ResponseWriter, err error) {
	code := xhttp.MapError(err)
	if code == 0 {
		code = http.StatusInternalServerError
	}
	w.WriteHeader(code)
}
