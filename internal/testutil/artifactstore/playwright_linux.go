//go:build linux

package artifactstore

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
)

const playwrightKeySchemaID = "cartulary.playwright_key.v1"

// playwrightImplementation is this producer's own source.
const playwrightImplementation = "internal/testutil/artifactstore/playwright_linux.go"

// playwrightInstaller installs Chromium with its native dependencies, pulls
// the pinned visual renderer image and prepares browser trust.
const playwrightInstaller = "tools/harness/browser/playwright-install.sh"

// playwrightInputs are the files an installation is made from: the installed
// Playwright engine and its browser revisions, the visual renderer profile,
// the browser trust preparation and the installer.
var playwrightInputs = []string{
	"node_modules/playwright-core/package.json", "node_modules/playwright-core/browsers.json",
	"tools/frontend_visual_renderer_profile.json", "tools/harness/browser/visual-renderer-profile.mjs",
	"tools/harness/browser/browser-trust.mjs", playwrightInstaller, playwrightImplementation,
}

// Playwright installs the browsers browser rows use. Its receipt records the
// key; reuse also needs the browser readiness proof, because Chromium and the
// renderer image live outside the receipt's directory, in Playwright's cache
// and the Docker store. A failed proof installs again, which pulls a missing
// or replaced image.
type Playwright struct {
	// Root is the repository root; Receipt the installation receipt.
	Root, Receipt string
	// Node and PNPM are the executables; NodeRuntime their runtime directory.
	Node, PNPM, NodeRuntime string
	Env                     []string
	Wrap                    func([]string) []string
}

// ID names the producer in receipts.
func (p Playwright) ID() string { return "playwright" }

func (p Playwright) key() (string, error) {
	inputs := map[string]string{}
	for _, name := range playwrightInputs {
		digest, err := fileDigest(filepath.Join(p.Root, filepath.FromSlash(name)))
		if err != nil {
			return "", err
		}
		inputs[name] = digest
	}
	data, err := json.Marshal(map[string]any{"schema_id": playwrightKeySchemaID, "producer": p.ID(), "inputs": inputs})
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:]), nil
}

func (p Playwright) environment() []string {
	return append(nodeEnvironment(slices.Clone(p.Env), p.NodeRuntime), "NODE_BIN="+p.Node, "PNPM="+p.PNPM)
}

// Current reports whether the receipt records the current key and the
// browser readiness proof passes.
func (p Playwright) Current() (bool, error) {
	key, err := p.key()
	if err != nil || !receiptMatches(p.Receipt, p.ID(), key) {
		return false, err
	}
	command := exec.Command(p.Node, readinessOwner, "browser")
	command.Dir, command.Env = p.Root, p.environment()
	return probe(command)
}

// Install runs the installer and writes the receipt.
func (p Playwright) Install() error {
	key, err := p.key()
	if err != nil {
		return err
	}
	if err := os.Remove(p.Receipt); err != nil && !os.IsNotExist(err) {
		return err
	}
	argv := []string{"bash", playwrightInstaller}
	if p.Wrap != nil {
		argv = p.Wrap(argv)
	}
	command := exec.Command(argv[0], argv[1:]...)
	command.Dir, command.Env = p.Root, p.environment()
	command.Stdout, command.Stderr = os.Stderr, os.Stderr
	if err := command.Run(); err != nil {
		return fmt.Errorf("install Playwright browsers: %w", err)
	}
	return writeReceipt(p.Receipt, p.ID(), key)
}
