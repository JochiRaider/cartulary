package incidentbundle

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
)

func applyPreparedIndicatorImportTx(
	ctx context.Context,
	tx pgx.Tx,
	prepared preparedIndicatorImport,
	importContext sourceport.ImportContext,
) error {
	if tx == nil || !prepared.binding.matches(importContext) ||
		importContext.ActorUserID == uuid.Nil || importContext.Attributions == nil {
		return indicatorSourceFailure(representationInvariant)
	}
	for _, row := range prepared.indicators {
		tag, err := tx.Exec(ctx, `
INSERT INTO indicators (
    record_id, incident_id, indicator_type, value_kind, display_value,
    normalized_value, dedupe_key, defanged_value, hash_algorithm, hash_value,
    stix_pattern, row_version, created_at, updated_at, created_by_user_id,
    updated_by_user_id, deleted_at, deleted_by_user_id
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)
`,
			row.RecordID, row.IncidentID, row.IndicatorType, row.ValueKind,
			row.DisplayValue, row.NormalizedValue, row.DedupeKey, row.DefangedValue,
			row.HashAlgorithm, row.HashValue, row.STIXPattern, row.RowVersion,
			row.CreatedAt, row.UpdatedAt, row.RuntimeCreatedByID,
			row.RuntimeUpdatedByID, row.DeletedAt, row.RuntimeDeletedByID,
		)
		if err != nil || tag.RowsAffected() != 1 {
			return classifyIndicatorApplyError(err, representationInvariant)
		}
		if err := recordIndicatorAttributions(importContext, row); err != nil {
			return err
		}
	}
	for _, row := range prepared.observations {
		if err := requirePortableRecordReference(
			ctx, tx, row.IncidentID, row.SourceRecordID, observationIncidentInvariant,
		); err != nil {
			return err
		}
		if row.ResolvedIndicatorID != nil {
			if err := requirePortableIndicatorReference(
				ctx, tx, row.IncidentID, *row.ResolvedIndicatorID, observationIncidentInvariant,
			); err != nil {
				return err
			}
		}
		tag, err := tx.Exec(ctx, `
INSERT INTO indicator_observations (
    indicator_observation_id, incident_id, source_record_id, source_field_key,
    origin_kind, origin_locator, observed_text, parsed_indicator_type,
    normalized_candidate, resolution_status, resolved_indicator_record_id,
    row_version, created_by_user_id, created_at, resolved_by_user_id,
    resolved_at, resolution_method, deleted_at, deleted_by_user_id
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19)
`,
			row.ObservationID, row.IncidentID, row.SourceRecordID, row.SourceFieldKey,
			row.OriginKind.String(), row.OriginLocator, row.ObservedText,
			row.ParsedIndicatorType, row.NormalizedCandidate, row.ResolutionStatus,
			row.ResolvedIndicatorID, row.RowVersion, row.RuntimeCreatedByID,
			row.CreatedAt, row.RuntimeResolvedByID, row.ResolvedAt,
			row.ResolutionMethod, row.DeletedAt, row.RuntimeDeletedByID,
		)
		if err != nil || tag.RowsAffected() != 1 {
			return classifyIndicatorApplyError(err, observationCoherentInvariant)
		}
		if err := recordObservationAttributions(importContext, row); err != nil {
			return err
		}
	}
	for _, row := range prepared.intervals {
		if err := requirePortableIndicatorReference(
			ctx, tx, row.IncidentID, row.IndicatorRecordID, intervalIncidentInvariant,
		); err != nil {
			return err
		}
		for _, supportID := range row.SupportRefs {
			if err := requirePortableRecordReference(
				ctx, tx, row.IncidentID, supportID, intervalIncidentInvariant,
			); err != nil {
				return err
			}
		}
		supportRefs, err := json.Marshal(uuidStrings(row.SupportRefs))
		if err != nil {
			return indicatorSourceFailure(intervalCoherentInvariant)
		}
		tag, err := tx.Exec(ctx, `
INSERT INTO indicator_state_intervals (
    indicator_state_interval_id, incident_id, indicator_record_id,
    lifecycle_state, valid_from, valid_to, confidence, rationale, support_refs,
    assessor, assessed_at, row_version, created_by_user_id, created_at,
    deleted_at, deleted_by_user_id
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9::jsonb, $10, $11, $12, $13, $14, $15, $16)
`,
			row.IntervalID, row.IncidentID, row.IndicatorRecordID, row.LifecycleState,
			row.ValidFrom, row.ValidTo, row.Confidence, row.Rationale, supportRefs,
			row.Assessor, row.AssessedAt, row.RowVersion, row.RuntimeCreatedByID,
			row.CreatedAt, row.DeletedAt, row.RuntimeDeletedByID,
		)
		if err != nil || tag.RowsAffected() != 1 {
			return classifyIndicatorApplyError(err, intervalCoherentInvariant)
		}
		if err := recordIntervalAttributions(importContext, row); err != nil {
			return err
		}
	}
	return nil
}

func requirePortableRecordReference(
	ctx context.Context,
	tx pgx.Tx,
	incidentID uuid.UUID,
	recordID uuid.UUID,
	invariant string,
) error {
	var present bool
	if err := tx.QueryRow(ctx, `
SELECT EXISTS (
    SELECT 1
      FROM records
     WHERE incident_id = $1
       AND record_id = $2
)
`, incidentID, recordID).Scan(&present); err != nil || !present {
		return indicatorSourceFailure(invariant)
	}
	return nil
}

func requirePortableIndicatorReference(
	ctx context.Context,
	tx pgx.Tx,
	incidentID uuid.UUID,
	recordID uuid.UUID,
	invariant string,
) error {
	var present bool
	if err := tx.QueryRow(ctx, `
SELECT EXISTS (
    SELECT 1
      FROM indicators
     WHERE incident_id = $1
       AND record_id = $2
)
`, incidentID, recordID).Scan(&present); err != nil || !present {
		return indicatorSourceFailure(invariant)
	}
	return nil
}

func recordIndicatorAttributions(importContext sourceport.ImportContext, row portableIndicatorRow) error {
	return recordPortableAttributions(importContext, "indicators", row.RecordID.String(), []portableAttribution{
		{column: "created_by_user_id", actor: &row.PortableCreatedByID},
		{column: "updated_by_user_id", actor: &row.PortableUpdatedByID},
		{column: "deleted_by_user_id", actor: row.PortableDeletedByID},
	}, representationInvariant)
}

func recordObservationAttributions(importContext sourceport.ImportContext, row portableObservationRow) error {
	return recordPortableAttributions(importContext, "indicator_observations", row.ObservationID.String(), []portableAttribution{
		{column: "created_by_user_id", actor: &row.PortableCreatedByID},
		{column: "resolved_by_user_id", actor: row.PortableResolvedByID},
		{column: "deleted_by_user_id", actor: row.PortableDeletedByID},
	}, observationCoherentInvariant)
}

func recordIntervalAttributions(importContext sourceport.ImportContext, row portableIntervalRow) error {
	return recordPortableAttributions(importContext, "indicator_state_intervals", row.IntervalID.String(), []portableAttribution{
		{column: "created_by_user_id", actor: &row.PortableCreatedByID},
		{column: "deleted_by_user_id", actor: row.PortableDeletedByID},
	}, intervalCoherentInvariant)
}

type portableAttribution struct {
	column string
	actor  *uuid.UUID
}

func recordPortableAttributions(
	importContext sourceport.ImportContext,
	table string,
	rowID string,
	attributions []portableAttribution,
	invariant string,
) error {
	for _, attribution := range attributions {
		if attribution.actor == nil {
			continue
		}
		if err := importContext.Attributions.RecordImportedAttribution(
			table, rowID, attribution.column, attribution.actor.String(),
		); err != nil {
			return indicatorSourceFailure(invariant)
		}
	}
	return nil
}

func classifyIndicatorApplyError(err error, invariant string) error {
	if err == nil {
		return indicatorSourceFailure(invariant)
	}
	var postgresFailure *pgconn.PgError
	if errors.As(err, &postgresFailure) {
		return indicatorSourceFailure(invariant)
	}
	// Source failures are deliberately closed and do not disclose adapter,
	// relation, constraint, or hostile row details to the coordinator.
	return indicatorSourceFailure(invariant)
}
