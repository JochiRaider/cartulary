package evidence

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	evidencepolicy "github.com/JochiRaider/cartulary/internal/modules/evidence/internal/policy"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/objectstore"
)

func (s *service) handlePreviewHandle(w http.ResponseWriter, r *http.Request) {
	s.handleIssueHandle(w, r, "preview")
}

func (s *service) handleDownloadHandle(w http.ResponseWriter, r *http.Request) {
	s.handleIssueHandle(w, r, "download")
}

func (s *service) handleIssueHandle(w http.ResponseWriter, r *http.Request, kind string) {
	principal, apiErr := s.admission.authenticate(r, true)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	recordID, ok := pathUUID(w, r, "record_id")
	if !ok {
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
	if _, apiErr := s.admission.requireRole(r.Context(), access.IncidentID, principal.User.ID, "viewer", "editor", "reviewer", "admin"); apiErr != nil {
		writeAPIError(w, r, evidenceRecordNotFound())
		return
	}
	if apiErr := decodeHandleIssueRequest(r.Body); apiErr != nil {
		writeAPIError(w, r, apiErr)
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
	handle := handleRecord{
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

func (s *service) handleRedeemHandle(w http.ResponseWriter, r *http.Request) {
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
	if handle.SessionID != principal.Session.ID {
		writeAPIError(w, r, handleNotFoundOrRevoked())
		return
	}
	if _, apiErr := s.admission.requireRole(r.Context(), handle.IncidentID, principal.User.ID, "viewer", "editor", "reviewer", "admin"); apiErr != nil {
		writeAPIError(w, r, handleNotFoundOrRevoked())
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
	if reasonCode, err := s.operations.CheckHandleAccess(r.Context(), handle); err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	} else if reasonCode != "" {
		writeAPIError(w, r, evidenceAccessUnavailable(reasonCode))
		return
	}
	if err := evidencepolicy.ValidatePersistedObjectBlobStorageKey(handle.StorageKey, handle.IncidentID, handle.ObjectBlobID); err != nil {
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
