package submoduleb

import (
	"context"
	"errors"
	"testing"

	"github.com/raymondwoongso/goerp/domain"
	"github.com/raymondwoongso/goerp/domain/xerror"
	examplemock "github.com/raymondwoongso/goerp/example/mock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type getTestSuite struct {
	exampleReader *examplemock.MockExampleReader
	exampleWriter *examplemock.MockExampleWriter
}

func newGetTestSuite(t *testing.T) *getTestSuite {
	ctrl := gomock.NewController(t)

	return &getTestSuite{
		exampleReader: examplemock.NewMockExampleReader(ctrl),
		exampleWriter: examplemock.NewMockExampleWriter(ctrl),
	}
}

func Test_Get(t *testing.T) {
	ctx := context.Background()
	validExample := domain.Example{ID: 1, Name: "Nama"}

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		ts := newGetTestSuite(t)

		ts.exampleReader.EXPECT().
			Get(ctx, int64(1)).
			Return(validExample, nil)

		ts.exampleWriter.EXPECT().
			Insert(ctx, validExample).
			Return(validExample, nil)

		res, err := NewGet(ts.exampleReader, ts.exampleWriter).Invoke(ctx, 1)
		assert.NoError(t, err)
		assert.True(t, res)
	})

	t.Run("error — error example reader get", func(t *testing.T) {
		t.Parallel()
		ts := newGetTestSuite(t)

		ts.exampleReader.EXPECT().
			Get(ctx, int64(1)).
			Return(domain.Example{}, errors.New("some error"))

		res, err := NewGet(ts.exampleReader, ts.exampleWriter).Invoke(ctx, 1)
		assert.Empty(t, res)
		assert.Error(t, err)
	})

	t.Run("error — error example reader get — wrapped", func(t *testing.T) {
		t.Parallel()
		ts := newGetTestSuite(t)

		ts.exampleReader.EXPECT().
			Get(ctx, int64(1)).
			Return(domain.Example{}, xerror.New(xerror.CodeInternal, "some error"))

		res, err := NewGet(ts.exampleReader, ts.exampleWriter).Invoke(ctx, 1)
		assert.Empty(t, res)
		assert.Error(t, err)
		assert.Equal(t, xerror.CodeInternal, xerror.GetCode(err))
	})

	t.Run("error — error example writer insert", func(t *testing.T) {
		t.Parallel()
		ts := newGetTestSuite(t)

		ts.exampleReader.EXPECT().
			Get(ctx, int64(1)).
			Return(validExample, nil)

		ts.exampleWriter.EXPECT().
			Insert(ctx, validExample).
			Return(domain.Example{}, xerror.New(xerror.CodeDuplicate, "some error"))

		res, err := NewGet(ts.exampleReader, ts.exampleWriter).Invoke(ctx, 1)
		assert.Empty(t, res)
		assert.Error(t, err)
		assert.Equal(t, xerror.CodeDuplicate, xerror.GetCode(err))
	})
}
