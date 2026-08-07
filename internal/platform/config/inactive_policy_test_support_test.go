package config

import (
	"encoding/json"
	"fmt"
	"sort"
)

type testExtensionPolicy struct {
	registrations []ClaimRegistration
	claims        map[string]string
	parse         func(string, string) (any, error)
	validate      func(map[string]any) [][2]string
}

func (policy testExtensionPolicy) ClaimRegistrations() []ClaimRegistration {
	if policy.registrations == nil {
		return characterizedClaimRegistrations()
	}
	return append([]ClaimRegistration(nil), policy.registrations...)
}

func (policy testExtensionPolicy) Keys() []string {
	keys := make([]string, 0, len(policy.claims))
	for key := range policy.claims {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func (policy testExtensionPolicy) ClaimKey(key string) (string, bool) {
	claimKey, ok := policy.claims[key]
	return claimKey, ok
}

func (policy testExtensionPolicy) ParseOverlay(key string, raw string) (any, error) {
	if policy.parse != nil {
		return policy.parse(key, raw)
	}
	return raw, nil
}

func (policy testExtensionPolicy) ValidateAndDiscard(values map[string]any) [][2]string {
	if policy.validate != nil {
		return policy.validate(values)
	}
	findings := make([][2]string, 0, len(values))
	for key := range values {
		findings = append(findings, [2]string{key, "extension_config_without_claim"})
	}
	sort.Slice(findings, func(i, j int) bool { return findings[i][0] < findings[j][0] })
	return findings
}

func syntaxOnlyTestExtensionPolicy() ExtensionPolicy {
	const syntaxKey = "future_profile.syntax"
	const forbiddenKey = "future_profile.forbidden"
	return testExtensionPolicy{
		claims: map[string]string{
			syntaxKey:    "enterprise_authentication.claimed",
			forbiddenKey: "enterprise_authentication.claimed",
		},
		parse: func(key string, raw string) (any, error) {
			if key == forbiddenKey {
				return raw, nil
			}
			var value map[string]any
			if err := json.Unmarshal([]byte(raw), &value); err != nil {
				return nil, err
			}
			return value, nil
		},
		validate: func(values map[string]any) [][2]string {
			findings := make([][2]string, 0)
			for key, value := range values {
				if key == forbiddenKey {
					findings = append(findings, [2]string{key, "extension_config_without_claim"})
					continue
				}
				object, ok := value.(map[string]any)
				mode, modeOK := object["mode"].(string)
				if !ok || len(object) != 1 || !modeOK || mode != "strict" && mode != "relaxed" {
					findings = append(findings, [2]string{key, "extension_validation_result_invalid"})
				}
			}
			sort.Slice(findings, func(i, j int) bool {
				return fmt.Sprint(findings[i]) < fmt.Sprint(findings[j])
			})
			return findings
		},
	}
}

func characterizedClaimRegistrations() []ClaimRegistration {
	return []ClaimRegistration{
		{ID: "enterprise_authentication", Path: "enterprise_authentication.claimed"},
		{ID: "import", Path: "import.claimed"},
		{ID: "incident_portability", Path: "incident_portability.claimed"},
		{ID: "network_flow_activity", Path: "network_flow_activity.claimed"},
		{ID: "reference_pack", Path: "reference_pack.claimed"},
		{ID: "snapshot_reporting", Path: "snapshot_reporting.claimed"},
	}
}
