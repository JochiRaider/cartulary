package projection

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	indicatorprojection "github.com/JochiRaider/cartulary/internal/modules/indicators/workbookprojection"
)

func LoadProjectionInputTx(
	ctx context.Context,
	tx pgx.Tx,
	recordID uuid.UUID,
) (indicatorprojection.ProjectionInput, bool, error) {
	input, err := scanProjectionInput(tx.QueryRow(ctx, indicatorProjectionSourceSQL+`
 WHERE i.record_id = $1
   AND r.deleted_at IS NULL
`, recordID))
	if errors.Is(err, pgx.ErrNoRows) {
		return indicatorprojection.ProjectionInput{}, false, nil
	}
	if err != nil {
		return indicatorprojection.ProjectionInput{}, false, fmt.Errorf("load Indicator projection input: %w", err)
	}
	return input, true, nil
}

func ListProjectionInputsTx(
	ctx context.Context,
	tx pgx.Tx,
	incidentID uuid.UUID,
	afterRecordID *uuid.UUID,
	limit int,
) (indicatorprojection.ProjectionInputPage, error) {
	if limit < 1 || limit > 1000 {
		return indicatorprojection.ProjectionInputPage{}, fmt.Errorf("indicator projection source page limit %d is outside 1..1000", limit)
	}
	rows, err := tx.Query(ctx, indicatorProjectionSourceSQL+`
 WHERE i.incident_id = $1
   AND ($2::uuid IS NULL OR i.record_id > $2)
   AND r.deleted_at IS NULL
 ORDER BY i.record_id
 LIMIT $3
`, incidentID, afterRecordID, limit+1)
	if err != nil {
		return indicatorprojection.ProjectionInputPage{}, fmt.Errorf("list Indicator projection inputs: %w", err)
	}
	defer rows.Close()

	inputs := make([]indicatorprojection.ProjectionInput, 0, limit+1)
	for rows.Next() {
		input, scanErr := scanProjectionInput(rows)
		if scanErr != nil {
			return indicatorprojection.ProjectionInputPage{}, fmt.Errorf("scan Indicator projection input: %w", scanErr)
		}
		inputs = append(inputs, input)
	}
	if err := rows.Err(); err != nil {
		return indicatorprojection.ProjectionInputPage{}, fmt.Errorf("iterate Indicator projection inputs: %w", err)
	}

	page := indicatorprojection.ProjectionInputPage{Inputs: inputs}
	if len(inputs) > limit {
		page.Inputs = inputs[:limit]
		next := page.Inputs[len(page.Inputs)-1].RecordID
		page.NextRecordID = &next
	}
	return page, nil
}

func scanProjectionInput(scanner interface{ Scan(...any) error }) (indicatorprojection.ProjectionInput, error) {
	var (
		input               indicatorprojection.ProjectionInput
		normalizedValue     pgtype.Text
		defangedValue       pgtype.Text
		hashAlgorithm       pgtype.Text
		hashValue           pgtype.Text
		stixPattern         pgtype.Text
		firstObservedAt     pgtype.Timestamptz
		lastObservedAt      pgtype.Timestamptz
		observationCount    int32
		lifecycleSummary    pgtype.Text
		supportingLinkCount int32
	)
	if err := scanner.Scan(
		&input.RecordID,
		&input.IncidentID,
		&input.RowVersion,
		&input.IndicatorType,
		&input.ValueKind,
		&input.DisplayValue,
		&normalizedValue,
		&input.DedupeKey,
		&defangedValue,
		&hashAlgorithm,
		&hashValue,
		&stixPattern,
		&firstObservedAt,
		&lastObservedAt,
		&observationCount,
		&lifecycleSummary,
		&supportingLinkCount,
		&input.EditedAt,
	); err != nil {
		return indicatorprojection.ProjectionInput{}, err
	}
	input.NormalizedValue = textPointer(normalizedValue)
	input.DefangedValue = textPointer(defangedValue)
	input.HashAlgorithm = textPointer(hashAlgorithm)
	input.HashValue = textPointer(hashValue)
	input.STIXPattern = textPointer(stixPattern)
	input.FirstObservedAt = timePointer(firstObservedAt)
	input.LastObservedAt = timePointer(lastObservedAt)
	input.ObservationCount = int(observationCount)
	input.LifecycleSummary = textPointer(lifecycleSummary)
	input.SupportingLinkCount = int(supportingLinkCount)
	return input, nil
}

func textPointer(value pgtype.Text) *string {
	if !value.Valid {
		return nil
	}
	result := value.String
	return &result
}

func timePointer(value pgtype.Timestamptz) *time.Time {
	if !value.Valid {
		return nil
	}
	result := value.Time.UTC()
	return &result
}

const indicatorProjectionSourceSQL = `
SELECT
    i.record_id,
    i.incident_id,
    r.row_version,
    i.indicator_type,
    i.value_kind,
    i.display_value,
    i.normalized_value,
    i.dedupe_key,
    i.defanged_value,
    i.hash_algorithm,
    i.hash_value,
    i.stix_pattern,
    obs.first_observed_at,
    obs.last_observed_at,
    COALESCE(obs.observation_count, 0)::integer,
    lifecycle.lifecycle_summary,
    COALESCE(links.supporting_link_count, 0)::integer,
    r.updated_at
  FROM indicators i
  JOIN records r
    ON r.record_id = i.record_id
  LEFT JOIN (
        SELECT
            resolved_indicator_record_id,
            MIN(created_at) AS first_observed_at,
            MAX(created_at) AS last_observed_at,
            COUNT(*) AS observation_count
          FROM indicator_observations
         WHERE resolution_status = 'resolved'
           AND resolved_indicator_record_id IS NOT NULL
           AND deleted_at IS NULL
         GROUP BY resolved_indicator_record_id
  ) obs
    ON obs.resolved_indicator_record_id = i.record_id
  LEFT JOIN (
        SELECT DISTINCT ON (indicator_record_id)
            indicator_record_id,
            lifecycle_state AS lifecycle_summary
          FROM indicator_state_intervals
         WHERE deleted_at IS NULL
         ORDER BY indicator_record_id, CASE WHEN valid_to IS NULL THEN 0 ELSE 1 END ASC, valid_from DESC, indicator_state_interval_id DESC
  ) lifecycle
    ON lifecycle.indicator_record_id = i.record_id
  LEFT JOIN (
        SELECT dst_record_id, COUNT(*) AS supporting_link_count
          FROM active_record_links_v1
         GROUP BY dst_record_id
  ) links
    ON links.dst_record_id = i.record_id`
