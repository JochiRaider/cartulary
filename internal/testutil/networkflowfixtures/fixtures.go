// Package networkflowfixtures admits immutable fixture bytes and retains comparison
// evidence. Product behavior and scenario dispatch belong to the consuming tests.
package networkflowfixtures

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"syscall"

	"github.com/JochiRaider/cartulary/internal/testutil/harnessschema"
	"github.com/JochiRaider/cartulary/internal/testutil/suiteservices"
)

const summarySchema = "cartulary.network_flow_fixture_execution.v1"
const manifestSchema = "cartulary.network_flow_fixture_manifest.v2"
const scenarioSchema = "cartulary.network_flow_fixture_scenario.v2"
const maxManifestBytes = 1024 * 1024
const maxArtifacts = 64

var fixtureIDPattern = regexp.MustCompile(`^NF-FIX-[0-9]{3}-[a-z0-9][a-z0-9-]*$`)
var artifactNamePattern = regexp.MustCompile(`^[A-Za-z0-9._@+-]+$`)
var ErrDraft = errors.New("draft fixture cannot qualify behavior")

type File struct {
	Path   string `json:"logical_path"`
	Size   int64  `json:"size_bytes"`
	SHA256 string `json:"sha256"`
}

type Manifest struct {
	FixtureID string `json:"fixture_id"`
	Freeze    struct {
		Status   string `json:"status"`
		Revision int    `json:"revision"`
	} `json:"freeze"`
	Sources        []File `json:"source_files"`
	Expected       []File `json:"expected_artifacts"`
	Transcripts    []File `json:"transcript_files"`
	SourceSHA256   string `json:"source_bundle_sha256"`
	ExpectedSHA256 string `json:"expected_bundle_sha256"`
}

// Config requires the revision selected by the executable fixture registry.
// RunRoot is the already selected invocation root, never a global scratch path.
type Config struct {
	CorpusRoot string
	RunRoot    string
	FixtureID  string
	Revision   int
}

type Summary struct {
	SchemaID       string  `json:"schema_id"`
	FixtureID      string  `json:"fixture_id"`
	Revision       *int    `json:"freeze_revision"`
	ManifestSHA256 *string `json:"manifest_sha256"`
	SourceSHA256   *string `json:"source_bundle_sha256"`
	ExpectedSHA256 *string `json:"expected_bundle_sha256"`
	InputRoot      *string `json:"materialized_input_root"`
	Artifacts      []File  `json:"produced_artifacts"`
	Status         string  `json:"comparison_status"`
	FailurePhase   *string `json:"failure_phase"`
}

type Bundle struct {
	Manifest   Manifest
	inputRoot  string
	outputRoot string
	runRoot    string
	files      map[string]File
	artifacts  []File
}

// SelectedRunRoot resolves the public harness environment without inspecting
// documentation or starting services. An explicit run identity is mandatory.
func SelectedRunRoot() (string, error) {
	if os.Getenv("CARTULARY_TEST_RUN_ID") == "" {
		return "", errors.New("fixture execution requires an explicit run ID")
	}
	root, err := suiteservices.ResolveResultsRoot(nil)
	if err != nil {
		return "", err
	}
	id := suiteservices.ResolveRunID(nil)
	if !safePath(id) || strings.Contains(id, "/") {
		return "", errors.New("invalid run ID")
	}
	return filepath.Join(root, id), nil
}

// Run admits every byte before invoking compare. A failing callback, including
// testing.Fatal's Goexit, retains a failed summary. Evidence publication failures
// are returned alongside comparison failures and never qualify as success.
func Run(config Config, compare func(*Bundle) error) (summaryPath string, err error) {
	if !fixtureIDPattern.MatchString(config.FixtureID) || len(config.FixtureID) > 120 || config.Revision < 1 || compare == nil {
		return "", errors.New("invalid fixture execution configuration")
	}
	if err := secureDirectory(config.RunRoot); err != nil {
		return "", err
	}
	workspace, err := os.MkdirTemp(config.RunRoot, config.FixtureID+"-")
	if err != nil {
		return "", err
	}
	summaryPath = filepath.Join(workspace, "comparison-summary.json")
	phase := "admission"
	summary := Summary{SchemaID: summarySchema, FixtureID: config.FixtureID, Artifacts: []File{}, Status: "failed", FailurePhase: &phase}
	var bundle *Bundle
	defer func() {
		if bundle != nil {
			summary.Artifacts = bundle.artifacts
		}
		data, publishErr := json.MarshalIndent(summary, "", "  ")
		if publishErr == nil {
			publishErr = validateDocument(data, summarySchema)
		}
		if publishErr == nil {
			publishErr = os.WriteFile(summaryPath, append(data, '\n'), 0600)
		}
		err = errors.Join(err, publishErr)
	}()
	bundle, err = admit(config, workspace, &summary)
	if err != nil {
		if errors.Is(err, ErrDraft) {
			summary.Status = "blocked"
		}
		return summaryPath, err
	}
	phase = "comparison"
	if err = compare(bundle); err != nil {
		return summaryPath, err
	}
	summary.Status, summary.FailurePhase = "passed", nil
	return summaryPath, nil
}

func admit(config Config, workspace string, summary *Summary) (*Bundle, error) {
	dir := filepath.Join(config.CorpusRoot, config.FixtureID)
	if err := rejectSymlinkComponents(dir); err != nil {
		return nil, err
	}
	root, err := os.OpenRoot(dir)
	if err != nil {
		return nil, err
	}
	defer root.Close()
	manifestBytes, err := readRegular(root, "manifest.json", maxManifestBytes)
	if err != nil {
		return nil, err
	}
	summary.ManifestSHA256 = ptr(digest(manifestBytes))
	if err := validateDocument(manifestBytes, manifestSchema); err != nil {
		return nil, fmt.Errorf("manifest: %w", err)
	}
	var manifest Manifest
	if err := json.Unmarshal(manifestBytes, &manifest); err != nil {
		return nil, err
	}
	summary.Revision = &manifest.Freeze.Revision
	summary.SourceSHA256, summary.ExpectedSHA256 = &manifest.SourceSHA256, &manifest.ExpectedSHA256
	if manifest.FixtureID != config.FixtureID || manifest.Freeze.Revision != config.Revision {
		return nil, errors.New("fixture identity or frozen revision mismatch")
	}
	if manifest.Freeze.Status != "frozen" {
		return nil, ErrDraft
	}
	listed := map[string]File{}
	for _, files := range [][]File{manifest.Sources, manifest.Expected, manifest.Transcripts} {
		previous := ""
		for _, file := range files {
			if !safePath(file.Path) || file.Path == "manifest.json" || file.Path <= previous {
				return nil, errors.New("invalid or unordered fixture path")
			}
			if _, duplicate := listed[file.Path]; duplicate {
				return nil, errors.New("duplicate fixture path")
			}
			listed[file.Path], previous = file, file.Path
		}
	}
	if bundleDigest(manifest.Sources) != manifest.SourceSHA256 || bundleDigest(append(slices.Clone(manifest.Expected), manifest.Transcripts...)) != manifest.ExpectedSHA256 {
		return nil, errors.New("fixture bundle digest mismatch")
	}
	if err := checkInventory(root, listed); err != nil {
		return nil, err
	}
	inputs, outputs := filepath.Join(workspace, "inputs"), filepath.Join(workspace, "produced")
	if err := os.Mkdir(inputs, 0700); err != nil {
		return nil, err
	}
	if err := os.Mkdir(outputs, 0700); err != nil {
		return nil, err
	}
	// Copy before hashing: the callback only sees the exact bytes verified here,
	// even if the original fixture is replaced after admission.
	for _, file := range listed {
		if err := copyVerified(root, file, inputs); err != nil {
			return nil, err
		}
		if strings.HasSuffix(file.Path, ".scenario.json") {
			data, err := os.ReadFile(filepath.Join(inputs, file.Path))
			if err != nil {
				return nil, err
			}
			if err := validateDocument(data, scenarioSchema); err != nil {
				return nil, fmt.Errorf("scenario: %w", err)
			}
			var scenario struct {
				FixtureID string `json:"fixture_id"`
			}
			if err := json.Unmarshal(data, &scenario); err != nil {
				return nil, err
			}
			if scenario.FixtureID != manifest.FixtureID {
				return nil, errors.New("scenario identity mismatch")
			}
		}
	}
	if err := filepath.WalkDir(inputs, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return os.Chmod(path, 0500)
		}
		return os.Chmod(path, 0400)
	}); err != nil {
		return nil, err
	}
	relative, err := filepath.Rel(config.RunRoot, inputs)
	if err != nil {
		return nil, err
	}
	summary.InputRoot = &relative
	return &Bundle{Manifest: manifest, inputRoot: inputs, outputRoot: outputs, runRoot: config.RunRoot, files: listed, artifacts: []File{}}, nil
}

func (b *Bundle) Read(path string) ([]byte, error) {
	input, err := b.InputPath(path)
	if err != nil {
		return nil, err
	}
	return os.ReadFile(input)
}

// InputPath admits file-based product readers to the same sealed bytes as Read.
// It exposes only listed inputs, never the corpus or writable output directory.
func (b *Bundle) InputPath(path string) (string, error) {
	if _, ok := b.files[path]; !ok {
		return "", errors.New("unlisted fixture input")
	}
	return filepath.Join(b.inputRoot, path), nil
}

// Record retains bounded artifact metadata; it never interprets product output.
func (b *Bundle) Record(name string, data []byte) error {
	if !safePath(name) || !artifactNamePattern.MatchString(name) || len(name) > 120 || len(b.artifacts) >= maxArtifacts {
		return errors.New("invalid or excessive produced artifacts")
	}
	path := filepath.Join(b.outputRoot, name)
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return err
	}
	_, writeErr := file.Write(data)
	if err := errors.Join(writeErr, file.Close()); err != nil {
		return err
	}
	relative, err := filepath.Rel(b.runRoot, path)
	if err != nil {
		return err
	}
	b.artifacts = append(b.artifacts, File{Path: relative, Size: int64(len(data)), SHA256: digest(data)})
	return nil
}

func validateDocument(data []byte, schema string) error {
	value, err := harnessschema.Decode(data)
	if err != nil {
		return err
	}
	return harnessschema.Validate(schema, value)
}

func safePath(path string) bool {
	return path != "" && fs.ValidPath(path) && path != "." && !strings.ContainsAny(path, "\\:\x00") && filepath.IsLocal(path)
}

func rejectSymlinkComponents(path string) error {
	abs, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	current := string(filepath.Separator)
	for _, part := range strings.Split(strings.TrimPrefix(abs, current), string(filepath.Separator)) {
		current = filepath.Join(current, part)
		info, err := os.Lstat(current)
		if err != nil {
			return err
		}
		if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
			return errors.New("unsafe fixture directory")
		}
	}
	return nil
}

func secureDirectory(path string) error {
	if !filepath.IsAbs(path) {
		return errors.New("fixture run root must be absolute")
	}
	current := string(filepath.Separator)
	for _, part := range strings.Split(strings.TrimPrefix(filepath.Clean(path), current), string(filepath.Separator)) {
		current = filepath.Join(current, part)
		info, err := os.Lstat(current)
		if errors.Is(err, os.ErrNotExist) {
			if err := os.Mkdir(current, 0700); err != nil && !errors.Is(err, os.ErrExist) {
				return err
			}
			info, err = os.Lstat(current)
		}
		if err != nil {
			return err
		}
		if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
			return errors.New("unsafe run directory")
		}
	}
	return nil
}

func openRegular(root *os.Root, path string) (*os.File, error) {
	if !safePath(path) {
		return nil, errors.New("unsafe fixture path")
	}
	parts := strings.Split(path, "/")
	for i := range parts {
		info, err := root.Lstat(strings.Join(parts[:i+1], "/"))
		if err != nil {
			return nil, err
		}
		if info.Mode()&os.ModeSymlink != 0 || i < len(parts)-1 && !info.IsDir() || i == len(parts)-1 && !info.Mode().IsRegular() {
			return nil, errors.New("fixture path must contain only directories and a regular file")
		}
	}
	file, err := root.OpenFile(path, os.O_RDONLY|syscall.O_NOFOLLOW, 0)
	if err != nil {
		return nil, err
	}
	info, err := file.Stat()
	if err != nil || !info.Mode().IsRegular() {
		_ = file.Close()
		return nil, errors.New("nonregular fixture file")
	}
	return file, nil
}

func readRegular(root *os.Root, path string, limit int64) ([]byte, error) {
	file, err := openRegular(root, path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	data, err := io.ReadAll(io.LimitReader(file, limit+1))
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > limit {
		return nil, errors.New("fixture document exceeds admission bound")
	}
	return data, nil
}

func checkInventory(root *os.Root, listed map[string]File) error {
	seen := map[string]bool{}
	err := fs.WalkDir(root.FS(), ".", func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return errors.New("symlink in fixture inventory")
		}
		if entry.IsDir() {
			return nil
		}
		if !entry.Type().IsRegular() {
			return errors.New("nonregular fixture inventory entry")
		}
		if path == "manifest.json" {
			return nil
		}
		if _, ok := listed[path]; !ok {
			return errors.New("unlisted fixture file")
		}
		seen[path] = true
		return nil
	})
	if err != nil {
		return err
	}
	if len(seen) != len(listed) {
		return errors.New("missing fixture file")
	}
	return nil
}

func copyVerified(root *os.Root, expected File, inputs string) error {
	source, err := openRegular(root, expected.Path)
	if err != nil {
		return err
	}
	defer source.Close()
	destination := filepath.Join(inputs, expected.Path)
	if err := os.MkdirAll(filepath.Dir(destination), 0700); err != nil {
		return err
	}
	out, err := os.OpenFile(destination, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return err
	}
	hash := sha256.New()
	n, copyErr := io.Copy(io.MultiWriter(out, hash), io.LimitReader(source, expected.Size+1))
	if err := errors.Join(copyErr, out.Close()); err != nil {
		return err
	}
	if n != expected.Size || hex.EncodeToString(hash.Sum(nil)) != expected.SHA256 {
		return errors.New("fixture size or hash mismatch")
	}
	return nil
}

func bundleDigest(files []File) string {
	hash := sha256.New()
	for _, file := range files {
		fmt.Fprintf(hash, "%s\x00%s\x00%d\n", file.Path, file.SHA256, file.Size)
	}
	return hex.EncodeToString(hash.Sum(nil))
}
func digest(data []byte) string { sum := sha256.Sum256(data); return hex.EncodeToString(sum[:]) }
func ptr[T any](value T) *T     { return &value }
