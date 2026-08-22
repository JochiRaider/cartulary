package revisions

import (
	"errors"

	"github.com/JochiRaider/cartulary/internal/modules/revisions/rollbackcontract"
)

var (
	ErrDuplicateTargetSemantics  = errors.New("revisions: duplicate target semantics")
	ErrMissingTargetSemantics    = errors.New("revisions: missing target semantics")
	ErrUnexpectedTargetSemantics = errors.New("revisions: unexpected target semantics")
	ErrInvalidTargetSemantics    = errors.New("revisions: invalid target semantics")
)

type compiledTargetSemantics struct {
	dispatchClass       rollbackcontract.DispatchClass
	admittedRecordTypes []string
	history             HistoryFacet
	historyValidator    HistoryValidator
	rowProviders        map[string]rollbackcontract.RowSourceProvider
	nonRowProvider      rollbackcontract.NonRowTargetProvider
}

type TargetSemanticsCatalog struct {
	byTargetKind                 map[string]compiledTargetSemantics
	defaultRowTargetByRecordType map[string]string
}

// NewTargetSemanticsCatalog compiles exact source-owner closure against the
// current generated registry. Callers cannot select or synthesize production
// requirements.
func NewTargetSemanticsCatalog(contributions []ProviderContribution) (*TargetSemanticsCatalog, error) {
	requirements, err := currentTargetSemanticsRequirements()
	if err != nil {
		return nil, err
	}
	return compileTargetSemanticsCatalog(requirements, contributions)
}
