package reporting

import (
	"context"
	"sort"
)

// The public case-content value has one schema before and after redaction.
// Start with identity only: no source collection can survive merely because a
// new projection forgot to overwrite an unredacted copy.
func retainedExportIdentity(source ExportModel) (ExportModel, error) {
	return copyReportingModel(ExportModel{
		SchemaID: ExportModelSchemaID, ExportModelID: source.ExportModelID, SnapshotModelID: source.SnapshotModelID,
		RenderIdentity: source.RenderIdentity, ExportModelCreatedAt: source.ExportModelCreatedAt,
		ReleaseScope: source.ReleaseScope, RecipientPartitionRefs: cloneStrings(source.RecipientPartitionRefs),
		PresentationTexts: []ReportingPresentationText{},
		SnapshotContent: SnapshotContent{
			ReferencePacks: source.ReferencePacks, IncidentID: source.IncidentID, SnapshotID: source.SnapshotID,
			SnapshotAt: source.SnapshotAt, SourceChangeSetHighWatermark: source.SourceChangeSetHighWatermark,
			SnapshotBoundaryKind: source.SnapshotBoundaryKind, DerivationVersion: source.DerivationVersion,
			ExportModelGeneratorID: source.ExportModelGeneratorID, ExportModelGeneratorVersion: source.ExportModelGeneratorVersion,
			Sections: []ReportingSection{}, Records: []ReportingRecordSummary{}, Relationships: []ReportingRelationshipSummary{},
			TimelineEvents: []ReportingTimelineEvent{}, Subjects: []TokenizableSubject{}, Diagrams: []ReportingDiagram{},
			Assets: []ReportingAssetDeclaration{}, SupportIndex: []ReportingSupportRef{}, SourceIndex: []ReportingSourceRef{},
			ValidationSummary: ReportingExportModelValidation{SchemaID: "cartulary.reporting_export_model_validation.v1", Result: "passed", Issues: []ReportingValidationIssue{}},
		},
	})
}

func projectRetainedMetadata(ctx context.Context, source ExportModel, retained *RedactedExportModel, fields map[string]ReportingField, effective []string, profile RedactionProfileView) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	// These collections are derived later from admitted template/composition
	// declarations. Source admission must not smuggle renderer-owned objects.
	if len(source.Diagrams) != 0 || len(source.Assets) != 0 {
		return newRenderValidationError("export_model_invalid", "unsupported_derivation_algorithm")
	}
	index, err := redactionSubjectIndex(source.SnapshotContent)
	if err != nil {
		return err
	}
	subjectRefs := map[string]reportingContainerRefs{}
	for id, field := range fields {
		for _, subject := range index[id] {
			refs, exists := subjectRefs[subject.StableSubjectRef]
			if !exists {
				refs = emptyContainerRefs()
			}
			refs.include(field.SourceRefs, field.SupportRefs, field.DisclosurePartitionRefs)
			subjectRefs[subject.StableSubjectRef] = refs
		}
	}
	for _, subject := range source.Subjects {
		refs, exists := subjectRefs[subject.StableSubjectRef]
		if !exists {
			continue
		}
		refs = refs.normalized()
		subject.SourceRefs, subject.SupportRefs, subject.DisclosurePartitionRefs = refs.source, refs.support, refs.partitions
		retained.Subjects = append(retained.Subjects, subject)
	}
	retained.Relationships = projectRetainedRelationships(source, retained.Records, retained.Subjects, fields)
	for i := range retained.PresentationTexts {
		body := &retained.PresentationTexts[i].Body
		refs := emptyContainerRefs()
		for _, id := range body.SourceRefs {
			if field, exists := fields[id]; exists {
				refs.include(field.SourceRefs, field.SupportRefs, nil)
			}
		}
		refs = refs.normalized()
		body.SourceRefs, body.SupportRefs = refs.source, refs.support
	}
	if err := projectRetainedIndexes(source.SnapshotContent, retained, fields, effective, profile); err != nil {
		return err
	}
	// Retained arrays and optional identity pointers must not alias immutable
	// admission content or another render attempt.
	copy, err := copyReportingModel(retained.ExportModel)
	if err != nil {
		return err
	}
	retained.ExportModel = copy
	if err := validateRetainedProjection(retained.ExportModel); err != nil {
		return err
	}
	return ctx.Err()
}

func projectRetainedRelationships(source ExportModel, records []ReportingRecordSummary, subjects []TokenizableSubject, fields map[string]ReportingField) []ReportingRelationshipSummary {
	result := []ReportingRelationshipSummary{}
	recordsByID := map[string]ReportingRecordSummary{}
	for _, record := range records {
		recordsByID[record.RecordID] = record
	}
	subjectIDs := map[string]bool{}
	for _, subject := range subjects {
		subjectIDs[subject.StableSubjectRef] = true
	}
	sources := reportingFieldSources(source.SnapshotContent)
	for _, relationship := range source.Relationships {
		src, srcExists := recordsByID[relationship.SrcRecordRef.SourceRecordRef.SourceRecordID]
		dst, dstExists := recordsByID[relationship.DstRecordRef.SourceRecordRef.SourceRecordID]
		if !srcExists || !dstExists {
			continue
		}
		refs := emptyContainerRefs()
		values := map[string]any{}
		for _, ref := range relationship.SourceRefs {
			field, exists := fields[ref]
			if !exists {
				continue
			}
			values[sources[ref].Field.FieldKey] = field.Value
			refs.include(field.SourceRefs, field.SupportRefs, field.DisclosurePartitionRefs)
		}
		// A derived edge requires all structural inputs to remain exact. A
		// mask, stub or dropped ID cannot create a new relationship endpoint.
		if values["src_record_id"] != src.RecordID || values["dst_record_id"] != dst.RecordID || values["link_type"] != relationship.RelationshipKind {
			continue
		}
		endpoint := func(original RelationshipEndpointRef, record ReportingRecordSummary) RelationshipEndpointRef {
			original.DisplayRef = record.DisplayName
			original.StableSubjectRef = nil
			ref := canonicalSubjectRef(record.RecordType, record.RecordID)
			if subjectIDs[ref] {
				original.StableSubjectRef = &ref
			}
			return original
		}
		relationship.SrcRecordRef = endpoint(relationship.SrcRecordRef, src)
		relationship.DstRecordRef = endpoint(relationship.DstRecordRef, dst)
		refs.include(src.SourceRefs, src.SupportRefs, src.DisclosurePartitionRefs)
		refs.include(dst.SourceRefs, dst.SupportRefs, dst.DisclosurePartitionRefs)
		refs = refs.normalized()
		relationship.SourceRefs, relationship.SupportRefs, relationship.DisclosurePartitionRefs = refs.source, refs.support, refs.partitions
		result = append(result, relationship)
	}
	return result
}

// Retain exact source/support identities needed by selected content. Historical
// source_path values are immutable snapshot addresses, not renumbered output
// paths. A citation may retain a metadata-only anchor; it never restores fields.
func projectRetainedIndexes(source SnapshotContent, retained *RedactedExportModel, fields map[string]ReportingField, effective []string, profile RedactionProfileView) error {
	invalid := newRenderValidationError("export_model_invalid", "dangling_source_ref")
	index := snapshotSourceIndex(source)
	anchors := map[sourceOwner]string{}
	for _, ref := range source.SourceIndex {
		if ref.SourcePath == nil {
			anchors[sourceOwner{ref.SourceFamily, ref.SourceRecordID}] = ref.SourceRefID
		}
	}
	wantedSources, wantedSupport := map[string][]string{}, map[string][]string{}
	mark := func(id string, labels []string) error {
		ref, exists := index[id]
		if !exists {
			return invalid
		}
		wantedSources[id] = orderedDisclosurePartitions(append(wantedSources[id], labels...))
		if ref.SourcePath != nil {
			anchor := anchors[sourceOwner{ref.SourceFamily, ref.SourceRecordID}]
			if anchor == "" {
				return invalid
			}
			wantedSources[anchor] = orderedDisclosurePartitions(append(wantedSources[anchor], labels...))
		}
		return nil
	}
	include := func(sources, support, labels []string) error {
		for _, id := range sources {
			if err := mark(id, labels); err != nil {
				return err
			}
		}
		for _, id := range support {
			wantedSupport[id] = orderedDisclosurePartitions(append(wantedSupport[id], labels...))
		}
		return nil
	}
	for _, field := range fields {
		if err := include(field.SourceRefs, field.SupportRefs, field.DisclosurePartitionRefs); err != nil {
			return err
		}
	}
	for _, text := range retained.PresentationTexts {
		if err := include(text.Body.SourceRefs, text.Body.SupportRefs, text.Body.DisclosurePartitionRefs); err != nil {
			return err
		}
	}
	for _, record := range retained.Records {
		if record.SourceRecordRef.SourceRefID == nil {
			return invalid
		}
		if err := mark(*record.SourceRecordRef.SourceRefID, record.DisclosurePartitionRefs); err != nil {
			return err
		}
	}
	for _, subject := range retained.Subjects {
		family, id, err := subjectSourceIdentity(subject)
		if err != nil {
			return err
		}
		if err := mark(anchors[sourceOwner{family, id}], subject.DisclosurePartitionRefs); err != nil {
			return err
		}
	}
	for _, relation := range retained.Relationships {
		for _, endpoint := range []RelationshipEndpointRef{relation.SrcRecordRef, relation.DstRecordRef} {
			if endpoint.SourceRecordRef.SourceRefID == nil {
				return invalid
			}
			if err := mark(*endpoint.SourceRecordRef.SourceRefID, relation.DisclosurePartitionRefs); err != nil {
				return err
			}
		}
	}
	for _, support := range source.SupportIndex {
		labels, needed := wantedSupport[support.SupportRefID]
		if !needed {
			continue
		}
		if support.SourceRefID != nil {
			if err := mark(*support.SourceRefID, labels); err != nil {
				return err
			}
		}
		// No source excerpt is copied into an independent unchecked label.
		support.SourceSummary = nil
		support.DisclosurePartitionRefs = labels
		retained.SupportIndex = append(retained.SupportIndex, support)
		delete(wantedSupport, support.SupportRefID)
	}
	if len(wantedSupport) != 0 {
		return invalidSupportRef()
	}
	allowed, declared := map[string]bool{}, map[string]bool{}
	for _, label := range profile.AllowedDisclosurePartitionRefs {
		declared[label] = true
	}
	for _, label := range effective {
		allowed[label] = declared[label]
	}
	for _, original := range source.SourceIndex {
		labels, needed := wantedSources[original.SourceRefID]
		if !needed {
			continue
		}
		if validateDisclosurePartitionSet(labels) != nil {
			return invalid
		}
		for _, label := range labels {
			if !allowed[label] {
				return newRenderValidationError("redaction_manifest_invalid", "disclosure_partition_unresolved")
			}
		}
		original.SourceSummary = nil
		original.DisclosurePartitionRefs = labels
		retained.SourceIndex = append(retained.SourceIndex, original)
	}
	sort.Slice(retained.Subjects, func(i, j int) bool {
		return retained.Subjects[i].StableSubjectRef < retained.Subjects[j].StableSubjectRef
	})
	return nil
}
