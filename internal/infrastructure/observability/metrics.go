package observability

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.39.0"
)

type MetricsConfig struct {
	OTLPProtocol   string
	ServiceName    string
	ServiceVersion string
	Environment    string
	OTLPEndpoint   string
	Enabled        bool
	ExportInterval time.Duration // How often to export metrics (default: 60s)
}

type Metrics struct {
	meterProvider *sdkmetric.MeterProvider
	meter         metric.Meter
	config        MetricsConfig
}

func NewMetrics(config MetricsConfig) (*Metrics, error) {
	if !config.Enabled {
		return &Metrics{
			meter:  otel.Meter(config.ServiceName),
			config: config,
		}, nil
	}

	if config.ExportInterval == 0 {
		config.ExportInterval = 60 * time.Second
	}

	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(config.ServiceName),
			semconv.ServiceVersion(config.ServiceVersion),
			semconv.DeploymentEnvironmentName(config.Environment),
		),
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	exp, err := newMeterExporter(ctx, config)

	meterProvider := newMeterProvider(res, exp, config.ExportInterval)
	otel.SetMeterProvider(meterProvider)

	meter := meterProvider.Meter(
		config.ServiceName,
		metric.WithInstrumentationVersion(config.ServiceVersion),
	)
	return &Metrics{
		meterProvider: meterProvider,
		meter:         meter,
		config:        config,
	}, nil
}

func newMeterProvider(res *resource.Resource, exp sdkmetric.Exporter, exportInterval time.Duration) *sdkmetric.MeterProvider {
	meterProvider := sdkmetric.NewMeterProvider(
		sdkmetric.WithResource(res),
		sdkmetric.WithReader(
			sdkmetric.NewPeriodicReader(
				exp,
				sdkmetric.WithInterval(exportInterval),
			),
		),
	)

	return meterProvider
}

func newMeterExporter(ctx context.Context, config MetricsConfig) (exp sdkmetric.Exporter, err error) {
	fmt.Printf("OTLPProtocol : %v\n", config.OTLPProtocol)
	switch config.OTLPProtocol {
	case "grpc":
		exp, err = otlpmetricgrpc.New(ctx,
			otlpmetricgrpc.WithEndpoint(config.OTLPEndpoint),
			otlpmetricgrpc.WithInsecure(),
		)
	case "stdout":
		exp, err = stdoutmetric.New()
	default:
		exp, err = stdoutmetric.New()
	}
	return
}

func (m *Metrics) Shutdown(ctx context.Context) error {
	if m.meterProvider != nil {
		return m.meterProvider.Shutdown(ctx)
	}
	return nil
}

func (m *Metrics) Meter() metric.Meter {
	return m.meter
}

func (m *Metrics) CreateCounter(name, description, unit string) (metric.Int64Counter, error) {
	return m.meter.Int64Counter(
		name,
		metric.WithDescription(description),
		metric.WithUnit(unit),
	)
}

func (m *Metrics) CreateHistogram(name, description, unit string) (metric.Float64Histogram, error) {
	return m.meter.Float64Histogram(
		name,
		metric.WithDescription(description),
		metric.WithUnit(unit),
	)
}

func (m *Metrics) CreateUpDownCounter(name, description, unit string) (metric.Int64UpDownCounter, error) {
	return m.meter.Int64UpDownCounter(
		name,
		metric.WithDescription(description),
		metric.WithUnit(unit),
	)
}

func (m *Metrics) CreateGauge(name, description, unit string, callback metric.Int64Callback) error {
	_, err := m.meter.Int64ObservableGauge(
		name,
		metric.WithDescription(description),
		metric.WithUnit(unit),
		metric.WithInt64Callback(callback),
	)
	return err
}
