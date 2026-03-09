package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/akaporn-katip/go-project-structure-template/internal/application/repositories"
	"github.com/akaporn-katip/go-project-structure-template/internal/application/unitofwork"
	"github.com/jmoiron/sqlx"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

type UnitOfWork struct {
	unitofwork.UnitOfWork
	db                   *sqlx.DB
	meter                metric.Meter
	tracer               trace.Tracer
	transactionCounter   metric.Int64Counter
	transactionDuration  metric.Float64Histogram
	transactionRollbacks metric.Int64Counter
	activeTransactions   metric.Int64UpDownCounter
}

func NewUnitOfWork(db *sqlx.DB) (*UnitOfWork, error) {
	meter := otel.Meter("api.katipwork.com/crm/internal/infrastructure/persistence/postgres/unit_of_work")
	transactionCounter, err := meter.Int64Counter(
		"db.transaction.total",
		metric.WithDescription("Total number of database transactions"),
		metric.WithUnit("{transaction}"),
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create transaction counter: %w", err)
	}

	transactionDuration, err := meter.Float64Histogram(
		"db.transaction.duration",
		metric.WithDescription("Duration of database transactions"),
		metric.WithUnit("ms"),
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create transaction duration histogram: %w", err)
	}

	transactionRollbacks, err := meter.Int64Counter(
		"db.transaction.rollbacks",
		metric.WithDescription("Number of transaction rollbacks"),
		metric.WithUnit("{rollback}"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create rollback counter: %w", err)
	}

	activeTransactions, err := meter.Int64UpDownCounter(
		"db.transaction.active",
		metric.WithDescription("Number of active transactions"),
		metric.WithUnit("{transaction}"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create active transactions counter: %w", err)
	}

	return &UnitOfWork{
		db:     db,
		tracer: otel.Tracer("api.katipwork.com/crm/internal/infrastructure/persistence/postgres/unit_of_work"),

		// metrics
		meter:                meter,
		transactionCounter:   transactionCounter,
		transactionDuration:  transactionDuration,
		transactionRollbacks: transactionRollbacks,
		activeTransactions:   activeTransactions,
	}, nil
}

func (u *UnitOfWork) CreateTransaction() Transaction {
	return Transaction{
		uow: u,
	}
}

func (u *UnitOfWork) ExecuteTx(ctx context.Context, fn unitofwork.TxFunction) error {
	start := time.Now()
	tx := u.CreateTransaction()

	_, span := u.tracer.Start(ctx, "UnitOfWork.WithTx",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			attribute.String("db.system", "postgresql"),
			attribute.String("transaction.type", "read_write"),
		),
	)
	defer span.End()

	span.AddEvent("begin_transaction")
	if err := tx.Begin(ctx); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to begin transaction")
		return err
	}

	span.AddEvent("executing_function")

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback(ctx)
			panic(p)
		}
	}()

	err := fn(ctx, tx.CreatePostgresRepositoriesWithTx(ctx))
	if err != nil {
		span.AddEvent("function_error", trace.WithAttributes(
			attribute.String("error.message", err.Error()),
		))
		span.RecordError(err)

		span.AddEvent("attempting_rollback")
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			span.AddEvent("rollback_failed", trace.WithAttributes(
				attribute.String("rollback.error", rbErr.Error()),
			))
			span.SetStatus(codes.Error, "transaction failed and rollback failed")

			// Record duration even on error
			duration := time.Since(start).Milliseconds()
			u.transactionDuration.Record(ctx, float64(duration),
				metric.WithAttributes(
					attribute.String("db.system", "postgresql"),
					attribute.String("outcome", "error"),
				),
			)

			return fmt.Errorf("error: %v, rollback error: %v", err, rbErr)
		}

		span.AddEvent("rollback_successful")
		span.SetStatus(codes.Error, "transaction rolled back due to error")
		return err
	}

	span.AddEvent("attempting_commit")
	err = tx.Commit(ctx)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to commit transaction")

		duration := time.Since(start).Milliseconds()
		u.transactionDuration.Record(ctx, float64(duration),
			metric.WithAttributes(
				attribute.String("db.system", "postgresql"),
				attribute.String("outcome", "commit_failed"),
			),
		)

		return err
	}

	span.AddEvent("commit_successful")
	span.SetStatus(codes.Ok, "transaction completed successfully")

	duration := time.Since(start).Milliseconds()
	u.transactionDuration.Record(ctx, float64(duration),
		metric.WithAttributes(
			attribute.String("db.system", "postgresql"),
			attribute.String("outcome", "committed"),
		),
	)
	return nil
}

func WithTx[T any](ctx context.Context, fn unitofwork.TxFunctionWithResult[T], uow unitofwork.UnitOfWork) (T, error) {
	start := time.Now()

	var zero T
	concreteUow, ok := uow.(*UnitOfWork)
	if !ok {
		return zero, fmt.Errorf("infrastructure mismatch: expected postgres.UnitOfWork")
	}
	tx := concreteUow.CreateTransaction()

	_, span := concreteUow.tracer.Start(ctx, "UnitOfWork.WithTx",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			attribute.String("db.system", "postgresql"),
			attribute.String("transaction.type", "read_write"),
		),
	)
	defer span.End()

	span.AddEvent("begin_transaction")
	if err := tx.Begin(ctx); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to begin transaction")
		return zero, err
	}

	span.AddEvent("executing_function")

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback(ctx)
			panic(p)
		}
	}()

	rs, err := fn(ctx, tx.CreatePostgresRepositoriesWithTx(ctx))
	if err != nil {
		span.AddEvent("function_error", trace.WithAttributes(
			attribute.String("error.message", err.Error()),
		))
		span.RecordError(err)

		span.AddEvent("attempting_rollback")
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			span.AddEvent("rollback_failed", trace.WithAttributes(
				attribute.String("rollback.error", rbErr.Error()),
			))
			span.SetStatus(codes.Error, "transaction failed and rollback failed")

			// Record duration even on error
			duration := time.Since(start).Milliseconds()
			concreteUow.transactionDuration.Record(ctx, float64(duration),
				metric.WithAttributes(
					attribute.String("db.system", "postgresql"),
					attribute.String("outcome", "error"),
				),
			)

			return zero, fmt.Errorf("error: %v, rollback error: %v", err, rbErr)
		}

		span.AddEvent("rollback_successful")
		span.SetStatus(codes.Error, "transaction rolled back due to error")
		return zero, err
	}

	span.AddEvent("attempting_commit")
	err = tx.Commit(ctx)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to commit transaction")

		duration := time.Since(start).Milliseconds()
		concreteUow.transactionDuration.Record(ctx, float64(duration),
			metric.WithAttributes(
				attribute.String("db.system", "postgresql"),
				attribute.String("outcome", "commit_failed"),
			),
		)

		return zero, err
	}

	span.AddEvent("commit_successful")
	span.SetStatus(codes.Ok, "transaction completed successfully")

	duration := time.Since(start).Milliseconds()
	concreteUow.transactionDuration.Record(ctx, float64(duration),
		metric.WithAttributes(
			attribute.String("db.system", "postgresql"),
			attribute.String("outcome", "committed"),
		),
	)
	return rs, nil
}

func (uow *UnitOfWork) Repositories() repositories.Repositories {
	return NewPostgresRepositories(uow.db)
}
