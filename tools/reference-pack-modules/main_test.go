package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIndependentExecutableInputDiscovery(t *testing.T) {
	for _, source := range []string{
		"package p\nimport sys \"os\"\nvar discover = sys.Getwd",
		"package p\nimport \"os\"\nfunc f(){os.LookupEnv(\"SOURCE\")}",
		"package p\nimport . \"os\"\nvar discover = Getenv",
		"package p\nimport _ \"os/exec\"",
		"package p\nimport \"runtime\"\nvar source = runtime.Caller",
		"package p\n//go:generate read-application-input\n",
	} {
		if sourceInputs("producer.go", []byte(source)) == nil {
			t.Fatal("implicit input mechanism admitted")
		}
	}
	if err := sourceInputs("producer.go", []byte("package p\nimport \"os\"\nfunc explicit(name string){os.OpenRoot(name)}")); err != nil {
		t.Fatal(err)
	}
}

func TestIndependentModuleArchiveInputs(t *testing.T) {
	root, output := t.TempDir(), t.TempDir()
	write := func(name, value string) {
		t.Helper()
		p := filepath.Join(root, name)
		if err := os.MkdirAll(filepath.Dir(p), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(value), 0600); err != nil {
			t.Fatal(err)
		}
	}
	write("go.mod", "module github.com/JochiRaider/cartulary/referencepacks/example\n\ngo 1.27.3\n")
	write("example.go", "package example\n")
	write("go.sum", "")
	module := "github.com/JochiRaider/cartulary/referencepacks/example"
	first, err := export(root, module, "v0.1.0", output)
	if err != nil || len(first.Files) != 3 {
		t.Fatal("explicit archive", err)
	}
	write(".excluded/unused.json", "{}")
	second, err := export(root, module, "v0.1.0", output)
	if err != nil || first.SourceSHA256 != second.SourceSHA256 || first.ArchiveSHA256 != second.ArchiveSHA256 {
		t.Fatal("excluded input altered executable identity", err)
	}
	write("generate.go", "package example\n//go:generate fetch-source\n")
	if _, err := export(root, module, "v0.1.0", output); err == nil {
		t.Fatal("implicit generator admitted")
	}
	if err := os.Remove(filepath.Join(root, "generate.go")); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(root, "example.go"), filepath.Join(root, "linked.go")); err != nil {
		t.Fatal(err)
	}
	if _, err := export(root, module, "v0.1.0", output); err == nil {
		t.Fatal("linked archive input admitted")
	}
}
