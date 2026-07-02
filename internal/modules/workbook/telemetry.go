package workbook

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"

	"github.com/JochiRaider/cartulary/internal/modules/entities"
	"github.com/JochiRaider/cartulary/internal/modules/indicators"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	"github.com/JochiRaider/cartulary/internal/modules/timeline"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/telemetry"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
)

func (s *Service) startWorkbookQuery(ctx context.Context, viewSchemaID string) (context.Context, func(result string, errorCode string, rowCount int)) {
	viewSchemaID = safeWorkbookViewSchemaID(viewSchemaID)
	started := time.Now()
	ctx, span := telemetry.Tracer(telemetry.ScopeWorkbook, s.telemetryServiceVersion()).Start(
		ctx,
		"cartulary.workbook.query",
		trace.WithSpanKind(trace.SpanKindInternal),
		trace.WithAttributes(telemetry.SafeAttributes(
			attribute.String("cartulary.view_schema_id", viewSchemaID),
			attribute.String("cartulary.operation", "query"),
		)...),
	)
	return ctx, func(result string, errorCode string, rowCount int) {
		result = safeWorkbookResult(result)
		attrs := telemetry.SafeAttributes(
			attribute.String("cartulary.view_schema_id", viewSchemaID),
			attribute.String("cartulary.operation", "query"),
			attribute.String("cartulary.result", result),
		)
		if safeWorkbookToken(errorCode) {
			attrs = telemetry.SafeAttributes(append(attrs, attribute.String("cartulary.error_code", errorCode))...)
		}
		if result != "success" {
			span.SetStatus(codes.Error, "")
		}
		span.SetAttributes(attrs...)
		span.End()

		meter := telemetry.Meter(telemetry.ScopeWorkbook, s.telemetryServiceVersion())
		duration, _ := meter.Float64Histogram(
			"cartulary.workbook.query.duration",
			metric.WithUnit("s"),
			metric.WithDescription("Workbook query duration."),
		)
		duration.Record(ctx, time.Since(started).Seconds(), metric.WithAttributes(attrs...))
		if result == "success" && rowCount >= 0 {
			rowsReturned, _ := meter.Int64Histogram(
				"cartulary.workbook.rows.returned",
				metric.WithUnit("{row}"),
				metric.WithDescription("Serialized rows returned by one successful workbook view-query response."),
			)
			rowsReturned.Record(ctx, int64(rowCount), metric.WithAttributes(
				telemetry.SafeAttributes(
					attribute.String("cartulary.view_schema_id", viewSchemaID),
					attribute.String("cartulary.result", result),
				)...,
			))
		}
	}
}

func (s *Service) startWorkbookMutation(ctx context.Context, viewSchemaID string, operation string) (context.Context, func(result string, errorCode string)) {
	viewSchemaID = safeWorkbookViewSchemaID(viewSchemaID)
	recordType := safeWorkbookRecordType(viewSchemaID)
	operation = safeWorkbookOperation(operation)
	started := time.Now()
	ctx, span := telemetry.Tracer(telemetry.ScopeWorkbook, s.telemetryServiceVersion()).Start(
		ctx,
		"cartulary.workbook.mutation",
		trace.WithSpanKind(trace.SpanKindInternal),
		trace.WithAttributes(telemetry.SafeAttributes(
			attribute.String("cartulary.view_schema_id", viewSchemaID),
			attribute.String("cartulary.record_type", recordType),
			attribute.String("cartulary.operation", operation),
		)...),
	)
	return ctx, func(result string, errorCode string) {
		result = safeWorkbookResult(result)
		attrs := telemetry.SafeAttributes(
			attribute.String("cartulary.view_schema_id", viewSchemaID),
			attribute.String("cartulary.record_type", recordType),
			attribute.String("cartulary.operation", operation),
			attribute.String("cartulary.result", result),
		)
		if safeWorkbookToken(errorCode) {
			attrs = telemetry.SafeAttributes(append(attrs, attribute.String("cartulary.error_code", errorCode))...)
		}
		if result != "success" {
			span.SetStatus(codes.Error, "")
		}
		span.SetAttributes(attrs...)
		span.End()

		duration, _ := telemetry.Meter(telemetry.ScopeWorkbook, s.telemetryServiceVersion()).Float64Histogram(
			"cartulary.workbook.mutation.duration",
			metric.WithUnit("s"),
			metric.WithDescription("Workbook mutation duration."),
		)
		duration.Record(ctx, time.Since(started).Seconds(), metric.WithAttributes(attrs...))
	}
}

func (s *Service) telemetryServiceVersion() string {
	if s != nil && strings.TrimSpace(s.serviceVersion) != "" {
		return s.serviceVersion
	}
	return telemetry.VersionUnknown
}

func workbookAPIErrorTelemetry(apiErr *httpapi.APIError) (string, string) {
	if apiErr == nil {
		return "success", ""
	}
	switch {
	case apiErr.Status == http.StatusConflict:
		return "conflict", apiErr.Code
	case apiErr.Status >= 400 && apiErr.Status < 500:
		return "rejected", apiErr.Code
	case apiErr.Status >= 500:
		return "failed", apiErr.Code
	default:
		return "failed", apiErr.Code
	}
}

func workbookMutationErrorTelemetry(err error, clientTxnID string) (string, string) {
	if err == nil {
		return "success", ""
	}
	return workbookAPIErrorTelemetry(mutationAPIErrorForTelemetry(err, clientTxnID))
}

func mutationAPIErrorForTelemetry(err error, clientTxnID string) *httpapi.APIError {
	var (
		validationErr *MutationValidationError
		lifecycleErr  *LifecycleValidationError
		rowConflict   *RowVersionConflictError
		sameConflict  *SameFieldConflictError
	)
	switch {
	case errors.Is(err, authn.ErrClientTxnConflict):
		return httpapi.ClientTxnConflictError(clientTxnID)
	case errors.Is(err, pgx.ErrNoRows):
		return incidentNotFoundError()
	case errors.Is(err, revisions.ErrRecordDeletedUseRestore):
		return &httpapi.APIError{Status: http.StatusConflict, Code: "record_deleted_use_restore", Details: map[string]any{}}
	case errors.As(err, &validationErr):
		return invalidMutationPayload(validationErr.Field, validationErr.ReasonCode)
	case errors.As(err, &lifecycleErr):
		return &httpapi.APIError{Status: http.StatusConflict, Code: "illegal_transition", Details: map[string]any{}}
	case errors.As(err, &sameConflict):
		return sameFieldConflictError(sameConflict)
	case errors.As(err, &rowConflict):
		return rowVersionConflictError(map[string]any{})
	default:
		return internalAPIError(err)
	}
}

func safeWorkbookViewSchemaID(viewSchemaID string) string {
	if _, ok := viewschema.Lookup(viewSchemaID); ok {
		return viewSchemaID
	}
	switch viewSchemaID {
	case timeline.TimelineViewSchemaID, entities.HostsViewSchemaID, entities.IdentitiesViewSchemaID, indicators.ViewSchemaID:
		return viewSchemaID
	default:
		return "unknown"
	}
}

func safeWorkbookRecordType(viewSchemaID string) string {
	switch viewSchemaID {
	case timeline.TimelineViewSchemaID:
		return "timeline_event"
	case entities.HostsViewSchemaID:
		return "host"
	case entities.IdentitiesViewSchemaID:
		return "identity"
	default:
		if recordType := recordTypeForView(viewSchemaID); safeWorkbookToken(recordType) {
			return recordType
		}
		return "unknown"
	}
}

func safeWorkbookOperation(operation string) string {
	switch operation {
	case "create", "patch", "query":
		return operation
	default:
		return "unknown"
	}
}

func safeWorkbookResult(result string) string {
	switch result {
	case "success", "rejected", "conflict", "canceled", "failed", "timeout", "dropped":
		return result
	default:
		return "failed"
	}
}

func safeWorkbookToken(value string) bool {
	value = strings.TrimSpace(value)
	if value == "" || len(value) > 128 {
		return false
	}
	for _, r := range value {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '.' || r == '_' || r == '-' {
			continue
		}
		return false
	}
	return true
}
