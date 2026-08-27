package runtime

import (
	"fmt"
	"slices"
	"strings"

	"github.com/JochiRaider/cartulary/internal/modules/projections/internal/queryengine"
	"github.com/JochiRaider/cartulary/internal/modules/projections/providercontract"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
)

func validateProvider(provider Provider) error {
	descriptor := provider.descriptor
	if descriptor.Capabilities.RefreshRow && provider.refreshRowTx == nil {
		return fmt.Errorf("projection provider %q declares refresh support without implementation", descriptor.ProviderID)
	}
	if !descriptor.Capabilities.RefreshRow && provider.refreshRowTx != nil {
		return fmt.Errorf("projection provider %q has refresh implementation without capability", descriptor.ProviderID)
	}
	if descriptor.Capabilities.IncidentRebuild && provider.rebuildIncidentTx == nil {
		return fmt.Errorf("projection provider %q declares incident rebuild support without implementation", descriptor.ProviderID)
	}
	if !descriptor.Capabilities.IncidentRebuild && provider.rebuildIncidentTx != nil {
		return fmt.Errorf("projection provider %q has incident rebuild implementation without capability", descriptor.ProviderID)
	}
	if descriptor.Capabilities.Query != (provider.queryStrategy != queryStrategyNone) {
		return fmt.Errorf("projection provider %q query capability does not match registered query strategy", descriptor.ProviderID)
	}
	switch provider.queryStrategy {
	case queryStrategyNone:
		if len(provider.queryPlans) != 0 {
			return fmt.Errorf("projection provider %q has compiled plans without a query strategy", descriptor.ProviderID)
		}
	case queryStrategyCompiledPlan:
		if len(provider.queryPlans) == 0 {
			return fmt.Errorf("projection provider %q compiled query strategy has no plans", descriptor.ProviderID)
		}
	case queryStrategySourceOwnerHydration:
		if len(provider.queryPlans) != 0 {
			return fmt.Errorf("projection provider %q source-owner hydration strategy has compiled plans", descriptor.ProviderID)
		}
	default:
		return fmt.Errorf("projection provider %q has unsupported query strategy %d", descriptor.ProviderID, provider.queryStrategy)
	}
	declaredViews := map[string]struct{}{}
	for _, viewSchemaID := range descriptor.ViewSchemaIDs {
		declaredViews[viewSchemaID] = struct{}{}
	}
	seenPlans := map[string]struct{}{}
	querySurfaces, err := providerPlans(provider)
	if err != nil {
		return err
	}
	for _, surface := range querySurfaces {
		if surface.ViewSchemaID == "" {
			return fmt.Errorf("projection provider %q declares query surface with empty view_schema_id", descriptor.ProviderID)
		}
		if _, ok := declaredViews[surface.ViewSchemaID]; !ok {
			return fmt.Errorf("projection provider %q query surface %q is not one of its view_schema_ids", descriptor.ProviderID, surface.ViewSchemaID)
		}
		if _, exists := seenPlans[surface.ViewSchemaID]; exists {
			return fmt.Errorf("projection provider %q declares duplicate query surface %q", descriptor.ProviderID, surface.ViewSchemaID)
		}
		seenPlans[surface.ViewSchemaID] = struct{}{}
	}
	switch descriptor.RestoreRebuild {
	case providercontract.RestoreRebuildRequired:
		if provider.rebuildIncidentTx == nil {
			return fmt.Errorf("projection provider %q declares required restore rebuild without implementation", descriptor.ProviderID)
		}
	}
	return nil
}

func validateSemanticIntents(
	registry *providerRegistry,
	intents []providercontract.SurfaceIntent,
	intentOwners map[string]string,
) error {
	if registry == nil || len(intents) == 0 {
		return fmt.Errorf("projection query intents are empty")
	}
	intentByView := make(map[string]providercontract.SurfaceIntent, len(intents))
	for _, intent := range intents {
		if strings.TrimSpace(intent.ViewSchemaID) == "" {
			return fmt.Errorf("projection query intent has empty view_schema_id")
		}
		if len(intent.FieldKeys) == 0 {
			return fmt.Errorf("projection query intent %q has no field keys", intent.ViewSchemaID)
		}
		if _, exists := intentByView[intent.ViewSchemaID]; exists {
			return fmt.Errorf("duplicate projection query intent %q", intent.ViewSchemaID)
		}
		if err := validateUniqueStrings(intent.ViewSchemaID, "field_keys", intent.FieldKeys); err != nil {
			return err
		}
		provider, exists := registry.providerForView(intent.ViewSchemaID)
		if !exists {
			return fmt.Errorf("projection surface intent %q has no provider", intent.ViewSchemaID)
		}
		if owner := intentOwners[intent.ViewSchemaID]; owner != provider.descriptor.SourceOwnerModule {
			return fmt.Errorf(
				"projection surface intent %q is supplied by owner %q, want %q",
				intent.ViewSchemaID,
				owner,
				provider.descriptor.SourceOwnerModule,
			)
		}
		if !provider.descriptor.Capabilities.Query {
			return fmt.Errorf("projection provider %q has semantic intent without query capability", provider.descriptor.ProviderID)
		}
		schema, exists := viewschema.Lookup(intent.ViewSchemaID)
		if !exists {
			return fmt.Errorf("projection query intent %q has no view schema", intent.ViewSchemaID)
		}
		schemaFields := make([]string, 0, len(schema.Fields()))
		for fieldKey := range schema.Fields() {
			schemaFields = append(schemaFields, fieldKey)
		}
		if !equalStringSets(intent.FieldKeys, schemaFields) {
			return fmt.Errorf(
				"projection query intent %q fields do not match its view schema",
				intent.ViewSchemaID,
			)
		}
		if provider.queryStrategy == queryStrategyCompiledPlan {
			plan, exists := registry.querySurfaces[intent.ViewSchemaID]
			if !exists {
				return fmt.Errorf("projection query intent %q has no private compiled plan", intent.ViewSchemaID)
			}
			planFields := make([]string, 0, len(plan.Fields))
			for _, field := range plan.Fields {
				planFields = append(planFields, field.Key)
			}
			if !equalStringSets(intent.FieldKeys, planFields) {
				return fmt.Errorf(
					"projection query intent %q fields do not match its private compiled plan",
					intent.ViewSchemaID,
				)
			}
		}
		intentByView[intent.ViewSchemaID] = intent.Clone()
	}
	for _, provider := range registry.providers {
		for _, viewSchemaID := range provider.descriptor.ViewSchemaIDs {
			_, hasIntent := intentByView[viewSchemaID]
			if provider.descriptor.Capabilities.Query && !hasIntent {
				return fmt.Errorf(
					"projection provider %q query surface %q has no semantic intent",
					provider.descriptor.ProviderID,
					viewSchemaID,
				)
			}
			if !provider.descriptor.Capabilities.Query && hasIntent {
				return fmt.Errorf(
					"projection provider %q has semantic intent without query capability",
					provider.descriptor.ProviderID,
				)
			}
		}
	}
	for viewSchemaID := range registry.querySurfaces {
		if _, exists := intentByView[viewSchemaID]; !exists {
			return fmt.Errorf("private compiled plan %q has no semantic intent", viewSchemaID)
		}
	}
	return nil
}

func equalStringSets(left []string, right []string) bool {
	leftCopy := slices.Clone(left)
	rightCopy := slices.Clone(right)
	slices.Sort(leftCopy)
	slices.Sort(rightCopy)
	return slices.Equal(leftCopy, rightCopy)
}

func validateUniqueStrings(providerKey string, field string, values []string) error {
	seen := map[string]struct{}{}
	for _, value := range values {
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("projection provider %q declares empty %s value", providerKey, field)
		}
		if _, exists := seen[value]; exists {
			return fmt.Errorf("projection provider %q declares duplicate %s value %q", providerKey, field, value)
		}
		seen[value] = struct{}{}
	}
	return nil
}

func providerPlans(provider Provider) ([]queryengine.Surface, error) {
	surfaces := make([]queryengine.Surface, 0, len(provider.queryPlans))
	for _, surface := range provider.queryPlans {
		compiled, err := queryengine.CompileSurface(surface)
		if err != nil {
			return nil, fmt.Errorf("projection provider %q query surface: %w", provider.descriptor.ProviderID, err)
		}
		surfaces = append(surfaces, compiled)
	}
	return surfaces, nil
}
