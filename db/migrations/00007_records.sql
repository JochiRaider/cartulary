-- +goose Up
--
-- Name: change_set_mutations; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.change_set_mutations (
    change_set_id uuid NOT NULL,
    sequence_no integer NOT NULL,
    target_kind text NOT NULL,
    target_id text NOT NULL,
    operation_kind text NOT NULL,
    before_version_id text,
    after_version_id text,
    before_value jsonb,
    after_value jsonb
);

--
-- Name: change_sets; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.change_sets (
    change_set_id uuid DEFAULT gen_random_uuid() NOT NULL,
    incident_id uuid NOT NULL,
    actor_user_id uuid NOT NULL,
    source text NOT NULL,
    reason text,
    client_txn_id text,
    request_id text,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);

--
-- Name: record_history_entry_refs; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.record_history_entry_refs (
    history_entry_ref text NOT NULL,
    record_id uuid NOT NULL,
    change_set_id uuid NOT NULL,
    mutation_sequence_no integer NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);

--
-- Name: record_revisions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.record_revisions (
    revision_id bigint NOT NULL,
    change_set_id uuid NOT NULL,
    record_id uuid NOT NULL,
    row_version bigint NOT NULL,
    before_json jsonb,
    after_json jsonb,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);

--
-- Name: record_revisions_revision_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.record_revisions_revision_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

--
-- Name: record_revisions_revision_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.record_revisions_revision_id_seq OWNED BY public.record_revisions.revision_id;

--
-- Name: records; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.records (
    record_id uuid DEFAULT gen_random_uuid() NOT NULL,
    incident_id uuid NOT NULL,
    record_type text NOT NULL,
    created_by_user_id uuid NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_by_user_id uuid NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    row_version bigint DEFAULT 1 NOT NULL,
    deleted_at timestamp with time zone,
    deleted_by_user_id uuid,
    CONSTRAINT records_delete_state_ck CHECK ((((deleted_at IS NULL) AND (deleted_by_user_id IS NULL)) OR ((deleted_at IS NOT NULL) AND (deleted_by_user_id IS NOT NULL)))),
    CONSTRAINT records_record_type_check CHECK ((record_type = ANY (ARRAY['timeline_event'::text, 'host'::text, 'identity'::text, 'party'::text, 'indicator'::text, 'artifact'::text, 'task_request'::text, 'decision'::text, 'evidence'::text, 'assessment'::text]))),
    CONSTRAINT records_row_version_check CHECK ((row_version >= 1))
);

--
-- Name: record_revisions revision_id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.record_revisions ALTER COLUMN revision_id SET DEFAULT nextval('public.record_revisions_revision_id_seq'::regclass);

--
-- Name: change_set_mutations change_set_mutations_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.change_set_mutations
    ADD CONSTRAINT change_set_mutations_pkey PRIMARY KEY (change_set_id, sequence_no);

--
-- Name: change_sets change_sets_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.change_sets
    ADD CONSTRAINT change_sets_pkey PRIMARY KEY (change_set_id);

--
-- Name: record_history_entry_refs record_history_entry_refs_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.record_history_entry_refs
    ADD CONSTRAINT record_history_entry_refs_pkey PRIMARY KEY (history_entry_ref);

--
-- Name: record_history_entry_refs record_history_entry_refs_record_id_change_set_id_mutation__key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.record_history_entry_refs
    ADD CONSTRAINT record_history_entry_refs_record_id_change_set_id_mutation__key UNIQUE (record_id, change_set_id, mutation_sequence_no);

--
-- Name: record_revisions record_revisions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.record_revisions
    ADD CONSTRAINT record_revisions_pkey PRIMARY KEY (revision_id);

--
-- Name: record_revisions record_revisions_record_id_row_version_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.record_revisions
    ADD CONSTRAINT record_revisions_record_id_row_version_key UNIQUE (record_id, row_version);

--
-- Name: records records_incident_id_record_id_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.records
    ADD CONSTRAINT records_incident_id_record_id_key UNIQUE (incident_id, record_id);

--
-- Name: records records_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.records
    ADD CONSTRAINT records_pkey PRIMARY KEY (record_id);

--
-- Name: change_set_mutations_target_lookup_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX change_set_mutations_target_lookup_idx ON public.change_set_mutations USING btree (target_kind, target_id);

--
-- Name: change_sets_incident_lookup_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX change_sets_incident_lookup_idx ON public.change_sets USING btree (incident_id, created_at DESC, change_set_id DESC);

--
-- Name: record_history_entry_refs_record_lookup_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX record_history_entry_refs_record_lookup_idx ON public.record_history_entry_refs USING btree (record_id, change_set_id, mutation_sequence_no);

--
-- Name: record_revisions_record_lookup_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX record_revisions_record_lookup_idx ON public.record_revisions USING btree (record_id, row_version DESC);

--
-- Name: records_active_incident_type_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX records_active_incident_type_idx ON public.records USING btree (incident_id, record_type, updated_at DESC, record_id) WHERE (deleted_at IS NULL);

--
-- Name: records_incident_lookup_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX records_incident_lookup_idx ON public.records USING btree (incident_id, record_id);

--
-- Name: change_set_mutations change_set_mutations_change_set_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.change_set_mutations
    ADD CONSTRAINT change_set_mutations_change_set_id_fkey FOREIGN KEY (change_set_id) REFERENCES public.change_sets(change_set_id) ON UPDATE NO ACTION ON DELETE CASCADE;

--
-- Name: change_sets change_sets_actor_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.change_sets
    ADD CONSTRAINT change_sets_actor_user_id_fkey FOREIGN KEY (actor_user_id) REFERENCES public.users(id) ON UPDATE NO ACTION ON DELETE NO ACTION;

--
-- Name: change_sets change_sets_incident_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.change_sets
    ADD CONSTRAINT change_sets_incident_id_fkey FOREIGN KEY (incident_id) REFERENCES public.incidents(id) ON UPDATE NO ACTION ON DELETE CASCADE;

--
-- Name: record_history_entry_refs record_history_entry_refs_change_set_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.record_history_entry_refs
    ADD CONSTRAINT record_history_entry_refs_change_set_id_fkey FOREIGN KEY (change_set_id) REFERENCES public.change_sets(change_set_id) ON UPDATE NO ACTION ON DELETE CASCADE;

--
-- Name: record_history_entry_refs record_history_entry_refs_change_set_id_mutation_sequence__fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.record_history_entry_refs
    ADD CONSTRAINT record_history_entry_refs_change_set_id_mutation_sequence__fkey FOREIGN KEY (change_set_id, mutation_sequence_no) REFERENCES public.change_set_mutations(change_set_id, sequence_no) ON UPDATE NO ACTION ON DELETE CASCADE;

--
-- Name: record_history_entry_refs record_history_entry_refs_record_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.record_history_entry_refs
    ADD CONSTRAINT record_history_entry_refs_record_id_fkey FOREIGN KEY (record_id) REFERENCES public.records(record_id) ON UPDATE NO ACTION ON DELETE CASCADE;

--
-- Name: record_revisions record_revisions_change_set_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.record_revisions
    ADD CONSTRAINT record_revisions_change_set_id_fkey FOREIGN KEY (change_set_id) REFERENCES public.change_sets(change_set_id) ON UPDATE NO ACTION ON DELETE CASCADE;

--
-- Name: record_revisions record_revisions_record_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.record_revisions
    ADD CONSTRAINT record_revisions_record_id_fkey FOREIGN KEY (record_id) REFERENCES public.records(record_id) ON UPDATE NO ACTION ON DELETE CASCADE;

--
-- Name: records records_created_by_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.records
    ADD CONSTRAINT records_created_by_user_id_fkey FOREIGN KEY (created_by_user_id) REFERENCES public.users(id) ON UPDATE NO ACTION ON DELETE NO ACTION;

--
-- Name: records records_deleted_by_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.records
    ADD CONSTRAINT records_deleted_by_user_id_fkey FOREIGN KEY (deleted_by_user_id) REFERENCES public.users(id) ON UPDATE NO ACTION ON DELETE NO ACTION;

--
-- Name: records records_incident_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.records
    ADD CONSTRAINT records_incident_id_fkey FOREIGN KEY (incident_id) REFERENCES public.incidents(id) ON UPDATE NO ACTION ON DELETE CASCADE;

--
-- Name: records records_updated_by_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.records
    ADD CONSTRAINT records_updated_by_user_id_fkey FOREIGN KEY (updated_by_user_id) REFERENCES public.users(id) ON UPDATE NO ACTION ON DELETE NO ACTION;

CREATE INDEX change_sets_actor_user_id_fk_idx ON public.change_sets (actor_user_id);
CREATE INDEX record_history_entry_refs_change_set_id_mutation_seque_824b16fa ON public.record_history_entry_refs (change_set_id, mutation_sequence_no);
CREATE INDEX record_history_entry_refs_change_set_id_fk_idx ON public.record_history_entry_refs (change_set_id);
CREATE INDEX record_revisions_change_set_id_fk_idx ON public.record_revisions (change_set_id);
CREATE INDEX records_created_by_user_id_fk_idx ON public.records (created_by_user_id);
CREATE INDEX records_deleted_by_user_id_fk_idx ON public.records (deleted_by_user_id);
CREATE INDEX records_updated_by_user_id_fk_idx ON public.records (updated_by_user_id);

-- +goose Down
ALTER TABLE public.record_revisions ALTER COLUMN revision_id DROP DEFAULT;
ALTER SEQUENCE public.record_revisions_revision_id_seq OWNED BY NONE;
DROP SEQUENCE public.record_revisions_revision_id_seq;
DROP TABLE public.record_history_entry_refs ;
DROP TABLE public.record_revisions ;
DROP TABLE public.change_set_mutations ;
DROP TABLE public.change_sets ;
DROP TABLE public.records ;
