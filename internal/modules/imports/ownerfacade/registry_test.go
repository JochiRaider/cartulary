package ownerfacade

import (
	"context"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
)

func TestImportOwnerCreateRegistryRequiresExactBindings(t *testing.T) {
	t.Parallel()

	expected := []ImportOwnerCreateBinding{
		{TargetViewSchemaID: "cartulary.view.hosts.v1", FacadeID: "entities.host.import_create"},
		{TargetViewSchemaID: "cartulary.view.identities.v1", FacadeID: "entities.identity.import_create"},
	}
	host := importOwnerCreateTestFacade(
		t,
		expected[0],
	)
	identity := importOwnerCreateTestFacade(
		t,
		expected[1],
	)

	registry, err := NewImportOwnerCreateRegistry(expected, identity, host)
	if err != nil {
		t.Fatalf("construct exact registry: %v", err)
	}
	bindings := registry.Bindings()
	if len(bindings) != 2 ||
		bindings[0].TargetViewSchemaID != "cartulary.view.hosts.v1" ||
		bindings[1].TargetViewSchemaID != "cartulary.view.identities.v1" {
		t.Fatalf("registry bindings are not stable: %#v", bindings)
	}
	if resolved, ok := registry.Resolve(
		"cartulary.view.hosts.v1",
		"entities.host.import_create",
	); !ok || resolved != host {
		t.Fatalf("resolve exact host binding = %#v, %v", resolved, ok)
	}
	if _, ok := registry.Resolve(
		"cartulary.view.hosts.v1",
		"entities.identity.import_create",
	); ok {
		t.Fatal("registry resolved a facade-id mismatch")
	}
}

func TestImportOwnerCreateRegistryFailsClosed(t *testing.T) {
	t.Parallel()

	hostBinding := ImportOwnerCreateBinding{
		TargetViewSchemaID: "cartulary.view.hosts.v1",
		FacadeID:           "entities.host.import_create",
	}
	identityBinding := ImportOwnerCreateBinding{
		TargetViewSchemaID: "cartulary.view.identities.v1",
		FacadeID:           "entities.identity.import_create",
	}
	host := importOwnerCreateTestFacade(t, hostBinding)
	identity := importOwnerCreateTestFacade(t, identityBinding)

	tests := []struct {
		name     string
		expected []ImportOwnerCreateBinding
		facades  []ImportOwnerCreateFacade
		want     string
	}{
		{
			name:     "missing",
			expected: []ImportOwnerCreateBinding{hostBinding, identityBinding},
			facades:  []ImportOwnerCreateFacade{host},
			want:     "missing import owner-create facade",
		},
		{
			name:     "duplicate",
			expected: []ImportOwnerCreateBinding{hostBinding},
			facades:  []ImportOwnerCreateFacade{host, host},
			want:     "duplicate import owner-create facade",
		},
		{
			name:     "unexpected",
			expected: []ImportOwnerCreateBinding{hostBinding},
			facades:  []ImportOwnerCreateFacade{host, identity},
			want:     "unexpected import owner-create facade",
		},
		{
			name:     "mismatch",
			expected: []ImportOwnerCreateBinding{hostBinding},
			facades: []ImportOwnerCreateFacade{importOwnerCreateTestFacade(t, ImportOwnerCreateBinding{
				TargetViewSchemaID: hostBinding.TargetViewSchemaID,
				FacadeID:           identityBinding.FacadeID,
			})},
			want: "facade mismatch",
		},
		{
			name:     "nil",
			expected: []ImportOwnerCreateBinding{hostBinding},
			facades:  []ImportOwnerCreateFacade{nil},
			want:     "nil facade",
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			_, err := NewImportOwnerCreateRegistry(test.expected, test.facades...)
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("registry error = %v, want containing %q", err, test.want)
			}
		})
	}
}

func TestBoundImportOwnerCreateFacadeRejectsWrongTarget(t *testing.T) {
	t.Parallel()

	facade := importOwnerCreateTestFacade(t, ImportOwnerCreateBinding{
		TargetViewSchemaID: "cartulary.view.hosts.v1",
		FacadeID:           "entities.host.import_create",
	})
	_, err := facade.CreateImportRowTx(context.Background(), nil, ImportOwnerCreateCommand{
		Request: ImportOwnerCreateRequest{TargetViewSchemaID: "cartulary.view.identities.v1"},
	})
	if err == nil || !strings.Contains(err.Error(), "is bound to") {
		t.Fatalf("wrong-target error = %v", err)
	}
}

func importOwnerCreateTestFacade(
	t testing.TB,
	binding ImportOwnerCreateBinding,
) ImportOwnerCreateFacade {
	t.Helper()
	facade, err := NewImportOwnerCreateFacade(
		binding,
		func(
			context.Context,
			pgx.Tx,
			ImportOwnerCreateCommand,
		) (ImportOwnerCreateResponse, error) {
			return ImportOwnerCreateResponse{}, nil
		},
	)
	if err != nil {
		t.Fatalf("construct test facade: %v", err)
	}
	return facade
}
