package google

import (
	"context"
	"errors"
	"testing"

	"github.com/raymondwongso/goerp/domain"
	mockdomain "github.com/raymondwongso/goerp/domain/mock"
	domainauth "github.com/raymondwongso/goerp/domain/auth"
	mockgoogle "github.com/raymondwongso/goerp/domain/google/mock"
	"github.com/raymondwongso/goerp/domain/xerror"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type loginTestSuite struct {
	tokenProvider    *mockgoogle.MockTokenProvider
	oauthStateWriter *mockdomain.MockOAuthStateWriter
}

func newLoginTestSuite(t *testing.T) *loginTestSuite {
	ctrl := gomock.NewController(t)
	return &loginTestSuite{
		tokenProvider:    mockgoogle.NewMockTokenProvider(ctrl),
		oauthStateWriter: mockdomain.NewMockOAuthStateWriter(ctrl),
	}
}

func Test_Login(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		ts := newLoginTestSuite(t)

		ts.oauthStateWriter.EXPECT().
			Insert(ctx, gomock.Any()).
			Return(domain.OAuthState{}, nil)

		ts.tokenProvider.EXPECT().
			GetAuthURL(gomock.Any(), gomock.Any()).
			Return("https://accounts.google.com/o/oauth2/v2/auth?state=abc")

		res, err := NewLogin(ts.tokenProvider, ts.oauthStateWriter).Invoke(ctx, domainauth.GoogleLoginRequest{
			RedirectTo: "/dashboard",
		})

		assert.NoError(t, err)
		assert.NotEmpty(t, res.RedirectURL)
	})

	t.Run("success — empty redirect_to", func(t *testing.T) {
		t.Parallel()
		ts := newLoginTestSuite(t)

		ts.oauthStateWriter.EXPECT().
			Insert(ctx, gomock.Any()).
			Return(domain.OAuthState{}, nil)

		ts.tokenProvider.EXPECT().
			GetAuthURL(gomock.Any(), gomock.Any()).
			Return("https://accounts.google.com/o/oauth2/v2/auth?state=abc")

		res, err := NewLogin(ts.tokenProvider, ts.oauthStateWriter).Invoke(ctx, domainauth.GoogleLoginRequest{})

		assert.NoError(t, err)
		assert.NotEmpty(t, res.RedirectURL)
	})

	t.Run("error — oauth state insert failed", func(t *testing.T) {
		t.Parallel()
		ts := newLoginTestSuite(t)

		ts.oauthStateWriter.EXPECT().
			Insert(ctx, gomock.Any()).
			Return(domain.OAuthState{}, errors.New("db error"))

		res, err := NewLogin(ts.tokenProvider, ts.oauthStateWriter).Invoke(ctx, domainauth.GoogleLoginRequest{})

		assert.Empty(t, res)
		assert.Error(t, err)
		assert.Equal(t, xerror.CodeInternal, xerror.GetCode(err))
	})
}
