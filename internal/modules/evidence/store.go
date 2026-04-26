package evidence

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

var (
	ErrBlobNotFound       = errors.New("evidence: blob not found")
	ErrEvidenceNotFound   = errors.New("evidence: evidence not found")
	ErrBlobNotAttachable  = errors.New("evidence: blob not attachable")
	ErrIncidentMismatch   = errors.New("evidence: incident mismatch")
	ErrRowVersionConflict = errors.New("evidence: row version conflict")
)

type Store struct {
	pool          postgres.DB
	authStore     *authn.Store
	revisionStore *revisions.Store
}

type BlobSlotParams struct {
	ObjectBlobID      uuid.UUID
	IncidentID        uuid.UUID
	ActorUserID       uuid.UUID
	StorageKey        string
	ByteSize          int64
	FilenameHint      *string
	ContentTypeHint   *string
	ExpectedSHA256Hex *string
	TargetExpiresAt   time.Time
	PendingExpiresAt  time.Time
	UploadTarget      map[string]any
	AcceptedContract  map[string]any
	RequestHash       []byte
	ClientTxnID       string
}

type BlobSlotResult struct {
	Payload    map[string]any
	StatusCode int
	Replayed   bool
}

type BlobRecord struct {
	ObjectBlobID        uuid.UUID
	IncidentID          uuid.UUID
	StorageKey          string
	UploadState         string
	ByteSize            int64
	FilenameHint        *string
	ContentTypeHint     *string
	ExpectedSHA256Hex   *string
	ObservedSize        *int64
	ObservedContentType *string
	ObservedSHA256Hex   *string
	TargetExpiresAt     time.Time
	PendingExpiresAt    time.Time
}

type ObservedObject struct {
	Size        int64
	ContentType string
	SHA256Hex   string
}

type AttachBlobResult struct {
	Payload          map[string]any
	StatusCode       int
	Replayed         bool
	IncidentID       uuid.UUID
	RecordID         uuid.UUID
	ChangeSetID      uuid.UUID
	ClientTxnID      string
	RowVersion       int64
	ChangedFieldKeys []string
}

type EvidenceAccessRecord struct {
	IncidentID             uuid.UUID
	RecordID               uuid.UUID
	ObjectBlobID           uuid.UUID
	StorageKey             string
	EvidenceLifecycleState string
	UploadState            string
	FilenameSource         string
	ContentType            string
	SizeBytes              int64
	SHA256                 *string
	MediaClass             string
	PreviewKind            *string
}

type HandleRecord struct {
	Token        string
	IncidentID   uuid.UUID
	RecordID     uuid.UUID
	ObjectBlobID uuid.UUID
	StorageKey   string
	SessionID    uuid.UUID
	HandleKind   string
	MediaClass   string
	PreviewKind  *string
	Disposition  string
	Filename     string
	ContentType  string
	SizeBytes    int64
	SHA256       *string
	ExpiresAt    time.Time
	ConsumedAt   *time.Time
}

func NewStore(pool postgres.DB) *Store {
	return &Store{pool: pool, authStore: authn.NewStore(pool), revisionStore: revisions.NewStore()}
}

func (s *Store) CreateBlobSlot(ctx context.Context, params BlobSlotParams) (BlobSlotResult, error) {
	key := authn.RouteIdempotencyKey{
		RouteKey: blobCreateRouteKey, ActorUserID: params.ActorUserID,
		ScopeKey: params.IncidentID.String(), ClientTxnID: params.ClientTxnID,
	}
	if existing, err := s.authStore.GetRouteIdempotency(ctx, key); err == nil {
		if !bytes.Equal(existing.RequestHash, params.RequestHash) {
			return BlobSlotResult{}, authn.ErrClientTxnConflict
		}
		payload, err := decodeStoredPayload(existing.ResponseJSON)
		if err != nil {
			return BlobSlotResult{}, err
		}
		return BlobSlotResult{Payload: payload, StatusCode: http.StatusOK, Replayed: true}, nil
	} else if !errors.Is(err, authn.ErrNotFound) {
		return BlobSlotResult{}, err
	}

	payload := map[string]any{
		"incident_id":        params.IncidentID.String(),
		"object_blob_id":     params.ObjectBlobID.String(),
		"upload_state":       "pending",
		"target_expires_at":  formatHTTPTime(params.TargetExpiresAt),
		"pending_expires_at": formatHTTPTime(params.PendingExpiresAt),
		"upload_target":      params.UploadTarget,
		"accepted_contract":  params.AcceptedContract,
	}

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return BlobSlotResult{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if err := ensureIncidentVisibleTx(ctx, tx, params.IncidentID); err != nil {
		return BlobSlotResult{}, err
	}
	_, err = tx.Exec(ctx, `
INSERT INTO object_blobs (
    object_blob_id, incident_id, created_by_user_id, storage_key, upload_state,
    byte_size, filename_hint, content_type_hint, expected_sha256_hex,
    target_expires_at, pending_expires_at, created_at, updated_at
) VALUES ($1, $2, $3, $4, 'pending', $5, $6, $7, $8, $9, $10, $11, $11)
`, params.ObjectBlobID, params.IncidentID, params.ActorUserID, params.StorageKey, params.ByteSize,
		params.FilenameHint, params.ContentTypeHint, params.ExpectedSHA256Hex,
		params.TargetExpiresAt.UTC(), params.PendingExpiresAt.UTC(), params.TargetExpiresAt.Add(-60*time.Minute).UTC())
	if err != nil {
		return BlobSlotResult{}, err
	}
	if err := authn.InsertRouteIdempotencyPayload(ctx, tx, key, nil, params.RequestHash, http.StatusCreated, payload); err != nil {
		if authn.IsUniqueViolation(err) {
			return BlobSlotResult{}, authn.ErrClientTxnConflict
		}
		return BlobSlotResult{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return BlobSlotResult{}, err
	}
	return BlobSlotResult{Payload: payload, StatusCode: http.StatusCreated}, nil
}

func (s *Store) GetBlob(ctx context.Context, objectBlobID uuid.UUID) (BlobRecord, error) {
	row := s.pool.QueryRow(ctx, `
SELECT object_blob_id, incident_id, storage_key, upload_state, byte_size,
       filename_hint, content_type_hint, expected_sha256_hex,
       observed_size, observed_content_type, observed_sha256_hex,
       target_expires_at, pending_expires_at
  FROM object_blobs
 WHERE object_blob_id = $1
`, objectBlobID)
	var record BlobRecord
	if err := row.Scan(&record.ObjectBlobID, &record.IncidentID, &record.StorageKey, &record.UploadState, &record.ByteSize,
		&record.FilenameHint, &record.ContentTypeHint, &record.ExpectedSHA256Hex,
		&record.ObservedSize, &record.ObservedContentType, &record.ObservedSHA256Hex,
		&record.TargetExpiresAt, &record.PendingExpiresAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return BlobRecord{}, ErrBlobNotFound
		}
		return BlobRecord{}, err
	}
	return record, nil
}

func (s *Store) AttachBlob(ctx context.Context, actor authn.UserRecord, recordID uuid.UUID, request AttachBlobRequest, requestHash []byte, observed *ObservedObject, requestID string, now time.Time) (AttachBlobResult, error) {
	key := authn.RouteIdempotencyKey{
		RouteKey: blobAttachRouteKey, ActorUserID: actor.ID,
		ScopeKey: recordID.String(), ClientTxnID: request.ClientTxnID,
	}
	if existing, err := s.authStore.GetRouteIdempotency(ctx, key); err == nil {
		if !bytes.Equal(existing.RequestHash, requestHash) {
			return AttachBlobResult{}, authn.ErrClientTxnConflict
		}
		payload, err := decodeStoredPayload(existing.ResponseJSON)
		if err != nil {
			return AttachBlobResult{}, err
		}
		return AttachBlobResult{Payload: payload, StatusCode: http.StatusOK, Replayed: true, RecordID: recordID, ClientTxnID: request.ClientTxnID}, nil
	} else if !errors.Is(err, authn.ErrNotFound) {
		return AttachBlobResult{}, err
	}

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return AttachBlobResult{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	meta, err := loadEvidenceMetaForUpdateTx(ctx, tx, recordID)
	if err != nil {
		return AttachBlobResult{}, err
	}
	if meta.RowVersion != request.BaseRowVersion {
		return AttachBlobResult{}, &rowVersionConflictError{RecordID: recordID, BaseRowVersion: request.BaseRowVersion, CurrentRowVersion: meta.RowVersion}
	}
	beforeRow, err := loadEvidenceRowTx(ctx, tx, recordID)
	if err != nil {
		return AttachBlobResult{}, err
	}
	blob, err := loadBlobForUpdateTx(ctx, tx, request.ObjectBlobID)
	if err != nil {
		return AttachBlobResult{}, err
	}
	if blob.IncidentID != meta.IncidentID {
		return AttachBlobResult{}, ErrIncidentMismatch
	}
	if blob.UploadState == "quarantined" || blob.UploadState == "failed" {
		return AttachBlobResult{}, ErrBlobNotAttachable
	}
	if blob.UploadState == "pending" {
		if now.After(blob.PendingExpiresAt) {
			if err := failBlobTx(ctx, tx, request.ObjectBlobID, "pending_timeout", now); err != nil {
				return AttachBlobResult{}, err
			}
			return AttachBlobResult{}, ErrBlobNotAttachable
		}
		if observed == nil {
			return AttachBlobResult{}, ErrBlobNotAttachable
		}
		if observed.Size != blob.ByteSize {
			if err := failBlobTx(ctx, tx, request.ObjectBlobID, "declared_size_mismatch", now); err != nil {
				return AttachBlobResult{}, err
			}
			return AttachBlobResult{}, ErrBlobNotAttachable
		}
		if blob.ExpectedSHA256Hex != nil && observed.SHA256Hex != *blob.ExpectedSHA256Hex {
			if err := failBlobTx(ctx, tx, request.ObjectBlobID, "expected_sha256_mismatch", now); err != nil {
				return AttachBlobResult{}, err
			}
			return AttachBlobResult{}, ErrBlobNotAttachable
		}
		if err := markBlobAvailableTx(ctx, tx, request.ObjectBlobID, observed, now); err != nil {
			return AttachBlobResult{}, err
		}
		blob.UploadState = "available"
		blob.ObservedSize = &observed.Size
		blob.ObservedContentType = &observed.ContentType
		blob.ObservedSHA256Hex = &observed.SHA256Hex
	}
	if blob.UploadState != "available" {
		return AttachBlobResult{}, ErrBlobNotAttachable
	}
	sha := blob.ObservedSHA256Hex
	if sha == nil && observed != nil {
		sha = &observed.SHA256Hex
	}
	_, err = tx.Exec(ctx, `
UPDATE evidence
   SET object_blob_id = $2,
       lifecycle_state = CASE WHEN lifecycle_state IN ('requested', 'pending_receipt', 'received') THEN 'available' ELSE lifecycle_state END,
       upload_state = 'available',
       storage_ref = $3,
       blob_hash = $4,
       received_at = COALESCE(received_at, $5),
       updated_at = $5
 WHERE record_id = $1
`, recordID, request.ObjectBlobID, "object://"+request.ObjectBlobID.String(), sha, now.UTC())
	if err != nil {
		return AttachBlobResult{}, err
	}
	rowVersion, err := advanceRecordVersionTx(ctx, tx, recordID, actor.ID, now)
	if err != nil {
		return AttachBlobResult{}, err
	}
	afterRow, err := loadEvidenceRowTx(ctx, tx, recordID)
	if err != nil {
		return AttachBlobResult{}, err
	}
	changeSetID, err := s.revisionStore.InsertChangeSetTx(ctx, tx, revisions.ChangeSetParams{
		IncidentID: meta.IncidentID, ActorUserID: actor.ID, Source: blobAttachRouteKey,
		ClientTxnID: &request.ClientTxnID, RequestID: &requestID, CreatedAt: now.UTC(),
	})
	if err != nil {
		return AttachBlobResult{}, err
	}
	beforeVersionID := fmt.Sprintf("%s:%d", recordID, request.BaseRowVersion)
	afterVersionID := fmt.Sprintf("%s:%d", recordID, rowVersion)
	if err := s.revisionStore.InsertMutationTx(ctx, tx, revisions.MutationParams{
		ChangeSetID: changeSetID, SequenceNo: 1, TargetKind: "record", TargetID: recordID.String(),
		OperationKind: "patch", BeforeVersionID: &beforeVersionID, AfterVersionID: &afterVersionID,
		BeforeValue: beforeRow, AfterValue: afterRow,
	}); err != nil {
		return AttachBlobResult{}, err
	}
	if err := s.revisionStore.InsertRecordRevisionTx(ctx, tx, revisions.RecordRevisionParams{
		ChangeSetID: changeSetID, RecordID: recordID, RowVersion: rowVersion, BeforeValue: beforeRow, AfterValue: afterRow,
	}); err != nil {
		return AttachBlobResult{}, err
	}
	payload := map[string]any{
		"view_schema_id": evidenceViewSchemaID,
		"change_set_id":  changeSetID.String(),
		"object_blob_id": request.ObjectBlobID.String(),
		"row":            afterRow,
	}
	if err := authn.InsertRouteIdempotencyPayload(ctx, tx, key, nil, requestHash, http.StatusOK, payload); err != nil {
		if authn.IsUniqueViolation(err) {
			return AttachBlobResult{}, authn.ErrClientTxnConflict
		}
		return AttachBlobResult{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return AttachBlobResult{}, err
	}
	return AttachBlobResult{
		Payload: payload, StatusCode: http.StatusOK, IncidentID: meta.IncidentID, RecordID: recordID,
		ChangeSetID: changeSetID, ClientTxnID: request.ClientTxnID, RowVersion: rowVersion,
		ChangedFieldKeys: sortedChangedKeys(beforeRow, afterRow),
	}, nil
}

func (s *Store) LoadEvidenceAccess(ctx context.Context, recordID uuid.UUID) (EvidenceAccessRecord, error) {
	row := s.pool.QueryRow(ctx, `
SELECT e.incident_id, e.record_id, e.object_blob_id, b.storage_key,
       e.lifecycle_state, b.upload_state,
       COALESCE(b.filename_hint, e.title, ''),
       COALESCE(b.observed_content_type, b.content_type_hint, 'application/octet-stream'),
       COALESCE(b.observed_size, b.byte_size),
       b.observed_sha256_hex
  FROM evidence e
  JOIN records r ON r.record_id = e.record_id AND r.deleted_at IS NULL
  JOIN object_blobs b ON b.object_blob_id = e.object_blob_id
 WHERE e.record_id = $1
`, recordID)
	var access EvidenceAccessRecord
	if err := row.Scan(&access.IncidentID, &access.RecordID, &access.ObjectBlobID, &access.StorageKey,
		&access.EvidenceLifecycleState, &access.UploadState, &access.FilenameSource, &access.ContentType,
		&access.SizeBytes, &access.SHA256); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return EvidenceAccessRecord{}, ErrEvidenceNotFound
		}
		return EvidenceAccessRecord{}, err
	}
	access.MediaClass, access.PreviewKind = classifyMedia(access.ContentType)
	return access, nil
}

func (s *Store) InsertHandle(ctx context.Context, handle HandleRecord, issuedByUserID uuid.UUID) error {
	_, err := s.pool.Exec(ctx, `
INSERT INTO evidence_access_handles (
    handle_token, incident_id, record_id, object_blob_id, issued_by_user_id, issuing_session_id,
    handle_kind, media_class, preview_kind, disposition, filename, content_type,
    size_bytes, sha256, evidence_lifecycle_state, upload_state, expires_at, created_at
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, 'available', 'available', $15, $16)
`, handle.Token, handle.IncidentID, handle.RecordID, handle.ObjectBlobID, issuedByUserID, handle.SessionID,
		handle.HandleKind, handle.MediaClass, handle.PreviewKind, handle.Disposition, handle.Filename,
		handle.ContentType, handle.SizeBytes, handle.SHA256, handle.ExpiresAt.UTC(), time.Now().UTC())
	return err
}

func (s *Store) LoadHandle(ctx context.Context, token string) (HandleRecord, error) {
	row := s.pool.QueryRow(ctx, `
SELECT h.handle_token, h.incident_id, h.record_id, h.object_blob_id, b.storage_key,
       h.issuing_session_id, h.handle_kind, h.media_class, h.preview_kind, h.disposition,
       h.filename, h.content_type, h.size_bytes, h.sha256, h.expires_at, h.consumed_at
  FROM evidence_access_handles h
  JOIN object_blobs b ON b.object_blob_id = h.object_blob_id
 WHERE h.handle_token = $1
`, token)
	var handle HandleRecord
	if err := row.Scan(&handle.Token, &handle.IncidentID, &handle.RecordID, &handle.ObjectBlobID, &handle.StorageKey,
		&handle.SessionID, &handle.HandleKind, &handle.MediaClass, &handle.PreviewKind, &handle.Disposition,
		&handle.Filename, &handle.ContentType, &handle.SizeBytes, &handle.SHA256, &handle.ExpiresAt, &handle.ConsumedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return HandleRecord{}, ErrBlobNotFound
		}
		return HandleRecord{}, err
	}
	return handle, nil
}

func (s *Store) ConsumeDownloadHandle(ctx context.Context, token string, now time.Time) error {
	tag, err := s.pool.Exec(ctx, `
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

func (s *Store) CheckHandleAccess(ctx context.Context, handle HandleRecord) error {
	var ok bool
	err := s.pool.QueryRow(ctx, `
SELECT EXISTS (
    SELECT 1
      FROM evidence e
      JOIN records r ON r.record_id = e.record_id AND r.deleted_at IS NULL
      JOIN object_blobs b ON b.object_blob_id = e.object_blob_id
     WHERE e.incident_id = $1
       AND e.record_id = $2
       AND e.object_blob_id = $3
       AND e.lifecycle_state IN ('available', 'released')
       AND b.upload_state = 'available'
)
`, handle.IncidentID, handle.RecordID, handle.ObjectBlobID).Scan(&ok)
	if err != nil {
		return err
	}
	if !ok {
		return ErrBlobNotAttachable
	}
	return nil
}

type evidenceMeta struct {
	IncidentID uuid.UUID
	RowVersion int64
}

type rowVersionConflictError struct {
	RecordID          uuid.UUID
	BaseRowVersion    int64
	CurrentRowVersion int64
}

func (e *rowVersionConflictError) Error() string { return ErrRowVersionConflict.Error() }
func (e *rowVersionConflictError) Unwrap() error { return ErrRowVersionConflict }

func loadEvidenceMetaForUpdateTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) (evidenceMeta, error) {
	var meta evidenceMeta
	err := tx.QueryRow(ctx, `
SELECT r.incident_id, r.row_version
  FROM records r
  JOIN evidence e ON e.record_id = r.record_id
 WHERE r.record_id = $1
   AND r.record_type = 'evidence'
   AND r.deleted_at IS NULL
 FOR UPDATE OF r, e
`, recordID).Scan(&meta.IncidentID, &meta.RowVersion)
	if errors.Is(err, pgx.ErrNoRows) {
		return evidenceMeta{}, ErrEvidenceNotFound
	}
	return meta, err
}

func loadBlobForUpdateTx(ctx context.Context, tx pgx.Tx, objectBlobID uuid.UUID) (BlobRecord, error) {
	row := tx.QueryRow(ctx, `
SELECT object_blob_id, incident_id, storage_key, upload_state, byte_size,
       filename_hint, content_type_hint, expected_sha256_hex,
       observed_size, observed_content_type, observed_sha256_hex,
       target_expires_at, pending_expires_at
  FROM object_blobs
 WHERE object_blob_id = $1
 FOR UPDATE
`, objectBlobID)
	var record BlobRecord
	if err := row.Scan(&record.ObjectBlobID, &record.IncidentID, &record.StorageKey, &record.UploadState, &record.ByteSize,
		&record.FilenameHint, &record.ContentTypeHint, &record.ExpectedSHA256Hex,
		&record.ObservedSize, &record.ObservedContentType, &record.ObservedSHA256Hex,
		&record.TargetExpiresAt, &record.PendingExpiresAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return BlobRecord{}, ErrBlobNotFound
		}
		return BlobRecord{}, err
	}
	return record, nil
}

func failBlobTx(ctx context.Context, tx pgx.Tx, objectBlobID uuid.UUID, reason string, now time.Time) error {
	_, err := tx.Exec(ctx, `
UPDATE object_blobs
   SET upload_state = 'failed',
       terminal_reason = $2,
       failed_at = $3,
       cleanup_due_at = $3 + interval '1 hour',
       updated_at = $3
 WHERE object_blob_id = $1
`, objectBlobID, reason, now.UTC())
	return err
}

func markBlobAvailableTx(ctx context.Context, tx pgx.Tx, objectBlobID uuid.UUID, observed *ObservedObject, now time.Time) error {
	_, err := tx.Exec(ctx, `
UPDATE object_blobs
   SET upload_state = 'available',
       observed_size = $2,
       observed_content_type = $3,
       observed_sha256_hex = $4,
       finalized_at = $5,
       updated_at = $5
 WHERE object_blob_id = $1
`, objectBlobID, observed.Size, observed.ContentType, observed.SHA256Hex, now.UTC())
	return err
}

func loadEvidenceRowTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) (map[string]any, error) {
	row := tx.QueryRow(ctx, `
SELECT e.record_id, r.row_version,
       e.title, e.lifecycle_state, e.requested_at, e.received_at, e.storage_ref, e.blob_hash,
       e.collector_party_text, e.collector_party_id, e.source_party_text, e.source_party_id,
       COALESCE(b.upload_state, e.upload_state), 0::integer, e.updated_at
  FROM evidence e
  JOIN records r ON r.record_id = e.record_id
  LEFT JOIN object_blobs b ON b.object_blob_id = e.object_blob_id
 WHERE e.record_id = $1
   AND r.deleted_at IS NULL
`, recordID)
	values := make([]any, 15)
	targets := make([]any, len(values))
	for index := range values {
		targets[index] = &values[index]
	}
	if err := row.Scan(targets...); err != nil {
		return nil, err
	}
	return evidenceRowFromValues(values)
}

func evidenceRowFromValues(values []any) (map[string]any, error) {
	recordID, err := uuidFromDBValue(values[0])
	if err != nil {
		return nil, err
	}
	keys := []string{
		"evidence.title", "evidence.lifecycle_state", "evidence.requested_at", "evidence.received_at",
		"evidence.storage_ref", "evidence.blob_hash", "evidence.collector_party_text", "evidence.collector_party_id",
		"evidence.source_party_text", "evidence.source_party_id", "evidence.upload_state",
		"evidence.linked_record_count", "evidence.edited_at",
	}
	cells := make(map[string]any, len(keys))
	for index, key := range keys {
		cells[key] = map[string]any{"value": publicCellValue(values[index+2])}
	}
	return map[string]any{
		"record_id":    recordID.String(),
		"row_version":  values[1],
		"cells":        cells,
		"group_values": map[string]any{},
	}, nil
}

func uuidFromDBValue(value any) (uuid.UUID, error) {
	switch typed := value.(type) {
	case uuid.UUID:
		return typed, nil
	case [16]byte:
		return uuid.UUID(typed), nil
	case []byte:
		if len(typed) == 16 {
			return uuid.FromBytes(typed)
		}
		return uuid.Parse(string(typed))
	case string:
		return uuid.Parse(typed)
	default:
		return uuid.UUID{}, fmt.Errorf("evidence row record_id was %T", value)
	}
}

func publicCellValue(value any) any {
	if value == nil {
		return nil
	}
	switch typed := value.(type) {
	case time.Time:
		return typed.UTC().Format(time.RFC3339Nano)
	case uuid.UUID:
		return typed.String()
	case []byte:
		if len(typed) == 16 {
			if id, err := uuid.FromBytes(typed); err == nil {
				return id.String()
			}
		}
		return string(typed)
	default:
		return typed
	}
}

func advanceRecordVersionTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID, actorUserID uuid.UUID, now time.Time) (int64, error) {
	var rowVersion int64
	err := tx.QueryRow(ctx, `
UPDATE records
   SET row_version = row_version + 1,
       updated_at = $2,
       updated_by_user_id = $3
 WHERE record_id = $1
RETURNING row_version
`, recordID, now.UTC(), actorUserID).Scan(&rowVersion)
	return rowVersion, err
}

func ensureIncidentVisibleTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID) error {
	var exists bool
	if err := tx.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM incidents WHERE id = $1)`, incidentID).Scan(&exists); err != nil {
		return err
	}
	if !exists {
		return pgx.ErrNoRows
	}
	return nil
}

func decodeStoredPayload(data []byte) (map[string]any, error) {
	var payload map[string]any
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, err
	}
	return payload, nil
}

func firstNonEmptyPtr(left *string, right *string, fallback string) string {
	if left != nil && *left != "" {
		return *left
	}
	if right != nil && *right != "" {
		return *right
	}
	return fallback
}
