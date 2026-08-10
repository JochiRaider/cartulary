package revisions

import (
	"errors"
	"os"
	"strings"
	"testing"
)

func TestProductionReadinessComponentsRemainDecomposed_Unit(t *testing.T) {
	required := []string{
		"appender_change_set_storage.go",
		"appender_facade.go",
		"appender_intent_publication.go",
		"appender_mutation_storage.go",
		"appender_revision_storage.go",
		"target_history_facets.go",
		"target_semantics_catalog.go",
		"target_semantics_compiler.go",
		"target_semantics_lookup.go",
		"target_semantics_registry.go",
		"incident_bundle_apply.go",
		"incident_bundle_export.go",
		"incident_bundle_models.go",
		"incident_bundle_parse_prepare.go",
		"incident_bundle_sequence.go",
		"incident_bundle_validation.go",
		"rollback_query_affected.go",
		"rollback_query_companions.go",
		"rollback_query_currentness.go",
		"rollback_query_facade.go",
		"rollback_query_selectors.go",
		"rollback_apply_coordinator.go",
		"rollback_apply_effects.go",
		"rollback_apply_nonrow.go",
		"rollback_apply_row.go",
		"catalog_admission_test.go",
	}
	for _, path := range required {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("required production-readiness component %s: %v", path, err)
		}
	}

	retired := []string{
		"appender.go",
		"target_semantics.go",
		"incident_bundle_portability.go",
		"rollback_query_repository.go",
		"rollback_apply.go",
		"candidate_semantics_test.go",
	}
	for _, path := range retired {
		if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
			t.Fatalf("retired production-readiness component %s must remain absent, stat error = %v", path, err)
		}
	}
}

func TestProductionReadinessRetiredIdentifiersRemainAbsent_Unit(t *testing.T) {
	prohibited := []string{
		"BuildPatchConflictWindowWith" + "Descriptors",
		"DecodeRevision" + "Row",
		"NewFieldDescriptor" + "Set",
		"NewRecordSnapshotCaptureCatalogFor" + "Requirements",
		"DeclaredRecord" + "Types",
		"type TargetSemantics" + "Descriptor",
		"type RecordView" + "Descriptor",
		"Projection" + "Services",
		"LiveRecordChange" + "Policy",
		"LiveRecordChange" + "Required",
		"LiveRecordChange" + "None",
		"HistoryTarget" + "Semantics",
		"Rollback" + "Dispatch",
		"Captured" + "Record",
		"Append" + "Captured",
		"ErrInvalid" + "CapturedSnapshot",
		"candidate" + "Semantics",
	}
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("read Revisions package: %v", err)
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") {
			continue
		}
		contents, err := os.ReadFile(entry.Name())
		if err != nil {
			t.Fatalf("read Revisions source %s: %v", entry.Name(), err)
		}
		for _, identifier := range prohibited {
			if strings.Contains(string(contents), identifier) {
				t.Fatalf("Revisions source %s contains retired identifier %q", entry.Name(), identifier)
			}
		}
	}
}
