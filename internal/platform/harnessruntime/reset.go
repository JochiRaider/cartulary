package harnessruntime

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/JochiRaider/cartulary/internal/platform/harnessredact"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/objectstore"
)

const (
	testRuntimeResetSchemaID    = "cartulary.test.runtime_reset.v1"
	testRuntimeIdentitySchemaID = "cartulary.test.runtime_identity.v1"
	testClockModuleOverrideKey  = "test_clock"
	testRoutesEnabledEnv        = httpapi.TestRoutesEnabledEnv
	testRouteTokenEnv           = httpapi.TestRouteTokenEnv
	testRuntimeMarkerEnv        = httpapi.TestRuntimeMarkerEnv
	testRuntimeMarkerValue      = httpapi.TestRuntimeMarkerValue
	testRouteTokenHeader        = httpapi.TestRouteTokenHeader
	testRuntimeResetTimeout     = 30 * time.Second
)

type testRuntimeResetService struct {
	resetDatabase func(context.Context) error
	postgres      *pgxpool.Pool
	objectStore   objectstore.Store
	guard         httpapi.TestRouteGuard
	resetHooks    []func()
	resetMu       sync.Mutex
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
	PartialFailure             bool              `json:"partial_failure"`
	PartialFailureDetails      map[string]any    `json:"partial_failure_details,omitempty"`
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

type testRuntimeIdentityResult struct {
	SchemaID      string `json:"schema_id"`
	RuntimeMarker string `json:"runtime_marker"`
	ServerPID     int    `json:"server_pid"`
	TestRoutes    bool   `json:"test_routes_enabled"`
}

func RegisterTestRuntimeResetRoute(resetHooks ...func()) httpapi.RouteRegistrar {
	return func(mux *http.ServeMux, deps httpapi.DependencySet) error {
		if !httpapi.TestRoutesEnabled(deps.Env) {
			return nil
		}
		guard, err := httpapi.NewTestRouteGuard(deps.Env)
		if err != nil {
			return fmt.Errorf("register test runtime reset route: %w", err)
		}
		if deps.Postgres == nil {
			return fmt.Errorf("register test runtime reset route: postgres dependency is required")
		}
		if deps.ObjectStore == nil {
			return fmt.Errorf("register test runtime reset route: object store dependency is required")
		}
		effectiveResetHooks := append([]func(){}, resetHooks...)
		if clearable, ok := deps.PublicErrorFaults.(interface{ Clear() }); ok {
			effectiveResetHooks = append(effectiveResetHooks, clearable.Clear)
		}
		if resettable, ok := deps.ModuleOverrides[testClockModuleOverrideKey].(interface{ Reset() time.Time }); ok {
			effectiveResetHooks = append(effectiveResetHooks, func() {
				_ = resettable.Reset()
			})
		}
		service := &testRuntimeResetService{
			resetDatabase: deps.TestResetDatabase,
			postgres:      deps.Postgres,
			objectStore:   deps.ObjectStore,
			guard:         guard,
			resetHooks:    effectiveResetHooks,
		}
		mux.HandleFunc("GET /api/v1/test/runtime/identity", service.handleIdentity)
		mux.HandleFunc("POST /api/v1/test/runtime/reset", service.handleReset)
		return nil
	}
}

func (s *testRuntimeResetService) handleIdentity(w http.ResponseWriter, r *http.Request) {
	if !s.guard.Authorize(w, r) {
		return
	}
	_ = httpapi.WriteSuccess(w, r, http.StatusOK, testRuntimeIdentityResult{
		SchemaID:      testRuntimeIdentitySchemaID,
		RuntimeMarker: testRuntimeMarkerValue,
		ServerPID:     os.Getpid(),
		TestRoutes:    true,
	})
}

func (s *testRuntimeResetService) handleReset(w http.ResponseWriter, r *http.Request) {
	if !s.guard.Authorize(w, r) {
		return
	}
	if err := validateTestRuntimeResetBody(r); err != nil {
		_ = httpapi.WriteError(w, r, http.StatusBadRequest, "invalid_test_reset_request", "invalid test runtime reset request", map[string]any{
			"reason": err.Error(),
		})
		return
	}
	if !s.resetMu.TryLock() {
		_ = httpapi.WriteError(w, r, http.StatusConflict, "test_runtime_reset_in_progress", "test runtime reset already in progress", map[string]any{})
		return
	}
	defer s.resetMu.Unlock()
	for _, resetHook := range s.resetHooks {
		if resetHook != nil {
			resetHook()
		}
	}

	ctx, cancel := context.WithTimeout(r.Context(), testRuntimeResetTimeout)
	defer cancel()

	beforeGooseVersions, err := countTableRows(ctx, s.postgres, "goose_db_version")
	if err != nil {
		writeTestRuntimeResetError(w, r, "count migration metadata before reset", err, false, nil)
		return
	}
	beforeLineageRows, err := countTableRows(ctx, s.postgres, "schema_migration_lineage")
	if err != nil {
		writeTestRuntimeResetError(w, r, "count migration lineage before reset", err, false, nil)
		return
	}

	tables, err := listMutablePublicTables(ctx, s.postgres)
	if err != nil {
		writeTestRuntimeResetError(w, r, "list mutable tables", err, false, nil)
		return
	}
	if s.resetDatabase != nil {
		if err := s.resetDatabase(ctx); err != nil {
			writeTestRuntimeResetError(w, r, "reset database through Recovery adapter", err, false, map[string]any{
				"tables_reset": tables,
			})
			return
		}
	}
	afterGooseVersions, err := countTableRows(ctx, s.postgres, "goose_db_version")
	if err != nil {
		writeTestRuntimeResetError(w, r, "count migration metadata after reset", err, false, map[string]any{
			"tables_reset": tables,
		})
		return
	}
	afterLineageRows, err := countTableRows(ctx, s.postgres, "schema_migration_lineage")
	if err != nil {
		writeTestRuntimeResetError(w, r, "count migration lineage after reset", err, false, map[string]any{
			"tables_reset": tables,
		})
		return
	}
	counts, err := readPostResetCounts(ctx, s.postgres)
	if err != nil {
		writeTestRuntimeResetError(w, r, "read post-reset counts", err, false, map[string]any{
			"tables_reset": tables,
		})
		return
	}
	if beforeGooseVersions != afterGooseVersions || afterGooseVersions == 0 || beforeLineageRows != afterLineageRows || afterLineageRows != 1 {
		writeTestRuntimeResetError(w, r, "verify migration metadata preservation", errors.New("migration metadata was not preserved"), false, map[string]any{
			"tables_reset": tables,
		})
		return
	}
	if counts.ActiveDeploymentAdmins != 1 || counts.BootstrapMarkers != 1 {
		writeTestRuntimeResetError(w, r, "verify bootstrap admin restored", errors.New("bootstrap admin was not restored"), false, map[string]any{
			"tables_reset": tables,
		})
		return
	}
	objectsRemoved, err := clearConfiguredObjectBucket(ctx, s.objectStore)
	if err != nil {
		details := map[string]any{
			"tables_reset":         tables,
			"object_count_removed": objectsRemoved,
		}
		if objectsAfter, countErr := countConfiguredObjectBucket(context.Background(), s.objectStore); countErr == nil {
			details["object_count_after"] = objectsAfter
		}
		writeTestRuntimeResetError(w, r, "clear object store bucket", err, true, details)
		return
	}
	objectsAfter, err := countConfiguredObjectBucket(ctx, s.objectStore)
	if err != nil {
		writeTestRuntimeResetError(w, r, "count object store bucket after reset", err, true, map[string]any{
			"tables_reset":         tables,
			"object_count_removed": objectsRemoved,
		})
		return
	}
	if objectsAfter != 0 {
		writeTestRuntimeResetError(w, r, "verify object store bucket empty", errors.New("object store bucket was not empty after reset"), true, map[string]any{
			"tables_reset":         tables,
			"object_count_removed": objectsRemoved,
			"object_count_after":   objectsAfter,
		})
		return
	}

	result := testRuntimeResetResult{
		SchemaID:                   testRuntimeResetSchemaID,
		ResetID:                    uuid.NewString(),
		TablesReset:                tables,
		MutableTableCount:          len(tables),
		ObjectCountRemoved:         objectsRemoved,
		ObjectCountAfter:           objectsAfter,
		MigrationMetadataPreserved: beforeGooseVersions == afterGooseVersions && afterGooseVersions > 0 && beforeLineageRows == afterLineageRows && afterLineageRows == 1,
		BootstrapAdminRestored:     counts.ActiveDeploymentAdmins == 1 && counts.BootstrapMarkers == 1,
		PartialFailure:             false,
		PostResetCounts:            counts,
	}

	_ = httpapi.WriteSuccess(w, r, http.StatusOK, result)
}

// ResetDatabase performs the destructive database portion of a harness reset
// through a caller-supplied Recovery pool. Product server composition never
// supplies this capability; browser and service fixtures invoke it out of
// process before calling the route that clears runtime-local state.
func ResetDatabase(ctx context.Context, recovery *pgxpool.Pool, restoreBootstrap func(context.Context, pgx.Tx) error) error {
	if recovery == nil || restoreBootstrap == nil {
		return errors.New("recovery reset dependencies are required")
	}
	tx, err := recovery.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return newResetFailure("recovery_reset_begin_failed", err)
	}
	defer func() {
		_ = tx.Rollback(context.Background())
	}()
	tables, err := listMutablePublicTables(ctx, tx)
	if err != nil {
		return newResetFailure("recovery_reset_inventory_failed", err)
	}
	if err := truncateTables(ctx, tx, tables); err != nil {
		return newResetFailure("recovery_reset_truncate_failed", err)
	}
	if err := resetTableSequences(ctx, tx, tables); err != nil {
		return newResetFailure("recovery_reset_sequence_failed", err)
	}
	if err := restoreBootstrap(ctx, tx); err != nil {
		return newResetFailure("recovery_reset_bootstrap_failed", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return newResetFailure("recovery_reset_commit_failed", err)
	}
	return nil
}

func validateTestRuntimeResetBody(r *http.Request) error {
	if r.Body == nil {
		return nil
	}
	defer r.Body.Close()
	raw, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		return fmt.Errorf("read body: %w", err)
	}
	trimmed := strings.TrimSpace(string(raw))
	if trimmed == "" {
		return nil
	}
	if !strings.HasPrefix(trimmed, "{") {
		return errors.New("body must be an empty JSON object")
	}
	var payload map[string]json.RawMessage
	decoder := json.NewDecoder(strings.NewReader(trimmed))
	if err := decoder.Decode(&payload); err != nil {
		return fmt.Errorf("body must be an empty JSON object: %w", err)
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		return errors.New("body must contain one JSON value")
	}
	if len(payload) > 0 {
		return errors.New("body object must not contain members")
	}
	return nil
}

func writeTestRuntimeResetError(w http.ResponseWriter, r *http.Request, action string, err error, partialFailure bool, details map[string]any) {
	status := http.StatusInternalServerError
	code := "test_runtime_reset_failed"
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(r.Context().Err(), context.DeadlineExceeded) {
		status = http.StatusServiceUnavailable
		code = "test_runtime_reset_timeout"
	}
	responseDetails := map[string]any{
		"failed_action":           action,
		"partial_failure":         partialFailure,
		"partial_failure_details": map[string]any{},
		"error":                   harnessredact.String(err.Error()),
	}
	for key, value := range details {
		responseDetails["partial_failure_details"].(map[string]any)[key] = value
		if key == "object_count_after" {
			responseDetails[key] = value
		}
	}
	_ = httpapi.WriteError(w, r, status, code, action, responseDetails)
}

type resetDB interface {
	Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
	Query(context.Context, string, ...any) (pgx.Rows, error)
	QueryRow(context.Context, string, ...any) pgx.Row
}

type resetFailure struct {
	reason string
	cause  error
}

func (failure *resetFailure) Error() string {
	return failure.reason
}

func (failure *resetFailure) Unwrap() error {
	return failure.cause
}

func newResetFailure(reason string, cause error) error {
	return &resetFailure{reason: reason, cause: cause}
}

func listMutablePublicTables(ctx context.Context, db resetDB) ([]string, error) {
	rows, err := db.Query(ctx, `
		SELECT table_name
		FROM information_schema.tables
		WHERE table_schema = 'public'
		  AND table_type = 'BASE TABLE'
		  AND table_name NOT IN ('goose_db_version', 'schema_migration_lineage')
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

func truncateTables(ctx context.Context, db resetDB, tables []string) error {
	if len(tables) == 0 {
		return nil
	}

	identifiers := make([]string, 0, len(tables))
	for _, table := range tables {
		identifiers = append(identifiers, pgx.Identifier{"public", table}.Sanitize())
	}
	_, err := db.Exec(ctx, "TRUNCATE TABLE "+strings.Join(identifiers, ", ")+" CASCADE")
	return err
}

func resetTableSequences(ctx context.Context, db resetDB, tables []string) error {
	for _, table := range tables {
		rows, err := db.Query(ctx, `
SELECT column_name
FROM information_schema.columns
WHERE table_schema = 'public'
  AND table_name = $1
  AND column_default LIKE 'nextval(%'
ORDER BY ordinal_position`, table)
		if err != nil {
			return err
		}
		columns := make([]string, 0)
		for rows.Next() {
			var column string
			if err := rows.Scan(&column); err != nil {
				rows.Close()
				return err
			}
			columns = append(columns, column)
		}
		if err := rows.Err(); err != nil {
			rows.Close()
			return err
		}
		rows.Close()
		for _, column := range columns {
			var sequence string
			if err := db.QueryRow(ctx, `SELECT pg_get_serial_sequence($1, $2)`, "public."+table, column).Scan(&sequence); err != nil {
				return err
			}
			if _, err := db.Exec(ctx, `SELECT setval($1, 1, false)`, sequence); err != nil {
				return err
			}
		}
	}
	return nil
}

func clearConfiguredObjectBucket(ctx context.Context, client objectstore.Store) (int, error) {
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

func countConfiguredObjectBucket(ctx context.Context, client objectstore.Store) (int, error) {
	objects, err := client.ListObjects(ctx, "")
	if err != nil {
		return 0, err
	}
	return len(objects), nil
}

func countTableRows(ctx context.Context, db resetDB, table string) (int, error) {
	var count int
	err := db.QueryRow(ctx, "SELECT COUNT(*) FROM "+pgx.Identifier{"public", table}.Sanitize()).Scan(&count)
	return count, err
}

func readPostResetCounts(ctx context.Context, db resetDB) (testRuntimeCounts, error) {
	var counts testRuntimeCounts
	err := db.QueryRow(ctx, `
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
