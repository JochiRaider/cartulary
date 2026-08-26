package runtime

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	entityprojection "github.com/JochiRaider/cartulary/internal/modules/entities/workbookprojection"
	"github.com/JochiRaider/cartulary/internal/modules/projections/internal/queryengine"
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
	if got, want := providerKeys(registry.evidenceAssociationProviders("host")), []string{"host"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("host Evidence association providers: got %#v want %#v", got, want)
	}
	if got, want := providerKeys(registry.evidenceAssociationProviders("identity")), []string{"identity"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("identity Evidence association providers: got %#v want %#v", got, want)
	}
	if got := registry.evidenceAssociationProviders("note"); len(got) != 0 {
		t.Fatalf("unregistered note Evidence association providers: got %#v want none", providerKeys(got))
	}

	multiViewProviders := registryValidationProviders()
	multiViewProviders[1].descriptor.SourceRecordTypes = []string{"host"}
	multiViewRegistry, err := newProviderRegistry(multiViewProviders)
	if err != nil {
		t.Fatalf("multi-view provider registry: %v", err)
	}
	if got, want := providerKeys(multiViewRegistry.evidenceAssociationProviders("host")), []string{"host", "identity"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("deterministic host Evidence association providers: got %#v want %#v", got, want)
	}

	t.Run("generic rebuild callers share descriptor order and failure boundary", func(t *testing.T) {
		providers := registryValidationProviders()
		calls := make([]string, 0, len(providers))
		for index := range providers {
			providerID := providers[index].descriptor.ProviderID
			providers[index].rebuildIncidentTx = func(context.Context, *Store, pgx.Tx, uuid.UUID) error {
				calls = append(calls, providerID)
				return nil
			}
		}
		registry, err := newProviderRegistry(providers)
		if err != nil {
			t.Fatalf("generic rebuild registry: %v", err)
		}
		store := &Store{registry: registry}
		incidentID := uuid.New()
		want := []string{"host", "identity"}
		for name, rebuild := range map[string]func() error{
			"incident": func() error {
				return store.RebuildIncidentTx(t.Context(), nil, incidentID)
			},
			"import": func() error {
				return store.RebuildImportedIncidentTx(t.Context(), nil, incidentID)
			},
			"selected views": func() error {
				return store.RebuildIncidentViewsTx(
					t.Context(),
					nil,
					incidentID,
					[]string{identitiesViewSchemaID, hostsViewSchemaID},
				)
			},
		} {
			t.Run(name, func(t *testing.T) {
				calls = calls[:0]
				if err := rebuild(); err != nil {
					t.Fatalf("%s rebuild: %v", name, err)
				}
				if !reflect.DeepEqual(calls, want) {
					t.Fatalf("%s rebuild order: got %#v want %#v", name, calls, want)
				}
			})
		}

		sentinel := errors.New("characterized provider failure")
		registry.rebuildOrder[1].rebuildIncidentTx = func(context.Context, *Store, pgx.Tx, uuid.UUID) error {
			return sentinel
		}
		calls = calls[:0]
		if err := store.RebuildImportedIncidentTx(t.Context(), nil, incidentID); !errors.Is(err, sentinel) {
			t.Fatalf("generic rebuild failure = %v, want %v", err, sentinel)
		}
		if got := calls; !reflect.DeepEqual(got, []string{"host"}) {
			t.Fatalf("providers called before failure: got %#v want host only", got)
		}
	})
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
				providers[0].descriptor.Status = providercontract.ProviderStatus("retired")
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
			want: "omit source_owner_module",
		},
		"duplicate provider key": {
			mutate: func(providers []Provider) []Provider {
				providers[1].descriptor.ProviderID = providers[0].descriptor.ProviderID
				return providers
			},
			want: "duplicate projection provider_id",
		},
		"duplicate view ownership": {
			mutate: func(providers []Provider) []Provider {
				providers[1].descriptor.ViewSchemaIDs = []string{hostsViewSchemaID}
				return providers
			},
			want: "duplicate projection view ownership",
		},
		"duplicate projection table family": {
			mutate: func(providers []Provider) []Provider {
				providers[0].descriptor.ProjectionTableIDs = []string{"host_grid_projection", "host_grid_projection"}
				return providers
			},
			want: "duplicate projection_table_ids",
		},
		"projection storage owner mismatch": {
			mutate: func(providers []Provider) []Provider {
				providers[0].descriptor.ProjectionStorageOwnerModule = "entities"
				return providers
			},
			want: "must be projections",
		},
		"query capability mismatch": {
			mutate: func(providers []Provider) []Provider {
				providers[0].descriptor.Capabilities.Query = false
				return providers
			},
			want: "query capability does not match",
		},
		"unsupported query strategy": {
			mutate: func(providers []Provider) []Provider {
				providers[0].queryStrategy = queryStrategy(255)
				return providers
			},
			want: "unsupported query strategy",
		},
		"compiled query strategy without plans": {
			mutate: func(providers []Provider) []Provider {
				providers[0].queryStrategy = queryStrategyCompiledPlan
				return providers
			},
			want: "compiled query strategy has no plans",
		},
		"source hydration strategy with compiled plans": {
			mutate: func(providers []Provider) []Provider {
				providers[0].queryPlans = queryengine.TimelinePlans()
				return providers
			},
			want: "source-owner hydration strategy has compiled plans",
		},
		"compiled plans without query strategy": {
			mutate: func(providers []Provider) []Provider {
				providers[0].descriptor.Capabilities.Query = false
				providers[0].queryStrategy = queryStrategyNone
				providers[0].queryPlans = queryengine.TimelinePlans()
				return providers
			},
			want: "compiled plans without a query strategy",
		},
		"refresh implementation without capability": {
			mutate: func(providers []Provider) []Provider {
				providers[0].descriptor.Capabilities.RefreshRow = false
				return providers
			},
			want: "refresh implementation without capability",
		},
		"association fields without state loader": {
			mutate: func(providers []Provider) []Provider {
				providers[0].loadEvidenceAssociationStateTx = nil
				return providers
			},
			want: "Evidence association effects require a state loader",
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
				providers[0].descriptor.RestoreRebuild = providercontract.RestoreRebuildUnsupported
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
	base := func(key, view, recordType, table string, after []string) providercontract.ProviderDescriptor {
		return providercontract.ProviderDescriptor{
			SchemaVersion:                providercontract.DescriptorSchemaVersion,
			Status:                       providercontract.ProviderStatusActive,
			ProviderID:                   key,
			SourceOwnerModule:            "entities",
			ViewSchemaIDs:                []string{view},
			SourceRecordTypes:            []string{recordType},
			SourceAuthorityModules:       []string{"entities"},
			ProjectionTableIDs:           []string{table},
			ProjectionStorageOwnerModule: "projections",
			Capabilities: providercontract.ProviderCapabilities{
				Query:           true,
				RefreshRow:      true,
				RestoreRebuild:  true,
				IncidentRebuild: true,
			},
			RestoreRebuild: providercontract.RestoreRebuildRequired,
			FacadePackages: []string{"internal/modules/entities"},
			RebuildAfter:   after,
		}
	}
	return []Provider{
		NewHostProvider(base("host", hostsViewSchemaID, "host", "host_grid_projection", nil), &registryEntitySource{}),
		NewIdentityProvider(base("identity", identitiesViewSchemaID, "identity", "identity_grid_projection", []string{"host"}), &registryEntitySource{}),
	}
}

type registryEntitySource struct{ entityprojection.SourceReader }

func cloneProjectionProviders(providers []Provider) []Provider {
	cloned := make([]Provider, len(providers))
	copy(cloned, providers)
	for index := range cloned {
		cloned[index].descriptor.ViewSchemaIDs = append([]string(nil), cloned[index].descriptor.ViewSchemaIDs...)
		cloned[index].descriptor.SourceRecordTypes = append([]string(nil), cloned[index].descriptor.SourceRecordTypes...)
		cloned[index].descriptor.SourceAuthorityModules = append([]string(nil), cloned[index].descriptor.SourceAuthorityModules...)
		cloned[index].descriptor.ProjectionTableIDs = append([]string(nil), cloned[index].descriptor.ProjectionTableIDs...)
		cloned[index].descriptor.FacadePackages = append([]string(nil), cloned[index].descriptor.FacadePackages...)
		cloned[index].descriptor.RebuildAfter = append([]string(nil), cloned[index].descriptor.RebuildAfter...)
	}
	return cloned
}

func providerKeys(providers []*Provider) []string {
	keys := make([]string, 0, len(providers))
	for _, provider := range providers {
		keys = append(keys, provider.descriptor.ProviderID)
	}
	return keys
}
