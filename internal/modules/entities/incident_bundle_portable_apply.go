package entities

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
	"github.com/JochiRaider/cartulary/internal/modules/incidentportability"
	"github.com/JochiRaider/cartulary/internal/modules/records"
	"github.com/JochiRaider/cartulary/internal/modules/records/enveloperead"
)

type entityPortableEnvelope struct {
	recordID, incidentID             uuid.UUID
	recordType                       string
	rowVersion                       int64
	createdAt, updatedAt             time.Time
	createdByUserID, updatedByUserID uuid.UUID
	deletedAt                        *time.Time
}

var (
	portableHostColumns = []string{
		"record_id", "incident_id", "display_name", "hostname", "aad_device_id", "fqdn",
		"entity_origin", "seed_entity_mention_id", "host_state", "merged_into_record_id",
		"row_version", "created_at", "updated_at", "created_by_user_id", "updated_by_user_id",
		"location", "os_platform", "business_owner", "criticality", "containment_status",
	}
	portableIdentityColumns = []string{
		"record_id", "incident_id", "display_name", "upn", "email", "sam_account_name",
		"aad_object_id", "sid", "entity_origin", "seed_entity_mention_id", "identity_state",
		"merged_into_record_id", "row_version", "created_at", "updated_at",
		"created_by_user_id", "updated_by_user_id", "privilege_level", "mfa_state", "reset_status",
	}
	portableMentionColumns = []string{
		"entity_mention_id", "source_record_id", "entity_type", "source_field_key",
		"origin_kind", "origin_locator", "raw_text", "normalized_text", "resolution_status",
		"row_version", "ordinal", "created_by_user_id", "created_at", "resolved_record_id",
		"resolved_by_user_id", "resolved_at", "resolution_method",
	}
	portablePreservedColumns = []string{
		"entity_preserved_identifier_id", "incident_id", "record_id", "entity_type",
		"identifier_type", "raw_value", "normalized_value", "classification",
		"created_by_user_id", "created_at", "deleted_at",
	}
	portableAliasColumns = []string{
		"entity_alias_id", "incident_id", "record_id", "entity_type", "raw_text",
		"normalized_text", "classification", "created_by_user_id", "created_at", "deleted_at",
	}
)

func portableHostValues(host stagedEntityHost) []any {
	row := host.Row
	return []any{row.RecordID, row.IncidentID, row.DisplayName, row.Hostname, row.AADDeviceID,
		row.FQDN, row.EntityOrigin, nil, row.State, row.MergedIntoRecordID, row.RowVersion,
		row.CreatedAt, row.UpdatedAt, row.RuntimeCreatedByID, row.RuntimeUpdatedByID,
		row.Location, row.OSPlatform, row.BusinessOwner, row.Criticality, row.ContainmentStatus}
}

func portableIdentityValues(identity stagedEntityIdentity) []any {
	row := identity.Row
	return []any{row.RecordID, row.IncidentID, row.DisplayName, row.UPN, row.Email,
		row.SAMAccountName, row.AADObjectID, row.SID, row.EntityOrigin, nil, row.State,
		row.MergedIntoRecordID, row.RowVersion, row.CreatedAt, row.UpdatedAt,
		row.RuntimeCreatedByID, row.RuntimeUpdatedByID, row.PrivilegeLevel,
		row.MFAState, row.ResetStatus}
}

// applyPreparedEntityImportTx writes the staged rows with one COPY per family
// and phase, in the live import's order: unmerged hosts and identities, then
// merged ones, mentions, seed links, preserved identifiers and aliases.
// Attribution was staged while preparing; the core flushes it in step 9.
func applyPreparedEntityImportTx(ctx context.Context, tx pgx.Tx, prepared preparedEntityImport, importContext sourceport.ImportContext) error {
	if tx == nil || !prepared.binding.matches(importContext) || importContext.Attributions == nil || importContext.ActorUserID == uuid.Nil || !prepared.staged || importContext.Preparation == nil {
		return entitySourceFailure(entitySourceIdentity)
	}
	if err := validateEntityCrossRowsBeforeApply(ctx, tx, prepared, importContext); err != nil {
		return err
	}
	store := importContext.Preparation
	for _, merged := range []bool{false, true} {
		hosts, err := incidentportability.OpenPreparedFilteredCopy(ctx, store, prepared.hostRows, hostCodec{}, func(host stagedEntityHost) bool {
			return (host.Row.State == "merged") == merged
		})
		if err != nil {
			return classifyEntityApplyError(err, entityEnvelopeTypeScope)
		}
		written, err := tx.CopyFrom(ctx, pgx.Identifier{"hosts"}, portableHostColumns, hosts)
		if err := hosts.Finish(written, err); err != nil {
			return classifyEntityApplyError(err, entityEnvelopeTypeScope)
		}
		identities, err := incidentportability.OpenPreparedFilteredCopy(ctx, store, prepared.identityRows, identityCodec{}, func(identity stagedEntityIdentity) bool {
			return (identity.Row.State == "merged") == merged
		})
		if err != nil {
			return classifyEntityApplyError(err, entityEnvelopeTypeScope)
		}
		written, err = tx.CopyFrom(ctx, pgx.Identifier{"identities"}, portableIdentityColumns, identities)
		if err := identities.Finish(written, err); err != nil {
			return classifyEntityApplyError(err, entityEnvelopeTypeScope)
		}
	}
	mentions, err := incidentportability.OpenPreparedCopy(ctx, store, prepared.mentionRows, mentionCodec{})
	if err != nil {
		return classifyEntityApplyError(err, entityResolutionMerge)
	}
	written, err := tx.CopyFrom(ctx, pgx.Identifier{"entity_mentions"}, portableMentionColumns, mentions)
	if err := mentions.Finish(written, err); err != nil {
		return classifyEntityApplyError(err, entityResolutionMerge)
	}
	if err := readEntityBatches(ctx, prepared, importContext, prepared.hostRows, hostCodec{}, func(batch []stagedEntityHost) error {
		ids, seeds := []uuid.UUID{}, []uuid.UUID{}
		for _, host := range batch {
			if host.Row.SeedMentionID != nil {
				ids, seeds = append(ids, host.Row.RecordID), append(seeds, *host.Row.SeedMentionID)
			}
		}
		return linkEntitySeeds(ctx, tx, "host", ids, seeds)
	}); err != nil {
		return err
	}
	if err := readEntityBatches(ctx, prepared, importContext, prepared.identityRows, identityCodec{}, func(batch []stagedEntityIdentity) error {
		ids, seeds := []uuid.UUID{}, []uuid.UUID{}
		for _, identity := range batch {
			if identity.Row.SeedMentionID != nil {
				ids, seeds = append(ids, identity.Row.RecordID), append(seeds, *identity.Row.SeedMentionID)
			}
		}
		return linkEntitySeeds(ctx, tx, "identity", ids, seeds)
	}); err != nil {
		return err
	}
	preserved, err := incidentportability.OpenPreparedCopy(ctx, store, prepared.preservedRows, preservedCodec{})
	if err != nil {
		return classifyEntityApplyError(err, entityUnique)
	}
	written, err = tx.CopyFrom(ctx, pgx.Identifier{"entity_preserved_identifiers"}, portablePreservedColumns, preserved)
	if err := preserved.Finish(written, err); err != nil {
		return classifyEntityApplyError(err, entityUnique)
	}
	aliases, err := incidentportability.OpenPreparedCopy(ctx, store, prepared.aliasRows, aliasCodec{})
	if err != nil {
		return classifyEntityApplyError(err, entityUnique)
	}
	written, err = tx.CopyFrom(ctx, pgx.Identifier{"entity_aliases"}, portableAliasColumns, aliases)
	if err := aliases.Finish(written, err); err != nil {
		return classifyEntityApplyError(err, entityUnique)
	}
	return nil
}

// linkEntitySeeds sets one batch of seed mentions with one statement; every
// staged link must update exactly its row.
func linkEntitySeeds(ctx context.Context, tx pgx.Tx, kind string, ids, seeds []uuid.UUID) error {
	if len(ids) == 0 {
		return nil
	}
	var tag pgconn.CommandTag
	var err error
	if kind == "host" {
		tag, err = tx.Exec(ctx, `UPDATE hosts AS target SET seed_entity_mention_id = link.seed FROM unnest($1::uuid[],$2::uuid[]) AS link(id,seed) WHERE target.record_id = link.id`, ids, seeds)
	} else {
		tag, err = tx.Exec(ctx, `UPDATE identities AS target SET seed_entity_mention_id = link.seed FROM unnest($1::uuid[],$2::uuid[]) AS link(id,seed) WHERE target.record_id = link.id`, ids, seeds)
	}
	if err != nil || tag.RowsAffected() != int64(len(ids)) {
		return classifyEntityApplyError(err, entitySourceIdentity)
	}
	return nil
}

func loadEntityPortableEnvelope(ctx context.Context, tx pgx.Tx, incidentID, recordID uuid.UUID) (entityPortableEnvelope, bool, error) {
	envelope, err := records.NewStore().LoadEnvelopeTx(ctx, tx, recordID, false)
	if errors.Is(err, enveloperead.ErrEnvelopeNotFound) {
		return entityPortableEnvelope{}, false, nil
	}
	if err != nil {
		return entityPortableEnvelope{}, false, err
	}
	if envelope.IncidentID != incidentID {
		return entityPortableEnvelope{}, false, nil
	}
	return entityPortableEnvelopeFrom(envelope), true, nil
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
