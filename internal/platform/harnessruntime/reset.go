package harnessruntime

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/processlease"
	"github.com/JochiRaider/cartulary/internal/platform/recoverystate"
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

// DatabaseResetPlan supplies owner-composed state semantics, never another table registry.
type DatabaseResetPlan struct {
	Catalog          *recoverystate.Catalog
	ResolveRebuild   func(string) (func(context.Context, pgx.Tx) error, error)
	RestoreBootstrap func(context.Context, pgx.Tx) error
}

type DatabaseResetResult struct {
	SettlementFailures         []*DatabaseResetFailure
	TransactionSettlement      string
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
func ResetDatabase(ctx context.Context, recovery *pgxpool.Pool, plan DatabaseResetPlan) (result DatabaseResetResult, resultErr error) {
	result.TransactionSettlement = "not_started"
	if ctx == nil || recovery == nil || plan.Catalog == nil || plan.ResolveRebuild == nil || plan.RestoreBootstrap == nil {
		return result, newResetFailure("recovery_reset_dependencies_invalid", errors.New("recovery reset dependencies are required"))
	}
	totalDeadline := time.Now().Add(testRuntimeResetTimeout)
	if deadline, ok := ctx.Deadline(); ok && deadline.Before(totalDeadline) {
		totalDeadline = deadline
	}
	work, cancelWork := context.WithDeadline(ctx, totalDeadline.Add(-recoveryLeaseLossDetection))
	defer cancelWork()
	admission, err := processlease.AcquireRecoveryTarget(work, recovery, testRuntimeResetTimeout, recoveryLeaseLossDetection)
	if err != nil {
		return result, newResetFailure("recovery_reset_target_admission_failed", err)
	}
	result, resultErr = resetAdmittedDatabase(admission.Context(), recovery, plan, admission.AssertHeld, totalDeadline)
	release, cancelRelease := context.WithDeadline(context.WithoutCancel(ctx), totalDeadline)
	defer cancelRelease()
	if err := admission.Release(release); err != nil {
		failure := NewDatabaseResetFailure("recovery_reset_target_release_failed", err)
		result.SettlementFailures = append(result.SettlementFailures, failure)
		resultErr = errors.Join(resultErr, failure)
	}
	return result, resultErr
}

func resetAdmittedDatabase(ctx context.Context, recovery *pgxpool.Pool, plan DatabaseResetPlan, assertAdmission func() error, deadline time.Time) (result DatabaseResetResult, resultErr error) {
	result.TransactionSettlement = "not_started"
	if err := assertAdmission(); err != nil {
		return result, newResetFailure("recovery_reset_target_admission_failed", err)
	}
	tx, err := recovery.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return result, newResetFailure("recovery_reset_begin_failed", err)
	}
	result.TransactionSettlement = "unknown"
	commitAttempted := false
	defer func() {
		if result.TransactionSettlement == "committed" {
			return
		}
		settle, cancel := context.WithDeadline(context.WithoutCancel(ctx), deadline)
		defer cancel()
		err := tx.Rollback(settle)
		if err == nil && !commitAttempted {
			result.TransactionSettlement = "rolled_back"
		}
		if err != nil && !errors.Is(err, pgx.ErrTxClosed) {
			failure := NewDatabaseResetFailure("recovery_reset_rollback_failed", err)
			result.SettlementFailures = append(result.SettlementFailures, failure)
			resultErr = errors.Join(resultErr, failure)
		}
	}()
	tables, writable, rebuilds, err := inventoryResetTables(ctx, tx, plan)
	if err != nil {
		return result, newResetFailure("recovery_reset_inventory_failed", err)
	}
	result.TablesReset, result.MutableTableCount = tables, len(tables)
	before, err := resetMetadataSnapshot(ctx, tx)
	if err != nil {
		return result, newResetFailure("recovery_reset_metadata_inventory_failed", err)
	}
	identifiers := make([]string, len(tables))
	for i, table := range tables {
		identifiers[i] = pgx.Identifier{"public", table}.Sanitize()
	}
	if len(identifiers) > 0 {
		if _, err := tx.Exec(ctx, "LOCK TABLE "+strings.Join(identifiers, ", ")+" IN ACCESS EXCLUSIVE MODE"); err != nil {
			return result, newResetFailure("recovery_reset_lock_failed", err)
		}
	}
	if _, err := tx.Exec(ctx, `SET LOCAL row_security = off; SET LOCAL session_replication_role = replica`); err != nil {
		return result, newResetFailure("recovery_reset_clear_failed", err)
	}
	for _, table := range writable {
		if _, err := tx.Exec(ctx, "DELETE FROM "+pgx.Identifier{"public", table}.Sanitize()); err != nil {
			return result, newResetFailure("recovery_reset_clear_failed", err)
		}
	}
	for _, rebuild := range rebuilds {
		if err := rebuild(ctx, tx); err != nil {
			return result, newResetFailure("recovery_reset_rebuild_failed", err)
		}
	}
	empty, err := readMutableTableCounts(ctx, tx, tables)
	if err != nil {
		return result, newResetFailure("recovery_reset_state_verification_failed", err)
	}
	for _, table := range empty {
		if table.Rows != 0 {
			return result, newResetFailure("recovery_reset_state_verification_failed", errors.New("mutable state remains before bootstrap"))
		}
	}
	if _, err := tx.Exec(ctx, `SET LOCAL session_replication_role = origin`); err != nil {
		return result, newResetFailure("recovery_reset_bootstrap_failed", err)
	}
	if err := plan.RestoreBootstrap(ctx, tx); err != nil {
		return result, newResetFailure("recovery_reset_bootstrap_failed", err)
	}
	after, err := resetMetadataSnapshot(ctx, tx)
	if err != nil {
		return result, newResetFailure("recovery_reset_metadata_verification_failed", err)
	}
	result.MigrationMetadataPreserved = before == after
	if !result.MigrationMetadataPreserved {
		return result, newResetFailure("recovery_reset_metadata_verification_failed", errors.New("protected metadata changed"))
	}
	result.TableCounts, err = readMutableTableCounts(ctx, tx, tables)
	if err != nil {
		return result, newResetFailure("recovery_reset_state_verification_failed", err)
	}
	for _, table := range result.TableCounts {
		want := 0
		switch table.Table {
		case "users", "deployment_bootstrap_state", "deployment_admin_audit_events", "administrative_audit_projections":
			want = 1
		}
		if table.Rows != want {
			return result, newResetFailure("recovery_reset_bootstrap_verification_failed", errors.New("unexpected bootstrap state"))
		}
	}
	result.PostResetCounts, err = readDatabaseResetCounts(ctx, tx)
	if err != nil {
		return result, newResetFailure("recovery_reset_state_verification_failed", err)
	}
	result.BootstrapAdminRestored = result.PostResetCounts.ActiveDeploymentAdmins == 1 && result.PostResetCounts.BootstrapMarkers == 1
	if !result.BootstrapAdminRestored {
		return result, newResetFailure("recovery_reset_bootstrap_verification_failed", errors.New("bootstrap admin not restored"))
	}
	if err := assertAdmission(); err != nil {
		return result, newResetFailure("recovery_reset_target_admission_lost", err)
	}
	commitAttempted = true
	if err := tx.Commit(ctx); err != nil {
		return result, newResetFailure("recovery_reset_commit_failed", err)
	}
	result.TransactionSettlement = "committed"
	return result, nil
}

func inventoryResetTables(ctx context.Context, tx pgx.Tx, plan DatabaseResetPlan) ([]string, []string, []func(context.Context, pgx.Tx) error, error) {
	rows, err := tx.Query(ctx, `SELECT c.relname, has_table_privilege(current_user,c.oid,'DELETE')
FROM pg_catalog.pg_class c JOIN pg_catalog.pg_namespace n ON n.oid=c.relnamespace
WHERE n.nspname='public' AND c.relkind IN ('r','p','f') ORDER BY c.relname COLLATE "C"`)
	if err != nil {
		return nil, nil, nil, err
	}
	names := []string{}
	deletable := map[string]bool{}
	for rows.Next() {
		var name string
		var allowed bool
		if err := rows.Scan(&name, &allowed); err != nil {
			rows.Close()
			return nil, nil, nil, err
		}
		names = append(names, name)
		deletable[name] = allowed
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return nil, nil, nil, err
	}
	if err := plan.Catalog.ValidateDatabaseTableNames(names); err != nil {
		return nil, nil, nil, err
	}
	tables, writable := []string{}, []string{}
	algorithms := map[string]func(context.Context, pgx.Tx) error{}
	for _, table := range plan.Catalog.Document().Tables {
		if table.StateClass == recoverystate.StateMigration {
			continue
		}
		tables = append(tables, table.TableName)
		if deletable[table.TableName] {
			writable = append(writable, table.TableName)
			continue
		}
		if table.StateClass != recoverystate.StateDerived || table.RestoreAction != recoverystate.RebuildState || table.AlgorithmID == nil {
			return nil, nil, nil, errors.New("mutable table lacks clearing authority")
		}
		rebuild, err := plan.ResolveRebuild(*table.AlgorithmID)
		if err != nil || rebuild == nil {
			return nil, nil, nil, errors.Join(errors.New("unsupported derived clearing algorithm"), err)
		}
		algorithms[*table.AlgorithmID] = rebuild
	}
	sort.Strings(tables)
	sort.Strings(writable)
	ids := []string{}
	for id := range algorithms {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	rebuilds := []func(context.Context, pgx.Tx) error{}
	for _, id := range ids {
		rebuilds = append(rebuilds, algorithms[id])
	}
	return tables, writable, rebuilds, nil
}

func resetMetadataSnapshot(ctx context.Context, db resetDB) (string, error) {
	var snapshot string
	err := db.QueryRow(ctx, `SELECT jsonb_build_array(
 (SELECT COALESCE(jsonb_agg(to_jsonb(t) ORDER BY to_jsonb(t)::text COLLATE "C"),'[]'::jsonb) FROM public.goose_db_version t),
 (SELECT COALESCE(jsonb_agg(to_jsonb(t) ORDER BY to_jsonb(t)::text COLLATE "C"),'[]'::jsonb) FROM public.schema_migration_lineage t),
 (SELECT COALESCE(jsonb_agg(to_jsonb(t) ORDER BY to_jsonb(t)::text COLLATE "C"),'[]'::jsonb) FROM public.application_crypto_format t))::text`).Scan(&snapshot)
	return snapshot, err
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
