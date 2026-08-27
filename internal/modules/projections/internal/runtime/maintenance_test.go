package runtime

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/projections/providercontract"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	telemetrytest "github.com/JochiRaider/cartulary/internal/platform/telemetry/testsupport"
)

func TestMaintenanceRebuildIsOneCompleteCatalogTransaction(t *testing.T) {
	incidentID := uuid.New()
	calls := make([]string, 0, 3)
	providers := maintenanceProviders(func(providerID string, _ pgx.Tx) error {
		calls = append(calls, providerID)
		return nil
	})
	tracked := &trackingMaintenanceDB{}
	store := &Store{pool: tracked, registry: mustCharacterizationCatalog(t, providers).registry}
	if err := store.RebuildIncident(t.Context(), incidentID); err != nil {
		t.Fatalf("maintenance rebuild: %v", err)
	}
	if got, want := calls, []string{"first", "middle", "final"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("provider order = %#v, want %#v", got, want)
	}
	if tracked.begins != 1 || tracked.commits != 1 || tracked.rollbacks != 1 {
		t.Fatalf("transaction calls begin=%d commit=%d rollback=%d", tracked.begins, tracked.commits, tracked.rollbacks)
	}
}

func TestMaintenanceRebuildFailuresRollbackAndRemainContextual(t *testing.T) {
	sentinel := errors.New("injected maintenance provider failure")
	t.Run("begin", func(t *testing.T) {
		store := &Store{pool: beginFailMaintenanceDB{err: sentinel}}
		if err := store.RebuildIncident(t.Context(), uuid.New()); err == nil || !strings.Contains(err.Error(), "begin incident projection rebuild") || !errors.Is(err, sentinel) {
			t.Fatalf("begin failure = %v", err)
		}
	})

	t.Run("provider", func(t *testing.T) {
		tracked := &trackingMaintenanceDB{}
		calls := make([]string, 0, 3)
		providers := maintenanceProviders(func(providerID string, _ pgx.Tx) error {
			calls = append(calls, providerID)
			if providerID == "middle" {
				return sentinel
			}
			return nil
		})
		store := &Store{pool: tracked, registry: mustCharacterizationCatalog(t, providers).registry}
		err := store.RebuildIncident(t.Context(), uuid.New())
		if err == nil || !errors.Is(err, sentinel) || !strings.Contains(err.Error(), "rebuild incident projections") {
			t.Fatalf("provider failure = %v", err)
		}
		if got, want := calls, []string{"first", "middle"}; !reflect.DeepEqual(got, want) {
			t.Fatalf("provider failure calls = %#v, want %#v", got, want)
		}
		if tracked.begins != 1 || tracked.commits != 0 || tracked.rollbacks != 1 {
			t.Fatalf("provider failure transaction calls begin=%d commit=%d rollback=%d", tracked.begins, tracked.commits, tracked.rollbacks)
		}
	})

	t.Run("commit", func(t *testing.T) {
		tracked := &trackingMaintenanceDB{commitErr: sentinel}
		store := &Store{pool: tracked, registry: mustCharacterizationCatalog(t, maintenanceProviders(func(string, pgx.Tx) error { return nil })).registry}
		err := store.RebuildIncident(t.Context(), uuid.New())
		if err == nil || !errors.Is(err, sentinel) || !strings.Contains(err.Error(), "commit incident projection rebuild") {
			t.Fatalf("commit failure = %v", err)
		}
		if tracked.begins != 1 || tracked.commits != 1 || tracked.rollbacks != 1 {
			t.Fatalf("commit failure transaction calls begin=%d commit=%d rollback=%d", tracked.begins, tracked.commits, tracked.rollbacks)
		}
	})

	t.Run("cancellation", func(t *testing.T) {
		called := false
		store := &Store{pool: &trackingMaintenanceDB{}, registry: mustCharacterizationCatalog(t, maintenanceProviders(func(string, pgx.Tx) error { called = true; return nil })).registry}
		ctx, cancel := context.WithCancel(t.Context())
		cancel()
		err := store.RebuildIncident(ctx, uuid.New())
		if err == nil || !errors.Is(err, context.Canceled) || called {
			t.Fatalf("canceled maintenance rebuild error=%v provider_called=%t", err, called)
		}
	})
}

func TestMaintenanceRebuildTelemetryIsOneBoundedSpan(t *testing.T) {
	capture := telemetrytest.StartCapture()
	defer capture.Close(t.Context())
	sentinel := errors.New("injected begin failure")
	store := &Store{pool: beginFailMaintenanceDB{err: sentinel}}
	if err := store.RebuildIncident(t.Context(), uuid.New()); !errors.Is(err, sentinel) {
		t.Fatalf("maintenance telemetry failure: %v", err)
	}
	spans := capture.EndedSpans()
	if len(spans) != 1 {
		t.Fatalf("maintenance span count = %d, want 1", len(spans))
	}
	want := map[string]string{
		"cartulary.view_schema_id": "unknown",
		"cartulary.operation":      "rebuild",
		"cartulary.result":         "failed",
	}
	if spans[0].Name != "cartulary.workbook.projection" || !reflect.DeepEqual(spans[0].Attributes, want) {
		t.Fatalf("maintenance telemetry = %#v, want name and attributes %#v", spans[0], want)
	}
}

func maintenanceProviders(rebuild func(string, pgx.Tx) error) []Provider {
	providerIDs := []string{"first", "middle", "final"}
	providers := make([]Provider, 0, len(providerIDs))
	for index, providerID := range providerIDs {
		after := []string(nil)
		if index > 0 {
			after = []string{providerIDs[index-1]}
		}
		descriptor := characterizationProviderDescriptor(
			providerID,
			"cartulary.view.characterization."+providerID+".v1",
			[]string{"host_grid_projection", "identity_grid_projection", "indicator_grid_projection"}[index],
			after,
			providercontract.ProviderStatusActive,
			providercontract.RestoreRebuildRequired,
		)
		descriptor.Capabilities = providercontract.ProviderCapabilities{IncidentRebuild: true, RestoreRebuild: true}
		key := providerID
		providers = append(providers, Provider{
			descriptor: descriptor,
			rebuildIncidentTx: func(_ context.Context, _ *Store, tx pgx.Tx, _ uuid.UUID) error {
				return rebuild(key, tx)
			},
		})
	}
	return providers
}

type trackingMaintenanceDB struct {
	postgres.DB
	begins    int
	commits   int
	rollbacks int
	commitErr error
}

func (db *trackingMaintenanceDB) BeginTx(ctx context.Context, options pgx.TxOptions) (pgx.Tx, error) {
	db.begins++
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return &trackingMaintenanceTx{db: db}, nil
}

type trackingMaintenanceTx struct {
	pgx.Tx
	db *trackingMaintenanceDB
}

func (tx *trackingMaintenanceTx) Commit(ctx context.Context) error {
	tx.db.commits++
	if tx.db.commitErr != nil {
		return tx.db.commitErr
	}
	return nil
}

func (tx *trackingMaintenanceTx) Rollback(ctx context.Context) error {
	tx.db.rollbacks++
	return nil
}

type beginFailMaintenanceDB struct {
	postgres.DB
	err error
}

func (db beginFailMaintenanceDB) BeginTx(context.Context, pgx.TxOptions) (pgx.Tx, error) {
	return nil, db.err
}
