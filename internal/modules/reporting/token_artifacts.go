package reporting

import (
	"bytes"
	"errors"
	"fmt"
	"sort"
	"strings"
)

// This boundary is shared by native output persistence and portable admission.
// Import never upgrades old bytes or trusts their digest as proof that their
// schema, selected rules and subject inventory satisfy the current contract.
func validatePublicRedactionArtifacts(bundle RenderBundleManifest, files []RenderBundleFile) error {
	invalid := errors.New("reporting: invalid public redaction artifacts")
	evidence := map[string][]byte{}
	for _, file := range files {
		switch file.Role {
		case renderBundleRoleRedactionManifest, renderBundleRoleRedactionProfileView, renderBundleRoleTokenManifest:
			if _, exists := evidence[file.Role]; exists {
				return invalid
			}
			evidence[file.Role] = file.Bytes
		}
	}
	raw := evidence[renderBundleRoleRedactionManifest]
	manifest, canonical, err := decodePortableObject[RedactionManifest](raw, bundle.RedactionManifestSHA256)
	if err != nil || !bytes.Equal(raw, canonical) || validateRedactionManifest(manifest) != nil ||
		manifest.SnapshotID != bundle.SnapshotID || !sameReportingValue(manifest.RenderIdentity, bundle.RenderIdentity) {
		return invalid
	}
	raw = evidence[renderBundleRoleRedactionProfileView]
	profile, canonical, err := decodePortableObject[RedactionProfileView](raw, reportingObjectDigest(RedactionProfileViewSchemaID, raw))
	if err != nil || !bytes.Equal(raw, canonical) || validateRedactionProfileView(profile) != nil ||
		profile.ProfileID != manifest.ProfileID || profile.ProfileVersion != manifest.ProfileVersion || profile.ProfileSHA256 != manifest.ProfileSHA256 ||
		!profileSupportsScope(profile, bundle.ReleaseScope) {
		return invalid
	}
	tokenized := false
	for _, entry := range manifest.Entries {
		ruleID, action, trace := selectRedactionRule(ExportField{ContentClass: entry.ContentClass}, entry.Path, profile)
		if entry.RuleID != ruleID || entry.Action != action.Type || !sameReportingValue(entry.SelectedRuleTrace, trace) || action.TokenBacked != (entry.TokenID != nil) {
			return invalid
		}
		tokenized = tokenized || entry.TokenID != nil
	}
	raw, hasTokens := evidence[renderBundleRoleTokenManifest]
	if hasTokens != tokenized || hasTokens != (bundle.TokenManifestSHA256 != nil) {
		return invalid
	}
	if !hasTokens {
		return nil
	}
	tokens, canonical, err := decodePortableObject[RedactionTokenManifest](raw, *bundle.TokenManifestSHA256)
	if err != nil || !bytes.Equal(raw, canonical) || validateTokenManifestBindings(tokens, manifest, bundle.RedactionManifestSHA256) != nil {
		return invalid
	}
	for _, entry := range tokens.Entries {
		if !sameReportingValue(entry.RecipientPartitionRefs, tokens.Entries[0].RecipientPartitionRefs) ||
			bundle.ReleaseScope != ReleaseScopeExternal && len(entry.RecipientPartitionRefs) != 0 {
			return invalid
		}
	}
	return nil
}

// Rule selection belongs to each model address in the redaction manifest.
// Token artifacts describe one subject, even when different fields selected
// different token-backed actions. The digest edges are acyclic:
// redaction manifest <- token manifest <- private reveal map.
func buildTokenArtifacts(model ExportModel, tokens map[string]redactionDisplayToken, manifestSHA, profileSHA string, recipients []string) (*RedactionTokenManifest, []byte, string, *RedactionRevealMap, []byte, string, error) {
	if len(tokens) == 0 {
		return nil, nil, "", nil, nil, "", nil
	}
	byID := map[string]redactionDisplayToken{}
	for _, token := range tokens {
		byID[token.TokenID] = token
	}
	ids := make([]string, 0, len(byID))
	for id := range byID {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	manifest := RedactionTokenManifest{
		RenderIdentity: model.RenderIdentity.copy(), SnapshotID: model.SnapshotID,
		CreatedAt: model.RenderAdmittedAt, SchemaID: RedactionTokenManifestSchemaID,
		ProfileSHA256: profileSHA, RedactionManifestSHA256: manifestSHA,
		Entries: make([]RedactionTokenManifestEntry, 0, len(ids)),
	}
	reveal := RedactionRevealMap{
		RenderIdentity: model.RenderIdentity.copy(), SnapshotID: model.SnapshotID,
		CreatedAt: model.RenderAdmittedAt, SchemaID: RedactionRevealMapSchemaID,
		ProfileSHA256: profileSHA, Entries: make([]RedactionRevealMapEntry, 0, len(ids)),
	}
	for _, id := range ids {
		token := byID[id]
		subject := token.Subject
		manifest.Entries = append(manifest.Entries, RedactionTokenManifestEntry{
			TokenID: id, DisplayToken: token.DisplayToken, StableSubjectRef: subject.StableSubjectRef,
			SubjectFamily: subject.SubjectFamily, SourceRecordID: copyProfileValue(subject.SourceRecordID),
			EntityMentionID: copyProfileValue(subject.EntityMentionID), CanonicalDisplayValueSHA256: subject.CanonicalDisplayValueSHA256,
			RecipientPartitionRefs: append([]string{}, recipients...),
		})
		reveal.Entries = append(reveal.Entries, RedactionRevealMapEntry{
			TokenID: id, DisplayToken: token.DisplayToken, StableSubjectRef: subject.StableSubjectRef,
			SubjectFamily: subject.SubjectFamily, SourceRecordID: copyProfileValue(subject.SourceRecordID),
			EntityMentionID: copyProfileValue(subject.EntityMentionID), CanonicalDisplayValueSHA256: subject.CanonicalDisplayValueSHA256,
			CanonicalDisplayValue: token.DisplayValue, RecipientPartitionRefs: append([]string{}, recipients...),
		})
	}
	manifestBytes, err := canonicalJSON(manifest)
	if err != nil {
		return nil, nil, "", nil, nil, "", err
	}
	manifestDigest := reportingObjectDigest(manifest.SchemaID, manifestBytes)
	reveal.TokenManifestSHA256 = manifestDigest
	revealBytes, err := canonicalJSON(reveal)
	if err != nil {
		return nil, nil, "", nil, nil, "", err
	}
	return &manifest, manifestBytes, manifestDigest, &reveal, revealBytes, reportingObjectDigest(reveal.SchemaID, revealBytes), nil
}

func validRedactionEntry(entry RedactionManifestEntry) bool {
	if !validExactExportModelPath(entry.Path) || !validReportingContentClass(entry.ContentClass) ||
		!validSelectedRuleTrace(entry.SelectedRuleTrace) || entry.RuleID != entry.SelectedRuleTrace.RuleID ||
		entry.RulePrecedence != entry.SelectedRuleTrace.RulePrecedence || entry.DisclosurePartitionRefs == nil ||
		validateDisclosurePartitionSet(entry.DisclosurePartitionRefs) != nil ||
		!sameReportingValue(entry.DisclosurePartitionRefs, orderedDisclosurePartitions(entry.DisclosurePartitionRefs)) {
		return false
	}
	if entry.RulePrecedence == "path" && (entry.SelectedRuleTrace.SelectorValue == nil || *entry.SelectedRuleTrace.SelectorValue != entry.Path) {
		return false
	}
	if entry.RulePrecedence == "content_class" && (entry.SelectedRuleTrace.SelectorValue == nil || *entry.SelectedRuleTrace.SelectorValue != entry.ContentClass) {
		return false
	}
	if entry.DisclosurePartitionHandling != disclosurePartitionHandling(ExportField{DisclosurePartitionRefs: entry.DisclosurePartitionRefs}, entry.Outcome) {
		return false
	}
	if entry.Outcome == "tokenized" {
		return (entry.Action == ActionMask || entry.Action == ActionStub) && entry.TokenID != nil && validReportingIdentifier(*entry.TokenID)
	}
	if entry.TokenID != nil {
		return false
	}
	return map[string]string{ActionAllow: "unchanged", ActionDrop: "dropped", ActionMask: "masked", ActionTruncate: "truncated", ActionStub: "stubbed"}[entry.Action] == entry.Outcome && entry.Outcome != ""
}

func validateRedactionManifest(manifest RedactionManifest) error {
	invalid := errors.New("reporting: invalid redaction manifest")
	if manifest.SchemaID != RedactionManifestSchemaID || manifest.RenderIdentity.validate() != nil || manifest.CreatedAt != manifest.RenderAdmittedAt ||
		!validReportingIdentifier(manifest.SnapshotID) || !validReportingIdentifier(manifest.ProfileID) || !validReportingIdentifier(manifest.ProfileVersion) ||
		!sha256HexPattern.MatchString(manifest.ProfileSHA256) || manifest.Entries == nil {
		return invalid
	}
	previous := ""
	for _, entry := range manifest.Entries {
		if entry.Path <= previous || !validRedactionEntry(entry) {
			return invalid
		}
		previous = entry.Path
	}
	if err := validateContainerDecisions(manifest); err != nil {
		return err
	}
	return validateStructuralRemovals(structuralDecisionEntries(manifest.Entries, manifest.ContainerDecisions), manifest.StructuralRemovals)
}

func validateTokenManifestBindings(tokens RedactionTokenManifest, manifest RedactionManifest, manifestSHA string) error {
	invalid := errors.New("reporting: invalid token manifest binding")
	if validateRedactionManifest(manifest) != nil || tokens.SchemaID != RedactionTokenManifestSchemaID || tokens.RedactionManifestSHA256 != manifestSHA ||
		tokens.ProfileSHA256 != manifest.ProfileSHA256 || tokens.SnapshotID != manifest.SnapshotID || tokens.CreatedAt != manifest.CreatedAt ||
		!sameReportingValue(tokens.RenderIdentity, manifest.RenderIdentity) || len(tokens.Entries) == 0 {
		return invalid
	}
	uses := map[string]bool{}
	for _, entry := range manifest.Entries {
		if entry.TokenID != nil {
			uses[*entry.TokenID] = true
		}
	}
	previous := ""
	displays, subjects := map[string]bool{}, map[string]bool{}
	for _, token := range tokens.Entries {
		if token.TokenID <= previous || !uses[token.TokenID] || displays[token.DisplayToken] || subjects[token.StableSubjectRef] ||
			!sha256HexPattern.MatchString(token.CanonicalDisplayValueSHA256) || token.RecipientPartitionRefs == nil || validateRecipientPartitionRefs(token.RecipientPartitionRefs) != nil {
			return invalid
		}
		previous = token.TokenID
		subject := TokenizableSubject{SchemaID: tokenizableSubjectSchemaID, StableSubjectRef: token.StableSubjectRef, SubjectFamily: token.SubjectFamily, SourceRecordID: token.SourceRecordID, EntityMentionID: token.EntityMentionID}
		if token.SourceRecordID != nil {
			subject.SourceRecordType = &token.SubjectFamily
		}
		if _, _, err := subjectSourceIdentity(subject); err != nil {
			return invalid
		}
		id, err := tokens.RenderIdentity.generatedID("token", "tok_", map[string]any{"stable_subject_ref": token.StableSubjectRef})
		if err != nil || id != token.TokenID {
			return invalid
		}
		displays[token.DisplayToken], subjects[token.StableSubjectRef] = true, true
		delete(uses, token.TokenID)
	}
	if len(uses) != 0 || !validTokenLabels(tokens.Entries) {
		return invalid
	}
	return nil
}

func validateRecipientPartitionRefs(refs []string) error {
	previous := ""
	for _, ref := range refs {
		if !validPartyPartitionRef(ref) || ref <= previous {
			return errors.New("invalid recipient partition")
		}
		previous = ref
	}
	return nil
}

func validTokenLabels(entries []RedactionTokenManifestEntry) bool {
	if len(entries) == 0 {
		return false
	}
	ordered := append([]RedactionTokenManifestEntry(nil), entries...)
	sort.Slice(ordered, func(i, j int) bool { return ordered[i].StableSubjectRef < ordered[j].StableSubjectRef })
	neutral := strings.HasPrefix(ordered[0].DisplayToken, "SUBJECT-")
	ordinals := map[string]int{}
	for _, entry := range ordered {
		family := strings.ToUpper(entry.SubjectFamily)
		if neutral {
			family = "SUBJECT"
		}
		ordinals[family]++
		if entry.DisplayToken != fmt.Sprintf("%s-%04d", family, ordinals[family]) {
			return false
		}
	}
	return true
}
