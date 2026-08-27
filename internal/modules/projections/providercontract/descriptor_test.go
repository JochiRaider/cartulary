package providercontract

import (
	"reflect"
	"strings"
	"testing"
)

func TestDescriptorSetOwnsDeclarativeValidation(t *testing.T) {
	tests := map[string]struct {
		mutate func([]ProviderDescriptor) []ProviderDescriptor
		want   string
	}{
		"unsupported schema version": {
			mutate: mutateFirstDescriptor(func(descriptor *ProviderDescriptor) { descriptor.SchemaVersion = "projection_provider_descriptor.v1" }),
			want:   "unsupported schema_version",
		},
		"unsupported status": {
			mutate: mutateFirstDescriptor(func(descriptor *ProviderDescriptor) { descriptor.Status = ProviderStatus("retired") }),
			want:   "unsupported status",
		},
		"missing provider id": {
			mutate: mutateFirstDescriptor(func(descriptor *ProviderDescriptor) { descriptor.ProviderID = "" }),
			want:   "empty provider_id",
		},
		"missing source owner": {
			mutate: mutateFirstDescriptor(func(descriptor *ProviderDescriptor) { descriptor.SourceOwnerModule = "" }),
			want:   "empty source_owner_module",
		},
		"missing view schemas": {
			mutate: mutateFirstDescriptor(func(descriptor *ProviderDescriptor) { descriptor.ViewSchemaIDs = nil }),
			want:   "no view_schema_ids",
		},
		"duplicate view schema": {
			mutate: mutateFirstDescriptor(func(descriptor *ProviderDescriptor) {
				descriptor.ViewSchemaIDs = []string{"cartulary.view.hosts.v1", "cartulary.view.hosts.v1"}
			}),
			want: "duplicate view_schema_ids",
		},
		"missing source record types": {
			mutate: mutateFirstDescriptor(func(descriptor *ProviderDescriptor) { descriptor.SourceRecordTypes = nil }),
			want:   "no source_record_types",
		},
		"duplicate source record type": {
			mutate: mutateFirstDescriptor(func(descriptor *ProviderDescriptor) { descriptor.SourceRecordTypes = []string{"host", "host"} }),
			want:   "duplicate source_record_types",
		},
		"missing source authority modules": {
			mutate: mutateFirstDescriptor(func(descriptor *ProviderDescriptor) { descriptor.SourceAuthorityModules = nil }),
			want:   "no source_authority_modules",
		},
		"source authority modules omit owner": {
			mutate: mutateFirstDescriptor(func(descriptor *ProviderDescriptor) { descriptor.SourceAuthorityModules = []string{"links"} }),
			want:   "omit source_owner_module",
		},
		"duplicate provider": {
			mutate: func(descriptors []ProviderDescriptor) []ProviderDescriptor {
				descriptors[1].ProviderID = descriptors[0].ProviderID
				return descriptors
			},
			want: "duplicate projection provider_id",
		},
		"duplicate view ownership": {
			mutate: func(descriptors []ProviderDescriptor) []ProviderDescriptor {
				descriptors[1].ViewSchemaIDs = descriptors[0].ViewSchemaIDs
				return descriptors
			},
			want: "duplicate projection view ownership",
		},
		"missing projection table": {
			mutate: mutateFirstDescriptor(func(descriptor *ProviderDescriptor) { descriptor.ProjectionTableIDs = nil }),
			want:   "no projection_table_ids",
		},
		"duplicate projection table ownership": {
			mutate: func(descriptors []ProviderDescriptor) []ProviderDescriptor {
				descriptors[1].ProjectionTableIDs = descriptors[0].ProjectionTableIDs
				return descriptors
			},
			want: "duplicate projection table ownership",
		},
		"storage ownership mismatch": {
			mutate: mutateFirstDescriptor(func(descriptor *ProviderDescriptor) { descriptor.ProjectionStorageOwnerModule = "entities" }),
			want:   "must be projections",
		},
		"restore without incident rebuild": {
			mutate: mutateFirstDescriptor(func(descriptor *ProviderDescriptor) { descriptor.Capabilities.IncidentRebuild = false }),
			want:   "restore rebuild without incident rebuild capability",
		},
		"required restore without capability": {
			mutate: mutateFirstDescriptor(func(descriptor *ProviderDescriptor) { descriptor.Capabilities.RestoreRebuild = false }),
			want:   "required restore rebuild without capability",
		},
		"active unsupported restore": {
			mutate: mutateFirstDescriptor(func(descriptor *ProviderDescriptor) {
				descriptor.RestoreRebuild = RestoreRebuildUnsupported
				descriptor.Capabilities.RestoreRebuild = false
			}),
			want: "active but declares unsupported restore rebuild",
		},
		"missing facade": {
			mutate: mutateFirstDescriptor(func(descriptor *ProviderDescriptor) { descriptor.FacadePackages = nil }),
			want:   "no facade_packages",
		},
		"duplicate facade": {
			mutate: mutateFirstDescriptor(func(descriptor *ProviderDescriptor) {
				descriptor.FacadePackages = []string{"internal/modules/entities", "internal/modules/entities"}
			}),
			want: "duplicate facade_packages",
		},
		"projection internal facade": {
			mutate: mutateFirstDescriptor(func(descriptor *ProviderDescriptor) {
				descriptor.FacadePackages = []string{"internal/modules/projections"}
			}),
			want: "must not expose projection internals",
		},
		"unknown dependency": {
			mutate: mutateFirstDescriptor(func(descriptor *ProviderDescriptor) { descriptor.RebuildAfter = []string{"missing"} }),
			want:   "references unknown provider",
		},
		"dependency cycle": {
			mutate: mutateFirstDescriptor(func(descriptor *ProviderDescriptor) { descriptor.RebuildAfter = []string{"identity"} }),
			want:   "cycle",
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := NewDescriptorSet(test.mutate(validDescriptorSetFacts()))
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("NewDescriptorSet error = %v, want containing %q", err, test.want)
			}
		})
	}
}

func TestDescriptorSetRebuildOrderAndCopiesAreImmutable(t *testing.T) {
	descriptors := validDescriptorSetFacts()
	set, err := NewDescriptorSet(descriptors)
	if err != nil {
		t.Fatalf("NewDescriptorSet: %v", err)
	}
	descriptors[0].ViewSchemaIDs[0] = "mutated-input"
	first := set.RebuildOrder()
	if got, want := descriptorIDs(first), []string{"host", "identity"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("rebuild order = %#v, want %#v", got, want)
	}
	first[0].ProviderID = "mutated-output"
	first[0].ViewSchemaIDs[0] = "mutated-output"
	second := set.RebuildOrder()
	if got, want := descriptorIDs(second), []string{"host", "identity"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("rebuild order after mutation = %#v, want %#v", got, want)
	}
	if second[0].ViewSchemaIDs[0] != "cartulary.view.hosts.v1" {
		t.Fatalf("rebuild descriptor slice escaped: %#v", second[0].ViewSchemaIDs)
	}
}

func validDescriptorSetFacts() []ProviderDescriptor {
	descriptor := func(providerID, viewSchemaID, recordType, tableID string, after []string) ProviderDescriptor {
		return ProviderDescriptor{
			SchemaVersion:                DescriptorSchemaVersion,
			Status:                       ProviderStatusActive,
			ProviderID:                   providerID,
			SourceOwnerModule:            "entities",
			ViewSchemaIDs:                []string{viewSchemaID},
			SourceRecordTypes:            []string{recordType},
			SourceAuthorityModules:       []string{"entities"},
			ProjectionTableIDs:           []string{tableID},
			ProjectionStorageOwnerModule: "projections",
			Capabilities: ProviderCapabilities{
				Query:           true,
				RefreshRow:      true,
				RestoreRebuild:  true,
				IncidentRebuild: true,
			},
			RestoreRebuild: RestoreRebuildRequired,
			FacadePackages: []string{"internal/modules/entities"},
			RebuildAfter:   after,
		}
	}
	return []ProviderDescriptor{
		descriptor("host", "cartulary.view.hosts.v1", "host", "host_grid_projection", nil),
		descriptor("identity", "cartulary.view.identities.v1", "identity", "identity_grid_projection", []string{"host"}),
	}
}

func mutateFirstDescriptor(mutate func(*ProviderDescriptor)) func([]ProviderDescriptor) []ProviderDescriptor {
	return func(descriptors []ProviderDescriptor) []ProviderDescriptor {
		mutate(&descriptors[0])
		return descriptors
	}
}

func descriptorIDs(descriptors []ProviderDescriptor) []string {
	ids := make([]string, 0, len(descriptors))
	for _, descriptor := range descriptors {
		ids = append(ids, descriptor.ProviderID)
	}
	return ids
}
