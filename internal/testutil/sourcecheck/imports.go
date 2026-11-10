// Package sourcecheck inspects Go import syntax without owning dependency policy.
package sourcecheck

import (
	"fmt"
	"go/parser"
	"go/token"
	"io/fs"
	"sort"
	"strconv"
	"strings"
)

type FileImports struct {
	Path    string   `json:"path"`
	Imports []string `json:"imports"`
}

// ParseImports decodes every import declaration. Build constraints do not filter
// source, and function bodies are outside this inspection's syntax boundary.
func ParseImports(path string, source []byte) ([]string, error) {
	parsed, err := parser.ParseFile(token.NewFileSet(), path, source, parser.ImportsOnly)
	if err != nil {
		return nil, fmt.Errorf("parse imports %s: %w", path, err)
	}
	imports := make([]string, 0, len(parsed.Imports))
	for _, spec := range parsed.Imports {
		value, err := strconv.Unquote(spec.Path.Value)
		if err != nil {
			return nil, fmt.Errorf("decode import %s: %w", path, err)
		}
		imports = append(imports, value)
	}
	return imports, nil
}

// ReadImports walks the caller-rooted filesystem without following links. Only
// selected production Go files are opened; callers apply package/support policy.
func ReadImports(source fs.FS) ([]FileImports, error) {
	return ReadImportsUnder(source, []string{"."}, nil)
}

// ReadImportsUnder scans explicit roots, pruning caller-owned exclusions before
// opening them. Policies remain with the caller; build constraints never omit files.
func ReadImportsUnder(source fs.FS, roots []string, exclude func(string) bool) ([]FileImports, error) {
	files := make([]FileImports, 0)
	for _, root := range roots {
		if !fs.ValidPath(root) {
			return nil, fmt.Errorf("invalid import root %q", root)
		}
		err := fs.WalkDir(source, root, func(path string, entry fs.DirEntry, err error) error {
			if err != nil {
				return fmt.Errorf("walk imports %s: %w", path, err)
			}
			if exclude != nil && exclude(path) {
				if entry.IsDir() {
					return fs.SkipDir
				}
				return nil
			}
			if entry.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
				return nil
			}
			if !entry.Type().IsRegular() {
				return fmt.Errorf("inspect imports %s: not a regular file", path)
			}
			data, err := fs.ReadFile(source, path)
			if err != nil {
				return fmt.Errorf("read imports %s: %w", path, err)
			}
			imports, err := ParseImports(path, data)
			if err != nil {
				return err
			}
			files = append(files, FileImports{Path: path, Imports: imports})
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	sort.Slice(files, func(i, j int) bool { return files[i].Path < files[j].Path })
	return files, nil
}
