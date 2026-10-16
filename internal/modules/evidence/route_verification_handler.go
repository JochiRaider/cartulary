package evidence

import (
	"errors"
	"net/http"

	"github.com/JochiRaider/cartulary/internal/modules/incidents/admission"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
)

func (s *service) handleVerifyBlob(w http.ResponseWriter, r *http.Request) {
	principal, apiErr := s.admission.authenticate(r, true)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	blobID, ok := pathUUID(w, r, "object_blob_id")
	if !ok {
		return
	}
	blob, err := s.operations.GetBlob(r.Context(), blobID)
	if errors.Is(err, ErrBlobNotFound) {
		writeAPIError(w, r, evidenceAttachRejected(AttachReasonBlobNotVisible))
		return
	}
	if err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	if _, apiErr = s.admission.visibleIncident(r.Context(), blob.IncidentID, principal.User.ID); apiErr != nil {
		writeAPIError(w, r, evidenceAttachRejected(AttachReasonBlobNotVisible))
		return
	}
	if _, apiErr = s.admission.requireRole(r.Context(), blob.IncidentID, principal.User.ID, admission.RolesEditorReviewerAdmin, "editor|reviewer|admin"); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	raw, apiErr := decodeStrictObject(r.Body, "invalid_mutation_payload")
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	for key := range raw {
		if key != "client_txn_id" {
			writeAPIError(w, r, invalidMutationPayload(key, "unknown_field"))
			return
		}
	}
	transactionID, apiErr := requiredString(raw, "client_txn_id", "invalid_mutation_payload")
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	job, err := s.verification.admit(r.Context(), principal.User.ID, blobID, transactionID)
	switch {
	case errors.Is(err, authn.ErrClientTxnConflict):
		apiErr = clientTxnConflict(transactionID)
	case admission.IsDenied(err, admission.DenialNotVisible):
		apiErr = evidenceAttachRejected(AttachReasonBlobNotVisible)
	case admission.IsDenied(err, admission.DenialInsufficientRole):
		apiErr = &httpapi.APIError{Status: http.StatusForbidden, Code: "authorization_denied", Details: map[string]any{"required_role": "editor|reviewer|admin"}}
	default:
		apiErr = translateAttachError(err, transactionID)
	}
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	if err = s.admission.slide(r.Context(), &principal, r.Method, r.URL.Path); err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	_ = httpapi.WriteSuccess(w, r, http.StatusAccepted, job)
}
