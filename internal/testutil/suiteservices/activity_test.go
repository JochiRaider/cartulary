package suiteservices

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
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
		SuiteRuntimeRootEnv: root, SuiteRuntimeRunIDEnv: "activity-test", SuiteRuntimeLeaseIDEnv: "lease-test",
		"CARTULARY_HARNESS_EXECUTION_ENABLED": "1", "CARTULARY_HARNESS_ALLOCATION_REF": "",
		"CARTULARY_HARNESS_LAUNCH_CONTEXT": `{"run_id":"activity-test","invocation_id":"00000000-0000-4000-8000-000000000001","unit_id":"u"}`} {
		t.Setenv(key, value)
	}
	write("runtime-owner.json", fmt.Sprintf(`{"schema_id":"cartulary.harness_suite_runtime_owner.v1","lease_id":"lease-test","run_id":"activity-test","owner_uid":%d,"created_at":"2026-10-09T00:00:00Z"}`, os.Getuid()))
	if err := os.Mkdir(filepath.Join(root, "execution-observations"), 0o700); err != nil {
		t.Fatal(err)
	}
	write("execution-observations/open", "1\n")
	write("execution-observations/limits.json", `{"maximum_execution_records":1,"maximum_record_bytes":65536}`)
	finish := migrationObservation()
	finish(7500*time.Microsecond, "failed")
	bytes, err := os.ReadFile(filepath.Join(root, "execution-observations/slot-0"))
	if err != nil {
		t.Fatal(err)
	}
	var record map[string]any
	if json.Unmarshal(bytes, &record) != nil || record["outcome"] != "failed" || record["duration_ms"] != 7.5 {
		t.Fatalf("bad activity: %s", bytes)
	}
	migrationObservation()(time.Second, "passed")
	if _, err := os.Stat(filepath.Join(root, "execution-observations/truncated")); err != nil {
		t.Fatal("missing cap exhaustion marker", err)
	}
	write("execution-observations/open", "0\n")
	finish(time.Second, "passed")
	after, _ := os.ReadFile(filepath.Join(root, "execution-observations/slot-0"))
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
		}, now, func() func(time.Duration, string) {
			observed = true
			return func(d time.Duration, status string) { duration, outcome = d, status }
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
			func() func(time.Duration, string) {
				started = true
				return func(time.Duration, string) { finished = true }
			})
	}()
	<-done
	if !started || finished {
		t.Fatal("fatal exit fabricated a completion")
	}
}
