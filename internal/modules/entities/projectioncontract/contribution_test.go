package projectioncontract

import (
	"reflect"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/entities/projectionports"
	"github.com/JochiRaider/cartulary/internal/modules/projections/providercontract"
)

type contributionSource struct{ SourceReader }

func TestNewContributionRequiresSource(t *testing.T) {
	t.Parallel()
	for _, source := range []SourceReader{nil, (*contributionSource)(nil)} {
		got, err := NewContribution(source)
		if err == nil || err.Error() != "entities projection source is required" || got.Source() != nil {
			t.Fatalf("nil source admitted: %#v, %v", got, err)
		}
	}
	for _, source := range []SourceReader{&contributionSource{}, contributionSource{}} {
		got, err := NewContribution(source)
		if err != nil || got.Source() != source {
			t.Fatalf("valid source not retained: %#v, %v", got, err)
		}
	}
}

func TestRuntimeContributionOwnsTypedEntityDescriptorFacts(t *testing.T) {
	contribution, err := NewContribution(&contributionSource{})
	if err != nil {
		t.Fatalf("construct Entities projection contribution: %v", err)
	}
	descriptors := contribution.ProjectionContribution().Descriptors()
	if len(descriptors) != 2 {
		t.Fatalf("descriptor count = %d, want 2", len(descriptors))
	}
	wantDescriptors := []providercontract.ProviderDescriptor{
		{
			SchemaVersion: "projection_provider_descriptor.v3", Status: "active", ProviderID: "host",
			SourceOwnerModule: "entities", ProjectionStorageOwnerModule: "projections",
			ViewSchemaIDs: []string{"cartulary.view.hosts.v1"}, SourceRecordTypes: []string{"host"},
			SourceAuthorityModules: []string{"entities", "evidence", "links", "records"}, ProjectionTableIDs: []string{"host_grid_projection"},
			Capabilities:   providercontract.ProviderCapabilities{Query: true, RefreshRow: true, RestoreRebuild: true, IncidentRebuild: true},
			RestoreRebuild: "required", FacadePackages: []string{"internal/modules/entities/projectioncontract"}, RebuildAfter: []string{"timeline"},
			CharacterizationRefs: []string{"internal/modules/entities/origin_upsert_integration_test.go", "internal/modules/entities/resolution_route_integration_test.go"},
		},
		{
			SchemaVersion: "projection_provider_descriptor.v3", Status: "active", ProviderID: "identity",
			SourceOwnerModule: "entities", ProjectionStorageOwnerModule: "projections",
			ViewSchemaIDs: []string{"cartulary.view.identities.v1"}, SourceRecordTypes: []string{"identity"},
			SourceAuthorityModules: []string{"entities", "evidence", "links", "records"}, ProjectionTableIDs: []string{"identity_grid_projection"},
			Capabilities:   providercontract.ProviderCapabilities{Query: true, RefreshRow: true, RestoreRebuild: true, IncidentRebuild: true},
			RestoreRebuild: "required", FacadePackages: []string{"internal/modules/entities/projectioncontract"}, RebuildAfter: []string{"host"},
			CharacterizationRefs: []string{"internal/modules/entities/origin_upsert_integration_test.go", "internal/modules/entities/resolution_route_integration_test.go"},
		},
	}
	if contribution.ProjectionContribution().SourceOwnerModule() != "entities" || !reflect.DeepEqual(Descriptors(), wantDescriptors) {
		t.Fatal("Entities declaration differs from its independent owner expectation")
	}
	for index, want := range wantDescriptors {
		got := descriptors[index]
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("descriptor %d drifted:\ngot  %#v\nwant %#v", index, got, want)
		}
	}
	intents := contribution.ProjectionContribution().SurfaceIntents()
	wantIntents := []providercontract.SurfaceIntent{
		{ViewSchemaID: "cartulary.view.hosts.v1", FieldKeys: strings.Fields(`
host.display_name host.hostname host.aad_device_id host.fqdn host.aliases
host.reusable_identifiers host.host_state host.linked_event_count host.evidence_count
host.location host.os_platform host.business_owner host.criticality host.containment_status
host.edited_at host.subject_party_ids host.responsible_party_ids host.observer_party_ids`)},
		{ViewSchemaID: "cartulary.view.identities.v1", FieldKeys: strings.Fields(`
identity.display_name identity.aad_object_id identity.sid identity.upn identity.email
identity.sam_account_name identity.aliases identity.reusable_identifiers identity.identity_state
identity.linked_event_count identity.evidence_count identity.privilege_level identity.mfa_state
identity.reset_status identity.edited_at identity.subject_party_ids identity.responsible_party_ids identity.observer_party_ids`)},
	}
	if !reflect.DeepEqual(intents, wantIntents) {
		t.Fatalf("Entities surface intents drifted: %#v", intents)
	}

	assertInterfaceMethods(t, (*projectionports.MutationRows)(nil), []string{
		"DeleteHostTx",
		"DeleteIdentityTx",
		"RefreshHostTx",
		"RefreshIdentityTx",
	})
	assertInterfaceMethods(t, (*projectionports.QueryReader)(nil), []string{
		"SelectHostQueryProjections",
		"SelectHostQueryProjectionsTx",
		"SelectIdentityQueryProjections",
		"SelectIdentityQueryProjectionsTx",
	})
	assertInterfaceMethods(t, (*projectionports.ReportingReader)(nil), []string{
		"CollectHostDerivedFactsTx",
		"CollectIdentityDerivedFactsTx",
	})
}

func assertInterfaceMethods(t *testing.T, pointer any, want []string) {
	t.Helper()
	interfaceType := reflect.TypeOf(pointer).Elem()
	got := make([]string, interfaceType.NumMethod())
	for index := range interfaceType.NumMethod() {
		got[index] = interfaceType.Method(index).Name
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("%s methods = %v, want %v", interfaceType, got, want)
	}
}
