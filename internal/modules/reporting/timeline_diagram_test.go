package reporting

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"
)

func fixtureTimelineComposition(t *testing.T, contract TemplateContract, incidentID, kind string, ids []string) json.RawMessage {
	t.Helper()
	decl := fixtureTimelineDeclaration(kind, ids)
	raw, err := json.Marshal(map[string]any{
		"schema_id": "cartulary.report_composition.v1", "deck_ops": []any{}, "authored_texts": []any{},
		"diagram_decls": []any{map[string]any{
			"decl_id": decl.DeclID, "replaces_decl_id": nil, "diagram_kind": kind,
			"diagram_source_kind": "timeline", "source_graph_view_id": nil,
			"selection_rule": decl.SelectionRule, "layout_mode": "auto", "layout": nil, "label_overrides": []any{},
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	return sealedReportingCompositionFixture(t, raw, incidentID, contract)
}

func fixtureTimelineDeclaration(kind string, ids []string) compositionDiagramDecl {
	return compositionDiagramDecl{
		DeclID: "timeline", DiagramKind: kind, DiagramSourceKind: "timeline", LayoutMode: "auto",
		SelectionRule: diagramSelectionRule{SchemaID: "diagram_selection_rule.v1", RuleID: "timeline_sequence",
			VertexRefs: []string{}, EdgeRefs: []string{}, SeedVertexRefs: []string{}, EdgeKindFilter: []string{},
			TimelineRecordIDs: ids, OverflowPolicy: "fail"},
	}
}

func fixtureRetainedTimelineEvents() []ReportingTimelineEvent {
	early, late := "2026-04-01T00:00:00.000000Z", "2026-04-02T00:00:00.000000Z"
	return []ReportingTimelineEvent{
		{RecordID: "event-early", ActivitySortTS: &early, ActivityParseState: "parsed", ActivityPrecisionRank: 6, RecordCreatedAt: early},
		{RecordID: "event-late", ActivitySortTS: &late, ActivityParseState: "parsed", ActivityPrecisionRank: 6, RecordCreatedAt: early},
	}
}

func testTimelineDiagramSelection(t *testing.T) {
	events := fixtureRetainedTimelineEvents()
	decl := fixtureTimelineDeclaration("flowchart", []string{"event-late", "event-early"})
	diagram, err := timelineDiagramFromEvents(decl, events, ReleaseScopeInternalReview)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(diagram.IncludedVertexRefs, []string{"event-early", "event-late"}) || diagram.SourceProjectionRef != nil || len(diagram.Nodes) != 2 || len(diagram.Edges) != 1 || diagram.Edges[0].Label != "next" || diagram.Nodes[0].Label != "event-early" {
		t.Fatal("Timeline topology, ordering or label fallback is not exact", diagram)
	}
	source, err := serializeMermaidSource(diagram)
	if err != nil || !strings.Contains(string(source), `n0001 -->|"next"| n0002`) || strings.Contains(string(source), "2026-") {
		t.Fatal("private sort data became display text", string(source), err)
	}
	ref, err := timelineAdjacencyRef("event-early", "event-late")
	if err != nil || diagram.IncludedEdgeRefs[0] != ref {
		t.Fatal("adjacency identity mismatch", err)
	}
	reverse, _ := timelineAdjacencyRef("event-late", "event-early")
	if reverse == ref {
		t.Fatal("edge identity lost direction")
	}
	for name, ids := range map[string][]string{"missing": {"absent"}, "empty": {}, "duplicate": {"event-early", "event-early"}} {
		t.Run(name, func(t *testing.T) {
			_, err := timelineDiagramFromEvents(fixtureTimelineDeclaration("flowchart", ids), events, ReleaseScopeInternalReview)
			var invalid *renderValidationError
			want := "diagram_selection_missing_ref"
			if name == "duplicate" {
				want = "diagram_selection_duplicate_ref"
			}
			if !errors.As(err, &invalid) || invalid.ReasonCode != want {
				t.Fatal("invalid selection admitted", err)
			}
		})
	}
	for _, kind := range []string{"flowchart", "sequence"} {
		limit, _ := diagramVertexLimit(kind)
		many := make([]ReportingTimelineEvent, limit+1)
		ids := make([]string, len(many))
		for i := range many {
			many[i] = events[0]
			many[i].RecordID = fmt.Sprintf("event-%03d", i)
			ids[i] = many[i].RecordID
		}
		d := fixtureTimelineDeclaration(kind, ids)
		if _, err := timelineDiagramFromEvents(d, many, ReleaseScopeInternalReview); err == nil {
			t.Fatal("hard limit ignored", kind)
		}
		d.SelectionRule.OverflowPolicy = "summarize"
		bounded, err := timelineDiagramFromEvents(d, many, ReleaseScopeInternalReview)
		if err != nil || len(bounded.Nodes) != limit || len(bounded.Edges) != limit-1 || bounded.OverflowSummary == nil || bounded.OverflowSummary.OmittedVertexCount != 1 || bounded.OverflowSummary.OmittedEdgeCount != 1 || *bounded.OverflowSummary.FirstOmittedRef != ids[limit] {
			t.Fatal("incorrect bounded selection", kind, err)
		}
		if _, err := serializeMermaidSource(bounded); err != nil {
			t.Fatal(err)
		}
		if _, err := timelineDiagramFromEvents(d, many, ReleaseScopeExternal); err == nil {
			t.Fatal("composition authorized external truncation")
		}
	}
	override := diagramLabelOverride{Label: "Presentation"}
	override.Target.Kind, override.Target.Ref = "vertex", "event-early"
	if err := applyDiagramLabelOverrides(&diagram, []diagramLabelOverride{override}); err != nil || diagram.Nodes[0].Label != "Presentation" {
		t.Fatal("exact label target unresolved", err)
	}
	override.Target.Ref = "n0001"
	if err := applyDiagramLabelOverrides(&diagram, []diagramLabelOverride{override}); err == nil {
		t.Fatal("renderer identity used as source ref")
	}
}

func testRetainedTimeline(t *testing.T) {
	model := fixtureStructuredRenderModel(t, []ExportField{
		{Path: "/timeline/a", SourceFamily: "timeline_event", SourceRecordID: "a", SourceRecordType: "timeline_event", ContentClass: ContentClassSourceEvidence, Value: map[string]any{"capture_state": "rough", "value": "raw-a"}},
		{Path: "/timeline/b", SourceFamily: "timeline_event", SourceRecordID: "b", SourceRecordType: "timeline_event", ContentClass: ContentClassSourceEvidence, Value: map[string]any{"capture_state": "rough", "value": "raw-b"}},
	})
	before, _ := canonicalJSON(model)
	profile, _, err := ResolveRedactionProfile(InternalRedactionProfileID, DefaultRedactionProfileVersion, nil, model.SnapshotContent)
	if err != nil {
		t.Fatal(err)
	}
	paths := fixtureModelPaths(t, model.SnapshotContent, "/timeline/a/capture_state", "/timeline/a/value", "/timeline/b/capture_state", "/timeline/b/value")
	path, replacement := paths["/timeline/a/value"], "retained"
	profile.DefaultAction = RedactionActionSpec{Type: ActionDrop}
	profile.Rules = []RedactionRule{{RuleID: "mask-a", Path: &path, Action: RedactionActionSpec{Type: ActionMask, ReplacementText: &replacement}}}
	fixtureAllowContainerRules(t, &profile, model.SnapshotContent)
	digest, err := ValidateRedactionProfile(profile)
	if err != nil {
		t.Fatal(err)
	}
	result, err := RedactExportModel(model, fixtureSourceFactsForContent(model.SnapshotContent), profile, digest, "", ReleaseScopeInternalReview, nil)
	if err != nil {
		t.Fatal(err)
	}
	events := result.Model.TimelineEvents
	if len(events) != 1 || events[0].RecordID != "a" || len(events[0].Fields) != 1 || events[0].Fields[0].Value != "retained" || events[0].DisplayTimes != (ReportingDisplayTimes{}) {
		t.Fatal("raw or removed event retained", events)
	}
	if _, err := timelineDiagramFromEvents(fixtureTimelineDeclaration("flowchart", []string{"b"}), events, ReleaseScopeInternalReview); err == nil {
		t.Fatal("diagram resurrected removed event")
	}
	events[0].RecordCreatedAt = "changed"
	if events[0].ActivitySortTS != nil {
		*events[0].ActivitySortTS = "changed"
	}
	after, _ := canonicalJSON(model)
	if string(before) != string(after) {
		t.Fatal("retained event aliases source")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if got, err := projectRetainedTimeline(ctx, model.TimelineEvents, nil); !errors.Is(err, context.Canceled) || got != nil {
		t.Fatal("cancellation ignored")
	}
}

func testNoInventedDiagram(t *testing.T) {
	contract, _ := ResolveTemplateContract(DefaultTemplateID, DefaultTemplateVersion)
	model := RedactedExportModel{ExportModel: ExportModel{RenderIdentity: fixtureRenderIdentity()}, Fields: []RedactedField{{Value: "must not become a vertex"}}}
	bundle, err := fixtureRenderer().fixtureRenderReportBundle(context.Background(), contract, OutputKindMermaid, model, RedactionProfileView{}, strings.Repeat("a", 64), ReleaseScopeInternalReview, fixtureOutputOptions(OutputKindMermaid, ReleaseScopeInternalReview), nil, RedactionBundleArtifacts{})
	var invalid *renderValidationError
	if !errors.As(err, &invalid) || invalid.ReasonCode != "no_diagrams_selected" || len(bundle.Files) != 0 {
		t.Fatal("missing selection invented output", err)
	}
}
