package reporting

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"testing"
)

func testContainerRedaction(t *testing.T) {
	model := fixtureStructuredRenderModel(t, []ExportField{
		{Path: "/incident/private", ContentClass: ContentClassSourceEvidence, Value: "excluded private value", DisclosurePartitionRefs: []string{"blocked"}},
		{Path: "/incident/retained", ContentClass: ContentClassWorkingMaterial, Value: "retained value", DisclosurePartitionRefs: []string{"public"}},
	})
	census, err := redactionTreeCandidates(model.SnapshotContent)
	if err != nil || len(census) != 5 {
		t.Fatal("unexpected canonical fixture census", err, census)
	}
	section, block, field := census[0].path, census[1].path, census[2].path
	profileFor := func(t *testing.T) RedactionProfile {
		t.Helper()
		profile, _, err := ResolveRedactionProfile(InternalRedactionProfileID, DefaultRedactionProfileVersion, nil, model.SnapshotContent)
		if err != nil {
			t.Fatal(err)
		}
		return profile
	}
	set := func(profile *RedactionProfile, path string, action RedactionActionSpec) {
		for i := range profile.Rules {
			if profile.Rules[i].Path != nil && *profile.Rules[i].Path == path {
				profile.Rules[i].Action = action
				return
			}
		}
		t.Fatal("fixture path rule missing", path)
	}
	render := func(profile RedactionProfile) (RedactionResult, error) {
		digest, err := ValidateRedactionProfile(profile)
		if err != nil {
			return RedactionResult{}, err
		}
		raw, _ := canonicalJSON(model)
		return RedactExportModel(model, fixtureSourceFactsForContent(model.SnapshotContent), profile, digest, reportingObjectDigest(model.SchemaID, raw), ReleaseScopeInternalReview, nil)
	}
	t.Run("block drop excludes values before disclosure and all projections", func(t *testing.T) {
		profile := profileFor(t)
		set(&profile, block, RedactionActionSpec{Type: ActionDrop})
		result, err := render(profile)
		if err != nil {
			t.Fatal(err)
		}
		if len(result.Model.Fields) != 1 || len(result.Manifest.Entries) != 1 || len(result.Model.Sections) != 1 || len(result.Model.Sections[0].Blocks) != 1 || len(result.Manifest.StructuralRemovals) != 0 {
			t.Fatal("drop did not exclude exactly its canonical subtree")
		}
		if result.Manifest.ContainerDecisions[1].Outcome != "dropped" || !sameReportingValue(result.Manifest.ContainerDecisions[1].SuppressedPaths, []string{field}) {
			t.Fatal("drop lost exact rule/subtree evidence")
		}
		raw, _ := canonicalJSON(result.Model)
		if bytes.Contains(raw, []byte("excluded private value")) || bytes.Contains(result.ManifestJSON, []byte("excluded private value")) {
			t.Fatal("excluded value reached derived output")
		}
		if result.Model.Sections[0].Blocks[0].BlockOrdinal != 1 {
			t.Fatal("retained block not renumbered")
		}
		for _, key := range []string{"content_class", "suppressed_paths", "selected_rule_trace", "unexpected"} {
			var object map[string]any
			if err := json.Unmarshal(result.ManifestJSON, &object); err != nil {
				t.Fatal(err)
			}
			decision := object["container_decisions"].([]any)[0].(map[string]any)
			if key == "unexpected" {
				decision[key] = true
			} else {
				delete(decision, key)
			}
			raw, _ := canonicalJSON(object)
			if _, _, err := decodePortableObject[RedactionManifest](raw, reportingObjectDigest(RedactionManifestSchemaID, raw)); err == nil {
				t.Fatal("unclosed container evidence admitted", key)
			}
		}
		for name, mutate := range map[string]func(*RedactionManifest){
			"missing container census": func(m *RedactionManifest) { m.ContainerDecisions = nil },
			"missing ancestor":         func(m *RedactionManifest) { m.ContainerDecisions = m.ContainerDecisions[1:] },
			"invented section class": func(m *RedactionManifest) {
				value := ContentClassSourceEvidence
				m.ContainerDecisions[0].ContentClass = &value
			},
			"wrong block class": func(m *RedactionManifest) { m.ContainerDecisions[1].ContentClass = nil },
			"retained child suppression": func(m *RedactionManifest) {
				m.ContainerDecisions[1].SuppressedPaths = append(m.ContainerDecisions[1].SuppressedPaths, m.Entries[0].Path)
			},
			"stale retained partitions": func(m *RedactionManifest) { m.ContainerDecisions[0].DisclosurePartitionRefs = []string{"blocked"} },
			"transformed container":     func(m *RedactionManifest) { m.ContainerDecisions[1].Action = ActionMask },
			"suppressed path outside tree": func(m *RedactionManifest) {
				m.ContainerDecisions[1].SuppressedPaths = []string{block + ".unknown['x']"}
			},
		} {
			t.Run(name, func(t *testing.T) {
				var candidate RedactionManifest
				raw, _ := json.Marshal(result.Manifest)
				err := json.Unmarshal(raw, &candidate)
				if err != nil {
					t.Fatal(err)
				}
				mutate(&candidate)
				if validateRedactionManifest(candidate) == nil {
					t.Fatal("tampered container evidence accepted")
				}
			})
		}
	})
	t.Run("section drop does not manufacture child rule decisions", func(t *testing.T) {
		profile := profileFor(t)
		set(&profile, section, RedactionActionSpec{Type: ActionDrop})
		result, err := render(profile)
		if err != nil {
			t.Fatal(err)
		}
		if len(result.Manifest.Entries) != 0 || len(result.Manifest.ContainerDecisions) != 1 || len(result.Manifest.ContainerDecisions[0].SuppressedPaths) != 4 || len(result.Manifest.StructuralRemovals) != 0 || len(result.Model.Fields) != 0 || len(result.Model.Sections) != 0 || len(result.Model.Records) != 0 || len(result.Model.SourceIndex) != 0 {
			t.Fatal("section drop left content or invented child actions")
		}
		missing := block + ".fields['missing']"
		profile.Rules = append(profile.Rules, RedactionRule{RuleID: "missing-under-drop", Path: &missing, Action: RedactionActionSpec{Type: ActionAllow}})
		if _, err := render(profile); err == nil {
			t.Fatal("parent drop hid unresolved exact selector")
		}
	})
	t.Run("class and default precedence remain explicit", func(t *testing.T) {
		profile := profileFor(t)
		rules := []RedactionRule{}
		for _, rule := range profile.Rules {
			if rule.Path == nil || *rule.Path != block {
				rules = append(rules, rule)
			}
		}
		class := ContentClassSourceEvidence
		profile.Rules = append(rules, RedactionRule{RuleID: "class-drop", ContentClass: &class, Action: RedactionActionSpec{Type: ActionDrop}})
		result, err := render(profile)
		if err != nil {
			t.Fatal(err)
		}
		if result.Manifest.ContainerDecisions[1].SelectedRuleTrace.SelectorKind != "content_class" {
			t.Fatal("class rule not selected")
		}
		profile.Rules = nil
		profile.DefaultAction = RedactionActionSpec{Type: ActionDrop}
		result, err = render(profile)
		if err != nil {
			t.Fatal(err)
		}
		if len(result.Manifest.ContainerDecisions) != 1 || result.Manifest.ContainerDecisions[0].SelectedRuleTrace.SelectorKind != "profile_default" {
			t.Fatal("classless section default not selected")
		}
	})
	t.Run("nested drop records complete ancestry and no descendant token use", func(t *testing.T) {
		field := ReportingField{FieldKey: "value", SourceRefs: []string{"source"}}
		nested := SnapshotContent{Sections: []ReportingSection{{SectionID: "section", SectionKind: "appendix", Blocks: []ReportingBlock{{BlockID: "parent", ContentClass: ContentClassSourceEvidence, DisclosurePartitionRefs: []string{"blocked"}, Children: []ReportingBlock{{BlockID: "child'quoted", ContentClass: ContentClassSourceEvidence, Fields: []ReportingField{field}}}}}}}}
		parent := "$.sections['section'].blocks['parent']"
		profile := internalRedactionProfile()
		profile.Rules = []RedactionRule{{RuleID: "drop-parent", Path: &parent, Action: RedactionActionSpec{Type: ActionDrop}}}
		digest, _ := ValidateRedactionProfile(profile)
		view, _, _, err := BuildRedactionProfileView(profile, digest)
		if err != nil {
			t.Fatal(err)
		}
		fields, decisions, err := selectContainerRedactions(context.Background(), nested, []ExportField{{Path: "source"}}, view)
		if err != nil || len(fields) != 0 || len(decisions) != 2 || len(decisions[1].SuppressedPaths) != 2 {
			t.Fatal("nested suppression incomplete", err)
		}
		entries := structuralDecisionEntries(nil, decisions)
		removals, err := structuralRedactionRemovals(nested, entries)
		if err != nil || len(removals) != 1 || !sameReportingValue(removals[0].CausalPaths, []string{parent}) || validateStructuralRemovals(entries, removals) != nil {
			t.Fatal("parent drop invented descendant causes", err)
		}
		retained := explicitlyRetainedTree(nested, decisions)
		if len(retained.Sections[0].Blocks) != 0 {
			t.Fatal("nested excluded content retained")
		}
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		if _, _, err := selectContainerRedactions(ctx, nested, nil, view); !errors.Is(err, context.Canceled) {
			t.Fatal("canceled selection succeeded", err)
		}
	})
	for _, action := range []RedactionActionSpec{{Type: ActionAllow}, {Type: ActionMask}, {Type: ActionStub}, {Type: ActionTruncate, MaxChars: copyProfileValue(new(2))}, {Type: ActionMask, TokenBacked: true}} {
		t.Run("reject unresolved "+action.Type, func(t *testing.T) {
			profile := profileFor(t)
			set(&profile, block, action)
			_, err := render(profile)
			var failure *renderValidationError
			if err == nil || !errors.As(err, &failure) {
				t.Fatal("unresolved container output accepted", err)
			}
			want := "redaction_action_unresolved"
			if action.Type == ActionAllow {
				want = "disclosure_partition_unresolved"
			}
			if failure.ReasonCode != want {
				t.Fatal("wrong closed failure", failure.ReasonCode, want)
			}
		})
	}
}
