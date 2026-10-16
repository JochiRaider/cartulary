package reporting

import (
	"bytes"
	"errors"
	"reflect"
	"slices"
	"testing"
	"time"
)

func assertExplicitSubjectTokenRules(t *testing.T) {
	t.Helper()
	assertUnresolvedMentionReveal(t)
	fields := []ExportField{
		{Path: "/hosts/z", SourceFamily: "host", ContentClass: ContentClassSourceEvidence, Value: map[string]any{"display_name": "secret-z"}},
		{Path: "/hosts/a", SourceFamily: "host", ContentClass: ContentClassSourceEvidence, Value: map[string]any{"display_name": "secret-a"}},
		{Path: "/hosts_alias/a", SourceFamily: "host", ContentClass: ContentClassSourceEvidence, Value: map[string]any{"dns_name": "secret-alias"}},
		{Path: "/identities/b", SourceFamily: "identity", ContentClass: ContentClassSourceEvidence, Value: map[string]any{"display_name": "secret-b"}},
		{Path: "/parties/c", SourceFamily: "party", ContentClass: ContentClassSourceEvidence, Value: map[string]any{"display_name": "secret-c"}},
	}
	model, sourceSHA, err := buildStructuredTestExportModel("incident", fixtureSnapshotID, time.Date(2026, 5, 23, 12, 0, 0, 0, time.UTC), "boundary", ReleaseScopeInternalReview, nil, fields)
	if err != nil {
		t.Fatal(err)
	}
	profile := RedactionProfile{ProfileStatus: "adopted_current", AllowedReleaseScopes: []string{ReleaseScopeInternalDraft, ReleaseScopeInternalReview, ReleaseScopeExternal},
		SchemaID: RedactionProfileSchemaID, ProfileID: "test.tokens", Version: "1", DefaultAction: RedactionActionSpec{Type: ActionMask, TokenBacked: true},
		AllowedDisclosurePartitionRefs: []string{"party:c", "internal_only"}}
	fixtureAllowContainerRules(t, &profile, model.SnapshotContent)
	run := func(model ExportModel, profile RedactionProfile) RedactionResult {
		t.Helper()
		fixtureAllowContainerRules(t, &profile, model.SnapshotContent)
		digest, err := ValidateRedactionProfile(profile)
		if err != nil {
			t.Fatal(err)
		}
		result, err := RedactExportModel(model, fixtureSourceFactsForContent(model.SnapshotContent), profile, digest, sourceSHA, ReleaseScopeInternalReview, nil)
		if err != nil {
			t.Fatal(err)
		}
		return result
	}
	values := func(result RedactionResult) map[string]any {
		out := map[string]any{}
		for _, field := range result.Model.Fields {
			out[fixtureSourcesByModelPath(t, model.SnapshotContent, "/hosts/a/display_name", "/hosts/z/display_name", "/hosts_alias/a/dns_name", "/identities/b/display_name", "/parties/c/display_name")[field.Path]] = field.Value
		}
		return out
	}
	t.Run("container drop precedes token allocation", func(t *testing.T) {
		dropped := profile
		dropped.Rules = append([]RedactionRule{}, profile.Rules...)
		section := "$.sections['sec-0001']"
		for i := range dropped.Rules {
			if dropped.Rules[i].Path != nil && *dropped.Rules[i].Path == section {
				dropped.Rules[i].Action = RedactionActionSpec{Type: ActionDrop}
			}
		}
		result := run(model, dropped)
		if !slices.Contains(result.Manifest.ContainerDecisions[0].DisclosurePartitionRefs, "internal_only") {
			t.Fatal("container evidence lost subject-derived restrictions")
		}
		if len(result.Model.Fields) != 0 || result.TokenManifest != nil || result.RevealMap != nil || len(result.Manifest.Entries) != 0 {
			t.Fatal("suppressed subject allocated a token or reveal value")
		}
	})
	first := run(model, profile)
	want := map[string]any{"/hosts/a/display_name": "HOST-0001", "/hosts_alias/a/dns_name": "HOST-0001", "/hosts/z/display_name": "HOST-0002", "/identities/b/display_name": "IDENTITY-0001", "/parties/c/display_name": "PARTY-0001"}
	if !reflect.DeepEqual(values(first), want) || len(first.TokenManifest.Entries) != 4 {
		t.Fatal("tokens must use canonical per-family subject ordinals", values(first))
	}
	if bytes.Contains(first.TokenManifestJSON, []byte("secret-")) || bytes.Contains(first.ManifestJSON, []byte("secret-")) {
		t.Fatal("public token evidence retained a raw display value")
	}
	if len(first.RevealMap.Entries) != 4 {
		t.Fatal("reveal entries must be per subject, not per field")
	}
	for _, entry := range first.Manifest.Entries {
		if !validExactExportModelPath(entry.Path) || entry.TokenID == nil {
			t.Fatal("token use lacks canonical rule binding")
		}
	}
	// Reordering model arrays and replacing display values cannot affect identity
	// order. A separate admission changes token IDs but not the ordinal labels.
	reordered := model
	reordered.Subjects = append([]TokenizableSubject(nil), model.Subjects...)
	for left, right := 0, len(reordered.Subjects)-1; left < right; left, right = left+1, right-1 {
		reordered.Subjects[left], reordered.Subjects[right] = reordered.Subjects[right], reordered.Subjects[left]
	}
	if got := run(reordered, profile); !bytes.Equal(got.TokenManifestJSON, first.TokenManifestJSON) {
		t.Fatal("subject traversal order changed token artifacts")
	}
	previewID := *model.ReleaseID
	reordered.RenderIdentity = RenderIdentity{PreviewAttemptID: &previewID, RenderAdmittedAt: model.RenderAdmittedAt}
	preview := run(reordered, profile)
	if !reflect.DeepEqual(values(preview), want) || preview.TokenManifest.Entries[0].TokenID == first.TokenManifest.Entries[0].TokenID {
		t.Fatal("token display labels and admission identities were conflated")
	}
	// Only subjects receiving a selected substitution consume an ordinal.
	dropped := profile
	path, alias := "$.sections['sec-0001'].blocks['sec-0001.b0001'].fields['display_name']", "$.sections['sec-0001'].blocks['sec-0001.b0003'].fields['dns_name']"
	dropped.Rules = []RedactionRule{
		{RuleID: "drop-host", Path: &path, Action: RedactionActionSpec{Type: ActionDrop}},
		{RuleID: "drop-host-alias", Path: &alias, Action: RedactionActionSpec{Type: ActionDrop}},
	}
	withoutFirst := run(model, dropped)
	if values(withoutFirst)["/hosts/z/display_name"] != "HOST-0001" || len(withoutFirst.TokenManifest.Entries) != 3 {
		t.Fatal("dropped subjects consumed display ordinals")
	}
	stub := profile
	stub.DefaultAction.Type = ActionStub
	if !reflect.DeepEqual(values(run(model, stub)), want) {
		t.Fatal("token-backed stub changed the subject namespace")
	}
	bound, boundSHA, err := ResolveRedactionProfile(TokenizedRedactionProfileID, DefaultRedactionProfileVersion, nil, model.SnapshotContent)
	if err != nil || boundSHA == "" {
		t.Fatal("snapshot profile binding failed", err)
	}
	if got := values(run(model, bound)); got["/hosts/z/display_name"] != "SUBJECT-0002" {
		t.Fatal("built-in profile did not materialize explicit subject rules", got)
	}
	_, err = RedactExportModel(model, fixtureSourceFactsForContent(model.SnapshotContent), bound, "altered-profile", sourceSHA, ReleaseScopeInternalReview, nil)
	if !errors.Is(err, ErrInvalidRedactionProfile) {
		t.Fatal("unbound profile digest accepted")
	}
	profile.NeutralTokenFamily = true
	neutral := values(run(model, profile))
	if neutral["/hosts/a/display_name"] != "SUBJECT-0001" || neutral["/hosts/z/display_name"] != "SUBJECT-0002" || neutral["/identities/b/display_name"] != "SUBJECT-0003" || neutral["/parties/c/display_name"] != "SUBJECT-0004" {
		t.Fatal("neutral ordinals must share one sorted namespace", neutral)
	}
	profile.NeutralTokenFamily = false
	profile.DefaultAction.TokenBacked = false
	literal := run(model, profile)
	if literal.TokenManifest != nil || literal.RevealMap != nil {
		t.Fatal("literal masks created implicit token artifacts")
	}
	for _, value := range values(literal) {
		if value != DefaultMaskText {
			t.Fatal("literal mask was silently tokenized")
		}
	}
	profile.DefaultAction.TokenBacked = true
	t.Run("source has no display value", func(t *testing.T) {
		withoutDisplay, sourceDigest, err := buildStructuredTestExportModel("incident", fixtureSnapshotID, model.SnapshotAt, "boundary", ReleaseScopeInternalReview, nil,
			[]ExportField{{Path: "/hosts/no-label", SourceFamily: "host", ContentClass: ContentClassSourceEvidence, Value: map[string]any{"hostname": "private-hostname"}}})
		if err != nil {
			t.Fatal(err)
		}
		profile := profile
		profile.Rules = nil
		fixtureAllowContainerRules(t, &profile, withoutDisplay.SnapshotContent)
		digest, err := ValidateRedactionProfile(profile)
		if err != nil {
			t.Fatal(err)
		}
		_, err = RedactExportModel(withoutDisplay, fixtureSourceFactsForContent(withoutDisplay.SnapshotContent), profile, digest, sourceDigest, ReleaseScopeInternalReview, nil)
		var failure *renderValidationError
		if !errors.As(err, &failure) || failure.FailureCode != "token_manifest_invalid" || failure.ReasonCode != "token_subject_not_stable" {
			t.Fatal("tokenization invented a missing display value", err)
		}
	})
	for name, change := range map[string]func(*ExportModel){
		"missing subject": func(m *ExportModel) { m.Subjects = nil },
		"missing display": func(m *ExportModel) {
			m.Records = append([]ReportingRecordSummary(nil), m.Records...)
			m.Subjects = append([]TokenizableSubject(nil), m.Subjects...)
			for i := range m.Records {
				m.Records[i].DisplayName = nil
			}
			for i := range m.Subjects {
				m.Subjects[i].CanonicalDisplayValueSHA256, _ = subjectDisplayValueDigest(m.Subjects[i].StableSubjectRef, nil)
			}
		},
		"ambiguous subject": func(m *ExportModel) {
			m.Subjects = append(append([]TokenizableSubject(nil), m.Subjects...), m.Subjects[0])
		},
		"unbound source path": func(m *ExportModel) { m.Records = nil },
		"invalid stable identity": func(m *ExportModel) {
			m.Subjects = append([]TokenizableSubject(nil), m.Subjects...)
			m.Subjects[0].StableSubjectRef = "host:legacy"
		},
		"cross snapshot": func(m *ExportModel) { m.SnapshotID = "different" },
	} {
		t.Run(name, func(t *testing.T) {
			candidate := model
			change(&candidate)
			digest, err := ValidateRedactionProfile(profile)
			if err != nil {
				t.Fatal(err)
			}
			_, err = RedactExportModel(candidate, fixtureSourceFactsForContent(candidate.SnapshotContent), profile, digest, sourceSHA, ReleaseScopeInternalReview, nil)
			wantCode, wantReason := "token_manifest_invalid", "token_subject_not_unique"
			switch name {
			case "invalid stable identity", "missing subject":
				wantReason = "token_subject_not_stable"
			case "missing display":
				// The unchanged source still supplies display names. Altering
				// their summaries is frozen-model tampering before tokenization.
				wantCode, wantReason = "export_model_invalid", "invalid_scalar_value"
			case "cross snapshot":
				wantCode, wantReason = "export_model_invalid", "snapshot_ref_mismatch"
			case "unbound source path":
				wantCode, wantReason = "export_model_invalid", "dangling_source_ref"
			}
			var failure *renderValidationError
			if !errors.As(err, &failure) || failure.FailureCode != wantCode || failure.ReasonCode != wantReason {
				t.Fatal("invalid subject binding was not rejected at its owning boundary", err)
			}
		})
	}
	for _, action := range []string{ActionAllow, ActionDrop, ActionTruncate, ActionHash} {
		candidate := profile
		candidate.DefaultAction.Type = action
		if _, err := ValidateRedactionProfile(candidate); !errors.Is(err, ErrInvalidRedactionProfile) {
			t.Fatal("invalid token-backed action admitted", action, err)
		}
	}
	duplicateRules := profile
	duplicateRules.Rules = []RedactionRule{
		{RuleID: "same", Path: &path, Action: RedactionActionSpec{Type: ActionDrop}},
		{RuleID: "same", Path: &alias, Action: RedactionActionSpec{Type: ActionDrop}},
	}
	if _, err := ValidateRedactionProfile(duplicateRules); !errors.Is(err, ErrInvalidRedactionProfile) {
		t.Fatal("duplicate rule identity admitted")
	}
	candidate := profile
	literalText := "replacement"
	candidate.DefaultAction.ReplacementText = &literalText
	if _, err := ValidateRedactionProfile(candidate); !errors.Is(err, ErrInvalidRedactionProfile) {
		t.Fatal("token and literal parameters must not be ambiguous")
	}
}

func assertUnresolvedMentionReveal(t *testing.T) {
	t.Helper()
	model := fixtureStructuredRenderModel(t, []ExportField{{
		Path: "/entity_mentions/mention-1", SourceFamily: "entity_mention", SourceRecordID: "mention-1", SourceRecordType: "entity_mention",
		ContentClass: ContentClassSourceEvidence, Value: map[string]any{"entity_type": "host", "raw_text": "Original Host", "normalized_text": "normalized-host", "source_record_id": "origin", "source_field_key": "evidence.description", "resolution_status": "unresolved", "resolved_record_id": nil},
	}, {Path: "/evidence/origin", SourceFamily: "evidence", SourceRecordID: "origin", SourceRecordType: "evidence", ContentClass: ContentClassSourceEvidence, Value: "origin text"}})
	profile, profileSHA, err := ResolveRedactionProfile(TokenizedRedactionProfileID, DefaultRedactionProfileVersion, nil, model.SnapshotContent)
	if err != nil {
		t.Fatal(err)
	}
	raw, _ := canonicalJSON(model)
	result, err := RedactExportModel(model, fixtureSourceFactsForContent(model.SnapshotContent), profile, profileSHA, reportingObjectDigest(model.SchemaID, raw), ReleaseScopeInternalReview, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.RevealMap.Entries) != 1 {
		t.Fatal("unresolved subject produced multiple private values")
	}
	entry := result.RevealMap.Entries[0]
	if entry.SourceRecordID != nil || entry.EntityMentionID == nil || *entry.EntityMentionID != "mention-1" || entry.CanonicalDisplayValue != "Original Host" ||
		bytes.Contains(result.RevealMapJSON, []byte("normalized-host")) || bytes.Contains(result.TokenManifestJSON, []byte("Original Host")) {
		t.Fatal("mention reveal changed source identity or display provenance")
	}
}
