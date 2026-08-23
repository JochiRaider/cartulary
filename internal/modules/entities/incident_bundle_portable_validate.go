package entities

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
)

func validatePreparedEntityImportTx(ctx context.Context, tx pgx.Tx, prepared preparedEntityImport, importContext sourceport.ImportContext) error {
	if tx == nil || !prepared.binding.matches(importContext) || importContext.Attributions == nil {
		return entitySourceFailure(entitySourceIdentity)
	}
	actualHosts, err := loadPortableHosts(ctx, sourceport.ExportContext{Query: tx, IncidentID: importContext.IncidentID})
	if err != nil {
		return entitySourceFailure(entityEnvelopeTypeScope)
	}
	actualIdentities, err := loadPortableIdentities(ctx, sourceport.ExportContext{Query: tx, IncidentID: importContext.IncidentID})
	if err != nil {
		return entitySourceFailure(entityEnvelopeTypeScope)
	}
	actualMentions, err := loadPortableEntityMentions(ctx, sourceport.ExportContext{Query: tx, IncidentID: importContext.IncidentID})
	if err != nil {
		return entitySourceFailure(entityResolutionMerge)
	}
	actualPreserved, err := loadPortablePreservedIdentifiers(ctx, sourceport.ExportContext{Query: tx, IncidentID: importContext.IncidentID})
	if err != nil {
		return entitySourceFailure(entitySameIncident)
	}
	actualAliases, err := loadPortableAliases(ctx, sourceport.ExportContext{Query: tx, IncidentID: importContext.IncidentID})
	if err != nil {
		return entitySourceFailure(entitySameIncident)
	}

	var failures []entityFailureCandidate
	comparePortableHosts(prepared.hosts, actualHosts, &failures)
	comparePortableIdentities(prepared.identities, actualIdentities, &failures)
	comparePortableMentions(prepared.mentions, actualMentions, &failures)
	comparePortablePreserved(prepared.preserved, actualPreserved, &failures)
	comparePortableAliases(prepared.aliases, actualAliases, &failures)
	if !entityAttributionsEqual(prepared, importContext) {
		failures = append(failures, entityFailure(entitySourceIdentity, hostsBundlePath, "", importContext.IncidentID))
	}
	var claimsValid bool
	if err := tx.QueryRow(ctx, `
WITH expected AS (
    SELECT expected.incident_id, expected.entity_type, expected.identifier_type,
           expected.normalized_value, expected.record_id
      FROM entities_expected_active_identifier_claims_v1(NULL::uuid) AS expected
     WHERE expected.incident_id = $1
), actual AS (
    SELECT incident_id, entity_type, identifier_type, normalized_value, record_id
      FROM entity_active_identifier_claims
     WHERE incident_id = $1
), difference AS (
    (SELECT * FROM expected EXCEPT SELECT * FROM actual)
    UNION ALL
    (SELECT * FROM actual EXCEPT SELECT * FROM expected)
)
SELECT NOT EXISTS (SELECT 1 FROM difference)
`, importContext.IncidentID).Scan(&claimsValid); err != nil || !claimsValid {
		failures = append(failures, entityFailure(entityUnique, preservedIDsBundlePath, "", importContext.IncidentID))
	}
	return selectedEntityFailure(failures)
}

func comparePortableHosts(expected, actual []portableHostRow, failures *[]entityFailureCandidate) {
	want := make(map[uuid.UUID]portableHostRow, len(expected))
	for _, row := range expected {
		want[row.RecordID] = row
	}
	seen := map[uuid.UUID]struct{}{}
	for _, row := range actual {
		expectedRow, present := want[row.RecordID]
		if !present || !portableHostEqual(expectedRow, row) {
			*failures = append(*failures, entityFailure(entityEnvelopeTypeScope, hostsBundlePath, row.RecordID.String(), row.RecordID))
		}
		seen[row.RecordID] = struct{}{}
	}
	for _, row := range expected {
		if _, present := seen[row.RecordID]; !present {
			*failures = append(*failures, entityFailure(entityEnvelopeTypeScope, hostsBundlePath, row.RecordID.String(), row.RecordID))
		}
	}
}

func comparePortableIdentities(expected, actual []portableIdentityRow, failures *[]entityFailureCandidate) {
	want := make(map[uuid.UUID]portableIdentityRow, len(expected))
	for _, row := range expected {
		want[row.RecordID] = row
	}
	seen := map[uuid.UUID]struct{}{}
	for _, row := range actual {
		expectedRow, present := want[row.RecordID]
		if !present || !portableIdentityEqual(expectedRow, row) {
			*failures = append(*failures, entityFailure(entityEnvelopeTypeScope, identitiesBundlePath, row.RecordID.String(), row.RecordID))
		}
		seen[row.RecordID] = struct{}{}
	}
	for _, row := range expected {
		if _, present := seen[row.RecordID]; !present {
			*failures = append(*failures, entityFailure(entityEnvelopeTypeScope, identitiesBundlePath, row.RecordID.String(), row.RecordID))
		}
	}
}

func comparePortableMentions(expected, actual []portableMentionRow, failures *[]entityFailureCandidate) {
	want := make(map[uuid.UUID]portableMentionRow, len(expected))
	for _, row := range expected {
		want[row.MentionID] = row
	}
	seen := map[uuid.UUID]struct{}{}
	for _, row := range actual {
		expectedRow, present := want[row.MentionID]
		if !present || !portableMentionEqual(expectedRow, row) {
			*failures = append(*failures, entityFailure(entityResolutionMerge, entityMentionsBundlePath, row.MentionID.String(), row.MentionID))
		}
		seen[row.MentionID] = struct{}{}
	}
	for _, row := range expected {
		if _, present := seen[row.MentionID]; !present {
			*failures = append(*failures, entityFailure(entityResolutionMerge, entityMentionsBundlePath, row.MentionID.String(), row.MentionID))
		}
	}
}

func comparePortablePreserved(expected, actual []portablePreservedIdentifierRow, failures *[]entityFailureCandidate) {
	want := make(map[uuid.UUID]portablePreservedIdentifierRow, len(expected))
	for _, row := range expected {
		want[row.PreservedID] = row
	}
	seen := map[uuid.UUID]struct{}{}
	for _, row := range actual {
		expectedRow, present := want[row.PreservedID]
		if !present || !portablePreservedEqual(expectedRow, row) {
			*failures = append(*failures, entityFailure(entitySameIncident, preservedIDsBundlePath, row.PreservedID.String(), row.PreservedID))
		}
		seen[row.PreservedID] = struct{}{}
	}
	for _, row := range expected {
		if _, present := seen[row.PreservedID]; !present {
			*failures = append(*failures, entityFailure(entitySameIncident, preservedIDsBundlePath, row.PreservedID.String(), row.PreservedID))
		}
	}
}

func comparePortableAliases(expected, actual []portableAliasRow, failures *[]entityFailureCandidate) {
	want := make(map[uuid.UUID]portableAliasRow, len(expected))
	for _, row := range expected {
		want[row.AliasID] = row
	}
	seen := map[uuid.UUID]struct{}{}
	for _, row := range actual {
		expectedRow, present := want[row.AliasID]
		if !present || !portableAliasEqual(expectedRow, row) {
			*failures = append(*failures, entityFailure(entitySameIncident, entityAliasesBundlePath, row.AliasID.String(), row.AliasID))
		}
		seen[row.AliasID] = struct{}{}
	}
	for _, row := range expected {
		if _, present := seen[row.AliasID]; !present {
			*failures = append(*failures, entityFailure(entitySameIncident, entityAliasesBundlePath, row.AliasID.String(), row.AliasID))
		}
	}
}

func portableHostEqual(left, right portableHostRow) bool {
	return left.RecordID == right.RecordID && left.IncidentID == right.IncidentID && left.DisplayName == right.DisplayName &&
		entityStringPtrEqual(left.Hostname, right.Hostname) && entityStringPtrEqual(left.AADDeviceID, right.AADDeviceID) &&
		entityStringPtrEqual(left.FQDN, right.FQDN) && left.EntityOrigin == right.EntityOrigin &&
		entityUUIDPtrEqual(left.SeedMentionID, right.SeedMentionID) && left.State == right.State &&
		entityUUIDPtrEqual(left.MergedIntoRecordID, right.MergedIntoRecordID) && left.RowVersion == right.RowVersion &&
		left.CreatedAt.Equal(right.CreatedAt) && left.UpdatedAt.Equal(right.UpdatedAt) &&
		left.RuntimeCreatedByID == right.RuntimeCreatedByID && left.RuntimeUpdatedByID == right.RuntimeUpdatedByID &&
		entityStringPtrEqual(left.Location, right.Location) && entityStringPtrEqual(left.OSPlatform, right.OSPlatform) &&
		entityStringPtrEqual(left.BusinessOwner, right.BusinessOwner) && entityStringPtrEqual(left.Criticality, right.Criticality) &&
		entityStringPtrEqual(left.ContainmentStatus, right.ContainmentStatus)
}

func portableIdentityEqual(left, right portableIdentityRow) bool {
	return left.RecordID == right.RecordID && left.IncidentID == right.IncidentID && left.DisplayName == right.DisplayName &&
		entityStringPtrEqual(left.UPN, right.UPN) && entityStringPtrEqual(left.Email, right.Email) &&
		entityStringPtrEqual(left.SAMAccountName, right.SAMAccountName) && entityStringPtrEqual(left.AADObjectID, right.AADObjectID) &&
		entityStringPtrEqual(left.SID, right.SID) && left.EntityOrigin == right.EntityOrigin &&
		entityUUIDPtrEqual(left.SeedMentionID, right.SeedMentionID) && left.State == right.State &&
		entityUUIDPtrEqual(left.MergedIntoRecordID, right.MergedIntoRecordID) && left.RowVersion == right.RowVersion &&
		left.CreatedAt.Equal(right.CreatedAt) && left.UpdatedAt.Equal(right.UpdatedAt) &&
		left.RuntimeCreatedByID == right.RuntimeCreatedByID && left.RuntimeUpdatedByID == right.RuntimeUpdatedByID &&
		entityStringPtrEqual(left.PrivilegeLevel, right.PrivilegeLevel) && entityStringPtrEqual(left.MFAState, right.MFAState) &&
		entityStringPtrEqual(left.ResetStatus, right.ResetStatus)
}

func portableMentionEqual(left, right portableMentionRow) bool {
	return left.MentionID == right.MentionID && left.SourceRecordID == right.SourceRecordID && left.EntityType == right.EntityType &&
		left.SourceFieldKey == right.SourceFieldKey && left.OriginKind == right.OriginKind && left.OriginLocator == right.OriginLocator &&
		left.RawText == right.RawText && left.NormalizedText == right.NormalizedText && left.ResolutionStatus == right.ResolutionStatus &&
		left.RowVersion == right.RowVersion && left.Ordinal == right.Ordinal && left.RuntimeCreatedByID == right.RuntimeCreatedByID &&
		left.CreatedAt.Equal(right.CreatedAt) && entityUUIDPtrEqual(left.ResolvedRecordID, right.ResolvedRecordID) &&
		entityUUIDPtrEqual(left.RuntimeResolvedByID, right.RuntimeResolvedByID) && entityTimePtrEqual(left.ResolvedAt, right.ResolvedAt) &&
		entityStringPtrEqual(left.ResolutionMethod, right.ResolutionMethod)
}

func portablePreservedEqual(left, right portablePreservedIdentifierRow) bool {
	return left.PreservedID == right.PreservedID && left.IncidentID == right.IncidentID && left.RecordID == right.RecordID &&
		left.EntityType == right.EntityType && left.IdentifierType == right.IdentifierType && left.RawValue == right.RawValue &&
		left.NormalizedValue == right.NormalizedValue && left.Classification == right.Classification &&
		left.RuntimeCreatedByID == right.RuntimeCreatedByID && left.CreatedAt.Equal(right.CreatedAt) && entityTimePtrEqual(left.DeletedAt, right.DeletedAt)
}

func portableAliasEqual(left, right portableAliasRow) bool {
	return left.AliasID == right.AliasID && left.IncidentID == right.IncidentID && left.RecordID == right.RecordID &&
		left.EntityType == right.EntityType && left.RawText == right.RawText && left.NormalizedText == right.NormalizedText &&
		left.Classification == right.Classification && left.RuntimeCreatedByID == right.RuntimeCreatedByID &&
		left.CreatedAt.Equal(right.CreatedAt) && entityTimePtrEqual(left.DeletedAt, right.DeletedAt)
}

func entityAttributionsEqual(prepared preparedEntityImport, importContext sourceport.ImportContext) bool {
	want := map[string]string{}
	add := func(table, rowID, column string, actor *uuid.UUID) {
		if actor != nil {
			want[table+"\x00"+rowID+"\x00"+column] = actor.String()
		}
	}
	for _, row := range prepared.hosts {
		add("hosts", row.RecordID.String(), "created_by_user_id", &row.PortableCreatedByID)
		add("hosts", row.RecordID.String(), "updated_by_user_id", &row.PortableUpdatedByID)
	}
	for _, row := range prepared.identities {
		add("identities", row.RecordID.String(), "created_by_user_id", &row.PortableCreatedByID)
		add("identities", row.RecordID.String(), "updated_by_user_id", &row.PortableUpdatedByID)
	}
	for _, row := range prepared.mentions {
		add("entity_mentions", row.MentionID.String(), "created_by_user_id", &row.PortableCreatedByID)
		add("entity_mentions", row.MentionID.String(), "resolved_by_user_id", row.PortableResolvedByID)
	}
	for _, row := range prepared.preserved {
		add("entity_preserved_identifiers", row.PreservedID.String(), "created_by_user_id", &row.PortableCreatedByID)
	}
	for _, row := range prepared.aliases {
		add("entity_aliases", row.AliasID.String(), "created_by_user_id", &row.PortableCreatedByID)
	}
	found := map[string]string{}
	for _, attribution := range importContext.Attributions.ImportedAttributions() {
		key := attribution.SourceTable + "\x00" + attribution.SourceRowID + "\x00" + attribution.SourceColumn
		if _, relevant := want[key]; relevant {
			found[key] = attribution.SourceActorID
		}
	}
	if len(want) != len(found) {
		return false
	}
	for key, actorID := range want {
		if found[key] != actorID {
			return false
		}
	}
	return true
}
