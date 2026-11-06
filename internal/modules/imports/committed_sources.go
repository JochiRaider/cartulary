package imports

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// CommittedSourceID identifies the immutable application, not its live workflow.
// The origin unit supplies stable entropy so preparation and exact replay agree.
func CommittedSourceID(unitID uuid.UUID) uuid.UUID {
	return uuid.NewSHA1(unitID, []byte("cartulary.imports.committed_source.v1"))
}

// commitSourceTx joins the existing successful unit publication transaction.
// Copying between Imports-owned relations avoids buffering source bytes in Go.
func commitSourceTx(ctx context.Context, tx pgx.Tx, start applyStartResult, unit applyUnitData, actorID uuid.UUID, now time.Time) error {
	stream, err := loadSourceMetadata(ctx, tx, unit.SourceStreamRef, true)
	if err != nil {
		return err
	}
	if stream.ImportSessionID != start.ImportSessionID || stream.ImportUnitID != unit.UnitID || stream.SourceContentSHA256 != unit.SourceContentSHA256 {
		return importApplyBlockedError("source_changed")
	}
	reader := newSourceReader(ctx, tx, stream.ImportSourceCapability, false)
	defer reader.Close()
	if _, err := io.Copy(io.Discard, reader); err != nil {
		return err
	}
	mapping, err := json.Marshal(unit.approvedMapping)
	if err != nil {
		return err
	}
	result, err := tx.Exec(ctx, `INSERT INTO import_committed_source_blobs (content_sha256, byte_size)
SELECT source_content_sha256, source_byte_size
FROM import_source_streams WHERE source_stream_ref=$1 AND source_content_sha256=$2
ON CONFLICT (content_sha256) DO NOTHING`, unit.SourceStreamRef, unit.SourceContentSHA256)
	if err != nil {
		return fmt.Errorf("retain committed source bytes: %w", err)
	}
	if result.RowsAffected() == 0 {
		var size int64
		if err := tx.QueryRow(ctx, `SELECT byte_size FROM import_committed_source_blobs WHERE content_sha256=$1`, unit.SourceContentSHA256).Scan(&size); err != nil || size != stream.SourceByteSize {
			return importApplyBlockedError("source_changed")
		}
		retained := newCommittedBlobReader(ctx, tx, unit.SourceContentSHA256, size)
		defer retained.Close()
		if _, err := io.Copy(io.Discard, retained); err != nil {
			return err
		}
	} else {
		chunks, err := tx.Exec(ctx, `INSERT INTO import_committed_source_chunks (content_sha256,chunk_ordinal,source_bytes)
SELECT s.source_content_sha256,part.byte_offset/65536,substring(s.source_bytes FROM (part.byte_offset+1)::integer FOR 65536)
FROM import_source_streams s CROSS JOIN LATERAL generate_series(0::bigint,s.source_byte_size-1,65536::bigint) AS part(byte_offset)
WHERE s.source_stream_ref=$1`, unit.SourceStreamRef)
		if err != nil {
			return fmt.Errorf("retain committed source chunks: %w", err)
		}
		if chunks.RowsAffected() != (stream.SourceByteSize+sourceReadChunkBytes-1)/sourceReadChunkBytes {
			return importApplyBlockedError("source_changed")
		}
	}
	result, err = tx.Exec(ctx, `INSERT INTO import_committed_sources (
committed_source_id, incident_id, content_sha256, source_file_kind, source_media_type,
original_filename, parser_profile_id, parser_version, locator_kind, locator, source_rect_a1, header_row_ref, data_start_row_ref,
approved_mapping_json, mapping_fingerprint, target_kind, target_view_schema_id, extension_profile_id,
origin_import_session_id, origin_import_unit_id, created_by_user_id, committed_at)
SELECT $1,$2,$3,$4,s.source_media_type,s.original_filename,$5,$6,$7,$8,$9,u.approved_header_row_ref,u.approved_data_start_row_ref,$10,$11,$12,$13,$14,$15,$16,$17,$18
FROM import_sessions s JOIN import_units u USING(import_session_id)
WHERE s.import_session_id=$15 AND s.incident_id=$2 AND u.import_unit_id=$16`,
		CommittedSourceID(unit.UnitID), start.IncidentID, unit.SourceContentSHA256, unit.SourceFileKind,
		unit.ParserProfileID, unit.ParserVersion, unit.LocatorKind, unit.Locator, unit.SourceRectA1,
		mapping, unit.MappingFingerprint, unit.approvedMapping.targetKindOrDefault(),
		unit.approvedMapping.TargetViewSchemaID, unit.approvedMapping.ExtensionProfileID,
		start.ImportSessionID, unit.UnitID, actorID, now)
	if err != nil {
		return fmt.Errorf("publish committed source provenance: %w", err)
	}
	if result.RowsAffected() != 1 {
		return importApplyBlockedError("source_changed")
	}
	if unit.approvedMapping.targetKindOrDefault() == ImportTargetKindViewSchema {
		_, err = tx.Exec(ctx, `INSERT INTO import_committed_source_records
(incident_id,committed_source_id,source_row_ref,record_id)
SELECT $1,$2,source_row_ref,record_id FROM import_apply_journal
WHERE import_unit_id=$3 AND mapping_fingerprint=$4`, start.IncidentID, CommittedSourceID(unit.UnitID), unit.UnitID, unit.MappingFingerprint)
		if err != nil {
			return fmt.Errorf("publish committed record provenance: %w", err)
		}
	}
	return nil
}
