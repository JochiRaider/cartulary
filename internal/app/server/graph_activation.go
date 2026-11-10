package server

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/app/graphassembly"
	graphrestore "github.com/JochiRaider/cartulary/internal/modules/graphprojection/restore"
	"github.com/JochiRaider/cartulary/internal/modules/recovery"
	"github.com/JochiRaider/cartulary/internal/platform/postgres/dbport"
)

// A claim activation owns one reconstruction attempt. Recovery completion is
// already durable; this gate controls only consumers of derived Graph state.
type claimedGraphActivation struct {
	ready          atomic.Bool
	dependentReady chan struct{}
	expected       *graphrestore.RestoreRebuildRequest
	rebuild        func(context.Context) (graphrestore.RestoreRebuildResult, error)
	fatal          func()
	mu             sync.Mutex
	cancel         context.CancelFunc
	done           chan struct{}
	closed         bool
}

func newClaimedGraphActivation(db dbport.DB, control recovery.TargetControl, claimed bool, now func() time.Time, fatal func()) (*claimedGraphActivation, error) {
	if err := control.Validate(); err != nil || (control.State != recovery.TargetInitialized && control.State != recovery.TargetComplete) || fatal == nil || now == nil || db == nil {
		return nil, errors.New("graph claim activation requires an admitted serving target")
	}
	activation := &claimedGraphActivation{dependentReady: make(chan struct{}), fatal: fatal}
	if !claimed {
		return activation, nil
	}
	if control.State == recovery.TargetInitialized {
		activation.ready.Store(true)
		return activation, nil
	}
	request, err := claimedGraphRebuildRequest(control)
	if err != nil {
		return nil, err
	}
	activation.expected = &request
	activation.rebuild = func(ctx context.Context) (graphrestore.RestoreRebuildResult, error) {
		request.Context = ctx
		request.PublishedAt = now().UTC()
		return graphassembly.RebuildClaimedGraph(ctx, db, request)
	}
	return activation, nil
}

func claimedGraphRebuildRequest(control recovery.TargetControl) (graphrestore.RestoreRebuildRequest, error) {
	if err := control.Validate(); err != nil || control.State != recovery.TargetComplete || *control.Purpose != "restore_target" {
		return graphrestore.RestoreRebuildRequest{}, errors.New("graph reconstruction requires structural completion")
	}
	operation, err := uuid.Parse(*control.OperationID)
	if err != nil {
		return graphrestore.RestoreRebuildRequest{}, err
	}
	generation, err := uuid.Parse(*control.TargetGenerationID)
	if err != nil {
		return graphrestore.RestoreRebuildRequest{}, err
	}
	binding := graphrestore.CurrentRestoreImplementationBinding()
	if *control.RecoveryStateCatalogSHA256 != binding.Binding.RecoveryStateCatalogSHA256 {
		return graphrestore.RestoreRebuildRequest{}, errors.New("graph reconstruction catalog mismatch")
	}
	registry := graphrestore.CurrentRestoreSourceRegistry()
	return graphrestore.RestoreRebuildRequest{
		RestoreOperationID: operation, TargetGenerationID: generation,
		RecoveryStateCatalog: graphrestore.RestoreRecoveryCatalogRef{
			DigestSHA256:  *control.RecoveryStateCatalogSHA256,
			AlgorithmID:   graphrestore.RestoreAlgorithmID,
			GraphTableIDs: graphrestore.RestoreGraphTableIDs(),
		},
		SourceRegistry:        graphrestore.RestoreSourceRegistryRef{Registry: registry, SHA256: registry.DigestSHA256()},
		ImplementationBinding: binding,
	}, nil
}

func (activation *claimedGraphActivation) Ready() bool { return activation.ready.Load() }

func (activation *claimedGraphActivation) Prepare(ctx context.Context, serving <-chan struct{}) error {
	activation.mu.Lock()
	defer activation.mu.Unlock()
	if activation.closed || ctx == nil || serving == nil {
		return errors.New("graph claim activation unavailable")
	}
	if activation.cancel != nil {
		return nil
	}
	ctx, activation.cancel = context.WithCancel(ctx)
	activation.done = make(chan struct{})
	go func() {
		defer close(activation.done)
		defer func() {
			if recover() != nil {
				activation.fatal()
			}
		}()
		select {
		case <-ctx.Done():
			return
		case <-serving:
		}
		if ctx.Err() != nil {
			return
		}
		if activation.rebuild != nil {
			result, err := activation.rebuild(ctx)
			if code, ok := graphrestore.RestoreErrorCodeOf(err); ok && code == graphrestore.RestoreErrorOutcomeIndeterminate {
				activation.fatal()
				return
			}
			if err != nil || !activation.admitsResult(result) || ctx.Err() != nil {
				return
			}
			activation.ready.Store(true)
		}
		if activation.Ready() {
			close(activation.dependentReady)
		}
	}()
	return nil
}

func (activation *claimedGraphActivation) Close(ctx context.Context) error {
	activation.mu.Lock()
	activation.closed = true
	cancel, done := activation.cancel, activation.done
	activation.mu.Unlock()
	if cancel == nil {
		return nil
	}
	cancel()
	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (activation *claimedGraphActivation) admitsResult(result graphrestore.RestoreRebuildResult) bool {
	request := activation.expected
	return request != nil && result.ReadinessSatisfied() &&
		result.RestoreOperationID == request.RestoreOperationID.String() &&
		result.TargetGenerationID == request.TargetGenerationID.String() &&
		result.SourceRegistrySHA256 == request.SourceRegistry.SHA256 &&
		result.ImplementationBindingSHA256 == request.ImplementationBinding.SHA256
}
