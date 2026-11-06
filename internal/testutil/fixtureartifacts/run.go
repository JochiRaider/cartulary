package fixtureartifacts

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/JochiRaider/cartulary/internal/testutil/suiteservices"
)

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
	if !SafePath(id) || strings.Contains(id, "/") {
		return "", errors.New("invalid run ID")
	}
	return filepath.Join(root, id), nil
}
