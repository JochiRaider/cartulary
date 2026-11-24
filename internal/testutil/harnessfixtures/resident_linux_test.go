//go:build linux

package harnessfixtures

import (
	"slices"
	"testing"
)

func TestPhysicalResidents(t *testing.T) {
	claims := map[string]int64{"memory_mb": 256}
	physical := map[string]map[string]int64{"object_store": claims, "postgres": claims, "browser_stack": claims}
	residents := func(mode string, request Request) []Resident {
		p := &Production{Options: ProductionOptions{Environment: map[string]string{"CARTULARY_TEST_SERVICES_MODE": mode}}}
		return p.physicalResidents(WorkRequirement{Request: request, Physical: physical})
	}
	keys := func(residents []Resident) []string {
		var result []string
		for _, resident := range residents {
			result = append(result, resident.Key+"/"+resident.Scope)
		}
		return result
	}
	for name, test := range map[string]struct {
		mode    string
		request Request
		want    []string
	}{
		"owned exclusive generation":                  {"owned", Request{UnitID: "go:a", PostgresIsolation: "exclusive"}, []string{"object_store:shared/", "postgres:unit:go:a/"}},
		"owned shared generation":                     {"owned", Request{UnitID: "go:a", PostgresIsolation: "shared"}, []string{"object_store:shared/", "postgres:shared/"}},
		"attached shared generation":                  {"attach", Request{UnitID: "go:a", PostgresIsolation: "shared"}, []string{"object_store:shared/attached", "postgres:shared/attached"}},
		"attached exclusive generation stays charged": {"attach", Request{UnitID: "go:a", PostgresIsolation: "exclusive"}, []string{"object_store:shared/attached", "postgres:unit:go:a/"}},
	} {
		if got := keys(residents(test.mode, test.request)); !slices.Equal(got, test.want) {
			t.Errorf("%s: %v, want %v", name, got, test.want)
		}
	}
	stack := residents("owned", Request{UnitID: "browser:a", PostgresIsolation: "exclusive", Capability: "browser_stack"})
	if len(stack) != 3 || stack[2].Key != stackKey(Request{UnitID: "browser:a", PostgresIsolation: "exclusive", Capability: "browser_stack"}) {
		t.Fatalf("browser stack residents = %v", keys(stack))
	}
}
