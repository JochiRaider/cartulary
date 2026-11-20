package jobs

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestTypedExecutionContract_Unit(t *testing.T) {
	t.Run("issued owner commit outlives attempt cancellation within the bound", assertAwaitCommit)
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

// assertAwaitCommit pins AwaitCommit's contract: the commit never sees the
// attempt's cancellation directly, and is cancelled only once the bound after
// that cancellation has elapsed. The commit model loses its acknowledgement
// whenever its own context ends, as the PostgreSQL driver does.
func assertAwaitCommit(t *testing.T) {
	if got := ProductionRuntimePolicy().AttemptOperationTimeout; got != 10*time.Second {
		t.Fatalf("attempt-operation timeout = %s", got)
	}
	t.Run("acknowledged after cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		err := awaitCommit(ctx, 5*time.Second, func(commitCtx context.Context) error {
			cancel()
			if commitCtx.Err() != nil {
				return errors.New("acknowledgement lost to attempt cancellation")
			}
			return nil
		})
		if err != nil {
			t.Fatalf("issued commit = %v", err)
		}
	})
	t.Run("pending past the bound", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		started := time.Now()
		err := awaitCommit(ctx, 20*time.Millisecond, func(commitCtx context.Context) error {
			cancel()
			<-commitCtx.Done()
			return commitCtx.Err()
		})
		if !errors.Is(err, context.Canceled) || time.Since(started) < 20*time.Millisecond {
			t.Fatalf("pending commit past the bound = %v after %s", err, time.Since(started))
		}
	})
	t.Run("detached caller keeps its own bound", func(t *testing.T) {
		err := AwaitCommit(context.WithoutCancel(context.Background()), func(commitCtx context.Context) error {
			if commitCtx.Err() != nil {
				return commitCtx.Err()
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	})
}
