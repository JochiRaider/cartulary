package networkflow_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/graphprojection"
	"github.com/JochiRaider/cartulary/internal/modules/graphprojection/postgresresult"
	. "github.com/JochiRaider/cartulary/internal/modules/networkflow"
	"github.com/JochiRaider/cartulary/internal/testutil/revisionsupport"
)

func TestReportingGraphSourceValidatesLeasesReadsAndReleasesExactResult_Integration(t *testing.T) {
	harness, actor, incidentID := startNetworkFlowStoreTest(t, "network-flow-reporting-graph-source")
	store := newTestNetworkFlowStore(t, harness.DB, revisionsupport.MustAppender(t))
	ctx := context.Background()
	now := time.Date(2026, 8, 15, 21, 0, 0, 0, time.UTC)
	graphViewID := "nfgv_cccccccccccccccccccccccccccccccc"
	result := graphResultForReportingSource(graphViewID, now)
	declaration := graphViewDeclarationFixture(graphViewID, incidentID, actor.ID, now)
	declaration.SelectedResult = &GraphViewSelectedResultBinding{
		ProjectionResultID:            result.Binding.ProjectionResultID,
		SourceSnapshotID:              result.Binding.SourceSnapshotID,
		ProjectionSchemaID:            result.Binding.ProjectionSchemaID,
		ProjectionVersion:             result.Binding.ProjectionVersion,
		NormalizedConfigurationSHA256: result.Binding.NormalizedConfigurationSHA256,
		NormalizedSourceSHA256:        result.Binding.NormalizedSourceSHA256,
		CanonicalOutputSHA256:         result.Binding.CanonicalOutputSHA256,
	}

	tx, err := harness.DB.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		t.Fatalf("begin fixture transaction: %v", err)
	}
	publisher, err := postgresresult.NewPublisher(tx)
	if err != nil {
		t.Fatalf("construct result publisher: %v", err)
	}
	if err := publisher.PublishResult(ctx, result); err != nil {
		t.Fatalf("publish exact graph result: %v", err)
	}
	if err := store.InsertGraphViewDeclarationTx(ctx, tx, declaration); err != nil {
		t.Fatalf("insert selected graph declaration: %v", err)
	}
	if err := tx.Commit(ctx); err != nil {
		t.Fatalf("commit graph source fixture: %v", err)
	}

	source, err := NewReportingGraphSource(harness.DB, store)
	if err != nil {
		t.Fatalf("construct Reporting graph source: %v", err)
	}
	jobID := uuid.New()
	tx, err = harness.DB.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		t.Fatalf("begin release admission transaction: %v", err)
	}
	lease, err := source.ValidateAndLeaseResultTx(ctx, tx, incidentID, jobID, result.Binding, now, now.Add(5*time.Minute))
	if err != nil {
		t.Fatalf("validate and lease exact result: %v", err)
	}
	var leaseCount int
	if err := tx.QueryRow(ctx, `SELECT count(*) FROM graph_projection_result_leases WHERE lease_id = $1`, lease.LeaseID).Scan(&leaseCount); err != nil || leaseCount != 1 {
		t.Fatalf("lease is not visible in release admission transaction: count=%d err=%v", leaseCount, err)
	}
	if err := tx.Commit(ctx); err != nil {
		t.Fatalf("commit release admission transaction: %v", err)
	}

	loaded, err := source.ReadAndRenewLeasedResult(ctx, jobID, result.Binding, now.Add(time.Minute), now.Add(6*time.Minute))
	if err != nil {
		t.Fatalf("read and renew leased result: %v", err)
	}
	if loaded.Projection.Binding != result.Binding || len(loaded.Projection.Vertices) != 2 || len(loaded.Projection.Edges) != 1 ||
		len(loaded.LabelCandidates.VertexLabelCandidates) != 2 || len(loaded.LabelCandidates.EdgeLabelCandidates) != 1 ||
		loaded.LabelCandidates.VertexLabelCandidates[0].Endpoint.StringValue != "192.0.2.1" ||
		loaded.LabelCandidates.EdgeLabelCandidates[0].Protocol.IntegerValue != 6 {
		t.Fatalf("exact ordered result drifted: %#v", loaded)
	}

	t.Run("stored identity expiry and committed renewal", func(t *testing.T) {
		storedID := uuid.NewString()
		if _, err := harness.DB.Exec(ctx, `UPDATE graph_projection_result_leases SET lease_id = $1 WHERE lease_id = $2`, storedID, lease.LeaseID); err != nil {
			t.Fatal(err)
		}
		lease.LeaseID = storedID
		if _, err := source.ReadAndRenewLeasedResult(ctx, jobID, result.Binding, now.Add(2*time.Minute), now.Add(3*time.Minute)); err != nil {
			t.Fatalf("stored identity / shorter expiry: %v", err)
		}
		if _, err := source.ReadAndRenewLeasedResult(ctx, jobID, result.Binding, now.Add(3*time.Minute), now.Add(4*time.Minute)); !errors.Is(err, graphprojection.ErrResultV2LeaseNotFound) {
			t.Fatalf("expiry equality: %v", err)
		}
		wrong := result.Binding
		wrong.NormalizedSourceSHA256 = testSHA4
		if _, err := source.ReadAndRenewLeasedResult(ctx, jobID, wrong, now.Add(2*time.Minute), now.Add(8*time.Minute)); !errors.Is(err, graphprojection.ErrResultV2BindingMismatch) {
			t.Fatalf("wrong exact binding: %v", err)
		}
		var expires time.Time
		if err := harness.DB.QueryRow(ctx, `SELECT leased_until FROM graph_projection_result_leases WHERE lease_id = $1`, storedID).Scan(&expires); err != nil || !expires.Equal(now.Add(8*time.Minute)) {
			t.Fatalf("successful renewal must survive failed read: %v %v", expires, err)
		}
		for _, wrongJob := range []uuid.UUID{uuid.Nil, uuid.New()} {
			if content, err := source.ReadAndRenewLeasedResult(ctx, wrongJob, result.Binding, now.Add(2*time.Minute), now.Add(9*time.Minute)); err == nil || len(content.Projection.Vertices) != 0 {
				t.Fatalf("wrong job disclosed content: %#v %v", content, err)
			}
		}
	})
	t.Run("release belongs to caller transaction", func(t *testing.T) {
		tx, err := harness.DB.BeginTx(ctx, pgx.TxOptions{})
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = tx.Rollback(ctx) }()
		if err := source.ReleaseJobLeasesTx(ctx, tx, jobID); err != nil {
			t.Fatal(err)
		}
		var count int
		if err := tx.QueryRow(ctx, `SELECT count(*) FROM graph_projection_result_leases WHERE lease_id = $1`, lease.LeaseID).Scan(&count); err != nil || count != 0 {
			t.Fatalf("transaction deletion: %d %v", count, err)
		}
		if err := tx.Rollback(ctx); err != nil {
			t.Fatal(err)
		}
		if err := harness.DB.QueryRow(ctx, `SELECT count(*) FROM graph_projection_result_leases WHERE lease_id = $1`, lease.LeaseID).Scan(&count); err != nil || count != 1 {
			t.Fatalf("rollback restored lease: %d %v", count, err)
		}
	})

	t.Run("complete scoped release includes expired leases", func(t *testing.T) {
		tx, err := harness.DB.BeginTx(ctx, pgx.TxOptions{})
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = tx.Rollback(ctx) }()
		if _, err := tx.Exec(ctx, `INSERT INTO graph_projection_results SELECT (jsonb_populate_record(NULL::graph_projection_results, to_jsonb(r) || jsonb_build_object('projection_result_id', 'gpres_' || repeat(md5(n::text), 2), 'source_owner_id', CASE WHEN n = 1003 THEN 'neighbor' ELSE $2 END))).* FROM graph_projection_results r CROSS JOIN generate_series(1,1003) n WHERE projection_result_id = $1`, result.Binding.ProjectionResultID, ProfileID); err != nil {
			t.Fatal(err)
		}
		if _, err := tx.Exec(ctx, `INSERT INTO graph_projection_result_leases SELECT (jsonb_populate_record(NULL::graph_projection_result_leases, to_jsonb(l) || jsonb_build_object('lease_id', md5(n::text)::uuid, 'projection_result_id', 'gpres_' || repeat(md5(n::text), 2), 'renewed_at', l.created_at, 'leased_until', CASE WHEN n % 2 = 0 THEN $2::timestamptz ELSE $3::timestamptz END))).* FROM graph_projection_result_leases l CROSS JOIN generate_series(1,1003) n WHERE lease_id = $1`, lease.LeaseID, now.Add(time.Second), now.Add(time.Hour)); err != nil {
			t.Fatal(err)
		}
		for _, dimensions := range []struct{ owner, resource, purpose string }{{"other_owner", jobID.String(), "render"}, {"snapshot_reporting", uuid.NewString(), "render"}, {"snapshot_reporting", jobID.String(), "other_purpose"}} {
			if _, err := tx.Exec(ctx, `INSERT INTO graph_projection_result_leases SELECT (jsonb_populate_record(NULL::graph_projection_result_leases, to_jsonb(l) || jsonb_build_object('lease_id', $2::text, 'lease_owner_id', $3::text, 'lease_owner_resource_id', $4::text, 'lease_purpose', $5::text))).* FROM graph_projection_result_leases l WHERE lease_id = $1`, lease.LeaseID, uuid.NewString(), dimensions.owner, dimensions.resource, dimensions.purpose); err != nil {
				t.Fatal(err)
			}
		}
		for range 2 {
			if err := source.ReleaseJobLeasesTx(ctx, tx, jobID); err != nil {
				t.Fatal(err)
			}
		}
		var left int
		if err := tx.QueryRow(ctx, `SELECT count(*) FROM graph_projection_result_leases`).Scan(&left); err != nil || left != 4 {
			t.Fatalf("neighbor scope leases=%d want=4 err=%v", left, err)
		}
	})

	stale := result.Binding
	stale.SourceSnapshotID = "network-flow-source-stale"
	tx, err = harness.DB.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		t.Fatalf("begin stale validation transaction: %v", err)
	}
	if _, err := source.ValidateAndLeaseResultTx(ctx, tx, incidentID, uuid.New(), stale, now, now.Add(time.Minute)); !errors.Is(err, graphprojection.ErrResultV2SourceStale) {
		t.Fatalf("stale source binding error = %v", err)
	}
	if err := tx.Rollback(ctx); err != nil {
		t.Fatalf("rollback stale validation: %v", err)
	}

	tx, err = harness.DB.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		t.Fatalf("begin lease release transaction: %v", err)
	}
	if err := source.ReleaseJobLeasesTx(ctx, tx, jobID); err != nil {
		t.Fatalf("release job leases: %v", err)
	}
	if err := tx.Commit(ctx); err != nil {
		t.Fatalf("commit lease release: %v", err)
	}
	if _, err := source.ReadAndRenewLeasedResult(ctx, jobID, result.Binding, now.Add(2*time.Minute), now.Add(7*time.Minute)); !errors.Is(err, graphprojection.ErrResultV2LeaseNotFound) {
		t.Fatalf("released lease read error = %v", err)
	}
}

func graphResultForReportingSource(graphViewID string, now time.Time) graphprojection.CompletedResultV2 {
	return graphprojection.CompletedResultV2{
		Binding: graphprojection.ResultBindingV2{
			ProjectionResultID:            graphViewResultID,
			GraphViewID:                   graphViewID,
			SourceOwnerID:                 ProfileID,
			SourceSnapshotID:              "network-flow-source-1",
			ProjectionSchemaID:            graphprojection.ProjectionSchemaIDV2,
			ProjectionVersion:             "network_flow_activity.v1",
			NormalizedConfigurationSHA256: testSHA1,
			NormalizedSourceSHA256:        testSHA2,
			CanonicalOutputSHA256:         testSHA3,
		},
		ResultJSON: []byte(`{"projection_schema_id":"graph_projection.v2"}`),
		Vertices: []graphprojection.ResultVertexV2{
			{VertexID: "vx_aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", VertexKind: "network_flow.ip_endpoint.v1", SortKey: "a", JSON: []byte(`{"vertex_id":"vx_aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","source_entity_ref":{"source_entity_id":"nfe_aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"},"properties":{"endpoint_value":"192.0.2.1"}}`)},
			{VertexID: "vx_bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb", VertexKind: "network_flow.ip_endpoint.v1", SortKey: "b", JSON: []byte(`{"vertex_id":"vx_bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb","source_entity_ref":{"source_entity_id":"nfe_bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"},"properties":{"endpoint_value":"192.0.2.2"}}`)},
		},
		Edges: []graphprojection.ResultEdgeV2{
			{EdgeID: "ed_aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", EdgeKind: "network_flow.flow_edge.v1", SrcVertexID: "vx_aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", DstVertexID: "vx_bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb", Direction: "directed", SortKey: "a", JSON: []byte(`{"edge_id":"ed_aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","source_relationship_ref":{"source_relationship_id":"nff_aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"},"properties":{"ip_protocol":6,"dst_port":443}}`)},
		},
		PublishedAt: now,
	}
}
