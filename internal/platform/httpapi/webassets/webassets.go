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
	"path"
	"sort"
	"strings"
)

//go:embed fallback/index.html all:dist
var embeddedFiles embed.FS

const distArchivePath = "dist/web-assets.zip"
const clientAssetManifestPath = "dist/client-asset-set-manifest.json"
const clientSupportRegistryPath = "dist/client-extension-support-registry.json"

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

func ReadIndexHTML() ([]byte, error) {
	return readIndexHTML(embeddedFiles)
}

// ReadBrowserRootHTML verifies the build-bound asset contracts and injects the
// final client support registry into the dynamic browser bootstrap response.
// The injected response is intentionally outside the immutable static mapping.
func ReadBrowserRootHTML() ([]byte, error) {
	return readBrowserRootHTML(embeddedFiles)
}

func ClientSupportRegistrySHA256() (string, bool, error) {
	registry, present, err := readClientBuildContracts(embeddedFiles)
	if err != nil || !present {
		return "", present, err
	}
	digest := sha256.Sum256(registry)
	return hex.EncodeToString(digest[:]), true, nil
}

func StaticFS() (fs.FS, error) {
	return staticFS(embeddedFiles)
}

func readIndexHTML(files fs.FS) ([]byte, error) {
	archive, ok, err := openDistArchive(files)
	if err != nil {
		return nil, err
	}
	if ok {
		return fs.ReadFile(archive, "index.html")
	}
	return fs.ReadFile(files, "fallback/index.html")
}

func staticFS(files fs.FS) (fs.FS, error) {
	archive, ok, err := openDistArchive(files)
	if err != nil {
		return nil, err
	}
	if ok {
		return archive, nil
	}
	return fs.Sub(files, "fallback")
}

func readBrowserRootHTML(files fs.FS) ([]byte, error) {
	root, err := readIndexHTML(files)
	if err != nil {
		return nil, err
	}
	registry, present, err := readClientBuildContracts(files)
	if err != nil {
		return nil, err
	}
	if !present {
		return root, nil
	}
	closingHead := []byte("</head>")
	if bytes.Count(root, closingHead) != 1 {
		return nil, errors.New("browser root must contain exactly one closing head element")
	}
	registry = bytes.TrimSuffix(registry, []byte("\n"))
	bootstrap := append([]byte(`<script id="cartulary-client-extension-support-registry" type="application/json">`), registry...)
	bootstrap = append(bootstrap, []byte("</script></head>")...)
	return bytes.Replace(root, closingHead, bootstrap, 1), nil
}

func readClientBuildContracts(files fs.FS) ([]byte, bool, error) {
	archive, present, err := openDistArchive(files)
	if err != nil || !present {
		return nil, false, err
	}
	manifestBytes, err := fs.ReadFile(files, clientAssetManifestPath)
	if err != nil {
		return nil, false, fmt.Errorf("read client asset manifest: %w", err)
	}
	var manifest clientAssetManifest
	if err := decodeCanonicalJSON(manifestBytes, 16_777_216, &manifest); err != nil {
		return nil, false, fmt.Errorf("validate client asset manifest: %w", err)
	}
	if manifest.SchemaID != "cartulary.client_asset_set_manifest.v1" || len(manifest.Assets) == 0 || len(manifest.Assets) > 65_536 {
		return nil, false, errors.New("client asset manifest root is invalid")
	}
	archiveAssets := make(map[string]clientAssetManifestRow, len(archive.File))
	for _, file := range archive.File {
		if file.FileInfo().IsDir() {
			continue
		}
		if !validLogicalPath(file.Name) {
			return nil, false, fmt.Errorf("invalid archived client asset path %q", file.Name)
		}
		if _, duplicate := archiveAssets[file.Name]; duplicate {
			return nil, false, fmt.Errorf("duplicate archived client asset %q", file.Name)
		}
		data, readErr := fs.ReadFile(archive, file.Name)
		if readErr != nil {
			return nil, false, readErr
		}
		digest := sha256.Sum256(data)
		archiveAssets[file.Name] = clientAssetManifestRow{
			ByteLength: int64(len(data)), LogicalPath: file.Name, SHA256: hex.EncodeToString(digest[:]),
		}
	}
	previousPath := ""
	for _, row := range manifest.Assets {
		if !validLogicalPath(row.LogicalPath) || row.LogicalPath <= previousPath || row.ByteLength < 0 || row.ByteLength > 1_073_741_824 || !validSHA256(row.SHA256) {
			return nil, false, fmt.Errorf("invalid client asset manifest row %q", row.LogicalPath)
		}
		actual, ok := archiveAssets[row.LogicalPath]
		if !ok || actual != row {
			return nil, false, fmt.Errorf("client asset manifest mismatch for %q", row.LogicalPath)
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
		return nil, false, fmt.Errorf("client asset manifest omits %q", paths[0])
	}
	manifestDigest := sha256.Sum256(manifestBytes)
	registryBytes, err := fs.ReadFile(files, clientSupportRegistryPath)
	if err != nil {
		return nil, false, fmt.Errorf("read client support registry: %w", err)
	}
	var registry clientSupportRegistry
	if err := decodeCanonicalJSON(registryBytes, 1_048_576, &registry); err != nil {
		return nil, false, fmt.Errorf("validate client support registry: %w", err)
	}
	if registry.SchemaID != "cartulary.client_extension_support_registry.v1" || registry.ClientBuildClass != "standard" || registry.AssetSetSHA256 != hex.EncodeToString(manifestDigest[:]) || registry.ClientBuildID != "cartulary.web.standard.sha256:"+registry.AssetSetSHA256 || registry.Profiles == nil || len(registry.Profiles) > 256 {
		return nil, false, errors.New("client support registry root is invalid")
	}
	previousProfileID := ""
	for _, profile := range registry.Profiles {
		if !validToken(profile.ProfileID) || profile.ProfileID <= previousProfileID || len(profile.SupportedContractMajors) != 1 || profile.SupportedContractMajors[0] <= 0 || profile.WorkspaceKeys == nil || profile.CapabilityIDs == nil || len(profile.CapabilityIDs) != 0 || profile.PublicSchemaIDs == nil || !sortedUniqueStrings(profile.WorkspaceKeys, true) || !sortedUniqueStrings(profile.PublicSchemaIDs, false) {
			return nil, false, fmt.Errorf("invalid client support profile %q", profile.ProfileID)
		}
		previousProfileID = profile.ProfileID
	}
	return append([]byte(nil), registryBytes...), true, nil
}

func decodeCanonicalJSON(data []byte, maximum int, destination any) error {
	if len(data) == 0 || len(data) > maximum || !bytes.HasSuffix(data, []byte("\n")) || bytes.HasSuffix(data, []byte("\n\n")) {
		return errors.New("canonical JSON byte envelope is invalid")
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(destination); err != nil {
		return err
	}
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		if err == nil {
			return errors.New("canonical JSON contains multiple values")
		}
		return fmt.Errorf("canonical JSON trailing bytes are invalid: %w", err)
	}
	canonical, err := json.Marshal(destination)
	if err != nil {
		return err
	}
	canonical = append(canonical, '\n')
	if !bytes.Equal(canonical, data) {
		return errors.New("JSON bytes are not canonical")
	}
	return nil
}

func validLogicalPath(value string) bool {
	if value == "" || len(value) > 512 || strings.HasPrefix(value, "/") || strings.ContainsAny(value, "\\\x00") || path.Clean(value) != value {
		return false
	}
	for _, segment := range strings.Split(value, "/") {
		if segment == "" || segment == "." || segment == ".." {
			return false
		}
	}
	return true
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

func openDistArchive(files fs.FS) (*zip.Reader, bool, error) {
	data, err := fs.ReadFile(files, distArchivePath)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, false, nil
		}
		return nil, false, err
	}
	reader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, false, err
	}
	return reader, true, nil
}
