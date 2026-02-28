package domain

import (
	"context"
)

//go:generate mockgen -package=mockdomain -source=$GOFILE -destination=mock/mock_$GOFILE

// Example is a temporary struct used by example module. It doesn't serve any real purpose
type Example struct {
	ID   int64  `json:"id" db:"id"`
	Name string `json:"name" db:"name"`
}

// ExampleReader defines interface for Example repository read operations such as Get, List, Fetch
type ExampleReader interface {
	Get(ctx context.Context, id int64) (Example, error)
}

// ExampleWriter defines interface for Example repository write operations such as Insert, Update, SoftDelete
type ExampleWriter interface {
	Insert(ctx context.Context, req Example) (Example, error)
}
