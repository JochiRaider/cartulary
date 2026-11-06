package graphprojection

import (
	"context"
	"time"
)

// ReconstructExactV2 recomputes retained semantic inputs and admits only the
// selected immutable result. Transfer and deployment recovery share this pure
// boundary; neither operation's publication lifecycle is invoked here.
func ReconstructExactV2(ctx context.Context, expected ResultBindingV2, semanticInput []byte, publishedAt time.Time, cancellationCheck func(context.Context) error) (CompletedResultV2, error) {
	if ctx == nil || publishedAt.IsZero() || expected.ValidateIdentity() != nil {
		return CompletedResultV2{}, ErrResultV2Invalid
	}
	projection, err := ProjectV2(ctx, InvocationContextV2{GraphViewID: expected.GraphViewID, SourceOwnerID: expected.SourceOwnerID, CancellationCheck: func(c context.Context) error {
		if err := ctx.Err(); err != nil {
			return err
		}
		if cancellationCheck != nil {
			return cancellationCheck(c)
		}
		return nil
	}}, semanticInput)
	if err != nil {
		return CompletedResultV2{}, err
	}
	if projection.ResultBindingV2() != expected {
		return CompletedResultV2{}, ErrResultV2BindingMismatch
	}
	completed, err := projection.CompletedResult()
	if err != nil {
		return CompletedResultV2{}, err
	}
	completed.PublishedAt = publishedAt.UTC()
	return completed, nil
}
