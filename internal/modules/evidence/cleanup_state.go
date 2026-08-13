package evidence

import (
	"errors"

	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

// cleanupService owns durable failed-blob claim coordination. Application
// assembly activates it only after publication readiness.
type cleanupService struct {
	pool          postgres.DB
	blobLifecycle blobLifecycleRepository
}

func newCleanupService(pool postgres.DB) (*cleanupService, error) {
	if pool == nil {
		return nil, errors.New("compose Evidence cleanup: Postgres is required")
	}
	return &cleanupService{pool: pool, blobLifecycle: blobLifecycleRepository{db: pool}}, nil
}
