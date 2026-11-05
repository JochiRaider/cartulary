package postgresresult_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	_ "embed"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"slices"

	"github.com/jackc/pgx/v5/pgconn"
	"maps"
	"strings"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/graphprojection"
	"github.com/JochiRaider/cartulary/internal/modules/graphprojection/postgresresult"
	"github.com/jackc/pgx/v5"
)

//go:embed testdata/result-vectors-v2.json
var provenanceVectors []byte

//go:embed testdata/shared-reverse-input.json
var sharedReverseInput []byte

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
	// This is valid graph metadata in the current envelope. Only the content binding is stale.
	value["mapped_metadata"] = map[string]any{"tampered": "changed without a new identity"}
	value["schema_registry"].(map[string]any)["metadata_keys"] = []any{map[string]any{
		"target_scope": "graph_view", "target_kind": "*", "projected_metadata_key": "tampered", "projected_type": "string",
		"required": true, "nullable_output": false, "missing_behavior": "error", "source_null_behavior": "error",
	}}
	changed, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return changed
}

func checkCurrentResultProofs(t *testing.T, ctx context.Context, db interface {
	BeginTx(context.Context, pgx.TxOptions) (pgx.Tx, error)
}) {
	var vectors map[string]graphprojection.CompletedResultV2
	if err := json.Unmarshal(provenanceVectors, &vectors); err != nil {
		t.Fatal(err)
	}
	t.Run("stored envelope byte boundary", func(t *testing.T) {
		tx, err := db.BeginTx(ctx, pgx.TxOptions{})
		if err != nil {
			t.Fatal(err)
		}
		defer tx.Rollback(ctx)
		r := vectors["direct"]
		r.PublishedAt = time.Date(2026, 11, 4, 0, 0, 0, 0, time.UTC)
		seedStoredResult(t, ctx, tx, r)
		reader, err := postgresresult.NewReader(tx)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := reader.WithResultLimitForTest(len(r.ResultJSON)).ReadExactResult(ctx, r.Binding); err != nil {
			t.Fatal(err)
		}
		assertInvalidStoredResult(t, ctx, reader.WithResultLimitForTest(len(r.ResultJSON)-1), r)
	})
	t.Run("stored child bytes exceed envelope", func(t *testing.T) {
		for _, scope := range []string{"vertex", "edge"} {
			t.Run(scope, func(t *testing.T) {
				tx, err := db.BeginTx(ctx, pgx.TxOptions{})
				if err != nil {
					t.Fatal(err)
				}
				defer tx.Rollback(ctx)
				r := vectors["direct"]
				r.PublishedAt = time.Date(2026, 11, 4, 0, 0, 0, 0, time.UTC)
				r.Edges = slices.Clone(r.Edges)
				r.Vertices = slices.Clone(r.Vertices)
				excess := bytes.Repeat([]byte("x"), len(r.ResultJSON)+1)
				if scope == "edge" {
					r.Edges[0].JSON = excess
				} else {
					r.Vertices[0].JSON = excess
				}
				seedStoredResult(t, ctx, tx, r)
				reader, err := postgresresult.NewReader(tx)
				if err != nil {
					t.Fatal(err)
				}
				assertInvalidStoredResult(t, ctx, reader, r)
			})
		}
	})
	bad := map[string]graphprojection.CompletedResultV2{"null-entity": vectors["null-entity"], "null-relationship": vectors["null-relationship"]}
	bad["undeclared graph metadata"] = reboundStoredResult(t, vectors["empty"], func(e map[string]any) {
		e["mapped_metadata"] = map[string]any{"review_probe": "x"}
	})
	bad["duplicate direct source"] = reboundStoredResult(t, vectors["direct"], func(e map[string]any) {
		vertices := e["vertices"].([]any)
		v := maps.Clone(fixtureRow(t, e, "vertices", "vertex_id", "vx_9c4e4d9612a63ec17a81250e411c24aa75708474e1a10e1a5925aa5ec1668961"))
		for _, field := range []string{"metadata", "source_entity_ref"} {
			row := maps.Clone(v[field].(map[string]any))
			row["mapping_rule_id"] = "map-other"
			v[field] = row
		}
		// Independently framed for map-other. This key sorts after both originals.
		v["vertex_id"] = "vx_ca821420bee7208766b22b6e235041c354819796e00cc763b632822098e37cbd"
		v["sort_key"] = "sk_40c45aa2ad0b93240d9fd4f4c9657a65a25d213b4f9aa1b2efade919fe3d298b"
		e["vertices"] = append(vertices, v)
	})
	bad["required score omitted"] = reboundStoredResult(t, vectors["aggregate"], func(e map[string]any) {
		v := e["vertices"].([]any)[0].(map[string]any)
		delete(v["properties"].(map[string]any), "score")
	})
	bad["empty object label"] = reboundStoredResult(t, vectors["direct"], func(e map[string]any) {
		e["vertices"].([]any)[0].(map[string]any)["labels"] = []any{""}
	})
	bad["duplicate direct relationship"] = reboundStoredResult(t, vectors["direct"], func(e map[string]any) {
		edges := e["edges"].([]any)
		row := maps.Clone(edges[0].(map[string]any))
		for _, name := range []string{"metadata", "source_relationship_ref"} {
			ref := maps.Clone(row[name].(map[string]any))
			ref["mapping_rule_id"] = "relation-other"
			row[name] = ref
		}
		// Independent GPEDGE1 and GPSORT1 transcripts for the second mapping.
		row["edge_id"] = "ed_e1e2db5e42b7164ee1e8c9bb865b02237a34dfb8c1ea0aed4288c7f23b750c4a"
		row["sort_key"] = "sk_06758658c5cfc3d0983ccb711af83401230c508b47f9c16a4af7e16f17e54752"
		e["edges"] = append([]any{row}, edges...)
	})
	shared := func(t *testing.T, input []byte) graphprojection.CompletedResultV2 {
		t.Helper()
		r, err := graphprojection.ProjectV2(ctx, graphprojection.InvocationContextV2{GraphViewID: "mixed-graph", SourceOwnerID: "network_flow_activity"}, input)
		if err != nil {
			t.Fatal(err)
		}
		result, err := r.CompletedResult()
		if err != nil {
			t.Fatal(err)
		}
		return result
	}
	paired := shared(t, sharedReverseInput)
	if paired.Binding.ProjectionResultID != "gpres_31d2ebc7412ef6e6eaf67a99b731c9d64d8136555a53a8e489941f44893d2efa" || paired.Validate(ctx) != nil {
		t.Fatal("shared reverse control")
	}
	bad["divergent paired labels"] = reboundStoredResult(t, paired, func(e map[string]any) {
		fixtureRow(t, e, "edges", "edge_family", "reverse")["labels"] = []any{}
	})
	bad["fabricated reverse ID"] = reboundStoredResult(t, vectors["reverse"], func(e map[string]any) {
		fixtureRow(t, e, "edges", "edge_family", "reverse")["edge_id"] = "ed_" + strings.Repeat("0", 64)
	})
	bad["retired envelope member"] = reboundStoredResult(t, vectors["empty"], func(e map[string]any) { e["properties"] = map[string]any{} })
	bad["retired property scope"] = reboundStoredResult(t, vectors["empty"], func(e map[string]any) {
		e["schema_registry"].(map[string]any)["property_keys"] = []any{map[string]any{"target_scope": "graph_view", "target_kind": "*", "projected_key": "probe", "projected_type": "integer", "required": true, "nullable_output": true, "missing_behavior": "error", "source_null_behavior": "error"}}
	})
	bad["unreachable graph null"] = reboundStoredResult(t, vectors["empty"], func(e map[string]any) {
		e["mapped_metadata"] = map[string]any{"probe": nil}
		e["schema_registry"].(map[string]any)["metadata_keys"] = []any{map[string]any{"target_scope": "graph_view", "target_kind": "*", "projected_metadata_key": "probe", "projected_type": "integer", "required": true, "nullable_output": true, "missing_behavior": "error", "source_null_behavior": "error"}}
	})
	bad["producerless kind"] = reboundStoredResult(t, vectors["empty"], func(e map[string]any) {
		e["schema_registry"].(map[string]any)["vertex_kinds"] = []any{map[string]any{"vertex_kind": "ghost", "source_entity_kinds": []any{}, "aggregation_rule_ids": []any{}, "labels": []any{}, "source_labels_preserved": false, "properties": []any{}}}
	})
	t.Run("publish full label set", func(t *testing.T) {
		var input map[string]any
		if err := json.Unmarshal(sharedReverseInput, &input); err != nil {
			t.Fatal(err)
		}
		labels := make([]string, 255)
		for i := range labels {
			labels[i] = fmt.Sprintf("label-%03d", i)
		}
		config := input["projection_config"].(map[string]any)
		config["default_edge_labels"] = labels
		// A second, idle producer widens the reverse kind registry beyond 256.
		config["relationship_mappings"].([]any)[1].(map[string]any)["mapping_labels"] = []string{"configured", "another"}
		raw, err := json.Marshal(input)
		if err != nil {
			t.Fatal(err)
		}
		result := shared(t, raw)
		result.PublishedAt = time.Date(2026, 11, 4, 0, 0, 0, 0, time.UTC)
		tx, err := db.BeginTx(ctx, pgx.TxOptions{})
		if err != nil {
			t.Fatal(err)
		}
		defer tx.Rollback(ctx)
		pub, err := postgresresult.NewPublisher(tx)
		if err != nil {
			t.Fatal(err)
		}
		if err := pub.PublishResult(ctx, result); err != nil {
			t.Fatal(err)
		}
		reader, err := postgresresult.NewReader(tx)
		if err != nil {
			t.Fatal(err)
		}
		got, err := reader.ReadExactResult(ctx, result.Binding)
		if err != nil || !bytes.Equal(got.ResultJSON, result.ResultJSON) || !reflect.DeepEqual(got.Edges, result.Edges) {
			t.Fatalf("256-label read: %v", err)
		}
	})
	bad["undeclared raw source kind"] = reboundStoredResult(t, vectors["aggregate"], func(e map[string]any) {
		meta := fixtureRow(t, e, "vertices", "vertex_kind", "endpoints")["metadata"].(map[string]any)
		meta["aggregation_source_refs"] = []any{map[string]any{"ref_kind": "source_entity", "ref_kind_name": "unconfigured", "ref_id": "raw-only", "contributor_sort_key": "csk_756e636f6e66696775726564.7261772d6f6e6c79.."}}
	})
	bad["contradictory scope defaults"] = reboundStoredResult(t, vectors["aggregate"], func(e map[string]any) {
		fixtureRow(t, e["schema_registry"].(map[string]any), "vertex_kinds", "vertex_kind", "endpoints")["labels"] = []any{"invented-default"}
	})
	bad["obsolete direction union"] = reboundStoredResult(t, vectors["reverse"], func(e map[string]any) {
		fixtureRow(t, e["schema_registry"].(map[string]any), "edge_kinds", "edge_kind", "edge")["directions"] = []any{"directed", "undirected", "bidirectional"}
	})
	bad["contradictory raw contributor"] = reboundStoredResult(t, vectors["aggregate"], func(e map[string]any) {
		meta := fixtureRow(t, e, "vertices", "vertex_kind", "endpoints")["metadata"].(map[string]any)
		meta["aggregation_source_refs"] = []any{map[string]any{"ref_kind": "source_entity", "ref_kind_name": "kind-b", "ref_id": "entity-a", "contributor_sort_key": "csk_6b696e642d62.656e746974792d61.."}}
	})
	bad["arbitrary aggregate sort key"] = reboundStoredResult(t, vectors["aggregate"], func(e map[string]any) {
		edges := e["edges"].([]any)
		edges[1].(map[string]any)["sort_key"] = "a"
		edges[0], edges[1] = edges[1], edges[0]
	})
	bad["aggregation-only preserves labels"] = reboundStoredResult(t, vectors["aggregate"], func(e map[string]any) {
		fixtureRow(t, e["schema_registry"].(map[string]any), "vertex_kinds", "vertex_kind", "endpoints")["source_labels_preserved"] = true
	})

	bad["inconsistent preservation"] = reboundStoredResult(t, vectors["reverse"], func(e map[string]any) {
		fixtureRow(t, e["schema_registry"].(map[string]any), "edge_kinds", "edge_kind", "back")["source_labels_preserved"] = true
	})
	bad["same-kind mapped fields"] = reboundStoredResult(t, vectors["reverse"], func(e map[string]any) {
		registry := e["schema_registry"].(map[string]any)
		registry["edge_kinds"] = []any{fixtureRow(t, registry, "edge_kinds", "edge_kind", "edge")}
		registry["metadata_keys"] = []any{map[string]any{"target_scope": "edge", "target_kind": "edge", "projected_metadata_key": "probe", "projected_type": "string", "nullable_output": false, "required": true, "missing_behavior": "error", "source_null_behavior": "error"}}
		forward := fixtureRow(t, e, "edges", "edge_family", "direct")
		reverse := fixtureRow(t, e, "edges", "edge_family", "reverse")
		// Independently framed GPEDGE1 and GPSORT1 transcripts for reverse kind "edge".
		forward["edge_id"], forward["sort_key"] = "ed_cb7fdbc4588bd4fa05a393b866df938f65b3d22a3be2d0fc1ff4633251989e9b", "sk_601d1e17d0dedb57ef9b3dc375552c012e2990a528441ded66cc98a7ef09e58c"
		reverse["edge_kind"], reverse["edge_id"], reverse["sort_key"] = "edge", "ed_d857db1e878a165d47f094c9611b889d2e25c85b8c6f8b2250249ce3991e7f2c", "sk_c9f3d03874c7ee2da5e19b3f57dc918cf5ce77853070ea2a1b330023a96c1555"
		reverse["metadata"].(map[string]any)["reverse_of_edge_id"] = forward["edge_id"]
		forward["metadata"].(map[string]any)["mapped_metadata"] = map[string]any{"probe": "one"}
		reverse["metadata"].(map[string]any)["mapped_metadata"] = map[string]any{"probe": "two"}
		e["edges"] = []any{forward, reverse}
	})
	for name, id := range map[string]string{
		"same-kind mapped fields":           "gpres_cd8fdf8804cb1a183d81dbefbf9c6c9c1d8fb695da8c9df203a079a20b9540ea",
		"inconsistent preservation":         "gpres_822e9620e75545111344ff03faecdec6f8707d423d25fd3706e1f2cd7d0f02c8",
		"undeclared raw source kind":        "gpres_9a36d1988dcbbfa0d0fc176c56f50387a5210c5a2f936d4b03b7df82df29ea04",
		"contradictory scope defaults":      "gpres_1dd74cf06a286996fb248a7b6dc8b93a705da2dc4ad819aabb4c39fb059ce3a3",
		"obsolete direction union":          "gpres_ff7b37ae821e9c62b8b0b2d471bcd2a30ee2ed7a830f0f18148ee42b849ae350",
		"contradictory raw contributor":     "gpres_bd259a07cd2af8f1c8824fa2810f72d3304f0a3d2a6223fa3e72517cdbf8255a",
		"arbitrary aggregate sort key":      "gpres_f2ca5db1456cb302866057c21bfefb956b452fe334d5589a8be044b9492beac3",
		"aggregation-only preserves labels": "gpres_e2a80ed909ebf99b692faf17bf5e3786b5a76a67d1143d936c5ab7e597e731ab",
		"fabricated reverse ID":             "gpres_74e123a0b38f016318a5b8a5d535499baf7586adcbb97bafd91db7ddc69d3013",
		"divergent paired labels":           "gpres_31b408984613d965c4ecc2874baf8ecd4c325be6e3fbc191d4f271bee23d7456",
		"unreachable graph null":            "gpres_d0c96559d538afff4d5d77cb09406e76f8766af0b43b4b789740ad52f7f2492e",
		"producerless kind":                 "gpres_55548c8036eb1b4d4bf0cf85ea4e3ee9de0528fdb180a0f4daa8377e7049ce77",
		"required score omitted":            "gpres_ef6d655120013924c251dd39f6b2c821af2e70ea278e6a80e67ec8f6602c471c",
		"duplicate direct relationship":     "gpres_45ac6817d6403e57220b0774f3097fda619fd1937b5a527136670b43027c8f84",
		"empty object label":                "gpres_c562ac8bdc6c66c007cda2257b143351cccfcf9b9ec78860459d8fd1c5b7ccb1",
		"undeclared graph metadata":         "gpres_53c892812e2de7aaf9e697a0fdee0b80cfa4bbb3b3beee9281265cb0d35dd9c9",
		"duplicate direct source":           "gpres_e8ea0a89cbe90e7a0b42359d6a56c23ebee8c47793f8d7f6474ffb947cb5acfa",
	} {
		if bad[name].Binding.ProjectionResultID != id {
			t.Fatalf("%s rebound binding: %s", name, bad[name].Binding.ProjectionResultID)
		}
	}
	bad["direction outside own kind"] = reboundStoredResult(t, vectors["direct"], func(e map[string]any) {
		k := fixtureRow(t, e["schema_registry"].(map[string]any), "edge_kinds", "edge_kind", "edge")
		k["directions"] = []any{"bidirectional"}
	})
	bad["invented producer"] = reboundStoredResult(t, vectors["aggregate"], func(e map[string]any) {
		fixtureRow(t, e, "edges", "edge_kind", "tail")["metadata"].(map[string]any)["aggregation_rule_id"] = "invented"
	})
	bad["endpoint cycle"] = reboundStoredResult(t, vectors["aggregate"], func(e map[string]any) {
		fixtureRow(t, e, "edges", "edge_kind", "bridge")["src_vertex_id"] = fixtureRow(t, e, "vertices", "vertex_kind", "final")["vertex_id"]
	})
	bad["overlapping groups"] = reboundStoredResult(t, vectors["aggregate"], func(e map[string]any) {
		group := maps.Clone(fixtureRow(t, e, "vertices", "vertex_kind", "endpoints"))
		group["vertex_id"] = "vx_" + strings.Repeat("1", 64)
		group["sort_key"] = "sk_" + strings.Repeat("f", 64)
		e["vertices"] = append(e["vertices"].([]any), group)
	})
	bad["cycle through another producer of the input kind"] = reboundStoredResult(t, vectors["aggregate"], func(e map[string]any) {
		registry := e["schema_registry"].(map[string]any)
		fixtureRow(t, registry, "edge_kinds", "edge_kind", "bridge")["aggregation_rule_ids"] = []any{"bridge", "tail"}
		fixtureRow(t, registry, "edge_kinds", "edge_kind", "tail")["aggregation_rule_ids"] = []any{"unused"}
		fixtureRow(t, e, "edges", "edge_kind", "tail")["edge_kind"] = "bridge"
	})
	bad["substituted vertex source"] = reboundStoredResult(t, vectors["direct"], func(e map[string]any) {
		v := fixtureRow(t, e, "vertices", "vertex_id", "vx_9c4e4d9612a63ec17a81250e411c24aa75708474e1a10e1a5925aa5ec1668961")
		v["source_entity_ref"].(map[string]any)["source_entity_id"] = "other-entity"
	})
	bad["incorrect direct vertex sort key"] = reboundStoredResult(t, vectors["direct"], func(e map[string]any) {
		e["vertices"].([]any)[0].(map[string]any)["sort_key"] = "sk_" + strings.Repeat("0", 64)
	})
	for _, family := range []string{"direct", "reverse"} {
		for _, side := range []string{"src_vertex_id", "dst_vertex_id"} {
			bad[family+" edge to aggregate "+side] = reboundStoredResult(t, vectors[family], func(e map[string]any) {
				aggregate := maps.Clone(e["vertices"].([]any)[0].(map[string]any))
				aggregate["vertex_id"] = "vx_" + strings.Repeat("f", 64)
				aggregate["vertex_family"] = "aggregated"
				aggregate["sort_key"] = "sk_" + strings.Repeat("f", 64)
				aggregate["source_entity_ref"] = nil
				aggregate["metadata"] = map[string]any{"mapped_metadata": map[string]any{}, "mapping_rule_id": nil, "aggregation_rule_id": "group", "aggregation_source_refs": []any{
					map[string]any{"ref_kind": "source_entity", "ref_kind_name": "a", "ref_id": "b", "contributor_sort_key": "csk_61.62.."},
				}}
				e["vertices"] = append(e["vertices"].([]any), aggregate)
				e["schema_registry"].(map[string]any)["vertex_kinds"].([]any)[0].(map[string]any)["aggregation_rule_ids"] = []any{"group"}
				forward := fixtureRow(t, e, "edges", "edge_family", "direct")
				forward[side] = aggregate["vertex_id"]
				if family == "reverse" {
					reverse := fixtureRow(t, e, "edges", "edge_family", "reverse")
					reverse["src_vertex_id"], reverse["dst_vertex_id"] = forward["dst_vertex_id"], forward["src_vertex_id"]
				}
			})
		}
	}
	for name, mutate := range map[string]func(map[string]any){
		"null property": func(r map[string]any) { r["property_keys"] = []any{nil} },
		"invalid metadata declaration": func(r map[string]any) {
			r["metadata_keys"] = []any{map[string]any{"target_scope": "vertex", "target_kind": "vertex", "projected_metadata_key": "test"}}
		},
		"missing direction list":           func(r map[string]any) { delete(r["edge_kinds"].([]any)[0].(map[string]any), "directions") },
		"unknown kind member":              func(r map[string]any) { r["vertex_kinds"].([]any)[0].(map[string]any)["unknown"] = true },
		"contradictory property reference": func(r map[string]any) { r["property_keys"].([]any)[0].(map[string]any)["required"] = false },
		"invalid declaration policy": func(r map[string]any) {
			r["property_keys"].([]any)[0].(map[string]any)["missing_behavior"] = "invented"
		},
	} {
		bad[name] = reboundStoredResult(t, vectors["aggregate"], func(e map[string]any) { mutate(e["schema_registry"].(map[string]any)) })
	}
	for name, result := range bad {
		t.Run(name, func(t *testing.T) {
			tx, err := db.BeginTx(ctx, pgx.TxOptions{})
			if err != nil {
				t.Fatal(err)
			}
			defer tx.Rollback(ctx)
			if err := result.Validate(ctx); !errors.Is(err, graphprojection.ErrResultV2Invalid) {
				t.Errorf("pure proof accepted malformed result: %v", err)
			}
			result.PublishedAt = time.Date(2026, 11, 3, 0, 0, 0, 0, time.UTC)
			observed := &observedPublicationTx{Tx: tx}
			publisher, err := postgresresult.NewPublisher(observed)
			if err != nil {
				t.Fatal(err)
			}
			if err := publisher.PublishResult(ctx, result); !errors.Is(err, graphprojection.ErrResultV2Invalid) {
				t.Fatalf("malformed provenance publication: %v", err)
			}
			if observed.executions != 0 {
				t.Fatalf("invalid publication attempted %d SQL executions", observed.executions)
			}
			var count int
			if err := tx.QueryRow(ctx, `SELECT
				(SELECT count(*) FROM graph_projection_results WHERE projection_result_id=$1) +
				(SELECT count(*) FROM graph_projection_result_vertices WHERE projection_result_id=$1) +
				(SELECT count(*) FROM graph_projection_result_edges WHERE projection_result_id=$1)`, result.Binding.ProjectionResultID).Scan(&count); err != nil || count != 0 {
				t.Fatalf("rejected result wrote %d rows: %v", count, err)
			}
			seedStoredResult(t, ctx, tx, result)
			reader, err := postgresresult.NewReader(tx)
			if err != nil {
				t.Fatal(err)
			}
			assertInvalidStoredResult(t, ctx, reader, result)
		})
	}
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
		assertInvalidStoredResult(t, ctx, reader, result)
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

type observedPublicationTx struct {
	pgx.Tx
	executions int
}

func (tx *observedPublicationTx) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	tx.executions++
	return tx.Tx.Exec(ctx, sql, args...)
}

func (tx *observedPublicationTx) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	tx.executions++
	return tx.Tx.Query(ctx, sql, args...)
}
func (tx *observedPublicationTx) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	tx.executions++
	return tx.Tx.QueryRow(ctx, sql, args...)
}

func fixtureRow(t *testing.T, root map[string]any, collection, key, value string) map[string]any {
	t.Helper()
	for _, raw := range root[collection].([]any) {
		row := raw.(map[string]any)
		if row[key] == value {
			return row
		}
	}
	t.Fatalf("fixture lacks %s %s=%s", collection, key, value)
	return nil
}

// Rebind envelope, child bytes and indexed columns independently of production.
func reboundStoredResult(t *testing.T, result graphprojection.CompletedResultV2, mutate func(map[string]any)) graphprojection.CompletedResultV2 {
	t.Helper()
	var envelope map[string]any
	decoder := json.NewDecoder(bytes.NewReader(result.ResultJSON))
	decoder.UseNumber()
	if err := decoder.Decode(&envelope); err != nil {
		t.Fatal(err)
	}
	encode := func(v any) []byte {
		b, err := json.Marshal(v)
		if err != nil {
			t.Fatal(err)
		}
		return b
	}
	mutate(envelope)
	result.Vertices = nil
	result.Edges = nil
	for _, raw := range envelope["vertices"].([]any) {
		v := raw.(map[string]any)
		result.Vertices = append(result.Vertices, graphprojection.ResultVertexV2{VertexID: v["vertex_id"].(string), VertexKind: v["vertex_kind"].(string), SortKey: v["sort_key"].(string), JSON: encode(v)})
	}
	for _, raw := range envelope["edges"].([]any) {
		e := raw.(map[string]any)
		result.Edges = append(result.Edges, graphprojection.ResultEdgeV2{EdgeID: e["edge_id"].(string), EdgeKind: e["edge_kind"].(string), SrcVertexID: e["src_vertex_id"].(string), DstVertexID: e["dst_vertex_id"].(string), Direction: e["direction"].(string), SortKey: e["sort_key"].(string), JSON: encode(e)})
	}
	output := maps.Clone(envelope)
	for _, key := range []string{"projection_schema_id", "projection_result_id", "graph_view_id", "source_owner_id", "source_snapshot_id", "projection_version", "normalized_configuration_sha256", "normalized_source_sha256", "canonical_output_sha256"} {
		delete(output, key)
	}
	digest := sha256.Sum256(encode(output))
	b := &result.Binding
	b.CanonicalOutputSHA256 = hex.EncodeToString(digest[:])
	var transcript []byte
	for _, field := range []string{"cartulary.graph_projection_result_identity.v2", b.GraphViewID, b.SourceOwnerID, b.SourceSnapshotID, b.ProjectionVersion, b.NormalizedConfigurationSHA256, b.NormalizedSourceSHA256, b.CanonicalOutputSHA256} {
		transcript = binary.BigEndian.AppendUint64(transcript, uint64(len(field)))
		transcript = append(transcript, field...)
	}
	id := sha256.Sum256(transcript)
	b.ProjectionResultID = "gpres_" + hex.EncodeToString(id[:])
	envelope["projection_result_id"] = b.ProjectionResultID
	envelope["canonical_output_sha256"] = b.CanonicalOutputSHA256
	result.ResultJSON = encode(envelope)
	return result
}

func seedStoredResult(t *testing.T, ctx context.Context, tx pgx.Tx, result graphprojection.CompletedResultV2) {
	t.Helper()
	b := result.Binding
	// Import the existing rows verbatim, without executing the candidate publisher.
	_, err := tx.Exec(ctx, `INSERT INTO graph_projection_results (projection_result_id, graph_view_id,source_owner_id,source_snapshot_id,projection_schema_id,projection_version,normalized_configuration_sha256,normalized_source_sha256,canonical_output_sha256,vertex_count,edge_count,result_json,published_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)`, b.ProjectionResultID, b.GraphViewID, b.SourceOwnerID, b.SourceSnapshotID, b.ProjectionSchemaID, b.ProjectionVersion, b.NormalizedConfigurationSHA256, b.NormalizedSourceSHA256, b.CanonicalOutputSHA256, len(result.Vertices), len(result.Edges), []byte(result.ResultJSON), result.PublishedAt)
	if err != nil {
		t.Fatal(err)
	}
	for i, row := range result.Vertices {
		if _, err := tx.Exec(ctx, `INSERT INTO graph_projection_result_vertices (projection_result_id,vertex_id,vertex_kind,sort_ordinal,sort_key,vertex_json) VALUES ($1,$2,$3,$4,$5,$6)`, b.ProjectionResultID, row.VertexID, row.VertexKind, i, row.SortKey, []byte(row.JSON)); err != nil {
			t.Fatal(err)
		}
	}
	for i, row := range result.Edges {
		if _, err := tx.Exec(ctx, `INSERT INTO graph_projection_result_edges
			(projection_result_id,edge_id,edge_kind,src_vertex_id,dst_vertex_id,direction,sort_ordinal,sort_key,edge_json)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`, b.ProjectionResultID, row.EdgeID, row.EdgeKind, row.SrcVertexID, row.DstVertexID, row.Direction, i, row.SortKey, []byte(row.JSON)); err != nil {
			t.Fatal(err)
		}
	}
}

func assertInvalidStoredResult(t *testing.T, ctx context.Context, reader *postgresresult.Reader, result graphprojection.CompletedResultV2) {
	t.Helper()
	if binding, err := reader.LockResultEnvelope(ctx, result.Binding.ProjectionResultID); err != nil || binding != result.Binding {
		t.Fatalf("metadata lock changed content-admission ordering: %v", err)
	}
	seed := "vx_" + strings.Repeat("0", 64)
	if len(result.Vertices) > 0 {
		seed = result.Vertices[0].VertexID
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
			_, err := reader.Traverse(ctx, graphprojection.TraversalRequestV2{ProjectionResultID: result.Binding.ProjectionResultID, SeedVertexIDs: []string{seed}, Direction: graphprojection.TraversalOutgoingV2, MaximumDepth: 1, MaximumVertices: 2, MaximumEdges: 1})
			return err
		},
	} {
		if err := read(); !errors.Is(err, graphprojection.ErrResultV2Invalid) {
			t.Errorf("%s returned divergent stored row: %v", name, err)
		}
	}
}
