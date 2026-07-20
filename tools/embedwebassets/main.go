package main

import (
	"archive/zip"
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

	flags := flag.NewFlagSet("embedwebassets", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	flags.StringVar(&sourceDir, "source-dir", "", "frontend dist directory")
	flags.StringVar(&sourceIndex, "source-index", "", "frontend dist index.html")
	flags.StringVar(&output, "output", "", "output zip path")
	flags.StringVar(&tempDir, "temp-dir", "", "temporary archive directory")
	flags.StringVar(&assetManifestOutput, "asset-manifest", "", "client asset-set manifest output")
	flags.StringVar(&clientSupportSourcePath, "client-support-source", "", "authored client support source")
	flags.StringVar(&clientSupportRegistryOutput, "client-support-registry", "", "build-bound client support registry output")
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
	if flags.NArg() != 0 {
		return fmt.Errorf("unexpected positional arguments: %s", strings.Join(flags.Args(), " "))
	}
	if err := writeArchive(sourceDir, sourceIndex, output, tempDir); err != nil {
		return err
	}
	return writeClientBuildContracts(sourceDir, clientSupportSourcePath, assetManifestOutput, clientSupportRegistryOutput)
}

func writeClientBuildContracts(sourceDir, supportSourcePath, manifestOutput, supportOutput string) error {
	sourceRoot, err := filepath.Abs(sourceDir)
	if err != nil {
		return fmt.Errorf("resolve source dir for client contracts: %w", err)
	}
	entries, err := collectEntries(sourceRoot)
	if err != nil {
		return err
	}
	rows := make([]clientAssetManifestRow, 0, len(entries))
	for _, entry := range entries {
		if entry.isDir {
			continue
		}
		data, readErr := os.ReadFile(entry.absPath)
		if readErr != nil {
			return fmt.Errorf("read client asset %s: %w", entry.name, readErr)
		}
		if len(data) > 1_073_741_824 {
			return fmt.Errorf("client asset exceeds byte limit: %s", entry.name)
		}
		digest := sha256.Sum256(data)
		rows = append(rows, clientAssetManifestRow{
			ByteLength: int64(len(data)), LogicalPath: entry.name, SHA256: hex.EncodeToString(digest[:]),
		})
	}
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

	sourceFile, err := os.Open(supportSourcePath)
	if err != nil {
		return fmt.Errorf("read client support source: %w", err)
	}
	defer sourceFile.Close()
	var source clientSupportSource
	decoder := json.NewDecoder(sourceFile)
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
		ClientBuildID:    "cartulary.web.standard.sha256:" + digestHex,
		Profiles:         profiles,
		SchemaID:         "cartulary.client_extension_support_registry.v1",
	})
	if err != nil {
		return err
	}
	if len(supportBytes) > 1_048_576 {
		return errors.New("client support registry exceeds byte limit")
	}
	if err := writeCanonicalOutput(manifestOutput, manifestBytes); err != nil {
		return err
	}
	return writeCanonicalOutput(supportOutput, supportBytes)
}

func canonicalJSONBytes(value any) ([]byte, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return nil, fmt.Errorf("encode canonical client contract: %w", err)
	}
	return append(data, '\n'), nil
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
	if err != nil || indexRel == "." || strings.HasPrefix(indexRel, ".."+string(filepath.Separator)) || filepath.IsAbs(indexRel) {
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
