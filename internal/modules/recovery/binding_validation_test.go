package recovery

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
	"github.com/JochiRaider/cartulary/internal/platform/extensioncodec"
	"github.com/JochiRaider/cartulary/internal/testutil/recoverytest"
)

func TestPostgresBindingProofEnforcement_Unit(t *testing.T) {
	binding := logicalTestBinding(t, "binding", "family", "physical_rows", 0)
	scratch := recoverytest.ProofScratch(t)
	entries := func(rows ...[2]string) extensioncodec.Entries {
		return func(yield func(extensioncodec.Entry) error) error {
			for _, row := range rows {
				entry, err := binding.CurrentCodec.Postgres.EncodeRow([]any{row[0], row[1]})
				if err != nil {
					return err
				}
				if err := yield(entry); err != nil {
					return err
				}
			}
			return nil
		}
	}
	source := entries([2]string{"b", "second"}, [2]string{"a", "first"})
	proof, err := extensioncodec.ComputeProof(t.Context(), binding.BindingID, "postgres", binding.CurrentCodec.Postgres.Binding().Limits, scratch, source)
	if err != nil {
		t.Fatal(err)
	}
	expected := currentTestProof(binding)
	expected.ItemCount, expected.ContentBytes, expected.BindingDigest = proof.ItemCount, proof.ContentBytes, proof.BindingDigest
	check := func(t *testing.T, expected PostgresBindingProof, read extensioncodec.Entries, valid bool) {
		t.Helper()
		result := validatePostgresBinding(t.Context(), binding, expected, read, scratch)
		err := classifyBindingRestoreResult(binding.BindingID, result, nil)
		if (err == nil) != valid || (result.Status == "valid") != valid {
			t.Fatalf("result=%#v err=%v want valid=%v", result, err, valid)
		}
		body, marshalErr := canonicaljson.MarshalExtension(result)
		if marshalErr != nil || bytes.Contains(body, []byte("private-storage-secret")) {
			t.Fatal("unsafe or invalid result", marshalErr)
		}
	}
	t.Run("logical ordering", func(t *testing.T) {
		check(t, expected, entries([2]string{"a", "first"}, [2]string{"b", "second"}), true)
	})
	for name, read := range map[string]extensioncodec.Entries{
		"missing":                    entries([2]string{"a", "first"}),
		"extra":                      entries([2]string{"a", "first"}, [2]string{"b", "second"}, [2]string{"c", "third"}),
		"same size changed identity": entries([2]string{"c", "first"}, [2]string{"b", "second"}),
		"same size changed bytes":    entries([2]string{"a", "First"}, [2]string{"b", "second"}),
		"duplicate identity":         entries([2]string{"a", "first"}, [2]string{"a", "first"}),
		"entry byte limit":           entries([2]string{"a", strings.Repeat("x", 129)}),
		"storage failure":            func(func(extensioncodec.Entry) error) error { return errors.New("private-storage-secret") },
	} {
		t.Run(name, func(t *testing.T) { check(t, expected, read, false) })
	}
	for _, field := range []string{"profile", "binding", "physical", "implementation", "codec", "codec digest", "storage", "count", "bytes", "digest"} {
		t.Run(field, func(t *testing.T) {
			changed := expected
			switch field {
			case "profile":
				changed.ProfileID = "another_profile"
			case "binding":
				changed.BindingID = "test_profile.other"
			case "physical":
				changed.PhysicalStateBindingSHA256 = strings.Repeat("a", 64)
			case "implementation":
				changed.ImplementationBindingSHA256 = strings.Repeat("a", 64)
			case "codec":
				changed.BackupCodecID = "retired.codec"
			case "codec digest":
				changed.BackupCodecSHA256 = strings.Repeat("a", 64)
			case "storage":
				changed.StorageKind = "filesystem"
			case "count":
				changed.ItemCount++
			case "bytes":
				changed.ContentBytes++
			case "digest":
				changed.BindingDigest = strings.Repeat("a", 64)
			}
			check(t, changed, source, false)
		})
	}
	t.Run("empty bindings have an explicit proof", func(t *testing.T) {
		proof, err := extensioncodec.ComputeProof(t.Context(), binding.BindingID, "postgres", binding.CurrentCodec.Postgres.Binding().Limits, scratch, entries())
		if err != nil {
			t.Fatal(err)
		}
		empty := expected
		empty.ItemCount, empty.ContentBytes, empty.BindingDigest = 0, 0, proof.BindingDigest
		check(t, empty, entries(), true)
		check(t, expected, entries(), false)
	})
	t.Run("result precedence", func(t *testing.T) {
		for _, count := range []int{0, 256, 257, 4096, 4097, 5000} {
			t.Run(fmt.Sprint(count), func(t *testing.T) {
				result := bindingRestoreResult{SchemaID: "wrong", BindingID: "wrong", Status: "wrong", Findings: make([]bindingRestoreFinding, count)}
				err := classifyBindingRestoreResult(binding.BindingID, result, nil)
				want := "extension_validation_result_invalid"
				if count > 4096 {
					want = "extension_diagnostic_overflow limit=4096 actual=4097"
				}
				if err == nil || !strings.Contains(err.Error(), want) {
					t.Fatalf("count=%d classification=%v", count, err)
				}
				if err := classifyBindingRestoreResult(binding.BindingID, result, context.Canceled); err != context.Canceled {
					t.Fatalf("invocation did not take precedence: %v", err)
				}
			})
		}
		result := bindingRestoreResult{SchemaID: bindingRestoreResultSchemaID, BindingID: binding.BindingID, Status: "valid"}
		if err := classifyBindingRestoreResult(binding.BindingID, result, nil); err == nil {
			t.Fatal("null findings admitted")
		}
	})
	t.Run("strict snapshot byte boundary", func(t *testing.T) {
		snapshot := VNextPostgresSnapshotArtifact{SchemaID: PostgresSnapshotArtifactV3SchemaID, BackupSetID: "00000000-0000-0000-0000-000000000001", ConsistencyPointAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), RecoveryStateCatalogSHA256: strings.Repeat("1", 64), TransactionIsolation: VNextTransactionIsolation, Units: []VNextPostgresUnitDescriptor{}, ExtensionBindingProofs: []PostgresBindingProof{expected}}
		snapshot.SnapshotSHA256, err = postgresSnapshotDigest(snapshot)
		if err != nil {
			t.Fatal(err)
		}
		body, err := canonicaljson.MarshalExtension(snapshot)
		if err != nil {
			t.Fatal(err)
		}
		var decoded VNextPostgresSnapshotArtifact
		if err := decodePostgresSnapshot(body, &decoded); err != nil {
			t.Fatal(err)
		}
		for _, malformed := range [][]byte{[]byte{0xff}, []byte("[]\n"), body[:len(body)-1], append([]byte(" "), body...), bytes.Replace(body, []byte(`"schema_id":`), []byte(`"extra":0,"schema_id":`), 1), bytes.Replace(body, []byte(`"binding_id":`), []byte(`"binding_id":"duplicate","binding_id":`), 1)} {
			if err := decodePostgresSnapshot(malformed, &decoded); err == nil {
				t.Fatal("malformed snapshot accepted")
			}
		}
	})
}
