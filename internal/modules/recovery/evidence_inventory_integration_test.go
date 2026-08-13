package recovery_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	"github.com/JochiRaider/cartulary/internal/modules/evidence"
	"github.com/JochiRaider/cartulary/internal/modules/evidence/recoveryprovider"
	"github.com/JochiRaider/cartulary/internal/modules/incidents/testsupport/scenariotest"
	"github.com/JochiRaider/cartulary/internal/modules/recovery"
	"github.com/JochiRaider/cartulary/internal/platform/recoverystate"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
)

func TestRecoveryEvidenceInventoryClosure_Integration(t *testing.T) {
	runtime := appsupport.StartRuntime(t)
	harness := runtime.StartDefaultServer(t, "recovery-evidence-inventory-closure")
	admin, actorID := flowtest.ProvisionBootstrapAdmin(t, harness.Server.HTTP.URL)
	incident := scenariotest.CreateIncident(t, harness.Server, admin, map[string]any{
		"client_txn_id": "txn-recovery-evidence-inventory-incident",
		"incident_key":  "RECOVERY-EVIDENCE-INVENTORY",
		"title":         "Recovery Evidence inventory closure",
	})
	incidentID := uuid.MustParse(incident["incident_id"].(string))
	ctx := context.Background()
	fixedAt := time.Date(2026, time.July, 7, 13, 0, 0, 0, time.UTC)
	availableAttachedID := uuid.MustParse("00000000-0000-0000-0000-000000710001")
	availableUnattachedID := uuid.MustParse("00000000-0000-0000-0000-000000710002")
	pendingID := uuid.MustParse("00000000-0000-0000-0000-000000710003")
	failedID := uuid.MustParse("00000000-0000-0000-0000-000000710004")
	evidenceRecordID := uuid.MustParse("00000000-0000-0000-0000-000000710005")
	digestA := strings.Repeat("a", 64)
	digestB := strings.Repeat("b", 64)

	if _, err := harness.Pool.Exec(ctx, `
INSERT INTO object_blobs (
    object_blob_id, incident_id, created_by_user_id, storage_key, upload_state,
    byte_size, expected_sha256_hex, observed_size, observed_content_type,
    observed_sha256_hex, target_expires_at, pending_expires_at, finalized_at,
    terminal_reason, failed_at, created_at, updated_at
) VALUES
    ($1, $5, $6, 'evidence/available-attached', 'available',
     11, $7, 11, 'text/plain', $7, $8, $8, $9, NULL, NULL, $9, $9),
    ($2, $5, $6, 'evidence/available-unattached', 'available',
     12, $10, 12, 'application/pdf', $10, $8, $8, $9, NULL, NULL, $9, $9),
    ($3, $5, $6, 'evidence/pending-private', 'pending',
     13, NULL, NULL, NULL, NULL, $8, $8, NULL, NULL, NULL, $9, $9),
    ($4, $5, $6, 'evidence/failed-private', 'failed',
     14, NULL, NULL, NULL, NULL, $8, $8, NULL, 'declared_size_mismatch', $9, $9, $9)
`, availableAttachedID, availableUnattachedID, pendingID, failedID, incidentID, actorID,
		digestA, fixedAt.Add(time.Hour), fixedAt, digestB); err != nil {
		t.Fatalf("seed Evidence recovery object states: %v", err)
	}
	if _, err := harness.Pool.Exec(ctx, `
WITH record AS (
    INSERT INTO records (
        record_id, incident_id, record_type, created_by_user_id, updated_by_user_id,
        created_at, updated_at
    ) VALUES ($1, $2, 'evidence', $3, $3, $4, $4)
)
INSERT INTO evidence (
    record_id, incident_id, title, lifecycle_state, received_at,
    blob_hash, upload_state, object_blob_id, created_at, updated_at
) VALUES ($1, $2, 'Attached recovery Evidence', 'available', $4, $5, 'complete', $6, $4, $4)
`, evidenceRecordID, incidentID, actorID, fixedAt, digestA, availableAttachedID); err != nil {
		t.Fatalf("seed attached Evidence recovery row: %v", err)
	}

	contribution := evidence.RecoveryStateContribution()
	wantTables := []struct {
		name      string
		class     recoverystate.StateClass
		inclusion recoverystate.BackupInclusion
		action    recoverystate.RestoreAction
	}{
		{name: "evidence", class: recoverystate.StateAuthoritative, inclusion: recoverystate.InclusionRequired, action: recoverystate.RestoreState},
		{name: "evidence_custody_events", class: recoverystate.StateAuthoritative, inclusion: recoverystate.InclusionRequired, action: recoverystate.RestoreState},
		{name: "object_blobs", class: recoverystate.StateAuthoritative, inclusion: recoverystate.InclusionRequired, action: recoverystate.RestoreState},
		{name: "evidence_access_handles", class: recoverystate.StateSecurity, inclusion: recoverystate.InclusionSecurity, action: recoverystate.InvalidateState},
		{name: "evidence_object_upload_leases", class: recoverystate.StateSecurity, inclusion: recoverystate.InclusionSecurity, action: recoverystate.InvalidateState},
		{name: "evidence_blob_cleanup_claims", class: recoverystate.StateTransient, inclusion: recoverystate.InclusionTransient, action: recoverystate.InvalidateState},
	}
	if len(contribution.Tables) != len(wantTables) {
		t.Fatalf("Evidence recovery relation count = %d, want %d: %#v", len(contribution.Tables), len(wantTables), contribution.Tables)
	}
	for index, want := range wantTables {
		got := contribution.Tables[index]
		if got.TableName != want.name || got.StateClass != want.class || got.BackupInclusion != want.inclusion || got.RestoreAction != want.action {
			t.Fatalf("Evidence recovery relation %d = %#v, want %#v", index, got, want)
		}
	}
	if len(contribution.ObjectFamilies) != 1 {
		t.Fatalf("Evidence recovery object-family count = %d, want 1", len(contribution.ObjectFamilies))
	}
	family := contribution.ObjectFamilies[0]
	if family.ObjectFamilyID != "evidence.blobs" || family.StateClass != recoverystate.StateAuthoritative ||
		family.BackupInclusion != recoverystate.InclusionRequired || family.RestoreAction != recoverystate.RestoreState {
		t.Fatalf("unexpected Evidence recovery object-family posture: %#v", family)
	}

	provider := recoveryprovider.New(harness.Pool)
	objects, err := provider.ListAvailableRecoveryObjects(ctx)
	if err != nil {
		t.Fatalf("list Evidence recovery objects: %v", err)
	}
	if len(objects) != 2 {
		t.Fatalf("available Evidence recovery object count = %d, want 2: %#v", len(objects), objects)
	}
	if objects[0].ObjectBlobID != availableAttachedID || objects[0].StorageKey != "evidence/available-attached" ||
		objects[1].ObjectBlobID != availableUnattachedID || objects[1].StorageKey != "evidence/available-unattached" {
		t.Fatalf("available Evidence recovery objects are incomplete or unordered: %#v", objects)
	}
	if objects[0].BlobHash == nil || *objects[0].BlobHash != digestA || objects[1].BlobHash != nil {
		t.Fatalf("Evidence recovery attachment metadata = %#v, want attached hash then nil", objects)
	}
	index, err := recovery.AvailableBlobObjectIDsByStorageRef(ctx, provider)
	if err != nil {
		t.Fatalf("build Evidence recovery object index: %v", err)
	}
	wantIndex := map[string]uuid.UUID{
		"evidence/available-attached":   availableAttachedID,
		"evidence/available-unattached": availableUnattachedID,
	}
	if len(index) != len(wantIndex) {
		t.Fatalf("Evidence recovery object index = %#v, want %#v", index, wantIndex)
	}
	for key, wantID := range wantIndex {
		if index[key] != wantID {
			t.Fatalf("Evidence recovery object index[%q] = %s, want %s", key, index[key], wantID)
		}
	}
	rowCount, err := provider.CountRecoveryRows(ctx)
	if err != nil {
		t.Fatalf("count Evidence recovery rows: %v", err)
	}
	if rowCount != 4 {
		t.Fatalf("Evidence recovery row count = %d, want 4", rowCount)
	}

	tx, err := harness.Pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly})
	if err != nil {
		t.Fatalf("begin Evidence vNext inventory snapshot: %v", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	vNext := evidence.VNextRecoveryObjectInventory(evidenceInventoryObjectSource{})
	if vNext.OwnerID() != "module.evidence" || vNext.ObjectFamilyID() != "evidence.blobs" ||
		vNext.InventoryAlgorithmID() != "evidence.snapshot_blob_inventory.v1" {
		t.Fatalf("unexpected Evidence vNext inventory identity: owner=%q family=%q algorithm=%q", vNext.OwnerID(), vNext.ObjectFamilyID(), vNext.InventoryAlgorithmID())
	}
	members, err := vNext.Inventory(ctx, evidenceInventorySnapshot{tx: tx})
	if err != nil {
		t.Fatalf("inventory Evidence vNext objects: %v", err)
	}
	if len(members) != 2 || members[0].LogicalObjectID != availableAttachedID.String() || members[1].LogicalObjectID != availableUnattachedID.String() {
		t.Fatalf("Evidence vNext inventory members = %#v, want the two available blobs exactly once", members)
	}
	if members[0].StorageKey != "evidence/available-attached" || members[0].ContentType != "text/plain" || members[0].PlaintextBytes != 11 || members[0].PlaintextSHA256 != digestA {
		t.Fatalf("unexpected attached Evidence vNext member: %#v", members[0])
	}
	if members[1].StorageKey != "evidence/available-unattached" || members[1].ContentType != "application/pdf" || members[1].PlaintextBytes != 12 || members[1].PlaintextSHA256 != digestB {
		t.Fatalf("unexpected unattached Evidence vNext member: %#v", members[1])
	}
}

type evidenceInventorySnapshot struct {
	tx pgx.Tx
}

func (snapshot evidenceInventorySnapshot) StreamCanonicalTableRows(
	context.Context,
	string,
	func(json.RawMessage) error,
) error {
	return fmt.Errorf("table streaming is outside the Evidence inventory fixture")
}

func (snapshot evidenceInventorySnapshot) QueryRows(ctx context.Context, query string, args ...any) (recovery.VNextRows, error) {
	return snapshot.tx.Query(ctx, query, args...)
}

type evidenceInventoryObjectSource struct{}

func (evidenceInventoryObjectSource) OpenRecoveryObject(context.Context, string) (io.ReadCloser, error) {
	return nil, fmt.Errorf("object reads are outside the Evidence inventory fixture")
}

func (evidenceInventoryObjectSource) StatRecoveryObject(context.Context, string) (recovery.VNextObjectSourceInfo, error) {
	return recovery.VNextObjectSourceInfo{}, fmt.Errorf("object stats are outside the Evidence inventory fixture")
}
