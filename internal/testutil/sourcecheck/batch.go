package sourcecheck

import (
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"strings"
)

type inputFile struct {
	Path   string `json:"path"`
	Source string `json:"source"`
}
type request struct {
	Files               []inputFile `json:"files"`
	IncludeDeclarations bool        `json:"include_declarations,omitempty"`
	IncludeSQL          bool        `json:"include_sql,omitempty"`
}
type outputFile struct {
	Path         string            `json:"path"`
	Imports      []string          `json:"imports"`
	Declarations *[]Declaration    `json:"declarations,omitempty"`
	SQL          *[]SQLDeclaration `json:"sql,omitempty"`
}

type response struct {
	Files []outputFile `json:"files"`
}

// InspectBatch serves the private source-inspection protocol for command facades.
func InspectBatch(input io.Reader, output io.Writer) error {
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
	result := response{Files: make([]outputFile, 0, len(batch.Files))}
	seen := make(map[string]bool, len(batch.Files))
	for _, file := range batch.Files {
		if !fs.ValidPath(file.Path) || strings.ContainsAny(file.Path, "\\\x00") || !strings.HasSuffix(file.Path, ".go") || seen[file.Path] {
			return fmt.Errorf("invalid or duplicate import path %q", file.Path)
		}
		seen[file.Path] = true

	}
	var sql map[string][]SQLDeclaration
	if batch.IncludeSQL {
		sources := map[string]string{}
		for _, file := range batch.Files {
			sources[file.Path] = file.Source
		}
		var err error
		sql, err = AnalyzeSQL(sources)
		if err != nil {
			return err
		}
	}
	for _, file := range batch.Files {
		imports, err := ParseImports(file.Path, []byte(file.Source))
		if err != nil {
			return err
		}
		entry := outputFile{Path: file.Path, Imports: imports}
		if batch.IncludeSQL {
			declarations := sql[file.Path]
			entry.SQL = &declarations
		}
		if batch.IncludeDeclarations {
			declarations, err := ParseDeclarations(file.Path, []byte(file.Source))
			if err != nil {
				return err
			}
			entry.Declarations = &declarations
		}
		result.Files = append(result.Files, entry)
	}
	return json.NewEncoder(output).Encode(result)
}
