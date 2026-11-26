package harnessruntime

import (
	"encoding/json"
	"maps"
	"slices"
	"strings"
)

// unitOverlayKey names the variable that records what a unit's own overlay
// replaced: its planned command environment (for a Go unit, the fixed Go build
// context) and the runtime's Go binding. Each name maps to the value it had
// before the overlay, or null when it was unset.
const unitOverlayKey = "CARTULARY_HARNESS_UNIT_OVERLAY"

// recordOverlay returns the overlay record for names, read from the
// environment as it stood before the overlay.
func recordOverlay(before map[string]string, names []string) (string, error) {
	record := map[string]*string{}
	for _, name := range names {
		if value, ok := before[name]; ok {
			record[name] = &value
		} else {
			record[name] = nil
		}
	}
	data, err := json.Marshal(record)
	return string(data), err
}

// NestedEnvironment returns environ with the enclosing unit's overlay undone.
// A harness invocation launched from inside a unit (a nested graph or a real
// make command) starts from it, so it prepares and builds with the
// environment a top-level invocation has, never its host test's Go build
// context. Outside a unit, environ is returned unchanged.
func NestedEnvironment(environ []string) ([]string, error) {
	environment := map[string]string{}
	for _, entry := range environ {
		if name, value, ok := strings.Cut(entry, "="); ok {
			environment[name] = value
		}
	}
	if data, ok := environment[unitOverlayKey]; ok {
		var record map[string]*string
		if err := json.Unmarshal([]byte(data), &record); err != nil {
			return nil, err
		}
		for name, value := range record {
			if value == nil {
				delete(environment, name)
			} else {
				environment[name] = *value
			}
		}
		delete(environment, unitOverlayKey)
	}
	result := make([]string, 0, len(environment))
	for _, name := range slices.Sorted(maps.Keys(environment)) {
		result = append(result, name+"="+environment[name])
	}
	return result, nil
}
