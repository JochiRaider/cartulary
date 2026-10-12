package suiteservices

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

func TestMigrationActivityPrivatePublicationAndShutdown(t *testing.T) {
	root := t.TempDir()
	if err := os.Chmod(root, 0o700); err != nil {
		t.Fatal(err)
	}
	write := func(name, value string) {
		if err := os.WriteFile(filepath.Join(root, name), []byte(value), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	for key, value := range map[string]string{SuiteIDEnv: "", testRunIDEnv: "activity-test", testResultsDirEnv: t.TempDir(),
		SuiteRuntimeRootEnv: root, SuiteRuntimeRunIDEnv: "activity-test", SuiteRuntimeLeaseIDEnv: "00000000-0000-4000-8000-000000000001",
		"CARTULARY_HARNESS_EXECUTION_ENABLED": "1", "CARTULARY_HARNESS_ALLOCATION_REF": "",
		"CARTULARY_HARNESS_LAUNCH_CONTEXT": `{"run_id":"activity-test","invocation_id":"00000000-0000-4000-8000-000000000001","unit_id":"u"}`} {
		t.Setenv(key, value)
	}
	repository, err := FindRepoRoot()
	if err != nil {
		t.Fatal(err)
	}
	owner, _ := json.Marshal(map[string]any{"schema_id": "cartulary.harness_suite_runtime_owner.v2", "lease_id": os.Getenv(SuiteRuntimeLeaseIDEnv), "run_id": "activity-test", "owner_uid": os.Getuid(), "created_at": "2026-10-09T00:00:00Z", "repository": repository, "run_root": filepath.Join(os.Getenv(testResultsDirEnv), "activity-test"), "output_roots": []string{}})
	write("runtime-owner.json", string(owner))
	if err := os.MkdirAll(filepath.Join(root, "detail/execution-observations"), 0o700); err != nil {
		t.Fatal(err)
	}
	write("detail/execution-observations/open", "1\n")
	write("detail/execution-observations/limits.json", `{"maximum_execution_records":1,"maximum_record_bytes":65536}`)
	finish := migrationObservation()
	finish(time.Now(), 7500*time.Microsecond, "failed")
	bytes, err := os.ReadFile(filepath.Join(root, "detail/execution-observations/slot-0"))
	if err != nil {
		t.Fatal(err)
	}
	var record map[string]any
	if json.Unmarshal(bytes, &record) != nil || record["outcome"] != "failed" || record["duration_ms"] != 7.5 {
		t.Fatalf("bad activity: %s", bytes)
	}

	if len(record["anchors"].([]any)) != 2 || record["local_start_ms"] == nil || record["local_end_ms"].(float64)-record["local_start_ms"].(float64) < 7.49999 {
		t.Fatal("missing producer clock coordinates", record)
	}
	migrationObservation()(time.Now(), time.Second, "passed")
	if _, err := os.Stat(filepath.Join(root, "detail/execution-observations/truncated")); err != nil {
		t.Fatal("missing cap exhaustion marker", err)
	}
	write("detail/execution-observations/open", "0\n")
	finish(time.Now(), time.Second, "passed")
	after, _ := os.ReadFile(filepath.Join(root, "detail/execution-observations/slot-0"))
	if string(after) != string(bytes) {
		t.Fatal("late completion changed retained private fact")
	}
}

func TestMigrationActivityMeasuresOnlyOperationAndRetainsFailure(t *testing.T) {
	for _, failure := range []error{nil, errors.New("migration failed"), context.Canceled} {
		var duration time.Duration
		var outcome string
		clock := []time.Time{time.Unix(1, 0), time.Unix(1, 7300000)}
		now := func() time.Time { value := clock[0]; clock = clock[1:]; return value }
		observed := false
		got, err := measureMigration(context.Background(), func() error {
			if !observed {
				t.Fatal("operation started before incomplete publication")
			}
			return failure
		}, now, func() func(time.Time, time.Duration, string) {
			observed = true
			return func(_ time.Time, d time.Duration, status string) { duration, outcome = d, status }
		})
		want := "passed"
		if failure != nil {
			want = "failed"
		}
		if errors.Is(failure, context.Canceled) {
			want = "cancelled"
		}
		if got != 7300*time.Microsecond || duration != got || err != failure || outcome != want {
			t.Fatalf("duration=%v outcome=%s err=%v", duration, outcome, err)
		}
	}
}

func TestMigrationActivityFatalExitRetainsIncompleteBeginning(t *testing.T) {
	started, finished := false, false
	done := make(chan struct{})
	go func() {
		defer close(done)
		_, _ = measureMigration(context.Background(), func() error { runtime.Goexit(); return nil }, time.Now,
			func() func(time.Time, time.Duration, string) {
				started = true
				return func(time.Time, time.Duration, string) { finished = true }
			})
	}()
	<-done
	if !started || finished {
		t.Fatal("fatal exit fabricated a completion")
	}
}
