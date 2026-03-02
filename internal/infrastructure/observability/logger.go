package observability

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutlog"
	"go.opentelemetry.io/otel/log/global"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.39.0"
)

type LoggerConfig struct {
	OTLPProtocol   string
	ServiceName    string
	ServiceVersion string
	Environment    string
	OTLPEndpoint   string
	Enabled        bool
	LogLevel       string // debug, info, warn, error
}

type Logger struct {
	loggerProvider *sdklog.LoggerProvider
	logger         *slog.Logger
	config         LoggerConfig
}

func NewLogger(config LoggerConfig) (*Logger, error) {
	// Parse log level
	level := parseLogLevel(config.LogLevel)

	if !config.Enabled {
		// Return default slog logger with JSON handler
		handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: level,
		})
		return &Logger{
			logger: slog.New(handler),
			config: config,
		}, nil
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

	exp, err := newLogExporter(ctx, config)
	if err != nil {
		return nil, err
	}

	loggerProvider := newLoggerProvider(res, exp)

	// Set global logger provider
	global.SetLoggerProvider(loggerProvider)

	// Create OpenTelemetry slog handler
	otelHandler := otelslog.NewHandler(
		config.ServiceName,
		otelslog.WithLoggerProvider(loggerProvider),
	)

	// Wrap with JSON handler for local development or combine handlers
	var handler slog.Handler
	if config.OTLPProtocol == "stdout" {
		// Combine JSON handler with OpenTelemetry handler
		jsonHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: level,
		})
		handler = &multiHandler{
			handlers: []slog.Handler{jsonHandler, otelHandler},
		}
	} else {
		// Use only OpenTelemetry handler for remote export
		handler = otelHandler
	}

	logger := slog.New(handler)

	return &Logger{
		loggerProvider: loggerProvider,
		logger:         logger,
		config:         config,
	}, nil
}

func newLoggerProvider(res *resource.Resource, exp sdklog.Exporter) *sdklog.LoggerProvider {
	loggerProvider := sdklog.NewLoggerProvider(
		sdklog.WithResource(res),
		sdklog.WithProcessor(
			sdklog.NewBatchProcessor(exp),
		),
	)

	return loggerProvider
}

func newLogExporter(ctx context.Context, config LoggerConfig) (exp sdklog.Exporter, err error) {
	fmt.Printf("Log OTLPProtocol : %v\n", config.OTLPProtocol)
	switch config.OTLPProtocol {
	case "grpc":
		exp, err = otlploggrpc.New(ctx,
			otlploggrpc.WithEndpoint(config.OTLPEndpoint),
			otlploggrpc.WithInsecure(),
		)
	case "stdout":
		exp, err = stdoutlog.New()
	default:
		exp, err = stdoutlog.New()
	}
	return
}

func parseLogLevel(level string) slog.Level {
	switch level {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// Shutdown gracefully shuts down the logger provider
func (l *Logger) Shutdown(ctx context.Context) error {
	if l.loggerProvider != nil {
		return l.loggerProvider.Shutdown(ctx)
	}
	return nil
}

// Logger returns the configured slog logger
func (l *Logger) Logger() *slog.Logger {
	return l.logger
}

// multiHandler allows multiple handlers to process the same log record
type multiHandler struct {
	handlers []slog.Handler
}

func (m *multiHandler) Enabled(ctx context.Context, level slog.Level) bool {
	for _, h := range m.handlers {
		if h.Enabled(ctx, level) {
			return true
		}
	}
	return false
}

func (m *multiHandler) Handle(ctx context.Context, record slog.Record) error {
	for _, h := range m.handlers {
		if err := h.Handle(ctx, record); err != nil {
			return err
		}
	}
	return nil
}

func (m *multiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	handlers := make([]slog.Handler, len(m.handlers))
	for i, h := range m.handlers {
		handlers[i] = h.WithAttrs(attrs)
	}
	return &multiHandler{handlers: handlers}
}

func (m *multiHandler) WithGroup(name string) slog.Handler {
	handlers := make([]slog.Handler, len(m.handlers))
	for i, h := range m.handlers {
		handlers[i] = h.WithGroup(name)
	}
	return &multiHandler{handlers: handlers}
}
