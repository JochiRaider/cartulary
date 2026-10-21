//go:build linux

package harnessdiag

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestDiagnosticSchema(t *testing.T) {
	if err := Validate(observationSchema, emptyObservation()); err != nil {
		t.Fatal(err)
	}
	for _, text := range []string{`{"x":1,"x":2}`, `-0`, `9007199254740992`, `"\ud800"`, `{"x":NaN}`, `{} {}`} {
		if _, err := Decode([]byte(text)); err == nil {
			t.Fatalf("accepted %s", text)
		}
	}
	value, err := Decode([]byte(`{"x":1.25,"unicode":"\ud83d\ude00"}`))
	if err != nil || value.(object)["x"] != 1.25 {
		t.Fatalf("decode: %v %v", value, err)
	}
	r := emptyObservation()
	r["unexpected"] = true
	if Validate(observationSchema, r) == nil {
		t.Fatal("accepted extra field")
	}
	if Validate("cartulary.harness_run_observation.v1", emptyObservation()) == nil {
		t.Fatal("accepted obsolete schema")
	}
	for _, path := range []string{"a/b", "a_b/long-name.1", "a@b"} {
		if !schemaPattern(`^(?!/)(?!.*(?:^|/)\.\.(?:/|$))(?!.*(?:^|/)\.(?:/|$))(?!.*//)[A-Za-z0-9._@+,:=-]+(?:/[A-Za-z0-9._@+,:=-]+)*$`, path) {
			t.Fatalf("rejected %s", path)
		}
	}
	for _, path := range []string{"/a", "a/../b", "a/./b", "a//b", "a/"} {
		if schemaPattern(`^(?!/)(?!.*(?:^|/)\.\.(?:/|$))(?!.*(?:^|/)\.(?:/|$))(?!.*//)[A-Za-z0-9._@+,:=-]+(?:/[A-Za-z0-9._@+,:=-]+)*$`, path) {
			t.Fatalf("accepted %s", path)
		}
	}
}
func TestDiagnosticSelection(t *testing.T) {
	root := t.TempDir()
	missing := filepath.Join(root, "missing")
	r := Observe(context.Background(), Inputs{ResultsDir: missing})
	if r["observation"] != "not_published" || r["operation_exit_code"] != float64(0) {
		t.Fatal(r)
	}
	if _, err := os.Stat(missing); !os.IsNotExist(err) {
		t.Fatal("reader created directory")
	}
	for _, in := range []Inputs{{ResultsDir: root, WaitSeconds: "1"}, {ResultsDir: root, AfterRevision: "-1"}, {ResultsDir: "../invalid"}, {ResultsDir: root, Target: "/invalid"}} {
		if r := Observe(context.Background(), in); r["failure_reason"] != "usage_error" {
			t.Fatal(r)
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	r = Observe(ctx, Inputs{ResultsDir: missing})
	if r["operation_exit_code"] != float64(15) {
		t.Fatal(r)
	}
}
func TestDiagnosticSecureFiles(t *testing.T) {
	t.Run("closed_rejection_codes", testRejectionCodes)
	root := t.TempDir()
	file := filepath.Join(root, "data")
	if err := os.WriteFile(file, []byte(`{}`), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := readLocal(file, 2, false); err != nil {
		t.Fatal(err)
	}
	if _, err := readLocal(file, 1, false); err == nil {
		t.Fatal("accepted exhausted bound")
	}
	link := filepath.Join(root, "link")
	if err := os.Symlink(file, link); err != nil {
		t.Fatal(err)
	}
	if _, err := readLocal(link, 2, false); err == nil {
		t.Fatal("accepted symlink")
	}
	ancestor := filepath.Join(root, "ancestor")
	if err := os.Symlink(root, ancestor); err != nil {
		t.Fatal(err)
	}
	if _, err := readLocal(filepath.Join(ancestor, "data"), 2, false); err == nil {
		t.Fatal("accepted symlink ancestor")
	}
	hard := filepath.Join(root, "hard")
	if err := os.Link(file, hard); err != nil {
		t.Fatal(err)
	}
	if _, err := readLocal(file, 2, false); err == nil {
		t.Fatal("accepted hardlink")
	}
	if err := os.Remove(hard); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(file, 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := readLocal(file, 2, false); err == nil {
		t.Fatal("accepted public file")
	}
}
