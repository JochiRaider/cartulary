// Package goenv is the one Go environment every harness build and test runs
// under (TH-HARNESS, P5-3d): the settings tools/toolchain_pins.json declares
// under go_environment, the pinned toolchain and the pinned cryptographic
// module. Builds and tests share it, so a build started from inside a test
// keys like one started from a shell, and shipped binaries are built the way
// they are tested: linux/amd64.v1, cgo disabled, no environment file or
// workspace, no flags or experiments.
package goenv

import (
	"bytes"
	"encoding/json"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"slices"
)

// PinsPath is the toolchain pins, relative to the repository root.
const PinsPath = "tools/toolchain_pins.json"

// Settings are the names go_environment declares, each exactly once.
var Settings = []string{"CGO_ENABLED", "GO111MODULE", "GOAMD64", "GOARCH", "GOENV", "GOEXPERIMENT", "GOFLAGS", "GOOS", "GOWORK"}

// Parse returns the pinned Go environment from the toolchain pins: the
// declared settings, GOTOOLCHAIN and GOFIPS140.
func Parse(pins []byte) (map[string]string, error) {
	var pinned struct {
		GoToolchain         string            `json:"go_toolchain"`
		GoEnvironment       map[string]string `json:"go_environment"`
		CryptographicModule struct {
			Selector string `json:"selector"`
		} `json:"cryptographic_module"`
	}
	if err := json.NewDecoder(bytes.NewReader(pins)).Decode(&pinned); err != nil {
		return nil, fmt.Errorf("%s: %w", PinsPath, err)
	}
	if names := slices.Sorted(maps.Keys(pinned.GoEnvironment)); !slices.Equal(names, Settings) {
		return nil, fmt.Errorf("%s: go_environment declares %v, want exactly %v", PinsPath, names, Settings)
	}
	if pinned.GoToolchain == "" || pinned.CryptographicModule.Selector == "" {
		return nil, fmt.Errorf("%s: the Go environment needs go_toolchain and cryptographic_module.selector", PinsPath)
	}
	environment := maps.Clone(pinned.GoEnvironment)
	environment["GOTOOLCHAIN"] = pinned.GoToolchain
	environment["GOFIPS140"] = pinned.CryptographicModule.Selector
	return environment, nil
}

// Read returns the pinned Go environment of the checkout at root.
func Read(root string) (map[string]string, error) {
	pins, err := os.ReadFile(filepath.Join(root, PinsPath))
	if err != nil {
		return nil, err
	}
	return Parse(pins)
}

// Race is environment for a race-detector build, which needs cgo.
func Race(environment map[string]string) map[string]string {
	race := maps.Clone(environment)
	race["CGO_ENABLED"] = "1"
	return race
}

// Entries returns environment as NAME=value entries in name order, to append
// to a process environment so that they take precedence.
func Entries(environment map[string]string) []string {
	entries := make([]string, 0, len(environment))
	for _, name := range slices.Sorted(maps.Keys(environment)) {
		entries = append(entries, name+"="+environment[name])
	}
	return entries
}
