package graphprojection

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/google/uuid"

	contractrecovery "github.com/JochiRaider/cartulary/internal/gen/contractrecovery"
)

func TestGraphRestoreCurrentRegistryAndBindingsAreExact(t *testing.T) {
	registry := CurrentRestoreSourceRegistry()
	if registry.DigestSHA256() != contractrecovery.CurrentGraphProjectionRestoreSourceRegistrySHA256 ||
		registry.DigestSHA256() != CurrentRestoreSourceRegistry().DigestSHA256() ||
		len(registry.Document().Entries) != 1 {
		t.Fatalf("current v2 restore source registry drifted: %#v", registry.Document())
	}
	current := CurrentRestoreImplementationBinding()
	if current.Binding.SchemaID != RestoreImplementationBindingSchemaID ||
		current.Binding.AlgorithmID != RestoreAlgorithmID ||
		current.Binding.SourceRegistrySHA256 != registry.DigestSHA256() ||
		!equalStrings(current.Binding.GraphTableIDs, RestoreGraphTableIDs()) ||
		current.SHA256 != contractrecovery.CurrentGraphProjectionRestoreImplementationBindingSHA256 {
		t.Fatalf("current restore implementation binding drifted: %#v", current)
	}
	var canonical map[string]any
	if err := json.Unmarshal([]byte(contractrecovery.CurrentGraphProjectionRestoreImplementationBindingJSON), &canonical); err != nil || len(canonical) != 13 {
		t.Fatalf("generated implementation binding is not a closed canonical object: fields=%d err=%v", len(canonical), err)
	}
	for path := range contractrecovery.Index {
		if strings.Contains(path, "graph-projection-restore-") && strings.Contains(path, ".v1") {
			t.Fatalf("retired Graph restore v1 artifact remains packaged: %s", path)
		}
	}
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
		ReconciledNonterminalJobCount: 0, ReconciledLeaseCount: 0,
		PostconditionSHA256: &postcondition, Warnings: []RestoreSafeMessage{}, Errors: []RestoreSafeMessage{},
	}
	if err := result.Validate(); err != nil || !result.ReadinessSatisfied() {
		t.Fatalf("valid successful restore result rejected: %v", err)
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
