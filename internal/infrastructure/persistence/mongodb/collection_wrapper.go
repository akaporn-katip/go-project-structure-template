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

// SingleResult is an interface describing a single result from MongoDB.
type SingleResult interface {
	Decode(v interface{}) error
	Err() error
}

// CollectionExecutor replaces DatabaseExecutor
type CollectionExecutor interface {
	InsertOne(ctx context.Context, document interface{}, opts ...*options.InsertOneOptions) (*mongo.InsertOneResult, error)
	FindOne(ctx context.Context, filter interface{}, opts ...*options.FindOneOptions) SingleResult
	Aggregate(ctx context.Context, pipeline interface{}, opts ...*options.AggregateOptions) (*mongo.Cursor, error)
	BulkWrite(ctx context.Context, models []mongo.WriteModel, opts ...*options.BulkWriteOptions) (*mongo.BulkWriteResult, error)
	CountDocuments(ctx context.Context, filter interface{}, opts ...*options.CountOptions) (int64, error)
	DeleteMany(ctx context.Context, filter interface{}, opts ...*options.DeleteOptions) (*mongo.DeleteResult, error)
	DeleteOne(ctx context.Context, filter interface{}, opts ...*options.DeleteOptions) (*mongo.DeleteResult, error)
	Distinct(ctx context.Context, fieldName string, filter interface{}, opts ...*options.DistinctOptions) ([]interface{}, error)
	Find(ctx context.Context, filter interface{}, opts ...*options.FindOptions) (*mongo.Cursor, error)
	FindOneAndDelete(ctx context.Context, filter interface{}, opts ...*options.FindOneAndDeleteOptions) SingleResult
	FindOneAndReplace(ctx context.Context, filter interface{}, replacement interface{}, opts ...*options.FindOneAndReplaceOptions) SingleResult
	FindOneAndUpdate(ctx context.Context, filter interface{}, update interface{}, opts ...*options.FindOneAndUpdateOptions) SingleResult
	InsertMany(ctx context.Context, documents []interface{}, opts ...*options.InsertManyOptions) (*mongo.InsertManyResult, error)
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
		tracer:         otel.Tracer("github.com/akaporn-katip/go-project-structure-template/internal/infrastructure/persistence/mongodb/collection_wrapper"),
	}
	meter := otel.Meter("github.com/akaporn-katip/go-project-structure-template/internal/infrastructure/persistence/mongodb/collection_wrapper")

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

func (cw *CollectionWrapper) FindOne(ctx context.Context, filter interface{}, opts ...*options.FindOneOptions) SingleResult {
	start := time.Now()
	_, span := cw.tracer.Start(ctx, "CollectionWrapper.FindOne", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	res := cw.coll.FindOne(ctx, filter, opts...)
	cw.recordTelemetry(ctx, "FindOne", start, res.Err(), span)
	return res
}

func (cw *CollectionWrapper) Aggregate(ctx context.Context, pipeline interface{}, opts ...*options.AggregateOptions) (*mongo.Cursor, error) {
	start := time.Now()
	_, span := cw.tracer.Start(ctx, "CollectionWrapper.Aggregate", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	res, err := cw.coll.Aggregate(ctx, pipeline, opts...)
	cw.recordTelemetry(ctx, "Aggregate", start, err, span)
	return res, err
}

func (cw *CollectionWrapper) BulkWrite(ctx context.Context, models []mongo.WriteModel, opts ...*options.BulkWriteOptions) (*mongo.BulkWriteResult, error) {
	start := time.Now()
	_, span := cw.tracer.Start(ctx, "CollectionWrapper.BulkWrite", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	res, err := cw.coll.BulkWrite(ctx, models, opts...)
	cw.recordTelemetry(ctx, "BulkWrite", start, err, span)
	return res, err
}

func (cw *CollectionWrapper) CountDocuments(ctx context.Context, filter interface{}, opts ...*options.CountOptions) (int64, error) {
	start := time.Now()
	_, span := cw.tracer.Start(ctx, "CollectionWrapper.CountDocuments", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	res, err := cw.coll.CountDocuments(ctx, filter, opts...)
	cw.recordTelemetry(ctx, "CountDocuments", start, err, span)
	return res, err
}

func (cw *CollectionWrapper) DeleteMany(ctx context.Context, filter interface{}, opts ...*options.DeleteOptions) (*mongo.DeleteResult, error) {
	start := time.Now()
	_, span := cw.tracer.Start(ctx, "CollectionWrapper.DeleteMany", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	res, err := cw.coll.DeleteMany(ctx, filter, opts...)
	cw.recordTelemetry(ctx, "DeleteMany", start, err, span)
	return res, err
}

func (cw *CollectionWrapper) DeleteOne(ctx context.Context, filter interface{}, opts ...*options.DeleteOptions) (*mongo.DeleteResult, error) {
	start := time.Now()
	_, span := cw.tracer.Start(ctx, "CollectionWrapper.DeleteOne", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	res, err := cw.coll.DeleteOne(ctx, filter, opts...)
	cw.recordTelemetry(ctx, "DeleteOne", start, err, span)
	return res, err
}

func (cw *CollectionWrapper) Distinct(ctx context.Context, fieldName string, filter interface{}, opts ...*options.DistinctOptions) ([]interface{}, error) {
	start := time.Now()
	_, span := cw.tracer.Start(ctx, "CollectionWrapper.Distinct", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	res, err := cw.coll.Distinct(ctx, fieldName, filter, opts...)
	cw.recordTelemetry(ctx, "Distinct", start, err, span)
	return res, err
}

func (cw *CollectionWrapper) Find(ctx context.Context, filter interface{}, opts ...*options.FindOptions) (*mongo.Cursor, error) {
	start := time.Now()
	_, span := cw.tracer.Start(ctx, "CollectionWrapper.Find", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	res, err := cw.coll.Find(ctx, filter, opts...)
	cw.recordTelemetry(ctx, "Find", start, err, span)
	return res, err
}

func (cw *CollectionWrapper) FindOneAndDelete(ctx context.Context, filter interface{}, opts ...*options.FindOneAndDeleteOptions) SingleResult {
	start := time.Now()
	_, span := cw.tracer.Start(ctx, "CollectionWrapper.FindOneAndDelete", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	res := cw.coll.FindOneAndDelete(ctx, filter, opts...)
	cw.recordTelemetry(ctx, "FindOneAndDelete", start, res.Err(), span)
	return res
}

func (cw *CollectionWrapper) FindOneAndReplace(ctx context.Context, filter interface{}, replacement interface{}, opts ...*options.FindOneAndReplaceOptions) SingleResult {
	start := time.Now()
	_, span := cw.tracer.Start(ctx, "CollectionWrapper.FindOneAndReplace", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	res := cw.coll.FindOneAndReplace(ctx, filter, replacement, opts...)
	cw.recordTelemetry(ctx, "FindOneAndReplace", start, res.Err(), span)
	return res
}

func (cw *CollectionWrapper) FindOneAndUpdate(ctx context.Context, filter interface{}, update interface{}, opts ...*options.FindOneAndUpdateOptions) SingleResult {
	start := time.Now()
	_, span := cw.tracer.Start(ctx, "CollectionWrapper.FindOneAndUpdate", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	res := cw.coll.FindOneAndUpdate(ctx, filter, update, opts...)
	cw.recordTelemetry(ctx, "FindOneAndUpdate", start, res.Err(), span)
	return res
}

func (cw *CollectionWrapper) InsertMany(ctx context.Context, documents []interface{}, opts ...*options.InsertManyOptions) (*mongo.InsertManyResult, error) {
	start := time.Now()
	_, span := cw.tracer.Start(ctx, "CollectionWrapper.InsertMany", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	res, err := cw.coll.InsertMany(ctx, documents, opts...)
	cw.recordTelemetry(ctx, "InsertMany", start, err, span)
	return res, err
}
