package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func TestRecoveryProjectionClosure(t *testing.T) {
	source := filepath.Join("..", "..", "contracts", "recovery")
	copyInputs := func(t *testing.T) string {
		t.Helper()
		base := t.TempDir()
		err := filepath.WalkDir(source, func(path string, entry os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if entry.IsDir() || !strings.HasSuffix(path, ".json") {
				return nil
			}
			relative, err := filepath.Rel(source, path)
			if err != nil {
				return err
			}
			target := filepath.Join(base, relative)
			if err := os.MkdirAll(filepath.Dir(target), 0o700); err != nil {
				return err
			}
			body, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			return os.WriteFile(target, body, 0o600)
		})
		if err != nil {
			t.Fatal(err)
		}
		return base
	}
	writeObject := func(t *testing.T, path string, value any) {
		t.Helper()
		body, err := json.Marshal(value)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, body, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	if err := validateRecoveryContractFamily(filepath.Join("..", "..")); err != nil {
		t.Fatal(err)
	}
	t.Run("coherently rebound unsupported Graph projection", func(t *testing.T) {
		base := copyInputs(t)
		path := "fixtures/graph-projection-restore-implementation-binding.v8.json"
		binding, oldSHA, err := readRecoveryCanonicalArtifact(base, path)
		if err != nil {
			t.Fatal(err)
		}
		binding["projection_schema_id"] = "graph_projection.v999"
		writeObject(t, filepath.Join(base, path), binding)
		_, newSHA, err := readRecoveryCanonicalArtifact(base, path)
		if err != nil {
			t.Fatal(err)
		}
		err = filepath.WalkDir(base, func(path string, entry os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if entry.IsDir() || !strings.HasSuffix(path, ".json") {
				return nil
			}
			body, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			body = []byte(strings.ReplaceAll(string(body), oldSHA, newSHA))
			if filepath.Base(path) == "graph-projection-restore-implementation-binding.v8.schema.json" {
				body = []byte(strings.ReplaceAll(string(body), "graph_projection.v2", "graph_projection.v999"))
			}
			return os.WriteFile(path, body, 0o600)
		})
		if err != nil {
			t.Fatal(err)
		}
		if err := validateGraphProjectionRestoreV5Contracts(base); err == nil || !strings.Contains(err.Error(), "drifted from its authored source set") {
			t.Fatalf("unsupported typed projection must reject before downstream artifact checks: %v", err)
		}
	})
	for _, name := range []string{"missing role", "extra role", "wrong role", "retired", "unresolved", "unpackaged"} {
		t.Run(name, func(t *testing.T) {
			base := copyInputs(t)
			registry, err := readRecoveryObject(base, "index.json")
			if err != nil {
				t.Fatal(err)
			}
			roles := registry["codec_schemas"].(map[string]any)
			switch name {
			case "missing role":
				delete(roles, "postgres_snapshot_unit")
			case "extra role":
				roles["unused"] = roles["postgres_snapshot_unit"]
			case "wrong role":
				roles["postgres_snapshot_unit"] = roles["postgres_snapshot_artifact"]
			case "retired":
				roles["backup_artifact_envelope"] = "cartulary.backup_artifact_envelope.v2"
			case "unresolved":
				roles["postgres_snapshot_unit"] = "cartulary.postgres_snapshot_unit.v999"
				registry["current_schema_ids"] = append(registry["current_schema_ids"].([]any), "cartulary.postgres_snapshot_unit.v999")
			case "unpackaged":
				paths := registry["schemas"].([]any)
				for i, path := range paths {
					if path == "postgres-snapshot-unit.v1.schema.json" {
						registry["schemas"] = append(paths[:i], paths[i+1:]...)
						break
					}
				}
			}
			writeObject(t, filepath.Join(base, "index.json"), registry)
			if _, err := resolveRecoveryCodecSchemas(base); err == nil {
				t.Fatal("invalid codec role admitted")
			}
		})
	}
	required, err := resolveRecoveryCodecSchemas(source)
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"retired", "omitted", "unresolved", "extra current"} {
		t.Run("binding "+name, func(t *testing.T) {
			ids := slices.Clone(required)
			switch name {
			case "retired":
				ids[0] = "cartulary.backup_artifact_envelope.v2"
			case "omitted":
				ids = ids[1:]
			case "unresolved":
				ids[0] = "cartulary.backup_artifact_envelope.v999"
			case "extra current":
				ids = append(ids, "cartulary.operator_recovery_progress.v2")
			}
			slices.Sort(ids)
			values := make([]any, len(ids))
			for i, id := range ids {
				values[i] = id
			}
			sum := sha256.Sum256([]byte("CARTULARY-RECOVERY-CODEC-REGISTRY-VNEXT\n" + strings.Join(ids, "\n") + "\n"))
			binding := map[string]any{"schema_ids": values, "digest_sha256": hex.EncodeToString(sum[:])}
			if _, err := validateRecoveryGenerationCodecBinding(binding, required, "test binding"); err == nil {
				t.Fatal("self-consistent incompatible codec set admitted")
			}
		})
	}
	t.Run("bounded shape and exact membership", func(t *testing.T) {
		base := copyInputs(t)
		registry, err := readRecoveryObject(base, "index.json")
		if err != nil {
			t.Fatal(err)
		}
		schema, err := readRecoveryObject(base, "postgres-snapshot-artifact.v3.schema.json")
		if err != nil {
			t.Fatal(err)
		}
		units := schema["properties"].(map[string]any)["units"].(map[string]any)
		units["maxItems"] = 84
		writeObject(t, filepath.Join(base, "postgres-snapshot-artifact.v3.schema.json"), schema)
		if err := validateRecoverySnapshotSchemaBounds(base, registry); err == nil {
			t.Fatal("stale structural bound admitted")
		}
		snapshot := map[string]any{"units": []any{map[string]any{"table_name": "a"}, map[string]any{"table_name": "b"}}}
		if err := validateRecoverySnapshotFixture(snapshot, []string{"a", "b"}); err != nil {
			t.Fatal(err)
		}
		for _, required := range [][]string{{"a"}, {"a", "b", "c"}, {"b", "a"}} {
			if err := validateRecoverySnapshotFixture(snapshot, required); err == nil {
				t.Fatal("inexact table membership admitted")
			}
		}
	})
}
