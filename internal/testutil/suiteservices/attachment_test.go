package suiteservices

import (
	"encoding/json"
	"errors"
	"maps"
	"os"
	"path/filepath"
	"testing"
)

func TestAttachmentEnvironmentOmitsUnsetModes(t *testing.T) {
	for _, tc := range []struct {
		attachment Attachment
		want       map[string]string
	}{
		{Attachment{SuiteID: "suite"}, map[string]string{SuiteIDEnv: "suite"}},
		{Attachment{SuiteID: "suite", CallMode: "owned", LifecycleMode: "owned"}, map[string]string{SuiteIDEnv: "suite", CallModeEnv: "owned", LifecycleModeEnv: "owned"}},
		{Attachment{SuiteID: "suite", CallMode: "attach", PersistentBorrower: true}, map[string]string{SuiteIDEnv: "suite", CallModeEnv: "attach", PersistentBorrowerEnv: "1"}},
	} {
		if got := tc.attachment.Environment(); !maps.Equal(got, tc.want) {
			t.Errorf("%+v.Environment() = %v, want %v", tc.attachment, got, tc.want)
		}
	}
}

func TestPostgresAllocationAlwaysWritesEveryName(t *testing.T) {
	if got, want := (PostgresAllocation{}).Environment(), map[string]string{PGIsolationEnv: "", PGAllocationRefEnv: "", PGDataBytesEnv: ""}; !maps.Equal(got, want) {
		t.Fatalf("empty allocation wrote %v, want %v", got, want)
	}
	if got, want := (PostgresAllocation{Isolation: "shared", Ref: "session:a"}).Environment(), map[string]string{PGIsolationEnv: "shared", PGAllocationRefEnv: "session:a", PGDataBytesEnv: ""}; !maps.Equal(got, want) {
		t.Fatalf("allocation wrote %v, want %v", got, want)
	}
	allocation := PostgresAllocation{Isolation: "exclusive", Ref: "postgres:unit:a", DataBytes: 512 << 20}
	if got, want := allocation.Environment(), map[string]string{PGIsolationEnv: "exclusive", PGAllocationRefEnv: "postgres:unit:a", PGDataBytesEnv: "536870912"}; !maps.Equal(got, want) {
		t.Fatalf("sized allocation wrote %v, want %v", got, want)
	}
	if got, err := PostgresDataBytes(allocation.Environment()); err != nil || got != 512<<20 {
		t.Fatalf("read back %d, %v", got, err)
	}
	if got, err := PostgresDataBytes(map[string]string{PGDataBytesEnv: ""}); err != nil || got != 0 {
		t.Fatalf("an unnamed capacity read %d, %v", got, err)
	}
	for _, invalid := range []string{"0", "-1", "1g", " 5"} {
		if _, err := PostgresDataBytes(map[string]string{PGDataBytesEnv: invalid}); err == nil {
			t.Fatalf("capacity %q accepted", invalid)
		}
	}
}

// Only a broker allocation records a footprint, in its suite artifact root.
func TestRecordPostgresFootprint(t *testing.T) {
	results := t.TempDir()
	env := map[string]string{testResultsDirEnv: results, testRunIDEnv: "footprint-run", SuiteIDEnv: "suite-a", PGAllocationRefEnv: "session:a"}
	file := filepath.Join(results, "footprint-run", "_shared", "test-services", "suite-a", PostgresFootprintFile)
	if err := RecordPostgresFootprint(env, 1<<30, 700<<20); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(file); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("a session allocation recorded a footprint: %v", err)
	}
	env[PGAllocationRefEnv] = "postgres:unit:go:a"
	if err := RecordPostgresFootprint(env, 1<<30, 700<<20); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	var footprint PostgresFootprint
	if err := json.Unmarshal(data, &footprint); err != nil {
		t.Fatal(err)
	}
	if footprint != (PostgresFootprint{SchemaID: PostgresFootprintSchema, AllocationRef: "postgres:unit:go:a", DataCapacityBytes: 1 << 30, MemoryPeakBytes: 700 << 20}) {
		t.Fatalf("footprint %+v", footprint)
	}
}
