package httptestx

import (
	"bytes"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/JochiRaider/cartulary/internal/platform/harnessruntime/runevidence"
	"github.com/JochiRaider/cartulary/internal/platform/processlifecycle"
	"github.com/JochiRaider/cartulary/internal/testutil/harnessschema"
)

// privateEvidenceRun is a run root of its own, so a deliberately forced fatal
// never fails the enclosing harness run.
func privateEvidenceRun(t *testing.T) (func(string) (string, bool), string) {
	t.Helper()
	results := t.TempDir()
	if err := os.Chmod(results, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(results, "run"), 0o700); err != nil {
		t.Fatal(err)
	}
	env := map[string]string{"CARTULARY_TEST_RESULTS_DIR": results, "CARTULARY_TEST_RUN_ID": "run", "CARTULARY_WORK_UNIT_ID": "go:harness.browser:fatal"}
	lookup := func(key string) (string, bool) {
		value, ok := env[key]
		return value, ok
	}
	return lookup, filepath.Join(results, "run", runevidence.Directory, strconv.Itoa(os.Getpid())+".jsonl")
}

// fatalRecords returns the stream's schema-valid records and fails on any
// other line.
func fatalRecords(t *testing.T, stream string) []map[string]any {
	t.Helper()
	data, err := os.ReadFile(stream)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		t.Fatal(err)
	}
	records := []map[string]any{}
	for _, line := range bytes.Split(bytes.TrimSuffix(data, []byte("\n")), []byte("\n")) {
		value, err := harnessschema.Decode(line)
		if err != nil || harnessschema.Validate(runevidence.SchemaID, value) != nil {
			t.Fatalf("record is not %s: %s", runevidence.SchemaID, line)
		}
		records = append(records, value.(map[string]any))
	}
	return records
}

// TestServerFatalObserverWritesOneRecord covers TH-HARNESS-REQ-863 and
// AC-158: an in-process server's lifecycle fatal writes exactly one
// server_fatal record with its reason code, process kind, work unit and
// process ID; a server without a fatal writes nothing.
func TestServerFatalObserverWritesOneRecord(t *testing.T) {
	t.Run("a forced fatal writes one record", func(t *testing.T) {
		lookup, stream := privateEvidenceRun(t)
		lifecycle := processlifecycle.New()
		if err := lifecycle.Publish(); err != nil {
			t.Fatal(err)
		}
		stop := watchServerFatal(lifecycle, lookup)
		if !lifecycle.Fatal("published_component_lost") || lifecycle.Fatal("indeterminate_database_commit") {
			t.Fatal("lifecycle did not take exactly one fatal")
		}
		stop()
		stop()
		records := fatalRecords(t, stream)
		if len(records) != 1 {
			t.Fatalf("records = %v, want one", records)
		}
		record := records[0]
		if record["kind"] != "server_fatal" || record["reason_code"] != "published_component_lost" || record["process_kind"] != "in_process" ||
			record["work_unit_id"] != "go:harness.browser:fatal" || record["pid"] != float64(os.Getpid()) {
			t.Fatalf("record = %v", record)
		}
	})
	t.Run("a fatal before stop is still recorded once", func(t *testing.T) {
		for range 50 {
			lookup, stream := privateEvidenceRun(t)
			lifecycle := processlifecycle.New()
			stop := watchServerFatal(lifecycle, lookup)
			lifecycle.Fatal("application_process_lease_lost")
			stop()
			if records := fatalRecords(t, stream); len(records) != 1 || records[0]["reason_code"] != "application_process_lease_lost" {
				t.Fatalf("records = %v, want one", records)
			}
		}
	})
	t.Run("a server without a fatal writes nothing", func(t *testing.T) {
		lookup, stream := privateEvidenceRun(t)
		lifecycle := processlifecycle.New()
		if err := lifecycle.Publish(); err != nil {
			t.Fatal(err)
		}
		stop := watchServerFatal(lifecycle, lookup)
		stop()
		lifecycle.Fatal("published_component_lost")
		if records := fatalRecords(t, stream); len(records) != 0 {
			t.Fatalf("records = %v, want none", records)
		}
	})
}
