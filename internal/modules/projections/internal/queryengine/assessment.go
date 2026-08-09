package queryengine

import (
	"strconv"
	"strings"
)

const assessmentsViewSchemaID = "cartulary.view.assessments.v1"

func AssessmentPlans() []Surface {
	return []Surface{{
		ViewSchemaID: assessmentsViewSchemaID,
		FromSQL: `FROM assessment_grid_projection a
JOIN records r ON r.record_id = a.record_id
LEFT JOIN LATERAL (
    SELECT COALESCE(jsonb_agg(jsonb_build_object(
        'item_ref', 'record_ref:' || dst.record_id::text,
        'item_kind', 'record_ref',
        'display_text', dst.record_type || ':' || dst.record_id::text,
        'linked_record_id', dst.record_id::text
    ) ORDER BY dst.record_type ASC, dst.record_id ASC), '[]'::jsonb) AS support_refs
      FROM active_record_links_v1 rl
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
) support ON true`,
		RecordExpr:   "a.record_id",
		IncidentExpr: "a.incident_id",
		Fields: []Field{
			{Key: "assessment.subject_ref", Expr: "a.subject_ref", Kind: FieldKindText},
			{Key: "assessment.subject_type", Expr: "a.subject_type", Kind: FieldKindText},
			{Key: "assessment.assessment_state", Expr: "a.assessment_state", Kind: FieldKindText},
			{Key: "assessment.confidence_band", Expr: "a.confidence_band", SortExpr: assessmentEnumSortExpr("a.confidence_band", "unset", "low", "medium", "high"), Kind: FieldKindText},
			{Key: "assessment.confidence_score", Expr: "a.confidence_score", Kind: FieldKindNumber},
			{Key: "assessment.rationale", Expr: "a.rationale", Kind: FieldKindText},
			{Key: "assessment.assessor", Expr: "a.assessor", Kind: FieldKindText},
			{Key: "assessment.assessed_at", Expr: "a.assessed_at", Kind: FieldKindTimestamp},
			{Key: "assessment.support_refs", Expr: "support.support_refs", Kind: FieldKindCollection},
			{Key: "assessment.supporting_link_count", Expr: "a.supporting_link_count", Kind: FieldKindNumber},
		},
	}}
}

func assessmentEnumSortExpr(expr string, values ...string) string {
	var builder strings.Builder
	builder.WriteString("CASE ")
	builder.WriteString(expr)
	for index, value := range values {
		builder.WriteString(" WHEN '")
		builder.WriteString(value)
		builder.WriteString("' THEN ")
		builder.WriteString(strconv.Itoa(index))
	}
	builder.WriteString(" ELSE ")
	builder.WriteString(strconv.Itoa(len(values)))
	builder.WriteString(" END")
	return builder.String()
}
