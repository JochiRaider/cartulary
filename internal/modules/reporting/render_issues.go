package reporting

import (
	"bytes"
	"errors"
	"slices"
	"strings"
)

const retainedRenderIssueLimit = 500
const evaluatedRenderIssueLimit = 2000

// One collector belongs to one validation stage. It retains only the smallest
// ordered prefix while evaluating the entire stage, so source traversal order
// cannot choose the terminal failure or grow retained diagnostic memory.
type renderIssueCollector struct {
	stage string
	set   renderIssueSet
}

type renderIssueSet struct {
	count    int
	retained []reportingRenderValidationIssue
}

func (c *renderIssueCollector) add(err error) {
	if err == nil {
		return
	}
	failure := renderFailureDetails(renderFailureAt(err, c.stage, "export_model_invalid", "invalid_type"))
	if failure.issues != nil && failure.Stage == c.stage {
		c.set.count += failure.issues.count
		for _, issue := range failure.issues.retained {
			if issue.Severity == "error" {
				c.retain(issue)
			}
		}
		return
	}
	issue := reportingRenderValidationIssue{Stage: c.stage, Severity: "error", FailureCode: &failure.FailureCode, ReasonCode: &failure.ReasonCode, SafeDetails: failure.SafeDetails}
	c.set.count++
	c.retain(issue)
}

func (c *renderIssueCollector) retain(issue reportingRenderValidationIssue) {
	index, _ := slices.BinarySearchFunc(c.set.retained, issue, compareRenderIssues)
	if index >= retainedRenderIssueLimit {
		return
	}
	c.set.retained = slices.Insert(c.set.retained, index, issue)
	if len(c.set.retained) > retainedRenderIssueLimit {
		c.set.retained = c.set.retained[:retainedRenderIssueLimit]
	}
}

func (c *renderIssueCollector) finish() error {
	if c.set.count == 0 {
		return nil
	}
	if c.set.count > evaluatedRenderIssueLimit {
		// The bounded warning is retained even when 500 earlier errors exist.
		c.set.retained = append(c.set.retained[:retainedRenderIssueLimit-1], renderIssueLimitWarning(c.stage, c.set.count))
	}
	first := c.set.retained[0]
	return &renderValidationError{Stage: first.Stage, FailureCode: *first.FailureCode, ReasonCode: *first.ReasonCode, SafeDetails: first.SafeDetails, issues: &c.set}
}

func renderIssueLimitWarning(stage string, count int) reportingRenderValidationIssue {
	code, reason := "validation_summary_truncated", "validation_issue_limit_reached"
	return reportingRenderValidationIssue{Stage: stage, Severity: "warning", FailureCode: &code, ReasonCode: &reason,
		SafeDetails: map[string]any{"limit_key": "validation_issues.count", "limit": evaluatedRenderIssueLimit, "actual": count}}
}

func (s *renderIssueSet) validate(first RenderFailure) error {
	invalid := errors.New("reporting: invalid retained render issue set")
	if s.count < 1 || len(s.retained) != min(s.count, retainedRenderIssueLimit) {
		return invalid
	}
	for index, issue := range s.retained {
		// Current producers expose only closed count and timer diagnostics. Do
		// not admit arbitrary paths or richer maps merely because a future
		// producer might support them.
		if issue.Stage != first.Stage || issue.ExportModelPath != nil || issue.BundlePath != nil || issue.FailureCode == nil || issue.ReasonCode == nil {
			return invalid
		}
		if index == retainedRenderIssueLimit-1 && s.count > evaluatedRenderIssueLimit {
			expected, _ := canonicalJSON(renderIssueLimitWarning(first.Stage, s.count))
			actual, err := canonicalJSON(issue)
			if err != nil || !bytes.Equal(expected, actual) {
				return invalid
			}
			continue
		}
		if issue.Severity != "error" || validateRenderFailure(RenderFailure{Stage: issue.Stage, FailureCode: *issue.FailureCode, ReasonCode: *issue.ReasonCode, SafeDetails: issue.SafeDetails}) != nil {
			return invalid
		}
		if index > 0 && compareRenderIssues(s.retained[index-1], issue) > 0 {
			return invalid
		}
	}
	issue := s.retained[0]
	if *issue.FailureCode != first.FailureCode || *issue.ReasonCode != first.ReasonCode || !sameReportingValue(issue.SafeDetails, first.SafeDetails) {
		return invalid
	}
	return nil
}

func compareRenderIssues(a, b reportingRenderValidationIssue) int {
	stages := strings.Fields(renderStages)
	if n := slices.Index(stages, a.Stage) - slices.Index(stages, b.Stage); n != 0 {
		return n
	}
	if n := strings.Compare(a.Severity, b.Severity); n != 0 {
		return n
	}
	// Current issues are safe-detail-only or unattributed. Paths remain null.
	if (len(a.SafeDetails) == 0) != (len(b.SafeDetails) == 0) {
		if len(a.SafeDetails) != 0 {
			return -1
		}
		return 1
	}
	if n := strings.Compare(*a.FailureCode, *b.FailureCode); n != 0 {
		return n
	}
	if n := strings.Compare(*a.ReasonCode, *b.ReasonCode); n != 0 {
		return n
	}
	left, _ := canonicalJSON(a.SafeDetails)
	right, _ := canonicalJSON(b.SafeDetails)
	return bytes.Compare(left, right)
}
