package reporting

import (
	"context"
	"encoding/json"
	"sort"
	"strings"
	"unicode/utf8"

	"golang.org/x/text/unicode/norm"
)

// Presentation text is a derived value, not a fabricated source record. Its
// decision address remains stable when deck operations reorder or remove slides.
type ReportingPresentationText struct {
	SchemaID       string         `json:"schema_id"`
	AuthoredTextID string         `json:"authored_text_id"`
	TextRole       string         `json:"text_role"`
	ContentClass   string         `json:"content_class"`
	Body           ReportingField `json:"body"`
}

type renderSelectionInput struct {
	template    *TemplateContract
	composition *renderComposition
	limits      renderLimits
	graphs      graphRedactionSelection
}

type compositionTextSelection struct {
	text     compositionAuthoredText
	subjects []TokenizableSubject
}

func redactRenderInput(ctx context.Context, input renderInput) (RedactionResult, error) {
	if err := validateTemplateContract(input.TemplateContract, input.OutputKind, input.ReleaseScope); err != nil {
		return RedactionResult{}, classifyBundleFailure(err)
	}
	view, _, _, err := BuildRedactionProfileView(input.RedactionProfile, input.RedactionProfileSHA256)
	if err != nil {
		return RedactionResult{}, err
	}
	identity := RedactedExportModel{ExportModel: ExportModel{SnapshotContent: SnapshotContent{IncidentID: input.ExportModel.IncidentID}, RenderIdentity: input.ExportModel.RenderIdentity}}
	if err := validateCompositionRenderBinding(input.CompositionJSON, input.TemplateContract, identity, input.ReleaseScope, view); err != nil {
		return RedactionResult{}, err
	}
	composition, err := parseRenderComposition(input.CompositionJSON)
	if err != nil {
		return RedactionResult{}, err
	}
	limits, err := resolveRenderLimits(input.TemplateContract.DeclaredLimits)
	if err != nil {
		return RedactionResult{}, err
	}
	graphs, err := prepareGraphRedaction(ctx, input, composition)
	if err != nil {
		return RedactionResult{}, err
	}
	result, err := redactExportModel(ctx, input.ExportModel, input.SourceFacts, input.RedactionProfile, input.RedactionProfileSHA256, input.ExportModelSHA256, input.ReleaseScope, input.RecipientPartitionRefs, renderSelectionInput{composition: composition, limits: limits, template: &input.TemplateContract, graphs: graphs})
	if err != nil {
		return RedactionResult{}, err
	}
	return assembleRenderModel(ctx, input, result, graphs.diagrams)
}

func compositionTextPolicy(role string, limits renderLimits) (int64, string, bool, error) {
	key, reason, multiline := "composition.authored_text_chars", "authored_text_limit_exceeded", true
	switch role {
	case "title_override":
		key, reason, multiline = "composition.authored_title_chars", "authored_title_limit_exceeded", false
	case "speaker_notes":
		key = "speaker_notes_chars_per_slide"
	case "authored_text":
	default:
		return 0, "", false, compositionFailure("invalid_schema_id")
	}
	return limits[key], reason, multiline, nil
}

func validateCompositionTextBody(text compositionAuthoredText, limits renderLimits) error {
	limit, reason, multiline, err := compositionTextPolicy(text.TextRole, limits)
	if err != nil {
		return err
	}
	if !norm.NFC.IsNormalString(text.Body) || !validReportingText(text.Body, multiline) || strings.TrimSpace(text.Body) == "" || int64(utf8.RuneCountInString(text.Body)) > limit {
		return compositionFailure(reason)
	}
	return nil
}

// Parsing is deliberately limited to the companion's closed placeholder
// grammar. Literal prose is never scanned for inferred subjects or case facts.
func substituteAuthoredSubjects(body string, limit int64, reason string, resolve func(string) (string, error)) (string, error) {
	var result strings.Builder
	var count int64
	appendBounded := func(part string) error {
		size := int64(utf8.RuneCountInString(part))
		if size > limit-count {
			return compositionFailure(reason)
		}
		count += size
		result.WriteString(part)
		return nil
	}
	for {
		start := strings.Index(body, "{{subject:")
		if start < 0 {
			if err := appendBounded(body); err != nil {
				return "", err
			}
			return result.String(), nil
		}
		if err := appendBounded(body[:start]); err != nil {
			return "", err
		}
		rest := body[start+len("{{subject:"):]
		end := strings.Index(rest, "}}")
		if end <= 0 {
			return "", compositionFailure("authored_subject_ref_unresolved")
		}
		value, err := resolve(rest[:end])
		if err != nil {
			return "", err
		}
		if err := appendBounded(value); err != nil {
			return "", err
		}
		body = rest[end+2:]
	}
}

func presentationTextPath(id string) (string, error) {
	path, err := keyedModelPath("$", "presentation_texts", id)
	if err != nil {
		return "", err
	}
	return path + ".body", nil
}

func addCompositionTextDecisionPaths(raw json.RawMessage, paths map[string]string) error {
	composition, err := parseRenderComposition(raw)
	if err != nil {
		return err
	}
	if composition == nil {
		return nil
	}
	for _, text := range composition.AuthoredTexts {
		path, err := presentationTextPath(text.AuthoredTextID)
		if err != nil {
			return err
		}
		if paths[path] != "" {
			return newRenderValidationError("export_model_invalid", "duplicate_export_model_path")
		}
		paths[path] = path
	}
	return nil
}

func prepareCompositionTextFields(input renderSelectionInput, model ExportModel, selected []selectedRedactionField, sources map[string]reportingFieldSource, paths map[string]string) ([]ExportField, map[string]compositionTextSelection, error) {
	fields, texts := []ExportField{}, map[string]compositionTextSelection{}
	if input.composition == nil {
		return fields, texts, nil
	}
	index, err := redactionSubjectIndex(model.SnapshotContent)
	if err != nil {
		return nil, nil, err
	}
	// A subject must already have an eligible selected display value or a
	// selected token-backed use. Merely existing in the frozen model is not enough.
	eligible := map[string]TokenizableSubject{}
	for _, selection := range selected {
		if !selection.Include {
			continue
		}
		field := sources[selection.Field.Path].Field
		for _, subject := range index[selection.Field.Path] {
			if selection.Action.TokenBacked || (field.FieldKey == "display_name" || field.FieldKey == "raw_text") && selection.Value != nil {
				eligible[subject.StableSubjectRef] = subject
			}
		}
	}
	for _, text := range input.composition.AuthoredTexts {
		if err := validateCompositionTextBody(text, input.limits); err != nil {
			return nil, nil, err
		}
		path, err := presentationTextPath(text.AuthoredTextID)
		if err != nil {
			return nil, nil, err
		}
		if paths[path] != "" {
			return nil, nil, newRenderValidationError("export_model_invalid", "duplicate_stable_id")
		}
		refs := map[string]TokenizableSubject{}
		limit, reason, _, _ := compositionTextPolicy(text.TextRole, input.limits)
		_, err = substituteAuthoredSubjects(text.Body, limit, reason, func(ref string) (string, error) {
			subject, exists := eligible[ref]
			if !exists {
				return "", compositionFailure("authored_subject_ref_unresolved")
			}
			refs[ref] = subject
			return "", nil
		})
		if err != nil {
			return nil, nil, err
		}
		ordered := make([]string, 0, len(refs))
		for ref := range refs {
			ordered = append(ordered, ref)
		}
		sort.Strings(ordered)
		selection := compositionTextSelection{text: text, subjects: []TokenizableSubject{}}
		field := ReportingField{SchemaID: "cartulary.reporting_field.v1", FieldKey: "body", FieldOrdinal: 1, SourceValueState: "derived", RedactedValueState: "unchanged", Value: text.Body, SourceRefs: []string{}, SupportRefs: []string{}, DisclosurePartitionRefs: []string{text.DisclosurePartitionRef}}
		for _, ref := range ordered {
			subject := refs[ref]
			selection.subjects = append(selection.subjects, subject)
			field.SourceRefs = appendUniqueStrings(field.SourceRefs, subject.SourceRefs...)
			field.SupportRefs = appendUniqueStrings(field.SupportRefs, subject.SupportRefs...)
		}
		paths[path], sources[path], texts[path] = path, reportingFieldSource{Field: field, Multiline: text.TextRole != "title_override"}, selection
		fields = append(fields, ExportField{Path: path, ContentClass: ContentClassPresentationText, Value: text.Body, DisclosurePartitionRefs: []string{text.DisclosurePartitionRef}, SupportRefs: field.SupportRefs, GeneratedPresentation: true})
	}
	return fields, texts, nil
}

func resolveCompositionTextValues(input renderSelectionInput, model ExportModel, selected []selectedRedactionField, tokens map[string]redactionDisplayToken, sources map[string]reportingFieldSource, texts map[string]compositionTextSelection) error {
	if len(texts) == 0 {
		return nil
	}
	index, err := redactionSubjectIndex(model.SnapshotContent)
	if err != nil {
		return err
	}
	displays, ambiguous, tokenDisplays := map[string]string{}, map[string]bool{}, map[string]string{}
	for _, token := range tokens {
		tokenDisplays[token.StableSubjectRef] = token.DisplayToken
	}
	for _, selection := range selected {
		if !selection.Include {
			continue
		}
		source := sources[selection.Field.Path]
		if source.Field.FieldKey != "display_name" && source.Field.FieldKey != "raw_text" {
			continue
		}
		value, ok := selection.Value.(string)
		if !ok || !validReportingText(value, true) {
			continue
		}
		for _, subject := range index[selection.Field.Path] {
			ref := subject.StableSubjectRef
			if previous, exists := displays[ref]; exists && previous != value {
				ambiguous[ref] = true
			}
			displays[ref] = value
		}
	}
	for i := range selected {
		selection := &selected[i]
		item, exists := texts[selection.Field.Path]
		if !exists {
			continue
		}
		limit, reason, _, _ := compositionTextPolicy(item.text.TextRole, input.limits)
		body, err := substituteAuthoredSubjects(item.text.Body, limit, reason, func(ref string) (string, error) {
			if token, exists := tokenDisplays[ref]; exists {
				return token, nil
			}
			value, exists := displays[ref]
			if !exists || ambiguous[ref] {
				return "", compositionFailure("authored_subject_ref_unresolved")
			}
			return value, nil
		})
		if err != nil {
			return err
		}
		text := item.text
		text.Body = body
		if err := validateCompositionTextBody(text, input.limits); err != nil {
			return err
		}
		value, include, outcome, err := applyRedactionAction(body, selection.Action)
		if err != nil {
			return err
		}
		selection.Value, selection.Include, selection.Outcome = value, include, outcome
		if selection.Action.TokenBacked {
			text.Body = tokens[selection.Field.Path].DisplayToken
		} else if include {
			var ok bool
			text.Body, ok = value.(string)
			if !ok {
				return compositionFailure("authored_text_limit_exceeded")
			}
		}
		if include {
			if err := validateCompositionTextBody(text, input.limits); err != nil {
				return err
			}
		}
	}
	return nil
}

// Deck derivation receives only the retained projection. Keeping a removed ID
// privately allows its operation to have no effect without restoring raw text.
func bindCompositionPresentation(composition *renderComposition, model RedactedExportModel) error {
	if composition == nil || len(composition.AuthoredTexts) == 0 {
		return nil
	}
	retained := map[string]ReportingPresentationText{}
	for _, text := range model.PresentationTexts {
		retained[text.AuthoredTextID] = text
	}
	for i, raw := range composition.AuthoredTexts {
		path, err := presentationTextPath(raw.AuthoredTextID)
		if err != nil {
			return err
		}
		encoded, err := canonicalJSON(raw)
		if err != nil || model.admittedPresentationTexts[path] != hashHex(encoded) {
			return compositionFailure("authored_subject_ref_unresolved")
		}
		text, exists := retained[raw.AuthoredTextID]
		if !exists {
			composition.AuthoredTexts[i].omitted = true
			composition.AuthoredTexts[i].Body = ""
			continue
		}
		value, ok := text.Body.Value.(string)
		if !ok || text.TextRole != raw.TextRole {
			return compositionFailure("authored_subject_ref_unresolved")
		}
		composition.AuthoredTexts[i].Body = value
		body := text.Body
		composition.AuthoredTexts[i].retainedBody = &body
	}
	return nil
}
