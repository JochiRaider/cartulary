package graphassembly

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"reflect"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/graphprojection"
	"github.com/JochiRaider/cartulary/internal/modules/graphprojection/postgresrestore"
	"github.com/JochiRaider/cartulary/internal/modules/graphprojection/postgresresult"
	graphrestore "github.com/JochiRaider/cartulary/internal/modules/graphprojection/restore"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/jackc/pgx/v5"
)

//go:embed testdata/reverse-input.json
var reverseInput []byte

// Captured on the GP-07 engine with a first-contributor endpoint selection.
//
//go:embed testdata/first-contributor.json
var historicAmbiguousInput []byte

func checkReverseReconstruction(t *testing.T, db postgres.DB) {
	ctx := t.Context()
	req := claimedGraphRequest(ctx)
	projected, err := graphprojection.ProjectV2(ctx, graphprojection.InvocationContextV2{GraphViewID: "mixed-graph", SourceOwnerID: "network_flow_activity"}, reverseInput)
	if err != nil || len(projected.Edges) != 2 {
		t.Fatalf("reverse projection: %v", err)
	}
	result, err := projected.CompletedResult()
	if err != nil {
		t.Fatal(err)
	}
	if err = result.Validate(ctx); err != nil {
		t.Fatal(err)
	}
	result.PublishedAt = req.PublishedAt
	tx, err := db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback(ctx)
	pub, err := postgresresult.NewPublisher(tx)
	if err != nil {
		t.Fatal(err)
	}
	broken := result
	broken.ResultJSON = []byte("{}")
	if err := pub.PublishResult(ctx, broken); !errors.Is(err, graphprojection.ErrResultV2Invalid) {
		t.Fatalf("invalid publication: %v", err)
	}
	var count int
	if err := tx.QueryRow(ctx, "SELECT count(*) FROM graph_projection_results WHERE projection_result_id=$1", result.Binding.ProjectionResultID).Scan(&count); err != nil || count != 0 {
		t.Fatalf("partial publication: %d %v", count, err)
	}
	for i := 0; i < 2; i++ {
		if err := pub.PublishResult(ctx, result); err != nil {
			t.Fatal(err)
		}
	}
	if err := tx.Commit(ctx); err != nil {
		t.Fatal(err)
	}
	reader, err := postgresresult.NewReader(db)
	if err != nil {
		t.Fatal(err)
	}
	assertExact := func() {
		t.Helper()
		got, err := reader.ReadExactResult(ctx, result.Binding)
		if err != nil || got.Binding != result.Binding || !bytes.Equal(got.ResultJSON, result.ResultJSON) || !reflect.DeepEqual(got.Vertices, result.Vertices) || !reflect.DeepEqual(got.Edges, result.Edges) {
			t.Fatalf("exact read lost reverse result: %v", err)
		}
	}
	assertExact()
	var historical struct {
		Input   json.RawMessage
		Binding graphprojection.ResultBindingV2
	}
	if err := json.Unmarshal(historicAmbiguousInput, &historical); err != nil {
		t.Fatal(err)
	}
	for _, scenario := range []string{"reverse", "ambiguous", "cancelled"} {
		t.Run(scenario, func(t *testing.T) {
			entry := graphrestore.CurrentRestoreSourceRegistry().Document().Entries[0]
			requestCtx, cancel := context.WithCancel(ctx)
			defer cancel()
			registry, err := graphrestore.NewCurrentRestoreSourceRegistry(graphrestore.RestoreSourceRegistration{Entry: entry, Enumerate: func(context.Context) ([]graphrestore.RestoreCandidate, error) {
				raw, binding := reverseInput, result.Binding
				if scenario == "ambiguous" {
					raw, binding = historical.Input, historical.Binding
				}
				if scenario == "cancelled" {
					cancel()
				}
				return []graphrestore.RestoreCandidate{{CandidateID: binding.GraphViewID, GraphViewID: binding.GraphViewID, SemanticQuerySchemaID: entry.SemanticQuerySchemaIDs[0], SemanticInput: raw, ExpectedBinding: binding}}, nil
			}})
			if err != nil {
				t.Fatal(err)
			}
			request := claimedGraphRequest(requestCtx)
			tx, err := db.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead})
			if err != nil {
				t.Fatal(err)
			}
			defer tx.Rollback(ctx)
			writer, err := postgresrestore.New(tx, db, graphProjectionDerivedStateReconciler{})
			if err != nil {
				t.Fatal(err)
			}
			service, err := graphrestore.NewRestoreService(writer, registry)
			if err != nil {
				t.Fatal(err)
			}
			rebuilt, err := service.Rebuild(ctx, request)
			if scenario == "reverse" {
				if err != nil || !rebuilt.ReadinessSatisfied() || len(rebuilt.RebuiltViews) != 1 || rebuilt.RebuiltViews[0].ProjectionResultID != result.Binding.ProjectionResultID {
					t.Fatalf("reverse reconstruction: %v", err)
				}
			} else {
				if err == nil || rebuilt.ReadinessSatisfied() || len(rebuilt.RebuiltViews) != 0 || len(rebuilt.ClearedTableIDs) != 0 {
					t.Fatalf("failed rebuild claimed committed state: %v", err)
				}
				if err := tx.Rollback(ctx); err != nil {
					t.Fatal(err)
				}
			}
			assertExact()
		})
	}
}
