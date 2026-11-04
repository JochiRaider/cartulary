package postgresresult_test

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/graphprojection"
	"github.com/JochiRaider/cartulary/internal/modules/graphprojection/postgresresult"
	"github.com/jackc/pgx/v5"
)

// Captured from the pre-GP-07 writer. Raw envelope and row JSON retain exponent spelling.
//
//go:embed testdata/direct-numeric-v2.json
var historicalNumericFixture []byte

func checkHistoricalNumericRead(t *testing.T, ctx context.Context, tx pgx.Tx) {
	var fixture struct {
		Input  json.RawMessage
		Result graphprojection.CompletedResultV2
	}
	if err := json.Unmarshal(historicalNumericFixture, &fixture); err != nil {
		t.Fatal(err)
	}
	want := fixture.Result
	b := want.Binding
	want.PublishedAt = time.Date(2026, 11, 2, 12, 0, 0, 0, time.UTC)
	if !bytes.Contains(want.ResultJSON, []byte("1e06")) || len(want.Vertices) != 2 || !bytes.Contains(want.Vertices[0].JSON, []byte("1e06")) {
		t.Fatal("incomplete historical fixture")
	}
	// Import the existing rows verbatim, without executing the candidate publisher.
	_, err := tx.Exec(ctx, `INSERT INTO graph_projection_results (projection_result_id, graph_view_id,source_owner_id,source_snapshot_id,projection_schema_id,projection_version,normalized_configuration_sha256,normalized_source_sha256,canonical_output_sha256,vertex_count,edge_count,result_json,published_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,0,$11,$12)`, b.ProjectionResultID, b.GraphViewID, b.SourceOwnerID, b.SourceSnapshotID, b.ProjectionSchemaID, b.ProjectionVersion, b.NormalizedConfigurationSHA256, b.NormalizedSourceSHA256, b.CanonicalOutputSHA256, len(want.Vertices), []byte(want.ResultJSON), want.PublishedAt)
	if err != nil {
		t.Fatal(err)
	}
	for i, row := range want.Vertices {
		if _, err := tx.Exec(ctx, `INSERT INTO graph_projection_result_vertices (projection_result_id,vertex_id,vertex_kind,sort_ordinal,sort_key,vertex_json) VALUES ($1,$2,$3,$4,$5,$6)`, b.ProjectionResultID, row.VertexID, row.VertexKind, i, row.SortKey, []byte(row.JSON)); err != nil {
			t.Fatal(err)
		}
	}
	reader, err := postgresresult.NewReader(tx)
	if err != nil {
		t.Fatal(err)
	}
	check := func(got graphprojection.CompletedResultV2, err error) {
		t.Helper()
		if err != nil || got.Binding != b || !bytes.Equal(got.ResultJSON, want.ResultJSON) || !reflect.DeepEqual(got.Vertices, want.Vertices) || len(got.Edges) != 0 {
			t.Errorf("historical bytes/binding diverged: %v", err)
		}
	}
	check(reader.ReadExactResult(ctx, b))
	check(reader.ReadExactResultForUpdate(ctx, b))
	r, err := graphprojection.ProjectV2(ctx, graphprojection.InvocationContextV2{GraphViewID: b.GraphViewID, SourceOwnerID: b.SourceOwnerID}, fixture.Input)
	if err != nil {
		t.Fatal(err)
	}
	check(r.CompletedResult())
}
