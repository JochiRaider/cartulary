package workbookprojection

import (
	"github.com/JochiRaider/cartulary/internal/modules/links/readshape"
	"github.com/JochiRaider/cartulary/internal/modules/projections/providercontract"
)

const timelineViewSchemaID = "cartulary.view.timeline.v2"

func QuerySurfaces() []providercontract.QuerySurface {
	return []providercontract.QuerySurface{{
		ViewSchemaID: timelineViewSchemaID,
		FromSQL: `FROM timeline_grid_projection t
JOIN records r ON r.record_id = t.record_id
LEFT JOIN LATERAL (
    SELECT COALESCE(jsonb_agg(
      jsonb_build_object(
        'item_ref', 'entity_mention:' || em.entity_mention_id::text,
        'item_kind', CASE WHEN em.resolution_status = 'resolved' THEN 'resolved_ref' ELSE 'unresolved_mention' END,
        'entity_type', em.entity_type,
        'display_text', em.raw_text,
        'raw_text', em.raw_text,
        'mention_row_version', em.row_version
      )
      || CASE WHEN em.resolution_status = 'resolved' THEN jsonb_strip_nulls(jsonb_build_object(
        'resolved_record_id', em.resolved_record_id::text,
        'resolution_method', em.resolution_method,
        'auto_resolved', CASE WHEN em.resolution_method = 'auto_match' THEN true ELSE NULL END,
        'matched_alias_text', CASE WHEN em.resolution_method = 'auto_match' THEN matched_alias.raw_text ELSE NULL END
      )) ELSE '{}'::jsonb END
      || CASE WHEN relationship.provenance IS NOT NULL THEN jsonb_build_object(
        'provenance', relationship.provenance,
        'confidence', relationship.confidence
      ) ELSE '{}'::jsonb END
      ORDER BY em.ordinal ASC, em.entity_mention_id ASC), '[]'::jsonb) AS host_refs
      FROM entity_mentions em
      LEFT JOIN LATERAL (
          SELECT l.provenance, l.confidence
            FROM ` + readshape.ActiveRecordLinksAlias("l") + `
           WHERE l.incident_id = t.incident_id
             AND l.src_record_id = t.record_id
             AND l.dst_record_id = em.resolved_record_id
             AND l.link_type = 'observed_on_host'
           ORDER BY l.created_at DESC, l.record_link_id DESC
           LIMIT 1
      ) relationship ON true
      LEFT JOIN LATERAL (
          SELECT ea.raw_text
            FROM entity_aliases ea
           WHERE ea.record_id = em.resolved_record_id
             AND ea.entity_type = em.entity_type
             AND regexp_replace(btrim(ea.raw_text), '[[:space:]]+', ' ', 'g')::citext =
                 em.normalized_text::citext
             AND ea.deleted_at IS NULL
           ORDER BY ea.created_at ASC, ea.entity_alias_id ASC
           LIMIT 1
      ) matched_alias ON true
     WHERE em.source_record_id = t.record_id
       AND em.entity_type = 'host'
       AND em.resolution_status IN ('unresolved', 'resolved')
) host_mentions ON true
LEFT JOIN LATERAL (
    SELECT COALESCE(jsonb_agg(
      jsonb_build_object(
        'item_ref', 'entity_mention:' || em.entity_mention_id::text,
        'item_kind', CASE WHEN em.resolution_status = 'resolved' THEN 'resolved_ref' ELSE 'unresolved_mention' END,
        'entity_type', em.entity_type,
        'display_text', em.raw_text,
        'raw_text', em.raw_text,
        'mention_row_version', em.row_version
      )
      || CASE WHEN em.resolution_status = 'resolved' THEN jsonb_strip_nulls(jsonb_build_object(
        'resolved_record_id', em.resolved_record_id::text,
        'resolution_method', em.resolution_method,
        'auto_resolved', CASE WHEN em.resolution_method = 'auto_match' THEN true ELSE NULL END,
        'matched_alias_text', CASE WHEN em.resolution_method = 'auto_match' THEN matched_alias.raw_text ELSE NULL END
      )) ELSE '{}'::jsonb END
      || CASE WHEN relationship.provenance IS NOT NULL THEN jsonb_build_object(
        'provenance', relationship.provenance,
        'confidence', relationship.confidence
      ) ELSE '{}'::jsonb END
      ORDER BY em.ordinal ASC, em.entity_mention_id ASC), '[]'::jsonb) AS identity_refs
      FROM entity_mentions em
      LEFT JOIN LATERAL (
          SELECT l.provenance, l.confidence
            FROM ` + readshape.ActiveRecordLinksAlias("l") + `
           WHERE l.incident_id = t.incident_id
             AND l.src_record_id = t.record_id
             AND l.dst_record_id = em.resolved_record_id
             AND l.link_type = 'observed_as_identity'
           ORDER BY l.created_at DESC, l.record_link_id DESC
           LIMIT 1
      ) relationship ON true
      LEFT JOIN LATERAL (
          SELECT ea.raw_text
            FROM entity_aliases ea
           WHERE ea.record_id = em.resolved_record_id
             AND ea.entity_type = em.entity_type
             AND regexp_replace(btrim(ea.raw_text), '[[:space:]]+', ' ', 'g')::citext =
                 em.normalized_text::citext
             AND ea.deleted_at IS NULL
           ORDER BY ea.created_at ASC, ea.entity_alias_id ASC
           LIMIT 1
      ) matched_alias ON true
     WHERE em.source_record_id = t.record_id
       AND em.entity_type = 'identity'
       AND em.resolution_status IN ('unresolved', 'resolved')
) identity_mentions ON true
LEFT JOIN LATERAL (
    SELECT COALESCE(jsonb_agg(jsonb_build_object(
        'item_ref', ` + readshape.RecordTagItemRefSQL("t.record_id", "rt.record_tag_id") + `,
        'item_kind', 'tag',
        'display_text', rt.tag_name,
        'tag_id', rt.record_tag_id::text
    ) ORDER BY rt.normalized_tag_name ASC, rt.record_tag_id ASC), '[]'::jsonb) AS tags
      FROM ` + readshape.ActiveRecordTagsAlias("rt") + `
     WHERE rt.incident_id = t.incident_id
       AND rt.record_id = t.record_id
) tags ON true
LEFT JOIN LATERAL (
    SELECT COALESCE(jsonb_agg(jsonb_build_object(
        'item_ref', ` + readshape.RecordRefItemRefSQL("rl.dst_record_id") + `,
        'item_kind', 'record_ref',
        'display_text', COALESCE(ev.title, rl.dst_record_id::text),
        'linked_record_id', rl.dst_record_id::text
    ) ORDER BY COALESCE(ev.title, rl.dst_record_id::text) ASC, rl.dst_record_id ASC), '[]'::jsonb) AS attached_evidence
      FROM ` + readshape.ActiveRecordLinksAlias("rl") + `
      JOIN evidence ev
        ON ev.incident_id = rl.incident_id
       AND ev.record_id = rl.dst_record_id
     WHERE rl.incident_id = t.incident_id
       AND rl.src_record_id = t.record_id
       AND rl.link_type = 'attached_evidence'
) attached_evidence ON true`,
		RecordExpr:   "t.record_id",
		IncidentExpr: "t.incident_id",
		Fields: []providercontract.QueryField{
			{Key: "timeline.date_entered_text", Expr: "t.date_entered_text", Kind: providercontract.FieldKindText},
			{Key: "timeline.analyst_text", Expr: "t.analyst_text", Kind: providercontract.FieldKindText},
			{Key: "timeline.mitre_stage_text", Expr: "t.mitre_stage_text", Kind: providercontract.FieldKindText},
			{Key: "timeline.device_object_text", Expr: "t.device_object_text", Kind: providercontract.FieldKindText},
			{Key: "timeline.ip_address_text", Expr: "t.ip_address_text", Kind: providercontract.FieldKindText},
			{Key: "timeline.activity_utc_text", Expr: "t.activity_utc_text", Kind: providercontract.FieldKindText},
			{Key: "timeline.activity_local_text", Expr: "t.activity_local_text", Kind: providercontract.FieldKindText},
			{Key: "timeline.raw_activity_text", Expr: "t.raw_activity_text", Kind: providercontract.FieldKindText},
			{Key: "timeline.activity_synopsis_text", Expr: "t.activity_synopsis_text", Kind: providercontract.FieldKindText},
			{Key: "timeline.data_source_text", Expr: "t.data_source_text", Kind: providercontract.FieldKindText},
			{Key: "timeline.host_refs", Expr: "host_mentions.host_refs", Kind: providercontract.FieldKindCollection, Ordered: true},
			{Key: "timeline.identity_refs", Expr: "identity_mentions.identity_refs", Kind: providercontract.FieldKindCollection, Ordered: true},
			{Key: "timeline.tags", Expr: "tags.tags", Kind: providercontract.FieldKindCollection},
			{Key: "timeline.attached_evidence_ids", Expr: "attached_evidence.attached_evidence", Kind: providercontract.FieldKindCollection},
			{Key: "timeline.evidence_count", Expr: "t.evidence_count", Kind: providercontract.FieldKindNumber},
			{Key: "timeline.recorded_at", Expr: "t.recorded_at", Kind: providercontract.FieldKindTimestamp},
			{Key: "timeline.edited_at", Expr: "t.edited_at", Kind: providercontract.FieldKindTimestamp},
			{Key: "timeline.activity_sort_ts", Expr: "t.activity_sort_ts", Kind: providercontract.FieldKindTimestamp},
			{Key: "timeline.date_entered_sort_day", Expr: "t.date_entered_sort_day", Kind: providercontract.FieldKindDate},
			{Key: "timeline.activity_time_pair_state", Expr: "t.activity_time_pair_state", Kind: providercontract.FieldKindText},
			{Key: "timeline.capture_state", Expr: "t.capture_state", Kind: providercontract.FieldKindText},
			{Key: "timeline.replacement_record_id", Expr: "t.replacement_record_id", Kind: providercontract.FieldKindText},
			{Key: "timeline.has_evidence", Expr: "t.has_evidence", Kind: providercontract.FieldKindBool},
			{Key: "timeline.has_unresolved_mentions", Expr: "t.has_unresolved_mentions", Kind: providercontract.FieldKindBool},
		},
	}}
}
