package reporting

import (
	"sort"
	"strings"
)

// A source group is an internal view of explicitly bound fields. Mentions,
// links, tags and envelope metadata are sources, not additional Core records.
// The view is reconstructed from the frozen source index, never a path prefix.
type reportingSourceGroup struct {
	RecordID                string
	RecordType              string
	DeletedState            string
	SourceRecordRef         SourceRecordRef
	DisplayName             *string
	Fields                  []ReportingField
	SourceRefs              []string
	SupportRefs             []string
	DisclosurePartitionRefs []string
}

func coreTypeForSourceFamily(family string) string {
	switch family {
	case "timeline_event", "host", "identity", "party", "indicator", "artifact", "task_request", "decision", "evidence", "assessment":
		return family
	case "note", "finding_hypothesis", "comm_log", "handoff", "status_review", "lesson":
		return "artifact"
	default:
		return ""
	}
}

func validCoreRecordType(value string) bool {
	return value != "" && coreTypeForSourceFamily(value) == value
}

func (group *reportingSourceGroup) normalize() error {
	if group.SourceRecordRef.SourceFamily == "entity_mention" {
		if _, err := reportingMentionSource(*group); err != nil {
			return err
		}
	}
	sort.Slice(group.Fields, func(i, j int) bool { return group.Fields[i].FieldKey < group.Fields[j].FieldKey })
	for i := range group.Fields {
		if i > 0 && group.Fields[i-1].FieldKey == group.Fields[i].FieldKey {
			return newRenderValidationError("export_model_invalid", "duplicate_stable_id")
		}
		group.Fields[i].FieldOrdinal = i + 1
	}
	sort.Strings(group.SourceRefs)
	sort.Strings(group.SupportRefs)
	group.DisclosurePartitionRefs = orderedDisclosurePartitions(group.DisclosurePartitionRefs)
	family := group.SourceRecordRef.SourceFamily
	expected := coreTypeForSourceFamily(family)
	if family == "record_envelope" {
		for _, field := range group.Fields {
			if field.FieldKey == "record_type" {
				expected, _ = field.Value.(string)
			}
		}
		if !validCoreRecordType(expected) {
			return newRenderValidationError("export_model_invalid", "invalid_type")
		}
	} else if expected == "" {
		switch family {
		case "entity_mention", "record_link", "record_tag":
			expected = family
		default:
			return newRenderValidationError("export_model_invalid", "invalid_type")
		}
	}
	if group.RecordType != "" && group.RecordType != expected {
		return newRenderValidationError("export_model_invalid", "invalid_type")
	}
	group.RecordType = expected
	state, err := sourceGroupLifecycle(family, group.Fields)
	if err != nil {
		return err
	}
	group.DeletedState = state
	group.DisplayName = sourceGroupDisplayName(family, group.Fields)
	return nil
}

// Display selection uses only the supplied fields. The same rule applies to
// admitted source fields and to their independently selected retained values.
func sourceGroupDisplayName(family string, fields []ReportingField) *string {
	keys := []string{"display_name", "title", "summary", "activity_synopsis_text", "statement"}
	switch family {
	case "host", "identity", "party":
		keys = []string{"display_name"}
	case "entity_mention":
		keys = []string{"raw_text"}
	}
	for _, key := range keys {
		for _, field := range fields {
			if value, ok := field.Value.(string); field.FieldKey == key && ok && strings.TrimSpace(value) != "" {
				return &value
			}
		}
	}
	return nil
}

func (group reportingSourceGroup) coreSummary() ReportingRecordSummary {
	return ReportingRecordSummary{
		SchemaID: "cartulary.reporting_record_summary.v1", RecordID: group.RecordID, RecordType: group.RecordType,
		SourceRecordRef: group.SourceRecordRef, DisplayName: group.DisplayName, DeletedState: group.DeletedState,
		Fields: group.Fields, SourceRefs: group.SourceRefs, SupportRefs: group.SupportRefs,
		DisclosurePartitionRefs: group.DisclosurePartitionRefs,
	}
}

// One Core envelope has one summary. A selected domain projection supplies its
// display fields; otherwise the explicit envelope projection supplies metadata.
// Their fields remain separate in sections/source_index, avoiding collisions
// between domain and envelope row versions or similarly named timestamps.
func coreRecordSummaries(groups []reportingSourceGroup) ([]ReportingRecordSummary, error) {
	byID := map[string]reportingSourceGroup{}
	for _, group := range groups {
		family := group.SourceRecordRef.SourceFamily
		if family != "record_envelope" && coreTypeForSourceFamily(family) == "" {
			continue
		}
		previous, exists := byID[group.RecordID]
		if exists {
			if previous.RecordType != group.RecordType || previous.SourceRecordRef.SourceFamily != "record_envelope" && family != "record_envelope" {
				return nil, newRenderValidationError("export_model_invalid", "duplicate_stable_id")
			}
			if family == "record_envelope" {
				continue
			}
		}
		byID[group.RecordID] = group
	}
	result := make([]ReportingRecordSummary, 0, len(byID))
	for _, group := range byID {
		if group.SourceRecordRef.SourceFamily == "record_envelope" && (group.RecordType == "timeline_event" || group.RecordType == "decision") {
			return nil, newRenderValidationError("export_model_invalid", "dangling_source_ref")
		}
		result = append(result, group.coreSummary())
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].RecordType == result[j].RecordType {
			return result[i].RecordID < result[j].RecordID
		}
		return result[i].RecordType < result[j].RecordType
	})
	return result, nil
}

func validateCoreRecordProjection(model SnapshotContent) error {
	groups, err := snapshotSourceGroups(model)
	if err != nil {
		return err
	}
	expected, err := coreRecordSummaries(groups)
	if err != nil {
		return err
	}
	if len(expected) != len(model.Records) {
		return newRenderValidationError("export_model_invalid", "dangling_source_ref")
	}
	for i, record := range expected {
		if record.RecordID != model.Records[i].RecordID || record.SourceRecordRef.SourceFamily != model.Records[i].SourceRecordRef.SourceFamily {
			return newRenderValidationError("export_model_invalid", "dangling_source_ref")
		}
	}
	if !sameReportingValue(expected, model.Records) {
		return newRenderValidationError("export_model_invalid", "invalid_scalar_value")
	}
	return nil
}

func snapshotSourceGroups(model SnapshotContent) ([]reportingSourceGroup, error) {
	invalid := newRenderValidationError("export_model_invalid", "dangling_source_ref")
	index := snapshotSourceIndex(model)
	groups := map[sourceOwner]*reportingSourceGroup{}
	anchors := map[sourceOwner]string{}
	for _, ref := range model.SourceIndex {
		if ref.SourcePath == nil {
			anchors[sourceOwner{ref.SourceFamily, ref.SourceRecordID}] = ref.SourceRefID
		}
	}
	var collect func([]ReportingBlock) error
	collect = func(blocks []ReportingBlock) error {
		for _, block := range blocks {
			for _, field := range block.Fields {
				if len(field.SourceRefs) == 0 {
					return invalid
				}
				var owner sourceOwner
				for i, id := range field.SourceRefs {
					ref, exists := index[id]
					if !exists || ref.SourcePath == nil {
						return invalid
					}
					key := sourceOwner{ref.SourceFamily, ref.SourceRecordID}
					if i != 0 && key != owner {
						return invalid
					}
					owner = key
				}
				if owner.family == "incident_metadata" {
					continue
				}
				group := groups[owner]
				if group == nil {
					anchor, exists := anchors[owner]
					if !exists {
						return invalid
					}
					group = &reportingSourceGroup{RecordID: owner.id,
						SourceRecordRef: SourceRecordRef{SchemaID: "source_record_ref.v1", SourceFamily: owner.family, SourceRecordID: owner.id, SourceSnapshotID: index[anchor].SourceSnapshotID, SourceRefID: &anchor},
						Fields:          []ReportingField{}, SourceRefs: []string{}, SupportRefs: []string{}, DisclosurePartitionRefs: []string{}}
					groups[owner] = group
				}
				group.Fields = append(group.Fields, field)
				group.SourceRefs = appendUniqueStrings(group.SourceRefs, field.SourceRefs...)
				group.SupportRefs = appendUniqueStrings(group.SupportRefs, field.SupportRefs...)
				group.DisclosurePartitionRefs = append(group.DisclosurePartitionRefs, field.DisclosurePartitionRefs...)
			}
			if err := collect(block.Children); err != nil {
				return err
			}
		}
		return nil
	}
	for _, section := range model.Sections {
		if err := collect(section.Blocks); err != nil {
			return nil, err
		}
	}
	result := make([]reportingSourceGroup, 0, len(groups))
	for _, group := range groups {
		if err := group.normalize(); err != nil {
			return nil, err
		}
		result = append(result, *group)
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].SourceRecordRef.SourceFamily == result[j].SourceRecordRef.SourceFamily {
			return result[i].RecordID < result[j].RecordID
		}
		return result[i].SourceRecordRef.SourceFamily < result[j].SourceRecordRef.SourceFamily
	})
	return result, nil
}
