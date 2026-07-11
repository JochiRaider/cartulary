package entities_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"fmt"
	"io"
	"net/http"
	"slices"
	"strings"
	"testing"

	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/JochiRaider/cartulary/internal/modules/projections"
	"github.com/JochiRaider/cartulary/internal/modules/records/testsupport/golden"
	workbookscenariotest "github.com/JochiRaider/cartulary/internal/modules/workbook/testsupport/scenariotest"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/contractassert"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
	"github.com/JochiRaider/cartulary/internal/testutil/wstest"
)

func TestSupportPhase4Integration_SurfaceEnvelope(t *testing.T) {
	suite := newPhase4SupportSuite(t, "surface-envelope")
	for _, route := range workbookscenariotest.RoutesForHarness(
		t,
		workbookscenariotest.WorkbookRouteInventory(workbookscenariotest.RouteInventoryContext{}),
		workbookscenariotest.RouteHarnessSurfaceEnvelope,
	) {
		t.Run(string(route.Key), func(t *testing.T) {
			scenario := suite.newScenario(t, route)
			data := scenario.requireRouteSuccess(t, route, supportTxn("surface", route.Key), nil)
			assertRouteSuccessShape(t, route, scenario, data)
		})
	}
}

func TestSupportPhase4Integration_CSRFProtection(t *testing.T) {
	suite := newPhase4SupportSuite(t, "csrf")
	for _, route := range workbookscenariotest.RoutesForHarness(
		t,
		workbookscenariotest.WorkbookRouteInventory(workbookscenariotest.RouteInventoryContext{}),
		workbookscenariotest.RouteHarnessCSRF,
	) {
		t.Run(string(route.Key), func(t *testing.T) {
			scenario := suite.newScenario(t, route)
			resp := scenario.doRoute(t, route, supportTxn("csrf", route.Key), nil, false)
			workbookscenariotest.RequireErrorBody(t, resp, http.StatusForbidden, "csrf_verification_failed")
		})
	}
}

func TestSupportPhase4Integration_ReplayAndDivergentConflict(t *testing.T) {
	suite := newPhase4SupportSuite(t, "replay")
	for _, route := range workbookscenariotest.RoutesForHarness(
		t,
		workbookscenariotest.WorkbookRouteInventory(workbookscenariotest.RouteInventoryContext{}),
		workbookscenariotest.RouteHarnessReplayDivergent,
	) {
		t.Run(string(route.Key), func(t *testing.T) {
			scenario := suite.newScenario(t, route)
			clientTxnID := supportTxn("replay", route.Key)

			firstData := scenario.requireRouteSuccess(t, route, clientTxnID, nil)
			replayResp := scenario.doRoute(t, route, clientTxnID, nil, true)
			replayData := scenario.requireRouteSuccessStatus(t, route, replayResp, route.ReplayStatus)
			requireStableReplayPayload(t, route, firstData, replayData)

			divergentResp := scenario.doRoute(t, route, clientTxnID, route.BuildDivergentBody(scenario.routeCtx, clientTxnID), true)
			divergentBody := workbookscenariotest.RequireErrorBody(t, divergentResp, route.DivergentStatus, route.DivergentCode)
			contractassert.RequireDivergentReplayRejected(
				t,
				divergentResp.StatusCode,
				divergentBody["error"].(map[string]any)["code"].(string),
				route.DivergentCode,
			)
		})
	}
}

func TestSupportPhase4Integration_AuthorizationReDerivation(t *testing.T) {
	suite := newPhase4SupportSuite(t, "authorization")
	for _, route := range workbookscenariotest.RoutesForHarness(
		t,
		workbookscenariotest.WorkbookRouteInventory(workbookscenariotest.RouteInventoryContext{}),
		workbookscenariotest.RouteHarnessAuthorization,
	) {
		t.Run(string(route.Key), func(t *testing.T) {
			scenario := suite.newScenario(t, route)
			scenario.applyAuthorizationChange(t, route)

			resp := scenario.doRoute(t, route, supportTxn("authorization", route.Key), nil, true)
			body := workbookscenariotest.RequireErrorBody(t, resp, route.AuthorizationStatus, route.AuthorizationCode)
			contractassert.RequireAuthorizationReDerived(
				t,
				contractassert.AuthorizationOutcome{Status: route.SuccessStatus},
				contractassert.AuthorizationOutcome{Status: resp.StatusCode, Code: body["error"].(map[string]any)["code"].(string)},
			)
		})
	}
}

func TestSupportPhase4Integration_DefaultQueryMetaAndFieldKeyConformance(t *testing.T) {
	suite := newPhase4SupportSuite(t, "query-matrix")
	for _, route := range workbookscenariotest.RoutesForHarness(
		t,
		workbookscenariotest.WorkbookRouteInventory(workbookscenariotest.RouteInventoryContext{}),
		workbookscenariotest.RouteHarnessQueryFieldMatrix,
	) {
		t.Run(string(route.Key), func(t *testing.T) {
			scenario := suite.newScenario(t, route)
			data := scenario.requireRouteSuccess(t, route, supportTxn("query", route.Key), nil)
			recordID := requireAffectedRecordID(t, route, scenario, data)

			envelope, row := scenario.queryAffectedRow(t, route, recordID)
			contractassert.RequireDefaultQueryMeta(t, envelope, route.ExpectedViewSchemaID)
			contractassert.RequireFieldKeyConformance(
				t,
				workbookscenariotest.SortedRowFieldKeys(t, row),
				workbookscenariotest.AllowedFieldKeys(t, string(route.Key), route.ExpectedViewSchemaID),
			)
		})
	}
}

func TestSupportPhase4Integration_ProjectionAndWebsocketConsequences(t *testing.T) {
	suite := newPhase4SupportSuite(t, "effects")
	for _, route := range workbookscenariotest.RoutesForHarness(
		t,
		workbookscenariotest.WorkbookRouteInventory(workbookscenariotest.RouteInventoryContext{}),
		workbookscenariotest.RouteHarnessEffects,
	) {
		t.Run(string(route.Key), func(t *testing.T) {
			scenario := suite.newScenario(t, route)
			var wsClient *wstest.Client
			if route.WebSocketExpectation == workbookscenariotest.RouteWebSocketRecordChanged {
				wsClient = workbookscenariotest.ConnectViewSocket(
					t,
					scenario.harness.Server,
					scenario.IncidentID.String(),
					route.WebSocketViewSchemaID,
					scenario.actorLogin.SessionCookie.Value,
				)
				defer wsClient.Close(1000, "test_complete")
			}

			data := scenario.requireRouteSuccess(t, route, supportTxn("effects", route.Key), nil)

			if wsClient != nil {
				socketChange := workbookscenariotest.RequireRecordChanged(
					t,
					wsClient,
					route.BuildWebSocketRecordID(scenario.routeCtx),
					route.WebSocketRowVersion,
				)
				requireRouteSocketChange(t, route.Key, data, socketChange, route.WebSocketViewSchemaID, nil)
				for _, expectation := range route.AdditionalWebSocketChanges {
					additionalChange := workbookscenariotest.RequireRecordChanged(
						t,
						wsClient,
						expectation.BuildRecordID(scenario.routeCtx),
						expectation.RowVersion,
					)
					requireRouteSocketChange(t, route.Key, data, additionalChange, expectation.ViewSchemaID, expectation.ChangedKeys)
				}
				workbookscenariotest.ExpectNoSocketMessage(t, wsClient)
			}

			recordID := requireAffectedRecordID(t, route, scenario, data)
			_, rowBefore := scenario.queryAffectedRow(t, route, recordID)
			scenario.rebuildProjection(t, route)
			_, rowAfter := scenario.queryAffectedRow(t, route, recordID)
			contractassert.RequireProjectionDeterminism(t, rowBefore["cells"], rowAfter["cells"])
		})
	}
}

func requireRouteSocketChange(t testing.TB, routeKey workbookscenariotest.RouteKey, responseData map[string]any, socketChange workbookscenariotest.RecordChangeSocketPayload, viewSchemaID string, changedKeys []string) {
	t.Helper()
	if changeSetID, ok := responseData["change_set_id"].(string); ok && changeSetID != "" && socketChange.ChangeSetID != changeSetID {
		t.Fatalf("expected websocket change_set_id to match route response for %s: payload=%#v response=%#v", routeKey, socketChange, responseData)
	}
	if viewSchemaID != "" && !socketChangeIncludesView(socketChange, viewSchemaID) {
		t.Fatalf("expected websocket affected view %s for %s, got %#v", viewSchemaID, routeKey, socketChange)
	}
	for _, key := range changedKeys {
		if !slices.Contains(socketChange.ChangedFieldKeys, key) {
			t.Fatalf("expected websocket changed key %s for %s, got %#v", key, routeKey, socketChange)
		}
	}
}

func socketChangeIncludesView(socketChange workbookscenariotest.RecordChangeSocketPayload, viewSchemaID string) bool {
	for _, view := range socketChange.AffectedViews {
		if view.ViewSchemaID == viewSchemaID {
			return true
		}
	}
	return false
}

func TestSupportPhase4Integration_RecordEnvelopeHeadSchema(t *testing.T) {
	postgresHarness := pgtest.Start(t)
	testDB := postgresHarness.PrepareIsolatedDatabaseT(t, "phase4-records-head")

	db, err := sql.Open("pgx", testDB.DSN)
	if err != nil {
		t.Fatalf("open head schema database: %v", err)
	}
	defer db.Close()

	requireColumns(t, db, "records", "record_id", "incident_id", "record_type", "created_by_user_id", "created_at", "updated_by_user_id", "updated_at", "row_version", "deleted_at", "deleted_by_user_id")
	requireNoColumns(t, db, "record_id", "users", "user_sessions", "bootstrap_tokens", "pending_totp_enrollments", "incident_memberships", "deployment_bootstrap_state")
	requireGenericRecordFKsDoNotTargetTimeline(t, db)
}

type phase4SupportScenario struct {
	harness               *workbookscenariotest.ServerHarness
	bootstrapUserID       uuid.UUID
	actorLogin            workbookscenariotest.LoginResult
	actorUserID           uuid.UUID
	IncidentID            uuid.UUID
	routeCtx              workbookscenariotest.RouteInventoryContext
	label                 string
	routeKey              workbookscenariotest.RouteKey
	timelineID            uuid.UUID
	mentionID             uuid.UUID
	canonicalHostID       uuid.UUID
	duplicateHostID       uuid.UUID
	canonicalIdentityID   uuid.UUID
	duplicateIdentityID   uuid.UUID
	duplicateLinkID       uuid.UUID
	tagIDSurvivor         uuid.UUID
	tagIDLoser            uuid.UUID
	assessmentHostID      uuid.UUID
	partyID               uuid.UUID
	assessmentID          uuid.UUID
	evidenceID            uuid.UUID
	objectBlobID          uuid.UUID
	alternateObjectBlobID uuid.UUID
	noteID                uuid.UUID
	taskRequestID         uuid.UUID
	decisionID            uuid.UUID
	commLogID             uuid.UUID
	handoffID             uuid.UUID
	statusReviewID        uuid.UUID
	lessonID              uuid.UUID
}

type phase4SupportSuite struct {
	label           string
	harness         *workbookscenariotest.ServerHarness
	bootstrapLogin  workbookscenariotest.LoginResult
	bootstrapUserID uuid.UUID
}

func newPhase4SupportSuite(t *testing.T, label string) *phase4SupportSuite {
	t.Helper()

	harness := workbookscenariotest.StartRuntime(t).StartServer(t, "phase4-support-"+label)
	bootstrapLogin, bootstrapUserID := workbookscenariotest.ProvisionBootstrapAdmin(t, harness.Server)
	return &phase4SupportSuite{
		label:           label,
		harness:         harness,
		bootstrapLogin:  bootstrapLogin,
		bootstrapUserID: bootstrapUserID,
	}
}

func (s *phase4SupportSuite) newScenario(t *testing.T, route workbookscenariotest.RouteInventoryEntry) *phase4SupportScenario {
	t.Helper()

	incident := workbookscenariotest.CreateIncident(t, s.harness.Server, s.bootstrapLogin, map[string]any{
		"client_txn_id": supportTxn(s.label+"-incident", route.Key),
		"incident_key":  "IR-P4-" + strings.ToUpper(strings.ReplaceAll(string(route.Key), "_", "-")),
		"title":         "Phase 4 support matrix " + s.label + " " + string(route.Key),
	})
	incidentID := workbookscenariotest.MustUUID(t, incident["incident_id"].(string))

	const actorPassword = "SupportAdminPass1!"
	actorRecord := workbookscenariotest.SeedLocalUserFlags(
		t,
		s.harness.DB,
		"phase4-support-"+s.label+"-"+string(route.Key)+"@example.test",
		"Phase 4 Support Admin",
		actorPassword,
		false,
		false,
		true,
	)
	workbookscenariotest.SeedIncidentMembership(t, s.harness.DB, incidentID, actorRecord.ID, actorRecord.DisplayName, "admin", s.bootstrapUserID)
	actorLogin := loginLocalSupportUser(t, s.harness, actorRecord.Email, actorPassword)

	timelineID := supportUUID(s.label, route.Key, "timeline")
	mentionID := supportUUID(s.label, route.Key, "mention")
	canonicalHostID := supportUUID(s.label, route.Key, "canonical-host")
	duplicateHostID := supportUUID(s.label, route.Key, "duplicate-host")
	canonicalIdentityID := supportUUID(s.label, route.Key, "canonical-identity")
	duplicateIdentityID := supportUUID(s.label, route.Key, "duplicate-identity")

	scenario := &phase4SupportScenario{
		harness:               s.harness,
		bootstrapUserID:       s.bootstrapUserID,
		actorLogin:            actorLogin,
		actorUserID:           actorRecord.ID,
		IncidentID:            incidentID,
		label:                 s.label,
		routeKey:              route.Key,
		timelineID:            timelineID,
		mentionID:             mentionID,
		canonicalHostID:       canonicalHostID,
		duplicateHostID:       duplicateHostID,
		canonicalIdentityID:   canonicalIdentityID,
		duplicateIdentityID:   duplicateIdentityID,
		duplicateLinkID:       supportUUID(s.label, route.Key, "duplicate-link"),
		tagIDSurvivor:         supportUUID(s.label, route.Key, "tag-survivor"),
		tagIDLoser:            supportUUID(s.label, route.Key, "tag-loser"),
		assessmentHostID:      supportUUID(s.label, route.Key, "assessment-host"),
		partyID:               supportUUID(s.label, route.Key, "party"),
		assessmentID:          supportUUID(s.label, route.Key, "assessment"),
		evidenceID:            supportUUID(s.label, route.Key, "evidence"),
		objectBlobID:          supportUUID(s.label, route.Key, "object-blob"),
		alternateObjectBlobID: supportUUID(s.label, route.Key, "alternate-object-blob"),
		noteID:                supportUUID(s.label, route.Key, "note"),
		taskRequestID:         supportUUID(s.label, route.Key, "task-request"),
		decisionID:            supportUUID(s.label, route.Key, "decision"),
		commLogID:             supportUUID(s.label, route.Key, "comm-log"),
		handoffID:             supportUUID(s.label, route.Key, "handoff"),
		statusReviewID:        supportUUID(s.label, route.Key, "status-review"),
		lessonID:              supportUUID(s.label, route.Key, "lesson"),
		routeCtx: workbookscenariotest.RouteInventoryContext{
			IncidentID:            incidentID.String(),
			ActorUserID:           actorRecord.ID.String(),
			TimelineRecordID:      timelineID.String(),
			MentionID:             mentionID.String(),
			MergeSurvivorRecordID: canonicalHostID.String(),
			MergeLoserRecordID:    duplicateHostID.String(),
			HostRecordID:          canonicalHostID.String(),
			IdentityRecordID:      canonicalIdentityID.String(),
			PartyRecordID:         supportUUID(s.label, route.Key, "party").String(),
			AssessmentRecordID:    supportUUID(s.label, route.Key, "assessment").String(),
			EvidenceRecordID:      supportUUID(s.label, route.Key, "evidence").String(),
			ObjectBlobID:          supportUUID(s.label, route.Key, "object-blob").String(),
			AlternateObjectBlobID: supportUUID(s.label, route.Key, "alternate-object-blob").String(),
			NoteRecordID:          supportUUID(s.label, route.Key, "note").String(),
			TaskRequestRecordID:   supportUUID(s.label, route.Key, "task-request").String(),
			DecisionRecordID:      supportUUID(s.label, route.Key, "decision").String(),
			CommLogRecordID:       supportUUID(s.label, route.Key, "comm-log").String(),
			HandoffRecordID:       supportUUID(s.label, route.Key, "handoff").String(),
			StatusReviewRecordID:  supportUUID(s.label, route.Key, "status-review").String(),
			LessonRecordID:        supportUUID(s.label, route.Key, "lesson").String(),
		},
	}

	scenario.seedBaseData(t, route)
	return scenario
}

func (s *phase4SupportScenario) seedBaseData(t *testing.T, route workbookscenariotest.RouteInventoryEntry) {
	t.Helper()

	workbookscenariotest.SeedTimelineRecord(t, s.harness.DB, s.IncidentID, s.actorUserID, s.timelineID)
	workbookscenariotest.SeedHostRecord(t, s.harness.DB, s.IncidentID, s.actorUserID, s.canonicalHostID, "WS-023", "WS-023", "", "")
	workbookscenariotest.SeedHostRecord(t, s.harness.DB, s.IncidentID, s.actorUserID, s.duplicateHostID, "WS-024", "WS-024", "ws-024.corp.example.test", "")
	workbookscenariotest.SeedIdentityRecord(t, s.harness.DB, s.IncidentID, s.actorUserID, s.canonicalIdentityID, "Alex Analyst", "alex.analyst@example.test", "alex.analyst@example.test", "ALEXA")
	workbookscenariotest.SeedIdentityRecord(t, s.harness.DB, s.IncidentID, s.actorUserID, s.duplicateIdentityID, "Legacy Analyst", "legacy.analyst@example.test", "legacy.analyst@example.test", "LEGACYA")
	s.seedWorkbookRouteFamilyData(t, route)

	switch route.Key {
	case workbookscenariotest.RouteMentionResolve:
		workbookscenariotest.SeedMention(
			t,
			s.harness.DB,
			s.actorUserID,
			s.mentionID,
			s.timelineID,
			golden.RecordFieldTimelineHostRefs,
			"host",
			"WS-023",
			"unresolved",
			nil,
			nil,
		)
	case workbookscenariotest.RouteExplicitMerge:
		workbookscenariotest.SeedResolvedMention(
			t,
			s.harness.DB,
			s.actorUserID,
			s.mentionID,
			s.timelineID,
			s.duplicateHostID,
			golden.RecordFieldTimelineHostRefs,
			"host",
			"WS-024",
		)
		workbookscenariotest.SeedRecordLink(
			t,
			s.harness.DB,
			s.IncidentID,
			s.actorUserID,
			s.duplicateLinkID,
			s.timelineID,
			s.duplicateHostID,
			"observed_on_host",
			"manual",
			nil,
		)
		workbookscenariotest.SeedRecordTag(t, s.harness.DB, s.IncidentID, s.actorUserID, s.tagIDSurvivor, s.canonicalHostID, "critical-host")
		workbookscenariotest.SeedRecordTag(t, s.harness.DB, s.IncidentID, s.actorUserID, s.tagIDLoser, s.duplicateHostID, "critical-host")
		workbookscenariotest.SeedAssessment(t, s.harness.DB, s.IncidentID, s.actorUserID, s.assessmentHostID, s.duplicateHostID, "host", "confirmed")
	}

	s.rebuildBaseProjections(t)

	if route.Key == workbookscenariotest.RouteIndicatorsQuery {
		data := s.requireRouteSuccess(
			t,
			findRouteByKey(t, workbookscenariotest.RouteIndicatorsCreate),
			supportTxn(s.label+"-seed-indicator", route.Key),
			nil,
		)
		s.routeCtx.IndicatorRecordID = requireRowRecordID(t, data)
	}
}

func (s *phase4SupportScenario) seedWorkbookRouteFamilyData(t *testing.T, route workbookscenariotest.RouteInventoryEntry) {
	t.Helper()

	s.routeCtx.PartyRecordID = s.seedWorkbookCreate(t, workbookscenariotest.WorkbookPartiesViewSchemaID, map[string]any{
		"client_txn_id":      supportTxn(s.label+"-seed-party", s.routeKey),
		"party.display_name": "Seed Support Party",
		"party.party_kind":   "organization",
	})
	s.routeCtx.EvidenceRecordID = s.seedWorkbookCreate(t, workbookscenariotest.WorkbookEvidenceViewSchemaID, map[string]any{
		"client_txn_id":  supportTxn(s.label+"-seed-evidence", s.routeKey),
		"evidence.title": "Seed Support Evidence",
	})
	s.routeCtx.NoteRecordID = s.seedWorkbookCreate(t, workbookscenariotest.WorkbookNotesViewSchemaID, map[string]any{
		"client_txn_id": supportTxn(s.label+"-seed-note", s.routeKey),
		"note.title":    "Seed Support Note",
	})
	s.routeCtx.DecisionRecordID = s.seedWorkbookCreate(t, workbookscenariotest.WorkbookDecisionsViewSchemaID, map[string]any{
		"client_txn_id":          supportTxn(s.label+"-seed-decision", s.routeKey),
		"decision.summary":       "Seed Support Decision",
		"decision.decision_type": "containment",
		"decision.rationale":     "Seed support rationale",
	})
	s.routeCtx.TaskRequestRecordID = s.seedWorkbookCreate(t, workbookscenariotest.WorkbookTaskRequestsViewSchemaID, map[string]any{
		"client_txn_id":  supportTxn(s.label+"-seed-task", s.routeKey),
		"task.title":     "Seed Support Task",
		"task.task_kind": "collection",
	})
	s.routeCtx.AssessmentRecordID = s.seedWorkbookCreate(t, workbookscenariotest.WorkbookAssessmentsViewSchemaID, map[string]any{
		"client_txn_id":               supportTxn(s.label+"-seed-assessment", s.routeKey),
		"assessment.subject_ref":      s.routeCtx.HostRecordID,
		"assessment.subject_type":     "host",
		"assessment.assessment_state": "confirmed",
		"assessment.confidence_score": 55,
		"assessment.rationale":        "Seed support assessment",
	})
	s.routeCtx.CommLogRecordID = s.seedWorkbookCreate(t, workbookscenariotest.WorkbookCommLogViewSchemaID, map[string]any{
		"client_txn_id":               supportTxn(s.label+"-seed-comm-log", s.routeKey),
		"comm_log.comm_type":          "briefing",
		"comm_log.audience":           "leadership",
		"comm_log.channel_or_meeting": "Bridge",
		"comm_log.summary":            "Seed support communication",
	})
	s.routeCtx.HandoffRecordID = s.seedWorkbookCreate(t, workbookscenariotest.WorkbookHandoffViewSchemaID, map[string]any{
		"client_txn_id":                  supportTxn(s.label+"-seed-handoff", s.routeKey),
		"handoff.incoming_owner_user_id": s.actorUserID.String(),
		"handoff.current_state_summary":  "Seed support handoff",
	})
	s.routeCtx.StatusReviewRecordID = s.seedWorkbookCreate(t, workbookscenariotest.WorkbookStatusReviewViewSchemaID, map[string]any{
		"client_txn_id":                       supportTxn(s.label+"-seed-status-review", s.routeKey),
		"status_review.current_state_summary": "Seed support status",
	})
	s.routeCtx.LessonRecordID = s.seedWorkbookCreate(t, workbookscenariotest.WorkbookLessonViewSchemaID, map[string]any{
		"client_txn_id":  supportTxn(s.label+"-seed-lesson", s.routeKey),
		"lesson.summary": "Seed support lesson",
	})

	if route.Key == workbookscenariotest.RouteEvidenceAttachBlob {
		s.routeCtx.ObjectBlobID = s.seedUploadedObjectBlob(t, "attach-primary")
		s.routeCtx.AlternateObjectBlobID = s.seedUploadedObjectBlob(t, "attach-alternate")
	}
	if route.Key == workbookscenariotest.RouteEvidencePreviewHandle || route.Key == workbookscenariotest.RouteEvidenceDownloadHandle {
		objectBlobID := s.seedUploadedObjectBlob(t, "handle")
		s.attachSeededBlob(t, objectBlobID)
		s.routeCtx.ObjectBlobID = objectBlobID
	}
}

func (s *phase4SupportScenario) seedWorkbookCreate(t *testing.T, viewSchemaID string, body map[string]any) string {
	t.Helper()

	resp := workbookscenariotest.DoJSON(
		t,
		http.MethodPost,
		s.harness.Server.HTTP.URL+"/api/v1/incidents/"+s.IncidentID.String()+"/views/"+viewSchemaID+"/rows",
		body,
		workbookscenariotest.WithCookies(s.actorLogin.SessionCookie, s.actorLogin.CSRFCookie),
		workbookscenariotest.WithHeader(authn.CSRFHeaderName, s.actorLogin.CSRFCookie.Value),
	)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("seed workbook create %s failed: status=%d body=%#v request=%#v", viewSchemaID, resp.StatusCode, httptestx.ReadJSONBody(t, resp), body)
	}
	data := httptestx.RequireSuccessEnvelope(t, resp, http.StatusCreated)["data"].(map[string]any)
	return requireRowRecordID(t, data)
}

func (s *phase4SupportScenario) seedUploadedObjectBlob(t *testing.T, label string) string {
	t.Helper()

	payload := []byte("phase4 support object " + label)
	sum := sha256.Sum256(payload)
	resp := workbookscenariotest.DoJSON(
		t,
		http.MethodPost,
		s.harness.Server.HTTP.URL+"/api/v1/object-blobs",
		map[string]any{
			"incident_id":       s.IncidentID.String(),
			"client_txn_id":     supportTxn(s.label+"-seed-blob-"+label, s.routeKey),
			"byte_size":         len(payload),
			"filename_hint":     label + ".txt",
			"content_type_hint": "text/plain",
			"sha256_hex":        fmt.Sprintf("%x", sum[:]),
		},
		workbookscenariotest.WithCookies(s.actorLogin.SessionCookie, s.actorLogin.CSRFCookie),
		workbookscenariotest.WithHeader(authn.CSRFHeaderName, s.actorLogin.CSRFCookie.Value),
	)
	data := httptestx.RequireSuccessEnvelope(t, resp, http.StatusCreated)["data"].(map[string]any)
	putResp, err := http.DefaultClient.Do(mustPutRequest(t, s.harness.Server.HTTP.URL, data["upload_target"].(map[string]any)["href"].(string), payload))
	if err != nil {
		t.Fatalf("upload support object %s: %v", label, err)
	}
	defer putResp.Body.Close()
	if putResp.StatusCode < 200 || putResp.StatusCode >= 300 {
		body, _ := io.ReadAll(putResp.Body)
		t.Fatalf("upload support object %s status %d: %s", label, putResp.StatusCode, string(body))
	}
	return data["object_blob_id"].(string)
}

func (s *phase4SupportScenario) attachSeededBlob(t *testing.T, objectBlobID string) {
	t.Helper()

	resp := workbookscenariotest.DoJSON(
		t,
		http.MethodPost,
		s.harness.Server.HTTP.URL+"/api/v1/evidence-records/"+s.routeCtx.EvidenceRecordID+"/attach-blob",
		map[string]any{
			"object_blob_id":   objectBlobID,
			"base_row_version": 1,
			"client_txn_id":    supportTxn(s.label+"-seed-attach", s.routeKey),
		},
		workbookscenariotest.WithCookies(s.actorLogin.SessionCookie, s.actorLogin.CSRFCookie),
		workbookscenariotest.WithHeader(authn.CSRFHeaderName, s.actorLogin.CSRFCookie.Value),
	)
	httptestx.RequireSuccessEnvelope(t, resp, http.StatusOK)
}

func mustPutRequest(t testing.TB, baseURL string, url string, payload []byte) *http.Request {
	t.Helper()

	if strings.HasPrefix(url, "/") {
		url = baseURL + url
	}
	req, err := http.NewRequest(http.MethodPut, url, bytes.NewReader(payload))
	if err != nil {
		t.Fatalf("create object upload request: %v", err)
	}
	req.Header.Set("Content-Type", "text/plain")
	return req
}

func (s *phase4SupportScenario) rebuildBaseProjections(t *testing.T) {
	t.Helper()

	store := projections.NewStore(s.harness.Server.Runtime.Postgres)
	if err := store.RebuildIncidentTimeline(context.Background(), s.IncidentID); err != nil {
		t.Fatalf("rebuild timeline projections: %v", err)
	}
	if err := store.RebuildIncidentHosts(context.Background(), s.IncidentID); err != nil {
		t.Fatalf("rebuild host projections: %v", err)
	}
	if err := store.RebuildIncidentIdentities(context.Background(), s.IncidentID); err != nil {
		t.Fatalf("rebuild identity projections: %v", err)
	}
}

func (s *phase4SupportScenario) requireRouteSuccess(t *testing.T, route workbookscenariotest.RouteInventoryEntry, clientTxnID string, body any) map[string]any {
	t.Helper()

	resp := s.doRoute(t, route, clientTxnID, body, true)
	return s.requireRouteSuccessStatus(t, route, resp, route.SuccessStatus)
}

func (s *phase4SupportScenario) requireRouteSuccessStatus(t *testing.T, route workbookscenariotest.RouteInventoryEntry, resp *http.Response, wantStatus int) map[string]any {
	t.Helper()

	if resp.StatusCode != wantStatus {
		t.Fatalf("route %s unexpected status: got %d want %d body=%#v", route.Key, resp.StatusCode, wantStatus, httptestx.ReadJSONBody(t, resp))
	}
	data := workbookscenariotest.RequireSuccessData(t, resp, wantStatus)
	assertRouteSuccessShape(t, route, s, data)
	return data
}

func (s *phase4SupportScenario) doRoute(
	t *testing.T,
	route workbookscenariotest.RouteInventoryEntry,
	clientTxnID string,
	body any,
	includeCSRFFHeader bool,
) *http.Response {
	t.Helper()

	requestBody := body
	if requestBody == nil {
		requestBody = route.BuildBody(s.routeCtx, clientTxnID)
	}

	options := []func(*http.Request){}
	if route.RequiresCSRF {
		options = append(options, workbookscenariotest.WithCookies(s.actorLogin.SessionCookie, s.actorLogin.CSRFCookie))
		if includeCSRFFHeader {
			options = append(options, workbookscenariotest.WithHeader(authn.CSRFHeaderName, s.actorLogin.CSRFCookie.Value))
		}
	} else {
		options = append(options, workbookscenariotest.WithCookies(s.actorLogin.SessionCookie))
	}

	return workbookscenariotest.DoJSON(
		t,
		route.Method,
		s.harness.Server.HTTP.URL+route.BuildPath(s.routeCtx),
		requestBody,
		options...,
	)
}

func (s *phase4SupportScenario) applyAuthorizationChange(t *testing.T, route workbookscenariotest.RouteInventoryEntry) {
	t.Helper()

	switch route.AuthorizationChange {
	case workbookscenariotest.RouteAuthorizationDemoteViewer:
		if _, err := s.harness.DB.ExecContext(
			context.Background(),
			`
UPDATE incident_memberships
   SET role = 'viewer',
       updated_at = now(),
       updated_by_user_id = $3
 WHERE incident_id = $1
   AND user_id = $2
`,
			s.IncidentID,
			s.actorUserID,
			s.bootstrapUserID,
		); err != nil {
			t.Fatalf("demote support actor membership: %v", err)
		}
	case workbookscenariotest.RouteAuthorizationRemoveMember:
		if _, err := s.harness.DB.ExecContext(
			context.Background(),
			`DELETE FROM incident_memberships WHERE incident_id = $1 AND user_id = $2`,
			s.IncidentID,
			s.actorUserID,
		); err != nil {
			t.Fatalf("remove support actor membership: %v", err)
		}
	case workbookscenariotest.RouteAuthorizationNotApplicable:
		t.Fatalf("route %s does not declare authorization change", route.Key)
	default:
		t.Fatalf("unsupported authorization change %s", route.AuthorizationChange)
	}
}

func (s *phase4SupportScenario) queryAffectedRow(t *testing.T, route workbookscenariotest.RouteInventoryEntry, recordID string) (map[string]any, map[string]any) {
	t.Helper()

	envelope := workbookscenariotest.QueryViewEnvelope(
		t,
		s.harness.Server.HTTP.URL,
		s.IncidentID.String(),
		route.ExpectedViewSchemaID,
		s.actorLogin,
	)
	rows := rowsFromQueryData(t, envelope["data"].(map[string]any))
	return envelope, workbookscenariotest.FindRow(t, rows, recordID)
}

func (s *phase4SupportScenario) rebuildProjection(t *testing.T, route workbookscenariotest.RouteInventoryEntry) {
	t.Helper()

	store := projections.NewStore(s.harness.Server.Runtime.Postgres)
	var err error
	switch route.ProjectionTarget {
	case workbookscenariotest.RouteProjectionHosts:
		err = store.RebuildIncidentHosts(context.Background(), s.IncidentID)
	case workbookscenariotest.RouteProjectionIdentities:
		err = store.RebuildIncidentIdentities(context.Background(), s.IncidentID)
	case workbookscenariotest.RouteProjectionIndicators:
		err = store.RebuildIncidentIndicators(context.Background(), s.IncidentID)
	case workbookscenariotest.RouteProjectionTimeline:
		err = store.RebuildIncidentTimeline(context.Background(), s.IncidentID)
	case workbookscenariotest.RouteProjectionNotApplicable:
		return
	default:
		t.Fatalf("unsupported projection target %s", route.ProjectionTarget)
	}
	if err != nil {
		t.Fatalf("rebuild %s projection: %v", route.ProjectionTarget, err)
	}
}

func assertRouteSuccessShape(t testing.TB, route workbookscenariotest.RouteInventoryEntry, scenario *phase4SupportScenario, data map[string]any) {
	t.Helper()

	switch route.SuccessShape {
	case workbookscenariotest.RouteSuccessShapeMentionResolution:
		if data["incident_id"] != scenario.IncidentID.String() {
			t.Fatalf("unexpected incident_id for %s: %#v", route.Key, data)
		}
		if _, ok := data["source_record"].(map[string]any); !ok {
			t.Fatalf("expected mention resolution source_record for %s, got %#v", route.Key, data)
		}
		requireNonEmptyString(t, data, "change_set_id")
	case workbookscenariotest.RouteSuccessShapeMerge:
		if data["survivor_record_id"] != scenario.routeCtx.MergeSurvivorRecordID {
			t.Fatalf("unexpected survivor_record_id for %s: %#v", route.Key, data)
		}
		if data["loser_record_id"] != scenario.routeCtx.MergeLoserRecordID {
			t.Fatalf("unexpected loser_record_id for %s: %#v", route.Key, data)
		}
		if _, ok := data["merge_summary"].(map[string]any); !ok {
			t.Fatalf("expected merge_summary for %s, got %#v", route.Key, data)
		}
		requireNonEmptyString(t, data, "change_set_id")
	case workbookscenariotest.RouteSuccessShapeMutationRow:
		requireNonEmptyString(t, data, "change_set_id")
		if requireRowRecordID(t, data) == "" {
			t.Fatalf("expected row record_id for %s, got %#v", route.Key, data)
		}
	case workbookscenariotest.RouteSuccessShapeQueryRows:
		rows := rowsFromQueryData(t, data)
		if len(rows) == 0 {
			t.Fatalf("expected non-empty query rows for %s, got %#v", route.Key, data)
		}
	case workbookscenariotest.RouteSuccessShapeObjectBlob:
		requireNonEmptyString(t, data, "object_blob_id")
		if data["incident_id"] != scenario.IncidentID.String() {
			t.Fatalf("unexpected blob incident_id for %s: %#v", route.Key, data)
		}
		if data["upload_state"] != "pending" {
			t.Fatalf("expected pending object blob for %s, got %#v", route.Key, data)
		}
		if _, ok := data["upload_target"].(map[string]any); !ok {
			t.Fatalf("expected upload_target for %s, got %#v", route.Key, data)
		}
		if _, ok := data["accepted_contract"].(map[string]any); !ok {
			t.Fatalf("expected accepted_contract for %s, got %#v", route.Key, data)
		}
	case workbookscenariotest.RouteSuccessShapeEvidenceAttach:
		requireNonEmptyString(t, data, "change_set_id")
		requireNonEmptyString(t, data, "object_blob_id")
		if requireRowRecordID(t, data) != scenario.routeCtx.EvidenceRecordID {
			t.Fatalf("expected attach row for %s to match evidence record, got %#v", route.Key, data)
		}
	case workbookscenariotest.RouteSuccessShapeEvidenceHandle:
		requireNonEmptyString(t, data, "href")
		if data["record_id"] != scenario.routeCtx.EvidenceRecordID {
			t.Fatalf("expected handle record_id for %s, got %#v", route.Key, data)
		}
		if data["handle_kind"] == "" {
			t.Fatalf("expected handle_kind for %s, got %#v", route.Key, data)
		}
	default:
		t.Fatalf("unsupported success shape %s", route.SuccessShape)
	}
}

func requireStableReplayPayload(t testing.TB, route workbookscenariotest.RouteInventoryEntry, firstData map[string]any, replayData map[string]any) {
	t.Helper()

	firstChangeSet, firstHasChangeSet := firstData["change_set_id"].(string)
	replayChangeSet, replayHasChangeSet := replayData["change_set_id"].(string)
	if firstHasChangeSet && replayHasChangeSet && firstChangeSet != replayChangeSet {
		t.Fatalf("expected replay to preserve change_set_id for %s: first=%#v replay=%#v", route.Key, firstData, replayData)
	}
	if route.SuccessShape == workbookscenariotest.RouteSuccessShapeMutationRow {
		if requireRowRecordID(t, firstData) != requireRowRecordID(t, replayData) {
			t.Fatalf("expected replay to preserve row record_id for %s: first=%#v replay=%#v", route.Key, firstData, replayData)
		}
	}
	if route.SuccessShape == workbookscenariotest.RouteSuccessShapeObjectBlob {
		if firstData["object_blob_id"] != replayData["object_blob_id"] {
			t.Fatalf("expected replay to preserve object_blob_id for %s: first=%#v replay=%#v", route.Key, firstData, replayData)
		}
	}
}

func requireColumns(t testing.TB, db *sql.DB, table string, columns ...string) {
	t.Helper()
	for _, column := range columns {
		assertCount(t, db, `
SELECT COUNT(*)
  FROM information_schema.columns
 WHERE table_schema = 'public'
   AND table_name = $1
   AND column_name = $2
`, 1, table, column)
	}
}

func requireNoColumns(t testing.TB, db *sql.DB, column string, tables ...string) {
	t.Helper()
	for _, table := range tables {
		assertCount(t, db, `
SELECT COUNT(*)
  FROM information_schema.columns
 WHERE table_schema = 'public'
   AND table_name = $1
   AND column_name = $2
`, 0, table, column)
	}
}

func requireGenericRecordFKsDoNotTargetTimeline(t testing.TB, db *sql.DB) {
	t.Helper()
	assertCount(t, db, `
SELECT COUNT(*)
  FROM pg_constraint c
  JOIN pg_class source_table ON source_table.oid = c.conrelid
  JOIN pg_class target_table ON target_table.oid = c.confrelid
 WHERE c.contype = 'f'
   AND source_table.relname IN ('record_revisions', 'record_links', 'record_tags', 'entity_mentions', 'indicator_observations')
   AND target_table.relname = 'timeline_events'
`, 0)
}

func assertCount(t testing.TB, db *sql.DB, query string, want int, args ...any) {
	t.Helper()
	var got int
	if err := db.QueryRowContext(context.Background(), query, args...).Scan(&got); err != nil {
		t.Fatalf("query count failed: %v\n%s", err, query)
	}
	if got != want {
		t.Fatalf("unexpected count: got %d want %d\n%s", got, want, query)
	}
}

func requireAffectedRecordID(t testing.TB, route workbookscenariotest.RouteInventoryEntry, scenario *phase4SupportScenario, data map[string]any) string {
	t.Helper()

	recordID := route.AffectedRecordID(scenario.routeCtx, data)
	if recordID == "" {
		t.Fatalf("route %s returned no affected record id for %#v", route.Key, data)
	}
	return recordID
}

func rowsFromQueryData(t testing.TB, data map[string]any) []map[string]any {
	t.Helper()

	rawRows, ok := data["rows"].([]any)
	if !ok {
		t.Fatalf("expected query rows array, got %#v", data)
	}
	rows := make([]map[string]any, 0, len(rawRows))
	for _, rawRow := range rawRows {
		row, ok := rawRow.(map[string]any)
		if !ok {
			t.Fatalf("expected query row object, got %#v", rawRow)
		}
		rows = append(rows, row)
	}
	return rows
}

func requireRowRecordID(t testing.TB, data map[string]any) string {
	t.Helper()

	row, ok := data["row"].(map[string]any)
	if !ok {
		t.Fatalf("expected row payload, got %#v", data)
	}
	recordID, ok := row["record_id"].(string)
	if !ok || recordID == "" {
		t.Fatalf("expected non-empty row.record_id, got %#v", row)
	}
	return recordID
}

func requireNonEmptyString(t testing.TB, payload map[string]any, key string) string {
	t.Helper()

	value, ok := payload[key].(string)
	if !ok || value == "" {
		t.Fatalf("expected non-empty %s in %#v", key, payload)
	}
	return value
}

func loginLocalSupportUser(t testing.TB, harness *workbookscenariotest.ServerHarness, username string, password string) workbookscenariotest.LoginResult {
	t.Helper()

	resp := workbookscenariotest.DoJSON(
		t,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/auth/login",
		map[string]any{
			"username": username,
			"password": password,
		},
	)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("support login failed: status=%d body=%#v", resp.StatusCode, httptestx.ReadJSONBody(t, resp))
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
		t.Fatalf("expected support login to set session and csrf cookies, got %#v", resp.Cookies())
	}
	return workbookscenariotest.LoginResult{
		SessionCookie: sessionCookie,
		CSRFCookie:    csrfCookie,
	}
}

func findRouteByKey(t testing.TB, key workbookscenariotest.RouteKey) workbookscenariotest.RouteInventoryEntry {
	t.Helper()

	for _, route := range workbookscenariotest.WorkbookRouteInventory(workbookscenariotest.RouteInventoryContext{}) {
		if route.Key == key {
			return route
		}
	}
	t.Fatalf("missing phase4 route inventory entry %s", key)
	return workbookscenariotest.RouteInventoryEntry{}
}

func supportTxn(label string, routeKey workbookscenariotest.RouteKey) string {
	return "txn-phase4-support-" + label + "-" + string(routeKey)
}

func supportUUID(label string, routeKey workbookscenariotest.RouteKey, name string) uuid.UUID {
	return uuid.NewSHA1(uuid.NameSpaceOID, []byte("cartulary-phase4-support:"+label+":"+string(routeKey)+":"+name))
}
