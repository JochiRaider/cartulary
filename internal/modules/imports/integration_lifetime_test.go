package imports_test

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/authcookietest"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
)

// startImportJSON owns the worker and any unconsumed response through test
// cleanup, including failure before a concurrent-start gate is released.
func startImportJSON(t testing.TB, base string, login authcookietest.AuthCookies, path string, body any, start <-chan struct{}) func() *http.Response {
	t.Helper()
	request := httptestx.NewJSONRequest(t, http.MethodPost, base+path, body)
	ctx, cancel := context.WithCancel(t.Context())
	request = request.WithContext(ctx)
	request.Header.Set(authn.CSRFHeaderName, login.CSRF.Value)
	request.AddCookie(login.Session)
	request.AddCookie(login.CSRF)
	done := make(chan struct{})
	var response *http.Response
	var requestErr error
	t.Cleanup(func() {
		cancel()
		select {
		case <-done:
			if response != nil {
				if err := response.Body.Close(); err != nil {
					t.Error(err)
				}
			}
		case <-time.After(5 * time.Second):
			t.Error("concurrent import HTTP worker did not settle during cleanup")
		}
	})
	go func() {
		defer close(done)
		if start != nil {
			select {
			case <-start:
			case <-ctx.Done():
				requestErr = ctx.Err()
				return
			}
		}
		response, requestErr = http.DefaultClient.Do(request)
	}()
	return func() *http.Response {
		t.Helper()
		select {
		case <-done:
			if requestErr != nil {
				t.Fatalf("concurrent import HTTP request failed: %v", requestErr)
			}
			return response
		case <-time.After(10 * time.Second):
			t.Fatal("concurrent import HTTP response exceeded observation bound")
			return nil
		}
	}
}

// closeImportAdvisoryGate releases only a dedicated test blocker session. Register
// it immediately after acquisition so failed setup cannot return a locked session.
func closeImportAdvisoryGate(t testing.TB, conn *sql.Conn) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if _, err := conn.ExecContext(ctx, "SELECT pg_advisory_unlock_all()"); err != nil {
		t.Errorf("release import advisory gate: %v", err)
		_ = conn.Raw(func(any) error { return driver.ErrBadConn })
	}
	if err := conn.Close(); err != nil && !errors.Is(err, sql.ErrConnDone) {
		t.Errorf("close import advisory gate connection: %v", err)
	}
}
