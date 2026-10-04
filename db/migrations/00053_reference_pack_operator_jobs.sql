-- +goose Up
-- Refuse to relabel an existing human-only Reference Pack import contract.
-- No identity, proof, or historical outcome is synthesized during cutover.
-- +goose StatementBegin
DO $$ DECLARE affected bigint; BEGIN
    LOCK TABLE public.jobs IN ACCESS EXCLUSIVE MODE;
    SELECT count(*) INTO affected FROM public.jobs
      WHERE job_kind='reference_pack.import_v2'
      AND extension_idempotency_identity->>'schema_id' IS DISTINCT FROM 'cartulary.route_scoped_idempotency_identity.v2';
    IF affected > 0 THEN
        RAISE EXCEPTION 'reference_pack_cutover_required: retired_import_attribution=%; preserve a matching backup and explicitly reset disposable development data', affected USING ERRCODE='55000';
    END IF;
END $$;
-- +goose StatementEnd
ALTER TABLE public.jobs ADD COLUMN submitting_operator_operation_id uuid;
ALTER TABLE public.jobs ALTER COLUMN submitted_by_user_id DROP NOT NULL;
ALTER TABLE public.jobs ADD CONSTRAINT jobs_submitting_actor_ck CHECK (
    (submitted_by_user_id IS NOT NULL AND submitting_operator_operation_id IS NULL)
    OR (
        submitted_by_user_id IS NULL
        AND submitting_operator_operation_id IS NOT NULL
        AND submitting_operator_operation_id <> '00000000-0000-0000-0000-000000000000'::uuid
        AND scope_kind='deployment' AND incident_id IS NULL
        AND auth_policy='deployment_admin' AND extension_owner_profile_id IS NOT NULL
        AND (expired_at IS NOT NULL OR (
            extension_idempotency_identity->>'schema_id'='cartulary.route_scoped_idempotency_identity.v2'
            AND extension_idempotency_identity->>'actor_kind'='local_operator'
            AND extension_idempotency_identity->'actor_user_id'='null'::jsonb
            AND extension_idempotency_identity->>'operator_operation_id'=submitting_operator_operation_id::text
            AND extension_idempotency_identity->>'client_txn_id'=submitting_operator_operation_id::text
        ))
    ) IS TRUE
);

-- +goose Down
-- +goose StatementBegin
DO $$ BEGIN
    LOCK TABLE public.jobs IN ACCESS EXCLUSIVE MODE;
    IF EXISTS(SELECT 1 FROM public.jobs WHERE submitted_by_user_id IS NULL) THEN
        RAISE EXCEPTION 'local operator jobs cannot be represented by the historical schema' USING ERRCODE='55000';
    END IF;
END $$;
-- +goose StatementEnd
ALTER TABLE public.jobs DROP CONSTRAINT jobs_submitting_actor_ck;
ALTER TABLE public.jobs ALTER COLUMN submitted_by_user_id SET NOT NULL;
ALTER TABLE public.jobs DROP COLUMN submitting_operator_operation_id;
