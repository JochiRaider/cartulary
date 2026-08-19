package evidence

import (
	"errors"

	"github.com/JochiRaider/cartulary/internal/modules/incidents/admission"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
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
	switch {
	case errors.Is(err, authn.ErrClientTxnConflict):
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
	case errors.Is(err, ErrEvidenceQuarantined):
		return evidenceAttachRejected(AttachReasonEvidenceQuarantined)
	case errors.Is(err, ErrBlobNotAttachable):
		return evidenceAttachRejected(AttachReasonEvidenceInconsistent)
	default:
		return internalAPIError(err)
	}
}
