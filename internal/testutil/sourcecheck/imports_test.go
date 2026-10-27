package sourcecheck

import (
	"errors"
	"io/fs"
	"reflect"
	"strings"
	"testing"
	"testing/fstest"
)

func TestParseImports(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name, source string
		want         []string
	}{
		{"empty", "package fixture", []string{}},
		{"all declarations", "package fixture\nimport alias \"example/first\"\nimport (\n _ `example/second`\n . \"example/third\"\n)\nimport (\"example/fourth\"; renamed \"example/fi\\x66th\")", []string{"example/first", "example/second", "example/third", "example/fourth", "example/fifth"}},
		{"comments and body strings", "package fixture\n// import \"fake/comment\"\n/* import (`fake/block`) */\nimport /* gap */ \"real/owner\"\nconst sample = `import (\"fake/string\")`", []string{"real/owner"}},
		{"host independent", "//go:build impossible_sourcecheck_tag\n\npackage fixture\nimport \"example/tagged\"", []string{"example/tagged"}},
	} {
		t.Run(test.name, func(t *testing.T) {
			got, err := ParseImports("fixture.go", []byte(test.source))
			if err != nil || !reflect.DeepEqual(got, test.want) {
				t.Fatalf("imports = %v, %v; want %v", got, err, test.want)
			}
		})
	}
	for _, source := range []string{"not a package", "package p\nimport (", "package p\nimport \"unterminated", "package p\nimport \"bad\\q\""} {
		if got, err := ParseImports("bad.go", []byte(source)); err == nil || got != nil || !strings.Contains(err.Error(), "bad.go") {
			t.Fatalf("malformed imports = %v, %v", got, err)
		}
	}
}

type observedFS struct {
	fs.FS
	opened *[]string
	fail   string
}

func (f observedFS) Open(path string) (fs.File, error) {
	*f.opened = append(*f.opened, path)
	if path == f.fail {
		return nil, fs.ErrPermission
	}
	return f.FS.Open(path)
}

func TestReadImports(t *testing.T) {
	t.Parallel()
	opened := []string{}
	source := observedFS{FS: fstest.MapFS{
		"z.go":             {Data: []byte("package z\nimport `z/owner`")},
		"child/a.go":       {Data: []byte("package child\nimport \"child/owner\"")},
		"a.go":             {Data: []byte("package a")},
		"ignored_test.go":  {Data: []byte("malformed")},
		"README.md":        {Data: []byte("not source")},
		"child/spec.md":    {Data: []byte("not source")},
		"linked-directory": {Mode: fs.ModeSymlink, Data: []byte("external")},
	}, opened: &opened}
	got, err := ReadImports(source)
	want := []FileImports{{Path: "a.go", Imports: []string{}}, {Path: "child/a.go", Imports: []string{"child/owner"}}, {Path: "z.go", Imports: []string{"z/owner"}}}
	if err != nil || !reflect.DeepEqual(got, want) {
		t.Fatalf("files = %#v, %v", got, err)
	}
	for _, path := range opened {
		if strings.HasSuffix(path, ".md") || strings.HasSuffix(path, "_test.go") || path == "linked-directory" {
			t.Fatalf("opened excluded entry %s", path)
		}
	}
	if got, err := ReadImports(fstest.MapFS{}); err != nil || len(got) != 0 {
		t.Fatalf("empty scope = %v, %v", got, err)
	}
}

func TestReadImportsFailures(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name string
		mode fs.FileMode
		data string
		fail string
	}{
		{"parse", 0, "package p\nimport (", ""},
		{"read", 0, "package p", "z.go"},
		{"walk", 0, "package p", "."},
		{"symlink", fs.ModeSymlink, "external.go", ""},
		{"device", fs.ModeDevice, "", ""},
	} {
		t.Run(test.name, func(t *testing.T) {
			opened := []string{}
			source := observedFS{FS: fstest.MapFS{"a.go": {Data: []byte("package a")}, "z.go": {Mode: test.mode, Data: []byte(test.data)}}, opened: &opened, fail: test.fail}
			got, err := ReadImports(source)
			if err == nil || got != nil {
				t.Fatalf("partial success: %v, %v", got, err)
			}
			if test.fail != "" && !errors.Is(err, fs.ErrPermission) {
				t.Fatalf("read error lost: %v", err)
			}
		})
	}
}
