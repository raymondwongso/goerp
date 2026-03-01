package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	domainauth "github.com/raymondwongso/goerp/domain/auth"
	mockdomainauth "github.com/raymondwongso/goerp/domain/auth/mock"
	"github.com/raymondwongso/goerp/domain/xerror"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type handlerTestSuite struct {
	googleLogin    *mockdomainauth.MockGoogleLogin
	googleCallback *mockdomainauth.MockGoogleCallback
}

func newHandlerTestSuite(t *testing.T) *handlerTestSuite {
	ctrl := gomock.NewController(t)
	return &handlerTestSuite{
		googleLogin:    mockdomainauth.NewMockGoogleLogin(ctrl),
		googleCallback: mockdomainauth.NewMockGoogleCallback(ctrl),
	}
}

func (ts *handlerTestSuite) newHandler() *Handler {
	return NewHandler(HandlerParam{
		GoogleLogin:    ts.googleLogin,
		GoogleCallback: ts.googleCallback,
	})
}

func TestNewHandler(t *testing.T) {
	t.Run("panics when GoogleLogin is nil", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		googleCallback := mockdomainauth.NewMockGoogleCallback(ctrl)

		assert.Panics(t, func() {
			NewHandler(HandlerParam{
				GoogleLogin:    nil,
				GoogleCallback: googleCallback,
			})
		})
	})

	t.Run("panics when GoogleCallback is nil", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		googleLogin := mockdomainauth.NewMockGoogleLogin(ctrl)

		assert.Panics(t, func() {
			NewHandler(HandlerParam{
				GoogleLogin:    googleLogin,
				GoogleCallback: nil,
			})
		})
	})
}

func TestHandler_GoogleLogin(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		t.Parallel()
		ts := newHandlerTestSuite(t)

		ts.googleLogin.EXPECT().
			Invoke(gomock.Any(), domainauth.GoogleLoginRequest{RedirectTo: "/dashboard", IPAddress: "192.0.2.1"}).
			Return(domainauth.GoogleLoginResult{RedirectURL: "https://accounts.google.com/o/oauth2/v2/auth?state=abc"}, nil)

		req := httptest.NewRequest(http.MethodGet, "/auth/google/login?redirect_to=/dashboard", nil)
		w := httptest.NewRecorder()

		ts.newHandler().GoogleLogin(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

		var res map[string]string
		assert.NoError(t, json.NewDecoder(w.Body).Decode(&res))
		assert.Equal(t, "https://accounts.google.com/o/oauth2/v2/auth?state=abc", res["redirect_url"])
	})

	t.Run("success — no redirect_to", func(t *testing.T) {
		t.Parallel()
		ts := newHandlerTestSuite(t)

		ts.googleLogin.EXPECT().
			Invoke(gomock.Any(), domainauth.GoogleLoginRequest{IPAddress: "192.0.2.1"}).
			Return(domainauth.GoogleLoginResult{RedirectURL: "https://accounts.google.com/o/oauth2/v2/auth?state=abc"}, nil)

		req := httptest.NewRequest(http.MethodGet, "/auth/google/login", nil)
		w := httptest.NewRecorder()

		ts.newHandler().GoogleLogin(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("error — usecase internal error", func(t *testing.T) {
		t.Parallel()
		ts := newHandlerTestSuite(t)

		ts.googleLogin.EXPECT().
			Invoke(gomock.Any(), gomock.Any()).
			Return(domainauth.GoogleLoginResult{}, xerror.New(xerror.CodeInternal, "internal error"))

		req := httptest.NewRequest(http.MethodGet, "/auth/google/login", nil)
		w := httptest.NewRecorder()

		ts.newHandler().GoogleLogin(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("error — usecase unknown error", func(t *testing.T) {
		t.Parallel()
		ts := newHandlerTestSuite(t)

		ts.googleLogin.EXPECT().
			Invoke(gomock.Any(), gomock.Any()).
			Return(domainauth.GoogleLoginResult{}, errors.New("unexpected error"))

		req := httptest.NewRequest(http.MethodGet, "/auth/google/login", nil)
		w := httptest.NewRecorder()

		ts.newHandler().GoogleLogin(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestHandler_GoogleCallback(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		t.Parallel()
		ts := newHandlerTestSuite(t)

		ts.googleCallback.EXPECT().
			Invoke(gomock.Any(), gomock.Any()).
			Return(domainauth.GoogleCallbackResult{
				SessionID:  "session-id-1",
				RedirectTo: "/dashboard",
			}, nil)

		req := httptest.NewRequest(http.MethodGet, "/auth/google/callback?code=authcode&state=statetoken", nil)
		w := httptest.NewRecorder()

		ts.newHandler().GoogleCallback(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

		cookies := w.Result().Cookies()
		assert.Len(t, cookies, 1)
		assert.Equal(t, "session_id", cookies[0].Name)
		assert.Equal(t, "session-id-1", cookies[0].Value)
		assert.True(t, cookies[0].HttpOnly)
		assert.True(t, cookies[0].Secure)
		assert.Equal(t, sessionCookieMaxAge, cookies[0].MaxAge)

		var res map[string]string
		assert.NoError(t, json.NewDecoder(w.Body).Decode(&res))
		assert.Equal(t, "/dashboard", res["redirect_to"])
	})

	t.Run("error — usecase invalid parameter", func(t *testing.T) {
		t.Parallel()
		ts := newHandlerTestSuite(t)

		ts.googleCallback.EXPECT().
			Invoke(gomock.Any(), gomock.Any()).
			Return(domainauth.GoogleCallbackResult{}, xerror.New(xerror.CodeInvalidParameter, "code is required"))

		req := httptest.NewRequest(http.MethodGet, "/auth/google/callback?state=statetoken", nil)
		w := httptest.NewRecorder()

		ts.newHandler().GoogleCallback(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("error — usecase unauthorized", func(t *testing.T) {
		t.Parallel()
		ts := newHandlerTestSuite(t)

		ts.googleCallback.EXPECT().
			Invoke(gomock.Any(), gomock.Any()).
			Return(domainauth.GoogleCallbackResult{}, xerror.New(xerror.CodeUnauthorized, "invalid state"))

		req := httptest.NewRequest(http.MethodGet, "/auth/google/callback?code=code&state=invalid", nil)
		w := httptest.NewRecorder()

		ts.newHandler().GoogleCallback(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("error — usecase internal error", func(t *testing.T) {
		t.Parallel()
		ts := newHandlerTestSuite(t)

		ts.googleCallback.EXPECT().
			Invoke(gomock.Any(), gomock.Any()).
			Return(domainauth.GoogleCallbackResult{}, xerror.New(xerror.CodeInternal, "db error"))

		req := httptest.NewRequest(http.MethodGet, "/auth/google/callback?code=code&state=state", nil)
		w := httptest.NewRecorder()

		ts.newHandler().GoogleCallback(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}
