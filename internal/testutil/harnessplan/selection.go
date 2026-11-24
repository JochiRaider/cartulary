package harnessplan

import (
	"fmt"
	"maps"
	"slices"
	"strings"
)

// Request is one graph invocation as the runtime receives it.
type Request struct {
	// Kind says how Target is read: "command" (a public command and its
	// selection inputs), "target" (an operation, a private union or a public
	// operation command), "owner" or "rows" (private slices run under the
	// public command Target). "aggregate" is not supported yet.
	Kind              string
	Target            string
	Owner             string
	Rows              []string
	ServiceBackedOnly bool
	// Environment is the runtime's environment. Preparation reads nothing
	// from the process environment.
	Environment map[string]string
	// CacheMode and InstrumentationMode, when set, are imposed by the
	// runtime's own policy (accounted runs) over the command's inputs.
	CacheMode, InstrumentationMode string
}

// selectionControls are public inputs the planner consumes to select work;
// they are never forwarded to an operation's command.
var selectionControls = []string{"BROWSER_LANES", "CHECK", "COMPONENT", "DETAIL", "GO_SHUFFLE_SEED", "JSON", "OWNER", "ROWS", "SCOPE", "SUITE"}

// command is a request resolved against its public command.
type command struct {
	entry     SurfaceTarget
	inputs    map[string]resolvedInput
	selection Selection
}

// resolve reads the request's public command, resolves its inputs and turns
// the request into a selection.
func (in *Inputs) resolve(request Request) (command, error) {
	fail := func(format string, args ...any) (command, error) {
		return command{}, fmt.Errorf("%w: %s", ErrSelection, fmt.Sprintf(format, args...))
	}
	entry, ok := in.TaskSurface.target(request.Target)
	if !ok || entry.CommandID == "" {
		return fail("%s is not a graph command", request.Target)
	}
	// Selection arguments are command-line inputs of the enclosing command.
	selected := map[string]string{}
	if request.Kind != "rows" {
		if request.Owner != "" {
			selected["OWNER"] = request.Owner
		}
		if len(request.Rows) != 0 {
			selected["ROWS"] = strings.Join(dedupe(request.Rows), ",")
		}
		if request.ServiceBackedOnly {
			selected["SCOPE"] = "services"
		}
	}
	environment := maps.Clone(request.Environment)
	var sources []string
	for _, token := range strings.Fields(environment[makeInputSources]) {
		if _, ok := selected[strings.SplitN(token, "=", 2)[0]]; !ok {
			sources = append(sources, token)
		}
	}
	for _, name := range slices.Sorted(maps.Keys(selected)) {
		environment[name] = selected[name]
		sources = append(sources, name+"="+originCommandLine)
	}
	environment[makeInputSources] = strings.Join(sources, " ")
	inputs, err := resolveInputs(entry, environment)
	if err != nil {
		return command{}, err
	}
	resolved := command{entry: entry, inputs: inputs}
	switch request.Kind {
	case "rows":
		if len(request.Rows) == 0 {
			return fail("a rows selection needs rows")
		}
		resolved.selection = Selection{Rows: dedupe(request.Rows)}
	case "owner":
		if request.Owner == "" || request.ServiceBackedOnly && inputs["SCOPE"].Value != "services" {
			return fail("an owner selection needs an owner")
		}
		rows, err := in.ownerRows(request.Owner, request.Rows, inputs)
		if err != nil {
			return command{}, err
		}
		resolved.selection = Selection{Rows: rows}
	case "target":
		resolved.selection, err = in.targetSelection(request.Target, entry, inputs)
	case "command":
		resolved.selection, err = in.commandSelection(entry, inputs)
	case "aggregate":
		resolved.selection, err = in.aggregateSelection(request.Target)
	default:
		return fail("unknown selection kind %q", request.Kind)
	}
	if err != nil {
		return command{}, err
	}
	if _, ok := inputs["GO_SHUFFLE_SEED"]; ok && !in.goOnly(resolved.selection) {
		return fail("GO_SHUFFLE_SEED requires a nonempty Go-only selection")
	}
	return resolved, nil
}

// ownerRows selects an owner's rows: the named rows, or every row the owner
// places in aggregates, narrowed to service-backed rows for SCOPE=services.
func (in *Inputs) ownerRows(owner string, named []string, inputs map[string]resolvedInput) ([]string, error) {
	var all []string
	placed := []string{}
	for _, row := range in.Catalog.Rows {
		if row.OwnerID == owner {
			all = append(all, row.RowID)
			if row.AggregatePlacement != "explicit_only" {
				placed = append(placed, row.RowID)
			}
		}
	}
	if len(all) == 0 {
		return nil, fmt.Errorf("%w: unknown active owner %s", ErrSelection, owner)
	}
	rows := placed
	if len(named) != 0 {
		for _, row := range named {
			if !slices.Contains(all, row) {
				return nil, fmt.Errorf("%w: %s does not belong to %s", ErrSelection, row, owner)
			}
		}
		rows = dedupe(named)
	}
	return in.serviceScope(rows, nil, len(named) != 0, inputs)
}

// serviceScope keeps only service-backed rows when SCOPE=services.
func (in *Inputs) serviceScope(rows, operations []string, named bool, inputs map[string]resolvedInput) ([]string, error) {
	if inputs["SCOPE"].Value != "services" {
		return rows, nil
	}
	var scoped []string
	for _, id := range rows {
		if len(in.row(id).ServiceDependencies) != 0 {
			scoped = append(scoped, id)
		}
	}
	if named && len(scoped) != len(rows) {
		return nil, fmt.Errorf("%w: a requested row is outside service scope", ErrSelection)
	}
	if len(scoped) == 0 || len(operations) != 0 {
		return nil, fmt.Errorf("%w: the selection has no service-only row scope", ErrSelection)
	}
	return scoped, nil
}

// targetSelection selects an operation, a private union, or the operation
// or command selection a public target stands for.
func (in *Inputs) targetSelection(name string, entry SurfaceTarget, inputs map[string]resolvedInput) (Selection, error) {
	if _, ok := in.Owner.TargetMembers[name]; ok {
		return Selection{Targets: []string{name}}, nil
	}
	if _, ok := in.Owner.Operations[name]; ok {
		return Selection{Targets: []string{name}}, nil
	}
	if entry.TargetClass != "public" {
		return Selection{}, fmt.Errorf("%w: unknown public command %s", ErrSelection, name)
	}
	if _, ok := in.Owner.AggregateTiers[name]; ok {
		return in.aggregateSelection(name)
	}
	if entry.SelectionBinding != nil {
		return in.commandSelection(entry, inputs)
	}
	if recipe := in.TaskSurface.MakeRecipes[name]; recipe.OperationID != "" {
		return Selection{Targets: []string{recipe.OperationID}, Forwarded: forwardedInputs(entry, inputs)}, nil
	}
	return Selection{}, fmt.Errorf("%w: command %s has no execution selection", ErrSelection, name)
}

// aggregateSelection selects an aggregate's tier: every row placed at or
// below the tier, and the policy roots of the aggregate and of the
// aggregates it inherits.
func (in *Inputs) aggregateSelection(name string) (Selection, error) {
	tier, ok := in.Owner.AggregateTiers[name]
	if !ok {
		return Selection{}, fmt.Errorf("%w: %s is not an aggregate", ErrSelection, name)
	}
	rank := slices.Index(in.Owner.TierOrder, tier)
	var rows []string
	for _, row := range in.Catalog.Rows {
		if placement := slices.Index(in.Owner.TierOrder, row.AggregatePlacement); placement >= 0 && placement <= rank {
			rows = append(rows, row.RowID)
		}
	}
	var roots []string
	pending, seen := []string{name}, map[string]bool{}
	for len(pending) != 0 {
		aggregate := pending[0]
		pending = pending[1:]
		if seen[aggregate] {
			continue
		}
		seen[aggregate] = true
		roots = append(roots, in.Owner.AggregatePolicyRoots[aggregate]...)
		pending = append(pending, in.Owner.AggregatePolicyInherits[aggregate]...)
	}
	return Selection{Rows: rows, Targets: dedupe(roots)}, nil
}

// commandSelection resolves a command's selection binding: an owner slice
// (owner_or_suite with OWNER), or the private selection its inputs name.
func (in *Inputs) commandSelection(entry SurfaceTarget, inputs map[string]resolvedInput) (Selection, error) {
	binding := entry.SelectionBinding
	if binding == nil {
		return Selection{}, fmt.Errorf("%w: %s has no command selection binding", ErrSelection, entry.Name)
	}
	if owner, ok := inputs["OWNER"]; ok && binding.Kind == "owner_or_suite" {
		var named []string
		if rows, ok := inputs["ROWS"]; ok {
			named = strings.Split(rows.Value, ",")
		}
		rows, err := in.ownerRows(owner.Value, named, inputs)
		return Selection{Rows: rows}, err
	}
	name := binding.Selection
	if binding.Kind != "fixed" {
		key, ok := inputs[binding.Input]
		name = binding.DefaultSelection
		if ok {
			name = binding.Variants[key.Value]
		}
	}
	if name == "" {
		return Selection{}, fmt.Errorf("%w: %s has no selection for %s", ErrSelection, entry.Name, binding.Input)
	}
	rows, operations, err := in.flatten(name, nil)
	if err != nil {
		return Selection{}, err
	}
	var placed []string
	for _, id := range dedupe(rows) {
		if in.row(id).AggregatePlacement != "explicit_only" {
			placed = append(placed, id)
		}
	}
	operations = dedupe(operations)
	if placed, err = in.serviceScope(placed, operations, false, inputs); err != nil {
		return Selection{}, err
	}
	return Selection{Rows: placed, Targets: operations, Forwarded: forwardedInputs(entry, inputs)}, nil
}

// flatten expands a private selection into its rows and operations.
func (in *Inputs) flatten(name string, visiting []string) (rows, operations []string, err error) {
	if slices.Contains(visiting, name) {
		return nil, nil, fmt.Errorf("%w: private selection cycle at %s", ErrSelection, name)
	}
	if _, ok := in.Owner.Operations[name]; ok {
		return nil, []string{name}, nil
	}
	members, ok := in.Owner.TargetMembers[name]
	if !ok {
		return nil, nil, fmt.Errorf("%w: unknown private selection %s", ErrSelection, name)
	}
	visiting = append(slices.Clone(visiting), name)
	for _, member := range members {
		switch {
		case member.RowID != "":
			rows = append(rows, member.RowID)
		case member.Catalog != nil:
			selected := in.catalogRows(*member.Catalog)
			if len(selected) == 0 {
				return nil, nil, fmt.Errorf("%w: %s selects no catalog rows", ErrSelection, name)
			}
			rows = append(rows, selected...)
		default:
			next := member.Name
			if member.OperationID != "" {
				next = member.OperationID
			}
			memberRows, memberOperations, err := in.flatten(next, visiting)
			if err != nil {
				return nil, nil, err
			}
			rows, operations = append(rows, memberRows...), append(operations, memberOperations...)
		}
	}
	return rows, operations, nil
}

// forwardedInputs are the command's argv inputs that reach its operations:
// present, not redacted and not consumed by selection.
func forwardedInputs(entry SurfaceTarget, inputs map[string]resolvedInput) map[string]string {
	forwarded := map[string]string{}
	for name, input := range inputs {
		control := slices.Contains(selectionControls, name) && (name != "OWNER" || entry.SelectionBinding != nil)
		if input.Definition.ChildForwarding == "argv" && input.Definition.SummaryEmission != "redacted_value" && !control {
			forwarded[name] = input.Value
		}
	}
	return forwarded
}

// goOnly reports whether a selection is nonempty Go rows and nothing else.
func (in *Inputs) goOnly(selection Selection) bool {
	if len(selection.Rows) == 0 || len(selection.Targets) != 0 {
		return false
	}
	for _, id := range selection.Rows {
		if in.row(id).Runner != RunnerGo {
			return false
		}
	}
	return true
}

// row returns a catalog row, or the zero row when the ID is unknown; the
// compiler rejects unknown rows.
func (in *Inputs) row(id string) Row {
	index, ok := in.rowIndex[id]
	if !ok {
		return Row{}
	}
	return in.Catalog.Rows[index]
}
