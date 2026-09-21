-- +goose Up
CREATE INDEX record_links_active_null_field_outgoing_page_idx
    ON public.record_links (incident_id, src_record_id, link_type, created_at DESC, record_link_id DESC)
    WHERE deleted_at IS NULL AND field_key IS NULL;
CREATE INDEX record_links_active_null_field_incoming_page_idx
    ON public.record_links (incident_id, dst_record_id, link_type, created_at DESC, record_link_id DESC)
    WHERE deleted_at IS NULL AND field_key IS NULL;

-- +goose Down
DROP INDEX public.record_links_active_null_field_incoming_page_idx;
DROP INDEX public.record_links_active_null_field_outgoing_page_idx;
