// Package recoverybrowsertest owns the disposable backup/restore browser fixture.
package recoverybrowsertest

import (
	"bufio"
	"bytes"
	"context"
	"crypto/tls"
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
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/JochiRaider/cartulary/internal/app/configassembly"
	"github.com/JochiRaider/cartulary/internal/app/extensionassembly"
	"github.com/JochiRaider/cartulary/internal/app/projectionassembly"
	"github.com/JochiRaider/cartulary/internal/app/recoveryassembly"
	"github.com/JochiRaider/cartulary/internal/app/referenceassembly"
	"github.com/JochiRaider/cartulary/internal/app/server"
	"github.com/JochiRaider/cartulary/internal/modules/evidence"
	"github.com/JochiRaider/cartulary/internal/modules/recovery"
	"github.com/JochiRaider/cartulary/internal/modules/reference_data"
	"github.com/JochiRaider/cartulary/internal/modules/revisions/conflicts"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/cryptography"
	"github.com/JochiRaider/cartulary/internal/platform/objectstore"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/testutil/fixtures"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
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

// Run serves a freshly restored deployment until cancellation or input closure.
// It closes borrowed handles and retires every database it creates before returning.
func Run(ctx context.Context, runtimeRoot string, input io.Reader, output io.Writer, lifecycle *Lifecycle) (retErr error) {
	lifecycle.event("fixture", "startup", "started", 90*time.Second, nil)
	defer func() {
		outcome := "succeeded"
		if retErr != nil {
			outcome = "failed"
		}
		lifecycle.event("fixture", "terminal", outcome, 0, retErr)
	}()
	runtimeRoot = strings.TrimSpace(runtimeRoot)
	if runtimeRoot == "" {
		return errors.New("runtime root is required")
	}
	if err := os.MkdirAll(runtimeRoot, 0o700); err != nil {
		return fmt.Errorf("create runtime root: %w", err)
	}

	// The browser harness publishes this owner-only credential in its private
	// runtime root; suite administration credentials are not inherited by browsers.
	credential, err := os.ReadFile(filepath.Join(runtimeRoot, postgres.FilesystemRecoveryDSNFile))
	if err != nil {
		return fmt.Errorf("read private browser Recovery binding: %w", err)
	}
	baseDSN := strings.TrimSpace(string(credential))
	if _, err := (&pgtest.TestDatabase{DSN: baseDSN}).DSNForPurpose(postgres.PurposeRecovery); err != nil {
		return err
	}
	templateDatabase := strings.TrimSpace(os.Getenv("CARTULARY_PGTEST_TEMPLATE_DB"))
	if templateDatabase == "" {
		return errors.New("browser Recovery fixture requires the suite template database")
	}
	tlsDirectory := filepath.Join(runtimeRoot, "tls")
	serverTLS, err := cryptography.TLSServer("127.0.0.1", filepath.Join(tlsDirectory, "server.crt"), filepath.Join(tlsDirectory, "server.key"))
	if err != nil {
		return fmt.Errorf("admit fixture server identity: %w", err)
	}
	clientTLS, err := cryptography.TLSClient(cryptography.TLSClientOptions{ServerName: "127.0.0.1", RootCertificatePath: filepath.Join(tlsDirectory, "ca.pem")})
	if err != nil {
		return fmt.Errorf("admit fixture server trust: %w", err)
	}

	sourceRoot := filepath.Join(runtimeRoot, "restore-browser-restore-source")
	if err := os.RemoveAll(sourceRoot); err != nil {
		return fmt.Errorf("reset source root: %w", err)
	}
	if err := os.MkdirAll(sourceRoot, 0o700); err != nil {
		return fmt.Errorf("create source root: %w", err)
	}
	sourceTokenEnv, err := prepareConflictTokenHarness(sourceRoot, tlsDirectory)
	if err != nil {
		return err
	}
	sourceName := "cartulary_restore_browser_source_" + safeSuffix(time.Now().UTC().Format("20060102150405.000000000"))
	sourceDSN, err := createFromTemplateDB(ctx, baseDSN, sourceName, templateDatabase)
	if err != nil {
		return err
	}
	lifecycle.event("source.database", "acquisition", "succeeded", 0, nil)
	defer lifecycle.cleanup("source.database", 10*time.Second, func() error {
		dropCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return dropDatabase(dropCtx, baseDSN, sourceName)
	}, &retErr)
	sourceRecovery, err := admitPostgres(ctx, sourceDSN, postgres.PurposeRecovery)
	if err != nil {
		return fmt.Errorf("open source postgres: %w", err)
	}
	lifecycle.event("source.recovery", "acquisition", "succeeded", 0, nil)
	defer lifecycle.cleanup("source.recovery", 0, closeWithoutError(sourceRecovery.Close), &retErr)
	sourcePool := sourceRecovery.Pool()

	sourceObjectStore, err := objectstore.NewFilesystemStore(filepath.Join(sourceRoot, "object-storage"))
	if err != nil {
		return fmt.Errorf("open source object store: %w", err)
	}
	lifecycle.event("source.objects", "acquisition", "succeeded", 0, nil)
	defer lifecycle.cleanup("source.objects", 0, sourceObjectStore.Close, &retErr)

	const sourceAdminEmail = "restore-browser-admin@example.test"
	const sourceAdminPassword = "RestoreBrowserAdmin1!"
	seedDatabase, err := admitPostgres(ctx, sourceDSN, postgres.PurposeMigration)
	if err != nil {
		return err
	}
	err = seedDeploymentAdmin(ctx, seedDatabase.Pool(), sourceAdminEmail, sourceAdminPassword)
	seedDatabase.Close()
	if err != nil {
		return err
	}
	sourceAdmission, err := admitPostgres(ctx, sourceDSN, postgres.PurposeRuntime)
	if err != nil {
		return fmt.Errorf("admit source postgres: %w", err)
	}
	lifecycle.event("source.runtime_pool", "acquisition", "succeeded", 0, nil)
	defer lifecycle.cleanup("source.runtime_pool", 0, closeWithoutError(sourceAdmission.Close), &retErr)
	sourceListener, err := tls.Listen("tcp", "127.0.0.1:0", serverTLS.Clone())
	if err != nil {
		return err
	}
	defer sourceListener.Close()
	sourceOrigin := "https://" + sourceListener.Addr().String()
	sourceConfig, err := targetConfig(sourceRoot, sourceOrigin, sourceTokenEnv)
	if err != nil {
		return fmt.Errorf("load source configuration: %w", err)
	}
	sourceRuntime, err := server.NewRuntime(ctx, sourceConfig, server.Options{
		Postgres:    sourceAdmission,
		ObjectStore: sourceObjectStore,
		Env:         sourceTokenEnv,
	})
	if err != nil {
		return fmt.Errorf("start source runtime: %w", err)
	}
	lifecycle.event("source.runtime", "acquisition", "succeeded", 0, nil)
	defer lifecycle.cleanup("source.runtime", 0, closeWithoutError(sourceRuntime.Close), &retErr)
	sourceServer, sourceServerErr, err := startRuntimeServer(ctx, sourceListener, sourceRuntime.HTTPHandler(), sourceRuntime.ActivatePublication)
	if err != nil {
		return fmt.Errorf("start source server: %w", err)
	}
	lifecycle.event("source.http", "acquisition", "succeeded", 0, nil)
	defer lifecycle.cleanup("source.http", 5*time.Second, sourceServer.Close, &retErr)
	seed, err := seedSourceDeployment(ctx, sourceOrigin, clientTLS, sourceAdminEmail, sourceAdminPassword)
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
	lifecycle.event("backup.storage", "acquisition", "succeeded", 0, nil)
	defer lifecycle.cleanup("backup.storage", 0, func() error { return recovery.CloseBackupStorage(backupStorage) }, &retErr)
	extensionBackups, err := extensionassembly.GeneratedRecoveryCatalog()
	if err != nil {
		return fmt.Errorf("construct extension recovery catalog: %w", err)
	}

	sourceStore := recovery.NewStore(sourcePool)
	now := time.Now().UTC()
	stateCatalog, err := recoveryassembly.CurrentRecoveryStateCatalog()
	if err != nil {
		return err
	}
	sourcePacks, err := referenceassembly.NewRecoveryStorage(filepath.Join(sourceRoot, "temporary-work"), filepath.Join(sourceRoot, "reference-pack-storage"), reference_data.DefaultLimits())
	if err != nil {
		return err
	}
	lifecycle.event("source.reference_packs", "acquisition", "succeeded", 0, nil)
	defer lifecycle.cleanup("source.reference_packs", 0, closeWithoutError(sourcePacks.Close), &retErr)
	inventories, err := recoveryassembly.CurrentVNextObjectInventoryCatalog(recoveryassembly.NewVNextObjectSource(sourceObjectStore), sourcePacks, nil)
	if err != nil {
		return err
	}
	streaming, err := recovery.RequireStreamingBackupStorage(backupStorage)
	if err != nil {
		return err
	}
	capture, err := recovery.NewVNextCaptureService(recoveryassembly.NewVNextSnapshotRepository(sourcePool), streaming, stateCatalog, inventories)
	if err != nil {
		return err
	}
	captured, err := capture.Capture(ctx, recovery.VNextCaptureParams{
		BackupSetID: uuid.New(), ConsistencyPointAt: now, CreatedAt: now,
		RetainedUntil: now.Add(31 * 24 * time.Hour),
	})
	if err != nil {
		return fmt.Errorf("capture retained backup set: %w", err)
	}
	backupSet, err := sourceStore.PublishVNextCapturedBackup(ctx, captured)
	if err != nil {
		return err
	}

	targetRoot := filepath.Join(runtimeRoot, "restore-browser-restore-target")
	if err := os.RemoveAll(targetRoot); err != nil {
		return fmt.Errorf("reset target root: %w", err)
	}
	if err := os.MkdirAll(targetRoot, 0o700); err != nil {
		return fmt.Errorf("create target root: %w", err)
	}
	targetTokenEnv, err := prepareConflictTokenHarness(targetRoot, tlsDirectory)
	if err != nil {
		return err
	}
	targetName := "cartulary_restore_browser_" + safeSuffix(time.Now().UTC().Format("20060102150405.000000000"))
	targetDSN, err := createFromTemplateDB(ctx, baseDSN, targetName, templateDatabase)
	if err != nil {
		return err
	}
	lifecycle.event("target.database", "acquisition", "succeeded", 0, nil)
	defer lifecycle.cleanup("target.database", 10*time.Second, func() error {
		dropCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return dropDatabase(dropCtx, baseDSN, targetName)
	}, &retErr)

	targetRecovery, err := admitPostgres(ctx, targetDSN, postgres.PurposeRecovery)
	if err != nil {
		return fmt.Errorf("open target postgres: %w", err)
	}
	lifecycle.event("target.recovery", "acquisition", "succeeded", 0, nil)
	defer lifecycle.cleanup("target.recovery", 0, closeWithoutError(targetRecovery.Close), &retErr)
	targetPool := targetRecovery.Pool()
	targetObjectStore, err := objectstore.NewFilesystemStore(filepath.Join(targetRoot, "object-storage"))
	if err != nil {
		return fmt.Errorf("open target object store: %w", err)
	}
	lifecycle.event("target.objects", "acquisition", "succeeded", 0, nil)
	defer lifecycle.cleanup("target.objects", 0, targetObjectStore.Close, &retErr)

	projectionRebuilder, workbookProbe, err := projectionassembly.NewRecoveryServices(targetPool)
	if err != nil {
		return fmt.Errorf("compose restore projection services: %w", err)
	}
	targetPacks, err := referenceassembly.NewRecoveryStorage(filepath.Join(targetRoot, "temporary-work"), filepath.Join(targetRoot, "reference-pack-storage"), reference_data.DefaultLimits())
	if err != nil {
		return err
	}
	lifecycle.event("target.reference_packs", "acquisition", "succeeded", 0, nil)
	defer lifecycle.cleanup("target.reference_packs", 0, closeWithoutError(targetPacks.Close), &retErr)
	graph, err := recoveryassembly.NewGraphProjectionRestoreParticipant(targetPool)
	if err != nil {
		return err
	}
	result, err := recovery.NewVersionedRestoreRunner(sourceStore, backupStorage, extensionBackups, stateCatalog).RestoreLatestSuccessfulRetained(ctx, recovery.RestoreTarget{
		RestoreOperationID: uuid.New(), TargetGenerationID: uuid.New(),
		Postgres: targetPool, ObjectStore: targetObjectStore, ReferencePacks: targetPacks,
		EvidenceObjects: evidence.NewRecoveryProvider(targetPool),
		GraphProjection: graph, Projections: projectionRebuilder,
	}, now.Add(time.Second))
	if err != nil {
		return fmt.Errorf("restore target: %w", err)
	}
	if result.BackupSet.BackupSetID != backupSet.BackupSetID {
		return fmt.Errorf("restored backup_set_id %s, want %s", result.BackupSet.BackupSetID, backupSet.BackupSetID)
	}
	result.SelectedIncidentID = &seed.IncidentID
	if err := (recovery.RestoreVerificationWorkbookProbe{Executor: workbookProbe}).ProbeRestoredBackup(ctx, &result); err != nil {
		return fmt.Errorf("probe restored workbook query: %w", err)
	}

	listener, err := tls.Listen("tcp", "127.0.0.1:0", serverTLS.Clone())
	if err != nil {
		return err
	}
	defer listener.Close()
	origin := "https://" + listener.Addr().String()
	cfg, err := targetConfig(targetRoot, origin, targetTokenEnv)
	if err != nil {
		return fmt.Errorf("load target configuration: %w", err)
	}
	sourceClaims := sourceConfig.RequestedClaims().ProfileIDs()
	targetClaims := cfg.RequestedClaims().ProfileIDs()
	if !slices.Equal(sourceClaims, targetClaims) {
		return fmt.Errorf("restore source and target claim postures differ: source=%v target=%v", sourceClaims, targetClaims)
	}
	targetAdmission, err := admitPostgres(ctx, targetDSN, postgres.PurposeRuntime)
	if err != nil {
		return fmt.Errorf("admit target postgres: %w", err)
	}
	lifecycle.event("target.runtime_pool", "acquisition", "succeeded", 0, nil)
	defer lifecycle.cleanup("target.runtime_pool", 0, closeWithoutError(targetAdmission.Close), &retErr)
	runtime, err := server.NewRuntime(ctx, cfg, server.Options{
		Postgres:    targetAdmission,
		ObjectStore: targetObjectStore,
		Env:         targetTokenEnv,
	})
	if err != nil {
		return fmt.Errorf("start target runtime: %w", err)
	}
	lifecycle.event("target.runtime", "acquisition", "succeeded", 0, nil)
	defer lifecycle.cleanup("target.runtime", 0, closeWithoutError(runtime.Close), &retErr)

	server, serverErr, err := startRuntimeServer(ctx, listener, runtime.HTTPHandler(), runtime.ActivatePublication)
	if err != nil {
		return err
	}
	lifecycle.event("target.http", "acquisition", "succeeded", 0, nil)
	defer lifecycle.cleanup("target.http", 5*time.Second, server.Close, &retErr)

	incidentIDs, err := restoredIncidentIDs(ctx, targetPool)
	if err != nil {
		return fmt.Errorf("list restored incidents: %w", err)
	}
	if err := json.NewEncoder(output).Encode(readyPayload{
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
		return fmt.Errorf("write ready payload: %w", err)
	}

	lifecycle.event("fixture", "startup", "succeeded", 90*time.Second, nil)
	stdinDone := make(chan struct{})
	go func() {
		_, _ = bufio.NewReader(input).ReadBytes('\n')
		close(stdinDone)
	}()

	select {
	case <-ctx.Done():
	case <-stdinDone:
	case err := <-sourceServerErr:
		return err
	case err := <-serverErr:
		if err != nil {
			return err
		}
	}

	return nil
}

func admitPostgres(ctx context.Context, dsn string, purpose postgres.Purpose) (postgres.AdmittedPool, error) {
	selected, err := (&pgtest.TestDatabase{DSN: dsn}).DSNForPurpose(purpose)
	if err != nil {
		return nil, err
	}
	roles := map[postgres.Purpose]string{postgres.PurposeRuntime: "cartulary_runtime", postgres.PurposeMigration: "cartulary_schema_owner", postgres.PurposeRecovery: "cartulary_recovery"}
	return postgres.Setup(ctx, postgres.Settings{BindingKind: "managed_service", DSN: selected, Purpose: purpose, ExpectedRole: roles[purpose]})
}

type runtimeServer struct {
	http          *http.Server
	cancel        context.CancelFunc
	closeOnce     sync.Once
	active        sync.WaitGroup
	closeErr      error
	connectionsMu sync.Mutex
	connections   map[net.Conn]http.ConnState
	retiring      bool
}

func (server *runtimeServer) connectionStates() string {
	server.connectionsMu.Lock()
	defer server.connectionsMu.Unlock()
	counts := map[http.ConnState]int{}
	for _, state := range server.connections {
		counts[state]++
	}
	return fmt.Sprintf("new=%d active=%d idle=%d", counts[http.StateNew], counts[http.StateActive], counts[http.StateIdle])
}

func (server *runtimeServer) Close() error {
	server.closeOnce.Do(func() {
		// A preconnected browser socket has no HTTP request to drain. Stop
		// admitting that lifetime before Shutdown, whose StateNew grace period
		// can outlast this fixture's shutdown deadline. The ConnState hook also
		// retires connections accepted concurrently with listener shutdown.
		server.connectionsMu.Lock()
		server.retiring = true
		var unused []net.Conn
		for connection, state := range server.connections {
			if state == http.StateNew {
				unused = append(unused, connection)
			}
		}
		server.connectionsMu.Unlock()
		for _, connection := range unused {
			_ = connection.Close()
		}
		// Shutdown does not cancel active handlers or hijacked WebSockets. Retire
		// their request lifetime before draining and releasing borrowed databases.
		server.cancel()
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		before := server.connectionStates()
		shutdownErr := server.http.Shutdown(ctx)
		if shutdownErr != nil {
			server.closeErr = fmt.Errorf("HTTP shutdown (start: %s; deadline: %s): %w", before, server.connectionStates(), shutdownErr)
			server.closeErr = errors.Join(server.closeErr, server.http.Close())
		}
		drained := make(chan struct{})
		go func() { server.active.Wait(); close(drained) }()
		select {
		case <-drained:
		case <-ctx.Done():
			server.closeErr = errors.Join(server.closeErr, fmt.Errorf("drain fixture requests: %w", ctx.Err()))
		}
	})
	return server.closeErr
}

func startRuntimeServer(parent context.Context, listener net.Listener, handler http.Handler, activatePublication func() error) (*runtimeServer, <-chan error, error) {
	if activatePublication == nil {
		return nil, nil, errors.New("extension_publication_failed")
	}
	if err := activatePublication(); err != nil {
		return nil, nil, fmt.Errorf("activate publication: %w", err)
	}
	ctx, cancel := context.WithCancel(parent)
	server := &runtimeServer{cancel: cancel, connections: make(map[net.Conn]http.ConnState)}
	server.http = &http.Server{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			server.active.Add(1)
			defer server.active.Done()
			handler.ServeHTTP(w, r)
		}),
		ConnState: func(connection net.Conn, state http.ConnState) {
			server.connectionsMu.Lock()
			if state == http.StateClosed || state == http.StateHijacked {
				delete(server.connections, connection)
			} else {
				server.connections[connection] = state
			}
			retire := server.retiring && state == http.StateNew
			server.connectionsMu.Unlock()
			if retire {
				_ = connection.Close()
			}
		},
		ReadHeaderTimeout: 5 * time.Second,
		BaseContext:       func(net.Listener) context.Context { return ctx },
	}
	done := make(chan error, 1)
	go func() {
		err := server.http.Serve(listener)
		if errors.Is(err, http.ErrServerClosed) {
			err = nil
		}
		done <- err
	}()
	return server, done, nil
}

func encryptedBackupStorage(root string) (recovery.BackupStorage, error) {
	rawStorage, err := recoveryassembly.NewFilesystemStorage(root)
	if err != nil {
		return nil, err
	}
	key, err := recovery.ParseRecoveryEncryptionKey(restoreBrowserRecoveryMasterKey)
	if err != nil {
		return nil, errors.Join(err, rawStorage.Close())
	}
	storage, err := recovery.NewEncryptedBackupStorage(rawStorage, key)
	if err != nil {
		return nil, errors.Join(err, rawStorage.Close())
	}
	return storage, nil
}

func seedDeploymentAdmin(ctx context.Context, pool *pgxpool.Pool, email string, password string) error {
	hash, err := authn.HashPassword(ctx, password)
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

func seedSourceDeployment(ctx context.Context, origin string, trust *tls.Config, adminEmail string, adminPassword string) (seededSource, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return seededSource{}, fmt.Errorf("create source cookie jar: %w", err)
	}
	transport := &http.Transport{TLSClientConfig: trust}
	defer transport.CloseIdleConnections()
	client := &http.Client{Jar: jar, Transport: transport, Timeout: 15 * time.Second}
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

func targetConfig(root string, origin string, runtimeEnv map[string]string) (configassembly.Loaded, error) {
	env := map[string]string{
		"CARTULARY__APPLICATION__PUBLIC_ORIGIN":                       origin,
		"CARTULARY__ROOTS__DATABASE_STORAGE__PATH":                    filepath.Join(root, "database-storage"),
		"CARTULARY__ROOTS__OBJECT_STORAGE__PATH":                      filepath.Join(root, "object-storage"),
		"CARTULARY__ROOTS__BACKUP_STORAGE__PATH":                      filepath.Join(root, "backup-storage"),
		"CARTULARY__ROOTS__REFERENCE_PACK_STORAGE__PATH":              filepath.Join(root, "reference-pack-storage"),
		"CARTULARY__ROOTS__TEMPORARY_WORK__PATH":                      filepath.Join(root, "temporary-work"),
		"CARTULARY__ROOTS__EXPORT_OUTPUTS__PATH":                      filepath.Join(root, "export-outputs"),
		"CARTULARY__BOOTSTRAP__FIRST_ADMIN_MANIFEST_PATH":             filepath.Join(root, "bootstrap-admin.json"),
		"CARTULARY__REFERENCE_PACKS__TRUST_BOOTSTRAP_PATH":            fixtures.Path("reference-packs", "trust-bootstrap.json"),
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

func prepareConflictTokenHarness(root string, tlsDirectory string) (map[string]string, error) {
	const secret = "oVmbXT5kH1Q59Lur9tmdNgYUW3L41EGpcjT73_5CgSQ"
	manifest := []byte(`{"schema_id":"cartulary.revisions_conflict_token_key_ring.v2","algorithm":"hkdf_sha256_aes_256_gcm_v2","keys":[{"conflict_token_key_id":"restore-browser-fixture","state":"active","secret_ref":{"kind":"env","name":"revisions-conflict-token-fixture-v1"}}]}`)
	if err := os.WriteFile(filepath.Join(root, "revisions-conflict-token-key-ring.json"), manifest, 0o600); err != nil {
		return nil, fmt.Errorf("write restore-browser Revisions key ring: %w", err)
	}
	return map[string]string{
		"CARTULARY__APPLICATION__TLS_CERTIFICATE_PATH":         filepath.Join(tlsDirectory, "server.crt"),
		"CARTULARY__APPLICATION__TLS_PRIVATE_KEY_PATH":         filepath.Join(tlsDirectory, "server.key"),
		authn.AuthMasterKeyEnv:                                 strings.TrimSpace(string(fixtures.MustRead("auth", "master-key.base64"))),
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
	if err := pgtest.ProvisionDatabase(ctx, adminDSN, databaseName); err != nil {
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return "", errors.Join(err, dropDatabase(cleanupCtx, baseDSN, databaseName))
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
