//go:build linux

package artifactstore

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
)

// ReportRenderer builds the isolated reporting runtime installation, whose
// layout the Reporting NLSpec owns: content-addressed roots and their
// manifests beside a receipt, in its directory. The store builds its
// confinement launcher; the packaging script keys, reuses, builds, seals and
// collects installations under the renderer's lock.
type ReportRenderer struct {
	// Root is the repository root; Directory the installation directory.
	Root, Directory string
	Store           Store
	// Sandbox is the confinement launcher.
	Sandbox                 GoBinary
	Node, PNPM, NodeRuntime string
	Env                     []string
	Wrap                    func([]string) []string
	Controls                Controls
}

// ID names the producer.
func (r ReportRenderer) ID() string { return "report-renderer" }

// Produce builds the launcher through the store, then runs the packaging
// script with the launcher's root while holding the renderer's lock.
func (r ReportRenderer) Produce() error {
	if err := os.MkdirAll(r.Directory, 0o755); err != nil {
		return err
	}
	unlock, err := lockFile(r.Directory + ".lock")
	if err != nil {
		return err
	}
	defer unlock()
	launcher, err := r.Store.Produce(r.Sandbox, map[string]string{}, r.Controls)
	if err != nil {
		return fmt.Errorf("the confinement launcher: %w", err)
	}
	argv := []string{r.Node, "tools/report-renderer/build-runtime.mjs"}
	if r.Wrap != nil {
		argv = r.Wrap(argv)
	}
	command := exec.Command(argv[0], argv[1:]...)
	command.Dir = r.Root
	command.Env = append(nodeEnvironment(slices.Clone(r.Env), r.NodeRuntime), "NODE_BIN="+r.Node, "PNPM="+r.PNPM,
		"RENDER_SANDBOX_BIN="+filepath.Join(r.Store.RootPath(launcher.Root), r.Sandbox.Name), "RENDER_SANDBOX_ROOT="+launcher.Root)
	command.Stdout, command.Stderr = os.Stderr, os.Stderr
	if err := command.Run(); err != nil {
		return fmt.Errorf("build the reporting runtime: %w", err)
	}
	return nil
}
