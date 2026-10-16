package reporting

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/JochiRaider/cartulary/internal/platform/strictjson"
)

const renderStageRequestSchema = "cartulary.reporting_render_stage_request.v5"
const renderStageResultSchema = "cartulary.reporting_render_stage_result.v5"
const renderStageRequestLimit = 768 << 20

var errRenderStageProtocol = errors.New("reporting render stage protocol invalid")

// Stage data cannot grant executable, filesystem, environment or network
// capabilities. Those belong to the installed runtime and its supervisor.
type renderStageRequest struct {
	SchemaID       string             `json:"schema_id"`
	Operation      string             `json:"operation"`
	Source         string             `json:"source"`
	AdmittedAt     string             `json:"admitted_at"`
	IdentitySHA256 string             `json:"identity_sha256"`
	ExpectedPages  *int               `json:"expected_pages,omitempty"`
	Assets         []renderStageAsset `json:"assets,omitempty"`
}

type renderStageAsset struct {
	Path      string `json:"path"`
	MediaType string `json:"media_type"`
	Bytes     []byte `json:"bytes"`
	SHA256    string `json:"sha256"`
}

var renderDiagramPath = regexp.MustCompile(`^diagrams/[A-Za-z0-9][A-Za-z0-9_.-]{0,127}\.svg$`)

func validateRenderStageAssets(request renderStageRequest) error {
	if len(request.Assets) > 600 || len(request.Source) > 64<<20 || request.Operation != "slidev_pdf" && len(request.Assets) != 0 {
		return errRenderStageProtocol
	}
	localBytes, diagramBytes, localCount, diagramCount := 0, 0, 0, 0
	themeFiles := []RenderBundleFile{}
	for i, asset := range request.Assets {
		if i > 0 && request.Assets[i-1].Path >= asset.Path || len(asset.Bytes) == 0 || hashHex(asset.Bytes) != asset.SHA256 {
			return errRenderStageProtocol
		}
		if renderDiagramPath.MatchString(asset.Path) && asset.MediaType == "image/svg+xml" {
			diagramCount++
			diagramBytes += len(asset.Bytes)
			if diagramCount > 100 || len(asset.Bytes) > 8<<20 || diagramBytes > 32<<20 {
				return errRenderStageProtocol
			}
		} else if shareableBundleMember(RenderBundleManifestFile{Path: asset.Path, Role: renderBundleRoleLocalAsset, MediaType: asset.MediaType, RequiredForRelease: true}) || templateThemeMember(themeFileDeclaration(asset.Path, asset.MediaType)) {
			if strings.HasPrefix(asset.Path, "theme/") {
				themeFiles = append(themeFiles, newBundleFile(asset.Path, renderBundleRoleLocalTheme, asset.MediaType, asset.Bytes, true))
			}
			localCount++
			localBytes += len(asset.Bytes)
			if localCount > 500 || len(asset.Bytes) > 50<<20 || localBytes > 250<<20 {
				return errRenderStageProtocol
			}
		} else {
			return errRenderStageProtocol
		}
		if err := validateTemplateFileContent(asset.MediaType, asset.Bytes); err != nil {
			return errRenderStageProtocol
		}
	}
	if validateThemeBundleFiles(themeFiles) != nil {
		return errRenderStageProtocol
	}
	return nil
}

func renderStageAssetsDigest(assets []renderStageAsset) string {
	// Length framing binds sorted paths, media types and exact binary contents.
	// Base64 is a private transport encoding, never the artifact's hash preimage.
	var input bytes.Buffer
	input.WriteString("cartulary.reporting_render_assets.v2\x00")
	for _, a := range assets {
		fmt.Fprintf(&input, "%d:%s%d:%s%d:", len(a.Path), a.Path, len(a.MediaType), a.MediaType, len(a.Bytes))
		input.Write(a.Bytes)
	}
	return hashHex(input.Bytes())
}

type renderStageObservation struct {
	SchemaID         string `json:"schema_id"`
	NetworkPolicyID  string `json:"network_policy_id"`
	Stage            string `json:"stage"`
	ObservationKind  string `json:"observation_kind"`
	DestinationClass string `json:"destination_class"`
	Allowed          bool   `json:"allowed"`
	AttemptCount     int    `json:"attempt_count"`
	LoopbackPortRef  string `json:"loopback_port_ref"`
}

type renderStageArtifact struct {
	Path      string `json:"path"`
	MediaType string `json:"media_type"`
	Bytes     int64  `json:"bytes"`
	SHA256    string `json:"sha256"`
}

type renderStageResult struct {
	SchemaID            string                   `json:"schema_id"`
	Operation           string                   `json:"operation"`
	IdentitySHA256      string                   `json:"identity_sha256"`
	SourceSHA256        string                   `json:"source_sha256"`
	AssetsSHA256        string                   `json:"assets_sha256"`
	Artifact            renderStageArtifact      `json:"artifact"`
	BrowserObservations []renderStageObservation `json:"browser_observations"`
}

func decodeClosedStageObject(data []byte, target any, keys ...string) error {
	if !utf8.Valid(data) {
		return errRenderStageProtocol
	}
	object, err := strictjson.DecodeObject(bytes.NewReader(data))
	if err != nil || len(object) != len(keys) {
		return errRenderStageProtocol
	}
	for _, key := range keys {
		value, ok := object[key]
		if !ok || bytes.Equal(bytes.TrimSpace(value), []byte("null")) {
			return errRenderStageProtocol
		}
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return errRenderStageProtocol
	}
	return nil
}

// This validates transport integrity, the actual emitted artifact and browser
// observations. Native sandbox enforcement/observations and measured toolchain
// admission are separate mandatory gates, not implied by a successful decode.
func admitRenderStageResult(request renderStageRequest, protocol []byte, files map[string][]byte) (renderStageResult, error) {
	invalid := func() (renderStageResult, error) { return renderStageResult{}, errRenderStageProtocol }
	if len(protocol) == 0 || len(protocol) > 64<<10 || request.SchemaID != renderStageRequestSchema || validateRenderStageAssets(request) != nil {
		return invalid()
	}
	var result renderStageResult
	if err := decodeClosedStageObject(protocol, &result, "schema_id", "operation", "identity_sha256", "source_sha256", "assets_sha256", "artifact", "browser_observations"); err != nil {
		return invalid()
	}
	if result.SchemaID != renderStageResultSchema || result.Operation != request.Operation || result.IdentitySHA256 != request.IdentitySHA256 || result.SourceSHA256 != hashHex([]byte(request.Source)) || result.AssetsSHA256 != renderStageAssetsDigest(request.Assets) {
		return invalid()
	}
	var members map[string]json.RawMessage
	if err := json.Unmarshal(protocol, &members); err != nil {
		return invalid()
	}
	if err := decodeClosedStageObject(members["artifact"], &result.Artifact, "path", "media_type", "bytes", "sha256"); err != nil {
		return invalid()
	}
	var observations []json.RawMessage
	if err := json.Unmarshal(members["browser_observations"], &observations); err != nil || len(observations) != 1 {
		return invalid()
	}
	var observation renderStageObservation
	if err := decodeClosedStageObject(observations[0], &observation, "schema_id", "network_policy_id", "stage", "observation_kind", "destination_class", "allowed", "attempt_count", "loopback_port_ref"); err != nil {
		return invalid()
	}
	path, media, stage := "", "", ""
	switch request.Operation {
	case "slidev_pdf":
		path, media, stage = "deck.pdf", "application/pdf", "slidev_export"
	case "mermaid_svg":
		path, media, stage = "diagram.svg", "image/svg+xml", "mermaid_render"
	default:
		return invalid()
	}
	// Successful browser traffic is coalesced into exactly one declared-loopback
	// observation. Any blocked observation makes the attempt ineligible.
	if observation.SchemaID != "render_sandbox_observation.v1" || observation.NetworkPolicyID != "render_sandbox_policy_v1" || observation.Stage != stage || observation.ObservationKind != "loopback_allowed" || observation.DestinationClass != "loopback" || !observation.Allowed || observation.AttemptCount < 1 || observation.LoopbackPortRef != "render-loopback-1" {
		return invalid()
	}
	data, ok := files[path]
	if len(files) != 1 || !ok || result.Artifact.Path != path || result.Artifact.MediaType != media || result.Artifact.Bytes != int64(len(data)) || len(data) == 0 || len(data) > 500<<20 || result.Artifact.SHA256 != hashHex(data) {
		return invalid()
	}
	if request.Operation == "mermaid_svg" {
		if err := validateRenderedSVG(data); err != nil {
			return renderStageResult{}, err
		}
	} else if !bytes.HasPrefix(data, []byte("%PDF-")) {
		return invalid()
	}
	return result, nil
}
