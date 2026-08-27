package evidence_test

// Evidence projection and cross-owner effect contracts.
import (
	"context"
	"crypto/sha256"
	"fmt"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/JochiRaider/cartulary/internal/testutil/revisionsupport"
	"net/http"
	"reflect"
	"strings"
	"testing"
)

func TestObjectUploadAttachWorkbookProjection_Integration(t *testing.T) {
	harness := appsupport.StartServer(t, "evidence_lifecycle-upload-attach-projection")
	login, _ := appsupport.ProvisionBootstrapAdmin(t, harness.Server)
	incident := appsupport.CreateIncident(t, harness.Server, login, map[string]any{
		"client_txn_id": "txn-evidence_lifecycle-i-01-incident",
		"incident_key":  "evidence_lifecycle-i-01",
		"title":         "Evidence upload attach projection",
	})
	incidentID := appsupport.MustUUID(t, incident["incident_id"].(string))

	timelineData := requireHTTPWorkbookCreate(t, harness, login, incidentID, "cartulary.view.timeline.v2", map[string]any{
		"client_txn_id":                   "txn-evidence_lifecycle-i-01-timeline",
		"timeline.activity_synopsis_text": "Endpoint screenshot received",
	})
	timelineRow := timelineData["row"].(map[string]any)
	timelineRecordID := appsupport.MustUUID(t, timelineRow["record_id"].(string))
	timelineRowVersion := int(timelineRow["row_version"].(float64))
	requireTimelineEvidenceProjection(t, harness, login, incidentID, timelineRecordID, 0, false)
	evidenceData := requireHTTPWorkbookCreate(t, harness, login, incidentID, "cartulary.view.evidence.v1", map[string]any{
		"client_txn_id":                 "txn-evidence_lifecycle-i-01-evidence",
		"evidence.title":                "Endpoint screenshot",
		"evidence.collector_party_text": "IR collector",
	})
	evidenceRecordID := appsupport.MustUUID(t, evidenceData["row"].(map[string]any)["record_id"].(string))
	requireEvidenceProjectionLinkedCount(t, harness, login, incidentID, evidenceRecordID, 0)
	requireHTTPWorkbookPatch(t, harness, login, timelineRecordID, map[string]any{
		"view_schema_id":   "cartulary.view.timeline.v2",
		"base_row_version": timelineRowVersion,
		"client_txn_id":    "txn-evidence_lifecycle-i-01-link-evidence",
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
	// Requested Evidence is linked but does not count as attached until the blob
	// becomes available.
	requireTimelineEvidenceProjection(t, harness, login, incidentID, timelineRecordID, 0, false)
	if got := countAttachedEvidenceLinks(t, harness, incidentID, timelineRecordID, evidenceRecordID); got != 1 {
		t.Fatalf("workbook patch wrote attached evidence links: got %d want 1", got)
	}
	requireEvidenceProjectionLinkedCount(t, harness, login, incidentID, evidenceRecordID, 1)

	secondTimelineData := requireHTTPWorkbookCreate(t, harness, login, incidentID, "cartulary.view.timeline.v2", map[string]any{
		"client_txn_id":                   "txn-evidence_lifecycle-i-01-second-timeline",
		"timeline.activity_synopsis_text": "Second Evidence link",
	})
	secondTimelineRow := secondTimelineData["row"].(map[string]any)
	secondTimelineRecordID := appsupport.MustUUID(t, secondTimelineRow["record_id"].(string))
	requireHTTPWorkbookPatch(t, harness, login, secondTimelineRecordID, map[string]any{
		"view_schema_id":   "cartulary.view.timeline.v2",
		"base_row_version": 1,
		"client_txn_id":    "txn-evidence_lifecycle-i-01-second-link-evidence",
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
	requireEvidenceProjectionLinkedCount(t, harness, login, incidentID, evidenceRecordID, 2)
	requireHTTPWorkbookPatch(t, harness, login, secondTimelineRecordID, map[string]any{
		"view_schema_id":   "cartulary.view.timeline.v2",
		"base_row_version": 2,
		"client_txn_id":    "txn-evidence_lifecycle-i-01-remove-second-link",
		"changes": []map[string]any{{
			"field_key": "timeline.attached_evidence_ids",
			"action_payload": map[string]any{
				"kind": "collection_actions_v1",
				"actions": []map[string]any{{
					"op":       "remove_record_ref",
					"item_ref": "record_ref:" + evidenceRecordID.String(),
				}},
			},
		}},
	})
	requireEvidenceProjectionLinkedCount(t, harness, login, incidentID, evidenceRecordID, 1)

	payload := []byte("evidence_lifecycle projection object")
	sum := sha256.Sum256(payload)
	createResp := appsupport.DoJSON(t, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/object-blobs", map[string]any{
		"incident_id":       incidentID.String(),
		"client_txn_id":     "txn-evidence_lifecycle-i-01-blob",
		"byte_size":         len(payload),
		"filename_hint":     " endpoint.txt ",
		"content_type_hint": "text/plain",
		"sha256_hex":        fmt.Sprintf("%x", sum[:]),
	}, authOptions(login)...)
	createData := httptestx.RequireSuccessEnvelope(t, createResp, http.StatusCreated)["data"].(map[string]any)
	uploadTarget := createData["upload_target"].(map[string]any)
	href, _ := uploadTarget["href"].(string)
	if !strings.HasPrefix(href, "/api/v1/object-uploads/upl_") || strings.Contains(href, "://") {
		t.Fatalf("create returned non-opaque same-origin upload target: %#v", uploadTarget)
	}
	putObject(t, harness.Server.HTTP.URL, href, payload, "text/plain", login)

	attachBody := map[string]any{
		"object_blob_id":   createData["object_blob_id"],
		"base_row_version": 1,
		"client_txn_id":    "txn-evidence_lifecycle-i-01-attach",
	}
	attachResp := appsupport.DoJSON(t, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/evidence-records/"+evidenceRecordID.String()+"/attach-blob", attachBody, authOptions(login)...)
	attachData := httptestx.RequireSuccessEnvelope(t, attachResp, http.StatusOK)["data"].(map[string]any)
	if attachData["object_blob_id"] != createData["object_blob_id"] {
		t.Fatalf("attach object_blob_id got %#v want %#v", attachData["object_blob_id"], createData["object_blob_id"])
	}
	attachRow := attachData["row"].(map[string]any)
	attachCells := attachRow["cells"].(map[string]any)
	if got := int(attachCells["evidence.linked_record_count"].(map[string]any)["value"].(float64)); got != 1 {
		t.Fatalf("attach row evidence.linked_record_count got %d want 1: %#v", got, attachRow)
	}
	queriedRow := requireEvidenceProjectionRow(t, harness, login, incidentID, evidenceRecordID)
	if !reflect.DeepEqual(attachRow, queriedRow) {
		t.Fatalf("canonical mutation row differs from provider query row:\nmutation=%#v\nquery=%#v", attachRow, queriedRow)
	}
	revisionsupport.RequireOneRecordChangeIntentPerRevisionSQL(t, harness.DB, attachData["change_set_id"].(string))
	requireTimelineEvidenceProjection(t, harness, login, incidentID, timelineRecordID, 0, false)

	replayResp := appsupport.DoJSON(t, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/evidence-records/"+evidenceRecordID.String()+"/attach-blob", attachBody, authOptions(login)...)
	replayData := httptestx.RequireSuccessEnvelope(t, replayResp, http.StatusOK)["data"].(map[string]any)
	if replayData["change_set_id"] != attachData["change_set_id"] {
		t.Fatalf("attach replay changed change_set_id: replay=%#v first=%#v", replayData["change_set_id"], attachData["change_set_id"])
	}
	revisionsupport.RequireOneRecordChangeIntentPerRevisionSQL(t, harness.DB, attachData["change_set_id"].(string))
	requireTimelineEvidenceProjection(t, harness, login, incidentID, timelineRecordID, 0, false)
	availableData := requireHTTPWorkbookPatch(t, harness, login, evidenceRecordID, map[string]any{
		"view_schema_id":   "cartulary.view.evidence.v1",
		"base_row_version": 2,
		"client_txn_id":    "txn-evidence_lifecycle-i-01-available",
		"changes": []map[string]any{{
			"field_key": "evidence.lifecycle_state",
			"value":     "available",
		}},
	})
	revisionsupport.RequireOneRecordChangeIntentPerRevisionSQL(t, harness.DB, availableData["change_set_id"].(string))
	requireTimelineEvidenceProjection(t, harness, login, incidentID, timelineRecordID, 1, true)
	if got := countEvidenceRevisions(t, harness, evidenceRecordID); got != 3 {
		t.Fatalf("attach replay or lifecycle transition produced wrong revision count: got %d want 3", got)
	}
	if got := countEvidenceBlobLinks(t, harness, evidenceRecordID); got != 1 {
		t.Fatalf("evidence row has duplicate or missing blob link: got %d want 1", got)
	}
}

func TestAttachedEvidenceProjectionRebuild_Integration(t *testing.T) {
	harness := appsupport.StartServer(t, "evidence_lifecycle-projection-rebuild")
	login, _ := appsupport.ProvisionBootstrapAdmin(t, harness.Server)
	incident := appsupport.CreateIncident(t, harness.Server, login, map[string]any{
		"client_txn_id": "txn-evidence_lifecycle-projection-incident",
		"incident_key":  "evidence_lifecycle-projection",
		"title":         "Evidence projection rebuild",
	})
	incidentID := appsupport.MustUUID(t, incident["incident_id"].(string))

	timelineData := requireHTTPWorkbookCreate(t, harness, login, incidentID, "cartulary.view.timeline.v2", map[string]any{
		"client_txn_id":                   "txn-evidence_lifecycle-projection-timeline",
		"timeline.activity_synopsis_text": "Projection rebuild row",
	})
	timelineRow := timelineData["row"].(map[string]any)
	timelineRecordID := appsupport.MustUUID(t, timelineRow["record_id"].(string))
	timelineRowVersion := int(timelineRow["row_version"].(float64))

	evidenceData := requireHTTPWorkbookCreate(t, harness, login, incidentID, "cartulary.view.evidence.v1", map[string]any{
		"client_txn_id":  "txn-evidence_lifecycle-projection-evidence",
		"evidence.title": "Projection evidence",
	})
	evidenceRecordID := appsupport.MustUUID(t, evidenceData["row"].(map[string]any)["record_id"].(string))
	attachUploadedBlobWithMetadata(t, harness, login, incidentID, evidenceRecordID, []byte("evidence_lifecycle projection rebuild"), "projection.txt", "text/plain", "txn-evidence_lifecycle-projection-blob", "txn-evidence_lifecycle-projection-attach")
	var (
		evidenceLifecycleState string
		evidenceUploadState    string
		blobUploadState        string
	)
	if err := harness.DB.QueryRowContext(context.Background(), `
SELECT e.lifecycle_state, e.upload_state, b.upload_state
  FROM evidence e
  JOIN object_blobs b ON b.object_blob_id = e.object_blob_id
 WHERE e.record_id = $1
`, evidenceRecordID).Scan(&evidenceLifecycleState, &evidenceUploadState, &blobUploadState); err != nil {
		t.Fatalf("load attached evidence availability state: %v", err)
	}
	if evidenceLifecycleState != "available" || evidenceUploadState != "available" || blobUploadState != "available" {
		t.Fatalf(
			"attached evidence availability state got lifecycle=%q evidence_upload=%q blob_upload=%q",
			evidenceLifecycleState,
			evidenceUploadState,
			blobUploadState,
		)
	}

	requireHTTPWorkbookPatch(t, harness, login, timelineRecordID, map[string]any{
		"view_schema_id":   "cartulary.view.timeline.v2",
		"base_row_version": timelineRowVersion,
		"client_txn_id":    "txn-evidence_lifecycle-projection-link",
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

	sourceBefore := ProcessCounts(t, harness, incidentID, timelineRecordID)
	if _, err := harness.DB.ExecContext(context.Background(), `
UPDATE timeline_grid_projection
   SET evidence_count = 0,
       has_evidence = false
 WHERE record_id = $1
`, timelineRecordID); err != nil {
		t.Fatalf("corrupt timeline projection: %v", err)
	}
	requireTimelineProjectionStorage(t, harness, timelineRecordID, 0, false)
	requireTimelineEvidenceProjection(t, harness, login, incidentID, timelineRecordID, 0, false)

	if err := harness.Projections.RebuildIncident(context.Background(), incidentID); err != nil {
		t.Fatalf("rebuild timeline projection: %v", err)
	}
	requireTimelineProjectionStorage(t, harness, timelineRecordID, 1, true)
	requireTimelineEvidenceProjection(t, harness, login, incidentID, timelineRecordID, 1, true)
	if sourceAfter := ProcessCounts(t, harness, incidentID, timelineRecordID); sourceAfter != sourceBefore {
		t.Fatalf("projection rebuild mutated source/history state: before=%+v after=%+v", sourceBefore, sourceAfter)
	}
	if got := countEvidenceBlobLinks(t, harness, evidenceRecordID); got != 1 {
		t.Fatalf("projection rebuild changed evidence blob link count: got %d want 1", got)
	}
}
