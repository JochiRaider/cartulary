-- +goose Up
--
-- Assessment-create idempotency rows are durable committed replay history.
-- Classify every target row before changing any payload. Failure details are
-- aggregate counts only and intentionally exclude stored identities or values.
--

-- +goose StatementBegin
DO $$
DECLARE
    target_count bigint;
    canonical_count bigint;
    legacy_count bigint;
    invalid_count bigint;
BEGIN
    SELECT
        count(*),
        count(*) FILTER (WHERE
            status_code = 201
            AND octet_length(request_hash) > 0
            AND scope_key ~ '^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}:cartulary\.view\.assessments\.v1$'
            AND split_part(scope_key, ':', 1) <> '00000000-0000-0000-0000-000000000000'
            AND jsonb_typeof(response_json) = 'object'
            AND response_json ?& ARRAY['schema_id', 'record_id', 'change_set_id', 'row_version', 'row']
            AND response_json - ARRAY['schema_id', 'record_id', 'change_set_id', 'row_version', 'row'] = '{}'::jsonb
            AND jsonb_typeof(response_json -> 'schema_id') = 'string'
            AND response_json ->> 'schema_id' = 'cartulary.assessments.create_result.v1'
            AND jsonb_typeof(response_json -> 'record_id') = 'string'
            AND response_json ->> 'record_id' ~ '^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$'
            AND response_json ->> 'record_id' <> '00000000-0000-0000-0000-000000000000'
            AND jsonb_typeof(response_json -> 'change_set_id') = 'string'
            AND response_json ->> 'change_set_id' ~ '^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$'
            AND response_json ->> 'change_set_id' <> '00000000-0000-0000-0000-000000000000'
            AND jsonb_typeof(response_json -> 'row_version') = 'number'
            AND response_json ->> 'row_version' ~ '^[1-9][0-9]*$'
            AND (length(response_json ->> 'row_version') < 19
                OR (length(response_json ->> 'row_version') = 19
                    AND response_json ->> 'row_version' <= '9223372036854775807'))
            AND jsonb_typeof(response_json -> 'row') = 'object'
            AND response_json -> 'row' ?& ARRAY['record_id', 'row_version']
            AND jsonb_typeof(response_json -> 'row' -> 'record_id') = 'string'
            AND response_json -> 'row' ->> 'record_id' = response_json ->> 'record_id'
            AND (NOT (response_json -> 'row' ? 'incident_id')
                OR (jsonb_typeof(response_json -> 'row' -> 'incident_id') = 'string'
                    AND response_json -> 'row' ->> 'incident_id' = split_part(scope_key, ':', 1)))
            AND jsonb_typeof(response_json -> 'row' -> 'row_version') = 'number'
            AND response_json -> 'row' ->> 'row_version' = response_json ->> 'row_version'
        ),
        count(*) FILTER (WHERE
            status_code = 201
            AND octet_length(request_hash) > 0
            AND scope_key ~ '^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}:cartulary\.view\.assessments\.v1$'
            AND split_part(scope_key, ':', 1) <> '00000000-0000-0000-0000-000000000000'
            AND jsonb_typeof(response_json) = 'object'
            AND response_json ?& ARRAY['view_schema_id', 'change_set_id', 'row']
            AND response_json - ARRAY['view_schema_id', 'change_set_id', 'row'] = '{}'::jsonb
            AND jsonb_typeof(response_json -> 'view_schema_id') = 'string'
            AND response_json ->> 'view_schema_id' = 'cartulary.view.assessments.v1'
            AND jsonb_typeof(response_json -> 'change_set_id') = 'string'
            AND response_json ->> 'change_set_id' ~ '^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$'
            AND response_json ->> 'change_set_id' <> '00000000-0000-0000-0000-000000000000'
            AND jsonb_typeof(response_json -> 'row') = 'object'
            AND response_json -> 'row' ?& ARRAY['record_id', 'row_version']
            AND jsonb_typeof(response_json -> 'row' -> 'record_id') = 'string'
            AND response_json -> 'row' ->> 'record_id' ~ '^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$'
            AND response_json -> 'row' ->> 'record_id' <> '00000000-0000-0000-0000-000000000000'
            AND (NOT (response_json -> 'row' ? 'incident_id')
                OR (jsonb_typeof(response_json -> 'row' -> 'incident_id') = 'string'
                    AND response_json -> 'row' ->> 'incident_id' = split_part(scope_key, ':', 1)))
            AND jsonb_typeof(response_json -> 'row' -> 'row_version') = 'number'
            AND response_json -> 'row' ->> 'row_version' ~ '^[1-9][0-9]*$'
            AND (length(response_json -> 'row' ->> 'row_version') < 19
                OR (length(response_json -> 'row' ->> 'row_version') = 19
                    AND response_json -> 'row' ->> 'row_version' <= '9223372036854775807'))
        )
      INTO target_count, canonical_count, legacy_count
      FROM public.route_idempotency
     WHERE route_key = 'assessments.rows.create';

    invalid_count := target_count - canonical_count - legacy_count;
    IF invalid_count <> 0 THEN
        RAISE EXCEPTION USING
            ERRCODE = '23514',
            MESSAGE = 'assessment_create_idempotency_v1_preflight_failed',
            DETAIL = format(
                'target=%s canonical=%s legacy=%s invalid=%s',
                target_count,
                canonical_count,
                legacy_count,
                invalid_count
            ),
            HINT = 'Keep the pre-cutover binary active and disposition unsupported assessment-create replay state before retrying.';
    END IF;

    UPDATE public.route_idempotency
       SET response_json = CASE
            WHEN response_json ->> 'view_schema_id' = 'cartulary.view.assessments.v1'
                THEN jsonb_build_object(
                    'schema_id', 'cartulary.assessments.create_result.v1',
                    'record_id', response_json -> 'row' -> 'record_id',
                    'change_set_id', response_json -> 'change_set_id',
                    'row_version', response_json -> 'row' -> 'row_version',
                    'row', response_json -> 'row' || jsonb_build_object(
                        'incident_id', split_part(scope_key, ':', 1)
                    )
                )
            ELSE jsonb_set(
                response_json,
                '{row}',
                response_json -> 'row' || jsonb_build_object(
                    'incident_id', split_part(scope_key, ':', 1)
                )
            )
       END
     WHERE route_key = 'assessments.rows.create'
       AND (response_json ->> 'view_schema_id' = 'cartulary.view.assessments.v1'
            OR (response_json ->> 'schema_id' = 'cartulary.assessments.create_result.v1'
                AND NOT (response_json -> 'row' ? 'incident_id')));
END;
$$;
-- +goose StatementEnd

-- +goose Down
-- Down supports deterministic disposable-database evidence. Roll back the
-- strict binary before using it; production migration remains up-only.

-- +goose StatementBegin
DO $$
DECLARE
    target_count bigint;
    canonical_count bigint;
    legacy_count bigint;
    invalid_count bigint;
BEGIN
    SELECT
        count(*),
        count(*) FILTER (WHERE
            status_code = 201
            AND octet_length(request_hash) > 0
            AND scope_key ~ '^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}:cartulary\.view\.assessments\.v1$'
            AND split_part(scope_key, ':', 1) <> '00000000-0000-0000-0000-000000000000'
            AND jsonb_typeof(response_json) = 'object'
            AND response_json ?& ARRAY['schema_id', 'record_id', 'change_set_id', 'row_version', 'row']
            AND response_json - ARRAY['schema_id', 'record_id', 'change_set_id', 'row_version', 'row'] = '{}'::jsonb
            AND jsonb_typeof(response_json -> 'schema_id') = 'string'
            AND response_json ->> 'schema_id' = 'cartulary.assessments.create_result.v1'
            AND jsonb_typeof(response_json -> 'record_id') = 'string'
            AND response_json ->> 'record_id' ~ '^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$'
            AND response_json ->> 'record_id' <> '00000000-0000-0000-0000-000000000000'
            AND jsonb_typeof(response_json -> 'change_set_id') = 'string'
            AND response_json ->> 'change_set_id' ~ '^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$'
            AND response_json ->> 'change_set_id' <> '00000000-0000-0000-0000-000000000000'
            AND jsonb_typeof(response_json -> 'row_version') = 'number'
            AND response_json ->> 'row_version' ~ '^[1-9][0-9]*$'
            AND (length(response_json ->> 'row_version') < 19
                OR (length(response_json ->> 'row_version') = 19
                    AND response_json ->> 'row_version' <= '9223372036854775807'))
            AND jsonb_typeof(response_json -> 'row') = 'object'
            AND response_json -> 'row' ?& ARRAY['record_id', 'incident_id', 'row_version']
            AND jsonb_typeof(response_json -> 'row' -> 'record_id') = 'string'
            AND response_json -> 'row' ->> 'record_id' = response_json ->> 'record_id'
            AND jsonb_typeof(response_json -> 'row' -> 'incident_id') = 'string'
            AND response_json -> 'row' ->> 'incident_id' = split_part(scope_key, ':', 1)
            AND jsonb_typeof(response_json -> 'row' -> 'row_version') = 'number'
            AND response_json -> 'row' ->> 'row_version' = response_json ->> 'row_version'
        ),
        count(*) FILTER (WHERE
            status_code = 201
            AND octet_length(request_hash) > 0
            AND scope_key ~ '^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}:cartulary\.view\.assessments\.v1$'
            AND split_part(scope_key, ':', 1) <> '00000000-0000-0000-0000-000000000000'
            AND jsonb_typeof(response_json) = 'object'
            AND response_json ?& ARRAY['view_schema_id', 'change_set_id', 'row']
            AND response_json - ARRAY['view_schema_id', 'change_set_id', 'row'] = '{}'::jsonb
            AND jsonb_typeof(response_json -> 'view_schema_id') = 'string'
            AND response_json ->> 'view_schema_id' = 'cartulary.view.assessments.v1'
            AND jsonb_typeof(response_json -> 'change_set_id') = 'string'
            AND response_json ->> 'change_set_id' ~ '^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$'
            AND response_json ->> 'change_set_id' <> '00000000-0000-0000-0000-000000000000'
            AND jsonb_typeof(response_json -> 'row') = 'object'
            AND response_json -> 'row' ?& ARRAY['record_id', 'row_version']
            AND jsonb_typeof(response_json -> 'row' -> 'record_id') = 'string'
            AND response_json -> 'row' ->> 'record_id' ~ '^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$'
            AND response_json -> 'row' ->> 'record_id' <> '00000000-0000-0000-0000-000000000000'
            AND (NOT (response_json -> 'row' ? 'incident_id')
                OR (jsonb_typeof(response_json -> 'row' -> 'incident_id') = 'string'
                    AND response_json -> 'row' ->> 'incident_id' = split_part(scope_key, ':', 1)))
            AND jsonb_typeof(response_json -> 'row' -> 'row_version') = 'number'
            AND response_json -> 'row' ->> 'row_version' ~ '^[1-9][0-9]*$'
            AND (length(response_json -> 'row' ->> 'row_version') < 19
                OR (length(response_json -> 'row' ->> 'row_version') = 19
                    AND response_json -> 'row' ->> 'row_version' <= '9223372036854775807'))
        )
      INTO target_count, canonical_count, legacy_count
      FROM public.route_idempotency
     WHERE route_key = 'assessments.rows.create';

    invalid_count := target_count - canonical_count - legacy_count;
    IF invalid_count <> 0 THEN
        RAISE EXCEPTION USING
            ERRCODE = '23514',
            MESSAGE = 'assessment_create_idempotency_v1_rollback_preflight_failed',
            DETAIL = format(
                'target=%s canonical=%s legacy=%s invalid=%s',
                target_count,
                canonical_count,
                legacy_count,
                invalid_count
            ),
            HINT = 'Roll back the strict binary and disposition unsupported assessment-create replay state before retrying.';
    END IF;

    UPDATE public.route_idempotency
       SET response_json = jsonb_build_object(
            'view_schema_id', 'cartulary.view.assessments.v1',
            'change_set_id', response_json -> 'change_set_id',
            'row', (response_json -> 'row') - 'incident_id'
       )
     WHERE route_key = 'assessments.rows.create'
       AND response_json ->> 'schema_id' = 'cartulary.assessments.create_result.v1';
END;
$$;
-- +goose StatementEnd
