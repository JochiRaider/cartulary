-- +goose Up
-- Serialize the audit, source correction, and projection refresh with writers.
LOCK TABLE public.records, public.record_tags, public.timeline_grid_projection
    IN SHARE ROW EXCLUSIVE MODE;

-- Include tag tombstones and tags on deleted records. Both the upgrade and
-- the subsequent logical-restore check use this Links-owned read relation.
CREATE VIEW public.links_timeline_tag_labels_v1 WITH (security_invoker = true) AS
SELECT t.record_tag_id, t.tag_name, t.normalized_tag_name, t.deleted_at
FROM public.record_tags t
JOIN public.records r USING (incident_id, record_id)
WHERE r.record_type = 'timeline_event';
REVOKE ALL ON TABLE public.links_timeline_tag_labels_v1 FROM PUBLIC;
REVOKE ALL ON TYPE public.links_timeline_tag_labels_v1 FROM PUBLIC;
GRANT SELECT ON TABLE public.links_timeline_tag_labels_v1 TO cartulary_runtime;
GRANT SELECT ON TABLE public.links_timeline_tag_labels_v1 TO cartulary_recovery;

-- PostgreSQL 18 uses Unicode 16; the writing application uses Unicode 17.
-- UnicodeData 17 adds combining classes for these marks but no canonical
-- decomposition or composition pairs. For those inputs, let the database
-- order and compose with inert stand-ins of the same class. Stable ordering
-- of equal classes lets us restore even literal stand-ins without ambiguity.
-- This migration-local function is dropped before commit.
-- +goose StatementBegin
CREATE FUNCTION public.links_v30_tag_nfc(raw text) RETURNS text
LANGUAGE plpgsql IMMUTABLE STRICT SET search_path = pg_catalog, public AS $nfc$
DECLARE
    class_230 constant text := U&'\1ACF\1AD0\1AD1\1AD2\1AD3\1AD4\1AD5\1AD6\1AD7\1AD8\1AD9\1ADA\1ADB\1ADC\1AE0\1AE1\1AE2\1AE3\1AE4\1AE5\1AE7\1AE8\1AE9\1AEA\+01E6E3\+01E6E6\+01E6EE\+01E6EF\+01E6F5';
    class_220 constant text := U&'\1ADD\1AE6\+010EFA\+010EFB';
    class_234 constant text := U&'\1AEB';
    marks constant text := class_230 || class_220 || class_234;
    proxies constant text := U&'\035B\0359\035D';
    -- CCC-zero, NFKD-irreducible NFC_QC=Maybe characters from Unicode 17
    -- DerivedNormalizationProps. The Go writer counts backward-combining
    -- characters (including Hangul V and T) as nonstarters as well.
    backward_pattern constant text := U&'[\09BE\09D7\0B3E\0B56-\0B57\0BBE\0BD7\0CC2\0CD5-\0CD6\0D3E\0D57\0DCF\0DDF\102E\1161-\1175\11A8-\11C2\1B35\+011127\+01133E\+011357\+0113B8\+0113BB\+0113C2\+0113C9\+0114B0\+0114BA\+0114BD\+0115AF\+011930\+01611E-\+016120\+016129\+016D67]';
    compatibility text;
    nonstarters integer := 0;
    probe text;
    decomposed text;
    shadow text;
    originals text[];
    pos integer[] := ARRAY[0, 0, 0];
    result text := '';
    ch text;
    slot integer;
BEGIN
    -- A valid 64-scalar NFC label expands to at most 256 NFD scalars.
    IF char_length(raw) > 256 THEN
        RAISE EXCEPTION 'Timeline tag audit found an oversized label outside the NFC repair';
    END IF;
    -- Go NFC would insert a combining grapheme joiner after 30 consecutive
    -- nonstarters. Inserting a new scalar is outside a composition-only repair,
    -- even when a corrupt comparison key agrees with the unguarded spelling.
    IF octet_length(raw) <> char_length(raw) THEN
        compatibility := normalize(raw, NFKD);
        IF char_length(compatibility) > 30 THEN
            FOREACH ch IN ARRAY string_to_array(compatibility, NULL) LOOP
                -- On an NFKD scalar, a starter separates these two sentinels.
                -- A nonzero combining class lets NFD reorder class 1 ahead of 240.
                probe := U&'\0345' || ch || U&'\0334';
                IF normalize(probe, NFD) COLLATE "C" <> probe COLLATE "C"
                    OR strpos(marks, ch) > 0 OR ch COLLATE "C" ~ backward_pattern THEN
                    nonstarters := nonstarters + 1;
                    IF nonstarters > 30 THEN
                        RAISE EXCEPTION 'Timeline tag audit found a non-stream-safe label outside the NFC repair';
                    END IF;
                ELSE
                    nonstarters := 0;
                END IF;
            END LOOP;
        END IF;
    END IF;
    IF raw COLLATE "C" !~ ('[' || marks || ']') THEN
        RETURN normalize(raw, NFC);
    END IF;
    decomposed := normalize(raw, NFD);
    originals := ARRAY[
        regexp_replace(decomposed COLLATE "C", '[^' || class_230 || U&'\035B]', '', 'g'),
        regexp_replace(decomposed COLLATE "C", '[^' || class_220 || U&'\0359]', '', 'g'),
        regexp_replace(decomposed COLLATE "C", '[^' || class_234 || U&'\035D]', '', 'g')
    ];
    shadow := normalize(translate(decomposed, marks,
        repeat(U&'\035B', length(class_230)) || repeat(U&'\0359', length(class_220)) || U&'\035D'), NFC);
    FOREACH ch IN ARRAY string_to_array(shadow, NULL) LOOP
        slot := strpos(proxies, ch);
        IF slot > 0 THEN
            pos[slot] := pos[slot] + 1;
            ch := substr(originals[slot], pos[slot], 1);
        END IF;
        result := result || ch;
    END LOOP;
    RETURN result;
END
$nfc$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION public.links_v30_tag_nfc(text) FROM PUBLIC;

-- +goose StatementBegin
DO $repair$
DECLARE
    -- Additional full folds in CaseFolding 17: three Latin letters and
    -- Beria Erfe capitals U+16EA0..16EB8, whose lowercase offset is 27.
    capitals text;
    lowercase text;
    invalid_count bigint;
    active_count bigint;
    deleted_count bigint;
BEGIN
    SELECT U&'\A7CE\A7D2\A7D4' || string_agg(chr(cp), '' ORDER BY cp),
           U&'\A7CF\A7D3\A7D5' || string_agg(chr(cp + 27), '' ORDER BY cp)
    INTO capitals, lowercase FROM generate_series(0x16EA0, 0x16EB8) cp;

    WITH labels AS MATERIALIZED (
        SELECT t.tag_name, t.normalized_tag_name, t.deleted_at,
               public.links_v30_tag_nfc(t.tag_name) COLLATE "C" AS label
        FROM public.links_timeline_tag_labels_v1 t
    )
    SELECT count(*) FILTER (WHERE
               char_length(label) NOT BETWEEN 1 AND 64
               OR label ~ U&'[\0001-\001F\007F-\009F]'
               OR label <> btrim(label, U&'\0020\00A0\1680\2000\2001\2002\2003\2004\2005\2006\2007\2008\2009\200A\2028\2029\202F\205F\3000')
               OR normalized_tag_name COLLATE "C" <> casefold(translate(label, capitals, lowercase) COLLATE pg_catalog.pg_unicode_fast) COLLATE "C"),
           count(*) FILTER (WHERE tag_name COLLATE "C" <> label AND deleted_at IS NULL),
           count(*) FILTER (WHERE tag_name COLLATE "C" <> label AND deleted_at IS NOT NULL)
    INTO invalid_count, active_count, deleted_count FROM labels;

    IF invalid_count <> 0 THEN
        RAISE EXCEPTION 'Timeline tag audit found % invalid bindings outside the NFC repair', invalid_count;
    END IF;
    RAISE NOTICE 'Timeline NFC tag repair: % active and % tombstoned bindings', active_count, deleted_count;
END
$repair$;
-- +goose StatementEnd

-- A data-modifying CTE and its outer query share a snapshot. Normalize the
-- source label in the projection expression as well. Its order matches the
-- Timeline facts reader (normalized key in byte order, then tag identity).
WITH repaired AS (
    UPDATE public.record_tags t SET tag_name = public.links_v30_tag_nfc(t.tag_name)
    FROM public.links_timeline_tag_labels_v1 v
    WHERE v.record_tag_id = t.record_tag_id
      AND v.tag_name COLLATE "C" <> public.links_v30_tag_nfc(v.tag_name) COLLATE "C"
    RETURNING t.incident_id, t.record_id
)
UPDATE public.timeline_grid_projection p
SET tags = (
    SELECT coalesce(jsonb_agg(jsonb_build_object(
               'item_ref', 'record_tag:' || t.record_id::text || ':' || t.record_tag_id::text,
               'item_kind', 'tag', 'tag_id', t.record_tag_id,
               'display_text', public.links_v30_tag_nfc(t.tag_name)
           ) ORDER BY t.normalized_tag_name COLLATE "C", t.record_tag_id), '[]'::jsonb)
    FROM public.active_record_tags_v1 t
    WHERE (t.incident_id, t.record_id) = (p.incident_id, p.record_id)
)
WHERE (p.incident_id, p.record_id) IN (SELECT incident_id, record_id FROM repaired);

DROP FUNCTION public.links_v30_tag_nfc(text);

-- +goose Down
-- The previous display bytes are intentionally not retained by the migration.
-- +goose StatementBegin
DO $repair$
BEGIN
    RAISE EXCEPTION 'Timeline tag NFC correction is irreversible; restore the pre-upgrade backup';
END
$repair$;
-- +goose StatementEnd
