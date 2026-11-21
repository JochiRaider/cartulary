package timeline

import (
	"context"
	"errors"
	"reflect"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
	"github.com/JochiRaider/cartulary/internal/modules/incidentportability"
	"github.com/JochiRaider/cartulary/internal/modules/timeline/internal/sourcevalue"
)

func (p preparedTimelineImport) rawEvents(ctx context.Context, d sourceport.Descriptor, visit func(map[string]any) error) error {
	err := incidentportability.WalkPath(ctx, p.members, timelineBundleRecordsPath, visit)
	var malformed *incidentportability.StrictNDJSONDecodeError
	if errors.As(err, &malformed) {
		return d.DeclaredFailure("timeline.version_shape_exact")
	}
	return err
}

func (p preparedTimelineImport) events(ctx context.Context, d sourceport.Descriptor, visit func(sourcevalue.Row) error) error {
	return p.rawEvents(ctx, d, func(raw map[string]any) error {
		row, ok := sourcevalue.Parse(raw)
		if !ok {
			return d.DeclaredFailure("timeline.version_shape_exact")
		}
		return visit(row)
	})
}

func (p preparedTimelineImport) prepareEvents(ctx context.Context, input sourceport.ImportContext, d sourceport.Descriptor) error {
	previous := ""
	if err := p.rawEvents(ctx, d, func(raw map[string]any) error {
		if !portableTimelineUUID(raw["record_id"]) {
			return d.DeclaredFailure("timeline.source_identity_admitted")
		}
		id := raw["record_id"].(string)
		if id <= previous {
			return d.DeclaredFailure("timeline.source_identity_admitted")
		}
		previous = id
		return nil
	}); err != nil {
		return err
	}
	for _, phase := range []struct {
		invariant string
		valid     func(sourcevalue.Row) bool
	}{
		{"version_shape_exact", func(sourcevalue.Row) bool { return true }},
		{"envelope_type_scope", func(r sourcevalue.Row) bool { return r.IncidentID == p.incidentID }},
		{"lifecycle_coherent", sourcevalue.Row.LifecycleLegal},
		{"generated_time_coherent", sourcevalue.Row.GeneratedTimeLegal},
		{"paired_time_coherent", sourcevalue.Row.PairStateLegal},
		{"event_attribution_exact", func(r sourcevalue.Row) bool {
			for _, id := range []*uuid.UUID{r.ReviewedByUserID, r.SupersededByUserID} {
				if id != nil {
					if _, ok := input.Actors.Lookup(id.String()); !ok {
						return false
					}
				}
			}
			return true
		}},
	} {
		if err := p.events(ctx, d, func(row sourcevalue.Row) error {
			if !phase.valid(row) {
				return d.DeclaredFailure("timeline." + phase.invariant)
			}
			return nil
		}); err != nil {
			return err
		}
	}
	return nil
}

// stagedEventBatches visits the staged events in admitted identity order, at
// most timelineSetBatch at a time, for one set-based statement per batch.
func (p preparedTimelineImport) stagedEventBatches(ctx context.Context, store incidentportability.MemberReader, visit func([]sourcevalue.Row) error) error {
	batch := make([]sourcevalue.Row, 0, timelineSetBatch)
	if err := incidentportability.ReadPreparedRows(ctx, store, p.eventRows, timelineEventCodec{}, func(row sourcevalue.Row) error {
		batch = append(batch, row)
		if len(batch) < timelineSetBatch {
			return nil
		}
		err := visit(batch)
		batch = batch[:0]
		return err
	}); err != nil {
		return err
	}
	if len(batch) == 0 {
		return nil
	}
	return visit(batch)
}

// preflightEvents checks every staged event's envelope and identity occupancy
// with one statement per batch. The first failing event in identity order
// selects the invariant, as the per-row check did.
func (p preparedTimelineImport) preflightEvents(ctx context.Context, tx pgx.Tx, store incidentportability.MemberReader, d sourceport.Descriptor) error {
	return p.stagedEventBatches(ctx, store, func(batch []sourcevalue.Row) error {
		ids := make([]uuid.UUID, len(batch))
		for index, row := range batch {
			ids[index] = row.RecordID
		}
		rows, err := tx.Query(ctx, `
SELECT EXISTS(SELECT 1 FROM records WHERE record_id=u.id AND incident_id=$2 AND record_type='timeline_event'),
       EXISTS(SELECT 1 FROM timeline_events WHERE record_id=u.id)
  FROM unnest($1::uuid[]) WITH ORDINALITY AS u(id, ordinal)
 ORDER BY u.ordinal`, ids, p.incidentID)
		if err != nil {
			return err
		}
		checks, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) ([2]bool, error) {
			var check [2]bool
			return check, row.Scan(&check[0], &check[1])
		})
		if err != nil {
			return err
		}
		if len(checks) != len(batch) {
			return incidentportability.ErrPreparedStream
		}
		for _, check := range checks {
			if !check[0] {
				return d.DeclaredFailure("timeline.envelope_type_scope")
			}
			if check[1] {
				return d.DeclaredFailure("timeline.source_identity_admitted")
			}
		}
		return nil
	})
}

// applyEvents inserts each staged batch with one statement that copies the
// envelope mirrors from the applied records, in identity order.
func (p preparedTimelineImport) applyEvents(ctx context.Context, tx pgx.Tx, store incidentportability.MemberReader, d sourceport.Descriptor) error {
	return p.stagedEventBatches(ctx, store, func(batch []sourcevalue.Row) error {
		var (
			ids, incidents                   = make([]uuid.UUID, len(batch)), make([]uuid.UUID, len(batch))
			captureStates, pairStates        = make([]string, len(batch)), make([]string, len(batch))
			reviewers, superseders           = make([]*uuid.UUID, len(batch)), make([]*uuid.UUID, len(batch))
			reviewedAt, supersededAt         = make([]*time.Time, len(batch)), make([]*time.Time, len(batch))
			utcGenerated, localGenerated     = make([]bool, len(batch)), make([]bool, len(batch))
			dates, analysts, stages, devices = make([]*string, len(batch)), make([]*string, len(batch)), make([]*string, len(batch)), make([]*string, len(batch))
			addresses, utcTexts, localTexts  = make([]*string, len(batch)), make([]*string, len(batch)), make([]*string, len(batch))
			rawTexts, synopses, dataSources  = make([]*string, len(batch)), make([]*string, len(batch)), make([]*string, len(batch))
		)
		for index, row := range batch {
			ids[index], incidents[index], captureStates[index], pairStates[index] = row.RecordID, row.IncidentID, row.CaptureState, row.ActivityTimePairState
			reviewers[index], superseders[index] = p.importedEventActor(row.ReviewedByUserID), p.importedEventActor(row.SupersededByUserID)
			reviewedAt[index], supersededAt[index] = row.ReviewedAt, row.SupersededAt
			utcGenerated[index], localGenerated[index] = row.ActivityUTCGenerated, row.ActivityLocalGenerated
			dates[index], analysts[index], stages[index], devices[index] = row.DateEnteredText, row.AnalystText, row.MitreStageText, row.DeviceObjectText
			addresses[index], utcTexts[index], localTexts[index] = row.IPAddressText, row.ActivityUTCText, row.ActivityLocalText
			rawTexts[index], synopses[index], dataSources[index] = row.RawActivityText, row.ActivitySynopsisText, row.DataSourceText
		}
		tag, err := tx.Exec(ctx, `INSERT INTO timeline_events (
record_id,incident_id,capture_state,row_version,recorded_at,edited_at,created_by_user_id,updated_by_user_id,
reviewed_by_user_id,reviewed_at,superseded_by_user_id,superseded_at,date_entered_text,analyst_text,mitre_stage_text,
device_object_text,ip_address_text,activity_utc_text,activity_local_text,raw_activity_text,activity_synopsis_text,
data_source_text,activity_utc_generated,activity_local_generated,activity_time_pair_state)
SELECT u.record_id,u.incident_id,u.capture_state,r.row_version,r.created_at,r.updated_at,r.created_by_user_id,r.updated_by_user_id,
u.reviewer,u.reviewed_at,u.superseder,u.superseded_at,u.date_entered,u.analyst,u.mitre_stage,u.device_object,u.ip_address,
u.activity_utc,u.activity_local,u.raw_activity,u.activity_synopsis,u.data_source,u.utc_generated,u.local_generated,u.pair_state
FROM unnest($1::uuid[],$2::uuid[],$3::text[],$4::uuid[],$5::timestamptz[],$6::uuid[],$7::timestamptz[],$8::text[],$9::text[],$10::text[],
$11::text[],$12::text[],$13::text[],$14::text[],$15::text[],$16::text[],$17::text[],$18::bool[],$19::bool[],$20::text[])
  WITH ORDINALITY AS u(record_id,incident_id,capture_state,reviewer,reviewed_at,superseder,superseded_at,date_entered,analyst,
       mitre_stage,device_object,ip_address,activity_utc,activity_local,raw_activity,activity_synopsis,data_source,utc_generated,
       local_generated,pair_state,ordinal)
  JOIN records r ON r.record_id=u.record_id AND r.incident_id=u.incident_id AND r.record_type='timeline_event'
 ORDER BY u.ordinal`,
			ids, incidents, captureStates, reviewers, reviewedAt, superseders, supersededAt, dates, analysts, stages, devices, addresses,
			utcTexts, localTexts, rawTexts, synopses, dataSources, utcGenerated, localGenerated, pairStates)
		if err != nil {
			return err
		}
		if tag.RowsAffected() != int64(len(batch)) {
			return d.DeclaredFailure("timeline.envelope_type_scope")
		}
		// Historical attribution was staged while preparing; the core flushes
		// it in step 9 and checks the counts.
		return nil
	})
}

// validateEvents compares the stored events with the staged events in one
// ordered pass. A missing, changed or extra event is an exact-state failure.
func (p preparedTimelineImport) validateEvents(ctx context.Context, tx pgx.Tx, store incidentportability.MemberReader, d sourceport.Descriptor) error {
	rows, err := tx.Query(ctx, `SELECT t.record_id,t.incident_id,t.capture_state,t.reviewed_by_user_id,t.reviewed_at,t.superseded_by_user_id,t.superseded_at,
t.date_entered_text,t.analyst_text,t.mitre_stage_text,t.device_object_text,t.ip_address_text,t.activity_utc_text,t.activity_local_text,
t.raw_activity_text,t.activity_synopsis_text,t.data_source_text,t.activity_utc_generated,t.activity_local_generated,t.activity_time_pair_state,
(r.record_id IS NOT NULL AND t.row_version=r.row_version AND t.recorded_at=r.created_at AND t.edited_at=r.updated_at AND t.created_by_user_id=r.created_by_user_id AND t.updated_by_user_id=r.updated_by_user_id)
FROM timeline_events t LEFT JOIN records r ON r.record_id=t.record_id WHERE t.incident_id=$1 ORDER BY t.record_id`, p.incidentID)
	if err != nil {
		return err
	}
	defer rows.Close()
	err = incidentportability.ReadPreparedRows(ctx, store, p.eventRows, timelineEventCodec{}, func(source sourcevalue.Row) error {
		if !rows.Next() {
			if err := rows.Err(); err != nil {
				return err
			}
			return d.DeclaredFailure("timeline.event_state_exact")
		}
		var actual sourcevalue.Row
		var mirrors bool
		if err := rows.Scan(&actual.RecordID, &actual.IncidentID, &actual.CaptureState, &actual.ReviewedByUserID, &actual.ReviewedAt, &actual.SupersededByUserID, &actual.SupersededAt, &actual.DateEnteredText, &actual.AnalystText, &actual.MitreStageText, &actual.DeviceObjectText, &actual.IPAddressText, &actual.ActivityUTCText, &actual.ActivityLocalText, &actual.RawActivityText, &actual.ActivitySynopsisText, &actual.DataSourceText, &actual.ActivityUTCGenerated, &actual.ActivityLocalGenerated, &actual.ActivityTimePairState, &mirrors); err != nil {
			return err
		}
		expected := source
		expected.ReviewedByUserID = p.importedEventActor(source.ReviewedByUserID)
		expected.SupersededByUserID = p.importedEventActor(source.SupersededByUserID)
		actual.NormalizeClocks()
		if !mirrors || !reflect.DeepEqual(actual, expected) {
			return d.DeclaredFailure("timeline.event_state_exact")
		}
		return nil
	})
	if err != nil {
		return err
	}
	if rows.Next() {
		return d.DeclaredFailure("timeline.event_state_exact")
	}
	return rows.Err()
}
