-- +goose Up
--
-- Name: saved_views; Type: TABLE; Schema: public; Owner: -
--

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
    CONSTRAINT saved_views_owner_scope_ck CHECK (
        (scope IN ('private', 'shared') AND owner_user_id IS NOT NULL)
        OR (scope = 'system' AND owner_user_id IS NULL)
    ),
    CONSTRAINT saved_views_scope_check CHECK ((scope = ANY (ARRAY['private'::text, 'shared'::text, 'system'::text]))),
    CONSTRAINT saved_views_version_positive_ck CHECK (saved_view_version >= 1),
    CONSTRAINT saved_views_timestamp_order_ck CHECK (updated_at >= created_at)
);

--
-- Name: saved_views saved_views_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.saved_views
    ADD CONSTRAINT saved_views_pkey PRIMARY KEY (saved_view_id);

--
-- Name: saved_views_incident_order_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX saved_views_incident_order_idx ON public.saved_views USING btree (incident_id, updated_at DESC, saved_view_id);

--
-- Name: saved_views_owner_lookup_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX saved_views_owner_lookup_idx ON public.saved_views USING btree (incident_id, owner_user_id) WHERE (scope = 'private'::text);

--
-- Name: saved_views saved_views_incident_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.saved_views
    ADD CONSTRAINT saved_views_incident_id_fkey FOREIGN KEY (incident_id) REFERENCES public.incidents(id) ON UPDATE NO ACTION ON DELETE CASCADE;

--
-- Name: saved_views saved_views_owner_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.saved_views
    ADD CONSTRAINT saved_views_owner_user_id_fkey FOREIGN KEY (owner_user_id) REFERENCES public.users(id) ON UPDATE NO ACTION ON DELETE NO ACTION;

CREATE INDEX saved_views_owner_user_id_fk_idx ON public.saved_views (owner_user_id);

-- +goose Down
DROP TABLE public.saved_views;
