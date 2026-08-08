package collaboration_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	platformws "github.com/JochiRaider/cartulary/internal/modules/collaboration"
	"github.com/JochiRaider/cartulary/internal/modules/collaboration/testsupport/incidentwstest"
	collabscenariotest "github.com/JochiRaider/cartulary/internal/modules/collaboration/testsupport/scenariotest"
	incidentscenariotest "github.com/JochiRaider/cartulary/internal/modules/incidents/testsupport/scenariotest"
	timelineroutetest "github.com/JochiRaider/cartulary/internal/modules/timeline/testsupport/routetest"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/jobs"
	"github.com/JochiRaider/cartulary/internal/testutil/collaborationsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/JochiRaider/cartulary/internal/testutil/wstest"
)

func TestTwoClientsPresenceReplay_Integration(t *testing.T) {
	runtime := collabscenariotest.StartRuntime(t)

	t.Run("two clients exchange canonical presence and replay in order", func(t *testing.T) {
		harness, admin, incidentID := setupSocketIncident(t, runtime, "collaboration-i-6-01-presence-replay")

		first := incidentwstest.ConnectAndHello(t, harness.Server.HTTP.URL, incidentID, incidentwstest.ConnectOptions{
			SessionToken:     admin.SessionCookie.Value,
			ClientInstanceID: "collaboration-i-6-01-first",
			Presence:         timelinePresence(),
		})
		if got := len(first.PresenceSnapshot); got != 1 {
			t.Fatalf("first presence_snapshot length = %d want 1: %#v", got, first.PresenceSnapshot)
		}
		firstResumeToken := first.HelloAck.ResumeToken

		second := incidentwstest.ConnectAndHello(t, harness.Server.HTTP.URL, incidentID, incidentwstest.ConnectOptions{
			SessionToken:     admin.SessionCookie.Value,
			ClientInstanceID: "collaboration-i-6-01-second",
			Presence:         timelinePresence(),
		})
		defer second.Close(wstest.StatusNormalClosure, "test_complete")
		if got := len(second.PresenceSnapshot); got != 2 {
			t.Fatalf("second presence_snapshot length = %d want 2: %#v", got, second.PresenceSnapshot)
		}

		delta := requirePresenceDelta(t, first, "upsert")
		if delta.ConnectionID != second.HelloAck.ConnectionID {
			t.Fatalf("first client saw presence_delta for %s want second connection %s", delta.ConnectionID, second.HelloAck.ConnectionID)
		}

		first.Close(wstest.StatusNormalClosure, "test_complete")
		publishJobProgress(t, harness, incidentID, "collaboration-i-6-01-job-a", platformws.JobStatusQueued)
		publishJobProgress(t, harness, incidentID, "collaboration-i-6-01-job-b", platformws.JobStatusRunning)
		waitForReplayEventCount(t, harness, incidentID, 2)

		resumed := incidentwstest.ConnectAndResume(t, harness.Server.HTTP.URL, incidentID, incidentwstest.ConnectOptions{
			SessionToken:     admin.SessionCookie.Value,
			ClientInstanceID: "collaboration-i-6-01-first",
			Presence:         timelinePresence(),
		}, firstResumeToken, 0)
		defer resumed.Close(wstest.StatusNormalClosure, "test_complete")
		if resumed.ResumeAck.Status != platformws.ResumeStatusReplayed {
			t.Fatalf("resume status = %q want %q", resumed.ResumeAck.Status, platformws.ResumeStatusReplayed)
		}
		requireReplayTypes(t, resumed.ReplayedMessages, "job_progress", "job_progress")
		if got := len(resumed.PresenceSnapshot); got != 2 {
			t.Fatalf("resumed presence_snapshot length = %d want 2: %#v", got, resumed.PresenceSnapshot)
		}
	})
}

func TestIncidentSocketRevocationSources(t *testing.T) {
	runtime := collabscenariotest.StartRuntime(t)
	harness, admin, _, incidentID := setupSocketIncidentWithAdminID(t, runtime, "collaboration-support-socket-revocations")

	logoutUser := flowtest.SeedLocalUserRecord(t, harness.DB, "collaboration-support-socket-logout@example.test", "Collaboration Logout", "CollaborationLogoutPass1!", false, false, true)
	expiryUser := flowtest.SeedLocalUserRecord(t, harness.DB, "collaboration-support-socket-expiry@example.test", "Collaboration Expiry", "CollaborationExpiryPass1!", false, false, true)
	concurrencyUser := flowtest.SeedLocalUserRecord(t, harness.DB, "collaboration-support-socket-concurrency@example.test", "Collaboration Concurrency", "CollaborationConcurrencyPass1!", false, false, true)
	member := flowtest.SeedLocalUserRecord(t, harness.DB, "collaboration-support-socket-member@example.test", "Collaboration Member", "CollaborationMemberPass1!", false, false, true)
	incidentscenariotest.CreateMembershipForUser(t, harness.Server, admin, incidentID, logoutUser.ID.String(), logoutUser.Email, "editor")
	incidentscenariotest.CreateMembershipForUser(t, harness.Server, admin, incidentID, expiryUser.ID.String(), expiryUser.Email, "editor")
	incidentscenariotest.CreateMembershipForUser(t, harness.Server, admin, incidentID, concurrencyUser.ID.String(), concurrencyUser.Email, "editor")
	incidentscenariotest.CreateMembershipForUser(t, harness.Server, admin, incidentID, member.ID.String(), member.Email, "editor")

	logoutSession, logoutCSRF := flowtest.LoginLocalUser(t, harness.Server.HTTP.URL, logoutUser.Email, "CollaborationLogoutPass1!", nil)
	expirySession, _ := flowtest.LoginLocalUser(t, harness.Server.HTTP.URL, expiryUser.Email, "CollaborationExpiryPass1!", nil)
	concurrencySession, _ := flowtest.LoginLocalUser(t, harness.Server.HTTP.URL, concurrencyUser.Email, "CollaborationConcurrencyPass1!", nil)
	memberSession, _ := flowtest.LoginLocalUser(t, harness.Server.HTTP.URL, member.Email, "CollaborationMemberPass1!", nil)

	t.Run("current session logout", func(t *testing.T) {
		logoutSocket := incidentwstest.ConnectAndHello(t, harness.Server.HTTP.URL, incidentID, incidentwstest.ConnectOptions{
			SessionToken:     logoutSession.Value,
			ClientInstanceID: "collaboration-support-socket-logout",
			Presence:         timelinePresence(),
		})
		logoutResp := httptestx.DoJSON(
			t,
			http.MethodPost,
			harness.Server.HTTP.URL+"/api/v1/auth/logout",
			map[string]any{},
			httptestx.WithCookies(logoutSession, logoutCSRF),
			httptestx.WithHeader(authn.CSRFHeaderName, logoutCSRF.Value),
		)
		httptestx.RequireSuccessEnvelope(t, logoutResp, http.StatusOK)
		incidentwstest.ExpectSessionRevoked(t, logoutSocket, "session_revoked")
	})

	t.Run("concurrency limit eviction", func(t *testing.T) {
		concurrencySocket := incidentwstest.ConnectAndHello(t, harness.Server.HTTP.URL, incidentID, incidentwstest.ConnectOptions{
			SessionToken:     concurrencySession.Value,
			ClientInstanceID: "collaboration-support-socket-concurrency",
			Presence:         timelinePresence(),
		})
		sessionID := uuid.MustParse(sessionIDForCookie(t, harness, concurrencyUser.ID.String()))
		harness.Collaboration.RevokeSession(sessionID, authn.ConcurrencyLimitReasonCode)
		incidentwstest.ExpectSessionRevoked(t, concurrencySocket, authn.ConcurrencyLimitReasonCode)
	})

	t.Run("incident membership removal", func(t *testing.T) {
		memberSocket := incidentwstest.ConnectAndHello(t, harness.Server.HTTP.URL, incidentID, incidentwstest.ConnectOptions{
			SessionToken:     memberSession.Value,
			ClientInstanceID: "collaboration-support-socket-membership",
			Presence:         timelinePresence(),
		})
		incidentscenariotest.DeleteMembershipVersion(t, harness.Server, admin, incidentID, member.ID.String(), queryMembershipVersion(t, harness, incidentID, member.ID.String()))
		incidentwstest.ExpectSessionRevoked(t, memberSocket, "incident_access_revoked")
	})

	t.Run("incident close", func(t *testing.T) {
		closeIncident := incidentscenariotest.CreateIncident(t, harness.Server, admin, map[string]any{
			"client_txn_id": "txn-collaboration-support-socket-incident-close-incident",
			"incident_key":  "IR-COLLABORATIONSUPPORTSOCKETCLOSE",
			"title":         "Collaboration support socket close",
		})
		closeIncidentID := closeIncident["incident_id"].(string)
		closeSocket := incidentwstest.ConnectAndHello(t, harness.Server.HTTP.URL, closeIncidentID, incidentwstest.ConnectOptions{
			SessionToken:     admin.SessionCookie.Value,
			ClientInstanceID: "collaboration-support-socket-incident-close",
			Presence:         timelinePresence(),
		})
		closeResp := httptestx.DoJSON(
			t,
			http.MethodPost,
			harness.Server.HTTP.URL+"/api/v1/incidents/"+closeIncidentID+"/close",
			map[string]any{
				"base_incident_version": 1,
				"client_txn_id":         "txn-collaboration-support-socket-incident-close",
				"reason":                "Close incident to terminate writable collaboration.",
			},
			httptestx.WithCookies(admin.SessionCookie, admin.CSRFCookie),
			httptestx.WithHeader(authn.CSRFHeaderName, admin.CSRFCookie.Value),
		)
		httptestx.RequireSuccessEnvelope(t, closeResp, http.StatusOK)
		incidentwstest.ExpectIncidentClosed(t, closeSocket)
	})

	t.Run("idle expiry", func(t *testing.T) {
		expirySocket := incidentwstest.ConnectAndHello(t, harness.Server.HTTP.URL, incidentID, incidentwstest.ConnectOptions{
			SessionToken:     expirySession.Value,
			ClientInstanceID: "collaboration-support-socket-expiry",
			Presence:         timelinePresence(),
		})
		if err := expirySocket.Send(context.Background(), platformws.Message{
			Type:    "pong",
			Payload: platformws.RawPayload(map[string]any{}),
		}); err != nil {
			t.Fatalf("send pong before idle expiry: %v", err)
		}
		harness.Server.Clock.Advance(31 * time.Minute)
		incidentwstest.ExpectSessionRevoked(t, expirySocket, "session_expired")
	})
}

func TestClosedIncidentSocketTerminatesBeforeWritableAck(t *testing.T) {
	runtime := collabscenariotest.StartRuntime(t)
	harness, admin, incidentID := setupSocketIncident(t, runtime, "collaboration-support-closed-socket")

	initial := incidentwstest.ConnectAndHello(t, harness.Server.HTTP.URL, incidentID, incidentwstest.ConnectOptions{
		SessionToken:     admin.SessionCookie.Value,
		ClientInstanceID: "collaboration-support-closed-socket-source",
		Presence:         timelinePresence(),
	})
	resumeToken := initial.HelloAck.ResumeToken
	initial.Close(wstest.StatusNormalClosure, "test_complete")

	closeResp := httptestx.DoJSON(
		t,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/close",
		map[string]any{
			"base_incident_version": 1,
			"client_txn_id":         "txn-collaboration-support-closed-socket-close",
			"reason":                "Close incident before new writable socket attempts.",
		},
		httptestx.WithCookies(admin.SessionCookie, admin.CSRFCookie),
		httptestx.WithHeader(authn.CSRFHeaderName, admin.CSRFCookie.Value),
	)
	httptestx.RequireSuccessEnvelope(t, closeResp, http.StatusOK)

	t.Run("hello", func(t *testing.T) {
		conn, _, err := incidentwstest.TryConnect(harness.Server.HTTP.URL, incidentID, incidentwstest.ConnectOptions{
			SessionToken: admin.SessionCookie.Value,
		})
		if err != nil {
			t.Fatalf("dial closed incident websocket: %v", err)
		}
		defer conn.CloseNow()

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := conn.WriteJSON(ctx, platformws.Message{
			Type: "hello",
			Payload: platformws.RawPayload(map[string]any{
				"client_instance_id": "collaboration-support-closed-socket-hello",
				"presence":           timelinePresence(),
			}),
		}); err != nil {
			t.Fatalf("send closed incident hello: %v", err)
		}
		requireClosedIncidentTerminal(t, ctx, conn)
	})

	t.Run("resume", func(t *testing.T) {
		conn, _, err := incidentwstest.TryConnect(harness.Server.HTTP.URL, incidentID, incidentwstest.ConnectOptions{
			SessionToken: admin.SessionCookie.Value,
		})
		if err != nil {
			t.Fatalf("dial closed incident websocket: %v", err)
		}
		defer conn.CloseNow()

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := conn.WriteJSON(ctx, platformws.Message{
			Type: "resume",
			Payload: platformws.RawPayload(map[string]any{
				"client_instance_id":   "collaboration-support-closed-socket-source",
				"resume_token":         resumeToken,
				"last_seen_stream_seq": 0,
				"presence":             timelinePresence(),
			}),
		}); err != nil {
			t.Fatalf("send closed incident resume: %v", err)
		}
		requireClosedIncidentTerminal(t, ctx, conn)
	})
}

func requireClosedIncidentTerminal(t testing.TB, ctx context.Context, conn *wstest.Client) {
	t.Helper()

	var message platformws.Message
	if err := conn.ReadJSON(ctx, &message); err != nil {
		t.Fatalf("read closed incident terminal error: %v", err)
	}
	if message.Type != "error" {
		t.Fatalf("closed incident first response type = %q want error", message.Type)
	}
	var payload map[string]any
	if err := json.Unmarshal(message.Payload, &payload); err != nil {
		t.Fatalf("decode closed incident terminal payload: %v", err)
	}
	if payload["code"] != platformws.IncidentTerminalClosed || payload["retryable"] != false {
		t.Fatalf("unexpected closed incident terminal payload: %#v", payload)
	}

	var next platformws.Message
	err := conn.ReadJSON(ctx, &next)
	wstest.RequireClose(t, err, wstest.StatusPolicyViolation, "incident_closed")
}

func TestResumeReplaysReplayableMessagesOnly_Integration(t *testing.T) {
	runtime := collabscenariotest.StartRuntime(t)
	harness, admin, incidentID := setupSocketIncident(t, runtime, "collaboration-i-6-02-replayable-only")

	source := incidentwstest.ConnectAndHello(t, harness.Server.HTTP.URL, incidentID, incidentwstest.ConnectOptions{
		SessionToken:     admin.SessionCookie.Value,
		ClientInstanceID: "collaboration-i-6-02-source",
		Presence:         timelinePresence(),
	})
	resumeToken := source.HelloAck.ResumeToken
	source.Close(wstest.StatusNormalClosure, "test_complete")

	other := incidentwstest.ConnectAndHello(t, harness.Server.HTTP.URL, incidentID, incidentwstest.ConnectOptions{
		SessionToken:     admin.SessionCookie.Value,
		ClientInstanceID: "collaboration-i-6-02-other",
		Presence:         timelinePresence(),
	})
	defer other.Close(wstest.StatusNormalClosure, "test_complete")

	timelineroutetest.CreateRow(t, harness.Server, admin, incidentID, map[string]any{
		"client_txn_id":                   "txn-collaboration-i-6-02-record",
		"timeline.activity_synopsis_text": "Collaboration replayable record change",
	})

	publishJobProgress(t, harness, incidentID, "collaboration-i-6-02-job", platformws.JobStatusRunning)
	waitForReplayEventCount(t, harness, incidentID, 2)

	resumed := incidentwstest.ConnectAndResume(t, harness.Server.HTTP.URL, incidentID, incidentwstest.ConnectOptions{
		SessionToken:     admin.SessionCookie.Value,
		ClientInstanceID: "collaboration-i-6-02-source",
		Presence:         timelinePresence(),
	}, resumeToken, 0)
	defer resumed.Close(wstest.StatusNormalClosure, "test_complete")

	if resumed.ResumeAck.Status != platformws.ResumeStatusReplayed {
		t.Fatalf("resume status = %q want %q", resumed.ResumeAck.Status, platformws.ResumeStatusReplayed)
	}
	requireReplayTypes(t, resumed.ReplayedMessages, "record_changed", "job_progress")
	for _, message := range resumed.ReplayedMessages {
		if message.Type == "presence_snapshot" || message.Type == "presence_delta" {
			t.Fatalf("resume replay must not include presence message: %#v", resumed.ReplayedMessages)
		}
	}
	if got := len(resumed.PresenceSnapshot); got != 2 {
		t.Fatalf("resume must rehydrate presence through fresh snapshot, got %d presences: %#v", got, resumed.PresenceSnapshot)
	}
}

func TestCookieSocketRejectsUntrustedOrigin_Integration(t *testing.T) {
	runtime := collabscenariotest.StartRuntime(t)
	harness, admin, incidentID := setupSocketIncident(t, runtime, "collaboration-i-6-04-origin")

	incidentwstest.RequireDialRejectedStatus(t, harness.Server.HTTP.URL, incidentID, incidentwstest.ConnectOptions{
		Cookies: []*http.Cookie{admin.SessionCookie},
		Origin:  "https://untrusted.example.test",
	}, http.StatusForbidden)

	client := incidentwstest.ConnectAndHello(t, harness.Server.HTTP.URL, incidentID, incidentwstest.ConnectOptions{
		SessionToken:     admin.SessionCookie.Value,
		ClientInstanceID: "collaboration-i-6-04-authorized",
		Presence:         timelinePresence(),
	})
	defer client.Close(wstest.StatusNormalClosure, "test_complete")
	if got := len(client.PresenceSnapshot); got != 1 {
		t.Fatalf("rejected origin must not subscribe before a valid client connects, snapshot len=%d snapshot=%#v", got, client.PresenceSnapshot)
	}
}

func setupSocketIncidentWithAdminID(t testing.TB, runtime *collabscenariotest.RuntimeHarness, prefix string) (*collabscenariotest.ServerHarness, flowtest.LoginResult, uuid.UUID, string) {
	t.Helper()

	harness := runtime.StartServer(t, prefix)
	admin, adminID := flowtest.ProvisionBootstrapAdminUUID(t, harness.Server.HTTP.URL)
	incident := incidentscenariotest.CreateIncident(t, harness.Server, admin, map[string]any{
		"client_txn_id": "txn-" + prefix,
		"incident_key":  "IR-" + prefix,
		"title":         prefix,
	})
	return harness, admin, adminID, incident["incident_id"].(string)
}

func publishJobProgress(t testing.TB, harness *collabscenariotest.ServerHarness, incidentID string, jobID string, status string) {
	t.Helper()
	_ = jobID
	_ = status

	var actorUserID string
	if err := harness.DB.QueryRowContext(context.Background(), `
SELECT user_id::text
  FROM incident_memberships
 WHERE incident_id = $1::uuid
 ORDER BY joined_at, user_id
 LIMIT 1
`, incidentID).Scan(&actorUserID); err != nil {
		t.Fatalf("load job progress actor: %v", err)
	}
	now := harness.Server.Clock.Now().UTC()
	tx, err := harness.Pool.Begin(context.Background())
	if err != nil {
		t.Fatalf("begin job progress transaction: %v", err)
	}
	defer func() { _ = tx.Rollback(context.Background()) }()
	incidentUUID := uuid.MustParse(incidentID)
	actorUUID := uuid.MustParse(actorUserID)
	if _, err := collaborationsupport.NewJobTransactions().CreateQueuedTx(context.Background(), tx, jobs.CreateParams{
		JobKind:           collaborationsupport.TestJobKind,
		Scope:             jobs.Scope{Kind: jobs.ScopeKindIncident, IncidentID: &incidentUUID},
		SubmittedByUserID: actorUUID,
		Cancelable:        true,
		Progress:          jobs.Progress{Completed: 1, Total: intPointer(2)},
	}, now); err != nil {
		t.Fatalf("append durable job progress: %v", err)
	}
	if err := tx.Commit(context.Background()); err != nil {
		t.Fatalf("commit job progress: %v", err)
	}
}

func intPointer(value int) *int { return &value }

func waitForReplayEventCount(t testing.TB, harness *collabscenariotest.ServerHarness, incidentID string, want int) {
	t.Helper()

	deadline := time.Now().Add(5 * time.Second)
	for {
		var count int
		if err := harness.DB.QueryRowContext(context.Background(), `
SELECT count(*)
  FROM collaboration_replay_events
 WHERE incident_id = $1::uuid
`, incidentID).Scan(&count); err != nil {
			t.Fatalf("count durable replay events: %v", err)
		}
		if count >= want {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("durable replay event count = %d want at least %d", count, want)
		}
		time.Sleep(20 * time.Millisecond)
	}
}

func queryMembershipVersion(t testing.TB, harness *collabscenariotest.ServerHarness, incidentID string, userID string) int64 {
	t.Helper()

	var version int64
	if err := harness.DB.QueryRowContext(context.Background(), `
SELECT membership_version
  FROM incident_memberships
 WHERE incident_id = $1::uuid
   AND user_id = $2::uuid
`, incidentID, userID).Scan(&version); err != nil {
		t.Fatalf("query membership version: %v", err)
	}
	return version
}
