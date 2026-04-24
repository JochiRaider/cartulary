-- +goose Up
ALTER TABLE hosts
    DROP CONSTRAINT IF EXISTS hosts_entity_origin_check,
    DROP CONSTRAINT IF EXISTS hosts_entity_origin_core02_ck;

ALTER TABLE identities
    DROP CONSTRAINT IF EXISTS identities_entity_origin_check,
    DROP CONSTRAINT IF EXISTS identities_entity_origin_core02_ck;

UPDATE hosts
   SET entity_origin = 'entity_sheet'
 WHERE entity_origin = 'direct_create';

UPDATE identities
   SET entity_origin = 'entity_sheet'
 WHERE entity_origin = 'direct_create';

ALTER TABLE hosts
    ALTER COLUMN entity_origin SET DEFAULT 'entity_sheet',
    ADD CONSTRAINT hosts_entity_origin_core02_ck CHECK (entity_origin IN ('entity_sheet', 'entity_import', 'created_from_mention', 'system_upsert'));

ALTER TABLE identities
    ALTER COLUMN entity_origin SET DEFAULT 'entity_sheet',
    ADD CONSTRAINT identities_entity_origin_core02_ck CHECK (entity_origin IN ('entity_sheet', 'entity_import', 'created_from_mention', 'system_upsert'));

-- +goose Down
ALTER TABLE hosts
    DROP CONSTRAINT IF EXISTS hosts_entity_origin_core02_ck,
    DROP CONSTRAINT IF EXISTS hosts_entity_origin_check;

ALTER TABLE identities
    DROP CONSTRAINT IF EXISTS identities_entity_origin_core02_ck,
    DROP CONSTRAINT IF EXISTS identities_entity_origin_check;

ALTER TABLE hosts
    ALTER COLUMN entity_origin SET DEFAULT 'entity_sheet',
    ADD CONSTRAINT hosts_entity_origin_check CHECK (entity_origin IN ('entity_sheet', 'entity_import', 'created_from_mention', 'system_upsert'));

ALTER TABLE identities
    ALTER COLUMN entity_origin SET DEFAULT 'entity_sheet',
    ADD CONSTRAINT identities_entity_origin_check CHECK (entity_origin IN ('entity_sheet', 'entity_import', 'created_from_mention', 'system_upsert'));
