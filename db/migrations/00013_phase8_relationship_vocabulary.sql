-- +goose Up

ALTER TABLE record_links
    DROP CONSTRAINT IF EXISTS record_links_link_type_check;

ALTER TABLE record_links
    ADD CONSTRAINT record_links_link_type_check CHECK (link_type IN (
        'observed_on_host',
        'observed_as_identity',
        'references_indicator',
        'attached_evidence',
        'references_artifact',
        'derived_from',
        'merged_into',
        'supported_by',
        'references_record',
        'supersedes'
    ));

-- +goose Down

ALTER TABLE record_links
    DROP CONSTRAINT IF EXISTS record_links_link_type_check;

ALTER TABLE record_links
    ADD CONSTRAINT record_links_link_type_check CHECK (link_type IN (
        'supersedes',
        'observed_on_host',
        'observed_as_identity',
        'supported_by',
        'references_record',
        'attached_evidence'
    ));
