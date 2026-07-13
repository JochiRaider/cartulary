-- +goose Up
-- +goose StatementBegin
DO $$
DECLARE
    referenced_run_count bigint;
BEGIN
    SELECT count(*)
      INTO referenced_run_count
      FROM public.reporting_releases AS release,
           LATERAL jsonb_array_elements(release.graph_projection_refs) AS ref
     WHERE EXISTS (
           SELECT 1
             FROM public.graph_projection_runs AS run
            WHERE run.projection_run_id = ref ->> 'projection_run_id'
     );

    IF referenced_run_count > 0 THEN
        RAISE EXCEPTION USING
            ERRCODE = 'check_violation',
            MESSAGE = format(
                'graph projection conformance migration blocked: referenced_projection_run_count=%s',
                referenced_run_count
            ),
            HINT = 'Export or rebuild the affected Reporting release tuples under the adopted Graph Projection contract before retrying migration 32.';
    END IF;
END
$$;
-- +goose StatementEnd

TRUNCATE TABLE public.graph_projection_idempotency;
TRUNCATE TABLE public.graph_projection_views CASCADE;

ALTER TABLE public.graph_projection_runs
    ADD COLUMN started_at timestamp with time zone,
    ADD COLUMN generated_at timestamp with time zone,
    ADD COLUMN replaced_at timestamp with time zone,
	ADD COLUMN invalidated_at timestamp with time zone,
    ADD COLUMN retention_policy_json jsonb NOT NULL DEFAULT '{
      "retain_replaced_results": true,
      "retention_count": 5,
      "retention_duration_seconds": 2592000,
      "retain_failed_results": true,
      "failed_retention_count": 20,
      "failed_retention_duration_seconds": 2592000
    }'::jsonb,
    ADD CONSTRAINT graph_projection_runs_state_ck
        CHECK (state = ANY (ARRAY['accepted', 'computing', 'available', 'failed', 'replaced', 'invalidated'])),
    ADD CONSTRAINT graph_projection_runs_retention_policy_ck
        CHECK (jsonb_typeof(retention_policy_json) = 'object');

ALTER TABLE public.graph_projection_views
    ADD COLUMN invalidation_json jsonb,
	ADD COLUMN selected_projection_run_id text,
    ADD CONSTRAINT graph_projection_views_state_ck
        CHECK (state = ANY (ARRAY['creating', 'available', 'refreshing', 'failed', 'invalidated'])),
    ADD CONSTRAINT graph_projection_views_invalidation_ck
        CHECK (invalidation_json IS NULL OR jsonb_typeof(invalidation_json) = 'object');

ALTER TABLE public.graph_projection_views
    ADD CONSTRAINT graph_projection_views_selected_run_fkey
        FOREIGN KEY (selected_projection_run_id)
        REFERENCES public.graph_projection_runs(projection_run_id)
        DEFERRABLE INITIALLY DEFERRED;

ALTER TABLE public.graph_projection_idempotency
    DROP CONSTRAINT graph_projection_idempotency_pkey,
    ADD COLUMN scope_key text NOT NULL,
    ADD CONSTRAINT graph_projection_idempotency_pkey
        PRIMARY KEY (operation, scope_key, idempotency_key);

CREATE UNIQUE INDEX graph_projection_runs_one_active_per_view_idx
    ON public.graph_projection_runs (graph_view_id)
    WHERE state IN ('accepted', 'computing');

CREATE INDEX graph_projection_runs_replaced_retention_idx
    ON public.graph_projection_runs (graph_view_id, replaced_at DESC, projection_run_id ASC)
    WHERE state = 'replaced';

CREATE INDEX graph_projection_runs_failed_retention_idx
    ON public.graph_projection_runs (graph_view_id, completed_at DESC, projection_run_id ASC)
    WHERE state = 'failed';

CREATE INDEX graph_projection_runs_invalidated_retention_idx
    ON public.graph_projection_runs (graph_view_id, invalidated_at DESC, projection_run_id ASC)
    WHERE state = 'invalidated';

CREATE INDEX graph_projection_vertices_run_id_lookup_idx
    ON public.graph_projection_vertices (graph_view_id, projection_run_id, vertex_id);

CREATE INDEX graph_projection_edges_run_id_lookup_idx
    ON public.graph_projection_edges (graph_view_id, projection_run_id, edge_id);

-- +goose Down
DROP INDEX IF EXISTS public.graph_projection_edges_run_id_lookup_idx;
DROP INDEX IF EXISTS public.graph_projection_vertices_run_id_lookup_idx;
DROP INDEX IF EXISTS public.graph_projection_runs_invalidated_retention_idx;
DROP INDEX IF EXISTS public.graph_projection_runs_failed_retention_idx;
DROP INDEX IF EXISTS public.graph_projection_runs_replaced_retention_idx;
DROP INDEX IF EXISTS public.graph_projection_runs_one_active_per_view_idx;

ALTER TABLE public.graph_projection_idempotency
    DROP CONSTRAINT graph_projection_idempotency_pkey,
    DROP COLUMN scope_key,
    ADD CONSTRAINT graph_projection_idempotency_pkey
        PRIMARY KEY (operation, idempotency_key);

ALTER TABLE public.graph_projection_views
	DROP CONSTRAINT graph_projection_views_selected_run_fkey,
    DROP CONSTRAINT graph_projection_views_invalidation_ck,
    DROP CONSTRAINT graph_projection_views_state_ck,
    DROP COLUMN selected_projection_run_id,
	DROP COLUMN invalidation_json;

ALTER TABLE public.graph_projection_runs
    DROP CONSTRAINT graph_projection_runs_retention_policy_ck,
    DROP CONSTRAINT graph_projection_runs_state_ck,
    DROP COLUMN retention_policy_json,
	DROP COLUMN invalidated_at,
    DROP COLUMN replaced_at,
    DROP COLUMN generated_at,
    DROP COLUMN started_at;
