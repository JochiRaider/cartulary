package recovery_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	"github.com/JochiRaider/cartulary/internal/modules/evidence"
	"github.com/JochiRaider/cartulary/internal/modules/incidents/testsupport/scenariotest"
	"github.com/JochiRaider/cartulary/internal/modules/recovery"
	"github.com/JochiRaider/cartulary/internal/platform/recoverystate"
	"github.com/JochiRaider/cartulary/internal/platform/sqlshape"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
)

func TestRecoveryEvidenceInventoryClosure_Integration(t *testing.T) {
	runtime := appsupport.StartRuntime(t)
	harness := runtime.StartServer(t, appsupport.ServerOptions{Prefix: "recovery-evidence-inventory-closure", TestRouteMode: httptestx.TestRouteModeDisabled})
	admin, actorID := flowtest.ProvisionBootstrapAdmin(t, http.DefaultClient, harness.Server.HTTP.URL)
	incident := scenariotest.CreateIncident(t, harness.Server, admin, map[string]any{
		"client_txn_id": "txn-recovery-evidence-inventory-incident",
		"incident_key":  "RECOVERY-EVIDENCE-INVENTORY",
		"title":         "Recovery Evidence inventory closure",
	})
	incidentID := uuid.MustParse(incident["incident_id"].(string))
	ctx := context.Background()
	fixedAt := time.Now().UTC().Truncate(time.Microsecond)
	availableAttachedID := uuid.MustParse("00000000-0000-0000-0000-000000710001")
	availableUnattachedID := uuid.MustParse("00000000-0000-0000-0000-000000710002")
	pendingID := uuid.MustParse("00000000-0000-0000-0000-000000710003")
	failedID := uuid.MustParse("00000000-0000-0000-0000-000000710004")
	evidenceRecordID := uuid.MustParse("00000000-0000-0000-0000-000000710005")
	digestA := strings.Repeat("a", 64)
	digestB := strings.Repeat("b", 64)

	quarantinedAttachedID := uuid.MustParse("00000000-0000-0000-0000-000000710006")
	quarantinedUnattachedID := uuid.MustParse("00000000-0000-0000-0000-000000710007")
	deletedEvidenceID := uuid.MustParse("00000000-0000-0000-0000-000000710008")
	digestC, digestD := strings.Repeat("c", 64), strings.Repeat("d", 64)
	key := func(id uuid.UUID) string {
		return fmt.Sprintf("incidents/%s/object-blobs/%s/objects/00000000-0000-4000-8000-000000710009", incidentID, id)
	}
	for _, item := range []struct {
		id       uuid.UUID
		state    string
		size     int64
		digest   any
		expected any
	}{
		{availableAttachedID, "available", 11, digestA, digestA},
		{availableUnattachedID, "available", 12, digestB, digestB},
		{pendingID, "pending", 13, nil, nil}, {failedID, "failed", 14, nil, nil},
		{quarantinedAttachedID, "quarantined", 15, digestC, digestC},
		{quarantinedUnattachedID, "quarantined", 16, digestD, nil},
	} {
		var observed, finalized, reason, failed, cleanup, contentType any
		if item.state == "available" || item.state == "quarantined" {
			observed = item.size
			finalized = fixedAt
			contentType = "text/plain"
		}
		if item.state == "failed" {
			reason = "declared_size_mismatch"
			failed = fixedAt
			cleanup = fixedAt.Add(45 * time.Minute)
		}
		if _, err := harness.Pool.Exec(ctx, `INSERT INTO object_blobs(object_blob_id,incident_id,created_by_user_id,storage_key,upload_state,byte_size,expected_sha256_hex,observed_size,observed_content_type,observed_sha256_hex,target_expires_at,pending_expires_at,finalized_at,terminal_reason,failed_at,cleanup_due_at,created_at,updated_at) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$17,$9,$10,$11,$12,$13,$14,$15,$16,$16)`, item.id, incidentID, actorID, key(item.id), item.state, item.size, item.expected, observed, item.digest, fixedAt.Add(time.Hour), fixedAt.Add(24*time.Hour), finalized, reason, failed, cleanup, fixedAt, contentType); err != nil {
			t.Fatal(err)
		}
	}
	for _, item := range []struct {
		record, blob  uuid.UUID
		state, digest string
		deleted       bool
	}{{evidenceRecordID, availableAttachedID, "available", digestA, false}, {deletedEvidenceID, quarantinedAttachedID, "quarantined", digestC, true}} {
		if _, err := harness.Pool.Exec(ctx, `WITH record AS (INSERT INTO records(record_id,incident_id,record_type,created_by_user_id,updated_by_user_id,created_at,updated_at) VALUES($1,$2,'evidence',$3,$3,$4,$4)) INSERT INTO evidence(record_id,incident_id,title,lifecycle_state,received_at,blob_hash,upload_state,object_blob_id,created_at,updated_at) VALUES($1,$2,'Retained recovery Evidence',$5,$4,$6,$5,$7,$4,$4)`, item.record, incidentID, actorID, fixedAt, item.state, item.digest, item.blob); err != nil {
			t.Fatal(err)
		}
		if item.deleted {
			if _, err := harness.Pool.Exec(ctx, `UPDATE records SET deleted_at=$2,deleted_by_user_id=$3 WHERE record_id=$1`, item.record, fixedAt, actorID); err != nil {
				t.Fatal(err)
			}
		}
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
		{name: "evidence_blob_verifications", class: recoverystate.StateSecurity, inclusion: recoverystate.InclusionSecurity, action: recoverystate.InvalidateState},
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
		family.BackupInclusion != recoverystate.InclusionRequired || family.RestoreAction != recoverystate.RestoreState || family.InventoryAlgorithmID == nil || *family.InventoryAlgorithmID != "evidence.snapshot_blob_inventory.v2" {
		t.Fatalf("unexpected Evidence recovery object-family posture: %#v", family)
	}

	provider := evidence.NewRecoveryProvider(harness.Pool)
	objects, err := provider.ListRequiredRecoveryObjects(ctx)
	if err != nil {
		t.Fatalf("list Evidence recovery objects: %v", err)
	}
	ptr := func(value string) *string { return &value }
	sizePtr := func(value int64) *int64 { return &value }
	wantObjects := []recovery.EvidenceObjectState{
		{ObjectBlobID: availableAttachedID, StorageKey: key(availableAttachedID), ByteSize: 11, ObservedSize: sizePtr(11), ExpectedSHA256Hex: ptr(digestA), ObservedSHA256Hex: ptr(digestA), BlobHash: ptr(digestA)},
		{ObjectBlobID: availableUnattachedID, StorageKey: key(availableUnattachedID), ByteSize: 12, ObservedSize: sizePtr(12), ExpectedSHA256Hex: ptr(digestB), ObservedSHA256Hex: ptr(digestB)},
		{ObjectBlobID: quarantinedAttachedID, StorageKey: key(quarantinedAttachedID), ByteSize: 15, ObservedSize: sizePtr(15), ExpectedSHA256Hex: ptr(digestC), ObservedSHA256Hex: ptr(digestC), BlobHash: ptr(digestC)},
		{ObjectBlobID: quarantinedUnattachedID, StorageKey: key(quarantinedUnattachedID), ByteSize: 16, ObservedSize: sizePtr(16), ObservedSHA256Hex: ptr(digestD)},
	}
	if !reflect.DeepEqual(objects, wantObjects) {
		t.Fatalf("required Evidence objects=%#v want=%#v", objects, wantObjects)
	}

	index, err := recovery.RequiredBlobObjectIDsByStorageRef(ctx, provider)
	if err != nil {
		t.Fatalf("build Evidence recovery object index: %v", err)
	}
	wantIndex := map[string]uuid.UUID{
		key(availableAttachedID):     availableAttachedID,
		key(availableUnattachedID):   availableUnattachedID,
		key(quarantinedAttachedID):   quarantinedAttachedID,
		key(quarantinedUnattachedID): quarantinedUnattachedID,
	}
	if len(index) != len(wantIndex) {
		t.Fatalf("Evidence recovery object index = %#v, want %#v", index, wantIndex)
	}
	for key, wantID := range wantIndex {
		if index[key] != wantID {
			t.Fatalf("Evidence recovery object index[%q] = %s, want %s", key, index[key], wantID)
		}
	}

	tx, err := harness.Pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly})
	if err != nil {
		t.Fatalf("begin Evidence vNext inventory snapshot: %v", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	objectSource := &evidenceInventoryObjectSource{}
	vNext := evidence.VNextRecoveryObjectInventory(objectSource)
	if vNext.OwnerID() != "module.evidence" || vNext.ObjectFamilyID() != "evidence.blobs" ||
		vNext.InventoryAlgorithmID() != "evidence.snapshot_blob_inventory.v2" {
		t.Fatalf("unexpected Evidence vNext inventory identity: owner=%q family=%q algorithm=%q", vNext.OwnerID(), vNext.ObjectFamilyID(), vNext.InventoryAlgorithmID())
	}
	members, err := vNext.Inventory(ctx, evidenceInventorySnapshot{tx: tx})
	if err != nil {
		t.Fatalf("inventory Evidence vNext objects: %v", err)
	}
	assertMembers := func(t testing.TB, members []recovery.VNextObjectMember) {
		t.Helper()
		if len(members) != 4 {
			t.Fatalf("required inventory count=%d want4", len(members))
		}
		for i, want := range wantObjects {
			got := members[i]
			if got.LogicalObjectID != want.ObjectBlobID.String() || got.StorageKey != want.StorageKey || got.ContentType != "text/plain" || got.PlaintextBytes != want.ByteSize || got.PlaintextSHA256 != *want.ObservedSHA256Hex || got.Open == nil {
				t.Fatalf("inventory[%d]=%#v want=%#v", i, got, want)
			}
		}
	}
	assertMembers(t, members)
	// Freeze the read snapshot, then corrupt a finalized row outside it. Both
	// inventories reject fresh reads, while the original snapshot stays coherent.
	for _, state := range []string{"available", "quarantined"} {
		for _, change := range []struct {
			name, field string
			value       any
		}{
			{"missing observed size", "observed_size", nil}, {"mismatched observed size", "observed_size", int64(17)},
			{"missing digest", "observed_sha256_hex", nil},
			{"expected mismatch", "expected_sha256_hex", digestA},
		} {
			t.Run(state+"/"+change.name, func(t *testing.T) {
				if _, err := harness.Pool.Exec(ctx, "UPDATE object_blobs SET upload_state=$2,"+change.field+"=$3 WHERE object_blob_id=$1", quarantinedUnattachedID, state, change.value); err != nil {
					t.Fatal(err)
				}
				defer func() {
					if _, err := harness.Pool.Exec(ctx, `UPDATE object_blobs SET upload_state='quarantined',observed_size=16,observed_sha256_hex=$2,expected_sha256_hex=NULL WHERE object_blob_id=$1`, quarantinedUnattachedID, digestD); err != nil {
						t.Error(err)
					}
				}()
				actual, err := provider.ListRequiredRecoveryObjects(ctx)
				if !errors.Is(err, recovery.ErrInvalidBackupArtifact) || actual != nil {
					t.Fatalf("invalid required row produced partial inventory=%#v err=%v", actual, err)
				}
				fresh, err := harness.Pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly})
				if err != nil {
					t.Fatal(err)
				}
				defer fresh.Rollback(ctx)
				actualMembers, err := vNext.Inventory(ctx, evidenceInventorySnapshot{tx: fresh})
				if !errors.Is(err, recovery.ErrInvalidBackupArtifact) || actualMembers != nil {
					t.Fatalf("invalid required row produced snapshot members=%#v err=%v", actualMembers, err)
				}
				original, err := vNext.Inventory(ctx, evidenceInventorySnapshot{tx: tx})
				if err != nil {
					t.Fatal(err)
				}
				assertMembers(t, original)
			})
		}
	}
	if objectSource.reads != 0 || objectSource.stats != 0 {
		t.Fatalf("inventory accessed bytes: reads=%d stats=%d", objectSource.reads, objectSource.stats)
	}
	var deleted bool
	var state string
	if err := harness.Pool.QueryRow(ctx, `SELECT r.deleted_at IS NOT NULL,b.upload_state FROM records r JOIN evidence e USING(record_id) JOIN object_blobs b ON b.object_blob_id=e.object_blob_id WHERE r.record_id=$1`, deletedEvidenceID).Scan(&deleted, &state); err != nil || !deleted || state != "quarantined" {
		t.Fatalf("inventory changed retained state: deleted=%v blob=%s err=%v", deleted, state, err)
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

func (snapshot evidenceInventorySnapshot) QueryRows(ctx context.Context, query sqlshape.Fragment, args ...any) (recovery.VNextRows, error) {
	return snapshot.tx.Query(ctx, sqlshape.Text(query), args...)
}

type evidenceInventoryObjectSource struct{ reads, stats int }

func (source *evidenceInventoryObjectSource) OpenRecoveryObject(context.Context, string) (io.ReadCloser, error) {
	source.reads++
	return nil, fmt.Errorf("object reads are outside the Evidence inventory fixture")
}

func (source *evidenceInventoryObjectSource) StatRecoveryObject(context.Context, string) (recovery.VNextObjectSourceInfo, error) {
	source.stats++
	return recovery.VNextObjectSourceInfo{}, fmt.Errorf("object stats are outside the Evidence inventory fixture")
}
