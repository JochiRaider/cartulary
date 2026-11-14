package performancefixture

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
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
	if err := os.Chmod(path, 0o600); err != nil {
		t.Fatal(err)
	}
	raw, err := json.Marshal(bundle)
	if err != nil {
		t.Fatal(err)
	}
	valid := string(raw)
	privateMarker := bundle.CredentialSets[0].Credentials[0].Secret
	for name, document := range map[string]string{
		"unknown_private_field":   strings.Replace(valid, `"schema_id":`, `"`+privateMarker+`":null,"schema_id":`, 1),
		"nested_unknown_field":    strings.Replace(valid, `"principal":`, `"`+privateMarker+`":null,"principal":`, 1),
		"duplicate_member":        strings.Replace(valid, `"schema_id":`, `"schema_id":"ignored","schema_id":`, 1),
		"nested_duplicate_member": strings.Replace(valid, `"principal":`, `"principal":"ignored","principal":`, 1),
		"wrong_field_case":        strings.Replace(valid, `"schema_id":`, `"SCHEMA_ID":`, 1),
		"nested_wrong_field_case": strings.Replace(valid, `"secret":`, `"SECRET":`, 1),
		"missing_required_member": strings.Replace(valid, `"schema_id":"cartulary.performance_fixture_runtime.v2",`, "", 1),
		"null_required_member":    strings.Replace(valid, `"schema_id":"cartulary.performance_fixture_runtime.v2"`, `"schema_id":null`, 1),
		"wrong_schema":            strings.Replace(valid, "cartulary.performance_fixture_runtime.v2", "cartulary.performance_fixture_runtime.v1", 1),
		"wrong_profile":           strings.Replace(valid, "synthetic_profile_v1", "another_profile_v1", 1),
		"wrong_type":              strings.Replace(valid, `"secret":"`+privateMarker+`"`, `"secret":1`, 1),
		"second_document":         valid + `{}`,
		"trailing_garbage":        valid + "private-trailing-content",
		"null_document":           "null",
	} {
		t.Run(name, func(t *testing.T) {
			if err := os.WriteFile(path, []byte(document), 0o600); err != nil {
				t.Fatal(err)
			}
			if _, err := ReadRuntimeBundle(profile, path); err == nil {
				t.Fatal("invalid runtime document was admitted")
			} else if strings.Contains(err.Error(), privateMarker) || strings.Contains(err.Error(), bundle.CredentialSets[0].Credentials[0].Principal) {
				t.Fatal("runtime admission diagnostic disclosed private content")
			}
		})
	}
	if err := os.WriteFile(path, []byte(" \n"+valid+"\n\t"), 0o600); err != nil {
		t.Fatal(err)
	}
	if admitted, err := ReadRuntimeBundle(profile, path); err != nil || admitted.CredentialSets[0].Credentials[0] != bundle.CredentialSets[0].Credentials[0] {
		t.Fatal("valid runtime document with whitespace was not preserved")
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
