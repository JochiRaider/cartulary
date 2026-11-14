package graphassembly

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"reflect"
	"slices"
	"testing"

	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/graphprojection"
	"github.com/JochiRaider/cartulary/internal/modules/graphprojection/postgresrestore"
	"github.com/JochiRaider/cartulary/internal/modules/graphprojection/postgresresult"
	graphrestore "github.com/JochiRaider/cartulary/internal/modules/graphprojection/restore"
	"github.com/JochiRaider/cartulary/internal/platform/postgres/dbport"
)

//go:embed testdata/reverse-input.json
var reverseInput []byte

// Captured on the GP-07 engine with a first-contributor endpoint selection.
//
//go:embed testdata/first-contributor.json
var historicAmbiguousInput []byte

//go:embed testdata/registry-order.json
var registryOrderVector []byte

//go:embed testdata/label-capabilities.json
var labelCapabilityVectors []byte

type labelMigration struct {
	Name           string
	Input          json.RawMessage
	Before, After  graphprojection.CompletedResultV2
	ReadableBefore bool
}

func checkReverseReconstruction(t *testing.T, db dbport.DB) {
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
	assertExact := func(t *testing.T, want graphprojection.CompletedResultV2) {
		t.Helper()
		got, err := reader.ReadExactResult(ctx, want.Binding)
		if err != nil || got.Binding != want.Binding || !bytes.Equal(got.ResultJSON, want.ResultJSON) || !reflect.DeepEqual(got.Vertices, want.Vertices) || !reflect.DeepEqual(got.Edges, want.Edges) {
			t.Fatalf("exact read lost reverse result: %v", err)
		}
	}
	assertExact(t, result)
	var migrations []labelMigration
	if err := json.Unmarshal(labelCapabilityVectors, &migrations); err != nil || len(migrations) != 2 {
		t.Fatalf("label fixtures: %v", err)
	}
	for _, m := range migrations {
		t.Run(m.Name+" projection", func(t *testing.T) {
			got, err := graphprojection.ProjectV2(ctx, graphprojection.InvocationContextV2{GraphViewID: m.After.Binding.GraphViewID, SourceOwnerID: m.After.Binding.SourceOwnerID}, m.Input)
			if err != nil {
				t.Fatal(err)
			}
			completed, err := got.CompletedResult()
			if err != nil {
				t.Fatal(err)
			}
			if completed.Binding == m.Before.Binding && bytes.Equal(completed.ResultJSON, m.Before.ResultJSON) {
				t.Fatal("writer reproduced independently pinned historical label registry")
			}
			if completed.Binding != m.After.Binding || !bytes.Equal(completed.ResultJSON, m.After.ResultJSON) ||
				!reflect.DeepEqual(completed.Vertices, m.Before.Vertices) || !reflect.DeepEqual(completed.Edges, m.Before.Edges) {
				t.Fatalf("label cutover changed unexpected bytes: %#v", completed.Binding)
			}
			if err := m.Before.Validate(ctx); (err == nil) != m.ReadableBefore || err != nil && !errors.Is(err, graphprojection.ErrResultV2Invalid) {
				t.Fatalf("historic label admission: %v", err)
			}
		})
	}
	var historical struct {
		Input   json.RawMessage
		Binding graphprojection.ResultBindingV2
	}
	if err := json.Unmarshal(historicAmbiguousInput, &historical); err != nil {
		t.Fatal(err)
	}
	var order struct {
		Input         json.RawMessage
		Before, After graphprojection.CompletedResultV2
	}
	if err := json.Unmarshal(registryOrderVector, &order); err != nil {
		t.Fatal(err)
	}
	got, err := graphprojection.ProjectV2(ctx, graphprojection.InvocationContextV2{GraphViewID: order.After.Binding.GraphViewID, SourceOwnerID: order.After.Binding.SourceOwnerID}, order.Input)
	if err != nil {
		t.Fatal(err)
	}
	current, err := got.CompletedResult()
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Equal(current.ResultJSON, order.Before.ResultJSON) && current.Binding == order.Before.Binding {
		t.Error("writer still reproduces the fixed joined-order envelope and binding")
	}
	if !reflect.DeepEqual(current, order.After) {
		t.Errorf("component-order fixture mismatch: got %+v, want %+v", current.Binding, order.After.Binding)
	}
	var permuted map[string]any
	if err := json.Unmarshal(order.Input, &permuted); err != nil {
		t.Fatal(err)
	}
	slices.Reverse(permuted["property_definitions"].([]any))
	config := permuted["projection_config"].(map[string]any)
	slices.Reverse(config["metadata_mappings"].([]any))
	slices.Reverse(config["entity_mappings"].([]any))
	raw, err := json.Marshal(permuted)
	if err != nil {
		t.Fatal(err)
	}
	reordered, err := graphprojection.ProjectV2(ctx, graphprojection.InvocationContextV2{GraphViewID: order.After.Binding.GraphViewID, SourceOwnerID: order.After.Binding.SourceOwnerID}, raw)
	if err != nil || reordered.ResultBindingV2() != order.After.Binding {
		t.Fatalf("registry input permutation: %v", err)
	}
	if err := order.Before.Validate(ctx); !errors.Is(err, graphprojection.ErrResultV2Invalid) {
		t.Errorf("old declaration order accepted: %v", err)
	}
	// The fixed reverse vector uses this input without the optional field definitions.
	// Preserve the exact numeric tokens when constructing its source request.
	var members map[string]json.RawMessage
	if err := json.Unmarshal(reverseInput, &members); err != nil {
		t.Fatal(err)
	}
	members["property_definitions"] = json.RawMessage("[]")
	directionInput, err := json.Marshal(members)
	if err != nil {
		t.Fatal(err)
	}
	rebuiltDirection, err := graphprojection.ProjectV2(ctx, graphprojection.InvocationContextV2{GraphViewID: result.Binding.GraphViewID, SourceOwnerID: result.Binding.SourceOwnerID}, directionInput)
	if err != nil {
		t.Fatal(err)
	}
	expectedDirection, err := rebuiltDirection.CompletedResult()
	if err != nil {
		t.Fatal(err)
	}
	const correctedID = "gpres_ec577029dd9f34da22993c844f4987c5454c22fef4b561d8190a83cf758be485"
	if expectedDirection.Binding.ProjectionResultID != correctedID || expectedDirection.Binding.CanonicalOutputSHA256 != "fd927b7d014da97e7c896c3e3518b51b18a286e4981a04b9a4ecb938e32b6403" {
		t.Errorf("direction rematerialization: %+v", expectedDirection.Binding)
	}
	// The old registry overclaims the policy proved by its only mapping.
	obsoleteDirection := expectedDirection.Binding
	obsoleteDirection.ProjectionResultID = "gpres_57e5a323f6e8c51160008243fd83be74cdaf1d592021d459cb384981d013510c"
	obsoleteDirection.CanonicalOutputSHA256 = "0b68a8d17031188653464ad60e36eb63575cd2d7495c52c03ce1ecefbd79f98a"
	// A failing restore clears inside its borrowed transaction. Verify the
	// previously committed result survives rollback and the replacement is absent.
	type restoreCase struct {
		name    string
		input   json.RawMessage
		binding graphprojection.ResultBindingV2
		want    *graphprojection.CompletedResultV2

		replacement *graphprojection.CompletedResultV2
	}
	// Simulate a source owner correcting a relationship mapping with intersecting key lists.
	var correctedInput map[string]any
	decoder := json.NewDecoder(bytes.NewReader(reverseInput))
	decoder.UseNumber()
	if err := decoder.Decode(&correctedInput); err != nil {
		t.Fatal(err)
	}
	relationship := correctedInput["projection_config"].(map[string]any)["relationship_mappings"].([]any)[0].(map[string]any)
	relationship["required_property_keys"] = []string{"group"}
	relationship["optional_property_keys"] = []string{"group"}
	badInput, err := json.Marshal(correctedInput)
	if err != nil {
		t.Fatal(err)
	}
	view := graphprojection.InvocationContextV2{GraphViewID: result.Binding.GraphViewID, SourceOwnerID: result.Binding.SourceOwnerID}
	_, badErr := graphprojection.ProjectV2(ctx, view, badInput)
	relationship["optional_property_keys"] = []string{}
	fixedInput, err := json.Marshal(correctedInput)
	if err != nil {
		t.Fatal(err)
	}
	fixed, err := graphprojection.ProjectV2(ctx, view, fixedInput)
	if err != nil {
		t.Fatal(err)
	}
	repaired, err := fixed.CompletedResult()
	if err != nil {
		t.Fatal(err)
	}
	var pe *graphprojection.ProjectionErrorV2
	if !errors.As(badErr, &pe) || pe.ReasonCode != "invalid_mapping_rule" {
		t.Fatalf("overlap verdict: %v", badErr)
	}
	if repaired.Binding.ProjectionResultID != "gpres_3ccdb10211fdcd175e6fac2bb15f3f6d9d87566e98090502da9a9b0183113d05" {
		t.Fatal("corrected binding diverged")
	}
	// Captured on GP-11. Correcting the key lists changes only configuration and binding.
	oldMapping := repaired.Binding
	oldMapping.CanonicalOutputSHA256 = "274db2d52afbe43c4df1dd1d19d08b1430b3e53ba297da83b6ba8eae447ec055"
	oldMapping.NormalizedConfigurationSHA256 = "8938bebc1b56d4ac4203d837875fe07bb10aea978bdf0c9bf008ba6a830a22c7"
	oldMapping.ProjectionResultID = "gpres_3d981c993fec303fb1e1cb1e34fb680cfe77251d96334001c988308a1aae455b"

	previousEnvelope := expectedDirection.Binding
	previousEnvelope.CanonicalOutputSHA256 = "ee71acff104428c64d31a48613a545900713ac9dbce854f0f373f689158f1b2f"
	previousEnvelope.ProjectionResultID = "gpres_b356854df17070a45dbf005adfc715170cab59b2e5badb022dea0dfe1854a586"
	cases := []restoreCase{
		{name: "previous envelope shape", input: directionInput, binding: previousEnvelope, replacement: &expectedDirection},
		{name: "reverse", input: reverseInput, binding: result.Binding, want: &result},
		{name: "ambiguous", input: historical.Input, binding: historical.Binding},
		{name: "cancelled", input: reverseInput, binding: result.Binding},
		{name: "old declaration order", input: order.Input, binding: order.Before.Binding, replacement: &order.After},
		{name: "new declaration order", input: order.Input, binding: order.After.Binding, want: &order.After},
		{name: "old direction union", input: directionInput, binding: obsoleteDirection, replacement: &expectedDirection},
		{name: "new direction union", input: directionInput, binding: expectedDirection.Binding, want: &expectedDirection},
	}
	cases = append(cases,
		restoreCase{name: "old overlapping declaration", input: fixedInput, binding: oldMapping, replacement: &repaired},
		restoreCase{name: "corrected disjoint declaration", input: fixedInput, binding: repaired.Binding, want: &repaired})
	for i := range migrations {
		m := &migrations[i]
		old := restoreCase{name: "old " + m.Name, input: m.Input, binding: m.Before.Binding, replacement: &m.After}
		cases = append(cases, old, restoreCase{name: "new " + m.Name, input: m.Input, binding: m.After.Binding, want: &m.After})
	}
	shuffled, tied := timestampTieReconstruction(t, view)
	cases = append(cases, restoreCase{name: "equal timestamp extrema", input: shuffled, binding: tied.Binding, want: &tied})

	publish := func(t *testing.T, selected graphprojection.CompletedResultV2) {
		t.Helper()
		tx, err := db.BeginTx(ctx, pgx.TxOptions{})
		if err != nil {
			t.Fatal(err)
		}
		defer tx.Rollback(ctx)
		publisher, err := postgresresult.NewPublisher(tx)
		if err != nil {
			t.Fatal(err)
		}
		selected.PublishedAt = req.PublishedAt
		if err := publisher.PublishResult(ctx, selected); err != nil {
			t.Fatal(err)
		}
		if err := tx.Commit(ctx); err != nil {
			t.Fatal(err)
		}
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.want == nil {
				publish(t, result)
			}
			entry := graphrestore.CurrentRestoreSourceRegistry().Document().Entries[0]
			requestCtx, cancel := context.WithCancel(ctx)
			defer cancel()
			registry, err := graphrestore.NewCurrentRestoreSourceRegistry(graphrestore.RestoreSourceRegistration{Entry: entry, Enumerate: func(context.Context) ([]graphrestore.RestoreCandidate, error) {
				if tc.name == "cancelled" {
					cancel()
				}
				return []graphrestore.RestoreCandidate{{CandidateID: tc.binding.GraphViewID, GraphViewID: tc.binding.GraphViewID, SemanticQuerySchemaID: entry.SemanticQuerySchemaIDs[0], SemanticInput: tc.input, ExpectedBinding: tc.binding}}, nil
			}})
			if err != nil {
				t.Fatal(err)
			}
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
			rebuilt, err := service.Rebuild(ctx, claimedGraphRequest(requestCtx))
			if tc.want != nil {
				if err != nil || !rebuilt.ReadinessSatisfied() || len(rebuilt.RebuiltViews) != 1 || rebuilt.RebuiltViews[0].ProjectionResultID != tc.want.Binding.ProjectionResultID {
					t.Fatalf("rebuild: %v", err)
				}
				assertExact(t, *tc.want)
				publish(t, *tc.want)
				assertExact(t, *tc.want)
				return
			}
			if err == nil || rebuilt.ReadinessSatisfied() || len(rebuilt.RebuiltViews) != 0 || len(rebuilt.ClearedTableIDs) != 0 {
				t.Fatalf("failed rebuild claimed committed state: %v", err)
			}
			if tc.replacement != nil {
				if code, ok := graphrestore.RestoreErrorCodeOf(err); !ok || code != graphrestore.RestoreErrorInvalidCandidate {
					t.Fatalf("expected binding mismatch after reprojection: %v", err)
				}
			}
			if err := tx.Rollback(ctx); err != nil {
				t.Fatal(err)
			}
			assertExact(t, result)
			if tc.replacement != nil {
				if _, err := reader.ReadExactResult(ctx, tc.replacement.Binding); !errors.Is(err, graphprojection.ErrResultV2NotFound) {
					t.Fatalf("rollback exposed replacement: %v", err)
				}
			}
		})
	}
}

func timestampTieReconstruction(t *testing.T, view graphprojection.InvocationContextV2) (json.RawMessage, graphprojection.CompletedResultV2) {
	t.Helper()
	// Materialize a timestamp tie, then reconstruct from a different submitted order.
	const prefix = "2026-01-01T00:00:00"
	var ties map[string]any
	if err := json.Unmarshal(reverseInput, &ties); err != nil {
		t.Fatal(err)
	}
	ties["source_relationships"] = []any{}
	rows := []any{}
	for _, entry := range []struct{ id, spelling string }{{"a", ".0Z"}, {"b", ".000000Z"}, {"c", "Z"}} {
		rows = append(rows, map[string]any{"source_entity_kind": "kind-a", "source_entity_id": entry.id, "labels": []any{}, "metadata": map[string]any{}, "properties": map[string]any{"group": "one", "instant": prefix + entry.spelling}})
	}
	ties["source_entities"] = rows
	ties["projection_config"].(map[string]any)["aggregation_rules"] = []any{map[string]any{"aggregation_rule_id": "time", "target_scope": "vertex", "input_scope": "source_entity", "input_kind": "kind-a", "projected_kind": "summary", "grouping_keys": []any{"properties.group"}, "missing_grouping_key_behavior": "error", "property_merge_behavior": map[string]any{}}}
	defs := []any{}
	for _, op := range []string{"min", "max"} {
		defs = append(defs, map[string]any{"property_definition_id": op, "target_scope": "vertex", "target_kind": "summary", "projected_key": op, "projected_type": "timestamp", "required": true, "missing_behavior": "error", "source_null_behavior": "error", "null_output_policy": "omit", "merge_behavior": op, "source_field_path": "properties.instant"})
	}
	ties["property_definitions"] = defs
	original, err := json.Marshal(ties)
	if err != nil {
		t.Fatal(err)
	}
	extrema, err := graphprojection.ProjectV2(t.Context(), view, original)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, v := range extrema.Vertices {
		if v.VertexKind == "summary" {
			found = true
			if v.Properties["min"] != prefix+".0Z" || v.Properties["max"] != prefix+".0Z" {
				t.Fatal("tie lost first canonical spelling", v.Properties)
			}
		}
	}
	if !found {
		t.Fatal("missing extrema")
	}
	tied, err := extrema.CompletedResult()
	if err != nil {
		t.Fatal(err)
	}
	slices.Reverse(rows)
	shuffled, err := json.Marshal(ties)
	if err != nil {
		t.Fatal(err)
	}
	return shuffled, tied
}
