package daemon

import (
	"context"
	"log"
	"log/slog"
	nethttp "net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/akaporn-katip/go-project-structure-template/config"
	"github.com/akaporn-katip/go-project-structure-template/internal/application/customer_profile/command"
	"github.com/akaporn-katip/go-project-structure-template/internal/infrastructure/http"
	"github.com/akaporn-katip/go-project-structure-template/internal/infrastructure/http/handler"
	"github.com/akaporn-katip/go-project-structure-template/internal/infrastructure/http/middleware"
	"github.com/akaporn-katip/go-project-structure-template/internal/infrastructure/observability"
	"github.com/akaporn-katip/go-project-structure-template/internal/infrastructure/persistence/postgres"
	"github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"

	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

func Serve(configPath string) {
	cfg, err := config.LoadWithPath(configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// ====================================================================
	// Initialize OpenTelemetry - Tracing
	// ====================================================================
	trace, err := observability.NewTelemetry(observability.TelemetryConfig{
		ServiceName:    cfg.Observability.ServiceName,
		ServiceVersion: cfg.Observability.ServiceVersion,
		Environment:    cfg.Observability.Environment,
		OTLPProtocol:   cfg.Observability.OTLPProtocol,
		OTLPEndpoint:   cfg.Observability.OTLPEndpoint,
		Enabled:        cfg.Observability.EnableTracing,
	})

	if err != nil {
		log.Fatalf("Failed to initialize telemetry: %v", err)
	}

	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := trace.Shutdown(ctx); err != nil {
			log.Printf("Failed to shutdown telemetry: %v", err)
		}
	}()

	// ====================================================================
	// Initialize OpenTelemetry - Metrics
	// ====================================================================
	metrics, err := observability.NewMetrics(observability.MetricsConfig{
		ServiceName:    cfg.Observability.ServiceName,
		ServiceVersion: cfg.Observability.ServiceVersion,
		Environment:    cfg.Observability.Environment,
		OTLPProtocol:   cfg.Observability.OTLPProtocol,
		OTLPEndpoint:   cfg.Observability.OTLPEndpoint,
		Enabled:        cfg.Observability.EnableMetrics,
		ExportInterval: cfg.Observability.MetricsInterval,
	})

	if err != nil {
		log.Fatalf("Failed to initialize metrics: %v", err)
	}

	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := metrics.Shutdown(ctx); err != nil {
			log.Printf("Failed to shutdown metric: %v", err)
		}
	}()

	// ====================================================================
	// Initialize OpenTelemetry - Logging
	// ====================================================================
	logger, err := observability.NewLogger(observability.LoggerConfig{
		ServiceName:    cfg.Observability.ServiceName,
		ServiceVersion: cfg.Observability.ServiceVersion,
		Environment:    cfg.Observability.Environment,
		OTLPProtocol:   cfg.Observability.OTLPProtocol,
		OTLPEndpoint:   cfg.Observability.OTLPEndpoint,
		Enabled:        cfg.Observability.EnableLogging,
		LogLevel:       cfg.Observability.LogLevel,
	})

	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}

	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := logger.Shutdown(ctx); err != nil {
			log.Printf("Failed to shutdown logger: %v", err)
		}
	}()

	// Set as default slog logger
	slog.SetDefault(logger.Logger())
	slog.Info("Logger initialized successfully",
		"service", cfg.Observability.ServiceName,
		"environment", cfg.Observability.Environment,
		"log_level", cfg.Observability.LogLevel,
	)

	// ====================================================================
	// Initialize Database
	// ====================================================================
	// db, err := setupInMemoryDB()

	db, err := sqlx.Connect("postgres", cfg.Database.DSN)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()
	// Create Unit of work
	uow, err := postgres.NewUnitOfWork(db)
	if err != nil {
		log.Fatalf("Failed to create unit of work %v", err)
	}
	// ====================================================================
	// Initialize Application Layer
	// ====================================================================
	createCustomerProfileHandler := command.NewCreateCustomerProfileHandler(uow)

	// ====================================================================
	// Initialize HTTP Layer with Metrics Middleware
	// ====================================================================
	validate := validator.New()
	customerProfileHandler := handler.NewCustomerProfileHandler(createCustomerProfileHandler, validate)

	// Create Middleware
	traceMiddleware := middleware.NewTraceMiddleware(cfg.Observability.ServiceName)
	metricMiddleware, err := middleware.NewMetricsMiddleware(metrics.Meter())
	if err != nil {
		log.Fatalf("Failed to create metrics middleware: %v", err)
	}
	loggingMiddleware := middleware.NewLoggingMiddleware(logger.Logger())

	// Setup router with all middlewares
	r := http.NewRouter(http.Handlers{
		CustomerProfileHandler: customerProfileHandler,
	}, traceMiddleware, metricMiddleware, loggingMiddleware)

	// ====================================================================
	// Start HTTP Server
	// ====================================================================
	srv := &nethttp.Server{
		Addr:    cfg.Server.Port,
		Handler: r.Handler(),
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != nethttp.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// ====================================================================
	// Graceful Shutdown
	// ====================================================================
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutdown Server ...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Println("Server Shutdown:", err)
	}
	log.Println("Server exiting")
}
