//go:build linux

package artifactstore

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
)

const webBundleKeySchemaID = "cartulary.web_bundle_key.v1"

// webBundleImplementation is this producer's own source.
const webBundleImplementation = "internal/testutil/artifactstore/web_linux.go"

// documentationFile is human documentation, which no build reads.
var documentationFile = regexp.MustCompile(`(?i)(\.(md|markdown)$|(^|/)readme(\.[^/]*)?$)`)

// WebBundle builds one Vite bundle of apps/web. Its key covers every
// repository file under its input roots, its fixed build arguments, the Node
// and pnpm it runs with and the VITE_ settings and NODE_ENV it sees. The
// arguments never depend on how verbosely a command reports: a bundle always
// carries source maps.
type WebBundle struct {
	// Root is the repository root; Inputs are the repository paths whose
	// files the build can read.
	Root   string
	Inputs []string
	// Mode is the Vite mode: production or measurement.
	Mode string
	// Node is the Node executable; PNPM the pnpm launcher; NodeRuntime the
	// runtime directory both come from; PNPMVersion the pinned version.
	Node, PNPM, NodeRuntime, PNPMVersion string
	Env                                  []string
}

// ID names the producer in key records.
func (w WebBundle) ID() string { return "web-bundle" }

func (w WebBundle) buildArgs() []string {
	args := []string{"--dir", "apps/web", "exec", "vite", "build", "--sourcemap", "--logLevel", "warn"}
	if w.Mode != "production" {
		args = append(args, "--mode", w.Mode)
	}
	return args
}

// repositoryFiles lists the tracked and untracked, non-ignored files under
// paths, without human documentation.
func repositoryFiles(root string, paths []string) ([]string, error) {
	command := exec.Command("git", append([]string{"ls-files", "--cached", "--others", "--exclude-standard", "-z", "--"}, paths...)...)
	command.Dir = root
	var stderr bytes.Buffer
	command.Stderr = &stderr
	output, err := command.Output()
	if err != nil {
		return nil, fmt.Errorf("git ls-files: %w: %s", err, strings.TrimSpace(stderr.String()))
	}
	files := []string{}
	for _, name := range strings.Split(string(output), "\x00") {
		if name != "" && !documentationFile.MatchString(name) {
			files = append(files, name)
		}
	}
	slices.Sort(files)
	return slices.Compact(files), nil
}

// Key digests the input files, build arguments, toolchain and settings.
func (w WebBundle) Key() (string, error) {
	files, err := repositoryFiles(w.Root, w.Inputs)
	if err != nil {
		return "", err
	}
	if len(files) == 0 {
		return "", fmt.Errorf("web bundle inputs %v hold no files", w.Inputs)
	}
	inputs := map[string]string{}
	for _, name := range append(files, webBundleImplementation) {
		path := filepath.Join(w.Root, filepath.FromSlash(name))
		info, err := os.Lstat(path)
		if os.IsNotExist(err) {
			// A deleted tracked file is absent from the build too.
			continue
		}
		if err != nil {
			return "", err
		}
		if !info.Mode().IsRegular() {
			inputs[name] = "non-regular:" + info.Mode().Type().String()
			continue
		}
		if inputs[name], err = fileDigest(path); err != nil {
			return "", err
		}
	}
	version, err := exec.Command(w.Node, "--version").Output()
	if err != nil {
		return "", fmt.Errorf("node --version: %w", err)
	}
	settings := map[string]string{}
	for _, entry := range w.Env {
		name, value, _ := strings.Cut(entry, "=")
		if strings.HasPrefix(name, "VITE_") || name == "NODE_ENV" {
			settings[name] = value
		}
	}
	data, err := json.Marshal(map[string]any{
		"schema_id": webBundleKeySchemaID, "producer": w.ID(), "mode": w.Mode, "build_arguments": w.buildArgs(),
		"node_version": strings.TrimSpace(string(version)), "pnpm_version": w.PNPMVersion, "settings": settings, "inputs": inputs,
	})
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:]), nil
}

// Build renders the bundle into staging.
func (w WebBundle) Build(staging string) error {
	command := exec.Command(w.PNPM, append(w.buildArgs(), "--outDir", staging, "--emptyOutDir")...)
	command.Dir = w.Root
	command.Env = append(slices.Clone(w.Env),
		"PATH="+filepath.Join(w.NodeRuntime, "bin")+string(os.PathListSeparator)+environmentValue(w.Env, "PATH"),
		"COREPACK_HOME="+filepath.Join(w.NodeRuntime, "corepack"))
	command.Stdout, command.Stderr = os.Stderr, os.Stderr
	if err := command.Run(); err != nil {
		return fmt.Errorf("vite build (%s): %w", w.Mode, err)
	}
	if _, err := os.Stat(filepath.Join(staging, "index.html")); err != nil {
		return fmt.Errorf("vite build (%s) produced no index.html", w.Mode)
	}
	return nil
}

// environmentValue returns env's last value for name.
func environmentValue(env []string, name string) string {
	for i := len(env) - 1; i >= 0; i-- {
		if value, ok := strings.CutPrefix(env[i], name+"="); ok {
			return value
		}
	}
	return ""
}
