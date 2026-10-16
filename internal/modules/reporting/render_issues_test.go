package reporting

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"testing"

	"github.com/JochiRaider/cartulary/internal/platform/jobs"
)

func testRenderIssueCollection(t *testing.T) {
	t.Run("complete model stage", testCompleteSnapshotValidation)
	payload, err := (&fakeReportingJobStore{}).ReleasePayloadForJob(context.Background(), jobs.Execution{})
	if err != nil {
		t.Fatal(err)
	}
	limits, err := resolveRenderLimits(nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, count := range []int{1, 2, 499, 500, 501, 2000, 2001, 2300} {
		t.Run(fmt.Sprint(count), func(t *testing.T) {
			var canonical []byte
			for _, reverse := range []bool{false, true} {
				collector := renderIssueCollector{stage: "slidev_source"}
				for i := 0; i < count; i++ {
					value := i
					if reverse {
						value = count - i - 1
					}
					collector.add(limits.check("slides.count", 121+int64(value)))
					if len(collector.set.retained) > retainedRenderIssueLimit {
						t.Fatal("unbounded retention")
					}
				}
				failure := renderFailureDetails(collector.finish())
				raw, err := failedRenderSummary(payload.renderInput, failure)
				if err != nil || validateFailedRenderSummary(raw, payload.renderInput, failure.ReasonCode) != nil {
					t.Fatal("invalid stage evidence", err)
				}
				if reverse && !bytes.Equal(canonical, raw) {
					t.Fatal("evaluation order changed canonical evidence")
				}
				canonical = raw
				var summary reportingRenderValidationSummary
				if err := json.Unmarshal(raw, &summary); err != nil {
					t.Fatal(err)
				}
				if summary.IssueCount != count || len(summary.Issues) != min(count, 500) {
					t.Fatal("lost evaluated count")
				}
				last := summary.Issues[len(summary.Issues)-1]
				if (last.Severity == "warning") != (count > 2000) {
					t.Fatal("incorrect retention warning")
				}
				if count > 2000 && (*last.ReasonCode != "validation_issue_limit_reached" || last.SafeDetails["actual"] != float64(count)) {
					t.Fatal("lost truncation reason/count")
				}
				if count < 2 {
					continue
				}
				for name, change := range map[string]func(*reportingRenderValidationSummary){
					"removed issue":   func(s *reportingRenderValidationSummary) { s.Issues = s.Issues[:len(s.Issues)-1] },
					"wrong order":     func(s *reportingRenderValidationSummary) { s.Issues[0], s.Issues[1] = s.Issues[1], s.Issues[0] },
					"unknown key":     func(s *reportingRenderValidationSummary) { s.Issues[1].SafeDetails["raw_error"] = "sensitive" },
					"different stage": func(s *reportingRenderValidationSummary) { s.Issues[1].Stage = "redaction" },
				} {
					var changed reportingRenderValidationSummary
					if err := json.Unmarshal(raw, &changed); err != nil {
						t.Fatal(err)
					}
					change(&changed)
					mutated, err := canonicalJSON(changed)
					if err != nil || validateFailedRenderSummary(mutated, payload.renderInput, failure.ReasonCode) == nil {
						t.Fatal("accepted modified diagnostics", name)
					}
				}
			}
		})
	}
	// The production validator must finish every slide and both independent
	// checks. Canonical failure choice must not depend on input traversal order.
	limits["slides.count"], limits["blocks_per_slide"], limits["speaker_notes_chars_per_slide"] = 1, 1, 1
	deck := reportingDeckModel{Slides: []reportingDeckSlide{
		{Blocks: make([]ReportingBlock, 2), SpeakerNotes: fixtureNote("xx")},
		{Blocks: make([]ReportingBlock, 3), SpeakerNotes: fixtureNote("xxx")},
	}}
	failure := renderFailureDetails(validateDeckLimits(limits, deck))
	if failure.issues == nil || failure.issues.count != 5 {
		t.Fatal("deck validation stopped early")
	}
	first, err := failedRenderSummary(payload.renderInput, failure)
	if err != nil {
		t.Fatal(err)
	}
	slices.Reverse(deck.Slides)
	second, err := failedRenderSummary(payload.renderInput, renderFailureDetails(validateDeckLimits(limits, deck)))
	if err != nil || !bytes.Equal(first, second) {
		t.Fatal("deck traversal selected terminal failure", err)
	}
	// No partial Mermaid source escapes on failure, and every node and edge
	// label is evaluated even when an earlier label already failed.
	diagram := reportingDiagramModel{DiagramKind: "flowchart",
		Nodes: []reportingDiagramNode{{NodeID: "a", Label: "<unsafe>"}, {NodeID: "b", Label: "<unsafe>"}},
		Edges: []reportingDiagramEdge{{EdgeID: "edge", From: "a", To: "b", Label: "<unsafe>"}},
	}
	source, err := serializeMermaidSource(diagram)
	failure = renderFailureDetails(err)
	if source != nil || failure.issues == nil || failure.issues.count != 3 {
		t.Fatal("Mermaid validation stopped before completing its stage")
	}
	collector := renderIssueCollector{stage: "mermaid_source"}
	collector.add(err)
	collector.add(err)
	failure = renderFailureDetails(collector.finish())
	if failure.issues.count != 6 {
		t.Fatal("nested stage collection lost evaluated issues")
	}
	deck = reportingDeckModel{Title: "Safe", Slides: []reportingDeckSlide{
		{Title: "bad\nheading", Layout: "unknown", Blocks: []ReportingBlock{{BlockKind: "unknown"}, {BlockKind: "paragraph", Fields: []ReportingField{{Value: "bad\x00"}}}}, SpeakerNotes: fixtureNote("bad\x00 -->")},
		{Title: "Safe", Layout: "default", SpeakerNotes: fixtureNote("-->")},
	}}
	template, _ := ResolveTemplateContract(DefaultTemplateID, DefaultTemplateVersion)
	template.Author = "Safe"
	source, err = serializeSlidevMarkdown(deck, template)
	failure = renderFailureDetails(err)
	if source != nil || failure.issues == nil || failure.issues.count != 7 {
		t.Fatal("Slidev validation stopped before completing its stage", failure.issues)
	}
}
