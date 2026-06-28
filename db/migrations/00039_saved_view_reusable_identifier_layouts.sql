-- +goose Up
WITH additive_fields(view_schema_id, field_key) AS (
    VALUES
        ('cartulary.view.hosts.v1', 'host.reusable_identifiers'),
        ('cartulary.view.identities.v1', 'identity.reusable_identifiers')
)
UPDATE saved_views AS sv
   SET layout_json = jsonb_set(
       jsonb_set(
           sv.layout_json,
           '{column_order}',
           CASE
               WHEN sv.layout_json->'column_order' ? additive_fields.field_key THEN sv.layout_json->'column_order'
               ELSE (sv.layout_json->'column_order') || jsonb_build_array(additive_fields.field_key)
           END,
           false
       ),
       '{hidden_field_keys}',
       (
           SELECT COALESCE(jsonb_agg(key_value ORDER BY key_value), '[]'::jsonb)
             FROM (
                 SELECT DISTINCT key_value
                   FROM jsonb_array_elements_text(
                       CASE
                           WHEN sv.layout_json->'hidden_field_keys' ? additive_fields.field_key THEN sv.layout_json->'hidden_field_keys'
                           ELSE (sv.layout_json->'hidden_field_keys') || jsonb_build_array(additive_fields.field_key)
                       END
                   ) AS existing_keys(key_value)
             ) AS distinct_keys
       ),
       false
   )
  FROM additive_fields
 WHERE sv.view_schema_id = additive_fields.view_schema_id
   AND sv.layout_json->>'layout_schema_id' = 'cartulary.layout.v1'
   AND jsonb_typeof(sv.layout_json->'column_order') = 'array'
   AND jsonb_typeof(sv.layout_json->'hidden_field_keys') = 'array'
   AND (
       NOT (sv.layout_json->'column_order' ? additive_fields.field_key)
       OR NOT (sv.layout_json->'hidden_field_keys' ? additive_fields.field_key)
   );

-- +goose Down
WITH additive_fields(view_schema_id, field_key) AS (
    VALUES
        ('cartulary.view.hosts.v1', 'host.reusable_identifiers'),
        ('cartulary.view.identities.v1', 'identity.reusable_identifiers')
)
UPDATE saved_views AS sv
   SET layout_json = jsonb_set(
       jsonb_set(
           sv.layout_json,
           '{column_order}',
           (
               SELECT COALESCE(jsonb_agg(key_value ORDER BY key_ordinal), '[]'::jsonb)
                 FROM jsonb_array_elements_text(sv.layout_json->'column_order') WITH ORDINALITY AS existing_keys(key_value, key_ordinal)
                WHERE key_value <> additive_fields.field_key
           ),
           false
       ),
       '{hidden_field_keys}',
       (
           SELECT COALESCE(jsonb_agg(key_value ORDER BY key_value), '[]'::jsonb)
             FROM jsonb_array_elements_text(sv.layout_json->'hidden_field_keys') AS existing_keys(key_value)
            WHERE key_value <> additive_fields.field_key
       ),
       false
   )
  FROM additive_fields
 WHERE sv.view_schema_id = additive_fields.view_schema_id
   AND sv.layout_json->>'layout_schema_id' = 'cartulary.layout.v1'
   AND jsonb_typeof(sv.layout_json->'column_order') = 'array'
   AND jsonb_typeof(sv.layout_json->'hidden_field_keys') = 'array';
