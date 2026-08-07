package appsupport

import (
	"testing"

	"github.com/JochiRaider/cartulary/internal/app/configassembly"
	"github.com/JochiRaider/cartulary/internal/platform/config"
)

// LoadDeploymentConfiguration selects an explicit deployment artifact for a
// test and returns its admitted defensive application projection.
func LoadDeploymentConfiguration(t testing.TB, path string) configassembly.Deployment {
	t.Helper()
	loaded, err := configassembly.Load(config.LoadOptions{Path: path})
	if err != nil {
		t.Fatalf("load deployment configuration %s: %v", path, err)
	}
	return loaded.Deployment()
}
