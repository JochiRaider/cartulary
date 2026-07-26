package scenariotest

import (
	"database/sql"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	"github.com/JochiRaider/cartulary/internal/modules/timeline"
	"github.com/JochiRaider/cartulary/internal/modules/timeline/testsupport/asserttest"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	platformws "github.com/JochiRaider/cartulary/internal/platform/ws"
	"github.com/JochiRaider/cartulary/internal/testutil/auditassert"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
)

const timelineViewSchemaID = "cartulary.view.timeline.v2"

func CreateTimelineRow(t testing.TB, server *httptestx.Server, incidentID string, actor flowtest.LoginResult, body map[string]any) map[string]any {
	t.Helper()

	resp := httptestx.DoJSON(
		t,
		http.MethodPost,
		server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/views/"+timelineViewSchemaID+"/rows",
		body,
		httptestx.WithCookies(actor.SessionCookie, actor.CSRFCookie),
		httptestx.WithHeader(authn.CSRFHeaderName, actor.CSRFCookie.Value),
	)
	return httptestx.RequireSuccessEnvelope(t, resp, http.StatusCreated)["data"].(map[string]any)
}

func MustUUID(t testing.TB, raw string) uuid.UUID {
	t.Helper()

	value, err := uuid.Parse(raw)
	if err != nil {
		t.Fatalf("parse uuid: %v", err)
	}
	return value
}

func QueryTimelineEnvelope(t testing.TB, server *httptestx.Server, incidentID string, login flowtest.LoginResult, body map[string]any) map[string]any {
	t.Helper()

	queryResp := httptestx.DoJSON(
		t,
		http.MethodPost,
		server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/views/"+timeline.TimelineViewSchemaID+"/query",
		body,
		httptestx.WithCookies(login.SessionCookie),
	)
	return httptestx.RequireSuccessEnvelope(t, queryResp, http.StatusOK)
}

func QueryTimelineRows(t testing.TB, server *httptestx.Server, incidentID string, login flowtest.LoginResult) []any {
	t.Helper()

	return QueryTimelineEnvelope(t, server, incidentID, login, map[string]any{})["data"].(map[string]any)["rows"].([]any)
}

func FindRow(t testing.TB, rows []any, recordID string) map[string]any {
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

func RequireNoTimelineCollaborationEmission(t testing.TB, client *TimelineSocketClient, changes <-chan platformws.Message) {
	t.Helper()

	ExpectNoTimelineSocketMessage(t, client)
	asserttest.RequireNoRecordChange(t, changes, 300*time.Millisecond)
}

func RequireMutationRecorded(t testing.TB, db *sql.DB, changeSetID string, recordID string, wantActorUserID string, wantSource string, wantClientTxnID string, wantMutationRows int, wantRevisions int) {
	t.Helper()

	database := asserttest.SQLDatabase(db)
	changeSet := asserttest.LookupChangeSet(t, database, changeSetID)
	auditassert.RequireMutationAttribution(t, auditassert.MutationAttribution{
		ActorUserID: changeSet.ActorUserID,
		Source:      changeSet.Source,
		ClientTxnID: changeSet.ClientTxnID,
		RequestID:   changeSet.RequestID,
		CreatedAt:   changeSet.CreatedAt,
	}, wantActorUserID, wantSource, wantClientTxnID)
	if got := asserttest.CountChangeSetMutations(t, database, changeSetID); got != wantMutationRows {
		t.Fatalf("unexpected mutation row count for %s: got %d want %d", changeSetID, got, wantMutationRows)
	}
	if got := asserttest.CountRecordRevisions(t, database, recordID); got != wantRevisions {
		t.Fatalf("unexpected record revision count for %s: got %d want %d", recordID, got, wantRevisions)
	}
}
