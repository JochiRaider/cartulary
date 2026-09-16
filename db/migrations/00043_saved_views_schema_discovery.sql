-- +goose Up
CREATE INDEX saved_views_incident_schema_order_idx
    ON public.saved_views (incident_id, view_schema_id, updated_at DESC, saved_view_id);

-- +goose Down
DROP INDEX public.saved_views_incident_schema_order_idx;
