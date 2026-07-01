package projections

import (
	"context"
	"fmt"
	"sort"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type ProviderDescriptor struct {
	ProviderKey              string
	SourceOwnerKey           string
	ViewSchemaIDs            []string
	SourceRecordTypes        []string
	ProjectionTableFamilies  []string
	SchemaOwnerKey           string
	RefreshRowSupported      bool
	RebuildIncidentSupported bool
	RebuildAfter             []string
	CharacterizationRefs     []string
}

type projectionProvider struct {
	descriptor        ProviderDescriptor
	refreshRowTx      func(context.Context, *Store, pgx.Tx, uuid.UUID) error
	rebuildIncidentTx func(context.Context, *Store, pgx.Tx, uuid.UUID) error
}

type providerRegistry struct {
	providers    []*projectionProvider
	byViewSchema map[string]*projectionProvider
	rebuildOrder []*projectionProvider
}

var requiredProjectionViewSchemaIDs = map[string]struct{}{
	timelineViewSchemaID:     {},
	hostsViewSchemaID:        {},
	identitiesViewSchemaID:   {},
	indicatorsViewSchemaID:   {},
	assessmentsViewSchemaID:  {},
	evidenceViewSchemaID:     {},
	notesViewSchemaID:        {},
	partiesViewSchemaID:      {},
	taskRequestsViewSchemaID: {},
	decisionsViewSchemaID:    {},
	commLogViewSchemaID:      {},
	handoffViewSchemaID:      {},
	statusReviewViewSchemaID: {},
	lessonViewSchemaID:       {},
}

var projectionTableSchemaOwners = map[string]string{
	"timeline_grid_projection":     "projections",
	"host_grid_projection":         "projections",
	"identity_grid_projection":     "projections",
	"indicator_grid_projection":    "projections",
	"assessment_grid_projection":   "projections",
	"artifact_grid_projection":     "projections",
	"evidence_grid_projection":     "projections",
	"party_grid_projection":        "projections",
	"task_request_grid_projection": "projections",
	"decision_grid_projection":     "projections",
}

func defaultProviderRegistry() *providerRegistry {
	registry, err := newProviderRegistry(builtInProjectionProviders())
	if err != nil {
		panic(fmt.Sprintf("invalid built-in projection provider registry: %v", err))
	}
	return registry
}

func (s *Store) providerRegistry() *providerRegistry {
	if s == nil || s.registry == nil {
		return defaultProviderRegistry()
	}
	return s.registry
}

func newProviderRegistry(providers []projectionProvider) (*providerRegistry, error) {
	if len(providers) == 0 {
		return nil, fmt.Errorf("projection provider registry is empty")
	}
	registry := &providerRegistry{
		providers:    make([]*projectionProvider, 0, len(providers)),
		byViewSchema: map[string]*projectionProvider{},
	}
	byProviderKey := map[string]*projectionProvider{}
	for index := range providers {
		provider := providers[index]
		if err := validateProvider(provider); err != nil {
			return nil, err
		}
		if _, exists := byProviderKey[provider.descriptor.ProviderKey]; exists {
			return nil, fmt.Errorf("duplicate projection provider key %q", provider.descriptor.ProviderKey)
		}
		providerCopy := provider
		providerPointer := &providerCopy
		byProviderKey[provider.descriptor.ProviderKey] = providerPointer
		registry.providers = append(registry.providers, providerPointer)
		seenViews := map[string]struct{}{}
		for _, viewSchemaID := range provider.descriptor.ViewSchemaIDs {
			if _, exists := seenViews[viewSchemaID]; exists {
				return nil, fmt.Errorf("projection provider %q declares duplicate view_schema_id %q", provider.descriptor.ProviderKey, viewSchemaID)
			}
			seenViews[viewSchemaID] = struct{}{}
			if existing := registry.byViewSchema[viewSchemaID]; existing != nil {
				return nil, fmt.Errorf("duplicate projection provider ownership for %q: %q and %q", viewSchemaID, existing.descriptor.ProviderKey, provider.descriptor.ProviderKey)
			}
			registry.byViewSchema[viewSchemaID] = providerPointer
		}
	}
	for viewSchemaID := range requiredProjectionViewSchemaIDs {
		if registry.byViewSchema[viewSchemaID] == nil {
			return nil, fmt.Errorf("required projection surface %q has no provider", viewSchemaID)
		}
	}
	rebuildOrder, err := topologicalProviderOrder(registry.providers, byProviderKey)
	if err != nil {
		return nil, err
	}
	registry.rebuildOrder = rebuildOrder
	return registry, nil
}

func validateProvider(provider projectionProvider) error {
	descriptor := provider.descriptor
	if descriptor.ProviderKey == "" {
		return fmt.Errorf("projection provider has empty provider_key")
	}
	if descriptor.SourceOwnerKey == "" {
		return fmt.Errorf("projection provider %q has empty source_owner_key", descriptor.ProviderKey)
	}
	if len(descriptor.ViewSchemaIDs) == 0 {
		return fmt.Errorf("projection provider %q declares no view_schema_ids", descriptor.ProviderKey)
	}
	if len(descriptor.SourceRecordTypes) == 0 {
		return fmt.Errorf("projection provider %q declares no source_record_types", descriptor.ProviderKey)
	}
	if len(descriptor.ProjectionTableFamilies) == 0 {
		return fmt.Errorf("projection provider %q declares no projection_table_families", descriptor.ProviderKey)
	}
	if descriptor.SchemaOwnerKey == "" {
		return fmt.Errorf("projection provider %q has empty schema_owner_key", descriptor.ProviderKey)
	}
	if descriptor.RefreshRowSupported && provider.refreshRowTx == nil {
		return fmt.Errorf("projection provider %q declares refresh support without implementation", descriptor.ProviderKey)
	}
	if descriptor.RebuildIncidentSupported && provider.rebuildIncidentTx == nil {
		return fmt.Errorf("projection provider %q declares rebuild support without implementation", descriptor.ProviderKey)
	}
	for _, family := range descriptor.ProjectionTableFamilies {
		owner, ok := projectionTableSchemaOwners[family]
		if !ok {
			return fmt.Errorf("projection provider %q declares unknown projection table family %q", descriptor.ProviderKey, family)
		}
		if owner != descriptor.SchemaOwnerKey {
			return fmt.Errorf("projection provider %q schema_owner_key=%q does not match %s owner %q", descriptor.ProviderKey, descriptor.SchemaOwnerKey, family, owner)
		}
	}
	return nil
}

func topologicalProviderOrder(providers []*projectionProvider, byProviderKey map[string]*projectionProvider) ([]*projectionProvider, error) {
	remaining := map[string]*projectionProvider{}
	indegree := map[string]int{}
	outgoing := map[string][]string{}
	for _, provider := range providers {
		key := provider.descriptor.ProviderKey
		remaining[key] = provider
		indegree[key] = 0
	}
	for _, provider := range providers {
		key := provider.descriptor.ProviderKey
		for _, dependency := range provider.descriptor.RebuildAfter {
			if byProviderKey[dependency] == nil {
				return nil, fmt.Errorf("projection provider %q rebuild_after references unknown provider %q", key, dependency)
			}
			indegree[key]++
			outgoing[dependency] = append(outgoing[dependency], key)
		}
	}
	ordered := make([]*projectionProvider, 0, len(providers))
	for len(remaining) > 0 {
		ready := make([]*projectionProvider, 0)
		for key, provider := range remaining {
			if indegree[key] == 0 {
				ready = append(ready, provider)
			}
		}
		if len(ready) == 0 {
			return nil, fmt.Errorf("projection provider rebuild graph has a cycle")
		}
		sort.Slice(ready, func(left, right int) bool {
			return providerSortKey(ready[left]) < providerSortKey(ready[right])
		})
		next := ready[0]
		nextKey := next.descriptor.ProviderKey
		ordered = append(ordered, next)
		delete(remaining, nextKey)
		for _, dependent := range outgoing[nextKey] {
			indegree[dependent]--
		}
	}
	return ordered, nil
}

func providerSortKey(provider *projectionProvider) string {
	viewIDs := append([]string(nil), provider.descriptor.ViewSchemaIDs...)
	sort.Strings(viewIDs)
	firstView := ""
	if len(viewIDs) > 0 {
		firstView = viewIDs[0]
	}
	return provider.descriptor.ProviderKey + "\x00" + firstView
}

func (r *providerRegistry) providerForView(viewSchemaID string) (*projectionProvider, bool) {
	if r == nil {
		return nil, false
	}
	provider := r.byViewSchema[viewSchemaID]
	return provider, provider != nil
}

func (s *Store) refreshProjectionRowTx(ctx context.Context, tx pgx.Tx, viewSchemaID string, recordID uuid.UUID) error {
	provider, ok := s.providerRegistry().providerForView(viewSchemaID)
	if !ok || !provider.descriptor.RefreshRowSupported || provider.refreshRowTx == nil {
		return fmt.Errorf("projection refresh surface %q not mapped", viewSchemaID)
	}
	return provider.refreshRowTx(ctx, s, tx, recordID)
}

func (s *Store) rebuildProjectionIncidentTx(ctx context.Context, tx pgx.Tx, viewSchemaID string, incidentID uuid.UUID) error {
	provider, ok := s.providerRegistry().providerForView(viewSchemaID)
	if !ok || !provider.descriptor.RebuildIncidentSupported || provider.rebuildIncidentTx == nil {
		return fmt.Errorf("projection rebuild surface %q not mapped", viewSchemaID)
	}
	return provider.rebuildIncidentTx(ctx, s, tx, incidentID)
}

func builtInProjectionProviders() []projectionProvider {
	return []projectionProvider{
		{
			descriptor: ProviderDescriptor{
				ProviderKey:              "timeline",
				SourceOwnerKey:           "timeline",
				ViewSchemaIDs:            []string{timelineViewSchemaID},
				SourceRecordTypes:        []string{"timeline_event"},
				ProjectionTableFamilies:  []string{"timeline_grid_projection"},
				SchemaOwnerKey:           "projections",
				RebuildIncidentSupported: true,
				CharacterizationRefs:     []string{"internal/modules/timeline/phase3_projection_contract_test.go"},
			},
			rebuildIncidentTx: func(ctx context.Context, store *Store, tx pgx.Tx, incidentID uuid.UUID) error {
				return store.rebuildIncidentTimelineTxCore(ctx, tx, incidentID)
			},
		},
		{
			descriptor: ProviderDescriptor{
				ProviderKey:              "host",
				SourceOwnerKey:           "entities",
				ViewSchemaIDs:            []string{hostsViewSchemaID},
				SourceRecordTypes:        []string{"host"},
				ProjectionTableFamilies:  []string{"host_grid_projection"},
				SchemaOwnerKey:           "projections",
				RebuildIncidentSupported: true,
				RebuildAfter:             []string{"timeline"},
				CharacterizationRefs:     []string{"internal/modules/entities/phase4_integration_test.go"},
			},
			rebuildIncidentTx: func(ctx context.Context, store *Store, tx pgx.Tx, incidentID uuid.UUID) error {
				return store.rebuildIncidentHostsTxCore(ctx, tx, incidentID)
			},
		},
		{
			descriptor: ProviderDescriptor{
				ProviderKey:              "identity",
				SourceOwnerKey:           "entities",
				ViewSchemaIDs:            []string{identitiesViewSchemaID},
				SourceRecordTypes:        []string{"identity"},
				ProjectionTableFamilies:  []string{"identity_grid_projection"},
				SchemaOwnerKey:           "projections",
				RebuildIncidentSupported: true,
				RebuildAfter:             []string{"host"},
				CharacterizationRefs:     []string{"internal/modules/entities/phase4_integration_test.go"},
			},
			rebuildIncidentTx: func(ctx context.Context, store *Store, tx pgx.Tx, incidentID uuid.UUID) error {
				return store.rebuildIncidentIdentitiesTxCore(ctx, tx, incidentID)
			},
		},
		{
			descriptor: ProviderDescriptor{
				ProviderKey:              "indicator",
				SourceOwnerKey:           "indicators",
				ViewSchemaIDs:            []string{indicatorsViewSchemaID},
				SourceRecordTypes:        []string{"indicator"},
				ProjectionTableFamilies:  []string{"indicator_grid_projection"},
				SchemaOwnerKey:           "projections",
				RebuildIncidentSupported: true,
				RebuildAfter:             []string{"identity"},
				CharacterizationRefs:     []string{"internal/modules/entities/phase9_indicators_test.go"},
			},
			rebuildIncidentTx: func(ctx context.Context, store *Store, tx pgx.Tx, incidentID uuid.UUID) error {
				return store.rebuildIncidentIndicatorsTxCore(ctx, tx, incidentID)
			},
		},
		{
			descriptor: ProviderDescriptor{
				ProviderKey:              "assessment",
				SourceOwnerKey:           "assessments",
				ViewSchemaIDs:            []string{assessmentsViewSchemaID},
				SourceRecordTypes:        []string{"assessment"},
				ProjectionTableFamilies:  []string{"assessment_grid_projection"},
				SchemaOwnerKey:           "projections",
				RefreshRowSupported:      true,
				RebuildIncidentSupported: true,
				RebuildAfter:             []string{"indicator"},
				CharacterizationRefs:     []string{"internal/modules/assessments/phase9_assessment_contract_test.go", "internal/modules/projections/query_test.go"},
			},
			refreshRowTx: func(ctx context.Context, store *Store, tx pgx.Tx, recordID uuid.UUID) error {
				return store.refreshAssessmentTxCore(ctx, tx, recordID)
			},
			rebuildIncidentTx: func(ctx context.Context, store *Store, tx pgx.Tx, incidentID uuid.UUID) error {
				return store.rebuildIncidentAssessmentsTxCore(ctx, tx, incidentID)
			},
		},
		{
			descriptor: ProviderDescriptor{
				ProviderKey:    "artifact",
				SourceOwnerKey: "artifacts",
				ViewSchemaIDs: []string{
					notesViewSchemaID,
					commLogViewSchemaID,
					handoffViewSchemaID,
					statusReviewViewSchemaID,
					lessonViewSchemaID,
					findingsViewSchemaID,
					investigativeQueriesViewSchemaID,
					forensicKeywordsViewSchemaID,
				},
				SourceRecordTypes:        []string{"artifact"},
				ProjectionTableFamilies:  []string{"artifact_grid_projection"},
				SchemaOwnerKey:           "projections",
				RefreshRowSupported:      true,
				RebuildIncidentSupported: true,
				RebuildAfter:             []string{"assessment"},
				CharacterizationRefs:     []string{"internal/modules/workbook/phase9_coordination_surfaces_test.go", "internal/modules/projections/query_test.go"},
			},
			refreshRowTx: func(ctx context.Context, store *Store, tx pgx.Tx, recordID uuid.UUID) error {
				return store.refreshArtifactTxCore(ctx, tx, recordID)
			},
			rebuildIncidentTx: func(ctx context.Context, store *Store, tx pgx.Tx, incidentID uuid.UUID) error {
				return store.rebuildIncidentArtifactsTxCore(ctx, tx, incidentID)
			},
		},
		{
			descriptor: ProviderDescriptor{
				ProviderKey:              "evidence",
				SourceOwnerKey:           "evidence",
				ViewSchemaIDs:            []string{evidenceViewSchemaID},
				SourceRecordTypes:        []string{"evidence"},
				ProjectionTableFamilies:  []string{"evidence_grid_projection"},
				SchemaOwnerKey:           "projections",
				RefreshRowSupported:      true,
				RebuildIncidentSupported: true,
				RebuildAfter:             []string{"artifact"},
				CharacterizationRefs:     []string{"internal/modules/evidence/phase5_integration_test.go", "internal/modules/projections/query_test.go"},
			},
			refreshRowTx: func(ctx context.Context, store *Store, tx pgx.Tx, recordID uuid.UUID) error {
				return store.refreshEvidenceTxCore(ctx, tx, recordID)
			},
			rebuildIncidentTx: func(ctx context.Context, store *Store, tx pgx.Tx, incidentID uuid.UUID) error {
				return store.rebuildIncidentEvidenceTxCore(ctx, tx, incidentID)
			},
		},
		{
			descriptor: ProviderDescriptor{
				ProviderKey:              "party",
				SourceOwnerKey:           "parties",
				ViewSchemaIDs:            []string{partiesViewSchemaID},
				SourceRecordTypes:        []string{"party"},
				ProjectionTableFamilies:  []string{"party_grid_projection"},
				SchemaOwnerKey:           "projections",
				RefreshRowSupported:      true,
				RebuildIncidentSupported: true,
				RebuildAfter:             []string{"evidence"},
				CharacterizationRefs:     []string{"internal/modules/workbook/phase9_parties_integration_test.go", "internal/modules/projections/query_test.go"},
			},
			refreshRowTx: func(ctx context.Context, store *Store, tx pgx.Tx, recordID uuid.UUID) error {
				return store.refreshPartyTxCore(ctx, tx, recordID)
			},
			rebuildIncidentTx: func(ctx context.Context, store *Store, tx pgx.Tx, incidentID uuid.UUID) error {
				return store.rebuildIncidentPartiesTxCore(ctx, tx, incidentID)
			},
		},
		{
			descriptor: ProviderDescriptor{
				ProviderKey:              "task_request",
				SourceOwnerKey:           "tasksdecisions",
				ViewSchemaIDs:            []string{taskRequestsViewSchemaID},
				SourceRecordTypes:        []string{"task_request"},
				ProjectionTableFamilies:  []string{"task_request_grid_projection"},
				SchemaOwnerKey:           "projections",
				RefreshRowSupported:      true,
				RebuildIncidentSupported: true,
				RebuildAfter:             []string{"party"},
				CharacterizationRefs:     []string{"internal/modules/workbook/phase9_task_decisions_store_test.go", "internal/modules/projections/query_test.go"},
			},
			refreshRowTx: func(ctx context.Context, store *Store, tx pgx.Tx, recordID uuid.UUID) error {
				return store.refreshTaskRequestTxCore(ctx, tx, recordID)
			},
			rebuildIncidentTx: func(ctx context.Context, store *Store, tx pgx.Tx, incidentID uuid.UUID) error {
				return store.rebuildIncidentTaskRequestsTxCore(ctx, tx, incidentID)
			},
		},
		{
			descriptor: ProviderDescriptor{
				ProviderKey:              "decision",
				SourceOwnerKey:           "tasksdecisions",
				ViewSchemaIDs:            []string{decisionsViewSchemaID},
				SourceRecordTypes:        []string{"decision"},
				ProjectionTableFamilies:  []string{"decision_grid_projection"},
				SchemaOwnerKey:           "projections",
				RefreshRowSupported:      true,
				RebuildIncidentSupported: true,
				RebuildAfter:             []string{"task_request"},
				CharacterizationRefs:     []string{"internal/modules/workbook/phase9_task_decisions_store_test.go", "internal/modules/projections/query_test.go"},
			},
			refreshRowTx: func(ctx context.Context, store *Store, tx pgx.Tx, recordID uuid.UUID) error {
				return store.refreshDecisionTxCore(ctx, tx, recordID)
			},
			rebuildIncidentTx: func(ctx context.Context, store *Store, tx pgx.Tx, incidentID uuid.UUID) error {
				return store.rebuildIncidentDecisionsTxCore(ctx, tx, incidentID)
			},
		},
	}
}
