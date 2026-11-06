-- +goose Up
-- Incident Bundles owns historical attribution. Source owners declare the exact
-- mutable attribution tuples below. Effective native mutations retire only the
-- affected override in their transaction; immutable creation/history is retained.
-- No source-table query, native-account mutation, or historical backfill occurs.

-- +goose StatementBegin
CREATE FUNCTION public.incident_bundle_retire_mutated_attribution() RETURNS trigger
LANGUAGE plpgsql SET search_path = pg_catalog, public AS $$
DECLARE
    before_row jsonb := to_jsonb(OLD);
    after_row jsonb := to_jsonb(NEW);
    changed boolean := false;
    position integer;
    scoped_incident uuid;
BEGIN
    IF TG_OP <> 'UPDATE' OR TG_NARGS < 3 THEN
        RAISE EXCEPTION 'invalid historical attribution mutation binding';
    END IF;
    FOR position IN 2..TG_NARGS-1 LOOP
        IF NOT (before_row ? TG_ARGV[position]) OR NOT (after_row ? TG_ARGV[position]) THEN
            RAISE EXCEPTION 'historical attribution mutation field absent';
        END IF;
        changed := changed OR (before_row -> TG_ARGV[position]) IS DISTINCT FROM (after_row -> TG_ARGV[position]);
    END LOOP;
    IF changed THEN
        IF before_row ->> TG_ARGV[1] IS NULL OR (TG_ARGV[0] <> '' AND before_row ->> TG_ARGV[0] IS NULL) THEN
            RAISE EXCEPTION 'historical attribution mutation identity absent';
        END IF;
        scoped_incident := CASE WHEN TG_ARGV[0] <> '' THEN (before_row ->> TG_ARGV[0])::uuid ELSE NULL END;
        DELETE FROM public.incident_bundle_imported_attributions
        WHERE (scoped_incident IS NULL OR incident_id = scoped_incident)
          AND source_table = TG_TABLE_NAME
          AND source_row_id = before_row ->> TG_ARGV[1]
          AND source_column = TG_ARGV[2];
    END IF;
    RETURN NEW;
END;
$$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION public.incident_bundle_retire_mutated_attribution() FROM PUBLIC;

CREATE TRIGGER incident_bundle_attribution_mutation_01
AFTER UPDATE OF updated_by_user_id, incident_version, updated_at ON public.incidents
FOR EACH ROW EXECUTE FUNCTION public.incident_bundle_retire_mutated_attribution('id', 'id', 'updated_by_user_id', 'incident_version', 'updated_at');

CREATE TRIGGER incident_bundle_attribution_mutation_02
AFTER UPDATE OF updated_by_user_id, row_version, updated_at ON public.records
FOR EACH ROW EXECUTE FUNCTION public.incident_bundle_retire_mutated_attribution('incident_id', 'record_id', 'updated_by_user_id', 'row_version', 'updated_at');

CREATE TRIGGER incident_bundle_attribution_mutation_03
AFTER UPDATE OF deleted_by_user_id, deleted_at ON public.records
FOR EACH ROW EXECUTE FUNCTION public.incident_bundle_retire_mutated_attribution('incident_id', 'record_id', 'deleted_by_user_id', 'deleted_at');

CREATE TRIGGER incident_bundle_attribution_mutation_04
AFTER UPDATE OF updated_by_user_id, row_version, updated_at ON public.hosts
FOR EACH ROW EXECUTE FUNCTION public.incident_bundle_retire_mutated_attribution('incident_id', 'record_id', 'updated_by_user_id', 'row_version', 'updated_at');

CREATE TRIGGER incident_bundle_attribution_mutation_05
AFTER UPDATE OF updated_by_user_id, row_version, updated_at ON public.identities
FOR EACH ROW EXECUTE FUNCTION public.incident_bundle_retire_mutated_attribution('incident_id', 'record_id', 'updated_by_user_id', 'row_version', 'updated_at');

CREATE TRIGGER incident_bundle_attribution_mutation_06
AFTER UPDATE OF updated_by_user_id, profile_version, updated_at ON public.timeline_time_conversion_profiles
FOR EACH ROW EXECUTE FUNCTION public.incident_bundle_retire_mutated_attribution('incident_id', 'incident_id', 'updated_by_user_id', 'profile_version', 'updated_at');

CREATE TRIGGER incident_bundle_attribution_mutation_07
AFTER UPDATE OF reviewed_by_user_id, reviewed_at ON public.timeline_events
FOR EACH ROW EXECUTE FUNCTION public.incident_bundle_retire_mutated_attribution('incident_id', 'record_id', 'reviewed_by_user_id', 'reviewed_at');

CREATE TRIGGER incident_bundle_attribution_mutation_08
AFTER UPDATE OF superseded_by_user_id, superseded_at ON public.timeline_events
FOR EACH ROW EXECUTE FUNCTION public.incident_bundle_retire_mutated_attribution('incident_id', 'record_id', 'superseded_by_user_id', 'superseded_at');

-- Entity mention identity is a global primary key; the row has no incident column.
-- Its declared relation/identity still names one source row across all incidents.
CREATE TRIGGER incident_bundle_attribution_mutation_09
AFTER UPDATE OF resolved_by_user_id, resolved_at ON public.entity_mentions
FOR EACH ROW EXECUTE FUNCTION public.incident_bundle_retire_mutated_attribution('', 'entity_mention_id', 'resolved_by_user_id', 'resolved_at');

CREATE TRIGGER incident_bundle_attribution_mutation_10
AFTER UPDATE OF resolved_by_user_id, resolved_at ON public.indicator_observations
FOR EACH ROW EXECUTE FUNCTION public.incident_bundle_retire_mutated_attribution('incident_id', 'indicator_observation_id', 'resolved_by_user_id', 'resolved_at');

CREATE TRIGGER incident_bundle_attribution_mutation_11
AFTER UPDATE OF deleted_by_user_id, deleted_at ON public.indicator_observations
FOR EACH ROW EXECUTE FUNCTION public.incident_bundle_retire_mutated_attribution('incident_id', 'indicator_observation_id', 'deleted_by_user_id', 'deleted_at');

CREATE TRIGGER incident_bundle_attribution_mutation_12
AFTER UPDATE OF deleted_by_user_id, deleted_at ON public.indicator_state_intervals
FOR EACH ROW EXECUTE FUNCTION public.incident_bundle_retire_mutated_attribution('incident_id', 'indicator_state_interval_id', 'deleted_by_user_id', 'deleted_at');

CREATE TRIGGER incident_bundle_attribution_mutation_13
AFTER UPDATE OF assessor_user_id, assessed_at ON public.assessments
FOR EACH ROW EXECUTE FUNCTION public.incident_bundle_retire_mutated_attribution('incident_id', 'record_id', 'assessor_user_id', 'assessed_at');

CREATE TRIGGER incident_bundle_attribution_mutation_14
AFTER UPDATE OF deleted_by_user_id, deleted_at ON public.assessments
FOR EACH ROW EXECUTE FUNCTION public.incident_bundle_retire_mutated_attribution('incident_id', 'record_id', 'deleted_by_user_id', 'deleted_at');

CREATE TRIGGER incident_bundle_attribution_mutation_15
AFTER UPDATE OF owner_user_id, decided_at ON public.record_links
FOR EACH ROW EXECUTE FUNCTION public.incident_bundle_retire_mutated_attribution('incident_id', 'record_link_id', 'owner_user_id', 'decided_at');

CREATE TRIGGER incident_bundle_attribution_mutation_16
AFTER UPDATE OF deleted_by_user_id, deleted_at ON public.record_links
FOR EACH ROW EXECUTE FUNCTION public.incident_bundle_retire_mutated_attribution('incident_id', 'record_link_id', 'deleted_by_user_id', 'deleted_at');

CREATE TRIGGER incident_bundle_attribution_mutation_17
AFTER UPDATE OF deleted_by_user_id, deleted_at ON public.record_tags
FOR EACH ROW EXECUTE FUNCTION public.incident_bundle_retire_mutated_attribution('incident_id', 'record_tag_id', 'deleted_by_user_id', 'deleted_at');

CREATE TRIGGER incident_bundle_attribution_mutation_18
AFTER UPDATE OF owner_user_id ON public.task_requests
FOR EACH ROW EXECUTE FUNCTION public.incident_bundle_retire_mutated_attribution('incident_id', 'record_id', 'owner_user_id');

CREATE TRIGGER incident_bundle_attribution_mutation_19
AFTER UPDATE OF owner_user_id ON public.decisions
FOR EACH ROW EXECUTE FUNCTION public.incident_bundle_retire_mutated_attribution('incident_id', 'record_id', 'owner_user_id');

CREATE TRIGGER incident_bundle_attribution_mutation_20
AFTER UPDATE OF owner_user_id ON public.saved_views
FOR EACH ROW EXECUTE FUNCTION public.incident_bundle_retire_mutated_attribution('incident_id', 'saved_view_id', 'owner_user_id');

-- +goose Down
-- +goose StatementBegin
DO $$ BEGIN
    RAISE EXCEPTION 'incident portability requires fresh state and has no in-place downgrade; retain the prior executable with its own data';
END $$;
-- +goose StatementEnd
