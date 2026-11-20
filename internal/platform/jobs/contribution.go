package jobs

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
)

// Contribution adds one job kind, its worker and its handler to a process in
// addition to the published job catalog. Only harness assemblies supply one
// (TH-HARNESS-REQ-450); production assembly passes none, so a contributed job
// kind never reaches a production catalog.
type Contribution struct {
	Definition Definition
	Worker     WorkerRuntimeContract
	// Bind receives the process's Jobs capabilities once they are assembled
	// and returns the worker's handler.
	Bind func(ContributionPorts) HandlerFunc
}

// ContributionPorts are the Jobs capabilities a contributed job kind receives.
type ContributionPorts struct {
	// Enqueue commits one queued job of the contributed kind, submitted by the
	// given user, and wakes the runner. It returns the job ID.
	Enqueue func(ctx context.Context, submittedBy uuid.UUID, payload json.RawMessage) (uuid.UUID, error)
	// Payload reads the handler payload of a running attempt.
	Payload func(ctx context.Context, execution Execution) (json.RawMessage, error)
}
