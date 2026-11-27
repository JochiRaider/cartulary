package harnessplan

import (
	"regexp"
	"slices"
	"strings"
)

// environmentPath is the declared harness environment (TH-HARNESS Section
// 5.5, environment registry).
const environmentPath = "tools/harness_environment_registry.json"

// EnvironmentRegistry declares every harness environment name by role, and
// the process and tool names a unit may take from the environment its graph
// starts in. The static check validates its shape and coverage; the planner
// derives each unit's inherited environment from it.
type EnvironmentRegistry struct {
	Roles     map[string][]string `json:"roles"`
	Inherited struct {
		Names    []string `json:"names"`
		Prefixes []string `json:"prefixes"`
	} `json:"inherited"`
}

// uninheritedRoles never legitimately come from a graph's starting
// environment: product configuration the harness sets for product processes,
// whether product code names it or derives its name. An operator's shell must
// not supply any of them to a unit. A test's own names are not registered, so
// a unit never inherits them either.
var uninheritedRoles = map[string]bool{"product_derived": true, "product_interface": true}

// operationBinding is a named Make variable an operation's execution uses;
// the operation runner resolves each one from its environment.
var operationBinding = regexp.MustCompile(`\$\(([A-Z][A-Z0-9_]*)\)`)

// inheritance is what a unit may take from its graph's environment
// (default-deny, P5-4b): registered harness names outside the uninherited
// roles, the registry's inherited process and tool names, the task surface's
// global inputs, and the Make variables operations name. Everything else a
// unit receives is set explicitly: forwarded inputs, plan bindings and the
// runtime's own variables.
func (in *Inputs) inheritance() (names, prefixes []string) {
	for _, operation := range in.Owner.Operations {
		for _, match := range operationBinding.FindAllSubmatch(operation.Execution, -1) {
			names = append(names, string(match[1]))
		}
	}
	for role, registered := range in.Environment.Roles {
		if !uninheritedRoles[role] {
			names = append(names, registered...)
		}
	}
	names = append(names, in.Environment.Inherited.Names...)
	for _, input := range in.TaskSurface.GlobalInputs {
		names = append(names, input.Name)
	}
	slices.Sort(names)
	return slices.Compact(names), slices.Clone(in.Environment.Inherited.Prefixes)
}

// validEnvironmentName is a portable environment variable name.
func validEnvironmentName(name string) bool {
	if name == "" || name[0] >= '0' && name[0] <= '9' {
		return false
	}
	return strings.Trim(name, "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789_") == ""
}
