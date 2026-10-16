package reporting

import (
	"context"
	"encoding/json"
	"errors"
	"math"
	"strings"
	"testing"
)

func TestSlidevMarkdownEscapeV2(t *testing.T) {
	for _, tc := range []struct{ input, want string }{
		{"Café 漢字 (report).", "Café 漢字 (report)."},
		{"{{ 7 * 7 }}", "{{ 7 \\* 7 }}"},
		{"&lt; <script> &amp;", "\\&lt; \\<script> \\&amp;"},
		{"# title\n  > quote\n- list\n+ list\n= heading", "\\# title\n  \\> quote\n\\- list\n\\+ list\n\\= heading"},
		{"123. list\n  4) list\ntext 1.2", "123\\. list\n  4\\) list\ntext 1.2"},
		{"\\`|*_[]<~$&", "\\\\\\`\\|\\*\\_\\[\\]\\<\\~\\$\\&"},
		{"https://example.invalid/a_(b)?q=x#fragment", "https://example.invalid/a\\_(b)?q=x#fragment"},
	} {
		if got := escapeMarkdownText(tc.input); got != tc.want {
			t.Fatalf("escaped %q = %q, want %q", tc.input, got, tc.want)
		}
	}
}

func TestSlidevFieldTextCanonicalValues(t *testing.T) {
	for _, tc := range []struct {
		name  string
		value any
		want  string
	}{
		{"null", nil, ""},
		{"text", "Café 漢字 < &", "Café 漢字 < &"},
		{"paragraph lines", "first\nsecond", "first\nsecond"},
		{"integer", json.Number("-9007199254740991"), "-9007199254740991"},
		{"producer integer", int64(42), "42"},
		{"decoded integer", float64(42), "42"},
		{"true", true, "true"},
		{"false", false, "false"},
		{"empty array", []any{}, ""},
		{"scalar array", []any{"alpha", nil, json.Number("2"), false}, "alpha, , 2, false"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := fieldValueToText(tc.value, true)
			if err != nil || got != tc.want {
				t.Fatalf("text = %q, error = %v; want %q", got, err, tc.want)
			}
		})
	}
	for _, value := range []any{map[string]any{}, []any{[]any{}}, []any{map[string]any{}},
		[]byte("opaque"), "bad\xff", "bad\x00", "bad\t", "bad\r", "bad\u0085", "bad\u007f",
		json.Number("-0"), json.Number("+1"), json.Number("01"), json.Number("1.0"), json.Number("1e2"),
		json.Number("9007199254740992"), int64(9007199254740992), math.NaN(), math.Inf(1), 1.25, math.Copysign(0, -1),
	} {
		got, err := fieldValueToText(value, true)
		assertSlidevSourceFailure(t, []byte(got), err, "invalid_field_value_for_text")
	}
	for _, value := range []any{"first\nsecond", []any{"first\nsecond"}} {
		got, err := fieldValueToText(value, false)
		assertSlidevSourceFailure(t, []byte(got), err, "invalid_field_value_for_text")
	}
}

func fixtureNote(value string) *string { return &value }

func TestSlidevSpeakerNotesRejectBeforeOutput(t *testing.T) {
	deck := reportingDeckModel{Title: "Report", Slides: []reportingDeckSlide{
		{Title: "Safe slide", Layout: "default", SpeakerNotes: fixtureNote("First note\nSecond note")},
	}}
	template, _ := ResolveTemplateContract(DefaultTemplateID, DefaultTemplateVersion)
	data, err := serializeSlidevMarkdown(deck, template)
	if err != nil || !strings.Contains(string(data), "<!--\nFirst note\nSecond note\n-->\n") || strings.Count(string(data), "<!--") != 1 {
		t.Fatalf("notes must be a single final block: %q %v", data, err)
	}
	for _, note := range []string{"-->", "prefix --> suffix", "line\n-->\n<script>"} {
		deck.Slides = append(deck.Slides[:1], reportingDeckSlide{Title: "Invalid slide", Layout: "default", SpeakerNotes: fixtureNote(note)})
		data, err := serializeSlidevMarkdown(deck, template)
		assertSlidevSourceFailure(t, data, err, "invalid_speaker_note_sequence")
		if strings.Contains(err.Error(), "prefix") || strings.Contains(err.Error(), "script") {
			t.Fatal("diagnostic leaked note content")
		}
	}
	for _, note := range []string{"bad\x00", "bad\u0085", "bad\xff", "bad\r"} {
		deck.Slides[1].SpeakerNotes = fixtureNote(note)
		data, err := serializeSlidevMarkdown(deck, template)
		assertSlidevSourceFailure(t, data, err, "invalid_field_value_for_text")
	}
}

func TestSlidevInvalidFieldPublishesNoBundle(t *testing.T) {
	contract, ok := ResolveTemplateContract(DefaultTemplateID, DefaultTemplateVersion)
	if !ok {
		t.Fatal("missing template")
	}
	model := RedactedExportModel{ExportModel: ExportModel{RenderIdentity: fixtureRenderIdentity()}, Fields: []RedactedField{
		{Path: "/incident/status", ContentClass: ContentClassSourceEvidence, Value: "active"}, {Path: "/incident/title", ContentClass: ContentClassSourceEvidence, Value: map[string]any{"private": "must not serialize"}},
	}}
	bindFixtureRetainedSections(&model)
	bundle, err := fixtureRenderer().fixtureRenderReportBundle(context.Background(), contract, OutputKindSlidev, model, RedactionProfileView{}, strings.Repeat("a", 64), ReleaseScopeInternalDraft, fixtureOutputOptions(OutputKindSlidev, ReleaseScopeInternalDraft), nil, RedactionBundleArtifacts{})
	assertSlidevSourceFailure(t, bundle.ManifestJSON, err, "invalid_field_value_for_text")
	if len(bundle.Files) != 0 || bundle.ManifestSHA256 != "" {
		t.Fatal("failed source validation published bundle material")
	}
	model.Fields[1].Value = []any{"alpha", nil, 2, false}
	bindFixtureRetainedSections(&model)
	bundle, err = fixtureRenderer().fixtureRenderReportBundle(context.Background(), contract, OutputKindSlidev, model, RedactionProfileView{}, strings.Repeat("a", 64), ReleaseScopeInternalDraft, fixtureOutputOptions(OutputKindSlidev, ReleaseScopeInternalDraft), nil, RedactionBundleArtifacts{})
	if err != nil {
		t.Fatal(err)
	}
	for _, file := range bundle.Files {
		if file.Path == "slides.md" && strings.Contains(string(file.Bytes), "alpha, , 2, false") {
			return
		}
	}
	t.Fatal("bundle did not use scalar-array source conversion")
}

func assertSlidevSourceFailure(t *testing.T, data []byte, err error, reason string) {
	t.Helper()
	var failure *renderValidationError
	if len(data) != 0 || !errors.As(err, &failure) || failure.FailureCode != "slidev_source_invalid" || failure.ReasonCode != reason {
		t.Fatalf("expected no bytes and slidev_source_invalid/%s; bytes=%q error=%v", reason, data, err)
	}
}
