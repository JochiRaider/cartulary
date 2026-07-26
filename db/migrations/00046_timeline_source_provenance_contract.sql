-- +goose Up
ALTER TABLE public.timeline_events
    DROP COLUMN raw_capture;

-- +goose Down
ALTER TABLE public.timeline_events
    ADD COLUMN raw_capture jsonb DEFAULT '{}'::jsonb NOT NULL;

UPDATE public.timeline_events AS event
   SET raw_capture = jsonb_build_object(
       'import_columns',
       COALESCE((
           SELECT jsonb_agg(
               provenance.source_metadata
               || jsonb_build_object(
                   'source_row_ordinal', provenance.source_row_ordinal,
                   'source_column_ordinal', provenance.source_column_ordinal,
                   'source_header_text', provenance.source_header_json,
                   'raw_value', provenance.raw_value
               )
               || CASE
                   WHEN provenance.cell_kind IS NULL THEN '{}'::jsonb
                   ELSE jsonb_build_object('cell_kind', provenance.cell_kind)
               END
               ORDER BY
                   provenance.source_row_ordinal,
                   provenance.source_column_ordinal,
                   provenance.source_identity_hash
           )
             FROM public.timeline_source_provenance AS provenance
            WHERE provenance.record_id = event.record_id
       ), '[]'::jsonb)
   )
 WHERE EXISTS (
       SELECT 1
         FROM public.timeline_source_provenance AS provenance
        WHERE provenance.record_id = event.record_id
   );
