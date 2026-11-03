package extensions

import (
	"fmt"
	"slices"
	"sort"

	"github.com/JochiRaider/cartulary/internal/modules/extensions/registryfacts"
)

// validateClaimResolution rebinds the immutable execution decision at every
// execution boundary. Canonical membership remains independent of graph order.
func (c *Coordinator) validateClaimResolution(resolution ClaimResolution) ([]string, error) {
	invalid := func() ([]string, error) {
		return nil, validationFailure(Finding{Code: "extension_registry_invalid", Phase: "dependency_validation", Actual: "claim_resolution_registry_mismatch"})
	}
	if c == nil || resolution.registrySHA256 == "" || resolution.registrySHA256 != c.registrySHA256 {
		return invalid()
	}
	current, err := c.ResolveClaims(resolution.claims.ProfileIDs())
	if err != nil {
		return nil, err
	}
	if current.claims.SHA256() != resolution.claims.SHA256() || !slices.Equal(current.admissionOrder, resolution.admissionOrder) {
		return invalid()
	}
	return current.AdmissionOrder(), nil
}

func (c *Coordinator) dependencyOrder(requested map[string]struct{}) ([]string, error) {
	vertices := make([]string, 0, len(requested))
	for id := range requested {
		vertices = append(vertices, id)
	}
	sort.Strings(vertices)
	graph := make(map[string][]string, len(vertices))
	findings := []Finding{}
	emit := func(finding Finding) bool {
		if len(findings) == registryfacts.DiagnosticLimit {
			return false
		}
		findings = append(findings, finding)
		return true
	}
	overflow := func() ([]string, error) {
		return nil, validationFailure(Finding{Code: "extension_diagnostic_overflow", Phase: "dependency_validation"})
	}
	for _, id := range vertices {
		counts := map[string]int{}
		for _, dependency := range c.profiles[id].descriptor.Dependencies {
			target := dependency.ProfileID
			if !extensionProfileIDPattern.MatchString(target) || target == "base" || dependency.RequiredContractMajor < 1 {
				if !emit(Finding{Code: "extension_registry_invalid", Phase: "dependency_validation", ProfileID: id, Actual: "invalid_dependency"}) {
					return overflow()
				}
				continue
			}
			counts[target]++
			if target == id {
				if !emit(dependencyCollision("dependency_self", []string{id}, []string{id + "->" + id}, 0)) {
					return overflow()
				}
			}
			if _, claimed := requested[target]; !claimed {
				if !emit(Finding{Code: "extension_dependency_not_claimed", Phase: "dependency_validation", ProfileID: id, DependencyProfileID: target}) {
					return overflow()
				}
			} else if counts[target] == 1 {
				graph[id] = append(graph[id], target)
			}
			record, recognized := c.profiles[target]
			switch {
			case !recognized:
				if !emit(Finding{Code: "extension_profile_unrecognized", Phase: "dependency_validation", ProfileID: target}) {
					return overflow()
				}
			case !record.descriptor.Claimable:
				if !emit(Finding{Code: "extension_profile_not_claimable", Phase: "dependency_validation", ProfileID: target}) {
					return overflow()
				}
			case record.bindingSHA256 == "":
				if !emit(Finding{Code: "extension_implementation_unavailable", Phase: "implementation_binding", ProfileID: target, Actual: "binding_missing"}) {
					return overflow()
				}
			case record.descriptor.ContractMajor != dependency.RequiredContractMajor:
				if !emit(Finding{Code: "extension_dependency_incompatible", Phase: "dependency_validation", ProfileID: id, DependencyProfileID: target, Expected: fmt.Sprint(dependency.RequiredContractMajor), Actual: fmt.Sprint(record.descriptor.ContractMajor)}) {
					return overflow()
				}
			}
		}
		for target, count := range counts {
			if count > 1 {
				if !emit(dependencyCollision("dependency_duplicate", []string{id}, []string{id + "->" + target}, count)) {
					return overflow()
				}
			}
		}
		sort.Strings(graph[id])
	}
	for _, component := range registryfacts.DependencyComponents(vertices, graph) {
		if len(component) < 2 {
			continue
		}
		members := map[string]bool{}
		for _, id := range component {
			members[id] = true
		}
		edges := []string{}
		for _, id := range component {
			for _, target := range graph[id] {
				if members[target] {
					edges = append(edges, id+"->"+target)
				}
			}
		}
		if !emit(dependencyCollision("dependency_cycle", component, edges, 0)) {
			return overflow()
		}
	}
	if len(findings) > 0 {
		sortFindings(findings)
		return nil, validationFailure(findings...)
	}
	order := make([]string, 0, len(vertices))
	emitted := map[string]bool{}
	for len(order) < len(vertices) {
		next := ""
		for _, id := range vertices {
			if emitted[id] {
				continue
			}
			ready := true
			for _, target := range graph[id] {
				if !emitted[target] {
					ready = false
					break
				}
			}
			if ready {
				next = id
				break
			}
		}
		if next == "" {
			return nil, validationFailure(Finding{Code: "extension_registry_invalid", Phase: "dependency_validation", Actual: "dependency_order_incomplete"})
		}
		emitted[next] = true
		order = append(order, next)
	}
	return order, nil
}

func dependencyCollision(class string, profiles, edges []string, count int) Finding {
	return Finding{Code: "extension_registry_conflict", Phase: "dependency_validation", CollisionClass: class,
		ConflictingProfileIDs: append([]string(nil), profiles...), ConflictingTokens: append([]string(nil), edges...), DuplicateCount: count}
}
