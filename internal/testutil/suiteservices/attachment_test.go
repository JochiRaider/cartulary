package suiteservices

import (
	"maps"
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

func TestPostgresAllocationAlwaysWritesBoth(t *testing.T) {
	if got, want := (PostgresAllocation{}).Environment(), map[string]string{PGIsolationEnv: "", PGAllocationRefEnv: ""}; !maps.Equal(got, want) {
		t.Fatalf("empty allocation wrote %v, want %v", got, want)
	}
	if got, want := (PostgresAllocation{Isolation: "shared", Ref: "session:a"}).Environment(), map[string]string{PGIsolationEnv: "shared", PGAllocationRefEnv: "session:a"}; !maps.Equal(got, want) {
		t.Fatalf("allocation wrote %v, want %v", got, want)
	}
}
