package submoduleb

import (
	"context"
	"fmt"

	"github.com/raymondwoongso/goerp/example"
)

type Get struct {
	exampleReader example.ExampleReader
	exampleWriter example.ExampleWriter
}

func NewGet(exampleReader example.ExampleReader, exampleWriter example.ExampleWriter) *Get {
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
