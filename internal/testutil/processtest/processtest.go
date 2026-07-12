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

	"github.com/JochiRaider/cartulary/internal/testutil/diagnosticstest"
	"github.com/JochiRaider/cartulary/internal/testutil/wstest"
)

const httpAddrEnv = "CARTULARY_HTTP_ADDR"
const httpListenFDEnv = "CARTULARY_HTTP_LISTEN_FD"
const serverHarnessBinEnv = "CARTULARY_SERVER_HARNESS_BIN"

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
	listenAddr := strings.TrimSpace(env[httpAddrEnv])
	if listenAddr == "" {
		listenAddr = "127.0.0.1:0"
	}
	listener, err := net.Listen("tcp", listenAddr)
	if err != nil {
		t.Fatalf("open cmd/server listener: %v", err)
	}
	listenerFile := listenerFile(t, listener)
	env[httpAddrEnv] = listener.Addr().String()
	env[httpListenFDEnv] = "3"

	ctx, cancel := context.WithCancel(context.Background())
	command, args := serverCommand(t)
	cmd := exec.CommandContext(ctx, command, args...)
	cmd.Dir = repoRoot()
	cmd.Env = append(os.Environ(), envPairs(env)...)
	cmd.ExtraFiles = []*os.File{listenerFile}
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	address := clientAddress(listener.Addr())
	server := &Server{
		Address: address,
		BaseURL: "http://" + address,
		cancel:  cancel,
		done:    make(chan error, 1),
		cmd:     cmd,
	}
	cmd.Stdout = &server.stdout
	cmd.Stderr = &server.stderr

	if err := cmd.Start(); err != nil {
		_ = listenerFile.Close()
		_ = listener.Close()
		cancel()
		t.Fatalf("start cmd/server: %v", err)
	}
	_ = listenerFile.Close()
	_ = listener.Close()

	go func() {
		server.done <- cmd.Wait()
	}()

	return server
}

func repoRoot() string {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		return "."
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", "..", ".."))
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

func (s *Server) DiagnosticsJSON(t testing.TB) string {
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

	return stderr
}

func (s *Server) Diagnostics(t testing.TB) map[string]any {
	t.Helper()

	stderr := s.DiagnosticsJSON(t)

	var payload map[string]any
	if err := json.Unmarshal([]byte(stderr), &payload); err != nil {
		t.Fatalf("decode stderr diagnostics JSON: %v\nstderr:\n%s", err, stderr)
	}
	return payload
}

func (s *Server) RequireDiagnosticsMatchGolden(t testing.TB, goldenParts []string) {
	t.Helper()

	diagnosticstest.RequireJSONMatchesGolden(t, s.DiagnosticsJSON(t), goldenParts)
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

func serverCommand(t testing.TB) (string, []string) {
	t.Helper()
	configured := strings.TrimSpace(os.Getenv(serverHarnessBinEnv))
	if configured == "" {
		t.Fatalf("%s is required; run the owning public Make target", serverHarnessBinEnv)
	}
	info, err := os.Lstat(configured)
	if err != nil {
		t.Fatalf("%s is not usable: %v", serverHarnessBinEnv, err)
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() || info.Mode()&0o111 == 0 {
		t.Fatalf("%s must name a regular executable file: %s", serverHarnessBinEnv, configured)
	}
	return configured, nil
}

func listenerFile(t testing.TB, listener net.Listener) *os.File {
	t.Helper()

	tcpListener, ok := listener.(*net.TCPListener)
	if !ok {
		t.Fatalf("expected TCP listener, got %T", listener)
	}
	file, err := tcpListener.File()
	if err != nil {
		t.Fatalf("duplicate cmd/server listener fd: %v", err)
	}
	return file
}

func clientAddress(addr net.Addr) string {
	host, port, err := net.SplitHostPort(addr.String())
	if err != nil {
		return addr.String()
	}
	switch host {
	case "", "0.0.0.0", "::":
		host = "127.0.0.1"
	}
	return net.JoinHostPort(host, port)
}
