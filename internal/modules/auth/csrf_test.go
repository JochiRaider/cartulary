package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/platform/authn"
)

func TestCookieMutationCSRF_Unit(t *testing.T) {
	now := time.Date(2026, time.October, 28, 12, 0, 0, 0, time.UTC)
	keys := loadUnitMasterKeys(t)
	userID := uuid.MustParse("10000000-0000-0000-0000-000000000124")
	sessionID := uuid.MustParse("10000000-0000-0000-0000-000000000125")
	const target = "/api/v1/users/10000000-0000-0000-0000-000000000126"
	for _, operation := range []struct {
		name, method, path, body string
		handler                  func(*Service, http.ResponseWriter, *http.Request)
		malformedBody            bool
	}{
		{"logout", http.MethodPost, "/api/v1/auth/logout", `{}`, (*Service).handleLogout, false},
		{"password change", http.MethodPost, "/api/v1/auth/password/change", `{"client_txn_id":"csrf-password","current_password":"CurrentPassword1!","new_password":"ChangedPassword1!"}`, (*Service).handlePasswordChange, true},
		{"totp begin", http.MethodPost, "/api/v1/auth/mfa/totp/begin", `{"client_txn_id":"csrf-begin"}`, (*Service).handleTOTPBegin, true},
		{"totp complete", http.MethodPost, "/api/v1/auth/mfa/totp/complete", `{"client_txn_id":"csrf-complete","enrollment_id":"10000000-0000-0000-0000-000000000127","code":"123456"}`, (*Service).handleTOTPComplete, true},
		{"user create", http.MethodPost, "/api/v1/users", `{"client_txn_id":"csrf-create","auth_kind":"local","email":"csrf@example.test","display_name":"CSRF User","initial_password":"InitialPassword1!"}`, (*Service).handleUsersCollection, false},
		{"user patch", http.MethodPatch, target, `{"base_user_version":1,"display_name":"Changed User"}`, (*Service).handleUsersMember, false},
		{"password reset", http.MethodPost, target + "/password/reset", `{"base_user_version":1,"client_txn_id":"csrf-reset-password","new_password":"ResetPassword1!"}`, (*Service).handleUsersMember, false},
		{"totp reset", http.MethodPost, target + "/mfa/totp/reset", `{"base_user_version":1,"client_txn_id":"csrf-reset-totp"}`, (*Service).handleUsersMember, false},
		{"revoke all", http.MethodPost, target + "/sessions/revoke-all", `{"client_txn_id":"csrf-revoke","reason":"revoke sessions"}`, (*Service).handleUsersMember, false},
	} {
		t.Run(operation.name, func(t *testing.T) {
			cases := []struct{ name, header, body string }{
				{"missing header", "", operation.body},
				{"invalid header", "wrong-csrf-token", operation.body},
			}
			if operation.malformedBody {
				cases = append(cases, struct{ name, header, body string }{"malformed body", "", `{`})
			}
			for _, requestCase := range cases {
				t.Run(requestCase.name, func(t *testing.T) {
					hub := &hubStub{}
					store := &authStoreStub{
						// Session reads are an implementation choice; harmful effects are not.
						getSessionByFingerprintFunc: func(context.Context, []byte) (authn.SessionRecord, authn.UserRecord, error) {
							return activeSessionRecord(sessionID, userID, now), deploymentAdminUserRecord(userID, "csrf-admin@example.test"), nil
						},
						revokeSessionFunc: func(context.Context, uuid.UUID, string, time.Time) error {
							t.Fatal("CSRF rejection must not revoke a session")
							return nil
						},
						changePasswordFunc: func(context.Context, authn.UserRecord, string, []byte, string, string, time.Time) (authn.PasswordChangeResult, error) {
							t.Fatal("CSRF rejection must not change a password")
							return authn.PasswordChangeResult{}, nil
						},
						beginTOTPEnrollmentFunc: func(context.Context, uuid.UUID, uuid.UUID, string, *uuid.UUID, *uuid.UUID, string, []byte, bool, time.Time) (authn.PendingTOTPEnrollmentRecord, bool, error) {
							t.Fatal("CSRF rejection must not create an enrollment")
							return authn.PendingTOTPEnrollmentRecord{}, false, nil
						},
						activateTOTPEnrollmentFunc: func(context.Context, authn.UserRecord, uuid.UUID, string, *uuid.UUID, *uuid.UUID, []byte, []byte, time.Time) (authn.TOTPCompleteResult, error) {
							t.Fatal("CSRF rejection must not activate a factor")
							return authn.TOTPCompleteResult{}, nil
						},
						createUserFunc: func(context.Context, authn.UserRecord, string, string, string, bool, bool, string, []byte, string, time.Time) (authn.UserCreateResult, error) {
							t.Fatal("CSRF rejection must not create a user")
							return authn.UserCreateResult{}, nil
						},
						updateUserFunc: func(context.Context, authn.UserRecord, uuid.UUID, int64, *string, *string, *bool, *bool, *bool, string, time.Time) (authn.UserRecord, []uuid.UUID, error) {
							t.Fatal("CSRF rejection must not update a user")
							return authn.UserRecord{}, nil, nil
						},
						adminResetPasswordFunc: func(context.Context, authn.UserRecord, uuid.UUID, int64, string, string, []byte, string, time.Time) (authn.AdminPasswordResetResult, error) {
							t.Fatal("CSRF rejection must not reset a password")
							return authn.AdminPasswordResetResult{}, nil
						},
						adminResetTOTPFunc: func(context.Context, authn.UserRecord, uuid.UUID, int64, string, []byte, string, time.Time) (authn.AdminTOTPResetResult, error) {
							t.Fatal("CSRF rejection must not reset a factor")
							return authn.AdminTOTPResetResult{}, nil
						},
						adminRevokeAllSessionsFunc: func(context.Context, authn.UserRecord, uuid.UUID, string, []byte, string, time.Time) (authn.AdminRevokeAllResult, error) {
							t.Fatal("CSRF rejection must not revoke user sessions")
							return authn.AdminRevokeAllResult{}, nil
						},
					}
					service := newUnitService(t, store, hub, keys, now)
					request := newJSONRequest(t, operation.method, operation.path, requestCase.body)
					addSessionCookiesOnly(request, keys, "csrf-cookie-session")
					if requestCase.header != "" {
						request.Header.Set(authn.CSRFHeaderName, requestCase.header)
					}
					recorder := httptest.NewRecorder()
					operation.handler(service, recorder, request)
					requireErrorEnvelope(t, recorder, http.StatusForbidden, "csrf_verification_failed", "")
					if len(hub.revocations) != 0 {
						t.Fatal("CSRF rejection must not publish revocations")
					}
				})
			}
		})
	}
}
