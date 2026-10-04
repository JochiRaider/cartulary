package packformat

import (
	"context"
	"errors"
	"slices"
	"strconv"
	"strings"
)

// DependencyResolver authenticates the exact retained tuple and its physical
// availability. Unavailable content returns false; storage execution errors
// remain errors and must never become dependency verdicts.
type DependencyResolver func(context.Context, Dependency) (Manifest, bool, error)

type dependencyNode struct {
	key, version string
	edges        []Dependency
}

// VerifyDependencies executes all three dependency ranks before registry
// compatibility or indexing. Reachable state is already retained trusted
// content; hostile candidate declarations never supply recursive manifests.
func VerifyDependencies(ctx context.Context, candidate Manifest, resolve DependencyResolver, scratch DiagnosticScratch) error {
	program, err := DependencyChecks(candidate, resolve)
	if err != nil {
		return err
	}
	return runCheckRange(ctx, "dependencies", "conflicts", scratch, program, true)
}

func DependencyChecks(candidate Manifest, resolve DependencyResolver) (map[string]CheckFunc, error) {
	if resolve == nil {
		return nil, errors.New("reference pack: dependency resolver required")
	}
	identity := func(key, version string) string { return key + "\x00" + version }
	root := identity(candidate.Key, candidate.Version)
	nodes := map[string]dependencyNode{root: {candidate.Key, candidate.Version, candidate.Dependencies}}
	return map[string]CheckFunc{
		"dependencies": func(ctx context.Context, emit FindingSink) error {
			if profiles[candidate.Key].Required {
				for i := range candidate.Dependencies {
					if err := emit(Finding{Path: "$.dependencies[" + strconv.Itoa(i) + "]"}); err != nil {
						return err
					}
				}
				for i := range candidate.Conflicts {
					if err := emit(Finding{Path: "$.conflicts[" + strconv.Itoa(i) + "]"}); err != nil {
						return err
					}
				}
			}
			queue := []string{root}
			resolved := map[Dependency]bool{}
			for len(queue) > 0 {
				if err := ctx.Err(); err != nil {
					return err
				}
				id := queue[0]
				queue = queue[1:]
				for _, edge := range nodes[id].edges {
					if _, seen := resolved[edge]; seen {
						continue
					}
					_, supported := profiles[edge.Key]
					var target Manifest
					var available bool
					if supported {
						var err error
						target, available, err = resolve(ctx, edge)
						if err != nil {
							return err
						}
					}
					resolved[edge] = available
					if !available {
						if err := emit(dependencyFinding("$.dependencies", edge.Key, edge.Version)); err != nil {
							return err
						}
						continue
					}
					if target.Key != edge.Key || target.Version != edge.Version {
						return errors.New("reference pack: resolved dependency identity mismatch")
					}
					next := identity(edge.Key, edge.Version)
					if _, seen := nodes[next]; !seen {
						nodes[next] = dependencyNode{edge.Key, edge.Version, target.Dependencies}
						queue = append(queue, next)
					}
				}
			}
			return nil
		},
		"dependency_cycle": func(ctx context.Context, emit FindingSink) error {
			// Iterative Tarjan traversal bounds the process stack independently
			// of retained graph depth and identifies every cyclic participant.
			type frame struct {
				id   string
				edge int
			}
			frames := []frame{{id: root}}
			serial := 1
			index, low := map[string]int{root: serial}, map[string]int{root: serial}
			onStack, cyclic := map[string]bool{root: true}, map[string]bool{}
			stack := []string{root}
			for len(frames) > 0 {
				if err := ctx.Err(); err != nil {
					return err
				}
				at := len(frames) - 1
				current := &frames[at]
				id := current.id
				if current.edge < len(nodes[id].edges) {
					edge := nodes[id].edges[current.edge]
					current.edge++
					next := identity(edge.Key, edge.Version)
					if index[next] == 0 {
						serial++
						index[next], low[next] = serial, serial
						onStack[next] = true
						stack = append(stack, next)
						frames = append(frames, frame{id: next})
					} else if onStack[next] {
						low[id] = min(low[id], index[next])
					}
					continue
				}
				if low[id] == index[id] {
					component := []string{}
					for {
						next := stack[len(stack)-1]
						stack = stack[:len(stack)-1]
						delete(onStack, next)
						component = append(component, next)
						if next == id {
							break
						}
					}
					isCycle := len(component) > 1
					for _, edge := range nodes[id].edges {
						isCycle = isCycle || identity(edge.Key, edge.Version) == id
					}
					if isCycle {
						for _, member := range component {
							cyclic[member] = true
						}
					}
				}
				frames = frames[:at]
				if at > 0 {
					parent := frames[at-1].id
					low[parent] = min(low[parent], low[id])
				}
			}
			keys := make([]string, 0, len(cyclic))
			for id := range cyclic {
				keys = append(keys, id)
			}
			slices.Sort(keys)
			for _, id := range keys {
				node := nodes[id]
				if err := emit(dependencyFinding("$.dependencies", node.key, node.version)); err != nil {
					return err
				}
			}
			return nil
		},
		"conflicts": func(_ context.Context, emit FindingSink) error {
			for i, conflict := range candidate.Conflicts {
				bad := profiles[conflict.Key].Required || conflictMatches(conflict, candidate.Key, candidate.Version)
				for _, edge := range candidate.Dependencies {
					bad = bad || conflictMatches(conflict, edge.Key, edge.Version)
				}
				if bad {
					if err := emit(Finding{Path: "$.conflicts[" + strconv.Itoa(i) + "]"}); err != nil {
						return err
					}
				}
			}
			return nil
		},
	}, nil
}

func conflictMatches(conflict Conflict, key, version string) bool {
	return conflict.Key == key && (conflict.Version == nil || *conflict.Version == version)
}

func dependencyFinding(path, key, version string) Finding {
	return Finding{Path: path, Details: SafeDetails{RelatedPackKey: &key, RelatedPackVersion: &version}}
}

// SupportedDependencyKeys is a stable key-order projection for admission.
// Unknown keys cannot become available during one application/config revision.
func SupportedDependencyKeys() []string {
	keys := make([]string, 0, len(profiles))
	for key := range profiles {
		keys = append(keys, key)
	}
	slices.SortFunc(keys, strings.Compare)
	return keys
}
