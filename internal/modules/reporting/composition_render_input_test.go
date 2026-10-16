package reporting

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"testing"
)

func sealedReportingCompositionFixture(t *testing.T, body []byte, incidentID string, contract TemplateContract) json.RawMessage {
	t.Helper()
	var object map[string]any
	if err := json.Unmarshal(body, &object); err != nil {
		t.Fatal(err)
	}
	object["composition_id"] = "composition-1"
	object["composition_version"] = "v1"
	object["incident_id"] = incidentID
	object["template_id"], object["template_version"] = contract.TemplateID, contract.TemplateVersion
	object["authored_against_snapshot_id"] = nil
	delete(object, "composition_sha256")
	base, err := canonicalJSON(object)
	if err != nil {
		t.Fatal(err)
	}
	digest := sha256.Sum256(base)
	object["composition_sha256"] = hex.EncodeToString(digest[:])
	sealed, err := canonicalJSON(object)
	if err != nil {
		t.Fatal(err)
	}
	return sealed
}

func TestCompositionRenderBinding_Unit(t *testing.T) {
	t.Run("presentation text redaction", testCompositionTextRedaction)
	t.Run("external authored text permission", testCompositionAuthoredTextPermission)
	t.Run("authored prose is presentation text", func(t *testing.T) {
		blocks := []ReportingBlock{{BlockID: "derived", BlockKind: "paragraph", Fields: []ReportingField{{Value: "source"}}}}
		slides := []reportingDeckSlide{{Blocks: blocks, sectionOrigin: &deckSectionOrigin{DeclarationID: "summary"}, derivedBlockIDs: derivedDeckBlockIDs(blocks)}}
		composition := &renderComposition{
			AuthoredTexts: []compositionAuthoredText{{AuthoredTextID: "prose", TextRole: "authored_text", Body: "Presenter context", DisclosurePartitionRef: "public"}},
			DeckOps:       []compositionDeckOp{{OpKind: "insert_authored_block", Payload: compositionOpPayload{AuthoredTextRef: "prose", Position: "after", BlockAnchor: &compositionBlockAnchor{AnchorKind: "block_anchor", SectionAnchor: compositionSectionAnchor{AnchorKind: "section_anchor", DeclarationID: "summary"}, BlockKind: "paragraph"}}}},
		}
		if err := applyCompositionToDeck(&slides, composition, TemplateContract{}, ReleaseScopeInternalDraft); err != nil {
			t.Fatal(err)
		}
		if len(slides[0].Blocks) != 2 || slides[0].Blocks[1].BlockKind != "paragraph" || slides[0].Blocks[1].ContentClass != "presentation_text" {
			t.Fatal("authored prose was misclassified as a curated case claim")
		}
	})
	t.Run("semantic operations", testCompositionDeckOperations)
	t.Run("record-qualified anchors", testCompositionRecordAnchors)
	contract := TemplateContract{TemplateID: "template", TemplateVersion: "1"}
	model := RedactedExportModel{ExportModel: ExportModel{SnapshotContent: SnapshotContent{IncidentID: "incident"}}}
	raw := sealedReportingCompositionFixture(t, []byte(`{"schema_id":"cartulary.report_composition.v1","deck_ops":[],"diagram_decls":[],"authored_texts":[]}`), model.IncidentID, contract)
	if err := validateCompositionRenderBinding(raw, contract, model, ReleaseScopeInternalReview, RedactionProfileView{}); err != nil {
		t.Fatal(err)
	}
	for name, change := range map[string]func(*TemplateContract, *RedactedExportModel){
		"template":         func(c *TemplateContract, _ *RedactedExportModel) { c.TemplateVersion = "2" },
		"incident":         func(_ *TemplateContract, m *RedactedExportModel) { m.IncidentID = "other" },
		"preview identity": func(_ *TemplateContract, m *RedactedExportModel) { value := "preview"; m.PreviewAttemptID = &value },
	} {
		t.Run(name, func(t *testing.T) {
			c, m := contract, model
			change(&c, &m)
			var invalid *renderValidationError
			if err := validateCompositionRenderBinding(raw, c, m, ReleaseScopeInternalReview, RedactionProfileView{}); !errors.As(err, &invalid) || invalid.FailureCode != "composition_invalid" {
				t.Fatalf("binding admitted: %v", err)
			}
		})
	}
	for _, invalid := range []json.RawMessage{json.RawMessage(`{}`), json.RawMessage(`null`), append(json.RawMessage(" "), raw...)} {
		if _, err := parseRenderComposition(invalid); err == nil {
			t.Fatal("unvalidated composition projection")
		}
	}
	drop := sealedReportingCompositionFixture(t, []byte(`{"schema_id":"cartulary.report_composition.v1","deck_ops":[{"op_id":"drop","op_kind":"exclude_section","on_unresolved":"drop","payload":{"section_anchor":{"anchor_kind":"section_anchor","template_section_decl_id":"missing","expansion_key":null}}}],"diagram_decls":[],"authored_texts":[]}`), model.IncidentID, contract)
	var failure *renderValidationError
	if err := validateCompositionRenderBinding(drop, contract, model, ReleaseScopeExternal, RedactionProfileView{}); !errors.As(err, &failure) || failure.ReasonCode != "composition_drop_invalid_for_external_release" {
		t.Fatal("external drop escaped common output-kind boundary", err)
	}
}

func testCompositionAuthoredTextPermission(t *testing.T) {
	contract := TemplateContract{TemplateID: "template", TemplateVersion: "1"}
	model := RedactedExportModel{ExportModel: ExportModel{SnapshotContent: SnapshotContent{IncidentID: "incident"}}}
	for _, role := range []string{"title_override", "speaker_notes", "authored_text"} {
		t.Run(role, func(t *testing.T) {
			body, err := json.Marshal(map[string]any{
				"schema_id": "cartulary.report_composition.v1", "deck_ops": []any{}, "diagram_decls": []any{},
				"authored_texts": []any{map[string]any{"authored_text_id": "text", "text_role": role, "body": "Unreleased presentation text", "disclosure_partition_ref": "public"}},
			})
			if err != nil {
				t.Fatal(err)
			}
			raw := sealedReportingCompositionFixture(t, body, model.IncidentID, contract)
			for _, profile := range []RedactionProfileView{{}, {ProfileID: ExternalRedactionProfileID, AllowAuthoredPresentationText: false}} {
				bundle, err := fixtureRenderer().fixtureRenderReportBundle(context.Background(), contract, OutputKindSlidev, model, profile, "unused", ReleaseScopeExternal, nil, raw, RedactionBundleArtifacts{})
				var invalid *renderValidationError
				if !errors.As(err, &invalid) || invalid.FailureCode != "composition_invalid" || invalid.ReasonCode != "authored_text_not_permitted" || len(bundle.Files) != 0 || len(bundle.ManifestJSON) != 0 {
					t.Fatal("external text reached rendering without permission", err)
				}
			}
			// These assertions cover this permission gate only. Partition,
			// redaction, anchors and publication remain independent obligations.
			if err := validateCompositionRenderBinding(raw, contract, model, ReleaseScopeExternal, RedactionProfileView{AllowAuthoredPresentationText: true}); err != nil {
				t.Fatal("explicit permission rejected by the permission gate", err)
			}
			for _, scope := range []string{ReleaseScopeInternalDraft, ReleaseScopeInternalReview} {
				if err := validateCompositionRenderBinding(raw, contract, model, scope, RedactionProfileView{}); err != nil {
					t.Fatal("external control changed internal applicability", err)
				}
			}
		})
	}
}
