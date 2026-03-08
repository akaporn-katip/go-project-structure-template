package mongodb

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/noop"
	"go.opentelemetry.io/otel/trace"
)

// CollectionExecutor replaces DatabaseExecutor
type CollectionExecutor interface {
	InsertOne(ctx context.Context, document interface{}, opts ...*options.InsertOneOptions) (*mongo.InsertOneResult, error)
	FindOne(ctx context.Context, filter interface{}, opts ...*options.FindOneOptions) *mongo.SingleResult
	// Add UpdateOne, DeleteOne, etc., as needed
}

type CollectionWrapper struct {
	coll           *mongo.Collection
	collectionName string
	tracer         trace.Tracer

	queryCounter  metric.Int64Counter
	queryDuration metric.Float64Histogram
	queryErrors   metric.Int64Counter
}

func NewCollectionWrapper(coll *mongo.Collection) *CollectionWrapper {
	wrapper := &CollectionWrapper{
		coll:           coll,
		collectionName: coll.Name(),
		tracer:         otel.Tracer("api.katipwork.com/crm/internal/infrastructure/persistence/mongodb/collection_wrapper"),
	}
	meter := otel.Meter("api.katipwork.com/crm/internal/infrastructure/persistence/mongodb/collection_wrapper")

	wrapper.initMetrics(meter)
	return wrapper
}

func (r *CollectionWrapper) initMetrics(meter metric.Meter) error {
	var err error
	r.queryCounter, err = meter.Int64Counter("db.query.total", metric.WithDescription("Total number of database queries"))
	if err != nil {
		r.queryCounter = noop.Int64Counter{}
	}

	r.queryDuration, err = meter.Float64Histogram("db.query.duration", metric.WithDescription("Duration of database queries"), metric.WithUnit("ms"))
	if err != nil {
		r.queryDuration = noop.Float64Histogram{}
	}

	r.queryErrors, err = meter.Int64Counter("db.query.errors", metric.WithDescription("Number of database query errors"))
	if err != nil {
		r.queryErrors = noop.Int64Counter{}
	}

	return nil
}

func (cw *CollectionWrapper) recordTelemetry(ctx context.Context, operation string, start time.Time, err error, span trace.Span) {
	duration := time.Since(start).Milliseconds()
	attrs := []attribute.KeyValue{
		attribute.String("db.system", "mongodb"),
		attribute.String("db.operation", operation),
		attribute.String("db.collection", cw.collectionName),
	}

	cw.queryCounter.Add(ctx, 1, metric.WithAttributes(attrs...))
	cw.queryDuration.Record(ctx, float64(duration), metric.WithAttributes(attrs...))

	if err != nil && err != mongo.ErrNoDocuments {
		cw.queryErrors.Add(ctx, 1, metric.WithAttributes(append(attrs, attribute.String("error.type", operation+"_error"))...))
		span.SetStatus(codes.Error, err.Error())
	} else {
		span.SetStatus(codes.Ok, operation+" success")
	}
}

func (cw *CollectionWrapper) InsertOne(ctx context.Context, document interface{}, opts ...*options.InsertOneOptions) (*mongo.InsertOneResult, error) {
	start := time.Now()
	_, span := cw.tracer.Start(ctx, "CollectionWrapper.InsertOne", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	res, err := cw.coll.InsertOne(ctx, document, opts...)
	cw.recordTelemetry(ctx, "InsertOne", start, err, span)
	return res, err
}

func (cw *CollectionWrapper) FindOne(ctx context.Context, filter interface{}, opts ...*options.FindOneOptions) *mongo.SingleResult {
	start := time.Now()
	_, span := cw.tracer.Start(ctx, "CollectionWrapper.FindOne", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	res := cw.coll.FindOne(ctx, filter, opts...)
	cw.recordTelemetry(ctx, "FindOne", start, res.Err(), span)
	return res
}
