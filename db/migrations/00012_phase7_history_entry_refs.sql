-- +goose Up
CREATE TABLE IF NOT EXISTS record_history_entry_refs (
    history_entry_ref text PRIMARY KEY,
    record_id uuid NOT NULL REFERENCES records (record_id) ON DELETE CASCADE,
    change_set_id uuid NOT NULL REFERENCES change_sets (change_set_id) ON DELETE CASCADE,
    mutation_sequence_no integer NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (record_id, change_set_id, mutation_sequence_no),
    FOREIGN KEY (change_set_id, mutation_sequence_no)
        REFERENCES change_set_mutations (change_set_id, sequence_no)
        ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS record_history_entry_refs_record_lookup_idx
    ON record_history_entry_refs (record_id, change_set_id, mutation_sequence_no);

-- +goose Down
DROP INDEX IF EXISTS record_history_entry_refs_record_lookup_idx;
DROP TABLE IF EXISTS record_history_entry_refs;
