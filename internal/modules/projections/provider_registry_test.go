package projections

import (
	"reflect"
	"strings"
	"testing"
)

func TestProjectionProviderRegistryBuiltIns(t *testing.T) {
	registry, err := newProviderRegistry(builtInProjectionProviders())
	if err != nil {
		t.Fatalf("built-in provider registry: %v", err)
	}
	wantOrder := []string{
		"timeline",
		"host",
		"identity",
		"indicator",
		"assessment",
		"artifact",
		"evidence",
		"party",
		"task_request",
		"decision",
	}
	if got := providerKeys(registry.rebuildOrder); !reflect.DeepEqual(got, wantOrder) {
		t.Fatalf("built-in rebuild order changed:\ngot  %#v\nwant %#v", got, wantOrder)
	}
	for viewSchemaID := range requiredProjectionViewSchemaIDs {
		if provider, ok := registry.providerForView(viewSchemaID); !ok || provider == nil {
			t.Fatalf("required projection surface %s has no provider", viewSchemaID)
		}
	}
	for _, optionalView := range []string{findingsViewSchemaID, investigativeQueriesViewSchemaID, forensicKeywordsViewSchemaID} {
		provider, ok := registry.providerForView(optionalView)
		if !ok || provider == nil || provider.descriptor.ProviderKey != "artifact" {
			t.Fatalf("optional artifact surface %s provider = %#v, %v", optionalView, provider, ok)
		}
	}
}

func TestProjectionProviderRegistryRejectsInvalidProviderSets(t *testing.T) {
	tests := map[string]struct {
		mutate func([]projectionProvider) []projectionProvider
		want   string
	}{
		"duplicate provider key": {
			mutate: func(providers []projectionProvider) []projectionProvider {
				providers[1].descriptor.ProviderKey = providers[0].descriptor.ProviderKey
				return providers
			},
			want: "duplicate projection provider key",
		},
		"duplicate view ownership": {
			mutate: func(providers []projectionProvider) []projectionProvider {
				providers[1].descriptor.ViewSchemaIDs = append(providers[1].descriptor.ViewSchemaIDs, timelineViewSchemaID)
				return providers
			},
			want: "duplicate projection provider ownership",
		},
		"missing required surface": {
			mutate: func(providers []projectionProvider) []projectionProvider {
				return providers[1:]
			},
			want: "has no provider",
		},
		"unknown projection table family": {
			mutate: func(providers []projectionProvider) []projectionProvider {
				providers[0].descriptor.ProjectionTableFamilies = []string{"unknown_grid_projection"}
				return providers
			},
			want: "unknown projection table family",
		},
		"schema owner mismatch": {
			mutate: func(providers []projectionProvider) []projectionProvider {
				providers[0].descriptor.SchemaOwnerKey = "timeline"
				return providers
			},
			want: "does not match",
		},
		"missing rebuild dependency": {
			mutate: func(providers []projectionProvider) []projectionProvider {
				providers[0].descriptor.RebuildAfter = []string{"missing"}
				return providers
			},
			want: "references unknown provider",
		},
		"rebuild cycle": {
			mutate: func(providers []projectionProvider) []projectionProvider {
				providers[0].descriptor.RebuildAfter = []string{"decision"}
				return providers
			},
			want: "cycle",
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := newProviderRegistry(tc.mutate(cloneProjectionProviders(builtInProjectionProviders())))
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("newProviderRegistry error = %v, want containing %q", err, tc.want)
			}
		})
	}
}

func cloneProjectionProviders(providers []projectionProvider) []projectionProvider {
	cloned := make([]projectionProvider, len(providers))
	copy(cloned, providers)
	for index := range cloned {
		cloned[index].descriptor.ViewSchemaIDs = append([]string(nil), cloned[index].descriptor.ViewSchemaIDs...)
		cloned[index].descriptor.SourceRecordTypes = append([]string(nil), cloned[index].descriptor.SourceRecordTypes...)
		cloned[index].descriptor.ProjectionTableFamilies = append([]string(nil), cloned[index].descriptor.ProjectionTableFamilies...)
		cloned[index].descriptor.RebuildAfter = append([]string(nil), cloned[index].descriptor.RebuildAfter...)
		cloned[index].descriptor.CharacterizationRefs = append([]string(nil), cloned[index].descriptor.CharacterizationRefs...)
	}
	return cloned
}

func providerKeys(providers []*projectionProvider) []string {
	keys := make([]string, 0, len(providers))
	for _, provider := range providers {
		keys = append(keys, provider.descriptor.ProviderKey)
	}
	return keys
}
