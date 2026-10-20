//go:build linux

package harnessfixtures

import (
	"context"
	"errors"
	"slices"

	"github.com/JochiRaider/cartulary/internal/testutil/harnesssession"
)

var errCleanupUncertain = errors.New("cleanup relationships are incomplete or contradictory")

// evaluateCleanup is the sole dependency policy for ordinary and recovery
// closure. Edges point from consumers to dependencies. A selected target closes
// only after all of its retained consumers; empty target selects the whole set.
func evaluateCleanup(resources []harnesssession.Resource, target string, attempt func(harnesssession.Resource) error, report func(CleanupOutcome)) error {
	nodes := map[string]harnesssession.Resource{}
	for _, resource := range resources {
		nodes[resource.Reference()] = resource
	}
	selected := map[string]bool{}
	if target == "" {
		for ref := range nodes {
			selected[ref] = true
		}
	} else {
		selected[target] = true
	}
	for changed := true; changed; {
		changed = false
		for ref, resource := range nodes {
			if selected[ref] {
				continue
			}
			for _, dependency := range resource.Dependencies {
				if selected[dependency] {
					selected[ref] = true
					changed = true
					break
				}
			}
		}
	}
	invalid := map[string]bool{}
	for ref, resource := range nodes {
		if resource.Dependencies == nil {
			invalid[ref] = true
		}
		for _, dependency := range resource.Dependencies {
			if _, ok := nodes[dependency]; !ok || dependency == ref {
				invalid[ref] = true
			}
		}
	}
	refs := make([]string, 0, len(nodes))
	for ref := range nodes {
		refs = append(refs, ref)
	}
	slices.Sort(refs)
	failures := []error{}
	finish := func(ref, outcome string, err error) {
		if err != nil {
			failures = append(failures, err)
		}
		if report != nil {
			report(CleanupOutcome{Kind: nodes[ref].Kind, Outcome: outcome, Err: err})
		}
	}
	// An unknown edge may name any retained dependency. There is no evidence
	// that permits treating the other nodes as independent of that consumer.
	if len(invalid) > 0 {
		for _, ref := range refs {
			if invalid[ref] {
				finish(ref, "failed", harnesssession.AtCleanupStage("resource_graph", errCleanupUncertain))
			} else if selected[ref] {
				finish(ref, "blocked", harnesssession.AtCleanupStage("resource_graph", ErrCleanupBlocked))
			}
		}
		return errors.Join(failures...)
	}
	pending, blocked := map[string]bool{}, map[string]bool{}
	for ref := range nodes {
		pending[ref] = selected[ref]
	}
	var block func(string)
	block = func(ref string) {
		if blocked[ref] {
			return
		}
		blocked[ref] = true
		for _, dependency := range nodes[ref].Dependencies {
			block(dependency)
		}
	}
	for {
		progress := false
		for _, ref := range refs {
			if !pending[ref] {
				continue
			}
			consumed := false
			for consumer, resource := range nodes {
				if pending[consumer] && slices.Contains(resource.Dependencies, ref) {
					consumed = true
					break
				}
			}
			if consumed {
				continue
			}
			pending[ref] = false
			progress = true
			if blocked[ref] {
				finish(ref, "blocked", harnesssession.AtCleanupStage("resource_graph", ErrCleanupBlocked))
				continue
			}
			err := attempt(nodes[ref])
			if err == nil {
				finish(ref, "completed", nil)
				continue
			}
			outcome := "failed"
			if errors.Is(err, ErrCleanupBlocked) {
				outcome = "blocked"
			}
			finish(ref, outcome, err)
			for _, dependency := range nodes[ref].Dependencies {
				block(dependency)
			}
			if errors.Is(err, errCleanupUncertain) {
				for other := range nodes {
					block(other)
				}
			}
		}
		if !progress {
			break
		}
	}
	// A remaining strongly connected component cannot be ordered safely. Its
	// dependencies remain pending as well; unrelated acyclic branches progressed.
	for _, ref := range refs {
		if pending[ref] {
			finish(ref, "blocked", harnesssession.AtCleanupStage("resource_graph", ErrCleanupBlocked))
		}
	}
	return errors.Join(failures...)
}

func (p ProductionOptions) closeResources(ctx context.Context, target string, override func(context.Context) error) error {
	resources, err := p.Runtime.Resources(4096)
	if err != nil {
		return artifactFailure(harnesssession.AtCleanupStage("resource_inventory", err))
	}
	return CleanupFailure(evaluateCleanup(resources, target, func(resource harnesssession.Resource) error {
		var err error
		if target != "" && resource.Reference() == target && override != nil {
			err = override(ctx)
		} else {
			err = p.cleanupResource(ctx, resource, resources)
		}
		if err == nil {
			err = p.releaseResource(resource)
		}
		return err
	}, p.ReportCleanup))
}

func (p ProductionOptions) closeResource(ctx context.Context, kind string, target any, fn func(context.Context) error) error {
	reference, err := p.Runtime.ResourceReference(kind, target)
	if err != nil {
		return err
	}
	return p.closeResources(ctx, reference, fn)
}
