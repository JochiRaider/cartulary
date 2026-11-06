package entities

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
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
		if err := prepared.hosts(ctx, importContext, func(row portableHostRow) error {
			if (row.State == "merged") != merged {
				return nil
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

			return nil
		}); err != nil {
			return err
		}
		if err := prepared.identities(ctx, importContext, func(row portableIdentityRow) error {
			if (row.State == "merged") != merged {
				return nil
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

			return nil
		}); err != nil {
			return err
		}
	}
	if err := prepared.mentions(ctx, importContext, func(row portableMentionRow) error {
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

		return nil
	}); err != nil {
		return err
	}
	if err := prepared.hosts(ctx, importContext, func(row portableHostRow) error {
		if row.SeedMentionID == nil {
			return nil
		}
		tag, err := tx.Exec(ctx, `UPDATE hosts SET seed_entity_mention_id = $2 WHERE record_id = $1`, row.RecordID, row.SeedMentionID)
		if err != nil || tag.RowsAffected() != 1 {
			return classifyEntityApplyError(err, entitySourceIdentity)
		}

		return nil
	}); err != nil {
		return err
	}
	if err := prepared.identities(ctx, importContext, func(row portableIdentityRow) error {
		if row.SeedMentionID == nil {
			return nil
		}
		tag, err := tx.Exec(ctx, `UPDATE identities SET seed_entity_mention_id = $2 WHERE record_id = $1`, row.RecordID, row.SeedMentionID)
		if err != nil || tag.RowsAffected() != 1 {
			return classifyEntityApplyError(err, entitySourceIdentity)
		}

		return nil
	}); err != nil {
		return err
	}
	if err := prepared.preserved(ctx, importContext, func(row portablePreservedIdentifierRow) error {
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

		return nil
	}); err != nil {
		return err
	}
	if err := prepared.aliases(ctx, importContext, func(row portableAliasRow) error {
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

		return nil
	}); err != nil {
		return err
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

func loadEntityPortableEnvelope(ctx context.Context, tx pgx.Tx, incidentID, recordID uuid.UUID) (entityPortableEnvelope, bool, error) {
	var envelope entityPortableEnvelope
	var deletedAt pgtype.Timestamptz
	err := tx.QueryRow(ctx, `SELECT record_id, incident_id, record_type, row_version, created_at, updated_at, created_by_user_id, updated_by_user_id, deleted_at FROM records WHERE incident_id = $1 AND record_id = $2`, incidentID, recordID).Scan(&envelope.recordID, &envelope.incidentID, &envelope.recordType, &envelope.rowVersion, &envelope.createdAt, &envelope.updatedAt, &envelope.createdByUserID, &envelope.updatedByUserID, &deletedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return envelope, false, nil
	}
	if err != nil {
		return envelope, false, err
	}
	if deletedAt.Valid {
		envelope.deletedAt = &deletedAt.Time
	}
	return envelope, true, nil
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
