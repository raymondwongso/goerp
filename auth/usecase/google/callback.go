package google

import (
	"context"

	"github.com/guregu/null"
	"github.com/raymondwongso/goerp/domain"
	domainauth "github.com/raymondwongso/goerp/domain/auth"
	domaingoogle "github.com/raymondwongso/goerp/domain/google"
	"github.com/raymondwongso/goerp/domain/xerror"
)

// Callback handles the Google OAuth callback flow
type Callback struct {
	tokenProvider      domaingoogle.TokenProvider
	oauthStateWriter   domain.OAuthStateWriter
	userWriter         domain.UserWriter
	oauthAccountWriter domain.OAuthAccountWriter
	sessionWriter      domain.SessionWriter
}

// NewCallback creates a new Callback use case
func NewCallback(
	tokenProvider domaingoogle.TokenProvider,
	oauthStateWriter domain.OAuthStateWriter,
	userWriter domain.UserWriter,
	oauthAccountWriter domain.OAuthAccountWriter,
	sessionWriter domain.SessionWriter,
) *Callback {
	return &Callback{
		tokenProvider:      tokenProvider,
		oauthStateWriter:   oauthStateWriter,
		userWriter:         userWriter,
		oauthAccountWriter: oauthAccountWriter,
		sessionWriter:      sessionWriter,
	}
}

// Invoke processes the Google OAuth callback.
// It consumes the oauth state, exchanges the code for tokens, upserts user and oauth account,
// creates a session, and returns the session ID and post-login redirect destination.
func (u *Callback) Invoke(ctx context.Context, req domainauth.GoogleCallbackRequest) (domainauth.GoogleCallbackResult, error) {
	if err := u.validate(req); err != nil {
		return domainauth.GoogleCallbackResult{}, err
	}

	oauthState, err := u.oauthStateWriter.DeleteByState(ctx, req.State)
	if err != nil {
		return domainauth.GoogleCallbackResult{}, xerror.NewWithCause(xerror.CodeUnauthorized, "invalid or expired state", err)
	}

	claims, err := u.tokenProvider.Exchange(ctx, req.Code, oauthState.CodeVerifier)
	if err != nil {
		return domainauth.GoogleCallbackResult{}, xerror.NewWithCause(xerror.CodeUnauthorized, "token exchange failed", err)
	}

	user, err := u.userWriter.Upsert(ctx, domain.User{
		Email:       claims.Email,
		DisplayName: null.NewString(claims.Name, claims.Name != ""),
		AvatarURL:   null.NewString(claims.Picture, claims.Picture != ""),
	})
	if err != nil {
		return domainauth.GoogleCallbackResult{}, xerror.NewWithCause(xerror.CodeInternal, "failed to upsert user", err)
	}

	_, err = u.oauthAccountWriter.Upsert(ctx, domain.OAuthAccount{
		UserID:      user.ID,
		Provider:    domain.OAuthProviderGoogle,
		ProviderSub: claims.Sub,
		Email:       claims.Email,
	})
	if err != nil {
		return domainauth.GoogleCallbackResult{}, xerror.NewWithCause(xerror.CodeInternal, "failed to upsert oauth account", err)
	}

	session, err := u.sessionWriter.Insert(ctx, domain.Session{
		UserID:    user.ID,
		IPAddress: null.NewString(req.IPAddress, req.IPAddress != ""),
		UserAgent: null.NewString(req.UserAgent, req.UserAgent != ""),
	})
	if err != nil {
		return domainauth.GoogleCallbackResult{}, xerror.NewWithCause(xerror.CodeInternal, "failed to create session", err)
	}

	redirectTo := "/"
	if oauthState.RedirectTo.Valid && oauthState.RedirectTo.String != "" {
		redirectTo = oauthState.RedirectTo.String
	}

	return domainauth.GoogleCallbackResult{
		SessionID:  session.ID,
		RedirectTo: redirectTo,
	}, nil
}

func (u *Callback) validate(req domainauth.GoogleCallbackRequest) error {
	if req.Code == "" {
		return xerror.New(xerror.CodeInvalidParameter, "code is required")
	}
	if req.State == "" {
		return xerror.New(xerror.CodeInvalidParameter, "state is required")
	}
	return nil
}
