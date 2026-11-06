package restore_test

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/google/uuid"

	contractrecovery "github.com/JochiRaider/cartulary/internal/gen/contractrecovery"
	. "github.com/JochiRaider/cartulary/internal/modules/graphprojection/restore"
)

func TestGraphRestoreCurrentRegistryAndBindingsAreExact(t *testing.T) {
	t.Run("registry owns admitted metadata", checkRestoreRegistryOwnership)
	t.Run("definition bytes reject duplicate and aliased members", func(t *testing.T) {
		body := contractrecovery.CurrentGraphProjectionRestoreImplementationBindingJSON
		for name, malformed := range map[string]string{
			"duplicate":  `{"schema_id":"retired",` + body[1:],
			"case alias": strings.Replace(body, `"schema_id"`, `"SCHEMA_ID"`, 1),
			"unknown":    `{"unknown":true,` + body[1:],
			"trailing":   body + `{}`,
			"unicode":    strings.Replace(body, `"schema_id":`, `"\ud800":`, 1),
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
	for _, field := range []string{"schema_id", "algorithm_id", "projection_schema_id", "binding_id", "source_registry_sha256", "recovery_state_catalog_sha256", "semantic_query_schema_ids", "graph_table_ids"} {
		t.Run("unsupported "+field, func(t *testing.T) {
			var object map[string]any
			if err := json.Unmarshal([]byte(contractrecovery.CurrentGraphProjectionRestoreImplementationBindingJSON), &object); err != nil {
				t.Fatal(err)
			}
			switch object[field].(type) {
			case []any:
				object[field] = []string{"unsupported.v999"}
			default:
				object[field] = "unsupported.v999"
			}
			body, err := json.Marshal(object)
			if err != nil {
				t.Fatal(err)
			}
			sum := sha256.Sum256(body)
			for _, digest := range []string{current.SHA256, fmt.Sprintf("%x", sum[:])} {
				if ref := FrozenRestoreImplementationBinding(body, digest); ref.SHA256 != "" {
					t.Fatal("altered executable definition admitted with an original or recomputed digest")
				}
			}
		})
	}
	if len(contractrecovery.RecoveryGenerations) != 1 ||
		contractrecovery.RecoveryGenerations[0].GenerationID == "" ||
		!contractrecovery.RecoveryGenerations[0].CaptureCurrent {
		t.Fatalf("recovery must project exactly one current Graph v5 generation: %#v", contractrecovery.RecoveryGenerations)
	}
	currentPaths := map[string]bool{
		"contracts/recovery/graph-projection-restore-implementation-binding.v9.schema.json":   true,
		"contracts/recovery/graph-projection-restore-source-registry.v6.schema.json":          true,
		"contracts/recovery/graph-projection-restore-rebuild-result.v6.schema.json":           true,
		"contracts/recovery/fixtures/graph-projection-restore-implementation-binding.v9.json": true,
		"contracts/recovery/fixtures/graph-projection-restore-source-registry.v6.json":        true,
		"contracts/recovery/fixtures/graph-projection-restore-rebuild-result.v6.json":         true,
	}
	for path := range contractrecovery.Index {
		if strings.Contains(path, "graph-projection-restore-") && !currentPaths[path] {
			t.Fatalf("unselected Graph restore artifact remains packaged: %s", path)
		}
	}
}

func checkRestoreRegistryOwnership(t *testing.T) {
	for _, boundary := range []string{"constructor", "document", "registrations"} {
		t.Run(boundary, func(t *testing.T) {
			_, registration, expected := currentRestoreFixture(t)
			calls := 0
			enumerate := registration.Enumerate
			registration.Enumerate = func(ctx context.Context) ([]RestoreCandidate, error) {
				calls++
				return enumerate(ctx)
			}
			inputs := []RestoreSourceRegistration{registration}
			registry, err := NewCurrentRestoreSourceRegistry(inputs...)
			if err != nil {
				t.Fatal(err)
			}
			publisher := &recordingRestorePublisher{}
			service, err := NewRestoreService(publisher, registry)
			if err != nil {
				t.Fatal(err)
			}
			want := CurrentRestoreSourceRegistry().Document()
			switch boundary {
			case "constructor":
				inputs[0].Entry.SemanticQuerySchemaIDs[0] = "mutated"
				inputs[0].Entry.SourceOwnerID = "mutated"
				inputs[0].Enumerate = nil
			case "document":
				document := registry.Document()
				document.Entries[0].SemanticQuerySchemaIDs[0] = "mutated"
				document.Entries[0].SourceOwnerID = "mutated"
			case "registrations":
				registrations := registry.Registrations()
				registrations[0].Entry.SemanticQuerySchemaIDs[0] = "mutated"
				registrations[0].Entry.SourceOwnerID = "mutated"
				registrations[0].Enumerate = nil
			}
			if !reflect.DeepEqual(registry.Document(), want) || !reflect.DeepEqual(registry.Registrations()[0].Entry, want.Entries[0]) || registry.DigestSHA256() != contractrecovery.CurrentGraphProjectionRestoreSourceRegistrySHA256 {
				t.Fatal("external mutation changed admitted registry contents or digest")
			}
			request := currentRestoreRequest(context.Background(), registry, CurrentRestoreImplementationBinding())
			result, err := service.Rebuild(request.Context, request)
			if err != nil || !result.ReadinessSatisfied() || calls != 1 || len(publisher.plans) != 1 || publisher.plans[0].Projections[0].Result.Binding != expected {
				t.Fatalf("external mutation changed reconstruction or callback: calls=%d result=%#v err=%v", calls, result, err)
			}
		})
	}
	packaged := CurrentRestoreSourceRegistry()
	document := packaged.Document()
	document.Entries[0].SemanticQuerySchemaIDs[0] = "mutated"
	if !reflect.DeepEqual(packaged.Document(), CurrentRestoreSourceRegistry().Document()) {
		t.Fatal("packaged registry accessor leaked mutable metadata")
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
