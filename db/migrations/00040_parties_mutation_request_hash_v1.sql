-- +goose Up

-- Party request hashes before cartulary.parties.mutation_request_hash.v1 did
-- not carry an owner algorithm identity and cannot be reconstructed from the
-- retained response. Dispose only replay rows whose existing route scope
-- deterministically identifies the Party owner. This is a clean cutover: no
-- legacy hash reader or dual comparison remains.
DELETE FROM public.route_idempotency AS replay
 WHERE (
        replay.route_key = 'workbook.rows.create'
        AND replay.scope_key ~
            '^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}:cartulary\.view\.parties\.v1$'
       )
    OR (
        replay.route_key IN (
            'workbook.records.patch',
            'workbook.records.conflicts.resolve'
        )
        AND replay.scope_key ~
            '^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$'
        AND EXISTS (
            SELECT 1
              FROM public.records AS record
             WHERE record.record_id::text = replay.scope_key
               AND record.record_type = 'party'
        )
       );

-- +goose Down

-- Replay responses do not contain the original normalized request, so deleted
-- Party hash rows cannot be restored honestly. Disposable-database rollback
-- intentionally leaves the replay table unchanged.
SELECT 1;
