-- +goose Up
-- Commit chronology cannot be reconstructed from clocks or UUIDs.
-- +goose StatementBegin
DO $$ BEGIN
    IF EXISTS (SELECT 1 FROM public.change_sets) THEN
        RAISE EXCEPTION 'revision chronology requires fresh history state';
    END IF;
END $$;
-- +goose StatementEnd

CREATE TABLE public.revision_history_orders (
    incident_id uuid NOT NULL,
    last_order bigint NOT NULL,
    CONSTRAINT revision_history_orders_pkey PRIMARY KEY (incident_id),
    CONSTRAINT revision_history_orders_incident_id_fkey FOREIGN KEY (incident_id) REFERENCES public.incidents(id) ON DELETE CASCADE ON UPDATE NO ACTION,
    CONSTRAINT revision_history_orders_positive_ck CHECK (last_order > 0)
);
GRANT INSERT, SELECT, UPDATE, DELETE ON TABLE public.revision_history_orders TO cartulary_runtime;
GRANT INSERT, SELECT, UPDATE, DELETE, TRUNCATE ON TABLE public.revision_history_orders TO cartulary_recovery;

ALTER TABLE public.change_sets
    ADD COLUMN history_order bigint NOT NULL,
    ADD CONSTRAINT change_sets_history_order_positive_ck CHECK (history_order > 0),
    ADD CONSTRAINT change_sets_incident_history_order_key UNIQUE (incident_id, history_order);

-- The upsert holds the incident's allocator row lock until transaction end.
-- Explicit portable orders may arrive in UUID order; never lower the high water.
-- +goose StatementBegin
CREATE FUNCTION public.revisions_assign_history_order_v1() RETURNS trigger
LANGUAGE plpgsql SET search_path = pg_catalog, public AS $$
DECLARE allocated bigint;
BEGIN
    INSERT INTO public.revision_history_orders(incident_id, last_order)
    VALUES (NEW.incident_id, COALESCE(NEW.history_order, 1))
    ON CONFLICT (incident_id) DO UPDATE
      SET last_order = CASE WHEN NEW.history_order IS NULL
        THEN revision_history_orders.last_order + 1
        ELSE GREATEST(revision_history_orders.last_order, NEW.history_order) END
    RETURNING last_order INTO allocated;
    NEW.history_order := COALESCE(NEW.history_order, allocated);
    RETURN NEW;
END $$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION public.revisions_assign_history_order_v1() FROM PUBLIC;
GRANT EXECUTE ON FUNCTION public.revisions_assign_history_order_v1() TO cartulary_runtime;
GRANT EXECUTE ON FUNCTION public.revisions_assign_history_order_v1() TO cartulary_recovery;
CREATE TRIGGER change_sets_assign_history_order
BEFORE INSERT ON public.change_sets FOR EACH ROW
EXECUTE FUNCTION public.revisions_assign_history_order_v1();

-- +goose Down
-- +goose StatementBegin
DO $$ BEGIN
    RAISE EXCEPTION 'revision chronology requires fresh state and has no in-place downgrade';
END $$;
-- +goose StatementEnd
