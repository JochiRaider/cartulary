//go:build linux

package harnessfixtures

import (
	"context"
	"crypto/x509"
	"encoding/json"
	"errors"
	"github.com/JochiRaider/cartulary/internal/testutil/harnessadmission"
	"github.com/JochiRaider/cartulary/internal/testutil/harnesssession"
	"github.com/JochiRaider/cartulary/internal/testutil/objectstoreproxy"
	"github.com/JochiRaider/cartulary/internal/testutil/tlstest"
	"io"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"testing"
	"time"
)

func TestFixtureProxyLifecycle(t *testing.T) {
	if mode := os.Getenv("CARTULARY_PROXY_RECOVERY_TEST"); mode != "" {
		proxyRecoveryHelper(t, mode)
		return
	}
	t.Run("owner_restart_after_failed_stop", testProxyRecoveryRestart)
	for _, variant := range []string{"promotion_failure", "missing_ready_proof"} {
		t.Run(variant, func(t *testing.T) {
			runtime, root, run := runtimeFixture(t)
			if err := os.WriteFile(filepath.Join(runtime.Root, "root.pem"), []byte("fixture certificate"), 0600); err != nil {
				t.Fatal(err)
			}
			suiteFile := filepath.Join(runtime.Root, "suite-lease.json")
			if err := runtime.RecordDependentResource("managed_suite", suiteFile, "acquired", []string{}); err != nil {
				t.Fatal(err)
			}
			defer runtime.RecordResource("managed_suite", suiteFile, "released")
			suite := &SuiteOwner{Acquire: func(context.Context) (*Suite, error) {
				return &Suite{LeaseFile: suiteFile, ResourceRef: fixtureReference(t, runtime, "managed_suite", suiteFile), Environment: map[string]string{"CARTULARY_S3TEST_ENDPOINT": "localhost:1", "CARTULARY_S3TEST_PROXY_CERTIFICATE_PATH": "/fixture/cert", "CARTULARY_S3TEST_PROXY_PRIVATE_KEY_PATH": "/fixture/key", "SSL_CERT_FILE": filepath.Join(runtime.Root, "root.pem")}}, nil
			}}
			ready := ""
			options := ProductionOptions{Root: root, RunRoot: run, Runtime: runtime, Environment: map[string]string{"GO": "fixture-go", "GO_CACHE_DIR": root, "GO_MOD_CACHE_DIR": root, "GO_TMP_DIR": root, "PATH": "/usr/bin:/bin"}}
			options.Run = func(_ context.Context, _ string, args []string, _ map[string]string) (CommandResult, error) {
				flag := func(key string) string {
					for i := range args {
						if args[i] == key && i+1 < len(args) {
							return args[i+1]
						}
					}
					return ""
				}
				switch args[0] {
				case "build":
					if err := os.WriteFile(flag("-o"), []byte("#!/bin/sh\ntrap 'exit 0' TERM\nwhile :; do sleep 1; done\n"), 0700); err != nil {
						t.Fatal(err)
					}
				case "attempt":
					writeFixtureJSON(t, flag("--attempt-file"), map[string]string{"proof": "private attempt"})
				case "status":
				case "promote":
					if variant == "promotion_failure" {
						return CommandResult{Started: true, Exit: 1}, errors.New("interrupted publication")
					}
					ready = flag("--lease-file")
					writeFixtureJSON(t, ready, map[string]string{"proof": "private ready"})
				default:
					t.Fatalf("unexpected adapter call %v", args)
				}
				return CommandResult{Started: true}, nil
			}
			provider := &proxyProvider{options: options, suite: suite}
			allocation, err := provider.Acquire(context.Background(), Request{})
			if variant == "promotion_failure" {
				if err == nil {
					t.Fatal("failed promotion admitted")
				}
			} else {
				if err != nil || allocation == nil {
					t.Fatal(err)
				}
				if err := os.Remove(ready); err != nil {
					t.Fatal(err)
				}
				if err := allocation.Release(context.Background(), true); err == nil {
					t.Fatal("missing proof silently accepted")
				}
			}
			resources, err := runtime.Resources(64)
			if err != nil {
				t.Fatal(err)
			}
			for _, resource := range resources {
				if resource.Kind == "helper_process" {
					t.Fatal("proxy child was not drained after proof failure")
				}
				path, err := resource.Path()
				if err != nil || filepath.Dir(path) == "." {
					t.Fatal(resource)
				}
				// This test owns the synthetic physical protocol. A failed proof deliberately
				// remains available; remove only the test's journal after checking the drain.
				if err := runtime.RecordResource(resource.Kind, path, "released"); err != nil {
					t.Fatal(err)
				}
			}
		})
	}
}

func proxyRecoveryHelper(t *testing.T, mode string) {
	if mode == "serve" {
		var args []string
		if err := json.Unmarshal([]byte(os.Getenv("CARTULARY_PROXY_RECOVERY_ARGS")), &args); err != nil {
			t.Fatal(err)
		}
		os.Exit(objectstoreproxy.Main(args))
	}
	base := os.Getenv("CARTULARY_PROXY_RECOVERY_BASE")
	r, err := harnesssession.NewRuntime(filepath.Join(base, "repo"), filepath.Join(base, "run"), "proxy-recovery", filepath.Join(base, "scratch"))
	if err != nil {
		t.Fatal(err)
	}
	ca, err := tlstest.NewAuthority()
	if err != nil {
		t.Fatal(err)
	}
	identity, err := ca.Issue("proxy", []string{"127.0.0.1"}, x509.ExtKeyUsageServerAuth)
	if err != nil {
		t.Fatal(err)
	}
	write := func(name string, data []byte) string {
		file, err := tlstest.WriteFile(r.Root, name, data)
		if err != nil {
			t.Fatal(err)
		}
		return file
	}
	cert, key, trust := write("server.pem", identity.CertificatePEM), write("server.key", identity.PrivateKeyPEM), write("root.pem", ca.CertificatePEM)
	listener, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	address := listener.Addr().String()
	listener.Close()
	attempt, lease := filepath.Join(r.Root, "attempt.json"), filepath.Join(r.Root, "lease.json")
	common := []string{"--listen", address, "--upstream", "https://127.0.0.1:1", "--origin", "https://localhost:5173", "--tls-certificate", cert, "--tls-private-key", key, "--tls-root-certificate", trust}
	args := append(append([]string{"attempt"}, common...), "--attempt-file", attempt, "--instance-id", "restart-test", "--log-path", filepath.Join(r.Root, "proxy.log"))
	if code := objectstoreproxy.Main(args); code != 0 {
		t.Fatal("attempt", code)
	}
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	in, err := os.Open(executable)
	if err != nil {
		t.Fatal(err)
	}
	defer in.Close()
	disposable := filepath.Join(r.Root, "proxy-binary")
	out, err := os.OpenFile(disposable, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0700)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := io.Copy(out, in); err != nil {
		t.Fatal(err)
	}
	if err := out.Close(); err != nil {
		t.Fatal(err)
	}
	args[0] = "serve"
	encoded, _ := json.Marshal(args)
	child := exec.Command(disposable, "-test.run=^TestFixtureProxyLifecycle$")
	child.Env = append(os.Environ(), "CARTULARY_PROXY_RECOVERY_TEST=serve", "CARTULARY_PROXY_RECOVERY_ARGS="+string(encoded))
	if err := child.Start(); err != nil {
		t.Fatal(err)
	}
	success := false
	defer func() {
		if !success {
			child.Process.Kill()
			child.Wait()
		}
	}()
	status := append(append([]string{"status"}, common...), "--state-file", attempt)
	ready := false
	for i := 0; i < 200; i++ {
		if objectstoreproxy.Main(status) == 0 {
			ready = true
			break
		}
		time.Sleep(25 * time.Millisecond)
	}
	if !ready {
		t.Fatal("proxy readiness")
	}
	promotion := append(append([]string{"promote"}, common...), "--attempt-file", attempt, "--lease-file", lease)
	if code := objectstoreproxy.Main(promotion); code != 0 {
		t.Fatal("promote", code)
	}
	if err := r.RecordDependentResource("object_store_proxy", lease, "acquired", []string{}, trust); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(trust, []byte("invalid certificate"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := objectstoreproxy.Stop(lease); err == nil {
		t.Fatal("stop should fail without TLS authority")
	}
	if err := child.Process.Signal(syscall.Signal(0)); err != nil {
		t.Fatal("failed stop killed process", err)
	}
	if err := os.WriteFile(trust, ca.CertificatePEM, 0600); err != nil {
		t.Fatal(err)
	}
	write("credentials.log", []byte("disposable private credential"))
	if err := r.PreserveRecovery(); err != nil {
		t.Fatal(err)
	}
	data, _ := json.Marshal(map[string]string{"root": r.Root, "lease": r.LeaseID})
	if err := os.WriteFile(filepath.Join(base, "identity.json"), data, 0600); err != nil {
		t.Fatal(err)
	}
	success = true
	// Exit with the physical proxy unresolved. The next OS process must recover
	// using durable proof; its deleted executable cannot be invoked again.
	os.Exit(0)
}

func testProxyRecoveryRestart(t *testing.T) {
	base := t.TempDir()
	for _, name := range []string{"repo", "run", "scratch"} {
		if err := os.Mkdir(filepath.Join(base, name), 0700); err != nil {
			t.Fatal(err)
		}
	}
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	owner := exec.Command(executable, "-test.run=^TestFixtureProxyLifecycle$")
	owner.Env = append(os.Environ(), "CARTULARY_PROXY_RECOVERY_TEST=create", "CARTULARY_PROXY_RECOVERY_BASE="+base)
	if output, err := owner.CombinedOutput(); err != nil {
		t.Fatalf("owner: %v %s", err, output)
	}
	data, err := os.ReadFile(filepath.Join(base, "identity.json"))
	if err != nil {
		t.Fatal(err)
	}
	identity := map[string]string{}
	if err := json.Unmarshal(data, &identity); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := objectstoreproxy.Stop(filepath.Join(identity["root"], "lease.json")); err != nil {
			t.Error("test proxy cleanup", err)
		}
	})
	caller, err := harnessadmission.Identity(os.Getpid())
	if err != nil {
		t.Fatal(err)
	}
	r, err := harnesssession.OwnRuntime(filepath.Join(base, "repo"), filepath.Join(base, "run"), identity["root"], "proxy-recovery", identity["lease"], caller, true)
	if err != nil {
		t.Fatal("new owner", err)
	}
	defer r.ReleaseHandles()
	lease, trust := filepath.Join(r.Root, "lease.json"), filepath.Join(r.Root, "root.pem")
	for _, name := range []string{"proxy-binary", "server.pem", "server.key", "credentials.log"} {
		if _, err := os.Stat(filepath.Join(r.Root, name)); !os.IsNotExist(err) {
			t.Fatal("unnecessary detail retained", name, err)
		}
	}
	for _, name := range []string{lease, trust} {
		if _, err := os.Stat(name); err != nil {
			t.Fatal("recovery dependency lost", err)
		}
	}
	records, err := r.Resources(64)
	if err != nil || len(records) != 1 {
		t.Fatal(records, err)
	}
	if err := r.StopResource(context.Background(), filepath.Join(base, "repo"), filepath.Join(base, "run"), nil, records[0]); err != nil {
		t.Fatal("new-process physical recovery", err)
	}
	if _, err := os.Stat(lease); err != nil {
		t.Fatal("physical stop removed unresolved authority before journal release", err)
	}
	if err := r.RecordResource("object_store_proxy", lease, "released"); err != nil {
		t.Fatal(err)
	}
	if err := r.Close(); err != nil {
		t.Fatal(err)
	}
}
