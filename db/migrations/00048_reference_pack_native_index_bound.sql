-- +goose Up
ALTER TABLE public.reference_pack_indexes DROP CONSTRAINT rp_indexes_ck3;
ALTER TABLE public.reference_pack_indexes ADD CONSTRAINT rp_indexes_ck3 CHECK (
    octet_length(canonical_item) >= 1 AND octet_length(canonical_item) <=
    CASE WHEN convert_from(canonical_item, 'UTF8')::jsonb->>'schema_id' IN
      ('cartulary.reference_pack.framework.attack.object.v2',
       'cartulary.reference_pack.framework.attack.relationship.v2')
    THEN 4194303 ELSE 1048576 END
);

-- +goose Down
ALTER TABLE public.reference_pack_indexes DROP CONSTRAINT rp_indexes_ck3;
ALTER TABLE public.reference_pack_indexes ADD CONSTRAINT rp_indexes_ck3 CHECK
    (octet_length(canonical_item) >= 1 AND octet_length(canonical_item) <= 1048576);
