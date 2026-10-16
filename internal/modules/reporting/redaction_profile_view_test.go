package reporting

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func testNormalizedRedactionProfile(t *testing.T) {
	t.Run("exact identifier contract", testProfileIdentifierContract)
	t.Run("exact model paths", testExactExportModelPaths)
	path, class, empty, limit := "$.sections['sec-0001'].blocks['sec-0001.b0001'].fields['value']", ContentClassSourceEvidence, "", 3
	profile := internalRedactionProfile()
	profile.Rules = []RedactionRule{
		{RuleID: "z-path", Path: &path, Action: RedactionActionSpec{Type: ActionMask, ReplacementText: &empty}},
		{RuleID: "a-class", ContentClass: &class, Action: RedactionActionSpec{Type: ActionTruncate, MaxChars: &limit}},
	}
	profile.DefaultAction = RedactionActionSpec{Type: ActionStub}
	build := func(p RedactionProfile) (RedactionProfileView, []byte) {
		t.Helper()
		digest, err := ValidateRedactionProfile(p)
		if err != nil {
			t.Fatal(err)
		}
		view, encoded, sum, err := BuildRedactionProfileView(p, digest)
		if err != nil || sum != reportingObjectDigest(RedactionProfileViewSchemaID, encoded) {
			t.Fatal("profile normalization", err)
		}
		return view, encoded
	}
	view, encoded := build(profile)
	if view.PathRules[0].RuleOrder != 1 || view.ContentClassRules[0].RuleOrder != 2 ||
		view.PathRules[0].LiteralOutput == nil || *view.PathRules[0].LiteralOutput != "" ||
		view.ProfileDefault.LiteralOutput == nil || *view.ProfileDefault.LiteralOutput != DefaultStubText ||
		view.ContentClassRules[0].LiteralOutput != nil || *view.ContentClassRules[0].TruncateMaxScalars != 3 {
		t.Fatal("authored order or action defaults lost")
	}
	var object map[string]json.RawMessage
	if err := json.Unmarshal(encoded, &object); err != nil {
		t.Fatal(err)
	}
	if len(object) != 13 || object["time_bucket_granularity"] != nil || object["redaction_profile_id"] == nil {
		t.Fatal("profile view members/defaults differ", string(encoded))
	}
	trace, _ := canonicalJSON(view.ProfileDefault.SelectedRuleTrace)
	if string(trace) != `{"rule_id":"profile_default","rule_precedence":"profile_default","schema_id":"selected_rule_trace.v1","selector_kind":"profile_default","selector_value":null}` {
		t.Fatal("trace must contain only safe selected-rule metadata", string(trace))
	}
	// The view owns its values, even when the original profile is changed later.
	empty, path, limit = "mutated literal", "/mutated", 1
	again, _ := canonicalJSON(view)
	if !bytes.Equal(encoded, again) {
		t.Fatal("profile view aliases authored storage")
	}
	_, raw := build(internalRedactionProfile())
	if !bytes.Contains(raw, []byte(`"path_rules":[]`)) || !bytes.Contains(raw, []byte(`"content_class_rules":[]`)) {
		t.Fatal("empty rules must be arrays")
	}
	for name, change := range map[string]func(*RedactionProfileView){
		"unknown partition":      func(v *RedactionProfileView) { v.AllowedDisclosurePartitionRefs = []string{"public_summary"} },
		"duplicate partition":    func(v *RedactionProfileView) { v.AllowedDisclosurePartitionRefs = []string{"public", "public"} },
		"partition order":        func(v *RedactionProfileView) { v.AllowedDisclosurePartitionRefs = []string{"internal_only", "public"} },
		"nested party partition": func(v *RedactionProfileView) { v.AllowedDisclosurePartitionRefs = []string{"party:one:two"} },
		"old schema":             func(v *RedactionProfileView) { v.SchemaID = "cartulary.redaction_profile_view.v1" },
		"empty scopes":           func(v *RedactionProfileView) { v.AllowedReleaseScopes = nil },
		"unknown scope":          func(v *RedactionProfileView) { v.AllowedReleaseScopes = []string{"external"} },
		"duplicate scope": func(v *RedactionProfileView) {
			v.AllowedReleaseScopes = []string{ReleaseScopeInternalDraft, ReleaseScopeInternalDraft}
		},
		"scope order": func(v *RedactionProfileView) {
			v.AllowedReleaseScopes = []string{ReleaseScopeExternal, ReleaseScopeInternalDraft}
		},
		"unknown status":             func(v *RedactionProfileView) { v.ProfileStatus = "approved" },
		"nil rules":                  func(v *RedactionProfileView) { v.PathRules = nil },
		"duplicate rule id":          func(v *RedactionProfileView) { v.ContentClassRules[0].RuleID = v.PathRules[0].RuleID },
		"missing literal":            func(v *RedactionProfileView) { v.ProfileDefault.LiteralOutput = nil },
		"source literal on truncate": func(v *RedactionProfileView) { s := "private"; v.ContentClassRules[0].LiteralOutput = &s },
		"token backed truncate":      func(v *RedactionProfileView) { v.ContentClassRules[0].TokenBacked = true },
		"zero order":                 func(v *RedactionProfileView) { v.PathRules[0].RuleOrder = 0 },
		"wrong selector":             func(v *RedactionProfileView) { v.PathRules[0].ContentClassSelector = &class },
		"trace precedence":           func(v *RedactionProfileView) { v.PathRules[0].SelectedRuleTrace.RulePrecedence = "profile_default" },
		"trace selector":             func(v *RedactionProfileView) { s := "/different"; v.PathRules[0].SelectedRuleTrace.SelectorValue = &s },
		"default selector":           func(v *RedactionProfileView) { v.ProfileDefault.PathSelector = &path },
	} {
		t.Run(name, func(t *testing.T) {
			var invalid RedactionProfileView
			if err := json.Unmarshal(encoded, &invalid); err != nil {
				t.Fatal(err)
			}
			change(&invalid)
			if validateRedactionProfileView(invalid) == nil {
				t.Fatal("invalid normalized profile accepted")
			}
		})
	}
	for _, fragment := range []string{`"time_bucket_granularity":null,`, `"time_bucket_granularity":"day",`, `"unknown":true,`, `"rules":[],`, `"default_action":{},`} {
		bad := []byte("{" + fragment + string(encoded[1:]))
		if _, _, err := decodePortableObject[RedactionProfileView](bad, reportingObjectDigest(RedactionProfileViewSchemaID, bad)); err == nil {
			t.Fatal("unknown or retired profile member accepted")
		}
	}
	for _, status := range []string{"adopted_current", "adopted_deprecated", "future_only"} {
		candidate := view
		candidate.ProfileStatus = status
		candidate.AllowedReleaseScopes = []string{ReleaseScopeInternalReview, ReleaseScopeExternal}
		if !profileSupportsScope(candidate, ReleaseScopeInternalReview) || profileSupportsScope(candidate, ReleaseScopeExternal) != (status == "adopted_current") {
			t.Fatal("profile lifecycle scope policy ignored")
		}
	}
	if _, _, sum, err := BuildRedactionProfileView(profile, strings.Repeat("0", 64)); err == nil || sum != "" {
		t.Fatal("unbound profile digest accepted")
	}
	retired := internalRedactionProfile()
	retired.SchemaID = "cartulary.redaction_profile.v1"
	if _, err := ValidateRedactionProfile(retired); err == nil {
		t.Fatal("retired authored profile schema accepted")
	}
	for _, id := range []string{InternalRedactionProfileID, ExternalRedactionProfileID, TokenizedRedactionProfileID} {
		if isSupportedRedactionProfileSelector(id, "1") || isSupportedRedactionProfileSelector(id, "2") {
			t.Fatal("API accepts a retired built-in profile tuple")
		}
		if _, _, err := ResolveRedactionProfile(id, "2", nil, SnapshotContent{}); err == nil {
			t.Fatal("retired built-in profile was reconstructed")
		}
	}
	for _, partitions := range [][]string{{}, {"public", "party:é", "internal_only", "blocked"}} {
		if validateDisclosurePartitionSet(partitions) != nil {
			t.Fatal("closed partition vocabulary rejected")
		}
	}
}

func testProfileIdentifierContract(t *testing.T) {
	for _, id := range []string{"", " leading", "trailing\u00a0", "line\nfeed", "c1\u0085", "nul\x00", "path/name", "path\\name", "fragment#id", string([]byte{0xff}), strings.Repeat("é", 129)} {
		for _, member := range []string{"profile", "version", "rule"} {
			profile := internalRedactionProfile()
			switch member {
			case "profile":
				profile.ProfileID = id
			case "version":
				profile.Version = id
			case "rule":
				class := ContentClassSourceEvidence
				profile.Rules = []RedactionRule{{RuleID: id, ContentClass: &class, Action: RedactionActionSpec{Type: ActionDrop}}}
			}
			if _, err := ValidateRedactionProfile(profile); err == nil {
				t.Fatalf("invalid %s identifier accepted", member)
			}
		}
	}
	for _, id := range []string{"id", "space inside", "é", "e\u0301", strings.Repeat("é", 128)} {
		profile := internalRedactionProfile()
		profile.ProfileID = id
		if !validReportingIdentifier(id) {
			t.Fatal("valid exact identifier rejected")
		}
		digest, err := ValidateRedactionProfile(profile)
		if err != nil {
			t.Fatal(err)
		}
		view, _, _, err := BuildRedactionProfileView(profile, digest)
		if err != nil || view.ProfileID != id {
			t.Fatal("identifier rewritten", err)
		}
	}
}
