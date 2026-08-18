package timeline

import (
	"context"
	"encoding/hex"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/incidentportability"
)

const (
	timelineBundleRecordsPath    = "data/timeline_records.ndjson"
	timelineBundleProfilesPath   = "data/timeline_time_profiles.ndjson"
	timelineBundleProvenancePath = "data/timeline_source_provenance.ndjson"
	timelineBundleV1RecordsPath  = "data/timeline_events.ndjson"
	timelineBundleV1ProfilesPath = "data/timeline_time_conversion_profiles.ndjson"
)

var timelineBundleRecordFields = fieldSet(
	"record_id", "incident_id", "capture_state",
	"reviewed_by_user_id", "reviewed_at",
	"superseded_by_user_id", "superseded_at",
	"date_entered_text", "analyst_text", "mitre_stage_text",
	"device_object_text", "ip_address_text", "activity_utc_text",
	"activity_local_text", "raw_activity_text", "activity_synopsis_text",
	"data_source_text", "activity_utc_generated", "activity_local_generated",
	"activity_time_pair_state",
)

var timelineBundleV1RecordFields = fieldSet(
	"record_id", "incident_id", "capture_state", "row_version",
	"recorded_at", "edited_at", "created_by_user_id", "updated_by_user_id",
	"reviewed_by_user_id", "reviewed_at",
	"superseded_by_user_id", "superseded_at", "raw_capture",
	"date_entered_text", "analyst_text", "mitre_stage_text",
	"device_object_text", "ip_address_text", "activity_utc_text",
	"activity_local_text", "raw_activity_text", "activity_synopsis_text",
	"data_source_text", "activity_utc_generated", "activity_local_generated",
	"activity_time_pair_state",
)

var timelineBundleProfileFields = fieldSet(
	"incident_id", "enabled", "local_offset_minutes", "local_label",
	"profile_version", "updated_at", "updated_by_user_id",
)

var timelineBundleProvenanceFields = fieldSet(
	"record_id", "source_identity_sha256", "source_row_ordinal",
	"source_column_ordinal", "source_kind", "source_metadata",
	"source_header", "raw_value", "cell_kind", "created_at",
)

func ExportIncidentBundleFiles(ctx context.Context, q incidentportability.Queryer, incidentID uuid.UUID) ([]incidentportability.File, error) {
	specs := []struct {
		path  string
		query string
	}{
		{timelineBundleProfilesPath, `
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
 WHERE profile.incident_id = $1
 ORDER BY profile.incident_id`},
		{timelineBundleRecordsPath, `
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
 WHERE event.incident_id = $1
 ORDER BY event.record_id`},
		{timelineBundleProvenancePath, `
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
          provenance.source_column_ordinal, provenance.source_identity_hash`},
	}
	files := make([]incidentportability.File, 0, len(specs))
	for _, spec := range specs {
		file, err := incidentportability.ExportNDJSON(ctx, q, incidentID, spec.path, spec.query)
		if err != nil {
			return nil, err
		}
		files = append(files, file)
	}
	return files, nil
}

func ImportIncidentBundleFilesTx(
	ctx context.Context,
	tx pgx.Tx,
	files map[string][]byte,
	bundleVersion int,
	actorUserID uuid.UUID,
	attributions incidentportability.AttributionRecorder,
) error {
	switch bundleVersion {
	case 2:
		if err := importTimelineProfilesTx(ctx, tx, timelineBundleProfilesPath, files[timelineBundleProfilesPath], actorUserID, attributions); err != nil {
			return err
		}
		if err := importTimelineRecordsV2Tx(ctx, tx, files[timelineBundleRecordsPath], actorUserID, attributions); err != nil {
			return err
		}
		return importTimelineProvenanceV2Tx(ctx, tx, files[timelineBundleProvenancePath])
	case 1:
		if err := importTimelineProfilesTx(ctx, tx, timelineBundleV1ProfilesPath, files[timelineBundleV1ProfilesPath], actorUserID, attributions); err != nil {
			return err
		}
		return importTimelineRecordsV1Tx(ctx, tx, files[timelineBundleV1RecordsPath], actorUserID, attributions)
	default:
		return malformedTimelineBundle()
	}
}

func importTimelineProfilesTx(
	ctx context.Context,
	tx pgx.Tx,
	logicalPath string,
	payload []byte,
	actorUserID uuid.UUID,
	attributions incidentportability.AttributionRecorder,
) error {
	rows, err := decodeLogicalRows(payload)
	if err != nil {
		return err
	}
	for _, row := range rows {
		if err := requireExactFields(row, timelineBundleProfileFields, "incident_id", "enabled", "profile_version", "updated_at"); err != nil {
			return err
		}
		if err := incidentportability.RemapTopLevelUserFields(row, "timeline_time_conversion_profiles", []string{"incident_id"}, actorUserID, attributions); err != nil {
			return err
		}
		raw, err := json.Marshal(row)
		if err != nil {
			return err
		}
		tag, err := tx.Exec(ctx, `
INSERT INTO timeline_time_conversion_profiles (
    incident_id, enabled, local_offset_minutes, local_label,
    profile_version, updated_at, updated_by_user_id
)
SELECT
    (payload ->> 'incident_id')::uuid,
    (payload ->> 'enabled')::boolean,
    NULLIF(payload ->> 'local_offset_minutes', '')::integer,
    payload ->> 'local_label',
    (payload ->> 'profile_version')::bigint,
    (payload ->> 'updated_at')::timestamp with time zone,
    NULLIF(payload ->> 'updated_by_user_id', '')::uuid
  FROM (SELECT $1::jsonb AS payload) AS input
`, raw)
		if err != nil {
			return err
		}
		if tag.RowsAffected() != 1 {
			return incidentportability.FixedImportFailure(logicalPath)
		}
	}
	return nil
}

func importTimelineRecordsV2Tx(
	ctx context.Context,
	tx pgx.Tx,
	payload []byte,
	actorUserID uuid.UUID,
	attributions incidentportability.AttributionRecorder,
) error {
	rows, err := decodeLogicalRows(payload)
	if err != nil {
		return err
	}
	for _, row := range rows {
		if err := requireExactFields(row, timelineBundleRecordFields, "record_id", "incident_id", "capture_state", "activity_utc_generated", "activity_local_generated", "activity_time_pair_state"); err != nil {
			return err
		}
		if err := incidentportability.RemapTopLevelUserFields(row, "timeline_events", []string{"record_id"}, actorUserID, attributions); err != nil {
			return err
		}
		raw, err := json.Marshal(row)
		if err != nil {
			return err
		}
		tag, err := tx.Exec(ctx, timelineLogicalRecordInsertSQL, raw)
		if err != nil {
			return err
		}
		if tag.RowsAffected() != 1 {
			return incidentportability.FixedImportFailure(timelineBundleRecordsPath)
		}
	}
	return nil
}

func importTimelineRecordsV1Tx(
	ctx context.Context,
	tx pgx.Tx,
	payload []byte,
	actorUserID uuid.UUID,
	attributions incidentportability.AttributionRecorder,
) error {
	rows, err := decodeLogicalRows(payload)
	if err != nil {
		return err
	}
	for _, row := range rows {
		if err := requireExactFields(row, timelineBundleV1RecordFields, "record_id", "incident_id", "capture_state", "row_version", "recorded_at", "edited_at", "created_by_user_id", "updated_by_user_id", "raw_capture"); err != nil {
			return err
		}
		rawCapture, err := decodeLegacyRawCapture(row["raw_capture"])
		if err != nil {
			return err
		}
		delete(row, "raw_capture")
		if err := incidentportability.RemapTopLevelUserFields(row, "timeline_events", []string{"record_id"}, actorUserID, attributions); err != nil {
			return err
		}
		raw, err := json.Marshal(row)
		if err != nil {
			return err
		}
		tag, err := tx.Exec(ctx, timelineLegacyRecordInsertSQL, raw)
		if err != nil {
			return err
		}
		if tag.RowsAffected() != 1 {
			return incidentportability.FixedImportFailure(timelineBundleV1RecordsPath)
		}
		recordID, err := uuid.Parse(incidentportability.StringFromAny(row["record_id"]))
		if err != nil {
			return malformedTimelineBundle()
		}
		if err := insertSourceProvenanceTx(ctx, tx, recordID, rawCapture); err != nil {
			return err
		}
	}
	return nil
}

func importTimelineProvenanceV2Tx(ctx context.Context, tx pgx.Tx, payload []byte) error {
	rows, err := decodeLogicalRows(payload)
	if err != nil {
		return err
	}
	for _, row := range rows {
		if err := requireExactFields(row, timelineBundleProvenanceFields,
			"record_id", "source_identity_sha256", "source_row_ordinal",
			"source_column_ordinal", "source_kind", "source_metadata",
			"source_header", "raw_value", "created_at",
		); err != nil {
			return err
		}
		identity, err := hex.DecodeString(incidentportability.StringFromAny(row["source_identity_sha256"]))
		if err != nil || len(identity) != sha256DigestBytes {
			return malformedTimelineBundle()
		}
		raw, err := json.Marshal(row)
		if err != nil {
			return err
		}
		tag, err := tx.Exec(ctx, `
INSERT INTO timeline_source_provenance (
    record_id, source_identity_hash, source_row_ordinal,
    source_column_ordinal, source_kind, source_metadata,
    source_header_json, raw_value, cell_kind, created_at
)
SELECT
    (payload ->> 'record_id')::uuid,
    $2::bytea,
    (payload ->> 'source_row_ordinal')::integer,
    (payload ->> 'source_column_ordinal')::integer,
    payload ->> 'source_kind',
    payload -> 'source_metadata',
    payload -> 'source_header',
    payload ->> 'raw_value',
    NULLIF(payload ->> 'cell_kind', ''),
    (payload ->> 'created_at')::timestamp with time zone
  FROM (SELECT $1::jsonb AS payload) AS input
`, raw, identity)
		if err != nil {
			return err
		}
		if tag.RowsAffected() != 1 {
			return incidentportability.FixedImportFailure(timelineBundleProvenancePath)
		}
	}
	return nil
}

func decodeLegacyRawCapture(value any) ([]ClipboardRawImportColumn, error) {
	raw, err := json.Marshal(value)
	if err != nil {
		return nil, malformedTimelineBundle()
	}
	var object map[string]json.RawMessage
	if err := json.Unmarshal(raw, &object); err != nil {
		return nil, malformedTimelineBundle()
	}
	if len(object) == 0 {
		return nil, nil
	}
	if len(object) != 1 {
		return nil, malformedTimelineBundle()
	}
	columns, ok := object["import_columns"]
	if !ok {
		return nil, malformedTimelineBundle()
	}
	var result []ClipboardRawImportColumn
	if err := json.Unmarshal(columns, &result); err != nil {
		return nil, malformedTimelineBundle()
	}
	return result, nil
}

func decodeLogicalRows(payload []byte) ([]map[string]any, error) {
	return incidentportability.DecodeNDJSON(payload)
}

func fieldSet(fields ...string) map[string]struct{} {
	result := make(map[string]struct{}, len(fields))
	for _, field := range fields {
		result[field] = struct{}{}
	}
	return result
}

func requireExactFields(row map[string]any, allowed map[string]struct{}, required ...string) error {
	for field := range row {
		if _, ok := allowed[field]; !ok {
			return malformedTimelineBundle()
		}
	}
	for _, field := range required {
		if _, ok := row[field]; !ok {
			return malformedTimelineBundle()
		}
	}
	return nil
}

func malformedTimelineBundle() error {
	return &incidentportability.VerificationFailure{ReasonCode: "malformed_manifest"}
}

const sha256DigestBytes = 32

const timelineLogicalRecordInsertSQL = `
INSERT INTO timeline_events (
    record_id, incident_id, capture_state, row_version,
    recorded_at, edited_at, created_by_user_id, updated_by_user_id,
    reviewed_by_user_id, reviewed_at, superseded_by_user_id, superseded_at,
    date_entered_text, analyst_text, mitre_stage_text, device_object_text,
    ip_address_text, activity_utc_text, activity_local_text, raw_activity_text,
    activity_synopsis_text, data_source_text, activity_utc_generated,
    activity_local_generated, activity_time_pair_state
)
SELECT
    (payload ->> 'record_id')::uuid,
    (payload ->> 'incident_id')::uuid,
    payload ->> 'capture_state',
    record.row_version,
    record.created_at,
    record.updated_at,
    record.created_by_user_id,
    record.updated_by_user_id,
    NULLIF(payload ->> 'reviewed_by_user_id', '')::uuid,
    NULLIF(payload ->> 'reviewed_at', '')::timestamp with time zone,
    NULLIF(payload ->> 'superseded_by_user_id', '')::uuid,
    NULLIF(payload ->> 'superseded_at', '')::timestamp with time zone,
    payload ->> 'date_entered_text',
    payload ->> 'analyst_text',
    payload ->> 'mitre_stage_text',
    payload ->> 'device_object_text',
    payload ->> 'ip_address_text',
    payload ->> 'activity_utc_text',
    payload ->> 'activity_local_text',
    payload ->> 'raw_activity_text',
    payload ->> 'activity_synopsis_text',
    payload ->> 'data_source_text',
    (payload ->> 'activity_utc_generated')::boolean,
    (payload ->> 'activity_local_generated')::boolean,
    payload ->> 'activity_time_pair_state'
  FROM (SELECT $1::jsonb AS payload) AS input
  JOIN records AS record
    ON record.record_id = (payload ->> 'record_id')::uuid
   AND record.incident_id = (payload ->> 'incident_id')::uuid
   AND record.record_type = 'timeline_event'
`

const timelineLegacyRecordInsertSQL = `
INSERT INTO timeline_events (
    record_id, incident_id, capture_state, row_version,
    recorded_at, edited_at, created_by_user_id, updated_by_user_id,
    reviewed_by_user_id, reviewed_at, superseded_by_user_id, superseded_at,
    date_entered_text, analyst_text, mitre_stage_text, device_object_text,
    ip_address_text, activity_utc_text, activity_local_text, raw_activity_text,
    activity_synopsis_text, data_source_text, activity_utc_generated,
    activity_local_generated, activity_time_pair_state
)
SELECT
    (payload ->> 'record_id')::uuid,
    (payload ->> 'incident_id')::uuid,
    payload ->> 'capture_state',
    (payload ->> 'row_version')::bigint,
    (payload ->> 'recorded_at')::timestamp with time zone,
    (payload ->> 'edited_at')::timestamp with time zone,
    (payload ->> 'created_by_user_id')::uuid,
    (payload ->> 'updated_by_user_id')::uuid,
    NULLIF(payload ->> 'reviewed_by_user_id', '')::uuid,
    NULLIF(payload ->> 'reviewed_at', '')::timestamp with time zone,
    NULLIF(payload ->> 'superseded_by_user_id', '')::uuid,
    NULLIF(payload ->> 'superseded_at', '')::timestamp with time zone,
    payload ->> 'date_entered_text',
    payload ->> 'analyst_text',
    payload ->> 'mitre_stage_text',
    payload ->> 'device_object_text',
    payload ->> 'ip_address_text',
    payload ->> 'activity_utc_text',
    payload ->> 'activity_local_text',
    payload ->> 'raw_activity_text',
    payload ->> 'activity_synopsis_text',
    payload ->> 'data_source_text',
    (payload ->> 'activity_utc_generated')::boolean,
    (payload ->> 'activity_local_generated')::boolean,
    payload ->> 'activity_time_pair_state'
  FROM (SELECT $1::jsonb AS payload) AS input
`
