package revisions_test

import (
	"context"
	"net/http"
	"slices"
	"strings"
	"testing"

	"github.com/google/uuid"

	envelopetest "github.com/JochiRaider/cartulary/internal/modules/records/testsupport/envelopetest"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/authcookietest"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
)

func requireEvidenceRestoreBlobBridge(t *testing.T, harness *appsupport.ServerHarness, login authcookietest.AuthCookies, incident, actor uuid.UUID) {
	t.Helper()
	cases := []struct {
		lifecycle string
		allowed   []string
	}{
		{"requested", []string{"none", "available", "quarantined"}},
		{"pending_receipt", []string{"none", "available", "quarantined"}},
		{"received", []string{"none", "available", "quarantined"}},
		{"available", []string{"available"}},
		{"released", []string{"available"}},
		{"quarantined", []string{"none", "quarantined"}},
	}
	for _, test := range cases {
		for _, blobState := range []string{"none", "available", "quarantined", "pending", "failed"} {
			t.Run("Evidence restore Blob/"+test.lifecycle+"/"+blobState, func(t *testing.T) {
				evidence, blob := uuid.New(), uuid.New()
				copiedState := "available"
				if test.lifecycle == "quarantined" {
					copiedState = "quarantined"
				}
				digest := strings.Repeat("a", 64)
				mustExec(t, harness.DB, `INSERT INTO object_blobs(object_blob_id,incident_id,created_by_user_id,storage_key,upload_state,byte_size,observed_size,observed_content_type,observed_sha256_hex,target_expires_at,pending_expires_at,finalized_at) VALUES($1,$2,$3,$4,$5,0,0,'text/plain',$6,now()+interval '1 hour',now()+interval '24 hours',now())`, blob, incident, actor, "restore-blob/"+blob.String(), copiedState, digest)
				envelopetest.SeedRecordEnvelope(t, harness.DB, incident, actor, evidence, "evidence")
				mustExec(t, harness.DB, `INSERT INTO evidence(record_id,incident_id,title,lifecycle_state,object_blob_id,upload_state,blob_hash) VALUES($1,$2,'Blob restore fixture',$3,$4,$5,$6)`, evidence, incident, test.lifecycle, blob, copiedState, digest)
				httptestx.RequireSuccessEnvelope(t, deleteRecord(t, harness, login, evidence, map[string]any{"base_row_version": 1, "client_txn_id": "blob-bridge-delete"}), http.StatusOK)
				body := map[string]any{"base_row_version": 2, "client_txn_id": "blob-bridge-restore"}
				assertBlocked := func() {
					t.Helper()
					if countRows(t, harness.DB, `SELECT count(*) FROM records WHERE record_id=$1 AND row_version=2 AND deleted_at IS NOT NULL AND deleted_by_user_id=$2`, evidence, actor) != 1 || countRecordMutations(t, harness.DB, evidence, "restore") != 0 || countRecordRevisions(t, harness.DB, evidence) != 1 || countRows(t, harness.DB, `SELECT count(*) FROM evidence_grid_projection WHERE record_id=$1`, evidence) != 0 {
						t.Fatal("blocked Blob restore changed envelope, history or projection")
					}
				}
				if test.lifecycle == "available" && blobState == "available" {
					lock, err := harness.Pool.Begin(t.Context())
					if err != nil {
						t.Fatal(err)
					}
					defer lock.Rollback(context.Background())
					if _, err := lock.Exec(t.Context(), `SELECT object_blob_id FROM object_blobs WHERE object_blob_id=$1 FOR UPDATE`, blob); err != nil {
						t.Fatal(err)
					}
					locked := httptestx.RequireErrorEnvelope(t, restoreRecord(t, harness, login, evidence, body), http.StatusConflict, "record_locked")["error"].(map[string]any)
					details := locked["details"].(map[string]any)
					if locked["retryable"] != true || len(details) != 1 || details["record_id"] != evidence.String() {
						t.Fatalf("unsafe Blob contention details=%#v", locked)
					}
					if err := lock.Rollback(t.Context()); err != nil {
						t.Fatal(err)
					}
					assertBlocked()
				}
				var associated *uuid.UUID
				if blobState == "none" {
					mustExec(t, harness.DB, `UPDATE evidence SET object_blob_id=NULL WHERE record_id=$1`, evidence)
				} else {
					associated = &blob
					mustExec(t, harness.DB, `UPDATE object_blobs SET upload_state=$2, terminal_reason=CASE WHEN $2='failed' THEN 'pending_timeout' END, failed_at=CASE WHEN $2='failed' THEN now() END WHERE object_blob_id=$1`, blob, blobState)
				}
				assertSource := func() {
					t.Helper()
					if countRows(t, harness.DB, `SELECT count(*) FROM evidence WHERE record_id=$1 AND lifecycle_state=$2 AND upload_state=$3 AND blob_hash=$4 AND object_blob_id IS NOT DISTINCT FROM $5::uuid AND title='Blob restore fixture'`, evidence, test.lifecycle, copiedState, digest, associated) != 1 {
						t.Fatal("restore changed retained Evidence source or copied Blob scalars")
					}
				}
				if !slices.Contains(test.allowed, blobState) {
					blocked := httptestx.RequireErrorEnvelope(t, restoreRecord(t, harness, login, evidence, body), http.StatusConflict, "record_restore_blocked")["error"].(map[string]any)
					details := blocked["details"].(map[string]any)
					if blocked["retryable"] != false || len(details) != 2 || details["record_id"] != evidence.String() || details["reason_code"] != "evidence_blob_state_conflict" {
						t.Fatalf("unsafe Blob state blocker=%#v", blocked)
					}
					assertBlocked()
					assertSource()
					associated = &blob
					mustExec(t, harness.DB, `UPDATE evidence SET object_blob_id=$2 WHERE record_id=$1`, evidence, blob)
					mustExec(t, harness.DB, `UPDATE object_blobs SET upload_state=$2,terminal_reason=NULL,failed_at=NULL WHERE object_blob_id=$1`, blob, copiedState)
				}
				restored := httptestx.RequireSuccessEnvelope(t, restoreRecord(t, harness, login, evidence, body), http.StatusOK)["data"].(map[string]any)
				if restored["row_version"] != float64(3) || restored["deleted"] != false || restored["deleted_at"] != nil || restored["deleted_by_user_id"] != nil || countRecordMutations(t, harness.DB, evidence, "restore") != 1 || countRecordRevisions(t, harness.DB, evidence) != 2 || countRows(t, harness.DB, `SELECT count(*) FROM evidence_grid_projection WHERE record_id=$1`, evidence) != 1 {
					t.Fatalf("restore after Blob admission=%#v", restored)
				}
				assertSource()
				requireSingleRecordChangeIntent(t, harness.DB, restored["change_set_id"].(string), evidence, "invalidate")
			})
		}
	}
}

func requireEvidenceRestorePartyReferences(t *testing.T, harness *appsupport.ServerHarness, login authcookietest.AuthCookies, incident, actor uuid.UUID, variant string) {
	t.Helper()
	seedParty := func() uuid.UUID {
		id := uuid.New()
		envelopetest.SeedRecordEnvelope(t, harness.DB, incident, actor, id, "party")
		mustExec(t, harness.DB, `INSERT INTO parties(record_id,incident_id,display_name,party_kind) VALUES($1,$2,'Restore Party','organization')`, id, incident)
		return id
	}
	var collector, origin *uuid.UUID
	if variant == "collector" || variant == "both_same" || variant == "both_distinct" {
		id := seedParty()
		collector = &id
	}
	if variant == "source" || variant == "both_distinct" {
		id := seedParty()
		origin = &id
	}
	if variant == "both_same" {
		origin = collector
	}
	evidence := uuid.New()
	envelopetest.SeedRecordEnvelope(t, harness.DB, incident, actor, evidence, "evidence")
	mustExec(t, harness.DB, `INSERT INTO evidence(record_id,incident_id,title,collector_party_id,source_party_id) VALUES($1,$2,'Party restore fixture',$3,$4)`, evidence, incident, collector, origin)
	deleted := httptestx.RequireSuccessEnvelope(t, deleteRecord(t, harness, login, evidence, map[string]any{"base_row_version": 1, "client_txn_id": "evidence-party-delete-" + variant}), http.StatusOK)["data"].(map[string]any)
	if deleted["row_version"] != float64(2) || deleted["deleted"] != true {
		t.Fatalf("Evidence delete=%#v", deleted)
	}
	body := map[string]any{"base_row_version": 2, "client_txn_id": "evidence-party-restore-" + variant}
	assertUnchanged := func() {
		t.Helper()
		if countRows(t, harness.DB, `SELECT count(*) FROM records WHERE record_id=$1 AND row_version=2 AND deleted_at IS NOT NULL AND deleted_by_user_id=$2`, evidence, actor) != 1 || countRecordMutations(t, harness.DB, evidence, "restore") != 0 || countRecordRevisions(t, harness.DB, evidence) != 1 || countRows(t, harness.DB, `SELECT count(*) FROM evidence_grid_projection WHERE record_id=$1`, evidence) != 0 {
			t.Fatal("rejected restore changed Evidence tombstone, version, history or projection")
		}
		if countRows(t, harness.DB, `SELECT count(*) FROM evidence WHERE record_id=$1 AND collector_party_id IS NOT DISTINCT FROM $2::uuid AND source_party_id IS NOT DISTINCT FROM $3::uuid AND title='Party restore fixture' AND lifecycle_state='requested'`, evidence, collector, origin) != 1 {
			t.Fatal("rejected restore changed Evidence source")
		}
	}
	party := collector
	if origin != nil {
		party = origin
	}
	if party != nil {
		lock, err := harness.Pool.Begin(t.Context())
		if err != nil {
			t.Fatal(err)
		}
		defer lock.Rollback(context.Background())
		if _, err := lock.Exec(t.Context(), `SELECT record_id FROM records WHERE record_id=$1 FOR UPDATE`, *party); err != nil {
			t.Fatal(err)
		}
		locked := httptestx.RequireErrorEnvelope(t, restoreRecord(t, harness, login, evidence, body), http.StatusConflict, "record_locked")["error"].(map[string]any)
		details := locked["details"].(map[string]any)
		if locked["retryable"] != true || len(details) != 1 || details["record_id"] != evidence.String() {
			t.Fatalf("unsafe owner lock response=%#v", locked)
		}
		if err := lock.Rollback(t.Context()); err != nil {
			t.Fatal(err)
		}
		assertUnchanged()
		httptestx.RequireSuccessEnvelope(t, deleteRecord(t, harness, login, *party, map[string]any{"base_row_version": 1, "client_txn_id": "party-delete-before-evidence-restore-" + variant}), http.StatusOK)
		blocked := httptestx.RequireErrorEnvelope(t, restoreRecord(t, harness, login, evidence, body), http.StatusConflict, "record_restore_blocked")["error"].(map[string]any)
		details = blocked["details"].(map[string]any)
		if blocked["retryable"] != false || len(details) != 2 || details["record_id"] != evidence.String() || details["reason_code"] != "inactive_evidence_party_reference" {
			t.Fatalf("unsafe Party reference blocker=%#v", blocked)
		}
		assertUnchanged()
		httptestx.RequireSuccessEnvelope(t, restoreRecord(t, harness, login, *party, map[string]any{"base_row_version": 2, "client_txn_id": "party-restore-before-evidence-" + variant}), http.StatusOK)
	}
	restored := httptestx.RequireSuccessEnvelope(t, restoreRecord(t, harness, login, evidence, body), http.StatusOK)["data"].(map[string]any)
	if restored["row_version"] != float64(3) || restored["deleted"] != false || restored["deleted_at"] != nil || restored["deleted_by_user_id"] != nil || countRecordMutations(t, harness.DB, evidence, "restore") != 1 || countRecordRevisions(t, harness.DB, evidence) != 2 || countRows(t, harness.DB, `SELECT count(*) FROM evidence_grid_projection WHERE record_id=$1`, evidence) != 1 {
		t.Fatalf("Evidence restore after reference repair=%#v", restored)
	}
	requireSingleRecordChangeIntent(t, harness.DB, restored["change_set_id"].(string), evidence, "invalidate")
	if party != nil {
		blocked := httptestx.RequireErrorEnvelope(t, deleteRecord(t, harness, login, *party, map[string]any{"base_row_version": 3, "client_txn_id": "party-delete-after-evidence-restore-" + variant}), http.StatusConflict, "record_delete_blocked")["error"].(map[string]any)
		if blocked["details"].(map[string]any)["reason_code"] != "active_incoming_party_reference" {
			t.Fatalf("restored Evidence did not block Party delete=%#v", blocked)
		}
	}
}
