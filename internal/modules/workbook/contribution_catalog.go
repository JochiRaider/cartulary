package workbook

import (
	"context"
	"errors"
	"fmt"
	"reflect"
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

type queryProvider struct {
	query func(context.Context, QueryCommand) (querypage.Result, error)
}

func NewQueryProvider(
	query func(context.Context, QueryCommand) (querypage.Result, error),
) (QueryProvider, error) {
	if query == nil {
		return nil, errors.New("query provider requires query function")
	}
	return &queryProvider{query: query}, nil
}

func (provider *queryProvider) QueryRowsPage(
	ctx context.Context,
	command QueryCommand,
) (querypage.Result, error) {
	return provider.query(ctx, command)
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

// ActionCapabilityRequirements are supplied by application assembly. Workbook
// validates an exact immutable catalog against these keys without knowing
// which source modules currently provide the capabilities.
type ActionCapabilityRequirements struct {
	ClipboardViewSchemaIDs []string
	BulkViewSchemaIDs      []string
	LinkedNoteRecordTypes  []string
	SupersedeRecordTypes   []string
}

type ContributionCatalogInput struct {
	ProjectionDescriptors providercontract.DescriptorSet
	Queries               []QueryContribution
	Creates               []CreateContribution
	Patches               []PatchContribution
	Conflicts             []ConflictContribution
	ActionRequirements    ActionCapabilityRequirements
	Actions               MutationActionContributions
}

// WorkbookContributionCatalog is immutable after construction.
type WorkbookContributionCatalog struct {
	queries     map[string]QueryProvider
	creates     map[string]CreateProvider
	patches     map[string]PatchProvider
	conflicts   map[string]ConflictProvider
	clipboards  map[string]ClipboardProvider
	bulk        map[string]BulkProvider
	linkedNotes map[string]LinkedNoteProvider
	supersedes  map[string]SupersedeProvider
}

func NewWorkbookContributionCatalog(input ContributionCatalogInput) (*WorkbookContributionCatalog, error) {
	input = cloneContributionCatalogInput(input)
	descriptors := input.ProjectionDescriptors
	queryContributions := input.Queries
	createContributions := input.Creates
	patchContributions := input.Patches
	conflictContributions := input.Conflicts
	actionRequirements := input.ActionRequirements
	actionContributions := input.Actions
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
		if isNilContributionProvider(contribution.Provider) {
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
	requirements, err := validateActionCapabilityRequirements(actionRequirements)
	if err != nil {
		return nil, err
	}
	clipboards, err := validateClipboardContributions(actionContributions.Clipboard, requirements.clipboard)
	if err != nil {
		return nil, err
	}
	bulk, err := validateBulkContributions(actionContributions.Bulk, requirements.bulk)
	if err != nil {
		return nil, err
	}
	linkedNotes, err := validateLinkedNoteContributions(actionContributions.LinkedNote, requirements.linkedNote)
	if err != nil {
		return nil, err
	}
	supersedes, err := validateSupersedeContributions(actionContributions.Supersede, requirements.supersede)
	if err != nil {
		return nil, err
	}
	return newWorkbookContributionCatalog(contributionCatalogIndexes{
		queries: queries, creates: creates, patches: patches, conflicts: conflicts,
		clipboards: clipboards, bulk: bulk, linkedNotes: linkedNotes, supersedes: supersedes,
	}), nil
}

func (c *WorkbookContributionCatalog) ClipboardFor(viewSchemaID string) (ClipboardProvider, bool) {
	if c == nil {
		return nil, false
	}
	provider, ok := c.clipboards[viewSchemaID]
	return provider, ok
}

func (c *WorkbookContributionCatalog) BulkFor(viewSchemaID string) (BulkProvider, bool) {
	if c == nil {
		return nil, false
	}
	provider, ok := c.bulk[viewSchemaID]
	return provider, ok
}

func (c *WorkbookContributionCatalog) LinkedNoteFor(recordType string) (LinkedNoteProvider, bool) {
	if c == nil {
		return nil, false
	}
	provider, ok := c.linkedNotes[recordType]
	return provider, ok
}

func (c *WorkbookContributionCatalog) SupersedeFor(recordType string) (SupersedeProvider, bool) {
	if c == nil {
		return nil, false
	}
	provider, ok := c.supersedes[recordType]
	return provider, ok
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
		if isNilContributionProvider(contribution.Provider) {
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
		if isNilContributionProvider(contribution.Provider) {
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
		if isNilContributionProvider(contribution.Provider) {
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

type validatedActionRequirements struct {
	clipboard  map[string]struct{}
	bulk       map[string]struct{}
	linkedNote map[string]struct{}
	supersede  map[string]struct{}
}

func validateActionCapabilityRequirements(input ActionCapabilityRequirements) (validatedActionRequirements, error) {
	clipboard, err := validateRequiredKeys("clipboard", input.ClipboardViewSchemaIDs, true)
	if err != nil {
		return validatedActionRequirements{}, err
	}
	bulk, err := validateRequiredKeys("bulk", input.BulkViewSchemaIDs, true)
	if err != nil {
		return validatedActionRequirements{}, err
	}
	linkedNote, err := validateRequiredKeys("linked-note", input.LinkedNoteRecordTypes, false)
	if err != nil {
		return validatedActionRequirements{}, err
	}
	supersede, err := validateRequiredKeys("supersede", input.SupersedeRecordTypes, false)
	if err != nil {
		return validatedActionRequirements{}, err
	}
	return validatedActionRequirements{
		clipboard: clipboard, bulk: bulk, linkedNote: linkedNote, supersede: supersede,
	}, nil
}

func validateRequiredKeys(family string, keys []string, viewSchemaKeys bool) (map[string]struct{}, error) {
	required := make(map[string]struct{}, len(keys))
	for _, key := range append([]string(nil), keys...) {
		if key == "" {
			return nil, fmt.Errorf("workbook %s capability requirement has empty key", family)
		}
		if _, duplicate := required[key]; duplicate {
			return nil, fmt.Errorf("duplicate workbook %s capability requirement %q", family, key)
		}
		if viewSchemaKeys {
			if _, known := viewschema.Lookup(key); !known {
				return nil, fmt.Errorf("workbook %s capability requirement references unknown surface %q", family, key)
			}
		}
		required[key] = struct{}{}
	}
	return required, nil
}

func validateClipboardContributions(contributions []ClipboardContribution, expected map[string]struct{}) (map[string]ClipboardProvider, error) {
	providers := make(map[string]ClipboardProvider, len(contributions))
	for _, contribution := range contributions {
		if _, known := viewschema.Lookup(contribution.ViewSchemaID); !known {
			return nil, fmt.Errorf("workbook clipboard contribution references unknown surface %q", contribution.ViewSchemaID)
		}
		if _, required := expected[contribution.ViewSchemaID]; !required {
			return nil, fmt.Errorf("workbook clipboard contribution references unsupported surface %q", contribution.ViewSchemaID)
		}
		if _, duplicate := providers[contribution.ViewSchemaID]; duplicate {
			return nil, fmt.Errorf("duplicate workbook clipboard contribution for %q", contribution.ViewSchemaID)
		}
		if isNilContributionProvider(contribution.Provider) {
			return nil, fmt.Errorf("workbook clipboard contribution %q has nil provider", contribution.ViewSchemaID)
		}
		providers[contribution.ViewSchemaID] = contribution.Provider
	}
	for viewSchemaID := range expected {
		if _, ok := providers[viewSchemaID]; !ok {
			return nil, fmt.Errorf("workbook clipboard contribution missing required surface %q", viewSchemaID)
		}
	}
	return providers, nil
}

func validateBulkContributions(contributions []BulkContribution, expected map[string]struct{}) (map[string]BulkProvider, error) {
	providers := make(map[string]BulkProvider, len(contributions))
	for _, contribution := range contributions {
		if _, known := viewschema.Lookup(contribution.ViewSchemaID); !known {
			return nil, fmt.Errorf("workbook bulk contribution references unknown surface %q", contribution.ViewSchemaID)
		}
		if _, required := expected[contribution.ViewSchemaID]; !required {
			return nil, fmt.Errorf("workbook bulk contribution references unsupported surface %q", contribution.ViewSchemaID)
		}
		if _, duplicate := providers[contribution.ViewSchemaID]; duplicate {
			return nil, fmt.Errorf("duplicate workbook bulk contribution for %q", contribution.ViewSchemaID)
		}
		if isNilContributionProvider(contribution.Provider) {
			return nil, fmt.Errorf("workbook bulk contribution %q has nil provider", contribution.ViewSchemaID)
		}
		providers[contribution.ViewSchemaID] = contribution.Provider
	}
	for viewSchemaID := range expected {
		if _, ok := providers[viewSchemaID]; !ok {
			return nil, fmt.Errorf("workbook bulk contribution missing required surface %q", viewSchemaID)
		}
	}
	return providers, nil
}

func validateLinkedNoteContributions(contributions []LinkedNoteContribution, expected map[string]struct{}) (map[string]LinkedNoteProvider, error) {
	providers := make(map[string]LinkedNoteProvider, len(contributions))
	for _, contribution := range contributions {
		if _, required := expected[contribution.RecordType]; !required {
			return nil, fmt.Errorf("workbook linked-note contribution references unsupported record type %q", contribution.RecordType)
		}
		if _, duplicate := providers[contribution.RecordType]; duplicate {
			return nil, fmt.Errorf("duplicate workbook linked-note contribution for record type %q", contribution.RecordType)
		}
		if isNilContributionProvider(contribution.Provider) {
			return nil, fmt.Errorf("workbook linked-note contribution %q has nil provider", contribution.RecordType)
		}
		providers[contribution.RecordType] = contribution.Provider
	}
	for recordType := range expected {
		if _, ok := providers[recordType]; !ok {
			return nil, fmt.Errorf("workbook linked-note contribution missing required record type %q", recordType)
		}
	}
	return providers, nil
}

func validateSupersedeContributions(contributions []SupersedeContribution, expected map[string]struct{}) (map[string]SupersedeProvider, error) {
	providers := make(map[string]SupersedeProvider, len(contributions))
	for _, contribution := range contributions {
		if _, required := expected[contribution.RecordType]; !required {
			return nil, fmt.Errorf("workbook supersede contribution references unsupported record type %q", contribution.RecordType)
		}
		if _, duplicate := providers[contribution.RecordType]; duplicate {
			return nil, fmt.Errorf("duplicate workbook supersede contribution for record type %q", contribution.RecordType)
		}
		if isNilContributionProvider(contribution.Provider) {
			return nil, fmt.Errorf("workbook supersede contribution %q has nil provider", contribution.RecordType)
		}
		providers[contribution.RecordType] = contribution.Provider
	}
	for recordType := range expected {
		if _, ok := providers[recordType]; !ok {
			return nil, fmt.Errorf("workbook supersede contribution missing required record type %q", recordType)
		}
	}
	return providers, nil
}

func isNilContributionProvider(provider any) bool {
	if provider == nil {
		return true
	}
	value := reflect.ValueOf(provider)
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return value.IsNil()
	default:
		return false
	}
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
