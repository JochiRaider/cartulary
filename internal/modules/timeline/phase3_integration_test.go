package timeline_test

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/JochiRaider/cartulary/internal/modules/projections"
	"github.com/JochiRaider/cartulary/internal/modules/timeline"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	platformws "github.com/JochiRaider/cartulary/internal/platform/ws"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/JochiRaider/cartulary/internal/testutil/phase3test"
	"github.com/JochiRaider/cartulary/internal/testutil/timelinetest"
)

func TestPhase3_I_3_01_CreatePatchReplayAndRollback(t *testing.T) {
	runtime := phase3test.StartRuntime(t)

	t.Run("create patch review and supersede replays stay single-write and preserve substrate", func(t *testing.T) {
		server, db := startPhase3Server(t, runtime, "phase3-i-3-01-replay")

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

		createResp := doPhase3JSON(
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
		if createCells["timeline.summary"].(map[string]any)["value"] != nil {
			t.Fatalf("expected zero-field create to keep summary null, got %#v", createCells["timeline.summary"])
		}
		if createCells["timeline.capture_state"].(map[string]any)["value"] != "rough" {
			t.Fatalf("expected zero-field create to start rough, got %#v", createCells["timeline.capture_state"])
		}
		requireTimelineSocketChange(t, socket, recordID, 1)
		timelinetest.AwaitRecordChange(t, hubChanges, 5*time.Second)
		requireMutationRecorded(t, db, createData["change_set_id"].(string), recordID, adminID, "timeline.rows.create", "txn-i-3-01-row-create-zero", 1, 1)
		projection := timelinetest.LookupProjectionRow(t, db, recordID)
		if projection.CaptureState != "rough" || projection.ReplacementRecordID != nil {
			t.Fatalf("unexpected zero-field projection row: %#v", projection)
		}

		countersAfterCreate := timelinetest.SnapshotCounters(t, db, incidentID, recordID)
		createReplay := doPhase3JSON(
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
		httptestx.RequireReplayScaffold(t, httptestx.ReplayExpectation{
			FirstStatus:     http.StatusCreated,
			ReplayStatus:    http.StatusOK,
			DivergentStatus: http.StatusConflict,
			DivergentCode:   "client_txn_conflict",
			StableBefore: httptestx.ReplayCounts{
				ChangeSets:   countersAfterCreate.ChangeSets,
				MutationRows: countersAfterCreate.MutationRows,
				Revisions:    countersAfterCreate.Revisions,
			},
			StableAfter: httptestx.ReplayCounts{
				ChangeSets:   timelinetest.SnapshotCounters(t, db, incidentID, recordID).ChangeSets,
				MutationRows: timelinetest.SnapshotCounters(t, db, incidentID, recordID).MutationRows,
				Revisions:    timelinetest.SnapshotCounters(t, db, incidentID, recordID).Revisions,
			},
		})

		divergentCreate := doPhase3JSON(
			t,
			http.MethodPost,
			server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/views/"+timeline.TimelineViewSchemaID+"/rows",
			map[string]any{
				"client_txn_id":    "txn-i-3-01-row-create-zero",
				"timeline.summary": "Different capture",
			},
			withCookies(adminLogin.sessionCookie, adminLogin.csrfCookie),
			withHeader(authn.CSRFHeaderName, adminLogin.csrfCookie.Value),
		)
		httptestx.RequireErrorEnvelope(t, divergentCreate, http.StatusConflict, "client_txn_conflict")
		requireNoTimelineCollaborationEmission(t, socket, hubChanges)

		replacement := createTimelineRow(t, server, incidentID, adminLogin, map[string]any{
			"client_txn_id":    "txn-i-3-01-replacement",
			"timeline.summary": "Replacement row",
		})
		replacementRow := replacement["row"].(map[string]any)
		replacementID := replacementRow["record_id"].(string)
		if replacementRow["cells"].(map[string]any)["timeline.capture_state"].(map[string]any)["value"] != "rough" {
			t.Fatalf("expected one-value create to start rough, got %#v", replacementRow)
		}
		requireTimelineSocketChange(t, socket, replacementID, 1)
		timelinetest.AwaitRecordChange(t, hubChanges, 5*time.Second)

		patchResp := doPhase3JSON(
			t,
			http.MethodPatch,
			server.HTTP.URL+"/api/v1/records/"+recordID,
			map[string]any{
				"view_schema_id":   timeline.TimelineViewSchemaID,
				"base_row_version": 1,
				"client_txn_id":    "txn-i-3-01-row-patch",
				"changes": []map[string]any{
					{"field_key": "timeline.summary", "value": "Enriched capture"},
					{"field_key": "timeline.details", "value": "Details from patch"},
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
		timelinetest.AwaitRecordChange(t, hubChanges, 5*time.Second)
		requireMutationRecorded(t, db, patchData["change_set_id"].(string), recordID, adminID, "timeline.records.patch", "txn-i-3-01-row-patch", 1, 2)
		projection = timelinetest.LookupProjectionRow(t, db, recordID)
		if projection.CaptureState != "enriched" {
			t.Fatalf("expected projection to reflect enriched patch state, got %#v", projection)
		}

		countersAfterPatch := timelinetest.SnapshotCounters(t, db, incidentID, recordID)
		patchReplay := doPhase3JSON(
			t,
			http.MethodPatch,
			server.HTTP.URL+"/api/v1/records/"+recordID,
			map[string]any{
				"view_schema_id":   timeline.TimelineViewSchemaID,
				"base_row_version": 1,
				"client_txn_id":    "txn-i-3-01-row-patch",
				"changes": []map[string]any{
					{"field_key": "timeline.details", "value": "Details from patch"},
					{"field_key": "timeline.summary", "value": "Enriched capture"},
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
		countersAfterPatchReplay := timelinetest.SnapshotCounters(t, db, incidentID, recordID)
		httptestx.RequireReplayScaffold(t, httptestx.ReplayExpectation{
			FirstStatus:     http.StatusOK,
			ReplayStatus:    http.StatusOK,
			DivergentStatus: http.StatusConflict,
			DivergentCode:   "client_txn_conflict",
			StableBefore: httptestx.ReplayCounts{
				ChangeSets:   countersAfterPatch.ChangeSets,
				MutationRows: countersAfterPatch.MutationRows,
				Revisions:    countersAfterPatch.Revisions,
			},
			StableAfter: httptestx.ReplayCounts{
				ChangeSets:   countersAfterPatchReplay.ChangeSets,
				MutationRows: countersAfterPatchReplay.MutationRows,
				Revisions:    countersAfterPatchReplay.Revisions,
			},
		})

		divergentPatch := doPhase3JSON(
			t,
			http.MethodPatch,
			server.HTTP.URL+"/api/v1/records/"+recordID,
			map[string]any{
				"view_schema_id":   timeline.TimelineViewSchemaID,
				"base_row_version": 1,
				"client_txn_id":    "txn-i-3-01-row-patch",
				"changes": []map[string]any{
					{"field_key": "timeline.summary", "value": "Divergent replay"},
				},
			},
			withCookies(adminLogin.sessionCookie, adminLogin.csrfCookie),
			withHeader(authn.CSRFHeaderName, adminLogin.csrfCookie.Value),
		)
		httptestx.RequireErrorEnvelope(t, divergentPatch, http.StatusConflict, "client_txn_conflict")
		requireNoTimelineCollaborationEmission(t, socket, hubChanges)

		reviewResp := doPhase3JSON(
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
		timelinetest.AwaitRecordChange(t, hubChanges, 5*time.Second)
		requireMutationRecorded(t, db, reviewData["change_set_id"].(string), recordID, adminID, "timeline.records.mark_reviewed", "txn-i-3-01-row-review", 1, 3)
		projection = timelinetest.LookupProjectionRow(t, db, recordID)
		if projection.CaptureState != "reviewed" {
			t.Fatalf("expected reviewed projection state, got %#v", projection)
		}

		reviewCounts := timelinetest.SnapshotCounters(t, db, incidentID, recordID)
		reviewReplay := doPhase3JSON(
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
		reviewCountsAfterReplay := timelinetest.SnapshotCounters(t, db, incidentID, recordID)
		httptestx.RequireReplayScaffold(t, httptestx.ReplayExpectation{
			FirstStatus:     http.StatusOK,
			ReplayStatus:    http.StatusOK,
			DivergentStatus: http.StatusConflict,
			DivergentCode:   "client_txn_conflict",
			StableBefore: httptestx.ReplayCounts{
				ChangeSets:   reviewCounts.ChangeSets,
				MutationRows: reviewCounts.MutationRows,
				Revisions:    reviewCounts.Revisions,
			},
			StableAfter: httptestx.ReplayCounts{
				ChangeSets:   reviewCountsAfterReplay.ChangeSets,
				MutationRows: reviewCountsAfterReplay.MutationRows,
				Revisions:    reviewCountsAfterReplay.Revisions,
			},
		})

		reviewDivergent := doPhase3JSON(
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

		supersedeResp := doPhase3JSON(
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
		timelinetest.AwaitRecordChange(t, hubChanges, 5*time.Second)
		requireMutationRecorded(t, db, supersedeData["change_set_id"].(string), recordID, adminID, "timeline.records.supersede", "txn-i-3-01-row-supersede", 2, 4)
		if got := timelinetest.CountActiveSupersedesLinks(t, db, incidentID, replacementID, recordID); got != 1 {
			t.Fatalf("expected one active supersedes link, got %d", got)
		}
		projection = timelinetest.LookupProjectionRow(t, db, recordID)
		if projection.CaptureState != "superseded" || projection.ReplacementRecordID == nil || *projection.ReplacementRecordID != replacementID {
			t.Fatalf("unexpected superseded projection row: %#v", projection)
		}

		supersedeCounts := timelinetest.SnapshotCounters(t, db, incidentID, recordID)
		supersedeReplay := doPhase3JSON(
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
		supersedeCountsAfterReplay := timelinetest.SnapshotCounters(t, db, incidentID, recordID)
		httptestx.RequireReplayScaffold(t, httptestx.ReplayExpectation{
			FirstStatus:     http.StatusOK,
			ReplayStatus:    http.StatusOK,
			DivergentStatus: http.StatusConflict,
			DivergentCode:   "client_txn_conflict",
			StableBefore: httptestx.ReplayCounts{
				ChangeSets:   supersedeCounts.ChangeSets,
				MutationRows: supersedeCounts.MutationRows,
				Revisions:    supersedeCounts.Revisions,
			},
			StableAfter: httptestx.ReplayCounts{
				ChangeSets:   supersedeCountsAfterReplay.ChangeSets,
				MutationRows: supersedeCountsAfterReplay.MutationRows,
				Revisions:    supersedeCountsAfterReplay.Revisions,
			},
		})

		supersedeDivergentReason := doPhase3JSON(
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

		supersedeDivergentReplacement := doPhase3JSON(
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
		restoreHooks := timeline.SetStoreHooksForTesting(timeline.StoreHooks{
			BeforeCommit: func(routeKey string, recordID uuid.UUID) error {
				return errors.New("forced timeline rollback")
			},
		})
		defer restoreHooks()

		server, db := startPhase3Server(t, runtime, "phase3-i-3-01-rollback")
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

		createResp := doPhase3JSON(
			t,
			http.MethodPost,
			server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/views/"+timeline.TimelineViewSchemaID+"/rows",
			map[string]any{
				"client_txn_id":    "txn-i-3-01-rollback-row",
				"timeline.summary": "Rollback row",
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
		timelinetest.RequireNoRecordChange(t, hubChanges, 300*time.Millisecond)
	})
}

func TestPhase3_PatchSameFieldConflictEnvelope_I_3_04(t *testing.T) {
	runtime := phase3test.StartRuntime(t)
	server, db := startPhase3Server(t, runtime, "phase3-i-3-04-same-field-conflict")
	defer db.Close()

	adminLogin, adminID := provisionBootstrapAdmin(t, server)
	incident := createIncident(t, server, adminLogin, map[string]any{
		"client_txn_id": "txn-i-3-04-conflict-incident",
		"incident_key":  "IR-I304",
		"title":         "Same field conflict",
	})
	incidentID := incident["incident_id"].(string)
	created := createTimelineRow(t, server, incidentID, adminLogin, map[string]any{
		"client_txn_id":    "txn-i-3-04-conflict-row",
		"timeline.summary": "Base summary",
	})
	recordID := created["row"].(map[string]any)["record_id"].(string)

	serverPatch := doPhase3JSON(
		t,
		http.MethodPatch,
		server.HTTP.URL+"/api/v1/records/"+recordID,
		map[string]any{
			"view_schema_id":   timeline.TimelineViewSchemaID,
			"base_row_version": 1,
			"client_txn_id":    "txn-i-3-04-conflict-server",
			"changes": []map[string]any{
				{"field_key": "timeline.summary", "value": "Server summary"},
			},
		},
		withCookies(adminLogin.sessionCookie, adminLogin.csrfCookie),
		withHeader(authn.CSRFHeaderName, adminLogin.csrfCookie.Value),
	)
	httptestx.RequireSuccessEnvelope(t, serverPatch, http.StatusOK)
	beforeConflict := timelinetest.SnapshotCounters(t, db, incidentID, recordID)

	clientPatch := doPhase3JSON(
		t,
		http.MethodPatch,
		server.HTTP.URL+"/api/v1/records/"+recordID,
		map[string]any{
			"view_schema_id":   timeline.TimelineViewSchemaID,
			"base_row_version": 1,
			"client_txn_id":    "txn-i-3-04-conflict-client",
			"changes": []map[string]any{
				{"field_key": "timeline.summary", "value": "Client summary"},
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
		conflict["field_key"] != "timeline.summary" ||
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
	afterConflict := timelinetest.SnapshotCounters(t, db, incidentID, recordID)
	if beforeConflict != afterConflict {
		t.Fatalf("same-field HTTP conflict must not create writes: before=%#v after=%#v", beforeConflict, afterConflict)
	}
	if got := queryCount(t, db, `SELECT COUNT(*) FROM route_idempotency WHERE route_key = 'timeline.records.patch' AND client_txn_id = $1`, "txn-i-3-04-conflict-client"); got != 0 {
		t.Fatalf("same-field HTTP conflict must not persist idempotency row, got %d", got)
	}
}

func TestPhase3_I_3_02_ProjectionQueryUsesDeterministicRebuild(t *testing.T) {
	runtime := phase3test.StartRuntime(t)

	server, db := startPhase3Server(t, runtime, "phase3-i-3-02")
	defer db.Close()

	adminLogin, _ := provisionBootstrapAdmin(t, server)
	incident := createIncident(t, server, adminLogin, map[string]any{
		"client_txn_id": "txn-i-3-02-incident",
		"incident_key":  "IR-I302",
		"title":         "Projection reads",
	})
	incidentID := incident["incident_id"].(string)

	first := createTimelineRow(t, server, incidentID, adminLogin, map[string]any{
		"client_txn_id":        "txn-i-3-02-row-a",
		"timeline.summary":     "Tie A",
		"timeline.occurred_at": "2026-04-10T10:00:00Z",
	})
	second := createTimelineRow(t, server, incidentID, adminLogin, map[string]any{
		"client_txn_id":        "txn-i-3-02-row-b",
		"timeline.summary":     "Tie B",
		"timeline.occurred_at": "2026-04-10T10:00:00Z",
	})
	zeroField := createTimelineRow(t, server, incidentID, adminLogin, map[string]any{
		"client_txn_id": "txn-i-3-02-row-c",
	})
	firstID := first["row"].(map[string]any)["record_id"].(string)
	secondID := second["row"].(map[string]any)["record_id"].(string)
	zeroFieldID := zeroField["row"].(map[string]any)["record_id"].(string)

	patch := doPhase3JSON(
		t,
		http.MethodPatch,
		server.HTTP.URL+"/api/v1/records/"+secondID,
		map[string]any{
			"view_schema_id":   timeline.TimelineViewSchemaID,
			"base_row_version": 1,
			"client_txn_id":    "txn-i-3-02-row-b-patch",
			"changes": []map[string]any{
				{"field_key": "timeline.details", "value": "Projected details"},
			},
		},
		withCookies(adminLogin.sessionCookie, adminLogin.csrfCookie),
		withHeader(authn.CSRFHeaderName, adminLogin.csrfCookie.Value),
	)
	httptestx.RequireSuccessEnvelope(t, patch, http.StatusOK)

	beforeEnvelope := queryTimelineEnvelope(t, server, incidentID, adminLogin, map[string]any{})
	httptestx.RequireDefaultQueryMeta(t, beforeEnvelope, timeline.TimelineViewSchemaID)
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
	if summary := secondRow["cells"].(map[string]any)["timeline.summary"].(map[string]any)["value"]; summary != "Tie B" {
		t.Fatalf("expected projected summary, got %#v", secondRow)
	}
	if details := secondRow["cells"].(map[string]any)["timeline.details"].(map[string]any)["value"]; details != "Projected details" {
		t.Fatalf("expected projection-backed details, got %#v", secondRow)
	}
	if captureState := secondRow["group_values"].(map[string]any)["timeline.capture_state"]; captureState != "enriched" {
		t.Fatalf("expected projected group_values capture state, got %#v", secondRow["group_values"])
	}
	if occurredDay := secondRow["group_values"].(map[string]any)["timeline.occurred_day"]; occurredDay != "2026-04-10" {
		t.Fatalf("expected occurred_day group value, got %#v", secondRow["group_values"])
	}
	zeroFieldRow := findRow(t, beforeRows, zeroFieldID)
	if zeroFieldRow["cells"].(map[string]any)["timeline.summary"].(map[string]any)["value"] != nil {
		t.Fatalf("expected zero-field query row summary to remain null, got %#v", zeroFieldRow)
	}
	if zeroFieldRow["cells"].(map[string]any)["timeline.replacement_record_id"].(map[string]any)["value"] != nil {
		t.Fatalf("expected zero-field query row replacement cell to remain null, got %#v", zeroFieldRow)
	}
	if zeroFieldRow["group_values"].(map[string]any)["timeline.capture_state"] != "rough" {
		t.Fatalf("expected zero-field query row to remain rough, got %#v", zeroFieldRow)
	}

	sortedQuery := doPhase3JSON(
		t,
		http.MethodPost,
		server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/views/"+timeline.TimelineViewSchemaID+"/query",
		map[string]any{
			"sort": []map[string]any{{"field_key": "timeline.summary", "direction": "asc"}},
		},
		withCookies(adminLogin.sessionCookie),
	)
	sortedEnvelope := httptestx.RequireSuccessEnvelope(t, sortedQuery, http.StatusOK)
	sortedMeta := sortedEnvelope["meta"].(map[string]any)["query"].(map[string]any)
	sortedSort := sortedMeta["sort"].([]any)
	if len(sortedSort) != 3 {
		t.Fatalf("expected effective sort to include user sort plus default tail, got %#v", sortedSort)
	}
	if sortedSort[0].(map[string]any)["field_key"] != "timeline.summary" || sortedSort[0].(map[string]any)["direction"] != "asc" {
		t.Fatalf("expected primary user sort on timeline.summary asc, got %#v", sortedSort)
	}
	if sortedSort[1].(map[string]any)["field_key"] != "timeline.sort_ts" || sortedSort[1].(map[string]any)["direction"] != "asc" {
		t.Fatalf("expected default tail timeline.sort_ts asc, got %#v", sortedSort)
	}
	if sortedSort[2].(map[string]any)["field_key"] != "record_id" || sortedSort[2].(map[string]any)["direction"] != "asc" {
		t.Fatalf("expected stable record_id asc tie-break, got %#v", sortedSort)
	}
	sortedRows := sortedEnvelope["data"].(map[string]any)["rows"].([]any)
	if got := sortedRows[0].(map[string]any)["record_id"]; got != firstID {
		t.Fatalf("expected timeline.summary asc to place Tie A first, got %#v", sortedRows)
	}
	if got := sortedRows[1].(map[string]any)["record_id"]; got != secondID {
		t.Fatalf("expected timeline.summary asc to place Tie B second, got %#v", sortedRows)
	}
	if got := sortedRows[2].(map[string]any)["record_id"]; got != zeroFieldID {
		t.Fatalf("expected null summary row to sort last, got %#v", sortedRows)
	}

	if _, err := db.ExecContext(context.Background(), `DELETE FROM timeline_grid_projection WHERE incident_id::text = $1`, incidentID); err != nil {
		t.Fatalf("clear projection rows: %v", err)
	}
	emptyEnvelope := queryTimelineEnvelope(t, server, incidentID, adminLogin, map[string]any{})
	httptestx.RequireDefaultQueryMeta(t, emptyEnvelope, timeline.TimelineViewSchemaID)
	emptyRows := emptyEnvelope["data"].(map[string]any)["rows"].([]any)
	if len(emptyRows) != 0 {
		t.Fatalf("query route must read projection rows, got %#v", emptyRows)
	}

	projectionStore := projections.NewStore(server.Runtime.Postgres)
	if err := projectionStore.RebuildIncidentTimeline(context.Background(), mustUUID(t, incidentID)); err != nil {
		t.Fatalf("rebuild timeline projection: %v", err)
	}
	rebuiltEnvelope := queryTimelineEnvelope(t, server, incidentID, adminLogin, map[string]any{})
	httptestx.RequireDefaultQueryMeta(t, rebuiltEnvelope, timeline.TimelineViewSchemaID)
	rebuiltRows := rebuiltEnvelope["data"].(map[string]any)["rows"].([]any)
	httptestx.RequireProjectionDeterminism(t, beforeRows, rebuiltRows)

	if got := queryCount(t, db, `SELECT COUNT(*) FROM timeline_events WHERE incident_id::text = $1`, incidentID); got != 3 {
		t.Fatalf("rebuild must preserve source rows, got %d", got)
	}
}

func TestPhase3_I_3_03_AuthorizationLifecycleAndSupersedeTransitions(t *testing.T) {
	runtime := phase3test.StartRuntime(t)

	t.Run("authorization re-derives, reasons normalize, and supersede guards hold", func(t *testing.T) {
		server, db := startPhase3Server(t, runtime, "phase3-i-3-03")
		defer db.Close()

		adminLogin, _ := provisionBootstrapAdmin(t, server)
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
			"client_txn_id":    "txn-i-3-03-replacement",
			"timeline.summary": "Replacement row",
		})
		replacementID := replacement["row"].(map[string]any)["record_id"].(string)
		alternateReplacement := createTimelineRow(t, server, incidentID, adminLogin, map[string]any{
			"client_txn_id":    "txn-i-3-03-alternate-replacement",
			"timeline.summary": "Alternate replacement row",
		})
		alternateReplacementID := alternateReplacement["row"].(map[string]any)["record_id"].(string)
		created := createTimelineRow(t, server, incidentID, adminLogin, map[string]any{
			"client_txn_id":    "txn-i-3-03-primary",
			"timeline.summary": "Primary row",
		})
		recordID := created["row"].(map[string]any)["record_id"].(string)
		otherReplacement := createTimelineRow(t, server, otherIncidentID, adminLogin, map[string]any{
			"client_txn_id":    "txn-i-3-03-cross-incident",
			"timeline.summary": "Cross incident replacement",
		})
		otherReplacementID := otherReplacement["row"].(map[string]any)["record_id"].(string)

		reviewerSession, reviewerCSRF := loginLocalUser(t, server, "reviewer-target@example.test", "ReviewerTargetPass1!")
		reviewerLogin := loginResult{sessionCookie: reviewerSession, csrfCookie: reviewerCSRF}
		socket := connectTimelineSocket(t, server, incidentID, reviewerSession.Value)
		defer socket.Close(1000, "test_complete")
		hubChanges, unsubscribe := server.Runtime.WSHub.SubscribeRecordChanges(16)
		defer unsubscribe()

		reviewDenied := doPhase3JSON(
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
		beforeAuth := httptestx.AuthorizationOutcome{Status: http.StatusForbidden, Code: "authorization_denied"}

		markReviewed := doPhase3JSON(
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
		timelinetest.AwaitRecordChange(t, hubChanges, 5*time.Second)
		requireMutationRecorded(t, db, reviewedData["change_set_id"].(string), recordID, reviewerID, "timeline.records.mark_reviewed", "txn-i-3-03-reviewed-1", 1, 2)
		if projection := timelinetest.LookupProjectionRow(t, db, recordID); projection.CaptureState != "reviewed" {
			t.Fatalf("expected reviewed projection row, got %#v", projection)
		}
		httptestx.RequireAuthorizationReDerived(t, beforeAuth, httptestx.AuthorizationOutcome{Status: http.StatusOK})

		reviewCounts := timelinetest.SnapshotCounters(t, db, incidentID, recordID)
		reviewReplay := doPhase3JSON(
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
		reviewCountsAfterReplay := timelinetest.SnapshotCounters(t, db, incidentID, recordID)
		httptestx.RequireReplayScaffold(t, httptestx.ReplayExpectation{
			FirstStatus:     http.StatusOK,
			ReplayStatus:    http.StatusOK,
			DivergentStatus: http.StatusConflict,
			DivergentCode:   "client_txn_conflict",
			StableBefore: httptestx.ReplayCounts{
				ChangeSets:   reviewCounts.ChangeSets,
				MutationRows: reviewCounts.MutationRows,
				Revisions:    reviewCounts.Revisions,
			},
			StableAfter: httptestx.ReplayCounts{
				ChangeSets:   reviewCountsAfterReplay.ChangeSets,
				MutationRows: reviewCountsAfterReplay.MutationRows,
				Revisions:    reviewCountsAfterReplay.Revisions,
			},
		})

		reviewDivergent := doPhase3JSON(
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

		materialEdit := doPhase3JSON(
			t,
			http.MethodPatch,
			server.HTTP.URL+"/api/v1/records/"+recordID,
			map[string]any{
				"view_schema_id":   timeline.TimelineViewSchemaID,
				"base_row_version": 2,
				"client_txn_id":    "txn-i-3-03-demote",
				"changes": []map[string]any{
					{"field_key": "timeline.details", "value": "Material edit after review"},
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
		timelinetest.AwaitRecordChange(t, hubChanges, 5*time.Second)

		selfSupersede := doPhase3JSON(
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
		httptestx.RequireErrorEnvelope(t, selfSupersede, http.StatusConflict, "illegal_transition")
		requireNoTimelineCollaborationEmission(t, socket, hubChanges)

		crossIncidentSupersede := doPhase3JSON(
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
		httptestx.RequireErrorEnvelope(t, crossIncidentSupersede, http.StatusConflict, "illegal_transition")
		requireNoTimelineCollaborationEmission(t, socket, hubChanges)

		supersede := doPhase3JSON(
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
		timelinetest.AwaitRecordChange(t, hubChanges, 5*time.Second)
		requireMutationRecorded(t, db, superseded["change_set_id"].(string), recordID, reviewerID, "timeline.records.supersede", "txn-i-3-03-supersede", 2, 4)
		timelinetest.RequireSupersedeCoupledChangeSet(t, db, superseded["change_set_id"].(string), recordID, replacementID, 4)
		if got := timelinetest.CountActiveSupersedesLinks(t, db, incidentID, replacementID, recordID); got != 1 {
			t.Fatalf("expected one active supersedes link, got %d", got)
		}
		if projection := timelinetest.LookupProjectionRow(t, db, recordID); projection.CaptureState != "superseded" || projection.ReplacementRecordID == nil || *projection.ReplacementRecordID != replacementID {
			t.Fatalf("unexpected superseded projection row: %#v", projection)
		}

		queryEnvelope := queryTimelineEnvelope(t, server, incidentID, reviewerLogin, map[string]any{})
		httptestx.RequireDefaultQueryMeta(t, queryEnvelope, timeline.TimelineViewSchemaID)
		supersededRow := findRow(t, queryEnvelope["data"].(map[string]any)["rows"].([]any), recordID)
		if supersededRow["cells"].(map[string]any)["timeline.replacement_record_id"].(map[string]any)["value"] != replacementID {
			t.Fatalf("expected query to surface replacement_record_id, got %#v", supersededRow)
		}
		if supersededRow["group_values"].(map[string]any)["timeline.capture_state"] != "superseded" {
			t.Fatalf("expected query to surface superseded capture_state, got %#v", supersededRow)
		}

		supersedeCounts := timelinetest.SnapshotCounters(t, db, incidentID, recordID)
		supersedeReplay := doPhase3JSON(
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
		supersedeCountsAfterReplay := timelinetest.SnapshotCounters(t, db, incidentID, recordID)
		httptestx.RequireReplayScaffold(t, httptestx.ReplayExpectation{
			FirstStatus:     http.StatusOK,
			ReplayStatus:    http.StatusOK,
			DivergentStatus: http.StatusConflict,
			DivergentCode:   "client_txn_conflict",
			StableBefore: httptestx.ReplayCounts{
				ChangeSets:   supersedeCounts.ChangeSets,
				MutationRows: supersedeCounts.MutationRows,
				Revisions:    supersedeCounts.Revisions,
			},
			StableAfter: httptestx.ReplayCounts{
				ChangeSets:   supersedeCountsAfterReplay.ChangeSets,
				MutationRows: supersedeCountsAfterReplay.MutationRows,
				Revisions:    supersedeCountsAfterReplay.Revisions,
			},
		})

		supersedeDivergentReason := doPhase3JSON(
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

		supersedeDivergentReplacement := doPhase3JSON(
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

		patchAfterSupersede := doPhase3JSON(
			t,
			http.MethodPatch,
			server.HTTP.URL+"/api/v1/records/"+recordID,
			map[string]any{
				"view_schema_id":   timeline.TimelineViewSchemaID,
				"base_row_version": 4,
				"client_txn_id":    "txn-i-3-03-post-supersede-patch",
				"changes": []map[string]any{
					{"field_key": "timeline.summary", "value": "must fail while superseded"},
				},
			},
			withCookies(reviewerSession, reviewerCSRF),
			withHeader(authn.CSRFHeaderName, reviewerCSRF.Value),
		)
		httptestx.RequireErrorEnvelope(t, patchAfterSupersede, http.StatusConflict, "illegal_transition")
		requireNoTimelineCollaborationEmission(t, socket, hubChanges)

		illegalReview := doPhase3JSON(
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
		patchDeniedAgain := doPhase3JSON(
			t,
			http.MethodPatch,
			server.HTTP.URL+"/api/v1/records/"+recordID,
			map[string]any{
				"view_schema_id":   timeline.TimelineViewSchemaID,
				"base_row_version": 4,
				"client_txn_id":    "txn-i-3-03-post-downgrade",
				"changes": []map[string]any{
					{"field_key": "timeline.summary", "value": "must fail"},
				},
			},
			withCookies(reviewerSession, reviewerCSRF),
			withHeader(authn.CSRFHeaderName, reviewerCSRF.Value),
		)
		httptestx.RequireErrorEnvelope(t, patchDeniedAgain, http.StatusForbidden, "authorization_denied")
		requireNoTimelineCollaborationEmission(t, socket, hubChanges)
		httptestx.RequireAuthorizationReDerived(t, httptestx.AuthorizationOutcome{Status: http.StatusOK}, httptestx.AuthorizationOutcome{Status: http.StatusForbidden, Code: "authorization_denied"})
	})

	t.Run("supersede rollback clears source history projection link and collaboration", func(t *testing.T) {
		restoreHooks := timeline.SetStoreHooksForTesting(timeline.StoreHooks{
			BeforeCommit: func(routeKey string, recordID uuid.UUID) error {
				if routeKey == "timeline.records.supersede" {
					return errors.New("forced supersede rollback")
				}
				return nil
			},
		})
		defer restoreHooks()

		server, db := startPhase3Server(t, runtime, "phase3-i-3-03-rollback")
		defer db.Close()

		adminLogin, _ := provisionBootstrapAdmin(t, server)
		incident := createIncident(t, server, adminLogin, map[string]any{
			"client_txn_id": "txn-i-3-03-rollback-incident",
			"incident_key":  "IR-I303R",
			"title":         "Supersede rollback",
		})
		incidentID := incident["incident_id"].(string)
		replacement := createTimelineRow(t, server, incidentID, adminLogin, map[string]any{
			"client_txn_id":    "txn-i-3-03-rollback-replacement",
			"timeline.summary": "Replacement row",
		})
		replacementID := replacement["row"].(map[string]any)["record_id"].(string)
		created := createTimelineRow(t, server, incidentID, adminLogin, map[string]any{
			"client_txn_id":    "txn-i-3-03-rollback-primary",
			"timeline.summary": "Primary row",
		})
		recordID := created["row"].(map[string]any)["record_id"].(string)

		socket := connectTimelineSocket(t, server, incidentID, adminLogin.sessionCookie.Value)
		defer socket.Close(1000, "test_complete")
		hubChanges, unsubscribe := server.Runtime.WSHub.SubscribeRecordChanges(8)
		defer unsubscribe()

		before := timelinetest.SnapshotCounters(t, db, incidentID, recordID)
		supersede := doPhase3JSON(
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

		after := timelinetest.SnapshotCounters(t, db, incidentID, recordID)
		if before != after {
			t.Fatalf("expected rollback to keep counters stable, before=%+v after=%+v", before, after)
		}
		if got := timelinetest.CountActiveSupersedesLinks(t, db, incidentID, replacementID, recordID); got != 0 {
			t.Fatalf("rollback must clear supersedes link, got %d", got)
		}
		projection := timelinetest.LookupProjectionRow(t, db, recordID)
		if projection.CaptureState != "rough" || projection.ReplacementRecordID != nil {
			t.Fatalf("rollback must restore pre-supersede projection row, got %#v", projection)
		}
		envelope := queryTimelineEnvelope(t, server, incidentID, adminLogin, map[string]any{})
		httptestx.RequireDefaultQueryMeta(t, envelope, timeline.TimelineViewSchemaID)
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

type loginResult struct {
	sessionCookie *http.Cookie
	csrfCookie    *http.Cookie
}

func toPhase3Login(login loginResult) phase3test.LoginResult {
	return phase3test.LoginResult{
		SessionCookie: login.sessionCookie,
		CSRFCookie:    login.csrfCookie,
	}
}

type recordChangeSocketPayload = phase3test.RecordChangeSocketPayload

type timelineSocketClient = phase3test.TimelineSocketClient

func startPhase3Server(t testing.TB, runtime *phase3test.RuntimeHarness, prefix string) (*httptestx.Server, *sql.DB) {
	t.Helper()

	harness := runtime.StartServer(t, prefix)
	return harness.Server, harness.DB
}

func provisionBootstrapAdmin(t testing.TB, server *httptestx.Server) (loginResult, string) {
	t.Helper()

	login, userID := phase3test.ProvisionBootstrapAdmin(t, server)
	return loginResult{
		sessionCookie: login.SessionCookie,
		csrfCookie:    login.CSRFCookie,
	}, userID.String()
}

func createIncident(t testing.TB, server *httptestx.Server, admin loginResult, body map[string]any) map[string]any {
	t.Helper()

	return phase3test.CreateIncident(t, server, toPhase3Login(admin), body)
}

func createTimelineRow(t testing.TB, server *httptestx.Server, incidentID string, admin loginResult, body map[string]any) map[string]any {
	t.Helper()

	return phase3test.CreateTimelineRow(t, server, incidentID, toPhase3Login(admin), body)
}

func createMembership(t testing.TB, server *httptestx.Server, incidentID string, userID string, email string, role string, admin loginResult) {
	t.Helper()

	phase3test.CreateMembership(t, server, incidentID, userID, email, role, toPhase3Login(admin))
}

func updateMembershipRole(t testing.TB, server *httptestx.Server, incidentID string, userID string, baseVersion int, role string, admin loginResult) {
	t.Helper()

	phase3test.UpdateMembershipRole(t, server, incidentID, userID, baseVersion, role, toPhase3Login(admin))
}

func seedLocalUserFlags(t testing.TB, db *sql.DB, email string, displayName string, password string, mfaRequired bool, isDeploymentAdmin bool, isActive bool) string {
	t.Helper()

	return phase3test.SeedLocalUserFlags(t, db, email, displayName, password, mfaRequired, isDeploymentAdmin, isActive).ID.String()
}

func loginLocalUser(t testing.TB, server *httptestx.Server, username string, password string) (*http.Cookie, *http.Cookie) {
	t.Helper()

	return phase3test.LoginLocalUser(t, server, username, password)
}

func connectTimelineSocket(t testing.TB, server *httptestx.Server, incidentID string, sessionToken string) *timelineSocketClient {
	t.Helper()

	return phase3test.ConnectTimelineSocket(t, server, incidentID, sessionToken)
}

func requireTimelineSocketChange(t testing.TB, client *timelineSocketClient, wantRecordID string, wantRowVersion int64) recordChangeSocketPayload {
	t.Helper()

	return phase3test.RequireTimelineSocketChange(t, client, wantRecordID, wantRowVersion)
}

func expectNoTimelineSocketMessage(t testing.TB, client *timelineSocketClient) {
	t.Helper()

	phase3test.ExpectNoTimelineSocketMessage(t, client)
}

func requireMutationRecorded(t testing.TB, db *sql.DB, changeSetID string, recordID string, wantActorUserID string, wantSource string, wantClientTxnID string, wantMutationRows int, wantRevisions int) {
	t.Helper()

	phase3test.RequireMutationRecorded(t, db, changeSetID, recordID, wantActorUserID, wantSource, wantClientTxnID, wantMutationRows, wantRevisions)
}

func requireNoTimelineCollaborationEmission(t testing.TB, client *timelineSocketClient, changes <-chan platformws.RecordChange) {
	t.Helper()

	phase3test.RequireNoTimelineCollaborationEmission(t, client, changes)
}

func queryTimelineEnvelope(t testing.TB, server *httptestx.Server, incidentID string, login loginResult, body map[string]any) map[string]any {
	t.Helper()

	return phase3test.QueryTimelineEnvelope(t, server, incidentID, toPhase3Login(login), body)
}

func queryTimelineRows(t testing.TB, server *httptestx.Server, incidentID string, login loginResult) []any {
	t.Helper()

	return phase3test.QueryTimelineRows(t, server, incidentID, toPhase3Login(login))
}

func findRow(t testing.TB, rows []any, recordID string) map[string]any {
	t.Helper()

	return phase3test.FindRow(t, rows, recordID)
}

func mustUUID(t testing.TB, raw string) uuid.UUID {
	t.Helper()

	return phase3test.MustUUID(t, raw)
}

func requireBootstrapLogin(t testing.TB, server *httptestx.Server, username string, password string) string {
	t.Helper()

	login := phase3test.RequireBootstrapLogin(t, server, username, password)
	return login
}

func beginTOTPEnrollment(t testing.TB, server *httptestx.Server, bootstrapToken string, body map[string]any) map[string]any {
	t.Helper()

	return phase3test.BeginTOTPEnrollment(t, server, bootstrapToken, body)
}

func completeInitialEnrollment(t testing.TB, server *httptestx.Server, bootstrapToken string, enrollmentID string, secretBase32 string, clientTxnID string) {
	t.Helper()

	phase3test.CompleteInitialEnrollment(t, server, bootstrapToken, enrollmentID, secretBase32, clientTxnID)
}

func loginLocalUserWithSecondFactor(t testing.TB, server *httptestx.Server, username string, password string, code string) loginResult {
	t.Helper()

	login := phase3test.LoginLocalUserWithSecondFactor(t, server, username, password, code)
	return loginResult{
		sessionCookie: login.SessionCookie,
		csrfCookie:    login.CSRFCookie,
	}
}

func generateTOTPCode(t testing.TB, secretBase32 string) string {
	t.Helper()

	return phase3test.GenerateTOTPCode(t, secretBase32)
}

func doPhase3JSON(t testing.TB, method string, url string, body any, options ...func(*http.Request)) *http.Response {
	t.Helper()

	return phase3test.DoJSON(t, method, url, body, options...)
}

func withCookies(cookies ...*http.Cookie) func(*http.Request) {
	return phase3test.WithCookies(cookies...)
}

func withHeader(key string, value string) func(*http.Request) {
	return phase3test.WithHeader(key, value)
}

func queryCount(t testing.TB, db *sql.DB, query string, args ...any) int {
	t.Helper()

	return phase3test.QueryCount(t, db, query, args...)
}
