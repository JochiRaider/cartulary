package reporting

import (
	"context"

	"github.com/JochiRaider/cartulary/internal/modules/reporting/exportprovider"
)

// Subject membership is a frozen source fact, not a label, current database
// lookup or permission asserted by a renderer. Existing source restrictions are
// retained: assigning a Party never removes an earlier blocked/internal ref.
func assignSubjectFieldPartitions(ctx context.Context, model SnapshotContent, facts exportprovider.SourceFacts, fields []ExportField, scope string) ([]ExportField, error) {
	if err := validateSnapshotSourceFacts(facts, model); err != nil {
		return nil, err
	}
	if !validReleaseScope(scope) {
		return nil, newRenderValidationError("invalid_release_request", "unsupported_release_scope")
	}
	states := map[string]lifecycleRestriction{}
	for _, record := range model.Records {
		if record.DeletedState == "superseded" {
			states[record.RecordID] = lifecycleSuperseded
		} else {
			states[record.RecordID] = lifecycleActive
		}
	}
	for _, record := range facts.DeletedRecords {
		states[record.RecordID] = lifecycleDeleted
	}
	type subjectKey struct{ kind, id string }
	assigned := map[subjectKey][]string{}
	parties := map[string]string{}
	for _, party := range facts.Parties {
		parties[party.PartyID] = "party:" + party.PartitionSegment
	}
	for _, assignment := range facts.Assignments {
		// Responsibility and observation are not recipient permissions.
		if assignment.AssignmentKind != "subject_party" {
			continue
		}
		key := subjectKey{assignment.SubjectRecordType, assignment.SubjectRecordID.String()}
		assigned[key] = append(assigned[key], parties[assignment.PartyID.String()])
	}
	unassigned := "internal_only"
	if scope == ReleaseScopeExternal {
		unassigned = "blocked"
	}
	result := make([]ExportField, len(fields))
	for i, field := range fields {
		result[i] = field
		if field.SourceFamily == "record_envelope" || coreTypeForSourceFamily(field.SourceFamily) != "" {
			result[i].lifecycle = states[field.SourceRecordID]
		}
		var partitions []string
		// Both envelope and domain projections carry the declared Core type.
		// No fallback lookup through rendered summaries can change that identity.
		kind := field.SourceRecordType
		switch kind {
		case "party":
			ref, exists := parties[field.SourceRecordID]
			if !exists {
				return nil, newRenderValidationError("export_model_invalid", "dangling_source_ref")
			}
			// Public-directory eligibility alone is insufficient. No current
			// built-in profile grants the additional public-label permission.
			partitions = []string{ref}
		case "host", "identity":
			partitions = assigned[subjectKey{kind, field.SourceRecordID}]
			if len(partitions) == 0 {
				partitions = []string{unassigned}
			}
		default:
			// Reference inheritance follows below. This direct-subject pass
			// does not classify unlinked free text or assert aggregate proof.
			partitions = []string{}
		}
		result[i].DisclosurePartitionRefs = orderedDisclosurePartitions(append(append([]string{}, field.DisclosurePartitionRefs...), partitions...))
	}
	return assignReferencedFieldPartitions(ctx, model, result, scope, states)
}
