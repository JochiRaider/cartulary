-- +goose Up

-- +goose StatementBegin
DO $$
BEGIN
    IF EXISTS (
        SELECT 1
          FROM public.import_sessions
         WHERE session_status = 'applying'
    ) OR EXISTS (
        SELECT 1
          FROM public.import_units
         WHERE unit_status = 'applying'
    ) THEN
        RAISE EXCEPTION
            'migration 00049 requires no active import apply; reset or reseed ambiguous applying state';
    END IF;
END
$$;
-- +goose StatementEnd

ALTER TABLE public.import_units
    DROP CONSTRAINT import_units_unit_status_check,
    ADD CONSTRAINT import_units_unit_status_check
        CHECK (unit_status = ANY (ARRAY[
            'discovered'::text,
            'selected'::text,
            'mapped'::text,
            'ready'::text,
            'skipped'::text,
            'applying'::text,
            'applied'::text,
            'rejected'::text,
            'failed'::text,
            'canceled'::text
        ]));

CREATE TABLE public.import_apply_unit_plans (
    import_session_id uuid NOT NULL,
    import_unit_id uuid NOT NULL,
    apply_job_id uuid NOT NULL,
    discovery_sequence integer NOT NULL,
    source_file_kind text NOT NULL,
    source_content_sha256 text NOT NULL,
    source_stream_ref text NOT NULL,
    source_rows_sha256 text NOT NULL,
    parser_profile_id text NOT NULL,
    parser_version text NOT NULL,
    locator_kind text NOT NULL,
    locator text NOT NULL,
    source_rect_a1 text NOT NULL,
    mapping_fingerprint text NOT NULL,
    approved_mapping_json jsonb NOT NULL,
    approved_mapping_sha256 text NOT NULL,
    target_kind text NOT NULL,
    target_view_schema_id text,
    extension_profile_id text,
    owner_binding_id text NOT NULL,
    target_registry_sha256 text NOT NULL,
    admitted_at timestamp with time zone NOT NULL,
    CONSTRAINT import_apply_unit_plans_pk
        PRIMARY KEY (import_session_id, import_unit_id, apply_job_id),
    CONSTRAINT import_apply_unit_plans_unit_uq UNIQUE (import_unit_id),
    CONSTRAINT import_apply_unit_plans_session_sequence_uq
        UNIQUE (import_session_id, discovery_sequence),
    CONSTRAINT import_apply_unit_plans_sequence_ck
        CHECK (discovery_sequence > 0),
    CONSTRAINT import_apply_unit_plans_digest_ck
        CHECK (
            source_content_sha256 ~ '^[0-9a-f]{64}$'
            AND source_rows_sha256 ~ '^[0-9a-f]{64}$'
            AND mapping_fingerprint ~ '^[0-9a-f]{64}$'
            AND approved_mapping_sha256 ~ '^[0-9a-f]{64}$'
            AND target_registry_sha256 ~ '^[0-9a-f]{64}$'
        ),
    CONSTRAINT import_apply_unit_plans_mapping_json_ck
        CHECK (jsonb_typeof(approved_mapping_json) = 'object'),
    CONSTRAINT import_apply_unit_plans_target_kind_ck
        CHECK (target_kind = ANY (ARRAY[
            'view_schema'::text,
            'network_flow_table'::text
        ])),
    CONSTRAINT import_apply_unit_plans_target_shape_ck
        CHECK (
            (
                target_kind = 'view_schema'
                AND target_view_schema_id IS NOT NULL
                AND extension_profile_id IS NULL
            )
            OR
            (
                target_kind = 'network_flow_table'
                AND target_view_schema_id IS NULL
                AND extension_profile_id IS NOT NULL
            )
        ),
    CONSTRAINT import_apply_unit_plans_session_fk
        FOREIGN KEY (import_session_id)
        REFERENCES public.import_sessions(import_session_id)
        ON DELETE CASCADE,
    CONSTRAINT import_apply_unit_plans_unit_fk
        FOREIGN KEY (import_unit_id)
        REFERENCES public.import_units(import_unit_id)
        ON DELETE CASCADE,
    CONSTRAINT import_apply_unit_plans_job_fk
        FOREIGN KEY (apply_job_id)
        REFERENCES public.jobs(job_id)
);

CREATE TABLE public.import_unit_apply_outcomes (
    import_unit_apply_outcome_id uuid DEFAULT gen_random_uuid() PRIMARY KEY,
    import_session_id uuid NOT NULL,
    import_unit_id uuid NOT NULL,
    apply_job_id uuid NOT NULL,
    discovery_sequence integer NOT NULL,
    unit_commit_id text NOT NULL,
    outcome_status text NOT NULL,
    actor_user_id uuid NOT NULL,
    target_kind text NOT NULL,
    target_view_schema_id text,
    extension_profile_id text,
    owner_binding_id text NOT NULL,
    source_content_sha256 text NOT NULL,
    mapping_fingerprint text NOT NULL,
    owner_result_json jsonb DEFAULT '{}'::jsonb NOT NULL,
    resource_refs_json jsonb DEFAULT '[]'::jsonb NOT NULL,
    change_set_id uuid,
    error_code text,
    reason_code text,
    committed_at timestamp with time zone NOT NULL,
    CONSTRAINT import_unit_apply_outcomes_unit_uq UNIQUE (import_unit_id),
    CONSTRAINT import_unit_apply_outcomes_commit_uq UNIQUE (unit_commit_id),
    CONSTRAINT import_unit_apply_outcomes_session_sequence_uq
        UNIQUE (import_session_id, discovery_sequence),
    CONSTRAINT import_unit_apply_outcomes_sequence_ck
        CHECK (discovery_sequence > 0),
    CONSTRAINT import_unit_apply_outcomes_commit_id_ck
        CHECK (unit_commit_id ~ '^import-unit:[0-9a-f-]{36}:[0-9a-f-]{36}$'),
    CONSTRAINT import_unit_apply_outcomes_status_ck
        CHECK (outcome_status = ANY (ARRAY[
            'applied'::text,
            'failed'::text,
            'canceled'::text
        ])),
    CONSTRAINT import_unit_apply_outcomes_target_kind_ck
        CHECK (target_kind = ANY (ARRAY[
            'view_schema'::text,
            'network_flow_table'::text
        ])),
    CONSTRAINT import_unit_apply_outcomes_target_shape_ck
        CHECK (
            (
                target_kind = 'view_schema'
                AND target_view_schema_id IS NOT NULL
                AND extension_profile_id IS NULL
                AND change_set_id IS NOT NULL
            )
            OR (
                target_kind = 'network_flow_table'
                AND target_view_schema_id IS NULL
                AND extension_profile_id IS NOT NULL
                AND change_set_id IS NULL
            )
            OR (
                outcome_status <> 'applied'
                AND (
                    (target_kind = 'view_schema' AND target_view_schema_id IS NOT NULL
                        AND extension_profile_id IS NULL)
                    OR
                    (target_kind = 'network_flow_table' AND target_view_schema_id IS NULL
                        AND extension_profile_id IS NOT NULL)
                )
            )
        ),
    CONSTRAINT import_unit_apply_outcomes_source_digest_ck
        CHECK (source_content_sha256 ~ '^[0-9a-f]{64}$'),
    CONSTRAINT import_unit_apply_outcomes_mapping_digest_ck
        CHECK (mapping_fingerprint ~ '^[0-9a-f]{64}$'),
    CONSTRAINT import_unit_apply_outcomes_owner_result_ck
        CHECK (jsonb_typeof(owner_result_json) IN ('object', 'array')),
    CONSTRAINT import_unit_apply_outcomes_resource_refs_ck
        CHECK (jsonb_typeof(resource_refs_json) = 'array'),
    CONSTRAINT import_unit_apply_outcomes_error_shape_ck
        CHECK (
            (outcome_status = 'applied' AND error_code IS NULL AND reason_code IS NULL)
            OR
            (outcome_status IN ('failed', 'canceled') AND error_code IS NOT NULL)
        ),
    CONSTRAINT import_unit_apply_outcomes_session_fk
        FOREIGN KEY (import_session_id)
        REFERENCES public.import_sessions(import_session_id)
        ON DELETE CASCADE,
    CONSTRAINT import_unit_apply_outcomes_unit_fk
        FOREIGN KEY (import_unit_id)
        REFERENCES public.import_units(import_unit_id)
        ON DELETE CASCADE,
    CONSTRAINT import_unit_apply_outcomes_job_fk
        FOREIGN KEY (apply_job_id)
        REFERENCES public.jobs(job_id),
    CONSTRAINT import_unit_apply_outcomes_plan_fk
        FOREIGN KEY (import_session_id, import_unit_id, apply_job_id)
        REFERENCES public.import_apply_unit_plans(
            import_session_id,
            import_unit_id,
            apply_job_id
        ),
    CONSTRAINT import_unit_apply_outcomes_actor_fk
        FOREIGN KEY (actor_user_id)
        REFERENCES public.users(id),
    CONSTRAINT import_unit_apply_outcomes_change_set_fk
        FOREIGN KEY (change_set_id)
        REFERENCES public.change_sets(change_set_id)
        ON DELETE CASCADE
);

CREATE INDEX import_unit_apply_outcomes_finalize_idx
    ON public.import_unit_apply_outcomes (
        import_session_id,
        discovery_sequence,
        import_unit_id
    );

-- +goose Down

DROP TABLE IF EXISTS public.import_unit_apply_outcomes;
DROP TABLE IF EXISTS public.import_apply_unit_plans;

UPDATE public.import_units
   SET unit_status = 'failed'
 WHERE unit_status = 'canceled';

ALTER TABLE public.import_units
    DROP CONSTRAINT import_units_unit_status_check,
    ADD CONSTRAINT import_units_unit_status_check
        CHECK (unit_status = ANY (ARRAY[
            'discovered'::text,
            'selected'::text,
            'mapped'::text,
            'ready'::text,
            'skipped'::text,
            'applying'::text,
            'applied'::text,
            'rejected'::text,
            'failed'::text
        ]));
