package openapicompat

import "testing"

func oauthSecurityScheme(flow string, scopes map[string]any) map[string]any {
	configuration := map[string]any{"scopes": scopes}
	if flow == "implicit" || flow == "authorizationCode" {
		configuration["authorizationUrl"] = "https://example.test/authorize"
	}
	if flow != "implicit" {
		configuration["tokenUrl"] = "https://example.test/token"
	}
	return map[string]any{"type": "oauth2", "flows": map[string]any{flow: configuration}}
}

func oauthSecurityDocument(t *testing.T, version string, schemes map[string]any) []byte {
	t.Helper()
	return securityRootDocument(t, map[string]any{
		"info":       map[string]any{"title": "OAuth fixture", "version": version},
		"components": map[string]any{"securitySchemes": schemes},
	})
}

func TestCompareOAuthScopeDescriptions(t *testing.T) {
	for _, flow := range []string{"implicit", "password", "clientCredentials", "authorizationCode"} {
		for _, name := range []string{"read", "description", "security", "x-cartulary-state-changing", "read/~"} {
			// Escaped scheme keys also exercise comparator robustness for inputs
			// outside OpenAPI's component-key grammar; names never select policy.
			for _, values := range [][2]string{{"old", "new"}, {"", "new"}, {"old", ""}} {
				t.Run(flow+"/"+name+"/"+values[0]+"-to-"+values[1], func(t *testing.T) {
					before := oauthSecurityDocument(t, "1.0.0", map[string]any{name: oauthSecurityScheme(flow, map[string]any{name: values[0]})})
					after := oauthSecurityDocument(t, "1.0.1", map[string]any{name: oauthSecurityScheme(flow, map[string]any{name: values[1]})})
					report, err := Compare(before, after)
					if err != nil {
						t.Fatal(err)
					}
					pointer := "/components/securitySchemes/" + escapePointer(name) + "/flows/" + flow + "/scopes/" + escapePointer(name)
					if len(report.Changes) != 1 || report.Changes[0].Pointer != pointer || report.Changes[0].Kind != "change" || report.Changes[0].Classification != NonBehavioral {
						t.Fatalf("description-only change: %#v", report.Changes)
					}
					if err := validateVersionPolicy(report); err != nil {
						t.Fatalf("patch description edit rejected: %v", err)
					}
				})
			}
		}
	}
}

func TestCompareOAuthScopeIdentityAndConfiguration(t *testing.T) {
	for _, flow := range []string{"implicit", "password", "clientCredentials", "authorizationCode"} {
		for _, name := range []string{"read", "description", "security", "x-cartulary-state-changing", "read/~"} {
			for _, operation := range []string{"add", "remove", "rename"} {
				t.Run(flow+"/"+name+"/"+operation, func(t *testing.T) {
					left, right := map[string]any{}, map[string]any{name: "description"}
					want := 1
					if operation == "remove" {
						left, right = right, left
					} else if operation == "rename" {
						left, right, want = right, map[string]any{"renamed": "description"}, 2
					}
					before := oauthSecurityDocument(t, "1.0.0", map[string]any{"oauth": oauthSecurityScheme(flow, left)})
					for _, version := range []string{"1.1.0", "2.0.0"} {
						after := oauthSecurityDocument(t, version, map[string]any{"oauth": oauthSecurityScheme(flow, right)})
						report, err := Compare(before, after)
						if err != nil || len(report.Changes) != want {
							t.Fatalf("scope identity: %#v %v", report, err)
						}
						for _, change := range report.Changes {
							if change.Classification != SecuritySensitive {
								t.Fatalf("scope identity downgraded: %#v", change)
							}
						}
						if err := validateVersionPolicy(report); (err == nil) != (version == "2.0.0") {
							t.Fatalf("scope identity version %s: %v", version, err)
						}
					}
				})
			}
		}
	}
	for _, invalid := range []any{nil, true, float64(1), []any{}, map[string]any{"description": "text"}} {
		for _, reverse := range []bool{false, true} {
			left, right := any("description"), invalid
			if reverse {
				left, right = right, left
			}
			before := oauthSecurityDocument(t, "1.0.0", map[string]any{"oauth": oauthSecurityScheme("implicit", map[string]any{"description": left})})
			after := oauthSecurityDocument(t, "2.0.0", map[string]any{"oauth": oauthSecurityScheme("implicit", map[string]any{"description": right})})
			report, err := Compare(before, after)
			if err != nil || len(report.Changes) != 1 || report.Changes[0].Classification != SecuritySensitive {
				t.Fatalf("invalid value downgraded (%T, reverse=%v): %#v %v", invalid, reverse, report, err)
			}
		}
	}
	for _, change := range []string{"endpoint", "unknown flow", "nested description", "whole scheme add", "whole scheme remove"} {
		t.Run(change, func(t *testing.T) {
			left := oauthSecurityScheme("implicit", map[string]any{"read": "old"})
			right := oauthSecurityScheme("implicit", map[string]any{"read": "old"})
			beforeSchemes, afterSchemes := map[string]any{"oauth": left}, map[string]any{"oauth": right}
			switch change {
			case "endpoint":
				right["flows"].(map[string]any)["implicit"].(map[string]any)["authorizationUrl"] = "https://example.test/new-authorize"
			case "unknown flow":
				beforeSchemes["oauth"] = oauthSecurityScheme("futureFlow", map[string]any{"description": "old"})
				afterSchemes["oauth"] = oauthSecurityScheme("futureFlow", map[string]any{"description": "new"})
			case "nested description":
				beforeSchemes["oauth"] = oauthSecurityScheme("implicit", map[string]any{"read": map[string]any{"description": "old"}})
				afterSchemes["oauth"] = oauthSecurityScheme("implicit", map[string]any{"read": map[string]any{"description": "new"}})
			case "whole scheme add":
				beforeSchemes = map[string]any{}
			case "whole scheme remove":
				afterSchemes = map[string]any{}
			}
			report, err := Compare(oauthSecurityDocument(t, "1.0.0", beforeSchemes), oauthSecurityDocument(t, "2.0.0", afterSchemes))
			if err != nil || len(report.Changes) != 1 || report.Changes[0].Classification != SecuritySensitive {
				t.Fatalf("security configuration downgraded: %#v %v", report, err)
			}
		})
	}
}

func TestCompareOAuthScopeReviewPolicy(t *testing.T) {
	before := oauthSecurityDocument(t, "1.0.0", map[string]any{"oauth": oauthSecurityScheme("implicit", map[string]any{"read": "old"})})
	after := oauthSecurityDocument(t, "1.0.1", map[string]any{"oauth": oauthSecurityScheme("implicit", map[string]any{"read": "new"})})
	report, err := Compare(before, after)
	if err != nil || len(report.Changes) != 1 || report.Changes[0].Classification != NonBehavioral {
		t.Fatalf("description change: %#v %v", report, err)
	}
	change := report.Changes[0]
	approved := changeSet{SchemaID: changeSetSchemaID, BaselineVersion: "1.0.0", TargetVersion: "1.0.1", Changes: []approvedChangeRef{
		{Fingerprint: change.Fingerprint, Classification: NonBehavioral, OwnerID: "platform.openapi", Rationale: "Clarify existing scope description"},
	}}
	if err := validateChangeSet(report, approved); err != nil {
		t.Fatal(err)
	}
	for _, kind := range []string{"missing", "stale", "duplicate", "classification"} {
		bad := approved
		bad.Changes = append([]approvedChangeRef(nil), approved.Changes...)
		switch kind {
		case "missing":
			bad.Changes = nil
		case "stale":
			bad.Changes[0].Fingerprint = "stale"
		case "duplicate":
			bad.Changes = append(bad.Changes, bad.Changes[0])
		case "classification":
			bad.Changes[0].Classification = SecuritySensitive
		}
		if err := validateChangeSet(report, bad); err == nil {
			t.Fatalf("accepted %s description review", kind)
		}
	}
}
