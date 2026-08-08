package jobs

type transitionReason string

const (
	transitionReasonState    transitionReason = "state_not_allowed"
	transitionReasonProgress transitionReason = "progress_not_allowed"
)

type transitionMutation struct {
	resource Resource
	changed  bool
}

var terminalSources = map[string]map[string]struct{}{
	StatusSucceeded: {
		StatusRunning:         {},
		StatusCancelRequested: {},
	},
	StatusFailed: {
		StatusRunning:         {},
		StatusCancelRequested: {},
	},
	StatusCanceled: {
		StatusCancelRequested: {},
	},
}

func validateProgress(progress Progress) error {
	if progress.Completed < 0 ||
		(progress.Total != nil && (*progress.Total <= 0 || progress.Completed > *progress.Total)) {
		return invalidTransition(transitionReasonProgress)
	}
	return nil
}

func validateInitialProgress(progress Progress) error {
	if err := validateProgress(progress); err != nil {
		return ErrInvalidJobDefinition
	}
	return nil
}

func validateProgressAdvance(current Progress, next Progress) error {
	if next.Completed < current.Completed {
		return invalidTransition(transitionReasonProgress)
	}
	if current.Total != nil {
		if next.Total == nil || *next.Total < *current.Total {
			return invalidTransition(transitionReasonProgress)
		}
	}
	return nil
}

func progressEqual(left Progress, right Progress) bool {
	if left.Completed != right.Completed {
		return false
	}
	if left.Total == nil || right.Total == nil {
		return left.Total == nil && right.Total == nil
	}
	return *left.Total == *right.Total
}

func optionalStringEqual(left *string, right *string) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}
	return *left == *right
}

func invalidTransition(_ transitionReason) error {
	// Reasons are deliberately closed and private. Public callers receive only
	// the stable sentinel and cannot observe storage state or identifiers.
	return ErrInvalidTransition
}
