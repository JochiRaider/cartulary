package evidence

import (
	"errors"

	"github.com/JochiRaider/cartulary/internal/modules/incidents/admission"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
)

// translateAttachError is the single public translation boundary for the
// attach application operation. Its order is intentional and concealment-safe.
// A rejected Blob transition is not among them: attach classifies the Blob it
// holds locked as pending before the guarded transition, so a transition that
// matches no row is corrupt state and stays internal_error (REQ-01-706).
func translateAttachError(err error, clientTxnID string) *httpapi.APIError {
	if err == nil {
		return nil
	}
	var rowConflict *rowVersionConflictError
	var attachRejected AttachRejectedError
	switch {
	case errors.Is(err, ErrClientTxnConflict):
		return clientTxnConflict(clientTxnID)
	case errors.Is(err, errAttachTargetDeleted):
		return recordDeletedUseRestore()
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
	default:
		return internalAPIError(err)
	}
}
