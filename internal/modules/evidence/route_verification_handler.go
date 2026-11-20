package evidence

import (
	"errors"
	"net/http"

	"github.com/JochiRaider/cartulary/internal/modules/incidents/admission"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
)

func (s *service) handleVerifyBlob(w http.ResponseWriter, r *http.Request) {
	principal, apiErr := admittedPrincipal(r)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	// A malformed object_blob_id received the declared 409 blob_not_visible
	// from central admission, identical to a missing Blob (REQ-01-707,
	// REQ-01-452a).
	blobID, apiErr := admittedPathUUID(r, "object_blob_id")
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
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
	// Only an incident the caller cannot see is concealed as an invisible
	// Blob; any other admission failure stays internal_error (REQ-01-706).
	// One read decides visibility and then the role.
	if _, apiErr = s.admission.concealedRole(r.Context(), blob.IncidentID, principal.User.ID, admission.RolesEditorReviewerAdmin, "editor|reviewer|admin", blobNotVisible); apiErr != nil {
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
	if apiErr = translateVerifyError(err); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	if err = s.admission.slide(r.Context(), &principal); err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	_ = httpapi.WriteSuccess(w, r, http.StatusAccepted, job)
}

// translateVerifyError maps verification admission failures to the route's
// declared codes (Core 01 REQ-01-452a). Admission checks the caller's role and
// the incident lifecycle under the transaction and then the Blob's
// eligibility; a Blob removed after the route read it is concealed like a
// missing one. Every other failure is internal_error (REQ-01-706), including a
// retained receipt whose request hash differs: the hash covers only the
// object_blob_id that the receipt's identity already binds, so a mismatch is
// corrupt stored state, never a divergent request.
func translateVerifyError(err error) *httpapi.APIError {
	if err == nil {
		return nil
	}
	var attachRejected AttachRejectedError
	switch {
	case admission.IsDenied(err, admission.DenialNotVisible):
		return evidenceAttachRejected(AttachReasonBlobNotVisible)
	case admission.IsDenied(err, admission.DenialInsufficientRole):
		return &httpapi.APIError{Status: http.StatusForbidden, Code: "authorization_denied", Details: map[string]any{"required_role": "editor|reviewer|admin"}}
	case admission.IsDenied(err, admission.DenialIncidentClosed):
		return incidentClosedError()
	case errors.As(err, &attachRejected):
		return evidenceAttachRejected(attachRejected.ReasonCode)
	case errors.Is(err, ErrBlobNotFound):
		return evidenceAttachRejected(AttachReasonBlobNotVisible)
	default:
		return internalAPIError(err)
	}
}

// blobNotVisible is the concealed response for a Blob the caller cannot see.
func blobNotVisible() *httpapi.APIError {
	return evidenceAttachRejected(AttachReasonBlobNotVisible)
}
