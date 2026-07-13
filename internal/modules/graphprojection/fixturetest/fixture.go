// Package fixturetest provides test-only loading and byte comparison for the
// authored Graph Projection conformance corpus. It intentionally contains no
// projection derivation or expected-value construction.
package fixturetest

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

const ManifestSchemaID = "cartulary.graph_projection_fixture_manifest.v1"

type Manifest struct {
	SchemaID       string      `json:"schema_id"`
	FixtureVersion int         `json:"fixture_version"`
	FixtureID      string      `json:"fixture_id"`
	ExecutionLayer string      `json:"execution_layer"`
	OwnerSections  []string    `json:"owner_sections"`
	AcceptanceIDs  []string    `json:"acceptance_ids"`
	TestSymbol     string      `json:"test_symbol"`
	Determinism    Determinism `json:"determinism"`
	Steps          []Step      `json:"steps"`
	Artifacts      []Artifact  `json:"artifacts"`
	Golden         Golden      `json:"golden"`
}

type Determinism struct {
	Clock        string `json:"clock"`
	Nonce        string `json:"nonce"`
	CursorKeyHex string `json:"cursor_key_hex"`
}

type Step struct {
	ID               string `json:"id"`
	Operation        string `json:"operation"`
	InputArtifact    string `json:"input_artifact"`
	ExpectedArtifact string `json:"expected_artifact"`
}

type Artifact struct {
	Path   string `json:"path"`
	Role   string `json:"role"`
	SHA256 string `json:"sha256"`
}

type Golden struct {
	Provenance   string `json:"provenance"`
	ReviewStatus string `json:"review_status"`
}

func RepoRoot(start string) (string, error) {
	current, err := filepath.Abs(start)
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(current, "go.mod")); err == nil {
			return current, nil
		}
		next := filepath.Dir(current)
		if next == current {
			return "", fmt.Errorf("fixturetest: repository root not found from %s", start)
		}
		current = next
	}
}

func Load(root, fixtureID string) (Manifest, string, error) {
	if !strings.HasPrefix(fixtureID, "GP-FIX-") || len(fixtureID) != len("GP-FIX-000") {
		return Manifest{}, "", fmt.Errorf("fixturetest: invalid fixture id %q", fixtureID)
	}
	directory := filepath.Join(root, "contracts", "graph-projection", "fixtures", fixtureID)
	manifestPath := filepath.Join(directory, "fixture.json")
	body, err := os.ReadFile(manifestPath)
	if err != nil {
		return Manifest{}, "", err
	}
	decoder := json.NewDecoder(strings.NewReader(string(body)))
	decoder.DisallowUnknownFields()
	var manifest Manifest
	if err := decoder.Decode(&manifest); err != nil {
		return Manifest{}, "", fmt.Errorf("decode %s: %w", manifestPath, err)
	}
	if manifest.SchemaID != ManifestSchemaID || manifest.FixtureVersion != 1 || manifest.FixtureID != fixtureID {
		return Manifest{}, "", fmt.Errorf("invalid fixture identity in %s", manifestPath)
	}
	if len(manifest.Steps) == 0 || len(manifest.Artifacts) == 0 || manifest.TestSymbol == "" {
		return Manifest{}, "", fmt.Errorf("fixture %s is incomplete", fixtureID)
	}
	seen := map[string]bool{}
	for _, artifact := range manifest.Artifacts {
		if artifact.Path == "" || seen[artifact.Path] {
			return Manifest{}, "", fmt.Errorf("fixture %s has duplicate or empty artifact path", fixtureID)
		}
		seen[artifact.Path] = true
		path, err := containedPath(directory, artifact.Path)
		if err != nil {
			return Manifest{}, "", err
		}
		info, err := os.Lstat(path)
		if err != nil {
			return Manifest{}, "", err
		}
		if info.Mode()&fs.ModeSymlink != 0 || !info.Mode().IsRegular() {
			return Manifest{}, "", fmt.Errorf("fixture %s artifact %s is not a regular file", fixtureID, artifact.Path)
		}
		bytes, err := os.ReadFile(path)
		if err != nil {
			return Manifest{}, "", err
		}
		digest := sha256.Sum256(bytes)
		if artifact.SHA256 != hex.EncodeToString(digest[:]) {
			return Manifest{}, "", fmt.Errorf("fixture %s artifact digest mismatch for %s", fixtureID, artifact.Path)
		}
	}
	return manifest, directory, nil
}

func ReadArtifact(directory, relativePath string) ([]byte, error) {
	path, err := containedPath(directory, relativePath)
	if err != nil {
		return nil, err
	}
	return os.ReadFile(path)
}

func CompareBytes(name string, actual, expected []byte) error {
	if string(actual) == string(expected) {
		return nil
	}
	return fmt.Errorf("%s differs: actual_len=%d expected_len=%d", name, len(actual), len(expected))
}

func containedPath(directory, relativePath string) (string, error) {
	if relativePath == "" || filepath.IsAbs(relativePath) || strings.Contains(relativePath, "\\") {
		return "", fmt.Errorf("unsafe fixture artifact path %q", relativePath)
	}
	path := filepath.Clean(filepath.Join(directory, filepath.FromSlash(relativePath)))
	rel, err := filepath.Rel(directory, path)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("unsafe fixture artifact path %q", relativePath)
	}
	return path, nil
}
