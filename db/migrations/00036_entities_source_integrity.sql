-- +goose Up
--
-- Entities source rows are authoritative inputs, not best-effort projections.
-- Classify the complete existing source set before creating enforcement
-- objects. Failure details contain aggregate counts only; operators choose the
-- disposition and the migration never repairs, merges, or discards source
-- state.

-- +goose StatementBegin
DO $$
DECLARE
    host_count bigint;
    identity_count bigint;
    alias_count bigint := 0;
    preserved_count bigint := 0;
    mention_count bigint;
    source_row record;
    candidate text;
    codepoint integer;
    character_ordinal integer;
    invalid boolean;
    unicode_space text := chr(9) || chr(10) || chr(11) || chr(12) || chr(13)
        || chr(32) || chr(133) || chr(160) || chr(5760)
        || chr(8192) || chr(8193) || chr(8194) || chr(8195) || chr(8196)
        || chr(8197) || chr(8198) || chr(8199) || chr(8200) || chr(8201)
        || chr(8202) || chr(8232) || chr(8233) || chr(8239) || chr(8287)
        || chr(12288);
BEGIN
    SELECT count(*)
      INTO host_count
      FROM public.hosts h
      LEFT JOIN public.records r
        ON r.record_id = h.record_id
      LEFT JOIN public.hosts target
        ON target.record_id = h.merged_into_record_id
     WHERE r.record_id IS NULL
        OR r.incident_id IS DISTINCT FROM h.incident_id
        OR r.record_type IS DISTINCT FROM 'host'
        OR r.row_version IS DISTINCT FROM h.row_version
        OR r.created_at IS DISTINCT FROM h.created_at
        OR r.updated_at IS DISTINCT FROM h.updated_at
        OR r.created_by_user_id IS DISTINCT FROM h.created_by_user_id
        OR r.updated_by_user_id IS DISTINCT FROM h.updated_by_user_id
        OR h.entity_origin NOT IN (
            'entity_sheet', 'entity_import', 'created_from_mention', 'system_upsert'
        )
        OR h.host_state NOT IN ('stub', 'canonical', 'merged')
        OR (h.host_state IN ('stub', 'canonical') AND h.merged_into_record_id IS NOT NULL)
        OR (h.host_state = 'merged' AND (
            h.merged_into_record_id IS NULL
            OR h.merged_into_record_id = h.record_id
            OR target.record_id IS NULL
            OR target.incident_id IS DISTINCT FROM h.incident_id
            OR target.host_state = 'merged'
        ));

    SELECT count(*)
      INTO identity_count
      FROM public.identities i
      LEFT JOIN public.records r
        ON r.record_id = i.record_id
      LEFT JOIN public.identities target
        ON target.record_id = i.merged_into_record_id
     WHERE r.record_id IS NULL
        OR r.incident_id IS DISTINCT FROM i.incident_id
        OR r.record_type IS DISTINCT FROM 'identity'
        OR r.row_version IS DISTINCT FROM i.row_version
        OR r.created_at IS DISTINCT FROM i.created_at
        OR r.updated_at IS DISTINCT FROM i.updated_at
        OR r.created_by_user_id IS DISTINCT FROM i.created_by_user_id
        OR r.updated_by_user_id IS DISTINCT FROM i.updated_by_user_id
        OR i.entity_origin NOT IN (
            'entity_sheet', 'entity_import', 'created_from_mention', 'system_upsert'
        )
        OR i.identity_state NOT IN ('stub', 'canonical', 'merged')
        OR (i.identity_state IN ('stub', 'canonical') AND i.merged_into_record_id IS NOT NULL)
        OR (i.identity_state = 'merged' AND (
            i.merged_into_record_id IS NULL
            OR i.merged_into_record_id = i.record_id
            OR target.record_id IS NULL
            OR target.incident_id IS DISTINCT FROM i.incident_id
            OR target.identity_state = 'merged'
        ));

    FOR source_row IN
        SELECT a.*,
               EXISTS (
                   SELECT 1
                     FROM public.hosts h
                    WHERE a.entity_type = 'host'
                      AND h.record_id = a.record_id
                      AND h.incident_id = a.incident_id
               ) OR EXISTS (
                   SELECT 1
                     FROM public.identities i
                    WHERE a.entity_type = 'identity'
                      AND i.record_id = a.record_id
                      AND i.incident_id = a.incident_id
               ) AS owner_valid
          FROM public.entity_aliases a
    LOOP
        candidate := normalize(btrim(source_row.raw_text, unicode_space), NFC);
        invalid := source_row.entity_type NOT IN ('host', 'identity')
            OR source_row.classification <> 'suggestion_only'
            OR NOT source_row.owner_valid
            OR candidate = ''
            OR char_length(candidate) > 256
            OR source_row.raw_text IS DISTINCT FROM candidate
            OR source_row.normalized_text::text IS DISTINCT FROM candidate;
        IF NOT invalid THEN
            FOR character_ordinal IN 1..char_length(candidate) LOOP
                codepoint := ascii(substring(candidate FROM character_ordinal FOR 1));
                IF codepoint BETWEEN 0 AND 31 OR codepoint BETWEEN 127 AND 159 THEN
                    invalid := true;
                    EXIT;
                END IF;
            END LOOP;
        END IF;
        IF invalid THEN
            alias_count := alias_count + 1;
        END IF;
    END LOOP;

    FOR source_row IN
        SELECT p.*,
               EXISTS (
                   SELECT 1
                     FROM public.hosts h
                    WHERE p.entity_type = 'host'
                      AND h.record_id = p.record_id
                      AND h.incident_id = p.incident_id
               ) OR EXISTS (
                   SELECT 1
                     FROM public.identities i
                    WHERE p.entity_type = 'identity'
                      AND i.record_id = p.record_id
                      AND i.incident_id = p.incident_id
               ) AS owner_valid
          FROM public.entity_preserved_identifiers p
    LOOP
        candidate := normalize(btrim(source_row.raw_value, unicode_space), NFC);
        invalid := source_row.classification NOT IN (
                'exact_match_reuse', 'suggestion_only', 'provenance_only'
            )
            OR NOT source_row.owner_valid
            OR NOT (
                (source_row.entity_type = 'host'
                    AND source_row.identifier_type IN (
                        'aad_device_id', 'fqdn', 'hostname'
                    ))
                OR (source_row.entity_type = 'identity'
                    AND source_row.identifier_type IN (
                        'aad_object_id', 'sid', 'upn', 'email', 'sam_account_name'
                    ))
            )
            OR candidate = '';
        IF NOT invalid THEN
            FOR character_ordinal IN 1..char_length(candidate) LOOP
                codepoint := ascii(substring(candidate FROM character_ordinal FOR 1));
                IF codepoint BETWEEN 0 AND 31
                    OR codepoint BETWEEN 127 AND 159
                    OR codepoint = 173
                    OR codepoint BETWEEN 1536 AND 1541
                    OR codepoint = 1564
                    OR codepoint = 1757
                    OR codepoint = 1807
                    OR codepoint BETWEEN 2192 AND 2193
                    OR codepoint = 2274
                    OR codepoint = 6158
                    OR codepoint BETWEEN 8203 AND 8207
                    OR codepoint BETWEEN 8234 AND 8238
                    OR codepoint BETWEEN 8288 AND 8292
                    OR codepoint BETWEEN 8294 AND 8303
                    OR codepoint = 65279
                    OR codepoint BETWEEN 65529 AND 65531
                    OR codepoint IN (69821, 69837)
                    OR codepoint BETWEEN 78896 AND 78911
                    OR codepoint BETWEEN 113824 AND 113827
                    OR codepoint BETWEEN 119155 AND 119162
                    OR codepoint = 917505
                    OR codepoint BETWEEN 917536 AND 917631 THEN
                    invalid := true;
                    EXIT;
                END IF;
            END LOOP;
        END IF;
        IF NOT invalid THEN
            IF source_row.identifier_type = 'sid' THEN
                candidate := upper(candidate);
            ELSE
                candidate := lower(candidate);
            END IF;
            invalid := source_row.normalized_value IS DISTINCT FROM candidate;
        END IF;
        IF invalid THEN
            preserved_count := preserved_count + 1;
        END IF;
    END LOOP;

    SELECT count(*)
      INTO mention_count
      FROM public.entity_mentions m
      LEFT JOIN public.records source_record
        ON source_record.record_id = m.source_record_id
      LEFT JOIN public.records resolved_record
        ON resolved_record.record_id = m.resolved_record_id
     WHERE m.entity_type NOT IN ('host', 'identity')
        OR m.origin_kind NOT IN (
            'manual_entry', 'clipboard_paste', 'csv_import', 'xlsx_import',
            'api_import', 'extraction', 'system'
        )
        OR btrim(m.source_field_key) = ''
        OR btrim(m.origin_locator) = ''
        OR btrim(m.raw_text) = ''
        OR btrim(m.normalized_text) = ''
        OR m.row_version < 1
        OR m.ordinal < 1
        OR source_record.record_id IS NULL
        OR (
            m.resolution_status = 'resolved'
            AND (
                m.resolved_record_id IS NULL
                OR m.resolved_by_user_id IS NULL
                OR m.resolved_at IS NULL
                OR btrim(coalesce(m.resolution_method, '')) = ''
                OR resolved_record.record_id IS NULL
                OR resolved_record.incident_id IS DISTINCT FROM source_record.incident_id
                OR resolved_record.record_type IS DISTINCT FROM m.entity_type
            )
        )
        OR (
            m.resolution_status IN ('unresolved', 'dismissed')
            AND (
                m.resolved_record_id IS NOT NULL
                OR m.resolved_by_user_id IS NOT NULL
                OR m.resolved_at IS NOT NULL
                OR m.resolution_method IS NOT NULL
            )
        );

    IF host_count + identity_count + alias_count + preserved_count + mention_count <> 0 THEN
        RAISE EXCEPTION USING
            ERRCODE = '23514',
            MESSAGE = 'entities_source_integrity_preflight_failed',
            DETAIL = format(
                'hosts=%s identities=%s aliases=%s preserved_identifiers=%s mentions=%s',
                host_count,
                identity_count,
                alias_count,
                preserved_count,
                mention_count
            ),
            HINT = 'Keep the pre-cutover binary active and explicitly disposition invalid Entities source rows before retrying; this migration never repairs or chooses a winner.';
    END IF;
END;
$$;
-- +goose StatementEnd

ALTER TABLE public.entity_preserved_identifiers
    ADD CONSTRAINT entity_preserved_identifiers_identifier_class_ck CHECK (
        (entity_type = 'host' AND identifier_type IN (
            'aad_device_id', 'fqdn', 'hostname'
        ))
        OR (entity_type = 'identity' AND identifier_type IN (
            'aad_object_id', 'sid', 'upn', 'email', 'sam_account_name'
        ))
    ),
    ADD CONSTRAINT entity_preserved_identifiers_value_nonempty_ck CHECK (
        btrim(raw_value) <> '' AND btrim(normalized_value) <> ''
    );

ALTER TABLE public.entity_mentions
    ADD CONSTRAINT entity_mentions_origin_kind_ck CHECK (
        origin_kind IN (
            'manual_entry', 'clipboard_paste', 'csv_import', 'xlsx_import',
            'api_import', 'extraction', 'system'
        )
    ),
    ADD CONSTRAINT entity_mentions_required_text_ck CHECK (
        btrim(source_field_key) <> ''
        AND btrim(origin_locator) <> ''
        AND btrim(raw_text) <> ''
        AND btrim(normalized_text) <> ''
    ),
    ADD CONSTRAINT entity_mentions_row_version_ck CHECK (row_version >= 1),
    ADD CONSTRAINT entity_mentions_resolution_tuple_ck CHECK (
        (resolution_status = 'resolved'
            AND resolved_record_id IS NOT NULL
            AND resolved_by_user_id IS NOT NULL
            AND resolved_at IS NOT NULL
            AND btrim(coalesce(resolution_method, '')) <> '')
        OR (resolution_status IN ('unresolved', 'dismissed')
            AND resolved_record_id IS NULL
            AND resolved_by_user_id IS NULL
            AND resolved_at IS NULL
            AND resolution_method IS NULL)
    );

-- +goose StatementBegin
CREATE FUNCTION public.entities_source_codepoints_admitted_v1(
    candidate text,
    reject_format boolean
)
RETURNS boolean
LANGUAGE plpgsql
IMMUTABLE
STRICT
PARALLEL SAFE
SET search_path = pg_catalog, public
AS $$
DECLARE
    codepoint integer;
    character_ordinal integer;
BEGIN
    FOR character_ordinal IN 1..char_length(candidate) LOOP
        codepoint := ascii(substring(candidate FROM character_ordinal FOR 1));
        IF codepoint BETWEEN 0 AND 31
            OR codepoint BETWEEN 127 AND 159
            OR (reject_format AND (
                codepoint = 173
                OR codepoint BETWEEN 1536 AND 1541
                OR codepoint = 1564
                OR codepoint = 1757
                OR codepoint = 1807
                OR codepoint BETWEEN 2192 AND 2193
                OR codepoint = 2274
                OR codepoint = 6158
                OR codepoint BETWEEN 8203 AND 8207
                OR codepoint BETWEEN 8234 AND 8238
                OR codepoint BETWEEN 8288 AND 8292
                OR codepoint BETWEEN 8294 AND 8303
                OR codepoint = 65279
                OR codepoint BETWEEN 65529 AND 65531
                OR codepoint IN (69821, 69837)
                OR codepoint BETWEEN 78896 AND 78911
                OR codepoint BETWEEN 113824 AND 113827
                OR codepoint BETWEEN 119155 AND 119162
                OR codepoint = 917505
                OR codepoint BETWEEN 917536 AND 917631
            )) THEN
            RETURN false;
        END IF;
    END LOOP;
    RETURN true;
END;
$$;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE FUNCTION public.entities_trim_unicode_space_v1(candidate text)
RETURNS text
LANGUAGE sql
IMMUTABLE
STRICT
PARALLEL SAFE
SET search_path = pg_catalog, public
RETURN btrim(
    candidate,
    chr(9) || chr(10) || chr(11) || chr(12) || chr(13)
        || chr(32) || chr(133) || chr(160) || chr(5760)
        || chr(8192) || chr(8193) || chr(8194) || chr(8195) || chr(8196)
        || chr(8197) || chr(8198) || chr(8199) || chr(8200) || chr(8201)
        || chr(8202) || chr(8232) || chr(8233) || chr(8239) || chr(8287)
        || chr(12288)
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE FUNCTION public.entities_assert_entity_source_v1()
RETURNS trigger
LANGUAGE plpgsql
SET search_path = pg_catalog, public
AS $$
DECLARE
    envelope public.records%ROWTYPE;
    target_incident_id uuid;
    target_state text;
BEGIN
    SELECT *
      INTO envelope
      FROM public.records
     WHERE record_id = NEW.record_id;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING
            ERRCODE = '23514',
            MESSAGE = 'entities_source_envelope_missing';
    END IF;
    IF envelope.incident_id IS DISTINCT FROM NEW.incident_id
        OR envelope.record_type IS DISTINCT FROM TG_ARGV[0]
        OR envelope.row_version IS DISTINCT FROM NEW.row_version
        OR envelope.created_at IS DISTINCT FROM NEW.created_at
        OR envelope.updated_at IS DISTINCT FROM NEW.updated_at
        OR envelope.created_by_user_id IS DISTINCT FROM NEW.created_by_user_id
        OR envelope.updated_by_user_id IS DISTINCT FROM NEW.updated_by_user_id THEN
        RAISE EXCEPTION USING
            ERRCODE = '23514',
            MESSAGE = 'entities_source_envelope_mismatch';
    END IF;

    IF TG_ARGV[0] = 'host' THEN
        IF NEW.host_state = 'merged' THEN
            IF NEW.merged_into_record_id = NEW.record_id THEN
                RAISE EXCEPTION USING ERRCODE = '23514', MESSAGE = 'entities_source_merge_self_reference';
            END IF;
            SELECT incident_id, host_state
              INTO target_incident_id, target_state
              FROM public.hosts
             WHERE record_id = NEW.merged_into_record_id;
        ELSE
            RETURN NEW;
        END IF;
    ELSE
        IF NEW.identity_state = 'merged' THEN
            IF NEW.merged_into_record_id = NEW.record_id THEN
                RAISE EXCEPTION USING ERRCODE = '23514', MESSAGE = 'entities_source_merge_self_reference';
            END IF;
            SELECT incident_id, identity_state
              INTO target_incident_id, target_state
              FROM public.identities
             WHERE record_id = NEW.merged_into_record_id;
        ELSE
            RETURN NEW;
        END IF;
    END IF;
    IF NOT FOUND
        OR target_incident_id IS DISTINCT FROM NEW.incident_id
        OR target_state = 'merged' THEN
        RAISE EXCEPTION USING
            ERRCODE = '23514',
            MESSAGE = 'entities_source_merge_target_invalid';
    END IF;
    RETURN NEW;
END;
$$;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE FUNCTION public.entities_assert_entity_envelope_update_v1()
RETURNS trigger
LANGUAGE plpgsql
SET search_path = pg_catalog, public
AS $$
DECLARE
    source_count integer;
BEGIN
    IF NEW.record_type = 'host' THEN
        SELECT count(*)
          INTO source_count
          FROM public.hosts h
         WHERE h.record_id = NEW.record_id
           AND h.incident_id = NEW.incident_id
           AND h.row_version = NEW.row_version
           AND h.created_at = NEW.created_at
           AND h.updated_at = NEW.updated_at
           AND h.created_by_user_id = NEW.created_by_user_id
           AND h.updated_by_user_id = NEW.updated_by_user_id;
    ELSIF NEW.record_type = 'identity' THEN
        SELECT count(*)
          INTO source_count
          FROM public.identities i
         WHERE i.record_id = NEW.record_id
           AND i.incident_id = NEW.incident_id
           AND i.row_version = NEW.row_version
           AND i.created_at = NEW.created_at
           AND i.updated_at = NEW.updated_at
           AND i.created_by_user_id = NEW.created_by_user_id
           AND i.updated_by_user_id = NEW.updated_by_user_id;
    ELSE
        RETURN NEW;
    END IF;
    IF source_count <> 1 THEN
        RAISE EXCEPTION USING
            ERRCODE = '23514',
            MESSAGE = 'entities_envelope_source_mismatch';
    END IF;
    RETURN NEW;
END;
$$;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE FUNCTION public.entities_reject_entity_source_delete_v1()
RETURNS trigger
LANGUAGE plpgsql
SET search_path = pg_catalog, public
AS $$
BEGIN
    IF EXISTS (
        SELECT 1
          FROM public.records r
         WHERE r.record_id = OLD.record_id
           AND r.record_type = TG_ARGV[0]
    ) THEN
        RAISE EXCEPTION USING
            ERRCODE = '23514',
            MESSAGE = 'entities_source_delete_requires_envelope_delete';
    END IF;
    RETURN OLD;
END;
$$;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE FUNCTION public.entities_assert_entity_child_v1()
RETURNS trigger
LANGUAGE plpgsql
SET search_path = pg_catalog, public
AS $$
DECLARE
    owner_valid boolean;
    source_incident_id uuid;
    resolved_incident_id uuid;
    resolved_record_type text;
    candidate text;
BEGIN
    IF TG_TABLE_NAME = 'entity_mentions' THEN
        SELECT incident_id
          INTO source_incident_id
          FROM public.records
         WHERE record_id = NEW.source_record_id;
        IF NOT FOUND THEN
            RAISE EXCEPTION USING ERRCODE = '23514', MESSAGE = 'entities_mention_source_invalid';
        END IF;
        IF NEW.resolution_status = 'resolved' THEN
            SELECT incident_id, record_type
              INTO resolved_incident_id, resolved_record_type
              FROM public.records
             WHERE record_id = NEW.resolved_record_id;
            IF NOT FOUND
                OR resolved_incident_id IS DISTINCT FROM source_incident_id
                OR resolved_record_type IS DISTINCT FROM NEW.entity_type THEN
                RAISE EXCEPTION USING ERRCODE = '23514', MESSAGE = 'entities_mention_target_invalid';
            END IF;
        END IF;
        RETURN NEW;
    END IF;

    IF NEW.entity_type = 'host' THEN
        SELECT EXISTS (
            SELECT 1
              FROM public.hosts h
             WHERE h.record_id = NEW.record_id
               AND h.incident_id = NEW.incident_id
        ) INTO owner_valid;
    ELSIF NEW.entity_type = 'identity' THEN
        SELECT EXISTS (
            SELECT 1
              FROM public.identities i
             WHERE i.record_id = NEW.record_id
               AND i.incident_id = NEW.incident_id
        ) INTO owner_valid;
    ELSE
        owner_valid := false;
    END IF;
    IF NOT owner_valid THEN
        RAISE EXCEPTION USING ERRCODE = '23514', MESSAGE = 'entities_child_owner_invalid';
    END IF;

    IF TG_TABLE_NAME = 'entity_aliases' THEN
        candidate := normalize(public.entities_trim_unicode_space_v1(NEW.raw_text), NFC);
        -- Leave the existing, field-specific empty/control/length CHECKs as the
        -- authoritative error source for values they already cover.  This
        -- trigger closes only the Unicode trim/NFC and exact-projection gap.
        IF candidate <> ''
            AND char_length(candidate) <= 256
            AND public.entities_source_codepoints_admitted_v1(candidate, false)
            AND (
                NEW.raw_text IS DISTINCT FROM candidate
                OR NEW.normalized_text::text IS DISTINCT FROM candidate
            ) THEN
            RAISE EXCEPTION USING ERRCODE = '23514', MESSAGE = 'entities_alias_normalization_invalid';
        END IF;
        RETURN NEW;
    END IF;

    IF NOT (
        (NEW.entity_type = 'host' AND NEW.identifier_type IN (
            'aad_device_id', 'fqdn', 'hostname'
        ))
        OR (NEW.entity_type = 'identity' AND NEW.identifier_type IN (
            'aad_object_id', 'sid', 'upn', 'email', 'sam_account_name'
        ))
    ) THEN
        RAISE EXCEPTION USING ERRCODE = '23514', MESSAGE = 'entities_identifier_class_invalid';
    END IF;
    candidate := normalize(public.entities_trim_unicode_space_v1(NEW.raw_value), NFC);
    IF candidate = ''
        OR NOT public.entities_source_codepoints_admitted_v1(candidate, true) THEN
        RAISE EXCEPTION USING ERRCODE = '23514', MESSAGE = 'entities_identifier_value_invalid';
    END IF;
    IF NEW.identifier_type = 'sid' THEN
        candidate := upper(candidate);
    ELSE
        candidate := lower(candidate);
    END IF;
    IF NEW.normalized_value IS DISTINCT FROM candidate THEN
        RAISE EXCEPTION USING ERRCODE = '23514', MESSAGE = 'entities_identifier_normalization_invalid';
    END IF;
    RETURN NEW;
END;
$$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION public.entities_source_codepoints_admitted_v1(text, boolean) FROM PUBLIC;
REVOKE ALL ON FUNCTION public.entities_trim_unicode_space_v1(text) FROM PUBLIC;
REVOKE ALL ON FUNCTION public.entities_assert_entity_source_v1() FROM PUBLIC;
REVOKE ALL ON FUNCTION public.entities_assert_entity_envelope_update_v1() FROM PUBLIC;
REVOKE ALL ON FUNCTION public.entities_reject_entity_source_delete_v1() FROM PUBLIC;
REVOKE ALL ON FUNCTION public.entities_assert_entity_child_v1() FROM PUBLIC;

GRANT EXECUTE ON FUNCTION public.entities_source_codepoints_admitted_v1(text, boolean)
    TO cartulary_runtime, cartulary_recovery;
GRANT EXECUTE ON FUNCTION public.entities_trim_unicode_space_v1(text)
    TO cartulary_runtime, cartulary_recovery;

CREATE TRIGGER entities_assert_host_source
BEFORE INSERT OR UPDATE ON public.hosts
FOR EACH ROW EXECUTE FUNCTION public.entities_assert_entity_source_v1('host');

CREATE TRIGGER entities_assert_identity_source
BEFORE INSERT OR UPDATE ON public.identities
FOR EACH ROW EXECUTE FUNCTION public.entities_assert_entity_source_v1('identity');

CREATE CONSTRAINT TRIGGER entities_assert_entity_envelope_update
AFTER UPDATE OF incident_id, record_type, row_version, created_at, updated_at,
    created_by_user_id, updated_by_user_id ON public.records
DEFERRABLE INITIALLY DEFERRED
FOR EACH ROW EXECUTE FUNCTION public.entities_assert_entity_envelope_update_v1();

CREATE CONSTRAINT TRIGGER entities_reject_host_source_delete
AFTER DELETE ON public.hosts
DEFERRABLE INITIALLY DEFERRED
FOR EACH ROW EXECUTE FUNCTION public.entities_reject_entity_source_delete_v1('host');

CREATE CONSTRAINT TRIGGER entities_reject_identity_source_delete
AFTER DELETE ON public.identities
DEFERRABLE INITIALLY DEFERRED
FOR EACH ROW EXECUTE FUNCTION public.entities_reject_entity_source_delete_v1('identity');

CREATE TRIGGER entities_assert_entity_alias
BEFORE INSERT OR UPDATE ON public.entity_aliases
FOR EACH ROW EXECUTE FUNCTION public.entities_assert_entity_child_v1();

CREATE TRIGGER entities_assert_entity_preserved_identifier
BEFORE INSERT OR UPDATE ON public.entity_preserved_identifiers
FOR EACH ROW EXECUTE FUNCTION public.entities_assert_entity_child_v1();

CREATE TRIGGER entities_assert_entity_mention
BEFORE INSERT OR UPDATE ON public.entity_mentions
FOR EACH ROW EXECUTE FUNCTION public.entities_assert_entity_child_v1();

-- +goose Down
DROP TRIGGER entities_assert_entity_mention ON public.entity_mentions;
DROP TRIGGER entities_assert_entity_preserved_identifier ON public.entity_preserved_identifiers;
DROP TRIGGER entities_assert_entity_alias ON public.entity_aliases;
DROP TRIGGER entities_reject_identity_source_delete ON public.identities;
DROP TRIGGER entities_reject_host_source_delete ON public.hosts;
DROP TRIGGER entities_assert_entity_envelope_update ON public.records;
DROP TRIGGER entities_assert_identity_source ON public.identities;
DROP TRIGGER entities_assert_host_source ON public.hosts;

DROP FUNCTION public.entities_assert_entity_child_v1();
DROP FUNCTION public.entities_reject_entity_source_delete_v1();
DROP FUNCTION public.entities_assert_entity_envelope_update_v1();
DROP FUNCTION public.entities_assert_entity_source_v1();
DROP FUNCTION public.entities_trim_unicode_space_v1(text);
DROP FUNCTION public.entities_source_codepoints_admitted_v1(text, boolean);

ALTER TABLE public.entity_mentions
    DROP CONSTRAINT entity_mentions_resolution_tuple_ck,
    DROP CONSTRAINT entity_mentions_row_version_ck,
    DROP CONSTRAINT entity_mentions_required_text_ck,
    DROP CONSTRAINT entity_mentions_origin_kind_ck;

ALTER TABLE public.entity_preserved_identifiers
    DROP CONSTRAINT entity_preserved_identifiers_value_nonempty_ck,
    DROP CONSTRAINT entity_preserved_identifiers_identifier_class_ck;
