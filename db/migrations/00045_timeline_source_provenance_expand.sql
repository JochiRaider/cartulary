-- +goose Up
-- +goose StatementBegin
DO $$
BEGIN
    IF EXISTS (
        SELECT 1
          FROM public.timeline_events
         WHERE raw_capture IS NOT NULL
           AND raw_capture <> '{}'::jsonb
           AND (
               jsonb_typeof(raw_capture) <> 'object'
               OR EXISTS (
                   SELECT 1
                     FROM jsonb_object_keys(raw_capture) AS key
                    WHERE key <> 'import_columns'
               )
               OR (
                   raw_capture ? 'import_columns'
                   AND jsonb_typeof(raw_capture -> 'import_columns') <> 'array'
               )
           )
    ) THEN
        RAISE EXCEPTION 'timeline raw_capture contains an unsupported shape';
    END IF;
END
$$;
-- +goose StatementEnd

CREATE TABLE public.timeline_source_provenance (
    record_id uuid NOT NULL REFERENCES public.timeline_events(record_id) ON DELETE CASCADE,
    source_identity_hash bytea NOT NULL,
    source_row_ordinal integer NOT NULL,
    source_column_ordinal integer NOT NULL,
    source_kind text NOT NULL,
    source_metadata jsonb NOT NULL,
    source_header_json jsonb NOT NULL,
    raw_value text NOT NULL,
    cell_kind text,
    created_at timestamp with time zone NOT NULL,
    PRIMARY KEY (
        record_id,
        source_identity_hash,
        source_row_ordinal,
        source_column_ordinal
    ),
    CONSTRAINT timeline_source_provenance_identity_hash_ck
        CHECK (octet_length(source_identity_hash) = 32),
    CONSTRAINT timeline_source_provenance_ordinals_ck
        CHECK (source_row_ordinal >= 0 AND source_column_ordinal >= 0),
    CONSTRAINT timeline_source_provenance_source_kind_ck
        CHECK (octet_length(source_kind) BETWEEN 1 AND 64 AND source_kind !~ '[[:cntrl:]]'),
    CONSTRAINT timeline_source_provenance_metadata_ck
        CHECK (
            jsonb_typeof(source_metadata) = 'object'
            AND octet_length(source_metadata::text) <= 65536
        ),
    CONSTRAINT timeline_source_provenance_header_ck
        CHECK (octet_length(source_header_json::text) <= 65536),
    CONSTRAINT timeline_source_provenance_raw_value_ck
        CHECK (octet_length(raw_value) <= 65536),
    CONSTRAINT timeline_source_provenance_cell_kind_ck
        CHECK (
            cell_kind IS NULL
            OR (
                octet_length(cell_kind) BETWEEN 1 AND 64
                AND cell_kind !~ '[[:cntrl:]]'
            )
        )
);

CREATE INDEX idx_timeline_source_provenance_record
    ON public.timeline_source_provenance (
        record_id,
        source_row_ordinal,
        source_column_ordinal,
        source_identity_hash
    );

WITH provenance_rows AS (
    SELECT
        event.record_id,
        column_value,
        (
            column_value
            - 'source_row_ordinal'
            - 'source_column_ordinal'
            - 'source_header_text'
            - 'raw_value'
            - 'cell_kind'
        ) AS source_metadata,
        event.recorded_at
      FROM public.timeline_events AS event
      CROSS JOIN LATERAL jsonb_array_elements(
          COALESCE(event.raw_capture -> 'import_columns', '[]'::jsonb)
      ) AS column_value
)
INSERT INTO public.timeline_source_provenance (
    record_id,
    source_identity_hash,
    source_row_ordinal,
    source_column_ordinal,
    source_kind,
    source_metadata,
    source_header_json,
    raw_value,
    cell_kind,
    created_at
)
SELECT
    record_id,
    digest(source_metadata::text, 'sha256'),
    (column_value ->> 'source_row_ordinal')::integer,
    (column_value ->> 'source_column_ordinal')::integer,
    column_value ->> 'source_kind',
    source_metadata,
    COALESCE(column_value -> 'source_header_text', 'null'::jsonb),
    COALESCE(column_value ->> 'raw_value', ''),
    NULLIF(column_value ->> 'cell_kind', ''),
    recorded_at
  FROM provenance_rows;

-- +goose StatementBegin
DO $$
BEGIN
    IF EXISTS (
        SELECT record_id
          FROM public.timeline_source_provenance
         GROUP BY record_id
        HAVING count(*) > 4096
    ) THEN
        RAISE EXCEPTION 'timeline source provenance exceeds 4096 entries per record';
    END IF;
END
$$;
-- +goose StatementEnd

-- +goose Down
DROP TABLE public.timeline_source_provenance;
