package harnessdiag

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func testRejectionCodes(t *testing.T) {
	root := t.TempDir()
	file := filepath.Join(root, "secret-name")
	for _, tc := range []struct{ body, schema, stage string }{{"secret invalid", "", "decoding"}, {`{"schema_id":"foreign"}`, observationSchema, "schema"}} {
		if err := os.WriteFile(file, []byte(tc.body), 0600); err != nil {
			t.Fatal(err)
		}
		_, err := readObject(file, tc.schema, "live_snapshot", false, true)
		var rejection *observationRejection
		if !errors.As(err, &rejection) || rejection.stage != tc.stage {
			t.Fatal("lost boundary", err)
		}
		line := rejectionLine(err)
		if len(line) > 256 || strings.Count(line, "\n") != 1 || strings.Contains(line, "secret") || strings.Contains(line, root) {
			t.Fatal("unsafe rejection line")
		}
	}
	_, err := readObject(file+"-missing", "", "manifest", true, false)
	if err != nil {
		t.Fatal("optional absence changed", err)
	}
	line := rejectionLine(rejected("secret-role", "secret-stage", "secret-reason"))
	if strings.Contains(line, "secret") {
		t.Fatal("open diagnostic vocabulary")
	}
	if err := os.Chmod(root, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "run-manifest.json"), []byte("secret invalid"), 0600); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	if code := Main([]string{"--results-dir", root, "--json"}, &stdout, &stderr); code != 11 {
		t.Fatal("incorrect rejection exit", code)
	}
	value, err := Decode(stdout.Bytes())
	if err != nil || Validate(observationSchema, value) != nil {
		t.Fatal("rejection changed public JSON")
	}
	if stderr.String() != "[ARTIFACT] role=manifest stage=decoding reason=invalid_document\n" {
		t.Fatal("command did not retain one closed rejection")
	}
}
