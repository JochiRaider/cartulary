package exportprovider

import (
	"errors"
	"slices"
)

// BoundField carries the owner's explicit identity through materialization.
// A field path is a selector, not a source-record identity encoding.
type BoundField struct {
	FieldFact
	RecordID   string
	RecordType string
}

func (output ProviderOutput) BoundFields() ([]BoundField, error) {
	bindings, err := output.validatedRecordBindings()
	if err != nil {
		return nil, err
	}
	fields := make([]BoundField, 0, len(output.FieldFacts))
	for _, field := range output.FieldFacts {
		record := bindings[field.Path]
		fields = append(fields, BoundField{FieldFact: field, RecordID: record.RecordID, RecordType: record.RecordType})
	}
	return fields, nil
}

func (output ProviderOutput) recordBindings() (map[string]RecordFact, error) {
	fields := map[string]FieldFact{}
	for _, field := range output.FieldFacts {
		fields[field.Path] = field
	}
	bindings := map[string]RecordFact{}
	type identity struct{ family, id string }
	records := map[identity]RecordFact{}
	for _, record := range output.RecordFacts {
		key := identity{record.SourceFamily, record.RecordID}
		if _, exists := records[key]; exists || len(record.FieldPaths) == 0 {
			return nil, errors.New("invalid record fact identity or membership")
		}
		records[key] = record
		partitions, support := map[string]bool{}, map[string]bool{}
		for _, path := range record.FieldPaths {
			field, exists := fields[path]
			if !exists || field.SourceFamily != record.SourceFamily {
				return nil, errors.New("record fact has a dangling or foreign field binding")
			}
			if _, exists := bindings[path]; exists {
				return nil, errors.New("field fact has overlapping record bindings")
			}
			bindings[path] = record
			for _, ref := range field.DisclosurePartitionRefs {
				partitions[ref] = true
			}
			for _, ref := range field.SupportRefs {
				support[ref] = true
			}
		}
		if !sameReferences(record.DisclosurePartitionRefs, partitions) || !sameReferences(record.SupportRefs, support) {
			return nil, errors.New("record fact metadata differs from its fields")
		}
	}
	if len(bindings) != len(fields) {
		return nil, errors.New("field fact has no source record binding")
	}
	seenSubjects := map[string]bool{}
	subjectRecords := map[identity]bool{}
	for _, subject := range output.SubjectFacts {
		if seenSubjects[subject.StableSubjectRef] || len(subject.FieldPaths) == 0 {
			return nil, errors.New("duplicate or empty subject fact")
		}
		seenSubjects[subject.StableSubjectRef] = true
		if err := validateFactPaths(subject.FieldPaths, subject.SourceFamily, subject.SourceRecordID, bindings); err != nil {
			return nil, err
		}
		key := identity{subject.SourceFamily, subject.SourceRecordID}
		record := records[key]
		kind, ref, err := subjectIdentity(record, boundRecordFields(record, fields))
		if err != nil || subjectRecords[key] || ref == "" || subject.SubjectKind != kind || subject.StableSubjectRef != ref || !sameMembers(subject.FieldPaths, record.FieldPaths) || !sameMembers(subject.DisclosurePartitionRefs, record.DisclosurePartitionRefs) {
			return nil, errors.New("subject fact differs from its declared source record")
		}
		subjectRecords[key] = true
	}
	seenEvents := map[string]bool{}
	for _, event := range output.TimelineEventFacts {
		if event.SourceFamily != "timeline_event" || seenEvents[event.TimelineEventID] || len(event.FieldPaths) == 0 {
			return nil, errors.New("duplicate or invalid timeline event fact")
		}
		seenEvents[event.TimelineEventID] = true
		if err := validateFactPaths(event.FieldPaths, event.SourceFamily, event.TimelineEventID, bindings); err != nil {
			return nil, err
		}
		record := records[identity{event.SourceFamily, event.TimelineEventID}]
		if !sameMembers(event.FieldPaths, record.FieldPaths) || !sameMembers(event.DisclosurePartitionRefs, record.DisclosurePartitionRefs) || !sameMembers(event.SupportRefs, record.SupportRefs) {
			return nil, errors.New("timeline event fact differs from its declared source record")
		}
	}
	for key := range records {
		switch key.family {
		case "host", "identity", "party", "entity_mention":
			if !subjectRecords[key] {
				return nil, errors.New("missing subject fact")
			}
		case "timeline_event":
			if !seenEvents[key.id] {
				return nil, errors.New("missing timeline event fact")
			}
		}
	}
	seenSupport := map[[2]string]bool{}
	supportTargets := map[string]string{}
	for _, support := range output.SupportRefFacts {
		field, exists := fields[support.SourcePath]
		key := [2]string{support.SourcePath, support.TargetPath}
		if !exists || !slices.Contains(field.SupportRefs, support.TargetPath) || support.SupportTargetRef != support.TargetPath || seenSupport[key] {
			return nil, errors.New("invalid or duplicate support fact binding")
		}
		if target, exists := supportTargets[support.SupportRefID]; exists && target != support.TargetPath {
			return nil, errors.New("support identity names conflicting targets")
		}
		supportTargets[support.SupportRefID] = support.TargetPath
		seenSupport[key] = true
	}
	seenPartitions := map[[2]string]bool{}
	for _, partition := range output.DisclosurePartitionFacts {
		field, exists := fields[partition.FieldPath]
		record := bindings[partition.FieldPath]
		key := [2]string{partition.FieldPath, partition.PartitionRef}
		if !exists || partition.SourceFamily != record.SourceFamily || partition.SourceRecordID != record.RecordID || !slices.Contains(field.DisclosurePartitionRefs, partition.PartitionRef) || seenPartitions[key] {
			return nil, errors.New("invalid or duplicate partition fact binding")
		}
		seenPartitions[key] = true
	}
	for _, field := range output.FieldFacts {
		for _, ref := range field.SupportRefs {
			if !seenSupport[[2]string{field.Path, ref}] {
				return nil, errors.New("missing support fact binding")
			}
		}
		for _, ref := range field.DisclosurePartitionRefs {
			if !seenPartitions[[2]string{field.Path, ref}] {
				return nil, errors.New("missing partition fact binding")
			}
		}
	}
	return bindings, nil
}

func boundRecordFields(record RecordFact, fields map[string]FieldFact) []BoundField {
	bound := make([]BoundField, 0, len(record.FieldPaths))
	for _, path := range record.FieldPaths {
		bound = append(bound, BoundField{FieldFact: fields[path], RecordID: record.RecordID, RecordType: record.RecordType})
	}
	return bound
}

func sameMembers(values, expected []string) bool {
	set := map[string]bool{}
	for _, value := range expected {
		set[value] = true
	}
	return sameReferences(values, set)
}

func validateFactPaths(paths []string, family, id string, bindings map[string]RecordFact) error {
	seen := map[string]bool{}
	for _, path := range paths {
		record, exists := bindings[path]
		if !exists || record.SourceFamily != family || record.RecordID != id || seen[path] {
			return errors.New("derived fact has a dangling, foreign or repeated field binding")
		}
		seen[path] = true
	}
	return nil
}

func sameReferences(values []string, want map[string]bool) bool {
	if len(values) != len(want) {
		return false
	}
	seen := map[string]bool{}
	for _, value := range values {
		if value == "" || !want[value] || seen[value] {
			return false
		}
		seen[value] = true
	}
	return true
}
