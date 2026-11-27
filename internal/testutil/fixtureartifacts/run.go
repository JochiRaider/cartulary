package fixtureartifacts

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/JochiRaider/cartulary/internal/platform/harnessruntime/runevidence"
	"github.com/JochiRaider/cartulary/internal/testutil/suiteservices"
)

// unitArtifactDirEnv names the executing unit's own directory, which the
// graph runtime gives every unit it launches.
const unitArtifactDirEnv = "CARTULARY_STEP_ARTIFACT_DIR"

// SelectedUnitRoot is the executing unit's evidence directory,
// unit-artifacts/<unit> under the selected run (TH-HARNESS-REQ-874). Fixture
// copies and test evidence a unit writes belong there, never at the run
// root's top level. It resolves the public harness environment without
// inspecting documentation or starting services; an explicit run identity and
// the unit's directory are mandatory.
func SelectedUnitRoot() (string, error) {
	if _, err := suiteservices.SelectedResultsRoot(); err != nil {
		return "", err
	}
	run, err := runevidence.RunRoot(os.LookupEnv)
	if err != nil {
		return "", err
	}
	run, err = filepath.Abs(run)
	if err != nil {
		return "", err
	}
	unit := os.Getenv(unitArtifactDirEnv)
	if !filepath.IsAbs(unit) || filepath.Clean(unit) != unit || filepath.Dir(unit) != filepath.Join(run, "unit-artifacts") {
		return "", errors.New("test evidence requires the executing unit's artifact directory")
	}
	return unit, nil
}
