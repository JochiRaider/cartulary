package sourcevalue

import "fmt"

func ValidateSnapshotSource(record, source map[string]any) error {
	row, valid := Parse(source)
	if !valid || source["record_id"] != record["record_id"] || source["incident_id"] != record["incident_id"] {
		return fmt.Errorf("invalid historical Evidence source shape")
	}
	if !row.LifecycleLegal() {
		return fmt.Errorf("invalid historical Evidence lifecycle or blob bridge")
	}
	if !row.StorageReferenceLegal() {
		return fmt.Errorf("invalid historical Evidence logical reference association")
	}
	return nil
}
