package runcontext

import (
	"maps"
	"slices"
	"testing"
)

// Names are assembled so this file names none itself.
func name(suffix string) string { return "CARTULARY" + "_" + suffix }

func TestContextWritesOnlySetFields(t *testing.T) {
	got := Context{ResultsDir: "/results", RunID: "run", Target: "check"}.Environment()
	want := map[string]string{name("TEST_RESULTS_DIR"): "/results", name("TEST_RUN_ID"): "run", name("TEST_TARGET"): "check"}
	if !maps.Equal(got, want) {
		t.Fatalf("Environment() = %v, want %v", got, want)
	}
	if got := (Context{}).Environment(); len(got) != 0 {
		t.Fatalf("an empty context wrote %v", got)
	}
	environment := map[string]string{name("TEST_RUN_ID"): "parent", "PATH": "/bin"}
	Context{RunID: "child", AdmissionLease: "lease"}.Apply(environment)
	if want := map[string]string{name("TEST_RUN_ID"): "child", name("HOST_ADMISSION_LEASE"): "lease", "PATH": "/bin"}; !maps.Equal(environment, want) {
		t.Fatalf("Apply left %v, want %v", environment, want)
	}
}

func TestListsAlwaysWrite(t *testing.T) {
	if got := Services(nil)[name("HARNESS_SERVICE_DEPENDENCIES")]; got != "" {
		t.Fatalf("no services wrote %q", got)
	}
	if got := Services([]string{"postgres", "object_store"})[name("HARNESS_SERVICE_DEPENDENCIES")]; got != "postgres,object_store" {
		t.Fatalf("services wrote %q", got)
	}
	if got, ok := InputSources(nil)[name("MAKE_INPUT_SOURCES")]; !ok || got != "" {
		t.Fatalf("no input sources wrote %q, %v", got, ok)
	}
	if got := InputSources([]string{"OWNER=cli", "ROWS=cli"})[name("MAKE_INPUT_SOURCES")]; got != "OWNER=cli ROWS=cli" {
		t.Fatalf("input sources wrote %q", got)
	}
	if got := Entries(map[string]string{"B": "2", "A": "1"}); !slices.Equal(got, []string{"A=1", "B=2"}) {
		t.Fatalf("Entries = %q", got)
	}
}
