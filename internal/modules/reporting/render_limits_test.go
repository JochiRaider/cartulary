package reporting

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/platform/renderprocess"
)

func TestReportingDeclaredLimits(t *testing.T) {
	defaults, err := resolveRenderLimits(nil)
	if err != nil {
		t.Fatal(err)
	}
	if defaults.timeout("total_attempt") != 15*time.Minute || defaults.rerenderTimeout(OutputKindSlidev) != 10*time.Minute || defaults.rerenderTimeout(OutputKindMermaid) != time.Minute || defaults["rendered_pdf.byte_size"] != 250<<20 {
		t.Fatal("incorrect defaults")
	}
	for _, test := range []struct{ key, raw, reason string }{
		{"timeout.slidev_export", `0`, "invalid_template_timeout_key"},
		{"timeout.slidev_export", `1.0`, "invalid_template_timeout_key"},
		{"timeout.slidev_export", `1e2`, "invalid_template_timeout_key"},
		{"timeout.slidev_export", `"60s"`, "invalid_template_timeout_key"},
		{"timeout.slidev_export", `"1 MiB"`, "invalid_template_timeout_key"},
		{"timeout.slidev_export", `901`, "template_limit_exceeds_hard_limit"},
		{"timeout.unowned", `1`, "invalid_template_timeout_key"},
		{"unowned.count", `1`, "template_manifest_invalid"},
		{"validation_issues.count", `500`, "template_manifest_invalid"},
		{"validation_issues.count", `1`, "template_manifest_invalid"},
		{"sections.count", `"1 MiB"`, "invalid_scalar_value"},
		{"sections.count", `null`, "invalid_scalar_value"},
		{"sections.count", `true`, "invalid_scalar_value"},
		{"sections.count", `-1`, "invalid_scalar_value"},
		{"sections.count", `501`, "template_limit_exceeds_hard_limit"},
		{"rendered_pdf.byte_size", `"1.5 MiB"`, "invalid_scalar_value"},
		{"rendered_pdf.byte_size", `"01 MiB"`, "invalid_scalar_value"},
		{"rendered_pdf.byte_size", `"501 MiB"`, "template_limit_exceeds_hard_limit"},
		{"rendered_pdf.byte_size", `"9999999999999999999999999 MiB"`, "template_limit_exceeds_hard_limit"},
	} {
		t.Run(test.key+"/"+test.raw, func(t *testing.T) {
			_, err := resolveRenderLimits(map[string]json.RawMessage{test.key: json.RawMessage(test.raw)})
			var failure *renderValidationError
			if !errors.As(err, &failure) || failure.FailureCode != "invalid_release_request" || failure.ReasonCode != test.reason {
				t.Fatalf("got %v", err)
			}
		})
	}
	limits, err := resolveRenderLimits(map[string]json.RawMessage{"timeout.slidev_export": json.RawMessage(`2`), "timeout.determinism_rerender": json.RawMessage(`3`), "rendered_pdf.byte_size": json.RawMessage(`"500 MiB"`)})
	if err != nil || limits.rerenderTimeout(OutputKindSlidev) != 3*time.Second || limits["rendered_pdf.byte_size"] != 500<<20 {
		t.Fatal(limits, err)
	}
	if err := limits.check("rendered_pdf.byte_size", 500<<20); err != nil {
		t.Fatal(err)
	}
	var failure *renderValidationError
	if err := limits.check("rendered_pdf.byte_size", 500<<20+1); !errors.As(err, &failure) || failure.ReasonCode != "rendered_pdf_bytes_exceeded" {
		t.Fatal(err)
	}
	deck := reportingDeckModel{Slides: []reportingDeckSlide{{Blocks: make([]ReportingBlock, 30), SpeakerNotes: fixtureNote(strings.Repeat("é", 10000))}}}
	if err := validateDeckLimits(defaults, deck); err != nil {
		t.Fatal(err)
	}
	deck.Slides[0].SpeakerNotes = fixtureNote(*deck.Slides[0].SpeakerNotes + "x")
	if err := validateDeckLimits(defaults, deck); !errors.As(err, &failure) || failure.ReasonCode != "speaker_notes_limit_exceeded" {
		t.Fatal(err)
	}
	deck.Slides[0].SpeakerNotes = nil
	deck.Slides[0].Blocks = append(deck.Slides[0].Blocks, ReportingBlock{})
	if err := validateDeckLimits(defaults, deck); !errors.As(err, &failure) || failure.ReasonCode != "slide_block_limit_exceeded" {
		t.Fatal(err)
	}
}

type limitObservingRuntime struct {
	fixtureRenderRuntime
	deadline time.Time
	maxBytes int64
	err      error
}

func (r *limitObservingRuntime) Run(ctx context.Context, _ []byte, outputs []renderprocess.Output) (renderprocess.Result, error) {
	r.deadline, _ = ctx.Deadline()
	r.maxBytes = outputs[0].MaxBytes
	return renderprocess.Result{}, r.err
}

func TestReportingStageEnforcesDeclaredLimits(t *testing.T) {
	limits, err := resolveRenderLimits(map[string]json.RawMessage{"timeout.slidev_export": json.RawMessage(`2`), "rendered_pdf.byte_size": json.RawMessage(`1024`)})
	if err != nil {
		t.Fatal(err)
	}
	runtime := &limitObservingRuntime{err: renderprocess.ErrOutputLimit}
	start := time.Now()
	bytes, err := (&Renderer{runtime: runtime}).renderStage(context.Background(), renderStageRequest{Operation: "slidev_pdf"}, limits)
	var failure *renderValidationError
	if len(bytes) != 0 || !errors.As(err, &failure) || failure.ReasonCode != "rendered_pdf_bytes_exceeded" {
		t.Fatal(err)
	}
	if runtime.maxBytes != 1024 || runtime.deadline.Before(start.Add(2*time.Second)) || runtime.deadline.After(time.Now().Add(2*time.Second)) {
		t.Fatal("stage did not receive declared limits")
	}
	// Cleanup failure remains the primary security outcome even after a size violation.
	runtime.err = errors.Join(renderprocess.ErrOutputLimit, renderprocess.ErrCleanup)
	_, err = (&Renderer{runtime: runtime}).renderStage(context.Background(), renderStageRequest{Operation: "slidev_pdf"}, limits)
	if !errors.As(err, &failure) || failure.ReasonCode != "partial_output_cleanup_failed" {
		t.Fatal(err)
	}
	// A runtime returning bytes despite the requested cap cannot bypass admission.
	request := renderStageRequest{SchemaID: renderStageRequestSchema, Operation: "slidev_pdf", Source: "synthetic", IdentitySHA256: strings.Repeat("a", 64)}
	limits["rendered_pdf.byte_size"] = 1
	_, err = fixtureRenderer().renderStage(context.Background(), request, limits)
	if !errors.As(err, &failure) || failure.ReasonCode != "rendered_pdf_bytes_exceeded" {
		t.Fatal(err)
	}
	t.Run("deadline attribution and discarded late output", func(t *testing.T) {
		for _, key := range []string{"timeout.total_attempt", "timeout.slidev_export"} {
			ctx, cancel := withRenderDeadline(context.Background(), -time.Second, "slidev_export", key)
			defer cancel()
			late := &limitObservingRuntime{}
			output, err := (&Renderer{runtime: late}).renderStage(ctx, request, limits)
			var failure *renderValidationError
			if len(output) != 0 || !errors.As(err, &failure) || failure.ReasonCode != "render_timeout" || failure.SafeDetails["limit_key"] != key {
				t.Fatal("deadline was replaced", err)
			}
			kind := "stage"
			if key == "timeout.total_attempt" {
				kind = "total"
			}
			if failure.SafeDetails["timeout_kind"] != kind {
				t.Fatal("lost timer attribution", failure.SafeDetails)
			}
			if got := classifyRenderContextError(ctx, errors.Join(ctx.Err(), renderprocess.ErrCleanup)); got.ReasonCode != "partial_output_cleanup_failed" {
				t.Fatal("cleanup failure hidden", got)
			}
		}
	})
}

func TestReportingDeckChunkingPreservesContent(t *testing.T) {
	limits, err := resolveRenderLimits(map[string]json.RawMessage{"blocks_per_slide": json.RawMessage(`2`)})
	if err != nil {
		t.Fatal(err)
	}
	blocks := make([]ReportingBlock, 5)
	for i := range blocks {
		blocks[i] = ReportingBlock{BlockID: fmt.Sprintf("block-%d", i), Fields: []ReportingField{{FieldKey: "value", FieldOrdinal: 1, Value: fmt.Sprintf("value-%d", i)}}}
	}
	input := []reportingDeckSlide{{Title: "Evidence", Layout: "default", Blocks: blocks, SpeakerNotes: fixtureNote("whole section note")}, {Title: "Summary", Layout: "section", Blocks: []ReportingBlock{}, SpeakerNotes: nil}}
	slides, err := chunkDeckSlides(input, limits)
	if err != nil {
		t.Fatal(err)
	}
	if len(slides) != 4 || slides[0].SlideID != "sld-0001__chunk_0001" || slides[2].SlideID != "sld-0001__chunk_0003" || slides[3].SlideID != "sld-0004" || slides[2].Title != "Evidence (3/3)" {
		t.Fatal(slides)
	}
	var collected []ReportingBlock
	for i, slide := range slides {
		if slide.Ordinal != i+1 || len(slide.Blocks) > 2 {
			t.Fatal(slides)
		}
		if i > 0 && slide.SpeakerNotes != nil {
			t.Fatal("notes duplicated")
		}
		collected = append(collected, slide.Blocks...)
	}
	if len(collected) != len(blocks) || !reflect.DeepEqual(slides[0].SpeakerNotes, input[0].SpeakerNotes) {
		t.Fatal("chunking changed content")
	}
	for i, block := range collected {
		if block.BlockID != blocks[i].BlockID || block.Fields[0].Value != blocks[i].Fields[0].Value || block.BlockOrdinal != i%2+1 {
			t.Fatal("chunking changed retained identities, values or local ordinals")
		}
	}
	slides[0].Blocks[0].Fields[0].Value = "mutated result"
	if input[0].Blocks[0].Fields[0].Value == "mutated result" {
		t.Fatal("chunk shares mutable block storage")
	}
	withClicks := append([]reportingDeckSlide{}, input...)
	withClicks[0].ClickSteps = []reportingClickStep{{ClickStepID: "unbound"}}
	if result, err := chunkDeckSlides(withClicks, limits); err == nil || result != nil {
		t.Fatal("chunking silently discarded unbound click state")
	}
	limits["slides.count"] = 3
	if _, err := chunkDeckSlides(input, limits); err == nil {
		t.Fatal("chunking bypassed total slide limit")
	}
}
