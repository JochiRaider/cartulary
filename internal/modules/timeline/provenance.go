package timeline

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

const (
	maxProvenanceEntriesPerRecord = 4096
	maxProvenanceValueBytes       = 64 * 1024
)

var ErrProvenanceIdentityCollision = errors.New("timeline provenance identity collision")

func insertSourceProvenanceTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID, additions []ClipboardRawImportColumn) error {
	if len(additions) == 0 {
		return nil
	}
	if recordID == uuid.Nil || len(additions) > maxProvenanceEntriesPerRecord {
		return errors.New("timeline provenance entry count is invalid")
	}
	var existingCount int
	if err := tx.QueryRow(ctx, `
SELECT count(*)
  FROM timeline_source_provenance
 WHERE record_id = $1
`, recordID).Scan(&existingCount); err != nil {
		return fmt.Errorf("count timeline source provenance: %w", err)
	}
	if existingCount+len(additions) > maxProvenanceEntriesPerRecord {
		return fmt.Errorf("timeline provenance exceeds %d entries", maxProvenanceEntriesPerRecord)
	}

	for _, addition := range additions {
		metadata := provenanceSourceMetadata(addition)
		metadataJSON, err := json.Marshal(metadata)
		if err != nil {
			return fmt.Errorf("encode timeline provenance metadata: %w", err)
		}
		headerJSON, err := json.Marshal(addition.SourceHeaderText)
		if err != nil {
			return fmt.Errorf("encode timeline provenance header: %w", err)
		}
		if len(metadataJSON) > maxProvenanceValueBytes ||
			len(headerJSON) > maxProvenanceValueBytes ||
			len([]byte(addition.RawValue)) > maxProvenanceValueBytes {
			return fmt.Errorf("timeline provenance value exceeds %d bytes", maxProvenanceValueBytes)
		}
		if addition.SourceKind == "" || addition.SourceRowOrdinal < 0 || addition.SourceColumnOrdinal < 0 {
			return errors.New("timeline provenance identity is invalid")
		}
		identityHash := sha256.Sum256(metadataJSON)
		tag, err := tx.Exec(ctx, `
INSERT INTO timeline_source_provenance (
    record_id,
    source_identity_hash,
    source_row_ordinal,
    source_column_ordinal,
    source_kind,
    source_metadata,
    source_header_json,
    raw_value,
    cell_kind,
    created_at
) VALUES ($1, $2, $3, $4, $5, $6::jsonb, $7::jsonb, $8, NULLIF($9, ''), transaction_timestamp())
ON CONFLICT (
    record_id,
    source_identity_hash,
    source_row_ordinal,
    source_column_ordinal
) DO NOTHING
`, recordID, identityHash[:], addition.SourceRowOrdinal, addition.SourceColumnOrdinal,
			addition.SourceKind, string(metadataJSON), string(headerJSON), addition.RawValue, addition.CellKind)
		if err != nil {
			return fmt.Errorf("insert timeline source provenance: %w", err)
		}
		if tag.RowsAffected() == 1 {
			continue
		}
		var exactReplay bool
		if err := tx.QueryRow(ctx, `
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
`, recordID, identityHash[:], addition.SourceRowOrdinal, addition.SourceColumnOrdinal,
			addition.SourceKind, string(metadataJSON), string(headerJSON), addition.RawValue, addition.CellKind).Scan(&exactReplay); err != nil {
			return fmt.Errorf("verify timeline provenance replay: %w", err)
		}
		if !exactReplay {
			return ErrProvenanceIdentityCollision
		}
	}
	return nil
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
