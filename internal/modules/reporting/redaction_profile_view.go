package reporting

import (
	"fmt"
	"slices"
)

const selectedRuleTraceSchemaID = "selected_rule_trace.v1"
const redactionRuleViewSchemaID = "cartulary.reporting_redaction_rule_view.v1"

type SelectedRuleTrace struct {
	SchemaID       string  `json:"schema_id"`
	RuleID         string  `json:"rule_id"`
	RulePrecedence string  `json:"rule_precedence"`
	SelectorKind   string  `json:"selector_kind"`
	SelectorValue  *string `json:"selector_value"`
}

type RedactionProfileView struct {
	SchemaID                       string              `json:"schema_id"`
	ProfileID                      string              `json:"redaction_profile_id"`
	ProfileVersion                 string              `json:"redaction_profile_version"`
	ProfileSHA256                  string              `json:"redaction_profile_sha256"`
	ProfileStatus                  string              `json:"profile_status"`
	AllowedReleaseScopes           []string            `json:"allowed_release_scopes"`
	AllowedDisclosurePartitionRefs []string            `json:"allowed_disclosure_partition_refs"`
	IncludeSuperseded              bool                `json:"include_superseded"`
	NeutralTokenFamily             bool                `json:"neutral_token_family"`
	AllowAuthoredPresentationText  bool                `json:"allow_authored_presentation_text"`
	PathRules                      []RedactionRuleView `json:"path_rules"`
	ContentClassRules              []RedactionRuleView `json:"content_class_rules"`
	ProfileDefault                 RedactionRuleView   `json:"profile_default"`
}

type RedactionRuleView struct {
	SchemaID             string            `json:"schema_id"`
	RuleID               string            `json:"rule_id"`
	RuleOrder            int               `json:"rule_order"`
	SelectorKind         string            `json:"selector_kind"`
	PathSelector         *string           `json:"path_selector"`
	ContentClassSelector *string           `json:"content_class_selector"`
	Action               string            `json:"action"`
	TokenBacked          bool              `json:"token_backed"`
	LiteralOutput        *string           `json:"literal_output"`
	TruncateMaxScalars   *int              `json:"truncate_max_scalars"`
	SelectedRuleTrace    SelectedRuleTrace `json:"selected_rule_trace"`
}

func profilePolicyError() error {
	return fmt.Errorf("%w: invalid profile policy", ErrInvalidRedactionProfile)
}

func validateProfilePolicy(status string, scopes []string) error {
	if status != "adopted_current" && status != "adopted_deprecated" && status != "future_only" {
		return profilePolicyError()
	}
	order := []string{ReleaseScopeInternalDraft, ReleaseScopeInternalReview, ReleaseScopeExternal}
	previous := -1
	if len(scopes) == 0 {
		return profilePolicyError()
	}
	for _, scope := range scopes {
		index := slices.Index(order, scope)
		if index <= previous {
			return profilePolicyError()
		}
		previous = index
	}
	return nil
}

func profileSupportsScope(view RedactionProfileView, scope string) bool {
	return slices.Contains(view.AllowedReleaseScopes, scope) &&
		(scope != ReleaseScopeExternal || view.ProfileStatus == "adopted_current")
}

// Normalize only authored policy. No source value participates in this view.
// The declaration order is preserved independently for each precedence class.
func BuildRedactionProfileView(profile RedactionProfile, profileSHA256 string) (RedactionProfileView, []byte, string, error) {
	digest, err := ValidateRedactionProfile(profile)
	if err != nil || digest != profileSHA256 {
		return RedactionProfileView{}, nil, "", profilePolicyError()
	}
	view := RedactionProfileView{
		SchemaID: RedactionProfileViewSchemaID, ProfileID: profile.ProfileID,
		ProfileVersion: profile.Version, ProfileSHA256: profileSHA256,
		ProfileStatus: profile.ProfileStatus, AllowedReleaseScopes: append([]string{}, profile.AllowedReleaseScopes...),
		AllowedDisclosurePartitionRefs: append([]string{}, profile.AllowedDisclosurePartitionRefs...),
		IncludeSuperseded:              profile.IncludeSuperseded, NeutralTokenFamily: profile.NeutralTokenFamily,
		AllowAuthoredPresentationText: profile.AllowAuthoredPresentationText,
		PathRules:                     []RedactionRuleView{}, ContentClassRules: []RedactionRuleView{},
		ProfileDefault: normalizedRedactionRule("profile_default", 1, "profile_default", nil, profile.DefaultAction),
	}
	for index, rule := range profile.Rules {
		if rule.Path != nil {
			view.PathRules = append(view.PathRules, normalizedRedactionRule(rule.RuleID, index+1, "path", rule.Path, rule.Action))
		} else {
			view.ContentClassRules = append(view.ContentClassRules, normalizedRedactionRule(rule.RuleID, index+1, "content_class", rule.ContentClass, rule.Action))
		}
	}
	if err := validateRedactionProfileView(view); err != nil {
		return RedactionProfileView{}, nil, "", err
	}
	encoded, err := canonicalJSON(view)
	if err != nil {
		return RedactionProfileView{}, nil, "", err
	}
	return view, encoded, reportingObjectDigest(view.SchemaID, encoded), nil
}

func copyProfileValue[T any](value *T) *T {
	if value == nil {
		return nil
	}
	copy := *value
	return &copy
}

func normalizedRedactionRule(id string, order int, kind string, selector *string, action RedactionActionSpec) RedactionRuleView {
	rule := RedactionRuleView{
		SchemaID: redactionRuleViewSchemaID, RuleID: id, RuleOrder: order,
		SelectorKind: kind, Action: action.Type, TokenBacked: action.TokenBacked,
		TruncateMaxScalars: copyProfileValue(action.MaxChars),
		SelectedRuleTrace: SelectedRuleTrace{SchemaID: selectedRuleTraceSchemaID, RuleID: id,
			RulePrecedence: kind, SelectorKind: kind, SelectorValue: copyProfileValue(selector)},
	}
	if kind == "path" {
		rule.PathSelector = copyProfileValue(selector)
	} else if kind == "content_class" {
		rule.ContentClassSelector = copyProfileValue(selector)
	}
	if !action.TokenBacked {
		switch action.Type {
		case ActionMask:
			literal := DefaultMaskText
			if action.ReplacementText != nil {
				literal = *action.ReplacementText
			}
			rule.LiteralOutput = &literal
		case ActionStub:
			literal := DefaultStubText
			if action.StubText != nil {
				literal = *action.StubText
			}
			rule.LiteralOutput = &literal
		}
	}
	return rule
}

func validSelectedRuleTrace(trace SelectedRuleTrace) bool {
	if trace.SchemaID != selectedRuleTraceSchemaID || !validReportingIdentifier(trace.RuleID) || trace.RulePrecedence != trace.SelectorKind {
		return false
	}
	switch trace.SelectorKind {
	case "path":
		return trace.SelectorValue != nil && validExactExportModelPath(*trace.SelectorValue)
	case "content_class":
		return trace.SelectorValue != nil && validReportingContentClass(*trace.SelectorValue)
	case "profile_default":
		return trace.SelectorValue == nil
	}
	return false
}

func (rule RedactionRuleView) actionSpec() RedactionActionSpec {
	action := RedactionActionSpec{Type: rule.Action, TokenBacked: rule.TokenBacked, MaxChars: rule.TruncateMaxScalars}
	if rule.Action == ActionMask {
		action.ReplacementText = rule.LiteralOutput
	} else if rule.Action == ActionStub {
		action.StubText = rule.LiteralOutput
	}
	return action
}

func validateRedactionProfileView(view RedactionProfileView) error {
	if view.SchemaID != RedactionProfileViewSchemaID || !validReportingIdentifier(view.ProfileID) || !validReportingIdentifier(view.ProfileVersion) ||
		!sha256HexPattern.MatchString(view.ProfileSHA256) || validateProfilePolicy(view.ProfileStatus, view.AllowedReleaseScopes) != nil ||
		view.AllowedDisclosurePartitionRefs == nil || validateDisclosurePartitionSet(view.AllowedDisclosurePartitionRefs) != nil ||
		view.PathRules == nil || view.ContentClassRules == nil {
		return profilePolicyError()
	}
	ids, selectors := map[string]bool{}, map[string]bool{}
	for _, group := range []struct {
		kind  string
		rules []RedactionRuleView
	}{{"path", view.PathRules}, {"content_class", view.ContentClassRules}, {"profile_default", []RedactionRuleView{view.ProfileDefault}}} {
		for index, rule := range group.rules {
			if rule.SchemaID != redactionRuleViewSchemaID || !validReportingIdentifier(rule.RuleID) || ids[rule.RuleID] || rule.RuleOrder < 1 || rule.SelectorKind != group.kind ||
				!validSelectedRuleTrace(rule.SelectedRuleTrace) || rule.SelectedRuleTrace.RuleID != rule.RuleID || rule.SelectedRuleTrace.SelectorKind != rule.SelectorKind ||
				validateAction(rule.actionSpec()) != nil {
				return profilePolicyError()
			}
			ids[rule.RuleID] = true
			if index > 0 && (group.rules[index-1].RuleOrder > rule.RuleOrder ||
				group.rules[index-1].RuleOrder == rule.RuleOrder && group.rules[index-1].RuleID >= rule.RuleID) {
				return profilePolicyError()
			}
			selector := rule.PathSelector
			if group.kind == "content_class" {
				selector = rule.ContentClassSelector
			}
			if (group.kind == "path" && (rule.PathSelector == nil || rule.ContentClassSelector != nil)) ||
				(group.kind == "content_class" && (rule.ContentClassSelector == nil || rule.PathSelector != nil)) ||
				(group.kind == "profile_default" && (rule.PathSelector != nil || rule.ContentClassSelector != nil)) ||
				!equalOptionalString(selector, rule.SelectedRuleTrace.SelectorValue) {
				return profilePolicyError()
			}
			if selector != nil {
				key := group.kind + "\x00" + *selector
				if selectors[key] {
					return profilePolicyError()
				}
				selectors[key] = true
			}
			literalRequired := !rule.TokenBacked && (rule.Action == ActionMask || rule.Action == ActionStub)
			if (rule.LiteralOutput != nil) != literalRequired {
				return profilePolicyError()
			}
		}
	}
	return nil
}

func selectRedactionRule(field ExportField, modelPath string, view RedactionProfileView) (string, RedactionActionSpec, SelectedRuleTrace) {
	for _, rule := range view.PathRules {
		if *rule.PathSelector == modelPath {
			return rule.RuleID, rule.actionSpec(), rule.SelectedRuleTrace
		}
	}
	for _, rule := range view.ContentClassRules {
		if *rule.ContentClassSelector == field.ContentClass {
			return rule.RuleID, rule.actionSpec(), rule.SelectedRuleTrace
		}
	}
	rule := view.ProfileDefault
	return rule.RuleID, rule.actionSpec(), rule.SelectedRuleTrace
}
