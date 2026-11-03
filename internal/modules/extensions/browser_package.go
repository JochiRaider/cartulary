package extensions

import (
	"crypto/sha256"
	"encoding/hex"
	"sort"
	"strings"
)

// BrowserPackage contains the final build's metadata and independently packaged
// integrity pin. Physical asset verification belongs to the platform asset owner;
// application composition joins that admitted asset view to this logical catalog.
type BrowserPackage struct {
	IntegrityJSON         []byte
	IntegritySHA256       string
	ClientSupportJSON     []byte
	AssetManifestJSON     []byte
	BootstrapTemplateJSON []byte
}

type browserArtifactSource struct {
	ArtifactSource
	overrides map[string]PackagedArtifact
	paths     []string
	pin       string
}

func (s *browserArtifactSource) Artifact(path string) (PackagedArtifact, bool) {
	if a, ok := s.overrides[path]; ok {
		a.JSON = append([]byte(nil), a.JSON...)
		return a, true
	}
	return s.ArtifactSource.Artifact(path)
}
func (s *browserArtifactSource) ArtifactPaths() []string { return append([]string(nil), s.paths...) }
func (s *browserArtifactSource) IntegritySHA256() string { return s.pin }

func bindBrowserPackage(base ArtifactSource, browser BrowserPackage) (ArtifactSource, error) {
	source := &browserArtifactSource{ArtifactSource: base, overrides: map[string]PackagedArtifact{}, pin: browser.IntegritySHA256}
	for name, body := range map[string][]byte{"registry-integrity.json": browser.IntegrityJSON, "client-support-registry.json": browser.ClientSupportJSON, "client-asset-set-manifest.json": browser.AssetManifestJSON, "client-bootstrap-template.json": browser.BootstrapTemplateJSON} {
		path := generatedExtensionsRoot + name
		if len(body) == 0 {
			return nil, registryInvalid("browser_contract_missing")
		}
		if len(body) > packagedArtifactByteLimit(path) {
			return nil, registryLimit("browser_contract_bytes", packagedArtifactByteLimit(path), len(body))
		}
		bodyCopy := append([]byte(nil), body...)
		sum := sha256.Sum256(bodyCopy)
		source.overrides[path] = PackagedArtifact{JSON: bodyCopy, SHA256: hex.EncodeToString(sum[:])}
	}
	source.paths = append(base.ArtifactPaths(), generatedExtensionsRoot+"client-asset-set-manifest.json", generatedExtensionsRoot+"client-bootstrap-template.json")
	sort.Strings(source.paths)
	integrity, _, err := readArtifactObject(source, generatedExtensionsRoot+"registry-integrity.json")
	if err != nil {
		return nil, err
	}
	if integrity["generator_id"] != "cartulary.embedwebassets.extensions.v1" {
		return nil, registryInvalid("browser_integrity_producer_invalid")
	}
	manifest, manifestArtifact, err := readArtifactObject(source, generatedExtensionsRoot+"client-asset-set-manifest.json")
	if err != nil {
		return nil, err
	}
	if requireExactKeys(manifest, "schema_id", "assets") != nil || manifest["schema_id"] != "cartulary.client_asset_set_manifest.v1" {
		return nil, registryInvalid("browser_manifest_invalid")
	}
	assets, ok := objectSlice(manifest["assets"])
	if !ok || len(assets) == 0 {
		return nil, registryInvalid("browser_assets_missing")
	}
	if len(assets) > 65536 {
		return nil, registryLimit("browser_assets", 65536, len(assets))
	}
	previous := ""
	for _, asset := range assets {
		path := stringValue(asset["logical_path"])
		length, ok := integerValue(asset["byte_length"])
		if requireExactKeys(asset, "logical_path", "byte_length", "sha256") != nil || !validSourceRef(path) || !strings.HasPrefix(path, "assets/") || path <= previous || !ok || length < 0 || length > 1<<30 || !lowerDigest(stringValue(asset["sha256"])) {
			return nil, registryInvalid("browser_asset_identity_invalid")
		}
		previous = path
	}
	support, _, err := readArtifactObject(source, generatedExtensionsRoot+"client-support-registry.json")
	if err != nil {
		return nil, err
	}
	template, _, err := readArtifactObject(base, generatedExtensionsRoot+"client-support-registry.json")
	if err != nil {
		return nil, err
	}
	if requireExactKeys(support, "schema_id", "asset_set_sha256", "client_build_class", "client_build_id", "profiles") != nil || support["schema_id"] != "cartulary.client_extension_support_registry.v1" || support["client_build_class"] != "standard" || support["asset_set_sha256"] != manifestArtifact.SHA256 || !validBrowserBuildID(stringValue(support["client_build_id"])) || canonicalDigest(support["profiles"]) != canonicalDigest(template["profiles"]) {
		return nil, registryInvalid("browser_support_projection_mismatch")
	}
	bootstrap, _, err := readArtifactObject(source, generatedExtensionsRoot+"client-bootstrap-template.json")
	if err != nil {
		return nil, err
	}
	if requireExactKeys(bootstrap, "schema_id", "html_utf8") != nil || bootstrap["schema_id"] != "cartulary.client_bootstrap_template.v1" || strings.Count(stringValue(bootstrap["html_utf8"]), "</head>") != 1 {
		return nil, registryInvalid("browser_bootstrap_invalid")
	}
	return source, nil
}

func validBrowserBuildID(value string) bool {
	if len(value) < 1 || len(value) > 160 {
		return false
	}
	for i, c := range []byte(value) {
		if c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' {
			continue
		}
		if i == 0 || !strings.ContainsRune("_.:-", rune(c)) {
			return false
		}
	}
	return true
}
