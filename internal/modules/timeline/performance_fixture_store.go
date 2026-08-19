package timeline

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/timeline/sourcerepository"
	"github.com/JochiRaider/cartulary/internal/modules/timeline/workbookprojection"
)

type performanceFixtureRecordBatchPort interface {
	InsertPerformanceFixtureBatchTx(context.Context, pgx.Tx, []RecordCreateParams) error
}

func (s *store) createPerformanceFixtureRows(ctx context.Context, command PerformanceFixtureCommand) (PerformanceFixtureResult, error) {
	recordBatch, ok := s.recordStore.(performanceFixtureRecordBatchPort)
	if !ok {
		return PerformanceFixtureResult{}, fmt.Errorf("timeline performance fixture record batch port is unavailable")
	}
	projectionBatch, ok := s.projectionStore.(workbookprojection.FixtureBatchPort)
	if !ok {
		return PerformanceFixtureResult{}, fmt.Errorf("timeline performance fixture projection batch port is unavailable")
	}
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return PerformanceFixtureResult{}, fmt.Errorf("begin timeline performance fixture transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if err := s.incidentAccess.RequireOpenTx(ctx, tx, command.IncidentID); err != nil {
		return PerformanceFixtureResult{}, err
	}
	var existing int
	if err := tx.QueryRow(ctx, `SELECT COUNT(*) FROM timeline_events WHERE incident_id = $1`, command.IncidentID).Scan(&existing); err != nil {
		return PerformanceFixtureResult{}, fmt.Errorf("query preexisting timeline performance fixture rows: %w", err)
	}
	if existing != 0 {
		return PerformanceFixtureResult{}, fmt.Errorf("timeline performance fixture requires an empty incident timeline")
	}

	now := command.Now.UTC()
	envelopes := make([]RecordCreateParams, len(command.Rows))
	sourceRows := make([][]any, len(command.Rows))
	projections := make([]workbookprojection.ProjectionInput, len(command.Rows))
	recordIDs := make([]uuid.UUID, len(command.Rows))
	for index, row := range command.Rows {
		recordID := uuid.New()
		recordIDs[index] = recordID
		envelopes[index] = RecordCreateParams{
			RecordID: &recordID, IncidentID: command.IncidentID, RecordType: "timeline_event",
			CreatedByUserID: command.Actor.ID, CreatedAt: now, UpdatedByUserID: command.Actor.ID,
			UpdatedAt: now, RowVersion: 1,
		}
		summary := row.Summary
		var dataSource *string
		if row.DataSource != "" {
			value := row.DataSource
			dataSource = &value
		}
		snapshot := sourcerepository.Snapshot{
			RecordID: recordID, IncidentID: command.IncidentID,
			ActivitySynopsisText: &summary, DataSourceText: dataSource,
			ActivityTimePairState: "disabled", CaptureState: InitialCaptureState(), RowVersion: 1,
			RecordedAt: now, EditedAt: now, CreatedByUserID: command.Actor.ID, UpdatedByUserID: command.Actor.ID,
		}
		projections[index] = projectRecord(snapshot, nil).ProjectionInput()
		sourceRows[index] = []any{
			recordID, command.IncidentID, snapshot.CaptureState, int64(1), now, now,
			command.Actor.ID, command.Actor.ID, snapshot.ActivitySynopsisText,
			snapshot.DataSourceText, snapshot.ActivityTimePairState,
		}
	}
	if err := recordBatch.InsertPerformanceFixtureBatchTx(ctx, tx, envelopes); err != nil {
		return PerformanceFixtureResult{}, err
	}
	inserted, err := tx.CopyFrom(
		ctx,
		pgx.Identifier{"timeline_events"},
		[]string{
			"record_id", "incident_id", "capture_state", "row_version", "recorded_at", "edited_at",
			"created_by_user_id", "updated_by_user_id", "activity_synopsis_text", "data_source_text",
			"activity_time_pair_state",
		},
		pgx.CopyFromRows(sourceRows),
	)
	if err != nil {
		return PerformanceFixtureResult{}, fmt.Errorf("insert Timeline performance fixture source rows: %w", err)
	}
	if inserted != int64(len(sourceRows)) {
		return PerformanceFixtureResult{}, fmt.Errorf("insert Timeline performance fixture source rows: inserted %d rows, want %d", inserted, len(sourceRows))
	}

	var entityRefresh mentionProjectionRefresh
	result := PerformanceFixtureResult{RowCount: len(command.Rows)}
	for index, row := range command.Rows {
		if row.HostRef == "" {
			continue
		}
		result.RelationshipRows++
		cells := make([]ownerBatchCellV1, 0, 3)
		for _, field := range []struct{ key, value string }{
			{"timeline.host_refs", row.HostRef},
			{"timeline.identity_refs", row.IdentityRef},
			{"timeline.tags", row.Tag},
		} {
			change, ok := clipboardValueToPatchChange(field.key, field.value)
			if !ok {
				return PerformanceFixtureResult{}, fmt.Errorf("build timeline performance fixture collection %s", field.key)
			}
			cells = append(cells, ownerBatchCellV1{FieldKey: field.key, Value: field.value, Change: change})
		}
		refresh, err := s.applyPasteMentionActionsTx(ctx, tx, command.Actor, command.IncidentID, recordIDs[index], cells, "clipboard_paste", now)
		if err != nil {
			return PerformanceFixtureResult{}, err
		}
		entityRefresh.merge(refresh)
		if _, err := s.applyPasteTagActionsTx(ctx, tx, command.Actor.ID, command.IncidentID, recordIDs[index], cells, now); err != nil {
			return PerformanceFixtureResult{}, err
		}
		derived := projectRecord(sourcerepository.Snapshot{
			RecordID: recordIDs[index], IncidentID: command.IncidentID,
			ActivitySynopsisText:  projections[index].ActivitySynopsisText,
			DataSourceText:        projections[index].DataSourceText,
			ActivityTimePairState: "disabled", CaptureState: InitialCaptureState(), RowVersion: 1,
			RecordedAt: now, EditedAt: now, CreatedByUserID: command.Actor.ID, UpdatedByUserID: command.Actor.ID,
		}, nil)
		if err := s.hydrateProjectedCollections(ctx, tx, &derived); err != nil {
			return PerformanceFixtureResult{}, err
		}
		projections[index] = derived.ProjectionInput()
	}
	if err := s.refreshMentionEntityProjectionsTx(ctx, tx, entityRefresh); err != nil {
		return PerformanceFixtureResult{}, err
	}
	if err := projectionBatch.ApplyTimelineFixtureBatchTx(ctx, tx, projections); err != nil {
		return PerformanceFixtureResult{}, err
	}
	actual, err := performanceFixtureCountsTx(ctx, tx, command.IncidentID, projectionBatch)
	if err != nil {
		return PerformanceFixtureResult{}, err
	}
	if actual != result {
		return PerformanceFixtureResult{}, fmt.Errorf("timeline performance fixture transaction produced rows=%d relationships=%d, want rows=%d relationships=%d", actual.RowCount, actual.RelationshipRows, result.RowCount, result.RelationshipRows)
	}
	if err := tx.Commit(ctx); err != nil {
		return PerformanceFixtureResult{}, fmt.Errorf("commit timeline performance fixture transaction: %w", err)
	}
	return result, nil
}

func (s *store) performanceFixtureCounts(ctx context.Context, incidentID uuid.UUID) (PerformanceFixtureResult, error) {
	projectionBatch, ok := s.projectionStore.(workbookprojection.FixtureBatchPort)
	if !ok {
		return PerformanceFixtureResult{}, fmt.Errorf("timeline performance fixture projection batch port is unavailable")
	}
	var result PerformanceFixtureResult
	err := s.pool.QueryRow(ctx, `
SELECT COUNT(*), COUNT(*) FILTER (WHERE data_source_text LIKE 'https://fixture-%')
  FROM timeline_events
 WHERE incident_id = $1
`, incidentID).Scan(&result.RowCount, &result.RelationshipRows)
	if err != nil {
		return PerformanceFixtureResult{}, fmt.Errorf("query Timeline performance fixture counts: %w", err)
	}
	projectionCount, err := projectionBatch.CountTimelineFixtureRows(ctx, incidentID)
	if err != nil {
		return PerformanceFixtureResult{}, err
	}
	if projectionCount != result.RowCount {
		return PerformanceFixtureResult{}, fmt.Errorf("timeline performance fixture projection count=%d rows=%d", projectionCount, result.RowCount)
	}
	return result, nil
}

func performanceFixtureCountsTx(
	ctx context.Context,
	tx pgx.Tx,
	incidentID uuid.UUID,
	projectionBatch workbookprojection.FixtureBatchPort,
) (PerformanceFixtureResult, error) {
	var result PerformanceFixtureResult
	err := tx.QueryRow(ctx, `
SELECT COUNT(*), COUNT(*) FILTER (WHERE data_source_text LIKE 'https://fixture-%')
  FROM timeline_events
 WHERE incident_id = $1
`, incidentID).Scan(&result.RowCount, &result.RelationshipRows)
	if err != nil {
		return PerformanceFixtureResult{}, fmt.Errorf("validate Timeline performance fixture source rows: %w", err)
	}
	projectionCount, err := projectionBatch.CountTimelineFixtureRowsTx(ctx, tx, incidentID)
	if err != nil {
		return PerformanceFixtureResult{}, err
	}
	if projectionCount != result.RowCount {
		return PerformanceFixtureResult{}, fmt.Errorf("validate Timeline performance fixture projections: got %d want %d", projectionCount, result.RowCount)
	}
	return result, nil
}
