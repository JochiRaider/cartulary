package reportingprovider

import (
	"context"
	"time"

	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/reporting/exportprovider"
)

func CollectFactsTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, supportRefs map[string][]string) (exportprovider.ProviderOutput, error) {
	rows, err := tx.Query(ctx, `SELECT t.record_id::text, 'timeline_event'::text, r.record_type::text, 'source_evidence'::text, jsonb_build_object(
       'record_id', t.record_id,
       'capture_state', t.capture_state,
       'row_version', t.row_version,
       'recorded_at', t.recorded_at,
       'edited_at', t.edited_at,
       'created_by_user_id', t.created_by_user_id,
       'updated_by_user_id', t.updated_by_user_id,
       'reviewed_by_user_id', t.reviewed_by_user_id,
       'reviewed_at', t.reviewed_at,
       'superseded_by_user_id', t.superseded_by_user_id,
       'superseded_at', t.superseded_at,
       'date_entered_text', t.date_entered_text,
       'analyst_text', t.analyst_text,
       'mitre_stage_text', t.mitre_stage_text,
       'device_object_text', t.device_object_text,
       'ip_address_text', t.ip_address_text,
       'activity_utc_text', t.activity_utc_text,
       'activity_local_text', t.activity_local_text,
       'raw_activity_text', t.raw_activity_text,
       'activity_synopsis_text', t.activity_synopsis_text,
       'data_source_text', t.data_source_text,
       'activity_utc_generated', t.activity_utc_generated,
       'activity_local_generated', t.activity_local_generated,
       'activity_time_pair_state', t.activity_time_pair_state
   ), r.created_at, t.activity_utc_text, t.activity_local_text, t.date_entered_text
  FROM timeline_events t
  JOIN records r ON r.incident_id = t.incident_id AND r.record_id = t.record_id AND r.deleted_at IS NULL
 WHERE t.incident_id = $1`, incidentID)
	if err != nil {
		return exportprovider.ProviderOutput{}, err
	}
	defer rows.Close()
	fields := []exportprovider.BoundField{}
	timeline := []exportprovider.TimelineSortFact{}
	for rows.Next() {
		var id, family, recordType, contentClass string
		var raw []byte
		var createdAt time.Time
		var utcText, localText, dateText *string
		if err := rows.Scan(&id, &family, &recordType, &contentClass, &raw, &createdAt, &utcText, &localText, &dateText); err != nil {
			return exportprovider.ProviderOutput{}, err
		}
		value, err := canonicaljson.DecodeStrict(raw)
		if err != nil {
			return exportprovider.ProviderOutput{}, err
		}
		fact, err := DeriveSortFact(id, createdAt, utcText, localText, dateText)
		if err != nil {
			return exportprovider.ProviderOutput{}, err
		}
		timeline = append(timeline, fact)
		fields = append(fields, exportprovider.BoundField{RecordID: id, RecordType: recordType, FieldFact: exportprovider.FieldFact{
			SchemaID: exportprovider.FieldFactSchemaID, Path: "/timeline/" + id, SourceFamily: family, ContentClass: contentClass, Value: value, SupportRefs: exportprovider.CloneStrings(supportRefs[id]),
		}})
	}
	if err := rows.Err(); err != nil {
		return exportprovider.ProviderOutput{}, err
	}
	return exportprovider.NewProviderOutput("timeline", fields, timeline)
}
