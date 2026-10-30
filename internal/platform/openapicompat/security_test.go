package openapicompat

import (
	"encoding/json"
	"testing"
)

func securityDocument(t *testing.T, paths map[string]any) []byte {
	t.Helper()
	raw, err := json.Marshal(map[string]any{"openapi": "3.1.0", "info": map[string]any{"version": "1.0.0"}, "paths": paths})
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func TestCompareNestedSecurityDeclarations(t *testing.T) {
	for _, value := range []bool{false, true} {
		operation := map[string]any{"x-cartulary-state-changing": value, "security": []any{}}
		for _, scope := range []string{"path", "operation", "property"} {
			before := map[string]any{}
			after := map[string]any{"/source-inspection": map[string]any{"post": operation}}
			if scope == "operation" {
				before["/source-inspection"] = map[string]any{}
			}
			if scope == "property" {
				before["/source-inspection"] = map[string]any{"post": map[string]any{}}
			}
			for _, remove := range []bool{false, true} {
				left, right := securityDocument(t, before), securityDocument(t, after)
				kind := "add"
				if remove {
					left, right, kind = right, left, "remove"
				}
				report, err := Compare(left, right)
				if err != nil {
					t.Fatal(err)
				}
				for _, suffix := range []string{"x-cartulary-state-changing", "security"} {
					pointer := "/paths/~1source-inspection/post/" + suffix
					count := 0
					for _, change := range report.Changes {
						if change.Pointer == pointer && change.Kind == kind && change.Classification == SecuritySensitive {
							count++
						}
					}
					if count != 1 {
						t.Fatalf("scope=%s value=%v kind=%s pointer=%s: got %d security records: %#v", scope, value, kind, pointer, count, report.Changes)
					}
				}
			}
		}
	}
}

func securityRootDocument(t *testing.T, fields map[string]any) []byte {
	t.Helper()
	root := map[string]any{"openapi": "3.1.0", "info": map[string]any{"version": "1.0.0"}, "paths": map[string]any{}}
	for key, value := range fields {
		root[key] = value
	}
	raw, err := json.Marshal(root)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func TestCompareSecurityContainersAndReplacements(t *testing.T) {
	operation := map[string]any{"x-cartulary-state-changing": false, "security": []any{}}
	scheme := map[string]any{"type": "apiKey", "in": "header", "name": "X-Key"}
	cases := []struct {
		name          string
		before, after map[string]any
		pointers      []string
	}{
		{"root security", nil, map[string]any{"security": []any{}}, []string{"/security"}},
		{"components", nil, map[string]any{"components": map[string]any{"securitySchemes": map[string]any{"description": scheme, "a/~b": scheme}}}, []string{"/components/securitySchemes/description", "/components/securitySchemes/a~1~0b"}},
		{"schemes", map[string]any{"components": map[string]any{}}, map[string]any{"components": map[string]any{"securitySchemes": map[string]any{"key": scheme}}}, []string{"/components/securitySchemes/key"}},
		{"scheme", map[string]any{"components": map[string]any{"securitySchemes": map[string]any{}}}, map[string]any{"components": map[string]any{"securitySchemes": map[string]any{"key": scheme}}}, []string{"/components/securitySchemes/key"}},
		{"replace path", map[string]any{"paths": map[string]any{"/a/~b": nil}}, map[string]any{"paths": map[string]any{"/a/~b": map[string]any{"get": operation, "post": operation}}}, []string{"/paths/~1a~1~0b/get/security", "/paths/~1a~1~0b/get/x-cartulary-state-changing", "/paths/~1a~1~0b/post/security", "/paths/~1a~1~0b/post/x-cartulary-state-changing"}},
		{"replace operation", map[string]any{"paths": map[string]any{"/a": map[string]any{"post": nil}}}, map[string]any{"paths": map[string]any{"/a": map[string]any{"post": operation}}}, []string{"/paths/~1a/post/security", "/paths/~1a/post/x-cartulary-state-changing"}},
		{"webhook", nil, map[string]any{"webhooks": map[string]any{"event": map[string]any{"post": operation}}}, []string{"/webhooks/event/post/security", "/webhooks/event/post/x-cartulary-state-changing"}},
		{"reusable path", nil, map[string]any{"components": map[string]any{"pathItems": map[string]any{"shared": map[string]any{"get": operation}}}}, []string{"/components/pathItems/shared/get/security", "/components/pathItems/shared/get/x-cartulary-state-changing"}},
		{"callback", nil, map[string]any{"paths": map[string]any{"/a": map[string]any{"post": map[string]any{"callbacks": map[string]any{"event": map[string]any{"{$request.body#/url}": map[string]any{"post": operation}}}}}}}, []string{"/paths/~1a/post/callbacks/event/{$request.body#~1url}/post/security", "/paths/~1a/post/callbacks/event/{$request.body#~1url}/post/x-cartulary-state-changing"}},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			for _, remove := range []bool{false, true} {
				left, right := securityRootDocument(t, test.before), securityRootDocument(t, test.after)
				kind := "add"
				if remove {
					left, right, kind = right, left, "remove"
				}
				report, err := Compare(left, right)
				if err != nil {
					t.Fatal(err)
				}
				seen := map[string]bool{}
				for _, change := range report.Changes {
					if seen[change.Pointer] {
						t.Fatalf("duplicate pointer: %s", change.Pointer)
					}
					seen[change.Pointer] = true
				}
				securityCount := 0
				for _, change := range report.Changes {
					if change.Classification == SecuritySensitive {
						securityCount++
					}
				}
				if securityCount != len(test.pointers) {
					t.Fatalf("unexpected security changes: %#v", report.Changes)
				}
				for _, pointer := range test.pointers {
					found := false
					for _, change := range report.Changes {
						if change.Pointer == pointer && change.Kind == kind && change.Classification == SecuritySensitive {
							found = true
						}
					}
					if !found {
						t.Fatalf("missing %s %s: %#v", kind, pointer, report.Changes)
					}
				}
				for range 5 {
					again, err := Compare(left, right)
					if err != nil || !equalJSON(report, again) {
						t.Fatalf("nondeterministic comparison: %v", err)
					}
				}
			}
		})
	}
}

func TestCompareSecurityLeavesAndFalsePositives(t *testing.T) {
	for _, test := range []struct {
		name, before, after string
		want                Classification
	}{
		{"mutation", `{"paths":{"/a":{"post":{"x-cartulary-state-changing":false}}}}`, `{"paths":{"/a":{"post":{"x-cartulary-state-changing":true}}}}`, SecuritySensitive},
		{"requirements", `{"security":[]}`, `{"security":[{"description":[]}]}`, SecuritySensitive},
		{"operation requirements", `{"paths":{"/a":{"get":{"security":[]}}}}`, `{"paths":{"/a":{"get":{"security":[{"key":[]}]}}}}`, SecuritySensitive},
		{"scheme type", `{"components":{"securitySchemes":{"key":{"type":"apiKey"}}}}`, `{"components":{"securitySchemes":{"key":{"type":"http"}}}}`, SecuritySensitive},
		{"scheme description", `{"components":{"securitySchemes":{"key":{"description":"old"}}}}`, `{"components":{"securitySchemes":{"key":{"description":"new"}}}}`, NonBehavioral},
		{"scope named description", `{"components":{"securitySchemes":{"key":{"flows":{"implicit":{"scopes":{"description":"old"}}}}}}}`, `{"components":{"securitySchemes":{"key":{"flows":{"implicit":{"scopes":{"description":"new"}}}}}}}`, NonBehavioral},
		{"operation description", `{"paths":{"/a":{"get":{"description":"old"}}}}`, `{"paths":{"/a":{"get":{"description":"new"}}}}`, NonBehavioral},
		{"schema security property", `{"components":{"schemas":{"Example":{"properties":{}}}}}`, `{"components":{"schemas":{"Example":{"properties":{"security":{"type":"string"}}}}}}`, Additive},
		{"schema mutation property", `{"components":{"schemas":{"Example":{"properties":{}}}}}`, `{"components":{"schemas":{"Example":{"properties":{"x-cartulary-state-changing":{"type":"boolean"}}}}}}`, Additive},
		{"security prefix", `{"paths":{"/a":{"get":{}}}}`, `{"paths":{"/a":{"get":{"securityNote":"text"}}}}`, Additive},
		{"nested example", `{"components":{"schemas":{"Example":{}}}}`, `{"components":{"schemas":{"Example":{"example":{"paths":{"/a":{"get":{"security":[],"x-cartulary-state-changing":false}}}}}}}}`, NonBehavioral},
	} {
		t.Run(test.name, func(t *testing.T) {
			var before, after map[string]any
			if err := json.Unmarshal([]byte(test.before), &before); err != nil {
				t.Fatal(err)
			}
			if err := json.Unmarshal([]byte(test.after), &after); err != nil {
				t.Fatal(err)
			}
			report, err := Compare(securityRootDocument(t, before), securityRootDocument(t, after))
			if err != nil {
				t.Fatal(err)
			}
			if len(report.Changes) != 1 || report.Changes[0].Classification != test.want {
				t.Fatalf("want %s: %#v", test.want, report.Changes)
			}
		})
	}
}

func TestCompareNestedSecurityRequiresExactReview(t *testing.T) {
	before := securityDocument(t, map[string]any{})
	after := securityDocument(t, map[string]any{"/inspect": map[string]any{"post": map[string]any{"x-cartulary-state-changing": false, "security": []any{}}}})
	report, err := Compare(before, after)
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Changes) != 3 || report.Changes[0].Classification != Additive {
		t.Fatalf("structural change lost: %#v", report.Changes)
	}
	report.TargetVersion = "1.1.0"
	if err := validateVersionPolicy(report); err == nil {
		t.Fatal("nested security accepted without major version")
	}
	report.TargetVersion = "2.0.0"
	if err := validateVersionPolicy(report); err != nil {
		t.Fatal(err)
	}
	approved := changeSet{SchemaID: changeSetSchemaID, BaselineVersion: report.BaselineVersion, TargetVersion: report.TargetVersion}
	for _, change := range report.Changes {
		approved.Changes = append(approved.Changes, approvedChangeRef{Fingerprint: change.Fingerprint, Classification: change.Classification, OwnerID: "platform.openapi", Rationale: "Review fixture declaration"})
	}
	if err := validateChangeSet(report, approved); err != nil {
		t.Fatal(err)
	}
	for _, kind := range []string{"missing", "stale", "duplicate", "classification", "owner", "rationale"} {
		bad := approved
		bad.Changes = append([]approvedChangeRef(nil), approved.Changes...)
		switch kind {
		case "missing":
			bad.Changes = bad.Changes[:1]
		case "stale":
			bad.Changes[1].Fingerprint = "stale"
		case "duplicate":
			bad.Changes = append(bad.Changes, bad.Changes[1])
		case "classification":
			bad.Changes[1].Classification = Additive
		case "owner":
			bad.Changes[1].OwnerID = ""
		case "rationale":
			bad.Changes[1].Rationale = " "
		}
		if err := validateChangeSet(report, bad); err == nil {
			t.Fatalf("accepted %s review", kind)
		}
	}
}
