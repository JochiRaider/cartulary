-- +goose Up
-- +goose StatementBegin
DO $$
BEGIN
    LOCK TABLE public.extension_staged_objects IN ACCESS EXCLUSIVE MODE;
    IF EXISTS (SELECT 1 FROM public.extension_staged_objects) THEN
        RAISE EXCEPTION 'operation-scoped streaming requires fresh retained state';
    END IF;
END
$$;
-- +goose StatementEnd

ALTER TABLE public.extension_staged_objects
    ADD COLUMN operation_id text NOT NULL,
    ADD COLUMN reserved_resource_kind text,
    ADD COLUMN reserved_resource_id text,
    ADD CONSTRAINT extension_staged_objects_resource_ck CHECK (
        (reserved_resource_kind IS NULL) = (reserved_resource_id IS NULL)
        AND (reserved_resource_kind IS NULL OR (reserved_resource_kind ~ '^[a-z][a-z0-9_]{0,127}$' AND octet_length(reserved_resource_id) BETWEEN 1 AND 512 AND reserved_resource_id !~ '[[:cntrl:]]'))
    ),
    ALTER COLUMN expected_byte_size DROP NOT NULL,
    ALTER COLUMN expected_sha256 DROP NOT NULL,
    ADD CONSTRAINT extension_staged_objects_operation_ck CHECK (octet_length(operation_id) BETWEEN 1 AND 512 AND operation_id !~ '[[:cntrl:]]'),
    ADD CONSTRAINT extension_staged_objects_seal_ck CHECK (
        (expected_byte_size IS NULL) = (expected_sha256 IS NULL)
        AND (state NOT IN ('ready', 'published') OR expected_byte_size IS NOT NULL)
        AND (state <> 'allocated' OR expected_byte_size IS NULL)
    );

CREATE INDEX extension_staged_objects_operation_idx ON public.extension_staged_objects (operation_id, owner_profile_id, staging_id);

-- +goose Down
-- +goose StatementBegin
DO $$ BEGIN
    RAISE EXCEPTION 'operation-scoped streaming has no in-place downgrade';
END $$;
-- +goose StatementEnd
