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

	"github.com/JochiRaider/cartulary/internal/app/revisionassembly"
	authstoretest "github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/storetest"
	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
	"github.com/JochiRaider/cartulary/internal/modules/incidentportability"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
)

const (
	revisionsTestChangeSetsPath = "data/change_sets.ndjson"
	revisionsTestMutationsPath  = "data/change_set_mutations.ndjson"
	revisionsTestHistoryPath    = "data/record_revisions.ndjson"
)

type revisionsPortabilityHarness struct {
	db         postgres.DB
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
	rows []incidentportability.ImportedAttribution
}

type revisionsPortableAttributionResolver struct {
	rows []incidentportability.ImportedAttribution
}

type revisionsPortabilityEnvelopeReader struct{}

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
	r.rows = append(r.rows, incidentportability.ImportedAttribution{
		SourceTable: table, SourceRowID: rowID, SourceColumn: column,
		SourceActorID: actorID,
	})
	return nil
}

func (r *revisionsAttributionRecorder) ImportedAttributions() []incidentportability.ImportedAttribution {
	return append([]incidentportability.ImportedAttribution(nil), r.rows...)
}

func TestRevisionsIncidentBundleInvariantAttribution(t *testing.T) {
	harness := newRevisionsPortabilityHarness(t, "invariant-attribution")
	valid := harness.validBundle(t)

	tests := []struct {
		name      string
		invariant string
		mutate    func(sourceport.MapBundle)
	}{
		{
			name:      "references_complete",
			invariant: "revisions.references_complete",
			mutate: func(bundle sourceport.MapBundle) {
				rows := decodePortableRows(t, bundle[revisionsTestHistoryPath])
				rows[0]["record_id"] = uuid.NewString()
				bundle[revisionsTestHistoryPath] = encodePortableRows(t, rows)
			},
		},
		{
			name:      "actor_references_complete",
			invariant: "revisions.actor_references_complete",
			mutate: func(bundle sourceport.MapBundle) {
				rows := decodePortableRows(t, bundle[revisionsTestChangeSetsPath])
				rows[0]["actor_user_id"] = uuid.NewString()
				bundle[revisionsTestChangeSetsPath] = encodePortableRows(t, rows)
			},
		},
		{
			name:      "mutation_sequence_contiguous",
			invariant: "revisions.mutation_sequence_contiguous",
			mutate: func(bundle sourceport.MapBundle) {
				rows := decodePortableRows(t, bundle[revisionsTestMutationsPath])
				rows[0]["sequence_no"] = 2
				bundle[revisionsTestMutationsPath] = encodePortableRows(t, rows)
			},
		},
		{
			name:      "record_version_unique",
			invariant: "revisions.record_version_unique",
			mutate: func(bundle sourceport.MapBundle) {
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
			mutate: func(bundle sourceport.MapBundle) {
				rows := decodePortableRows(t, bundle[revisionsTestHistoryPath])
				after := rows[0]["after_json"].(map[string]any)
				after["source"].(map[string]any)["display_name"] = "not-current"
				bundle[revisionsTestHistoryPath] = encodePortableRows(t, rows)
			},
		},
		{
			name:      "schema_less_history_rejected",
			invariant: "revisions.history_reconstruction",
			mutate: func(bundle sourceport.MapBundle) {
				rows := decodePortableRows(t, bundle[revisionsTestHistoryPath])
				delete(rows[0]["after_json"].(map[string]any), "snapshot_schema_id")
				bundle[revisionsTestHistoryPath] = encodePortableRows(t, rows)
			},
		},
		{
			name:      "schema_less_mutation_history_rejected",
			invariant: "revisions.history_reconstruction",
			mutate: func(bundle sourceport.MapBundle) {
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
		mutate func(sourceport.MapBundle)
	}{
		{
			name: "unknown_member",
			mutate: func(bundle sourceport.MapBundle) {
				rows := decodePortableRows(t, bundle[revisionsTestChangeSetsPath])
				rows[0]["hostile_unknown"] = "must-not-escape"
				bundle[revisionsTestChangeSetsPath] = encodePortableRows(t, rows)
			},
		},
		{
			name: "missing_nullable_member",
			mutate: func(bundle sourceport.MapBundle) {
				rows := decodePortableRows(t, bundle[revisionsTestMutationsPath])
				delete(rows[0], "before_value")
				bundle[revisionsTestMutationsPath] = encodePortableRows(t, rows)
			},
		},
		{
			name: "wrong_integer_type",
			mutate: func(bundle sourceport.MapBundle) {
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
	t.Run("links_history_is_strict_before_import_mutation", func(t *testing.T) {
		exerciseStrictLinksIncidentBundleAdmission(t)
	})
}

func TestRevisionsIncidentBundleInvariantSelectionIsPermutationIndependent(t *testing.T) {
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
	linkID := uuid.MustParse("88888888-8888-4888-8888-888888888888")
	tagID := uuid.MustParse("99999999-9999-4999-8999-999999999999")

	linkBundle := func() sourceport.MapBundle {
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
	tagBundle := func() sourceport.MapBundle {
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
		bundle func() sourceport.MapBundle
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
				requireRevisionsInvariant(t, err, "revisions.references_complete")
				return
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
	for _, version := range []int{3} {
		t.Run(fmt.Sprintf("bundle_version_%d", version), func(t *testing.T) {
			importContext := harness.context
			importContext.BundleVersion = version
			importContext.OperationID = fmt.Sprintf("revisions-round-trip-v%d", version)
			importContext.Actors = actors
			attributions := &revisionsAttributionRecorder{}
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
			exported, err := harness.port.Export(context.Background(), sourceport.ExportContext{
				Query: tx, IncidentID: harness.incidentID,
				PortableAttributions: revisionsPortableAttributionResolver{
					rows: attributions.ImportedAttributions(),
				},
			})
			if err != nil {
				t.Fatalf("re-export Revisions bundle: %v", err)
			}
			for _, file := range exported {
				if !bytes.Equal(bundle[file.Path], file.Payload) {
					t.Fatalf("Revisions round trip changed %s\noriginal=%s\nexported=%s", file.Path, bundle[file.Path], file.Payload)
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
	actor := authstoretest.SeedLocalUserRecord(
		t,
		store.DB,
		"revisions-portability-"+uuid.NewString()+"@example.test",
		"Revisions Portability",
		"RevisionsPortability1!",
		false,
		false,
		true,
	)
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
	targetSemantics, err := revisionassembly.CurrentTargetSemanticsCatalog()
	if err != nil {
		t.Fatalf("build Revisions target-semantics catalog: %v", err)
	}
	validation, err := revisions.NewIncidentBundleValidationCatalog(
		revisionsPortabilityEnvelopeReader{},
		targetSemantics,
		contributions,
	)
	if err != nil {
		t.Fatalf("build Revisions portability validation catalog: %v", err)
	}
	port := revisions.NewIncidentBundleSourcePort(validation)
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
	attributions := &revisionsAttributionRecorder{}
	return revisionsPortabilityHarness{
		db: store.DB, actor: actor, incidentID: incident.ID, recordID: recordID,
		changeSet: uuid.New(), createdAt: createdAt, snapshot: snapshot, port: port,
		context: sourceport.ImportContext{
			IncidentID: incident.ID, ActorUserID: actor.ID, BundleVersion: 3,
			OperationID:  "revisions-portability-" + suffix,
			Attributions: attributions, Actors: actors,
		},
	}
}

func (h revisionsPortabilityHarness) validBundle(t testing.TB) sourceport.MapBundle {
	t.Helper()
	changeSets := []map[string]any{{
		"change_set_id": h.changeSet.String(), "incident_id": h.incidentID.String(),
		"actor_user_id": h.actor.ID.String(), "source": "portable.test",
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
	return sourceport.MapBundle{
		revisionsTestChangeSetsPath: encodePortableRows(t, changeSets),
		revisionsTestMutationsPath:  encodePortableRows(t, mutations),
		revisionsTestHistoryPath:    encodePortableRows(t, revisionRows),
	}
}

func (h revisionsPortabilityHarness) applyAndValidate(t testing.TB, bundle sourceport.MapBundle) error {
	t.Helper()
	prepared, err := h.port.PrepareImport(context.Background(), bundle, h.context)
	if err != nil {
		return err
	}
	tx, err := h.db.BeginTx(context.Background(), pgx.TxOptions{})
	if err != nil {
		t.Fatalf("begin Revisions portability transaction: %v", err)
	}
	defer func() { _ = tx.Rollback(context.Background()) }()
	if err := h.port.ApplyImportTx(context.Background(), tx, prepared, h.context); err != nil {
		return err
	}
	return h.port.ValidateImportTx(context.Background(), tx, prepared, h.context)
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

func requireImportMutationCount(t testing.TB, h revisionsPortabilityHarness, bundle sourceport.MapBundle, want int) {
	t.Helper()
	prepared, err := h.port.PrepareImport(context.Background(), bundle, h.context)
	if err != nil {
		t.Fatalf("prepare canonical Links history import: %v", err)
	}
	tx, err := h.db.BeginTx(context.Background(), pgx.TxOptions{})
	if err != nil {
		t.Fatalf("begin canonical Links history import: %v", err)
	}
	defer func() { _ = tx.Rollback(context.Background()) }()
	if err := h.port.ApplyImportTx(context.Background(), tx, prepared, h.context); err != nil {
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

func clonePortableBundle(bundle sourceport.MapBundle) sourceport.MapBundle {
	clone := make(sourceport.MapBundle, len(bundle))
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
