package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/akaporn-katip/go-project-structure-template/internal/application/unitofwork"
	"github.com/akaporn-katip/go-project-structure-template/internal/domain/customerprofile"
	"github.com/jmoiron/sqlx"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

type UnitOfWork struct {
	tx     *sqlx.Tx
	db     *sqlx.DB
	tracer trace.Tracer

	// metrics
	meter                metric.Meter
	transactionCounter   metric.Int64Counter
	transactionDuration  metric.Float64Histogram
	transactionRollbacks metric.Int64Counter
	activeTransactions   metric.Int64UpDownCounter
}

func NewUnitOfWork(db *sqlx.DB, meter metric.Meter) (*UnitOfWork, error) {

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

func (uow *UnitOfWork) Begin(ctx context.Context) error {
	if uow.tx != nil {
		return fmt.Errorf("transaction already exists")
	}

	// Increment active transactions
	uow.activeTransactions.Add(ctx, 1,
		metric.WithAttributes(
			attribute.String("db.system", "postgresql"),
		),
	)

	tx, err := uow.db.BeginTxx(ctx, nil)
	if err != nil {
		uow.activeTransactions.Add(ctx, -1)

		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	uow.tx = tx
	return nil
}

func (uow *UnitOfWork) Commit(ctx context.Context) error {
	if uow.tx == nil {
		return fmt.Errorf("no active transaction")
	}

	err := uow.tx.Commit()
	uow.tx = nil

	// Decrement active transactions
	uow.activeTransactions.Add(ctx, -1,
		metric.WithAttributes(
			attribute.String("db.system", "postgresql"),
		),
	)

	// Record transaction completion
	uow.transactionCounter.Add(ctx, 1,
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

func (uow *UnitOfWork) Rollback(ctx context.Context) error {
	if uow.tx == nil {
		return nil
	}

	err := uow.tx.Rollback()
	uow.tx = nil

	// Decrement active transactions
	uow.activeTransactions.Add(ctx, -1,
		metric.WithAttributes(
			attribute.String("db.system", "postgresql"),
		),
	)

	// Record rollback
	uow.transactionRollbacks.Add(ctx, 1,
		metric.WithAttributes(
			attribute.String("db.system", "postgresql"),
		),
	)

	// Record transaction completion
	uow.transactionCounter.Add(ctx, 1,
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

func (uow *UnitOfWork) WithTx(ctx context.Context, fn func(ctx context.Context) error) error {
	if err := uow.Begin(ctx); err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			uow.Rollback(ctx)
			panic(p)
		}
	}()

	if err := fn(ctx); err != nil {
		if rbErr := uow.Rollback(ctx); rbErr != nil {
			return fmt.Errorf("error: %v, rollback error: %v", err, rbErr)
		}

		return err
	}

	return uow.Commit(ctx)
}

type TxFunction[T any] = func(ctx context.Context) (*T, error)

func WithTx[T any](ctx context.Context, fn TxFunction[T], uow unitofwork.UnitOfWork) (*T, error) {
	start := time.Now()
	concreteUow, ok := uow.(*UnitOfWork)
	if !ok {
		return nil, fmt.Errorf("invalid type: %T", uow)
	}

	_, span := concreteUow.tracer.Start(ctx, "UnitOfWork.WithTx",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			attribute.String("db.system", "postgresql"),
			attribute.String("transaction.type", "read_write"),
		),
	)
	defer span.End()

	span.AddEvent("begin_transaction")
	if err := concreteUow.Begin(ctx); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to begin transaction")
		return nil, err
	}

	span.AddEvent("executing_function")

	defer func() {
		if p := recover(); p != nil {
			concreteUow.Rollback(ctx)
			panic(p)
		}
	}()

	rs, err := fn(ctx)
	if err != nil {
		span.AddEvent("function_error", trace.WithAttributes(
			attribute.String("error.message", err.Error()),
		))
		span.RecordError(err)

		span.AddEvent("attempting_rollback")
		if rbErr := concreteUow.Rollback(ctx); rbErr != nil {
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

			return nil, fmt.Errorf("error: %v, rollback error: %v", err, rbErr)
		}

		span.AddEvent("rollback_successful")
		span.SetStatus(codes.Error, "transaction rolled back due to error")
		return nil, err
	}

	span.AddEvent("attempting_commit")
	err = concreteUow.Commit(ctx)
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

		return nil, err
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

func (uow *UnitOfWork) GetCustomerProfileRepository() customerprofile.Repository {
	executor := uow.getDb()
	wrapper := NewDatabaseWrapper(executor, uow.meter)
	return NewCustomerProfileRespository(wrapper)
}

func (uow *UnitOfWork) getDb() DatabaseExecutor {
	if uow.tx != nil {
		return uow.tx
	}
	return uow.db
}
