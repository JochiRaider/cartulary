package processtest

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"
)

const helperModeEnv = "CARTULARY_PROCESSTEST_HELPER_MODE"
const helperExpectedBaseURLEnv = "CARTULARY_PROCESSTEST_EXPECT_BASE_URL"

func TestMain(m *testing.M) {
	if mode := os.Getenv(helperModeEnv); mode != "" {
		os.Exit(runHelperProcess(mode))
	}
	os.Exit(m.Run())
}

func TestStartServerFinalizesEnvironmentFromInheritedListener(t *testing.T) {
	configureHelperProcess(t, "ready")
	callerEnv := map[string]string{
		"CALLER_OWNED":    "original",
		httpAddrEnv:       "127.0.0.1:0",
		httpListenFDEnv:   "caller-value",
		helperModeEnv:     "ready",
		"CALLBACK_MARKER": "before",
	}
	var callbackEnv map[string]string
	var callbackBaseURL string

	server := StartServer(t, ServerOptions{
		Env: callerEnv,
		FinalizeEnv: func(env map[string]string, baseURL string) {
			callbackEnv = env
			callbackBaseURL = baseURL
			env[helperExpectedBaseURLEnv] = baseURL
			env[httpAddrEnv] = "127.0.0.1:1"
			env[httpListenFDEnv] = "99"
			env["CALLBACK_MARKER"] = "finalized"
		},
	})
	t.Cleanup(func() { server.Stop(t) })

	callbackEnv["CALLBACK_MARKER"] = "late-mutation"
	callerEnv["CALLER_OWNED"] = "late-caller-mutation"
	if callbackBaseURL != server.BaseURL {
		t.Fatalf("FinalizeEnv base URL got %q want %q", callbackBaseURL, server.BaseURL)
	}
	if got := "http://" + clientAddress(stringAddress(callbackEnv[httpAddrEnv])); got != server.BaseURL {
		t.Fatalf("authoritative listener address got %q want %q", got, server.BaseURL)
	}
	if got := callbackEnv[httpListenFDEnv]; got != "3" {
		t.Fatalf("authoritative listener fd got %q want 3", got)
	}
	if got := callerEnv[httpListenFDEnv]; got != "caller-value" {
		t.Fatalf("StartServer mutated caller environment: %#v", callerEnv)
	}
	if server.cmd == nil || server.cmd.Process == nil || server.cmd.Process.Pid <= 0 {
		t.Fatalf("StartServer did not create a real child process: %#v", server.cmd)
	}
	if got, want := server.cmd.Path, os.Getenv(serverHarnessBinEnv); got != want {
		t.Fatalf("StartServer child command got %q want %q", got, want)
	}

	server.WaitForReady(t)
	server.RequireStatus(t, "/healthz", http.StatusOK)
	server.RequireStatus(t, "/readyz", http.StatusOK)
}

type stringAddress string

func (a stringAddress) Network() string { return "tcp" }

func (a stringAddress) String() string { return string(a) }

func TestServerProcessContractDefaults(t *testing.T) {
	if readinessDeadline != 15*time.Second ||
		readinessRequestTimeout != 500*time.Millisecond ||
		readinessPollInterval != 200*time.Millisecond ||
		exitDeadline != 15*time.Second ||
		statusRequestTimeout != 2*time.Second ||
		gracefulStopDeadline != 4*time.Second ||
		forcedStopDeadline != time.Second ||
		stopDeadline != 5*time.Second ||
		refusalRequestTimeout != 500*time.Millisecond {
		t.Fatalf("unexpected process contract deadlines")
	}
}

func TestResolveServerCommandRequiresRegularExecutable(t *testing.T) {
	dir := t.TempDir()
	executable := filepath.Join(dir, "server-harness")
	if err := os.WriteFile(executable, []byte("#!/bin/sh\n"), 0o700); err != nil {
		t.Fatalf("write executable fixture: %v", err)
	}
	nonExecutable := filepath.Join(dir, "not-executable")
	if err := os.WriteFile(nonExecutable, []byte("fixture"), 0o600); err != nil {
		t.Fatalf("write non-executable fixture: %v", err)
	}
	directory := filepath.Join(dir, "directory")
	if err := os.Mkdir(directory, 0o700); err != nil {
		t.Fatalf("write directory fixture: %v", err)
	}
	symlink := filepath.Join(dir, "server-harness-link")
	if err := os.Symlink(executable, symlink); err != nil {
		t.Fatalf("write symlink fixture: %v", err)
	}

	if got, err := resolveServerCommand("  " + executable + "  "); err != nil || got != executable {
		t.Fatalf("resolve executable got %q err=%v", got, err)
	}
	for name, configured := range map[string]string{
		"missing":        "",
		"unknown":        filepath.Join(dir, "missing"),
		"non executable": nonExecutable,
		"directory":      directory,
		"symlink":        symlink,
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := resolveServerCommand(configured); err == nil {
				t.Fatalf("resolveServerCommand(%q) unexpectedly succeeded", configured)
			}
		})
	}
}

func TestResolveRepoRootHasNoFallback(t *testing.T) {
	t.Chdir(t.TempDir())
	if root, err := resolveRepoRoot(); err == nil {
		t.Fatalf("resolveRepoRoot unexpectedly returned fallback %q", root)
	}
}

func TestWaitForExitIsRepeatableAndBroadcast(t *testing.T) {
	configureHelperProcess(t, "exit-error")
	server := StartServer(t, ServerOptions{})

	const waiterCount = 8
	results := make(chan error, waiterCount)
	var waiters sync.WaitGroup
	for range waiterCount {
		waiters.Add(1)
		go func() {
			defer waiters.Done()
			results <- server.WaitForExit(t)
		}()
	}
	waiters.Wait()
	close(results)
	var first string
	for err := range results {
		if err == nil {
			t.Fatal("expected non-zero helper exit")
		}
		if first == "" {
			first = err.Error()
		} else if err.Error() != first {
			t.Fatalf("waiters observed different results: got %q want %q", err, first)
		}
	}
	if err := server.WaitForExit(t); err == nil || err.Error() != first {
		t.Fatalf("repeated wait got %v want %q", err, first)
	}
	server.RequireDiagnosticsCode(t, "invalid_deployment_config")
	server.RequireDiagnosticsField(t, "x", "y")
}

func TestStopIsIdempotentAndConcurrent(t *testing.T) {
	configureHelperProcess(t, "ready")
	server := StartServer(t, ServerOptions{})
	server.WaitForReady(t)

	const stopperCount = 8
	results := make(chan error, stopperCount)
	var stoppers sync.WaitGroup
	for range stopperCount {
		stoppers.Add(1)
		go func() {
			defer stoppers.Done()
			results <- server.stop()
		}()
	}
	stoppers.Wait()
	close(results)
	for err := range results {
		if err != nil {
			t.Fatalf("concurrent stop: %v", err)
		}
	}
	server.Stop(t)
	server.Stop(t)
	if err := server.WaitForExit(t); err != nil {
		t.Fatalf("graceful helper exit: %v", err)
	}
	server.RequireConnectionRefused(t, "/healthz")
	server.RequireWebsocketConnectionRefused(t, "/ws/v1/bootstrap-harness")
}

func TestStopAfterNaturalExit(t *testing.T) {
	configureHelperProcess(t, "exit-zero")
	server := StartServer(t, ServerOptions{})
	if err := server.WaitForExit(t); err != nil {
		t.Fatalf("natural helper exit: %v", err)
	}
	server.Stop(t)
	server.Stop(t)
}

func TestStopForcesIgnoredSIGTERMWithinTotalBound(t *testing.T) {
	configureHelperProcess(t, "ignore-term")
	server := StartServer(t, ServerOptions{})
	server.WaitForReady(t)

	started := time.Now()
	if err := server.stop(); err != nil {
		t.Fatal(err)
	}
	elapsed := time.Since(started)
	if elapsed < gracefulStopDeadline {
		t.Fatalf("ignored SIGTERM was not given the graceful interval: %s", elapsed)
	}
	if elapsed > stopDeadline+250*time.Millisecond {
		t.Fatalf("forced stop exceeded total bound: %s", elapsed)
	}
	if err := server.WaitForExit(t); err == nil {
		t.Fatal("expected forced process exit to report a signal")
	}
}

func TestStopTerminatesProcessGroupDescendants(t *testing.T) {
	configureHelperProcess(t, "descendant")
	server := StartServer(t, ServerOptions{})
	server.WaitForReady(t)
	descendantPID := waitForDescendantPID(t, server)

	if err := server.stop(); err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(forcedStopDeadline)
	for processIsRunning(descendantPID) && time.Now().Before(deadline) {
		time.Sleep(processPollInterval)
	}
	if processIsRunning(descendantPID) {
		t.Fatalf("descendant process %d survived process-group stop", descendantPID)
	}
}

func TestProcessOutputIsRetainedAfterExit(t *testing.T) {
	configureHelperProcess(t, "exit-output")
	server := StartServer(t, ServerOptions{})
	if err := server.WaitForExit(t); err != nil {
		t.Fatalf("output helper exit: %v", err)
	}
	if got := server.stdout.String(); !strings.Contains(got, "stdout-retained") {
		t.Fatalf("missing retained stdout: %q", got)
	}
	if got := server.stderr.String(); !strings.Contains(got, "stderr-retained") {
		t.Fatalf("missing retained stderr: %q", got)
	}
}

func TestDiagnosticsJSONRequiresExactlyOneJSONObject(t *testing.T) {
	valid := `{"error":{"code":"invalid_deployment_config","details":{"items":[{"path":"x","reason_code":"y"}]}}}`
	if got, payload, err := parseDiagnosticsJSON(" \n" + valid + "\n"); err != nil || got != valid || payload["error"] == nil {
		t.Fatalf("parse valid diagnostics got=%q payload=%#v err=%v", got, payload, err)
	}
	for name, diagnostics := range map[string]string{
		"empty":         " \n\t",
		"malformed":     `{"error":`,
		"array":         `[{"error":{}}]`,
		"scalar":        `"error"`,
		"trailing":      valid + "\nexit status 1",
		"multiple":      valid + "\n" + valid,
		"trailing-junk": valid + " trailing",
	} {
		t.Run(name, func(t *testing.T) {
			if got, payload, err := parseDiagnosticsJSON(diagnostics); err == nil {
				t.Fatalf("unexpectedly accepted diagnostics got=%q payload=%#v", got, payload)
			}
		})
	}
}

func configureHelperProcess(t testing.TB, mode string) {
	t.Helper()
	executable, err := os.Executable()
	if err != nil {
		t.Fatalf("resolve processtest helper executable: %v", err)
	}
	t.Setenv(serverHarnessBinEnv, executable)
	t.Setenv(helperModeEnv, mode)
}

func runHelperProcess(mode string) int {
	switch mode {
	case "ready":
		return serveInheritedListener(false, false)
	case "ignore-term":
		return serveInheritedListener(true, false)
	case "descendant":
		return serveInheritedListener(false, true)
	case "descendant-child":
		signal.Ignore(syscall.SIGTERM)
		ready := os.NewFile(3, "processtest-descendant-ready")
		if ready == nil {
			return 2
		}
		_, _ = ready.Write([]byte{1})
		_ = ready.Close()
		select {}
	case "exit-zero":
		return 0
	case "exit-error":
		_, _ = fmt.Fprintln(os.Stderr, `{"error":{"code":"invalid_deployment_config","details":{"items":[{"path":"x","reason_code":"y"}]}}}`)
		return 2
	case "exit-output":
		_, _ = fmt.Fprintln(os.Stdout, "stdout-retained")
		_, _ = fmt.Fprintln(os.Stderr, "stderr-retained")
		return 0
	default:
		_, _ = fmt.Fprintf(os.Stderr, "unknown processtest helper mode %q\n", mode)
		return 2
	}
}

func serveInheritedListener(ignoreTermination bool, withDescendant bool) int {
	file := os.NewFile(3, "processtest-listener")
	if file == nil {
		_, _ = fmt.Fprintln(os.Stderr, "missing inherited listener fd 3")
		return 2
	}
	listener, err := net.FileListener(file)
	_ = file.Close()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "open inherited listener: %v\n", err)
		return 2
	}

	if expected := os.Getenv(helperExpectedBaseURLEnv); expected != "" {
		actual := "http://" + clientAddress(listener.Addr())
		if expected != actual || os.Getenv(httpAddrEnv) != listener.Addr().String() || os.Getenv(httpListenFDEnv) != "3" || os.Getenv("CALLBACK_MARKER") != "finalized" || os.Getenv("CALLER_OWNED") != "original" {
			_, _ = fmt.Fprintf(os.Stderr, `{"error":{"code":"invalid_final_environment","expected":%q,"actual":%q}}`, expected, actual)
			return 2
		}
	}

	if withDescendant {
		readyReader, readyWriter, err := os.Pipe()
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "create descendant readiness pipe: %v\n", err)
			return 2
		}
		child := exec.Command(os.Args[0])
		child.Env = replaceEnvValue(os.Environ(), helperModeEnv, "descendant-child")
		child.ExtraFiles = []*os.File{readyWriter}
		if err := child.Start(); err != nil {
			_ = readyReader.Close()
			_ = readyWriter.Close()
			_, _ = fmt.Fprintf(os.Stderr, "start descendant: %v\n", err)
			return 2
		}
		_ = readyWriter.Close()
		_ = readyReader.SetReadDeadline(time.Now().Add(time.Second))
		readiness := []byte{0}
		if _, err := readyReader.Read(readiness); err != nil || readiness[0] != 1 {
			_ = readyReader.Close()
			_, _ = fmt.Fprintf(os.Stderr, "wait for descendant readiness: byte=%d err=%v\n", readiness[0], err)
			return 2
		}
		_ = readyReader.Close()
		_, _ = fmt.Fprintf(os.Stdout, "DESCENDANT_PID=%d\n", child.Process.Pid)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(writer http.ResponseWriter, _ *http.Request) { writer.WriteHeader(http.StatusOK) })
	mux.HandleFunc("/readyz", func(writer http.ResponseWriter, _ *http.Request) { writer.WriteHeader(http.StatusOK) })
	server := &http.Server{Handler: mux, ReadHeaderTimeout: time.Second}
	if ignoreTermination {
		signal.Ignore(syscall.SIGTERM)
	} else {
		signals := make(chan os.Signal, 1)
		signal.Notify(signals, syscall.SIGTERM)
		defer signal.Stop(signals)
		go func() {
			<-signals
			_ = server.Close()
		}()
	}
	if err := server.Serve(listener); err != nil && err != http.ErrServerClosed {
		_, _ = fmt.Fprintf(os.Stderr, "serve inherited listener: %v\n", err)
		return 2
	}
	return 0
}

func replaceEnvValue(env []string, key string, value string) []string {
	prefix := key + "="
	replaced := make([]string, 0, len(env)+1)
	for _, pair := range env {
		if !strings.HasPrefix(pair, prefix) {
			replaced = append(replaced, pair)
		}
	}
	return append(replaced, prefix+value)
}

func waitForDescendantPID(t testing.TB, server *Server) int {
	t.Helper()
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		for _, line := range strings.Split(server.stdout.String(), "\n") {
			if rawPID, ok := strings.CutPrefix(line, "DESCENDANT_PID="); ok {
				pid, err := strconv.Atoi(rawPID)
				if err != nil {
					t.Fatalf("parse descendant PID %q: %v", rawPID, err)
				}
				return pid
			}
		}
		time.Sleep(processPollInterval)
	}
	t.Fatalf("missing descendant PID in stdout %q", server.stdout.String())
	return 0
}

func processIsRunning(pid int) bool {
	if pid <= 0 {
		return false
	}
	stat, err := os.ReadFile(filepath.Join("/proc", strconv.Itoa(pid), "stat"))
	if errors.Is(err, os.ErrNotExist) {
		return false
	}
	if err != nil {
		return true
	}
	fields := strings.Fields(string(stat))
	return len(fields) < 3 || fields[2] != "Z"
}
