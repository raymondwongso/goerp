package example

import (
	"context"

	"github.com/raymondwongso/goerp/domain"
)

//go:generate mockgen -package=mockexample -source=$GOFILE -destination=mock/mock_$GOFILE

// SubmoduleACreate defines interface for submodulea/create usecase
type SubmoduleACreate interface {
	Invoke(ctx context.Context, example domain.Example) (domain.Example, error)
}

// SubmoduleBGet defines interface for submoduleb/get usecase
type SubmoduleBGet interface {
	Invoke(ctx context.Context, id int64) (domain.Example, error)
}
