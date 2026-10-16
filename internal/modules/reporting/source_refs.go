package reporting

import (
	"slices"
	"sort"
	"strings"
)

// ReportingSourceRef is the public provenance address. Private owner selectors
// exist only during materialization; they never become retained identities.
type ReportingSourceRef struct {
	SourceRefID             string   `json:"source_ref_id"`
	SourceFamily            string   `json:"source_family"`
	SourceRecordID          string   `json:"source_record_id"`
	SourcePath              *string  `json:"source_path"`
	SourceSnapshotID        string   `json:"source_snapshot_id"`
	SourceSummary           *string  `json:"source_summary"`
	DisclosurePartitionRefs []string `json:"disclosure_partition_refs"`
}

type sourceOwner struct{ family, id string }

func supportSourceOwner(ref ReportingSupportRef) (sourceOwner, bool) {
	_, target, ok := strings.Cut(ref.SupportTargetRef, ":")
	if !ok {
		return sourceOwner{}, false
	}
	switch ref.SupportKind {
	case "source_record":
		family, id, _ := strings.Cut(target, ":")
		return sourceOwner{family, id}, true
	case "evidence_item":
		return sourceOwner{"evidence", target}, true
	case "artifact_record":
		return sourceOwner{"artifact", target}, true
	case "timeline_event":
		return sourceOwner{"timeline_event", target}, true
	case "relationship":
		return sourceOwner{"record_link", target}, true
	default:
		return sourceOwner{}, false
	}
}

func sourceReferenceID(snapshot string, owner sourceOwner, path *string) (string, error) {
	return generatedReportingID("cartulary.reporting_source_ref_id.v1", "src_", map[string]any{
		"snapshot_id": snapshot, "source_family": owner.family, "source_record_id": owner.id, "source_path": path,
	})
}

func snapshotSourceIndex(model SnapshotContent) map[string]ReportingSourceRef {
	index := make(map[string]ReportingSourceRef, len(model.SourceIndex))
	for _, ref := range model.SourceIndex {
		index[ref.SourceRefID] = ref
	}
	return index
}

// This one-way projection runs only on freshly collected owner facts. Frozen
// readers validate exact bindings; they never upgrade selector-based history.
func materializeSnapshotSourceIndex(model *SnapshotContent, owners map[string]sourceOwner, groups []reportingSourceGroup) error {
	paths, err := redactionFieldModelPaths(*model)
	if err != nil {
		return err
	}
	fields := reportingFieldSources(*model)
	replacements := map[string]string{}
	index := map[string]ReportingSourceRef{}
	records := map[sourceOwner]ReportingSourceRef{}
	add := func(owner sourceOwner, path *string, partitions []string) (ReportingSourceRef, error) {
		if !validSupportIdentifier(owner.family) || owner.family == "post" || !validSupportIdentifier(owner.id) {
			return ReportingSourceRef{}, newRenderValidationError("export_model_invalid", "dangling_source_ref")
		}
		id, err := sourceReferenceID(model.SnapshotID, owner, path)
		if err != nil {
			return ReportingSourceRef{}, err
		}
		ref := ReportingSourceRef{SourceRefID: id, SourceFamily: owner.family, SourceRecordID: owner.id,
			SourcePath: path, SourceSnapshotID: model.SnapshotID, DisclosurePartitionRefs: orderedDisclosurePartitions(partitions)}
		index[id] = ref
		return ref, nil
	}
	incidentPartitions := []string{}
	hasIncidentFields := false
	for selector, path := range paths {
		owner, exists := owners[selector]
		if !exists {
			return newRenderValidationError("export_model_invalid", "dangling_source_ref")
		}
		ref, err := add(owner, &path, fields[selector].Field.DisclosurePartitionRefs)
		if err != nil {
			return err
		}
		replacements[selector] = ref.SourceRefID
		if owner.family == "incident_metadata" {
			hasIncidentFields = true
			incidentPartitions = append(incidentPartitions, ref.DisclosurePartitionRefs...)
		}
	}
	for _, record := range groups {
		owner := sourceOwner{record.SourceRecordRef.SourceFamily, record.RecordID}
		ref, err := add(owner, nil, record.DisclosurePartitionRefs)
		if err != nil {
			return err
		}
		records[owner] = ref
	}
	// The incident itself is a source-backed support target even when no
	// incident field was requested. Its owned disclosure default is internal.
	if !hasIncidentFields {
		incidentPartitions = []string{"internal_only"}
	}
	incidentOwner := sourceOwner{"incident_metadata", model.IncidentID}
	incident, err := add(incidentOwner, nil, incidentPartitions)
	if err != nil {
		return err
	}
	records[incidentOwner] = incident
	mapRefs := func(refs []string) ([]string, error) {
		out := make([]string, len(refs))
		for i, ref := range refs {
			id, exists := replacements[ref]
			if !exists {
				return nil, newRenderValidationError("export_model_invalid", "dangling_source_ref")
			}
			out[i] = id
		}
		sort.Strings(out)
		return out, nil
	}
	mapFields := func(fields []ReportingField) error {
		for i := range fields {
			refs, err := mapRefs(fields[i].SourceRefs)
			if err != nil {
				return err
			}
			fields[i].SourceRefs = refs
		}
		return nil
	}
	var mapBlocks func([]ReportingBlock) error
	mapBlocks = func(blocks []ReportingBlock) error {
		for i := range blocks {
			refs, err := mapRefs(blocks[i].SourceRefs)
			if err != nil {
				return err
			}
			blocks[i].SourceRefs = refs
			if err := mapFields(blocks[i].Fields); err != nil {
				return err
			}
			if err := mapBlocks(blocks[i].Children); err != nil {
				return err
			}
		}
		return nil
	}
	for i := range model.Sections {
		refs, err := mapRefs(model.Sections[i].SourceRefs)
		if err != nil {
			return err
		}
		model.Sections[i].SourceRefs = refs
		if err := mapBlocks(model.Sections[i].Blocks); err != nil {
			return err
		}
	}
	bindRecord := func(ref *SourceRecordRef) error {
		source, exists := records[sourceOwner{ref.SourceFamily, ref.SourceRecordID}]
		if !exists {
			return newRenderValidationError("export_model_invalid", "dangling_source_ref")
		}
		id := source.SourceRefID
		ref.SourceRefID = &id
		return nil
	}
	for i := range model.Records {
		record := &model.Records[i]
		if err := bindRecord(&record.SourceRecordRef); err != nil {
			return err
		}
		refs, err := mapRefs(record.SourceRefs)
		if err != nil {
			return err
		}
		record.SourceRefs = refs
		if err := mapFields(record.Fields); err != nil {
			return err
		}
	}
	for i := range model.TimelineEvents {
		refs, err := mapRefs(model.TimelineEvents[i].SourceRefs)
		if err != nil {
			return err
		}
		model.TimelineEvents[i].SourceRefs = refs
		if err := mapFields(model.TimelineEvents[i].Fields); err != nil {
			return err
		}
	}
	for i := range model.Subjects {
		refs, err := mapRefs(model.Subjects[i].SourceRefs)
		if err != nil {
			return err
		}
		model.Subjects[i].SourceRefs = refs
	}
	for i := range model.Relationships {
		rel := &model.Relationships[i]
		refs, err := mapRefs(rel.SourceRefs)
		if err != nil {
			return err
		}
		rel.SourceRefs = refs
		if err := bindRecord(&rel.SrcRecordRef.SourceRecordRef); err != nil {
			return err
		}
		if err := bindRecord(&rel.DstRecordRef.SourceRecordRef); err != nil {
			return err
		}
	}
	for i := range model.SupportIndex {
		support := &model.SupportIndex[i]
		owner, sourceBacked := supportSourceOwner(*support)
		if !sourceBacked {
			continue
		}
		ref, exists := records[owner]
		if !exists {
			return invalidSupportRef()
		}
		id := ref.SourceRefID
		support.SourceRefID, support.DisclosurePartitionRefs = &id, append([]string{}, ref.DisclosurePartitionRefs...)
	}
	model.SourceIndex = make([]ReportingSourceRef, 0, len(index))
	for _, ref := range index {
		model.SourceIndex = append(model.SourceIndex, ref)
	}
	sort.Slice(model.SourceIndex, func(i, j int) bool { return model.SourceIndex[i].SourceRefID < model.SourceIndex[j].SourceRefID })
	return nil
}

func validateSnapshotSourceReferences(model SnapshotContent) error {
	issues := renderIssueCollector{stage: "export_model_materialization"}
	invalid := func(reason string) { issues.add(newRenderValidationError("export_model_invalid", reason)) }
	if model.SourceIndex == nil {
		invalid("dangling_source_ref")
	}
	index := snapshotSourceIndex(model)
	paths, err := redactionFieldModelPaths(model)
	issues.add(err)
	fields := reportingFieldSources(model)
	groups, groupErr := snapshotSourceGroups(model)
	issues.add(groupErr)
	records := map[sourceOwner][]string{{"incident_metadata", model.IncidentID}: nil}
	for _, record := range groups {
		records[sourceOwner{record.SourceRecordRef.SourceFamily, record.RecordID}] = record.DisclosurePartitionRefs
	}
	incidentPartitions := []string{}
	hasIncidentFields := false
	for _, ref := range model.SourceIndex {
		if ref.SourceFamily == "incident_metadata" && ref.SourcePath != nil {
			hasIncidentFields = true
			incidentPartitions = append(incidentPartitions, ref.DisclosurePartitionRefs...)
		}
	}
	if !hasIncidentFields {
		incidentPartitions = []string{"internal_only"}
	}
	records[sourceOwner{"incident_metadata", model.IncidentID}] = orderedDisclosurePartitions(incidentPartitions)
	previous := ""
	anchors := map[sourceOwner]int{}
	for _, ref := range model.SourceIndex {
		owner := sourceOwner{ref.SourceFamily, ref.SourceRecordID}
		expected, err := sourceReferenceID(ref.SourceSnapshotID, owner, ref.SourcePath)
		if err != nil || ref.SourceRefID != expected {
			invalid("invalid_generated_identifier")
		}
		if ref.SourceRefID <= previous {
			invalid("duplicate_stable_id")
		}
		previous = ref.SourceRefID
		if ref.SourceSnapshotID != model.SnapshotID {
			invalid("snapshot_ref_mismatch")
		}
		partitions, exists := records[owner]
		if !exists || !validSupportIdentifier(owner.family) || owner.family == "post" || !validSupportIdentifier(owner.id) || ref.SourceSummary != nil {
			invalid("dangling_source_ref")
		}
		if ref.SourcePath != nil {
			if !validExactExportModelPath(*ref.SourcePath) || paths[ref.SourceRefID] != *ref.SourcePath {
				invalid("dangling_source_ref")
			}
			partitions = fields[ref.SourceRefID].Field.DisclosurePartitionRefs
		} else {
			anchors[owner]++
		}
		if ref.DisclosurePartitionRefs == nil || validateDisclosurePartitionSet(ref.DisclosurePartitionRefs) != nil || !slices.Equal(partitions, ref.DisclosurePartitionRefs) {
			invalid("invalid_scalar_value")
		}
	}
	for owner := range records {
		if anchors[owner] != 1 {
			invalid("dangling_source_ref")
		}
	}
	check := func(id string, owner *sourceOwner, record bool) {
		ref, exists := index[id]
		if !exists || (ref.SourcePath == nil) != record || owner != nil && (ref.SourceFamily != owner.family || ref.SourceRecordID != owner.id) {
			invalid("dangling_source_ref")
		}
	}
	for id := range paths {
		check(id, nil, false)
	}
	bindRecord := func(ref SourceRecordRef) {
		if ref.SourceRefID == nil {
			invalid("dangling_source_ref")
			return
		}
		check(*ref.SourceRefID, &sourceOwner{ref.SourceFamily, ref.SourceRecordID}, true)
	}
	for _, record := range model.Records {
		bindRecord(record.SourceRecordRef)
		owner := sourceOwner{record.SourceRecordRef.SourceFamily, record.RecordID}
		for _, ref := range record.SourceRefs {
			check(ref, &owner, false)
		}
	}
	for _, rel := range model.Relationships {
		bindRecord(rel.SrcRecordRef.SourceRecordRef)
		bindRecord(rel.DstRecordRef.SourceRecordRef)
	}
	for _, support := range model.SupportIndex {
		owner, sourceBacked := supportSourceOwner(support)
		if sourceBacked {
			if support.SourceRefID == nil {
				issues.add(invalidSupportRef())
				continue
			}
			ref, exists := index[*support.SourceRefID]
			if !exists || ref.SourcePath != nil || ref.SourceFamily != owner.family || ref.SourceRecordID != owner.id {
				issues.add(invalidSupportRef())
			}
			if !slices.Equal(support.DisclosurePartitionRefs, records[owner]) {
				issues.add(invalidSupportRef())
			}
		} else if support.SourceRefID != nil {
			issues.add(invalidSupportRef())
		}
	}
	return issues.finish()
}
