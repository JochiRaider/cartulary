package harnessplan

import (
	"encoding/json"
	"errors"
	"slices"
	"strings"
)

// projections names the parts of a command or aggregate graph that targets
// report on, each closed under needs:
//
//   - the suite each row belongs to (backend-<suite>, frontend-unit, or a
//     shell row's operation) and each operation;
//   - every other public command, and every gate smoke check, whose own plan
//     the graph contains: its rows are bound here and its operations are
//     here with identical digests.
//
// A public command's projection carries its own command ID; every other
// projection carries the selected command's.
func (in *Inputs) projections(graph *Graph, selected SurfaceTarget, options Options) (map[string]Projection, error) {
	members := map[string]map[string]bool{}
	add := func(name, unit string) {
		if members[name] == nil {
			members[name] = map[string]bool{}
		}
		members[name][unit] = true
	}
	bound := map[string]string{}
	for id, unit := range graph.Units {
		for _, binding := range unit.RowBindings {
			bound[binding.RowID] = id
			add(in.rowSuite(in.row(binding.RowID)), id)
		}
		if operation, ok := cutOperation(id); ok {
			add(operation, id)
		}
		if target, ok := strings.CutPrefix(id, "browser_target_summary:"); ok {
			add(target, id)
		}
	}
	commandIDs := map[string]string{}
	contains := func(name string, selection Selection) error {
		// A candidate the planner cannot compile cannot be contained.
		candidate, err := Compile(in, selection, options)
		if errors.Is(err, ErrSelection) {
			return nil
		}
		if err != nil {
			return err
		}
		var ids []string
		for id, unit := range candidate.Units {
			if _, operation := cutOperation(id); !operation {
				for _, binding := range unit.RowBindings {
					owner, ok := bound[binding.RowID]
					if !ok {
						return nil
					}
					ids = append(ids, owner)
				}
				continue
			}
			if present, ok := graph.Units[id]; !ok || present.SemanticDigest != unit.SemanticDigest {
				return nil
			}
			ids = append(ids, id)
		}
		for _, id := range ids {
			add(name, id)
		}
		return nil
	}
	for _, target := range in.TaskSurface.Targets {
		if !in.projectable(target, selected) {
			continue
		}
		defaults, err := resolveInputs(target, map[string]string{})
		if err != nil {
			continue
		}
		selection, err := in.targetSelection(target.Name, target, defaults)
		if err != nil {
			continue
		}
		commandIDs[target.Name] = target.CommandID
		if err := contains(target.Name, selection); err != nil {
			return nil, err
		}
	}
	for _, check := range in.TaskSurface.HarnessChecks {
		if check.GateSmokeRole == "" {
			continue
		}
		selection := Selection{}
		switch {
		case check.Selection.RowID != "":
			selection.Rows = []string{check.Selection.RowID}
		case check.Selection.OperationID != "":
			selection.Targets = []string{check.Selection.OperationID}
		default:
			selection.Targets = []string{check.Selection.Name}
		}
		if err := contains(check.Name, selection); err != nil {
			return nil, err
		}
	}
	projections := map[string]Projection{}
	for name, units := range members {
		var pending []string
		for id := range units {
			pending = append(pending, id)
		}
		for len(pending) != 0 {
			id := pending[0]
			pending = pending[1:]
			for _, need := range graph.Units[id].Needs {
				if !units[need] {
					units[need] = true
					pending = append(pending, need)
				}
			}
		}
		commandID := commandIDs[name]
		if commandID == "" {
			commandID = selected.CommandID
		}
		ids := make([]string, 0, len(units))
		for id := range units {
			ids = append(ids, id)
		}
		slices.Sort(ids)
		projections[name] = Projection{CommandID: commandID, UnitIDs: ids}
	}
	return projections, nil
}

// projectable reports whether a public command can be projected from another
// command's graph: it is a different graph command without required inputs,
// owner slices, browser modes or source writes, selected through a binding
// or a graph-entry operation.
func (in *Inputs) projectable(target, selected SurfaceTarget) bool {
	if target.Name == selected.Name || target.TargetClass != "public" || target.CommandID == "" {
		return false
	}
	if _, aggregate := in.Owner.AggregateTiers[target.Name]; aggregate {
		return false
	}
	if target.InputContract != nil && slices.ContainsFunc(target.InputContract.Inputs, func(input InputDefinition) bool { return input.Required }) {
		return false
	}
	if slices.ContainsFunc(target.SideEffects, func(effect SideEffect) bool { return effect.Class == "authored_source_write" }) {
		return false
	}
	recipe := in.TaskSurface.MakeRecipes[target.Name]
	if binding := target.SelectionBinding; binding != nil {
		return recipe.Type == "work_graph" && binding.Kind != "owner_or_suite" && binding.BrowserMode == ""
	}
	if recipe.OperationID == "" {
		return false
	}
	var execution struct {
		GraphEntry bool `json:"graph_entry"`
	}
	return json.Unmarshal(in.Owner.Operations[recipe.OperationID].Execution, &execution) == nil && execution.GraphEntry
}

// rowSuite is the suite a row reports under: its Go suite, the frontend
// unit suite, its browser stage's target, or a shell row's operation.
func (in *Inputs) rowSuite(row Row) string {
	switch row.Runner {
	case RunnerGo:
		return "backend-" + goSuite(row)
	case RunnerVitest:
		return "frontend-unit"
	case RunnerPlaywright:
		for _, stage := range in.Browser.Stages {
			for _, membership := range stage.Groups {
				if slices.Contains(membership.SelectedRowIDs, row.RowID) {
					return stage.Target
				}
			}
		}
		return ""
	default:
		return row.Selector.OperationID
	}
}

func cutOperation(id string) (string, bool) { return strings.CutPrefix(id, "operation:") }
