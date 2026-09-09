package networkflow

import (
	"time"

	"github.com/JochiRaider/cartulary/internal/platform/jobs"
)

// NewGraphViewAuthorityTestHandler exercises the production handler with an
// explicitly controlled finalization boundary, without starting a worker.
func NewGraphViewAuthorityTestHandler(store *Store, manager GraphViewJobManager, finalizer GraphViewJobFinalizer) jobs.HandlerFunc {
	module := &Module{store: store, limits: store.limits, now: time.Now,
		graphProjection: newGraphProjectionAdapter(), jobManager: manager, jobFinalizer: finalizer}
	return module.handleGraphViewMaterialization
}
