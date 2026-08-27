package timeline

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/imports/ownerfacade"
	"github.com/JochiRaider/cartulary/internal/modules/timeline/mutationpolicy"
)

func TestTimelineImportNormalizerPreservesOwnerSemantics(t *testing.T) {
	t.Parallel()

	t.Run("visible text is preserved exactly", func(t *testing.T) {
		t.Parallel()
		raw := " \tTabbed\n=HYPERLINK(\"https://example.test\")\r "
		value, include, err := normalizeImportField(
			"timeline.activity_synopsis_text",
			raw,
			"omit_field",
		)
		text, textOK := value.Text()
		if err != nil || !include || value.Kind() != ownerfacade.ImportScalarText ||
			!textOK || text != raw {
			t.Fatalf(
				"unexpected Timeline visible-text normalization: value=%#v include=%t err=%v",
				value,
				include,
				err,
			)
		}
	})

	t.Run("visible text rune boundaries match mutation admission", func(t *testing.T) {
		t.Parallel()
		for _, runeCount := range []int{mutationpolicy.MaxVisibleTextRunes - 1, mutationpolicy.MaxVisibleTextRunes} {
			raw := strings.Repeat("界", runeCount)
			value, include, err := normalizeImportField(
				"timeline.activity_synopsis_text",
				raw,
				"omit_field",
			)
			text, textOK := value.Text()
			if err != nil || !include || !textOK || text != raw {
				t.Fatalf("%d-rune Timeline text mismatch: value=%#v include=%t err=%v", runeCount, value, include, err)
			}
		}
		if _, _, err := normalizeImportField(
			"timeline.activity_synopsis_text",
			strings.Repeat("界", mutationpolicy.MaxVisibleTextRunes+1),
			"omit_field",
		); err == nil {
			t.Fatal("32769-rune Timeline text unexpectedly succeeded")
		}
	})

	t.Run("mention and tag tokens use owner normalization", func(t *testing.T) {
		t.Parallel()
		mention, include, err := normalizeImportField(
			"timeline.host_refs",
			"  Host   One  ",
			"omit_field",
		)
		mentionToken, mentionOK := mention.CollectionToken()
		if err != nil || !include || !mentionOK ||
			mentionToken.RawText != "  Host   One  " ||
			mentionToken.NormalizedText != "Host One" {
			t.Fatalf(
				"unexpected Timeline mention normalization: value=%#v include=%t err=%v",
				mention,
				include,
				err,
			)
		}
		tag, include, err := normalizeImportField(
			"timeline.tags",
			" Urgent ",
			"omit_field",
		)
		tagToken, tagOK := tag.CollectionToken()
		if err != nil || !include || !tagOK ||
			tagToken.RawText != "Urgent" ||
			tagToken.NormalizedText != "urgent" {
			t.Fatalf(
				"unexpected Timeline tag normalization: value=%#v include=%t err=%v",
				tag,
				include,
				err,
			)
		}
	})

	t.Run("empty policy remains owner exact", func(t *testing.T) {
		t.Parallel()
		if _, include, err := normalizeImportField(
			"timeline.activity_synopsis_text",
			"",
			"omit_field",
		); err != nil || include {
			t.Fatalf("Timeline omit_field = include %t, error %v", include, err)
		}
		value, include, err := normalizeImportField(
			"timeline.activity_synopsis_text",
			"",
			"write_null",
		)
		if err != nil || !include || value.Kind() != ownerfacade.ImportScalarNull {
			t.Fatalf(
				"Timeline write_null = value %#v, include %t, error %v",
				value,
				include,
				err,
			)
		}
		if _, _, err := normalizeImportField(
			"timeline.host_refs",
			"",
			"write_null",
		); err == nil {
			t.Fatal("Timeline collection write_null unexpectedly succeeded")
		}
	})

	t.Run("invalid and unsupported fields fail closed", func(t *testing.T) {
		t.Parallel()
		if _, _, err := normalizeImportField(
			"timeline.activity_synopsis_text",
			strings.Repeat("x", 32769),
			"omit_field",
		); err == nil {
			t.Fatal("oversized Timeline visible text unexpectedly succeeded")
		}
		if _, _, err := normalizeImportField(
			"timeline.attached_evidence_ids",
			uuid.NewString(),
			"omit_field",
		); err == nil {
			t.Fatal("unsupported Timeline evidence import unexpectedly succeeded")
		}
	})
}

func TestTimelineImportRequestRetainsDurableProvenance(t *testing.T) {
	t.Parallel()

	sessionID := uuid.New()
	unitID := uuid.New()
	fieldKey := "timeline.activity_synopsis_text"
	text := "Imported row"
	request, err := timelineCreateRequestFromImport(
		ownerfacade.ImportOwnerCreateRequest{
			TargetViewSchemaID:  TimelineViewSchemaID,
			ImportSessionID:     sessionID,
			ImportUnitID:        unitID,
			MappingFingerprint:  strings.Repeat("a", 64),
			SourceFileKind:      "csv",
			SourceContentSHA256: strings.Repeat("b", 64),
			ParserProfileID:     "csv_rfc4180_v1",
			ParserVersion:       "1",
			LocatorKind:         "csv_file",
			Locator:             "file",
			SourceRectA1:        "A1:B2",
			SourceRowRef:        2,
			ClientTxnID:         "import-row-2",
			FieldValues: []ownerfacade.ImportFieldValue{{
				FieldKey:        fieldKey,
				NormalizedValue: ownerfacade.NewTextImportScalar(text),
			}},
			UnknownValues: []ownerfacade.ImportUnknownValue{{
				SourceColumnOrdinal: 2,
				SourceHeaderText:    "source_note",
				RawValue:            "raw",
				CellKind:            "inline_string",
			}},
		},
	)
	if err != nil {
		t.Fatalf("build Timeline owner import request: %v", err)
	}
	if request.ActivitySynopsisText == nil ||
		*request.ActivitySynopsisText != text ||
		len(request.RawCaptureColumns) != 1 {
		t.Fatalf("unexpected Timeline owner request: %#v", request)
	}
	provenance := request.RawCaptureColumns[0]
	if provenance.ImportSessionID != sessionID.String() ||
		provenance.ImportUnitID != unitID.String() ||
		provenance.SourceRowOrdinal != 2 ||
		provenance.SourceColumnOrdinal != 2 ||
		provenance.RawValue != "raw" {
		t.Fatalf("unexpected Timeline import provenance: %#v", provenance)
	}
}
