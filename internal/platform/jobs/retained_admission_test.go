package jobs

import (
	"encoding/json"
	"github.com/google/uuid"
	"testing"
	"time"
)

func TestRetainedJobAdmissionResource_Unit(t *testing.T) {
	now := time.Date(2026, 9, 9, 12, 0, 0, 0, time.UTC)
	retention := now.Add(24 * time.Hour)
	incident := uuid.New()
	resource := RetainedExtensionJob{Resource: Resource{Scope: Scope{Kind: ScopeKindIncident, IncidentID: &incident}, Status: StatusFailed, SubmittedAt: now, UpdatedAt: now, FinishedAt: &now, RetainedUntil: &retention}}
	validFailure := []byte(`{"code":"job_failed","message":"Failed.","retryable":false}`)
	if err := validateRetainedAdmissionResource(resource, nil, validFailure); err != nil {
		t.Fatal(err)
	}
	for _, raw := range []string{`{}`, `{"code":"job_failed","message":"Failed.","retryable":"false"}`, `{"code":"job_failed","message":"Failed.","retryable":false,"details":null}`, `{"code":"job_failed","message":"Failed.","retryable":false,"private":"unsupported"}`} {
		if validateRetainedAdmissionResource(resource, nil, []byte(raw)) == nil {
			t.Fatal("malformed retained failure accepted")
		}
	}
	resource.Resource.Status = StatusCanceled
	resource.Resource.ResultSummary = &ResultSummary{Code: "job_canceled", Message: "Canceled."}
	raw, _ := json.Marshal(resource.Resource.ResultSummary)
	if err := validateRetainedAdmissionResource(resource, raw, nil); err != nil {
		t.Fatal(err)
	}
	resource.Resource.ResultSummary.Code = "unsupported"
	raw, _ = json.Marshal(resource.Resource.ResultSummary)
	if validateRetainedAdmissionResource(resource, raw, nil) == nil {
		t.Fatal("unknown cancellation code accepted")
	}
	resource.ExpiredAt = &retention
	resource.Resource.ResultSummary = nil
	if err := validateRetainedAdmissionResource(resource, nil, nil); err != nil {
		t.Fatal(err)
	}
	if validateRetainedAdmissionResource(resource, []byte(`{}`), nil) == nil {
		t.Fatal("expired job restored an erased summary")
	}
	resource.Resource.Status = StatusRunning
	if validateRetainedAdmissionResource(resource, nil, nil) == nil {
		t.Fatal("nonterminal expired state accepted")
	}
}
