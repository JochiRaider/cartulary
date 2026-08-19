package evidence

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/evidence/blobref"
	evidencepolicy "github.com/JochiRaider/cartulary/internal/modules/evidence/internal/policy"
	"github.com/JochiRaider/cartulary/internal/modules/incidents/admission"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/objectstore"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (s *service) handleCreateBlob(w http.ResponseWriter, r *http.Request) {
	principal, apiErr := s.admission.authenticate(r, true)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	body, apiErr := readBoundedBlobCreateRequest(r.Body)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	incidentID, apiErr := decodeBlobCreateIncidentID(body)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	incident, apiErr := s.admission.visibleIncident(r.Context(), incidentID, principal.User.ID)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	if _, apiErr := s.admission.requireRole(r.Context(), incidentID, principal.User.ID, admission.RolesEditorReviewerAdmin, "editor|reviewer|admin"); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	request, apiErr := decodeBlobCreateRequest(bytes.NewReader(body), s.maxBlobBytes)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	if incident.IncidentStatus == admission.IncidentStatusClosed {
		writeAPIError(w, r, incidentClosedError())
		return
	}
	now := s.now().UTC().Truncate(time.Microsecond)
	objectBlobID := uuid.New()
	leaseID := uuid.New()
	storageKey, err := blobref.ObjectBlobStorageKey(request.IncidentID, objectBlobID)
	if err != nil {
		writeAPIError(w, r, internalAPIError(fmt.Errorf("build object blob storage key: %w", err)))
		return
	}
	targetExpiresAt := now.Add(60 * time.Minute)
	pendingExpiresAt := now.Add(24 * time.Hour)
	requiredHeaders := map[string]string{
		"Content-Type": firstNonEmptyPtr(request.ContentTypeHint, nil, "application/octet-stream"),
	}
	requiredHeadersDigest, requiredHeadersDigestHex, err := objectUploadBindingDigest(requiredHeaders)
	if err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	_ = requiredHeadersDigest
	acceptedContractDigest, acceptedContractDigestHex, err := objectUploadBindingDigest(request.AcceptedContract)
	if err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	expectedSHA256Hex := ""
	if request.SHA256Hex != nil {
		expectedSHA256Hex = *request.SHA256Hex
	}
	createdTarget, err := s.uploads.createTarget(objectUploadTokenClaims{
		Version:                objectUploadTokenVersion,
		LeaseID:                leaseID,
		ObjectBlobID:           objectBlobID,
		IncidentID:             request.IncidentID,
		IssuingUserID:          principal.User.ID,
		IssuingSessionID:       principal.Session.ID,
		StorageKey:             storageKey,
		ByteSize:               request.ByteSize,
		ExpectedSHA256Hex:      expectedSHA256Hex,
		RequiredMethod:         http.MethodPut,
		RequiredHeadersSHA256:  requiredHeadersDigestHex,
		AcceptedContractSHA256: acceptedContractDigestHex,
		IssuedAtUnixNano:       now.UnixNano(),
		ExpiresAtUnixNano:      targetExpiresAt.UnixNano(),
	})
	if err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	target := createdTarget.Target
	target.Headers = copyUploadTargetHeaders(requiredHeaders)
	result, err := s.operations.CreateBlobSlot(r.Context(), blobSlotParams{
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
		RequestHash:      blobCreateRequestHash(request),
		ClientTxnID:      request.ClientTxnID,
		UploadLease: uploadLeaseCreateParams{
			LeaseID: leaseID, CapabilityHash: objectUploadTokenDigest(createdTarget.Token),
			IssuingUserID: principal.User.ID, IssuingSessionID: principal.Session.ID,
			IssuedAt: now, ExpiresAt: targetExpiresAt, RequiredMethod: http.MethodPut,
			RequiredHeaders: requiredHeaders, AcceptedContractSHA256: acceptedContractDigest,
		},
	})
	switch {
	case errors.Is(err, authn.ErrClientTxnConflict):
		writeAPIError(w, r, clientTxnConflict(request.ClientTxnID))
		return
	case errors.Is(err, pgx.ErrNoRows):
		writeAPIError(w, r, incidentNotFoundError())
		return
	case admission.IsDenied(err, admission.DenialIncidentClosed):
		writeAPIError(w, r, incidentClosedError())
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

func (s *service) handleUploadTarget(w http.ResponseWriter, r *http.Request) {
	principal, apiErr := s.admission.authenticate(r, true)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	token := r.PathValue("upload_token")
	claims, err := decodeObjectUploadToken(s.keys, token)
	if err != nil {
		writeAPIError(w, r, objectUploadNotFoundOrRevoked())
		return
	}
	incident, apiErr := s.admission.visibleIncident(r.Context(), claims.IncidentID, principal.User.ID)
	if apiErr != nil {
		writeAPIError(w, r, objectUploadNotFoundOrRevoked())
		return
	}
	if _, apiErr := s.admission.requireRole(r.Context(), claims.IncidentID, principal.User.ID, admission.RolesEditorReviewerAdmin, "editor|reviewer|admin"); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	lease, err := s.operations.GetUploadLease(r.Context(), claims.LeaseID)
	if errors.Is(err, errUploadLeaseNotFound) || errors.Is(err, ErrBlobNotFound) {
		writeAPIError(w, r, objectUploadNotFoundOrRevoked())
		return
	}
	if err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	now := s.now().UTC()
	capabilityHash := objectUploadTokenDigest(token)
	if !uploadLeaseBindingsMatch(lease, claims, capabilityHash, principal.User.ID, principal.Session.ID) {
		writeAPIError(w, r, objectUploadNotFoundOrRevoked())
		return
	}
	if !lease.ExpiresAt.After(now) || !time.Unix(0, claims.ExpiresAtUnixNano).After(now) {
		writeAPIError(w, r, objectUploadExpired("target_expired"))
		return
	}
	if lease.LeaseState != "issued" {
		writeAPIError(w, r, objectUploadNotFoundOrRevoked())
		return
	}
	blob := lease.Blob
	if blob.IncidentID != claims.IncidentID || blob.StorageKey != claims.StorageKey || blob.ByteSize != claims.ByteSize {
		writeAPIError(w, r, objectUploadRejected(http.StatusConflict, "upload_contract_mismatch", nil))
		return
	}
	acceptedContract := map[string]any{
		"incident_id": blob.IncidentID.String(), "byte_size": blob.ByteSize,
		"filename_hint": nullableString(blob.FilenameHint), "content_type_hint": nullableString(blob.ContentTypeHint),
		"sha256_hex": nullableString(blob.ExpectedSHA256Hex),
	}
	_, acceptedDigestHex, digestErr := objectUploadBindingDigest(acceptedContract)
	if digestErr != nil || acceptedDigestHex != claims.AcceptedContractSHA256 {
		writeAPIError(w, r, objectUploadRejected(http.StatusConflict, "upload_contract_mismatch", nil))
		return
	}
	_, headerDigestHex, digestErr := objectUploadBindingDigest(lease.RequiredHeaders)
	if digestErr != nil || headerDigestHex != claims.RequiredHeadersSHA256 {
		writeAPIError(w, r, objectUploadRejected(http.StatusConflict, "upload_contract_mismatch", nil))
		return
	}
	if err := evidencepolicy.ValidatePersistedObjectBlobStorageKey(blob.StorageKey, blob.IncidentID, blob.ObjectBlobID); err != nil {
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
	for name, required := range lease.RequiredHeaders {
		if strings.TrimSpace(r.Header.Get(name)) != required {
			writeAPIError(w, r, objectUploadRejected(http.StatusBadRequest, "upload_contract_mismatch", nil))
			return
		}
	}
	if incident.IncidentStatus == admission.IncidentStatusClosed {
		writeAPIError(w, r, incidentClosedError())
		return
	}
	if err := s.operations.ClaimUploadLease(r.Context(), claims.LeaseID, capabilityHash, now); err != nil {
		switch {
		case admission.IsDenied(err, admission.DenialIncidentClosed):
			writeAPIError(w, r, incidentClosedError())
		case errors.Is(err, errUploadLeaseNotFound):
			writeAPIError(w, r, objectUploadNotFoundOrRevoked())
		case errors.Is(err, errUploadLeaseUnavailable):
			writeAPIError(w, r, objectUploadNotFoundOrRevoked())
		default:
			writeAPIError(w, r, internalAPIError(err))
		}
		return
	}
	contentType := lease.RequiredHeaders["Content-Type"]
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
	if err := s.operations.CompleteUploadLease(r.Context(), claims.LeaseID, s.now().UTC()); err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	if err := s.admission.slide(r.Context(), &principal, r.Method, r.URL.Path); err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func uploadLeaseBindingsMatch(
	lease uploadLeaseRecord,
	claims objectUploadTokenClaims,
	capabilityHash []byte,
	currentUserID uuid.UUID,
	currentSessionID uuid.UUID,
) bool {
	if claims.IssuingUserID != currentUserID || claims.IssuingSessionID != currentSessionID ||
		lease.LeaseID != claims.LeaseID || lease.ObjectBlobID != claims.ObjectBlobID ||
		lease.IncidentID != claims.IncidentID || lease.IssuingUserID != claims.IssuingUserID ||
		lease.IssuingSessionID != claims.IssuingSessionID || lease.RequiredMethod != claims.RequiredMethod ||
		lease.IssuedAt.UnixNano() != claims.IssuedAtUnixNano || lease.ExpiresAt.UnixNano() != claims.ExpiresAtUnixNano ||
		!bytes.Equal(lease.CapabilityHash, capabilityHash) ||
		!digestHexEquals(lease.AcceptedContractSHA256, claims.AcceptedContractSHA256) {
		return false
	}
	expectedHash := ""
	if lease.Blob.ExpectedSHA256Hex != nil {
		expectedHash = *lease.Blob.ExpectedSHA256Hex
	}
	return expectedHash == claims.ExpectedSHA256Hex
}

func digestHexEquals(digest []byte, encoded string) bool {
	decoded, err := hex.DecodeString(encoded)
	return err == nil && bytes.Equal(digest, decoded)
}

func (s *service) handleAttachBlob(w http.ResponseWriter, r *http.Request) {
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
	incident, apiErr := s.admission.visibleIncident(r.Context(), access.IncidentID, principal.User.ID)
	if apiErr != nil {
		writeAPIError(w, r, evidenceRecordNotFound())
		return
	}
	if _, apiErr := s.admission.requireRole(r.Context(), access.IncidentID, principal.User.ID, admission.RolesEditorReviewerAdmin, "editor|reviewer|admin"); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	request, apiErr := decodeAttachBlobRequest(r.Body)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	if incident.IncidentStatus == admission.IncidentStatusClosed {
		writeAPIError(w, r, incidentClosedError())
		return
	}
	now := s.now().UTC()
	requestHash := attachBlobRequestHash(request)
	preflight, err := s.operations.PreflightAttachBlob(r.Context(), principal.User, recordID, request, requestHash, now)
	if apiErr := translateAttachError(err, request.ClientTxnID); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	if preflight.Replay != nil {
		if err := s.admission.slide(r.Context(), &principal, r.Method, r.URL.Path); err != nil {
			writeAPIError(w, r, internalAPIError(err))
			return
		}
		_ = httpapi.WriteSuccess(w, r, preflight.Replay.StatusCode, preflight.Replay.Payload)
		return
	}
	blob := preflight.Blob
	var observed *observedObject
	if blob.UploadState == "pending" {
		if err := evidencepolicy.ValidatePersistedObjectBlobStorageKey(blob.StorageKey, blob.IncidentID, blob.ObjectBlobID); err != nil {
			writeAPIError(w, r, objectStoreDependencyAPIError(err))
			return
		}
		observed, err = s.objects.observeUploadedObject(r.Context(), blob)
		if err != nil {
			if apiErr := objectStoreDependencyAPIError(err); apiErr != nil {
				writeAPIError(w, r, apiErr)
				return
			}
			writeAPIError(w, r, evidenceAttachRejected(AttachReasonBlobFailed))
			return
		}
	}
	result, err := s.operations.AttachBlob(r.Context(), principal.User, recordID, request, requestHash, observed, httpapi.RequestIDFromContext(r.Context()), now)
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
