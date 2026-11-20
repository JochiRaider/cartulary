package fixtureartifacts

import (
	"os"

	"github.com/JochiRaider/cartulary/internal/platform/harnessruntime/runevidence"
	"github.com/JochiRaider/cartulary/internal/testutil/suiteservices"
)

// SelectedRunRoot resolves the public harness environment without inspecting
// documentation or starting services. An explicit run identity is mandatory.
// The run-root rule is shared with the contract evidence writer.
func SelectedRunRoot() (string, error) {
	if _, err := suiteservices.SelectedResultsRoot(); err != nil {
		return "", err
	}
	return runevidence.RunRoot(os.LookupEnv)
}
