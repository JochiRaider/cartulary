package workbook_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/collaboration/testsupport/incidentwstest"
	recordstoretest "github.com/JochiRaider/cartulary/internal/modules/records/testsupport/storetest"
	"github.com/JochiRaider/cartulary/internal/modules/savedviews"
	"github.com/JochiRaider/cartulary/internal/modules/workbook"
	workbookstartup "github.com/JochiRaider/cartulary/internal/modules/workbook/startup"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
	platformws "github.com/JochiRaider/cartulary/internal/platform/ws"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
)

func TestPhase9Sprint7_CoordinationMinimumDefaultsAndRejection_U_9_08(t *testing.T) {
	ctx := context.Background()
	harness := recordstoretest.StartStore(t, "phase9-sprint7-coordination-defaults")
	store := workbook.NewStore(harness.DB)
	actor := recordstoretest.SeedLocalUserFlags(t, harness.DB, "sprint7-coordination@example.test", "Sprint7 Coordination", "Sprint7Coordination1!", false, false, true)
	incident := recordstoretest.CreateIncidentInStore(t, harness.DB, actor, "txn-phase9-sprint7-coordination-incident", "IR-S7-COORD", "Phase 9 Sprint 7 coordination defaults")

	for _, tc := range []struct {
		name         string
		viewSchemaID string
		values       map[string]workbook.ValueChange
		wantField    string
	}{
		{
			name:         "comm-log-missing-summary",
			viewSchemaID: workbook.CommLogViewSchemaID,
			values: map[string]workbook.ValueChange{
				"comm_log.comm_type":          sprint7Text("briefing"),
				"comm_log.audience":           sprint7Text("leadership"),
				"comm_log.channel_or_meeting": sprint7Text("Bridge"),
			},
			wantField: "comm_log.summary",
		},
		{
			name:         "handoff-missing-summary",
			viewSchemaID: workbook.HandoffViewSchemaID,
			values: map[string]workbook.ValueChange{
				"handoff.incoming_owner_user_id": sprint7UUID(actor.ID),
			},
			wantField: "handoff.current_state_summary",
		},
		{
			name:         "status-review-missing-summary",
			viewSchemaID: workbook.StatusReviewViewSchemaID,
			values:       map[string]workbook.ValueChange{},
			wantField:    "status_review.current_state_summary",
		},
		{
			name:         "lesson-missing-summary",
			viewSchemaID: workbook.LessonViewSchemaID,
			values:       map[string]workbook.ValueChange{},
			wantField:    "lesson.summary",
		},
	} {
		before := countSprint7DurableState(t, harness, incident.ID)
		_, err := store.CreateWorkbookRow(ctx, actor, incident.ID, workbook.CreateRequest{
			ViewSchemaID: tc.viewSchemaID,
			ClientTxnID:  "txn-phase9-sprint7-" + tc.name,
			Values:       tc.values,
		}, []byte("txn-phase9-sprint7-"+tc.name), "req-phase9-sprint7-"+tc.name, sprint7Time(0))
		requireSprint7MutationValidation(t, err, tc.wantField, "missing_required_field")
		requireSprint7DurableState(t, harness, incident.ID, before, tc.name)
	}

	expectSprint7DecodePatchRejected(t, workbook.CommLogViewSchemaID, "comm_log.comm_id", "client-supplied")
	expectSprint7DecodePatchRejected(t, workbook.HandoffViewSchemaID, "handoff.handoff_id", "client-supplied")
	expectSprint7DecodePatchRejected(t, workbook.StatusReviewViewSchemaID, "status_review.status_review_id", "client-supplied")
	expectSprint7DecodePatchRejected(t, workbook.LessonViewSchemaID, "lesson.lesson_id", "client-supplied")

	comm := mustCreateSprint7Row(t, store, actor, incident.ID, workbook.CommLogViewSchemaID, "txn-phase9-sprint7-comm-defaults", map[string]workbook.ValueChange{
		"comm_log.comm_type":          sprint7Text("briefing"),
		"comm_log.audience":           sprint7Text("Leadership Team"),
		"comm_log.channel_or_meeting": sprint7Text("Bridge"),
		"comm_log.summary":            sprint7Text("Daily coordination update"),
	}, nil, sprint7Time(0))
	commRow := comm.Payload["row"].(map[string]any)
	requireSprint7ArtifactType(t, harness, comm.RecordID, "comm_log")
	requireSprint7CellNonEmpty(t, commRow, "comm_log.comm_id")
	requireSprint7CellNonEmpty(t, commRow, "comm_log.timestamp_utc")
	requireSprint7CellValue(t, commRow, "comm_log.next_report_at", nil)
	requireSprint7CellValue(t, commRow, "comm_log.privilege_tag", nil)
	requireSprint7CollectionItemCount(t, commRow, "comm_log.decision_ids", 0)
	requireSprint7CollectionItemCount(t, commRow, "comm_log.action_task_ids", 0)
	requireSprint7CollectionItemCount(t, commRow, "comm_log.audience_party_ids", 0)
	requireSprint7CollectionItemCount(t, commRow, "comm_log.attendee_party_ids", 0)
	beforeCommVersion := sprint6RecordVersion(t, harness.DB, comm.RecordID)
	_, err := sprint6Patch(store, actor, comm.RecordID, workbook.CommLogViewSchemaID, beforeCommVersion, "txn-phase9-sprint7-comm-invalid-type",
		sprint6ValueChange("comm_log.comm_type", workbook.ValueChange{Kind: "text", Text: stringPtrU911("emergency")}))
	requireSprint7MutationValidation(t, err, "comm_log.comm_type", "invalid_value")
	if got := sprint6RecordVersion(t, harness.DB, comm.RecordID); got != beforeCommVersion {
		t.Fatalf("invalid comm type changed row version: got %d want %d", got, beforeCommVersion)
	}

	handoff := mustCreateSprint7Row(t, store, actor, incident.ID, workbook.HandoffViewSchemaID, "txn-phase9-sprint7-handoff-defaults", map[string]workbook.ValueChange{
		"handoff.incoming_owner_user_id": sprint7UUID(actor.ID),
		"handoff.current_state_summary":  sprint7Text("Night shift owns containment"),
	}, nil, sprint7Time(time.Hour))
	handoffRow := handoff.Payload["row"].(map[string]any)
	requireSprint7ArtifactType(t, harness, handoff.RecordID, "handoff")
	requireSprint7CellNonEmpty(t, handoffRow, "handoff.handoff_id")
	requireSprint7CellNonEmpty(t, handoffRow, "handoff.timestamp_utc")
	requireSprint7CellValue(t, handoffRow, "handoff.outgoing_owner_user_id", actor.ID.String())
	requireSprint7CellValue(t, handoffRow, "handoff.acknowledged_at", nil)
	requireSprint7CellValue(t, handoffRow, "handoff.ack_state", "pending")
	requireSprint7CellValue(t, handoffRow, "handoff.next_checks", nil)
	requireSprint7CollectionItemCount(t, handoffRow, "handoff.open_task_ids", 0)
	requireSprint7CollectionItemCount(t, handoffRow, "handoff.open_decision_ids", 0)
	requireSprint7CollectionItemCount(t, handoffRow, "handoff.open_risk_refs", 0)
	acknowledgedAt := sprint7Time(2 * time.Hour)
	handoffAck := mustSprint6Patch(t, store, actor, handoff.RecordID, workbook.HandoffViewSchemaID, 1, "txn-phase9-sprint7-handoff-ack",
		sprint6ValueChange("handoff.acknowledged_at", workbook.ValueChange{Kind: "timestamp", Timestamp: &acknowledgedAt}))
	requireSprint7CellValue(t, handoffAck.Payload["row"].(map[string]any), "handoff.ack_state", "acknowledged")
	handoffClear := mustSprint6Patch(t, store, actor, handoff.RecordID, workbook.HandoffViewSchemaID, 2, "txn-phase9-sprint7-handoff-clear-ack",
		sprint6ValueChange("handoff.acknowledged_at", workbook.ValueChange{Kind: "null"}))
	requireSprint7CellValue(t, handoffClear.Payload["row"].(map[string]any), "handoff.ack_state", "pending")

	status := mustCreateSprint7Row(t, store, actor, incident.ID, workbook.StatusReviewViewSchemaID, "txn-phase9-sprint7-status-defaults", map[string]workbook.ValueChange{
		"status_review.current_state_summary": sprint7Text("Containment is stable"),
	}, nil, sprint7Time(2*time.Hour))
	statusRow := status.Payload["row"].(map[string]any)
	requireSprint7ArtifactType(t, harness, status.RecordID, "status_review")
	requireSprint7CellNonEmpty(t, statusRow, "status_review.status_review_id")
	requireSprint7CellNonEmpty(t, statusRow, "status_review.timestamp_utc")
	requireSprint7CellValue(t, statusRow, "status_review.review_owner_user_id", actor.ID.String())
	requireSprint7CellValue(t, statusRow, "status_review.next_report_at", nil)
	requireSprint7CellValue(t, statusRow, "status_review.active_risks_summary", nil)
	requireSprint7CollectionItemCount(t, statusRow, "status_review.blocked_task_ids", 0)
	requireSprint7CollectionItemCount(t, statusRow, "status_review.pending_evidence_ids", 0)
	requireSprint7CollectionItemCount(t, statusRow, "status_review.open_decision_ids", 0)

	lesson := mustCreateSprint7Row(t, store, actor, incident.ID, workbook.LessonViewSchemaID, "txn-phase9-sprint7-lesson-defaults", map[string]workbook.ValueChange{
		"lesson.summary": sprint7Text("Preserve VPN logs earlier"),
	}, nil, sprint7Time(3*time.Hour))
	lessonRow := lesson.Payload["row"].(map[string]any)
	requireSprint7ArtifactType(t, harness, lesson.RecordID, "lesson")
	requireSprint7CellNonEmpty(t, lessonRow, "lesson.lesson_id")
	requireSprint7CellNonEmpty(t, lessonRow, "lesson.timestamp_utc")
	requireSprint7CellValue(t, lessonRow, "lesson.owner_user_id", actor.ID.String())
	requireSprint7CellValue(t, lessonRow, "lesson.closure_state", "open")
	requireSprint7CollectionItemCount(t, lessonRow, "lesson.follow_up_task_ids", 0)
	requireSprint7CollectionItemCount(t, lessonRow, "lesson.evidence_refs", 0)
	beforeLessonVersion := sprint6RecordVersion(t, harness.DB, lesson.RecordID)
	_, err = sprint6Patch(store, actor, lesson.RecordID, workbook.LessonViewSchemaID, beforeLessonVersion, "txn-phase9-sprint7-lesson-invalid-closure",
		sprint6ValueChange("lesson.closure_state", workbook.ValueChange{Kind: "text", Text: stringPtrU911("archived")}))
	requireSprint7MutationValidation(t, err, "lesson.closure_state", "invalid_value")
	if got := sprint6RecordVersion(t, harness.DB, lesson.RecordID); got != beforeLessonVersion {
		t.Fatalf("invalid lesson closure state changed row version: got %d want %d", got, beforeLessonVersion)
	}
}

func TestPhase9Sprint7_RejectedCoordinationCreateEmitsNoRecordChanged_U_9_08(t *testing.T) {
	harness := recordstoretest.StartServer(t, "phase9-sprint7-rejected-create-ws")
	login, _ := recordstoretest.ProvisionBootstrapAdmin(t, harness.Server)
	incident := recordstoretest.CreateIncident(t, harness.Server, login, map[string]any{
		"client_txn_id": "txn-phase9-sprint7-rejected-ws-incident",
		"incident_key":  "IR-S7-REJECTED-WS",
		"title":         "Phase 9 Sprint 7 rejected create websocket",
	})
	incidentID := incident["incident_id"].(string)
	socket := incidentwstest.ConnectAndHello(t, harness.Server.HTTP.URL, incidentID, incidentwstest.ConnectOptions{
		Cookies:          []*http.Cookie{login.SessionCookie},
		ClientInstanceID: "phase9-sprint7-rejected-create-listener",
		Presence: platformws.PresenceInput{
			SheetRef: map[string]string{"kind": "view_schema", "id": workbook.CommLogViewSchemaID},
			Mode:     "viewing",
		},
	})
	defer socket.Close(websocket.StatusNormalClosure, "test_complete")

	for _, tc := range []struct {
		viewSchemaID string
		artifactType string
		clientTxnID  string
		wantField    string
	}{
		{viewSchemaID: workbook.CommLogViewSchemaID, artifactType: "comm_log", clientTxnID: "txn-phase9-sprint7-zero-field-comm", wantField: "comm_log.comm_type"},
		{viewSchemaID: workbook.HandoffViewSchemaID, artifactType: "handoff", clientTxnID: "txn-phase9-sprint7-zero-field-handoff", wantField: "handoff.incoming_owner_user_id"},
		{viewSchemaID: workbook.StatusReviewViewSchemaID, artifactType: "status_review", clientTxnID: "txn-phase9-sprint7-zero-field-status", wantField: "status_review.current_state_summary"},
		{viewSchemaID: workbook.LessonViewSchemaID, artifactType: "lesson", clientTxnID: "txn-phase9-sprint7-zero-field-lesson", wantField: "lesson.summary"},
	} {
		t.Run(tc.viewSchemaID, func(t *testing.T) {
			before := countSprint7SQLDurableState(t, harness.DB, incidentID)
			resp := recordstoretest.DoJSON(
				t,
				http.MethodPost,
				harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/views/"+tc.viewSchemaID+"/rows",
				map[string]any{"client_txn_id": tc.clientTxnID},
				recordstoretest.WithCookies(login.SessionCookie, login.CSRFCookie),
				recordstoretest.WithHeader(authn.CSRFHeaderName, login.CSRFCookie.Value),
			)
			body := httptestx.RequireErrorEnvelope(t, resp, http.StatusBadRequest, "invalid_mutation_payload")
			httptestx.RequireErrorDetail(t, body, "field", tc.wantField)
			requireSprint7SQLDurableState(t, harness.DB, incidentID, before, "rejected zero-field create route")
			if got := countSprint7SQLProjectionRows(t, harness.DB, incidentID, tc.artifactType); got != 0 {
				t.Fatalf("%s rejected zero-field create left %d projection rows", tc.viewSchemaID, got)
			}
			requireSprint7NoRecordChanged(t, socket, 300*time.Millisecond)
		})
	}
}

func TestPhase9Sprint7_CoordinationSavedViewsRemainAdditive_U_9_08(t *testing.T) {
	ctx := context.Background()
	harness := recordstoretest.StartStore(t, "phase9-sprint7-coordination-saved-views")
	actor := recordstoretest.SeedLocalUserFlags(t, harness.DB, "sprint7-saved-views@example.test", "Sprint7 Saved Views", "Sprint7SavedViews1!", false, false, true)
	incident := recordstoretest.CreateIncidentInStore(t, harness.DB, actor, "txn-phase9-sprint7-saved-views-incident", "IR-S7-SAVED-VIEWS", "Phase 9 Sprint 7 coordination saved views")
	savedViewStore := savedviews.NewStore(harness.DB)
	startupStore := workbookstartup.NewStore(harness.DB)
	workbookStore := workbook.NewStore(harness.DB)

	for _, viewSchemaID := range []string{
		workbook.CommLogViewSchemaID,
		workbook.HandoffViewSchemaID,
		workbook.StatusReviewViewSchemaID,
		workbook.LessonViewSchemaID,
	} {
		t.Run(viewSchemaID, func(t *testing.T) {
			createRequest := sprint7SavedViewCreateRequest(t, viewSchemaID)
			created, err := savedViewStore.Create(ctx, actor, incident.ID, createRequest, sprint7Time(0))
			if err != nil {
				t.Fatalf("create coordination saved view: %v", err)
			}
			if created.ViewSchemaID != viewSchemaID {
				t.Fatalf("saved view stored wrong view_schema_id: got %q want %q", created.ViewSchemaID, viewSchemaID)
			}
			savedViewID := created.SavedViewID.String()

			startup, err := startupStore.Resolve(ctx, incident.ID, actor.ID, "admin", sprint7SheetRefJSON(t, "saved_view", savedViewID), sprint7Time(1))
			if err != nil {
				t.Fatalf("resolve startup saved view: %v", err)
			}
			if startup.SelectedViewSchemaID == nil || *startup.SelectedViewSchemaID != viewSchemaID {
				t.Fatalf("startup selected wrong view schema: got %v want %q", startup.SelectedViewSchemaID, viewSchemaID)
			}
			requireSprint7SheetRefJSON(t, startup.SelectedSheetRef, "saved_view", savedViewID)
			if startup.SelectedSavedView == nil {
				t.Fatalf("startup selected saved view details must be present")
			}
			if startup.SelectedSavedView.ViewSchemaID != viewSchemaID {
				t.Fatalf("startup selected saved view changed identity: got %q want %q", startup.SelectedSavedView.ViewSchemaID, viewSchemaID)
			}
			if _, err := workbookStore.QueryRows(ctx, incident.ID, viewSchemaID, sprint7DefaultQueryMeta(t, viewSchemaID)); err != nil {
				t.Fatalf("canonical query changed identity for %s: %v", viewSchemaID, err)
			}
		})
	}
}

func sprint7SavedViewCreateRequest(t testing.TB, viewSchemaID string) savedviews.CreateRequest {
	t.Helper()
	payload, err := json.Marshal(map[string]any{
		"view_schema_id": viewSchemaID,
		"display_name":   "Saved " + viewSchemaID,
		"query_json":     map[string]any{},
		"layout_json":    map[string]any{},
	})
	if err != nil {
		t.Fatalf("marshal saved view request: %v", err)
	}
	request, apiErr := savedviews.DecodeCreateRequest(strings.NewReader(string(payload)))
	if apiErr != nil {
		t.Fatalf("decode saved view request: %v", apiErr)
	}
	return request
}

func sprint7SheetRefJSON(t testing.TB, kind string, id string) []byte {
	t.Helper()
	payload, err := json.Marshal(map[string]string{"kind": kind, "id": id})
	if err != nil {
		t.Fatalf("marshal sheet ref: %v", err)
	}
	return payload
}

func requireSprint7SheetRefJSON(t testing.TB, raw []byte, wantKind string, wantID string) {
	t.Helper()
	var ref map[string]string
	if err := json.Unmarshal(raw, &ref); err != nil {
		t.Fatalf("decode selected sheet ref: %v", err)
	}
	if ref["kind"] != wantKind || ref["id"] != wantID {
		t.Fatalf("unexpected sheet ref: got %#v want kind=%q id=%q", ref, wantKind, wantID)
	}
}

func sprint7DefaultQueryMeta(t testing.TB, viewSchemaID string) viewschema.QueryMeta {
	t.Helper()
	schema, ok := viewschema.Lookup(viewSchemaID)
	if !ok {
		t.Fatalf("missing view schema %s", viewSchemaID)
	}
	return schema.DefaultQueryMeta()
}

func TestPhase9Sprint7_CoordinationProjectionSortFilterGroup_U_9_08(t *testing.T) {
	harness := recordstoretest.StartStore(t, "phase9-sprint7-coordination-projections")
	store := workbook.NewStore(harness.DB)
	actor := recordstoretest.SeedLocalUserFlags(t, harness.DB, "sprint7-projection@example.test", "Sprint7 Projection", "Sprint7Projection1!", false, false, true)
	incident := recordstoretest.CreateIncidentInStore(t, harness.DB, actor, "txn-phase9-sprint7-projection-incident", "IR-S7-PROJECTION", "Phase 9 Sprint 7 coordination projections")

	commOneNextReport := time.Date(2026, 5, 20, 9, 0, 0, 0, time.UTC)
	_ = mustCreateSprint7Row(t, store, actor, incident.ID, workbook.CommLogViewSchemaID, "txn-phase9-sprint7-comm-projection-one", map[string]workbook.ValueChange{
		"comm_log.timestamp_utc":      sprint7Timestamp(time.Date(2026, 5, 18, 9, 0, 0, 0, time.UTC)),
		"comm_log.comm_type":          sprint7Text("briefing"),
		"comm_log.audience":           sprint7Text("Leads"),
		"comm_log.channel_or_meeting": sprint7Text("Morning bridge"),
		"comm_log.summary":            sprint7Text("Projection briefing"),
		"comm_log.next_report_at":     sprint7Timestamp(commOneNextReport),
		"comm_log.privilege_tag":      sprint7Text("internal"),
	}, nil, sprint7Time(0))
	commTwoNextReport := time.Date(2026, 5, 21, 9, 0, 0, 0, time.UTC)
	commTwo := mustCreateSprint7Row(t, store, actor, incident.ID, workbook.CommLogViewSchemaID, "txn-phase9-sprint7-comm-projection-two", map[string]workbook.ValueChange{
		"comm_log.timestamp_utc":      sprint7Timestamp(time.Date(2026, 5, 19, 9, 0, 0, 0, time.UTC)),
		"comm_log.comm_type":          sprint7Text("notification"),
		"comm_log.audience":           sprint7Text("Duty managers"),
		"comm_log.channel_or_meeting": sprint7Text("Email"),
		"comm_log.summary":            sprint7Text("Projection notification"),
		"comm_log.next_report_at":     sprint7Timestamp(commTwoNextReport),
	}, nil, sprint7Time(time.Minute))
	requireSprint7ProjectedRow(t, store, incident.ID, workbook.CommLogViewSchemaID, viewschema.QueryMeta{
		Filters: []viewschema.Filter{
			{FieldKey: "comm_log.comm_type", Op: "eq", Arg: map[string]any{"value": "notification"}},
			{FieldKey: "comm_log.timestamp_day", Op: "eq", Arg: map[string]any{"value": "2026-05-19"}},
			{FieldKey: "comm_log.next_report_day", Op: "eq", Arg: map[string]any{"value": "2026-05-21"}},
		},
		Sort:    []viewschema.SortEntry{{FieldKey: "comm_log.next_report_day", Direction: "desc"}, {FieldKey: "record_id", Direction: "asc"}},
		GroupBy: sprint7StringPtr("comm_log.comm_type"),
	}, commTwo.RecordID, "comm_log.comm_type", "notification")

	handoffPending := mustCreateSprint7Row(t, store, actor, incident.ID, workbook.HandoffViewSchemaID, "txn-phase9-sprint7-handoff-projection-pending", map[string]workbook.ValueChange{
		"handoff.timestamp_utc":          sprint7Timestamp(time.Date(2026, 5, 18, 10, 0, 0, 0, time.UTC)),
		"handoff.incoming_owner_user_id": sprint7UUID(actor.ID),
		"handoff.current_state_summary":  sprint7Text("Pending projection handoff"),
	}, nil, sprint7Time(2*time.Minute))
	acknowledgedAt := time.Date(2026, 5, 19, 11, 0, 0, 0, time.UTC)
	handoffAck := mustCreateSprint7Row(t, store, actor, incident.ID, workbook.HandoffViewSchemaID, "txn-phase9-sprint7-handoff-projection-ack", map[string]workbook.ValueChange{
		"handoff.timestamp_utc":          sprint7Timestamp(time.Date(2026, 5, 19, 10, 0, 0, 0, time.UTC)),
		"handoff.incoming_owner_user_id": sprint7UUID(actor.ID),
		"handoff.current_state_summary":  sprint7Text("Acknowledged projection handoff"),
		"handoff.acknowledged_at":        sprint7Timestamp(acknowledgedAt),
	}, nil, sprint7Time(3*time.Minute))
	requireSprint7ProjectedRow(t, store, incident.ID, workbook.HandoffViewSchemaID, viewschema.QueryMeta{
		Filters: []viewschema.Filter{
			{FieldKey: "handoff.incoming_owner_user_id", Op: "eq", Arg: map[string]any{"value": actor.ID.String()}},
			{FieldKey: "handoff.ack_state", Op: "eq", Arg: map[string]any{"value": "acknowledged"}},
			{FieldKey: "handoff.timestamp_day", Op: "eq", Arg: map[string]any{"value": "2026-05-19"}},
		},
		Sort:    []viewschema.SortEntry{{FieldKey: "handoff.outgoing_owner_user_id", Direction: "asc"}, {FieldKey: "record_id", Direction: "asc"}},
		GroupBy: sprint7StringPtr("handoff.ack_state"),
	}, handoffAck.RecordID, "handoff.ack_state", "acknowledged")
	ackAsc, err := store.QueryRows(context.Background(), incident.ID, workbook.HandoffViewSchemaID, viewschema.QueryMeta{
		Sort: []viewschema.SortEntry{{FieldKey: "handoff.ack_state", Direction: "asc"}, {FieldKey: "record_id", Direction: "asc"}},
	})
	if err != nil {
		t.Fatalf("query handoff ack_state ascending order: %v", err)
	}
	requireSprint7RecordOrder(t, ackAsc, []uuid.UUID{handoffPending.RecordID, handoffAck.RecordID})
	ackDesc, err := store.QueryRows(context.Background(), incident.ID, workbook.HandoffViewSchemaID, viewschema.QueryMeta{
		Sort: []viewschema.SortEntry{{FieldKey: "handoff.ack_state", Direction: "desc"}, {FieldKey: "record_id", Direction: "asc"}},
	})
	if err != nil {
		t.Fatalf("query handoff ack_state descending order: %v", err)
	}
	requireSprint7RecordOrder(t, ackDesc, []uuid.UUID{handoffAck.RecordID, handoffPending.RecordID})

	_ = mustCreateSprint7Row(t, store, actor, incident.ID, workbook.StatusReviewViewSchemaID, "txn-phase9-sprint7-status-projection-one", map[string]workbook.ValueChange{
		"status_review.timestamp_utc":         sprint7Timestamp(time.Date(2026, 5, 18, 12, 0, 0, 0, time.UTC)),
		"status_review.current_state_summary": sprint7Text("Status review baseline"),
	}, nil, sprint7Time(4*time.Minute))
	statusNextReport := time.Date(2026, 5, 22, 12, 0, 0, 0, time.UTC)
	status := mustCreateSprint7Row(t, store, actor, incident.ID, workbook.StatusReviewViewSchemaID, "txn-phase9-sprint7-status-projection-two", map[string]workbook.ValueChange{
		"status_review.timestamp_utc":         sprint7Timestamp(time.Date(2026, 5, 19, 12, 0, 0, 0, time.UTC)),
		"status_review.review_owner_user_id":  sprint7UUID(actor.ID),
		"status_review.current_state_summary": sprint7Text("Status review next report"),
		"status_review.next_report_at":        sprint7Timestamp(statusNextReport),
	}, nil, sprint7Time(5*time.Minute))
	requireSprint7ProjectedRow(t, store, incident.ID, workbook.StatusReviewViewSchemaID, viewschema.QueryMeta{
		Filters: []viewschema.Filter{
			{FieldKey: "status_review.review_owner_user_id", Op: "eq", Arg: map[string]any{"value": actor.ID.String()}},
			{FieldKey: "status_review.timestamp_day", Op: "eq", Arg: map[string]any{"value": "2026-05-19"}},
			{FieldKey: "status_review.next_report_day", Op: "eq", Arg: map[string]any{"value": "2026-05-22"}},
		},
		Sort:    []viewschema.SortEntry{{FieldKey: "status_review.next_report_day", Direction: "desc"}, {FieldKey: "record_id", Direction: "asc"}},
		GroupBy: sprint7StringPtr("status_review.review_owner_user_id"),
	}, status.RecordID, "status_review.review_owner_user_id", actor.ID.String())

	_ = mustCreateSprint7Row(t, store, actor, incident.ID, workbook.LessonViewSchemaID, "txn-phase9-sprint7-lesson-projection-open", map[string]workbook.ValueChange{
		"lesson.timestamp_utc": sprint7Timestamp(time.Date(2026, 5, 18, 13, 0, 0, 0, time.UTC)),
		"lesson.summary":       sprint7Text("Open lesson projection"),
	}, nil, sprint7Time(6*time.Minute))
	lesson := mustCreateSprint7Row(t, store, actor, incident.ID, workbook.LessonViewSchemaID, "txn-phase9-sprint7-lesson-projection-closed", map[string]workbook.ValueChange{
		"lesson.timestamp_utc": sprint7Timestamp(time.Date(2026, 5, 19, 13, 0, 0, 0, time.UTC)),
		"lesson.summary":       sprint7Text("Closed lesson projection"),
		"lesson.owner_user_id": sprint7UUID(actor.ID),
		"lesson.closure_state": sprint7Text("closed"),
	}, nil, sprint7Time(7*time.Minute))
	requireSprint7ProjectedRow(t, store, incident.ID, workbook.LessonViewSchemaID, viewschema.QueryMeta{
		Filters: []viewschema.Filter{
			{FieldKey: "lesson.closure_state", Op: "eq", Arg: map[string]any{"value": "closed"}},
			{FieldKey: "lesson.owner_user_id", Op: "eq", Arg: map[string]any{"value": actor.ID.String()}},
			{FieldKey: "lesson.timestamp_day", Op: "eq", Arg: map[string]any{"value": "2026-05-19"}},
		},
		Sort:    []viewschema.SortEntry{{FieldKey: "lesson.closure_state", Direction: "asc"}, {FieldKey: "record_id", Direction: "asc"}},
		GroupBy: sprint7StringPtr("lesson.closure_state"),
	}, lesson.RecordID, "lesson.closure_state", "closed")
}

func TestPhase9Sprint7_CoordinationDeclaredQueryFieldsAreMapped_U_9_08(t *testing.T) {
	harness := recordstoretest.StartStore(t, "phase9-sprint7-coordination-query-fields")
	store := workbook.NewStore(harness.DB)
	actor := recordstoretest.SeedLocalUserFlags(t, harness.DB, "sprint7-query-fields@example.test", "Sprint7 Query Fields", "Sprint7QueryFields1!", false, false, true)
	incident := recordstoretest.CreateIncidentInStore(t, harness.DB, actor, "txn-phase9-sprint7-query-fields-incident", "IR-S7-QUERY-FIELDS", "Phase 9 Sprint 7 declared query fields")

	mustCreateSprint7Row(t, store, actor, incident.ID, workbook.CommLogViewSchemaID, "txn-phase9-sprint7-query-fields-comm", map[string]workbook.ValueChange{
		"comm_log.comm_type":          sprint7Text("briefing"),
		"comm_log.audience":           sprint7Text("Query field audience"),
		"comm_log.channel_or_meeting": sprint7Text("Bridge"),
		"comm_log.summary":            sprint7Text("Query field comm log"),
		"comm_log.privilege_tag":      sprint7Text("internal"),
	}, nil, sprint7Time(0))
	mustCreateSprint7Row(t, store, actor, incident.ID, workbook.HandoffViewSchemaID, "txn-phase9-sprint7-query-fields-handoff", map[string]workbook.ValueChange{
		"handoff.incoming_owner_user_id": sprint7UUID(actor.ID),
		"handoff.current_state_summary":  sprint7Text("Query field handoff"),
	}, nil, sprint7Time(time.Minute))
	mustCreateSprint7Row(t, store, actor, incident.ID, workbook.StatusReviewViewSchemaID, "txn-phase9-sprint7-query-fields-status", map[string]workbook.ValueChange{
		"status_review.current_state_summary": sprint7Text("Query field status review"),
	}, nil, sprint7Time(2*time.Minute))
	mustCreateSprint7Row(t, store, actor, incident.ID, workbook.LessonViewSchemaID, "txn-phase9-sprint7-query-fields-lesson", map[string]workbook.ValueChange{
		"lesson.summary": sprint7Text("Query field lesson"),
	}, nil, sprint7Time(3*time.Minute))

	for _, viewSchemaID := range []string{
		workbook.CommLogViewSchemaID,
		workbook.HandoffViewSchemaID,
		workbook.StatusReviewViewSchemaID,
		workbook.LessonViewSchemaID,
	} {
		t.Run(viewSchemaID, func(t *testing.T) {
			schema, ok := viewschema.Lookup(viewSchemaID)
			if !ok {
				t.Fatalf("missing schema %s", viewSchemaID)
			}
			for _, fieldKey := range schema.SortFields() {
				sort := []viewschema.SortEntry{{FieldKey: fieldKey, Direction: "asc"}, {FieldKey: "record_id", Direction: "asc"}}
				if _, err := store.QueryRows(context.Background(), incident.ID, viewSchemaID, viewschema.QueryMeta{Sort: sort}); err != nil {
					t.Fatalf("%s sort field %s is not queryable: %v", viewSchemaID, fieldKey, err)
				}
			}
			for _, fieldKey := range schema.FilterFields() {
				query := viewschema.QueryMeta{
					Filters: []viewschema.Filter{{FieldKey: fieldKey, Op: "eq", Arg: map[string]any{"value": nil}}},
					Sort:    schema.DefaultSort(),
				}
				if _, err := store.QueryRows(context.Background(), incident.ID, viewSchemaID, query); err != nil {
					t.Fatalf("%s filter field %s is not queryable: %v", viewSchemaID, fieldKey, err)
				}
			}
			for _, fieldKey := range schema.GroupingFields() {
				groupBy := fieldKey
				rows, err := store.QueryRows(context.Background(), incident.ID, viewSchemaID, viewschema.QueryMeta{
					Sort:    schema.DefaultSort(),
					GroupBy: &groupBy,
				})
				if err != nil {
					t.Fatalf("%s grouping field %s is not queryable: %v", viewSchemaID, fieldKey, err)
				}
				if len(rows) == 0 {
					t.Fatalf("%s grouping field %s returned no rows", viewSchemaID, fieldKey)
				}
				if _, exists := rows[0]["group_values"].(map[string]any)[fieldKey]; !exists {
					t.Fatalf("%s grouping field %s missing group value in %#v", viewSchemaID, fieldKey, rows[0])
				}
			}
		})
	}
}

func TestPhase9Sprint7_CoordinationSemanticFilters_U_9_08(t *testing.T) {
	harness := recordstoretest.StartStore(t, "phase9-sprint7-coordination-semantic-filters")
	store := workbook.NewStore(harness.DB)
	actor := recordstoretest.SeedLocalUserFlags(t, harness.DB, "sprint7-filters@example.test", "Sprint7 Filters", "Sprint7Filters1!", false, false, true)
	alternate := recordstoretest.SeedLocalUserFlags(t, harness.DB, "sprint7-filters-alt@example.test", "Sprint7 Filters Alt", "Sprint7FiltersAlt1!", false, false, true)
	incident := recordstoretest.CreateIncidentInStore(t, harness.DB, actor, "txn-phase9-sprint7-filters-incident", "IR-S7-FILTERS", "Phase 9 Sprint 7 semantic filters")

	commPositive := mustCreateSprint7Row(t, store, actor, incident.ID, workbook.CommLogViewSchemaID, "txn-phase9-sprint7-filter-comm-positive", map[string]workbook.ValueChange{
		"comm_log.timestamp_utc":      sprint7Timestamp(time.Date(2026, 5, 19, 9, 0, 0, 0, time.UTC)),
		"comm_log.comm_type":          sprint7Text("notification"),
		"comm_log.audience":           sprint7Text("Incident Command"),
		"comm_log.channel_or_meeting": sprint7Text("Bridge Alpha"),
		"comm_log.summary":            sprint7Text("Semantic filter positive comm log"),
		"comm_log.next_report_at":     sprint7Timestamp(time.Date(2026, 5, 22, 9, 0, 0, 0, time.UTC)),
		"comm_log.privilege_tag":      sprint7Text("privileged"),
	}, nil, sprint7Time(0))
	mustCreateSprint7Row(t, store, actor, incident.ID, workbook.CommLogViewSchemaID, "txn-phase9-sprint7-filter-comm-negative", map[string]workbook.ValueChange{
		"comm_log.timestamp_utc":      sprint7Timestamp(time.Date(2026, 5, 20, 9, 0, 0, 0, time.UTC)),
		"comm_log.comm_type":          sprint7Text("briefing"),
		"comm_log.audience":           sprint7Text("Engineering"),
		"comm_log.channel_or_meeting": sprint7Text("Email"),
		"comm_log.summary":            sprint7Text("Semantic filter negative comm log"),
		"comm_log.next_report_at":     sprint7Timestamp(time.Date(2026, 5, 23, 9, 0, 0, 0, time.UTC)),
		"comm_log.privilege_tag":      sprint7Text("public"),
	}, nil, sprint7Time(time.Minute))

	handoffPositive := mustCreateSprint7Row(t, store, actor, incident.ID, workbook.HandoffViewSchemaID, "txn-phase9-sprint7-filter-handoff-positive", map[string]workbook.ValueChange{
		"handoff.timestamp_utc":          sprint7Timestamp(time.Date(2026, 5, 19, 10, 0, 0, 0, time.UTC)),
		"handoff.outgoing_owner_user_id": sprint7UUID(actor.ID),
		"handoff.incoming_owner_user_id": sprint7UUID(alternate.ID),
		"handoff.current_state_summary":  sprint7Text("Semantic filter positive handoff"),
		"handoff.acknowledged_at":        sprint7Timestamp(time.Date(2026, 5, 19, 11, 0, 0, 0, time.UTC)),
	}, nil, sprint7Time(2*time.Minute))
	mustCreateSprint7Row(t, store, actor, incident.ID, workbook.HandoffViewSchemaID, "txn-phase9-sprint7-filter-handoff-negative", map[string]workbook.ValueChange{
		"handoff.timestamp_utc":          sprint7Timestamp(time.Date(2026, 5, 20, 10, 0, 0, 0, time.UTC)),
		"handoff.outgoing_owner_user_id": sprint7UUID(alternate.ID),
		"handoff.incoming_owner_user_id": sprint7UUID(actor.ID),
		"handoff.current_state_summary":  sprint7Text("Semantic filter negative handoff"),
	}, nil, sprint7Time(3*time.Minute))

	statusPositive := mustCreateSprint7Row(t, store, actor, incident.ID, workbook.StatusReviewViewSchemaID, "txn-phase9-sprint7-filter-status-positive", map[string]workbook.ValueChange{
		"status_review.timestamp_utc":         sprint7Timestamp(time.Date(2026, 5, 19, 12, 0, 0, 0, time.UTC)),
		"status_review.review_owner_user_id":  sprint7UUID(actor.ID),
		"status_review.current_state_summary": sprint7Text("Semantic filter positive status"),
		"status_review.next_report_at":        sprint7Timestamp(time.Date(2026, 5, 22, 12, 0, 0, 0, time.UTC)),
	}, nil, sprint7Time(4*time.Minute))
	mustCreateSprint7Row(t, store, actor, incident.ID, workbook.StatusReviewViewSchemaID, "txn-phase9-sprint7-filter-status-negative", map[string]workbook.ValueChange{
		"status_review.timestamp_utc":         sprint7Timestamp(time.Date(2026, 5, 20, 12, 0, 0, 0, time.UTC)),
		"status_review.review_owner_user_id":  sprint7UUID(alternate.ID),
		"status_review.current_state_summary": sprint7Text("Semantic filter negative status"),
		"status_review.next_report_at":        sprint7Timestamp(time.Date(2026, 5, 23, 12, 0, 0, 0, time.UTC)),
	}, nil, sprint7Time(5*time.Minute))

	lessonPositive := mustCreateSprint7Row(t, store, actor, incident.ID, workbook.LessonViewSchemaID, "txn-phase9-sprint7-filter-lesson-positive", map[string]workbook.ValueChange{
		"lesson.timestamp_utc": sprint7Timestamp(time.Date(2026, 5, 19, 13, 0, 0, 0, time.UTC)),
		"lesson.summary":       sprint7Text("Semantic filter positive lesson"),
		"lesson.owner_user_id": sprint7UUID(actor.ID),
		"lesson.closure_state": sprint7Text("closed"),
	}, nil, sprint7Time(6*time.Minute))
	mustCreateSprint7Row(t, store, actor, incident.ID, workbook.LessonViewSchemaID, "txn-phase9-sprint7-filter-lesson-negative", map[string]workbook.ValueChange{
		"lesson.timestamp_utc": sprint7Timestamp(time.Date(2026, 5, 20, 13, 0, 0, 0, time.UTC)),
		"lesson.summary":       sprint7Text("Semantic filter negative lesson"),
		"lesson.owner_user_id": sprint7UUID(alternate.ID),
		"lesson.closure_state": sprint7Text("open"),
	}, nil, sprint7Time(7*time.Minute))

	for _, tc := range []struct {
		name         string
		viewSchemaID string
		fieldKey     string
		op           string
		value        any
		wantRecordID uuid.UUID
	}{
		{"comm-type-eq", workbook.CommLogViewSchemaID, "comm_log.comm_type", "eq", "notification", commPositive.RecordID},
		{"comm-type-prefix", workbook.CommLogViewSchemaID, "comm_log.comm_type", "prefix", "not", commPositive.RecordID},
		{"comm-audience-eq", workbook.CommLogViewSchemaID, "comm_log.audience", "eq", "incident command", commPositive.RecordID},
		{"comm-audience-prefix", workbook.CommLogViewSchemaID, "comm_log.audience", "prefix", "incident", commPositive.RecordID},
		{"comm-channel-eq", workbook.CommLogViewSchemaID, "comm_log.channel_or_meeting", "eq", "bridge alpha", commPositive.RecordID},
		{"comm-channel-prefix", workbook.CommLogViewSchemaID, "comm_log.channel_or_meeting", "prefix", "bridge", commPositive.RecordID},
		{"comm-privilege-eq", workbook.CommLogViewSchemaID, "comm_log.privilege_tag", "eq", "privileged", commPositive.RecordID},
		{"comm-privilege-prefix", workbook.CommLogViewSchemaID, "comm_log.privilege_tag", "prefix", "priv", commPositive.RecordID},
		{"comm-timestamp-day", workbook.CommLogViewSchemaID, "comm_log.timestamp_day", "eq", "2026-05-19", commPositive.RecordID},
		{"comm-next-report-day", workbook.CommLogViewSchemaID, "comm_log.next_report_day", "eq", "2026-05-22", commPositive.RecordID},
		{"handoff-outgoing-eq", workbook.HandoffViewSchemaID, "handoff.outgoing_owner_user_id", "eq", actor.ID.String(), handoffPositive.RecordID},
		{"handoff-outgoing-prefix", workbook.HandoffViewSchemaID, "handoff.outgoing_owner_user_id", "prefix", actor.ID.String()[:8], handoffPositive.RecordID},
		{"handoff-incoming-eq", workbook.HandoffViewSchemaID, "handoff.incoming_owner_user_id", "eq", alternate.ID.String(), handoffPositive.RecordID},
		{"handoff-incoming-prefix", workbook.HandoffViewSchemaID, "handoff.incoming_owner_user_id", "prefix", alternate.ID.String()[:8], handoffPositive.RecordID},
		{"handoff-timestamp-day", workbook.HandoffViewSchemaID, "handoff.timestamp_day", "eq", "2026-05-19", handoffPositive.RecordID},
		{"handoff-ack-eq", workbook.HandoffViewSchemaID, "handoff.ack_state", "eq", "acknowledged", handoffPositive.RecordID},
		{"handoff-ack-prefix", workbook.HandoffViewSchemaID, "handoff.ack_state", "prefix", "ack", handoffPositive.RecordID},
		{"status-owner-eq", workbook.StatusReviewViewSchemaID, "status_review.review_owner_user_id", "eq", actor.ID.String(), statusPositive.RecordID},
		{"status-owner-prefix", workbook.StatusReviewViewSchemaID, "status_review.review_owner_user_id", "prefix", actor.ID.String()[:8], statusPositive.RecordID},
		{"status-timestamp-day", workbook.StatusReviewViewSchemaID, "status_review.timestamp_day", "eq", "2026-05-19", statusPositive.RecordID},
		{"status-next-report-day", workbook.StatusReviewViewSchemaID, "status_review.next_report_day", "eq", "2026-05-22", statusPositive.RecordID},
		{"lesson-owner-eq", workbook.LessonViewSchemaID, "lesson.owner_user_id", "eq", actor.ID.String(), lessonPositive.RecordID},
		{"lesson-owner-prefix", workbook.LessonViewSchemaID, "lesson.owner_user_id", "prefix", actor.ID.String()[:8], lessonPositive.RecordID},
		{"lesson-closure-eq", workbook.LessonViewSchemaID, "lesson.closure_state", "eq", "closed", lessonPositive.RecordID},
		{"lesson-closure-prefix", workbook.LessonViewSchemaID, "lesson.closure_state", "prefix", "clo", lessonPositive.RecordID},
		{"lesson-timestamp-day", workbook.LessonViewSchemaID, "lesson.timestamp_day", "eq", "2026-05-19", lessonPositive.RecordID},
	} {
		t.Run(tc.name, func(t *testing.T) {
			requireSprint7FilterMatchesOnly(t, store, incident.ID, tc.viewSchemaID, viewschema.Filter{
				FieldKey: tc.fieldKey,
				Op:       tc.op,
				Arg:      map[string]any{"value": tc.value},
			}, tc.wantRecordID)
		})
	}
}

func TestPhase9Sprint7_CoordinationCollectionItemShapes_U_9_08(t *testing.T) {
	harness := recordstoretest.StartStore(t, "phase9-sprint7-coordination-collection-shapes")
	store := workbook.NewStore(harness.DB)
	actor := recordstoretest.SeedLocalUserFlags(t, harness.DB, "sprint7-shapes@example.test", "Sprint7 Shapes", "Sprint7Shapes1!", false, false, true)
	incident := recordstoretest.CreateIncidentInStore(t, harness.DB, actor, "txn-phase9-sprint7-shapes-incident", "IR-S7-SHAPES", "Phase 9 Sprint 7 collection item shapes")

	partyID := mustCreatePartyForU911(t, store, actor, incident.ID, "txn-phase9-sprint7-shapes-party", "Coordination Shape Party")
	attendeePartyID := mustCreatePartyForU911(t, store, actor, incident.ID, "txn-phase9-sprint7-shapes-attendee", "Coordination Shape Attendee")
	decisionID := mustCreateSprint6Decision(t, store, actor, incident.ID, "txn-phase9-sprint7-shapes-decision", "approved", "Shape decision")
	taskID := mustCreateTaskForU911(t, store, actor, incident.ID, "txn-phase9-sprint7-shapes-task", "Shape task")
	evidenceID := mustCreateEvidenceForU911(t, store, actor, incident.ID, "txn-phase9-sprint7-shapes-evidence", "Shape evidence")

	comm := mustCreateSprint7Row(t, store, actor, incident.ID, workbook.CommLogViewSchemaID, "txn-phase9-sprint7-shapes-comm", sprint7MinimumValues(actor.ID, workbook.CommLogViewSchemaID), map[string]workbook.CollectionActionPayload{
		"comm_log.decision_ids":       sprint6Collection(addSprint6RecordRef(decisionID)),
		"comm_log.action_task_ids":    sprint6Collection(addSprint6RecordRef(taskID)),
		"comm_log.audience_party_ids": sprint6Collection(workbook.CollectionAction{Op: "add_party_ref", PartyID: &partyID}),
		"comm_log.attendee_party_ids": sprint6Collection(workbook.CollectionAction{Op: "add_party_ref", PartyID: &attendeePartyID}),
	}, sprint7Time(0))
	commRow := comm.Payload["row"].(map[string]any)
	requireSprint7RecordRefItemShape(t, commRow, "comm_log.decision_ids", decisionID)
	requireSprint7RecordRefItemShape(t, commRow, "comm_log.action_task_ids", taskID)
	requireSprint7PartyRefItemShape(t, commRow, "comm_log.audience_party_ids", partyID)
	requireSprint7PartyRefItemShape(t, commRow, "comm_log.attendee_party_ids", attendeePartyID)

	handoff := mustCreateSprint7Row(t, store, actor, incident.ID, workbook.HandoffViewSchemaID, "txn-phase9-sprint7-shapes-handoff", sprint7MinimumValues(actor.ID, workbook.HandoffViewSchemaID), map[string]workbook.CollectionActionPayload{
		"handoff.open_task_ids":     sprint6Collection(addSprint6RecordRef(taskID)),
		"handoff.open_decision_ids": sprint6Collection(addSprint6RecordRef(decisionID)),
		"handoff.open_risk_refs":    {Actions: []workbook.CollectionAction{{Op: "add_risk_ref", RiskRefText: "Escalate outbound access", NormalizedText: "escalate outbound access"}}},
	}, sprint7Time(time.Minute))
	handoffRow := handoff.Payload["row"].(map[string]any)
	requireSprint7RecordRefItemShape(t, handoffRow, "handoff.open_task_ids", taskID)
	requireSprint7RecordRefItemShape(t, handoffRow, "handoff.open_decision_ids", decisionID)
	requireSprint7RiskRefItemShape(t, handoffRow, "handoff.open_risk_refs", "Escalate outbound access")

	status := mustCreateSprint7Row(t, store, actor, incident.ID, workbook.StatusReviewViewSchemaID, "txn-phase9-sprint7-shapes-status", sprint7MinimumValues(actor.ID, workbook.StatusReviewViewSchemaID), map[string]workbook.CollectionActionPayload{
		"status_review.blocked_task_ids":     sprint6Collection(addSprint6RecordRef(taskID)),
		"status_review.pending_evidence_ids": sprint6Collection(addSprint6RecordRef(evidenceID)),
		"status_review.open_decision_ids":    sprint6Collection(addSprint6RecordRef(decisionID)),
	}, sprint7Time(2*time.Minute))
	statusRow := status.Payload["row"].(map[string]any)
	requireSprint7RecordRefItemShape(t, statusRow, "status_review.blocked_task_ids", taskID)
	requireSprint7RecordRefItemShape(t, statusRow, "status_review.pending_evidence_ids", evidenceID)
	requireSprint7RecordRefItemShape(t, statusRow, "status_review.open_decision_ids", decisionID)

	lesson := mustCreateSprint7Row(t, store, actor, incident.ID, workbook.LessonViewSchemaID, "txn-phase9-sprint7-shapes-lesson", sprint7MinimumValues(actor.ID, workbook.LessonViewSchemaID), map[string]workbook.CollectionActionPayload{
		"lesson.follow_up_task_ids": sprint6Collection(addSprint6RecordRef(taskID)),
		"lesson.evidence_refs":      sprint6Collection(addSprint6RecordRef(evidenceID)),
	}, sprint7Time(3*time.Minute))
	lessonRow := lesson.Payload["row"].(map[string]any)
	requireSprint7RecordRefItemShape(t, lessonRow, "lesson.follow_up_task_ids", taskID)
	requireSprint7RecordRefItemShape(t, lessonRow, "lesson.evidence_refs", evidenceID)
}

func TestPhase9Sprint7_CoordinationDuplicateCoalescing_U_9_08(t *testing.T) {
	harness := recordstoretest.StartStore(t, "phase9-sprint7-coordination-duplicate-coalescing")
	store := workbook.NewStore(harness.DB)
	actor := recordstoretest.SeedLocalUserFlags(t, harness.DB, "sprint7-duplicates@example.test", "Sprint7 Duplicates", "Sprint7Duplicates1!", false, false, true)
	incident := recordstoretest.CreateIncidentInStore(t, harness.DB, actor, "txn-phase9-sprint7-duplicates-incident", "IR-S7-DUPLICATES", "Phase 9 Sprint 7 duplicate coalescing")

	partyID := mustCreatePartyForU911(t, store, actor, incident.ID, "txn-phase9-sprint7-duplicates-party", "Duplicate Party")
	decisionID := mustCreateSprint6Decision(t, store, actor, incident.ID, "txn-phase9-sprint7-duplicates-decision", "approved", "Duplicate decision")
	taskID := mustCreateTaskForU911(t, store, actor, incident.ID, "txn-phase9-sprint7-duplicates-task", "Duplicate task")
	evidenceID := mustCreateEvidenceForU911(t, store, actor, incident.ID, "txn-phase9-sprint7-duplicates-evidence", "Duplicate evidence")

	for _, tc := range []struct {
		name         string
		viewSchemaID string
		fieldKey     string
		action       workbook.CollectionAction
		wantLinkID   uuid.UUID
	}{
		{"comm-decision", workbook.CommLogViewSchemaID, "comm_log.decision_ids", addSprint6RecordRef(decisionID), decisionID},
		{"comm-task", workbook.CommLogViewSchemaID, "comm_log.action_task_ids", addSprint6RecordRef(taskID), taskID},
		{"comm-audience-party", workbook.CommLogViewSchemaID, "comm_log.audience_party_ids", workbook.CollectionAction{Op: "add_party_ref", PartyID: &partyID}, partyID},
		{"comm-attendee-party", workbook.CommLogViewSchemaID, "comm_log.attendee_party_ids", workbook.CollectionAction{Op: "add_party_ref", PartyID: &partyID}, partyID},
		{"handoff-task", workbook.HandoffViewSchemaID, "handoff.open_task_ids", addSprint6RecordRef(taskID), taskID},
		{"handoff-decision", workbook.HandoffViewSchemaID, "handoff.open_decision_ids", addSprint6RecordRef(decisionID), decisionID},
		{"status-task", workbook.StatusReviewViewSchemaID, "status_review.blocked_task_ids", addSprint6RecordRef(taskID), taskID},
		{"status-evidence", workbook.StatusReviewViewSchemaID, "status_review.pending_evidence_ids", addSprint6RecordRef(evidenceID), evidenceID},
		{"status-decision", workbook.StatusReviewViewSchemaID, "status_review.open_decision_ids", addSprint6RecordRef(decisionID), decisionID},
		{"lesson-task", workbook.LessonViewSchemaID, "lesson.follow_up_task_ids", addSprint6RecordRef(taskID), taskID},
		{"lesson-evidence", workbook.LessonViewSchemaID, "lesson.evidence_refs", addSprint6RecordRef(evidenceID), evidenceID},
	} {
		t.Run(tc.name, func(t *testing.T) {
			clientTxnID := "txn-phase9-sprint7-duplicates-" + tc.name
			values := sprint7MinimumValues(actor.ID, tc.viewSchemaID)
			collections := map[string]workbook.CollectionActionPayload{
				tc.fieldKey: sprint6Collection(tc.action, tc.action),
			}
			result := mustCreateSprint7Row(t, store, actor, incident.ID, tc.viewSchemaID, clientTxnID, values, collections, sprint7Time(0))
			requireSprint7CollectionItemCount(t, result.Payload["row"].(map[string]any), tc.fieldKey, 1)
			if got := countSprint7LinksForField(t, harness, result.RecordID, tc.fieldKey); got != 1 {
				t.Fatalf("%s duplicate create links: got %d want 1", tc.fieldKey, got)
			}
			requireSprint7ManualReferenceLink(t, harness, result.RecordID, tc.wantLinkID, tc.fieldKey, "references_record")

			replay := mustCreateSprint7Row(t, store, actor, incident.ID, tc.viewSchemaID, clientTxnID, values, collections, sprint7Time(30*time.Minute))
			if replay.RecordID != result.RecordID {
				t.Fatalf("%s replay record_id: got %s want %s", tc.fieldKey, replay.RecordID, result.RecordID)
			}
			requireSprint7CollectionItemCount(t, replay.Payload["row"].(map[string]any), tc.fieldKey, 1)
			if got := countSprint7LinksForField(t, harness, result.RecordID, tc.fieldKey); got != 1 {
				t.Fatalf("%s duplicate replay links: got %d want 1", tc.fieldKey, got)
			}
		})
	}

	riskClientTxnID := "txn-phase9-sprint7-duplicates-risk"
	riskValues := sprint7MinimumValues(actor.ID, workbook.HandoffViewSchemaID)
	riskCollections := map[string]workbook.CollectionActionPayload{
		"handoff.open_risk_refs": {
			Actions: []workbook.CollectionAction{
				{Op: "add_risk_ref", RiskRefText: " Repeated risk reference ", NormalizedText: "repeated risk reference"},
				{Op: "add_risk_ref", RiskRefText: "Repeated risk reference", NormalizedText: "repeated risk reference"},
			},
		},
	}
	handoff := mustCreateSprint7Row(t, store, actor, incident.ID, workbook.HandoffViewSchemaID, riskClientTxnID, riskValues, riskCollections, sprint7Time(time.Minute))
	requireSprint7CollectionItemCount(t, handoff.Payload["row"].(map[string]any), "handoff.open_risk_refs", 1)
	if got := countSprint7ActiveRiskRefs(t, harness, handoff.RecordID); got != 1 {
		t.Fatalf("handoff risk duplicate create refs: got %d want 1", got)
	}

	riskReplay := mustCreateSprint7Row(t, store, actor, incident.ID, workbook.HandoffViewSchemaID, riskClientTxnID, riskValues, riskCollections, sprint7Time(31*time.Minute))
	if riskReplay.RecordID != handoff.RecordID {
		t.Fatalf("handoff risk replay record_id: got %s want %s", riskReplay.RecordID, handoff.RecordID)
	}
	requireSprint7CollectionItemCount(t, riskReplay.Payload["row"].(map[string]any), "handoff.open_risk_refs", 1)
	if got := countSprint7ActiveRiskRefs(t, harness, handoff.RecordID); got != 1 {
		t.Fatalf("handoff risk duplicate replay refs: got %d want 1", got)
	}
}

func TestPhase9Sprint7_CoordinationCollectionsAndValidation_U_9_08(t *testing.T) {
	ctx := context.Background()
	harness := recordstoretest.StartStore(t, "phase9-sprint7-coordination-collections")
	store := workbook.NewStore(harness.DB)
	actor := recordstoretest.SeedLocalUserFlags(t, harness.DB, "sprint7-collections@example.test", "Sprint7 Collections", "Sprint7Collections1!", false, false, true)
	incident := recordstoretest.CreateIncidentInStore(t, harness.DB, actor, "txn-phase9-sprint7-collections-incident", "IR-S7-COLLECTIONS", "Phase 9 Sprint 7 coordination collections")

	partyID := mustCreatePartyForU911(t, store, actor, incident.ID, "txn-phase9-sprint7-party", "Coordination Legal")
	otherPartyID := mustCreatePartyForU911(t, store, actor, incident.ID, "txn-phase9-sprint7-party-other", "Coordination Legal Alternate")
	decisionID := mustCreateSprint6Decision(t, store, actor, incident.ID, "txn-phase9-sprint7-decision", "approved", "Approve coordination plan")
	taskID := mustCreateTaskForU911(t, store, actor, incident.ID, "txn-phase9-sprint7-task", "Coordinate endpoint logs")
	evidenceID := mustCreateEvidenceForU911(t, store, actor, incident.ID, "txn-phase9-sprint7-evidence", "Coordination evidence")

	otherIncident := recordstoretest.CreateIncidentInStore(t, harness.DB, actor, "txn-phase9-sprint7-foreign-incident", "IR-S7-COLLECTIONS-FOREIGN", "Phase 9 Sprint 7 foreign incident")
	foreignEvidenceID := mustCreateEvidenceForU911(t, store, actor, otherIncident.ID, "txn-phase9-sprint7-foreign-evidence", "Foreign evidence")
	foreignTaskID := mustCreateTaskForU911(t, store, actor, otherIncident.ID, "txn-phase9-sprint7-foreign-task", "Foreign task")
	foreignDecisionID := mustCreateSprint6Decision(t, store, actor, otherIncident.ID, "txn-phase9-sprint7-foreign-decision", "approved", "Foreign decision")
	foreignPartyID := mustCreatePartyForU911(t, store, actor, otherIncident.ID, "txn-phase9-sprint7-foreign-party", "Foreign party")
	deletedEvidenceID := mustCreateEvidenceForU911(t, store, actor, incident.ID, "txn-phase9-sprint7-deleted-evidence", "Deleted evidence")
	deletedTaskID := mustCreateTaskForU911(t, store, actor, incident.ID, "txn-phase9-sprint7-deleted-task", "Deleted task")
	deletedDecisionID := mustCreateSprint6Decision(t, store, actor, incident.ID, "txn-phase9-sprint7-deleted-decision", "approved", "Deleted decision")
	deletedPartyID := mustCreatePartyForU911(t, store, actor, incident.ID, "txn-phase9-sprint7-deleted-party", "Deleted party")
	for label, recordID := range map[string]uuid.UUID{
		"evidence": deletedEvidenceID,
		"task":     deletedTaskID,
		"decision": deletedDecisionID,
		"party":    deletedPartyID,
	} {
		if _, err := harness.DB.Exec(ctx, `UPDATE records SET deleted_at = $2, deleted_by_user_id = $3 WHERE record_id = $1`, recordID, sprint7Time(time.Hour), actor.ID); err != nil {
			t.Fatalf("soft delete %s: %v", label, err)
		}
	}

	invalidTargets := map[string]struct {
		wrongType uuid.UUID
		foreign   uuid.UUID
		deleted   uuid.UUID
	}{
		"decision":     {wrongType: taskID, foreign: foreignDecisionID, deleted: deletedDecisionID},
		"task_request": {wrongType: decisionID, foreign: foreignTaskID, deleted: deletedTaskID},
		"evidence":     {wrongType: taskID, foreign: foreignEvidenceID, deleted: deletedEvidenceID},
	}
	for _, field := range []struct {
		viewSchemaID string
		fieldKey     string
		targetType   string
	}{
		{workbook.CommLogViewSchemaID, "comm_log.decision_ids", "decision"},
		{workbook.CommLogViewSchemaID, "comm_log.action_task_ids", "task_request"},
		{workbook.HandoffViewSchemaID, "handoff.open_task_ids", "task_request"},
		{workbook.HandoffViewSchemaID, "handoff.open_decision_ids", "decision"},
		{workbook.StatusReviewViewSchemaID, "status_review.blocked_task_ids", "task_request"},
		{workbook.StatusReviewViewSchemaID, "status_review.pending_evidence_ids", "evidence"},
		{workbook.StatusReviewViewSchemaID, "status_review.open_decision_ids", "decision"},
		{workbook.LessonViewSchemaID, "lesson.follow_up_task_ids", "task_request"},
		{workbook.LessonViewSchemaID, "lesson.evidence_refs", "evidence"},
	} {
		for _, invalid := range []struct {
			name string
			id   uuid.UUID
		}{
			{name: "wrong-type", id: invalidTargets[field.targetType].wrongType},
			{name: "foreign", id: invalidTargets[field.targetType].foreign},
			{name: "deleted", id: invalidTargets[field.targetType].deleted},
		} {
			before := countSprint7DurableState(t, harness, incident.ID)
			clientTxnID := "txn-phase9-sprint7-invalid-" + strings.ReplaceAll(field.fieldKey, ".", "-") + "-" + invalid.name
			_, err := store.CreateWorkbookRow(ctx, actor, incident.ID, workbook.CreateRequest{
				ViewSchemaID: field.viewSchemaID,
				ClientTxnID:  clientTxnID,
				Values:       sprint7MinimumValues(actor.ID, field.viewSchemaID),
				Collections:  map[string]workbook.CollectionActionPayload{field.fieldKey: sprint6Collection(addSprint6RecordRef(invalid.id))},
			}, []byte(clientTxnID), "req-"+clientTxnID, sprint7Time(2*time.Hour))
			requireSprint7MutationValidation(t, err, field.fieldKey, "invalid_value")
			requireSprint7DurableState(t, harness, incident.ID, before, field.fieldKey+" "+invalid.name)
		}
	}

	for _, fieldKey := range []string{"comm_log.audience_party_ids", "comm_log.attendee_party_ids"} {
		for _, invalid := range []struct {
			name string
			id   uuid.UUID
		}{
			{name: "wrong-type", id: taskID},
			{name: "foreign", id: foreignPartyID},
			{name: "deleted", id: deletedPartyID},
		} {
			before := countSprint7DurableState(t, harness, incident.ID)
			clientTxnID := "txn-phase9-sprint7-invalid-" + strings.ReplaceAll(fieldKey, ".", "-") + "-" + invalid.name
			_, err := store.CreateWorkbookRow(ctx, actor, incident.ID, workbook.CreateRequest{
				ViewSchemaID: workbook.CommLogViewSchemaID,
				ClientTxnID:  clientTxnID,
				Values:       sprint7MinimumValues(actor.ID, workbook.CommLogViewSchemaID),
				Collections:  map[string]workbook.CollectionActionPayload{fieldKey: sprint6Collection(workbook.CollectionAction{Op: "add_party_ref", PartyID: &invalid.id})},
			}, []byte(clientTxnID), "req-"+clientTxnID, sprint7Time(3*time.Hour))
			requireSprint7MutationValidation(t, err, fieldKey, "invalid_value")
			requireSprint7DurableState(t, harness, incident.ID, before, fieldKey+" "+invalid.name)
		}
	}

	comm := mustCreateSprint7Row(t, store, actor, incident.ID, workbook.CommLogViewSchemaID, "txn-phase9-sprint7-comm-collections", map[string]workbook.ValueChange{
		"comm_log.comm_type":          sprint7Text("briefing"),
		"comm_log.audience":           sprint7Text("Leadership source text"),
		"comm_log.channel_or_meeting": sprint7Text("Bridge"),
		"comm_log.summary":            sprint7Text("Collection coordination update"),
	}, map[string]workbook.CollectionActionPayload{
		"comm_log.decision_ids":       sprint6Collection(addSprint6RecordRef(decisionID), addSprint6RecordRef(decisionID)),
		"comm_log.action_task_ids":    sprint6Collection(addSprint6RecordRef(taskID), addSprint6RecordRef(taskID)),
		"comm_log.audience_party_ids": sprint6Collection(workbook.CollectionAction{Op: "add_party_ref", PartyID: &partyID}, workbook.CollectionAction{Op: "add_party_ref", PartyID: &partyID}),
	}, sprint7Time(4*time.Hour))
	commRow := comm.Payload["row"].(map[string]any)
	requireSprint7CellValue(t, commRow, "comm_log.audience", "Leadership source text")
	requireSprint7CollectionItemCount(t, commRow, "comm_log.decision_ids", 1)
	requireSprint7CollectionItemKind(t, commRow, "comm_log.decision_ids", "record_ref")
	requireSprint7CollectionItemCount(t, commRow, "comm_log.action_task_ids", 1)
	requireSprint7CollectionItemCount(t, commRow, "comm_log.audience_party_ids", 1)
	requireSprint7CollectionItemKind(t, commRow, "comm_log.audience_party_ids", "party_ref")
	requireSprint7ManualReferenceLink(t, harness, comm.RecordID, decisionID, "comm_log.decision_ids", "references_record")
	requireSprint7ManualReferenceLink(t, harness, comm.RecordID, taskID, "comm_log.action_task_ids", "references_record")
	requireSprint7ManualReferenceLink(t, harness, comm.RecordID, partyID, "comm_log.audience_party_ids", "references_record")

	commWithAttendee := mustSprint6Patch(t, store, actor, comm.RecordID, workbook.CommLogViewSchemaID, 1, "txn-phase9-sprint7-comm-add-attendee",
		sprint6CollectionChange("comm_log.attendee_party_ids", sprint6Collection(workbook.CollectionAction{Op: "add_party_ref", PartyID: &otherPartyID})))
	requireSprint7CellValue(t, commWithAttendee.Payload["row"].(map[string]any), "comm_log.audience", "Leadership source text")
	commWithoutAttendee := mustSprint6Patch(t, store, actor, comm.RecordID, workbook.CommLogViewSchemaID, 2, "txn-phase9-sprint7-comm-remove-attendee",
		sprint6CollectionChange("comm_log.attendee_party_ids", sprint6Collection(workbook.CollectionAction{Op: "remove_party_ref", ItemRef: "party_ref:" + otherPartyID.String()})))
	requireSprint7CellValue(t, commWithoutAttendee.Payload["row"].(map[string]any), "comm_log.audience", "Leadership source text")

	handoff := mustCreateSprint7Row(t, store, actor, incident.ID, workbook.HandoffViewSchemaID, "txn-phase9-sprint7-handoff-collections", map[string]workbook.ValueChange{
		"handoff.incoming_owner_user_id": sprint7UUID(actor.ID),
		"handoff.current_state_summary":  sprint7Text("Open coordination work"),
	}, map[string]workbook.CollectionActionPayload{
		"handoff.open_task_ids":     sprint6Collection(addSprint6RecordRef(taskID)),
		"handoff.open_decision_ids": sprint6Collection(addSprint6RecordRef(decisionID)),
		"handoff.open_risk_refs": {
			Actions: []workbook.CollectionAction{
				{Op: "add_risk_ref", RiskRefText: " Pending outbound access review ", NormalizedText: "pending outbound access review"},
				{Op: "add_risk_ref", RiskRefText: "Pending outbound access review", NormalizedText: "pending outbound access review"},
			},
		},
	}, sprint7Time(5*time.Hour))
	handoffRow := handoff.Payload["row"].(map[string]any)
	requireSprint7CollectionItemCount(t, handoffRow, "handoff.open_task_ids", 1)
	requireSprint7CollectionItemCount(t, handoffRow, "handoff.open_decision_ids", 1)
	requireSprint7CollectionItemCount(t, handoffRow, "handoff.open_risk_refs", 1)
	requireSprint7CollectionItemKind(t, handoffRow, "handoff.open_risk_refs", "risk_ref")
	riskRef := requireSprint7SingleItemRef(t, handoffRow, "handoff.open_risk_refs", "risk_ref:")
	if got := countSprint7ActiveRiskRefs(t, harness, handoff.RecordID); got != 1 {
		t.Fatalf("handoff duplicate risk refs: got %d want 1", got)
	}
	if got := countSprint7LinksForField(t, harness, handoff.RecordID, "handoff.open_risk_refs"); got != 0 {
		t.Fatalf("risk refs must not create generic record links, got %d", got)
	}
	secondHandoff := mustCreateSprint7Row(t, store, actor, incident.ID, workbook.HandoffViewSchemaID, "txn-phase9-sprint7-handoff-risk-scoped", map[string]workbook.ValueChange{
		"handoff.incoming_owner_user_id": sprint7UUID(actor.ID),
		"handoff.current_state_summary":  sprint7Text("Separate handoff same risk text"),
	}, map[string]workbook.CollectionActionPayload{
		"handoff.open_risk_refs": {Actions: []workbook.CollectionAction{{Op: "add_risk_ref", RiskRefText: "Pending outbound access review", NormalizedText: "pending outbound access review"}}},
	}, sprint7Time(6*time.Hour))
	secondRiskRef := requireSprint7SingleItemRef(t, secondHandoff.Payload["row"].(map[string]any), "handoff.open_risk_refs", "risk_ref:")
	if got := countSprint7ActiveRiskRefs(t, harness, secondHandoff.RecordID); got != 1 {
		t.Fatalf("same risk text must be scoped per handoff: got %d want 1", got)
	}
	removedRisk := mustSprint6Patch(t, store, actor, handoff.RecordID, workbook.HandoffViewSchemaID, 1, "txn-phase9-sprint7-handoff-remove-risk",
		sprint6CollectionChange("handoff.open_risk_refs", workbook.CollectionActionPayload{Actions: []workbook.CollectionAction{{Op: "remove_risk_ref", ItemRef: riskRef}}}))
	requireSprint7CollectionItemCount(t, removedRisk.Payload["row"].(map[string]any), "handoff.open_risk_refs", 0)
	if got := countSprint7ActiveRiskRefs(t, harness, secondHandoff.RecordID); got != 1 {
		t.Fatalf("removing one handoff risk ref affected another handoff: got %d want 1", got)
	}

	status := mustCreateSprint7Row(t, store, actor, incident.ID, workbook.StatusReviewViewSchemaID, "txn-phase9-sprint7-status-collections", map[string]workbook.ValueChange{
		"status_review.current_state_summary": sprint7Text("Status with open coordination work"),
	}, map[string]workbook.CollectionActionPayload{
		"status_review.blocked_task_ids":     sprint6Collection(addSprint6RecordRef(taskID)),
		"status_review.pending_evidence_ids": sprint6Collection(addSprint6RecordRef(evidenceID)),
		"status_review.open_decision_ids":    sprint6Collection(addSprint6RecordRef(decisionID)),
	}, sprint7Time(7*time.Hour))
	statusRow := status.Payload["row"].(map[string]any)
	requireSprint7CollectionItemCount(t, statusRow, "status_review.blocked_task_ids", 1)
	requireSprint7CollectionItemCount(t, statusRow, "status_review.pending_evidence_ids", 1)
	requireSprint7CollectionItemCount(t, statusRow, "status_review.open_decision_ids", 1)

	lesson := mustCreateSprint7Row(t, store, actor, incident.ID, workbook.LessonViewSchemaID, "txn-phase9-sprint7-lesson-collections", map[string]workbook.ValueChange{
		"lesson.summary": sprint7Text("Follow up on evidence and work"),
	}, map[string]workbook.CollectionActionPayload{
		"lesson.follow_up_task_ids": sprint6Collection(addSprint6RecordRef(taskID)),
		"lesson.evidence_refs":      sprint6Collection(addSprint6RecordRef(evidenceID)),
	}, sprint7Time(8*time.Hour))
	lessonRow := lesson.Payload["row"].(map[string]any)
	requireSprint7CollectionItemCount(t, lessonRow, "lesson.follow_up_task_ids", 1)
	requireSprint7CollectionItemCount(t, lessonRow, "lesson.evidence_refs", 1)

	for _, field := range []struct {
		viewSchemaID string
		recordID     uuid.UUID
		fieldKey     string
		targetType   string
	}{
		{workbook.CommLogViewSchemaID, comm.RecordID, "comm_log.decision_ids", "decision"},
		{workbook.CommLogViewSchemaID, comm.RecordID, "comm_log.action_task_ids", "task_request"},
		{workbook.HandoffViewSchemaID, handoff.RecordID, "handoff.open_task_ids", "task_request"},
		{workbook.HandoffViewSchemaID, handoff.RecordID, "handoff.open_decision_ids", "decision"},
		{workbook.StatusReviewViewSchemaID, status.RecordID, "status_review.blocked_task_ids", "task_request"},
		{workbook.StatusReviewViewSchemaID, status.RecordID, "status_review.pending_evidence_ids", "evidence"},
		{workbook.StatusReviewViewSchemaID, status.RecordID, "status_review.open_decision_ids", "decision"},
		{workbook.LessonViewSchemaID, lesson.RecordID, "lesson.follow_up_task_ids", "task_request"},
		{workbook.LessonViewSchemaID, lesson.RecordID, "lesson.evidence_refs", "evidence"},
	} {
		for _, invalid := range []struct {
			name    string
			itemRef string
		}{
			{name: "wrong-type", itemRef: "record_ref:" + invalidTargets[field.targetType].wrongType.String()},
			{name: "foreign", itemRef: "record_ref:" + invalidTargets[field.targetType].foreign.String()},
			{name: "deleted", itemRef: "record_ref:" + invalidTargets[field.targetType].deleted.String()},
		} {
			requireSprint7InvalidCollectionPatch(t, harness, store, actor, incident.ID, field.recordID, field.viewSchemaID, field.fieldKey, sprint6Collection(workbook.CollectionAction{Op: "remove_record_ref", ItemRef: invalid.itemRef}), field.fieldKey+" remove "+invalid.name)
		}
	}

	for _, field := range []struct {
		recordID uuid.UUID
		fieldKey string
	}{
		{comm.RecordID, "comm_log.audience_party_ids"},
		{comm.RecordID, "comm_log.attendee_party_ids"},
	} {
		for _, invalid := range []struct {
			name    string
			itemRef string
		}{
			{name: "wrong-type", itemRef: "party_ref:" + taskID.String()},
			{name: "foreign", itemRef: "party_ref:" + foreignPartyID.String()},
			{name: "deleted", itemRef: "party_ref:" + deletedPartyID.String()},
		} {
			requireSprint7InvalidCollectionPatch(t, harness, store, actor, incident.ID, field.recordID, workbook.CommLogViewSchemaID, field.fieldKey, sprint6Collection(workbook.CollectionAction{Op: "remove_party_ref", ItemRef: invalid.itemRef}), field.fieldKey+" remove "+invalid.name)
		}
	}

	for _, invalid := range []struct {
		name    string
		itemRef string
	}{
		{name: "foreign-source", itemRef: secondRiskRef},
		{name: "deleted", itemRef: riskRef},
		{name: "invalid", itemRef: "risk_ref:00000000-0000-0000-0000-000000000000"},
	} {
		requireSprint7InvalidCollectionPatch(t, harness, store, actor, incident.ID, handoff.RecordID, workbook.HandoffViewSchemaID, "handoff.open_risk_refs", sprint6Collection(workbook.CollectionAction{Op: "remove_risk_ref", ItemRef: invalid.itemRef}), "handoff.open_risk_refs remove "+invalid.name)
	}
}

func mustCreateSprint7Row(t testing.TB, store *workbook.Store, actor authn.UserRecord, incidentID uuid.UUID, viewSchemaID string, clientTxnID string, values map[string]workbook.ValueChange, collections map[string]workbook.CollectionActionPayload, now time.Time) workbook.MutationResult {
	t.Helper()
	result, err := store.CreateWorkbookRow(context.Background(), actor, incidentID, workbook.CreateRequest{
		ViewSchemaID: viewSchemaID,
		ClientTxnID:  clientTxnID,
		Values:       values,
		Collections:  collections,
	}, []byte(clientTxnID), "req-"+clientTxnID, now)
	if err != nil {
		t.Fatalf("create %s: %v", clientTxnID, err)
	}
	return result
}

func sprint7Text(value string) workbook.ValueChange {
	return workbook.ValueChange{Kind: "text", Text: stringPtrU911(value)}
}

func sprint7UUID(value uuid.UUID) workbook.ValueChange {
	return workbook.ValueChange{Kind: "uuid", UUID: &value}
}

func sprint7Timestamp(value time.Time) workbook.ValueChange {
	value = value.UTC()
	return workbook.ValueChange{Kind: "timestamp", Timestamp: &value}
}

func sprint7StringPtr(value string) *string {
	return &value
}

func sprint7Time(offset time.Duration) time.Time {
	return time.Date(2026, 5, 19, 15, 0, 0, 0, time.UTC).Add(offset)
}

func sprint7MinimumValues(actorID uuid.UUID, viewSchemaID string) map[string]workbook.ValueChange {
	switch viewSchemaID {
	case workbook.CommLogViewSchemaID:
		return map[string]workbook.ValueChange{
			"comm_log.comm_type":          sprint7Text("briefing"),
			"comm_log.audience":           sprint7Text("Audience text"),
			"comm_log.channel_or_meeting": sprint7Text("Bridge"),
			"comm_log.summary":            sprint7Text("Coordination update"),
		}
	case workbook.HandoffViewSchemaID:
		return map[string]workbook.ValueChange{
			"handoff.incoming_owner_user_id": sprint7UUID(actorID),
			"handoff.current_state_summary":  sprint7Text("Handoff state"),
		}
	case workbook.StatusReviewViewSchemaID:
		return map[string]workbook.ValueChange{
			"status_review.current_state_summary": sprint7Text("Status review state"),
		}
	case workbook.LessonViewSchemaID:
		return map[string]workbook.ValueChange{
			"lesson.summary": sprint7Text("Lesson summary"),
		}
	default:
		return map[string]workbook.ValueChange{}
	}
}

type sprint7DurableState struct {
	Records     int
	Artifacts   int
	RecordLinks int
	RiskRefs    int
}

func countSprint7DurableState(t testing.TB, harness *recordstoretest.StoreHarness, incidentID uuid.UUID) sprint7DurableState {
	t.Helper()
	var state sprint7DurableState
	if err := harness.DB.QueryRow(context.Background(), `
SELECT
    (SELECT COUNT(*) FROM records WHERE incident_id = $1),
    (SELECT COUNT(*) FROM artifacts WHERE incident_id = $1),
    (SELECT COUNT(*) FROM record_links WHERE incident_id = $1),
    (SELECT COUNT(*) FROM handoff_risk_refs WHERE incident_id = $1)
`, incidentID).Scan(&state.Records, &state.Artifacts, &state.RecordLinks, &state.RiskRefs); err != nil {
		t.Fatalf("count durable state: %v", err)
	}
	return state
}

func requireSprint7DurableState(t testing.TB, harness *recordstoretest.StoreHarness, incidentID uuid.UUID, want sprint7DurableState, context string) {
	t.Helper()
	got := countSprint7DurableState(t, harness, incidentID)
	if got != want {
		t.Fatalf("%s changed durable state: got %#v want %#v", context, got, want)
	}
}

func countSprint7SQLDurableState(t testing.TB, db *sql.DB, incidentID string) sprint7DurableState {
	t.Helper()
	var state sprint7DurableState
	if err := db.QueryRowContext(context.Background(), `
SELECT
    (SELECT COUNT(*) FROM records WHERE incident_id = $1),
    (SELECT COUNT(*) FROM artifacts WHERE incident_id = $1),
    (SELECT COUNT(*) FROM record_links WHERE incident_id = $1),
    (SELECT COUNT(*) FROM handoff_risk_refs WHERE incident_id = $1)
`, incidentID).Scan(&state.Records, &state.Artifacts, &state.RecordLinks, &state.RiskRefs); err != nil {
		t.Fatalf("count durable SQL state: %v", err)
	}
	return state
}

func requireSprint7SQLDurableState(t testing.TB, db *sql.DB, incidentID string, want sprint7DurableState, context string) {
	t.Helper()
	got := countSprint7SQLDurableState(t, db, incidentID)
	if got != want {
		t.Fatalf("%s changed durable SQL state: got %#v want %#v", context, got, want)
	}
}

func countSprint7SQLProjectionRows(t testing.TB, db *sql.DB, incidentID string, artifactType string) int {
	t.Helper()
	var count int
	if err := db.QueryRowContext(context.Background(), `
SELECT COUNT(*)
  FROM artifact_grid_projection
 WHERE incident_id = $1
   AND artifact_type = $2
`, incidentID, artifactType).Scan(&count); err != nil {
		t.Fatalf("count projection rows for %s: %v", artifactType, err)
	}
	return count
}

func requireSprint7ArtifactType(t testing.TB, harness *recordstoretest.StoreHarness, recordID uuid.UUID, artifactType string) {
	t.Helper()
	var recordType, gotArtifactType string
	if err := harness.DB.QueryRow(context.Background(), `
SELECT r.record_type, a.artifact_type
  FROM records r
  JOIN artifacts a ON a.incident_id = r.incident_id AND a.record_id = r.record_id
 WHERE r.record_id = $1
`, recordID).Scan(&recordType, &gotArtifactType); err != nil {
		t.Fatalf("query artifact type: %v", err)
	}
	if recordType != "artifact" || gotArtifactType != artifactType {
		t.Fatalf("unexpected artifact identity: record_type=%q artifact_type=%q want artifact/%s", recordType, gotArtifactType, artifactType)
	}
}

func requireSprint7MutationValidation(t testing.TB, err error, field string, reason string) {
	t.Helper()
	var validationErr *workbook.MutationValidationError
	if !errors.As(err, &validationErr) {
		t.Fatalf("expected mutation validation error, got %v", err)
	}
	if validationErr.Field != field || validationErr.ReasonCode != reason {
		t.Fatalf("unexpected validation error: %#v", validationErr)
	}
}

func requireSprint7InvalidCollectionPatch(t testing.TB, harness *recordstoretest.StoreHarness, store *workbook.Store, actor authn.UserRecord, incidentID uuid.UUID, recordID uuid.UUID, viewSchemaID string, fieldKey string, collection workbook.CollectionActionPayload, context string) {
	t.Helper()
	before := countSprint7DurableState(t, harness, incidentID)
	baseVersion := sprint6RecordVersion(t, harness.DB, recordID)
	_, err := sprint6Patch(store, actor, recordID, viewSchemaID, baseVersion, "txn-phase9-sprint7-invalid-"+strings.ReplaceAll(context, " ", "-"), sprint6CollectionChange(fieldKey, collection))
	requireSprint7MutationValidation(t, err, fieldKey, "invalid_value")
	requireSprint7DurableState(t, harness, incidentID, before, context)
	if got := sprint6RecordVersion(t, harness.DB, recordID); got != baseVersion {
		t.Fatalf("%s changed row version: got %d want %d", context, got, baseVersion)
	}
}

func requireSprint7NoRecordChanged(t testing.TB, client *incidentwstest.Client, timeout time.Duration) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for {
		remaining := time.Until(deadline)
		if remaining <= 0 {
			return
		}
		message, err := client.AwaitNextMessage(remaining)
		if err != nil {
			if errors.Is(err, context.DeadlineExceeded) {
				return
			}
			t.Fatalf("wait for no record_changed: %v", err)
		}
		if message.Type == "record_changed" {
			t.Fatalf("unexpected record_changed after rejected create: %s", string(message.Payload))
		}
	}
}

func expectSprint7DecodePatchRejected(t testing.TB, viewSchemaID string, fieldKey string, value any) {
	t.Helper()
	payload, err := json.Marshal(map[string]any{
		"view_schema_id":   viewSchemaID,
		"base_row_version": 1,
		"client_txn_id":    "txn-phase9-sprint7-decode-reject-" + fieldKey,
		"changes": []map[string]any{
			{
				"field_key": fieldKey,
				"value":     value,
			},
		},
	})
	if err != nil {
		t.Fatalf("marshal patch payload: %v", err)
	}
	if _, apiErr := workbook.DecodePatchRequest(strings.NewReader(string(payload))); apiErr == nil {
		t.Fatalf("expected patch decode to reject %s", fieldKey)
	}
}

func requireSprint7CellValue(t testing.TB, row map[string]any, fieldKey string, want any) {
	t.Helper()
	got := row["cells"].(map[string]any)[fieldKey].(map[string]any)["value"]
	if got != want {
		t.Fatalf("unexpected %s value: got %#v want %#v", fieldKey, got, want)
	}
}

func requireSprint7CellNonEmpty(t testing.TB, row map[string]any, fieldKey string) {
	t.Helper()
	got := row["cells"].(map[string]any)[fieldKey].(map[string]any)["value"]
	if got == nil || got == "" {
		t.Fatalf("expected non-empty %s value, got %#v", fieldKey, got)
	}
}

func requireSprint7CollectionItemCount(t testing.TB, row map[string]any, fieldKey string, want int) {
	t.Helper()
	items := sprint7CollectionItems(t, row, fieldKey)
	if len(items) != want {
		t.Fatalf("unexpected %s item count: got %d want %d items=%#v", fieldKey, len(items), want, items)
	}
}

func requireSprint7CollectionItemKind(t testing.TB, row map[string]any, fieldKey string, itemKind string) {
	t.Helper()
	for _, item := range sprint7CollectionItems(t, row, fieldKey) {
		if item["item_kind"] != itemKind {
			t.Fatalf("unexpected %s item kind: got %#v want %s", fieldKey, item["item_kind"], itemKind)
		}
	}
}

func requireSprint7SingleItemRef(t testing.TB, row map[string]any, fieldKey string, prefix string) string {
	t.Helper()
	items := sprint7CollectionItems(t, row, fieldKey)
	if len(items) != 1 {
		t.Fatalf("expected one %s item, got %#v", fieldKey, items)
	}
	itemRef, ok := items[0]["item_ref"].(string)
	if !ok || !strings.HasPrefix(itemRef, prefix) {
		t.Fatalf("unexpected %s item_ref: %#v", fieldKey, items[0]["item_ref"])
	}
	return itemRef
}

func sprint7CollectionItems(t testing.TB, row map[string]any, fieldKey string) []map[string]any {
	t.Helper()
	value := sprint7CollectionValue(t, row, fieldKey)
	switch items := value["items"].(type) {
	case []map[string]any:
		return items
	case []any:
		result := make([]map[string]any, 0, len(items))
		for _, item := range items {
			itemMap, ok := item.(map[string]any)
			if !ok {
				t.Fatalf("unexpected %s item shape: %#v", fieldKey, item)
			}
			result = append(result, itemMap)
		}
		return result
	default:
		t.Fatalf("unexpected %s items shape: %#v", fieldKey, value["items"])
		return nil
	}
}

func sprint7CollectionValue(t testing.TB, row map[string]any, fieldKey string) map[string]any {
	t.Helper()
	value := row["cells"].(map[string]any)[fieldKey].(map[string]any)["value"].(map[string]any)
	if value["kind"] != "collection_value_v1" {
		t.Fatalf("expected %s to be collection_value_v1, got %#v", fieldKey, value)
	}
	return value
}

func requireSprint7CollectionUnordered(t testing.TB, row map[string]any, fieldKey string) {
	t.Helper()
	value := sprint7CollectionValue(t, row, fieldKey)
	if value["ordered"] != false {
		t.Fatalf("%s ordered: got %#v want false", fieldKey, value["ordered"])
	}
}

func requireSprint7RecordRefItemShape(t testing.TB, row map[string]any, fieldKey string, targetID uuid.UUID) {
	t.Helper()
	requireSprint7CollectionUnordered(t, row, fieldKey)
	items := sprint7CollectionItems(t, row, fieldKey)
	if len(items) != 1 {
		t.Fatalf("expected one %s item, got %#v", fieldKey, items)
	}
	item := items[0]
	requireSprint7ExactItemKeys(t, item, "display_text", "item_kind", "item_ref", "linked_record_id")
	target := targetID.String()
	requireSprint7ItemString(t, item, "item_kind", "record_ref")
	requireSprint7ItemString(t, item, "item_ref", "record_ref:"+target)
	requireSprint7ItemString(t, item, "linked_record_id", target)
	if displayText, ok := item["display_text"].(string); !ok || displayText == "" {
		t.Fatalf("%s display_text must be non-empty string, got %#v", fieldKey, item["display_text"])
	}
}

func requireSprint7PartyRefItemShape(t testing.TB, row map[string]any, fieldKey string, partyID uuid.UUID) {
	t.Helper()
	requireSprint7CollectionUnordered(t, row, fieldKey)
	items := sprint7CollectionItems(t, row, fieldKey)
	if len(items) != 1 {
		t.Fatalf("expected one %s item, got %#v", fieldKey, items)
	}
	item := items[0]
	requireSprint7ExactItemKeys(t, item, "display_text", "item_kind", "item_ref", "party_id")
	target := partyID.String()
	requireSprint7ItemString(t, item, "item_kind", "party_ref")
	requireSprint7ItemString(t, item, "item_ref", "party_ref:"+target)
	requireSprint7ItemString(t, item, "party_id", target)
	if displayText, ok := item["display_text"].(string); !ok || displayText == "" {
		t.Fatalf("%s display_text must be non-empty string, got %#v", fieldKey, item["display_text"])
	}
}

func requireSprint7RiskRefItemShape(t testing.TB, row map[string]any, fieldKey string, riskText string) {
	t.Helper()
	requireSprint7CollectionUnordered(t, row, fieldKey)
	items := sprint7CollectionItems(t, row, fieldKey)
	if len(items) != 1 {
		t.Fatalf("expected one %s item, got %#v", fieldKey, items)
	}
	item := items[0]
	requireSprint7ExactItemKeys(t, item, "display_text", "item_kind", "item_ref", "risk_ref_id", "risk_ref_text")
	requireSprint7ItemString(t, item, "item_kind", "risk_ref")
	requireSprint7ItemString(t, item, "display_text", riskText)
	requireSprint7ItemString(t, item, "risk_ref_text", riskText)
	riskRefID, ok := item["risk_ref_id"].(string)
	if !ok || riskRefID == "" {
		t.Fatalf("%s risk_ref_id must be non-empty string, got %#v", fieldKey, item["risk_ref_id"])
	}
	requireSprint7ItemString(t, item, "item_ref", "risk_ref:"+riskRefID)
}

func requireSprint7ExactItemKeys(t testing.TB, item map[string]any, keys ...string) {
	t.Helper()
	if len(item) != len(keys) {
		t.Fatalf("unexpected item keys: got %#v want %v", item, keys)
	}
	for _, key := range keys {
		if _, ok := item[key]; !ok {
			t.Fatalf("missing item key %s in %#v", key, item)
		}
	}
}

func requireSprint7ItemString(t testing.TB, item map[string]any, key string, want string) {
	t.Helper()
	if got, ok := item[key].(string); !ok || got != want {
		t.Fatalf("unexpected %s: got %#v want %q in %#v", key, item[key], want, item)
	}
}

func requireSprint7FilterMatchesOnly(t testing.TB, store *workbook.Store, incidentID uuid.UUID, viewSchemaID string, filter viewschema.Filter, wantRecordID uuid.UUID) {
	t.Helper()
	rows, err := store.QueryRows(context.Background(), incidentID, viewSchemaID, viewschema.QueryMeta{
		Filters: []viewschema.Filter{filter},
		Sort:    []viewschema.SortEntry{{FieldKey: "record_id", Direction: "asc"}},
	})
	if err != nil {
		t.Fatalf("query %s filter %#v: %v", viewSchemaID, filter, err)
	}
	if len(rows) != 1 || rows[0]["record_id"] != wantRecordID.String() {
		t.Fatalf("query %s filter %#v returned rows %#v, want only %s", viewSchemaID, filter, rows, wantRecordID)
	}
}

func requireSprint7ProjectedRow(t testing.TB, store *workbook.Store, incidentID uuid.UUID, viewSchemaID string, query viewschema.QueryMeta, recordID uuid.UUID, groupField string, groupValue any) {
	t.Helper()
	rows, err := store.QueryRows(context.Background(), incidentID, viewSchemaID, query)
	if err != nil {
		t.Fatalf("query %s projections: %v", viewSchemaID, err)
	}
	if len(rows) != 1 || rows[0]["record_id"] != recordID.String() {
		t.Fatalf("query %s returned rows %#v, want only %s", viewSchemaID, rows, recordID)
	}
	groupValues := rows[0]["group_values"].(map[string]any)
	if groupValues[groupField] != groupValue {
		t.Fatalf("unexpected group value for %s: got %#v want %#v", groupField, groupValues[groupField], groupValue)
	}
}

func requireSprint7RecordOrder(t testing.TB, rows []map[string]any, want []uuid.UUID) {
	t.Helper()
	if len(rows) != len(want) {
		t.Fatalf("unexpected row count: got %d want %d rows=%#v", len(rows), len(want), rows)
	}
	for index, id := range want {
		if rows[index]["record_id"] != id.String() {
			t.Fatalf("unexpected row order at %d: got %v want %s rows=%#v", index, rows[index]["record_id"], id, rows)
		}
	}
}

func requireSprint7ManualReferenceLink(t testing.TB, harness *recordstoretest.StoreHarness, sourceID uuid.UUID, targetID uuid.UUID, fieldKey string, linkType string) {
	t.Helper()
	var provenance string
	var confidence sql.NullInt64
	if err := harness.DB.QueryRow(context.Background(), `
SELECT provenance, confidence
  FROM record_links
 WHERE src_record_id = $1
   AND dst_record_id = $2
   AND field_key = $3
   AND link_type = $4
   AND deleted_at IS NULL
`, sourceID, targetID, fieldKey, linkType).Scan(&provenance, &confidence); err != nil {
		t.Fatalf("query manual reference link %s %s -> %s: %v", fieldKey, sourceID, targetID, err)
	}
	if provenance != "manual" || confidence.Valid {
		t.Fatalf("manual link %s must preserve provenance=manual confidence=NULL, got provenance=%q confidence=%#v", fieldKey, provenance, confidence)
	}
}

func countSprint7ActiveRiskRefs(t testing.TB, harness *recordstoretest.StoreHarness, handoffID uuid.UUID) int {
	t.Helper()
	var count int
	if err := harness.DB.QueryRow(context.Background(), `
SELECT COUNT(*)
  FROM handoff_risk_refs
 WHERE handoff_record_id = $1
   AND deleted_at IS NULL
`, handoffID).Scan(&count); err != nil {
		t.Fatalf("count risk refs: %v", err)
	}
	return count
}

func countSprint7LinksForField(t testing.TB, harness *recordstoretest.StoreHarness, sourceID uuid.UUID, fieldKey string) int {
	t.Helper()
	var count int
	if err := harness.DB.QueryRow(context.Background(), `
SELECT COUNT(*)
  FROM record_links
 WHERE src_record_id = $1
   AND field_key = $2
   AND deleted_at IS NULL
`, sourceID, fieldKey).Scan(&count); err != nil {
		t.Fatalf("count links for %s: %v", fieldKey, err)
	}
	return count
}
