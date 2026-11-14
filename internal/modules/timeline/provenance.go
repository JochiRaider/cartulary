package timeline

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/timeline/internal/provenancevalue"
	"github.com/JochiRaider/cartulary/internal/modules/timeline/sourcelocking"
)

var ErrInvalidProvenance = errors.New("timeline provenance is invalid")
var ErrProvenanceIdentityCollision = fmt.Errorf("%w: identity collision", ErrInvalidProvenance)

func insertSourceProvenanceTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID, additions []ClipboardRawImportColumn) error {
	if len(additions) == 0 {
		return nil
	}
	if recordID == uuid.Nil || len(additions) > provenancevalue.MaxEntriesPerRecord {
		return fmt.Errorf("%w: entry count", ErrInvalidProvenance)
	}
	// Hold the parent through the caller's transaction so capacity preflight and
	// insertion serialize with every ordinary writer for this event.
	if err := sourcelocking.LockEventTx(ctx, tx, recordID); err != nil {
		return fmt.Errorf("lock timeline provenance target: %w", err)
	}
	var existingCount int
	if err := tx.QueryRow(ctx, `
SELECT count(*)
  FROM timeline_source_provenance
 WHERE record_id = $1
`, recordID).Scan(&existingCount); err != nil {
		return fmt.Errorf("count timeline source provenance: %w", err)
	}
	type identity struct {
		digest      [32]byte
		row, column int
	}
	seen := make(map[identity]int, len(additions))
	absent := make([]int, 0, len(additions))
	for index, addition := range additions {
		value, err := admitProvenanceAddition(addition)
		if err != nil {
			return err
		}
		key := identity{value.Digest, addition.SourceRowOrdinal, addition.SourceColumnOrdinal}
		if previous, exists := seen[key]; exists {
			// Both headers have passed text/null admission, so the interface member
			// is comparable. Every other member is retained scalar source state.
			if addition != additions[previous] {
				return ErrProvenanceIdentityCollision
			}
			continue
		}
		seen[key] = index
		var exactReplay bool
		err = tx.QueryRow(ctx, `
SELECT source_kind = $5
   AND source_metadata = $6::jsonb
   AND source_header_json = $7::jsonb
   AND raw_value = $8
   AND cell_kind IS NOT DISTINCT FROM NULLIF($9, '')
  FROM timeline_source_provenance
 WHERE record_id = $1
   AND source_identity_hash = $2
   AND source_row_ordinal = $3
   AND source_column_ordinal = $4
`, recordID, value.Digest[:], addition.SourceRowOrdinal, addition.SourceColumnOrdinal,
			addition.SourceKind, string(value.MetadataJSON), string(value.HeaderJSON), addition.RawValue, addition.CellKind).Scan(&exactReplay)
		if errors.Is(err, pgx.ErrNoRows) {
			absent = append(absent, index)
			continue
		}
		if err != nil {
			return fmt.Errorf("verify timeline provenance replay: %w", err)
		}
		if !exactReplay {
			return ErrProvenanceIdentityCollision
		}
	}
	if existingCount+len(absent) > provenancevalue.MaxEntriesPerRecord {
		return fmt.Errorf("%w: exceeds %d entries", ErrInvalidProvenance, provenancevalue.MaxEntriesPerRecord)
	}
	for _, index := range absent {
		addition := additions[index]
		// Re-encode one admitted item at a time instead of retaining up to4096
		// pairs of bounded JSON buffers through the preflight.
		value, err := admitProvenanceAddition(addition)
		if err != nil {
			return err
		}
		tag, err := tx.Exec(ctx, `
INSERT INTO timeline_source_provenance (
    record_id, source_identity_hash, source_row_ordinal, source_column_ordinal,
    source_kind, source_metadata, source_header_json, raw_value, cell_kind, created_at
) VALUES ($1, $2, $3, $4, $5, $6::jsonb, $7::jsonb, $8, NULLIF($9, ''), transaction_timestamp())
`, recordID, value.Digest[:], addition.SourceRowOrdinal, addition.SourceColumnOrdinal,
			addition.SourceKind, string(value.MetadataJSON), string(value.HeaderJSON), addition.RawValue, addition.CellKind)
		if err != nil {
			return fmt.Errorf("insert timeline source provenance: %w", err)
		}
		if tag.RowsAffected() != 1 {
			return ErrProvenanceIdentityCollision
		}
	}
	return nil
}

func admitProvenanceAddition(addition ClipboardRawImportColumn) (provenancevalue.Values, error) {
	var cellKind any
	if addition.CellKind != "" {
		cellKind = addition.CellKind
	}
	value, ok := provenancevalue.ParseValues(addition.SourceKind, provenanceSourceMetadata(addition), addition.SourceHeaderText, addition.RawValue, cellKind)
	if !ok || !value.Coherent() {
		return provenancevalue.Values{}, fmt.Errorf("%w: values", ErrInvalidProvenance)
	}
	if addition.SourceRowOrdinal < 1 || addition.SourceRowOrdinal > 2147483647 || addition.SourceColumnOrdinal < 1 || addition.SourceColumnOrdinal > 2147483647 {
		return provenancevalue.Values{}, fmt.Errorf("%w: identity", ErrInvalidProvenance)
	}
	return value, nil
}

func provenanceSourceMetadata(column ClipboardRawImportColumn) map[string]any {
	metadata := map[string]any{"source_kind": column.SourceKind}
	addString := func(key string, value string) {
		if value != "" {
			metadata[key] = value
		}
	}
	addString("paste_client_txn_id", column.PasteClientTxnID)
	addString("import_session_id", column.ImportSessionID)
	addString("import_unit_id", column.ImportUnitID)
	addString("mapping_fingerprint", column.MappingFingerprint)
	addString("source_file_kind", column.SourceFileKind)
	addString("source_content_sha256", column.SourceContentSHA256)
	addString("parser_profile_id", column.ParserProfileID)
	addString("parser_version", column.ParserVersion)
	addString("locator_kind", column.LocatorKind)
	addString("locator", column.Locator)
	addString("source_rect_a1", column.SourceRectA1)
	return metadata
}
