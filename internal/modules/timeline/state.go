package timeline

import "github.com/google/uuid"

const (
	captureStateRough      = "rough"
	captureStateEnriched   = "enriched"
	captureStateReviewed   = "reviewed"
	captureStateSuperseded = "superseded"
)

var captureStateVocabulary = map[string]struct{}{
	captureStateRough:      {},
	captureStateEnriched:   {},
	captureStateReviewed:   {},
	captureStateSuperseded: {},
}

func InitialCaptureState() string {
	return captureStateRough
}

func IsSupportedCaptureState(value string) bool {
	_, ok := captureStateVocabulary[value]
	return ok
}

func CreateRequestHasUserValue(request CreateRequest) bool {
	return request.OccurredAt != nil ||
		request.Summary != nil ||
		request.Details != nil ||
		request.SourceText != nil ||
		(request.HostRefs != nil && len(request.HostRefs.Actions) > 0) ||
		(request.IdentityRefs != nil && len(request.IdentityRefs.Actions) > 0) ||
		(request.Tags != nil && len(request.Tags.Actions) > 0)
}

func CaptureStateAfterMaterialPatch(current string) (string, error) {
	switch current {
	case captureStateRough, captureStateReviewed:
		return captureStateEnriched, nil
	case captureStateEnriched:
		return captureStateEnriched, nil
	default:
		return "", ErrIllegalTransition
	}
}

func CaptureStateAllowsMarkReviewed(current string) bool {
	return current == captureStateRough || current == captureStateEnriched
}

func CaptureStateAllowsSupersede(current string) bool {
	return current == captureStateRough || current == captureStateEnriched || current == captureStateReviewed
}

func ValidateSupersedeReplacement(currentRecordID uuid.UUID, currentIncidentID uuid.UUID, replacementRecordID *uuid.UUID, replacementIncidentID *uuid.UUID) error {
	if replacementRecordID == nil {
		return nil
	}
	if *replacementRecordID == currentRecordID {
		return ErrIllegalTransition
	}
	if replacementIncidentID != nil && *replacementIncidentID != currentIncidentID {
		return ErrIllegalTransition
	}
	return nil
}
