package workbook

import (
	"context"
	"fmt"
	"slices"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/projections/providercontract"
	"github.com/JochiRaider/cartulary/internal/platform/querypage"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
)

type QueryBackendKind string

const (
	QueryBackendSourceOwner QueryBackendKind = "source_owner"
	QueryBackendProjection  QueryBackendKind = "projection"
)

// QueryProvider is an exact-surface query contribution. Implementations do
// not receive a view-schema ID because the catalog binds that identity once
// during application assembly.
type QueryProvider interface {
	QueryRowsPage(context.Context, QueryCommand) (querypage.Result, error)
}

// QueryCommand carries the incident scope already authorized by Workbook,
// the exact catalog key, the normalized query, and the validated keyset
// window.
type QueryCommand struct {
	IncidentID   uuid.UUID
	ViewSchemaID string
	Query        viewschema.QueryMeta
	Window       querypage.Window
}

type QueryProviderFunc func(context.Context, QueryCommand) (querypage.Result, error)

func (f QueryProviderFunc) QueryRowsPage(
	ctx context.Context,
	command QueryCommand,
) (querypage.Result, error) {
	return f(ctx, command)
}

type QueryContribution struct {
	ViewSchemaID      string
	SourceOwnerKey    string
	SourceRecordTypes []string
	BackendKind       QueryBackendKind
	Provider          QueryProvider
}

type CreateContribution struct {
	ViewSchemaID      string
	SourceOwnerKey    string
	SourceRecordTypes []string
	Provider          CreateProvider
}

type PatchContribution struct {
	RecordType    string
	ViewSchemaIDs []string
	Provider      PatchProvider
}

type ConflictContribution struct {
	RecordType    string
	ViewSchemaIDs []string
	Provider      ConflictProvider
}

// WorkbookContributionCatalog is immutable after construction. Later slices
// add the conflict index to this same catalog.
type WorkbookContributionCatalog struct {
	queries   map[string]QueryProvider
	creates   map[string]CreateProvider
	patches   map[string]PatchProvider
	conflicts map[string]ConflictProvider
}

func NewWorkbookContributionCatalog(
	descriptors providercontract.DescriptorSet,
	queryContributions []QueryContribution,
	createContributions []CreateContribution,
	patchContributions []PatchContribution,
	conflictContributions []ConflictContribution,
) (*WorkbookContributionCatalog, error) {
	expected, err := expectedQuerySurfaces(descriptors)
	if err != nil {
		return nil, err
	}

	queries := make(map[string]QueryProvider, len(queryContributions))
	seen := make(map[string]struct{}, len(queryContributions))
	for _, contribution := range queryContributions {
		if contribution.ViewSchemaID == "" {
			return nil, fmt.Errorf("workbook query contribution has empty view_schema_id")
		}
		if _, exists := seen[contribution.ViewSchemaID]; exists {
			return nil, fmt.Errorf("duplicate workbook query contribution for %q", contribution.ViewSchemaID)
		}
		seen[contribution.ViewSchemaID] = struct{}{}

		surface, known := expected[contribution.ViewSchemaID]
		if !known {
			return nil, fmt.Errorf("workbook query contribution references unknown active surface %q", contribution.ViewSchemaID)
		}
		if contribution.SourceOwnerKey != surface.sourceOwnerKey {
			return nil, fmt.Errorf(
				"workbook query contribution %q source owner %q does not match descriptor owner %q",
				contribution.ViewSchemaID,
				contribution.SourceOwnerKey,
				surface.sourceOwnerKey,
			)
		}
		recordTypes := append([]string(nil), contribution.SourceRecordTypes...)
		slices.Sort(recordTypes)
		if !slices.Equal(recordTypes, surface.sourceRecordTypes) {
			return nil, fmt.Errorf(
				"workbook query contribution %q record types %v do not match active schema %v",
				contribution.ViewSchemaID,
				recordTypes,
				surface.sourceRecordTypes,
			)
		}
		if contribution.BackendKind != surface.backendKind {
			return nil, fmt.Errorf(
				"workbook query contribution %q backend %q does not match descriptor capability backend %q",
				contribution.ViewSchemaID,
				contribution.BackendKind,
				surface.backendKind,
			)
		}
		if contribution.Provider == nil {
			return nil, fmt.Errorf("workbook query contribution %q has nil provider", contribution.ViewSchemaID)
		}
		queries[contribution.ViewSchemaID] = contribution.Provider
	}

	for viewSchemaID := range expected {
		if _, ok := queries[viewSchemaID]; !ok {
			return nil, fmt.Errorf("workbook query contribution missing active surface %q", viewSchemaID)
		}
	}
	creates, err := validateCreateContributions(expected, createContributions)
	if err != nil {
		return nil, err
	}
	patches, err := validatePatchContributions(patchContributions)
	if err != nil {
		return nil, err
	}
	conflicts, err := validateConflictContributions(conflictContributions)
	if err != nil {
		return nil, err
	}
	return &WorkbookContributionCatalog{
		queries:   queries,
		creates:   creates,
		patches:   patches,
		conflicts: conflicts,
	}, nil
}

func (c *WorkbookContributionCatalog) ConflictFor(recordType string) (ConflictProvider, bool) {
	if c == nil {
		return nil, false
	}
	provider, ok := c.conflicts[recordType]
	return provider, ok
}

func (c *WorkbookContributionCatalog) QueryFor(viewSchemaID string) (QueryProvider, bool) {
	if c == nil {
		return nil, false
	}
	provider, ok := c.queries[viewSchemaID]
	return provider, ok
}

func (c *WorkbookContributionCatalog) CreateFor(viewSchemaID string) (CreateProvider, bool) {
	if c == nil {
		return nil, false
	}
	provider, ok := c.creates[viewSchemaID]
	return provider, ok
}

func (c *WorkbookContributionCatalog) PatchFor(recordType string) (PatchProvider, bool) {
	if c == nil {
		return nil, false
	}
	provider, ok := c.patches[recordType]
	return provider, ok
}

func (c *WorkbookContributionCatalog) QuerySurfaceIDs() []string {
	if c == nil {
		return []string{}
	}
	ids := make([]string, 0, len(c.queries))
	for viewSchemaID := range c.queries {
		ids = append(ids, viewSchemaID)
	}
	slices.Sort(ids)
	return ids
}

func (c *WorkbookContributionCatalog) QueryRows(
	ctx context.Context,
	incidentID uuid.UUID,
	viewSchemaID string,
	query viewschema.QueryMeta,
) ([]map[string]any, error) {
	page, err := c.QueryRowsPage(
		ctx,
		incidentID,
		viewSchemaID,
		query,
		querypage.Window{Limit: int(^uint(0)>>1) - 1},
	)
	return page.Rows, err
}

func (c *WorkbookContributionCatalog) QueryRowsPage(
	ctx context.Context,
	incidentID uuid.UUID,
	viewSchemaID string,
	query viewschema.QueryMeta,
	window querypage.Window,
) (querypage.Result, error) {
	provider, ok := c.QueryFor(viewSchemaID)
	if !ok {
		return querypage.Result{}, fmt.Errorf("workbook query surface %q is not registered", viewSchemaID)
	}
	return provider.QueryRowsPage(ctx, QueryCommand{
		IncidentID:   incidentID,
		ViewSchemaID: viewSchemaID,
		Query:        query,
		Window:       window,
	})
}

type expectedQuerySurface struct {
	sourceOwnerKey    string
	sourceRecordTypes []string
	backendKind       QueryBackendKind
}

func validateCreateContributions(
	expected map[string]expectedQuerySurface,
	contributions []CreateContribution,
) (map[string]CreateProvider, error) {
	creates := make(map[string]CreateProvider, len(contributions))
	for _, contribution := range contributions {
		schema, known := viewschema.Lookup(contribution.ViewSchemaID)
		if !known {
			return nil, fmt.Errorf("workbook create contribution references unknown active surface %q", contribution.ViewSchemaID)
		}
		if !schema.CreateCapable {
			return nil, fmt.Errorf("workbook create contribution registered for non-create-capable surface %q", contribution.ViewSchemaID)
		}
		if _, exists := creates[contribution.ViewSchemaID]; exists {
			return nil, fmt.Errorf("duplicate workbook create contribution for %q", contribution.ViewSchemaID)
		}
		surface := expected[contribution.ViewSchemaID]
		if contribution.SourceOwnerKey != surface.sourceOwnerKey {
			return nil, fmt.Errorf(
				"workbook create contribution %q source owner %q does not match descriptor owner %q",
				contribution.ViewSchemaID,
				contribution.SourceOwnerKey,
				surface.sourceOwnerKey,
			)
		}
		recordTypes := append([]string(nil), contribution.SourceRecordTypes...)
		slices.Sort(recordTypes)
		if !slices.Equal(recordTypes, surface.sourceRecordTypes) {
			return nil, fmt.Errorf(
				"workbook create contribution %q record types %v do not match active schema %v",
				contribution.ViewSchemaID,
				recordTypes,
				surface.sourceRecordTypes,
			)
		}
		if contribution.Provider == nil {
			return nil, fmt.Errorf("workbook create contribution %q has nil provider", contribution.ViewSchemaID)
		}
		creates[contribution.ViewSchemaID] = contribution.Provider
	}
	for viewSchemaID := range expected {
		schema, _ := viewschema.Lookup(viewSchemaID)
		_, registered := creates[viewSchemaID]
		if schema.CreateCapable && !registered {
			return nil, fmt.Errorf("workbook create contribution missing create-capable surface %q", viewSchemaID)
		}
		if !schema.CreateCapable && registered {
			return nil, fmt.Errorf("workbook create contribution unexpectedly registered surface %q", viewSchemaID)
		}
	}
	return creates, nil
}

func validatePatchContributions(contributions []PatchContribution) (map[string]PatchProvider, error) {
	expected := expectedPatchSurfaces()
	patches := make(map[string]PatchProvider, len(contributions))
	for _, contribution := range contributions {
		if contribution.RecordType == "" {
			return nil, fmt.Errorf("workbook patch contribution has empty record type")
		}
		if _, exists := patches[contribution.RecordType]; exists {
			return nil, fmt.Errorf("duplicate workbook patch contribution for record type %q", contribution.RecordType)
		}
		expectedViews, required := expected[contribution.RecordType]
		if !required {
			return nil, fmt.Errorf("workbook patch contribution references non-writable record type %q", contribution.RecordType)
		}
		viewSchemaIDs := append([]string(nil), contribution.ViewSchemaIDs...)
		slices.Sort(viewSchemaIDs)
		if !slices.Equal(viewSchemaIDs, expectedViews) {
			return nil, fmt.Errorf(
				"workbook patch contribution %q surfaces %v do not match active writable surfaces %v",
				contribution.RecordType,
				viewSchemaIDs,
				expectedViews,
			)
		}
		if contribution.Provider == nil {
			return nil, fmt.Errorf("workbook patch contribution %q has nil provider", contribution.RecordType)
		}
		patches[contribution.RecordType] = contribution.Provider
	}
	for recordType := range expected {
		if _, ok := patches[recordType]; !ok {
			return nil, fmt.Errorf("workbook patch contribution missing writable record type %q", recordType)
		}
	}
	return patches, nil
}

func validateConflictContributions(
	contributions []ConflictContribution,
) (map[string]ConflictProvider, error) {
	expected := expectedConflictSurfaces()
	conflicts := make(map[string]ConflictProvider, len(contributions))
	for _, contribution := range contributions {
		if contribution.RecordType == "" {
			return nil, fmt.Errorf("workbook conflict contribution has empty record type")
		}
		if _, exists := conflicts[contribution.RecordType]; exists {
			return nil, fmt.Errorf(
				"duplicate workbook conflict contribution for record type %q",
				contribution.RecordType,
			)
		}
		expectedViews, required := expected[contribution.RecordType]
		if !required {
			return nil, fmt.Errorf(
				"workbook conflict contribution references non-conflict-capable record type %q",
				contribution.RecordType,
			)
		}
		viewSchemaIDs := append([]string(nil), contribution.ViewSchemaIDs...)
		slices.Sort(viewSchemaIDs)
		if !slices.Equal(viewSchemaIDs, expectedViews) {
			return nil, fmt.Errorf(
				"workbook conflict contribution %q surfaces %v do not match active conflict-capable surfaces %v",
				contribution.RecordType,
				viewSchemaIDs,
				expectedViews,
			)
		}
		if contribution.Provider == nil {
			return nil, fmt.Errorf(
				"workbook conflict contribution %q has nil provider",
				contribution.RecordType,
			)
		}
		conflicts[contribution.RecordType] = contribution.Provider
	}
	for recordType := range expected {
		if _, ok := conflicts[recordType]; !ok {
			return nil, fmt.Errorf(
				"workbook conflict contribution missing conflict-capable record type %q",
				recordType,
			)
		}
	}
	return conflicts, nil
}

func expectedConflictSurfaces() map[string][]string {
	expected := map[string][]string{}
	for _, resource := range viewschema.ListPublicResources() {
		schema, ok := viewschema.Lookup(resource.ViewSchemaID)
		if !ok || !schemaHasConflictCapableField(schema) {
			continue
		}
		for _, recordType := range resource.SourceRecordTypes {
			expected[recordType] = append(expected[recordType], resource.ViewSchemaID)
		}
	}
	for recordType := range expected {
		slices.Sort(expected[recordType])
	}
	return expected
}

func schemaHasConflictCapableField(schema viewschema.Schema) bool {
	for _, field := range schema.Fields() {
		if field.Writable && field.ConflictResolutionClass != "" {
			return true
		}
	}
	return false
}

func expectedPatchSurfaces() map[string][]string {
	expected := map[string][]string{}
	for _, resource := range viewschema.ListPublicResources() {
		schema, ok := viewschema.Lookup(resource.ViewSchemaID)
		if !ok || !schemaHasWritableField(schema) {
			continue
		}
		for _, recordType := range resource.SourceRecordTypes {
			expected[recordType] = append(expected[recordType], resource.ViewSchemaID)
		}
	}
	for recordType := range expected {
		slices.Sort(expected[recordType])
	}
	return expected
}

func schemaHasWritableField(schema viewschema.Schema) bool {
	for _, field := range schema.Fields() {
		if field.Writable {
			return true
		}
	}
	return false
}

func expectedQuerySurfaces(descriptors providercontract.DescriptorSet) (map[string]expectedQuerySurface, error) {
	descriptorByView := make(map[string]providercontract.ProviderDescriptor)
	for _, descriptor := range descriptors.All() {
		if descriptor.Status != providercontract.ProviderStatusActive {
			continue
		}
		for _, viewSchemaID := range descriptor.ViewSchemaIDs {
			if prior, exists := descriptorByView[viewSchemaID]; exists {
				return nil, fmt.Errorf(
					"duplicate active projection descriptor ownership for %q: %q and %q",
					viewSchemaID,
					prior.ProviderID,
					descriptor.ProviderID,
				)
			}
			descriptorByView[viewSchemaID] = descriptor
		}
	}

	resources := viewschema.ListPublicResources()
	expected := make(map[string]expectedQuerySurface, len(resources))
	for _, resource := range resources {
		descriptor, ok := descriptorByView[resource.ViewSchemaID]
		if !ok {
			return nil, fmt.Errorf("active workbook surface %q has no active projection descriptor", resource.ViewSchemaID)
		}
		recordTypes := append([]string(nil), resource.SourceRecordTypes...)
		slices.Sort(recordTypes)
		descriptorRecordTypes := append([]string(nil), descriptor.SourceRecordTypes...)
		slices.Sort(descriptorRecordTypes)
		if !slices.Equal(recordTypes, descriptorRecordTypes) {
			return nil, fmt.Errorf(
				"active workbook surface %q record types %v do not match projection descriptor %q record types %v",
				resource.ViewSchemaID,
				recordTypes,
				descriptor.ProviderID,
				descriptorRecordTypes,
			)
		}
		backendKind := QueryBackendSourceOwner
		if descriptor.Capabilities.Query {
			backendKind = QueryBackendProjection
		}
		expected[resource.ViewSchemaID] = expectedQuerySurface{
			sourceOwnerKey:    descriptor.SourceOwnerModule,
			sourceRecordTypes: recordTypes,
			backendKind:       backendKind,
		}
	}
	for viewSchemaID, descriptor := range descriptorByView {
		if _, ok := expected[viewSchemaID]; !ok {
			return nil, fmt.Errorf(
				"active projection descriptor %q references unknown workbook surface %q",
				descriptor.ProviderID,
				viewSchemaID,
			)
		}
	}
	return expected, nil
}

func (s *Store) QueryRows(
	ctx context.Context,
	incidentID uuid.UUID,
	viewSchemaID string,
	query viewschema.QueryMeta,
) ([]map[string]any, error) {
	if s == nil || s.contributions == nil {
		return nil, fmt.Errorf("workbook contribution catalog is required")
	}
	return s.contributions.QueryRows(ctx, incidentID, viewSchemaID, query)
}

func (s *Store) QueryRowsPage(
	ctx context.Context,
	incidentID uuid.UUID,
	viewSchemaID string,
	query viewschema.QueryMeta,
	window querypage.Window,
) (querypage.Result, error) {
	if s == nil || s.contributions == nil {
		return querypage.Result{}, fmt.Errorf("workbook contribution catalog is required")
	}
	return s.contributions.QueryRowsPage(ctx, incidentID, viewSchemaID, query, window)
}
