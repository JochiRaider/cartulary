package reportcomposition

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/google/uuid"
)

func TestCompositionCanonicalDigests_Unit(t *testing.T) {
	// Independently stated UTF-8 bytes and Python hashlib SHA-256 vector.
	canonical := []byte("{\"n\":9007199254740991,\"nested\":{\"text\":\"R&D <review>\u2028\u2029\"},\"schema_id\":\"cartulary.report_composition.v1\"}")
	const expected = "1f4becedac6db2d612b1c5e19957c8f26fd7bfc81d0b302e8af6fb5a0f50361d"
	for _, member := range []string{"composition_sha256", "preview_source_sha256"} {
		raw := append(append([]byte(nil), canonical[:len(canonical)-1]...), []byte(",\""+member+"\":\"ignored\"}")...)
		digest := CompositionDigest
		if member == "preview_source_sha256" {
			digest = PreviewSourceDigest
		}
		got, err := digest(raw)
		if err != nil || got != expected {
			t.Fatalf("%s digest = %s, %v", member, got, err)
		}
		changed := bytes.ReplaceAll(raw, []byte("R&D"), []byte("R and D"))
		if altered, err := digest(changed); err != nil || altered == expected {
			t.Fatal("altered content not distinguished", err)
		}
		for _, invalid := range []string{`null`, `[]`, `{"x":1,"x":2}`, `{"nested":{"x":1,"x":2}}`, `{"n":1.0}`, `{"n":1e0}`, `{"n":-0}`, `{"n":9007199254740992}`, `{"x":"\ud800"}`, "{\"x\":\"\xff\"}"} {
			if _, err := digest([]byte(invalid)); err == nil {
				t.Errorf("%s admitted %q", member, invalid)
			}
		}
	}
	record := ResourceRecord{CompositionID: uuid.MustParse("10000000-0000-4000-8000-000000000001"), IncidentID: uuid.MustParse("20000000-0000-4000-8000-000000000002"), TemplateID: "template", TemplateVersion: "1", DraftVersion: 1, DeckOps: json.RawMessage(`[]`), DiagramDecls: json.RawMessage(`[]`), AuthoredTexts: json.RawMessage(`[]`)}
	data, sha, err := canonicalComposition(record, 1)
	if err != nil {
		t.Fatal(err)
	}
	if got, err := CompositionDigest(data); err != nil || got != sha {
		t.Fatal("immutable digest", got, err)
	}
	const expectedComposition = `{"authored_against_snapshot_id":null,"authored_texts":[],"composition_id":"10000000-0000-4000-8000-000000000001","composition_sha256":"6a79f96bb86f31c263169d3d5e32c58a8f3fffb11a31d047e0653fcc40bf7f0b","composition_version":"v1","deck_ops":[],"diagram_decls":[],"incident_id":"20000000-0000-4000-8000-000000000002","schema_id":"cartulary.report_composition.v1","template_id":"template","template_version":"1"}`
	if string(data) != expectedComposition {
		t.Fatalf("immutable canonical bytes = %s", data)
	}
	if summary := validateInlineComposition(record, data); !summaryValid(summary) || summary["composition_sha256"] != sha {
		t.Fatal("inline digest disagrees with version creation", summary)
	}
	for _, altered := range [][]byte{
		bytes.Replace(data, []byte(sha), []byte(strings.Repeat("0", 64)), 1),
		bytes.Replace(data, []byte(`"authored_texts":[]`), []byte(`"authored_texts":[{}]`), 1),
	} {
		if summaryValid(validateInlineComposition(record, altered)) {
			t.Fatal("inline validation accepted altered identity")
		}
	}
	version := &VersionRecord{CanonicalComposition: data, CompositionSHA256: sha, CompositionVersion: 1}
	for _, kind := range []string{"draft", "version"} {
		var selected *VersionRecord
		if kind == "version" {
			selected = version
		}
		source, expected, _, compositionSHA, err := previewSource(record, selected, PreviewRequest{SourceKind: kind})
		if err != nil {
			t.Fatal(err)
		}
		if got, err := PreviewSourceDigest(source); err != nil || got != expected {
			t.Fatal("preview digest", got, err)
		}
		if (kind == "draft") != (compositionSHA == nil) {
			t.Fatal("preview/release digest conflated")
		}
	}
	// Original raw tokens must survive until strict admission, including persisted data.
	for _, raw := range []string{`[{"n":1.0}]`, `[{"n":1,"n":2}]`, `[{"x":"\ud800"}]`} {
		record.AuthoredTexts = json.RawMessage(raw)
		if _, _, err := canonicalComposition(record, 1); err == nil {
			t.Errorf("composition repaired %s", raw)
		}
		if _, _, _, _, err := previewSource(record, nil, PreviewRequest{SourceKind: "draft"}); err == nil {
			t.Errorf("preview repaired %s", raw)
		}
		if _, apiErr := requiredArray(json.RawMessage(raw), "authored_texts"); apiErr == nil || apiErr.Status != 400 {
			t.Errorf("invalid raw input response: %#v", apiErr)
		}
	}
	omitted, e1 := DecodeCreateDraftRequest(strings.NewReader(`{"client_txn_id":"same","template_id":"template","template_version":"1"}`))
	explicit, e2 := DecodeCreateDraftRequest(strings.NewReader(`{"client_txn_id":"same","template_id":"template","template_version":"1","deck_ops":[],"diagram_decls":[],"authored_texts":[]}`))
	if e1 != nil || e2 != nil || !bytes.Equal(omitted.Normalized, explicit.Normalized) {
		t.Fatal("default materialization drift", e1, e2)
	}
	for _, invalid := range []string{`"\ud800"`, `"a","template_id":"b"`} {
		if _, apiErr := DecodeCreateDraftRequest(strings.NewReader(`{"client_txn_id":"same","template_id":` + invalid + `,"template_version":"1"}`)); apiErr == nil || apiErr.Status != 400 {
			t.Fatal("malformed scalar admitted", invalid, apiErr)
		}
	}
}
