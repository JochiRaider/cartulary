package reference_data

import (
	"context"
	"errors"
	extensiondeadline "github.com/JochiRaider/cartulary/internal/modules/extensions/deadline"
	"github.com/JochiRaider/cartulary/internal/modules/reference_data/internal/packformat"
	"testing"
	"time"
)

func testVerificationLifetime(t testing.TB) *VerificationLifetime {
	t.Helper()
	lifetime, err := NewVerificationLifetime(time.Second, func(err error) { t.Errorf("unexpected verification fatal: %v", err) })
	if err != nil {
		t.Fatal(err)
	}
	return lifetime
}

func verifyTestCanonicalContainer(ctx context.Context, storage VerificationStorage, attempt verificationAttempt, index func(context.Context, packformat.Manifest, string, string) (packformat.ContentSink, error)) (result *verifiedContent, err error) {
	lifetime, _ := NewVerificationLifetime(time.Second, func(err error) { panic(err) })
	err = lifetime.Run(ctx, func(ctx context.Context) error {
		var inner error
		result, inner = verifyCanonicalContainer(ctx, storage, attempt, index)
		return inner
	})
	return
}

// These cases exercise the owner completion boundary, not a second deadline
// implementation. A simulated process sink lets the test release a late callback.
func TestVerificationLifetimeCleanup_Unit(t *testing.T) {
	t.Run("portable operation completion", testPortableExecutionCompletion)
	t.Run("portable fatal precedence", testPortableFatalPrecedence)
	for _, mode := range []string{"canceled", "deadline", "live"} {
		t.Run(mode, func(t *testing.T) {
			type key struct{}
			ctx := context.WithValue(t.Context(), key{}, "retained")
			var cancel context.CancelFunc
			switch mode {
			case "canceled":
				ctx, cancel = context.WithCancel(ctx)
				cancel()
			case "deadline":
				ctx, cancel = context.WithDeadline(ctx, time.Now().Add(-time.Millisecond))
				defer cancel()
			}
			lifetime := testVerificationLifetime(t)
			closed := 0
			err := lifetime.Run(ctx, func(scope context.Context) error {
				for range 2 {
					workspace := &completionWorkspace{close: func(cleanup context.Context) error {
						if cleanup.Err() != nil || cleanup.Value(key{}) != "retained" {
							t.Error("cleanup lost live completion capability or values")
						}
						closed++
						return nil
					}}
					if err := closeVerificationWorkspace(scope, workspace); err != nil {
						return err
					}
				}
				return scope.Err()
			})
			if closed != 2 || (mode == "live" && err != nil) || (mode != "live" && err == nil) {
				t.Fatal(closed, err)
			}
		})
	}
	for _, grace := range []time.Duration{0, 10 * time.Millisecond} {
		t.Run(grace.String(), func(t *testing.T) {
			fatal := make(chan error, 2)
			lifetime, _ := NewVerificationLifetime(grace, func(err error) { fatal <- err })
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()
			entered, release := make(chan struct{}), make(chan struct{})
			finished := make(chan error, 1)
			go func() {
				finished <- lifetime.Run(ctx, func(scope context.Context) error {
					return closeVerificationWorkspace(scope, &completionWorkspace{close: func(context.Context) error { close(entered); <-release; return nil }})
				})
			}()
			<-entered
			cancel()
			select {
			case <-fatal:
			case <-time.After(time.Second):
				t.Fatal("missing fatal consequence")
			}
			select {
			case <-finished:
				t.Fatal("returned while cleanup still owned resources")
			default:
			}
			close(release)
			if err := <-finished; !errors.Is(err, extensiondeadline.ErrGraceExhausted) {
				t.Fatal("late cleanup replaced fatal outcome", err)
			}
			select {
			case <-fatal:
				t.Fatal("fatal consequence emitted twice")
			default:
			}
		})
	}
	t.Run("resources and nested refresh members retain original grace", func(t *testing.T) {
		fatal := make(chan error, 1)
		lifetime, _ := NewVerificationLifetime(40*time.Millisecond, func(err error) { fatal <- err })
		ctx, cancel := context.WithCancel(t.Context())
		defer cancel()
		entered, release := make(chan struct{}), make(chan struct{})
		finished := make(chan error, 1)
		go func() {
			finished <- lifetime.Run(ctx, func(scope context.Context) error {
				// The first member owns multiple resources; the next nested member
				// must inherit the same observation, including its fatal state.
				for range 2 {
					if err := closeVerificationWorkspace(scope, &completionWorkspace{close: func(context.Context) error { return nil }}); err != nil {
						return err
					}
				}
				return lifetime.Run(scope, func(member context.Context) error {
					cancel()
					_ = extensiondeadline.Sample(member, lifetime.policy)
					close(entered)
					<-release
					return lifetime.Run(member, func(next context.Context) error {
						return closeVerificationWorkspace(next, &completionWorkspace{close: func(context.Context) error { return nil }})
					})
				})
			})
		}()
		<-entered
		select {
		case <-fatal:
		case <-time.After(time.Second):
			t.Fatal("original grace was not enforced")
		}
		close(release)
		if err := <-finished; !errors.Is(err, extensiondeadline.ErrGraceExhausted) {
			t.Fatal("nested member reset exhausted grace", err)
		}
	})

	if _, err := NewVerificationLifetime(time.Second, nil); err == nil {
		t.Fatal("missing fatal sink accepted")
	}
}

type completionWorkspace struct {
	VerificationWorkspace
	close func(context.Context) error
}

func (w *completionWorkspace) Close(ctx context.Context) error { return w.close(ctx) }
