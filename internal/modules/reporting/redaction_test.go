package reporting

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestRedactionProfilePrecedenceActionsAndManifest_Unit(t *testing.T) {
	descriptionPath := "/incident/description"
	sourceClass := ContentClassSourceEvidence
	maxChars := 4
	maskText := "[MASKED]"
	profile := RedactionProfile{
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
	profileSHA, err := ValidateRedactionProfile(profile)
	if err != nil {
		t.Fatalf("validate profile: %v", err)
	}
	model := ExportModel{
		SchemaID:                     ExportModelSchemaID,
		IncidentID:                   "incident-1",
		SnapshotAt:                   time.Date(2026, 5, 23, 12, 0, 0, 0, time.UTC),
		SourceChangeSetHighWatermark: SourceBoundaryTokenPrefix + strings.Repeat("0", 64),
		DerivationVersion:            DerivationVersion,
		Fields: []ExportField{
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
		},
	}
	sourceBytes, err := canonicalJSON(model)
	if err != nil {
		t.Fatalf("source model json: %v", err)
	}
	result, err := RedactExportModel(model, profile, profileSHA, hashHex(sourceBytes), ReleaseScopeInternalReview, nil)
	if err != nil {
		t.Fatalf("redact model: %v", err)
	}
	values := map[string]any{}
	for _, field := range result.Model.Fields {
		values[field.Path] = field.Value
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
		entries[entry.Path] = entry
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
	if entries["/incident/description"].SelectedRuleTrace.SelectionKind != "path" ||
		entries["/incident/raw_note"].SelectedRuleTrace.SelectionKind != "content_class" ||
		entries["/incident/internal_note"].SelectedRuleTrace.SelectionKind != "profile_default" {
		t.Fatalf("manifest must record selected-rule trace objects, got %#v", result.Manifest.Entries)
	}
	assertRedactionUsesStructuredExportModelBlocks(t)
	assertCanonicalExportModelGoldenHash(t)
	assertRedactionTokensRevealMapAndProfileViewAreCanonical(t)
}

func assertRedactionTokensRevealMapAndProfileViewAreCanonical(t *testing.T) {
	t.Helper()
	profile := RedactionProfile{
		SchemaID:      RedactionProfileSchemaID,
		ProfileID:     "test.tokenized",
		Version:       "1",
		DefaultAction: RedactionActionSpec{Type: ActionMask},
	}
	profileSHA, err := ValidateRedactionProfile(profile)
	if err != nil {
		t.Fatalf("validate profile: %v", err)
	}
	model := ExportModel{
		SchemaID:                     ExportModelSchemaID,
		IncidentID:                   "incident-1",
		SnapshotAt:                   time.Date(2026, 5, 23, 12, 0, 0, 0, time.UTC),
		SourceChangeSetHighWatermark: SourceBoundaryTokenPrefix + strings.Repeat("9", 64),
		DerivationVersion:            DerivationVersion,
		Fields: []ExportField{
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
		},
	}
	sourceBytes, err := canonicalJSON(model)
	if err != nil {
		t.Fatalf("source model json: %v", err)
	}
	result, err := RedactExportModel(model, profile, profileSHA, hashHex(sourceBytes), ReleaseScopeInternalReview, nil)
	if err != nil {
		t.Fatalf("redact model: %v", err)
	}
	if len(result.Model.Fields) != 1 {
		t.Fatalf("expected one tokenized field, got %#v", result.Model.Fields)
	}
	displayToken, ok := result.Model.Fields[0].Value.(string)
	if !ok || !strings.HasPrefix(displayToken, "SUBJECT-") {
		t.Fatalf("mask/stub subject field must receive a deterministic display token, got %#v", result.Model.Fields[0].Value)
	}
	if result.ProfileView.SchemaID != RedactionProfileViewSchemaID || result.ProfileViewSHA256 == "" || len(result.ProfileViewJSON) == 0 {
		t.Fatalf("redaction result must carry a canonical profile view, got %#v", result)
	}
	if result.Manifest.ProfileViewSHA256 != result.ProfileViewSHA256 {
		t.Fatalf("redaction manifest must bind the profile view digest")
	}
	if result.TokenManifest == nil || result.TokenManifestSHA256 == "" || len(result.TokenManifestJSON) == 0 {
		t.Fatalf("tokenized redaction must carry a token manifest, got %#v", result)
	}
	if result.RevealMap == nil || result.RevealMapSHA256 == "" || len(result.RevealMapJSON) == 0 {
		t.Fatalf("tokenized redaction must carry an internal reveal map, got %#v", result)
	}
	if result.Manifest.TokenManifestSHA256 == nil || *result.Manifest.TokenManifestSHA256 != result.TokenManifestSHA256 {
		t.Fatalf("redaction manifest must bind the token manifest digest")
	}
	tokenEntry := result.TokenManifest.Entries[0]
	if tokenEntry.DisplayToken != displayToken || tokenEntry.StableSubjectRef != "party:party_a" {
		t.Fatalf("token manifest must bind display token to stable subject, got %#v", tokenEntry)
	}
	revealEntry := result.RevealMap.Entries[0]
	if revealEntry.DisplayToken != displayToken || revealEntry.OriginalValueSHA256 == "" || revealEntry.OriginalValue == nil {
		t.Fatalf("reveal map must retain sensitive original material internally, got %#v", revealEntry)
	}
	if result.Manifest.Entries[0].Outcome != "tokenized" || result.Manifest.Entries[0].SelectedRuleTrace.SelectionKind != "profile_default" {
		t.Fatalf("manifest must record tokenized outcome and rule trace, got %#v", result.Manifest.Entries[0])
	}
	repeated, err := RedactExportModel(model, profile, profileSHA, hashHex(sourceBytes), ReleaseScopeInternalReview, nil)
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
		SchemaID:      RedactionProfileSchemaID,
		ProfileID:     "test.structured",
		Version:       "1",
		DefaultAction: RedactionActionSpec{Type: ActionAllow},
	}
	profileSHA, err := ValidateRedactionProfile(profile)
	if err != nil {
		t.Fatalf("validate profile: %v", err)
	}
	model, sourceSHA, err := buildStructuredExportModel(
		"incident-1",
		"snapshot-1",
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
	model.Fields = []ExportField{{
		Path:         "/incident/title",
		ContentClass: ContentClassCuratedNarrative,
		SourceFamily: "incident_metadata",
		Value:        "legacy cache title",
	}}
	result, err := RedactExportModel(model, profile, profileSHA, sourceSHA, ReleaseScopeInternalReview, nil)
	if err != nil {
		t.Fatalf("redact model: %v", err)
	}
	if len(result.Model.Fields) != 1 || result.Model.Fields[0].Value != "structured title" {
		t.Fatalf("redaction must use structured export model blocks over flat compatibility cache, got %#v", result.Model.Fields)
	}
}

func TestRedactionProfileRejectsConflictsHashAndUnsafeBounds_Unit(t *testing.T) {
	path := "/incident/title"
	class := ContentClassCuratedNarrative
	maxZero := 0
	cases := []struct {
		name    string
		profile RedactionProfile
	}{
		{
			name: "duplicate path rules",
			profile: RedactionProfile{
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
				SchemaID:      RedactionProfileSchemaID,
				ProfileID:     "test.hash",
				Version:       "1",
				DefaultAction: RedactionActionSpec{Type: ActionHash},
			},
		},
		{
			name: "truncate missing safe bound",
			profile: RedactionProfile{
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
	profileSHA, err := ValidateRedactionProfile(profile)
	if err != nil {
		t.Fatalf("validate profile: %v", err)
	}
	model := ExportModel{
		SchemaID:                     ExportModelSchemaID,
		IncidentID:                   "incident-1",
		SnapshotAt:                   time.Date(2026, 5, 23, 12, 0, 0, 0, time.UTC),
		SourceChangeSetHighWatermark: SourceBoundaryTokenPrefix + strings.Repeat("1", 64),
		DerivationVersion:            DerivationVersion,
		Fields: []ExportField{
			{
				Path:          "/evidence/blob",
				ContentClass:  ContentClassSourceEvidence,
				Value:         "raw bytes",
				RawBlobSource: true,
				OpaqueBinary:  true,
			},
			{
				Path:         "/incident/internal_note",
				ContentClass: ContentClassWorkingMaterial,
				Value:        "internal",
			},
		},
	}
	sourceBytes, err := canonicalJSON(model)
	if err != nil {
		t.Fatalf("source model json: %v", err)
	}
	if _, err := RedactExportModel(model, profile, profileSHA, hashHex(sourceBytes), ReleaseScopeExternal, nil); !errors.Is(err, ErrRedactionValidation) {
		t.Fatalf("external release must reject raw bytes and working material, got %v", err)
	}
}

func TestDisclosurePartitionsAndCuratedSupportRefsFailClosed_Unit(t *testing.T) {
	allowedPath := "/incident/summary"
	restrictedPath := "/incident/restricted"
	profile := RedactionProfile{
		SchemaID:  RedactionProfileSchemaID,
		ProfileID: "test.partitions",
		Version:   "1",
		AllowedDisclosurePartitionRefs: []string{
			"public_summary",
		},
		DefaultAction: RedactionActionSpec{Type: ActionAllow},
		Rules: []RedactionRule{
			{RuleID: "allow-summary", Path: &allowedPath, Action: RedactionActionSpec{Type: ActionAllow}},
			{RuleID: "allow-restricted", Path: &restrictedPath, Action: RedactionActionSpec{Type: ActionAllow}},
		},
	}
	profileSHA, err := ValidateRedactionProfile(profile)
	if err != nil {
		t.Fatalf("validate profile: %v", err)
	}
	model := ExportModel{
		SchemaID:                     ExportModelSchemaID,
		IncidentID:                   "incident-1",
		SnapshotAt:                   time.Date(2026, 5, 23, 12, 0, 0, 0, time.UTC),
		SourceChangeSetHighWatermark: SourceBoundaryTokenPrefix + strings.Repeat("2", 64),
		DerivationVersion:            DerivationVersion,
		Fields: []ExportField{
			{
				Path:                    allowedPath,
				ContentClass:            ContentClassCuratedNarrative,
				Value:                   "supported public summary",
				DisclosurePartitionRefs: []string{"public_summary"},
				SupportRefs:             []string{"/incident/source"},
			},
			{
				Path:                    restrictedPath,
				ContentClass:            ContentClassDerivedAnalytic,
				Value:                   "restricted analytic",
				DisclosurePartitionRefs: []string{"legal_hold"},
			},
		},
	}
	sourceBytes, err := canonicalJSON(model)
	if err != nil {
		t.Fatalf("source model json: %v", err)
	}
	result, err := RedactExportModel(model, profile, profileSHA, hashHex(sourceBytes), ReleaseScopeExternal, nil)
	if err != nil {
		t.Fatalf("redact with partition filter: %v", err)
	}
	if len(result.Model.Fields) != 1 || result.Model.Fields[0].Path != allowedPath {
		t.Fatalf("external profile must drop fields outside allowed partitions, got %#v", result.Model.Fields)
	}
	entries := map[string]RedactionManifestEntry{}
	for _, entry := range result.Manifest.Entries {
		entries[entry.Path] = entry
	}
	if entries[restrictedPath].Outcome != "dropped_disclosure_partition" {
		t.Fatalf("restricted field must be represented as partition drop, got %#v", entries[restrictedPath])
	}

	model.Fields[0].SupportRefs = nil
	if _, err := RedactExportModel(model, profile, profileSHA, hashHex(sourceBytes), ReleaseScopeExternal, nil); !errors.Is(err, ErrRedactionValidation) {
		t.Fatalf("external curated narrative without support refs must fail closed, got %v", err)
	}
}

func TestBuildExportModelUsesReportingOwnedMetadataSnapshotStableHash(t *testing.T) {
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
	snapshotAt := time.Date(2026, 5, 23, 12, 0, 0, 123, time.FixedZone("test", -5*60*60))
	watermark := SourceBoundaryTokenPrefix + strings.Repeat("4", 64)
	workbookFields := []ExportField{
		{
			Path:         "/workbook/rows/indicator-1/value",
			ContentClass: ContentClassSourceEvidence,
			SourceFamily: "workbook",
			Value:        "example.test",
		},
	}

	first, firstSHA, err := BuildExportModel(incident, snapshotAt, watermark, append([]ExportField(nil), workbookFields...))
	if err != nil {
		t.Fatalf("build first export model: %v", err)
	}
	second, secondSHA, err := BuildExportModel(incident, snapshotAt, watermark, append([]ExportField(nil), workbookFields...))
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
	if firstSHA != hashHex(firstJSON) {
		t.Fatalf("export model sha must be derived from canonical reporting model json")
	}
	if firstSHA != secondSHA || string(firstJSON) != string(secondJSON) {
		t.Fatalf("identical reporting metadata snapshots must produce stable export models: first=%s second=%s", firstSHA, secondSHA)
	}
	if first.IncidentID != incident.ID || !first.SnapshotAt.Equal(snapshotAt.UTC()) {
		t.Fatalf("export model must preserve reporting-owned incident id and normalized snapshot time: %#v", first)
	}
	paths := map[string]any{}
	for _, field := range first.RedactionFields() {
		paths[field.Path] = field.Value
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

func assertCanonicalExportModelGoldenHash(t *testing.T) {
	t.Helper()
	description := "Golden summary"
	incident := IncidentMetadataSnapshot{
		ID:          "00000000-0000-0000-0000-000000000099",
		Title:       "Golden incident",
		Description: &description,
		Status:      "active",
		Version:     7,
	}
	model, modelSHA, err := BuildExportModel(
		incident,
		time.Date(2026, 5, 23, 12, 0, 0, 0, time.UTC),
		SourceBoundaryTokenPrefix+strings.Repeat("7", 64),
		[]ExportField{{
			Path:         "/timeline/tl-1",
			ContentClass: ContentClassSourceEvidence,
			SourceFamily: "timeline_event",
			Value:        "First observed event",
			SupportRefs:  []string{"/record_envelopes/evidence-1"},
		}},
	)
	if err != nil {
		t.Fatalf("build export model: %v", err)
	}
	canonical, err := canonicalJSON(model)
	if err != nil {
		t.Fatalf("canonical export model: %v", err)
	}
	if modelSHA != hashHex(canonical) {
		t.Fatalf("model sha = %s, canonical hash = %s", modelSHA, hashHex(canonical))
	}
	const wantSHA = "90b89be2b42fc61a8a78690f9e4cfcba0f686568e3fdafd98c9e890aa94cf06a"
	if modelSHA != wantSHA {
		t.Fatalf("canonical export model golden hash = %s, want %s\n%s", modelSHA, wantSHA, canonical)
	}
}

func TestDecoderNormalizationAndRegisteredReasons_Unit(t *testing.T) {
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
		"redaction_profile_version":"1",
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
		"redaction_profile_version":"1",
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
		"redaction_profile_version":"1",
		"output_kind":"slidev",
		"output_options":{"source_only":true},
		"graph_projection_refs":[{
			"projection_schema_id":"graph_projection.v1",
			"graph_view_id":"gv_a",
			"source_snapshot_id":"00000000-0000-0000-0000-000000000001",
			"projection_run_id":"gr_a",
			"projection_version":"1",
			"projection_config_digest":"` + graphSHA + `",
			"projection_source_digest":"` + graphSHA + `",
			"projection_output_digest":"` + graphSHA + `"
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
		"redaction_profile_version":"1",
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
		"redaction_profile_version":"1",
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
		"redaction_profile_version":"1",
		"output_kind":"slidev",
		"graph_projection_refs":[
			{"projection_schema_id":"graph_projection.v1","graph_view_id":"gv_a","source_snapshot_id":"s","projection_run_id":"r1","projection_version":"1","projection_config_digest":"` + graphSHA + `","projection_source_digest":"` + graphSHA + `","projection_output_digest":"` + graphSHA + `"},
			{"projection_schema_id":"graph_projection.v1","graph_view_id":"gv_a","source_snapshot_id":"s","projection_run_id":"r2","projection_version":"1","projection_config_digest":"` + graphSHA + `","projection_source_digest":"` + graphSHA + `","projection_output_digest":"` + graphSHA + `"}
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
		"redaction_profile_version":"1",
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
		"redaction_profile_version":"1",
		"output_kind":"slidev",
		"recipient_partition_refs":["party:a"]
	}`))
	if apiErr != nil {
		t.Fatalf("structural release decode should not perform hidden-resource selector validation, got %#v", apiErr)
	}
	_, internalPartitions := validateCreateReleaseRequestSemantics(internalPartitionRequest)
	if internalPartitions == nil || internalPartitions.Details["reason_code"] != "recipient_partitions_not_allowed" {
		t.Fatalf("internal recipient partitions must use closed rejection reason, got %#v", internalPartitions)
	}

	recipientModel := ExportModel{
		Fields: []ExportField{
			{Path: "/parties/party_a", SourceFamily: "party", DisclosurePartitionRefs: []string{"party:party_a"}},
		},
	}
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
				RedactionProfileVersion: "1",
			},
			wantReason: "recipient_partition_profile_mismatch",
		},
		{
			name: "external recipient ref grammar",
			request: CreateReleaseRequest{
				ReleaseScope:            ReleaseScopeExternal,
				RedactionProfileID:      ExternalRedactionProfileID,
				RedactionProfileVersion: "1",
				RecipientPartitionRefs:  []string{"party:bad/ref"},
			},
			wantReason: "invalid_recipient_partition_ref",
		},
		{
			name: "external recipient must resolve in snapshot",
			request: CreateReleaseRequest{
				ReleaseScope:            ReleaseScopeExternal,
				RedactionProfileID:      ExternalRedactionProfileID,
				RedactionProfileVersion: "1",
				RecipientPartitionRefs:  []string{"party:missing"},
			},
			wantReason: "unknown_recipient_partition",
		},
		{
			name: "profile allowed party subset must match recipient set",
			request: CreateReleaseRequest{
				ReleaseScope:            ReleaseScopeExternal,
				RedactionProfileID:      InternalRedactionProfileID,
				RedactionProfileVersion: "1",
				RecipientPartitionRefs:  []string{"party:party_a"},
			},
			wantReason: "recipient_partition_profile_mismatch",
		},
	}
	for _, tc := range recipientValidationCases {
		t.Run(tc.name, func(t *testing.T) {
			apiErr := validateCreateReleaseRecipientPartitions(tc.request, recipientModel)
			if apiErr == nil || apiErr.Details["reason_code"] != tc.wantReason {
				t.Fatalf("recipient validation reason = %#v want %q", apiErr, tc.wantReason)
			}
		})
	}
	validRecipientRequest := CreateReleaseRequest{
		ReleaseScope:            ReleaseScopeExternal,
		RedactionProfileID:      ExternalRedactionProfileID,
		RedactionProfileVersion: "1",
		RecipientPartitionRefs:  []string{"party:party_a"},
	}
	if apiErr := validateCreateReleaseRecipientPartitions(validRecipientRequest, recipientModel); apiErr != nil {
		t.Fatalf("valid recipient partition rejected: %#v", apiErr)
	}

	contract, ok := ResolveTemplateContract(DefaultTemplateID, DefaultTemplateVersion)
	if !ok {
		t.Fatal("default template contract must resolve")
	}
	baseRequest := CreateReleaseRequest{
		SnapshotID:              mustTestUUID("00000000-0000-0000-0000-000000000001"),
		ClientTxnID:             "txn-render-reasons",
		TemplateID:              DefaultTemplateID,
		TemplateVersion:         DefaultTemplateVersion,
		RedactionProfileID:      InternalRedactionProfileID,
		RedactionProfileVersion: "1",
		OutputKind:              OutputKindSlidev,
		ReleaseScope:            ReleaseScopeInternalDraft,
	}
	baseModel := ExportModel{
		SchemaID:                     ExportModelSchemaID,
		IncidentID:                   "00000000-0000-0000-0000-000000000001",
		SnapshotAt:                   time.Date(2026, 5, 23, 12, 0, 0, 0, time.UTC),
		SourceChangeSetHighWatermark: SourceBoundaryTokenPrefix + strings.Repeat("3", 64),
		DerivationVersion:            DerivationVersion,
		Fields: []ExportField{
			{Path: "/incident/status", ContentClass: ContentClassDerivedAnalytic, Value: "open"},
			{Path: "/incident/title", ContentClass: ContentClassCuratedNarrative, Value: "Renderable", SupportRefs: []string{"/incident/status"}},
		},
	}
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
	reasonCases := []struct {
		name     string
		request  CreateReleaseRequest
		contract TemplateContract
		model    ExportModel
		want     string
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
			want:     "invalid_redaction_profile",
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
				model.Fields = append(model.Fields, ExportField{
					Path:          "/evidence/raw",
					ContentClass:  ContentClassSourceEvidence,
					Value:         "raw",
					RawBlobSource: true,
					OpaqueBinary:  true,
				})
				return model
			}(),
			want: "post_redaction_validation_failed",
		},
		{
			name: "template render failure",
			request: func() CreateReleaseRequest {
				req := baseRequest
				req.OutputKind = "html"
				req.ReleaseScope = ReleaseScopeExternal
				return req
			}(),
			contract: contract,
			model:    baseModel,
			want:     "template_render_failed",
		},
		{
			name:    "undeclared template binding",
			request: baseRequest,
			contract: func() TemplateContract {
				next := contract
				next.RenderBindings = append(next.RenderBindings, TemplateBinding{Name: "undeclared"})
				return next
			}(),
			model: baseModel,
			want:  "undeclared_template_binding",
		},
		{
			name:    "missing required field",
			request: baseRequest,
			contract: func() TemplateContract {
				next := contract
				next.RequiredFieldPaths = append(next.RequiredFieldPaths, "/incident/missing")
				return next
			}(),
			model: baseModel,
			want:  "missing_required_field",
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
				model.Fields = append(model.Fields, ExportField{
					Path:         "/incident/description",
					ContentClass: ContentClassCuratedNarrative,
					Value:        "![remote](https://cdn.example.test/a.png)",
					SupportRefs:  []string{"/incident/status"},
				})
				return model
			}(),
			want: "template_render_failed",
		},
		{
			name: "recognized composition op cannot silently no-op",
			request: func() CreateReleaseRequest {
				req := baseRequest
				req.CompositionJSON = json.RawMessage(`{
					"schema_id":"cartulary.report_composition.v1",
					"deck_ops":[{"op_id":"op-1","op_kind":"exclude_section","payload":{}}],
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
				req.GraphProjectionRefs = json.RawMessage(`[
					{"projection_schema_id":"graph_projection.v1","graph_view_id":"gv_bound","source_snapshot_id":"snap","projection_run_id":"run_1","projection_version":"1","projection_config_digest":"` + graphRenderSHA + `","projection_source_digest":"` + graphRenderSHA + `","projection_output_digest":"` + graphRenderSHA + `"}
				]`)
				req.CompositionJSON = json.RawMessage(`{
					"schema_id":"cartulary.report_composition.v1",
					"deck_ops":[],
					"diagram_decls":[{"decl_id":"diag-1","diagram_kind":"flowchart","diagram_source_kind":"graph","source_graph_view_id":"gv_missing","layout_mode":"auto"}],
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
				req.GraphProjectionRefs = json.RawMessage(`[
					{"projection_schema_id":"graph_projection.v1","graph_view_id":"gv_bound","source_snapshot_id":"snap","projection_run_id":"run_1","projection_version":"1","projection_config_digest":"` + graphRenderSHA + `","projection_source_digest":"` + graphRenderSHA + `","projection_output_digest":"` + graphRenderSHA + `"}
				]`)
				req.CompositionJSON = json.RawMessage(`{
					"schema_id":"cartulary.report_composition.v1",
					"deck_ops":[],
					"diagram_decls":[{"decl_id":"diag-1","diagram_kind":"flowchart","diagram_source_kind":"graph","source_graph_view_id":"gv_bound","layout_mode":"auto"}],
					"authored_texts":[]
				}`)
				return req
			}(),
			contract: contract,
			model:    baseModel,
			want:     "",
		},
		{
			name: "manual layout unsupported for mermaid",
			request: func() CreateReleaseRequest {
				req := baseRequest
				req.OutputKind = OutputKindMermaid
				req.CompositionJSON = json.RawMessage(`{
					"schema_id":"cartulary.report_composition.v1",
					"deck_ops":[],
					"diagram_decls":[{"decl_id":"diag-1","diagram_kind":"flowchart","diagram_source_kind":"timeline","layout_mode":"manual"}],
					"authored_texts":[]
				}`)
				return req
			}(),
			contract: contract,
			model:    baseModel,
			want:     "manual_layout_not_supported_for_output_kind",
		},
		{
			name: "mermaid labels reject raw angle brackets",
			request: func() CreateReleaseRequest {
				req := baseRequest
				req.OutputKind = OutputKindMermaid
				return req
			}(),
			contract: contract,
			model: func() ExportModel {
				model := baseModel
				model.Fields = append(model.Fields, ExportField{
					Path:         "/incident/<bad>",
					ContentClass: ContentClassDerivedAnalytic,
					Value:        "unsafe",
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
			_, reasonCode, err := renderReleaseCandidate(tc.request, tc.contract, tc.model, hashHex(source))
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
			return json.Marshal(value)
		}
		defer func() { encodeCanonicalJSON = original }()

		source, err := canonicalJSON(baseModel)
		if err != nil {
			t.Fatalf("source json: %v", err)
		}
		_, reasonCode, err := renderReleaseCandidate(baseRequest, contract, baseModel, hashHex(source))
		if err == nil || reasonCode != "manifest_encoding_failed" {
			t.Fatalf("manifest encoding reason = %q, err=%v", reasonCode, err)
		}
	})
}

func TestRenderBundleManifestBindsOutputHash(t *testing.T) {
	contract, ok := ResolveTemplateContract(DefaultTemplateID, DefaultTemplateVersion)
	if !ok {
		t.Fatal("resolve template contract")
	}
	model := RedactedExportModel{
		SchemaID:          ExportModelSchemaID,
		DerivationVersion: DerivationVersion,
		Fields: []RedactedField{
			{
				Path:         "/incident/title",
				ContentClass: ContentClassDerivedAnalytic,
				Value:        "Example",
			},
		},
	}
	redactionManifestSHA256 := strings.Repeat("a", 64)
	bundle, err := renderReportBundle(contract, OutputKindSlidev, model, redactionManifestSHA256, ReleaseScopeInternalDraft, nil, nil, nil, RedactionBundleArtifacts{})
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
	if bundle.Manifest.ToolchainSnapshotSHA256 == nil || bundle.Manifest.SandboxObservationSHA256 == nil || bundle.Manifest.ValidationSummarySHA256 == "" {
		t.Fatalf("bundle manifest must bind toolchain, sandbox, and validation artifacts: %#v", bundle.Manifest)
	}
}

func TestRenderBundleExternalReleaseRecordsDeterminismDigest(t *testing.T) {
	contract, ok := ResolveTemplateContract(DefaultTemplateID, DefaultTemplateVersion)
	if !ok {
		t.Fatal("resolve template contract")
	}
	model := RedactedExportModel{
		SchemaID:          ExportModelSchemaID,
		DerivationVersion: DerivationVersion,
		Fields: []RedactedField{
			{
				Path:         "/incident/title",
				ContentClass: ContentClassCuratedNarrative,
				Value:        "External deterministic report",
			},
		},
	}
	first, err := renderReportBundle(contract, OutputKindSlidev, model, strings.Repeat("a", 64), ReleaseScopeExternal, nil, nil, nil, RedactionBundleArtifacts{})
	if err != nil {
		t.Fatalf("build external render bundle: %v", err)
	}
	second, err := renderReportBundle(contract, OutputKindSlidev, model, strings.Repeat("a", 64), ReleaseScopeExternal, nil, nil, nil, RedactionBundleArtifacts{})
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

func TestRenderBundleCarriesRedactionArtifactsWithSensitiveRevealRole(t *testing.T) {
	contract, ok := ResolveTemplateContract(DefaultTemplateID, DefaultTemplateVersion)
	if !ok {
		t.Fatal("resolve template contract")
	}
	model := RedactedExportModel{
		SchemaID:          ExportModelSchemaID,
		DerivationVersion: DerivationVersion,
		Fields: []RedactedField{
			{
				Path:         "/incident/title",
				ContentClass: ContentClassDerivedAnalytic,
				Value:        "Example",
			},
		},
	}
	bundle, err := renderReportBundle(contract, OutputKindSlidev, model, strings.Repeat("a", 64), ReleaseScopeInternalDraft, nil, nil, nil, RedactionBundleArtifacts{
		RedactionManifestJSON: []byte(`{"schema_id":"` + RedactionManifestSchemaID + `"}`),
		ProfileViewJSON:       []byte(`{"schema_id":"` + RedactionProfileViewSchemaID + `"}`),
		TokenManifestJSON:     []byte(`{"schema_id":"` + RedactionTokenManifestSchemaID + `"}`),
		TokenManifestSHA256:   strings.Repeat("b", 64),
		RevealMapJSON:         []byte(`{"schema_id":"` + RedactionRevealMapSchemaID + `"}`),
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
	if roles[renderBundleRoleSensitiveRevealMap].Path != "internal/reveal-map.json" ||
		roles[renderBundleRoleSensitiveRevealMap].StorageKind != renderBundleStorageInline {
		t.Fatalf("reveal map must be persisted only as an internal sensitive bundle artifact, got %#v", roles[renderBundleRoleSensitiveRevealMap])
	}
	if bundle.Manifest.TokenManifestSHA256 == nil || *bundle.Manifest.TokenManifestSHA256 != strings.Repeat("b", 64) {
		t.Fatalf("bundle manifest must bind token manifest digest, got %#v", bundle.Manifest)
	}
	if len(bundle.Manifest.Files) < 9 || bundle.ManifestSHA256 == "" {
		t.Fatalf("bundle manifest must bind primary output and redaction artifacts, got %#v", bundle.Manifest)
	}
}

func TestRenderBundleMermaidPipelineEmitsSourceSVGAndDeterministicManifest(t *testing.T) {
	contract, ok := ResolveTemplateContract(DefaultTemplateID, DefaultTemplateVersion)
	if !ok {
		t.Fatal("resolve template contract")
	}
	model := RedactedExportModel{
		SchemaID:          ExportModelSchemaID,
		DerivationVersion: DerivationVersion,
		Fields: []RedactedField{
			{Path: "/incident/title", ContentClass: ContentClassCuratedNarrative, Value: "Example"},
			{Path: "/incident/status", ContentClass: ContentClassDerivedAnalytic, Value: "open"},
		},
	}
	first, err := renderReportBundle(contract, OutputKindMermaid, model, strings.Repeat("a", 64), ReleaseScopeInternalReview, nil, nil, nil, RedactionBundleArtifacts{})
	if err != nil {
		t.Fatalf("render mermaid bundle: %v", err)
	}
	second, err := renderReportBundle(contract, OutputKindMermaid, model, strings.Repeat("a", 64), ReleaseScopeInternalReview, nil, nil, nil, RedactionBundleArtifacts{})
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
	if roles[renderBundleRoleSourceMermaid].Path != "diagrams/default.mmd" ||
		!strings.Contains(string(roles[renderBundleRoleSourceMermaid].Bytes), "flowchart TD") {
		t.Fatalf("mermaid source missing or malformed: %#v", roles[renderBundleRoleSourceMermaid])
	}
	if roles[renderBundleRoleRenderedSVG].Path != "diagrams/default.svg" ||
		!strings.Contains(string(roles[renderBundleRoleRenderedSVG].Bytes), "<svg") {
		t.Fatalf("mermaid SVG missing or malformed: %#v", roles[renderBundleRoleRenderedSVG])
	}
	if roles[renderBundleRoleValidationSummary].Path != "validation/summary.json" ||
		roles[renderBundleRoleToolchainSnapshot].Path != "validation/toolchain.json" {
		t.Fatalf("validation/toolchain artifacts missing: %#v", first.Files)
	}
}

func TestRenderBundleRejectsUnsafeSVGLabelInput(t *testing.T) {
	contract, ok := ResolveTemplateContract(DefaultTemplateID, DefaultTemplateVersion)
	if !ok {
		t.Fatal("resolve template contract")
	}
	model := RedactedExportModel{
		SchemaID:          ExportModelSchemaID,
		DerivationVersion: DerivationVersion,
		Fields: []RedactedField{
			{Path: "/incident/url(evil)", ContentClass: ContentClassDerivedAnalytic, Value: "unsafe"},
		},
	}
	if _, err := renderReportBundle(contract, OutputKindMermaid, model, strings.Repeat("a", 64), ReleaseScopeInternalReview, nil, nil, nil, RedactionBundleArtifacts{}); err == nil {
		t.Fatalf("unsafe SVG label input must fail closed")
	}
}

func mustTestUUID(value string) uuid.UUID {
	parsed, err := uuid.Parse(value)
	if err != nil {
		panic(err)
	}
	return parsed
}
