package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"

	"github.com/JochiRaider/cartulary/internal/platform/telemetry"
)

type telemetryDB struct {
	inner          DB
	serviceVersion string
	tracer         trace.Tracer
	duration       metric.Float64Histogram
}

func InstrumentDB(inner DB, serviceVersion string) DB {
	if inner == nil {
		return nil
	}
	meter := telemetry.Meter(telemetry.ScopePostgres, serviceVersion)
	duration, _ := meter.Float64Histogram(
		"cartulary.postgres.operation.duration",
		metric.WithUnit("s"),
		metric.WithDescription("Postgres dependency operation duration."),
	)
	return &telemetryDB{
		inner:          inner,
		serviceVersion: serviceVersion,
		tracer:         telemetry.Tracer(telemetry.ScopePostgres, serviceVersion),
		duration:       duration,
	}
}

func (db *telemetryDB) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	ctx, span, started := db.start(ctx, "exec")
	tag, err := db.inner.Exec(ctx, sql, args...)
	db.finish(ctx, span, started, "exec", err)
	return tag, err
}

func (db *telemetryDB) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	ctx, span, started := db.start(ctx, "query")
	rows, err := db.inner.Query(ctx, sql, args...)
	db.finish(ctx, span, started, "query", err)
	return rows, err
}

func (db *telemetryDB) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	ctx, span, started := db.start(ctx, "query_row")
	return telemetryRow{
		inner:     db.inner.QueryRow(ctx, sql, args...),
		db:        db,
		ctx:       ctx,
		span:      span,
		started:   started,
		operation: "query_row",
	}
}

func (db *telemetryDB) BeginTx(ctx context.Context, options pgx.TxOptions) (pgx.Tx, error) {
	ctx, span, started := db.start(ctx, "begin_tx")
	tx, err := db.inner.BeginTx(ctx, options)
	db.finish(ctx, span, started, "begin_tx", err)
	return tx, err
}

func (db *telemetryDB) start(ctx context.Context, operation string) (context.Context, trace.Span, time.Time) {
	started := time.Now()
	ctx, span := db.tracer.Start(
		ctx,
		"cartulary.postgres.operation",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(telemetry.SafeAttributes(
			attribute.String("db.system.name", "postgresql"),
			attribute.String("cartulary.operation", operation),
		)...),
	)
	return ctx, span, started
}

func (db *telemetryDB) finish(ctx context.Context, span trace.Span, started time.Time, operation string, err error) {
	result := "success"
	attrs := telemetry.SafeAttributes(
		attribute.String("db.system.name", "postgresql"),
		attribute.String("cartulary.operation", operation),
	)
	if err != nil {
		result = "failed"
		if errors.Is(err, context.Canceled) {
			result = "canceled"
		}
		attrs = telemetry.SafeAttributes(append(attrs, attribute.String("cartulary.error_class", postgresErrorClass(err)))...)
		span.SetStatus(codes.Error, "")
	}
	attrs = telemetry.SafeAttributes(append(attrs, attribute.String("cartulary.result", result))...)
	span.SetAttributes(attrs...)
	span.End()
	db.duration.Record(ctx, time.Since(started).Seconds(), metric.WithAttributes(attrs...))
}

type telemetryRow struct {
	inner     pgx.Row
	db        *telemetryDB
	ctx       context.Context
	span      trace.Span
	started   time.Time
	operation string
}

func (row telemetryRow) Scan(dest ...any) error {
	err := row.inner.Scan(dest...)
	row.db.finish(row.ctx, row.span, row.started, row.operation, err)
	return err
}

func postgresErrorClass(err error) string {
	if errors.Is(err, context.DeadlineExceeded) {
		return "timeout"
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "40001":
			return "serialization_conflict"
		case "23502", "23503", "23505", "23514":
			return "constraint_violation"
		default:
			return "internal_error"
		}
	}
	return "dependency_unavailable"
}
