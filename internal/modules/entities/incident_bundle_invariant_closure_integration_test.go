package entities

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
	"github.com/JochiRaider/cartulary/internal/modules/incidentportability"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
)

func TestEntityIncidentBundleInvariantClosure_Integration(t *testing.T) {
	ctx := context.Background()
	db := pgtest.Start(t).BeginRollbackDBT(t, "entities-incident-bundle-invariants")
	actorID, incidentID := seedEntityPortableOwner(t, ctx, db)
	hostRecordID := uuid.MustParse("00000000-0000-4000-8000-000000000101")
	mismatchedRecordID := uuid.MustParse("00000000-0000-4000-8000-000000000102")
	timestamp := time.Date(2026, time.August, 22, 20, 0, 0, 123000, time.UTC)
	seedEntityPortableEnvelope(t, ctx, db, incidentID, hostRecordID, "host", actorID, timestamp)
	seedEntityPortableEnvelope(t, ctx, db, incidentID, mismatchedRecordID, "identity", actorID, timestamp)
	importContext := entityPortableTestContext(t, incidentID, actorID, "entities-portable")
	base := entityPortableTestBundle(t, entityPortableHostMap(hostRecordID, incidentID, actorID, timestamp))

	t.Run("valid v2 rows import export byte exactly and reconstruct claims", func(t *testing.T) {
		attributions := &entityPortableTestAttributions{}
		validContext := importContext
		validContext.OperationID = "entities-portable-valid"
		validContext.Attributions = attributions
		prepared, err := prepareEntityImport(base, validContext)
		if err != nil {
			t.Fatalf("prepare valid entity source: %v", err)
		}
		tx, err := db.BeginTx(ctx, pgx.TxOptions{})
		if err != nil {
			t.Fatalf("begin valid entity source transaction: %v", err)
		}
		defer func() { _ = tx.Rollback(ctx) }()
		if err := applyPreparedEntityImportTx(ctx, tx, prepared, validContext); err != nil {
			t.Fatalf("apply valid entity source: %v", err)
		}
		if err := validatePreparedEntityImportTx(ctx, tx, prepared, validContext); err != nil {
			t.Fatalf("validate valid entity source: %v", err)
		}
		files, err := exportEntityIncidentBundleFiles(ctx, sourceport.ExportContext{Query: tx, IncidentID: incidentID})
		if err != nil {
			t.Fatalf("export imported entity source: %v", err)
		}
		for _, file := range files {
			want, _ := base.File(file.Path)
			if !bytes.Equal(file.Payload, want) {
				t.Fatalf("portable member %s changed across import/export", file.Path)
			}
		}
		var claimCount int
		if err := tx.QueryRow(ctx, "SELECT count(*) FROM entity_active_identifier_claims WHERE incident_id = $1 AND record_id = $2", incidentID, hostRecordID).Scan(&claimCount); err != nil {
			t.Fatalf("count reconstructed entity claims: %v", err)
		}
		if claimCount != 1 {
			t.Fatalf("reconstructed entity claims = %d, want 1", claimCount)
		}
	})

	prepareFailures := []struct {
		name, invariant string
		mutate          func(sourceport.MapBundle)
	}{
		{"source identity admitted", entitySourceIdentity, func(bundle sourceport.MapBundle) {
			row := entityPortableHostMap(hostRecordID, incidentID, actorID, timestamp)
			row["undeclared"] = true
			bundle[hostsBundlePath] = entityMarshalPortableRows(t, row)
		}},
		{"mentions observational", entityMentionsObserved, func(bundle sourceport.MapBundle) {
			row := entityPortableMentionMap(uuid.New(), hostRecordID, actorID, timestamp)
			row["origin_kind"] = "legacy"
			bundle[entityMentionsBundlePath] = entityMarshalPortableRows(t, row)
		}},
		{"resolution merge coherent", entityResolutionMerge, func(bundle sourceport.MapBundle) {
			row := entityPortableHostMap(hostRecordID, incidentID, actorID, timestamp)
			row["host_state"], row["merged_into_record_id"] = "merged", hostRecordID.String()
			bundle[hostsBundlePath] = entityMarshalPortableRows(t, row)
		}},
		{"alias identifier normalized", entityNormalized, func(bundle sourceport.MapBundle) {
			row := entityPortableAliasMap(uuid.New(), incidentID, hostRecordID, actorID, timestamp)
			row["raw_text"], row["normalized_text"] = " padded ", "wrong"
			bundle[entityAliasesBundlePath] = entityMarshalPortableRows(t, row)
		}},
		{"alias identifier classified", entityClassified, func(bundle sourceport.MapBundle) {
			row := entityPortablePreservedMap(uuid.New(), incidentID, hostRecordID, actorID, timestamp)
			row["identifier_type"], row["raw_value"], row["normalized_value"] = "sid", "S-1-5-21", "S-1-5-21"
			bundle[preservedIDsBundlePath] = entityMarshalPortableRows(t, row)
		}},
		{"alias identifier unique", entityUnique, func(bundle sourceport.MapBundle) {
			first := entityPortableAliasMap(uuid.MustParse("00000000-0000-4000-8000-000000000201"), incidentID, hostRecordID, actorID, timestamp)
			second := entityPortableAliasMap(uuid.MustParse("00000000-0000-4000-8000-000000000202"), incidentID, hostRecordID, actorID, timestamp)
			bundle[entityAliasesBundlePath] = entityMarshalPortableRows(t, first, second)
		}},
	}
	for _, test := range prepareFailures {
		t.Run(test.name, func(t *testing.T) {
			bundle := cloneEntityPortableBundle(base)
			test.mutate(bundle)
			_, err := prepareEntityImport(bundle, importContext)
			requireEntityPortableFailure(t, err, test.invariant)
			assertNoEntityPortableRows(t, ctx, db, incidentID)
		})
	}

	t.Run("envelope type scope", func(t *testing.T) {
		bundle := entityPortableTestBundle(t, entityPortableHostMap(mismatchedRecordID, incidentID, actorID, timestamp))
		requireEntityPortableApplyFailure(t, ctx, db, bundle, importContext, entityEnvelopeTypeScope)
	})
	t.Run("alias identifier same incident", func(t *testing.T) {
		bundle := cloneEntityPortableBundle(base)
		bundle[entityAliasesBundlePath] = entityMarshalPortableRows(t, entityPortableAliasMap(uuid.New(), incidentID, uuid.New(), actorID, timestamp))
		requireEntityPortableApplyFailure(t, ctx, db, bundle, importContext, entitySameIncident)
	})
	t.Run("multiple defects use declared invariant precedence", func(t *testing.T) {
		bundle := cloneEntityPortableBundle(base)
		host := entityPortableHostMap(hostRecordID, incidentID, actorID, timestamp)
		host["undeclared"] = true
		bundle[hostsBundlePath] = entityMarshalPortableRows(t, host)
		alias := entityPortableAliasMap(uuid.New(), incidentID, hostRecordID, actorID, timestamp)
		alias["normalized_text"] = "wrong"
		bundle[entityAliasesBundlePath] = entityMarshalPortableRows(t, alias)
		_, err := prepareEntityImport(bundle, importContext)
		requireEntityPortableFailure(t, err, entitySourceIdentity)
	})
}

func seedEntityPortableOwner(t testing.TB, ctx context.Context, db *pgtest.RollbackDB) (uuid.UUID, uuid.UUID) {
	t.Helper()
	actorID, incidentID := uuid.New(), uuid.New()
	if _, err := db.Exec(ctx, `
INSERT INTO users (id, email, display_name, password_hash, mfa_required, is_active, is_deployment_admin)
VALUES ($1, $2, 'Entity portable actor', 'test-only', false, true, true)
`, actorID, "entity-portable-"+actorID.String()+"@example.test"); err != nil {
		t.Fatalf("seed entity portable actor: %v", err)
	}
	if _, err := db.Exec(ctx, `
INSERT INTO incidents (id, incident_key, incident_key_canonical, title, status, created_by_user_id, updated_by_user_id)
VALUES ($1, $2, $2, 'Entity portable incident', 'active', $3, $3)
`, incidentID, "ENTITY-PORTABLE-"+incidentID.String(), actorID); err != nil {
		t.Fatalf("seed entity portable incident: %v", err)
	}
	return actorID, incidentID
}

func seedEntityPortableEnvelope(t testing.TB, ctx context.Context, db *pgtest.RollbackDB, incidentID, recordID uuid.UUID, recordType string, actorID uuid.UUID, timestamp time.Time) {
	t.Helper()
	if _, err := db.Exec(ctx, `
INSERT INTO records (
    record_id, incident_id, record_type, row_version, created_at, updated_at,
    created_by_user_id, updated_by_user_id
)
VALUES ($1, $2, $3, 1, $4, $4, $5, $5)
`, recordID, incidentID, recordType, timestamp, actorID); err != nil {
		t.Fatalf("seed entity portable envelope: %v", err)
	}
}

func entityPortableTestContext(t testing.TB, incidentID, actorID uuid.UUID, operationID string) sourceport.ImportContext {
	t.Helper()
	actors, err := sourceport.NewActorCatalog([]sourceport.ActorDescriptor{{SourceActorID: actorID.String()}})
	if err != nil {
		t.Fatalf("construct entity portable actor catalog: %v", err)
	}
	return sourceport.ImportContext{
		IncidentID: incidentID, ActorUserID: actorID, BundleVersion: 2,
		OperationID: operationID, Actors: actors,
	}
}

func entityPortableTestBundle(t testing.TB, host map[string]any) sourceport.MapBundle {
	t.Helper()
	return sourceport.MapBundle{
		entityMentionsBundlePath: {},
		hostsBundlePath:          entityMarshalPortableRows(t, host),
		identitiesBundlePath:     {},
		preservedIDsBundlePath:   {},
		entityAliasesBundlePath:  {},
	}
}

func entityPortableHostMap(recordID, incidentID, actorID uuid.UUID, timestamp time.Time) map[string]any {
	return map[string]any{
		"record_id": recordID.String(), "incident_id": incidentID.String(), "display_name": "Portable Host",
		"hostname": "portable.example.test", "aad_device_id": nil, "fqdn": nil, "entity_origin": "entity_import",
		"seed_entity_mention_id": nil, "host_state": "canonical", "merged_into_record_id": nil, "row_version": 1,
		"created_at": entityFormatTimestamp(timestamp), "updated_at": entityFormatTimestamp(timestamp),
		"created_by_user_id": actorID.String(), "updated_by_user_id": actorID.String(), "location": nil,
		"os_platform": nil, "business_owner": nil, "criticality": nil, "containment_status": nil,
	}
}

func entityPortableMentionMap(mentionID, sourceRecordID, actorID uuid.UUID, timestamp time.Time) map[string]any {
	return map[string]any{
		"entity_mention_id": mentionID.String(), "source_record_id": sourceRecordID.String(), "entity_type": "host",
		"source_field_key": "timeline.activity_synopsis_text", "origin_kind": "manual_entry", "origin_locator": "fixture",
		"raw_text": "Portable Host", "normalized_text": "Portable Host", "resolution_status": "unresolved",
		"row_version": 1, "ordinal": 1, "created_by_user_id": actorID.String(), "created_at": entityFormatTimestamp(timestamp),
		"resolved_record_id": nil, "resolved_by_user_id": nil, "resolved_at": nil, "resolution_method": nil,
	}
}

func entityPortableAliasMap(aliasID, incidentID, recordID, actorID uuid.UUID, timestamp time.Time) map[string]any {
	return map[string]any{
		"entity_alias_id": aliasID.String(), "incident_id": incidentID.String(), "record_id": recordID.String(),
		"entity_type": "host", "raw_text": "Portable Alias", "normalized_text": "Portable Alias",
		"classification": "suggestion_only", "created_by_user_id": actorID.String(),
		"created_at": entityFormatTimestamp(timestamp), "deleted_at": nil,
	}
}

func entityPortablePreservedMap(preservedID, incidentID, recordID, actorID uuid.UUID, timestamp time.Time) map[string]any {
	return map[string]any{
		"entity_preserved_identifier_id": preservedID.String(), "incident_id": incidentID.String(), "record_id": recordID.String(),
		"entity_type": "host", "identifier_type": "hostname", "raw_value": "Portable.Example.Test",
		"normalized_value": "portable.example.test", "classification": "exact_match_reuse",
		"created_by_user_id": actorID.String(), "created_at": entityFormatTimestamp(timestamp), "deleted_at": nil,
	}
}

func entityMarshalPortableRows(t testing.TB, rows ...map[string]any) []byte {
	t.Helper()
	var payload []byte
	for _, row := range rows {
		encoded, err := json.Marshal(row)
		if err != nil {
			t.Fatalf("marshal entity portable row: %v", err)
		}
		payload = append(payload, encoded...)
		payload = append(payload, '\n')
	}
	return payload
}

func cloneEntityPortableBundle(bundle sourceport.MapBundle) sourceport.MapBundle {
	result := sourceport.MapBundle{}
	for path, payload := range bundle {
		result[path] = append([]byte(nil), payload...)
	}
	return result
}

func requireEntityPortableFailure(t testing.TB, err error, invariant string) {
	t.Helper()
	var failure *sourceport.Failure
	if !errors.As(err, &failure) || failure.FamilyID() != "entities" || failure.InvariantID() != invariant {
		t.Fatalf("entity portable failure = %v, want entities/%s", err, invariant)
	}
}

func requireEntityPortableApplyFailure(t testing.TB, ctx context.Context, db *pgtest.RollbackDB, bundle sourceport.MapBundle, baseContext sourceport.ImportContext, invariant string) {
	t.Helper()
	attributions := &entityPortableTestAttributions{}
	importContext := baseContext
	importContext.OperationID = "entities-portable-" + invariant
	importContext.Attributions = attributions
	prepared, err := prepareEntityImport(bundle, importContext)
	if err != nil {
		t.Fatalf("prepare cross-row entity failure: %v", err)
	}
	tx, err := db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		t.Fatalf("begin cross-row entity failure: %v", err)
	}
	err = applyPreparedEntityImportTx(ctx, tx, prepared, importContext)
	requireEntityPortableFailure(t, err, invariant)
	if err := tx.Rollback(ctx); err != nil {
		t.Fatalf("roll back cross-row entity failure: %v", err)
	}
	assertNoEntityPortableRows(t, ctx, db, baseContext.IncidentID)
}

func assertNoEntityPortableRows(t testing.TB, ctx context.Context, db *pgtest.RollbackDB, incidentID uuid.UUID) {
	t.Helper()
	queries := []string{
		"SELECT count(*) FROM hosts WHERE incident_id = $1",
		"SELECT count(*) FROM identities WHERE incident_id = $1",
		"SELECT count(*) FROM entity_aliases WHERE incident_id = $1",
		"SELECT count(*) FROM entity_preserved_identifiers WHERE incident_id = $1",
		"SELECT count(*) FROM entity_mentions AS mention JOIN records AS source ON source.record_id = mention.source_record_id WHERE source.incident_id = $1",
	}
	for _, query := range queries {
		var count int
		if err := db.QueryRow(ctx, query, incidentID).Scan(&count); err != nil {
			t.Fatalf("count entity portable rows: %v", err)
		}
		if count != 0 {
			t.Fatalf("failed entity import left %d source rows", count)
		}
	}
}

type entityPortableTestAttributions struct {
	rows []incidentportability.ImportedAttribution
}

func (r *entityPortableTestAttributions) RecordImportedAttribution(table, rowID, column, actorID string) error {
	r.rows = append(r.rows, incidentportability.ImportedAttribution{
		SourceTable: table, SourceRowID: rowID, SourceColumn: column, SourceActorID: actorID,
	})
	return nil
}

func (r *entityPortableTestAttributions) ImportedAttributions() []incidentportability.ImportedAttribution {
	return append([]incidentportability.ImportedAttribution(nil), r.rows...)
}
