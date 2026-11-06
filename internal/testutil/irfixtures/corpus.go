// Package irfixtures admits the incident-response corpus. It supplies data,
// never a workflow interpreter or an alternative product import path.
package irfixtures

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"slices"

	"github.com/JochiRaider/cartulary/internal/testutil/fixtureartifacts"
	"github.com/JochiRaider/cartulary/internal/testutil/harnessschema"
)

const ID = "ir-incident"
const Revision = 1
const maxFileBytes = 32 << 20

type Source struct {
	fixtureartifacts.File
	Records   int      `json:"records"`
	Columns   []string `json:"columns"`
	Use       string   `json:"use"`
	FieldKeys []string `json:"field_keys"`
}

type Dependency struct {
	ID             string   `json:"dependency_id"`
	Revision       int      `json:"revision"`
	ManifestPath   string   `json:"manifest_path"`
	ManifestSHA256 string   `json:"manifest_sha256"`
	Datasets       []string `json:"datasets"`
}

type Manifest struct {
	SchemaID     string                  `json:"schema_id"`
	FixtureID    string                  `json:"fixture_id"`
	Revision     int                     `json:"revision"`
	Sources      []Source                `json:"source_files"`
	Dependencies []Dependency            `json:"dependencies"`
	Expected     fixtureartifacts.File   `json:"expected_file"`
	Supplement   []fixtureartifacts.File `json:"supplement_files"`
}

type Corpus struct {
	Manifest       Manifest
	ManifestSHA256 string
	root           string
	files          map[string]fixtureartifacts.File
}

func decode(data []byte, result any) error {
	d := json.NewDecoder(bytes.NewReader(data))
	d.DisallowUnknownFields()
	if err := d.Decode(result); err != nil {
		return err
	}
	if d.Decode(new(any)) != io.EOF {
		return errors.New("trailing fixture JSON")
	}
	return nil
}

func digest(data []byte) string {
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}

// Admit copies and verifies every selected input before returning. fixtureRoot
// is the repository's fixtures directory; destination must be a fresh directory
// inside the caller's test-owned workspace. Consumers only see sealed copies.
func Admit(fixtureRoot, destination string) (_ *Corpus, err error) {
	if err := fixtureartifacts.RejectSymlinkComponents(fixtureRoot); err != nil {
		return nil, err
	}
	if err := fixtureartifacts.SecureDirectory(filepath.Dir(destination)); err != nil {
		return nil, err
	}
	if err := os.Mkdir(destination, 0700); err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			_ = os.RemoveAll(destination)
		}
	}()
	root, err := os.OpenRoot(fixtureRoot)
	if err != nil {
		return nil, err
	}
	defer root.Close()
	raw, err := fixtureartifacts.ReadRegular(root, "ir-incident/v1/manifest.json", 1<<20)
	if err != nil {
		return nil, err
	}
	c := &Corpus{root: destination, ManifestSHA256: digest(raw), files: map[string]fixtureartifacts.File{}}
	value, err := harnessschema.Decode(raw)
	if err != nil {
		return nil, err
	}
	if err := harnessschema.Validate("cartulary.ir_fixture_manifest.v1", value); err != nil {
		return nil, err
	}
	if err := decode(raw, &c.Manifest); err != nil {
		return nil, err
	}
	m := c.Manifest
	if m.SchemaID != "cartulary.ir_fixture_manifest.v1" || m.FixtureID != ID || m.Revision != Revision || len(m.Sources) != 5 || len(m.Dependencies) != 1 {
		return nil, errors.New("incompatible IR corpus manifest")
	}
	add := func(prefix string, file fixtureartifacts.File) error {
		if !fixtureartifacts.SafePath(file.Path) || file.Size < 0 || file.Size > maxFileBytes || len(file.SHA256) != 64 {
			return errors.New("invalid IR fixture member")
		}
		if _, err := hex.DecodeString(file.SHA256); err != nil {
			return err
		}
		file.Path = prefix + file.Path
		if _, exists := c.files[file.Path]; exists {
			return errors.New("duplicate IR fixture member")
		}
		if err := fixtureartifacts.CopyVerified(root, file, destination); err != nil {
			return err
		}
		c.files[file.Path] = file
		return nil
	}
	for _, source := range m.Sources {
		if source.Records < 1 || len(source.Columns) == 0 {
			return nil, errors.New("invalid source shape")
		}
		switch source.Use {
		case "file_upload", "clipboard":
			if len(source.FieldKeys) != len(source.Columns) {
				return nil, errors.New("incomplete source mapping")
			}
		case "entity_alias_operations":
			if len(source.FieldKeys) != 0 {
				return nil, errors.New("aliases are not an import target")
			}
		default:
			return nil, errors.New("unsupported fixture use")
		}
		if err := add("ir-incident/v1/", source.File); err != nil {
			return nil, err
		}
	}
	if err := add("ir-incident/v1/", m.Expected); err != nil {
		return nil, err
	}
	for _, file := range m.Supplement {
		if err := add("ir-incident/v1/", file); err != nil {
			return nil, err
		}
	}
	dep := m.Dependencies[0]
	if dep.ID != "network-flow-orientation" || dep.Revision != 1 || dep.ManifestPath != "network-flow-orientation/v1/manifest.json" || !slices.Equal(dep.Datasets, []string{"day1", "day2", "overlap"}) {
		return nil, errors.New("unsupported fixture dependency")
	}
	dependencyBytes, err := fixtureartifacts.ReadRegular(root, dep.ManifestPath, 1<<20)
	if err != nil {
		return nil, err
	}
	if digest(dependencyBytes) != dep.ManifestSHA256 {
		return nil, errors.New("fixture dependency changed")
	}
	// Dependency schema is owned by Network Flow. Its exact admitted bytes pin
	// the complete descriptor; this consumer selects only named datasets.
	var orientation struct {
		SchemaID string `json:"schema_id"`
		Revision int    `json:"expectation_revision"`
		Datasets []struct {
			Name         string `json:"name"`
			CSV          string `json:"csv"`
			CSVHash      string `json:"csv_sha256"`
			Expected     string `json:"expected"`
			ExpectedHash string `json:"expected_sha256"`
		} `json:"datasets"`
	}
	if err := json.Unmarshal(dependencyBytes, &orientation); err != nil {
		return nil, err
	}
	if orientation.SchemaID != "cartulary.network_flow.orientation_fixture.v1" || orientation.Revision != dep.Revision {
		return nil, errors.New("incompatible orientation dependency")
	}
	found := map[string]bool{}
	for _, dataset := range orientation.Datasets {
		if !slices.Contains(dep.Datasets, dataset.Name) {
			continue
		}
		if found[dataset.Name] {
			return nil, errors.New("duplicate dependency dataset")
		}
		found[dataset.Name] = true
		for _, file := range []fixtureartifacts.File{{Path: dataset.CSV, SHA256: dataset.CSVHash}, {Path: dataset.Expected, SHA256: dataset.ExpectedHash}} {
			data, err := fixtureartifacts.ReadRegular(root, "network-flow-orientation/v1/"+file.Path, maxFileBytes)
			if err != nil {
				return nil, err
			}
			file.Size = int64(len(data))
			if err := add("network-flow-orientation/v1/", file); err != nil {
				return nil, err
			}
		}
	}
	if len(found) != len(dep.Datasets) {
		return nil, errors.New("missing dependency dataset")
	}
	for name, data := range map[string][]byte{"ir-incident/v1/manifest.json": raw, dep.ManifestPath: dependencyBytes} {
		if err := add("", fixtureartifacts.File{Path: name, Size: int64(len(data)), SHA256: digest(data)}); err != nil {
			return nil, err
		}
	}
	if err := filepath.WalkDir(destination, func(path string, entry fs.DirEntry, err error) error {
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
	return c, nil
}

func (c *Corpus) Path(name string) (string, error) {
	if _, exists := c.files[name]; !exists {
		return "", fmt.Errorf("unadmitted fixture member: %s", name)
	}
	return filepath.Join(c.root, name), nil
}

func (c *Corpus) Read(name string) ([]byte, error) {
	p, err := c.Path(name)
	if err != nil {
		return nil, err
	}
	return os.ReadFile(p)
}

// Remove releases only this admitted copy. Source fixtures remain immutable.
// Callers retaining comparison inputs under a run root need not call Remove.
func (c *Corpus) Remove() error {
	if err := fixtureartifacts.RejectSymlinkComponents(c.root); err != nil {
		return err
	}
	if err := filepath.WalkDir(c.root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return os.Chmod(path, 0700)
		}
		return nil
	}); err != nil {
		return err
	}
	return os.RemoveAll(c.root)
}
