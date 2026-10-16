//go:build linux

package harnessfixtures

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/JochiRaider/cartulary/internal/testutil/harnesssession"
)

func testRecoveryCommand(t *testing.T) {
	t.Run("exact_recovery_command", func(t *testing.T) {
		root, run, scratch := t.TempDir(), filepath.Join(t.TempDir(), "fixture"), t.TempDir()
		if err := os.Mkdir(run, 0700); err != nil {
			t.Fatal(err)
		}
		r, err := harnesssession.NewRuntime(root, run, "fixture", scratch)
		if err != nil {
			t.Fatal(err)
		}
		defer r.ReleaseHandles()
		original := filepath.Join(run, "original-result.json")
		if err := os.WriteFile(original, []byte("failed\n"), 0600); err != nil {
			t.Fatal(err)
		}
		input := map[string]string{"RESULTS_DIR": run, "RUNTIME_ROOT": r.Root, "RUNTIME_LEASE_ID": r.LeaseID}
		var output bytes.Buffer
		for _, key := range []string{"RESULTS_DIR", "RUNTIME_ROOT", "RUNTIME_LEASE_ID"} {
			valid := input[key]
			input[key] = ""
			if code := RecoverRun(context.Background(), root, input, &output); code != 2 {
				t.Fatal(key, code)
			}
			input[key] = valid
		}
		input["RESULTS_DIR"] = filepath.Join(filepath.Dir(run), "foreign")
		if code := RecoverRun(context.Background(), root, input, &output); code != 2 {
			t.Fatal("foreign run admitted", code)
		}
		input["RESULTS_DIR"] = run
		input["RUNTIME_LEASE_ID"] = "00000000-0000-4000-8000-000000000000"
		if code := RecoverRun(context.Background(), root, input, &output); code != 2 {
			t.Fatal("foreign lease admitted", code)
		}
		input["RUNTIME_LEASE_ID"] = r.LeaseID
		link := filepath.Join(scratch, "link")
		if err := os.Symlink(r.Root, link); err != nil {
			t.Fatal(err)
		}
		input["RUNTIME_ROOT"] = link
		if code := RecoverRun(context.Background(), root, input, &output); code != 2 {
			t.Fatal("symlink admitted", code)
		}
		input["RUNTIME_ROOT"] = r.Root
		if code := RecoverRun(context.Background(), root, input, &output); code != 0 {
			t.Fatal("empty exact recovery", code, output.String())
		}
		if _, err := os.Stat(r.Root); !os.IsNotExist(err) {
			t.Fatal("runtime retained", err)
		}
		data, err := os.ReadFile(original)
		if err != nil || string(data) != "failed\n" {
			t.Fatal("original evidence mutated", err)
		}
	})
}
