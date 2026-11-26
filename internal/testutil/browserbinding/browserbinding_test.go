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
