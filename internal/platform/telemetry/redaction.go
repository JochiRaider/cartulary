package telemetry

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

const (
	TelemetryItemDroppedMetricName = "cartulary.telemetry.item.dropped"
	RedactionRejectedDropReason    = "redaction_rejected"
)

type ForbiddenValueFamily string

const (
	ForbiddenIncidentAuthoredContent ForbiddenValueFamily = "incident_authored_content"
	ForbiddenEvidenceIdentity        ForbiddenValueFamily = "evidence_identity"
	ForbiddenStableIdentifier        ForbiddenValueFamily = "stable_identifier"
	ForbiddenSecurityMaterial        ForbiddenValueFamily = "security_material"
	ForbiddenRequestContent          ForbiddenValueFamily = "request_content"
	ForbiddenDatabaseContent         ForbiddenValueFamily = "database_content"
	ForbiddenInfrastructureIdentity  ForbiddenValueFamily = "infrastructure_identity"
	ForbiddenExceptionDetail         ForbiddenValueFamily = "exception_detail"
	ForbiddenBaggageTraceState       ForbiddenValueFamily = "baggage_trace_state"
)

type ForbiddenValueTreatment string

const (
	ForbiddenTreatmentOmit               ForbiddenValueTreatment = "omit"
	ForbiddenTreatmentReplaceClosedClass ForbiddenValueTreatment = "replace_with_closed_class"
	ForbiddenTreatmentDropItem           ForbiddenValueTreatment = "drop_item"
)

type ForbiddenValueAction struct {
	Family             ForbiddenValueFamily
	DefaultTreatment   ForbiddenValueTreatment
	ReplacementAllowed bool
	DropMetricReason   string
	DiagnosticFamily   string
}

var forbiddenValueActions = []ForbiddenValueAction{
	{Family: ForbiddenIncidentAuthoredContent, DefaultTreatment: ForbiddenTreatmentOmit, ReplacementAllowed: true, DropMetricReason: RedactionRejectedDropReason, DiagnosticFamily: "incident_authored_content"},
	{Family: ForbiddenEvidenceIdentity, DefaultTreatment: ForbiddenTreatmentOmit, ReplacementAllowed: true, DropMetricReason: RedactionRejectedDropReason, DiagnosticFamily: "evidence_identity"},
	{Family: ForbiddenStableIdentifier, DefaultTreatment: ForbiddenTreatmentOmit, ReplacementAllowed: true, DropMetricReason: RedactionRejectedDropReason, DiagnosticFamily: "stable_identifier"},
	{Family: ForbiddenSecurityMaterial, DefaultTreatment: ForbiddenTreatmentOmit, ReplacementAllowed: false, DropMetricReason: RedactionRejectedDropReason, DiagnosticFamily: "security_material"},
	{Family: ForbiddenRequestContent, DefaultTreatment: ForbiddenTreatmentOmit, ReplacementAllowed: true, DropMetricReason: RedactionRejectedDropReason, DiagnosticFamily: "request_content"},
	{Family: ForbiddenDatabaseContent, DefaultTreatment: ForbiddenTreatmentOmit, ReplacementAllowed: true, DropMetricReason: RedactionRejectedDropReason, DiagnosticFamily: "database_content"},
	{Family: ForbiddenInfrastructureIdentity, DefaultTreatment: ForbiddenTreatmentOmit, ReplacementAllowed: true, DropMetricReason: RedactionRejectedDropReason, DiagnosticFamily: "infrastructure_identity"},
	{Family: ForbiddenExceptionDetail, DefaultTreatment: ForbiddenTreatmentReplaceClosedClass, ReplacementAllowed: true, DropMetricReason: RedactionRejectedDropReason, DiagnosticFamily: "exception_detail"},
	{Family: ForbiddenBaggageTraceState, DefaultTreatment: ForbiddenTreatmentOmit, ReplacementAllowed: false, DropMetricReason: RedactionRejectedDropReason, DiagnosticFamily: "baggage_trace_state"},
}

func ForbiddenValueActions() []ForbiddenValueAction {
	return append([]ForbiddenValueAction(nil), forbiddenValueActions...)
}

func ActionForForbiddenValueFamily(family ForbiddenValueFamily) (ForbiddenValueAction, bool) {
	for _, action := range forbiddenValueActions {
		if action.Family == family {
			return action, true
		}
	}
	return ForbiddenValueAction{}, false
}

func RedactionDropMetric(family ForbiddenValueFamily, recursive bool) (string, []attribute.KeyValue, bool) {
	action, ok := ActionForForbiddenValueFamily(family)
	if !ok || recursive || action.DropMetricReason == "" {
		return "", nil, false
	}
	attrs := SafeAttributes(attribute.String("cartulary.drop_reason", action.DropMetricReason))
	if len(attrs) != 1 {
		return "", nil, false
	}
	return TelemetryItemDroppedMetricName, attrs, true
}

func RecordRedactionDrop(ctx context.Context, family ForbiddenValueFamily, serviceVersion string, recursive bool) bool {
	name, attrs, ok := RedactionDropMetric(family, recursive)
	if !ok {
		return false
	}
	counter, err := Meter(ScopeTelemetry, serviceVersion).Int64Counter(
		name,
		metric.WithUnit("{item}"),
		metric.WithDescription("Telemetry items dropped before recording because redaction could not prove safety."),
	)
	if err != nil {
		return false
	}
	counter.Add(ctx, 1, metric.WithAttributes(attrs...))
	return true
}
