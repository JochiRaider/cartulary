package reporting

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/platform/jobs"
)

func TestReportingFailureDiagnosticsBindAdmission(t *testing.T) {
	t.Run("exhaustive bounded stage diagnostics", testRenderIssueCollection)
	payload, err := (&fakeReportingJobStore{}).ReleasePayloadForJob(context.Background(), jobs.Execution{})
	if err != nil {
		t.Fatal(err)
	}
	for _, preview := range []bool{false, true} {
		input := payload.renderInput
		if preview {
			input.ExportModel.PreviewAttemptID, input.ExportModel.ReleaseID = input.ExportModel.ReleaseID, nil
		}
		failure := RenderFailure{Stage: "release_state", FailureCode: "render_timeout", ReasonCode: "render_timeout", SafeDetails: map[string]any{"stage": "release_state", "timeout_kind": "total", "limit_key": "timeout.total_attempt", "limit": 1}}
		raw, err := failedRenderSummary(input, failure)
		if err != nil {
			t.Fatal(err)
		}
		again, err := failedRenderSummary(input, failure)
		if err != nil || !bytes.Equal(raw, again) || validateFailedRenderSummary(raw, input, "render_timeout") != nil {
			t.Fatal("failure bytes changed or were rejected", err)
		}
		for name, change := range map[string]func(map[string]any){
			"identity":                func(m map[string]any) { m["release_id"] = input.SnapshotID },
			"snapshot":                func(m map[string]any) { m["snapshot_id"] = nil },
			"time":                    func(m map[string]any) { m["created_at"] = "2026-01-01T00:00:00.000000Z" },
			"passed":                  func(m map[string]any) { m["result"] = "passed" },
			"missing first failure":   func(m map[string]any) { delete(m, "first_failure") },
			"different first failure": func(m map[string]any) { m["first_failure"].(map[string]any)["reason_code"] = "render_canceled" },
			"empty issues":            func(m map[string]any) { m["issues"] = []any{} },
			"wrong count":             func(m map[string]any) { m["issue_count"] = 2 },
			"unknown failure":         func(m map[string]any) { m["failure_code"] = "private failure" },
			"unknown reason":          func(m map[string]any) { m["reason_code"] = "private reason" },
			"unknown stage":           func(m map[string]any) { m["terminal_stage"] = "private stage" },
			"raw detail": func(m map[string]any) {
				m["issues"].([]any)[0].(map[string]any)["safe_details"].(map[string]any)["raw_error"] = "do-not-echo"
			},
		} {
			t.Run(name, func(t *testing.T) {
				var mutated map[string]any
				if err := json.Unmarshal(raw, &mutated); err != nil {
					t.Fatal(err)
				}
				change(mutated)
				encoded, err := canonicalJSON(mutated)
				if err != nil {
					t.Fatal(err)
				}
				if err := validateFailedRenderSummary(encoded, input, failure.ReasonCode); err == nil || strings.Contains(err.Error(), "do-not-echo") {
					t.Fatal("invalid diagnostic admitted or raw error disclosed", err)
				}
			})
		}
		failure.SafeDetails = map[string]any{"actual": "private content"}
		if _, err := failedRenderSummary(input, failure); err == nil || strings.Contains(err.Error(), "private content") {
			t.Fatal("unbounded diagnostic emitted", err)
		}
	}
}
