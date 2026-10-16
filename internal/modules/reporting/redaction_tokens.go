package reporting

import (
	"fmt"
	"sort"
	"strings"
)

type redactionDisplayToken struct {
	TokenID          string
	DisplayToken     string
	StableSubjectRef string
	SubjectKind      string
	Subject          TokenizableSubject
	DisplayValue     string
}

func canonicalSubjectRef(kind, recordID string) string {
	return "subject:" + kind + ":record:" + recordID
}

// Subject binding comes from the immutable model's explicit record/source
// relationships. A path spelling alone must never create a subject identity.
func redactionSubjectIndex(model SnapshotContent) (map[string][]TokenizableSubject, error) {
	groups, err := snapshotSourceGroups(model)
	if err != nil {
		return nil, err
	}
	type sourceKey struct{ family, id string }
	recordPaths := map[sourceKey][]string{}
	for _, record := range groups {
		ref := record.SourceRecordRef
		if ref.SourceSnapshotID != model.SnapshotID {
			continue
		}
		key := sourceKey{ref.SourceFamily, ref.SourceRecordID}
		if _, exists := recordPaths[key]; exists {
			return nil, newRenderValidationError("token_manifest_invalid", "token_subject_not_unique")
		}
		recordPaths[key] = record.SourceRefs
	}
	index := map[string][]TokenizableSubject{}
	subjects := map[string]TokenizableSubject{}
	for _, subject := range model.Subjects {
		subjects[subject.StableSubjectRef] = subject
		family, id, err := subjectSourceIdentity(subject)
		if err != nil {
			return nil, err
		}
		paths, exists := recordPaths[sourceKey{family, id}]
		if !exists {
			return nil, newRenderValidationError("token_manifest_invalid", "token_subject_not_stable")
		}
		for _, path := range paths {
			index[path] = append(index[path], subject)
		}
	}
	for _, record := range groups {
		if record.SourceRecordRef.SourceFamily != "entity_mention" {
			continue
		}
		binding, err := reportingMentionSource(record)
		if err != nil {
			return nil, err
		}
		if binding.ResolvedRecordID == nil {
			continue
		}
		subject, exists := subjects[canonicalSubjectRef(binding.Family, *binding.ResolvedRecordID)]
		if !exists {
			return nil, newRenderValidationError("token_manifest_invalid", "token_subject_not_stable")
		}
		for _, path := range record.SourceRefs {
			index[path] = append(index[path], subject)
		}
	}
	return index, nil
}

func assignDisplayTokens(model ExportModel, selected []selectedRedactionField, neutralFamily bool, presentation map[string][]TokenizableSubject) (map[string]redactionDisplayToken, error) {
	subjectIndex, err := redactionSubjectIndex(model.SnapshotContent)
	if err != nil {
		return nil, err
	}
	for path, subjects := range presentation {
		if _, exists := subjectIndex[path]; exists {
			return nil, newRenderValidationError("token_manifest_invalid", "token_subject_not_unique")
		}
		subjectIndex[path] = subjects
	}
	byPath := map[string]redactionDisplayToken{}
	bySubject := map[string]redactionDisplayToken{}
	groups, err := snapshotSourceGroups(model.SnapshotContent)
	if err != nil {
		return nil, err
	}
	displayValues := map[[2]string]*string{}
	for _, record := range groups {
		displayValues[[2]string{record.SourceRecordRef.SourceFamily, record.SourceRecordRef.SourceRecordID}] = record.DisplayName
	}
	for _, selection := range selected {
		field := selection.Field
		if !selection.Action.TokenBacked {
			continue
		}
		subjects := subjectIndex[field.Path]
		if len(subjects) != 1 {
			return nil, newRenderValidationError("token_manifest_invalid", "token_subject_not_unique")
		}
		subject := subjects[0]
		family, id, err := subjectSourceIdentity(subject)
		if err != nil {
			return nil, err
		}
		display := displayValues[[2]string{family, id}]
		if display == nil || !validReportingText(*display, true) {
			return nil, newRenderValidationError("token_manifest_invalid", "token_subject_not_stable")
		}
		digest, err := subjectDisplayValueDigest(subject.StableSubjectRef, display)
		if err != nil || digest != subject.CanonicalDisplayValueSHA256 {
			return nil, newRenderValidationError("token_manifest_invalid", "token_subject_not_stable")
		}
		token := redactionDisplayToken{StableSubjectRef: subject.StableSubjectRef, SubjectKind: subject.SubjectFamily, Subject: subject, DisplayValue: *display}
		byPath[field.Path] = token
		bySubject[subject.StableSubjectRef] = token
	}
	refs := make([]string, 0, len(bySubject))
	for ref := range bySubject {
		refs = append(refs, ref)
	}
	sort.Strings(refs)
	ordinals := map[string]int{}
	for _, ref := range refs {
		token := bySubject[ref]
		family := strings.ToUpper(token.SubjectKind)
		if neutralFamily {
			family = "SUBJECT"
		}
		ordinals[family]++
		token.DisplayToken = fmt.Sprintf("%s-%04d", family, ordinals[family])
		id, err := model.RenderIdentity.generatedID("token", "tok_", map[string]any{"stable_subject_ref": ref})
		if err != nil {
			return nil, err
		}
		token.TokenID = id
		bySubject[ref] = token
	}
	for path, token := range byPath {
		byPath[path] = bySubject[token.StableSubjectRef]
	}
	return byPath, nil
}
