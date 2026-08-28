package harnessruntime

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/processlease"
)

const (
	testRuntimeIdentitySchemaID = "cartulary.test.runtime_identity.v1"
	testRoutesEnabledEnv        = httpapi.TestRoutesEnabledEnv
	testRouteTokenEnv           = httpapi.TestRouteTokenEnv
	testRuntimeMarkerEnv        = httpapi.TestRuntimeMarkerEnv
	testRuntimeMarkerValue      = httpapi.TestRuntimeMarkerValue
	testRouteTokenHeader        = httpapi.TestRouteTokenHeader
	testRuntimeResetTimeout     = 30 * time.Second
	recoveryLeaseLossDetection  = 5 * time.Second
)

type DatabaseResetResult struct {
	TablesReset                []string
	TableCounts                []DatabaseResetTableCount
	MutableTableCount          int
	MigrationMetadataPreserved bool
	BootstrapAdminRestored     bool
	PostResetCounts            DatabaseResetCounts
}

type DatabaseResetTableCount struct {
	Table string
	Rows  int
}

type DatabaseResetCounts struct {
	ActiveDeploymentAdmins int
	BootstrapMarkers       int
	Incidents              int
	Records                int
	UserSessions           int
	RouteIdempotency       int
}

type testRuntimeIdentityResult struct {
	SchemaID      string `json:"schema_id"`
	RuntimeMarker string `json:"runtime_marker"`
	ServerPID     int    `json:"server_pid"`
	TestRoutes    bool   `json:"test_routes_enabled"`
}

func RegisterTestRuntimeIdentityRoute() httpapi.RouteRegistrar {
	return func(mux *http.ServeMux, deps httpapi.DependencySet) error {
		if !httpapi.TestRoutesEnabled(deps.Env) {
			return nil
		}
		guard, err := httpapi.NewTestRouteGuard(deps.Env)
		if err != nil {
			return fmt.Errorf("register test runtime identity route: %w", err)
		}
		mux.HandleFunc("GET /api/v1/test/runtime/identity", func(w http.ResponseWriter, r *http.Request) {
			handleTestRuntimeIdentity(w, r, guard)
		})
		return nil
	}
}

func handleTestRuntimeIdentity(w http.ResponseWriter, r *http.Request, guard httpapi.TestRouteGuard) {
	if !guard.Authorize(w, r) {
		return
	}
	_ = httpapi.WriteSuccess(w, r, http.StatusOK, testRuntimeIdentityResult{
		SchemaID:      testRuntimeIdentitySchemaID,
		RuntimeMarker: testRuntimeMarkerValue,
		ServerPID:     os.Getpid(),
		TestRoutes:    true,
	})
}

// ResetDatabase performs the destructive database portion of a harness reset
// through a caller-supplied Recovery pool. Product server composition never
// supplies this capability; browser and service fixtures invoke it out of
// process. Backend replacement clears all runtime-local state before the
// replacement process reacquires ordinary serving admission.
func ResetDatabase(ctx context.Context, recovery *pgxpool.Pool, restoreBootstrap func(context.Context, pgx.Tx) error) (DatabaseResetResult, error) {
	if recovery == nil || restoreBootstrap == nil {
		return DatabaseResetResult{}, newResetFailure("recovery_reset_dependencies_invalid", errors.New("recovery reset dependencies are required"))
	}
	admission, err := processlease.AcquireRecoveryTarget(
		ctx,
		recovery,
		testRuntimeResetTimeout,
		recoveryLeaseLossDetection,
	)
	if err != nil {
		return DatabaseResetResult{}, newResetFailure("recovery_reset_target_admission_failed", err)
	}
	result, resetErr := resetAdmittedDatabase(admission.Context(), recovery, restoreBootstrap, admission.AssertHeld)
	releaseCtx, cancelRelease := context.WithTimeout(context.Background(), recoveryLeaseLossDetection)
	releaseErr := admission.Release(releaseCtx)
	cancelRelease()
	if resetErr != nil {
		return result, resetErr
	}
	if releaseErr != nil {
		return result, newResetFailure("recovery_reset_target_release_failed", releaseErr)
	}
	return result, nil
}

func resetAdmittedDatabase(
	ctx context.Context,
	recovery *pgxpool.Pool,
	restoreBootstrap func(context.Context, pgx.Tx) error,
	assertAdmission func() error,
) (DatabaseResetResult, error) {
	var result DatabaseResetResult
	if err := assertAdmission(); err != nil {
		return result, newResetFailure("recovery_reset_target_admission_failed", err)
	}
	tx, err := recovery.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return result, newResetFailure("recovery_reset_begin_failed", err)
	}
	defer func() {
		_ = tx.Rollback(context.Background())
	}()
	beforeGooseVersions, err := countTableRows(ctx, tx, "goose_db_version")
	if err != nil {
		return result, newResetFailure("recovery_reset_metadata_inventory_failed", err)
	}
	beforeLineageRows, err := countTableRows(ctx, tx, "schema_migration_lineage")
	if err != nil {
		return result, newResetFailure("recovery_reset_metadata_inventory_failed", err)
	}
	tables, err := listMutablePublicTables(ctx, tx)
	if err != nil {
		return result, newResetFailure("recovery_reset_inventory_failed", err)
	}
	result.TablesReset = append([]string(nil), tables...)
	result.MutableTableCount = len(tables)
	if err := truncateTables(ctx, tx, tables); err != nil {
		return result, newResetFailure("recovery_reset_truncate_failed", err)
	}
	if err := resetTableSequences(ctx, tx, tables); err != nil {
		return result, newResetFailure("recovery_reset_sequence_failed", err)
	}
	if err := restoreBootstrap(ctx, tx); err != nil {
		return result, newResetFailure("recovery_reset_bootstrap_failed", err)
	}
	afterGooseVersions, err := countTableRows(ctx, tx, "goose_db_version")
	if err != nil {
		return result, newResetFailure("recovery_reset_metadata_verification_failed", err)
	}
	afterLineageRows, err := countTableRows(ctx, tx, "schema_migration_lineage")
	if err != nil {
		return result, newResetFailure("recovery_reset_metadata_verification_failed", err)
	}
	counts, err := readDatabaseResetCounts(ctx, tx)
	if err != nil {
		return result, newResetFailure("recovery_reset_state_verification_failed", err)
	}
	tableCounts, err := readMutableTableCounts(ctx, tx, tables)
	if err != nil {
		return result, newResetFailure("recovery_reset_state_verification_failed", err)
	}
	result.TableCounts = tableCounts
	result.MigrationMetadataPreserved = beforeGooseVersions == afterGooseVersions && afterGooseVersions > 0 && beforeLineageRows == afterLineageRows && afterLineageRows == 1
	result.BootstrapAdminRestored = counts.ActiveDeploymentAdmins == 1 && counts.BootstrapMarkers == 1
	result.PostResetCounts = counts
	if !result.MigrationMetadataPreserved {
		return result, newResetFailure("recovery_reset_metadata_verification_failed", errors.New("migration metadata was not preserved"))
	}
	if !result.BootstrapAdminRestored {
		return result, newResetFailure("recovery_reset_bootstrap_verification_failed", errors.New("bootstrap state was not restored"))
	}
	if counts.RouteIdempotency != 0 {
		return result, newResetFailure("recovery_reset_state_verification_failed", errors.New("route idempotency state was not cleared"))
	}
	if err := assertAdmission(); err != nil {
		return result, newResetFailure("recovery_reset_target_admission_lost", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return result, newResetFailure("recovery_reset_commit_failed", err)
	}
	return result, nil
}

type resetDB interface {
	Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
	Query(context.Context, string, ...any) (pgx.Rows, error)
	QueryRow(context.Context, string, ...any) pgx.Row
}

type DatabaseResetFailure struct {
	reason string
	cause  error
}

func (failure *DatabaseResetFailure) Error() string {
	return failure.reason
}

func (failure *DatabaseResetFailure) Unwrap() error {
	return failure.cause
}

func (failure *DatabaseResetFailure) Stage() string {
	if failure == nil {
		return ""
	}
	return failure.reason
}

func (failure *DatabaseResetFailure) SQLState() string {
	if failure == nil {
		return ""
	}
	var postgresError *pgconn.PgError
	if !errors.As(failure.cause, &postgresError) || !validSQLState(postgresError.Code) {
		return ""
	}
	return postgresError.Code
}

func (failure *DatabaseResetFailure) TimedOut() bool {
	return failure != nil && errors.Is(failure.cause, context.DeadlineExceeded)
}

func validSQLState(value string) bool {
	if len(value) != 5 {
		return false
	}
	for _, character := range value {
		if (character < '0' || character > '9') && (character < 'A' || character > 'Z') {
			return false
		}
	}
	return true
}

func newResetFailure(reason string, cause error) error {
	return &DatabaseResetFailure{reason: reason, cause: cause}
}

func NewDatabaseResetFailure(stage string, cause error) *DatabaseResetFailure {
	return &DatabaseResetFailure{reason: stage, cause: cause}
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

func countTableRows(ctx context.Context, db resetDB, table string) (int, error) {
	var count int
	err := db.QueryRow(ctx, "SELECT COUNT(*) FROM "+pgx.Identifier{"public", table}.Sanitize()).Scan(&count)
	return count, err
}

func readMutableTableCounts(ctx context.Context, db resetDB, tables []string) ([]DatabaseResetTableCount, error) {
	counts := make([]DatabaseResetTableCount, 0, len(tables))
	for _, table := range tables {
		count, err := countTableRows(ctx, db, table)
		if err != nil {
			return nil, err
		}
		counts = append(counts, DatabaseResetTableCount{Table: table, Rows: count})
	}
	return counts, nil
}

func readDatabaseResetCounts(ctx context.Context, db resetDB) (DatabaseResetCounts, error) {
	var counts DatabaseResetCounts
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
