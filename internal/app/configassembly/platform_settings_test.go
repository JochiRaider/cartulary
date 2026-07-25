package configassembly

import (
	"context"
	"database/sql"
	"reflect"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/JochiRaider/cartulary/internal/platform/bootstrap"
	"github.com/JochiRaider/cartulary/internal/platform/config"
	"github.com/JochiRaider/cartulary/internal/platform/objectstore"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/platform/telemetry"
	telemetryconfiguration "github.com/JochiRaider/cartulary/internal/platform/telemetry/configuration"
)

func TestPlatformSettingsProjection_Unit(t *testing.T) {
	cfg := Deployment{
		Roots: config.RootBindings{
			DatabaseStorage: config.RootBinding{BindingKind: "managed_service", Path: "db-root", ServiceRef: "database-primary"},
			ObjectStorage:   config.RootBinding{BindingKind: "filesystem_root", Path: "object-root", ServiceRef: "object-primary"},
		},
		Bootstrap: config.BootstrapConfig{FirstAdminManifestPath: "/deployment/bootstrap.json"},
		Telemetry: telemetryconfiguration.Config{
			Enabled:  true,
			Resource: telemetryconfiguration.ResourceConfig{ServiceVersion: "2026.7.25"},
		},
	}
	if got, want := PostgresBinding(cfg), (postgres.Binding{
		BindingKind: "managed_service", RootPath: "db-root", ServiceRef: "database-primary",
	}); !reflect.DeepEqual(got, want) {
		t.Fatalf("postgres binding = %#v, want %#v", got, want)
	}
	if got, want := ObjectStoreBinding(cfg), (objectstore.Binding{
		BindingKind: "filesystem_root", RootPath: "object-root", ServiceRef: "object-primary",
	}); !reflect.DeepEqual(got, want) {
		t.Fatalf("object-store binding = %#v, want %#v", got, want)
	}
	if got, want := ObjectStoreInstrumentation(cfg), (objectstore.Instrumentation{
		Enabled: true, ServiceVersion: "2026.7.25",
	}); !reflect.DeepEqual(got, want) {
		t.Fatalf("object-store instrumentation = %#v, want %#v", got, want)
	}
	if got, want := BootstrapSettings(cfg), (bootstrap.Settings{
		ManifestPath: "/deployment/bootstrap.json",
	}); !reflect.DeepEqual(got, want) {
		t.Fatalf("bootstrap settings = %#v, want %#v", got, want)
	}
}

func TestPlatformAdapterConstructorSignatures_Unit(t *testing.T) {
	var (
		_ func(context.Context, postgres.Settings) (*pgxpool.Pool, error)                                     = postgres.Setup
		_ func(postgres.Settings) (*sql.DB, error)                                                            = postgres.OpenSQL
		_ func(context.Context, objectstore.Settings, objectstore.Instrumentation) (objectstore.Store, error) = objectstore.Setup
		_ func(context.Context, objectstore.Settings) (objectstore.EnsureBucketResult, error)                 = objectstore.EnsureBucket
		_ func(context.Context, bootstrap.Settings, *pgxpool.Pool) error                                      = bootstrap.Preflight
		_ func(
			context.Context,
			telemetryconfiguration.Config,
			string,
			map[string]string,
			...telemetry.BootstrapOption,
		) (*telemetry.Runtime, error) = telemetry.Bootstrap
	)
}
