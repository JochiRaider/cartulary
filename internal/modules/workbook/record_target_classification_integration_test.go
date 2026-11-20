package workbook_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
)

// TestWorkbookRecordTargetClassification_Integration verifies AC-614's first
// bullet through the public single-record routes: for every owner of a
// mutable record type, a patch, a conflict resolution (keep_saved and a value
// resolution) and, where the type has one, a lifecycle action on a soft-deleted
// record return 409 record_deleted_use_restore and commit nothing, and a
// missing record returns the route's not-found response (REQ-01-705). The
// deleted state precedes the record type, the incident lifecycle and the base
// version.
func TestWorkbookRecordTargetClassification_Integration(t *testing.T) {
	harness, login, _, incidentID := ConflictFixture(t, "workbook-record-target-classification", "IR-RECORD-TARGET-CLASSIFICATION")
	recordsURL := harness.Server.HTTP.URL + "/api/v1/records/"
	authOptions := []func(*http.Request){httptestx.WithCookies(login.Session, login.CSRF), httptestx.WithHeader(authn.CSRFHeaderName, login.CSRF.Value)}
	post := func(path string, body map[string]any) *http.Response {
		t.Helper()
		return httptestx.DoJSON(t, http.DefaultClient, http.MethodPost, recordsURL+path, body, authOptions...)
	}
	patch := func(recordID uuid.UUID, view, field string, base int64, txn string) *http.Response {
		t.Helper()
		return doWorkbookJSON(t, harness, login, http.MethodPatch, uuid.Nil, "", recordID, map[string]any{
			"view_schema_id": view, "base_row_version": base, "client_txn_id": txn,
			"changes": []map[string]any{{"field_key": field, "value": "Rejected " + txn}},
		})
	}
	resolve := func(recordID uuid.UUID, token, kind, txn string) *http.Response {
		t.Helper()
		body := map[string]any{"conflict_token": token, "resolution_kind": kind, "client_txn_id": txn}
		if kind != "keep_saved" {
			body["resolved_value"] = "Rejected " + txn
		}
		return ResolveConflictRaw(t, harness, login, recordID, token, body)
	}
	type effects struct{ changeSets, receipts, version int }
	snapshot := func(recordID uuid.UUID) effects {
		t.Helper()
		return effects{
			changeSets: appsupport.QueryCount(t, harness.DB, `SELECT count(*) FROM change_sets WHERE incident_id = $1`, incidentID),
			receipts:   appsupport.QueryCount(t, harness.DB, `SELECT count(*) FROM route_idempotency WHERE scope_key = $1`, recordID.String()),
			version:    appsupport.QueryCount(t, harness.DB, `SELECT COALESCE(max(row_version), 0) FROM records WHERE record_id = $1`, recordID),
		}
	}
	// requireRejected sends the request and requires the error envelope with no
	// change set, receipt or version change.
	requireRejected := func(t *testing.T, name string, recordID uuid.UUID, status int, code string, send func() *http.Response) {
		t.Helper()
		before := snapshot(recordID)
		response := send()
		if response.StatusCode != status {
			t.Fatalf("%s: status %d, want %d %s; body=%#v", name, response.StatusCode, status, code, httptestx.ReadJSONBody(t, response))
		}
		httptestx.RequireErrorEnvelope(t, response, status, code)
		if after := snapshot(recordID); after != before {
			t.Fatalf("%s committed effects: %+v -> %+v", name, before, after)
		}
	}

	type ownerCase struct {
		name, view, field string
		create            map[string]any
		lifecycle         []string
	}
	owners := []ownerCase{
		{name: "timeline", view: TimelineViewSchemaID, field: "timeline.activity_synopsis_text",
			create: map[string]any{"timeline.activity_synopsis_text": "Timeline base"}, lifecycle: []string{"mark-reviewed", "supersede"}},
		{name: "note", view: "cartulary.view.notes.v1", field: "note.title",
			create: map[string]any{"note.title": "Note base", "note.body": "Note body"}},
		{name: "lesson", view: "cartulary.view.lesson.v1", field: "lesson.summary",
			create: map[string]any{"lesson.summary": "Lesson base"}},
		{name: "evidence", view: "cartulary.view.evidence.v1", field: "evidence.title",
			create: map[string]any{"evidence.title": "Evidence base"}},
		{name: "host", view: "cartulary.view.hosts.v1", field: "host.display_name",
			create: map[string]any{"host.display_name": "Host base", "host.hostname": "classification-host"}},
		{name: "identity", view: "cartulary.view.identities.v1", field: "identity.display_name",
			create: map[string]any{"identity.display_name": "Identity base", "identity.email": "classification@example.test"}},
		{name: "party", view: "cartulary.view.parties.v1", field: "party.display_name",
			create: map[string]any{"party.display_name": "Party base", "party.party_kind": "person"}},
		{name: "task", view: "cartulary.view.task_requests.v1", field: "task.title",
			create: map[string]any{"task.title": "Task base", "task.task_kind": "question"}},
		{name: "decision", view: "cartulary.view.decisions.v1", field: "decision.summary",
			create:    map[string]any{"decision.summary": "Decision base", "decision.decision_type": "scope", "decision.rationale": "Classification fixture"},
			lifecycle: []string{"supersede"}},
	}
	type deletedTarget struct {
		owner    ownerCase
		recordID uuid.UUID
		token    string
	}
	deletedTargets := make([]deletedTarget, 0, len(owners))
	for _, owner := range owners {
		t.Run(owner.name+"/deleted", func(t *testing.T) {
			prefix := "rtc-" + owner.name
			body := map[string]any{"client_txn_id": prefix + "-create"}
			for key, value := range owner.create {
				body[key] = value
			}
			created := requireWorkbookCreate(t, harness, login, incidentID, owner.view, body)["row"].(map[string]any)
			recordID := appsupport.MustUUID(t, created["record_id"].(string))
			requireWorkbookPatch(t, harness, login, recordID, map[string]any{
				"view_schema_id": owner.view, "base_row_version": 1, "client_txn_id": prefix + "-saved",
				"changes": []map[string]any{{"field_key": owner.field, "value": "Saved " + owner.name}},
			})
			stale := patch(recordID, owner.view, owner.field, 1, prefix+"-stale")
			conflict := httptestx.RequireErrorEnvelope(t, stale, http.StatusConflict, "same_field_conflict")["error"].(map[string]any)["conflict"].(map[string]any)
			token := conflict["conflict_token"].(string)
			httptestx.RequireSuccessEnvelope(t, deleteRecordViaWorkbookRoute(t, harness, login, recordID, map[string]any{
				"base_row_version": 2, "client_txn_id": prefix + "-delete",
			}), http.StatusOK)

			requireRejected(t, "patch", recordID, http.StatusConflict, "record_deleted_use_restore", func() *http.Response { return patch(recordID, owner.view, owner.field, 3, prefix+"-patch") })
			requireRejected(t, "keep_saved", recordID, http.StatusConflict, "record_deleted_use_restore", func() *http.Response { return resolve(recordID, token, "keep_saved", prefix+"-keep") })
			requireRejected(t, "use_unsaved", recordID, http.StatusConflict, "record_deleted_use_restore", func() *http.Response { return resolve(recordID, token, "use_unsaved", prefix+"-use") })
			for _, action := range owner.lifecycle {
				body := map[string]any{"base_row_version": 3, "client_txn_id": prefix + "-" + action, "reason": "Classification fixture"}
				if owner.name == "decision" {
					body["replacement_record_id"] = uuid.NewString()
				}
				requireRejected(t, action, recordID, http.StatusConflict, "record_deleted_use_restore", func() *http.Response { return post(recordID.String()+"/"+action, body) })
			}
			if owner.name == "note" {
				requireRejected(t, "note associations", recordID, http.StatusConflict, "record_deleted_use_restore", func() *http.Response {
					return post(recordID.String()+"/note-associations", map[string]any{
						"kind": "source", "base_row_version": 3, "client_txn_id": prefix + "-associations",
						"actions": []map[string]any{{"op": "add", "counterpart_record_id": uuid.NewString()}},
					})
				})
			}
			deletedTargets = append(deletedTargets, deletedTarget{owner: owner, recordID: recordID, token: token})
		})
	}

	t.Run("deleted record of a type without a patch owner", func(t *testing.T) {
		indicator := requireWorkbookCreate(t, harness, login, incidentID, "cartulary.view.indicators.v1", map[string]any{
			"client_txn_id": "rtc-indicator-create", "indicator.indicator_type": "domain_name",
			"indicator.value_kind": "atomic", "indicator.display_value": "classification.example.test",
		})["row"].(map[string]any)
		indicatorID := appsupport.MustUUID(t, indicator["record_id"].(string))
		requireRejected(t, "live indicator patch", indicatorID, http.StatusNotFound, "incident_not_found", func() *http.Response {
			return patch(indicatorID, "cartulary.view.indicators.v1", "indicator.display_value", 1, "rtc-indicator-live")
		})
		httptestx.RequireSuccessEnvelope(t, deleteRecordViaWorkbookRoute(t, harness, login, indicatorID, map[string]any{
			"base_row_version": 1, "client_txn_id": "rtc-indicator-delete",
		}), http.StatusOK)
		requireRejected(t, "deleted indicator patch", indicatorID, http.StatusConflict, "record_deleted_use_restore", func() *http.Response {
			return patch(indicatorID, "cartulary.view.indicators.v1", "indicator.display_value", 2, "rtc-indicator-deleted")
		})
	})

	t.Run("missing record is not found on every route", func(t *testing.T) {
		missingID := uuid.New()
		requireRejected(t, "missing patch", missingID, http.StatusNotFound, "incident_not_found", func() *http.Response {
			return patch(missingID, TimelineViewSchemaID, "timeline.activity_synopsis_text", 1, "rtc-missing-patch")
		})
		requireRejected(t, "missing conflict resolution", missingID, http.StatusNotFound, "incident_not_found", func() *http.Response { return resolve(missingID, "missing-token", "keep_saved", "rtc-missing-keep") })
		for _, action := range []string{"mark-reviewed", "supersede"} {
			requireRejected(t, "missing "+action, missingID, http.StatusNotFound, "incident_not_found", func() *http.Response {
				return post(missingID.String()+"/"+action, map[string]any{
					"base_row_version": 1, "client_txn_id": "rtc-missing-" + action, "reason": "Classification fixture",
				})
			})
		}
		requireRejected(t, "missing note associations", missingID, http.StatusNotFound, "incident_not_found", func() *http.Response {
			return post(missingID.String()+"/note-associations", map[string]any{
				"kind": "source", "base_row_version": 1, "client_txn_id": "rtc-missing-associations",
				"actions": []map[string]any{{"op": "add", "counterpart_record_id": uuid.NewString()}},
			})
		})
	})

	if len(deletedTargets) != len(owners) {
		t.Fatalf("deleted fixtures = %d, want %d", len(deletedTargets), len(owners))
	}
	t.Run("deleted precedes the incident lifecycle and a stale version", func(t *testing.T) {
		live := CreateNote(t, harness, login, incidentID, "rtc-closed-live", "Live before close", "Live body")
		liveID := appsupport.MustUUID(t, live["record_id"].(string))
		if _, err := harness.DB.ExecContext(context.Background(), `UPDATE incidents SET status = 'closed', closed_at = now() WHERE id = $1`, incidentID); err != nil {
			t.Fatalf("close incident: %v", err)
		}
		requireRejected(t, "live patch after close", liveID, http.StatusConflict, "incident_closed", func() *http.Response {
			return patch(liveID, "cartulary.view.notes.v1", "note.title", 1, "rtc-closed-live-patch")
		})
		for _, target := range deletedTargets {
			prefix := "rtc-closed-" + target.owner.name
			requireRejected(t, target.owner.name+" stale patch after close", target.recordID, http.StatusConflict, "record_deleted_use_restore", func() *http.Response {
				return patch(target.recordID, target.owner.view, target.owner.field, 1, prefix+"-patch")
			})
			requireRejected(t, target.owner.name+" keep_saved after close", target.recordID, http.StatusConflict, "record_deleted_use_restore", func() *http.Response { return resolve(target.recordID, target.token, "keep_saved", prefix+"-keep") })
		}
	})
}
