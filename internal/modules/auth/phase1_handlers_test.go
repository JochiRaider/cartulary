package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"

	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/pagination"
	"github.com/JochiRaider/cartulary/internal/testutil/authcookietest"
	phase1test "github.com/JochiRaider/cartulary/internal/testutil/phase1test/inventory"
)

func TestPhase1_LoginNormalizationAndPasswordExactness_U_1_02(t *testing.T) {
	now := time.Date(2026, time.April, 17, 11, 55, 0, 0, time.UTC)
	keys := loadUnitMasterKeys(t)
	userID := uuid.MustParse("10000000-0000-0000-0000-000000000091")
	sessionID := uuid.MustParse("10000000-0000-0000-0000-000000000092")
	password := "  Exact Login Secret  "
	passwordHash, err := authn.HashPassword(password)
	if err != nil {
		t.Fatalf("hash login password: %v", err)
	}

	t.Run("username trimming remains route-owned while password bytes stay exact", func(t *testing.T) {
		createCalls := 0
		store := &authStoreStub{
			getUserByNormalizedEmailFunc: func(_ context.Context, email string) (authn.UserRecord, error) {
				if email != "Analyst@Example.Test" {
					t.Fatalf("unexpected normalized email: got %q want Analyst@Example.Test", email)
				}
				return authn.UserRecord{
					ID:           userID,
					Email:        "Analyst@Example.Test",
					DisplayName:  "Analyst",
					PasswordHash: passwordHash,
					IsActive:     true,
				}, nil
			},
			createSessionWithConcurrencyFunc: func(_ context.Context, user authn.UserRecord, fingerprint []byte, timing authn.SessionTiming, requestID string) (authn.SessionRecord, *authn.SessionRecord, error) {
				createCalls++
				if user.ID != userID {
					t.Fatalf("unexpected login actor: got %s want %s", user.ID, userID)
				}
				if len(fingerprint) == 0 {
					t.Fatal("expected login fingerprint to be populated")
				}
				if requestID != "" {
					t.Fatalf("unexpected request id in direct handler test: got %q", requestID)
				}
				return activeSessionRecord(sessionID, userID, now), nil, nil
			},
			listIncidentMembershipSummariesFunc: func(_ context.Context, gotUserID uuid.UUID) ([]authn.IncidentMembershipSummary, error) {
				if gotUserID != userID {
					t.Fatalf("unexpected membership lookup user: got %s want %s", gotUserID, userID)
				}
				return nil, nil
			},
		}
		service := newUnitService(t, store, &hubStub{}, keys, now)

		recorder := httptest.NewRecorder()
		request := newJSONRequest(t, http.MethodPost, "/api/v1/auth/login", `{
			"username":"\u00a0Analyst@Example.Test\t",
			"password":"  Exact Login Secret  "
		}`)
		service.handleLogin(recorder, request)

		if recorder.Code != http.StatusOK {
			t.Fatalf("unexpected login status: got %d want %d", recorder.Code, http.StatusOK)
		}
		if createCalls != 1 {
			t.Fatalf("expected one session creation for exact password, got %d", createCalls)
		}
	})

	t.Run("password normalization would fail before session creation", func(t *testing.T) {
		createCalls := 0
		store := &authStoreStub{
			getUserByNormalizedEmailFunc: func(_ context.Context, email string) (authn.UserRecord, error) {
				if email != "Analyst@Example.Test" {
					t.Fatalf("unexpected normalized email: got %q want Analyst@Example.Test", email)
				}
				return authn.UserRecord{
					ID:           userID,
					Email:        "Analyst@Example.Test",
					DisplayName:  "Analyst",
					PasswordHash: passwordHash,
					IsActive:     true,
				}, nil
			},
			createSessionWithConcurrencyFunc: func(context.Context, authn.UserRecord, []byte, authn.SessionTiming, string) (authn.SessionRecord, *authn.SessionRecord, error) {
				createCalls++
				return authn.SessionRecord{}, nil, nil
			},
		}
		service := newUnitService(t, store, &hubStub{}, keys, now)

		recorder := httptest.NewRecorder()
		request := newJSONRequest(t, http.MethodPost, "/api/v1/auth/login", `{
			"username":" \u00a0Analyst@Example.Test\t",
			"password":"Exact Login Secret"
		}`)
		service.handleLogin(recorder, request)

		requireErrorEnvelope(t, recorder, http.StatusUnauthorized, "invalid_credentials", "")
		if createCalls != 0 {
			t.Fatalf("session creation must not run after exact-password mismatch, got %d calls", createCalls)
		}
	})
}

func TestPhase1_LoginCreatesSessionAndResource_U_1_03(t *testing.T) {
	now := time.Date(2026, time.April, 17, 12, 0, 0, 0, time.UTC)
	keys := loadUnitMasterKeys(t)

	userID := uuid.MustParse("10000000-0000-0000-0000-000000000101")
	sessionID := uuid.MustParse("10000000-0000-0000-0000-000000000102")
	incidentID := uuid.MustParse("10000000-0000-0000-0000-000000000103")
	password := "  Exact Login Secret  "
	passwordHash, err := authn.HashPassword(password)
	if err != nil {
		t.Fatalf("hash login password: %v", err)
	}

	var capturedTiming authn.SessionTiming
	createCalls := 0
	store := &authStoreStub{
		getUserByNormalizedEmailFunc: func(_ context.Context, email string) (authn.UserRecord, error) {
			if email != "analyst@example.test" {
				t.Fatalf("unexpected normalized login email: got %q", email)
			}
			return authn.UserRecord{
				ID:           userID,
				Email:        email,
				DisplayName:  "Analyst",
				PasswordHash: passwordHash,
				IsActive:     true,
			}, nil
		},
		createSessionWithConcurrencyFunc: func(_ context.Context, user authn.UserRecord, fingerprint []byte, timing authn.SessionTiming, requestID string) (authn.SessionRecord, *authn.SessionRecord, error) {
			createCalls++
			if user.ID != userID {
				t.Fatalf("login created session for wrong user: got %s want %s", user.ID, userID)
			}
			if len(fingerprint) == 0 {
				t.Fatal("session fingerprint must be populated")
			}
			if requestID != "" {
				t.Fatalf("unexpected request id in direct handler test: got %q", requestID)
			}
			capturedTiming = timing
			return authn.SessionRecord{
				ID:                       sessionID,
				UserID:                   userID,
				AuthenticatedAt:          timing.AuthenticatedAt,
				LastQualifyingActivityAt: timing.LastQualifyingActivityAt,
				IdleExpiresAt:            timing.IdleExpiresAt,
				AbsoluteExpiresAt:        timing.AbsoluteExpiresAt,
				SessionExpiresAt:         timing.SessionExpiresAt,
				CreatedAt:                timing.AuthenticatedAt,
				UpdatedAt:                timing.AuthenticatedAt,
			}, nil, nil
		},
		listIncidentMembershipSummariesFunc: func(_ context.Context, gotUserID uuid.UUID) ([]authn.IncidentMembershipSummary, error) {
			if gotUserID != userID {
				t.Fatalf("unexpected membership lookup user: got %s want %s", gotUserID, userID)
			}
			return []authn.IncidentMembershipSummary{
				{IncidentID: incidentID, Role: "commander"},
			}, nil
		},
	}

	service := newUnitService(t, store, &hubStub{}, keys, now)
	recorder := httptest.NewRecorder()
	request := newJSONRequest(t, http.MethodPost, "/api/v1/auth/login", `{
		"username":" analyst@example.test ",
		"password":"  Exact Login Secret  "
	}`)

	service.handleLogin(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected login status: got %d want %d", recorder.Code, http.StatusOK)
	}
	if createCalls != 1 {
		t.Fatalf("expected one session creation, got %d", createCalls)
	}
	if !capturedTiming.AuthenticatedAt.Equal(now) {
		t.Fatalf("unexpected authenticated_at: got %s want %s", capturedTiming.AuthenticatedAt, now)
	}
	if !capturedTiming.LastQualifyingActivityAt.Equal(now) {
		t.Fatalf("unexpected last_qualifying_activity_at: got %s want %s", capturedTiming.LastQualifyingActivityAt, now)
	}
	if !capturedTiming.IdleExpiresAt.Equal(now.Add(30 * time.Minute)) {
		t.Fatalf("unexpected idle_expires_at: got %s", capturedTiming.IdleExpiresAt)
	}
	if !capturedTiming.SessionExpiresAt.Equal(capturedTiming.IdleExpiresAt) {
		t.Fatalf("expected session_expires_at to clamp to idle expiry on create: session=%s idle=%s", capturedTiming.SessionExpiresAt, capturedTiming.IdleExpiresAt)
	}

	data := decodeSuccessData(t, recorder)
	if got := data["user_id"]; got != userID.String() {
		t.Fatalf("unexpected session user_id: got %v want %s", got, userID)
	}
	if got := data["provider_type"]; got != "local" {
		t.Fatalf("unexpected provider_type: got %v", got)
	}
	if got := data["authenticated_at"]; got != now.Format(time.RFC3339Nano) {
		t.Fatalf("unexpected authenticated_at response: got %v", got)
	}
	if got := data["idle_expires_at"]; got != now.Add(30*time.Minute).Format(time.RFC3339Nano) {
		t.Fatalf("unexpected idle_expires_at response: got %v", got)
	}
	if got := data["absolute_expires_at"]; got != now.Add(12*time.Hour).Format(time.RFC3339Nano) {
		t.Fatalf("unexpected absolute_expires_at response: got %v", got)
	}
	if got := data["session_expires_at"]; got != now.Add(30*time.Minute).Format(time.RFC3339Nano) {
		t.Fatalf("unexpected session_expires_at response: got %v", got)
	}
	memberships, ok := data["memberships"].([]any)
	if !ok || len(memberships) != 1 {
		t.Fatalf("expected one membership summary, got %#v", data["memberships"])
	}
	member, ok := memberships[0].(map[string]any)
	if !ok || member["incident_id"] != incidentID.String() || member["role"] != "commander" {
		t.Fatalf("unexpected membership summary: %#v", memberships[0])
	}
	authcookietest.RequireAuthCookies(t, recorder.Result().Cookies())
}

func TestPhase1_SessionInspectionRoute_U_1_04(t *testing.T) {
	now := time.Date(2026, time.April, 17, 12, 15, 0, 0, time.UTC)
	keys := loadUnitMasterKeys(t)

	t.Run("returns the singleton session resource without sliding idle expiry", func(t *testing.T) {
		userID := uuid.MustParse("10000000-0000-0000-0000-000000000104")
		sessionID := uuid.MustParse("10000000-0000-0000-0000-000000000105")
		incidentID := uuid.MustParse("10000000-0000-0000-0000-000000000106")
		token := "session-inspection-token"

		sessionLookups := 0
		slideCalls := 0
		store := &authStoreStub{
			getSessionByFingerprintFunc: func(_ context.Context, fingerprint []byte) (authn.SessionRecord, authn.UserRecord, error) {
				sessionLookups++
				if !bytes.Equal(fingerprint, authn.FingerprintToken(keys, token)) {
					t.Fatal("session inspection used the wrong session fingerprint")
				}
				user := activeUserRecord(userID, "inspection@example.test")
				user.DisplayName = "Inspector"
				return activeSessionRecord(sessionID, userID, now), user, nil
			},
			listIncidentMembershipSummariesFunc: func(_ context.Context, gotUserID uuid.UUID) ([]authn.IncidentMembershipSummary, error) {
				if gotUserID != userID {
					t.Fatalf("unexpected membership lookup user: got %s want %s", gotUserID, userID)
				}
				return []authn.IncidentMembershipSummary{
					{IncidentID: incidentID, Role: "viewer"},
				}, nil
			},
			slideSessionFunc: func(_ context.Context, _ uuid.UUID, timing authn.SessionTiming) (authn.SessionTiming, error) {
				slideCalls++
				return timing, nil
			},
		}
		service := newUnitService(t, store, &hubStub{}, keys, now)

		recorder := httptest.NewRecorder()
		request := newJSONRequest(t, http.MethodGet, "/api/v1/auth/session", "")
		addSessionCookiesOnly(request, keys, token)
		service.handleSession(recorder, request)

		if recorder.Code != http.StatusOK {
			t.Fatalf("unexpected session inspection status: got %d want %d", recorder.Code, http.StatusOK)
		}
		if sessionLookups != 1 {
			t.Fatalf("expected one session lookup, got %d", sessionLookups)
		}
		if slideCalls != 0 {
			t.Fatalf("session inspection must not slide idle expiry, got %d slide calls", slideCalls)
		}

		data := decodeSuccessData(t, recorder)
		if got := data["user_id"]; got != userID.String() {
			t.Fatalf("unexpected session user_id: got %v want %s", got, userID)
		}
		if got := data["display_name"]; got != "Inspector" {
			t.Fatalf("unexpected session display_name: got %v", got)
		}
		if got := data["provider_type"]; got != "local" {
			t.Fatalf("unexpected session provider_type: got %v", got)
		}
		if got := data["idle_expires_at"]; got != now.Add(30*time.Minute).Format(time.RFC3339Nano) {
			t.Fatalf("unexpected idle_expires_at: got %v", got)
		}
		if got := data["session_expires_at"]; got != now.Add(30*time.Minute).Format(time.RFC3339Nano) {
			t.Fatalf("unexpected session_expires_at: got %v", got)
		}
		memberships, ok := data["memberships"].([]any)
		if !ok || len(memberships) != 1 {
			t.Fatalf("expected one membership summary, got %#v", data["memberships"])
		}
		member, ok := memberships[0].(map[string]any)
		if !ok || member["incident_id"] != incidentID.String() || member["role"] != "viewer" {
			t.Fatalf("unexpected membership summary: %#v", memberships[0])
		}
	})

	t.Run("rejects pagination members on the actual route before session lookup", func(t *testing.T) {
		sessionLookups := 0
		store := &authStoreStub{
			getSessionByFingerprintFunc: func(context.Context, []byte) (authn.SessionRecord, authn.UserRecord, error) {
				sessionLookups++
				return authn.SessionRecord{}, authn.UserRecord{}, nil
			},
		}
		service := newUnitService(t, store, &hubStub{}, keys, now)

		recorder := httptest.NewRecorder()
		request := newJSONRequest(t, http.MethodGet, "/api/v1/auth/session?limit=10", "")
		service.handleSession(recorder, request)

		requireErrorEnvelope(t, recorder, http.StatusBadRequest, "invalid_pagination_request", "")
		if sessionLookups != 0 {
			t.Fatalf("session lookup must not run after pagination rejection, got %d lookups", sessionLookups)
		}
	})
}

func TestPhase1_UserCreateRouteDefaults_U_1_07(t *testing.T) {
	now := time.Date(2026, time.April, 17, 12, 20, 0, 0, time.UTC)
	keys := loadUnitMasterKeys(t)

	t.Run("rejects client-driven is_active before the create path runs", func(t *testing.T) {
		adminID := uuid.MustParse("10000000-0000-0000-0000-000000000107")
		sessionID := uuid.MustParse("10000000-0000-0000-0000-000000000108")
		token := "user-create-invalid-field-token"

		createCalls := 0
		store := &authStoreStub{
			getSessionByFingerprintFunc: func(_ context.Context, fingerprint []byte) (authn.SessionRecord, authn.UserRecord, error) {
				if !bytes.Equal(fingerprint, authn.FingerprintToken(keys, token)) {
					t.Fatal("user create used the wrong session fingerprint")
				}
				return activeSessionRecord(sessionID, adminID, now), deploymentAdminUserRecord(adminID, "admin@example.test"), nil
			},
			createUserFunc: func(context.Context, authn.UserRecord, string, string, string, bool, bool, string, []byte, string, time.Time) (authn.UserCreateResult, error) {
				createCalls++
				return authn.UserCreateResult{}, nil
			},
		}
		service := newUnitService(t, store, &hubStub{}, keys, now)

		recorder := httptest.NewRecorder()
		request := newJSONRequest(t, http.MethodPost, "/api/v1/users", `{
			"client_txn_id":"txn-user-create-invalid-field",
			"auth_kind":"local",
			"email":"phase1-invalid@example.test",
			"display_name":"Phase 1 Invalid",
			"initial_password":"Phase1InvalidPass!",
			"is_active":false
		}`)
		addSessionAuth(request, keys, token, true)
		service.handleUsersCollection(recorder, request)

		requireErrorEnvelope(t, recorder, http.StatusBadRequest, "invalid_mutation_payload", "is_active")
		if createCalls != 0 {
			t.Fatalf("CreateUser must not run for client-driven is_active, got %d calls", createCalls)
		}
	})

	t.Run("rejects invalid auth_kind before the create path runs", func(t *testing.T) {
		adminID := uuid.MustParse("10000000-0000-0000-0000-000000000107")
		sessionID := uuid.MustParse("10000000-0000-0000-0000-000000000108")
		token := "user-create-invalid-auth-kind-token"

		createCalls := 0
		store := &authStoreStub{
			getSessionByFingerprintFunc: func(_ context.Context, fingerprint []byte) (authn.SessionRecord, authn.UserRecord, error) {
				if !bytes.Equal(fingerprint, authn.FingerprintToken(keys, token)) {
					t.Fatal("user create used the wrong session fingerprint")
				}
				return activeSessionRecord(sessionID, adminID, now), deploymentAdminUserRecord(adminID, "admin@example.test"), nil
			},
			createUserFunc: func(context.Context, authn.UserRecord, string, string, string, bool, bool, string, []byte, string, time.Time) (authn.UserCreateResult, error) {
				createCalls++
				return authn.UserCreateResult{}, nil
			},
		}
		service := newUnitService(t, store, &hubStub{}, keys, now)

		recorder := httptest.NewRecorder()
		request := newJSONRequest(t, http.MethodPost, "/api/v1/users", `{
			"client_txn_id":"txn-user-create-invalid-auth-kind",
			"auth_kind":"ldap",
			"email":"phase1-invalid-auth-kind@example.test",
			"display_name":"Phase 1 Invalid Auth Kind",
			"initial_password":"Phase1InvalidAuthKindPass!"
		}`)
		addSessionAuth(request, keys, token, true)
		service.handleUsersCollection(recorder, request)

		requireErrorEnvelope(t, recorder, http.StatusBadRequest, "invalid_mutation_payload", "auth_kind")
		if got := decodeErrorDetails(t, recorder)["reason_code"]; got != "invalid_auth_kind" {
			t.Fatalf("unexpected auth_kind rejection reason_code: got %v want invalid_auth_kind", got)
		}
		if createCalls != 0 {
			t.Fatalf("CreateUser must not run for invalid auth_kind, got %d calls", createCalls)
		}
	})

	t.Run("applies defaults and returns a secret-safe resource", func(t *testing.T) {
		adminID := uuid.MustParse("10000000-0000-0000-0000-000000000109")
		sessionID := uuid.MustParse("10000000-0000-0000-0000-000000000110")
		userID := uuid.MustParse("10000000-0000-0000-0000-00000000010b")
		token := "user-create-defaults-token"
		initialPassword := "Phase1DefaultsPass!"
		responseCreatedAt := now.Add(time.Minute)

		var capturedPasswordHash string
		var capturedRequestHash []byte
		store := &authStoreStub{
			getSessionByFingerprintFunc: func(_ context.Context, fingerprint []byte) (authn.SessionRecord, authn.UserRecord, error) {
				if !bytes.Equal(fingerprint, authn.FingerprintToken(keys, token)) {
					t.Fatal("user create used the wrong session fingerprint")
				}
				return activeSessionRecord(sessionID, adminID, now), deploymentAdminUserRecord(adminID, "admin@example.test"), nil
			},
			createUserFunc: func(_ context.Context, actor authn.UserRecord, email string, displayName string, passwordHash string, mfaRequired bool, isDeploymentAdmin bool, clientTxnID string, requestHash []byte, requestID string, createdAt time.Time) (authn.UserCreateResult, error) {
				if actor.ID != adminID || !actor.IsDeploymentAdmin {
					t.Fatalf("unexpected create actor: %#v", actor)
				}
				if email != "phase1-create@example.test" {
					t.Fatalf("unexpected create email: got %q", email)
				}
				if displayName != "Phase 1 Create" {
					t.Fatalf("unexpected create display_name: got %q", displayName)
				}
				if !mfaRequired {
					t.Fatal("omitted mfa_required must default to true")
				}
				if isDeploymentAdmin {
					t.Fatal("omitted is_deployment_admin must default to false")
				}
				if clientTxnID != "txn-user-create-defaults" {
					t.Fatalf("unexpected create client_txn_id: got %q", clientTxnID)
				}
				if requestID != "" {
					t.Fatalf("unexpected request id in direct handler test: got %q", requestID)
				}
				if !createdAt.Equal(now) {
					t.Fatalf("unexpected create time: got %s want %s", createdAt, now)
				}
				capturedPasswordHash = passwordHash
				capturedRequestHash = append([]byte(nil), requestHash...)
				return authn.UserCreateResult{
					StatusCode: http.StatusCreated,
					ResponseJSON: mustJSON(t, map[string]any{
						"user_id":             userID,
						"email":               email,
						"display_name":        displayName,
						"is_active":           true,
						"mfa_required":        true,
						"is_deployment_admin": false,
						"user_version":        1,
						"created_at":          responseCreatedAt,
						"updated_at":          responseCreatedAt,
						"auth_bindings": []map[string]any{
							{
								"provider_type": "local",
								"provider_key":  "local",
								"username":      email,
								"created_at":    responseCreatedAt,
							},
						},
					}),
				}, nil
			},
		}
		service := newUnitService(t, store, &hubStub{}, keys, now)

		recorder := httptest.NewRecorder()
		request := newJSONRequest(t, http.MethodPost, "/api/v1/users", `{
			"client_txn_id":"txn-user-create-defaults",
			"auth_kind":"local",
			"email":" phase1-create@example.test ",
			"display_name":" Phase 1 Create ",
			"initial_password":"Phase1DefaultsPass!"
		}`)
		addSessionAuth(request, keys, token, true)
		service.handleUsersCollection(recorder, request)

		if recorder.Code != http.StatusCreated {
			t.Fatalf("unexpected create status: got %d want %d", recorder.Code, http.StatusCreated)
		}
		if ok, err := authn.VerifyPasswordHash(capturedPasswordHash, initialPassword); err != nil || !ok {
			t.Fatalf("captured password hash must verify initial password, ok=%v err=%v", ok, err)
		}
		expectedHash := hashRequestPayload(map[string]any{
			"auth_kind":           "local",
			"email":               "phase1-create@example.test",
			"display_name":        "Phase 1 Create",
			"initial_password":    requestSecretFingerprint(keys, initialPassword),
			"mfa_required":        true,
			"is_deployment_admin": false,
		})
		if !hashesEqual(capturedRequestHash, expectedHash) {
			t.Fatalf("unexpected create request hash: got %x want %x", capturedRequestHash, expectedHash)
		}

		data := decodeSuccessData(t, recorder)
		if got := data["email"]; got != "phase1-create@example.test" {
			t.Fatalf("unexpected normalized create response email: got %v want phase1-create@example.test", got)
		}
		if got := data["display_name"]; got != "Phase 1 Create" {
			t.Fatalf("unexpected normalized create response display_name: got %v want Phase 1 Create", got)
		}
		if data["is_active"] != true || data["mfa_required"] != true || data["is_deployment_admin"] != false {
			t.Fatalf("unexpected create response defaults: %#v", data)
		}
		requireNoSecretKeys(t, data, "password_hash", "initial_password", "bootstrap_token", "secret_base32")
	})
}

func TestPhase1_CSRFProtectionRoutes_U_1_09(t *testing.T) {
	now := time.Date(2026, time.April, 17, 12, 40, 0, 0, time.UTC)
	keys := loadUnitMasterKeys(t)
	fixture := phase1test.RouteInventoryFixture{
		UserID: "10000000-0000-0000-0000-000000000124",
	}
	token := "csrf-state-changing-token"
	routes := phase1test.RoutesForHarness(t, phase1test.PublicRouteInventory(), phase1test.RouteHarnessCSRF)

	for _, tc := range []struct {
		name       string
		headerName string
		headerVal  string
	}{
		{
			name: "missing csrf header",
		},
		{
			name:       "invalid csrf header",
			headerName: authn.CSRFHeaderName,
			headerVal:  "wrong-csrf-token",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			for _, route := range routes {
				t.Run(string(route.ID), func(t *testing.T) {
					hub := &hubStub{}
					service := newUnitService(t, &authStoreStub{
						getSessionByFingerprintFunc: func(context.Context, []byte) (authn.SessionRecord, authn.UserRecord, error) {
							t.Fatal("csrf failure must not look up the session")
							return authn.SessionRecord{}, authn.UserRecord{}, nil
						},
					}, hub, keys, now)

					recorder := httptest.NewRecorder()
					request := newJSONRequest(t, route.Method, phase1test.BuildRoutePath(route.Template, fixture), phase1RouteCSRFPayload(t, route))
					addSessionCookiesOnly(request, keys, token)
					if tc.headerName != "" {
						request.Header.Set(tc.headerName, tc.headerVal)
					}
					dispatchPhase1UnitRoute(t, service, route, recorder, request)

					requireErrorEnvelope(t, recorder, http.StatusForbidden, "csrf_verification_failed", "")
					if len(hub.revocations) != 0 {
						t.Fatalf("csrf failure must not publish revocations, got %#v", hub.revocations)
					}
				})
			}
		})
	}

	for _, route := range routes {
		switch route.ID {
		case phase1test.RoutePasswordChange, phase1test.RouteTOTPBegin, phase1test.RouteTOTPComplete:
		default:
			continue
		}

		t.Run(string(route.ID)+"/malformed_body_still_fails_csrf", func(t *testing.T) {
			hub := &hubStub{}
			service := newUnitService(t, &authStoreStub{
				getSessionByFingerprintFunc: func(context.Context, []byte) (authn.SessionRecord, authn.UserRecord, error) {
					t.Fatal("csrf failure must not look up the session")
					return authn.SessionRecord{}, authn.UserRecord{}, nil
				},
			}, hub, keys, now)

			recorder := httptest.NewRecorder()
			request := newJSONRequest(t, route.Method, phase1test.BuildRoutePath(route.Template, fixture), `{`)
			addSessionCookiesOnly(request, keys, token)
			dispatchPhase1UnitRoute(t, service, route, recorder, request)

			requireErrorEnvelope(t, recorder, http.StatusForbidden, "csrf_verification_failed", "")
			if len(hub.revocations) != 0 {
				t.Fatalf("csrf failure must not publish revocations, got %#v", hub.revocations)
			}
		})
	}
}

func TestPhase1_CredentialStateRoute_U_1_10(t *testing.T) {
	now := time.Date(2026, time.April, 17, 12, 50, 0, 0, time.UTC)
	keys := loadUnitMasterKeys(t)

	t.Run("returns the safe credential-state resource for each declared totp.state", func(t *testing.T) {
		userID := uuid.MustParse("10000000-0000-0000-0000-000000000126")
		sessionID := uuid.MustParse("10000000-0000-0000-0000-000000000127")
		token := "credential-state-token"
		changedAt := now.Add(-2 * time.Hour)
		enrolledAt := now.Add(-time.Hour)
		pendingExpiresAt := now.Add(5 * time.Minute)

		for _, tc := range []struct {
			name            string
			user            authn.UserRecord
			pending         *authn.PendingTOTPEnrollmentRecord
			wantState       string
			wantPendingText string
		}{
			{
				name: "not enrolled",
				user: authn.UserRecord{
					ID:                userID,
					Email:             "state-not-enrolled@example.test",
					DisplayName:       "Not Enrolled",
					IsActive:          true,
					MFARequired:       true,
					PasswordChangedAt: changedAt,
				},
				wantState: "not_enrolled",
			},
			{
				name: "pending",
				user: authn.UserRecord{
					ID:                userID,
					Email:             "state-pending@example.test",
					DisplayName:       "Pending",
					IsActive:          true,
					MFARequired:       true,
					PasswordChangedAt: changedAt,
				},
				pending: &authn.PendingTOTPEnrollmentRecord{
					ID:        uuid.MustParse("10000000-0000-0000-0000-000000000128"),
					UserID:    userID,
					ExpiresAt: pendingExpiresAt,
				},
				wantState:       "pending",
				wantPendingText: pendingExpiresAt.Format(time.RFC3339Nano),
			},
			{
				name: "active",
				user: authn.UserRecord{
					ID:                userID,
					Email:             "state-active@example.test",
					DisplayName:       "Active",
					IsActive:          true,
					MFARequired:       true,
					PasswordChangedAt: changedAt,
					TOTPEnrolledAt:    &enrolledAt,
				},
				wantState: "active",
			},
		} {
			t.Run(tc.name, func(t *testing.T) {
				slideCalls := 0
				store := &authStoreStub{
					getSessionByFingerprintFunc: func(_ context.Context, fingerprint []byte) (authn.SessionRecord, authn.UserRecord, error) {
						if !bytes.Equal(fingerprint, authn.FingerprintToken(keys, token)) {
							t.Fatal("credential state used the wrong session fingerprint")
						}
						return activeSessionRecord(sessionID, userID, now), tc.user, nil
					},
					slideSessionFunc: func(_ context.Context, gotSessionID uuid.UUID, timing authn.SessionTiming) (authn.SessionTiming, error) {
						slideCalls++
						if gotSessionID != sessionID {
							t.Fatalf("unexpected slide session id: got %s want %s", gotSessionID, sessionID)
						}
						if !timing.LastQualifyingActivityAt.Equal(now) {
							t.Fatalf("unexpected slide last_qualifying_activity_at: got %s want %s", timing.LastQualifyingActivityAt, now)
						}
						return timing, nil
					},
					getPendingTOTPEnrollmentForUserFunc: func(_ context.Context, gotUserID uuid.UUID, current time.Time) (*authn.PendingTOTPEnrollmentRecord, error) {
						if gotUserID != userID {
							t.Fatalf("unexpected pending-enrollment user: got %s want %s", gotUserID, userID)
						}
						if !current.Equal(now) {
							t.Fatalf("unexpected pending-enrollment time: got %s want %s", current, now)
						}
						return tc.pending, nil
					},
				}
				service := newUnitService(t, store, &hubStub{}, keys, now)

				recorder := httptest.NewRecorder()
				request := newJSONRequest(t, http.MethodGet, "/api/v1/auth/credential-state", "")
				addSessionCookiesOnly(request, keys, token)
				service.handleCredentialState(recorder, request)

				if recorder.Code != http.StatusOK {
					t.Fatalf("unexpected credential-state status: got %d want %d", recorder.Code, http.StatusOK)
				}
				if slideCalls != 1 {
					t.Fatalf("credential-state inspection must slide idle expiry once, got %d calls", slideCalls)
				}

				data := decodeSuccessData(t, recorder)
				if got := data["auth_kind"]; got != "local" {
					t.Fatalf("unexpected auth_kind: got %v", got)
				}
				if got := data["user_id"]; got != userID.String() {
					t.Fatalf("unexpected credential-state user_id: got %v want %s", got, userID)
				}
				if got := data["recovery_model"]; got != "admin_assisted" {
					t.Fatalf("unexpected recovery_model: got %v", got)
				}
				totpState := data["totp"].(map[string]any)
				if got := totpState["state"]; got != tc.wantState {
					t.Fatalf("unexpected totp.state: got %v want %s", got, tc.wantState)
				}
				if tc.wantPendingText == "" {
					if pendingValue, ok := totpState["pending_expires_at"]; ok && pendingValue != nil {
						t.Fatalf("unexpected pending_expires_at for %s: %#v", tc.wantState, pendingValue)
					}
				} else if got := totpState["pending_expires_at"]; got != tc.wantPendingText {
					t.Fatalf("unexpected pending_expires_at: got %v want %s", got, tc.wantPendingText)
				}
				requireNoSecretKeys(t, data, "password_hash", "bootstrap_token", "secret_base32", "otpauth_uri")
			})
		}
	})

	t.Run("rejects bootstrap-token auth on the ordinary credential-state route", func(t *testing.T) {
		userID := uuid.MustParse("10000000-0000-0000-0000-000000000129")
		bootstrapTokenID := uuid.MustParse("10000000-0000-0000-0000-00000000012a")
		bootstrapToken := "credential-state-bootstrap-token"

		store := &authStoreStub{
			getSessionByFingerprintFunc: func(_ context.Context, fingerprint []byte) (authn.SessionRecord, authn.UserRecord, error) {
				if !bytes.Equal(fingerprint, authn.FingerprintToken(keys, bootstrapToken)) {
					t.Fatal("unexpected bearer fingerprint during credential-state bootstrap rejection")
				}
				return authn.SessionRecord{}, authn.UserRecord{}, authn.ErrNotFound
			},
			getBootstrapTokenByFingerprintFunc: func(_ context.Context, fingerprint []byte) (authn.BootstrapTokenRecord, authn.UserRecord, error) {
				if !bytes.Equal(fingerprint, authn.FingerprintToken(keys, bootstrapToken)) {
					t.Fatal("unexpected bootstrap fingerprint during credential-state bootstrap rejection")
				}
				return authn.BootstrapTokenRecord{
					ID:        bootstrapTokenID,
					UserID:    userID,
					IssuedAt:  now.Add(-time.Minute),
					ExpiresAt: now.Add(9 * time.Minute),
				}, activeUserRecord(userID, "bootstrap-rejected@example.test"), nil
			},
		}
		service := newUnitService(t, store, &hubStub{}, keys, now)

		recorder := httptest.NewRecorder()
		request := newJSONRequest(t, http.MethodGet, "/api/v1/auth/credential-state", "")
		request.Header.Set("Authorization", "Bearer "+bootstrapToken)
		service.handleCredentialState(recorder, request)

		requireErrorEnvelope(t, recorder, http.StatusConflict, "credential_bootstrap_rejected", "")
		details := decodeErrorDetails(t, recorder)
		if got := details["reason_code"]; got != "not_allowed_for_route" {
			t.Fatalf("unexpected bootstrap rejection reason_code: got %v want not_allowed_for_route", got)
		}
	})
}

func TestPhase1_PasswordChangeRouteContracts_U_1_11(t *testing.T) {
	now := time.Date(2026, time.April, 17, 13, 0, 0, 0, time.UTC)
	keys := loadUnitMasterKeys(t)

	t.Run("current password verification remains exact", func(t *testing.T) {
		userID := uuid.MustParse("10000000-0000-0000-0000-000000000131")
		sessionID := uuid.MustParse("10000000-0000-0000-0000-000000000132")
		token := "password-change-exact-token"
		currentPassword := "  Exact Current  "
		passwordHash, err := authn.HashPassword(currentPassword)
		if err != nil {
			t.Fatalf("hash current password: %v", err)
		}

		changeCalls := 0
		store := &authStoreStub{
			getSessionByFingerprintFunc: func(_ context.Context, fingerprint []byte) (authn.SessionRecord, authn.UserRecord, error) {
				if !bytes.Equal(fingerprint, authn.FingerprintToken(keys, token)) {
					t.Fatal("password change used wrong session fingerprint")
				}
				return activeSessionRecord(sessionID, userID, now), authn.UserRecord{
					ID:           userID,
					Email:        "exact@example.test",
					DisplayName:  "Exact",
					PasswordHash: passwordHash,
					IsActive:     true,
				}, nil
			},
			getRouteIdempotencyFunc: func(context.Context, authn.RouteIdempotencyKey) (authn.RouteIdempotencyRecord, error) {
				return authn.RouteIdempotencyRecord{}, authn.ErrNotFound
			},
			changePasswordFunc: func(context.Context, authn.UserRecord, string, []byte, string, string, time.Time) (authn.PasswordChangeResult, error) {
				changeCalls++
				return authn.PasswordChangeResult{}, nil
			},
		}
		service := newUnitService(t, store, &hubStub{}, keys, now)

		recorder := httptest.NewRecorder()
		request := newJSONRequest(t, http.MethodPost, "/api/v1/auth/password/change", `{
			"client_txn_id":"txn-password-exact",
			"current_password":"Exact Current",
			"new_password":"Replacement passphrase 1"
		}`)
		addSessionAuth(request, keys, token, true)
		service.handlePasswordChange(recorder, request)

		requireErrorEnvelope(t, recorder, http.StatusConflict, "invalid_current_password", "")
		if changeCalls != 0 {
			t.Fatalf("store.ChangePassword must not run on invalid current password, got %d calls", changeCalls)
		}
	})

	t.Run("active totp remains required before the mutation is delegated", func(t *testing.T) {
		userID := uuid.MustParse("10000000-0000-0000-0000-000000000141")
		sessionID := uuid.MustParse("10000000-0000-0000-0000-000000000142")
		token := "password-change-mfa-token"
		user, _ := activeTOTPUserRecord(t, keys, userID, "mfa@example.test", "  Exact Current  ", now)

		changeCalls := 0
		store := &authStoreStub{
			getSessionByFingerprintFunc: func(_ context.Context, fingerprint []byte) (authn.SessionRecord, authn.UserRecord, error) {
				if !bytes.Equal(fingerprint, authn.FingerprintToken(keys, token)) {
					t.Fatal("password change used wrong session fingerprint")
				}
				return activeSessionRecord(sessionID, userID, now), user, nil
			},
			getRouteIdempotencyFunc: func(context.Context, authn.RouteIdempotencyKey) (authn.RouteIdempotencyRecord, error) {
				return authn.RouteIdempotencyRecord{}, authn.ErrNotFound
			},
			changePasswordFunc: func(context.Context, authn.UserRecord, string, []byte, string, string, time.Time) (authn.PasswordChangeResult, error) {
				changeCalls++
				return authn.PasswordChangeResult{}, nil
			},
		}
		service := newUnitService(t, store, &hubStub{}, keys, now)

		recorder := httptest.NewRecorder()
		request := newJSONRequest(t, http.MethodPost, "/api/v1/auth/password/change", `{
			"client_txn_id":"txn-password-mfa",
			"current_password":"  Exact Current  ",
			"new_password":"Replacement passphrase 1"
		}`)
		addSessionAuth(request, keys, token, true)
		service.handlePasswordChange(recorder, request)

		requireErrorEnvelope(t, recorder, http.StatusUnauthorized, "invalid_second_factor", "")
		if changeCalls != 0 {
			t.Fatalf("store.ChangePassword must not run without the active totp assertion, got %d calls", changeCalls)
		}
	})

	t.Run("successful password change fingerprints secrets and revokes every active session", func(t *testing.T) {
		userID := uuid.MustParse("10000000-0000-0000-0000-000000000151")
		currentSessionID := uuid.MustParse("10000000-0000-0000-0000-000000000152")
		otherSessionID := uuid.MustParse("10000000-0000-0000-0000-000000000153")
		token := "password-change-success-token"
		currentPassword := "  Exact Current  "
		newPassword := "Replacement passphrase 1"
		user, secretBase32 := activeTOTPUserRecord(t, keys, userID, "change@example.test", currentPassword, now)
		code := generateTOTPCodeAt(t, secretBase32, now)

		var capturedHash []byte
		var capturedNewPasswordHash string
		store := &authStoreStub{
			getSessionByFingerprintFunc: func(_ context.Context, fingerprint []byte) (authn.SessionRecord, authn.UserRecord, error) {
				if !bytes.Equal(fingerprint, authn.FingerprintToken(keys, token)) {
					t.Fatal("password change used wrong session fingerprint")
				}
				return activeSessionRecord(currentSessionID, userID, now), user, nil
			},
			getRouteIdempotencyFunc: func(context.Context, authn.RouteIdempotencyKey) (authn.RouteIdempotencyRecord, error) {
				return authn.RouteIdempotencyRecord{}, authn.ErrNotFound
			},
			changePasswordFunc: func(_ context.Context, actor authn.UserRecord, clientTxnID string, requestHash []byte, newPasswordHash string, requestID string, changedAt time.Time) (authn.PasswordChangeResult, error) {
				if actor.ID != userID {
					t.Fatalf("password change actor mismatch: got %s want %s", actor.ID, userID)
				}
				if clientTxnID != "txn-password-success" {
					t.Fatalf("unexpected password change client_txn_id: got %q", clientTxnID)
				}
				if requestID != "" {
					t.Fatalf("unexpected request id in direct handler test: got %q", requestID)
				}
				if !changedAt.Equal(now) {
					t.Fatalf("unexpected password change time: got %s want %s", changedAt, now)
				}
				capturedHash = append([]byte(nil), requestHash...)
				capturedNewPasswordHash = newPasswordHash
				return authn.PasswordChangeResult{
					RevokedSessionIDs: []uuid.UUID{currentSessionID, otherSessionID},
					ResponseJSON: mustJSON(t, map[string]any{
						"user_id":          userID,
						"password":         map[string]any{"changed_at": changedAt},
						"sessions_revoked": true,
					}),
				}, nil
			},
		}
		hub := &hubStub{}
		service := newUnitService(t, store, hub, keys, now)

		recorder := httptest.NewRecorder()
		request := newJSONRequest(t, http.MethodPost, "/api/v1/auth/password/change", `{
			"client_txn_id":"txn-password-success",
			"current_password":"  Exact Current  ",
			"new_password":"Replacement passphrase 1",
			"second_factor":{"kind":"totp","assertion":{"code":"`+code+`"}}
		}`)
		addSessionAuth(request, keys, token, true)
		service.handlePasswordChange(recorder, request)

		if recorder.Code != http.StatusOK {
			t.Fatalf("unexpected password change status: got %d want %d", recorder.Code, http.StatusOK)
		}
		if ok, err := authn.VerifyPasswordHash(capturedNewPasswordHash, newPassword); err != nil || !ok {
			t.Fatalf("new password hash must verify the replacement password, ok=%v err=%v", ok, err)
		}
		if ok, err := authn.VerifyPasswordHash(capturedNewPasswordHash, currentPassword); err != nil {
			t.Fatalf("verifying old password against replacement hash: %v", err)
		} else if ok {
			t.Fatal("replacement hash must not accept the current password")
		}
		expectedHash := hashRequestPayload(map[string]any{
			"current_password": requestSecretFingerprint(keys, currentPassword),
			"new_password":     requestSecretFingerprint(keys, newPassword),
			"second_factor": requestSecondFactorHashPayload(keys, &SecondFactorAssertion{
				Kind: "totp",
				Code: code,
			}),
		})
		if !hashesEqual(capturedHash, expectedHash) {
			t.Fatalf("unexpected password change request hash: got %x want %x", capturedHash, expectedHash)
		}
		requireRevocations(t, hub.revocations, []revocationCall{
			{sessionID: currentSessionID, reasonCode: "session_revoked"},
			{sessionID: otherSessionID, reasonCode: "session_revoked"},
		})
		data := decodeSuccessData(t, recorder)
		if data["sessions_revoked"] != true {
			t.Fatalf("unexpected password change response: %#v", data)
		}
		requireNoSecretKeys(t, data, "password_hash", "current_password", "new_password", "secret_base32", "bootstrap_token")
		requireCookieCleared(t, recorder.Result().Cookies(), authn.SessionCookieName)
		requireCookieCleared(t, recorder.Result().Cookies(), authn.CSRFCookieName)
	})

	t.Run("replay returns the original secret-safe stored payload without mutating again", func(t *testing.T) {
		userID := uuid.MustParse("10000000-0000-0000-0000-000000000154")
		sessionID := uuid.MustParse("10000000-0000-0000-0000-000000000155")
		token := "password-change-replay-token"
		currentPassword := "Replay Current Password!"
		newPassword := "Replay Replacement Password!"
		passwordHash, err := authn.HashPassword(currentPassword)
		if err != nil {
			t.Fatalf("hash replay current password: %v", err)
		}

		replayPayload := map[string]any{
			"user_id":          userID,
			"password":         map[string]any{"changed_at": now},
			"sessions_revoked": true,
		}
		expectedHash := hashRequestPayload(map[string]any{
			"current_password": requestSecretFingerprint(keys, currentPassword),
			"new_password":     requestSecretFingerprint(keys, newPassword),
			"second_factor":    requestSecondFactorHashPayload(keys, nil),
		})
		idempotencyLookups := 0
		changeCalls := 0
		store := &authStoreStub{
			getSessionByFingerprintFunc: func(_ context.Context, fingerprint []byte) (authn.SessionRecord, authn.UserRecord, error) {
				if !bytes.Equal(fingerprint, authn.FingerprintToken(keys, token)) {
					t.Fatal("password-change replay used the wrong session fingerprint")
				}
				return activeSessionRecord(sessionID, userID, now), authn.UserRecord{
					ID:           userID,
					Email:        "replay@example.test",
					DisplayName:  "Replay",
					PasswordHash: passwordHash,
					IsActive:     true,
				}, nil
			},
			getRouteIdempotencyFunc: func(_ context.Context, key authn.RouteIdempotencyKey) (authn.RouteIdempotencyRecord, error) {
				idempotencyLookups++
				if key.RouteKey != "auth.password.change" || key.ActorUserID != userID || key.ScopeKey != "actor" || key.ClientTxnID != "txn-password-replay" {
					t.Fatalf("unexpected idempotency lookup: key=%#v", key)
				}
				return authn.RouteIdempotencyRecord{
					RequestHash:  expectedHash,
					ResponseJSON: mustJSON(t, replayPayload),
				}, nil
			},
			changePasswordFunc: func(context.Context, authn.UserRecord, string, []byte, string, string, time.Time) (authn.PasswordChangeResult, error) {
				changeCalls++
				return authn.PasswordChangeResult{}, nil
			},
		}
		service := newUnitService(t, store, &hubStub{}, keys, now)

		recorder := httptest.NewRecorder()
		request := newJSONRequest(t, http.MethodPost, "/api/v1/auth/password/change", `{
			"client_txn_id":"txn-password-replay",
			"current_password":"Replay Current Password!",
			"new_password":"Replay Replacement Password!"
		}`)
		addSessionAuth(request, keys, token, true)
		service.handlePasswordChange(recorder, request)

		if recorder.Code != http.StatusOK {
			t.Fatalf("unexpected replay password-change status: got %d want %d", recorder.Code, http.StatusOK)
		}
		if idempotencyLookups != 1 {
			t.Fatalf("expected one idempotency lookup, got %d", idempotencyLookups)
		}
		if changeCalls != 0 {
			t.Fatalf("ChangePassword must not run for an exact replay, got %d calls", changeCalls)
		}
		data := decodeSuccessData(t, recorder)
		if data["sessions_revoked"] != true {
			t.Fatalf("unexpected replay password-change response: %#v", data)
		}
		requireNoSecretKeys(t, data, "password_hash", "current_password", "new_password", "secret_base32", "bootstrap_token")
		requireCookieCleared(t, recorder.Result().Cookies(), authn.SessionCookieName)
		requireCookieCleared(t, recorder.Result().Cookies(), authn.CSRFCookieName)
	})
}

func TestPhase1_TOTPRouteContracts_U_1_12(t *testing.T) {
	now := time.Date(2026, time.April, 17, 14, 0, 0, 0, time.UTC)
	keys := loadUnitMasterKeys(t)

	t.Run("begin rejects mixed bootstrap and session auth modes", func(t *testing.T) {
		userID := uuid.MustParse("10000000-0000-0000-0000-000000000161")
		sessionID := uuid.MustParse("10000000-0000-0000-0000-000000000162")
		bootstrapTokenID := uuid.MustParse("10000000-0000-0000-0000-000000000163")
		sessionToken := "totp-begin-session-token"
		bootstrapToken := "totp-begin-bootstrap-token"

		beginCalls := 0
		store := &authStoreStub{
			getSessionByFingerprintFunc: func(_ context.Context, fingerprint []byte) (authn.SessionRecord, authn.UserRecord, error) {
				switch {
				case bytes.Equal(fingerprint, authn.FingerprintToken(keys, sessionToken)):
					return activeSessionRecord(sessionID, userID, now), activeUserRecord(userID, "mixed@example.test"), nil
				case bytes.Equal(fingerprint, authn.FingerprintToken(keys, bootstrapToken)):
					return authn.SessionRecord{}, authn.UserRecord{}, authn.ErrNotFound
				default:
					t.Fatal("unexpected fingerprint lookup during mixed-auth test")
					return authn.SessionRecord{}, authn.UserRecord{}, authn.ErrNotFound
				}
			},
			getBootstrapTokenByFingerprintFunc: func(_ context.Context, fingerprint []byte) (authn.BootstrapTokenRecord, authn.UserRecord, error) {
				if !bytes.Equal(fingerprint, authn.FingerprintToken(keys, bootstrapToken)) {
					t.Fatal("unexpected bootstrap fingerprint during mixed-auth test")
				}
				return authn.BootstrapTokenRecord{
					ID:        bootstrapTokenID,
					UserID:    userID,
					IssuedAt:  now.Add(-time.Minute),
					ExpiresAt: now.Add(9 * time.Minute),
				}, activeUserRecord(userID, "mixed@example.test"), nil
			},
			beginTOTPEnrollmentFunc: func(context.Context, uuid.UUID, string, *uuid.UUID, *uuid.UUID, string, []byte, []byte, bool, time.Time) (authn.PendingTOTPEnrollmentRecord, bool, error) {
				beginCalls++
				return authn.PendingTOTPEnrollmentRecord{}, false, nil
			},
		}
		service := newUnitService(t, store, &hubStub{}, keys, now)

		recorder := httptest.NewRecorder()
		request := newJSONRequest(t, http.MethodPost, "/api/v1/auth/mfa/totp/begin", `{"client_txn_id":"txn-begin-mixed"}`)
		request.Header.Set("Authorization", "Bearer "+bootstrapToken)
		addSessionCookiesOnly(request, keys, sessionToken)
		service.handleTOTPBegin(recorder, request)

		requireErrorEnvelope(t, recorder, http.StatusBadRequest, "invalid_auth_request", "authorization")
		if beginCalls != 0 {
			t.Fatalf("BeginTOTPEnrollment must not run when multiple auth modes are present, got %d calls", beginCalls)
		}
	})

	t.Run("complete rejects mixed bootstrap and session auth modes", func(t *testing.T) {
		userID := uuid.MustParse("10000000-0000-0000-0000-000000000164")
		sessionID := uuid.MustParse("10000000-0000-0000-0000-000000000165")
		bootstrapTokenID := uuid.MustParse("10000000-0000-0000-0000-000000000166")
		sessionToken := "totp-complete-session-token"
		bootstrapToken := "totp-complete-bootstrap-token"
		enrollmentID := uuid.MustParse("10000000-0000-0000-0000-000000000167")

		completeCalls := 0
		store := &authStoreStub{
			getSessionByFingerprintFunc: func(_ context.Context, fingerprint []byte) (authn.SessionRecord, authn.UserRecord, error) {
				switch {
				case bytes.Equal(fingerprint, authn.FingerprintToken(keys, sessionToken)):
					return activeSessionRecord(sessionID, userID, now), activeUserRecord(userID, "mixed-complete@example.test"), nil
				case bytes.Equal(fingerprint, authn.FingerprintToken(keys, bootstrapToken)):
					return authn.SessionRecord{}, authn.UserRecord{}, authn.ErrNotFound
				default:
					t.Fatal("unexpected fingerprint lookup during mixed-auth complete test")
					return authn.SessionRecord{}, authn.UserRecord{}, authn.ErrNotFound
				}
			},
			getBootstrapTokenByFingerprintFunc: func(_ context.Context, fingerprint []byte) (authn.BootstrapTokenRecord, authn.UserRecord, error) {
				if !bytes.Equal(fingerprint, authn.FingerprintToken(keys, bootstrapToken)) {
					t.Fatal("unexpected bootstrap fingerprint during mixed-auth complete test")
				}
				return authn.BootstrapTokenRecord{
					ID:        bootstrapTokenID,
					UserID:    userID,
					IssuedAt:  now.Add(-time.Minute),
					ExpiresAt: now.Add(9 * time.Minute),
				}, activeUserRecord(userID, "mixed-complete@example.test"), nil
			},
			getPendingTOTPEnrollmentByIDFunc: func(context.Context, uuid.UUID) (*authn.PendingTOTPEnrollmentRecord, error) {
				t.Fatal("complete must not load pending enrollment when auth modes are mixed")
				return nil, nil
			},
			activateTOTPEnrollmentFunc: func(context.Context, authn.UserRecord, uuid.UUID, string, *uuid.UUID, *uuid.UUID, time.Time) (authn.TOTPCompleteResult, error) {
				completeCalls++
				return authn.TOTPCompleteResult{}, nil
			},
		}
		service := newUnitService(t, store, &hubStub{}, keys, now)

		recorder := httptest.NewRecorder()
		request := newJSONRequest(t, http.MethodPost, "/api/v1/auth/mfa/totp/complete", `{
			"client_txn_id":"txn-complete-mixed",
			"enrollment_id":"`+enrollmentID.String()+`",
			"code":"123456"
		}`)
		request.Header.Set("Authorization", "Bearer "+bootstrapToken)
		addSessionCookiesOnly(request, keys, sessionToken)
		service.handleTOTPComplete(recorder, request)

		requireErrorEnvelope(t, recorder, http.StatusBadRequest, "invalid_auth_request", "authorization")
		if completeCalls != 0 {
			t.Fatalf("ActivateTOTPEnrollment must not run when multiple auth modes are present, got %d calls", completeCalls)
		}
	})

	t.Run("begin emits setup material only on begin and scopes replay to the bootstrap token", func(t *testing.T) {
		userID := uuid.MustParse("10000000-0000-0000-0000-000000000171")
		bootstrapTokenID := uuid.MustParse("10000000-0000-0000-0000-000000000172")
		bootstrapToken := "totp-begin-only-bootstrap-token"
		enrollmentID := uuid.MustParse("10000000-0000-0000-0000-000000000173")
		secretBytes := []byte("01234567890123456789")
		ciphertext, nonce, err := authn.EncryptSecret(keys, secretBytes)
		if err != nil {
			t.Fatalf("encrypt totp secret: %v", err)
		}

		store := &authStoreStub{
			getSessionByFingerprintFunc: func(_ context.Context, fingerprint []byte) (authn.SessionRecord, authn.UserRecord, error) {
				if bytes.Equal(fingerprint, authn.FingerprintToken(keys, bootstrapToken)) {
					return authn.SessionRecord{}, authn.UserRecord{}, authn.ErrNotFound
				}
				t.Fatal("unexpected session fingerprint during bootstrap begin")
				return authn.SessionRecord{}, authn.UserRecord{}, authn.ErrNotFound
			},
			getBootstrapTokenByFingerprintFunc: func(_ context.Context, fingerprint []byte) (authn.BootstrapTokenRecord, authn.UserRecord, error) {
				if !bytes.Equal(fingerprint, authn.FingerprintToken(keys, bootstrapToken)) {
					t.Fatal("unexpected bootstrap fingerprint during bootstrap begin")
				}
				return authn.BootstrapTokenRecord{
					ID:        bootstrapTokenID,
					UserID:    userID,
					IssuedAt:  now.Add(-time.Minute),
					ExpiresAt: now.Add(9 * time.Minute),
				}, activeUserRecord(userID, "bootstrap@example.test"), nil
			},
			beginTOTPEnrollmentFunc: func(_ context.Context, gotUserID uuid.UUID, authScopeKind string, sessionID *uuid.UUID, gotBootstrapTokenID *uuid.UUID, clientTxnID string, _, _ []byte, replacesActive bool, createdAt time.Time) (authn.PendingTOTPEnrollmentRecord, bool, error) {
				if gotUserID != userID {
					t.Fatalf("unexpected bootstrap begin user: got %s want %s", gotUserID, userID)
				}
				if authScopeKind != "bootstrap_token" {
					t.Fatalf("unexpected bootstrap begin auth scope: got %q", authScopeKind)
				}
				if sessionID != nil {
					t.Fatalf("bootstrap begin must not carry session scope, got %v", *sessionID)
				}
				if gotBootstrapTokenID == nil || *gotBootstrapTokenID != bootstrapTokenID {
					t.Fatalf("unexpected bootstrap begin token scope: got %v want %s", gotBootstrapTokenID, bootstrapTokenID)
				}
				if clientTxnID != "txn-begin-bootstrap" {
					t.Fatalf("unexpected bootstrap begin client_txn_id: got %q", clientTxnID)
				}
				if replacesActive {
					t.Fatal("first bootstrap enrollment must not be marked as replacement")
				}
				if !createdAt.Equal(now) {
					t.Fatalf("unexpected bootstrap begin time: got %s want %s", createdAt, now)
				}
				return authn.PendingTOTPEnrollmentRecord{
					ID:               enrollmentID,
					UserID:           userID,
					AuthScopeKind:    authScopeKind,
					ClientTxnID:      clientTxnID,
					SecretCiphertext: ciphertext,
					SecretNonce:      nonce,
					CreatedAt:        now,
					ExpiresAt:        now.Add(10 * time.Minute),
				}, false, nil
			},
		}
		service := newUnitService(t, store, &hubStub{}, keys, now)

		recorder := httptest.NewRecorder()
		request := newJSONRequest(t, http.MethodPost, "/api/v1/auth/mfa/totp/begin", `{"client_txn_id":"txn-begin-bootstrap"}`)
		request.Header.Set("Authorization", "Bearer "+bootstrapToken)
		service.handleTOTPBegin(recorder, request)

		if recorder.Code != http.StatusOK {
			t.Fatalf("unexpected bootstrap begin status: got %d want %d", recorder.Code, http.StatusOK)
		}
		data := decodeSuccessData(t, recorder)
		if got := data["enrollment_id"]; got != enrollmentID.String() {
			t.Fatalf("unexpected enrollment id: got %v want %s", got, enrollmentID)
		}
		setup, ok := data["totp_setup"].(map[string]any)
		if !ok {
			t.Fatalf("expected totp_setup payload, got %#v", data["totp_setup"])
		}
		if got, ok := setup["secret_base32"].(string); !ok || got == "" {
			t.Fatalf("expected non-empty secret_base32 on begin, got %#v", setup["secret_base32"])
		}
		if got, ok := setup["otpauth_uri"].(string); !ok || !strings.Contains(got, "secret=") {
			t.Fatalf("unexpected otpauth_uri: %#v", setup["otpauth_uri"])
		}
	})

	t.Run("begin replays within one auth scope and rejects divergent replay conflicts", func(t *testing.T) {
		userID := uuid.MustParse("10000000-0000-0000-0000-000000000174")
		bootstrapTokenID := uuid.MustParse("10000000-0000-0000-0000-000000000175")
		bootstrapToken := "totp-begin-replay-bootstrap-token"
		enrollmentID := uuid.MustParse("10000000-0000-0000-0000-000000000176")
		beginCalls := 0
		var storedCiphertext []byte
		var storedNonce []byte
		store := &authStoreStub{
			getSessionByFingerprintFunc: func(_ context.Context, fingerprint []byte) (authn.SessionRecord, authn.UserRecord, error) {
				if bytes.Equal(fingerprint, authn.FingerprintToken(keys, bootstrapToken)) {
					return authn.SessionRecord{}, authn.UserRecord{}, authn.ErrNotFound
				}
				t.Fatal("unexpected session fingerprint during replay begin")
				return authn.SessionRecord{}, authn.UserRecord{}, authn.ErrNotFound
			},
			getBootstrapTokenByFingerprintFunc: func(_ context.Context, fingerprint []byte) (authn.BootstrapTokenRecord, authn.UserRecord, error) {
				if !bytes.Equal(fingerprint, authn.FingerprintToken(keys, bootstrapToken)) {
					t.Fatal("unexpected bootstrap fingerprint during replay begin")
				}
				return authn.BootstrapTokenRecord{
					ID:        bootstrapTokenID,
					UserID:    userID,
					IssuedAt:  now.Add(-time.Minute),
					ExpiresAt: now.Add(9 * time.Minute),
				}, activeUserRecord(userID, "replay-begin@example.test"), nil
			},
			beginTOTPEnrollmentFunc: func(_ context.Context, gotUserID uuid.UUID, authScopeKind string, sessionID *uuid.UUID, gotBootstrapTokenID *uuid.UUID, clientTxnID string, secretCiphertext []byte, secretNonce []byte, replacesActive bool, createdAt time.Time) (authn.PendingTOTPEnrollmentRecord, bool, error) {
				beginCalls++
				if gotUserID != userID || authScopeKind != "bootstrap_token" || sessionID != nil {
					t.Fatalf("unexpected replay begin scope: user=%s scope=%q session=%v", gotUserID, authScopeKind, sessionID)
				}
				if gotBootstrapTokenID == nil || *gotBootstrapTokenID != bootstrapTokenID {
					t.Fatalf("unexpected replay begin bootstrap token scope: got %v want %s", gotBootstrapTokenID, bootstrapTokenID)
				}
				if replacesActive {
					t.Fatal("bootstrap replay begin must not mark replacement enrollment")
				}
				if !createdAt.Equal(now) {
					t.Fatalf("unexpected replay begin time: got %s want %s", createdAt, now)
				}
				switch clientTxnID {
				case "txn-begin-replay":
					if beginCalls == 1 {
						storedCiphertext = append([]byte(nil), secretCiphertext...)
						storedNonce = append([]byte(nil), secretNonce...)
					}
					return authn.PendingTOTPEnrollmentRecord{
						ID:                        enrollmentID,
						UserID:                    userID,
						AuthScopeKind:             "bootstrap_token",
						AuthScopeBootstrapTokenID: &bootstrapTokenID,
						ClientTxnID:               clientTxnID,
						SecretCiphertext:          storedCiphertext,
						SecretNonce:               storedNonce,
						CreatedAt:                 now,
						ExpiresAt:                 now.Add(10 * time.Minute),
					}, beginCalls > 1, nil
				case "txn-begin-conflict":
					return authn.PendingTOTPEnrollmentRecord{}, false, authn.ErrClientTxnConflict
				default:
					t.Fatalf("unexpected begin client_txn_id: got %q", clientTxnID)
					return authn.PendingTOTPEnrollmentRecord{}, false, nil
				}
			},
		}
		service := newUnitService(t, store, &hubStub{}, keys, now)

		firstRecorder := httptest.NewRecorder()
		firstRequest := newJSONRequest(t, http.MethodPost, "/api/v1/auth/mfa/totp/begin", `{"client_txn_id":"txn-begin-replay"}`)
		firstRequest.Header.Set("Authorization", "Bearer "+bootstrapToken)
		service.handleTOTPBegin(firstRecorder, firstRequest)
		if firstRecorder.Code != http.StatusOK {
			t.Fatalf("unexpected first replay-begin status: got %d want %d", firstRecorder.Code, http.StatusOK)
		}
		firstData := decodeSuccessData(t, firstRecorder)
		firstSetup := firstData["totp_setup"].(map[string]any)

		secondRecorder := httptest.NewRecorder()
		secondRequest := newJSONRequest(t, http.MethodPost, "/api/v1/auth/mfa/totp/begin", `{"client_txn_id":"txn-begin-replay"}`)
		secondRequest.Header.Set("Authorization", "Bearer "+bootstrapToken)
		service.handleTOTPBegin(secondRecorder, secondRequest)
		if secondRecorder.Code != http.StatusOK {
			t.Fatalf("unexpected second replay-begin status: got %d want %d", secondRecorder.Code, http.StatusOK)
		}
		secondData := decodeSuccessData(t, secondRecorder)
		secondSetup := secondData["totp_setup"].(map[string]any)

		if firstData["enrollment_id"] != secondData["enrollment_id"] {
			t.Fatalf("begin replay must return the original enrollment_id: first=%v second=%v", firstData["enrollment_id"], secondData["enrollment_id"])
		}
		if firstSetup["secret_base32"] != secondSetup["secret_base32"] {
			t.Fatalf("begin replay must return the original setup secret: first=%v second=%v", firstSetup["secret_base32"], secondSetup["secret_base32"])
		}

		conflictRecorder := httptest.NewRecorder()
		conflictRequest := newJSONRequest(t, http.MethodPost, "/api/v1/auth/mfa/totp/begin", `{"client_txn_id":"txn-begin-conflict"}`)
		conflictRequest.Header.Set("Authorization", "Bearer "+bootstrapToken)
		service.handleTOTPBegin(conflictRecorder, conflictRequest)
		requireErrorEnvelope(t, conflictRecorder, http.StatusConflict, "client_txn_conflict", "")
	})

	t.Run("complete omits setup material and preserves bootstrap scope for first enrollment", func(t *testing.T) {
		userID := uuid.MustParse("10000000-0000-0000-0000-000000000181")
		bootstrapTokenID := uuid.MustParse("10000000-0000-0000-0000-000000000182")
		enrollmentID := uuid.MustParse("10000000-0000-0000-0000-000000000183")
		bootstrapToken := "totp-complete-bootstrap-token"
		secretBytes := []byte("01234567890123456789")
		secretBase32 := authn.EncodeSecretBase32(secretBytes)
		ciphertext, nonce, err := authn.EncryptSecret(keys, secretBytes)
		if err != nil {
			t.Fatalf("encrypt pending totp secret: %v", err)
		}
		code := generateTOTPCodeAt(t, secretBase32, now)

		store := &authStoreStub{
			getSessionByFingerprintFunc: func(_ context.Context, fingerprint []byte) (authn.SessionRecord, authn.UserRecord, error) {
				if bytes.Equal(fingerprint, authn.FingerprintToken(keys, bootstrapToken)) {
					return authn.SessionRecord{}, authn.UserRecord{}, authn.ErrNotFound
				}
				t.Fatal("unexpected session fingerprint during bootstrap complete")
				return authn.SessionRecord{}, authn.UserRecord{}, authn.ErrNotFound
			},
			getBootstrapTokenByFingerprintFunc: func(_ context.Context, fingerprint []byte) (authn.BootstrapTokenRecord, authn.UserRecord, error) {
				if !bytes.Equal(fingerprint, authn.FingerprintToken(keys, bootstrapToken)) {
					t.Fatal("unexpected bootstrap fingerprint during bootstrap complete")
				}
				return authn.BootstrapTokenRecord{
					ID:        bootstrapTokenID,
					UserID:    userID,
					IssuedAt:  now.Add(-time.Minute),
					ExpiresAt: now.Add(9 * time.Minute),
				}, activeUserRecord(userID, "complete@example.test"), nil
			},
			getPendingTOTPEnrollmentByIDFunc: func(_ context.Context, gotEnrollmentID uuid.UUID) (*authn.PendingTOTPEnrollmentRecord, error) {
				if gotEnrollmentID != enrollmentID {
					t.Fatalf("unexpected enrollment lookup id: got %s want %s", gotEnrollmentID, enrollmentID)
				}
				return &authn.PendingTOTPEnrollmentRecord{
					ID:                        enrollmentID,
					UserID:                    userID,
					AuthScopeKind:             "bootstrap_token",
					AuthScopeBootstrapTokenID: &bootstrapTokenID,
					ClientTxnID:               "txn-begin-bootstrap",
					SecretCiphertext:          ciphertext,
					SecretNonce:               nonce,
					CreatedAt:                 now,
					ExpiresAt:                 now.Add(10 * time.Minute),
				}, nil
			},
			activateTOTPEnrollmentFunc: func(_ context.Context, user authn.UserRecord, gotEnrollmentID uuid.UUID, authScopeKind string, sessionID *uuid.UUID, gotBootstrapTokenID *uuid.UUID, completedAt time.Time) (authn.TOTPCompleteResult, error) {
				if user.ID != userID {
					t.Fatalf("unexpected bootstrap complete user: got %s want %s", user.ID, userID)
				}
				if gotEnrollmentID != enrollmentID {
					t.Fatalf("unexpected bootstrap complete enrollment: got %s want %s", gotEnrollmentID, enrollmentID)
				}
				if authScopeKind != "bootstrap_token" {
					t.Fatalf("unexpected bootstrap complete auth scope: got %q", authScopeKind)
				}
				if sessionID != nil {
					t.Fatalf("bootstrap complete must not carry session scope, got %v", *sessionID)
				}
				if gotBootstrapTokenID == nil || *gotBootstrapTokenID != bootstrapTokenID {
					t.Fatalf("unexpected bootstrap complete token scope: got %v want %s", gotBootstrapTokenID, bootstrapTokenID)
				}
				if !completedAt.Equal(now) {
					t.Fatalf("unexpected bootstrap complete time: got %s want %s", completedAt, now)
				}
				return authn.TOTPCompleteResult{
					EnrolledAt:      completedAt,
					SessionsRevoked: false,
				}, nil
			},
		}
		service := newUnitService(t, store, &hubStub{}, keys, now)

		recorder := httptest.NewRecorder()
		request := newJSONRequest(t, http.MethodPost, "/api/v1/auth/mfa/totp/complete", `{
			"client_txn_id":"txn-complete-bootstrap",
			"enrollment_id":"`+enrollmentID.String()+`",
			"code":"`+code+`"
		}`)
		request.Header.Set("Authorization", "Bearer "+bootstrapToken)
		service.handleTOTPComplete(recorder, request)

		if recorder.Code != http.StatusOK {
			t.Fatalf("unexpected bootstrap complete status: got %d want %d", recorder.Code, http.StatusOK)
		}
		data := decodeSuccessData(t, recorder)
		if data["sessions_revoked"] != false {
			t.Fatalf("unexpected bootstrap complete response: %#v", data)
		}
		if _, ok := data["totp_setup"]; ok {
			t.Fatalf("totp complete must not return setup material, got %#v", data)
		}
	})

	t.Run("successful bootstrap completion consumes the bootstrap token for later requests", func(t *testing.T) {
		userID := uuid.MustParse("10000000-0000-0000-0000-000000000184")
		bootstrapTokenID := uuid.MustParse("10000000-0000-0000-0000-000000000185")
		enrollmentID := uuid.MustParse("10000000-0000-0000-0000-000000000186")
		bootstrapToken := "totp-complete-consume-bootstrap-token"
		secretBytes := []byte("klmnopqrst0123456789")
		secretBase32 := authn.EncodeSecretBase32(secretBytes)
		ciphertext, nonce, err := authn.EncryptSecret(keys, secretBytes)
		if err != nil {
			t.Fatalf("encrypt bootstrap-consumption secret: %v", err)
		}
		code := generateTOTPCodeAt(t, secretBase32, now)

		bootstrapLookups := 0
		store := &authStoreStub{
			getSessionByFingerprintFunc: func(_ context.Context, fingerprint []byte) (authn.SessionRecord, authn.UserRecord, error) {
				if bytes.Equal(fingerprint, authn.FingerprintToken(keys, bootstrapToken)) {
					return authn.SessionRecord{}, authn.UserRecord{}, authn.ErrNotFound
				}
				t.Fatal("unexpected session fingerprint during bootstrap-consumption test")
				return authn.SessionRecord{}, authn.UserRecord{}, authn.ErrNotFound
			},
			getBootstrapTokenByFingerprintFunc: func(_ context.Context, fingerprint []byte) (authn.BootstrapTokenRecord, authn.UserRecord, error) {
				if !bytes.Equal(fingerprint, authn.FingerprintToken(keys, bootstrapToken)) {
					t.Fatal("unexpected bootstrap fingerprint during bootstrap-consumption test")
				}
				bootstrapLookups++
				record := authn.BootstrapTokenRecord{
					ID:        bootstrapTokenID,
					UserID:    userID,
					IssuedAt:  now.Add(-time.Minute),
					ExpiresAt: now.Add(9 * time.Minute),
				}
				if bootstrapLookups > 1 {
					consumedAt := now
					record.ConsumedAt = &consumedAt
				}
				return record, activeUserRecord(userID, "consume-bootstrap@example.test"), nil
			},
			getPendingTOTPEnrollmentByIDFunc: func(_ context.Context, gotEnrollmentID uuid.UUID) (*authn.PendingTOTPEnrollmentRecord, error) {
				if gotEnrollmentID != enrollmentID {
					t.Fatalf("unexpected bootstrap-consumption enrollment lookup: got %s want %s", gotEnrollmentID, enrollmentID)
				}
				return &authn.PendingTOTPEnrollmentRecord{
					ID:                        enrollmentID,
					UserID:                    userID,
					AuthScopeKind:             "bootstrap_token",
					AuthScopeBootstrapTokenID: &bootstrapTokenID,
					ClientTxnID:               "txn-complete-consume",
					SecretCiphertext:          ciphertext,
					SecretNonce:               nonce,
					CreatedAt:                 now,
					ExpiresAt:                 now.Add(10 * time.Minute),
				}, nil
			},
			activateTOTPEnrollmentFunc: func(_ context.Context, user authn.UserRecord, gotEnrollmentID uuid.UUID, authScopeKind string, sessionID *uuid.UUID, gotBootstrapTokenID *uuid.UUID, completedAt time.Time) (authn.TOTPCompleteResult, error) {
				if user.ID != userID || gotEnrollmentID != enrollmentID || authScopeKind != "bootstrap_token" || sessionID != nil {
					t.Fatalf("unexpected bootstrap-consumption activation scope: user=%s enrollment=%s scope=%q session=%v", user.ID, gotEnrollmentID, authScopeKind, sessionID)
				}
				if gotBootstrapTokenID == nil || *gotBootstrapTokenID != bootstrapTokenID {
					t.Fatalf("unexpected bootstrap-consumption token scope: got %v want %s", gotBootstrapTokenID, bootstrapTokenID)
				}
				if !completedAt.Equal(now) {
					t.Fatalf("unexpected bootstrap-consumption complete time: got %s want %s", completedAt, now)
				}
				return authn.TOTPCompleteResult{
					EnrolledAt:      completedAt,
					SessionsRevoked: false,
				}, nil
			},
		}
		service := newUnitService(t, store, &hubStub{}, keys, now)

		completeRecorder := httptest.NewRecorder()
		completeRequest := newJSONRequest(t, http.MethodPost, "/api/v1/auth/mfa/totp/complete", `{
			"client_txn_id":"txn-complete-consume",
			"enrollment_id":"`+enrollmentID.String()+`",
			"code":"`+code+`"
		}`)
		completeRequest.Header.Set("Authorization", "Bearer "+bootstrapToken)
		service.handleTOTPComplete(completeRecorder, completeRequest)
		if completeRecorder.Code != http.StatusOK {
			t.Fatalf("unexpected bootstrap-consumption complete status: got %d want %d", completeRecorder.Code, http.StatusOK)
		}

		beginRecorder := httptest.NewRecorder()
		beginRequest := newJSONRequest(t, http.MethodPost, "/api/v1/auth/mfa/totp/begin", `{"client_txn_id":"txn-begin-after-consume"}`)
		beginRequest.Header.Set("Authorization", "Bearer "+bootstrapToken)
		service.handleTOTPBegin(beginRecorder, beginRequest)

		requireErrorEnvelope(t, beginRecorder, http.StatusConflict, "credential_bootstrap_rejected", "")
		details := decodeErrorDetails(t, beginRecorder)
		if got := details["reason_code"]; got != "consumed" {
			t.Fatalf("unexpected bootstrap-consumption reason_code: got %v want consumed", got)
		}
	})

	t.Run("replacement completion revokes returned sessions and clears the current cookie", func(t *testing.T) {
		userID := uuid.MustParse("10000000-0000-0000-0000-000000000191")
		currentSessionID := uuid.MustParse("10000000-0000-0000-0000-000000000192")
		otherSessionID := uuid.MustParse("10000000-0000-0000-0000-000000000193")
		enrollmentID := uuid.MustParse("10000000-0000-0000-0000-000000000194")
		token := "totp-complete-session-token"
		secretBytes := []byte("abcdefghij0123456789")
		secretBase32 := authn.EncodeSecretBase32(secretBytes)
		ciphertext, nonce, err := authn.EncryptSecret(keys, secretBytes)
		if err != nil {
			t.Fatalf("encrypt replacement totp secret: %v", err)
		}
		code := generateTOTPCodeAt(t, secretBase32, now)

		store := &authStoreStub{
			getSessionByFingerprintFunc: func(_ context.Context, fingerprint []byte) (authn.SessionRecord, authn.UserRecord, error) {
				if !bytes.Equal(fingerprint, authn.FingerprintToken(keys, token)) {
					t.Fatal("unexpected session fingerprint during replacement complete")
				}
				return activeSessionRecord(currentSessionID, userID, now), activeUserRecord(userID, "replace@example.test"), nil
			},
			getPendingTOTPEnrollmentByIDFunc: func(_ context.Context, gotEnrollmentID uuid.UUID) (*authn.PendingTOTPEnrollmentRecord, error) {
				if gotEnrollmentID != enrollmentID {
					t.Fatalf("unexpected replacement enrollment lookup: got %s want %s", gotEnrollmentID, enrollmentID)
				}
				return &authn.PendingTOTPEnrollmentRecord{
					ID:                 enrollmentID,
					UserID:             userID,
					AuthScopeKind:      "session",
					AuthScopeSessionID: &currentSessionID,
					SecretCiphertext:   ciphertext,
					SecretNonce:        nonce,
					ReplacesActive:     true,
					CreatedAt:          now,
					ExpiresAt:          now.Add(10 * time.Minute),
				}, nil
			},
			activateTOTPEnrollmentFunc: func(_ context.Context, user authn.UserRecord, gotEnrollmentID uuid.UUID, authScopeKind string, sessionID *uuid.UUID, bootstrapTokenID *uuid.UUID, completedAt time.Time) (authn.TOTPCompleteResult, error) {
				if user.ID != userID {
					t.Fatalf("unexpected replacement complete user: got %s want %s", user.ID, userID)
				}
				if gotEnrollmentID != enrollmentID {
					t.Fatalf("unexpected replacement complete enrollment: got %s want %s", gotEnrollmentID, enrollmentID)
				}
				if authScopeKind != "session" {
					t.Fatalf("unexpected replacement complete auth scope: got %q", authScopeKind)
				}
				if sessionID == nil || *sessionID != currentSessionID {
					t.Fatalf("unexpected replacement complete session scope: got %v want %s", sessionID, currentSessionID)
				}
				if bootstrapTokenID != nil {
					t.Fatalf("replacement complete must not carry bootstrap scope, got %v", *bootstrapTokenID)
				}
				if !completedAt.Equal(now) {
					t.Fatalf("unexpected replacement complete time: got %s want %s", completedAt, now)
				}
				return authn.TOTPCompleteResult{
					EnrolledAt:        completedAt,
					SessionsRevoked:   true,
					RevokedSessionIDs: []uuid.UUID{currentSessionID, otherSessionID},
				}, nil
			},
		}
		hub := &hubStub{}
		service := newUnitService(t, store, hub, keys, now)

		recorder := httptest.NewRecorder()
		request := newJSONRequest(t, http.MethodPost, "/api/v1/auth/mfa/totp/complete", `{
			"client_txn_id":"txn-complete-replacement",
			"enrollment_id":"`+enrollmentID.String()+`",
			"code":"`+code+`"
		}`)
		addSessionAuth(request, keys, token, true)
		service.handleTOTPComplete(recorder, request)

		if recorder.Code != http.StatusOK {
			t.Fatalf("unexpected replacement complete status: got %d want %d", recorder.Code, http.StatusOK)
		}
		requireRevocations(t, hub.revocations, []revocationCall{
			{sessionID: currentSessionID, reasonCode: "session_revoked"},
			{sessionID: otherSessionID, reasonCode: "session_revoked"},
		})
		data := decodeSuccessData(t, recorder)
		if data["sessions_revoked"] != true {
			t.Fatalf("unexpected replacement complete response: %#v", data)
		}
		if _, ok := data["totp_setup"]; ok {
			t.Fatalf("totp complete must not return setup material, got %#v", data)
		}
		requireCookieCleared(t, recorder.Result().Cookies(), authn.SessionCookieName)
		requireCookieCleared(t, recorder.Result().Cookies(), authn.CSRFCookieName)
	})
}

func TestPhase1_AdminCredentialActionGuards_U_1_13(t *testing.T) {
	now := time.Date(2026, time.April, 17, 15, 0, 0, 0, time.UTC)
	keys := loadUnitMasterKeys(t)
	targetUserID := uuid.MustParse("10000000-0000-0000-0000-000000000201")

	t.Run("incident admins do not gain deployment-admin credential routes", func(t *testing.T) {
		actorID := uuid.MustParse("10000000-0000-0000-0000-000000000202")
		sessionID := uuid.MustParse("10000000-0000-0000-0000-000000000203")
		token := "incident-admin-token"

		passwordResetCalls := 0
		totpResetCalls := 0
		revokeAllCalls := 0
		store := &authStoreStub{
			getSessionByFingerprintFunc: func(_ context.Context, fingerprint []byte) (authn.SessionRecord, authn.UserRecord, error) {
				if !bytes.Equal(fingerprint, authn.FingerprintToken(keys, token)) {
					t.Fatal("unexpected session fingerprint for denied admin action")
				}
				user := activeUserRecord(actorID, "incident-admin@example.test")
				user.IsDeploymentAdmin = false
				return activeSessionRecord(sessionID, actorID, now), user, nil
			},
			adminResetPasswordFunc: func(context.Context, authn.UserRecord, uuid.UUID, int64, string, string, []byte, string, time.Time) (authn.AdminPasswordResetResult, error) {
				passwordResetCalls++
				return authn.AdminPasswordResetResult{}, nil
			},
			adminResetTOTPFunc: func(context.Context, authn.UserRecord, uuid.UUID, int64, string, []byte, string, time.Time) (authn.AdminTOTPResetResult, error) {
				totpResetCalls++
				return authn.AdminTOTPResetResult{}, nil
			},
			adminRevokeAllSessionsFunc: func(context.Context, authn.UserRecord, uuid.UUID, string, []byte, string, time.Time) (authn.AdminRevokeAllResult, error) {
				revokeAllCalls++
				return authn.AdminRevokeAllResult{}, nil
			},
		}
		service := newUnitService(t, store, &hubStub{}, keys, now)

		for _, tc := range []struct {
			name string
			path string
			body string
		}{
			{
				name: "password reset",
				path: "/api/v1/users/" + targetUserID.String() + "/password/reset",
				body: `{"base_user_version":1,"client_txn_id":"txn-reset","new_password":"Replacement passphrase 1"}`,
			},
			{
				name: "totp reset",
				path: "/api/v1/users/" + targetUserID.String() + "/mfa/totp/reset",
				body: `{"base_user_version":1,"client_txn_id":"txn-totp-reset"}`,
			},
			{
				name: "revoke all",
				path: "/api/v1/users/" + targetUserID.String() + "/sessions/revoke-all",
				body: `{"client_txn_id":"txn-revoke-all"}`,
			},
		} {
			t.Run(tc.name, func(t *testing.T) {
				recorder := httptest.NewRecorder()
				request := newJSONRequest(t, http.MethodPost, tc.path, tc.body)
				addSessionAuth(request, keys, token, true)
				service.handleUsersMember(recorder, request)
				requireErrorEnvelope(t, recorder, http.StatusUnauthorized, unauthorizedCode, "")
			})
		}

		if passwordResetCalls != 0 || totpResetCalls != 0 || revokeAllCalls != 0 {
			t.Fatalf("admin credential stores must not run for incident admins: password=%d totp=%d revoke_all=%d", passwordResetCalls, totpResetCalls, revokeAllCalls)
		}
	})

	t.Run("deployment admins can execute each credential route and preserve route-local consequences", func(t *testing.T) {
		actorID := uuid.MustParse("10000000-0000-0000-0000-000000000211")
		currentSessionID := uuid.MustParse("10000000-0000-0000-0000-000000000212")
		otherSessionID := uuid.MustParse("10000000-0000-0000-0000-000000000213")
		token := "deployment-admin-token"
		adminUser := deploymentAdminUserRecord(actorID, "deployment-admin@example.test")
		session := activeSessionRecord(currentSessionID, actorID, now)

		passwordResetCalled := false
		totpResetCalled := false
		revokeAllCalled := false
		store := &authStoreStub{
			getSessionByFingerprintFunc: func(_ context.Context, fingerprint []byte) (authn.SessionRecord, authn.UserRecord, error) {
				if !bytes.Equal(fingerprint, authn.FingerprintToken(keys, token)) {
					t.Fatal("unexpected session fingerprint for allowed admin action")
				}
				return session, adminUser, nil
			},
			adminResetPasswordFunc: func(_ context.Context, actor authn.UserRecord, gotTargetUserID uuid.UUID, baseUserVersion int64, newPasswordHash string, clientTxnID string, requestHash []byte, requestID string, changedAt time.Time) (authn.AdminPasswordResetResult, error) {
				passwordResetCalled = true
				if actor.ID != actorID || gotTargetUserID != targetUserID || baseUserVersion != 4 || clientTxnID != "txn-reset-allowed" {
					t.Fatalf("unexpected password reset parameters: actor=%s target=%s version=%d txn=%q", actor.ID, gotTargetUserID, baseUserVersion, clientTxnID)
				}
				if len(requestHash) == 0 || requestID != "" || !changedAt.Equal(now) {
					t.Fatalf("unexpected password reset routing metadata: hash=%x request_id=%q changed_at=%s", requestHash, requestID, changedAt)
				}
				if ok, err := authn.VerifyPasswordHash(newPasswordHash, "Replacement passphrase 1"); err != nil || !ok {
					t.Fatalf("password reset hash must verify replacement password, ok=%v err=%v", ok, err)
				}
				return authn.AdminPasswordResetResult{
					RevokedSessionIDs: []uuid.UUID{currentSessionID, otherSessionID},
					ResponseJSON: mustJSON(t, map[string]any{
						"user_id":             targetUserID,
						"email":               "target@example.test",
						"display_name":        "Target User",
						"is_active":           true,
						"mfa_required":        true,
						"is_deployment_admin": false,
						"user_version":        5,
						"auth_bindings": []map[string]any{
							{
								"provider_type": "local",
								"provider_key":  "local",
								"username":      "target@example.test",
								"created_at":    now,
							},
						},
					}),
				}, nil
			},
			adminResetTOTPFunc: func(_ context.Context, actor authn.UserRecord, gotTargetUserID uuid.UUID, baseUserVersion int64, clientTxnID string, requestHash []byte, requestID string, changedAt time.Time) (authn.AdminTOTPResetResult, error) {
				totpResetCalled = true
				if actor.ID != actorID || gotTargetUserID != targetUserID || baseUserVersion != 5 || clientTxnID != "txn-totp-reset-allowed" {
					t.Fatalf("unexpected totp reset parameters: actor=%s target=%s version=%d txn=%q", actor.ID, gotTargetUserID, baseUserVersion, clientTxnID)
				}
				if len(requestHash) == 0 || requestID != "" || !changedAt.Equal(now) {
					t.Fatalf("unexpected totp reset routing metadata: hash=%x request_id=%q changed_at=%s", requestHash, requestID, changedAt)
				}
				return authn.AdminTOTPResetResult{
					RevokedSessionIDs: []uuid.UUID{currentSessionID, otherSessionID},
					ResponseJSON: mustJSON(t, map[string]any{
						"user_id":             targetUserID,
						"email":               "target@example.test",
						"display_name":        "Target User",
						"is_active":           true,
						"mfa_required":        true,
						"is_deployment_admin": false,
						"user_version":        6,
						"auth_bindings": []map[string]any{
							{
								"provider_type": "local",
								"provider_key":  "local",
								"username":      "target@example.test",
								"created_at":    now,
							},
						},
					}),
				}, nil
			},
			adminRevokeAllSessionsFunc: func(_ context.Context, actor authn.UserRecord, gotTargetUserID uuid.UUID, clientTxnID string, requestHash []byte, requestID string, revokedAt time.Time) (authn.AdminRevokeAllResult, error) {
				revokeAllCalled = true
				if actor.ID != actorID || gotTargetUserID != targetUserID || clientTxnID != "txn-revoke-all-allowed" {
					t.Fatalf("unexpected revoke-all parameters: actor=%s target=%s txn=%q", actor.ID, gotTargetUserID, clientTxnID)
				}
				if len(requestHash) == 0 || requestID != "" || !revokedAt.Equal(now) {
					t.Fatalf("unexpected revoke-all routing metadata: hash=%x request_id=%q revoked_at=%s", requestHash, requestID, revokedAt)
				}
				return authn.AdminRevokeAllResult{
					RevokedAt:         revokedAt,
					RevokedSessionIDs: []uuid.UUID{currentSessionID, otherSessionID},
					ResponseJSON: mustJSON(t, map[string]any{
						"user_id":          targetUserID,
						"sessions_revoked": true,
						"revoked_at":       revokedAt,
					}),
				}, nil
			},
		}
		hub := &hubStub{}
		service := newUnitService(t, store, hub, keys, now)

		for _, tc := range []struct {
			name string
			path string
			body string
		}{
			{
				name: "password reset",
				path: "/api/v1/users/" + targetUserID.String() + "/password/reset",
				body: `{"base_user_version":4,"client_txn_id":"txn-reset-allowed","new_password":"Replacement passphrase 1"}`,
			},
			{
				name: "totp reset",
				path: "/api/v1/users/" + targetUserID.String() + "/mfa/totp/reset",
				body: `{"base_user_version":5,"client_txn_id":"txn-totp-reset-allowed","reason":"reset totp"}`,
			},
			{
				name: "revoke all",
				path: "/api/v1/users/" + targetUserID.String() + "/sessions/revoke-all",
				body: `{"client_txn_id":"txn-revoke-all-allowed","reason":"incident response"}`,
			},
		} {
			t.Run(tc.name, func(t *testing.T) {
				hub.revocations = nil
				recorder := httptest.NewRecorder()
				request := newJSONRequest(t, http.MethodPost, tc.path, tc.body)
				addSessionAuth(request, keys, token, true)
				service.handleUsersMember(recorder, request)
				if recorder.Code != http.StatusOK {
					t.Fatalf("unexpected allowed admin action status: got %d want %d", recorder.Code, http.StatusOK)
				}
				requireRevocations(t, hub.revocations, []revocationCall{
					{sessionID: currentSessionID, reasonCode: "session_revoked"},
					{sessionID: otherSessionID, reasonCode: "session_revoked"},
				})
				requireCookieCleared(t, recorder.Result().Cookies(), authn.SessionCookieName)
				requireCookieCleared(t, recorder.Result().Cookies(), authn.CSRFCookieName)
			})
		}

		if !passwordResetCalled || !totpResetCalled || !revokeAllCalled {
			t.Fatalf("expected all admin credential actions to run, password=%v totp=%v revoke_all=%v", passwordResetCalled, totpResetCalled, revokeAllCalled)
		}
	})
}

type authStoreStub struct {
	getUserByNormalizedEmailFunc        func(context.Context, string) (authn.UserRecord, error)
	getUserByIDFunc                     func(context.Context, uuid.UUID) (authn.UserRecord, error)
	listIncidentMembershipSummariesFunc func(context.Context, uuid.UUID) ([]authn.IncidentMembershipSummary, error)
	getSessionByFingerprintFunc         func(context.Context, []byte) (authn.SessionRecord, authn.UserRecord, error)
	createSessionWithConcurrencyFunc    func(context.Context, authn.UserRecord, []byte, authn.SessionTiming, string) (authn.SessionRecord, *authn.SessionRecord, error)
	slideSessionFunc                    func(context.Context, uuid.UUID, authn.SessionTiming) (authn.SessionTiming, error)
	revokeSessionFunc                   func(context.Context, uuid.UUID, string, time.Time) error
	issueBootstrapTokenFunc             func(context.Context, uuid.UUID, []byte, time.Time) (authn.BootstrapTokenRecord, error)
	getBootstrapTokenByFingerprintFunc  func(context.Context, []byte) (authn.BootstrapTokenRecord, authn.UserRecord, error)
	getPendingTOTPEnrollmentForUserFunc func(context.Context, uuid.UUID, time.Time) (*authn.PendingTOTPEnrollmentRecord, error)
	getPendingTOTPEnrollmentByIDFunc    func(context.Context, uuid.UUID) (*authn.PendingTOTPEnrollmentRecord, error)
	beginTOTPEnrollmentFunc             func(context.Context, uuid.UUID, string, *uuid.UUID, *uuid.UUID, string, []byte, []byte, bool, time.Time) (authn.PendingTOTPEnrollmentRecord, bool, error)
	activateTOTPEnrollmentFunc          func(context.Context, authn.UserRecord, uuid.UUID, string, *uuid.UUID, *uuid.UUID, time.Time) (authn.TOTPCompleteResult, error)
	getRouteIdempotencyFunc             func(context.Context, authn.RouteIdempotencyKey) (authn.RouteIdempotencyRecord, error)
	changePasswordFunc                  func(context.Context, authn.UserRecord, string, []byte, string, string, time.Time) (authn.PasswordChangeResult, error)
	listUsersFunc                       func(context.Context) ([]authn.UserRecord, error)
	createUserFunc                      func(context.Context, authn.UserRecord, string, string, string, bool, bool, string, []byte, string, time.Time) (authn.UserCreateResult, error)
	updateUserFunc                      func(context.Context, authn.UserRecord, uuid.UUID, int64, *string, *string, *bool, *bool, *bool, string, time.Time) (authn.UserRecord, []uuid.UUID, error)
	adminResetPasswordFunc              func(context.Context, authn.UserRecord, uuid.UUID, int64, string, string, []byte, string, time.Time) (authn.AdminPasswordResetResult, error)
	adminResetTOTPFunc                  func(context.Context, authn.UserRecord, uuid.UUID, int64, string, []byte, string, time.Time) (authn.AdminTOTPResetResult, error)
	adminRevokeAllSessionsFunc          func(context.Context, authn.UserRecord, uuid.UUID, string, []byte, string, time.Time) (authn.AdminRevokeAllResult, error)
}

func (s *authStoreStub) GetUserByNormalizedEmail(ctx context.Context, email string) (authn.UserRecord, error) {
	return callStub1(s.getUserByNormalizedEmailFunc, ctx, email)
}

func (s *authStoreStub) GetUserByID(ctx context.Context, userID uuid.UUID) (authn.UserRecord, error) {
	return callStub1(s.getUserByIDFunc, ctx, userID)
}

func (s *authStoreStub) ListIncidentMembershipSummaries(ctx context.Context, userID uuid.UUID) ([]authn.IncidentMembershipSummary, error) {
	return callStub1(s.listIncidentMembershipSummariesFunc, ctx, userID)
}

func (s *authStoreStub) GetSessionByFingerprint(ctx context.Context, fingerprint []byte) (authn.SessionRecord, authn.UserRecord, error) {
	return callStub1Result2(s.getSessionByFingerprintFunc, ctx, fingerprint)
}

func (s *authStoreStub) CreateSessionWithConcurrency(ctx context.Context, user authn.UserRecord, fingerprint []byte, timing authn.SessionTiming, requestID string) (authn.SessionRecord, *authn.SessionRecord, error) {
	return callStub4Result2(s.createSessionWithConcurrencyFunc, ctx, user, fingerprint, timing, requestID)
}

func (s *authStoreStub) SlideSession(ctx context.Context, sessionID uuid.UUID, timing authn.SessionTiming) (authn.SessionTiming, error) {
	return callStub2(s.slideSessionFunc, ctx, sessionID, timing)
}

func (s *authStoreStub) RevokeSession(ctx context.Context, sessionID uuid.UUID, reasonCode string, now time.Time) error {
	return callStubNoResult3(s.revokeSessionFunc, ctx, sessionID, reasonCode, now)
}

func (s *authStoreStub) IssueBootstrapToken(ctx context.Context, userID uuid.UUID, fingerprint []byte, now time.Time) (authn.BootstrapTokenRecord, error) {
	return callStub3(s.issueBootstrapTokenFunc, ctx, userID, fingerprint, now)
}

func (s *authStoreStub) GetBootstrapTokenByFingerprint(ctx context.Context, fingerprint []byte) (authn.BootstrapTokenRecord, authn.UserRecord, error) {
	return callStub1Result2(s.getBootstrapTokenByFingerprintFunc, ctx, fingerprint)
}

func (s *authStoreStub) GetPendingTOTPEnrollmentForUser(ctx context.Context, userID uuid.UUID, now time.Time) (*authn.PendingTOTPEnrollmentRecord, error) {
	return callStub2Ptr(s.getPendingTOTPEnrollmentForUserFunc, ctx, userID, now)
}

func (s *authStoreStub) GetPendingTOTPEnrollmentByID(ctx context.Context, enrollmentID uuid.UUID) (*authn.PendingTOTPEnrollmentRecord, error) {
	return callStub1(s.getPendingTOTPEnrollmentByIDFunc, ctx, enrollmentID)
}

func (s *authStoreStub) BeginTOTPEnrollment(ctx context.Context, userID uuid.UUID, authScopeKind string, sessionID *uuid.UUID, bootstrapTokenID *uuid.UUID, clientTxnID string, secretCiphertext []byte, secretNonce []byte, replacesActive bool, now time.Time) (authn.PendingTOTPEnrollmentRecord, bool, error) {
	return callStub9Result2(s.beginTOTPEnrollmentFunc, ctx, userID, authScopeKind, sessionID, bootstrapTokenID, clientTxnID, secretCiphertext, secretNonce, replacesActive, now)
}

func (s *authStoreStub) ActivateTOTPEnrollment(ctx context.Context, user authn.UserRecord, enrollmentID uuid.UUID, authScopeKind string, sessionID *uuid.UUID, bootstrapTokenID *uuid.UUID, now time.Time) (authn.TOTPCompleteResult, error) {
	return callStub6(s.activateTOTPEnrollmentFunc, ctx, user, enrollmentID, authScopeKind, sessionID, bootstrapTokenID, now)
}

func (s *authStoreStub) GetRouteIdempotency(ctx context.Context, key authn.RouteIdempotencyKey) (authn.RouteIdempotencyRecord, error) {
	return callStub1(s.getRouteIdempotencyFunc, ctx, key)
}

func (s *authStoreStub) ChangePassword(ctx context.Context, user authn.UserRecord, clientTxnID string, requestHash []byte, newPasswordHash string, requestID string, now time.Time) (authn.PasswordChangeResult, error) {
	return callStub6(s.changePasswordFunc, ctx, user, clientTxnID, requestHash, newPasswordHash, requestID, now)
}

func (s *authStoreStub) ListUsers(ctx context.Context) ([]authn.UserRecord, error) {
	if s.listUsersFunc == nil {
		var zero []authn.UserRecord
		return zero, nil
	}
	return s.listUsersFunc(ctx)
}

func (s *authStoreStub) CreateUser(ctx context.Context, actor authn.UserRecord, email string, displayName string, passwordHash string, mfaRequired bool, isDeploymentAdmin bool, clientTxnID string, requestHash []byte, requestID string, now time.Time) (authn.UserCreateResult, error) {
	return callStub10(s.createUserFunc, ctx, actor, email, displayName, passwordHash, mfaRequired, isDeploymentAdmin, clientTxnID, requestHash, requestID, now)
}

func (s *authStoreStub) UpdateUser(ctx context.Context, actor authn.UserRecord, targetUserID uuid.UUID, baseUserVersion int64, email *string, displayName *string, isActive *bool, mfaRequired *bool, isDeploymentAdmin *bool, requestID string, now time.Time) (authn.UserRecord, []uuid.UUID, error) {
	return callStub10Result2(s.updateUserFunc, ctx, actor, targetUserID, baseUserVersion, email, displayName, isActive, mfaRequired, isDeploymentAdmin, requestID, now)
}

func (s *authStoreStub) AdminResetPassword(ctx context.Context, actor authn.UserRecord, targetUserID uuid.UUID, baseUserVersion int64, newPasswordHash string, clientTxnID string, requestHash []byte, requestID string, now time.Time) (authn.AdminPasswordResetResult, error) {
	return callStub8(s.adminResetPasswordFunc, ctx, actor, targetUserID, baseUserVersion, newPasswordHash, clientTxnID, requestHash, requestID, now)
}

func (s *authStoreStub) AdminResetTOTP(ctx context.Context, actor authn.UserRecord, targetUserID uuid.UUID, baseUserVersion int64, clientTxnID string, requestHash []byte, requestID string, now time.Time) (authn.AdminTOTPResetResult, error) {
	return callStub7(s.adminResetTOTPFunc, ctx, actor, targetUserID, baseUserVersion, clientTxnID, requestHash, requestID, now)
}

func (s *authStoreStub) AdminRevokeAllSessions(ctx context.Context, actor authn.UserRecord, targetUserID uuid.UUID, clientTxnID string, requestHash []byte, requestID string, now time.Time) (authn.AdminRevokeAllResult, error) {
	return callStub6(s.adminRevokeAllSessionsFunc, ctx, actor, targetUserID, clientTxnID, requestHash, requestID, now)
}

type hubStub struct {
	revocations []revocationCall
}

type revocationCall struct {
	sessionID  uuid.UUID
	reasonCode string
}

func (h *hubStub) RegisterSession(uuid.UUID) (<-chan string, func()) {
	return nil, func() {}
}

func (h *hubStub) RevokeSession(sessionID uuid.UUID, reasonCode string) {
	h.revocations = append(h.revocations, revocationCall{sessionID: sessionID, reasonCode: reasonCode})
}

func newUnitService(t testing.TB, store authStore, hub sessionHub, keys authn.MasterKeys, now time.Time) *Service {
	t.Helper()
	return &Service{
		store:      store,
		hub:        hub,
		keys:       keys,
		pagination: pagination.NewRegistry(),
		now:        func() time.Time { return now },
	}
}

func loadUnitMasterKeys(t testing.TB) authn.MasterKeys {
	t.Helper()
	keys, err := authn.LoadMasterKeys(nil)
	if err != nil {
		t.Fatalf("load master keys: %v", err)
	}
	return keys
}

func newJSONRequest(t testing.TB, method string, target string, body string) *http.Request {
	t.Helper()
	request := httptest.NewRequest(method, target, strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	return request
}

func addSessionAuth(request *http.Request, keys authn.MasterKeys, sessionToken string, stateChanging bool) {
	addSessionCookiesOnly(request, keys, sessionToken)
	if stateChanging {
		request.Header.Set(authn.CSRFHeaderName, authn.CSRFTokenForSessionToken(keys, sessionToken))
	}
}

func addSessionCookiesOnly(request *http.Request, keys authn.MasterKeys, sessionToken string) {
	request.AddCookie(&http.Cookie{Name: authn.SessionCookieName, Value: sessionToken})
	request.AddCookie(&http.Cookie{Name: authn.CSRFCookieName, Value: authn.CSRFTokenForSessionToken(keys, sessionToken)})
}

func decodeSuccessData(t testing.TB, recorder *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var envelope struct {
		Data map[string]any `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode success envelope: %v", err)
	}
	return envelope.Data
}

func requireErrorEnvelope(t testing.TB, recorder *httptest.ResponseRecorder, wantStatus int, wantCode string, wantField string) {
	t.Helper()
	if recorder.Code != wantStatus {
		t.Fatalf("unexpected error status: got %d want %d body=%s", recorder.Code, wantStatus, recorder.Body.String())
	}
	var envelope struct {
		Error struct {
			Code    string         `json:"code"`
			Details map[string]any `json:"details"`
		} `json:"error"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode error envelope: %v", err)
	}
	if envelope.Error.Code != wantCode {
		t.Fatalf("unexpected error code: got %q want %q", envelope.Error.Code, wantCode)
	}
	if wantField == "" {
		return
	}
	if got := envelope.Error.Details["field"]; got != wantField {
		t.Fatalf("unexpected error field: got %v want %s", got, wantField)
	}
}

func decodeErrorDetails(t testing.TB, recorder *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var envelope struct {
		Error struct {
			Details map[string]any `json:"details"`
		} `json:"error"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode error envelope details: %v", err)
	}
	return envelope.Error.Details
}

func requireCookieCleared(t testing.TB, cookies []*http.Cookie, name string) {
	t.Helper()
	cookie := findCookie(cookies, name)
	if cookie == nil || cookie.Value != "" || cookie.MaxAge >= 0 {
		t.Fatalf("expected %s cookie to be cleared, got %#v", name, cookie)
	}
}

func findCookie(cookies []*http.Cookie, name string) *http.Cookie {
	for _, cookie := range cookies {
		if cookie.Name == name {
			return cookie
		}
	}
	return nil
}

func requireRevocations(t testing.TB, got []revocationCall, want []revocationCall) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("unexpected revocation count: got %#v want %#v", got, want)
	}
	for index := range want {
		if got[index] != want[index] {
			t.Fatalf("unexpected revocation at %d: got %#v want %#v", index, got[index], want[index])
		}
	}
}

func phase1RouteCSRFPayload(t testing.TB, route phase1test.RouteInventoryEntry) string {
	t.Helper()

	switch route.ID {
	case phase1test.RouteLogout:
		return `{}`
	case phase1test.RoutePasswordChange:
		return `{
			"client_txn_id":"txn-password-csrf",
			"current_password":"Phase1CSRFCurrent!",
			"new_password":"Phase1CSRFFresh!"
		}`
	case phase1test.RouteTOTPBegin:
		return `{"client_txn_id":"txn-totp-begin-csrf"}`
	case phase1test.RouteTOTPComplete:
		return `{
			"client_txn_id":"txn-totp-complete-csrf",
			"enrollment_id":"10000000-0000-0000-0000-000000000129",
			"code":"123456"
		}`
	case phase1test.RouteUsersCreate:
		return `{
			"client_txn_id":"txn-user-create-csrf",
			"auth_kind":"local",
			"email":"csrf-route@example.test",
			"display_name":"CSRF Route",
			"initial_password":"Phase1CSRFFresh!"
		}`
	case phase1test.RouteUsersPatch:
		return `{
			"base_user_version":1,
			"display_name":"CSRF Patch"
		}`
	case phase1test.RouteUsersPasswordReset:
		return `{
			"base_user_version":1,
			"client_txn_id":"txn-user-password-reset-csrf",
			"new_password":"Phase1CSRFFresh!"
		}`
	case phase1test.RouteUsersTOTPReset:
		return `{
			"base_user_version":1,
			"client_txn_id":"txn-user-totp-reset-csrf"
		}`
	case phase1test.RouteUsersRevokeAll:
		return `{
			"client_txn_id":"txn-user-revoke-all-csrf",
			"reason":"csrf guard"
		}`
	default:
		t.Fatalf("missing CSRF test payload for route %s", route.ID)
		return ""
	}
}

func dispatchPhase1UnitRoute(t testing.TB, service *Service, route phase1test.RouteInventoryEntry, recorder *httptest.ResponseRecorder, request *http.Request) {
	t.Helper()

	switch route.ID {
	case phase1test.RouteLogout:
		service.handleLogout(recorder, request)
	case phase1test.RoutePasswordChange:
		service.handlePasswordChange(recorder, request)
	case phase1test.RouteTOTPBegin:
		service.handleTOTPBegin(recorder, request)
	case phase1test.RouteTOTPComplete:
		service.handleTOTPComplete(recorder, request)
	case phase1test.RouteUsersCreate:
		service.handleUsersCollection(recorder, request)
	case phase1test.RouteUsersPatch, phase1test.RouteUsersPasswordReset, phase1test.RouteUsersTOTPReset, phase1test.RouteUsersRevokeAll:
		service.handleUsersMember(recorder, request)
	default:
		t.Fatalf("missing unit route dispatcher for %s", route.ID)
	}
}

func activeSessionRecord(sessionID uuid.UUID, userID uuid.UUID, now time.Time) authn.SessionRecord {
	timing := authn.NewSessionTiming(now)
	return authn.SessionRecord{
		ID:                       sessionID,
		UserID:                   userID,
		AuthenticatedAt:          timing.AuthenticatedAt,
		LastQualifyingActivityAt: timing.LastQualifyingActivityAt,
		IdleExpiresAt:            timing.IdleExpiresAt,
		AbsoluteExpiresAt:        timing.AbsoluteExpiresAt,
		SessionExpiresAt:         timing.SessionExpiresAt,
		CreatedAt:                timing.AuthenticatedAt,
		UpdatedAt:                timing.AuthenticatedAt,
	}
}

func activeUserRecord(userID uuid.UUID, email string) authn.UserRecord {
	return authn.UserRecord{
		ID:          userID,
		Email:       email,
		DisplayName: "User",
		IsActive:    true,
	}
}

func deploymentAdminUserRecord(userID uuid.UUID, email string) authn.UserRecord {
	user := activeUserRecord(userID, email)
	user.IsDeploymentAdmin = true
	return user
}

func activeTOTPUserRecord(t testing.TB, keys authn.MasterKeys, userID uuid.UUID, email string, password string, now time.Time) (authn.UserRecord, string) {
	t.Helper()
	passwordHash, err := authn.HashPassword(password)
	if err != nil {
		t.Fatalf("hash active totp password: %v", err)
	}
	secretBytes := []byte("01234567890123456789")
	secretBase32 := authn.EncodeSecretBase32(secretBytes)
	ciphertext, nonce, err := authn.EncryptSecret(keys, secretBytes)
	if err != nil {
		t.Fatalf("encrypt active totp secret: %v", err)
	}
	enrolledAt := now.Add(-time.Hour)
	return authn.UserRecord{
		ID:                   userID,
		Email:                email,
		DisplayName:          "User",
		PasswordHash:         passwordHash,
		MFARequired:          true,
		IsActive:             true,
		TOTPEnrolledAt:       &enrolledAt,
		TOTPSecretCiphertext: ciphertext,
		TOTPSecretNonce:      nonce,
	}, secretBase32
}

func generateTOTPCodeAt(t testing.TB, secretBase32 string, now time.Time) string {
	t.Helper()
	code, err := totp.GenerateCodeCustom(secretBase32, now, totp.ValidateOpts{
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

func mustJSON(t testing.TB, value any) []byte {
	t.Helper()
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("marshal json payload: %v", err)
	}
	return data
}

func callStub1[Arg any, Result any](fn func(context.Context, Arg) (Result, error), ctx context.Context, arg Arg) (Result, error) {
	if fn == nil {
		var zero Result
		return zero, errors.New("unexpected authStoreStub call")
	}
	return fn(ctx, arg)
}

func callStub2[Arg1 any, Arg2 any, Result any](fn func(context.Context, Arg1, Arg2) (Result, error), ctx context.Context, arg1 Arg1, arg2 Arg2) (Result, error) {
	if fn == nil {
		var zero Result
		return zero, errors.New("unexpected authStoreStub call")
	}
	return fn(ctx, arg1, arg2)
}

func callStub1Result2[Arg any, Result1 any, Result2 any](fn func(context.Context, Arg) (Result1, Result2, error), ctx context.Context, arg Arg) (Result1, Result2, error) {
	if fn == nil {
		var zero1 Result1
		var zero2 Result2
		return zero1, zero2, errors.New("unexpected authStoreStub call")
	}
	return fn(ctx, arg)
}

func callStub2Ptr[Arg1 any, Arg2 any, Result any](fn func(context.Context, Arg1, Arg2) (*Result, error), ctx context.Context, arg1 Arg1, arg2 Arg2) (*Result, error) {
	if fn == nil {
		return nil, errors.New("unexpected authStoreStub call")
	}
	return fn(ctx, arg1, arg2)
}

func callStub3[Arg1 any, Arg2 any, Arg3 any, Result any](fn func(context.Context, Arg1, Arg2, Arg3) (Result, error), ctx context.Context, arg1 Arg1, arg2 Arg2, arg3 Arg3) (Result, error) {
	if fn == nil {
		var zero Result
		return zero, errors.New("unexpected authStoreStub call")
	}
	return fn(ctx, arg1, arg2, arg3)
}

func callStub4Result2[Arg1 any, Arg2 any, Arg3 any, Arg4 any, Result1 any, Result2 any](fn func(context.Context, Arg1, Arg2, Arg3, Arg4) (Result1, Result2, error), ctx context.Context, arg1 Arg1, arg2 Arg2, arg3 Arg3, arg4 Arg4) (Result1, Result2, error) {
	if fn == nil {
		var zero1 Result1
		var zero2 Result2
		return zero1, zero2, errors.New("unexpected authStoreStub call")
	}
	return fn(ctx, arg1, arg2, arg3, arg4)
}

func callStub6[Arg1 any, Arg2 any, Arg3 any, Arg4 any, Arg5 any, Arg6 any, Result any](fn func(context.Context, Arg1, Arg2, Arg3, Arg4, Arg5, Arg6) (Result, error), ctx context.Context, arg1 Arg1, arg2 Arg2, arg3 Arg3, arg4 Arg4, arg5 Arg5, arg6 Arg6) (Result, error) {
	if fn == nil {
		var zero Result
		return zero, errors.New("unexpected authStoreStub call")
	}
	return fn(ctx, arg1, arg2, arg3, arg4, arg5, arg6)
}

func callStub7[Arg1 any, Arg2 any, Arg3 any, Arg4 any, Arg5 any, Arg6 any, Arg7 any, Result any](fn func(context.Context, Arg1, Arg2, Arg3, Arg4, Arg5, Arg6, Arg7) (Result, error), ctx context.Context, arg1 Arg1, arg2 Arg2, arg3 Arg3, arg4 Arg4, arg5 Arg5, arg6 Arg6, arg7 Arg7) (Result, error) {
	if fn == nil {
		var zero Result
		return zero, errors.New("unexpected authStoreStub call")
	}
	return fn(ctx, arg1, arg2, arg3, arg4, arg5, arg6, arg7)
}

func callStub8[Arg1 any, Arg2 any, Arg3 any, Arg4 any, Arg5 any, Arg6 any, Arg7 any, Arg8 any, Result any](fn func(context.Context, Arg1, Arg2, Arg3, Arg4, Arg5, Arg6, Arg7, Arg8) (Result, error), ctx context.Context, arg1 Arg1, arg2 Arg2, arg3 Arg3, arg4 Arg4, arg5 Arg5, arg6 Arg6, arg7 Arg7, arg8 Arg8) (Result, error) {
	if fn == nil {
		var zero Result
		return zero, errors.New("unexpected authStoreStub call")
	}
	return fn(ctx, arg1, arg2, arg3, arg4, arg5, arg6, arg7, arg8)
}

func callStub9Result2[Arg1 any, Arg2 any, Arg3 any, Arg4 any, Arg5 any, Arg6 any, Arg7 any, Arg8 any, Arg9 any, Result1 any, Result2 any](fn func(context.Context, Arg1, Arg2, Arg3, Arg4, Arg5, Arg6, Arg7, Arg8, Arg9) (Result1, Result2, error), ctx context.Context, arg1 Arg1, arg2 Arg2, arg3 Arg3, arg4 Arg4, arg5 Arg5, arg6 Arg6, arg7 Arg7, arg8 Arg8, arg9 Arg9) (Result1, Result2, error) {
	if fn == nil {
		var zero1 Result1
		var zero2 Result2
		return zero1, zero2, errors.New("unexpected authStoreStub call")
	}
	return fn(ctx, arg1, arg2, arg3, arg4, arg5, arg6, arg7, arg8, arg9)
}

func callStub10[Arg1 any, Arg2 any, Arg3 any, Arg4 any, Arg5 any, Arg6 any, Arg7 any, Arg8 any, Arg9 any, Arg10 any, Result any](fn func(context.Context, Arg1, Arg2, Arg3, Arg4, Arg5, Arg6, Arg7, Arg8, Arg9, Arg10) (Result, error), ctx context.Context, arg1 Arg1, arg2 Arg2, arg3 Arg3, arg4 Arg4, arg5 Arg5, arg6 Arg6, arg7 Arg7, arg8 Arg8, arg9 Arg9, arg10 Arg10) (Result, error) {
	if fn == nil {
		var zero Result
		return zero, errors.New("unexpected authStoreStub call")
	}
	return fn(ctx, arg1, arg2, arg3, arg4, arg5, arg6, arg7, arg8, arg9, arg10)
}

func callStub10Result2[Arg1 any, Arg2 any, Arg3 any, Arg4 any, Arg5 any, Arg6 any, Arg7 any, Arg8 any, Arg9 any, Arg10 any, Result1 any, Result2 any](fn func(context.Context, Arg1, Arg2, Arg3, Arg4, Arg5, Arg6, Arg7, Arg8, Arg9, Arg10) (Result1, Result2, error), ctx context.Context, arg1 Arg1, arg2 Arg2, arg3 Arg3, arg4 Arg4, arg5 Arg5, arg6 Arg6, arg7 Arg7, arg8 Arg8, arg9 Arg9, arg10 Arg10) (Result1, Result2, error) {
	if fn == nil {
		var zero1 Result1
		var zero2 Result2
		return zero1, zero2, errors.New("unexpected authStoreStub call")
	}
	return fn(ctx, arg1, arg2, arg3, arg4, arg5, arg6, arg7, arg8, arg9, arg10)
}

func callStubNoResult3[Arg1 any, Arg2 any, Arg3 any](fn func(context.Context, Arg1, Arg2, Arg3) error, ctx context.Context, arg1 Arg1, arg2 Arg2, arg3 Arg3) error {
	if fn == nil {
		return errors.New("unexpected authStoreStub call")
	}
	return fn(ctx, arg1, arg2, arg3)
}
