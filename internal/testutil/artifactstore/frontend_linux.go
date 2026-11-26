//go:build linux

package artifactstore

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"slices"
	"strings"
)

// readinessOwner owns the frontend installation receipt and the readiness
// probes: its identity binds the pins, package manifests, workspace
// configuration, lockfile and installed lock metadata.
const readinessOwner = "tools/harness/readiness/installed-readiness.mjs"

// pnpmInstallFlags never depend on how a command reports, so the output mode
// cannot change an installation.
var pnpmInstallFlags = []string{"--frozen-lockfile", "--reporter=append-only", "--loglevel=warn"}

// pnpmSettings are the workspace settings an installation requires.
var pnpmSettings = [][2]string{{"store-dir", ".pnpm-store"}, {"confirmModulesPurge", "false"}}

// FrontendInstall installs the workspace's frontend dependencies with a
// frozen lockfile. Its receipt is the frontend installation receipt, whose
// owner decides whether it is current; the installation writes it last.
type FrontendInstall struct {
	// Root is the repository root; Receipt the installation receipt.
	Root, Receipt string
	// Node and PNPM are the executables; NodeRuntime their runtime directory.
	Node, PNPM, NodeRuntime string
	Env                     []string
}

// ID names the producer.
func (f FrontendInstall) ID() string { return "frontend-install" }

func (f FrontendInstall) command(name string, args ...string) *exec.Cmd {
	command := exec.Command(name, args...)
	command.Dir, command.Env = f.Root, nodeEnvironment(slices.Clone(f.Env), f.NodeRuntime)
	return command
}

// Current asks the receipt's owner whether the receipt matches the
// installation's inputs and the installed lock.
func (f FrontendInstall) Current() (bool, error) {
	return probe(f.command(f.Node, readinessOwner, "frontend-installation"))
}

// Install checks the workspace's pnpm settings, removes the receipt, installs
// with a frozen lockfile and has the receipt's owner publish it.
func (f FrontendInstall) Install() error {
	for _, setting := range pnpmSettings {
		output, err := f.command(f.PNPM, "config", "get", setting[0]).Output()
		if got := strings.TrimSpace(string(output)); err != nil || got != setting[1] {
			return fmt.Errorf("pnpm %s must be %s; got %q: %v", setting[0], setting[1], got, err)
		}
	}
	if err := os.Remove(f.Receipt); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	var output bytes.Buffer
	install := f.command(f.PNPM, append([]string{"install"}, pnpmInstallFlags...)...)
	install.Env = append(install.Env, "CI=true")
	install.Stdout, install.Stderr = io.MultiWriter(os.Stderr, &output), io.MultiWriter(os.Stderr, &output)
	if err := install.Run(); err != nil {
		if strings.Contains(output.String(), "ERR_PNPM_OUTDATED_LOCKFILE") {
			return errors.New("pnpm-lock.yaml does not match the package manifests; run make frontend-lockfile-update")
		}
		return fmt.Errorf("pnpm install: %w", err)
	}
	publish := f.command(f.Node, readinessOwner, "publish")
	publish.Stdout, publish.Stderr = os.Stderr, os.Stderr
	if err := publish.Run(); err != nil {
		return fmt.Errorf("publish the frontend installation receipt: %w", err)
	}
	return nil
}
