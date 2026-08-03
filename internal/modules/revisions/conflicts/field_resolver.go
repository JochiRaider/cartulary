package conflicts

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

var (
	ErrDuplicateFieldResolver  = errors.New("revisions conflicts: duplicate field resolver")
	ErrMissingFieldResolver    = errors.New("revisions conflicts: missing field resolver")
	ErrUnexpectedFieldResolver = errors.New("revisions conflicts: unexpected field resolver")
	ErrInvalidFieldContract    = errors.New("revisions conflicts: invalid field contract")
	ErrFieldNotFound           = errors.New("revisions conflicts: field not found")
	ErrFieldNotWritable        = errors.New("revisions conflicts: field is not writable")
)

type FieldDescriptor struct {
	FieldKey                string
	ValueKind               string
	Writable                bool
	ConflictResolutionClass string
}

type FieldDescriptorSet struct {
	fields map[string]FieldDescriptor
}

type FieldResolverContribution struct {
	ProviderID    string
	SourceOwnerID string
	ViewSchemaID  string
	Fields        []FieldDescriptor
}

type FieldResolver interface {
	ResolveViewSchema(viewSchemaID string) (FieldDescriptorSet, error)
	ResolveWritableField(viewSchemaID string, fieldKey string) (FieldDescriptor, error)
}

// FieldResolverCatalog is an immutable startup-built snapshot. It retains no
// reference to a mutable or process-global view-schema registry.
type FieldResolverCatalog struct {
	byViewSchema map[string]FieldDescriptorSet
}

func NewFieldResolverCatalog(requiredViewSchemaIDs []string, contributions ...FieldResolverContribution) (*FieldResolverCatalog, error) {
	required := make(map[string]struct{}, len(requiredViewSchemaIDs))
	for _, viewSchemaID := range requiredViewSchemaIDs {
		if strings.TrimSpace(viewSchemaID) == "" {
			return nil, fmt.Errorf("%w: required view schema id is blank", ErrInvalidFieldContract)
		}
		if _, duplicate := required[viewSchemaID]; duplicate {
			return nil, fmt.Errorf("%w: required view schema %q", ErrDuplicateFieldResolver, viewSchemaID)
		}
		required[viewSchemaID] = struct{}{}
	}

	byViewSchema := make(map[string]FieldDescriptorSet, len(contributions))
	providerIDs := make(map[string]struct{}, len(contributions))
	for _, contribution := range contributions {
		if strings.TrimSpace(contribution.ProviderID) == "" ||
			strings.TrimSpace(contribution.SourceOwnerID) == "" ||
			strings.TrimSpace(contribution.ViewSchemaID) == "" {
			return nil, fmt.Errorf("%w: provider, source owner, and view schema are required", ErrInvalidFieldContract)
		}
		if _, expected := required[contribution.ViewSchemaID]; !expected {
			return nil, fmt.Errorf("%w: view schema %q", ErrUnexpectedFieldResolver, contribution.ViewSchemaID)
		}
		if _, duplicate := providerIDs[contribution.ProviderID]; duplicate {
			return nil, fmt.Errorf("%w: provider %q", ErrDuplicateFieldResolver, contribution.ProviderID)
		}
		if _, duplicate := byViewSchema[contribution.ViewSchemaID]; duplicate {
			return nil, fmt.Errorf("%w: view schema %q", ErrDuplicateFieldResolver, contribution.ViewSchemaID)
		}
		providerIDs[contribution.ProviderID] = struct{}{}
		set, err := newValidatedFieldDescriptorSet(contribution.ViewSchemaID, contribution.Fields)
		if err != nil {
			return nil, err
		}
		byViewSchema[contribution.ViewSchemaID] = set
	}

	missing := make([]string, 0)
	for viewSchemaID := range required {
		if _, present := byViewSchema[viewSchemaID]; !present {
			missing = append(missing, viewSchemaID)
		}
	}
	if len(missing) > 0 {
		sort.Strings(missing)
		return nil, fmt.Errorf("%w: view schemas %v", ErrMissingFieldResolver, missing)
	}
	return &FieldResolverCatalog{byViewSchema: byViewSchema}, nil
}

func NewFieldDescriptorSet(fields []FieldDescriptor) FieldDescriptorSet {
	set, _ := newValidatedFieldDescriptorSet("direct", fields)
	return set
}

func newValidatedFieldDescriptorSet(viewSchemaID string, fields []FieldDescriptor) (FieldDescriptorSet, error) {
	result := FieldDescriptorSet{fields: make(map[string]FieldDescriptor, len(fields))}
	for _, field := range fields {
		if strings.TrimSpace(field.FieldKey) == "" || strings.TrimSpace(field.ValueKind) == "" {
			return FieldDescriptorSet{}, fmt.Errorf("%w: view schema %q has an incomplete field", ErrInvalidFieldContract, viewSchemaID)
		}
		if field.Writable && strings.TrimSpace(field.ConflictResolutionClass) == "" {
			return FieldDescriptorSet{}, fmt.Errorf("%w: writable field %q in view schema %q has no conflict class", ErrInvalidFieldContract, field.FieldKey, viewSchemaID)
		}
		if _, duplicate := result.fields[field.FieldKey]; duplicate {
			return FieldDescriptorSet{}, fmt.Errorf("%w: field %q in view schema %q", ErrDuplicateFieldResolver, field.FieldKey, viewSchemaID)
		}
		result.fields[field.FieldKey] = field
	}
	return result, nil
}

func (c *FieldResolverCatalog) ResolveViewSchema(viewSchemaID string) (FieldDescriptorSet, error) {
	if c == nil {
		return FieldDescriptorSet{}, fmt.Errorf("%w: view schema %q", ErrMissingFieldResolver, viewSchemaID)
	}
	set, ok := c.byViewSchema[viewSchemaID]
	if !ok {
		return FieldDescriptorSet{}, fmt.Errorf("%w: view schema %q", ErrMissingFieldResolver, viewSchemaID)
	}
	return set, nil
}

func (c *FieldResolverCatalog) ResolveWritableField(viewSchemaID string, fieldKey string) (FieldDescriptor, error) {
	set, err := c.ResolveViewSchema(viewSchemaID)
	if err != nil {
		return FieldDescriptor{}, err
	}
	field, ok := set.fields[fieldKey]
	if !ok {
		return FieldDescriptor{}, fmt.Errorf("%w: field %q in view schema %q", ErrFieldNotFound, fieldKey, viewSchemaID)
	}
	if !field.Writable || isReadOnlySystemField(fieldKey) {
		return FieldDescriptor{}, fmt.Errorf("%w: field %q in view schema %q", ErrFieldNotWritable, fieldKey, viewSchemaID)
	}
	return field, nil
}

func (s FieldDescriptorSet) Writable(fieldKey string) bool {
	if s.fields == nil || isReadOnlySystemField(fieldKey) {
		return false
	}
	field, ok := s.fields[fieldKey]
	return ok && field.Writable
}

func (s FieldDescriptorSet) ConflictResolutionClass(fieldKey string) string {
	if s.fields == nil {
		return ""
	}
	return s.fields[fieldKey].ConflictResolutionClass
}

func isReadOnlySystemField(fieldKey string) bool {
	switch fieldKey {
	case "record_id", "row_version", "version_id", "updated_at", "created_at", "created_by_user_id", "updated_by_user_id":
		return true
	default:
		return false
	}
}
