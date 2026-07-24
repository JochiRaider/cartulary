-- +goose Up
--
-- Name: deterministic Import unit discovery order; Type: COLUMN/CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE public.import_units
    ADD COLUMN discovery_sequence integer;

WITH ranked AS (
    SELECT import_unit_id,
           row_number() OVER (
               PARTITION BY import_session_id
               ORDER BY created_at ASC, import_unit_id ASC
           ) AS discovery_sequence
      FROM public.import_units
)
UPDATE public.import_units AS unit
   SET discovery_sequence = ranked.discovery_sequence
  FROM ranked
 WHERE ranked.import_unit_id = unit.import_unit_id;

ALTER TABLE public.import_units
    ALTER COLUMN discovery_sequence SET NOT NULL,
    ADD CONSTRAINT import_units_discovery_sequence_ck
        CHECK (discovery_sequence BETWEEN 1 AND 2147483647),
    ADD CONSTRAINT import_units_session_discovery_sequence_uq
        UNIQUE (import_session_id, discovery_sequence);

-- +goose Down
ALTER TABLE public.import_units
    DROP CONSTRAINT import_units_session_discovery_sequence_uq,
    DROP CONSTRAINT import_units_discovery_sequence_ck,
    DROP COLUMN discovery_sequence;
