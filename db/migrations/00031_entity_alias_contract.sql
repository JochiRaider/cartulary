-- +goose Up
-- +goose StatementBegin
DO $$
DECLARE
    invalid_count bigint;
    invalid_sample text;
BEGIN
    SELECT count(*)
      INTO invalid_count
      FROM entity_aliases
     WHERE normalized_text = ''
        OR char_length(normalized_text) > 256
        OR normalized_text ~ '[[:cntrl:]]';
    SELECT string_agg(entity_alias_id::text, ', ' ORDER BY entity_alias_id)
      INTO invalid_sample
      FROM (
          SELECT entity_alias_id
            FROM entity_aliases
           WHERE normalized_text = ''
              OR char_length(normalized_text) > 256
              OR normalized_text ~ '[[:cntrl:]]'
           ORDER BY entity_alias_id
           LIMIT 10
      ) invalid;
    IF invalid_count > 0 THEN
        RAISE EXCEPTION 'entity alias migration preflight failed: invalid_count=%, sample_entity_alias_ids=%', invalid_count, invalid_sample;
    END IF;
END
$$;
-- +goose StatementEnd

DROP INDEX entity_aliases_record_unique_idx;

ALTER TABLE entity_aliases
    ALTER COLUMN normalized_text TYPE citext USING normalized_text::citext;

UPDATE entity_aliases
   SET raw_text = normalized_text::text
 WHERE raw_text IS DISTINCT FROM normalized_text::text;

WITH ranked AS (
    SELECT entity_alias_id,
           row_number() OVER (
               PARTITION BY record_id, entity_type, normalized_text
               ORDER BY created_at ASC, entity_alias_id ASC
           ) AS duplicate_rank
      FROM entity_aliases
     WHERE deleted_at IS NULL
)
UPDATE entity_aliases aliases
   SET deleted_at = statement_timestamp()
  FROM ranked
 WHERE aliases.entity_alias_id = ranked.entity_alias_id
   AND ranked.duplicate_rank > 1;

ALTER TABLE entity_aliases
    ADD CONSTRAINT entity_aliases_alias_text_nonempty_ck CHECK (normalized_text::text <> ''),
    ADD CONSTRAINT entity_aliases_alias_text_length_ck CHECK (char_length(normalized_text::text) <= 256),
    ADD CONSTRAINT entity_aliases_alias_text_controls_ck CHECK (normalized_text::text !~ '[[:cntrl:]]'),
    ADD CONSTRAINT entity_aliases_alias_text_storage_ck CHECK (raw_text = normalized_text::text);

CREATE UNIQUE INDEX entity_aliases_record_unique_idx
    ON entity_aliases (record_id, entity_type, normalized_text)
    WHERE deleted_at IS NULL;

-- +goose Down
DROP INDEX entity_aliases_record_unique_idx;

ALTER TABLE entity_aliases
    DROP CONSTRAINT entity_aliases_alias_text_storage_ck,
    DROP CONSTRAINT entity_aliases_alias_text_controls_ck,
    DROP CONSTRAINT entity_aliases_alias_text_length_ck,
    DROP CONSTRAINT entity_aliases_alias_text_nonempty_ck,
    ALTER COLUMN normalized_text TYPE text USING normalized_text::text;

CREATE UNIQUE INDEX entity_aliases_record_unique_idx
    ON entity_aliases (record_id, entity_type, normalized_text)
    WHERE deleted_at IS NULL;
