package collaboration_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/platform/authn"
	platformws "github.com/JochiRaider/cartulary/internal/platform/ws"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/JochiRaider/cartulary/internal/testutil/incidentwstest"
	"github.com/JochiRaider/cartulary/internal/testutil/phase3test"
)

func TestPhase6_TwoClientsPresenceReplay_I_6_01(t *testing.T) {
	runtime := phase3test.StartRuntime(t)

	t.Run("two clients exchange canonical presence and replay in order", func(t *testing.T) {
		harness, admin, incidentID := setupPhase6SocketIncident(t, runtime, "phase6-i-6-01-presence-replay")

		first := incidentwstest.ConnectAndHello(t, harness.Server.HTTP.URL, incidentID, incidentwstest.ConnectOptions{
			SessionToken:     admin.SessionCookie.Value,
			ClientInstanceID: "phase6-i-6-01-first",
			Presence:         timelinePresence(),
		})
		if got := len(first.PresenceSnapshot); got != 1 {
			t.Fatalf("first presence_snapshot length = %d want 1: %#v", got, first.PresenceSnapshot)
		}
		firstResumeToken := first.HelloAck.ResumeToken

		second := incidentwstest.ConnectAndHello(t, harness.Server.HTTP.URL, incidentID, incidentwstest.ConnectOptions{
			SessionToken:     admin.SessionCookie.Value,
			ClientInstanceID: "phase6-i-6-01-second",
			Presence:         timelinePresence(),
		})
		defer second.Close(websocket.StatusNormalClosure, "test_complete")
		if got := len(second.PresenceSnapshot); got != 2 {
			t.Fatalf("second presence_snapshot length = %d want 2: %#v", got, second.PresenceSnapshot)
		}

		delta := requirePresenceDelta(t, first, "upsert")
		if delta.ConnectionID != second.HelloAck.ConnectionID {
			t.Fatalf("first client saw presence_delta for %s want second connection %s", delta.ConnectionID, second.HelloAck.ConnectionID)
		}

		first.Close(websocket.StatusNormalClosure, "test_complete")
		publishPhase6JobProgress(t, harness, incidentID, "phase6-i-6-01-job-a", platformws.JobStatusQueued)
		publishPhase6JobProgress(t, harness, incidentID, "phase6-i-6-01-job-b", platformws.JobStatusRunning)

		resumed := incidentwstest.ConnectAndResume(t, harness.Server.HTTP.URL, incidentID, incidentwstest.ConnectOptions{
			SessionToken:     admin.SessionCookie.Value,
			ClientInstanceID: "phase6-i-6-01-first",
			Presence:         timelinePresence(),
		}, firstResumeToken, 0)
		defer resumed.Close(websocket.StatusNormalClosure, "test_complete")
		if resumed.ResumeAck.Status != platformws.ResumeStatusReplayed {
			t.Fatalf("resume status = %q want %q", resumed.ResumeAck.Status, platformws.ResumeStatusReplayed)
		}
		requireReplayTypes(t, resumed.ReplayedMessages, "job_progress", "job_progress")
		if got := len(resumed.PresenceSnapshot); got != 2 {
			t.Fatalf("resumed presence_snapshot length = %d want 2: %#v", got, resumed.PresenceSnapshot)
		}
	})

	t.Run("revocations emit session_revoked before close", func(t *testing.T) {
		harness, admin, _, incidentID := setupPhase6SocketIncidentWithAdminID(t, runtime, "phase6-i-6-01-revocations")

		logoutUser := phase3test.SeedLocalUserFlags(t, harness.DB, "phase6-i-6-01-logout@example.test", "Phase 6 Logout", "Phase6LogoutPass1!", false, false, true)
		expiryUser := phase3test.SeedLocalUserFlags(t, harness.DB, "phase6-i-6-01-expiry@example.test", "Phase 6 Expiry", "Phase6ExpiryPass1!", false, false, true)
		concurrencyUser := phase3test.SeedLocalUserFlags(t, harness.DB, "phase6-i-6-01-concurrency@example.test", "Phase 6 Concurrency", "Phase6ConcurrencyPass1!", false, false, true)
		member := phase3test.SeedLocalUserFlags(t, harness.DB, "phase6-i-6-01-member@example.test", "Phase 6 Member", "Phase6MemberPass1!", false, false, true)
		phase3test.CreateMembership(t, harness.Server, incidentID, logoutUser.ID.String(), logoutUser.Email, "editor", admin)
		phase3test.CreateMembership(t, harness.Server, incidentID, expiryUser.ID.String(), expiryUser.Email, "editor", admin)
		phase3test.CreateMembership(t, harness.Server, incidentID, concurrencyUser.ID.String(), concurrencyUser.Email, "editor", admin)
		phase3test.CreateMembership(t, harness.Server, incidentID, member.ID.String(), member.Email, "editor", admin)

		logoutSession, logoutCSRF := phase3test.LoginLocalUser(t, harness.Server, logoutUser.Email, "Phase6LogoutPass1!")
		expirySession, _ := phase3test.LoginLocalUser(t, harness.Server, expiryUser.Email, "Phase6ExpiryPass1!")
		concurrencySession, _ := phase3test.LoginLocalUser(t, harness.Server, concurrencyUser.Email, "Phase6ConcurrencyPass1!")
		memberSession, _ := phase3test.LoginLocalUser(t, harness.Server, member.Email, "Phase6MemberPass1!")

		logoutSocket := incidentwstest.ConnectAndHello(t, harness.Server.HTTP.URL, incidentID, incidentwstest.ConnectOptions{
			SessionToken:     logoutSession.Value,
			ClientInstanceID: "phase6-i-6-01-logout",
			Presence:         timelinePresence(),
		})
		logoutResp := phase3test.DoJSON(
			t,
			http.MethodPost,
			harness.Server.HTTP.URL+"/api/v1/auth/logout",
			map[string]any{},
			phase3test.WithCookies(logoutSession, logoutCSRF),
			phase3test.WithHeader(authn.CSRFHeaderName, logoutCSRF.Value),
		)
		httptestx.RequireSuccessEnvelope(t, logoutResp, http.StatusOK)
		incidentwstest.ExpectSessionRevoked(t, logoutSocket, "session_revoked")

		concurrencySocket := incidentwstest.ConnectAndHello(t, harness.Server.HTTP.URL, incidentID, incidentwstest.ConnectOptions{
			SessionToken:     concurrencySession.Value,
			ClientInstanceID: "phase6-i-6-01-concurrency",
			Presence:         timelinePresence(),
		})
		sessionID := phase3test.MustUUID(t, sessionIDForCookie(t, harness, concurrencyUser.ID.String()))
		harness.Server.Runtime.WSHub.RevokeSession(sessionID, authn.ConcurrencyLimitReasonCode)
		incidentwstest.ExpectSessionRevoked(t, concurrencySocket, authn.ConcurrencyLimitReasonCode)

		memberSocket := incidentwstest.ConnectAndHello(t, harness.Server.HTTP.URL, incidentID, incidentwstest.ConnectOptions{
			SessionToken:     memberSession.Value,
			ClientInstanceID: "phase6-i-6-01-membership",
			Presence:         timelinePresence(),
		})
		phase3test.DeleteMembership(t, harness.Server, incidentID, member.ID.String(), queryMembershipVersion(t, harness, incidentID, member.ID.String()), admin)
		incidentwstest.ExpectSessionRevoked(t, memberSocket, "incident_access_revoked")

		expirySocket := incidentwstest.ConnectAndHello(t, harness.Server.HTTP.URL, incidentID, incidentwstest.ConnectOptions{
			SessionToken:     expirySession.Value,
			ClientInstanceID: "phase6-i-6-01-expiry",
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

func TestPhase6_ResumeReplaysReplayableMessagesOnly_I_6_02(t *testing.T) {
	runtime := phase3test.StartRuntime(t)
	harness, admin, incidentID := setupPhase6SocketIncident(t, runtime, "phase6-i-6-02-replayable-only")

	source := incidentwstest.ConnectAndHello(t, harness.Server.HTTP.URL, incidentID, incidentwstest.ConnectOptions{
		SessionToken:     admin.SessionCookie.Value,
		ClientInstanceID: "phase6-i-6-02-source",
		Presence:         timelinePresence(),
	})
	resumeToken := source.HelloAck.ResumeToken
	source.Close(websocket.StatusNormalClosure, "test_complete")

	other := incidentwstest.ConnectAndHello(t, harness.Server.HTTP.URL, incidentID, incidentwstest.ConnectOptions{
		SessionToken:     admin.SessionCookie.Value,
		ClientInstanceID: "phase6-i-6-02-other",
		Presence:         timelinePresence(),
	})
	defer other.Close(websocket.StatusNormalClosure, "test_complete")

	phase3test.CreateTimelineRow(t, harness.Server, incidentID, admin, map[string]any{
		"client_txn_id":    "txn-phase6-i-6-02-record",
		"timeline.summary": "Phase 6 replayable record change",
	})
	publishPhase6JobProgress(t, harness, incidentID, "phase6-i-6-02-job", platformws.JobStatusRunning)

	resumed := incidentwstest.ConnectAndResume(t, harness.Server.HTTP.URL, incidentID, incidentwstest.ConnectOptions{
		SessionToken:     admin.SessionCookie.Value,
		ClientInstanceID: "phase6-i-6-02-source",
		Presence:         timelinePresence(),
	}, resumeToken, 0)
	defer resumed.Close(websocket.StatusNormalClosure, "test_complete")

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

func TestPhase6_CookieSocketRejectsUntrustedOrigin_I_6_04(t *testing.T) {
	runtime := phase3test.StartRuntime(t)
	harness, admin, incidentID := setupPhase6SocketIncident(t, runtime, "phase6-i-6-04-origin")

	incidentwstest.RequireDialRejectedStatus(t, harness.Server.HTTP.URL, incidentID, incidentwstest.ConnectOptions{
		Cookies: []*http.Cookie{admin.SessionCookie},
		Origin:  "https://untrusted.example.test",
	}, http.StatusForbidden)

	client := incidentwstest.ConnectAndHello(t, harness.Server.HTTP.URL, incidentID, incidentwstest.ConnectOptions{
		SessionToken:     admin.SessionCookie.Value,
		ClientInstanceID: "phase6-i-6-04-authorized",
		Presence:         timelinePresence(),
	})
	defer client.Close(websocket.StatusNormalClosure, "test_complete")
	if got := len(client.PresenceSnapshot); got != 1 {
		t.Fatalf("rejected origin must not subscribe before a valid client connects, snapshot len=%d snapshot=%#v", got, client.PresenceSnapshot)
	}
}

func setupPhase6SocketIncidentWithAdminID(t testing.TB, runtime *phase3test.RuntimeHarness, prefix string) (*phase3test.ServerHarness, phase3test.LoginResult, uuid.UUID, string) {
	t.Helper()

	harness := runtime.StartServer(t, prefix)
	admin, adminID := phase3test.ProvisionBootstrapAdmin(t, harness.Server)
	incident := phase3test.CreateIncident(t, harness.Server, admin, map[string]any{
		"client_txn_id": "txn-" + prefix,
		"incident_key":  "IR-" + prefix,
		"title":         prefix,
	})
	return harness, admin, adminID, incident["incident_id"].(string)
}

func publishPhase6JobProgress(t testing.TB, harness *phase3test.ServerHarness, incidentID string, jobID string, status string) {
	t.Helper()

	parsedIncidentID := phase3test.MustUUID(t, incidentID)
	total := int64(2)
	payload := platformws.NewIncidentJobProgressPayload(jobID, parsedIncidentID, status, platformws.JobProgress{
		Completed: 1,
		Total:     &total,
	}, time.Now().UTC())
	if err := harness.Server.Runtime.WSHub.PublishJobProgress(parsedIncidentID, payload); err != nil {
		t.Fatalf("publish job progress: %v", err)
	}
}

func queryMembershipVersion(t testing.TB, harness *phase3test.ServerHarness, incidentID string, userID string) int64 {
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
