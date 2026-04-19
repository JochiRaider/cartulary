package timeline_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"

	"github.com/JochiRaider/cartulary/internal/modules/projections"
	"github.com/JochiRaider/cartulary/internal/modules/timeline"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/fixtures"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
	"github.com/JochiRaider/cartulary/internal/testutil/s3test"
	"github.com/JochiRaider/cartulary/internal/testutil/timelinetest"
	"github.com/JochiRaider/cartulary/internal/testutil/wstest"
)

func TestPhase3_I_3_01_CreatePatchReplayAndRollback(t *testing.T) {
	postgresHarness := pgtest.Start(t)
	s3Harness := s3test.Start(t)

	t.Run("same replay stays single-write and divergent replay conflicts", func(t *testing.T) {
		server, db := startPhase3Server(t, postgresHarness, s3Harness, "phase3-i-3-01-replay")
		defer db.Close()

		adminLogin, adminID := provisionBootstrapAdmin(t, server)
		incident := createIncident(t, server, adminLogin, map[string]any{
			"client_txn_id": "txn-i-3-01-incident",
			"incident_key":  "IR-I301",
			"title":         "Timeline substrate",
		})
		incidentID := incident["incident_id"].(string)
		socket := connectTimelineSocket(t, server, incidentID, adminLogin.sessionCookie.Value)
		defer socket.Close(1000, "test_complete")
		hubChanges, unsubscribe := server.Runtime.WSHub.SubscribeRecordChanges(16)
		defer unsubscribe()

		createResp := doPhase3JSON(
			t,
			http.MethodPost,
			server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/views/"+timeline.TimelineViewSchemaID+"/rows",
			map[string]any{
				"client_txn_id":    "txn-i-3-01-row-create",
				"timeline.summary": "Initial capture",
			},
			withCookies(adminLogin.sessionCookie, adminLogin.csrfCookie),
			withHeader(authn.CSRFHeaderName, adminLogin.csrfCookie.Value),
		)
		createData := httptestx.RequireSuccessEnvelope(t, createResp, http.StatusCreated)["data"].(map[string]any)
		createRow := createData["row"].(map[string]any)
		recordID := createRow["record_id"].(string)
		requireTimelineSocketChange(t, socket, recordID, 1)
		timelinetest.AwaitRecordChange(t, hubChanges, 5*time.Second)

		createAttribution := timelinetest.LookupChangeSet(t, db, createData["change_set_id"].(string))
		httptestx.RequireMutationAttribution(t, httptestx.MutationAttribution{
			ActorUserID: createAttribution.ActorUserID,
			Source:      createAttribution.Source,
			ClientTxnID: createAttribution.ClientTxnID,
			RequestID:   createAttribution.RequestID,
			CreatedAt:   createAttribution.CreatedAt,
		}, adminID, "timeline.rows.create", "txn-i-3-01-row-create")

		countersAfterCreate := timelinetest.SnapshotCounters(t, db, incidentID, recordID)
		createReplay := doPhase3JSON(
			t,
			http.MethodPost,
			server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/views/"+timeline.TimelineViewSchemaID+"/rows",
			map[string]any{
				"client_txn_id":    "txn-i-3-01-row-create",
				"timeline.summary": "Initial capture",
			},
			withCookies(adminLogin.sessionCookie, adminLogin.csrfCookie),
			withHeader(authn.CSRFHeaderName, adminLogin.csrfCookie.Value),
		)
		createReplayData := httptestx.RequireSuccessEnvelope(t, createReplay, http.StatusOK)["data"].(map[string]any)
		if createReplayData["change_set_id"] != createData["change_set_id"] {
			t.Fatalf("expected create replay to return original payload, got %#v", createReplayData)
		}
		timelinetest.RequireNoRecordChange(t, hubChanges, 300*time.Millisecond)
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
				"client_txn_id":    "txn-i-3-01-row-create",
				"timeline.summary": "Different capture",
			},
			withCookies(adminLogin.sessionCookie, adminLogin.csrfCookie),
			withHeader(authn.CSRFHeaderName, adminLogin.csrfCookie.Value),
		)
		httptestx.RequireErrorEnvelope(t, divergentCreate, http.StatusConflict, "client_txn_conflict")
		timelinetest.RequireNoRecordChange(t, hubChanges, 300*time.Millisecond)

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

		patchAttribution := timelinetest.LookupChangeSet(t, db, patchData["change_set_id"].(string))
		httptestx.RequireMutationAttribution(t, httptestx.MutationAttribution{
			ActorUserID: patchAttribution.ActorUserID,
			Source:      patchAttribution.Source,
			ClientTxnID: patchAttribution.ClientTxnID,
			RequestID:   patchAttribution.RequestID,
			CreatedAt:   patchAttribution.CreatedAt,
		}, adminID, "timeline.records.patch", "txn-i-3-01-row-patch")

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
		timelinetest.RequireNoRecordChange(t, hubChanges, 300*time.Millisecond)
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
		timelinetest.RequireNoRecordChange(t, hubChanges, 300*time.Millisecond)
	})

	t.Run("late transaction failures roll back source history projection and collaboration", func(t *testing.T) {
		restoreHooks := timeline.SetStoreHooksForTesting(timeline.StoreHooks{
			BeforeCommit: func(routeKey string, recordID uuid.UUID) error {
				return errors.New("forced timeline rollback")
			},
		})
		defer restoreHooks()

		server, db := startPhase3Server(t, postgresHarness, s3Harness, "phase3-i-3-01-rollback")
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

func TestPhase3_I_3_02_ProjectionQueryUsesDeterministicRebuild(t *testing.T) {
	postgresHarness := pgtest.Start(t)
	s3Harness := s3test.Start(t)

	server, db := startPhase3Server(t, postgresHarness, s3Harness, "phase3-i-3-02")
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
		"timeline.summary":     "Earlier",
		"timeline.occurred_at": "2026-04-10T10:00:00Z",
	})
	second := createTimelineRow(t, server, incidentID, adminLogin, map[string]any{
		"client_txn_id":        "txn-i-3-02-row-b",
		"timeline.summary":     "Later",
		"timeline.occurred_at": "2026-04-10T11:00:00Z",
	})
	_ = first
	secondID := second["row"].(map[string]any)["record_id"].(string)

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

	beforeRows := queryTimelineRows(t, server, incidentID, adminLogin)
	if len(beforeRows) != 2 {
		t.Fatalf("expected two projected rows, got %#v", beforeRows)
	}
	secondRow := findRow(t, beforeRows, secondID)
	if secondRow["row_version"] != float64(2) {
		t.Fatalf("expected patched row_version in projection query, got %#v", secondRow)
	}
	if details := secondRow["cells"].(map[string]any)["timeline.details"].(map[string]any)["value"]; details != "Projected details" {
		t.Fatalf("expected projection-backed details, got %#v", secondRow)
	}

	if _, err := db.ExecContext(context.Background(), `DELETE FROM timeline_grid_projection WHERE incident_id::text = $1`, incidentID); err != nil {
		t.Fatalf("clear projection rows: %v", err)
	}
	emptyRows := queryTimelineRows(t, server, incidentID, adminLogin)
	if len(emptyRows) != 0 {
		t.Fatalf("query route must read projection rows, got %#v", emptyRows)
	}

	projectionStore := projections.NewStore(server.Runtime.Postgres)
	if err := projectionStore.RebuildIncidentTimeline(context.Background(), mustUUID(t, incidentID)); err != nil {
		t.Fatalf("rebuild timeline projection: %v", err)
	}
	rebuiltRows := queryTimelineRows(t, server, incidentID, adminLogin)
	httptestx.RequireProjectionDeterminism(t, beforeRows, rebuiltRows)

	if got := queryCount(t, db, `SELECT COUNT(*) FROM timeline_events WHERE incident_id::text = $1`, incidentID); got != 2 {
		t.Fatalf("rebuild must preserve source rows, got %d", got)
	}
}

func TestPhase3_I_3_03_AuthorizationLifecycleAndSupersedeTransitions(t *testing.T) {
	postgresHarness := pgtest.Start(t)
	s3Harness := s3test.Start(t)

	server, db := startPhase3Server(t, postgresHarness, s3Harness, "phase3-i-3-03")
	defer db.Close()

	adminLogin, _ := provisionBootstrapAdmin(t, server)
	reviewerID := seedLocalUserFlags(t, db, "reviewer-target@example.test", "Reviewer Target", "ReviewerTargetPass1!", false, false, true)

	incident := createIncident(t, server, adminLogin, map[string]any{
		"client_txn_id": "txn-i-3-03-incident",
		"incident_key":  "IR-I303",
		"title":         "Lifecycle",
	})
	incidentID := incident["incident_id"].(string)
	createMembership(t, server, incidentID, reviewerID, "reviewer-target@example.test", "viewer", adminLogin)

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
			"reason":           "viewer cannot review",
		},
		withCookies(reviewerSession, reviewerCSRF),
		withHeader(authn.CSRFHeaderName, reviewerCSRF.Value),
	)
	httptestx.RequireErrorEnvelope(t, reviewDenied, http.StatusForbidden, "authorization_denied")

	updateMembershipRole(t, server, incidentID, reviewerID, 1, "reviewer", adminLogin)
	beforeAuth := httptestx.AuthorizationOutcome{Status: http.StatusForbidden, Code: "authorization_denied"}

	markReviewed := doPhase3JSON(
		t,
		http.MethodPost,
		server.HTTP.URL+"/api/v1/records/"+recordID+"/mark-reviewed",
		map[string]any{
			"base_row_version": 1,
			"client_txn_id":    "txn-i-3-03-reviewed-1",
			"reason":           "Initial review",
		},
		withCookies(reviewerSession, reviewerCSRF),
		withHeader(authn.CSRFHeaderName, reviewerCSRF.Value),
	)
	reviewedData := httptestx.RequireSuccessEnvelope(t, markReviewed, http.StatusOK)["data"].(map[string]any)
	if reviewedData["capture_state"] != "reviewed" || reviewedData["row_version"] != float64(2) {
		t.Fatalf("unexpected reviewed payload: %#v", reviewedData)
	}
	requireTimelineSocketChange(t, socket, recordID, 2)
	timelinetest.AwaitRecordChange(t, hubChanges, 5*time.Second)
	httptestx.RequireAuthorizationReDerived(t, beforeAuth, httptestx.AuthorizationOutcome{Status: http.StatusOK})

	reviewCounts := timelinetest.SnapshotCounters(t, db, incidentID, recordID)
	reviewReplay := doPhase3JSON(
		t,
		http.MethodPost,
		server.HTTP.URL+"/api/v1/records/"+recordID+"/mark-reviewed",
		map[string]any{
			"base_row_version": 1,
			"client_txn_id":    "txn-i-3-03-reviewed-1",
			"reason":           "Initial review",
		},
		withCookies(reviewerSession, reviewerCSRF),
		withHeader(authn.CSRFHeaderName, reviewerCSRF.Value),
	)
	reviewReplayData := httptestx.RequireSuccessEnvelope(t, reviewReplay, http.StatusOK)["data"].(map[string]any)
	if reviewReplayData["change_set_id"] != reviewedData["change_set_id"] {
		t.Fatalf("expected review replay to return original payload, got %#v", reviewReplayData)
	}
	timelinetest.RequireNoRecordChange(t, hubChanges, 300*time.Millisecond)
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
	timelinetest.RequireNoRecordChange(t, hubChanges, 300*time.Millisecond)

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

	supersede := doPhase3JSON(
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
	superseded := httptestx.RequireSuccessEnvelope(t, supersede, http.StatusOK)["data"].(map[string]any)
	if superseded["capture_state"] != "superseded" || superseded["row_version"] != float64(4) {
		t.Fatalf("unexpected supersede payload: %#v", superseded)
	}
	requireTimelineSocketChange(t, socket, recordID, 4)
	timelinetest.AwaitRecordChange(t, hubChanges, 5*time.Second)
	if got := queryCount(t, db, `SELECT COUNT(*) FROM record_links WHERE incident_id::text = $1 AND src_record_id::text = $2 AND dst_record_id::text = $3 AND link_type = 'supersedes' AND deleted_at IS NULL`, incidentID, replacementID, recordID); got != 1 {
		t.Fatalf("expected one active supersedes link, got %d", got)
	}
	if got := queryCount(t, db, `SELECT COUNT(*) FROM change_set_mutations WHERE change_set_id::text = $1`, superseded["change_set_id"]); got != 2 {
		t.Fatalf("expected one coupled history entry with two mutation rows, got %d", got)
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
	timelinetest.RequireNoRecordChange(t, hubChanges, 300*time.Millisecond)
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

	supersedeDivergent := doPhase3JSON(
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
	httptestx.RequireErrorEnvelope(t, supersedeDivergent, http.StatusConflict, "client_txn_conflict")
	timelinetest.RequireNoRecordChange(t, hubChanges, 300*time.Millisecond)

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
	httptestx.RequireAuthorizationReDerived(t, httptestx.AuthorizationOutcome{Status: http.StatusOK}, httptestx.AuthorizationOutcome{Status: http.StatusForbidden, Code: "authorization_denied"})
}

type loginResult struct {
	sessionCookie *http.Cookie
	csrfCookie    *http.Cookie
}

type recordChangeSocketPayload struct {
	RecordID         string   `json:"record_id"`
	RowVersion       float64  `json:"row_version"`
	ChangeSetID      string   `json:"change_set_id"`
	ClientTxnID      string   `json:"client_txn_id"`
	ChangedFieldKeys []string `json:"changed_field_keys"`
}

func startPhase3Server(t testing.TB, postgresHarness *pgtest.Harness, s3Harness *s3test.Harness, prefix string) (*httptestx.Server, *sql.DB) {
	t.Helper()

	testDB, _, err := postgresHarness.PrepareDatabase(context.Background(), prefix)
	if err != nil {
		t.Fatalf("prepare postgres database: %v", err)
	}
	t.Cleanup(func() {
		if err := postgresHarness.DropDatabase(context.Background(), testDB.Name); err != nil {
			t.Fatalf("drop postgres database: %v", err)
		}
	})

	bucket, err := s3Harness.BootstrapBucket(context.Background(), prefix)
	if err != nil {
		t.Fatalf("bootstrap bucket: %v", err)
	}
	t.Cleanup(func() {
		if err := s3Harness.CleanupBucket(context.Background(), bucket); err != nil {
			t.Logf("cleanup bucket: %v", err)
		}
	})

	env := testDB.Env()
	for key, value := range s3Harness.Env(bucket) {
		env[key] = value
	}
	env["CARTULARY__BOOTSTRAP__FIRST_ADMIN_MANIFEST_PATH"] = fixtures.Path("bootstrap-admin", "canonical.json")

	server := httptestx.StartServer(t, httptestx.ServerOptions{Env: env})
	db, err := sql.Open("pgx", testDB.DSN)
	if err != nil {
		t.Fatalf("open sql db: %v", err)
	}
	return server, db
}

func provisionBootstrapAdmin(t testing.TB, server *httptestx.Server) (loginResult, string) {
	t.Helper()

	bootstrapToken := requireBootstrapLogin(t, server, "bootstrap-admin@example.test", "BootstrapPass1!")
	begin := beginTOTPEnrollment(t, server, bootstrapToken, map[string]any{
		"client_txn_id": "txn-bootstrap-admin-begin",
	})
	secretBase32 := begin["totp_setup"].(map[string]any)["secret_base32"].(string)
	completeInitialEnrollment(t, server, bootstrapToken, begin["enrollment_id"].(string), secretBase32, "txn-bootstrap-admin-complete")
	login := loginLocalUserWithSecondFactor(t, server, "bootstrap-admin@example.test", "BootstrapPass1!", generateTOTPCode(t, secretBase32))

	sessionResp := doPhase3JSON(t, http.MethodGet, server.HTTP.URL+"/api/v1/auth/session", nil, withCookies(login.sessionCookie))
	sessionData := httptestx.RequireSuccessEnvelope(t, sessionResp, http.StatusOK)["data"].(map[string]any)
	return login, sessionData["user_id"].(string)
}

func createIncident(t testing.TB, server *httptestx.Server, admin loginResult, body map[string]any) map[string]any {
	t.Helper()

	resp := doPhase3JSON(
		t,
		http.MethodPost,
		server.HTTP.URL+"/api/v1/incidents",
		body,
		withCookies(admin.sessionCookie, admin.csrfCookie),
		withHeader(authn.CSRFHeaderName, admin.csrfCookie.Value),
	)
	return httptestx.RequireSuccessEnvelope(t, resp, http.StatusCreated)["data"].(map[string]any)
}

func createTimelineRow(t testing.TB, server *httptestx.Server, incidentID string, admin loginResult, body map[string]any) map[string]any {
	t.Helper()

	resp := doPhase3JSON(
		t,
		http.MethodPost,
		server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/views/"+timeline.TimelineViewSchemaID+"/rows",
		body,
		withCookies(admin.sessionCookie, admin.csrfCookie),
		withHeader(authn.CSRFHeaderName, admin.csrfCookie.Value),
	)
	return httptestx.RequireSuccessEnvelope(t, resp, http.StatusCreated)["data"].(map[string]any)
}

func createMembership(t testing.TB, server *httptestx.Server, incidentID string, userID string, email string, role string, admin loginResult) {
	t.Helper()

	resp := doPhase3JSON(
		t,
		http.MethodPost,
		server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/memberships",
		map[string]any{
			"client_txn_id": "txn-phase3-membership-create-" + userID,
			"email":         email,
			"role":          role,
		},
		withCookies(admin.sessionCookie, admin.csrfCookie),
		withHeader(authn.CSRFHeaderName, admin.csrfCookie.Value),
	)
	body := httptestx.RequireSuccessEnvelope(t, resp, http.StatusCreated)["data"].(map[string]any)
	if body["user_id"] != userID {
		t.Fatalf("unexpected membership create payload: %#v", body)
	}
}

func updateMembershipRole(t testing.TB, server *httptestx.Server, incidentID string, userID string, baseVersion int, role string, admin loginResult) {
	t.Helper()

	resp := doPhase3JSON(
		t,
		http.MethodPatch,
		server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/memberships/"+userID,
		map[string]any{
			"base_membership_version": baseVersion,
			"role":                    role,
		},
		withCookies(admin.sessionCookie, admin.csrfCookie),
		withHeader(authn.CSRFHeaderName, admin.csrfCookie.Value),
	)
	httptestx.RequireSuccessEnvelope(t, resp, http.StatusOK)
}

func seedLocalUserFlags(t testing.TB, db *sql.DB, email string, displayName string, password string, mfaRequired bool, isDeploymentAdmin bool, isActive bool) string {
	t.Helper()

	hash, err := authn.HashPassword(password)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	var userID string
	if err := db.QueryRowContext(context.Background(), `
INSERT INTO users (email, display_name, password_hash, mfa_required, is_active, is_deployment_admin)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id::text
`, email, displayName, hash, mfaRequired, isActive, isDeploymentAdmin).Scan(&userID); err != nil {
		t.Fatalf("seed local user with flags: %v", err)
	}
	return userID
}

func loginLocalUser(t testing.TB, server *httptestx.Server, username string, password string) (*http.Cookie, *http.Cookie) {
	t.Helper()

	resp := doPhase3JSON(t, http.MethodPost, server.HTTP.URL+"/api/v1/auth/login", map[string]any{
		"username": username,
		"password": password,
	})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("login failed: status=%d body=%#v", resp.StatusCode, httptestx.ReadJSONBody(t, resp))
	}
	httptestx.RequireSuccessEnvelope(t, resp, http.StatusOK)

	var sessionCookie *http.Cookie
	var csrfCookie *http.Cookie
	for _, cookie := range resp.Cookies() {
		switch cookie.Name {
		case authn.SessionCookieName:
			sessionCookie = cookie
		case authn.CSRFCookieName:
			csrfCookie = cookie
		}
	}
	if sessionCookie == nil || csrfCookie == nil {
		t.Fatalf("expected login to set both session and csrf cookies, got %#v", resp.Cookies())
	}
	return sessionCookie, csrfCookie
}

func connectTimelineSocket(t testing.TB, server *httptestx.Server, incidentID string, sessionToken string) *wstest.Client {
	t.Helper()

	headers := http.Header{}
	headers.Set("Authorization", "Bearer "+sessionToken)
	client := wstest.ConnectWithHeaders(t, server.HTTP.URL, "/ws/v1/incidents/"+incidentID+"/views/"+timeline.TimelineViewSchemaID+"/changes", headers)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	message, err := client.Receive(ctx)
	if err != nil {
		t.Fatalf("receive websocket connected message: %v", err)
	}
	wstest.RequireMessageType(t, message, "connected")
	return client
}

func requireTimelineSocketChange(t testing.TB, client *wstest.Client, wantRecordID string, wantRowVersion int64) recordChangeSocketPayload {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	message, err := client.Receive(ctx)
	if err != nil {
		t.Fatalf("receive websocket record_changed: %v", err)
	}
	wstest.RequireMessageType(t, message, "record_changed")

	var payload recordChangeSocketPayload
	if err := json.Unmarshal(message.Payload, &payload); err != nil {
		t.Fatalf("decode record_changed payload: %v", err)
	}
	if payload.RecordID != wantRecordID || payload.RowVersion != float64(wantRowVersion) {
		t.Fatalf("unexpected record_changed payload: %#v", payload)
	}
	if payload.ClientTxnID == "" {
		t.Fatalf("expected websocket payload client_txn_id, got %#v", payload)
	}
	return payload
}

func expectNoTimelineSocketMessage(t testing.TB, client *wstest.Client) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	_, err := client.Receive(ctx)
	if err == nil {
		t.Fatal("expected no websocket message")
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected websocket receive timeout, got %v", err)
	}
}

func queryTimelineRows(t testing.TB, server *httptestx.Server, incidentID string, login loginResult) []any {
	t.Helper()

	queryResp := doPhase3JSON(
		t,
		http.MethodPost,
		server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/views/"+timeline.TimelineViewSchemaID+"/query",
		map[string]any{},
		withCookies(login.sessionCookie),
	)
	return httptestx.RequireSuccessEnvelope(t, queryResp, http.StatusOK)["data"].(map[string]any)["rows"].([]any)
}

func findRow(t testing.TB, rows []any, recordID string) map[string]any {
	t.Helper()

	for _, candidate := range rows {
		row := candidate.(map[string]any)
		if row["record_id"] == recordID {
			return row
		}
	}
	t.Fatalf("record_id %s not found in rows %#v", recordID, rows)
	return nil
}

func mustUUID(t testing.TB, raw string) uuid.UUID {
	t.Helper()

	value, err := uuid.Parse(raw)
	if err != nil {
		t.Fatalf("parse uuid: %v", err)
	}
	return value
}

func requireBootstrapLogin(t testing.TB, server *httptestx.Server, username string, password string) string {
	t.Helper()

	resp := doPhase3JSON(t, http.MethodPost, server.HTTP.URL+"/api/v1/auth/login", map[string]any{
		"username": username,
		"password": password,
	})
	body := httptestx.RequireErrorEnvelope(t, resp, http.StatusUnauthorized, "mfa_setup_required")
	details := body["error"].(map[string]any)["details"].(map[string]any)
	return details["bootstrap_token"].(string)
}

func beginTOTPEnrollment(t testing.TB, server *httptestx.Server, bootstrapToken string, body map[string]any) map[string]any {
	t.Helper()

	resp := doPhase3JSON(t, http.MethodPost, server.HTTP.URL+"/api/v1/auth/mfa/totp/begin", body, withHeader("Authorization", "Bearer "+bootstrapToken))
	return httptestx.RequireSuccessEnvelope(t, resp, http.StatusOK)["data"].(map[string]any)
}

func completeInitialEnrollment(t testing.TB, server *httptestx.Server, bootstrapToken string, enrollmentID string, secretBase32 string, clientTxnID string) {
	t.Helper()

	resp := doPhase3JSON(t, http.MethodPost, server.HTTP.URL+"/api/v1/auth/mfa/totp/complete", map[string]any{
		"client_txn_id": clientTxnID,
		"enrollment_id": enrollmentID,
		"code":          generateTOTPCode(t, secretBase32),
	}, withHeader("Authorization", "Bearer "+bootstrapToken))
	httptestx.RequireSuccessEnvelope(t, resp, http.StatusOK)
}

func loginLocalUserWithSecondFactor(t testing.TB, server *httptestx.Server, username string, password string, code string) loginResult {
	t.Helper()

	resp := doPhase3JSON(t, http.MethodPost, server.HTTP.URL+"/api/v1/auth/login", map[string]any{
		"username": username,
		"password": password,
		"second_factor": map[string]any{
			"kind": "totp",
			"assertion": map[string]any{
				"code": code,
			},
		},
	})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("login with second factor failed: status=%d body=%#v", resp.StatusCode, httptestx.ReadJSONBody(t, resp))
	}
	httptestx.RequireSuccessEnvelope(t, resp, http.StatusOK)

	var sessionCookie *http.Cookie
	var csrfCookie *http.Cookie
	for _, cookie := range resp.Cookies() {
		switch cookie.Name {
		case authn.SessionCookieName:
			sessionCookie = cookie
		case authn.CSRFCookieName:
			csrfCookie = cookie
		}
	}
	if sessionCookie == nil || csrfCookie == nil {
		t.Fatalf("expected login to set both session and csrf cookies, got %#v", resp.Cookies())
	}
	return loginResult{sessionCookie: sessionCookie, csrfCookie: csrfCookie}
}

func generateTOTPCode(t testing.TB, secretBase32 string) string {
	t.Helper()

	code, err := totp.GenerateCodeCustom(secretBase32, time.Now().UTC(), totp.ValidateOpts{
		Period:    30,
		Skew:      1,
		Digits:    otp.DigitsSix,
		Algorithm: otp.AlgorithmSHA1,
	})
	if err != nil {
		t.Fatalf("generate totp code: %v", err)
	}
	return code
}

func doPhase3JSON(t testing.TB, method string, url string, body any, options ...func(*http.Request)) *http.Response {
	t.Helper()

	req := httptestx.NewJSONRequest(t, method, url, body)
	for _, option := range options {
		option(req)
	}
	return httptestx.Do(t, http.DefaultClient, req)
}

func withCookies(cookies ...*http.Cookie) func(*http.Request) {
	return func(req *http.Request) {
		for _, cookie := range cookies {
			if cookie != nil {
				req.AddCookie(cookie)
			}
		}
	}
}

func withHeader(key string, value string) func(*http.Request) {
	return func(req *http.Request) {
		req.Header.Set(key, value)
	}
}

func queryCount(t testing.TB, db *sql.DB, query string, args ...any) int {
	t.Helper()

	var count int
	if err := db.QueryRowContext(context.Background(), query, args...).Scan(&count); err != nil {
		t.Fatalf("query count: %v", err)
	}
	return count
}
