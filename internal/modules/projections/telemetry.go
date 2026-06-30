package projections

import (
	"context"
	"strings"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/JochiRaider/cartulary/internal/platform/telemetry"
)

const (
	timelineViewSchemaID     = "cartulary.view.timeline.v2"
	hostsViewSchemaID        = "cartulary.view.hosts.v1"
	identitiesViewSchemaID   = "cartulary.view.identities.v1"
	indicatorsViewSchemaID   = "cartulary.view.indicators.v1"
	assessmentsViewSchemaID  = "cartulary.view.assessments.v1"
	evidenceViewSchemaID     = "cartulary.view.evidence.v1"
	notesViewSchemaID        = "cartulary.view.notes.v1"
	partiesViewSchemaID      = "cartulary.view.parties.v1"
	taskRequestsViewSchemaID = "cartulary.view.task_requests.v1"
	decisionsViewSchemaID    = "cartulary.view.decisions.v1"
)

func (s *Store) startProjectionSpan(ctx context.Context, viewSchemaID string) (context.Context, func(error)) {
	viewSchemaID = safeProjectionViewSchemaID(viewSchemaID)
	ctx, span := telemetry.Tracer(telemetry.ScopeWorkbook, s.telemetryServiceVersion()).Start(
		ctx,
		"cartulary.workbook.projection",
		trace.WithSpanKind(trace.SpanKindInternal),
		trace.WithAttributes(telemetry.SafeAttributes(
			attribute.String("cartulary.view_schema_id", viewSchemaID),
			attribute.String("cartulary.operation", "rebuild"),
		)...),
	)
	return ctx, func(err error) {
		result := "success"
		if err != nil {
			result = "failed"
			span.SetStatus(codes.Error, "")
		}
		span.SetAttributes(telemetry.SafeAttributes(
			attribute.String("cartulary.view_schema_id", viewSchemaID),
			attribute.String("cartulary.operation", "rebuild"),
			attribute.String("cartulary.result", result),
		)...)
		span.End()
	}
}

func (s *Store) telemetryServiceVersion() string {
	return telemetry.VersionUnknown
}

func safeProjectionViewSchemaID(viewSchemaID string) string {
	switch strings.TrimSpace(viewSchemaID) {
	case timelineViewSchemaID,
		hostsViewSchemaID,
		identitiesViewSchemaID,
		indicatorsViewSchemaID,
		assessmentsViewSchemaID,
		evidenceViewSchemaID,
		notesViewSchemaID,
		partiesViewSchemaID,
		taskRequestsViewSchemaID,
		decisionsViewSchemaID:
		return strings.TrimSpace(viewSchemaID)
	default:
		return "unknown"
	}
}
