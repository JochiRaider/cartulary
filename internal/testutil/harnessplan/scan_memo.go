package harnessplan

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

// scanMemoPath keeps what the planner scans out of source files, by content
// digest, between plans of one checkout (P6-5). The graph cache directory is
// machine state; the memo is never an input to a key.
const scanMemoPath = GraphCacheDirectory + "/source-scans.json"

const scanMemoSchemaID = "cartulary.harness_source_scan_memo.v1"

// scannerIdentity names the scanning rules a memo entry was made with, so a
// change to any pattern discards every entry.
var scannerIdentity = func() string {
	sum := sha256.Sum256([]byte(strings.Join([]string{
		scanMemoSchemaID, moduleFrom.String(), moduleBare.String(), moduleDynamic.String(),
		moduleRequire.String(), moduleComputed.String(), goPathLiteral.String(),
	}, "\x00")))
	return hex.EncodeToString(sum[:])
}()

// moduleScan is what a Node module's text says about its imports: whether it
// has a computed dynamic import, and its relative specifiers in pattern order.
type moduleScan struct {
	Computed   bool     `json:"computed"`
	Specifiers []string `json:"specifiers"`
}

// sourceScans memoizes each scan by the scanned file's content digest. The
// digest is the snapshot's own, so an entry applies exactly to the bytes it
// was made from; only resolution against the snapshot happens per plan.
type sourceScans struct {
	SchemaID string                `json:"schema_id"`
	Scanner  string                `json:"scanner"`
	Modules  map[string]moduleScan `json:"modules"`
	// Literals are a Go file's path-shaped string literals, before they are
	// matched against the snapshot.
	Literals map[string][]string `json:"literals"`
	changed  bool
	// kept says the memo is read and written; otherwise it lives for one plan.
	kept bool
}

// loadSourceScans reads the memo when the checkout keeps one. A missing,
// unreadable or foreign memo is empty.
func loadSourceScans(root string, kept bool) *sourceScans {
	empty := &sourceScans{SchemaID: scanMemoSchemaID, Scanner: scannerIdentity, Modules: map[string]moduleScan{}, Literals: map[string][]string{}, kept: kept}
	if !kept {
		return empty
	}
	data, err := readBounded(filepath.Join(root, scanMemoPath))
	if err != nil {
		return empty
	}
	var memo sourceScans
	if json.Unmarshal(data, &memo) != nil || memo.SchemaID != scanMemoSchemaID || memo.Scanner != scannerIdentity || memo.Modules == nil || memo.Literals == nil {
		return empty
	}
	memo.kept = true
	return &memo
}

// save keeps only entries for content in the current snapshot and writes the
// memo when this plan added one. A memo that cannot be written is left as it
// was: it only saves work.
func (m *sourceScans) save(root string, source []SourceInput) {
	if !m.kept || !m.changed {
		return
	}
	current := make(map[string]bool, len(source))
	for _, input := range source {
		current[input.ByteDigest] = true
	}
	for digest := range m.Modules {
		if !current[digest] {
			delete(m.Modules, digest)
		}
	}
	for digest := range m.Literals {
		if !current[digest] {
			delete(m.Literals, digest)
		}
	}
	data, err := json.Marshal(m)
	if err != nil {
		return
	}
	path := filepath.Join(root, scanMemoPath)
	if os.MkdirAll(filepath.Dir(path), 0o755) == nil {
		_ = writeAtomic(path, data)
	}
}

// moduleScan returns a snapshot module's scan, from the memo when its content
// was scanned before.
func (p *cachePlanner) moduleScan(index int) (moduleScan, error) {
	digest := p.source[index].ByteDigest
	if scan, ok := p.scans.Modules[digest]; ok {
		return scan, nil
	}
	source, err := p.text(p.source[index].Path)
	if err != nil {
		return moduleScan{}, err
	}
	scan := moduleScan{Computed: moduleComputed.MatchString(source), Specifiers: []string{}}
	for _, pattern := range modulePatterns {
		for _, match := range pattern.FindAllStringSubmatch(source, -1) {
			if specifier := match[1]; strings.HasPrefix(specifier, "./") || strings.HasPrefix(specifier, "../") {
				scan.Specifiers = append(scan.Specifiers, specifier)
			}
		}
	}
	p.scans.Modules[digest], p.scans.changed = scan, true
	return scan, nil
}

// goLiterals returns a snapshot Go file's path-shaped string literals, from
// the memo when its content was scanned before.
func (p *cachePlanner) goLiterals(index int) ([]string, error) {
	digest := p.source[index].ByteDigest
	if literals, ok := p.scans.Literals[digest]; ok {
		return literals, nil
	}
	source, err := p.text(p.source[index].Path)
	if err != nil {
		return nil, err
	}
	literals := []string{}
	for _, match := range goPathLiteral.FindAllStringSubmatch(source, -1) {
		literals = append(literals, match[1])
	}
	p.scans.Literals[digest], p.scans.changed = literals, true
	return literals, nil
}
