package recovery

import (
	"context"
	"errors"

	"github.com/JochiRaider/cartulary/internal/platform/recoverystate"
	"github.com/jackc/pgx/v5"
)

// AtomicRestoreTargetForTest exposes the shipped PostgreSQL mutation adapter to
// external-package integration tests after real current input/target admission.
func AtomicRestoreTargetForTest(fence *FencedRestore) VNextRestoreTarget {
	return admittedAtomicRestoreTestTarget{fence}
}

type admittedAtomicRestoreTestTarget struct{ fence *FencedRestore }

func (target admittedAtomicRestoreTestTarget) WithAtomicRestore(ctx context.Context, catalog *recoverystate.Catalog, run func(VNextRestoreMutation) error) (err error) {
	fence := target.fence
	fence.operation.Lock()
	defer fence.operation.Unlock()
	if fence.closed || fence.used {
		return ErrTargetIncomplete
	}
	fence.used = true
	defer func() {
		if !fence.committed {
			err = errors.Join(err, rollbackRestoreTransaction(fence.tx))
		}
	}()
	return (&vNextRestoreTarget{target: fence.target, extensionBackups: fence.runner.extensionBackups}).WithAtomicRestore(ctx, catalog, run)
}

func RestoreMutationTransactionForTest(mutation VNextRestoreMutation) pgx.Tx {
	return mutation.(*vNextRestoreMutation).tx
}
