package jobapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/httpauth"
	"github.com/JochiRaider/cartulary/internal/platform/jobs"
	"github.com/JochiRaider/cartulary/internal/platform/postgres/dbport"
)

func TestJobInternalErrorDoesNotRetainPrivateCause(t *testing.T) {
	const private = "job-private-marker SELECT private_table /private/jobs"
	apiErr := internalAPIError(errors.New(private))
	if strings.Contains(apiErr.Message, private) || apiErr.Message != "internal_error" {
		t.Fatal("job API error retains private cause")
	}
	recorder := httptest.NewRecorder()
	writeAPIError(recorder, httptest.NewRequest(http.MethodGet, "/api/v1/jobs/example", nil), apiErr)
	if recorder.Code != 500 || strings.Contains(recorder.Body.String(), private) {
		t.Fatal("job API failure response discloses private cause")
	}
}

type failingJobRouteDB struct{ dbport.DB }
type failingJobRouteRow struct{}

func (failingJobRouteDB) QueryRow(context.Context, string, ...any) pgx.Row {
	return failingJobRouteRow{}
}
func (failingJobRouteRow) Scan(...any) error {
	return errors.New("route-secret SELECT private_relation /private/database")
}

// activeSessionStore authenticates every token as one active session, so a
// request reaches the job handler.
type activeSessionStore struct{}

func (activeSessionStore) GetSessionByFingerprint(context.Context, []byte) (authn.SessionRecord, authn.UserRecord, error) {
	userID := uuid.MustParse("00000000-0000-4000-8000-0000000000a1")
	return authn.SessionRecord{ID: uuid.MustParse("00000000-0000-4000-8000-0000000000b1"), UserID: userID, SessionExpiresAt: time.Now().Add(time.Hour)},
		authn.UserRecord{ID: userID, IsActive: true}, nil
}

func (activeSessionStore) GetBootstrapTokenByFingerprint(context.Context, []byte) (authn.BootstrapTokenRecord, authn.UserRecord, error) {
	return authn.BootstrapTokenRecord{}, authn.UserRecord{}, authn.ErrNotFound
}

func (activeSessionStore) RevokeSession(context.Context, uuid.UUID, string, time.Time) error {
	return nil
}

// failingJobs fails every job read and cancel with a private cause.
type failingJobs struct{}

func (failingJobs) Get(context.Context, uuid.UUID) (jobs.Resource, error) {
	return jobs.Resource{}, errors.New("job-secret SELECT private_jobs /private/jobs")
}

func (failingJobs) Cancel(context.Context, jobs.CancelParams) (jobs.CancelResult, error) {
	return jobs.CancelResult{}, errors.New("job-secret SELECT private_jobs /private/jobs")
}

// TestJobHandlersConcealInternalDatabaseFailure routes each job operation
// through its registry binding, so central admission authenticates first
// (Core 01 REQ-01-707). A failing session store and a failing job read each
// yield the canonical internal_error envelope without the private cause.
func TestJobHandlersConcealInternalDatabaseFailure(t *testing.T) {
	for _, failure := range []struct {
		name    string
		store   httpauth.SessionStore
		markers []string
	}{
		{"session store", authn.NewStore(failingJobRouteDB{}), []string{"route-secret", "SELECT private_relation", "/private/database"}},
		{"job read", activeSessionStore{}, []string{"job-secret", "SELECT private_jobs", "/private/jobs"}},
	} {
		registry, err := httpapi.NewRouteRegistry(nil)
		if err != nil {
			t.Fatal(err)
		}
		deps := httpapi.DependencySet{
			PublicRoutes:  registry,
			PostgresDB:    failingJobRouteDB{},
			Authenticator: httpauth.NewAuthenticator(httpauth.Options{Store: failure.store, Now: time.Now}),
		}
		mux := http.NewServeMux()
		if err := RegisterRoutes(failingJobs{})(mux, deps); err != nil {
			t.Fatal(err)
		}
		handler := registry.Handler(mux)
		for _, method := range []string{http.MethodGet, http.MethodPost} {
			t.Run(failure.name+" "+method, func(t *testing.T) {
				path := "/api/v1/jobs/00000000-0000-4000-8000-000000000001"
				body := ""
				if method == http.MethodPost {
					path += "/cancel"
					body = `{"client_txn_id":"txn-conceal"}`
				}
				request := httptest.NewRequest(method, path, strings.NewReader(body))
				request.Header.Set("Authorization", "Bearer route-test-token")
				recorder := httptest.NewRecorder()
				handler.ServeHTTP(recorder, request)
				if recorder.Code != http.StatusInternalServerError {
					t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body)
				}
				for _, marker := range failure.markers {
					if strings.Contains(recorder.Body.String(), marker) {
						t.Fatal("handler response discloses private marker")
					}
				}
				var envelope struct {
					Error struct {
						Code      string
						Message   string
						Status    int
						Retryable bool
						Details   map[string]any
						Conflict  any
					}
				}
				if err := json.Unmarshal(recorder.Body.Bytes(), &envelope); err != nil {
					t.Fatal(err)
				}
				value := envelope.Error
				if value.Code != "internal_error" || value.Message != "internal_error" || value.Status != 500 || value.Retryable || value.Details == nil || len(value.Details) != 0 || value.Conflict != nil {
					t.Fatal("handler response is not canonical")
				}
			})
		}
	}
}
