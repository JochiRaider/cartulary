package testruntime

import (
	"context"
	"crypto/subtle"
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

	"github.com/JochiRaider/cartulary/internal/platform/bootstrap"
	"github.com/JochiRaider/cartulary/internal/platform/config"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/objectstore"
	"github.com/JochiRaider/cartulary/internal/testutil/harnessredact"
)

const (
	testRuntimeResetSchemaID = "cartulary.test.runtime_reset.v1"
	testRoutesEnabledEnv     = "CARTULARY_ENABLE_TEST_ROUTES"
	testRouteTokenEnv        = "CARTULARY_TEST_ROUTE_TOKEN"
	testRuntimeMarkerEnv     = "CARTULARY_TEST_RUNTIME_MARKER"
	testRuntimeMarkerValue   = "harness-owned"
	testRouteTokenHeader     = "X-Cartulary-Test-Route-Token"
	testRuntimeResetTimeout  = 30 * time.Second
)

type testRuntimeResetService struct {
	cfg         config.Config
	env         map[string]string
	postgres    *pgxpool.Pool
	objectStore objectstore.Store
	token       string
	resetMu     sync.Mutex
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

func RegisterTestRuntimeResetRoute() httpapi.RouteRegistrar {
	return func(mux *http.ServeMux, deps httpapi.DependencySet) error {
		if lookupTestRuntimeEnv(deps.Env, testRoutesEnabledEnv) != "1" {
			return nil
		}
		if lookupTestRuntimeEnv(deps.Env, testRuntimeMarkerEnv) != testRuntimeMarkerValue {
			return fmt.Errorf("register test runtime reset route: %s must be %q", testRuntimeMarkerEnv, testRuntimeMarkerValue)
		}
		token := lookupTestRuntimeEnv(deps.Env, testRouteTokenEnv)
		if !validTestRouteToken(token) {
			return fmt.Errorf("register test runtime reset route: %s must be a harness-generated token with at least 128 bits of entropy", testRouteTokenEnv)
		}
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
			token:       token,
		}
		mux.HandleFunc("POST /api/v1/test/runtime/reset", service.handleReset)
		return nil
	}
}

func (s *testRuntimeResetService) handleReset(w http.ResponseWriter, r *http.Request) {
	if !s.authorized(r) {
		_ = httpapi.WriteError(w, r, http.StatusForbidden, "test_route_forbidden", "test route token is required", map[string]any{})
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

	ctx, cancel := context.WithTimeout(r.Context(), testRuntimeResetTimeout)
	defer cancel()

	tx, err := s.postgres.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		writeTestRuntimeResetError(w, r, "begin reset transaction", err, false, nil)
		return
	}
	defer func() {
		_ = tx.Rollback(context.Background())
	}()

	beforeGooseVersions, err := countTableRows(ctx, tx, "goose_db_version")
	if err != nil {
		writeTestRuntimeResetError(w, r, "count migration metadata before reset", err, false, nil)
		return
	}

	tables, err := listMutablePublicTables(ctx, tx)
	if err != nil {
		writeTestRuntimeResetError(w, r, "list mutable tables", err, false, nil)
		return
	}
	if err := truncateTables(ctx, tx, tables); err != nil {
		writeTestRuntimeResetError(w, r, "truncate mutable tables", err, false, map[string]any{
			"tables_reset": tables,
		})
		return
	}
	if err := bootstrap.PreflightTx(ctx, s.cfg, tx); err != nil {
		writeTestRuntimeResetError(w, r, "restore bootstrap admin", err, false, map[string]any{
			"tables_reset": tables,
		})
		return
	}
	afterGooseVersions, err := countTableRows(ctx, tx, "goose_db_version")
	if err != nil {
		writeTestRuntimeResetError(w, r, "count migration metadata after reset", err, false, map[string]any{
			"tables_reset": tables,
		})
		return
	}
	counts, err := readPostResetCounts(ctx, tx)
	if err != nil {
		writeTestRuntimeResetError(w, r, "read post-reset counts", err, false, map[string]any{
			"tables_reset": tables,
		})
		return
	}
	if beforeGooseVersions != afterGooseVersions || afterGooseVersions == 0 {
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
	if err := tx.Commit(ctx); err != nil {
		writeTestRuntimeResetError(w, r, "commit reset transaction", err, true, map[string]any{
			"tables_reset": tables,
		})
		return
	}

	objectsRemoved, err := clearConfiguredObjectBucket(ctx, s.cfg, s.env, s.objectStore)
	if err != nil {
		details := map[string]any{
			"tables_reset":         tables,
			"object_count_removed": objectsRemoved,
		}
		if objectsAfter, countErr := countConfiguredObjectBucket(context.Background(), s.cfg, s.env, s.objectStore); countErr == nil {
			details["object_count_after"] = objectsAfter
		}
		writeTestRuntimeResetError(w, r, "clear object store bucket", err, true, details)
		return
	}
	objectsAfter, err := countConfiguredObjectBucket(ctx, s.cfg, s.env, s.objectStore)
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
		MigrationMetadataPreserved: beforeGooseVersions == afterGooseVersions && afterGooseVersions > 0,
		BootstrapAdminRestored:     counts.ActiveDeploymentAdmins == 1 && counts.BootstrapMarkers == 1,
		PartialFailure:             false,
		PostResetCounts:            counts,
	}

	_ = httpapi.WriteSuccess(w, r, http.StatusOK, result)
}

func (s *testRuntimeResetService) authorized(r *http.Request) bool {
	got := r.Header.Get(testRouteTokenHeader)
	if got == "" || len(got) != len(s.token) {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(got), []byte(s.token)) == 1
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

func listMutablePublicTables(ctx context.Context, db resetDB) ([]string, error) {
	rows, err := db.Query(ctx, `
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

func truncateTables(ctx context.Context, db resetDB, tables []string) error {
	if len(tables) == 0 {
		return nil
	}

	identifiers := make([]string, 0, len(tables))
	for _, table := range tables {
		identifiers = append(identifiers, pgx.Identifier{"public", table}.Sanitize())
	}
	_, err := db.Exec(ctx, "TRUNCATE TABLE "+strings.Join(identifiers, ", ")+" RESTART IDENTITY CASCADE")
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

func lookupTestRuntimeEnv(env map[string]string, key string) string {
	if env != nil {
		return env[key]
	}
	return os.Getenv(key)
}

func validTestRouteToken(token string) bool {
	if len(token) < 22 {
		return false
	}
	for _, r := range token {
		if r <= ' ' || r > '~' {
			return false
		}
	}
	return true
}
