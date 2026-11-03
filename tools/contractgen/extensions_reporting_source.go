package main

import (
	"fmt"
	"strings"
)

func validateExtensionReportingSource(object map[string]any, path string) error {
	if err := requireExtensionExactKeys(object, stringSet("schema_id", "profile_id", "participant_id", "participant_kind", "projection_schema_id", "label_schema_id", "read_algorithm_id", "ordering_algorithm_id", "state_family_ids", "max_vertices", "max_edges", "max_label_bytes"), path); err != nil {
		return err
	}
	for _, key := range []string{"profile_id", "participant_id", "projection_schema_id", "label_schema_id", "read_algorithm_id", "ordering_algorithm_id"} {
		if _, err := requiredString(object, key, path); err != nil {
			return err
		}
	}
	if object["participant_kind"] != "snapshot_reporting" || !strings.HasPrefix(stringValue(object["participant_id"]), stringValue(object["profile_id"])+".") {
		return fmt.Errorf("%s has invalid source ownership", path)
	}
	if _, err := sortedUniqueStringArray(object["state_family_ids"], path+".state_family_ids", false); err != nil {
		return err
	}
	for key, max := range map[string]int64{"max_vertices": 100000, "max_edges": 250000, "max_label_bytes": 67108864} {
		n, err := positiveJSONInt(object[key], path+"."+key)
		if err != nil || n > max {
			return fmt.Errorf("%s has invalid source bounds", path)
		}
	}
	return nil
}
