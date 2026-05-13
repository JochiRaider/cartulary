package workbook

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/entities"
	"github.com/JochiRaider/cartulary/internal/modules/records"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	"github.com/JochiRaider/cartulary/internal/modules/timeline"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
)

const (
	AssessmentsViewSchemaID  = "cartulary.view.assessments.v1"
	CommLogViewSchemaID      = "cartulary.view.comm_log.v1"
	DecisionsViewSchemaID    = "cartulary.view.decisions.v1"
	EvidenceViewSchemaID     = "cartulary.view.evidence.v1"
	HandoffViewSchemaID      = "cartulary.view.handoff.v1"
	LessonViewSchemaID       = "cartulary.view.lesson.v1"
	NotesViewSchemaID        = "cartulary.view.notes.v1"
	PartiesViewSchemaID      = "cartulary.view.parties.v1"
	StatusReviewViewSchemaID = "cartulary.view.status_review.v1"
	TaskRequestsViewSchemaID = "cartulary.view.task_requests.v1"
)

type Store struct {
	pool          postgres.DB
	authStore     *authn.Store
	recordStore   *records.Store
	revisionStore *revisions.Store
	timelineStore *timeline.Store
	entityStore   *entities.Store
}

func NewStore(pool postgres.DB) *Store {
	return &Store{
		pool:          pool,
		authStore:     authn.NewStore(pool),
		recordStore:   records.NewStore(),
		revisionStore: revisions.NewStore(),
		timelineStore: timeline.NewStore(pool),
		entityStore:   entities.NewStore(pool),
	}
}

func (s *Store) QueryRows(ctx context.Context, incidentID uuid.UUID, viewSchemaID string, query viewschema.QueryMeta) ([]map[string]any, error) {
	switch viewSchemaID {
	case timeline.TimelineViewSchemaID:
		return s.timelineStore.QueryRows(ctx, incidentID, query)
	case entities.HostsViewSchemaID:
		return s.entityStore.QueryHostRows(ctx, incidentID, query)
	case entities.IdentitiesViewSchemaID:
		return s.entityStore.QueryIdentityRows(ctx, incidentID, query)
	case entities.IndicatorsViewSchemaID:
		return s.entityStore.QueryIndicatorRows(ctx, incidentID, query)
	default:
		definition, ok := genericSurfaces[viewSchemaID]
		if !ok {
			return nil, fmt.Errorf("workbook query surface %q not mapped", viewSchemaID)
		}
		return s.queryGenericRows(ctx, incidentID, definition, query)
	}
}

type fieldKind string

const (
	fieldKindText       fieldKind = "text"
	fieldKindTimestamp  fieldKind = "timestamp"
	fieldKindDate       fieldKind = "date"
	fieldKindBool       fieldKind = "bool"
	fieldKindNumber     fieldKind = "number"
	fieldKindCollection fieldKind = "collection"
)

type genericField struct {
	key     string
	expr    string
	kind    fieldKind
	ordered bool
}

type genericSurface struct {
	viewSchemaID string
	fromSQL      string
	recordExpr   string
	incidentExpr string
	whereSQL     string
	fields       []genericField
}

func (d genericSurface) field(key string) (genericField, bool) {
	if key == "record_id" {
		return genericField{key: "record_id", expr: d.recordExpr, kind: fieldKindText}, true
	}
	for _, field := range d.fields {
		if field.key == key {
			return field, true
		}
	}
	return genericField{}, false
}

func (s *Store) queryGenericRows(ctx context.Context, incidentID uuid.UUID, definition genericSurface, query viewschema.QueryMeta) ([]map[string]any, error) {
	sqlText, args, err := buildGenericQuerySQL(incidentID, definition, query)
	if err != nil {
		return nil, err
	}
	rows, err := s.pool.Query(ctx, sqlText, args...)
	if err != nil {
		return nil, fmt.Errorf("query workbook rows: %w", err)
	}
	defer rows.Close()

	result := make([]map[string]any, 0)
	for rows.Next() {
		values, err := rows.Values()
		if err != nil {
			return nil, fmt.Errorf("read workbook row values: %w", err)
		}
		if len(values) != len(definition.fields)+2 {
			return nil, fmt.Errorf("query workbook rows: unexpected value count %d", len(values))
		}
		row, err := buildGenericRow(definition, query.GroupBy, values)
		if err != nil {
			return nil, err
		}
		result = append(result, row)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate workbook rows: %w", err)
	}
	return result, nil
}

func buildGenericQuerySQL(incidentID uuid.UUID, definition genericSurface, query viewschema.QueryMeta) (string, []any, error) {
	var builder strings.Builder
	builder.WriteString("SELECT ")
	builder.WriteString(definition.recordExpr)
	builder.WriteString(", r.row_version")
	for _, field := range definition.fields {
		builder.WriteString(", ")
		builder.WriteString(field.expr)
	}
	builder.WriteString(" ")
	builder.WriteString(definition.fromSQL)
	builder.WriteString(" WHERE ")
	builder.WriteString(definition.incidentExpr)
	builder.WriteString(" = $1 AND r.deleted_at IS NULL")
	if definition.whereSQL != "" {
		builder.WriteString(" AND ")
		builder.WriteString(definition.whereSQL)
	}
	args := []any{incidentID}

	for _, filter := range query.Filters {
		if err := appendGenericFilter(&builder, &args, definition, filter); err != nil {
			return "", nil, err
		}
	}

	builder.WriteString(" ORDER BY ")
	for index, entry := range query.Sort {
		if index > 0 {
			builder.WriteString(", ")
		}
		field, ok := definition.field(entry.FieldKey)
		if !ok {
			return "", nil, fmt.Errorf("sort field %q not mapped for %s", entry.FieldKey, definition.viewSchemaID)
		}
		builder.WriteString(field.expr)
		if entry.Direction == "desc" {
			builder.WriteString(" DESC")
		} else {
			builder.WriteString(" ASC")
		}
	}
	return builder.String(), args, nil
}

func appendGenericFilter(builder *strings.Builder, args *[]any, definition genericSurface, filter viewschema.Filter) error {
	if filter.FieldKey == "note.full_text" {
		query, _ := filter.Arg["query"].(string)
		for _, token := range strings.Fields(query) {
			builder.WriteString("\n   AND ")
			builder.WriteString(bind(args, token))
			builder.WriteString(" = ANY(regexp_split_to_array(lower(coalesce(p.title, '') || ' ' || coalesce(p.body, '')), '[^[:alnum:]]+'))")
		}
		return nil
	}
	if filter.FieldKey == "note.tags" {
		return appendTagFilter(builder, args, definition.recordExpr, filter)
	}
	field, ok := definition.field(filter.FieldKey)
	if !ok {
		return fmt.Errorf("filter field %q not mapped for %s", filter.FieldKey, definition.viewSchemaID)
	}
	switch filter.Op {
	case "eq":
		return appendEqualityFilter(builder, args, field, filter.Arg)
	case "prefix":
		value, _ := filter.Arg["value"].(string)
		builder.WriteString("\n   AND lower(coalesce((")
		builder.WriteString(field.expr)
		builder.WriteString(")::text, '')) LIKE ")
		builder.WriteString(bind(args, value+"%"))
		return nil
	case "range":
		return appendRangeFilter(builder, args, field, filter.Arg)
	default:
		return fmt.Errorf("filter operator %q not mapped for %s", filter.Op, filter.FieldKey)
	}
}

func appendTagFilter(builder *strings.Builder, args *[]any, recordExpr string, filter viewschema.Filter) error {
	values, _ := filter.Arg["values"].([]any)
	textValues := make([]string, 0, len(values))
	for _, value := range values {
		if text, ok := value.(string); ok {
			textValues = append(textValues, text)
		}
	}
	switch filter.Op {
	case "contains_any":
		builder.WriteString("\n   AND EXISTS (SELECT 1 FROM record_tags rt WHERE rt.record_id = ")
		builder.WriteString(recordExpr)
		builder.WriteString(" AND rt.deleted_at IS NULL AND rt.normalized_tag_name = ANY(")
		builder.WriteString(bind(args, textValues))
		builder.WriteString("::text[]))")
		return nil
	case "contains_all":
		for _, value := range values {
			builder.WriteString("\n   AND EXISTS (SELECT 1 FROM record_tags rt WHERE rt.record_id = ")
			builder.WriteString(recordExpr)
			builder.WriteString(" AND rt.deleted_at IS NULL AND rt.normalized_tag_name = ")
			builder.WriteString(bind(args, value))
			builder.WriteString(")")
		}
		return nil
	default:
		return fmt.Errorf("tag filter operator %q not mapped", filter.Op)
	}
}

func appendEqualityFilter(builder *strings.Builder, args *[]any, field genericField, arg map[string]any) error {
	if value, ok := arg["value"]; ok {
		builder.WriteString("\n   AND ")
		if value == nil {
			builder.WriteString(field.expr)
			builder.WriteString(" IS NULL")
			return nil
		}
		appendComparableExpr(builder, field)
		builder.WriteString(" = ")
		builder.WriteString(bindWithFieldCast(args, value, field))
		return nil
	}
	values, _ := arg["values"].([]any)
	builder.WriteString("\n   AND (")
	for index, value := range values {
		if index > 0 {
			builder.WriteString(" OR ")
		}
		appendComparableExpr(builder, field)
		builder.WriteString(" = ")
		builder.WriteString(bindWithFieldCast(args, value, field))
	}
	builder.WriteString(")")
	return nil
}

func appendRangeFilter(builder *strings.Builder, args *[]any, field genericField, arg map[string]any) error {
	for _, bound := range []struct {
		key string
		op  string
	}{
		{key: "gt", op: ">"},
		{key: "gte", op: ">="},
		{key: "lt", op: "<"},
		{key: "lte", op: "<="},
	} {
		value, ok := arg[bound.key]
		if !ok {
			continue
		}
		builder.WriteString("\n   AND ")
		builder.WriteString(field.expr)
		builder.WriteByte(' ')
		builder.WriteString(bound.op)
		builder.WriteByte(' ')
		builder.WriteString(bindWithFieldCast(args, value, field))
	}
	return nil
}

func appendComparableExpr(builder *strings.Builder, field genericField) {
	switch field.kind {
	case fieldKindText:
		builder.WriteString("lower(coalesce((")
		builder.WriteString(field.expr)
		builder.WriteString(")::text, ''))")
	default:
		builder.WriteString(field.expr)
	}
}

func bind(args *[]any, value any) string {
	*args = append(*args, value)
	return fmt.Sprintf("$%d", len(*args))
}

func bindWithFieldCast(args *[]any, value any, field genericField) string {
	placeholder := bind(args, value)
	switch field.kind {
	case fieldKindTimestamp:
		return placeholder + "::timestamptz"
	case fieldKindDate:
		return placeholder + "::date"
	default:
		return placeholder
	}
}

func buildGenericRow(definition genericSurface, groupBy *string, values []any) (map[string]any, error) {
	recordID, err := uuidValue(values[0])
	if err != nil {
		return nil, err
	}
	cells := make(map[string]any, len(definition.fields))
	fieldValues := make(map[string]any, len(definition.fields))
	for index, field := range definition.fields {
		value := genericCellValue(field, values[index+2])
		fieldValues[field.key] = value
		cells[field.key] = map[string]any{"value": value}
	}
	row := map[string]any{
		"record_id":   recordID.String(),
		"row_version": values[1],
		"cells":       cells,
	}
	groupValues := map[string]any{}
	if groupBy != nil {
		groupValues[*groupBy] = fieldValues[*groupBy]
	}
	row["group_values"] = groupValues
	return row, nil
}

func uuidValue(value any) (uuid.UUID, error) {
	switch typed := value.(type) {
	case uuid.UUID:
		return typed, nil
	case string:
		parsed, err := uuid.Parse(typed)
		if err != nil {
			return uuid.UUID{}, fmt.Errorf("query workbook rows: invalid record_id %q", typed)
		}
		return parsed, nil
	case []byte:
		if len(typed) == 16 {
			parsed, err := uuid.FromBytes(typed)
			if err != nil {
				return uuid.UUID{}, fmt.Errorf("query workbook rows: invalid binary record_id: %w", err)
			}
			return parsed, nil
		}
		parsed, err := uuid.Parse(string(typed))
		if err != nil {
			return uuid.UUID{}, fmt.Errorf("query workbook rows: invalid record_id bytes")
		}
		return parsed, nil
	case [16]byte:
		return uuid.UUID(typed), nil
	default:
		return uuid.UUID{}, fmt.Errorf("query workbook rows: record_id was %T", value)
	}
}

func genericCellValue(field genericField, value any) any {
	if field.kind == fieldKindCollection {
		if value != nil {
			if items, ok := collectionItemsFromValue(value); ok {
				return map[string]any{
					"kind":    "collection_value_v1",
					"ordered": field.ordered,
					"items":   items,
				}
			}
		}
		return map[string]any{
			"kind":    "collection_value_v1",
			"ordered": field.ordered,
			"items":   []map[string]any{},
		}
	}
	if value == nil {
		return nil
	}
	switch typed := value.(type) {
	case time.Time:
		if field.kind == fieldKindDate {
			return typed.UTC().Format("2006-01-02")
		}
		return typed.UTC().Format(time.RFC3339Nano)
	case uuid.UUID:
		return typed.String()
	case []byte:
		if field.kind == fieldKindText && len(typed) == 16 {
			if parsed, err := uuid.FromBytes(typed); err == nil {
				return parsed.String()
			}
		}
		return string(typed)
	case [16]byte:
		if field.kind == fieldKindText {
			return uuid.UUID(typed).String()
		}
		return typed
	default:
		return typed
	}
}

func collectionItemsFromValue(value any) ([]map[string]any, bool) {
	var data []byte
	switch typed := value.(type) {
	case []byte:
		data = typed
	case string:
		data = []byte(typed)
	default:
		return nil, false
	}
	var items []map[string]any
	if err := json.Unmarshal(data, &items); err != nil {
		return nil, false
	}
	if items == nil {
		items = []map[string]any{}
	}
	return items, true
}

var genericSurfaces = map[string]genericSurface{
	AssessmentsViewSchemaID: {
		viewSchemaID: AssessmentsViewSchemaID,
		fromSQL: `FROM assessment_grid_projection a
JOIN records r ON r.record_id = a.record_id
LEFT JOIN LATERAL (
    SELECT COALESCE(jsonb_agg(jsonb_build_object(
        'item_ref', 'record_ref:' || dst.record_id::text,
        'item_kind', 'record_ref',
        'display_text', dst.record_type || ':' || dst.record_id::text,
        'linked_record_id', dst.record_id::text
    ) ORDER BY dst.record_type ASC, dst.record_id ASC), '[]'::jsonb) AS support_refs
      FROM record_links rl
      JOIN records src
        ON src.incident_id = rl.incident_id
       AND src.record_id = rl.src_record_id
       AND src.deleted_at IS NULL
      JOIN records dst
        ON dst.incident_id = rl.incident_id
       AND dst.record_id = rl.dst_record_id
       AND dst.deleted_at IS NULL
     WHERE rl.incident_id = a.incident_id
       AND rl.src_record_id = a.record_id
       AND rl.link_type = 'supported_by'
       AND rl.deleted_at IS NULL
) support ON true`,
		recordExpr:   "a.record_id",
		incidentExpr: "a.incident_id",
		fields: []genericField{
			{key: "assessment.subject_ref", expr: "a.subject_ref", kind: fieldKindText},
			{key: "assessment.subject_type", expr: "a.subject_type", kind: fieldKindText},
			{key: "assessment.assessment_state", expr: "a.assessment_state", kind: fieldKindText},
			{key: "assessment.confidence_band", expr: "a.confidence_band", kind: fieldKindText},
			{key: "assessment.confidence_score", expr: "a.confidence_score", kind: fieldKindNumber},
			{key: "assessment.rationale", expr: "a.rationale", kind: fieldKindText},
			{key: "assessment.assessor", expr: "a.assessor", kind: fieldKindText},
			{key: "assessment.assessed_at", expr: "a.assessed_at", kind: fieldKindTimestamp},
			{key: "assessment.support_refs", expr: "support.support_refs", kind: fieldKindCollection},
			{key: "assessment.supporting_link_count", expr: "a.supporting_link_count", kind: fieldKindNumber},
		},
	},
	EvidenceViewSchemaID: {
		viewSchemaID: EvidenceViewSchemaID,
		fromSQL:      "FROM evidence e JOIN records r ON r.record_id = e.record_id LEFT JOIN object_blobs b ON b.object_blob_id = e.object_blob_id",
		recordExpr:   "e.record_id",
		incidentExpr: "e.incident_id",
		fields: []genericField{
			{key: "evidence.title", expr: "e.title", kind: fieldKindText},
			{key: "evidence.lifecycle_state", expr: "e.lifecycle_state", kind: fieldKindText},
			{key: "evidence.requested_at", expr: "e.requested_at", kind: fieldKindTimestamp},
			{key: "evidence.received_at", expr: "e.received_at", kind: fieldKindTimestamp},
			{key: "evidence.storage_ref", expr: "e.storage_ref", kind: fieldKindText},
			{key: "evidence.blob_hash", expr: "COALESCE(b.observed_sha256_hex, e.blob_hash)", kind: fieldKindText},
			{key: "evidence.collector_party_text", expr: "e.collector_party_text", kind: fieldKindText},
			{key: "evidence.collector_party_id", expr: "e.collector_party_id", kind: fieldKindText},
			{key: "evidence.source_party_text", expr: "e.source_party_text", kind: fieldKindText},
			{key: "evidence.source_party_id", expr: "e.source_party_id", kind: fieldKindText},
			{key: "evidence.upload_state", expr: "COALESCE(b.upload_state, e.upload_state)", kind: fieldKindText},
			{key: "evidence.linked_record_count", expr: "0", kind: fieldKindNumber},
			{key: "evidence.edited_at", expr: "e.updated_at", kind: fieldKindTimestamp},
		},
	},
	DecisionsViewSchemaID: {
		viewSchemaID: DecisionsViewSchemaID,
		fromSQL:      "FROM decisions d JOIN records r ON r.record_id = d.record_id",
		recordExpr:   "d.record_id",
		incidentExpr: "d.incident_id",
		fields: []genericField{
			{key: "decision.summary", expr: "d.summary", kind: fieldKindText},
			{key: "decision.status", expr: "d.status", kind: fieldKindText},
			{key: "decision.owner_user_id", expr: "d.owner_user_id", kind: fieldKindText},
			{key: "decision.decision_type", expr: "d.decision_type", kind: fieldKindText},
			{key: "decision.decided_at", expr: "d.decided_at", kind: fieldKindTimestamp},
			{key: "decision.rationale", expr: "d.rationale", kind: fieldKindText},
			{key: "decision.support_refs", expr: recordRefCollectionExprFor("d", "decision.support_refs", "supported_by"), kind: fieldKindCollection},
			{key: "decision.affected_record_count", expr: linkCountExprFor("d", "decision.affected_record_ids", "references_record"), kind: fieldKindNumber},
			{key: "decision.supersedes_record_id", expr: supersedesRecordIDExprFor("d"), kind: fieldKindText},
			{key: "decision.updated_at", expr: "d.updated_at", kind: fieldKindTimestamp},
			{key: "decision.is_superseded", expr: isSupersededExprFor("d"), kind: fieldKindBool},
		},
	},
	PartiesViewSchemaID: {
		viewSchemaID: PartiesViewSchemaID,
		fromSQL:      "FROM parties p JOIN records r ON r.record_id = p.record_id",
		recordExpr:   "p.record_id",
		incidentExpr: "p.incident_id",
		fields: []genericField{
			{key: "party.display_name", expr: "p.display_name", kind: fieldKindText},
			{key: "party.party_kind", expr: "p.party_kind", kind: fieldKindText},
			{key: "party.organization_name", expr: "p.organization_name", kind: fieldKindText},
			{key: "party.role_title", expr: "p.role_title", kind: fieldKindText},
			{key: "party.primary_email", expr: "p.primary_email", kind: fieldKindText},
			{key: "party.timezone_name", expr: "p.timezone_name", kind: fieldKindText},
			{key: "party.external_ref", expr: "p.external_ref", kind: fieldKindText},
			{key: "party.notes", expr: "p.notes", kind: fieldKindText},
			{key: "party.updated_at", expr: "p.updated_at", kind: fieldKindTimestamp},
		},
	},
	TaskRequestsViewSchemaID: {
		viewSchemaID: TaskRequestsViewSchemaID,
		fromSQL:      "FROM task_requests t JOIN records r ON r.record_id = t.record_id",
		recordExpr:   "t.record_id",
		incidentExpr: "t.incident_id",
		fields: []genericField{
			{key: "task.title", expr: "t.title", kind: fieldKindText},
			{key: "task.status", expr: "t.status", kind: fieldKindText},
			{key: "task.owner_user_id", expr: "t.owner_user_id", kind: fieldKindText},
			{key: "task.priority", expr: "t.priority", kind: fieldKindText},
			{key: "task.task_kind", expr: "t.task_kind", kind: fieldKindText},
			{key: "task.workstream", expr: "t.workstream", kind: fieldKindText},
			{key: "task.due_at", expr: "t.due_at", kind: fieldKindTimestamp},
			{key: "task.requester_party_text", expr: "t.requester_party_text", kind: fieldKindText},
			{key: "task.requester_party_id", expr: "t.requester_party_id", kind: fieldKindText},
			{key: "task.blocked_reason", expr: "t.blocked_reason", kind: fieldKindText},
			{key: "task.completed_at", expr: "t.completed_at", kind: fieldKindTimestamp},
			{key: "task.external_ticket_ref", expr: "t.external_ticket_ref", kind: fieldKindText},
			{key: "task.closure_summary", expr: "t.closure_summary", kind: fieldKindText},
			{key: "task.linked_record_ids", expr: recordRefCollectionExprFor("t", "task.linked_record_ids", "references_record"), kind: fieldKindCollection},
			{key: "task.decision_record_id", expr: "t.decision_record_id", kind: fieldKindText},
			{key: "task.linked_record_count", expr: linkCountExprFor("t", "task.linked_record_ids", "references_record"), kind: fieldKindNumber},
			{key: "task.updated_at", expr: "t.updated_at", kind: fieldKindTimestamp},
			{key: "task.no_owner", expr: "t.owner_user_id IS NULL", kind: fieldKindBool},
		},
	},
	NotesViewSchemaID: artifactSurface(NotesViewSchemaID, "note", []genericField{
		{key: "note.title", expr: "p.title", kind: fieldKindText},
		{key: "note.body", expr: "p.body", kind: fieldKindText},
		{key: "note.tags", expr: tagCollectionExprFor("p"), kind: fieldKindCollection},
		{key: "note.linked_record_count", expr: "p.linked_record_count", kind: fieldKindNumber},
		{key: "note.updated_at", expr: "p.updated_at", kind: fieldKindTimestamp},
		{key: "note.created_by_user_id", expr: "p.created_by_user_id", kind: fieldKindText},
	}),
	CommLogViewSchemaID: artifactSurface(CommLogViewSchemaID, "comm_log", []genericField{
		{key: "comm_log.timestamp_utc", expr: "p.timestamp_utc", kind: fieldKindTimestamp},
		{key: "comm_log.comm_type", expr: "p.comm_type", kind: fieldKindText},
		{key: "comm_log.audience", expr: "p.audience", kind: fieldKindText},
		{key: "comm_log.channel_or_meeting", expr: "p.channel_or_meeting", kind: fieldKindText},
		{key: "comm_log.summary", expr: "p.summary", kind: fieldKindText},
		{key: "comm_log.next_report_at", expr: "p.next_report_at", kind: fieldKindTimestamp},
		{key: "comm_log.privilege_tag", expr: "p.privilege_tag", kind: fieldKindText},
		{key: "comm_log.decision_ids", expr: recordRefCollectionExpr("comm_log.decision_ids"), kind: fieldKindCollection},
		{key: "comm_log.action_task_ids", expr: recordRefCollectionExpr("comm_log.action_task_ids"), kind: fieldKindCollection},
		{key: "comm_log.audience_party_ids", expr: partyRefCollectionExpr("comm_log.audience_party_ids"), kind: fieldKindCollection},
		{key: "comm_log.attendee_party_ids", expr: partyRefCollectionExpr("comm_log.attendee_party_ids"), kind: fieldKindCollection},
		{key: "comm_log.comm_id", expr: "p.comm_id", kind: fieldKindText},
		{key: "comm_log.timestamp_day", expr: "p.timestamp_day", kind: fieldKindDate},
		{key: "comm_log.next_report_day", expr: "p.next_report_day", kind: fieldKindDate},
		{key: "comm_log.updated_at", expr: "p.updated_at", kind: fieldKindTimestamp},
	}),
	HandoffViewSchemaID: artifactSurface(HandoffViewSchemaID, "handoff", []genericField{
		{key: "handoff.timestamp_utc", expr: "p.timestamp_utc", kind: fieldKindTimestamp},
		{key: "handoff.outgoing_owner_user_id", expr: "p.outgoing_owner_user_id", kind: fieldKindText},
		{key: "handoff.incoming_owner_user_id", expr: "p.incoming_owner_user_id", kind: fieldKindText},
		{key: "handoff.current_state_summary", expr: "p.current_state_summary", kind: fieldKindText},
		{key: "handoff.open_task_ids", expr: recordRefCollectionExpr("handoff.open_task_ids"), kind: fieldKindCollection},
		{key: "handoff.open_decision_ids", expr: recordRefCollectionExpr("handoff.open_decision_ids"), kind: fieldKindCollection},
		{key: "handoff.open_risk_refs", expr: riskRefCollectionExpr(), kind: fieldKindCollection},
		{key: "handoff.next_checks", expr: "p.next_checks", kind: fieldKindText},
		{key: "handoff.acknowledged_at", expr: "p.acknowledged_at", kind: fieldKindTimestamp},
		{key: "handoff.handoff_id", expr: "p.handoff_id", kind: fieldKindText},
		{key: "handoff.timestamp_day", expr: "p.timestamp_day", kind: fieldKindDate},
		{key: "handoff.ack_state", expr: "p.ack_state", kind: fieldKindText},
		{key: "handoff.updated_at", expr: "p.updated_at", kind: fieldKindTimestamp},
	}),
	StatusReviewViewSchemaID: artifactSurface(StatusReviewViewSchemaID, "status_review", []genericField{
		{key: "status_review.timestamp_utc", expr: "p.timestamp_utc", kind: fieldKindTimestamp},
		{key: "status_review.review_owner_user_id", expr: "p.review_owner_user_id", kind: fieldKindText},
		{key: "status_review.current_state_summary", expr: "p.current_state_summary", kind: fieldKindText},
		{key: "status_review.blocked_task_ids", expr: recordRefCollectionExpr("status_review.blocked_task_ids"), kind: fieldKindCollection},
		{key: "status_review.pending_evidence_ids", expr: recordRefCollectionExpr("status_review.pending_evidence_ids"), kind: fieldKindCollection},
		{key: "status_review.open_decision_ids", expr: recordRefCollectionExpr("status_review.open_decision_ids"), kind: fieldKindCollection},
		{key: "status_review.active_risks_summary", expr: "p.active_risks_summary", kind: fieldKindText},
		{key: "status_review.next_report_at", expr: "p.next_report_at", kind: fieldKindTimestamp},
		{key: "status_review.status_review_id", expr: "p.status_review_id", kind: fieldKindText},
		{key: "status_review.timestamp_day", expr: "p.timestamp_day", kind: fieldKindDate},
		{key: "status_review.next_report_day", expr: "p.next_report_day", kind: fieldKindDate},
		{key: "status_review.updated_at", expr: "p.updated_at", kind: fieldKindTimestamp},
	}),
	LessonViewSchemaID: artifactSurface(LessonViewSchemaID, "lesson", []genericField{
		{key: "lesson.timestamp_utc", expr: "p.timestamp_utc", kind: fieldKindTimestamp},
		{key: "lesson.summary", expr: "p.summary", kind: fieldKindText},
		{key: "lesson.owner_user_id", expr: "p.owner_user_id", kind: fieldKindText},
		{key: "lesson.closure_state", expr: "p.closure_state", kind: fieldKindText},
		{key: "lesson.follow_up_task_ids", expr: recordRefCollectionExpr("lesson.follow_up_task_ids"), kind: fieldKindCollection},
		{key: "lesson.evidence_refs", expr: recordRefCollectionExpr("lesson.evidence_refs"), kind: fieldKindCollection},
		{key: "lesson.lesson_id", expr: "p.lesson_id", kind: fieldKindText},
		{key: "lesson.timestamp_day", expr: "p.timestamp_day", kind: fieldKindDate},
		{key: "lesson.updated_at", expr: "p.updated_at", kind: fieldKindTimestamp},
	}),
}

func recordRefCollectionExpr(fieldKey string) string {
	return recordRefCollectionExprFor("p", fieldKey, "references_record")
}

func recordRefCollectionExprFor(alias string, fieldKey string, linkType string) string {
	return `(SELECT COALESCE(jsonb_agg(jsonb_build_object(
        'item_ref', 'record_ref:' || dst.record_id::text,
        'item_kind', 'record_ref',
        'display_text', dst.record_type || ':' || dst.record_id::text,
        'linked_record_id', dst.record_id::text
    ) ORDER BY dst.record_type ASC, dst.record_id ASC), '[]'::jsonb)
      FROM record_links rl
      JOIN records dst
        ON dst.incident_id = rl.incident_id
       AND dst.record_id = rl.dst_record_id
       AND dst.deleted_at IS NULL
     WHERE rl.incident_id = ` + alias + `.incident_id
       AND rl.src_record_id = ` + alias + `.record_id
       AND rl.link_type = '` + linkType + `'
       AND rl.field_key = '` + fieldKey + `'
       AND rl.deleted_at IS NULL)::text`
}

func tagCollectionExprFor(alias string) string {
	return `(SELECT COALESCE(jsonb_agg(jsonb_build_object(
        'item_ref', 'record_tag:' || rt.record_id::text || ':' || rt.record_tag_id::text,
        'item_kind', 'tag',
        'display_text', rt.tag_name,
        'tag_id', rt.record_tag_id::text
    ) ORDER BY rt.normalized_tag_name ASC, rt.record_tag_id ASC), '[]'::jsonb)
      FROM record_tags rt
     WHERE rt.incident_id = ` + alias + `.incident_id
       AND rt.record_id = ` + alias + `.record_id
       AND rt.deleted_at IS NULL)::text`
}

func linkCountExprFor(alias string, fieldKey string, linkType string) string {
	return `(SELECT COUNT(*)::integer
      FROM record_links rl
      JOIN records dst
        ON dst.incident_id = rl.incident_id
       AND dst.record_id = rl.dst_record_id
       AND dst.deleted_at IS NULL
     WHERE rl.incident_id = ` + alias + `.incident_id
       AND rl.src_record_id = ` + alias + `.record_id
       AND rl.link_type = '` + linkType + `'
       AND rl.field_key = '` + fieldKey + `'
       AND rl.deleted_at IS NULL)`
}

func supersedesRecordIDExprFor(alias string) string {
	return `(SELECT rl.dst_record_id
      FROM record_links rl
      JOIN records dst
        ON dst.incident_id = rl.incident_id
       AND dst.record_id = rl.dst_record_id
       AND dst.record_type = 'decision'
       AND dst.deleted_at IS NULL
     WHERE rl.incident_id = ` + alias + `.incident_id
       AND rl.src_record_id = ` + alias + `.record_id
       AND rl.link_type = 'supersedes'
       AND rl.deleted_at IS NULL
     ORDER BY rl.created_at DESC, rl.record_link_id DESC
     LIMIT 1)`
}

func isSupersededExprFor(alias string) string {
	return `EXISTS (
      SELECT 1
        FROM record_links rl
        JOIN records src
          ON src.incident_id = rl.incident_id
         AND src.record_id = rl.src_record_id
         AND src.record_type = 'decision'
         AND src.deleted_at IS NULL
       WHERE rl.incident_id = ` + alias + `.incident_id
         AND rl.dst_record_id = ` + alias + `.record_id
         AND rl.link_type = 'supersedes'
         AND rl.deleted_at IS NULL
    )`
}

func partyRefCollectionExpr(fieldKey string) string {
	return `(SELECT COALESCE(jsonb_agg(jsonb_build_object(
        'item_ref', 'party_ref:' || party.record_id::text,
        'item_kind', 'party_ref',
        'display_text', party.display_name,
        'party_id', party.record_id::text
    ) ORDER BY party.display_name ASC, party.record_id ASC), '[]'::jsonb)
      FROM record_links rl
      JOIN parties party
        ON party.incident_id = rl.incident_id
       AND party.record_id = rl.dst_record_id
      JOIN records dst
        ON dst.incident_id = rl.incident_id
       AND dst.record_id = rl.dst_record_id
       AND dst.deleted_at IS NULL
     WHERE rl.incident_id = p.incident_id
       AND rl.src_record_id = p.record_id
       AND rl.link_type = 'references_record'
       AND rl.field_key = '` + fieldKey + `'
       AND rl.deleted_at IS NULL)::text`
}

func riskRefCollectionExpr() string {
	return `(SELECT COALESCE(jsonb_agg(jsonb_build_object(
        'item_ref', 'risk_ref:' || risk_ref_id::text,
        'item_kind', 'risk_ref',
        'display_text', risk_ref_text,
        'risk_ref_id', risk_ref_id::text,
        'risk_ref_text', risk_ref_text
    ) ORDER BY risk_ref_text ASC, risk_ref_id ASC), '[]'::jsonb)
      FROM handoff_risk_refs hr
     WHERE hr.incident_id = p.incident_id
       AND hr.handoff_record_id = p.record_id
       AND hr.deleted_at IS NULL)::text`
}

func artifactSurface(viewSchemaID string, artifactType string, fields []genericField) genericSurface {
	return genericSurface{
		viewSchemaID: viewSchemaID,
		fromSQL:      "FROM artifact_grid_projection p JOIN records r ON r.record_id = p.record_id",
		recordExpr:   "p.record_id",
		incidentExpr: "p.incident_id",
		whereSQL:     "p.artifact_type = '" + artifactType + "'",
		fields:       fields,
	}
}
