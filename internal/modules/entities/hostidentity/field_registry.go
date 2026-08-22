package hostidentity

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
)

type entityFieldPatchStrategy string

const (
	entityFieldPatchNone       entityFieldPatchStrategy = "none"
	entityFieldPatchDirect     entityFieldPatchStrategy = "direct"
	entityFieldPatchCollection entityFieldPatchStrategy = "collection"
)

type entityFieldRegistryKey struct {
	viewSchemaID string
	fieldKey     string
}

type hostFieldMaterializer func(HostRecord) any
type identityFieldMaterializer func(IdentityRecord) any
type hostFieldPatchApplier func(*HostRecord, *string) (bool, error)
type identityFieldPatchApplier func(*IdentityRecord, *string) (bool, error)
type collectionFieldPatchApplier func(
	context.Context,
	pgx.Tx,
	uuid.UUID,
	uuid.UUID,
	[]CollectionAction,
	uuid.UUID,
	time.Time,
) ([]AliasAppliedMutation, error)

type entityFieldDescriptor struct {
	viewSchemaID string
	fieldKey     string
	patch        entityFieldPatchStrategy

	materializeHost     hostFieldMaterializer
	materializeIdentity identityFieldMaterializer
	patchHost           hostFieldPatchApplier
	patchIdentity       identityFieldPatchApplier
	patchCollection     collectionFieldPatchApplier

	owner viewschema.Field
}

func (descriptor entityFieldDescriptor) participatesInCreate() bool {
	return descriptor.owner.Writable || descriptor.owner.CreateWritable
}

func (descriptor entityFieldDescriptor) clearable() bool {
	return descriptor.owner.Clearable
}

func (descriptor entityFieldDescriptor) supportsGrouping() bool {
	return descriptor.owner.Groupable
}

type entityFieldRegistry struct {
	byKey   map[entityFieldRegistryKey]entityFieldDescriptor
	ordered map[string][]entityFieldDescriptor
}

var entityFields = mustNewEntityFieldRegistry(entityFieldDescriptors())

func mustNewEntityFieldRegistry(descriptors []entityFieldDescriptor) entityFieldRegistry {
	registry, err := newEntityFieldRegistry(descriptors)
	if err != nil {
		panic("hostidentity: build entity field registry: " + err.Error())
	}
	return registry
}

func newEntityFieldRegistry(descriptors []entityFieldDescriptor) (entityFieldRegistry, error) {
	registry := entityFieldRegistry{
		byKey:   make(map[entityFieldRegistryKey]entityFieldDescriptor, len(descriptors)),
		ordered: make(map[string][]entityFieldDescriptor, 2),
	}
	for index, descriptor := range descriptors {
		if descriptor.viewSchemaID != HostsViewSchemaID && descriptor.viewSchemaID != IdentitiesViewSchemaID {
			return entityFieldRegistry{}, fmt.Errorf("descriptor %d has unknown view schema %q", index, descriptor.viewSchemaID)
		}
		field, ok := viewschema.LookupField(descriptor.viewSchemaID, descriptor.fieldKey)
		if !ok {
			return entityFieldRegistry{}, fmt.Errorf("descriptor %d has unknown field %s/%s", index, descriptor.viewSchemaID, descriptor.fieldKey)
		}
		key := entityFieldRegistryKey{viewSchemaID: descriptor.viewSchemaID, fieldKey: descriptor.fieldKey}
		if _, duplicate := registry.byKey[key]; duplicate {
			return entityFieldRegistry{}, fmt.Errorf("duplicate descriptor %s/%s", descriptor.viewSchemaID, descriptor.fieldKey)
		}
		if err := validateEntityFieldDescriptor(descriptor, field); err != nil {
			return entityFieldRegistry{}, fmt.Errorf("descriptor %s/%s: %w", descriptor.viewSchemaID, descriptor.fieldKey, err)
		}
		descriptor.owner = field
		registry.byKey[key] = descriptor
		registry.ordered[descriptor.viewSchemaID] = append(registry.ordered[descriptor.viewSchemaID], descriptor)
	}

	for _, viewSchemaID := range []string{HostsViewSchemaID, IdentitiesViewSchemaID} {
		resource, ok := viewschema.LookupPublicResource(viewSchemaID)
		if !ok {
			return entityFieldRegistry{}, fmt.Errorf("owner projection %s is missing", viewSchemaID)
		}
		actual := registry.ordered[viewSchemaID]
		if len(actual) != len(resource.Fields) {
			return entityFieldRegistry{}, fmt.Errorf("%s descriptor count %d does not match owner field count %d", viewSchemaID, len(actual), len(resource.Fields))
		}
		for index, ownerField := range resource.Fields {
			if actual[index].fieldKey != ownerField.FieldKey {
				return entityFieldRegistry{}, fmt.Errorf("%s descriptor order mismatch at %d: got %s, want %s", viewSchemaID, index, actual[index].fieldKey, ownerField.FieldKey)
			}
		}
	}
	return registry, nil
}

func validateEntityFieldDescriptor(descriptor entityFieldDescriptor, owner viewschema.Field) error {
	if owner.FieldKey != descriptor.fieldKey {
		return fmt.Errorf("owner field key is %q", owner.FieldKey)
	}
	if owner.Writable && owner.WriteKind != "direct_value" && owner.WriteKind != "action_payload" {
		return fmt.Errorf("writable owner write kind %q is unsupported", owner.WriteKind)
	}
	expectedPatch := entityPatchStrategyForOwner(owner)
	if descriptor.patch != expectedPatch {
		return fmt.Errorf("patch strategy %q does not match owner strategy %q", descriptor.patch, expectedPatch)
	}

	host := descriptor.viewSchemaID == HostsViewSchemaID
	if host {
		if descriptor.materializeHost == nil || descriptor.materializeIdentity != nil {
			return fmt.Errorf("host descriptor must have exactly one host materializer")
		}
	} else if descriptor.materializeIdentity == nil || descriptor.materializeHost != nil {
		return fmt.Errorf("identity descriptor must have exactly one identity materializer")
	}

	switch descriptor.patch {
	case entityFieldPatchNone:
		if descriptor.patchHost != nil || descriptor.patchIdentity != nil || descriptor.patchCollection != nil {
			return fmt.Errorf("nonpatchable descriptor has a patch applier")
		}
	case entityFieldPatchDirect:
		if host && (descriptor.patchHost == nil || descriptor.patchIdentity != nil) {
			return fmt.Errorf("direct Host descriptor must have exactly one Host patch applier")
		}
		if !host && (descriptor.patchIdentity == nil || descriptor.patchHost != nil) {
			return fmt.Errorf("direct Identity descriptor must have exactly one Identity patch applier")
		}
		if descriptor.patchCollection != nil {
			return fmt.Errorf("direct descriptor has a collection patch applier")
		}
	case entityFieldPatchCollection:
		if descriptor.patchHost != nil || descriptor.patchIdentity != nil || descriptor.patchCollection == nil {
			return fmt.Errorf("collection descriptor must have exactly one collection patch applier")
		}
	default:
		return fmt.Errorf("unknown patch strategy %q", descriptor.patch)
	}

	if owner.Writable || owner.CreateWritable {
		if owner.EntityBindingMode == nil || *owner.EntityBindingMode != "entity_origin" {
			return fmt.Errorf("write-capable owner field must bind to entity_origin")
		}
		expectedConflict := "atomic_replace"
		if owner.WriteKind == "action_payload" {
			expectedConflict = "collection_review"
		}
		if owner.ConflictResolutionClass != expectedConflict {
			return fmt.Errorf("owner conflict class %q does not match %q", owner.ConflictResolutionClass, expectedConflict)
		}
	}
	return nil
}

func (registry entityFieldRegistry) lookup(viewSchemaID string, fieldKey string) (entityFieldDescriptor, bool) {
	descriptor, ok := registry.byKey[entityFieldRegistryKey{viewSchemaID: viewSchemaID, fieldKey: fieldKey}]
	return descriptor, ok
}

func (registry entityFieldRegistry) descriptors(viewSchemaID string) []entityFieldDescriptor {
	return append([]entityFieldDescriptor(nil), registry.ordered[viewSchemaID]...)
}

func (registry entityFieldRegistry) buildHostRow(record HostRecord) map[string]any {
	cells := make(map[string]any, len(registry.ordered[HostsViewSchemaID]))
	groupValues := make(map[string]any)
	for _, descriptor := range registry.ordered[HostsViewSchemaID] {
		value := descriptor.materializeHost(record)
		cells[descriptor.fieldKey] = map[string]any{"value": value}
		if descriptor.supportsGrouping() {
			groupValues[descriptor.fieldKey] = value
		}
	}
	return map[string]any{
		"record_id":    record.RecordID.String(),
		"row_version":  record.RowVersion,
		"cells":        cells,
		"group_values": groupValues,
	}
}

func (registry entityFieldRegistry) buildIdentityRow(record IdentityRecord) map[string]any {
	cells := make(map[string]any, len(registry.ordered[IdentitiesViewSchemaID]))
	groupValues := make(map[string]any)
	for _, descriptor := range registry.ordered[IdentitiesViewSchemaID] {
		value := descriptor.materializeIdentity(record)
		cells[descriptor.fieldKey] = map[string]any{"value": value}
		if descriptor.supportsGrouping() {
			groupValues[descriptor.fieldKey] = value
		}
	}
	return map[string]any{
		"record_id":    record.RecordID.String(),
		"row_version":  record.RowVersion,
		"cells":        cells,
		"group_values": groupValues,
	}
}

func (registry entityFieldRegistry) applyHostPatch(
	ctx context.Context,
	tx pgx.Tx,
	incidentID uuid.UUID,
	recordID uuid.UUID,
	actorID uuid.UUID,
	now time.Time,
	record *HostRecord,
	change PatchChange,
) (bool, []AliasAppliedMutation, error) {
	descriptor, ok := registry.lookup(HostsViewSchemaID, change.FieldKey)
	if !ok {
		return false, nil, ErrNoEffectivePatchChange
	}
	switch descriptor.patch {
	case entityFieldPatchDirect:
		if change.CollectionActions != nil {
			return false, nil, ErrInvalidAliasReference
		}
		changed, err := descriptor.patchHost(record, change.Value)
		return changed, nil, err
	case entityFieldPatchCollection:
		if change.CollectionActions == nil || change.Value != nil {
			return false, nil, ErrNoEffectivePatchChange
		}
		mutations, err := descriptor.patchCollection(ctx, tx, incidentID, recordID, change.CollectionActions, actorID, now)
		return len(mutations) > 0, mutations, err
	default:
		return false, nil, ErrNoEffectivePatchChange
	}
}

func (registry entityFieldRegistry) applyIdentityPatch(
	ctx context.Context,
	tx pgx.Tx,
	incidentID uuid.UUID,
	recordID uuid.UUID,
	actorID uuid.UUID,
	now time.Time,
	record *IdentityRecord,
	change PatchChange,
) (bool, []AliasAppliedMutation, error) {
	descriptor, ok := registry.lookup(IdentitiesViewSchemaID, change.FieldKey)
	if !ok {
		return false, nil, ErrNoEffectivePatchChange
	}
	switch descriptor.patch {
	case entityFieldPatchDirect:
		if change.CollectionActions != nil {
			return false, nil, ErrInvalidAliasReference
		}
		changed, err := descriptor.patchIdentity(record, change.Value)
		return changed, nil, err
	case entityFieldPatchCollection:
		if change.CollectionActions == nil || change.Value != nil {
			return false, nil, ErrNoEffectivePatchChange
		}
		mutations, err := descriptor.patchCollection(ctx, tx, incidentID, recordID, change.CollectionActions, actorID, now)
		return len(mutations) > 0, mutations, err
	default:
		return false, nil, ErrNoEffectivePatchChange
	}
}

func entityFieldDescriptors() []entityFieldDescriptor {
	aliasPatch := func(entityType string) collectionFieldPatchApplier {
		return func(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, recordID uuid.UUID, actions []CollectionAction, actorID uuid.UUID, now time.Time) ([]AliasAppliedMutation, error) {
			return applyEntityAliasActionsTx(ctx, tx, incidentID, recordID, entityType, actions, actorID, now)
		}
	}
	return []entityFieldDescriptor{
		hostField("host.display_name", entityFieldPatchDirect, func(record HostRecord) any { return record.DisplayName }, patchRequiredHostString(func(record *HostRecord) *string { return &record.DisplayName })),
		hostField("host.hostname", entityFieldPatchDirect, func(record HostRecord) any { return derefString(record.Hostname) }, patchOptionalHostString(func(record *HostRecord) **string { return &record.Hostname })),
		hostField("host.aad_device_id", entityFieldPatchNone, func(record HostRecord) any { return derefString(record.AADDeviceID) }, nil),
		hostField("host.fqdn", entityFieldPatchNone, func(record HostRecord) any { return derefString(record.FQDN) }, nil),
		hostCollectionField("host.aliases", func(record HostRecord) any {
			return collectionValue(false, aliasCollectionItems(record.SuggestionOnlyAliases))
		}, aliasPatch("host")),
		hostField("host.reusable_identifiers", entityFieldPatchNone, func(record HostRecord) any {
			return collectionValue(false, hostReusableIdentifierCollectionItems(record))
		}, nil),
		hostField("host.host_state", entityFieldPatchNone, func(record HostRecord) any { return record.HostState }, nil),
		hostField("host.linked_event_count", entityFieldPatchNone, func(record HostRecord) any { return record.LinkedEventCount }, nil),
		hostField("host.evidence_count", entityFieldPatchNone, func(record HostRecord) any { return record.EvidenceCount }, nil),
		hostField("host.location", entityFieldPatchDirect, func(record HostRecord) any { return derefString(record.Location) }, patchOptionalHostString(func(record *HostRecord) **string { return &record.Location })),
		hostField("host.os_platform", entityFieldPatchDirect, func(record HostRecord) any { return derefString(record.OSPlatform) }, patchOptionalHostString(func(record *HostRecord) **string { return &record.OSPlatform })),
		hostField("host.business_owner", entityFieldPatchDirect, func(record HostRecord) any { return derefString(record.BusinessOwner) }, patchOptionalHostString(func(record *HostRecord) **string { return &record.BusinessOwner })),
		hostField("host.criticality", entityFieldPatchDirect, func(record HostRecord) any { return derefString(record.Criticality) }, patchOptionalHostString(func(record *HostRecord) **string { return &record.Criticality })),
		hostField("host.containment_status", entityFieldPatchDirect, func(record HostRecord) any { return derefString(record.ContainmentStatus) }, patchOptionalHostString(func(record *HostRecord) **string { return &record.ContainmentStatus })),
		hostField("host.edited_at", entityFieldPatchNone, func(record HostRecord) any { return formatTimestamp(record.UpdatedAt) }, nil),

		identityField("identity.display_name", entityFieldPatchDirect, func(record IdentityRecord) any { return record.DisplayName }, patchRequiredIdentityString(func(record *IdentityRecord) *string { return &record.DisplayName })),
		identityField("identity.aad_object_id", entityFieldPatchNone, func(record IdentityRecord) any { return derefString(record.AADObjectID) }, nil),
		identityField("identity.sid", entityFieldPatchNone, func(record IdentityRecord) any { return derefString(record.SID) }, nil),
		identityField("identity.upn", entityFieldPatchDirect, func(record IdentityRecord) any { return derefString(record.UPN) }, patchOptionalIdentityString(func(record *IdentityRecord) **string { return &record.UPN })),
		identityField("identity.email", entityFieldPatchDirect, func(record IdentityRecord) any { return derefString(record.Email) }, patchOptionalIdentityString(func(record *IdentityRecord) **string { return &record.Email })),
		identityField("identity.sam_account_name", entityFieldPatchDirect, func(record IdentityRecord) any { return derefString(record.SamAccountName) }, patchOptionalIdentityString(func(record *IdentityRecord) **string { return &record.SamAccountName })),
		identityCollectionField("identity.aliases", func(record IdentityRecord) any {
			return collectionValue(false, aliasCollectionItems(record.SuggestionOnlyAliases))
		}, aliasPatch("identity")),
		identityField("identity.reusable_identifiers", entityFieldPatchNone, func(record IdentityRecord) any {
			return collectionValue(false, identityReusableIdentifierCollectionItems(record))
		}, nil),
		identityField("identity.identity_state", entityFieldPatchNone, func(record IdentityRecord) any { return record.IdentityState }, nil),
		identityField("identity.linked_event_count", entityFieldPatchNone, func(record IdentityRecord) any { return record.LinkedEventCount }, nil),
		identityField("identity.evidence_count", entityFieldPatchNone, func(record IdentityRecord) any { return record.EvidenceCount }, nil),
		identityField("identity.privilege_level", entityFieldPatchDirect, func(record IdentityRecord) any { return derefString(record.PrivilegeLevel) }, patchOptionalIdentityString(func(record *IdentityRecord) **string { return &record.PrivilegeLevel })),
		identityField("identity.mfa_state", entityFieldPatchDirect, func(record IdentityRecord) any { return derefString(record.MFAState) }, patchOptionalIdentityString(func(record *IdentityRecord) **string { return &record.MFAState })),
		identityField("identity.reset_status", entityFieldPatchDirect, func(record IdentityRecord) any { return derefString(record.ResetStatus) }, patchOptionalIdentityString(func(record *IdentityRecord) **string { return &record.ResetStatus })),
		identityField("identity.edited_at", entityFieldPatchNone, func(record IdentityRecord) any { return formatTimestamp(record.UpdatedAt) }, nil),
	}
}

func hostField(fieldKey string, patch entityFieldPatchStrategy, materialize hostFieldMaterializer, apply hostFieldPatchApplier) entityFieldDescriptor {
	return entityFieldDescriptor{viewSchemaID: HostsViewSchemaID, fieldKey: fieldKey, patch: patch, materializeHost: materialize, patchHost: apply}
}

func identityField(fieldKey string, patch entityFieldPatchStrategy, materialize identityFieldMaterializer, apply identityFieldPatchApplier) entityFieldDescriptor {
	return entityFieldDescriptor{viewSchemaID: IdentitiesViewSchemaID, fieldKey: fieldKey, patch: patch, materializeIdentity: materialize, patchIdentity: apply}
}

func hostCollectionField(fieldKey string, materialize hostFieldMaterializer, apply collectionFieldPatchApplier) entityFieldDescriptor {
	return entityFieldDescriptor{viewSchemaID: HostsViewSchemaID, fieldKey: fieldKey, patch: entityFieldPatchCollection, materializeHost: materialize, patchCollection: apply}
}

func identityCollectionField(fieldKey string, materialize identityFieldMaterializer, apply collectionFieldPatchApplier) entityFieldDescriptor {
	return entityFieldDescriptor{viewSchemaID: IdentitiesViewSchemaID, fieldKey: fieldKey, patch: entityFieldPatchCollection, materializeIdentity: materialize, patchCollection: apply}
}

func patchRequiredHostString(target func(*HostRecord) *string) hostFieldPatchApplier {
	return func(record *HostRecord, value *string) (bool, error) {
		if value == nil {
			return false, ErrNoEffectivePatchChange
		}
		current := target(record)
		if *current == *value {
			return false, nil
		}
		*current = *value
		return true, nil
	}
}

func patchOptionalHostString(target func(*HostRecord) **string) hostFieldPatchApplier {
	return func(record *HostRecord, value *string) (bool, error) {
		current := target(record)
		if stringPointersEqual(*current, value) {
			return false, nil
		}
		*current = cloneStringPointer(value)
		return true, nil
	}
}

func patchRequiredIdentityString(target func(*IdentityRecord) *string) identityFieldPatchApplier {
	return func(record *IdentityRecord, value *string) (bool, error) {
		if value == nil {
			return false, ErrNoEffectivePatchChange
		}
		current := target(record)
		if *current == *value {
			return false, nil
		}
		*current = *value
		return true, nil
	}
}

func patchOptionalIdentityString(target func(*IdentityRecord) **string) identityFieldPatchApplier {
	return func(record *IdentityRecord, value *string) (bool, error) {
		current := target(record)
		if stringPointersEqual(*current, value) {
			return false, nil
		}
		*current = cloneStringPointer(value)
		return true, nil
	}
}

func entityPatchStrategyForOwner(field viewschema.Field) entityFieldPatchStrategy {
	if !field.Writable {
		return entityFieldPatchNone
	}
	if field.WriteKind == "action_payload" {
		return entityFieldPatchCollection
	}
	return entityFieldPatchDirect
}
