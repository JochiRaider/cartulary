package restoreprobe

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"slices"
	"strings"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
	contract "github.com/JochiRaider/cartulary/internal/platform/workbookprobe"
)

const (
	RegistrationSchemaID = contract.RegistrationSchemaID
	BaseProfile          = contract.BaseProfile
)

var (
	ErrInvalidRegistration = errors.New("workbook restore probe: invalid registration")
	ErrRegistryConflict    = errors.New("workbook restore probe: registry conflict")
	ErrExecutionFailed     = errors.New("workbook restore probe: execution failed")

	identifierPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_.:-]{0,159}$`)
)

type Filter = contract.Filter
type Sort = contract.Sort
type Registration = contract.Registration
type Result = contract.Result

type ProjectionQuery interface {
	QueryRows(context.Context, uuid.UUID, string, viewschema.QueryMeta) ([]map[string]any, error)
}

type Executor = contract.Executor

type Registry struct {
	query    ProjectionQuery
	defaults map[string]Registration
}

func NewRegistry(query ProjectionQuery, registrations ...Registration) (*Registry, error) {
	if isNilProjectionQuery(query) {
		return nil, fmt.Errorf("%w: projection query is required", ErrInvalidRegistration)
	}
	if len(registrations) == 0 {
		return nil, fmt.Errorf("%w: at least one registration is required", ErrInvalidRegistration)
	}

	seenIDs := make(map[string]struct{}, len(registrations))
	profiles := make(map[string]struct{})
	defaults := make(map[string]Registration)
	for _, registration := range registrations {
		if err := validateRegistration(registration); err != nil {
			return nil, err
		}
		if _, duplicate := seenIDs[registration.RegistrationID]; duplicate {
			return nil, fmt.Errorf("%w: duplicate registration_id %q", ErrRegistryConflict, registration.RegistrationID)
		}
		seenIDs[registration.RegistrationID] = struct{}{}
		profiles[registration.Profile] = struct{}{}
		if registration.IsDefault {
			if existing, duplicate := defaults[registration.Profile]; duplicate {
				return nil, fmt.Errorf(
					"%w: profile %q has multiple defaults %q and %q",
					ErrRegistryConflict,
					registration.Profile,
					existing.RegistrationID,
					registration.RegistrationID,
				)
			}
			defaults[registration.Profile] = cloneRegistration(registration)
		}
	}
	for profile := range profiles {
		if _, ok := defaults[profile]; !ok {
			return nil, fmt.Errorf("%w: profile %q has no default", ErrRegistryConflict, profile)
		}
	}
	return &Registry{query: query, defaults: defaults}, nil
}

func isNilProjectionQuery(query ProjectionQuery) bool {
	if query == nil {
		return true
	}
	value := reflect.ValueOf(query)
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return value.IsNil()
	default:
		return false
	}
}

func (registry *Registry) ExecuteDefault(
	ctx context.Context,
	profile string,
	incidentID uuid.UUID,
) (Result, error) {
	if registry == nil || registry.query == nil {
		return Result{}, fmt.Errorf("%w: registry is unavailable", ErrExecutionFailed)
	}
	registration, ok := registry.defaults[profile]
	if !ok {
		return Result{}, fmt.Errorf("%w: profile %q has no default", ErrExecutionFailed, profile)
	}
	if incidentID == uuid.Nil {
		return Result{}, fmt.Errorf("%w: incident_id is required", ErrExecutionFailed)
	}
	meta := viewschema.QueryMeta{
		Filters: make([]viewschema.Filter, 0, len(registration.Filters)),
		Sort:    make([]viewschema.SortEntry, 0, len(registration.Sort)),
		GroupBy: registration.GroupBy,
	}
	for _, filter := range registration.Filters {
		meta.Filters = append(meta.Filters, viewschema.Filter{
			FieldKey: filter.FieldKey,
			Op:       filter.Op,
			Arg:      cloneMap(filter.Arg),
		})
	}
	for _, sortEntry := range registration.Sort {
		meta.Sort = append(meta.Sort, viewschema.SortEntry{
			FieldKey:  sortEntry.FieldKey,
			Direction: sortEntry.Direction,
		})
	}
	rows, err := registry.query.QueryRows(ctx, incidentID, registration.ViewSchemaID, meta)
	result := Result{
		RegistrationID: registration.RegistrationID,
		ViewSchemaID:   registration.ViewSchemaID,
	}
	if err != nil {
		return result, fmt.Errorf(
			"%w: registration %q: %v",
			ErrExecutionFailed,
			registration.RegistrationID,
			err,
		)
	}
	result.RowCount = int64(len(rows))
	return result, nil
}

func validateRegistration(registration Registration) error {
	if registration.SchemaID != RegistrationSchemaID {
		return fmt.Errorf("%w: unsupported schema_id %q", ErrInvalidRegistration, registration.SchemaID)
	}
	for name, value := range map[string]string{
		"registration_id": registration.RegistrationID,
		"owner_id":        registration.OwnerID,
		"profile":         registration.Profile,
		"view_schema_id":  registration.ViewSchemaID,
	} {
		if !identifierPattern.MatchString(value) {
			return fmt.Errorf("%w: %s is not an identifier", ErrInvalidRegistration, name)
		}
	}
	schema, ok := viewschema.Lookup(registration.ViewSchemaID)
	if !ok {
		return fmt.Errorf("%w: unknown view_schema_id %q", ErrInvalidRegistration, registration.ViewSchemaID)
	}
	sortFields := schema.SortFields()
	defaultSort := schema.DefaultSort()
	for index, entry := range registration.Sort {
		isDefaultEntry := index < len(defaultSort) &&
			defaultSort[index].FieldKey == entry.FieldKey &&
			defaultSort[index].Direction == entry.Direction
		if !slices.Contains(sortFields, entry.FieldKey) && !isDefaultEntry {
			return fmt.Errorf("%w: sort field %q is not admitted by %q", ErrInvalidRegistration, entry.FieldKey, registration.ViewSchemaID)
		}
		if entry.Direction != "asc" && entry.Direction != "desc" {
			return fmt.Errorf("%w: sort direction %q is unsupported", ErrInvalidRegistration, entry.Direction)
		}
	}
	filterFields := schema.FilterFields()
	for _, filter := range registration.Filters {
		if !slices.Contains(filterFields, filter.FieldKey) || strings.TrimSpace(filter.Op) == "" {
			return fmt.Errorf("%w: filter %q is not admitted by %q", ErrInvalidRegistration, filter.FieldKey, registration.ViewSchemaID)
		}
	}
	if registration.GroupBy != nil && !slices.Contains(schema.GroupingFields(), *registration.GroupBy) {
		return fmt.Errorf("%w: group_by %q is not admitted by %q", ErrInvalidRegistration, *registration.GroupBy, registration.ViewSchemaID)
	}
	if registration.RowRequirement != "zero_rows_allowed" {
		return fmt.Errorf("%w: unsupported row_requirement %q", ErrInvalidRegistration, registration.RowRequirement)
	}
	return nil
}

func cloneRegistration(registration Registration) Registration {
	cloned := registration
	cloned.Filters = append([]Filter(nil), registration.Filters...)
	for index := range cloned.Filters {
		cloned.Filters[index].Arg = cloneMap(cloned.Filters[index].Arg)
	}
	cloned.Sort = append([]Sort(nil), registration.Sort...)
	if registration.GroupBy != nil {
		value := *registration.GroupBy
		cloned.GroupBy = &value
	}
	return cloned
}

func cloneMap(source map[string]any) map[string]any {
	if source == nil {
		return nil
	}
	cloned := make(map[string]any, len(source))
	for key, value := range source {
		cloned[key] = value
	}
	return cloned
}
