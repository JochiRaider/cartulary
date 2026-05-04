package app

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/JochiRaider/cartulary/internal/platform/config"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/objectstore"
)

const testRuntimeResetSchemaID = "cartulary.test.runtime_reset.v1"

type testRuntimeResetService struct {
	cfg         config.Config
	env         map[string]string
	postgres    *pgxpool.Pool
	objectStore objectstore.Store
}

type testRuntimeResetResult struct {
	SchemaID                   string            `json:"schema_id"`
	ResetID                    string            `json:"reset_id"`
	TablesReset                []string          `json:"tables_reset"`
	MutableTableCount          int               `json:"mutable_table_count"`
	ObjectCountRemoved         int               `json:"object_count_removed"`
	ObjectCountAfter           int               `json:"object_count_after"`
	MigrationMetadataPreserved bool              `json:"migration_metadata_preserved"`
	BootstrapAdminRestored     bool              `json:"bootstrap_admin_restored"`
	PostResetCounts            testRuntimeCounts `json:"post_reset_counts"`
}

type testRuntimeCounts struct {
	ActiveDeploymentAdmins int `json:"active_deployment_admins"`
	BootstrapMarkers       int `json:"bootstrap_markers"`
	Incidents              int `json:"incidents"`
	Records                int `json:"records"`
	UserSessions           int `json:"user_sessions"`
	RouteIdempotency       int `json:"route_idempotency"`
}

func RegisterTestRuntimeResetRoute() httpapi.RouteRegistrar {
	return func(mux *http.ServeMux, deps httpapi.DependencySet) error {
		if deps.Postgres == nil {
			return fmt.Errorf("register test runtime reset route: postgres dependency is required")
		}
		if deps.ObjectStore == nil {
			return fmt.Errorf("register test runtime reset route: object store dependency is required")
		}
		service := &testRuntimeResetService{
			cfg:         deps.Config,
			env:         deps.Env,
			postgres:    deps.Postgres,
			objectStore: deps.ObjectStore,
		}
		mux.HandleFunc("POST /api/v1/test/runtime/reset", service.handleReset)
		return nil
	}
}

func (s *testRuntimeResetService) handleReset(w http.ResponseWriter, r *http.Request) {
	beforeGooseVersions, err := countTableRows(r.Context(), s.postgres, "goose_db_version")
	if err != nil {
		writeTestRuntimeResetError(w, r, "count migration metadata before reset", err)
		return
	}

	tables, err := listMutablePublicTables(r.Context(), s.postgres)
	if err != nil {
		writeTestRuntimeResetError(w, r, "list mutable tables", err)
		return
	}
	if err := truncateTables(r.Context(), s.postgres, tables); err != nil {
		writeTestRuntimeResetError(w, r, "truncate mutable tables", err)
		return
	}
	if err := runBootstrap(r.Context(), s.cfg, postgresBootstrapStore{pool: s.postgres}, osBootstrapManifestFS{}, deriveBootstrapPasswordHash); err != nil {
		writeTestRuntimeResetError(w, r, "restore bootstrap admin", err)
		return
	}

	objectsRemoved, err := clearConfiguredObjectBucket(r.Context(), s.cfg, s.env, s.objectStore)
	if err != nil {
		writeTestRuntimeResetError(w, r, "clear object store bucket", err)
		return
	}
	objectsAfter, err := countConfiguredObjectBucket(r.Context(), s.cfg, s.env, s.objectStore)
	if err != nil {
		writeTestRuntimeResetError(w, r, "count object store bucket after reset", err)
		return
	}
	afterGooseVersions, err := countTableRows(r.Context(), s.postgres, "goose_db_version")
	if err != nil {
		writeTestRuntimeResetError(w, r, "count migration metadata after reset", err)
		return
	}
	counts, err := readPostResetCounts(r.Context(), s.postgres)
	if err != nil {
		writeTestRuntimeResetError(w, r, "read post-reset counts", err)
		return
	}

	result := testRuntimeResetResult{
		SchemaID:                   testRuntimeResetSchemaID,
		ResetID:                    uuid.NewString(),
		TablesReset:                tables,
		MutableTableCount:          len(tables),
		ObjectCountRemoved:         objectsRemoved,
		ObjectCountAfter:           objectsAfter,
		MigrationMetadataPreserved: beforeGooseVersions == afterGooseVersions && afterGooseVersions > 0,
		BootstrapAdminRestored:     counts.ActiveDeploymentAdmins == 1 && counts.BootstrapMarkers == 1,
		PostResetCounts:            counts,
	}

	_ = httpapi.WriteSuccess(w, r, http.StatusOK, result)
}

func writeTestRuntimeResetError(w http.ResponseWriter, r *http.Request, action string, err error) {
	_ = httpapi.WriteError(w, r, http.StatusInternalServerError, "test_runtime_reset_failed", action, map[string]any{
		"error": err.Error(),
	})
}

func listMutablePublicTables(ctx context.Context, pool *pgxpool.Pool) ([]string, error) {
	rows, err := pool.Query(ctx, `
		SELECT table_name
		FROM information_schema.tables
		WHERE table_schema = 'public'
		  AND table_type = 'BASE TABLE'
		  AND table_name <> 'goose_db_version'
		ORDER BY table_name
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var table string
		if err := rows.Scan(&table); err != nil {
			return nil, err
		}
		tables = append(tables, table)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return tables, nil
}

func truncateTables(ctx context.Context, pool *pgxpool.Pool, tables []string) error {
	if len(tables) == 0 {
		return nil
	}

	identifiers := make([]string, 0, len(tables))
	for _, table := range tables {
		identifiers = append(identifiers, pgx.Identifier{"public", table}.Sanitize())
	}
	_, err := pool.Exec(ctx, "TRUNCATE TABLE "+strings.Join(identifiers, ", ")+" RESTART IDENTITY CASCADE")
	return err
}

func clearConfiguredObjectBucket(ctx context.Context, cfg config.Config, env map[string]string, client objectstore.Store) (int, error) {
	removed := 0
	objects, err := client.ListObjects(ctx, "")
	if err != nil {
		return 0, err
	}
	for _, objectInfo := range objects {
		if err := client.DeleteObject(ctx, objectInfo.Key); err != nil {
			return removed, err
		}
		removed++
	}
	return removed, nil
}

func countConfiguredObjectBucket(ctx context.Context, cfg config.Config, env map[string]string, client objectstore.Store) (int, error) {
	objects, err := client.ListObjects(ctx, "")
	if err != nil {
		return 0, err
	}
	return len(objects), nil
}

func countTableRows(ctx context.Context, pool *pgxpool.Pool, table string) (int, error) {
	var count int
	err := pool.QueryRow(ctx, "SELECT COUNT(*) FROM "+pgx.Identifier{"public", table}.Sanitize()).Scan(&count)
	return count, err
}

func readPostResetCounts(ctx context.Context, pool *pgxpool.Pool) (testRuntimeCounts, error) {
	var counts testRuntimeCounts
	err := pool.QueryRow(ctx, `
		SELECT
			(SELECT COUNT(*) FROM users WHERE is_active = true AND is_deployment_admin = true),
			(SELECT COUNT(*) FROM deployment_bootstrap_state),
			(SELECT COUNT(*) FROM incidents),
			(SELECT COUNT(*) FROM records),
			(SELECT COUNT(*) FROM user_sessions),
			(SELECT COUNT(*) FROM route_idempotency)
	`).Scan(
		&counts.ActiveDeploymentAdmins,
		&counts.BootstrapMarkers,
		&counts.Incidents,
		&counts.Records,
		&counts.UserSessions,
		&counts.RouteIdempotency,
	)
	return counts, err
}
