package runtime

import (
	"reflect"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/projections/providercontract"
)

func TestProjectionProviderRegistryOrdersAndIndexesContributions(t *testing.T) {
	registry, err := newProviderRegistry(registryValidationProviders())
	if err != nil {
		t.Fatalf("provider registry: %v", err)
	}
	if got, want := providerKeys(registry.rebuildOrder), []string{"host", "identity"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("rebuild order: got %#v want %#v", got, want)
	}
	for _, viewSchemaID := range []string{hostsViewSchemaID, identitiesViewSchemaID} {
		if provider, ok := registry.providerForView(viewSchemaID); !ok || provider == nil {
			t.Fatalf("projection surface %s has no provider", viewSchemaID)
		}
	}
}

func TestProjectionProviderRegistryRejectsInvalidContributions(t *testing.T) {
	tests := map[string]struct {
		mutate func([]Provider) []Provider
		want   string
	}{
		"unsupported schema version": {
			mutate: func(providers []Provider) []Provider {
				providers[0].descriptor.SchemaVersion = "projection_provider_descriptor.v1"
				return providers
			},
			want: "unsupported schema_version",
		},
		"unsupported status": {
			mutate: func(providers []Provider) []Provider {
				providers[0].descriptor.Status = ProviderStatus("retired")
				return providers
			},
			want: "unsupported status",
		},
		"missing source record types": {
			mutate: func(providers []Provider) []Provider {
				providers[0].descriptor.SourceRecordTypes = nil
				return providers
			},
			want: "no source_record_types",
		},
		"duplicate source record type": {
			mutate: func(providers []Provider) []Provider {
				providers[0].descriptor.SourceRecordTypes = []string{"host", "host"}
				return providers
			},
			want: "duplicate source_record_types",
		},
		"missing source authority modules": {
			mutate: func(providers []Provider) []Provider {
				providers[0].descriptor.SourceAuthorityModules = nil
				return providers
			},
			want: "no source_authority_modules",
		},
		"duplicate source authority module": {
			mutate: func(providers []Provider) []Provider {
				providers[0].descriptor.SourceAuthorityModules = []string{"entities", "entities"}
				return providers
			},
			want: "duplicate source_authority_modules",
		},
		"source authority modules omit source owner": {
			mutate: func(providers []Provider) []Provider {
				providers[0].descriptor.SourceAuthorityModules = []string{"links"}
				return providers
			},
			want: "omit source_owner_key",
		},
		"duplicate provider key": {
			mutate: func(providers []Provider) []Provider {
				providers[1].descriptor.ProviderKey = providers[0].descriptor.ProviderKey
				return providers
			},
			want: "duplicate projection provider key",
		},
		"duplicate view ownership": {
			mutate: func(providers []Provider) []Provider {
				providers[1].descriptor.ViewSchemaIDs = []string{hostsViewSchemaID}
				return providers
			},
			want: "duplicate projection provider ownership",
		},
		"duplicate projection table family": {
			mutate: func(providers []Provider) []Provider {
				providers[0].descriptor.ProjectionTableFamilies = []string{"host_grid_projection", "host_grid_projection"}
				return providers
			},
			want: "duplicate projection_table_families",
		},
		"projection storage owner mismatch": {
			mutate: func(providers []Provider) []Provider {
				providers[0].descriptor.ProjectionStorageOwnerKey = "entities"
				return providers
			},
			want: "must be projections",
		},
		"query capability mismatch": {
			mutate: func(providers []Provider) []Provider {
				providers[0].descriptor.Capabilities.Query = true
				return providers
			},
			want: "query capability does not match",
		},
		"refresh implementation without capability": {
			mutate: func(providers []Provider) []Provider {
				providers[0].descriptor.Capabilities.RefreshRow = false
				return providers
			},
			want: "refresh implementation without capability",
		},
		"restore rebuild required without capability": {
			mutate: func(providers []Provider) []Provider {
				providers[0].descriptor.Capabilities.RestoreRebuild = false
				return providers
			},
			want: "required restore rebuild without capability",
		},
		"active unsupported restore rebuild": {
			mutate: func(providers []Provider) []Provider {
				providers[0].descriptor.RestoreRebuild = RestoreRebuildUnsupported
				providers[0].descriptor.Capabilities.RestoreRebuild = false
				return providers
			},
			want: "active but declares unsupported restore rebuild",
		},
		"missing facade packages": {
			mutate: func(providers []Provider) []Provider {
				providers[0].descriptor.FacadePackages = nil
				return providers
			},
			want: "declares no facade_packages",
		},
		"duplicate facade package": {
			mutate: func(providers []Provider) []Provider {
				providers[0].descriptor.FacadePackages = []string{"internal/modules/entities", "internal/modules/entities"}
				return providers
			},
			want: "duplicate facade package",
		},
		"projection internal facade package": {
			mutate: func(providers []Provider) []Provider {
				providers[0].descriptor.FacadePackages = []string{"internal/modules/projections"}
				return providers
			},
			want: "must not expose projection internals",
		},
		"missing rebuild dependency": {
			mutate: func(providers []Provider) []Provider {
				providers[0].descriptor.RebuildAfter = []string{"missing"}
				return providers
			},
			want: "references unknown provider",
		},
		"rebuild cycle": {
			mutate: func(providers []Provider) []Provider {
				providers[0].descriptor.RebuildAfter = []string{"identity"}
				return providers
			},
			want: "cycle",
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := newProviderRegistry(test.mutate(cloneProjectionProviders(registryValidationProviders())))
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("newProviderRegistry error = %v, want containing %q", err, test.want)
			}
		})
	}
}

func registryValidationProviders() []Provider {
	base := func(key, view, recordType, table string, after []string) ProviderDescriptor {
		return ProviderDescriptor{
			SchemaVersion:             providercontract.DescriptorSchemaVersion,
			Status:                    ProviderStatusActive,
			ProviderKey:               key,
			SourceOwnerKey:            "entities",
			ViewSchemaIDs:             []string{view},
			SourceRecordTypes:         []string{recordType},
			SourceAuthorityModules:    []string{"entities"},
			ProjectionTableFamilies:   []string{table},
			ProjectionStorageOwnerKey: "projections",
			Capabilities: ProviderCapabilities{
				RefreshRow:      true,
				RestoreRebuild:  true,
				IncidentRebuild: true,
			},
			RestoreRebuild: RestoreRebuildRequired,
			FacadePackages: []string{"internal/modules/entities"},
			RebuildAfter:   after,
		}
	}
	return []Provider{
		NewHostProvider(base("host", hostsViewSchemaID, "host", "host_grid_projection", nil)),
		NewIdentityProvider(base("identity", identitiesViewSchemaID, "identity", "identity_grid_projection", []string{"host"})),
	}
}

func cloneProjectionProviders(providers []Provider) []Provider {
	cloned := make([]Provider, len(providers))
	copy(cloned, providers)
	for index := range cloned {
		cloned[index].descriptor.ViewSchemaIDs = append([]string(nil), cloned[index].descriptor.ViewSchemaIDs...)
		cloned[index].descriptor.SourceRecordTypes = append([]string(nil), cloned[index].descriptor.SourceRecordTypes...)
		cloned[index].descriptor.SourceAuthorityModules = append([]string(nil), cloned[index].descriptor.SourceAuthorityModules...)
		cloned[index].descriptor.ProjectionTableFamilies = append([]string(nil), cloned[index].descriptor.ProjectionTableFamilies...)
		cloned[index].descriptor.FacadePackages = append([]string(nil), cloned[index].descriptor.FacadePackages...)
		cloned[index].descriptor.RebuildAfter = append([]string(nil), cloned[index].descriptor.RebuildAfter...)
	}
	return cloned
}

func providerKeys(providers []*Provider) []string {
	keys := make([]string, 0, len(providers))
	for _, provider := range providers {
		keys = append(keys, provider.descriptor.ProviderKey)
	}
	return keys
}
