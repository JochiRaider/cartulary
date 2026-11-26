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
	"runtime"
	"strings"
)

const nodeRuntimeKeySchemaID = "cartulary.node_runtime_key.v1"

// nodeRuntimeImplementation is this producer's own source.
const nodeRuntimeImplementation = "internal/testutil/artifactstore/node_linux.go"

// nodeReleases is where Node publishes its release archives.
const nodeReleases = "https://nodejs.org/dist"

// NodeRuntimeReceipt is the receipt's name inside the runtime directory.
const NodeRuntimeReceipt = "installation.json"

// NodeRuntime installs the pinned Node release, from its archive with the
// committed digest, with the pinned pnpm prepared through Corepack. Its key
// is the archive's name and digest, the pnpm version and this producer.
type NodeRuntime struct {
	// Root is the repository root, where this producer's source is keyed;
	// Directory is the runtime directory, replaced as a whole.
	Root, Directory string
	// Version and PNPMVersion are the pinned versions; Archive is the release
	// archive's file name and SHA256 its committed digest.
	Version, PNPMVersion, Archive, SHA256 string
	// Archives keeps downloaded archives; Releases is the download base,
	// empty for Node's own.
	Archives, Releases string
	Env                []string
}

// NodeArchive names the release archive of version for this platform.
func NodeArchive(version string) (string, error) {
	platform := map[string]string{"amd64": "x64", "arm64": "arm64"}[runtime.GOARCH]
	if runtime.GOOS != "linux" || platform == "" {
		return "", fmt.Errorf("no Node release for %s/%s", runtime.GOOS, runtime.GOARCH)
	}
	return fmt.Sprintf("node-v%s-linux-%s.tar.xz", version, platform), nil
}

// ID names the producer in receipts.
func (n NodeRuntime) ID() string { return "node-runtime" }

func (n NodeRuntime) key() (string, error) {
	implementation, err := fileDigest(filepath.Join(n.Root, nodeRuntimeImplementation))
	if err != nil {
		return "", err
	}
	data, err := json.Marshal(map[string]any{
		"schema_id": nodeRuntimeKeySchemaID, "producer": n.ID(), "version": n.Version, "pnpm_version": n.PNPMVersion,
		"archive": n.Archive, "archive_sha256": n.SHA256, "inputs": map[string]string{nodeRuntimeImplementation: implementation},
	})
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:]), nil
}

// Current reports whether the runtime's receipt records the current key and
// its node and pnpm executables are present.
func (n NodeRuntime) Current() (bool, error) {
	key, err := n.key()
	if err != nil {
		return false, err
	}
	if !receiptMatches(filepath.Join(n.Directory, NodeRuntimeReceipt), n.ID(), key) {
		return false, nil
	}
	for _, name := range []string{"node", "pnpm"} {
		if info, err := os.Stat(filepath.Join(n.Directory, "bin", name)); err != nil || info.Mode().Perm()&0o111 == 0 {
			return false, nil
		}
	}
	return true, nil
}

// Install extracts the verified archive beside the runtime directory,
// prepares pnpm, checks both versions and writes the receipt, then swaps the
// new runtime into place by rename.
func (n NodeRuntime) Install() error {
	key, err := n.key()
	if err != nil {
		return err
	}
	releases := n.Releases
	if releases == "" {
		releases = nodeReleases
	}
	archive, err := verifiedArchive(n.Archives, n.Archive, releases+"/v"+n.Version+"/"+n.Archive, n.SHA256)
	if err != nil {
		return err
	}
	parent, base := filepath.Dir(n.Directory), filepath.Base(n.Directory)
	if err := os.MkdirAll(parent, 0o755); err != nil {
		return err
	}
	staging, err := os.MkdirTemp(parent, "."+base+"-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(staging)
	env := nodeEnvironment(n.Env, staging)
	corepack := filepath.Join(staging, "bin", "corepack")
	for _, command := range [][]string{
		{"tar", "-xJf", archive, "-C", staging, "--strip-components=1"},
		{corepack, "enable", "--install-directory", filepath.Join(staging, "bin"), "pnpm"},
		{corepack, "prepare", "pnpm@" + n.PNPMVersion, "--activate"},
	} {
		cmd := exec.Command(command[0], command[1:]...)
		cmd.Env, cmd.Stdout, cmd.Stderr = env, os.Stderr, os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("%s: %w", strings.Join(command, " "), err)
		}
	}
	for name, want := range map[string]string{"node": "v" + n.Version, "pnpm": n.PNPMVersion} {
		executable := filepath.Join(staging, "bin", name)
		cmd := exec.Command(executable, "--version")
		cmd.Env, cmd.Stderr = env, os.Stderr
		output, err := cmd.Output()
		if got := strings.TrimSpace(string(output)); err != nil || got != want {
			return fmt.Errorf("the installed %s reports %q, not %s: %v", name, got, want, err)
		}
	}
	if err := writeReceipt(filepath.Join(staging, NodeRuntimeReceipt), n.ID(), key); err != nil {
		return err
	}
	return swapDirectory(staging, n.Directory)
}

// swapDirectory replaces target with the directory staging by rename. A
// previous target is renamed aside first and removed after the swap, so a
// reader sees either the old tree or the new one.
func swapDirectory(staging, target string) error {
	old := ""
	if _, err := os.Lstat(target); err == nil {
		old = filepath.Join(filepath.Dir(target), "."+filepath.Base(target)+"-old-"+token())
		if err := os.Rename(target, old); err != nil {
			return err
		}
	} else if !os.IsNotExist(err) {
		return err
	}
	if err := os.Rename(staging, target); err != nil {
		if old != "" {
			_ = os.Rename(old, target)
		}
		return err
	}
	if old != "" {
		return os.RemoveAll(old)
	}
	return nil
}
