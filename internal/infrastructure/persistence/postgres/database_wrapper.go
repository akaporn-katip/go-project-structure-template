package postgres

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/noop"
	"go.opentelemetry.io/otel/trace"
)

type DatabaseWrapper struct {
	DatabaseExecutor
	db     DatabaseExecutor
	tracer trace.Tracer

	queryCounter  metric.Int64Counter
	queryDuration metric.Float64Histogram
	queryErrors   metric.Int64Counter
}

func NewDatabaseWrapper(db DatabaseExecutor) *DatabaseWrapper {
	wrapper := &DatabaseWrapper{
		db:     db,
		tracer: otel.Tracer("api.katipwork.com/crm/internal/infrastructure/persistence/postgres/database_wrapper"),
	}

	meter := otel.Meter("api.katipwork.com/crm/internal/infrastructure/persistence/postgres/unit_of_work")
	wrapper.initMetrics(meter)
	return wrapper
}

func (r *DatabaseWrapper) initMetrics(meter metric.Meter) error {
	var err error
	r.queryCounter, err = meter.Int64Counter(
		"db.query.total",
		metric.WithDescription("Total number of database queries"),
		metric.WithUnit("{query}"),
	)
	if err != nil {
		r.queryCounter = noop.Int64Counter{}
	}

	r.queryDuration, err = meter.Float64Histogram(
		"db.query.duration",
		metric.WithDescription("Duration of database queries"),
		metric.WithUnit("ms"),
	)
	if err != nil {
		r.queryDuration = noop.Float64Histogram{}
	}

	r.queryErrors, err = meter.Int64Counter(
		"db.query.errors",
		metric.WithDescription("Number of database query errors"),
		metric.WithUnit("{error}"),
	)
	if err != nil {
		r.queryErrors = noop.Int64Counter{}
	}

	return nil
}

func (pw *DatabaseWrapper) SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	start := time.Now()
	_, span := pw.tracer.Start(ctx, "DatabaseWrapper.SelectContext",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			attribute.String("db.statement", query),
		),
	)
	defer span.End()
	err := pw.db.SelectContext(ctx, dest, query, args...)
	operation, table := extractSQLOperationAndTable(query)
	duration := time.Since(start).Milliseconds()
	attrs := []attribute.KeyValue{
		attribute.String("db.system", "postgresql"),
		attribute.String("db.operation", operation),
		attribute.String("db.table", table),
	}

	pw.queryCounter.Add(ctx, 1, metric.WithAttributes(attrs...))
	pw.queryDuration.Record(ctx, float64(duration), metric.WithAttributes(attrs...))

	if err != nil {
		pw.queryErrors.Add(ctx, 1, metric.WithAttributes(
			append(attrs, attribute.String("error.type", "select_error"))...,
		))
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	span.SetStatus(codes.Ok, "SelectContext success")
	return nil
}

func (pw *DatabaseWrapper) GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	start := time.Now()

	_, span := pw.tracer.Start(ctx, "DatabaseWrapper.GetContext",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			attribute.String("db.statement", query),
		),
	)
	defer span.End()
	err := pw.db.GetContext(ctx, dest, query, args...)
	operation, table := extractSQLOperationAndTable(query)
	duration := time.Since(start).Milliseconds()
	attrs := []attribute.KeyValue{
		attribute.String("db.system", "postgresql"),
		attribute.String("db.operation", operation),
		attribute.String("db.table", table),
	}

	pw.queryCounter.Add(ctx, 1, metric.WithAttributes(attrs...))
	pw.queryDuration.Record(ctx, float64(duration), metric.WithAttributes(attrs...))

	if err != nil {
		pw.queryErrors.Add(ctx, 1, metric.WithAttributes(
			append(attrs, attribute.String("error.type", "get_error"))...,
		))

		if err.Error() != sql.ErrNoRows.Error() {
			span.SetStatus(codes.Error, err.Error())
		}
		return err
	}

	span.SetStatus(codes.Ok, "GetContext success")
	return nil
}

func (pw *DatabaseWrapper) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	start := time.Now()

	_, span := pw.tracer.Start(ctx, "DatabaseWrapper.ExecContext",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			attribute.String("db.statement", query),
		),
	)
	defer span.End()

	rs, err := pw.db.ExecContext(ctx, query, args...)
	duration := time.Since(start).Milliseconds()
	operation, table := extractSQLOperationAndTable(query)
	attrs := []attribute.KeyValue{
		attribute.String("db.system", "postgresql"),
		attribute.String("db.operation", operation),
		attribute.String("db.table", table),
	}

	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		pw.queryErrors.Add(ctx, 1, metric.WithAttributes(
			append(attrs, attribute.String("error.type", "execute_error"))...,
		))
		return rs, err
	}

	pw.queryCounter.Add(ctx, 1, metric.WithAttributes(attrs...))
	pw.queryDuration.Record(ctx, float64(duration), metric.WithAttributes(attrs...))

	span.SetStatus(codes.Ok, "ExecContext success")
	return rs, nil
}

func (pw *DatabaseWrapper) BindNamed(query string, arg interface{}) (string, []interface{}, error) {
	return pw.db.BindNamed(query, arg)
}

func extractSQLOperationAndTable(query string) (operation, table string) {
	q := strings.TrimSpace(query)
	if q == "" {
		return "", ""
	}
	uq := strings.ToUpper(q)

	// find first core operation token occurrence (prefer earliest in the text)
	ops := []string{"SELECT", "INSERT", "UPDATE", "DELETE"}
	opIdx := -1
	for _, o := range ops {
		i := strings.Index(uq, o)
		if i >= 0 && (opIdx == -1 || i < opIdx) {
			opIdx = i
			operation = o
		}
	}
	if operation == "" {
		return "", ""
	}

	// choose the keyword that precedes a table name for the given operation
	var key string
	switch operation {
	case "SELECT", "DELETE":
		key = "FROM"
	case "INSERT":
		key = "INTO"
	case "UPDATE":
		key = "UPDATE"
	default:
		return operation, ""
	}

	// search for the keyword starting from the operation occurrence
	searchBase := uq[opIdx:]
	keyPos := strings.Index(searchBase, key)
	if keyPos < 0 {
		return operation, ""
	}
	// absolute position of token start in original query
	start := opIdx + keyPos + len(key)

	// skip whitespace
	for start < len(q) && (q[start] == ' ' || q[start] == '\t' || q[start] == '\n' || q[start] == '\r') {
		start++
	}
	if start >= len(q) {
		return operation, ""
	}

	// if next char is an opening parenthesis it's likely a subquery - no table to extract
	switch q[start] {
	case '(':
		return operation, ""
	case '"', '\'', '`':
		// quoted identifier
		quote := q[start]
		end := start + 1
		for end < len(q) && q[end] != quote {
			end++
		}
		if end >= len(q) {
			table = strings.TrimSpace(q[start+1:])
		} else {
			table = q[start+1 : end]
		}
	default:
		// unquoted identifier: read until whitespace, comma, semicolon, or parenthesis
		end := start
		for end < len(q) {
			ch := q[end]
			if ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r' || ch == ',' || ch == ';' || ch == '(' || ch == ')' {
				break
			}
			end++
		}
		table = q[start:end]
	}

	table = strings.TrimSpace(table)
	// strip optional "AS" alias or plain alias (we stopped at space so alias shouldn't be included)
	// remove surrounding quotes/backticks if any remain
	table = strings.Trim(table, `"'`+"`")
	return operation, table
}
