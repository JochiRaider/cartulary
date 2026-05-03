package phase4test

import (
	"database/sql"
	"net/http"
	"testing"

	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/JochiRaider/cartulary/internal/testutil/timelinetest"
)

type RouteConformanceCase struct {
	Route                  RouteInventoryEntry
	Context                RouteInventoryContext
	ClientTxnID            string
	Login                  LoginResult
	ActorUserID            string
	ExpectedMutationSource string
}

func RequireRouteReplayHistoryConformance(t testing.TB, db *sql.DB, serverURL string, c RouteConformanceCase) map[string]any {
	t.Helper()
	if c.Route.ReplayCapability == RouteReplayNotApplicable {
		t.Fatalf("route %s does not declare replay conformance", c.Route.Key)
	}
	if c.Route.BuildDivergentBody == nil {
		t.Fatalf("route %s does not declare a divergent replay payload", c.Route.Key)
	}
	if c.ClientTxnID == "" {
		t.Fatalf("route %s conformance case missing client transaction id", c.Route.Key)
	}

	firstResp := DoRoute(t, serverURL, c.Route, c.Context, c.ClientTxnID, c.Login, c.Route.BuildBody(c.Context, c.ClientTxnID))
	firstData := RequireSuccessData(t, firstResp, c.Route.SuccessStatus)
	if c.ExpectedMutationSource != "" {
		changeSetID, ok := firstData["change_set_id"].(string)
		if !ok || changeSetID == "" {
			t.Fatalf("route %s expected attributed mutation response with change_set_id, got %#v", c.Route.Key, firstData)
		}
		RequireChangeSetAttribution(t, db, changeSetID, c.ActorUserID, c.ExpectedMutationSource, c.ClientTxnID)
	}

	affectedRecordID := ""
	if c.Route.AffectedRecordID != nil {
		affectedRecordID = c.Route.AffectedRecordID(c.Context, firstData)
	}
	stableBefore := SnapshotReplayCounts(t, db, c.Context.IncidentID, affectedRecordID)
	replayResp := DoRoute(t, serverURL, c.Route, c.Context, c.ClientTxnID, c.Login, c.Route.BuildBody(c.Context, c.ClientTxnID))
	replayData := RequireSuccessData(t, replayResp, c.Route.ReplayStatus)
	if firstChangeSet, ok := firstData["change_set_id"].(string); ok && firstChangeSet != "" {
		if replayData["change_set_id"] != firstChangeSet {
			t.Fatalf("route %s replay returned different change_set_id: first=%#v replay=%#v", c.Route.Key, firstData, replayData)
		}
	}
	stableAfter := SnapshotReplayCounts(t, db, c.Context.IncidentID, affectedRecordID)
	httptestx.RequireReplayScaffold(t, httptestx.ReplayExpectation{
		FirstStatus:     c.Route.SuccessStatus,
		ReplayStatus:    c.Route.ReplayStatus,
		DivergentStatus: c.Route.DivergentStatus,
		DivergentCode:   c.Route.DivergentCode,
		StableBefore:    stableBefore,
		StableAfter:     stableAfter,
	})

	divergentResp := DoRoute(t, serverURL, c.Route, c.Context, c.ClientTxnID, c.Login, c.Route.BuildDivergentBody(c.Context, c.ClientTxnID))
	divergentBody := RequireErrorBody(t, divergentResp, c.Route.DivergentStatus, c.Route.DivergentCode)
	httptestx.RequireDivergentReplayRejected(
		t,
		divergentResp.StatusCode,
		divergentBody["error"].(map[string]any)["code"].(string),
		c.Route.DivergentCode,
	)

	return firstData
}

func DoRoute(t testing.TB, serverURL string, route RouteInventoryEntry, ctx RouteInventoryContext, clientTxnID string, login LoginResult, body any) *http.Response {
	t.Helper()
	options := []func(*http.Request){WithCookies(login.SessionCookie, login.CSRFCookie)}
	if route.RequiresCSRF {
		options = append(options, WithHeader(authn.CSRFHeaderName, login.CSRFCookie.Value))
	}
	return DoJSON(t, route.Method, serverURL+route.BuildPath(ctx), body, options...)
}

func SnapshotReplayCounts(t testing.TB, db *sql.DB, incidentID string, recordID string) httptestx.ReplayCounts {
	t.Helper()

	counts := httptestx.ReplayCounts{
		ChangeSets: QueryCount(t, db, `SELECT COUNT(*) FROM change_sets WHERE incident_id::text = $1`, incidentID),
		MutationRows: QueryCount(t, db, `
SELECT COUNT(*)
  FROM change_set_mutations m
  JOIN change_sets c ON c.change_set_id = m.change_set_id
 WHERE c.incident_id::text = $1
`, incidentID),
	}
	if recordID != "" {
		counts.Revisions = QueryCount(t, db, `SELECT COUNT(*) FROM record_revisions WHERE record_id::text = $1`, recordID)
	}
	return counts
}

func RequireChangeSetAttribution(t testing.TB, db *sql.DB, changeSetID string, actorUserID string, source string, clientTxnID string) {
	t.Helper()
	changeSet := timelinetest.LookupChangeSet(t, db, changeSetID)
	httptestx.RequireMutationAttribution(t, httptestx.MutationAttribution{
		ActorUserID: changeSet.ActorUserID,
		Source:      changeSet.Source,
		ClientTxnID: changeSet.ClientTxnID,
		RequestID:   changeSet.RequestID,
		CreatedAt:   changeSet.CreatedAt,
	}, actorUserID, source, clientTxnID)
}

func MustRoute(t testing.TB, key RouteKey, ctx RouteInventoryContext) RouteInventoryEntry {
	t.Helper()
	for _, route := range Phase4RouteInventory(ctx) {
		if route.Key == key {
			return route
		}
	}
	t.Fatalf("phase4 route inventory missing %s", key)
	return RouteInventoryEntry{}
}
