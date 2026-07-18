package timeline_test

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	"github.com/JochiRaider/cartulary/internal/modules/collaboration/testsupport/incidentwstest"
	incidentscenariotest "github.com/JochiRaider/cartulary/internal/modules/incidents/testsupport/scenariotest"
	"github.com/JochiRaider/cartulary/internal/modules/projections"
	"github.com/JochiRaider/cartulary/internal/modules/timeline"
	"github.com/JochiRaider/cartulary/internal/modules/timeline/testsupport/asserttest"
	"github.com/JochiRaider/cartulary/internal/modules/timeline/testsupport/scenariotest"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	platformws "github.com/JochiRaider/cartulary/internal/platform/ws"
	"github.com/JochiRaider/cartulary/internal/testutil/contractassert"
	"github.com/JochiRaider/cartulary/internal/testutil/dbassert"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
)

func TestCreatePatchReplayAndRollback_Integration(t *testing.T) {
	runtime := scenariotest.StartRuntime(t)

	t.Run("create patch review and supersede replays stay single-write and preserve substrate", func(t *testing.T) {
		server, db := startServer(t, runtime, "phase3-i-3-01-replay")

		adminLogin, adminID := provisionBootstrapAdmin(t, server)
		incident := createIncident(t, server, adminLogin, map[string]any{
			"client_txn_id": "txn-i-3-01-incident",
			"incident_key":  "IR-I301",
			"title":         "Timeline substrate",
		})
		incidentID := incident["incident_id"].(string)
		socket := connectTimelineSocket(t, server, incidentID, adminLogin.sessionCookie.Value)
		defer socket.Close(1000, "test_complete")
		hubChanges, unsubscribe := server.Runtime.WSHub.SubscribeRecordChanges(24)
		defer unsubscribe()

		createResp := doJSON(
			t,
			http.MethodPost,
			server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/views/"+timeline.TimelineViewSchemaID+"/rows",
			map[string]any{
				"client_txn_id": "txn-i-3-01-row-create-zero",
			},
			withCookies(adminLogin.sessionCookie, adminLogin.csrfCookie),
			withHeader(authn.CSRFHeaderName, adminLogin.csrfCookie.Value),
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
		requireTimelineSocketChange(t, socket, recordID, 1)
		asserttest.AwaitRecordChange(t, hubChanges, 5*time.Second)
		requireMutationRecorded(t, db, createData["change_set_id"].(string), recordID, adminID, "timeline.rows.create", "txn-i-3-01-row-create-zero", 1, 1)
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

		countersAfterCreate := asserttest.SnapshotCounters(t, asserttest.SQLDatabase(db), incidentID, recordID)
		createReplay := doJSON(
			t,
			http.MethodPost,
			server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/views/"+timeline.TimelineViewSchemaID+"/rows",
			map[string]any{
				"client_txn_id": "txn-i-3-01-row-create-zero",
			},
			withCookies(adminLogin.sessionCookie, adminLogin.csrfCookie),
			withHeader(authn.CSRFHeaderName, adminLogin.csrfCookie.Value),
		)
		createReplayData := httptestx.RequireSuccessEnvelope(t, createReplay, http.StatusOK)["data"].(map[string]any)
		if createReplayData["change_set_id"] != createData["change_set_id"] {
			t.Fatalf("expected create replay to return original payload, got %#v", createReplayData)
		}
		requireNoTimelineCollaborationEmission(t, socket, hubChanges)
		contractassert.RequireReplayScaffold(t, contractassert.ReplayExpectation{
			FirstStatus:     http.StatusCreated,
			ReplayStatus:    http.StatusOK,
			DivergentStatus: http.StatusConflict,
			DivergentCode:   "client_txn_conflict",
			StableBefore: contractassert.ReplayCounts{
				ChangeSets:   countersAfterCreate.ChangeSets,
				MutationRows: countersAfterCreate.MutationRows,
				Revisions:    countersAfterCreate.Revisions,
			},
			StableAfter: contractassert.ReplayCounts{
				ChangeSets:   asserttest.SnapshotCounters(t, asserttest.SQLDatabase(db), incidentID, recordID).ChangeSets,
				MutationRows: asserttest.SnapshotCounters(t, asserttest.SQLDatabase(db), incidentID, recordID).MutationRows,
				Revisions:    asserttest.SnapshotCounters(t, asserttest.SQLDatabase(db), incidentID, recordID).Revisions,
			},
		})

		divergentCreate := doJSON(
			t,
			http.MethodPost,
			server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/views/"+timeline.TimelineViewSchemaID+"/rows",
			map[string]any{
				"client_txn_id":                   "txn-i-3-01-row-create-zero",
				"timeline.activity_synopsis_text": "Different capture",
			},
			withCookies(adminLogin.sessionCookie, adminLogin.csrfCookie),
			withHeader(authn.CSRFHeaderName, adminLogin.csrfCookie.Value),
		)
		httptestx.RequireErrorEnvelope(t, divergentCreate, http.StatusConflict, "client_txn_conflict")
		requireNoTimelineCollaborationEmission(t, socket, hubChanges)

		replacement := createTimelineRow(t, server, incidentID, adminLogin, map[string]any{
			"client_txn_id":                   "txn-i-3-01-replacement",
			"timeline.activity_synopsis_text": "Replacement row",
		})
		replacementRow := replacement["row"].(map[string]any)
		replacementID := replacementRow["record_id"].(string)
		if replacementRow["cells"].(map[string]any)["timeline.capture_state"].(map[string]any)["value"] != "rough" {
			t.Fatalf("expected one-value create to start rough, got %#v", replacementRow)
		}
		requireTimelineSocketChange(t, socket, replacementID, 1)
		asserttest.AwaitRecordChange(t, hubChanges, 5*time.Second)

		patchResp := doJSON(
			t,
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
			withCookies(adminLogin.sessionCookie, adminLogin.csrfCookie),
			withHeader(authn.CSRFHeaderName, adminLogin.csrfCookie.Value),
		)
		patchData := httptestx.RequireSuccessEnvelope(t, patchResp, http.StatusOK)["data"].(map[string]any)
		patchedRow := patchData["row"].(map[string]any)
		if patchedRow["row_version"] != float64(2) {
			t.Fatalf("unexpected patch row_version: %#v", patchedRow)
		}
		requireTimelineSocketChange(t, socket, recordID, 2)
		asserttest.AwaitRecordChange(t, hubChanges, 5*time.Second)
		requireMutationRecorded(t, db, patchData["change_set_id"].(string), recordID, adminID, "timeline.records.patch", "txn-i-3-01-row-patch", 1, 2)
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

		countersAfterPatch := asserttest.SnapshotCounters(t, asserttest.SQLDatabase(db), incidentID, recordID)
		patchReplay := doJSON(
			t,
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
			withCookies(adminLogin.sessionCookie, adminLogin.csrfCookie),
			withHeader(authn.CSRFHeaderName, adminLogin.csrfCookie.Value),
		)
		patchReplayData := httptestx.RequireSuccessEnvelope(t, patchReplay, http.StatusOK)["data"].(map[string]any)
		if patchReplayData["change_set_id"] != patchData["change_set_id"] {
			t.Fatalf("expected patch replay to return original payload, got %#v", patchReplayData)
		}
		requireNoTimelineCollaborationEmission(t, socket, hubChanges)
		countersAfterPatchReplay := asserttest.SnapshotCounters(t, asserttest.SQLDatabase(db), incidentID, recordID)
		contractassert.RequireReplayScaffold(t, contractassert.ReplayExpectation{
			FirstStatus:     http.StatusOK,
			ReplayStatus:    http.StatusOK,
			DivergentStatus: http.StatusConflict,
			DivergentCode:   "client_txn_conflict",
			StableBefore: contractassert.ReplayCounts{
				ChangeSets:   countersAfterPatch.ChangeSets,
				MutationRows: countersAfterPatch.MutationRows,
				Revisions:    countersAfterPatch.Revisions,
			},
			StableAfter: contractassert.ReplayCounts{
				ChangeSets:   countersAfterPatchReplay.ChangeSets,
				MutationRows: countersAfterPatchReplay.MutationRows,
				Revisions:    countersAfterPatchReplay.Revisions,
			},
		})

		divergentPatch := doJSON(
			t,
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
			withCookies(adminLogin.sessionCookie, adminLogin.csrfCookie),
			withHeader(authn.CSRFHeaderName, adminLogin.csrfCookie.Value),
		)
		httptestx.RequireErrorEnvelope(t, divergentPatch, http.StatusConflict, "client_txn_conflict")
		requireNoTimelineCollaborationEmission(t, socket, hubChanges)

		reviewResp := doJSON(
			t,
			http.MethodPost,
			server.HTTP.URL+"/api/v1/records/"+recordID+"/mark-reviewed",
			map[string]any{
				"base_row_version": 2,
				"client_txn_id":    "txn-i-3-01-row-review",
				"reason":           "  Initial review  ",
			},
			withCookies(adminLogin.sessionCookie, adminLogin.csrfCookie),
			withHeader(authn.CSRFHeaderName, adminLogin.csrfCookie.Value),
		)
		reviewData := httptestx.RequireSuccessEnvelope(t, reviewResp, http.StatusOK)["data"].(map[string]any)
		if reviewData["capture_state"] != "reviewed" || reviewData["row_version"] != float64(3) || reviewData["reason"] != "Initial review" {
			t.Fatalf("unexpected review payload: %#v", reviewData)
		}
		requireTimelineSocketChange(t, socket, recordID, 3)
		asserttest.AwaitRecordChange(t, hubChanges, 5*time.Second)
		requireMutationRecorded(t, db, reviewData["change_set_id"].(string), recordID, adminID, "timeline.records.mark_reviewed", "txn-i-3-01-row-review", 1, 3)
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

		reviewCounts := asserttest.SnapshotCounters(t, asserttest.SQLDatabase(db), incidentID, recordID)
		reviewReplay := doJSON(
			t,
			http.MethodPost,
			server.HTTP.URL+"/api/v1/records/"+recordID+"/mark-reviewed",
			map[string]any{
				"base_row_version": 2,
				"client_txn_id":    "txn-i-3-01-row-review",
				"reason":           "Initial review",
			},
			withCookies(adminLogin.sessionCookie, adminLogin.csrfCookie),
			withHeader(authn.CSRFHeaderName, adminLogin.csrfCookie.Value),
		)
		reviewReplayData := httptestx.RequireSuccessEnvelope(t, reviewReplay, http.StatusOK)["data"].(map[string]any)
		if reviewReplayData["change_set_id"] != reviewData["change_set_id"] {
			t.Fatalf("expected review replay to return original payload, got %#v", reviewReplayData)
		}
		requireNoTimelineCollaborationEmission(t, socket, hubChanges)
		reviewCountsAfterReplay := asserttest.SnapshotCounters(t, asserttest.SQLDatabase(db), incidentID, recordID)
		contractassert.RequireReplayScaffold(t, contractassert.ReplayExpectation{
			FirstStatus:     http.StatusOK,
			ReplayStatus:    http.StatusOK,
			DivergentStatus: http.StatusConflict,
			DivergentCode:   "client_txn_conflict",
			StableBefore: contractassert.ReplayCounts{
				ChangeSets:   reviewCounts.ChangeSets,
				MutationRows: reviewCounts.MutationRows,
				Revisions:    reviewCounts.Revisions,
			},
			StableAfter: contractassert.ReplayCounts{
				ChangeSets:   reviewCountsAfterReplay.ChangeSets,
				MutationRows: reviewCountsAfterReplay.MutationRows,
				Revisions:    reviewCountsAfterReplay.Revisions,
			},
		})

		reviewDivergent := doJSON(
			t,
			http.MethodPost,
			server.HTTP.URL+"/api/v1/records/"+recordID+"/mark-reviewed",
			map[string]any{
				"base_row_version": 2,
				"client_txn_id":    "txn-i-3-01-row-review",
				"reason":           "Different review reason",
			},
			withCookies(adminLogin.sessionCookie, adminLogin.csrfCookie),
			withHeader(authn.CSRFHeaderName, adminLogin.csrfCookie.Value),
		)
		httptestx.RequireErrorEnvelope(t, reviewDivergent, http.StatusConflict, "client_txn_conflict")
		requireNoTimelineCollaborationEmission(t, socket, hubChanges)

		supersedeResp := doJSON(
			t,
			http.MethodPost,
			server.HTTP.URL+"/api/v1/records/"+recordID+"/supersede",
			map[string]any{
				"base_row_version":      3,
				"client_txn_id":         "txn-i-3-01-row-supersede",
				"reason":                "  Superseded by a better row  ",
				"replacement_record_id": replacementID,
			},
			withCookies(adminLogin.sessionCookie, adminLogin.csrfCookie),
			withHeader(authn.CSRFHeaderName, adminLogin.csrfCookie.Value),
		)
		supersedeData := httptestx.RequireSuccessEnvelope(t, supersedeResp, http.StatusOK)["data"].(map[string]any)
		if supersedeData["capture_state"] != "superseded" || supersedeData["row_version"] != float64(4) || supersedeData["reason"] != "Superseded by a better row" || supersedeData["replacement_record_id"] != replacementID {
			t.Fatalf("unexpected supersede payload: %#v", supersedeData)
		}
		requireTimelineSocketChange(t, socket, recordID, 4)
		asserttest.AwaitRecordChange(t, hubChanges, 5*time.Second)
		requireMutationRecorded(t, db, supersedeData["change_set_id"].(string), recordID, adminID, "timeline.records.supersede", "txn-i-3-01-row-supersede", 2, 4)
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

		supersedeCounts := asserttest.SnapshotCounters(t, asserttest.SQLDatabase(db), incidentID, recordID)
		supersedeReplay := doJSON(
			t,
			http.MethodPost,
			server.HTTP.URL+"/api/v1/records/"+recordID+"/supersede",
			map[string]any{
				"base_row_version":      3,
				"client_txn_id":         "txn-i-3-01-row-supersede",
				"reason":                "Superseded by a better row",
				"replacement_record_id": replacementID,
			},
			withCookies(adminLogin.sessionCookie, adminLogin.csrfCookie),
			withHeader(authn.CSRFHeaderName, adminLogin.csrfCookie.Value),
		)
		supersedeReplayData := httptestx.RequireSuccessEnvelope(t, supersedeReplay, http.StatusOK)["data"].(map[string]any)
		if supersedeReplayData["change_set_id"] != supersedeData["change_set_id"] {
			t.Fatalf("expected supersede replay to return original payload, got %#v", supersedeReplayData)
		}
		requireNoTimelineCollaborationEmission(t, socket, hubChanges)
		supersedeCountsAfterReplay := asserttest.SnapshotCounters(t, asserttest.SQLDatabase(db), incidentID, recordID)
		contractassert.RequireReplayScaffold(t, contractassert.ReplayExpectation{
			FirstStatus:     http.StatusOK,
			ReplayStatus:    http.StatusOK,
			DivergentStatus: http.StatusConflict,
			DivergentCode:   "client_txn_conflict",
			StableBefore: contractassert.ReplayCounts{
				ChangeSets:   supersedeCounts.ChangeSets,
				MutationRows: supersedeCounts.MutationRows,
				Revisions:    supersedeCounts.Revisions,
			},
			StableAfter: contractassert.ReplayCounts{
				ChangeSets:   supersedeCountsAfterReplay.ChangeSets,
				MutationRows: supersedeCountsAfterReplay.MutationRows,
				Revisions:    supersedeCountsAfterReplay.Revisions,
			},
		})

		supersedeDivergentReason := doJSON(
			t,
			http.MethodPost,
			server.HTTP.URL+"/api/v1/records/"+recordID+"/supersede",
			map[string]any{
				"base_row_version":      3,
				"client_txn_id":         "txn-i-3-01-row-supersede",
				"reason":                "Different supersede reason",
				"replacement_record_id": replacementID,
			},
			withCookies(adminLogin.sessionCookie, adminLogin.csrfCookie),
			withHeader(authn.CSRFHeaderName, adminLogin.csrfCookie.Value),
		)
		httptestx.RequireErrorEnvelope(t, supersedeDivergentReason, http.StatusConflict, "client_txn_conflict")
		requireNoTimelineCollaborationEmission(t, socket, hubChanges)

		supersedeDivergentReplacement := doJSON(
			t,
			http.MethodPost,
			server.HTTP.URL+"/api/v1/records/"+recordID+"/supersede",
			map[string]any{
				"base_row_version":      3,
				"client_txn_id":         "txn-i-3-01-row-supersede",
				"reason":                "Superseded by a better row",
				"replacement_record_id": recordID,
			},
			withCookies(adminLogin.sessionCookie, adminLogin.csrfCookie),
			withHeader(authn.CSRFHeaderName, adminLogin.csrfCookie.Value),
		)
		httptestx.RequireErrorEnvelope(t, supersedeDivergentReplacement, http.StatusConflict, "client_txn_conflict")
		requireNoTimelineCollaborationEmission(t, socket, hubChanges)
	})

	t.Run("late transaction failures roll back source history projection and collaboration", func(t *testing.T) {
		server, db := startServerWithTimelineOptions(t, runtime, "phase3-i-3-01-rollback", timeline.WithBeforeCommitHookForTesting(
			func(routeKey string, recordID uuid.UUID) error {
				return errors.New("forced timeline rollback")
			},
		))
		defer db.Close()

		adminLogin, _ := provisionBootstrapAdmin(t, server)
		incident := createIncident(t, server, adminLogin, map[string]any{
			"client_txn_id": "txn-i-3-01-rollback-incident",
			"incident_key":  "IR-I301R",
			"title":         "Rollback proof",
		})
		incidentID := incident["incident_id"].(string)
		socket := connectTimelineSocket(t, server, incidentID, adminLogin.sessionCookie.Value)
		defer socket.Close(1000, "test_complete")
		hubChanges, unsubscribe := server.Runtime.WSHub.SubscribeRecordChanges(4)
		defer unsubscribe()

		createResp := doJSON(
			t,
			http.MethodPost,
			server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/views/"+timeline.TimelineViewSchemaID+"/rows",
			map[string]any{
				"client_txn_id":                   "txn-i-3-01-rollback-row",
				"timeline.activity_synopsis_text": "Rollback row",
			},
			withCookies(adminLogin.sessionCookie, adminLogin.csrfCookie),
			withHeader(authn.CSRFHeaderName, adminLogin.csrfCookie.Value),
		)
		httptestx.RequireErrorEnvelope(t, createResp, http.StatusInternalServerError, "internal_error")

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
		asserttest.RequireNoRecordChange(t, hubChanges, 300*time.Millisecond)
	})
}

func TestRouteIdempotencyIsActorScoped(t *testing.T) {
	runtime := scenariotest.StartRuntime(t)
	server, db := startServer(t, runtime, "phase3-idempotency-actor-scope")
	defer db.Close()

	adminLogin, adminID := provisionBootstrapAdmin(t, server)
	editorID := seedLocalUserFlags(t, db, "phase3-actor-scope-editor@example.test", "Actor Scope Editor", "ActorScopeEditor1!", false, false, true)
	incident := createIncident(t, server, adminLogin, map[string]any{
		"client_txn_id": "txn-phase3-actor-scope-incident",
		"incident_key":  "IR-ACTOR-SCOPE",
		"title":         "Actor-scoped idempotency",
	})
	incidentID := incident["incident_id"].(string)
	createMembership(t, server, incidentID, editorID, "phase3-actor-scope-editor@example.test", "editor", adminLogin)
	editorSession, editorCSRF := loginLocalUser(t, server, "phase3-actor-scope-editor@example.test", "ActorScopeEditor1!")
	editorLogin := loginResult{sessionCookie: editorSession, csrfCookie: editorCSRF}

	createPayload := map[string]any{
		"client_txn_id":                   "txn-phase3-shared-row-create",
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
		login loginResult
		want  map[string]any
	}{
		{login: adminLogin, want: adminCreate},
		{login: editorLogin, want: editorCreate},
	} {
		resp := doJSON(
			t,
			http.MethodPost,
			server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/views/"+timeline.TimelineViewSchemaID+"/rows",
			createPayload,
			withCookies(tc.login.sessionCookie, tc.login.csrfCookie),
			withHeader(authn.CSRFHeaderName, tc.login.csrfCookie.Value),
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
`, "timeline.rows.create", adminID, editorID, incidentID+":"+timeline.TimelineViewSchemaID, "txn-phase3-shared-row-create"); got != 2 {
		t.Fatalf("expected two actor-scoped create idempotency rows, got %d", got)
	}

	patchTarget := createTimelineRow(t, server, incidentID, adminLogin, map[string]any{
		"client_txn_id":                   "txn-phase3-actor-scope-patch-target",
		"timeline.activity_synopsis_text": "patch target",
	})
	patchRecordID := patchTarget["row"].(map[string]any)["record_id"].(string)
	adminPatchPayload := map[string]any{
		"view_schema_id":   timeline.TimelineViewSchemaID,
		"base_row_version": 1,
		"client_txn_id":    "txn-phase3-shared-row-patch",
		"changes": []map[string]any{
			{"field_key": "timeline.activity_synopsis_text", "value": "admin patch"},
		},
	}
	adminPatchResp := doJSON(
		t,
		http.MethodPatch,
		server.HTTP.URL+"/api/v1/records/"+patchRecordID,
		adminPatchPayload,
		withCookies(adminLogin.sessionCookie, adminLogin.csrfCookie),
		withHeader(authn.CSRFHeaderName, adminLogin.csrfCookie.Value),
	)
	adminPatch := httptestx.RequireSuccessEnvelope(t, adminPatchResp, http.StatusOK)["data"].(map[string]any)

	editorPatchPayload := map[string]any{
		"view_schema_id":   timeline.TimelineViewSchemaID,
		"base_row_version": 2,
		"client_txn_id":    "txn-phase3-shared-row-patch",
		"changes": []map[string]any{
			{"field_key": "timeline.raw_activity_text", "value": "editor patch"},
		},
	}
	editorPatchResp := doJSON(
		t,
		http.MethodPatch,
		server.HTTP.URL+"/api/v1/records/"+patchRecordID,
		editorPatchPayload,
		withCookies(editorLogin.sessionCookie, editorLogin.csrfCookie),
		withHeader(authn.CSRFHeaderName, editorLogin.csrfCookie.Value),
	)
	editorPatch := httptestx.RequireSuccessEnvelope(t, editorPatchResp, http.StatusOK)["data"].(map[string]any)
	if adminPatch["change_set_id"] == editorPatch["change_set_id"] {
		t.Fatalf("cross-actor patch must commit independent change_sets, got %#v", adminPatch)
	}
	for _, tc := range []struct {
		login   loginResult
		payload map[string]any
		want    map[string]any
	}{
		{login: adminLogin, payload: adminPatchPayload, want: adminPatch},
		{login: editorLogin, payload: editorPatchPayload, want: editorPatch},
	} {
		resp := doJSON(
			t,
			http.MethodPatch,
			server.HTTP.URL+"/api/v1/records/"+patchRecordID,
			tc.payload,
			withCookies(tc.login.sessionCookie, tc.login.csrfCookie),
			withHeader(authn.CSRFHeaderName, tc.login.csrfCookie.Value),
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
`, "timeline.records.patch", adminID, editorID, patchRecordID, "txn-phase3-shared-row-patch"); got != 2 {
		t.Fatalf("expected two actor-scoped patch idempotency rows, got %d", got)
	}
	if got := queryCount(t, db, `
SELECT COUNT(DISTINCT actor_user_id)
  FROM route_idempotency
 WHERE route_key IN ('timeline.rows.create', 'timeline.records.patch')
   AND client_txn_id IN ('txn-phase3-shared-row-create', 'txn-phase3-shared-row-patch')
   AND actor_user_id::text IN ($1, $2)
`, adminID, editorID); got != 2 {
		t.Fatalf("expected both actors represented in idempotency rows, got %d", got)
	}
}

func TestPatchSameFieldConflictEnvelope_Integration(t *testing.T) {
	runtime := scenariotest.StartRuntime(t)
	server, db := startServer(t, runtime, "phase3-i-3-04-same-field-conflict")
	defer db.Close()

	adminLogin, adminID := provisionBootstrapAdmin(t, server)
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

	serverPatch := doJSON(
		t,
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
		withCookies(adminLogin.sessionCookie, adminLogin.csrfCookie),
		withHeader(authn.CSRFHeaderName, adminLogin.csrfCookie.Value),
	)
	httptestx.RequireSuccessEnvelope(t, serverPatch, http.StatusOK)
	beforeConflict := asserttest.SnapshotCounters(t, asserttest.SQLDatabase(db), incidentID, recordID)

	clientPatch := doJSON(
		t,
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
		withCookies(adminLogin.sessionCookie, adminLogin.csrfCookie),
		withHeader(authn.CSRFHeaderName, adminLogin.csrfCookie.Value),
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
		conflict["server_updated_by"] != adminID ||
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
`, adminID, recordID, "txn-i-3-04-conflict-client"); got != 0 {
		t.Fatalf("same-field HTTP conflict must not persist idempotency row, got %d", got)
	}
}

func TestRouteEnvelopeMatrix_Integration(t *testing.T) {
	runtime := scenariotest.StartRuntime(t)
	server, db := startServer(t, runtime, "phase3-i-3-06-envelope-matrix")
	defer db.Close()

	adminLogin, _ := provisionBootstrapAdmin(t, server)
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
		withCookies(adminLogin.sessionCookie, adminLogin.csrfCookie),
		withHeader(authn.CSRFHeaderName, adminLogin.csrfCookie.Value),
	}
	queryOptions := []func(*http.Request){withCookies(adminLogin.sessionCookie)}
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
			resp := doRawJSON(t, tc.method, tc.url, tc.body, tc.options...)
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
	runtime := scenariotest.StartRuntime(t)
	server, db := startServer(t, runtime, "phase3-i-3-07-rough-preservation")
	defer db.Close()

	adminLogin, adminID := provisionBootstrapAdmin(t, server)
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
	seedHostRecord(t, db, mustUUID(t, incidentID), mustUUID(t, adminID), hostID, "VPN Gateway", "vpn-gateway")
	resolveResp := doJSON(
		t,
		http.MethodPost,
		server.HTTP.URL+"/api/v1/entity-mentions/"+mentionID.String()+"/resolve",
		map[string]any{
			"base_mention_row_version": 1,
			"client_txn_id":            "txn-i-3-07-resolve-host",
			"action":                   "resolve_item",
			"resolved_record_id":       hostID.String(),
		},
		withCookies(adminLogin.sessionCookie, adminLogin.csrfCookie),
		withHeader(authn.CSRFHeaderName, adminLogin.csrfCookie.Value),
	)
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
	runtime := scenariotest.StartRuntime(t)

	server, db := startServer(t, runtime, "phase3-i-3-02")
	defer db.Close()

	adminLogin, _ := provisionBootstrapAdmin(t, server)
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

	patch := doJSON(
		t,
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
		withCookies(adminLogin.sessionCookie, adminLogin.csrfCookie),
		withHeader(authn.CSRFHeaderName, adminLogin.csrfCookie.Value),
	)
	httptestx.RequireSuccessEnvelope(t, patch, http.StatusOK)

	beforeEnvelope := queryTimelineEnvelope(t, server, incidentID, adminLogin, map[string]any{})
	contractassert.RequireDefaultQueryMeta(t, beforeEnvelope, timeline.TimelineViewSchemaID)
	beforeRows := beforeEnvelope["data"].(map[string]any)["rows"].([]any)
	if len(beforeRows) != 3 {
		t.Fatalf("expected three projected rows, got %#v", beforeRows)
	}
	expectedFirstID, expectedSecondID := firstID, secondID
	if expectedSecondID < expectedFirstID {
		expectedFirstID, expectedSecondID = expectedSecondID, expectedFirstID
	}
	if got := beforeRows[0].(map[string]any)["record_id"]; got != expectedFirstID {
		t.Fatalf("expected first default-sorted row %s, got %#v", expectedFirstID, beforeRows)
	}
	if got := beforeRows[1].(map[string]any)["record_id"]; got != expectedSecondID {
		t.Fatalf("expected second default-sorted row %s, got %#v", expectedSecondID, beforeRows)
	}
	if got := beforeRows[2].(map[string]any)["record_id"]; got != zeroFieldID {
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

	sortedQuery := doJSON(
		t,
		http.MethodPost,
		server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/views/"+timeline.TimelineViewSchemaID+"/query",
		map[string]any{
			"sort": []map[string]any{{"field_key": "timeline.activity_synopsis_text", "direction": "asc"}},
		},
		withCookies(adminLogin.sessionCookie),
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
	emptyEnvelope := queryTimelineEnvelope(t, server, incidentID, adminLogin, map[string]any{})
	contractassert.RequireDefaultQueryMeta(t, emptyEnvelope, timeline.TimelineViewSchemaID)
	emptyRows := emptyEnvelope["data"].(map[string]any)["rows"].([]any)
	if len(emptyRows) != 0 {
		t.Fatalf("query route must read projection rows, got %#v", emptyRows)
	}

	projectionStore := projections.NewStore(server.Runtime.Postgres)
	if err := projectionStore.RebuildIncidentTimeline(context.Background(), mustUUID(t, incidentID)); err != nil {
		t.Fatalf("rebuild timeline projection: %v", err)
	}
	rebuiltEnvelope := queryTimelineEnvelope(t, server, incidentID, adminLogin, map[string]any{})
	contractassert.RequireDefaultQueryMeta(t, rebuiltEnvelope, timeline.TimelineViewSchemaID)
	rebuiltRows := rebuiltEnvelope["data"].(map[string]any)["rows"].([]any)
	contractassert.RequireProjectionDeterminism(t, beforeRows, rebuiltRows)

	if got := queryCount(t, db, `SELECT COUNT(*) FROM timeline_events WHERE incident_id::text = $1`, incidentID); got != 3 {
		t.Fatalf("rebuild must preserve source rows, got %d", got)
	}
}

func TestAuthorizationLifecycleAndSupersedeTransitions_Integration(t *testing.T) {
	runtime := scenariotest.StartRuntime(t)

	t.Run("authorization re-derives, reasons normalize, and supersede guards hold", func(t *testing.T) {
		server, db := startServer(t, runtime, "phase3-i-3-03")
		defer db.Close()

		adminLogin, adminID := provisionBootstrapAdmin(t, server)
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
		supersedeReplacementFixture := doJSON(
			t,
			http.MethodPost,
			server.HTTP.URL+"/api/v1/records/"+supersededReplacementID+"/supersede",
			map[string]any{
				"base_row_version":      1,
				"client_txn_id":         "txn-i-3-03-supersede-replacement-fixture",
				"reason":                "make replacement superseded",
				"replacement_record_id": supersededReplacementNextID,
			},
			withCookies(adminLogin.sessionCookie, adminLogin.csrfCookie),
			withHeader(authn.CSRFHeaderName, adminLogin.csrfCookie.Value),
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
`, incidentID, activeIncomingReplacementID, activeIncomingTargetID, adminID); err != nil {
			t.Fatalf("seed active incoming supersedes link: %v", err)
		}

		reviewerSession, reviewerCSRF := loginLocalUser(t, server, "reviewer-target@example.test", "ReviewerTargetPass1!")
		reviewerLogin := loginResult{sessionCookie: reviewerSession, csrfCookie: reviewerCSRF}
		socket := connectTimelineSocket(t, server, incidentID, reviewerSession.Value)
		defer socket.Close(1000, "test_complete")
		hubChanges, unsubscribe := server.Runtime.WSHub.SubscribeRecordChanges(16)
		defer unsubscribe()

		reviewDenied := doJSON(
			t,
			http.MethodPost,
			server.HTTP.URL+"/api/v1/records/"+recordID+"/mark-reviewed",
			map[string]any{
				"base_row_version": 1,
				"client_txn_id":    "txn-i-3-03-reviewed-denied",
				"reason":           "editor cannot review",
			},
			withCookies(reviewerSession, reviewerCSRF),
			withHeader(authn.CSRFHeaderName, reviewerCSRF.Value),
		)
		httptestx.RequireErrorEnvelope(t, reviewDenied, http.StatusForbidden, "authorization_denied")
		requireNoTimelineCollaborationEmission(t, socket, hubChanges)

		updateMembershipRole(t, server, incidentID, reviewerID, 1, "reviewer", adminLogin)
		beforeAuth := contractassert.AuthorizationOutcome{Status: http.StatusForbidden, Code: "authorization_denied"}

		markReviewed := doJSON(
			t,
			http.MethodPost,
			server.HTTP.URL+"/api/v1/records/"+recordID+"/mark-reviewed",
			map[string]any{
				"base_row_version": 1,
				"client_txn_id":    "txn-i-3-03-reviewed-1",
			},
			withCookies(reviewerSession, reviewerCSRF),
			withHeader(authn.CSRFHeaderName, reviewerCSRF.Value),
		)
		reviewedData := httptestx.RequireSuccessEnvelope(t, markReviewed, http.StatusOK)["data"].(map[string]any)
		if reviewedData["capture_state"] != "reviewed" || reviewedData["row_version"] != float64(2) || reviewedData["reason"] != nil {
			t.Fatalf("unexpected reviewed payload: %#v", reviewedData)
		}
		requireTimelineSocketChange(t, socket, recordID, 2)
		asserttest.AwaitRecordChange(t, hubChanges, 5*time.Second)
		requireMutationRecorded(t, db, reviewedData["change_set_id"].(string), recordID, reviewerID, "timeline.records.mark_reviewed", "txn-i-3-03-reviewed-1", 1, 2)
		if projection := asserttest.LookupProjectionRow(t, asserttest.SQLDatabase(db), recordID); projection.CaptureState != "reviewed" {
			t.Fatalf("expected reviewed projection row, got %#v", projection)
		}
		contractassert.RequireAuthorizationReDerived(t, beforeAuth, contractassert.AuthorizationOutcome{Status: http.StatusOK})

		reviewCounts := asserttest.SnapshotCounters(t, asserttest.SQLDatabase(db), incidentID, recordID)
		reviewReplay := doJSON(
			t,
			http.MethodPost,
			server.HTTP.URL+"/api/v1/records/"+recordID+"/mark-reviewed",
			map[string]any{
				"base_row_version": 1,
				"client_txn_id":    "txn-i-3-03-reviewed-1",
				"reason":           nil,
			},
			withCookies(reviewerSession, reviewerCSRF),
			withHeader(authn.CSRFHeaderName, reviewerCSRF.Value),
		)
		reviewReplayData := httptestx.RequireSuccessEnvelope(t, reviewReplay, http.StatusOK)["data"].(map[string]any)
		if reviewReplayData["change_set_id"] != reviewedData["change_set_id"] {
			t.Fatalf("expected review replay to return original payload, got %#v", reviewReplayData)
		}
		requireNoTimelineCollaborationEmission(t, socket, hubChanges)
		reviewCountsAfterReplay := asserttest.SnapshotCounters(t, asserttest.SQLDatabase(db), incidentID, recordID)
		contractassert.RequireReplayScaffold(t, contractassert.ReplayExpectation{
			FirstStatus:     http.StatusOK,
			ReplayStatus:    http.StatusOK,
			DivergentStatus: http.StatusConflict,
			DivergentCode:   "client_txn_conflict",
			StableBefore: contractassert.ReplayCounts{
				ChangeSets:   reviewCounts.ChangeSets,
				MutationRows: reviewCounts.MutationRows,
				Revisions:    reviewCounts.Revisions,
			},
			StableAfter: contractassert.ReplayCounts{
				ChangeSets:   reviewCountsAfterReplay.ChangeSets,
				MutationRows: reviewCountsAfterReplay.MutationRows,
				Revisions:    reviewCountsAfterReplay.Revisions,
			},
		})

		reviewDivergent := doJSON(
			t,
			http.MethodPost,
			server.HTTP.URL+"/api/v1/records/"+recordID+"/mark-reviewed",
			map[string]any{
				"base_row_version": 1,
				"client_txn_id":    "txn-i-3-03-reviewed-1",
				"reason":           "Different review reason",
			},
			withCookies(reviewerSession, reviewerCSRF),
			withHeader(authn.CSRFHeaderName, reviewerCSRF.Value),
		)
		httptestx.RequireErrorEnvelope(t, reviewDivergent, http.StatusConflict, "client_txn_conflict")
		requireNoTimelineCollaborationEmission(t, socket, hubChanges)

		materialEdit := doJSON(
			t,
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
			withCookies(reviewerSession, reviewerCSRF),
			withHeader(authn.CSRFHeaderName, reviewerCSRF.Value),
		)
		demoted := httptestx.RequireSuccessEnvelope(t, materialEdit, http.StatusOK)["data"].(map[string]any)
		demotedRow := demoted["row"].(map[string]any)
		if got := demotedRow["cells"].(map[string]any)["timeline.capture_state"].(map[string]any)["value"]; got != "enriched" {
			t.Fatalf("expected reviewed row to demote back to enriched, got %#v", demotedRow)
		}
		requireTimelineSocketChange(t, socket, recordID, 3)
		asserttest.AwaitRecordChange(t, hubChanges, 5*time.Second)
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
		selfSupersede := doJSON(
			t,
			http.MethodPost,
			server.HTTP.URL+"/api/v1/records/"+recordID+"/supersede",
			map[string]any{
				"base_row_version":      3,
				"client_txn_id":         "txn-i-3-03-self-supersede",
				"reason":                "self must fail",
				"replacement_record_id": recordID,
			},
			withCookies(reviewerSession, reviewerCSRF),
			withHeader(authn.CSRFHeaderName, reviewerCSRF.Value),
		)
		requireRejectedMutationStable(t, db, incidentID, recordID, selfBefore, selfSupersede, "enriched", "superseded", []string{"replacement_must_be_different_timeline_record"})
		requireNoTimelineCollaborationEmission(t, socket, hubChanges)

		crossBefore := asserttest.SnapshotCounters(t, asserttest.SQLDatabase(db), incidentID, recordID)
		crossIncidentSupersede := doJSON(
			t,
			http.MethodPost,
			server.HTTP.URL+"/api/v1/records/"+recordID+"/supersede",
			map[string]any{
				"base_row_version":      3,
				"client_txn_id":         "txn-i-3-03-cross-supersede",
				"reason":                "cross incident must fail",
				"replacement_record_id": otherReplacementID,
			},
			withCookies(reviewerSession, reviewerCSRF),
			withHeader(authn.CSRFHeaderName, reviewerCSRF.Value),
		)
		requireRejectedMutationStable(t, db, incidentID, recordID, crossBefore, crossIncidentSupersede, "enriched", "superseded", []string{"replacement_must_be_visible_active_same_incident_timeline_record"})
		requireNoTimelineCollaborationEmission(t, socket, hubChanges)

		supersededReplacementBefore := asserttest.SnapshotCounters(t, asserttest.SQLDatabase(db), incidentID, recordID)
		supersededReplacementResp := doJSON(
			t,
			http.MethodPost,
			server.HTTP.URL+"/api/v1/records/"+recordID+"/supersede",
			map[string]any{
				"base_row_version":      3,
				"client_txn_id":         "txn-i-3-03-superseded-replacement-target",
				"reason":                "superseded replacement must fail",
				"replacement_record_id": supersededReplacementID,
			},
			withCookies(reviewerSession, reviewerCSRF),
			withHeader(authn.CSRFHeaderName, reviewerCSRF.Value),
		)
		requireRejectedMutationStable(t, db, incidentID, recordID, supersededReplacementBefore, supersededReplacementResp, "enriched", "superseded", []string{"replacement_must_not_be_superseded"})
		requireNoTimelineCollaborationEmission(t, socket, hubChanges)

		missingReplacementID := uuid.New().String()
		missingReplacementBefore := asserttest.SnapshotCounters(t, asserttest.SQLDatabase(db), incidentID, recordID)
		missingReplacementResp := doJSON(
			t,
			http.MethodPost,
			server.HTTP.URL+"/api/v1/records/"+recordID+"/supersede",
			map[string]any{
				"base_row_version":      3,
				"client_txn_id":         "txn-i-3-03-missing-replacement-target",
				"reason":                "missing replacement must fail",
				"replacement_record_id": missingReplacementID,
			},
			withCookies(reviewerSession, reviewerCSRF),
			withHeader(authn.CSRFHeaderName, reviewerCSRF.Value),
		)
		requireRejectedMutationStable(t, db, incidentID, recordID, missingReplacementBefore, missingReplacementResp, "enriched", "superseded", []string{"replacement_must_be_visible_active_same_incident_timeline_record"})
		requireNoTimelineCollaborationEmission(t, socket, hubChanges)

		activeIncomingBefore := asserttest.SnapshotCounters(t, asserttest.SQLDatabase(db), incidentID, activeIncomingTargetID)
		activeIncomingResp := doJSON(
			t,
			http.MethodPost,
			server.HTTP.URL+"/api/v1/records/"+activeIncomingTargetID+"/supersede",
			map[string]any{
				"base_row_version":      1,
				"client_txn_id":         "txn-i-3-03-active-incoming-target",
				"reason":                "target already has replacement",
				"replacement_record_id": replacementID,
			},
			withCookies(reviewerSession, reviewerCSRF),
			withHeader(authn.CSRFHeaderName, reviewerCSRF.Value),
		)
		requireRejectedMutationStable(t, db, incidentID, activeIncomingTargetID, activeIncomingBefore, activeIncomingResp, "rough", "superseded", []string{"target_must_not_have_active_replacement"})
		requireNoTimelineCollaborationEmission(t, socket, hubChanges)

		supersede := doJSON(
			t,
			http.MethodPost,
			server.HTTP.URL+"/api/v1/records/"+recordID+"/supersede",
			map[string]any{
				"base_row_version":      3,
				"client_txn_id":         "txn-i-3-03-supersede",
				"reason":                "  Superseded by a better row  ",
				"replacement_record_id": replacementID,
			},
			withCookies(reviewerSession, reviewerCSRF),
			withHeader(authn.CSRFHeaderName, reviewerCSRF.Value),
		)
		superseded := httptestx.RequireSuccessEnvelope(t, supersede, http.StatusOK)["data"].(map[string]any)
		if superseded["capture_state"] != "superseded" || superseded["row_version"] != float64(4) || superseded["reason"] != "Superseded by a better row" || superseded["replacement_record_id"] != replacementID {
			t.Fatalf("unexpected supersede payload: %#v", superseded)
		}
		requireTimelineSocketChange(t, socket, recordID, 4)
		asserttest.AwaitRecordChange(t, hubChanges, 5*time.Second)
		requireMutationRecorded(t, db, superseded["change_set_id"].(string), recordID, reviewerID, "timeline.records.supersede", "txn-i-3-03-supersede", 2, 4)
		asserttest.RequireSupersedeCoupledChangeSet(t, asserttest.SQLDatabase(db), superseded["change_set_id"].(string), recordID, replacementID, 4)
		if got := asserttest.CountActiveSupersedesLinks(t, asserttest.SQLDatabase(db), incidentID, replacementID, recordID); got != 1 {
			t.Fatalf("expected one active supersedes link, got %d", got)
		}
		if projection := asserttest.LookupProjectionRow(t, asserttest.SQLDatabase(db), recordID); projection.CaptureState != "superseded" || projection.ReplacementRecordID == nil || *projection.ReplacementRecordID != replacementID {
			t.Fatalf("unexpected superseded projection row: %#v", projection)
		}

		queryEnvelope := queryTimelineEnvelope(t, server, incidentID, reviewerLogin, map[string]any{})
		contractassert.RequireDefaultQueryMeta(t, queryEnvelope, timeline.TimelineViewSchemaID)
		supersededRow := findRow(t, queryEnvelope["data"].(map[string]any)["rows"].([]any), recordID)
		if supersededRow["cells"].(map[string]any)["timeline.replacement_record_id"].(map[string]any)["value"] != replacementID {
			t.Fatalf("expected query to surface replacement_record_id, got %#v", supersededRow)
		}
		if supersededRow["group_values"].(map[string]any)["timeline.capture_state"] != "superseded" {
			t.Fatalf("expected query to surface superseded capture_state, got %#v", supersededRow)
		}

		supersedeCounts := asserttest.SnapshotCounters(t, asserttest.SQLDatabase(db), incidentID, recordID)
		supersedeReplay := doJSON(
			t,
			http.MethodPost,
			server.HTTP.URL+"/api/v1/records/"+recordID+"/supersede",
			map[string]any{
				"base_row_version":      3,
				"client_txn_id":         "txn-i-3-03-supersede",
				"reason":                "Superseded by a better row",
				"replacement_record_id": replacementID,
			},
			withCookies(reviewerSession, reviewerCSRF),
			withHeader(authn.CSRFHeaderName, reviewerCSRF.Value),
		)
		supersedeReplayData := httptestx.RequireSuccessEnvelope(t, supersedeReplay, http.StatusOK)["data"].(map[string]any)
		if supersedeReplayData["change_set_id"] != superseded["change_set_id"] {
			t.Fatalf("expected supersede replay to return original payload, got %#v", supersedeReplayData)
		}
		requireNoTimelineCollaborationEmission(t, socket, hubChanges)
		supersedeCountsAfterReplay := asserttest.SnapshotCounters(t, asserttest.SQLDatabase(db), incidentID, recordID)
		contractassert.RequireReplayScaffold(t, contractassert.ReplayExpectation{
			FirstStatus:     http.StatusOK,
			ReplayStatus:    http.StatusOK,
			DivergentStatus: http.StatusConflict,
			DivergentCode:   "client_txn_conflict",
			StableBefore: contractassert.ReplayCounts{
				ChangeSets:   supersedeCounts.ChangeSets,
				MutationRows: supersedeCounts.MutationRows,
				Revisions:    supersedeCounts.Revisions,
			},
			StableAfter: contractassert.ReplayCounts{
				ChangeSets:   supersedeCountsAfterReplay.ChangeSets,
				MutationRows: supersedeCountsAfterReplay.MutationRows,
				Revisions:    supersedeCountsAfterReplay.Revisions,
			},
		})

		supersedeDivergentReason := doJSON(
			t,
			http.MethodPost,
			server.HTTP.URL+"/api/v1/records/"+recordID+"/supersede",
			map[string]any{
				"base_row_version":      3,
				"client_txn_id":         "txn-i-3-03-supersede",
				"reason":                "Different supersede reason",
				"replacement_record_id": replacementID,
			},
			withCookies(reviewerSession, reviewerCSRF),
			withHeader(authn.CSRFHeaderName, reviewerCSRF.Value),
		)
		httptestx.RequireErrorEnvelope(t, supersedeDivergentReason, http.StatusConflict, "client_txn_conflict")
		requireNoTimelineCollaborationEmission(t, socket, hubChanges)

		supersedeDivergentReplacement := doJSON(
			t,
			http.MethodPost,
			server.HTTP.URL+"/api/v1/records/"+recordID+"/supersede",
			map[string]any{
				"base_row_version":      3,
				"client_txn_id":         "txn-i-3-03-supersede",
				"reason":                "Superseded by a better row",
				"replacement_record_id": alternateReplacementID,
			},
			withCookies(reviewerSession, reviewerCSRF),
			withHeader(authn.CSRFHeaderName, reviewerCSRF.Value),
		)
		httptestx.RequireErrorEnvelope(t, supersedeDivergentReplacement, http.StatusConflict, "client_txn_conflict")
		requireNoTimelineCollaborationEmission(t, socket, hubChanges)

		patchAfterSupersede := doJSON(
			t,
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
			withCookies(reviewerSession, reviewerCSRF),
			withHeader(authn.CSRFHeaderName, reviewerCSRF.Value),
		)
		httptestx.RequireErrorEnvelope(t, patchAfterSupersede, http.StatusConflict, "illegal_transition")
		requireNoTimelineCollaborationEmission(t, socket, hubChanges)

		illegalReview := doJSON(
			t,
			http.MethodPost,
			server.HTTP.URL+"/api/v1/records/"+recordID+"/mark-reviewed",
			map[string]any{
				"base_row_version": 4,
				"client_txn_id":    "txn-i-3-03-illegal-review",
			},
			withCookies(reviewerSession, reviewerCSRF),
			withHeader(authn.CSRFHeaderName, reviewerCSRF.Value),
		)
		httptestx.RequireErrorEnvelope(t, illegalReview, http.StatusConflict, "illegal_transition")
		requireNoTimelineCollaborationEmission(t, socket, hubChanges)

		updateMembershipRole(t, server, incidentID, reviewerID, 2, "viewer", adminLogin)
		patchDeniedAgain := doJSON(
			t,
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
			withCookies(reviewerSession, reviewerCSRF),
			withHeader(authn.CSRFHeaderName, reviewerCSRF.Value),
		)
		httptestx.RequireErrorEnvelope(t, patchDeniedAgain, http.StatusForbidden, "authorization_denied")
		requireNoTimelineCollaborationEmission(t, socket, hubChanges)
		contractassert.RequireAuthorizationReDerived(t, contractassert.AuthorizationOutcome{Status: http.StatusOK}, contractassert.AuthorizationOutcome{Status: http.StatusForbidden, Code: "authorization_denied"})
	})

	t.Run("supersede rollback clears source history projection link and collaboration", func(t *testing.T) {
		server, db := startServerWithTimelineOptions(t, runtime, "phase3-i-3-03-rollback", timeline.WithBeforeCommitHookForTesting(
			func(routeKey string, recordID uuid.UUID) error {
				if routeKey == "timeline.records.supersede" {
					return errors.New("forced supersede rollback")
				}
				return nil
			},
		))
		defer db.Close()

		adminLogin, _ := provisionBootstrapAdmin(t, server)
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

		socket := connectTimelineSocket(t, server, incidentID, adminLogin.sessionCookie.Value)
		defer socket.Close(1000, "test_complete")
		hubChanges, unsubscribe := server.Runtime.WSHub.SubscribeRecordChanges(8)
		defer unsubscribe()

		before := asserttest.SnapshotCounters(t, asserttest.SQLDatabase(db), incidentID, recordID)
		supersede := doJSON(
			t,
			http.MethodPost,
			server.HTTP.URL+"/api/v1/records/"+recordID+"/supersede",
			map[string]any{
				"base_row_version":      1,
				"client_txn_id":         "txn-i-3-03-rollback-supersede",
				"reason":                "rollback me",
				"replacement_record_id": replacementID,
			},
			withCookies(adminLogin.sessionCookie, adminLogin.csrfCookie),
			withHeader(authn.CSRFHeaderName, adminLogin.csrfCookie.Value),
		)
		httptestx.RequireErrorEnvelope(t, supersede, http.StatusInternalServerError, "internal_error")

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
		envelope := queryTimelineEnvelope(t, server, incidentID, adminLogin, map[string]any{})
		contractassert.RequireDefaultQueryMeta(t, envelope, timeline.TimelineViewSchemaID)
		row := findRow(t, envelope["data"].(map[string]any)["rows"].([]any), recordID)
		if row["cells"].(map[string]any)["timeline.replacement_record_id"].(map[string]any)["value"] != nil {
			t.Fatalf("rollback must clear replacement_record_id query surfacing, got %#v", row)
		}
		if row["group_values"].(map[string]any)["timeline.capture_state"] != "rough" {
			t.Fatalf("rollback must restore rough capture state, got %#v", row)
		}
		requireNoTimelineCollaborationEmission(t, socket, hubChanges)
	})
}

func TestCanonicalIncidentWebSocket_Integration(t *testing.T) {
	runtime := scenariotest.StartRuntime(t)

	t.Run("handshake membership and presence snapshot use the canonical incident route", func(t *testing.T) {
		server, db := startServer(t, runtime, "phase3-i-3-05-handshake")
		defer db.Close()

		adminLogin, _ := provisionBootstrapAdmin(t, server)
		incident := createIncident(t, server, adminLogin, map[string]any{
			"client_txn_id": "txn-i-3-05-incident",
			"incident_key":  "IR-I305-A",
			"title":         "Canonical socket handshake",
		})
		incidentID := incident["incident_id"].(string)

		first := incidentwstest.ConnectAndHello(t, server.HTTP.URL, incidentID, incidentwstest.ConnectOptions{
			SessionToken:     adminLogin.sessionCookie.Value,
			ClientInstanceID: "phase3-i-3-05-first",
			Presence:         timelinePresence(),
		})
		defer first.Close(websocket.StatusNormalClosure, "test_complete")
		if len(first.PresenceSnapshot) != 1 {
			t.Fatalf("expected first presence_snapshot to include one active connection, got %#v", first.PresenceSnapshot)
		}

		second := incidentwstest.ConnectAndHello(t, server.HTTP.URL, incidentID, incidentwstest.ConnectOptions{
			SessionToken:     adminLogin.sessionCookie.Value,
			ClientInstanceID: "phase3-i-3-05-second",
			Presence:         timelinePresence(),
		})
		defer second.Close(websocket.StatusNormalClosure, "test_complete")
		if len(second.PresenceSnapshot) != 2 {
			t.Fatalf("expected second presence_snapshot to include both active connections, got %#v", second.PresenceSnapshot)
		}
		if second.PresenceSnapshot[0].ConnectionID > second.PresenceSnapshot[1].ConnectionID {
			t.Fatalf("presence_snapshot must be sorted by connection_id: %#v", second.PresenceSnapshot)
		}

		outsiderID := seedLocalUserFlags(t, db, "phase3-i-3-05-outsider@example.test", "Phase 3 WS Outsider", "Phase3OutsiderPass1!", false, false, true)
		if outsiderID == "" {
			t.Fatal("expected outsider user")
		}
		outsiderSession, _ := loginLocalUser(t, server, "phase3-i-3-05-outsider@example.test", "Phase3OutsiderPass1!")
		incidentwstest.RequireDialErrorEnvelope(t, server.HTTP.URL, incidentID, incidentwstest.ConnectOptions{
			SessionToken: outsiderSession.Value,
		}, http.StatusNotFound, "incident_not_found")

		incidentwstest.RequireDialRejectedStatus(t, server.HTTP.URL, incidentID, incidentwstest.ConnectOptions{
			Cookies: []*http.Cookie{adminLogin.sessionCookie},
			Origin:  "https://untrusted.example.test",
		}, http.StatusForbidden)
	})

	t.Run("incident membership removal revokes only that incident socket", func(t *testing.T) {
		server, db := startServer(t, runtime, "phase3-i-3-05-revocation")
		defer db.Close()

		adminLogin, _ := provisionBootstrapAdmin(t, server)
		incidentA := createIncident(t, server, adminLogin, map[string]any{
			"client_txn_id": "txn-i-3-05-incident-a",
			"incident_key":  "IR-I305-B",
			"title":         "Canonical socket revoked incident",
		})
		incidentB := createIncident(t, server, adminLogin, map[string]any{
			"client_txn_id": "txn-i-3-05-incident-b",
			"incident_key":  "IR-I305-C",
			"title":         "Canonical socket retained incident",
		})
		incidentAID := incidentA["incident_id"].(string)
		incidentBID := incidentB["incident_id"].(string)

		userID := seedLocalUserFlags(t, db, "phase3-i-3-05-member@example.test", "Phase 3 WS Member", "Phase3MemberPass1!", false, false, true)
		createMembership(t, server, incidentAID, userID, "phase3-i-3-05-member@example.test", "editor", adminLogin)
		createMembership(t, server, incidentBID, userID, "phase3-i-3-05-member@example.test", "editor", adminLogin)
		membershipVersion := queryMembershipVersion(t, db, incidentAID, userID)

		sessionCookie, _ := loginLocalUser(t, server, "phase3-i-3-05-member@example.test", "Phase3MemberPass1!")
		socketA := incidentwstest.ConnectAndHello(t, server.HTTP.URL, incidentAID, incidentwstest.ConnectOptions{
			SessionToken:     sessionCookie.Value,
			ClientInstanceID: "phase3-i-3-05-revoked",
			Presence:         timelinePresence(),
		})
		defer socketA.Close(websocket.StatusNormalClosure, "test_complete")

		incidentscenariotest.DeleteMembershipVersion(t, server, toLogin(adminLogin), incidentAID, userID, membershipVersion)
		incidentwstest.ExpectSessionRevoked(t, socketA, "incident_access_revoked")

		socketB := incidentwstest.ConnectAndHello(t, server.HTTP.URL, incidentBID, incidentwstest.ConnectOptions{
			SessionToken:     sessionCookie.Value,
			ClientInstanceID: "phase3-i-3-05-retained",
			Presence:         timelinePresence(),
		})
		defer socketB.Close(websocket.StatusNormalClosure, "test_complete")
	})
}

func TestTimelineTimeConversionProfile(t *testing.T) {
	runtime := scenariotest.StartRuntime(t)
	server, db := startServer(t, runtime, "phase3-i-3-08-time-conversion")
	defer db.Close()

	adminLogin, adminID := provisionBootstrapAdmin(t, server)
	incident := createIncident(t, server, adminLogin, map[string]any{
		"client_txn_id": "txn-i-3-08-incident",
		"incident_key":  "IR-I308",
		"title":         "Timeline time conversion",
	})
	incidentID := incident["incident_id"].(string)
	profileURL := server.HTTP.URL + "/api/v1/incidents/" + incidentID + "/timeline-time-conversion-profile"

	defaultResp := doJSON(
		t,
		http.MethodGet,
		profileURL,
		nil,
		withCookies(adminLogin.sessionCookie),
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

	editorID := seedLocalUserFlags(t, db, "phase3-i-3-08-editor@example.test", "Phase 3 Time Editor", "Phase3TimeEditor1!", false, false, true)
	createMembership(t, server, incidentID, editorID, "phase3-i-3-08-editor@example.test", "editor", adminLogin)
	editorSession, editorCSRF := loginLocalUser(t, server, "phase3-i-3-08-editor@example.test", "Phase3TimeEditor1!")
	editorPut := doJSON(
		t,
		http.MethodPut,
		profileURL,
		map[string]any{
			"base_profile_version": 1,
			"enabled":              true,
			"local_offset_minutes": -300,
			"local_label":          "UTC-05",
		},
		withCookies(editorSession, editorCSRF),
		withHeader(authn.CSRFHeaderName, editorCSRF.Value),
	)
	httptestx.RequireErrorEnvelope(t, editorPut, http.StatusForbidden, "authorization_denied")

	disabledPut := doJSON(
		t,
		http.MethodPut,
		profileURL,
		map[string]any{
			"base_profile_version": 1,
			"enabled":              false,
			"local_offset_minutes": nil,
			"local_label":          nil,
		},
		withCookies(adminLogin.sessionCookie, adminLogin.csrfCookie),
		withHeader(authn.CSRFHeaderName, adminLogin.csrfCookie.Value),
	)
	disabledProfile := httptestx.RequireSuccessEnvelope(t, disabledPut, http.StatusOK)["data"].(map[string]any)
	if disabledProfile["enabled"] != false ||
		disabledProfile["local_offset_minutes"] != nil ||
		disabledProfile["local_label"] != nil ||
		disabledProfile["profile_version"] != float64(2) ||
		disabledProfile["updated_by_user_id"] != adminID {
		t.Fatalf("unexpected disabled time conversion profile: %#v", disabledProfile)
	}

	enabledPut := doJSON(
		t,
		http.MethodPut,
		profileURL,
		map[string]any{
			"base_profile_version": 2,
			"enabled":              true,
			"local_offset_minutes": -300,
			"local_label":          "UTC-05",
		},
		withCookies(adminLogin.sessionCookie, adminLogin.csrfCookie),
		withHeader(authn.CSRFHeaderName, adminLogin.csrfCookie.Value),
	)
	enabledProfile := httptestx.RequireSuccessEnvelope(t, enabledPut, http.StatusOK)["data"].(map[string]any)
	if enabledProfile["enabled"] != true ||
		enabledProfile["local_offset_minutes"] != float64(-300) ||
		enabledProfile["local_label"] != "UTC-05" ||
		enabledProfile["profile_version"] != float64(3) {
		t.Fatalf("unexpected enabled time conversion profile: %#v", enabledProfile)
	}

	stalePut := doJSON(
		t,
		http.MethodPut,
		profileURL,
		map[string]any{
			"base_profile_version": 2,
			"enabled":              true,
			"local_offset_minutes": 60,
			"local_label":          "UTC+01",
		},
		withCookies(adminLogin.sessionCookie, adminLogin.csrfCookie),
		withHeader(authn.CSRFHeaderName, adminLogin.csrfCookie.Value),
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
}

func timelinePresence() platformws.PresenceInput {
	return platformws.PresenceInput{
		SheetRef: map[string]string{
			"kind": "view_schema",
			"id":   timeline.TimelineViewSchemaID,
		},
		Mode: "viewing",
	}
}

type loginResult struct {
	sessionCookie *http.Cookie
	csrfCookie    *http.Cookie
}

func toLogin(login loginResult) flowtest.LoginResult {
	return flowtest.LoginResult{
		SessionCookie: login.sessionCookie,
		CSRFCookie:    login.csrfCookie,
	}
}

type recordChangeSocketPayload = scenariotest.RecordChangeSocketPayload

type timelineSocketClient = scenariotest.TimelineSocketClient

func startServer(t testing.TB, runtime *scenariotest.RuntimeHarness, prefix string) (*httptestx.Server, *sql.DB) {
	t.Helper()

	harness := runtime.StartServer(t, prefix)
	return harness.Server, harness.DB
}

func startServerWithTimelineOptions(t testing.TB, runtime *scenariotest.RuntimeHarness, prefix string, options ...timeline.TestFacadeOption) (*httptestx.Server, *sql.DB) {
	t.Helper()

	harness := runtime.StartServerWithDependencies(t, prefix, timeline.DependencySetForTesting(options...))
	return harness.Server, harness.DB
}

func provisionBootstrapAdmin(t testing.TB, server *httptestx.Server) (loginResult, string) {
	t.Helper()

	login, userID := flowtest.ProvisionBootstrapAdminUUID(t, server.HTTP.URL)
	return loginResult{
		sessionCookie: login.SessionCookie,
		csrfCookie:    login.CSRFCookie,
	}, userID.String()
}

func createIncident(t testing.TB, server *httptestx.Server, admin loginResult, body map[string]any) map[string]any {
	t.Helper()

	return incidentscenariotest.CreateIncident(t, server, toLogin(admin), body)
}

func createTimelineRow(t testing.TB, server *httptestx.Server, incidentID string, admin loginResult, body map[string]any) map[string]any {
	t.Helper()

	return scenariotest.CreateTimelineRow(t, server, incidentID, toLogin(admin), body)
}

func createMembership(t testing.TB, server *httptestx.Server, incidentID string, userID string, email string, role string, admin loginResult) {
	t.Helper()

	incidentscenariotest.CreateMembershipForUser(t, server, toLogin(admin), incidentID, userID, email, role)
}

func updateMembershipRole(t testing.TB, server *httptestx.Server, incidentID string, userID string, baseVersion int, role string, admin loginResult) {
	t.Helper()

	incidentscenariotest.UpdateMembershipRole(t, server, toLogin(admin), incidentID, userID, baseVersion, role)
}

func seedLocalUserFlags(t testing.TB, db *sql.DB, email string, displayName string, password string, mfaRequired bool, isDeploymentAdmin bool, isActive bool) string {
	t.Helper()

	return flowtest.SeedLocalUserRecord(t, db, email, displayName, password, mfaRequired, isDeploymentAdmin, isActive).ID.String()
}

func loginLocalUser(t testing.TB, server *httptestx.Server, username string, password string) (*http.Cookie, *http.Cookie) {
	t.Helper()

	return flowtest.LoginLocalUser(t, server.HTTP.URL, username, password, nil)
}

func connectTimelineSocket(t testing.TB, server *httptestx.Server, incidentID string, sessionToken string) *timelineSocketClient {
	t.Helper()

	return scenariotest.ConnectTimelineSocket(t, server, incidentID, sessionToken)
}

func requireTimelineSocketChange(t testing.TB, client *timelineSocketClient, wantRecordID string, wantRowVersion int64) recordChangeSocketPayload {
	t.Helper()

	return scenariotest.RequireTimelineSocketChange(t, client, wantRecordID, wantRowVersion)
}

func expectNoTimelineSocketMessage(t testing.TB, client *timelineSocketClient) {
	t.Helper()

	scenariotest.ExpectNoTimelineSocketMessage(t, client)
}

func requireMutationRecorded(t testing.TB, db *sql.DB, changeSetID string, recordID string, wantActorUserID string, wantSource string, wantClientTxnID string, wantMutationRows int, wantRevisions int) {
	t.Helper()

	scenariotest.RequireMutationRecorded(t, db, changeSetID, recordID, wantActorUserID, wantSource, wantClientTxnID, wantMutationRows, wantRevisions)
}

func requireNoTimelineCollaborationEmission(t testing.TB, client *timelineSocketClient, changes <-chan platformws.RecordChange) {
	t.Helper()

	scenariotest.RequireNoTimelineCollaborationEmission(t, client, changes)
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

func queryTimelineEnvelope(t testing.TB, server *httptestx.Server, incidentID string, login loginResult, body map[string]any) map[string]any {
	t.Helper()

	return scenariotest.QueryTimelineEnvelope(t, server, incidentID, toLogin(login), body)
}

func queryTimelineRows(t testing.TB, server *httptestx.Server, incidentID string, login loginResult) []any {
	t.Helper()

	return scenariotest.QueryTimelineRows(t, server, incidentID, toLogin(login))
}

func findRow(t testing.TB, rows []any, recordID string) map[string]any {
	t.Helper()

	return scenariotest.FindRow(t, rows, recordID)
}

func mustUUID(t testing.TB, raw string) uuid.UUID {
	t.Helper()

	return scenariotest.MustUUID(t, raw)
}

func doJSON(t testing.TB, method string, url string, body any, options ...func(*http.Request)) *http.Response {
	t.Helper()

	return httptestx.DoJSON(t, method, url, body, options...)
}

func doRawJSON(t testing.TB, method string, url string, body string, options ...func(*http.Request)) *http.Response {
	t.Helper()

	return httptestx.DoRawJSON(t, method, url, body, options...)
}

func withCookies(cookies ...*http.Cookie) func(*http.Request) {
	return httptestx.WithCookies(cookies...)
}

func withHeader(key string, value string) func(*http.Request) {
	return httptestx.WithHeader(key, value)
}

func queryCount(t testing.TB, db *sql.DB, query string, args ...any) int {
	t.Helper()

	return dbassert.CountSQL(t, db, query, args...)
}

func queryMembershipVersion(t testing.TB, db *sql.DB, incidentID string, userID string) int64 {
	t.Helper()

	var version int64
	if err := db.QueryRowContext(context.Background(), `
SELECT membership_version
FROM incident_memberships
WHERE incident_id::text = $1 AND user_id::text = $2
`, incidentID, userID).Scan(&version); err != nil {
		t.Fatalf("query incident membership version: %v", err)
	}
	return version
}
