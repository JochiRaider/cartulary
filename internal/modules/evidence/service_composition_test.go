package evidence_test

import (
	"github.com/JochiRaider/cartulary/internal/app/projectionassembly"
	"github.com/JochiRaider/cartulary/internal/modules/collaboration"
	"github.com/JochiRaider/cartulary/internal/modules/evidence"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

func newTestBlobLifecycleService(
	pool postgres.DB,
	appender *revisions.Appender,
	intents collaboration.RecordChangedAppender,
) *evidence.BlobLifecycleService {
	projectionRuntime, err := projectionassembly.Build(pool)
	if err != nil {
		panic(err)
	}
	service, err := evidence.NewBlobLifecycleService(evidence.BlobLifecycleDependencies{
		Postgres:       pool,
		Revisions:      appender,
		Projections:    projectionRuntime.EvidencePorts().Rows,
		SupportEffects: projectionRuntime.EvidencePorts().SupportEffects,
		Collaboration:  intents,
	})
	if err != nil {
		panic(err)
	}
	return service
}

func newTestRouteOperations(
	pool postgres.DB,
	appender *revisions.Appender,
	intents collaboration.RecordChangedAppender,
) *evidence.RouteOperations {
	blobs := newTestBlobLifecycleService(pool, appender, intents)
	access, err := evidence.NewAccessHandleService(pool)
	if err != nil {
		panic(err)
	}
	operations, err := evidence.NewRouteOperations(blobs, access)
	if err != nil {
		panic(err)
	}
	return operations
}
