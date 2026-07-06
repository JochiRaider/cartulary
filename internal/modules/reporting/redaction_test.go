package reporting

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestPhase11_U_11_REPORTING_01_RedactionProfilePrecedenceActionsAndManifest(t *testing.T) {
	descriptionPath := "/incident/description"
	sourceClass := ContentClassSourceEvidence
	maxChars := 4
	maskText := "[MASKED]"
	profile := RedactionProfile{
		SchemaID:  "cartulary.redaction_profile.v1",
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
}

func TestPhase11_U_11_REPORTING_02_RedactionProfileRejectsConflictsHashAndUnsafeBounds(t *testing.T) {
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
				SchemaID:      "cartulary.redaction_profile.v1",
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
				SchemaID:      "cartulary.redaction_profile.v1",
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
				SchemaID:      "cartulary.redaction_profile.v1",
				ProfileID:     "test.hash",
				Version:       "1",
				DefaultAction: RedactionActionSpec{Type: ActionHash},
			},
		},
		{
			name: "truncate missing safe bound",
			profile: RedactionProfile{
				SchemaID:      "cartulary.redaction_profile.v1",
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

func TestPhase11_U_11_REPORTING_03_ExternalValidationRejectsOpaqueBytesAndWorkingMaterial(t *testing.T) {
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

func TestPhase11_U_11_REPORTING_04_DisclosurePartitionsAndCuratedSupportRefsFailClosed(t *testing.T) {
	allowedPath := "/incident/summary"
	restrictedPath := "/incident/restricted"
	profile := RedactionProfile{
		SchemaID:  "cartulary.redaction_profile.v1",
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

func TestSupportPhase11_BuildExportModelUsesReportingOwnedMetadataSnapshotStableHash(t *testing.T) {
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
	for _, field := range first.Fields {
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

func TestPhase11_U_11_REPORTING_05_DecoderNormalizationAndRegisteredReasons(t *testing.T) {
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
		{name: "remote script", kind: OutputKindHTML, output: `<script src="https://cdn.example.test/app.js"></script>`},
		{name: "remote stylesheet", kind: OutputKindHTML, output: `<link rel="stylesheet" href="//cdn.example.test/app.css">`},
		{name: "remote font css import", kind: OutputKindHTML, output: `<style>@import url("https://fonts.example.test/font.css");</style>`},
		{name: "remote css url", kind: OutputKindHTML, output: `<style>body{background:url(https://cdn.example.test/bg.png)}</style>`},
		{name: "remote image", kind: OutputKindHTML, output: `<img src="https://cdn.example.test/a.png">`},
		{name: "markdown remote image", kind: OutputKindMarkdown, output: `![remote](https://cdn.example.test/a.png)`},
		{name: "markdown raw remote image", kind: OutputKindMarkdown, output: `<img src="https://cdn.example.test/a.png">`},
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
		if err := ValidateSelfContainedOutput(OutputKindHTML, output); err != nil {
			t.Fatalf("plain escaped URL text should remain renderable, got %v", err)
		}
	})
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
				req.OutputKind = OutputKindHTML
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
	}
	for _, tc := range reasonCases {
		t.Run(tc.name, func(t *testing.T) {
			source, err := canonicalJSON(tc.model)
			if err != nil {
				t.Fatalf("source json: %v", err)
			}
			_, reasonCode, err := renderReleaseCandidate(tc.request, tc.contract, tc.model, hashHex(source))
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

func mustTestUUID(value string) uuid.UUID {
	parsed, err := uuid.Parse(value)
	if err != nil {
		panic(err)
	}
	return parsed
}
