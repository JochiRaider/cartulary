// Package runcontext is the single Go writer of a harness command's invocation
// context (TH-HARNESS Section 5.5, single writer): where the run's results go,
// which run, target and work unit the command reports for, the accounted
// invocation and host admission it runs under, the operation it executes, the
// services it selects, the retained resources it depends on and where its
// public inputs came from. Readers keep reading the names; every writer outside
// the Make entry point calls this package or its Node form,
// tools/harness/contract/run-context.mjs.
//
// It has no dependencies, so the planner, runtime, fixtures and tools can all
// use it.
package runcontext

import (
	"encoding/json"
	"maps"
	"slices"
	"strings"
)

// Context is a command's invocation context. Environment writes only the
// fields that are set: an empty value is never a meaningful context, so a
// writer that means "none" removes the name instead.
type Context struct {
	ResultsDir         string
	RunID              string
	Target             string
	WorkUnit           string
	PhysicalOutputRoot string
	InvocationRoot     string
	AdmissionLease     string
	OperationDigest    string
}

// Environment returns c as environment entries.
func (c Context) Environment() map[string]string {
	environment := map[string]string{}
	for name, value := range map[string]string{
		"CARTULARY_TEST_RESULTS_DIR":             c.ResultsDir,
		"CARTULARY_TEST_RUN_ID":                  c.RunID,
		"CARTULARY_TEST_TARGET":                  c.Target,
		"CARTULARY_WORK_UNIT_ID":                 c.WorkUnit,
		"CARTULARY_HARNESS_PHYSICAL_OUTPUT_ROOT": c.PhysicalOutputRoot,
		"CARTULARY_HARNESS_INVOCATION_ROOT":      c.InvocationRoot,
		"CARTULARY_HOST_ADMISSION_LEASE":         c.AdmissionLease,
		"CARTULARY_OPERATION_DIGEST":             c.OperationDigest,
	} {
		if value != "" {
			environment[name] = value
		}
	}
	return environment
}

// Apply writes c into environment and returns it.
func (c Context) Apply(environment map[string]string) map[string]string {
	maps.Copy(environment, c.Environment())
	return environment
}

// Services selects the test services a command may use, in the given order.
// It always writes: an empty selection is a selection of none, which differs
// from inheriting one.
func Services(services []string) map[string]string {
	return map[string]string{"CARTULARY_HARNESS_SERVICE_DEPENDENCIES": strings.Join(services, ",")}
}

// InputSources records where each public input came from as NAME=source
// entries, in the given order. It always writes, so an empty list clears an
// inherited record.
func InputSources(sources []string) map[string]string {
	return map[string]string{"CARTULARY_MAKE_INPUT_SOURCES": strings.Join(sources, " ")}
}

// ResourceDependencies names the retained resources a command depends on, as
// a JSON list. It always writes: no references is the empty list, which
// differs from inheriting a parent's.
func ResourceDependencies(references []string) map[string]string {
	if references == nil {
		references = []string{}
	}
	data, _ := json.Marshal(references)
	return map[string]string{"CARTULARY_HARNESS_RESOURCE_DEPENDENCIES": string(data)}
}

// Entries returns environment as sorted NAME=value entries for a process.
func Entries(environment map[string]string) []string {
	entries := make([]string, 0, len(environment))
	for _, name := range slices.Sorted(maps.Keys(environment)) {
		entries = append(entries, name+"="+environment[name])
	}
	return entries
}
