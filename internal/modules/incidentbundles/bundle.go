package incidentbundles

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"path"
	"sort"
	"strings"
	"time"

	"github.com/JochiRaider/cartulary/internal/platform/config"
)

const (
	BundleFormat  = "cartulary.incident_bundle"
	BundleVersion = "v1"
)

var requiredStructuredFiles = []string{
	"data/incident.json",
	"actors.ndjson",
	"records.ndjson",
	"timeline_events.ndjson",
	"hosts.ndjson",
	"identities.ndjson",
	"entity_aliases.ndjson",
	"indicators.ndjson",
	"indicator_observations.ndjson",
	"indicator_state_intervals.ndjson",
	"artifacts.ndjson",
	"task_requests.ndjson",
	"decisions.ndjson",
	"evidence_records.ndjson",
	"evidence_custody_events.ndjson",
	"object_blobs.ndjson",
	"entity_mentions.ndjson",
	"compromise_assessments.ndjson",
	"record_links.ndjson",
	"tags.ndjson",
	"record_tags.ndjson",
	"change_sets.ndjson",
	"change_set_mutations.ndjson",
	"record_revisions.ndjson",
	"saved_views.ndjson",
	"reference_pack_refs.json",
}

type ManifestInput struct {
	BundleID             string
	IncidentID           string
	IncidentKey          string
	ExportedAt           string
	ReferencePackMode    string
	OptionalSections     []string
	RequiredCapabilities []string
}

type BundleArchive struct {
	Bytes          []byte
	ManifestSHA256 string
	ChecksumLines  []string
	Manifest       BundleManifest
}

type BundleManifest struct {
	BundleFormat         string         `json:"bundle_format"`
	BundleVersion        string         `json:"bundle_version"`
	BundleID             string         `json:"bundle_id"`
	IncidentID           string         `json:"incident_id"`
	IncidentKey          string         `json:"incident_key"`
	ExportedAt           string         `json:"exported_at"`
	SourceBoundary       string         `json:"source_boundary"`
	HistoryMode          string         `json:"history_mode"`
	BlobMode             string         `json:"blob_mode"`
	ReferencePackMode    string         `json:"reference_pack_mode"`
	OptionalSections     []string       `json:"optional_sections"`
	RequiredCapabilities []string       `json:"required_capabilities"`
	Files                []ManifestFile `json:"files"`
}

type ManifestFile struct {
	Path      string `json:"path"`
	SHA256    string `json:"sha256"`
	SizeBytes int64  `json:"size_bytes"`
	Required  bool   `json:"required"`
}

type VerificationInput struct {
	Bundle []byte
	Limits config.LimitConfig
}

type VerifiedBundle struct {
	Manifest       BundleManifest
	ManifestSHA256 string
	Files          map[string][]byte
	Checksums      map[string]string
}

type VerificationError struct {
	ReasonCode string
}

func (e *VerificationError) Error() string {
	return "incident bundle verification failed: " + e.ReasonCode
}

func BuildBundleArchive(input ManifestInput, files map[string][]byte) (BundleArchive, error) {
	normalizedFiles := map[string][]byte{}
	for pathName, content := range files {
		if !safeBundlePath(pathName) {
			return BundleArchive{}, fmt.Errorf("unsafe bundle path %q", pathName)
		}
		normalizedFiles[pathName] = append([]byte(nil), content...)
	}
	for _, pathName := range requiredStructuredFiles {
		if _, ok := normalizedFiles[pathName]; !ok {
			if strings.HasSuffix(pathName, ".json") {
				normalizedFiles[pathName] = []byte("[]\n")
			} else {
				normalizedFiles[pathName] = []byte{}
			}
		}
	}
	if _, ok := normalizedFiles["data/incident.json"]; !ok {
		return BundleArchive{}, fmt.Errorf("data/incident.json is required")
	}

	manifestFiles := manifestFilesFor(normalizedFiles, false)
	sourceBoundaryBytes, err := json.Marshal(manifestFiles)
	if err != nil {
		return BundleArchive{}, err
	}
	manifest := BundleManifest{
		BundleFormat:         BundleFormat,
		BundleVersion:        BundleVersion,
		BundleID:             input.BundleID,
		IncidentID:           input.IncidentID,
		IncidentKey:          input.IncidentKey,
		ExportedAt:           input.ExportedAt,
		SourceBoundary:       "cartulary.source_boundary.v1:" + hashHex(sourceBoundaryBytes),
		HistoryMode:          HistoryModeFull,
		BlobMode:             BlobModeFull,
		ReferencePackMode:    input.ReferencePackMode,
		OptionalSections:     canonicalStringSet(input.OptionalSections),
		RequiredCapabilities: canonicalStringSet(input.RequiredCapabilities),
		Files:                manifestFiles,
	}
	if manifest.ReferencePackMode == "" {
		manifest.ReferencePackMode = ReferencePackModeRefsOnly
	}
	manifestBytes, err := canonicalJSONString(manifest)
	if err != nil {
		return BundleArchive{}, err
	}
	manifestSHA := hashHex(manifestBytes)
	normalizedFiles["manifest.json"] = manifestBytes

	checksumLines := checksumLinesFor(normalizedFiles)
	normalizedFiles["integrity/checksums.sha256"] = []byte(strings.Join(checksumLines, "\n") + "\n")

	archiveBytes, err := zipFiles(normalizedFiles)
	if err != nil {
		return BundleArchive{}, err
	}
	return BundleArchive{
		Bytes:          archiveBytes,
		ManifestSHA256: manifestSHA,
		ChecksumLines:  checksumLines,
		Manifest:       manifest,
	}, nil
}

func VerifyBundle(input VerificationInput) (VerifiedBundle, error) {
	files, err := readBundleArchive(input.Bundle, input.Limits)
	if err != nil {
		return VerifiedBundle{}, err
	}
	manifestBytes, ok := files["manifest.json"]
	if !ok {
		return VerifiedBundle{}, &VerificationError{ReasonCode: "missing_required_file"}
	}
	var manifest BundleManifest
	if err := json.Unmarshal(manifestBytes, &manifest); err != nil {
		return VerifiedBundle{}, &VerificationError{ReasonCode: "malformed_manifest"}
	}
	if manifest.BundleFormat != BundleFormat || manifest.BundleVersion != BundleVersion || manifest.HistoryMode != HistoryModeFull || manifest.BlobMode != BlobModeFull {
		return VerifiedBundle{}, &VerificationError{ReasonCode: "malformed_manifest"}
	}
	if len(manifest.RequiredCapabilities) > 0 {
		return VerifiedBundle{}, &VerificationError{ReasonCode: "unsupported_required_capability"}
	}
	for _, pathName := range requiredStructuredFiles {
		if _, ok := files[pathName]; !ok {
			return VerifiedBundle{}, &VerificationError{ReasonCode: "missing_required_file"}
		}
	}
	checksumBytes, ok := files["integrity/checksums.sha256"]
	if !ok {
		return VerifiedBundle{}, &VerificationError{ReasonCode: "checksum_mismatch"}
	}
	checksums, err := parseChecksumInventory(string(checksumBytes))
	if err != nil {
		return VerifiedBundle{}, &VerificationError{ReasonCode: "checksum_mismatch"}
	}
	for pathName, content := range files {
		if strings.HasPrefix(pathName, "integrity/") {
			continue
		}
		want, ok := checksums[pathName]
		if !ok {
			return VerifiedBundle{}, &VerificationError{ReasonCode: "checksum_mismatch"}
		}
		if got := hashHex(content); got != want {
			return VerifiedBundle{}, &VerificationError{ReasonCode: "checksum_mismatch"}
		}
		if strings.HasPrefix(pathName, "blobs/sha256/") {
			if blobSHA := strings.TrimPrefix(pathName, "blobs/sha256/"); blobSHA != want {
				return VerifiedBundle{}, &VerificationError{ReasonCode: "blob_hash_mismatch"}
			}
		}
	}
	for listedPath := range checksums {
		if strings.HasPrefix(listedPath, "integrity/") {
			return VerifiedBundle{}, &VerificationError{ReasonCode: "checksum_mismatch"}
		}
		if _, ok := files[listedPath]; !ok {
			return VerifiedBundle{}, &VerificationError{ReasonCode: "missing_required_file"}
		}
	}
	return VerifiedBundle{
		Manifest:       manifest,
		ManifestSHA256: hashHex(manifestBytes),
		Files:          files,
		Checksums:      checksums,
	}, nil
}

func readBundleArchive(bundle []byte, limits config.LimitConfig) (map[string][]byte, error) {
	if len(bundle) == 0 {
		return nil, &VerificationError{ReasonCode: "missing_required_file"}
	}
	if files, err := readZipArchive(bundle, limits); err == nil {
		return files, nil
	} else if isVerificationFailure(err) {
		return nil, err
	}
	if files, err := readTarArchive(bytes.NewReader(bundle), int64(len(bundle)), limits); err == nil {
		return files, nil
	} else if isVerificationFailure(err) {
		return nil, err
	}
	gz, err := gzip.NewReader(bytes.NewReader(bundle))
	if err == nil {
		defer gz.Close()
		if files, tarErr := readTarArchive(gz, int64(len(bundle)), limits); tarErr == nil {
			return files, nil
		} else if isVerificationFailure(tarErr) {
			return nil, tarErr
		}
	}
	return nil, &VerificationError{ReasonCode: "unsupported_member_type"}
}

func readZipArchive(bundle []byte, limits config.LimitConfig) (map[string][]byte, error) {
	zr, err := zip.NewReader(bytes.NewReader(bundle), int64(len(bundle)))
	if err != nil {
		return nil, err
	}
	files := map[string][]byte{}
	var extracted int64
	for _, member := range zr.File {
		if err := checkMemberCount(len(files)+1, limits); err != nil {
			return nil, err
		}
		if !safeBundlePath(member.Name) {
			return nil, &VerificationError{ReasonCode: "invalid_member_path"}
		}
		if member.FileInfo().IsDir() || member.FileInfo().Mode().Type() != 0 {
			return nil, &VerificationError{ReasonCode: "unsupported_member_type"}
		}
		extracted += int64(member.UncompressedSize64)
		if err := checkExtractedSize(extracted, limits); err != nil {
			return nil, err
		}
		rc, err := member.Open()
		if err != nil {
			return nil, &VerificationError{ReasonCode: "missing_required_file"}
		}
		data, err := io.ReadAll(rc)
		_ = rc.Close()
		if err != nil {
			return nil, &VerificationError{ReasonCode: "missing_required_file"}
		}
		files[member.Name] = data
	}
	if err := checkCompressionRatio(extracted, int64(len(bundle)), limits); err != nil {
		return nil, err
	}
	return files, nil
}

func readTarArchive(reader io.Reader, compressedSize int64, limits config.LimitConfig) (map[string][]byte, error) {
	tr := tar.NewReader(reader)
	files := map[string][]byte{}
	var extracted int64
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if err := checkMemberCount(len(files)+1, limits); err != nil {
			return nil, err
		}
		if !safeBundlePath(header.Name) {
			return nil, &VerificationError{ReasonCode: "invalid_member_path"}
		}
		if header.Typeflag != tar.TypeReg {
			return nil, &VerificationError{ReasonCode: "unsupported_member_type"}
		}
		extracted += header.Size
		if err := checkExtractedSize(extracted, limits); err != nil {
			return nil, err
		}
		data, err := io.ReadAll(tr)
		if err != nil {
			return nil, &VerificationError{ReasonCode: "missing_required_file"}
		}
		files[header.Name] = data
	}
	if len(files) == 0 {
		return nil, &VerificationError{ReasonCode: "missing_required_file"}
	}
	if err := checkCompressionRatio(extracted, compressedSize, limits); err != nil {
		return nil, err
	}
	return files, nil
}

func checkExtractedSize(extracted int64, limits config.LimitConfig) error {
	max := limits.IncidentBundles.MaxExtractedBytes
	if max <= 0 {
		max = config.DefaultIncidentBundleMaxExtractedBytes
	}
	if extracted > max {
		return &VerificationError{ReasonCode: "archive_extracted_bytes_exceeded"}
	}
	return nil
}

func checkMemberCount(count int, limits config.LimitConfig) error {
	max := limits.Archives.MaxMembers
	if max <= 0 {
		max = config.DefaultArchiveMaxMembers
	}
	if int64(count) > max {
		return &VerificationError{ReasonCode: "archive_member_count_exceeded"}
	}
	return nil
}

func checkCompressionRatio(extracted int64, compressed int64, limits config.LimitConfig) error {
	max := limits.Archives.MaxCompressionRatio
	if max <= 0 {
		max = config.DefaultArchiveMaxCompressionRatio
	}
	if compressed <= 0 {
		compressed = 1
	}
	if extracted/compressed > max {
		return &VerificationError{ReasonCode: "archive_compression_ratio_exceeded"}
	}
	return nil
}

func manifestFilesFor(files map[string][]byte, includeIntegrity bool) []ManifestFile {
	paths := make([]string, 0, len(files))
	for pathName := range files {
		if pathName == "manifest.json" || (!includeIntegrity && strings.HasPrefix(pathName, "integrity/")) {
			continue
		}
		paths = append(paths, pathName)
	}
	sort.Strings(paths)
	result := make([]ManifestFile, 0, len(paths))
	for _, pathName := range paths {
		result = append(result, ManifestFile{
			Path:      pathName,
			SHA256:    "sha256:" + hashHex(files[pathName]),
			SizeBytes: int64(len(files[pathName])),
			Required:  !strings.HasPrefix(pathName, "ext/"),
		})
	}
	return result
}

func checksumLinesFor(files map[string][]byte) []string {
	paths := make([]string, 0, len(files))
	for pathName := range files {
		if strings.HasPrefix(pathName, "integrity/") {
			continue
		}
		paths = append(paths, pathName)
	}
	sort.Strings(paths)
	lines := make([]string, 0, len(paths))
	for _, pathName := range paths {
		lines = append(lines, hashHex(files[pathName])+"  "+pathName)
	}
	return lines
}

func parseChecksumInventory(content string) (map[string]string, error) {
	lines := strings.Split(strings.TrimSpace(content), "\n")
	if len(lines) == 1 && lines[0] == "" {
		return nil, fmt.Errorf("empty checksum inventory")
	}
	checksums := map[string]string{}
	previous := ""
	for _, line := range lines {
		parts := strings.SplitN(line, "  ", 2)
		if len(parts) != 2 || len(parts[0]) != sha256.Size*2 {
			return nil, fmt.Errorf("malformed checksum line")
		}
		if _, err := hex.DecodeString(parts[0]); err != nil {
			return nil, err
		}
		if previous != "" && parts[1] < previous {
			return nil, fmt.Errorf("checksum inventory not sorted")
		}
		previous = parts[1]
		checksums[parts[1]] = parts[0]
	}
	return checksums, nil
}

func zipFiles(files map[string][]byte) ([]byte, error) {
	paths := make([]string, 0, len(files))
	for pathName := range files {
		paths = append(paths, pathName)
	}
	sort.Strings(paths)
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	modified := time.Date(1980, 1, 1, 0, 0, 0, 0, time.UTC)
	for _, pathName := range paths {
		header := &zip.FileHeader{
			Name:     pathName,
			Method:   zip.Store,
			Modified: modified,
		}
		header.SetMode(0o600)
		w, err := zw.CreateHeader(header)
		if err != nil {
			return nil, err
		}
		if _, err := w.Write(files[pathName]); err != nil {
			return nil, err
		}
	}
	if err := zw.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func safeBundlePath(pathName string) bool {
	if pathName == "" || strings.HasPrefix(pathName, "/") || strings.Contains(pathName, "\\") {
		return false
	}
	clean := path.Clean(pathName)
	if clean == "." || clean != pathName {
		return false
	}
	for _, segment := range strings.Split(pathName, "/") {
		if segment == "" || segment == "." || segment == ".." {
			return false
		}
	}
	return true
}

func canonicalStringSet(values []string) []string {
	seen := map[string]struct{}{}
	for _, value := range values {
		if value != "" {
			seen[value] = struct{}{}
		}
	}
	result := make([]string, 0, len(seen))
	for value := range seen {
		result = append(result, value)
	}
	sort.Strings(result)
	return result
}

func isVerificationFailure(err error) bool {
	_, ok := err.(*VerificationError)
	return ok
}
