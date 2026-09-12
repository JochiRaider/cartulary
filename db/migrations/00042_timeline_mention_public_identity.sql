-- +goose Up

-- Enrich derived collection JSON only. The existing server-owned binding is
-- compared against source identity; no client selector is parsed as an ID.
-- Stop old projection writers before applying this forward upgrade.
LOCK TABLE public.timeline_grid_projection IN SHARE ROW EXCLUSIVE MODE;
LOCK TABLE public.entity_mentions IN SHARE MODE;

-- +goose StatementBegin
DO $$
BEGIN
    IF EXISTS (
        SELECT 1
          FROM public.timeline_grid_projection AS projection
          CROSS JOIN LATERAL (VALUES
              ('timeline.host_refs', 'host', projection.host_refs),
              ('timeline.identity_refs', 'identity', projection.identity_refs)
          ) AS collection(field_key, entity_type, items)
          CROSS JOIN LATERAL jsonb_array_elements(collection.items) AS item(value)
          LEFT JOIN public.entity_mentions AS mention
            ON mention.source_record_id = projection.record_id
           AND mention.source_field_key = collection.field_key
           AND mention.entity_type = collection.entity_type
           AND 'entity_mention:' || mention.entity_mention_id::text = item.value->>'item_ref'
         WHERE mention.entity_mention_id IS NULL
            OR mention.resolution_status NOT IN ('unresolved', 'resolved')
            OR item.value->>'entity_type' IS DISTINCT FROM mention.entity_type
            OR item.value->'mention_row_version' IS DISTINCT FROM to_jsonb(mention.row_version)
            OR (item.value ? 'entity_mention_id' AND
                item.value->>'entity_mention_id' IS DISTINCT FROM mention.entity_mention_id::text)
    ) THEN
        RAISE EXCEPTION USING ERRCODE = '23514',
            MESSAGE = 'timeline_mention_projection_binding_invalid';
    END IF;
END;
$$;
-- +goose StatementEnd

UPDATE public.timeline_grid_projection AS projection
   SET host_refs = (
       SELECT COALESCE(jsonb_agg(item.value || jsonb_build_object(
           'entity_mention_id', mention.entity_mention_id::text
       ) ORDER BY item.ordinal), '[]'::jsonb)
         FROM jsonb_array_elements(projection.host_refs) WITH ORDINALITY AS item(value, ordinal)
         JOIN public.entity_mentions AS mention
           ON mention.source_record_id = projection.record_id
          AND mention.source_field_key = 'timeline.host_refs'
          AND 'entity_mention:' || mention.entity_mention_id::text = item.value->>'item_ref'
   ), identity_refs = (
       SELECT COALESCE(jsonb_agg(item.value || jsonb_build_object(
           'entity_mention_id', mention.entity_mention_id::text
       ) ORDER BY item.ordinal), '[]'::jsonb)
         FROM jsonb_array_elements(projection.identity_refs) WITH ORDINALITY AS item(value, ordinal)
         JOIN public.entity_mentions AS mention
           ON mention.source_record_id = projection.record_id
          AND mention.source_field_key = 'timeline.identity_refs'
          AND 'entity_mention:' || mention.entity_mention_id::text = item.value->>'item_ref'
   );

-- +goose Down

-- Additive derived metadata is readable by old clients. Keep it on rollback;
-- rolling application code back never reverses committed source operations.
SELECT 1;
