package main

import (
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"
)

const testOrigin = "http://localhost:5173"

func TestProxyPreflightExactPolicy(t *testing.T) {
	upstream := httptest.NewServer(http.NotFoundHandler())
	t.Cleanup(upstream.Close)
	handler := newProxyHandler(parseTestURL(t, upstream.URL), testOrigin)

	req := httptest.NewRequest(http.MethodOptions, "http://object-store.test/cartulary/object.bin", nil)
	req.Header.Set("Origin", testOrigin)
	req.Header.Set("Access-Control-Request-Method", http.MethodPut)
	req.Header.Set("Access-Control-Request-Headers", "x-amz-checksum-sha256, content-type")
	resp := httptest.NewRecorder()
	handler.ServeHTTP(resp, req)

	if resp.Code != http.StatusNoContent {
		t.Fatalf("preflight status got %d want %d", resp.Code, http.StatusNoContent)
	}
	if got := resp.Header().Get("Access-Control-Allow-Origin"); got != testOrigin {
		t.Fatalf("allow origin got %q want %q", got, testOrigin)
	}
	if got := resp.Header().Get("Access-Control-Allow-Methods"); got != "PUT, OPTIONS" {
		t.Fatalf("allow methods got %q", got)
	}
	if got := resp.Header().Get("Access-Control-Allow-Headers"); got != "content-type, x-amz-checksum-sha256" {
		t.Fatalf("allow headers got %q", got)
	}
	if got := resp.Header().Get("Access-Control-Allow-Credentials"); got != "" {
		t.Fatalf("credentials header must be absent, got %q", got)
	}
	if got := resp.Header().Get("Access-Control-Max-Age"); got != "600" {
		t.Fatalf("max age got %q want 600", got)
	}
}

func TestProxyPreflightRejectsDisallowedOriginMethodAndHeader(t *testing.T) {
	upstream := httptest.NewServer(http.NotFoundHandler())
	t.Cleanup(upstream.Close)
	handler := newProxyHandler(parseTestURL(t, upstream.URL), testOrigin)

	tests := []struct {
		name    string
		origin  string
		method  string
		headers string
	}{
		{name: "null origin", origin: "null", method: http.MethodPut, headers: "content-type"},
		{name: "extra method", origin: testOrigin, method: http.MethodGet, headers: "content-type"},
		{name: "extra header", origin: testOrigin, method: http.MethodPut, headers: "content-type, authorization"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodOptions, "http://object-store.test/cartulary/object.bin", nil)
			req.Header.Set("Origin", tc.origin)
			req.Header.Set("Access-Control-Request-Method", tc.method)
			req.Header.Set("Access-Control-Request-Headers", tc.headers)
			resp := httptest.NewRecorder()
			handler.ServeHTTP(resp, req)

			if resp.Code < 400 {
				t.Fatalf("preflight status got %d want rejection", resp.Code)
			}
			if got := resp.Header().Get("Access-Control-Allow-Origin"); got != "" {
				t.Fatalf("rejected preflight must not grant origin, got %q", got)
			}
		})
	}
}

func TestProxyPUTPreservesHostAndNormalizesCORS(t *testing.T) {
	var gotHost string
	var gotBody []byte
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotHost = r.Host
		gotBody, _ = io.ReadAll(r.Body)
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Expose-Headers", "etag, x-amz-request-id")
		w.Header().Set("ETag", `"abc"`)
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(upstream.Close)
	handler := newProxyHandler(parseTestURL(t, upstream.URL), testOrigin)

	req := httptest.NewRequest(http.MethodPut, "http://signed-host.test/cartulary/object.bin", strings.NewReader("payload"))
	req.Header.Set("Origin", testOrigin)
	resp := httptest.NewRecorder()
	handler.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("PUT status got %d want %d", resp.Code, http.StatusOK)
	}
	if gotHost != "signed-host.test" {
		t.Fatalf("upstream host got %q want signed host", gotHost)
	}
	if string(gotBody) != "payload" {
		t.Fatalf("upstream body got %q", string(gotBody))
	}
	if got := resp.Header().Get("Access-Control-Allow-Origin"); got != testOrigin {
		t.Fatalf("allow origin got %q want %q", got, testOrigin)
	}
	if got := resp.Header().Get("Access-Control-Expose-Headers"); got != "etag" {
		t.Fatalf("expose headers got %q want etag", got)
	}
}

func TestProxyConfigurationRejectsAmbiguousOrNonLoopbackInputs(t *testing.T) {
	tests := []struct {
		name     string
		listen   string
		upstream string
		origin   string
	}{
		{
			name:     "non-loopback listener",
			listen:   "0.0.0.0:8333",
			upstream: "http://127.0.0.1:18333",
			origin:   testOrigin,
		},
		{
			name:     "upstream userinfo",
			listen:   "127.0.0.1:8333",
			upstream: "http://user@127.0.0.1:18333",
			origin:   testOrigin,
		},
		{
			name:     "upstream query",
			listen:   "127.0.0.1:8333",
			upstream: "http://127.0.0.1:18333?mode=stale",
			origin:   testOrigin,
		},
		{
			name:     "allowed-origin fragment",
			listen:   "127.0.0.1:8333",
			upstream: "http://127.0.0.1:18333",
			origin:   "http://localhost:5173#stale",
		},
		{
			name:     "allowed-origin path",
			listen:   "127.0.0.1:8333",
			upstream: "http://127.0.0.1:18333",
			origin:   "http://localhost:5173/path",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := normalizeConfig(tc.listen, tc.upstream, tc.origin); err == nil {
				t.Fatal("unsafe proxy configuration must be rejected")
			}
		})
	}
}

func TestProxyLifecycleTreatsPreProofAttemptAsUntrusted(t *testing.T) {
	temp := t.TempDir()
	config, err := normalizeConfig(
		availableLoopbackAddress(t),
		"http://127.0.0.1:18333",
		testOrigin,
	)
	if err != nil {
		t.Fatalf("normalize config: %v", err)
	}
	attemptFile := filepath.Join(temp, "attempt.json")
	options := commandOptions{
		attemptFile: attemptFile,
		instanceID:  "pre-proof-crash",
		logPath:     filepath.Join(temp, "pre-proof.log"),
	}
	if err := createAttempt(options, config); err != nil {
		t.Fatalf("create attempt: %v", err)
	}
	var coded *exitError
	if err := status(attemptFile, config); !errors.As(err, &coded) || coded.code != 1 {
		t.Fatalf("pre-proof attempt status got %v, want unproven exit code 1", err)
	}
	if err := stop(attemptFile); err == nil {
		t.Fatal("pre-proof attempt must never authorize signaling")
	}
	if err := secureRemove(attemptFile); err != nil {
		t.Fatalf("discard pre-proof attempt: %v", err)
	}
}

func TestProxyLifecycleUsesImmutableProofAndPidfdStop(t *testing.T) {
	if os.Getenv("CARTULARY_S3CORSPROXY_HELPER") == "1" {
		return
	}
	temp := t.TempDir()
	listen := availableLoopbackAddress(t)
	attemptFile := filepath.Join(temp, "attempt.json")
	leaseFile := filepath.Join(temp, "lease.json")
	logPath := filepath.Join(temp, "logs", "instance.log")
	instanceID := "test-instance"
	config, err := normalizeConfig(listen, "http://127.0.0.1:18333", testOrigin)
	if err != nil {
		t.Fatalf("normalize config: %v", err)
	}
	options := commandOptions{
		listen:      listen,
		upstream:    config.UpstreamOrigin,
		origin:      config.AllowedOrigin,
		attemptFile: attemptFile,
		leaseFile:   leaseFile,
		instanceID:  instanceID,
		logPath:     logPath,
	}
	if err := createAttempt(options, config); err != nil {
		t.Fatalf("create attempt: %v", err)
	}
	child := startProxyHelper(t, options)
	t.Cleanup(func() {
		if child.ProcessState == nil || !child.ProcessState.Exited() {
			_ = child.Process.Kill()
			_, _ = child.Process.Wait()
		}
	})
	waitForStatus(t, attemptFile, config)
	if err := promote(options, config); err != nil {
		t.Fatalf("promote attempt: %v", err)
	}
	if _, err := os.Stat(attemptFile); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("promoted attempt must be removed, got %v", err)
	}
	if err := status(leaseFile, config); err != nil {
		t.Fatalf("ready lease status: %v", err)
	}
	mismatched, err := normalizeConfig(listen, config.UpstreamOrigin, "http://localhost:5174")
	if err != nil {
		t.Fatalf("normalize mismatch config: %v", err)
	}
	var coded *exitError
	if err := status(leaseFile, mismatched); !errors.As(err, &coded) || coded.code != 3 {
		t.Fatalf("mismatched proven config got %v, want exit code 3", err)
	}
	if err := stop(leaseFile); err != nil {
		t.Fatalf("pidfd stop: %v", err)
	}
	if err := child.Wait(); err != nil {
		t.Fatalf("proxy helper exit: %v", err)
	}
	if err := stop(leaseFile); err != nil {
		t.Fatalf("repeated stop: %v", err)
	}
}

func TestProxyLifecycleRefusesMutatedProofAndUnownedListener(t *testing.T) {
	if os.Getenv("CARTULARY_S3CORSPROXY_HELPER") == "1" {
		return
	}
	temp := t.TempDir()
	listen := availableLoopbackAddress(t)
	config, err := normalizeConfig(listen, "http://127.0.0.1:18333", testOrigin)
	if err != nil {
		t.Fatalf("normalize config: %v", err)
	}
	attemptFile := filepath.Join(temp, "attempt.json")
	leaseFile := filepath.Join(temp, "lease.json")
	options := commandOptions{
		listen:      listen,
		upstream:    config.UpstreamOrigin,
		origin:      config.AllowedOrigin,
		attemptFile: attemptFile,
		leaseFile:   leaseFile,
		instanceID:  "proof-test",
		logPath:     filepath.Join(temp, "proof.log"),
	}
	if err := createAttempt(options, config); err != nil {
		t.Fatalf("create attempt: %v", err)
	}
	child := startProxyHelper(t, options)
	t.Cleanup(func() {
		if child.ProcessState == nil || !child.ProcessState.Exited() {
			_ = child.Process.Kill()
			_, _ = child.Process.Wait()
		}
	})
	waitForStatus(t, attemptFile, config)
	if err := promote(options, config); err != nil {
		t.Fatalf("promote attempt: %v", err)
	}
	lease, err := readState(leaseFile)
	if err != nil {
		t.Fatalf("read lease: %v", err)
	}
	raw, err := secureRead(leaseFile)
	if err != nil {
		t.Fatalf("read lease bytes: %v", err)
	}
	var payload proxyLease
	if err := json.Unmarshal(raw, &payload); err != nil {
		t.Fatalf("decode lease: %v", err)
	}
	proofMutations := []struct {
		name   string
		mutate func(*processProof)
	}{
		{
			name: "boot mismatch",
			mutate: func(proof *processProof) {
				proof.BootID = "different-boot"
			},
		},
		{
			name: "pid reuse start ticks",
			mutate: func(proof *processProof) {
				proof.StartTimeTicks++
			},
		},
		{
			name: "executable rebuild inode",
			mutate: func(proof *processProof) {
				proof.ExecutableInode++
			},
		},
		{
			name: "executable bytes mismatch",
			mutate: func(proof *processProof) {
				proof.ExecutableSHA256 = "sha256:" + strings.Repeat("0", 64)
			},
		},
	}
	for _, mutation := range proofMutations {
		t.Run(mutation.name, func(t *testing.T) {
			payload.Process = lease.process
			mutation.mutate(&payload.Process)
			if err := atomicWriteJSON(leaseFile, payload); err != nil {
				t.Fatalf("mutate lease: %v", err)
			}
			if err := status(leaseFile, config); err == nil {
				t.Fatal("mutated proof must not be reusable")
			}
			if err := stop(leaseFile); err == nil {
				t.Fatal("mutated proof must refuse signaling")
			}
			if err := child.Process.Signal(syscall.Signal(0)); err != nil {
				t.Fatalf("child must remain alive after refused signaling: %v", err)
			}
		})
	}
	payload.Process = lease.process
	if err := atomicWriteJSON(leaseFile, payload); err != nil {
		t.Fatalf("restore lease: %v", err)
	}
	if err := stop(leaseFile); err != nil {
		t.Fatalf("stop restored lease: %v", err)
	}
	if err := child.Wait(); err != nil {
		t.Fatalf("proxy helper exit: %v", err)
	}

	unowned, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen unowned: %v", err)
	}
	defer unowned.Close()
	unownedConfig, err := normalizeConfig(
		unowned.Addr().String(),
		"http://127.0.0.1:18333",
		testOrigin,
	)
	if err != nil {
		t.Fatalf("normalize unowned config: %v", err)
	}
	unownedOptions := options
	unownedOptions.listen = unownedConfig.Listen
	unownedOptions.attemptFile = filepath.Join(temp, "unowned-attempt.json")
	unownedOptions.instanceID = "unowned-test"
	if err := createAttempt(unownedOptions, unownedConfig); err != nil {
		t.Fatalf("create unowned attempt: %v", err)
	}
	conflicted := startProxyHelper(t, unownedOptions)
	if err := conflicted.Wait(); err == nil {
		t.Fatal("unowned listener must make the synchronous bind fail")
	}
	attempt, err := readAttempt(unownedOptions.attemptFile)
	if err != nil {
		t.Fatalf("read conflicted attempt: %v", err)
	}
	if attempt.State != "launching" || attempt.Process != nil {
		t.Fatalf("bind conflict must not publish process proof: %+v", attempt)
	}
}

func TestProxyHelperProcess(t *testing.T) {
	if os.Getenv("CARTULARY_S3CORSPROXY_HELPER") != "1" {
		return
	}
	var values map[string]string
	if err := json.Unmarshal([]byte(os.Getenv("CARTULARY_S3CORSPROXY_OPTIONS")), &values); err != nil {
		os.Exit(2)
	}
	options := commandOptions{
		listen:      values["listen"],
		upstream:    values["upstream"],
		origin:      values["origin"],
		attemptFile: values["attempt_file"],
		instanceID:  values["instance_id"],
		logPath:     values["log_path"],
	}
	config, err := normalizeConfig(options.listen, options.upstream, options.origin)
	if err == nil {
		err = serve(options, config)
	}
	if err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}

func availableLoopbackAddress(t *testing.T) string {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("allocate port: %v", err)
	}
	address := listener.Addr().String()
	if err := listener.Close(); err != nil {
		t.Fatalf("release port: %v", err)
	}
	return address
}

func startProxyHelper(t *testing.T, options commandOptions) *exec.Cmd {
	t.Helper()
	executable, err := os.Executable()
	if err != nil {
		t.Fatalf("test executable: %v", err)
	}
	serialized, err := json.Marshal(map[string]string{
		"listen":       options.listen,
		"upstream":     options.upstream,
		"origin":       options.origin,
		"attempt_file": options.attemptFile,
		"instance_id":  options.instanceID,
		"log_path":     options.logPath,
	})
	if err != nil {
		t.Fatalf("serialize helper options: %v", err)
	}
	command := exec.Command(executable, "-test.run=^TestProxyHelperProcess$")
	command.Env = append(
		os.Environ(),
		"CARTULARY_S3CORSPROXY_HELPER=1",
		"CARTULARY_S3CORSPROXY_OPTIONS="+string(serialized),
	)
	if err := command.Start(); err != nil {
		t.Fatalf("start proxy helper: %v", err)
	}
	return command
}

func waitForStatus(t *testing.T, stateFile string, config proxyConfig) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if err := status(stateFile, config); err == nil {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatalf("proxy state %s did not become proven", stateFile)
}

func parseTestURL(t *testing.T, raw string) *url.URL {
	t.Helper()
	parsed, err := url.Parse(raw)
	if err != nil {
		t.Fatalf("parse test URL: %v", err)
	}
	return parsed
}
