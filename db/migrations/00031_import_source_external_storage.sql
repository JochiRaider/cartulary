-- +goose Up
-- Imports reads source byte ranges without decompressing whole stored sources.
ALTER TABLE public.import_source_streams ALTER COLUMN source_bytes SET STORAGE EXTERNAL;
-- This unpublished pre-production revision requires a fresh environment.

-- +goose Down
-- Policy reversal does not change bytes or force compression of existing data.
ALTER TABLE public.import_source_streams ALTER COLUMN source_bytes SET STORAGE EXTENDED;
