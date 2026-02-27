package example

import (
	"context"

	"github.com/raymondwoongso/goerp/domain"
)

//go:generate mockgen -package=mockexample -source=$GOFILE -destination=mock/mock_$GOFILE

// ExampleReader defines interface for Example repository read operations such as Get, List, Fetch
type ExampleReader interface {
	Get(ctx context.Context, id int64) (domain.Example, error)
}

// ExampleWriter defines interface for Example repository write operations such as Insert, Update, SoftDelete
type ExampleWriter interface {
	Insert(ctx context.Context, req domain.Example) (domain.Example, error)
}

// SubmoduleACreate defines interface for submodulea/create usecase
type SubmoduleACreate interface {
	Invoke(ctx context.Context, example domain.Example) (domain.Example, error)
}

// SubmoduleBGet defines interface for submoduleb/get usecase
type SubmoduleBGet interface {
	Invoke(ctx context.Context, id int64) (domain.Example, error)
}
