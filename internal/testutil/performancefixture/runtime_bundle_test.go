package performancefixture

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestRuntimeBundleIsSuiteRandomPrivateCopiedAndRemoved(t *testing.T) {
	t.Parallel()
	snapshotKey := "abababababababababababababababababababababababababababababababab"
	first, err := generateRuntimeBundle(snapshotKey, bytes.NewReader(entropySequence(0x11)))
	if err != nil {
		t.Fatal(err)
	}
	second, err := generateRuntimeBundle(snapshotKey, bytes.NewReader(entropySequence(0x22)))
	if err != nil {
		t.Fatal(err)
	}
	if first.BackgroundAccounts[0] == second.BackgroundAccounts[0] {
		t.Fatal("independent suite entropy produced the same credential")
	}
	base := t.TempDir()
	sourceRoot := filepath.Join(base, "source")
	sourcePath, err := WriteRuntimeBundle(sourceRoot, first)
	if err != nil {
		t.Fatal(err)
	}
	assertMode(t, sourceRoot, 0o700)
	assertMode(t, sourcePath, 0o600)
	destinationRoot := filepath.Join(base, "predicate")
	destinationPath, err := CopyRuntimeBundle(sourcePath, destinationRoot)
	if err != nil {
		t.Fatal(err)
	}
	assertMode(t, destinationRoot, 0o700)
	assertMode(t, destinationPath, 0o600)
	copied, err := ReadRuntimeBundle(destinationPath)
	if err != nil {
		t.Fatal(err)
	}
	if copied.BackgroundAccounts[23] != first.BackgroundAccounts[23] {
		t.Fatal("predicate runtime copy changed credential material")
	}
	if err := RemoveRuntimeBundle(destinationRoot); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(destinationRoot); !os.IsNotExist(err) {
		t.Fatalf("predicate credential root remained after cleanup: %v", err)
	}
	if err := RemoveRuntimeBundle(sourceRoot); err != nil {
		t.Fatal(err)
	}
}

func entropySequence(seed byte) []byte {
	result := make([]byte, 48*24)
	for index := range result {
		result[index] = seed + byte(index/48)
	}
	return result
}

func TestRuntimeBundleRejectsInsecureAndUnknownContent(t *testing.T) {
	t.Parallel()
	root := filepath.Join(t.TempDir(), "insecure")
	if err := os.Mkdir(root, 0o755); err != nil {
		t.Fatal(err)
	}
	bundle, err := GenerateRuntimeBundle("cdcdcdcdcdcdcdcdcdcdcdcdcdcdcdcdcdcdcdcdcdcdcdcdcdcdcdcdcdcdcdcd")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := WriteRuntimeBundle(root, bundle); err == nil {
		t.Fatal("expected existing runtime root to fail closed")
	}
	privateRoot := filepath.Join(t.TempDir(), "private")
	path, err := WriteRuntimeBundle(privateRoot, bundle)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(path, 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := ReadRuntimeBundle(path); err == nil {
		t.Fatal("expected insecure runtime bundle mode to fail closed")
	}
}

func assertMode(t *testing.T, path string, want os.FileMode) {
	t.Helper()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if got := info.Mode().Perm(); got != want {
		t.Fatalf("%s mode=%#o want=%#o", path, got, want)
	}
}
