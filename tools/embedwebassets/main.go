package main

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
)

var stableZipTime = time.Date(1980, 1, 1, 0, 0, 0, 0, time.UTC)

type archiveEntry struct {
	absPath string
	name    string
	isDir   bool
}

type clientAssetManifest struct {
	Assets   []clientAssetManifestRow `json:"assets"`
	SchemaID string                   `json:"schema_id"`
}

type clientAssetManifestRow struct {
	ByteLength  int64  `json:"byte_length"`
	LogicalPath string `json:"logical_path"`
	SHA256      string `json:"sha256"`
}

type clientSupportSource struct {
	SchemaID         string                   `json:"schema_id"`
	ClientBuildClass string                   `json:"client_build_class"`
	Rows             []clientSupportSourceRow `json:"rows"`
}

type clientSupportSourceRow struct {
	ProfileID        string   `json:"profile_id"`
	ContractMajor    int64    `json:"contract_major"`
	WorkspaceKeys    []string `json:"workspace_keys"`
	CapabilityIDs    []string `json:"capability_ids"`
	PublicSchemaIDs  []string `json:"public_schema_ids"`
	ClientAssetSetID string   `json:"client_asset_set_id"`
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

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
}

func run(args []string) error {
	var sourceDir string
	var sourceIndex string
	var output string
	var tempDir string
	var assetManifestOutput string
	var clientSupportSourcePath string
	var clientSupportRegistryOutput string
	var integrityOutput, integrityDigestOutput, repositoryRoot, bootstrapOutput string

	flags := flag.NewFlagSet("embedwebassets", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	flags.StringVar(&sourceDir, "source-dir", "", "frontend dist directory")
	flags.StringVar(&sourceIndex, "source-index", "", "frontend dist index.html")
	flags.StringVar(&output, "output", "", "output zip path")
	flags.StringVar(&tempDir, "temp-dir", "", "temporary archive directory")
	flags.StringVar(&assetManifestOutput, "asset-manifest", "", "client asset-set manifest output")
	flags.StringVar(&clientSupportSourcePath, "client-support-source", "", "authored client support source")
	flags.StringVar(&clientSupportRegistryOutput, "client-support-registry", "", "build-bound client support registry output")
	flags.StringVar(&bootstrapOutput, "bootstrap-template", "", "private browser bootstrap template output")
	flags.StringVar(&integrityOutput, "registry-integrity", "", "final Extensions integrity output")
	flags.StringVar(&integrityDigestOutput, "registry-integrity-digest", "", "final Extensions integrity pin output")
	flags.StringVar(&repositoryRoot, "repository-root", ".", "repository source root")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if sourceDir == "" {
		return errors.New("--source-dir is required")
	}
	if sourceIndex == "" {
		return errors.New("--source-index is required")
	}
	if output == "" {
		return errors.New("--output is required")
	}
	if assetManifestOutput == "" {
		return errors.New("--asset-manifest is required")
	}
	if clientSupportSourcePath == "" {
		return errors.New("--client-support-source is required")
	}
	if clientSupportRegistryOutput == "" {
		return errors.New("--client-support-registry is required")
	}
	if integrityOutput == "" || integrityDigestOutput == "" || bootstrapOutput == "" {
		return errors.New("final registry integrity outputs are required")
	}
	destinations := map[string]bool{}
	for _, destination := range []string{output, assetManifestOutput, clientSupportRegistryOutput, integrityOutput, integrityDigestOutput, bootstrapOutput} {
		root, err := filepath.Abs(sourceDir)
		if err != nil {
			return err
		}
		target, err := filepath.Abs(destination)
		if err != nil {
			return err
		}
		if destinations[target] {
			return errors.New("packaged output paths must be distinct")
		}
		destinations[target] = true
		relative, err := filepath.Rel(root, target)
		if err != nil {
			return err
		}
		if relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator)) && !filepath.IsAbs(relative) {
			return errors.New("packaged outputs must be outside the served asset tree")
		}
	}
	if flags.NArg() != 0 {
		return fmt.Errorf("unexpected positional arguments: %s", strings.Join(flags.Args(), " "))
	}
	stage, err := os.MkdirTemp(tempDir, ".browser-package-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(stage)
	staticArchive := filepath.Join(stage, "static.zip")
	if err := writeArchive(sourceDir, sourceIndex, staticArchive, stage); err != nil {
		return err
	}
	if err := writeBootstrapTemplate(sourceIndex, bootstrapOutput); err != nil {
		return err
	}
	if err := writeClientBuildContracts(staticArchive, bootstrapOutput, clientSupportSourcePath, assetManifestOutput, clientSupportRegistryOutput); err != nil {
		return err
	}
	if err := writeExtensionIntegrity(repositoryRoot, assetManifestOutput, clientSupportRegistryOutput, bootstrapOutput, integrityOutput, integrityDigestOutput); err != nil {
		return err
	}
	return sealBrowserArchive(staticArchive, output, map[string]string{
		"client-asset-set-manifest.json": assetManifestOutput, "client-extension-support-registry.json": clientSupportRegistryOutput, "client-bootstrap-template.json": bootstrapOutput, "extension-registry-integrity.json": integrityOutput, "extension-registry-integrity.sha256": integrityDigestOutput,
	})
}

func writeClientBuildContracts(archivePath, bootstrapPath, supportSourcePath, manifestOutput, supportOutput string) error {
	archive, err := zip.OpenReader(archivePath)
	if err != nil {
		return err
	}
	defer archive.Close()
	rows := make([]clientAssetManifestRow, 0, len(archive.File))
	seen := map[string]bool{}
	for _, entry := range archive.File {
		if entry.FileInfo().IsDir() {
			continue
		}
		if !validAssetPath(entry.Name) || !strings.HasPrefix(entry.Name, "assets/") || !entry.Mode().IsRegular() || seen[entry.Name] || entry.UncompressedSize64 > 1<<30 || len(rows) >= 65536 {
			return errors.New("invalid or oversized packaged client asset")
		}
		seen[entry.Name] = true
		reader, err := entry.Open()
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
		if length > 1<<30 || uint64(length) != entry.UncompressedSize64 {
			return errors.New("client asset length mismatch")
		}
		rows = append(rows, clientAssetManifestRow{ByteLength: length, LogicalPath: entry.Name, SHA256: hex.EncodeToString(hash.Sum(nil))})
	}
	sort.Slice(rows, func(i, j int) bool { return rows[i].LogicalPath < rows[j].LogicalPath })
	if len(rows) == 0 || len(rows) > 65_536 {
		return errors.New("client asset manifest must contain 1..65536 rows")
	}
	manifestBytes, err := canonicalJSONBytes(clientAssetManifest{
		Assets: rows, SchemaID: "cartulary.client_asset_set_manifest.v1",
	})
	if err != nil {
		return err
	}
	if len(manifestBytes) > 16_777_216 {
		return errors.New("client asset manifest exceeds byte limit")
	}
	manifestDigest := sha256.Sum256(manifestBytes)
	bootstrap, err := os.ReadFile(bootstrapPath)
	if err != nil {
		return err
	}
	buildDigest := byteDigest([]byte(hex.EncodeToString(manifestDigest[:]) + "\n" + byteDigest(bootstrap) + "\n"))

	sourceFile, err := os.Open(supportSourcePath)
	if err != nil {
		return fmt.Errorf("read client support source: %w", err)
	}
	defer sourceFile.Close()
	sourceBytes, err := io.ReadAll(io.LimitReader(sourceFile, (64<<20)+1))
	if err != nil {
		return err
	}
	if len(sourceBytes) > 64<<20 {
		return errors.New("client support source exceeds byte limit")
	}
	if _, err := canonicaljson.DecodeStrict(sourceBytes); err != nil {
		return err
	}
	var source clientSupportSource
	decoder := json.NewDecoder(bytes.NewReader(sourceBytes))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&source); err != nil {
		return fmt.Errorf("decode client support source: %w", err)
	}
	var extra any
	if trailingErr := decoder.Decode(&extra); !errors.Is(trailingErr, io.EOF) {
		if trailingErr == nil {
			return errors.New("client support source must contain one JSON value")
		}
		return fmt.Errorf("decode trailing client support source bytes: %w", trailingErr)
	}
	if source.SchemaID != "cartulary.extension_client_support_registry_source.v1" || source.ClientBuildClass != "standard" || source.Rows == nil || len(source.Rows) > 256 {
		return errors.New("client support source has invalid root contract")
	}
	profiles := make([]clientSupportRegistryProfile, 0, len(source.Rows))
	previousProfileID := ""
	for _, row := range source.Rows {
		if !validExtensionToken(row.ProfileID) || row.ProfileID <= previousProfileID || row.ContractMajor <= 0 || row.WorkspaceKeys == nil || row.CapabilityIDs == nil || row.PublicSchemaIDs == nil || row.ClientAssetSetID == "" {
			return fmt.Errorf("invalid client support row for %q", row.ProfileID)
		}
		if len(row.CapabilityIDs) != 0 || !sortedUnique(row.WorkspaceKeys, true) || !sortedUnique(row.PublicSchemaIDs, false) {
			return fmt.Errorf("invalid client support collections for %s", row.ProfileID)
		}
		previousProfileID = row.ProfileID
		profiles = append(profiles, clientSupportRegistryProfile{
			CapabilityIDs: []string{}, ProfileID: row.ProfileID, PublicSchemaIDs: row.PublicSchemaIDs,
			SupportedContractMajors: []int64{row.ContractMajor}, WorkspaceKeys: row.WorkspaceKeys,
		})
	}
	digestHex := hex.EncodeToString(manifestDigest[:])
	supportBytes, err := canonicalJSONBytes(clientSupportRegistry{
		AssetSetSHA256:   digestHex,
		ClientBuildClass: "standard",
		ClientBuildID:    "cartulary.web.standard.sha256:" + buildDigest,
		Profiles:         profiles,
		SchemaID:         "cartulary.client_extension_support_registry.v1",
	})
	if err != nil {
		return err
	}
	if len(supportBytes) > 1<<20 {
		return errors.New("client support registry exceeds byte limit")
	}
	if err := writeCanonicalOutput(manifestOutput, manifestBytes); err != nil {
		return err
	}
	return writeCanonicalOutput(supportOutput, supportBytes)
}

func canonicalJSONBytes(value any) ([]byte, error) { return canonicaljson.MarshalExtension(value) }

func validAssetPath(value string) bool {
	return value != "." && len(value) <= 512 && utf8.ValidString(value) && fs.ValidPath(value) && !strings.ContainsAny(value, "\\\x00")
}

func writeCanonicalOutput(output string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(output), 0o755); err != nil {
		return fmt.Errorf("create client contract output directory: %w", err)
	}
	temp, err := os.CreateTemp(filepath.Dir(output), "."+filepath.Base(output)+".tmp-*")
	if err != nil {
		return err
	}
	tempPath := temp.Name()
	defer os.Remove(tempPath)
	if _, err := temp.Write(data); err != nil {
		_ = temp.Close()
		return err
	}
	if err := temp.Chmod(0o644); err != nil {
		_ = temp.Close()
		return err
	}
	if err := temp.Close(); err != nil {
		return err
	}
	return os.Rename(tempPath, output)
}

func validExtensionToken(value string) bool {
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

func sortedUnique(values []string, tokens bool) bool {
	previous := ""
	for _, value := range values {
		if value == "" || value <= previous || (tokens && !validExtensionToken(value)) {
			return false
		}
		previous = value
	}
	return true
}

func writeArchive(sourceDir, sourceIndex, output string, tempDir string) error {
	sourceRoot, err := filepath.Abs(sourceDir)
	if err != nil {
		return fmt.Errorf("resolve source dir: %w", err)
	}
	sourceInfo, err := os.Lstat(sourceRoot)
	if err != nil {
		return fmt.Errorf("stat source dir: %w", err)
	}
	if !sourceInfo.IsDir() || sourceInfo.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("source dir must be a real directory: %s", sourceDir)
	}

	indexPath, err := filepath.Abs(sourceIndex)
	if err != nil {
		return fmt.Errorf("resolve source index: %w", err)
	}
	indexRel, err := filepath.Rel(sourceRoot, indexPath)
	if err != nil || indexRel != "index.html" || strings.HasPrefix(indexRel, ".."+string(filepath.Separator)) || filepath.IsAbs(indexRel) {
		return fmt.Errorf("source index must be inside source dir: %s", sourceIndex)
	}
	indexInfo, err := os.Lstat(indexPath)
	if err != nil {
		return fmt.Errorf("stat source index: %w", err)
	}
	if !indexInfo.Mode().IsRegular() || indexInfo.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("source index must be a regular file: %s", sourceIndex)
	}

	entries, err := collectEntries(sourceRoot)
	if err != nil {
		return err
	}
	staticEntries := entries[:0]
	for _, entry := range entries {
		if entry.name == "index.html" {
			continue
		}
		if entry.name != "assets/" && !strings.HasPrefix(entry.name, "assets/") {
			return errors.New("browser static files must be under the served assets namespace")
		}
		staticEntries = append(staticEntries, entry)
	}
	entries = staticEntries
	if len(entries) == 0 {
		return fmt.Errorf("source dir contains no embeddable files: %s", sourceDir)
	}

	outputPath, err := filepath.Abs(output)
	if err != nil {
		return fmt.Errorf("resolve output: %w", err)
	}
	outputDir := filepath.Dir(outputPath)
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return fmt.Errorf("create output dir: %w", err)
	}
	tempRoot := outputDir
	if tempDir != "" {
		tempRoot, err = filepath.Abs(tempDir)
		if err != nil {
			return fmt.Errorf("resolve temp dir: %w", err)
		}
	}
	if err := os.MkdirAll(tempRoot, 0o755); err != nil {
		return fmt.Errorf("create temp dir: %w", err)
	}
	tempFile, err := os.CreateTemp(tempRoot, "."+filepath.Base(outputPath)+".tmp-*")
	if err != nil {
		return fmt.Errorf("create temp archive: %w", err)
	}
	tempPath := tempFile.Name()
	cleanupTemp := true
	defer func() {
		if cleanupTemp {
			_ = os.Remove(tempPath)
		}
	}()

	if err := writeZip(tempFile, entries); err != nil {
		_ = tempFile.Close()
		return err
	}
	if err := tempFile.Close(); err != nil {
		return fmt.Errorf("close temp archive: %w", err)
	}
	if err := os.Chmod(tempPath, 0o644); err != nil {
		return fmt.Errorf("set archive mode: %w", err)
	}
	if err := os.Rename(tempPath, outputPath); err != nil {
		return fmt.Errorf("publish archive: %w", err)
	}
	cleanupTemp = false
	return nil
}

func collectEntries(sourceRoot string) ([]archiveEntry, error) {
	var entries []archiveEntry
	if err := filepath.WalkDir(sourceRoot, func(current string, dirEntry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if current == sourceRoot {
			return nil
		}
		info, err := dirEntry.Info()
		if err != nil {
			return fmt.Errorf("stat source entry %s: %w", current, err)
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("source entry must not be a symlink: %s", current)
		}
		if !info.IsDir() && !info.Mode().IsRegular() {
			return fmt.Errorf("source entry must be a regular file or directory: %s", current)
		}
		name, err := archiveName(sourceRoot, current, info.IsDir())
		if err != nil {
			return err
		}
		entries = append(entries, archiveEntry{
			absPath: current,
			name:    name,
			isDir:   info.IsDir(),
		})
		return nil
	}); err != nil {
		return nil, err
	}
	sort.Slice(entries, func(left, right int) bool {
		return entries[left].name < entries[right].name
	})
	return entries, nil
}

func archiveName(sourceRoot, current string, isDir bool) (string, error) {
	rel, err := filepath.Rel(sourceRoot, current)
	if err != nil {
		return "", fmt.Errorf("resolve source-relative path: %w", err)
	}
	if rel == "." || rel == "" {
		return "", fmt.Errorf("invalid source-relative path: %s", current)
	}
	name := filepath.ToSlash(rel)
	clean := path.Clean(name)
	if clean == "." || clean == "" || strings.HasPrefix(clean, "../") || path.IsAbs(clean) || clean != name {
		return "", fmt.Errorf("unsafe archive path: %s", name)
	}
	if isDir {
		clean += "/"
	}
	return clean, nil
}

func writeZip(writer io.Writer, entries []archiveEntry) error {
	zipWriter := zip.NewWriter(writer)
	for _, entry := range entries {
		header := &zip.FileHeader{
			Name:     entry.name,
			Method:   zip.Deflate,
			Modified: stableZipTime,
		}
		if entry.isDir {
			header.Method = zip.Store
			header.SetMode(0o755 | fs.ModeDir)
		} else {
			header.SetMode(0o644)
		}
		zipEntry, err := zipWriter.CreateHeader(header)
		if err != nil {
			_ = zipWriter.Close()
			return fmt.Errorf("create zip entry %s: %w", entry.name, err)
		}
		if entry.isDir {
			continue
		}
		file, err := os.Open(entry.absPath)
		if err != nil {
			_ = zipWriter.Close()
			return fmt.Errorf("open source file %s: %w", entry.absPath, err)
		}
		if _, err := io.Copy(zipEntry, file); err != nil {
			_ = file.Close()
			_ = zipWriter.Close()
			return fmt.Errorf("copy source file %s: %w", entry.absPath, err)
		}
		if err := file.Close(); err != nil {
			_ = zipWriter.Close()
			return fmt.Errorf("close source file %s: %w", entry.absPath, err)
		}
	}
	if err := zipWriter.Close(); err != nil {
		return fmt.Errorf("close zip archive: %w", err)
	}
	return nil
}

// The bootstrap is a dynamic-response input, never a static asset.
func writeBootstrapTemplate(source, output string) error {
	f, err := os.Open(source)
	if err != nil {
		return err
	}
	defer f.Close()
	body, err := io.ReadAll(io.LimitReader(f, (64<<20)+1))
	if err != nil {
		return err
	}
	if !utf8.Valid(body) || bytes.Count(body, []byte("</head>")) != 1 {
		return errors.New("browser bootstrap must be UTF-8 with one closing head element")
	}
	contract, err := canonicalJSONBytes(map[string]any{"schema_id": "cartulary.client_bootstrap_template.v1", "html_utf8": string(body)})
	if err != nil {
		return err
	}
	if len(contract) > 64<<20 {
		return errors.New("browser bootstrap contract exceeds byte limit")
	}
	return writeCanonicalOutput(output, contract)
}

// Publish one complete container atomically. Contract entries are private and
// excluded from both the asset manifest and the runtime static filesystem.
func sealBrowserArchive(staticPath, output string, contracts map[string]string) error {
	source, err := zip.OpenReader(staticPath)
	if err != nil {
		return err
	}
	defer source.Close()
	if err := os.MkdirAll(filepath.Dir(output), 0755); err != nil {
		return err
	}
	temporary, err := os.CreateTemp(filepath.Dir(output), ".browser-package-*")
	if err != nil {
		return err
	}
	defer os.Remove(temporary.Name())
	defer temporary.Close()
	writer := zip.NewWriter(temporary)
	for _, file := range source.File {
		if err := writer.Copy(file); err != nil {
			return err
		}
	}
	names := make([]string, 0, len(contracts))
	for name := range contracts {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		body, err := os.ReadFile(contracts[name])
		if err != nil {
			return err
		}
		header := &zip.FileHeader{Name: "_contracts/" + name, Method: zip.Deflate}
		header.Modified = stableZipTime
		header.SetMode(0444)
		file, err := writer.CreateHeader(header)
		if err != nil {
			return err
		}
		if _, err := file.Write(body); err != nil {
			return err
		}
	}
	if err := writer.Close(); err != nil {
		return err
	}
	if err := temporary.Chmod(0644); err != nil {
		return err
	}
	if err := temporary.Close(); err != nil {
		return err
	}
	return os.Rename(temporary.Name(), output)
}
