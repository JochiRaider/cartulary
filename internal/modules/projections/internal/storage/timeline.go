package storage

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	timelineprojection "github.com/JochiRaider/cartulary/internal/modules/timeline/workbookprojection"
)

func (store *Store) UpsertTimelineRowTx(
	ctx context.Context,
	tx pgx.Tx,
	input timelineprojection.ProjectionInput,
) error {
	hostRefs, err := json.Marshal(input.HostRefs)
	if err != nil {
		return fmt.Errorf("encode timeline host refs: %w", err)
	}
	identityRefs, err := json.Marshal(input.IdentityRefs)
	if err != nil {
		return fmt.Errorf("encode timeline identity refs: %w", err)
	}
	tags, err := json.Marshal(input.Tags)
	if err != nil {
		return fmt.Errorf("encode timeline tags: %w", err)
	}
	attachedEvidence, err := json.Marshal(input.AttachedEvidence)
	if err != nil {
		return fmt.Errorf("encode timeline attached evidence refs: %w", err)
	}
	if _, err := tx.Exec(ctx, `
INSERT INTO timeline_grid_projection (
    record_id, incident_id, row_version, date_entered_text, analyst_text,
    mitre_stage_text, device_object_text, ip_address_text, activity_utc_text,
    activity_local_text, raw_activity_text, activity_synopsis_text,
    data_source_text, recorded_at, edited_at, activity_sort_ts,
    date_entered_sort_day, activity_time_pair_state, capture_state,
    replacement_record_id, evidence_count, has_evidence,
    has_unresolved_mentions, host_refs, identity_refs, tags,
    attached_evidence_refs
)
VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14,
    $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27
)
ON CONFLICT (record_id) DO UPDATE
SET incident_id = EXCLUDED.incident_id,
    row_version = EXCLUDED.row_version,
    date_entered_text = EXCLUDED.date_entered_text,
    analyst_text = EXCLUDED.analyst_text,
    mitre_stage_text = EXCLUDED.mitre_stage_text,
    device_object_text = EXCLUDED.device_object_text,
    ip_address_text = EXCLUDED.ip_address_text,
    activity_utc_text = EXCLUDED.activity_utc_text,
    activity_local_text = EXCLUDED.activity_local_text,
    raw_activity_text = EXCLUDED.raw_activity_text,
    activity_synopsis_text = EXCLUDED.activity_synopsis_text,
    data_source_text = EXCLUDED.data_source_text,
    recorded_at = EXCLUDED.recorded_at,
    edited_at = EXCLUDED.edited_at,
    activity_sort_ts = EXCLUDED.activity_sort_ts,
    date_entered_sort_day = EXCLUDED.date_entered_sort_day,
    activity_time_pair_state = EXCLUDED.activity_time_pair_state,
    capture_state = EXCLUDED.capture_state,
    replacement_record_id = EXCLUDED.replacement_record_id,
    evidence_count = EXCLUDED.evidence_count,
    has_evidence = EXCLUDED.has_evidence,
    has_unresolved_mentions = EXCLUDED.has_unresolved_mentions,
    host_refs = EXCLUDED.host_refs,
    identity_refs = EXCLUDED.identity_refs,
    tags = EXCLUDED.tags,
    attached_evidence_refs = EXCLUDED.attached_evidence_refs
`, input.RecordID, input.IncidentID, input.RowVersion, input.DateEnteredText,
		input.AnalystText, input.MitreStageText, input.DeviceObjectText,
		input.IPAddressText, input.ActivityUTCText, input.ActivityLocalText,
		input.RawActivityText, input.ActivitySynopsisText, input.DataSourceText,
		input.RecordedAt.UTC(), input.EditedAt.UTC(), input.ActivitySortTS,
		input.DateEnteredSortDay, input.ActivityTimePairState, input.CaptureState,
		input.ReplacementRecordID, input.EvidenceCount, input.HasEvidence,
		input.HasUnresolvedMentions, string(hostRefs), string(identityRefs),
		string(tags), string(attachedEvidence)); err != nil {
		return fmt.Errorf("upsert timeline projection row: %w", err)
	}
	return nil
}

func (store *Store) InsertTimelineFixtureBatchTx(
	ctx context.Context,
	tx pgx.Tx,
	inputs []timelineprojection.ProjectionInput,
) error {
	if len(inputs) == 0 {
		return fmt.Errorf("insert Timeline fixture projections: empty input")
	}
	rows := make([][]any, len(inputs))
	for index, input := range inputs {
		hostRefs, err := json.Marshal(input.HostRefs)
		if err != nil {
			return fmt.Errorf("encode Timeline fixture host refs %d: %w", index+1, err)
		}
		identityRefs, err := json.Marshal(input.IdentityRefs)
		if err != nil {
			return fmt.Errorf("encode Timeline fixture identity refs %d: %w", index+1, err)
		}
		tags, err := json.Marshal(input.Tags)
		if err != nil {
			return fmt.Errorf("encode Timeline fixture tags %d: %w", index+1, err)
		}
		attachedEvidence, err := json.Marshal(input.AttachedEvidence)
		if err != nil {
			return fmt.Errorf("encode Timeline fixture evidence refs %d: %w", index+1, err)
		}
		rows[index] = []any{
			input.RecordID, input.IncidentID, input.RowVersion, input.DateEnteredText,
			input.AnalystText, input.MitreStageText, input.DeviceObjectText,
			input.IPAddressText, input.ActivityUTCText, input.ActivityLocalText,
			input.RawActivityText, input.ActivitySynopsisText, input.DataSourceText,
			input.RecordedAt.UTC(), input.EditedAt.UTC(), input.ActivitySortTS,
			input.DateEnteredSortDay, input.ActivityTimePairState, input.CaptureState,
			input.ReplacementRecordID, input.EvidenceCount, input.HasEvidence,
			input.HasUnresolvedMentions, json.RawMessage(hostRefs), json.RawMessage(identityRefs),
			json.RawMessage(tags), json.RawMessage(attachedEvidence),
		}
	}
	inserted, err := tx.CopyFrom(
		ctx,
		pgx.Identifier{"timeline_grid_projection"},
		[]string{
			"record_id", "incident_id", "row_version", "date_entered_text", "analyst_text",
			"mitre_stage_text", "device_object_text", "ip_address_text", "activity_utc_text",
			"activity_local_text", "raw_activity_text", "activity_synopsis_text",
			"data_source_text", "recorded_at", "edited_at", "activity_sort_ts",
			"date_entered_sort_day", "activity_time_pair_state", "capture_state",
			"replacement_record_id", "evidence_count", "has_evidence",
			"has_unresolved_mentions", "host_refs", "identity_refs", "tags",
			"attached_evidence_refs",
		},
		pgx.CopyFromRows(rows),
	)
	if err != nil {
		return fmt.Errorf("insert Timeline fixture projection batch: %w", err)
	}
	if inserted != int64(len(rows)) {
		return fmt.Errorf("insert Timeline fixture projection batch: inserted %d rows, want %d", inserted, len(rows))
	}
	return nil
}

func (store *Store) CountTimelineFixtureRows(ctx context.Context, incidentID uuid.UUID) (int, error) {
	return countTimelineFixtureRows(ctx, store.db, incidentID)
}

func (store *Store) CountTimelineFixtureRowsTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID) (int, error) {
	return countTimelineFixtureRows(ctx, tx, incidentID)
}

type timelineFixtureRowQuerier interface {
	QueryRow(context.Context, string, ...any) pgx.Row
}

func countTimelineFixtureRows(ctx context.Context, querier timelineFixtureRowQuerier, incidentID uuid.UUID) (int, error) {
	var count int
	if err := querier.QueryRow(ctx, `SELECT COUNT(*) FROM timeline_grid_projection WHERE incident_id = $1`, incidentID).Scan(&count); err != nil {
		return 0, fmt.Errorf("count timeline fixture projection rows: %w", err)
	}
	return count, nil
}

func (store *Store) DeleteTimelineRowTx(
	ctx context.Context,
	tx pgx.Tx,
	recordID uuid.UUID,
) error {
	if _, err := tx.Exec(ctx, `DELETE FROM timeline_grid_projection WHERE record_id = $1`, recordID); err != nil {
		return fmt.Errorf("delete timeline projection row: %w", err)
	}
	return nil
}

func (store *Store) DeleteTimelineIncidentTx(
	ctx context.Context,
	tx pgx.Tx,
	incidentID uuid.UUID,
) error {
	if _, err := tx.Exec(ctx, `DELETE FROM timeline_grid_projection WHERE incident_id = $1`, incidentID); err != nil {
		return fmt.Errorf("clear timeline projection rows: %w", err)
	}
	return nil
}
