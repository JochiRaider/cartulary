-- +goose Up

CREATE INDEX graph_projection_results_cleanup_candidate_idx
    ON public.graph_projection_results (
        source_owner_id,
        published_at,
        projection_result_id
    );

-- +goose Down

DROP INDEX public.graph_projection_results_cleanup_candidate_idx;
