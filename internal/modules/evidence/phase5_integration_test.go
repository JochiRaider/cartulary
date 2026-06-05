package evidence_test

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/evidence"
	"github.com/JochiRaider/cartulary/internal/modules/evidence/blobref"
	"github.com/JochiRaider/cartulary/internal/modules/projections"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/objectstore"
	platformws "github.com/JochiRaider/cartulary/internal/platform/ws"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/JochiRaider/cartulary/internal/testutil/incidentwstest"
	"github.com/JochiRaider/cartulary/internal/testutil/phase4test"
)

func TestPhase5_ObjectUploadAttachWorkbookProjection_I_5_01(t *testing.T) {
	harness := phase4test.StartServer(t, "phase5-upload-attach-projection")
	login, _ := phase4test.ProvisionBootstrapAdmin(t, harness.Server)
	incident := phase4test.CreateIncident(t, harness.Server, login, map[string]any{
		"client_txn_id": "txn-phase5-i-01-incident",
		"incident_key":  "phase5-i-01",
		"title":         "Phase 5 upload attach projection",
	})
	incidentID := phase4test.MustUUID(t, incident["incident_id"].(string))

	timelineData := requirePhase5HTTPWorkbookCreate(t, harness, login, incidentID, "cartulary.view.timeline.v1", map[string]any{
		"client_txn_id":    "txn-phase5-i-01-timeline",
		"timeline.summary": "Endpoint screenshot received",
	})
	timelineRow := timelineData["row"].(map[string]any)
	timelineRecordID := phase4test.MustUUID(t, timelineRow["record_id"].(string))
	timelineRowVersion := int(timelineRow["row_version"].(float64))
	requireTimelineEvidenceProjection(t, harness, login, incidentID, timelineRecordID, 0, false)
	evidenceData := requirePhase5HTTPWorkbookCreate(t, harness, login, incidentID, "cartulary.view.evidence.v1", map[string]any{
		"client_txn_id":                 "txn-phase5-i-01-evidence",
		"evidence.title":                "Endpoint screenshot",
		"evidence.collector_party_text": "IR collector",
	})
	evidenceRecordID := phase4test.MustUUID(t, evidenceData["row"].(map[string]any)["record_id"].(string))

	payload := []byte("phase5 projection object")
	sum := sha256.Sum256(payload)
	createResp := phase4test.DoJSON(t, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/object-blobs", map[string]any{
		"incident_id":       incidentID.String(),
		"client_txn_id":     "txn-phase5-i-01-blob",
		"byte_size":         len(payload),
		"filename_hint":     " endpoint.txt ",
		"content_type_hint": "text/plain",
		"sha256_hex":        fmt.Sprintf("%x", sum[:]),
	}, authOptions(login)...)
	createData := httptestx.RequireSuccessEnvelope(t, createResp, http.StatusCreated)["data"].(map[string]any)
	putObject(t, createData["upload_target"].(map[string]any)["href"].(string), payload, "text/plain")

	attachBody := map[string]any{
		"object_blob_id":   createData["object_blob_id"],
		"base_row_version": 1,
		"client_txn_id":    "txn-phase5-i-01-attach",
	}
	attachResp := phase4test.DoJSON(t, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/evidence-records/"+evidenceRecordID.String()+"/attach-blob", attachBody, authOptions(login)...)
	attachData := httptestx.RequireSuccessEnvelope(t, attachResp, http.StatusOK)["data"].(map[string]any)
	if attachData["object_blob_id"] != createData["object_blob_id"] {
		t.Fatalf("attach object_blob_id got %#v want %#v", attachData["object_blob_id"], createData["object_blob_id"])
	}
	requireTimelineEvidenceProjection(t, harness, login, incidentID, timelineRecordID, 0, false)

	requirePhase5HTTPWorkbookPatch(t, harness, login, timelineRecordID, map[string]any{
		"view_schema_id":   "cartulary.view.timeline.v1",
		"base_row_version": timelineRowVersion,
		"client_txn_id":    "txn-phase5-i-01-link-evidence",
		"changes": []map[string]any{{
			"field_key": "timeline.attached_evidence_ids",
			"action_payload": map[string]any{
				"kind": "collection_actions_v1",
				"actions": []map[string]any{{
					"op":               "add_record_ref",
					"linked_record_id": evidenceRecordID.String(),
				}},
			},
		}},
	})
	requireTimelineEvidenceProjection(t, harness, login, incidentID, timelineRecordID, 1, true)
	if got := countAttachedEvidenceLinks(t, harness, incidentID, timelineRecordID, evidenceRecordID); got != 1 {
		t.Fatalf("workbook patch wrote attached evidence links: got %d want 1", got)
	}

	replayResp := phase4test.DoJSON(t, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/evidence-records/"+evidenceRecordID.String()+"/attach-blob", attachBody, authOptions(login)...)
	replayData := httptestx.RequireSuccessEnvelope(t, replayResp, http.StatusOK)["data"].(map[string]any)
	if replayData["change_set_id"] != attachData["change_set_id"] {
		t.Fatalf("attach replay changed change_set_id: replay=%#v first=%#v", replayData["change_set_id"], attachData["change_set_id"])
	}
	requireTimelineEvidenceProjection(t, harness, login, incidentID, timelineRecordID, 1, true)
	if got := countEvidenceRevisions(t, harness, evidenceRecordID); got != 2 {
		t.Fatalf("attach replay created duplicate record revision: got %d want 2", got)
	}
	if got := countEvidenceBlobLinks(t, harness, evidenceRecordID); got != 1 {
		t.Fatalf("evidence row has duplicate or missing blob link: got %d want 1", got)
	}
}

func TestPhase5_AttachRouteContract_I_5_05(t *testing.T) {
	runtime := phase4test.StartRuntime(t)
	testDB := runtime.PrepareServerDatabase(t, "phase5-attach-route-contract")
	harness := runtime.StartServerWithDatabase(t, "phase5-attach-route-contract", testDB)
	login, adminID := phase4test.ProvisionBootstrapAdmin(t, harness.Server)
	objectStoreAdmin := phase4test.SeedLocalUserFlags(t, harness.DB, "phase5-object-store-admin@example.test", "Phase5 Object Store Admin", "Phase5ObjectStoreAdmin1!", false, true, true)
	incident := phase4test.CreateIncident(t, harness.Server, login, map[string]any{
		"client_txn_id": "txn-phase5-attach-route-incident",
		"incident_key":  "phase5-attach-route",
		"title":         "Phase 5 attach route contract",
	})
	incidentID := phase4test.MustUUID(t, incident["incident_id"].(string))
	otherIncident := phase4test.CreateIncident(t, harness.Server, login, map[string]any{
		"client_txn_id": "txn-phase5-attach-route-other",
		"incident_key":  "phase5-attach-route-other",
		"title":         "Phase 5 attach route other",
	})
	otherIncidentID := phase4test.MustUUID(t, otherIncident["incident_id"].(string))

	recordID := uuid.New()
	seedEvidenceRecord(t, harness, incidentID, adminID, recordID)
	availableBlobID := insertPhase5RouteBlob(t, harness, incidentID, adminID, "available")
	attachURL := harness.Server.HTTP.URL + "/api/v1/evidence-records/" + recordID.String() + "/attach-blob"

	for _, tc := range []struct {
		name string
		body string
	}{
		{name: "array", body: `[]`},
		{name: "missing object_blob_id", body: `{"base_row_version":1,"client_txn_id":"txn-route-shape"}`},
		{name: "missing base_row_version", body: `{"object_blob_id":"` + availableBlobID.String() + `","client_txn_id":"txn-route-shape"}`},
		{name: "missing client_txn_id", body: `{"object_blob_id":"` + availableBlobID.String() + `","base_row_version":1}`},
		{name: "null client_txn_id", body: `{"object_blob_id":"` + availableBlobID.String() + `","base_row_version":1,"client_txn_id":null}`},
		{name: "unknown member", body: `{"object_blob_id":"` + availableBlobID.String() + `","base_row_version":1,"client_txn_id":"txn-route-shape","extra":true}`},
	} {
		t.Run("shape "+tc.name, func(t *testing.T) {
			httptestx.RequireErrorEnvelope(t, doRawJSON(t, http.MethodPost, attachURL, tc.body, authOptions(login)...), http.StatusBadRequest, "invalid_mutation_payload")
		})
	}

	t.Run("viewer cannot attach", func(t *testing.T) {
		viewer := phase4test.SeedLocalUserFlags(t, harness.DB, "phase5-attach-viewer@example.test", "Phase5 Attach Viewer", "Phase5AttachViewer1!", false, false, true)
		phase4test.SeedIncidentMembership(t, harness.DB, incidentID, viewer.ID, "Phase5 Attach Viewer", "viewer", adminID)
		viewerLogin := loginLocalUserNoMFA(t, harness, "phase5-attach-viewer@example.test", "Phase5AttachViewer1!")
		resp := phase4test.DoJSON(t, http.MethodPost, attachURL, map[string]any{
			"object_blob_id":   availableBlobID.String(),
			"base_row_version": 1,
			"client_txn_id":    "txn-phase5-viewer-denied",
		}, authOptions(viewerLogin)...)
		httptestx.RequireErrorEnvelope(t, resp, http.StatusForbidden, "authorization_denied")
		requireEvidenceStates(t, harness, recordID, "received", "pending")
	})

	t.Run("route-owned failures use canonical envelopes", func(t *testing.T) {
		cases := []struct {
			name       string
			blobID     uuid.UUID
			base       int
			wantStatus int
			wantCode   string
			wantReason string
		}{
			{name: "row version conflict", blobID: availableBlobID, base: 9, wantStatus: http.StatusConflict, wantCode: "row_version_conflict"},
			{name: "missing blob", blobID: uuid.New(), base: 1, wantStatus: http.StatusConflict, wantCode: "evidence_attach_rejected", wantReason: evidence.AttachReasonBlobNotVisible},
			{name: "foreign blob", blobID: insertPhase5RouteBlob(t, harness, otherIncidentID, adminID, "available"), base: 1, wantStatus: http.StatusConflict, wantCode: "evidence_attach_rejected", wantReason: evidence.AttachReasonBlobNotVisible},
			{name: "pending blob", blobID: insertPhase5RouteBlob(t, harness, incidentID, adminID, "pending"), base: 1, wantStatus: http.StatusConflict, wantCode: "evidence_attach_rejected", wantReason: evidence.AttachReasonBlobPending},
			{name: "failed blob", blobID: insertPhase5RouteBlob(t, harness, incidentID, adminID, "failed"), base: 1, wantStatus: http.StatusConflict, wantCode: "evidence_attach_rejected", wantReason: evidence.AttachReasonBlobFailed},
			{name: "quarantined blob", blobID: insertPhase5RouteBlob(t, harness, incidentID, adminID, "quarantined"), base: 1, wantStatus: http.StatusConflict, wantCode: "evidence_attach_rejected", wantReason: evidence.AttachReasonBlobQuarantined},
		}
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				resp := phase4test.DoJSON(t, http.MethodPost, attachURL, map[string]any{
					"object_blob_id":   tc.blobID.String(),
					"base_row_version": tc.base,
					"client_txn_id":    "txn-phase5-route-" + strings.ReplaceAll(tc.name, " ", "-"),
				}, authOptions(login)...)
				body := httptestx.RequireErrorEnvelope(t, resp, tc.wantStatus, tc.wantCode)
				if tc.wantReason != "" {
					details := body["error"].(map[string]any)["details"].(map[string]any)
					if got := details["reason_code"]; got != tc.wantReason {
						t.Fatalf("reason_code got %v want %s", got, tc.wantReason)
					}
				}
				requireEvidenceStates(t, harness, recordID, "received", "pending")
			})
		}
	})

	t.Run("divergent replay is rejected through HTTP", func(t *testing.T) {
		replayRecordID := uuid.New()
		seedEvidenceRecord(t, harness, incidentID, adminID, replayRecordID)
		firstBlobID := insertPhase5RouteBlob(t, harness, incidentID, adminID, "available")
		replayURL := harness.Server.HTTP.URL + "/api/v1/evidence-records/" + replayRecordID.String() + "/attach-blob"
		body := map[string]any{"object_blob_id": firstBlobID.String(), "base_row_version": 1, "client_txn_id": "txn-phase5-route-divergent"}
		first := httptestx.RequireSuccessEnvelope(t, phase4test.DoJSON(t, http.MethodPost, replayURL, body, authOptions(login)...), http.StatusOK)["data"].(map[string]any)
		beforeRevisions := countEvidenceRevisions(t, harness, replayRecordID)
		secondBlobID := insertPhase5RouteBlob(t, harness, incidentID, adminID, "available")
		resp := phase4test.DoJSON(t, http.MethodPost, replayURL, map[string]any{
			"object_blob_id":   secondBlobID.String(),
			"base_row_version": 2,
			"client_txn_id":    "txn-phase5-route-divergent",
		}, authOptions(login)...)
		httptestx.RequireErrorEnvelope(t, resp, http.StatusConflict, "client_txn_conflict")
		if got := countEvidenceRevisions(t, harness, replayRecordID); got != beforeRevisions {
			t.Fatalf("divergent replay changed revision count: got %d want %d", got, beforeRevisions)
		}
		if got := countEvidenceBlobLinks(t, harness, replayRecordID); got != 1 {
			t.Fatalf("divergent replay changed blob link count: got %d want 1", got)
		}
		if first["object_blob_id"] != firstBlobID.String() {
			t.Fatalf("first attach object_blob_id got %#v want %s", first["object_blob_id"], firstBlobID)
		}
	})

	t.Run("object-store dependency errors use owner public mapping", func(t *testing.T) {
		requirePhaseDObjectStoreDependencyErrorsUseOwnerPublicMapping(
			t,
			func(t testing.TB, prefix string, store objectstore.Store) *phase4test.ServerHarness {
				return runtime.StartServerWithDatabaseAndObjectStore(t, prefix, testDB, store)
			},
			phase5ObjectStoreDependencyAdmin{
				email:    objectStoreAdmin.Email,
				password: "Phase5ObjectStoreAdmin1!",
				userID:   objectStoreAdmin.ID,
			},
		)
	})
}

func TestPhase5_AttachedEvidenceProjectionRebuild_I_5_06(t *testing.T) {
	harness := phase4test.StartServer(t, "phase5-projection-rebuild")
	login, _ := phase4test.ProvisionBootstrapAdmin(t, harness.Server)
	incident := phase4test.CreateIncident(t, harness.Server, login, map[string]any{
		"client_txn_id": "txn-phase5-projection-incident",
		"incident_key":  "phase5-projection",
		"title":         "Phase 5 projection rebuild",
	})
	incidentID := phase4test.MustUUID(t, incident["incident_id"].(string))

	timelineData := requirePhase5HTTPWorkbookCreate(t, harness, login, incidentID, "cartulary.view.timeline.v1", map[string]any{
		"client_txn_id":    "txn-phase5-projection-timeline",
		"timeline.summary": "Projection rebuild row",
	})
	timelineRow := timelineData["row"].(map[string]any)
	timelineRecordID := phase4test.MustUUID(t, timelineRow["record_id"].(string))
	timelineRowVersion := int(timelineRow["row_version"].(float64))

	evidenceData := requirePhase5HTTPWorkbookCreate(t, harness, login, incidentID, "cartulary.view.evidence.v1", map[string]any{
		"client_txn_id":  "txn-phase5-projection-evidence",
		"evidence.title": "Projection evidence",
	})
	evidenceRecordID := phase4test.MustUUID(t, evidenceData["row"].(map[string]any)["record_id"].(string))
	attachUploadedBlobWithMetadata(t, harness, login, incidentID, evidenceRecordID, []byte("phase5 projection rebuild"), "projection.txt", "text/plain", "txn-phase5-projection-blob", "txn-phase5-projection-attach")

	requirePhase5HTTPWorkbookPatch(t, harness, login, timelineRecordID, map[string]any{
		"view_schema_id":   "cartulary.view.timeline.v1",
		"base_row_version": timelineRowVersion,
		"client_txn_id":    "txn-phase5-projection-link",
		"changes": []map[string]any{{
			"field_key": "timeline.attached_evidence_ids",
			"action_payload": map[string]any{
				"kind": "collection_actions_v1",
				"actions": []map[string]any{{
					"op":               "add_record_ref",
					"linked_record_id": evidenceRecordID.String(),
				}},
			},
		}},
	})
	requireTimelineEvidenceProjection(t, harness, login, incidentID, timelineRecordID, 1, true)

	sourceBefore := phase5ProcessCounts(t, harness, incidentID, timelineRecordID)
	if _, err := harness.DB.ExecContext(context.Background(), `
UPDATE timeline_grid_projection
   SET evidence_count = 0,
       has_evidence = false
 WHERE record_id = $1
`, timelineRecordID); err != nil {
		t.Fatalf("corrupt timeline projection: %v", err)
	}
	requireTimelineProjectionStorage(t, harness, timelineRecordID, 0, false)
	requireTimelineEvidenceProjection(t, harness, login, incidentID, timelineRecordID, 1, true)

	if err := projections.NewStore(harness.Server.Runtime.Postgres).RebuildIncidentTimeline(context.Background(), incidentID); err != nil {
		t.Fatalf("rebuild timeline projection: %v", err)
	}
	requireTimelineProjectionStorage(t, harness, timelineRecordID, 1, true)
	requireTimelineEvidenceProjection(t, harness, login, incidentID, timelineRecordID, 1, true)
	if sourceAfter := phase5ProcessCounts(t, harness, incidentID, timelineRecordID); sourceAfter != sourceBefore {
		t.Fatalf("projection rebuild mutated source/history state: before=%+v after=%+v", sourceBefore, sourceAfter)
	}
	if got := countEvidenceBlobLinks(t, harness, evidenceRecordID); got != 1 {
		t.Fatalf("projection rebuild changed evidence blob link count: got %d want 1", got)
	}
}

func TestPhase5_AttachPublishesWorkbookWebSocketRefresh_I_5_07(t *testing.T) {
	harness := phase4test.StartServer(t, "phase5-attach-websocket-refresh")
	login, adminID := phase4test.ProvisionBootstrapAdmin(t, harness.Server)
	incident := phase4test.CreateIncident(t, harness.Server, login, map[string]any{
		"client_txn_id": "txn-phase5-i-07-incident",
		"incident_key":  "phase5-i-07",
		"title":         "Phase 5 attach websocket refresh",
	})
	incidentID := phase4test.MustUUID(t, incident["incident_id"].(string))

	timelineData := requirePhase5HTTPWorkbookCreate(t, harness, login, incidentID, "cartulary.view.timeline.v1", map[string]any{
		"client_txn_id":    "txn-phase5-i-07-timeline",
		"timeline.summary": "WebSocket evidence count target",
	})
	timelineRow := timelineData["row"].(map[string]any)
	timelineRecordID := phase4test.MustUUID(t, timelineRow["record_id"].(string))
	timelineRowVersion := int64(timelineRow["row_version"].(float64))
	evidenceData := requirePhase5HTTPWorkbookCreate(t, harness, login, incidentID, "cartulary.view.evidence.v1", map[string]any{
		"client_txn_id":  "txn-phase5-i-07-evidence",
		"evidence.title": "WebSocket evidence",
	})
	evidenceRecordID := phase4test.MustUUID(t, evidenceData["row"].(map[string]any)["record_id"].(string))
	if _, err := harness.DB.ExecContext(context.Background(), `
INSERT INTO record_links (
    incident_id, src_record_id, dst_record_id, link_type, field_key,
    provenance, owner_user_id, created_by_user_id, decided_at, created_at
) VALUES ($1, $2, $3, 'attached_evidence', 'timeline.attached_evidence_ids', 'manual', $4, $4, now(), now())
`, incidentID, timelineRecordID, evidenceRecordID, adminID); err != nil {
		t.Fatalf("insert attached evidence link: %v", err)
	}

	socket := incidentwstest.ConnectAndHello(t, harness.Server.HTTP.URL, incidentID.String(), incidentwstest.ConnectOptions{
		SessionToken:     login.SessionCookie.Value,
		ClientInstanceID: "phase5-i-07-record-change-listener",
		Presence: platformws.PresenceInput{
			SheetRef: map[string]string{"kind": "view_schema", "id": "cartulary.view.timeline.v1"},
			Mode:     "viewing",
		},
	})
	defer socket.Close(websocket.StatusNormalClosure, "test_complete")

	attachData := attachUploadedBlobWithMetadata(t, harness, login, incidentID, evidenceRecordID, []byte("phase5 websocket attach"), "websocket.txt", "text/plain", "txn-phase5-i-07-blob", "txn-phase5-i-07-attach")
	rowVersion := int64(attachData["row"].(map[string]any)["row_version"].(float64))

	evidenceChange := phase5AwaitRecordChanged(t, socket, evidenceRecordID, rowVersion)
	phase5RequireAffectedView(t, evidenceChange, "cartulary.view.evidence.v1")
	if !containsString(phase5ChangedFieldKeys(t, evidenceChange), "evidence.upload_state") {
		t.Fatalf("evidence attach changed keys missing evidence.upload_state: %#v", evidenceChange)
	}
	timelineChange := phase5AwaitRecordChanged(t, socket, timelineRecordID, timelineRowVersion)
	phase5RequireAffectedView(t, timelineChange, "cartulary.view.timeline.v1")
	changedKeys := phase5ChangedFieldKeys(t, timelineChange)
	for _, key := range []string{"timeline.attached_evidence_ids", "timeline.evidence_count", "timeline.has_evidence"} {
		if !containsString(changedKeys, key) {
			t.Fatalf("timeline websocket changed keys missing %s: %#v", key, timelineChange)
		}
	}
}

func phase5AwaitRecordChanged(t testing.TB, client *incidentwstest.Client, recordID uuid.UUID, rowVersion int64) map[string]any {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		message, err := client.AwaitNextMessage(time.Until(deadline))
		if err != nil {
			t.Fatalf("wait for record_changed for %s: %v", recordID, err)
		}
		if message.Type != "record_changed" {
			continue
		}
		var payload map[string]any
		if err := json.Unmarshal(message.Payload, &payload); err != nil {
			t.Fatalf("decode record_changed payload: %v", err)
		}
		payloadRowVersion, ok := payload["row_version"].(float64)
		if !ok {
			t.Fatalf("record_changed payload missing numeric row_version: %#v", payload)
		}
		if payload["record_id"] == recordID.String() && int64(payloadRowVersion) == rowVersion {
			return payload
		}
	}
	t.Fatalf("timed out waiting for record_changed for %s version %d", recordID, rowVersion)
	return nil
}

func phase5RequireAffectedView(t testing.TB, payload map[string]any, viewSchemaID string) {
	t.Helper()
	affectedViews, ok := payload["affected_views"].([]any)
	if !ok {
		t.Fatalf("record_changed payload missing affected_views: %#v", payload)
	}
	for _, rawView := range affectedViews {
		view, ok := rawView.(map[string]any)
		if !ok {
			t.Fatalf("record_changed affected view has invalid shape: %#v", rawView)
		}
		if view["view_schema_id"] == viewSchemaID {
			if view["change_kind"] == "" {
				t.Fatalf("affected view missing change_kind: %#v", view)
			}
			return
		}
	}
	t.Fatalf("record_changed payload missing affected view %s: %#v", viewSchemaID, payload)
}

func phase5ChangedFieldKeys(t testing.TB, payload map[string]any) []string {
	t.Helper()
	rawKeys, ok := payload["changed_field_keys"].([]any)
	if !ok {
		t.Fatalf("record_changed payload missing changed_field_keys: %#v", payload)
	}
	keys := make([]string, 0, len(rawKeys))
	for _, rawKey := range rawKeys {
		key, ok := rawKey.(string)
		if !ok {
			t.Fatalf("record_changed changed field key is not a string: %#v", rawKey)
		}
		keys = append(keys, key)
	}
	return keys
}

func TestPhase5_ExpiredSlotReplay_I_5_02(t *testing.T) {
	harness := phase4test.StartServer(t, "phase5-expired-slot-replay")
	login, adminID := phase4test.ProvisionBootstrapAdmin(t, harness.Server)
	incident := phase4test.CreateIncident(t, harness.Server, login, map[string]any{
		"client_txn_id": "txn-phase5-expired-slot-incident",
		"incident_key":  "phase5-expired-slot",
		"title":         "Phase 5 expired slot replay",
	})
	incidentID := phase4test.MustUUID(t, incident["incident_id"].(string))

	issuedAt := time.Now().UTC().Add(time.Minute).Truncate(time.Second)
	httptestx.SetClockFixed(t, harness.Server, issuedAt)
	createBody := map[string]any{
		"incident_id":       incidentID.String(),
		"client_txn_id":     "txn-phase5-expired-slot",
		"byte_size":         17,
		"filename_hint":     " expired.txt ",
		"content_type_hint": "text/plain",
	}
	createResp := phase4test.DoJSON(t, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/object-blobs", createBody, authOptions(login)...)
	createData := httptestx.RequireSuccessEnvelope(t, createResp, http.StatusCreated)["data"].(map[string]any)
	requireCreateExpiry(t, createData, "target_expires_at", issuedAt.Add(60*time.Minute))
	requireCreateExpiry(t, createData, "pending_expires_at", issuedAt.Add(24*time.Hour))

	replayAt := issuedAt.Add(61 * time.Minute)
	extendSessionForClockJump(t, harness, adminID, replayAt.Add(30*time.Minute))
	httptestx.SetClockFixed(t, harness.Server, replayAt)
	replayResp := phase4test.DoJSON(t, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/object-blobs", createBody, authOptions(login)...)
	replayData := httptestx.RequireSuccessEnvelope(t, replayResp, http.StatusOK)["data"].(map[string]any)
	for _, key := range []string{"object_blob_id", "target_expires_at", "pending_expires_at"} {
		if replayData[key] != createData[key] {
			t.Fatalf("expired replay refreshed %s: replay=%#v create=%#v", key, replayData, createData)
		}
	}
	replayedTarget := mustParseTime(t, replayData["target_expires_at"].(string))
	if !replayedTarget.Before(replayAt) {
		t.Fatalf("replayed target should remain expired: got %s replay_at %s", replayedTarget, replayAt)
	}

	freshBody := map[string]any{
		"incident_id":       incidentID.String(),
		"client_txn_id":     "txn-phase5-expired-slot-fresh",
		"byte_size":         17,
		"filename_hint":     " expired.txt ",
		"content_type_hint": "text/plain",
	}
	freshResp := phase4test.DoJSON(t, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/object-blobs", freshBody, authOptions(login)...)
	freshData := httptestx.RequireSuccessEnvelope(t, freshResp, http.StatusCreated)["data"].(map[string]any)
	if freshData["object_blob_id"] == createData["object_blob_id"] {
		t.Fatalf("fresh client_txn_id reused expired object_blob_id: %#v", freshData)
	}
	requireCreateExpiry(t, freshData, "target_expires_at", replayAt.Add(60*time.Minute))
	requireCreateExpiry(t, freshData, "pending_expires_at", replayAt.Add(24*time.Hour))
	if got := countObjectBlobs(t, harness, incidentID); got != 2 {
		t.Fatalf("fresh target should create exactly one additional blob slot: got %d want 2", got)
	}
}

func TestPhase5_HandleRedeemInvalidatesOnCurrentStateLoss_I_5_03(t *testing.T) {
	harness := phase4test.StartServer(t, "phase5-handle-redeem-invalidates")
	login, adminID := phase4test.ProvisionBootstrapAdmin(t, harness.Server)
	incident := phase4test.CreateIncident(t, harness.Server, login, map[string]any{
		"client_txn_id": "txn-phase5-i-03-incident",
		"incident_key":  "phase5-i-03",
		"title":         "Phase 5 handle invalidation",
	})
	incidentID := phase4test.MustUUID(t, incident["incident_id"].(string))

	activeLogin := login
	activeAdminID := adminID
	activeIncidentID := incidentID

	t.Run("membership loss and wrong session hide the handle", func(t *testing.T) {
		recordID := uuid.New()
		seedEvidenceRecord(t, harness, activeIncidentID, activeAdminID, recordID)
		attachUploadedBlobWithMetadata(t, harness, activeLogin, activeIncidentID, recordID, []byte("membership body"), "membership.txt", "text/plain", "txn-phase5-i-03-membership-blob", "txn-phase5-i-03-membership-attach")
		handle := issueEvidenceHandle(t, harness, activeLogin, recordID, "preview-handle")
		if _, err := harness.DB.ExecContext(context.Background(), `
DELETE FROM incident_memberships
 WHERE incident_id = $1
   AND user_id = $2
`, activeIncidentID, activeAdminID); err != nil {
			t.Fatalf("remove incident membership: %v", err)
		}
		httptestx.RequireErrorEnvelope(t,
			phase4test.DoJSON(t, http.MethodGet, harness.Server.HTTP.URL+handle["href"].(string), nil, phase4test.WithCookies(activeLogin.SessionCookie)),
			http.StatusNotFound,
			"handle_not_found_or_revoked",
		)

		phase4test.SeedIncidentMembership(t, harness.DB, activeIncidentID, activeAdminID, "Bootstrap Admin", "admin", activeAdminID)
		otherUser := phase4test.SeedLocalUserFlags(t, harness.DB, "phase5-i-03-other@example.test", "Phase5 Other", "Phase5Other1!", false, false, true)
		phase4test.SeedIncidentMembership(t, harness.DB, activeIncidentID, otherUser.ID, "Phase5 Other", "admin", activeAdminID)
		otherLogin := loginLocalUserNoMFA(t, harness, "phase5-i-03-other@example.test", "Phase5Other1!")
		httptestx.RequireErrorEnvelope(t,
			phase4test.DoJSON(t, http.MethodGet, harness.Server.HTTP.URL+handle["href"].(string), nil, phase4test.WithCookies(otherLogin.SessionCookie)),
			http.StatusNotFound,
			"handle_not_found_or_revoked",
		)
	})

	scenarios := []struct {
		name       string
		reasonCode string
		mutate     func(recordID uuid.UUID, objectBlobID uuid.UUID)
	}{
		{name: "blob detach", reasonCode: "no_visible_blob", mutate: func(recordID uuid.UUID, objectBlobID uuid.UUID) {
			updateEvidenceBlobLink(t, harness, recordID, nil)
		}},
		{name: "blob replacement", reasonCode: "evidence_inconsistent", mutate: func(recordID uuid.UUID, objectBlobID uuid.UUID) {
			replacementID := linkSeededBlob(t, harness, activeIncidentID, activeAdminID, recordID, "available", "available", "phase5/i-03/replacement-"+recordID.String())
			updateEvidenceBlobLink(t, harness, recordID, &replacementID)
		}},
		{name: "evidence delete", reasonCode: "evidence_inconsistent", mutate: func(recordID uuid.UUID, objectBlobID uuid.UUID) {
			updateRecordDeletedState(t, harness, recordID, true)
		}},
		{name: "evidence restore", reasonCode: "evidence_inconsistent", mutate: func(recordID uuid.UUID, objectBlobID uuid.UUID) {
			updateRecordDeletedState(t, harness, recordID, true)
			updateRecordDeletedState(t, harness, recordID, false)
		}},
		{name: "quarantine", reasonCode: "evidence_quarantined", mutate: func(recordID uuid.UUID, objectBlobID uuid.UUID) {
			updateBlobState(t, harness, objectBlobID, "quarantined")
		}},
		{name: "pending transition", reasonCode: "blob_pending", mutate: func(recordID uuid.UUID, objectBlobID uuid.UUID) {
			updateBlobState(t, harness, objectBlobID, "pending")
		}},
		{name: "failed transition", reasonCode: "blob_failed", mutate: func(recordID uuid.UUID, objectBlobID uuid.UUID) {
			updateBlobState(t, harness, objectBlobID, "failed")
		}},
		{name: "object missing", reasonCode: "blob_missing", mutate: func(recordID uuid.UUID, objectBlobID uuid.UUID) {
			deleteBlobObject(t, harness, blobStorageKey(t, harness, objectBlobID))
		}},
		{name: "metadata mismatch", reasonCode: "evidence_inconsistent", mutate: func(recordID uuid.UUID, objectBlobID uuid.UUID) {
			updateBlobObservedMetadata(t, harness, objectBlobID, int64(999), "text/plain", "")
		}},
	}

	for _, scenario := range scenarios {
		for _, endpoint := range []string{"preview-handle", "download-handle"} {
			t.Run(endpoint+" "+scenario.name, func(t *testing.T) {
				recordID := uuid.New()
				seedEvidenceRecord(t, harness, activeIncidentID, activeAdminID, recordID)
				attachData := attachUploadedBlobWithMetadata(t, harness, activeLogin, activeIncidentID, recordID, []byte("phase5 invalidation"), "invalidate.txt", "text/plain", "txn-"+recordID.String()+"-blob", "txn-"+recordID.String()+"-attach")
				objectBlobID := phase4test.MustUUID(t, attachData["object_blob_id"].(string))
				handle := issueEvidenceHandle(t, harness, activeLogin, recordID, endpoint)
				scenario.mutate(recordID, objectBlobID)
				requireEvidenceAccessUnavailableReason(t,
					phase4test.DoJSON(t, http.MethodGet, harness.Server.HTTP.URL+handle["href"].(string), nil, phase4test.WithCookies(activeLogin.SessionCookie)),
					scenario.reasonCode,
				)
			})
		}
	}

	t.Run("logout revokes issuing session before lookup", func(t *testing.T) {
		recordID := uuid.New()
		seedEvidenceRecord(t, harness, activeIncidentID, activeAdminID, recordID)
		attachUploadedBlobWithMetadata(t, harness, activeLogin, activeIncidentID, recordID, []byte("logout body"), "logout.txt", "text/plain", "txn-phase5-i-03-logout-blob", "txn-phase5-i-03-logout-attach")
		handle := issueEvidenceHandle(t, harness, activeLogin, recordID, "preview-handle")
		logout := phase4test.DoJSON(t, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/auth/logout", map[string]any{}, authOptions(activeLogin)...)
		httptestx.RequireSuccessEnvelope(t, logout, http.StatusOK)
		httptestx.RequireErrorEnvelope(t,
			phase4test.DoJSON(t, http.MethodGet, harness.Server.HTTP.URL+handle["href"].(string), nil, phase4test.WithCookies(activeLogin.SessionCookie)),
			http.StatusUnauthorized,
			"session_required",
		)
	})
}

func TestPhase5_QuarantineBoundaryPreservesTwoStepAttach_I_5_04(t *testing.T) {
	t.Run("AC-405 object bytes stay outside structured state and loss fails closed", func(t *testing.T) {
		harness := phase4test.StartServer(t, "phase5-i-04-object-boundary")
		login, adminID := phase4test.ProvisionBootstrapAdmin(t, harness.Server)
		incident := phase4test.CreateIncident(t, harness.Server, login, map[string]any{
			"client_txn_id": "txn-phase5-i-04-boundary-incident",
			"incident_key":  "phase5-i-04-boundary",
			"title":         "Phase 5 object boundary",
		})
		incidentID := phase4test.MustUUID(t, incident["incident_id"].(string))
		recordID := uuid.New()
		seedEvidenceRecord(t, harness, incidentID, adminID, recordID)

		marker := "phase5-ac405-marker-" + uuid.NewString() + "-payload"
		payload := []byte("prefix-" + marker + "-suffix")
		attachData := attachUploadedBlobWithMetadata(t, harness, login, incidentID, recordID, payload, "boundary.txt", "text/plain", "txn-phase5-i-04-boundary-blob", "txn-phase5-i-04-boundary-attach")
		objectBlobID := phase4test.MustUUID(t, attachData["object_blob_id"].(string))

		preview := issueEvidenceHandle(t, harness, login, recordID, "preview-handle")
		if got := string(redeemHandle(t, harness.Server.HTTP.URL+preview["href"].(string), login)); got != string(payload) {
			t.Fatalf("preview before object loss got %q want %q", got, string(payload))
		}
		download := issueEvidenceHandle(t, harness, login, recordID, "download-handle")
		if got := string(redeemHandle(t, harness.Server.HTTP.URL+download["href"].(string), login)); got != string(payload) {
			t.Fatalf("download before object loss got %q want %q", got, string(payload))
		}

		structured := structuredTableText(t, harness,
			"records", "evidence", "object_blobs", "route_idempotency",
			"change_sets", "change_set_mutations", "record_revisions", "evidence_access_handles",
		)
		if strings.Contains(structured, marker) {
			t.Fatalf("structured Postgres state contains inline evidence payload marker %q", marker)
		}

		storageKey := blobStorageKey(t, harness, objectBlobID)
		if err := harness.Server.Runtime.ObjectStore.DeleteObject(context.Background(), storageKey); err != nil {
			t.Fatalf("delete object bytes: %v", err)
		}
		if got := countEvidenceBlobLinks(t, harness, recordID); got != 1 {
			t.Fatalf("object loss changed committed evidence blob link count: got %d want 1", got)
		}
		requireEvidenceStates(t, harness, recordID, "available", "available")
		for _, endpoint := range []string{"preview-handle", "download-handle"} {
			requireEvidenceAccessUnavailableReason(t,
				phase4test.DoJSON(t, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/evidence-records/"+recordID.String()+"/"+endpoint, map[string]any{}, authOptions(login)...),
				"blob_missing",
			)
		}
	})

	t.Run("failed unattached cleanup deletes bytes and retains metadata", func(t *testing.T) {
		harness := phase4test.StartServer(t, "phase5-i-04-cleanup")
		login, adminID := phase4test.ProvisionBootstrapAdmin(t, harness.Server)
		incident := phase4test.CreateIncident(t, harness.Server, login, map[string]any{
			"client_txn_id": "txn-phase5-i-04-cleanup-incident",
			"incident_key":  "phase5-i-04-cleanup",
			"title":         "Phase 5 cleanup",
		})
		incidentID := phase4test.MustUUID(t, incident["incident_id"].(string))
		now := time.Now().UTC().Truncate(time.Second)

		cleanupBlobID := uuid.New()
		cleanupKey := "phase5/i-04/cleanup/" + cleanupBlobID.String()
		cleanupPayload := "phase5 cleanup orphan bytes"
		if err := harness.Server.Runtime.ObjectStore.PutObject(context.Background(), cleanupKey, strings.NewReader(cleanupPayload), int64(len(cleanupPayload)), "text/plain"); err != nil {
			t.Fatalf("put cleanup candidate object: %v", err)
		}
		insertFailedCleanupBlob(t, harness, incidentID, adminID, cleanupBlobID, cleanupKey, now)

		expiredBlobID := uuid.New()
		insertExpiredPendingBlob(t, harness, incidentID, adminID, expiredBlobID, "phase5/i-04/expired/"+expiredBlobID.String(), now)

		result, err := evidence.NewStore(harness.Server.Runtime.Postgres).CleanupFailedUnattachedBlobBytes(context.Background(), harness.Server.Runtime.ObjectStore, now, 10)
		if err != nil {
			t.Fatalf("cleanup failed unattached blob bytes: %v", err)
		}
		if result.ExpiredPendingCount != 1 || result.CleanedBlobCount != 1 {
			t.Fatalf("cleanup result got expired=%d cleaned=%d want 1/1", result.ExpiredPendingCount, result.CleanedBlobCount)
		}
		if _, err := harness.Server.Runtime.ObjectStore.StatObject(context.Background(), cleanupKey); err == nil {
			t.Fatalf("cleanup candidate object bytes still exist at %s", cleanupKey)
		}
		requireCleanedFailedBlobMetadata(t, harness, cleanupBlobID)
		requireExpiredPendingFailed(t, harness, expiredBlobID)
		if got := countEvidenceRowsForBlob(t, harness, cleanupBlobID); got != 0 {
			t.Fatalf("cleanup created or retained evidence link rows: got %d want 0", got)
		}
	})

	t.Run("quarantine bridges evidence and blocks attach preview and download", func(t *testing.T) {
		harness := phase4test.StartServer(t, "phase5-i-04-quarantine")
		login, adminID := phase4test.ProvisionBootstrapAdmin(t, harness.Server)
		incident := phase4test.CreateIncident(t, harness.Server, login, map[string]any{
			"client_txn_id": "txn-phase5-i-04-quarantine-incident",
			"incident_key":  "phase5-i-04-quarantine",
			"title":         "Phase 5 quarantine",
		})
		incidentID := phase4test.MustUUID(t, incident["incident_id"].(string))
		recordID := uuid.New()
		seedEvidenceRecord(t, harness, incidentID, adminID, recordID)
		attachData := attachUploadedBlobWithMetadata(t, harness, login, incidentID, recordID, []byte("phase5 quarantine body"), "quarantine.txt", "text/plain", "txn-phase5-i-04-quarantine-blob", "txn-phase5-i-04-quarantine-attach")
		objectBlobID := phase4test.MustUUID(t, attachData["object_blob_id"].(string))
		preview := issueEvidenceHandle(t, harness, login, recordID, "preview-handle")
		download := issueEvidenceHandle(t, harness, login, recordID, "download-handle")
		beforeRevisions := countEvidenceRevisions(t, harness, recordID)

		store := evidence.NewStore(harness.Server.Runtime.Postgres)
		if _, err := store.QuarantineBlob(context.Background(), adminID, objectBlobID, "unsupported_trigger", "req-phase5-i-04-bad-trigger", time.Now().UTC()); !errors.Is(err, evidence.ErrIllegalBlobTransition) {
			t.Fatalf("unsupported quarantine trigger got %v want ErrIllegalBlobTransition", err)
		}
		result, err := store.QuarantineBlob(context.Background(), adminID, objectBlobID, "content_inspection_quarantine", "req-phase5-i-04-quarantine", time.Now().UTC())
		if err != nil {
			t.Fatalf("quarantine blob: %v", err)
		}
		if result.ChangedEvidenceRecord != 1 || result.ChangeSetID == uuid.Nil {
			t.Fatalf("quarantine result got changed=%d change_set=%s", result.ChangedEvidenceRecord, result.ChangeSetID)
		}
		requireEvidenceStates(t, harness, recordID, "quarantined", "quarantined")
		if got := countEvidenceRevisions(t, harness, recordID); got != beforeRevisions+1 {
			t.Fatalf("quarantine revision count got %d want %d", got, beforeRevisions+1)
		}
		requireChangeSetSource(t, harness, result.ChangeSetID, "evidence.blob.quarantine")

		for _, handle := range []map[string]any{preview, download} {
			requireEvidenceAccessUnavailableReason(t,
				phase4test.DoJSON(t, http.MethodGet, harness.Server.HTTP.URL+handle["href"].(string), nil, phase4test.WithCookies(login.SessionCookie)),
				"evidence_quarantined",
			)
		}
		for _, endpoint := range []string{"preview-handle", "download-handle"} {
			requireEvidenceAccessUnavailableReason(t,
				phase4test.DoJSON(t, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/evidence-records/"+recordID.String()+"/"+endpoint, map[string]any{}, authOptions(login)...),
				"evidence_quarantined",
			)
		}
		secondRecordID := uuid.New()
		seedEvidenceRecord(t, harness, incidentID, adminID, secondRecordID)
		attachBlocked := httptestx.RequireErrorEnvelope(t,
			phase4test.DoJSON(t, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/evidence-records/"+secondRecordID.String()+"/attach-blob", map[string]any{
				"object_blob_id":   objectBlobID.String(),
				"base_row_version": 1,
				"client_txn_id":    "txn-phase5-i-04-quarantine-attach-blocked",
			}, authOptions(login)...),
			http.StatusConflict,
			"evidence_attach_rejected",
		)
		attachDetails := attachBlocked["error"].(map[string]any)["details"].(map[string]any)
		if got := attachDetails["reason_code"]; got != evidence.AttachReasonBlobQuarantined {
			t.Fatalf("quarantined attach reason got %v want %s", got, evidence.AttachReasonBlobQuarantined)
		}

		pendingID := uuid.New()
		insertExpiredPendingBlob(t, harness, incidentID, adminID, pendingID, "phase5/i-04/pending-quarantine/"+pendingID.String(), time.Now().UTC().Add(2*time.Hour))
		if _, err := store.QuarantineBlob(context.Background(), adminID, pendingID, "admin_quarantine", "req-phase5-i-04-pending-quarantine", time.Now().UTC()); !errors.Is(err, evidence.ErrIllegalBlobTransition) {
			t.Fatalf("pending quarantine got %v want ErrIllegalBlobTransition", err)
		}
	})

	t.Run("active content uses observed media state for preview policy", func(t *testing.T) {
		for _, active := range []struct {
			name        string
			contentType string
			filename    string
		}{
			{name: "html", contentType: "text/html", filename: "pretend-image.png"},
			{name: "svg", contentType: "image/svg+xml", filename: "pretend-raster.png"},
		} {
			t.Run(active.name, func(t *testing.T) {
				harness := phase4test.StartServer(t, "phase5-i-04-active-"+active.name)
				login, adminID := phase4test.ProvisionBootstrapAdmin(t, harness.Server)
				incident := phase4test.CreateIncident(t, harness.Server, login, map[string]any{
					"client_txn_id": "txn-phase5-i-04-active-" + active.name + "-incident",
					"incident_key":  "phase5-i-04-active-" + active.name,
					"title":         "Phase 5 active content " + active.name,
				})
				incidentID := phase4test.MustUUID(t, incident["incident_id"].(string))
				recordID := uuid.New()
				seedEvidenceRecord(t, harness, incidentID, adminID, recordID)
				payload := []byte("<script>window.__cartulary_phase5_active_content = true</script>")
				attachData := attachUploadedBlobWithHints(t, harness, login, incidentID, recordID, payload, active.filename, "image/png", active.contentType, "txn-phase5-i-04-active-"+active.name+"-blob", "txn-phase5-i-04-active-"+active.name+"-attach")
				objectBlobID := phase4test.MustUUID(t, attachData["object_blob_id"].(string))
				requireObservedContentType(t, harness, objectBlobID, active.contentType)

				requireEvidenceAccessUnavailableReason(t,
					phase4test.DoJSON(t, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/evidence-records/"+recordID.String()+"/preview-handle", map[string]any{}, authOptions(login)...),
					"unsupported_preview",
				)
				download := issueEvidenceHandle(t, harness, login, recordID, "download-handle")
				if got := string(redeemHandle(t, harness.Server.HTTP.URL+download["href"].(string), login)); got != string(payload) {
					t.Fatalf("download active content got %q want %q", got, string(payload))
				}
			})
		}
	})
}

func requireCreateExpiry(t testing.TB, data map[string]any, field string, want time.Time) {
	t.Helper()
	gotRaw, ok := data[field].(string)
	if !ok {
		t.Fatalf("%s got %T in %#v", field, data[field], data)
	}
	got := mustParseTime(t, gotRaw)
	if !got.Equal(want.UTC()) {
		t.Fatalf("%s got %s want %s", field, got, want.UTC())
	}
	if field == "target_expires_at" {
		target, ok := data["upload_target"].(map[string]any)
		if !ok {
			t.Fatalf("upload_target got %T", data["upload_target"])
		}
		if target["expires_at"] != gotRaw {
			t.Fatalf("upload_target.expires_at got %#v want %q", target["expires_at"], gotRaw)
		}
	}
}

func mustParseTime(t testing.TB, value string) time.Time {
	t.Helper()
	parsed, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		t.Fatalf("parse time %q: %v", value, err)
	}
	return parsed.UTC()
}

func extendSessionForClockJump(t testing.TB, harness *phase4test.ServerHarness, userID any, expiresAt time.Time) {
	t.Helper()
	if _, err := harness.DB.ExecContext(context.Background(), `
UPDATE user_sessions
   SET idle_expires_at = $2,
       absolute_expires_at = $2,
       session_expires_at = $2,
       updated_at = now()
 WHERE user_id = $1
   AND revoked_at IS NULL
`, userID, expiresAt.UTC()); err != nil {
		t.Fatalf("extend test session for clock jump: %v", err)
	}
}

func requirePhase5HTTPWorkbookCreate(t testing.TB, harness *phase4test.ServerHarness, login phase4test.LoginResult, incidentID uuid.UUID, viewSchemaID string, body map[string]any) map[string]any {
	t.Helper()
	resp := phase4test.DoJSON(t, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID.String()+"/views/"+viewSchemaID+"/rows", body, authOptions(login)...)
	return httptestx.RequireSuccessEnvelope(t, resp, http.StatusCreated)["data"].(map[string]any)
}

func requirePhase5HTTPWorkbookPatch(t testing.TB, harness *phase4test.ServerHarness, login phase4test.LoginResult, recordID uuid.UUID, body map[string]any) map[string]any {
	t.Helper()
	resp := phase4test.DoJSON(t, http.MethodPatch, harness.Server.HTTP.URL+"/api/v1/records/"+recordID.String(), body, authOptions(login)...)
	return httptestx.RequireSuccessEnvelope(t, resp, http.StatusOK)["data"].(map[string]any)
}

func requireTimelineEvidenceProjection(t testing.TB, harness *phase4test.ServerHarness, login phase4test.LoginResult, incidentID uuid.UUID, recordID uuid.UUID, wantCount int, wantHasEvidence bool) {
	t.Helper()
	resp := phase4test.DoJSON(t, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID.String()+"/views/cartulary.view.timeline.v1/query", map[string]any{}, authOptions(login)...)
	data := httptestx.RequireSuccessEnvelope(t, resp, http.StatusOK)["data"].(map[string]any)
	for _, raw := range data["rows"].([]any) {
		row := raw.(map[string]any)
		if row["record_id"] != recordID.String() {
			continue
		}
		cells := row["cells"].(map[string]any)
		if got := int(cells["timeline.evidence_count"].(map[string]any)["value"].(float64)); got != wantCount {
			t.Fatalf("timeline.evidence_count got %d want %d in row %#v", got, wantCount, row)
		}
		if got := cells["timeline.has_evidence"].(map[string]any)["value"].(bool); got != wantHasEvidence {
			t.Fatalf("timeline.has_evidence got %v want %v in row %#v", got, wantHasEvidence, row)
		}
		return
	}
	t.Fatalf("timeline row %s not found in query %#v", recordID, data["rows"])
}

type phase5SourceHistoryCounts struct {
	ChangeSets   int
	Mutations    int
	Revisions    int
	RecordLinks  int
	TimelineRows int
}

func phase5ProcessCounts(t testing.TB, harness *phase4test.ServerHarness, incidentID uuid.UUID, recordID uuid.UUID) phase5SourceHistoryCounts {
	t.Helper()
	var counts phase5SourceHistoryCounts
	if err := harness.DB.QueryRowContext(context.Background(), `
SELECT
    (SELECT COUNT(*) FROM change_sets WHERE incident_id = $1),
    (SELECT COUNT(*) FROM change_set_mutations m JOIN change_sets cs ON cs.change_set_id = m.change_set_id WHERE cs.incident_id = $1),
    (SELECT COUNT(*) FROM record_revisions WHERE record_id = $2),
    (SELECT COUNT(*) FROM record_links WHERE incident_id = $1 AND deleted_at IS NULL),
    (SELECT COUNT(*) FROM timeline_events WHERE incident_id = $1)
`, incidentID, recordID).Scan(&counts.ChangeSets, &counts.Mutations, &counts.Revisions, &counts.RecordLinks, &counts.TimelineRows); err != nil {
		t.Fatalf("count phase5 source/history state: %v", err)
	}
	return counts
}

func requireTimelineProjectionStorage(t testing.TB, harness *phase4test.ServerHarness, recordID uuid.UUID, wantCount int, wantHasEvidence bool) {
	t.Helper()
	var count int
	var hasEvidence bool
	if err := harness.DB.QueryRowContext(context.Background(), `
SELECT evidence_count, has_evidence
  FROM timeline_grid_projection
 WHERE record_id = $1
`, recordID).Scan(&count, &hasEvidence); err != nil {
		t.Fatalf("load timeline projection storage: %v", err)
	}
	if count != wantCount || hasEvidence != wantHasEvidence {
		t.Fatalf("timeline projection storage got count=%d has_evidence=%v want count=%d has_evidence=%v", count, hasEvidence, wantCount, wantHasEvidence)
	}
}

func countEvidenceRevisions(t testing.TB, harness *phase4test.ServerHarness, recordID uuid.UUID) int {
	t.Helper()
	var count int
	if err := harness.DB.QueryRowContext(context.Background(), `SELECT COUNT(*) FROM record_revisions WHERE record_id = $1`, recordID).Scan(&count); err != nil {
		t.Fatalf("count evidence revisions: %v", err)
	}
	return count
}

func countEvidenceBlobLinks(t testing.TB, harness *phase4test.ServerHarness, recordID uuid.UUID) int {
	t.Helper()
	var count int
	if err := harness.DB.QueryRowContext(context.Background(), `SELECT COUNT(*) FROM evidence WHERE record_id = $1 AND object_blob_id IS NOT NULL`, recordID).Scan(&count); err != nil {
		t.Fatalf("count evidence blob links: %v", err)
	}
	return count
}

func countAttachedEvidenceLinks(t testing.TB, harness *phase4test.ServerHarness, incidentID uuid.UUID, srcRecordID uuid.UUID, dstRecordID uuid.UUID) int {
	t.Helper()
	var count int
	if err := harness.DB.QueryRowContext(context.Background(), `
SELECT COUNT(*)
  FROM record_links
 WHERE incident_id = $1
   AND src_record_id = $2
   AND dst_record_id = $3
   AND link_type = 'attached_evidence'
   AND field_key = 'timeline.attached_evidence_ids'
   AND deleted_at IS NULL
`, incidentID, srcRecordID, dstRecordID).Scan(&count); err != nil {
		t.Fatalf("count attached evidence links: %v", err)
	}
	return count
}

func insertPhase5RouteBlob(t testing.TB, harness *phase4test.ServerHarness, incidentID uuid.UUID, actorID uuid.UUID, uploadState string) uuid.UUID {
	t.Helper()
	objectBlobID := uuid.New()
	storageKey, err := blobref.ObjectBlobStorageKey(incidentID, objectBlobID)
	if err != nil {
		t.Fatalf("phase5 route blob storage key: %v", err)
	}
	now := time.Now().UTC()
	var finalizedAt any
	var terminalReason any
	var failedAt any
	if uploadState == "available" || uploadState == "quarantined" {
		finalizedAt = now
	}
	if uploadState == "failed" {
		terminalReason = "pending_timeout"
		failedAt = now
	}
	if _, err := harness.DB.ExecContext(context.Background(), `
INSERT INTO object_blobs (
    object_blob_id, incident_id, created_by_user_id, storage_key, upload_state,
    byte_size, filename_hint, content_type_hint, observed_size, observed_content_type,
    observed_sha256_hex, target_expires_at, pending_expires_at, finalized_at,
    terminal_reason, failed_at, cleanup_due_at, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5,
    11, 'route.txt', 'text/plain', 11, 'text/plain',
    '0000000000000000000000000000000000000000000000000000000000000000',
    $6, $7, $8, $9, $10, $11, $12, $12
)
`, objectBlobID, incidentID, actorID, storageKey, uploadState,
		now.Add(time.Hour), now.Add(24*time.Hour), finalizedAt, terminalReason, failedAt, nullableCleanupDue(uploadState, now), now); err != nil {
		t.Fatalf("insert phase5 route blob: %v", err)
	}
	return objectBlobID
}

func nullableCleanupDue(uploadState string, now time.Time) any {
	if uploadState == "failed" {
		return now.Add(time.Hour)
	}
	return nil
}

func loginLocalUserNoMFA(t testing.TB, harness *phase4test.ServerHarness, username string, password string) phase4test.LoginResult {
	t.Helper()
	resp := phase4test.DoJSON(t, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/auth/login", map[string]any{
		"username": username,
		"password": password,
	})
	httptestx.RequireSuccessEnvelope(t, resp, http.StatusOK)
	var sessionCookie *http.Cookie
	var csrfCookie *http.Cookie
	for _, cookie := range resp.Cookies() {
		if cookie.Name == authn.SessionCookieName {
			sessionCookie = cookie
		}
		if cookie.Name == authn.CSRFCookieName {
			csrfCookie = cookie
		}
	}
	if sessionCookie == nil || csrfCookie == nil {
		t.Fatalf("login did not set session and csrf cookies: %#v", resp.Cookies())
	}
	return phase4test.LoginResult{SessionCookie: sessionCookie, CSRFCookie: csrfCookie}
}

func updateRecordDeletedState(t testing.TB, harness *phase4test.ServerHarness, recordID uuid.UUID, deleted bool) {
	t.Helper()
	deletedAt := any(nil)
	if deleted {
		deletedAt = time.Now().UTC()
	}
	if _, err := harness.DB.ExecContext(context.Background(), `
UPDATE records
   SET deleted_at = $2,
       deleted_by_user_id = CASE WHEN $2::timestamptz IS NULL THEN NULL ELSE created_by_user_id END,
       row_version = row_version + 1,
       updated_at = now()
 WHERE record_id = $1
`, recordID, deletedAt); err != nil {
		t.Fatalf("update record deleted state: %v", err)
	}
}

func structuredTableText(t testing.TB, harness *phase4test.ServerHarness, tables ...string) string {
	t.Helper()
	var builder strings.Builder
	for _, table := range tables {
		query := fmt.Sprintf(`SELECT COALESCE(string_agg(to_jsonb(t)::text, E'\n'), '') FROM %s AS t`, quoteIdent(table))
		var text string
		if err := harness.DB.QueryRowContext(context.Background(), query).Scan(&text); err != nil {
			t.Fatalf("dump structured table %s: %v", table, err)
		}
		builder.WriteString(table)
		builder.WriteByte('\n')
		builder.WriteString(text)
		builder.WriteByte('\n')
	}
	return builder.String()
}

func quoteIdent(identifier string) string {
	return `"` + strings.ReplaceAll(identifier, `"`, `""`) + `"`
}

func insertFailedCleanupBlob(t testing.TB, harness *phase4test.ServerHarness, incidentID uuid.UUID, actorID uuid.UUID, objectBlobID uuid.UUID, storageKey string, now time.Time) {
	t.Helper()
	if _, err := harness.DB.ExecContext(context.Background(), `
INSERT INTO object_blobs (
    object_blob_id, incident_id, created_by_user_id, storage_key, upload_state,
    byte_size, filename_hint, content_type_hint, observed_size, observed_content_type,
    observed_sha256_hex, target_expires_at, pending_expires_at, finalized_at,
    terminal_reason, failed_at, cleanup_due_at, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, 'failed',
    27, 'cleanup.txt', 'text/plain', 27, 'text/plain',
    '0000000000000000000000000000000000000000000000000000000000000000',
    $5::timestamptz - interval '3 hours', $5::timestamptz - interval '2 hours', NULL,
    'pending_timeout', $5::timestamptz - interval '2 hours', $5::timestamptz - interval '1 hour',
    $5::timestamptz - interval '3 hours', $5::timestamptz - interval '2 hours'
)
`, objectBlobID, incidentID, actorID, storageKey, now.UTC()); err != nil {
		t.Fatalf("insert failed cleanup blob: %v", err)
	}
}

func insertExpiredPendingBlob(t testing.TB, harness *phase4test.ServerHarness, incidentID uuid.UUID, actorID uuid.UUID, objectBlobID uuid.UUID, storageKey string, now time.Time) {
	t.Helper()
	if _, err := harness.DB.ExecContext(context.Background(), `
INSERT INTO object_blobs (
    object_blob_id, incident_id, created_by_user_id, storage_key, upload_state,
    byte_size, filename_hint, content_type_hint,
    target_expires_at, pending_expires_at, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, 'pending',
    10, 'expired.txt', 'text/plain',
    $5::timestamptz - interval '2 hours', $5::timestamptz - interval '1 hour',
    $5::timestamptz - interval '25 hours', $5::timestamptz - interval '25 hours'
)
`, objectBlobID, incidentID, actorID, storageKey, now.UTC()); err != nil {
		t.Fatalf("insert expired pending blob: %v", err)
	}
}

func requireCleanedFailedBlobMetadata(t testing.TB, harness *phase4test.ServerHarness, objectBlobID uuid.UUID) {
	t.Helper()
	var uploadState string
	var cleaned bool
	if err := harness.DB.QueryRowContext(context.Background(), `
SELECT upload_state, cleaned_up_at IS NOT NULL
  FROM object_blobs
 WHERE object_blob_id = $1
`, objectBlobID).Scan(&uploadState, &cleaned); err != nil {
		t.Fatalf("load cleaned blob metadata: %v", err)
	}
	if uploadState != "failed" || !cleaned {
		t.Fatalf("cleaned failed blob metadata got state=%s cleaned=%v want failed/true", uploadState, cleaned)
	}
}

func requireExpiredPendingFailed(t testing.TB, harness *phase4test.ServerHarness, objectBlobID uuid.UUID) {
	t.Helper()
	var uploadState string
	var terminalReason string
	var cleanupDue bool
	var cleaned bool
	if err := harness.DB.QueryRowContext(context.Background(), `
SELECT upload_state, terminal_reason, cleanup_due_at IS NOT NULL, cleaned_up_at IS NOT NULL
  FROM object_blobs
 WHERE object_blob_id = $1
`, objectBlobID).Scan(&uploadState, &terminalReason, &cleanupDue, &cleaned); err != nil {
		t.Fatalf("load expired pending blob: %v", err)
	}
	if uploadState != "failed" || terminalReason != "pending_timeout" || !cleanupDue || cleaned {
		t.Fatalf("expired pending blob got state=%s reason=%s cleanup_due=%v cleaned=%v", uploadState, terminalReason, cleanupDue, cleaned)
	}
}

func countEvidenceRowsForBlob(t testing.TB, harness *phase4test.ServerHarness, objectBlobID uuid.UUID) int {
	t.Helper()
	var count int
	if err := harness.DB.QueryRowContext(context.Background(), `
SELECT count(*)
  FROM evidence
 WHERE object_blob_id = $1
`, objectBlobID).Scan(&count); err != nil {
		t.Fatalf("count evidence rows for blob: %v", err)
	}
	return count
}

func requireEvidenceStates(t testing.TB, harness *phase4test.ServerHarness, recordID uuid.UUID, wantLifecycle string, wantUpload string) {
	t.Helper()
	var lifecycle string
	var upload string
	if err := harness.DB.QueryRowContext(context.Background(), `
SELECT e.lifecycle_state, COALESCE(b.upload_state, e.upload_state)
  FROM evidence e
  LEFT JOIN object_blobs b ON b.object_blob_id = e.object_blob_id
 WHERE e.record_id = $1
`, recordID).Scan(&lifecycle, &upload); err != nil {
		t.Fatalf("load evidence states: %v", err)
	}
	if lifecycle != wantLifecycle || upload != wantUpload {
		t.Fatalf("evidence states got lifecycle=%s upload=%s want %s/%s", lifecycle, upload, wantLifecycle, wantUpload)
	}
}

func requireChangeSetSource(t testing.TB, harness *phase4test.ServerHarness, changeSetID uuid.UUID, wantSource string) {
	t.Helper()
	var source string
	if err := harness.DB.QueryRowContext(context.Background(), `
SELECT source
  FROM change_sets
 WHERE change_set_id = $1
`, changeSetID).Scan(&source); err != nil {
		t.Fatalf("load change set source: %v", err)
	}
	if source != wantSource {
		t.Fatalf("change set source got %s want %s", source, wantSource)
	}
}

func requireObservedContentType(t testing.TB, harness *phase4test.ServerHarness, objectBlobID uuid.UUID, want string) {
	t.Helper()
	var got string
	if err := harness.DB.QueryRowContext(context.Background(), `
SELECT observed_content_type
  FROM object_blobs
 WHERE object_blob_id = $1
`, objectBlobID).Scan(&got); err != nil {
		t.Fatalf("load observed content type: %v", err)
	}
	if got != want {
		t.Fatalf("observed content type got %q want %q", got, want)
	}
}

func attachUploadedBlobWithHints(t *testing.T, harness *phase4test.ServerHarness, login phase4test.LoginResult, incidentID uuid.UUID, recordID uuid.UUID, payload []byte, filename string, hintContentType string, uploadContentType string, createTxn string, attachTxn string) map[string]any {
	t.Helper()
	createResp := phase4test.DoJSON(t, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/object-blobs", map[string]any{
		"incident_id":       incidentID.String(),
		"client_txn_id":     createTxn,
		"byte_size":         len(payload),
		"filename_hint":     filename,
		"content_type_hint": hintContentType,
		"sha256_hex":        fmt.Sprintf("%x", sha256Sum(payload)),
	}, authOptions(login)...)
	createData := httptestx.RequireSuccessEnvelope(t, createResp, http.StatusCreated)["data"].(map[string]any)
	putObject(t, createData["upload_target"].(map[string]any)["href"].(string), payload, uploadContentType)
	attachResp := phase4test.DoJSON(t, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/evidence-records/"+recordID.String()+"/attach-blob", map[string]any{
		"object_blob_id":   createData["object_blob_id"],
		"base_row_version": 1,
		"client_txn_id":    attachTxn,
	}, authOptions(login)...)
	return httptestx.RequireSuccessEnvelope(t, attachResp, http.StatusOK)["data"].(map[string]any)
}
