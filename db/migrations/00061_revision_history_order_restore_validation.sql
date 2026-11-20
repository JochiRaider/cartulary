-- +goose Up
-- Recovery loads change sets with triggers disabled, so the commit-time
-- allocator never runs for restored rows. This Revisions-owned predicate lets
-- Recovery prove the restored chronology before commit: every change set has a
-- history order, and each incident's allocator covers its greatest order. It
-- reads only; no table, constraint, trigger or data changes.
-- +goose StatementBegin
CREATE FUNCTION public.revisions_history_order_is_valid_v1()
RETURNS boolean LANGUAGE sql STABLE SET search_path = pg_catalog, public AS $$
 SELECT NOT EXISTS (SELECT 1 FROM public.change_sets WHERE history_order IS NULL)
 AND NOT EXISTS (
   SELECT 1
     FROM (SELECT incident_id, max(history_order) AS high_water
             FROM public.change_sets
            GROUP BY incident_id) AS restored
     LEFT JOIN public.revision_history_orders AS allocator USING (incident_id)
    WHERE allocator.last_order IS NULL OR allocator.last_order < restored.high_water
 );
$$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION public.revisions_history_order_is_valid_v1() FROM PUBLIC;
GRANT EXECUTE ON FUNCTION public.revisions_history_order_is_valid_v1() TO cartulary_recovery;

-- +goose Down
-- Remove the predicate exactly; no data is transformed.
DROP FUNCTION public.revisions_history_order_is_valid_v1();
