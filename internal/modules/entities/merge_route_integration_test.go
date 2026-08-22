package entities_test

import (
	"context"
	"net/http"
	"reflect"
	"testing"

	"github.com/google/uuid"

	assessmenttest "github.com/JochiRaider/cartulary/internal/modules/assessments/testsupport"
	authflowtest "github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	"github.com/JochiRaider/cartulary/internal/modules/collaboration/testsupport/incidentwstest"
	entitytest "github.com/JochiRaider/cartulary/internal/modules/entities/testsupport"
	linktest "github.com/JochiRaider/cartulary/internal/modules/links/testsupport"
	timelinetest "github.com/JochiRaider/cartulary/internal/modules/timeline/testsupport"
	"github.com/JochiRaider/cartulary/internal/modules/timeline/testsupport/asserttest"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	viewtest "github.com/JochiRaider/cartulary/internal/platform/viewschema/testsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/auditassert"
	"github.com/JochiRaider/cartulary/internal/testutil/contractassert"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/JochiRaider/cartulary/internal/testutil/revisionsupport"
	workbookscenariotest "github.com/JochiRaider/cartulary/internal/testutil/workbookscenariotest"
)

func TestExplicitMergeRoute_Integration(t *testing.T) {
	t.Run("entity route failures enforce authentication csrf visibility role and body precedence", func(t *testing.T) {
		harness := appsupport.StartServer(t, "entity_linking-entity-route-failure-precedence")
		adminLogin, adminUserID := appsupport.ProvisionBootstrapAdmin(t, harness.Server)
		incident := appsupport.CreateIncident(t, harness.Server, adminLogin, map[string]any{
			"client_txn_id": "txn-entity_linking-route-precedence-incident",
			"incident_key":  "IR-ROUTE-PRECEDENCE",
			"title":         "Entity route failure precedence",
		})
		incidentID := appsupport.MustUUID(t, incident["incident_id"].(string))
		entitytest.SeedHostRecord(t, harness.DB, incidentID, adminUserID, entitytest.CanonicalHostRecordID, "Survivor", "SURVIVOR", "", "")
		entitytest.SeedHostRecord(t, harness.DB, incidentID, adminUserID, entitytest.DuplicateHostRecordID, "Loser", "LOSER", "", "")
		timelinetest.SeedTimelineRecord(t, harness.DB, incidentID, adminUserID, timelinetest.RecordID)
		entitytest.SeedMention(t, harness.DB, adminUserID, entitytest.HostMentionID, timelinetest.RecordID, timelinetest.FieldHostRefs, "host", "SURVIVOR", "unresolved", nil, nil)

		type failureCounts struct {
			changeSets      int
			mutations       int
			idempotencyRows int
			hostProjection  int
		}
		counts := func() failureCounts {
			return failureCounts{
				changeSets:      appsupport.QueryCount(t, harness.DB, `SELECT COUNT(*) FROM change_sets WHERE incident_id = $1`, incidentID),
				mutations:       appsupport.QueryCount(t, harness.DB, `SELECT COUNT(*) FROM change_set_mutations m JOIN change_sets c ON c.change_set_id = m.change_set_id WHERE c.incident_id = $1`, incidentID),
				idempotencyRows: appsupport.QueryCount(t, harness.DB, `SELECT COUNT(*) FROM route_idempotency WHERE scope_key IN ($1, $2, $3)`, entitytest.CanonicalHostRecordID.String(), entitytest.DuplicateHostRecordID.String(), entitytest.HostMentionID.String()),
				hostProjection:  appsupport.QueryCount(t, harness.DB, `SELECT COUNT(*) FROM host_grid_projection WHERE incident_id = $1`, incidentID),
			}
		}
		before := counts()
		socket := incidentwstest.ConnectViewSocket(t, harness.Server, incidentID.String(), viewtest.HostsViewSchemaID, adminLogin.SessionCookie.Value)
		defer socket.Close(1000, "test_complete")

		mergeMalformedPath := harness.Server.HTTP.URL + "/api/v1/records/not-a-uuid/merge"
		mentionMalformedPath := harness.Server.HTTP.URL + "/api/v1/entity-mentions/not-a-uuid/resolve"
		for _, url := range []string{mergeMalformedPath, mentionMalformedPath} {
			appsupport.RequireErrorBody(t, doEntitiesRawJSON(t, http.MethodPost, url, "{"), http.StatusUnauthorized, "session_required")
			appsupport.RequireErrorBody(t, doEntitiesRawJSON(t, http.MethodPost, url, "{", withCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie)), http.StatusForbidden, "csrf_verification_failed")
		}
		appsupport.RequireErrorBody(t, doEntitiesRawJSON(t, http.MethodPost, mergeMalformedPath, "{", withCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie), withHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value)), http.StatusNotFound, "incident_not_found")
		appsupport.RequireErrorBody(t, doEntitiesRawJSON(t, http.MethodPost, mentionMalformedPath, "{", withCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie), withHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value)), http.StatusNotFound, "entity_mention_not_found")

		hiddenRecordID := uuid.New()
		hiddenMentionID := uuid.New()
		appsupport.RequireErrorBody(t, doEntitiesRawJSON(t, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/records/"+hiddenRecordID.String()+"/merge", "{", withCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie), withHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value)), http.StatusNotFound, "incident_not_found")
		appsupport.RequireErrorBody(t, doEntitiesRawJSON(t, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/entity-mentions/"+hiddenMentionID.String()+"/resolve", "{", withCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie), withHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value)), http.StatusNotFound, "entity_mention_not_found")

		if _, err := harness.DB.ExecContext(context.Background(), `UPDATE incident_memberships SET role = 'viewer', updated_at = now(), updated_by_user_id = $3 WHERE incident_id = $1 AND user_id = $2`, incidentID, adminUserID, adminUserID); err != nil {
			t.Fatalf("demote entity route actor: %v", err)
		}
		mergeURL := harness.Server.HTTP.URL + "/api/v1/records/" + entitytest.CanonicalHostRecordID.String() + "/merge"
		mentionURL := harness.Server.HTTP.URL + "/api/v1/entity-mentions/" + entitytest.HostMentionID.String() + "/resolve"
		for _, url := range []string{mergeURL, mentionURL} {
			appsupport.RequireErrorBody(t, doEntitiesRawJSON(t, http.MethodPost, url, "{", withCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie), withHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value)), http.StatusForbidden, "authorization_denied")
		}
		if _, err := harness.DB.ExecContext(context.Background(), `UPDATE incident_memberships SET role = 'admin', updated_at = now(), updated_by_user_id = $3 WHERE incident_id = $1 AND user_id = $2`, incidentID, adminUserID, adminUserID); err != nil {
			t.Fatalf("restore entity route actor: %v", err)
		}
		for _, url := range []string{mergeURL, mentionURL} {
			appsupport.RequireErrorBody(t, doEntitiesRawJSON(t, http.MethodPost, url, "{", withCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie), withHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value)), http.StatusBadRequest, "invalid_mutation_payload")
		}

		otherActor := authflowtest.SeedLocalUserRecord(t, harness.DB, "entity_linking-route-precedence-other@example.test", "EntityLinking Route Other", "EntityLinkingRouteOtherPass1!", false, false, true)
		otherIncident := appsupport.CreateIncidentInStore(t, harness.Pool, otherActor, "txn-entity_linking-route-precedence-hidden-incident", "IR-ROUTE-PRECEDENCE-HIDDEN", "Entity route hidden incident")
		hiddenLoserID := uuid.New()
		entitytest.SeedHostRecord(t, harness.DB, otherIncident.ID, otherActor.ID, hiddenLoserID, "Hidden loser", "HIDDEN-LOSER", "", "")

		for _, loserID := range []string{"not-a-uuid", uuid.New().String(), hiddenLoserID.String()} {
			resp := doEntitiesJSON(t, http.MethodPost, mergeURL, map[string]any{
				"loser_record_id":           loserID,
				"survivor_base_row_version": 1,
				"loser_base_row_version":    1,
				"client_txn_id":             "txn-entity_linking-route-precedence-" + loserID,
			}, withCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie), withHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value))
			appsupport.RequireErrorBody(t, resp, http.StatusNotFound, "incident_not_found")
		}

		if after := counts(); after != before {
			t.Fatalf("entity route failures mutated durable state: before=%#v after=%#v", before, after)
		}
		incidentwstest.ExpectNoSocketMessage(t, socket)
	})

	t.Run("host merge repoints live fan-out and preserves survivor reuse", func(t *testing.T) {
		harness := appsupport.StartServer(t, "entity_linking-i-4-03")
		appsupport.RequireSchemaTables(t, harness.DB, "entity-resolution", "hosts", "identities", "entity_mentions", "record_tags", "assessments")

		adminLogin, adminUserID := provisionBootstrapAdmin(t, harness.Server)
		incident := createIncident(t, harness.Server, adminLogin, map[string]any{
			"client_txn_id": "txn-entity_linking-i-4-03-incident",
			"incident_key":  "IR-I403",
			"title":         "Entity merge",
		})
		incidentID := mustUUID(t, incident["incident_id"].(string))
		timelineSocket := incidentwstest.ConnectViewSocket(t, harness.Server, incidentID.String(), viewtest.TimelineViewSchemaID, adminLogin.sessionCookie.Value)
		defer timelineSocket.Close(1000, "test_complete")
		hostSocket := incidentwstest.ConnectViewSocket(t, harness.Server, incidentID.String(), viewtest.HostsViewSchemaID, adminLogin.sessionCookie.Value)
		defer hostSocket.Close(1000, "test_complete")

		seedHostRecord(t, harness.DB, incidentID, adminUserID, entitytest.CanonicalHostRecordID, "WS-023", "WS-023", "", "")
		seedHostRecord(t, harness.DB, incidentID, adminUserID, entitytest.DuplicateHostRecordID, "WS-023 duplicate", "WS-023-DUP", "ws-023.corp.example.test", "")
		seedEntityAlias(t, harness.DB, incidentID, adminUserID, entitytest.DuplicateHostRecordID, "host", "Workstation 23")
		seedTimelineRecord(t, harness.DB, incidentID, adminUserID, timelinetest.RecordID)
		seedResolvedMention(t, harness.DB, adminUserID, entitytest.HostMentionID, timelinetest.RecordID, entitytest.DuplicateHostRecordID, timelinetest.FieldHostRefs, "WS-023")
		seedRecordLink(t, harness.DB, incidentID, adminUserID, linktest.DuplicateLinkID, timelinetest.RecordID, entitytest.DuplicateHostRecordID, "observed_on_host", "manual", nil)
		seedRecordTag(t, harness.DB, incidentID, adminUserID, linktest.TagIDSurvivor, entitytest.CanonicalHostRecordID, "critical-host")
		seedRecordTag(t, harness.DB, incidentID, adminUserID, linktest.TagIDLoser, entitytest.DuplicateHostRecordID, "critical-host")
		seedAssessment(t, harness.DB, incidentID, adminUserID, assessmenttest.HostAssessmentID, entitytest.DuplicateHostRecordID, "host", "confirmed")

		mergeResp := doEntitiesJSON(
			t,
			http.MethodPost,
			harness.Server.HTTP.URL+"/api/v1/records/"+entitytest.CanonicalHostRecordID.String()+"/merge",
			map[string]any{
				"loser_record_id":           entitytest.DuplicateHostRecordID.String(),
				"survivor_base_row_version": 1,
				"loser_base_row_version":    1,
				"client_txn_id":             "txn-entity_linking-i-4-03-merge",
				"reason":                    "  merge duplicate host  ",
			},
			withCookies(adminLogin.sessionCookie, adminLogin.csrfCookie),
			withHeader(authn.CSRFHeaderName, adminLogin.csrfCookie.Value),
		)
		if mergeResp.StatusCode != http.StatusOK {
			t.Fatalf("unexpected status: got %d want %d body=%#v", mergeResp.StatusCode, http.StatusOK, httptestx.ReadJSONBody(t, mergeResp))
		}
		mergeData := httptestx.RequireSuccessEnvelope(t, mergeResp, http.StatusOK)["data"].(map[string]any)
		if mergeData["survivor_record_id"] != entitytest.CanonicalHostRecordID.String() {
			t.Fatalf("unexpected survivor_record_id: %#v", mergeData)
		}
		if mergeData["loser_record_id"] != entitytest.DuplicateHostRecordID.String() {
			t.Fatalf("unexpected loser_record_id: %#v", mergeData)
		}
		if mergeData["record_type"] != "host" {
			t.Fatalf("unexpected canonical record_type: %#v", mergeData)
		}
		if got := int64(mergeData["survivor_row_version"].(float64)); got != 2 {
			t.Fatalf("expected survivor_row_version=2, got %d", got)
		}
		if got := int64(mergeData["loser_row_version"].(float64)); got != 2 {
			t.Fatalf("expected loser_row_version=2, got %d", got)
		}
		if mergeData["merged_into_record_id"] != entitytest.CanonicalHostRecordID.String() {
			t.Fatalf("expected merged_into_record_id to echo survivor, got %#v", mergeData)
		}

		summary := mergeData["merge_summary"].(map[string]any)
		if summary["record_type"] != "host" {
			t.Fatalf("unexpected merge summary record_type: %#v", summary)
		}
		if got := int(summary["repointed_mention_resolution_count"].(float64)); got != 1 {
			t.Fatalf("expected one repointed mention resolution, got %d", got)
		}
		if got := int(summary["repointed_link_count"].(float64)); got != 1 {
			t.Fatalf("expected one repointed link, got %d", got)
		}
		if got := int(summary["deduped_tag_count"].(float64)); got != 1 {
			t.Fatalf("expected one deduped tag, got %d", got)
		}
		if got := int(summary["suggestion_aliases_copied_count"].(float64)); got != 1 {
			t.Fatalf("expected one copied suggestion alias, got %#v", summary)
		}
		if got := int(summary["suggestion_alias_duplicate_noop_count"].(float64)); got != 0 {
			t.Fatalf("expected zero duplicate alias no-ops, got %#v", summary)
		}
		if got := int(summary["provenance_only_retained_count"].(float64)); got != 0 {
			t.Fatalf("expected zero provenance-only retained identifiers, got %#v", summary)
		}
		exactMatchClasses := summary["exact_match_classes"].([]any)
		if len(exactMatchClasses) != 3 {
			t.Fatalf("expected three host exact-match classes, got %#v", exactMatchClasses)
		}
		if exactMatchClasses[0].(map[string]any)["identifier_class"] != "aad_device_id" {
			t.Fatalf("unexpected host exact-match precedence: %#v", exactMatchClasses)
		}
		if exactMatchClasses[1].(map[string]any)["identifier_class"] != "fqdn" {
			t.Fatalf("unexpected host exact-match precedence: %#v", exactMatchClasses)
		}
		if got := int(exactMatchClasses[1].(map[string]any)["promoted_count"].(float64)); got != 1 {
			t.Fatalf("expected fqdn promoted_count=1, got %#v", exactMatchClasses[1])
		}

		survivorState, survivorMergedInto, survivorRowVersion, survivorFQDN := lookupHostState(t, harness.DB, entitytest.CanonicalHostRecordID)
		if survivorState != "canonical" || survivorMergedInto != nil || survivorRowVersion != 2 || survivorFQDN != "ws-023.corp.example.test" {
			t.Fatalf("unexpected survivor host state after merge: state=%s merged_into=%v row_version=%d fqdn=%q", survivorState, survivorMergedInto, survivorRowVersion, survivorFQDN)
		}
		loserState, loserMergedInto, loserRowVersion, _ := lookupHostState(t, harness.DB, entitytest.DuplicateHostRecordID)
		if loserState != "merged" || loserMergedInto == nil || *loserMergedInto != entitytest.CanonicalHostRecordID || loserRowVersion != 2 {
			t.Fatalf("unexpected loser host state after merge: state=%s merged_into=%v row_version=%d", loserState, loserMergedInto, loserRowVersion)
		}

		mention := lookupMention(t, harness.DB, entitytest.HostMentionID)
		entitytest.RequireMentionStatus(t, mention, entitytest.MentionStatusResolved)
		if mention.ResolvedRecordID == nil || *mention.ResolvedRecordID != entitytest.CanonicalHostRecordID {
			t.Fatalf("expected merge to repoint mention to survivor, got %#v", mention)
		}
		if mention.RowVersion != 2 {
			t.Fatalf("expected merge to increment mention row_version, got %#v", mention)
		}

		link := lookupActiveLink(t, harness.DB, incidentID, timelinetest.RecordID, entitytest.CanonicalHostRecordID, "observed_on_host")
		linktest.RequireActiveLink(t, link, timelinetest.RecordID, entitytest.CanonicalHostRecordID, "observed_on_host", "manual", nil)
		if got := queryCount(t, harness.DB, `
SELECT COUNT(*)
  FROM record_links
 WHERE record_link_id = $1
   AND deleted_at IS NULL
`, linktest.DuplicateLinkID); got != 0 {
			t.Fatalf("expected loser-targeted active link to disappear, got %d active rows", got)
		}

		if got := queryCount(t, harness.DB, `
SELECT COUNT(*)
  FROM record_tags
 WHERE incident_id = $1
   AND record_id = $2
   AND normalized_tag_name = 'critical-host'
   AND deleted_at IS NULL
`, incidentID, entitytest.CanonicalHostRecordID); got != 1 {
			t.Fatalf("expected one active survivor tag after dedupe, got %d", got)
		}
		if got := queryCount(t, harness.DB, `
SELECT COUNT(*)
  FROM record_tags
 WHERE record_id = $1
   AND deleted_at IS NULL
`, entitytest.DuplicateHostRecordID); got != 0 {
			t.Fatalf("expected loser active tags to be cleared, got %d", got)
		}
		if got := lookupAssessmentSubject(t, harness.DB, assessmenttest.HostAssessmentID); got != entitytest.CanonicalHostRecordID {
			t.Fatalf("expected loser assessment to repoint to survivor, got %s", got)
		}
		revisionsupport.RequireOneRecordChangeIntentPerRevisionSQL(t, harness.DB, mergeData["change_set_id"].(string))

		timelineChange := incidentwstest.RequireRecordChanged(t, timelineSocket, timelinetest.RecordID.String(), 1)
		if timelineChange.ChangeSetID != mergeData["change_set_id"] {
			t.Fatalf("expected websocket invalidation to carry the merge change_set_id, got timeline=%#v merge=%#v", timelineChange, mergeData)
		}
		survivorChange := incidentwstest.RequireRecordChanged(t, hostSocket, entitytest.CanonicalHostRecordID.String(), 2)
		if len(survivorChange.AffectedViews) != 1 || survivorChange.AffectedViews[0].ChangeKind != "invalidate" {
			t.Fatalf("expected survivor invalidation, got %#v", survivorChange)
		}
		loserChange := incidentwstest.RequireRecordChanged(t, hostSocket, entitytest.DuplicateHostRecordID.String(), 2)
		if len(loserChange.AffectedViews) != 1 || loserChange.AffectedViews[0].ChangeKind != "remove" {
			t.Fatalf("expected explicit loser removal, got %#v", loserChange)
		}
		viewLogin := appsupport.LoginResult{SessionCookie: adminLogin.sessionCookie, CSRFCookie: adminLogin.csrfCookie}
		survivorRow := workbookscenariotest.FindRow(t, workbookscenariotest.QueryViewRows(t, harness.Server.HTTP.URL, incidentID.String(), viewtest.HostsViewSchemaID, viewLogin), entitytest.CanonicalHostRecordID.String())
		if got := survivorRow["cells"].(map[string]any)["host.linked_event_count"].(map[string]any)["value"]; got != float64(1) {
			t.Fatalf("expected merge to project the repointed linked event on the survivor, got %#v row=%#v", got, survivorRow)
		}

		replayResp := doEntitiesJSON(
			t,
			http.MethodPost,
			harness.Server.HTTP.URL+"/api/v1/records/"+entitytest.CanonicalHostRecordID.String()+"/merge",
			map[string]any{
				"loser_record_id":           entitytest.DuplicateHostRecordID.String(),
				"survivor_base_row_version": 1,
				"loser_base_row_version":    1,
				"client_txn_id":             "txn-entity_linking-i-4-03-merge",
				"reason":                    "  merge duplicate host  ",
			},
			withCookies(adminLogin.sessionCookie, adminLogin.csrfCookie),
			withHeader(authn.CSRFHeaderName, adminLogin.csrfCookie.Value),
		)
		replayData := httptestx.RequireSuccessEnvelope(t, replayResp, http.StatusOK)["data"].(map[string]any)
		if replayData["change_set_id"] != mergeData["change_set_id"] {
			t.Fatalf("expected replayed merge to return the stored payload, got %#v %#v", mergeData, replayData)
		}
		revisionsupport.RequireOneRecordChangeIntentPerRevisionSQL(t, harness.DB, mergeData["change_set_id"].(string))

		divergentResp := doEntitiesJSON(
			t,
			http.MethodPost,
			harness.Server.HTTP.URL+"/api/v1/records/"+entitytest.CanonicalHostRecordID.String()+"/merge",
			map[string]any{
				"loser_record_id":           entitytest.DuplicateHostRecordID.String(),
				"survivor_base_row_version": 1,
				"loser_base_row_version":    1,
				"client_txn_id":             "txn-entity_linking-i-4-03-merge",
				"reason":                    "different replay payload",
			},
			withCookies(adminLogin.sessionCookie, adminLogin.csrfCookie),
			withHeader(authn.CSRFHeaderName, adminLogin.csrfCookie.Value),
		)
		httptestx.RequireErrorEnvelope(t, divergentResp, http.StatusConflict, "client_txn_conflict")

		createResp := doEntitiesJSON(
			t,
			http.MethodPost,
			harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID.String()+"/views/"+viewtest.HostsViewSchemaID+"/rows",
			map[string]any{
				"client_txn_id": "txn-entity_linking-i-4-03-create-after-merge",
				"host.fqdn":     "ws-023.corp.example.test",
			},
			withCookies(adminLogin.sessionCookie, adminLogin.csrfCookie),
			withHeader(authn.CSRFHeaderName, adminLogin.csrfCookie.Value),
		)
		createData := httptestx.RequireSuccessEnvelope(t, createResp, http.StatusOK)["data"].(map[string]any)
		row := createData["row"].(map[string]any)
		if row["record_id"] != entitytest.CanonicalHostRecordID.String() {
			t.Fatalf("expected carried-forward exact match to reuse survivor, got %#v", createData)
		}
		_ = link
	})

	t.Run("merge conflict exposes both supplied and current versions", func(t *testing.T) {
		harness := appsupport.StartServer(t, "entity_linking-i-4-03-version-conflict")
		adminLogin, adminUserID := provisionBootstrapAdmin(t, harness.Server)
		incident := createIncident(t, harness.Server, adminLogin, map[string]any{
			"client_txn_id": "txn-entity_linking-i-4-03-version-conflict-incident",
			"incident_key":  "IR-I403-V",
			"title":         "Entity merge version conflict",
		})
		incidentID := mustUUID(t, incident["incident_id"].(string))
		seedHostRecord(t, harness.DB, incidentID, adminUserID, entitytest.CanonicalHostRecordID, "Survivor Host", "SURVIVOR", "", "")
		seedHostRecord(t, harness.DB, incidentID, adminUserID, entitytest.DuplicateHostRecordID, "Loser Host", "LOSER", "", "")
		if _, err := harness.DB.ExecContext(context.Background(), `UPDATE records SET row_version = CASE record_id WHEN $1 THEN 2 ELSE 3 END WHERE record_id IN ($1, $2)`, entitytest.CanonicalHostRecordID, entitytest.DuplicateHostRecordID); err != nil {
			t.Fatalf("advance merge fixture versions: %v", err)
		}

		response := doEntitiesJSON(
			t,
			http.MethodPost,
			harness.Server.HTTP.URL+"/api/v1/records/"+entitytest.CanonicalHostRecordID.String()+"/merge",
			map[string]any{
				"loser_record_id":           entitytest.DuplicateHostRecordID.String(),
				"survivor_base_row_version": 1,
				"loser_base_row_version":    2,
				"client_txn_id":             "txn-entity_linking-i-4-03-version-conflict",
			},
			withCookies(adminLogin.sessionCookie, adminLogin.csrfCookie),
			withHeader(authn.CSRFHeaderName, adminLogin.csrfCookie.Value),
		)
		body := httptestx.RequireErrorEnvelope(t, response, http.StatusConflict, "row_version_conflict")
		details := body["error"].(map[string]any)["details"].(map[string]any)
		expected := map[string]any{
			"survivor_record_id":           entitytest.CanonicalHostRecordID.String(),
			"loser_record_id":              entitytest.DuplicateHostRecordID.String(),
			"survivor_base_row_version":    float64(1),
			"loser_base_row_version":       float64(2),
			"survivor_current_row_version": float64(2),
			"loser_current_row_version":    float64(3),
		}
		if !reflect.DeepEqual(details, expected) {
			t.Fatalf("unexpected complete merge conflict details: got=%#v want=%#v", details, expected)
		}
	})

	t.Run("host merge collision uses owner precondition detail shape", func(t *testing.T) {
		harness := appsupport.StartServer(t, "entity_linking-i-4-03-collision-detail")
		adminLogin, adminUserID := provisionBootstrapAdmin(t, harness.Server)
		incident := createIncident(t, harness.Server, adminLogin, map[string]any{
			"client_txn_id": "txn-entity_linking-i-4-03-collision-detail-incident",
			"incident_key":  "IR-I403-COLLISION",
			"title":         "Entity merge collision",
		})
		incidentID := mustUUID(t, incident["incident_id"].(string))
		blockingRecordID := uuid.New()

		seedHostRecord(t, harness.DB, incidentID, adminUserID, entitytest.CanonicalHostRecordID, "Survivor Host", "SURVIVOR-HOST", "survivor-host.corp.example.test", "")
		seedHostRecord(t, harness.DB, incidentID, adminUserID, entitytest.DuplicateHostRecordID, "Loser Host", "LOSER-HOST", "blocked-host.corp.example.test", "")
		seedHostRecord(t, harness.DB, incidentID, adminUserID, blockingRecordID, "Blocking Host", "BLOCKING-HOST", "blocked-host.corp.example.test", "")

		mergeResp := doEntitiesJSON(
			t,
			http.MethodPost,
			harness.Server.HTTP.URL+"/api/v1/records/"+entitytest.CanonicalHostRecordID.String()+"/merge",
			map[string]any{
				"loser_record_id":           entitytest.DuplicateHostRecordID.String(),
				"survivor_base_row_version": 1,
				"loser_base_row_version":    1,
				"client_txn_id":             "txn-entity_linking-i-4-03-collision-detail-merge",
			},
			withCookies(adminLogin.sessionCookie, adminLogin.csrfCookie),
			withHeader(authn.CSRFHeaderName, adminLogin.csrfCookie.Value),
		)
		errBody := httptestx.RequireErrorEnvelope(t, mergeResp, http.StatusConflict, "merge_precondition_failed")
		details := errBody["error"].(map[string]any)["details"].(map[string]any)
		if details["reason_code"] != "carry_forward_identifier_collision" {
			t.Fatalf("unexpected merge collision reason_code: %#v", details)
		}
		if details["identifier_class"] != "fqdn" || details["normalized_value"] != "blocked-host.corp.example.test" {
			t.Fatalf("unexpected merge collision identifier details: %#v", details)
		}
		if details["blocking_record_id"] != blockingRecordID.String() {
			t.Fatalf("expected blocking_record_id=%s, got %#v", blockingRecordID, details)
		}
		if _, exists := details["conflicting_record_id"]; exists {
			t.Fatalf("merge collision details must not expose non-owner conflicting_record_id: %#v", details)
		}
	})

	t.Run("merge authorization re-derives current incident role", func(t *testing.T) {
		harness := appsupport.StartServer(t, "entity_linking-i-4-03-authz")
		adminLogin, adminUserID := provisionBootstrapAdmin(t, harness.Server)
		incident := createIncident(t, harness.Server, adminLogin, map[string]any{
			"client_txn_id": "txn-entity_linking-i-4-03-authz-incident",
			"incident_key":  "IR-I403-A",
			"title":         "Entity merge authz",
		})
		incidentID := mustUUID(t, incident["incident_id"].(string))
		seedHostRecord(t, harness.DB, incidentID, adminUserID, entitytest.CanonicalHostRecordID, "WS-023", "WS-023", "", "")
		seedHostRecord(t, harness.DB, incidentID, adminUserID, entitytest.DuplicateHostRecordID, "WS-023 duplicate", "WS-023-DUP", "", "")

		if _, err := harness.DB.ExecContext(context.Background(), `
UPDATE incident_memberships
   SET role = 'viewer',
       updated_at = now(),
       updated_by_user_id = $3
 WHERE incident_id = $1
   AND user_id = $2
`, incidentID, adminUserID, adminUserID); err != nil {
			t.Fatalf("demote merge actor membership: %v", err)
		}

		resp := doEntitiesJSON(
			t,
			http.MethodPost,
			harness.Server.HTTP.URL+"/api/v1/records/"+entitytest.CanonicalHostRecordID.String()+"/merge",
			map[string]any{
				"loser_record_id":           entitytest.DuplicateHostRecordID.String(),
				"survivor_base_row_version": 1,
				"loser_base_row_version":    1,
				"client_txn_id":             "txn-entity_linking-i-4-03-authz-merge",
			},
			withCookies(adminLogin.sessionCookie, adminLogin.csrfCookie),
			withHeader(authn.CSRFHeaderName, adminLogin.csrfCookie.Value),
		)
		httptestx.RequireErrorEnvelope(t, resp, http.StatusForbidden, "authorization_denied")
	})

	t.Run("identity merge preserves loser lineage, raw mention text, and current-state readback", func(t *testing.T) {
		harness := appsupport.StartServer(t, "entity_linking-i-4-03-identity")
		adminLogin, adminUserID := provisionBootstrapAdmin(t, harness.Server)
		incident := createIncident(t, harness.Server, adminLogin, map[string]any{
			"client_txn_id": "txn-entity_linking-i-4-03-identity-incident",
			"incident_key":  "IR-I403-I",
			"title":         "Entity identity merge",
		})
		incidentID := mustUUID(t, incident["incident_id"].(string))
		viewLogin := appsupport.LoginResult{SessionCookie: adminLogin.sessionCookie, CSRFCookie: adminLogin.csrfCookie}

		entitytest.SeedIdentityRecord(t, harness.DB, incidentID, adminUserID, entitytest.CanonicalIdentityRecordID, "Alex Analyst", "alex.survivor@example.test", "alex.survivor@example.test", "ALEXSURV")
		entitytest.SeedIdentityRecord(t, harness.DB, incidentID, adminUserID, entitytest.DuplicateIdentityRecordID, "Alex Duplicate", "alex.analyst@example.test", "alex.analyst@example.test", "ALEXA")
		entitytest.SeedEntityAlias(t, harness.DB, incidentID, adminUserID, entitytest.DuplicateIdentityRecordID, "identity", "Case Owner")
		timelinetest.SeedTimelineRecord(t, harness.DB, incidentID, adminUserID, timelinetest.RecordID)
		entitytest.SeedResolvedMention(t, harness.DB, adminUserID, entitytest.IdentityMentionID, timelinetest.RecordID, entitytest.DuplicateIdentityRecordID, timelinetest.FieldIdentityRefs, "identity", "Case Owner")
		linktest.SeedRecordLink(t, harness.DB, incidentID, adminUserID, linktest.DuplicateLinkID, timelinetest.RecordID, entitytest.DuplicateIdentityRecordID, "observed_as_identity", "manual", nil)
		assessmenttest.SeedAssessment(t, harness.DB, incidentID, adminUserID, assessmenttest.IdentityAssessmentID, entitytest.DuplicateIdentityRecordID, "identity", "confirmed")
		beforeMention := lookupMention(t, harness.DB, entitytest.IdentityMentionID)

		mergeResp := doEntitiesJSON(
			t,
			http.MethodPost,
			harness.Server.HTTP.URL+"/api/v1/records/"+entitytest.CanonicalIdentityRecordID.String()+"/merge",
			map[string]any{
				"loser_record_id":           entitytest.DuplicateIdentityRecordID.String(),
				"survivor_base_row_version": 1,
				"loser_base_row_version":    1,
				"client_txn_id":             "txn-entity_linking-i-4-03-identity-merge",
				"reason":                    "merge duplicate identity",
			},
			withCookies(adminLogin.sessionCookie, adminLogin.csrfCookie),
			withHeader(authn.CSRFHeaderName, adminLogin.csrfCookie.Value),
		)
		mergeData := httptestx.RequireSuccessEnvelope(t, mergeResp, http.StatusOK)["data"].(map[string]any)
		if mergeData["survivor_record_id"] != entitytest.CanonicalIdentityRecordID.String() || mergeData["loser_record_id"] != entitytest.DuplicateIdentityRecordID.String() {
			t.Fatalf("unexpected identity merge payload: %#v", mergeData)
		}
		if mergeData["merge_summary"].(map[string]any)["record_type"] != "identity" {
			t.Fatalf("expected identity merge summary, got %#v", mergeData)
		}

		changeSet := asserttest.LookupChangeSet(t, asserttest.SQLDatabase(harness.DB), mergeData["change_set_id"].(string))
		auditassert.RequireMutationAttribution(t, auditassert.MutationAttribution{
			ActorUserID: changeSet.ActorUserID,
			Source:      changeSet.Source,
			ClientTxnID: changeSet.ClientTxnID,
			RequestID:   changeSet.RequestID,
			CreatedAt:   changeSet.CreatedAt,
		}, adminUserID.String(), "entities.records.merge", "txn-entity_linking-i-4-03-identity-merge")
		if got := asserttest.CountChangeSetMutations(t, asserttest.SQLDatabase(harness.DB), mergeData["change_set_id"].(string)); got < 2 {
			t.Fatalf("expected identity merge to emit at least two mutation rows, got %d", got)
		}

		survivorState, survivorMergedInto, survivorRowVersion, survivorEmail := lookupIdentityState(t, harness.DB, entitytest.CanonicalIdentityRecordID)
		if survivorState != "canonical" || survivorMergedInto != nil || survivorRowVersion != 2 || survivorEmail != "alex.survivor@example.test" {
			t.Fatalf("unexpected survivor identity state: state=%s merged_into=%v row_version=%d email=%q", survivorState, survivorMergedInto, survivorRowVersion, survivorEmail)
		}
		loserState, loserMergedInto, loserRowVersion, _ := lookupIdentityState(t, harness.DB, entitytest.DuplicateIdentityRecordID)
		if loserState != "merged" || loserMergedInto == nil || *loserMergedInto != entitytest.CanonicalIdentityRecordID || loserRowVersion != 2 {
			t.Fatalf("unexpected loser identity state: state=%s merged_into=%v row_version=%d", loserState, loserMergedInto, loserRowVersion)
		}

		afterMention := lookupMention(t, harness.DB, entitytest.IdentityMentionID)
		entitytest.RequireMentionStatus(t, afterMention, entitytest.MentionStatusResolved)
		if afterMention.ResolvedRecordID == nil || *afterMention.ResolvedRecordID != entitytest.CanonicalIdentityRecordID {
			t.Fatalf("expected identity merge to repoint mention resolution to survivor, got %#v", afterMention)
		}
		entitytest.RequireRawTextPreserved(t, beforeMention.RawText, afterMention.RawText)

		link := linktest.LookupActiveLink(t, harness.DB, incidentID, timelinetest.RecordID, entitytest.CanonicalIdentityRecordID, "observed_as_identity")
		linktest.RequireActiveLink(t, link, timelinetest.RecordID, entitytest.CanonicalIdentityRecordID, "observed_as_identity", "manual", nil)
		if got := assessmenttest.LookupAssessmentSubject(t, harness.DB, assessmenttest.IdentityAssessmentID); got != entitytest.CanonicalIdentityRecordID {
			t.Fatalf("expected identity assessment to repoint to survivor, got %s", got)
		}

		identityEnvelope := workbookscenariotest.QueryViewEnvelope(t, harness.Server.HTTP.URL, incidentID.String(), viewtest.IdentitiesViewSchemaID, viewLogin)
		contractassert.RequireDefaultQueryMeta(t, identityEnvelope, viewtest.IdentitiesViewSchemaID)
		identityRows := workbookscenariotest.QueryViewRows(t, harness.Server.HTTP.URL, incidentID.String(), viewtest.IdentitiesViewSchemaID, viewLogin)
		identitySurvivorRow := workbookscenariotest.FindRow(t, identityRows, entitytest.CanonicalIdentityRecordID.String())
		if got := identitySurvivorRow["cells"].(map[string]any)["identity.linked_event_count"].(map[string]any)["value"]; got != float64(1) {
			t.Fatalf("expected identity merge to project the repointed linked event on the survivor, got %#v row=%#v", got, identitySurvivorRow)
		}
		for _, row := range identityRows {
			if row["record_id"] == entitytest.DuplicateIdentityRecordID.String() {
				t.Fatalf("expected merged loser to disappear from current-state identity rows, got %#v", identityRows)
			}
		}

		createAfterMerge := appsupport.DoJSON(
			t,
			http.MethodPost,
			harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID.String()+"/views/"+viewtest.IdentitiesViewSchemaID+"/rows",
			map[string]any{
				"client_txn_id":         "txn-entity_linking-i-4-03-identity-after-merge",
				"identity.email":        "alex.analyst@example.test",
				"identity.display_name": "Alex After Merge",
			},
			appsupport.WithCookies(adminLogin.sessionCookie, adminLogin.csrfCookie),
			appsupport.WithHeader(authn.CSRFHeaderName, adminLogin.csrfCookie.Value),
		)
		createAfterMergeData := appsupport.RequireSuccessData(t, createAfterMerge, http.StatusOK)
		if createAfterMergeData["row"].(map[string]any)["record_id"] != entitytest.CanonicalIdentityRecordID.String() {
			t.Fatalf("expected carried-forward identity exact match to reuse survivor, got %#v", createAfterMergeData)
		}
	})
}
