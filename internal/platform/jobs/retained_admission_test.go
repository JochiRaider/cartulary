package jobs

import (
	"encoding/json"
	"github.com/google/uuid"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/platform/jobs/jobcontract"
)

func TestRetainedJobAdmissionResource_Unit(t *testing.T) {
	t.Run("selected terminal contract", checkRetainedTerminalContract)
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

func checkRetainedTerminalContract(t *testing.T) {
	definition := Definition{JobKind: "test.run_v1", ProgressUnitID: "test.run.operation.v1", HandlerName: "test.worker_v1", Terminal: jobcontract.TerminalContract{SchemaID: jobcontract.TerminalSuccessSchema, SuccessCodes: []string{"done"}, Resources: []jobcontract.ResourceContract{{Kind: "incident", SchemaID: jobcontract.ResourceIDSchema, MinRefs: 1, MaxRefs: 1}}}}
	id := "11111111-1111-4111-8111-111111111111"
	retained := RetainedExtensionJob{JobKind: definition.JobKind, ProgressUnitID: definition.ProgressUnitID, HandlerName: definition.HandlerName, Resource: Resource{Status: StatusSucceeded, ResultSummary: &ResultSummary{Code: "done", Message: "Done.", ResourceRefs: []ResourceRef{{Kind: "incident", ID: id, Route: "/api/v1/incidents/" + id}}}}}
	if err := validateRetainedDefinition(retained, definition); err != nil {
		t.Fatal("valid selected result", err)
	}
	for _, name := range []string{"unknown code", "cancellation", "missing resource", "wrong owner", "wrong kind", "wrong progress", "wrong handler"} {
		t.Run(name, func(t *testing.T) {
			changed, summary := retained, *retained.Resource.ResultSummary
			changed.Resource.ResultSummary = &summary
			switch name {
			case "unknown code":
				summary.Code = "unregistered"
			case "cancellation":
				summary.Code = "job_canceled"
			case "missing resource":
				summary.ResourceRefs = nil
			case "wrong owner":
				changed.OwnerProfileID = "unexpected"
			case "wrong kind":
				changed.JobKind = "unexpected"
			case "wrong progress":
				changed.ProgressUnitID = "unexpected"
			case "wrong handler":
				changed.HandlerName = "unexpected"
			}
			if validateRetainedDefinition(changed, definition) == nil {
				t.Fatal("invalid retained definition admitted")
			}
		})
	}
	definition.Terminal.Resources[0].MinRefs = 0
	retained.Resource.ResultSummary.ResourceRefs = nil
	if err := validateRetainedDefinition(retained, definition); err != nil {
		t.Fatal("optional output rejected", err)
	}
	expired := time.Now().UTC()
	retained.ExpiredAt = &expired
	retained.Resource.ResultSummary = nil
	if err := validateRetainedDefinition(retained, definition); err != nil {
		t.Fatal("expired tombstone required erased result", err)
	}
}
