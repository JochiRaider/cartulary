-- +goose Up
-- Party v2 snapshots retain the Records envelope as the single clock authority.
-- Retained v1 source timestamps cannot be silently reinterpreted or discarded.
-- +goose StatementBegin
DO $$ BEGIN
    LOCK TABLE public.parties, public.records, public.change_set_mutations,
        public.record_revisions IN SHARE MODE;
    IF EXISTS (SELECT 1 FROM public.parties)
       OR EXISTS (SELECT 1 FROM public.records WHERE record_type='party')
       OR EXISTS (SELECT 1 FROM public.change_set_mutations
           WHERE before_value->>'snapshot_schema_id'='cartulary.revisions.snapshot.party.v1'
              OR after_value->>'snapshot_schema_id'='cartulary.revisions.snapshot.party.v1')
       OR EXISTS (SELECT 1 FROM public.record_revisions
           WHERE before_json->>'snapshot_schema_id'='cartulary.revisions.snapshot.party.v1'
              OR after_json->>'snapshot_schema_id'='cartulary.revisions.snapshot.party.v1') THEN
        RAISE EXCEPTION 'party_clock_cutover_required: preserve the matching application and data; initialize a fresh disposable environment explicitly' USING ERRCODE='55000';
    END IF;
END $$;
-- +goose StatementEnd
ALTER TABLE public.parties DROP COLUMN created_at, DROP COLUMN updated_at;

-- +goose Down
-- +goose StatementBegin
DO $$ BEGIN
    RAISE EXCEPTION 'party clock authority cannot be downgraded; restore a matching application/data backup' USING ERRCODE='55000';
END $$;
-- +goose StatementEnd
