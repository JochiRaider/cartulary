package policy

import "time"

const (
	EvidenceRequested      = "requested"
	EvidencePendingReceipt = "pending_receipt"
	EvidenceReceived       = "received"
	EvidenceAvailable      = "available"
	EvidenceQuarantined    = "quarantined"
	EvidenceReleased       = "released"

	BlobPending     = "pending"
	BlobAvailable   = "available"
	BlobFailed      = "failed"
	BlobQuarantined = "quarantined"

	QuarantineTriggerContentInspection = "content_inspection_quarantine"
	QuarantineTriggerAdmin             = "admin_quarantine"
	QuarantineClearContentInspection   = "content_inspection_clear"
	QuarantineClearAdmin               = "admin_clear"
)

// ValidEvidenceLifecycle reports whether value belongs to the adopted closed
// Evidence lifecycle vocabulary.
func ValidEvidenceLifecycle(value string) bool {
	switch value {
	case EvidenceRequested, EvidencePendingReceipt, EvidenceReceived,
		EvidenceAvailable, EvidenceQuarantined, EvidenceReleased:
		return true
	default:
		return false
	}
}

// ValidBlobUploadState reports whether value belongs to the adopted closed
// Object Blob upload-state vocabulary.
func ValidBlobUploadState(value string) bool {
	switch value {
	case BlobPending, BlobAvailable, BlobFailed, BlobQuarantined:
		return true
	default:
		return false
	}
}

func ValidQuarantineEntryTrigger(value string) bool {
	return value == QuarantineTriggerContentInspection || value == QuarantineTriggerAdmin
}

func ValidQuarantineClearTrigger(value string) bool {
	return value == QuarantineClearContentInspection || value == QuarantineClearAdmin
}

func LegalBlobTransition(from string, to string, trigger string) bool {
	if from == to {
		return ValidBlobUploadState(from)
	}
	switch {
	case from == BlobPending && (to == BlobAvailable || to == BlobFailed):
		return true
	case from == BlobAvailable && to == BlobQuarantined:
		return ValidQuarantineEntryTrigger(trigger)
	case from == BlobQuarantined && to == BlobAvailable:
		return ValidQuarantineClearTrigger(trigger)
	default:
		return false
	}
}

func LegalEvidenceTransition(from string, to string) bool {
	if from == to {
		return ValidEvidenceLifecycle(from)
	}
	switch from {
	case EvidenceRequested:
		return to == EvidencePendingReceipt || to == EvidenceReceived || to == EvidenceAvailable
	case EvidencePendingReceipt:
		return to == EvidenceRequested || to == EvidenceReceived || to == EvidenceAvailable
	case EvidenceReceived:
		return to == EvidencePendingReceipt || to == EvidenceAvailable || to == EvidenceQuarantined
	case EvidenceAvailable:
		return to == EvidenceReceived || to == EvidenceQuarantined || to == EvidenceReleased
	case EvidenceQuarantined:
		return to == EvidenceReceived || to == EvidenceAvailable
	case EvidenceReleased:
		return to == EvidenceAvailable || to == EvidenceQuarantined
	default:
		return false
	}
}

// ViolatesEvidenceBlobBridge evaluates the resulting durable Evidence/blob
// combination. A linked available blob does not itself promote requested,
// pending-receipt, or received Evidence.
func ViolatesEvidenceBlobBridge(evidenceState string, hasBlob bool, blobState string) bool {
	switch evidenceState {
	case EvidenceAvailable, EvidenceReleased:
		return !hasBlob || blobState != BlobAvailable
	case EvidenceQuarantined:
		return hasBlob && blobState != BlobQuarantined
	default:
		return false
	}
}

type InitialLifecycleDisposition uint8

const (
	InitialLifecycleAllowed InitialLifecycleDisposition = iota
	InitialLifecycleInvalid
	InitialLifecycleGuardViolation
	InitialLifecycleIllegalTransition
)

func InitialEvidenceLifecycleDisposition(state string, blobSupplied bool, blobFinalized bool) InitialLifecycleDisposition {
	if !ValidEvidenceLifecycle(state) {
		return InitialLifecycleInvalid
	}
	switch state {
	case EvidenceAvailable:
		if !blobFinalized && !blobSupplied {
			return InitialLifecycleGuardViolation
		}
	case EvidenceQuarantined:
		if blobFinalized || blobSupplied {
			return InitialLifecycleGuardViolation
		}
	case EvidenceReleased:
		return InitialLifecycleIllegalTransition
	}
	return InitialLifecycleAllowed
}

type AssociationBlobDisposition uint8

const (
	AssociationBlobAvailable AssociationBlobDisposition = iota
	AssociationBlobNeedsFinalization
	AssociationBlobExpired
	AssociationBlobFailed
	AssociationBlobQuarantined
	AssociationBlobInconsistent
)

func ClassifyBlobForAssociation(uploadState string, pendingExpiresAt time.Time, now time.Time) AssociationBlobDisposition {
	switch uploadState {
	case BlobAvailable:
		return AssociationBlobAvailable
	case BlobPending:
		if !pendingExpiresAt.After(now) {
			return AssociationBlobExpired
		}
		return AssociationBlobNeedsFinalization
	case BlobFailed:
		return AssociationBlobFailed
	case BlobQuarantined:
		return AssociationBlobQuarantined
	default:
		return AssociationBlobInconsistent
	}
}

// IncidentMutationBlocked is the Evidence-local interpretation used after a
// visible incident snapshot has been loaded. Transactional entry points still
// recheck current incident mutability through the Incidents owner.
func IncidentMutationBlocked(status string) bool {
	return status == "closed"
}
