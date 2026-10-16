package exportprovider

import (
	"errors"
	"sort"
	"strings"
)

// NewProviderOutput constructs a complete contribution from explicit owner
// bindings. Selectors are opaque; changing their spelling cannot change a
// record's identity. A failed construction returns no partial contribution.
func NewProviderOutput(providerKey string, fields []BoundField, timeline []TimelineSortFact) (ProviderOutput, error) {
	fields = append([]BoundField(nil), fields...)
	sort.Slice(fields, func(i, j int) bool { return fields[i].Path < fields[j].Path })
	output := ProviderOutput{
		SourceFacts: EmptySourceFacts(),
		SchemaID:    ProviderOutputSchemaID, ProviderKey: providerKey,
		FieldFacts: []FieldFact{}, RecordFacts: []RecordFact{}, TimelineEventFacts: []TimelineEventFact{},
		SubjectFacts: []SubjectFact{}, SupportRefFacts: []SupportRefFact{}, DisclosurePartitionFacts: []DisclosurePartitionFact{},
	}
	sortFacts := map[string]TimelineSortFact{}
	for _, fact := range timeline {
		if err := fact.Validate(); err != nil {
			return ProviderOutput{}, err
		}
		if _, exists := sortFacts[fact.RecordID]; exists {
			return ProviderOutput{}, errors.New("duplicate timeline sort fact")
		}
		sortFacts[fact.RecordID] = fact.Clone()
	}
	type identity struct{ family, id string }
	records := map[identity]RecordFact{}
	fieldsByRecord := map[identity][]BoundField{}
	for _, field := range fields {
		if field.RecordID == "" || field.RecordType == "" {
			return ProviderOutput{}, errors.New("owner field omits record identity or type")
		}
		output.FieldFacts = append(output.FieldFacts, field.FieldFact)
		key := identity{field.SourceFamily, field.RecordID}
		fieldsByRecord[key] = append(fieldsByRecord[key], field)
		record, exists := records[key]
		if exists && record.RecordType != field.RecordType {
			return ProviderOutput{}, errors.New("owner fields declare conflicting record types")
		}
		if !exists {
			record = RecordFact{SchemaID: RecordFactSchemaID, SourceFamily: field.SourceFamily, RecordID: field.RecordID, RecordType: field.RecordType}
		}
		record.FieldPaths = append(record.FieldPaths, field.Path)
		record.DisclosurePartitionRefs = appendUnique(record.DisclosurePartitionRefs, field.DisclosurePartitionRefs...)
		record.SupportRefs = appendUnique(record.SupportRefs, field.SupportRefs...)
		records[key] = record
	}
	for _, record := range records {
		sort.Strings(record.DisclosurePartitionRefs)
		sort.Strings(record.SupportRefs)
		output.RecordFacts = append(output.RecordFacts, record)
	}
	sort.Slice(output.RecordFacts, func(i, j int) bool {
		a, b := output.RecordFacts[i], output.RecordFacts[j]
		if a.SourceFamily == b.SourceFamily {
			return a.RecordID < b.RecordID
		}
		return a.SourceFamily < b.SourceFamily
	})
	for _, record := range output.RecordFacts {
		if record.SourceFamily == "timeline_event" {
			fact, exists := sortFacts[record.RecordID]
			if !exists {
				return ProviderOutput{}, errors.New("missing timeline sort fact")
			}
			delete(sortFacts, record.RecordID)
			output.TimelineEventFacts = append(output.TimelineEventFacts, TimelineEventFact{
				Sort: fact, SchemaID: TimelineEventFactSchemaID, TimelineEventID: record.RecordID, SourceFamily: record.SourceFamily,
				FieldPaths: CloneStrings(record.FieldPaths), DisclosurePartitionRefs: CloneStrings(record.DisclosurePartitionRefs), SupportRefs: CloneStrings(record.SupportRefs),
			})
		}
		kind, ref, err := subjectIdentity(record, fieldsByRecord[identity{record.SourceFamily, record.RecordID}])
		if err != nil {
			return ProviderOutput{}, err
		}
		if ref != "" {
			output.SubjectFacts = append(output.SubjectFacts, SubjectFact{
				SchemaID: SubjectFactSchemaID, StableSubjectRef: ref, SubjectKind: kind,
				SourceFamily: record.SourceFamily, SourceRecordID: record.RecordID,
				FieldPaths: CloneStrings(record.FieldPaths), DisclosurePartitionRefs: CloneStrings(record.DisclosurePartitionRefs),
			})
		}
	}
	if len(sortFacts) != 0 {
		return ProviderOutput{}, errors.New("orphaned timeline sort fact")
	}
	sort.Slice(output.SubjectFacts, func(i, j int) bool {
		return output.SubjectFacts[i].StableSubjectRef < output.SubjectFacts[j].StableSubjectRef
	})
	for _, field := range fields {
		for _, ref := range appendUnique(nil, field.SupportRefs...) {
			output.SupportRefFacts = append(output.SupportRefFacts, SupportRefFact{
				SchemaID: SupportRefFactSchemaID, SupportRefID: supportRefID(ref), SourcePath: field.Path, TargetPath: ref, SupportTargetRef: ref,
			})
		}
		for _, ref := range appendUnique(nil, field.DisclosurePartitionRefs...) {
			output.DisclosurePartitionFacts = append(output.DisclosurePartitionFacts, DisclosurePartitionFact{
				SchemaID: DisclosurePartitionFactSchemaID, PartitionRef: ref, SourceFamily: field.SourceFamily, SourceRecordID: field.RecordID, FieldPath: field.Path,
			})
		}
	}
	sort.Slice(output.SupportRefFacts, func(i, j int) bool {
		a, b := output.SupportRefFacts[i], output.SupportRefFacts[j]
		if a.SourcePath == b.SourcePath {
			return a.TargetPath < b.TargetPath
		}
		return a.SourcePath < b.SourcePath
	})
	sort.Slice(output.DisclosurePartitionFacts, func(i, j int) bool {
		a, b := output.DisclosurePartitionFacts[i], output.DisclosurePartitionFacts[j]
		if a.PartitionRef == b.PartitionRef {
			return a.FieldPath < b.FieldPath
		}
		return a.PartitionRef < b.PartitionRef
	})
	if err := output.Validate(); err != nil {
		return ProviderOutput{}, err
	}
	return output, nil
}

func subjectIdentity(record RecordFact, fields []BoundField) (string, string, error) {
	switch record.SourceFamily {
	case "host", "identity", "party":
		return record.SourceFamily, "subject:" + record.SourceFamily + ":record:" + record.RecordID, nil
	case "entity_mention":
		kind := ""
		for _, field := range fields {
			if field.SourceFamily != record.SourceFamily || field.RecordID != record.RecordID {
				continue
			}
			value, ok := field.Value.(map[string]any)
			if !ok {
				return "", "", errors.New("mention has no typed subject family")
			}
			family, ok := value["entity_type"].(string)
			if !ok || (family != "host" && family != "identity") || (kind != "" && kind != family) {
				return "", "", errors.New("mention has an invalid or conflicting subject family")
			}
			kind = family
		}
		return kind, "subject:" + kind + ":mention:" + record.RecordID, nil
	default:
		return "", "", nil
	}
}

func supportRefID(ref string) string {
	trimmed := strings.Trim(ref, "/")
	if trimmed == "" {
		return "support-root"
	}
	return "support-" + strings.NewReplacer("/", "-", ":", "-", " ", "-").Replace(trimmed)
}

func appendUnique(values []string, additions ...string) []string {
	seen := map[string]bool{}
	for _, value := range values {
		seen[value] = true
	}
	for _, value := range additions {
		if !seen[value] {
			seen[value] = true
			values = append(values, value)
		}
	}
	return values
}
