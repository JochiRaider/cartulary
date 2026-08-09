package revisions

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

var (
	ErrDuplicateRecordViewRoute  = errors.New("revisions: duplicate record/view route")
	ErrMissingRecordViewRoute    = errors.New("revisions: missing record/view route")
	ErrUnexpectedRecordViewRoute = errors.New("revisions: unexpected record/view route")
	ErrUnknownRecordViewSchema   = errors.New("revisions: unknown record/view schema")
	ErrUnsupportedRecordVariant  = errors.New("revisions: unsupported record variant")
	ErrAmbiguousRecordViewRoute  = errors.New("revisions: ambiguous record/view route")
)

type RecordViewDescriptor struct {
	ContributionID string
	SourceOwner    SourceOwnerModule
	RecordType     string
	Variant        *RecordVariant
	ViewSchemaIDs  []string
}

// RecordViewSurface is the owner-neutral public view fact consumed by the
// Revisions catalog. Projection runtime capabilities are intentionally absent.
type RecordViewSurface struct {
	SourceRecordTypes []string
	ViewSchemaID      string
}

type recordViewRoute struct {
	RecordViewDescriptor
}

// RecordViewCatalog is the immutable, composition-scoped resolver compiled
// from source-owner contributions and validated against public view surfaces.
type RecordViewCatalog struct {
	ordered  []recordViewRoute
	byRecord map[string][]recordViewRoute
}

func NewRecordViewCatalog(
	contributions []ProviderContribution,
	surfaces []RecordViewSurface,
	knownViewSchemaIDs []string,
) (*RecordViewCatalog, error) {
	knownViews := make(map[string]struct{}, len(knownViewSchemaIDs))
	for _, viewSchemaID := range knownViewSchemaIDs {
		if viewSchemaID != "" {
			knownViews[viewSchemaID] = struct{}{}
		}
	}
	expected, err := expectedRecordViews(surfaces, knownViews)
	if err != nil {
		return nil, err
	}
	seenContributionIDs := map[string]struct{}{}
	seenViews := map[string]string{}
	seenSelectors := map[string]string{}
	routes := make([]recordViewRoute, 0)

	for _, contribution := range contributions {
		for _, record := range contribution.Records {
			switch record.LiveRecordChangePolicy {
			case LiveRecordChangeRequired:
				if len(record.RecordViewRoutes) == 0 {
					return nil, fmt.Errorf("%w: record type %q has required live changes but no routes", ErrMissingRecordViewRoute, record.RecordType)
				}
			case LiveRecordChangeNone:
				if len(record.RecordViewRoutes) != 0 {
					return nil, fmt.Errorf("%w: record type %q declares routes with policy none", ErrUnexpectedRecordViewRoute, record.RecordType)
				}
			default:
				return nil, fmt.Errorf("%w: record type %q has live policy %q", ErrUnexpectedRecordViewRoute, record.RecordType, record.LiveRecordChangePolicy)
			}
			for _, contributed := range record.RecordViewRoutes {
				route, err := normalizeRecordViewRoute(contribution.SourceOwnerModule, record.RecordType, contributed)
				if err != nil {
					return nil, err
				}
				if _, exists := seenContributionIDs[route.ContributionID]; exists {
					return nil, fmt.Errorf("%w: contribution id %q", ErrDuplicateRecordViewRoute, route.ContributionID)
				}
				seenContributionIDs[route.ContributionID] = struct{}{}
				selector := recordViewSelectorKey(route.RecordType, route.Variant)
				if existing, exists := seenSelectors[selector]; exists {
					return nil, fmt.Errorf("%w: selector %q is declared by %q and %q", ErrAmbiguousRecordViewRoute, selector, existing, route.ContributionID)
				}
				seenSelectors[selector] = route.ContributionID

				for _, viewSchemaID := range route.ViewSchemaIDs {
					if _, exists := knownViews[viewSchemaID]; !exists {
						return nil, fmt.Errorf("%w: %q", ErrUnknownRecordViewSchema, viewSchemaID)
					}
					expectedForRecord, expectedRecord := expected[route.RecordType]
					if !expectedRecord {
						return nil, fmt.Errorf("%w: record type %q", ErrUnexpectedRecordViewRoute, route.RecordType)
					}
					if _, exists := expectedForRecord[viewSchemaID]; !exists {
						return nil, fmt.Errorf("%w: record type %q view %q", ErrUnexpectedRecordViewRoute, route.RecordType, viewSchemaID)
					}
					if existing, exists := seenViews[viewSchemaID]; exists {
						return nil, fmt.Errorf("%w: view %q is declared by %q and %q", ErrDuplicateRecordViewRoute, viewSchemaID, existing, route.ContributionID)
					}
					seenViews[viewSchemaID] = route.ContributionID
				}
				routes = append(routes, recordViewRoute{RecordViewDescriptor: route})
			}
		}
	}

	missing := make([]string, 0)
	for recordType, expectedViews := range expected {
		for viewSchemaID := range expectedViews {
			if _, exists := seenViews[viewSchemaID]; !exists {
				missing = append(missing, recordType+":"+viewSchemaID)
			}
		}
	}
	if len(missing) != 0 {
		sort.Strings(missing)
		return nil, fmt.Errorf("%w: %v", ErrMissingRecordViewRoute, missing)
	}

	sort.Slice(routes, func(left, right int) bool {
		a, b := routes[left], routes[right]
		if a.RecordType != b.RecordType {
			return a.RecordType < b.RecordType
		}
		aKind, aValue := recordVariantParts(a.Variant)
		bKind, bValue := recordVariantParts(b.Variant)
		if aKind != bKind {
			return aKind < bKind
		}
		if aValue != bValue {
			return aValue < bValue
		}
		return a.ContributionID < b.ContributionID
	})
	byRecord := make(map[string][]recordViewRoute)
	for _, route := range routes {
		byRecord[route.RecordType] = append(byRecord[route.RecordType], cloneRecordViewRoute(route))
	}
	return &RecordViewCatalog{ordered: cloneRecordViewRoutes(routes), byRecord: byRecord}, nil
}

func (c *RecordViewCatalog) Descriptors() []RecordViewDescriptor {
	if c == nil {
		return nil
	}
	result := make([]RecordViewDescriptor, 0, len(c.ordered))
	for _, route := range c.ordered {
		result = append(result, cloneRecordViewDescriptor(route.RecordViewDescriptor))
	}
	return result
}

func (c *RecordViewCatalog) Resolve(recordType string, row map[string]any) (string, error) {
	if c == nil {
		return "", errors.New("revisions: record/view catalog is required")
	}
	routes := c.byRecord[recordType]
	if len(routes) == 0 {
		return "", fmt.Errorf("%w: record type %q", ErrUnexpectedRecordViewRoute, recordType)
	}
	if len(routes) == 1 && routes[0].Variant == nil {
		return routes[0].ViewSchemaIDs[0], nil
	}
	variantValue := firstCellPrefix(row)
	if variantValue == "" {
		variantValue = artifactTypeFromSourceSnapshot(row)
	}
	for _, route := range routes {
		if route.Variant != nil &&
			route.Variant.Kind == "artifact_type" &&
			route.Variant.Value == variantValue {
			return route.ViewSchemaIDs[0], nil
		}
	}
	return "", fmt.Errorf("%w: record type %q variant %q", ErrUnsupportedRecordVariant, recordType, variantValue)
}

func normalizeRecordViewRoute(
	owner SourceOwnerModule,
	recordType string,
	contributed RecordViewRouteContribution,
) (RecordViewDescriptor, error) {
	contributionID := strings.TrimSpace(contributed.ContributionID)
	if contributionID == "" || strings.TrimSpace(recordType) == "" || owner == "" {
		return RecordViewDescriptor{}, fmt.Errorf("%w: incomplete route contribution", ErrUnexpectedRecordViewRoute)
	}
	if len(contributed.ViewSchemaIDs) != 1 || strings.TrimSpace(contributed.ViewSchemaIDs[0]) == "" {
		return RecordViewDescriptor{}, fmt.Errorf("%w: contribution %q must identify exactly one view", ErrAmbiguousRecordViewRoute, contributionID)
	}
	var variant *RecordVariant
	if contributed.Variant != nil {
		if contributed.Variant.Kind != "artifact_type" ||
			strings.TrimSpace(contributed.Variant.Value) == "" ||
			recordType != "artifact" {
			return RecordViewDescriptor{}, fmt.Errorf("%w: contribution %q variant %#v", ErrUnsupportedRecordVariant, contributionID, contributed.Variant)
		}
		copyVariant := *contributed.Variant
		variant = &copyVariant
	}
	return RecordViewDescriptor{
		ContributionID: contributionID,
		SourceOwner:    owner,
		RecordType:     recordType,
		Variant:        variant,
		ViewSchemaIDs:  append([]string(nil), contributed.ViewSchemaIDs...),
	}, nil
}

func expectedRecordViews(
	surfaces []RecordViewSurface,
	knownViews map[string]struct{},
) (map[string]map[string]struct{}, error) {
	expected := map[string]map[string]struct{}{}
	seenViews := map[string]struct{}{}
	for _, surface := range surfaces {
		if _, known := knownViews[surface.ViewSchemaID]; !known {
			return nil, fmt.Errorf("%w: %q", ErrUnknownRecordViewSchema, surface.ViewSchemaID)
		}
		if _, duplicate := seenViews[surface.ViewSchemaID]; duplicate {
			return nil, fmt.Errorf("%w: public surface %q", ErrDuplicateRecordViewRoute, surface.ViewSchemaID)
		}
		seenViews[surface.ViewSchemaID] = struct{}{}
		if len(surface.SourceRecordTypes) == 0 {
			return nil, fmt.Errorf("%w: public surface %q has no source record types", ErrUnexpectedRecordViewRoute, surface.ViewSchemaID)
		}
		for _, recordType := range surface.SourceRecordTypes {
			if strings.TrimSpace(recordType) == "" {
				return nil, fmt.Errorf("%w: public surface %q has an empty source record type", ErrUnexpectedRecordViewRoute, surface.ViewSchemaID)
			}
			if expected[recordType] == nil {
				expected[recordType] = map[string]struct{}{}
			}
			expected[recordType][surface.ViewSchemaID] = struct{}{}
		}
	}
	return expected, nil
}

func recordViewSelectorKey(recordType string, variant *RecordVariant) string {
	kind, value := recordVariantParts(variant)
	return recordType + "\x00" + kind + "\x00" + value
}

func recordVariantParts(variant *RecordVariant) (string, string) {
	if variant == nil {
		return "", ""
	}
	return variant.Kind, variant.Value
}

func cloneRecordViewRoutes(routes []recordViewRoute) []recordViewRoute {
	result := make([]recordViewRoute, len(routes))
	for index, route := range routes {
		result[index] = cloneRecordViewRoute(route)
	}
	return result
}

func cloneRecordViewRoute(route recordViewRoute) recordViewRoute {
	return recordViewRoute{RecordViewDescriptor: cloneRecordViewDescriptor(route.RecordViewDescriptor)}
}

func cloneRecordViewDescriptor(descriptor RecordViewDescriptor) RecordViewDescriptor {
	descriptor.ViewSchemaIDs = append([]string(nil), descriptor.ViewSchemaIDs...)
	if descriptor.Variant != nil {
		copyVariant := *descriptor.Variant
		descriptor.Variant = &copyVariant
	}
	return descriptor
}

func artifactTypeFromSourceSnapshot(row map[string]any) string {
	source, ok := row["source"].(map[string]any)
	if !ok {
		return ""
	}
	artifactType, _ := source["artifact_type"].(string)
	return artifactType
}

func firstCellPrefix(row map[string]any) string {
	cells, ok := row["cells"].(map[string]any)
	if !ok {
		return ""
	}
	keys := make([]string, 0, len(cells))
	for key := range cells {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	if len(keys) == 0 {
		return ""
	}
	return strings.SplitN(keys[0], ".", 2)[0]
}
