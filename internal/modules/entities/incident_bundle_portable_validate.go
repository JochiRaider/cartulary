package entities

import (
	"context"
	"errors"
	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
	"github.com/JochiRaider/cartulary/internal/modules/incidentportability"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func validatePreparedEntityImportTx(ctx context.Context, tx pgx.Tx, prepared preparedEntityImport, input sourceport.ImportContext) error {
	if tx == nil || !prepared.binding.matches(input) || input.Attributions == nil {
		return entitySourceFailure(entitySourceIdentity)
	}
	var failures entityFailureAccumulator
	if equal, err := compareEntitySource(ctx, tx, input, func(visit func(portableHostRow) error) error { return prepared.hosts(ctx, input, visit) }, loadPortableHostsPage, func(row portableHostRow) uuid.UUID { return row.RecordID }, portableHostEqual); err != nil {
		return err
	} else if !equal {
		failures.add(entityEnvelopeTypeScope)
	}
	if equal, err := compareEntitySource(ctx, tx, input, func(visit func(portableIdentityRow) error) error { return prepared.identities(ctx, input, visit) }, loadPortableIdentitiesPage, func(row portableIdentityRow) uuid.UUID { return row.RecordID }, portableIdentityEqual); err != nil {
		return err
	} else if !equal {
		failures.add(entityEnvelopeTypeScope)
	}
	if equal, err := compareEntitySource(ctx, tx, input, func(visit func(portableMentionRow) error) error { return prepared.mentions(ctx, input, visit) }, loadPortableEntityMentionsPage, func(row portableMentionRow) uuid.UUID { return row.MentionID }, portableMentionEqual); err != nil {
		return err
	} else if !equal {
		failures.add(entityResolutionMerge)
	}
	if equal, err := compareEntitySource(ctx, tx, input, func(visit func(portablePreservedIdentifierRow) error) error {
		return prepared.preserved(ctx, input, visit)
	}, loadPortablePreservedIdentifiersPage, func(row portablePreservedIdentifierRow) uuid.UUID { return row.PreservedID }, portablePreservedEqual); err != nil {
		return err
	} else if !equal {
		failures.add(entitySameIncident)
	}
	if equal, err := compareEntitySource(ctx, tx, input, func(visit func(portableAliasRow) error) error { return prepared.aliases(ctx, input, visit) }, loadPortableAliasesPage, func(row portableAliasRow) uuid.UUID { return row.AliasID }, portableAliasEqual); err != nil {
		return err
	} else if !equal {
		failures.add(entitySameIncident)
	}
	if err := validateEntityAttributions(ctx, prepared, input); err != nil {
		failures.add(entitySourceIdentity)
	}
	if err := validateStagedEntityClaims(ctx, tx, prepared, input); err != nil {
		var failure *sourceport.Failure
		if !errors.As(err, &failure) {
			return err
		}
		failures.add(failure.InvariantID())
	}
	if err := failures.err(); err != nil {
		return err
	}
	for _, statement := range []string{`DELETE FROM public.entities_bundle_validation_claims WHERE transaction_id=pg_current_xact_id()`, `DELETE FROM public.entities_bundle_validation_binding WHERE transaction_id=pg_current_xact_id()`} {
		if _, err := tx.Exec(ctx, statement); err != nil {
			return err
		}
	}
	return nil
}

// The source stream advances against bounded keyset pages; only one page and
// one admitted source record are live, independent of the collection size.
func compareEntitySource[T any](ctx context.Context, tx pgx.Tx, input sourceport.ImportContext, walk func(func(T) error) error, load func(context.Context, sourceport.ExportContext, *uuid.UUID) ([]T, error), identity func(T) uuid.UUID, equal func(T, T) bool) (bool, error) {
	var page []T
	var after *uuid.UUID
	index := 0
	same := true
	next := func() (T, bool, error) {
		if index == len(page) {
			var err error
			page, err = load(ctx, sourceport.ExportContext{Query: tx, IncidentID: input.IncidentID}, after)
			if err != nil {
				var zero T
				return zero, false, err
			}
			index = 0
			if len(page) > 0 {
				last := identity(page[len(page)-1])
				after = &last
			}
		}
		if index == len(page) {
			var zero T
			return zero, false, nil
		}
		row := page[index]
		index++
		return row, true, nil
	}
	if err := walk(func(expected T) error {
		actual, present, err := next()
		if err != nil {
			return err
		}
		if !present || !equal(expected, actual) {
			same = false
		}
		return nil
	}); err != nil {
		return false, err
	}
	_, extra, err := next()
	return same && !extra, err
}

func validateEntityAttributions(ctx context.Context, prepared preparedEntityImport, input sourceport.ImportContext) error {
	for _, family := range []struct{ path, table, identity string }{
		{hostsBundlePath, "hosts", "record_id"}, {identitiesBundlePath, "identities", "record_id"},
		{entityMentionsBundlePath, "entity_mentions", "entity_mention_id"},
		{preservedIDsBundlePath, "entity_preserved_identifiers", "entity_preserved_identifier_id"},
		{entityAliasesBundlePath, "entity_aliases", "entity_alias_id"},
	} {
		if err := incidentportability.ValidateTopLevelAttributions(ctx, prepared.members, family.path, family.table, family.identity, input.Attributions); err != nil {
			return err
		}
	}
	return nil
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
