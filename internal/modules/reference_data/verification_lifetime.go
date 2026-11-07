package reference_data

import (
	"context"
	"errors"
	"time"

	extensiondeadline "github.com/JochiRaider/cartulary/internal/modules/extensions/deadline"
)

// VerificationLifetime binds owner resource completion to the imported process
// lifetime. Assembly supplies the configured grace and mandatory fatal sink.
type VerificationLifetime struct {
	policy extensiondeadline.InvocationPolicy
}

func NewVerificationLifetime(grace time.Duration, fatal func(error)) (*VerificationLifetime, error) {
	if grace < 0 || fatal == nil {
		return nil, errors.New("reference pack: verification lifetime policy required")
	}
	return &VerificationLifetime{policy: extensiondeadline.InvocationPolicy{Grace: grace, Exhausted: fatal}}, nil
}

type verificationLifetimeKey struct{}

func (l *VerificationLifetime) track(ctx context.Context) context.Context {
	return context.WithValue(extensiondeadline.ReuseCancellationTracking(ctx), verificationLifetimeKey{}, l)
}

// Run joins the callback, including its cleanup, before returning. The owner
// callback classifies semantic outcomes and proven commits; this boundary only
// enforces the shared lifetime, whose fatal outcome cannot be replaced.
func (l *VerificationLifetime) Run(ctx context.Context, run func(context.Context) error) error {
	if l == nil || run == nil {
		return errors.New("reference pack: verification lifetime required")
	}
	ctx = l.track(ctx)
	completion, err := extensiondeadline.Complete(ctx, l.policy, func(ctx context.Context) (struct{}, error) {
		return struct{}{}, run(ctx)
	})
	if err != nil {
		return err
	}
	return completion.Err
}

func closeVerificationWorkspace(ctx context.Context, workspace VerificationWorkspace) error {
	l, ok := ctx.Value(verificationLifetimeKey{}).(*VerificationLifetime)
	if !ok || l == nil {
		return errors.New("reference pack: workspace completion outside verification lifetime")
	}
	return l.completeCleanup(ctx, workspace.Close)
}

func (l *VerificationLifetime) completeCleanup(ctx context.Context, cleanup func(context.Context) error) error {
	completion, err := extensiondeadline.Complete(ctx, l.policy, func(callCtx context.Context) (struct{}, error) {
		return struct{}{}, cleanup(context.WithoutCancel(callCtx))
	})
	if err != nil {
		return err
	}
	return completion.Err
}

func verificationCleanupError(primary, cleanup error) error {
	if verificationFatalError(primary) != nil {
		return primary
	}
	if cleanup != nil {
		return cleanup
	}
	return primary
}

func verificationFatalError(err error) error {
	var fatal interface{ FatalReasonCode() string }
	if errors.As(err, &fatal) && fatal.FatalReasonCode() != "" {
		return err
	}
	return nil
}
