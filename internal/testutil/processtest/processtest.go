package processtest

import (
	"bytes"
	"context"
	"encoding/json"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/testutil/wstest"
)

const httpAddrEnv = "CARTULARY_HTTP_ADDR"

type Server struct {
	Address string
	BaseURL string

	cancel context.CancelFunc
	done   chan error
	cmd    *exec.Cmd
	stderr bytes.Buffer
	stdout bytes.Buffer
}

type ServerOptions struct {
	Env map[string]string
}

func StartServer(t testing.TB, options ServerOptions) *Server {
	t.Helper()

	env := make(map[string]string, len(options.Env)+1)
	for key, value := range options.Env {
		env[key] = value
	}
	if strings.TrimSpace(env[httpAddrEnv]) == "" {
		env[httpAddrEnv] = freeAddress(t)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cmd := exec.CommandContext(ctx, "go", "run", "./cmd/server")
	cmd.Dir = repoRoot()
	cmd.Env = append(os.Environ(), envPairs(env)...)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	server := &Server{
		Address: env[httpAddrEnv],
		BaseURL: "http://" + env[httpAddrEnv],
		cancel:  cancel,
		done:    make(chan error, 1),
		cmd:     cmd,
	}
	cmd.Stdout = &server.stdout
	cmd.Stderr = &server.stderr

	if err := cmd.Start(); err != nil {
		cancel()
		t.Fatalf("start cmd/server: %v", err)
	}

	go func() {
		server.done <- cmd.Wait()
	}()

	return server
}

func (s *Server) WaitForReady(t testing.TB) {
	t.Helper()

	client := &http.Client{Timeout: 500 * time.Millisecond}
	deadline := time.Now().Add(15 * time.Second)
	for time.Now().Before(deadline) {
		select {
		case err := <-s.done:
			t.Fatalf("cmd/server exited before readiness: %v\nstdout:\n%s\nstderr:\n%s", err, s.stdout.String(), s.stderr.String())
		default:
		}

		healthResp, healthErr := client.Get(s.BaseURL + "/healthz")
		if healthErr == nil {
			healthResp.Body.Close()
			readyResp, readyErr := client.Get(s.BaseURL + "/readyz")
			if readyErr == nil {
				readyResp.Body.Close()
				if healthResp.StatusCode == http.StatusOK && readyResp.StatusCode == http.StatusOK {
					return
				}
			}
		}

		time.Sleep(200 * time.Millisecond)
	}

	t.Fatalf("timed out waiting for readiness\nstdout:\n%s\nstderr:\n%s", s.stdout.String(), s.stderr.String())
}

func (s *Server) WaitForExit(t testing.TB) error {
	t.Helper()

	select {
	case err := <-s.done:
		return err
	case <-time.After(15 * time.Second):
		t.Fatalf("timed out waiting for cmd/server exit\nstdout:\n%s\nstderr:\n%s", s.stdout.String(), s.stderr.String())
		return nil
	}
}

func (s *Server) Stop(t testing.TB) {
	t.Helper()

	if s == nil {
		return
	}
	s.cancel()
	if s.cmd != nil && s.cmd.Process != nil {
		_ = syscall.Kill(-s.cmd.Process.Pid, syscall.SIGTERM)
	}
	select {
	case <-s.done:
	case <-time.After(5 * time.Second):
		t.Fatalf("timed out stopping cmd/server\nstdout:\n%s\nstderr:\n%s", s.stdout.String(), s.stderr.String())
	}
}

func (s *Server) RequireStatus(t testing.TB, path string, want int) {
	t.Helper()

	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(s.BaseURL + path)
	if err != nil {
		t.Fatalf("request %s: %v", path, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != want {
		t.Fatalf("unexpected status for %s: got %d want %d", path, resp.StatusCode, want)
	}
}

func (s *Server) RequireConnectionRefused(t testing.TB, path string) {
	t.Helper()

	client := &http.Client{Timeout: 500 * time.Millisecond}
	if resp, err := client.Get(s.BaseURL + path); err == nil {
		resp.Body.Close()
		t.Fatalf("expected %s to be unreachable, got HTTP %d", path, resp.StatusCode)
	}
}

func (s *Server) RequireWebsocketConnectionRefused(t testing.TB, path string) {
	t.Helper()

	_, _, err := wstest.TryConnect(s.BaseURL, path, nil)
	wstest.RequireConnectionRefused(t, err)
}

func (s *Server) Diagnostics(t testing.TB) map[string]any {
	t.Helper()

	stderr := strings.TrimSpace(s.stderr.String())
	if stderr == "" {
		t.Fatal("expected structured startup diagnostics on stderr")
	}
	if idx := strings.LastIndex(stderr, "\nexit status "); idx >= 0 {
		stderr = strings.TrimSpace(stderr[:idx])
	}
	if strings.HasPrefix(stderr, "exit status ") {
		t.Fatalf("missing structured startup diagnostics on stderr\nstderr:\n%s", s.stderr.String())
	}

	var payload map[string]any
	if err := json.Unmarshal([]byte(stderr), &payload); err != nil {
		t.Fatalf("decode stderr diagnostics JSON: %v\nstderr:\n%s", err, stderr)
	}
	return payload
}

func (s *Server) RequireDiagnosticsCode(t testing.TB, wantCode string) {
	t.Helper()

	payload := s.Diagnostics(t)
	errorPayload := payload["error"].(map[string]any)
	if errorPayload["code"] != wantCode {
		t.Fatalf("unexpected diagnostics code: got %v want %s", errorPayload["code"], wantCode)
	}
}

func (s *Server) RequireReasonCode(t testing.TB, wantReasonCode string) {
	t.Helper()

	payload := s.Diagnostics(t)
	items := payload["error"].(map[string]any)["details"].(map[string]any)["items"].([]any)
	for _, item := range items {
		typed := item.(map[string]any)
		if typed["reason_code"] == wantReasonCode {
			return
		}
	}
	t.Fatalf("missing reason_code=%q in stderr payload %s", wantReasonCode, s.stderr.String())
}

func (s *Server) RequireDiagnosticsField(t testing.TB, wantPath string, wantReasonCode string) {
	t.Helper()

	payload := s.Diagnostics(t)
	items := payload["error"].(map[string]any)["details"].(map[string]any)["items"].([]any)
	for _, item := range items {
		typed := item.(map[string]any)
		if typed["path"] == wantPath && typed["reason_code"] == wantReasonCode {
			return
		}
	}
	t.Fatalf("missing diagnostic path=%q reason_code=%q in stderr payload %s", wantPath, wantReasonCode, s.stderr.String())
}

func envPairs(env map[string]string) []string {
	pairs := make([]string, 0, len(env))
	for key, value := range env {
		pairs = append(pairs, key+"="+value)
	}
	return pairs
}

func freeAddress(t testing.TB) string {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("reserve free tcp port: %v", err)
	}
	defer listener.Close()
	return listener.Addr().String()
}

func repoRoot() string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(file), "..", "..", "..")
}
