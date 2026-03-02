package observability

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.39.0"

	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"

	"go.opentelemetry.io/otel/trace"
)

type TelemetryConfig struct {
	OTLPProtocol   string
	ServiceName    string
	ServiceVersion string
	Environment    string
	OTLPEndpoint   string
	Enabled        bool
}

type Telemetry struct {
	tracerProvider *sdktrace.TracerProvider
	tracer         trace.Tracer
	config         TelemetryConfig
}

func NewTelemetry(config TelemetryConfig) (*Telemetry, error) {
	if !config.Enabled {
		return &Telemetry{
			tracer: otel.Tracer(config.ServiceName),
			config: config,
		}, nil
	}

	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(config.ServiceName),
			semconv.ServiceVersion(config.ServiceVersion),
			attribute.String("environment", config.Environment),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	exp, err := newTraceExporter(ctx, config)
	if err != nil {
		return nil, err
	}

	tracerProvider, err := newTraceProvider(res, exp)

	// Set global tracer provider
	otel.SetTracerProvider(tracerProvider)

	// Set global propagator for distributed tracing
	otel.SetTextMapPropagator(
		newPropagator(),
	)

	tracer := tracerProvider.Tracer(
		config.ServiceName,
		trace.WithInstrumentationVersion(config.ServiceVersion),
	)

	return &Telemetry{
		tracerProvider: tracerProvider,
		tracer:         tracer,
		config:         config,
	}, nil
}

func newTraceProvider(res *resource.Resource, exp sdktrace.SpanExporter) (*sdktrace.TracerProvider, error) {
	traceProvider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(
			sdktrace.ParentBased(sdktrace.AlwaysSample()),
		),
	)

	return traceProvider, nil
}

func newTraceExporter(ctx context.Context, config TelemetryConfig) (exp sdktrace.SpanExporter, err error) {
	fmt.Printf("OTLPProtocol : %v\n", config.OTLPProtocol)
	switch config.OTLPProtocol {
	case "grpc":
		exp, err = otlptracegrpc.New(ctx,
			otlptracegrpc.WithEndpoint(config.OTLPEndpoint),
			otlptracegrpc.WithInsecure(),
		)
	case "stdout":
		exp, err = stdouttrace.New()
	default:
		exp, err = stdouttrace.New()
	}
	return
}

func newPropagator() propagation.TextMapPropagator {
	return propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
}

// Shutdown gracefully shuts down the tracer provider
func (t *Telemetry) Shutdown(ctx context.Context) error {
	if t.tracerProvider != nil {
		return t.tracerProvider.Shutdown(ctx)
	}
	return nil
}

// Tracer returns the configured tracer
func (t *Telemetry) Tracer() trace.Tracer {
	return t.tracer
}

// StartSpan starts a new span with the given name
func (t *Telemetry) StartSpan(ctx context.Context, spanName string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	return t.tracer.Start(ctx, spanName, opts...)
}

// AddEvent adds an event to the current span
func AddEvent(ctx context.Context, name string, attrs ...attribute.KeyValue) {
	span := trace.SpanFromContext(ctx)
	span.AddEvent(name, trace.WithAttributes(attrs...))
}

// SetAttributes adds attributes to the current span
func SetAttributes(ctx context.Context, attrs ...attribute.KeyValue) {
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(attrs...)
}

// RecordError records an error in the current span
func RecordError(ctx context.Context, err error, attrs ...attribute.KeyValue) {
	span := trace.SpanFromContext(ctx)
	span.RecordError(err, trace.WithAttributes(attrs...))
}

// GetTraceID returns the trace ID from context
func GetTraceID(ctx context.Context) string {
	span := trace.SpanFromContext(ctx)
	return span.SpanContext().TraceID().String()
}

// GetSpanID returns the span ID from context
func GetSpanID(ctx context.Context) string {
	span := trace.SpanFromContext(ctx)
	return span.SpanContext().SpanID().String()
}
