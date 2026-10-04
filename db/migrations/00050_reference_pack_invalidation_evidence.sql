-- +goose Up
-- This fact is written only by the semantic finalizer, alongside the first
-- invalidation of established content. Preparation and operational aborts
-- cannot assert it. Keeping it with the attempt avoids a cohort-sized buffer.
ALTER TABLE public.reference_pack_attempt_members
    ADD COLUMN invalidated_content boolean NOT NULL DEFAULT false,
    ADD CONSTRAINT rp_attempt_members_invalidation_ck CHECK (
        NOT invalidated_content OR (verdict='content_rejected' AND pack_key IS NOT NULL)
    );

-- Prepared verification facts remain immutable. Only the finalizer's new
-- publication fact can advance once, while its owning attempt is nonterminal.
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
DROP TRIGGER reference_pack_attempt_members_immutable ON public.reference_pack_attempt_members;
CREATE TRIGGER reference_pack_attempt_members_guard BEFORE UPDATE OR DELETE ON public.reference_pack_attempt_members FOR EACH ROW EXECUTE FUNCTION public.reference_pack_attempt_member_guard();
GRANT UPDATE(invalidated_content) ON public.reference_pack_attempt_members TO cartulary_runtime;

-- +goose Down
-- Retain the distinction between verification evidence and a committed
-- invalidation. A downgrade cannot silently erase that historical fact.
-- +goose StatementBegin
DO $$ BEGIN
    IF EXISTS (SELECT 1 FROM public.reference_pack_attempt_members WHERE invalidated_content) THEN
        RAISE EXCEPTION 'reference pack invalidation evidence cannot be discarded; restore a complete matching backup' USING ERRCODE='55000';
    END IF;
END $$;
-- +goose StatementEnd
REVOKE UPDATE(invalidated_content) ON public.reference_pack_attempt_members FROM cartulary_runtime;
DROP TRIGGER reference_pack_attempt_members_guard ON public.reference_pack_attempt_members;
DROP FUNCTION public.reference_pack_attempt_member_guard();
CREATE TRIGGER reference_pack_attempt_members_immutable BEFORE UPDATE OR DELETE ON public.reference_pack_attempt_members FOR EACH ROW EXECUTE FUNCTION public.reference_pack_immutable_row();
ALTER TABLE public.reference_pack_attempt_members
    DROP CONSTRAINT rp_attempt_members_invalidation_ck,
    DROP COLUMN invalidated_content;
