package reportcomposition

import (
	"bytes"
	"cmp"
	"slices"
	"strings"
)

// Provenance is validation-local. It orders diagnostics without adding public
// schema members, trusting supplied ordinals, or exposing invalid source values.
type compositionFinding struct {
	code        string
	details     map[string]any
	sourceArray int
	sourceIndex int
}

func issue(code string, details map[string]any) compositionFinding {
	safe := make(map[string]any, len(details))
	for key, value := range details {
		switch key {
		case "composition_id", "composition_version", "composition_op_id", "diagram_id", "authored_text_id", "template_id", "template_version", "graph_view_id", "target_ref":
			if validCompositionIdentifier(value) {
				safe[key] = value
			}
		case "field", "limit", "composition_anchor_kind", "source_kind", "release_scope", "layout_mode", "output_kind":
			// These values are emitted by the validator's closed branches.
			safe[key] = value
		}
	}
	return compositionFinding{code: code, details: safe, sourceArray: 5, sourceIndex: -1}
}

func compositionIssueAt(array, index int) func(string, map[string]any) compositionFinding {
	return func(code string, details map[string]any) compositionFinding {
		finding := issue(code, details)
		finding.sourceArray, finding.sourceIndex = array, index
		return finding
	}
}

func attributeCompositionFindings(findings []compositionFinding, array, index int) []compositionFinding {
	for i := range findings {
		if findings[i].sourceArray == 5 {
			findings[i].sourceArray, findings[i].sourceIndex = array, index
		}
	}
	return findings
}

func (finding compositionFinding) public() map[string]any {
	return map[string]any{
		"severity": "error", "code": finding.code,
		"message_key":  "report_composition." + finding.code,
		"safe_details": finding.details,
	}
}

var compositionStages = []string{
	"schema_validation", "resource_binding_validation", "canonical_digest_validation",
	"template_vocab_validation", "anchor_validation", "authored_text_validation",
	"diagram_validation", "operation_validation",
}

func compositionFindingStage(code string) int {
	switch {
	case code == "composition_digest_mismatch":
		return 2
	case code == "composition_incident_mismatch", code == "composition_id_mismatch", code == "composition_template_mismatch":
		return 1
	case strings.HasPrefix(code, "composition_anchor_"), code == "composition_drop_invalid_for_external_release":
		return 4
	case strings.HasPrefix(code, "authored_"):
		return 5
	case strings.HasPrefix(code, "diagram_"), code == "raw_generated_source_invalid", code == "composition_replacement_target_missing", code == "manual_layout_not_supported_for_output_kind":
		return 6
	case code == "composition_duplicate_diagram_insert":
		return 7
	default:
		return 0
	}
}

func orderedCompositionFindings(findings []compositionFinding) ([]map[string]any, any) {
	slices.SortFunc(findings, func(a, b compositionFinding) int {
		for _, order := range []int{
			cmp.Compare(compositionFindingStage(a.code), compositionFindingStage(b.code)),
			cmp.Compare(a.sourceArray, b.sourceArray),
			compareCompositionSourceIndex(a.sourceIndex, b.sourceIndex),
		} {
			if order != 0 {
				return order
			}
		}
		for _, key := range []string{"composition_op_id", "diagram_id", "authored_text_id", "target_ref"} {
			left, leftPresent := a.details[key].(string)
			right, rightPresent := b.details[key].(string)
			if leftPresent != rightPresent {
				if leftPresent {
					return -1
				}
				return 1
			}
			if order := strings.Compare(left, right); order != 0 {
				return order
			}
		}
		if order := strings.Compare(a.code, b.code); order != 0 {
			return order
		}
		left, _ := canonicalJSON(a.details)
		right, _ := canonicalJSON(b.details)
		return bytes.Compare(left, right)
	})
	public := make([]map[string]any, 0, len(findings))
	var stage any
	for _, finding := range findings {
		public = append(public, finding.public())
	}
	if len(findings) > 0 {
		stage = compositionStages[compositionFindingStage(findings[0].code)]
	}
	return public, stage
}

func compareCompositionSourceIndex(a, b int) int {
	if a < 0 && b >= 0 {
		return 1
	}
	if b < 0 && a >= 0 {
		return -1
	}
	return cmp.Compare(a, b)
}
