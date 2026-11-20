-- +goose Up
-- Native history order is allocated when the writing transaction commits, so
-- order follows commit order and no writer holds the allocator during owner
-- work. Committed rows already carry their orders; no data is transformed.
DROP TRIGGER change_sets_assign_history_order ON public.change_sets;
DROP FUNCTION public.revisions_assign_history_order_v1();

-- PostgreSQL cannot defer NOT NULL. The commit-time allocator gives every
-- committed change set an order; the positive check and incident key remain.
ALTER TABLE public.change_sets ALTER COLUMN history_order DROP NOT NULL;

-- One deferred event per inserted change set. The allocator upsert takes the
-- incident's row lock at commit and holds it until transaction end. A row
-- removed earlier in the transaction allocates nothing. An explicit portable
-- order only raises the high water. Exhaustion fails the commit; never wrap.
-- +goose StatementBegin
CREATE FUNCTION public.revisions_allocate_history_order_v2() RETURNS trigger
LANGUAGE plpgsql SET search_path = pg_catalog, public AS $$
DECLARE
    assigned bigint;
    allocated bigint;
BEGIN
    SELECT history_order INTO assigned
      FROM public.change_sets
     WHERE change_set_id = NEW.change_set_id;
    IF NOT FOUND THEN
        RETURN NULL;
    END IF;
    IF assigned IS NOT NULL THEN
        INSERT INTO public.revision_history_orders AS allocator (incident_id, last_order)
        VALUES (NEW.incident_id, assigned)
        ON CONFLICT (incident_id) DO UPDATE
          SET last_order = GREATEST(allocator.last_order, EXCLUDED.last_order);
        RETURN NULL;
    END IF;
    INSERT INTO public.revision_history_orders AS allocator (incident_id, last_order)
    VALUES (NEW.incident_id, 1)
    ON CONFLICT (incident_id) DO UPDATE
      SET last_order = allocator.last_order + 1
      WHERE allocator.last_order < 9223372036854775807
    RETURNING allocator.last_order INTO allocated;
    IF allocated IS NULL THEN
        RAISE EXCEPTION 'revision history order is exhausted'
            USING ERRCODE = 'numeric_value_out_of_range';
    END IF;
    UPDATE public.change_sets
       SET history_order = allocated
     WHERE change_set_id = NEW.change_set_id;
    RETURN NULL;
END $$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION public.revisions_allocate_history_order_v2() FROM PUBLIC;
CREATE CONSTRAINT TRIGGER change_sets_allocate_history_order
AFTER INSERT ON public.change_sets
DEFERRABLE INITIALLY DEFERRED FOR EACH ROW
EXECUTE FUNCTION public.revisions_allocate_history_order_v2();

-- +goose Down
-- The predecessor allocator held the incident lock across owner work, and its
-- chronology migration has no in-place downgrade either.
-- +goose StatementBegin
DO $$ BEGIN
    RAISE EXCEPTION 'commit-order revision chronology has no in-place downgrade; retain the prior executable with its own data';
END $$;
-- +goose StatementEnd
