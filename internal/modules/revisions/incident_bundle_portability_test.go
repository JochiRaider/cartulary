package revisions_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	authstoretest "github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/storetest"
	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
	"github.com/JochiRaider/cartulary/internal/modules/incidentportability"
	"github.com/JochiRaider/cartulary/internal/modules/records"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/postgres/dbport"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/portabilitysupport"
)

const (
	revisionsTestChangeSetsPath = "data/change_sets.ndjson"
	revisionsTestMutationsPath  = "data/change_set_mutations.ndjson"
	revisionsTestHistoryPath    = "data/record_revisions.ndjson"
)

type revisionsPortabilityHarness struct {
	db         dbport.DB
	actor      authn.UserRecord
	incidentID uuid.UUID
	recordID   uuid.UUID
	changeSet  uuid.UUID
	createdAt  time.Time
	snapshot   map[string]any
	port       sourceport.Port
	context    sourceport.ImportContext
}

type revisionsAttributionRecorder struct {
	rows       []incidentportability.ImportedAttribution
	localActor uuid.UUID
	recordErr  error
	visitErr   error
}

type revisionsPortableAttributionResolver struct {
	rows []incidentportability.ImportedAttribution
}

type revisionsPortabilityEnvelopeReader struct{}

func (revisionsPortabilityEnvelopeReader) VisitIncidentRecordIDsTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, visit func(uuid.UUID) error) error {
	return records.NewStore().VisitIncidentRecordIDsTx(ctx, tx, incidentID, visit)
}

func (revisionsPortabilityEnvelopeReader) RecordTypeTx(
	ctx context.Context,
	tx pgx.Tx,
	incidentID uuid.UUID,
	recordID uuid.UUID,
) (string, error) {
	var recordType string
	err := tx.QueryRow(ctx, `
SELECT record_type
  FROM records
 WHERE incident_id = $1
   AND record_id = $2
`, incidentID, recordID).Scan(&recordType)
	return recordType, err
}

func (r revisionsPortableAttributionResolver) ResolvePortableSourceActors(
	_ context.Context,
	_ incidentportability.Queryer,
	_ uuid.UUID,
	table string,
	column string,
	rowIDs []string,
) (map[string]string, error) {
	wanted := make(map[string]struct{}, len(rowIDs))
	for _, rowID := range rowIDs {
		wanted[rowID] = struct{}{}
	}
	result := map[string]string{}
	for _, row := range r.rows {
		if row.SourceTable != table || row.SourceColumn != column {
			continue
		}
		if _, ok := wanted[row.SourceRowID]; ok {
			result[row.SourceRowID] = row.SourceActorID
		}
	}
	return result, nil
}

func (r *revisionsAttributionRecorder) RecordImportedAttribution(
	table string,
	rowID string,
	column string,
	actorID string,
) error {
	if r.recordErr != nil {
		return r.recordErr
	}
	r.rows = append(r.rows, incidentportability.ImportedAttribution{
		SourceTable: table, SourceRowID: rowID, SourceColumn: column,
		SourceActorID: actorID, LocalUserID: r.localActor,
	})
	return nil
}

func (r *revisionsAttributionRecorder) ImportedAttributions() []incidentportability.ImportedAttribution {
	return append([]incidentportability.ImportedAttribution(nil), r.rows...)
}

func TestRevisionsIncidentBundleInvariantAttribution(t *testing.T) {
	t.Run("historical_link_reference_scope", exerciseHistoricalLinksReferenceScope)
	t.Run("snapshot_version_agreement", exerciseRevisionSnapshotVersionAgreement)
	t.Run("complete_row_history", exerciseCompleteRowHistory)
	t.Run("actor_import_contexts", exerciseRevisionActorContextChains)
	t.Run("committed_history_order", exercisePortableHistoryOrder)
	t.Run("row_mutation_requires_after_state", func(t *testing.T) {
		h := newRevisionsPortabilityHarness(t, "row-after-state")
		bundle := h.validBundle(t)
		rows := decodePortableRows(t, bundle[revisionsTestMutationsPath])
		rows[0]["operation_kind"], rows[0]["before_value"], rows[0]["after_value"] = "patch", h.snapshot, nil
		rows[0]["before_version_id"], rows[0]["after_version_id"] = "host:"+h.recordID.String()+":1", nil
		bundle[revisionsTestMutationsPath] = encodePortableRows(t, rows)
		requireRevisionsInvariant(t, h.applyAndValidate(t, bundle), "revisions.history_reconstruction")
	})
	t.Run("incident_history_coverage", exerciseIncidentHistoryCoverage)
	t.Run("complete_mention_history", exerciseCompleteMentionHistory)
	t.Run("complete_preserved_identifier_history", exerciseCompletePreservedIdentifierHistory)
	t.Run("complete_alias_history", exerciseCompleteAliasHistory)
	t.Run("complete_interval_history", exerciseCompleteIntervalHistory)
	t.Run("complete_observation_history", exerciseCompleteObservationHistory)
	t.Run("nonrow_reference_scope", exerciseNonRowHistoryReferenceScope)
	t.Run("source_capture_retains_bigint_versions", func(t *testing.T) {
		h := newRevisionsPortabilityHarness(t, "bigint-version")
		for _, want := range []json.Number{"9007199254740993", "9223372036854775807"} {
			t.Run(string(want), func(t *testing.T) {
				tx, err := h.db.BeginTx(t.Context(), pgx.TxOptions{})
				if err != nil {
					t.Fatal(err)
				}
				defer func() { _ = tx.Rollback(t.Context()) }()
				version, err := want.Int64()
				if err != nil {
					t.Fatal(err)
				}
				for _, table := range []string{"records", "hosts"} {
					if _, err := tx.Exec(t.Context(), "UPDATE "+table+" SET row_version=$2::bigint WHERE record_id=$1", h.recordID, version); err != nil {
						t.Fatal(err)
					}
				}
				captures := 0
				for _, owner := range mustRevisionProviderContributions(t) {
					for _, record := range owner.Records {
						if record.RecordType != "host" {
							continue
						}
						captures++
						snapshot, err := record.DeleteRestoreSource.SnapshotTx(t.Context(), tx, h.recordID)
						if err != nil {
							t.Fatal(err)
						}
						for _, member := range []string{"record", "source"} {
							value, ok := snapshot[member].(map[string]any)
							if !ok || value["row_version"] != want {
								t.Errorf("%s version = %#v, want exact JSON integer %s", member, value["row_version"], want)
							}
						}
					}
				}
				if captures != 1 {
					t.Fatalf("Host source captures = %d, want 1", captures)
				}
			})
		}
	})
	for _, member := range []struct {
		name  string
		cases []struct {
			name   string
			mutate func(map[string]any)
		}
	}{
		{"record", []struct {
			name   string
			mutate func(map[string]any)
		}{
			{"complete", nil},
			{"missing nullable", func(record map[string]any) { delete(record, "deleted_at") }},
			{"unknown member", func(record map[string]any) { record["projection_label"] = "not source" }},
			{"foreign incident", func(record map[string]any) { record["incident_id"] = "ffffffff-ffff-4fff-8fff-ffffffffffff" }},
			{"mistyped version", func(record map[string]any) { record["row_version"] = "1" }},
			{"nonpositive version", func(record map[string]any) { record["row_version"] = json.Number("0") }},
			{"reversed clocks", func(record map[string]any) { record["updated_at"] = "2000-01-01T00:00:00Z" }},
			{"partial deletion tuple", func(record map[string]any) { record["deleted_at"] = record["updated_at"] }},
		}},
		{"source", []struct {
			name   string
			mutate func(map[string]any)
		}{
			{"complete", nil},
			{"missing nullable", func(source map[string]any) { delete(source, "hostname") }},
			{"unknown member", func(source map[string]any) { source["projection_label"] = "not source" }},
			{"foreign incident", func(source map[string]any) { source["incident_id"] = "ffffffff-ffff-4fff-8fff-ffffffffffff" }},
			{"foreign record", func(source map[string]any) { source["record_id"] = "ffffffff-ffff-4fff-8fff-ffffffffffff" }},
			{"mistyped version", func(source map[string]any) { source["row_version"] = "1" }},
			{"invalid clock", func(source map[string]any) { source["created_at"] = "yesterday" }},
			{"unknown origin", func(source map[string]any) { source["entity_origin"] = "unknown" }},
			{"unknown state", func(source map[string]any) { source["host_state"] = "active" }},
			{"missing merge target", func(source map[string]any) { source["host_state"] = "merged" }},
			{"self merge", func(source map[string]any) {
				source["host_state"] = "merged"
				source["merged_into_record_id"] = source["record_id"]
			}},
			{"nonpositive source version", func(source map[string]any) { source["row_version"] = json.Number("0") }},
			{"different source version", func(source map[string]any) { source["row_version"] = json.Number("3") }},
			{"different source clock", func(source map[string]any) { source["created_at"] = "2000-01-01T00:00:00Z" }},
			{"different source actor", func(source map[string]any) { source["created_by_user_id"] = "ffffffff-ffff-4fff-8fff-ffffffffffff" }},
		}},
	} {
		t.Run("nonterminal_"+member.name+"_is_complete", func(t *testing.T) {
			h := newRevisionsPortabilityHarness(t, "nonterminal-"+member.name)
			ctx := context.Background()
			for _, query := range []string{
				`UPDATE records SET row_version = 2 WHERE record_id = $1`,
				`UPDATE hosts SET row_version = 2 WHERE record_id = $1`,
			} {
				if _, err := h.db.Exec(ctx, query, h.recordID); err != nil {
					t.Fatalf("advance current fixture version: %v", err)
				}
			}
			current := clonePortableRow(t, h.snapshot)
			current["record"].(map[string]any)["row_version"] = json.Number("2")
			current["source"].(map[string]any)["row_version"] = json.Number("2")
			for _, test := range member.cases {
				t.Run(test.name, func(t *testing.T) {
					prior := clonePortableRow(t, h.snapshot)
					if test.mutate != nil {
						test.mutate(prior[member.name].(map[string]any))
					}
					bundle := h.validBundle(t)
					mutations := decodePortableRows(t, bundle[revisionsTestMutationsPath])
					mutations[0]["after_value"] = prior
					second := clonePortableRow(t, mutations[0])
					second["sequence_no"] = 2
					second["operation_kind"] = "patch"
					second["before_version_id"] = second["after_version_id"]
					second["after_version_id"] = "host:" + h.recordID.String() + ":2"
					second["before_value"], second["after_value"] = prior, current
					bundle[revisionsTestMutationsPath] = encodePortableRows(t, append(mutations, second))
					history := decodePortableRows(t, bundle[revisionsTestHistoryPath])
					history[0]["after_json"] = prior
					next := clonePortableRow(t, history[0])
					next["revision_id"], next["row_version"] = 2, 2
					next["before_json"], next["after_json"] = prior, current
					bundle[revisionsTestHistoryPath] = encodePortableRows(t, append(history, next))
					caseHarness := h
					if member.name == "source" && test.name == "different source actor" {
						actors, err := sourceport.NewActorCatalog([]sourceport.ActorDescriptor{
							{SourceActorID: h.actor.ID.String()},
							{SourceActorID: "ffffffff-ffff-4fff-8fff-ffffffffffff"},
						})
						if err != nil {
							t.Fatal(err)
						}
						caseHarness.context.Actors = actors
					}
					err := caseHarness.applyAndValidate(t, bundle)
					if test.mutate == nil {
						if err != nil {
							t.Fatalf("complete nonterminal %s rejected: %v", member.name, err)
						}
					} else {
						requireRevisionsInvariant(t, err, "revisions.history_reconstruction")
					}
				})
			}
		})
	}
	t.Run("row_target_binds_actual_record_owner", func(t *testing.T) {
		h := newRevisionsPortabilityHarness(t, "row-target-owner")
		for _, kind := range []string{"host", "record", "identity"} {
			t.Run(kind, func(t *testing.T) {
				ctx := context.Background()
				input := h.freshImportContext()
				bundle := h.validBundle(t)
				mutations := decodePortableRows(t, bundle[revisionsTestMutationsPath])
				mutations[0]["target_kind"] = kind
				bundle[revisionsTestMutationsPath] = encodePortableRows(t, mutations)
				prepared, err := h.port.PrepareImport(ctx, bundle, input)
				if err != nil {
					t.Fatalf("prepare target-owner fixture: %v", err)
				}
				tx, err := h.db.BeginTx(ctx, pgx.TxOptions{})
				if err != nil {
					t.Fatalf("begin target-owner fixture: %v", err)
				}
				defer func() { _ = tx.Rollback(ctx) }()
				err = h.port.ApplyImportTx(ctx, tx, prepared, input)
				if kind == "identity" {
					requireRevisionsInvariant(t, err, "revisions.references_complete")
					var count int
					if err := tx.QueryRow(ctx, `SELECT count(*) FROM change_sets WHERE change_set_id = $1`, h.changeSet).Scan(&count); err != nil || count != 0 {
						t.Fatalf("wrong target owner inserted %d change sets, err=%v", count, err)
					}
				} else {
					if err != nil {
						t.Fatalf("apply matching target owner: %v", err)
					}
					if err := h.port.ValidateImportTx(ctx, tx, prepared, input); err != nil {
						t.Fatalf("validate matching target owner: %v", err)
					}
				}
			})
		}
	})
	t.Run("row_invariants", func(t *testing.T) {
		harness := newRevisionsPortabilityHarness(t, "invariant-attribution")
		valid := harness.validBundle(t)

		tests := []struct {
			name      string
			invariant string
			mutate    func(portabilitysupport.MapBundle)
		}{
			{
				name:      "references_complete",
				invariant: "revisions.references_complete",
				mutate: func(bundle portabilitysupport.MapBundle) {
					rows := decodePortableRows(t, bundle[revisionsTestHistoryPath])
					rows[0]["record_id"] = uuid.NewString()
					bundle[revisionsTestHistoryPath] = encodePortableRows(t, rows)
				},
			},
			{
				name:      "actor_references_complete",
				invariant: "revisions.actor_references_complete",
				mutate: func(bundle portabilitysupport.MapBundle) {
					rows := decodePortableRows(t, bundle[revisionsTestChangeSetsPath])
					rows[0]["actor_user_id"] = uuid.NewString()
					bundle[revisionsTestChangeSetsPath] = encodePortableRows(t, rows)
				},
			},
			{
				name:      "mutation_sequence_contiguous",
				invariant: "revisions.mutation_sequence_contiguous",
				mutate: func(bundle portabilitysupport.MapBundle) {
					rows := decodePortableRows(t, bundle[revisionsTestMutationsPath])
					rows[0]["sequence_no"] = 2
					bundle[revisionsTestMutationsPath] = encodePortableRows(t, rows)
				},
			},
			{
				name:      "record_version_unique",
				invariant: "revisions.record_version_unique",
				mutate: func(bundle portabilitysupport.MapBundle) {
					rows := decodePortableRows(t, bundle[revisionsTestHistoryPath])
					duplicate := clonePortableRow(t, rows[0])
					duplicate["revision_id"] = 2
					rows = append(rows, duplicate)
					bundle[revisionsTestHistoryPath] = encodePortableRows(t, rows)
				},
			},
			{
				name:      "history_reconstruction",
				invariant: "revisions.history_reconstruction",
				mutate: func(bundle portabilitysupport.MapBundle) {
					rows := decodePortableRows(t, bundle[revisionsTestHistoryPath])
					after := rows[0]["after_json"].(map[string]any)
					after["source"].(map[string]any)["display_name"] = "not-current"
					bundle[revisionsTestHistoryPath] = encodePortableRows(t, rows)
				},
			},
			{
				name:      "schema_less_history_rejected",
				invariant: "revisions.history_reconstruction",
				mutate: func(bundle portabilitysupport.MapBundle) {
					rows := decodePortableRows(t, bundle[revisionsTestHistoryPath])
					delete(rows[0]["after_json"].(map[string]any), "snapshot_schema_id")
					bundle[revisionsTestHistoryPath] = encodePortableRows(t, rows)
				},
			},
			{
				name:      "schema_less_mutation_history_rejected",
				invariant: "revisions.history_reconstruction",
				mutate: func(bundle portabilitysupport.MapBundle) {
					rows := decodePortableRows(t, bundle[revisionsTestMutationsPath])
					delete(rows[0]["after_value"].(map[string]any), "snapshot_schema_id")
					bundle[revisionsTestMutationsPath] = encodePortableRows(t, rows)
				},
			},
		}
		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				bundle := clonePortableBundle(valid)
				test.mutate(bundle)
				err := harness.applyAndValidate(t, bundle)
				requireRevisionsInvariant(t, err, test.invariant)
			})
		}

		descriptor := harness.port.Descriptor()
		if got := descriptor.InvariantIDs[len(descriptor.InvariantIDs)-2:]; got[0] != "revisions.sequence_repair_after_validation" || got[1] != "revisions.source_identity_admitted" {
			t.Fatalf("revisions invariant tail = %#v", got)
		}

		strictFixtures := []struct {
			name   string
			mutate func(portabilitysupport.MapBundle)
		}{
			{
				name: "unknown_member",
				mutate: func(bundle portabilitysupport.MapBundle) {
					rows := decodePortableRows(t, bundle[revisionsTestChangeSetsPath])
					rows[0]["hostile_unknown"] = "must-not-escape"
					bundle[revisionsTestChangeSetsPath] = encodePortableRows(t, rows)
				},
			},
			{
				name: "missing_nullable_member",
				mutate: func(bundle portabilitysupport.MapBundle) {
					rows := decodePortableRows(t, bundle[revisionsTestMutationsPath])
					delete(rows[0], "before_value")
					bundle[revisionsTestMutationsPath] = encodePortableRows(t, rows)
				},
			},
			{
				name: "wrong_integer_type",
				mutate: func(bundle portabilitysupport.MapBundle) {
					rows := decodePortableRows(t, bundle[revisionsTestHistoryPath])
					rows[0]["row_version"] = "1"
					bundle[revisionsTestHistoryPath] = encodePortableRows(t, rows)
				},
			},
		}
		for _, fixture := range strictFixtures {
			t.Run("strict_"+fixture.name, func(t *testing.T) {
				bundle := clonePortableBundle(valid)
				fixture.mutate(bundle)
				_, err := harness.port.PrepareImport(context.Background(), bundle, harness.context)
				if err == nil {
					t.Fatal("malformed Revisions row passed strict preparation")
				}
				if bytes.Contains([]byte(err.Error()), []byte("must-not-escape")) ||
					bytes.Contains([]byte(err.Error()), []byte("hostile_unknown")) {
					t.Fatalf("strict failure exposed hostile input: %v", err)
				}
			})
		}
	})
	t.Run("links_history_is_strict_before_import_mutation", func(t *testing.T) {
		exerciseStrictLinksIncidentBundleAdmission(t)
	})
	t.Run("terminal snapshot is complete", func(t *testing.T) {
		h := newRevisionsPortabilityHarness(t, "complete-terminal")
		for _, test := range []struct {
			name, object, field string
			add                 bool
		}{
			{name: "record version omitted", object: "record", field: "row_version"},
			{name: "nullable deletion omitted", object: "record", field: "deleted_at"},
			{name: "source value omitted", object: "source", field: "hostname"},
			{name: "source version omitted", object: "source", field: "row_version"},
			{name: "unknown record member", object: "record", field: "projection_label", add: true},
			{name: "unknown source member", object: "source", field: "projection_label", add: true},
		} {
			t.Run(test.name, func(t *testing.T) {
				bundle := h.validBundle(t)
				rows := decodePortableRows(t, bundle[revisionsTestHistoryPath])
				object := rows[0]["after_json"].(map[string]any)[test.object].(map[string]any)
				if test.add {
					object[test.field] = "not source truth"
				} else {
					if _, exists := object[test.field]; !exists {
						t.Fatalf("fixture lacks required %s.%s", test.object, test.field)
					}
					delete(object, test.field)
				}
				bundle[revisionsTestHistoryPath] = encodePortableRows(t, rows)
				requireRevisionsInvariant(t, h.applyAndValidate(t, bundle), "revisions.history_reconstruction")
			})
		}
	})
	t.Run("timestamp-looking source text remains exact", func(t *testing.T) {
		for _, chain := range []bool{false, true} {
			for _, mismatch := range []bool{false, true} {
				t.Run(fmt.Sprintf("chain_%t_mismatch_%t", chain, mismatch), func(t *testing.T) {
					h := newRevisionsPortabilityHarness(t, "exact-history-text")
					const currentText = "2026-08-01T12:00:00Z"
					priorText := currentText
					if mismatch {
						priorText = "2026-08-01T08:00:00-04:00"
					}
					version := 1
					if chain {
						version = 2
					}
					if _, err := h.db.Exec(t.Context(), `UPDATE records SET row_version=$2 WHERE record_id=$1`, h.recordID, version); err != nil {
						t.Fatal(err)
					}
					if _, err := h.db.Exec(t.Context(), `UPDATE hosts SET display_name=$2,row_version=$3 WHERE record_id=$1`, h.recordID, currentText, version); err != nil {
						t.Fatal(err)
					}
					h.snapshot["source"].(map[string]any)["display_name"] = priorText
					if chain {
						h.snapshot["source"].(map[string]any)["display_name"] = currentText
					}
					bundle := h.validBundle(t)
					if chain {
						before := clonePortableRow(t, h.snapshot)
						before["source"].(map[string]any)["display_name"] = priorText
						after := clonePortableRow(t, h.snapshot)
						after["record"].(map[string]any)["row_version"] = 2
						after["source"].(map[string]any)["row_version"] = 2
						mutations := decodePortableRows(t, bundle[revisionsTestMutationsPath])
						mutations = append(mutations, map[string]any{
							"change_set_id": h.changeSet.String(), "sequence_no": 2,
							"target_kind": "host", "target_id": h.recordID.String(), "operation_kind": "patch",
							"before_version_id": "host:" + h.recordID.String() + ":1", "after_version_id": "host:" + h.recordID.String() + ":2",
							"before_value": before, "after_value": after,
						})
						bundle[revisionsTestMutationsPath] = encodePortableRows(t, mutations)
						rows := decodePortableRows(t, bundle[revisionsTestHistoryPath])
						rows = append(rows, map[string]any{
							"revision_id": 2, "change_set_id": h.changeSet.String(), "record_id": h.recordID.String(), "row_version": 2,
							"before_json": before, "after_json": after, "created_at": h.createdAt.Format(time.RFC3339Nano),
						})
						bundle[revisionsTestHistoryPath] = encodePortableRows(t, rows)
					}
					err := h.applyAndValidate(t, bundle)
					if mismatch {
						requireRevisionsInvariant(t, err, "revisions.history_reconstruction")
					} else if err != nil {
						t.Fatalf("exact source text rejected: %v", err)
					}
				})
			}
		}
	})
}

func TestRevisionsIncidentBundleInvariantSelectionIsPermutationIndependent(t *testing.T) {
	exerciseRevisionsSourceReferencePriority(t)
	t.Run("change_set_membership_precedes_writes", func(t *testing.T) {
		h := newRevisionsPortabilityHarness(t, "membership-priority")
		for _, path := range []string{revisionsTestMutationsPath, revisionsTestHistoryPath} {
			for _, invalidHistory := range []bool{false, true} {
				t.Run(fmt.Sprintf("%s_history_%t", path, invalidHistory), func(t *testing.T) {
					ctx := context.Background()
					bundle := h.validBundle(t)
					rows := decodePortableRows(t, bundle[path])
					rows[0]["change_set_id"] = "ffffffff-ffff-4fff-8fff-ffffffffffff"
					bundle[path] = encodePortableRows(t, rows)
					if invalidHistory {
						mutations := decodePortableRows(t, bundle[revisionsTestMutationsPath])
						delete(mutations[0]["after_value"].(map[string]any), "snapshot_schema_id")
						bundle[revisionsTestMutationsPath] = encodePortableRows(t, mutations)
					}
					prepared, err := h.port.PrepareImport(ctx, bundle, h.context)
					if err != nil {
						t.Fatalf("prepare membership fixture: %v", err)
					}
					tx, err := h.db.BeginTx(ctx, pgx.TxOptions{})
					if err != nil {
						t.Fatalf("begin membership fixture: %v", err)
					}
					defer func() { _ = tx.Rollback(ctx) }()
					requireRevisionsInvariant(t, h.port.ApplyImportTx(ctx, tx, prepared, h.context), "revisions.references_complete")
					var count int
					if err := tx.QueryRow(ctx, `SELECT count(*) FROM change_sets WHERE change_set_id = $1`, h.changeSet).Scan(&count); err != nil || count != 0 {
						t.Fatalf("missing membership inserted %d change sets, err=%v", count, err)
					}
				})
			}
		}
	})
	t.Run("change_set_membership_spans_batches", func(t *testing.T) {
		for _, missingLast := range []bool{false, true} {
			t.Run(fmt.Sprintf("missing_last_%t", missingLast), func(t *testing.T) {
				ctx := context.Background()
				h := newRevisionsPortabilityHarness(t, "membership-batches")
				bundle := h.validBundle(t)
				changeSet := decodePortableRows(t, bundle[revisionsTestChangeSetsPath])[0]
				mutation := decodePortableRows(t, bundle[revisionsTestMutationsPath])[0]
				var changeSets, mutations []map[string]any
				for i := 1; i <= 129; i++ {
					id := fmt.Sprintf("00000000-0000-4000-8000-%012d", i)
					row := clonePortableRow(t, changeSet)
					row["change_set_id"] = id
					row["history_order"] = i
					if !missingLast || i != 129 {
						changeSets = append(changeSets, row)
					}
					entry := clonePortableRow(t, mutation)
					entry["change_set_id"] = id
					if i != 1 {
						entry["operation_kind"] = "patch"
						entry["before_version_id"] = entry["after_version_id"]
						entry["before_value"] = clonePortableRow(t, entry["after_value"].(map[string]any))
					}
					mutations = append(mutations, entry)
				}
				history := decodePortableRows(t, bundle[revisionsTestHistoryPath])
				history[0]["change_set_id"] = changeSets[0]["change_set_id"]
				bundle[revisionsTestChangeSetsPath] = encodePortableRows(t, changeSets)
				bundle[revisionsTestMutationsPath] = encodePortableRows(t, mutations)
				bundle[revisionsTestHistoryPath] = encodePortableRows(t, history)
				prepared, err := h.port.PrepareImport(ctx, bundle, h.context)
				if err != nil {
					t.Fatalf("prepare membership batches: %v", err)
				}
				tx, err := h.db.BeginTx(ctx, pgx.TxOptions{})
				if err != nil {
					t.Fatalf("begin membership batches: %v", err)
				}
				defer func() { _ = tx.Rollback(ctx) }()
				err = h.port.ApplyImportTx(ctx, tx, prepared, h.context)
				wantCount := 129
				if missingLast {
					requireRevisionsInvariant(t, err, "revisions.references_complete")
					wantCount = 0
				} else {
					if err != nil {
						t.Fatalf("apply admitted membership batches: %v", err)
					}
					if err := h.port.ValidateImportTx(ctx, tx, prepared, h.context); err != nil {
						t.Fatalf("validate admitted membership batches: %v", err)
					}
				}
				var count int
				if err := tx.QueryRow(ctx, `SELECT count(*) FROM change_sets WHERE incident_id = $1 AND source = 'portable.test'`, h.incidentID).Scan(&count); err != nil || count != wantCount {
					t.Fatalf("membership batches inserted %d change sets, err=%v; want %d", count, err, wantCount)
				}
			})
		}
	})
	t.Run("history_defect_does_not_mask_prior_invariants", func(t *testing.T) {
		h := newRevisionsPortabilityHarness(t, "prewrite-priority")
		for _, test := range []struct {
			name, invariant string
			mutate          func([]map[string]any, []map[string]any) ([]map[string]any, []map[string]any)
		}{
			{"later mutation reference", "revisions.references_complete", func(mutations, history []map[string]any) ([]map[string]any, []map[string]any) {
				row := clonePortableRow(t, mutations[0])
				row["sequence_no"] = 2
				row["target_id"] = "ffffffff-ffff-4fff-8fff-ffffffffffff"
				return append(mutations, row), history
			}},
			{"later revision reference", "revisions.references_complete", func(mutations, history []map[string]any) ([]map[string]any, []map[string]any) {
				history[0]["record_id"] = "ffffffff-ffff-4fff-8fff-ffffffffffff"
				return mutations, history
			}},
			{"mutation sequence", "revisions.mutation_sequence_contiguous", func(mutations, history []map[string]any) ([]map[string]any, []map[string]any) {
				mutations[0]["sequence_no"] = 2
				return mutations, history
			}},
			{"revision identity", "revisions.record_version_unique", func(mutations, history []map[string]any) ([]map[string]any, []map[string]any) {
				return mutations, append(history, clonePortableRow(t, history[0]))
			}},
			{"record version pair", "revisions.record_version_unique", func(mutations, history []map[string]any) ([]map[string]any, []map[string]any) {
				row := clonePortableRow(t, history[0])
				row["revision_id"] = 2
				return mutations, append(history, row)
			}},
			{"record version order", "revisions.record_version_unique", func(mutations, history []map[string]any) ([]map[string]any, []map[string]any) {
				row := clonePortableRow(t, history[0])
				row["revision_id"] = 2
				history[0]["row_version"] = 2
				return mutations, append(history, row)
			}},
		} {
			t.Run(test.name, func(t *testing.T) {
				ctx := context.Background()
				bundle := h.validBundle(t)
				mutations := decodePortableRows(t, bundle[revisionsTestMutationsPath])
				history := decodePortableRows(t, bundle[revisionsTestHistoryPath])
				delete(mutations[0]["after_value"].(map[string]any), "snapshot_schema_id")
				mutations, history = test.mutate(mutations, history)
				bundle[revisionsTestMutationsPath] = encodePortableRows(t, mutations)
				bundle[revisionsTestHistoryPath] = encodePortableRows(t, history)
				prepared, err := h.port.PrepareImport(ctx, bundle, h.context)
				if err != nil {
					t.Fatalf("prepare competing pre-write defects: %v", err)
				}
				tx, err := h.db.BeginTx(ctx, pgx.TxOptions{})
				if err != nil {
					t.Fatalf("begin competing pre-write defects: %v", err)
				}
				defer func() { _ = tx.Rollback(ctx) }()
				err = h.port.ApplyImportTx(ctx, tx, prepared, h.context)
				requireRevisionsInvariant(t, err, test.invariant)
				var count int
				if err := tx.QueryRow(ctx, `SELECT count(*) FROM change_sets WHERE change_set_id = $1`, h.changeSet).Scan(&count); err != nil || count != 0 {
					t.Fatalf("pre-write rejection inserted %d change sets, err=%v", count, err)
				}
			})
		}
	})
	harness := newRevisionsPortabilityHarness(t, "permutation")
	for _, reverse := range []bool{false, true} {
		t.Run(fmt.Sprintf("reverse_%t", reverse), func(t *testing.T) {
			bundle := harness.validBundle(t)
			mutations := decodePortableRows(t, bundle[revisionsTestMutationsPath])
			secondMutation := clonePortableRow(t, mutations[0])
			secondMutation["sequence_no"] = 3
			mutations = append(mutations, secondMutation)

			revisionsRows := decodePortableRows(t, bundle[revisionsTestHistoryPath])
			missing := clonePortableRow(t, revisionsRows[0])
			missing["revision_id"] = 2
			missing["record_id"] = uuid.NewString()
			revisionsRows = append(revisionsRows, missing)
			if reverse {
				mutations[0], mutations[1] = mutations[1], mutations[0]
				revisionsRows[0], revisionsRows[1] = revisionsRows[1], revisionsRows[0]
			}
			bundle[revisionsTestMutationsPath] = encodePortableRows(t, mutations)
			bundle[revisionsTestHistoryPath] = encodePortableRows(t, revisionsRows)
			requireRevisionsInvariant(
				t,
				harness.applyAndValidate(t, bundle),
				"revisions.references_complete",
			)
		})
	}
}

func exerciseStrictLinksIncidentBundleAdmission(t *testing.T) {
	harness := newRevisionsPortabilityHarness(t, "strict-links-history")
	destinationID := uuid.MustParse("77777777-7777-4777-8777-777777777777")
	seedRevisionsReferenceHost(t, harness, destinationID, harness.incidentID)
	linkID := uuid.MustParse("88888888-8888-4888-8888-888888888888")
	tagID := uuid.MustParse("99999999-9999-4999-8999-999999999999")

	linkBundle := func() portabilitysupport.MapBundle {
		bundle := harness.validBundle(t)
		rows := decodePortableRows(t, bundle[revisionsTestMutationsPath])
		rows = append(rows, map[string]any{
			"change_set_id": harness.changeSet.String(), "sequence_no": 2,
			"target_kind": "record_link", "target_id": linkID.String(), "operation_kind": "create",
			"before_version_id": nil, "after_version_id": nil, "before_value": nil,
			"after_value": canonicalPortableLinkValue(harness, linkID, destinationID),
		})
		bundle[revisionsTestMutationsPath] = encodePortableRows(t, rows)
		return bundle
	}
	tagBundle := func() portabilitysupport.MapBundle {
		bundle := harness.validBundle(t)
		rows := decodePortableRows(t, bundle[revisionsTestMutationsPath])
		rows = append(rows, map[string]any{
			"change_set_id": harness.changeSet.String(), "sequence_no": 2,
			"target_kind": "record_tag", "target_id": "record_tag:" + harness.recordID.String() + ":" + tagID.String(),
			"operation_kind": "create", "before_version_id": nil, "after_version_id": nil,
			"before_value": nil, "after_value": canonicalPortableTagValue(harness, tagID),
		})
		bundle[revisionsTestMutationsPath] = encodePortableRows(t, rows)
		return bundle
	}

	requireImportMutationCount(t, harness, linkBundle(), 2)
	requireImportMutationCount(t, harness, tagBundle(), 2)

	tests := []struct {
		name   string
		bundle func() portabilitysupport.MapBundle
		mutate func(map[string]any)
	}{
		{name: "link unknown member", bundle: linkBundle, mutate: func(row map[string]any) { row["after_value"].(map[string]any)["legacy"] = true }},
		{name: "link missing nullable", bundle: linkBundle, mutate: func(row map[string]any) { delete(row["after_value"].(map[string]any), "confidence") }},
		{name: "link mistyped confidence", bundle: linkBundle, mutate: func(row map[string]any) { row["after_value"].(map[string]any)["confidence"] = "100" }},
		{name: "link noncanonical uuid", bundle: linkBundle, mutate: func(row map[string]any) {
			row["after_value"].(map[string]any)["record_link_id"] = "88888888-8888-4888-8888-88888888888A"
		}},
		{name: "link noncanonical timestamp", bundle: linkBundle, mutate: func(row map[string]any) {
			row["after_value"].(map[string]any)["created_at"] = "2026-08-01T08:00:00.123456-04:00"
		}},
		{name: "link invalid provenance", bundle: linkBundle, mutate: func(row map[string]any) { row["after_value"].(map[string]any)["provenance"] = "legacy" }},
		{name: "link compact shape", bundle: linkBundle, mutate: func(row map[string]any) {
			value := row["after_value"].(map[string]any)
			for _, key := range []string{"field_key", "provenance", "confidence", "owner_user_id", "created_by_user_id", "decided_at", "created_at", "deleted_at", "deleted_by_user_id"} {
				delete(value, key)
			}
		}},
		{name: "link illegal create sides", bundle: linkBundle, mutate: func(row map[string]any) {
			row["before_value"] = clonePortableRow(t, row["after_value"].(map[string]any))
		}},
		{name: "tag alias", bundle: tagBundle, mutate: func(row map[string]any) {
			value := row["after_value"].(map[string]any)
			value["tag_id"] = value["record_tag_id"]
			delete(value, "record_tag_id")
		}},
		{name: "tag bare target", bundle: tagBundle, mutate: func(row map[string]any) { row["target_id"] = tagID.String() }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			bundle := test.bundle()
			rows := decodePortableRows(t, bundle[revisionsTestMutationsPath])
			test.mutate(rows[1])
			bundle[revisionsTestMutationsPath] = encodePortableRows(t, rows)
			prepared, err := harness.port.PrepareImport(context.Background(), bundle, harness.context)
			if err != nil {
				t.Fatalf("prepare valid outer row before Links history validation: %v", err)
			}
			tx, err := harness.db.BeginTx(context.Background(), pgx.TxOptions{})
			if err != nil {
				t.Fatalf("begin strict Links import: %v", err)
			}
			defer func() { _ = tx.Rollback(context.Background()) }()
			err = harness.port.ApplyImportTx(context.Background(), tx, prepared, harness.context)
			requireRevisionsInvariant(t, err, "revisions.history_reconstruction")
			var count int
			if queryErr := tx.QueryRow(context.Background(), `SELECT count(*) FROM change_sets WHERE change_set_id = $1`, harness.changeSet).Scan(&count); queryErr != nil || count != 0 {
				t.Fatalf("invalid Links history mutated import state: count=%d err=%v", count, queryErr)
			}
		})
	}
}

func TestRevisionsIncidentBundleRoundTripIsDeterministic(t *testing.T) {
	harness := newRevisionsPortabilityHarness(t, "round-trip")
	bundle := harness.validBundle(t)
	portableActorID := uuid.New()
	changeSets := decodePortableRows(t, bundle[revisionsTestChangeSetsPath])
	changeSets[0]["actor_user_id"] = portableActorID.String()
	bundle[revisionsTestChangeSetsPath] = encodePortableRows(t, changeSets)
	for _, path := range []string{revisionsTestMutationsPath, revisionsTestHistoryPath} {
		rows := decodePortableRows(t, bundle[path])
		for _, row := range rows {
			replacePortableActorIDs(row, harness.actor.ID.String(), portableActorID.String())
		}
		bundle[path] = encodePortableRows(t, rows)
	}
	actors, err := sourceport.NewActorCatalog([]sourceport.ActorDescriptor{{
		SourceActorID: portableActorID.String(),
	}})
	if err != nil {
		t.Fatalf("build remapped portable actor catalog: %v", err)
	}
	for _, version := range []int{8} {
		t.Run(fmt.Sprintf("bundle_version_%d", version), func(t *testing.T) {
			importContext := harness.context
			importContext.BundleVersion = version
			importContext.OperationID = uuid.NewString()
			importContext.Actors = actors
			attributions := &revisionsAttributionRecorder{localActor: harness.actor.ID}
			importContext.Attributions = attributions
			prepared, err := harness.port.PrepareImport(context.Background(), bundle, importContext)
			if err != nil {
				t.Fatalf("prepare valid Revisions bundle: %v", err)
			}
			tx, err := harness.db.BeginTx(context.Background(), pgx.TxOptions{})
			if err != nil {
				t.Fatalf("begin Revisions round-trip: %v", err)
			}
			defer func() { _ = tx.Rollback(context.Background()) }()
			if err := harness.port.ApplyImportTx(context.Background(), tx, prepared, importContext); err != nil {
				t.Fatalf("apply valid Revisions bundle: %v", err)
			}
			var historyRecordIDs, historyEntryRecordIDs []uuid.UUID
			if err := tx.QueryRow(context.Background(), `
SELECT history_record_ids, history_entry_record_ids
  FROM change_set_mutations
 WHERE change_set_id = $1 AND sequence_no = 1
`, harness.changeSet).Scan(&historyRecordIDs, &historyEntryRecordIDs); err != nil {
				t.Fatalf("load recomputed imported history facts: %v", err)
			}
			if len(historyRecordIDs) != 1 || historyRecordIDs[0] != harness.recordID ||
				len(historyEntryRecordIDs) != 1 || historyEntryRecordIDs[0] != harness.recordID {
				t.Fatalf("recomputed imported history facts = %v / %v", historyRecordIDs, historyEntryRecordIDs)
			}
			if err := harness.port.ValidateImportTx(context.Background(), tx, prepared, importContext); err != nil {
				t.Fatalf("validate valid Revisions bundle: %v", err)
			}
			exported, err := portabilitysupport.Export(harness.port, context.Background(), sourceport.ExportContext{
				Query: tx, IncidentID: harness.incidentID,
				PortableAttributions: revisionsPortableAttributionResolver{
					rows: attributions.ImportedAttributions(),
				},
			})
			if err != nil {
				t.Fatalf("re-export Revisions bundle: %v", err)
			}
			if len(exported) != 3 {
				t.Fatalf("Revisions export returned %d files, want all three history files", len(exported))
			}
			seen := map[string]bool{}
			for _, file := range exported {
				switch file.Path {
				case revisionsTestChangeSetsPath, revisionsTestMutationsPath, revisionsTestHistoryPath:
				default:
					t.Fatalf("unexpected Revisions export path %q", file.Path)
				}
				if seen[file.Path] {
					t.Fatalf("duplicate Revisions export path %q", file.Path)
				}
				seen[file.Path] = true
				want := bundle[file.Path]
				if file.Path == revisionsTestChangeSetsPath {
					rows := decodePortableRows(t, want)
					rows[0]["actor_import_contexts"] = []any{map[string]any{"import_id": importContext.OperationID, "actor_user_id": harness.actor.ID.String()}}
					want = encodePortableRows(t, rows)
				}
				if !bytes.Equal(want, file.Payload) {
					t.Fatalf("Revisions round trip changed %s\nexpected=%s\nexported=%s", file.Path, want, file.Payload)
				}
			}
		})
	}
}

func TestRevisionsIncidentBundleFailureIsAtomic(t *testing.T) {
	harness := newRevisionsPortabilityHarness(t, "atomic")
	bundle := harness.validBundle(t)
	rows := decodePortableRows(t, bundle[revisionsTestHistoryPath])
	rows[0]["after_json"].(map[string]any)["source"].(map[string]any)["display_name"] = "not-current"
	bundle[revisionsTestHistoryPath] = encodePortableRows(t, rows)

	prepared, err := harness.port.PrepareImport(context.Background(), bundle, harness.context)
	if err != nil {
		t.Fatalf("prepare atomicity fixture: %v", err)
	}
	tx, err := harness.db.BeginTx(context.Background(), pgx.TxOptions{})
	if err != nil {
		t.Fatalf("begin atomicity fixture: %v", err)
	}
	if err := harness.port.ApplyImportTx(context.Background(), tx, prepared, harness.context); err != nil {
		_ = tx.Rollback(context.Background())
		t.Fatalf("apply atomicity fixture: %v", err)
	}
	err = harness.port.ValidateImportTx(context.Background(), tx, prepared, harness.context)
	requireRevisionsInvariant(t, err, "revisions.history_reconstruction")
	if rollbackErr := tx.Rollback(context.Background()); rollbackErr != nil {
		t.Fatalf("rollback invalid Revisions import: %v", rollbackErr)
	}
	for table, query := range map[string]string{
		"change_sets":          `SELECT count(*) FROM change_sets WHERE change_set_id = $1`,
		"change_set_mutations": `SELECT count(*) FROM change_set_mutations WHERE change_set_id = $1`,
		"record_revisions":     `SELECT count(*) FROM record_revisions WHERE change_set_id = $1`,
	} {
		var count int
		if err := harness.db.QueryRow(context.Background(), query, harness.changeSet).Scan(&count); err != nil {
			t.Fatalf("count atomic %s: %v", table, err)
		}
		if count != 0 {
			t.Fatalf("invalid import left %d rows in %s", count, table)
		}
	}
}

func newRevisionsPortabilityHarness(t testing.TB, suffix string) revisionsPortabilityHarness {
	t.Helper()
	store := appsupport.StartStore(t, "revisions-portability-"+suffix)
	actor := authstoretest.SeedActor(t, store.DB, authstoretest.ActorAttributes{Email: "revisions-portability-" + uuid.NewString() + "@example.test", DisplayName: "Revisions Portability"})
	incident := appsupport.CreateIncidentInStore(
		t,
		store.DB,
		actor,
		"txn-revisions-portability-"+uuid.NewString(),
		"IR-RP-"+uuid.NewString()[:8],
		"Revisions portability "+suffix,
	)
	recordID := uuid.New()
	createdAt := time.Date(2026, time.August, 1, 12, 0, 0, 123456000, time.UTC)
	if _, err := store.DB.Exec(context.Background(), `
INSERT INTO records (
    record_id, incident_id, record_type, created_by_user_id, created_at,
    updated_by_user_id, updated_at, row_version
) VALUES ($1, $2, 'host', $3, $4, $3, $4, 1)
`, recordID, incident.ID, actor.ID, createdAt); err != nil {
		t.Fatalf("seed portable record envelope: %v", err)
	}
	if _, err := store.DB.Exec(context.Background(), `
INSERT INTO hosts (
    record_id, incident_id, display_name, hostname, host_state,
    row_version, created_at, updated_at, created_by_user_id, updated_by_user_id
) VALUES ($1, $2, 'Portable host', 'portable-host', 'canonical', 1, $3, $3, $4, $4)
`, recordID, incident.ID, createdAt, actor.ID); err != nil {
		t.Fatalf("seed portable host source: %v", err)
	}
	contributions := mustRevisionProviderContributions(t)
	snapshotCaptures, err := revisions.NewRecordSnapshotCaptureCatalog(contributions, records.ValidateSnapshotEnvelope)
	if err != nil {
		t.Fatalf("build Revisions snapshot catalog: %v", err)
	}
	targetSemantics, err := revisions.NewTargetSemanticsCatalog(contributions)
	if err != nil {
		t.Fatalf("build Revisions target-semantics catalog: %v", err)
	}
	port, err := revisions.NewIncidentBundleSourcePort(
		revisionsPortabilityEnvelopeReader{},
		snapshotCaptures,
		targetSemantics,
	)
	if err != nil {
		t.Fatalf("build Revisions portability source port: %v", err)
	}
	var hostSource revisions.RecordProviderContribution
	for _, contribution := range contributions {
		for _, record := range contribution.Records {
			if record.RecordType == "host" {
				hostSource = record
			}
		}
	}
	if hostSource.DeleteRestoreSource == nil {
		t.Fatal("host canonical-current-row provider is missing")
	}
	tx, err := store.DB.BeginTx(context.Background(), pgx.TxOptions{})
	if err != nil {
		t.Fatalf("begin portable snapshot read: %v", err)
	}
	snapshot, err := hostSource.DeleteRestoreSource.SnapshotTx(context.Background(), tx, recordID)
	_ = tx.Rollback(context.Background())
	if err != nil {
		t.Fatalf("load portable canonical current row: %v", err)
	}
	snapshot["snapshot_schema_id"] = hostSource.SnapshotSchemaID
	actors, err := sourceport.NewActorCatalog([]sourceport.ActorDescriptor{{
		SourceActorID: actor.ID.String(),
	}})
	if err != nil {
		t.Fatalf("build portable actor catalog: %v", err)
	}
	attributions := &revisionsAttributionRecorder{localActor: actor.ID}
	return revisionsPortabilityHarness{
		db: store.DB, actor: actor, incidentID: incident.ID, recordID: recordID,
		changeSet: uuid.New(), createdAt: createdAt, snapshot: snapshot, port: port,
		context: sourceport.ImportContext{
			IncidentID: incident.ID, ActorUserID: actor.ID, BundleVersion: 8,
			OperationID:  uuid.NewString(),
			Attributions: attributions, Actors: actors,
		},
	}
}

func (h revisionsPortabilityHarness) validBundle(t testing.TB) portabilitysupport.MapBundle {
	t.Helper()
	changeSets := []map[string]any{{
		"change_set_id": h.changeSet.String(), "incident_id": h.incidentID.String(),
		"actor_user_id": h.actor.ID.String(), "source": "portable.test",
		"actor_import_contexts": []any{}, "history_order": 1,
		"reason": nil, "client_txn_id": nil, "request_id": nil,
		"created_at": h.createdAt.Format(time.RFC3339Nano),
	}}
	mutations := []map[string]any{{
		"change_set_id": h.changeSet.String(), "sequence_no": 1,
		"target_kind": "host", "target_id": h.recordID.String(),
		"operation_kind": "create", "before_version_id": nil,
		"after_version_id": "host:" + h.recordID.String() + ":1",
		"before_value":     nil, "after_value": h.snapshot,
	}}
	revisionRows := []map[string]any{{
		"revision_id": 1, "change_set_id": h.changeSet.String(),
		"record_id": h.recordID.String(), "row_version": 1,
		"before_json": nil, "after_json": h.snapshot,
		"created_at": h.createdAt.Format(time.RFC3339Nano),
	}}
	return portabilitysupport.MapBundle{
		revisionsTestChangeSetsPath: encodePortableRows(t, changeSets),
		revisionsTestMutationsPath:  encodePortableRows(t, mutations),
		revisionsTestHistoryPath:    encodePortableRows(t, revisionRows),
	}
}

func (h revisionsPortabilityHarness) freshImportContext() sourceport.ImportContext {
	input := h.context
	input.Attributions = &revisionsAttributionRecorder{localActor: h.actor.ID}
	return input
}

func (h revisionsPortabilityHarness) applyAndValidate(t testing.TB, bundle portabilitysupport.MapBundle) error {
	t.Helper()
	input := h.freshImportContext()
	prepared, err := h.port.PrepareImport(context.Background(), bundle, input)
	if err != nil {
		return err
	}
	tx, err := h.db.BeginTx(context.Background(), pgx.TxOptions{})
	if err != nil {
		t.Fatalf("begin Revisions portability transaction: %v", err)
	}
	defer func() { _ = tx.Rollback(context.Background()) }()
	if err := h.port.ApplyImportTx(context.Background(), tx, prepared, input); err != nil {
		return err
	}
	return h.port.ValidateImportTx(context.Background(), tx, prepared, input)
}

func canonicalPortableLinkValue(h revisionsPortabilityHarness, linkID uuid.UUID, destinationID uuid.UUID) map[string]any {
	return map[string]any{
		"record_link_id": linkID.String(), "incident_id": h.incidentID.String(),
		"src_record_id": h.recordID.String(), "dst_record_id": destinationID.String(),
		"link_type": "references_record", "field_key": nil, "provenance": "manual", "confidence": nil,
		"owner_user_id": h.actor.ID.String(), "created_by_user_id": h.actor.ID.String(),
		"decided_at": h.createdAt.Format(time.RFC3339Nano), "created_at": h.createdAt.Format(time.RFC3339Nano),
		"deleted_at": nil, "deleted_by_user_id": nil,
	}
}

func canonicalPortableTagValue(h revisionsPortabilityHarness, tagID uuid.UUID) map[string]any {
	return map[string]any{
		"record_tag_id": tagID.String(), "incident_id": h.incidentID.String(), "record_id": h.recordID.String(),
		"tag_name": "Portable", "normalized_tag_name": "portable", "created_by_user_id": h.actor.ID.String(),
		"created_at": h.createdAt.Format(time.RFC3339Nano), "updated_at": h.createdAt.Format(time.RFC3339Nano),
		"deleted_at": nil, "deleted_by_user_id": nil,
	}
}

func requireImportMutationCount(t testing.TB, h revisionsPortabilityHarness, bundle portabilitysupport.MapBundle, want int) {
	t.Helper()
	input := h.freshImportContext()
	prepared, err := h.port.PrepareImport(context.Background(), bundle, input)
	if err != nil {
		t.Fatalf("prepare canonical Links history import: %v", err)
	}
	tx, err := h.db.BeginTx(context.Background(), pgx.TxOptions{})
	if err != nil {
		t.Fatalf("begin canonical Links history import: %v", err)
	}
	defer func() { _ = tx.Rollback(context.Background()) }()
	if err := h.port.ApplyImportTx(context.Background(), tx, prepared, input); err != nil {
		t.Fatalf("apply canonical Links history import: %v", err)
	}
	var count int
	if err := tx.QueryRow(context.Background(), `SELECT count(*) FROM change_set_mutations WHERE change_set_id = $1`, h.changeSet).Scan(&count); err != nil || count != want {
		t.Fatalf("canonical Links history mutation count=%d want=%d err=%v", count, want, err)
	}
}

func requireRevisionsInvariant(t testing.TB, err error, invariantID string) {
	t.Helper()
	var failure *sourceport.Failure
	if !errors.As(err, &failure) || failure.FamilyID() != "revisions" ||
		failure.InvariantID() != invariantID {
		t.Fatalf("Revisions portability failure = %#v, %v; want %s", failure, err, invariantID)
	}
	for _, forbidden := range []string{"SELECT", "record_revisions", "must-not-escape"} {
		if bytes.Contains([]byte(err.Error()), []byte(forbidden)) {
			t.Fatalf("safe Revisions failure exposed %q: %v", forbidden, err)
		}
	}
}

func encodePortableRows(t testing.TB, rows []map[string]any) []byte {
	t.Helper()
	var payload bytes.Buffer
	for _, row := range rows {
		encoded, err := incidentportability.CanonicalJSONString(row)
		if err != nil {
			t.Fatalf("encode portable row: %v", err)
		}
		payload.Write(encoded)
	}
	return payload.Bytes()
}

func decodePortableRows(t testing.TB, payload []byte) []map[string]any {
	t.Helper()
	rows, err := incidentportability.DecodeStrictNDJSONObjects(payload, "test")
	if err != nil {
		t.Fatalf("decode portable rows: %v", err)
	}
	return rows
}

func clonePortableRow(t testing.TB, row map[string]any) map[string]any {
	t.Helper()
	raw, err := json.Marshal(row)
	if err != nil {
		t.Fatalf("marshal portable row clone: %v", err)
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	var clone map[string]any
	if err := decoder.Decode(&clone); err != nil {
		t.Fatalf("decode portable row clone: %v", err)
	}
	return clone
}

func clonePortableBundle(bundle portabilitysupport.MapBundle) portabilitysupport.MapBundle {
	clone := make(portabilitysupport.MapBundle, len(bundle))
	for path, payload := range bundle {
		clone[path] = append([]byte(nil), payload...)
	}
	return clone
}

func replacePortableActorIDs(value any, oldActorID string, newActorID string) {
	switch typed := value.(type) {
	case map[string]any:
		for key, member := range typed {
			if member == oldActorID && len(key) >= len("_user_id") &&
				key[len(key)-len("_user_id"):] == "_user_id" {
				typed[key] = newActorID
				continue
			}
			replacePortableActorIDs(member, oldActorID, newActorID)
		}
	case []any:
		for _, item := range typed {
			replacePortableActorIDs(item, oldActorID, newActorID)
		}
	}
}

func (r *revisionsAttributionRecorder) VisitImportedAttributions(table, rowID string, visit func(incidentportability.ImportedAttribution) error) error {
	if r.visitErr != nil {
		return r.visitErr
	}
	for _, row := range r.ImportedAttributions() {
		if row.SourceTable == table && (rowID == "" || row.SourceRowID == rowID) {
			if err := visit(row); err != nil {
				return err
			}
		}
	}
	return nil
}
