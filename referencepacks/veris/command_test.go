package producer

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestStandaloneProducerCommand(t *testing.T) {
	for _, args := range [][]string{{"--candidate", "proposal.json"}, {"unexpected"}, {"--action", "unsupported"}} {
		var out, diagnostic bytes.Buffer
		if status := RunCommand(context.Background(), args, &out, &diagnostic); status != 1 || out.Len() != 0 || !strings.Contains(diagnostic.String(), "veris producer:") {
			t.Fatal("closed command dispatch", status, out.String(), diagnostic.String())
		}
	}
	var out, diagnostic bytes.Buffer
	if status := RunCommand(context.Background(), []string{"--help"}, &out, &diagnostic); status != 0 || out.Len() != 0 || !strings.Contains(diagnostic.String(), "Usage of veris:") {
		t.Fatal("help", status, out.String(), diagnostic.String())
	}
	directory := t.TempDir()
	report := filepath.Join(directory, "identity.json")
	out.Reset()
	diagnostic.Reset()
	if status := RunCommand(context.Background(), []string{"--action", "identity", "--report-file", report}, &out, &diagnostic); status != 0 || diagnostic.Len() != 0 {
		t.Fatal("identity", status, diagnostic.String())
	}
	raw, err := os.ReadFile(report)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(bytes.TrimSuffix(out.Bytes(), []byte{'\n'}), raw) {
		t.Fatal("delivered report differs from canonical stdout")
	}
	var builder Builder
	if err := json.Unmarshal(raw, &builder); err != nil || builder.ID != "cartulary.reference-pack.veris" || len(builder.SHA256) != 64 {
		t.Fatal("builder identity", err)
	}
	before := bytes.Clone(raw)
	out.Reset()
	diagnostic.Reset()
	if status := RunCommand(context.Background(), []string{"--action", "identity", "--report-file", report}, &out, &diagnostic); status != 0 {
		t.Fatal("repeat identity", status)
	}
	raw, err = os.ReadFile(report)
	if err != nil || !bytes.Equal(raw, before) {
		t.Fatal("identity changed without inputs", err)
	}
	out.Reset()
	diagnostic.Reset()
	if status := RunCommand(context.Background(), []string{"--action", "identity", "--report-file", directory}, &out, &diagnostic); status != 1 || out.Len() != 0 {
		t.Fatal("failed report delivery", status)
	}
	input := filepath.Join(directory, "retired.json")
	if err := os.WriteFile(input, []byte(`{"schema_id":"cartulary.base_registry_reissue_input.v1","output_file":"unused"}`), 0600); err != nil {
		t.Fatal(err)
	}
	forbidden := filepath.Join(directory, "forbidden-report.json")
	out.Reset()
	diagnostic.Reset()
	if status := RunCommand(context.Background(), []string{"--action", "assemble", "--input", input, "--report-file", forbidden}, &out, &diagnostic); status != 1 || out.Len() != 0 {
		t.Fatal("foreign/retired input admitted", status)
	}
	for _, action := range []string{"check", "audit"} {
		out.Reset()
		diagnostic.Reset()
		if status := RunCommand(context.Background(), []string{"--action", action, "--lock", input, "--directory", directory, "--built-at", "2026-11-08T18:30:00Z", "--report-file", forbidden}, &out, &diagnostic); status != 1 || out.Len() != 0 {
			t.Fatal("unbound source produced report", action, status)
		}
	}

	if _, err := os.Stat(forbidden); !os.IsNotExist(err) {
		t.Fatal("rejected input produced report", err)
	}
}
