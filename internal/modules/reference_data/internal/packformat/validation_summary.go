package packformat

import (
	"bytes"
	"encoding/json"
	"errors"

	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
)

var validationSummaryShape = compileProjection("validation_summary.v1.schema.json")

// DecodeValidationSummary verifies canonical representation, closed shape,
// registered check mapping, issue identities, order and exact retained counts.
func DecodeValidationSummary(data []byte) (*ValidationSummary, error) {
	invalid := errors.New("reference pack: invalid validation summary")
	if !admittedValidationSummarySize(len(data)) {
		return nil, invalid
	}
	value, err := canonicaljson.DecodeStrict(data)
	if err != nil || !validationSummaryShape.matches(value) {
		return nil, invalid
	}
	canonical, err := canonicaljson.Marshal(value)
	if err != nil || !bytes.Equal(data, canonical) {
		return nil, invalid
	}
	var s ValidationSummary
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, invalid
	}
	if s.Retained != len(s.Issues) || s.Retained != min(s.Total, diagnosticRetained) || s.Truncated != (s.Total > diagnosticRetained) {
		return nil, invalid
	}
	if s.Result == "succeeded" {
		if s.Total != 0 || s.PrimaryIssueID != nil {
			return nil, invalid
		}
		return &s, nil
	}
	if s.Total == 0 || s.PrimaryIssueID == nil || *s.PrimaryIssueID != s.Issues[0].ID {
		return nil, invalid
	}
	var check Check
	for _, candidate := range checks {
		if candidate.ID == s.Issues[0].CheckID {
			check = candidate
			break
		}
	}
	if check.ID == "" {
		return nil, invalid
	}
	for i, issue := range s.Issues {
		expected, err := makeIssue(check, Finding{Path: issue.Path, EntryID: issue.EntryID, Details: issue.Details})
		if err != nil {
			return nil, invalid
		}
		a, _ := canonicaljson.Marshal(expected)
		b, _ := canonicaljson.Marshal(issue)
		if !bytes.Equal(a, b) || i > 0 && compareIssues(s.Issues[i-1], issue) >= 0 {
			return nil, invalid
		}
	}
	return &s, nil
}

func admittedValidationSummarySize(size int) bool { return size > 0 && size <= 16777216 }
