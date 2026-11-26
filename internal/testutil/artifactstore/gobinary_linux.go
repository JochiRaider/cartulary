//go:build linux

package artifactstore

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
)

const goBinaryKeySchemaID = "cartulary.go_binary_key.v1"

// producerImplementation is this producer's own source: how it builds is part
// of what it builds.
const producerImplementation = "internal/testutil/artifactstore/gobinary_linux.go"

// receiptWriter is the code and contract that turn a built binary into its
// cryptographic receipt; the receipt's bytes depend on them.
var receiptWriter = []string{
	"tools/harness/readiness/cryptographic-build-cli.mjs",
	"tools/harness/readiness/cryptographic-module.mjs",
	"tools/schemas/cartulary.cryptographic_build_receipt.v1.schema.json",
}

// goSettings are the effective Go settings a build reads, as `go env` reports
// them after every configuration source has applied.
var goSettings = []string{"CGO_ENABLED", "GOAMD64", "GOARCH", "GOEXPERIMENT", "GOFIPS140", "GOFLAGS", "GOOS", "GOTOOLCHAIN", "GOVERSION"}

// GoBinary produces one Go binary and its cryptographic receipt. Its key
// covers what the build reads and nothing a launcher contributes: this
// declaration, every repository file of the package's import closure, the
// module files, the effective Go settings, the pinned cryptographic module
// and the receipt writer.
type GoBinary struct {
	// Module is the module root the package belongs to.
	Module string
	// Package is the package path to build, relative to Module.
	Package string
	// Tags are the build tags, comma-separated.
	Tags string
	// Name is the binary's file name in its root; the receipt is Name.crypto.json.
	Name string
	// Go and Node are the executables; Env is the environment both run with.
	Go, Node string
	Env      []string
	// Wrap, when set, wraps the build command, such as in a run step.
	Wrap func(command []string) []string
}

// ID names the producer in key records.
func (b GoBinary) ID() string { return "go-binary" }

// buildArgs never stamp VCS state and trim source paths (D27), so a binary's
// bytes do not depend on the checkout it was built in.
func (b GoBinary) buildArgs() []string {
	args := []string{"build", "-buildvcs=false", "-trimpath"}
	if b.Tags != "" {
		args = append(args, "-tags", b.Tags)
	}
	return args
}

func (b GoBinary) run(name string, args []string, env []string) ([]byte, error) {
	command := exec.Command(name, args...)
	command.Dir = b.Module
	command.Env = env
	var stderr bytes.Buffer
	command.Stderr = &stderr
	output, err := command.Output()
	if err != nil {
		return nil, fmt.Errorf("%s %s: %w: %s", filepath.Base(name), strings.Join(args, " "), err, strings.TrimSpace(stderr.String()))
	}
	return output, nil
}

// closure lists the repository files of the package's import closure that a
// build reads.
func (b GoBinary) closure() ([]string, error) {
	return goClosure(b.Go, b.Module, b.Package, b.Tags, b.Env)
}

// goClosure lists the repository files of a package's import closure that a
// build reads: every non-test source of each kind Go compiles and every
// embedded file, under the build's own caches and with modules read-only.
func goClosure(goExecutable, module, pkg, tags string, env []string) ([]string, error) {
	args := []string{"list", "-deps", "-json"}
	if tags != "" {
		args = append(args, "-tags", tags)
	}
	command := exec.Command(goExecutable, append(args, "--", pkg)...)
	command.Dir, command.Env = module, append(slices.Clone(env), "GOFLAGS=-mod=readonly")
	var stderr bytes.Buffer
	command.Stderr = &stderr
	output, err := command.Output()
	if err != nil {
		return nil, fmt.Errorf("go list %s: %w: %s", pkg, err, strings.TrimSpace(stderr.String()))
	}
	module, err = filepath.Abs(module)
	if err != nil {
		return nil, err
	}
	files := []string{}
	decoder := json.NewDecoder(bytes.NewReader(output))
	for {
		var p struct {
			Dir                                                 string
			Standard                                            bool
			GoFiles, CgoFiles, CFiles, CXXFiles, MFiles, HFiles []string
			FFiles, SFiles, SwigFiles, SwigCXXFiles, SysoFiles  []string
			EmbedFiles                                          []string
		}
		if err := decoder.Decode(&p); errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			return nil, err
		}
		if p.Standard || !strings.HasPrefix(p.Dir+string(filepath.Separator), module+string(filepath.Separator)) {
			continue
		}
		for _, group := range [][]string{p.GoFiles, p.CgoFiles, p.CFiles, p.CXXFiles, p.MFiles, p.HFiles, p.FFiles, p.SFiles, p.SwigFiles, p.SwigCXXFiles, p.SysoFiles, p.EmbedFiles} {
			for _, name := range group {
				relative, err := filepath.Rel(module, filepath.Join(p.Dir, name))
				if err != nil {
					return nil, err
				}
				files = append(files, filepath.ToSlash(relative))
			}
		}
	}
	slices.Sort(files)
	return slices.Compact(files), nil
}

// Key digests everything the binary and its receipt depend on.
func (b GoBinary) Key() (string, error) {
	closure, err := b.closure()
	if err != nil {
		return "", err
	}
	if len(closure) == 0 {
		return "", fmt.Errorf("%s has no repository files", b.Package)
	}
	inputs := map[string]string{}
	for _, name := range slices.Concat([]string{"go.mod", "go.sum", producerImplementation}, receiptWriter, closure) {
		digest, err := fileDigest(filepath.Join(b.Module, filepath.FromSlash(name)))
		if err != nil {
			return "", err
		}
		inputs[name] = digest
	}
	effective, err := effectiveGoSettings(b.Go, b.Module, b.Env)
	if err != nil {
		return "", err
	}
	node, err := b.run(b.Node, []string{"--version"}, b.Env)
	if err != nil {
		return "", err
	}
	pins, err := os.ReadFile(filepath.Join(b.Module, "tools/toolchain_pins.json"))
	if err != nil {
		return "", err
	}
	var pinned struct {
		GoToolchain         string          `json:"go_toolchain"`
		CryptographicModule json.RawMessage `json:"cryptographic_module"`
	}
	if err := json.Unmarshal(pins, &pinned); err != nil {
		return "", err
	}
	data, err := json.Marshal(map[string]any{
		"schema_id": goBinaryKeySchemaID, "producer": b.ID(), "package": b.Package, "tags": b.Tags, "name": b.Name,
		"outputs": []string{b.Name, b.Name + ".crypto.json"}, "node_version": strings.TrimSpace(string(node)),
		"build_arguments": b.buildArgs(), "settings": effective, "inputs": inputs,
		"go_toolchain": pinned.GoToolchain, "cryptographic_module": pinned.CryptographicModule,
	})
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:]), nil
}

// Build verifies the pinned cryptographic module, builds the binary into
// staging and writes its receipt beside it, so a binary the receipt rejects
// never reaches a root.
func (b GoBinary) Build(staging string) error {
	env := append(slices.Clone(b.Env), "GO="+b.Go)
	if _, err := b.run(b.Node, []string{"tools/harness/readiness/cryptographic-build-cli.mjs", "verify"}, env); err != nil {
		return err
	}
	binary := filepath.Join(staging, b.Name)
	build := append(append([]string{b.Go}, b.buildArgs()...), "-o", binary, b.Package)
	if b.Wrap != nil {
		build = b.Wrap(build)
	}
	command := exec.Command(build[0], build[1:]...)
	command.Dir, command.Env, command.Stdout, command.Stderr = b.Module, b.Env, os.Stderr, os.Stderr
	if err := command.Run(); err != nil {
		return fmt.Errorf("build %s: %w", b.Package, err)
	}
	_, err := b.run(b.Node, []string{"tools/harness/readiness/cryptographic-build-cli.mjs", "receipt", binary, binary + ".crypto.json"}, env)
	return err
}
