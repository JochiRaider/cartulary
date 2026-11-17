package timeline

import "fmt"

const (
	captureStateRough      = "rough"
	captureStateEnriched   = "enriched"
	captureStateReviewed   = "reviewed"
	captureStateSuperseded = "superseded"
)

// Lifecycle guard identifiers registered in Core 01 §3.3.6.2.
const (
	supersedeGuardReplacementDifferent                 = "replacement_must_differ_from_target"
	supersedeGuardReplacementVisibleActiveSameIncident = "replacement_must_be_active_in_incident"
	supersedeGuardReplacementNotSuperseded             = "replacement_must_not_be_superseded"
	supersedeGuardTargetMustNotHaveActiveReplacement   = "target_must_not_have_active_replacement"
)

// Lifecycle-transition reason family (Core 01 §3.3.6.2, REQ-01-231).
const (
	reasonTransitionNotAllowed = "transition_not_allowed"
	reasonGuardsNotSatisfied   = "guards_not_satisfied"
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

// captureStateAfterMaterialPatch returns the state a capture-state-material
// patch produces. Callers reject superseded rows first; any other state outside
// the closed vocabulary is a storage invariant failure (REQ-03-109a): an
// unclassified error that reaches the shared internal_error boundary, never
// illegal_transition.
func captureStateAfterMaterialPatch(current string) (string, error) {
	switch current {
	case captureStateRough, captureStateReviewed:
		return captureStateEnriched, nil
	case captureStateEnriched:
		return captureStateEnriched, nil
	default:
		return "", fmt.Errorf("timeline: invalid persisted capture state %q", current)
	}
}

// supersededPatchTargetState is the to_status reported when an ordinary patch
// is rejected on a superseded row (REQ-01-231): a tag-only request would leave
// the state unchanged, while any other change is capture-state material.
func supersededPatchTargetState(fieldKeys []string) string {
	for _, fieldKey := range fieldKeys {
		if fieldKey != "timeline.tags" {
			return captureStateEnriched
		}
	}
	return captureStateSuperseded
}

func captureStateAllowsMarkReviewed(current string) bool {
	return current == captureStateRough || current == captureStateEnriched
}

func captureStateAllowsSupersede(current string) bool {
	return current == captureStateRough || current == captureStateEnriched || current == captureStateReviewed
}
