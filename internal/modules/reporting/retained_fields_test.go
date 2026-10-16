package reporting

import (
	"errors"
	"testing"
)

func testRetainedFieldActions(t *testing.T) {
	for _, sourceState := range []string{"present", "null", "missing", "unavailable", "withheld", "superseded", "derived", "deleted"} {
		for _, outcome := range []string{"unchanged", "masked", "truncated", "stubbed", "tokenized", "dropped", "blocked"} {
			for _, precedence := range []string{"profile_default", "path"} {
				t.Run(sourceState+"/"+outcome+"/"+precedence, func(t *testing.T) {
					var value any = "safe"
					if outcome == "unchanged" && (sourceState == "null" || sourceState == "missing" || sourceState == "unavailable") {
						value = nil
					}
					valid := false
					switch sourceState {
					case "present", "superseded", "derived":
						valid = outcome != "dropped" && outcome != "blocked"
					case "null":
						valid = outcome == "unchanged"
					case "missing":
						valid = outcome == "unchanged" && precedence == "profile_default"
					case "unavailable":
						valid = outcome == "stubbed" || outcome == "unchanged" && precedence == "profile_default"
					case "withheld":
						valid = outcome == "masked" || outcome == "stubbed" || outcome == "tokenized"
					}
					field := ReportingField{SchemaID: "cartulary.reporting_field.v1", FieldKey: "value", FieldOrdinal: 1, SourceValueState: sourceState,
						Value: "original", SourceRefs: []string{}, SupportRefs: []string{}, DisclosurePartitionRefs: []string{"public"}}
					result, err := retainedReportingField(field, value, outcome, precedence, true)
					if (err == nil) != valid {
						t.Fatalf("retained-state admission: valid=%v err=%v", valid, err)
					}
					if err != nil {
						var failure *renderValidationError
						if !errors.As(err, &failure) || failure.FailureCode != "redaction_manifest_invalid" || failure.ReasonCode != "redaction_action_unresolved" || result.SchemaID != "" {
							t.Fatal("invalid action returned partial data or wrong failure", err)
						}
					} else if result.SourceValueState != sourceState || result.RawValueSHA256 != nil {
						t.Fatal("selected action changed source state")
					}
					if field.Value != "original" {
						t.Fatal("source mutated")
					}
				})
			}
		}
	}

	model := fixtureStructuredRenderModel(t, []ExportField{
		{Path: "/incident/nullable", ContentClass: ContentClassSourceEvidence, Value: nil, DisclosurePartitionRefs: []string{"public"}},
		{Path: "/incident/present", ContentClass: ContentClassSourceEvidence, Value: "present", DisclosurePartitionRefs: []string{"public"}},
	})
	for _, action := range []RedactionActionSpec{{Type: ActionMask}, {Type: ActionStub}, {Type: ActionTruncate, MaxChars: func() *int { n := 4; return &n }()}} {
		profile := internalRedactionProfile()
		profile.DefaultAction = action
		digest, err := ValidateRedactionProfile(profile)
		if err != nil {
			t.Fatal(err)
		}
		result, err := RedactExportModel(model, fixtureSourceFactsForContent(model.SnapshotContent), profile, digest, "", ReleaseScopeInternalReview, nil)
		var failure *renderValidationError
		if !errors.As(err, &failure) || failure.ReasonCode != "redaction_action_unresolved" || len(result.ManifestJSON) != 0 || len(result.Model.Fields) != 0 {
			t.Fatal("production action on null escaped retained-state validation", action.Type, err)
		}
	}
	for _, id := range []string{ExternalRedactionProfileID, TokenizedRedactionProfileID} {
		scope := ReleaseScopeInternalReview
		if id == ExternalRedactionProfileID {
			scope = ReleaseScopeExternal
		}
		profile, digest, err := ResolveRedactionProfile(id, DefaultRedactionProfileVersion, nil, model.SnapshotContent)
		if err != nil {
			t.Fatal(err)
		}
		result, err := RedactExportModel(model, fixtureSourceFactsForContent(model.SnapshotContent), profile, digest, "", scope, nil)
		if err != nil {
			t.Fatal(err)
		}
		paths := fixtureModelPaths(t, model.SnapshotContent, "/incident/nullable", "/incident/present")
		found := false
		for _, entry := range result.Manifest.Entries {
			if entry.Path == paths["/incident/nullable"] {
				found = entry.Action == ActionAllow && entry.Outcome == "unchanged" && entry.TokenID == nil && entry.RulePrecedence == "path"
			}
		}
		if !found {
			t.Fatal("null did not use its admitted explicit rule")
		}
		for _, field := range result.Model.Fields {
			if field.Path == paths["/incident/nullable"] && field.Value != nil {
				t.Fatal("null was rendered as text")
			}
		}
	}
}
