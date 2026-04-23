package auth

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/authcookietest"
	"github.com/JochiRaider/cartulary/internal/testutil/phase1storetest"
)

type phase1StoreFixture struct {
	harness *phase1storetest.StoreHarness
	store   *authn.Store
	hub     *hubStub
	service *Service
	keys    authn.MasterKeys
	now     time.Time
}

func newPhase1StoreFixture(t testing.TB, prefix string, now time.Time) *phase1StoreFixture {
	t.Helper()

	harness := phase1storetest.StartStore(t, prefix)
	keys := loadUnitMasterKeys(t)
	store := authn.NewStore(harness.Pool)
	hub := &hubStub{}

	return &phase1StoreFixture{
		harness: harness,
		store:   store,
		hub:     hub,
		service: newUnitService(t, store, hub, keys, now),
		keys:    keys,
		now:     now,
	}
}

func TestPhase1_ConcurrencyLimitRevokesLRUNonCurrent_U_1_05(t *testing.T) {
	now := time.Date(2026, time.April, 17, 12, 18, 0, 0, time.UTC)
	fixture := newPhase1StoreFixture(t, "phase1-u-1-05", now)

	user := phase1storetest.SeedLocalUserRecord(
		t,
		fixture.harness.DB,
		"concurrency@example.test",
		"Concurrency",
		"Phase1ConcurrencyPass!",
		false,
		false,
		true,
	)

	victim := phase1storetest.SeedSession(t, fixture.harness.DB, fixture.keys, user.ID, "phase1-u-1-05-victim", now.Add(-5*time.Hour), now.Add(-10*time.Minute))
	tieLaterAuth := phase1storetest.SeedSession(t, fixture.harness.DB, fixture.keys, user.ID, "phase1-u-1-05-tie-later-auth", now.Add(-4*time.Hour), now.Add(-10*time.Minute))
	phase1storetest.SeedSession(t, fixture.harness.DB, fixture.keys, user.ID, "phase1-u-1-05-recent-a", now.Add(-3*time.Hour), now.Add(-9*time.Minute))
	phase1storetest.SeedSession(t, fixture.harness.DB, fixture.keys, user.ID, "phase1-u-1-05-recent-b", now.Add(-2*time.Hour), now.Add(-8*time.Minute))
	phase1storetest.SeedSession(t, fixture.harness.DB, fixture.keys, user.ID, "phase1-u-1-05-recent-c", now.Add(-time.Hour), now.Add(-7*time.Minute))

	recorder := httptest.NewRecorder()
	request := newJSONRequest(t, http.MethodPost, "/api/v1/auth/login", `{
		"username":" concurrency@example.test ",
		"password":"Phase1ConcurrencyPass!"
	}`)
	fixture.service.handleLogin(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected login status: got %d want %d", recorder.Code, http.StatusOK)
	}
	authcookietest.RequireAuthCookies(t, recorder.Result().Cookies())
	requireRevocations(t, fixture.hub.revocations, []revocationCall{
		{sessionID: victim.ID, reasonCode: authn.ConcurrencyLimitReasonCode},
	})

	data := decodeSuccessData(t, recorder)
	if got := data["user_id"]; got != user.ID.String() {
		t.Fatalf("unexpected session user_id: got %v want %s", got, user.ID)
	}

	requireSessionRevoked(t, fixture.harness.DB, victim.ID, now, authn.ConcurrencyLimitReasonCode)
	requireSessionActive(t, fixture.harness.DB, tieLaterAuth.ID)
	created := phase1storetest.QuerySessionRow(t, fixture.harness.DB, user.ID.String())
	if created.SessionID == victim.ID.String() {
		t.Fatalf("expected a fresh login session, got the revoked victim row %#v", created)
	}
	if created.RevokedAt.Valid || created.RevokeReasonCode.Valid {
		t.Fatalf("expected the newly created session to remain active, got %#v", created)
	}

	auditCount := phase1storetest.QueryCount(
		t,
		fixture.harness.DB,
		`SELECT COUNT(*) FROM deployment_admin_audit_events WHERE target_user_id = $1 AND reason_code = $2`,
		user.ID,
		authn.ConcurrencyLimitReasonCode,
	)
	if auditCount != 1 {
		t.Fatalf("expected one concurrency-limit audit event, got %d", auditCount)
	}
	event := phase1storetest.RequireAuditEventBySource(t, phase1storetest.LookupUserAuditEvents(t, fixture.harness.DB, user.ID.String()), "auth.login")
	if event.EventKind != "session_revoked" || event.ReasonCode != authn.ConcurrencyLimitReasonCode {
		t.Fatalf("unexpected concurrency-limit audit event: %#v", event)
	}
	if got := event.Before["session_id"]; got != victim.ID.String() {
		t.Fatalf("unexpected revoked audit victim: got %v want %s", got, victim.ID)
	}
	if got := event.After["reason_code"]; got != authn.ConcurrencyLimitReasonCode {
		t.Fatalf("unexpected revoked audit reason_code: got %v want %s", got, authn.ConcurrencyLimitReasonCode)
	}
}

func TestPhase1_RouteRevocationConsequences_U_1_06(t *testing.T) {
	now := time.Date(2026, time.April, 17, 12, 30, 0, 0, time.UTC)

	for _, tc := range []struct {
		name string
		run  func(*testing.T, *phase1StoreFixture)
	}{
		{
			name: "logout revokes only the current session",
			run: func(t *testing.T, fixture *phase1StoreFixture) {
				user := phase1storetest.SeedLocalUserRecord(
					t,
					fixture.harness.DB,
					"logout@example.test",
					"Logout",
					"Phase1LogoutPass!",
					false,
					false,
					true,
				)
				current := phase1storetest.SeedSession(t, fixture.harness.DB, fixture.keys, user.ID, "phase1-u-1-06-logout-current", fixture.now.Add(-2*time.Hour), fixture.now.Add(-5*time.Minute))
				other := phase1storetest.SeedSession(t, fixture.harness.DB, fixture.keys, user.ID, "phase1-u-1-06-logout-other", fixture.now.Add(-time.Hour), fixture.now.Add(-4*time.Minute))

				recorder := httptest.NewRecorder()
				request := newJSONRequest(t, http.MethodPost, "/api/v1/auth/logout", `{}`)
				addSessionAuth(request, fixture.keys, "phase1-u-1-06-logout-current", true)
				fixture.service.handleLogout(recorder, request)

				if recorder.Code != http.StatusOK {
					t.Fatalf("unexpected logout status: got %d want %d", recorder.Code, http.StatusOK)
				}
				requireRevocations(t, fixture.hub.revocations, []revocationCall{
					{sessionID: current.ID, reasonCode: "session_revoked"},
				})
				requireSessionRevoked(t, fixture.harness.DB, current.ID, fixture.now, "session_revoked")
				requireSessionActive(t, fixture.harness.DB, other.ID)
				data := decodeSuccessData(t, recorder)
				if data["logged_out"] != true || data["sessions_revoked"] != false {
					t.Fatalf("unexpected logout response: %#v", data)
				}
				requireCookieCleared(t, recorder.Result().Cookies(), authn.SessionCookieName)
				requireCookieCleared(t, recorder.Result().Cookies(), authn.CSRFCookieName)
			},
		},
		{
			name: "password change revokes every active session",
			run: func(t *testing.T, fixture *phase1StoreFixture) {
				secretBase32 := authn.EncodeSecretBase32([]byte("01234567890123456789"))
				user := phase1storetest.SeedLocalUserWithActiveTOTPRecord(
					t,
					fixture.harness.DB,
					"password-change@example.test",
					"Password Change",
					"Phase1PasswordCurrent!",
					true,
					false,
					secretBase32,
				)
				current := phase1storetest.SeedSession(t, fixture.harness.DB, fixture.keys, user.ID, "phase1-u-1-06-password-current", fixture.now.Add(-2*time.Hour), fixture.now.Add(-5*time.Minute))
				other := phase1storetest.SeedSession(t, fixture.harness.DB, fixture.keys, user.ID, "phase1-u-1-06-password-other", fixture.now.Add(-time.Hour), fixture.now.Add(-4*time.Minute))
				code := generateTOTPCodeAt(t, secretBase32, fixture.now)

				recorder := httptest.NewRecorder()
				request := newJSONRequest(t, http.MethodPost, "/api/v1/auth/password/change", `{
					"client_txn_id":"txn-phase1-u-1-06-password-change",
					"current_password":"Phase1PasswordCurrent!",
					"new_password":"Phase1PasswordChanged!",
					"second_factor":{"kind":"totp","assertion":{"code":"`+code+`"}}
				}`)
				addSessionAuth(request, fixture.keys, "phase1-u-1-06-password-current", true)
				fixture.service.handlePasswordChange(recorder, request)

				if recorder.Code != http.StatusOK {
					t.Fatalf("unexpected password-change status: got %d want %d", recorder.Code, http.StatusOK)
				}
				requireRevocations(t, fixture.hub.revocations, []revocationCall{
					{sessionID: current.ID, reasonCode: "session_revoked"},
					{sessionID: other.ID, reasonCode: "session_revoked"},
				})
				requireSessionRevoked(t, fixture.harness.DB, current.ID, fixture.now, "session_revoked")
				requireSessionRevoked(t, fixture.harness.DB, other.ID, fixture.now, "session_revoked")
				if data := decodeSuccessData(t, recorder); data["sessions_revoked"] != true {
					t.Fatalf("unexpected password-change response: %#v", data)
				}
				requireCookieCleared(t, recorder.Result().Cookies(), authn.SessionCookieName)
				requireCookieCleared(t, recorder.Result().Cookies(), authn.CSRFCookieName)
			},
		},
		{
			name: "totp replacement completion revokes every active session",
			run: func(t *testing.T, fixture *phase1StoreFixture) {
				currentSecretBase32 := authn.EncodeSecretBase32([]byte("01234567890123456789"))
				user := phase1storetest.SeedLocalUserWithActiveTOTPRecord(
					t,
					fixture.harness.DB,
					"totp-replacement@example.test",
					"TOTP Replacement",
					"Phase1TOTPReplacementPass!",
					true,
					false,
					currentSecretBase32,
				)
				current := phase1storetest.SeedSession(t, fixture.harness.DB, fixture.keys, user.ID, "phase1-u-1-06-totp-current", fixture.now.Add(-2*time.Hour), fixture.now.Add(-5*time.Minute))
				other := phase1storetest.SeedSession(t, fixture.harness.DB, fixture.keys, user.ID, "phase1-u-1-06-totp-other", fixture.now.Add(-time.Hour), fixture.now.Add(-4*time.Minute))
				replacementSecret := []byte("abcdefghij0123456789")
				enrollment := phase1storetest.SeedPendingTOTPEnrollment(
					t,
					fixture.harness.DB,
					fixture.keys,
					user.ID,
					&current.ID,
					nil,
					"txn-phase1-u-1-06-totp-begin",
					replacementSecret,
					true,
					fixture.now.Add(-time.Minute),
				)
				code := generateTOTPCodeAt(t, authn.EncodeSecretBase32(replacementSecret), fixture.now)

				recorder := httptest.NewRecorder()
				request := newJSONRequest(t, http.MethodPost, "/api/v1/auth/mfa/totp/complete", `{
					"client_txn_id":"txn-phase1-u-1-06-totp-complete",
					"enrollment_id":"`+enrollment.ID.String()+`",
					"code":"`+code+`"
				}`)
				addSessionAuth(request, fixture.keys, "phase1-u-1-06-totp-current", true)
				fixture.service.handleTOTPComplete(recorder, request)

				if recorder.Code != http.StatusOK {
					t.Fatalf("unexpected totp-complete status: got %d want %d", recorder.Code, http.StatusOK)
				}
				requireRevocations(t, fixture.hub.revocations, []revocationCall{
					{sessionID: current.ID, reasonCode: "session_revoked"},
					{sessionID: other.ID, reasonCode: "session_revoked"},
				})
				requireSessionRevoked(t, fixture.harness.DB, current.ID, fixture.now, "session_revoked")
				requireSessionRevoked(t, fixture.harness.DB, other.ID, fixture.now, "session_revoked")
				if data := decodeSuccessData(t, recorder); data["sessions_revoked"] != true {
					t.Fatalf("unexpected totp-complete response: %#v", data)
				}
				requireCookieCleared(t, recorder.Result().Cookies(), authn.SessionCookieName)
				requireCookieCleared(t, recorder.Result().Cookies(), authn.CSRFCookieName)
			},
		},
		{
			name: "account disablement revokes every active session and clears the current cookie when needed",
			run: func(t *testing.T, fixture *phase1StoreFixture) {
				admin := phase1storetest.SeedLocalUserRecord(
					t,
					fixture.harness.DB,
					"disablement-admin@example.test",
					"Disablement Admin",
					"Phase1DisablementPass!",
					false,
					true,
					true,
				)
				phase1storetest.SeedLocalUserRecord(
					t,
					fixture.harness.DB,
					"disablement-second-admin@example.test",
					"Disablement Second Admin",
					"Phase1DisablementSecondPass!",
					false,
					true,
					true,
				)
				current := phase1storetest.SeedSession(t, fixture.harness.DB, fixture.keys, admin.ID, "phase1-u-1-06-disable-current", fixture.now.Add(-2*time.Hour), fixture.now.Add(-5*time.Minute))
				other := phase1storetest.SeedSession(t, fixture.harness.DB, fixture.keys, admin.ID, "phase1-u-1-06-disable-other", fixture.now.Add(-time.Hour), fixture.now.Add(-4*time.Minute))

				recorder := httptest.NewRecorder()
				request := newJSONRequest(t, http.MethodPatch, "/api/v1/users/"+admin.ID.String(), `{
					"base_user_version":1,
					"is_active":false
				}`)
				addSessionAuth(request, fixture.keys, "phase1-u-1-06-disable-current", true)
				fixture.service.handleUsersMember(recorder, request)

				if recorder.Code != http.StatusOK {
					t.Fatalf("unexpected disablement status: got %d want %d", recorder.Code, http.StatusOK)
				}
				requireRevocations(t, fixture.hub.revocations, []revocationCall{
					{sessionID: current.ID, reasonCode: "session_revoked"},
					{sessionID: other.ID, reasonCode: "session_revoked"},
				})
				requireSessionRevoked(t, fixture.harness.DB, current.ID, fixture.now, "session_revoked")
				requireSessionRevoked(t, fixture.harness.DB, other.ID, fixture.now, "session_revoked")
				if data := decodeSuccessData(t, recorder); data["is_active"] != false {
					t.Fatalf("unexpected disablement response: %#v", data)
				}
				requireCookieCleared(t, recorder.Result().Cookies(), authn.SessionCookieName)
				requireCookieCleared(t, recorder.Result().Cookies(), authn.CSRFCookieName)
			},
		},
		{
			name: "admin password reset revokes every target session and leaves the actor cookie alone",
			run: func(t *testing.T, fixture *phase1StoreFixture) {
				admin := phase1storetest.SeedLocalUserRecord(
					t,
					fixture.harness.DB,
					"reset-admin@example.test",
					"Reset Admin",
					"Phase1ResetAdminPass!",
					false,
					true,
					true,
				)
				target := phase1storetest.SeedLocalUserRecord(
					t,
					fixture.harness.DB,
					"reset-target@example.test",
					"Reset Target",
					"Phase1ResetTargetPass!",
					false,
					false,
					true,
				)
				actorSession := phase1storetest.SeedSession(t, fixture.harness.DB, fixture.keys, admin.ID, "phase1-u-1-06-reset-admin", fixture.now.Add(-2*time.Hour), fixture.now.Add(-5*time.Minute))
				targetCurrent := phase1storetest.SeedSession(t, fixture.harness.DB, fixture.keys, target.ID, "phase1-u-1-06-reset-target-current", fixture.now.Add(-time.Hour), fixture.now.Add(-4*time.Minute))
				targetOther := phase1storetest.SeedSession(t, fixture.harness.DB, fixture.keys, target.ID, "phase1-u-1-06-reset-target-other", fixture.now.Add(-45*time.Minute), fixture.now.Add(-3*time.Minute))

				recorder := httptest.NewRecorder()
				request := newJSONRequest(t, http.MethodPost, "/api/v1/users/"+target.ID.String()+"/password/reset", `{
					"base_user_version":1,
					"client_txn_id":"txn-phase1-u-1-06-password-reset",
					"new_password":"Phase1ResetTargetChanged!"
				}`)
				addSessionAuth(request, fixture.keys, "phase1-u-1-06-reset-admin", true)
				fixture.service.handleUsersMember(recorder, request)

				if recorder.Code != http.StatusOK {
					t.Fatalf("unexpected admin password-reset status: got %d want %d", recorder.Code, http.StatusOK)
				}
				requireRevocations(t, fixture.hub.revocations, []revocationCall{
					{sessionID: targetCurrent.ID, reasonCode: "session_revoked"},
					{sessionID: targetOther.ID, reasonCode: "session_revoked"},
				})
				requireSessionActive(t, fixture.harness.DB, actorSession.ID)
				requireSessionRevoked(t, fixture.harness.DB, targetCurrent.ID, fixture.now, "session_revoked")
				requireSessionRevoked(t, fixture.harness.DB, targetOther.ID, fixture.now, "session_revoked")
				if data := decodeSuccessData(t, recorder); data["user_id"] != target.ID.String() {
					t.Fatalf("unexpected admin password-reset response: %#v", data)
				}
				requireNoAuthCookieMutation(t, recorder.Result().Cookies())
			},
		},
		{
			name: "admin totp reset revokes every target session and leaves the actor cookie alone",
			run: func(t *testing.T, fixture *phase1StoreFixture) {
				admin := phase1storetest.SeedLocalUserRecord(
					t,
					fixture.harness.DB,
					"totp-reset-admin@example.test",
					"TOTP Reset Admin",
					"Phase1TOTPResetAdminPass!",
					false,
					true,
					true,
				)
				targetSecretBase32 := authn.EncodeSecretBase32([]byte("jihgfedcba9876543210"))
				target := phase1storetest.SeedLocalUserWithActiveTOTPRecord(
					t,
					fixture.harness.DB,
					"totp-reset-target@example.test",
					"TOTP Reset Target",
					"Phase1TOTPResetTargetPass!",
					true,
					false,
					targetSecretBase32,
				)
				actorSession := phase1storetest.SeedSession(t, fixture.harness.DB, fixture.keys, admin.ID, "phase1-u-1-06-totp-reset-admin", fixture.now.Add(-2*time.Hour), fixture.now.Add(-5*time.Minute))
				targetCurrent := phase1storetest.SeedSession(t, fixture.harness.DB, fixture.keys, target.ID, "phase1-u-1-06-totp-reset-target-current", fixture.now.Add(-time.Hour), fixture.now.Add(-4*time.Minute))
				targetOther := phase1storetest.SeedSession(t, fixture.harness.DB, fixture.keys, target.ID, "phase1-u-1-06-totp-reset-target-other", fixture.now.Add(-45*time.Minute), fixture.now.Add(-3*time.Minute))

				recorder := httptest.NewRecorder()
				request := newJSONRequest(t, http.MethodPost, "/api/v1/users/"+target.ID.String()+"/mfa/totp/reset", `{
					"base_user_version":1,
					"client_txn_id":"txn-phase1-u-1-06-totp-reset"
				}`)
				addSessionAuth(request, fixture.keys, "phase1-u-1-06-totp-reset-admin", true)
				fixture.service.handleUsersMember(recorder, request)

				if recorder.Code != http.StatusOK {
					t.Fatalf("unexpected admin totp-reset status: got %d want %d", recorder.Code, http.StatusOK)
				}
				requireRevocations(t, fixture.hub.revocations, []revocationCall{
					{sessionID: targetCurrent.ID, reasonCode: "session_revoked"},
					{sessionID: targetOther.ID, reasonCode: "session_revoked"},
				})
				requireSessionActive(t, fixture.harness.DB, actorSession.ID)
				requireSessionRevoked(t, fixture.harness.DB, targetCurrent.ID, fixture.now, "session_revoked")
				requireSessionRevoked(t, fixture.harness.DB, targetOther.ID, fixture.now, "session_revoked")
				if data := decodeSuccessData(t, recorder); data["user_id"] != target.ID.String() {
					t.Fatalf("unexpected admin totp-reset response: %#v", data)
				}
				requireNoAuthCookieMutation(t, recorder.Result().Cookies())
			},
		},
		{
			name: "explicit revoke all revokes every target session and leaves the actor cookie alone",
			run: func(t *testing.T, fixture *phase1StoreFixture) {
				admin := phase1storetest.SeedLocalUserRecord(
					t,
					fixture.harness.DB,
					"revoke-all-admin@example.test",
					"Revoke All Admin",
					"Phase1RevokeAllAdminPass!",
					false,
					true,
					true,
				)
				target := phase1storetest.SeedLocalUserRecord(
					t,
					fixture.harness.DB,
					"revoke-all-target@example.test",
					"Revoke All Target",
					"Phase1RevokeAllTargetPass!",
					false,
					false,
					true,
				)
				actorSession := phase1storetest.SeedSession(t, fixture.harness.DB, fixture.keys, admin.ID, "phase1-u-1-06-revoke-all-admin", fixture.now.Add(-2*time.Hour), fixture.now.Add(-5*time.Minute))
				targetCurrent := phase1storetest.SeedSession(t, fixture.harness.DB, fixture.keys, target.ID, "phase1-u-1-06-revoke-all-target-current", fixture.now.Add(-time.Hour), fixture.now.Add(-4*time.Minute))
				targetOther := phase1storetest.SeedSession(t, fixture.harness.DB, fixture.keys, target.ID, "phase1-u-1-06-revoke-all-target-other", fixture.now.Add(-45*time.Minute), fixture.now.Add(-3*time.Minute))

				recorder := httptest.NewRecorder()
				request := newJSONRequest(t, http.MethodPost, "/api/v1/users/"+target.ID.String()+"/sessions/revoke-all", `{
					"client_txn_id":"txn-phase1-u-1-06-revoke-all",
					"reason":"incident response"
				}`)
				addSessionAuth(request, fixture.keys, "phase1-u-1-06-revoke-all-admin", true)
				fixture.service.handleUsersMember(recorder, request)

				if recorder.Code != http.StatusOK {
					t.Fatalf("unexpected revoke-all status: got %d want %d", recorder.Code, http.StatusOK)
				}
				requireRevocations(t, fixture.hub.revocations, []revocationCall{
					{sessionID: targetCurrent.ID, reasonCode: "session_revoked"},
					{sessionID: targetOther.ID, reasonCode: "session_revoked"},
				})
				requireSessionActive(t, fixture.harness.DB, actorSession.ID)
				requireSessionRevoked(t, fixture.harness.DB, targetCurrent.ID, fixture.now, "session_revoked")
				requireSessionRevoked(t, fixture.harness.DB, targetOther.ID, fixture.now, "session_revoked")
				if data := decodeSuccessData(t, recorder); data["sessions_revoked"] != true {
					t.Fatalf("unexpected revoke-all response: %#v", data)
				}
				requireNoAuthCookieMutation(t, recorder.Result().Cookies())
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			tc.run(t, newPhase1StoreFixture(t, "phase1-u-1-06-"+slugifyTestName(tc.name), now))
		})
	}
}

func TestPhase1_UserPatchAndLastAdminGuard_U_1_08(t *testing.T) {
	now := time.Date(2026, time.April, 17, 12, 35, 0, 0, time.UTC)

	t.Run("rejects missing base_user_version before store mutation", func(t *testing.T) {
		fixture := newPhase1StoreFixture(t, "phase1-u-1-08-missing-version", now)
		admin := phase1storetest.SeedLocalUserRecord(
			t,
			fixture.harness.DB,
			"patch-admin@example.test",
			"Patch Admin",
			"Phase1PatchAdminPass!",
			false,
			true,
			true,
		)
		target := phase1storetest.SeedLocalUserRecord(
			t,
			fixture.harness.DB,
			"patch-target@example.test",
			"Patch Target",
			"Phase1PatchTargetPass!",
			false,
			false,
			true,
		)
		phase1storetest.SeedSession(t, fixture.harness.DB, fixture.keys, admin.ID, "phase1-u-1-08-missing-version-admin", now.Add(-time.Hour), now.Add(-5*time.Minute))

		recorder := httptest.NewRecorder()
		request := newJSONRequest(t, http.MethodPatch, "/api/v1/users/"+target.ID.String(), `{
			"display_name":"Missing Version"
		}`)
		addSessionAuth(request, fixture.keys, "phase1-u-1-08-missing-version-admin", true)
		fixture.service.handleUsersMember(recorder, request)

		requireErrorEnvelope(t, recorder, http.StatusBadRequest, "invalid_mutation_payload", "base_user_version")
		if len(fixture.hub.revocations) != 0 {
			t.Fatalf("missing base_user_version must not publish revocations, got %#v", fixture.hub.revocations)
		}
		after := mustUserRecord(t, fixture.store, target.ID)
		if after.UserVersion != target.UserVersion || after.DisplayName != target.DisplayName {
			t.Fatalf("missing base_user_version must not mutate the target user: before=%#v after=%#v", target, after)
		}
	})

	t.Run("normalizes writable strings and allows demotion when another active deployment admin remains", func(t *testing.T) {
		fixture := newPhase1StoreFixture(t, "phase1-u-1-08-normalized-success", now)
		admin := phase1storetest.SeedLocalUserRecord(
			t,
			fixture.harness.DB,
			"patch-actor@example.test",
			"Patch Actor",
			"Phase1PatchActorPass!",
			false,
			true,
			true,
		)
		target := phase1storetest.SeedLocalUserRecord(
			t,
			fixture.harness.DB,
			"patch-demote@example.test",
			"Patch Demote",
			"Phase1PatchDemotePass!",
			false,
			true,
			true,
		)
		phase1storetest.SeedSession(t, fixture.harness.DB, fixture.keys, admin.ID, "phase1-u-1-08-normalized-admin", now.Add(-time.Hour), now.Add(-5*time.Minute))

		recorder := httptest.NewRecorder()
		request := newJSONRequest(t, http.MethodPatch, "/api/v1/users/"+target.ID.String(), `{
			"base_user_version":1,
			"email":" Phase1Patched@Example.Test ",
			"display_name":" Analyst Patched ",
			"is_deployment_admin":false
		}`)
		addSessionAuth(request, fixture.keys, "phase1-u-1-08-normalized-admin", true)
		fixture.service.handleUsersMember(recorder, request)

		if recorder.Code != http.StatusOK {
			t.Fatalf("unexpected patch status: got %d want %d", recorder.Code, http.StatusOK)
		}
		if len(fixture.hub.revocations) != 0 {
			t.Fatalf("pure normalization and demotion must not publish revocations, got %#v", fixture.hub.revocations)
		}
		data := decodeSuccessData(t, recorder)
		if data["email"] != "Phase1Patched@Example.Test" || data["display_name"] != "Analyst Patched" || data["is_deployment_admin"] != false {
			t.Fatalf("unexpected normalized patch response: %#v", data)
		}
		after := mustUserRecord(t, fixture.store, target.ID)
		if after.Email != "Phase1Patched@Example.Test" || after.DisplayName != "Analyst Patched" || after.IsDeploymentAdmin {
			t.Fatalf("unexpected normalized patch persistence: %#v", after)
		}
		if after.UserVersion != target.UserVersion+1 {
			t.Fatalf("expected user_version to advance: before=%d after=%d", target.UserVersion, after.UserVersion)
		}
	})

	t.Run("translates stale version conflicts to user_version_conflict without durable mutation", func(t *testing.T) {
		fixture := newPhase1StoreFixture(t, "phase1-u-1-08-stale-version", now)
		admin := phase1storetest.SeedLocalUserRecord(
			t,
			fixture.harness.DB,
			"patch-stale-admin@example.test",
			"Patch Stale Admin",
			"Phase1PatchStaleAdminPass!",
			false,
			true,
			true,
		)
		target := phase1storetest.SeedLocalUserRecord(
			t,
			fixture.harness.DB,
			"patch-stale-target@example.test",
			"Patch Stale Target",
			"Phase1PatchStaleTargetPass!",
			false,
			false,
			true,
		)
		phase1storetest.SeedSession(t, fixture.harness.DB, fixture.keys, admin.ID, "phase1-u-1-08-stale-admin", now.Add(-time.Hour), now.Add(-5*time.Minute))

		recorder := httptest.NewRecorder()
		request := newJSONRequest(t, http.MethodPatch, "/api/v1/users/"+target.ID.String(), `{
			"base_user_version":7,
			"email":" Phase1Stale@Example.Test ",
			"display_name":" Phase 1 Stale "
		}`)
		addSessionAuth(request, fixture.keys, "phase1-u-1-08-stale-admin", true)
		fixture.service.handleUsersMember(recorder, request)

		requireErrorEnvelope(t, recorder, http.StatusConflict, "user_version_conflict", "")
		if len(fixture.hub.revocations) != 0 {
			t.Fatalf("stale version conflicts must not publish revocations, got %#v", fixture.hub.revocations)
		}
		after := mustUserRecord(t, fixture.store, target.ID)
		if after.UserVersion != target.UserVersion || after.Email != target.Email || after.DisplayName != target.DisplayName {
			t.Fatalf("stale version conflicts must not mutate the target user: before=%#v after=%#v", target, after)
		}
	})

	t.Run("rejects demotion of the last active deployment admin without durable mutation", func(t *testing.T) {
		fixture := newPhase1StoreFixture(t, "phase1-u-1-08-last-admin-demotion", now)
		admin := phase1storetest.SeedLocalUserRecord(
			t,
			fixture.harness.DB,
			"patch-last-admin@example.test",
			"Patch Last Admin",
			"Phase1PatchLastAdminPass!",
			false,
			true,
			true,
		)
		phase1storetest.SeedLocalUserRecord(
			t,
			fixture.harness.DB,
			"patch-last-admin-inactive@example.test",
			"Patch Last Admin Inactive",
			"Phase1PatchLastAdminInactivePass!",
			false,
			true,
			false,
		)
		phase1storetest.SeedSession(t, fixture.harness.DB, fixture.keys, admin.ID, "phase1-u-1-08-last-admin-demotion", now.Add(-time.Hour), now.Add(-5*time.Minute))

		recorder := httptest.NewRecorder()
		request := newJSONRequest(t, http.MethodPatch, "/api/v1/users/"+admin.ID.String(), `{
			"base_user_version":1,
			"is_deployment_admin":false
		}`)
		addSessionAuth(request, fixture.keys, "phase1-u-1-08-last-admin-demotion", true)
		fixture.service.handleUsersMember(recorder, request)

		requireErrorEnvelope(t, recorder, http.StatusConflict, "last_deployment_admin", "")
		if len(fixture.hub.revocations) != 0 {
			t.Fatalf("last-admin demotion conflicts must not publish revocations, got %#v", fixture.hub.revocations)
		}
		after := mustUserRecord(t, fixture.store, admin.ID)
		if !after.IsDeploymentAdmin || !after.IsActive || after.UserVersion != admin.UserVersion {
			t.Fatalf("last-admin demotion conflicts must leave the admin unchanged: before=%#v after=%#v", admin, after)
		}
	})

	t.Run("rejects deactivation of the last active deployment admin without durable mutation", func(t *testing.T) {
		fixture := newPhase1StoreFixture(t, "phase1-u-1-08-last-admin-disable", now)
		admin := phase1storetest.SeedLocalUserRecord(
			t,
			fixture.harness.DB,
			"patch-last-disable@example.test",
			"Patch Last Disable",
			"Phase1PatchLastDisablePass!",
			false,
			true,
			true,
		)
		phase1storetest.SeedLocalUserRecord(
			t,
			fixture.harness.DB,
			"patch-last-disable-inactive@example.test",
			"Patch Last Disable Inactive",
			"Phase1PatchLastDisableInactivePass!",
			false,
			true,
			false,
		)
		phase1storetest.SeedSession(t, fixture.harness.DB, fixture.keys, admin.ID, "phase1-u-1-08-last-admin-disable", now.Add(-time.Hour), now.Add(-5*time.Minute))

		recorder := httptest.NewRecorder()
		request := newJSONRequest(t, http.MethodPatch, "/api/v1/users/"+admin.ID.String(), `{
			"base_user_version":1,
			"is_active":false
		}`)
		addSessionAuth(request, fixture.keys, "phase1-u-1-08-last-admin-disable", true)
		fixture.service.handleUsersMember(recorder, request)

		requireErrorEnvelope(t, recorder, http.StatusConflict, "last_deployment_admin", "")
		if len(fixture.hub.revocations) != 0 {
			t.Fatalf("last-admin disablement conflicts must not publish revocations, got %#v", fixture.hub.revocations)
		}
		after := mustUserRecord(t, fixture.store, admin.ID)
		if !after.IsDeploymentAdmin || !after.IsActive || after.UserVersion != admin.UserVersion {
			t.Fatalf("last-admin disablement conflicts must leave the admin unchanged: before=%#v after=%#v", admin, after)
		}
	})
}

func mustUserRecord(t testing.TB, store *authn.Store, userID uuid.UUID) authn.UserRecord {
	t.Helper()

	record, err := store.GetUserByID(context.Background(), userID)
	if err != nil {
		t.Fatalf("lookup user %s: %v", userID, err)
	}
	return record
}

func requireSessionRevoked(t testing.TB, db *sql.DB, sessionID uuid.UUID, revokedAt time.Time, reasonCode string) {
	t.Helper()

	row := phase1storetest.QuerySessionByID(t, db, sessionID.String())
	if !row.RevokedAt.Valid || !row.RevokedAt.Time.Equal(revokedAt) {
		t.Fatalf("expected session %s revoked_at=%s, got %#v", sessionID, revokedAt, row)
	}
	if !row.RevokeReasonCode.Valid || row.RevokeReasonCode.String != reasonCode {
		t.Fatalf("expected session %s revoke_reason_code=%q, got %#v", sessionID, reasonCode, row)
	}
}

func requireSessionActive(t testing.TB, db *sql.DB, sessionID uuid.UUID) {
	t.Helper()

	row := phase1storetest.QuerySessionByID(t, db, sessionID.String())
	if row.RevokedAt.Valid || row.RevokeReasonCode.Valid {
		t.Fatalf("expected session %s to remain active, got %#v", sessionID, row)
	}
}

func requireNoAuthCookieMutation(t testing.TB, cookies []*http.Cookie) {
	t.Helper()

	for _, cookie := range cookies {
		if cookie.Name == authn.SessionCookieName || cookie.Name == authn.CSRFCookieName {
			t.Fatalf("expected no auth-cookie mutation, got %#v", cookies)
		}
	}
}

func slugifyTestName(name string) string {
	result := make([]rune, 0, len(name))
	lastDash := false
	for _, r := range name {
		switch {
		case r >= 'a' && r <= 'z':
			result = append(result, r)
			lastDash = false
		case r >= 'A' && r <= 'Z':
			result = append(result, r+'a'-'A')
			lastDash = false
		case r >= '0' && r <= '9':
			result = append(result, r)
			lastDash = false
		case !lastDash:
			result = append(result, '-')
			lastDash = true
		}
	}
	if len(result) > 0 && result[len(result)-1] == '-' {
		result = result[:len(result)-1]
	}
	return string(result)
}
