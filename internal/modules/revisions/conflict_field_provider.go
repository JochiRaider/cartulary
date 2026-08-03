package revisions

import (
	"fmt"

	"github.com/JochiRaider/cartulary/internal/modules/revisions/conflicts"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
)

// ConflictFieldProvider is constructed by a source owner and consumed once by
// revision assembly. The immutable Revisions catalog retains only copied field
// contracts, never this provider or the process-global view-schema registry.
type ConflictFieldProvider interface {
	ConflictFields(viewSchemaID string) ([]conflicts.FieldDescriptor, error)
}

type viewSchemaConflictFieldProvider struct{}

func NewViewSchemaConflictFieldProvider() ConflictFieldProvider {
	return viewSchemaConflictFieldProvider{}
}

func (viewSchemaConflictFieldProvider) ConflictFields(viewSchemaID string) ([]conflicts.FieldDescriptor, error) {
	schema, ok := viewschema.Lookup(viewSchemaID)
	if !ok {
		return nil, fmt.Errorf("Revisions conflict field provider: unknown view schema %q", viewSchemaID)
	}
	fields := schema.Fields()
	descriptors := make([]conflicts.FieldDescriptor, 0, len(fields))
	for _, field := range fields {
		descriptors = append(descriptors, conflicts.FieldDescriptor{
			FieldKey:                field.FieldKey,
			ValueKind:               field.WriteKind,
			Writable:                field.Writable,
			ConflictResolutionClass: field.ConflictResolutionClass,
		})
	}
	return descriptors, nil
}
