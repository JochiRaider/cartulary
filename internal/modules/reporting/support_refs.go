package reporting

import (
	"sort"
	"strings"
	"unicode/utf8"
)

func invalidSupportRef() error {
	return newRenderValidationError("export_model_invalid", "invalid_support_ref")
}

type supportTarget struct{ kind, ref string }

// Source-owner support paths are logical record selectors, never storage paths.
// Reject malformed/unadmitted selectors without normalizing them or inventing
// a record identity from a hash. Frozen-model admission checks reachability.
func resolveSupportSelector(ref, incidentID string) (supportTarget, error) {
	parts := strings.Split(ref, "/")
	if len(parts) != 3 || parts[0] != "" || !validSupportIdentifier(parts[2]) {
		return supportTarget{}, invalidSupportRef()
	}
	family := sourceFamilyForPath(ref)
	if family == "" {
		return supportTarget{}, invalidSupportRef()
	}
	switch parts[1] {
	case "incident":
		if !validSupportIdentifier(incidentID) {
			return supportTarget{}, invalidSupportRef()
		}
		return supportTarget{"source_record", "record:incident_metadata:" + incidentID}, nil
	case "evidence":
		return supportTarget{"evidence_item", "evidence:" + parts[2]}, nil
	case "timeline":
		return supportTarget{"timeline_event", "timeline:" + parts[2]}, nil
	case "relationships":
		return supportTarget{"relationship", "relationship:" + parts[2]}, nil
	default:
		return supportTarget{"source_record", "record:" + family + ":" + parts[2]}, nil
	}
}

func supportReferenceID(snapshotID string, target supportTarget) (string, error) {
	return generatedReportingID("cartulary.reporting_support_ref_id.v1", "sup_", map[string]any{
		"snapshot_id": snapshotID, "support_kind": target.kind, "support_target_ref": target.ref,
	})
}

// Source selectors are confined to the owner port. Retained objects refer to
// the snapshot's support index by identity, never by an executable/source path.
func materializeSupportReferences(selectors []string, incidentID, snapshotID string) ([]string, map[string]supportTarget, error) {
	targets := map[string]supportTarget{}
	for _, selector := range selectors {
		target, err := resolveSupportSelector(selector, incidentID)
		if err != nil {
			return nil, nil, err
		}
		id, err := supportReferenceID(snapshotID, target)
		if err != nil {
			return nil, nil, err
		}
		targets[id] = target
	}
	ids := make([]string, 0, len(targets))
	for id := range targets {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids, targets, nil
}

// Validate against the frozen inventory only. A current database record or a
// syntactically valid target cannot supply missing historical provenance.
func validateSnapshotSupportReferences(model SnapshotContent) error {
	issues := renderIssueCollector{stage: "export_model_materialization"}
	available := map[supportTarget]int{
		{"source_record", "record:incident_metadata:" + model.IncidentID}: 1,
	}
	for _, source := range model.SourceIndex {
		if source.SourcePath == nil && source.SourceFamily != "incident_metadata" {
			available[supportTarget{"source_record", "record:" + source.SourceFamily + ":" + source.SourceRecordID}]++
		}
	}
	for _, record := range model.Records {
		if record.SourceRecordRef.SourceFamily == "evidence" && record.RecordType == "evidence" {
			available[supportTarget{"evidence_item", "evidence:" + record.RecordID}]++
		}
		if record.RecordType == "artifact" {
			available[supportTarget{"artifact_record", "artifact:" + record.RecordID}]++
		}
	}
	for _, event := range model.TimelineEvents {
		available[supportTarget{"timeline_event", "timeline:" + event.RecordID}]++
	}
	for _, relationship := range model.Relationships {
		available[supportTarget{"relationship", "relationship:" + relationship.RelationshipID}]++
	}
	for _, diagram := range model.Diagrams {
		available[supportTarget{"diagram_selection", "diagram:" + diagram.DiagramID}]++
	}
	index := map[string]bool{}
	previous := ""
	for _, ref := range model.SupportIndex {
		target := supportTarget{ref.SupportKind, ref.SupportTargetRef}
		expected, err := supportReferenceID(model.SnapshotID, target)
		if err != nil || ref.SchemaID != "cartulary.reporting_support_ref.v1" || ref.SupportRefID != expected ||
			ref.SupportRefID <= previous || ref.SourceSnapshotID != model.SnapshotID ||
			!validateSupportTarget(target.kind, target.ref) || available[target] != 1 {
			issues.add(invalidSupportRef())
		}
		previous = ref.SupportRefID
		index[ref.SupportRefID] = false
	}
	check := func(refs []string) {
		if !orderedReportingRefs(refs) {
			issues.add(invalidSupportRef())
		}
		for _, id := range refs {
			if _, exists := index[id]; !exists {
				issues.add(invalidSupportRef())
				continue
			}
			index[id] = true
		}
	}
	checkFields := func(fields []ReportingField) {
		for _, field := range fields {
			check(field.SupportRefs)
		}
	}
	var checkBlocks func([]ReportingBlock)
	checkBlocks = func(blocks []ReportingBlock) {
		for _, block := range blocks {
			check(block.SupportRefs)
			checkFields(block.Fields)
			checkBlocks(block.Children)
		}
	}
	for _, section := range model.Sections {
		check(section.SupportRefs)
		checkBlocks(section.Blocks)
	}
	for _, record := range model.Records {
		check(record.SupportRefs)
		checkFields(record.Fields)
	}
	for _, event := range model.TimelineEvents {
		check(event.SupportRefs)
		checkFields(event.Fields)
	}
	for _, relationship := range model.Relationships {
		check(relationship.SupportRefs)
	}
	for _, used := range index {
		if !used {
			issues.add(invalidSupportRef())
		}
	}
	return issues.finish()
}

func validSupportIdentifier(value string) bool {
	return validReportingText(value, false) && utf8.RuneCountInString(value) >= 1 &&
		utf8.RuneCountInString(value) <= 128 && strings.TrimSpace(value) == value &&
		!strings.ContainsAny(value, "/\\#")
}

// This validates retained output as well as newly materialized output. It is
// not a fallback reader for old unknown-target hashes or malformed target kinds.
func validateSupportTarget(kind, target string) bool {
	prefix, id, ok := strings.Cut(target, ":")
	if !ok {
		return false
	}
	switch kind {
	case "source_record":
		family, recordID, hasID := strings.Cut(id, ":")
		return prefix == "record" && hasID && validSupportIdentifier(family) && family != "post" && family != "unknown" && validSupportIdentifier(recordID)
	case "evidence_item":
		return prefix == "evidence" && validSupportIdentifier(id)
	case "artifact_record":
		return prefix == "artifact" && validSupportIdentifier(id)
	case "timeline_event":
		return prefix == "timeline" && validSupportIdentifier(id)
	case "relationship":
		return prefix == "relationship" && validSupportIdentifier(id)
	case "diagram_selection":
		return prefix == "diagram" && validSupportIdentifier(id)
	default:
		return false
	}
}
