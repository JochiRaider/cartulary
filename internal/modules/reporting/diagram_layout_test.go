package reporting

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

func manualDiagramFixture() reportingDiagramModel {
	return reportingDiagramModel{SchemaID: "cartulary.reporting_diagram.v1", DiagramID: "manual", DiagramKind: "flowchart", DiagramSourceKind: "timeline", LayoutMode: "manual",
		IncludedVertexRefs: []string{"opaque:b:ref", "opaque:a:ref"}, IncludedEdgeRefs: []string{"opaque:edge"},
		Nodes: []reportingDiagramNode{{NodeID: "n0001", Label: "B & literal"}, {NodeID: "n0002", Label: "A"}},
		Edges: []reportingDiagramEdge{{EdgeID: "e0001", From: "n0001", To: "n0002", Label: "next"}},
		CompositionLayout: &compositionDiagramLayout{SchemaID: "composition_diagram_layout.v1", CoordinateSpace: diagramCoordinateSpace{Unit: "css_px", Origin: "top_left", Width: 400, Height: 300},
			NodePositions: []diagramNodePosition{{TargetRef: "opaque:a:ref", X: 200, Y: 100, Width: 60, Height: 40}, {TargetRef: "opaque:b:ref", X: 10, Y: 20, Width: 80, Height: 40}}, EdgeRoutes: []diagramEdgeRoute{}},
	}
}

func testManualDiagramLayout(t *testing.T) {
	t.Run("declaration ownership and operations", testCompositionDiagramOperations)
	d := manualDiagramFixture()
	data, err := serializeManualDiagramSVG(d)
	if err != nil {
		t.Fatal(err)
	}
	want := "<svg xmlns=\"http://www.w3.org/2000/svg\" viewBox=\"0 0 400 300\">\n" +
		"  <path d=\"M90 40 L200 120\" fill=\"none\" stroke=\"#334155\" stroke-width=\"2\"/>\n" +
		"  <polygon points=\"200 120 194 117 194 123\" fill=\"#334155\"/>\n" +
		"  <text x=\"145\" y=\"80\" font-family=\"sans-serif\" font-size=\"12\" fill=\"#0f172a\">next</text>\n" +
		"  <rect x=\"10\" y=\"20\" width=\"80\" height=\"40\" fill=\"#ffffff\" stroke=\"#334155\" stroke-width=\"1\"/>\n" +
		"  <text x=\"50\" y=\"40\" font-family=\"sans-serif\" font-size=\"14\" text-anchor=\"middle\" fill=\"#0f172a\">B &amp; literal</text>\n" +
		"  <rect x=\"200\" y=\"100\" width=\"60\" height=\"40\" fill=\"#ffffff\" stroke=\"#334155\" stroke-width=\"1\"/>\n" +
		"  <text x=\"230\" y=\"120\" font-family=\"sans-serif\" font-size=\"14\" text-anchor=\"middle\" fill=\"#0f172a\">A</text>\n</svg>\n"
	if string(data) != want {
		t.Fatalf("manual SVG differs: %s", data)
	}
	if validateRenderedSVG(data) != nil {
		t.Fatal("manual output violates passive SVG profile")
	}
	before, _ := canonicalJSON(d)
	_, err = serializeManualDiagramSVG(d)
	after, _ := canonicalJSON(d)
	if err != nil || !bytes.Equal(before, after) {
		t.Fatal("layout mutated admitted input")
	}
	for _, example := range []struct{ toward, want diagramPoint }{
		{diagramPoint{200, 40}, diagramPoint{90, 40}}, {diagramPoint{0, 40}, diagramPoint{10, 40}},
		{diagramPoint{50, 200}, diagramPoint{50, 60}}, {diagramPoint{50, 0}, diagramPoint{50, 20}},
		{diagramPoint{70, 60}, diagramPoint{90, 40}}, {diagramPoint{30, 20}, diagramPoint{10, 40}},
	} {
		got, err := diagramEdgeEndpoint(d.CompositionLayout.NodePositions[1], example.toward)
		if err != nil || got != example.want {
			t.Fatal("endpoint side/tie rule", got, err)
		}
	}
	d.CompositionLayout.EdgeRoutes = []diagramEdgeRoute{{TargetRef: "opaque:edge", RouteKind: "polyline", Waypoints: []diagramPoint{{X: 50, Y: 150}, {X: 230, Y: 150}}}}
	data, err = serializeManualDiagramSVG(d)
	if err != nil || !bytes.Contains(data, []byte(`d="M50 60 L50 150 L230 150 L230 140"`)) {
		t.Fatal("authored route/endpoint order", err)
	}
	for name, test := range map[string]struct {
		change func(*reportingDiagramModel)
		reason string
	}{
		"missing position": {func(d *reportingDiagramModel) {
			d.CompositionLayout.NodePositions = d.CompositionLayout.NodePositions[:1]
		}, "diagram_layout_missing_node_position"},
		"duplicate position": {func(d *reportingDiagramModel) {
			d.CompositionLayout.NodePositions = append(d.CompositionLayout.NodePositions, d.CompositionLayout.NodePositions[0])
		}, "diagram_layout_duplicate_target"},
		"unknown position": {func(d *reportingDiagramModel) { d.CompositionLayout.NodePositions[0].TargetRef = "n0001" }, "diagram_layout_unknown_target"},
		"bounds":           {func(d *reportingDiagramModel) { d.CompositionLayout.NodePositions[0].Width = 401 }, "diagram_layout_invalid"},
		"negative":         {func(d *reportingDiagramModel) { d.CompositionLayout.NodePositions[0].X = -1 }, "diagram_layout_invalid"},
		"order": {func(d *reportingDiagramModel) {
			d.CompositionLayout.NodePositions[0], d.CompositionLayout.NodePositions[1] = d.CompositionLayout.NodePositions[1], d.CompositionLayout.NodePositions[0]
		}, "diagram_layout_invalid"},
		"unknown route": {func(d *reportingDiagramModel) {
			d.CompositionLayout.EdgeRoutes = []diagramEdgeRoute{{TargetRef: "n0001"}}
		}, "diagram_layout_unknown_target"},
		"zero vector": {func(d *reportingDiagramModel) {
			d.CompositionLayout.EdgeRoutes = []diagramEdgeRoute{{TargetRef: "opaque:edge", RouteKind: "polyline", Waypoints: []diagramPoint{{X: 50, Y: 40}}}}
		}, "diagram_layout_invalid"},
		"wrong kind": {func(d *reportingDiagramModel) { d.DiagramKind = "sequence" }, "diagram_layout_invalid"},
	} {
		t.Run(name, func(t *testing.T) {
			candidate := manualDiagramFixture()
			test.change(&candidate)
			data, err := serializeManualDiagramSVG(candidate)
			var invalid *renderValidationError
			if len(data) != 0 || !errors.As(err, &invalid) || invalid.ReasonCode != test.reason {
				t.Fatal("invalid layout accepted", err)
			}
		})
	}
	input := manualDiagramRenderInput(t)
	rendered, _, err := fixtureRenderer().Render(context.Background(), input)
	if err != nil {
		t.Fatal("manual candidate", err)
	}
	if err := validateRenderCandidateBinding(input, rendered); err != nil {
		t.Fatal("manual candidate binding", err)
	}
	files := map[string]RenderBundleFile{}
	for _, f := range rendered.RenderBundle.Files {
		files[f.Path] = f
	}
	if len(files["diagrams/manual.svg"].Bytes) == 0 || len(files["diagrams/manual.mmd"].Bytes) != 0 || !bytes.Contains(files["slides.md"].Bytes, []byte("![manual](diagrams/manual.svg)")) {
		t.Fatal("manual output membership")
	}
	if !bytes.Contains(files["slides.md"].Bytes, []byte("```mermaid\n")) || len(files["diagrams/automatic.mmd"].Bytes) == 0 {
		t.Fatal("mixed auto output membership")
	}
	var deck reportingDeckModel
	if json.Unmarshal(files["intermediate/deck-model.json"].Bytes, &deck) != nil || len(deck.Slides) != 3 || deck.Slides[1].Title != "manual" || deck.Slides[2].Title != "automatic" {
		t.Fatal("diagram insertion order changed")
	}
	for _, path := range []string{"diagrams/manual.svg", "diagrams/automatic.mmd", "intermediate/diagrams/manual.json"} {
		candidate := rendered
		candidate.RenderBundle.Files = append([]RenderBundleFile{}, rendered.RenderBundle.Files...)
		for i := range candidate.RenderBundle.Files {
			f := &candidate.RenderBundle.Files[i]
			if f.Path == path {
				f.Bytes = append(append([]byte{}, f.Bytes...), '\n')
			}
		}
		rehashPresentationFixture(t, &candidate)
		if validateRenderCandidateBinding(input, candidate) == nil {
			t.Fatal("coherent replacement admitted", path)
		}
	}
	input.OutputKind = OutputKindMermaid
	input.OutputOptions = fixtureOutputOptions(OutputKindMermaid, input.ReleaseScope)
	_, _, err = fixtureRenderer().Render(context.Background(), input)
	var invalid *renderValidationError
	if !errors.As(err, &invalid) || invalid.ReasonCode != "manual_layout_not_supported_for_output_kind" {
		t.Fatal("manual Mermaid accepted", err)
	}
}

func testCompositionDiagramOperations(t *testing.T) {
	contract, _ := ResolveTemplateContract(DefaultTemplateID, DefaultTemplateVersion)
	contract.Sections[0].DiagramDecls = []TemplateDiagramDeclaration{{DeclID: "original", DiagramKind: "flowchart", DiagramSourceKind: "timeline"}}
	replacement := "original"
	composition := &renderComposition{DiagramDecls: []compositionDiagramDecl{{DeclID: replacement, ReplacesDeclID: &replacement, DiagramKind: "sequence", DiagramSourceKind: "timeline", LayoutMode: "auto"}}}
	decls, owned, err := resolveDiagramDeclarations(contract, composition)
	if err != nil || len(decls) != 1 || !owned[replacement] || decls[0].DiagramKind != "sequence" {
		t.Fatal("replacement did not precede uniqueness check", err)
	}
	missing := "missing"
	for _, d := range []compositionDiagramDecl{{DeclID: "original"}, {DeclID: "missing", ReplacesDeclID: &missing}, {DeclID: "renamed", ReplacesDeclID: &replacement}} {
		if _, _, err := resolveDiagramDeclarations(contract, &renderComposition{DiagramDecls: []compositionDiagramDecl{d}}); err == nil {
			t.Fatal("invalid replacement/collision accepted")
		}
	}
	anchor := &compositionDiagramAnchor{AnchorKind: "diagram_anchor", DeclarationID: "manual"}
	d := manualDiagramFixture()
	override := diagramLabelOverride{Label: "First"}
	override.Target.Kind = "vertex"
	override.Target.Ref = d.IncludedVertexRefs[0]
	op := compositionDeckOp{OpKind: "override_diagram_labels", OnUnresolved: "fail", Payload: compositionOpPayload{DiagramAnchor: anchor, LabelOverrides: []diagramLabelOverride{override}}}
	later := op
	later.Payload.LabelOverrides = []diagramLabelOverride{override}
	later.Payload.LabelOverrides[0].Label = "Later"
	diagrams := []reportingDiagramModel{d}
	if err := applyCompositionDiagramOps(&diagrams, &renderComposition{DeckOps: []compositionDeckOp{op, later}}, ReleaseScopeInternalReview); err != nil || diagrams[0].Nodes[0].Label != "Later" {
		t.Fatal("later diagram label does not win", err)
	}
	exclude := compositionDeckOp{OpKind: "exclude_diagram", OnUnresolved: "fail", Payload: compositionOpPayload{DiagramAnchor: anchor}}
	for _, drop := range []bool{false, true} {
		diagrams = []reportingDiagramModel{manualDiagramFixture()}
		next := op
		if drop {
			next.OnUnresolved = "drop"
		}
		err := applyCompositionDiagramOps(&diagrams, &renderComposition{DeckOps: []compositionDeckOp{exclude, next}}, ReleaseScopeInternalReview)
		if drop && (err != nil || len(diagrams) != 0) || !drop && err == nil {
			t.Fatal("removed diagram target remained resolvable", err)
		}
	}
	d.compositionOwned = true
	diagrams = []reportingDiagramModel{d}
	if applyCompositionDiagramOps(&diagrams, &renderComposition{DeckOps: []compositionDeckOp{exclude}}, ReleaseScopeInternalReview) == nil {
		t.Fatal("template exclusion removed composition declaration")
	}
	slides := []reportingDeckSlide{{SlideID: "base", Title: "Base", Blocks: []ReportingBlock{}, sectionOrigin: &deckSectionOrigin{DeclarationID: "base"}}}
	insert := compositionDeckOp{OpKind: "insert_diagram_slide", Payload: compositionOpPayload{DiagramAnchor: anchor, SectionAnchor: &compositionSectionAnchor{AnchorKind: "section_anchor", DeclarationID: "base"}, Position: "before"}}
	if err := applyCompositionToDeck(&slides, &renderComposition{DeckOps: []compositionDeckOp{insert, insert}}, contract, ReleaseScopeInternalReview, diagrams...); err == nil {
		t.Fatal("duplicate diagram insertion admitted")
	}
}

func manualDiagramRenderInput(t *testing.T) renderInput {
	t.Helper()
	contract, _ := ResolveTemplateContract(DefaultTemplateID, DefaultTemplateVersion)
	model := fixtureStructuredRenderModel(t, []ExportField{{Path: "/timeline/event-a", SourceFamily: "timeline_event", SourceRecordID: "event-a", SourceRecordType: "timeline_event", ContentClass: ContentClassSourceEvidence, Value: map[string]any{"capture_state": "rough", "value": "Alpha"}}, {Path: "/timeline/event-b", SourceFamily: "timeline_event", SourceRecordID: "event-b", SourceRecordType: "timeline_event", ContentClass: ContentClassSourceEvidence, Value: map[string]any{"capture_state": "rough", "value": "Beta"}}})
	rule := diagramSelectionRule{SchemaID: "diagram_selection_rule.v1", RuleID: "timeline_sequence", VertexRefs: []string{}, EdgeRefs: []string{}, SeedVertexRefs: []string{}, EdgeKindFilter: []string{}, TimelineRecordIDs: []string{"event-a", "event-b"}, OverflowPolicy: "fail"}
	layout := manualDiagramFixture().CompositionLayout
	layout.NodePositions[0].TargetRef, layout.NodePositions[1].TargetRef = "event-a", "event-b"
	decl := compositionDiagramDecl{DeclID: "manual", DiagramKind: "flowchart", DiagramSourceKind: "timeline", SelectionRule: rule, LayoutMode: "manual", Layout: layout, LabelOverrides: []diagramLabelOverride{}}
	auto := decl
	auto.DeclID = "automatic"
	auto.LayoutMode = "auto"
	auto.Layout = nil
	anchor := &compositionSectionAnchor{AnchorKind: "section_anchor", DeclarationID: "snapshot"}
	ops := []compositionDeckOp{}
	for _, id := range []string{"manual", "automatic"} {
		ops = append(ops, compositionDeckOp{OpID: id, OpKind: "insert_diagram_slide", OnUnresolved: "fail", Payload: compositionOpPayload{SectionAnchor: anchor, DiagramAnchor: &compositionDiagramAnchor{AnchorKind: "diagram_anchor", DeclarationID: id}, Position: "after"}})
	}
	raw, err := json.Marshal(renderComposition{SchemaID: "cartulary.report_composition.v1", DeckOps: ops, DiagramDecls: []compositionDiagramDecl{decl, auto}, AuthoredTexts: []compositionAuthoredText{}})
	if err != nil {
		t.Fatal(err)
	}
	request := CreateReleaseRequest{OutputKind: OutputKindSlidev, OutputOptions: fixtureOutputOptions(OutputKindSlidev, ReleaseScopeInternalReview), ReleaseScope: ReleaseScopeInternalReview, RedactionProfileID: InternalRedactionProfileID, RedactionProfileVersion: DefaultRedactionProfileVersion, CompositionJSON: sealedReportingCompositionFixture(t, raw, model.IncidentID, contract)}
	encoded, _ := canonicalJSON(model)
	return renderCandidateFixture(request, contract, model, reportingObjectDigest(ExportModelSchemaID, encoded))
}

func testStageAssetAdmission(t *testing.T) {
	if templateAssetURL("assets/image/café !'()*+$&=.png") != "assets/image/caf%C3%A9%20%21%27%28%29%2A%2B%24%26%3D.png" {
		t.Fatal("asset URL does not preserve exact octets")
	}
	pack := fixtureTemplateAssetPack(t)
	image := renderStageAsset{Path: pack.Manifest.Assets[0].BundlePath, MediaType: "image/png", Bytes: pack.Files[0].Bytes, SHA256: pack.Manifest.Assets[0].SHA256}
	if validateRenderStageAssets(renderStageRequest{Operation: "slidev_pdf", Assets: []renderStageAsset{image}}) != nil {
		t.Fatal("binary PNG transport rejected")
	}
	svg := `<svg xmlns="http://www.w3.org/2000/svg"><text>é</text></svg>`
	asset := renderStageAsset{Path: "diagrams/a.svg", MediaType: "image/svg+xml", Bytes: []byte(svg), SHA256: hashHex([]byte(svg))}
	request := renderStageRequest{SchemaID: renderStageRequestSchema, Operation: "slidev_pdf", Assets: []renderStageAsset{asset}}
	if validateRenderStageAssets(request) != nil {
		t.Fatal("valid local asset rejected")
	}
	if renderStageAssetsDigest(request.Assets) != hashHex([]byte("cartulary.reporting_render_assets.v2\x0014:diagrams/a.svg13:image/svg+xml"+"61:"+svg)) {
		t.Fatal("asset framing differs")
	}
	for _, path := range []string{"../a.svg", "diagrams/../a.svg", "/diagrams/a.svg", "diagrams/a.svg?x", `diagrams\a.svg`, "diagrams/../../secret"} {
		bad := request
		bad.Assets = []renderStageAsset{{Path: path, MediaType: "image/svg+xml", Bytes: []byte(svg), SHA256: asset.SHA256}}
		if validateRenderStageAssets(bad) == nil {
			t.Fatal("asset escaped fixed directory")
		}
	}
	for _, content := range []string{`<svg xmlns="http://www.w3.org/2000/svg"><script>bad</script></svg>`, `<svg xmlns="http://www.w3.org/2000/svg"><image href="https://invalid.test"/></svg>`} {
		bad := request
		bad.Assets = []renderStageAsset{{Path: asset.Path, MediaType: "image/svg+xml", Bytes: []byte(content), SHA256: hashHex([]byte(content))}}
		if validateRenderStageAssets(bad) == nil {
			t.Fatal("active SVG reached stage")
		}
	}
	bad := request
	bad.Assets = []renderStageAsset{asset, asset}
	if validateRenderStageAssets(bad) == nil {
		t.Fatal("duplicate asset accepted")
	}
	bad = request
	bad.Operation = "mermaid_svg"
	if validateRenderStageAssets(bad) == nil {
		t.Fatal("Mermaid accepted local assets")
	}
	bad = request
	bad.Assets = []renderStageAsset{{Path: asset.Path, MediaType: "image/svg+xml", Bytes: []byte(svg), SHA256: strings.Repeat("0", 64)}}
	if validateRenderStageAssets(bad) == nil {
		t.Fatal("asset digest not enforced")
	}
}
