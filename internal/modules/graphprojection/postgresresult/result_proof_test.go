package postgresresult_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/graphprojection"
	"github.com/JochiRaider/cartulary/internal/modules/graphprojection/postgresresult"
	"github.com/jackc/pgx/v5"
)

// Exercise the actual current producer before testing the borrowed storage port.
func emptyResultFromEngine(t testing.TB) graphprojection.CompletedResultV2 {
	t.Helper()
	projected, err := graphprojection.ProjectV2(context.Background(), graphprojection.InvocationContextV2{GraphViewID: "proof-view", SourceOwnerID: "network_flow_activity"}, []byte(`{
  "projection_schema_id":"graph_projection.v2", "source_snapshot_id":"proof-snapshot",
  "projection_config":{"projection_version":"proof-v2","declared_source_entity_kinds":[],"declared_source_relationship_kinds":[],"entity_mappings":[],"relationship_mappings":[],"metadata_mappings":[],"aggregation_rules":[],"default_vertex_labels":[],"default_edge_labels":[],"allow_empty_kind_registry":true},
  "source_entities":[],"source_relationships":[],"source_metadata":{},"filters":{"entity_filters":[],"relationship_filters":[],"logic":"and"},"property_definitions":[]
 }`))
	if err != nil {
		t.Fatal(err)
	}
	result, err := projected.CompletedResult()
	if err != nil {
		t.Fatal(err)
	}
	result.PublishedAt = time.Date(2026, 11, 2, 0, 0, 0, 0, time.UTC)
	return result
}

func alteredResultContent(t testing.TB, raw []byte) []byte {
	t.Helper()
	var value map[string]any
	if err := json.Unmarshal(raw, &value); err != nil {
		t.Fatal(err)
	}
	value["properties"] = map[string]any{"tampered": "changed without a new identity"}
	changed, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return changed
}

func checkCurrentResultProofs(t *testing.T, ctx context.Context, db interface {
	BeginTx(context.Context, pgx.TxOptions) (pgx.Tx, error)
}) {
	t.Run("historical numeric bytes", func(t *testing.T) {
		tx, err := db.BeginTx(ctx, pgx.TxOptions{})
		if err != nil {
			t.Fatal(err)
		}
		defer tx.Rollback(ctx)
		checkHistoricalNumericRead(t, ctx, tx)
	})
	t.Run("row fidelity at publication and every content read", func(t *testing.T) {
		result := completedResult(t, "proof-row-fidelity", time.Date(2026, 11, 2, 0, 0, 0, 0, time.UTC))
		tx, err := db.BeginTx(ctx, pgx.TxOptions{})
		if err != nil {
			t.Fatal(err)
		}
		defer tx.Rollback(ctx)
		publisher, err := postgresresult.NewPublisher(tx)
		if err != nil {
			t.Fatal(err)
		}
		original := result.Edges[0].JSON
		result.Edges[0].JSON = []byte("{}")
		if err := publisher.PublishResult(ctx, result); !errors.Is(err, graphprojection.ErrResultV2Invalid) {
			t.Fatalf("divergent row published: %v", err)
		}
		var count int
		if err := tx.QueryRow(ctx, `SELECT count(*) FROM graph_projection_results WHERE projection_result_id=$1`, result.Binding.ProjectionResultID).Scan(&count); err != nil || count != 0 {
			t.Fatalf("invalid row publication mutated storage: %d %v", count, err)
		}
		result.Edges[0].JSON = original
		if err := publisher.PublishResult(ctx, result); err != nil {
			t.Fatal(err)
		}
		// Replace only the isolated edge fixture. Counts, identities, columns,
		// topology and the canonical envelope remain valid; child bytes differ.
		edge := result.Edges[0]
		if _, err := tx.Exec(ctx, `DELETE FROM graph_projection_result_edges WHERE projection_result_id=$1 AND edge_id=$2`, result.Binding.ProjectionResultID, edge.EdgeID); err != nil {
			t.Fatal(err)
		}
		if _, err := tx.Exec(ctx, `INSERT INTO graph_projection_result_edges (projection_result_id,edge_id,edge_kind,src_vertex_id,dst_vertex_id,direction,sort_ordinal,sort_key,edge_json) VALUES ($1,$2,$3,$4,$5,$6,0,$7,$8)`, result.Binding.ProjectionResultID, edge.EdgeID, edge.EdgeKind, edge.SrcVertexID, edge.DstVertexID, edge.Direction, edge.SortKey, []byte("{}")); err != nil {
			t.Fatal(err)
		}
		reader, err := postgresresult.NewReader(tx)
		if err != nil {
			t.Fatal(err)
		}
		if binding, err := reader.LockResultEnvelope(ctx, result.Binding.ProjectionResultID); err != nil || binding != result.Binding {
			t.Fatalf("metadata lock changed content-admission ordering: %v", err)
		}
		for name, read := range map[string]func() error{
			"exact":        func() error { _, err := reader.ReadExactResult(ctx, result.Binding); return err },
			"locked exact": func() error { _, err := reader.ReadExactResultForUpdate(ctx, result.Binding); return err },
			"vertices": func() error {
				_, err := reader.ReadVertices(ctx, result.Binding.ProjectionResultID, graphprojection.MaximumResultVerticesV2)
				return err
			},
			"edges": func() error {
				_, err := reader.ReadEdges(ctx, result.Binding.ProjectionResultID, graphprojection.MaximumResultEdgesV2)
				return err
			},
			"traversal": func() error {
				_, err := reader.Traverse(ctx, graphprojection.TraversalRequestV2{ProjectionResultID: result.Binding.ProjectionResultID, SeedVertexIDs: []string{edge.SrcVertexID}, Direction: graphprojection.TraversalOutgoingV2, MaximumDepth: 1, MaximumVertices: 2, MaximumEdges: 1})
				return err
			},
		} {
			if err := read(); !errors.Is(err, graphprojection.ErrResultV2Invalid) {
				t.Errorf("%s returned divergent stored row: %v", name, err)
			}
		}
	})
	t.Run("publication rejects changed content before mutation", func(t *testing.T) {
		result := emptyResultFromEngine(t)
		result.ResultJSON = alteredResultContent(t, result.ResultJSON)
		tx, err := db.BeginTx(ctx, pgx.TxOptions{})
		if err != nil {
			t.Fatal(err)
		}
		defer tx.Rollback(ctx)
		publisher, err := postgresresult.NewPublisher(tx)
		if err != nil {
			t.Fatal(err)
		}
		if err := publisher.PublishResult(ctx, result); !errors.Is(err, graphprojection.ErrResultV2Invalid) {
			t.Errorf("changed content admitted under original digest: %v", err)
		}
		var count int
		if err := tx.QueryRow(ctx, `SELECT count(*) FROM graph_projection_results WHERE projection_result_id=$1`, result.Binding.ProjectionResultID).Scan(&count); err != nil || count != 0 {
			t.Fatalf("invalid publication mutated storage: %d %v", count, err)
		}
	})
	t.Run("exact reads reject stored content corruption", func(t *testing.T) {
		result := emptyResultFromEngine(t)
		tx, err := db.BeginTx(ctx, pgx.TxOptions{})
		if err != nil {
			t.Fatal(err)
		}
		defer tx.Rollback(ctx)
		publisher, err := postgresresult.NewPublisher(tx)
		if err != nil {
			t.Fatal(err)
		}
		if err := publisher.PublishResult(ctx, result); err != nil {
			t.Fatal(err)
		}
		reader, err := postgresresult.NewReader(tx)
		if err != nil {
			t.Fatal(err)
		}
		loaded, err := reader.ReadExactResult(ctx, result.Binding)
		if err != nil || !bytes.Equal(loaded.ResultJSON, result.ResultJSON) {
			t.Fatal("current producer round trip failed", err)
		}
		// The database correctly forbids mutation of published rows. Replace
		// only this test's row in its rollback transaction to model corrupted
		// initial storage without weakening the production immutability trigger.
		if _, err := tx.Exec(ctx, `DELETE FROM graph_projection_results WHERE projection_result_id=$1`, result.Binding.ProjectionResultID); err != nil {
			t.Fatal(err)
		}
		b := result.Binding
		if _, err := tx.Exec(ctx, `INSERT INTO graph_projection_results
         (projection_result_id, graph_view_id, source_owner_id, source_snapshot_id, projection_schema_id, projection_version, normalized_configuration_sha256, normalized_source_sha256, canonical_output_sha256, vertex_count, edge_count, result_json, published_at)
         VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,0,0,$10,$11)`,
			b.ProjectionResultID, b.GraphViewID, b.SourceOwnerID, b.SourceSnapshotID, b.ProjectionSchemaID, b.ProjectionVersion, b.NormalizedConfigurationSHA256, b.NormalizedSourceSHA256, b.CanonicalOutputSHA256, alteredResultContent(t, result.ResultJSON), result.PublishedAt); err != nil {
			t.Fatal(err)
		}
		for _, read := range []func(context.Context, graphprojection.ResultBindingV2) (graphprojection.CompletedResultV2, error){reader.ReadExactResult, reader.ReadExactResultForUpdate} {
			if _, err := read(ctx, result.Binding); !errors.Is(err, graphprojection.ErrResultV2Invalid) {
				t.Errorf("stored corrupt content returned as exact: %v", err)
			}
		}
	})
}
