package restore_test

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/google/uuid"

	contractrecovery "github.com/JochiRaider/cartulary/internal/gen/contractrecovery"
	. "github.com/JochiRaider/cartulary/internal/modules/graphprojection/restore"
)

func TestGraphRestoreCurrentRegistryAndBindingsAreExact(t *testing.T) {
	t.Run("definition bytes reject duplicate and aliased members", func(t *testing.T) {
		body := contractrecovery.CurrentGraphProjectionRestoreImplementationBindingJSON
		for name, malformed := range map[string]string{
			"duplicate":  `{"schema_id":"retired",` + body[1:],
			"case alias": strings.Replace(body, `"schema_id"`, `"SCHEMA_ID"`, 1),
		} {
			t.Run(name, func(t *testing.T) {
				if ref := FrozenRestoreImplementationBinding([]byte(malformed), contractrecovery.CurrentGraphProjectionRestoreImplementationBindingSHA256); ref.SHA256 != "" {
					t.Fatal("malformed definition bytes admitted as the current frozen binding")
				}
			})
		}
	})
	registry := CurrentRestoreSourceRegistry()
	if registry.DigestSHA256() != contractrecovery.CurrentGraphProjectionRestoreSourceRegistrySHA256 ||
		registry.DigestSHA256() != CurrentRestoreSourceRegistry().DigestSHA256() ||
		len(registry.Document().Entries) != 1 {
		t.Fatalf("current v5 restore source registry drifted: %#v", registry.Document())
	}
	current := CurrentRestoreImplementationBinding()
	if current.Binding.SchemaID != RestoreImplementationBindingSchemaID ||
		current.Binding.AlgorithmID != RestoreAlgorithmID ||
		current.Binding.ProjectionSchemaID != "graph_projection.v2" ||
		current.Binding.SourceRegistrySHA256 != registry.DigestSHA256() ||
		!equalStrings(current.Binding.GraphTableIDs, RestoreGraphTableIDs()) ||
		current.SHA256 != contractrecovery.CurrentGraphProjectionRestoreImplementationBindingSHA256 {
		t.Fatalf("current restore implementation binding drifted: %#v", current)
	}
	var canonical map[string]any
	if err := json.Unmarshal([]byte(contractrecovery.CurrentGraphProjectionRestoreImplementationBindingJSON), &canonical); err != nil || len(canonical) != 8 {
		t.Fatalf("generated implementation binding is not a closed canonical object: fields=%d err=%v", len(canonical), err)
	}
	for _, field := range []string{"schema_id", "algorithm_id", "projection_schema_id"} {
		t.Run("unsupported "+field, func(t *testing.T) {
			var object map[string]any
			if err := json.Unmarshal([]byte(contractrecovery.CurrentGraphProjectionRestoreImplementationBindingJSON), &object); err != nil {
				t.Fatal(err)
			}
			object[field] = "unsupported.v999"
			body, err := json.Marshal(object)
			if err != nil {
				t.Fatal(err)
			}
			sum := sha256.Sum256(body)
			if ref := FrozenRestoreImplementationBinding(body, fmt.Sprintf("%x", sum[:])); ref.SHA256 != "" {
				t.Fatal("unsupported executable definition admitted with a matching digest")
			}
		})
	}
	if len(contractrecovery.RecoveryGenerations) != 1 ||
		contractrecovery.RecoveryGenerations[0].GenerationID == "" ||
		!contractrecovery.RecoveryGenerations[0].CaptureCurrent {
		t.Fatalf("recovery must project exactly one current Graph v5 generation: %#v", contractrecovery.RecoveryGenerations)
	}
	currentPaths := map[string]bool{
		"contracts/recovery/graph-projection-restore-implementation-binding.v7.schema.json":   true,
		"contracts/recovery/graph-projection-restore-source-registry.v5.schema.json":          true,
		"contracts/recovery/graph-projection-restore-rebuild-result.v5.schema.json":           true,
		"contracts/recovery/fixtures/graph-projection-restore-implementation-binding.v7.json": true,
		"contracts/recovery/fixtures/graph-projection-restore-source-registry.v5.json":        true,
		"contracts/recovery/fixtures/graph-projection-restore-rebuild-result.v5.json":         true,
	}
	for path := range contractrecovery.Index {
		if strings.Contains(path, "graph-projection-restore-") && !currentPaths[path] {
			t.Fatalf("unselected Graph restore artifact remains packaged: %s", path)
		}
	}
}

func equalStrings(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}

func TestGraphRestoreResultContractClosesStatesAndSafeErrors(t *testing.T) {
	postcondition := strings.Repeat("a", 64)
	result := RestoreRebuildResult{
		SchemaID: RestoreRebuildResultSchemaID, RestoreOperationID: uuid.NewString(),
		TargetGenerationID: uuid.NewString(), Status: RestoreStatusSucceeded,
		ReadinessOutcome: RestoreReadinessReady, AlgorithmID: RestoreAlgorithmID,
		ImplementationBindingSHA256:   strings.Repeat("b", 64),
		SourceRegistrySHA256:          strings.Repeat("c", 64),
		ClearedTableIDs:               RestoreGraphTableIDs(),
		RebuiltViews:                  []RestoreRebuiltView{},
		ReconciledNonterminalJobCount: 0,
		PostconditionSHA256:           &postcondition, Warnings: []RestoreSafeMessage{}, Errors: []RestoreSafeMessage{},
	}
	if err := result.Validate(); err != nil || !result.ReadinessSatisfied() {
		t.Fatalf("valid successful restore result rejected: %v", err)
	}
	retired := result
	retired.SchemaID = "cartulary.graph_projection_restore_rebuild_result.v4"
	if retired.Validate() == nil {
		t.Fatal("retired result schema admitted")
	}
	retired = result
	retired.AlgorithmID = "graphprojection.restore_rebuild.v4"
	if retired.Validate() == nil {
		t.Fatal("retired rebuild algorithm admitted")
	}
	retired = result
	retired.ClearedTableIDs = append(RestoreGraphTableIDs(), "graph_projection_result_leases")
	if retired.Validate() == nil {
		t.Fatal("retired table admitted")
	}
	invalid := result
	invalid.ReadinessOutcome = RestoreReadinessIncomplete
	if invalid.Validate() == nil {
		t.Fatal("invalid success/incomplete result combination was admitted")
	}
	failed := result
	failed.Status = RestoreStatusFailed
	failed.ReadinessOutcome = RestoreReadinessIncomplete
	failed.ClearedTableIDs = []string{}
	failed.PostconditionSHA256 = nil
	failed.Errors = []RestoreSafeMessage{{Code: RestoreErrorPublicationFailed}}
	if err := failed.Validate(); err != nil {
		t.Fatalf("valid failed/incomplete result rejected: %v", err)
	}
	secret := "postgres://operator:SECRET_PASSWORD@database/restore"
	err := NewRestoreError(RestoreErrorPublicationFailed)
	if strings.Contains(err.Error(), secret) || err.Error() != "graphprojection restore failed: restore_publication_failed" {
		t.Fatalf("restore error is not closed and safe: %q", err)
	}
}
