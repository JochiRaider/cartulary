// sourcecheck is the private batch import parser used by harness inspection.
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"strings"

	"github.com/JochiRaider/cartulary/internal/testutil/sourcecheck"
)

type inputFile struct {
	Path   string `json:"path"`
	Source string `json:"source"`
}
type request struct {
	Files []inputFile `json:"files"`
}
type response struct {
	Files []sourcecheck.FileImports `json:"files"`
}

func run(input io.Reader, output io.Writer) error {
	decoder := json.NewDecoder(input)
	decoder.DisallowUnknownFields()
	var batch request
	if err := decoder.Decode(&batch); err != nil {
		return fmt.Errorf("decode import batch: %w", err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		return fmt.Errorf("import batch requires exactly one JSON object")
	}
	if batch.Files == nil {
		return fmt.Errorf("import batch requires files")
	}
	result := response{Files: make([]sourcecheck.FileImports, 0, len(batch.Files))}
	seen := make(map[string]bool, len(batch.Files))
	for _, file := range batch.Files {
		if !fs.ValidPath(file.Path) || strings.ContainsAny(file.Path, "\\\x00") || !strings.HasSuffix(file.Path, ".go") || seen[file.Path] {
			return fmt.Errorf("invalid or duplicate import path %q", file.Path)
		}
		seen[file.Path] = true
		imports, err := sourcecheck.ParseImports(file.Path, []byte(file.Source))
		if err != nil {
			return err
		}
		result.Files = append(result.Files, sourcecheck.FileImports{Path: file.Path, Imports: imports})
	}
	return json.NewEncoder(output).Encode(result)
}
func main() {
	if len(os.Args) != 1 {
		fmt.Fprintln(os.Stderr, "sourcecheck accepts a JSON batch on stdin")
		os.Exit(1)
	}
	if err := run(os.Stdin, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
