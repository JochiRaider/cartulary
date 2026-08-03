package conflicts

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestFieldResolverCatalogRejectsDuplicateMissingAndInvalidProviders(t *testing.T) {
	t.Parallel()
	field := FieldDescriptor{
		FieldKey:                "note.body",
		ValueKind:               "text",
		Writable:                true,
		ConflictResolutionClass: "text_compare_merge",
	}
	valid := FieldResolverContribution{
		ProviderID:    "artifacts.notes#cartulary.view.notes.v1",
		SourceOwnerID: "artifacts",
		ViewSchemaID:  "cartulary.view.notes.v1",
		Fields:        []FieldDescriptor{field},
	}

	tests := []struct {
		name          string
		required      []string
		contributions []FieldResolverContribution
		want          error
	}{
		{name: "missing", required: []string{valid.ViewSchemaID}, want: ErrMissingFieldResolver},
		{name: "duplicate provider", required: []string{valid.ViewSchemaID}, contributions: []FieldResolverContribution{valid, valid}, want: ErrDuplicateFieldResolver},
		{name: "unexpected", contributions: []FieldResolverContribution{valid}, want: ErrUnexpectedFieldResolver},
		{name: "invalid field", required: []string{valid.ViewSchemaID}, contributions: []FieldResolverContribution{{ProviderID: valid.ProviderID, SourceOwnerID: valid.SourceOwnerID, ViewSchemaID: valid.ViewSchemaID, Fields: []FieldDescriptor{{FieldKey: field.FieldKey, Writable: true}}}}, want: ErrInvalidFieldContract},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			_, err := NewFieldResolverCatalog(test.required, test.contributions...)
			if !errors.Is(err, test.want) {
				t.Fatalf("catalog error = %v, want %v", err, test.want)
			}
		})
	}
}

func TestFieldResolverCatalogIsImmutableAndFailClosed(t *testing.T) {
	t.Parallel()
	fields := []FieldDescriptor{{
		FieldKey:                "note.body",
		ValueKind:               "text",
		Writable:                true,
		ConflictResolutionClass: "text_compare_merge",
	}}
	catalog, err := NewFieldResolverCatalog(
		[]string{"cartulary.view.notes.v1"},
		FieldResolverContribution{
			ProviderID:    "artifacts.notes#cartulary.view.notes.v1",
			SourceOwnerID: "artifacts",
			ViewSchemaID:  "cartulary.view.notes.v1",
			Fields:        fields,
		},
	)
	if err != nil {
		t.Fatalf("build catalog: %v", err)
	}
	fields[0].ConflictResolutionClass = "mutated"
	descriptor, err := catalog.ResolveWritableField("cartulary.view.notes.v1", "note.body")
	if err != nil {
		t.Fatalf("resolve writable field: %v", err)
	}
	if descriptor.ConflictResolutionClass != "text_compare_merge" {
		t.Fatalf("catalog changed through contribution slice: %#v", descriptor)
	}
	if _, err := catalog.ResolveWritableField("cartulary.view.notes.v1", "missing"); !errors.Is(err, ErrFieldNotFound) {
		t.Fatalf("missing field error = %v", err)
	}
	if _, err := catalog.ResolveViewSchema("cartulary.view.unknown.v1"); !errors.Is(err, ErrMissingFieldResolver) {
		t.Fatalf("missing view error = %v", err)
	}
}

func TestConflictWindowProductionSourcesDoNotReadGlobalViewSchemaRegistry(t *testing.T) {
	t.Parallel()
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve current source path")
	}
	directory := filepath.Dir(currentFile)
	targets := []string{
		filepath.Join(directory, "field_resolver.go"),
		filepath.Join(directory, "revision_window_reader.go"),
		filepath.Join(directory, "conflict_window.go"),
	}
	for _, target := range targets {
		data, err := os.ReadFile(target)
		if err != nil {
			t.Fatalf("read %s: %v", target, err)
		}
		for _, forbidden := range []string{"internal/platform/viewschema", "viewschema.Lookup", "viewschema.List"} {
			if strings.Contains(string(data), forbidden) {
				t.Fatalf("%s accesses global view-schema state through %q", target, forbidden)
			}
		}
	}
}
