package entities

import (
	"context"
	"strings"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
	"github.com/JochiRaider/cartulary/internal/modules/incidentportability"
	"github.com/JochiRaider/cartulary/internal/platform/fieldnorm"
)

func prepareEntityImport(ctx context.Context, bundle incidentportability.MemberReader, importContext sourceport.ImportContext) (preparedEntityImport, error) {
	prepared := preparedEntityImport{binding: entityPortableBinding{operationID: importContext.OperationID, incidentID: importContext.IncidentID, bundleVersion: importContext.BundleVersion, contractMajor: sourceport.ContractMajor}}
	if bundle == nil || importContext.OperationID == "" || importContext.IncidentID == uuid.Nil || importContext.ActorUserID == uuid.Nil || importContext.BundleVersion != 8 {
		return preparedEntityImport{}, entitySourceFailure(entitySourceIdentity)
	}
	members, err := sourceport.PrepareMembers(ctx, entitySourceDescriptor(), bundle, importContext.BundleVersion)
	if err != nil {
		return preparedEntityImport{}, entitySourceFailure(entitySourceIdentity)
	}
	prepared.members = members
	var failures entityFailureAccumulator
	for _, path := range []string{hostsBundlePath, identitiesBundlePath, entityMentionsBundlePath, preservedIDsBundlePath, entityAliasesBundlePath} {
		err := incidentportability.WalkPath(ctx, members, path, func(raw map[string]any) error {
			var rowFailures []entityFailureCandidate
			switch path {
			case hostsBundlePath:
				_, rowFailures = preparePortableHost(raw, importContext)
			case identitiesBundlePath:
				_, rowFailures = preparePortableIdentity(raw, importContext)
			case entityMentionsBundlePath:
				_, rowFailures = preparePortableMention(raw, importContext)
			case preservedIDsBundlePath:
				_, rowFailures = preparePortablePreserved(raw, importContext)
			case entityAliasesBundlePath:
				_, rowFailures = preparePortableAlias(raw, importContext)
			}
			failures.candidates(rowFailures)
			return nil
		})
		if err != nil {
			return preparedEntityImport{}, entitySourceFailure(entitySourceIdentity)
		}
	}
	if err := failures.err(); err != nil {
		return preparedEntityImport{}, err
	}
	return prepared, nil
}

func walkEntitySource[T any](ctx context.Context, prepared preparedEntityImport, input sourceport.ImportContext, path string, parse func(map[string]any, sourceport.ImportContext) (T, []entityFailureCandidate), visit func(T) error) error {
	if !prepared.binding.matches(input) || prepared.members == nil {
		return entitySourceFailure(entitySourceIdentity)
	}
	return incidentportability.WalkPath(ctx, prepared.members, path, func(raw map[string]any) error {
		row, failures := parse(raw, input)
		if err := selectedEntityFailure(failures); err != nil {
			return err
		}
		return visit(row)
	})
}
func (p preparedEntityImport) hosts(ctx context.Context, input sourceport.ImportContext, visit func(portableHostRow) error) error {
	return walkEntitySource(ctx, p, input, hostsBundlePath, preparePortableHost, visit)
}
func (p preparedEntityImport) identities(ctx context.Context, input sourceport.ImportContext, visit func(portableIdentityRow) error) error {
	return walkEntitySource(ctx, p, input, identitiesBundlePath, preparePortableIdentity, visit)
}
func (p preparedEntityImport) mentions(ctx context.Context, input sourceport.ImportContext, visit func(portableMentionRow) error) error {
	return walkEntitySource(ctx, p, input, entityMentionsBundlePath, preparePortableMention, visit)
}
func (p preparedEntityImport) preserved(ctx context.Context, input sourceport.ImportContext, visit func(portablePreservedIdentifierRow) error) error {
	return walkEntitySource(ctx, p, input, preservedIDsBundlePath, preparePortablePreserved, visit)
}
func (p preparedEntityImport) aliases(ctx context.Context, input sourceport.ImportContext, visit func(portableAliasRow) error) error {
	return walkEntitySource(ctx, p, input, entityAliasesBundlePath, preparePortableAlias, visit)
}

func preparePortableHost(raw map[string]any, importContext sourceport.ImportContext) (portableHostRow, []entityFailureCandidate) {
	identity := entityStableIdentity(raw["record_id"])
	fail := func(invariant string) (portableHostRow, []entityFailureCandidate) {
		return portableHostRow{}, []entityFailureCandidate{entityFailure(invariant, hostsBundlePath, identity, raw)}
	}
	if !exactEntityPortableMembers(raw, hostPortableMembers) {
		return fail(entitySourceIdentity)
	}
	row := portableHostRow{}
	var ok bool
	if row.RecordID, ok = entityCanonicalUUID(raw["record_id"]); !ok {
		return fail(entitySourceIdentity)
	}
	identity = row.RecordID.String()
	if row.IncidentID, ok = entityCanonicalUUID(raw["incident_id"]); !ok || row.IncidentID != importContext.IncidentID {
		return fail(entityEnvelopeTypeScope)
	}
	if row.DisplayName, ok = entityPortableText(raw["display_name"], false); !ok {
		return fail(entitySourceIdentity)
	}
	if row.Hostname, ok = entityNullableText(raw["hostname"], false); !ok {
		return fail(entitySourceIdentity)
	}
	if row.AADDeviceID, ok = entityNullableText(raw["aad_device_id"], false); !ok {
		return fail(entitySourceIdentity)
	}
	if row.FQDN, ok = entityNullableText(raw["fqdn"], false); !ok {
		return fail(entitySourceIdentity)
	}
	if row.EntityOrigin, ok = entityPortableText(raw["entity_origin"], false); !ok || !entityOriginAllowed(row.EntityOrigin) {
		return fail(entitySourceIdentity)
	}
	if row.SeedMentionID, ok = entityNullableUUID(raw["seed_entity_mention_id"]); !ok {
		return fail(entitySourceIdentity)
	}
	if row.State, ok = entityPortableText(raw["host_state"], false); !ok || !entityStateAllowed(row.State) {
		return fail(entityResolutionMerge)
	}
	if row.MergedIntoRecordID, ok = entityNullableUUID(raw["merged_into_record_id"]); !ok ||
		(row.State == "merged") != (row.MergedIntoRecordID != nil) ||
		(row.MergedIntoRecordID != nil && *row.MergedIntoRecordID == row.RecordID) {
		return fail(entityResolutionMerge)
	}
	if row.RowVersion, ok = entityCanonicalInteger(raw["row_version"]); !ok || row.RowVersion < 1 {
		return fail(entityEnvelopeTypeScope)
	}
	if row.CreatedAt, ok = entityCanonicalTimestamp(raw["created_at"]); !ok {
		return fail(entityEnvelopeTypeScope)
	}
	if row.UpdatedAt, ok = entityCanonicalTimestamp(raw["updated_at"]); !ok || row.UpdatedAt.Before(row.CreatedAt) {
		return fail(entityEnvelopeTypeScope)
	}
	if row.PortableCreatedByID, ok = entityAdmittedActor(raw["created_by_user_id"], importContext); !ok {
		return fail(entityEnvelopeTypeScope)
	}
	if row.PortableUpdatedByID, ok = entityAdmittedActor(raw["updated_by_user_id"], importContext); !ok {
		return fail(entityEnvelopeTypeScope)
	}
	row.RuntimeCreatedByID, row.RuntimeUpdatedByID = importContext.ActorUserID, importContext.ActorUserID
	if row.Location, ok = entityNullableText(raw["location"], true); !ok {
		return fail(entitySourceIdentity)
	}
	if row.OSPlatform, ok = entityNullableText(raw["os_platform"], true); !ok {
		return fail(entitySourceIdentity)
	}
	if row.BusinessOwner, ok = entityNullableText(raw["business_owner"], true); !ok {
		return fail(entitySourceIdentity)
	}
	if row.Criticality, ok = entityNullableText(raw["criticality"], true); !ok {
		return fail(entitySourceIdentity)
	}
	if row.ContainmentStatus, ok = entityNullableText(raw["containment_status"], true); !ok {
		return fail(entitySourceIdentity)
	}
	return row, nil
}

func preparePortableIdentity(raw map[string]any, importContext sourceport.ImportContext) (portableIdentityRow, []entityFailureCandidate) {
	identity := entityStableIdentity(raw["record_id"])
	fail := func(invariant string) (portableIdentityRow, []entityFailureCandidate) {
		return portableIdentityRow{}, []entityFailureCandidate{entityFailure(invariant, identitiesBundlePath, identity, raw)}
	}
	if !exactEntityPortableMembers(raw, identityPortableMembers) {
		return fail(entitySourceIdentity)
	}
	row := portableIdentityRow{}
	var ok bool
	if row.RecordID, ok = entityCanonicalUUID(raw["record_id"]); !ok {
		return fail(entitySourceIdentity)
	}
	identity = row.RecordID.String()
	if row.IncidentID, ok = entityCanonicalUUID(raw["incident_id"]); !ok || row.IncidentID != importContext.IncidentID {
		return fail(entityEnvelopeTypeScope)
	}
	if row.DisplayName, ok = entityPortableText(raw["display_name"], false); !ok {
		return fail(entitySourceIdentity)
	}
	for source, destination := range map[string]**string{
		"upn": &row.UPN, "email": &row.Email, "sam_account_name": &row.SAMAccountName,
		"aad_object_id": &row.AADObjectID, "sid": &row.SID, "privilege_level": &row.PrivilegeLevel,
		"mfa_state": &row.MFAState, "reset_status": &row.ResetStatus,
	} {
		allowEmpty := source == "privilege_level" || source == "mfa_state" || source == "reset_status"
		if *destination, ok = entityNullableText(raw[source], allowEmpty); !ok {
			return fail(entitySourceIdentity)
		}
	}
	if row.EntityOrigin, ok = entityPortableText(raw["entity_origin"], false); !ok || !entityOriginAllowed(row.EntityOrigin) {
		return fail(entitySourceIdentity)
	}
	if row.SeedMentionID, ok = entityNullableUUID(raw["seed_entity_mention_id"]); !ok {
		return fail(entitySourceIdentity)
	}
	if row.State, ok = entityPortableText(raw["identity_state"], false); !ok || !entityStateAllowed(row.State) {
		return fail(entityResolutionMerge)
	}
	if row.MergedIntoRecordID, ok = entityNullableUUID(raw["merged_into_record_id"]); !ok ||
		(row.State == "merged") != (row.MergedIntoRecordID != nil) ||
		(row.MergedIntoRecordID != nil && *row.MergedIntoRecordID == row.RecordID) {
		return fail(entityResolutionMerge)
	}
	if row.RowVersion, ok = entityCanonicalInteger(raw["row_version"]); !ok || row.RowVersion < 1 {
		return fail(entityEnvelopeTypeScope)
	}
	if row.CreatedAt, ok = entityCanonicalTimestamp(raw["created_at"]); !ok {
		return fail(entityEnvelopeTypeScope)
	}
	if row.UpdatedAt, ok = entityCanonicalTimestamp(raw["updated_at"]); !ok || row.UpdatedAt.Before(row.CreatedAt) {
		return fail(entityEnvelopeTypeScope)
	}
	if row.PortableCreatedByID, ok = entityAdmittedActor(raw["created_by_user_id"], importContext); !ok {
		return fail(entityEnvelopeTypeScope)
	}
	if row.PortableUpdatedByID, ok = entityAdmittedActor(raw["updated_by_user_id"], importContext); !ok {
		return fail(entityEnvelopeTypeScope)
	}
	row.RuntimeCreatedByID, row.RuntimeUpdatedByID = importContext.ActorUserID, importContext.ActorUserID
	return row, nil
}

func preparePortableMention(raw map[string]any, importContext sourceport.ImportContext) (portableMentionRow, []entityFailureCandidate) {
	identity := entityStableIdentity(raw["entity_mention_id"])
	fail := func(invariant string) (portableMentionRow, []entityFailureCandidate) {
		return portableMentionRow{}, []entityFailureCandidate{entityFailure(invariant, entityMentionsBundlePath, identity, raw)}
	}
	if !exactEntityPortableMembers(raw, mentionPortableMembers) {
		return fail(entitySourceIdentity)
	}
	row := portableMentionRow{}
	var ok bool
	if row.MentionID, ok = entityCanonicalUUID(raw["entity_mention_id"]); !ok {
		return fail(entitySourceIdentity)
	}
	identity = row.MentionID.String()
	if row.SourceRecordID, ok = entityCanonicalUUID(raw["source_record_id"]); !ok {
		return fail(entityMentionsObserved)
	}
	for source, destination := range map[string]*string{
		"entity_type": &row.EntityType, "source_field_key": &row.SourceFieldKey, "origin_kind": &row.OriginKind,
		"origin_locator": &row.OriginLocator, "raw_text": &row.RawText, "normalized_text": &row.NormalizedText,
		"resolution_status": &row.ResolutionStatus,
	} {
		if *destination, ok = entityPortableText(raw[source], false); !ok {
			return fail(entityMentionsObserved)
		}
	}
	if !entityTypeAllowed(row.EntityType) || !mentionOriginAllowed(row.OriginKind) || !resolutionStatusAllowed(row.ResolutionStatus) ||
		strings.TrimSpace(row.SourceFieldKey) == "" || strings.TrimSpace(row.OriginLocator) == "" ||
		strings.TrimSpace(row.RawText) == "" || strings.TrimSpace(row.NormalizedText) == "" {
		return fail(entityMentionsObserved)
	}
	if row.RowVersion, ok = entityCanonicalInteger(raw["row_version"]); !ok || row.RowVersion < 1 {
		return fail(entityMentionsObserved)
	}
	ordinal, ok := entityCanonicalInteger(raw["ordinal"])
	if !ok || ordinal < 1 || ordinal > int64(^uint32(0)>>1) {
		return fail(entityMentionsObserved)
	}
	row.Ordinal = int32(ordinal)
	if row.PortableCreatedByID, ok = entityAdmittedActor(raw["created_by_user_id"], importContext); !ok {
		return fail(entityMentionsObserved)
	}
	row.RuntimeCreatedByID = importContext.ActorUserID
	if row.CreatedAt, ok = entityCanonicalTimestamp(raw["created_at"]); !ok {
		return fail(entityMentionsObserved)
	}
	if row.ResolvedRecordID, ok = entityNullableUUID(raw["resolved_record_id"]); !ok {
		return fail(entityResolutionMerge)
	}
	if row.PortableResolvedByID, ok = entityNullableAdmittedActor(raw["resolved_by_user_id"], importContext); !ok {
		return fail(entityResolutionMerge)
	}
	row.RuntimeResolvedByID = entityRuntimeActorPointer(row.PortableResolvedByID, importContext.ActorUserID)
	if row.ResolvedAt, ok = entityNullableTimestamp(raw["resolved_at"]); !ok {
		return fail(entityResolutionMerge)
	}
	if row.ResolutionMethod, ok = entityNullableText(raw["resolution_method"], false); !ok {
		return fail(entityResolutionMerge)
	}
	if !mentionResolutionTupleValid(row) {
		return fail(entityResolutionMerge)
	}
	return row, nil
}

func preparePortablePreserved(raw map[string]any, importContext sourceport.ImportContext) (portablePreservedIdentifierRow, []entityFailureCandidate) {
	identity := entityStableIdentity(raw["entity_preserved_identifier_id"])
	fail := func(invariant string) (portablePreservedIdentifierRow, []entityFailureCandidate) {
		return portablePreservedIdentifierRow{}, []entityFailureCandidate{entityFailure(invariant, preservedIDsBundlePath, identity, raw)}
	}
	if !exactEntityPortableMembers(raw, preservedPortableMembers) {
		return fail(entitySourceIdentity)
	}
	row := portablePreservedIdentifierRow{}
	var ok bool
	if row.PreservedID, ok = entityCanonicalUUID(raw["entity_preserved_identifier_id"]); !ok {
		return fail(entitySourceIdentity)
	}
	identity = row.PreservedID.String()
	if row.IncidentID, ok = entityCanonicalUUID(raw["incident_id"]); !ok || row.IncidentID != importContext.IncidentID {
		return fail(entitySameIncident)
	}
	if row.RecordID, ok = entityCanonicalUUID(raw["record_id"]); !ok {
		return fail(entitySameIncident)
	}
	for source, destination := range map[string]*string{
		"entity_type": &row.EntityType, "identifier_type": &row.IdentifierType, "raw_value": &row.RawValue,
		"normalized_value": &row.NormalizedValue, "classification": &row.Classification,
	} {
		if *destination, ok = entityPortableText(raw[source], false); !ok {
			return fail(entityClassified)
		}
	}
	if !preservedClassAllowed(row.EntityType, row.IdentifierType, row.Classification) {
		return fail(entityClassified)
	}
	want, normalized := fieldnorm.NormalizeIdentifier(row.IdentifierType, row.RawValue)
	if !normalized || want != row.NormalizedValue {
		return fail(entityNormalized)
	}
	if row.PortableCreatedByID, ok = entityAdmittedActor(raw["created_by_user_id"], importContext); !ok {
		return fail(entitySourceIdentity)
	}
	row.RuntimeCreatedByID = importContext.ActorUserID
	if row.CreatedAt, ok = entityCanonicalTimestamp(raw["created_at"]); !ok {
		return fail(entitySourceIdentity)
	}
	if row.DeletedAt, ok = entityNullableTimestamp(raw["deleted_at"]); !ok || (row.DeletedAt != nil && row.DeletedAt.Before(row.CreatedAt)) {
		return fail(entitySourceIdentity)
	}
	return row, nil
}

func preparePortableAlias(raw map[string]any, importContext sourceport.ImportContext) (portableAliasRow, []entityFailureCandidate) {
	identity := entityStableIdentity(raw["entity_alias_id"])
	fail := func(invariant string) (portableAliasRow, []entityFailureCandidate) {
		return portableAliasRow{}, []entityFailureCandidate{entityFailure(invariant, entityAliasesBundlePath, identity, raw)}
	}
	if !exactEntityPortableMembers(raw, aliasPortableMembers) {
		return fail(entitySourceIdentity)
	}
	row := portableAliasRow{}
	var ok bool
	if row.AliasID, ok = entityCanonicalUUID(raw["entity_alias_id"]); !ok {
		return fail(entitySourceIdentity)
	}
	identity = row.AliasID.String()
	if row.IncidentID, ok = entityCanonicalUUID(raw["incident_id"]); !ok || row.IncidentID != importContext.IncidentID {
		return fail(entitySameIncident)
	}
	if row.RecordID, ok = entityCanonicalUUID(raw["record_id"]); !ok {
		return fail(entitySameIncident)
	}
	if row.EntityType, ok = entityPortableText(raw["entity_type"], false); !ok {
		return fail(entityClassified)
	}
	if row.RawText, ok = entityPortableText(raw["raw_text"], false); !ok {
		return fail(entityNormalized)
	}
	if row.NormalizedText, ok = entityPortableText(raw["normalized_text"], false); !ok {
		return fail(entityNormalized)
	}
	if row.Classification, ok = entityPortableText(raw["classification"], false); !ok ||
		!entityTypeAllowed(row.EntityType) || row.Classification != "suggestion_only" {
		return fail(entityClassified)
	}
	want, normalized := fieldnorm.NormalizeEntityAliasText(row.RawText)
	if !normalized || want != row.RawText || want != row.NormalizedText {
		return fail(entityNormalized)
	}
	if row.PortableCreatedByID, ok = entityAdmittedActor(raw["created_by_user_id"], importContext); !ok {
		return fail(entitySourceIdentity)
	}
	row.RuntimeCreatedByID = importContext.ActorUserID
	if row.CreatedAt, ok = entityCanonicalTimestamp(raw["created_at"]); !ok {
		return fail(entitySourceIdentity)
	}
	if row.DeletedAt, ok = entityNullableTimestamp(raw["deleted_at"]); !ok || (row.DeletedAt != nil && row.DeletedAt.Before(row.CreatedAt)) {
		return fail(entitySourceIdentity)
	}
	return row, nil
}

func entityStableIdentity(value any) string {
	parsed, ok := entityCanonicalUUID(value)
	if !ok {
		return ""
	}
	return parsed.String()
}

func entityOriginAllowed(value string) bool {
	switch value {
	case "entity_sheet", "entity_import", "created_from_mention", "system_upsert":
		return true
	default:
		return false
	}
}

func entityStateAllowed(value string) bool {
	return value == "stub" || value == "canonical" || value == "merged"
}

func entityTypeAllowed(value string) bool { return value == "host" || value == "identity" }

func mentionOriginAllowed(value string) bool {
	switch value {
	case "manual_entry", "clipboard_paste", "csv_import", "xlsx_import", "api_import", "extraction", "system":
		return true
	default:
		return false
	}
}

func resolutionStatusAllowed(value string) bool {
	return value == "unresolved" || value == "resolved" || value == "dismissed"
}

func mentionResolutionTupleValid(row portableMentionRow) bool {
	if row.ResolutionStatus == "resolved" {
		return row.ResolvedRecordID != nil && row.PortableResolvedByID != nil && row.ResolvedAt != nil && row.ResolutionMethod != nil
	}
	return row.ResolvedRecordID == nil && row.PortableResolvedByID == nil && row.ResolvedAt == nil && row.ResolutionMethod == nil
}

func preservedClassAllowed(entityType, identifierType, classification string) bool {
	if classification != "exact_match_reuse" && classification != "suggestion_only" && classification != "provenance_only" {
		return false
	}
	switch entityType {
	case "host":
		return identifierType == "aad_device_id" || identifierType == "fqdn" || identifierType == "hostname"
	case "identity":
		return identifierType == "aad_object_id" || identifierType == "sid" || identifierType == "upn" || identifierType == "email" || identifierType == "sam_account_name"
	default:
		return false
	}
}
