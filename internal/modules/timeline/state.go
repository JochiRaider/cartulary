package timeline

import "github.com/google/uuid"

const (
	captureStateRough      = "rough"
	captureStateEnriched   = "enriched"
	captureStateReviewed   = "reviewed"
	captureStateSuperseded = "superseded"
)

const (
	supersedeGuardReplacementDifferent                 = "replacement_must_be_different_timeline_record"
	supersedeGuardReplacementVisibleActiveSameIncident = "replacement_must_be_visible_active_same_incident_timeline_record"
	supersedeGuardReplacementNotSuperseded             = "replacement_must_not_be_superseded"
	supersedeGuardTargetMustNotHaveActiveReplacement   = "target_must_not_have_active_replacement"
)

type IllegalTransitionError struct {
	ReasonCode     string
	FromStatus     string
	ToStatus       string
	ViolatedGuards []string
}

func (e *IllegalTransitionError) Error() string {
	return ErrIllegalTransition.Error()
}

func (e *IllegalTransitionError) Unwrap() error {
	return ErrIllegalTransition
}

func newIllegalTransitionError(reasonCode string, fromStatus string, toStatus string, guards ...string) *IllegalTransitionError {
	return &IllegalTransitionError{
		ReasonCode:     reasonCode,
		FromStatus:     fromStatus,
		ToStatus:       toStatus,
		ViolatedGuards: append([]string{}, guards...),
	}
}

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
	guards := make([]string, 0, 2)
	if *replacementRecordID == currentRecordID {
		guards = append(guards, supersedeGuardReplacementDifferent)
	}
	if replacementIncidentID != nil && *replacementIncidentID != currentIncidentID {
		guards = append(guards, supersedeGuardReplacementVisibleActiveSameIncident)
	}
	if len(guards) > 0 {
		return newIllegalTransitionError("supersede_not_allowed", "", captureStateSuperseded, guards...)
	}
	return nil
}
