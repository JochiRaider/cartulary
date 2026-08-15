package performancefixture

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/JochiRaider/cartulary/internal/gen/performancefixtureprofile"
)

func TestRuntimeBundleIsSuiteRandomPrivateCopiedAndRemoved(t *testing.T) {
	t.Parallel()
	profile := runtimeTestProfile()
	snapshotKey := "abababababababababababababababababababababababababababababababab"
	first, err := generateRuntimeBundle(profile, snapshotKey, bytes.NewReader(entropySequence(0x11)))
	if err != nil {
		t.Fatal(err)
	}
	second, err := generateRuntimeBundle(profile, snapshotKey, bytes.NewReader(entropySequence(0x22)))
	if err != nil {
		t.Fatal(err)
	}
	if first.CredentialSets[0].Credentials[0] == second.CredentialSets[0].Credentials[0] {
		t.Fatal("independent suite entropy produced the same credential")
	}
	base := t.TempDir()
	sourceRoot := filepath.Join(base, "source")
	sourcePath, err := WriteRuntimeBundle(profile, sourceRoot, first)
	if err != nil {
		t.Fatal(err)
	}
	assertMode(t, sourceRoot, 0o700)
	assertMode(t, sourcePath, 0o600)
	destinationRoot := filepath.Join(base, "predicate")
	destinationPath, err := CopyRuntimeBundle(profile, sourcePath, destinationRoot)
	if err != nil {
		t.Fatal(err)
	}
	assertMode(t, destinationRoot, 0o700)
	assertMode(t, destinationPath, 0o600)
	copied, err := ReadRuntimeBundle(profile, destinationPath)
	if err != nil {
		t.Fatal(err)
	}
	if copied.CredentialSets[0].Credentials[23] != first.CredentialSets[0].Credentials[23] {
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
	profile := runtimeTestProfile()
	root := filepath.Join(t.TempDir(), "insecure")
	if err := os.Mkdir(root, 0o755); err != nil {
		t.Fatal(err)
	}
	bundle, err := GenerateRuntimeBundle(profile, "cdcdcdcdcdcdcdcdcdcdcdcdcdcdcdcdcdcdcdcdcdcdcdcdcdcdcdcdcdcdcdcd")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := WriteRuntimeBundle(profile, root, bundle); err == nil {
		t.Fatal("expected existing runtime root to fail closed")
	}
	privateRoot := filepath.Join(t.TempDir(), "private")
	path, err := WriteRuntimeBundle(profile, privateRoot, bundle)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(path, 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := ReadRuntimeBundle(profile, path); err == nil {
		t.Fatal("expected insecure runtime bundle mode to fail closed")
	}
}

func runtimeTestProfile() performancefixtureprofile.Profile {
	return performancefixtureprofile.Profile{
		FixtureProfileID: "synthetic_profile_v1",
		Status:           "active",
		ArtifactPolicy: performancefixtureprofile.ArtifactPolicy{
			RuntimeSchemaID: "cartulary.performance_fixture_runtime.v2",
		},
		RuntimeCredentialSets: []performancefixtureprofile.RuntimeCredentialSet{{
			SetID: "analysts", CredentialKind: "email_password", AccountCount: 24,
		}},
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
