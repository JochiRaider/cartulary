package projections

func assessmentQuerySurfaces() []genericSurface {
	return []genericSurface{{
		viewSchemaID: assessmentsViewSchemaID,
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
			{key: "assessment.confidence_band", expr: "a.confidence_band", sortExpr: enumSortExpr("a.confidence_band", "unset", "low", "medium", "high"), kind: fieldKindText},
			{key: "assessment.confidence_score", expr: "a.confidence_score", kind: fieldKindNumber},
			{key: "assessment.rationale", expr: "a.rationale", kind: fieldKindText},
			{key: "assessment.assessor", expr: "a.assessor", kind: fieldKindText},
			{key: "assessment.assessed_at", expr: "a.assessed_at", kind: fieldKindTimestamp},
			{key: "assessment.support_refs", expr: "support.support_refs", kind: fieldKindCollection},
			{key: "assessment.supporting_link_count", expr: "a.supporting_link_count", kind: fieldKindNumber},
		},
	}}
}
