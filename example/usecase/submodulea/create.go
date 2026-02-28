package submodulea

import (
	"context"

	"github.com/raymondwongso/goerp/domain"
	"github.com/raymondwongso/goerp/domain/xerror"
)

type Create struct{}

func NewCreate() *Create {
	return &Create{}
}

func (u *Create) Invoke(ctx context.Context, example domain.Example) (domain.Example, error) {
	if example.ID == 0 {
		return domain.Example{}, xerror.New(xerror.CodeUnprocessable, "some message")
	}
	example.ID = 2
	example.Name = "abc"
	return example, nil
}
