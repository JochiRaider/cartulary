-- +goose Up
--
-- Links-owned projection read contracts. These views expose active link and tag
-- semantics without making projection providers depend on source table details.
--

CREATE VIEW public.active_record_links_v1 AS
SELECT
    rl.record_link_id,
    rl.incident_id,
    rl.src_record_id,
    src.record_type AS src_record_type,
    rl.dst_record_id,
    dst.record_type AS dst_record_type,
    rl.link_type,
    rl.provenance,
    rl.confidence,
    rl.owner_user_id,
    rl.decided_at,
    rl.created_at,
    rl.deleted_at,
    rl.deleted_by_user_id,
    rl.created_by_user_id,
    rl.field_key
  FROM public.record_links rl
  JOIN public.records src
    ON src.incident_id = rl.incident_id
   AND src.record_id = rl.src_record_id
   AND src.deleted_at IS NULL
  JOIN public.records dst
    ON dst.incident_id = rl.incident_id
   AND dst.record_id = rl.dst_record_id
   AND dst.deleted_at IS NULL
 WHERE rl.deleted_at IS NULL;

CREATE VIEW public.active_record_tags_v1 AS
SELECT
    rt.record_tag_id,
    rt.incident_id,
    rt.record_id,
    r.record_type,
    rt.tag_name,
    rt.normalized_tag_name,
    rt.created_by_user_id,
    rt.created_at,
    rt.updated_at,
    rt.deleted_at,
    rt.deleted_by_user_id
  FROM public.record_tags rt
  JOIN public.records r
    ON r.incident_id = rt.incident_id
   AND r.record_id = rt.record_id
   AND r.deleted_at IS NULL
 WHERE rt.deleted_at IS NULL;

-- +goose Down
DROP VIEW IF EXISTS public.active_record_tags_v1;
DROP VIEW IF EXISTS public.active_record_links_v1;
