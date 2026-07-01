package workbook

import (
	"errors"
	"net/http"
	"testing"

	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
)

func TestWorkbookTelemetrySafeMappings(t *testing.T) {
	if got := safeWorkbookViewSchemaID(EvidenceViewSchemaID); got != EvidenceViewSchemaID {
		t.Fatalf("unexpected safe view schema: %q", got)
	}
	if got := safeWorkbookViewSchemaID("incident/10000000"); got != "unknown" {
		t.Fatalf("unexpected unsafe view schema mapping: %q", got)
	}
	if got := safeWorkbookRecordType(EvidenceViewSchemaID); got != "evidence" {
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
	result, code = workbookMutationErrorTelemetry(pgx.ErrNoRows, "txn")
	if result != "rejected" || code != "incident_not_found" {
		t.Fatalf("unexpected pgx no rows telemetry: result=%q code=%q", result, code)
	}
	result, code = workbookMutationErrorTelemetry(errors.New("database failure"), "txn")
	if result != "failed" || code != "internal_error" {
		t.Fatalf("unexpected internal error telemetry: result=%q code=%q", result, code)
	}
}

func TestWorkbookTelemetryNoSDK(t *testing.T) {
	service := &Service{serviceVersion: "0.0.0+unknown"}
	ctx, finishQuery := service.startWorkbookQuery(t.Context(), EvidenceViewSchemaID)
	finishQuery("success", "", 3)

	_, finishMutation := service.startWorkbookMutation(ctx, EvidenceViewSchemaID, "patch")
	finishMutation("conflict", "row_version_conflict")
}
