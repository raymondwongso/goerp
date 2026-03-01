package google

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"

	"github.com/guregu/null"
	domainauth "github.com/raymondwongso/goerp/domain/auth"
	"github.com/raymondwongso/goerp/domain"
	domaingoogle "github.com/raymondwongso/goerp/domain/google"
	"github.com/raymondwongso/goerp/domain/xerror"
)

// Login handles the Google OAuth login initiation
type Login struct {
	tokenProvider    domaingoogle.TokenProvider
	oauthStateWriter domain.OAuthStateWriter
}

// NewLogin creates a new Login use case
func NewLogin(tokenProvider domaingoogle.TokenProvider, oauthStateWriter domain.OAuthStateWriter) *Login {
	return &Login{
		tokenProvider:    tokenProvider,
		oauthStateWriter: oauthStateWriter,
	}
}

// Invoke initiates the Google OAuth login flow.
// It generates a state and PKCE challenge, stores the oauth state, and returns the Google redirect URL.
func (u *Login) Invoke(ctx context.Context, req domainauth.GoogleLoginRequest) (domainauth.GoogleLoginResult, error) {
	state, err := u.generateState()
	if err != nil {
		return domainauth.GoogleLoginResult{}, xerror.NewWithCause(xerror.CodeInternal, "failed to generate state", err)
	}

	verifier, challenge, err := u.generatePKCE()
	if err != nil {
		return domainauth.GoogleLoginResult{}, xerror.NewWithCause(xerror.CodeInternal, "failed to generate PKCE", err)
	}

	_, err = u.oauthStateWriter.Insert(ctx, domain.OAuthState{
		State:        state,
		CodeVerifier: verifier,
		RedirectTo:   null.NewString(req.RedirectTo, req.RedirectTo != ""),
		IPAddress:    null.NewString(req.IPAddress, req.IPAddress != ""),
	})
	if err != nil {
		return domainauth.GoogleLoginResult{}, xerror.NewWithCause(xerror.CodeInternal, "failed to insert oauth state", err)
	}

	return domainauth.GoogleLoginResult{
		RedirectURL: u.tokenProvider.GetAuthURL(state, challenge),
	}, nil
}

func (u *Login) generateState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func (u *Login) generatePKCE() (verifier, challenge string, err error) {
	b := make([]byte, 64)
	if _, err = rand.Read(b); err != nil {
		return
	}
	verifier = base64.RawURLEncoding.EncodeToString(b)
	sum := sha256.Sum256([]byte(verifier))
	challenge = base64.RawURLEncoding.EncodeToString(sum[:])
	return
}
