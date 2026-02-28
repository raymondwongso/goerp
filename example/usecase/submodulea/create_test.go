package submodulea

import (
	"context"
	"testing"

	"github.com/raymondwongso/goerp/domain"
	"github.com/raymondwongso/goerp/domain/xerror"
	"github.com/stretchr/testify/assert"
)

func Test_Create(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		res, err := NewCreate().Invoke(ctx, domain.Example{ID: 1, Name: "def"})
		assert.NoError(t, err)
		assert.Equal(t, int64(2), res.ID)
		assert.Equal(t, "abc", res.Name)
	})

	t.Run("error — empty example ID", func(t *testing.T) {
		t.Parallel()
		res, err := NewCreate().Invoke(ctx, domain.Example{})
		assert.Empty(t, res)
		assert.Error(t, err)
		assert.Equal(t, xerror.GetCode(err), xerror.CodeUnprocessable)
	})
}
