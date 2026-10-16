package reporting

import (
	"bytes"
	"encoding/json"
	"testing"
)

// Exercised by the renderer's artifact-boundary catalog entrypoint.
func assertSensitiveArtifactAdmission(t *testing.T) {
	t.Helper()
	assertShareableBundleMembers(t)
	numberPath := "$.sections['sec-0001'].blocks['sec-0001.b0001'].fields['number']"
	profile := RedactionProfile{ProfileStatus: "adopted_current", AllowedReleaseScopes: []string{ReleaseScopeInternalDraft, ReleaseScopeInternalReview, ReleaseScopeExternal},
		SchemaID: RedactionProfileSchemaID, ProfileID: "test.private-adjunct", Version: "1", DefaultAction: RedactionActionSpec{Type: ActionMask, TokenBacked: true},
		AllowedDisclosurePartitionRefs: []string{"public", "party:party_a", "internal_only"},
		Rules:                          []RedactionRule{{RuleID: "number-stub", Path: &numberPath, Action: RedactionActionSpec{Type: ActionStub, TokenBacked: true}}}}
	model := fixtureStructuredRenderModel(t, []ExportField{{Path: "/parties/party_a", ContentClass: ContentClassSourceEvidence, SourceFamily: "party", DisclosurePartitionRefs: []string{"public", "internal_only"}, Value: map[string]any{"display_name": "Private Party", "number": json.Number("9007199254740991")}}})
	fixtureAllowContainerRules(t, &profile, model.SnapshotContent)
	profileSHA, err := ValidateRedactionProfile(profile)
	if err != nil {
		t.Fatal(err)
	}
	data, err := canonicalJSON(model)
	if err != nil {
		t.Fatal(err)
	}
	redaction, err := RedactExportModel(model, fixtureSourceFactsForContent(model.SnapshotContent), profile, profileSHA, reportingObjectDigest(ExportModelSchemaID, data), ReleaseScopeInternalReview, nil)
	if err != nil {
		t.Fatal(err)
	}
	rendered := RenderedRelease{Redaction: redaction, RedactionManifestJSON: redaction.ManifestJSON, RedactionManifestSHA256: redaction.ManifestSHA256,
		RenderBundle: RenderBundle{Manifest: RenderBundleManifest{RenderIdentity: model.RenderIdentity, SnapshotID: model.SnapshotID, ReleaseScope: ReleaseScopeInternalReview,
			ExportModelSHA256: reportingObjectDigest(ExportModelSchemaID, data), TokenManifestSHA256: &redaction.TokenManifestSHA256, RedactionManifestSHA256: redaction.ManifestSHA256}}}
	if err := validateRevealMap(rendered); err != nil {
		t.Fatal("valid private adjunct rejected", err)
	}
	if len(redaction.RevealMap.Entries) != 1 || len(redaction.TokenManifest.Entries) != 1 || len(redaction.Manifest.Entries) != 2 {
		t.Fatal("two field rules must share one subject token and one reveal entry")
	}
	if bytes.Contains(redaction.RevealMapJSON, []byte("9007199254740991")) || bytes.Contains(redaction.RevealMapJSON, []byte("original_value")) {
		t.Fatal("private reveal map retained an unrelated original field value")
	}
	for name, mutate := range map[string]func(*RedactionRevealMap){
		"truncated":        func(m *RedactionRevealMap) { m.Entries = m.Entries[:0] },
		"foreign snapshot": func(m *RedactionRevealMap) { m.SnapshotID = "00000000-0000-0000-0000-000000000404" },
		"foreign identity": func(m *RedactionRevealMap) { id := "00000000-0000-0000-0000-000000000404"; m.ReleaseID = &id },
		"different token":  func(m *RedactionRevealMap) { m.Entries[0].DisplayToken = "different" },
		"altered display":  func(m *RedactionRevealMap) { m.Entries[0].CanonicalDisplayValue = "different" },
		"self-consistent altered display": func(m *RedactionRevealMap) {
			m.Entries[0].CanonicalDisplayValue = "different"
			m.Entries[0].CanonicalDisplayValueSHA256, _ = subjectDisplayValueDigest(m.Entries[0].StableSubjectRef, &m.Entries[0].CanonicalDisplayValue)
		},
		"duplicate entry":    func(m *RedactionRevealMap) { m.Entries = append(m.Entries, m.Entries[0]) },
		"foreign profile":    func(m *RedactionRevealMap) { m.ProfileSHA256 = "foreign" },
		"foreign source":     func(m *RedactionRevealMap) { v := "foreign"; m.Entries[0].SourceRecordID = &v },
		"mixed identity":     func(m *RedactionRevealMap) { v := "mention"; m.Entries[0].EntityMentionID = &v },
		"missing recipients": func(m *RedactionRevealMap) { m.Entries[0].RecipientPartitionRefs = nil },
		"foreign recipient":  func(m *RedactionRevealMap) { m.Entries[0].RecipientPartitionRefs = []string{"party:unselected"} },
	} {
		t.Run("private adjunct/"+name, func(t *testing.T) {
			m, _, err := decodePortableObject[RedactionRevealMap](redaction.RevealMapJSON, redaction.RevealMapSHA256)
			if err != nil {
				t.Fatal(err)
			}
			mutate(&m)
			changed := rendered
			changed.Redaction.RevealMapJSON, err = canonicalJSON(m)
			if err != nil {
				t.Fatal(err)
			}
			changed.Redaction.RevealMapSHA256 = reportingObjectDigest(m.SchemaID, changed.Redaction.RevealMapJSON)
			if validateRevealMap(changed) == nil {
				t.Fatal("altered adjunct with recomputed digest admitted")
			}
		})
	}
	for name, mutate := range map[string]func(*RedactionTokenManifest){
		"foreign profile":       func(m *RedactionTokenManifest) { m.ProfileSHA256 = "foreign" },
		"foreign rule manifest": func(m *RedactionTokenManifest) { m.RedactionManifestSHA256 = "foreign" },
		"missing subject":       func(m *RedactionTokenManifest) { m.Entries = nil },
		"duplicate subject":     func(m *RedactionTokenManifest) { m.Entries = append(m.Entries, m.Entries[0]) },
		"invalid family":        func(m *RedactionTokenManifest) { m.Entries[0].SubjectFamily = "host" },
		"invalid ordinal":       func(m *RedactionTokenManifest) { m.Entries[0].DisplayToken = "PARTY-0002" },
		"invalid digest":        func(m *RedactionTokenManifest) { m.Entries[0].CanonicalDisplayValueSHA256 = "foreign" },
		"invalid recipient":     func(m *RedactionTokenManifest) { m.Entries[0].RecipientPartitionRefs = []string{"internal_only"} },
	} {
		t.Run("token subject/"+name, func(t *testing.T) {
			token, _, err := decodePortableObject[RedactionTokenManifest](redaction.TokenManifestJSON, redaction.TokenManifestSHA256)
			if err != nil {
				t.Fatal(err)
			}
			mutate(&token)
			changed := rendered
			changed.Redaction.TokenManifestJSON, err = canonicalJSON(token)
			if err != nil {
				t.Fatal(err)
			}
			changed.Redaction.TokenManifestSHA256 = reportingObjectDigest(token.SchemaID, changed.Redaction.TokenManifestJSON)
			changed.RenderBundle.Manifest.TokenManifestSHA256 = &changed.Redaction.TokenManifestSHA256
			reveal := *redaction.RevealMap
			reveal.TokenManifestSHA256 = changed.Redaction.TokenManifestSHA256
			changed.Redaction.RevealMapJSON, err = canonicalJSON(reveal)
			if err != nil {
				t.Fatal(err)
			}
			changed.Redaction.RevealMapSHA256 = reportingObjectDigest(reveal.SchemaID, changed.Redaction.RevealMapJSON)
			if validateRevealMap(changed) == nil {
				t.Fatal("altered token identity admitted with recomputed digests")
			}
		})
	}
	for name, mutate := range map[string]func(*RedactionManifest){
		"rule trace":    func(m *RedactionManifest) { m.Entries[0].SelectedRuleTrace.RuleID = "foreign" },
		"precedence":    func(m *RedactionManifest) { m.Entries[0].RulePrecedence = "path" },
		"profile":       func(m *RedactionManifest) { m.ProfileSHA256 = "foreign" },
		"not tokenized": func(m *RedactionManifest) { m.Entries[0].Outcome = "unchanged" },
		"missing token": func(m *RedactionManifest) { m.Entries[0].TokenID = nil },
		"foreign token": func(m *RedactionManifest) { v := "tok_foreign"; m.Entries[0].TokenID = &v },
		"missing use":   func(m *RedactionManifest) { m.Entries = m.Entries[:1] },
		"duplicate use": func(m *RedactionManifest) { m.Entries = append(m.Entries, m.Entries[0]) },
	} {
		t.Run("selected rule/"+name, func(t *testing.T) {
			manifest, _, err := decodePortableObject[RedactionManifest](redaction.ManifestJSON, redaction.ManifestSHA256)
			if err != nil {
				t.Fatal(err)
			}
			mutate(&manifest)
			changed := rendered
			changed.RedactionManifestJSON, err = canonicalJSON(manifest)
			if err != nil {
				t.Fatal(err)
			}
			changed.RedactionManifestSHA256 = reportingObjectDigest(manifest.SchemaID, changed.RedactionManifestJSON)
			changed.RenderBundle.Manifest.RedactionManifestSHA256 = changed.RedactionManifestSHA256
			if validateRevealMap(changed) == nil {
				t.Fatal("altered selected-rule inventory admitted")
			}
		})
	}
	assertClosedTokenArtifactShapes(t, redaction)
	assertPublicTokenArtifactBindings(t, rendered)
	for name, raw := range map[string][]byte{
		"missing":        nil,
		"unknown member": bytes.Replace(redaction.RevealMapJSON, []byte(`"schema_id":`), []byte(`"unknown":true,"schema_id":`), 1),
		"whitespace":     append([]byte(" "), redaction.RevealMapJSON...),
	} {
		t.Run("private bytes/"+name, func(t *testing.T) {
			changed := rendered
			changed.Redaction.RevealMapJSON = raw
			if validateRevealMap(changed) == nil {
				t.Fatal("invalid private bytes admitted")
			}
		})
	}
	ordinary := RenderBundleManifestFile{Path: "validation/token-manifest.json", Role: renderBundleRoleTokenManifest, MediaType: "application/vnd.cartulary.reporting-token-manifest+json", RequiredForRelease: true}
	for name, mutate := range map[string]func(*RenderBundleManifestFile){
		"sensitive role":  func(f *RenderBundleManifestFile) { f.Role = "sensitive_reveal_map" },
		"unknown role":    func(f *RenderBundleManifestFile) { f.Role = "future_private_artifact" },
		"internal path":   func(f *RenderBundleManifestFile) { f.Path = "internal/reveal-map.json" },
		"sensitive media": func(f *RenderBundleManifestFile) { f.MediaType = "application/vnd.cartulary.reporting-reveal-map+json" },
	} {
		t.Run("shareable member/"+name, func(t *testing.T) {
			f := ordinary
			mutate(&f)
			if shareableBundleMember(f) {
				t.Fatal("private or unowned member admitted")
			}
		})
	}
}

func assertPublicTokenArtifactBindings(t *testing.T, rendered RenderedRelease) {
	t.Helper()
	r := rendered.Redaction
	files := []RenderBundleFile{
		newBundleFile("validation/redaction-manifest.json", renderBundleRoleRedactionManifest, "application/vnd.cartulary.redaction-manifest+json", r.ManifestJSON, true),
		newBundleFile("validation/redaction-profile-view.json", renderBundleRoleRedactionProfileView, "application/json", r.ProfileViewJSON, true),
		newBundleFile("validation/token-manifest.json", renderBundleRoleTokenManifest, "application/vnd.cartulary.reporting-token-manifest+json", r.TokenManifestJSON, true),
	}
	if err := validatePublicRedactionArtifacts(rendered.RenderBundle.Manifest, files); err != nil {
		t.Fatal("valid public artifact graph rejected", err)
	}
	for _, which := range []string{"retired redaction", "retired tokens", "changed rule", "truncated uses", "unknown profile member", "missing profile", "missing token", "duplicate token"} {
		t.Run("public artifact/"+which, func(t *testing.T) {
			bundle := rendered.RenderBundle.Manifest
			candidate := append([]RenderBundleFile(nil), files...)
			manifest, _, _ := decodePortableObject[RedactionManifest](r.ManifestJSON, r.ManifestSHA256)
			tokens, _, _ := decodePortableObject[RedactionTokenManifest](r.TokenManifestJSON, r.TokenManifestSHA256)
			switch which {
			case "retired redaction":
				manifest.SchemaID = "cartulary.redaction_manifest.v2"
			case "retired tokens":
				tokens.SchemaID = "cartulary.reporting_token_manifest.v2"
			case "changed rule":
				// Keep the entry structurally valid while disagreeing with its
				// frozen profile: integrity of the object alone is insufficient.
				manifest.Entries[0].Action = ActionStub
			case "truncated uses":
				manifest.Entries = manifest.Entries[:0]
			case "unknown profile member":
				candidate[1].Bytes = bytes.Replace(candidate[1].Bytes, []byte(`"schema_id":`), []byte(`"unknown":true,"schema_id":`), 1)
			}
			candidate[0].Bytes, _ = canonicalJSON(manifest)
			bundle.RedactionManifestSHA256 = reportingObjectDigest(manifest.SchemaID, candidate[0].Bytes)
			tokens.RedactionManifestSHA256 = bundle.RedactionManifestSHA256
			candidate[2].Bytes, _ = canonicalJSON(tokens)
			digest := reportingObjectDigest(tokens.SchemaID, candidate[2].Bytes)
			bundle.TokenManifestSHA256 = &digest
			switch which {
			case "missing profile":
				candidate = []RenderBundleFile{candidate[0], candidate[2]}
			case "missing token":
				candidate = candidate[:2]
			case "duplicate token":
				candidate = append(candidate, candidate[2])
			}
			if validatePublicRedactionArtifacts(bundle, candidate) == nil {
				t.Fatal("invalid public artifact admitted with recomputed digests")
			}
		})
	}
}

func assertClosedTokenArtifactShapes(t *testing.T, result RedactionResult) {
	t.Helper()
	assertClosedArtifactShape[RedactionManifest](t, result.ManifestJSON, RedactionManifestSchemaID)
	assertClosedArtifactShape[RedactionTokenManifest](t, result.TokenManifestJSON, RedactionTokenManifestSchemaID)
	assertClosedArtifactShape[RedactionRevealMap](t, result.RevealMapJSON, RedactionRevealMapSchemaID)
}

func assertClosedArtifactShape[T any](t *testing.T, raw []byte, schema string) {
	t.Helper()
	decode := func() map[string]any {
		t.Helper()
		var object map[string]any
		if err := json.Unmarshal(raw, &object); err != nil {
			t.Fatal(err)
		}
		return object
	}
	reject := func(object map[string]any) {
		t.Helper()
		data, err := canonicalJSON(object)
		if err != nil {
			t.Fatal(err)
		}
		if _, _, err := decodePortableObject[T](data, reportingObjectDigest(schema, data)); err == nil {
			t.Fatal("non-exact artifact shape admitted")
		}
	}
	for key := range decode() {
		t.Run(schema+"/missing "+key, func(t *testing.T) { object := decode(); delete(object, key); reject(object) })
	}
	for key := range decode()["entries"].([]any)[0].(map[string]any) {
		t.Run(schema+"/entry missing "+key, func(t *testing.T) {
			object := decode()
			delete(object["entries"].([]any)[0].(map[string]any), key)
			reject(object)
		})
	}
	for _, key := range []string{"unknown", "original_value", "paths", "actions", "subject_kind"} {
		t.Run(schema+"/retired "+key, func(t *testing.T) {
			object := decode()
			object["entries"].([]any)[0].(map[string]any)[key] = "retired"
			reject(object)
		})
	}
}
