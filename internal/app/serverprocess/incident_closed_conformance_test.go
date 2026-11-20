package serverprocess

import (
	"net/http"
	"slices"
	"testing"

	importtest "github.com/JochiRaider/cartulary/internal/modules/imports/testsupport"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/JochiRaider/cartulary/internal/testutil/workbookscenariotest"
)

// incidentClosedCases holds one case per operation that declares 409
// incident_closed. Each case prepares, on the open incident, requests that
// would succeed while the incident is open.
var incidentClosedCases = map[string]func(*conformanceEnv) []conformanceRequest{
	"resolveEntityMention": func(e *conformanceEnv) []conformanceRequest {
		_, mentionID, hostID := e.unresolvedMention("closed")
		return []conformanceRequest{e.mentionResolution(mentionID, hostID)}
	},
	"attachBlobToEvidenceRecord": func(e *conformanceEnv) []conformanceRequest {
		return []conformanceRequest{e.attach(e.evidence("Closed attach target"), e.verifiedBlob("closed-attach"))}
	},
	"applyImportSession": func(e *conformanceEnv) []conformanceRequest {
		sessionID, _ := importtest.PrepareReadyTimelineImport(e.t, e.harness.Server.HTTP.URL, e.login, e.incidentID.String(), e.txn("ready-import"))
		return []conformanceRequest{{
			label: "ready Timeline import", params: map[string]string{"import_session_id": sessionID},
			body: map[string]any{"client_txn_id": e.txn("apply")},
		}}
	},
	"createImportUnitRegion": func(e *conformanceEnv) []conformanceRequest {
		sessionID, unitIDs := importtest.StartXLSXSession(e.t, e.harness.Server.HTTP.URL, e.login, e.incidentID.String(), e.txn("xlsx-upload"), "closed-regions.xlsx")
		return []conformanceRequest{{
			label: "XLSX operator region", params: map[string]string{"import_session_id": sessionID, "base_unit_id": unitIDs[0]},
			body: map[string]any{"client_txn_id": e.txn("region"), "source_rect": map[string]any{"start_row": 1, "start_column": 1, "end_row": 2, "end_column": 2}},
		}}
	},
	"patchIncident": func(e *conformanceEnv) []conformanceRequest {
		return []conformanceRequest{{
			label: "current phase", params: map[string]string{"incident_id": e.incidentID.String()},
			// The version is read when sent, after the closure advanced it.
			body: func(t *testing.T) any {
				return map[string]any{"base_incident_version": e.incidentVersion(t), "current_phase": "Conformance phase"}
			},
		}}
	},
	"putTimelineTimeConversionProfile": func(e *conformanceEnv) []conformanceRequest {
		return []conformanceRequest{{
			label: "local offset", params: map[string]string{"incident_id": e.incidentID.String()},
			body: map[string]any{"base_profile_version": 1, "enabled": true, "local_offset_minutes": -300, "local_label": "UTC-05"},
		}}
	},
	"applyWorkbookBulkMutation": func(e *conformanceEnv) []conformanceRequest {
		recordID := e.timeline("Closed bulk target")
		return []conformanceRequest{{
			label: "Timeline fill-down", params: map[string]string{"incident_id": e.incidentID.String(), "view_schema_id": conformanceTimelineView},
			body: map[string]any{
				"view_schema_id": conformanceTimelineView, "client_txn_id": e.txn("bulk"), "kind": "fill_down_v1",
				"field_key": "timeline.raw_activity_text", "value": "Filled conformance text",
				"targets": []map[string]any{{"record_id": recordID.String(), "base_row_version": 1}},
			},
		}}
	},
	"pasteWorkbookClipboard": func(e *conformanceEnv) []conformanceRequest {
		return []conformanceRequest{{
			label: "Timeline create paste", params: map[string]string{"incident_id": e.incidentID.String(), "view_schema_id": conformanceTimelineView},
			body: map[string]any{
				"view_schema_id": conformanceTimelineView, "client_txn_id": e.txn("paste"),
				"clipboard_text": "pasted conformance text", "format": "csv", "header_mode": "none",
				"start_field_key": "timeline.raw_activity_text", "columns": []string{"timeline.raw_activity_text"},
				"targets": []map[string]any{{"kind": "create"}},
			},
		}}
	},
	// Every create-capable view has its own create owner, so each is a request.
	"createViewRow": func(e *conformanceEnv) []conformanceRequest {
		payloads := e.createViewPayloads(e.host("closed-assessment-subject"), e.timeline("Closed assessment support"))
		views := createCapableViews()
		if got := sortedKeys(payloads); !slices.Equal(got, views) {
			e.t.Fatalf("createViewRow requests cover views %v, want every create-capable view %v", got, views)
		}
		requests := make([]conformanceRequest, 0, len(views))
		for _, view := range views {
			body := map[string]any{"client_txn_id": e.txn("create")}
			for key, value := range payloads[view] {
				body[key] = value
			}
			requests = append(requests, conformanceRequest{
				label: view, params: map[string]string{"incident_id": e.incidentID.String(), "view_schema_id": view}, body: body,
			})
		}
		return requests
	},
	"dismissIndicatorObservation": func(e *conformanceEnv) []conformanceRequest {
		observationID := e.observation(e.timeline("closed-dismiss.conformance.test"), "timeline.raw_activity_text", "closed-dismiss.conformance.test")
		return []conformanceRequest{{
			label: "unresolved observation", params: map[string]string{"observation_id": observationID.String()},
			body: map[string]any{"client_txn_id": e.txn("dismiss"), "base_row_version": 1},
		}}
	},
	"resolveIndicatorObservation": func(e *conformanceEnv) []conformanceRequest {
		observationID := e.observation(e.timeline("closed-resolve.conformance.test"), "timeline.raw_activity_text", "closed-resolve.conformance.test")
		indicatorID := e.indicator("closed-resolve-target.conformance.test")
		return []conformanceRequest{{
			label: "unresolved observation", params: map[string]string{"observation_id": observationID.String()},
			body: map[string]any{"client_txn_id": e.txn("resolve"), "base_row_version": 1, "resolved_indicator_record_id": indicatorID.String()},
		}}
	},
	"restoreIndicatorObservation": func(e *conformanceEnv) []conformanceRequest {
		observationID := e.observation(e.timeline("closed-restore.conformance.test"), "timeline.raw_activity_text", "closed-restore.conformance.test")
		e.require(http.MethodPost, "/api/v1/indicator-observations/"+observationID.String()+"/dismiss",
			map[string]any{"client_txn_id": e.txn("dismiss-before-restore"), "base_row_version": 1}, http.StatusOK)
		return []conformanceRequest{{
			label: "dismissed observation", params: map[string]string{"observation_id": observationID.String()},
			body: map[string]any{"client_txn_id": e.txn("restore"), "base_row_version": 2},
		}}
	},
	"appendIndicatorStateInterval": func(e *conformanceEnv) []conformanceRequest {
		return []conformanceRequest{e.stateInterval(e.indicator("closed-interval.conformance.test"), e.timeline("Closed interval support"))}
	},
	"createObjectBlobSlot": func(e *conformanceEnv) []conformanceRequest {
		return []conformanceRequest{{label: "text Blob slot", body: e.blobCreateBody("closed-slot", []byte("closed slot payload"))}}
	},
	"verifyObjectBlob": func(e *conformanceEnv) []conformanceRequest {
		slot := e.uploadedBlob("closed-verify")
		return []conformanceRequest{{
			label: "uploaded Blob", params: map[string]string{"object_blob_id": slot.blobID},
			body: map[string]any{"client_txn_id": e.txn("verify")},
		}}
	},
	"uploadObjectBlobContent": func(e *conformanceEnv) []conformanceRequest {
		return []conformanceRequest{e.blobSlot("closed-upload").upload()}
	},
	"mutateNoteAssociations": func(e *conformanceEnv) []conformanceRequest {
		return []conformanceRequest{e.noteAssociation(e.note("Closed association"), e.note("Closed counterpart"))}
	},
	"deleteRecord": func(e *conformanceEnv) []conformanceRequest {
		noteID := e.note("Closed delete target")
		return []conformanceRequest{{
			label: "live Note", params: map[string]string{"record_id": noteID.String()},
			body: map[string]any{"base_row_version": 1, "client_txn_id": e.txn("delete"), "reason": "Conformance deletion"},
		}}
	},
	"patchRecord": func(e *conformanceEnv) []conformanceRequest {
		return []conformanceRequest{e.notePatch(e.note("Closed patch target"))}
	},
	// keep_saved commits only a receipt; it is a distinct owner path from a
	// value resolution (REQ-01-590).
	"resolveRecordSameFieldConflict": func(e *conformanceEnv) []conformanceRequest {
		recordID, token := e.timelineConflict("Closed conflict")
		return e.conflictResolutions(recordID, token)
	},
	"createRecordLinkedNote": func(e *conformanceEnv) []conformanceRequest {
		return []conformanceRequest{e.linkedNote(e.host("closed-linked-source"))}
	},
	"markTimelineRecordReviewed": func(e *conformanceEnv) []conformanceRequest {
		return []conformanceRequest{e.timelineAction(e.timeline("Closed review target"), "mark-reviewed")}
	},
	"restoreRecord": func(e *conformanceEnv) []conformanceRequest {
		noteID := e.note("Closed restore target")
		e.softDelete(noteID)
		return []conformanceRequest{{
			label: "deleted Note", params: map[string]string{"record_id": noteID.String()},
			body: map[string]any{"base_row_version": e.version(noteID), "client_txn_id": e.txn("restore")},
		}}
	},
	"rollbackRecord": func(e *conformanceEnv) []conformanceRequest {
		noteID := e.note("Closed rollback target")
		httptestx.RequireSuccessData(e.t, e.patch(noteID, "cartulary.view.notes.v1", "note.title", "Closed rollback revised", 1), http.StatusOK)
		return []conformanceRequest{{
			label: "Note row restore", params: map[string]string{"record_id": noteID.String()},
			body: map[string]any{
				"base_row_version": 2, "client_txn_id": e.txn("rollback"),
				"target": map[string]any{"kind": "row_restore", "restore_to_revision_no": 1},
			},
		}}
	},
	"supersedeRecord": func(e *conformanceEnv) []conformanceRequest {
		return []conformanceRequest{e.timelineAction(e.timeline("Closed supersede target"), "supersede")}
	},
	"createManualIndicatorObservation": func(e *conformanceEnv) []conformanceRequest {
		text := "closed-manual.conformance.test"
		return []conformanceRequest{e.manualObservation(e.timeline(text), "timeline.raw_activity_text", text)}
	},
	"mergeEntityRecord": func(e *conformanceEnv) []conformanceRequest {
		survivorID, loserID := e.host("closed-merge-survivor"), e.host("closed-merge-loser")
		return []conformanceRequest{{
			label: "Host merge", params: map[string]string{"survivor_record_id": survivorID.String()},
			body: map[string]any{
				"loser_record_id": loserID.String(), "survivor_base_row_version": 1, "loser_base_row_version": 1,
				"client_txn_id": e.txn("merge"), "reason": "Conformance merge",
			},
		}}
	},
}

// incidentClosedOperations is every catalog operation declaring 409
// incident_closed (Core 04 AC-616).
func incidentClosedOperations(operations map[string]httpapi.ContractOperation) []string {
	var ids []string
	for id, operation := range operations {
		if declaresErrorCode(operation, http.StatusConflict, "incident_closed") {
			ids = append(ids, id)
		}
	}
	slices.Sort(ids)
	return ids
}

// TestIncidentClosedConformance verifies Core 04 AC-616's closed-incident
// family (REQ-01-590, REQ-01-591): every operation that declares 409
// incident_closed returns it for a fresh request once its incident is closed,
// and commits no source, history, projection, idempotency, object, job or
// Collaboration effect. The case table is keyed by the catalog-derived
// operation set, so an operation without a case fails by name.
func TestIncidentClosedConformance(t *testing.T) {
	operations := incidentClosedOperations(catalogOperationsByID(t))
	requireCaseKeySet(t, "incident_closed", sortedKeys(incidentClosedCases), operations)

	e := startConformanceEnv(t, "incident-closed-conformance", "IR-CLOSED-CONFORMANCE")
	prepared := make(map[string][]conformanceRequest, len(operations))
	for _, id := range operations {
		prepared[id] = incidentClosedCases[id](e)
		if len(prepared[id]) == 0 {
			t.Fatalf("%s case prepared no request", id)
		}
	}
	beforeClose := workbookscenariotest.SnapshotIncidentEffects(t, e.harness, e.incidentID)
	e.closeIncident()
	// The snapshot observes the closure itself, so the unchanged snapshots
	// below are evidence rather than a blind comparison.
	if changed := workbookscenariotest.ChangedIncidentEffects(beforeClose, workbookscenariotest.SnapshotIncidentEffects(t, e.harness, e.incidentID)); !slices.Contains(changed, "incident") {
		t.Fatalf("closure changed effect classes %v, want the incident class", changed)
	}
	e.quiesce()

	for _, id := range operations {
		operation := e.operations[id]
		t.Run(id, func(t *testing.T) {
			for _, request := range prepared[id] {
				t.Run(request.label, func(t *testing.T) {
					e.requireRejected(t, operation, request, http.StatusConflict, "incident_closed")
				})
			}
		})
	}
}
