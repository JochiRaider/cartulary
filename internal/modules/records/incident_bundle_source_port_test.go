package records

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
)

func TestRecordsIncidentBundleSourcePortRejectsDuplicateIdentity_Unit(t *testing.T) {
	recordID := uuid.NewString()
	payload := []byte(
		`{"record_id":"` + recordID + `","incident_id":"` + uuid.NewString() + `","record_type":"timeline_event","row_version":1}` + "\n" +
			`{"record_id":"` + recordID + `","incident_id":"` + uuid.NewString() + `","record_type":"timeline_event","row_version":1}` + "\n",
	)
	port := NewIncidentBundleSourcePort()
	_, err := port.PrepareImport(context.Background(), sourceport.MapBundle{
		"data/records.ndjson": payload,
	}, sourceport.ImportContext{BundleVersion: 2, OperationID: "records-duplicate"})
	var failure *sourceport.Failure
	if !errors.As(err, &failure) || failure.FamilyID != "records" || failure.InvariantID != "records.incident_scope" {
		t.Fatalf("duplicate record identity failure = %#v, %v", failure, err)
	}
}

func TestRecordsIncidentBundleSourcePortAppliesAndRollsBack_Integration(t *testing.T) {
	ctx := context.Background()
	db := pgtest.Start(t).BeginRollbackDBT(t, "records-incident-bundle-source-port")
	actorID := uuid.New()
	incidentID := uuid.New()
	recordID := uuid.New()
	if _, err := db.Exec(ctx, `
INSERT INTO users (id, email, display_name, password_hash, mfa_required, is_active, is_deployment_admin)
VALUES ($1, $2, 'Records Port Actor', 'test-only', false, true, true)
`, actorID, "records-port-"+actorID.String()+"@example.test"); err != nil {
		t.Fatalf("seed records source-port actor: %v", err)
	}
	if _, err := db.Exec(ctx, `
INSERT INTO incidents (
    id, incident_key, incident_key_canonical, title, status,
    created_by_user_id, updated_by_user_id
) VALUES ($1, $2, $2, 'Records source-port incident', 'active', $3, $3)
`, incidentID, "RECORDS-PORT-"+incidentID.String(), actorID); err != nil {
		t.Fatalf("seed records source-port incident: %v", err)
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	payload := []byte(
		`{"record_id":"` + recordID.String() +
			`","incident_id":"` + incidentID.String() +
			`","record_type":"timeline_event","created_by_user_id":"` + actorID.String() +
			`","created_at":"` + now +
			`","updated_by_user_id":"` + actorID.String() +
			`","updated_at":"` + now +
			`","row_version":1}` + "\n",
	)
	port := NewIncidentBundleSourcePort()
	importContext := sourceport.ImportContext{
		IncidentID:    incidentID,
		ActorUserID:   actorID,
		BundleVersion: 2,
		OperationID:   "records-apply-rollback",
	}
	prepared, err := port.PrepareImport(ctx, sourceport.MapBundle{
		"data/records.ndjson": payload,
	}, importContext)
	if err != nil {
		t.Fatalf("prepare records import: %v", err)
	}
	tx, err := db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		t.Fatalf("begin records apply transaction: %v", err)
	}
	if err := port.ApplyImportTx(ctx, tx, prepared, importContext); err != nil {
		_ = tx.Rollback(ctx)
		t.Fatalf("apply records import: %v", err)
	}
	if err := port.ValidateImportTx(ctx, tx, prepared, importContext); err != nil {
		_ = tx.Rollback(ctx)
		t.Fatalf("validate records import: %v", err)
	}
	var insideCount int
	if err := tx.QueryRow(ctx, `SELECT count(*) FROM records WHERE record_id = $1`, recordID).Scan(&insideCount); err != nil {
		_ = tx.Rollback(ctx)
		t.Fatalf("query applied record: %v", err)
	}
	if insideCount != 1 {
		_ = tx.Rollback(ctx)
		t.Fatalf("applied record count = %d; want 1", insideCount)
	}
	if err := tx.Rollback(ctx); err != nil {
		t.Fatalf("roll back records import: %v", err)
	}
	var outsideCount int
	if err := db.QueryRow(ctx, `SELECT count(*) FROM records WHERE record_id = $1`, recordID).Scan(&outsideCount); err != nil {
		t.Fatalf("query rolled-back record: %v", err)
	}
	if outsideCount != 0 {
		t.Fatalf("rolled-back record count = %d; want 0", outsideCount)
	}
}
