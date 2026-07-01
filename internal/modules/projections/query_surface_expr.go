package projections

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
