-- +goose Up
-- Owner-scoped transaction work. Rows must be removed before commit.
-- No runtime DDL, TEMPORARY privilege, retained source state, or backfill.

CREATE TABLE public.entities_bundle_validation_binding (
    operation_id text NOT NULL,
    incident_id uuid NOT NULL,
    actor_id uuid NOT NULL,
    source_digest text NOT NULL,
    transaction_id xid8 DEFAULT pg_current_xact_id() NOT NULL,
    CONSTRAINT entities_bundle_validation_binding_tx_ck CHECK (transaction_id = pg_current_xact_id()),
    CONSTRAINT entities_bundle_validation_binding_pkey PRIMARY KEY (transaction_id)
 );
REVOKE ALL ON TABLE public.entities_bundle_validation_binding FROM PUBLIC;
REVOKE ALL ON TYPE public.entities_bundle_validation_binding FROM PUBLIC;
GRANT SELECT, INSERT, UPDATE, DELETE ON TABLE public.entities_bundle_validation_binding TO cartulary_runtime;
GRANT SELECT, INSERT, UPDATE, DELETE, TRUNCATE ON TABLE public.entities_bundle_validation_binding TO cartulary_recovery;

CREATE TABLE public.entities_bundle_validation_nodes (
    kind text NOT NULL,
    id uuid NOT NULL,
    state text NOT NULL,
    transaction_id xid8 DEFAULT pg_current_xact_id() NOT NULL,
    CONSTRAINT entities_bundle_validation_nodes_tx_ck CHECK (transaction_id = pg_current_xact_id()),
    CONSTRAINT entities_bundle_validation_nodes_pkey PRIMARY KEY (transaction_id, kind, id)
 );
REVOKE ALL ON TABLE public.entities_bundle_validation_nodes FROM PUBLIC;
REVOKE ALL ON TYPE public.entities_bundle_validation_nodes FROM PUBLIC;
GRANT SELECT, INSERT, UPDATE, DELETE ON TABLE public.entities_bundle_validation_nodes TO cartulary_runtime;
GRANT SELECT, INSERT, UPDATE, DELETE, TRUNCATE ON TABLE public.entities_bundle_validation_nodes TO cartulary_recovery;

CREATE TABLE public.entities_bundle_validation_mentions (
    id uuid NOT NULL,
    transaction_id xid8 DEFAULT pg_current_xact_id() NOT NULL,
    CONSTRAINT entities_bundle_validation_mentions_tx_ck CHECK (transaction_id = pg_current_xact_id()),
    CONSTRAINT entities_bundle_validation_mentions_pkey PRIMARY KEY (transaction_id, id)
 );
REVOKE ALL ON TABLE public.entities_bundle_validation_mentions FROM PUBLIC;
REVOKE ALL ON TYPE public.entities_bundle_validation_mentions FROM PUBLIC;
GRANT SELECT, INSERT, UPDATE, DELETE ON TABLE public.entities_bundle_validation_mentions TO cartulary_runtime;
GRANT SELECT, INSERT, UPDATE, DELETE, TRUNCATE ON TABLE public.entities_bundle_validation_mentions TO cartulary_recovery;

CREATE TABLE public.entities_bundle_validation_associations (
    kind text NOT NULL,
    entity_type text NOT NULL,
    record_id uuid NOT NULL,
    key1 text NOT NULL,
    key2 text NOT NULL,
    key3 text NOT NULL,
    validation_row_id uuid DEFAULT gen_random_uuid() NOT NULL,
    transaction_id xid8 DEFAULT pg_current_xact_id() NOT NULL,
    CONSTRAINT entities_bundle_validation_associations_tx_ck CHECK (transaction_id = pg_current_xact_id()),
    CONSTRAINT entities_bundle_validation_associations_pkey PRIMARY KEY (transaction_id, validation_row_id)
 );
REVOKE ALL ON TABLE public.entities_bundle_validation_associations FROM PUBLIC;
REVOKE ALL ON TYPE public.entities_bundle_validation_associations FROM PUBLIC;
GRANT SELECT, INSERT, UPDATE, DELETE ON TABLE public.entities_bundle_validation_associations TO cartulary_runtime;
GRANT SELECT, INSERT, UPDATE, DELETE, TRUNCATE ON TABLE public.entities_bundle_validation_associations TO cartulary_recovery;

CREATE TABLE public.entities_bundle_validation_claims (
    record_id uuid NOT NULL,
    entity_type text NOT NULL,
    identifier_type text NOT NULL,
    normalized_value text NOT NULL,
    transaction_id xid8 DEFAULT pg_current_xact_id() NOT NULL,
    CONSTRAINT entities_bundle_validation_claims_tx_ck CHECK (transaction_id = pg_current_xact_id()),
    CONSTRAINT entities_bundle_validation_claims_pkey PRIMARY KEY (transaction_id, entity_type, identifier_type, normalized_value, record_id)
 );
REVOKE ALL ON TABLE public.entities_bundle_validation_claims FROM PUBLIC;
REVOKE ALL ON TYPE public.entities_bundle_validation_claims FROM PUBLIC;
GRANT SELECT, INSERT, UPDATE, DELETE ON TABLE public.entities_bundle_validation_claims TO cartulary_runtime;
GRANT SELECT, INSERT, UPDATE, DELETE, TRUNCATE ON TABLE public.entities_bundle_validation_claims TO cartulary_recovery;

-- +goose StatementBegin
CREATE FUNCTION public.entities_reject_retained_bundle_validation() RETURNS trigger
LANGUAGE plpgsql SET search_path = pg_catalog, public AS $$
BEGIN
    IF EXISTS (SELECT 1 FROM public.entities_bundle_validation_binding WHERE transaction_id = NEW.transaction_id)
       OR EXISTS (SELECT 1 FROM public.entities_bundle_validation_nodes WHERE transaction_id = NEW.transaction_id)
       OR EXISTS (SELECT 1 FROM public.entities_bundle_validation_mentions WHERE transaction_id = NEW.transaction_id)
       OR EXISTS (SELECT 1 FROM public.entities_bundle_validation_associations WHERE transaction_id = NEW.transaction_id)
       OR EXISTS (SELECT 1 FROM public.entities_bundle_validation_claims WHERE transaction_id = NEW.transaction_id) THEN
        RAISE EXCEPTION 'incident import validation work must be released before commit';
    END IF;
    RETURN NULL;
END;
$$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION public.entities_reject_retained_bundle_validation() FROM PUBLIC;

CREATE CONSTRAINT TRIGGER entities_bundle_validation_binding_empty_commit
AFTER INSERT OR UPDATE ON public.entities_bundle_validation_binding
DEFERRABLE INITIALLY DEFERRED FOR EACH ROW EXECUTE FUNCTION public.entities_reject_retained_bundle_validation();

CREATE CONSTRAINT TRIGGER entities_bundle_validation_nodes_empty_commit
AFTER INSERT OR UPDATE ON public.entities_bundle_validation_nodes
DEFERRABLE INITIALLY DEFERRED FOR EACH ROW EXECUTE FUNCTION public.entities_reject_retained_bundle_validation();

CREATE CONSTRAINT TRIGGER entities_bundle_validation_mentions_empty_commit
AFTER INSERT OR UPDATE ON public.entities_bundle_validation_mentions
DEFERRABLE INITIALLY DEFERRED FOR EACH ROW EXECUTE FUNCTION public.entities_reject_retained_bundle_validation();

CREATE CONSTRAINT TRIGGER entities_bundle_validation_associations_empty_commit
AFTER INSERT OR UPDATE ON public.entities_bundle_validation_associations
DEFERRABLE INITIALLY DEFERRED FOR EACH ROW EXECUTE FUNCTION public.entities_reject_retained_bundle_validation();

CREATE CONSTRAINT TRIGGER entities_bundle_validation_claims_empty_commit
AFTER INSERT OR UPDATE ON public.entities_bundle_validation_claims
DEFERRABLE INITIALLY DEFERRED FOR EACH ROW EXECUTE FUNCTION public.entities_reject_retained_bundle_validation();

-- +goose Down
-- +goose StatementBegin
DO $$ BEGIN
    RAISE EXCEPTION 'incident portability requires fresh state and has no in-place downgrade; retain the prior executable with its own data';
END $$;
-- +goose StatementEnd
