package reporting

// Admission, redaction, replay and recovery use one source-content boundary.
// Each validator examines its complete owned invariant set; the collector
// retains only bounded, value-free diagnostics and rejects the whole model.
func validateSnapshotContent(model SnapshotContent) error {
	issues := renderIssueCollector{stage: "export_model_materialization"}
	issues.add(validateSnapshotSourceReferences(model))
	issues.add(validateSnapshotSupportReferences(model))
	issues.add(validateSnapshotBlockClasses(model))
	issues.add(validateSnapshotSectionStructure(model))
	issues.add(validateSnapshotFieldValues(model))
	issues.add(validateSnapshotFieldBindings(model))
	issues.add(validateCoreRecordProjection(model))
	issues.add(snapshotLifecycleFields(&model, false))
	issues.add(validateSnapshotSubjects(model))
	seenRecords := map[string]bool{}
	for _, record := range model.Records {
		if record.RecordID == "" || !validCoreRecordType(record.RecordType) || record.SourceRecordRef.SourceFamily == "" || record.SourceRecordRef.SourceFamily == "incident_metadata" || record.SourceRecordRef.SourceRecordID != record.RecordID {
			issues.add(newRenderValidationError("export_model_invalid", "dangling_source_ref"))
		}
		if seenRecords[record.RecordID] {
			issues.add(newRenderValidationError("export_model_invalid", "duplicate_stable_id"))
		}
		seenRecords[record.RecordID] = true
		if record.SourceRecordRef.SourceSnapshotID != model.SnapshotID {
			issues.add(newRenderValidationError("export_model_invalid", "snapshot_ref_mismatch"))
		}
	}
	for _, relationship := range model.Relationships {
		if relationship.SrcRecordRef.SourceRecordRef.SourceSnapshotID != model.SnapshotID || relationship.DstRecordRef.SourceRecordRef.SourceSnapshotID != model.SnapshotID {
			issues.add(newRenderValidationError("export_model_invalid", "snapshot_ref_mismatch"))
		}
	}
	return issues.finish()
}
