package auth

import "context"

//go:generate mockgen -package=mockdomainauth -source=$GOFILE -destination=mock/mock_auth.go

// GoogleLogin defines the interface for the google/login use case.
type GoogleLogin interface {
	Invoke(ctx context.Context, req GoogleLoginRequest) (GoogleLoginResult, error)
}

// GoogleCallback defines the interface for the google/callback use case.
type GoogleCallback interface {
	Invoke(ctx context.Context, req GoogleCallbackRequest) (GoogleCallbackResult, error)
}

// GoogleLoginRequest is the request for GoogleLogin usecase
type GoogleLoginRequest struct {
	RedirectTo string
	IPAddress  string
}

// GoogleLoginResult is the result for GoogleLogin usecase
type GoogleLoginResult struct {
	RedirectURL string
}

// GoogleCallbackRequest is the request for GoogleCallback usecase
type GoogleCallbackRequest struct {
	Code      string
	State     string
	UserAgent string
}

// GoogleCallbackResult is the result for GoogleCallback usecase
type GoogleCallbackResult struct {
	SessionID  string
	RedirectTo string
}
