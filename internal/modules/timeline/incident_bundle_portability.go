package timeline

import (
	"context"

	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
	"github.com/JochiRaider/cartulary/internal/modules/incidentportability"
	"github.com/JochiRaider/cartulary/internal/platform/sqlshape"
)

const (
	timelineBundleRecordsPath    = "data/timeline_records.ndjson"
	timelineBundleProfilesPath   = "data/timeline_time_profiles.ndjson"
	timelineBundleProvenancePath = "data/timeline_source_provenance.ndjson"
)

func ExportIncidentBundleFiles(ctx context.Context, input sourceport.ExportContext) error {
	specs := []struct {
		path  string
		query sqlshape.Fragment
	}{
		{timelineBundleProfilesPath, sqlshape.Literal(`
SELECT jsonb_build_object(
    'incident_id', profile.incident_id,
    'enabled', profile.enabled,
    'local_offset_minutes', profile.local_offset_minutes,
    'local_label', profile.local_label,
    'profile_version', profile.profile_version,
    'updated_at', profile.updated_at,
    'updated_by_user_id', profile.updated_by_user_id
)
  FROM timeline_time_conversion_profiles AS profile
 WHERE profile.incident_id = $1 AND ($2::uuid IS NULL OR profile.incident_id>$2)
 ORDER BY profile.incident_id LIMIT 128`)},
		{timelineBundleRecordsPath, sqlshape.Literal(`
SELECT jsonb_build_object(
    'record_id', event.record_id,
    'incident_id', event.incident_id,
    'capture_state', event.capture_state,
    'reviewed_by_user_id', event.reviewed_by_user_id,
    'reviewed_at', event.reviewed_at,
    'superseded_by_user_id', event.superseded_by_user_id,
    'superseded_at', event.superseded_at,
    'date_entered_text', event.date_entered_text,
    'analyst_text', event.analyst_text,
    'mitre_stage_text', event.mitre_stage_text,
    'device_object_text', event.device_object_text,
    'ip_address_text', event.ip_address_text,
    'activity_utc_text', event.activity_utc_text,
    'activity_local_text', event.activity_local_text,
    'raw_activity_text', event.raw_activity_text,
    'activity_synopsis_text', event.activity_synopsis_text,
    'data_source_text', event.data_source_text,
    'activity_utc_generated', event.activity_utc_generated,
    'activity_local_generated', event.activity_local_generated,
    'activity_time_pair_state', event.activity_time_pair_state
)
  FROM timeline_events AS event
 WHERE event.incident_id = $1 AND ($2::uuid IS NULL OR event.record_id>$2)
 ORDER BY event.record_id LIMIT 128`)},
		{timelineBundleProvenancePath, sqlshape.Literal(`
SELECT jsonb_build_object(
    'record_id', provenance.record_id,
    'source_identity_sha256', encode(provenance.source_identity_hash, 'hex'),
    'source_row_ordinal', provenance.source_row_ordinal,
    'source_column_ordinal', provenance.source_column_ordinal,
    'source_kind', provenance.source_kind,
    'source_metadata', provenance.source_metadata,
    'source_header', provenance.source_header_json,
    'raw_value', provenance.raw_value,
    'cell_kind', provenance.cell_kind,
    'created_at', provenance.created_at
)
  FROM timeline_source_provenance AS provenance
  JOIN timeline_events AS event ON event.record_id = provenance.record_id
 WHERE event.incident_id = $1
 ORDER BY provenance.record_id, provenance.source_row_ordinal,
          provenance.source_column_ordinal, provenance.source_identity_hash`)},
	}
	for _, spec := range specs {
		var err error
		switch spec.path {
		case timelineBundleRecordsPath:
			err = sourceport.ExportAttributedRows(ctx, input, sourceport.AttributedRows{Path: spec.path, Table: "timeline_events", Identity: "record_id", ActorColumns: []string{"reviewed_by_user_id", "superseded_by_user_id"}, Query: spec.query})
		case timelineBundleProfilesPath:
			err = sourceport.ExportAttributedRows(ctx, input, sourceport.AttributedRows{Path: spec.path, Table: "timeline_time_conversion_profiles", Identity: "incident_id", ActorColumns: []string{"updated_by_user_id"}, Query: spec.query})
		default:
			err = incidentportability.ExportNDJSONTo(ctx, input.Query, input.IncidentID, spec.path, spec.query, input.Members)
		}
		if err != nil {
			return err
		}
	}
	return nil
}
