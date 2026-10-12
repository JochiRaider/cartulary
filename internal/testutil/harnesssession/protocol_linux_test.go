//go:build linux

package harnesssession

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/testutil/harnessadmission"
)

func TestSessionProtocol(t *testing.T) {
	if os.Getenv("CARTULARY_SESSION_PROTOCOL_TEST_CHILD") == "1" {
		os.Exit(Main(nil, os.Stdin, os.Stdout))
	}
	t.Run("real_private_pipe_and_owner_binding", func(t *testing.T) {
		repo, run, scratch := t.TempDir(), t.TempDir(), t.TempDir()
		q := map[string]any{"schema_id": RequestSchema, "request_id": 1, "operation": "create", "repository": repo, "run_root": run, "run_id": "fixture", "scratch": scratch}
		call := func() map[string]any {
			t.Helper()
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestSessionProtocol$")
			cmd.Dir = repo
			cmd.Env = append(os.Environ(), "CARTULARY_SESSION_PROTOCOL_TEST_CHILD=1")
			data, _ := json.Marshal(q)
			cmd.Stdin = bytes.NewReader(data)
			out, err := cmd.Output()
			if err != nil {
				t.Fatal(q["operation"], err)
			}
			var reply map[string]any
			if json.Unmarshal(out, &reply) != nil || reply["ok"] != true || reply["operation"] != q["operation"] {
				t.Fatal("invalid response")
			}
			return reply
		}
		identity := call()["result"].(map[string]any)
		delete(q, "scratch")
		q["root"], q["lease_id"], q["operation"] = identity["root"], identity["lease_id"], "close"
		call()
		if _, err := os.Stat(identity["root"].(string)); !os.IsNotExist(err) {
			t.Fatal("root not removed", err)
		}
	})
	t.Run("strict_rejection_before_creation", func(t *testing.T) {
		repo, run, scratch := t.TempDir(), t.TempDir(), t.TempDir()
		caller, err := harnessadmission.Identity(os.Getpid())
		if err != nil {
			t.Fatal(err)
		}
		q := map[string]any{"schema_id": RequestSchema, "request_id": 1, "operation": "create", "repository": repo, "run_root": run, "run_id": "fixture", "scratch": scratch}
		valid, _ := json.Marshal(q)
		for _, data := range [][]byte{[]byte(`{"schema_id":"unknown"}`), append([]byte(`{"operation":"create",`), valid[1:]...), []byte(strings.Repeat(" ", MaximumFrame+1))} {
			if _, err := execute(data, repo, caller); err == nil {
				t.Fatal("invalid request accepted")
			}
		}
		q["command"] = "anything"
		data, _ := json.Marshal(q)
		if _, err := execute(data, repo, caller); err == nil {
			t.Fatal("arbitrary command accepted")
		}
		if _, err := os.Stat(filepath.Join(scratch, "suite-runtime")); !os.IsNotExist(err) {
			t.Fatal("rejected input created state", err)
		}
	})
	t.Run("short_lived_owner_and_borrower", func(t *testing.T) {
		repo, run, scratch := t.TempDir(), t.TempDir(), t.TempDir()
		if err := os.Chmod(run, 0700); err != nil {
			t.Fatal(err)
		}
		caller, err := harnessadmission.Identity(os.Getpid())
		if err != nil {
			t.Fatal(err)
		}
		q := map[string]any{"schema_id": RequestSchema, "request_id": 1, "operation": "create", "repository": repo, "run_root": run, "run_id": "fixture", "scratch": scratch}
		call := func() response {
			t.Helper()
			data, _ := json.Marshal(q)
			out, err := execute(data, repo, caller)
			if err != nil {
				t.Fatal(q["operation"], err)
			}
			return out
		}
		identity := call().Result.(map[string]string)
		delete(q, "scratch")
		q["root"] = identity["root"]
		q["lease_id"] = identity["lease_id"]
		q["operation"] = "register"
		q["values"] = []string{"private-sentinel-12345"}
		q["environment"] = map[string]string{"PASSWORD": "second-private-value", "CPU_TOKENS": "123456789"}
		call()
		delete(q, "values")
		delete(q, "environment")
		native, err := OwnRuntime(repo, run, identity["root"], "fixture", identity["lease_id"], caller, false)
		if err != nil {
			t.Fatal(err)
		}
		values, err := native.ForbiddenValues()
		if err != nil || len(values) != 2 {
			t.Fatal("native secret registration", err)
		}
		secrets, err := native.BeginFinalization()
		if err != nil {
			t.Fatal(err)
		}
		if _, err = ScanRetained(run, secrets); err != nil {
			t.Fatal(err)
		}
		native.ReleaseHandles()
		for _, operation := range []string{"secrets", "scan", "preserve", "prune", "resources"} {
			q["operation"] = operation
			data, _ := json.Marshal(q)
			if _, err := execute(data, repo, caller); err == nil {
				t.Fatal("removed operation accepted", operation)
			}
		}
		q["operation"] = "close"
		data, _ := json.Marshal(q)
		foreign := caller
		foreign.PID++
		if _, err := execute(data, repo, foreign); err == nil {
			t.Fatal("borrowed proof granted deletion")
		}
		call()
		if _, err := os.Stat(identity["root"]); !os.IsNotExist(err) {
			t.Fatal("runtime survived close", err)
		}
	})
	t.Run("independent_handles_serialize_with_close", func(t *testing.T) {
		r, repo, run := runtimeFixture(t)
		b, err := BorrowRuntime(repo, run, r.Root, r.RunID, r.LeaseID)
		if err != nil {
			t.Fatal(err)
		}
		defer b.ReleaseHandles()
		var group sync.WaitGroup
		for i := 0; i < 8; i++ {
			group.Go(func() {
				independent, err := BorrowRuntime(repo, run, r.Root, r.RunID, r.LeaseID)
				if err != nil {
					t.Error(err)
					return
				}
				defer independent.ReleaseHandles()
				for j := 0; j < 16; j++ {
					if err := independent.RegisterSecret("one-shared-secret"); err != nil {
						t.Error(err)
						return
					}
				}
			})
		}
		group.Wait()
		values, err := r.ForbiddenValues()
		if err != nil || len(values) != 1 {
			t.Fatal(values, err)
		}
		if _, err := r.BeginFinalization(); err != nil {
			t.Fatal(err)
		}
		if err := b.RegisterSecret("sealed-registration"); err == nil {
			t.Fatal("sealed registration accepted")
		}
		if err := r.Close(); err != nil {
			t.Fatal(err)
		}
		if err := b.RegisterSecret("must-never-be-written"); err == nil {
			t.Fatal("late registration accepted")
		}
	})
}
