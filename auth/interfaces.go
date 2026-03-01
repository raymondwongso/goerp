package auth

import (
	"context"

	domainauth "github.com/raymondwongso/goerp/domain/auth"
)

//go:generate mockgen -package=mockauth -source=$GOFILE -destination=mock/mock_$GOFILE

// GoogleLogin defines interface for google/login usecase
type GoogleLogin interface {
	Invoke(ctx context.Context, req domainauth.GoogleLoginRequest) (domainauth.GoogleLoginResult, error)
}

// GoogleCallback defines interface for google/callback usecase
type GoogleCallback interface {
	Invoke(ctx context.Context, req domainauth.GoogleCallbackRequest) (domainauth.GoogleCallbackResult, error)
}
