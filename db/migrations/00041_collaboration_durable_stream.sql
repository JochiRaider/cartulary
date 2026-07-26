-- +goose Up
CREATE TABLE public.collaboration_event_intents (
    intent_id uuid DEFAULT gen_random_uuid() PRIMARY KEY,
    intent_key text NOT NULL UNIQUE,
    incident_id uuid NOT NULL REFERENCES public.incidents(id) ON DELETE CASCADE,
    event_family text NOT NULL,
    canonical_payload jsonb NOT NULL,
    source_change_set_id uuid,
    source_record_id uuid,
    source_row_version bigint,
    source_identity text NOT NULL,
    mutation_ordinal integer NOT NULL,
    dispatch_state text NOT NULL DEFAULT 'pending',
    sequenced_event_id uuid,
    attempt_count integer NOT NULL DEFAULT 0,
    next_attempt_at timestamp with time zone NOT NULL,
    lease_owner uuid,
    lease_expires_at timestamp with time zone,
    sequenced_at timestamp with time zone,
    delivered_at timestamp with time zone,
    last_error_code text,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    CONSTRAINT collaboration_event_intents_key_ck
        CHECK (octet_length(intent_key) BETWEEN 1 AND 512 AND intent_key !~ '[[:cntrl:]]'),
    CONSTRAINT collaboration_event_intents_family_ck
        CHECK (event_family IN ('record_changed', 'job_progress', 'extension_resource_changed')),
    CONSTRAINT collaboration_event_intents_payload_ck
        CHECK (jsonb_typeof(canonical_payload) = 'object'),
    CONSTRAINT collaboration_event_intents_source_identity_ck
        CHECK (octet_length(source_identity) BETWEEN 1 AND 512 AND source_identity !~ '[[:cntrl:]]'),
    CONSTRAINT collaboration_event_intents_row_version_ck
        CHECK (source_row_version IS NULL OR source_row_version >= 1),
    CONSTRAINT collaboration_event_intents_ordinal_ck
        CHECK (mutation_ordinal >= 0),
    CONSTRAINT collaboration_event_intents_dispatch_state_ck
        CHECK (dispatch_state IN ('pending', 'sequenced')),
    CONSTRAINT collaboration_event_intents_attempt_ck
        CHECK (attempt_count BETWEEN 0 AND 2147483647),
    CONSTRAINT collaboration_event_intents_lease_ck
        CHECK ((lease_owner IS NULL) = (lease_expires_at IS NULL)),
    CONSTRAINT collaboration_event_intents_sequence_ck
        CHECK (
            (dispatch_state = 'pending' AND sequenced_event_id IS NULL AND sequenced_at IS NULL AND delivered_at IS NULL)
            OR
            (dispatch_state = 'sequenced' AND sequenced_event_id IS NOT NULL AND sequenced_at IS NOT NULL)
        ),
    CONSTRAINT collaboration_event_intents_delivery_ck
        CHECK (delivered_at IS NULL OR delivered_at >= sequenced_at),
    CONSTRAINT collaboration_event_intents_time_ck
        CHECK (updated_at >= created_at AND next_attempt_at >= created_at),
    CONSTRAINT collaboration_event_intents_error_ck
        CHECK (last_error_code IS NULL OR last_error_code ~ '^[a-z][a-z0-9_.-]{0,127}$')
);

CREATE INDEX idx_collaboration_event_intents_dispatch
    ON public.collaboration_event_intents (
        dispatch_state,
        next_attempt_at,
        created_at,
        intent_id
    )
    WHERE delivered_at IS NULL;

CREATE INDEX idx_collaboration_event_intents_lease
    ON public.collaboration_event_intents (lease_expires_at, intent_id)
    WHERE lease_owner IS NOT NULL;

CREATE TABLE public.collaboration_incident_stream_cursors (
    incident_id uuid PRIMARY KEY REFERENCES public.incidents(id) ON DELETE CASCADE,
    high_water_stream_seq bigint NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    CONSTRAINT collaboration_incident_stream_cursors_high_water_ck
        CHECK (high_water_stream_seq >= 0)
);

CREATE TABLE public.collaboration_replay_events (
    event_id uuid PRIMARY KEY,
    incident_id uuid NOT NULL REFERENCES public.incidents(id) ON DELETE CASCADE,
    stream_seq bigint NOT NULL,
    intent_key text NOT NULL UNIQUE,
    event_family text NOT NULL,
    canonical_payload jsonb NOT NULL,
    emitted_at timestamp with time zone NOT NULL,
    CONSTRAINT collaboration_replay_events_incident_sequence_uq
        UNIQUE (incident_id, stream_seq),
    CONSTRAINT collaboration_replay_events_sequence_ck
        CHECK (stream_seq >= 1),
    CONSTRAINT collaboration_replay_events_family_ck
        CHECK (event_family IN ('record_changed', 'job_progress', 'extension_resource_changed')),
    CONSTRAINT collaboration_replay_events_payload_ck
        CHECK (jsonb_typeof(canonical_payload) = 'object')
);

CREATE INDEX idx_collaboration_replay_events_retention
    ON public.collaboration_replay_events (incident_id, emitted_at, stream_seq);

CREATE TABLE public.collaboration_resume_tokens (
    token_hash bytea PRIMARY KEY,
    session_id uuid NOT NULL REFERENCES public.user_sessions(id) ON DELETE CASCADE,
    incident_id uuid NOT NULL REFERENCES public.incidents(id) ON DELETE CASCADE,
    client_instance_id text NOT NULL,
    issued_at timestamp with time zone NOT NULL,
    expires_at timestamp with time zone NOT NULL,
    CONSTRAINT collaboration_resume_tokens_hash_ck
        CHECK (octet_length(token_hash) = 32),
    CONSTRAINT collaboration_resume_tokens_client_ck
        CHECK (
            octet_length(client_instance_id) BETWEEN 1 AND 256
            AND client_instance_id !~ '[[:cntrl:]]'
        ),
    CONSTRAINT collaboration_resume_tokens_expiry_ck
        CHECK (expires_at > issued_at)
);

CREATE INDEX idx_collaboration_resume_tokens_expiry
    ON public.collaboration_resume_tokens (expires_at);

-- Record revisions are the shared, transaction-bound record mutation
-- substrate. This trigger is the durable fallback for every record family;
-- source owners may enrich the same deterministic intent later in the same
-- transaction without creating a second event.
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
    -- Incident Bundle restore replays historical source rows. Those rows must
    -- not be backfilled into the live collaboration stream.
    IF current_setting('cartulary.collaboration.suppress_historical_intents', true) = 'on' THEN
        RETURN NEW;
    END IF;

    SELECT *
      INTO STRICT record_row
      FROM public.records
     WHERE record_id = NEW.record_id;
    SELECT *
      INTO STRICT change_set_row
      FROM public.change_sets
     WHERE change_set_id = NEW.change_set_id;

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
    affected_view := jsonb_build_object(
        'view_schema_id', view_schema_id,
        'change_kind', change_kind
    );
    -- Ordinary owner mutations already provide the canonical public row to
    -- Revisions. Preserve the established sparse-patch delivery contract from
    -- that row. Restore and rollback intentionally remain invalidations because
    -- their source reconstruction can affect owner state beyond a cell delta.
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
        intent_key,
        incident_id,
        event_family,
        canonical_payload,
        source_change_set_id,
        source_record_id,
        source_row_version,
        source_identity,
        mutation_ordinal,
        next_attempt_at,
        created_at,
        updated_at
    ) VALUES (
        'record_changed:' || NEW.change_set_id::text || ':' || NEW.record_id::text || ':' || NEW.row_version::text,
        record_row.incident_id,
        'record_changed',
        payload,
        NEW.change_set_id,
        NEW.record_id,
        NEW.row_version,
        NEW.change_set_id::text || ':' || NEW.record_id::text,
        ordinal,
        change_set_row.created_at,
        change_set_row.created_at,
        change_set_row.created_at
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

-- Replayable job progress is captured at the source-table boundary so every
-- current and future incident-scoped job writer participates in the same
-- authoritative transaction. Deployment-scoped jobs are intentionally not
-- admitted to incident streams.
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
        'scope', jsonb_build_object(
            'kind', 'incident',
            'incident_id', NEW.incident_id::text
        ),
        'status', NEW.status,
        'progress', jsonb_build_object(
            'completed', NEW.progress_completed,
            'total', to_jsonb(NEW.progress_total)
        ),
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
        payload := payload || jsonb_build_object(
            'retained_until',
            public.collaboration_rfc3339_v1(NEW.retained_until)
        );
    END IF;

    deterministic_key := 'job_progress:' || NEW.job_id::text || ':'
        || encode(digest(payload::text, 'sha256'), 'hex');
    INSERT INTO public.collaboration_event_intents (
        intent_key,
        incident_id,
        event_family,
        canonical_payload,
        source_identity,
        mutation_ordinal,
        next_attempt_at,
        created_at,
        updated_at
    ) VALUES (
        deterministic_key,
        NEW.incident_id,
        'job_progress',
        payload,
        'job:' || NEW.job_id::text,
        0,
        NEW.updated_at,
        NEW.updated_at,
        NEW.updated_at
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

-- Network Flow is the currently admitted extension-resource producer. Capture
-- invalidation/removal in its source transaction and never persist connection
-- or subscriber data in the intent.
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
        intent_key,
        incident_id,
        event_family,
        canonical_payload,
        source_identity,
        mutation_ordinal,
        next_attempt_at,
        created_at,
        updated_at
    ) VALUES (
        deterministic_key,
        NEW.incident_id,
        'extension_resource_changed',
        payload,
        'network_flow_table:' || NEW.network_flow_table_id,
        0,
        NEW.updated_at,
        NEW.updated_at,
        NEW.updated_at
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

-- +goose Down
DROP TRIGGER network_flow_tables_collaboration_resource_intent ON public.network_flow_tables;
DROP FUNCTION public.collaboration_capture_network_flow_resource_v1();
DROP TRIGGER jobs_collaboration_progress_intent ON public.jobs;
DROP FUNCTION public.collaboration_capture_job_progress_v1();
DROP FUNCTION public.collaboration_rfc3339_v1(timestamp with time zone);
DROP TRIGGER record_revisions_collaboration_event_intent ON public.record_revisions;
DROP FUNCTION public.collaboration_capture_record_revision_v1();
DROP TABLE public.collaboration_resume_tokens;
DROP TABLE public.collaboration_replay_events;
DROP TABLE public.collaboration_incident_stream_cursors;
DROP TABLE public.collaboration_event_intents;
