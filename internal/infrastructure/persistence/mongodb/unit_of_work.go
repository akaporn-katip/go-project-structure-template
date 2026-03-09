package mongodb

import (
	"context"
	"fmt"
	"time"

	"github.com/akaporn-katip/go-project-structure-template/internal/application/repositories"
	"github.com/akaporn-katip/go-project-structure-template/internal/application/unitofwork"
	"go.mongodb.org/mongo-driver/mongo"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

type UnitOfWork struct {
	unitofwork.UnitOfWork
	client *mongo.Client
	dbName string
	meter  metric.Meter
	tracer trace.Tracer

	transactionCounter   metric.Int64Counter
	transactionDuration  metric.Float64Histogram
	transactionRollbacks metric.Int64Counter
}

func NewUnitOfWork(client *mongo.Client, dbName string) (*UnitOfWork, error) {
	meter := otel.Meter("api.katipwork.com/crm/internal/infrastructure/persistence/mongodb/unit_of_work")

	transactionCounter, _ := meter.Int64Counter("db.transaction.total", metric.WithDescription("Total number of database transactions"))
	transactionDuration, _ := meter.Float64Histogram("db.transaction.duration", metric.WithDescription("Duration of database transactions"), metric.WithUnit("ms"))
	transactionRollbacks, _ := meter.Int64Counter("db.transaction.rollbacks", metric.WithDescription("Number of transaction rollbacks"))

	return &UnitOfWork{
		client:               client,
		dbName:               dbName,
		tracer:               otel.Tracer("api.katipwork.com/crm/internal/infrastructure/persistence/mongodb/unit_of_work"),
		meter:                meter,
		transactionCounter:   transactionCounter,
		transactionDuration:  transactionDuration,
		transactionRollbacks: transactionRollbacks,
	}, nil
}

// ExecuteTx handles transactions without returning a result
func (u *UnitOfWork) ExecuteTx(ctx context.Context, fn unitofwork.TxFunction) error {
	start := time.Now()

	_, span := u.tracer.Start(ctx, "UnitOfWork.ExecuteTx",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(attribute.String("db.system", "mongodb")),
	)
	defer span.End()

	session, err := u.client.StartSession()
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to start session")
		return err
	}
	defer session.EndSession(ctx)

	// MongoDB handles the commit/rollback logic via WithTransaction
	callback := func(sessCtx mongo.SessionContext) (interface{}, error) {
		repos := NewMongoRepositories(u.client.Database(u.dbName))
		return nil, fn(sessCtx, repos)
	}

	_, err = session.WithTransaction(ctx, callback)

	duration := float64(time.Since(start).Milliseconds())
	if err != nil {
		u.transactionRollbacks.Add(ctx, 1, metric.WithAttributes(attribute.String("db.system", "mongodb")))
		u.transactionDuration.Record(ctx, duration, metric.WithAttributes(attribute.String("outcome", "error")))
		return err
	}

	u.transactionCounter.Add(ctx, 1, metric.WithAttributes(attribute.String("outcome", "committed")))
	u.transactionDuration.Record(ctx, duration, metric.WithAttributes(attribute.String("outcome", "committed")))
	return nil
}

// WithTx handles transactions that return a result
func WithTx[T any](ctx context.Context, fn unitofwork.TxFunctionWithResult[T], uow unitofwork.UnitOfWork) (T, error) {
	start := time.Now()
	var zero T

	concreteUow, ok := uow.(*UnitOfWork)
	if !ok {
		return zero, fmt.Errorf("infrastructure mismatch: expected mongodb.UnitOfWork")
	}

	_, span := concreteUow.tracer.Start(ctx, "UnitOfWork.WithTx",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(attribute.String("db.system", "mongodb")),
	)
	defer span.End()

	session, err := concreteUow.client.StartSession()
	if err != nil {
		span.RecordError(err)
		return zero, err
	}
	defer session.EndSession(ctx)

	var result T
	callback := func(sessCtx mongo.SessionContext) (interface{}, error) {
		repos := NewMongoRepositories(concreteUow.client.Database(concreteUow.dbName))
		res, fnErr := fn(sessCtx, repos)
		if fnErr != nil {
			return nil, fnErr
		}
		result = res
		return res, nil
	}

	_, err = session.WithTransaction(ctx, callback)

	duration := float64(time.Since(start).Milliseconds())
	if err != nil {
		concreteUow.transactionRollbacks.Add(ctx, 1, metric.WithAttributes(attribute.String("db.system", "mongodb")))
		concreteUow.transactionDuration.Record(ctx, duration, metric.WithAttributes(attribute.String("outcome", "error")))
		return zero, err
	}

	concreteUow.transactionCounter.Add(ctx, 1, metric.WithAttributes(attribute.String("outcome", "committed")))
	concreteUow.transactionDuration.Record(ctx, duration, metric.WithAttributes(attribute.String("outcome", "committed")))
	return result, nil
}

func (uow *UnitOfWork) Repositories() repositories.Repositories {
	return NewMongoRepositories(uow.client.Database(uow.dbName))
}
