package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/platform/authn"
)

func TestLoginCapacityAdmissionDoesNotDiscloseAccountExistence_Unit(t *testing.T) {
	first, err := authn.BeginPasswordWorkflow(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer first.Close()
	second, err := authn.BeginPasswordWorkflow(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer second.Close()
	var lookups atomic.Int32
	store := &authStoreStub{getUserByNormalizedEmailFunc: func(context.Context, string) (authn.UserRecord, error) {
		lookups.Add(1)
		return authn.UserRecord{}, authn.ErrNotFound
	}}
	service := &Service{loginStore: store}
	type result struct {
		response *httptest.ResponseRecorder
		elapsed  time.Duration
	}
	results := make(chan result, 64)
	start := make(chan struct{})
	for i := range 64 {
		go func() {
			<-start
			username := "existing@example.test"
			if i%2 == 0 {
				username = "unknown@example.test"
			}
			request := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(`{"username":"`+username+`","password":"Incorrect password 1!"}`))
			response := httptest.NewRecorder()
			at := time.Now()
			service.handleLogin(response, request)
			results <- result{response: response, elapsed: time.Since(at)}
		}()
	}
	close(start)
	for range 64 - authn.PasswordPendingWorkflows {
		select {
		case outcome := <-results:
			if outcome.response.Code != http.StatusServiceUnavailable || outcome.response.Header().Get("Retry-After") != "1" {
				t.Fatalf("overload response %d %s", outcome.response.Code, outcome.response.Body.String())
			}
			if outcome.elapsed > 250*time.Millisecond {
				t.Fatalf("excess request queued for %s", outcome.elapsed)
			}
			var envelope struct {
				Error struct {
					Code      string         `json:"code"`
					Status    int            `json:"status"`
					Retryable bool           `json:"retryable"`
					Details   map[string]any `json:"details"`
				} `json:"error"`
			}
			if err := json.Unmarshal(outcome.response.Body.Bytes(), &envelope); err != nil || envelope.Error.Code != "authentication_capacity_exhausted" || envelope.Error.Details == nil || len(envelope.Error.Details) != 0 || envelope.Error.Status != 503 || !envelope.Error.Retryable {
				t.Fatalf("unsafe overload envelope %v", err)
			}
		case <-time.After(time.Second):
			t.Fatal("overflow did not reject promptly")
		}
	}
	if lookups.Load() != 0 {
		t.Fatal("account lookup preceded admission")
	}
	first.Close()
	second.Close()
	for range authn.PasswordPendingWorkflows {
		select {
		case outcome := <-results:
			if outcome.response.Code != http.StatusUnauthorized {
				t.Fatalf("admitted request %d", outcome.response.Code)
			}
		case <-time.After(time.Second):
			t.Fatal("admitted request did not recover")
		}
	}
	if lookups.Load() != authn.PasswordPendingWorkflows {
		t.Fatal("unexpected admitted lookup count")
	}
}
