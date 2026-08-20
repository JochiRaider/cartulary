package timeline

import "github.com/JochiRaider/cartulary/internal/platform/workbookprobe"

const restoreWorkbookProbeRegistrationID = "timeline.base_restore_probe.v1"

// RestoreWorkbookProbeRegistration contributes Timeline's exact Base restore
// verification query. Workbook owns validation and execution.
func RestoreWorkbookProbeRegistration() workbookprobe.Registration {
	return workbookprobe.Registration{
		SchemaID:       workbookprobe.RegistrationSchemaID,
		RegistrationID: restoreWorkbookProbeRegistrationID,
		OwnerID:        "module.timeline",
		Profile:        workbookprobe.BaseProfile,
		IsDefault:      true,
		ViewSchemaID:   TimelineViewSchemaID,
		Filters:        []workbookprobe.Filter{},
		Sort: []workbookprobe.Sort{
			{FieldKey: "timeline.activity_sort_ts", Direction: "asc"},
			{FieldKey: "record_id", Direction: "asc"},
		},
		GroupBy:        nil,
		RowRequirement: "zero_rows_allowed",
	}
}
