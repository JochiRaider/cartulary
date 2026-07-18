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

func TestCoordinationMinimumDefaultsAndRejection_Unit(t *testing.T) {
	ctx := context.Background()
	harness := recordstoretest.StartStore(t, "phase9-sprint7-coordination-defaults")
	store := workbook.NewStore(harness.DB)
	actor := recordstoretest.SeedLocalUserFlags(t, harness.DB, "sprint7-coordination@example.test", "Sprint7 Coordination", "Sprint7Coordination1!", false, false, true)
	incident := recordstoretest.CreateIncidentInStore(t, harness.DB, actor, "txn-phase9-sprint7-coordination-incident", "IR-S7-COORD", "Workbook inspector Sprint 7 coordination defaults")

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
				"comm_log.comm_type":          Text("briefing"),
				"comm_log.audience":           Text("leadership"),
				"comm_log.channel_or_meeting": Text("Bridge"),
			},
			wantField: "comm_log.summary",
		},
		{
			name:         "handoff-missing-summary",
			viewSchemaID: workbook.HandoffViewSchemaID,
			values: map[string]workbook.ValueChange{
				"handoff.incoming_owner_user_id": UUID(actor.ID),
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
		before := countDurableState(t, harness, incident.ID)
		_, err := store.CreateWorkbookRow(ctx, actor, incident.ID, workbook.CreateRequest{
			ViewSchemaID: tc.viewSchemaID,
			ClientTxnID:  "txn-phase9-sprint7-" + tc.name,
			Values:       tc.values,
		}, []byte("txn-phase9-sprint7-"+tc.name), "req-phase9-sprint7-"+tc.name, Time(0))
		requireMutationValidation(t, err, tc.wantField, "missing_required_field")
		requireDurableState(t, harness, incident.ID, before, tc.name)
	}

	expectDecodePatchRejected(t, workbook.CommLogViewSchemaID, "comm_log.comm_id", "client-supplied")
	expectDecodePatchRejected(t, workbook.HandoffViewSchemaID, "handoff.handoff_id", "client-supplied")
	expectDecodePatchRejected(t, workbook.StatusReviewViewSchemaID, "status_review.status_review_id", "client-supplied")
	expectDecodePatchRejected(t, workbook.LessonViewSchemaID, "lesson.lesson_id", "client-supplied")

	comm := mustCreateRow(t, store, actor, incident.ID, workbook.CommLogViewSchemaID, "txn-phase9-sprint7-comm-defaults", map[string]workbook.ValueChange{
		"comm_log.comm_type":          Text("briefing"),
		"comm_log.audience":           Text("Leadership Team"),
		"comm_log.channel_or_meeting": Text("Bridge"),
		"comm_log.summary":            Text("Daily coordination update"),
	}, nil, Time(0))
	commRow := comm.Payload["row"].(map[string]any)
	requireArtifactType(t, harness, comm.RecordID, "comm_log")
	requireCellNonEmpty(t, commRow, "comm_log.comm_id")
	requireCellNonEmpty(t, commRow, "comm_log.timestamp_utc")
	requireCoordinationCellValue(t, commRow, "comm_log.next_report_at", nil)
	requireCoordinationCellValue(t, commRow, "comm_log.privilege_tag", nil)
	requireCoordinationCollectionItemCount(t, commRow, "comm_log.decision_ids", 0)
	requireCoordinationCollectionItemCount(t, commRow, "comm_log.action_task_ids", 0)
	requireCoordinationCollectionItemCount(t, commRow, "comm_log.audience_party_ids", 0)
	requireCoordinationCollectionItemCount(t, commRow, "comm_log.attendee_party_ids", 0)
	beforeCommVersion := RecordVersion(t, harness.DB, comm.RecordID)
	_, err := Patch(store, actor, comm.RecordID, workbook.CommLogViewSchemaID, beforeCommVersion, "txn-phase9-sprint7-comm-invalid-type",
		ValueChange("comm_log.comm_type", workbook.ValueChange{Kind: "text", Text: stringPtr("emergency")}))
	requireMutationValidation(t, err, "comm_log.comm_type", "invalid_value")
	if got := RecordVersion(t, harness.DB, comm.RecordID); got != beforeCommVersion {
		t.Fatalf("invalid comm type changed row version: got %d want %d", got, beforeCommVersion)
	}

	handoff := mustCreateRow(t, store, actor, incident.ID, workbook.HandoffViewSchemaID, "txn-phase9-sprint7-handoff-defaults", map[string]workbook.ValueChange{
		"handoff.incoming_owner_user_id": UUID(actor.ID),
		"handoff.current_state_summary":  Text("Night shift owns containment"),
	}, nil, Time(time.Hour))
	handoffRow := handoff.Payload["row"].(map[string]any)
	requireArtifactType(t, harness, handoff.RecordID, "handoff")
	requireCellNonEmpty(t, handoffRow, "handoff.handoff_id")
	requireCellNonEmpty(t, handoffRow, "handoff.timestamp_utc")
	requireCoordinationCellValue(t, handoffRow, "handoff.outgoing_owner_user_id", actor.ID.String())
	requireCoordinationCellValue(t, handoffRow, "handoff.acknowledged_at", nil)
	requireCoordinationCellValue(t, handoffRow, "handoff.ack_state", "pending")
	requireCoordinationCellValue(t, handoffRow, "handoff.next_checks", nil)
	requireCoordinationCollectionItemCount(t, handoffRow, "handoff.open_task_ids", 0)
	requireCoordinationCollectionItemCount(t, handoffRow, "handoff.open_decision_ids", 0)
	requireCoordinationCollectionItemCount(t, handoffRow, "handoff.open_risk_refs", 0)
	acknowledgedAt := Time(2 * time.Hour)
	handoffAck := mustPatch(t, store, actor, handoff.RecordID, workbook.HandoffViewSchemaID, 1, "txn-phase9-sprint7-handoff-ack",
		ValueChange("handoff.acknowledged_at", workbook.ValueChange{Kind: "timestamp", Timestamp: &acknowledgedAt}))
	requireCoordinationCellValue(t, handoffAck.Payload["row"].(map[string]any), "handoff.ack_state", "acknowledged")
	handoffClear := mustPatch(t, store, actor, handoff.RecordID, workbook.HandoffViewSchemaID, 2, "txn-phase9-sprint7-handoff-clear-ack",
		ValueChange("handoff.acknowledged_at", workbook.ValueChange{Kind: "null"}))
	requireCoordinationCellValue(t, handoffClear.Payload["row"].(map[string]any), "handoff.ack_state", "pending")

	status := mustCreateRow(t, store, actor, incident.ID, workbook.StatusReviewViewSchemaID, "txn-phase9-sprint7-status-defaults", map[string]workbook.ValueChange{
		"status_review.current_state_summary": Text("Containment is stable"),
	}, nil, Time(2*time.Hour))
	statusRow := status.Payload["row"].(map[string]any)
	requireArtifactType(t, harness, status.RecordID, "status_review")
	requireCellNonEmpty(t, statusRow, "status_review.status_review_id")
	requireCellNonEmpty(t, statusRow, "status_review.timestamp_utc")
	requireCoordinationCellValue(t, statusRow, "status_review.review_owner_user_id", actor.ID.String())
	requireCoordinationCellValue(t, statusRow, "status_review.next_report_at", nil)
	requireCoordinationCellValue(t, statusRow, "status_review.active_risks_summary", nil)
	requireCoordinationCollectionItemCount(t, statusRow, "status_review.blocked_task_ids", 0)
	requireCoordinationCollectionItemCount(t, statusRow, "status_review.pending_evidence_ids", 0)
	requireCoordinationCollectionItemCount(t, statusRow, "status_review.open_decision_ids", 0)

	lesson := mustCreateRow(t, store, actor, incident.ID, workbook.LessonViewSchemaID, "txn-phase9-sprint7-lesson-defaults", map[string]workbook.ValueChange{
		"lesson.summary": Text("Preserve VPN logs earlier"),
	}, nil, Time(3*time.Hour))
	lessonRow := lesson.Payload["row"].(map[string]any)
	requireArtifactType(t, harness, lesson.RecordID, "lesson")
	requireCellNonEmpty(t, lessonRow, "lesson.lesson_id")
	requireCellNonEmpty(t, lessonRow, "lesson.timestamp_utc")
	requireCoordinationCellValue(t, lessonRow, "lesson.owner_user_id", actor.ID.String())
	requireCoordinationCellValue(t, lessonRow, "lesson.closure_state", "open")
	requireCoordinationCollectionItemCount(t, lessonRow, "lesson.follow_up_task_ids", 0)
	requireCoordinationCollectionItemCount(t, lessonRow, "lesson.evidence_refs", 0)
	beforeLessonVersion := RecordVersion(t, harness.DB, lesson.RecordID)
	_, err = Patch(store, actor, lesson.RecordID, workbook.LessonViewSchemaID, beforeLessonVersion, "txn-phase9-sprint7-lesson-invalid-closure",
		ValueChange("lesson.closure_state", workbook.ValueChange{Kind: "text", Text: stringPtr("archived")}))
	requireMutationValidation(t, err, "lesson.closure_state", "invalid_value")
	if got := RecordVersion(t, harness.DB, lesson.RecordID); got != beforeLessonVersion {
		t.Fatalf("invalid lesson closure state changed row version: got %d want %d", got, beforeLessonVersion)
	}
}

func TestRejectedCoordinationCreateEmitsNoRecordChanged_Unit(t *testing.T) {
	harness := recordstoretest.StartServer(t, "phase9-sprint7-rejected-create-ws")
	login, _ := recordstoretest.ProvisionBootstrapAdmin(t, harness.Server)
	incident := recordstoretest.CreateIncident(t, harness.Server, login, map[string]any{
		"client_txn_id": "txn-phase9-sprint7-rejected-ws-incident",
		"incident_key":  "IR-S7-REJECTED-WS",
		"title":         "Workbook inspector Sprint 7 rejected create websocket",
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
			before := countSQLDurableState(t, harness.DB, incidentID)
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
			requireSQLDurableState(t, harness.DB, incidentID, before, "rejected zero-field create route")
			if got := countSQLProjectionRows(t, harness.DB, incidentID, tc.artifactType); got != 0 {
				t.Fatalf("%s rejected zero-field create left %d projection rows", tc.viewSchemaID, got)
			}
			requireNoRecordChanged(t, socket, 300*time.Millisecond)
		})
	}
}

func TestCoordinationSavedViewsRemainAdditive_Unit(t *testing.T) {
	ctx := context.Background()
	harness := recordstoretest.StartStore(t, "phase9-sprint7-coordination-saved-views")
	actor := recordstoretest.SeedLocalUserFlags(t, harness.DB, "sprint7-saved-views@example.test", "Sprint7 Saved Views", "Sprint7SavedViews1!", false, false, true)
	incident := recordstoretest.CreateIncidentInStore(t, harness.DB, actor, "txn-phase9-sprint7-saved-views-incident", "IR-S7-SAVED-VIEWS", "Workbook inspector Sprint 7 coordination saved views")
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
			createRequest := SavedViewCreateRequest(t, viewSchemaID)
			created, err := savedViewStore.Create(ctx, actor, incident.ID, createRequest, Time(0))
			if err != nil {
				t.Fatalf("create coordination saved view: %v", err)
			}
			if created.ViewSchemaID != viewSchemaID {
				t.Fatalf("saved view stored wrong view_schema_id: got %q want %q", created.ViewSchemaID, viewSchemaID)
			}
			savedViewID := created.SavedViewID.String()

			startup, err := startupStore.Resolve(ctx, incident.ID, actor.ID, "admin", SheetRefJSON(t, "saved_view", savedViewID), Time(1))
			if err != nil {
				t.Fatalf("resolve startup saved view: %v", err)
			}
			if startup.SelectedViewSchemaID == nil || *startup.SelectedViewSchemaID != viewSchemaID {
				t.Fatalf("startup selected wrong view schema: got %v want %q", startup.SelectedViewSchemaID, viewSchemaID)
			}
			requireSheetRefJSON(t, startup.SelectedSheetRef, "saved_view", savedViewID)
			if startup.SelectedSavedView == nil {
				t.Fatalf("startup selected saved view details must be present")
			}
			if startup.SelectedSavedView.ViewSchemaID != viewSchemaID {
				t.Fatalf("startup selected saved view changed identity: got %q want %q", startup.SelectedSavedView.ViewSchemaID, viewSchemaID)
			}
			if _, err := workbookStore.QueryRows(ctx, incident.ID, viewSchemaID, DefaultQueryMeta(t, viewSchemaID)); err != nil {
				t.Fatalf("canonical query changed identity for %s: %v", viewSchemaID, err)
			}
		})
	}
}

func SavedViewCreateRequest(t testing.TB, viewSchemaID string) savedviews.CreateRequest {
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

func SheetRefJSON(t testing.TB, kind string, id string) []byte {
	t.Helper()
	payload, err := json.Marshal(map[string]string{"kind": kind, "id": id})
	if err != nil {
		t.Fatalf("marshal sheet ref: %v", err)
	}
	return payload
}

func requireSheetRefJSON(t testing.TB, raw []byte, wantKind string, wantID string) {
	t.Helper()
	var ref map[string]string
	if err := json.Unmarshal(raw, &ref); err != nil {
		t.Fatalf("decode selected sheet ref: %v", err)
	}
	if ref["kind"] != wantKind || ref["id"] != wantID {
		t.Fatalf("unexpected sheet ref: got %#v want kind=%q id=%q", ref, wantKind, wantID)
	}
}

func DefaultQueryMeta(t testing.TB, viewSchemaID string) viewschema.QueryMeta {
	t.Helper()
	schema, ok := viewschema.Lookup(viewSchemaID)
	if !ok {
		t.Fatalf("missing view schema %s", viewSchemaID)
	}
	return schema.DefaultQueryMeta()
}

func TestCoordinationProjectionSortFilterGroup_Unit(t *testing.T) {
	harness := recordstoretest.StartStore(t, "phase9-sprint7-coordination-projections")
	store := workbook.NewStore(harness.DB)
	actor := recordstoretest.SeedLocalUserFlags(t, harness.DB, "sprint7-projection@example.test", "Sprint7 Projection", "Sprint7Projection1!", false, false, true)
	incident := recordstoretest.CreateIncidentInStore(t, harness.DB, actor, "txn-phase9-sprint7-projection-incident", "IR-S7-PROJECTION", "Workbook inspector Sprint 7 coordination projections")

	commOneNextReport := time.Date(2026, 5, 20, 9, 0, 0, 0, time.UTC)
	_ = mustCreateRow(t, store, actor, incident.ID, workbook.CommLogViewSchemaID, "txn-phase9-sprint7-comm-projection-one", map[string]workbook.ValueChange{
		"comm_log.timestamp_utc":      Timestamp(time.Date(2026, 5, 18, 9, 0, 0, 0, time.UTC)),
		"comm_log.comm_type":          Text("briefing"),
		"comm_log.audience":           Text("Leads"),
		"comm_log.channel_or_meeting": Text("Morning bridge"),
		"comm_log.summary":            Text("Projection briefing"),
		"comm_log.next_report_at":     Timestamp(commOneNextReport),
		"comm_log.privilege_tag":      Text("internal"),
	}, nil, Time(0))
	commTwoNextReport := time.Date(2026, 5, 21, 9, 0, 0, 0, time.UTC)
	commTwo := mustCreateRow(t, store, actor, incident.ID, workbook.CommLogViewSchemaID, "txn-phase9-sprint7-comm-projection-two", map[string]workbook.ValueChange{
		"comm_log.timestamp_utc":      Timestamp(time.Date(2026, 5, 19, 9, 0, 0, 0, time.UTC)),
		"comm_log.comm_type":          Text("notification"),
		"comm_log.audience":           Text("Duty managers"),
		"comm_log.channel_or_meeting": Text("Email"),
		"comm_log.summary":            Text("Projection notification"),
		"comm_log.next_report_at":     Timestamp(commTwoNextReport),
	}, nil, Time(time.Minute))
	requireProjectedRow(t, store, incident.ID, workbook.CommLogViewSchemaID, viewschema.QueryMeta{
		Filters: []viewschema.Filter{
			{FieldKey: "comm_log.comm_type", Op: "eq", Arg: map[string]any{"value": "notification"}},
			{FieldKey: "comm_log.timestamp_day", Op: "eq", Arg: map[string]any{"value": "2026-05-19"}},
			{FieldKey: "comm_log.next_report_day", Op: "eq", Arg: map[string]any{"value": "2026-05-21"}},
		},
		Sort:    []viewschema.SortEntry{{FieldKey: "comm_log.next_report_day", Direction: "desc"}, {FieldKey: "record_id", Direction: "asc"}},
		GroupBy: StringPtr("comm_log.comm_type"),
	}, commTwo.RecordID, "comm_log.comm_type", "notification")

	handoffPending := mustCreateRow(t, store, actor, incident.ID, workbook.HandoffViewSchemaID, "txn-phase9-sprint7-handoff-projection-pending", map[string]workbook.ValueChange{
		"handoff.timestamp_utc":          Timestamp(time.Date(2026, 5, 18, 10, 0, 0, 0, time.UTC)),
		"handoff.incoming_owner_user_id": UUID(actor.ID),
		"handoff.current_state_summary":  Text("Pending projection handoff"),
	}, nil, Time(2*time.Minute))
	acknowledgedAt := time.Date(2026, 5, 19, 11, 0, 0, 0, time.UTC)
	handoffAck := mustCreateRow(t, store, actor, incident.ID, workbook.HandoffViewSchemaID, "txn-phase9-sprint7-handoff-projection-ack", map[string]workbook.ValueChange{
		"handoff.timestamp_utc":          Timestamp(time.Date(2026, 5, 19, 10, 0, 0, 0, time.UTC)),
		"handoff.incoming_owner_user_id": UUID(actor.ID),
		"handoff.current_state_summary":  Text("Acknowledged projection handoff"),
		"handoff.acknowledged_at":        Timestamp(acknowledgedAt),
	}, nil, Time(3*time.Minute))
	requireProjectedRow(t, store, incident.ID, workbook.HandoffViewSchemaID, viewschema.QueryMeta{
		Filters: []viewschema.Filter{
			{FieldKey: "handoff.incoming_owner_user_id", Op: "eq", Arg: map[string]any{"value": actor.ID.String()}},
			{FieldKey: "handoff.ack_state", Op: "eq", Arg: map[string]any{"value": "acknowledged"}},
			{FieldKey: "handoff.timestamp_day", Op: "eq", Arg: map[string]any{"value": "2026-05-19"}},
		},
		Sort:    []viewschema.SortEntry{{FieldKey: "handoff.outgoing_owner_user_id", Direction: "asc"}, {FieldKey: "record_id", Direction: "asc"}},
		GroupBy: StringPtr("handoff.ack_state"),
	}, handoffAck.RecordID, "handoff.ack_state", "acknowledged")
	ackAsc, err := store.QueryRows(context.Background(), incident.ID, workbook.HandoffViewSchemaID, viewschema.QueryMeta{
		Sort: []viewschema.SortEntry{{FieldKey: "handoff.ack_state", Direction: "asc"}, {FieldKey: "record_id", Direction: "asc"}},
	})
	if err != nil {
		t.Fatalf("query handoff ack_state ascending order: %v", err)
	}
	requireRecordOrder(t, ackAsc, []uuid.UUID{handoffPending.RecordID, handoffAck.RecordID})
	ackDesc, err := store.QueryRows(context.Background(), incident.ID, workbook.HandoffViewSchemaID, viewschema.QueryMeta{
		Sort: []viewschema.SortEntry{{FieldKey: "handoff.ack_state", Direction: "desc"}, {FieldKey: "record_id", Direction: "asc"}},
	})
	if err != nil {
		t.Fatalf("query handoff ack_state descending order: %v", err)
	}
	requireRecordOrder(t, ackDesc, []uuid.UUID{handoffAck.RecordID, handoffPending.RecordID})

	_ = mustCreateRow(t, store, actor, incident.ID, workbook.StatusReviewViewSchemaID, "txn-phase9-sprint7-status-projection-one", map[string]workbook.ValueChange{
		"status_review.timestamp_utc":         Timestamp(time.Date(2026, 5, 18, 12, 0, 0, 0, time.UTC)),
		"status_review.current_state_summary": Text("Status review baseline"),
	}, nil, Time(4*time.Minute))
	statusNextReport := time.Date(2026, 5, 22, 12, 0, 0, 0, time.UTC)
	status := mustCreateRow(t, store, actor, incident.ID, workbook.StatusReviewViewSchemaID, "txn-phase9-sprint7-status-projection-two", map[string]workbook.ValueChange{
		"status_review.timestamp_utc":         Timestamp(time.Date(2026, 5, 19, 12, 0, 0, 0, time.UTC)),
		"status_review.review_owner_user_id":  UUID(actor.ID),
		"status_review.current_state_summary": Text("Status review next report"),
		"status_review.next_report_at":        Timestamp(statusNextReport),
	}, nil, Time(5*time.Minute))
	requireProjectedRow(t, store, incident.ID, workbook.StatusReviewViewSchemaID, viewschema.QueryMeta{
		Filters: []viewschema.Filter{
			{FieldKey: "status_review.review_owner_user_id", Op: "eq", Arg: map[string]any{"value": actor.ID.String()}},
			{FieldKey: "status_review.timestamp_day", Op: "eq", Arg: map[string]any{"value": "2026-05-19"}},
			{FieldKey: "status_review.next_report_day", Op: "eq", Arg: map[string]any{"value": "2026-05-22"}},
		},
		Sort:    []viewschema.SortEntry{{FieldKey: "status_review.next_report_day", Direction: "desc"}, {FieldKey: "record_id", Direction: "asc"}},
		GroupBy: StringPtr("status_review.review_owner_user_id"),
	}, status.RecordID, "status_review.review_owner_user_id", actor.ID.String())

	_ = mustCreateRow(t, store, actor, incident.ID, workbook.LessonViewSchemaID, "txn-phase9-sprint7-lesson-projection-open", map[string]workbook.ValueChange{
		"lesson.timestamp_utc": Timestamp(time.Date(2026, 5, 18, 13, 0, 0, 0, time.UTC)),
		"lesson.summary":       Text("Open lesson projection"),
	}, nil, Time(6*time.Minute))
	lesson := mustCreateRow(t, store, actor, incident.ID, workbook.LessonViewSchemaID, "txn-phase9-sprint7-lesson-projection-closed", map[string]workbook.ValueChange{
		"lesson.timestamp_utc": Timestamp(time.Date(2026, 5, 19, 13, 0, 0, 0, time.UTC)),
		"lesson.summary":       Text("Closed lesson projection"),
		"lesson.owner_user_id": UUID(actor.ID),
		"lesson.closure_state": Text("closed"),
	}, nil, Time(7*time.Minute))
	requireProjectedRow(t, store, incident.ID, workbook.LessonViewSchemaID, viewschema.QueryMeta{
		Filters: []viewschema.Filter{
			{FieldKey: "lesson.closure_state", Op: "eq", Arg: map[string]any{"value": "closed"}},
			{FieldKey: "lesson.owner_user_id", Op: "eq", Arg: map[string]any{"value": actor.ID.String()}},
			{FieldKey: "lesson.timestamp_day", Op: "eq", Arg: map[string]any{"value": "2026-05-19"}},
		},
		Sort:    []viewschema.SortEntry{{FieldKey: "lesson.closure_state", Direction: "asc"}, {FieldKey: "record_id", Direction: "asc"}},
		GroupBy: StringPtr("lesson.closure_state"),
	}, lesson.RecordID, "lesson.closure_state", "closed")
}

func TestCoordinationDeclaredQueryFieldsAreMapped_Unit(t *testing.T) {
	harness := recordstoretest.StartStore(t, "phase9-sprint7-coordination-query-fields")
	store := workbook.NewStore(harness.DB)
	actor := recordstoretest.SeedLocalUserFlags(t, harness.DB, "sprint7-query-fields@example.test", "Sprint7 Query Fields", "Sprint7QueryFields1!", false, false, true)
	incident := recordstoretest.CreateIncidentInStore(t, harness.DB, actor, "txn-phase9-sprint7-query-fields-incident", "IR-S7-QUERY-FIELDS", "Workbook inspector Sprint 7 declared query fields")

	mustCreateRow(t, store, actor, incident.ID, workbook.CommLogViewSchemaID, "txn-phase9-sprint7-query-fields-comm", map[string]workbook.ValueChange{
		"comm_log.comm_type":          Text("briefing"),
		"comm_log.audience":           Text("Query field audience"),
		"comm_log.channel_or_meeting": Text("Bridge"),
		"comm_log.summary":            Text("Query field comm log"),
		"comm_log.privilege_tag":      Text("internal"),
	}, nil, Time(0))
	mustCreateRow(t, store, actor, incident.ID, workbook.HandoffViewSchemaID, "txn-phase9-sprint7-query-fields-handoff", map[string]workbook.ValueChange{
		"handoff.incoming_owner_user_id": UUID(actor.ID),
		"handoff.current_state_summary":  Text("Query field handoff"),
	}, nil, Time(time.Minute))
	mustCreateRow(t, store, actor, incident.ID, workbook.StatusReviewViewSchemaID, "txn-phase9-sprint7-query-fields-status", map[string]workbook.ValueChange{
		"status_review.current_state_summary": Text("Query field status review"),
	}, nil, Time(2*time.Minute))
	mustCreateRow(t, store, actor, incident.ID, workbook.LessonViewSchemaID, "txn-phase9-sprint7-query-fields-lesson", map[string]workbook.ValueChange{
		"lesson.summary": Text("Query field lesson"),
	}, nil, Time(3*time.Minute))

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

func TestCoordinationSemanticFilters_Unit(t *testing.T) {
	harness := recordstoretest.StartStore(t, "phase9-sprint7-coordination-semantic-filters")
	store := workbook.NewStore(harness.DB)
	actor := recordstoretest.SeedLocalUserFlags(t, harness.DB, "sprint7-filters@example.test", "Sprint7 Filters", "Sprint7Filters1!", false, false, true)
	alternate := recordstoretest.SeedLocalUserFlags(t, harness.DB, "sprint7-filters-alt@example.test", "Sprint7 Filters Alt", "Sprint7FiltersAlt1!", false, false, true)
	incident := recordstoretest.CreateIncidentInStore(t, harness.DB, actor, "txn-phase9-sprint7-filters-incident", "IR-S7-FILTERS", "Workbook inspector Sprint 7 semantic filters")
	_, err := store.CreateWorkbookRow(context.Background(), actor, incident.ID, workbook.CreateRequest{
		ViewSchemaID: workbook.HandoffViewSchemaID,
		ClientTxnID:  "txn-phase9-sprint7-filter-nonmember-owner",
		Values: map[string]workbook.ValueChange{
			"handoff.incoming_owner_user_id": UUID(alternate.ID),
			"handoff.current_state_summary":  Text("Non-member owner must fail"),
		},
	}, []byte("txn-phase9-sprint7-filter-nonmember-owner"), "req-phase9-sprint7-filter-nonmember-owner", Time(0))
	requireMutationValidation(t, err, "handoff.incoming_owner_user_id", "invalid_value")
	recordstoretest.SeedIncidentMembership(t, harness.DB, incident.ID, alternate.ID, alternate.DisplayName, "editor", actor.ID)

	commPositive := mustCreateRow(t, store, actor, incident.ID, workbook.CommLogViewSchemaID, "txn-phase9-sprint7-filter-comm-positive", map[string]workbook.ValueChange{
		"comm_log.timestamp_utc":      Timestamp(time.Date(2026, 5, 19, 9, 0, 0, 0, time.UTC)),
		"comm_log.comm_type":          Text("notification"),
		"comm_log.audience":           Text("Incident Command"),
		"comm_log.channel_or_meeting": Text("Bridge Alpha"),
		"comm_log.summary":            Text("Semantic filter positive comm log"),
		"comm_log.next_report_at":     Timestamp(time.Date(2026, 5, 22, 9, 0, 0, 0, time.UTC)),
		"comm_log.privilege_tag":      Text("privileged"),
	}, nil, Time(0))
	mustCreateRow(t, store, actor, incident.ID, workbook.CommLogViewSchemaID, "txn-phase9-sprint7-filter-comm-negative", map[string]workbook.ValueChange{
		"comm_log.timestamp_utc":      Timestamp(time.Date(2026, 5, 20, 9, 0, 0, 0, time.UTC)),
		"comm_log.comm_type":          Text("briefing"),
		"comm_log.audience":           Text("Engineering"),
		"comm_log.channel_or_meeting": Text("Email"),
		"comm_log.summary":            Text("Semantic filter negative comm log"),
		"comm_log.next_report_at":     Timestamp(time.Date(2026, 5, 23, 9, 0, 0, 0, time.UTC)),
		"comm_log.privilege_tag":      Text("public"),
	}, nil, Time(time.Minute))

	handoffPositive := mustCreateRow(t, store, actor, incident.ID, workbook.HandoffViewSchemaID, "txn-phase9-sprint7-filter-handoff-positive", map[string]workbook.ValueChange{
		"handoff.timestamp_utc":          Timestamp(time.Date(2026, 5, 19, 10, 0, 0, 0, time.UTC)),
		"handoff.outgoing_owner_user_id": UUID(actor.ID),
		"handoff.incoming_owner_user_id": UUID(alternate.ID),
		"handoff.current_state_summary":  Text("Semantic filter positive handoff"),
		"handoff.acknowledged_at":        Timestamp(time.Date(2026, 5, 19, 11, 0, 0, 0, time.UTC)),
	}, nil, Time(2*time.Minute))
	mustCreateRow(t, store, actor, incident.ID, workbook.HandoffViewSchemaID, "txn-phase9-sprint7-filter-handoff-negative", map[string]workbook.ValueChange{
		"handoff.timestamp_utc":          Timestamp(time.Date(2026, 5, 20, 10, 0, 0, 0, time.UTC)),
		"handoff.outgoing_owner_user_id": UUID(alternate.ID),
		"handoff.incoming_owner_user_id": UUID(actor.ID),
		"handoff.current_state_summary":  Text("Semantic filter negative handoff"),
	}, nil, Time(3*time.Minute))

	statusPositive := mustCreateRow(t, store, actor, incident.ID, workbook.StatusReviewViewSchemaID, "txn-phase9-sprint7-filter-status-positive", map[string]workbook.ValueChange{
		"status_review.timestamp_utc":         Timestamp(time.Date(2026, 5, 19, 12, 0, 0, 0, time.UTC)),
		"status_review.review_owner_user_id":  UUID(actor.ID),
		"status_review.current_state_summary": Text("Semantic filter positive status"),
		"status_review.next_report_at":        Timestamp(time.Date(2026, 5, 22, 12, 0, 0, 0, time.UTC)),
	}, nil, Time(4*time.Minute))
	mustCreateRow(t, store, actor, incident.ID, workbook.StatusReviewViewSchemaID, "txn-phase9-sprint7-filter-status-negative", map[string]workbook.ValueChange{
		"status_review.timestamp_utc":         Timestamp(time.Date(2026, 5, 20, 12, 0, 0, 0, time.UTC)),
		"status_review.review_owner_user_id":  UUID(alternate.ID),
		"status_review.current_state_summary": Text("Semantic filter negative status"),
		"status_review.next_report_at":        Timestamp(time.Date(2026, 5, 23, 12, 0, 0, 0, time.UTC)),
	}, nil, Time(5*time.Minute))

	lessonPositive := mustCreateRow(t, store, actor, incident.ID, workbook.LessonViewSchemaID, "txn-phase9-sprint7-filter-lesson-positive", map[string]workbook.ValueChange{
		"lesson.timestamp_utc": Timestamp(time.Date(2026, 5, 19, 13, 0, 0, 0, time.UTC)),
		"lesson.summary":       Text("Semantic filter positive lesson"),
		"lesson.owner_user_id": UUID(actor.ID),
		"lesson.closure_state": Text("closed"),
	}, nil, Time(6*time.Minute))
	mustCreateRow(t, store, actor, incident.ID, workbook.LessonViewSchemaID, "txn-phase9-sprint7-filter-lesson-negative", map[string]workbook.ValueChange{
		"lesson.timestamp_utc": Timestamp(time.Date(2026, 5, 20, 13, 0, 0, 0, time.UTC)),
		"lesson.summary":       Text("Semantic filter negative lesson"),
		"lesson.owner_user_id": UUID(alternate.ID),
		"lesson.closure_state": Text("open"),
	}, nil, Time(7*time.Minute))

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
			requireFilterMatchesOnly(t, store, incident.ID, tc.viewSchemaID, viewschema.Filter{
				FieldKey: tc.fieldKey,
				Op:       tc.op,
				Arg:      map[string]any{"value": tc.value},
			}, tc.wantRecordID)
		})
	}
}

func TestCoordinationCollectionItemShapes_Unit(t *testing.T) {
	harness := recordstoretest.StartStore(t, "phase9-sprint7-coordination-collection-shapes")
	store := workbook.NewStore(harness.DB)
	actor := recordstoretest.SeedLocalUserFlags(t, harness.DB, "sprint7-shapes@example.test", "Sprint7 Shapes", "Sprint7Shapes1!", false, false, true)
	incident := recordstoretest.CreateIncidentInStore(t, harness.DB, actor, "txn-phase9-sprint7-shapes-incident", "IR-S7-SHAPES", "Workbook inspector Sprint 7 collection item shapes")

	partyID := mustCreatePartyFor(t, store, actor, incident.ID, "txn-phase9-sprint7-shapes-party", "Coordination Shape Party")
	attendeePartyID := mustCreatePartyFor(t, store, actor, incident.ID, "txn-phase9-sprint7-shapes-attendee", "Coordination Shape Attendee")
	decisionID := mustCreateDecision(t, store, actor, incident.ID, "txn-phase9-sprint7-shapes-decision", "approved", "Shape decision")
	taskID := mustCreateTaskFor(t, store, actor, incident.ID, "txn-phase9-sprint7-shapes-task", "Shape task")
	evidenceID := mustCreateEvidenceFor(t, store, actor, incident.ID, "txn-phase9-sprint7-shapes-evidence", "Shape evidence")

	comm := mustCreateRow(t, store, actor, incident.ID, workbook.CommLogViewSchemaID, "txn-phase9-sprint7-shapes-comm", MinimumValues(actor.ID, workbook.CommLogViewSchemaID), map[string]workbook.CollectionActionPayload{
		"comm_log.decision_ids":       Collection(addOptionalSurfaceRecordRef(decisionID)),
		"comm_log.action_task_ids":    Collection(addOptionalSurfaceRecordRef(taskID)),
		"comm_log.audience_party_ids": Collection(workbook.CollectionAction{Op: "add_party_ref", PartyID: &partyID}),
		"comm_log.attendee_party_ids": Collection(workbook.CollectionAction{Op: "add_party_ref", PartyID: &attendeePartyID}),
	}, Time(0))
	commRow := comm.Payload["row"].(map[string]any)
	requireRecordRefItemShape(t, commRow, "comm_log.decision_ids", decisionID)
	requireRecordRefItemShape(t, commRow, "comm_log.action_task_ids", taskID)
	requirePartyRefItemShape(t, commRow, "comm_log.audience_party_ids", partyID)
	requirePartyRefItemShape(t, commRow, "comm_log.attendee_party_ids", attendeePartyID)

	handoff := mustCreateRow(t, store, actor, incident.ID, workbook.HandoffViewSchemaID, "txn-phase9-sprint7-shapes-handoff", MinimumValues(actor.ID, workbook.HandoffViewSchemaID), map[string]workbook.CollectionActionPayload{
		"handoff.open_task_ids":     Collection(addOptionalSurfaceRecordRef(taskID)),
		"handoff.open_decision_ids": Collection(addOptionalSurfaceRecordRef(decisionID)),
		"handoff.open_risk_refs":    {Actions: []workbook.CollectionAction{{Op: "add_risk_ref", RiskRefText: "Escalate outbound access", NormalizedText: "escalate outbound access"}}},
	}, Time(time.Minute))
	handoffRow := handoff.Payload["row"].(map[string]any)
	requireRecordRefItemShape(t, handoffRow, "handoff.open_task_ids", taskID)
	requireRecordRefItemShape(t, handoffRow, "handoff.open_decision_ids", decisionID)
	requireRiskRefItemShape(t, handoffRow, "handoff.open_risk_refs", "Escalate outbound access")

	status := mustCreateRow(t, store, actor, incident.ID, workbook.StatusReviewViewSchemaID, "txn-phase9-sprint7-shapes-status", MinimumValues(actor.ID, workbook.StatusReviewViewSchemaID), map[string]workbook.CollectionActionPayload{
		"status_review.blocked_task_ids":     Collection(addOptionalSurfaceRecordRef(taskID)),
		"status_review.pending_evidence_ids": Collection(addOptionalSurfaceRecordRef(evidenceID)),
		"status_review.open_decision_ids":    Collection(addOptionalSurfaceRecordRef(decisionID)),
	}, Time(2*time.Minute))
	statusRow := status.Payload["row"].(map[string]any)
	requireRecordRefItemShape(t, statusRow, "status_review.blocked_task_ids", taskID)
	requireRecordRefItemShape(t, statusRow, "status_review.pending_evidence_ids", evidenceID)
	requireRecordRefItemShape(t, statusRow, "status_review.open_decision_ids", decisionID)

	lesson := mustCreateRow(t, store, actor, incident.ID, workbook.LessonViewSchemaID, "txn-phase9-sprint7-shapes-lesson", MinimumValues(actor.ID, workbook.LessonViewSchemaID), map[string]workbook.CollectionActionPayload{
		"lesson.follow_up_task_ids": Collection(addOptionalSurfaceRecordRef(taskID)),
		"lesson.evidence_refs":      Collection(addOptionalSurfaceRecordRef(evidenceID)),
	}, Time(3*time.Minute))
	lessonRow := lesson.Payload["row"].(map[string]any)
	requireRecordRefItemShape(t, lessonRow, "lesson.follow_up_task_ids", taskID)
	requireRecordRefItemShape(t, lessonRow, "lesson.evidence_refs", evidenceID)
}

func TestCoordinationDuplicateCoalescing_Unit(t *testing.T) {
	harness := recordstoretest.StartStore(t, "phase9-sprint7-coordination-duplicate-coalescing")
	store := workbook.NewStore(harness.DB)
	actor := recordstoretest.SeedLocalUserFlags(t, harness.DB, "sprint7-duplicates@example.test", "Sprint7 Duplicates", "Sprint7Duplicates1!", false, false, true)
	incident := recordstoretest.CreateIncidentInStore(t, harness.DB, actor, "txn-phase9-sprint7-duplicates-incident", "IR-S7-DUPLICATES", "Workbook inspector Sprint 7 duplicate coalescing")

	partyID := mustCreatePartyFor(t, store, actor, incident.ID, "txn-phase9-sprint7-duplicates-party", "Duplicate Party")
	decisionID := mustCreateDecision(t, store, actor, incident.ID, "txn-phase9-sprint7-duplicates-decision", "approved", "Duplicate decision")
	taskID := mustCreateTaskFor(t, store, actor, incident.ID, "txn-phase9-sprint7-duplicates-task", "Duplicate task")
	evidenceID := mustCreateEvidenceFor(t, store, actor, incident.ID, "txn-phase9-sprint7-duplicates-evidence", "Duplicate evidence")

	for _, tc := range []struct {
		name         string
		viewSchemaID string
		fieldKey     string
		action       workbook.CollectionAction
		wantLinkID   uuid.UUID
	}{
		{"comm-decision", workbook.CommLogViewSchemaID, "comm_log.decision_ids", addOptionalSurfaceRecordRef(decisionID), decisionID},
		{"comm-task", workbook.CommLogViewSchemaID, "comm_log.action_task_ids", addOptionalSurfaceRecordRef(taskID), taskID},
		{"comm-audience-party", workbook.CommLogViewSchemaID, "comm_log.audience_party_ids", workbook.CollectionAction{Op: "add_party_ref", PartyID: &partyID}, partyID},
		{"comm-attendee-party", workbook.CommLogViewSchemaID, "comm_log.attendee_party_ids", workbook.CollectionAction{Op: "add_party_ref", PartyID: &partyID}, partyID},
		{"handoff-task", workbook.HandoffViewSchemaID, "handoff.open_task_ids", addOptionalSurfaceRecordRef(taskID), taskID},
		{"handoff-decision", workbook.HandoffViewSchemaID, "handoff.open_decision_ids", addOptionalSurfaceRecordRef(decisionID), decisionID},
		{"status-task", workbook.StatusReviewViewSchemaID, "status_review.blocked_task_ids", addOptionalSurfaceRecordRef(taskID), taskID},
		{"status-evidence", workbook.StatusReviewViewSchemaID, "status_review.pending_evidence_ids", addOptionalSurfaceRecordRef(evidenceID), evidenceID},
		{"status-decision", workbook.StatusReviewViewSchemaID, "status_review.open_decision_ids", addOptionalSurfaceRecordRef(decisionID), decisionID},
		{"lesson-task", workbook.LessonViewSchemaID, "lesson.follow_up_task_ids", addOptionalSurfaceRecordRef(taskID), taskID},
		{"lesson-evidence", workbook.LessonViewSchemaID, "lesson.evidence_refs", addOptionalSurfaceRecordRef(evidenceID), evidenceID},
	} {
		t.Run(tc.name, func(t *testing.T) {
			clientTxnID := "txn-phase9-sprint7-duplicates-" + tc.name
			values := MinimumValues(actor.ID, tc.viewSchemaID)
			collections := map[string]workbook.CollectionActionPayload{
				tc.fieldKey: Collection(tc.action, tc.action),
			}
			result := mustCreateRow(t, store, actor, incident.ID, tc.viewSchemaID, clientTxnID, values, collections, Time(0))
			requireCoordinationCollectionItemCount(t, result.Payload["row"].(map[string]any), tc.fieldKey, 1)
			if got := countLinksForField(t, harness, result.RecordID, tc.fieldKey); got != 1 {
				t.Fatalf("%s duplicate create links: got %d want 1", tc.fieldKey, got)
			}
			requireManualReferenceLink(t, harness, result.RecordID, tc.wantLinkID, tc.fieldKey, "references_record")

			replay := mustCreateRow(t, store, actor, incident.ID, tc.viewSchemaID, clientTxnID, values, collections, Time(30*time.Minute))
			if replay.RecordID != result.RecordID {
				t.Fatalf("%s replay record_id: got %s want %s", tc.fieldKey, replay.RecordID, result.RecordID)
			}
			requireCoordinationCollectionItemCount(t, replay.Payload["row"].(map[string]any), tc.fieldKey, 1)
			if got := countLinksForField(t, harness, result.RecordID, tc.fieldKey); got != 1 {
				t.Fatalf("%s duplicate replay links: got %d want 1", tc.fieldKey, got)
			}
		})
	}

	riskClientTxnID := "txn-phase9-sprint7-duplicates-risk"
	riskValues := MinimumValues(actor.ID, workbook.HandoffViewSchemaID)
	riskCollections := map[string]workbook.CollectionActionPayload{
		"handoff.open_risk_refs": {
			Actions: []workbook.CollectionAction{
				{Op: "add_risk_ref", RiskRefText: " Repeated risk reference ", NormalizedText: "repeated risk reference"},
				{Op: "add_risk_ref", RiskRefText: "Repeated risk reference", NormalizedText: "repeated risk reference"},
			},
		},
	}
	handoff := mustCreateRow(t, store, actor, incident.ID, workbook.HandoffViewSchemaID, riskClientTxnID, riskValues, riskCollections, Time(time.Minute))
	requireCoordinationCollectionItemCount(t, handoff.Payload["row"].(map[string]any), "handoff.open_risk_refs", 1)
	if got := countActiveRiskRefs(t, harness, handoff.RecordID); got != 1 {
		t.Fatalf("handoff risk duplicate create refs: got %d want 1", got)
	}

	riskReplay := mustCreateRow(t, store, actor, incident.ID, workbook.HandoffViewSchemaID, riskClientTxnID, riskValues, riskCollections, Time(31*time.Minute))
	if riskReplay.RecordID != handoff.RecordID {
		t.Fatalf("handoff risk replay record_id: got %s want %s", riskReplay.RecordID, handoff.RecordID)
	}
	requireCoordinationCollectionItemCount(t, riskReplay.Payload["row"].(map[string]any), "handoff.open_risk_refs", 1)
	if got := countActiveRiskRefs(t, harness, handoff.RecordID); got != 1 {
		t.Fatalf("handoff risk duplicate replay refs: got %d want 1", got)
	}
}

func TestCoordinationCollectionsAndValidation_Unit(t *testing.T) {
	ctx := context.Background()
	harness := recordstoretest.StartStore(t, "phase9-sprint7-coordination-collections")
	store := workbook.NewStore(harness.DB)
	actor := recordstoretest.SeedLocalUserFlags(t, harness.DB, "sprint7-collections@example.test", "Sprint7 Collections", "Sprint7Collections1!", false, false, true)
	incident := recordstoretest.CreateIncidentInStore(t, harness.DB, actor, "txn-phase9-sprint7-collections-incident", "IR-S7-COLLECTIONS", "Workbook inspector Sprint 7 coordination collections")

	partyID := mustCreatePartyFor(t, store, actor, incident.ID, "txn-phase9-sprint7-party", "Coordination Legal")
	otherPartyID := mustCreatePartyFor(t, store, actor, incident.ID, "txn-phase9-sprint7-party-other", "Coordination Legal Alternate")
	decisionID := mustCreateDecision(t, store, actor, incident.ID, "txn-phase9-sprint7-decision", "approved", "Approve coordination plan")
	taskID := mustCreateTaskFor(t, store, actor, incident.ID, "txn-phase9-sprint7-task", "Coordinate endpoint logs")
	evidenceID := mustCreateEvidenceFor(t, store, actor, incident.ID, "txn-phase9-sprint7-evidence", "Coordination evidence")

	otherIncident := recordstoretest.CreateIncidentInStore(t, harness.DB, actor, "txn-phase9-sprint7-foreign-incident", "IR-S7-COLLECTIONS-FOREIGN", "Workbook inspector Sprint 7 foreign incident")
	foreignEvidenceID := mustCreateEvidenceFor(t, store, actor, otherIncident.ID, "txn-phase9-sprint7-foreign-evidence", "Foreign evidence")
	foreignTaskID := mustCreateTaskFor(t, store, actor, otherIncident.ID, "txn-phase9-sprint7-foreign-task", "Foreign task")
	foreignDecisionID := mustCreateDecision(t, store, actor, otherIncident.ID, "txn-phase9-sprint7-foreign-decision", "approved", "Foreign decision")
	foreignPartyID := mustCreatePartyFor(t, store, actor, otherIncident.ID, "txn-phase9-sprint7-foreign-party", "Foreign party")
	deletedEvidenceID := mustCreateEvidenceFor(t, store, actor, incident.ID, "txn-phase9-sprint7-deleted-evidence", "Deleted evidence")
	deletedTaskID := mustCreateTaskFor(t, store, actor, incident.ID, "txn-phase9-sprint7-deleted-task", "Deleted task")
	deletedDecisionID := mustCreateDecision(t, store, actor, incident.ID, "txn-phase9-sprint7-deleted-decision", "approved", "Deleted decision")
	deletedPartyID := mustCreatePartyFor(t, store, actor, incident.ID, "txn-phase9-sprint7-deleted-party", "Deleted party")
	for label, recordID := range map[string]uuid.UUID{
		"evidence": deletedEvidenceID,
		"task":     deletedTaskID,
		"decision": deletedDecisionID,
		"party":    deletedPartyID,
	} {
		if _, err := harness.DB.Exec(ctx, `UPDATE records SET deleted_at = $2, deleted_by_user_id = $3 WHERE record_id = $1`, recordID, Time(time.Hour), actor.ID); err != nil {
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
			before := countDurableState(t, harness, incident.ID)
			clientTxnID := "txn-phase9-sprint7-invalid-" + strings.ReplaceAll(field.fieldKey, ".", "-") + "-" + invalid.name
			_, err := store.CreateWorkbookRow(ctx, actor, incident.ID, workbook.CreateRequest{
				ViewSchemaID: field.viewSchemaID,
				ClientTxnID:  clientTxnID,
				Values:       MinimumValues(actor.ID, field.viewSchemaID),
				Collections:  map[string]workbook.CollectionActionPayload{field.fieldKey: Collection(addOptionalSurfaceRecordRef(invalid.id))},
			}, []byte(clientTxnID), "req-"+clientTxnID, Time(2*time.Hour))
			requireMutationValidation(t, err, field.fieldKey, "invalid_value")
			requireDurableState(t, harness, incident.ID, before, field.fieldKey+" "+invalid.name)
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
			before := countDurableState(t, harness, incident.ID)
			clientTxnID := "txn-phase9-sprint7-invalid-" + strings.ReplaceAll(fieldKey, ".", "-") + "-" + invalid.name
			_, err := store.CreateWorkbookRow(ctx, actor, incident.ID, workbook.CreateRequest{
				ViewSchemaID: workbook.CommLogViewSchemaID,
				ClientTxnID:  clientTxnID,
				Values:       MinimumValues(actor.ID, workbook.CommLogViewSchemaID),
				Collections:  map[string]workbook.CollectionActionPayload{fieldKey: Collection(workbook.CollectionAction{Op: "add_party_ref", PartyID: &invalid.id})},
			}, []byte(clientTxnID), "req-"+clientTxnID, Time(3*time.Hour))
			requireMutationValidation(t, err, fieldKey, "invalid_value")
			requireDurableState(t, harness, incident.ID, before, fieldKey+" "+invalid.name)
		}
	}

	comm := mustCreateRow(t, store, actor, incident.ID, workbook.CommLogViewSchemaID, "txn-phase9-sprint7-comm-collections", map[string]workbook.ValueChange{
		"comm_log.comm_type":          Text("briefing"),
		"comm_log.audience":           Text("Leadership source text"),
		"comm_log.channel_or_meeting": Text("Bridge"),
		"comm_log.summary":            Text("Collection coordination update"),
	}, map[string]workbook.CollectionActionPayload{
		"comm_log.decision_ids":       Collection(addOptionalSurfaceRecordRef(decisionID), addOptionalSurfaceRecordRef(decisionID)),
		"comm_log.action_task_ids":    Collection(addOptionalSurfaceRecordRef(taskID), addOptionalSurfaceRecordRef(taskID)),
		"comm_log.audience_party_ids": Collection(workbook.CollectionAction{Op: "add_party_ref", PartyID: &partyID}, workbook.CollectionAction{Op: "add_party_ref", PartyID: &partyID}),
	}, Time(4*time.Hour))
	commRow := comm.Payload["row"].(map[string]any)
	requireCoordinationCellValue(t, commRow, "comm_log.audience", "Leadership source text")
	requireCoordinationCollectionItemCount(t, commRow, "comm_log.decision_ids", 1)
	requireCollectionItemKind(t, commRow, "comm_log.decision_ids", "record_ref")
	requireCoordinationCollectionItemCount(t, commRow, "comm_log.action_task_ids", 1)
	requireCoordinationCollectionItemCount(t, commRow, "comm_log.audience_party_ids", 1)
	requireCollectionItemKind(t, commRow, "comm_log.audience_party_ids", "party_ref")
	requireManualReferenceLink(t, harness, comm.RecordID, decisionID, "comm_log.decision_ids", "references_record")
	requireManualReferenceLink(t, harness, comm.RecordID, taskID, "comm_log.action_task_ids", "references_record")
	requireManualReferenceLink(t, harness, comm.RecordID, partyID, "comm_log.audience_party_ids", "references_record")

	commWithAttendee := mustPatch(t, store, actor, comm.RecordID, workbook.CommLogViewSchemaID, 1, "txn-phase9-sprint7-comm-add-attendee",
		CollectionChange("comm_log.attendee_party_ids", Collection(workbook.CollectionAction{Op: "add_party_ref", PartyID: &otherPartyID})))
	requireCoordinationCellValue(t, commWithAttendee.Payload["row"].(map[string]any), "comm_log.audience", "Leadership source text")
	commWithoutAttendee := mustPatch(t, store, actor, comm.RecordID, workbook.CommLogViewSchemaID, 2, "txn-phase9-sprint7-comm-remove-attendee",
		CollectionChange("comm_log.attendee_party_ids", Collection(workbook.CollectionAction{Op: "remove_party_ref", ItemRef: "party_ref:" + otherPartyID.String()})))
	requireCoordinationCellValue(t, commWithoutAttendee.Payload["row"].(map[string]any), "comm_log.audience", "Leadership source text")

	handoff := mustCreateRow(t, store, actor, incident.ID, workbook.HandoffViewSchemaID, "txn-phase9-sprint7-handoff-collections", map[string]workbook.ValueChange{
		"handoff.incoming_owner_user_id": UUID(actor.ID),
		"handoff.current_state_summary":  Text("Open coordination work"),
	}, map[string]workbook.CollectionActionPayload{
		"handoff.open_task_ids":     Collection(addOptionalSurfaceRecordRef(taskID)),
		"handoff.open_decision_ids": Collection(addOptionalSurfaceRecordRef(decisionID)),
		"handoff.open_risk_refs": {
			Actions: []workbook.CollectionAction{
				{Op: "add_risk_ref", RiskRefText: " Pending outbound access review ", NormalizedText: "pending outbound access review"},
				{Op: "add_risk_ref", RiskRefText: "Pending outbound access review", NormalizedText: "pending outbound access review"},
			},
		},
	}, Time(5*time.Hour))
	handoffRow := handoff.Payload["row"].(map[string]any)
	requireCoordinationCollectionItemCount(t, handoffRow, "handoff.open_task_ids", 1)
	requireCoordinationCollectionItemCount(t, handoffRow, "handoff.open_decision_ids", 1)
	requireCoordinationCollectionItemCount(t, handoffRow, "handoff.open_risk_refs", 1)
	requireCollectionItemKind(t, handoffRow, "handoff.open_risk_refs", "risk_ref")
	riskRef := requireSingleItemRef(t, handoffRow, "handoff.open_risk_refs", "risk_ref:")
	if got := countActiveRiskRefs(t, harness, handoff.RecordID); got != 1 {
		t.Fatalf("handoff duplicate risk refs: got %d want 1", got)
	}
	if got := countLinksForField(t, harness, handoff.RecordID, "handoff.open_risk_refs"); got != 0 {
		t.Fatalf("risk refs must not create generic record links, got %d", got)
	}
	secondHandoff := mustCreateRow(t, store, actor, incident.ID, workbook.HandoffViewSchemaID, "txn-phase9-sprint7-handoff-risk-scoped", map[string]workbook.ValueChange{
		"handoff.incoming_owner_user_id": UUID(actor.ID),
		"handoff.current_state_summary":  Text("Separate handoff same risk text"),
	}, map[string]workbook.CollectionActionPayload{
		"handoff.open_risk_refs": {Actions: []workbook.CollectionAction{{Op: "add_risk_ref", RiskRefText: "Pending outbound access review", NormalizedText: "pending outbound access review"}}},
	}, Time(6*time.Hour))
	secondRiskRef := requireSingleItemRef(t, secondHandoff.Payload["row"].(map[string]any), "handoff.open_risk_refs", "risk_ref:")
	if got := countActiveRiskRefs(t, harness, secondHandoff.RecordID); got != 1 {
		t.Fatalf("same risk text must be scoped per handoff: got %d want 1", got)
	}
	removedRisk := mustPatch(t, store, actor, handoff.RecordID, workbook.HandoffViewSchemaID, 1, "txn-phase9-sprint7-handoff-remove-risk",
		CollectionChange("handoff.open_risk_refs", workbook.CollectionActionPayload{Actions: []workbook.CollectionAction{{Op: "remove_risk_ref", ItemRef: riskRef}}}))
	requireCoordinationCollectionItemCount(t, removedRisk.Payload["row"].(map[string]any), "handoff.open_risk_refs", 0)
	if got := countActiveRiskRefs(t, harness, secondHandoff.RecordID); got != 1 {
		t.Fatalf("removing one handoff risk ref affected another handoff: got %d want 1", got)
	}

	status := mustCreateRow(t, store, actor, incident.ID, workbook.StatusReviewViewSchemaID, "txn-phase9-sprint7-status-collections", map[string]workbook.ValueChange{
		"status_review.current_state_summary": Text("Status with open coordination work"),
	}, map[string]workbook.CollectionActionPayload{
		"status_review.blocked_task_ids":     Collection(addOptionalSurfaceRecordRef(taskID)),
		"status_review.pending_evidence_ids": Collection(addOptionalSurfaceRecordRef(evidenceID)),
		"status_review.open_decision_ids":    Collection(addOptionalSurfaceRecordRef(decisionID)),
	}, Time(7*time.Hour))
	statusRow := status.Payload["row"].(map[string]any)
	requireCoordinationCollectionItemCount(t, statusRow, "status_review.blocked_task_ids", 1)
	requireCoordinationCollectionItemCount(t, statusRow, "status_review.pending_evidence_ids", 1)
	requireCoordinationCollectionItemCount(t, statusRow, "status_review.open_decision_ids", 1)

	lesson := mustCreateRow(t, store, actor, incident.ID, workbook.LessonViewSchemaID, "txn-phase9-sprint7-lesson-collections", map[string]workbook.ValueChange{
		"lesson.summary": Text("Follow up on evidence and work"),
	}, map[string]workbook.CollectionActionPayload{
		"lesson.follow_up_task_ids": Collection(addOptionalSurfaceRecordRef(taskID)),
		"lesson.evidence_refs":      Collection(addOptionalSurfaceRecordRef(evidenceID)),
	}, Time(8*time.Hour))
	lessonRow := lesson.Payload["row"].(map[string]any)
	requireCoordinationCollectionItemCount(t, lessonRow, "lesson.follow_up_task_ids", 1)
	requireCoordinationCollectionItemCount(t, lessonRow, "lesson.evidence_refs", 1)

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
			requireInvalidCollectionPatch(t, harness, store, actor, incident.ID, field.recordID, field.viewSchemaID, field.fieldKey, Collection(workbook.CollectionAction{Op: "remove_record_ref", ItemRef: invalid.itemRef}), field.fieldKey+" remove "+invalid.name)
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
			requireInvalidCollectionPatch(t, harness, store, actor, incident.ID, field.recordID, workbook.CommLogViewSchemaID, field.fieldKey, Collection(workbook.CollectionAction{Op: "remove_party_ref", ItemRef: invalid.itemRef}), field.fieldKey+" remove "+invalid.name)
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
		requireInvalidCollectionPatch(t, harness, store, actor, incident.ID, handoff.RecordID, workbook.HandoffViewSchemaID, "handoff.open_risk_refs", Collection(workbook.CollectionAction{Op: "remove_risk_ref", ItemRef: invalid.itemRef}), "handoff.open_risk_refs remove "+invalid.name)
	}
}

func mustCreateRow(t testing.TB, store *workbook.Store, actor authn.UserRecord, incidentID uuid.UUID, viewSchemaID string, clientTxnID string, values map[string]workbook.ValueChange, collections map[string]workbook.CollectionActionPayload, now time.Time) workbook.MutationResult {
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

func Text(value string) workbook.ValueChange {
	return workbook.ValueChange{Kind: "text", Text: stringPtr(value)}
}

func UUID(value uuid.UUID) workbook.ValueChange {
	return workbook.ValueChange{Kind: "uuid", UUID: &value}
}

func Timestamp(value time.Time) workbook.ValueChange {
	value = value.UTC()
	return workbook.ValueChange{Kind: "timestamp", Timestamp: &value}
}

func StringPtr(value string) *string {
	return &value
}

func Time(offset time.Duration) time.Time {
	return time.Date(2026, 5, 19, 15, 0, 0, 0, time.UTC).Add(offset)
}

func MinimumValues(actorID uuid.UUID, viewSchemaID string) map[string]workbook.ValueChange {
	switch viewSchemaID {
	case workbook.CommLogViewSchemaID:
		return map[string]workbook.ValueChange{
			"comm_log.comm_type":          Text("briefing"),
			"comm_log.audience":           Text("Audience text"),
			"comm_log.channel_or_meeting": Text("Bridge"),
			"comm_log.summary":            Text("Coordination update"),
		}
	case workbook.HandoffViewSchemaID:
		return map[string]workbook.ValueChange{
			"handoff.incoming_owner_user_id": UUID(actorID),
			"handoff.current_state_summary":  Text("Handoff state"),
		}
	case workbook.StatusReviewViewSchemaID:
		return map[string]workbook.ValueChange{
			"status_review.current_state_summary": Text("Status review state"),
		}
	case workbook.LessonViewSchemaID:
		return map[string]workbook.ValueChange{
			"lesson.summary": Text("Lesson summary"),
		}
	default:
		return map[string]workbook.ValueChange{}
	}
}

type DurableState struct {
	Records     int
	Artifacts   int
	RecordLinks int
	RiskRefs    int
}

func countDurableState(t testing.TB, harness *recordstoretest.StoreHarness, incidentID uuid.UUID) DurableState {
	t.Helper()
	var state DurableState
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

func requireDurableState(t testing.TB, harness *recordstoretest.StoreHarness, incidentID uuid.UUID, want DurableState, context string) {
	t.Helper()
	got := countDurableState(t, harness, incidentID)
	if got != want {
		t.Fatalf("%s changed durable state: got %#v want %#v", context, got, want)
	}
}

func countSQLDurableState(t testing.TB, db *sql.DB, incidentID string) DurableState {
	t.Helper()
	var state DurableState
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

func requireSQLDurableState(t testing.TB, db *sql.DB, incidentID string, want DurableState, context string) {
	t.Helper()
	got := countSQLDurableState(t, db, incidentID)
	if got != want {
		t.Fatalf("%s changed durable SQL state: got %#v want %#v", context, got, want)
	}
}

func countSQLProjectionRows(t testing.TB, db *sql.DB, incidentID string, artifactType string) int {
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

func requireArtifactType(t testing.TB, harness *recordstoretest.StoreHarness, recordID uuid.UUID, artifactType string) {
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

func requireMutationValidation(t testing.TB, err error, field string, reason string) {
	t.Helper()
	var validationErr *workbook.MutationValidationError
	if !errors.As(err, &validationErr) {
		t.Fatalf("expected mutation validation error, got %v", err)
	}
	if validationErr.Field != field || validationErr.ReasonCode != reason {
		t.Fatalf("unexpected validation error: %#v", validationErr)
	}
}

func requireInvalidCollectionPatch(t testing.TB, harness *recordstoretest.StoreHarness, store *workbook.Store, actor authn.UserRecord, incidentID uuid.UUID, recordID uuid.UUID, viewSchemaID string, fieldKey string, collection workbook.CollectionActionPayload, context string) {
	t.Helper()
	before := countDurableState(t, harness, incidentID)
	baseVersion := RecordVersion(t, harness.DB, recordID)
	_, err := Patch(store, actor, recordID, viewSchemaID, baseVersion, "txn-phase9-sprint7-invalid-"+strings.ReplaceAll(context, " ", "-"), CollectionChange(fieldKey, collection))
	requireMutationValidation(t, err, fieldKey, "invalid_value")
	requireDurableState(t, harness, incidentID, before, context)
	if got := RecordVersion(t, harness.DB, recordID); got != baseVersion {
		t.Fatalf("%s changed row version: got %d want %d", context, got, baseVersion)
	}
}

func requireNoRecordChanged(t testing.TB, client *incidentwstest.Client, timeout time.Duration) {
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

func expectDecodePatchRejected(t testing.TB, viewSchemaID string, fieldKey string, value any) {
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

func requireCoordinationCellValue(t testing.TB, row map[string]any, fieldKey string, want any) {
	t.Helper()
	got := row["cells"].(map[string]any)[fieldKey].(map[string]any)["value"]
	if got != want {
		t.Fatalf("unexpected %s value: got %#v want %#v", fieldKey, got, want)
	}
}

func requireCellNonEmpty(t testing.TB, row map[string]any, fieldKey string) {
	t.Helper()
	got := row["cells"].(map[string]any)[fieldKey].(map[string]any)["value"]
	if got == nil || got == "" {
		t.Fatalf("expected non-empty %s value, got %#v", fieldKey, got)
	}
}

func requireCoordinationCollectionItemCount(t testing.TB, row map[string]any, fieldKey string, want int) {
	t.Helper()
	items := CollectionItems(t, row, fieldKey)
	if len(items) != want {
		t.Fatalf("unexpected %s item count: got %d want %d items=%#v", fieldKey, len(items), want, items)
	}
}

func requireCollectionItemKind(t testing.TB, row map[string]any, fieldKey string, itemKind string) {
	t.Helper()
	for _, item := range CollectionItems(t, row, fieldKey) {
		if item["item_kind"] != itemKind {
			t.Fatalf("unexpected %s item kind: got %#v want %s", fieldKey, item["item_kind"], itemKind)
		}
	}
}

func requireSingleItemRef(t testing.TB, row map[string]any, fieldKey string, prefix string) string {
	t.Helper()
	items := CollectionItems(t, row, fieldKey)
	if len(items) != 1 {
		t.Fatalf("expected one %s item, got %#v", fieldKey, items)
	}
	itemRef, ok := items[0]["item_ref"].(string)
	if !ok || !strings.HasPrefix(itemRef, prefix) {
		t.Fatalf("unexpected %s item_ref: %#v", fieldKey, items[0]["item_ref"])
	}
	return itemRef
}

func CollectionItems(t testing.TB, row map[string]any, fieldKey string) []map[string]any {
	t.Helper()
	value := CollectionValue(t, row, fieldKey)
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

func CollectionValue(t testing.TB, row map[string]any, fieldKey string) map[string]any {
	t.Helper()
	value := row["cells"].(map[string]any)[fieldKey].(map[string]any)["value"].(map[string]any)
	if value["kind"] != "collection_value_v1" {
		t.Fatalf("expected %s to be collection_value_v1, got %#v", fieldKey, value)
	}
	return value
}

func requireCollectionUnordered(t testing.TB, row map[string]any, fieldKey string) {
	t.Helper()
	value := CollectionValue(t, row, fieldKey)
	if value["ordered"] != false {
		t.Fatalf("%s ordered: got %#v want false", fieldKey, value["ordered"])
	}
}

func requireRecordRefItemShape(t testing.TB, row map[string]any, fieldKey string, targetID uuid.UUID) {
	t.Helper()
	requireCollectionUnordered(t, row, fieldKey)
	items := CollectionItems(t, row, fieldKey)
	if len(items) != 1 {
		t.Fatalf("expected one %s item, got %#v", fieldKey, items)
	}
	item := items[0]
	requireExactItemKeys(t, item, "display_text", "item_kind", "item_ref", "linked_record_id")
	target := targetID.String()
	requireItemString(t, item, "item_kind", "record_ref")
	requireItemString(t, item, "item_ref", "record_ref:"+target)
	requireItemString(t, item, "linked_record_id", target)
	if displayText, ok := item["display_text"].(string); !ok || displayText == "" {
		t.Fatalf("%s display_text must be non-empty string, got %#v", fieldKey, item["display_text"])
	}
}

func requirePartyRefItemShape(t testing.TB, row map[string]any, fieldKey string, partyID uuid.UUID) {
	t.Helper()
	requireCollectionUnordered(t, row, fieldKey)
	items := CollectionItems(t, row, fieldKey)
	if len(items) != 1 {
		t.Fatalf("expected one %s item, got %#v", fieldKey, items)
	}
	item := items[0]
	requireExactItemKeys(t, item, "display_text", "item_kind", "item_ref", "party_id")
	target := partyID.String()
	requireItemString(t, item, "item_kind", "party_ref")
	requireItemString(t, item, "item_ref", "party_ref:"+target)
	requireItemString(t, item, "party_id", target)
	if displayText, ok := item["display_text"].(string); !ok || displayText == "" {
		t.Fatalf("%s display_text must be non-empty string, got %#v", fieldKey, item["display_text"])
	}
}

func requireRiskRefItemShape(t testing.TB, row map[string]any, fieldKey string, riskText string) {
	t.Helper()
	requireCollectionUnordered(t, row, fieldKey)
	items := CollectionItems(t, row, fieldKey)
	if len(items) != 1 {
		t.Fatalf("expected one %s item, got %#v", fieldKey, items)
	}
	item := items[0]
	requireExactItemKeys(t, item, "display_text", "item_kind", "item_ref", "risk_ref_id", "risk_ref_text")
	requireItemString(t, item, "item_kind", "risk_ref")
	requireItemString(t, item, "display_text", riskText)
	requireItemString(t, item, "risk_ref_text", riskText)
	riskRefID, ok := item["risk_ref_id"].(string)
	if !ok || riskRefID == "" {
		t.Fatalf("%s risk_ref_id must be non-empty string, got %#v", fieldKey, item["risk_ref_id"])
	}
	requireItemString(t, item, "item_ref", "risk_ref:"+riskRefID)
}

func requireExactItemKeys(t testing.TB, item map[string]any, keys ...string) {
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

func requireItemString(t testing.TB, item map[string]any, key string, want string) {
	t.Helper()
	if got, ok := item[key].(string); !ok || got != want {
		t.Fatalf("unexpected %s: got %#v want %q in %#v", key, item[key], want, item)
	}
}

func requireFilterMatchesOnly(t testing.TB, store *workbook.Store, incidentID uuid.UUID, viewSchemaID string, filter viewschema.Filter, wantRecordID uuid.UUID) {
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

func requireProjectedRow(t testing.TB, store *workbook.Store, incidentID uuid.UUID, viewSchemaID string, query viewschema.QueryMeta, recordID uuid.UUID, groupField string, groupValue any) {
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

func requireRecordOrder(t testing.TB, rows []map[string]any, want []uuid.UUID) {
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

func requireManualReferenceLink(t testing.TB, harness *recordstoretest.StoreHarness, sourceID uuid.UUID, targetID uuid.UUID, fieldKey string, linkType string) {
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

func countActiveRiskRefs(t testing.TB, harness *recordstoretest.StoreHarness, handoffID uuid.UUID) int {
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

func countLinksForField(t testing.TB, harness *recordstoretest.StoreHarness, sourceID uuid.UUID, fieldKey string) int {
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
