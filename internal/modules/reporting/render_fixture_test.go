package reporting

import (
	"context"
	"encoding/json"
	"html"
	"strings"

	"github.com/JochiRaider/cartulary/internal/platform/renderprocess"
)

// This is a bundle-assembly fixture, not rendering or sandbox evidence. Real
// bytes, isolation and rerender equivalence belong to the installed-stage test.
func fixtureRenderer() *Renderer { return &Renderer{runtime: fixtureRenderRuntime{}} }

type fixtureRenderRuntime struct{}

type alteredRenderRuntime struct {
	fixtureRenderRuntime
	change func(*renderprocess.Result)
	runErr error
}

func (r alteredRenderRuntime) Run(ctx context.Context, input []byte, outputs []renderprocess.Output) (renderprocess.Result, error) {
	if r.runErr != nil {
		return renderprocess.Result{}, r.runErr
	}
	result, err := r.fixtureRenderRuntime.Run(ctx, input, outputs)
	if err == nil && r.change != nil {
		r.change(&result)
	}
	return result, err
}

func (fixtureRenderRuntime) Verify() error { return nil }
func (fixtureRenderRuntime) Toolchain() (ToolchainSnapshot, error) {
	version, executable, digest := "synthetic-unit-fixture", "bin/synthetic-browser", hashHex([]byte("synthetic-toolchain-bytes"))
	fonts := []ToolchainFontFile{{Path: "installed/fixture.font", SHA256: digest}}
	fontHash := hashHex([]byte(`[{"path":"installed/fixture.font","sha256":"` + digest + `"}]`))
	return ToolchainSnapshot{
		SchemaID: "cartulary.reporting_toolchain_snapshot.v2", NodeVersion: version, PackageManager: version, OSImageID: version,
		LockfileSHA256: digest, PackageStoreDigest: digest, FontManifestSHA256: fontHash, FontFiles: fonts,
		SlidevVersion: &version, MermaidVersion: &version, ChromiumVersion: &version, ChromiumExecutablePath: &executable, ChromiumExecutableSHA256: &digest,
		Locale: "C.UTF-8", Timezone: "UTC", ViewportCSSPX: "1280x720", DeviceScaleFactor: 1, ColorScheme: "light",
		NetworkPolicyID: "render_sandbox_policy_v1", BrowserLaunchArgs: []string{}, EnvAllowlist: []ToolchainEnvironmentValue{}, RenderCommand: []string{"synthetic-renderer"},
	}, nil
}
func (fixtureRenderRuntime) Run(ctx context.Context, input []byte, _ []renderprocess.Output) (renderprocess.Result, error) {
	if err := ctx.Err(); err != nil {
		return renderprocess.Result{}, err
	}
	var request renderStageRequest
	if err := json.Unmarshal(input, &request); err != nil {
		return renderprocess.Result{}, err
	}
	path, media, stage := "diagram.svg", "image/svg+xml", "mermaid_render"
	data := []byte(`<svg xmlns="http://www.w3.org/2000/svg"><text>` + html.EscapeString(strings.Join(strings.Fields(request.Source), " ")) + `</text></svg>`)
	if request.Operation == "slidev_pdf" {
		path, media, stage = "deck.pdf", "application/pdf", "slidev_export"
		data = []byte("%PDF-1.7\n% synthetic unit fixture " + hashHex([]byte(request.Source)) + "\n%%EOF\n")
	}
	result := renderStageResult{SchemaID: renderStageResultSchema, Operation: request.Operation, IdentitySHA256: request.IdentitySHA256, SourceSHA256: hashHex([]byte(request.Source)), AssetsSHA256: renderStageAssetsDigest(request.Assets), Artifact: renderStageArtifact{Path: path, MediaType: media, Bytes: int64(len(data)), SHA256: hashHex(data)}, BrowserObservations: []renderStageObservation{{SchemaID: "render_sandbox_observation.v1", NetworkPolicyID: "render_sandbox_policy_v1", Stage: stage, ObservationKind: "loopback_allowed", DestinationClass: "loopback", Allowed: true, AttemptCount: 1, LoopbackPortRef: "render-loopback-1"}}}
	protocol, err := json.Marshal(result)
	return renderprocess.Result{Protocol: protocol, Files: map[string][]byte{path: data}, NetworkObservations: []renderprocess.NetworkObservation{{Kind: "loopback_allowed", DestinationClass: "loopback", Allowed: true, Count: 1}}}, err
}
