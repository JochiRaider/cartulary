//go:build linux

package harnesssession

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/testutil/harnessadmission"
	"github.com/JochiRaider/cartulary/internal/testutil/harnesslocal"
	"github.com/JochiRaider/cartulary/internal/testutil/harnessschema"
	"golang.org/x/sys/unix"
)

func TestSessionProtocol(t *testing.T) {
	if os.Getenv("CARTULARY_SESSION_PROTOCOL_TEST_CHILD") == "1" {
		os.Exit(Main(nil, os.Stdin, os.Stdout))
	}
	t.Run("public_cleanup_diagnostic_preserves_classification_without_private_proof", func(t *testing.T) {
		cause := errors.Join(ErrUnresolved, &os.PathError{Op: "open", Path: "/proc/123/environ", Err: unix.EACCES}, errors.New("credential-sentinel"))
		err := AtCleanupStage("detail_producer", AtCleanupStage("process_discovery", cause))
		if !errors.Is(err, ErrUnresolved) || !errors.Is(err, unix.EACCES) {
			t.Fatal("diagnostic changed the native classification")
		}
		d := DescribeCleanupFailure(err)
		if d == nil || d.String() != "[RUNTIME-CLEANUP] stage=process_discovery cause=unresolved errno=13" {
			t.Fatal("missing closed cleanup diagnostic", d)
		}
		if strings.Contains(d.String(), "123") || strings.Contains(d.String(), "environ") || strings.Contains(d.String(), "credential-sentinel") {
			t.Fatal("public cleanup diagnostic retained private proof")
		}
		if DescribeCleanupFailure(nil) != nil || AtCleanupStage("detail_remove", nil) != nil {
			t.Fatal("successful cleanup acquired failure diagnostics")
		}
		if d := DescribeCleanupFailure(AtCleanupStage("untrusted-stage-sentinel", errors.New("private-path-sentinel"))); d.String() != "[RUNTIME-CLEANUP] stage=runtime_close cause=internal errno=0" {
			t.Fatal("unrecognized cleanup metadata escaped the closed boundary")
		}
	})
	t.Run("bounded_private_rejections_exclude_arbitrary_detail", func(t *testing.T) {
		for _, tc := range []struct {
			err    error
			reason string
			proc   bool
		}{
			{harnesslocal.ErrUnsafe, "unsafe", false},
			{errors.Join(ErrUnresolved, &os.PathError{Op: "open", Path: "/proc/123/environ", Err: unix.EACCES}), "unresolved", true},
			{&os.PathError{Op: "open", Path: "/private/credential-sentinel", Err: unix.ENOENT}, "io", false},
			{context.DeadlineExceeded, "deadline", false},
			{errors.New("credential-sentinel"), "internal", false},
		} {
			r := rejectReason(tc.err)
			if r.Reason != tc.reason || (r.PID != nil) != tc.proc {
				t.Fatal("incorrect private rejection", r)
			}
			if tc.proc && (*r.PID != 123 || r.ProofFile == nil || *r.ProofFile != "environ" || r.Errno == nil || *r.Errno != 13) {
				t.Fatal("missing proof-read diagnostic", r)
			}
			data, err := json.Marshal(response{Schema: RequestSchema, ID: 1, Operation: "inspect_diagnostic_scope", Error: r})
			if err != nil || bytes.Contains(data, []byte("credential-sentinel")) {
				t.Fatal("private error leaked arbitrary detail", err)
			}
			value, err := harnessschema.Decode(data)
			if err != nil || harnessschema.Validate(RequestSchema, value) != nil {
				t.Fatal("invalid rejection frame", err)
			}
		}
	})
	t.Run("physical_failure_diagnostic_is_closed_and_preserves_live_identity", func(t *testing.T) {
		proof, err := CaptureProcess(os.Getpid())
		if err != nil {
			t.Fatal(err)
		}
		proof.Group = false
		proof.Start = "0"
		request, _ := json.Marshal(map[string]any{"proof": proof, "grace_ms": 0, "kill_ms": 1})
		var output bytes.Buffer
		if code := ProcessMain(nil, bytes.NewReader(request), &output); code != 12 {
			t.Fatal("mismatched identity accepted", code)
		}
		d := DecodeCleanupDiagnostic(output.Bytes())
		if d == nil || d.Stage != "process_identity" || d.Cause != "unresolved" {
			t.Fatal("lost native diagnostic", output.String())
		}
		if unix.Kill(os.Getpid(), 0) != nil {
			t.Fatal("rejected proof affected the live process")
		}
		for _, input := range [][]byte{
			bytes.Repeat([]byte(" "), 4097),
			[]byte(`{"schema_id":"cartulary.harness_cleanup_diagnostic.v2","stage":"process_identity","cause":"unresolved","errno":0}`),
			[]byte(`{"schema_id":"cartulary.harness_cleanup_diagnostic.v3","stage":"process_identity","cause":"unresolved","errno":0,"pid":123}`),
			[]byte(`{"schema_id":"cartulary.harness_cleanup_diagnostic.v3","stage":"process_identity","cause":"credential-sentinel","errno":0}`),
			[]byte(`{"schema_id":"cartulary.harness_cleanup_diagnostic.v3","stage":"process_identity","cause":"unresolved","errno":0,"errno":13}`),
		} {
			if DecodeCleanupDiagnostic(input) != nil {
				t.Fatal("unsafe diagnostic accepted")
			}
		}
		output.Reset()
		if code := ProcessMain(nil, strings.NewReader(`{"unknown":true}`), &output); code != 11 || output.Len() != 0 {
			t.Fatal("malformed request reached physical cleanup")
		}
	})
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
