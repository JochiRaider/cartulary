package reporting

import "context"

// Timeline render inputs reuse the independently admitted Core ordering facts.
// Only selected retained field values may contribute display content. A source
// event with no retained fields cannot be resurrected by a diagram declaration.
func projectRetainedTimeline(ctx context.Context, source []ReportingTimelineEvent, fields map[string]ReportingField) ([]ReportingTimelineEvent, error) {
	events := make([]ReportingTimelineEvent, 0, len(source))
	for _, original := range source {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		event := original
		event.Fields = []ReportingField{}
		refs := emptyContainerRefs()
		for _, originalField := range original.Fields {
			if len(originalField.SourceRefs) == 0 {
				return nil, newRenderValidationError("export_model_invalid", "dangling_source_ref")
			}
			field, retained := fields[originalField.SourceRefs[0]]
			if !retained {
				continue
			}
			field.FieldOrdinal = len(event.Fields) + 1
			event.Fields = append(event.Fields, field)
			refs.include(field.SourceRefs, field.SupportRefs, field.DisclosurePartitionRefs)
		}
		if len(event.Fields) == 0 {
			continue
		}
		refs = refs.normalized()
		event.SourceRefs, event.SupportRefs, event.DisclosurePartitionRefs = refs.source, refs.support, refs.partitions
		// No template display binding has selected a time field at this stage.
		// In particular, Core's private sort timestamp is never display text.
		event.DisplayTimes = ReportingDisplayTimes{}
		if original.ActivitySortTS != nil {
			value := *original.ActivitySortTS
			event.ActivitySortTS = &value
		}
		if original.DateEnteredSortKey != nil {
			value := *original.DateEnteredSortKey
			event.DateEnteredSortKey = &value
		}
		events = append(events, event)
	}
	return events, nil
}
