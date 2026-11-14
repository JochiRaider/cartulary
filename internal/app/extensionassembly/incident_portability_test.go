package extensionassembly

import (
	"reflect"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/extensions"
	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type portabilityScopeQuery struct{ pgx.Tx }

func checkIncidentPortabilityPolicies(t *testing.T) {
	catalog := []extensions.PortabilityPolicy{{
		ProfileID: "profile", Claimable: true, ContractMajor: 4, Mode: "participant",
		ParticipantID: "profile.portability", ParticipantSHA256: "digest", ParticipantSchemaID: "specialization",
		Export:                 extensions.PortabilityOperationPolicy{OutputSchemaID: "profile.members", StateFamilyIDs: []string{"profile.export"}, MaximumInputBytes: 1024, MaximumOutputBytes: 2048, MaximumItems: 3},
		Import:                 extensions.PortabilityOperationPolicy{OutputSchemaID: "profile.prepared", StateFamilyIDs: []string{"profile.import"}, MaximumInputBytes: 4096, MaximumOutputBytes: 8192, MaximumItems: 7},
		AuthoritativeFamilyIDs: []string{"profile.export", "profile.import"},
	}}
	want := []incidentbundles.ExtensionPolicy{{
		ProfileID: "profile", ClaimState: "unclaimed", ContractMajor: 4, Mode: "participant",
		ParticipantID: "profile.portability", ParticipantSHA256: "digest", ParticipantSchemaID: "specialization",
		Export:                 incidentbundles.PortabilityOperationPolicy{OutputSchemaID: "profile.members", StateFamilyIDs: []string{"profile.export"}, MaximumInputBytes: 1024, MaximumOutputBytes: 2048, MaximumItems: 3},
		Import:                 incidentbundles.PortabilityOperationPolicy{OutputSchemaID: "profile.prepared", StateFamilyIDs: []string{"profile.import"}, MaximumInputBytes: 4096, MaximumOutputBytes: 8192, MaximumItems: 7},
		AuthoritativeFamilyIDs: []string{"profile.export", "profile.import"},
	}}
	got := IncidentPortabilityPolicies(catalog, extensions.ResolvedClaimSet{})
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("composed policy = %#v; want %#v", got, want)
	}
	catalog[0].Export.StateFamilyIDs[0] = "mutated export"
	catalog[0].Import.StateFamilyIDs[0] = "mutated import"
	catalog[0].AuthoritativeFamilyIDs[0] = "mutated family"
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("composed policy retained mutable catalog slices: %#v", got)
	}
	incident := uuid.MustParse("11111111-1111-4111-8111-111111111111")
	families := []string{"network_flow_activity.graph_views", "network_flow_activity.indicator_bindings", "network_flow_activity.rejected_row_diagnostics", "network_flow_activity.rows", "network_flow_activity.tables"}
	for _, tc := range []struct {
		name     string
		families []string
		valid    bool
	}{
		{"exact export selection", families, true},
		{"missing family", families[1:], false},
		{"foreign family", []string{"another_profile.rows"}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			access, err := (IncidentPortabilityReads{}).ForSnapshot(&portabilityScopeQuery{}, incident, "network_flow_activity", tc.families)
			if tc.valid {
				if err != nil || access == nil || access.ProfileScope() != "network_flow_activity" || access.IncidentScope() != incident {
					t.Fatalf("export capability: %v", err)
				}
			} else if err == nil || access != nil {
				t.Fatalf("out-of-scope family selection admitted: access=%v err=%v", access, err)
			}
		})
	}
}
