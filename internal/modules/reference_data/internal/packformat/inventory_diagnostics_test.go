package packformat

import (
	"context"
	"reflect"
	"strings"
	"testing"
)

func TestInventoryAndBindingDiagnosticsEnumerateEveryMember_Unit(t *testing.T) {
	m := Manifest{Files: []File{{Path: "notices/license.txt", Size: 3, SHA256: strings.Repeat("a", 64)}, {Path: "payload/entries.ndjson", Size: 4, SHA256: strings.Repeat("b", 64)}}}
	inventory := Inventory{"manifest.json": {}, "notices/license.txt": {Size: 5, SHA256: strings.Repeat("c", 64)}, "payload/entries.ndjson": {Size: 6, SHA256: strings.Repeat("d", 64)}, "hostile.secret": {}, "other.secret": {}}
	for _, id := range []string{"member_extra", "member_length", "member_hash"} {
		summary, err := CheckSummary(context.Background(), id, nil, func(_ context.Context, emit FindingSink) error {
			return InventoryFindings(id, m, inventory, false, emit)
		})
		if err != nil || summary.Total != 2 {
			t.Fatalf("%s: %#v %v", id, summary, err)
		}
		other := Inventory{}
		for _, path := range []string{"other.secret", "hostile.secret", "payload/entries.ndjson", "notices/license.txt", "manifest.json"} {
			other[path] = inventory[path]
		}
		repeated, err := CheckSummary(context.Background(), id, nil, func(_ context.Context, emit FindingSink) error { return InventoryFindings(id, m, other, false, emit) })
		if err != nil || !reflect.DeepEqual(summary, repeated) {
			t.Fatal("inventory traversal changed summary")
		}
		for _, issue := range summary.Issues {
			if strings.Contains(issue.Path, "secret") {
				t.Fatal("hostile path leaked")
			}
		}
	}
	missing, err := CheckSummary(context.Background(), "member_missing", nil, func(_ context.Context, emit FindingSink) error {
		return InventoryFindings("member_missing", m, Inventory{}, false, emit)
	})
	if err != nil || missing.Total != 3 {
		t.Fatal("missing member enumeration", missing, err)
	}
	summary, err := CheckSummary(context.Background(), "target_binding", nil, func(_ context.Context, emit FindingSink) error {
		return TrustBindingFindings(m, "manifest digest", "payload digest", TrustProposal{}, emit)
	})
	if err != nil || summary.Total != 6 {
		t.Fatal("binding enumeration", summary, err)
	}
}
