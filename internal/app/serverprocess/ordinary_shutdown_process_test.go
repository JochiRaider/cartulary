package serverprocess

import (
	"errors"
	"io"
	"net/http"
	"os/exec"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/app/server/holdjob"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/JochiRaider/cartulary/internal/testutil/processtest"
	"github.com/JochiRaider/cartulary/internal/testutil/wstest"
)

// ac604DrainSeconds is the configured shutdown deadline; the HTTP drain ends
// min(25%, 5 s) earlier, at 1.5 s (REQ-04-146a).
const ac604DrainSeconds = 2

// TestOrdinaryShutdownAC604_Process covers Core 04 AC-604 on the real
// server-harness binary. The server holds an open WebSocket, a request whose
// body never completes, and an active job when it receives SIGTERM, then a
// second SIGTERM. The socket closes with 1001; the request is cancelled at the
// close reserve boundary; a job that honours cancellation lets the process
// exit 0 within the deadline, and one that ignores it is named as open with
// exit 1. A repeated stop never extends the deadline, and neither case is
// component loss.
func TestOrdinaryShutdownAC604_Process(t *testing.T) {
	for _, tc := range []struct {
		name     string
		mode     string
		wantExit int
	}{
		{name: "job honours cancellation", mode: holdjob.ModeRespectCancellation, wantExit: 0},
		{name: "job ignores cancellation", mode: holdjob.ModeIgnoreCancellation, wantExit: 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			publicOrigin := "https://127.0.0.1:4173"
			server := startHarnessRuntimeServerProcess(t, "ac604-"+strings.ReplaceAll(tc.name, " ", "-"), map[string]string{
				"CARTULARY_ENABLE_TEST_ROUTES":                            "1",
				"CARTULARY_TEST_RUNTIME_MARKER":                           "harness-owned",
				"CARTULARY_TEST_ROUTE_TOKEN":                              httptestx.TestRouteToken,
				"CARTULARY_WEB_E2E_PUBLIC_ORIGIN":                         publicOrigin,
				"CARTULARY__TIMEOUTS__EXTENSIONS__SHUTDOWN_DRAIN_SECONDS": "2",
			}, func(env map[string]string, baseURL string) {
				env["CARTULARY_WEB_E2E_API_ORIGIN"] = baseURL
			})
			login, _ := provisionBootstrapAdmin(t, server)
			auth := []func(*http.Request){httptestx.WithCookies(login.Session, login.CSRF), httptestx.WithHeader(authn.CSRFHeaderName, login.CSRF.Value)}
			session := httptestx.RequireSuccessEnvelope(t, doJSON(t, server, http.MethodGet, "/api/v1/auth/session", nil, auth...), http.StatusOK)["data"].(map[string]any)

			// An open WebSocket.
			socket := connectSessionSocket(t, server, login, "ac604-"+tc.mode)

			// An active job.
			enqueued := holdJobControl(t, server, publicOrigin, map[string]any{"operation": "enqueue", "mode": tc.mode, "submitted_by_user_id": session["user_id"]})
			if enqueued["job_id"] == "" {
				t.Fatalf("hold job not enqueued: %#v", enqueued)
			}
			waitUntil(t, 10*time.Second, func() bool {
				return holdJobControl(t, server, publicOrigin, map[string]any{"operation": "status"})["running"] == float64(1)
			})

			// An in-flight request whose body never completes.
			body, holdBody := io.Pipe()
			defer holdBody.Close()
			request, err := http.NewRequest(http.MethodPost, server.BaseURL+"/api/v1/incidents", body)
			if err != nil {
				t.Fatal(err)
			}
			request.Header.Set("Content-Type", "application/json")
			// The body follows only the server's 100 Continue, which it sends
			// when the handler first reads the body. Once the first body
			// write returns, the request is in the handler. Without this the
			// stop could arrive while the server had not yet read the request,
			// and the drain, finding nothing in flight, would end at once.
			request.Header.Set("Expect", "100-continue")
			for _, option := range auth {
				option(request)
			}
			transport := server.Client.Transport.(*http.Transport).Clone()
			transport.ExpectContinueTimeout = 30 * time.Second
			defer transport.CloseIdleConnections()
			client := &http.Client{Transport: transport, Timeout: server.Client.Timeout}
			inFlight := make(chan error, 1)
			go func() {
				resp, err := client.Do(request)
				if resp != nil {
					_ = resp.Body.Close()
				}
				inFlight <- err
			}()
			if _, err := holdBody.Write([]byte(`{"client_txn_id":"ac604-in-flight",`)); err != nil {
				t.Fatalf("start in-flight body: %v", err)
			}

			stopped := time.Now()
			server.Signal(t, syscall.SIGTERM)
			time.Sleep(300 * time.Millisecond)
			server.Signal(t, syscall.SIGTERM) // a repeated stop must not extend the deadline

			wstest.RequireClose(t, socket.AwaitClose(5*time.Second), wstest.StatusGoingAway, "")
			exitErr := server.WaitForExit(t)
			elapsed := time.Since(stopped)
			if limit := ac604DrainSeconds*time.Second + 1500*time.Millisecond; elapsed > limit {
				t.Fatalf("shutdown took %s, beyond the %d s deadline (limit %s)", elapsed, ac604DrainSeconds, limit)
			}
			exitCode := 0
			var processExit *exec.ExitError
			if errors.As(exitErr, &processExit) {
				exitCode = processExit.ExitCode()
			} else if exitErr != nil {
				t.Fatalf("server wait: %v", exitErr)
			}
			if exitCode != tc.wantExit {
				t.Fatalf("exit=%d want %d\nstdout:\n%s\nstderr:\n%s", exitCode, tc.wantExit, server.Stdout(), server.Stderr())
			}
			// The server cancelled the request at the close reserve boundary (the
			// drain log below records it); release the client, which is still
			// blocked reading the never-finished body.
			_ = holdBody.Close()
			select {
			case <-inFlight:
			case <-time.After(5 * time.Second):
				t.Fatal("the in-flight request did not end after the process exited")
			}
			stdout, stderr := server.Stdout(), server.Stderr()
			for _, phase := range []string{"phase=stop_requested", "phase=http_drain", "cancelled_in_flight=true"} {
				if !strings.Contains(stdout, phase) {
					t.Fatalf("shutdown log lacks %q:\n%s", phase, stdout)
				}
			}
			if strings.Contains(stderr, "published_component_lost") || strings.Contains(stdout, "published_component_lost") {
				t.Fatalf("ordinary shutdown reported component loss:\n%s\n%s", stdout, stderr)
			}
			if tc.wantExit == 1 {
				diagnostic := server.DiagnosticsJSON(t)
				if !strings.Contains(diagnostic, `"code":"shutdown_incomplete"`) || !strings.Contains(diagnostic, `"jobs_runner"`) {
					t.Fatalf("diagnostic does not name the open jobs_runner: %s", diagnostic)
				}
			}
		})
	}
}

func holdJobControl(t testing.TB, server *processtest.Server, origin string, body map[string]any) map[string]any {
	t.Helper()
	body["schema_id"] = "cartulary.test.hold_job_control.v1"
	resp := doHarnessRuntimeJSON(t, server, http.MethodPost, "/api/v1/test/runtime/hold-jobs", body, httptestx.TestRouteToken, origin, "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("hold-job control %v: status %d", body["operation"], resp.StatusCode)
	}
	return httptestx.ReadJSONBody(t, resp)
}

func waitUntil(t testing.TB, timeout time.Duration, condition func() bool) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for !condition() {
		if time.Now().After(deadline) {
			t.Fatalf("condition not met within %s", timeout)
		}
		time.Sleep(50 * time.Millisecond)
	}
}
