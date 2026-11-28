//go:build linux

package artifactstore

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"slices"
	"strings"
)

const goToolKeySchemaID = "cartulary.go_tool_key.v1"

// goToolImplementation is this producer's own source.
const goToolImplementation = "internal/testutil/artifactstore/gotool_linux.go"

// GoTool installs one pinned Go tool. Its key covers the pinned module and
// version, the module files of the directory it builds from, the effective Go
// settings and this producer; nothing a launcher or location contributes. A
// tool with its own command builds that command from its module directory
// against the pinned module's library, and its key also covers the command's
// source closure.
type GoTool struct {
	// Root is the repository root; ModuleDir, when set, is a module directory
	// under it whose go.mod and go.sum select the tool's dependencies.
	Root, ModuleDir string
	// Module is the pinned module path and version, path@version.
	Module string
	// Package, when set, is the tool's own command, a ./package in ModuleDir
	// named as the pinned tool is.
	Package string
	// Go is the go executable; Env is the environment it runs with.
	Go  string
	Env []string
	// Wrap, when set, wraps the install command, such as in a run step.
	Wrap func(command []string) []string
}

// ID names the producer in key records.
func (t GoTool) ID() string { return "go-tool" }

// Name is the installed binary's file name: the package path's last element.
func (t GoTool) Name() string {
	modulePath, _, _ := strings.Cut(t.Module, "@")
	return path.Base(modulePath)
}

// Version is the pinned version.
func (t GoTool) Version() string {
	_, version, _ := strings.Cut(t.Module, "@")
	return version
}

func (t GoTool) installArgs() []string {
	if t.ModuleDir == "" {
		return []string{"install", t.Module}
	}
	if t.Package != "" {
		return []string{"-C", t.ModuleDir, "install", "-mod=readonly", t.Package}
	}
	modulePath, _, _ := strings.Cut(t.Module, "@")
	return []string{"-C", t.ModuleDir, "install", "-mod=readonly", modulePath}
}

// Key digests the pinned module, its module files, the Go settings and this
// producer.
func (t GoTool) Key() (string, error) {
	if t.Name() == "." || t.Version() == "" {
		return "", fmt.Errorf("go tool %q is not a pinned path@version", t.Module)
	}
	names := []string{goToolImplementation}
	if t.ModuleDir != "" {
		names = append(names, t.ModuleDir+"/go.mod", t.ModuleDir+"/go.sum")
	}
	if t.Package != "" {
		if path.Base(t.Package) != t.Name() {
			return "", fmt.Errorf("go tool command %s is not named %s", t.Package, t.Name())
		}
		closure, err := goClosure(t.Go, filepath.Join(t.Root, filepath.FromSlash(t.ModuleDir)), t.Package, "", t.Env)
		if err != nil {
			return "", err
		}
		for _, file := range closure {
			names = append(names, t.ModuleDir+"/"+file)
		}
	}
	inputs := map[string]string{}
	for _, name := range names {
		digest, err := fileDigest(filepath.Join(t.Root, filepath.FromSlash(name)))
		if err != nil {
			return "", err
		}
		inputs[name] = digest
	}
	settings, err := effectiveGoSettings(t.Go, t.Root, t.Env)
	if err != nil {
		return "", err
	}
	record := map[string]any{
		"schema_id": goToolKeySchemaID, "producer": t.ID(), "module": t.Module, "module_dir": t.ModuleDir,
		"install_arguments": t.installArgs(), "outputs": []string{t.Name()}, "settings": settings, "inputs": inputs,
	}
	if t.Package != "" {
		record["package"] = t.Package
	}
	data, err := json.Marshal(record)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:]), nil
}

// Build installs the tool into staging.
func (t GoTool) Build(staging string) error {
	install := append([]string{t.Go}, t.installArgs()...)
	if t.Wrap != nil {
		install = t.Wrap(install)
	}
	command := exec.Command(install[0], install[1:]...)
	command.Dir = t.Root
	command.Env = append(slices.Clone(t.Env), "GOBIN="+staging, "GOTELEMETRY=off")
	command.Stdout, command.Stderr = os.Stderr, os.Stderr
	if err := command.Run(); err != nil {
		return fmt.Errorf("install %s: %w", t.Module, err)
	}
	info, err := os.Stat(filepath.Join(staging, t.Name()))
	if err != nil || !info.Mode().IsRegular() || info.Mode().Perm()&0o100 == 0 {
		return fmt.Errorf("installing %s produced no executable %s", t.Module, t.Name())
	}
	return nil
}

// effectiveGoSettings reports the Go settings a build reads, after every
// configuration source has applied.
func effectiveGoSettings(goExecutable, dir string, env []string) (map[string]string, error) {
	command := exec.Command(goExecutable, append([]string{"env", "-json"}, goSettings...)...)
	command.Dir, command.Env = dir, env
	output, err := command.Output()
	if err != nil {
		return nil, fmt.Errorf("go env: %w", err)
	}
	var settings map[string]string
	if err := json.Unmarshal(output, &settings); err != nil {
		return nil, err
	}
	return settings, nil
}
