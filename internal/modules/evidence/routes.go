package evidence

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/auth"
	"github.com/JochiRaider/cartulary/internal/modules/evidence/blobref"
	"github.com/JochiRaider/cartulary/internal/modules/incidents"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/objectstore"
	platformws "github.com/JochiRaider/cartulary/internal/platform/ws"
)

type Service struct {
	store          *Store
	incidentStore  *incidents.Store
	authStore      *authn.Store
	objectStore    objectstore.Store
	hub            *platformws.Hub
	keys           authn.MasterKeys
	now            func() time.Time
	maxBlobBytes   int64
	previewMax     int64
	textPreviewMax int64
}

func RegisterRoutes() httpapi.RouteRegistrar {
	return func(mux *http.ServeMux, deps httpapi.DependencySet) error {
		service, err := newService(deps)
		if err != nil {
			return err
		}
		mux.HandleFunc("POST /api/v1/object-blobs", service.handleCreateBlob)
		mux.HandleFunc("PUT /api/v1/object-uploads/{upload_token}", service.handleUploadTarget)
		mux.HandleFunc("POST /api/v1/evidence-records/{record_id}/attach-blob", service.handleAttachBlob)
		mux.HandleFunc("POST /api/v1/evidence-records/{record_id}/preview-handle", service.handlePreviewHandle)
		mux.HandleFunc("POST /api/v1/evidence-records/{record_id}/download-handle", service.handleDownloadHandle)
		mux.HandleFunc("GET /api/v1/evidence-handles/{handle_token}", service.handleRedeemHandle)
		return nil
	}
}

func newService(deps httpapi.DependencySet) (*Service, error) {
	keys, err := authn.LoadMasterKeys(deps.Env)
	if err != nil {
		return nil, err
	}
	now := deps.Now
	if now == nil {
		now = func() time.Time { return time.Now().UTC() }
	}
	return &Service{
		store:          NewStore(deps.Postgres),
		incidentStore:  incidents.NewStore(deps.Postgres),
		authStore:      authn.NewStore(deps.Postgres),
		objectStore:    deps.ObjectStore,
		hub:            deps.WSHub,
		keys:           keys,
		now:            now,
		maxBlobBytes:   deps.Config.Limits.ObjectBlobs.MaxDeclaredByteSize,
		previewMax:     deps.Config.Limits.Previews.MaxPreviewablePayloadBytes,
		textPreviewMax: deps.Config.Limits.Previews.MaxTextInlineBytes,
	}, nil
}

func (s *Service) handleCreateBlob(w http.ResponseWriter, r *http.Request) {
	principal, apiErr := s.authenticateSessionRequest(r, true)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	request, apiErr := DecodeBlobCreateRequest(r.Body, s.maxBlobBytes)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	if _, apiErr := s.requireIncidentRole(r.Context(), request.IncidentID, principal.User.ID, "editor", "reviewer", "admin"); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	now := s.now().UTC()
	objectBlobID := uuid.New()
	storageKey, err := blobref.ObjectBlobStorageKey(request.IncidentID, objectBlobID)
	if err != nil {
		writeAPIError(w, r, internalAPIError(fmt.Errorf("build object blob storage key: %w", err)))
		return
	}
	targetExpiresAt := now.Add(60 * time.Minute)
	pendingExpiresAt := now.Add(24 * time.Hour)
	target, err := s.createObjectUploadTarget(objectUploadTokenClaims{
		Version:           1,
		ObjectBlobID:      objectBlobID,
		IncidentID:        request.IncidentID,
		StorageKey:        storageKey,
		ByteSize:          request.ByteSize,
		ExpiresAtUnixNano: targetExpiresAt.UnixNano(),
	})
	if err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	result, err := s.store.CreateBlobSlot(r.Context(), BlobSlotParams{
		ObjectBlobID: objectBlobID, IncidentID: request.IncidentID, ActorUserID: principal.User.ID,
		StorageKey: storageKey, ByteSize: request.ByteSize, FilenameHint: request.FilenameHint,
		ContentTypeHint: request.ContentTypeHint, ExpectedSHA256Hex: request.SHA256Hex,
		TargetExpiresAt: targetExpiresAt, PendingExpiresAt: pendingExpiresAt,
		UploadTarget: map[string]any{
			"href":       target.Href,
			"method":     target.Method,
			"expires_at": formatHTTPTime(targetExpiresAt),
			"headers":    target.Headers,
		},
		AcceptedContract: request.AcceptedContract,
		RequestHash:      BlobCreateRequestHash(request),
		ClientTxnID:      request.ClientTxnID,
	})
	switch {
	case errors.Is(err, authn.ErrClientTxnConflict):
		writeAPIError(w, r, clientTxnConflict(request.ClientTxnID))
		return
	case errors.Is(err, pgx.ErrNoRows):
		writeAPIError(w, r, incidentNotFoundError())
		return
	case err != nil:
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	if err := s.slideSessionIfNeeded(r.Context(), &principal, r.Method, r.URL.Path); err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	_ = httpapi.WriteSuccess(w, r, result.StatusCode, result.Payload)
}

func (s *Service) handleUploadTarget(w http.ResponseWriter, r *http.Request) {
	token := r.PathValue("upload_token")
	claims, err := decodeObjectUploadToken(s.keys, token)
	if err != nil {
		writeAPIError(w, r, objectUploadNotFoundOrRevoked())
		return
	}
	now := s.now().UTC()
	if !time.Unix(0, claims.ExpiresAtUnixNano).After(now) {
		writeAPIError(w, r, objectUploadExpired("target_expired"))
		return
	}
	blob, err := s.store.GetBlob(r.Context(), claims.ObjectBlobID)
	if errors.Is(err, ErrBlobNotFound) {
		writeAPIError(w, r, objectUploadNotFoundOrRevoked())
		return
	}
	if err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	if blob.IncidentID != claims.IncidentID || blob.StorageKey != claims.StorageKey || blob.ByteSize != claims.ByteSize {
		writeAPIError(w, r, objectUploadRejected(http.StatusConflict, "upload_contract_mismatch", nil))
		return
	}
	if err := validatePersistedObjectBlobStorageKey(blob.StorageKey, blob.IncidentID, blob.ObjectBlobID); err != nil {
		writeAPIError(w, r, objectStoreDependencyAPIError(err))
		return
	}
	if blob.UploadState != "pending" {
		writeAPIError(w, r, objectUploadRejected(http.StatusConflict, "blob_not_pending", nil))
		return
	}
	if !blob.TargetExpiresAt.After(now) || !blob.PendingExpiresAt.After(now) {
		writeAPIError(w, r, objectUploadExpired("slot_expired"))
		return
	}
	if r.ContentLength < 0 {
		writeAPIError(w, r, objectUploadRejected(http.StatusLengthRequired, "content_length_required", nil))
		return
	}
	if r.ContentLength != blob.ByteSize {
		status := http.StatusBadRequest
		reasonCode := "byte_size_mismatch"
		if r.ContentLength > blob.ByteSize {
			status = http.StatusRequestEntityTooLarge
			reasonCode = "byte_size_exceeds_contract"
		}
		writeAPIError(w, r, objectUploadRejected(status, reasonCode, map[string]any{
			"requested_byte_size":  r.ContentLength,
			"contracted_byte_size": blob.ByteSize,
		}))
		return
	}
	contentType := strings.TrimSpace(r.Header.Get("Content-Type"))
	if contentType == "" {
		contentType = firstNonEmptyPtr(blob.ContentTypeHint, nil, "application/octet-stream")
	}
	body := http.MaxBytesReader(w, r.Body, blob.ByteSize)
	defer body.Close()
	if err := s.putObject(r.Context(), blob.StorageKey, body, blob.ByteSize, contentType, objectstore.PurposeProductUpload); err != nil {
		if apiErr := objectStoreDependencyAPIError(err); apiErr != nil {
			writeAPIError(w, r, apiErr)
			return
		}
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Service) handleAttachBlob(w http.ResponseWriter, r *http.Request) {
	recordID, ok := pathUUID(w, r, "record_id")
	if !ok {
		return
	}
	principal, apiErr := s.authenticateSessionRequest(r, true)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	request, apiErr := DecodeAttachBlobRequest(r.Body)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	access, err := s.store.LoadEvidenceAccess(r.Context(), recordID)
	if errors.Is(err, ErrEvidenceNotFound) {
		writeAPIError(w, r, evidenceRecordNotFound())
		return
	}
	if err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	if _, apiErr := s.requireIncidentRole(r.Context(), access.IncidentID, principal.User.ID, "editor", "reviewer", "admin"); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	blob, err := s.store.GetBlob(r.Context(), request.ObjectBlobID)
	if err != nil && !errors.Is(err, ErrBlobNotFound) {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	var observed *ObservedObject
	if err == nil && blob.UploadState == "pending" {
		if err := validatePersistedObjectBlobStorageKey(blob.StorageKey, blob.IncidentID, blob.ObjectBlobID); err != nil {
			writeAPIError(w, r, objectStoreDependencyAPIError(err))
			return
		}
		observed, err = s.observeUploadedObject(r.Context(), blob)
		if err != nil {
			if apiErr := objectStoreDependencyAPIError(err); apiErr != nil {
				writeAPIError(w, r, apiErr)
				return
			}
			observed = nil
		}
	}
	result, err := s.store.AttachBlob(r.Context(), principal.User, recordID, request, AttachBlobRequestHash(request), observed, httpapi.RequestIDFromContext(r.Context()), s.now().UTC())
	var rowConflict *rowVersionConflictError
	var attachRejected AttachRejectedError
	switch {
	case errors.Is(err, authn.ErrClientTxnConflict):
		writeAPIError(w, r, clientTxnConflict(request.ClientTxnID))
		return
	case errors.As(err, &rowConflict):
		writeAPIError(w, r, rowVersionConflict(rowConflict.RecordID, rowConflict.BaseRowVersion, rowConflict.CurrentRowVersion))
		return
	case errors.Is(err, ErrEvidenceNotFound):
		writeAPIError(w, r, evidenceRecordNotFound())
		return
	case errors.As(err, &attachRejected):
		writeAPIError(w, r, evidenceAttachRejected(attachRejected.ReasonCode))
		return
	case errors.Is(err, ErrBlobNotFound):
		writeAPIError(w, r, evidenceAttachRejected(AttachReasonBlobNotVisible))
		return
	case errors.Is(err, ErrIncidentMismatch):
		writeAPIError(w, r, evidenceAttachRejected(AttachReasonBlobNotVisible))
		return
	case errors.Is(err, ErrEvidenceQuarantined):
		writeAPIError(w, r, evidenceAttachRejected(AttachReasonEvidenceQuarantined))
		return
	case errors.Is(err, ErrBlobNotAttachable):
		writeAPIError(w, r, evidenceAttachRejected(AttachReasonEvidenceInconsistent))
		return
	case err != nil:
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	if !result.Replayed {
		s.publishRecordChange(result, principal.User.ID)
	}
	if err := s.slideSessionIfNeeded(r.Context(), &principal, r.Method, r.URL.Path); err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	_ = httpapi.WriteSuccess(w, r, result.StatusCode, result.Payload)
}

func (s *Service) handlePreviewHandle(w http.ResponseWriter, r *http.Request) {
	s.handleIssueHandle(w, r, "preview")
}

func (s *Service) handleDownloadHandle(w http.ResponseWriter, r *http.Request) {
	s.handleIssueHandle(w, r, "download")
}

func (s *Service) handleIssueHandle(w http.ResponseWriter, r *http.Request, kind string) {
	recordID, ok := pathUUID(w, r, "record_id")
	if !ok {
		return
	}
	principal, apiErr := s.authenticateSessionRequest(r, true)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	if apiErr := DecodeHandleIssueRequest(r.Body); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	access, err := s.store.LoadEvidenceAccess(r.Context(), recordID)
	if errors.Is(err, ErrEvidenceNotFound) {
		writeAPIError(w, r, evidenceRecordNotFound())
		return
	}
	if err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	if _, apiErr := s.requireIncidentMembership(r.Context(), access.IncidentID, principal.User.ID); apiErr != nil {
		writeAPIError(w, r, evidenceRecordNotFound())
		return
	}
	if reasonCode := classifyEvidenceAccess(access, nil); reasonCode != "" {
		writeAPIError(w, r, evidenceAccessUnavailable(reasonCode))
		return
	}
	if reasonCode, apiErr := s.verifyEvidenceObjectAvailable(r.Context(), access); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	} else if reasonCode != "" {
		writeAPIError(w, r, evidenceAccessUnavailable(reasonCode))
		return
	}
	if kind == "preview" {
		if access.PreviewKind == nil {
			writeAPIError(w, r, evidenceAccessUnavailable("unsupported_preview"))
			return
		}
		if access.SizeBytes > s.previewMax || (*access.PreviewKind == "text_inline" && access.SizeBytes > s.textPreviewMax) {
			writeAPIError(w, r, evidenceAccessUnavailable("preview_payload_too_large"))
			return
		}
	}
	token, err := randomToken("hdl_")
	if err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	disposition := "attachment"
	singleUse := true
	expiresAt := s.now().UTC().Add(2 * time.Minute)
	var previewKind any
	if kind == "preview" {
		disposition = "inline"
		singleUse = false
		expiresAt = s.now().UTC().Add(5 * time.Minute)
		previewKind = *access.PreviewKind
	}
	filename := sanitizeFilename(access.FilenameSource, recordID, access.ContentType)
	handle := HandleRecord{
		Token: token, IncidentID: access.IncidentID, RecordID: access.RecordID, ObjectBlobID: *access.ObjectBlobID,
		RecordRowVersion: access.RecordRowVersion,
		StorageKey:       *access.StorageKey, SessionID: principal.Session.ID, HandleKind: kind,
		MediaClass: access.MediaClass, PreviewKind: access.PreviewKind, Disposition: disposition,
		Filename: filename, ContentType: access.ContentType, SizeBytes: access.SizeBytes, SHA256: access.SHA256,
		EvidenceLifecycleState: access.EvidenceLifecycleState, UploadState: access.UploadState,
		ExpiresAt: expiresAt,
	}
	if kind == "download" {
		handle.PreviewKind = nil
	}
	if err := s.store.InsertHandle(r.Context(), handle, principal.User.ID); err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	payload := map[string]any{
		"incident_id": access.IncidentID.String(), "record_id": access.RecordID.String(),
		"object_blob_id": access.ObjectBlobID.String(), "handle_kind": kind,
		"href": "/api/v1/evidence-handles/" + url.PathEscape(token), "method": "GET",
		"expires_at": formatHTTPTime(expiresAt), "single_use": singleUse, "media_class": access.MediaClass,
		"disposition": disposition, "filename": filename, "content_type": access.ContentType,
		"size_bytes": access.SizeBytes, "sha256": nullableString(access.SHA256),
		"evidence_lifecycle_state": access.EvidenceLifecycleState, "upload_state": access.UploadState,
	}
	if kind == "preview" {
		payload["preview_kind"] = previewKind
	}
	if err := s.slideSessionIfNeeded(r.Context(), &principal, r.Method, r.URL.Path); err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	_ = httpapi.WriteSuccess(w, r, http.StatusOK, payload)
}

func (s *Service) handleRedeemHandle(w http.ResponseWriter, r *http.Request) {
	principal, apiErr := s.authenticateSessionRequest(r, false)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	token := r.PathValue("handle_token")
	handle, err := s.store.LoadHandle(r.Context(), token)
	if errors.Is(err, ErrBlobNotFound) {
		writeAPIError(w, r, &auth.APIError{Status: http.StatusNotFound, Code: "handle_not_found_or_revoked", Details: map[string]any{}})
		return
	}
	if err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	now := s.now().UTC()
	if handle.ExpiresAt.Before(now) || handle.ExpiresAt.Equal(now) {
		writeAPIError(w, r, &auth.APIError{Status: http.StatusGone, Code: "handle_expired", Details: map[string]any{}})
		return
	}
	if handle.HandleKind == "download" && handle.ConsumedAt != nil {
		writeAPIError(w, r, &auth.APIError{Status: http.StatusGone, Code: "handle_consumed", Details: map[string]any{}})
		return
	}
	if handle.SessionID != principal.Session.ID {
		writeAPIError(w, r, handleNotFoundOrRevoked())
		return
	}
	if _, apiErr := s.requireIncidentMembership(r.Context(), handle.IncidentID, principal.User.ID); apiErr != nil {
		writeAPIError(w, r, handleNotFoundOrRevoked())
		return
	}
	if reasonCode, err := s.store.CheckHandleAccess(r.Context(), handle); err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	} else if reasonCode != "" {
		writeAPIError(w, r, evidenceAccessUnavailable(reasonCode))
		return
	}
	if err := validatePersistedObjectBlobStorageKey(handle.StorageKey, handle.IncidentID, handle.ObjectBlobID); err != nil {
		writeAPIError(w, r, objectStoreDependencyAPIError(err))
		return
	}
	readOptions := objectstore.ReadOptions{}
	status := http.StatusOK
	contentRange := ""
	if rangeHeader := r.Header.Get("Range"); rangeHeader != "" {
		if start, end, ok := parseByteRange(rangeHeader, handle.SizeBytes); ok {
			readOptions.RangeStart = &start
			readOptions.RangeEnd = &end
			status = http.StatusPartialContent
			contentRange = fmt.Sprintf("bytes %d-%d/%d", start, end, handle.SizeBytes)
		}
	}
	object, _, err := s.getObject(r.Context(), handle.StorageKey, readOptions, objectstore.PurposeProductRead)
	if err != nil {
		if apiErr := objectStoreDependencyAPIError(err); apiErr != nil {
			writeAPIError(w, r, apiErr)
			return
		}
		if adapterErr, ok := objectstore.AsAdapterError(err); ok && adapterErr.Code == objectstore.ErrorCodeRangeNotSatisfiable {
			writeAPIError(w, r, evidenceAccessUnavailable("evidence_inconsistent"))
			return
		}
		writeAPIError(w, r, evidenceAccessUnavailable("blob_missing"))
		return
	}
	defer object.Close()
	if handle.HandleKind == "download" {
		if err := s.store.ConsumeDownloadHandle(r.Context(), token, now); err != nil {
			writeAPIError(w, r, &auth.APIError{Status: http.StatusGone, Code: "handle_consumed", Details: map[string]any{}})
			return
		}
	}
	w.Header().Set("Content-Type", handle.ContentType)
	w.Header().Set("Content-Disposition", formatContentDisposition(handle.Disposition, handle.Filename))
	if contentRange != "" {
		w.Header().Set("Content-Range", contentRange)
	}
	if handle.SHA256 != nil {
		w.Header().Set("Digest", "sha-256="+*handle.SHA256)
	}
	w.WriteHeader(status)
	_, _ = io.Copy(w, object)
}

func (s *Service) observeUploadedObject(ctx context.Context, blob BlobRecord) (*ObservedObject, error) {
	if err := validatePersistedObjectBlobStorageKey(blob.StorageKey, blob.IncidentID, blob.ObjectBlobID); err != nil {
		return nil, err
	}
	stat, err := s.headObject(ctx, blob.StorageKey, objectstore.PurposeProductUpload)
	if err != nil {
		return nil, err
	}
	object, _, err := s.getObject(ctx, blob.StorageKey, objectstore.ReadOptions{}, objectstore.PurposeProductRead)
	if err != nil {
		return nil, err
	}
	defer object.Close()
	hash := sha256.New()
	if _, err := io.Copy(hash, object); err != nil {
		return nil, err
	}
	contentType := strings.TrimSpace(stat.ContentType)
	if contentType == "" {
		contentType = firstNonEmptyPtr(blob.ContentTypeHint, nil, "application/octet-stream")
	}
	return &ObservedObject{Size: stat.Size, ContentType: contentType, SHA256Hex: fmt.Sprintf("%x", hash.Sum(nil))}, nil
}

func (s *Service) verifyEvidenceObjectAvailable(ctx context.Context, access EvidenceAccessRecord) (string, *auth.APIError) {
	if access.StorageKey == nil {
		return "evidence_inconsistent", nil
	}
	if access.ObjectBlobID == nil {
		return "evidence_inconsistent", nil
	}
	if err := validatePersistedObjectBlobStorageKey(*access.StorageKey, access.IncidentID, *access.ObjectBlobID); err != nil {
		return "", objectStoreDependencyAPIError(err)
	}
	if _, err := s.headObject(ctx, *access.StorageKey, objectstore.PurposeProductRead); err != nil {
		if apiErr := objectStoreDependencyAPIError(err); apiErr != nil {
			return "", apiErr
		}
		return "blob_missing", nil
	}
	return "", nil
}

func (s *Service) createObjectUploadTarget(claims objectUploadTokenClaims) (objectstore.UploadTarget, error) {
	token, err := encodeObjectUploadToken(s.keys, claims)
	if err != nil {
		return objectstore.UploadTarget{}, err
	}
	return objectstore.UploadTarget{
		Href:    "/api/v1/object-uploads/" + url.PathEscape(token),
		Method:  "PUT",
		Headers: map[string]string{},
	}, nil
}

func (s *Service) putObject(ctx context.Context, key string, body io.Reader, size int64, contentType string, purpose objectstore.Purpose) error {
	if typed, ok := s.objectStore.(objectstore.TypedStore); ok {
		_, err := typed.Put(ctx, objectstore.PutObjectRequest{
			Key:         key,
			Body:        body,
			Size:        size,
			ContentType: contentType,
			Purpose:     purpose,
		})
		return err
	}
	return s.objectStore.PutObject(ctx, key, body, size, contentType)
}

func (s *Service) headObject(ctx context.Context, key string, purpose objectstore.Purpose) (objectstore.ObjectInfo, error) {
	if typed, ok := s.objectStore.(objectstore.TypedStore); ok {
		return typed.Head(ctx, objectstore.HeadObjectRequest{Key: key, Purpose: purpose})
	}
	return s.objectStore.StatObject(ctx, key)
}

func (s *Service) getObject(ctx context.Context, key string, options objectstore.ReadOptions, purpose objectstore.Purpose) (io.ReadCloser, objectstore.ObjectInfo, error) {
	if typed, ok := s.objectStore.(objectstore.TypedStore); ok {
		return typed.Get(ctx, objectstore.GetObjectRequest{Key: key, RangeStart: options.RangeStart, RangeEnd: options.RangeEnd, Purpose: purpose})
	}
	return s.objectStore.ReadObject(ctx, key, options)
}

func (s *Service) publishRecordChange(result AttachBlobResult, actorUserID uuid.UUID) {
	if s.hub == nil || result.RecordID == uuid.Nil || result.ChangeSetID == uuid.Nil {
		return
	}
	changes := append([]AttachRecordChange{{
		RecordID: result.RecordID, RowVersion: result.RowVersion,
		ViewSchemaID: evidenceViewSchemaID, ChangedFieldKeys: result.ChangedFieldKeys,
	}}, result.AffectedRecordChanges...)
	for _, change := range changes {
		changedKeys := append([]string(nil), change.ChangedFieldKeys...)
		slices.Sort(changedKeys)
		s.hub.PublishRecordChange(platformws.RecordChange{
			IncidentID: result.IncidentID, RecordID: change.RecordID, RowVersion: change.RowVersion,
			ChangeSetID: result.ChangeSetID, ClientTxnID: result.ClientTxnID, ActorUserID: actorUserID,
			ChangedFieldKeys: changedKeys, ViewSchemaID: change.ViewSchemaID,
		})
	}
}

func parseByteRange(value string, size int64) (int64, int64, bool) {
	if !strings.HasPrefix(value, "bytes=") || size <= 0 {
		return 0, 0, false
	}
	parts := strings.SplitN(strings.TrimPrefix(value, "bytes="), "-", 2)
	if len(parts) != 2 || parts[0] == "" {
		return 0, 0, false
	}
	start, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil || start < 0 || start >= size {
		return 0, 0, false
	}
	end := size - 1
	if parts[1] != "" {
		end, err = strconv.ParseInt(parts[1], 10, 64)
		if err != nil || end < start {
			return 0, 0, false
		}
	}
	if end >= size {
		end = size - 1
	}
	return start, end, true
}

func (s *Service) requireIncidentMembership(ctx context.Context, incidentID uuid.UUID, userID uuid.UUID) (incidents.MembershipRecord, *auth.APIError) {
	record, err := s.incidentStore.GetIncidentMembershipForUser(ctx, incidentID, userID)
	if errors.Is(err, incidents.ErrMembershipNotFound) {
		return incidents.MembershipRecord{}, incidentNotFoundError()
	}
	if err != nil {
		return incidents.MembershipRecord{}, internalAPIError(err)
	}
	return record, nil
}

func (s *Service) requireIncidentRole(ctx context.Context, incidentID uuid.UUID, userID uuid.UUID, roles ...string) (incidents.MembershipRecord, *auth.APIError) {
	record, apiErr := s.requireIncidentMembership(ctx, incidentID, userID)
	if apiErr != nil {
		return incidents.MembershipRecord{}, apiErr
	}
	if !slices.Contains(roles, record.Role) {
		return incidents.MembershipRecord{}, &auth.APIError{Status: http.StatusForbidden, Code: "authorization_denied", Details: map[string]any{"required_role": strings.Join(roles, "|")}}
	}
	return record, nil
}

func (s *Service) authenticateSessionRequest(r *http.Request, stateChanging bool) (auth.SessionPrincipal, *auth.APIError) {
	header := r.Header.Get("Authorization")
	if strings.HasPrefix(header, "Bearer ") {
		token := strings.TrimSpace(strings.TrimPrefix(header, "Bearer "))
		if token == "" {
			return auth.SessionPrincipal{}, &auth.APIError{Status: http.StatusUnauthorized, Code: "session_required", Details: map[string]any{}}
		}
		return s.authenticateSessionToken(r, auth.AuthSourceBearer, token, stateChanging)
	}
	cookie, err := r.Cookie(authn.SessionCookieName)
	if err != nil {
		if errors.Is(err, http.ErrNoCookie) {
			return auth.SessionPrincipal{}, &auth.APIError{Status: http.StatusUnauthorized, Code: "session_required", Details: map[string]any{}}
		}
		return auth.SessionPrincipal{}, internalAPIError(err)
	}
	return s.authenticateSessionToken(r, auth.AuthSourceCookie, cookie.Value, stateChanging)
}

func (s *Service) authenticateSessionToken(r *http.Request, authSource auth.AuthSource, sessionToken string, stateChanging bool) (auth.SessionPrincipal, *auth.APIError) {
	if stateChanging && authSource == auth.AuthSourceCookie {
		csrfCookie, _ := r.Cookie(authn.CSRFCookieName)
		if csrfCookie == nil || csrfCookie.Value != authn.CSRFTokenForSessionToken(s.keys, sessionToken) {
			return auth.SessionPrincipal{}, &auth.APIError{Status: http.StatusForbidden, Code: "csrf_verification_failed", Details: map[string]any{}}
		}
		if apiErr := auth.ValidateCSRF(r.Method, authSource, csrfCookie.Value, r.Header.Get(authn.CSRFHeaderName)); apiErr != nil {
			return auth.SessionPrincipal{}, apiErr
		}
	}
	session, user, err := s.authStore.GetSessionByFingerprint(r.Context(), authn.FingerprintToken(s.keys, sessionToken))
	if errors.Is(err, authn.ErrNotFound) {
		return auth.SessionPrincipal{}, &auth.APIError{Status: http.StatusUnauthorized, Code: "session_required", Details: map[string]any{}}
	}
	if err != nil {
		return auth.SessionPrincipal{}, internalAPIError(err)
	}
	if !user.IsActive || session.RevokedAt != nil {
		return auth.SessionPrincipal{}, &auth.APIError{Status: http.StatusUnauthorized, Code: "session_required", Details: map[string]any{}}
	}
	now := s.now()
	if !session.SessionExpiresAt.After(now) {
		_ = s.authStore.RevokeSession(context.Background(), session.ID, "session_expired", now)
		return auth.SessionPrincipal{}, &auth.APIError{Status: http.StatusUnauthorized, Code: "session_required", Details: map[string]any{}}
	}
	return auth.SessionPrincipal{AuthSource: authSource, SessionToken: sessionToken, Session: session, User: user}, nil
}

func (s *Service) slideSessionIfNeeded(ctx context.Context, principal *auth.SessionPrincipal, method string, path string) error {
	if principal == nil || !auth.ShouldSlideIdleExpiry(method, path) {
		return nil
	}
	sliding := authn.SessionTiming{
		AuthenticatedAt: principal.Session.AuthenticatedAt, LastQualifyingActivityAt: principal.Session.LastQualifyingActivityAt,
		IdleExpiresAt: principal.Session.IdleExpiresAt, AbsoluteExpiresAt: principal.Session.AbsoluteExpiresAt,
		SessionExpiresAt: principal.Session.SessionExpiresAt,
	}.Slide(s.now())
	persisted, err := s.authStore.SlideSession(ctx, principal.Session.ID, sliding)
	if err != nil {
		return err
	}
	principal.Session.LastQualifyingActivityAt = persisted.LastQualifyingActivityAt
	principal.Session.IdleExpiresAt = persisted.IdleExpiresAt
	principal.Session.SessionExpiresAt = persisted.SessionExpiresAt
	return nil
}

func writeAPIError(w http.ResponseWriter, r *http.Request, apiErr *auth.APIError) {
	message := apiErr.Message
	if message == "" {
		message = apiErr.Code
	}
	_ = httpapi.WriteErrorWithOptions(w, r, apiErr.Status, apiErr.Code, message, apiErr.Details, httpapi.ErrorOptions{
		Conflict:  apiErr.Conflict,
		Retryable: apiErr.Retryable || (apiErr.Status == http.StatusConflict && apiErr.Code == "record_locked"),
	})
}

func pathUUID(w http.ResponseWriter, r *http.Request, key string) (uuid.UUID, bool) {
	value, err := uuid.Parse(r.PathValue(key))
	if err != nil {
		http.NotFound(w, r)
		return uuid.UUID{}, false
	}
	return value, true
}

func internalAPIError(err error) *auth.APIError {
	return &auth.APIError{Status: http.StatusInternalServerError, Code: "internal_error", Message: err.Error(), Details: map[string]any{}}
}

const (
	objectBlobStorageKeyMalformedReason        = "object_blob_storage_key_malformed"
	objectBlobStorageKeyIdentityMismatchReason = "object_blob_storage_key_identity_mismatch"
)

type persistedObjectBlobStorageKeyError struct {
	reasonCode string
}

func (e *persistedObjectBlobStorageKeyError) Error() string {
	return "persisted object blob storage_key violates object_blob_storage_key_v1"
}

func validatePersistedObjectBlobStorageKey(key string, incidentID uuid.UUID, objectBlobID uuid.UUID) error {
	parts, err := blobref.ParseObjectBlobStorageKey(key)
	if err != nil {
		return &persistedObjectBlobStorageKeyError{reasonCode: objectBlobStorageKeyMalformedReason}
	}
	if parts.IncidentID != incidentID || parts.ObjectBlobID != objectBlobID {
		return &persistedObjectBlobStorageKeyError{reasonCode: objectBlobStorageKeyIdentityMismatchReason}
	}
	return nil
}

func objectStoreDependencyAPIError(err error) *auth.APIError {
	var storageKeyErr *persistedObjectBlobStorageKeyError
	if errors.As(err, &storageKeyErr) {
		return objectStoreInvalidRequestAPIError(storageKeyErr.reasonCode)
	}
	adapterErr, ok := objectstore.AsAdapterError(err)
	if !ok {
		return nil
	}
	switch adapterErr.Code {
	case objectstore.ErrorCodeUnavailable:
		return objectStoreUnavailableAPIError(unavailableReason(adapterErr), true)
	case objectstore.ErrorCodeDeadlineExceeded, objectstore.ErrorCodeRetryExhausted:
		return objectStoreUnavailableAPIError("retry_exhausted", true)
	case objectstore.ErrorCodeAccessRejected:
		return &auth.APIError{
			Status: http.StatusServiceUnavailable,
			Code:   "object_store_access_rejected",
			Details: map[string]any{
				"reason_code": accessRejectedReason(adapterErr),
			},
		}
	default:
		return nil
	}
}

func objectStoreInvalidRequestAPIError(reasonCode string) *auth.APIError {
	return &auth.APIError{
		Status: http.StatusInternalServerError,
		Code:   "object_store_invalid_request",
		Details: map[string]any{
			"reason_code": reasonCode,
		},
	}
}

func objectStoreUnavailableAPIError(reasonCode string, retryable bool) *auth.APIError {
	return &auth.APIError{
		Status:    http.StatusServiceUnavailable,
		Code:      "object_store_unavailable",
		Retryable: retryable,
		Details: map[string]any{
			"reason_code": reasonCode,
		},
	}
}

func objectUploadNotFoundOrRevoked() *auth.APIError {
	return &auth.APIError{Status: http.StatusNotFound, Code: "object_upload_not_found_or_revoked", Details: map[string]any{}}
}

func objectUploadExpired(reasonCode string) *auth.APIError {
	return &auth.APIError{Status: http.StatusGone, Code: "object_upload_expired", Details: map[string]any{"reason_code": reasonCode}}
}

func objectUploadRejected(status int, reasonCode string, details map[string]any) *auth.APIError {
	if details == nil {
		details = map[string]any{}
	}
	details["reason_code"] = reasonCode
	return &auth.APIError{Status: status, Code: "object_upload_rejected", Details: details}
}

func unavailableReason(adapterErr *objectstore.AdapterError) string {
	switch adapterErr.Reason {
	case objectstore.ReasonBucketMissing:
		return "bucket_missing"
	case objectstore.ReasonRetryExhausted, objectstore.ReasonDeadlineExceeded:
		return "retry_exhausted"
	default:
		return "endpoint_unreachable"
	}
}

func accessRejectedReason(adapterErr *objectstore.AdapterError) string {
	switch adapterErr.Reason {
	case objectstore.ReasonCredentialDenied:
		return "credential_denied"
	case objectstore.ReasonCORSRejected:
		return "cors_rejected"
	default:
		return "capability_missing"
	}
}

func incidentNotFoundError() *auth.APIError {
	return &auth.APIError{Status: http.StatusNotFound, Code: "incident_not_found", Details: map[string]any{}}
}

func evidenceRecordNotFound() *auth.APIError {
	return &auth.APIError{Status: http.StatusNotFound, Code: "evidence_record_not_found", Details: map[string]any{}}
}
