-- +goose Up
-- Owner-scoped transaction work. Rows must be removed before commit.
-- No runtime DDL, TEMPORARY privilege, retained source state, or backfill.

CREATE TABLE public.links_bundle_validation_links (
    record_link_id uuid DEFAULT gen_random_uuid() NOT NULL,
    incident_id uuid NOT NULL,
    src_record_id uuid NOT NULL,
    dst_record_id uuid NOT NULL,
    link_type text NOT NULL,
    provenance text NOT NULL,
    confidence integer,
    owner_user_id uuid NOT NULL,
    decided_at timestamp with time zone DEFAULT now() NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone,
    deleted_by_user_id uuid,
    created_by_user_id uuid NOT NULL,
    field_key text,
    transaction_id xid8 DEFAULT pg_current_xact_id() NOT NULL,
    CONSTRAINT links_bundle_validation_links_tx_ck CHECK (transaction_id = pg_current_xact_id()),
    CONSTRAINT links_bundle_validation_links_pkey PRIMARY KEY (transaction_id, record_link_id)
 );
REVOKE ALL ON TABLE public.links_bundle_validation_links FROM PUBLIC;
REVOKE ALL ON TYPE public.links_bundle_validation_links FROM PUBLIC;
GRANT SELECT, INSERT, UPDATE, DELETE ON TABLE public.links_bundle_validation_links TO cartulary_runtime;
GRANT SELECT, INSERT, UPDATE, DELETE, TRUNCATE ON TABLE public.links_bundle_validation_links TO cartulary_recovery;

CREATE TABLE public.links_bundle_validation_tags (
    record_tag_id uuid DEFAULT gen_random_uuid() NOT NULL,
    incident_id uuid NOT NULL,
    record_id uuid NOT NULL,
    tag_name text NOT NULL,
    normalized_tag_name text NOT NULL,
    created_by_user_id uuid NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone,
    deleted_by_user_id uuid,
    transaction_id xid8 DEFAULT pg_current_xact_id() NOT NULL,
    CONSTRAINT links_bundle_validation_tags_tx_ck CHECK (transaction_id = pg_current_xact_id()),
    CONSTRAINT links_bundle_validation_tags_pkey PRIMARY KEY (transaction_id, record_tag_id)
 );
REVOKE ALL ON TABLE public.links_bundle_validation_tags FROM PUBLIC;
REVOKE ALL ON TYPE public.links_bundle_validation_tags FROM PUBLIC;
GRANT SELECT, INSERT, UPDATE, DELETE ON TABLE public.links_bundle_validation_tags TO cartulary_runtime;
GRANT SELECT, INSERT, UPDATE, DELETE, TRUNCATE ON TABLE public.links_bundle_validation_tags TO cartulary_recovery;

CREATE TABLE public.links_bundle_validation_catalog (
    tag_name text COLLATE "C" NOT NULL,
    normalized_tag_name text COLLATE "C" NOT NULL,
    transaction_id xid8 DEFAULT pg_current_xact_id() NOT NULL,
    CONSTRAINT links_bundle_validation_catalog_tx_ck CHECK (transaction_id = pg_current_xact_id()),
    CONSTRAINT links_bundle_validation_catalog_pkey PRIMARY KEY (transaction_id, tag_name, normalized_tag_name)
 );
REVOKE ALL ON TABLE public.links_bundle_validation_catalog FROM PUBLIC;
REVOKE ALL ON TYPE public.links_bundle_validation_catalog FROM PUBLIC;
GRANT SELECT, INSERT, UPDATE, DELETE ON TABLE public.links_bundle_validation_catalog TO cartulary_runtime;
GRANT SELECT, INSERT, UPDATE, DELETE, TRUNCATE ON TABLE public.links_bundle_validation_catalog TO cartulary_recovery;

-- +goose StatementBegin
CREATE FUNCTION public.links_reject_retained_bundle_validation() RETURNS trigger
LANGUAGE plpgsql SET search_path = pg_catalog, public AS $$
BEGIN
    IF EXISTS (SELECT 1 FROM public.links_bundle_validation_links WHERE transaction_id = NEW.transaction_id)
       OR EXISTS (SELECT 1 FROM public.links_bundle_validation_tags WHERE transaction_id = NEW.transaction_id)
       OR EXISTS (SELECT 1 FROM public.links_bundle_validation_catalog WHERE transaction_id = NEW.transaction_id) THEN
        RAISE EXCEPTION 'incident import validation work must be released before commit';
    END IF;
    RETURN NULL;
END;
$$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION public.links_reject_retained_bundle_validation() FROM PUBLIC;

CREATE CONSTRAINT TRIGGER links_bundle_validation_links_empty_commit
AFTER INSERT OR UPDATE ON public.links_bundle_validation_links
DEFERRABLE INITIALLY DEFERRED FOR EACH ROW EXECUTE FUNCTION public.links_reject_retained_bundle_validation();

CREATE CONSTRAINT TRIGGER links_bundle_validation_tags_empty_commit
AFTER INSERT OR UPDATE ON public.links_bundle_validation_tags
DEFERRABLE INITIALLY DEFERRED FOR EACH ROW EXECUTE FUNCTION public.links_reject_retained_bundle_validation();

CREATE CONSTRAINT TRIGGER links_bundle_validation_catalog_empty_commit
AFTER INSERT OR UPDATE ON public.links_bundle_validation_catalog
DEFERRABLE INITIALLY DEFERRED FOR EACH ROW EXECUTE FUNCTION public.links_reject_retained_bundle_validation();

-- +goose Down
-- +goose StatementBegin
DO $$ BEGIN
    RAISE EXCEPTION 'incident portability requires fresh state and has no in-place downgrade; retain the prior executable with its own data';
END $$;
-- +goose StatementEnd
