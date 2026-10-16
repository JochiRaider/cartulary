package reporting

import "github.com/JochiRaider/cartulary/internal/modules/reporting/exportprovider"

// Validate source identity against the frozen content, not mutable live records
// or path spelling. Assignment and Party facts are inputs to disclosure policy;
// their presence cannot itself authorize a field in distributable output.
func validateSnapshotSourceFacts(facts exportprovider.SourceFacts, content SnapshotContent) error {
	issues := renderIssueCollector{stage: "export_model_materialization"}
	invalid := newRenderValidationError("export_model_invalid", "invalid_type")
	dangling := newRenderValidationError("export_model_invalid", "dangling_source_ref")
	if facts.SchemaID != exportprovider.SourceFactsSchemaID || facts.Parties == nil || facts.Assignments == nil || facts.DeletedRecords == nil {
		issues.add(invalid)
	}
	deleted := map[string]bool{}
	previousDeleted := ""
	for _, record := range facts.DeletedRecords {
		one := exportprovider.EmptySourceFacts()
		one.DeletedRecords = append(one.DeletedRecords, record)
		if one.Validate() != nil || record.RecordID <= previousDeleted {
			issues.add(invalid)
		}
		previousDeleted = record.RecordID
		deleted[record.RecordID] = true
	}
	for _, source := range content.SourceIndex {
		if deleted[source.SourceRecordID] && (source.SourceFamily == "record_envelope" || coreTypeForSourceFamily(source.SourceFamily) != "") {
			issues.add(dangling)
		}
	}
	types := map[string]string{}
	ambiguous := map[string]bool{}
	parties := map[string]bool{}
	for _, record := range content.Records {
		// Generic envelopes are a separate projection of the same source ID,
		// not a second Host, Identity or Party endpoint.
		switch record.RecordType {
		case "host", "identity", "party":
		default:
			continue
		}
		if record.SourceRecordRef.SourceFamily != record.RecordType {
			continue
		}
		id := record.SourceRecordRef.SourceRecordID
		if _, exists := types[id]; exists {
			ambiguous[id] = true
		}
		types[id] = record.RecordType
		if record.RecordType == "party" {
			parties[id] = false
		}
	}
	previous := ""
	for _, party := range facts.Parties {
		one := exportprovider.EmptySourceFacts()
		one.Parties = append(one.Parties, party)
		if one.Validate() != nil || party.PartyID <= previous {
			issues.add(invalid)
		}
		previous = party.PartyID
		if _, exists := parties[party.PartyID]; !exists {
			issues.add(dangling)
			continue
		}
		parties[party.PartyID] = true
	}
	for _, present := range parties {
		if !present {
			issues.add(dangling)
		}
	}
	previous = ""
	seen := map[[3]string]bool{}
	for _, assignment := range facts.Assignments {
		id := assignment.AssignmentID.String()
		if assignment.Validate() != nil || id <= previous {
			issues.add(invalid)
		}
		previous = id
		key := [3]string{assignment.SubjectRecordID.String(), assignment.PartyID.String(), assignment.AssignmentKind}
		if seen[key] {
			issues.add(newRenderValidationError("export_model_invalid", "duplicate_active_party_assignment"))
		}
		seen[key] = true
		if assignment.IncidentID.String() != content.IncidentID || ambiguous[key[0]] || ambiguous[key[1]] || types[key[0]] != assignment.SubjectRecordType || !parties[key[1]] {
			issues.add(dangling)
		}
	}
	return issues.finish()
}

func validateFrozenSourceFacts(input renderInput) error {
	if err := validateSnapshotSourceFacts(input.SourceFacts, input.ExportModel.SnapshotContent); err != nil {
		return err
	}
	digest, err := sourceFactsDigest(input.SourceFacts)
	if err != nil || digest != input.SourceFactsSHA256 {
		return newRenderValidationError("export_model_invalid", "invalid_type")
	}
	return nil
}

func sourceFactsDigest(facts exportprovider.SourceFacts) (string, error) {
	if err := facts.Validate(); err != nil {
		return "", err
	}
	raw, err := canonicalJSON(facts)
	if err != nil {
		return "", err
	}
	return reportingObjectDigest(exportprovider.SourceFactsSchemaID, raw), nil
}
