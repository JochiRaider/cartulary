package processtest

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"sort"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/testutil/diagnosticstest"
	"github.com/JochiRaider/cartulary/internal/testutil/suiteservices"
	"github.com/JochiRaider/cartulary/internal/testutil/wstest"
)

const httpAddrEnv = "CARTULARY_HTTP_ADDR"
const httpListenFDEnv = "CARTULARY_HTTP_LISTEN_FD"
const serverHarnessBinEnv = "CARTULARY_SERVER_HARNESS_BIN"

const (
	readinessDeadline       = 15 * time.Second
	readinessRequestTimeout = 500 * time.Millisecond
	readinessPollInterval   = 200 * time.Millisecond
	exitDeadline            = 15 * time.Second
	statusRequestTimeout    = 2 * time.Second
	gracefulStopDeadline    = 4 * time.Second
	forcedStopDeadline      = 1 * time.Second
	stopDeadline            = gracefulStopDeadline + forcedStopDeadline
	refusalRequestTimeout   = 500 * time.Millisecond
	processPollInterval     = 10 * time.Millisecond
)

type synchronizedBuffer struct {
	mu     sync.Mutex
	buffer bytes.Buffer
}

func (b *synchronizedBuffer) Write(payload []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buffer.Write(payload)
}

func (b *synchronizedBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buffer.String()
}

type processGroupController struct {
	pid int
}

func (c processGroupController) signal(signal syscall.Signal) error {
	if c.pid <= 0 {
		return nil
	}
	err := syscall.Kill(-c.pid, signal)
	if errors.Is(err, syscall.ESRCH) {
		return nil
	}
	return err
}

func (c processGroupController) alive() bool {
	if c.pid <= 0 {
		return false
	}
	err := syscall.Kill(-c.pid, 0)
	return err == nil || errors.Is(err, syscall.EPERM)
}

type Server struct {
	BaseURL string

	cmd        *exec.Cmd
	group      processGroupController
	done       chan struct{}
	resultMu   sync.RWMutex
	waitErr    error
	stopOnce   sync.Once
	stopResult error
	stderr     synchronizedBuffer
	stdout     synchronizedBuffer
}

type ServerOptions struct {
	Env         map[string]string
	FinalizeEnv func(env map[string]string, baseURL string)
}

func StartServer(t testing.TB, options ServerOptions) *Server {
	t.Helper()

	requestedEnv := cloneEnv(options.Env)
	listenAddr := strings.TrimSpace(requestedEnv[httpAddrEnv])
	if listenAddr == "" {
		listenAddr = "127.0.0.1:0"
	}
	listener, err := net.Listen("tcp", listenAddr)
	if err != nil {
		t.Fatalf("open cmd/server listener: %v", err)
	}
	listenerFile := listenerFile(t, listener)
	closeListener := func() {
		_ = listenerFile.Close()
		_ = listener.Close()
	}

	address := clientAddress(listener.Addr())
	baseURL := "http://" + address
	finalEnv := cloneEnv(requestedEnv)
	if options.FinalizeEnv != nil {
		options.FinalizeEnv(finalEnv, baseURL)
	}
	finalEnv[httpAddrEnv] = listener.Addr().String()
	finalEnv[httpListenFDEnv] = "3"

	command := serverCommand(t)
	repoRoot, err := resolveRepoRoot()
	if err != nil {
		closeListener()
		t.Fatalf("resolve repository root: %v", err)
	}
	cmd := exec.Command(command)
	cmd.Dir = repoRoot
	cmd.Env = append(os.Environ(), envPairs(finalEnv)...)
	cmd.ExtraFiles = []*os.File{listenerFile}
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	server := &Server{
		BaseURL: baseURL,
		cmd:     cmd,
		done:    make(chan struct{}),
	}
	cmd.Stdout = &server.stdout
	cmd.Stderr = &server.stderr

	if err := cmd.Start(); err != nil {
		closeListener()
		t.Fatalf("start cmd/server: %v", err)
	}
	server.group = processGroupController{pid: cmd.Process.Pid}
	closeListener()

	go func() {
		waitErr := cmd.Wait()
		server.resultMu.Lock()
		server.waitErr = waitErr
		server.resultMu.Unlock()
		close(server.done)
	}()

	return server
}

func resolveRepoRoot() (string, error) {
	return suiteservices.FindRepoRoot()
}

func cloneEnv(source map[string]string) map[string]string {
	cloned := make(map[string]string, len(source)+2)
	for key, value := range source {
		cloned[key] = value
	}
	return cloned
}

func (s *Server) WaitForReady(t testing.TB) {
	t.Helper()

	client := &http.Client{Timeout: readinessRequestTimeout}
	deadline := time.Now().Add(readinessDeadline)
	for time.Now().Before(deadline) {
		select {
		case <-s.done:
			t.Fatalf("cmd/server exited before readiness: %v\nstdout:\n%s\nstderr:\n%s", s.terminalResult(), s.stdout.String(), s.stderr.String())
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

		time.Sleep(readinessPollInterval)
	}

	t.Fatalf("timed out waiting for readiness\nstdout:\n%s\nstderr:\n%s", s.stdout.String(), s.stderr.String())
}

func (s *Server) WaitForExit(t testing.TB) error {
	t.Helper()

	select {
	case <-s.done:
		return s.terminalResult()
	case <-time.After(exitDeadline):
		t.Fatalf("timed out waiting for cmd/server exit\nstdout:\n%s\nstderr:\n%s", s.stdout.String(), s.stderr.String())
		return nil
	}
}

func (s *Server) terminalResult() error {
	if s == nil {
		return nil
	}
	<-s.done
	s.resultMu.RLock()
	defer s.resultMu.RUnlock()
	return s.waitErr
}

func (s *Server) Stop(t testing.TB) {
	t.Helper()

	if s == nil {
		return
	}
	if err := s.stop(); err != nil {
		t.Fatalf("stop cmd/server: %v\nstdout:\n%s\nstderr:\n%s", err, s.stdout.String(), s.stderr.String())
	}
}

func (s *Server) stop() error {
	s.stopOnce.Do(func() {
		if err := s.group.signal(syscall.SIGTERM); err != nil {
			s.stopResult = fmt.Errorf("signal process group SIGTERM: %w", err)
			return
		}
		if s.waitForProcessGroup(gracefulStopDeadline) {
			return
		}
		if err := s.group.signal(syscall.SIGKILL); err != nil {
			s.stopResult = fmt.Errorf("signal process group SIGKILL: %w", err)
			return
		}
		if !s.waitForProcessGroup(forcedStopDeadline) {
			s.stopResult = fmt.Errorf("process group %d did not terminate within %s", s.group.pid, stopDeadline)
		}
	})
	return s.stopResult
}

func (s *Server) waitForProcessGroup(timeout time.Duration) bool {
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	ticker := time.NewTicker(processPollInterval)
	defer ticker.Stop()

	for {
		if s.processExited() && !s.group.alive() {
			return true
		}
		select {
		case <-timer.C:
			return s.processExited() && !s.group.alive()
		case <-ticker.C:
		}
	}
}

func (s *Server) processExited() bool {
	select {
	case <-s.done:
		return true
	default:
		return false
	}
}

func (s *Server) RequireStatus(t testing.TB, path string, want int) {
	t.Helper()

	client := &http.Client{Timeout: statusRequestTimeout}
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

	client := &http.Client{Timeout: refusalRequestTimeout}
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

	diagnostics, _, err := parseDiagnosticsJSON(s.stderr.String())
	if err != nil {
		t.Fatalf("invalid structured startup diagnostics on stderr: %v\nstderr:\n%s", err, s.stderr.String())
	}
	return diagnostics
}

func parseDiagnosticsJSON(stderr string) (string, map[string]any, error) {
	diagnostics := strings.TrimSpace(stderr)
	if diagnostics == "" {
		return "", nil, errors.New("expected exactly one JSON object, got empty stderr")
	}

	decoder := json.NewDecoder(strings.NewReader(diagnostics))
	var payload map[string]any
	if err := decoder.Decode(&payload); err != nil {
		return "", nil, fmt.Errorf("decode JSON object: %w", err)
	}
	if payload == nil {
		return "", nil, errors.New("diagnostics must be a JSON object")
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		if err == nil {
			return "", nil, errors.New("diagnostics contain more than one JSON value")
		}
		return "", nil, fmt.Errorf("decode trailing diagnostics: %w", err)
	}
	return diagnostics, payload, nil
}

func (s *Server) RequireDiagnosticsMatchGolden(t testing.TB, goldenParts []string) {
	t.Helper()

	diagnosticstest.RequireJSONMatchesGolden(t, s.DiagnosticsJSON(t), goldenParts)
}

func (s *Server) RequireDiagnosticsCode(t testing.TB, wantCode string) {
	t.Helper()

	payload := s.diagnosticsPayload(t)
	errorPayload, ok := payload["error"].(map[string]any)
	if !ok || errorPayload["code"] != wantCode {
		t.Fatalf("unexpected diagnostics code: got %v want %s", errorPayload["code"], wantCode)
	}
}

func (s *Server) RequireDiagnosticsField(t testing.TB, wantPath string, wantReasonCode string) {
	t.Helper()

	payload := s.diagnosticsPayload(t)
	errorPayload, ok := payload["error"].(map[string]any)
	if !ok {
		t.Fatalf("missing diagnostics error object in stderr payload %s", s.stderr.String())
	}
	details, ok := errorPayload["details"].(map[string]any)
	if !ok {
		t.Fatalf("missing diagnostics details object in stderr payload %s", s.stderr.String())
	}
	items, ok := details["items"].([]any)
	if !ok {
		t.Fatalf("missing diagnostics items array in stderr payload %s", s.stderr.String())
	}
	for _, item := range items {
		typed, ok := item.(map[string]any)
		if ok && typed["path"] == wantPath && typed["reason_code"] == wantReasonCode {
			return
		}
	}
	t.Fatalf("missing diagnostic path=%q reason_code=%q in stderr payload %s", wantPath, wantReasonCode, s.stderr.String())
}

func (s *Server) diagnosticsPayload(t testing.TB) map[string]any {
	t.Helper()

	_, payload, err := parseDiagnosticsJSON(s.stderr.String())
	if err != nil {
		t.Fatalf("invalid structured startup diagnostics on stderr: %v\nstderr:\n%s", err, s.stderr.String())
	}
	return payload
}

func envPairs(env map[string]string) []string {
	keys := make([]string, 0, len(env))
	for key := range env {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	pairs := make([]string, 0, len(keys))
	for _, key := range keys {
		pairs = append(pairs, key+"="+env[key])
	}
	return pairs
}

func serverCommand(t testing.TB) string {
	t.Helper()
	configured, err := resolveServerCommand(os.Getenv(serverHarnessBinEnv))
	if err != nil {
		t.Fatal(err)
	}
	return configured
}

func resolveServerCommand(rawConfigured string) (string, error) {
	configured := strings.TrimSpace(rawConfigured)
	if configured == "" {
		return "", fmt.Errorf("%s is required; run the owning public Make target", serverHarnessBinEnv)
	}
	info, err := os.Lstat(configured)
	if err != nil {
		return "", fmt.Errorf("%s is not usable: %w", serverHarnessBinEnv, err)
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() || info.Mode()&0o111 == 0 {
		return "", fmt.Errorf("%s must name a regular executable file: %s", serverHarnessBinEnv, configured)
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
