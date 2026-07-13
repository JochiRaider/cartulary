// Package fixturetest provides test-only loading and byte comparison for the
// authored Graph Projection conformance corpus. It intentionally contains no
// projection derivation or expected-value construction.
package fixturetest

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
	"reflect"
	"strings"
)

const ManifestSchemaID = "cartulary.graph_projection_fixture_manifest.v1"

type Manifest struct {
	SchemaID       string       `json:"schema_id"`
	FixtureVersion int          `json:"fixture_version"`
	FixtureID      string       `json:"fixture_id"`
	ExecutionLayer string       `json:"execution_layer"`
	OwnerSections  []string     `json:"owner_sections"`
	AcceptanceIDs  []string     `json:"acceptance_ids"`
	TestSymbol     string       `json:"test_symbol"`
	Determinism    Determinism  `json:"determinism"`
	Comparison     Comparison   `json:"comparison"`
	StateEffects   StateEffects `json:"state_effects"`
	Steps          []Step       `json:"steps"`
	Artifacts      []Artifact   `json:"artifacts"`
	Golden         Golden       `json:"golden"`
}

type Determinism struct {
	Clock        string `json:"clock"`
	Nonce        string `json:"nonce"`
	CursorKeyHex string `json:"cursor_key_hex"`
}

type Comparison struct {
	Scope                    string `json:"scope"`
	Mode                     string `json:"mode"`
	RunIndependentValidation bool   `json:"run_independent_validation"`
}

type StateEffects struct {
	Mode string `json:"mode"`
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

// StepExecution is the observable output of one manifest operation. The
// executor owns Graph behavior; fixturetest owns only artifact comparison and
// retained-state-effect enforcement.
type StepExecution struct {
	Artifact        []byte
	StateEffectMode string
}

type Executor interface {
	ExecuteFixtureStep(manifest Manifest, step Step, input []byte) (StepExecution, error)
}

type ExecutorFunc func(manifest Manifest, step Step, input []byte) (StepExecution, error)

func (f ExecutorFunc) ExecuteFixtureStep(manifest Manifest, step Step, input []byte) (StepExecution, error) {
	return f(manifest, step, input)
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
	fixtureRoot, err := os.OpenRoot(directory)
	if err != nil {
		return Manifest{}, "", err
	}
	defer fixtureRoot.Close()
	body, err := fixtureRoot.ReadFile("fixture.json")
	if err != nil {
		return Manifest{}, "", err
	}
	manifestPath := filepath.Join(directory, "fixture.json")
	decoder := json.NewDecoder(strings.NewReader(string(body)))
	decoder.DisallowUnknownFields()
	var manifest Manifest
	if err := decoder.Decode(&manifest); err != nil {
		return Manifest{}, "", fmt.Errorf("decode %s: %w", manifestPath, err)
	}
	if manifest.SchemaID != ManifestSchemaID || manifest.FixtureVersion != 1 || manifest.FixtureID != fixtureID {
		return Manifest{}, "", fmt.Errorf("invalid fixture identity in %s", manifestPath)
	}
	if len(manifest.AcceptanceIDs) == 0 || len(manifest.Steps) == 0 || len(manifest.Artifacts) == 0 || manifest.TestSymbol == "" {
		return Manifest{}, "", fmt.Errorf("fixture %s is incomplete", fixtureID)
	}
	if manifest.Comparison.Mode != "exact_artifacts" || (manifest.Comparison.Scope != "run_specific" && manifest.Comparison.Scope != "run_independent") {
		return Manifest{}, "", fmt.Errorf("fixture %s has invalid comparison contract", fixtureID)
	}
	if manifest.Comparison.Scope == "run_specific" && manifest.Comparison.RunIndependentValidation {
		return Manifest{}, "", fmt.Errorf("fixture %s cannot exclude issue ids from run-specific comparison", fixtureID)
	}
	if manifest.StateEffects.Mode != "no_retained_state_change" && manifest.StateEffects.Mode != "retained_state_change" {
		return Manifest{}, "", fmt.Errorf("fixture %s has invalid state effect", fixtureID)
	}
	seen := map[string]bool{}
	for _, artifact := range manifest.Artifacts {
		if artifact.Path == "" || seen[artifact.Path] {
			return Manifest{}, "", fmt.Errorf("fixture %s has duplicate or empty artifact path", fixtureID)
		}
		seen[artifact.Path] = true
		path, err := cleanRelativePath(artifact.Path)
		if err != nil {
			return Manifest{}, "", err
		}
		info, err := fixtureRoot.Lstat(path)
		if err != nil {
			return Manifest{}, "", err
		}
		if info.Mode()&fs.ModeSymlink != 0 || !info.Mode().IsRegular() {
			return Manifest{}, "", fmt.Errorf("fixture %s artifact %s is not a regular file", fixtureID, artifact.Path)
		}
		bytes, err := fixtureRoot.ReadFile(path)
		if err != nil {
			return Manifest{}, "", err
		}
		digest := sha256.Sum256(bytes)
		if artifact.SHA256 != hex.EncodeToString(digest[:]) {
			return Manifest{}, "", fmt.Errorf("fixture %s artifact digest mismatch for %s", fixtureID, artifact.Path)
		}
	}
	for _, step := range manifest.Steps {
		if step.Operation == "advance_clock" {
			continue
		}
		if step.ExpectedArtifact == "" || !seen[step.ExpectedArtifact] {
			return Manifest{}, "", fmt.Errorf("fixture %s step %s has no declared expected artifact", fixtureID, step.ID)
		}
		if step.InputArtifact != "" && !seen[step.InputArtifact] {
			return Manifest{}, "", fmt.Errorf("fixture %s step %s has no declared input artifact", fixtureID, step.ID)
		}
	}
	return manifest, directory, nil
}

// Verify loads a reviewed manifest, executes its ordered steps through the
// supplied real-operation adapter, and compares every declared expected
// artifact. It never derives or rewrites an expected value.
func Verify(root, fixtureID string, executor Executor) error {
	if executor == nil {
		return fmt.Errorf("fixturetest: nil executor")
	}
	manifest, directory, err := Load(root, fixtureID)
	if err != nil {
		return err
	}
	if manifest.Golden.ReviewStatus != "reviewed" {
		return fmt.Errorf("fixture %s golden is not reviewed", fixtureID)
	}
	for _, step := range manifest.Steps {
		var input []byte
		if step.InputArtifact != "" {
			input, err = ReadArtifact(directory, step.InputArtifact)
			if err != nil {
				return fmt.Errorf("fixture %s step %s input: %w", fixtureID, step.ID, err)
			}
		}
		execution, err := executor.ExecuteFixtureStep(manifest, step, input)
		if err != nil {
			return fmt.Errorf("fixture %s step %s: %w", fixtureID, step.ID, err)
		}
		if execution.StateEffectMode != manifest.StateEffects.Mode {
			return fmt.Errorf("fixture %s step %s state effect = %q, expected %q", fixtureID, step.ID, execution.StateEffectMode, manifest.StateEffects.Mode)
		}
		if step.Operation == "advance_clock" {
			continue
		}
		expected, err := ReadArtifact(directory, step.ExpectedArtifact)
		if err != nil {
			return fmt.Errorf("fixture %s step %s expected: %w", fixtureID, step.ID, err)
		}
		if err := CompareArtifact(step.ExpectedArtifact, execution.Artifact, expected); err != nil {
			return fmt.Errorf("fixture %s step %s: %w", fixtureID, step.ID, err)
		}
	}
	return nil
}

func CompareArtifact(name string, actual, expected []byte) error {
	if !strings.HasSuffix(name, ".json") {
		return CompareBytes(name, actual, expected)
	}
	actualValue, err := decodeJSONArtifact(actual)
	if err != nil {
		return fmt.Errorf("%s actual JSON: %w", name, err)
	}
	expectedValue, err := decodeJSONArtifact(expected)
	if err != nil {
		return fmt.Errorf("%s expected JSON: %w", name, err)
	}
	if reflect.DeepEqual(actualValue, expectedValue) {
		return nil
	}
	actualCompact, _ := json.Marshal(actualValue)
	expectedCompact, _ := json.Marshal(expectedValue)
	return CompareBytes(name, actualCompact, expectedCompact)
}

func decodeJSONArtifact(body []byte) (any, error) {
	decoder := json.NewDecoder(strings.NewReader(string(body)))
	decoder.UseNumber()
	var value any
	if err := decoder.Decode(&value); err != nil {
		return nil, err
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		if err == nil {
			return nil, fmt.Errorf("multiple JSON values")
		}
		return nil, err
	}
	return value, nil
}

func ReadArtifact(directory, relativePath string) ([]byte, error) {
	path, err := cleanRelativePath(relativePath)
	if err != nil {
		return nil, err
	}
	fixtureRoot, err := os.OpenRoot(directory)
	if err != nil {
		return nil, err
	}
	defer fixtureRoot.Close()
	return fixtureRoot.ReadFile(path)
}

func CompareBytes(name string, actual, expected []byte) error {
	if string(actual) == string(expected) {
		return nil
	}
	limit := len(actual)
	if len(expected) < limit {
		limit = len(expected)
	}
	first := 0
	for first < limit && actual[first] == expected[first] {
		first++
	}
	windowStart := first - 16
	if windowStart < 0 {
		windowStart = 0
	}
	actualEnd := first + 16
	if actualEnd > len(actual) {
		actualEnd = len(actual)
	}
	expectedEnd := first + 16
	if expectedEnd > len(expected) {
		expectedEnd = len(expected)
	}
	return fmt.Errorf("%s differs at byte %d: actual_len=%d expected_len=%d actual_window=%q expected_window=%q", name, first, len(actual), len(expected), actual[windowStart:actualEnd], expected[windowStart:expectedEnd])
}

func cleanRelativePath(relativePath string) (string, error) {
	if relativePath == "" || filepath.IsAbs(relativePath) || strings.Contains(relativePath, "\\") {
		return "", fmt.Errorf("unsafe fixture artifact path %q", relativePath)
	}
	path := filepath.Clean(filepath.FromSlash(relativePath))
	if path == "." || path == ".." || strings.HasPrefix(path, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("unsafe fixture artifact path %q", relativePath)
	}
	return path, nil
}
