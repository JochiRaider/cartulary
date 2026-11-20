package harnessruntime

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/JochiRaider/cartulary/internal/testutil/harnessschema"
)

// PreparationDiagnostic is the closed record a rejected graph preparation
// publishes over the private compiler descriptor instead of a plan
// (TH-HARNESS-REQ-854). It carries typed facts only, never exception text.
type PreparationDiagnostic struct {
	FailureReason       string   `json:"failure_reason"`
	Cause               string   `json:"cause"`
	Subjects            []string `json:"subjects"`
	OmittedSubjectCount int      `json:"omitted_subject_count"`
}

const (
	preparationDiagnosticSchema       = "cartulary.harness_preparation_diagnostic.v1"
	maximumPreparationDiagnosticBytes = 4096
)

// preparationRejection carries a validated diagnostic beside its failure
// classification, so the terminal output can render it without text parsing.
type preparationRejection struct{ diagnostic PreparationDiagnostic }

func (e *preparationRejection) Error() string { return "graph preparation rejected" }

// decodePreparationDiagnostic admits only a schema-valid record within the
// transport bound; anything else fails closed in the caller.
func decodePreparationDiagnostic(data []byte) (PreparationDiagnostic, bool) {
	var diagnostic PreparationDiagnostic
	if len(data) == 0 || len(data) > maximumPreparationDiagnosticBytes {
		return diagnostic, false
	}
	value, err := harnessschema.Decode(data)
	if err != nil || harnessschema.Validate(preparationDiagnosticSchema, value) != nil {
		return diagnostic, false
	}
	if json.Unmarshal(data, &diagnostic) != nil {
		return diagnostic, false
	}
	return diagnostic, true
}

func preparationFailure(diagnostic PreparationDiagnostic) error {
	return errors.Join(failed("config", diagnostic.FailureReason), &preparationRejection{diagnostic: diagnostic})
}

// PreparationDiagnosticLines renders the validated diagnostic carried by err,
// one header line and one line per subject (TH-HARNESS-REQ-856). It returns
// nil when err carries none.
func PreparationDiagnosticLines(err error) []string {
	var rejection *preparationRejection
	if !errors.As(err, &rejection) {
		return nil
	}
	d := rejection.diagnostic
	lines := []string{fmt.Sprintf("[GRAPH-PREPARE] failure_reason=%s cause=%s subjects=%d omitted=%d", d.FailureReason, d.Cause, len(d.Subjects), d.OmittedSubjectCount)}
	for _, subject := range d.Subjects {
		lines = append(lines, "[GRAPH-PREPARE] subject="+subject)
	}
	return lines
}

// runtimePlanFailure attributes a rejected preparation to runtime_plan. A
// validated diagnostic keeps its own reason, so usage_error stays distinct
// (TH-HARNESS-REQ-305); any other rejection is a configuration error.
func runtimePlanFailure(err error) error {
	if PreparationDiagnosticLines(err) != nil {
		return entryFailureAt("runtime_plan", err)
	}
	return entryFailureAt("runtime_plan", errors.Join(failed("config", "configuration_error"), err))
}
