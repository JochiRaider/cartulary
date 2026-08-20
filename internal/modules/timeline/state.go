package timeline

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

func initialCaptureState() string {
	return captureStateRough
}

func CreateRequestHasUserValue(request CreateRequest) bool {
	return request.DateEnteredText != nil ||
		request.AnalystText != nil ||
		request.MitreStageText != nil ||
		request.DeviceObjectText != nil ||
		request.IPAddressText != nil ||
		request.ActivityUTCText != nil ||
		request.ActivityLocalText != nil ||
		request.RawActivityText != nil ||
		request.ActivitySynopsisText != nil ||
		request.DataSourceText != nil ||
		(request.HostRefs != nil && len(request.HostRefs.Actions) > 0) ||
		(request.IdentityRefs != nil && len(request.IdentityRefs.Actions) > 0) ||
		(request.Tags != nil && len(request.Tags.Actions) > 0) ||
		(request.AttachedEvidence != nil && len(request.AttachedEvidence.Actions) > 0)
}

func captureStateAfterMaterialPatch(current string) (string, error) {
	switch current {
	case captureStateRough, captureStateReviewed:
		return captureStateEnriched, nil
	case captureStateEnriched:
		return captureStateEnriched, nil
	default:
		return "", ErrIllegalTransition
	}
}

func captureStateAllowsMarkReviewed(current string) bool {
	return current == captureStateRough || current == captureStateEnriched
}

func captureStateAllowsSupersede(current string) bool {
	return current == captureStateRough || current == captureStateEnriched || current == captureStateReviewed
}
