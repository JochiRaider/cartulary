-- +goose Up
-- +goose StatementBegin
DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM route_idempotency
        WHERE actor_user_id IS NULL
    ) THEN
        RAISE EXCEPTION 'route_idempotency.actor_user_id must be populated before actor-scoped idempotency can be enforced';
    END IF;
END $$;
-- +goose StatementEnd

ALTER TABLE route_idempotency
    ALTER COLUMN actor_user_id SET NOT NULL;

ALTER TABLE route_idempotency
    DROP CONSTRAINT IF EXISTS route_idempotency_route_key_scope_key_client_txn_id_key;

CREATE UNIQUE INDEX IF NOT EXISTS route_idempotency_route_actor_scope_client_txn_idx
    ON route_idempotency (route_key, actor_user_id, scope_key, client_txn_id);

-- +goose Down
-- +goose StatementBegin
DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM route_idempotency
        GROUP BY route_key, scope_key, client_txn_id
        HAVING COUNT(*) > 1
    ) THEN
        RAISE EXCEPTION 'cannot restore route_idempotency legacy uniqueness: cross-actor duplicate idempotency keys exist';
    END IF;
END $$;
-- +goose StatementEnd

DROP INDEX IF EXISTS route_idempotency_route_actor_scope_client_txn_idx;

ALTER TABLE route_idempotency
    ADD CONSTRAINT route_idempotency_route_key_scope_key_client_txn_id_key
    UNIQUE (route_key, scope_key, client_txn_id);

ALTER TABLE route_idempotency
    ALTER COLUMN actor_user_id DROP NOT NULL;
