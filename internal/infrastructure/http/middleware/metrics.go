package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

type MetricsMiddleware struct {
	requestCounter  metric.Int64Counter
	requestDuration metric.Float64Histogram
	requestSize     metric.Int64Histogram
	responseSize    metric.Int64Histogram
	activeRequests  metric.Int64UpDownCounter
}

func NewMetricsMiddleware(meter metric.Meter) (*MetricsMiddleware, error) {
	requestCounter, err := meter.Int64Counter(
		"http.server.requests",
		metric.WithDescription("Total number of HTTP requests"),
		metric.WithUnit("{request}"),
	)
	if err != nil {
		return nil, err
	}

	// Request duration histogram
	requestDuration, err := meter.Float64Histogram(
		"http.server.duration",
		metric.WithDescription("Duration of HTTP requests"),
		metric.WithUnit("ms"),
	)
	if err != nil {
		return nil, err
	}

	// Request size histogram
	requestSize, err := meter.Int64Histogram(
		"http.server.request.size",
		metric.WithDescription("Size of HTTP request bodies"),
		metric.WithUnit("By"),
	)
	if err != nil {
		return nil, err
	}

	// Response size histogram
	responseSize, err := meter.Int64Histogram(
		"http.server.response.size",
		metric.WithDescription("Size of HTTP response bodies"),
		metric.WithUnit("By"),
	)
	if err != nil {
		return nil, err
	}

	// Active requests gauge
	activeRequests, err := meter.Int64UpDownCounter(
		"http.server.active_requests",
		metric.WithDescription("Number of active HTTP requests"),
		metric.WithUnit("{request}"),
	)
	if err != nil {
		return nil, err
	}

	return &MetricsMiddleware{
		requestCounter:  requestCounter,
		requestDuration: requestDuration,
		requestSize:     requestSize,
		responseSize:    responseSize,
		activeRequests:  activeRequests,
	}, nil

}

func (m *MetricsMiddleware) Handle() gin.HandlerFunc {

	return func(c *gin.Context) {
		start := time.Now()
		r := c.Request
		ctx := r.Context()

		m.activeRequests.Add(ctx, 1, metric.WithAttributes(
			attribute.String("http.method", r.Method),
			attribute.String("http.route", r.URL.Path),
		))

		if r.ContentLength > 0 {
			m.requestSize.Record(ctx, r.ContentLength,
				metric.WithAttributes(
					attribute.String("http.method", r.Method),
					attribute.String("http.route", r.URL.Path),
				),
			)
		}

		c.Next()
		rw := c.Writer

		duration := time.Since(start).Milliseconds()

		attrs := []attribute.KeyValue{
			attribute.String("http.method", r.Method),
			attribute.String("http.route", r.URL.Path),
			attribute.Int("http.status_code", rw.Status()),
			attribute.String("http.status_class", statusClass(rw.Status())),
		}

		m.requestCounter.Add(ctx, 1, metric.WithAttributes(attrs...))
		m.requestDuration.Record(ctx, float64(duration), metric.WithAttributes(attrs...))
		m.responseSize.Record(ctx, int64(rw.Size()), metric.WithAttributes(attrs...))

		m.activeRequests.Add(ctx, -1,
			metric.WithAttributes(
				attribute.String("http.method", r.Method),
				attribute.String("http.route", r.URL.Path),
			),
		)
	}
}

func statusClass(statusCode int) string {
	switch {
	case statusCode >= 200 && statusCode < 300:
		return "2xx"
	case statusCode >= 300 && statusCode < 400:
		return "3xx"
	case statusCode >= 400 && statusCode < 500:
		return "4xx"
	case statusCode >= 500:
		return "5xx"
	default:
		return "unknown"
	}
}
