package irfixtures

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/JochiRaider/cartulary/internal/testutil/fixtureartifacts"
)

// ForTest retains the exact admitted inputs and an assertion disposition under
// the already selected harness run. The harness row remains the test authority.
// Calling this helper during source discovery is neither needed nor supported.
func ForTest(t testing.TB, fixtureRoot string) *Corpus {
	t.Helper()
	runRoot, err := fixtureartifacts.SelectedRunRoot()
	if err != nil {
		t.Fatal(err)
	}
	if err := fixtureartifacts.SecureDirectory(runRoot); err != nil {
		t.Fatal(err)
	}
	workspace, err := os.MkdirTemp(runRoot, "ir-incident-")
	if err != nil {
		t.Fatal(err)
	}
	c, err := Admit(fixtureRoot, filepath.Join(workspace, "inputs"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		files := make([]fixtureartifacts.File, 0, len(c.files))
		for _, file := range c.files {
			files = append(files, file)
		}
		sort.Slice(files, func(i, j int) bool { return files[i].Path < files[j].Path })
		record := struct {
			FixtureID        string                  `json:"fixture_id"`
			Revision         int                     `json:"revision"`
			ManifestSHA256   string                  `json:"manifest_sha256"`
			Test             string                  `json:"test"`
			AssertionsPassed bool                    `json:"assertions_passed"`
			Files            []fixtureartifacts.File `json:"admitted_files"`
		}{ID, Revision, c.ManifestSHA256, t.Name(), !t.Failed(), files}
		raw, err := json.MarshalIndent(record, "", "  ")
		if err == nil {
			err = os.WriteFile(filepath.Join(workspace, "comparison.json"), append(raw, '\n'), 0600)
		}
		if err != nil {
			t.Error(err)
		}
	})
	t.Logf("IR fixture comparison: %s", filepath.Join(workspace, "comparison.json"))
	return c
}
