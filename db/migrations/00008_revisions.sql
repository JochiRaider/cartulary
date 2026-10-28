-- +goose Up
-- PostgreSQL 18.6 consolidated baseline: revisions.
-- Administrator provisioning and final admission are separate lifecycle steps.

CREATE SEQUENCE public.record_revisions_revision_id_seq AS bigint START WITH 1 INCREMENT BY 1 MINVALUE 1 MAXVALUE 9223372036854775807 CACHE 1 NO CYCLE;

-- +goose StatementBegin
CREATE FUNCTION public.change_set_mutations_history_ids_are_canonical(candidate uuid[])
 RETURNS boolean
 LANGUAGE sql
 IMMUTABLE PARALLEL SAFE STRICT
 SET search_path = pg_catalog, public
AS $$
    SELECT
        COALESCE(array_ndims(candidate), 1) = 1
        AND array_position(candidate, NULL::uuid) IS NULL
        AND array_position(candidate, '00000000-0000-0000-0000-000000000000'::uuid) IS NULL
        AND candidate = ARRAY(
            SELECT DISTINCT member
              FROM unnest(candidate) AS values_row(member)
             ORDER BY member
        );
$$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION public.change_set_mutations_history_ids_are_canonical(candidate uuid[]) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION public.change_set_mutations_history_ids_are_canonical(candidate uuid[]) TO cartulary_runtime;

GRANT EXECUTE ON FUNCTION public.change_set_mutations_history_ids_are_canonical(candidate uuid[]) TO cartulary_recovery;

CREATE TABLE public.change_set_mutations (
    change_set_id uuid NOT NULL,
    sequence_no integer NOT NULL,
    target_kind text NOT NULL,
    target_id text NOT NULL,
    operation_kind text NOT NULL,
    before_version_id text,
    after_version_id text,
    before_value jsonb,
    after_value jsonb,
    history_record_ids uuid[] NOT NULL,
    history_entry_record_ids uuid[] NOT NULL,
    CONSTRAINT change_set_mutations_history_entry_record_ids_ck CHECK (public.change_set_mutations_history_ids_are_canonical(history_entry_record_ids)),
    CONSTRAINT change_set_mutations_history_entry_subset_ck CHECK (history_entry_record_ids <@ history_record_ids),
    CONSTRAINT change_set_mutations_history_record_ids_ck CHECK (public.change_set_mutations_history_ids_are_canonical(history_record_ids)),
    CONSTRAINT change_set_mutations_pkey PRIMARY KEY (change_set_id, sequence_no)
);

REVOKE ALL ON TABLE public.change_set_mutations FROM PUBLIC;

REVOKE ALL ON TYPE public.change_set_mutations FROM PUBLIC;

GRANT INSERT, SELECT, UPDATE, DELETE ON TABLE public.change_set_mutations TO cartulary_runtime;

GRANT INSERT, SELECT, UPDATE, DELETE, TRUNCATE ON TABLE public.change_set_mutations TO cartulary_recovery;

CREATE TABLE public.change_sets (
    change_set_id uuid DEFAULT gen_random_uuid() NOT NULL,
    incident_id uuid NOT NULL,
    actor_user_id uuid NOT NULL,
    source text NOT NULL,
    reason text,
    client_txn_id text,
    request_id text,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT change_sets_pkey PRIMARY KEY (change_set_id)
);

REVOKE ALL ON TABLE public.change_sets FROM PUBLIC;

REVOKE ALL ON TYPE public.change_sets FROM PUBLIC;

GRANT INSERT, SELECT, UPDATE, DELETE ON TABLE public.change_sets TO cartulary_runtime;

GRANT INSERT, SELECT, UPDATE, DELETE, TRUNCATE ON TABLE public.change_sets TO cartulary_recovery;

CREATE TABLE public.record_history_entry_refs (
    history_entry_ref text NOT NULL,
    record_id uuid NOT NULL,
    change_set_id uuid NOT NULL,
    mutation_sequence_no integer NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT record_history_entry_refs_pkey PRIMARY KEY (history_entry_ref),
    CONSTRAINT record_history_entry_refs_record_id_change_set_id_mutation__key UNIQUE (record_id, change_set_id, mutation_sequence_no)
);

REVOKE ALL ON TABLE public.record_history_entry_refs FROM PUBLIC;

REVOKE ALL ON TYPE public.record_history_entry_refs FROM PUBLIC;

GRANT INSERT, SELECT, UPDATE, DELETE ON TABLE public.record_history_entry_refs TO cartulary_runtime;

GRANT INSERT, SELECT, UPDATE, DELETE, TRUNCATE ON TABLE public.record_history_entry_refs TO cartulary_recovery;

CREATE TABLE public.record_revision_conflict_facts (
    revision_id bigint NOT NULL,
    field_key text NOT NULL,
    before_present boolean NOT NULL,
    before_value jsonb,
    after_present boolean NOT NULL,
    after_value jsonb,
    CONSTRAINT record_revision_conflict_facts_after_value_ck CHECK (after_present = (after_value IS NOT NULL)),
    CONSTRAINT record_revision_conflict_facts_before_value_ck CHECK (before_present = (before_value IS NOT NULL)),
    CONSTRAINT record_revision_conflict_facts_field_key_ck CHECK (field_key = btrim(field_key) AND field_key <> ''::text),
    CONSTRAINT record_revision_conflict_facts_pkey PRIMARY KEY (revision_id, field_key),
    CONSTRAINT record_revision_conflict_facts_presence_ck CHECK (before_present OR after_present)
);

REVOKE ALL ON TABLE public.record_revision_conflict_facts FROM PUBLIC;

REVOKE ALL ON TYPE public.record_revision_conflict_facts FROM PUBLIC;

GRANT INSERT, SELECT, UPDATE, DELETE ON TABLE public.record_revision_conflict_facts TO cartulary_runtime;

GRANT INSERT, SELECT, UPDATE, DELETE, TRUNCATE ON TABLE public.record_revision_conflict_facts TO cartulary_recovery;

CREATE TABLE public.record_revisions (
    revision_id bigint DEFAULT nextval('record_revisions_revision_id_seq'::regclass) NOT NULL,
    change_set_id uuid NOT NULL,
    record_id uuid NOT NULL,
    row_version bigint NOT NULL,
    before_json jsonb,
    after_json jsonb,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT record_revisions_pkey PRIMARY KEY (revision_id),
    CONSTRAINT record_revisions_record_id_row_version_key UNIQUE (record_id, row_version)
);

REVOKE ALL ON TABLE public.record_revisions FROM PUBLIC;

REVOKE ALL ON TYPE public.record_revisions FROM PUBLIC;

GRANT INSERT, SELECT, UPDATE, DELETE ON TABLE public.record_revisions TO cartulary_runtime;

GRANT INSERT, SELECT, UPDATE, DELETE, TRUNCATE ON TABLE public.record_revisions TO cartulary_recovery;

-- +goose StatementBegin
CREATE FUNCTION public.revisions_incident_bundle_sequence_begin_v1()
 RETURNS bigint
 LANGUAGE plpgsql
 SECURITY DEFINER
 SET search_path = pg_catalog, public
AS $$
DECLARE
    original_next bigint;
BEGIN
    SELECT sequence_state.last_value +
           CASE WHEN sequence_state.is_called THEN sequence_catalog.seqincrement ELSE 0 END
      INTO STRICT original_next
      FROM public.record_revisions_revision_id_seq AS sequence_state
      JOIN pg_catalog.pg_sequence AS sequence_catalog
        ON sequence_catalog.seqrelid =
           'public.record_revisions_revision_id_seq'::pg_catalog.regclass;

    IF original_next < 1 THEN
        RAISE EXCEPTION USING
            ERRCODE = '22003',
            MESSAGE = 'record revision sequence has an invalid next value';
    END IF;

    EXECUTE pg_catalog.format(
        'ALTER SEQUENCE public.record_revisions_revision_id_seq RESTART WITH %s',
        original_next
    );
    RETURN original_next;
END;
$$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION public.revisions_incident_bundle_sequence_begin_v1() FROM PUBLIC;

GRANT EXECUTE ON FUNCTION public.revisions_incident_bundle_sequence_begin_v1() TO cartulary_runtime;

GRANT EXECUTE ON FUNCTION public.revisions_incident_bundle_sequence_begin_v1() TO cartulary_recovery;

-- +goose StatementBegin
CREATE FUNCTION public.revisions_incident_bundle_sequence_finish_v1(original_next bigint)
 RETURNS bigint
 LANGUAGE plpgsql
 SECURITY DEFINER
 SET search_path = pg_catalog, public
AS $$
DECLARE
    imported_max bigint;
    repaired_next bigint;
BEGIN
    IF original_next < 1 THEN
        RAISE EXCEPTION USING
            ERRCODE = '22003',
            MESSAGE = 'record revision sequence original next value is invalid';
    END IF;

    SELECT max(revision_id)
      INTO imported_max
      FROM public.record_revisions;

    IF imported_max = 9223372036854775807 THEN
        RAISE EXCEPTION USING
            ERRCODE = '22003',
            MESSAGE = 'record revision sequence is exhausted';
    END IF;

    repaired_next := greatest(
        original_next,
        CASE WHEN imported_max IS NULL THEN 1::bigint ELSE imported_max + 1 END
    );
    EXECUTE pg_catalog.format(
        'ALTER SEQUENCE public.record_revisions_revision_id_seq RESTART WITH %s',
        repaired_next
    );
    RETURN repaired_next;
END;
$$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION public.revisions_incident_bundle_sequence_finish_v1(original_next bigint) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION public.revisions_incident_bundle_sequence_finish_v1(original_next bigint) TO cartulary_runtime;

GRANT EXECUTE ON FUNCTION public.revisions_incident_bundle_sequence_finish_v1(original_next bigint) TO cartulary_recovery;

CREATE INDEX change_set_mutations_history_record_ids_idx ON public.change_set_mutations USING gin (history_record_ids);

CREATE INDEX change_set_mutations_target_lookup_idx ON public.change_set_mutations USING btree (target_kind, target_id);

CREATE INDEX change_sets_actor_user_id_fk_idx ON public.change_sets USING btree (actor_user_id);

CREATE INDEX change_sets_incident_lookup_idx ON public.change_sets USING btree (incident_id, created_at DESC, change_set_id DESC);

CREATE INDEX record_history_entry_refs_change_set_id_fk_idx ON public.record_history_entry_refs USING btree (change_set_id);

CREATE INDEX record_history_entry_refs_change_set_id_mutation_seque_824b16fa ON public.record_history_entry_refs USING btree (change_set_id, mutation_sequence_no);

CREATE INDEX record_revisions_change_set_id_fk_idx ON public.record_revisions USING btree (change_set_id);

CREATE INDEX record_revisions_record_lookup_idx ON public.record_revisions USING btree (record_id, row_version DESC);

ALTER TABLE public.change_set_mutations ADD CONSTRAINT change_set_mutations_change_set_id_fkey FOREIGN KEY (change_set_id) REFERENCES public.change_sets(change_set_id) ON DELETE CASCADE ON UPDATE NO ACTION;

ALTER TABLE public.change_sets ADD CONSTRAINT change_sets_actor_user_id_fkey FOREIGN KEY (actor_user_id) REFERENCES public.users(id) ON UPDATE NO ACTION ON DELETE NO ACTION;

ALTER TABLE public.change_sets ADD CONSTRAINT change_sets_incident_id_fkey FOREIGN KEY (incident_id) REFERENCES public.incidents(id) ON DELETE CASCADE ON UPDATE NO ACTION;

ALTER TABLE public.record_history_entry_refs ADD CONSTRAINT record_history_entry_refs_change_set_id_fkey FOREIGN KEY (change_set_id) REFERENCES public.change_sets(change_set_id) ON DELETE CASCADE ON UPDATE NO ACTION;

ALTER TABLE public.record_history_entry_refs ADD CONSTRAINT record_history_entry_refs_change_set_id_mutation_sequence__fkey FOREIGN KEY (change_set_id, mutation_sequence_no) REFERENCES public.change_set_mutations(change_set_id, sequence_no) ON DELETE CASCADE ON UPDATE NO ACTION;

ALTER TABLE public.record_history_entry_refs ADD CONSTRAINT record_history_entry_refs_record_id_fkey FOREIGN KEY (record_id) REFERENCES public.records(record_id) ON DELETE CASCADE ON UPDATE NO ACTION;

ALTER TABLE public.record_revision_conflict_facts ADD CONSTRAINT record_revision_conflict_facts_revision_id_fkey FOREIGN KEY (revision_id) REFERENCES public.record_revisions(revision_id) ON DELETE CASCADE ON UPDATE NO ACTION;

ALTER TABLE public.record_revisions ADD CONSTRAINT record_revisions_change_set_id_fkey FOREIGN KEY (change_set_id) REFERENCES public.change_sets(change_set_id) ON DELETE CASCADE ON UPDATE NO ACTION;

ALTER TABLE public.record_revisions ADD CONSTRAINT record_revisions_record_id_fkey FOREIGN KEY (record_id) REFERENCES public.records(record_id) ON DELETE CASCADE ON UPDATE NO ACTION;

ALTER SEQUENCE public.record_revisions_revision_id_seq OWNED BY public.record_revisions.revision_id;

REVOKE ALL ON SEQUENCE public.record_revisions_revision_id_seq FROM PUBLIC;

GRANT USAGE ON SEQUENCE public.record_revisions_revision_id_seq TO cartulary_runtime;

GRANT SELECT, UPDATE, USAGE ON SEQUENCE public.record_revisions_revision_id_seq TO cartulary_recovery;

-- +goose Down
-- Disposable harness cleanup only; no production downgrade or restore guarantee.
ALTER TABLE public.record_revisions DROP CONSTRAINT record_revisions_record_id_fkey;
ALTER TABLE public.record_revisions DROP CONSTRAINT record_revisions_change_set_id_fkey;
ALTER TABLE public.record_revision_conflict_facts DROP CONSTRAINT record_revision_conflict_facts_revision_id_fkey;
ALTER TABLE public.record_history_entry_refs DROP CONSTRAINT record_history_entry_refs_record_id_fkey;
ALTER TABLE public.record_history_entry_refs DROP CONSTRAINT record_history_entry_refs_change_set_id_mutation_sequence__fkey;
ALTER TABLE public.record_history_entry_refs DROP CONSTRAINT record_history_entry_refs_change_set_id_fkey;
ALTER TABLE public.change_sets DROP CONSTRAINT change_sets_incident_id_fkey;
ALTER TABLE public.change_sets DROP CONSTRAINT change_sets_actor_user_id_fkey;
ALTER TABLE public.change_set_mutations DROP CONSTRAINT change_set_mutations_change_set_id_fkey;
DROP FUNCTION public.revisions_incident_bundle_sequence_finish_v1(original_next bigint);
DROP FUNCTION public.revisions_incident_bundle_sequence_begin_v1();
DROP TABLE public.record_revisions;
DROP TABLE public.record_revision_conflict_facts;
DROP TABLE public.record_history_entry_refs;
DROP TABLE public.change_sets;
DROP TABLE public.change_set_mutations;
DROP FUNCTION public.change_set_mutations_history_ids_are_canonical(candidate uuid[]);
