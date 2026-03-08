package postgres

import (
	"context"
	"fmt"

	"github.com/akaporn-katip/go-project-structure-template/internal/application/repositories"
	"github.com/jmoiron/sqlx"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

type Transaction struct {
	uow *UnitOfWork
	tx  *sqlx.Tx
}

func (t *Transaction) Begin(ctx context.Context) error {
	if t.tx != nil {
		return fmt.Errorf("transaction already exists")
	}

	// Increment active transactions
	t.uow.activeTransactions.Add(ctx, 1,
		metric.WithAttributes(
			attribute.String("db.system", "postgresql"),
		),
	)

	tx, err := t.uow.db.BeginTxx(ctx, nil)
	if err != nil {
		t.uow.activeTransactions.Add(ctx, -1, metric.WithAttributes(
			attribute.String("db.system", "postgresql"),
		))

		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	t.tx = tx
	return nil
}

func (t *Transaction) Commit(ctx context.Context) error {
	if t.tx == nil {
		return fmt.Errorf("no active transaction")
	}

	err := t.tx.Commit()
	t.tx = nil

	// Decrement active transactions
	t.uow.activeTransactions.Add(ctx, -1,
		metric.WithAttributes(
			attribute.String("db.system", "postgresql"),
		),
	)

	// Record transaction completion
	t.uow.transactionCounter.Add(ctx, 1,
		metric.WithAttributes(
			attribute.String("db.system", "postgresql"),
			attribute.String("outcome", "committed"),
		),
	)

	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (t *Transaction) Rollback(ctx context.Context) error {
	if t.tx == nil {
		return nil
	}

	err := t.tx.Rollback()
	t.tx = nil

	// Decrement active transactions
	t.uow.activeTransactions.Add(ctx, -1,
		metric.WithAttributes(
			attribute.String("db.system", "postgresql"),
		),
	)

	// Record rollback
	t.uow.transactionRollbacks.Add(ctx, 1,
		metric.WithAttributes(
			attribute.String("db.system", "postgresql"),
		),
	)

	// Record transaction completion
	t.uow.transactionCounter.Add(ctx, 1,
		metric.WithAttributes(
			attribute.String("db.system", "postgresql"),
			attribute.String("outcome", "rolled_back"),
		),
	)

	if err != nil {
		return fmt.Errorf("failed to rollback transaction: %w", err)
	}

	return nil
}

func (t *Transaction) CreatePostgresRepositoriesWithTx(ctx context.Context) repositories.Repositories {
	return NewPostgresRepositories(t.tx)
}
