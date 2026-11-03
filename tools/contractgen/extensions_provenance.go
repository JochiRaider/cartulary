package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"unicode/utf8"
)

// These are code and build-control inputs to the Extensions producer. Runtime
// contracts have their own digest members; other catalogued generation inputs
// are added from the explicit input allowlist below, never directory discovery.
var extensionProducerSources = []string{
	"contracts/index.json", "contracts/extensions/index.json", "go.mod", "go.sum",
	"internal/platform/canonicaljson/canonicaljson.go", "internal/platform/canonicaljson/extension.go", "internal/platform/canonicaljson/strict.go",
	"internal/platform/extensioncodec/definition.go", "internal/platform/extensioncodec/postgres.go", "internal/platform/extensioncodec/proof.go",
	"tools/contractgen/extensions_codecs.go", "tools/contractgen/extensions_generation.go", "tools/contractgen/extensions_provenance.go",
	"tools/contractgen/extensions_reporting_source.go", "tools/contractgen/extensions_transaction.go", "tools/contractgen/extensions_validation.go", "tools/contractgen/families.go", "tools/contractgen/main.go", "tools/contractgen/validation.go",
	"internal/platform/jobs/jobcontract/identity.go", "internal/platform/jobs/jobcontract/terminal.go", "internal/platform/jobs/jobcontract/resources.go", "tools/contractgen/job_validation.go",
	"internal/modules/extensions/schemarefs/references.go",
	"internal/modules/extensions/registryfacts/facts.go",
	"internal/modules/extensions/registryfacts/owner_inputs.go",
	"internal/modules/extensions/registryfacts/limits.go",
	"internal/modules/extensions/registryfacts/collisions.go",
	"internal/modules/extensions/registryfacts/dependencies.go",
	"internal/modules/extensions/registryfacts/routes.go",
	"tools/contractgen/extensions_owner_diagnostics.go",
	"tools/toolchain_pins.json",
}

func extensionGeneratorSources(root string, indexed map[string]map[string]any) ([]map[string]any, error) {
	refs := append([]string(nil), extensionProducerSources...)
	for path := range indexed {
		full := "contracts/extensions/" + path
		if path != "dependencies.json" && !extensionRuntimeArtifactPath(full) {
			refs = append(refs, full)
		}
	}
	sort.Strings(refs)
	if len(refs) == 0 || len(refs) > 1024 {
		return nil, fmt.Errorf("extensions generator source count is invalid")
	}
	rows := make([]map[string]any, 0, len(refs))
	for i, ref := range refs {
		if i > 0 && refs[i-1] == ref {
			return nil, fmt.Errorf("duplicate Extensions generator source")
		}
		body, err := readExtensionSourceFile(root, ref)
		if err != nil {
			return nil, err
		}
		sum := sha256.Sum256(body)
		rows = append(rows, map[string]any{"source_ref": ref, "source_sha256": hex.EncodeToString(sum[:])})
	}
	return rows, nil
}

func readExtensionSourceFile(root, ref string) ([]byte, error) {
	allowedKind := ref == "go.mod" || ref == "go.sum" || strings.HasPrefix(ref, "contracts/") && strings.HasSuffix(ref, ".json") || strings.HasPrefix(ref, "internal/") && strings.HasSuffix(ref, ".go") || strings.HasPrefix(ref, "tools/") && (strings.HasSuffix(ref, ".go") || strings.HasSuffix(ref, ".json"))
	if !allowedKind || len(ref) == 0 || len(ref) > 512 || !utf8.ValidString(ref) || strings.ContainsAny(ref, "\\\x00") || strings.HasPrefix(ref, "/") {
		return nil, fmt.Errorf("invalid Extensions source path")
	}
	parts := strings.Split(ref, "/")
	for _, part := range parts {
		if part == "" || part == "." || part == ".." {
			return nil, fmt.Errorf("invalid Extensions source segment")
		}
	}
	capability, err := os.OpenRoot(root)
	if err != nil {
		return nil, err
	}
	defer capability.Close()
	for i := range parts {
		info, err := capability.Lstat(strings.Join(parts[:i+1], "/"))
		if err != nil {
			return nil, err
		}
		if info.Mode()&os.ModeSymlink != 0 || i < len(parts)-1 && !info.IsDir() || i == len(parts)-1 && !info.Mode().IsRegular() {
			return nil, fmt.Errorf("extensions source must be a regular non-symlink path")
		}
	}
	before, err := capability.Lstat(ref)
	if err != nil {
		return nil, err
	}
	file, err := capability.Open(ref)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	opened, err := file.Stat()
	if err != nil {
		return nil, err
	}
	if !opened.Mode().IsRegular() || !os.SameFile(before, opened) {
		return nil, fmt.Errorf("extensions source changed while opening")
	}
	data, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}
	after, err := capability.Lstat(ref)
	if err != nil {
		return nil, err
	}
	if after.Mode()&os.ModeSymlink != 0 || !os.SameFile(opened, after) || after.Size() != opened.Size() || !after.ModTime().Equal(opened.ModTime()) {
		return nil, fmt.Errorf("extensions source changed while reading")
	}
	return data, nil
}
