-- +goose Up
-- Sensitive reveal maps are private adjuncts, never shareable bundle members.
-- No conversion of retained bundles, catalogs, or pending render history.
-- +goose StatementBegin
DO $$ DECLARE snapshots bigint; jobs bigint; imports bigint; previews bigint; BEGIN
    LOCK TABLE public.reporting_snapshots, public.reporting_job_payloads,
        public.reporting_imported_artifact_files, public.report_composition_preview_attempts IN SHARE MODE;
    SELECT count(*) INTO snapshots FROM public.reporting_snapshots;
    SELECT count(*) INTO jobs FROM public.reporting_job_payloads;
    SELECT count(*) INTO imports FROM public.reporting_imported_artifact_files;
    SELECT count(*) INTO previews FROM public.report_composition_preview_attempts;
    IF snapshots + jobs + imports + previews > 0 THEN
        RAISE EXCEPTION 'reporting_cutover_required: snapshots=%, jobs=%, imported_artifacts=%, previews=%; use a matching backup or explicit disposable initialization', snapshots, jobs, imports, previews USING ERRCODE='55000';
    END IF;
END $$;
-- +goose StatementEnd

CREATE TABLE public.reporting_release_reveal_maps (
    release_id uuid NOT NULL,
    reveal_map_sha256 text NOT NULL,
    canonical_bytes bytea NOT NULL,
    created_at timestamptz NOT NULL,
    CONSTRAINT reporting_release_reveal_maps_pk PRIMARY KEY (release_id),
    CONSTRAINT reporting_release_reveal_maps_parent_fk FOREIGN KEY (release_id) REFERENCES public.reporting_render_bundles(release_id) ON UPDATE RESTRICT ON DELETE CASCADE,
    CONSTRAINT reporting_release_reveal_maps_digest_ck CHECK (reveal_map_sha256 ~ '^[a-f0-9]{64}$'),
    CONSTRAINT reporting_release_reveal_maps_bytes_ck CHECK (octet_length(canonical_bytes) > 0)
);

CREATE TABLE public.reporting_preview_reveal_maps (
    preview_attempt_id uuid NOT NULL,
    reveal_map_sha256 text NOT NULL,
    canonical_bytes bytea NOT NULL,
    created_at timestamptz NOT NULL,
    CONSTRAINT reporting_preview_reveal_maps_pk PRIMARY KEY (preview_attempt_id),
    CONSTRAINT reporting_preview_reveal_maps_parent_fk FOREIGN KEY (preview_attempt_id) REFERENCES public.reporting_composition_preview_outputs(preview_attempt_id) ON UPDATE RESTRICT ON DELETE CASCADE,
    CONSTRAINT reporting_preview_reveal_maps_digest_ck CHECK (reveal_map_sha256 ~ '^[a-f0-9]{64}$'),
    CONSTRAINT reporting_preview_reveal_maps_bytes_ck CHECK (octet_length(canonical_bytes) > 0)
);

ALTER TABLE public.reporting_render_bundle_files ADD CONSTRAINT reporting_bundle_public_members_ck CHECK (
    role <> 'sensitive_reveal_map' AND bundle_path NOT LIKE 'internal/%'
    AND media_type <> 'application/vnd.cartulary.reporting-reveal-map+json'
);
ALTER TABLE public.reporting_composition_preview_output_files ADD CONSTRAINT reporting_preview_public_members_ck CHECK (
    role <> 'sensitive_reveal_map' AND bundle_path NOT LIKE 'internal/%'
    AND media_type <> 'application/vnd.cartulary.reporting-reveal-map+json'
);

GRANT SELECT, INSERT ON TABLE public.reporting_release_reveal_maps, public.reporting_preview_reveal_maps TO cartulary_runtime;
GRANT SELECT, INSERT, UPDATE, DELETE, TRUNCATE ON TABLE public.reporting_release_reveal_maps, public.reporting_preview_reveal_maps TO cartulary_recovery;
REVOKE USAGE ON TYPE public.reporting_release_reveal_maps, public.reporting_preview_reveal_maps FROM PUBLIC;
GRANT USAGE ON TYPE public.reporting_release_reveal_maps, public.reporting_preview_reveal_maps TO cartulary_runtime, cartulary_recovery;

-- +goose Down
-- +goose StatementBegin
DO $$ BEGIN RAISE EXCEPTION 'private Reporting artifacts cannot be downgraded into shareable bundles; restore a matching backup' USING ERRCODE='55000'; END $$;
-- +goose StatementEnd
