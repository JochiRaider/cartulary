package timeline

import (
	"context"
	"errors"
	"reflect"

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

func (p preparedTimelineImport) preflightEvents(ctx context.Context, tx pgx.Tx, d sourceport.Descriptor) error {
	return p.events(ctx, d, func(row sourcevalue.Row) error {
		var envelope, occupied bool
		if err := tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM records WHERE record_id=$1 AND incident_id=$2 AND record_type='timeline_event'),EXISTS(SELECT 1 FROM timeline_events WHERE record_id=$1)`, row.RecordID, p.incidentID).Scan(&envelope, &occupied); err != nil {
			return err
		}
		if !envelope {
			return d.DeclaredFailure("timeline.envelope_type_scope")
		}
		if occupied {
			return d.DeclaredFailure("timeline.source_identity_admitted")
		}
		return nil
	})
}

func (p preparedTimelineImport) applyEvents(ctx context.Context, tx pgx.Tx, input sourceport.ImportContext, d sourceport.Descriptor) error {
	return p.events(ctx, d, func(row sourcevalue.Row) error {
		var reviewer, superseder *uuid.UUID
		if row.ReviewedByUserID != nil {
			reviewer = &p.actorID
		}
		if row.SupersededByUserID != nil {
			superseder = &p.actorID
		}
		tag, err := tx.Exec(ctx, `INSERT INTO timeline_events (
record_id,incident_id,capture_state,row_version,recorded_at,edited_at,created_by_user_id,updated_by_user_id,
reviewed_by_user_id,reviewed_at,superseded_by_user_id,superseded_at,date_entered_text,analyst_text,mitre_stage_text,
device_object_text,ip_address_text,activity_utc_text,activity_local_text,raw_activity_text,activity_synopsis_text,
data_source_text,activity_utc_generated,activity_local_generated,activity_time_pair_state)
SELECT $1,$2,$3,r.row_version,r.created_at,r.updated_at,r.created_by_user_id,r.updated_by_user_id,
$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20
FROM records r WHERE r.record_id=$1 AND r.incident_id=$2 AND r.record_type='timeline_event'`,
			row.RecordID, row.IncidentID, row.CaptureState, reviewer, row.ReviewedAt, superseder, row.SupersededAt, row.DateEnteredText, row.AnalystText, row.MitreStageText, row.DeviceObjectText, row.IPAddressText, row.ActivityUTCText, row.ActivityLocalText, row.RawActivityText, row.ActivitySynopsisText, row.DataSourceText, row.ActivityUTCGenerated, row.ActivityLocalGenerated, row.ActivityTimePairState)
		if err != nil {
			return err
		}
		if tag.RowsAffected() != 1 {
			return d.DeclaredFailure("timeline.envelope_type_scope")
		}
		for column, id := range map[string]*uuid.UUID{"reviewed_by_user_id": row.ReviewedByUserID, "superseded_by_user_id": row.SupersededByUserID} {
			if id != nil {
				if err := input.Attributions.RecordImportedAttribution("timeline_events", row.RecordID.String(), column, id.String()); err != nil {
					return err
				}
			}
		}
		return nil
	})
}

func (p preparedTimelineImport) validateEvents(ctx context.Context, tx pgx.Tx, input sourceport.ImportContext, d sourceport.Descriptor) error {
	var expectedRows, expectedAttrs int64
	err := p.events(ctx, d, func(source sourcevalue.Row) error {
		expectedRows++
		var actual sourcevalue.Row
		var mirrors bool
		err := tx.QueryRow(ctx, `SELECT t.record_id,t.incident_id,t.capture_state,t.reviewed_by_user_id,t.reviewed_at,t.superseded_by_user_id,t.superseded_at,
t.date_entered_text,t.analyst_text,t.mitre_stage_text,t.device_object_text,t.ip_address_text,t.activity_utc_text,t.activity_local_text,
t.raw_activity_text,t.activity_synopsis_text,t.data_source_text,t.activity_utc_generated,t.activity_local_generated,t.activity_time_pair_state,
(r.record_id IS NOT NULL AND t.row_version=r.row_version AND t.recorded_at=r.created_at AND t.edited_at=r.updated_at AND t.created_by_user_id=r.created_by_user_id AND t.updated_by_user_id=r.updated_by_user_id)
FROM timeline_events t LEFT JOIN records r ON r.record_id=t.record_id WHERE t.record_id=$1 AND t.incident_id=$2`, source.RecordID, p.incidentID).Scan(&actual.RecordID, &actual.IncidentID, &actual.CaptureState, &actual.ReviewedByUserID, &actual.ReviewedAt, &actual.SupersededByUserID, &actual.SupersededAt, &actual.DateEnteredText, &actual.AnalystText, &actual.MitreStageText, &actual.DeviceObjectText, &actual.IPAddressText, &actual.ActivityUTCText, &actual.ActivityLocalText, &actual.RawActivityText, &actual.ActivitySynopsisText, &actual.DataSourceText, &actual.ActivityUTCGenerated, &actual.ActivityLocalGenerated, &actual.ActivityTimePairState, &mirrors)
		if errors.Is(err, pgx.ErrNoRows) {
			return d.DeclaredFailure("timeline.event_state_exact")
		}
		if err != nil {
			return err
		}
		expected := source
		if expected.ReviewedByUserID != nil {
			expected.ReviewedByUserID = &p.actorID
		}
		if expected.SupersededByUserID != nil {
			expected.SupersededByUserID = &p.actorID
		}
		actual.NormalizeClocks()
		if !mirrors || !reflect.DeepEqual(actual, expected) {
			return d.DeclaredFailure("timeline.event_state_exact")
		}
		actors := map[string]uuid.UUID{}
		if source.ReviewedByUserID != nil {
			actors["reviewed_by_user_id"] = *source.ReviewedByUserID
		}
		if source.SupersededByUserID != nil {
			actors["superseded_by_user_id"] = *source.SupersededByUserID
		}
		expectedAttrs += int64(len(actors))
		if err := input.Attributions.VisitImportedAttributions("timeline_events", source.RecordID.String(), func(a incidentportability.ImportedAttribution) error {
			id, ok := actors[a.SourceColumn]
			if !ok || a.SourceActorID != id.String() || a.SourceTable != "timeline_events" || a.SourceRowID != source.RecordID.String() || a.LocalUserID != p.actorID {
				return d.DeclaredFailure("timeline.event_attribution_exact")
			}
			delete(actors, a.SourceColumn)
			return ctx.Err()
		}); err != nil {
			return err
		}
		if len(actors) != 0 {
			return d.DeclaredFailure("timeline.event_attribution_exact")
		}
		return nil
	})
	if err != nil {
		return err
	}
	var actualRows int64
	if err := tx.QueryRow(ctx, `SELECT count(*) FROM timeline_events WHERE incident_id=$1`, p.incidentID).Scan(&actualRows); err != nil {
		return err
	}
	if actualRows != expectedRows {
		return d.DeclaredFailure("timeline.event_state_exact")
	}
	var actualAttrs int64
	if err := input.Attributions.VisitImportedAttributions("timeline_events", "", func(incidentportability.ImportedAttribution) error { actualAttrs++; return ctx.Err() }); err != nil {
		return err
	}
	if actualAttrs != expectedAttrs {
		return d.DeclaredFailure("timeline.event_attribution_exact")
	}
	return nil
}
