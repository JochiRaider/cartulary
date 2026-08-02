package savedviews

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/incidentportability"
	"github.com/JochiRaider/cartulary/internal/platform/viewquery"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
)

func TestIncidentBundleSavedViewCanonicalExport_Integration(t *testing.T) {
	ctx := context.Background()
	db := pgtest.Start(t).BeginRollbackDBT(t, "saved-view-export-characterization")
	actorID := uuid.MustParse("00000000-0000-4000-8000-000000110301")
	incidentID := uuid.MustParse("00000000-0000-4000-8000-000000110302")
	emptyIncidentID := uuid.MustParse("00000000-0000-4000-8000-000000110303")
	canonicalLayout, layoutErr := viewschema.NormalizeLayout(nil, "cartulary.view.timeline.v2")
	if layoutErr != nil {
		t.Fatalf("build canonical export layout: %#v", layoutErr)
	}
	if _, err := db.Exec(ctx, `
INSERT INTO users (
    id, email, display_name, password_hash, mfa_required,
    is_active, is_deployment_admin, created_at, updated_at
) VALUES (
    $1, 'saved-view-export@example.test', 'Saved View Export', 'test-only',
    false, true, true, '2026-07-02T09:00:00Z', '2026-07-02T09:00:00Z'
)
`, actorID); err != nil {
		t.Fatalf("seed export actor: %v", err)
	}
	for _, incident := range []struct {
		id  uuid.UUID
		key string
	}{
		{id: incidentID, key: "SAVED-VIEW-EXPORT"},
		{id: emptyIncidentID, key: "SAVED-VIEW-EXPORT-EMPTY"},
	} {
		if _, err := db.Exec(ctx, `
INSERT INTO incidents (
    id, incident_key, incident_key_canonical, title, status,
    created_by_user_id, updated_by_user_id, created_at, updated_at
) VALUES (
    $1, $2, $2, 'Saved View export characterization', 'active',
    $3, $3, '2026-07-02T09:30:00Z', '2026-07-02T09:30:00Z'
)
`, incident.id, incident.key, actorID); err != nil {
			t.Fatalf("seed export incident %s: %v", incident.key, err)
		}
	}
	rows := []struct {
		id        string
		scope     string
		name      string
		owner     any
		createdAt string
		updatedAt string
		version   int
	}{
		{
			id: "00000000-0000-4000-8000-000000110311", scope: "private",
			name: "Private portable view", owner: actorID,
			createdAt: "2026-07-02T10:11:12Z", updatedAt: "2026-07-02T10:11:13Z", version: 1,
		},
		{
			id: "00000000-0000-4000-8000-000000110312", scope: "shared",
			name: "Shared portable view", owner: actorID,
			createdAt: "2026-07-02T10:12:12Z", updatedAt: "2026-07-02T10:12:13Z", version: 2,
		},
		{
			id: "00000000-0000-4000-8000-000000110313", scope: "system",
			name: "System portable view", owner: nil,
			createdAt: "2026-07-02T10:13:12Z", updatedAt: "2026-07-02T10:13:13Z", version: 3,
		},
	}
	for _, row := range rows {
		if _, err := db.Exec(ctx, `
INSERT INTO saved_views (
    saved_view_id, incident_id, view_schema_id, scope, display_name,
    query_json, layout_json, owner_user_id, created_at, updated_at,
    saved_view_version
) VALUES (
    $1, $2, 'cartulary.view.timeline.v2', $3, $4,
    '{"filters":[],"sort":[]}'::jsonb,
    $9::jsonb,
    $5, $6::timestamptz, $7::timestamptz, $8
)
`, row.id, incidentID, row.scope, row.name, row.owner, row.createdAt, row.updatedAt, row.version, canonicalLayout); err != nil {
			t.Fatalf("seed %s saved view: %v", row.scope, err)
		}
	}

	first, err := exportIncidentBundleFiles(ctx, savedViewExportContext{Query: db, IncidentID: incidentID})
	if err != nil {
		t.Fatalf("first saved-view export: %v", err)
	}
	second, err := exportIncidentBundleFiles(ctx, savedViewExportContext{Query: db, IncidentID: incidentID})
	if err != nil {
		t.Fatalf("second saved-view export: %v", err)
	}
	if len(first) != 1 || first[0].Path != "data/saved_views.ndjson" {
		t.Fatalf("saved-view export files = %#v", first)
	}
	if len(second) != 1 || !bytes.Equal(first[0].Payload, second[0].Payload) {
		t.Fatalf("identical saved-view state did not export byte-identically")
	}
	exportedRows, err := incidentportability.DecodeNDJSON(first[0].Payload)
	if err != nil {
		t.Fatalf("decode characterized export: %v", err)
	}
	if len(exportedRows) != len(rows) {
		t.Fatalf("exported row count = %d; want %d", len(exportedRows), len(rows))
	}
	expectedKeys := []string{
		"created_at", "display_name", "incident_id", "layout_json",
		"owner_user_id", "query_json", "saved_view_id", "saved_view_version",
		"scope", "updated_at", "view_schema_id",
	}
	for index, row := range exportedRows {
		keys := make([]string, 0, len(row))
		for key := range row {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		if !reflect.DeepEqual(keys, expectedKeys) {
			t.Fatalf("exported row %d keys = %v; want %v", index+1, keys, expectedKeys)
		}
		if row["saved_view_id"] != rows[index].id || row["scope"] != rows[index].scope {
			t.Fatalf("exported row %d identity/scope = %#v", index+1, row)
		}
		if _, ok := row["id"]; ok {
			t.Fatalf("exported row %d contains advisory id alias", index+1)
		}
		if _, ok := row["view_scope"]; ok {
			t.Fatalf("exported row %d contains advisory view_scope alias", index+1)
		}
		if row["created_at"] != rows[index].createdAt ||
			row["updated_at"] != rows[index].updatedAt {
			t.Fatalf("exported row %d timestamps = %#v", index+1, row)
		}
		if incidentportability.StringFromAny(row["saved_view_version"]) !=
			incidentportability.StringFromAny(rows[index].version) {
			t.Fatalf("exported row %d version = %#v", index+1, row["saved_view_version"])
		}
		if rows[index].owner == nil {
			if row["owner_user_id"] != nil {
				t.Fatalf("system owner = %#v; want null", row["owner_user_id"])
			}
		} else if row["owner_user_id"] != actorID.String() {
			t.Fatalf("%s owner = %#v; want %s", rows[index].scope, row["owner_user_id"], actorID)
		}
	}

	emptyFiles, err := exportIncidentBundleFiles(ctx, savedViewExportContext{Query: db, IncidentID: emptyIncidentID})
	if err != nil {
		t.Fatalf("zero-row saved-view export: %v", err)
	}
	if len(emptyFiles) != 1 ||
		emptyFiles[0].Path != "data/saved_views.ndjson" ||
		len(emptyFiles[0].Payload) != 0 {
		t.Fatalf("zero-row saved-view export = %#v; want required zero-byte member", emptyFiles)
	}
}

func TestIncidentBundleSavedViewPortableOwnerRoundTripAndRollback_Integration(t *testing.T) {
	ctx := context.Background()
	db := pgtest.Start(t).BeginRollbackDBT(t, "saved-view-portable-owner-round-trip")
	targetActorID := uuid.MustParse("00000000-0000-4000-8000-000000110451")
	sourceActorID := uuid.MustParse("00000000-0000-4000-8000-000000110403")
	incidentID := strictSavedViewIncidentID
	if _, err := db.Exec(ctx, `
INSERT INTO users (
    id, email, display_name, password_hash, mfa_required,
    is_active, is_deployment_admin, created_at, updated_at
) VALUES (
    $1, 'saved-view-import-target@example.test', 'Saved View Import Target', 'test-only',
    false, true, true, '2026-07-02T09:00:00Z', '2026-07-02T09:00:00Z'
)
`, targetActorID); err != nil {
		t.Fatalf("seed target actor: %v", err)
	}
	if _, err := db.Exec(ctx, `
INSERT INTO incidents (
    id, incident_key, incident_key_canonical, title, status,
    created_by_user_id, updated_by_user_id, created_at, updated_at
) VALUES (
    $1, 'SAVED-VIEW-PORTABLE-OWNER', 'SAVED-VIEW-PORTABLE-OWNER',
    'Saved View portable owner', 'active',
    $2, $2, '2026-07-02T09:30:00Z', '2026-07-02T09:30:00Z'
)
`, incidentID, targetActorID); err != nil {
		t.Fatalf("seed target incident: %v", err)
	}

	privateRow := validPortableSavedViewRow(t)
	sharedRow := validPortableSavedViewRow(t)
	sharedRow["saved_view_id"] = "00000000-0000-4000-8000-000000110404"
	sharedRow["scope"] = "shared"
	sharedRow["display_name"] = "Portable shared saved view"
	systemRow := validPortableSavedViewRow(t)
	systemRow["saved_view_id"] = "00000000-0000-4000-8000-000000110405"
	systemRow["scope"] = "system"
	systemRow["display_name"] = "Portable system saved view"
	systemRow["owner_user_id"] = nil
	payload := append([]byte(nil), encodePortableSavedViewRow(t, privateRow)...)
	payload = append(payload, encodePortableSavedViewRow(t, sharedRow)...)
	payload = append(payload, encodePortableSavedViewRow(t, systemRow)...)
	attributions := &recordingAttributionRecorder{localUserID: targetActorID}
	importContext := savedViewImportContext{
		IncidentID:    incidentID,
		ActorUserID:   targetActorID,
		Attributions:  attributions,
		ActorAdmitted: savedViewActorAdmission(sourceActorID),
	}
	prepared, err := prepareSavedViewImport(
		savedViewMapBundle{"data/saved_views.ndjson": payload},
		importContext,
	)
	if err != nil {
		t.Fatalf("prepare portable owner row: %v", err)
	}
	tx, err := db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		t.Fatalf("begin portable owner transaction: %v", err)
	}
	t.Cleanup(func() { _ = tx.Rollback(ctx) })
	if err := applyPreparedSavedViewImportTx(ctx, tx, prepared, importContext); err != nil {
		t.Fatalf("apply portable owner row: %v", err)
	}
	if err := validatePreparedSavedViewImportTx(ctx, tx, prepared, importContext); err != nil {
		t.Fatalf("validate portable owner row: %v", err)
	}

	runtimeRows, err := tx.Query(ctx, `
SELECT saved_view_id, scope, owner_user_id
  FROM saved_views
 WHERE incident_id = $1
 ORDER BY saved_view_id
`, incidentID)
	if err != nil {
		t.Fatalf("load runtime saved-view owners: %v", err)
	}
	defer runtimeRows.Close()
	runtimeOwners := map[string]*uuid.UUID{}
	for runtimeRows.Next() {
		var id string
		var scope string
		var owner *uuid.UUID
		if err := runtimeRows.Scan(&id, &scope, &owner); err != nil {
			t.Fatalf("scan runtime saved-view owner: %v", err)
		}
		runtimeOwners[scope] = owner
	}
	if err := runtimeRows.Err(); err != nil {
		t.Fatalf("iterate runtime saved-view owners: %v", err)
	}
	for _, scoped := range []string{"private", "shared"} {
		owner := runtimeOwners[scoped]
		if owner == nil || *owner != targetActorID {
			t.Fatalf("%s runtime owner = %v; want target actor %s", scoped, owner, targetActorID)
		}
	}
	if runtimeOwners["system"] != nil {
		t.Fatalf("system runtime owner = %v; want nil", runtimeOwners["system"])
	}
	if len(attributions.entries) != 2 {
		t.Fatalf("portable owner attribution = %#v; want source actor %s", attributions.entries, sourceActorID)
	}
	for _, entry := range attributions.entries {
		if entry.sourceActorID != sourceActorID.String() {
			t.Fatalf("portable owner attribution = %#v; want source actor %s", attributions.entries, sourceActorID)
		}
	}

	exported, err := exportIncidentBundleFiles(ctx, savedViewExportContext{
		Query:                tx,
		IncidentID:           incidentID,
		PortableAttributions: attributions,
	})
	if err != nil {
		t.Fatalf("re-export imported saved view: %v", err)
	}
	exportedRows, err := incidentportability.DecodeNDJSON(exported[0].Payload)
	if err != nil {
		t.Fatalf("decode re-exported saved view: %v", err)
	}
	if len(exportedRows) != 3 ||
		exportedRows[0]["owner_user_id"] != sourceActorID.String() ||
		exportedRows[1]["owner_user_id"] != sourceActorID.String() ||
		exportedRows[2]["owner_user_id"] != nil {
		t.Fatalf("re-exported portable owners = %#v; want source/source/null", exportedRows)
	}

	if err := tx.Rollback(ctx); err != nil {
		t.Fatalf("rollback owner round-trip transaction: %v", err)
	}
	var persisted int
	if err := db.QueryRow(ctx, `SELECT count(*) FROM saved_views WHERE incident_id = $1`, incidentID).Scan(&persisted); err != nil {
		t.Fatalf("count rolled-back saved views: %v", err)
	}
	if persisted != 0 {
		t.Fatalf("rollback retained %d saved-view rows", persisted)
	}

	validationCases := map[string]struct {
		invariant string
		mutate    func(pgx.Tx, *recordingAttributionRecorder) error
	}{
		"row count": {
			invariant: "saved_views.row_shape_exact",
			mutate: func(tx pgx.Tx, _ *recordingAttributionRecorder) error {
				_, err := tx.Exec(ctx, `DELETE FROM saved_views WHERE saved_view_id = $1`, privateRow["saved_view_id"])
				return err
			},
		},
		"identity and schema": {
			invariant: "saved_views.identity_scope_legal",
			mutate: func(tx pgx.Tx, _ *recordingAttributionRecorder) error {
				_, err := tx.Exec(ctx, `
UPDATE saved_views
   SET view_schema_id = 'cartulary.view.notes.v1'
 WHERE saved_view_id = $1
`, privateRow["saved_view_id"])
				return err
			},
		},
		"portable owner attribution": {
			invariant: "saved_views.owner_tuple_legal",
			mutate: func(_ pgx.Tx, recorder *recordingAttributionRecorder) error {
				recorder.entries[0].sourceActorID = uuid.NewString()
				return nil
			},
		},
		"display name": {
			invariant: "saved_views.display_name_normalized",
			mutate: func(tx pgx.Tx, _ *recordingAttributionRecorder) error {
				_, err := tx.Exec(ctx, `
UPDATE saved_views SET display_name = 'Changed after admission' WHERE saved_view_id = $1
`, privateRow["saved_view_id"])
				return err
			},
		},
		"query layout": {
			invariant: "saved_views.query_layout_legal",
			mutate: func(tx pgx.Tx, _ *recordingAttributionRecorder) error {
				_, err := tx.Exec(ctx, `
UPDATE saved_views
   SET query_json = '{"filters":[],"group_by":"timeline.capture_state","sort":[]}'::jsonb
 WHERE saved_view_id = $1
`, privateRow["saved_view_id"])
				return err
			},
		},
		"version timestamps": {
			invariant: "saved_views.version_timestamps_legal",
			mutate: func(tx pgx.Tx, _ *recordingAttributionRecorder) error {
				_, err := tx.Exec(ctx, `
UPDATE saved_views SET saved_view_version = saved_view_version + 1 WHERE saved_view_id = $1
`, privateRow["saved_view_id"])
				return err
			},
		},
	}
	for name, testCase := range validationCases {
		t.Run("validation "+name, func(t *testing.T) {
			recorder := &recordingAttributionRecorder{localUserID: targetActorID}
			caseContext := importContext
			caseContext.Attributions = recorder
			casePrepared, err := prepareSavedViewImport(
				savedViewMapBundle{
					"data/saved_views.ndjson": encodePortableSavedViewRow(t, privateRow),
				},
				caseContext,
			)
			if err != nil {
				t.Fatalf("prepare validation case: %v", err)
			}
			caseTx, err := db.BeginTx(ctx, pgx.TxOptions{})
			if err != nil {
				t.Fatalf("begin validation case: %v", err)
			}
			t.Cleanup(func() { _ = caseTx.Rollback(ctx) })
			if err := applyPreparedSavedViewImportTx(ctx, caseTx, casePrepared, caseContext); err != nil {
				t.Fatalf("apply validation case: %v", err)
			}
			if err := testCase.mutate(caseTx, recorder); err != nil {
				t.Fatalf("mutate validation case: %v", err)
			}
			err = validatePreparedSavedViewImportTx(ctx, caseTx, casePrepared, caseContext)
			requireSavedViewInvariantFailure(t, err, testCase.invariant)
			if err := caseTx.Rollback(ctx); err != nil {
				t.Fatalf("rollback validation case: %v", err)
			}
		})
	}

	t.Run("duplicate target row fails affected-row admission", func(t *testing.T) {
		recorder := &recordingAttributionRecorder{localUserID: targetActorID}
		caseContext := importContext
		caseContext.Attributions = recorder
		casePrepared, err := prepareSavedViewImport(
			savedViewMapBundle{
				"data/saved_views.ndjson": encodePortableSavedViewRow(t, privateRow),
			},
			caseContext,
		)
		if err != nil {
			t.Fatalf("prepare duplicate-target case: %v", err)
		}
		caseTx, err := db.BeginTx(ctx, pgx.TxOptions{})
		if err != nil {
			t.Fatalf("begin duplicate-target case: %v", err)
		}
		t.Cleanup(func() { _ = caseTx.Rollback(ctx) })
		if _, err := caseTx.Exec(ctx, `
INSERT INTO saved_views (
    saved_view_id, incident_id, view_schema_id, scope, display_name,
    query_json, layout_json, owner_user_id, created_at, updated_at,
    saved_view_version
)
VALUES (
    $1, $2, 'cartulary.view.timeline.v2', 'private', 'Existing row',
    '{"filters":[],"sort":[]}'::jsonb,
    '{"column_order":[],"column_widths":[],"hidden_field_keys":[],"layout_schema_id":"cartulary.layout.v1"}'::jsonb,
    $3, '2026-07-02T00:00:00Z', '2026-07-02T00:00:00Z', 1
)
`, privateRow["saved_view_id"], incidentID, targetActorID); err != nil {
			t.Fatalf("seed duplicate-target row: %v", err)
		}
		err = applyPreparedSavedViewImportTx(ctx, caseTx, casePrepared, caseContext)
		requireSavedViewInvariantFailure(t, err, "saved_views.identity_scope_legal")
		if len(recorder.entries) != 0 {
			t.Fatalf("failed insert recorded attribution: %#v", recorder.entries)
		}
		if err := caseTx.Rollback(ctx); err != nil {
			t.Fatalf("rollback duplicate-target case: %v", err)
		}
	})
}

type attributionEntry struct {
	table         string
	sourceRowID   string
	column        string
	sourceActorID string
}

type recordingAttributionRecorder struct {
	localUserID uuid.UUID
	entries     []attributionEntry
}

func (r *recordingAttributionRecorder) RecordImportedAttribution(table string, sourceRowID string, column string, sourceActorID string) error {
	r.entries = append(r.entries, attributionEntry{
		table:         table,
		sourceRowID:   sourceRowID,
		column:        column,
		sourceActorID: sourceActorID,
	})
	return nil
}

func (r *recordingAttributionRecorder) ImportedAttributions() []incidentportability.ImportedAttribution {
	result := make([]incidentportability.ImportedAttribution, 0, len(r.entries))
	for _, entry := range r.entries {
		result = append(result, incidentportability.ImportedAttribution{
			SourceTable:   entry.table,
			SourceRowID:   entry.sourceRowID,
			SourceColumn:  entry.column,
			SourceActorID: entry.sourceActorID,
			LocalUserID:   r.localUserID,
		})
	}
	return result
}

func (r *recordingAttributionRecorder) ResolvePortableSourceActors(
	_ context.Context,
	_ incidentportability.Queryer,
	_ uuid.UUID,
	table string,
	column string,
	sourceRowIDs []string,
) (map[string]string, error) {
	requested := map[string]struct{}{}
	for _, rowID := range sourceRowIDs {
		requested[rowID] = struct{}{}
	}
	resolved := map[string]string{}
	for _, entry := range r.entries {
		if entry.table != table || entry.column != column {
			continue
		}
		if _, ok := requested[entry.sourceRowID]; ok {
			resolved[entry.sourceRowID] = entry.sourceActorID
		}
	}
	return resolved, nil
}

func TestIncidentBundleSavedViewStrictPrepareFramingAndShape_Unit(t *testing.T) {
	importContext := strictSavedViewImportContext(t)
	canonicalRow := validPortableSavedViewRow(t)
	canonicalLine := encodePortableSavedViewRow(t, canonicalRow)

	if _, err := prepareSavedViewImport(
		savedViewMapBundle{"data/saved_views.ndjson": canonicalLine},
		importContext,
	); err != nil {
		t.Fatalf("canonical saved-view row rejected: %v", err)
	}

	rowWith := func(mutate func(map[string]any)) []byte {
		row := validPortableSavedViewRow(t)
		mutate(row)
		return encodePortableSavedViewRow(t, row)
	}
	filterLine := rowWith(func(row map[string]any) {
		row["query_json"] = map[string]any{
			"filters": []any{map[string]any{
				"field_key": "timeline.capture_state",
				"op":        "eq",
				"arg":       map[string]any{"value": "rough"},
			}},
			"sort": []any{},
		}
	})
	sortLine := rowWith(func(row map[string]any) {
		row["query_json"] = map[string]any{
			"filters": []any{},
			"sort": []any{map[string]any{
				"field_key": "timeline.activity_synopsis_text",
				"direction": "asc",
			}},
		}
	})
	widthLine := rowWith(func(row map[string]any) {
		row["layout_json"].(map[string]any)["column_widths"] = []any{map[string]any{
			"field_key": "timeline.activity_synopsis_text",
			"width_px":  json.Number("240"),
		}}
	})
	cases := map[string][]byte{
		"blank logical line": []byte("\n"),
		"two values one line": append(
			append(append([]byte(nil), bytes.TrimSpace(canonicalLine)...), ' '),
			bytes.TrimSpace(canonicalLine)...,
		),
		"trailing content": append(append([]byte(nil), bytes.TrimSpace(canonicalLine)...), []byte(" true\n")...),
		"malformed object": []byte("{\"saved_view_id\":\n"),
		"over bounded line": append(
			append([]byte(`{"saved_view_id":"`), bytes.Repeat([]byte("a"), 16*1024*1024)...),
			[]byte(`"}`)...,
		),
		"id alias": rowWith(func(row map[string]any) {
			row["id"] = row["saved_view_id"]
			delete(row, "saved_view_id")
		}),
		"view scope alias": rowWith(func(row map[string]any) {
			row["view_scope"] = row["scope"]
			delete(row, "scope")
		}),
		"unknown top member": rowWith(func(row map[string]any) {
			row["hostile_unknown_member"] = true
		}),
		"missing member": rowWith(func(row map[string]any) {
			delete(row, "layout_json")
		}),
		"wrong top type": rowWith(func(row map[string]any) {
			row["display_name"] = []any{"Portable saved view"}
		}),
		"null nonnullable": rowWith(func(row map[string]any) {
			row["query_json"] = nil
		}),
		"duplicate top member": append(
			append(append([]byte(nil), bytes.TrimSuffix(bytes.TrimSpace(canonicalLine), []byte("}"))...),
				[]byte(`,"saved_view_id":"00000000-0000-4000-8000-000000110401"}`)...),
			'\n',
		),
		"duplicate query member": append(
			bytes.Replace(
				append([]byte(nil), bytes.TrimSpace(canonicalLine)...),
				[]byte(`"query_json":{"filters":[],"sort":[]}`),
				[]byte(`"query_json":{"filters":[],"filters":[],"sort":[]}`),
				1,
			),
			'\n',
		),
		"unknown query member": rowWith(func(row map[string]any) {
			row["query_json"].(map[string]any)["cursor"] = "forbidden"
		}),
		"missing query member": rowWith(func(row map[string]any) {
			delete(row["query_json"].(map[string]any), "filters")
		}),
		"wrong query member type": rowWith(func(row map[string]any) {
			row["query_json"].(map[string]any)["sort"] = nil
		}),
		"duplicate sort member": append(
			bytes.Replace(
				append([]byte(nil), bytes.TrimSpace(sortLine)...),
				[]byte(`"field_key":"timeline.activity_synopsis_text"`),
				[]byte(`"field_key":"timeline.activity_synopsis_text","field_key":"timeline.activity_synopsis_text"`),
				1,
			),
			'\n',
		),
		"unknown sort member": rowWith(func(row map[string]any) {
			row["query_json"] = map[string]any{
				"filters": []any{},
				"sort": []any{map[string]any{
					"field_key": "timeline.activity_synopsis_text",
					"direction": "asc",
					"unknown":   true,
				}},
			}
		}),
		"missing sort member": rowWith(func(row map[string]any) {
			row["query_json"] = map[string]any{
				"filters": []any{},
				"sort": []any{map[string]any{
					"field_key": "timeline.activity_synopsis_text",
				}},
			}
		}),
		"wrong sort member type": rowWith(func(row map[string]any) {
			row["query_json"] = map[string]any{
				"filters": []any{},
				"sort": []any{map[string]any{
					"field_key": "timeline.activity_synopsis_text",
					"direction": nil,
				}},
			}
		}),
		"duplicate filter member": append(
			bytes.Replace(
				append([]byte(nil), bytes.TrimSpace(filterLine)...),
				[]byte(`"field_key":"timeline.capture_state"`),
				[]byte(`"field_key":"timeline.capture_state","field_key":"timeline.capture_state"`),
				1,
			),
			'\n',
		),
		"unknown filter member": rowWith(func(row map[string]any) {
			row["query_json"] = map[string]any{
				"filters": []any{map[string]any{
					"field_key": "timeline.capture_state",
					"op":        "eq",
					"arg":       map[string]any{"value": "rough"},
					"unknown":   true,
				}},
				"sort": []any{},
			}
		}),
		"missing filter member": rowWith(func(row map[string]any) {
			row["query_json"] = map[string]any{
				"filters": []any{map[string]any{
					"field_key": "timeline.capture_state",
					"op":        "eq",
				}},
				"sort": []any{},
			}
		}),
		"wrong filter member type": rowWith(func(row map[string]any) {
			row["query_json"] = map[string]any{
				"filters": []any{map[string]any{
					"field_key": "timeline.capture_state",
					"op":        nil,
					"arg":       map[string]any{"value": "rough"},
				}},
				"sort": []any{},
			}
		}),
		"duplicate filter arg member": append(
			bytes.Replace(
				append([]byte(nil), bytes.TrimSpace(filterLine)...),
				[]byte(`"value":"rough"`),
				[]byte(`"value":"rough","value":"rough"`),
				1,
			),
			'\n',
		),
		"unknown filter arg member": rowWith(func(row map[string]any) {
			row["query_json"] = map[string]any{
				"filters": []any{map[string]any{
					"field_key": "timeline.capture_state",
					"op":        "eq",
					"arg":       map[string]any{"value": "rough", "unknown": true},
				}},
				"sort": []any{},
			}
		}),
		"wrong filter arg member type": rowWith(func(row map[string]any) {
			row["query_json"] = map[string]any{
				"filters": []any{map[string]any{
					"field_key": "timeline.tags",
					"op":        "contains_any",
					"arg":       map[string]any{"values": "rough"},
				}},
				"sort": []any{},
			}
		}),
		"duplicate layout member": append(
			bytes.Replace(
				append([]byte(nil), bytes.TrimSpace(canonicalLine)...),
				[]byte(`"layout_schema_id":"cartulary.layout.v1"`),
				[]byte(`"layout_schema_id":"cartulary.layout.v1","layout_schema_id":"cartulary.layout.v1"`),
				1,
			),
			'\n',
		),
		"unknown layout member": rowWith(func(row map[string]any) {
			row["layout_json"].(map[string]any)["transient_grid_state"] = true
		}),
		"missing layout member": rowWith(func(row map[string]any) {
			delete(row["layout_json"].(map[string]any), "column_widths")
		}),
		"wrong layout member type": rowWith(func(row map[string]any) {
			row["layout_json"].(map[string]any)["hidden_field_keys"] = nil
		}),
		"duplicate width member": append(
			bytes.Replace(
				append([]byte(nil), bytes.TrimSpace(widthLine)...),
				[]byte(`"width_px":240`),
				[]byte(`"width_px":240,"width_px":240`),
				1,
			),
			'\n',
		),
		"unknown width member": rowWith(func(row map[string]any) {
			row["layout_json"].(map[string]any)["column_widths"] = []any{map[string]any{
				"field_key": "timeline.activity_synopsis_text",
				"width_px":  json.Number("240"),
				"unknown":   true,
			}}
		}),
		"missing width member": rowWith(func(row map[string]any) {
			row["layout_json"].(map[string]any)["column_widths"] = []any{map[string]any{
				"field_key": "timeline.activity_synopsis_text",
			}}
		}),
		"wrong width member type": rowWith(func(row map[string]any) {
			row["layout_json"].(map[string]any)["column_widths"] = []any{map[string]any{
				"field_key": "timeline.activity_synopsis_text",
				"width_px":  "240",
			}}
		}),
	}

	for name, payload := range cases {
		t.Run(name, func(t *testing.T) {
			_, err := prepareSavedViewImport(
				savedViewMapBundle{"data/saved_views.ndjson": payload},
				importContext,
			)
			requireSavedViewInvariantFailure(t, err, "saved_views.row_shape_exact")
			if strings.Contains(strings.ToLower(errorText(err)), "hostile_unknown_member") {
				t.Fatalf("public diagnostic disclosed hostile member name: %v", err)
			}
		})
	}
}

func TestIncidentBundleSavedViewStrictPrepareSemantics_Unit(t *testing.T) {
	importContext := strictSavedViewImportContext(t)
	cases := map[string]struct {
		invariant string
		mutate    func(map[string]any)
	}{
		"invalid saved view uuid": {
			invariant: "saved_views.identity_scope_legal",
			mutate:    func(row map[string]any) { row["saved_view_id"] = "not-a-uuid" },
		},
		"noncanonical saved view uuid": {
			invariant: "saved_views.identity_scope_legal",
			mutate: func(row map[string]any) {
				row["saved_view_id"] = "00000000-0000-4000-8000-00000011040A"
			},
		},
		"foreign incident": {
			invariant: "saved_views.identity_scope_legal",
			mutate:    func(row map[string]any) { row["incident_id"] = uuid.NewString() },
		},
		"unknown schema": {
			invariant: "saved_views.identity_scope_legal",
			mutate:    func(row map[string]any) { row["view_schema_id"] = "cartulary.view.unknown.v1" },
		},
		"invalid scope": {
			invariant: "saved_views.identity_scope_legal",
			mutate:    func(row map[string]any) { row["scope"] = "team" },
		},
		"private owner null": {
			invariant: "saved_views.owner_tuple_legal",
			mutate:    func(row map[string]any) { row["owner_user_id"] = nil },
		},
		"private owner is not uuid": {
			invariant: "saved_views.owner_tuple_legal",
			mutate:    func(row map[string]any) { row["owner_user_id"] = "not-a-uuid" },
		},
		"private owner descriptor missing": {
			invariant: "saved_views.owner_tuple_legal",
			mutate: func(row map[string]any) {
				row["owner_user_id"] = "00000000-0000-4000-8000-000000110498"
			},
		},
		"system owner present": {
			invariant: "saved_views.owner_tuple_legal",
			mutate:    func(row map[string]any) { row["scope"] = "system" },
		},
		"padded display name": {
			invariant: "saved_views.display_name_normalized",
			mutate:    func(row map[string]any) { row["display_name"] = " Portable saved view " },
		},
		"non NFC display name": {
			invariant: "saved_views.display_name_normalized",
			mutate:    func(row map[string]any) { row["display_name"] = "Cafe\u0301" },
		},
		"control display name": {
			invariant: "saved_views.display_name_normalized",
			mutate:    func(row map[string]any) { row["display_name"] = "Portable\u0007 view" },
		},
		"over limit display name": {
			invariant: "saved_views.display_name_normalized",
			mutate:    func(row map[string]any) { row["display_name"] = strings.Repeat("x", 257) },
		},
		"empty query object": {
			invariant: "saved_views.query_layout_legal",
			mutate:    func(row map[string]any) { row["query_json"] = map[string]any{} },
		},
		"null group by": {
			invariant: "saved_views.query_layout_legal",
			mutate: func(row map[string]any) {
				row["query_json"] = map[string]any{"filters": []any{}, "sort": []any{}, "group_by": nil}
			},
		},
		"unknown query field": {
			invariant: "saved_views.query_layout_legal",
			mutate: func(row map[string]any) {
				row["query_json"] = map[string]any{
					"filters": []any{map[string]any{
						"field_key": "timeline.unknown",
						"op":        "eq",
						"arg":       map[string]any{"value": "rough"},
					}},
					"sort": []any{},
				}
			},
		},
		"noncanonical query set": {
			invariant: "saved_views.query_layout_legal",
			mutate: func(row map[string]any) {
				row["query_json"] = map[string]any{
					"filters": []any{map[string]any{
						"field_key": "timeline.tags",
						"op":        "contains_any",
						"arg":       map[string]any{"values": []any{"beta", "alpha", "alpha"}},
					}},
					"sort": []any{},
				}
			},
		},
		"over limit sort": {
			invariant: "saved_views.query_layout_legal",
			mutate: func(row map[string]any) {
				items := make([]any, 9)
				for index := range items {
					items[index] = map[string]any{
						"field_key": "timeline.activity_synopsis_text",
						"direction": "asc",
					}
				}
				row["query_json"] = map[string]any{"filters": []any{}, "sort": items}
			},
		},
		"over limit filters": {
			invariant: "saved_views.query_layout_legal",
			mutate: func(row map[string]any) {
				items := make([]any, 17)
				for index := range items {
					items[index] = map[string]any{
						"field_key": "timeline.capture_state",
						"op":        "eq",
						"arg":       map[string]any{"value": "rough"},
					}
				}
				row["query_json"] = map[string]any{"filters": items, "sort": []any{}}
			},
		},
		"empty layout object": {
			invariant: "saved_views.query_layout_legal",
			mutate:    func(row map[string]any) { row["layout_json"] = map[string]any{} },
		},
		"invalid layout width": {
			invariant: "saved_views.query_layout_legal",
			mutate: func(row map[string]any) {
				row["layout_json"].(map[string]any)["column_widths"] = []any{map[string]any{
					"field_key": "timeline.activity_synopsis_text",
					"width_px":  json.Number("39"),
				}}
			},
		},
		"zero version": {
			invariant: "saved_views.version_timestamps_legal",
			mutate:    func(row map[string]any) { row["saved_view_version"] = json.Number("0") },
		},
		"fractional version": {
			invariant: "saved_views.version_timestamps_legal",
			mutate:    func(row map[string]any) { row["saved_view_version"] = json.Number("1.5") },
		},
		"offset timestamp": {
			invariant: "saved_views.version_timestamps_legal",
			mutate:    func(row map[string]any) { row["created_at"] = "2026-07-01T20:00:00-04:00" },
		},
		"reversed timestamps": {
			invariant: "saved_views.version_timestamps_legal",
			mutate: func(row map[string]any) {
				row["created_at"] = "2026-07-02T00:00:01Z"
				row["updated_at"] = "2026-07-02T00:00:00Z"
			},
		},
		"malformed timestamp": {
			invariant: "saved_views.version_timestamps_legal",
			mutate:    func(row map[string]any) { row["updated_at"] = "not-a-timestamp" },
		},
	}
	for name, testCase := range cases {
		t.Run(name, func(t *testing.T) {
			row := validPortableSavedViewRow(t)
			testCase.mutate(row)
			_, err := prepareSavedViewImport(
				savedViewMapBundle{"data/saved_views.ndjson": encodePortableSavedViewRow(t, row)},
				importContext,
			)
			requireSavedViewInvariantFailure(t, err, testCase.invariant)
		})
	}

	t.Run("duplicate stable identity", func(t *testing.T) {
		line := encodePortableSavedViewRow(t, validPortableSavedViewRow(t))
		payload := append(append([]byte(nil), line...), line...)
		_, err := prepareSavedViewImport(
			savedViewMapBundle{"data/saved_views.ndjson": payload},
			importContext,
		)
		requireSavedViewInvariantFailure(t, err, "saved_views.identity_scope_legal")
	})
}

func validPortableSavedViewRow(t testing.TB) map[string]any {
	t.Helper()
	queryJSON, queryErr := viewquery.NormalizePersisted(json.RawMessage(`{}`), "cartulary.view.timeline.v2")
	if queryErr != nil {
		t.Fatalf("build canonical saved-view query: %#v", queryErr)
	}
	layoutJSON, layoutErr := viewschema.NormalizeLayout(nil, "cartulary.view.timeline.v2")
	if layoutErr != nil {
		t.Fatalf("build canonical saved-view layout: %#v", layoutErr)
	}
	var query any
	if err := json.Unmarshal(queryJSON, &query); err != nil {
		t.Fatalf("decode canonical query fixture: %v", err)
	}
	var layout any
	if err := json.Unmarshal(layoutJSON, &layout); err != nil {
		t.Fatalf("decode canonical layout fixture: %v", err)
	}
	return map[string]any{
		"saved_view_id":      "00000000-0000-4000-8000-000000110401",
		"incident_id":        strictSavedViewIncidentID.String(),
		"view_schema_id":     "cartulary.view.timeline.v2",
		"scope":              "private",
		"display_name":       "Portable saved view",
		"query_json":         query,
		"layout_json":        layout,
		"owner_user_id":      "00000000-0000-4000-8000-000000110403",
		"created_at":         "2026-07-02T00:00:00Z",
		"updated_at":         "2026-07-02T00:00:00Z",
		"saved_view_version": json.Number("1"),
	}
}

var strictSavedViewIncidentID = uuid.MustParse("00000000-0000-4000-8000-000000110402")

type savedViewMapBundle map[string][]byte

func (b savedViewMapBundle) File(path string) ([]byte, bool) {
	payload, ok := b[path]
	return payload, ok
}

func strictSavedViewImportContext(t testing.TB) savedViewImportContext {
	t.Helper()
	return savedViewImportContext{
		IncidentID:  strictSavedViewIncidentID,
		ActorUserID: uuid.MustParse("00000000-0000-4000-8000-000000110499"),
		ActorAdmitted: savedViewActorAdmission(
			uuid.MustParse("00000000-0000-4000-8000-000000110403"),
		),
	}
}

func savedViewActorAdmission(actorIDs ...uuid.UUID) func(string) bool {
	admitted := make(map[string]struct{}, len(actorIDs))
	for _, actorID := range actorIDs {
		admitted[actorID.String()] = struct{}{}
	}
	return func(actorID string) bool {
		_, ok := admitted[actorID]
		return ok
	}
}

func encodePortableSavedViewRow(t testing.TB, row map[string]any) []byte {
	t.Helper()
	payload, err := incidentportability.CanonicalJSONString(row)
	if err != nil {
		t.Fatalf("encode portable saved-view row: %v", err)
	}
	return payload
}

func requireSavedViewInvariantFailure(t testing.TB, err error, invariantID string) {
	t.Helper()
	var failure *savedViewInvariantError
	if !errors.As(err, &failure) ||
		failure.InvariantID != invariantID {
		t.Fatalf("saved-view failure = %T %v; want invariant %s", err, err, invariantID)
	}
}

func errorText(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}
