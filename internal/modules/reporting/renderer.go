package reporting

import (
	"context"
	"encoding/json"
	"errors"
	"sort"
	"strconv"

	"github.com/JochiRaider/cartulary/internal/platform/renderprocess"
)

// RenderRuntime is an explicit assembly dependency. It grants only the pinned
// stage protocol, never caller-selected executables, environment or paths.
// Test compositions may supply a fixture; production uses NewInstalledRenderer.
type RenderRuntime interface {
	Toolchain() (ToolchainSnapshot, error)
	Verify() error
	Run(context.Context, []byte, []renderprocess.Output) (renderprocess.Result, error)
}

type Renderer struct{ runtime RenderRuntime }

func NewRenderer(runtime RenderRuntime) (*Renderer, error) {
	if runtime == nil {
		return nil, errors.New("reporting renderer runtime required")
	}
	if err := runtime.Verify(); err != nil {
		return nil, err
	}
	return &Renderer{runtime: runtime}, nil
}

// NewInstalledRenderer admits an exact complete installation. Missing or altered
// installations never select another binary or fall back to source placeholders.
func NewInstalledRenderer(directory, temporaryRoot string) (*Renderer, error) {
	installation, err := renderprocess.OpenInstallation(directory)
	if err != nil {
		return nil, err
	}
	// OpenInstallation already verifies the complete file set. Avoid repeating
	// that startup scan; every attempt still verifies before and after rendering.
	return &Renderer{runtime: &installedRenderRuntime{installation: installation, temporaryRoot: temporaryRoot}}, nil
}

type ToolchainEnvironmentValue struct {
	Name        string `json:"name"`
	ValueSHA256 string `json:"value_sha256"`
}

type ToolchainFontFile struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
}

// ToolchainSnapshot projects the installed bytes and the exact supervisor
// profile. Attempt time and actually used renderers are selected by Reporting.
type ToolchainSnapshot struct {
	CreatedAt                string                      `json:"created_at"`
	SchemaID                 string                      `json:"schema_id"`
	NodeVersion              string                      `json:"node_version"`
	PackageManager           string                      `json:"package_manager"`
	LockfileSHA256           string                      `json:"lockfile_sha256"`
	PackageStoreDigest       string                      `json:"package_store_digest"`
	SlidevVersion            *string                     `json:"slidev_version"`
	MermaidVersion           *string                     `json:"mermaid_version"`
	ChromiumVersion          *string                     `json:"chromium_version"`
	ChromiumExecutablePath   *string                     `json:"chromium_executable_path"`
	ChromiumExecutableSHA256 *string                     `json:"chromium_executable_sha256"`
	OSImageID                string                      `json:"os_image_id"`
	OSImageSHA256            *string                     `json:"os_image_sha256"`
	FontManifestSHA256       string                      `json:"font_manifest_sha256"`
	FontFiles                []ToolchainFontFile         `json:"font_files"`
	Locale                   string                      `json:"locale"`
	Timezone                 string                      `json:"timezone"`
	ViewportCSSPX            string                      `json:"viewport_css_px"`
	DeviceScaleFactor        int                         `json:"device_scale_factor"`
	ColorScheme              string                      `json:"color_scheme"`
	BrowserLaunchArgs        []string                    `json:"browser_launch_args"`
	EnvAllowlist             []ToolchainEnvironmentValue `json:"env_allowlist"`
	RenderCommand            []string                    `json:"render_command"`
	NetworkPolicyID          string                      `json:"network_policy_id"`
}

const installedRendererPort = 32123 // Each attempt owns a separate network namespace.

type installedRenderRuntime struct {
	installation  *renderprocess.Installation
	temporaryRoot string
}

func (r *installedRenderRuntime) Verify() error { return r.installation.Verify() }
func (r *installedRenderRuntime) Run(ctx context.Context, input []byte, outputs []renderprocess.Output) (renderprocess.Result, error) {
	return r.installation.Runtime(r.temporaryRoot, installedRendererPort).Run(ctx, input, outputs)
}
func (r *installedRenderRuntime) Toolchain() (ToolchainSnapshot, error) {
	manifest := r.installation.Manifest()
	profile := r.installation.Runtime(r.temporaryRoot, installedRendererPort)
	arguments, err := r.installation.BrowserLaunchArgs()
	if err != nil {
		return ToolchainSnapshot{}, err
	}
	environment := profile.Environment
	environment["CARTULARY_RENDER_LOOPBACK_PORT"] = strconv.Itoa(installedRendererPort)
	names := make([]string, 0, len(environment))
	for name := range environment {
		names = append(names, name)
	}
	sort.Strings(names)
	values := make([]ToolchainEnvironmentValue, 0, len(names))
	for _, name := range names {
		values = append(values, ToolchainEnvironmentValue{Name: name, ValueSHA256: hashHex([]byte(environment[name]))})
	}
	slidev, mermaid, digest := manifest.Packages["@slidev/cli"], manifest.Packages["mermaid"], r.installation.Digest()
	fonts := installedToolchainFontFiles(manifest.Files)
	fontJSON, err := canonicalJSON(fonts)
	if err != nil {
		return ToolchainSnapshot{}, err
	}
	return ToolchainSnapshot{
		SchemaID: "cartulary.reporting_toolchain_snapshot.v2", NodeVersion: manifest.NodeVersion,
		PackageManager: manifest.PackageManager, LockfileSHA256: manifest.LockfileSHA256, PackageStoreDigest: manifest.PackageStoreDigest,
		SlidevVersion: &slidev, MermaidVersion: &mermaid, ChromiumVersion: &manifest.ChromiumVersion,
		ChromiumExecutablePath: &manifest.ChromiumExecutablePath, ChromiumExecutableSHA256: &manifest.ChromiumExecutableSHA256,
		OSImageID: "cartulary.reporting.runtime/" + digest, OSImageSHA256: &digest, FontManifestSHA256: hashHex(fontJSON), FontFiles: fonts,
		Locale: "C.UTF-8", Timezone: "UTC", ViewportCSSPX: "1280x720", DeviceScaleFactor: 1, ColorScheme: "light",
		BrowserLaunchArgs: arguments, EnvAllowlist: values, RenderCommand: profile.Command, NetworkPolicyID: "render_sandbox_policy_v1",
	}, nil
}

func (r *Renderer) renderStage(ctx context.Context, request renderStageRequest, limits renderLimits) ([]byte, error) {
	if err := validateRenderStageAssets(request); err != nil {
		return nil, err
	}
	timeout, output := limits.timeout("mermaid_per_diagram"), "diagram.svg"
	maxBytes := int64(500 << 20) // Supervisor bound for SVG; no PDF allowance implied.
	if request.Operation == "slidev_pdf" {
		timeout, output, maxBytes = limits.timeout("slidev_export"), "deck.pdf", limits["rendered_pdf.byte_size"]
	}
	stage, key := "mermaid_render", "timeout.mermaid_per_diagram"
	if request.Operation == "slidev_pdf" {
		stage, key = "slidev_export", "timeout.slidev_export"
	}
	ctx, cancel := withRenderDeadline(ctx, timeout, stage, key)
	defer cancel()
	input, err := json.Marshal(request)
	if len(input) > renderStageRequestLimit {
		return nil, errRenderStageProtocol
	}
	if err != nil {
		return nil, err
	}
	result, err := r.runtime.Run(ctx, input, []renderprocess.Output{{Path: output, MaxBytes: maxBytes}})
	if err != nil {
		if request.Operation == "slidev_pdf" && errors.Is(err, renderprocess.ErrOutputLimit) && !errors.Is(err, renderprocess.ErrCleanup) {
			return nil, newRenderValidationError("bundle_resource_limit_exceeded", "rendered_pdf_bytes_exceeded")
		}
		classified := classifyRenderContextError(ctx, err)
		classified.cause = err
		return nil, classified
	}
	if err := ctx.Err(); err != nil {
		return nil, classifyRenderContextError(ctx, err)
	}
	if request.Operation == "slidev_pdf" {
		if err := limits.check("rendered_pdf.byte_size", int64(len(result.Files[output]))); err != nil {
			return nil, err
		}
	}
	if len(result.NetworkObservations) != 1 {
		return nil, newRenderValidationError("security_sandbox_violation", "sandbox_observation_missing")
	}
	observation := result.NetworkObservations[0]
	if !observation.Allowed || observation.Kind != "loopback_allowed" || observation.DestinationClass != "loopback" || observation.Count < 1 {
		return nil, newRenderValidationError("security_sandbox_violation", "sandbox_observation_missing")
	}
	if _, err := admitRenderStageResult(request, result.Protocol, result.Files); err != nil {
		return nil, newRenderValidationError("security_sandbox_violation", "external_reference_in_rendered_output")
	}
	if err := ctx.Err(); err != nil {
		return nil, classifyRenderContextError(ctx, err)
	}
	return result.Files[output], nil
}

func classifyRenderRuntimeError(err error) *renderValidationError {
	if errors.Is(err, renderprocess.ErrCleanup) {
		return newRenderValidationError("security_sandbox_violation", "partial_output_cleanup_failed")
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return newRenderValidationError("render_timeout", "render_timeout")
	}
	if errors.Is(err, context.Canceled) {
		return newRenderValidationError("render_canceled", "render_canceled")
	}
	var network *renderprocess.NetworkError
	if errors.As(err, &network) {
		if network.Kind == "loopback_blocked" {
			return newRenderValidationError("security_sandbox_violation", "loopback_not_declared")
		}
		return newRenderValidationError("remote_asset_reference", "outbound_request_observed")
	}
	var execution *renderprocess.ExecutionError
	if errors.As(err, &execution) {
		switch execution.Reason {
		case "click_export_page_count_mismatch":
			return newRenderValidationError("slidev_export_failed", execution.Reason)
		case "invalid_mermaid_construct":
			return newRenderValidationError("mermaid_invalid", execution.Reason)
		case "outbound_request_observed":
			return newRenderValidationError("remote_asset_reference", execution.Reason)
		case "loopback_not_declared", "external_reference_in_rendered_output":
			return newRenderValidationError("security_sandbox_violation", execution.Reason)
		case "renderer_svg_export_failed":
			return newRenderValidationError("mermaid_render_failed", "renderer_process_failed")
		case "renderer_server_start_failed", "renderer_page_load_failed", "renderer_print_surface_missing", "renderer_fonts_or_paint_failed", "renderer_pdf_export_failed", "browser_render_error", "browser_module_export_missing", "browser_slide_import_failed", "renderer_resource_missing", "renderer_resource_transform_failed", "renderer_resource_load_failed":
			return newRenderValidationError("slidev_export_failed", "renderer_process_failed")
		}
	}
	return newRenderValidationError("security_sandbox_violation", "render_runtime_unavailable")
}

func (r *Renderer) Render(ctx context.Context, payload renderInput) (RenderedRelease, string, error) {
	if err := validateFrozenSourceFacts(payload); err != nil {
		failure := renderFailureAt(err, "export_model_materialization", "export_model_invalid", "invalid_type")
		return RenderedRelease{}, failure.ReasonCode, failure
	}
	if r == nil || r.runtime == nil {
		return RenderedRelease{}, "sandbox_observation_missing", newRenderValidationError("security_sandbox_violation", "sandbox_observation_missing")
	}
	limits, err := resolveRenderLimits(payload.TemplateContract.DeclaredLimits)
	if err != nil {
		var failure *renderValidationError
		if errors.As(err, &failure) {
			return RenderedRelease{}, failure.ReasonCode, err
		}
		return RenderedRelease{}, "template_manifest_invalid", err
	}
	ctx, cancel := withRenderDeadline(ctx, limits.timeout("total_attempt"), "release_state", "timeout.total_attempt")
	defer cancel()
	if err := ctx.Err(); err != nil {
		classified := classifyRenderContextError(ctx, err)
		return RenderedRelease{}, classified.ReasonCode, classified
	}
	if err := r.runtime.Verify(); err != nil {
		return RenderedRelease{}, "toolchain_snapshot_mismatch", newRenderValidationError("nondeterministic_render", "toolchain_snapshot_mismatch")
	}
	contract := payload.TemplateContract
	if err := validateFrozenRenderProfile(payload); err != nil {
		failure := renderFailureAt(err, "redaction", "redaction_manifest_invalid", "redaction_profile_invalid")
		return RenderedRelease{}, failure.ReasonCode, failure
	}
	profile, profileSHA := payload.RedactionProfile, payload.RedactionProfileSHA256
	partial := RenderedRelease{Profile: profile, ProfileSHA256: profileSHA}
	redactionCtx, finishRedaction := withRenderDeadline(ctx, limits.timeout("redaction_tokenization"), "redaction", "timeout.redaction_tokenization")
	redaction, err := redactRenderInput(redactionCtx, payload)
	if redactionCtx.Err() != nil {
		err = classifyRenderContextError(redactionCtx, redactionCtx.Err())
	}
	finishRedaction()
	if errors.Is(err, ErrInvalidRedactionProfile) {
		failure := renderFailureAt(err, "redaction", "redaction_manifest_invalid", "redaction_profile_invalid")
		return partial, failure.ReasonCode, failure
	}
	if err != nil {
		failure := renderFailureAt(err, "redaction", "redaction_manifest_invalid", "invalid_scalar_value")
		return partial, failure.ReasonCode, failure
	}
	partial.Redaction = redaction
	if payload.ReleaseScope == ReleaseScopeExternal {
		repeatCtx, finish := withRenderDeadline(ctx, limits.timeout("redaction_tokenization"), "determinism", "timeout.redaction_tokenization")
		repeated, repeatErr := redactRenderInput(repeatCtx, payload)
		if repeatCtx.Err() != nil {
			repeatErr = classifyRenderContextError(repeatCtx, repeatCtx.Err())
		}
		finish()
		if repeatErr == nil {
			repeatErr = compareRenderMaterialization(redaction, repeated)
		}
		if repeatErr != nil {
			failure := renderFailureAt(repeatErr, "determinism", "nondeterministic_render", "export_model_mismatch")
			return partial, failure.ReasonCode, failure
		}
	}
	bundle, err := r.renderReportBundle(ctx, contract, payload.OutputKind, redaction.Model, redaction.ProfileView, redaction.ManifestSHA256, payload.ReleaseScope, payload.OutputOptions, payload.CompositionJSON, RedactionBundleArtifacts{
		RedactionManifestJSON: redaction.ManifestJSON,
		ProfileViewJSON:       redaction.ProfileViewJSON,
		TokenManifestJSON:     redaction.TokenManifestJSON,
		TokenManifestSHA256:   redaction.TokenManifestSHA256,
	})
	if err != nil {
		failure := classifyBundleFailure(err)
		return partial, failure.ReasonCode, failure
	}
	manifestJSON, err := canonicalJSON(redaction.Manifest)
	if err != nil {
		failure := renderFailureAt(err, "redaction", "redaction_manifest_invalid", "invalid_type")
		return partial, failure.ReasonCode, failure
	}
	if err := ctx.Err(); err != nil {
		classified := classifyRenderContextError(ctx, err)
		return partial, classified.ReasonCode, classified
	}
	if err := r.runtime.Verify(); err != nil {
		return partial, "toolchain_snapshot_mismatch", newRenderValidationError("nondeterministic_render", "toolchain_snapshot_mismatch")
	}
	return RenderedRelease{
		Profile:                 profile,
		ProfileSHA256:           profileSHA,
		Redaction:               redaction,
		OutputMediaType:         bundle.PrimaryMedia,
		OutputSHA256:            bundle.ManifestSHA256,
		RenderBundle:            bundle,
		RedactionManifestSHA256: redaction.ManifestSHA256,
		RedactionManifestJSON:   manifestJSON,
	}, "", nil
}
