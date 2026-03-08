package unitofwork

import (
	"context"

	"github.com/akaporn-katip/go-project-structure-template/internal/application/repositories"
)

type UnitOfWork interface {
	GetRepositories() repositories.Repositories
	ExecuteTx(ctx context.Context, fn TxFunction) error
}

type TxFunctionWithResult[T any] = func(ctx context.Context, repos repositories.Repositories) (T, error)
type TxFunction = func(ctx context.Context, repos repositories.Repositories) error

func ExecuteTx[T any](ctx context.Context, uow UnitOfWork, fn TxFunctionWithResult[T]) (T, error) {
	var result T
	err := uow.ExecuteTx(ctx, func(ctx context.Context, repos repositories.Repositories) error {
		res, err := fn(ctx, repos)
		result = res
		return err
	})

	if err != nil {
		var zero T
		return zero, err
	}

	return result, err
}
