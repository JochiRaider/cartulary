package evidence

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/evidence/blobref"
	"github.com/JochiRaider/cartulary/internal/modules/incidents"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/objectstore"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type Service struct {
	operations     RouteService
	admission      routeAdmission
	objects        routeObjectStoreAdapter
	uploads        uploadCapabilityService
	keys           authn.MasterKeys
	now            func() time.Time
	maxBlobBytes   int64
	previewMax     int64
	textPreviewMax int64
}

type Settings struct {
	MaxBlobBytes   int64
	PreviewMax     int64
	TextPreviewMax int64
}

type RouteOption func(*routeOptions)

type routeOptions struct {
	service RouteService
}

// RouteService is the narrow application capability required by Evidence
// transport. It exposes no database handle or provider construction surface.
type RouteService interface {
	CreateBlobSlot(context.Context, BlobSlotParams) (BlobSlotResult, error)
	GetBlob(context.Context, uuid.UUID) (BlobRecord, error)
	AttachBlob(context.Context, authn.UserRecord, uuid.UUID, AttachBlobRequest, []byte, *ObservedObject, string, time.Time) (AttachBlobResult, error)
	LoadEvidenceAccess(context.Context, uuid.UUID) (EvidenceAccessRecord, error)
	InsertHandle(context.Context, HandleRecord, uuid.UUID) error
	LoadHandle(context.Context, string) (HandleRecord, error)
	CheckHandleAccess(context.Context, HandleRecord) (string, error)
	ConsumeDownloadHandle(context.Context, string, time.Time) error
}

var _ RouteService = (*Store)(nil)

// WithRouteService injects the immutable Evidence application capability used
// by transport.
func WithRouteService(service RouteService) RouteOption {
	return func(options *routeOptions) {
		options.service = service
	}
}

func RegisterRoutes(settings Settings, options ...RouteOption) httpapi.RouteRegistrar {
	resolved := Settings{}
	resolved = settings
	resolvedOptions := routeOptions{}
	for _, option := range options {
		if option != nil {
			option(&resolvedOptions)
		}
	}
	return func(mux *http.ServeMux, deps httpapi.DependencySet) error {
		service, err := newService(deps, resolved, resolvedOptions)
		if err != nil {
			return err
		}
		return httpapi.BindOwnerRoutes(mux, deps, "module.evidence", map[string]http.HandlerFunc{
			"attachBlobToEvidenceRecord":  service.handleAttachBlob,
			"createObjectBlobSlot":        service.handleCreateBlob,
			"issueEvidenceDownloadHandle": service.handleDownloadHandle,
			"issueEvidencePreviewHandle":  service.handlePreviewHandle,
			"redeemEvidenceHandle":        service.handleRedeemHandle,
			"uploadObjectBlobContent":     service.handleUploadTarget,
		})
	}
}

func newService(deps httpapi.DependencySet, settings Settings, options routeOptions) (*Service, error) {
	keys, err := authn.LoadMasterKeys(deps.Env)
	if err != nil {
		return nil, err
	}
	now := deps.Now
	if now == nil {
		now = func() time.Time { return time.Now().UTC() }
	}
	operations := options.service
	if operations == nil {
		return nil, fmt.Errorf("compose Evidence routes: RouteService is required")
	}
	return &Service{
		operations: operations,
		admission: routeAdmission{
			incidents: incidents.NewAccess(deps.PostgresHandle()),
			auth:      authn.NewStore(deps.PostgresHandle()),
			keys:      keys,
			now:       now,
		},
		objects:        routeObjectStoreAdapter{store: deps.ObjectStore},
		uploads:        uploadCapabilityService{keys: keys},
		keys:           keys,
		now:            now,
		maxBlobBytes:   settings.MaxBlobBytes,
		previewMax:     settings.PreviewMax,
		textPreviewMax: settings.TextPreviewMax,
	}, nil
}

func (s *Service) handleCreateBlob(w http.ResponseWriter, r *http.Request) {
	principal, apiErr := s.admission.authenticate(r, true)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	request, apiErr := DecodeBlobCreateRequest(r.Body, s.maxBlobBytes)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	if _, apiErr := s.admission.requireRole(r.Context(), request.IncidentID, principal.User.ID, "editor", "reviewer", "admin"); apiErr != nil {
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
	target, err := s.uploads.createTarget(objectUploadTokenClaims{
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
	target.Headers = copyUploadTargetHeaders(target.Headers)
	target.Headers["Content-Type"] = firstNonEmptyPtr(request.ContentTypeHint, nil, "application/octet-stream")
	result, err := s.operations.CreateBlobSlot(r.Context(), BlobSlotParams{
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
	if err := s.admission.slide(r.Context(), &principal, r.Method, r.URL.Path); err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	_ = httpapi.WriteSuccess(w, r, result.StatusCode, result.Payload)
}

func copyUploadTargetHeaders(source map[string]string) map[string]string {
	result := make(map[string]string, len(source)+1)
	for name, value := range source {
		result[name] = value
	}
	return result
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
	blob, err := s.operations.GetBlob(r.Context(), claims.ObjectBlobID)
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
	if err := s.objects.put(r.Context(), blob.StorageKey, body, blob.ByteSize, contentType, objectstore.PurposeProductUpload); err != nil {
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
	principal, apiErr := s.admission.authenticate(r, true)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	request, apiErr := DecodeAttachBlobRequest(r.Body)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	access, err := s.operations.LoadEvidenceAccess(r.Context(), recordID)
	if errors.Is(err, ErrEvidenceNotFound) {
		writeAPIError(w, r, evidenceRecordNotFound())
		return
	}
	if err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	if _, apiErr := s.admission.requireRole(r.Context(), access.IncidentID, principal.User.ID, "editor", "reviewer", "admin"); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	blob, err := s.operations.GetBlob(r.Context(), request.ObjectBlobID)
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
		observed, err = s.objects.observeUploadedObject(r.Context(), blob)
		if err != nil {
			if apiErr := objectStoreDependencyAPIError(err); apiErr != nil {
				writeAPIError(w, r, apiErr)
				return
			}
			observed = nil
		}
	}
	result, err := s.operations.AttachBlob(r.Context(), principal.User, recordID, request, AttachBlobRequestHash(request), observed, httpapi.RequestIDFromContext(r.Context()), s.now().UTC())
	if apiErr := translateAttachError(err, request.ClientTxnID); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	if err := s.admission.slide(r.Context(), &principal, r.Method, r.URL.Path); err != nil {
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
	principal, apiErr := s.admission.authenticate(r, true)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	if apiErr := DecodeHandleIssueRequest(r.Body); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	access, err := s.operations.LoadEvidenceAccess(r.Context(), recordID)
	if errors.Is(err, ErrEvidenceNotFound) {
		writeAPIError(w, r, evidenceRecordNotFound())
		return
	}
	if err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	if _, apiErr := s.admission.requireMembership(r.Context(), access.IncidentID, principal.User.ID); apiErr != nil {
		writeAPIError(w, r, evidenceRecordNotFound())
		return
	}
	if reasonCode := classifyEvidenceAccess(access, nil); reasonCode != "" {
		writeAPIError(w, r, evidenceAccessUnavailable(reasonCode))
		return
	}
	if reasonCode, apiErr := s.objects.verifyEvidenceObjectAvailable(r.Context(), access); apiErr != nil {
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
	if err := s.operations.InsertHandle(r.Context(), handle, principal.User.ID); err != nil {
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
	if err := s.admission.slide(r.Context(), &principal, r.Method, r.URL.Path); err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	_ = httpapi.WriteSuccess(w, r, http.StatusOK, payload)
}

func (s *Service) handleRedeemHandle(w http.ResponseWriter, r *http.Request) {
	principal, apiErr := s.admission.authenticate(r, false)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	token := r.PathValue("handle_token")
	handle, err := s.operations.LoadHandle(r.Context(), token)
	if errors.Is(err, ErrBlobNotFound) {
		writeAPIError(w, r, &httpapi.APIError{Status: http.StatusNotFound, Code: "handle_not_found_or_revoked", Details: map[string]any{}})
		return
	}
	if err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	now := s.now().UTC()
	if handle.ExpiresAt.Before(now) || handle.ExpiresAt.Equal(now) {
		writeAPIError(w, r, &httpapi.APIError{Status: http.StatusGone, Code: "handle_expired", Details: map[string]any{}})
		return
	}
	if handle.HandleKind == "download" && handle.ConsumedAt != nil {
		writeAPIError(w, r, &httpapi.APIError{Status: http.StatusGone, Code: "handle_consumed", Details: map[string]any{}})
		return
	}
	if handle.SessionID != principal.Session.ID {
		writeAPIError(w, r, handleNotFoundOrRevoked())
		return
	}
	if _, apiErr := s.admission.requireMembership(r.Context(), handle.IncidentID, principal.User.ID); apiErr != nil {
		writeAPIError(w, r, handleNotFoundOrRevoked())
		return
	}
	if reasonCode, err := s.operations.CheckHandleAccess(r.Context(), handle); err != nil {
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
	object, _, err := s.objects.get(r.Context(), handle.StorageKey, readOptions, objectstore.PurposeProductRead)
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
		if err := s.operations.ConsumeDownloadHandle(r.Context(), token, now); err != nil {
			writeAPIError(w, r, &httpapi.APIError{Status: http.StatusGone, Code: "handle_consumed", Details: map[string]any{}})
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

func writeAPIError(w http.ResponseWriter, r *http.Request, apiErr *httpapi.APIError) {
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

func internalAPIError(err error) *httpapi.APIError {
	return &httpapi.APIError{Status: http.StatusInternalServerError, Code: "internal_error", Message: err.Error(), Details: map[string]any{}}
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

func PersistedObjectBlobStorageKeyErrorReason(err error) (string, bool) {
	var storageKeyError *persistedObjectBlobStorageKeyError
	if !errors.As(err, &storageKeyError) {
		return "", false
	}
	return storageKeyError.reasonCode, true
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

func objectStoreDependencyAPIError(err error) *httpapi.APIError {
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
		return &httpapi.APIError{
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

func objectStoreInvalidRequestAPIError(reasonCode string) *httpapi.APIError {
	return &httpapi.APIError{
		Status: http.StatusInternalServerError,
		Code:   "object_store_invalid_request",
		Details: map[string]any{
			"reason_code": reasonCode,
		},
	}
}

func objectStoreUnavailableAPIError(reasonCode string, retryable bool) *httpapi.APIError {
	return &httpapi.APIError{
		Status:    http.StatusServiceUnavailable,
		Code:      "object_store_unavailable",
		Retryable: retryable,
		Details: map[string]any{
			"reason_code": reasonCode,
		},
	}
}

func objectUploadNotFoundOrRevoked() *httpapi.APIError {
	return &httpapi.APIError{Status: http.StatusNotFound, Code: "object_upload_not_found_or_revoked", Details: map[string]any{}}
}

func objectUploadExpired(reasonCode string) *httpapi.APIError {
	return &httpapi.APIError{Status: http.StatusGone, Code: "object_upload_expired", Details: map[string]any{"reason_code": reasonCode}}
}

func objectUploadRejected(status int, reasonCode string, details map[string]any) *httpapi.APIError {
	if details == nil {
		details = map[string]any{}
	}
	details["reason_code"] = reasonCode
	return &httpapi.APIError{Status: status, Code: "object_upload_rejected", Details: details}
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

func incidentNotFoundError() *httpapi.APIError {
	return &httpapi.APIError{Status: http.StatusNotFound, Code: "incident_not_found", Details: map[string]any{}}
}

func incidentClosedError() *httpapi.APIError {
	return &httpapi.APIError{Status: http.StatusConflict, Code: "incident_closed", Message: "incident closed", Details: map[string]any{}}
}

func evidenceRecordNotFound() *httpapi.APIError {
	return &httpapi.APIError{Status: http.StatusNotFound, Code: "evidence_record_not_found", Details: map[string]any{}}
}
