package reporting

import (
	"bytes"
	"encoding/json"
	"errors"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"path"
	"sort"
	"strings"
)

// TemplatePack is an assembly input, never a public request or filesystem
// capability. Files are exact regular-file contents already owned by Reporting.
// The registry copies them; admitted jobs retain their own independent copy.
type TemplatePack struct {
	Manifest TemplateContract
	Files    []TemplatePackFile
}

type TemplatePackFile struct {
	SourcePath string `json:"source_path"`
	Bytes      []byte `json:"bytes"`
}

type TemplatePackRegistry struct{ packs map[[2]string]TemplatePack }

func NewTemplatePackRegistry(packs ...TemplatePack) (*TemplatePackRegistry, error) {
	builtin, _ := ResolveTemplateContract(DefaultTemplateID, DefaultTemplateVersion)
	r := &TemplatePackRegistry{packs: map[[2]string]TemplatePack{}}
	for _, pack := range append([]TemplatePack{{Manifest: builtin, Files: []TemplatePackFile{}}}, packs...) {
		key := [2]string{pack.Manifest.TemplateID, pack.Manifest.TemplateVersion}
		if _, exists := r.packs[key]; exists {
			return nil, errors.New("reporting: duplicate template pack identity")
		}
		for _, kind := range pack.Manifest.SupportedOutputKinds {
			for _, scope := range pack.Manifest.SupportedReleaseScopes {
				if err := validateTemplateContract(pack.Manifest, kind, scope); err != nil {
					return nil, err
				}
			}
		}
		if len(pack.Manifest.SupportedOutputKinds) == 0 || len(pack.Manifest.SupportedReleaseScopes) == 0 {
			return nil, errors.New("reporting: empty template pack applicability")
		}
		if err := validateTemplatePackFiles(pack.Manifest, pack.Files); err != nil {
			return nil, err
		}
		copy, err := cloneTemplatePack(pack)
		if err != nil {
			return nil, err
		}
		r.packs[key] = copy
	}
	return r, nil
}

func cloneTemplatePack(pack TemplatePack) (TemplatePack, error) {
	encoded, err := json.Marshal(pack)
	var copy TemplatePack
	if err == nil {
		decoder := json.NewDecoder(bytes.NewReader(encoded))
		decoder.UseNumber()
		err = decoder.Decode(&copy)
	}
	return copy, err
}

func (r *TemplatePackRegistry) resolve(id, version string) (TemplatePack, bool) {
	if r == nil {
		return TemplatePack{}, false
	}
	pack, ok := r.packs[[2]string{id, version}]
	if !ok {
		return TemplatePack{}, false
	}
	copy, err := cloneTemplatePack(pack)
	return copy, err == nil
}

func templateManifestDigest(contract TemplateContract) (string, error) {
	raw, err := canonicalJSON(contract)
	if err != nil {
		return "", err
	}
	return reportingObjectDigest(templateManifestSchema, raw), nil
}

func validateAdmittedTemplate(input renderInput, digest string) error {
	actual, err := templateManifestDigest(input.TemplateContract)
	if err != nil || !sha256HexPattern.MatchString(digest) || actual != digest ||
		input.TemplateID != input.TemplateContract.TemplateID || input.TemplateVersion != input.TemplateContract.TemplateVersion {
		return errors.New("reporting: altered admitted template pack")
	}
	if err := validateTemplateContract(input.TemplateContract, input.OutputKind, input.ReleaseScope); err != nil {
		return err
	}
	return validateTemplatePackFiles(input.TemplateContract, input.TemplateFiles)
}

func validateTemplateAssetDeclarations(contract TemplateContract) error {
	limits, err := resolveRenderLimits(contract.DeclaredLimits)
	if err != nil {
		return err
	}
	if err := limits.check("local_assets.count", int64(len(contract.Assets))); err != nil {
		return err
	}
	ids, paths, sources := map[string]bool{}, map[string]bool{}, map[string]bool{}
	var total int64
	for i, asset := range contract.Assets {
		if asset.SchemaID != "template_asset_item.v1" || !validReportingIdentifier(asset.AssetID) || ids[asset.AssetID] || paths[asset.BundlePath] || sources[asset.SourcePath] ||
			!portableMemberPath(asset.SourcePath) || !templateAssetMember(asset) || !sha256HexPattern.MatchString(asset.SHA256) || asset.ByteSize <= 0 || !asset.RequiredForRelease {
			return newRenderValidationError("bundle_manifest_invalid", "invalid_asset_media_type")
		}
		if i > 0 && contract.Assets[i-1].BundlePath >= asset.BundlePath {
			return newRenderValidationError("bundle_manifest_invalid", "array_order_invalid")
		}
		if err := limits.check("local_asset.byte_size", asset.ByteSize); err != nil {
			return err
		}
		total += asset.ByteSize
		if err := limits.check("local_assets.total_bytes", total); err != nil {
			return err
		}
		ids[asset.AssetID], paths[asset.BundlePath], sources[asset.SourcePath] = true, true, true
	}
	return nil
}

func templateAssetMember(asset TemplateAsset) bool {
	if asset.AssetRole == "local_theme" {
		return templateThemeMember(asset)
	}
	if !portableMemberPath(asset.BundlePath) || asset.AssetRole != "local_asset" {
		return false
	}
	parts := strings.Split(asset.BundlePath, "/")
	if len(parts) != 3 || parts[0] != "assets" || parts[1] != asset.AssetID || strings.ContainsAny(asset.BundlePath, "%?#") {
		return false
	}
	switch asset.MediaType {
	case "image/png":
		return path.Ext(parts[2]) == ".png"
	case "image/jpeg":
		return path.Ext(parts[2]) == ".jpg" || path.Ext(parts[2]) == ".jpeg"
	case "image/svg+xml":
		return path.Ext(parts[2]) == ".svg"
	default:
		return false
	}
}

func validateTemplatePackFiles(contract TemplateContract, files []TemplatePackFile) error {
	if err := validateTemplateAssetDeclarations(contract); err != nil {
		return err
	}
	if len(files) != len(contract.Assets) {
		return newRenderValidationError("undeclared_asset", "undeclared_asset_ref")
	}
	byPath := map[string]TemplateAsset{}
	for _, asset := range contract.Assets {
		byPath[asset.SourcePath] = asset
	}
	previous := ""
	for _, file := range files {
		asset, exists := byPath[file.SourcePath]
		if !exists || file.SourcePath <= previous || int64(len(file.Bytes)) != asset.ByteSize || hashHex(file.Bytes) != asset.SHA256 {
			return newRenderValidationError("undeclared_asset", "undeclared_asset_ref")
		}
		if err := validateTemplateFileContent(asset.MediaType, file.Bytes); err != nil {
			return err
		}
		previous = file.SourcePath
	}
	return validateTemplateTheme(contract, files)
}

func validateTemplateFileContent(media string, data []byte) error {
	switch media {
	case "font/woff2":
		return validateThemeFont(data)
	case themeCSSMedia:
		if len(data) == 0 || len(data) > 8192 {
			return newRenderValidationError("asset_limit_exceeded", "local_asset_bytes_exceeded")
		}
		return nil // The complete theme validator checks grammar and dependencies.
	default:
		return validateLocalImage(media, data)
	}
}

func validateLocalImage(media string, data []byte) error {
	if len(data) == 0 || len(data) > 50<<20 {
		return newRenderValidationError("asset_limit_exceeded", "local_asset_bytes_exceeded")
	}
	if media == "image/svg+xml" {
		return validateRenderedSVG(data)
	}
	// DecodeConfig inspects dimensions without allocating attacker-sized pixel
	// storage. Full decoding is left to the confined browser. Header/media
	// disagreement, non-image bytes and excessive decoded surfaces fail here.
	config, format, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil || media != "image/"+format || config.Width <= 0 || config.Height <= 0 || int64(config.Width) > 40_000_000/int64(config.Height) {
		return newRenderValidationError("bundle_manifest_invalid", "invalid_asset_media_type")
	}
	return nil
}

func retainedTemplateAssetFiles(contract TemplateContract, files []TemplatePackFile, assets []ReportingAssetDeclaration) ([]RenderBundleFile, error) {
	if err := validateTemplatePackFiles(contract, files); err != nil {
		return nil, err
	}
	bytesBySource := map[string][]byte{}
	for _, file := range files {
		bytesBySource[file.SourcePath] = file.Bytes
	}
	retained := map[string]ReportingAssetDeclaration{}
	for _, asset := range assets {
		if _, exists := retained[asset.AssetID]; exists {
			return nil, newRenderValidationError("export_model_invalid", "invalid_scalar_value")
		}
		retained[asset.AssetID] = asset
	}
	result := []RenderBundleFile{}
	for _, asset := range contract.Assets {
		if declaration, ok := retained[asset.AssetID]; ok {
			if !sameReportingValue(declaration, templateAssetDeclaration(asset)) {
				return nil, newRenderValidationError("export_model_invalid", "invalid_scalar_value")
			}
			result = append(result, newBundleFile(asset.BundlePath, asset.AssetRole, asset.MediaType, append([]byte(nil), bytesBySource[asset.SourcePath]...), asset.RequiredForRelease))
		}
	}
	if len(result) != len(retained) {
		return nil, newRenderValidationError("undeclared_asset", "undeclared_asset_ref")
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Path < result[j].Path })
	return result, nil
}
