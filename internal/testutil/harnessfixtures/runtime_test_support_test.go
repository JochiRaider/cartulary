//go:build linux

package harnessfixtures

import (
	"testing"

	"github.com/JochiRaider/cartulary/internal/testutil/harnesslocal"
	"github.com/JochiRaider/cartulary/internal/testutil/harnesssession"
)

func runtimeFixture(t *testing.T) (*harnesssession.Runtime, string, string) {
	t.Helper()
	root, run, scratch := t.TempDir(), t.TempDir(), t.TempDir()
	runtime, err := harnesssession.NewRuntime(root, run, "fixture", scratch)
	if err != nil {
		t.Fatal(err)
	}
	if harnesslocal.CheckDirectory(runtime.Directory, runtime.Directory.Name()) != nil {
		t.Fatal("runtime descriptor retained staging pathname")
	}
	t.Cleanup(func() {
		if err := runtime.Close(); err != nil {
			t.Error(err)
		}
	})
	return runtime, root, run
}
