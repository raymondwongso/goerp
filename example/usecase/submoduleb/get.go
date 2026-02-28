package submoduleb

import (
	"context"
	"fmt"

	"github.com/raymondwongso/goerp/domain"
)

type Get struct {
	exampleReader domain.ExampleReader
	exampleWriter domain.ExampleWriter
}

func NewGet(exampleReader domain.ExampleReader, exampleWriter domain.ExampleWriter) *Get {
	return &Get{
		exampleReader: exampleReader,
		exampleWriter: exampleWriter,
	}
}

func (u *Get) Invoke(ctx context.Context, id int64) (bool, error) {
	e, err := u.exampleReader.Get(ctx, id)
	if err != nil {
		return false, fmt.Errorf("error exampleReader.Get %w", err)
	}

	e, err = u.exampleWriter.Insert(ctx, e)
	if err != nil {
		return false, fmt.Errorf("error exampleWriter.Insert %w", err)
	}

	return true, nil
}
