package webassets

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"sort"
	"strings"
	"sync"
	"unicode/utf8"

	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
)

//go:embed all:assets
var embeddedFiles embed.FS

const distArchivePath = "assets/generated/web-assets.zip"
const clientAssetManifestPath = "_contracts/client-asset-set-manifest.json"
const clientSupportRegistryPath = "_contracts/client-extension-support-registry.json"

type clientAssetManifest struct {
	Assets   []clientAssetManifestRow `json:"assets"`
	SchemaID string                   `json:"schema_id"`
}

type clientAssetManifestRow struct {
	ByteLength  int64  `json:"byte_length"`
	LogicalPath string `json:"logical_path"`
	SHA256      string `json:"sha256"`
}

type clientSupportRegistry struct {
	AssetSetSHA256   string                         `json:"asset_set_sha256"`
	ClientBuildClass string                         `json:"client_build_class"`
	ClientBuildID    string                         `json:"client_build_id"`
	Profiles         []clientSupportRegistryProfile `json:"profiles"`
	SchemaID         string                         `json:"schema_id"`
}

type clientSupportRegistryProfile struct {
	CapabilityIDs           []string `json:"capability_ids"`
	ProfileID               string   `json:"profile_id"`
	PublicSchemaIDs         []string `json:"public_schema_ids"`
	SupportedContractMajors []int64  `json:"supported_contract_majors"`
	WorkspaceKeys           []string `json:"workspace_keys"`
}

// Contracts is a detached copy of the exact metadata packaged with a Build.
// Extensions admits its logical identities; this package proves physical bytes.
type Contracts struct {
	IntegrityJSON, ClientSupportJSON, AssetManifestJSON, BootstrapTemplateJSON []byte
	IntegritySHA256                                                            string
}

// Build is an immutable, physically verified view of one packaged browser build.
type Build struct {
	archive   *zip.Reader
	root      []byte
	contracts Contracts
}

func (b *Build) RootHTML() []byte { return append([]byte(nil), b.root...) }

type staticFiles struct{ archive *zip.Reader }

func (s staticFiles) Open(name string) (fs.File, error) {
	if name != "assets" && !strings.HasPrefix(name, "assets/") {
		return nil, &fs.PathError{Op: "open", Path: name, Err: fs.ErrNotExist}
	}
	return s.archive.Open(name)
}
func (b *Build) StaticFS() fs.FS { return staticFiles{archive: b.archive} }
func (b *Build) Contracts() Contracts {
	c := b.contracts
	c.IntegrityJSON = append([]byte(nil), c.IntegrityJSON...)
	c.ClientSupportJSON = append([]byte(nil), c.ClientSupportJSON...)
	c.AssetManifestJSON = append([]byte(nil), c.AssetManifestJSON...)
	c.BootstrapTemplateJSON = append([]byte(nil), c.BootstrapTemplateJSON...)
	return c
}

var embeddedBuild = sync.OnceValues(func() (*Build, error) { return Load(embeddedFiles) })

func LoadEmbedded() (*Build, error) { return embeddedBuild() }

// Load snapshots all package inputs before admitting them. No fallback package
// can become a serving build when any required final contract is absent.
func Load(files fs.FS) (*Build, error) {
	archiveBytes, err := fs.ReadFile(files, distArchivePath)
	if err != nil {
		return nil, err
	}
	archiveBytes = append([]byte(nil), archiveBytes...)
	archive, err := zip.NewReader(bytes.NewReader(archiveBytes), int64(len(archiveBytes)))
	if err != nil {
		return nil, err
	}
	b := &Build{archive: archive}
	for _, part := range []struct {
		path  string
		limit int
		out   *[]byte
	}{
		{clientAssetManifestPath, 16 << 20, &b.contracts.AssetManifestJSON},
		{clientSupportRegistryPath, 1 << 20, &b.contracts.ClientSupportJSON},
		{"_contracts/client-bootstrap-template.json", 64 << 20, &b.contracts.BootstrapTemplateJSON},
		{"_contracts/extension-registry-integrity.json", 8 << 20, &b.contracts.IntegrityJSON},
	} {
		f, err := archive.Open(part.path)
		if err != nil {
			return nil, err
		}
		body, readErr := io.ReadAll(io.LimitReader(f, int64(part.limit)+1))
		closeErr := f.Close()
		if readErr != nil {
			return nil, readErr
		}
		if closeErr != nil {
			return nil, closeErr
		}
		if len(body) == 0 || len(body) > part.limit {
			return nil, errors.New("browser contract exceeds byte limit")
		}
		*part.out = body
	}
	pinFile, err := archive.Open("_contracts/extension-registry-integrity.sha256")
	if err != nil {
		return nil, err
	}
	pin, err := io.ReadAll(io.LimitReader(pinFile, 66))
	closeErr := pinFile.Close()
	if closeErr != nil {
		return nil, closeErr
	}
	if err != nil {
		return nil, err
	}
	if len(pin) != 65 || pin[64] != '\n' || !validSHA256(string(pin[:64])) || string(pin[:64]) != digest(b.contracts.IntegrityJSON) {
		return nil, errors.New("browser integrity pin mismatch")
	}
	b.contracts.IntegritySHA256 = string(pin[:64])
	root, err := canonicaljson.DecodeExtension(b.contracts.IntegrityJSON)
	if err != nil {
		return nil, err
	}
	rootObject, ok := root.(map[string]any)
	if !ok || rootObject["schema_id"] != "cartulary.extension_registry_integrity.v2" || rootObject["generator_id"] != "cartulary.embedwebassets.extensions.v1" {
		return nil, errors.New("browser integrity root is invalid")
	}
	rows, ok := rootObject["supporting_contract_artifact_digests"].([]any)
	if !ok {
		return nil, errors.New("browser integrity contracts are missing")
	}
	expected := map[string]string{"client-asset-set-manifest": digest(b.contracts.AssetManifestJSON), "client-support-registry": digest(b.contracts.ClientSupportJSON), "client-bootstrap-template": digest(b.contracts.BootstrapTemplateJSON)}
	seen := map[string]bool{}
	for _, raw := range rows {
		row, ok := raw.(map[string]any)
		if !ok {
			return nil, errors.New("browser integrity row is invalid")
		}
		id, ok := row["artifact_id"].(string)
		if !ok || seen[id] {
			return nil, errors.New("browser integrity identity is invalid")
		}
		seen[id] = true
		if want, ok := expected[id]; ok {
			if row["artifact_sha256"] != want {
				return nil, errors.New("browser supporting contract digest mismatch")
			}
			delete(expected, id)
		}
	}
	if len(expected) != 0 {
		return nil, errors.New("browser supporting contract digest is missing")
	}
	var template struct {
		HTML     string `json:"html_utf8"`
		SchemaID string `json:"schema_id"`
	}
	if err := decodeCanonicalJSON(b.contracts.BootstrapTemplateJSON, 64<<20, &template); err != nil {
		return nil, err
	}
	if template.SchemaID != "cartulary.client_bootstrap_template.v1" || strings.Count(template.HTML, "</head>") != 1 {
		return nil, errors.New("browser bootstrap template is invalid")
	}
	if err := validateClientBuildContracts(archive, b.contracts); err != nil {
		return nil, err
	}
	// JSON embedded in an HTML script element must escape HTML delimiters. This
	// transport encoding does not change the separately admitted canonical bytes.
	registry, err := canonicaljson.DecodeExtension(b.contracts.ClientSupportJSON)
	if err != nil {
		return nil, err
	}
	bootstrap, err := json.Marshal(registry)
	if err != nil {
		return nil, err
	}
	script := append([]byte(`<script id="cartulary-client-extension-support-registry" type="application/json">`), bootstrap...)
	script = append(script, []byte("</script></head>")...)
	b.root = bytes.Replace([]byte(template.HTML), []byte("</head>"), script, 1)
	return b, nil
}

func digest(body []byte) string { sum := sha256.Sum256(body); return hex.EncodeToString(sum[:]) }

func validateClientBuildContracts(archive *zip.Reader, contracts Contracts) error {
	manifestBytes := contracts.AssetManifestJSON
	var manifest clientAssetManifest
	if err := decodeCanonicalJSON(manifestBytes, 16_777_216, &manifest); err != nil {
		return fmt.Errorf("validate client asset manifest: %w", err)
	}
	if manifest.SchemaID != "cartulary.client_asset_set_manifest.v1" || len(manifest.Assets) == 0 || len(manifest.Assets) > 65_536 {
		return errors.New("client asset manifest root is invalid")
	}
	archiveAssets := make(map[string]clientAssetManifestRow, len(archive.File))
	seenNames := map[string]bool{}
	private := map[string]bool{"_contracts/client-asset-set-manifest.json": true, "_contracts/client-extension-support-registry.json": true, "_contracts/client-bootstrap-template.json": true, "_contracts/extension-registry-integrity.json": true, "_contracts/extension-registry-integrity.sha256": true}
	for _, file := range archive.File {
		if seenNames[file.Name] {
			return errors.New("duplicate browser archive entry")
		}
		seenNames[file.Name] = true
		if private[file.Name] {
			if !file.Mode().IsRegular() {
				return errors.New("browser contract is not regular")
			}
			continue
		}
		if file.FileInfo().IsDir() {
			continue
		}
		if !validLogicalPath(file.Name) || !strings.HasPrefix(file.Name, "assets/") || !file.Mode().IsRegular() || file.UncompressedSize64 > 1<<30 || len(archiveAssets) >= 65536 {
			return fmt.Errorf("invalid archived client asset path %q", file.Name)
		}
		if _, duplicate := archiveAssets[file.Name]; duplicate {
			return fmt.Errorf("duplicate archived client asset %q", file.Name)
		}
		reader, err := file.Open()
		if err != nil {
			return err
		}
		hash := sha256.New()
		length, readErr := io.Copy(hash, io.LimitReader(reader, (1<<30)+1))
		closeErr := reader.Close()
		if readErr != nil {
			return readErr
		}
		if closeErr != nil {
			return closeErr
		}
		if length > 1<<30 || uint64(length) != file.UncompressedSize64 {
			return errors.New("client asset length mismatch")
		}
		archiveAssets[file.Name] = clientAssetManifestRow{ByteLength: length, LogicalPath: file.Name, SHA256: hex.EncodeToString(hash.Sum(nil))}
	}
	previousPath := ""
	for _, row := range manifest.Assets {
		if !validLogicalPath(row.LogicalPath) || row.LogicalPath <= previousPath || row.ByteLength < 0 || row.ByteLength > 1_073_741_824 || !validSHA256(row.SHA256) {
			return fmt.Errorf("invalid client asset manifest row %q", row.LogicalPath)
		}
		actual, ok := archiveAssets[row.LogicalPath]
		if !ok || actual != row {
			return fmt.Errorf("client asset manifest mismatch for %q", row.LogicalPath)
		}
		delete(archiveAssets, row.LogicalPath)
		previousPath = row.LogicalPath
	}
	if len(archiveAssets) != 0 {
		paths := make([]string, 0, len(archiveAssets))
		for assetPath := range archiveAssets {
			paths = append(paths, assetPath)
		}
		sort.Strings(paths)
		return fmt.Errorf("client asset manifest omits %q", paths[0])
	}
	manifestDigest := sha256.Sum256(manifestBytes)
	registryBytes := contracts.ClientSupportJSON
	var registry clientSupportRegistry
	if err := decodeCanonicalJSON(registryBytes, 1<<20, &registry); err != nil {
		return fmt.Errorf("validate client support registry: %w", err)
	}
	if registry.SchemaID != "cartulary.client_extension_support_registry.v1" || registry.ClientBuildClass != "standard" || registry.AssetSetSHA256 != hex.EncodeToString(manifestDigest[:]) || registry.ClientBuildID != "cartulary.web.standard.sha256:"+digest([]byte(registry.AssetSetSHA256+"\n"+digest(contracts.BootstrapTemplateJSON)+"\n")) || registry.Profiles == nil || len(registry.Profiles) > 256 {
		return errors.New("client support registry root is invalid")
	}
	previousProfileID := ""
	for _, profile := range registry.Profiles {
		if !validToken(profile.ProfileID) || profile.ProfileID <= previousProfileID || len(profile.SupportedContractMajors) != 1 || profile.SupportedContractMajors[0] <= 0 || profile.WorkspaceKeys == nil || profile.CapabilityIDs == nil || len(profile.CapabilityIDs) != 0 || profile.PublicSchemaIDs == nil || !sortedUniqueStrings(profile.WorkspaceKeys, true) || !sortedUniqueStrings(profile.PublicSchemaIDs, false) {
			return fmt.Errorf("invalid client support profile %q", profile.ProfileID)
		}
		previousProfileID = profile.ProfileID
	}
	return nil
}

func decodeCanonicalJSON(data []byte, maximum int, destination any) error {
	if len(data) == 0 || len(data) > maximum {
		return errors.New("canonical JSON byte envelope is invalid")
	}
	if _, err := canonicaljson.DecodeExtension(data); err != nil {
		return err
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(destination); err != nil {
		return err
	}
	canonical, err := canonicaljson.MarshalExtension(destination)
	if err != nil {
		return err
	}
	if !bytes.Equal(canonical, data) {
		return errors.New("JSON schema or canonical bytes are invalid")
	}
	return nil
}
func validLogicalPath(value string) bool {
	return value != "." && len(value) <= 512 && utf8.ValidString(value) && fs.ValidPath(value) && !strings.ContainsAny(value, "\\\x00")
}

func validSHA256(value string) bool {
	if len(value) != 64 {
		return false
	}
	_, err := hex.DecodeString(value)
	return err == nil && strings.ToLower(value) == value
}

func validToken(value string) bool {
	if len(value) < 1 || len(value) > 64 || value[0] < 'a' || value[0] > 'z' {
		return false
	}
	for index := 1; index < len(value); index++ {
		character := value[index]
		if (character < 'a' || character > 'z') && (character < '0' || character > '9') && character != '_' {
			return false
		}
	}
	return true
}

func sortedUniqueStrings(values []string, tokens bool) bool {
	previous := ""
	for _, value := range values {
		if value == "" || value <= previous || (tokens && !validToken(value)) {
			return false
		}
		previous = value
	}
	return true
}
