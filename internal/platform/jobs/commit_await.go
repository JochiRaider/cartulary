package jobs

import (
	"context"
	"time"
)

// AwaitCommit invokes an issued owner-transaction commit on a context that
// does not inherit the attempt's cancellation. Cancelling the driver call
// would discard the commit's acknowledgement and turn a determinable outcome
// into an indeterminate one. After the attempt context ends, the wait is
// bounded by the attempt-operation timeout, which bounds attempt persistence
// (Core 01 Jobs execution policy). A caller whose context is never cancelled,
// such as a coordinator that already detached the commit, keeps its own bound.
func AwaitCommit(ctx context.Context, commit func(context.Context) error) error {
	return awaitCommit(ctx, ProductionRuntimePolicy().AttemptOperationTimeout, commit)
}

func awaitCommit(ctx context.Context, bound time.Duration, commit func(context.Context) error) error {
	commitCtx, cancel := context.WithCancel(context.WithoutCancel(ctx))
	defer cancel()
	stop := context.AfterFunc(ctx, func() {
		timer := time.NewTimer(bound)
		defer timer.Stop()
		select {
		case <-timer.C:
			cancel()
		case <-commitCtx.Done():
		}
	})
	defer stop()
	return commit(commitCtx)
}
