package main

import (
	"bufio"
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/JochiRaider/cartulary/internal/app/configassembly"
	"github.com/JochiRaider/cartulary/internal/app/extensionassembly"
	"github.com/JochiRaider/cartulary/internal/app/projectionassembly"
	"github.com/JochiRaider/cartulary/internal/app/recoveryassembly"
	"github.com/JochiRaider/cartulary/internal/app/server"
	"github.com/JochiRaider/cartulary/internal/modules/evidence"
	"github.com/JochiRaider/cartulary/internal/modules/recovery"
	"github.com/JochiRaider/cartulary/internal/modules/revisions/conflicts"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/objectstore"
	"github.com/JochiRaider/cartulary/internal/testutil/fixtures"
	"github.com/JochiRaider/cartulary/internal/testutil/postgrescatalog"
	"github.com/JochiRaider/cartulary/internal/testutil/postgrescleanup"
)

const readySchemaID = "cartulary.restore.browser_restore_target.v1"
const restoreBrowserRecoveryMasterKey = "MDEyMzQ1Njc4OWFiY2RlZjAxMjM0NTY3ODlhYmNkZWY="
const restoreBrowserCatalogAdmissionLimit = 15 * time.Second

type readyPayload struct {
	SchemaID            string                            `json:"schema_id"`
	Origin              string                            `json:"origin"`
	BackupSetID         string                            `json:"backup_set_id"`
	ConsistencyPointAt  time.Time                         `json:"consistency_point_at"`
	RestoredIncidentIDs []string                          `json:"restored_incident_ids"`
	IncidentID          string                            `json:"incident_id"`
	UserEmail           string                            `json:"user_email"`
	UserPassword        string                            `json:"user_password"`
	TimelineSummary     string                            `json:"timeline_summary"`
	ConsistencyReport   recovery.RestoreConsistencyReport `json:"consistency_report"`
	TargetDatabase      string                            `json:"target_database"`
}

type seededSource struct {
	IncidentID      string
	UserEmail       string
	UserPassword    string
	TimelineSummary string
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "restore browser restore target: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	runtimeRoot := strings.TrimSpace(flagValue("--runtime-root"))
	if runtimeRoot == "" {
		return errors.New("--runtime-root is required")
	}
	if err := os.MkdirAll(runtimeRoot, 0o700); err != nil {
		return fmt.Errorf("create runtime root: %w", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	sourceEnv, err := sourceEnvironment(runtimeRoot)
	if err != nil {
		return err
	}
	baseDSN, err := sourcePostgresDSN(runtimeRoot, sourceEnv)
	if err != nil {
		return err
	}
	templateDatabase := strings.TrimSpace(sourceEnv["CARTULARY_PGTEST_TEMPLATE_DB"])
	if templateDatabase == "" {
		return errors.New("browser Recovery fixture requires the suite template database")
	}

	sourceRoot := filepath.Join(runtimeRoot, "restore-browser-restore-source")
	if err := os.RemoveAll(sourceRoot); err != nil {
		return fmt.Errorf("reset source root: %w", err)
	}
	if err := os.MkdirAll(sourceRoot, 0o700); err != nil {
		return fmt.Errorf("create source root: %w", err)
	}
	sourceTokenEnv, err := prepareConflictTokenHarness(sourceRoot)
	if err != nil {
		return err
	}
	sourceName := "cartulary_restore_browser_source_" + safeSuffix(time.Now().UTC().Format("20060102150405.000000000"))
	sourceDSN, err := createFromTemplateDB(ctx, baseDSN, sourceName, templateDatabase)
	if err != nil {
		return err
	}
	defer func() {
		dropCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = dropDatabase(dropCtx, baseDSN, sourceName)
	}()
	sourcePool, err := pgxpool.New(ctx, sourceDSN)
	if err != nil {
		return fmt.Errorf("open source postgres: %w", err)
	}
	defer sourcePool.Close()

	sourceObjectStore, err := objectstore.NewFilesystemStore(filepath.Join(sourceRoot, "object-storage"))
	if err != nil {
		return fmt.Errorf("open source object store: %w", err)
	}
	defer sourceObjectStore.Close()

	const sourceAdminEmail = "restore-browser-admin@example.test"
	const sourceAdminPassword = "RestoreBrowserAdmin1!"
	if err := seedDeploymentAdmin(ctx, sourcePool, sourceAdminEmail, sourceAdminPassword); err != nil {
		return err
	}
	sourceConfig, err := targetConfig(sourceRoot, "", sourceTokenEnv)
	if err != nil {
		return fmt.Errorf("load source configuration: %w", err)
	}
	sourceRuntime, err := server.NewRuntime(ctx, sourceConfig, server.Options{
		Postgres:    sourcePool,
		ObjectStore: sourceObjectStore,
		Env:         sourceTokenEnv,
	})
	if err != nil {
		return fmt.Errorf("start source runtime: %w", err)
	}
	defer sourceRuntime.Close()
	sourceServer, sourceOrigin, err := startRuntimeServer(sourceRuntime.HTTPHandler(), sourceRuntime.ActivatePublication)
	if err != nil {
		return fmt.Errorf("start source server: %w", err)
	}
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = sourceServer.Shutdown(shutdownCtx)
	}()
	seed, err := seedSourceDeployment(ctx, sourceOrigin, sourceAdminEmail, sourceAdminPassword)
	if err != nil {
		return err
	}

	backupRoot := filepath.Join(runtimeRoot, "restore-browser-restore-backup")
	if err := os.RemoveAll(backupRoot); err != nil {
		return fmt.Errorf("reset backup root: %w", err)
	}
	backupStorage, err := encryptedBackupStorage(backupRoot)
	if err != nil {
		return fmt.Errorf("open encrypted backup storage: %w", err)
	}
	extensionBackups, err := extensionassembly.GeneratedRecoveryCatalog()
	if err != nil {
		return fmt.Errorf("construct extension recovery catalog: %w", err)
	}

	sourceStore := recovery.NewStore(sourcePool)
	now := time.Now().UTC()
	postgresArtifact, err := recovery.CapturePostgresSnapshotArtifact(ctx, sourcePool)
	if err != nil {
		return fmt.Errorf("capture source postgres artifact: %w", err)
	}
	blobIndex, err := recovery.AvailableBlobObjectIDsByStorageRef(ctx, evidence.NewRecoveryProvider(sourcePool))
	if err != nil {
		return fmt.Errorf("index source blob storage refs: %w", err)
	}
	backupSetID := uuid.New()
	objectArtifacts, err := recovery.CaptureSeaweedFSS3ObjectStoreBackupArtifacts(ctx, sourceObjectStore, recovery.ObjectStoreBackupCaptureParams{
		BackupSetID:               backupSetID,
		ConsistencyPointAt:        now,
		Bucket:                    "restore-browser-restore-source",
		BlobObjectIDsByStorageRef: blobIndex,
	})
	if err != nil {
		return fmt.Errorf("capture source object artifact: %w", err)
	}
	backupSet, err := recovery.NewCaptureService(sourceStore, backupStorage, extensionBackups).CaptureBackupSet(ctx, recovery.CaptureBackupSetParams{
		BackupSetID:                       backupSetID,
		ConsistencyPointAt:                now,
		CreatedAt:                         now,
		RetainedUntil:                     now.Add(recovery.MinimumRetentionDuration),
		PostgresArtifact:                  recovery.BackupArtifact{Body: postgresArtifact, ContentType: "application/json"},
		ObjectStoreArtifact:               recovery.BackupArtifact{Body: objectArtifacts.SnapshotBody, ContentType: "application/json"},
		ObjectStoreBackupManifestArtifact: recovery.BackupArtifact{Body: objectArtifacts.ManifestBody, ContentType: "application/json"},
		ObjectStoreBackupSummaryArtifact:  recovery.BackupArtifact{Body: objectArtifacts.SummaryBody, ContentType: "application/json"},
	})
	if err != nil {
		return fmt.Errorf("capture retained backup set: %w", err)
	}

	targetRoot := filepath.Join(runtimeRoot, "restore-browser-restore-target")
	if err := os.RemoveAll(targetRoot); err != nil {
		return fmt.Errorf("reset target root: %w", err)
	}
	if err := os.MkdirAll(targetRoot, 0o700); err != nil {
		return fmt.Errorf("create target root: %w", err)
	}
	targetTokenEnv, err := prepareConflictTokenHarness(targetRoot)
	if err != nil {
		return err
	}
	targetName := "cartulary_restore_browser_" + safeSuffix(time.Now().UTC().Format("20060102150405.000000000"))
	targetDSN, err := createFromTemplateDB(ctx, baseDSN, targetName, templateDatabase)
	if err != nil {
		return err
	}
	defer func() {
		dropCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = dropDatabase(dropCtx, baseDSN, targetName)
	}()

	targetPool, err := pgxpool.New(ctx, targetDSN)
	if err != nil {
		return fmt.Errorf("open target postgres: %w", err)
	}
	targetObjectStore, err := objectstore.NewFilesystemStore(filepath.Join(targetRoot, "object-storage"))
	if err != nil {
		targetPool.Close()
		return fmt.Errorf("open target object store: %w", err)
	}

	projectionRebuilder, workbookProbe, err := projectionassembly.NewRecoveryServices(targetPool)
	if err != nil {
		targetPool.Close()
		_ = targetObjectStore.Close()
		return fmt.Errorf("compose restore projection services: %w", err)
	}
	result, err := recovery.NewRestoreRunner(sourceStore, backupStorage, extensionBackups).RestoreLatestSuccessfulRetained(ctx, recovery.RestoreTarget{
		Postgres:        targetPool,
		ObjectStore:     targetObjectStore,
		EvidenceObjects: evidence.NewRecoveryProvider(targetPool),
		Projections:     projectionRebuilder,
	}, now.Add(time.Second))
	if err != nil {
		targetPool.Close()
		_ = targetObjectStore.Close()
		return fmt.Errorf("restore target: %w", err)
	}
	if result.BackupSet.BackupSetID != backupSet.BackupSetID {
		targetPool.Close()
		_ = targetObjectStore.Close()
		return fmt.Errorf("restored backup_set_id %s, want %s", result.BackupSet.BackupSetID, backupSet.BackupSetID)
	}
	result.SelectedIncidentID = &seed.IncidentID
	if err := (recovery.RestoreVerificationWorkbookProbe{Executor: workbookProbe}).ProbeRestoredBackup(ctx, &result); err != nil {
		targetPool.Close()
		_ = targetObjectStore.Close()
		return fmt.Errorf("probe restored workbook query: %w", err)
	}

	cfg, err := targetConfig(targetRoot, "", targetTokenEnv)
	if err != nil {
		targetPool.Close()
		_ = targetObjectStore.Close()
		return fmt.Errorf("load target configuration: %w", err)
	}
	sourceClaims := sourceConfig.RequestedClaims().ProfileIDs()
	targetClaims := cfg.RequestedClaims().ProfileIDs()
	if !slices.Equal(sourceClaims, targetClaims) {
		targetPool.Close()
		_ = targetObjectStore.Close()
		return fmt.Errorf("restore source and target claim postures differ: source=%v target=%v", sourceClaims, targetClaims)
	}
	runtime, err := server.NewRuntime(ctx, cfg, server.Options{
		Postgres:    targetPool,
		ObjectStore: targetObjectStore,
		Env:         targetTokenEnv,
	})
	if err != nil {
		targetPool.Close()
		_ = targetObjectStore.Close()
		return fmt.Errorf("start target runtime: %w", err)
	}
	defer runtime.Close()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return fmt.Errorf("listen target server: %w", err)
	}
	if err := runtime.ActivatePublication(); err != nil {
		_ = listener.Close()
		return fmt.Errorf("activate target publication: %w", err)
	}
	origin := "http://" + listener.Addr().String()
	server := &http.Server{
		Handler:           runtime.HTTPHandler(),
		ReadHeaderTimeout: 5 * time.Second,
	}
	serverErr := make(chan error, 1)
	go func() {
		err := server.Serve(listener)
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
			return
		}
		serverErr <- nil
	}()

	incidentIDs, err := restoredIncidentIDs(ctx, targetPool)
	if err != nil {
		_ = server.Shutdown(context.Background())
		return fmt.Errorf("list restored incidents: %w", err)
	}
	if err := json.NewEncoder(os.Stdout).Encode(readyPayload{
		SchemaID:            readySchemaID,
		Origin:              origin,
		BackupSetID:         result.BackupSet.BackupSetID.String(),
		ConsistencyPointAt:  result.BackupSet.ConsistencyPointAt,
		RestoredIncidentIDs: incidentIDs,
		IncidentID:          seed.IncidentID,
		UserEmail:           seed.UserEmail,
		UserPassword:        seed.UserPassword,
		TimelineSummary:     seed.TimelineSummary,
		ConsistencyReport:   result.ConsistencyReport,
		TargetDatabase:      targetName,
	}); err != nil {
		_ = server.Shutdown(context.Background())
		return fmt.Errorf("write ready payload: %w", err)
	}

	stdinDone := make(chan struct{})
	go func() {
		_, _ = bufio.NewReader(os.Stdin).ReadBytes('\n')
		close(stdinDone)
	}()

	select {
	case <-ctx.Done():
	case <-stdinDone:
	case err := <-serverErr:
		if err != nil {
			return err
		}
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return server.Shutdown(shutdownCtx)
}

func startRuntimeServer(handler http.Handler, activatePublication func() error) (*http.Server, string, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, "", err
	}
	if activatePublication == nil {
		_ = listener.Close()
		return nil, "", errors.New("extension_publication_failed")
	}
	if err := activatePublication(); err != nil {
		_ = listener.Close()
		return nil, "", fmt.Errorf("activate publication: %w", err)
	}
	server := &http.Server{
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
	}
	go func() {
		err := server.Serve(listener)
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			fmt.Fprintf(os.Stderr, "restore browser source server: %v\n", err)
		}
	}()
	return server, "http://" + listener.Addr().String(), nil
}

func encryptedBackupStorage(root string) (recovery.BackupStorage, error) {
	rawStorage, err := recoveryassembly.NewFilesystemStorage(root)
	if err != nil {
		return nil, err
	}
	key, err := recovery.ParseRecoveryEncryptionKey(restoreBrowserRecoveryMasterKey)
	if err != nil {
		return nil, err
	}
	return recovery.NewEncryptedBackupStorage(rawStorage, key)
}

func seedDeploymentAdmin(ctx context.Context, pool *pgxpool.Pool, email string, password string) error {
	hash, err := authn.HashPassword(password)
	if err != nil {
		return fmt.Errorf("hash source admin password: %w", err)
	}
	if _, err := pool.Exec(ctx, `
INSERT INTO users (email, display_name, password_hash, mfa_required, is_active, is_deployment_admin)
VALUES ($1, $2, $3, false, true, true)
`, email, "Recovery Browser Admin", hash); err != nil {
		return fmt.Errorf("seed source deployment admin: %w", err)
	}
	return nil
}

func seedSourceDeployment(ctx context.Context, origin string, adminEmail string, adminPassword string) (seededSource, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return seededSource{}, fmt.Errorf("create source cookie jar: %w", err)
	}
	client := &http.Client{Jar: jar, Timeout: 15 * time.Second}
	csrf, err := loginSourceAdmin(ctx, client, origin, adminEmail, adminPassword)
	if err != nil {
		return seededSource{}, err
	}

	suffix := safeSuffix(time.Now().UTC().Format("20060102150405.000000000"))
	summary := "Recovery restored timeline " + suffix
	incident, err := doSourceJSON(ctx, client, http.MethodPost, origin+"/api/v1/incidents", csrf, map[string]any{
		"client_txn_id": "txn-restore-browser-incident-" + suffix,
		"incident_key":  "IR-RESTORE-BROWSER-" + suffix,
		"title":         "Recovery restored workbook evidence",
	})
	if err != nil {
		return seededSource{}, fmt.Errorf("create source incident: %w", err)
	}
	incidentID, err := stringAt(incident, "data", "incident_id")
	if err != nil {
		return seededSource{}, err
	}

	userEmail := "restore-browser-restored-" + suffix + "@example.test"
	userPassword := "RestoreBrowserUser1!"
	if _, err := doSourceJSON(ctx, client, http.MethodPost, origin+"/api/v1/users", csrf, map[string]any{
		"client_txn_id":       "txn-restore-browser-user-" + suffix,
		"auth_kind":           "local",
		"email":               userEmail,
		"display_name":        "Recovery Restored User",
		"initial_password":    userPassword,
		"mfa_required":        false,
		"is_deployment_admin": false,
	}); err != nil {
		return seededSource{}, fmt.Errorf("create source restored user: %w", err)
	}
	if _, err := doSourceJSON(ctx, client, http.MethodPost, origin+"/api/v1/incidents/"+url.PathEscape(incidentID)+"/memberships", csrf, map[string]any{
		"client_txn_id": "txn-restore-browser-membership-" + suffix,
		"email":         userEmail,
		"role":          "editor",
	}); err != nil {
		return seededSource{}, fmt.Errorf("create source membership: %w", err)
	}
	if _, err := doSourceJSON(ctx, client, http.MethodPost, origin+"/api/v1/incidents/"+url.PathEscape(incidentID)+"/views/cartulary.view.timeline.v2/rows", csrf, map[string]any{
		"client_txn_id":                   "txn-restore-browser-row-" + suffix,
		"timeline.activity_synopsis_text": summary,
	}); err != nil {
		return seededSource{}, fmt.Errorf("create source timeline row: %w", err)
	}

	return seededSource{
		IncidentID:      incidentID,
		UserEmail:       userEmail,
		UserPassword:    userPassword,
		TimelineSummary: summary,
	}, nil
}

func loginSourceAdmin(ctx context.Context, client *http.Client, origin string, email string, password string) (string, error) {
	payload, err := doSourceJSON(ctx, client, http.MethodPost, origin+"/api/v1/auth/login", "", map[string]any{
		"username": email,
		"password": password,
	})
	if err != nil {
		return "", fmt.Errorf("login source admin: %w", err)
	}
	if _, err := stringAt(payload, "data", "session_expires_at"); err != nil {
		return "", fmt.Errorf("login source admin returned unexpected payload: %w", err)
	}
	parsed, err := url.Parse(origin)
	if err != nil {
		return "", err
	}
	for _, cookie := range client.Jar.Cookies(parsed) {
		if cookie.Name == "cartulary_csrf" && strings.TrimSpace(cookie.Value) != "" {
			return cookie.Value, nil
		}
	}
	return "", errors.New("login source admin did not set csrf cookie")
}

func doSourceJSON(ctx context.Context, client *http.Client, method string, endpoint string, csrf string, body any) (map[string]any, error) {
	encoded, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, method, endpoint, bytes.NewReader(encoded))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if csrf != "" {
		req.Header.Set(authn.CSRFHeaderName, csrf)
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	raw, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return nil, readErr
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("%s %s returned %d: %s", method, endpoint, resp.StatusCode, strings.TrimSpace(string(raw)))
	}
	var decoded map[string]any
	if err := json.Unmarshal(raw, &decoded); err != nil {
		return nil, fmt.Errorf("decode %s %s response: %w", method, endpoint, err)
	}
	return decoded, nil
}

func stringAt(value map[string]any, path ...string) (string, error) {
	var current any = value
	for _, key := range path {
		object, ok := current.(map[string]any)
		if !ok {
			return "", fmt.Errorf("response path %s is not an object", strings.Join(path, "."))
		}
		current = object[key]
	}
	text, ok := current.(string)
	if !ok || strings.TrimSpace(text) == "" {
		return "", fmt.Errorf("response path %s is not a non-empty string", strings.Join(path, "."))
	}
	return text, nil
}

func flagValue(name string) string {
	for index, arg := range os.Args[1:] {
		if arg == name && index+2 <= len(os.Args[1:]) {
			return os.Args[index+2]
		}
		prefix := name + "="
		if strings.HasPrefix(arg, prefix) {
			return strings.TrimPrefix(arg, prefix)
		}
	}
	return ""
}

func sourceEnvironment(runtimeRoot string) (map[string]string, error) {
	env := map[string]string{
		"CARTULARY_S3_OBJECT_PRIMARY_ENDPOINT":             "localhost:9000",
		"CARTULARY_S3_OBJECT_PRIMARY_ACCESS_KEY_ID":        "cartulary-local",
		"CARTULARY_S3_OBJECT_PRIMARY_SECRET_ACCESS_KEY":    "cartulary-local-secret",
		"CARTULARY_S3_OBJECT_PRIMARY_SECURE":               "false",
		"CARTULARY_S3_OBJECT_PRIMARY_BUCKET":               "cartulary",
		"CARTULARY_POSTGRES_POSTGRES_PRIMARY_RECOVERY_DSN": "",
		"CARTULARY_POSTGRES_POSTGRES_PRIMARY_SERVICE_KIND": "postgres",
	}
	for key := range env {
		if value := strings.TrimSpace(os.Getenv(key)); value != "" {
			env[key] = value
		}
	}
	if value := strings.TrimSpace(os.Getenv("CARTULARY_WEB_E2E_DB")); value != "" {
		env["CARTULARY_WEB_E2E_DB"] = value
	}
	if value := strings.TrimSpace(os.Getenv("CARTULARY_PGTEST_TEMPLATE_DB")); value != "" {
		env["CARTULARY_PGTEST_TEMPLATE_DB"] = value
	}
	for _, candidate := range []string{
		filepath.Join(runtimeRoot, "test-services-web-e2e.env"),
		filepath.Join(runtimeRoot, "stack.env"),
	} {
		values, err := parseShellEnvFile(candidate)
		if err != nil {
			return nil, err
		}
		for key, value := range values {
			env[key] = value
		}
	}
	return env, nil
}

func parseShellEnvFile(path string) (map[string]string, error) {
	values := make(map[string]string)
	raw, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return values, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read env file %s: %w", path, err)
	}
	for _, line := range strings.Split(string(raw), "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		trimmed = strings.TrimPrefix(trimmed, "export ")
		key, value, ok := strings.Cut(trimmed, "=")
		if !ok {
			continue
		}
		values[strings.TrimSpace(key)] = unquoteShellValue(strings.TrimSpace(value))
	}
	return values, nil
}

func unquoteShellValue(value string) string {
	if len(value) >= 2 && value[0] == '\'' && value[len(value)-1] == '\'' {
		return strings.ReplaceAll(value[1:len(value)-1], `'"'"'`, `'`)
	}
	if len(value) >= 2 && value[0] == '"' && value[len(value)-1] == '"' {
		unquoted, err := strconvUnquote(value)
		if err == nil {
			return unquoted
		}
	}
	return value
}

func strconvUnquote(value string) (string, error) {
	var decoded string
	if err := json.Unmarshal([]byte(value), &decoded); err != nil {
		return "", err
	}
	return decoded, nil
}

func sourcePostgresDSN(runtimeRoot string, env map[string]string) (string, error) {
	if dsn := strings.TrimSpace(env["CARTULARY_POSTGRES_POSTGRES_PRIMARY_RECOVERY_DSN"]); dsn != "" {
		return dsn, nil
	}
	if database := strings.TrimSpace(env["CARTULARY_WEB_E2E_DB"]); database != "" {
		return "postgres://cartulary:cartulary@localhost:5432/" + url.PathEscape(database) + "?sslmode=disable", nil
	}
	stackPath := filepath.Join(runtimeRoot, "stack.json")
	raw, err := os.ReadFile(stackPath)
	if err != nil {
		return "", fmt.Errorf("read browser stack metadata: %w", err)
	}
	var stack struct {
		Database struct {
			LogicalID string `json:"logical_id"`
		} `json:"database"`
	}
	if err := json.Unmarshal(raw, &stack); err != nil {
		return "", fmt.Errorf("decode browser stack metadata: %w", err)
	}
	if strings.TrimSpace(stack.Database.LogicalID) == "" {
		return "", errors.New("browser stack metadata did not include source database")
	}
	return "postgres://cartulary:cartulary@localhost:5432/" + url.PathEscape(stack.Database.LogicalID) + "?sslmode=disable", nil
}

func targetConfig(root string, origin string, runtimeEnv map[string]string) (configassembly.Loaded, error) {
	if origin == "" {
		origin = "http://127.0.0.1"
	}
	env := map[string]string{
		"CARTULARY__APPLICATION__PUBLIC_ORIGIN":                       origin,
		"CARTULARY__ROOTS__DATABASE_STORAGE__PATH":                    filepath.Join(root, "database-storage"),
		"CARTULARY__ROOTS__OBJECT_STORAGE__PATH":                      filepath.Join(root, "object-storage"),
		"CARTULARY__ROOTS__BACKUP_STORAGE__PATH":                      filepath.Join(root, "backup-storage"),
		"CARTULARY__ROOTS__REFERENCE_PACK_STORAGE__PATH":              filepath.Join(root, "reference-pack-storage"),
		"CARTULARY__ROOTS__TEMPORARY_WORK__PATH":                      filepath.Join(root, "temporary-work"),
		"CARTULARY__ROOTS__EXPORT_OUTPUTS__PATH":                      filepath.Join(root, "export-outputs"),
		"CARTULARY__BOOTSTRAP__FIRST_ADMIN_MANIFEST_PATH":             filepath.Join(root, "bootstrap-admin.json"),
		"CARTULARY__REVISIONS__CONFLICT_TOKEN_KEY_RING_MANIFEST_PATH": filepath.Join(root, "revisions-conflict-token-key-ring.json"),
	}
	for key, value := range runtimeEnv {
		env[key] = value
	}
	return configassembly.Load(configassembly.LoadOptions{
		Path: fixtures.Path("config", "valid.toml"),
		Env:  env,
	})
}

func prepareConflictTokenHarness(root string) (map[string]string, error) {
	const secret = "oVmbXT5kH1Q59Lur9tmdNgYUW3L41EGpcjT73_5CgSQ"
	manifest := []byte(`{"schema_id":"cartulary.revisions_conflict_token_key_ring.v1","algorithm":"aes_256_gcm_v1","keys":[{"conflict_token_key_id":"restore-browser-fixture","state":"active","secret_ref":{"kind":"env","name":"revisions-conflict-token-fixture-v1"}}]}`)
	if err := os.WriteFile(filepath.Join(root, "revisions-conflict-token-key-ring.json"), manifest, 0o600); err != nil {
		return nil, fmt.Errorf("write restore-browser Revisions key ring: %w", err)
	}
	return map[string]string{
		conflicts.ConflictTokenFixtureRuntimeEnvName:           conflicts.ConflictTokenFixtureRuntimeMarker,
		"CARTULARY_SECRET_REVISIONS_CONFLICT_TOKEN_FIXTURE_V1": secret,
	}, nil
}

func createFromTemplateDB(ctx context.Context, baseDSN string, databaseName string, templateDatabase string) (string, error) {
	adminDSN, err := databaseDSN(baseDSN, "postgres")
	if err != nil {
		return "", err
	}
	dsn, err := databaseDSN(baseDSN, databaseName)
	if err != nil {
		return "", err
	}
	if err := createDatabase(ctx, adminDSN, databaseName, templateDatabase); err != nil {
		return "", err
	}
	return dsn, nil
}

func createDatabase(ctx context.Context, adminDSN string, name string, templateDatabase string) error {
	quotedTemplate, err := pgIdentifier(templateDatabase)
	if err != nil {
		return err
	}
	if err := postgrescleanup.ForceDropDatabase(ctx, adminDSN, name); err != nil {
		return fmt.Errorf("drop existing target db: %w", err)
	}
	quoted, err := pgIdentifier(name)
	if err != nil {
		return err
	}
	return postgrescatalog.WithMutation(
		ctx,
		adminDSN,
		name,
		restoreBrowserCatalogAdmissionLimit,
		func(admin *sql.DB) error {
			if _, err := admin.ExecContext(ctx, `CREATE DATABASE `+quoted+` TEMPLATE `+quotedTemplate); err != nil {
				return fmt.Errorf("create target db: %w", err)
			}
			return nil
		},
	)
}

func dropDatabase(ctx context.Context, sourceDSN string, name string) error {
	adminDSN, err := databaseDSN(sourceDSN, "postgres")
	if err != nil {
		return err
	}
	return postgrescleanup.ForceDropDatabase(ctx, adminDSN, name)
}

func databaseDSN(rawDSN string, database string) (string, error) {
	parsed, err := url.Parse(rawDSN)
	if err != nil {
		return "", fmt.Errorf("parse postgres dsn: %w", err)
	}
	parsed.Path = "/" + database
	return parsed.String(), nil
}

var identifierPattern = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)

func pgIdentifier(value string) (string, error) {
	if !identifierPattern.MatchString(value) {
		return "", fmt.Errorf("unsafe postgres identifier %q", value)
	}
	return `"` + value + `"`, nil
}

func safeSuffix(value string) string {
	replacer := strings.NewReplacer(".", "_", "-", "_", ":", "_")
	value = replacer.Replace(value)
	value = regexp.MustCompile(`[^a-zA-Z0-9_]`).ReplaceAllString(value, "_")
	return strings.Trim(value, "_")
}

func restoredIncidentIDs(ctx context.Context, db *pgxpool.Pool) ([]string, error) {
	rows, err := db.Query(ctx, `SELECT id::text FROM incidents ORDER BY id::text ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}
