package google

import (
	"context"
	"errors"
	"testing"

	"github.com/guregu/null"
	"github.com/raymondwongso/goerp/domain"
	domainauth "github.com/raymondwongso/goerp/domain/auth"
	mockdomain "github.com/raymondwongso/goerp/domain/mock"
	domaingoogle "github.com/raymondwongso/goerp/domain/google"
	mockgoogle "github.com/raymondwongso/goerp/domain/google/mock"
	"github.com/raymondwongso/goerp/domain/xerror"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type callbackTestSuite struct {
	tokenProvider      *mockgoogle.MockTokenProvider
	oauthStateWriter   *mockdomain.MockOAuthStateWriter
	userWriter         *mockdomain.MockUserWriter
	oauthAccountWriter *mockdomain.MockOAuthAccountWriter
	sessionWriter      *mockdomain.MockSessionWriter
}

func newCallbackTestSuite(t *testing.T) *callbackTestSuite {
	ctrl := gomock.NewController(t)
	return &callbackTestSuite{
		tokenProvider:      mockgoogle.NewMockTokenProvider(ctrl),
		oauthStateWriter:   mockdomain.NewMockOAuthStateWriter(ctrl),
		userWriter:         mockdomain.NewMockUserWriter(ctrl),
		oauthAccountWriter: mockdomain.NewMockOAuthAccountWriter(ctrl),
		sessionWriter:      mockdomain.NewMockSessionWriter(ctrl),
	}
}

func (ts *callbackTestSuite) newCallback() *Callback {
	return NewCallback(
		ts.tokenProvider,
		ts.oauthStateWriter,
		ts.userWriter,
		ts.oauthAccountWriter,
		ts.sessionWriter,
	)
}

func Test_Callback(t *testing.T) {
	ctx := context.Background()

	validReq := domainauth.GoogleCallbackRequest{
		Code:      "auth-code",
		State:     "random-state",
		IPAddress: "192.168.1.1",
		UserAgent: "Mozilla/5.0",
	}
	validOAuthState := domain.OAuthState{
		State:        "random-state",
		CodeVerifier: "code-verifier",
		RedirectTo:   null.StringFrom("/dashboard"),
	}
	validClaims := domaingoogle.Claims{
		Sub:     "google-sub-123",
		Email:   "user@example.com",
		Name:    "Test User",
		Picture: "https://example.com/pic.jpg",
	}
	validUser := domain.User{
		ID:          "user-id-1",
		Email:       "user@example.com",
		DisplayName: null.StringFrom("Test User"),
		AvatarURL:   null.StringFrom("https://example.com/pic.jpg"),
	}
	validSession := domain.Session{
		ID:     "session-id-1",
		UserID: "user-id-1",
	}

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		ts := newCallbackTestSuite(t)

		ts.oauthStateWriter.EXPECT().
			DeleteByState(ctx, "random-state").
			Return(validOAuthState, nil)

		ts.tokenProvider.EXPECT().
			Exchange(ctx, "auth-code", "code-verifier").
			Return(validClaims, nil)

		ts.userWriter.EXPECT().
			Upsert(ctx, domain.User{
				Email:       "user@example.com",
				DisplayName: null.StringFrom("Test User"),
				AvatarURL:   null.StringFrom("https://example.com/pic.jpg"),
			}).
			Return(validUser, nil)

		ts.oauthAccountWriter.EXPECT().
			Upsert(ctx, domain.OAuthAccount{
				UserID:      "user-id-1",
				Provider:    domain.OAuthProviderGoogle,
				ProviderSub: "google-sub-123",
				Email:       "user@example.com",
			}).
			Return(domain.OAuthAccount{}, nil)

		ts.sessionWriter.EXPECT().
			Insert(ctx, domain.Session{
				UserID:    "user-id-1",
				IPAddress: null.StringFrom("192.168.1.1"),
				UserAgent: null.StringFrom("Mozilla/5.0"),
			}).
			Return(validSession, nil)

		res, err := ts.newCallback().Invoke(ctx, validReq)

		assert.NoError(t, err)
		assert.Equal(t, "session-id-1", res.SessionID)
		assert.Equal(t, "/dashboard", res.RedirectTo)
	})

	t.Run("success — redirect_to defaults to /", func(t *testing.T) {
		t.Parallel()
		ts := newCallbackTestSuite(t)

		stateWithNoRedirect := domain.OAuthState{
			State:        "random-state",
			CodeVerifier: "code-verifier",
			RedirectTo:   null.NewString("", false),
		}

		ts.oauthStateWriter.EXPECT().
			DeleteByState(ctx, "random-state").
			Return(stateWithNoRedirect, nil)

		ts.tokenProvider.EXPECT().
			Exchange(ctx, "auth-code", "code-verifier").
			Return(validClaims, nil)

		ts.userWriter.EXPECT().
			Upsert(gomock.Any(), gomock.Any()).
			Return(validUser, nil)

		ts.oauthAccountWriter.EXPECT().
			Upsert(gomock.Any(), gomock.Any()).
			Return(domain.OAuthAccount{}, nil)

		ts.sessionWriter.EXPECT().
			Insert(gomock.Any(), gomock.Any()).
			Return(validSession, nil)

		res, err := ts.newCallback().Invoke(ctx, validReq)

		assert.NoError(t, err)
		assert.Equal(t, "/", res.RedirectTo)
	})

	t.Run("error — code is required", func(t *testing.T) {
		t.Parallel()
		ts := newCallbackTestSuite(t)

		res, err := ts.newCallback().Invoke(ctx, domainauth.GoogleCallbackRequest{State: "state"})

		assert.Empty(t, res)
		assert.Error(t, err)
		assert.Equal(t, xerror.CodeInvalidParameter, xerror.GetCode(err))
	})

	t.Run("error — state is required", func(t *testing.T) {
		t.Parallel()
		ts := newCallbackTestSuite(t)

		res, err := ts.newCallback().Invoke(ctx, domainauth.GoogleCallbackRequest{Code: "code"})

		assert.Empty(t, res)
		assert.Error(t, err)
		assert.Equal(t, xerror.CodeInvalidParameter, xerror.GetCode(err))
	})

	t.Run("error — oauth state delete failed", func(t *testing.T) {
		t.Parallel()
		ts := newCallbackTestSuite(t)

		ts.oauthStateWriter.EXPECT().
			DeleteByState(ctx, "random-state").
			Return(domain.OAuthState{}, errors.New("state not found"))

		res, err := ts.newCallback().Invoke(ctx, validReq)

		assert.Empty(t, res)
		assert.Error(t, err)
		assert.Equal(t, xerror.CodeUnauthorized, xerror.GetCode(err))
	})

	t.Run("error — token exchange failed", func(t *testing.T) {
		t.Parallel()
		ts := newCallbackTestSuite(t)

		ts.oauthStateWriter.EXPECT().
			DeleteByState(ctx, "random-state").
			Return(validOAuthState, nil)

		ts.tokenProvider.EXPECT().
			Exchange(ctx, "auth-code", "code-verifier").
			Return(domaingoogle.Claims{}, errors.New("exchange failed"))

		res, err := ts.newCallback().Invoke(ctx, validReq)

		assert.Empty(t, res)
		assert.Error(t, err)
		assert.Equal(t, xerror.CodeUnauthorized, xerror.GetCode(err))
	})

	t.Run("error — user upsert failed", func(t *testing.T) {
		t.Parallel()
		ts := newCallbackTestSuite(t)

		ts.oauthStateWriter.EXPECT().
			DeleteByState(ctx, "random-state").
			Return(validOAuthState, nil)

		ts.tokenProvider.EXPECT().
			Exchange(ctx, "auth-code", "code-verifier").
			Return(validClaims, nil)

		ts.userWriter.EXPECT().
			Upsert(ctx, gomock.Any()).
			Return(domain.User{}, errors.New("db error"))

		res, err := ts.newCallback().Invoke(ctx, validReq)

		assert.Empty(t, res)
		assert.Error(t, err)
		assert.Equal(t, xerror.CodeInternal, xerror.GetCode(err))
	})

	t.Run("error — oauth account upsert failed", func(t *testing.T) {
		t.Parallel()
		ts := newCallbackTestSuite(t)

		ts.oauthStateWriter.EXPECT().
			DeleteByState(ctx, "random-state").
			Return(validOAuthState, nil)

		ts.tokenProvider.EXPECT().
			Exchange(ctx, "auth-code", "code-verifier").
			Return(validClaims, nil)

		ts.userWriter.EXPECT().
			Upsert(ctx, gomock.Any()).
			Return(validUser, nil)

		ts.oauthAccountWriter.EXPECT().
			Upsert(ctx, gomock.Any()).
			Return(domain.OAuthAccount{}, errors.New("db error"))

		res, err := ts.newCallback().Invoke(ctx, validReq)

		assert.Empty(t, res)
		assert.Error(t, err)
		assert.Equal(t, xerror.CodeInternal, xerror.GetCode(err))
	})

	t.Run("error — session insert failed", func(t *testing.T) {
		t.Parallel()
		ts := newCallbackTestSuite(t)

		ts.oauthStateWriter.EXPECT().
			DeleteByState(ctx, "random-state").
			Return(validOAuthState, nil)

		ts.tokenProvider.EXPECT().
			Exchange(ctx, "auth-code", "code-verifier").
			Return(validClaims, nil)

		ts.userWriter.EXPECT().
			Upsert(ctx, gomock.Any()).
			Return(validUser, nil)

		ts.oauthAccountWriter.EXPECT().
			Upsert(ctx, gomock.Any()).
			Return(domain.OAuthAccount{}, nil)

		ts.sessionWriter.EXPECT().
			Insert(ctx, gomock.Any()).
			Return(domain.Session{}, errors.New("db error"))

		res, err := ts.newCallback().Invoke(ctx, validReq)

		assert.Empty(t, res)
		assert.Error(t, err)
		assert.Equal(t, xerror.CodeInternal, xerror.GetCode(err))
	})
}
