// Package diagramcontract owns the closed Reporting selection-rule wire contract
// shared with the Composition authoring boundary. It does not resolve live data.
package diagramcontract

import (
	"bytes"
	"encoding/json"
	"fmt"
	reportingjson "github.com/JochiRaider/cartulary/internal/modules/reporting/canonicaljson"
	"sort"
)

type SelectionRule struct {
	SchemaID          string   `json:"schema_id"`
	RuleID            string   `json:"rule_id"`
	VertexRefs        []string `json:"vertex_refs"`
	EdgeRefs          []string `json:"edge_refs"`
	SeedVertexRefs    []string `json:"seed_vertex_refs"`
	Depth             *int     `json:"depth"`
	EdgeKindFilter    []string `json:"edge_kind_filter"`
	TimelineRecordIDs []string `json:"timeline_record_ids"`
	OverflowPolicy    string   `json:"overflow_policy"`
}

func (rule *SelectionRule) UnmarshalJSON(data []byte) error {
	canonical, err := reportingjson.Canonicalize(data)
	if err != nil {
		return err
	}
	data = canonical
	var members map[string]json.RawMessage
	if err := json.Unmarshal(data, &members); err != nil {
		return err
	}
	allowed := map[string]struct{}{
		"schema_id": {}, "rule_id": {}, "vertex_refs": {}, "edge_refs": {},
		"seed_vertex_refs": {}, "depth": {}, "edge_kind_filter": {},
		"timeline_record_ids": {}, "overflow_policy": {},
	}
	if len(members) != len(allowed) {
		return fmt.Errorf("selection_rule must contain the closed member set")
	}
	for key := range members {
		if _, present := allowed[key]; !present {
			return fmt.Errorf("selection_rule contains unknown member %q", key)
		}
	}
	for _, key := range []string{"vertex_refs", "edge_refs", "seed_vertex_refs", "edge_kind_filter", "timeline_record_ids"} {
		if bytes.Equal(bytes.TrimSpace(members[key]), []byte("null")) {
			return fmt.Errorf("selection_rule member %q is not nullable", key)
		}
	}
	type wireRule SelectionRule
	var decoded wireRule
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}
	if decoded.SchemaID != "diagram_selection_rule.v1" ||
		(decoded.OverflowPolicy != "fail" && decoded.OverflowPolicy != "summarize") ||
		hasDuplicateStrings(decoded.EdgeKindFilter) || !sort.StringsAreSorted(decoded.EdgeKindFilter) {
		return fmt.Errorf("selection_rule has invalid common members")
	}
	if hasDuplicateStrings(decoded.TimelineRecordIDs) {
		return fmt.Errorf("selection_rule has duplicate timeline refs")
	}
	switch decoded.RuleID {
	case "explicit_refs":
		if decoded.Depth != nil || len(decoded.SeedVertexRefs) != 0 || len(decoded.EdgeKindFilter) != 0 || len(decoded.TimelineRecordIDs) != 0 || len(decoded.VertexRefs)+len(decoded.EdgeRefs) == 0 || hasDuplicateStrings(decoded.VertexRefs) || hasDuplicateStrings(decoded.EdgeRefs) {
			return fmt.Errorf("selection_rule explicit_refs members are invalid")
		}
	case "neighborhood":
		if decoded.Depth == nil || (*decoded.Depth != 1 && *decoded.Depth != 2) || len(decoded.SeedVertexRefs) == 0 || hasDuplicateStrings(decoded.SeedVertexRefs) || len(decoded.VertexRefs) != 0 || len(decoded.EdgeRefs) != 0 || len(decoded.TimelineRecordIDs) != 0 {
			return fmt.Errorf("selection_rule neighborhood members are invalid")
		}
	case "timeline_sequence":
		if decoded.Depth != nil || len(decoded.TimelineRecordIDs) == 0 || len(decoded.VertexRefs) != 0 || len(decoded.EdgeRefs) != 0 || len(decoded.SeedVertexRefs) != 0 || len(decoded.EdgeKindFilter) != 0 {
			return fmt.Errorf("selection_rule timeline_sequence members are invalid")
		}
	case "all_with_bounds":
		if decoded.Depth != nil || len(decoded.VertexRefs) != 0 || len(decoded.EdgeRefs) != 0 || len(decoded.SeedVertexRefs) != 0 || len(decoded.EdgeKindFilter) != 0 || len(decoded.TimelineRecordIDs) != 0 {
			return fmt.Errorf("selection_rule all_with_bounds members are invalid")
		}
	default:
		return fmt.Errorf("selection_rule rule_id is invalid")
	}
	for _, refs := range [][]string{decoded.VertexRefs, decoded.EdgeRefs, decoded.SeedVertexRefs, decoded.EdgeKindFilter, decoded.TimelineRecordIDs} {
		for _, ref := range refs {
			if ref == "" {
				return fmt.Errorf("selection_rule has an empty reference")
			}
		}
	}
	*rule = SelectionRule(decoded)
	return nil
}

func (rule SelectionRule) CompatibleSource(kind string) bool {
	switch kind {
	case "graph":
		return rule.RuleID == "explicit_refs" || rule.RuleID == "neighborhood" || rule.RuleID == "all_with_bounds"
	case "timeline":
		return rule.RuleID == "timeline_sequence"
	default:
		return false
	}
}

func hasDuplicateStrings(values []string) bool {
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		if _, exists := seen[value]; exists {
			return true
		}
		seen[value] = struct{}{}
	}
	return false
}
