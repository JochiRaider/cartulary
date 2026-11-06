package irfixtures

import (
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func ownedWorkspace(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	t.Cleanup(func() {
		_ = filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
			if err == nil && entry.IsDir() {
				return os.Chmod(path, 0700)
			}
			return err
		})
	})
	return root
}

func TestIRCorpusExpectations(t *testing.T) {
	root, err := filepath.Abs("../../../fixtures")
	if err != nil {
		t.Fatal(err)
	}
	c := ForTest(t, root)
	got, err := BuildExpectations(c)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := c.Read("ir-incident/v1/expected.json")
	if err != nil {
		t.Fatal(err)
	}
	var want Expectations
	if err := decode(raw, &want); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, want) {
		actual, _ := json.MarshalIndent(got, "", "  ")
		t.Fatalf("independent source expectations differ; revise only after review:\n%s", actual)
	}
	t.Logf("admitted %s revision %d manifest SHA-256 %s", ID, Revision, c.ManifestSHA256)
}

func TestIRCorpusAdmission(t *testing.T) {
	mutations := map[string]func(string){
		"changed original": func(root string) {
			_ = os.WriteFile(filepath.Join(root, "ir-incident/v1/hosts.csv"), []byte("changed"), 0600)
		},
		"missing source": func(root string) { _ = os.Remove(filepath.Join(root, "ir-incident/v1/aliases.csv")) },
		"changed expectation": func(root string) {
			_ = os.WriteFile(filepath.Join(root, "ir-incident/v1/expected.json"), []byte("{}"), 0600)
		},
		"changed dependency": func(root string) {
			_ = os.WriteFile(filepath.Join(root, "network-flow-orientation/v1/day1.csv"), []byte("changed"), 0600)
		},
		"changed dependency manifest": func(root string) {
			_ = os.WriteFile(filepath.Join(root, "network-flow-orientation/v1/manifest.json"), []byte("{}"), 0600)
		},
		"symlink": func(root string) {
			path := filepath.Join(root, "ir-incident/v1/hosts.csv")
			_ = os.Remove(path)
			_ = os.Symlink("identities.csv", path)
		},
		"path traversal": func(root string) {
			mutateManifest(t, root, func(m *Manifest) { m.Sources[0].Path = "../../outside.csv" })
		},
		"duplicate member": func(root string) { mutateManifest(t, root, func(m *Manifest) { m.Sources[1] = m.Sources[0] }) },
		"old revision":     func(root string) { mutateManifest(t, root, func(m *Manifest) { m.Revision = 0 }) },
	}
	for name, mutate := range mutations {
		t.Run(name, func(t *testing.T) {
			workspace := ownedWorkspace(t)
			root := filepath.Join(workspace, "fixtures")
			for _, name := range []string{"ir-incident", "network-flow-orientation"} {
				if err := os.CopyFS(filepath.Join(root, name), os.DirFS(filepath.Join("../../../fixtures", name))); err != nil {
					t.Fatal(err)
				}
			}
			mutate(root)
			destination := filepath.Join(workspace, "inputs")
			if _, err := Admit(root, destination); err == nil {
				t.Fatal("invalid fixture admitted")
			}
			if _, err := os.Stat(destination); !os.IsNotExist(err) {
				t.Fatalf("failed admission left materialized inputs: %v", err)
			}
		})
	}
}

func mutateManifest(t *testing.T, root string, mutate func(*Manifest)) {
	t.Helper()
	path := filepath.Join(root, "ir-incident/v1/manifest.json")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var m Manifest
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatal(err)
	}
	mutate(&m)
	raw, err = json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, raw, 0600); err != nil {
		t.Fatal(err)
	}
}
