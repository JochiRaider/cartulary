package sourcecatalog

import (
	"testing"

	"github.com/JochiRaider/cartulary/internal/gen/contracttasksdecisions"
)

func TestCatalogRejectsGeneratedOwnerFactDrift(t *testing.T) {
	t.Parallel()
	catalog, err := build(cloneGeneratedCatalog())
	if err != nil || len(catalog.Surfaces()) != 2 || len(catalog.Fields()) != 23 {
		t.Fatalf("canonical catalog = %d surfaces, %d fields, err=%v", len(catalog.Surfaces()), len(catalog.Fields()), err)
	}
	tests := []struct {
		name   string
		mutate func([]contracttasksdecisions.SourceSurface) []contracttasksdecisions.SourceSurface
	}{
		{name: "missing surface", mutate: func(catalog []contracttasksdecisions.SourceSurface) []contracttasksdecisions.SourceSurface {
			return catalog[:len(catalog)-1]
		}},
		{name: "duplicate surface", mutate: func(catalog []contracttasksdecisions.SourceSurface) []contracttasksdecisions.SourceSurface {
			catalog[1] = catalog[0]
			return catalog
		}},
		{name: "cross surface field", mutate: func(catalog []contracttasksdecisions.SourceSurface) []contracttasksdecisions.SourceSurface {
			field := catalog[0].DirectFields[len(catalog[0].DirectFields)-1]
			catalog[0].DirectFields = catalog[0].DirectFields[:len(catalog[0].DirectFields)-1]
			catalog[1].DirectFields = append(catalog[1].DirectFields, field)
			return catalog
		}},
		{name: "read only field", mutate: func(catalog []contracttasksdecisions.SourceSurface) []contracttasksdecisions.SourceSurface {
			catalog[0].DirectFields[0].FieldKey = "decision.record_id"
			return catalog
		}},
		{name: "duplicate field", mutate: func(catalog []contracttasksdecisions.SourceSurface) []contracttasksdecisions.SourceSurface {
			catalog[0].DirectFields[1] = catalog[0].DirectFields[0]
			return catalog
		}},
		{name: "stale view facts", mutate: func(catalog []contracttasksdecisions.SourceSurface) []contracttasksdecisions.SourceSurface {
			catalog[0].DirectFields[0].View.Clearable = !catalog[0].DirectFields[0].View.Clearable
			return catalog
		}},
		{name: "unowned storage relation", mutate: func(catalog []contracttasksdecisions.SourceSurface) []contracttasksdecisions.SourceSurface {
			catalog[0].SourceTable = "records"
			return catalog
		}},
		{name: "unsafe storage identifier", mutate: func(catalog []contracttasksdecisions.SourceSurface) []contracttasksdecisions.SourceSurface {
			catalog[0].DirectFields[0].Column = "title;drop_table"
			return catalog
		}},
		{name: "reference mismatch", mutate: func(catalog []contracttasksdecisions.SourceSurface) []contracttasksdecisions.SourceSurface {
			catalog[1].DirectFields[3].ExpectedTargetRecordType = "party"
			return catalog
		}},
		{name: "unknown collection operation", mutate: func(catalog []contracttasksdecisions.SourceSurface) []contracttasksdecisions.SourceSurface {
			catalog[0].CollectionFields[0].AllowedOperations[0] = "future_operation"
			return catalog
		}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := build(tc.mutate(cloneGeneratedCatalog())); err == nil {
				t.Fatalf("runtime catalog accepted %s drift", tc.name)
			}
		})
	}
}

func TestCatalogLookupsAreIndexedAndDefensive(t *testing.T) {
	t.Parallel()
	catalog, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	loadedAgain, err := Load()
	if err != nil || loadedAgain != catalog {
		t.Fatalf("Load() did not return cached catalog: same=%v err=%v", loadedAgain == catalog, err)
	}
	if len(catalog.Surfaces()) != 2 || len(catalog.Fields()) != 23 || len(catalog.WritableDirectStorageMappings()) != 20 {
		t.Fatalf("catalog counts = %d/%d/%d", len(catalog.Surfaces()), len(catalog.Fields()), len(catalog.WritableDirectStorageMappings()))
	}
	surfaces := catalog.Surfaces()
	surfaces[0].RecordType = "mutated"
	decision, ok := catalog.SurfaceByViewID(contracttasksdecisions.DecisionViewSchemaID)
	if !ok || decision.RecordType != "decision" {
		t.Fatalf("surface mutation escaped defensive copy: %#v, ok=%v", decision, ok)
	}
	byRecord, ok := catalog.SurfaceByRecordType(decision.RecordType)
	if !ok || byRecord != decision {
		t.Fatalf("record index = %#v, %v; want %#v", byRecord, ok, decision)
	}
	field, ok := catalog.Field("decision.affected_record_ids")
	if !ok || len(field.Collection.AllowedOperations) != 2 {
		t.Fatalf("collection field missing: %#v", field)
	}
	field.Collection.AllowedOperations[0] = "mutated"
	loadedField, _ := catalog.Field(field.FieldKey)
	if loadedField.Collection.AllowedOperations[0] == "mutated" {
		t.Fatal("collection mutation escaped defensive copy")
	}
	keys := catalog.ConflictFieldSourceKeys(contracttasksdecisions.TaskRequestViewSchemaID)
	if len(keys) != 14 || keys["task.decision_record_id"] != "decision_record_id" {
		t.Fatalf("task conflict source keys = %#v", keys)
	}
}

func cloneGeneratedCatalog() []contracttasksdecisions.SourceSurface {
	result := make([]contracttasksdecisions.SourceSurface, len(contracttasksdecisions.SourceCatalog))
	for index, surface := range contracttasksdecisions.SourceCatalog {
		result[index] = surface
		result[index].DirectFields = append([]contracttasksdecisions.DirectField(nil), surface.DirectFields...)
		result[index].CollectionFields = append([]contracttasksdecisions.CollectionField(nil), surface.CollectionFields...)
		for fieldIndex := range result[index].DirectFields {
			result[index].DirectFields[fieldIndex].View.EnumValues = append([]string(nil), surface.DirectFields[fieldIndex].View.EnumValues...)
		}
		for fieldIndex := range result[index].CollectionFields {
			result[index].CollectionFields[fieldIndex].AllowedOperations = append([]string(nil), surface.CollectionFields[fieldIndex].AllowedOperations...)
			result[index].CollectionFields[fieldIndex].View.EnumValues = append([]string(nil), surface.CollectionFields[fieldIndex].View.EnumValues...)
		}
	}
	return result
}
