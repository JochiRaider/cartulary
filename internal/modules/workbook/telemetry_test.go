package workbook

import (
	"net/http"
	"testing"

	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
)

func TestWorkbookTelemetrySafeMappings(t *testing.T) {
	const evidenceViewSchemaID = "cartulary.view.evidence.v1"
	if got := safeWorkbookViewSchemaID(evidenceViewSchemaID); got != evidenceViewSchemaID {
		t.Fatalf("unexpected safe view schema: %q", got)
	}
	if got := safeWorkbookViewSchemaID("incident/10000000"); got != "unknown" {
		t.Fatalf("unexpected unsafe view schema mapping: %q", got)
	}
	if got := safeWorkbookRecordType(evidenceViewSchemaID); got != "evidence" {
		t.Fatalf("unexpected evidence record type: %q", got)
	}
	if got := safeWorkbookOperation("patch"); got != "patch" {
		t.Fatalf("unexpected operation: %q", got)
	}
	if got := safeWorkbookOperation("patch/raw"); got != "unknown" {
		t.Fatalf("unexpected unsafe operation mapping: %q", got)
	}
}

func TestWorkbookAPIErrorTelemetry(t *testing.T) {
	result, code := workbookAPIErrorTelemetry(&httpapi.APIError{Status: http.StatusConflict, Code: "row_version_conflict"})
	if result != "conflict" || code != "row_version_conflict" {
		t.Fatalf("unexpected conflict telemetry: result=%q code=%q", result, code)
	}
	result, code = workbookAPIErrorTelemetry(&httpapi.APIError{Status: http.StatusBadRequest, Code: "invalid_mutation_payload"})
	if result != "rejected" || code != "invalid_mutation_payload" {
		t.Fatalf("unexpected rejection telemetry: result=%q code=%q", result, code)
	}
	result, code = workbookAPIErrorTelemetry(internalAPIError(nil))
	if result != "failed" || code != "internal_error" {
		t.Fatalf("unexpected internal error telemetry: result=%q code=%q", result, code)
	}
}

func TestWorkbookTelemetryNoSDK(t *testing.T) {
	const evidenceViewSchemaID = "cartulary.view.evidence.v1"
	service := &service{serviceVersion: "0.0.0+unknown"}
	ctx, finishQuery := service.startWorkbookQuery(t.Context(), evidenceViewSchemaID)
	finishQuery("success", "", 3)

	_, finishMutation := service.startWorkbookMutation(ctx, evidenceViewSchemaID, "patch")
	finishMutation("conflict", "row_version_conflict")
}
