package google

import "context"

//go:generate mockgen -package=mockgoogle -source=$GOFILE -destination=mock/mock_$GOFILE

// TokenProvider abstracts the Google OAuth2 token exchange and ID token verification
type TokenProvider interface {
	GetAuthURL(state, codeChallenge string) string
	Exchange(ctx context.Context, code, codeVerifier string) (Claims, error)
}

// Claims holds the claims extracted from a Google ID token
type Claims struct {
	Sub     string
	Email   string
	Name    string
	Picture string
}
