package extensions

import "testing"

func TestInactiveConfigurationCatalog_Unit(t *testing.T) {
	schema := map[string]any{
		"type":                 "object",
		"additionalProperties": false,
		"properties": map[string]any{
			"mode": map[string]any{"type": "string", "enum": []any{"strict", "relaxed"}},
		},
		"required": []any{"mode"},
	}
	catalog, err := NewInactiveConfigurationCatalog([]InactiveConfigurationPolicy{
		{
			ProfileID: "future_profile",
			ClaimKey:  "future_profile.claimed",
			Key:       "future_profile.syntax",
			Kind:      PolicySyntaxOnly,
			Schema:    schema,
		},
		{
			ProfileID: "future_profile",
			ClaimKey:  "future_profile.claimed",
			Key:       "future_profile.forbidden",
			Kind:      PolicyForbidden,
		},
	})
	if err != nil {
		t.Fatalf("build inactive configuration catalog: %v", err)
	}
	if claimKey, ok := catalog.ClaimKey("future_profile.syntax"); !ok || claimKey != "future_profile.claimed" {
		t.Fatalf("unexpected claim key: %q present=%t", claimKey, ok)
	}

	parsed, err := catalog.ParseOverlay("future_profile.syntax", `{"mode":"strict"}`)
	if err != nil {
		t.Fatalf("parse syntax-only overlay: %v", err)
	}
	if findings := catalog.ValidateAndDiscard(map[string]any{"future_profile.syntax": parsed}); len(findings) != 0 {
		t.Fatalf("valid syntax-only value was rejected: %v", findings)
	}
	findings := catalog.ValidateAndDiscard(map[string]any{
		"future_profile.syntax":    map[string]any{"mode": "active"},
		"future_profile.forbidden": "vault://must-not-leak",
	})
	if len(findings) != 2 ||
		findings[0] != [2]string{"future_profile.forbidden", "extension_config_without_claim"} ||
		findings[1] != [2]string{"future_profile.syntax", "extension_validation_result_invalid"} {
		t.Fatalf("unexpected redacted findings: %v", findings)
	}

	schema["properties"].(map[string]any)["mode"] = map[string]any{"type": "integer"}
	parsed, err = catalog.ParseOverlay("future_profile.syntax", `{"mode":"relaxed"}`)
	if err != nil {
		t.Fatalf("catalog was mutated through its construction input: %v", err)
	}
	if findings := catalog.ValidateAndDiscard(map[string]any{"future_profile.syntax": parsed}); len(findings) != 0 {
		t.Fatalf("catalog schema was not immutable: %v", findings)
	}

	if _, err := NewInactiveConfigurationCatalog([]InactiveConfigurationPolicy{
		{ProfileID: "a", ClaimKey: "a.claimed", Key: "same.key", Kind: PolicyForbidden},
		{ProfileID: "b", ClaimKey: "b.claimed", Key: "same.key", Kind: PolicyForbidden},
	}); err == nil {
		t.Fatal("duplicate inactive key was accepted")
	}
}
