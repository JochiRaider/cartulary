//go:build linux

package harnessfixtures

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestFixtureProxyLifecycle(t *testing.T) {
	for _, variant := range []string{"promotion_failure", "missing_ready_proof"} {
		t.Run(variant, func(t *testing.T) {
			runtime, root, run := runtimeFixture(t)
			suite := &SuiteOwner{Acquire: func(context.Context) (*Suite, error) {
				return &Suite{Environment: map[string]string{"CARTULARY_S3TEST_ENDPOINT": "localhost:1", "CARTULARY_S3TEST_PROXY_CERTIFICATE_PATH": "/fixture/cert", "CARTULARY_S3TEST_PROXY_PRIVATE_KEY_PATH": "/fixture/key", "SSL_CERT_FILE": "/fixture/root"}}, nil
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
