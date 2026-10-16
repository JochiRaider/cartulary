package reporting

import (
	"errors"
	"strings"
	"testing"
)

func testRedactionActionParameters(t *testing.T) {
	literal, one, maximum, excessive := "literal", 1, 240, 241
	for _, action := range []string{ActionAllow, ActionDrop, ActionMask, ActionStub, ActionTruncate} {
		for _, parameter := range []string{"replacement_text", "stub_text", "max_chars"} {
			if action == ActionMask && parameter == "replacement_text" ||
				action == ActionStub && parameter == "stub_text" ||
				action == ActionTruncate && parameter == "max_chars" {
				continue
			}
			t.Run(action+"/"+parameter, func(t *testing.T) {
				candidate := RedactionActionSpec{Type: action}
				if action == ActionTruncate {
					candidate.MaxChars = &one
				}
				switch parameter {
				case "replacement_text":
					candidate.ReplacementText = &literal
				case "stub_text":
					candidate.StubText = &literal
				case "max_chars":
					candidate.MaxChars = &one
				}
				assertInvalidActionProfile(t, candidate)
			})
		}
	}
	for _, value := range []string{"line\nbreak", "carriage\rreturn", "tab\tvalue", "nul\x00value", "c1\u0085value", string([]byte{0xff})} {
		for _, action := range []string{ActionMask, ActionStub} {
			candidate := RedactionActionSpec{Type: action}
			if action == ActionMask {
				candidate.ReplacementText = &value
			} else {
				candidate.StubText = &value
			}
			assertInvalidActionProfile(t, candidate)
		}
	}
	assertInvalidActionProfile(t, RedactionActionSpec{Type: ActionTruncate, MaxChars: &excessive})
	// Both endpoints are admitted; the bound counts Unicode scalars, not bytes.
	for _, limit := range []*int{&one, &maximum} {
		action := RedactionActionSpec{Type: ActionTruncate, MaxChars: limit}
		if err := validateAction(action); err != nil {
			t.Fatal(err)
		}
		source := strings.Repeat("é", *limit+1)
		value, included, outcome, err := applyRedactionAction(source, action)
		if err != nil || !included || outcome != "truncated" || value != strings.Repeat("é", *limit)+DefaultTruncateMark {
			t.Fatalf("scalar truncation: value=%v included=%t outcome=%s err=%v", value, included, outcome, err)
		}
	}
	for _, value := range []string{"", "é", "<script>literal only</script>", "{{ literal }}"} {
		for _, action := range []RedactionActionSpec{{Type: ActionMask, ReplacementText: &value}, {Type: ActionStub, StubText: &value}} {
			if err := validateAction(action); err != nil {
				t.Fatalf("safe literal rejected: %v", err)
			}
			got, included, _, err := applyRedactionAction("source", action)
			if err != nil || !included || got != value {
				t.Fatal("admitted literal was rewritten")
			}
		}
	}
}

func assertInvalidActionProfile(t testing.TB, action RedactionActionSpec) {
	t.Helper()
	// The same admission applies to profile defaults, exact paths and classes.
	path, class := "$.sections['sec-0001'].blocks['sec-0001.b0001'].fields['value']", ContentClassCuratedNarrative
	for _, placement := range []string{"default", "path", "class"} {
		profile := RedactionProfile{
			ProfileStatus: "adopted_current", AllowedReleaseScopes: []string{ReleaseScopeInternalDraft, ReleaseScopeInternalReview, ReleaseScopeExternal},
			SchemaID: RedactionProfileSchemaID, ProfileID: "test.closed.actions", Version: "1",
			DefaultAction: RedactionActionSpec{Type: ActionAllow},
		}
		switch placement {
		case "default":
			profile.DefaultAction = action
		case "path":
			profile.Rules = []RedactionRule{{RuleID: "selected", Path: &path, Action: action}}
		case "class":
			profile.Rules = []RedactionRule{{RuleID: "selected", ContentClass: &class, Action: action}}
		}
		if digest, err := ValidateRedactionProfile(profile); !errors.Is(err, ErrInvalidRedactionProfile) || digest != "" {
			t.Fatalf("%s accepted invalid action parameters: digest=%q error=%v", placement, digest, err)
		}
	}
}
