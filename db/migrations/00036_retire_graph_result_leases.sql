-- +goose Up
-- Current Graph consumers retain complete inputs at admission. Existing lease
-- state requires an explicit owner decision; even expired rows are preserved.
LOCK TABLE public.graph_projection_result_leases IN ACCESS EXCLUSIVE MODE;
-- +goose StatementBegin
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM public.graph_projection_result_leases) THEN
        RAISE EXCEPTION 'graph result lease retirement requires an empty table';
    END IF;
END
$$;
-- +goose StatementEnd
DROP TABLE public.graph_projection_result_leases;

-- +goose Down
-- Disposable harness cleanup only; no production downgrade or restore guarantee.
CREATE TABLE public.graph_projection_result_leases (
    lease_id uuid NOT NULL,
    projection_result_id text NOT NULL,
    lease_owner_id text NOT NULL,
    lease_owner_resource_id text NOT NULL,
    lease_purpose text NOT NULL,
    leased_until timestamp with time zone NOT NULL,
    created_at timestamp with time zone NOT NULL,
    renewed_at timestamp with time zone NOT NULL,
    CONSTRAINT graph_projection_result_leases_identity_key UNIQUE (projection_result_id, lease_owner_id, lease_owner_resource_id, lease_purpose),
    CONSTRAINT graph_projection_result_leases_owner_ck CHECK (char_length(lease_owner_id) >= 1 AND char_length(lease_owner_id) <= 255 AND char_length(lease_owner_resource_id) >= 1 AND char_length(lease_owner_resource_id) <= 255 AND lease_purpose ~ '^[a-z][a-z0-9_.:-]{0,254}$'::text),
    CONSTRAINT graph_projection_result_leases_pkey PRIMARY KEY (lease_id),
    CONSTRAINT graph_projection_result_leases_time_ck CHECK (renewed_at >= created_at AND leased_until > renewed_at)
);

REVOKE ALL ON TABLE public.graph_projection_result_leases FROM PUBLIC;

REVOKE ALL ON TYPE public.graph_projection_result_leases FROM PUBLIC;

GRANT INSERT, SELECT, UPDATE, DELETE ON TABLE public.graph_projection_result_leases TO cartulary_runtime;

GRANT INSERT, SELECT, UPDATE, DELETE, TRUNCATE ON TABLE public.graph_projection_result_leases TO cartulary_recovery;

GRANT USAGE ON TYPE public.graph_projection_result_leases TO cartulary_runtime;

GRANT USAGE ON TYPE public.graph_projection_result_leases TO cartulary_recovery;


CREATE INDEX graph_projection_result_leases_expiry_idx ON public.graph_projection_result_leases USING btree (leased_until, projection_result_id);
ALTER TABLE public.graph_projection_result_leases ADD CONSTRAINT graph_projection_result_leases_result_id_fkey FOREIGN KEY (projection_result_id) REFERENCES public.graph_projection_results(projection_result_id) ON DELETE RESTRICT ON UPDATE NO ACTION;
