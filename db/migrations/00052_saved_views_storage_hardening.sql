-- +goose Up
-- +goose StatementBegin
DO $$
DECLARE
    invalid_owner_scope_count bigint;
    invalid_version_count bigint;
    invalid_timestamp_count bigint;
BEGIN
    SELECT count(*)
      INTO invalid_owner_scope_count
      FROM public.saved_views
     WHERE NOT (
        (scope IN ('private', 'shared') AND owner_user_id IS NOT NULL)
        OR (scope = 'system' AND owner_user_id IS NULL)
     );

    SELECT count(*)
      INTO invalid_version_count
      FROM public.saved_views
     WHERE saved_view_version < 1;

    SELECT count(*)
      INTO invalid_timestamp_count
      FROM public.saved_views
     WHERE updated_at < created_at;

    IF invalid_owner_scope_count > 0
        OR invalid_version_count > 0
        OR invalid_timestamp_count > 0 THEN
        RAISE EXCEPTION USING
            ERRCODE = '23514',
            MESSAGE = format(
                'saved views storage hardening preflight failed: invalid_owner_scope_count=%s invalid_version_count=%s invalid_timestamp_count=%s',
                invalid_owner_scope_count,
                invalid_version_count,
                invalid_timestamp_count
            );
    END IF;
END
$$;
-- +goose StatementEnd

ALTER TABLE public.saved_views
    DROP CONSTRAINT saved_views_owner_scope_ck,
    ADD CONSTRAINT saved_views_owner_scope_ck
        CHECK (
            (scope IN ('private', 'shared') AND owner_user_id IS NOT NULL)
            OR (scope = 'system' AND owner_user_id IS NULL)
        ),
    ADD CONSTRAINT saved_views_version_positive_ck
        CHECK (saved_view_version >= 1),
    ADD CONSTRAINT saved_views_timestamp_order_ck
        CHECK (updated_at >= created_at);

-- +goose Down
-- Tightened Saved Views ownership, version, and timestamp guarantees are
-- intentionally forward-only. Relaxing them would admit state that the active
-- portability and ordinary-write contracts reject.
SELECT 1;
