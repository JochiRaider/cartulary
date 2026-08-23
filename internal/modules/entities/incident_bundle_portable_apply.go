package entities

import (
	"context"
	"errors"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
	"github.com/JochiRaider/cartulary/internal/platform/fieldnorm"
)

type entityPortableEnvelope struct {
	recordID, incidentID             uuid.UUID
	recordType                       string
	rowVersion                       int64
	createdAt, updatedAt             time.Time
	createdByUserID, updatedByUserID uuid.UUID
	deletedAt                        *time.Time
}

func applyPreparedEntityImportTx(ctx context.Context, tx pgx.Tx, prepared preparedEntityImport, importContext sourceport.ImportContext) error {
	if tx == nil || !prepared.binding.matches(importContext) || importContext.Attributions == nil || importContext.ActorUserID == uuid.Nil {
		return entitySourceFailure(entitySourceIdentity)
	}
	if err := validateEntityCrossRowsBeforeApply(ctx, tx, prepared, importContext); err != nil {
		return err
	}
	for _, merged := range []bool{false, true} {
		for _, row := range prepared.hosts {
			if (row.State == "merged") != merged {
				continue
			}
			if err := insertPortableHost(ctx, tx, row); err != nil {
				return err
			}
			if err := recordEntityAttributions(importContext, "hosts", row.RecordID.String(), []entityPortableAttribution{
				{column: "created_by_user_id", actor: &row.PortableCreatedByID},
				{column: "updated_by_user_id", actor: &row.PortableUpdatedByID},
			}, entityEnvelopeTypeScope); err != nil {
				return err
			}
		}
		for _, row := range prepared.identities {
			if (row.State == "merged") != merged {
				continue
			}
			if err := insertPortableIdentity(ctx, tx, row); err != nil {
				return err
			}
			if err := recordEntityAttributions(importContext, "identities", row.RecordID.String(), []entityPortableAttribution{
				{column: "created_by_user_id", actor: &row.PortableCreatedByID},
				{column: "updated_by_user_id", actor: &row.PortableUpdatedByID},
			}, entityEnvelopeTypeScope); err != nil {
				return err
			}
		}
	}
	for _, row := range prepared.mentions {
		tag, err := tx.Exec(ctx, `
INSERT INTO entity_mentions (
    entity_mention_id, source_record_id, entity_type, source_field_key,
    origin_kind, origin_locator, raw_text, normalized_text, resolution_status,
    row_version, ordinal, created_by_user_id, created_at, resolved_record_id,
    resolved_by_user_id, resolved_at, resolution_method
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
`, row.MentionID, row.SourceRecordID, row.EntityType, row.SourceFieldKey,
			row.OriginKind, row.OriginLocator, row.RawText, row.NormalizedText,
			row.ResolutionStatus, row.RowVersion, row.Ordinal, row.RuntimeCreatedByID,
			row.CreatedAt, row.ResolvedRecordID, row.RuntimeResolvedByID, row.ResolvedAt,
			row.ResolutionMethod)
		if err != nil || tag.RowsAffected() != 1 {
			return classifyEntityApplyError(err, entityResolutionMerge)
		}
		if err := recordEntityAttributions(importContext, "entity_mentions", row.MentionID.String(), []entityPortableAttribution{
			{column: "created_by_user_id", actor: &row.PortableCreatedByID},
			{column: "resolved_by_user_id", actor: row.PortableResolvedByID},
		}, entityMentionsObserved); err != nil {
			return err
		}
	}
	for _, row := range prepared.hosts {
		if row.SeedMentionID == nil {
			continue
		}
		tag, err := tx.Exec(ctx, `UPDATE hosts SET seed_entity_mention_id = $2 WHERE record_id = $1`, row.RecordID, row.SeedMentionID)
		if err != nil || tag.RowsAffected() != 1 {
			return classifyEntityApplyError(err, entitySourceIdentity)
		}
	}
	for _, row := range prepared.identities {
		if row.SeedMentionID == nil {
			continue
		}
		tag, err := tx.Exec(ctx, `UPDATE identities SET seed_entity_mention_id = $2 WHERE record_id = $1`, row.RecordID, row.SeedMentionID)
		if err != nil || tag.RowsAffected() != 1 {
			return classifyEntityApplyError(err, entitySourceIdentity)
		}
	}
	for _, row := range prepared.preserved {
		tag, err := tx.Exec(ctx, `
INSERT INTO entity_preserved_identifiers (
    entity_preserved_identifier_id, incident_id, record_id, entity_type,
    identifier_type, raw_value, normalized_value, classification,
    created_by_user_id, created_at, deleted_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
`, row.PreservedID, row.IncidentID, row.RecordID, row.EntityType, row.IdentifierType,
			row.RawValue, row.NormalizedValue, row.Classification, row.RuntimeCreatedByID,
			row.CreatedAt, row.DeletedAt)
		if err != nil || tag.RowsAffected() != 1 {
			return classifyEntityApplyError(err, entityUnique)
		}
		if err := recordEntityAttributions(importContext, "entity_preserved_identifiers", row.PreservedID.String(), []entityPortableAttribution{
			{column: "created_by_user_id", actor: &row.PortableCreatedByID},
		}, entitySourceIdentity); err != nil {
			return err
		}
	}
	for _, row := range prepared.aliases {
		tag, err := tx.Exec(ctx, `
INSERT INTO entity_aliases (
    entity_alias_id, incident_id, record_id, entity_type, raw_text,
    normalized_text, classification, created_by_user_id, created_at, deleted_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
`, row.AliasID, row.IncidentID, row.RecordID, row.EntityType, row.RawText,
			row.NormalizedText, row.Classification, row.RuntimeCreatedByID,
			row.CreatedAt, row.DeletedAt)
		if err != nil || tag.RowsAffected() != 1 {
			return classifyEntityApplyError(err, entityUnique)
		}
		if err := recordEntityAttributions(importContext, "entity_aliases", row.AliasID.String(), []entityPortableAttribution{
			{column: "created_by_user_id", actor: &row.PortableCreatedByID},
		}, entitySourceIdentity); err != nil {
			return err
		}
	}
	return nil
}

func insertPortableHost(ctx context.Context, tx pgx.Tx, row portableHostRow) error {
	tag, err := tx.Exec(ctx, `
INSERT INTO hosts (
    record_id, incident_id, display_name, hostname, aad_device_id, fqdn,
    entity_origin, seed_entity_mention_id, host_state, merged_into_record_id,
    row_version, created_at, updated_at, created_by_user_id, updated_by_user_id,
    location, os_platform, business_owner, criticality, containment_status
)
VALUES ($1, $2, $3, $4, $5, $6, $7, NULL, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19)
`, row.RecordID, row.IncidentID, row.DisplayName, row.Hostname, row.AADDeviceID,
		row.FQDN, row.EntityOrigin, row.State, row.MergedIntoRecordID, row.RowVersion,
		row.CreatedAt, row.UpdatedAt, row.RuntimeCreatedByID, row.RuntimeUpdatedByID,
		row.Location, row.OSPlatform, row.BusinessOwner, row.Criticality, row.ContainmentStatus)
	if err != nil || tag.RowsAffected() != 1 {
		return classifyEntityApplyError(err, entityEnvelopeTypeScope)
	}
	return nil
}

func insertPortableIdentity(ctx context.Context, tx pgx.Tx, row portableIdentityRow) error {
	tag, err := tx.Exec(ctx, `
INSERT INTO identities (
    record_id, incident_id, display_name, upn, email, sam_account_name,
    aad_object_id, sid, entity_origin, seed_entity_mention_id, identity_state,
    merged_into_record_id, row_version, created_at, updated_at,
    created_by_user_id, updated_by_user_id, privilege_level, mfa_state, reset_status
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NULL, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19)
`, row.RecordID, row.IncidentID, row.DisplayName, row.UPN, row.Email,
		row.SAMAccountName, row.AADObjectID, row.SID, row.EntityOrigin, row.State,
		row.MergedIntoRecordID, row.RowVersion, row.CreatedAt, row.UpdatedAt,
		row.RuntimeCreatedByID, row.RuntimeUpdatedByID, row.PrivilegeLevel,
		row.MFAState, row.ResetStatus)
	if err != nil || tag.RowsAffected() != 1 {
		return classifyEntityApplyError(err, entityEnvelopeTypeScope)
	}
	return nil
}

func validateEntityCrossRowsBeforeApply(ctx context.Context, tx pgx.Tx, prepared preparedEntityImport, importContext sourceport.ImportContext) error {
	envelopes, err := loadEntityPortableEnvelopes(ctx, tx, importContext.IncidentID)
	if err != nil {
		return entitySourceFailure(entityEnvelopeTypeScope)
	}
	var failures []entityFailureCandidate
	hosts := make(map[uuid.UUID]portableHostRow, len(prepared.hosts))
	identities := make(map[uuid.UUID]portableIdentityRow, len(prepared.identities))
	for _, row := range prepared.hosts {
		hosts[row.RecordID] = row
		envelope, present := envelopes[row.RecordID]
		if !present || !hostMatchesPortableEnvelope(row, envelope) {
			failures = append(failures, entityFailure(entityEnvelopeTypeScope, hostsBundlePath, row.RecordID.String(), row.RecordID))
		}
	}
	for _, row := range prepared.identities {
		identities[row.RecordID] = row
		envelope, present := envelopes[row.RecordID]
		if !present || !identityMatchesPortableEnvelope(row, envelope) {
			failures = append(failures, entityFailure(entityEnvelopeTypeScope, identitiesBundlePath, row.RecordID.String(), row.RecordID))
		}
	}
	for _, row := range prepared.hosts {
		if row.MergedIntoRecordID != nil {
			target, present := hosts[*row.MergedIntoRecordID]
			if !present || target.State == "merged" {
				failures = append(failures, entityFailure(entityResolutionMerge, hostsBundlePath, row.RecordID.String(), row.MergedIntoRecordID))
			}
		}
		if row.SeedMentionID != nil && !preparedHasMention(prepared.mentions, *row.SeedMentionID) {
			failures = append(failures, entityFailure(entitySourceIdentity, hostsBundlePath, row.RecordID.String(), row.SeedMentionID))
		}
	}
	for _, row := range prepared.identities {
		if row.MergedIntoRecordID != nil {
			target, present := identities[*row.MergedIntoRecordID]
			if !present || target.State == "merged" {
				failures = append(failures, entityFailure(entityResolutionMerge, identitiesBundlePath, row.RecordID.String(), row.MergedIntoRecordID))
			}
		}
		if row.SeedMentionID != nil && !preparedHasMention(prepared.mentions, *row.SeedMentionID) {
			failures = append(failures, entityFailure(entitySourceIdentity, identitiesBundlePath, row.RecordID.String(), row.SeedMentionID))
		}
	}
	for _, row := range prepared.mentions {
		source, present := envelopes[row.SourceRecordID]
		if !present || source.incidentID != importContext.IncidentID {
			failures = append(failures, entityFailure(entityResolutionMerge, entityMentionsBundlePath, row.MentionID.String(), row.SourceRecordID))
		}
		if row.ResolvedRecordID != nil {
			target, present := envelopes[*row.ResolvedRecordID]
			if !present || target.incidentID != importContext.IncidentID || target.recordType != row.EntityType {
				failures = append(failures, entityFailure(entityResolutionMerge, entityMentionsBundlePath, row.MentionID.String(), row.ResolvedRecordID))
			}
		}
	}
	for _, row := range prepared.preserved {
		if !entityOwnerPrepared(hosts, identities, row.RecordID, row.EntityType) {
			failures = append(failures, entityFailure(entitySameIncident, preservedIDsBundlePath, row.PreservedID.String(), row.RecordID))
		}
	}
	for _, row := range prepared.aliases {
		if !entityOwnerPrepared(hosts, identities, row.RecordID, row.EntityType) {
			failures = append(failures, entityFailure(entitySameIncident, entityAliasesBundlePath, row.AliasID.String(), row.RecordID))
		}
	}
	failures = append(failures, entityClaimFailures(prepared, envelopes)...)
	if existing, err := existingEntityClaimFailures(ctx, tx, importContext.IncidentID, prepared); err != nil {
		return entitySourceFailure(entityUnique)
	} else {
		failures = append(failures, existing...)
	}
	return selectedEntityFailure(failures)
}

func loadEntityPortableEnvelopes(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID) (map[uuid.UUID]entityPortableEnvelope, error) {
	rows, err := tx.Query(ctx, `
SELECT record_id, incident_id, record_type, row_version, created_at, updated_at,
       created_by_user_id, updated_by_user_id, deleted_at
  FROM records
 WHERE incident_id = $1
 ORDER BY record_id
`, incidentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := map[uuid.UUID]entityPortableEnvelope{}
	for rows.Next() {
		var envelope entityPortableEnvelope
		var deletedAt pgtype.Timestamptz
		if err := rows.Scan(&envelope.recordID, &envelope.incidentID, &envelope.recordType,
			&envelope.rowVersion, &envelope.createdAt, &envelope.updatedAt,
			&envelope.createdByUserID, &envelope.updatedByUserID, &deletedAt); err != nil {
			return nil, err
		}
		if deletedAt.Valid {
			value := deletedAt.Time
			envelope.deletedAt = &value
		}
		result[envelope.recordID] = envelope
	}
	return result, rows.Err()
}

func hostMatchesPortableEnvelope(row portableHostRow, envelope entityPortableEnvelope) bool {
	return envelope.recordID == row.RecordID && envelope.incidentID == row.IncidentID && envelope.recordType == "host" &&
		envelope.rowVersion == row.RowVersion && envelope.createdAt.Equal(row.CreatedAt) && envelope.updatedAt.Equal(row.UpdatedAt) &&
		envelope.createdByUserID == row.RuntimeCreatedByID && envelope.updatedByUserID == row.RuntimeUpdatedByID
}

func identityMatchesPortableEnvelope(row portableIdentityRow, envelope entityPortableEnvelope) bool {
	return envelope.recordID == row.RecordID && envelope.incidentID == row.IncidentID && envelope.recordType == "identity" &&
		envelope.rowVersion == row.RowVersion && envelope.createdAt.Equal(row.CreatedAt) && envelope.updatedAt.Equal(row.UpdatedAt) &&
		envelope.createdByUserID == row.RuntimeCreatedByID && envelope.updatedByUserID == row.RuntimeUpdatedByID
}

func preparedHasMention(rows []portableMentionRow, id uuid.UUID) bool {
	index := sort.Search(len(rows), func(i int) bool { return rows[i].MentionID.String() >= id.String() })
	return index < len(rows) && rows[index].MentionID == id
}

func entityOwnerPrepared(hosts map[uuid.UUID]portableHostRow, identities map[uuid.UUID]portableIdentityRow, recordID uuid.UUID, entityType string) bool {
	if entityType == "host" {
		_, present := hosts[recordID]
		return present
	}
	if entityType == "identity" {
		_, present := identities[recordID]
		return present
	}
	return false
}

func entityClaimFailures(prepared preparedEntityImport, envelopes map[uuid.UUID]entityPortableEnvelope) []entityFailureCandidate {
	type claimOwner struct {
		recordID uuid.UUID
		path     string
	}
	claims := map[string]claimOwner{}
	var failures []entityFailureCandidate
	add := func(recordID uuid.UUID, entityType, identifierType string, rawValue *string, path string, identity string) {
		if rawValue == nil {
			return
		}
		normalized, ok := fieldnorm.NormalizeIdentifier(identifierType, *rawValue)
		if !ok {
			failures = append(failures, entityFailure(entitySourceIdentity, path, identity, rawValue))
			return
		}
		key := strings.Join([]string{entityType, identifierType, normalized}, "\x00")
		if prior, duplicate := claims[key]; duplicate && prior.recordID != recordID {
			failures = append(failures,
				entityFailure(entityUnique, prior.path, prior.recordID.String(), key),
				entityFailure(entityUnique, path, identity, key),
			)
			return
		}
		claims[key] = claimOwner{recordID: recordID, path: path}
	}
	for _, row := range prepared.hosts {
		if envelope := envelopes[row.RecordID]; envelope.deletedAt == nil && row.State != "merged" {
			add(row.RecordID, "host", "aad_device_id", row.AADDeviceID, hostsBundlePath, row.RecordID.String())
			add(row.RecordID, "host", "fqdn", row.FQDN, hostsBundlePath, row.RecordID.String())
			add(row.RecordID, "host", "hostname", row.Hostname, hostsBundlePath, row.RecordID.String())
		}
	}
	for _, row := range prepared.identities {
		if envelope := envelopes[row.RecordID]; envelope.deletedAt == nil && row.State != "merged" {
			add(row.RecordID, "identity", "aad_object_id", row.AADObjectID, identitiesBundlePath, row.RecordID.String())
			add(row.RecordID, "identity", "sid", row.SID, identitiesBundlePath, row.RecordID.String())
			add(row.RecordID, "identity", "upn", row.UPN, identitiesBundlePath, row.RecordID.String())
			add(row.RecordID, "identity", "email", row.Email, identitiesBundlePath, row.RecordID.String())
			add(row.RecordID, "identity", "sam_account_name", row.SAMAccountName, identitiesBundlePath, row.RecordID.String())
		}
	}
	for _, row := range prepared.preserved {
		envelope := envelopes[row.RecordID]
		if envelope.deletedAt == nil && row.DeletedAt == nil && row.Classification == "exact_match_reuse" {
			value := row.RawValue
			add(row.RecordID, row.EntityType, row.IdentifierType, &value, preservedIDsBundlePath, row.PreservedID.String())
		}
	}
	return failures
}

func existingEntityClaimFailures(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, prepared preparedEntityImport) ([]entityFailureCandidate, error) {
	var count int
	if err := tx.QueryRow(ctx, `SELECT count(*) FROM entity_active_identifier_claims WHERE incident_id = $1`, incidentID).Scan(&count); err != nil {
		return nil, err
	}
	if count == 0 {
		return nil, nil
	}
	return []entityFailureCandidate{entityFailure(entityUnique, hostsBundlePath, "", prepared.binding.incidentID)}, nil
}

type entityPortableAttribution struct {
	column string
	actor  *uuid.UUID
}

func recordEntityAttributions(importContext sourceport.ImportContext, table, rowID string, values []entityPortableAttribution, invariant string) error {
	for _, value := range values {
		if value.actor == nil {
			continue
		}
		if err := importContext.Attributions.RecordImportedAttribution(table, rowID, value.column, value.actor.String()); err != nil {
			return entitySourceFailure(invariant)
		}
	}
	return nil
}

func classifyEntityApplyError(err error, invariant string) error {
	if err == nil {
		return entitySourceFailure(invariant)
	}
	var postgresFailure *pgconn.PgError
	if errors.As(err, &postgresFailure) {
		return entitySourceFailure(invariant)
	}
	return entitySourceFailure(invariant)
}
