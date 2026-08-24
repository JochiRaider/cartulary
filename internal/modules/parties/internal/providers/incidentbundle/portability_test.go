package incidentbundle

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
	"github.com/JochiRaider/cartulary/internal/modules/incidentportability"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
)

func TestPartyIncidentBundlePreparationAndInvariantPrecedence_Unit(t *testing.T) {
	t.Parallel()
	incidentID := uuid.MustParse("11111111-1111-4111-8111-111111111111")
	recordID := uuid.MustParse("22222222-2222-4222-8222-222222222222")
	importContext := partyPortableImportContext(incidentID, "party-prepare")
	port := NewContribution()

	for _, payload := range [][]byte{
		{},
		encodePartyPortableRows(t, partyPortableTestRow(recordID, incidentID, "person")),
	} {
		if _, err := port.PrepareImport(context.Background(), sourceport.MapBundle{partyIncidentBundlePath: payload}, importContext); err != nil {
			t.Fatalf("valid Party preparation: %v", err)
		}
	}

	base := partyPortableTestRow(recordID, incidentID, "organization")
	tests := []struct {
		name      string
		payload   []byte
		invariant string
	}{
		{
			name: "missing identity",
			payload: encodePartyPortableRows(t, mutatePartyPortableRow(base, func(row map[string]any) {
				delete(row, "record_id")
			})),
			invariant: "parties.source_identity_admitted",
		},
		{
			name: "noncanonical identity",
			payload: encodePartyPortableRows(t, mutatePartyPortableRow(base, func(row map[string]any) {
				row["record_id"] = "22222222-2222-4222-8222-22222222222A"
			})),
			invariant: "parties.source_identity_admitted",
		},
		{
			name:      "duplicate identity",
			payload:   encodePartyPortableRows(t, base, base),
			invariant: "parties.source_identity_admitted",
		},
		{
			name: "unknown member",
			payload: encodePartyPortableRows(t, mutatePartyPortableRow(base, func(row map[string]any) {
				row["created_at"] = "2026-08-24T00:00:00Z"
			})),
			invariant: "parties.version_shape_exact",
		},
		{
			name: "omitted optional",
			payload: encodePartyPortableRows(t, mutatePartyPortableRow(base, func(row map[string]any) {
				delete(row, "notes")
			})),
			invariant: "parties.version_shape_exact",
		},
		{
			name: "wrong type",
			payload: encodePartyPortableRows(t, mutatePartyPortableRow(base, func(row map[string]any) {
				row["timezone_name"] = 3
			})),
			invariant: "parties.version_shape_exact",
		},
		{
			name: "duplicate member",
			payload: []byte(fmt.Sprintf(
				`{"record_id":%q,"record_id":%q,"incident_id":%q,"display_name":"Party","party_kind":"organization","organization_name":null,"role_title":null,"primary_email":null,"timezone_name":null,"external_ref":null,"notes":null}`+"\n",
				recordID.String(), recordID.String(), incidentID.String(),
			)),
			invariant: "parties.version_shape_exact",
		},
		{
			name:      "blank line",
			payload:   append(encodePartyPortableRows(t, base), '\n'),
			invariant: "parties.version_shape_exact",
		},
		{
			name:      "trailing content",
			payload:   append(bytes.TrimSuffix(encodePartyPortableRows(t, base), []byte{'\n'}), []byte(" true\n")...),
			invariant: "parties.version_shape_exact",
		},
		{
			name: "required field",
			payload: encodePartyPortableRows(t, mutatePartyPortableRow(base, func(row map[string]any) {
				row["display_name"] = ""
			})),
			invariant: "parties.identity_lifecycle",
		},
		{
			name: "party kind",
			payload: encodePartyPortableRows(t, mutatePartyPortableRow(base, func(row map[string]any) {
				row["party_kind"] = "Organization"
			})),
			invariant: "parties.identity_lifecycle",
		},
		{
			name: "display normalization",
			payload: encodePartyPortableRows(t, mutatePartyPortableRow(base, func(row map[string]any) {
				row["display_name"] = " Party "
			})),
			invariant: "parties.normalization_exact",
		},
		{
			name: "timezone registry",
			payload: encodePartyPortableRows(t, mutatePartyPortableRow(base, func(row map[string]any) {
				row["timezone_name"] = "utc"
			})),
			invariant: "parties.normalization_exact",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := port.PrepareImport(
				context.Background(),
				sourceport.MapBundle{partyIncidentBundlePath: test.payload},
				importContext,
			)
			requirePartyPortableFailure(t, err, test.invariant)
		})
	}

	lowerRecordID := uuid.MustParse("22222222-2222-4222-8222-222222222221")
	shapeDefect := mutatePartyPortableRow(partyPortableTestRow(recordID, incidentID, "team"), func(row map[string]any) {
		delete(row, "notes")
	})
	normalizationDefect := mutatePartyPortableRow(partyPortableTestRow(lowerRecordID, incidentID, "team"), func(row map[string]any) {
		row["display_name"] = " Party "
	})
	for _, rows := range [][]map[string]any{
		{shapeDefect, normalizationDefect},
		{normalizationDefect, shapeDefect},
	} {
		_, err := port.PrepareImport(
			context.Background(),
			sourceport.MapBundle{partyIncidentBundlePath: encodePartyPortableRows(t, rows...)},
			importContext,
		)
		requirePartyPortableFailure(t, err, "parties.version_shape_exact")
	}

	if _, err := port.PrepareImport(
		context.Background(),
		sourceport.MapBundle{partyIncidentBundlePath: encodePartyPortableRows(t, base)},
		sourceport.ImportContext{
			IncidentID: incidentID, ActorUserID: importContext.ActorUserID,
			BundleVersion: 2, OperationID: "retired-version",
		},
	); !errors.Is(err, sourceport.ErrPreparedBinding) {
		t.Fatalf("version 2 preparation error = %v; want ErrPreparedBinding", err)
	}
}

func TestPartyIncidentBundleExactStateClaimsAndFailureAtomicity_Integration(t *testing.T) {
	harness := pgtest.Start(t)
	db := harness.BeginRollbackDBT(t, "party-incident-bundle-v3")
	ctx := context.Background()
	actorID := uuid.MustParse("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa")
	incidentID := uuid.MustParse("bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb")
	seedPartyPortableIncident(t, ctx, db, actorID, incidentID)

	recordIDs := []uuid.UUID{
		uuid.MustParse("10000000-0000-4000-8000-000000000001"),
		uuid.MustParse("10000000-0000-4000-8000-000000000002"),
		uuid.MustParse("10000000-0000-4000-8000-000000000003"),
		uuid.MustParse("10000000-0000-4000-8000-000000000004"),
		uuid.MustParse("10000000-0000-4000-8000-000000000005"),
	}
	kinds := []string{"person", "team", "organization", "distribution_list", "other"}
	portableRows := make([]map[string]any, 0, len(recordIDs))
	for index, recordID := range recordIDs {
		row := partyPortableTestRow(recordID, incidentID, kinds[index])
		if index == 0 {
			row["display_name"] = "Álpha Party"
			row["organization_name"] = "Coordination"
			row["role_title"] = "Lead"
			row["primary_email"] = "Owner@Example.Test"
			row["timezone_name"] = "US/Eastern"
			row["external_ref"] = "EXT-Alpha"
			row["notes"] = "Line one\n\nLine two"
		}
		portableRows = append(portableRows, row)
	}
	payload := encodePartyPortableRows(t, portableRows...)
	port := NewContribution()
	importContext := partyPortableImportContext(incidentID, "party-apply")
	importContext.ActorUserID = actorID
	prepared, err := port.PrepareImport(ctx, sourceport.MapBundle{partyIncidentBundlePath: payload}, importContext)
	if err != nil {
		t.Fatalf("prepare Party rows: %v", err)
	}
	var preApplyCount int
	if err := db.QueryRow(ctx, `SELECT count(*) FROM parties WHERE incident_id = $1`, incidentID).Scan(&preApplyCount); err != nil || preApplyCount != 0 {
		t.Fatalf("prepare changed Party state: count=%d err=%v", preApplyCount, err)
	}

	tx, err := db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		t.Fatalf("begin Party import transaction: %v", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	for _, recordID := range recordIDs {
		if _, err := tx.Exec(ctx, `
INSERT INTO records (
    record_id, incident_id, record_type, created_by_user_id, updated_by_user_id
) VALUES ($1, $2, 'party', $3, $3)
`, recordID, incidentID, actorID); err != nil {
			t.Fatalf("seed Party envelope: %v", err)
		}
	}
	if err := port.ApplyImportTx(ctx, tx, prepared, importContext); err != nil {
		t.Fatalf("apply Party rows: %v", err)
	}
	if err := port.ValidateImportTx(ctx, tx, prepared, importContext); err != nil {
		t.Fatalf("validate Party rows: %v", err)
	}
	exported, err := port.Export(ctx, sourceport.ExportContext{Query: tx, IncidentID: incidentID})
	if err != nil {
		t.Fatalf("export Party rows: %v", err)
	}
	if len(exported) != 1 || exported[0].Path != partyIncidentBundlePath || !bytes.Equal(exported[0].Payload, payload) {
		t.Fatalf("Party semantic byte round trip drift:\nwant=%s\ngot=%s", payload, exported[0].Payload)
	}
	var claimCount int
	if err := tx.QueryRow(ctx, `SELECT count(*) FROM party_active_key_claims WHERE incident_id = $1`, incidentID).Scan(&claimCount); err != nil || claimCount != 2 {
		t.Fatalf("Party exact claim count = %d, err=%v; want 2", claimCount, err)
	}

	if _, err := tx.Exec(ctx, `DELETE FROM party_active_key_claims WHERE incident_id = $1`, incidentID); err != nil {
		t.Fatalf("remove Party claim sentinel: %v", err)
	}
	requirePartyPortableFailure(t, port.ValidateImportTx(ctx, tx, prepared, importContext), "parties.identity_lifecycle")
	if _, err := tx.Exec(ctx, `SELECT parties_refresh_active_key_claims_v1($1)`, recordIDs[0]); err != nil {
		t.Fatalf("restore Party claim sentinel: %v", err)
	}
	if _, err := tx.Exec(ctx, `UPDATE records SET record_type = 'host' WHERE record_id = $1`, recordIDs[0]); err != nil {
		t.Fatalf("set Party envelope sentinel: %v", err)
	}
	requirePartyPortableFailure(t, port.ValidateImportTx(ctx, tx, prepared, importContext), "parties.envelope_type_scope")
	if _, err := tx.Exec(ctx, `UPDATE records SET record_type = 'party' WHERE record_id = $1`, recordIDs[0]); err != nil {
		t.Fatalf("restore Party envelope sentinel: %v", err)
	}
	if _, err := tx.Exec(ctx, `UPDATE parties SET display_name = ' Noncanonical ' WHERE record_id = $1`, recordIDs[0]); err != nil {
		t.Fatalf("set Party normalization sentinel: %v", err)
	}
	requirePartyPortableFailure(t, port.ValidateImportTx(ctx, tx, prepared, importContext), "parties.normalization_exact")

	canceledContext, cancel := context.WithCancel(ctx)
	cancel()
	if err := port.ValidateImportTx(canceledContext, tx, prepared, importContext); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancelled Party validation error = %v; want context.Canceled", err)
	}

	wrongContext := importContext
	wrongContext.BundleVersion = 2
	if err := port.ApplyImportTx(ctx, tx, prepared, wrongContext); !errors.Is(err, sourceport.ErrPreparedBinding) {
		t.Fatalf("cross-version prepared apply error = %v; want ErrPreparedBinding", err)
	}
}

func TestPartyIncidentBundleClaimCollisionRollsBack_Integration(t *testing.T) {
	harness := pgtest.Start(t)
	db := harness.BeginRollbackDBT(t, "party-incident-bundle-collision")
	ctx := context.Background()
	actorID := uuid.MustParse("cccccccc-cccc-4ccc-8ccc-cccccccccccc")
	incidentID := uuid.MustParse("dddddddd-dddd-4ddd-8ddd-dddddddddddd")
	seedPartyPortableIncident(t, ctx, db, actorID, incidentID)
	recordIDs := []uuid.UUID{
		uuid.MustParse("20000000-0000-4000-8000-000000000001"),
		uuid.MustParse("20000000-0000-4000-8000-000000000002"),
	}
	rows := []map[string]any{
		partyPortableTestRow(recordIDs[0], incidentID, "person"),
		partyPortableTestRow(recordIDs[1], incidentID, "person"),
	}
	rows[0]["primary_email"] = "same@example.test"
	rows[1]["primary_email"] = "SAME@example.test"
	port := NewContribution()
	importContext := partyPortableImportContext(incidentID, "party-collision")
	importContext.ActorUserID = actorID
	prepared, err := port.PrepareImport(ctx, sourceport.MapBundle{
		partyIncidentBundlePath: encodePartyPortableRows(t, rows...),
	}, importContext)
	if err != nil {
		t.Fatalf("prepare colliding Party rows: %v", err)
	}
	tx, err := db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		t.Fatalf("begin collision transaction: %v", err)
	}
	for _, recordID := range recordIDs {
		if _, err := tx.Exec(ctx, `
INSERT INTO records (
    record_id, incident_id, record_type, created_by_user_id, updated_by_user_id
) VALUES ($1, $2, 'party', $3, $3)
`, recordID, incidentID, actorID); err != nil {
			t.Fatalf("seed collision envelope: %v", err)
		}
	}
	requirePartyPortableFailure(t, port.ApplyImportTx(ctx, tx, prepared, importContext), "parties.identity_lifecycle")
	if err := tx.Rollback(ctx); err != nil {
		t.Fatalf("roll back collision transaction: %v", err)
	}
	var partiesCount, claimsCount, recordsCount int
	if err := db.QueryRow(ctx, `SELECT count(*) FROM parties WHERE incident_id = $1`, incidentID).Scan(&partiesCount); err != nil {
		t.Fatalf("count rolled-back Parties: %v", err)
	}
	if err := db.QueryRow(ctx, `SELECT count(*) FROM party_active_key_claims WHERE incident_id = $1`, incidentID).Scan(&claimsCount); err != nil {
		t.Fatalf("count rolled-back Party claims: %v", err)
	}
	if err := db.QueryRow(ctx, `SELECT count(*) FROM records WHERE incident_id = $1`, incidentID).Scan(&recordsCount); err != nil {
		t.Fatalf("count rolled-back Party envelopes: %v", err)
	}
	if partiesCount != 0 || claimsCount != 0 || recordsCount != 0 {
		t.Fatalf("collision left partial state: parties=%d claims=%d records=%d", partiesCount, claimsCount, recordsCount)
	}
}

func partyPortableImportContext(incidentID uuid.UUID, operationID string) sourceport.ImportContext {
	return sourceport.ImportContext{
		IncidentID:    incidentID,
		ActorUserID:   uuid.MustParse("ffffffff-ffff-4fff-8fff-ffffffffffff"),
		BundleVersion: 3,
		OperationID:   operationID,
	}
}

func partyPortableTestRow(recordID, incidentID uuid.UUID, kind string) map[string]any {
	return map[string]any{
		"record_id":         recordID.String(),
		"incident_id":       incidentID.String(),
		"display_name":      "Party",
		"party_kind":        kind,
		"organization_name": nil,
		"role_title":        nil,
		"primary_email":     nil,
		"timezone_name":     nil,
		"external_ref":      nil,
		"notes":             nil,
	}
}

func mutatePartyPortableRow(row map[string]any, mutate func(map[string]any)) map[string]any {
	cloned := make(map[string]any, len(row)+1)
	for key, value := range row {
		cloned[key] = value
	}
	mutate(cloned)
	return cloned
}

func encodePartyPortableRows(t testing.TB, rows ...map[string]any) []byte {
	t.Helper()
	var payload []byte
	for _, row := range rows {
		encoded, err := incidentportability.CanonicalJSONString(row)
		if err != nil {
			t.Fatalf("encode Party portable row: %v", err)
		}
		payload = append(payload, encoded...)
	}
	return payload
}

func requirePartyPortableFailure(t testing.TB, err error, invariant string) {
	t.Helper()
	var failure *sourceport.Failure
	if !errors.As(err, &failure) || failure.FamilyID() != "parties" || failure.InvariantID() != invariant {
		t.Fatalf("Party portability failure = %#v, %v; want %s", failure, err, invariant)
	}
	if stringsContainAny(err.Error(), "data/parties", "record_id", "party_active_key_claims", "constraint") {
		t.Fatalf("Party portability failure leaked internal detail: %q", err.Error())
	}
}

func stringsContainAny(value string, candidates ...string) bool {
	for _, candidate := range candidates {
		if bytes.Contains([]byte(value), []byte(candidate)) {
			return true
		}
	}
	return false
}

func seedPartyPortableIncident(
	t testing.TB,
	ctx context.Context,
	db interface {
		Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
	},
	actorID uuid.UUID,
	incidentID uuid.UUID,
) {
	t.Helper()
	if _, err := db.Exec(ctx, `
INSERT INTO users (
    id, email, display_name, password_hash, mfa_required, is_active,
    is_deployment_admin
) VALUES ($1, $2, 'Party portable actor', 'fixture-hash', false, true, true)
`, actorID, actorID.String()+"@example.test"); err != nil {
		t.Fatalf("seed Party portable actor: %v", err)
	}
	if _, err := db.Exec(ctx, `
INSERT INTO incidents (
    id, incident_key, incident_key_canonical, title, status,
    created_by_user_id, updated_by_user_id
) VALUES ($1, $2, lower($2), 'Party portable incident', 'active', $3, $3)
`, incidentID, "PARTY-PORTABLE-"+incidentID.String(), actorID); err != nil {
		t.Fatalf("seed Party portable incident: %v", err)
	}
}
