-- +goose Up
DELETE FROM route_idempotency old
USING route_idempotency newer
WHERE old.route_key = 'entities.indicators.rows.create'
  AND newer.route_key = 'indicators.rows.create'
  AND old.actor_user_id = newer.actor_user_id
  AND old.scope_key = newer.scope_key
  AND old.client_txn_id = newer.client_txn_id;

UPDATE route_idempotency
   SET route_key = 'indicators.rows.create'
 WHERE route_key = 'entities.indicators.rows.create';

UPDATE change_sets
   SET source = 'indicators.rows.create'
 WHERE source = 'entities.indicators.rows.create';

UPDATE change_sets
   SET source = 'indicators.observations.capture'
 WHERE source = 'entities.indicators.observations.capture';

UPDATE change_sets
   SET source = 'indicators.observations.resolve'
 WHERE source = 'entities.indicators.observations.resolve';

UPDATE change_sets
   SET source = 'indicators.lifecycle.append'
 WHERE source = 'entities.indicators.lifecycle.append';

UPDATE indicator_observations
   SET resolution_method = 'indicators.observations.resolve'
 WHERE resolution_method = 'entities.indicators.observations.resolve';

-- +goose Down
DELETE FROM route_idempotency old
USING route_idempotency newer
WHERE old.route_key = 'indicators.rows.create'
  AND newer.route_key = 'entities.indicators.rows.create'
  AND old.actor_user_id = newer.actor_user_id
  AND old.scope_key = newer.scope_key
  AND old.client_txn_id = newer.client_txn_id;

UPDATE route_idempotency
   SET route_key = 'entities.indicators.rows.create'
 WHERE route_key = 'indicators.rows.create';

UPDATE change_sets
   SET source = 'entities.indicators.rows.create'
 WHERE source = 'indicators.rows.create';

UPDATE change_sets
   SET source = 'entities.indicators.observations.capture'
 WHERE source = 'indicators.observations.capture';

UPDATE change_sets
   SET source = 'entities.indicators.observations.resolve'
 WHERE source = 'indicators.observations.resolve';

UPDATE change_sets
   SET source = 'entities.indicators.lifecycle.append'
 WHERE source = 'indicators.lifecycle.append';

UPDATE indicator_observations
   SET resolution_method = 'entities.indicators.observations.resolve'
 WHERE resolution_method = 'indicators.observations.resolve';
