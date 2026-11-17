package evidence

import (
	"errors"
	"net/http"

	"github.com/JochiRaider/cartulary/internal/modules/incidents/admission"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
)

// translateAttachError is the single public translation boundary for the
// attach application operation. Its order is intentional and concealment-safe.
func translateAttachError(err error, clientTxnID string) *httpapi.APIError {
	if err == nil {
		return nil
	}
	var rowConflict *rowVersionConflictError
	var attachRejected AttachRejectedError
	var blobTransition *blobTransitionError
	switch {
	case errors.Is(err, ErrClientTxnConflict):
		return clientTxnConflict(clientTxnID)
	case admission.IsDenied(err, admission.DenialIncidentClosed):
		return incidentClosedError()
	case errors.As(err, &rowConflict):
		return rowVersionConflict(rowConflict.RecordID, rowConflict.BaseRowVersion, rowConflict.CurrentRowVersion)
	case errors.Is(err, ErrEvidenceNotFound):
		return evidenceRecordNotFound()
	case errors.As(err, &attachRejected):
		return evidenceAttachRejected(attachRejected.ReasonCode)
	case errors.Is(err, ErrBlobNotFound), errors.Is(err, ErrIncidentMismatch):
		return evidenceAttachRejected(AttachReasonBlobNotVisible)
	case errors.Is(err, errEvidenceQuarantined):
		return evidenceAttachRejected(attachReasonEvidenceQuarantined)
	case errors.Is(err, errBlobNotAttachable):
		return evidenceAttachRejected(attachReasonEvidenceInconsistent)
	case errors.As(err, &blobTransition):
		// Core 01 §3.3.6.2: the Blob machine is matrix-only.
		return &httpapi.APIError{Status: http.StatusConflict, Code: "illegal_transition", Message: "illegal transition", Details: map[string]any{
			"from_status": blobTransition.fromStatus, "to_status": blobTransition.toStatus,
			"violated_guards": []string{}, "reason_code": "transition_not_allowed",
		}}
	default:
		return internalAPIError(err)
	}
}
