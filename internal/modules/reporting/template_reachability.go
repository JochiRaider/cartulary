package reporting

import (
	"context"
	"strings"

	"github.com/JochiRaider/cartulary/internal/modules/reporting/exportprovider"
)

// Reachability and lifecycle eligibility precede Core action selection. The
// allowlist grants binding syntax; it is never an implicit request for data.
// Values outside this closure cannot allocate tokens or enter public evidence.
func selectReachableSourceFields(ctx context.Context, model ExportModel, facts exportprovider.SourceFacts, fields []ExportField, bindings map[string][]string, profile RedactionProfile, scope string, recipients []string, input renderSelectionInput) ([]ExportField, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	states := map[string]lifecycleRestriction{}
	for _, field := range fields {
		states[field.Path] = field.lifecycle
	}
	eligible := func(ref string, permission bool) bool {
		state, exists := states[ref]
		return exists && state != lifecycleDeleted && (scope != ReleaseScopeExternal || state != lifecycleSuperseded || profile.IncludeSuperseded && permission)
	}
	wanted := map[string]bool{} // value: this use explicitly permits superseded data
	mark := func(ref string, permission bool) {
		if eligible(ref, permission) {
			wanted[ref] = wanted[ref] || permission
		}
	}
	if input.template == nil {
		// Unscoped owner-level redaction is not a production render admission.
		for _, field := range fields {
			mark(field.Path, false)
		}
	} else {
		subjects, err := redactionSubjectIndex(model.SnapshotContent)
		if err != nil {
			return nil, err
		}
		parties := []string{}
		partySet, recipientSet := map[string]bool{}, map[string]bool{}
		for _, ref := range recipients {
			recipientSet[ref] = true
		}
		for _, party := range facts.Parties {
			if scope != ReleaseScopeExternal || recipientSet["party:"+party.PartyID] {
				parties = append(parties, party.PartyID)
				partySet[party.PartyID] = true
			}
		}
		for _, section := range input.template.Sections {
			selectors := append([]string{}, section.Bindings...)
			for _, slot := range input.template.NarrativeSlots {
				if slot.SectionDeclID == section.DeclID {
					selectors = append(selectors, slot.Binding)
				}
			}
			labelParties := map[string]bool{}
			for _, field := range fields {
				if err := ctx.Err(); err != nil {
					return nil, err
				}
				if !eligible(field.Path, section.IncludeSuperseded) || !templateBindingsSelect(selectors, bindings[field.Path]) {
					continue
				}
				switch section.Expansion {
				case "none":
				case "per_subject":
					if len(subjects[field.Path]) == 0 {
						continue
					}
				case "per_party":
					members := candidatePartyDimensions(field.DisclosurePartitionRefs, parties, partySet, scope)
					if len(members) == 0 {
						continue
					}
					for _, id := range members {
						labelParties[id] = true
					}
				default:
					return nil, newRenderValidationError("export_model_invalid", "unsupported_derivation_algorithm")
				}
				mark(field.Path, section.IncludeSuperseded)
			}
			if section.Expansion == "per_party" && strings.Contains(section.Title, "{expansion_label}") {
				if section.SectionKind == "validation" || len(section.DiagramDecls) != 0 {
					for _, id := range parties {
						labelParties[id] = true
					}
				}
				for _, record := range model.Records {
					if record.RecordType != "party" || !labelParties[record.RecordID] {
						continue
					}
					for _, field := range record.Fields {
						if field.FieldKey == "display_name" {
							for _, ref := range field.SourceRefs {
								mark(ref, section.IncludeSuperseded)
							}
						}
					}
				}
			}
		}
		if err := selectDiagramSourceFields(model, input, mark); err != nil {
			return nil, err
		}
		if err := selectCompositionSourceFields(model, input, mark); err != nil {
			return nil, err
		}
	}
	// A selected relationship needs its structural source fields and endpoint
	// records. Support-only citations do not select record values this way.
	recordsByID := map[string]ReportingRecordSummary{}
	for _, record := range model.Records {
		recordsByID[record.RecordID] = record
	}
	for _, relationship := range model.Relationships {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		selected, superseded := false, false
		for _, ref := range relationship.SourceRefs {
			permission, exists := wanted[ref]
			selected, superseded = selected || exists, superseded || permission
		}
		if !selected {
			continue
		}
		for _, ref := range relationship.SourceRefs {
			mark(ref, superseded)
		}
		for _, endpoint := range []RelationshipEndpointRef{relationship.SrcRecordRef, relationship.DstRecordRef} {
			// Endpoint identity is an admitted Core record ID, even when its
			// source anchor names the parallel envelope projection.
			for _, ref := range recordsByID[endpoint.SourceRecordRef.SourceRecordID].SourceRefs {
				mark(ref, superseded)
			}
		}
	}
	selected := make([]ExportField, 0, len(wanted))
	for _, field := range fields {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if _, exists := wanted[field.Path]; !exists {
			continue
		}
		selected = append(selected, field)
	}
	return selected, nil
}

func candidatePartyDimensions(partitions, parties []string, admitted map[string]bool, scope string) []string {
	if len(partitions) == 0 {
		return nil
	}
	party := ""
	for _, partition := range partitions {
		if partition == "public" || partition == "internal_only" && scope != ReleaseScopeExternal {
			continue
		}
		if !strings.HasPrefix(partition, "party:") {
			return nil
		}
		id := strings.TrimPrefix(partition, "party:")
		if !admitted[id] || party != "" && party != id {
			return nil
		}
		party = id
	}
	if party != "" {
		return []string{party}
	}
	return parties
}

func selectDiagramSourceFields(model ExportModel, input renderSelectionInput, mark func(string, bool)) error {
	decls, _, err := resolveDiagramDeclarations(*input.template, input.composition)
	if err != nil {
		return err
	}
	events := map[string]ReportingTimelineEvent{}
	for _, event := range model.TimelineEvents {
		events[event.RecordID] = event
	}
	for _, decl := range decls {
		if decl.DiagramSourceKind != "timeline" {
			continue // Graph source admission has its own exact result binding.
		}
		permission := diagramAllowsSuperseded(*input.template, input.composition, decl.DeclID)
		for _, id := range decl.SelectionRule.TimelineRecordIDs {
			event, exists := events[id]
			if !exists {
				return newRenderValidationError("graph_projection_unavailable", "diagram_selection_missing_ref")
			}
			for _, ref := range event.SourceRefs {
				mark(ref, permission)
			}
		}
	}
	return nil
}

func selectCompositionSourceFields(model ExportModel, input renderSelectionInput, mark func(string, bool)) error {
	if input.composition == nil {
		return nil
	}
	subjects := map[string]TokenizableSubject{}
	for _, subject := range model.Subjects {
		subjects[subject.StableSubjectRef] = subject
	}
	sources := reportingFieldSources(model.SnapshotContent)
	for _, text := range input.composition.AuthoredTexts {
		limit, reason, _, err := compositionTextPolicy(text.TextRole, input.limits)
		if err != nil {
			return err
		}
		_, err = substituteAuthoredSubjects(text.Body, limit, reason, func(ref string) (string, error) {
			subject, exists := subjects[ref]
			if !exists {
				return "", compositionFailure("authored_subject_ref_unresolved")
			}
			for _, id := range subject.SourceRefs {
				field := sources[id].Field
				if field.FieldKey == "display_name" || field.FieldKey == "raw_text" {
					mark(id, false)
				}
			}
			return "", nil
		})
		if err != nil {
			return err
		}
	}
	for _, op := range input.composition.DeckOps {
		anchor := op.Payload.BlockAnchor
		if anchor == nil || anchor.RecordAnchor == nil || op.OpKind == "exclude_block" {
			continue
		}
		permission := false
		for _, section := range input.template.Sections {
			if section.DeclID == anchor.SectionAnchor.DeclarationID {
				permission = section.IncludeSuperseded
			}
		}
		for _, record := range model.Records {
			if record.RecordID == anchor.RecordAnchor.RecordID {
				for _, ref := range record.SourceRefs {
					mark(ref, permission)
				}
			}
		}
	}
	return nil
}
