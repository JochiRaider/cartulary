package timeline_test

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/JochiRaider/cartulary/internal/app/timelineassembly"
	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	incidentscenariotest "github.com/JochiRaider/cartulary/internal/modules/incidents/testsupport/scenariotest"
	"github.com/JochiRaider/cartulary/internal/modules/timeline"
	timelineadmission "github.com/JochiRaider/cartulary/internal/modules/timeline/admission"
	"github.com/JochiRaider/cartulary/internal/modules/timeline/testsupport/asserttest"
	"github.com/JochiRaider/cartulary/internal/modules/timeline/testsupport/fakeports"
	"github.com/JochiRaider/cartulary/internal/modules/timeline/testsupport/routetest"
	"github.com/JochiRaider/cartulary/internal/modules/timeline/workbookprojection"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/authcookietest"
	"github.com/JochiRaider/cartulary/internal/testutil/collaborationsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/collaborationsupport/incidentwstest"
	"github.com/JochiRaider/cartulary/internal/testutil/conflicttest"
	"github.com/JochiRaider/cartulary/internal/testutil/contractassert"
	"github.com/JochiRaider/cartulary/internal/testutil/dbassert"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	workbookscenariotest "github.com/JochiRaider/cartulary/internal/testutil/workbookscenariotest"
)

func TestCreatePatchReplayAndRollback_Integration(t *testing.T) {
	runtime := appsupport.StartRuntime(t)

	t.Run("create patch review and supersede replays stay single-write and preserve substrate", func(t *testing.T) {
		server, db, _ := startServer(t, runtime, "timeline_mutation-i-3-01-replay")

		adminLogin, adminID := flowtest.ProvisionBootstrapAdmin(t, http.DefaultClient, server.HTTP.URL)
		incident := createIncident(t, server, adminLogin, map[string]any{
			"client_txn_id": "txn-i-3-01-incident",
			"incident_key":  "IR-I301",
			"title":         "Timeline substrate",
		})
		incidentID := incident["incident_id"].(string)
		socket := incidentwstest.ConnectViewSocket(t, server, incidentID, timeline.TimelineViewSchemaID, adminLogin.Session.Value)
		defer socket.Close(1000, "test_complete")

		createResp := httptestx.DoJSON(t, http.DefaultClient,
			http.MethodPost,
			server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/views/"+timeline.TimelineViewSchemaID+"/rows",
			map[string]any{
				"client_txn_id": "txn-i-3-01-row-create-zero",
			},
			httptestx.WithCookies(adminLogin.Session, adminLogin.CSRF),
			httptestx.WithHeader(authn.CSRFHeaderName, adminLogin.CSRF.Value),
		)
		createData := httptestx.RequireSuccessEnvelope(t, createResp, http.StatusCreated)["data"].(map[string]any)
		createRow := createData["row"].(map[string]any)
		recordID := createRow["record_id"].(string)
		if createRow["row_version"] != float64(1) {
			t.Fatalf("unexpected zero-field create row_version: %#v", createRow)
		}
		createCells := createRow["cells"].(map[string]any)
		if createCells["timeline.activity_synopsis_text"].(map[string]any)["value"] != nil {
			t.Fatalf("expected zero-field create to keep summary null, got %#v", createCells["timeline.activity_synopsis_text"])
		}
		if createCells["timeline.capture_state"].(map[string]any)["value"] != "rough" {
			t.Fatalf("expected zero-field create to start rough, got %#v", createCells["timeline.capture_state"])
		}
		incidentwstest.RequireRecordChanged(t, socket, recordID, 1)
		requireMutationRecorded(t, db, createData["change_set_id"].(string), recordID, adminID.String(), "timeline.rows.create", "txn-i-3-01-row-create-zero", 1, 1)
		asserttest.RequireTimelineRecordMutation(t, asserttest.SQLDatabase(db), createData["change_set_id"].(string), asserttest.TimelineRecordMutationExpectation{
			SequenceNo:      1,
			RecordID:        recordID,
			OperationKind:   "create",
			AfterRowVersion: asserttest.RowVersion(1),
			AfterCells: map[string]any{
				"timeline.activity_synopsis_text": nil,
				"timeline.capture_state":          "rough",
			},
		})

		projection := asserttest.LookupProjectionRow(t, asserttest.SQLDatabase(db), recordID)
		if projection.CaptureState != "rough" || projection.ReplacementRecordID != nil {
			t.Fatalf("unexpected zero-field projection row: %#v", projection)
		}
		countersAfterCreateState := snapshotTimelineMutationState(t, db, incidentID, recordID)
		createReplay := httptestx.DoJSON(t, http.DefaultClient,
			http.MethodPost,
			server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/views/"+timeline.TimelineViewSchemaID+"/rows",
			map[string]any{
				"client_txn_id": "txn-i-3-01-row-create-zero",
			},
			httptestx.WithCookies(adminLogin.Session, adminLogin.CSRF),
			httptestx.WithHeader(authn.CSRFHeaderName, adminLogin.CSRF.Value),
		)
		createReplayData := httptestx.RequireSuccessEnvelope(t, createReplay, http.StatusOK)["data"].(map[string]any)
		if createReplayData["change_set_id"] != createData["change_set_id"] {
			t.Fatalf("expected create replay to return original payload, got %#v", createReplayData)
		}
		incidentwstest.ExpectNoSocketMessage(t, socket)
		if got := snapshotTimelineMutationState(t, db, incidentID, recordID); got != countersAfterCreateState {
			t.Fatalf("replay or rejection changed mutation state: before=%+v after=%+v", countersAfterCreateState, got)
		}

		divergentCreate := httptestx.DoJSON(t, http.DefaultClient,
			http.MethodPost,
			server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/views/"+timeline.TimelineViewSchemaID+"/rows",
			map[string]any{
				"client_txn_id":                   "txn-i-3-01-row-create-zero",
				"timeline.activity_synopsis_text": "Different capture",
			},
			httptestx.WithCookies(adminLogin.Session, adminLogin.CSRF),
			httptestx.WithHeader(authn.CSRFHeaderName, adminLogin.CSRF.Value),
		)
		httptestx.RequireErrorEnvelope(t, divergentCreate, http.StatusConflict, "client_txn_conflict")
		if got := snapshotTimelineMutationState(t, db, incidentID, recordID); got != countersAfterCreateState {
			t.Fatalf("replay or rejection changed mutation state: before=%+v after=%+v", countersAfterCreateState, got)
		}
		incidentwstest.ExpectNoSocketMessage(t, socket)

		replacement := createTimelineRow(t, server, incidentID, adminLogin, map[string]any{
			"client_txn_id":                   "txn-i-3-01-replacement",
			"timeline.activity_synopsis_text": "Replacement row",
		})
		replacementRow := replacement["row"].(map[string]any)
		replacementID := replacementRow["record_id"].(string)
		if replacementRow["cells"].(map[string]any)["timeline.capture_state"].(map[string]any)["value"] != "rough" {
			t.Fatalf("expected one-value create to start rough, got %#v", replacementRow)
		}
		incidentwstest.RequireRecordChanged(t, socket, replacementID, 1)

		patchResp := httptestx.DoJSON(t, http.DefaultClient,
			http.MethodPatch,
			server.HTTP.URL+"/api/v1/records/"+recordID,
			map[string]any{
				"view_schema_id":   timeline.TimelineViewSchemaID,
				"base_row_version": 1,
				"client_txn_id":    "txn-i-3-01-row-patch",
				"changes": []map[string]any{
					{"field_key": "timeline.activity_synopsis_text", "value": "Enriched capture"},
					{"field_key": "timeline.raw_activity_text", "value": "Details from patch"},
				},
			},
			httptestx.WithCookies(adminLogin.Session, adminLogin.CSRF),
			httptestx.WithHeader(authn.CSRFHeaderName, adminLogin.CSRF.Value),
		)
		patchData := httptestx.RequireSuccessEnvelope(t, patchResp, http.StatusOK)["data"].(map[string]any)
		patchedRow := patchData["row"].(map[string]any)
		if patchedRow["row_version"] != float64(2) {
			t.Fatalf("unexpected patch row_version: %#v", patchedRow)
		}
		incidentwstest.RequireRecordChanged(t, socket, recordID, 2)
		requireMutationRecorded(t, db, patchData["change_set_id"].(string), recordID, adminID.String(), "timeline.records.patch", "txn-i-3-01-row-patch", 1, 2)
		asserttest.RequireTimelineRecordMutation(t, asserttest.SQLDatabase(db), patchData["change_set_id"].(string), asserttest.TimelineRecordMutationExpectation{
			SequenceNo:       1,
			RecordID:         recordID,
			OperationKind:    "patch",
			BeforeRowVersion: asserttest.RowVersion(1),
			AfterRowVersion:  asserttest.RowVersion(2),
			BeforeCells: map[string]any{
				"timeline.activity_synopsis_text": nil,
				"timeline.raw_activity_text":      nil,
				"timeline.capture_state":          "rough",
			},
			AfterCells: map[string]any{
				"timeline.activity_synopsis_text": "Enriched capture",
				"timeline.raw_activity_text":      "Details from patch",
				"timeline.capture_state":          "enriched",
			},
		})

		projection = asserttest.LookupProjectionRow(t, asserttest.SQLDatabase(db), recordID)
		if projection.CaptureState != "enriched" {
			t.Fatalf("expected projection to reflect enriched patch state, got %#v", projection)
		}
		countersAfterPatchState := snapshotTimelineMutationState(t, db, incidentID, recordID)
		patchReplay := httptestx.DoJSON(t, http.DefaultClient,
			http.MethodPatch,
			server.HTTP.URL+"/api/v1/records/"+recordID,
			map[string]any{
				"view_schema_id":   timeline.TimelineViewSchemaID,
				"base_row_version": 1,
				"client_txn_id":    "txn-i-3-01-row-patch",
				"changes": []map[string]any{
					{"field_key": "timeline.raw_activity_text", "value": "Details from patch"},
					{"field_key": "timeline.activity_synopsis_text", "value": "Enriched capture"},
				},
			},
			httptestx.WithCookies(adminLogin.Session, adminLogin.CSRF),
			httptestx.WithHeader(authn.CSRFHeaderName, adminLogin.CSRF.Value),
		)
		patchReplayData := httptestx.RequireSuccessEnvelope(t, patchReplay, http.StatusOK)["data"].(map[string]any)
		if patchReplayData["change_set_id"] != patchData["change_set_id"] {
			t.Fatalf("expected patch replay to return original payload, got %#v", patchReplayData)
		}
		incidentwstest.ExpectNoSocketMessage(t, socket)
		if got := snapshotTimelineMutationState(t, db, incidentID, recordID); got != countersAfterPatchState {
			t.Fatalf("replay or rejection changed mutation state: before=%+v after=%+v", countersAfterPatchState, got)
		}

		divergentPatch := httptestx.DoJSON(t, http.DefaultClient,
			http.MethodPatch,
			server.HTTP.URL+"/api/v1/records/"+recordID,
			map[string]any{
				"view_schema_id":   timeline.TimelineViewSchemaID,
				"base_row_version": 1,
				"client_txn_id":    "txn-i-3-01-row-patch",
				"changes": []map[string]any{
					{"field_key": "timeline.activity_synopsis_text", "value": "Divergent replay"},
				},
			},
			httptestx.WithCookies(adminLogin.Session, adminLogin.CSRF),
			httptestx.WithHeader(authn.CSRFHeaderName, adminLogin.CSRF.Value),
		)
		httptestx.RequireErrorEnvelope(t, divergentPatch, http.StatusConflict, "client_txn_conflict")
		if got := snapshotTimelineMutationState(t, db, incidentID, recordID); got != countersAfterPatchState {
			t.Fatalf("replay or rejection changed mutation state: before=%+v after=%+v", countersAfterPatchState, got)
		}
		incidentwstest.ExpectNoSocketMessage(t, socket)

		reviewResp := httptestx.DoJSON(t, http.DefaultClient,
			http.MethodPost,
			server.HTTP.URL+"/api/v1/records/"+recordID+"/mark-reviewed",
			map[string]any{
				"base_row_version": 2,
				"client_txn_id":    "txn-i-3-01-row-review",
				"reason":           "  Initial review  ",
			},
			httptestx.WithCookies(adminLogin.Session, adminLogin.CSRF),
			httptestx.WithHeader(authn.CSRFHeaderName, adminLogin.CSRF.Value),
		)
		reviewData := httptestx.RequireSuccessEnvelope(t, reviewResp, http.StatusOK)["data"].(map[string]any)
		if reviewData["capture_state"] != "reviewed" || reviewData["row_version"] != float64(3) || reviewData["reason"] != "Initial review" {
			t.Fatalf("unexpected review payload: %#v", reviewData)
		}
		incidentwstest.RequireRecordChanged(t, socket, recordID, 3)
		requireMutationRecorded(t, db, reviewData["change_set_id"].(string), recordID, adminID.String(), "timeline.records.mark_reviewed", "txn-i-3-01-row-review", 1, 3)
		asserttest.RequireTimelineRecordMutation(t, asserttest.SQLDatabase(db), reviewData["change_set_id"].(string), asserttest.TimelineRecordMutationExpectation{
			SequenceNo:       1,
			RecordID:         recordID,
			OperationKind:    "patch",
			BeforeRowVersion: asserttest.RowVersion(2),
			AfterRowVersion:  asserttest.RowVersion(3),
			BeforeCells:      map[string]any{"timeline.capture_state": "enriched"},
			AfterCells:       map[string]any{"timeline.capture_state": "reviewed"},
		})

		projection = asserttest.LookupProjectionRow(t, asserttest.SQLDatabase(db), recordID)
		if projection.CaptureState != "reviewed" {
			t.Fatalf("expected reviewed projection state, got %#v", projection)
		}
		reviewCountsState := snapshotTimelineMutationState(t, db, incidentID, recordID)
		reviewReplay := httptestx.DoJSON(t, http.DefaultClient,
			http.MethodPost,
			server.HTTP.URL+"/api/v1/records/"+recordID+"/mark-reviewed",
			map[string]any{
				"base_row_version": 2,
				"client_txn_id":    "txn-i-3-01-row-review",
				"reason":           "Initial review",
			},
			httptestx.WithCookies(adminLogin.Session, adminLogin.CSRF),
			httptestx.WithHeader(authn.CSRFHeaderName, adminLogin.CSRF.Value),
		)
		reviewReplayData := httptestx.RequireSuccessEnvelope(t, reviewReplay, http.StatusOK)["data"].(map[string]any)
		if reviewReplayData["change_set_id"] != reviewData["change_set_id"] {
			t.Fatalf("expected review replay to return original payload, got %#v", reviewReplayData)
		}
		incidentwstest.ExpectNoSocketMessage(t, socket)
		if got := snapshotTimelineMutationState(t, db, incidentID, recordID); got != reviewCountsState {
			t.Fatalf("replay or rejection changed mutation state: before=%+v after=%+v", reviewCountsState, got)
		}

		reviewDivergent := httptestx.DoJSON(t, http.DefaultClient,
			http.MethodPost,
			server.HTTP.URL+"/api/v1/records/"+recordID+"/mark-reviewed",
			map[string]any{
				"base_row_version": 2,
				"client_txn_id":    "txn-i-3-01-row-review",
				"reason":           "Different review reason",
			},
			httptestx.WithCookies(adminLogin.Session, adminLogin.CSRF),
			httptestx.WithHeader(authn.CSRFHeaderName, adminLogin.CSRF.Value),
		)
		httptestx.RequireErrorEnvelope(t, reviewDivergent, http.StatusConflict, "client_txn_conflict")
		if got := snapshotTimelineMutationState(t, db, incidentID, recordID); got != reviewCountsState {
			t.Fatalf("replay or rejection changed mutation state: before=%+v after=%+v", reviewCountsState, got)
		}
		incidentwstest.ExpectNoSocketMessage(t, socket)

		supersedeResp := httptestx.DoJSON(t, http.DefaultClient,
			http.MethodPost,
			server.HTTP.URL+"/api/v1/records/"+recordID+"/supersede",
			map[string]any{
				"base_row_version":      3,
				"client_txn_id":         "txn-i-3-01-row-supersede",
				"reason":                "  Superseded by a better row  ",
				"replacement_record_id": replacementID,
			},
			httptestx.WithCookies(adminLogin.Session, adminLogin.CSRF),
			httptestx.WithHeader(authn.CSRFHeaderName, adminLogin.CSRF.Value),
		)
		supersedeData := httptestx.RequireSuccessEnvelope(t, supersedeResp, http.StatusOK)["data"].(map[string]any)
		if supersedeData["capture_state"] != "superseded" || supersedeData["row_version"] != float64(4) || supersedeData["reason"] != "Superseded by a better row" || supersedeData["replacement_record_id"] != replacementID {
			t.Fatalf("unexpected supersede payload: %#v", supersedeData)
		}
		incidentwstest.RequireRecordChanged(t, socket, recordID, 4)
		requireMutationRecorded(t, db, supersedeData["change_set_id"].(string), recordID, adminID.String(), "timeline.records.supersede", "txn-i-3-01-row-supersede", 2, 4)
		asserttest.RequireSupersedeCoupledChangeSet(t, asserttest.SQLDatabase(db), supersedeData["change_set_id"].(string), recordID, replacementID, 4)
		asserttest.RequireRecordLinkCreateMutation(t, asserttest.SQLDatabase(db), supersedeData["change_set_id"].(string), asserttest.RecordLinkMutationExpectation{
			SequenceNo:          2,
			IncidentID:          incidentID,
			SourceRecordID:      replacementID,
			DestinationRecordID: recordID,
			LinkType:            "supersedes",
		})

		if got := asserttest.CountActiveSupersedesLinks(t, asserttest.SQLDatabase(db), incidentID, replacementID, recordID); got != 1 {
			t.Fatalf("expected one active supersedes link, got %d", got)
		}
		projection = asserttest.LookupProjectionRow(t, asserttest.SQLDatabase(db), recordID)
		if projection.CaptureState != "superseded" || projection.ReplacementRecordID == nil || *projection.ReplacementRecordID != replacementID {
			t.Fatalf("unexpected superseded projection row: %#v", projection)
		}
		supersedeCountsState := snapshotTimelineMutationState(t, db, incidentID, recordID)
		supersedeReplay := httptestx.DoJSON(t, http.DefaultClient,
			http.MethodPost,
			server.HTTP.URL+"/api/v1/records/"+recordID+"/supersede",
			map[string]any{
				"base_row_version":      3,
				"client_txn_id":         "txn-i-3-01-row-supersede",
				"reason":                "Superseded by a better row",
				"replacement_record_id": replacementID,
			},
			httptestx.WithCookies(adminLogin.Session, adminLogin.CSRF),
			httptestx.WithHeader(authn.CSRFHeaderName, adminLogin.CSRF.Value),
		)
		supersedeReplayData := httptestx.RequireSuccessEnvelope(t, supersedeReplay, http.StatusOK)["data"].(map[string]any)
		if supersedeReplayData["change_set_id"] != supersedeData["change_set_id"] {
			t.Fatalf("expected supersede replay to return original payload, got %#v", supersedeReplayData)
		}
		incidentwstest.ExpectNoSocketMessage(t, socket)
		if got := snapshotTimelineMutationState(t, db, incidentID, recordID); got != supersedeCountsState {
			t.Fatalf("replay or rejection changed mutation state: before=%+v after=%+v", supersedeCountsState, got)
		}

		supersedeDivergentReason := httptestx.DoJSON(t, http.DefaultClient,
			http.MethodPost,
			server.HTTP.URL+"/api/v1/records/"+recordID+"/supersede",
			map[string]any{
				"base_row_version":      3,
				"client_txn_id":         "txn-i-3-01-row-supersede",
				"reason":                "Different supersede reason",
				"replacement_record_id": replacementID,
			},
			httptestx.WithCookies(adminLogin.Session, adminLogin.CSRF),
			httptestx.WithHeader(authn.CSRFHeaderName, adminLogin.CSRF.Value),
		)
		httptestx.RequireErrorEnvelope(t, supersedeDivergentReason, http.StatusConflict, "client_txn_conflict")
		if got := snapshotTimelineMutationState(t, db, incidentID, recordID); got != supersedeCountsState {
			t.Fatalf("replay or rejection changed mutation state: before=%+v after=%+v", supersedeCountsState, got)
		}
		incidentwstest.ExpectNoSocketMessage(t, socket)

		supersedeDivergentReplacement := httptestx.DoJSON(t, http.DefaultClient,
			http.MethodPost,
			server.HTTP.URL+"/api/v1/records/"+recordID+"/supersede",
			map[string]any{
				"base_row_version":      3,
				"client_txn_id":         "txn-i-3-01-row-supersede",
				"reason":                "Superseded by a better row",
				"replacement_record_id": recordID,
			},
			httptestx.WithCookies(adminLogin.Session, adminLogin.CSRF),
			httptestx.WithHeader(authn.CSRFHeaderName, adminLogin.CSRF.Value),
		)
		httptestx.RequireErrorEnvelope(t, supersedeDivergentReplacement, http.StatusConflict, "client_txn_conflict")
		incidentwstest.ExpectNoSocketMessage(t, socket)
	})

	t.Run("late transaction failures roll back source history projection and collaboration", func(t *testing.T) {
		server, db, harness := startServer(t, runtime, "timeline_mutation-i-3-01-rollback")
		defer db.Close()

		adminLogin, adminID := flowtest.ProvisionBootstrapAdmin(t, http.DefaultClient, server.HTTP.URL)
		incident := createIncident(t, server, adminLogin, map[string]any{
			"client_txn_id": "txn-i-3-01-rollback-incident",
			"incident_key":  "IR-I301R",
			"title":         "Rollback proof",
		})
		incidentID := incident["incident_id"].(string)
		socket := incidentwstest.ConnectViewSocket(t, server, incidentID, timeline.TimelineViewSchemaID, adminLogin.Session.Value)
		defer socket.Close(1000, "test_complete")

		facade := timelineFacadeWithProjectionFailure(t, harness, func(workbookprojection.ProjectionMutation) error {
			return errors.New("forced timeline rollback")
		})
		summary := "Rollback row"
		request := timeline.CreateRequest{
			ClientTxnID:          "txn-i-3-01-rollback-row",
			ActivitySynopsisText: &summary,
		}
		_, err := facade.CreateRow(context.Background(), timeline.CreateRowCommand{
			Actor:       loadTimelineTestUser(t, harness, adminID.String()),
			IncidentID:  mustUUID(t, incidentID),
			Request:     request,
			RequestHash: timelineadmission.CreateRequestHash(request),
			RequestID:   "req-i-3-01-rollback-row",
			Now:         time.Now().UTC(),
		})
		if err == nil || !strings.Contains(err.Error(), "forced timeline rollback") {
			t.Fatalf("expected forced projection rollback, got %v", err)
		}

		if got := queryCount(t, db, `SELECT COUNT(*) FROM timeline_events WHERE incident_id::text = $1`, incidentID); got != 0 {
			t.Fatalf("rollback must clear timeline_events, got %d", got)
		}
		if got := queryCount(t, db, `SELECT COUNT(*) FROM change_sets WHERE incident_id::text = $1`, incidentID); got != 0 {
			t.Fatalf("rollback must clear change_sets, got %d", got)
		}
		if got := queryCount(t, db, `SELECT COUNT(*) FROM change_set_mutations m JOIN change_sets c ON c.change_set_id = m.change_set_id WHERE c.incident_id::text = $1`, incidentID); got != 0 {
			t.Fatalf("rollback must clear change_set_mutations, got %d", got)
		}
		if got := queryCount(t, db, `SELECT COUNT(*) FROM record_revisions rr JOIN timeline_events e ON e.record_id = rr.record_id WHERE e.incident_id::text = $1`, incidentID); got != 0 {
			t.Fatalf("rollback must clear record_revisions, got %d", got)
		}
		if got := queryCount(t, db, `SELECT COUNT(*) FROM timeline_grid_projection WHERE incident_id::text = $1`, incidentID); got != 0 {
			t.Fatalf("rollback must clear timeline_grid_projection, got %d", got)
		}
		incidentwstest.ExpectNoSocketMessage(t, socket)
		if got := collaborationsupport.CountIntents(t, db, collaborationsupport.IntentSelector{IncidentID: incidentID}); got != 0 {
			t.Fatalf("rollback left %d collaboration intents", got)
		}
	})
}

func TestRouteIdempotencyIsActorScoped(t *testing.T) {
	runtime := appsupport.StartRuntime(t)
	server, db, _ := startServer(t, runtime, "timeline_mutation-idempotency-actor-scope")
	defer db.Close()

	adminLogin, adminID := flowtest.ProvisionBootstrapAdmin(t, http.DefaultClient, server.HTTP.URL)
	editorID := seedLocalUserFlags(t, db, "timeline_mutation-actor-scope-editor@example.test", "Actor Scope Editor", "ActorScopeEditor1!", false, false, true)
	incident := createIncident(t, server, adminLogin, map[string]any{
		"client_txn_id": "txn-timeline_mutation-actor-scope-incident",
		"incident_key":  "IR-ACTOR-SCOPE",
		"title":         "Actor-scoped idempotency",
	})
	incidentID := incident["incident_id"].(string)
	createMembership(t, server, incidentID, editorID, "timeline_mutation-actor-scope-editor@example.test", "editor", adminLogin)
	editorAuth := flowtest.LoginLocalUser(t, http.DefaultClient, server.HTTP.URL, "timeline_mutation-actor-scope-editor@example.test", "ActorScopeEditor1!", nil)
	editorLogin := editorAuth

	createPayload := map[string]any{
		"client_txn_id":                   "txn-timeline_mutation-shared-row-create",
		"timeline.activity_synopsis_text": "shared create txn",
	}
	adminCreate := createTimelineRow(t, server, incidentID, adminLogin, createPayload)
	editorCreate := createTimelineRow(t, server, incidentID, editorLogin, createPayload)
	adminCreateRecordID := adminCreate["row"].(map[string]any)["record_id"].(string)
	editorCreateRecordID := editorCreate["row"].(map[string]any)["record_id"].(string)
	if adminCreateRecordID == editorCreateRecordID {
		t.Fatalf("cross-actor row create must commit independent records, got %s", adminCreateRecordID)
	}
	for _, tc := range []struct {
		login authcookietest.AuthCookies
		want  map[string]any
	}{
		{login: adminLogin, want: adminCreate},
		{login: editorLogin, want: editorCreate},
	} {
		resp := httptestx.DoJSON(t, http.DefaultClient,
			http.MethodPost,
			server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/views/"+timeline.TimelineViewSchemaID+"/rows",
			createPayload,
			httptestx.WithCookies(tc.login.Session, tc.login.CSRF),
			httptestx.WithHeader(authn.CSRFHeaderName, tc.login.CSRF.Value),
		)
		replay := httptestx.RequireSuccessEnvelope(t, resp, http.StatusOK)["data"].(map[string]any)
		if replay["change_set_id"] != tc.want["change_set_id"] {
			t.Fatalf("create replay returned wrong actor payload: got %#v want %#v", replay, tc.want)
		}
	}
	if got := queryCount(t, db, `
SELECT COUNT(*)
  FROM route_idempotency
 WHERE route_key = $1
   AND actor_user_id::text IN ($2, $3)
   AND scope_key = $4
   AND client_txn_id = $5
`, "timeline.rows.create", adminID.String(), editorID, incidentID+":"+timeline.TimelineViewSchemaID, "txn-timeline_mutation-shared-row-create"); got != 2 {
		t.Fatalf("expected two actor-scoped create idempotency rows, got %d", got)
	}

	patchTarget := createTimelineRow(t, server, incidentID, adminLogin, map[string]any{
		"client_txn_id":                   "txn-timeline_mutation-actor-scope-patch-target",
		"timeline.activity_synopsis_text": "patch target",
	})
	patchRecordID := patchTarget["row"].(map[string]any)["record_id"].(string)
	adminPatchPayload := map[string]any{
		"view_schema_id":   timeline.TimelineViewSchemaID,
		"base_row_version": 1,
		"client_txn_id":    "txn-timeline_mutation-shared-row-patch",
		"changes": []map[string]any{
			{"field_key": "timeline.activity_synopsis_text", "value": "admin patch"},
		},
	}
	adminPatchResp := httptestx.DoJSON(t, http.DefaultClient,
		http.MethodPatch,
		server.HTTP.URL+"/api/v1/records/"+patchRecordID,
		adminPatchPayload,
		httptestx.WithCookies(adminLogin.Session, adminLogin.CSRF),
		httptestx.WithHeader(authn.CSRFHeaderName, adminLogin.CSRF.Value),
	)
	adminPatch := httptestx.RequireSuccessEnvelope(t, adminPatchResp, http.StatusOK)["data"].(map[string]any)

	editorPatchPayload := map[string]any{
		"view_schema_id":   timeline.TimelineViewSchemaID,
		"base_row_version": 2,
		"client_txn_id":    "txn-timeline_mutation-shared-row-patch",
		"changes": []map[string]any{
			{"field_key": "timeline.raw_activity_text", "value": "editor patch"},
		},
	}
	editorPatchResp := httptestx.DoJSON(t, http.DefaultClient,
		http.MethodPatch,
		server.HTTP.URL+"/api/v1/records/"+patchRecordID,
		editorPatchPayload,
		httptestx.WithCookies(editorLogin.Session, editorLogin.CSRF),
		httptestx.WithHeader(authn.CSRFHeaderName, editorLogin.CSRF.Value),
	)
	editorPatch := httptestx.RequireSuccessEnvelope(t, editorPatchResp, http.StatusOK)["data"].(map[string]any)
	if adminPatch["change_set_id"] == editorPatch["change_set_id"] {
		t.Fatalf("cross-actor patch must commit independent change_sets, got %#v", adminPatch)
	}
	for _, tc := range []struct {
		login   authcookietest.AuthCookies
		payload map[string]any
		want    map[string]any
	}{
		{login: adminLogin, payload: adminPatchPayload, want: adminPatch},
		{login: editorLogin, payload: editorPatchPayload, want: editorPatch},
	} {
		resp := httptestx.DoJSON(t, http.DefaultClient,
			http.MethodPatch,
			server.HTTP.URL+"/api/v1/records/"+patchRecordID,
			tc.payload,
			httptestx.WithCookies(tc.login.Session, tc.login.CSRF),
			httptestx.WithHeader(authn.CSRFHeaderName, tc.login.CSRF.Value),
		)
		replay := httptestx.RequireSuccessEnvelope(t, resp, http.StatusOK)["data"].(map[string]any)
		if replay["change_set_id"] != tc.want["change_set_id"] {
			t.Fatalf("patch replay returned wrong actor payload: got %#v want %#v", replay, tc.want)
		}
	}
	if got := queryCount(t, db, `
SELECT COUNT(*)
  FROM route_idempotency
 WHERE route_key = $1
   AND actor_user_id::text IN ($2, $3)
   AND scope_key = $4
   AND client_txn_id = $5
`, "timeline.records.patch", adminID.String(), editorID, patchRecordID, "txn-timeline_mutation-shared-row-patch"); got != 2 {
		t.Fatalf("expected two actor-scoped patch idempotency rows, got %d", got)
	}
	if got := queryCount(t, db, `
SELECT COUNT(DISTINCT actor_user_id)
  FROM route_idempotency
 WHERE route_key IN ('timeline.rows.create', 'timeline.records.patch')
   AND client_txn_id IN ('txn-timeline_mutation-shared-row-create', 'txn-timeline_mutation-shared-row-patch')
   AND actor_user_id::text IN ($1, $2)
`, adminID.String(), editorID); got != 2 {
		t.Fatalf("expected both actors represented in idempotency rows, got %d", got)
	}
}

func TestPatchSameFieldConflictEnvelope_Integration(t *testing.T) {
	runtime := appsupport.StartRuntime(t)
	server, db, _ := startServer(t, runtime, "timeline_mutation-i-3-04-same-field-conflict")
	defer db.Close()

	adminLogin, adminID := flowtest.ProvisionBootstrapAdmin(t, http.DefaultClient, server.HTTP.URL)
	incident := createIncident(t, server, adminLogin, map[string]any{
		"client_txn_id": "txn-i-3-04-conflict-incident",
		"incident_key":  "IR-I304",
		"title":         "Same field conflict",
	})
	incidentID := incident["incident_id"].(string)
	created := createTimelineRow(t, server, incidentID, adminLogin, map[string]any{
		"client_txn_id":                   "txn-i-3-04-conflict-row",
		"timeline.activity_synopsis_text": "Base summary",
	})
	recordID := created["row"].(map[string]any)["record_id"].(string)

	serverPatch := httptestx.DoJSON(t, http.DefaultClient,
		http.MethodPatch,
		server.HTTP.URL+"/api/v1/records/"+recordID,
		map[string]any{
			"view_schema_id":   timeline.TimelineViewSchemaID,
			"base_row_version": 1,
			"client_txn_id":    "txn-i-3-04-conflict-server",
			"changes": []map[string]any{
				{"field_key": "timeline.activity_synopsis_text", "value": "Server summary"},
			},
		},
		httptestx.WithCookies(adminLogin.Session, adminLogin.CSRF),
		httptestx.WithHeader(authn.CSRFHeaderName, adminLogin.CSRF.Value),
	)
	httptestx.RequireSuccessEnvelope(t, serverPatch, http.StatusOK)
	beforeConflict := asserttest.SnapshotCounters(t, asserttest.SQLDatabase(db), incidentID, recordID)

	clientPatch := httptestx.DoJSON(t, http.DefaultClient,
		http.MethodPatch,
		server.HTTP.URL+"/api/v1/records/"+recordID,
		map[string]any{
			"view_schema_id":   timeline.TimelineViewSchemaID,
			"base_row_version": 1,
			"client_txn_id":    "txn-i-3-04-conflict-client",
			"changes": []map[string]any{
				{"field_key": "timeline.activity_synopsis_text", "value": "Client summary"},
			},
		},
		httptestx.WithCookies(adminLogin.Session, adminLogin.CSRF),
		httptestx.WithHeader(authn.CSRFHeaderName, adminLogin.CSRF.Value),
	)
	body := httptestx.RequireErrorEnvelope(t, clientPatch, http.StatusConflict, "same_field_conflict")
	errorObject := body["error"].(map[string]any)
	conflict, ok := errorObject["conflict"].(map[string]any)
	if !ok {
		t.Fatalf("expected same-field conflict object, got %#v", errorObject)
	}
	if conflict["record_id"] != recordID ||
		conflict["field_key"] != "timeline.activity_synopsis_text" ||
		conflict["conflict_resolution_class"] != "text_compare_merge" ||
		conflict["base_row_version"] != float64(1) ||
		conflict["current_row_version"] != float64(2) ||
		conflict["base_value"] != "Base summary" ||
		conflict["server_value"] != "Server summary" ||
		conflict["client_value"] != "Client summary" ||
		conflict["server_updated_by"] != adminID.String() ||
		conflict["server_updated_at"] == "" ||
		conflict["conflict_token"] == "" {
		t.Fatalf("unexpected same-field conflict object: %#v", conflict)
	}
	afterConflict := asserttest.SnapshotCounters(t, asserttest.SQLDatabase(db), incidentID, recordID)
	if beforeConflict != afterConflict {
		t.Fatalf("same-field HTTP conflict must not create writes: before=%#v after=%#v", beforeConflict, afterConflict)
	}
	if got := queryCount(t, db, `
SELECT COUNT(*)
  FROM route_idempotency
 WHERE route_key = 'timeline.records.patch'
   AND actor_user_id::text = $1
   AND scope_key = $2
   AND client_txn_id = $3
`, adminID.String(), recordID, "txn-i-3-04-conflict-client"); got != 0 {
		t.Fatalf("same-field HTTP conflict must not persist idempotency row, got %d", got)
	}
}

func TestRouteEnvelopeMatrix_Integration(t *testing.T) {
	runtime := appsupport.StartRuntime(t)
	server, db, _ := startServer(t, runtime, "timeline_mutation-i-3-06-envelope-matrix")
	defer db.Close()

	adminLogin, _ := flowtest.ProvisionBootstrapAdmin(t, http.DefaultClient, server.HTTP.URL)
	incident := createIncident(t, server, adminLogin, map[string]any{
		"client_txn_id": "txn-i-3-06-incident",
		"incident_key":  "IR-I306",
		"title":         "Malformed envelope matrix",
	})
	incidentID := incident["incident_id"].(string)
	created := createTimelineRow(t, server, incidentID, adminLogin, map[string]any{
		"client_txn_id":                   "txn-i-3-06-row",
		"timeline.activity_synopsis_text": "Envelope row",
	})
	recordID := created["row"].(map[string]any)["record_id"].(string)
	replacement := createTimelineRow(t, server, incidentID, adminLogin, map[string]any{
		"client_txn_id":                   "txn-i-3-06-replacement",
		"timeline.activity_synopsis_text": "Envelope replacement",
	})
	replacementID := replacement["row"].(map[string]any)["record_id"].(string)

	authOptions := []func(*http.Request){
		httptestx.WithCookies(adminLogin.Session, adminLogin.CSRF),
		httptestx.WithHeader(authn.CSRFHeaderName, adminLogin.CSRF.Value),
	}
	queryOptions := []func(*http.Request){httptestx.WithCookies(adminLogin.Session)}
	cases := []struct {
		name       string
		method     string
		url        string
		body       string
		options    []func(*http.Request)
		status     int
		code       string
		detailWant map[string]any
		mutating   bool
	}{
		{
			name:       "create malformed JSON",
			method:     http.MethodPost,
			url:        server.HTTP.URL + "/api/v1/incidents/" + incidentID + "/views/" + timeline.TimelineViewSchemaID + "/rows",
			body:       `{"client_txn_id":`,
			options:    authOptions,
			status:     http.StatusBadRequest,
			code:       "invalid_mutation_payload",
			detailWant: map[string]any{"reason_code": "request_not_object"},
			mutating:   true,
		},
		{
			name:       "create unknown field",
			method:     http.MethodPost,
			url:        server.HTTP.URL + "/api/v1/incidents/" + incidentID + "/views/" + timeline.TimelineViewSchemaID + "/rows",
			body:       `{"client_txn_id":"txn-i-3-06-create-unknown","timeline.activity_synopsis_text":"x","timeline.unknown":true}`,
			options:    authOptions,
			status:     http.StatusBadRequest,
			code:       "invalid_mutation_payload",
			detailWant: map[string]any{"field": "timeline.unknown", "reason_code": "unknown_field"},
			mutating:   true,
		},
		{
			name:       "patch malformed JSON",
			method:     http.MethodPatch,
			url:        server.HTTP.URL + "/api/v1/records/" + recordID,
			body:       `{"view_schema_id":"` + timeline.TimelineViewSchemaID + `",`,
			options:    authOptions,
			status:     http.StatusBadRequest,
			code:       "invalid_mutation_payload",
			detailWant: map[string]any{"reason_code": "request_not_object"},
			mutating:   true,
		},
		{
			name:       "patch missing changes",
			method:     http.MethodPatch,
			url:        server.HTTP.URL + "/api/v1/records/" + recordID,
			body:       `{"view_schema_id":"` + timeline.TimelineViewSchemaID + `","base_row_version":1,"client_txn_id":"txn-i-3-06-patch-missing"}`,
			options:    authOptions,
			status:     http.StatusBadRequest,
			code:       "invalid_mutation_payload",
			detailWant: map[string]any{"field": "changes", "reason_code": "missing_required_field"},
			mutating:   true,
		},
		{
			name:       "mark reviewed malformed JSON",
			method:     http.MethodPost,
			url:        server.HTTP.URL + "/api/v1/records/" + recordID + "/mark-reviewed",
			body:       `[`,
			options:    authOptions,
			status:     http.StatusBadRequest,
			code:       "invalid_mutation_payload",
			detailWant: map[string]any{"reason_code": "request_not_object"},
			mutating:   true,
		},
		{
			name:       "supersede invalid reason",
			method:     http.MethodPost,
			url:        server.HTTP.URL + "/api/v1/records/" + recordID + "/supersede",
			body:       `{"base_row_version":1,"client_txn_id":"txn-i-3-06-supersede-invalid","reason":5,"replacement_record_id":"` + replacementID + `"}`,
			options:    authOptions,
			status:     http.StatusBadRequest,
			code:       "invalid_mutation_payload",
			detailWant: map[string]any{"field": "reason", "reason_code": "invalid_value"},
			mutating:   true,
		},
		{
			name:       "supersede explicit null replacement",
			method:     http.MethodPost,
			url:        server.HTTP.URL + "/api/v1/records/" + recordID + "/supersede",
			body:       `{"base_row_version":1,"client_txn_id":"txn-i-3-06-supersede-null-replacement","reason":"replacement omitted by null","replacement_record_id":null}`,
			options:    authOptions,
			status:     http.StatusBadRequest,
			code:       "invalid_mutation_payload",
			detailWant: map[string]any{"field": "replacement_record_id", "reason_code": "field_not_nullable"},
			mutating:   true,
		},
		{
			name:       "query malformed JSON",
			method:     http.MethodPost,
			url:        server.HTTP.URL + "/api/v1/incidents/" + incidentID + "/views/" + timeline.TimelineViewSchemaID + "/query",
			body:       `[`,
			options:    queryOptions,
			status:     http.StatusBadRequest,
			code:       "invalid_view_query",
			detailWant: map[string]any{"reason_code": "request_not_object"},
		},
		{
			name:       "query unknown member",
			method:     http.MethodPost,
			url:        server.HTTP.URL + "/api/v1/incidents/" + incidentID + "/views/" + timeline.TimelineViewSchemaID + "/query",
			body:       `{"unknown":true}`,
			options:    queryOptions,
			status:     http.StatusBadRequest,
			code:       "invalid_view_query",
			detailWant: map[string]any{"field": "unknown", "reason_code": "unknown_field"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var before asserttest.Counters
			if tc.mutating {
				before = asserttest.SnapshotCounters(t, asserttest.SQLDatabase(db), incidentID, recordID)
			}
			resp := httptestx.DoRawJSON(t, http.DefaultClient, tc.method, tc.url, tc.body, tc.options...)
			body := httptestx.RequireErrorEnvelope(t, resp, tc.status, tc.code)
			for key, want := range tc.detailWant {
				httptestx.RequireErrorDetail(t, body, key, want)
			}
			if tc.mutating {
				after := asserttest.SnapshotCounters(t, asserttest.SQLDatabase(db), incidentID, recordID)
				if before != after {
					t.Fatalf("malformed route request must not mutate history: before=%#v after=%#v", before, after)
				}
			}
		})
	}

	incidentwstest.RequireDialErrorEnvelope(t, server.HTTP.URL, incidentID, incidentwstest.ConnectOptions{}, http.StatusUnauthorized, "session_required")
}

func TestRoughUncertainCapturePreservation_Integration(t *testing.T) {
	runtime := appsupport.StartRuntime(t)
	server, db, _ := startServer(t, runtime, "timeline_mutation-i-3-07-rough-preservation")
	defer db.Close()

	adminLogin, adminID := flowtest.ProvisionBootstrapAdmin(t, http.DefaultClient, server.HTTP.URL)
	incident := createIncident(t, server, adminLogin, map[string]any{
		"client_txn_id": "txn-i-3-07-incident",
		"incident_key":  "IR-I307",
		"title":         "Rough uncertainty preservation",
	})
	incidentID := incident["incident_id"].(string)
	details := "Analyst pasted partial notes: maybe gateway, owner unknown."
	sourceText := "raw paste: host?  vpn   gateway ; acct maybe pending"
	rawHostText := " vpn   gateway "
	created := createTimelineRow(t, server, incidentID, adminLogin, map[string]any{
		"client_txn_id":                   "txn-i-3-07-rough-row",
		"timeline.activity_synopsis_text": details,
		"timeline.raw_activity_text":      sourceText,
		"timeline.host_refs": map[string]any{
			"kind": "collection_actions_v1",
			"actions": []map[string]any{
				{"op": "add_token", "raw_text": rawHostText},
			},
		},
	})
	row := created["row"].(map[string]any)
	recordID := row["record_id"].(string)
	cells := row["cells"].(map[string]any)
	if cells["timeline.activity_utc_text"].(map[string]any)["value"] != nil ||
		cells["timeline.activity_synopsis_text"].(map[string]any)["value"] != details ||
		cells["timeline.raw_activity_text"].(map[string]any)["value"] != sourceText {
		t.Fatalf("rough create did not preserve uncertain/null capture fields: %#v", row)
	}
	if cells["timeline.has_unresolved_mentions"].(map[string]any)["value"] != true {
		t.Fatalf("rough create should surface unresolved mention state, got %#v", cells["timeline.has_unresolved_mentions"])
	}
	hostItem := requireSingleCollectionItem(t, row, "timeline.host_refs")
	if hostItem["item_kind"] != "unresolved_mention" || hostItem["raw_text"] != rawHostText {
		t.Fatalf("expected unresolved host mention with raw text preserved, got %#v", hostItem)
	}
	mentionID := mentionIDFromItemRef(t, hostItem["item_ref"].(string))
	mentionBefore := lookupMention(t, db, mentionID)
	if mentionBefore.RawText != rawHostText || mentionBefore.ResolutionStatus != "unresolved" || mentionBefore.RowVersion != 1 {
		t.Fatalf("unexpected rough mention before resolution: %#v", mentionBefore)
	}

	hostID := uuid.New()
	seedHostRecord(t, db, mustUUID(t, incidentID), adminID, hostID, "VPN Gateway", "vpn-gateway")
	resolveResp := httptestx.DoJSON(t, http.DefaultClient,
		http.MethodPost,
		server.HTTP.URL+"/api/v1/entity-mentions/"+mentionID.String()+"/resolve",
		map[string]any{
			"base_mention_row_version": 1,
			"client_txn_id":            "txn-i-3-07-resolve-host",
			"action":                   "resolve_item",
			"resolved_record_id":       hostID.String(),
		},
		httptestx.WithCookies(adminLogin.Session, adminLogin.CSRF),
		httptestx.WithHeader(authn.CSRFHeaderName, adminLogin.CSRF.Value),
	)
	if resolveResp.StatusCode != http.StatusOK {
		t.Fatalf("resolve mention failed: status=%d body=%#v", resolveResp.StatusCode, httptestx.ReadJSONBody(t, resolveResp))
	}
	resolveData := httptestx.RequireSuccessEnvelope(t, resolveResp, http.StatusOK)["data"].(map[string]any)
	if resolveData["source_record"].(map[string]any)["row_version"] != float64(2) {
		t.Fatalf("expected mention resolution to advance source row_version, got %#v", resolveData)
	}

	mentionAfter := lookupMention(t, db, mentionID)
	if mentionAfter.RawText != rawHostText || mentionAfter.ResolutionStatus != "resolved" || mentionAfter.ResolvedRecordID == nil || *mentionAfter.ResolvedRecordID != hostID {
		t.Fatalf("mention resolution must preserve raw text while resolving target, got %#v", mentionAfter)
	}
	refreshed := findRow(t, queryTimelineRows(t, server, incidentID, adminLogin), recordID)
	refreshedCells := refreshed["cells"].(map[string]any)
	if refreshedCells["timeline.activity_synopsis_text"].(map[string]any)["value"] != details ||
		refreshedCells["timeline.raw_activity_text"].(map[string]any)["value"] != sourceText {
		t.Fatalf("resolution must not overwrite original rough capture fields, got %#v", refreshed)
	}
	refreshedHostItem := requireSingleCollectionItem(t, refreshed, "timeline.host_refs")
	if refreshedHostItem["item_kind"] != "resolved_ref" ||
		refreshedHostItem["raw_text"] != rawHostText ||
		refreshedHostItem["resolved_record_id"] != hostID.String() {
		t.Fatalf("resolved row must preserve original host token lineage, got %#v", refreshedHostItem)
	}
}

func TestProjectionQueryUsesDeterministicRebuild_Integration(t *testing.T) {
	runtime := appsupport.StartRuntime(t)

	server, db, harness := startServer(t, runtime, "timeline_mutation-i-3-02")
	defer db.Close()

	adminLogin, _ := flowtest.ProvisionBootstrapAdmin(t, http.DefaultClient, server.HTTP.URL)
	incident := createIncident(t, server, adminLogin, map[string]any{
		"client_txn_id": "txn-i-3-02-incident",
		"incident_key":  "IR-I302",
		"title":         "Projection reads",
	})
	incidentID := incident["incident_id"].(string)

	first := createTimelineRow(t, server, incidentID, adminLogin, map[string]any{
		"client_txn_id":                   "txn-i-3-02-row-a",
		"timeline.activity_synopsis_text": "Tie A",
		"timeline.activity_utc_text":      "2026-04-10T10:00:00Z",
	})
	second := createTimelineRow(t, server, incidentID, adminLogin, map[string]any{
		"client_txn_id":                   "txn-i-3-02-row-b",
		"timeline.activity_synopsis_text": "Tie B",
		"timeline.activity_utc_text":      "2026-04-10T10:00:00Z",
	})
	zeroField := createTimelineRow(t, server, incidentID, adminLogin, map[string]any{
		"client_txn_id": "txn-i-3-02-row-c",
	})
	firstID := first["row"].(map[string]any)["record_id"].(string)
	secondID := second["row"].(map[string]any)["record_id"].(string)
	zeroFieldID := zeroField["row"].(map[string]any)["record_id"].(string)

	patch := httptestx.DoJSON(t, http.DefaultClient,
		http.MethodPatch,
		server.HTTP.URL+"/api/v1/records/"+secondID,
		map[string]any{
			"view_schema_id":   timeline.TimelineViewSchemaID,
			"base_row_version": 1,
			"client_txn_id":    "txn-i-3-02-row-b-patch",
			"changes": []map[string]any{
				{"field_key": "timeline.raw_activity_text", "value": "Projected details"},
			},
		},
		httptestx.WithCookies(adminLogin.Session, adminLogin.CSRF),
		httptestx.WithHeader(authn.CSRFHeaderName, adminLogin.CSRF.Value),
	)
	httptestx.RequireSuccessEnvelope(t, patch, http.StatusOK)

	beforeEnvelope := queryTimelineEnvelope(t, server, incidentID, adminLogin)
	contractassert.RequireDefaultQueryMeta(t, beforeEnvelope, timeline.TimelineViewSchemaID)
	rawBeforeRows := beforeEnvelope["data"].(map[string]any)["rows"].([]any)
	beforeRows := make([]map[string]any, 0, len(rawBeforeRows))
	for _, rawRow := range rawBeforeRows {
		beforeRows = append(beforeRows, rawRow.(map[string]any))
	}
	if len(beforeRows) != 3 {
		t.Fatalf("expected three projected rows, got %#v", beforeRows)
	}
	expectedFirstID, expectedSecondID := firstID, secondID
	if expectedSecondID < expectedFirstID {
		expectedFirstID, expectedSecondID = expectedSecondID, expectedFirstID
	}
	if got := beforeRows[0]["record_id"]; got != expectedFirstID {
		t.Fatalf("expected first default-sorted row %s, got %#v", expectedFirstID, beforeRows)
	}
	if got := beforeRows[1]["record_id"]; got != expectedSecondID {
		t.Fatalf("expected second default-sorted row %s, got %#v", expectedSecondID, beforeRows)
	}
	if got := beforeRows[2]["record_id"]; got != zeroFieldID {
		t.Fatalf("expected zero-field create to sort after explicit occurred_at rows, got %#v", beforeRows)
	}
	secondRow := findRow(t, beforeRows, secondID)
	if secondRow["row_version"] != float64(2) {
		t.Fatalf("expected patched row_version in projection query, got %#v", secondRow)
	}
	if summary := secondRow["cells"].(map[string]any)["timeline.activity_synopsis_text"].(map[string]any)["value"]; summary != "Tie B" {
		t.Fatalf("expected projected summary, got %#v", secondRow)
	}
	if details := secondRow["cells"].(map[string]any)["timeline.raw_activity_text"].(map[string]any)["value"]; details != "Projected details" {
		t.Fatalf("expected projection-backed details, got %#v", secondRow)
	}
	if captureState := secondRow["group_values"].(map[string]any)["timeline.capture_state"]; captureState != "enriched" {
		t.Fatalf("expected projected group_values capture state, got %#v", secondRow["group_values"])
	}
	if enteredDay := secondRow["group_values"].(map[string]any)["timeline.date_entered_sort_day"]; enteredDay != nil {
		t.Fatalf("expected date_entered_sort_day to stay null when Date Entered is unauthored, got %#v", secondRow["group_values"])
	}
	zeroFieldRow := findRow(t, beforeRows, zeroFieldID)
	if zeroFieldRow["cells"].(map[string]any)["timeline.activity_synopsis_text"].(map[string]any)["value"] != nil {
		t.Fatalf("expected zero-field query row summary to remain null, got %#v", zeroFieldRow)
	}
	if zeroFieldRow["cells"].(map[string]any)["timeline.replacement_record_id"].(map[string]any)["value"] != nil {
		t.Fatalf("expected zero-field query row replacement cell to remain null, got %#v", zeroFieldRow)
	}
	if zeroFieldRow["group_values"].(map[string]any)["timeline.capture_state"] != "rough" {
		t.Fatalf("expected zero-field query row to remain rough, got %#v", zeroFieldRow)
	}

	sortedQuery := httptestx.DoJSON(t, http.DefaultClient,
		http.MethodPost,
		server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/views/"+timeline.TimelineViewSchemaID+"/query",
		map[string]any{
			"sort": []map[string]any{{"field_key": "timeline.activity_synopsis_text", "direction": "asc"}},
		},
		httptestx.WithCookies(adminLogin.Session),
	)
	sortedEnvelope := httptestx.RequireSuccessEnvelope(t, sortedQuery, http.StatusOK)
	sortedRows := sortedEnvelope["data"].(map[string]any)["rows"].([]any)
	if got := sortedRows[0].(map[string]any)["record_id"]; got != firstID {
		t.Fatalf("expected timeline.activity_synopsis_text asc to place Tie A first, got %#v", sortedRows)
	}
	if got := sortedRows[1].(map[string]any)["record_id"]; got != secondID {
		t.Fatalf("expected timeline.activity_synopsis_text asc to place Tie B second, got %#v", sortedRows)
	}
	if got := sortedRows[2].(map[string]any)["record_id"]; got != zeroFieldID {
		t.Fatalf("expected null summary row to sort last, got %#v", sortedRows)
	}

	if _, err := db.ExecContext(context.Background(), `DELETE FROM timeline_grid_projection WHERE incident_id::text = $1`, incidentID); err != nil {
		t.Fatalf("clear projection rows: %v", err)
	}
	emptyEnvelope := queryTimelineEnvelope(t, server, incidentID, adminLogin)
	contractassert.RequireDefaultQueryMeta(t, emptyEnvelope, timeline.TimelineViewSchemaID)
	emptyRows := emptyEnvelope["data"].(map[string]any)["rows"].([]any)
	if len(emptyRows) != 0 {
		t.Fatalf("query route must read projection rows, got %#v", emptyRows)
	}

	if err := harness.Projections.RebuildIncident(context.Background(), mustUUID(t, incidentID)); err != nil {
		t.Fatalf("rebuild timeline projection: %v", err)
	}
	rebuiltEnvelope := queryTimelineEnvelope(t, server, incidentID, adminLogin)
	contractassert.RequireDefaultQueryMeta(t, rebuiltEnvelope, timeline.TimelineViewSchemaID)
	rebuiltRows := queryTimelineRows(t, server, incidentID, adminLogin)
	contractassert.RequireProjectionDeterminism(t, beforeRows, rebuiltRows)

	if got := queryCount(t, db, `SELECT COUNT(*) FROM timeline_events WHERE incident_id::text = $1`, incidentID); got != 3 {
		t.Fatalf("rebuild must preserve source rows, got %d", got)
	}
}

func TestAuthorizationLifecycleAndSupersedeTransitions_Integration(t *testing.T) {
	runtime := appsupport.StartRuntime(t)

	t.Run("authorization re-derives, reasons normalize, and supersede guards hold", func(t *testing.T) {
		server, db, _ := startServer(t, runtime, "timeline_mutation-i-3-03")
		defer db.Close()

		adminLogin, adminID := flowtest.ProvisionBootstrapAdmin(t, http.DefaultClient, server.HTTP.URL)
		reviewerID := seedLocalUserFlags(t, db, "reviewer-target@example.test", "Reviewer Target", "ReviewerTargetPass1!", false, false, true)

		incident := createIncident(t, server, adminLogin, map[string]any{
			"client_txn_id": "txn-i-3-03-incident",
			"incident_key":  "IR-I303",
			"title":         "Lifecycle",
		})
		incidentID := incident["incident_id"].(string)
		createMembership(t, server, incidentID, reviewerID, "reviewer-target@example.test", "editor", adminLogin)

		otherIncident := createIncident(t, server, adminLogin, map[string]any{
			"client_txn_id": "txn-i-3-03-other-incident",
			"incident_key":  "IR-I303X",
			"title":         "Other incident",
		})
		otherIncidentID := otherIncident["incident_id"].(string)

		replacement := createTimelineRow(t, server, incidentID, adminLogin, map[string]any{
			"client_txn_id":                   "txn-i-3-03-replacement",
			"timeline.activity_synopsis_text": "Replacement row",
		})
		replacementID := replacement["row"].(map[string]any)["record_id"].(string)
		alternateReplacement := createTimelineRow(t, server, incidentID, adminLogin, map[string]any{
			"client_txn_id":                   "txn-i-3-03-alternate-replacement",
			"timeline.activity_synopsis_text": "Alternate replacement row",
		})
		alternateReplacementID := alternateReplacement["row"].(map[string]any)["record_id"].(string)
		created := createTimelineRow(t, server, incidentID, adminLogin, map[string]any{
			"client_txn_id":                   "txn-i-3-03-primary",
			"timeline.activity_synopsis_text": "Primary row",
		})
		recordID := created["row"].(map[string]any)["record_id"].(string)
		otherReplacement := createTimelineRow(t, server, otherIncidentID, adminLogin, map[string]any{
			"client_txn_id":                   "txn-i-3-03-cross-incident",
			"timeline.activity_synopsis_text": "Cross incident replacement",
		})
		otherReplacementID := otherReplacement["row"].(map[string]any)["record_id"].(string)
		supersededReplacement := createTimelineRow(t, server, incidentID, adminLogin, map[string]any{
			"client_txn_id":                   "txn-i-3-03-superseded-replacement",
			"timeline.activity_synopsis_text": "Superseded replacement row",
		})
		supersededReplacementID := supersededReplacement["row"].(map[string]any)["record_id"].(string)
		supersededReplacementNext := createTimelineRow(t, server, incidentID, adminLogin, map[string]any{
			"client_txn_id":                   "txn-i-3-03-superseded-replacement-next",
			"timeline.activity_synopsis_text": "Replacement for superseded replacement",
		})
		supersededReplacementNextID := supersededReplacementNext["row"].(map[string]any)["record_id"].(string)
		supersedeReplacementFixture := httptestx.DoJSON(t, http.DefaultClient,
			http.MethodPost,
			server.HTTP.URL+"/api/v1/records/"+supersededReplacementID+"/supersede",
			map[string]any{
				"base_row_version":      1,
				"client_txn_id":         "txn-i-3-03-supersede-replacement-fixture",
				"reason":                "make replacement superseded",
				"replacement_record_id": supersededReplacementNextID,
			},
			httptestx.WithCookies(adminLogin.Session, adminLogin.CSRF),
			httptestx.WithHeader(authn.CSRFHeaderName, adminLogin.CSRF.Value),
		)
		httptestx.RequireSuccessEnvelope(t, supersedeReplacementFixture, http.StatusOK)
		activeIncomingTarget := createTimelineRow(t, server, incidentID, adminLogin, map[string]any{
			"client_txn_id":                   "txn-i-3-03-active-incoming-target",
			"timeline.activity_synopsis_text": "Target with incoming replacement",
		})
		activeIncomingTargetID := activeIncomingTarget["row"].(map[string]any)["record_id"].(string)
		activeIncomingReplacement := createTimelineRow(t, server, incidentID, adminLogin, map[string]any{
			"client_txn_id":                   "txn-i-3-03-active-incoming-replacement",
			"timeline.activity_synopsis_text": "Existing incoming replacement",
		})
		activeIncomingReplacementID := activeIncomingReplacement["row"].(map[string]any)["record_id"].(string)
		if _, err := db.ExecContext(context.Background(), `
INSERT INTO record_links (incident_id, src_record_id, dst_record_id, link_type, provenance, owner_user_id, created_by_user_id)
VALUES ($1, $2, $3, 'supersedes', 'manual', $4, $4)
`, incidentID, activeIncomingReplacementID, activeIncomingTargetID, adminID.String()); err != nil {
			t.Fatalf("seed active incoming supersedes link: %v", err)
		}

		reviewerAuth := flowtest.LoginLocalUser(t, http.DefaultClient, server.HTTP.URL, "reviewer-target@example.test", "ReviewerTargetPass1!", nil)
		reviewerLogin := reviewerAuth
		asserttest.AwaitIncidentStreamIdle(t, asserttest.SQLDatabase(db), incidentID)
		socket := incidentwstest.ConnectViewSocket(t, server, incidentID, timeline.TimelineViewSchemaID, reviewerAuth.Session.Value)
		defer socket.Close(1000, "test_complete")

		reviewDeniedState := snapshotTimelineMutationState(t, db, incidentID, recordID)
		reviewDenied := httptestx.DoJSON(t, http.DefaultClient,
			http.MethodPost,
			server.HTTP.URL+"/api/v1/records/"+recordID+"/mark-reviewed",
			map[string]any{
				"base_row_version": 1,
				"client_txn_id":    "txn-i-3-03-reviewed-denied",
				"reason":           "editor cannot review",
			},
			httptestx.WithCookies(reviewerAuth.Session, reviewerAuth.CSRF),
			httptestx.WithHeader(authn.CSRFHeaderName, reviewerAuth.CSRF.Value),
		)
		httptestx.RequireErrorEnvelope(t, reviewDenied, http.StatusForbidden, "authorization_denied")
		if got := snapshotTimelineMutationState(t, db, incidentID, recordID); got != reviewDeniedState {
			t.Fatalf("denied request changed mutation state: before=%+v after=%+v", reviewDeniedState, got)
		}
		incidentwstest.ExpectNoSocketMessage(t, socket)

		updateMembershipRole(t, server, incidentID, reviewerID, 1, "reviewer", adminLogin)

		markReviewed := httptestx.DoJSON(t, http.DefaultClient,
			http.MethodPost,
			server.HTTP.URL+"/api/v1/records/"+recordID+"/mark-reviewed",
			map[string]any{
				"base_row_version": 1,
				"client_txn_id":    "txn-i-3-03-reviewed-1",
			},
			httptestx.WithCookies(reviewerAuth.Session, reviewerAuth.CSRF),
			httptestx.WithHeader(authn.CSRFHeaderName, reviewerAuth.CSRF.Value),
		)
		reviewedData := httptestx.RequireSuccessEnvelope(t, markReviewed, http.StatusOK)["data"].(map[string]any)
		if reviewedData["capture_state"] != "reviewed" || reviewedData["row_version"] != float64(2) || reviewedData["reason"] != nil {
			t.Fatalf("unexpected reviewed payload: %#v", reviewedData)
		}
		incidentwstest.RequireRecordChanged(t, socket, recordID, 2)
		requireMutationRecorded(t, db, reviewedData["change_set_id"].(string), recordID, reviewerID, "timeline.records.mark_reviewed", "txn-i-3-03-reviewed-1", 1, 2)
		if projection := asserttest.LookupProjectionRow(t, asserttest.SQLDatabase(db), recordID); projection.CaptureState != "reviewed" {
			t.Fatalf("expected reviewed projection row, got %#v", projection)
		}
		reviewCountsState := snapshotTimelineMutationState(t, db, incidentID, recordID)
		reviewReplay := httptestx.DoJSON(t, http.DefaultClient,
			http.MethodPost,
			server.HTTP.URL+"/api/v1/records/"+recordID+"/mark-reviewed",
			map[string]any{
				"base_row_version": 1,
				"client_txn_id":    "txn-i-3-03-reviewed-1",
				"reason":           nil,
			},
			httptestx.WithCookies(reviewerAuth.Session, reviewerAuth.CSRF),
			httptestx.WithHeader(authn.CSRFHeaderName, reviewerAuth.CSRF.Value),
		)
		reviewReplayData := httptestx.RequireSuccessEnvelope(t, reviewReplay, http.StatusOK)["data"].(map[string]any)
		if reviewReplayData["change_set_id"] != reviewedData["change_set_id"] {
			t.Fatalf("expected review replay to return original payload, got %#v", reviewReplayData)
		}
		incidentwstest.ExpectNoSocketMessage(t, socket)
		if got := snapshotTimelineMutationState(t, db, incidentID, recordID); got != reviewCountsState {
			t.Fatalf("replay or rejection changed mutation state: before=%+v after=%+v", reviewCountsState, got)
		}

		reviewDivergent := httptestx.DoJSON(t, http.DefaultClient,
			http.MethodPost,
			server.HTTP.URL+"/api/v1/records/"+recordID+"/mark-reviewed",
			map[string]any{
				"base_row_version": 1,
				"client_txn_id":    "txn-i-3-03-reviewed-1",
				"reason":           "Different review reason",
			},
			httptestx.WithCookies(reviewerAuth.Session, reviewerAuth.CSRF),
			httptestx.WithHeader(authn.CSRFHeaderName, reviewerAuth.CSRF.Value),
		)
		httptestx.RequireErrorEnvelope(t, reviewDivergent, http.StatusConflict, "client_txn_conflict")
		if got := snapshotTimelineMutationState(t, db, incidentID, recordID); got != reviewCountsState {
			t.Fatalf("replay or rejection changed mutation state: before=%+v after=%+v", reviewCountsState, got)
		}
		incidentwstest.ExpectNoSocketMessage(t, socket)

		materialEdit := httptestx.DoJSON(t, http.DefaultClient,
			http.MethodPatch,
			server.HTTP.URL+"/api/v1/records/"+recordID,
			map[string]any{
				"view_schema_id":   timeline.TimelineViewSchemaID,
				"base_row_version": 2,
				"client_txn_id":    "txn-i-3-03-demote",
				"changes": []map[string]any{
					{"field_key": "timeline.raw_activity_text", "value": "Material edit after review"},
				},
			},
			httptestx.WithCookies(reviewerAuth.Session, reviewerAuth.CSRF),
			httptestx.WithHeader(authn.CSRFHeaderName, reviewerAuth.CSRF.Value),
		)
		demoted := httptestx.RequireSuccessEnvelope(t, materialEdit, http.StatusOK)["data"].(map[string]any)
		demotedRow := demoted["row"].(map[string]any)
		if got := demotedRow["cells"].(map[string]any)["timeline.capture_state"].(map[string]any)["value"]; got != "enriched" {
			t.Fatalf("expected reviewed row to demote back to enriched, got %#v", demotedRow)
		}
		incidentwstest.RequireRecordChanged(t, socket, recordID, 3)
		requireMutationRecorded(t, db, demoted["change_set_id"].(string), recordID, reviewerID, "timeline.records.patch", "txn-i-3-03-demote", 1, 3)
		asserttest.RequireTimelineRecordMutation(t, asserttest.SQLDatabase(db), demoted["change_set_id"].(string), asserttest.TimelineRecordMutationExpectation{
			SequenceNo:       1,
			RecordID:         recordID,
			OperationKind:    "patch",
			BeforeRowVersion: asserttest.RowVersion(2),
			AfterRowVersion:  asserttest.RowVersion(3),
			BeforeCells:      map[string]any{"timeline.capture_state": "reviewed", "timeline.raw_activity_text": nil},
			AfterCells:       map[string]any{"timeline.capture_state": "enriched", "timeline.raw_activity_text": "Material edit after review"},
		})

		selfBefore := asserttest.SnapshotCounters(t, asserttest.SQLDatabase(db), incidentID, recordID)
		selfSupersede := httptestx.DoJSON(t, http.DefaultClient,
			http.MethodPost,
			server.HTTP.URL+"/api/v1/records/"+recordID+"/supersede",
			map[string]any{
				"base_row_version":      3,
				"client_txn_id":         "txn-i-3-03-self-supersede",
				"reason":                "self must fail",
				"replacement_record_id": recordID,
			},
			httptestx.WithCookies(reviewerAuth.Session, reviewerAuth.CSRF),
			httptestx.WithHeader(authn.CSRFHeaderName, reviewerAuth.CSRF.Value),
		)
		requireRejectedMutationStable(t, db, incidentID, recordID, selfBefore, selfSupersede, "enriched", "superseded", []string{"replacement_must_be_different_timeline_record"})
		incidentwstest.ExpectNoSocketMessage(t, socket)

		crossBefore := asserttest.SnapshotCounters(t, asserttest.SQLDatabase(db), incidentID, recordID)
		crossIncidentSupersede := httptestx.DoJSON(t, http.DefaultClient,
			http.MethodPost,
			server.HTTP.URL+"/api/v1/records/"+recordID+"/supersede",
			map[string]any{
				"base_row_version":      3,
				"client_txn_id":         "txn-i-3-03-cross-supersede",
				"reason":                "cross incident must fail",
				"replacement_record_id": otherReplacementID,
			},
			httptestx.WithCookies(reviewerAuth.Session, reviewerAuth.CSRF),
			httptestx.WithHeader(authn.CSRFHeaderName, reviewerAuth.CSRF.Value),
		)
		requireRejectedMutationStable(t, db, incidentID, recordID, crossBefore, crossIncidentSupersede, "enriched", "superseded", []string{"replacement_must_be_visible_active_same_incident_timeline_record"})
		incidentwstest.ExpectNoSocketMessage(t, socket)

		supersededReplacementBefore := asserttest.SnapshotCounters(t, asserttest.SQLDatabase(db), incidentID, recordID)
		supersededReplacementResp := httptestx.DoJSON(t, http.DefaultClient,
			http.MethodPost,
			server.HTTP.URL+"/api/v1/records/"+recordID+"/supersede",
			map[string]any{
				"base_row_version":      3,
				"client_txn_id":         "txn-i-3-03-superseded-replacement-target",
				"reason":                "superseded replacement must fail",
				"replacement_record_id": supersededReplacementID,
			},
			httptestx.WithCookies(reviewerAuth.Session, reviewerAuth.CSRF),
			httptestx.WithHeader(authn.CSRFHeaderName, reviewerAuth.CSRF.Value),
		)
		requireRejectedMutationStable(t, db, incidentID, recordID, supersededReplacementBefore, supersededReplacementResp, "enriched", "superseded", []string{"replacement_must_not_be_superseded"})
		incidentwstest.ExpectNoSocketMessage(t, socket)

		missingReplacementID := uuid.New().String()
		missingReplacementBefore := asserttest.SnapshotCounters(t, asserttest.SQLDatabase(db), incidentID, recordID)
		missingReplacementResp := httptestx.DoJSON(t, http.DefaultClient,
			http.MethodPost,
			server.HTTP.URL+"/api/v1/records/"+recordID+"/supersede",
			map[string]any{
				"base_row_version":      3,
				"client_txn_id":         "txn-i-3-03-missing-replacement-target",
				"reason":                "missing replacement must fail",
				"replacement_record_id": missingReplacementID,
			},
			httptestx.WithCookies(reviewerAuth.Session, reviewerAuth.CSRF),
			httptestx.WithHeader(authn.CSRFHeaderName, reviewerAuth.CSRF.Value),
		)
		requireRejectedMutationStable(t, db, incidentID, recordID, missingReplacementBefore, missingReplacementResp, "enriched", "superseded", []string{"replacement_must_be_visible_active_same_incident_timeline_record"})
		incidentwstest.ExpectNoSocketMessage(t, socket)

		activeIncomingBefore := asserttest.SnapshotCounters(t, asserttest.SQLDatabase(db), incidentID, activeIncomingTargetID)
		activeIncomingResp := httptestx.DoJSON(t, http.DefaultClient,
			http.MethodPost,
			server.HTTP.URL+"/api/v1/records/"+activeIncomingTargetID+"/supersede",
			map[string]any{
				"base_row_version":      1,
				"client_txn_id":         "txn-i-3-03-active-incoming-target",
				"reason":                "target already has replacement",
				"replacement_record_id": replacementID,
			},
			httptestx.WithCookies(reviewerAuth.Session, reviewerAuth.CSRF),
			httptestx.WithHeader(authn.CSRFHeaderName, reviewerAuth.CSRF.Value),
		)
		requireRejectedMutationStable(t, db, incidentID, activeIncomingTargetID, activeIncomingBefore, activeIncomingResp, "rough", "superseded", []string{"target_must_not_have_active_replacement"})
		incidentwstest.ExpectNoSocketMessage(t, socket)

		supersede := httptestx.DoJSON(t, http.DefaultClient,
			http.MethodPost,
			server.HTTP.URL+"/api/v1/records/"+recordID+"/supersede",
			map[string]any{
				"base_row_version":      3,
				"client_txn_id":         "txn-i-3-03-supersede",
				"reason":                "  Superseded by a better row  ",
				"replacement_record_id": replacementID,
			},
			httptestx.WithCookies(reviewerAuth.Session, reviewerAuth.CSRF),
			httptestx.WithHeader(authn.CSRFHeaderName, reviewerAuth.CSRF.Value),
		)
		superseded := httptestx.RequireSuccessEnvelope(t, supersede, http.StatusOK)["data"].(map[string]any)
		if superseded["capture_state"] != "superseded" || superseded["row_version"] != float64(4) || superseded["reason"] != "Superseded by a better row" || superseded["replacement_record_id"] != replacementID {
			t.Fatalf("unexpected supersede payload: %#v", superseded)
		}
		incidentwstest.RequireRecordChanged(t, socket, recordID, 4)
		requireMutationRecorded(t, db, superseded["change_set_id"].(string), recordID, reviewerID, "timeline.records.supersede", "txn-i-3-03-supersede", 2, 4)
		asserttest.RequireSupersedeCoupledChangeSet(t, asserttest.SQLDatabase(db), superseded["change_set_id"].(string), recordID, replacementID, 4)
		if got := asserttest.CountActiveSupersedesLinks(t, asserttest.SQLDatabase(db), incidentID, replacementID, recordID); got != 1 {
			t.Fatalf("expected one active supersedes link, got %d", got)
		}
		if projection := asserttest.LookupProjectionRow(t, asserttest.SQLDatabase(db), recordID); projection.CaptureState != "superseded" || projection.ReplacementRecordID == nil || *projection.ReplacementRecordID != replacementID {
			t.Fatalf("unexpected superseded projection row: %#v", projection)
		}

		queryEnvelope := queryTimelineEnvelope(t, server, incidentID, reviewerLogin)
		contractassert.RequireDefaultQueryMeta(t, queryEnvelope, timeline.TimelineViewSchemaID)
		rawQueryRows := queryEnvelope["data"].(map[string]any)["rows"].([]any)
		queryRows := make([]map[string]any, 0, len(rawQueryRows))
		for _, rawRow := range rawQueryRows {
			queryRows = append(queryRows, rawRow.(map[string]any))
		}
		supersededRow := findRow(t, queryRows, recordID)
		if supersededRow["cells"].(map[string]any)["timeline.replacement_record_id"].(map[string]any)["value"] != replacementID {
			t.Fatalf("expected query to surface replacement_record_id, got %#v", supersededRow)
		}
		if supersededRow["group_values"].(map[string]any)["timeline.capture_state"] != "superseded" {
			t.Fatalf("expected query to surface superseded capture_state, got %#v", supersededRow)
		}
		supersedeCountsState := snapshotTimelineMutationState(t, db, incidentID, recordID)
		supersedeReplay := httptestx.DoJSON(t, http.DefaultClient,
			http.MethodPost,
			server.HTTP.URL+"/api/v1/records/"+recordID+"/supersede",
			map[string]any{
				"base_row_version":      3,
				"client_txn_id":         "txn-i-3-03-supersede",
				"reason":                "Superseded by a better row",
				"replacement_record_id": replacementID,
			},
			httptestx.WithCookies(reviewerAuth.Session, reviewerAuth.CSRF),
			httptestx.WithHeader(authn.CSRFHeaderName, reviewerAuth.CSRF.Value),
		)
		supersedeReplayData := httptestx.RequireSuccessEnvelope(t, supersedeReplay, http.StatusOK)["data"].(map[string]any)
		if supersedeReplayData["change_set_id"] != superseded["change_set_id"] {
			t.Fatalf("expected supersede replay to return original payload, got %#v", supersedeReplayData)
		}
		incidentwstest.ExpectNoSocketMessage(t, socket)
		if got := snapshotTimelineMutationState(t, db, incidentID, recordID); got != supersedeCountsState {
			t.Fatalf("replay or rejection changed mutation state: before=%+v after=%+v", supersedeCountsState, got)
		}

		supersedeDivergentReason := httptestx.DoJSON(t, http.DefaultClient,
			http.MethodPost,
			server.HTTP.URL+"/api/v1/records/"+recordID+"/supersede",
			map[string]any{
				"base_row_version":      3,
				"client_txn_id":         "txn-i-3-03-supersede",
				"reason":                "Different supersede reason",
				"replacement_record_id": replacementID,
			},
			httptestx.WithCookies(reviewerAuth.Session, reviewerAuth.CSRF),
			httptestx.WithHeader(authn.CSRFHeaderName, reviewerAuth.CSRF.Value),
		)
		httptestx.RequireErrorEnvelope(t, supersedeDivergentReason, http.StatusConflict, "client_txn_conflict")
		if got := snapshotTimelineMutationState(t, db, incidentID, recordID); got != supersedeCountsState {
			t.Fatalf("replay or rejection changed mutation state: before=%+v after=%+v", supersedeCountsState, got)
		}
		incidentwstest.ExpectNoSocketMessage(t, socket)

		supersedeDivergentReplacement := httptestx.DoJSON(t, http.DefaultClient,
			http.MethodPost,
			server.HTTP.URL+"/api/v1/records/"+recordID+"/supersede",
			map[string]any{
				"base_row_version":      3,
				"client_txn_id":         "txn-i-3-03-supersede",
				"reason":                "Superseded by a better row",
				"replacement_record_id": alternateReplacementID,
			},
			httptestx.WithCookies(reviewerAuth.Session, reviewerAuth.CSRF),
			httptestx.WithHeader(authn.CSRFHeaderName, reviewerAuth.CSRF.Value),
		)
		httptestx.RequireErrorEnvelope(t, supersedeDivergentReplacement, http.StatusConflict, "client_txn_conflict")
		incidentwstest.ExpectNoSocketMessage(t, socket)

		patchAfterSupersede := httptestx.DoJSON(t, http.DefaultClient,
			http.MethodPatch,
			server.HTTP.URL+"/api/v1/records/"+recordID,
			map[string]any{
				"view_schema_id":   timeline.TimelineViewSchemaID,
				"base_row_version": 4,
				"client_txn_id":    "txn-i-3-03-post-supersede-patch",
				"changes": []map[string]any{
					{"field_key": "timeline.activity_synopsis_text", "value": "must fail while superseded"},
				},
			},
			httptestx.WithCookies(reviewerAuth.Session, reviewerAuth.CSRF),
			httptestx.WithHeader(authn.CSRFHeaderName, reviewerAuth.CSRF.Value),
		)
		httptestx.RequireErrorEnvelope(t, patchAfterSupersede, http.StatusConflict, "illegal_transition")
		incidentwstest.ExpectNoSocketMessage(t, socket)

		illegalReview := httptestx.DoJSON(t, http.DefaultClient,
			http.MethodPost,
			server.HTTP.URL+"/api/v1/records/"+recordID+"/mark-reviewed",
			map[string]any{
				"base_row_version": 4,
				"client_txn_id":    "txn-i-3-03-illegal-review",
			},
			httptestx.WithCookies(reviewerAuth.Session, reviewerAuth.CSRF),
			httptestx.WithHeader(authn.CSRFHeaderName, reviewerAuth.CSRF.Value),
		)
		httptestx.RequireErrorEnvelope(t, illegalReview, http.StatusConflict, "illegal_transition")
		incidentwstest.ExpectNoSocketMessage(t, socket)

		updateMembershipRole(t, server, incidentID, reviewerID, 2, "viewer", adminLogin)
		patchDeniedAgainState := snapshotTimelineMutationState(t, db, incidentID, recordID)
		patchDeniedAgain := httptestx.DoJSON(t, http.DefaultClient,
			http.MethodPatch,
			server.HTTP.URL+"/api/v1/records/"+recordID,
			map[string]any{
				"view_schema_id":   timeline.TimelineViewSchemaID,
				"base_row_version": 4,
				"client_txn_id":    "txn-i-3-03-post-downgrade",
				"changes": []map[string]any{
					{"field_key": "timeline.activity_synopsis_text", "value": "must fail"},
				},
			},
			httptestx.WithCookies(reviewerAuth.Session, reviewerAuth.CSRF),
			httptestx.WithHeader(authn.CSRFHeaderName, reviewerAuth.CSRF.Value),
		)
		httptestx.RequireErrorEnvelope(t, patchDeniedAgain, http.StatusForbidden, "authorization_denied")
		if got := snapshotTimelineMutationState(t, db, incidentID, recordID); got != patchDeniedAgainState {
			t.Fatalf("denied request changed mutation state: before=%+v after=%+v", patchDeniedAgainState, got)
		}
		incidentwstest.ExpectNoSocketMessage(t, socket)

	})

	t.Run("supersede rollback clears source history projection link and collaboration", func(t *testing.T) {
		rollbackEnabled := false
		server, db, harness := startServer(t, runtime, "timeline_mutation-i-3-03-rollback")
		defer db.Close()

		adminLogin, adminID := flowtest.ProvisionBootstrapAdmin(t, http.DefaultClient, server.HTTP.URL)
		incident := createIncident(t, server, adminLogin, map[string]any{
			"client_txn_id": "txn-i-3-03-rollback-incident",
			"incident_key":  "IR-I303R",
			"title":         "Supersede rollback",
		})
		incidentID := incident["incident_id"].(string)
		replacement := createTimelineRow(t, server, incidentID, adminLogin, map[string]any{
			"client_txn_id":                   "txn-i-3-03-rollback-replacement",
			"timeline.activity_synopsis_text": "Replacement row",
		})
		replacementID := replacement["row"].(map[string]any)["record_id"].(string)
		created := createTimelineRow(t, server, incidentID, adminLogin, map[string]any{
			"client_txn_id":                   "txn-i-3-03-rollback-primary",
			"timeline.activity_synopsis_text": "Primary row",
		})
		recordID := created["row"].(map[string]any)["record_id"].(string)

		asserttest.AwaitIncidentStreamIdle(t, asserttest.SQLDatabase(db), incidentID)
		socket := incidentwstest.ConnectViewSocket(t, server, incidentID, timeline.TimelineViewSchemaID, adminLogin.Session.Value)
		defer socket.Close(1000, "test_complete")

		before := asserttest.SnapshotCounters(t, asserttest.SQLDatabase(db), incidentID, recordID)
		rollbackEnabled = true
		facade := timelineFacadeWithProjectionFailure(t, harness, func(workbookprojection.ProjectionMutation) error {
			if rollbackEnabled {
				return errors.New("forced supersede rollback")
			}
			return nil
		})
		replacementRecordID := mustUUID(t, replacementID)
		request := timeline.SupersedeRequest{
			BaseRowVersion:      1,
			ClientTxnID:         "txn-i-3-03-rollback-supersede",
			Reason:              "rollback me",
			ReplacementRecordID: &replacementRecordID,
		}
		_, err := facade.SupersedeRow(context.Background(), timeline.SupersedeCommand{
			Actor:       loadTimelineTestUser(t, harness, adminID.String()),
			RecordID:    mustUUID(t, recordID),
			Request:     request,
			RequestHash: timelineadmission.ActionRequestHash(request.BaseRowVersion, request.ClientTxnID, &request.Reason, request.ReplacementRecordID),
			RequestID:   "req-i-3-03-rollback-supersede",
			Now:         time.Now().UTC(),
		})
		if err == nil || !strings.Contains(err.Error(), "forced supersede rollback") {
			t.Fatalf("expected forced supersede rollback, got %v", err)
		}

		after := asserttest.SnapshotCounters(t, asserttest.SQLDatabase(db), incidentID, recordID)
		if before != after {
			t.Fatalf("expected rollback to keep counters stable, before=%+v after=%+v", before, after)
		}
		if got := asserttest.CountActiveSupersedesLinks(t, asserttest.SQLDatabase(db), incidentID, replacementID, recordID); got != 0 {
			t.Fatalf("rollback must clear supersedes link, got %d", got)
		}
		projection := asserttest.LookupProjectionRow(t, asserttest.SQLDatabase(db), recordID)
		if projection.CaptureState != "rough" || projection.ReplacementRecordID != nil {
			t.Fatalf("rollback must restore pre-supersede projection row, got %#v", projection)
		}
		envelope := queryTimelineEnvelope(t, server, incidentID, adminLogin)
		contractassert.RequireDefaultQueryMeta(t, envelope, timeline.TimelineViewSchemaID)
		rawEnvelopeRows := envelope["data"].(map[string]any)["rows"].([]any)
		envelopeRows := make([]map[string]any, 0, len(rawEnvelopeRows))
		for _, rawRow := range rawEnvelopeRows {
			envelopeRows = append(envelopeRows, rawRow.(map[string]any))
		}
		row := findRow(t, envelopeRows, recordID)
		if row["cells"].(map[string]any)["timeline.replacement_record_id"].(map[string]any)["value"] != nil {
			t.Fatalf("rollback must clear replacement_record_id query surfacing, got %#v", row)
		}
		if row["group_values"].(map[string]any)["timeline.capture_state"] != "rough" {
			t.Fatalf("rollback must restore rough capture state, got %#v", row)
		}
		incidentwstest.ExpectNoSocketMessage(t, socket)
	})
}

func TestTimelineTimeConversionProfile(t *testing.T) {
	runtime := appsupport.StartRuntime(t)
	server, db, _ := startServer(t, runtime, "timeline_mutation-i-3-08-time-conversion")
	defer db.Close()

	adminLogin, adminID := flowtest.ProvisionBootstrapAdmin(t, http.DefaultClient, server.HTTP.URL)
	incident := createIncident(t, server, adminLogin, map[string]any{
		"client_txn_id": "txn-i-3-08-incident",
		"incident_key":  "IR-I308",
		"title":         "Timeline time conversion",
	})
	incidentID := incident["incident_id"].(string)
	profileURL := server.HTTP.URL + "/api/v1/incidents/" + incidentID + "/timeline-time-conversion-profile"

	defaultResp := httptestx.DoJSON(t, http.DefaultClient,
		http.MethodGet,
		profileURL,
		nil,
		httptestx.WithCookies(adminLogin.Session),
	)
	defaultProfile := httptestx.RequireSuccessEnvelope(t, defaultResp, http.StatusOK)["data"].(map[string]any)
	if defaultProfile["incident_id"] != incidentID ||
		defaultProfile["enabled"] != false ||
		defaultProfile["local_offset_minutes"] != nil ||
		defaultProfile["local_label"] != nil ||
		defaultProfile["profile_version"] != float64(1) ||
		defaultProfile["updated_at"] == "" ||
		defaultProfile["updated_by_user_id"] != nil {
		t.Fatalf("unexpected default time conversion profile: %#v", defaultProfile)
	}

	editorID := seedLocalUserFlags(t, db, "timeline_mutation-i-3-08-editor@example.test", "Timeline Time Editor", "TimelineMutationTimeEditor1!", false, false, true)
	createMembership(t, server, incidentID, editorID, "timeline_mutation-i-3-08-editor@example.test", "editor", adminLogin)
	editorLogin := flowtest.LoginLocalUser(t, http.DefaultClient, server.HTTP.URL, "timeline_mutation-i-3-08-editor@example.test", "TimelineMutationTimeEditor1!", nil)
	editorPut := httptestx.DoJSON(t, http.DefaultClient,
		http.MethodPut,
		profileURL,
		map[string]any{
			"base_profile_version": 1,
			"enabled":              true,
			"local_offset_minutes": -300,
			"local_label":          "UTC-05",
		},
		httptestx.WithCookies(editorLogin.Session, editorLogin.CSRF),
		httptestx.WithHeader(authn.CSRFHeaderName, editorLogin.CSRF.Value),
	)
	httptestx.RequireErrorEnvelope(t, editorPut, http.StatusForbidden, "authorization_denied")

	disabledPut := httptestx.DoJSON(t, http.DefaultClient,
		http.MethodPut,
		profileURL,
		map[string]any{
			"base_profile_version": 1,
			"enabled":              false,
			"local_offset_minutes": nil,
			"local_label":          nil,
		},
		httptestx.WithCookies(adminLogin.Session, adminLogin.CSRF),
		httptestx.WithHeader(authn.CSRFHeaderName, adminLogin.CSRF.Value),
	)
	disabledProfile := httptestx.RequireSuccessEnvelope(t, disabledPut, http.StatusOK)["data"].(map[string]any)
	if disabledProfile["enabled"] != false ||
		disabledProfile["local_offset_minutes"] != nil ||
		disabledProfile["local_label"] != nil ||
		disabledProfile["profile_version"] != float64(2) ||
		disabledProfile["updated_by_user_id"] != adminID.String() {
		t.Fatalf("unexpected disabled time conversion profile: %#v", disabledProfile)
	}

	enabledPut := httptestx.DoJSON(t, http.DefaultClient,
		http.MethodPut,
		profileURL,
		map[string]any{
			"base_profile_version": 2,
			"enabled":              true,
			"local_offset_minutes": -300,
			"local_label":          "UTC-05",
		},
		httptestx.WithCookies(adminLogin.Session, adminLogin.CSRF),
		httptestx.WithHeader(authn.CSRFHeaderName, adminLogin.CSRF.Value),
	)
	enabledProfile := httptestx.RequireSuccessEnvelope(t, enabledPut, http.StatusOK)["data"].(map[string]any)
	if enabledProfile["enabled"] != true ||
		enabledProfile["local_offset_minutes"] != float64(-300) ||
		enabledProfile["local_label"] != "UTC-05" ||
		enabledProfile["profile_version"] != float64(3) {
		t.Fatalf("unexpected enabled time conversion profile: %#v", enabledProfile)
	}

	stalePut := httptestx.DoJSON(t, http.DefaultClient,
		http.MethodPut,
		profileURL,
		map[string]any{
			"base_profile_version": 2,
			"enabled":              true,
			"local_offset_minutes": 60,
			"local_label":          "UTC+01",
		},
		httptestx.WithCookies(adminLogin.Session, adminLogin.CSRF),
		httptestx.WithHeader(authn.CSRFHeaderName, adminLogin.CSRF.Value),
	)
	staleBody := httptestx.RequireErrorEnvelope(t, stalePut, http.StatusConflict, "row_version_conflict")
	httptestx.RequireErrorDetail(t, staleBody, "base_row_version", float64(2))
	httptestx.RequireErrorDetail(t, staleBody, "current_row_version", float64(3))

	generatedFromLocal := createTimelineRow(t, server, incidentID, adminLogin, map[string]any{
		"client_txn_id":                "txn-i-3-08-local-create",
		"timeline.activity_local_text": "2026-06-28T12:34:56-05:00",
	})
	generatedCells := generatedFromLocal["row"].(map[string]any)["cells"].(map[string]any)
	if generatedCells["timeline.activity_local_text"].(map[string]any)["value"] != "2026-06-28T12:34:56-05:00" ||
		generatedCells["timeline.activity_utc_text"].(map[string]any)["value"] != "2026-06-28T17:34:56Z" ||
		generatedCells["timeline.activity_time_pair_state"].(map[string]any)["value"] != "paired_generated" {
		t.Fatalf("expected local-only create to generate UTC with fixed offset, got %#v", generatedCells)
	}

	preservedMismatch := createTimelineRow(t, server, incidentID, adminLogin, map[string]any{
		"client_txn_id":                "txn-i-3-08-preserve-paired-create",
		"timeline.activity_local_text": "2026-06-28T12:34:56-05:00",
		"timeline.activity_utc_text":   "2026-06-28T18:34:56Z",
	})
	preservedCells := preservedMismatch["row"].(map[string]any)["cells"].(map[string]any)
	if preservedCells["timeline.activity_local_text"].(map[string]any)["value"] != "2026-06-28T12:34:56-05:00" ||
		preservedCells["timeline.activity_utc_text"].(map[string]any)["value"] != "2026-06-28T18:34:56Z" ||
		preservedCells["timeline.activity_time_pair_state"].(map[string]any)["value"] != "paired_mismatch" {
		t.Fatalf("expected paired user values to be preserved as mismatch, got %#v", preservedCells)
	}

	conversionUnavailable := createTimelineRow(t, server, incidentID, adminLogin, map[string]any{
		"client_txn_id":                "txn-i-3-08-unparseable-local-create",
		"timeline.activity_local_text": "not-a-local-time",
	})
	unavailableCells := conversionUnavailable["row"].(map[string]any)["cells"].(map[string]any)
	if unavailableCells["timeline.activity_local_text"].(map[string]any)["value"] != "not-a-local-time" ||
		unavailableCells["timeline.activity_utc_text"].(map[string]any)["value"] != nil ||
		unavailableCells["timeline.activity_time_pair_state"].(map[string]any)["value"] != "conversion_unavailable" {
		t.Fatalf("expected unparseable local time to be preserved without generated UTC, got %#v", unavailableCells)
	}

	// Representative fixtures deliberately disable conversion. These separate
	// specimens keep every enabled-profile outcome and nullable source intent visible.
	for _, specimen := range []struct {
		name, state string
		fields      map[string]any
	}{
		{"empty", "empty", map[string]any{}},
		{"explicit-null", "empty", map[string]any{"timeline.activity_utc_text": nil, "timeline.activity_local_text": nil}},
		{"explicit-empty", "empty", map[string]any{"timeline.activity_utc_text": "", "timeline.activity_local_text": ""}},
		{"preserved-pair", "paired_user_preserved", map[string]any{"timeline.activity_utc_text": "2026-06-28T17:34:56Z", "timeline.activity_local_text": "2026-06-28T12:34:56-05:00"}},
	} {
		t.Run(specimen.name, func(t *testing.T) {
			specimen.fields["client_txn_id"] = "txn-time-specimen-" + specimen.name
			created := createTimelineRow(t, server, incidentID, adminLogin, specimen.fields)["row"].(map[string]any)
			cells := created["cells"].(map[string]any)
			if cells["timeline.activity_time_pair_state"].(map[string]any)["value"] != specimen.state {
				t.Fatalf("unexpected time-pair state: %#v", cells)
			}
			for _, field := range []string{"timeline.activity_utc_text", "timeline.activity_local_text"} {
				if cells[field].(map[string]any)["value"] != specimen.fields[field] {
					t.Fatalf("source changed on create: %s", field)
				}
			}
			patched := httptestx.DoJSON(t, http.DefaultClient, http.MethodPatch, server.HTTP.URL+"/api/v1/records/"+created["record_id"].(string), map[string]any{
				"client_txn_id":  "txn-time-specimen-patch-" + specimen.name,
				"view_schema_id": timeline.TimelineViewSchemaID, "base_row_version": created["row_version"],
				"changes": []map[string]any{{"field_key": "timeline.activity_synopsis_text", "value": "Separate semantic specimen"}},
			}, httptestx.WithCookies(adminLogin.Session, adminLogin.CSRF), httptestx.WithHeader(authn.CSRFHeaderName, adminLogin.CSRF.Value))
			current := httptestx.RequireSuccessEnvelope(t, patched, http.StatusOK)["data"].(map[string]any)["row"].(map[string]any)
			for _, field := range []string{"timeline.activity_utc_text", "timeline.activity_local_text"} {
				if current["cells"].(map[string]any)[field].(map[string]any)["value"] != specimen.fields[field] {
					t.Fatalf("patch omission changed source: %s", field)
				}
			}
		})
	}
}

func startServer(t testing.TB, runtime *appsupport.Runtime, prefix string) (*httptestx.Server, *sql.DB, *appsupport.ServerHarness) {
	t.Helper()

	harness := runtime.StartServer(t, appsupport.ServerOptions{Prefix: prefix, TestRouteMode: httptestx.TestRouteModeDisabled})
	return harness.Server, harness.DB, harness
}

func timelineFacadeWithProjectionFailure(t testing.TB, harness *appsupport.ServerHarness, fail func(workbookprojection.ProjectionMutation) error) *timeline.Facade {
	t.Helper()

	return timelineFacadeWithCollaboratorMutation(t, harness, func(collaborators *timeline.Collaborators) {
		collaborators.Commit.Projection = fakeports.Projection{
			Delegate:  collaborators.Commit.Projection,
			FailApply: fail,
		}
	})
}

func timelineFacadeWithCollaboratorMutation(t testing.TB, harness *appsupport.ServerHarness, mutate func(*timeline.Collaborators)) *timeline.Facade {
	t.Helper()

	projections := mustBuildProjectionRuntime(t, harness.Pool)
	conflictTokens := conflicttest.NewCodec("timeline")
	evidenceOwner := appsupport.NewEvidenceSourceRuntimeForTimeline(
		harness.Pool,
		conflictTokens,
		harness.Revisions.Appender(),
		harness.Collaboration.RecordChanges(),
		projections,
	)
	collaborators, err := timelineassembly.NewCollaborators(timelineassembly.Dependencies{
		Postgres:            harness.Pool,
		ConflictTokens:      conflictTokens,
		ConflictFields:      harness.Revisions.ConflictFieldResolver(),
		Revisions:           harness.Revisions.Appender(),
		Collaboration:       harness.Collaboration.RecordChanges(),
		EvidenceAttachments: evidenceOwner.TimelineAttachmentContribution(),
		TimelineProjection:  projections.TimelinePorts().Writer,
		EntityProjection:    projections.EntityMutationRows(),
		AssessmentRows:      projections.AssessmentPorts().Rows,
	})
	if err != nil {
		t.Fatalf("compose Timeline failure collaborators: %v", err)
	}
	if mutate != nil {
		mutate(&collaborators)
	}
	return timeline.NewFacade(
		harness.Pool,
		collaborators,
		conflictTokens,
	)
}

func loadTimelineTestUser(t testing.TB, harness *appsupport.ServerHarness, userID string) authn.UserRecord {
	t.Helper()

	user, err := authn.NewStore(harness.Pool).GetUserByID(context.Background(), mustUUID(t, userID))
	if err != nil {
		t.Fatalf("load Timeline test user: %v", err)
	}
	return user
}

func createIncident(t testing.TB, server *httptestx.Server, admin authcookietest.AuthCookies, body map[string]any) map[string]any {
	t.Helper()

	return incidentscenariotest.CreateIncident(t, server, admin, body)
}

func createTimelineRow(t testing.TB, server *httptestx.Server, incidentID string, admin authcookietest.AuthCookies, body map[string]any) map[string]any {
	t.Helper()

	return routetest.CreateRow(t, server, admin, incidentID, body)
}

func createMembership(t testing.TB, server *httptestx.Server, incidentID string, userID string, email string, role string, admin authcookietest.AuthCookies) {
	t.Helper()

	incidentscenariotest.CreateMembershipForUser(t, server, admin, incidentID, userID, email, role)
}

func updateMembershipRole(t testing.TB, server *httptestx.Server, incidentID string, userID string, baseVersion int, role string, admin authcookietest.AuthCookies) {
	t.Helper()

	incidentscenariotest.UpdateMembershipRole(t, server, admin, incidentID, userID, baseVersion, role)
}

func seedLocalUserFlags(t testing.TB, db *sql.DB, email string, displayName string, password string, mfaRequired bool, isDeploymentAdmin bool, isActive bool) string {
	t.Helper()

	return flowtest.SeedLocalUserRecord(t, db, email, displayName, password, mfaRequired, isDeploymentAdmin, isActive).ID.String()
}

func requireMutationRecorded(t testing.TB, db *sql.DB, changeSetID string, recordID string, wantActorUserID string, wantSource string, wantClientTxnID string, wantMutationRows int, wantRevisions int) {
	t.Helper()

	workbookscenariotest.RequireChangeSetAttribution(t, db, changeSetID, wantActorUserID, wantSource, wantClientTxnID)
	database := asserttest.SQLDatabase(db)
	if got := asserttest.CountChangeSetMutations(t, database, changeSetID); got != wantMutationRows {
		t.Fatalf("unexpected mutation row count for %s: got %d want %d", changeSetID, got, wantMutationRows)
	}
	if got := asserttest.CountRecordRevisions(t, database, recordID); got != wantRevisions {
		t.Fatalf("unexpected record revision count for %s: got %d want %d", recordID, got, wantRevisions)
	}
}

func requireRejectedMutationStable(t testing.TB, db *sql.DB, incidentID string, recordID string, before asserttest.Counters, resp *http.Response, wantFrom string, wantTo string, wantGuards []string) {
	t.Helper()

	body := httptestx.RequireErrorEnvelope(t, resp, http.StatusConflict, "illegal_transition")
	httptestx.RequireErrorDetail(t, body, "from_status", wantFrom)
	httptestx.RequireErrorDetail(t, body, "to_status", wantTo)
	httptestx.RequireErrorDetailStrings(t, body, "violated_guards", wantGuards)
	after := asserttest.SnapshotCounters(t, asserttest.SQLDatabase(db), incidentID, recordID)
	if before != after {
		t.Fatalf("rejected mutation must not write history or projection rows: before=%#v after=%#v", before, after)
	}
}

func queryTimelineEnvelope(t testing.TB, server *httptestx.Server, incidentID string, login authcookietest.AuthCookies) map[string]any {
	t.Helper()

	return workbookscenariotest.QueryViewEnvelope(t, server.HTTP.URL, incidentID, timeline.TimelineViewSchemaID, login)
}

func queryTimelineRows(t testing.TB, server *httptestx.Server, incidentID string, login authcookietest.AuthCookies) []map[string]any {
	t.Helper()

	return workbookscenariotest.QueryViewRows(t, server.HTTP.URL, incidentID, timeline.TimelineViewSchemaID, login)
}

func findRow(t testing.TB, rows []map[string]any, recordID string) map[string]any {
	t.Helper()

	return workbookscenariotest.FindRow(t, rows, recordID)
}

func mustUUID(t testing.TB, raw string) uuid.UUID {
	t.Helper()

	return appsupport.MustUUID(t, raw)
}

func queryCount(t testing.TB, db *sql.DB, query string, args ...any) int {
	t.Helper()

	return dbassert.CountSQL(t, db, query, args...)
}
