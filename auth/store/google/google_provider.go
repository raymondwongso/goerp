package google

import (
	"context"
	"fmt"

	"github.com/coreos/go-oidc/v3/oidc"
	domaingoogle "github.com/raymondwongso/goerp/domain/google"
	"golang.org/x/oauth2"
	googleoauth "golang.org/x/oauth2/google"
)

// ProviderParam holds the configuration for creating a Google OAuth2 provider
type ProviderParam struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
	Scopes       []string
}

type provider struct {
	oauthConfig  *oauth2.Config
	oidcVerifier *oidc.IDTokenVerifier
}

// NewProvider creates a new Google OAuth2 + OIDC provider.
// It fetches Google's OIDC discovery document on initialization.
func NewProvider(ctx context.Context, param ProviderParam) (domaingoogle.TokenProvider, error) {
	oidcProvider, err := oidc.NewProvider(ctx, "https://accounts.google.com")
	if err != nil {
		return nil, fmt.Errorf("failed to create oidc provider: %w", err)
	}

	scopes := param.Scopes
	if len(scopes) == 0 {
		scopes = []string{oidc.ScopeOpenID, "email", "profile"}
	}

	oauthConfig := &oauth2.Config{
		ClientID:     param.ClientID,
		ClientSecret: param.ClientSecret,
		RedirectURL:  param.RedirectURL,
		Scopes:       scopes,
		Endpoint:     googleoauth.Endpoint,
	}

	return &provider{
		oauthConfig:  oauthConfig,
		oidcVerifier: oidcProvider.Verifier(&oidc.Config{ClientID: param.ClientID}),
	}, nil
}

func (p *provider) GetAuthURL(state, codeChallenge string) string {
	return p.oauthConfig.AuthCodeURL(state,
		oauth2.SetAuthURLParam("code_challenge", codeChallenge),
		oauth2.SetAuthURLParam("code_challenge_method", "S256"),
	)
}

func (p *provider) Exchange(ctx context.Context, code, codeVerifier string) (domaingoogle.Claims, error) {
	token, err := p.oauthConfig.Exchange(ctx, code,
		oauth2.SetAuthURLParam("code_verifier", codeVerifier),
	)
	if err != nil {
		return domaingoogle.Claims{}, err
	}

	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		return domaingoogle.Claims{}, fmt.Errorf("missing id_token in token response")
	}

	idToken, err := p.oidcVerifier.Verify(ctx, rawIDToken)
	if err != nil {
		return domaingoogle.Claims{}, err
	}

	var claims struct {
		Sub     string `json:"sub"`
		Email   string `json:"email"`
		Name    string `json:"name"`
		Picture string `json:"picture"`
	}
	if err := idToken.Claims(&claims); err != nil {
		return domaingoogle.Claims{}, err
	}

	return domaingoogle.Claims{
		Sub:     claims.Sub,
		Email:   claims.Email,
		Name:    claims.Name,
		Picture: claims.Picture,
	}, nil
}
