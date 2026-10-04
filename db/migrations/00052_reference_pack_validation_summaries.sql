-- +goose Up
-- No omitted historical diagnostic can be reconstructed from a reason code.
-- Reject incompatible retained attempts before changing their schema.
-- +goose StatementBegin
DO $$ DECLARE affected bigint; BEGIN
    LOCK TABLE public.reference_pack_attempt_members IN ACCESS EXCLUSIVE MODE;
    SELECT count(*) INTO affected FROM public.reference_pack_attempt_members WHERE verdict='content_rejected';
    IF affected > 0 THEN
        RAISE EXCEPTION 'reference_pack_cutover_required: historical_validation_summaries=%; preserve a matching backup and explicitly reset disposable development data', affected USING ERRCODE='55000';
    END IF;
END $$;
-- +goose StatementEnd
ALTER TABLE public.reference_pack_attempt_members
    ADD COLUMN canonical_validation_summary bytea,
    ADD COLUMN validation_summary_id text GENERATED ALWAYS AS (
        CASE WHEN canonical_validation_summary IS NULL THEN NULL ELSE 'rpvs_' || encode(public.digest(canonical_validation_summary,'sha256'),'hex') END
    ) STORED,
    ADD CONSTRAINT rp_attempt_members_summary_ck CHECK (
        (verdict='content_rejected')=(canonical_validation_summary IS NOT NULL)
        AND (canonical_validation_summary IS NULL OR octet_length(canonical_validation_summary) BETWEEN 1 AND 16777216)
    );
CREATE INDEX reference_pack_attempt_members_summary_idx ON public.reference_pack_attempt_members(validation_summary_id);

-- Generated columns are not populated in NEW during a BEFORE trigger. The
-- canonical bytes remain immutable; PostgreSQL alone derives their identity.
DROP TRIGGER reference_pack_attempt_members_guard ON public.reference_pack_attempt_members;
DROP FUNCTION public.reference_pack_attempt_member_guard();
-- +goose StatementBegin
CREATE FUNCTION public.reference_pack_attempt_member_guard() RETURNS trigger LANGUAGE plpgsql SET search_path = pg_catalog, public AS $$
DECLARE terminal timestamptz;
BEGIN
    IF TG_OP='DELETE' THEN
        RAISE EXCEPTION 'immutable reference pack preparation' USING ERRCODE='55000';
    END IF;
    IF OLD.invalidated_content OR NOT NEW.invalidated_content
       OR (to_jsonb(NEW)-'invalidated_content'-'validation_summary_id') IS DISTINCT FROM (to_jsonb(OLD)-'invalidated_content'-'validation_summary_id') THEN
        RAISE EXCEPTION 'immutable reference pack preparation' USING ERRCODE='55000';
    END IF;
    SELECT completed_at INTO STRICT terminal FROM public.reference_pack_attempts WHERE attempt_id=OLD.attempt_id FOR UPDATE;
    IF terminal IS NOT NULL THEN
        RAISE EXCEPTION 'immutable reference pack terminal effect' USING ERRCODE='55000';
    END IF;
    RETURN NEW;
END $$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION public.reference_pack_attempt_member_guard() FROM PUBLIC;
CREATE TRIGGER reference_pack_attempt_members_guard BEFORE UPDATE OR DELETE ON public.reference_pack_attempt_members FOR EACH ROW EXECUTE FUNCTION public.reference_pack_attempt_member_guard();

-- +goose Down
-- +goose StatementBegin
DO $$ BEGIN
    IF EXISTS (SELECT 1 FROM public.reference_pack_attempt_members WHERE canonical_validation_summary IS NOT NULL) THEN
        RAISE EXCEPTION 'reference pack validation evidence cannot be discarded; restore a complete matching backup' USING ERRCODE='55000';
    END IF;
END $$;
-- +goose StatementEnd
DROP INDEX public.reference_pack_attempt_members_summary_idx;
ALTER TABLE public.reference_pack_attempt_members
    DROP CONSTRAINT rp_attempt_members_summary_ck,
    DROP COLUMN validation_summary_id,
    DROP COLUMN canonical_validation_summary;
DROP TRIGGER reference_pack_attempt_members_guard ON public.reference_pack_attempt_members;
DROP FUNCTION public.reference_pack_attempt_member_guard();
-- +goose StatementBegin
CREATE FUNCTION public.reference_pack_attempt_member_guard() RETURNS trigger LANGUAGE plpgsql SET search_path = pg_catalog, public AS $$
DECLARE terminal timestamptz;
BEGIN
    IF TG_OP='DELETE' THEN
        RAISE EXCEPTION 'immutable reference pack preparation' USING ERRCODE='55000';
    END IF;
    IF OLD.invalidated_content OR NOT NEW.invalidated_content
       OR (to_jsonb(NEW)-'invalidated_content') IS DISTINCT FROM (to_jsonb(OLD)-'invalidated_content') THEN
        RAISE EXCEPTION 'immutable reference pack preparation' USING ERRCODE='55000';
    END IF;
    SELECT completed_at INTO STRICT terminal FROM public.reference_pack_attempts WHERE attempt_id=OLD.attempt_id FOR UPDATE;
    IF terminal IS NOT NULL THEN
        RAISE EXCEPTION 'immutable reference pack terminal effect' USING ERRCODE='55000';
    END IF;
    RETURN NEW;
END $$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION public.reference_pack_attempt_member_guard() FROM PUBLIC;
CREATE TRIGGER reference_pack_attempt_members_guard BEFORE UPDATE OR DELETE ON public.reference_pack_attempt_members FOR EACH ROW EXECUTE FUNCTION public.reference_pack_attempt_member_guard();
