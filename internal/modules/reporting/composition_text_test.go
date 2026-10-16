package reporting

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

func compositionTextFixture(t *testing.T, texts []compositionAuthoredText, ops []compositionDeckOp) renderInput {
	t.Helper()
	model := fixtureStructuredRenderModel(t, []ExportField{{Path: "/hosts/host-one", SourceFamily: "host", ContentClass: ContentClassSourceEvidence, Value: map[string]any{"display_name": "Sensitive host"}}})
	contract, _ := ResolveTemplateContract(DefaultTemplateID, DefaultTemplateVersion)
	raw, err := canonicalJSON(map[string]any{"schema_id": "cartulary.report_composition.v1", "authored_texts": texts, "deck_ops": ops, "diagram_decls": []any{}})
	if err != nil {
		t.Fatal(err)
	}
	raw = sealedReportingCompositionFixture(t, raw, model.IncidentID, contract)
	encoded, _ := canonicalJSON(model)
	request := CreateReleaseRequest{ReleaseScope: ReleaseScopeInternalReview, OutputKind: OutputKindSlidev, OutputOptions: fixtureOutputOptions(OutputKindSlidev, ReleaseScopeInternalReview), RedactionProfileID: InternalRedactionProfileID, RedactionProfileVersion: DefaultRedactionProfileVersion, CompositionJSON: raw}
	return renderCandidateFixture(request, contract, model, reportingObjectDigest(ExportModelSchemaID, encoded))
}

func setCompositionFixtureProfile(t *testing.T, input *renderInput, profile RedactionProfile) {
	t.Helper()
	fixtureAllowContainerRules(t, &profile, input.ExportModel.SnapshotContent)
	digest, err := ValidateRedactionProfile(profile)
	if err != nil {
		t.Fatal(err)
	}
	input.RedactionProfile, input.RedactionProfileSHA256 = profile, digest
	input.RedactionProfileID, input.RedactionProfileVersion = profile.ProfileID, profile.Version
}

func testCompositionTextRedaction(t *testing.T) {
	subject := "{{subject:subject:host:record:host-one}}"
	anchor := &compositionSectionAnchor{AnchorKind: "section_anchor", DeclarationID: "snapshot"}
	texts := []compositionAuthoredText{
		{AuthoredTextID: "title", TextRole: "title_override", Body: "Review " + subject, DisclosurePartitionRef: "public"},
		{AuthoredTextID: "notes", TextRole: "speaker_notes", Body: "Discuss\n" + subject, DisclosurePartitionRef: "internal_only"},
		{AuthoredTextID: "prose", TextRole: "authored_text", Body: "Context " + subject, DisclosurePartitionRef: "public"},
	}
	ops := []compositionDeckOp{
		{OpID: "title", OpKind: "override_title", OnUnresolved: "fail", Payload: compositionOpPayload{SectionAnchor: anchor, AuthoredTextRef: "title"}},
		{OpID: "notes", OpKind: "set_speaker_notes", OnUnresolved: "fail", Payload: compositionOpPayload{SectionAnchor: anchor, AuthoredTextRef: "notes"}},
		{OpID: "prose", OpKind: "insert_authored_block", OnUnresolved: "fail", Payload: compositionOpPayload{BlockAnchor: &compositionBlockAnchor{AnchorKind: "block_anchor", SectionAnchor: *anchor, BlockKind: "paragraph"}, Position: "after", AuthoredTextRef: "prose"}},
	}
	input := compositionTextFixture(t, texts, ops)
	class, sourceClass, literal := ContentClassPresentationText, ContentClassSourceEvidence, "Safe presentation"
	for _, action := range []string{ActionAllow, ActionMask, ActionDrop} {
		t.Run(action, func(t *testing.T) {
			candidate := input
			profile := input.RedactionProfile
			profile.Rules = []RedactionRule{{RuleID: "presentation", ContentClass: &class, Action: RedactionActionSpec{Type: action}}}
			if action == ActionMask {
				profile.Rules[0].Action.ReplacementText = &literal
			}
			setCompositionFixtureProfile(t, &candidate, profile)
			before, _ := canonicalJSON(candidate)
			rendered, _, err := fixtureRenderer().Render(context.Background(), candidate)
			if err != nil {
				t.Fatal(err)
			}
			if err := validateRenderCandidateBinding(candidate, rendered); err != nil {
				t.Fatal(err)
			}
			after, _ := canonicalJSON(candidate)
			if !bytes.Equal(before, after) {
				t.Fatal("redaction mutated admission")
			}
			deck, err := deriveDeckModel(candidate.TemplateContract, rendered.Redaction.Model, renderOutputOptions{}, candidate.CompositionJSON)
			if err != nil {
				t.Fatal(err)
			}
			if action == ActionDrop {
				if len(rendered.Redaction.Model.PresentationTexts) != 0 || len(deck.Slides[0].Blocks) != 1 || deck.Slides[0].SpeakerNotes != nil || deck.Slides[0].Title == "Review Sensitive host" {
					t.Fatal("dropped prose reached the deck")
				}
			} else {
				wantTitle, wantProse := "Review Sensitive host", "Context Sensitive host"
				if action == ActionMask {
					wantTitle, wantProse = literal, literal
				}
				if len(rendered.Redaction.Model.PresentationTexts) != 3 || deck.Slides[0].Title != wantTitle || deck.Slides[0].Blocks[1].Fields[0].Value != wantProse {
					t.Fatal("presentation projection", deck)
				}
				if len(deck.Slides[0].Blocks[1].SourceRefs) == 0 {
					t.Fatal("subject provenance lost")
				}
				if action == ActionMask && deck.Slides[0].Blocks[1].Fields[0].RedactedValueState != "masked" {
					t.Fatal("deck erased redaction state")
				}
			}
			count := 0
			for _, entry := range rendered.Redaction.Manifest.Entries {
				if strings.HasPrefix(entry.Path, "$.presentation_texts[") {
					count++
					if entry.RuleID != "presentation" || entry.Action != action {
						t.Fatal("authored decision missing")
					}
				}
			}
			if count != 3 {
				t.Fatal("missing authored manifest entries", count)
			}
		})
	}
	t.Run("placeholder uses selected token", func(t *testing.T) {
		candidate := input
		profile := candidate.RedactionProfile
		profile.Rules = []RedactionRule{{RuleID: "source-token", ContentClass: &sourceClass, Action: RedactionActionSpec{Type: ActionMask, TokenBacked: true}}}
		setCompositionFixtureProfile(t, &candidate, profile)
		result, err := redactRenderInput(context.Background(), candidate)
		if err != nil {
			t.Fatal(err)
		}
		encoded, _ := canonicalJSON(result.Model)
		if bytes.Contains(encoded, []byte("Sensitive host")) || !bytes.Contains(encoded, []byte("Review HOST-0001")) {
			t.Fatal("placeholder bypassed selected token")
		}
		if len(result.TokenManifest.Entries) != 1 {
			t.Fatal("placeholder duplicated token identity")
		}
	})
	t.Run("whole text token requires one bound subject", func(t *testing.T) {
		candidate := input
		profile := candidate.RedactionProfile
		profile.Rules = []RedactionRule{{RuleID: "text-token", ContentClass: &class, Action: RedactionActionSpec{Type: ActionMask, TokenBacked: true}}}
		setCompositionFixtureProfile(t, &candidate, profile)
		result, err := redactRenderInput(context.Background(), candidate)
		if err != nil {
			t.Fatal(err)
		}
		for _, text := range result.Model.PresentationTexts {
			if text.Body.Value != "HOST-0001" {
				t.Fatal("whole-field token mismatch")
			}
		}
		if len(result.TokenManifest.Entries) != 1 || len(result.RevealMap.Entries) != 1 {
			t.Fatal("one subject must retain one identity")
		}
	})
	for _, tc := range []struct {
		name, reason string
		change       func(*renderInput)
	}{
		{"filtered subject", "authored_subject_ref_unresolved", func(in *renderInput) {
			profile := in.RedactionProfile
			profile.Rules = []RedactionRule{{RuleID: "drop-subject", ContentClass: &sourceClass, Action: RedactionActionSpec{Type: ActionDrop}}}
			setCompositionFixtureProfile(t, in, profile)
		}},
		{"partition permission", "disclosure_partition_unresolved", func(in *renderInput) {
			profile := in.RedactionProfile
			profile.AllowedDisclosurePartitionRefs = []string{"internal_only"}
			setCompositionFixtureProfile(t, in, profile)
		}},
		{"role limit", "authored_title_limit_exceeded", func(in *renderInput) {
			in.TemplateContract.DeclaredLimits = map[string]json.RawMessage{"composition.authored_title_chars": json.RawMessage(`5`)}
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			candidate := input
			tc.change(&candidate)
			result, err := redactRenderInput(context.Background(), candidate)
			var failure *renderValidationError
			if !errors.As(err, &failure) || failure.ReasonCode != tc.reason || len(result.ManifestJSON) != 0 {
				t.Fatal("unsafe authored admission", err)
			}
		})
	}
	t.Run("malformed and unknown placeholders", func(t *testing.T) {
		for _, body := range []string{"{{subject:", "{{subject:}}", "{{subject:not-selected}}", "{{subject:subject:host:record:unknown}}"} {
			candidate := compositionTextFixture(t, []compositionAuthoredText{{AuthoredTextID: "text", TextRole: "authored_text", Body: body, DisclosurePartitionRef: "public"}}, []compositionDeckOp{})
			if _, err := redactRenderInput(context.Background(), candidate); err == nil {
				t.Fatal("invalid placeholder admitted")
			}
		}
	})
	t.Run("rewritten composition cannot reuse a retained decision", func(t *testing.T) {
		result, err := redactRenderInput(context.Background(), input)
		if err != nil {
			t.Fatal(err)
		}
		composition, _ := parseRenderComposition(input.CompositionJSON)
		composition.AuthoredTexts[0].Body = "Changed after redaction"
		if bindCompositionPresentation(composition, result.Model) == nil {
			t.Fatal("changed composition bypassed redaction")
		}
	})
	t.Run("external permission does not allow a forbidden partition", func(t *testing.T) {
		for _, partition := range []string{"public", "internal_only", "party:unknown"} {
			candidate := compositionTextFixture(t, []compositionAuthoredText{{AuthoredTextID: "text", TextRole: "authored_text", Body: "Ordinary presentation", DisclosurePartitionRef: partition}}, []compositionDeckOp{})
			// This policy test intentionally drops all source content; an empty
			// optional section must not mask the authored-text partition result.
			candidate.TemplateContract.Sections[0].Required = false
			candidate.CompositionJSON = sealedReportingCompositionFixture(t, candidate.CompositionJSON, candidate.ExportModel.IncidentID, candidate.TemplateContract)
			candidate.ReleaseScope = ReleaseScopeExternal
			profile := candidate.RedactionProfile
			profile.ProfileID = "test.external.presentation"
			profile.AllowAuthoredPresentationText = true
			profile.AllowedReleaseScopes = []string{ReleaseScopeExternal}
			profile.AllowedDisclosurePartitionRefs = []string{"public"}
			profile.Rules = []RedactionRule{{RuleID: "source-drop", ContentClass: &sourceClass, Action: RedactionActionSpec{Type: ActionDrop}}}
			setCompositionFixtureProfile(t, &candidate, profile)
			result, err := redactRenderInput(context.Background(), candidate)
			if partition == "public" {
				if err != nil || len(result.Model.PresentationTexts) != 1 {
					t.Fatal("permitted public prose", err)
				}
			} else if err == nil || len(result.ManifestJSON) != 0 {
				t.Fatal("profile permission bypassed partition")
			}
		}
	})
	t.Run("substitution and replacement obey role limits", func(t *testing.T) {
		candidate := compositionTextFixture(t, []compositionAuthoredText{{AuthoredTextID: "title", TextRole: "title_override", Body: subject, DisclosurePartitionRef: "public"}}, []compositionDeckOp{})
		candidate.ExportModel = fixtureStructuredRenderModel(t, []ExportField{{Path: "/hosts/host-one", SourceFamily: "host", ContentClass: sourceClass, Value: map[string]any{"display_name": strings.Repeat("x", 121)}}})
		var failure *renderValidationError
		if _, err := redactRenderInput(context.Background(), candidate); !errors.As(err, &failure) || failure.ReasonCode != "authored_title_limit_exceeded" {
			t.Fatal("post-substitution limit", err)
		}
		candidate = input
		profile := candidate.RedactionProfile
		long := strings.Repeat("x", 121)
		profile.Rules = []RedactionRule{{RuleID: "long-mask", ContentClass: &class, Action: RedactionActionSpec{Type: ActionMask, ReplacementText: &long}}}
		setCompositionFixtureProfile(t, &candidate, profile)
		if _, err := redactRenderInput(context.Background(), candidate); !errors.As(err, &failure) || failure.ReasonCode != "authored_title_limit_exceeded" {
			t.Fatal("post-redaction limit", err)
		}
	})
	t.Run("dropped text still resolves its anchor", func(t *testing.T) {
		candidate := compositionTextFixture(t, []compositionAuthoredText{{AuthoredTextID: "title", TextRole: "title_override", Body: "Dropped", DisclosurePartitionRef: "public"}}, []compositionDeckOp{{OpID: "title", OpKind: "override_title", OnUnresolved: "fail", Payload: compositionOpPayload{AuthoredTextRef: "title", SectionAnchor: &compositionSectionAnchor{AnchorKind: "section_anchor", DeclarationID: "absent"}}}})
		profile := candidate.RedactionProfile
		profile.Rules = []RedactionRule{{RuleID: "drop-text", ContentClass: &class, Action: RedactionActionSpec{Type: ActionDrop}}}
		setCompositionFixtureProfile(t, &candidate, profile)
		result, err := redactRenderInput(context.Background(), candidate)
		if err != nil {
			t.Fatal(err)
		}
		var failure *renderValidationError
		if _, err := deriveDeckModel(candidate.TemplateContract, result.Model, renderOutputOptions{}, candidate.CompositionJSON); !errors.As(err, &failure) || failure.ReasonCode != "composition_anchor_unresolved" {
			t.Fatal("redaction concealed invalid anchor", err)
		}
	})
	t.Run("exact text path wins over class", func(t *testing.T) {
		candidate := input
		profile := candidate.RedactionProfile
		path := "$.presentation_texts['title'].body"
		profile.Rules = []RedactionRule{{RuleID: "only-title", Path: &path, Action: RedactionActionSpec{Type: ActionAllow}}, {RuleID: "drop-prose", ContentClass: &class, Action: RedactionActionSpec{Type: ActionDrop}}}
		setCompositionFixtureProfile(t, &candidate, profile)
		if err := validateFrozenRenderProfile(candidate); err != nil {
			t.Fatal("frozen profile lost authored path", err)
		}
		result, err := redactRenderInput(context.Background(), candidate)
		if err != nil || len(result.Model.PresentationTexts) != 1 || result.Model.PresentationTexts[0].AuthoredTextID != "title" {
			t.Fatal("text path precedence", err)
		}
	})
	t.Run("placeholder expansion rejects before building oversized output", func(t *testing.T) {
		calls := 0
		_, err := substituteAuthoredSubjects("{{subject:one}}{{subject:two}}", 120, "authored_title_limit_exceeded", func(string) (string, error) { calls++; return strings.Repeat("x", 121), nil })
		if err == nil || calls != 1 {
			t.Fatal("expansion continued beyond its operative bound")
		}
	})
	t.Run("unbound literal cannot receive a subject token", func(t *testing.T) {
		candidate := compositionTextFixture(t, []compositionAuthoredText{{AuthoredTextID: "text", TextRole: "authored_text", Body: "Literal prose", DisclosurePartitionRef: "public"}}, []compositionDeckOp{})
		profile := candidate.RedactionProfile
		profile.Rules = []RedactionRule{{RuleID: "text-token", ContentClass: &class, Action: RedactionActionSpec{Type: ActionMask, TokenBacked: true}}}
		setCompositionFixtureProfile(t, &candidate, profile)
		var failure *renderValidationError
		if _, err := redactRenderInput(context.Background(), candidate); !errors.As(err, &failure) || failure.ReasonCode != "token_subject_not_unique" {
			t.Fatal("literal acquired an inferred subject", err)
		}
	})
}
