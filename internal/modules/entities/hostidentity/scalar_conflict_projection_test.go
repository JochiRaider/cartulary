package hostidentity

import (
	"encoding/json"
	"reflect"
	"testing"
)

func exerciseEntityConflictSnapshotFields(t *testing.T) {
	projectors, err := newEntityConflictSnapshotProjectors()
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		view, schema, prefix string
		source               map[string]any
		expected             map[string]any
	}{
		{"cartulary.view.hosts.v1", "cartulary.revisions.snapshot.host.v1", "host", map[string]any{"display_name": "Host display", "hostname": "host.example.test", "location": "Room", "os_platform": "OS", "business_owner": "Owner", "criticality": "Critical", "containment_status": nil, "fqdn": "Ignored"}, map[string]any{"host.display_name": "Host display", "host.hostname": "host.example.test", "host.location": "Room", "host.os_platform": "OS", "host.business_owner": "Owner", "host.criticality": "Critical", "host.containment_status": nil}},
		{"cartulary.view.identities.v1", "cartulary.revisions.snapshot.identity.v1", "identity", map[string]any{"display_name": "Identity display", "upn": "u@example.test", "email": "e@example.test", "sam_account_name": "account", "privilege_level": "Privileged", "mfa_state": "Enabled", "reset_status": nil, "sid": "Ignored"}, map[string]any{"identity.display_name": "Identity display", "identity.upn": "u@example.test", "identity.email": "e@example.test", "identity.sam_account_name": "account", "identity.privilege_level": "Privileged", "identity.mfa_state": "Enabled", "identity.reset_status": nil}},
	} {
		t.Run(tc.prefix, func(t *testing.T) {
			encoded, err := json.Marshal(map[string]any{"snapshot_schema_id": tc.schema, "record": map[string]any{}, "source": tc.source})
			if err != nil {
				t.Fatal(err)
			}
			row, ok := projectors[tc.view].Project(encoded)
			if !ok {
				t.Fatal("source snapshot rejected")
			}
			cells := row["cells"].(map[string]any)
			if len(cells) != len(tc.expected) {
				t.Fatalf("projected cells=%#v", cells)
			}
			for field, want := range tc.expected {
				got, ok := cells[field].(map[string]any)
				if !ok {
					t.Fatalf("missing cell %s", field)
				}
				value, present := got["value"]
				if !present || !reflect.DeepEqual(value, want) {
					t.Fatalf("cell %s=%#v, want %#v", field, got, want)
				}
			}
		})
	}
}
