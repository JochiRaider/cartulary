package evidence_test

import (
	"context"
	"crypto/sha256"
	"fmt"
	"maps"
	"net/http"
	"reflect"
	"testing"

	"github.com/google/uuid"

	authflowtest "github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	incidentstoretest "github.com/JochiRaider/cartulary/internal/modules/incidents/testsupport/storetest"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/collaborationsupport/incidentwstest"
	"github.com/JochiRaider/cartulary/internal/testutil/contractassert"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/JochiRaider/cartulary/internal/testutil/workbookscenariotest"
)

func TestEvidenceSurfaceContracts_Integration(t *testing.T) {
	for _, kind := range []string{"blob", "attach", "preview", "download"} {
		t.Run(kind, func(t *testing.T) {
			harness := appsupport.StartRuntime(t).StartServer(t, appsupport.ServerOptions{Prefix: "evidence-contract-" + kind, TestRouteMode: httptestx.TestRouteModeDisabled})
			admin, adminID := appsupport.ProvisionBootstrapAdmin(t, harness.Server)
			incident := appsupport.CreateIncident(t, harness.Server, admin, map[string]any{"client_txn_id": "incident", "incident_key": "EVIDENCE-" + kind, "title": "Evidence " + kind})
			incidentID := appsupport.MustUUID(t, incident["incident_id"].(string))
			actor := authflowtest.SeedLocalUserRecord(t, harness.DB, "evidence-contract@example.test", "Evidence actor", "EvidenceContracts1!", false, false, true)
			incidentstoretest.SeedMembership(t, harness.DB, incidentID, actor.ID, actor.DisplayName, "admin", adminID)
			login := loginLocalUserNoMFA(t, harness, actor.Email, "EvidenceContracts1!")
			if kind == "blob" {
				body := map[string]any{"incident_id": incidentID.String(), "client_txn_id": "blob", "byte_size": 14, "filename_hint": " support-object.txt ", "content_type_hint": "text/plain", "sha256_hex": "9af4c73b2a919f220f4b008e466b52808a1987122d95ff0f2dde00968e36e844"}
				url := harness.Server.HTTP.URL + "/api/v1/object-blobs"
				appsupport.RequireErrorBody(t, appsupport.DoJSON(t, http.MethodPost, url, body, appsupport.WithCookies(login.SessionCookie, login.CSRFCookie)), http.StatusForbidden, "csrf_verification_failed")
				first := appsupport.RequireSuccessData(t, appsupport.DoJSON(t, http.MethodPost, url, body, authOptions(login)...), http.StatusCreated)
				if id, ok := first["object_blob_id"].(string); !ok || id == "" {
					t.Fatalf("missing blob ID: %#v", first)
				}
				if first["incident_id"] != incidentID.String() || first["upload_state"] != "pending" {
					t.Fatalf("wrong blob state: %#v", first)
				}
				for _, key := range []string{"upload_target", "accepted_contract"} {
					if _, ok := first[key].(map[string]any); !ok {
						t.Fatalf("missing %s: %#v", key, first)
					}
				}
				before := snapshotEvidenceContractState(t, harness, incidentID)
				replay := appsupport.RequireSuccessData(t, appsupport.DoJSON(t, http.MethodPost, url, body, authOptions(login)...), http.StatusOK)
				for _, key := range []string{"incident_id", "object_blob_id", "upload_state", "target_expires_at", "pending_expires_at", "upload_target", "accepted_contract"} {
					if !reflect.DeepEqual(replay[key], first[key]) {
						t.Fatalf("blob replay changed %s", key)
					}
				}
				if got := snapshotEvidenceContractState(t, harness, incidentID); got != before {
					t.Fatalf("blob replay changed source: %+v -> %+v", before, got)
				}
				divergent := maps.Clone(body)
				divergent["byte_size"] = 1
				appsupport.RequireErrorBody(t, appsupport.DoJSON(t, http.MethodPost, url, divergent, authOptions(login)...), http.StatusConflict, "client_txn_conflict")
				if got := snapshotEvidenceContractState(t, harness, incidentID); got != before {
					t.Fatalf("divergent blob changed source: %+v -> %+v", before, got)
				}
				demoteEvidenceContractActor(t, harness, incidentID, actor.ID)
				denied := maps.Clone(body)
				denied["client_txn_id"] = "denied"
				appsupport.RequireErrorBody(t, appsupport.DoJSON(t, http.MethodPost, url, denied, authOptions(login)...), http.StatusForbidden, "authorization_denied")
				if got := snapshotEvidenceContractState(t, harness, incidentID); got != before {
					t.Fatalf("denied blob changed source: %+v -> %+v", before, got)
				}
				return
			}
			const view = "cartulary.view.evidence.v1"
			created := requireHTTPWorkbookCreate(t, harness, login, incidentID, view, map[string]any{"client_txn_id": "evidence", "evidence.title": "Evidence contract"})
			id := appsupport.MustUUID(t, created["row"].(map[string]any)["record_id"].(string))
			if kind == "preview" || kind == "download" {
				attachUploadedBlobWithMetadata(t, harness, login, incidentID, id, []byte("Evidence handle"), "handle.txt", "text/plain", "blob", "attach")
				url := harness.Server.HTTP.URL + "/api/v1/evidence-records/" + id.String() + "/" + kind + "-handle"
				appsupport.RequireErrorBody(t, appsupport.DoJSON(t, http.MethodPost, url, map[string]any{}, appsupport.WithCookies(login.SessionCookie, login.CSRFCookie)), http.StatusForbidden, "csrf_verification_failed")
				data := appsupport.RequireSuccessData(t, appsupport.DoJSON(t, http.MethodPost, url, map[string]any{}, authOptions(login)...), http.StatusOK)
				if href, ok := data["href"].(string); !ok || href == "" {
					t.Fatalf("missing handle href: %#v", data)
				}
				if data["record_id"] != id.String() || data["handle_kind"] != kind {
					t.Fatalf("wrong handle identity: %#v", data)
				}
				before := snapshotEvidenceContractState(t, harness, incidentID)
				if _, err := harness.DB.ExecContext(context.Background(), `DELETE FROM incident_memberships WHERE incident_id=$1 AND user_id=$2`, incidentID, actor.ID); err != nil {
					t.Fatal(err)
				}
				appsupport.RequireErrorBody(t, appsupport.DoJSON(t, http.MethodPost, url, map[string]any{}, authOptions(login)...), http.StatusNotFound, "evidence_record_not_found")
				if got := snapshotEvidenceContractState(t, harness, incidentID); got != before {
					t.Fatalf("denied handle changed evidence: %+v -> %+v", before, got)
				}
				return
			}
			blob := uploadEvidenceContractBlob(t, harness, login, incidentID, "primary")
			alternate := uploadEvidenceContractBlob(t, harness, login, incidentID, "alternate")
			url := harness.Server.HTTP.URL + "/api/v1/evidence-records/" + id.String() + "/attach-blob"
			body := map[string]any{"object_blob_id": blob, "base_row_version": 1, "client_txn_id": "attach"}
			appsupport.RequireErrorBody(t, appsupport.DoJSON(t, http.MethodPost, url, body, appsupport.WithCookies(login.SessionCookie, login.CSRFCookie)), http.StatusForbidden, "csrf_verification_failed")
			socket := incidentwstest.ConnectViewSocket(t, harness.Server, incidentID.String(), view, login.SessionCookie.Value)
			defer socket.Close(1000, "test_complete")
			first := appsupport.RequireSuccessData(t, appsupport.DoJSON(t, http.MethodPost, url, body, authOptions(login)...), http.StatusOK)
			row := first["row"].(map[string]any)
			if first["object_blob_id"] != blob || row["record_id"] != id.String() || first["change_set_id"] == nil || first["change_set_id"] == "" {
				t.Fatalf("wrong attachment identity: %#v", first)
			}
			change := incidentwstest.RequireRecordChanged(t, socket, id.String(), 2)
			if change.ChangeSetID != first["change_set_id"] {
				t.Fatalf("wrong socket change identity: %#v", change)
			}
			found := false
			for _, affected := range change.AffectedViews {
				if affected.ViewSchemaID == view {
					found = true
				}
			}
			if !found {
				t.Fatalf("missing evidence affected view: %#v", change)
			}
			incidentwstest.ExpectNoSocketMessage(t, socket)
			envelope := workbookscenariotest.QueryViewEnvelope(t, harness.Server.HTTP.URL, incidentID.String(), view, login)
			contractassert.RequireDefaultQueryMeta(t, envelope, view)
			projected := requireEvidenceProjectionRow(t, harness, login, incidentID, id)
			contractassert.RequireFieldKeyConformance(t, workbookscenariotest.SortedRowFieldKeys(t, projected), workbookscenariotest.AllowedFieldKeys(t, "attach", view))
			before := snapshotEvidenceContractState(t, harness, incidentID)
			history := workbookscenariotest.SnapshotIncidentMutationEffects(t, harness.DB, incidentID.String())
			replay := appsupport.RequireSuccessData(t, appsupport.DoJSON(t, http.MethodPost, url, body, authOptions(login)...), http.StatusOK)
			if replay["change_set_id"] != first["change_set_id"] || replay["object_blob_id"] != blob || !reflect.DeepEqual(row, replay["row"]) {
				t.Fatalf("attachment replay changed result: %#v %#v", first, replay)
			}
			if got := snapshotEvidenceContractState(t, harness, incidentID); got != before {
				t.Fatalf("attachment replay changed source: %+v -> %+v", before, got)
			}
			if got := workbookscenariotest.SnapshotIncidentMutationEffects(t, harness.DB, incidentID.String()); got != history {
				t.Fatal("exact attachment replay changed effects")
			}
			contractassert.RequireProjectionDeterminism(t, projected, requireEvidenceProjectionRow(t, harness, login, incidentID, id))
			incidentwstest.ExpectNoSocketMessage(t, socket)
			divergent := maps.Clone(body)
			divergent["object_blob_id"] = alternate
			appsupport.RequireErrorBody(t, appsupport.DoJSON(t, http.MethodPost, url, divergent, authOptions(login)...), http.StatusConflict, "client_txn_conflict")
			if got := snapshotEvidenceContractState(t, harness, incidentID); got != before {
				t.Fatalf("divergent attachment changed source: %+v -> %+v", before, got)
			}
			if got := workbookscenariotest.SnapshotIncidentMutationEffects(t, harness.DB, incidentID.String()); got != history {
				t.Fatalf("attachment changed history: %+v -> %+v", history, got)
			}
			contractassert.RequireProjectionDeterminism(t, projected, requireEvidenceProjectionRow(t, harness, login, incidentID, id))
			incidentwstest.ExpectNoSocketMessage(t, socket)
		})
	}
}

func uploadEvidenceContractBlob(t *testing.T, harness *appsupport.ServerHarness, login appsupport.LoginResult, incidentID uuid.UUID, label string) string {
	t.Helper()
	payload := []byte("Evidence " + label)
	sum := sha256.Sum256(payload)
	data := appsupport.RequireSuccessData(t, appsupport.DoJSON(t, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/object-blobs", map[string]any{"incident_id": incidentID.String(), "client_txn_id": "blob-" + label, "byte_size": len(payload), "filename_hint": label + ".txt", "content_type_hint": "text/plain", "sha256_hex": fmt.Sprintf("%x", sum)}, authOptions(login)...), http.StatusCreated)
	putObject(t, harness.Server.HTTP.URL, data["upload_target"].(map[string]any)["href"].(string), payload, "text/plain", login)
	verifyUploadedBlob(t, harness, login, data["object_blob_id"], "verify-"+label, "succeeded")
	return data["object_blob_id"].(string)
}

func demoteEvidenceContractActor(t testing.TB, harness *appsupport.ServerHarness, incidentID, actorID uuid.UUID) {
	t.Helper()
	if _, err := harness.DB.ExecContext(context.Background(), `UPDATE incident_memberships SET role='viewer',membership_version=membership_version+1 WHERE incident_id=$1 AND user_id=$2`, incidentID, actorID); err != nil {
		t.Fatal(err)
	}
}

type evidenceContractState struct {
	Source  string
	Effects workbookscenariotest.MutationEffects
}

func snapshotEvidenceContractState(t testing.TB, harness *appsupport.ServerHarness, incidentID uuid.UUID) evidenceContractState {
	t.Helper()
	var state string
	if err := harness.DB.QueryRowContext(context.Background(), `SELECT jsonb_build_object(
 'blobs',(SELECT coalesce(jsonb_agg(to_jsonb(b) ORDER BY object_blob_id),'[]') FROM object_blobs b WHERE incident_id=$1),
 'evidence',(SELECT coalesce(jsonb_agg(to_jsonb(e) ORDER BY record_id),'[]') FROM evidence e WHERE incident_id=$1),
 'records',(SELECT coalesce(jsonb_agg(to_jsonb(r) ORDER BY record_id),'[]') FROM records r WHERE incident_id=$1)
 )::text`, incidentID).Scan(&state); err != nil {
		t.Fatal(err)
	}
	return evidenceContractState{Source: state, Effects: workbookscenariotest.SnapshotIncidentMutationEffects(t, harness.DB, incidentID.String())}
}
