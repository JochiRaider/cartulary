package telemetry

import (
	"context"
	"testing"

	"go.opentelemetry.io/otel/attribute"
)

func TestSafeAttributesPreservesAdoptedLowCardinalityAttributes(t *testing.T) {
	attrs := SafeAttributes(
		attribute.String("http.request.method", "GET"),
		attribute.String("http.route", "/api/v1/incidents/{incident_route}"),
		attribute.Int("http.response.status_code", 200),
		attribute.String("db.system.name", "postgresql"),
		attribute.String("cartulary.module", "workbook"),
		attribute.String("cartulary.operation", "query"),
		attribute.String("cartulary.result", "success"),
		attribute.String("cartulary.error_class", "request_invalid"),
		attribute.String("cartulary.profile.claims", "base,import"),
		attribute.String("cartulary.incident.hash64", "0123456789abcdef"),
	)

	if len(attrs) != 10 {
		t.Fatalf("expected all adopted attributes to be preserved, got %#v", attrs)
	}
}

func TestSafeAttributesOmitsForbiddenValueFamiliesBeforeRecording(t *testing.T) {
	attrs := SafeAttributes(
		attribute.String("http.route", "/api/v1/incidents/10000000-0000-4000-8000-000000000001?include=rows"),
		attribute.String("cartulary.error_code", "row_version_conflict:10000000-0000-4000-8000-000000000001"),
		attribute.String("cartulary.operation", "select * from incidents where id=$1"),
		attribute.String("db.statement", "select * from incidents"),
		attribute.String("cartulary.object_key", "incidents/secret/file.txt"),
		attribute.String("cartulary.result", "success"),
		attribute.String("cartulary.drop_reason", "queue_full"),
	)

	got := attributesByName(attrs)
	for _, forbidden := range []string{"http.route", "cartulary.error_code", "cartulary.operation", "db.statement", "cartulary.object_key"} {
		if _, ok := got[forbidden]; ok {
			t.Fatalf("forbidden attribute %s was preserved in %#v", forbidden, got)
		}
	}
	if got["cartulary.result"] != "success" || got["cartulary.drop_reason"] != "queue_full" {
		t.Fatalf("safe attributes were not preserved: %#v", got)
	}
}

func TestSafeAttributesOmitNullEquivalentAndUnknownCartularyKeys(t *testing.T) {
	attrs := SafeAttributes(
		attribute.String("cartulary.result", ""),
		attribute.String("cartulary.raw_user_id", "user-1"),
		attribute.String("cartulary.result", "null"),
		attribute.String("cartulary.result", "failed"),
	)

	got := attributesByName(attrs)
	if len(got) != 1 || got["cartulary.result"] != "failed" {
		t.Fatalf("expected only valid result attribute, got %#v", got)
	}
}

func TestNullLikeValuesOmittedBeforeRecordingForAllSignalFamilies(t *testing.T) {
	signalCases := map[string][]attribute.KeyValue{
		"spans": {
			attribute.String("cartulary.result", ""),
			attribute.String("cartulary.operation", "null"),
			attribute.String("cartulary.result", "success"),
		},
		"metrics": {
			attribute.String("cartulary.signal_kind", ""),
			attribute.String("cartulary.drop_reason", "null"),
			attribute.String("cartulary.signal_kind", "metrics"),
		},
		"logs": {
			attribute.String("cartulary.error_class", ""),
			attribute.String("cartulary.result", "null"),
			attribute.String("cartulary.error_class", "internal_error"),
		},
		"resource": {
			attribute.String("deployment.environment.name", ""),
			attribute.String("cartulary.profile.claims", "null"),
			attribute.String("cartulary.profile.claims", "base"),
		},
	}
	for signal, attrs := range signalCases {
		t.Run(signal, func(t *testing.T) {
			got := attributesByName(SafeAttributes(attrs...))
			if len(got) != 1 {
				t.Fatalf("expected only one safe attribute for %s, got %#v", signal, got)
			}
			for _, value := range got {
				if value == "" || value == "null" {
					t.Fatalf("null-like value reached %s telemetry attributes: %#v", signal, got)
				}
			}
		})
	}

	if _, _, ok := dropMetric("", QueueFullDropReason, false); ok {
		t.Fatal("self-diagnostic/drop metric must not record null signal_kind")
	}
	if result := PlanSelfDiagnostic(SelfDiagnosticPlan{Recursive: true, MetricAllowed: true}); result.MetricName != "" || len(result.Attributes) != 0 {
		t.Fatalf("self-diagnostics must omit null-like signal attributes before recording: %#v", result)
	}
}

func TestForbiddenValueActionsAreClosed(t *testing.T) {
	want := []struct {
		family      ForbiddenValueFamily
		treatment   ForbiddenValueTreatment
		replacement bool
		diagnostic  string
	}{
		{ForbiddenIncidentAuthoredContent, ForbiddenTreatmentOmit, true, "incident_authored_content"},
		{ForbiddenEvidenceIdentity, ForbiddenTreatmentOmit, true, "evidence_identity"},
		{ForbiddenStableIdentifier, ForbiddenTreatmentOmit, true, "stable_identifier"},
		{ForbiddenSecurityMaterial, ForbiddenTreatmentOmit, false, "security_material"},
		{ForbiddenRequestContent, ForbiddenTreatmentOmit, true, "request_content"},
		{ForbiddenDatabaseContent, ForbiddenTreatmentOmit, true, "database_content"},
		{ForbiddenInfrastructureIdentity, ForbiddenTreatmentOmit, true, "infrastructure_identity"},
		{ForbiddenExceptionDetail, ForbiddenTreatmentReplaceClosedClass, true, "exception_detail"},
		{ForbiddenBaggageTraceState, ForbiddenTreatmentOmit, false, "baggage_trace_state"},
	}

	actions := ForbiddenValueActions()
	if len(actions) != len(want) {
		t.Fatalf("forbidden-value action count mismatch: got %d want %d", len(actions), len(want))
	}
	for i, expected := range want {
		action := actions[i]
		if action.Family != expected.family ||
			action.DefaultTreatment != expected.treatment ||
			action.ReplacementAllowed != expected.replacement ||
			action.DropMetricReason != RedactionRejectedDropReason ||
			action.DiagnosticFamily != expected.diagnostic {
			t.Fatalf("action %d mismatch: got %#v want family=%s treatment=%s replacement=%t diagnostic=%s",
				i, action, expected.family, expected.treatment, expected.replacement, expected.diagnostic)
		}
		lookedUp, ok := ActionForForbiddenValueFamily(expected.family)
		if !ok || lookedUp != action {
			t.Fatalf("lookup mismatch for %s: got %#v ok=%t", expected.family, lookedUp, ok)
		}
	}

	actions[0].DiagnosticFamily = "mutated"
	if got := ForbiddenValueActions()[0].DiagnosticFamily; got != "incident_authored_content" {
		t.Fatalf("action registry returned mutable backing storage: got %q", got)
	}
	if _, ok := ActionForForbiddenValueFamily(ForbiddenValueFamily("unknown")); ok {
		t.Fatal("unknown forbidden-value family must not resolve to an action")
	}
}

func TestRedactionDropMetricRecordsOnlyWhenNonRecursive(t *testing.T) {
	for _, action := range ForbiddenValueActions() {
		name, attrs, ok := RedactionDropMetric(action.Family, false)
		if !ok {
			t.Fatalf("expected non-recursive drop metric for %s", action.Family)
		}
		if name != TelemetryItemDroppedMetricName {
			t.Fatalf("drop metric name mismatch: got %q", name)
		}
		got := attributesByName(attrs)
		if len(got) != 1 || got["cartulary.drop_reason"] != RedactionRejectedDropReason {
			t.Fatalf("drop metric attributes must contain only redaction_rejected reason, got %#v", got)
		}
		if _, _, ok := RedactionDropMetric(action.Family, true); ok {
			t.Fatalf("recursive redaction drop must not record a metric for %s", action.Family)
		}
		if !RecordRedactionDrop(context.Background(), action.Family, "", false) {
			t.Fatalf("non-recursive redaction drop should be recorded for %s", action.Family)
		}
		if RecordRedactionDrop(context.Background(), action.Family, "", true) {
			t.Fatalf("recursive redaction drop should not be recorded for %s", action.Family)
		}
	}
	if _, _, ok := RedactionDropMetric(ForbiddenValueFamily("unknown"), false); ok {
		t.Fatal("unknown forbidden family must not produce a drop metric")
	}
}

func TestResourceIdentityFailsWhenResourceAttributeLeavesRegistry(t *testing.T) {
	cfg := validTelemetryBootstrapConfig(t)
	cfg.Telemetry.Resource.DeploymentEnvironmentName = "/var/lib/cartulary"
	if _, err := BuildResourceIdentity(cfg, nil); err == nil {
		t.Fatal("expected unsafe deployment environment resource value to fail")
	}
}

func attributesByName(attrs []attribute.KeyValue) map[string]string {
	result := make(map[string]string, len(attrs))
	for _, attr := range attrs {
		result[string(attr.Key)] = attr.Value.Emit()
	}
	return result
}
