package networkflow

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/graphprojection"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/google/uuid"
)

type fixtureGraphReader struct {
	*store
	visited int
}

func (r *fixtureGraphReader) IterateRowsForTables(ctx context.Context, incident uuid.UUID, ids []string, visit func(flowRow) error) error {
	return r.store.IterateRowsForTables(ctx, incident, ids, func(row flowRow) error {
		r.visited++
		return visit(row)
	})
}

// CaptureFixtureGraph observes the private production composition boundary over
// already imported rows. The adapter still invokes the real Graph Projection
// engine. No production hook or alternative graph implementation is introduced.
func CaptureFixtureGraph(t testing.TB, db postgres.DB, limits EffectiveLimits, incident uuid.UUID, raw json.RawMessage) map[string]any {
	t.Helper()
	invocations := []map[string]any{}
	reader := &fixtureGraphReader{store: newStore(db, limits)}
	adapter := &graphProjectionAdapter{project: func(ctx context.Context, invocation graphprojection.InvocationContextV2, input []byte) (graphprojection.ProjectionResultV2, error) {
		invocations = append(invocations, map[string]any{
			"source_owner_id": invocation.SourceOwnerID, "graph_view_id": invocation.GraphViewID,
			"cancellation_check_present": invocation.CancellationCheck != nil,
			"projection_input":           json.RawMessage(append([]byte(nil), input...)),
		})
		return graphprojection.ProjectV2(ctx, invocation, input)
	}}
	composer, err := newGraphSourceComposer(reader, limits, adapter, nil)
	if err != nil {
		t.Fatal(err)
	}
	request, failure := decodeGraphQueryRequestValue(bytes.NewReader(raw), limits)
	var result any
	if failure == nil {
		var composition graphComposition
		composition, failure = composer.composeGraph(context.Background(), incident, request)
		if failure == nil {
			result = graphQueryResultResource(composition)
		}
	}
	var rejected any
	if failure != nil {
		presented := semanticHTTPError(failure)
		rejected = map[string]any{"status": presented.Status, "code": presented.Code, "message": presented.Message, "retryable": presented.Retryable, "details": presented.Details}
	}
	return map[string]any{"projection_invocations": invocations, "visited_rows": reader.visited, "result": result, "failure": rejected}
}
