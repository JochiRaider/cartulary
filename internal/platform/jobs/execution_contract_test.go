package jobs

import (
	"context"
	"testing"

	"github.com/google/uuid"
)

func TestTypedExecutionContract_Unit(t *testing.T) {
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
