package graphprojection

import (
	"context"
	"reflect"
	"slices"
)

// Reverse rows can precede their forwards in canonical order. Resolve links
// against the complete index and count their distinct forwards by mapping.
func validateRelationshipClaimsV2(ctx context.Context, view string, edges map[string]map[string]any, entityMappings, sourceIDs map[string]string, registry resultRegistryIndexV2) error {
	type claim struct {
		sourceKind, forwardKind, reverseKind string
		forwards, reverses                   int
		candidates                           []edgePolicyCandidateV2
	}
	// A source kind has exactly one relationship mapping. A false output
	// rules that mapping out as a supplier for every other output.
	excluded := make(map[string]bool)
	for _, kind := range registry.edges {
		if err := ctx.Err(); err != nil {
			return err
		}
		if !kind.preservesLabels {
			for source := range kind.sourceKinds {
				excluded[source] = true
			}
		}
	}
	claims := make(map[string]*claim)
	mappingByKind := make(map[string]string)
	aggregateDirections := make(map[string]string)
	reversed := make(map[string]bool)
	for _, edge := range edges {
		if err := ctx.Err(); err != nil {
			return err
		}
		meta := edge["metadata"].(map[string]any)
		reverse, ok := meta["is_reverse_edge"].(bool)
		link, hasLink := meta["reverse_of_edge_id"]
		if !ok || !hasLink || reverse != (edge["edge_family"] == "reverse") || (!reverse && link != nil) {
			return ErrResultV2Invalid
		}
		if edge["edge_family"] == "aggregated" {
			rule := meta["aggregation_rule_id"].(string)
			direction := edge["direction"].(string)
			if prior := aggregateDirections[rule]; prior != "" && prior != direction {
				return ErrResultV2Invalid
			}
			aggregateDirections[rule] = direction
			continue
		}
		ref := edge["source_relationship_ref"].(map[string]any)
		rule := ref["mapping_rule_id"].(string)
		source := ref["source_relationship_kind"].(string)
		kind := edge["edge_kind"].(string)
		if excluded[source] {
			for _, label := range edge["labels"].([]any) {
				if !registry.edges[kind].labels[label.(string)] {
					return ErrResultV2Invalid
				}
			}
		}
		if entityMappings[rule] != "" || (mappingByKind[source] != "" && mappingByKind[source] != rule) {
			return ErrResultV2Invalid
		}
		mappingByKind[source] = rule
		c := claims[rule]
		if c == nil {
			c = &claim{sourceKind: source}
			claims[rule] = c
		}
		if c.sourceKind != source {
			return ErrResultV2Invalid
		}
		if reverse {
			forward, _ := link.(string)
			if reversed[forward] || !validReverseLinkV2(edge, edges[forward]) || (c.reverseKind != "" && c.reverseKind != kind) {
				return ErrResultV2Invalid
			}
			reversed[forward] = true
			c.reverseKind = kind
			c.reverses++
		} else {
			id := ref["source_relationship_id"].(string)
			if len(sourceIDs) >= graphProjectionLimits.MaxSourceRelationships || sourceIDs[id] != "" || (c.forwardKind != "" && c.forwardKind != kind) {
				return ErrResultV2Invalid
			}
			sourceIDs[id] = source
			c.forwardKind = kind
			c.forwards++
		}
	}
	for rule, c := range claims {
		if err := ctx.Err(); err != nil {
			return err
		}
		if (c.reverses > 0 || registry.relationshipOutputs[c.sourceKind] == 2) && c.reverses != c.forwards {
			return ErrResultV2Invalid
		}
		if registry.relationshipOutputs[c.sourceKind] == 2 && c.forwardKind == c.reverseKind {
			return ErrResultV2Invalid
		}
		c.candidates = edgePolicyCandidatesV2(rule, c.sourceKind, c.forwardKind, c.reverseKind, registry.edges[c.forwardKind].directions)
	}
	// Only a policy that reproduces every observed row can explain this mapping.
	for _, edge := range edges {
		if err := ctx.Err(); err != nil {
			return err
		}
		family := edge["edge_family"].(string)
		if family == "aggregated" {
			continue
		}
		ref := edge["source_relationship_ref"].(map[string]any)
		c := claims[ref["mapping_rule_id"].(string)]
		direction := edge["direction"].(string)
		c.candidates = slices.DeleteFunc(c.candidates, func(candidate edgePolicyCandidateV2) bool {
			if !slices.Contains(candidate.directions, direction) {
				return true
			}
			id, key := mappedEdgeIdentityV2(view, family, edge["edge_kind"].(string), c.sourceKind, ref["source_relationship_id"].(string), edge["src_vertex_id"].(string), edge["dst_vertex_id"].(string), direction, candidate.digest)
			return id != edge["edge_id"] || key != edge["sort_key"]
		})
		if len(c.candidates) == 0 {
			return ErrResultV2Invalid
		}
	}
	// Each configured producer contributes exactly once, including dormant ones.
	// A source declared in two kinds necessarily uses normalized reverse emission.
	var singletonDirections, dormantMappingDirections []int
	for _, d := range registryDirectionsV2 {
		singletonDirections = append(singletonDirections, directionMaskV2(d))
	}
	for _, policy := range mappingDirectionPoliciesV2 {
		mask := directionMaskV2(policy.directions...)
		if !slices.Contains(dormantMappingDirections, mask) {
			dormantMappingDirections = append(dormantMappingDirections, mask)
		}
	}
	for _, kind := range registry.edges {
		if err := ctx.Err(); err != nil {
			return err
		}
		unions := directionUnionsV2{0: true}
		canPreserve := false
		for source := range kind.sourceKinds {
			canPreserve = canPreserve || !excluded[source]
			switch c := claims[mappingByKind[source]]; {
			case registry.relationshipOutputs[source] == 2:
				unions = unions.add(directionMaskV2("directed"))
			case c == nil:
				unions = unions.add(dormantMappingDirections...)
			default:
				var choices []int
				for _, candidate := range c.candidates {
					choices = append(choices, directionMaskV2(candidate.directions...))
				}
				unions = unions.add(choices...)
			}
		}
		for rule := range kind.rules {
			if direction := aggregateDirections[rule]; direction != "" {
				unions = unions.add(directionMaskV2(direction))
			} else {
				unions = unions.add(singletonDirections...)
			}
		}
		declared := 0
		for _, d := range kind.directions {
			declared |= directionMaskV2(d.(string))
		}
		if !unions[declared] || kind.preservesLabels && !canPreserve {
			return ErrResultV2Invalid
		}
	}
	return nil
}

// The three registry directions give eight possible unions, independent of producer count.
type directionUnionsV2 [8]bool

func directionMaskV2(directions ...string) int {
	mask := 0
	for _, d := range directions {
		mask |= 1 << slices.Index(registryDirectionsV2[:], d)
	}
	return mask
}

func (previous directionUnionsV2) add(choices ...int) (next directionUnionsV2) {
	for mask, possible := range previous {
		if possible {
			for _, choice := range choices {
				next[mask|choice] = true
			}
		}
	}
	return next
}

func validReverseLinkV2(reverse, forward map[string]any) bool {
	if forward["edge_family"] != "direct" || forward["direction"] != "directed" || reverse["direction"] != "directed" ||
		!slices.Equal(reverse["labels"].([]any), forward["labels"].([]any)) ||
		reverse["src_vertex_id"] != forward["dst_vertex_id"] || reverse["dst_vertex_id"] != forward["src_vertex_id"] {
		return false
	}
	// The envelope was verified canonical and decoded with UseNumber.
	// Deep equality preserves exact numeric spelling as well as member presence.
	if reverse["edge_kind"] == forward["edge_kind"] && (!reflect.DeepEqual(reverse["properties"], forward["properties"]) ||
		!reflect.DeepEqual(reverse["metadata"].(map[string]any)["mapped_metadata"], forward["metadata"].(map[string]any)["mapped_metadata"])) {
		return false
	}
	left, _ := reverse["source_relationship_ref"].(map[string]any)
	right, _ := forward["source_relationship_ref"].(map[string]any)
	reverseMapping, _ := reverse["metadata"].(map[string]any)["mapping_rule_id"].(string)
	forwardMapping, _ := forward["metadata"].(map[string]any)["mapping_rule_id"].(string)
	return left != nil && right != nil && reverseMapping != "" && reverseMapping == forwardMapping &&
		left["mapping_rule_id"] == reverseMapping && right["mapping_rule_id"] == forwardMapping &&
		left["source_relationship_kind"] == right["source_relationship_kind"] && left["source_relationship_id"] == right["source_relationship_id"]
}

type edgePolicyCandidateV2 struct {
	digest     string
	directions []string
}

func edgePolicyCandidatesV2(rule, sourceKind, forwardKind, reverseKind string, declared []any) []edgePolicyCandidateV2 {
	reverse := reverseKind != ""
	if !reverse {
		reverseKind = forwardKind
	}
	var candidates []edgePolicyCandidateV2
	for policy, capabilities := range mappingDirectionPoliciesV2 {
		if reverse && !capabilities.allowsReverse {
			continue
		}
		directions := capabilities.directions
		if slices.ContainsFunc(directions, func(d string) bool { return !slices.Contains(declared, any(d)) }) {
			continue
		}
		digest := relationshipMappingDigestV2(rule, sourceKind, forwardKind, policy, reverse, reverseKind)
		candidates = append(candidates, edgePolicyCandidateV2{digest: digest, directions: directions})
	}
	return candidates
}
