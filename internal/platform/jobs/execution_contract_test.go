package jobs

import (
	"context"
	"testing"

	"github.com/google/uuid"
)

func TestTypedExecutionContract_Unit(t *testing.T) {
	t.Run("fixed cancellation code", func(t *testing.T) {
		for _, code := range []string{"job_canceled", "done", "owner_canceled"} {
			result := ResultSummary{Code: code, Message: "Canceled."}
			_, _, err := marshalSummaries(&result, nil, StatusCanceled)
			if (err == nil) != (code == "job_canceled") {
				t.Fatal("cancellation writer classification", code, err)
			}
			record := Resource{Status: StatusCanceled}
			err = hydrateJobResource(&record, uuid.New(), nil, nil, nil, []byte(`{"code":"`+code+`","message":"Canceled."}`), nil)
			if (err == nil) != (code == "job_canceled") {
				t.Fatal("cancellation reader classification", code, err)
			}
		}
	})
	t.Run("terminal and retained resource admission", func(t *testing.T) {
		invalid := ResultSummary{Code: "done", Message: "Done.", ResourceRefs: []ResourceRef{{Kind: "incident", ID: "arbitrary", Route: "/api/v1/jobs/arbitrary"}}}
		if _, _, err := marshalSummaries(&invalid, nil, StatusSucceeded); err == nil {
			t.Error("terminal writer accepted invalid public resource identity")
		}
		if validRetainedSummary([]byte(`{"code":"done","message":"Done.","resource_refs":[{"kind":"incident","id":"arbitrary","route":"/api/v1/jobs/arbitrary"}]}`), false) {
			t.Error("retained success accepted invalid public resource identity")
		}
	})
	jobID := uuid.MustParse("59000000-0000-4000-8000-000000000001")
	attemptID := uuid.MustParse("59000000-0000-4000-8000-000000000002")
	execution := newExecution(jobID, attemptID)
	if execution.JobID() != jobID {
		t.Fatalf("execution job ID = %s want %s", execution.JobID(), jobID)
	}
	var handler HandlerFunc = func(context.Context, Execution) error { return nil }
	if handler == nil {
		t.Fatal("typed handler is nil")
	}
	_ = SuccessCompletion{Progress: Progress{Completed: 1}, ResultSummary: ResultSummary{Code: "done", Message: "Done."}}
	_ = FailureCompletion{Progress: Progress{Completed: 0}, ErrorSummary: ErrorSummary{Code: "failed", Message: "Failed."}}
	_ = CancellationCompletion{Progress: Progress{Completed: 0}}
}
