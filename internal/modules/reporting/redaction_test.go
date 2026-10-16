package reporting

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/graphprojection"
	"github.com/JochiRaider/cartulary/internal/platform/renderprocess"
)

func TestRedactionProfilePrecedenceActionsAndManifest_Unit(t *testing.T) {
	t.Run("release lifecycle eligibility", testReleaseLifecycleEligibility)
	t.Run("retained sections", testRetainedSections)
	t.Run("retained Core records", testRetainedRecords)
	t.Run("canonical retained model", testRetainedModelBytes)
	t.Run("retained field actions", testRetainedFieldActions)
	t.Run("structural removal evidence", testStructuralRedactionRemovals)
	t.Run("container selection and suppression", testContainerRedaction)
	descriptionPath := "$.sections['sec-0001'].blocks['sec-0001.b0001'].fields['value']"
	sourceClass := ContentClassSourceEvidence
	maxChars := 4
	maskText := "[MASKED]"
	profile := RedactionProfile{
		ProfileStatus: "adopted_current", AllowedReleaseScopes: []string{ReleaseScopeInternalDraft, ReleaseScopeInternalReview, ReleaseScopeExternal},
		SchemaID:  RedactionProfileSchemaID,
		ProfileID: "test.redaction",
		Version:   "1",
		DefaultAction: RedactionActionSpec{
			Type: ActionStub,
		},
		Rules: []RedactionRule{
			{
				RuleID:       "class-source-mask",
				ContentClass: &sourceClass,
				Action: RedactionActionSpec{
					Type:            ActionMask,
					ReplacementText: &maskText,
				},
			},
			{
				RuleID: "path-description-truncate",
				Path:   &descriptionPath,
				Action: RedactionActionSpec{
					Type:     ActionTruncate,
					MaxChars: &maxChars,
				},
			},
		},
	}
	if _, err := ValidateRedactionProfile(profile); err != nil {
		t.Fatalf("validate profile: %v", err)
	}
	model := fixtureStructuredRenderModel(t, []ExportField{
		{
			Path:         "/incident/description",
			ContentClass: ContentClassSourceEvidence,
			Value:        "abcdefgh",
		},
		{
			Path:         "/incident/raw_note",
			ContentClass: ContentClassSourceEvidence,
			Value:        "source value",
		},
		{
			Path:         "/incident/internal_note",
			ContentClass: ContentClassWorkingMaterial,
			Value:        "internal value",
		},
	})
	sourceBytes, err := canonicalJSON(model)
	if err != nil {
		t.Fatalf("source model json: %v", err)
	}
	fixtureAllowContainerRules(t, &profile, model.SnapshotContent)
	profileSHA, err := ValidateRedactionProfile(profile)
	if err != nil {
		t.Fatal(err)
	}
	result, err := RedactExportModel(model, fixtureSourceFactsForContent(model.SnapshotContent), profile, profileSHA, reportingObjectDigest(ExportModelSchemaID, sourceBytes), ReleaseScopeInternalReview, nil)
	if err != nil {
		t.Fatalf("redact model: %v", err)
	}
	values := map[string]any{}
	for _, field := range result.Model.Fields {
		values[fixtureSourcesByModelPath(t, model.SnapshotContent, "/incident/description", "/incident/internal_note", "/incident/raw_note")[field.Path]] = field.Value
	}
	if values["/incident/description"] != "abcd"+DefaultTruncateMark {
		t.Fatalf("path rule must override class rule, got %#v", values["/incident/description"])
	}
	if values["/incident/raw_note"] != maskText {
		t.Fatalf("class rule must override default, got %#v", values["/incident/raw_note"])
	}
	if values["/incident/internal_note"] != DefaultStubText {
		t.Fatalf("default rule must apply when no path or class rule matches, got %#v", values["/incident/internal_note"])
	}
	entries := map[string]RedactionManifestEntry{}
	for _, entry := range result.Manifest.Entries {
		entries[fixtureSourcesByModelPath(t, model.SnapshotContent, "/incident/description", "/incident/internal_note", "/incident/raw_note")[entry.Path]] = entry
	}
	if entries["/incident/description"].RuleID != "path-description-truncate" || entries["/incident/description"].Outcome != "truncated" {
		t.Fatalf("manifest must bind path rule and outcome, got %#v", entries["/incident/description"])
	}
	if entries["/incident/raw_note"].RuleID != "class-source-mask" || entries["/incident/raw_note"].Outcome != "masked" {
		t.Fatalf("manifest must bind class rule and outcome, got %#v", entries["/incident/raw_note"])
	}
	if entries["/incident/internal_note"].RuleID != "profile_default" || entries["/incident/internal_note"].Outcome != "stubbed" {
		t.Fatalf("manifest must bind default rule and outcome, got %#v", entries["/incident/internal_note"])
	}
	if result.Manifest.ProfileSHA256 != profileSHA || result.ManifestSHA256 == "" {
		t.Fatalf("manifest must carry immutable profile digest and manifest digest")
	}
	for _, entry := range result.Manifest.Entries {
		if !validExactExportModelPath(entry.Path) {
			t.Fatal("non-model artifact address")
		}
	}
	if entries["/incident/description"].Path != descriptionPath || bytes.Contains(result.ManifestJSON, []byte(`/incident/`)) || bytes.Contains(result.ManifestJSON, []byte(`"path":`)) {
		t.Fatal("manifest exposed a source selector instead of the exact pre-removal address")
	}
	required, _ := ResolveTemplateContract(DefaultTemplateID, DefaultTemplateVersion)
	required.Sections[0].Bindings = []string{descriptionPath}
	if _, err := deriveTemplateSections(context.Background(), required, result.Model); err != nil {
		t.Fatal("required retained binding did not resolve", err)
	}
	missing := result.Model
	missing.Sections = nil
	if _, err := deriveTemplateSections(context.Background(), required, missing); err == nil {
		t.Fatal("missing retained tree was satisfied by frozen source or flat decision fields")
	}
	if entries["/incident/description"].SelectedRuleTrace.SelectorKind != "path" ||
		entries["/incident/raw_note"].SelectedRuleTrace.SelectorKind != "content_class" ||
		entries["/incident/internal_note"].SelectedRuleTrace.SelectorKind != "profile_default" {
		t.Fatalf("manifest must record selected-rule trace objects, got %#v", result.Manifest.Entries)
	}
	assertRedactionUsesStructuredExportModelBlocks(t)
	assertCanonicalExportModelRoundTrip(t)
	assertRedactionTokensRevealMapAndProfileViewAreCanonical(t)
	assertExplicitSubjectTokenRules(t)
}

func assertRedactionTokensRevealMapAndProfileViewAreCanonical(t *testing.T) {
	t.Helper()
	profile := RedactionProfile{
		ProfileStatus: "adopted_current", AllowedReleaseScopes: []string{ReleaseScopeInternalDraft, ReleaseScopeInternalReview, ReleaseScopeExternal},
		SchemaID:                       RedactionProfileSchemaID,
		ProfileID:                      "test.tokenized",
		Version:                        "1",
		AllowedDisclosurePartitionRefs: []string{"party:party_a"},
		DefaultAction:                  RedactionActionSpec{Type: ActionMask, TokenBacked: true},
	}
	if _, err := ValidateRedactionProfile(profile); err != nil {
		t.Fatalf("validate profile: %v", err)
	}
	model := fixtureStructuredRenderModel(t, []ExportField{
		{
			Path:         "/parties/party_a",
			ContentClass: ContentClassSourceEvidence,
			SourceFamily: "party",
			Value: map[string]any{
				"display_name": "Alice Example",
				"role":         "recipient",
			},
			DisclosurePartitionRefs: []string{"party:party_a"},
		},
	})
	sourceBytes, err := canonicalJSON(model)
	if err != nil {
		t.Fatalf("source model json: %v", err)
	}
	fixtureAllowContainerRules(t, &profile, model.SnapshotContent)
	profileSHA, err := ValidateRedactionProfile(profile)
	if err != nil {
		t.Fatal(err)
	}
	result, err := RedactExportModel(model, fixtureSourceFactsForContent(model.SnapshotContent), profile, profileSHA, reportingObjectDigest(ExportModelSchemaID, sourceBytes), ReleaseScopeInternalReview, nil)
	if err != nil {
		t.Fatalf("redact model: %v", err)
	}
	if len(result.Model.Fields) != 2 || len(result.TokenManifest.Entries) != 1 {
		t.Fatalf("two scalar fields must share one subject token, got %#v", result.Model.Fields)
	}
	for _, field := range result.Model.Fields {
		if field.Value != "PARTY-0001" {
			t.Fatal("subject field did not retain shared token")
		}
	}
	displayToken, ok := result.Model.Fields[0].Value.(string)
	if !ok || displayToken != "PARTY-0001" {
		t.Fatalf("mask/stub subject field must receive a deterministic display token, got %#v", result.Model.Fields[0].Value)
	}
	if result.ProfileView.SchemaID != RedactionProfileViewSchemaID || result.ProfileViewSHA256 == "" || len(result.ProfileViewJSON) == 0 {
		t.Fatalf("redaction result must carry a canonical profile view, got %#v", result)
	}
	if result.ProfileView.ProfileSHA256 != result.Manifest.ProfileSHA256 {
		t.Fatalf("profile view must bind the selected profile digest")
	}
	if result.TokenManifest == nil || result.TokenManifestSHA256 == "" || len(result.TokenManifestJSON) == 0 {
		t.Fatalf("tokenized redaction must carry a token manifest, got %#v", result)
	}
	if result.RevealMap == nil || result.RevealMapSHA256 == "" || len(result.RevealMapJSON) == 0 {
		t.Fatalf("tokenized redaction must carry an internal reveal map, got %#v", result)
	}
	if result.TokenManifest.RedactionManifestSHA256 != result.ManifestSHA256 {
		t.Fatalf("token manifest must bind the complete selected-rule manifest")
	}
	tokenEntry := result.TokenManifest.Entries[0]
	if tokenEntry.DisplayToken != displayToken || tokenEntry.StableSubjectRef != "subject:party:record:party_a" {
		t.Fatalf("token manifest must bind display token to stable subject, got %#v", tokenEntry)
	}
	revealEntry := result.RevealMap.Entries[0]
	if revealEntry.DisplayToken != displayToken || revealEntry.CanonicalDisplayValueSHA256 == "" || revealEntry.CanonicalDisplayValue == "" {
		t.Fatalf("reveal map must retain sensitive original material internally, got %#v", revealEntry)
	}
	if result.Manifest.Entries[0].Outcome != "tokenized" || result.Manifest.Entries[0].SelectedRuleTrace.SelectorKind != "profile_default" {
		t.Fatalf("manifest must record tokenized outcome and rule trace, got %#v", result.Manifest.Entries[0])
	}
	repeated, err := RedactExportModel(model, fixtureSourceFactsForContent(model.SnapshotContent), profile, profileSHA, reportingObjectDigest(ExportModelSchemaID, sourceBytes), ReleaseScopeInternalReview, nil)
	if err != nil {
		t.Fatalf("redact model again: %v", err)
	}
	if repeated.ManifestSHA256 != result.ManifestSHA256 ||
		repeated.TokenManifestSHA256 != result.TokenManifestSHA256 ||
		repeated.RevealMapSHA256 != result.RevealMapSHA256 ||
		repeated.Model.Fields[0].Value != displayToken {
		t.Fatalf("tokenized redaction must be deterministic: first=%#v second=%#v", result, repeated)
	}
}

func assertRedactionUsesStructuredExportModelBlocks(t *testing.T) {
	t.Helper()
	profile := RedactionProfile{
		ProfileStatus: "adopted_current", AllowedReleaseScopes: []string{ReleaseScopeInternalDraft, ReleaseScopeInternalReview, ReleaseScopeExternal},
		SchemaID:      RedactionProfileSchemaID,
		ProfileID:     "test.structured",
		Version:       "1",
		DefaultAction: RedactionActionSpec{Type: ActionAllow},
	}
	profileSHA, err := ValidateRedactionProfile(profile)
	if err != nil {
		t.Fatalf("validate profile: %v", err)
	}
	model, sourceSHA, err := buildStructuredTestExportModel(
		"incident-1",
		fixtureSnapshotID,
		time.Date(2026, 5, 23, 12, 0, 0, 0, time.UTC),
		SourceBoundaryTokenPrefix+strings.Repeat("a", 64),
		ReleaseScopeInternalReview,
		nil,
		[]ExportField{{
			Path:         "/incident/title",
			ContentClass: ContentClassCuratedNarrative,
			SourceFamily: "incident_metadata",
			Value:        "structured title",
			SupportRefs:  []string{"/incident/status"},
		}},
	)
	if err != nil {
		t.Fatalf("build structured export model: %v", err)
	}

	result, err := RedactExportModel(model, fixtureSourceFactsForContent(model.SnapshotContent), profile, profileSHA, sourceSHA, ReleaseScopeInternalReview, nil)
	if err != nil {
		t.Fatalf("redact model: %v", err)
	}
	if len(result.Model.Fields) != 1 || result.Model.Fields[0].Value != "structured title" {
		t.Fatalf("redaction must use structured export model blocks, got %#v", result.Model.Fields)
	}
	raw, err := canonicalJSON(model)
	if err != nil {
		t.Fatal(err)
	}
	var object map[string]any
	if err := json.Unmarshal(raw, &object); err != nil {
		t.Fatal(err)
	}
	object["fields"] = []any{map[string]any{"path": "/incident/title", "value": "legacy cache title"}}
	legacy, err := canonicalJSON(object)
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := decodePortableObject[ExportModel](legacy, reportingObjectDigest(ExportModelSchemaID, legacy)); err == nil {
		t.Fatal("legacy flat fields were admitted as canonical structured content")
	}
}

func TestRedactionProfileRejectsConflictsHashAndUnsafeBounds_Unit(t *testing.T) {
	t.Run("closed class selectors", testRedactionContentClassSelectors)
	t.Run("closed action parameters", testRedactionActionParameters)
	t.Run("normalized profile contract", testNormalizedRedactionProfile)
	path := "$.sections['sec-0001'].blocks['sec-0001.b0001'].fields['value']"
	class := ContentClassCuratedNarrative
	maxZero := 0
	cases := []struct {
		name    string
		profile RedactionProfile
	}{
		{
			name: "duplicate path rules",
			profile: RedactionProfile{
				ProfileStatus: "adopted_current", AllowedReleaseScopes: []string{ReleaseScopeInternalDraft, ReleaseScopeInternalReview, ReleaseScopeExternal},
				SchemaID:      RedactionProfileSchemaID,
				ProfileID:     "test.duplicate.path",
				Version:       "1",
				DefaultAction: RedactionActionSpec{Type: ActionAllow},
				Rules: []RedactionRule{
					{RuleID: "a", Path: &path, Action: RedactionActionSpec{Type: ActionAllow}},
					{RuleID: "b", Path: &path, Action: RedactionActionSpec{Type: ActionDrop}},
				},
			},
		},
		{
			name: "duplicate content class rules",
			profile: RedactionProfile{
				ProfileStatus: "adopted_current", AllowedReleaseScopes: []string{ReleaseScopeInternalDraft, ReleaseScopeInternalReview, ReleaseScopeExternal},
				SchemaID:      RedactionProfileSchemaID,
				ProfileID:     "test.duplicate.class",
				Version:       "1",
				DefaultAction: RedactionActionSpec{Type: ActionAllow},
				Rules: []RedactionRule{
					{RuleID: "a", ContentClass: &class, Action: RedactionActionSpec{Type: ActionAllow}},
					{RuleID: "b", ContentClass: &class, Action: RedactionActionSpec{Type: ActionDrop}},
				},
			},
		},
		{
			name: "reserved hash action",
			profile: RedactionProfile{
				ProfileStatus: "adopted_current", AllowedReleaseScopes: []string{ReleaseScopeInternalDraft, ReleaseScopeInternalReview, ReleaseScopeExternal},
				SchemaID:      RedactionProfileSchemaID,
				ProfileID:     "test.hash",
				Version:       "1",
				DefaultAction: RedactionActionSpec{Type: ActionHash},
			},
		},
		{
			name: "truncate missing safe bound",
			profile: RedactionProfile{
				ProfileStatus: "adopted_current", AllowedReleaseScopes: []string{ReleaseScopeInternalDraft, ReleaseScopeInternalReview, ReleaseScopeExternal},
				SchemaID:      RedactionProfileSchemaID,
				ProfileID:     "test.truncate",
				Version:       "1",
				DefaultAction: RedactionActionSpec{Type: ActionTruncate, MaxChars: &maxZero},
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := ValidateRedactionProfile(tc.profile); !errors.Is(err, ErrInvalidRedactionProfile) {
				t.Fatalf("expected ErrInvalidRedactionProfile, got %v", err)
			}
		})
	}
}

func TestExternalValidationRejectsOpaqueBytesAndWorkingMaterial_Unit(t *testing.T) {
	profile := internalRedactionProfile()
	// Exercise content eligibility with a profile explicitly permitting external scope.
	profile.AllowedReleaseScopes = append(profile.AllowedReleaseScopes, ReleaseScopeExternal)
	profileSHA, err := ValidateRedactionProfile(profile)
	if err != nil {
		t.Fatal(err)
	}
	model := fixtureStructuredRenderModel(t, []ExportField{{Path: "/incident/internal_note", ContentClass: ContentClassWorkingMaterial, Value: "internal"}})
	sourceBytes, err := canonicalJSON(model)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := RedactExportModel(model, fixtureSourceFactsForContent(model.SnapshotContent), profile, profileSHA, reportingObjectDigest(ExportModelSchemaID, sourceBytes), ReleaseScopeExternal, nil); !errors.Is(err, ErrRedactionValidation) {
		t.Fatalf("external working material was retained: %v", err)
	}
	for name, field := range map[string]ExportField{
		"raw source":        {RawBlobSource: true, Value: "encoded raw data"},
		"opaque bytes":      {OpaqueBinary: true, Value: "encoded bytes"},
		"generated content": {GeneratedPresentation: true, Value: "presentation"},
		"binary value":      {Value: []byte{1, 2, 3}},
	} {
		t.Run(name, func(t *testing.T) {
			field.Path = "/evidence/source"
			field.SourceFamily = "evidence"
			field.SourceRecordID = "source"
			field.SourceRecordType = "evidence"
			field.ContentClass = ContentClassSourceEvidence
			model, digest, err := buildStructuredSnapshotModel("incident", fixtureSnapshotID, time.Date(2026, 5, 23, 12, 0, 0, 0, time.UTC), "boundary", fixtureSnapshotSource([]ExportField{field}))
			if err == nil || digest != "" || model.SchemaID != "" {
				t.Fatal("forbidden source survived snapshot materialization")
			}
		})
	}
}

func TestDisclosurePartitionsAndCuratedSupportRefsFailClosed_Unit(t *testing.T) {
	t.Run("effective scope authority", testEffectiveDisclosurePartitions)
	allowedPath := "/incident/summary"
	restrictedPath := "/incident/restricted"
	var allowedModelPath, restrictedModelPath string
	profile := RedactionProfile{
		ProfileStatus: "adopted_current", AllowedReleaseScopes: []string{ReleaseScopeInternalDraft, ReleaseScopeInternalReview, ReleaseScopeExternal},
		SchemaID:  RedactionProfileSchemaID,
		ProfileID: "test.partitions",
		Version:   "1",
		AllowedDisclosurePartitionRefs: []string{
			"public",
		},
		DefaultAction: RedactionActionSpec{Type: ActionAllow},
		Rules: []RedactionRule{
			{RuleID: "allow-summary", Path: &allowedModelPath, Action: RedactionActionSpec{Type: ActionAllow}},
			{RuleID: "allow-restricted", Path: &restrictedModelPath, Action: RedactionActionSpec{Type: ActionAllow}},
		},
	}
	model := fixtureStructuredRenderModel(t, []ExportField{
		{
			Path:                    allowedPath,
			ContentClass:            ContentClassCuratedNarrative,
			Value:                   "supported public summary",
			DisclosurePartitionRefs: []string{"public"},
			SupportRefs:             []string{"/evidence/public-source"},
		},
		{
			Path:                    restrictedPath,
			ContentClass:            ContentClassDerivedAnalytic,
			Value:                   "restricted analytic",
			DisclosurePartitionRefs: []string{"internal_only"},
		},
		{
			Path: "/evidence/public-source", ContentClass: ContentClassSourceEvidence,
			Value: "public source", DisclosurePartitionRefs: []string{"public"},
		},
	})
	modelPaths := fixtureModelPaths(t, model.SnapshotContent, "/evidence/public-source", restrictedPath, allowedPath)
	allowedModelPath, restrictedModelPath = modelPaths[allowedPath], modelPaths[restrictedPath]
	sourceBytes, err := canonicalJSON(model)
	if err != nil {
		t.Fatalf("source model json: %v", err)
	}
	// Every selected retaining action must reject an unresolved partition. An
	// implementation must not replace the selected action with a silent drop.
	maxChars := 3
	for _, action := range []RedactionActionSpec{
		{Type: ActionAllow}, {Type: ActionMask}, {Type: ActionStub},
		{Type: ActionTruncate, MaxChars: &maxChars},
	} {
		for _, selection := range []string{"path", "content_class", "profile_default"} {
			t.Run(action.Type+"/"+selection, func(t *testing.T) {
				candidate := profile
				candidate.Rules = append([]RedactionRule(nil), profile.Rules[:1]...)
				selected := RedactionRule{RuleID: "restricted-rule", Action: action}
				switch selection {
				case "path":
					selected.Path = &restrictedModelPath
					candidate.Rules = append(candidate.Rules, selected)
				case "content_class":
					class := ContentClassDerivedAnalytic
					selected.ContentClass = &class
					candidate.Rules = append(candidate.Rules, selected)
				case "profile_default":
					candidate.DefaultAction = action
				}
				fixtureAllowContainerRules(t, &candidate, model.SnapshotContent)
				digest, err := ValidateRedactionProfile(candidate)
				if err != nil {
					t.Fatal(err)
				}
				result, err := RedactExportModel(model, fixtureSourceFactsForContent(model.SnapshotContent), candidate, digest, reportingObjectDigest(ExportModelSchemaID, sourceBytes), ReleaseScopeExternal, nil)
				var validation *renderValidationError
				if !errors.As(err, &validation) || validation.FailureCode != "redaction_manifest_invalid" || validation.ReasonCode != "disclosure_partition_unresolved" {
					t.Fatalf("unresolved selected rule must fail with the closed disclosure reason, got %v", err)
				}
				if len(result.Model.Fields) != 0 || result.ModelSHA256 != "" || len(result.Manifest.Entries) != 0 {
					t.Fatal("failed redaction published partial content or success evidence")
				}
				if strings.Contains(err.Error(), "restricted analytic") {
					t.Fatal("redaction failure disclosed the source value")
				}
			})
		}
	}

	// A drop remains valid only when it is the selected rule, with its original
	// rule identity, action and trace retained in the manifest.
	profile.Rules[1] = RedactionRule{RuleID: "drop-restricted", Path: &restrictedModelPath, Action: RedactionActionSpec{Type: ActionDrop}}
	profileSHA, err := ValidateRedactionProfile(profile)
	if err != nil {
		t.Fatal(err)
	}
	result, err := RedactExportModel(model, fixtureSourceFactsForContent(model.SnapshotContent), profile, profileSHA, reportingObjectDigest(ExportModelSchemaID, sourceBytes), ReleaseScopeExternal, nil)
	if err != nil {
		t.Fatalf("redact with partition filter: %v", err)
	}
	if len(result.Model.Fields) != 2 || fixtureSourcesByModelPath(t, model.SnapshotContent, "/evidence/public-source", restrictedPath, allowedPath)[result.Model.Fields[1].Path] != allowedPath {
		t.Fatalf("selected drop must remove only the restricted field, got %#v", result.Model.Fields)
	}
	entries := map[string]RedactionManifestEntry{}
	for _, entry := range result.Manifest.Entries {
		entries[fixtureSourcesByModelPath(t, model.SnapshotContent, "/evidence/public-source", restrictedPath, allowedPath)[entry.Path]] = entry
	}
	entry := entries[restrictedPath]
	if entry.Outcome != "dropped" || entry.Action != ActionDrop || entry.RuleID != "drop-restricted" || entry.SelectedRuleTrace.RuleID != "drop-restricted" {
		t.Fatalf("selected drop identity must be preserved, got %#v", entry)
	}

	// Materialize an unsupported narrative, rather than corrupting a frozen
	// field's required array shape. This exercises the redaction support gate.
	unsupported := fixtureSourceFields(t, model.SnapshotContent)
	changed := false
	for i := range unsupported {
		if unsupported[i].Value == "supported public summary" {
			unsupported[i].SupportRefs = []string{}
			changed = true
		}
	}
	if !changed {
		t.Fatal("curated source field was not found")
	}
	model = fixtureStructuredRenderModel(t, unsupported)
	sourceBytes, err = canonicalJSON(model)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := RedactExportModel(model, fixtureSourceFactsForContent(model.SnapshotContent), profile, profileSHA, reportingObjectDigest(ExportModelSchemaID, sourceBytes), ReleaseScopeExternal, nil); !errors.Is(err, ErrRedactionValidation) {
		t.Fatalf("external curated narrative without support refs must fail closed, got %v", err)
	}
}

func TestBuildExportModelUsesReportingOwnedMetadataSnapshotStableHash(t *testing.T) {
	t.Run("canonical bytes and hash domains", testReportingCanonicalBytes)
	description := "Stable public summary"
	severity := "high"
	tlp := "TLP:AMBER"
	currentPhase := "containment"
	incident := IncidentMetadataSnapshot{
		ID:           "00000000-0000-0000-0000-000000000011",
		Title:        "Stable export model",
		Description:  &description,
		Status:       "active",
		Severity:     &severity,
		TLP:          &tlp,
		CurrentPhase: &currentPhase,
		Version:      42,
	}
	snapshotAt := time.Date(2026, 5, 23, 12, 0, 0, 123000, time.FixedZone("test", -5*60*60))
	watermark := SourceBoundaryTokenPrefix + strings.Repeat("4", 64)
	workbookFields := []ExportField{
		{
			Path:           "/workbook/rows/indicator-1/value",
			ContentClass:   ContentClassSourceEvidence,
			SourceFamily:   "indicator",
			SourceRecordID: "indicator-1", SourceRecordType: "indicator",
			Value: "example.test",
		},
	}

	first, firstSHA, err := BuildSnapshotModel(incident, fixtureSnapshotID, snapshotAt, watermark, fixtureSnapshotSource(fixtureOwnerFields(workbookFields)))
	if err != nil {
		t.Fatalf("build first export model: %v", err)
	}
	second, secondSHA, err := BuildSnapshotModel(incident, fixtureSnapshotID, snapshotAt, watermark, fixtureSnapshotSource(fixtureOwnerFields(workbookFields)))
	if err != nil {
		t.Fatalf("build second export model: %v", err)
	}
	firstJSON, err := canonicalJSON(first)
	if err != nil {
		t.Fatalf("first canonical json: %v", err)
	}
	secondJSON, err := canonicalJSON(second)
	if err != nil {
		t.Fatalf("second canonical json: %v", err)
	}
	if firstSHA != reportingObjectDigest(SnapshotModelSchemaID, firstJSON) {
		t.Fatalf("export model sha must be derived from canonical reporting model json")
	}
	if firstSHA != secondSHA || string(firstJSON) != string(secondJSON) {
		t.Fatalf("identical reporting metadata snapshots must produce stable export models: first=%s second=%s", firstSHA, secondSHA)
	}
	if first.IncidentID != incident.ID || !first.SnapshotAt.Equal(snapshotAt.UTC()) {
		t.Fatalf("export model must preserve reporting-owned incident id and normalized snapshot time: %#v", first)
	}
	paths := map[string]any{}
	redactionFields, err := first.RedactionFields()
	if err != nil {
		t.Fatal(err)
	}
	sourcePaths := fixtureSourcesByModelPath(t, first.SnapshotContent, "/incident/current_phase", "/incident/description", "/incident/severity", "/incident/status", "/incident/title", "/incident/tlp", "/workbook/rows/indicator-1/value")
	modelPaths, _ := redactionFieldModelPaths(first.SnapshotContent)
	for _, field := range redactionFields {
		paths[sourcePaths[modelPaths[field.Path]]] = field.Value
	}
	for path, want := range map[string]any{
		"/incident/title":                  incident.Title,
		"/incident/status":                 incident.Status,
		"/incident/description":            description,
		"/incident/severity":               severity,
		"/incident/tlp":                    tlp,
		"/incident/current_phase":          currentPhase,
		"/workbook/rows/indicator-1/value": "example.test",
	} {
		if paths[path] != want {
			t.Fatalf("export model field %s = %#v, want %#v", path, paths[path], want)
		}
	}
}

func assertCanonicalExportModelRoundTrip(t *testing.T) {
	t.Helper()
	description := "Golden summary"
	incident := IncidentMetadataSnapshot{
		ID:          "00000000-0000-0000-0000-000000000099",
		Title:       "Golden incident",
		Description: &description,
		Status:      "active",
		Version:     7,
	}
	model, modelSHA, err := BuildSnapshotModel(
		incident, fixtureSnapshotID,
		time.Date(2026, 5, 23, 12, 0, 0, 0, time.UTC),
		SourceBoundaryTokenPrefix+strings.Repeat("7", 64), fixtureSnapshotSource([]ExportField{{
			SourceRecordID: "tl-1", SourceRecordType: "timeline_event",
			Path:         "/timeline/tl-1",
			ContentClass: ContentClassSourceEvidence,
			SourceFamily: "timeline_event",
			Value:        map[string]any{"capture_state": "rough", "value": "First observed event"},
			SupportRefs:  []string{"/timeline/tl-1"},
		}}),
	)
	if err != nil {
		t.Fatalf("build export model: %v", err)
	}
	canonical, err := canonicalJSON(model)
	if err != nil {
		t.Fatalf("canonical export model: %v", err)
	}
	if modelSHA != reportingObjectDigest(SnapshotModelSchemaID, canonical) {
		t.Fatalf("model sha = %s, canonical hash = %s", modelSHA, reportingObjectDigest(SnapshotModelSchemaID, canonical))
	}
	var restored SnapshotModel
	if err := json.Unmarshal(canonical, &restored); err != nil {
		t.Fatal(err)
	}
	reencoded, err := canonicalJSON(restored)
	if err != nil || !bytes.Equal(reencoded, canonical) {
		t.Fatal("complete model changed after JSON persistence", err)
	}
	// Canonical byte order, scalar admission and the schema hash domain have
	// independent literal vectors in testReportingCanonicalBytes. The retired
	// golden bound Go struct declaration order rather than the owner algorithm.

}

func TestDecoderNormalizationAndRegisteredReasons_Unit(t *testing.T) {
	t.Run("internal profile omission", testInternalProfileRequestDefault)
	_, unknownAliasErr := DecodeCreateSnapshotRequest(strings.NewReader(`{"incident_id":"00000000-0000-0000-0000-000000000001","client_txn_id":"txn","source_change_set_high_watermark_id":"legacy"}`))
	if unknownAliasErr == nil || unknownAliasErr.Details["reason_code"] != "unknown_field" {
		t.Fatalf("legacy source boundary alias must be rejected as unknown field, got %#v", unknownAliasErr)
	}
	_, malformed := DecodeCreateReleaseRequest(strings.NewReader(`[]`))
	if malformed == nil || malformed.Details["reason_code"] != "request_not_object" {
		t.Fatalf("non-object release body must use registered request_not_object reason, got %#v", malformed)
	}
	_, duplicateMember := DecodeCreateReleaseRequest(strings.NewReader(`{
		"snapshot_id":"00000000-0000-0000-0000-000000000001",
		"snapshot_id":"00000000-0000-0000-0000-000000000002",
		"client_txn_id":"txn-release-duplicate",
		"template_id":"cartulary.report.default",
		"template_version":"1",
		"redaction_profile_id":"cartulary.redaction.internal",
		"redaction_profile_version":"5",
		"output_kind":"slidev"
	}`))
	if duplicateMember == nil || duplicateMember.Details["reason_code"] != "duplicate_object_member" {
		t.Fatalf("duplicate JSON members must fail before normalization, got %#v", duplicateMember)
	}
	emptyReason, apiErr := DecodeReleaseActionRequest(strings.NewReader(`{"client_txn_id":"txn","reason":""}`))
	if apiErr != nil {
		t.Fatalf("empty reason should normalize, got %v", apiErr)
	}
	nullReason, apiErr := DecodeReleaseActionRequest(strings.NewReader(`{"client_txn_id":"txn","reason":null}`))
	if apiErr != nil {
		t.Fatalf("null reason should normalize, got %v", apiErr)
	}
	omittedReason, apiErr := DecodeReleaseActionRequest(strings.NewReader(`{"client_txn_id":"txn"}`))
	if apiErr != nil {
		t.Fatalf("omitted reason should normalize, got %v", apiErr)
	}
	if string(emptyReason.Normalized) != string(nullReason.Normalized) || string(emptyReason.Normalized) != string(omittedReason.Normalized) {
		t.Fatalf("omitted, null, and empty reasons must compare equal: empty=%s null=%s omitted=%s", emptyReason.Normalized, nullReason.Normalized, omittedReason.Normalized)
	}
	partitioned, apiErr := DecodeCreateReleaseRequest(strings.NewReader(`{
		"snapshot_id":"00000000-0000-0000-0000-000000000001",
		"client_txn_id":"txn-release",
		"template_id":"cartulary.report.default",
		"template_version":"1",
		"redaction_profile_id":"cartulary.redaction.external",
		"redaction_profile_version":"5",
		"output_kind":"slidev",
		"release_scope":"external_release",
		"recipient_partition_refs":["party:b","party:a","party:b"]
	}`))
	if apiErr != nil {
		t.Fatalf("valid recipient partitions should decode, got %v", apiErr)
	}
	if got := strings.Join(partitioned.RecipientPartitionRefs, ","); got != "party:a,party:b" {
		t.Fatalf("recipient partitions must coalesce and sort, got %q", got)
	}
	var defaultOptions map[string]any
	if err := json.Unmarshal(partitioned.OutputOptions, &defaultOptions); err != nil {
		t.Fatalf("decode materialized output options: %v", err)
	}
	if defaultOptions["schema_id"] != OutputOptionsSchemaID || defaultOptions["pdf"] != true || defaultOptions["rendered_diagrams"] != true {
		t.Fatalf("omitted output_options must materialize slidev defaults, got %#v", defaultOptions)
	}
	graphSHA := strings.Repeat("a", 64)
	withTuple, apiErr := DecodeCreateReleaseRequest(strings.NewReader(`{
		"snapshot_id":"00000000-0000-0000-0000-000000000001",
		"client_txn_id":"txn-release-tuple",
		"template_id":"cartulary.report.default",
		"template_version":"1",
		"redaction_profile_id":"cartulary.redaction.internal",
		"redaction_profile_version":"5",
		"output_kind":"slidev",
		"output_options":{"source_only":true},
		"graph_projection_refs":[{
			"source_owner_id":"network_flow_activity",
			"graph_view_id":"gv_a",
			"projection_result_id":"gpres_` + graphSHA + `",
			"source_snapshot_id":"00000000-0000-0000-0000-000000000001",
			"projection_schema_id":"graph_projection.v2",
			"projection_version":"v2",
			"normalized_configuration_sha256":"` + graphSHA + `",
			"normalized_source_sha256":"` + graphSHA + `",
			"canonical_output_sha256":"` + graphSHA + `"
		}],
		"composition_id":"00000000-0000-0000-0000-000000000002",
		"composition_version":"v2",
		"composition_sha256":"` + strings.Repeat("b", 64) + `"
	}`))
	if apiErr != nil {
		t.Fatalf("valid release tuple should decode, got %v", apiErr)
	}
	var sourceOnlyOptions map[string]any
	if err := json.Unmarshal(withTuple.OutputOptions, &sourceOnlyOptions); err != nil {
		t.Fatalf("decode source-only output options: %v", err)
	}
	if sourceOnlyOptions["source_only"] != true || sourceOnlyOptions["pdf"] != false || sourceOnlyOptions["rendered_diagrams"] != false {
		t.Fatalf("source_only must force render output options false, got %#v", sourceOnlyOptions)
	}
	if withTuple.CompositionID == nil || withTuple.CompositionVersion == nil || withTuple.CompositionSHA256 == nil {
		t.Fatalf("composition tuple not decoded: %#v", withTuple)
	}
	if *withTuple.CompositionVersion != "v2" || *withTuple.CompositionSHA256 != strings.Repeat("b", 64) {
		t.Fatalf("composition tuple not decoded: %#v", withTuple)
	}
	_, partialComposition := DecodeCreateReleaseRequest(strings.NewReader(`{
		"snapshot_id":"00000000-0000-0000-0000-000000000001",
		"client_txn_id":"txn-release-partial-composition",
		"template_id":"cartulary.report.default",
		"template_version":"1",
		"redaction_profile_id":"cartulary.redaction.internal",
		"redaction_profile_version":"5",
		"output_kind":"slidev",
		"composition_id":"00000000-0000-0000-0000-000000000002"
	}`))
	if partialComposition == nil || partialComposition.Details["reason_code"] != "composition_tuple_incomplete" {
		t.Fatalf("partial composition tuple must fail closed, got %#v", partialComposition)
	}
	_, sourceOnlyExternal := DecodeCreateReleaseRequest(strings.NewReader(`{
		"snapshot_id":"00000000-0000-0000-0000-000000000001",
		"client_txn_id":"txn-source-only-external",
		"template_id":"cartulary.report.default",
		"template_version":"1",
		"redaction_profile_id":"cartulary.redaction.external",
		"redaction_profile_version":"5",
		"output_kind":"slidev",
		"release_scope":"external_release",
		"output_options":{"source_only":true}
	}`))
	if sourceOnlyExternal == nil || sourceOnlyExternal.Details["reason_code"] != "source_only_external_release_invalid" {
		t.Fatalf("external source_only must fail closed, got %#v", sourceOnlyExternal)
	}
	_, duplicateGraphRef := DecodeCreateReleaseRequest(strings.NewReader(`{
		"snapshot_id":"00000000-0000-0000-0000-000000000001",
		"client_txn_id":"txn-duplicate-graph",
		"template_id":"cartulary.report.default",
		"template_version":"1",
		"redaction_profile_id":"cartulary.redaction.internal",
		"redaction_profile_version":"5",
		"output_kind":"slidev",
		"graph_projection_refs":[
			{"source_owner_id":"network_flow_activity","graph_view_id":"gv_a","projection_result_id":"gpres_` + graphSHA + `","source_snapshot_id":"s","projection_schema_id":"graph_projection.v2","projection_version":"v2","normalized_configuration_sha256":"` + graphSHA + `","normalized_source_sha256":"` + graphSHA + `","canonical_output_sha256":"` + graphSHA + `"},
			{"source_owner_id":"network_flow_activity","graph_view_id":"gv_a","projection_result_id":"gpres_` + strings.Repeat("b", 64) + `","source_snapshot_id":"s","projection_schema_id":"graph_projection.v2","projection_version":"v2","normalized_configuration_sha256":"` + graphSHA + `","normalized_source_sha256":"` + graphSHA + `","canonical_output_sha256":"` + graphSHA + `"}
		]
	}`))
	if duplicateGraphRef == nil || duplicateGraphRef.Details["reason_code"] != "graph_projection_ambiguous" {
		t.Fatalf("duplicate graph view ref must fail closed, got %#v", duplicateGraphRef)
	}
	_, nullPartitions := DecodeCreateReleaseRequest(strings.NewReader(`{
		"snapshot_id":"00000000-0000-0000-0000-000000000001",
		"client_txn_id":"txn-release-null",
		"template_id":"cartulary.report.default",
		"template_version":"1",
		"redaction_profile_id":"cartulary.redaction.external",
		"redaction_profile_version":"5",
		"output_kind":"slidev",
		"release_scope":"external_release",
		"recipient_partition_refs":null
	}`))
	if nullPartitions == nil || nullPartitions.Details["reason_code"] != "field_not_nullable" {
		t.Fatalf("recipient_partition_refs null must be rejected as non-nullable, got %#v", nullPartitions)
	}
	internalPartitionRequest, apiErr := DecodeCreateReleaseRequest(strings.NewReader(`{
		"snapshot_id":"00000000-0000-0000-0000-000000000001",
		"client_txn_id":"txn-release-internal-partitions",
		"template_id":"cartulary.report.default",
		"template_version":"1",
		"redaction_profile_id":"cartulary.redaction.internal",
		"redaction_profile_version":"5",
		"output_kind":"slidev",
		"recipient_partition_refs":["party:a"]
	}`))
	if apiErr != nil {
		t.Fatalf("structural release decode should not perform hidden-resource selector validation, got %#v", apiErr)
	}
	templates, err := NewTemplatePackRegistry()
	if err != nil {
		t.Fatal(err)
	}
	_, internalPartitions := validateCreateReleaseRequestSemantics(internalPartitionRequest, templates)
	if internalPartitions == nil || internalPartitions.Details["reason_code"] != "recipient_partitions_not_allowed" {
		t.Fatalf("internal recipient partitions must use closed rejection reason, got %#v", internalPartitions)
	}

	recipientModel := fixtureStructuredRenderModel(t, []ExportField{
		{Path: "/parties/party_a", ContentClass: ContentClassSourceEvidence, SourceFamily: "party", DisclosurePartitionRefs: []string{"party:party_a"}},
	})
	recipientValidationCases := []struct {
		name       string
		request    CreateReleaseRequest
		wantReason string
	}{
		{
			name: "external requires non-empty recipient set",
			request: CreateReleaseRequest{
				ReleaseScope:            ReleaseScopeExternal,
				RedactionProfileID:      ExternalRedactionProfileID,
				RedactionProfileVersion: DefaultRedactionProfileVersion,
			},
			wantReason: "recipient_partition_profile_mismatch",
		},
		{
			name: "external recipient ref grammar",
			request: CreateReleaseRequest{
				ReleaseScope:            ReleaseScopeExternal,
				RedactionProfileID:      ExternalRedactionProfileID,
				RedactionProfileVersion: DefaultRedactionProfileVersion,
				RecipientPartitionRefs:  []string{"party:bad/ref"},
			},
			wantReason: "invalid_recipient_partition_ref",
		},
		{
			name: "external recipient must resolve in snapshot",
			request: CreateReleaseRequest{
				ReleaseScope:            ReleaseScopeExternal,
				RedactionProfileID:      ExternalRedactionProfileID,
				RedactionProfileVersion: DefaultRedactionProfileVersion,
				RecipientPartitionRefs:  []string{"party:missing"},
			},
			wantReason: "unknown_recipient_partition",
		},
		{
			name: "profile allowed party subset must match recipient set",
			request: CreateReleaseRequest{
				ReleaseScope:            ReleaseScopeExternal,
				RedactionProfileID:      InternalRedactionProfileID,
				RedactionProfileVersion: DefaultRedactionProfileVersion,
				RecipientPartitionRefs:  []string{"party:party_a"},
			},
			wantReason: "recipient_partition_profile_mismatch",
		},
	}
	for _, tc := range recipientValidationCases {
		t.Run(tc.name, func(t *testing.T) {
			profile, err := selectedRedactionProfile(tc.request.RedactionProfileID, tc.request.RedactionProfileVersion, tc.request.RecipientPartitionRefs)
			if err != nil {
				t.Fatal(err)
			}
			failure := validateRenderRecipientPartitions(tc.request.ReleaseScope, tc.request.RecipientPartitionRefs, recipientModel.SnapshotContent, profile)
			if failure == nil || failure.ReasonCode != tc.wantReason {
				t.Fatalf("recipient validation reason = %#v want %q", failure, tc.wantReason)
			}
			if err := validateRenderFailure(renderFailureDetails(failure)); err != nil {
				t.Fatal("recipient failure cannot cross the terminal Job boundary", err)
			}
		})
	}
	validRecipientRequest := CreateReleaseRequest{
		ReleaseScope:            ReleaseScopeExternal,
		RedactionProfileID:      ExternalRedactionProfileID,
		RedactionProfileVersion: DefaultRedactionProfileVersion,
		RecipientPartitionRefs:  []string{"party:party_a"},
	}
	retainedProfile, err := selectedRedactionProfile(validRecipientRequest.RedactionProfileID, validRecipientRequest.RedactionProfileVersion, validRecipientRequest.RecipientPartitionRefs)
	if err != nil {
		t.Fatal(err)
	}
	if err := validateRenderRecipientPartitions(validRecipientRequest.ReleaseScope, validRecipientRequest.RecipientPartitionRefs, recipientModel.SnapshotContent, retainedProfile); err != nil {
		t.Fatalf("valid recipient partition rejected: %#v", err)
	}
	retainedProfile.AllowedDisclosurePartitionRefs = []string{"public"}
	if failure := validateRenderRecipientPartitions(validRecipientRequest.ReleaseScope, validRecipientRequest.RecipientPartitionRefs, recipientModel.SnapshotContent, retainedProfile); failure == nil || failure.ReasonCode != "recipient_partition_profile_mismatch" {
		t.Fatal("recipient validation substituted the current registry for the retained policy", failure)
	}

	contract, ok := ResolveTemplateContract(DefaultTemplateID, DefaultTemplateVersion)
	if !ok {
		t.Fatal("default template contract must resolve")
	}
	baseRequest := CreateReleaseRequest{
		SnapshotID:              mustTestUUID("00000000-0000-0000-0000-000000000001"),
		ClientTxnID:             "txn-render-reasons",
		OutputOptions:           fixtureOutputOptions(OutputKindSlidev, ReleaseScopeInternalDraft),
		TemplateID:              DefaultTemplateID,
		TemplateVersion:         DefaultTemplateVersion,
		RedactionProfileID:      InternalRedactionProfileID,
		RedactionProfileVersion: DefaultRedactionProfileVersion,
		OutputKind:              OutputKindSlidev,
		ReleaseScope:            ReleaseScopeInternalDraft,
	}
	baseModel := fixtureStructuredRenderModel(t, []ExportField{
		{Path: "/incident/status", ContentClass: ContentClassDerivedAnalytic, Value: "open"},
		{Path: "/incident/title", ContentClass: ContentClassCuratedNarrative, Value: "Renderable", SupportRefs: []string{"/incident/status"}},
	})
	selfContainedCases := []struct {
		name   string
		kind   string
		output string
	}{
		{name: "remote script", kind: OutputKindSlidev, output: `<script src="https://cdn.example.test/app.js"></script>`},
		{name: "remote stylesheet", kind: OutputKindSlidev, output: `<link rel="stylesheet" href="//cdn.example.test/app.css">`},
		{name: "remote font css import", kind: OutputKindSlidev, output: `<style>@import url("https://fonts.example.test/font.css");</style>`},
		{name: "remote css url", kind: OutputKindSlidev, output: `<style>body{background:url(https://cdn.example.test/bg.png)}</style>`},
		{name: "remote image", kind: OutputKindSlidev, output: `<img src="https://cdn.example.test/a.png">`},
		{name: "slidev remote image", kind: OutputKindSlidev, output: `![remote](https://cdn.example.test/a.png)`},
		{name: "slidev raw remote image", kind: OutputKindSlidev, output: `<img src="https://cdn.example.test/a.png">`},
	}
	for _, tc := range selfContainedCases {
		t.Run("self contained rejects "+tc.name, func(t *testing.T) {
			if err := ValidateSelfContainedOutput(tc.kind, []byte(tc.output)); !errors.Is(err, ErrRemoteRuntimeAsset) {
				t.Fatalf("remote output asset should fail closed, got %v", err)
			}
		})
	}
	t.Run("self contained allows plain escaped URL text", func(t *testing.T) {
		output := []byte(`<p>Observed URL text: https://portal.example.test/login</p>`)
		if err := ValidateSelfContainedOutput(OutputKindSlidev, output); err != nil {
			t.Fatalf("plain escaped URL text should remain renderable, got %v", err)
		}
	})
	graphRenderSHA := strings.Repeat("c", 64)
	graphRenderRef := sourceProjectionRef{
		SourceOwnerID:                 "network_flow_activity",
		GraphViewID:                   "gv_bound",
		ProjectionResultID:            "gpres_" + graphRenderSHA,
		SourceSnapshotID:              "snap",
		ProjectionSchemaID:            graphprojection.ProjectionSchemaIDV2,
		ProjectionVersion:             "v2",
		NormalizedConfigurationSHA256: graphRenderSHA,
		NormalizedSourceSHA256:        graphRenderSHA,
		CanonicalOutputSHA256:         graphRenderSHA,
	}
	graphRenderResult := resolvedGraphResult{
		Ref: graphRenderRef,
		Result: graphprojection.CompletedResultV2{
			Binding: graphRenderRef.binding(),
			Vertices: []graphprojection.ResultVertexV2{
				{VertexID: "vertex-a", VertexKind: "actor", SortKey: "0001"},
				{VertexID: "vertex-b", VertexKind: "resource", SortKey: "0002"},
			},
			Edges: []graphprojection.ResultEdgeV2{
				{EdgeID: "edge-a-b", EdgeKind: "access", SrcVertexID: "vertex-a", DstVertexID: "vertex-b", SortKey: "0001"},
			},
		},
	}
	graphRenderResult.LabelCandidates = graphLabelCandidatesForResult(graphRenderResult.Result)
	reasonCases := []struct {
		name         string
		request      CreateReleaseRequest
		contract     TemplateContract
		model        ExportModel
		graphResults []resolvedGraphResult
		want         string
	}{
		{
			name: "invalid redaction profile",
			request: func() CreateReleaseRequest {
				req := baseRequest
				req.RedactionProfileVersion = "missing"
				return req
			}(),
			contract: contract,
			model:    baseModel,
			want:     "redaction_profile_invalid",
		},
		{
			name: "post redaction validation",
			request: func() CreateReleaseRequest {
				req := baseRequest
				req.RedactionProfileID = ExternalRedactionProfileID
				req.ReleaseScope = ReleaseScopeExternal
				return req
			}(),
			contract: contract,
			model: func() ExportModel {
				model := baseModel
				model = fixtureWithFields(t, model, ExportField{
					Path:         "/incident/unsupported_summary",
					ContentClass: ContentClassCuratedNarrative,
					Value:        "curated narrative without support",
				})
				return model
			}(),
			want: "invalid_scalar_value",
		},
		{
			name: "template render failure",
			request: func() CreateReleaseRequest {
				req := baseRequest
				req.OutputKind = "html"
				req.ReleaseScope = ReleaseScopeExternal
				req.RedactionProfileID = ExternalRedactionProfileID
				return req
			}(),
			contract: contract,
			model:    baseModel,
			want:     "invalid_scalar_value",
		},
		{
			name:    "undeclared template binding",
			request: baseRequest,
			contract: func() TemplateContract {
				next := contract
				next.Sections = append([]TemplateSection{}, contract.Sections...)
				next.Sections[0].Bindings = []string{"$.undeclared"}
				return next
			}(),
			model: baseModel,
			want:  "invalid_template_binding_pattern",
		},
		{
			name:    "missing required section",
			request: baseRequest,
			contract: func() TemplateContract {
				next := contract
				next.Sections = append([]TemplateSection{}, contract.Sections...)
				next.Sections[0].Bindings = []string{"$.sections['missing']"}
				return next
			}(),
			model: baseModel,
			want:  "required_section_empty",
		},
		{
			name: "self contained output validation",
			request: func() CreateReleaseRequest {
				req := baseRequest
				req.OutputKind = OutputKindSlidev
				return req
			}(),
			contract: contract,
			model: func() ExportModel {
				model := baseModel
				model = fixtureWithFields(t, model, ExportField{
					Path:         "/incident/description",
					ContentClass: ContentClassCuratedNarrative,
					Value:        "![remote](https://cdn.example.test/a.png)",
					SupportRefs:  []string{"/incident/status"},
				})
				return model
			}(),
			want: "remote_asset_reference_invalid",
		},
		{
			name: "recognized composition op cannot silently no-op",
			request: func() CreateReleaseRequest {
				req := baseRequest
				req.CompositionJSON = json.RawMessage(`{
					"schema_id":"cartulary.report_composition.v1",
					"deck_ops":[{"op_id":"op-1","op_kind":"exclude_section","on_unresolved":"fail","payload":{"section_anchor":{"anchor_kind":"section_anchor","template_section_decl_id":"missing","expansion_key":null}}}],
					"diagram_decls":[],
					"authored_texts":[]
				}`)
				return req
			}(),
			contract: contract,
			model:    baseModel,
			want:     "composition_anchor_unresolved",
		},
		{
			name: "graph diagram declaration must bind tuple ref",
			request: func() CreateReleaseRequest {
				req := baseRequest
				req.OutputKind = OutputKindMermaid
				req.OutputOptions = fixtureOutputOptions(OutputKindMermaid, req.ReleaseScope)
				req.GraphProjectionRefs = json.RawMessage(`[
					{"source_owner_id":"network_flow_activity","graph_view_id":"gv_bound","projection_result_id":"gpres_` + graphRenderSHA + `","source_snapshot_id":"snap","projection_schema_id":"graph_projection.v2","projection_version":"v2","normalized_configuration_sha256":"` + graphRenderSHA + `","normalized_source_sha256":"` + graphRenderSHA + `","canonical_output_sha256":"` + graphRenderSHA + `"}
				]`)
				req.CompositionJSON = json.RawMessage(`{
					"schema_id":"cartulary.report_composition.v1",
					"deck_ops":[],
					"diagram_decls":[{"decl_id":"diag-1","replaces_decl_id":null,"label_overrides":[],"layout":null,"diagram_kind":"flowchart","diagram_source_kind":"graph","source_graph_view_id":"gv_missing","selection_rule":{"schema_id":"diagram_selection_rule.v1","rule_id":"all_with_bounds","vertex_refs":[],"edge_refs":[],"seed_vertex_refs":[],"depth":null,"edge_kind_filter":[],"timeline_record_ids":[],"overflow_policy":"fail"},"layout_mode":"auto"}],
					"authored_texts":[]
				}`)
				return req
			}(),
			contract: contract,
			model:    baseModel,
			want:     "graph_projection_not_bound",
		},
		{
			name: "graph diagram declaration with tuple ref renders",
			request: func() CreateReleaseRequest {
				req := baseRequest
				req.OutputKind = OutputKindMermaid
				req.OutputOptions = fixtureOutputOptions(OutputKindMermaid, req.ReleaseScope)
				req.GraphProjectionRefs = json.RawMessage(`[
					{"source_owner_id":"network_flow_activity","graph_view_id":"gv_bound","projection_result_id":"gpres_` + graphRenderSHA + `","source_snapshot_id":"snap","projection_schema_id":"graph_projection.v2","projection_version":"v2","normalized_configuration_sha256":"` + graphRenderSHA + `","normalized_source_sha256":"` + graphRenderSHA + `","canonical_output_sha256":"` + graphRenderSHA + `"}
				]`)
				req.CompositionJSON = json.RawMessage(`{
					"schema_id":"cartulary.report_composition.v1",
					"deck_ops":[],
					"diagram_decls":[{"decl_id":"diag-1","replaces_decl_id":null,"label_overrides":[],"layout":null,"diagram_kind":"flowchart","diagram_source_kind":"graph","source_graph_view_id":"gv_bound","selection_rule":{"schema_id":"diagram_selection_rule.v1","rule_id":"all_with_bounds","vertex_refs":[],"edge_refs":[],"seed_vertex_refs":[],"depth":null,"edge_kind_filter":[],"timeline_record_ids":[],"overflow_policy":"fail"},"layout_mode":"auto"}],
					"authored_texts":[]
				}`)
				return req
			}(),
			contract:     contract,
			model:        baseModel,
			graphResults: []resolvedGraphResult{graphRenderResult},
			want:         "",
		},
		{
			name: "manual layout unsupported for mermaid",
			request: func() CreateReleaseRequest {
				req := baseRequest
				req.OutputKind = OutputKindMermaid
				req.OutputOptions = fixtureOutputOptions(OutputKindMermaid, req.ReleaseScope)
				req.CompositionJSON = json.RawMessage(`{
					"schema_id":"cartulary.report_composition.v1",
					"deck_ops":[],
					"diagram_decls":[{"decl_id":"diag-1","replaces_decl_id":null,"diagram_kind":"flowchart","diagram_source_kind":"timeline","source_graph_view_id":null,"selection_rule":{"schema_id":"diagram_selection_rule.v1","rule_id":"timeline_sequence","vertex_refs":[],"edge_refs":[],"seed_vertex_refs":[],"depth":null,"edge_kind_filter":[],"timeline_record_ids":["record-1"],"overflow_policy":"fail"},"label_overrides":[],"layout_mode":"manual","layout":{"schema_id":"composition_diagram_layout.v1","coordinate_space":{"unit":"css_px","origin":"top_left","width":400,"height":300},"node_positions":[],"edge_routes":[]}}],
					"authored_texts":[]
				}`)
				return req
			}(),
			contract: contract,
			model: fixtureWithFields(t, baseModel, ExportField{
				Path: "/timeline/record-1", SourceFamily: "timeline_event", SourceRecordID: "record-1", SourceRecordType: "timeline_event",
				ContentClass: ContentClassDerivedAnalytic, Value: map[string]any{"capture_state": "reviewed", "value": "observed"},
			}),
			want: "manual_layout_not_supported_for_output_kind",
		},
		{
			name: "mermaid labels reject raw angle brackets",
			request: func() CreateReleaseRequest {
				req := baseRequest
				req.OutputKind = OutputKindMermaid
				req.OutputOptions = fixtureOutputOptions(OutputKindMermaid, req.ReleaseScope)
				req.CompositionJSON = json.RawMessage(`{"schema_id":"cartulary.report_composition.v1","deck_ops":[],"authored_texts":[],"diagram_decls":[{"decl_id":"timeline","replaces_decl_id":null,"diagram_kind":"flowchart","diagram_source_kind":"timeline","source_graph_view_id":null,"selection_rule":{"schema_id":"diagram_selection_rule.v1","rule_id":"timeline_sequence","vertex_refs":[],"edge_refs":[],"seed_vertex_refs":[],"depth":null,"edge_kind_filter":[],"timeline_record_ids":["unsafe-event"],"overflow_policy":"fail"},"layout_mode":"auto","layout":null,"label_overrides":[{"target":{"target_kind":"vertex","ref":"unsafe-event"},"label":"<bad>"}]}]}`)
				return req
			}(),
			contract: contract,
			model: func() ExportModel {
				model := baseModel
				model = fixtureWithFields(t, model, ExportField{
					Path: "/timeline/unsafe-event", SourceFamily: "timeline_event", SourceRecordID: "unsafe-event", SourceRecordType: "timeline_event",
					ContentClass: ContentClassDerivedAnalytic,
					Value:        map[string]any{"capture_state": "rough", "value": "observed"},
				})
				return model
			}(),
			want: "invalid_mermaid_construct",
		},
	}
	for _, tc := range reasonCases {
		t.Run(tc.name, func(t *testing.T) {
			source, err := canonicalJSON(tc.model)
			if err != nil {
				t.Fatalf("source json: %v", err)
			}
			if len(tc.request.CompositionJSON) > 0 {
				tc.request.CompositionJSON = sealedReportingCompositionFixture(t, tc.request.CompositionJSON, tc.model.IncidentID, tc.contract)
			}
			_, reasonCode, err := fixtureRenderer().Render(context.Background(), renderCandidateFixture(tc.request, tc.contract, tc.model, reportingObjectDigest(ExportModelSchemaID, source), tc.graphResults...))
			if tc.want == "" {
				if err != nil || reasonCode != "" {
					t.Fatalf("render reason = %q, err=%v want success", reasonCode, err)
				}
				return
			}
			if err == nil || reasonCode != tc.want {
				t.Fatalf("render reason = %q, err=%v want %q", reasonCode, err, tc.want)
			}
		})
	}

	t.Run("manifest encoding failure", func(t *testing.T) {
		original := encodeCanonicalJSON
		manifestEncodes := 0
		encodeCanonicalJSON = func(value any) ([]byte, error) {
			if _, ok := value.(RedactionManifest); ok {
				manifestEncodes++
				if manifestEncodes == 2 {
					return nil, errors.New("forced manifest encoding failure")
				}
			}
			return original(value)
		}
		defer func() { encodeCanonicalJSON = original }()

		source, err := canonicalJSON(baseModel)
		if err != nil {
			t.Fatalf("source json: %v", err)
		}
		_, reasonCode, err := fixtureRenderer().Render(context.Background(), renderCandidateFixture(baseRequest, contract, baseModel, reportingObjectDigest(ExportModelSchemaID, source)))
		if err == nil || reasonCode != "invalid_type" {
			t.Fatalf("manifest encoding reason = %q, err=%v", reasonCode, err)
		}
	})
}

func TestRenderBundleManifestBindsOutputHash(t *testing.T) {
	assertPassingRenderSummary(t)
	contract, ok := ResolveTemplateContract(DefaultTemplateID, DefaultTemplateVersion)
	if !ok {
		t.Fatal("resolve template contract")
	}
	model := RedactedExportModel{ExportModel: ExportModel{RenderIdentity: fixtureRenderIdentity(), SchemaID: ExportModelSchemaID, SnapshotContent: SnapshotContent{DerivationVersion: DerivationVersion}},
		Fields: []RedactedField{
			{Path: "/incident/status", ContentClass: ContentClassDerivedAnalytic, Value: "active"},
			{
				Path:         "/incident/title",
				ContentClass: ContentClassDerivedAnalytic,
				Value:        "Example",
			},
		},
	}
	bindFixtureRetainedSections(&model)
	redactionManifestSHA256 := strings.Repeat("a", 64)
	bindFixtureRetainedSections(&model)
	bundle, err := fixtureRenderer().fixtureRenderReportBundle(context.Background(), contract, OutputKindSlidev, model, RedactionProfileView{}, redactionManifestSHA256, ReleaseScopeInternalDraft, fixtureOutputOptions(OutputKindSlidev, ReleaseScopeInternalDraft), nil, RedactionBundleArtifacts{})
	if err != nil {
		t.Fatalf("build render bundle: %v", err)
	}
	if bundle.Manifest.SchemaID != RenderBundleManifestSchemaID || bundle.Manifest.PrimaryPath != "slides.md" {
		t.Fatalf("unexpected bundle manifest: %#v", bundle.Manifest)
	}
	var primaryBytes []byte
	for _, file := range bundle.Files {
		if file.Path == bundle.PrimaryPath {
			primaryBytes = file.Bytes
			break
		}
	}
	if len(primaryBytes) == 0 {
		t.Fatalf("primary bundle file missing: %#v", bundle.Files)
	}
	if bundle.ManifestSHA256 == "" || bundle.ManifestSHA256 == hashHex(primaryBytes) {
		t.Fatalf("output hash must bind the manifest, not the primary bytes: manifest=%q output=%q", bundle.ManifestSHA256, hashHex(primaryBytes))
	}
	if len(bundle.Files) < 5 || !strings.Contains(string(primaryBytes), "Incident Report") {
		t.Fatalf("primary bundle file does not preserve rendered bytes: %#v", bundle.Files)
	}
	if bundle.Manifest.ToolchainSnapshotSHA256 == nil || bundle.Manifest.ValidationSummarySHA256 == "" {
		t.Fatalf("bundle manifest must bind toolchain and validation artifacts: %#v", bundle.Manifest)
	}
}

func TestRenderBundleExternalReleaseRecordsDeterminismDigest(t *testing.T) {
	contract, ok := ResolveTemplateContract(DefaultTemplateID, DefaultTemplateVersion)
	if !ok {
		t.Fatal("resolve template contract")
	}
	model := RedactedExportModel{ExportModel: ExportModel{RenderIdentity: fixtureRenderIdentity(), SchemaID: ExportModelSchemaID, SnapshotContent: SnapshotContent{DerivationVersion: DerivationVersion}},
		Fields: []RedactedField{
			{Path: "/incident/status", ContentClass: ContentClassDerivedAnalytic, Value: "active"},
			{
				Path:         "/incident/title",
				ContentClass: ContentClassCuratedNarrative,
				Value:        "External deterministic report",
			},
		},
	}
	bindFixtureRetainedSections(&model)
	first, err := fixtureRenderer().fixtureRenderReportBundle(context.Background(), contract, OutputKindSlidev, model, RedactionProfileView{}, strings.Repeat("a", 64), ReleaseScopeExternal, fixtureOutputOptions(OutputKindSlidev, ReleaseScopeExternal), nil, RedactionBundleArtifacts{})
	if err != nil {
		t.Fatalf("build external render bundle: %v", err)
	}
	renderer := &Renderer{runtime: alteredRenderRuntime{change: func(result *renderprocess.Result) {
		result.NetworkObservations[0].Count = 200
		var protocol renderStageResult
		if err := json.Unmarshal(result.Protocol, &protocol); err != nil {
			t.Fatal(err)
		}
		protocol.BrowserObservations[0].AttemptCount = 100
		var err error
		result.Protocol, err = json.Marshal(protocol)
		if err != nil {
			t.Fatal(err)
		}
	}}}
	second, err := renderer.fixtureRenderReportBundle(context.Background(), contract, OutputKindSlidev, model, RedactionProfileView{}, strings.Repeat("a", 64), ReleaseScopeExternal, fixtureOutputOptions(OutputKindSlidev, ReleaseScopeExternal), nil, RedactionBundleArtifacts{})
	if err != nil {
		t.Fatalf("build external render bundle again: %v", err)
	}
	if first.Manifest.ExternalDeterminismSHA256 == nil || *first.Manifest.ExternalDeterminismSHA256 == "" {
		t.Fatalf("external release must record determinism digest: %#v", first.Manifest)
	}
	if first.ManifestSHA256 != second.ManifestSHA256 ||
		*first.Manifest.ExternalDeterminismSHA256 != *second.Manifest.ExternalDeterminismSHA256 {
		t.Fatalf("external release render must be deterministic: first=%#v second=%#v", first.Manifest, second.Manifest)
	}
}

func TestRenderBundleSeparatesSensitiveRedactionArtifacts(t *testing.T) {
	contract, ok := ResolveTemplateContract(DefaultTemplateID, DefaultTemplateVersion)
	if !ok {
		t.Fatal("resolve template contract")
	}
	model := RedactedExportModel{ExportModel: ExportModel{RenderIdentity: fixtureRenderIdentity(), SchemaID: ExportModelSchemaID, SnapshotContent: SnapshotContent{DerivationVersion: DerivationVersion}},
		Fields: []RedactedField{
			{Path: "/incident/status", ContentClass: ContentClassDerivedAnalytic, Value: "active"},
			{
				Path:         "/incident/title",
				ContentClass: ContentClassDerivedAnalytic,
				Value:        "Example",
			},
		},
	}
	bindFixtureRetainedSections(&model)
	bundle, err := fixtureRenderer().fixtureRenderReportBundle(context.Background(), contract, OutputKindSlidev, model, RedactionProfileView{}, strings.Repeat("a", 64), ReleaseScopeInternalDraft, fixtureOutputOptions(OutputKindSlidev, ReleaseScopeInternalDraft), nil, RedactionBundleArtifacts{
		RedactionManifestJSON: []byte(`{"schema_id":"` + RedactionManifestSchemaID + `"}`),
		ProfileViewJSON:       []byte(`{"schema_id":"` + RedactionProfileViewSchemaID + `"}`),
		TokenManifestJSON:     []byte(`{"schema_id":"` + RedactionTokenManifestSchemaID + `"}`),
		TokenManifestSHA256:   strings.Repeat("b", 64),
	})
	if err != nil {
		t.Fatalf("build render bundle: %v", err)
	}
	roles := map[string]RenderBundleFile{}
	for _, file := range bundle.Files {
		roles[file.Role] = file
	}
	if roles[renderBundleRoleRedactionManifest].Path != "validation/redaction-manifest.json" {
		t.Fatalf("redaction manifest artifact missing from bundle: %#v", bundle.Files)
	}
	if roles[renderBundleRoleRedactionProfileView].Path != "validation/redaction-profile-view.json" {
		t.Fatalf("profile view artifact missing from bundle: %#v", bundle.Files)
	}
	if roles[renderBundleRoleTokenManifest].Path != "validation/token-manifest.json" {
		t.Fatalf("token manifest artifact missing from bundle: %#v", bundle.Files)
	}
	for _, file := range bundle.Files {
		if !shareableBundleMember(file.RenderBundleManifestFile) || bytes.Contains(file.Bytes, []byte(RedactionRevealMapSchemaID)) {
			t.Fatal("sensitive artifact entered ordinary render output")
		}
	}
	for _, file := range bundle.Manifest.Files {
		if !shareableBundleMember(file) {
			t.Fatal("sensitive artifact entered ordinary manifest")
		}
	}
	assertSensitiveArtifactAdmission(t)
	if bundle.Manifest.TokenManifestSHA256 == nil || *bundle.Manifest.TokenManifestSHA256 != strings.Repeat("b", 64) {
		t.Fatalf("bundle manifest must bind token manifest digest, got %#v", bundle.Manifest)
	}
	if len(bundle.Manifest.Files) < 7 || bundle.ManifestSHA256 == "" {
		t.Fatalf("bundle manifest must bind primary output and redaction artifacts, got %#v", bundle.Manifest)
	}
}

func TestRenderBundleMermaidPipelineEmitsSourceSVGAndDeterministicManifest(t *testing.T) {
	t.Run("selected timeline", testTimelineDiagramSelection)
	t.Run("retained timeline", testRetainedTimeline)
	t.Run("no invented diagram", testNoInventedDiagram)
	contract, ok := ResolveTemplateContract(DefaultTemplateID, DefaultTemplateVersion)
	if !ok {
		t.Fatal("resolve template contract")
	}
	model := RedactedExportModel{ExportModel: ExportModel{RenderIdentity: fixtureRenderIdentity(), SchemaID: ExportModelSchemaID, SnapshotContent: SnapshotContent{DerivationVersion: DerivationVersion}},
		Fields: []RedactedField{
			{Path: "/incident/title", ContentClass: ContentClassCuratedNarrative, Value: "Example"},
			{Path: "/incident/status", ContentClass: ContentClassDerivedAnalytic, Value: "open"},
		},
	}
	bindFixtureRetainedSections(&model)
	model.IncidentID = "incident"
	model.TimelineEvents = fixtureRetainedTimelineEvents()
	composition := fixtureTimelineComposition(t, contract, model.IncidentID, "flowchart", []string{"event-late", "event-early"})
	first, err := fixtureRenderer().fixtureRenderReportBundle(context.Background(), contract, OutputKindMermaid, model, RedactionProfileView{}, strings.Repeat("a", 64), ReleaseScopeInternalReview, fixtureOutputOptions(OutputKindMermaid, ReleaseScopeInternalReview), composition, RedactionBundleArtifacts{})
	if err != nil {
		t.Fatalf("render mermaid bundle: %v", err)
	}
	second, err := fixtureRenderer().fixtureRenderReportBundle(context.Background(), contract, OutputKindMermaid, model, RedactionProfileView{}, strings.Repeat("a", 64), ReleaseScopeInternalReview, fixtureOutputOptions(OutputKindMermaid, ReleaseScopeInternalReview), composition, RedactionBundleArtifacts{})
	if err != nil {
		t.Fatalf("render mermaid bundle again: %v", err)
	}
	if first.ManifestSHA256 != second.ManifestSHA256 {
		t.Fatalf("mermaid render bundle must be deterministic: first=%q second=%q", first.ManifestSHA256, second.ManifestSHA256)
	}
	roles := map[string]RenderBundleFile{}
	for _, file := range first.Files {
		roles[file.Role] = file
	}
	if roles[renderBundleRoleSourceMermaid].Path != "diagrams/timeline.mmd" ||
		!strings.Contains(string(roles[renderBundleRoleSourceMermaid].Bytes), "flowchart TD") {
		t.Fatalf("mermaid source missing or malformed: %#v", roles[renderBundleRoleSourceMermaid])
	}
	if roles[renderBundleRoleRenderedSVG].Path != "diagrams/timeline.svg" ||
		!strings.Contains(string(roles[renderBundleRoleRenderedSVG].Bytes), "<svg") {
		t.Fatalf("mermaid SVG missing or malformed: %#v", roles[renderBundleRoleRenderedSVG])
	}
	if roles[renderBundleRoleValidationSummary].Path != "validation/summary.json" ||
		roles[renderBundleRoleToolchainSnapshot].Path != "validation/toolchain.json" {
		t.Fatalf("validation/toolchain artifacts missing: %#v", first.Files)
	}
}

func TestRenderBundlePreservesInertSourceLabelText(t *testing.T) {
	contract, ok := ResolveTemplateContract(DefaultTemplateID, DefaultTemplateVersion)
	if !ok {
		t.Fatal("resolve template contract")
	}
	model := RedactedExportModel{ExportModel: ExportModel{RenderIdentity: fixtureRenderIdentity(), SchemaID: ExportModelSchemaID, SnapshotContent: SnapshotContent{DerivationVersion: DerivationVersion}},
		Fields: []RedactedField{
			{Path: "/incident/title", ContentClass: ContentClassCuratedNarrative, Value: "Inert url(evil)"},
			{Path: "/incident/status", ContentClass: ContentClassDerivedAnalytic, Value: "open"},
			{Path: "/incident/url(evil)", ContentClass: ContentClassDerivedAnalytic, Value: "inert label"},
		},
	}
	bindFixtureRetainedSections(&model)
	model.IncidentID = "incident"
	model.TimelineEvents = fixtureRetainedTimelineEvents()
	composition := fixtureTimelineComposition(t, contract, model.IncidentID, "flowchart", []string{"event-late", "event-early"})
	label := "Inert url(evil)"
	model.TimelineEvents[0].DisplayTimes.PrimaryDisplay = &label
	bundle, err := fixtureRenderer().fixtureRenderReportBundle(context.Background(), contract, OutputKindMermaid, model, RedactionProfileView{}, strings.Repeat("a", 64), ReleaseScopeInternalReview, fixtureOutputOptions(OutputKindMermaid, ReleaseScopeInternalReview), composition, RedactionBundleArtifacts{})
	if err != nil {
		t.Fatalf("inert source label text must remain displayable: %v", err)
	}
	found := false
	for _, file := range bundle.Files {
		if file.Role == renderBundleRoleSourceMermaid {
			found = strings.Contains(string(file.Bytes), "Inert url(evil)")
			if strings.Contains(string(file.Bytes), "/incident/") {
				t.Fatal("private selector rendered as a diagram label")
			}
		}
	}
	if !found {
		t.Fatal("redacted inert value was lost from the rendered diagram")
	}
	deck, err := fixtureDerivedDeck(contract, model, renderOutputOptions{}, nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, slide := range deck.Slides {
		for _, block := range slide.Blocks {
			if strings.Contains(fmt.Sprint(block.Fields[0].Value), "/incident/") {
				t.Fatal("private selector rendered as slide text")
			}
		}
	}
}

func mustTestUUID(value string) uuid.UUID {
	parsed, err := uuid.Parse(value)
	if err != nil {
		panic(err)
	}
	return parsed
}

// Explicitly materialize the selected profile for renderer fixtures. An invalid
// selector deliberately produces an invalid frozen input for the denial case.
func renderCandidateFixture(request CreateReleaseRequest, contract TemplateContract, model ExportModel, digest string, graphs ...resolvedGraphResult) renderInput {
	profile, profileSHA, _ := ResolveRedactionProfile(request.RedactionProfileID, request.RedactionProfileVersion, request.RecipientPartitionRefs, model.SnapshotContent)
	return renderInput{SnapshotID: model.SnapshotID, IncidentID: model.IncidentID, TemplateID: contract.TemplateID, TemplateVersion: contract.TemplateVersion, SourceFacts: fixtureSourceFactsForContent(model.SnapshotContent), SourceFactsSHA256: fixtureSourceFactsDigest(model.SnapshotContent), ReleaseScope: request.ReleaseScope, RecipientPartitionRefs: request.RecipientPartitionRefs, OutputKind: request.OutputKind, OutputOptions: request.OutputOptions, GraphProjectionRefs: request.GraphProjectionRefs, CompositionJSON: request.CompositionJSON, TemplateContract: contract, ExportModel: model, ExportModelSHA256: digest, RedactionProfileID: request.RedactionProfileID, RedactionProfileVersion: request.RedactionProfileVersion, RedactionProfile: profile, RedactionProfileSHA256: profileSHA, GraphSources: frozenGraphSources{SchemaID: frozenGraphSourcesSchema, Results: append([]resolvedGraphResult{}, graphs...)}}
}
