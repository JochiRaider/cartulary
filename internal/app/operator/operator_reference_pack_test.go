package operator

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/app/configassembly"
	"github.com/JochiRaider/cartulary/internal/app/referenceassembly"
	"github.com/JochiRaider/cartulary/internal/modules/reference_data"
	"github.com/google/uuid"
)

type referencePackImportFunc func(context.Context, uuid.UUID, string) referencePackOperatorResult

func (f referencePackImportFunc) importAndObserve(ctx context.Context, id uuid.UUID, name string) referencePackOperatorResult {
	return f(ctx, id, name)
}

func TestOperatorReferencePackClosedTransport_Unit(t *testing.T) {
	t.Run("typed staging rejection", testOperatorStagingRejection)
	id := uuid.MustParse("60000000-0000-4000-8000-000000000001")
	for _, tc := range []struct {
		name  string
		args  []string
		phase string
		exit  int
	}{
		{"missing", []string{"reference-pack", "import"}, "syntax", 2},
		{"extra", []string{"reference-pack", "import", "a.zip", "extra"}, "syntax", 2},
		{"escape", []string{"reference-pack", "import", "../SECRET"}, "syntax", 2},
		{"leading dot", []string{"reference-pack", "import", ".a"}, "syntax", 2},
		{"too long", []string{"reference-pack", "import", strings.Repeat("a", 129)}, "syntax", 2},
		{"config failure", []string{"reference-pack", "import", "pack.zip"}, "config", 3},
		{"open failure", []string{"reference-pack", "import", "pack.zip"}, "open", 3},
		{"rejection", []string{"reference-pack", "import", "pack.zip"}, "rejected", 3},
		{"success", []string{"reference-pack", "import", strings.Repeat("a", 128)}, "success", 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			loads, opens, imports, closes := 0, 0, 0, 0
			e := referencePackExecutor{transport: operatorTransport{stdout: &stdout, stderr: &stderr}, newOperationID: func() uuid.UUID { return id }, loadConfig: func(string) (configassembly.Loaded, error) {
				loads++
				if tc.phase == "config" {
					return configassembly.Loaded{}, errors.New("SECRET /private/path")
				}
				return configassembly.Loaded{}, nil
			}}
			e.open = func(context.Context, configassembly.Loaded) (referencePackLocalImport, func(), error) {
				opens++
				if tc.phase == "open" {
					return nil, nil, errors.New("SECRET database credential")
				}
				return referencePackImportFunc(func(_ context.Context, got uuid.UUID, name string) referencePackOperatorResult {
					imports++
					if got != id || name != tc.args[2] {
						t.Fatal("operation identity or name changed")
					}
					r := referencePackOperatorResult{SchemaID: "cartulary.reference_pack_operator_result.v1", OperationID: id.String(), Result: "succeeded"}
					if tc.phase == "rejected" {
						r.Result = "failed"
						code, reason := "reference_pack_operation_rejected", "clock_untrusted"
						r.ErrorCode = &code
						r.ReasonCode = &reason
					}
					return r
				}), func() { closes++ }, nil
			}
			if exit := e.runCommand(context.Background(), tc.args); exit != tc.exit {
				t.Fatalf("exit=%d output=%s", exit, stdout.String())
			}
			if strings.Contains(stdout.String()+stderr.String(), "SECRET") || strings.Count(stdout.String(), "\n") != 1 || !strings.HasSuffix(stdout.String(), "\n") {
				t.Fatal("unsafe or non-single result", stdout.String(), stderr.String())
			}
			var result map[string]any
			if err := json.Unmarshal(stdout.Bytes(), &result); err != nil || len(result) != 9 || result["operation_id"] != id.String() {
				t.Fatal("not closed result", result, err)
			}
			for _, field := range []string{"container_sha256", "pack_key", "pack_version", "job_id", "error_code", "reason_code"} {
				if _, ok := result[field]; !ok {
					t.Fatal("nullable member omitted", field)
				}
			}
			if !strings.HasPrefix(stdout.String(), `{"container_sha256":null,"error_code":`) {
				t.Fatal("noncanonical member ordering")
			}
			if tc.phase == "syntax" {
				if loads != 0 || opens != 0 || imports != 0 || closes != 0 || stderr.String() != "usage: "+referencePackImportUsage+"\n" {
					t.Fatal("invalid syntax caused effects")
				}
			} else if stderr.Len() != 0 {
				t.Fatal("ordinary result leaked stderr")
			}
			if tc.phase == "success" || tc.phase == "rejected" {
				if loads != 1 || opens != 1 || imports != 1 || closes != 1 {
					t.Fatal("operation or ownership repeated")
				}
			}
		})
	}
}

type referencePackAdmissionFunc func(context.Context, io.Reader) (*reference_data.PendingImport, error)

func (f referencePackAdmissionFunc) PrepareImport(ctx context.Context, reader io.Reader) (*reference_data.PendingImport, error) {
	return f(ctx, reader)
}

func testOperatorStagingRejection(t *testing.T) {
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, "incoming"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "incoming/pack.zip"), []byte("input"), 0600); err != nil {
		t.Fatal(err)
	}
	storage, err := referenceassembly.NewRootStorage(t.TempDir(), root)
	if err != nil {
		t.Fatal(err)
	}
	defer storage.Close()
	bound := &reference_data.ContentRejection{Code: "container_bytes_exceeded", CheckID: "container_bytes"}
	for _, tc := range []struct {
		err  error
		code string
	}{{bound, "reference_pack_verification_failed"}, {errors.New("storage unavailable"), "internal_error"}, {errors.Join(bound, errors.New("cleanup unavailable")), "internal_error"}} {
		runtime := referencePackLocalRuntime{ctx: context.Background(), storage: storage, admission: referencePackAdmissionFunc(func(context.Context, io.Reader) (*reference_data.PendingImport, error) { return nil, tc.err })}
		result := runtime.importAndObserve(context.Background(), uuid.New(), "pack.zip")
		if result.ErrorCode == nil || *result.ErrorCode != tc.code || result.JobID != nil || result.ContainerSHA256 != nil || result.PackKey != nil {
			t.Fatalf("admission result: %#v", result)
		}
		if tc.code == "reference_pack_verification_failed" && (result.ReasonCode == nil || *result.ReasonCode != "container_bytes_exceeded") {
			t.Fatal(result)
		}
	}
}
