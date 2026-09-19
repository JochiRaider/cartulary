-- +goose Up
CREATE INDEX jobs_restored_nonterminal_order_idx
    ON public.jobs (extension_owner_profile_id, job_kind, job_id)
    WHERE status IN ('queued', 'running', 'cancel_requested');

-- +goose Down
DROP INDEX public.jobs_restored_nonterminal_order_idx;
