// Package reportingfixture supplies explicit protocol fixtures for application
// transaction/authorization tests. It never qualifies rendering or confinement.
package reportingfixture

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"html"
	"strings"

	"github.com/JochiRaider/cartulary/internal/modules/reporting"
	"github.com/JochiRaider/cartulary/internal/platform/renderprocess"
)

type Runtime struct{}

func (Runtime) Verify() error { return nil }
func (Runtime) Toolchain() (reporting.ToolchainSnapshot, error) {
	version, executable, hash := "synthetic-application-fixture", "bin/synthetic-browser", digest([]byte("synthetic-toolchain-bytes"))
	fonts := []reporting.ToolchainFontFile{{Path: "installed/fixture.font", SHA256: hash}}
	fontHash := digest([]byte(`[{"path":"installed/fixture.font","sha256":"` + hash + `"}]`))
	return reporting.ToolchainSnapshot{
		SchemaID: "cartulary.reporting_toolchain_snapshot.v2", NodeVersion: version, PackageManager: version, OSImageID: version,
		LockfileSHA256: hash, PackageStoreDigest: hash, FontManifestSHA256: fontHash, FontFiles: fonts,
		SlidevVersion: &version, MermaidVersion: &version, ChromiumVersion: &version, ChromiumExecutablePath: &executable, ChromiumExecutableSHA256: &hash,
		Locale: "C.UTF-8", Timezone: "UTC", ViewportCSSPX: "1280x720", DeviceScaleFactor: 1, ColorScheme: "light",
		NetworkPolicyID: "render_sandbox_policy_v1", BrowserLaunchArgs: []string{}, EnvAllowlist: []reporting.ToolchainEnvironmentValue{}, RenderCommand: []string{"synthetic-renderer"},
	}, nil
}

func (Runtime) Run(ctx context.Context, input []byte, _ []renderprocess.Output) (renderprocess.Result, error) {
	if err := ctx.Err(); err != nil {
		return renderprocess.Result{}, err
	}
	var request struct {
		Operation string `json:"operation"`
		Source    string `json:"source"`
		Identity  string `json:"identity_sha256"`
		Assets    []struct {
			Path      string `json:"path"`
			MediaType string `json:"media_type"`
			Bytes     []byte `json:"bytes"`
		} `json:"assets"`
	}
	if err := json.Unmarshal(input, &request); err != nil {
		return renderprocess.Result{}, err
	}
	path, media, stage := "diagram.svg", "image/svg+xml", "mermaid_render"
	data := []byte(`<svg xmlns="http://www.w3.org/2000/svg"><text>` + html.EscapeString(strings.Join(strings.Fields(request.Source), " ")) + `</text></svg>`)
	if request.Operation == "slidev_pdf" {
		path, media, stage = "deck.pdf", "application/pdf", "slidev_export"
		data = []byte("%PDF-1.7\n% synthetic application fixture " + digest([]byte(request.Source)) + "\n%%EOF\n")
	}
	assetInput := "cartulary.reporting_render_assets.v2\x00"
	for _, asset := range request.Assets {
		assetInput += fmt.Sprintf("%d:%s%d:%s%d:%s", len(asset.Path), asset.Path, len(asset.MediaType), asset.MediaType, len(asset.Bytes), asset.Bytes)
	}
	protocol, err := json.Marshal(map[string]any{
		"schema_id": "cartulary.reporting_render_stage_result.v5", "operation": request.Operation, "identity_sha256": request.Identity, "source_sha256": digest([]byte(request.Source)), "assets_sha256": digest([]byte(assetInput)),
		"artifact":             map[string]any{"path": path, "media_type": media, "bytes": len(data), "sha256": digest(data)},
		"browser_observations": []any{map[string]any{"schema_id": "render_sandbox_observation.v1", "network_policy_id": "render_sandbox_policy_v1", "stage": stage, "observation_kind": "loopback_allowed", "destination_class": "loopback", "allowed": true, "attempt_count": 1, "loopback_port_ref": "render-loopback-1"}},
	})
	return renderprocess.Result{Protocol: protocol, Files: map[string][]byte{path: data}, NetworkObservations: []renderprocess.NetworkObservation{{Kind: "loopback_allowed", DestinationClass: "loopback", Allowed: true, Count: 1}}}, err
}

func digest(data []byte) string { sum := sha256.Sum256(data); return hex.EncodeToString(sum[:]) }
