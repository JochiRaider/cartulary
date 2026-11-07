package reference_data

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	extensiondeadline "github.com/JochiRaider/cartulary/internal/modules/extensions/deadline"
	"github.com/JochiRaider/cartulary/internal/platform/jobs"
)

func testPortableExecutionDeadline(t *testing.T) {
	t.Helper()
	t.Run("execution budget includes time before container verification", func(t *testing.T) {
		ctx, cancel := context.WithCancelCause(context.Background())
		e := &IncidentReferenceExecution{started: time.Now().Add(-time.Minute), ctx: ctx, cancel: cancel}
		defer cancel(nil)
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
			defer cancel(nil)
			outcome, code, details, classifyErr := e.ClassifyAbsent(context.Background(), context.DeadlineExceeded)
			if classifyErr != nil {
				t.Fatal(classifyErr)
			}
			if outcome != test.want {
				t.Fatal(outcome, test.want)
			}
			if outcome == "timed_out" && (code != "reference_pack_verification_failed" || details["reason_code"] != "verification_timeout") {
				t.Fatal(code, details)
			}
			// Stopping observation does not alter an already classified outcome.
			e.observation.finish()
			if e.failure != test.want {
				t.Fatal(e.failure)
			}
		})
	}
	t.Run("lost execution is recoverable", func(t *testing.T) {
		e := &IncidentReferenceExecution{}
		if outcome, _, _, _ := e.ClassifyAbsent(context.Background(), jobs.ErrExecutionLost); outcome != "recoverable" {
			t.Fatal(outcome)
		}
	})
}

func testPortableExecutionCompletion(t *testing.T) {
	t.Run("observer shutdown retains owner deadline", func(t *testing.T) {
		fatal := make(chan error, 1)
		lifetime, _ := NewVerificationLifetime(10*time.Millisecond, func(err error) { fatal <- err })
		ctx, cancel := context.WithCancelCause(lifetime.track(t.Context()))
		owner := &incidentReferences{verifier: &verificationService{referenceDependencies: &referenceDependencies{lifetime: lifetime}}}
		release := make(chan struct{})
		stopping := make(chan struct{})
		scope := &IncidentReferenceExecution{owner: owner, ctx: ctx, cancel: cancel,
			observation: &executionObservation{stop: func() { close(stopping) }, done: release}}
		result := make(chan error, 1)
		go func() {
			result <- scope.run(func(context.Context, *IncidentReferenceExecution) error {
				scope.timer = time.AfterFunc(50*time.Millisecond, func() { cancel(context.DeadlineExceeded) })
				return nil
			})
		}()
		<-stopping
		select {
		case <-fatal:
		case <-time.After(time.Second):
			t.Error("observer shutdown escaped the owner deadline")
		}
		close(release)
		if err := <-result; !errors.Is(err, extensiondeadline.ErrGraceExhausted) {
			t.Fatal(err)
		}
	})
	t.Run("later stages inherit an already exhausted grace", func(t *testing.T) {
		fatal := make(chan error, 2)
		lifetime, _ := NewVerificationLifetime(time.Millisecond, func(err error) { fatal <- err })
		tracked := extensiondeadline.WithCancellationTracking(t.Context())
		interrupted, stop := context.WithDeadline(tracked, time.Now().Add(-time.Second))
		defer stop()
		if err := lifetime.Run(interrupted, func(context.Context) error { return nil }); !errors.Is(err, extensiondeadline.ErrGraceExhausted) {
			t.Fatal(err)
		}
		ctx, cancel := context.WithCancelCause(lifetime.track(tracked))
		owner := &incidentReferences{verifier: &verificationService{referenceDependencies: &referenceDependencies{lifetime: lifetime}}}
		scope := &IncidentReferenceExecution{owner: owner, ctx: ctx, cancel: cancel}
		err := scope.run(func(context.Context, *IncidentReferenceExecution) error {
			for _, stage := range []string{"next container", "source owner", "parent completion"} {
				if err := scope.CompleteCleanup(func(context.Context) error { return nil }); !errors.Is(err, extensiondeadline.ErrGraceExhausted) {
					t.Error(stage, "renewed grace", err)
				}
			}
			return nil
		})
		if !errors.Is(err, extensiondeadline.ErrGraceExhausted) || len(fatal) != 1 {
			t.Fatal("inherited fatal was reset or emitted twice", err, len(fatal))
		}
	})
	for _, mode := range []string{"normal zero grace", "prior cancellation", "parent deadline", "owner deadline", "shared grace", "zero grace"} {
		t.Run(mode, func(t *testing.T) {
			type valueKey struct{}
			ctx := context.WithValue(t.Context(), valueKey{}, "preserved")
			if mode == "parent deadline" {
				var stop context.CancelFunc
				ctx, stop = context.WithDeadline(ctx, time.Now().Add(-time.Millisecond))
				defer stop()
			}
			grace := 100 * time.Millisecond
			if mode == "zero grace" || mode == "normal zero grace" {
				grace = 0
			}
			fatal := make(chan error, 2)
			lifetime, err := NewVerificationLifetime(grace, func(err error) { fatal <- err })
			if err != nil {
				t.Fatal(err)
			}
			ctx, cancel := context.WithCancelCause(lifetime.track(ctx))
			owner := &incidentReferences{verifier: &verificationService{referenceDependencies: &referenceDependencies{lifetime: lifetime}}}
			scope := &IncidentReferenceExecution{owner: owner, ctx: ctx, cancel: cancel, started: time.Now()}
			if mode == "prior cancellation" {
				cancel(context.Canceled)
			}
			release := make(chan struct{})
			finished := make(chan error, 1)
			go func() {
				finished <- scope.run(func(_ context.Context, e *IncidentReferenceExecution) error {
					if mode == "owner deadline" {
						e.started = time.Now().Add(-time.Minute + 20*time.Millisecond)
						if err := e.activate(60); err != nil {
							return err
						}
						<-ctx.Done()
					}
					if mode == "shared grace" || mode == "zero grace" {
						cancel(context.Canceled)
					}
					for i := 0; i < 2; i++ {
						err := e.CompleteCleanup(func(cleanup context.Context) error {
							if cleanup.Err() != nil || cleanup.Value(valueKey{}) != "preserved" {
								return errors.New("cleanup context lost")
							}
							if i == 1 && (mode == "shared grace" || mode == "zero grace") {
								<-release
							}
							return nil
						})
						if err != nil {
							return err
						}
					}
					if mode == "normal zero grace" {
						return nil
					}
					return ctx.Err()
				})
			}()
			exhausted := mode == "shared grace" || mode == "zero grace"
			if exhausted {
				select {
				case <-fatal:
				case <-time.After(time.Second):
					t.Fatal("missing fatal consequence")
				}
				// Zero grace may stop before the second callback. Either way no scope can
				// outlive its caller; the outer result retains the fatal selection.
				close(release)
			}
			err = <-finished
			if exhausted && !errors.Is(err, extensiondeadline.ErrGraceExhausted) {
				t.Fatal("lost fatal completion", err)
			}
			if !exhausted && mode != "normal zero grace" && err == nil {
				t.Fatal("lost interruption")
			}
			if mode == "normal zero grace" && err != nil {
				t.Fatal("normal teardown manufactured cancellation", err)
			}
			if scope.active.Load() {
				t.Fatal("scope remained active")
			}
			if err := scope.CompleteCleanup(func(context.Context) error { t.Fatal("escaped callback ran"); return nil }); err == nil {
				t.Fatal("escaped completion accepted")
			}
		})
	}
}

func testPortableFatalPrecedence(t *testing.T) {
	t.Run("exhaustion while joining observation precedes lease loss", func(t *testing.T) {
		fatal := make(chan error, 1)
		lifetime, _ := NewVerificationLifetime(10*time.Millisecond, func(err error) { fatal <- err })
		ctx, cancel := context.WithCancelCause(lifetime.track(t.Context()))
		owner := &incidentReferences{verifier: &verificationService{referenceDependencies: &referenceDependencies{lifetime: lifetime}}}
		release := make(chan struct{})
		scope := &IncidentReferenceExecution{owner: owner, ctx: ctx, cancel: cancel,
			observation: &executionObservation{stop: func() { cancel(context.Canceled) }, done: release, err: jobs.ErrExecutionLost}}
		done := make(chan error, 1)
		go func() {
			done <- scope.run(func(context.Context, *IncidentReferenceExecution) error {
				outcome, _, _, err := scope.ClassifyAbsent(context.Background(), errors.New("late failure"))
				if outcome != "" || !errors.Is(err, extensiondeadline.ErrGraceExhausted) {
					t.Error("lease loss replaced fatal", outcome, err)
				}
				return err
			})
		}()
		select {
		case <-fatal:
		case <-time.After(time.Second):
			t.Error("missing observer exhaustion")
		}
		close(release)
		if err := <-done; !errors.Is(err, extensiondeadline.ErrGraceExhausted) {
			t.Fatal(err)
		}
	})
	for _, late := range []error{nil, errors.New("close failed"), context.Canceled, context.DeadlineExceeded, &ContentRejection{Code: "bad_content"}} {
		t.Run(fmt.Sprint(late), func(t *testing.T) {
			fatal := make(chan error, 1)
			lifetime, _ := NewVerificationLifetime(10*time.Millisecond, func(err error) { fatal <- err })
			ctx, cancel := context.WithCancelCause(lifetime.track(t.Context()))
			owner := &incidentReferences{verifier: &verificationService{referenceDependencies: &referenceDependencies{lifetime: lifetime}}}
			scope := &IncidentReferenceExecution{owner: owner, ctx: ctx, cancel: cancel}
			release := make(chan struct{})
			result := make(chan error, 1)
			go func() {
				result <- scope.run(func(_ context.Context, e *IncidentReferenceExecution) error {
					cancel(context.Canceled)
					err := e.CompleteCleanup(func(context.Context) error { <-release; return late })
					if !errors.Is(err, extensiondeadline.ErrGraceExhausted) {
						t.Error("cleanup replaced fatal")
						return errors.New("cleanup replaced fatal")
					}
					err = verificationCleanupError(err, late)
					if !errors.Is(err, extensiondeadline.ErrGraceExhausted) {
						t.Error("merge replaced fatal")
						return errors.New("merge replaced fatal")
					}
					for _, cause := range []error{err, late} {
						outcome, code, details, failure := e.ClassifyAbsent(context.Background(), cause)
						if outcome != "" || code != "" || details != nil || !errors.Is(failure, extensiondeadline.ErrGraceExhausted) || e.failure != "" {
							t.Error("fatal authorized classification")
							return errors.New("fatal authorized classification")
						}
					}
					if !errors.Is(e.AbortTx(context.Background(), nil), extensiondeadline.ErrGraceExhausted) {
						t.Error("fatal authorized abort")
						return errors.New("fatal authorized abort")
					}
					return err
				})
			}()
			select {
			case <-fatal:
			case <-time.After(time.Second):
				t.Fatal("no fatal selection")
			}
			close(release)
			if err := <-result; !errors.Is(err, extensiondeadline.ErrGraceExhausted) {
				t.Fatal(err)
			}
		})
	}
}
