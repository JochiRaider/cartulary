package jobs

import (
	"testing"
	"time"
)

func TestJobTelemetryVocabularyHelpers(t *testing.T) {
	if got := jobKindFromScope(Scope{Kind: ScopeKindIncident}); got != ScopeKindIncident {
		t.Fatalf("unexpected incident job kind: %q", got)
	}
	if got := jobKindFromScope(Scope{Kind: "unknown-kind"}); got != "unknown" {
		t.Fatalf("unexpected unknown job kind: %q", got)
	}
	for status, want := range map[string]string{
		StatusSucceeded: "success",
		StatusCanceled:  "canceled",
		StatusFailed:    "failed",
		"other":         "failed",
	} {
		if got := resultForTerminalStatus(status); got != want {
			t.Fatalf("terminal status %q result: got %q want %q", status, got, want)
		}
	}
}

func TestSafeJobTelemetryToken(t *testing.T) {
	for _, value := range []string{"invalid_import_request", "internal.error", "object-store"} {
		if !safeJobTelemetryToken(value) {
			t.Fatalf("expected safe token %q", value)
		}
	}
	for _, value := range []string{"", "has space", "sql=select *", "incident/10000000"} {
		if safeJobTelemetryToken(value) {
			t.Fatalf("expected unsafe token %q", value)
		}
	}
}

func TestJobTelemetryNoSDK(t *testing.T) {
	manager := NewManager()
	manager.ConfigureTelemetry("0.0.0+unknown")

	ctx, span := manager.startJobSpan(t.Context(), "cartulary.jobs.enqueue", ScopeKindIncident, "enqueue")
	manager.finishJobSpan(span, "enqueue", ScopeKindIncident, "", "success", nil)

	startedAt := time.Now().UTC()
	finishedAt := startedAt.Add(2 * time.Second)
	manager.recordJobDuration(ctx, Resource{
		Scope:      Scope{Kind: ScopeKindIncident},
		Status:     StatusSucceeded,
		StartedAt:  &startedAt,
		FinishedAt: &finishedAt,
	}, ScopeKindIncident, "success")
}
