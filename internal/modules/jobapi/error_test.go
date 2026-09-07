package jobapi

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/jackc/pgx/v5"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
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

type failingJobRouteDB struct{ postgres.DB }
type failingJobRouteRow struct{}

func (failingJobRouteDB) QueryRow(context.Context, string, ...any) pgx.Row {
	return failingJobRouteRow{}
}
func (failingJobRouteRow) Scan(...any) error {
	return errors.New("route-secret SELECT private_relation /private/database")
}

func TestJobHandlersConcealInternalDatabaseFailure(t *testing.T) {
	service := &Service{authStore: authn.NewStore(failingJobRouteDB{}), now: time.Now}
	for _, method := range []string{http.MethodGet, http.MethodPost} {
		t.Run(method, func(t *testing.T) {
			path := "/api/v1/jobs/00000000-0000-4000-8000-000000000001"
			if method == http.MethodPost {
				path += "/cancel"
			}
			request := httptest.NewRequest(method, path, nil)
			request.Header.Set("Authorization", "Bearer route-test-token")
			recorder := httptest.NewRecorder()
			service.handleJobsMember(recorder, request)
			if recorder.Code != http.StatusInternalServerError {
				t.Fatalf("status=%d", recorder.Code)
			}
			for _, marker := range []string{"route-secret", "SELECT private_relation", "/private/database"} {
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
