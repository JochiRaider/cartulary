-- +goose Up
-- A row belongs to the Network Flow table of its incident and creator. A
-- per-row foreign key checked that once per row, which a default-limit import
-- pays 250,000 times inside its participant budget (NF-REQ-171a). Statement
-- triggers enforce the same relation once per statement over the statement's
-- transition rows; rows stay immutable, so inserts, parent deletes, parent key
-- updates and parent truncation are the only ways to break it.
ALTER TABLE public.network_flow_rows DROP CONSTRAINT network_flow_rows_table_creator_fkey;
-- This index served only the key's checks and cascade; the unique
-- (network_flow_table_id, source_row_number) index serves the cascade.
DROP INDEX public.network_flow_rows_table_creator_scope_idx;

-- Inserted rows must name an existing table of the same incident and creator.
-- The referenced tables are share-locked, as a foreign key locks them, so a
-- concurrent delete or key update waits for this transaction.
-- +goose StatementBegin
CREATE FUNCTION public.network_flow_rows_require_table()
 RETURNS trigger
 LANGUAGE plpgsql
 SET search_path = pg_catalog, public
AS $$
BEGIN
    PERFORM 1 FROM public.network_flow_tables t
     WHERE (t.incident_id, t.network_flow_table_id, t.created_by_user_id) IN
           (SELECT DISTINCT r.incident_id, r.network_flow_table_id, r.created_by_user_id FROM inserted_rows r)
       FOR KEY SHARE OF t;
    IF EXISTS (
        SELECT 1 FROM (SELECT DISTINCT incident_id, network_flow_table_id, created_by_user_id FROM inserted_rows) r
         WHERE NOT EXISTS (
            SELECT 1 FROM public.network_flow_tables t
             WHERE t.incident_id = r.incident_id
               AND t.network_flow_table_id = r.network_flow_table_id
               AND t.created_by_user_id = r.created_by_user_id)) THEN
        RAISE EXCEPTION 'Network Flow rows must belong to a table of their incident and creator'
            USING ERRCODE = '23503', CONSTRAINT = 'network_flow_rows_table_key',
                  SCHEMA = 'public', TABLE = 'network_flow_rows';
    END IF;
    RETURN NULL;
END;
$$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION public.network_flow_rows_require_table() FROM PUBLIC;

GRANT EXECUTE ON FUNCTION public.network_flow_rows_require_table() TO cartulary_runtime;

GRANT EXECUTE ON FUNCTION public.network_flow_rows_require_table() TO cartulary_recovery;

CREATE TRIGGER network_flow_rows_table_key AFTER INSERT ON public.network_flow_rows
    REFERENCING NEW TABLE AS inserted_rows
    FOR EACH STATEMENT EXECUTE FUNCTION public.network_flow_rows_require_table();

-- Deleting a table deletes its rows, as the key's ON DELETE CASCADE did.
-- +goose StatementBegin
CREATE FUNCTION public.network_flow_tables_delete_rows()
 RETURNS trigger
 LANGUAGE plpgsql
 SET search_path = pg_catalog, public
AS $$
BEGIN
    DELETE FROM public.network_flow_rows r
     USING deleted_tables d
     WHERE r.network_flow_table_id = d.network_flow_table_id
       AND r.incident_id = d.incident_id
       AND r.created_by_user_id = d.created_by_user_id;
    RETURN NULL;
END;
$$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION public.network_flow_tables_delete_rows() FROM PUBLIC;

GRANT EXECUTE ON FUNCTION public.network_flow_tables_delete_rows() TO cartulary_runtime;

GRANT EXECUTE ON FUNCTION public.network_flow_tables_delete_rows() TO cartulary_recovery;

CREATE TRIGGER network_flow_tables_delete_rows AFTER DELETE ON public.network_flow_tables
    REFERENCING OLD TABLE AS deleted_tables
    FOR EACH STATEMENT EXECUTE FUNCTION public.network_flow_tables_delete_rows();

-- A table's key cannot change while rows name it, and truncating the tables
-- cannot leave rows behind; both match the key's NO ACTION behavior.
-- +goose StatementBegin
CREATE FUNCTION public.network_flow_tables_keep_row_key()
 RETURNS trigger
 LANGUAGE plpgsql
 SET search_path = pg_catalog, public
AS $$
BEGIN
    IF TG_OP = 'TRUNCATE' THEN
        IF EXISTS (SELECT 1 FROM public.network_flow_rows) THEN
            RAISE EXCEPTION 'Network Flow rows must belong to a table of their incident and creator'
                USING ERRCODE = '23503', CONSTRAINT = 'network_flow_rows_table_key',
                      SCHEMA = 'public', TABLE = 'network_flow_rows';
        END IF;
        RETURN NULL;
    END IF;
    IF (NEW.incident_id, NEW.network_flow_table_id, NEW.created_by_user_id)
           IS DISTINCT FROM (OLD.incident_id, OLD.network_flow_table_id, OLD.created_by_user_id)
       AND EXISTS (
           SELECT 1 FROM public.network_flow_rows r
            WHERE r.network_flow_table_id = OLD.network_flow_table_id
              AND r.incident_id = OLD.incident_id
              AND r.created_by_user_id = OLD.created_by_user_id) THEN
        RAISE EXCEPTION 'Network Flow rows must belong to a table of their incident and creator'
            USING ERRCODE = '23503', CONSTRAINT = 'network_flow_rows_table_key',
                  SCHEMA = 'public', TABLE = 'network_flow_tables';
    END IF;
    RETURN NEW;
END;
$$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION public.network_flow_tables_keep_row_key() FROM PUBLIC;

GRANT EXECUTE ON FUNCTION public.network_flow_tables_keep_row_key() TO cartulary_runtime;

GRANT EXECUTE ON FUNCTION public.network_flow_tables_keep_row_key() TO cartulary_recovery;

CREATE TRIGGER network_flow_tables_keep_row_key BEFORE UPDATE OF incident_id, network_flow_table_id, created_by_user_id
    ON public.network_flow_tables
    FOR EACH ROW EXECUTE FUNCTION public.network_flow_tables_keep_row_key();

CREATE TRIGGER network_flow_tables_keep_row_key_on_truncate AFTER TRUNCATE ON public.network_flow_tables
    FOR EACH STATEMENT EXECUTE FUNCTION public.network_flow_tables_keep_row_key();

-- A logical restore loads rows with triggers disabled; Recovery proves the
-- relation afterwards with this owner predicate.
-- +goose StatementBegin
CREATE FUNCTION public.network_flow_rows_belong_to_tables_v1()
 RETURNS boolean
 LANGUAGE sql
 STABLE
 SET search_path = pg_catalog, public
AS $$
    SELECT NOT EXISTS (
        SELECT 1 FROM public.network_flow_rows r
         WHERE NOT EXISTS (
            SELECT 1 FROM public.network_flow_tables t
             WHERE t.incident_id = r.incident_id
               AND t.network_flow_table_id = r.network_flow_table_id
               AND t.created_by_user_id = r.created_by_user_id))
$$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION public.network_flow_rows_belong_to_tables_v1() FROM PUBLIC;

GRANT EXECUTE ON FUNCTION public.network_flow_rows_belong_to_tables_v1() TO cartulary_recovery;

-- +goose Down
DROP FUNCTION public.network_flow_rows_belong_to_tables_v1();
DROP TRIGGER network_flow_tables_keep_row_key_on_truncate ON public.network_flow_tables;
DROP TRIGGER network_flow_tables_keep_row_key ON public.network_flow_tables;
DROP FUNCTION public.network_flow_tables_keep_row_key();
DROP TRIGGER network_flow_tables_delete_rows ON public.network_flow_tables;
DROP FUNCTION public.network_flow_tables_delete_rows();
DROP TRIGGER network_flow_rows_table_key ON public.network_flow_rows;
DROP FUNCTION public.network_flow_rows_require_table();
CREATE INDEX network_flow_rows_table_creator_scope_idx ON public.network_flow_rows USING btree (incident_id, network_flow_table_id, created_by_user_id);
ALTER TABLE public.network_flow_rows ADD CONSTRAINT network_flow_rows_table_creator_fkey FOREIGN KEY (incident_id, network_flow_table_id, created_by_user_id) REFERENCES public.network_flow_tables(incident_id, network_flow_table_id, created_by_user_id) ON DELETE CASCADE ON UPDATE NO ACTION;
