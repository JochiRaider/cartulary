-- +goose Up
-- PostgreSQL 18.6 consolidated baseline: saved views.
-- Administrator provisioning and final admission are separate lifecycle steps.

CREATE TABLE public.saved_views (
    saved_view_id uuid DEFAULT gen_random_uuid() NOT NULL,
    incident_id uuid NOT NULL,
    view_schema_id text NOT NULL,
    scope text NOT NULL,
    display_name text NOT NULL,
    query_json jsonb NOT NULL,
    layout_json jsonb NOT NULL,
    owner_user_id uuid,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    saved_view_version bigint DEFAULT 1 NOT NULL,
    CONSTRAINT saved_views_owner_scope_ck CHECK ((scope = ANY (ARRAY['private'::text, 'shared'::text])) AND owner_user_id IS NOT NULL OR scope = 'system'::text AND owner_user_id IS NULL),
    CONSTRAINT saved_views_pkey PRIMARY KEY (saved_view_id),
    CONSTRAINT saved_views_scope_check CHECK (scope = ANY (ARRAY['private'::text, 'shared'::text, 'system'::text])),
    CONSTRAINT saved_views_timestamp_order_ck CHECK (updated_at >= created_at),
    CONSTRAINT saved_views_version_positive_ck CHECK (saved_view_version >= 1)
);

REVOKE ALL ON TABLE public.saved_views FROM PUBLIC;

REVOKE ALL ON TYPE public.saved_views FROM PUBLIC;

GRANT INSERT, SELECT, UPDATE, DELETE ON TABLE public.saved_views TO cartulary_runtime;

GRANT INSERT, SELECT, UPDATE, DELETE, TRUNCATE ON TABLE public.saved_views TO cartulary_recovery;

CREATE INDEX saved_views_incident_order_idx ON public.saved_views USING btree (incident_id, updated_at DESC, saved_view_id);

CREATE INDEX saved_views_incident_schema_order_idx ON public.saved_views USING btree (incident_id, view_schema_id, updated_at DESC, saved_view_id);

CREATE INDEX saved_views_owner_lookup_idx ON public.saved_views USING btree (incident_id, owner_user_id) WHERE (scope = 'private'::text);

CREATE INDEX saved_views_owner_user_id_fk_idx ON public.saved_views USING btree (owner_user_id);

ALTER TABLE public.saved_views ADD CONSTRAINT saved_views_incident_id_fkey FOREIGN KEY (incident_id) REFERENCES public.incidents(id) ON DELETE CASCADE ON UPDATE NO ACTION;

ALTER TABLE public.saved_views ADD CONSTRAINT saved_views_owner_user_id_fkey FOREIGN KEY (owner_user_id) REFERENCES public.users(id) ON UPDATE NO ACTION ON DELETE NO ACTION;

-- +goose Down
-- Disposable harness cleanup only; no production downgrade or restore guarantee.
ALTER TABLE public.saved_views DROP CONSTRAINT saved_views_owner_user_id_fkey;
ALTER TABLE public.saved_views DROP CONSTRAINT saved_views_incident_id_fkey;
DROP TABLE public.saved_views;
