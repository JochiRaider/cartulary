package codegenboundary_test

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/JochiRaider/cartulary/internal/testutil/sourcecheck"
)

const testutilImportPathPrefix = repoImportPrefix + "internal/testutil"
const revisionsSupportImportPath = testutilImportPathPrefix + "/revisionsupport"

func TestRuntimeGoFilesDoNotImportInternalTestutil(t *testing.T) {
	t.Run("scanner fixtures", func(t *testing.T) {
		fixture := fstest.MapFS{
			"cmd/a/main.go":                         {Data: []byte("package main\nimport alias `" + testutilImportPathPrefix + "/child`")},
			"internal/modules/x/deep/runtime.go":    {Data: []byte("//go:build never_enabled\n\npackage x\nimport (\n_ \"" + testutilImportPathPrefix + "\"\n)")},
			"internal/modules/x/comment.go":         {Data: []byte("package x\n// import `" + testutilImportPathPrefix + "`\nimport `" + testutilImportPathPrefix + "suffix`")},
			"internal/modules/x/testsupport/bad.go": {Data: []byte("invalid")},
			"internal/gen/bad.go":                   {Data: []byte("invalid")},
			"internal/testutil/bad.go":              {Data: []byte("invalid")},
			"internal/testharness/bad.go":           {Data: []byte("invalid")},
			"internal/platform/contracttest/bad.go": {Data: []byte("invalid")},
			"internal/modules/x/ignored_test.go":    {Data: []byte("invalid")},
			"outside/poison.go":                     {Data: []byte("invalid")},
		}
		files, err := runtimeImports(fixture)
		if err != nil {
			t.Fatal(err)
		}
		got := strings.Join(importOffenders(files, testutilImportPathPrefix), "\n")
		if got != "cmd/a/main.go\ninternal/modules/x/deep/runtime.go" {
			t.Fatalf("offenders: %s", got)
		}
		fixture["internal/modules/x/broken.go"] = &fstest.MapFile{Data: []byte("package x\nimport (")}
		if _, err := runtimeImports(fixture); err == nil {
			t.Fatal("malformed source accepted")
		}
		delete(fixture, "internal/modules/x/broken.go")
		fixture["cmd/link.go"] = &fstest.MapFile{Mode: fs.ModeSymlink, Data: []byte("elsewhere.go")}
		if _, err := runtimeImports(fixture); err == nil {
			t.Fatal("linked source accepted")
		}
		delete(fixture, "cmd/link.go")
		delete(fixture, "cmd/a/main.go")
		if _, err := runtimeImports(fixture); err == nil {
			t.Fatal("missing root accepted")
		}
	})

	offenders := runtimeGoImportOffenders(t, testutilImportPathPrefix)
	if len(offenders) > 0 {
		t.Fatalf("runtime Go files must not import internal/testutil:\n%s", strings.Join(offenders, "\n"))
	}
}

// The narrow witness preserves Revisions-only selection under single-owner
// catalog routing; mechanics and production scanning remain shared.
func TestRuntimeGoFilesDoNotImportRevisionsSupport(t *testing.T) {
	if _, err := runtimeImports(fstest.MapFS{"cmd": {Mode: fs.ModeDir}, "internal": {Mode: fs.ModeDir}}); err == nil {
		t.Fatal("empty runtime scope accepted")
	}
	files := []sourcecheck.FileImports{{Path: "internal/modules/peer/runtime.go", Imports: []string{revisionsSupportImportPath + "/child"}}}
	if got := importOffenders(files, revisionsSupportImportPath); len(got) != 1 {
		t.Fatal("Revisions witness did not reject support descendant")
	}
	offenders := runtimeGoImportOffenders(t, revisionsSupportImportPath)
	if len(offenders) > 0 {
		t.Fatalf("runtime Go files must not import Revisions test support:\n%s", strings.Join(offenders, "\n"))
	}
}

func runtimeGoImportOffenders(t *testing.T, importPathPrefix string) []string {
	t.Helper()
	files, err := runtimeImports(os.DirFS(filepath.Join("..", "..", "..")))
	if err != nil {
		t.Fatal(err)
	}
	return importOffenders(files, importPathPrefix)
}

func runtimeImports(source fs.FS) ([]sourcecheck.FileImports, error) {
	files, err := sourcecheck.ReadImportsUnder(source, []string{"cmd", "internal"}, func(path string) bool {
		return path == "internal/gen" || path == "internal/testutil" || path == "internal/testharness" ||
			path == "internal/platform/contracttest" ||
			(strings.HasPrefix(path, "internal/modules/") && strings.HasSuffix(path, "/testsupport"))
	})
	if err == nil && len(files) == 0 {
		return nil, fmt.Errorf("runtime import scope contains no production Go files")
	}
	return files, err
}

func importOffenders(files []sourcecheck.FileImports, prefix string) []string {
	var offenders []string
	for _, file := range files {
		for _, imported := range file.Imports {
			if imported == prefix || strings.HasPrefix(imported, prefix+"/") {
				offenders = append(offenders, file.Path)
				break
			}
		}
	}
	return offenders
}
