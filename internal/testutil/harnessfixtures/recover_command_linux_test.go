//go:build linux

package harnessfixtures

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/testutil/harnesslocal"
	"github.com/JochiRaider/cartulary/internal/testutil/harnesssession"
)

func testRecoveryCommand(t *testing.T) {
	t.Run("failure_reports_only_bounded_diagnostic", func(t *testing.T) {
		root, run, scratch := t.TempDir(), filepath.Join(t.TempDir(), "fixture"), t.TempDir()
		if err := os.Mkdir(run, 0700); err != nil {
			t.Fatal(err)
		}
		r, err := harnesssession.NewRuntime(root, run, "fixture", scratch)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() {
			if err := r.Close(); err != nil {
				t.Error(err)
			}
		})
		proof, err := r.RecoveryPath("resident-allocations", "invalid.json")
		if err != nil {
			t.Fatal(err)
		}
		if err := os.MkdirAll(filepath.Dir(proof), 0700); err != nil {
			t.Fatal(err)
		}
		const privateDetail = "private-proof-payload-must-not-be-reported"
		if err := os.WriteFile(proof, []byte(privateDetail), 0600); err != nil {
			t.Fatal(err)
		}
		if err := r.RecordDependentResource("resident_allocation", proof, "acquired", []string{}, proof); err != nil {
			t.Fatal(err)
		}
		// The fixture is synthetic; no physical resource was acquired.
		defer func() {
			if err := r.RecordResource("resident_allocation", proof, "released"); err != nil {
				t.Error(err)
			}
		}()
		var output bytes.Buffer
		code := RecoverRun(t.Context(), root, map[string]string{"RESULTS_DIR": run, "RUNTIME_ROOT": r.Root, "RUNTIME_LEASE_ID": r.LeaseID}, &output)
		if code != 12 || !strings.Contains(output.String(), "[RUNTIME-CLEANUP]") ||
			strings.Contains(output.String(), privateDetail) || strings.Contains(output.String(), proof) {
			t.Fatal("missing or unsafe cleanup diagnostic", code, output.String())
		}
		if data, err := os.ReadFile(proof); err != nil || string(data) != privateDetail {
			t.Fatal("unresolved proof changed", err)
		}
	})
	t.Run("exact_recovery_command", func(t *testing.T) {
		root, run, scratch := t.TempDir(), filepath.Join(t.TempDir(), "fixture"), t.TempDir()
		if err := os.Mkdir(run, 0700); err != nil {
			t.Fatal(err)
		}
		r, err := harnesssession.NewRuntime(root, run, "fixture", scratch)
		if err != nil {
			t.Fatal(err)
		}
		defer r.ReleaseHandles()
		original := filepath.Join(run, "original-result.json")
		if err := os.WriteFile(original, []byte("failed\n"), 0600); err != nil {
			t.Fatal(err)
		}
		input := map[string]string{"RESULTS_DIR": run, "RUNTIME_ROOT": r.Root, "RUNTIME_LEASE_ID": r.LeaseID}
		var output bytes.Buffer
		for _, key := range []string{"RESULTS_DIR", "RUNTIME_ROOT", "RUNTIME_LEASE_ID"} {
			valid := input[key]
			input[key] = ""
			if code := RecoverRun(context.Background(), root, input, &output); code != 2 {
				t.Fatal(key, code)
			}
			input[key] = valid
		}
		input["RESULTS_DIR"] = filepath.Join(filepath.Dir(run), "foreign")
		if code := RecoverRun(context.Background(), root, input, &output); code != 2 {
			t.Fatal("foreign run admitted", code)
		}
		input["RESULTS_DIR"] = run
		input["RUNTIME_LEASE_ID"] = "00000000-0000-4000-8000-000000000000"
		if code := RecoverRun(context.Background(), root, input, &output); code != 2 {
			t.Fatal("foreign lease admitted", code)
		}
		input["RUNTIME_LEASE_ID"] = r.LeaseID
		link := filepath.Join(scratch, "link")
		if err := os.Symlink(r.Root, link); err != nil {
			t.Fatal(err)
		}
		input["RUNTIME_ROOT"] = link
		if code := RecoverRun(context.Background(), root, input, &output); code != 2 {
			t.Fatal("symlink admitted", code)
		}
		input["RUNTIME_ROOT"] = ""
		input["RUNTIME_LEASE_ID"] = ""
		input["CARTULARY_HARNESS_SCRATCH_ROOT"] = scratch
		output.Reset()
		if code := RecoveryStatus(root, input, &output); code != 0 || !strings.Contains(output.String(), "runtime_bound") || strings.Contains(output.String(), r.Root) {
			t.Fatal("unsafe or missing status", code, output.String())
		}
		if code := RecoverRun(context.Background(), root, input, &output); code != 0 {
			t.Fatal("empty exact recovery", code, output.String())
		}
		if _, err := os.Stat(r.Root); !os.IsNotExist(err) {
			t.Fatal("runtime retained", err)
		}
		output.Reset()
		if code := RecoverRun(context.Background(), root, input, &output); code != 0 {
			t.Fatal("repeat recovery was not idempotent", code, output.String())
		}
		output.Reset()
		if code := RecoveryStatus(root, input, &output); code != 0 || !strings.Contains(output.String(), "settled") {
			t.Fatal("settlement receipt missing", code, output.String())
		}
		data, err := os.ReadFile(original)
		if err != nil || string(data) != "failed\n" {
			t.Fatal("original evidence mutated", err)
		}
	})
	t.Run("startup_recovers_dead_runtimes_of_the_repository", func(t *testing.T) {
		root, scratch := t.TempDir(), t.TempDir()
		open := func(id string) (*harnesssession.Runtime, string) {
			run := filepath.Join(t.TempDir(), id)
			if err := os.Mkdir(run, 0700); err != nil {
				t.Fatal(err)
			}
			r, err := harnesssession.NewRuntime(root, run, id, scratch)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = r.ReleaseHandles() })
			return r, run
		}
		// The original owner of a stopped run is proven dead by its boot identity.
		stop := func(r *harnesssession.Runtime) []byte {
			file := filepath.Join(r.Root, "runtime-process.json")
			original, err := os.ReadFile(file)
			if err != nil {
				t.Fatal(err)
			}
			data, _ := json.Marshal(map[string]any{"boot": "retired-boot", "pid": os.Getpid(), "start": "1"})
			if err := harnesslocal.ReplaceFile(file, data); err != nil {
				t.Fatal(err)
			}
			return original
		}
		stopped, _ := open("stopped")
		stop(stopped)
		live, _ := open("live")
		failing, failingRun := open("failing")
		proof, err := failing.RecoveryPath("resident-allocations", "invalid.json")
		if err != nil {
			t.Fatal(err)
		}
		if err := os.MkdirAll(filepath.Dir(proof), 0700); err != nil {
			t.Fatal(err)
		}
		const privateDetail = "private-proof-payload-must-not-be-reported"
		if err := os.WriteFile(proof, []byte(privateDetail), 0600); err != nil {
			t.Fatal(err)
		}
		if err := failing.RecordDependentResource("resident_allocation", proof, "acquired", []string{}, proof); err != nil {
			t.Fatal(err)
		}
		original := stop(failing)
		var output bytes.Buffer
		RecoverDeadRuntimes(t.Context(), root, map[string]string{"CARTULARY_HARNESS_SCRATCH_ROOT": scratch}, &output)
		report := output.String()
		if !strings.Contains(report, "[RECOVERED] run_id=stopped resources=0\n") ||
			!strings.Contains(report, "[RECOVERY-FAIL] run_id=failing failure_class=harness failure_reason=cleanup_error recover with: make test-run-recover RESULTS_DIR="+failingRun+"\n") ||
			strings.Contains(report, "run_id=live") || strings.Contains(report, privateDetail) || strings.Contains(report, proof) {
			t.Fatal("startup recovery report", report)
		}
		if _, err := os.Stat(stopped.Root); !os.IsNotExist(err) {
			t.Fatal("dead runtime retained", err)
		}
		if _, err := os.Stat(live.Root); err != nil {
			t.Fatal("live runtime recovered", err)
		}
		if data, err := os.ReadFile(proof); err != nil || string(data) != privateDetail {
			t.Fatal("unresolved proof changed", err)
		}
		// The synthetic proof acquired nothing; restore its owner to settle it.
		if err := harnesslocal.ReplaceFile(filepath.Join(failing.Root, "runtime-process.json"), original); err != nil {
			t.Fatal(err)
		}
		if err := failing.RecordResource("resident_allocation", proof, "released"); err != nil {
			t.Fatal(err)
		}
		for _, r := range []*harnesssession.Runtime{live, failing} {
			if err := r.Close(); err != nil {
				t.Error(err)
			}
		}
	})
}
