package evidence

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

// accessHandleRepository owns Evidence access snapshots and opaque handle
// persistence. It has no route, authorization, or object-store dependency.
type accessHandleRepository struct {
	db postgres.DB
}

func (repository accessHandleRepository) loadEvidence(
	ctx context.Context,
	recordID uuid.UUID,
) (evidenceAccessRecord, error) {
	row := repository.db.QueryRow(ctx, `
SELECT e.incident_id, e.record_id, r.row_version, e.object_blob_id::text, b.object_blob_id IS NOT NULL, b.storage_key,
       e.lifecycle_state, COALESCE(b.upload_state, ''),
       COALESCE(b.filename_hint, e.title, ''),
       COALESCE(b.observed_content_type, b.content_type_hint, 'application/octet-stream'),
       COALESCE(b.observed_size, b.byte_size, 0),
       b.observed_sha256_hex
  FROM evidence e
  JOIN records r ON r.record_id = e.record_id AND r.deleted_at IS NULL
  LEFT JOIN object_blobs b ON b.object_blob_id = e.object_blob_id
 WHERE e.record_id = $1
`, recordID)
	var access evidenceAccessRecord
	var objectBlobID sql.NullString
	var storageKey sql.NullString
	var sha256 sql.NullString
	if err := row.Scan(&access.IncidentID, &access.RecordID, &access.RecordRowVersion, &objectBlobID, &access.BlobMetadataVisible, &storageKey,
		&access.EvidenceLifecycleState, &access.UploadState, &access.FilenameSource, &access.ContentType,
		&access.SizeBytes, &sha256); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return evidenceAccessRecord{}, ErrEvidenceNotFound
		}
		return evidenceAccessRecord{}, err
	}
	if objectBlobID.Valid {
		parsed, err := uuid.Parse(objectBlobID.String)
		if err != nil {
			return evidenceAccessRecord{}, err
		}
		access.ObjectBlobID = &parsed
	}
	if storageKey.Valid {
		value := storageKey.String
		access.StorageKey = &value
	}
	if sha256.Valid {
		value := sha256.String
		access.SHA256 = &value
	}
	access.MediaClass, access.PreviewKind = classifyMedia(access.ContentType)
	return access, nil
}

func (repository accessHandleRepository) insert(
	ctx context.Context,
	handle handleRecord,
	issuedByUserID uuid.UUID,
) error {
	_, err := repository.db.Exec(ctx, `
INSERT INTO evidence_access_handles (
    handle_token, incident_id, record_id, record_row_version, object_blob_id, issued_by_user_id, issuing_session_id,
    handle_kind, media_class, preview_kind, disposition, filename, content_type,
    size_bytes, sha256, evidence_lifecycle_state, upload_state, expires_at, created_at
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19)
`, handle.Token, handle.IncidentID, handle.RecordID, handle.RecordRowVersion, handle.ObjectBlobID, issuedByUserID, handle.SessionID,
		handle.HandleKind, handle.MediaClass, handle.PreviewKind, handle.Disposition, handle.Filename,
		handle.ContentType, handle.SizeBytes, handle.SHA256, handle.EvidenceLifecycleState, handle.UploadState, handle.ExpiresAt.UTC(), time.Now().UTC())
	return err
}

func (repository accessHandleRepository) load(
	ctx context.Context,
	token string,
) (handleRecord, error) {
	row := repository.db.QueryRow(ctx, `
SELECT h.handle_token, h.incident_id, h.record_id, h.object_blob_id, b.storage_key,
       h.record_row_version, h.issuing_session_id, h.handle_kind, h.media_class, h.preview_kind, h.disposition,
       h.filename, h.content_type, h.size_bytes, h.sha256, h.evidence_lifecycle_state, h.upload_state, h.expires_at, h.consumed_at
  FROM evidence_access_handles h
  JOIN object_blobs b ON b.object_blob_id = h.object_blob_id
 WHERE h.handle_token = $1
`, token)
	var handle handleRecord
	if err := row.Scan(&handle.Token, &handle.IncidentID, &handle.RecordID, &handle.ObjectBlobID, &handle.StorageKey,
		&handle.RecordRowVersion, &handle.SessionID, &handle.HandleKind, &handle.MediaClass, &handle.PreviewKind, &handle.Disposition,
		&handle.Filename, &handle.ContentType, &handle.SizeBytes, &handle.SHA256, &handle.EvidenceLifecycleState, &handle.UploadState, &handle.ExpiresAt, &handle.ConsumedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return handleRecord{}, ErrBlobNotFound
		}
		return handleRecord{}, err
	}
	return handle, nil
}

func (repository accessHandleRepository) consumeDownload(
	ctx context.Context,
	token string,
	now time.Time,
) error {
	tag, err := repository.db.Exec(ctx, `
UPDATE evidence_access_handles
   SET consumed_at = $2
 WHERE handle_token = $1
   AND consumed_at IS NULL
`, token, now.UTC())
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return errors.New("handle already consumed")
	}
	return nil
}

func (repository accessHandleRepository) checkCurrent(
	ctx context.Context,
	handle handleRecord,
) (string, error) {
	access, err := repository.loadEvidence(ctx, handle.RecordID)
	if errors.Is(err, ErrEvidenceNotFound) {
		return "evidence_inconsistent", nil
	}
	if err != nil {
		return "", err
	}
	if access.IncidentID != handle.IncidentID {
		return "evidence_inconsistent", nil
	}
	if access.EvidenceLifecycleState == "quarantined" || access.UploadState == "quarantined" {
		return "evidence_quarantined", nil
	}
	if access.RecordRowVersion != handle.RecordRowVersion {
		return "evidence_inconsistent", nil
	}
	if reasonCode := classifyEvidenceAccess(access, &handle.ObjectBlobID); reasonCode != "" {
		return reasonCode, nil
	}
	if access.ContentType != handle.ContentType ||
		access.SizeBytes != handle.SizeBytes ||
		access.MediaClass != handle.MediaClass ||
		access.EvidenceLifecycleState != handle.EvidenceLifecycleState ||
		access.UploadState != handle.UploadState ||
		sanitizeFilename(access.FilenameSource, access.RecordID, access.ContentType) != handle.Filename {
		return "evidence_inconsistent", nil
	}
	if !nullableStringEqual(access.SHA256, handle.SHA256) {
		return "evidence_inconsistent", nil
	}
	if handle.HandleKind == "preview" {
		if access.PreviewKind == nil || handle.PreviewKind == nil || *access.PreviewKind != *handle.PreviewKind || handle.Disposition != "inline" {
			return "evidence_inconsistent", nil
		}
		return "", nil
	}
	if handle.HandleKind == "download" {
		if access.PreviewKind != nil && handle.PreviewKind != nil {
			return "evidence_inconsistent", nil
		}
		if handle.PreviewKind != nil || handle.Disposition != "attachment" {
			return "evidence_inconsistent", nil
		}
		return "", nil
	}
	return "evidence_inconsistent", nil
}
