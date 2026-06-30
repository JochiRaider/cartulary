package timeline

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"

	sqlc "github.com/JochiRaider/cartulary/internal/gen/sql"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
)

func (s *store) QueryRows(ctx context.Context, incidentID uuid.UUID, query viewschema.QueryMeta) ([]map[string]any, error) {
	sqlText, args, err := buildTimelineQuerySQL(incidentID, query)
	if err != nil {
		return nil, err
	}
	rows, err := s.pool.Query(ctx, sqlText, args...)
	if err != nil {
		return nil, fmt.Errorf("list timeline projection rows: %w", err)
	}
	defer rows.Close()

	projectedRows := make([]projectedRecord, 0)
	for rows.Next() {
		projected, err := scanProjectedRecord(rows)
		if err != nil {
			return nil, err
		}
		projectedRows = append(projectedRows, projected)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate timeline projection rows: %w", err)
	}
	rows.Close()

	result := make([]map[string]any, 0, len(projectedRows))
	for index := range projectedRows {
		if err := hydrateProjectedCollections(ctx, s.pool, &projectedRows[index]); err != nil {
			return nil, err
		}
		result = append(result, buildRow(projectedRows[index]))
	}
	return result, nil
}

func projectRecord(record sourceRecord, replacementRecordID *uuid.UUID) projectedRecord {
	return projectedRecord{
		RecordID:              record.RecordID,
		IncidentID:            record.IncidentID,
		RowVersion:            record.RowVersion,
		DateEnteredText:       cloneStringPointer(record.DateEnteredText),
		AnalystText:           cloneStringPointer(record.AnalystText),
		MitreStageText:        cloneStringPointer(record.MitreStageText),
		DeviceObjectText:      cloneStringPointer(record.DeviceObjectText),
		IPAddressText:         cloneStringPointer(record.IPAddressText),
		ActivityUTCText:       cloneStringPointer(record.ActivityUTCText),
		ActivityLocalText:     cloneStringPointer(record.ActivityLocalText),
		RawActivityText:       cloneStringPointer(record.RawActivityText),
		ActivitySynopsisText:  cloneStringPointer(record.ActivitySynopsisText),
		DataSourceText:        cloneStringPointer(record.DataSourceText),
		RecordedAt:            record.RecordedAt.UTC(),
		EditedAt:              record.EditedAt.UTC(),
		ActivitySortTS:        deriveActivitySortTS(record.ActivityUTCText, record.ActivityLocalText),
		DateEnteredSortDay:    deriveDateEnteredSortDay(record.DateEnteredText),
		ActivityTimePairState: record.ActivityTimePairState,
		CaptureState:          record.CaptureState,
		ReplacementRecordID:   replacementRecordID,
		EvidenceCount:         0,
		HasEvidence:           false,
		HasUnresolvedMentions: false,
	}
}

func projectionInput(record projectedRecord) timelineProjectionInput {
	return timelineProjectionInput{
		RecordID:              record.RecordID,
		IncidentID:            record.IncidentID,
		RowVersion:            record.RowVersion,
		DateEnteredText:       record.DateEnteredText,
		AnalystText:           record.AnalystText,
		MitreStageText:        record.MitreStageText,
		DeviceObjectText:      record.DeviceObjectText,
		IPAddressText:         record.IPAddressText,
		ActivityUTCText:       record.ActivityUTCText,
		ActivityLocalText:     record.ActivityLocalText,
		RawActivityText:       record.RawActivityText,
		ActivitySynopsisText:  record.ActivitySynopsisText,
		DataSourceText:        record.DataSourceText,
		RecordedAt:            record.RecordedAt,
		EditedAt:              record.EditedAt,
		ActivitySortTS:        record.ActivitySortTS,
		DateEnteredSortDay:    record.DateEnteredSortDay,
		ActivityTimePairState: record.ActivityTimePairState,
		CaptureState:          record.CaptureState,
		ReplacementRecordID:   record.ReplacementRecordID,
		EvidenceCount:         record.EvidenceCount,
		HasEvidence:           record.HasEvidence,
		HasUnresolvedMentions: record.HasUnresolvedMentions,
	}
}

func projectedRecordFromSQL(row sqlc.GetTimelineProjectionRowRow) (projectedRecord, error) {
	recordID, err := uuidFromPG(row.RecordID)
	if err != nil {
		return projectedRecord{}, err
	}
	incidentID, err := uuidFromPG(row.IncidentID)
	if err != nil {
		return projectedRecord{}, err
	}
	recordedAt, err := timeFromPG(row.RecordedAt)
	if err != nil {
		return projectedRecord{}, err
	}
	editedAt, err := timeFromPG(row.EditedAt)
	if err != nil {
		return projectedRecord{}, err
	}
	return projectedRecord{
		RecordID:              recordID,
		IncidentID:            incidentID,
		RowVersion:            row.RowVersion,
		DateEnteredText:       optionalTextFromPG(row.DateEnteredText),
		AnalystText:           optionalTextFromPG(row.AnalystText),
		MitreStageText:        optionalTextFromPG(row.MitreStageText),
		DeviceObjectText:      optionalTextFromPG(row.DeviceObjectText),
		IPAddressText:         optionalTextFromPG(row.IpAddressText),
		ActivityUTCText:       optionalTextFromPG(row.ActivityUtcText),
		ActivityLocalText:     optionalTextFromPG(row.ActivityLocalText),
		RawActivityText:       optionalTextFromPG(row.RawActivityText),
		ActivitySynopsisText:  optionalTextFromPG(row.ActivitySynopsisText),
		DataSourceText:        optionalTextFromPG(row.DataSourceText),
		RecordedAt:            recordedAt,
		EditedAt:              editedAt,
		ActivitySortTS:        optionalTimeFromPG(row.ActivitySortTs),
		DateEnteredSortDay:    optionalDateFromPG(row.DateEnteredSortDay),
		ActivityTimePairState: row.ActivityTimePairState,
		CaptureState:          row.CaptureState,
		ReplacementRecordID:   optionalUUIDFromPG(row.ReplacementRecordID),
		EvidenceCount:         int(row.EvidenceCount),
		HasEvidence:           row.HasEvidence,
		HasUnresolvedMentions: row.HasUnresolvedMentions,
	}, nil
}

func scanProjectedRecord(scanner interface {
	Scan(dest ...any) error
}) (projectedRecord, error) {
	var row sqlc.GetTimelineProjectionRowRow
	if err := scanner.Scan(
		&row.RecordID,
		&row.IncidentID,
		&row.RowVersion,
		&row.DateEnteredText,
		&row.AnalystText,
		&row.MitreStageText,
		&row.DeviceObjectText,
		&row.IpAddressText,
		&row.ActivityUtcText,
		&row.ActivityLocalText,
		&row.RawActivityText,
		&row.ActivitySynopsisText,
		&row.DataSourceText,
		&row.RecordedAt,
		&row.EditedAt,
		&row.ActivitySortTs,
		&row.DateEnteredSortDay,
		&row.ActivityTimePairState,
		&row.CaptureState,
		&row.ReplacementRecordID,
		&row.EvidenceCount,
		&row.HasEvidence,
		&row.HasUnresolvedMentions,
	); err != nil {
		return projectedRecord{}, fmt.Errorf("scan timeline projection row: %w", err)
	}
	return projectedRecordFromSQL(row)
}

var timelineSortExpressions = map[string]string{
	"record_id":                        "t.record_id",
	"timeline.activity_sort_ts":        "t.activity_sort_ts",
	"timeline.date_entered_sort_day":   "t.date_entered_sort_day",
	"timeline.activity_synopsis_text":  "t.activity_synopsis_text",
	"timeline.analyst_text":            "t.analyst_text",
	"timeline.mitre_stage_text":        "t.mitre_stage_text",
	"timeline.device_object_text":      "t.device_object_text",
	"timeline.ip_address_text":         "t.ip_address_text",
	"timeline.data_source_text":        "t.data_source_text",
	"timeline.edited_at":               "t.edited_at",
	"timeline.capture_state":           "t.capture_state",
	"timeline.evidence_count":          "t.evidence_count",
	"timeline.has_evidence":            "t.has_evidence",
	"timeline.has_unresolved_mentions": "t.has_unresolved_mentions",
}

func buildTimelineQuerySQL(incidentID uuid.UUID, query viewschema.QueryMeta) (string, []any, error) {
	var builder strings.Builder
	builder.WriteString(`
SELECT
    t.record_id,
    t.incident_id,
    r.row_version,
    t.date_entered_text,
    t.analyst_text,
    t.mitre_stage_text,
    t.device_object_text,
    t.ip_address_text,
    t.activity_utc_text,
    t.activity_local_text,
    t.raw_activity_text,
    t.activity_synopsis_text,
    t.data_source_text,
    t.recorded_at,
    t.edited_at,
    t.activity_sort_ts,
    t.date_entered_sort_day,
    t.activity_time_pair_state,
    t.capture_state,
    t.replacement_record_id,
    t.evidence_count,
    t.has_evidence,
    t.has_unresolved_mentions
  FROM timeline_grid_projection t
  JOIN records r
    ON r.record_id = t.record_id
 WHERE t.incident_id = $1
   AND r.deleted_at IS NULL`)

	args := []any{incidentID}
	for _, filter := range query.Filters {
		if err := appendTimelineFilter(&builder, &args, filter); err != nil {
			return "", nil, err
		}
	}

	builder.WriteString(" ORDER BY ")
	for index, sortEntry := range query.Sort {
		if index > 0 {
			builder.WriteString(", ")
		}
		expr, ok := timelineSortExpressions[sortEntry.FieldKey]
		if !ok {
			return "", nil, fmt.Errorf("timeline query sort field %q not mapped", sortEntry.FieldKey)
		}
		builder.WriteString(expr)
		if sortEntry.Direction == "desc" {
			builder.WriteString(" DESC")
		} else {
			builder.WriteString(" ASC")
		}
		builder.WriteString(" NULLS LAST")
	}

	return builder.String(), args, nil
}

func appendTimelineFilter(builder *strings.Builder, args *[]any, filter viewschema.Filter) error {
	switch filter.FieldKey {
	case "timeline.date_entered_sort_day":
		return appendDateFilterClause(builder, args, "t.date_entered_sort_day", filter)
	case "timeline.activity_time_pair_state":
		return appendStringFilterClause(builder, args, "t.activity_time_pair_state", filter)
	case "timeline.capture_state":
		return appendStringFilterClause(builder, args, "t.capture_state", filter)
	case "timeline.has_evidence":
		return appendBoolFilterClause(builder, args, "t.has_evidence", filter)
	case "timeline.has_unresolved_mentions":
		return appendBoolFilterClause(builder, args, "t.has_unresolved_mentions", filter)
	case "timeline.tags":
		return appendTimelineTagFilterClause(builder, args, filter)
	default:
		return fmt.Errorf("timeline query filter field %q not mapped", filter.FieldKey)
	}
}

func appendTimelineTagFilterClause(builder *strings.Builder, args *[]any, filter viewschema.Filter) error {
	values := stringValues(filter.Arg["values"])
	switch filter.Op {
	case "contains_any":
		builder.WriteString(`
   AND EXISTS (
        SELECT 1
          FROM record_tags rt
         WHERE rt.incident_id = t.incident_id
           AND rt.record_id = t.record_id
           AND rt.deleted_at IS NULL
           AND rt.normalized_tag_name = ANY(`)
		builder.WriteString(bindWithCast(args, values, "text[]"))
		builder.WriteString(`)
   )`)
		return nil
	case "contains_all":
		for _, value := range values {
			builder.WriteString(`
   AND EXISTS (
        SELECT 1
          FROM record_tags rt
         WHERE rt.incident_id = t.incident_id
           AND rt.record_id = t.record_id
           AND rt.deleted_at IS NULL
           AND rt.normalized_tag_name = `)
			builder.WriteString(bind(args, value))
			builder.WriteString(`
   )`)
		}
		return nil
	default:
		return fmt.Errorf("timeline tag operator %q not mapped", filter.Op)
	}
}

func appendStringFilterClause(builder *strings.Builder, args *[]any, expr string, filter viewschema.Filter) error {
	switch filter.Op {
	case "eq":
		return appendEqualityClause(builder, args, expr, filter.Arg)
	case "prefix":
		value, _ := filter.Arg["value"].(string)
		builder.WriteString("\n   AND left(lower(coalesce((")
		builder.WriteString(expr)
		builder.WriteString(")::text, '')), char_length(")
		builder.WriteString(bindWithCast(args, value, ""))
		builder.WriteString(")) = ")
		builder.WriteString(bindWithCast(args, value, ""))
		return nil
	default:
		return fmt.Errorf("string filter operator %q not mapped", filter.Op)
	}
}

func appendBoolFilterClause(builder *strings.Builder, args *[]any, expr string, filter viewschema.Filter) error {
	switch filter.Op {
	case "eq":
		return appendEqualityClause(builder, args, expr, filter.Arg)
	default:
		return fmt.Errorf("bool filter operator %q not mapped", filter.Op)
	}
}

func appendDateFilterClause(builder *strings.Builder, args *[]any, expr string, filter viewschema.Filter) error {
	switch filter.Op {
	case "eq":
		return appendEqualityClauseWithCast(builder, args, expr, filter.Arg, "date")
	case "range":
		return appendRangeClause(builder, args, expr, filter.Arg, "date")
	default:
		return fmt.Errorf("date filter operator %q not mapped", filter.Op)
	}
}

func appendEqualityClause(builder *strings.Builder, args *[]any, expr string, arg map[string]any) error {
	return appendEqualityClauseWithCast(builder, args, expr, arg, "")
}

func appendEqualityClauseWithCast(builder *strings.Builder, args *[]any, expr string, arg map[string]any, cast string) error {
	if value, ok := arg["value"]; ok {
		if value == nil {
			builder.WriteString("\n   AND ")
			builder.WriteString(expr)
			builder.WriteString(" IS NULL")
			return nil
		}
		builder.WriteString("\n   AND ")
		builder.WriteString(expr)
		builder.WriteString(" = ")
		builder.WriteString(bindWithCast(args, value, cast))
		return nil
	}

	values, ok := arg["values"].([]any)
	if !ok || len(values) == 0 {
		return fmt.Errorf("missing equality values for %s", expr)
	}
	builder.WriteString("\n   AND (")
	for index, value := range values {
		if index > 0 {
			builder.WriteString(" OR ")
		}
		builder.WriteString(expr)
		builder.WriteString(" = ")
		builder.WriteString(bindWithCast(args, value, cast))
	}
	builder.WriteString(")")
	return nil
}

func appendRangeClause(builder *strings.Builder, args *[]any, expr string, arg map[string]any, cast string) error {
	for _, bound := range []struct {
		Key string
		Op  string
	}{
		{Key: "gt", Op: ">"},
		{Key: "gte", Op: ">="},
		{Key: "lt", Op: "<"},
		{Key: "lte", Op: "<="},
	} {
		value, ok := arg[bound.Key]
		if !ok {
			continue
		}
		builder.WriteString("\n   AND ")
		builder.WriteString(expr)
		builder.WriteByte(' ')
		builder.WriteString(bound.Op)
		builder.WriteByte(' ')
		builder.WriteString(bindWithCast(args, value, cast))
	}
	return nil
}

func bind(args *[]any, value any) string {
	return bindWithCast(args, value, "")
}

func bindWithCast(args *[]any, value any, cast string) string {
	*args = append(*args, value)
	placeholder := fmt.Sprintf("$%d", len(*args))
	if cast == "" {
		return placeholder
	}
	return placeholder + "::" + cast
}

func stringValues(raw any) []string {
	items, ok := raw.([]any)
	if !ok {
		return nil
	}
	values := make([]string, 0, len(items))
	for _, item := range items {
		value, ok := item.(string)
		if !ok {
			continue
		}
		values = append(values, value)
	}
	return values
}
