package reference_data

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/platform/jobs"
)

func testPortableExecutionDeadline(t *testing.T) {
	t.Helper()
	t.Run("execution budget includes time before container verification", func(t *testing.T) {
		ctx, cancel := context.WithCancelCause(context.Background())
		e := &IncidentReferenceExecution{started: time.Now().Add(-time.Minute), ctx: ctx, cancel: cancel}
		defer e.Close()
		if err := e.activate(60); !errors.Is(err, context.DeadlineExceeded) || !errors.Is(context.Cause(ctx), context.DeadlineExceeded) {
			t.Fatal("expired execution admitted", err)
		}
	})
	for _, test := range []struct {
		name   string
		sample int64
		want   string
	}{
		{"cancellation before equality", 60*int64(time.Second) - 1, "canceled"},
		{"cancellation at equality", 60 * int64(time.Second), "timed_out"},
		{"cancellation after equality", 60*int64(time.Second) + 1, "timed_out"},
	} {
		t.Run(test.name, func(t *testing.T) {
			done := make(chan struct{})
			close(done)
			ctx, cancel := context.WithCancelCause(context.Background())
			e := &IncidentReferenceExecution{started: time.Now().Add(-61 * time.Second), seconds: 60, ctx: ctx, cancel: cancel, observation: &executionObservation{cancellation: &test.sample, err: jobs.ErrCancellationRequested, stop: func() {}, done: done}}
			defer e.Close()
			outcome, code, details := e.ClassifyAbsent(context.Background(), context.DeadlineExceeded)
			if outcome != test.want {
				t.Fatal(outcome, test.want)
			}
			if outcome == "timed_out" && (code != "reference_pack_verification_failed" || details["reason_code"] != "verification_timeout") {
				t.Fatal(code, details)
			}
			// Closing twice does not alter an already classified outcome.
			e.Close()
			if e.failure != test.want {
				t.Fatal(e.failure)
			}
		})
	}
	t.Run("lost execution is recoverable", func(t *testing.T) {
		e := &IncidentReferenceExecution{}
		if outcome, _, _ := e.ClassifyAbsent(context.Background(), jobs.ErrExecutionLost); outcome != "recoverable" {
			t.Fatal(outcome)
		}
	})
}
