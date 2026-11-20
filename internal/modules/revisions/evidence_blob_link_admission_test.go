package revisions_test

import (
	"net/http"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	"github.com/JochiRaider/cartulary/internal/modules/evidence/testsupport/blobfixture"
	envelopetest "github.com/JochiRaider/cartulary/internal/modules/records/testsupport/envelopetest"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
)

// TestEvidenceBlobLinkAdmission_Integration covers AC-609 (REQ-02-189a,
// REQ-02-190, REQ-02-196a) for the destructive-history Evidence paths: a
// rollback that would re-link a pending, failed, foreign, associated or
// cleanup-claimed Blob fails with 409 target_not_reversible before any write,
// and a restore or rollback that loads a Blob violating its lifecycle tuple, or
// a restore of an Evidence lifecycle outside the vocabulary, fails with
// internal_error. Every refusal commits nothing.
func TestEvidenceBlobLinkAdmission_Integration(t *testing.T) {
	harness := appsupport.StartServer(t, appsupport.ServerOptions{Prefix: "evidence-blob-link-admission", TestRouteMode: httptestx.TestRouteModeDisabled})
	login, actor := flowtest.ProvisionBootstrapAdmin(t, http.DefaultClient, harness.Server.HTTP.URL)
	incident, _ := seedRecord(t, harness.DB, harness.Server, login, actor, "IR-BLOB-LINK-ADMISSION")
	foreign := uuid.MustParse(appsupport.CreateIncident(t, harness.Server, login, map[string]any{
		"client_txn_id": "blob-link-admission-foreign", "incident_key": "IR-BLOB-LINK-ADMISSION-FOREIGN", "title": "Foreign Blob owner",
	})["incident_id"].(string))
	blobSnapshot := func(t *testing.T, blob uuid.UUID) string {
		t.Helper()
		var snapshot string
		if err := harness.DB.QueryRowContext(t.Context(), `
SELECT jsonb_build_object(
           'blob', (SELECT to_jsonb(b) FROM object_blobs b WHERE b.object_blob_id = $1),
           'claims', (SELECT COALESCE(jsonb_agg(to_jsonb(c)), '[]'::jsonb) FROM evidence_blob_cleanup_claims c WHERE c.object_blob_id = $1),
           'links', (SELECT COALESCE(jsonb_agg(e.record_id ORDER BY e.record_id), '[]'::jsonb) FROM evidence e WHERE e.object_blob_id = $1)
       )::text
`, blob).Scan(&snapshot); err != nil {
			t.Fatalf("snapshot Blob: %v", err)
		}
		return snapshot
	}

	t.Run("rollback re-link", func(t *testing.T) {
		cases := []struct {
			name, state  string
			foreign      bool
			associate    bool
			claim        bool
			breach       bool
			admitted     bool
			internalFail bool
		}{
			{name: "pending", state: blobfixture.Pending},
			{name: "failed", state: blobfixture.Failed},
			{name: "foreign-incident", state: blobfixture.Available, foreign: true},
			{name: "associated", state: blobfixture.Available, associate: true},
			{name: "cleanup-claimed", state: blobfixture.Available, claim: true},
			{name: "cleanup-claimed-failed", state: blobfixture.Failed, claim: true},
			{name: "tuple-violation", state: blobfixture.Available, breach: true, internalFail: true},
			{name: "admitted", state: blobfixture.Available, admitted: true},
		}
		for _, test := range cases {
			t.Run(test.name, func(t *testing.T) {
				evidence, blob, changeSet := uuid.New(), uuid.New(), uuid.New()
				blobIncident := incident
				if test.foreign {
					blobIncident = foreign
				}
				digest := strings.Repeat("c", 64)
				blobfixture.Seed(t, harness.Pool, blobfixture.Spec{
					ObjectBlobID: blob, IncidentID: blobIncident, CreatedByUserID: actor, State: test.state,
					StorageKey: "rollback-link-blob/" + blob.String(), ByteSize: 3, ExpectedSHA256Hex: &digest,
				})
				var other uuid.UUID
				if test.associate {
					other = uuid.New()
					envelopetest.SeedRecordEnvelope(t, harness.DB, incident, actor, other, "evidence")
					mustExec(t, harness.DB, `INSERT INTO evidence(record_id,incident_id,title,lifecycle_state,object_blob_id,upload_state,storage_ref,blob_hash) VALUES($1,$2,'Current owner','received',$3,'available',$4,$5)`, other, incident, blob, "object://"+blob.String(), digest)
				}
				if test.claim {
					blobfixture.ClaimForCleanup(t, harness.Pool, blob)
				}
				if test.breach {
					blobfixture.BreachTupleInvariant(t, harness.Pool, blob)
				}
				envelopetest.SeedRecordEnvelope(t, harness.DB, incident, actor, evidence, "evidence")
				mustExec(t, harness.DB, `INSERT INTO evidence(record_id,incident_id,title,lifecycle_state,upload_state) VALUES($1,$2,'Link after','requested','pending')`, evidence, incident)
				advanceRecordFixture(t, harness.DB, evidence, 2)
				// The retained lifecycle (received) satisfies the bridge with any
				// Blob state, so only the association rule decides the re-link.
				before := canonicalRowSnapshot(evidence, incident, actor, "evidence", "cartulary.revisions.snapshot.evidence.v2", 1, map[string]any{
					"title": "Link before", "lifecycle_state": "received", "upload_state": "available",
					"object_blob_id": blob.String(), "storage_ref": "object://" + blob.String(), "blob_hash": digest,
				})
				after := canonicalRowSnapshot(evidence, incident, actor, "evidence", "cartulary.revisions.snapshot.evidence.v2", 2, map[string]any{
					"title": "Link after", "lifecycle_state": "requested", "upload_state": "pending",
				})
				seedRollbackMutationWithRef(t, harness.DB, incident, actor, evidence, changeSet, 1, "record", evidence.String(), "patch", before, after, "href-rollback-link-"+test.name)
				beforeBlob := blobSnapshot(t, blob)
				changeSets := countRows(t, harness.DB, `SELECT count(*) FROM change_sets WHERE incident_id=$1`, incident)
				body := map[string]any{"base_row_version": 2, "client_txn_id": "rollback-link-" + test.name, "target": map[string]any{"kind": "change_set", "change_set_id": changeSet.String()}}
				response := rollbackRecord(t, harness, login, evidence, body)
				if test.admitted {
					httptestx.RequireSuccessEnvelope(t, response, http.StatusOK)
					if countRows(t, harness.DB, `SELECT count(*) FROM evidence WHERE record_id=$1 AND title='Link before' AND lifecycle_state='received' AND object_blob_id=$2`, evidence, blob) != 1 {
						t.Fatal("admitted Evidence rollback did not re-link the finalized Blob")
					}
					return
				}
				if test.internalFail {
					httptestx.RequireErrorEnvelope(t, response, http.StatusInternalServerError, "internal_error")
				} else {
					requireRollbackReasonCode(t, response, "target_not_reversible")
				}
				if countRows(t, harness.DB, `SELECT count(*) FROM evidence e JOIN records r USING(record_id) WHERE e.record_id=$1 AND e.title='Link after' AND e.lifecycle_state='requested' AND e.object_blob_id IS NULL AND r.row_version=2`, evidence) != 1 ||
					countRecordMutations(t, harness.DB, evidence, "rollback") != 0 ||
					countRows(t, harness.DB, `SELECT count(*) FROM change_sets WHERE incident_id=$1`, incident) != changeSets {
					t.Fatal("refused Evidence rollback changed the source, envelope or history")
				}
				if afterBlob := blobSnapshot(t, blob); afterBlob != beforeBlob {
					t.Fatalf("refused Evidence rollback changed the Blob, its claim or its association:\nbefore %s\nafter  %s", beforeBlob, afterBlob)
				}
			})
		}
	})

	t.Run("restore", func(t *testing.T) {
		cases := []struct {
			name      string
			linked    bool
			lifecycle string
			breach    bool
		}{
			{name: "tuple-violation", linked: true, lifecycle: "available", breach: true},
			{name: "lifecycle-outside-vocabulary", lifecycle: "archived"},
		}
		for _, test := range cases {
			t.Run(test.name, func(t *testing.T) {
				evidence := uuid.New()
				envelopetest.SeedRecordEnvelope(t, harness.DB, incident, actor, evidence, "evidence")
				var blob *uuid.UUID
				if test.linked {
					digest := strings.Repeat("d", 64)
					id := blobfixture.Seed(t, harness.Pool, blobfixture.Spec{
						IncidentID: incident, CreatedByUserID: actor, State: blobfixture.Available, ByteSize: 3, ExpectedSHA256Hex: &digest,
					})
					blob = &id
					mustExec(t, harness.DB, `INSERT INTO evidence(record_id,incident_id,title,lifecycle_state,object_blob_id,upload_state,storage_ref,blob_hash) VALUES($1,$2,'Restore invariant','available',$3,'available',$4,$5)`, evidence, incident, id, "object://"+id.String(), digest)
				} else {
					mustExec(t, harness.DB, `INSERT INTO evidence(record_id,incident_id,title,lifecycle_state,upload_state) VALUES($1,$2,'Restore invariant','requested','pending')`, evidence, incident)
				}
				httptestx.RequireSuccessEnvelope(t, deleteRecord(t, harness, login, evidence, map[string]any{"base_row_version": 1, "client_txn_id": "restore-invariant-delete-" + test.name}), http.StatusOK)
				if test.breach {
					blobfixture.BreachTupleInvariant(t, harness.Pool, *blob)
				}
				if test.lifecycle != "available" {
					// Invariant-breach fixture: the Evidence lifecycle has no
					// database vocabulary constraint, only the REQ-02-196a rule.
					mustExec(t, harness.DB, `UPDATE evidence SET lifecycle_state=$2 WHERE record_id=$1`, evidence, test.lifecycle)
				}
				var beforeBlob string
				if blob != nil {
					beforeBlob = blobSnapshot(t, *blob)
				}
				body := map[string]any{"base_row_version": 2, "client_txn_id": "restore-invariant-" + test.name}
				httptestx.RequireErrorEnvelope(t, restoreRecord(t, harness, login, evidence, body), http.StatusInternalServerError, "internal_error")
				if countRows(t, harness.DB, `SELECT count(*) FROM records WHERE record_id=$1 AND row_version=2 AND deleted_at IS NOT NULL AND deleted_by_user_id=$2`, evidence, actor) != 1 ||
					countRecordMutations(t, harness.DB, evidence, "restore") != 0 || countRecordRevisions(t, harness.DB, evidence) != 1 ||
					countRows(t, harness.DB, `SELECT count(*) FROM evidence_grid_projection WHERE record_id=$1`, evidence) != 0 ||
					countRows(t, harness.DB, `SELECT count(*) FROM evidence WHERE record_id=$1 AND lifecycle_state=$2 AND title='Restore invariant'`, evidence, test.lifecycle) != 1 {
					t.Fatal("failed Evidence restore changed the envelope, source, history or projection")
				}
				if blob != nil {
					if afterBlob := blobSnapshot(t, *blob); afterBlob != beforeBlob {
						t.Fatalf("failed Evidence restore changed the Blob:\nbefore %s\nafter  %s", beforeBlob, afterBlob)
					}
				}
			})
		}
	})
}
