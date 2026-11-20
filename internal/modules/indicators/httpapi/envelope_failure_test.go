package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/incidents/admission"
	"github.com/JochiRaider/cartulary/internal/modules/indicators"
	recordenvelope "github.com/JochiRaider/cartulary/internal/modules/records/enveloperead"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	platformhttpapi "github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/httpauth"
)

// serveAdmittedIndicatorRoutes binds every Indicator operation of s for
// central admission behind the registry's route resolution, as production
// does (Core 01 REQ-01-707). The authenticator reads sessions, so a request
// reaches its handler only after authentication and path-parameter admission.
func serveAdmittedIndicatorRoutes(t testing.TB, s *service, sessions httpauth.SessionStore, now time.Time) http.Handler {
	t.Helper()
	registry, err := platformhttpapi.NewRouteRegistry(nil)
	if err != nil {
		t.Fatalf("create route registry: %v", err)
	}
	deps := platformhttpapi.DependencySet{
		PublicRoutes:  registry,
		Authenticator: httpauth.NewAuthenticator(httpauth.Options{Store: sessions, Now: func() time.Time { return now }}),
	}
	mux := http.NewServeMux()
	if err := platformhttpapi.BindAdmittedOwnerRoutes(mux, deps, "module.indicators", s.operationHandlers()); err != nil {
		t.Fatalf("bind Indicator routes: %v", err)
	}
	for operationID, mode := range registry.AdmissionModes() {
		if mode != platformhttpapi.AdmissionCentral {
			t.Fatalf("Indicator operation %q admission mode = %q, want central", operationID, mode)
		}
	}
	return registry.Handler(mux)
}

type failingEnvelopeReader struct{ err error }

func (r failingEnvelopeReader) LoadEnvelope(context.Context, uuid.UUID) (recordenvelope.Envelope, error) {
	return recordenvelope.Envelope{}, r.err
}

// failingObservationOwner answers the observation lookup of an observation
// action with err. Any other owner call is outside the lookup under test.
type failingObservationOwner struct {
	ownerApplication
	err error
}

func (o failingObservationOwner) GetIndicatorObservation(context.Context, uuid.UUID) (indicators.IndicatorObservationRecord, error) {
	return indicators.IndicatorObservationRecord{}, o.err
}

type unreachedIncidentAdmission struct{ t *testing.T }

func (a unreachedIncidentAdmission) Check(context.Context, uuid.UUID, uuid.UUID, admission.Requirement) (admission.Grant, error) {
	a.t.Helper()
	a.t.Error("incident admission ran after an unresolved record envelope")
	return admission.Grant{}, errors.New("unreached")
}

type activeBearerSessions struct{ now time.Time }

func (s activeBearerSessions) GetSessionByFingerprint(context.Context, []byte) (authn.SessionRecord, authn.UserRecord, error) {
	user := authn.UserRecord{ID: uuid.MustParse("0c1d2e3f-4a5b-4c6d-8e7f-9a0b1c2d3e01"), IsActive: true}
	return authn.SessionRecord{
		ID: uuid.MustParse("0c1d2e3f-4a5b-4c6d-8e7f-9a0b1c2d3e02"), UserID: user.ID,
		AuthenticatedAt: s.now, LastQualifyingActivityAt: s.now,
		IdleExpiresAt: s.now.Add(time.Hour), AbsoluteExpiresAt: s.now.Add(time.Hour), SessionExpiresAt: s.now.Add(time.Hour),
	}, user, nil
}

func (activeBearerSessions) GetBootstrapTokenByFingerprint(context.Context, []byte) (authn.BootstrapTokenRecord, authn.UserRecord, error) {
	return authn.BootstrapTokenRecord{}, authn.UserRecord{}, authn.ErrNotFound
}

func (activeBearerSessions) RevokeSession(context.Context, uuid.UUID, string, time.Time) error {
	return nil
}

func (activeBearerSessions) SlideSession(_ context.Context, _ uuid.UUID, timing authn.SessionTiming) (authn.SessionTiming, error) {
	return timing, nil
}

// TestIndicatorEnvelopeReadFailuresAreInternal_Unit drives every Indicator
// route through the lookup of its path target: the record envelope, or the
// observation an observation action addresses. Each request goes through the
// registry binding, so central admission authenticates and admits the path
// identifier first (REQ-01-707). Only a missing target is the route's
// not-found response; an injected lookup failure is 500 internal_error on
// reads and mutations alike, before any incident admission (REQ-01-706).
func TestIndicatorEnvelopeReadFailuresAreInternal_Unit(t *testing.T) {
	now := time.Date(2026, 11, 18, 9, 30, 0, 0, time.UTC)
	recordID := uuid.MustParse("0c1d2e3f-4a5b-4c6d-8e7f-9a0b1c2d3e03")
	type lookupCase struct {
		name   string
		err    error
		status int
		code   func(notFound string) string
	}
	id := recordID.String()
	routes := []struct {
		name, method, path, notFound string
		observation                  bool
	}{
		{"listSourceRecordIndicatorObservations", http.MethodGet, "/api/v1/records/" + id + "/indicator-observations", "indicator_source_record_not_found", false},
		{"listIndicatorObservations", http.MethodGet, "/api/v1/indicators/" + id + "/observations", "indicator_not_found", false},
		{"listIndicatorStateIntervals", http.MethodGet, "/api/v1/indicators/" + id + "/state-intervals", "indicator_not_found", false},
		{"createManualIndicatorObservation", http.MethodPost, "/api/v1/records/" + id + "/indicator-observations", "indicator_source_record_not_found", false},
		{"appendIndicatorStateInterval", http.MethodPost, "/api/v1/indicators/" + id + "/state-intervals", "indicator_not_found", false},
		{"resolveIndicatorObservation", http.MethodPost, "/api/v1/indicator-observations/" + id + "/resolve", "indicator_observation_not_found", true},
		{"dismissIndicatorObservation", http.MethodPost, "/api/v1/indicator-observations/" + id + "/dismiss", "indicator_observation_not_found", true},
		{"restoreIndicatorObservation", http.MethodPost, "/api/v1/indicator-observations/" + id + "/restore", "indicator_observation_not_found", true},
	}
	injected := errors.New("injected record envelope database failure")
	envelopeCases := []lookupCase{
		{"missing envelope", fmt.Errorf("load envelope: %w", recordenvelope.ErrEnvelopeNotFound), http.StatusNotFound, func(notFound string) string { return notFound }},
		{"envelope read failure", fmt.Errorf("load envelope: %w", injected), http.StatusInternalServerError, func(string) string { return "internal_error" }},
	}
	observationCases := []lookupCase{
		{"missing observation", fmt.Errorf("load observation: %w", indicators.ErrIndicatorObservationNotFound), http.StatusNotFound, func(notFound string) string { return notFound }},
		{"observation read failure", fmt.Errorf("load indicator observation: %w", errors.New("injected observation database failure")), http.StatusInternalServerError, func(string) string { return "internal_error" }},
	}
	for _, route := range routes {
		cases := envelopeCases
		if route.observation {
			cases = observationCases
		}
		for _, tc := range cases {
			t.Run(route.name+"/"+tc.name, func(t *testing.T) {
				sessions := activeBearerSessions{now: now}
				s := &service{
					records:   failingEnvelopeReader{err: tc.err},
					incidents: unreachedIncidentAdmission{t: t},
					sessions:  sessions,
					now:       func() time.Time { return now },
				}
				if route.observation {
					s.records = nil
					s.owner = failingObservationOwner{err: tc.err}
				}
				request := httptest.NewRequest(route.method, route.path, strings.NewReader(`{}`))
				request.Header.Set("Authorization", "Bearer indicator-envelope-failure")
				request.Header.Set("Content-Type", "application/json")
				recorder := httptest.NewRecorder()
				serveAdmittedIndicatorRoutes(t, s, sessions, now).ServeHTTP(recorder, request)
				var envelope struct {
					Error struct {
						Code    string `json:"code"`
						Message string `json:"message"`
					} `json:"error"`
				}
				if err := json.Unmarshal(recorder.Body.Bytes(), &envelope); err != nil {
					t.Fatalf("decode error envelope %q: %v", recorder.Body.String(), err)
				}
				if want := tc.code(route.notFound); recorder.Code != tc.status || envelope.Error.Code != want {
					t.Fatalf("response = %d %s, want %d %s; body=%s", recorder.Code, envelope.Error.Code, tc.status, want, recorder.Body.String())
				}
				if strings.Contains(recorder.Body.String(), "injected") {
					t.Fatalf("response leaks the storage diagnostic: %s", recorder.Body.String())
				}
			})
		}
	}
}
