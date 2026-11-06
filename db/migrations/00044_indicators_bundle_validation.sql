-- +goose Up
-- Owner-scoped transaction work. Rows must be removed before commit.
-- No runtime DDL, TEMPORARY privilege, retained source state, or backfill.

CREATE TABLE public.indicators_bundle_validation_sources (
    id uuid NOT NULL,
    kind text NOT NULL,
    dedupe_key text NOT NULL,
    active boolean NOT NULL,
    transaction_id xid8 DEFAULT pg_current_xact_id() NOT NULL,
    CONSTRAINT indicators_bundle_validation_sources_tx_ck CHECK (transaction_id = pg_current_xact_id()),
    CONSTRAINT indicators_bundle_validation_sources_pkey PRIMARY KEY (transaction_id, id)
 );
REVOKE ALL ON TABLE public.indicators_bundle_validation_sources FROM PUBLIC;
REVOKE ALL ON TYPE public.indicators_bundle_validation_sources FROM PUBLIC;
GRANT SELECT, INSERT, UPDATE, DELETE ON TABLE public.indicators_bundle_validation_sources TO cartulary_runtime;
GRANT SELECT, INSERT, UPDATE, DELETE, TRUNCATE ON TABLE public.indicators_bundle_validation_sources TO cartulary_recovery;

-- +goose StatementBegin
CREATE FUNCTION public.indicators_reject_retained_bundle_validation() RETURNS trigger
LANGUAGE plpgsql SET search_path = pg_catalog, public AS $$
BEGIN
    IF EXISTS (SELECT 1 FROM public.indicators_bundle_validation_sources WHERE transaction_id = NEW.transaction_id) THEN
        RAISE EXCEPTION 'incident import validation work must be released before commit';
    END IF;
    RETURN NULL;
END;
$$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION public.indicators_reject_retained_bundle_validation() FROM PUBLIC;

CREATE CONSTRAINT TRIGGER indicators_bundle_validation_sources_empty_commit
AFTER INSERT OR UPDATE ON public.indicators_bundle_validation_sources
DEFERRABLE INITIALLY DEFERRED FOR EACH ROW EXECUTE FUNCTION public.indicators_reject_retained_bundle_validation();

-- +goose Down
-- +goose StatementBegin
DO $$ BEGIN
    RAISE EXCEPTION 'incident portability requires fresh state and has no in-place downgrade; retain the prior executable with its own data';
END $$;
-- +goose StatementEnd
