-- +goose Up
DROP TRIGGER network_flow_tables_collaboration_resource_intent ON public.network_flow_tables;
DROP FUNCTION public.collaboration_capture_network_flow_resource_v1();
DROP TRIGGER jobs_collaboration_progress_intent ON public.jobs;
DROP FUNCTION public.collaboration_capture_job_progress_v1();
DROP FUNCTION public.collaboration_rfc3339_v1(timestamp with time zone);
DROP TRIGGER record_revisions_collaboration_event_intent ON public.record_revisions;
DROP FUNCTION public.collaboration_capture_record_revision_v1();

-- +goose Down
-- +goose StatementBegin
CREATE FUNCTION public.collaboration_capture_record_revision_v1()
RETURNS trigger
LANGUAGE plpgsql
AS $$
DECLARE
    record_row public.records%ROWTYPE;
    change_set_row public.change_sets%ROWTYPE;
    row_snapshot jsonb;
    field_prefix text;
    view_schema_id text;
    change_kind text;
    changed_field_keys jsonb;
    patch_cells jsonb;
    patch_group_values jsonb;
    affected_view jsonb;
    payload jsonb;
    ordinal integer;
BEGIN
    IF current_setting('cartulary.collaboration.suppress_historical_intents', true) = 'on' THEN
        RETURN NEW;
    END IF;
    SELECT * INTO STRICT record_row FROM public.records WHERE record_id = NEW.record_id;
    SELECT * INTO STRICT change_set_row FROM public.change_sets WHERE change_set_id = NEW.change_set_id;
    row_snapshot := COALESCE(NEW.after_json, NEW.before_json, '{}'::jsonb);
    SELECT split_part(key, '.', 1)
      INTO field_prefix
      FROM jsonb_object_keys(COALESCE(row_snapshot -> 'cells', '{}'::jsonb)) AS key
     ORDER BY key
     LIMIT 1;
    view_schema_id := CASE record_row.record_type
        WHEN 'timeline_event' THEN 'cartulary.view.timeline.v2'
        WHEN 'host' THEN 'cartulary.view.hosts.v1'
        WHEN 'identity' THEN 'cartulary.view.identities.v1'
        WHEN 'party' THEN 'cartulary.view.parties.v1'
        WHEN 'indicator' THEN 'cartulary.view.indicators.v1'
        WHEN 'evidence' THEN 'cartulary.view.evidence.v1'
        WHEN 'task_request' THEN 'cartulary.view.task_requests.v1'
        WHEN 'decision' THEN 'cartulary.view.decisions.v1'
        WHEN 'assessment' THEN 'cartulary.view.assessments.v1'
        WHEN 'artifact' THEN CASE field_prefix
            WHEN 'finding' THEN 'cartulary.view.findings.v1'
            WHEN 'query' THEN 'cartulary.view.investigative_queries.v1'
            WHEN 'keyword' THEN 'cartulary.view.forensic_keywords.v1'
            WHEN 'comm_log' THEN 'cartulary.view.comm_log.v1'
            WHEN 'status_review' THEN 'cartulary.view.status_review.v1'
            WHEN 'handoff' THEN 'cartulary.view.handoff.v1'
            WHEN 'lesson' THEN 'cartulary.view.lesson.v1'
            ELSE 'cartulary.view.notes.v1'
        END
        ELSE NULL
    END;
    IF view_schema_id IS NULL THEN
        RAISE EXCEPTION 'record_changed view schema is not mapped for record type %', record_row.record_type
            USING ERRCODE = '23514';
    END IF;
    SELECT COALESCE(jsonb_agg(candidate.key ORDER BY candidate.key), '[]'::jsonb)
      INTO changed_field_keys
      FROM (
          SELECT key
            FROM (
                SELECT jsonb_object_keys(COALESCE(NEW.before_json -> 'cells', '{}'::jsonb)) AS key
                UNION
                SELECT jsonb_object_keys(COALESCE(NEW.after_json -> 'cells', '{}'::jsonb)) AS key
            ) AS keys
           WHERE COALESCE(NEW.before_json -> 'cells', '{}'::jsonb) -> keys.key
                 IS DISTINCT FROM
                 COALESCE(NEW.after_json -> 'cells', '{}'::jsonb) -> keys.key
      ) AS candidate;
    change_kind := CASE WHEN record_row.deleted_at IS NULL THEN 'invalidate' ELSE 'remove' END;
    affected_view := jsonb_build_object('view_schema_id', view_schema_id, 'change_kind', change_kind);
    IF change_kind <> 'remove'
       AND change_set_row.source NOT IN ('records.restore', 'rollback')
       AND jsonb_typeof(NEW.after_json -> 'cells') = 'object' THEN
        SELECT COALESCE(jsonb_object_agg(changed.key, NEW.after_json -> 'cells' -> changed.key), '{}'::jsonb)
          INTO patch_cells
          FROM jsonb_array_elements_text(changed_field_keys) AS changed(key)
         WHERE (NEW.after_json -> 'cells') ? changed.key;
        IF patch_cells <> '{}'::jsonb THEN
            SELECT COALESCE(jsonb_object_agg(changed.key, NEW.after_json -> 'group_values' -> changed.key), '{}'::jsonb)
              INTO patch_group_values
              FROM jsonb_array_elements_text(changed_field_keys) AS changed(key)
             WHERE jsonb_typeof(NEW.after_json -> 'group_values') = 'object'
               AND (NEW.after_json -> 'group_values') ? changed.key;
            patch_cells := jsonb_build_object(
                'record_id', COALESCE(NEW.after_json -> 'record_id', to_jsonb(NEW.record_id::text)),
                'row_version', COALESCE(NEW.after_json -> 'row_version', to_jsonb(NEW.row_version)),
                'cells', patch_cells
            );
            IF patch_group_values <> '{}'::jsonb THEN
                patch_cells := patch_cells || jsonb_build_object('group_values', patch_group_values);
            END IF;
            affected_view := jsonb_build_object(
                'view_schema_id', view_schema_id,
                'change_kind', 'patch',
                'patch_cells', patch_cells
            );
        END IF;
    END IF;
    payload := jsonb_build_object(
        'record_id', NEW.record_id::text,
        'row_version', NEW.row_version,
        'change_set_id', NEW.change_set_id::text,
        'client_txn_id', COALESCE(change_set_row.client_txn_id, ''),
        'actor_user_id', change_set_row.actor_user_id::text,
        'changed_field_keys', changed_field_keys,
        'affected_views', jsonb_build_array(affected_view)
    );
    SELECT GREATEST(COALESCE(min(sequence_no), 1) - 1, 0)
      INTO ordinal
      FROM public.change_set_mutations
     WHERE change_set_id = NEW.change_set_id
       AND target_id = NEW.record_id::text;
    INSERT INTO public.collaboration_event_intents (
        intent_key, incident_id, event_family, canonical_payload,
        source_change_set_id, source_record_id, source_row_version,
        source_identity, mutation_ordinal, next_attempt_at, created_at, updated_at
    ) VALUES (
        'record_changed:' || NEW.change_set_id::text || ':' || NEW.record_id::text || ':' || NEW.row_version::text,
        record_row.incident_id, 'record_changed', payload, NEW.change_set_id,
        NEW.record_id, NEW.row_version, NEW.change_set_id::text || ':' || NEW.record_id::text,
        ordinal, change_set_row.created_at, change_set_row.created_at, change_set_row.created_at
    )
    ON CONFLICT (intent_key) DO NOTHING;
    RETURN NEW;
END;
$$;
-- +goose StatementEnd

CREATE TRIGGER record_revisions_collaboration_event_intent
    AFTER INSERT ON public.record_revisions
    FOR EACH ROW
    EXECUTE FUNCTION public.collaboration_capture_record_revision_v1();

-- +goose StatementBegin
CREATE FUNCTION public.collaboration_rfc3339_v1(value timestamp with time zone)
RETURNS text
LANGUAGE sql
IMMUTABLE
STRICT
AS $$
    SELECT to_char(value AT TIME ZONE 'UTC', 'YYYY-MM-DD"T"HH24:MI:SS')
        || CASE
            WHEN extract(microseconds FROM value)::bigint % 1000000 = 0 THEN ''
            ELSE '.' || rtrim(to_char(value AT TIME ZONE 'UTC', 'US'), '0')
        END
        || 'Z'
$$;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE FUNCTION public.collaboration_capture_job_progress_v1()
RETURNS trigger
LANGUAGE plpgsql
AS $$
DECLARE
    payload jsonb;
    deterministic_key text;
BEGIN
    IF NEW.scope_kind <> 'incident' OR NEW.incident_id IS NULL THEN
        RETURN NEW;
    END IF;
    payload := jsonb_build_object(
        'job_id', NEW.job_id::text,
        'scope', jsonb_build_object('kind', 'incident', 'incident_id', NEW.incident_id::text),
        'status', NEW.status,
        'progress', jsonb_build_object('completed', NEW.progress_completed, 'total', to_jsonb(NEW.progress_total)),
        'updated_at', public.collaboration_rfc3339_v1(NEW.updated_at),
        'cancelable', NEW.cancelable
    );
    IF NEW.message IS NOT NULL THEN
        payload := payload || jsonb_build_object('message', NEW.message);
    END IF;
    IF NEW.result_summary_json IS NOT NULL THEN
        payload := payload || jsonb_build_object('result_summary', NEW.result_summary_json);
    END IF;
    IF NEW.error_summary_json IS NOT NULL THEN
        payload := payload || jsonb_build_object('error_summary', NEW.error_summary_json);
    END IF;
    IF NEW.retained_until IS NOT NULL THEN
        payload := payload || jsonb_build_object('retained_until', public.collaboration_rfc3339_v1(NEW.retained_until));
    END IF;
    deterministic_key := 'job_progress:' || NEW.job_id::text || ':'
        || encode(digest(payload::text, 'sha256'), 'hex');
    INSERT INTO public.collaboration_event_intents (
        intent_key, incident_id, event_family, canonical_payload, source_identity,
        mutation_ordinal, next_attempt_at, created_at, updated_at
    ) VALUES (
        deterministic_key, NEW.incident_id, 'job_progress', payload,
        'job:' || NEW.job_id::text, 0, NEW.updated_at, NEW.updated_at, NEW.updated_at
    )
    ON CONFLICT (intent_key) DO NOTHING;
    RETURN NEW;
END;
$$;
-- +goose StatementEnd

CREATE TRIGGER jobs_collaboration_progress_intent
    AFTER INSERT OR UPDATE ON public.jobs
    FOR EACH ROW
    EXECUTE FUNCTION public.collaboration_capture_job_progress_v1();

-- +goose StatementBegin
CREATE FUNCTION public.collaboration_capture_network_flow_resource_v1()
RETURNS trigger
LANGUAGE plpgsql
AS $$
DECLARE
    change_kind text;
    reason_code text;
    payload jsonb;
    deterministic_key text;
BEGIN
    IF NEW.table_status = 'soft_deleted' AND OLD.table_status IS DISTINCT FROM NEW.table_status THEN
        change_kind := 'remove';
        reason_code := 'soft_deleted';
    ELSIF OLD.display_name IS DISTINCT FROM NEW.display_name THEN
        change_kind := 'invalidate';
        reason_code := 'renamed';
    ELSE
        RETURN NEW;
    END IF;
    payload := jsonb_build_object(
        'extension_profile_id', 'network_flow_activity',
        'resource_kind', 'network_flow_table',
        'resource_id', NEW.network_flow_table_id,
        'change_kind', change_kind,
        'reason_code', reason_code,
        'workspace_refs', jsonb_build_array(jsonb_build_object(
            'kind', 'extension_workspace',
            'extension_profile_id', 'network_flow_activity',
            'workspace_key', 'network_analysis'
        ))
    );
    deterministic_key := 'extension_resource_changed:network_flow_table:'
        || NEW.network_flow_table_id || ':' || NEW.table_version::text || ':' || reason_code;
    INSERT INTO public.collaboration_event_intents (
        intent_key, incident_id, event_family, canonical_payload, source_identity,
        mutation_ordinal, next_attempt_at, created_at, updated_at
    ) VALUES (
        deterministic_key, NEW.incident_id, 'extension_resource_changed', payload,
        'network_flow_table:' || NEW.network_flow_table_id, 0,
        NEW.updated_at, NEW.updated_at, NEW.updated_at
    )
    ON CONFLICT (intent_key) DO NOTHING;
    RETURN NEW;
END;
$$;
-- +goose StatementEnd

CREATE TRIGGER network_flow_tables_collaboration_resource_intent
    AFTER UPDATE ON public.network_flow_tables
    FOR EACH ROW
    EXECUTE FUNCTION public.collaboration_capture_network_flow_resource_v1();
