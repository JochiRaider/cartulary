-- +goose Up
-- +goose StatementBegin
DO $$
DECLARE
    duplicate_association_count bigint;
BEGIN
    SELECT count(*)
      INTO duplicate_association_count
      FROM (
        SELECT object_blob_id
          FROM public.evidence
         WHERE object_blob_id IS NOT NULL
         GROUP BY object_blob_id
        HAVING count(*) > 1
      ) duplicates;

    IF duplicate_association_count > 0 THEN
        RAISE EXCEPTION USING
            ERRCODE = '23505',
            MESSAGE = format(
                'evidence blob association uniqueness preflight failed: duplicate_association_count=%s',
                duplicate_association_count
            );
    END IF;
END
$$;
-- +goose StatementEnd

DROP INDEX public.evidence_object_blob_idx;

CREATE UNIQUE INDEX evidence_object_blob_unique_idx
    ON public.evidence USING btree (object_blob_id)
    WHERE object_blob_id IS NOT NULL;

-- +goose Down
DROP INDEX public.evidence_object_blob_unique_idx;

CREATE INDEX evidence_object_blob_idx
    ON public.evidence USING btree (object_blob_id)
    WHERE object_blob_id IS NOT NULL;
