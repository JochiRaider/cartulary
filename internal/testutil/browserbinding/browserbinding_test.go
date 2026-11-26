package browserbinding

import (
	"maps"
	"testing"
)

// Names are assembled so this file names none itself.
func name(suffix string) string { return "CARTULARY" + "_" + suffix }

func TestBindingWritesOnlySetFields(t *testing.T) {
	got := Binding{Stage: "default", RuntimeProfile: "network_flow_claimed", SessionLeaseFile: "/lease"}.Environment()
	want := map[string]string{name("BROWSER_STAGE"): "default", name("BROWSER_RUNTIME_PROFILE_ID"): "network_flow_claimed", name("WEB_E2E_SESSION_LEASE_FILE"): "/lease"}
	if !maps.Equal(got, want) {
		t.Fatalf("Environment() = %v, want %v", got, want)
	}
	environment := map[string]string{name("BROWSER_STAGE"): "parent", "PATH": "/bin"}
	Binding{Stage: "child"}.Apply(environment)
	if want := map[string]string{name("BROWSER_STAGE"): "child", "PATH": "/bin"}; !maps.Equal(environment, want) {
		t.Fatalf("Apply left %v, want %v", environment, want)
	}
	if got := (Binding{}).Environment(); len(got) != 0 {
		t.Fatalf("an empty binding wrote %v", got)
	}
}

// A performance-fixture cohort's snapshot clone travels in the binding; the
// first clone's ordinal is "0", which is a set value.
func TestBindingCarriesTheFixtureSnapshot(t *testing.T) {
	got := Binding{FixtureProfile: "timeline", SnapshotKey: "key", SnapshotBuilderUnit: "builder", FixtureRow: "row", FixturePredicate: "p95", CloneLease: "lease", CloneOrdinal: "0"}.Environment()
	want := map[string]string{
		name("FIXTURE_PROFILE_ID"): "timeline", name("FIXTURE_SNAPSHOT_KEY"): "key", name("FIXTURE_SNAPSHOT_BUILDER_UNIT_ID"): "builder",
		name("FIXTURE_ROW_ID"): "row", name("FIXTURE_PREDICATE_ID"): "p95", name("FIXTURE_CLONE_LEASE_ID"): "lease", name("FIXTURE_CLONE_ORDINAL"): "0",
	}
	if !maps.Equal(got, want) {
		t.Fatalf("Environment() = %v, want %v", got, want)
	}
}
