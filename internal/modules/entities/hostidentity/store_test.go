package hostidentity

import (
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/collaboration"
	"github.com/JochiRaider/cartulary/internal/modules/entities/entitycontract"
	"github.com/JochiRaider/cartulary/internal/modules/entities/projectionports"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	"github.com/JochiRaider/cartulary/internal/modules/revisions/conflicts"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

type storePostgresStub struct{ postgres.DB }
type storeProjectionMutationRowsStub struct{ projectionports.MutationRows }
type storeProjectionQueryReaderStub struct{ projectionports.QueryReader }
type storeIdempotencyStub struct{ conflicts.IdempotencyPort }
type storePublicationStub struct {
	collaboration.RecordChangedAppender
}

func TestHostIdentityStoreCompositionRequiresCompleteDependencies_Unit(t *testing.T) {
	valid := func() StoreDependencies {
		return StoreDependencies{
			Postgres:               storePostgresStub{},
			Revisions:              &revisions.Appender{},
			ProjectionMutationRows: storeProjectionMutationRowsStub{},
			ProjectionQueryReader:  storeProjectionQueryReaderStub{},
			KeepSavedIdempotency:   storeIdempotencyStub{},
			Collaboration:          storePublicationStub{},
		}
	}
	tests := []struct {
		name       string
		dependency string
		mutate     func(*StoreDependencies)
	}{
		{name: "missing Postgres", dependency: "Postgres", mutate: func(dependencies *StoreDependencies) { dependencies.Postgres = nil }},
		{name: "typed-nil Postgres", dependency: "Postgres", mutate: func(dependencies *StoreDependencies) { dependencies.Postgres = (*storePostgresStub)(nil) }},
		{name: "missing Revisions", dependency: "Revisions", mutate: func(dependencies *StoreDependencies) { dependencies.Revisions = nil }},
		{name: "typed-nil Revisions", dependency: "Revisions", mutate: func(dependencies *StoreDependencies) { dependencies.Revisions = (*revisions.Appender)(nil) }},
		{name: "missing ProjectionMutationRows", dependency: "ProjectionMutationRows", mutate: func(dependencies *StoreDependencies) { dependencies.ProjectionMutationRows = nil }},
		{name: "typed-nil ProjectionMutationRows", dependency: "ProjectionMutationRows", mutate: func(dependencies *StoreDependencies) {
			dependencies.ProjectionMutationRows = (*storeProjectionMutationRowsStub)(nil)
		}},
		{name: "missing ProjectionQueryReader", dependency: "ProjectionQueryReader", mutate: func(dependencies *StoreDependencies) { dependencies.ProjectionQueryReader = nil }},
		{name: "typed-nil ProjectionQueryReader", dependency: "ProjectionQueryReader", mutate: func(dependencies *StoreDependencies) {
			dependencies.ProjectionQueryReader = (*storeProjectionQueryReaderStub)(nil)
		}},
		{name: "missing KeepSavedIdempotency", dependency: "KeepSavedIdempotency", mutate: func(dependencies *StoreDependencies) { dependencies.KeepSavedIdempotency = nil }},
		{name: "typed-nil KeepSavedIdempotency", dependency: "KeepSavedIdempotency", mutate: func(dependencies *StoreDependencies) {
			dependencies.KeepSavedIdempotency = (*storeIdempotencyStub)(nil)
		}},
		{name: "missing Collaboration", dependency: "Collaboration", mutate: func(dependencies *StoreDependencies) { dependencies.Collaboration = nil }},
		{name: "typed-nil Collaboration", dependency: "Collaboration", mutate: func(dependencies *StoreDependencies) {
			dependencies.Collaboration = (*storePublicationStub)(nil)
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			dependencies := valid()
			test.mutate(&dependencies)
			store, err := NewStore(dependencies)
			if store != nil {
				t.Fatalf("NewStore result = %#v, want nil", store)
			}
			want := "compose Host/Identity store: " + test.dependency + " is required"
			if err == nil || err.Error() != want {
				t.Fatalf("NewStore error = %v, want %q", err, want)
			}
		})
	}

	store, err := NewStore(valid())
	if err != nil || store == nil {
		t.Fatalf("valid NewStore result = %#v, %v", store, err)
	}

	t.Run("declaration order", func(t *testing.T) {
		store, err := NewStore(StoreDependencies{})
		if store != nil || err == nil || err.Error() != "compose Host/Identity store: Postgres is required" {
			t.Fatalf("empty NewStore result = %#v, %v", store, err)
		}
	})

	t.Run("import create facade dependencies", func(t *testing.T) {
		validImport := func() ImportDependencies {
			return ImportDependencies{
				Revisions:              &revisions.Appender{},
				ProjectionMutationRows: storeProjectionMutationRowsStub{},
				Collaboration:          storePublicationStub{},
			}
		}
		importTests := []struct {
			name       string
			dependency string
			mutate     func(*ImportDependencies)
		}{
			{name: "missing Revisions", dependency: "Revisions", mutate: func(dependencies *ImportDependencies) { dependencies.Revisions = nil }},
			{name: "typed-nil Revisions", dependency: "Revisions", mutate: func(dependencies *ImportDependencies) { dependencies.Revisions = (*revisions.Appender)(nil) }},
			{name: "missing ProjectionMutationRows", dependency: "ProjectionMutationRows", mutate: func(dependencies *ImportDependencies) { dependencies.ProjectionMutationRows = nil }},
			{name: "typed-nil ProjectionMutationRows", dependency: "ProjectionMutationRows", mutate: func(dependencies *ImportDependencies) {
				dependencies.ProjectionMutationRows = (*storeProjectionMutationRowsStub)(nil)
			}},
			{name: "missing Collaboration", dependency: "Collaboration", mutate: func(dependencies *ImportDependencies) { dependencies.Collaboration = nil }},
			{name: "typed-nil Collaboration", dependency: "Collaboration", mutate: func(dependencies *ImportDependencies) {
				dependencies.Collaboration = (*storePublicationStub)(nil)
			}},
		}
		for _, test := range importTests {
			t.Run(test.name, func(t *testing.T) {
				dependencies := validImport()
				test.mutate(&dependencies)
				facade, err := NewImportCreateFacade(entitycontract.HostsViewSchemaID, "entities.hosts", dependencies)
				if facade != nil {
					t.Fatalf("NewImportCreateFacade result = %#v, want nil", facade)
				}
				want := "compose Host/Identity import create facade: " + test.dependency + " is required"
				if err == nil || err.Error() != want {
					t.Fatalf("NewImportCreateFacade error = %v, want %q", err, want)
				}
			})
		}
		facade, err := NewImportCreateFacade(entitycontract.HostsViewSchemaID, "entities.hosts", validImport())
		if err != nil || facade == nil {
			t.Fatalf("valid NewImportCreateFacade result = %#v, %v", facade, err)
		}
	})
}
