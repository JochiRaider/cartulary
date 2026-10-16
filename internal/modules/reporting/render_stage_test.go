package reporting

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"errors"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/platform/renderprocess"
)

func TestReportingInstalledSandboxRenders(t *testing.T) {
	directory, err := filepath.Abs("../../../build/report-renderer")
	if err != nil {
		t.Fatal(err)
	}
	installation, err := renderprocess.OpenInstallation(directory)
	if err != nil {
		t.Fatalf("required build-report-renderer installation: %v", err)
	}
	temporaryRoot := t.TempDir()
	for _, operation := range []string{"mermaid_svg", "slidev_pdf", "slidev_list_clicks"} {
		t.Run(operation, func(t *testing.T) {
			request := renderStageRequest{SchemaID: renderStageRequestSchema, Operation: operation, Source: "flowchart TD\n  n0001[\"First\"]\n  n0002[\"Second\"]\n  n0001 -->|\"Edge\"| n0002\n", AdmittedAt: "2026-04-05T12:00:00Z", IdentitySHA256: strings.Repeat("a", 64)}
			output := "diagram.svg"
			if operation != "mermaid_svg" {
				request.Operation = "slidev_pdf"
				data, err := os.ReadFile("../../../contracts/reporting/fixtures/slidev-source.v2.json")
				if err != nil {
					t.Fatal(err)
				}
				var fixture struct {
					Deck     reportingDeckModel `json:"deck"`
					Template TemplateContract   `json:"template"`
				}
				if err := json.Unmarshal(data, &fixture); err != nil {
					t.Fatal(err)
				}
				if operation == "slidev_list_clicks" {
					fixture.Deck.Slides = []reportingDeckSlide{{SlideID: "sld-0001", Title: "List reveal", Layout: "default", Blocks: []ReportingBlock{{BlockID: "list", BlockKind: "bullet_list", Children: []ReportingBlock{
						{BlockID: "item-a", BlockKind: "paragraph", Fields: []ReportingField{{Value: "First"}}},
						{BlockID: "item-b", BlockKind: "paragraph", Fields: []ReportingField{{Value: "Second"}}},
						{BlockID: "item-c", BlockKind: "paragraph", Fields: []ReportingField{{Value: "Third"}}},
					}}}, clickProfile: "reveal_list_items"}}
					limits, err := resolveRenderLimits(nil)
					if err != nil {
						t.Fatal(err)
					}
					if err := materializeSlideClicks(&fixture.Deck.Slides[0], limits); err != nil {
						t.Fatal(err)
					}
				}
				source, err := serializeSlidevMarkdown(fixture.Deck, fixture.Template)
				if err != nil {
					t.Fatal(err)
				}
				pages := len(fixture.Deck.Slides)
				for _, slide := range fixture.Deck.Slides {
					pages += len(slide.ClickSteps)
				}
				request.Source, request.ExpectedPages, output = string(source), &pages, "deck.pdf"
			}
			input, err := json.Marshal(request)
			if err != nil {
				t.Fatal(err)
			}
			var first []byte
			for attempt := 0; attempt < 2; attempt++ {
				ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
				runtime := installation.Runtime(temporaryRoot, 32123)
				result, err := runtime.Run(ctx, input, []renderprocess.Output{{Path: output, MaxBytes: 32 << 20}})
				cancel()
				if err != nil {
					t.Fatalf("installed native stage: %v", err)
				}
				if _, err := admitRenderStageResult(request, result.Protocol, result.Files); err != nil {
					t.Fatal(err)
				}
				if len(result.NetworkObservations) != 1 || result.NetworkObservations[0].Kind != "loopback_allowed" || !result.NetworkObservations[0].Allowed || result.NetworkObservations[0].Count < 2 {
					t.Fatalf("native traffic not observed: %#v", result.NetworkObservations)
				}
				if attempt == 0 {
					first = result.Files[output]
				} else if !bytes.Equal(first, result.Files[output]) {
					t.Fatal("fresh native attempts produced different rendered bytes")
				}
				entries, err := os.ReadDir(temporaryRoot)
				if err != nil || len(entries) != 0 {
					t.Fatal("prior attempt survived into rerender")
				}
			}
		})
	}
	if err := installation.Verify(); err != nil {
		t.Fatal("renderer changed its installation")
	}
	t.Run("candidate_pipeline", func(t *testing.T) {
		renderer, err := NewInstalledRenderer(directory, temporaryRoot)
		if err != nil {
			t.Fatal(err)
		}
		contract, _ := ResolveTemplateContract(DefaultTemplateID, DefaultTemplateVersion)
		model := fixtureStructuredRenderModel(t, []ExportField{
			{Path: "/incident/status", ContentClass: ContentClassDerivedAnalytic, Value: "open"},
			{Path: "/incident/title", ContentClass: ContentClassCuratedNarrative, Value: "Native candidate", SupportRefs: []string{"/incident/status"}},
			{Path: "/timeline/native-event", SourceFamily: "timeline_event", SourceRecordID: "native-event", SourceRecordType: "timeline_event", ContentClass: ContentClassSourceEvidence, Value: map[string]any{"capture_state": "rough", "value": "Observed"}},
		})
		encoded, err := canonicalJSON(model)
		if err != nil {
			t.Fatal(err)
		}
		for _, kind := range []string{OutputKindSlidev, OutputKindMermaid} {
			request := CreateReleaseRequest{OutputKind: kind, OutputOptions: fixtureOutputOptions(kind, ReleaseScopeInternalReview), ReleaseScope: ReleaseScopeInternalReview, RedactionProfileID: InternalRedactionProfileID, RedactionProfileVersion: DefaultRedactionProfileVersion}
			if kind == OutputKindMermaid {
				request.CompositionJSON = fixtureTimelineComposition(t, contract, model.IncidentID, "flowchart", []string{"native-event"})
			} else {
				request.CompositionJSON = sealedReportingCompositionFixture(t, []byte(`{"schema_id":"cartulary.report_composition.v1","authored_texts":[],"diagram_decls":[],"deck_ops":[{"op_id":"clicks","op_kind":"override_click_profile","on_unresolved":"fail","payload":{"section_anchor":{"anchor_kind":"section_anchor","template_section_decl_id":"snapshot","expansion_key":null},"click_profile":"reveal_blocks"}}]}`), model.IncidentID, contract)
			}
			input := renderCandidateFixture(request, contract, model, reportingObjectDigest(ExportModelSchemaID, encoded))
			rendered, reason, err := renderer.Render(context.Background(), input)
			if err != nil {
				t.Fatalf("installed candidate %s: %s: %v (supervisor: %v)", kind, reason, err, errors.Unwrap(err))
			}
			found := false
			for _, file := range rendered.RenderBundle.Files {
				if file.Role == renderBundleRoleRenderedPDF || file.Role == renderBundleRoleRenderedSVG {
					found = true
					if len(file.Bytes) < 1000 || bytes.Contains(file.Bytes, []byte("synthetic")) {
						t.Fatal("candidate emitted placeholder output")
					}
				}
				if file.Role == renderBundleRoleToolchainSnapshot {
					var snapshot ToolchainSnapshot
					if json.Unmarshal(file.Bytes, &snapshot) != nil || snapshot.NodeVersion != installation.Manifest().NodeVersion || snapshot.OSImageSHA256 == nil || *snapshot.OSImageSHA256 != installation.Digest() || len(snapshot.BrowserLaunchArgs) == 0 || len(snapshot.EnvAllowlist) == 0 {
						t.Fatal("candidate toolchain is not bound to its installation")
					}
				}
				if file.Path == "validation/sandbox-observation.json" {
					t.Fatal("attempt observations entered approval content")
				}
			}
			if !found {
				t.Fatal("candidate omitted requested rendering")
			}
		}
	})
	t.Run("declared local assets", func(t *testing.T) {
		renderer, err := NewInstalledRenderer(directory, temporaryRoot)
		if err != nil {
			t.Fatal(err)
		}
		for _, scope := range []string{ReleaseScopeInternalReview, ReleaseScopeExternal} {
			input := fixtureTemplateAssetInputForScope(t, scope)
			rendered, reason, err := renderer.Render(context.Background(), input)
			if err != nil {
				t.Fatalf("native declared assets in %s: %s: %v", scope, reason, err)
			}
			assertTemplateAssetOutput(t, input, rendered)
			if scope == ReleaseScopeExternal && rendered.RenderBundle.Manifest.ExternalDeterminismSHA256 == nil {
				t.Fatal("external native images omitted fresh-attempt determinism evidence")
			}
		}
	})
	t.Run("declared local themes", func(t *testing.T) { assertNativeTemplateThemes(t, directory, temporaryRoot) })
	t.Run("external Mermaid bundles", func(t *testing.T) { testNativeExternalMermaid(t, directory, temporaryRoot) })
	t.Run("authored presentation uses retained redaction", func(t *testing.T) {
		renderer, err := NewInstalledRenderer(directory, temporaryRoot)
		if err != nil {
			t.Fatal(err)
		}
		input := compositionTextFixture(t, []compositionAuthoredText{{AuthoredTextID: "title", TextRole: "title_override", Body: "Review {{subject:subject:host:record:host-one}}", DisclosurePartitionRef: "public"}}, []compositionDeckOp{{OpID: "title", OpKind: "override_title", OnUnresolved: "fail", Payload: compositionOpPayload{AuthoredTextRef: "title", SectionAnchor: &compositionSectionAnchor{AnchorKind: "section_anchor", DeclarationID: "snapshot"}}}})
		profile := input.RedactionProfile
		class := ContentClassSourceEvidence
		profile.Rules = []RedactionRule{{RuleID: "source-token", ContentClass: &class, Action: RedactionActionSpec{Type: ActionMask, TokenBacked: true}}}
		setCompositionFixtureProfile(t, &input, profile)
		rendered, reason, err := renderer.Render(context.Background(), input)
		if err != nil {
			t.Fatalf("native authored candidate: %s: %v", reason, err)
		}
		if err := validateRenderCandidateBinding(input, rendered); err != nil {
			t.Fatal(err)
		}
		pdf, source := false, false
		for _, file := range rendered.RenderBundle.Files {
			if file.Role == renderBundleRoleRenderedPDF {
				pdf = len(file.Bytes) > 1000
			}
			if file.Path == "slides.md" {
				source = bytes.Contains(file.Bytes, []byte("Review HOST-0001"))
			}
			if bytes.Contains(file.Bytes, []byte("Sensitive host")) {
				t.Fatal("raw subject entered shareable output")
			}
		}
		if !pdf || !source {
			t.Fatal("native authored output missing")
		}
	})
	t.Run("mixed manual and auto diagram candidate", func(t *testing.T) {
		renderer, err := NewInstalledRenderer(directory, temporaryRoot)
		if err != nil {
			t.Fatal(err)
		}
		input := manualDiagramRenderInput(t)
		var previous string
		previousFiles := map[string]string{}
		previousBytes := map[string][]byte{}
		// Each comparison uses fresh native stages and clean working directories.
		for i := 0; i < 2; i++ {
			rendered, reason, err := renderer.Render(context.Background(), input)
			if err != nil {
				t.Fatalf("native diagram candidate: %s: %v (supervisor: %v)", reason, err, errors.Unwrap(err))
			}
			if err := validateRenderCandidateBinding(input, rendered); err != nil {
				t.Fatal("native diagram binding", err)
			}
			if i > 0 && previous != rendered.OutputSHA256 {
				for _, file := range rendered.RenderBundle.Files {
					if file.Path == "slides.md" {
						t.Logf("synthetic source: %s", file.Bytes)
					}
				}
				for _, file := range rendered.RenderBundle.Files {
					if before := previousFiles[file.Path]; before != hashHex(file.Bytes) {
						t.Logf("changed artifact %s: %s -> %s (%d bytes)", file.Path, before, hashHex(file.Bytes), len(file.Bytes))
						// Only this fixed synthetic fixture emits bounded byte
						// diagnostics; production never logs rendered content.
						old := previousBytes[file.Path]
						for offset := 0; offset < len(old) && offset < len(file.Bytes); offset++ {
							if old[offset] != file.Bytes[offset] {
								start := max(0, offset-80)
								t.Logf("first differing byte %d; prior=%q current=%q", offset, old[start:min(len(old), offset+160)], file.Bytes[start:min(len(file.Bytes), offset+160)])
								break
							}
						}
					}
				}
				t.Fatal("mixed diagrams changed bytes between fresh attempts")
			}
			previous = rendered.OutputSHA256
			for _, file := range rendered.RenderBundle.Files {
				previousFiles[file.Path] = hashHex(file.Bytes)
				previousBytes[file.Path] = append([]byte{}, file.Bytes...)
			}
			entries, err := os.ReadDir(temporaryRoot)
			if err != nil || len(entries) != 0 {
				t.Fatal("diagram assets survived attempt cleanup")
			}
		}
	})
	entries, err := os.ReadDir(temporaryRoot)
	if err != nil || len(entries) != 0 {
		t.Fatal("renderer attempt cleanup incomplete")
	}
}

func testNativeExternalMermaid(t *testing.T, directory, temporaryRoot string) {
	renderer, err := NewInstalledRenderer(directory, temporaryRoot)
	if err != nil {
		t.Fatal(err)
	}
	contract, _ := ResolveTemplateContract(DefaultTemplateID, DefaultTemplateVersion)
	model, digest, err := buildStructuredTestExportModel("00000000-0000-0000-0000-000000000001", fixtureSnapshotID,
		time.Date(2026, 4, 5, 12, 0, 0, 0, time.UTC), SourceBoundaryTokenPrefix+strings.Repeat("0", 64), ReleaseScopeExternal, nil,
		[]ExportField{
			{Path: "/incident/title", ContentClass: ContentClassDerivedAnalytic, Value: "Public timeline", DisclosurePartitionRefs: []string{"public"}},
			{Path: "/timeline/native-event", SourceFamily: "timeline_event", SourceRecordID: "native-event", SourceRecordType: "timeline_event", ContentClass: ContentClassSourceEvidence, Value: map[string]any{"capture_state": "rough", "value": "Observed"}, DisclosurePartitionRefs: []string{"public"}},
		})
	if err != nil {
		t.Fatal(err)
	}
	for _, count := range []int{1, 3} {
		t.Run(strconv.Itoa(count), func(t *testing.T) {
			declarations := []any{}
			for i := 1; i <= count; i++ {
				decl := fixtureTimelineDeclaration("flowchart", []string{"native-event"})
				declarations = append(declarations, map[string]any{
					"decl_id": "timeline-" + strconv.Itoa(i), "replaces_decl_id": nil, "diagram_kind": "flowchart",
					"diagram_source_kind": "timeline", "source_graph_view_id": nil, "selection_rule": decl.SelectionRule,
					"layout_mode": "auto", "layout": nil, "label_overrides": []any{},
				})
			}
			raw, err := json.Marshal(map[string]any{"schema_id": "cartulary.report_composition.v1", "deck_ops": []any{}, "authored_texts": []any{}, "diagram_decls": declarations})
			if err != nil {
				t.Fatal(err)
			}
			request := CreateReleaseRequest{OutputKind: OutputKindMermaid, OutputOptions: fixtureOutputOptions(OutputKindMermaid, ReleaseScopeExternal), ReleaseScope: ReleaseScopeExternal, RedactionProfileID: ExternalRedactionProfileID, RedactionProfileVersion: DefaultRedactionProfileVersion,
				CompositionJSON: sealedReportingCompositionFixture(t, raw, model.IncidentID, contract)}
			input := renderCandidateFixture(request, contract, model, digest)
			rendered, reason, err := renderer.Render(context.Background(), input)
			if err != nil {
				t.Fatalf("external Mermaid: %s: %v", reason, err)
			}
			if err := validateRenderCandidateBinding(input, rendered); err != nil {
				t.Fatal(err)
			}
			sources, images := []string{}, []string{}
			for _, file := range rendered.RenderBundle.Files {
				switch file.Role {
				case renderBundleRoleSourceMermaid:
					sources = append(sources, file.Path)
				case renderBundleRoleRenderedSVG:
					images = append(images, strings.TrimSuffix(file.Path, ".svg")+".mmd")
					if len(file.Bytes) < 1000 || validateRenderedSVG(file.Bytes) != nil {
						t.Fatal("missing actual passive SVG output")
					}
				}
			}
			if len(sources) != count || len(images) != count || strings.Join(sources, "\n") != strings.Join(images, "\n") || rendered.RenderBundle.Manifest.ExternalDeterminismSHA256 == nil {
				t.Fatal("external bundle lost diagram membership or fresh determinism evidence")
			}
		})
	}
}

func TestReportingRenderStageAdmitsRealSVG(t *testing.T) {
	// This is cross-language adapter evidence. The native namespace boundary is
	// covered separately; this test must not be reported as packaged-sandbox proof.
	repository, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	node, err := exec.LookPath("node")
	if err != nil {
		t.Fatal(err)
	}
	sources := []string{
		"flowchart TD\n  n0001[\"First\"]\n  n0002[\"Second\"]\n  n0001 -->|\"Edge\"| n0002\n",
		"flowchart LR\n  n0001[\"First\"]\n  n0002[\"Second\"]\n  n0003[\"Third\"]\n  n0001 -->|\"Edge A\"| n0002\n  n0001 -->|\"Edge B\"| n0003\n  n0002 -->|\"Edge C\"| n0003\n",
		"sequenceDiagram\n  participant p0001 as First\n  participant p0002 as Second\n  p0001->>p0002: Message\n",
	}
	data, err := os.ReadFile("../../../contracts/reporting/fixtures/mermaid-source.v3.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Cases []struct {
			Diagram     reportingDiagramModel `json:"diagram"`
			SourceLines []string              `json:"source_lines"`
		} `json:"cases"`
	}
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	for _, vector := range fixture.Cases {
		source, err := serializeMermaidSource(vector.Diagram)
		if err != nil || string(source) != strings.Join(vector.SourceLines, "\n") {
			t.Fatalf("canonical source diverged: %v", err)
		}
		sources = append(sources, string(source))
	}
	for _, source := range sources {
		request := renderStageRequest{SchemaID: renderStageRequestSchema, Operation: "mermaid_svg", Source: source, AdmittedAt: "2026-04-05T12:00:00Z", IdentitySHA256: strings.Repeat("a", 64)}
		data := exerciseReportingStage(t, repository, node, request, "diagram.svg")
		if !bytes.Contains(data, []byte("First")) || !bytes.Contains(data, []byte("Second")) {
			t.Fatal("stage lost admitted content")
		}
	}
}

func TestReportingRenderStageAdmitsRealPDF(t *testing.T) {
	repository, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	node, err := exec.LookPath("node")
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile("../../../contracts/reporting/fixtures/slidev-source.v2.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Deck        reportingDeckModel `json:"deck"`
		Template    TemplateContract   `json:"template"`
		SourceLines []string           `json:"source_lines"`
	}
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	source, err := serializeSlidevMarkdown(fixture.Deck, fixture.Template)
	if err != nil || string(source) != strings.Join(fixture.SourceLines, "\n") {
		t.Fatalf("canonical source diverged: %v", err)
	}
	pages := len(fixture.Deck.Slides)
	request := renderStageRequest{SchemaID: renderStageRequestSchema, Operation: "slidev_pdf", Source: string(source), AdmittedAt: "2026-04-05T12:00:00Z", IdentitySHA256: strings.Repeat("a", 64), ExpectedPages: &pages}
	first := exerciseReportingStage(t, repository, node, request, "deck.pdf")
	second := exerciseReportingStage(t, repository, node, request, "deck.pdf")
	if !bytes.Equal(first, second) {
		t.Fatal("isolated stage renders changed immutable admitted PDF bytes")
	}
}

func exerciseReportingStage(t *testing.T, repository, node string, request renderStageRequest, outputPath string) []byte {
	t.Helper()
	input, err := json.Marshal(request)
	if err != nil {
		t.Fatal(err)
	}
	work := t.TempDir()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	command := exec.CommandContext(ctx, node, filepath.Join(repository, "packages/report-renderer/src/stage-main.ts"))
	listener, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	if err := listener.Close(); err != nil {
		t.Fatal(err)
	}
	command.Env = append(os.Environ(), "CARTULARY_RENDER_LOOPBACK_PORT="+strconv.Itoa(port))
	command.Dir = work
	command.Stdin = bytes.NewReader(input)
	command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	command.Cancel = func() error { return syscall.Kill(-command.Process.Pid, syscall.SIGKILL) }
	command.WaitDelay = 2 * time.Second
	var output, diagnostic bytes.Buffer
	command.Stdout, command.Stderr = &output, &diagnostic
	err = command.Run()
	cancel()
	if err != nil {
		t.Fatalf("synthetic renderer stage: %v; %s", err, diagnostic.String())
	}
	data, err := os.ReadFile(filepath.Join(work, outputPath))
	if err != nil {
		t.Fatal(err)
	}
	result, err := admitRenderStageResult(request, output.Bytes(), map[string][]byte{outputPath: data})
	if err != nil {
		decoder := xml.NewDecoder(bytes.NewReader(data))
		for {
			token, decodeErr := decoder.Token()
			if decodeErr != nil {
				break
			}
			if element, ok := token.(xml.StartElement); ok {
				for _, attr := range element.Attr {
					if !svgAttributeAllowed(element.Name.Local, attr) {
						t.Logf("synthetic SVG rejected attribute: %s %s=%q", element.Name.Local, attr.Name.Local, attr.Value)
					}
				}
			}
		}
		t.Fatalf("actual renderer output failed independent admission: %v", err)
	}
	if result.Artifact.SHA256 != hashHex(data) {
		t.Fatal("stage lost byte binding")
	}
	return data
}

func TestReportingRenderStageRejectsUnboundOutputs(t *testing.T) {
	t.Run("local diagram assets", testStageAssetAdmission)
	request := renderStageRequest{SchemaID: renderStageRequestSchema, Operation: "mermaid_svg", Source: "synthetic source", AdmittedAt: "2026-04-05T12:00:00Z", IdentitySHA256: strings.Repeat("a", 64)}
	artifact := []byte(`<svg xmlns="http://www.w3.org/2000/svg"><text>Safe</text></svg>`)
	result := renderStageResult{SchemaID: renderStageResultSchema, Operation: request.Operation, IdentitySHA256: request.IdentitySHA256, SourceSHA256: hashHex([]byte(request.Source)), AssetsSHA256: renderStageAssetsDigest(request.Assets), Artifact: renderStageArtifact{Path: "diagram.svg", MediaType: "image/svg+xml", Bytes: int64(len(artifact)), SHA256: hashHex(artifact)}, BrowserObservations: []renderStageObservation{{SchemaID: "render_sandbox_observation.v1", NetworkPolicyID: "render_sandbox_policy_v1", Stage: "mermaid_render", ObservationKind: "loopback_allowed", DestinationClass: "loopback", Allowed: true, AttemptCount: 2, LoopbackPortRef: "render-loopback-1"}}}
	encode := func(value any) []byte {
		t.Helper()
		data, err := json.Marshal(value)
		if err != nil {
			t.Fatal(err)
		}
		return data
	}
	protocol := encode(result)
	files := map[string][]byte{"diagram.svg": artifact}
	if _, err := admitRenderStageResult(request, protocol, files); err != nil {
		t.Fatal(err)
	}
	for _, change := range []func(*renderStageResult){
		func(r *renderStageResult) { r.SchemaID = "cartulary.reporting_render_stage_result.v1" },
		func(r *renderStageResult) { r.Operation = "slidev_pdf" },
		func(r *renderStageResult) { r.IdentitySHA256 = strings.Repeat("b", 64) },
		func(r *renderStageResult) { r.SourceSHA256 = strings.Repeat("b", 64) },
		func(r *renderStageResult) { r.AssetsSHA256 = strings.Repeat("b", 64) },
		func(r *renderStageResult) { r.Artifact.Path = "../diagram.svg" },
		func(r *renderStageResult) { r.Artifact.MediaType = "application/pdf" },
		func(r *renderStageResult) { r.Artifact.Bytes++ },
		func(r *renderStageResult) { r.Artifact.SHA256 = strings.Repeat("b", 64) },
		func(r *renderStageResult) { r.BrowserObservations = nil },
		func(r *renderStageResult) {
			r.BrowserObservations = append(r.BrowserObservations, r.BrowserObservations[0])
		},
		func(r *renderStageResult) { r.BrowserObservations[0].AttemptCount = 0 },
		func(r *renderStageResult) { r.BrowserObservations[0].Allowed = false },
		func(r *renderStageResult) { r.BrowserObservations[0].LoopbackPortRef = "unallocated" },
		func(r *renderStageResult) { r.BrowserObservations[0].Stage = "slidev_export" },
		func(r *renderStageResult) { r.BrowserObservations[0].NetworkPolicyID = "network_none" },
	} {
		var candidate renderStageResult
		if err := json.Unmarshal(protocol, &candidate); err != nil {
			t.Fatal(err)
		}
		change(&candidate)
		if _, err := admitRenderStageResult(request, encode(candidate), files); !errors.Is(err, errRenderStageProtocol) {
			t.Fatalf("admitted altered result: %v", err)
		}
	}
	for _, candidate := range [][]byte{
		append(append([]byte(nil), protocol...), []byte(`{}`)...),
		bytes.Replace(protocol, []byte(`"artifact":{`), []byte(`"artifact":{"path":"diagram.svg",`), 1),
		bytes.Replace(protocol, []byte(`"artifact":{`), []byte(`"artifact":{"unknown":true,`), 1),
		bytes.Replace(protocol, []byte(`"allowed":true`), []byte(`"allowed":null`), 1),
		bytes.Replace(protocol, []byte(`"attempt_count":2`), []byte(`"attempt_count":2e0`), 1),
	} {
		if _, err := admitRenderStageResult(request, candidate, files); !errors.Is(err, errRenderStageProtocol) {
			t.Fatalf("admitted malformed protocol: %v", err)
		}
	}
	for _, candidate := range []map[string][]byte{nil, {"diagram.svg": []byte("changed")}, {"diagram.svg": artifact, "undeclared": artifact}} {
		if _, err := admitRenderStageResult(request, protocol, candidate); !errors.Is(err, errRenderStageProtocol) {
			t.Fatalf("admitted unbound artifacts: %v", err)
		}
	}
	unsafe := []byte(`<svg xmlns="http://www.w3.org/2000/svg"><script>void 0</script></svg>`)
	result.Artifact.Bytes, result.Artifact.SHA256 = int64(len(unsafe)), hashHex(unsafe)
	var validation *renderValidationError
	if _, err := admitRenderStageResult(request, encode(result), map[string][]byte{"diagram.svg": unsafe}); !errors.As(err, &validation) || validation.ReasonCode != "external_reference_in_rendered_output" {
		t.Fatalf("admitted unsafe digest-bound SVG: %v", err)
	}
	t.Run("native observations are independently mandatory", func(t *testing.T) {
		for _, observations := range [][]renderprocess.NetworkObservation{
			nil,
			{{Kind: "loopback_allowed", DestinationClass: "loopback", Allowed: true, Count: 0}},
			{{Kind: "loopback_allowed", DestinationClass: "external", Allowed: true, Count: 1}},
			{{Kind: "loopback_allowed", DestinationClass: "loopback", Allowed: false, Count: 1}},
			{{Kind: "loopback_allowed", DestinationClass: "loopback", Allowed: true, Count: 1}, {Kind: "loopback_allowed", DestinationClass: "loopback", Allowed: true, Count: 1}},
		} {
			renderer := &Renderer{runtime: alteredRenderRuntime{change: func(result *renderprocess.Result) { result.NetworkObservations = observations }}}
			limits, err := resolveRenderLimits(nil)
			if err != nil {
				t.Fatal(err)
			}
			bytes, err := renderer.renderStage(context.Background(), request, limits)
			var failure *renderValidationError
			if len(bytes) != 0 || !errors.As(err, &failure) || failure.ReasonCode != "sandbox_observation_missing" {
				t.Fatalf("native evidence admitted: %v", err)
			}
		}
	})
	t.Run("supervisor failures return no publishable output", func(t *testing.T) {
		for _, test := range []struct {
			err             error
			failure, reason string
		}{
			{context.Canceled, "render_canceled", "render_canceled"},
			{context.DeadlineExceeded, "render_timeout", "render_timeout"},
			{errors.Join(context.Canceled, renderprocess.ErrCleanup), "security_sandbox_violation", "partial_output_cleanup_failed"},
			{&renderprocess.NetworkError{Kind: "loopback_blocked"}, "security_sandbox_violation", "loopback_not_declared"},
			{&renderprocess.NetworkError{Kind: "outbound_blocked"}, "remote_asset_reference", "outbound_request_observed"},
			{renderprocess.ErrStart, "security_sandbox_violation", "render_runtime_unavailable"},
		} {
			renderer := &Renderer{runtime: alteredRenderRuntime{runErr: test.err}}
			limits, err := resolveRenderLimits(nil)
			if err != nil {
				t.Fatal(err)
			}
			bytes, err := renderer.renderStage(context.Background(), request, limits)
			var failure *renderValidationError
			if len(bytes) != 0 || !errors.As(err, &failure) || failure.FailureCode != test.failure || failure.ReasonCode != test.reason {
				t.Fatalf("classified output: %v", err)
			}
		}
	})
}
