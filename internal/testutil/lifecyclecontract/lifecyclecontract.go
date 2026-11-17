// Package lifecyclecontract validates illegal_transition details against the
// canonical error registry (Core 01 REQ-01-231 and REQ-01-238, Core 04
// AC-606). It reads only the generated contract projection.
package lifecyclecontract

import (
	"encoding/json"
	"errors"
	"slices"
	"sync"
	"testing"

	"github.com/JochiRaider/cartulary/internal/gen/contracterrors"
)

const errorRegistryPath = "contracts/errors/index.json"

// AnyMachine marks a guard that applies to every lifecycle action.
const AnyMachine = "any"

type registry struct {
	reasons map[string]bool
	guards  map[string]map[string]bool
}

var loadRegistry = sync.OnceValues(func() (registry, error) {
	artifact, ok := contracterrors.Index[errorRegistryPath]
	if !ok {
		return registry{}, errors.New("generated error registry is missing")
	}
	var raw struct {
		ReasonRegistries []struct {
			ErrorCode   string `json:"error_code"`
			ReasonCodes []struct {
				Code string `json:"code"`
			} `json:"reason_codes"`
		} `json:"reason_registries"`
		LifecycleGuards []struct {
			Code     string   `json:"code"`
			Machines []string `json:"machines"`
		} `json:"lifecycle_guards"`
	}
	if err := json.Unmarshal([]byte(artifact.JSON), &raw); err != nil {
		return registry{}, err
	}
	loaded := registry{reasons: map[string]bool{}, guards: map[string]map[string]bool{}}
	for _, family := range raw.ReasonRegistries {
		if family.ErrorCode != "illegal_transition" {
			continue
		}
		for _, reason := range family.ReasonCodes {
			loaded.reasons[reason.Code] = true
		}
	}
	for _, guard := range raw.LifecycleGuards {
		machines := map[string]bool{}
		for _, machine := range guard.Machines {
			machines[machine] = true
		}
		loaded.guards[guard.Code] = machines
	}
	if len(loaded.reasons) == 0 || len(loaded.guards) == 0 {
		return registry{}, errors.New("generated error registry has no illegal_transition vocabulary")
	}
	return loaded, nil
})

// RequireDetails validates one illegal_transition error.details object for
// the named lifecycle machine.
func RequireDetails(t testing.TB, machine string, details map[string]any) {
	t.Helper()
	for _, member := range []string{"from_status", "to_status", "violated_guards", "reason_code"} {
		if _, present := details[member]; !present {
			t.Fatalf("%s illegal_transition details omit %s: %#v", machine, member, details)
		}
	}
	var from *string
	switch value := details["from_status"].(type) {
	case nil:
	case string:
		from = &value
	default:
		t.Fatalf("%s illegal_transition from_status is not a string or null: %#v", machine, details["from_status"])
	}
	to, ok := details["to_status"].(string)
	if !ok {
		t.Fatalf("%s illegal_transition to_status is not a string: %#v", machine, details["to_status"])
	}
	reason, ok := details["reason_code"].(string)
	if !ok {
		t.Fatalf("%s illegal_transition reason_code is not a string: %#v", machine, details["reason_code"])
	}
	rawGuards, ok := details["violated_guards"].([]any)
	if !ok {
		// Typed callers may decode into []string.
		if typed, typedOK := details["violated_guards"].([]string); typedOK {
			for _, guard := range typed {
				rawGuards = append(rawGuards, guard)
			}
		} else {
			t.Fatalf("%s illegal_transition violated_guards is not an array: %#v", machine, details["violated_guards"])
		}
	}
	guards := make([]string, 0, len(rawGuards))
	for _, entry := range rawGuards {
		guard, ok := entry.(string)
		if !ok {
			t.Fatalf("%s illegal_transition guard is not a string: %#v", machine, entry)
		}
		guards = append(guards, guard)
	}
	RequireTransition(t, machine, from, to, reason, guards)
}

// RequireTransition validates typed illegal_transition values for the named
// lifecycle machine: a null from_status means a create or a record type
// without a machine.
func RequireTransition(t testing.TB, machine string, from *string, to string, reason string, guards []string) {
	t.Helper()
	loaded, err := loadRegistry()
	if err != nil {
		t.Fatalf("load illegal_transition registry: %v", err)
	}
	if from != nil && *from == "" {
		t.Fatalf("%s illegal_transition from_status is an empty string; use null", machine)
	}
	if to == "" {
		t.Fatalf("%s illegal_transition to_status is empty", machine)
	}
	if !loaded.reasons[reason] {
		t.Fatalf("%s illegal_transition reason_code %q is not registered", machine, reason)
	}
	if guards == nil {
		t.Fatalf("%s illegal_transition violated_guards is null; use []", machine)
	}
	if reason == "guards_not_satisfied" && len(guards) == 0 {
		t.Fatalf("%s illegal_transition guards_not_satisfied names no guard", machine)
	}
	for _, guard := range guards {
		machines, registered := loaded.guards[guard]
		if !registered || !(machines[machine] || machines[AnyMachine]) {
			t.Fatalf("%s illegal_transition guard %q is not registered for this machine", machine, guard)
		}
	}
	if slices.Contains(guards, "") {
		t.Fatalf("%s illegal_transition names an empty guard", machine)
	}
}
