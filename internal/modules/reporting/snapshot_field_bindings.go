package reporting

import (
	"bytes"
	"slices"
	"sort"
)

// Sections, records and Timeline events provide different views of one frozen
// source. Examine every binding before returning the bounded stage result;
// conflicting copies never become a usable partial lookup.
func validateSnapshotFieldBindings(model SnapshotContent) error {
	issues := renderIssueCollector{stage: "export_model_materialization"}
	invalid := func(reason string) { issues.add(newRenderValidationError("export_model_invalid", reason)) }
	bySource := map[string][]byte{}
	ambiguousSources := map[string]bool{}
	var collect func([]ReportingBlock)
	collect = func(blocks []ReportingBlock) {
		for _, block := range blocks {
			for _, field := range block.Fields {
				if len(field.SourceRefs) == 0 {
					invalid("dangling_source_ref")
					continue
				}
				encoded, err := reportingFieldContent(field)
				if err != nil {
					invalid("invalid_scalar_value")
					continue
				}
				for _, ref := range field.SourceRefs {
					if _, exists := bySource[ref]; exists {
						invalid("duplicate_export_model_path")
						ambiguousSources[ref] = true
						continue
					}
					bySource[ref] = encoded
				}
			}
			collect(block.Children)
		}
	}
	for _, section := range model.Sections {
		collect(section.Blocks)
	}
	type sourceKey struct{ family, id string }
	records := map[sourceKey]ReportingRecordSummary{}
	ambiguousRecords := map[sourceKey]bool{}
	bound := map[string]bool{}
	for _, record := range model.Records {
		key := sourceKey{record.SourceRecordRef.SourceFamily, record.RecordID}
		if _, exists := records[key]; exists {
			invalid("duplicate_stable_id")
			ambiguousRecords[key] = true
		} else {
			records[key] = record
		}
		refs := []string{}
		support, partitions := []string{}, []string{}
		for _, field := range record.Fields {
			support = appendUniqueStrings(support, field.SupportRefs...)
			partitions = orderedDisclosurePartitions(append(partitions, field.DisclosurePartitionRefs...))
			if len(field.SourceRefs) == 0 {
				invalid("dangling_source_ref")
				continue
			}
			encoded, err := reportingFieldContent(field)
			if err != nil {
				invalid("invalid_scalar_value")
			}
			for _, ref := range field.SourceRefs {
				if bySource[ref] == nil {
					invalid("dangling_source_ref")
				} else if !ambiguousSources[ref] && err == nil && !bytes.Equal(encoded, bySource[ref]) {
					invalid("invalid_scalar_value")
				}
				if bound[ref] {
					invalid("duplicate_export_model_path")
				}
				bound[ref] = true
				refs = append(refs, ref)
			}
		}
		sort.Strings(refs)
		if !orderedReportingRefs(record.SourceRefs) || !slices.Equal(refs, record.SourceRefs) {
			invalid("dangling_source_ref")
		}
		if !orderedReportingRefs(record.SupportRefs) || !slices.Equal(support, record.SupportRefs) ||
			record.DisclosurePartitionRefs == nil || validateDisclosurePartitionSet(record.DisclosurePartitionRefs) != nil || !slices.Equal(partitions, record.DisclosurePartitionRefs) {
			invalid("invalid_scalar_value")
		}
	}
	// Auxiliary source fields and the parallel envelope projection remain in
	// sections/source_index. Core summary completeness is checked separately against
	// the complete frozen source groups, not by making every field a Core record.
	seenEvents := map[sourceKey]bool{}
	for _, event := range model.TimelineEvents {
		key := sourceKey{"timeline_event", event.RecordID}
		record, exists := records[key]
		if !ambiguousRecords[key] && (!exists || record.RecordType != "timeline_event" || event.RecordID != record.RecordID) {
			invalid("dangling_source_ref")
		} else if exists && !ambiguousRecords[key] && (!sameReportingValue(event.Fields, record.Fields) ||
			!sameReportingValue(event.SourceRefs, record.SourceRefs) || !sameReportingValue(event.SupportRefs, record.SupportRefs) ||
			!sameReportingValue(event.DisclosurePartitionRefs, record.DisclosurePartitionRefs)) {
			invalid("invalid_scalar_value")
		}
		if seenEvents[key] {
			invalid("duplicate_stable_id")
		}
		seenEvents[key] = true
	}
	for key, record := range records {
		if key.family == "timeline_event" && record.RecordType == "timeline_event" && !seenEvents[key] {
			invalid("dangling_source_ref")
		}
	}
	issues.add(validateSnapshotTimelineEvents(model.TimelineEvents))
	return issues.finish()
}

func reportingFieldContent(field ReportingField) ([]byte, error) {
	// Ordinals belong to each containing list; all other field members must be
	// identical. Canonical bytes compare JSON values, not Go decoder types.
	field.FieldOrdinal = 0
	return canonicalJSON(field)
}

func sameReportingValue(left, right any) bool {
	a, err := canonicalJSON(left)
	if err != nil {
		return false
	}
	b, err := canonicalJSON(right)
	return err == nil && bytes.Equal(a, b)
}
