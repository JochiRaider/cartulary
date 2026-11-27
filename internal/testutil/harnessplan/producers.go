package harnessplan

import (
	"encoding/json"
	"fmt"
	"maps"
	"slices"
	"strings"
)

// waitForProducers checks that every operation unit waits, directly or
// through its needs, for the unit that produces each artifact its execution
// names (P6-5). The command entry produces only the harness runtime it
// launches; the graph alone produces everything its units use, so an
// operation that names an artifact without waiting for its producer would
// run before the artifact exists.
func (c *compiler) waitForProducers() error {
	type execution struct {
		Type          string   `json:"type"`
		Prerequisites []string `json:"prerequisites"`
	}
	executions := map[string]execution{}
	producers := map[string][]string{}
	for id, operation := range c.in.Owner.Operations {
		var e execution
		if err := json.Unmarshal(operation.Execution, &e); err != nil {
			return fmt.Errorf("%w: operation %s execution: %v", ErrInput, id, err)
		}
		executions[id] = e
		if e.Type == "artifact_binding" {
			for _, binding := range e.Prerequisites {
				producers[binding] = append(producers[binding], id)
			}
		}
	}
	ancestors := map[string]map[string]bool{}
	var walk func(id string) map[string]bool
	walk = func(id string) map[string]bool {
		if known, ok := ancestors[id]; ok {
			return known
		}
		set := map[string]bool{}
		ancestors[id] = set
		for _, need := range c.units[id].Needs {
			set[need] = true
			for ancestor := range walk(need) {
				set[ancestor] = true
			}
		}
		return set
	}
	for _, unitID := range slices.Sorted(maps.Keys(c.units)) {
		id, ok := strings.CutPrefix(unitID, "operation:")
		if !ok || executions[id].Type == "artifact_binding" || executions[id].Type == "installation" {
			continue
		}
		waits := walk(unitID)
		for _, binding := range executions[id].Prerequisites {
			candidates := producers[binding]
			if len(candidates) == 0 || slices.Contains(candidates, id) {
				continue
			}
			if !slices.ContainsFunc(candidates, func(producer string) bool { return waits["operation:"+producer] }) {
				return fmt.Errorf("%w: operation %s uses %s but does not wait for its producer %s", ErrInput, id, binding, strings.Join(candidates, " or "))
			}
		}
	}
	return nil
}
