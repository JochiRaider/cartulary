package main

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	dbmigrations "github.com/JochiRaider/cartulary/db/migrations"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
	"github.com/JochiRaider/cartulary/internal/testutil/s3test"
	"github.com/JochiRaider/cartulary/internal/testutil/suiteservices"
	"github.com/JochiRaider/cartulary/internal/testutil/testcontainersx"
)

const (
	postgresStartupTimeout     = 2 * time.Minute
	templateStartupTimeout     = 2 * time.Minute
	minioStartupTimeout        = 5 * time.Minute
	cleanupTimeout             = 2 * time.Minute
	webE2ELeakCheckTimeout     = 5 * time.Second
	signalWaitTimeout          = 15 * time.Second
	webE2ECleanupWorkers       = 4
	webE2ECleanupMaxWorkers    = 16
	staleFixtureMaxCandidates  = 32
	staleFixtureJanitorTimeout = 10 * time.Second

	webE2ECleanupWorkersEnv = "CARTULARY_TEST_SERVICES_WEB_E2E_CLEANUP_WORKERS"

	stagePostgresStart    = "postgres-start"
	stagePostgresTemplate = "postgres-template"
	stageMinIOStart       = "minio-start"
	stageChildStart       = "child-start"
	stageCleanupWebE2E    = "cleanup-web-e2e"
	stageCleanupJanitor   = "cleanup-janitor"
	stageCleanupPostgres  = "cleanup-postgres"
	stageCleanupMinIO     = "cleanup-minio"

	webE2EReclaimStrategyOwnedStack = "owned_stack_termination"

	bucketSetup       = "setup"
	bucketServiceWait = "service_wait"
	bucketMigration   = "migration"
	bucketTeardown    = "teardown"
)

type postgresService struct {
	adminDSN    string
	dsnTemplate string
	host        string
	port        string
	user        string
	close       func(context.Context) error
}

type minioService struct {
	endpoint  string
	accessKey string
	secretKey string
	secure    bool
	close     func(context.Context) error
}

type postgresStartResult struct {
	service postgresService
	start   time.Time
	end     time.Time
	err     error
}

type minioStartResult struct {
	service minioService
	start   time.Time
	end     time.Time
	err     error
}

type serviceCleanupResult struct {
	service string
	stage   string
	label   string
	start   time.Time
	end     time.Time
	err     error
}

type webE2EFixture struct {
	DatabaseName string
	DSN          string
	Bucket       string
	S3Endpoint   string
	S3AccessKey  string
	S3SecretKey  string
	S3Secure     bool
}

type webE2EMetadata struct {
	DatabaseName string `json:"database_name"`
	Bucket       string `json:"bucket"`
	Target       string `json:"target,omitempty"`
}

type childProcess interface {
	Wait() error
	Signal(os.Signal) error
	Kill() error
	PID() int
}

type dependencies struct {
	startPostgres       func(context.Context) (postgresService, error)
	startMinIO          func(context.Context) (minioService, error)
	startChild          func(argv []string, env map[string]string) (childProcess, error)
	createTemplate      func(context.Context, string, string) error
	prepareWebE2E       func(context.Context, map[string]string) (webE2EFixture, error)
	cleanupWebE2EDB     func(context.Context, webE2EMetadata, map[string]string) error
	cleanupWebE2EBucket func(context.Context, webE2EMetadata, map[string]string) error
	detectWebE2ELeaks   func(context.Context, []webE2EMetadata, map[string]string) error
	warmImages          func(context.Context, []string) error
	recordEvent         func(map[string]string, suiteservices.Event)
	refreshSummary      func(map[string]string)
	suiteID             func() (string, error)
	notifySignals       func(chan<- os.Signal, ...os.Signal)
	stopSignals         func(chan<- os.Signal)
}

func main() {
	os.Exit(run(os.Args[1:], envMap(os.Environ()), defaultDependencies()))
}

func run(args []string, env map[string]string, deps dependencies) int {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "usage: testservices run -- <command> [args...] | prepare-web-e2e --env-file <path> --metadata-file <path> | cleanup-web-e2e --metadata-file <path> | warm-images")
		return 2
	}

	switch args[0] {
	case "run":
		return runWrappedCommand(args, env, deps)
	case "prepare-web-e2e":
		return runPrepareWebE2E(args[1:], env, deps)
	case "cleanup-web-e2e":
		return runCleanupWebE2E(args[1:], env, deps)
	case "warm-images":
		return runWarmImages(args[1:], deps)
	default:
		fmt.Fprintf(os.Stderr, "usage: unknown testservices command %q\n", args[0])
		return 2
	}
}

func startPostgresAsync(parent context.Context, deps dependencies) <-chan postgresStartResult {
	resultCh := make(chan postgresStartResult, 1)
	go func() {
		ctx, cancel := context.WithTimeout(parent, postgresStartupTimeout)
		defer cancel()

		result := postgresStartResult{start: time.Now().UTC()}
		result.service, result.err = deps.startPostgres(ctx)
		result.end = time.Now().UTC()
		resultCh <- result
	}()
	return resultCh
}

func startMinIOAsync(parent context.Context, deps dependencies) <-chan minioStartResult {
	resultCh := make(chan minioStartResult, 1)
	go func() {
		ctx, cancel := context.WithTimeout(parent, minioStartupTimeout)
		defer cancel()

		result := minioStartResult{start: time.Now().UTC()}
		result.service, result.err = deps.startMinIO(ctx)
		result.end = time.Now().UTC()
		resultCh <- result
	}()
	return resultCh
}

func runWrappedCommand(args []string, env map[string]string, deps dependencies) int {
	wrapperStart := time.Now().UTC()
	command, usageErr := parseRunCommand(args)
	if usageErr != nil {
		fmt.Fprintln(os.Stderr, usageErr)
		return 2
	}

	if suiteservices.SuiteActive(env) {
		deps.recordEvent(env, suiteservices.Event{Type: suiteservices.EventWrapperPassThrough})
		child, err := deps.startChild(command, env)
		if err != nil {
			recordFailureAndRefresh(deps, env, failureSummary("", stageChildStart, "start child command", err))
			return 1
		}
		return waitForChild(command, child, deps)
	}

	suiteID, err := deps.suiteID()
	if err != nil {
		fmt.Fprintf(os.Stderr, "generate suite id: %v\n", err)
		return 1
	}

	ownedEnv := cloneEnv(env)
	ownedEnv[suiteservices.ActiveEnv] = "1"
	ownedEnv[suiteservices.SuiteIDEnv] = suiteID

	deps.recordEvent(ownedEnv, suiteservices.Event{Type: suiteservices.EventWrapperOwnedStart})
	deps.refreshSummary(ownedEnv)
	recordTimingSpan(deps, ownedEnv, bucketSetup, "test-services wrapper setup", wrapperStart, time.Now().UTC(), "pass")

	var postgresSvc postgresService
	var minioSvc minioService
	childExitCode := 1
	cleanupStatus := "startup_failed"

	startupCtx, cancelStartup := context.WithCancel(context.Background())
	defer cancelStartup()
	postgresResultCh := startPostgresAsync(startupCtx, deps)
	minioResultCh := startMinIOAsync(startupCtx, deps)

	postgresResult := <-postgresResultCh
	postgresSvc = postgresResult.service
	recordTimingSpanStatusAt(deps, ownedEnv, bucketServiceWait, "test-services start postgres", postgresResult.start, postgresResult.end, postgresResult.err)
	if postgresResult.err != nil {
		cancelStartup()
		minioResult := <-minioResultCh
		minioSvc = minioResult.service
		recordTimingSpanStatusAt(deps, ownedEnv, bucketServiceWait, "test-services start minio", minioResult.start, minioResult.end, minioResult.err)
		recordFailureAndRefresh(deps, ownedEnv, failureSummary(suiteservices.ServicePostgres, stagePostgresStart, "start suite postgres", postgresResult.err))
		cleanupOwnedServices(deps, ownedEnv, postgresSvc, minioSvc, "startup_failed", childExitCode)
		return 1
	}
	defer func() {
		cleanupOwnedServices(deps, ownedEnv, postgresSvc, minioSvc, cleanupStatus, childExitCode)
	}()

	templateDB := templateDatabaseName(suiteID)
	deps.recordEvent(ownedEnv, suiteservices.Event{
		Type:    suiteservices.EventServiceStarted,
		Service: suiteservices.ServicePostgres,
		Details: map[string]any{
			"host":              postgresSvc.host,
			"port":              postgresSvc.port,
			"user":              postgresSvc.user,
			"template_database": templateDB,
		},
	})
	deps.refreshSummary(ownedEnv)

	templateCtx, cancelTemplate := context.WithTimeout(context.Background(), templateStartupTimeout)
	templateStart := time.Now().UTC()
	err = deps.createTemplate(templateCtx, postgresSvc.adminDSN, templateDB)
	recordTimingSpanStatus(deps, ownedEnv, bucketMigration, "test-services prepare postgres template database", templateStart, err)
	cancelTemplate()
	if err != nil {
		cancelStartup()
		minioResult := <-minioResultCh
		minioSvc = minioResult.service
		recordTimingSpanStatusAt(deps, ownedEnv, bucketServiceWait, "test-services start minio", minioResult.start, minioResult.end, minioResult.err)
		recordFailureAndRefresh(deps, ownedEnv, failureSummary(suiteservices.ServicePostgres, stagePostgresTemplate, "prepare postgres template database", err))
		return 1
	}
	deps.recordEvent(ownedEnv, suiteservices.Event{
		Type:    suiteservices.EventPostgresDBCreated,
		Name:    templateDB,
		Kind:    "template",
		Details: postgresPreparationDetails(ownedEnv, suiteservices.PostgresPreparationTemplate, ""),
	})
	deps.recordEvent(ownedEnv, suiteservices.Event{
		Type:    suiteservices.EventPostgresDBMigrated,
		Name:    templateDB,
		Kind:    "template",
		Details: postgresPreparationDetails(ownedEnv, suiteservices.PostgresPreparationTemplate, ""),
	})
	deps.refreshSummary(ownedEnv)

	minioResult := <-minioResultCh
	minioSvc = minioResult.service
	recordTimingSpanStatusAt(deps, ownedEnv, bucketServiceWait, "test-services start minio", minioResult.start, minioResult.end, minioResult.err)
	if minioResult.err != nil {
		recordFailureAndRefresh(deps, ownedEnv, failureSummary(suiteservices.ServiceMinIO, stageMinIOStart, "start suite minio", minioResult.err))
		return 1
	}
	deps.recordEvent(ownedEnv, suiteservices.Event{
		Type:    suiteservices.EventServiceStarted,
		Service: suiteservices.ServiceMinIO,
		Details: map[string]any{
			"endpoint": minioSvc.endpoint,
			"secure":   minioSvc.secure,
		},
	})
	deps.refreshSummary(ownedEnv)

	janitorCtx, cancelJanitor := context.WithTimeout(context.Background(), staleFixtureJanitorTimeout)
	janitorStart := time.Now().UTC()
	err = cleanupStaleWebE2EFixtures(janitorCtx, deps, serviceBackedCleanupEnv(ownedEnv, postgresSvc, minioSvc))
	recordTimingSpanStatus(deps, ownedEnv, bucketSetup, "test-services janitor stale browser fixtures", janitorStart, err)
	cancelJanitor()
	if err != nil {
		recordFailureAndRefresh(deps, ownedEnv, failureSummary("", stageCleanupJanitor, "janitor stale browser e2e fixtures", err))
		cleanupStatus = "startup_failed"
		return 1
	}

	childEnv := cloneEnv(ownedEnv)
	childEnv[suiteservices.PGAdminDSNEnv] = postgresSvc.adminDSN
	childEnv[suiteservices.PGDSNTemplateEnv] = postgresSvc.dsnTemplate
	childEnv[suiteservices.PGTemplateDBEnv] = templateDB
	childEnv[suiteservices.S3EndpointEnv] = minioSvc.endpoint
	childEnv[suiteservices.S3AccessKeyEnv] = minioSvc.accessKey
	childEnv[suiteservices.S3SecretKeyEnv] = minioSvc.secretKey
	childEnv[suiteservices.S3SecureEnv] = fmt.Sprintf("%t", minioSvc.secure)

	child, err := deps.startChild(command, childEnv)
	if err != nil {
		recordFailureAndRefresh(deps, ownedEnv, failureSummary("", stageChildStart, "start child command", err))
		cleanupStatus = "child_start_failed"
		return 1
	}

	childExitCode = waitForChild(command, child, deps)
	if childExitCode == 0 {
		cleanupStatus = "succeeded"
		return 0
	}
	cleanupStatus = "failed"
	return childExitCode
}

func runPrepareWebE2E(args []string, env map[string]string, deps dependencies) int {
	envFile, metadataFile, err := parsePrepareWebE2EArgs(args)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 2
	}
	if !suiteservices.SuiteActive(env) {
		fmt.Fprintf(os.Stderr, "prepare-web-e2e requires %s=1\n", suiteservices.ActiveEnv)
		return 1
	}
	if strings.TrimSpace(suiteservices.LookupEnvValue(env, suiteservices.PGTemplateDBEnv)) == "" {
		fmt.Fprintf(os.Stderr, "prepare-web-e2e requires %s to clone the migrated suite template database\n", suiteservices.PGTemplateDBEnv)
		return 1
	}

	ctx, cancel := context.WithTimeout(context.Background(), templateStartupTimeout)
	defer cancel()

	prepareStart := time.Now().UTC()
	fixture, err := deps.prepareWebE2E(ctx, env)
	recordTimingSpanStatus(deps, env, bucketMigration, "test-services prepare browser e2e fixture", prepareStart, err)
	if err != nil {
		fmt.Fprintf(os.Stderr, "prepare browser e2e fixture: %v\n", err)
		return 1
	}

	metadata := webE2EMetadata{
		DatabaseName: fixture.DatabaseName,
		Bucket:       fixture.Bucket,
		Target:       suiteservices.LookupEnvValue(env, suiteservices.TargetEnv),
	}
	if err := writeWebE2EMetadata(metadataFile, metadata); err != nil {
		_ = cleanupWebE2EFixture(context.Background(), deps, env, metadata)
		fmt.Fprintf(os.Stderr, "write browser e2e metadata: %v\n", err)
		return 1
	}
	if err := writeWebE2EEnv(envFile, fixture); err != nil {
		_ = cleanupWebE2EFixture(context.Background(), deps, env, metadata)
		fmt.Fprintf(os.Stderr, "write browser e2e env: %v\n", err)
		return 1
	}
	deps.refreshSummary(env)

	return 0
}

func runCleanupWebE2E(args []string, env map[string]string, deps dependencies) int {
	metadataFile, err := parseCleanupWebE2EArgs(args)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 2
	}
	if !suiteservices.SuiteActive(env) {
		fmt.Fprintf(os.Stderr, "cleanup-web-e2e requires %s=1\n", suiteservices.ActiveEnv)
		return 1
	}

	metadata, err := readWebE2EMetadata(metadataFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "read browser e2e metadata: %v\n", err)
		return 1
	}

	retireStart := time.Now().UTC()
	recordWebE2EFixtureEvent(deps, env, suiteservices.EventWebE2EFixtureRetired, metadata)
	recordTimingSpan(deps, env, bucketTeardown, "test-services retire browser e2e fixture", retireStart, time.Now().UTC(), "pass")
	deps.refreshSummary(env)
	return 0
}

func runWarmImages(args []string, deps dependencies) int {
	if len(args) != 0 {
		fmt.Fprintln(os.Stderr, "usage: testservices warm-images")
		return 2
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
	defer cancel()

	if err := deps.warmImages(ctx, serviceImages()); err != nil {
		fmt.Fprintf(os.Stderr, "warm service images: %v\n", err)
		return 1
	}
	return 0
}

func parseRunCommand(args []string) ([]string, error) {
	if len(args) < 3 || args[0] != "run" || args[1] != "--" {
		return nil, errors.New("usage: testservices run -- <command> [args...]")
	}
	command := slices.Clone(args[2:])
	if len(command) == 0 || strings.TrimSpace(command[0]) == "" {
		return nil, errors.New("usage: testservices run -- <command> [args...]")
	}
	return command, nil
}

func parsePrepareWebE2EArgs(args []string) (string, string, error) {
	values, err := parseFlagPairs(args, map[string]struct{}{
		"--env-file":      {},
		"--metadata-file": {},
	})
	if err != nil {
		return "", "", err
	}
	envFile := strings.TrimSpace(values["--env-file"])
	metadataFile := strings.TrimSpace(values["--metadata-file"])
	if envFile == "" || metadataFile == "" {
		return "", "", errors.New("usage: testservices prepare-web-e2e --env-file <path> --metadata-file <path>")
	}
	return envFile, metadataFile, nil
}

func parseCleanupWebE2EArgs(args []string) (string, error) {
	values, err := parseFlagPairs(args, map[string]struct{}{
		"--metadata-file": {},
	})
	if err != nil {
		return "", err
	}
	metadataFile := strings.TrimSpace(values["--metadata-file"])
	if metadataFile == "" {
		return "", errors.New("usage: testservices cleanup-web-e2e --metadata-file <path>")
	}
	return metadataFile, nil
}

func parseFlagPairs(args []string, allowed map[string]struct{}) (map[string]string, error) {
	if len(args)%2 != 0 {
		return nil, errors.New("testservices command flags require <flag> <value> pairs")
	}

	values := make(map[string]string, len(allowed))
	for i := 0; i < len(args); i += 2 {
		flag := args[i]
		if _, ok := allowed[flag]; !ok {
			return nil, fmt.Errorf("unsupported testservices flag %q", flag)
		}
		if strings.TrimSpace(args[i+1]) == "" {
			return nil, fmt.Errorf("testservices flag %s requires a value", flag)
		}
		if _, exists := values[flag]; exists {
			return nil, fmt.Errorf("testservices flag %s must be specified once", flag)
		}
		values[flag] = args[i+1]
	}
	return values, nil
}

func defaultDependencies() dependencies {
	return dependencies{
		startPostgres:       startPostgresService,
		startMinIO:          startMinIOService,
		startChild:          startChildProcess,
		createTemplate:      createTemplateDatabase,
		prepareWebE2E:       prepareWebE2EFixture,
		cleanupWebE2EDB:     cleanupWebE2EDatabase,
		cleanupWebE2EBucket: cleanupWebE2EBucket,
		detectWebE2ELeaks:   detectWebE2EFixtureLeaks,
		warmImages:          warmServiceImages,
		recordEvent: func(env map[string]string, event suiteservices.Event) {
			_ = suiteservices.RecordEvent(env, event)
		},
		refreshSummary: func(env map[string]string) {
			_ = suiteservices.RefreshSummary(env)
		},
		suiteID:       generateSuiteID,
		notifySignals: signal.Notify,
		stopSignals:   signal.Stop,
	}
}

func serviceImages() []string {
	return []string{
		pgtest.ContainerImage(),
		s3test.ContainerImage(),
	}
}

func warmServiceImages(ctx context.Context, images []string) error {
	var wg sync.WaitGroup
	errCh := make(chan error, len(images))

	for _, image := range images {
		image := strings.TrimSpace(image)
		if image == "" {
			continue
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := warmServiceImage(ctx, image); err != nil {
				errCh <- err
			}
		}()
	}

	wg.Wait()
	close(errCh)

	var errs []error
	for err := range errCh {
		errs = append(errs, err)
	}
	return errors.Join(errs...)
}

func warmServiceImage(ctx context.Context, image string) error {
	inspect := exec.CommandContext(ctx, "docker", "image", "inspect", image)
	inspect.Stdout = ioDiscard{}
	inspect.Stderr = ioDiscard{}
	if err := inspect.Run(); err == nil {
		fmt.Fprintf(os.Stderr, "service image already present: %s\n", image)
		return nil
	}

	fmt.Fprintf(os.Stderr, "pulling service image: %s\n", image)
	pull := exec.CommandContext(ctx, "docker", "pull", image)
	pull.Stdout = os.Stdout
	pull.Stderr = os.Stderr
	if err := pull.Run(); err != nil {
		return fmt.Errorf("pull %s: %w", image, err)
	}
	return nil
}

type ioDiscard struct{}

func (ioDiscard) Write(p []byte) (int, error) {
	return len(p), nil
}

func startPostgresService(ctx context.Context) (postgresService, error) {
	harness, err := pgtest.StartOwned(ctx)
	if err != nil {
		return postgresService{}, err
	}
	return postgresService{
		adminDSN: harness.AdminDSN(),
		dsnTemplate: fmt.Sprintf(
			"postgres://%s:%s@%s:%s/{database}?sslmode=disable",
			harness.User,
			harness.Password,
			harness.Host,
			harness.Port,
		),
		host:  harness.Host,
		port:  harness.Port,
		user:  harness.User,
		close: harness.Close,
	}, nil
}

func startMinIOService(ctx context.Context) (minioService, error) {
	harness, err := s3test.StartOwned(ctx)
	if err != nil {
		return minioService{}, err
	}
	return minioService{
		endpoint:  harness.Endpoint,
		accessKey: harness.AccessKey,
		secretKey: harness.SecretKey,
		secure:    harness.Secure,
		close:     harness.Close,
	}, nil
}

func createTemplateDatabase(ctx context.Context, adminDSN string, templateDB string) error {
	if err := createDatabase(ctx, adminDSN, templateDB); err != nil {
		return err
	}

	templateDSN, err := replaceDatabaseInDSN(adminDSN, templateDB)
	if err != nil {
		return err
	}
	db, err := sql.Open("pgx", templateDSN)
	if err != nil {
		return fmt.Errorf("open template database: %w", err)
	}
	if _, err := postgres.Migrate(db, dbmigrations.Source(), "up"); err != nil {
		_ = db.Close()
		return err
	}
	if err := db.Close(); err != nil {
		return fmt.Errorf("close template database handle: %w", err)
	}

	admin, err := sql.Open("pgx", adminDSN)
	if err != nil {
		return fmt.Errorf("open postgres admin handle: %w", err)
	}
	defer admin.Close()

	if _, err := admin.ExecContext(ctx, `SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = $1 AND pid <> pg_backend_pid()`, templateDB); err != nil {
		return fmt.Errorf("terminate template database connections: %w", err)
	}
	if _, err := admin.ExecContext(ctx, fmt.Sprintf(`ALTER DATABASE "%s" WITH ALLOW_CONNECTIONS false`, templateDB)); err != nil {
		return fmt.Errorf("disable template database connections: %w", err)
	}
	if _, err := admin.ExecContext(ctx, fmt.Sprintf(`ALTER DATABASE "%s" WITH IS_TEMPLATE true`, templateDB)); err != nil {
		return fmt.Errorf("mark template database as template: %w", err)
	}
	return nil
}

func createDatabase(ctx context.Context, adminDSN string, name string) error {
	admin, err := sql.Open("pgx", adminDSN)
	if err != nil {
		return fmt.Errorf("open postgres admin handle: %w", err)
	}
	defer admin.Close()

	if _, err := admin.ExecContext(ctx, fmt.Sprintf(`CREATE DATABASE "%s"`, name)); err != nil {
		return fmt.Errorf("create database %s: %w", name, err)
	}
	return nil
}

func dropDatabase(ctx context.Context, adminDSN string, name string) error {
	admin, err := sql.Open("pgx", adminDSN)
	if err != nil {
		return fmt.Errorf("open postgres admin handle: %w", err)
	}
	defer admin.Close()

	if _, err := admin.ExecContext(ctx, fmt.Sprintf(`DROP DATABASE IF EXISTS "%s" WITH (FORCE)`, name)); err != nil {
		return fmt.Errorf("drop test database %s: %w", name, err)
	}
	return nil
}

func postgresPreparationDetails(env map[string]string, strategy string, templateDB string) map[string]any {
	details := map[string]any{
		"preparation_strategy": strategy,
		"target":               suiteservices.LookupEnvValue(env, suiteservices.TargetEnv),
	}
	if templateDB != "" {
		details["template_database"] = templateDB
	}
	return details
}

func replaceDatabaseInDSN(adminDSN string, database string) (string, error) {
	parsed, err := url.Parse(adminDSN)
	if err != nil {
		return "", fmt.Errorf("parse postgres dsn: %w", err)
	}
	parsed.Path = "/" + database
	return parsed.String(), nil
}

func prepareWebE2EFixture(ctx context.Context, env map[string]string) (webE2EFixture, error) {
	_ = env
	postgresHarness, err := pgtest.StartShared(ctx)
	if err != nil {
		return webE2EFixture{}, fmt.Errorf("attach suite postgres: %w", err)
	}
	testDB, _, err := postgresHarness.PrepareDatabase(ctx, "web_e2e")
	if err != nil {
		return webE2EFixture{}, fmt.Errorf("prepare browser e2e database: %w", err)
	}

	s3Harness, err := s3test.StartShared(ctx)
	if err != nil {
		_ = postgresHarness.DropDatabase(context.Background(), testDB.Name)
		return webE2EFixture{}, fmt.Errorf("attach suite minio: %w", err)
	}
	bucket, err := s3Harness.BootstrapBucket(ctx, "web-e2e")
	if err != nil {
		_ = postgresHarness.DropDatabase(context.Background(), testDB.Name)
		return webE2EFixture{}, fmt.Errorf("prepare browser e2e bucket: %w", err)
	}

	return webE2EFixture{
		DatabaseName: testDB.Name,
		DSN:          testDB.DSN,
		Bucket:       bucket,
		S3Endpoint:   s3Harness.Endpoint,
		S3AccessKey:  s3Harness.AccessKey,
		S3SecretKey:  s3Harness.SecretKey,
		S3Secure:     s3Harness.Secure,
	}, nil
}

func cleanupWebE2EFixture(ctx context.Context, deps dependencies, env map[string]string, metadata webE2EMetadata) error {
	errCh := make(chan error, 2)
	var wg sync.WaitGroup

	if strings.TrimSpace(metadata.DatabaseName) != "" {
		wg.Add(1)
		go func() {
			defer wg.Done()
			start := time.Now().UTC()
			err := deps.cleanupWebE2EDB(ctx, metadata, env)
			recordTimingSpanStatus(deps, env, bucketTeardown, "test-services cleanup browser e2e fixture database", start, err)
			if err != nil {
				errCh <- err
			}
		}()
	}

	if strings.TrimSpace(metadata.Bucket) != "" {
		wg.Add(1)
		go func() {
			defer wg.Done()
			start := time.Now().UTC()
			err := deps.cleanupWebE2EBucket(ctx, metadata, env)
			recordTimingSpanStatus(deps, env, bucketTeardown, "test-services cleanup browser e2e fixture bucket", start, err)
			if err != nil {
				errCh <- err
			}
		}()
	}

	wg.Wait()
	close(errCh)
	var errs []error
	for err := range errCh {
		errs = append(errs, err)
	}
	return errors.Join(errs...)
}

func cleanupWebE2EDatabase(ctx context.Context, metadata webE2EMetadata, env map[string]string) error {
	if adminDSN := strings.TrimSpace(suiteservices.LookupEnvValue(env, suiteservices.PGAdminDSNEnv)); adminDSN != "" {
		return dropDatabase(ctx, adminDSN, metadata.DatabaseName)
	}
	postgresHarness, err := pgtest.StartShared(ctx)
	if err != nil {
		return fmt.Errorf("attach suite postgres: %w", err)
	}
	return postgresHarness.DropDatabase(ctx, metadata.DatabaseName)
}

func cleanupWebE2EBucket(ctx context.Context, metadata webE2EMetadata, env map[string]string) error {
	if endpoint := strings.TrimSpace(suiteservices.LookupEnvValue(env, suiteservices.S3EndpointEnv)); endpoint != "" {
		s3Harness := &s3test.Harness{
			Endpoint:  endpoint,
			AccessKey: suiteservices.LookupEnvValue(env, suiteservices.S3AccessKeyEnv),
			SecretKey: suiteservices.LookupEnvValue(env, suiteservices.S3SecretKeyEnv),
			Secure:    strings.EqualFold(suiteservices.LookupEnvValue(env, suiteservices.S3SecureEnv), "true"),
		}
		return s3Harness.CleanupBucket(ctx, metadata.Bucket)
	}
	s3Harness, err := s3test.StartShared(ctx)
	if err != nil {
		return fmt.Errorf("attach suite minio: %w", err)
	}
	return s3Harness.CleanupBucket(ctx, metadata.Bucket)
}

func detectWebE2EFixtureLeaks(ctx context.Context, fixtures []webE2EMetadata, env map[string]string) error {
	adminDSN := strings.TrimSpace(suiteservices.LookupEnvValue(env, suiteservices.PGAdminDSNEnv))
	if adminDSN == "" || len(fixtures) == 0 {
		return nil
	}

	databases := make([]string, 0, len(fixtures))
	seen := make(map[string]struct{})
	for _, fixture := range fixtures {
		databaseName := strings.TrimSpace(fixture.DatabaseName)
		if databaseName == "" {
			continue
		}
		if _, ok := seen[databaseName]; ok {
			continue
		}
		seen[databaseName] = struct{}{}
		databases = append(databases, databaseName)
	}
	if len(databases) == 0 {
		return nil
	}

	admin, err := sql.Open("pgx", adminDSN)
	if err != nil {
		return fmt.Errorf("open postgres admin handle: %w", err)
	}
	defer admin.Close()

	placeholders := make([]string, 0, len(databases))
	args := make([]any, 0, len(databases))
	for index, databaseName := range databases {
		placeholders = append(placeholders, fmt.Sprintf("$%d", index+1))
		args = append(args, databaseName)
	}
	query := fmt.Sprintf(
		`SELECT datname, count(*) FROM pg_stat_activity WHERE datname IN (%s) AND pid <> pg_backend_pid() GROUP BY datname ORDER BY datname`,
		strings.Join(placeholders, ","),
	)
	rows, err := admin.QueryContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("query browser e2e fixture postgres connections: %w", err)
	}
	defer rows.Close()

	var errs []error
	for rows.Next() {
		var databaseName string
		var count int
		if err := rows.Scan(&databaseName, &count); err != nil {
			return fmt.Errorf("scan browser e2e fixture postgres connections: %w", err)
		}
		if count > 0 {
			errs = append(errs, fmt.Errorf("browser e2e fixture database %q has %d active postgres connection(s)", databaseName, count))
		}
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("read browser e2e fixture postgres connections: %w", err)
	}
	return errors.Join(errs...)
}

func writeWebE2EEnv(path string, fixture webE2EFixture) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	lines := []string{
		"# Generated by cartulary-test-services prepare-web-e2e.",
		shellExport("CARTULARY_POSTGRES_DSN", fixture.DSN),
		shellExport("CARTULARY_S3_ENDPOINT", fixture.S3Endpoint),
		shellExport("CARTULARY_S3_ACCESS_KEY_ID", fixture.S3AccessKey),
		shellExport("CARTULARY_S3_SECRET_ACCESS_KEY", fixture.S3SecretKey),
		shellExport("CARTULARY_S3_SECURE", fmt.Sprintf("%t", fixture.S3Secure)),
		shellExport("CARTULARY_S3_BUCKET", fixture.Bucket),
		"",
	}
	return os.WriteFile(path, []byte(strings.Join(lines, "\n")), 0o600)
}

func shellExport(key string, value string) string {
	return fmt.Sprintf("export %s=%s", key, shellQuote(value))
}

func shellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\"'\"'") + "'"
}

func writeWebE2EMetadata(path string, metadata webE2EMetadata) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	payload, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(payload, '\n'), 0o600)
}

func readWebE2EMetadata(path string) (webE2EMetadata, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return webE2EMetadata{}, err
	}
	var metadata webE2EMetadata
	if err := json.Unmarshal(raw, &metadata); err != nil {
		return webE2EMetadata{}, err
	}
	return metadata, nil
}

func recordCleanupAndRefresh(deps dependencies, env map[string]string, status string, childExitCode int) {
	deps.recordEvent(env, suiteservices.Event{
		Type:   suiteservices.EventCleanupCompleted,
		Status: status,
		Details: map[string]any{
			"child_exit_status": childExitCode,
		},
	})
	deps.refreshSummary(env)
}

func recordTimingSpanStatus(deps dependencies, env map[string]string, bucket string, label string, start time.Time, err error) {
	status := "pass"
	if err != nil {
		status = "fail"
	}
	recordTimingSpan(deps, env, bucket, label, start, time.Now().UTC(), status)
}

func recordTimingSpanStatusAt(deps dependencies, env map[string]string, bucket string, label string, start time.Time, end time.Time, err error) {
	status := "pass"
	if err != nil {
		status = "fail"
	}
	recordTimingSpan(deps, env, bucket, label, start, end, status)
}

func recordTimingSpan(deps dependencies, env map[string]string, bucket string, label string, start time.Time, end time.Time, status string) {
	if end.Before(start) {
		end = start
	}
	duration := end.Sub(start)
	if duration < 0 {
		duration = 0
	}
	deps.recordEvent(env, suiteservices.Event{
		Type:   suiteservices.EventTimingSpan,
		Status: status,
		Details: map[string]any{
			"target":      suiteservices.LookupEnvValue(env, suiteservices.TargetEnv),
			"bucket":      bucket,
			"label":       label,
			"start_time":  start.Format(time.RFC3339Nano),
			"end_time":    end.Format(time.RFC3339Nano),
			"duration_ms": duration.Milliseconds(),
			"status":      status,
		},
	})
}

func recordWebE2EFixtureEvent(deps dependencies, env map[string]string, eventType string, metadata webE2EMetadata) {
	target := metadata.Target
	if strings.TrimSpace(target) == "" {
		target = suiteservices.LookupEnvValue(env, suiteservices.TargetEnv)
	}
	deps.recordEvent(env, suiteservices.Event{
		Type: eventType,
		Details: map[string]any{
			"database_name": metadata.DatabaseName,
			"bucket":        metadata.Bucket,
			"target":        target,
		},
	})
}

func recordWebE2EFixtureReclaimedEvent(deps dependencies, env map[string]string, metadata webE2EMetadata, strategy string) {
	target := metadata.Target
	if strings.TrimSpace(target) == "" {
		target = suiteservices.LookupEnvValue(env, suiteservices.TargetEnv)
	}
	deps.recordEvent(env, suiteservices.Event{
		Type: suiteservices.EventWebE2EFixtureReclaimed,
		Details: map[string]any{
			"database_name":    metadata.DatabaseName,
			"bucket":           metadata.Bucket,
			"target":           target,
			"reclaim_strategy": strategy,
		},
	})
}

func cleanupRetiredWebE2EFixtures(ctx context.Context, deps dependencies, env map[string]string) error {
	scope, ok, err := suiteservices.Summarize(env)
	if err != nil || !ok {
		return err
	}

	return cleanupWebE2EFixtures(ctx, deps, env, pendingWebE2EFixtures(scope))
}

func pendingRetiredWebE2EFixtures(env map[string]string) ([]webE2EMetadata, error) {
	scope, ok, err := suiteservices.Summarize(env)
	if err != nil || !ok {
		return nil, err
	}
	return pendingWebE2EFixtures(scope), nil
}

func reclaimOwnedWebE2EFixtures(deps dependencies, env map[string]string, fixtures []webE2EMetadata) {
	for _, fixture := range fixtures {
		recordWebE2EFixtureReclaimedEvent(deps, env, fixture, webE2EReclaimStrategyOwnedStack)
	}
}

func cleanupWebE2EFixtures(ctx context.Context, deps dependencies, env map[string]string, fixtures []webE2EMetadata) error {
	if len(fixtures) == 0 {
		return nil
	}
	cleanupEnv := cloneEnv(env)
	var (
		mu      sync.Mutex
		errs    []error
		cleaned []webE2EMetadata
	)
	jobs := make(chan webE2EMetadata)
	workerCount := min(resolveWebE2ECleanupWorkers(env), len(fixtures))
	var wg sync.WaitGroup
	wg.Add(workerCount)
	for range workerCount {
		go func() {
			defer wg.Done()
			for fixture := range jobs {
				if err := cleanupWebE2EFixture(ctx, deps, cleanupEnv, fixture); err != nil {
					mu.Lock()
					errs = append(errs, fmt.Errorf("cleanup browser e2e fixture database=%q bucket=%q: %w", fixture.DatabaseName, fixture.Bucket, err))
					mu.Unlock()
					continue
				}
				mu.Lock()
				cleaned = append(cleaned, fixture)
				mu.Unlock()
			}
		}()
	}
	for _, fixture := range fixtures {
		jobs <- fixture
	}
	close(jobs)
	wg.Wait()

	slices.SortFunc(cleaned, func(left, right webE2EMetadata) int {
		leftKey := webE2EFixtureKey(left)
		rightKey := webE2EFixtureKey(right)
		return strings.Compare(leftKey, rightKey)
	})
	for _, fixture := range cleaned {
		recordWebE2EFixtureEvent(deps, env, suiteservices.EventWebE2EFixtureCleaned, fixture)
	}

	return errors.Join(errs...)
}

func resolveWebE2ECleanupWorkers(env map[string]string) int {
	raw := strings.TrimSpace(suiteservices.LookupEnvValue(env, webE2ECleanupWorkersEnv))
	if raw == "" {
		return webE2ECleanupWorkers
	}
	workers, err := strconv.Atoi(raw)
	if err != nil || workers < 1 {
		return webE2ECleanupWorkers
	}
	if workers > webE2ECleanupMaxWorkers {
		return webE2ECleanupMaxWorkers
	}
	return workers
}

func cleanupStaleWebE2EFixtures(ctx context.Context, deps dependencies, env map[string]string) error {
	fixtures, err := staleWebE2EFixtures(env)
	if err != nil || len(fixtures) == 0 {
		return err
	}
	return cleanupWebE2EFixtures(ctx, deps, env, fixtures)
}

func staleWebE2EFixtures(env map[string]string) ([]webE2EMetadata, error) {
	resultsRoot, err := suiteservices.ResolveResultsRoot(env)
	if err != nil {
		return nil, err
	}
	servicesRoot := filepath.Join(resultsRoot, suiteservices.ResolveRunID(env), "_shared", "test-services")

	completed := make(map[string]struct{})
	retired := make(map[string]webE2EMetadata)
	err = filepath.WalkDir(servicesRoot, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || entry.Name() != "service-scope.json" {
			return nil
		}
		raw, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		var scope suiteservices.ServiceScope
		if err := json.Unmarshal(raw, &scope); err != nil {
			return err
		}
		for _, fixture := range scope.BrowserE2E.CleanedFixtures {
			metadata := webE2EMetadata{
				DatabaseName: fixture.DatabaseName,
				Bucket:       fixture.Bucket,
				Target:       fixture.Target,
			}
			if key := webE2EFixtureKey(metadata); key != "" {
				completed[key] = struct{}{}
			}
		}
		for _, fixture := range scope.BrowserE2E.ReclaimedFixtures {
			metadata := webE2EMetadata{
				DatabaseName: fixture.DatabaseName,
				Bucket:       fixture.Bucket,
				Target:       fixture.Target,
			}
			if key := webE2EFixtureKey(metadata); key != "" {
				completed[key] = struct{}{}
			}
		}
		for _, fixture := range scope.BrowserE2E.RetiredFixtures {
			metadata := webE2EMetadata{
				DatabaseName: fixture.DatabaseName,
				Bucket:       fixture.Bucket,
				Target:       fixture.Target,
			}
			key := webE2EFixtureKey(metadata)
			if key == "" || !generatedWebE2EFixture(metadata) {
				continue
			}
			retired[key] = metadata
		}
		return nil
	})
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}

	fixtures := make([]webE2EMetadata, 0, len(retired))
	for key, fixture := range retired {
		if _, ok := completed[key]; ok {
			continue
		}
		fixtures = append(fixtures, fixture)
	}
	slices.SortFunc(fixtures, func(left, right webE2EMetadata) int {
		return strings.Compare(webE2EFixtureKey(left), webE2EFixtureKey(right))
	})
	if len(fixtures) > staleFixtureMaxCandidates {
		fixtures = fixtures[:staleFixtureMaxCandidates]
	}
	return fixtures, nil
}

func generatedWebE2EFixture(metadata webE2EMetadata) bool {
	databaseName := strings.TrimSpace(metadata.DatabaseName)
	bucket := strings.TrimSpace(metadata.Bucket)
	return generatedWebE2EDatabaseName(databaseName) && generatedWebE2EBucketName(bucket)
}

func generatedWebE2EDatabaseName(name string) bool {
	if !strings.HasPrefix(name, "ct_") || !strings.HasSuffix(name, "_web_e2e") {
		return false
	}
	for _, char := range name {
		if char >= 'a' && char <= 'z' {
			continue
		}
		if char >= '0' && char <= '9' {
			continue
		}
		if char == '_' {
			continue
		}
		return false
	}
	return true
}

func generatedWebE2EBucketName(name string) bool {
	if !strings.HasPrefix(name, "ct-") || !strings.HasSuffix(name, "-web-e2e") {
		return false
	}
	for _, char := range name {
		if char >= 'a' && char <= 'z' {
			continue
		}
		if char >= '0' && char <= '9' {
			continue
		}
		if char == '-' {
			continue
		}
		return false
	}
	return true
}

func pendingWebE2EFixtures(scope suiteservices.ServiceScope) []webE2EMetadata {
	completed := make(map[string]struct{})
	for _, fixture := range scope.BrowserE2E.CleanedFixtures {
		metadata := webE2EMetadata{
			DatabaseName: fixture.DatabaseName,
			Bucket:       fixture.Bucket,
			Target:       fixture.Target,
		}
		completed[webE2EFixtureKey(metadata)] = struct{}{}
	}
	for _, fixture := range scope.BrowserE2E.ReclaimedFixtures {
		metadata := webE2EMetadata{
			DatabaseName: fixture.DatabaseName,
			Bucket:       fixture.Bucket,
			Target:       fixture.Target,
		}
		completed[webE2EFixtureKey(metadata)] = struct{}{}
	}

	pending := make(map[string]webE2EMetadata)
	for _, fixture := range scope.BrowserE2E.RetiredFixtures {
		metadata := webE2EMetadata{
			DatabaseName: fixture.DatabaseName,
			Bucket:       fixture.Bucket,
			Target:       fixture.Target,
		}
		key := webE2EFixtureKey(metadata)
		if key == "" {
			continue
		}
		if _, ok := completed[key]; ok {
			continue
		}
		pending[key] = metadata
	}

	fixtures := make([]webE2EMetadata, 0, len(pending))
	for _, fixture := range pending {
		fixtures = append(fixtures, fixture)
	}
	slices.SortFunc(fixtures, func(left, right webE2EMetadata) int {
		return strings.Compare(webE2EFixtureKey(left), webE2EFixtureKey(right))
	})
	return fixtures
}

func webE2EFixtureKey(metadata webE2EMetadata) string {
	return strings.Join([]string{
		strings.TrimSpace(metadata.DatabaseName),
		strings.TrimSpace(metadata.Bucket),
	}, "\x1f")
}

func cleanupOwnedServices(deps dependencies, env map[string]string, postgresSvc postgresService, minioSvc minioService, status string, childExitCode int) {
	cleanupCtx, cancel := context.WithTimeout(context.Background(), cleanupTimeout)
	defer cancel()

	cleanupStatus := status
	if cleanupStatus == "" {
		cleanupStatus = "succeeded"
	}
	browserCleanupFailed := false

	cleanupEnv := serviceBackedCleanupEnv(env, postgresSvc, minioSvc)
	pendingFixtures, err := pendingRetiredWebE2EFixtures(cleanupEnv)
	if err != nil {
		cleanupStatus = "cleanup_failed"
		browserCleanupFailed = true
		printSuiteFailure(env, failureSummary("", stageCleanupWebE2E, "summarize retired browser e2e fixtures", err))
		pendingFixtures = nil
	}

	if !browserCleanupFailed {
		leakCtx, cancelLeakCheck := context.WithTimeout(cleanupCtx, webE2ELeakCheckTimeout)
		leakCheckStart := time.Now().UTC()
		leakErr := deps.detectWebE2ELeaks(leakCtx, pendingFixtures, cleanupEnv)
		cancelLeakCheck()
		if leakErr != nil {
			cleanupStatus = "cleanup_failed"
			browserCleanupFailed = true
			recordTimingSpanStatus(deps, env, bucketTeardown, "test-services check browser e2e fixture leaks", leakCheckStart, leakErr)
			printSuiteFailure(env, failureSummary("", stageCleanupWebE2E, "check browser e2e fixture leaks", leakErr))
		} else {
			recordTimingSpan(deps, env, bucketTeardown, "test-services check browser e2e fixture leaks", leakCheckStart, time.Now().UTC(), "pass")

			reclaimStart := time.Now().UTC()
			reclaimOwnedWebE2EFixtures(deps, env, pendingFixtures)
			recordTimingSpan(deps, env, bucketTeardown, "test-services reclaim pooled browser fixtures by owned stack", reclaimStart, time.Now().UTC(), "pass")
		}
	}
	deps.refreshSummary(env)

	if !browserCleanupFailed {
		janitorStart := time.Now().UTC()
		if err := cleanupStaleWebE2EFixtures(cleanupCtx, deps, cleanupEnv); err != nil {
			cleanupStatus = "cleanup_failed"
			recordTimingSpanStatus(deps, env, bucketTeardown, "test-services janitor stale browser fixtures", janitorStart, err)
			printSuiteFailure(env, failureSummary("", stageCleanupJanitor, "janitor stale browser e2e fixtures", err))
		} else {
			recordTimingSpan(deps, env, bucketTeardown, "test-services janitor stale browser fixtures", janitorStart, time.Now().UTC(), "pass")
		}
	}

	for _, result := range cleanupServices(cleanupCtx, postgresSvc, minioSvc) {
		recordTimingSpanStatusAt(deps, env, bucketTeardown, result.label, result.start, result.end, result.err)
		if result.err != nil {
			cleanupStatus = "cleanup_failed"
			printSuiteFailure(env, failureSummary(result.service, result.stage, result.label, result.err))
		}
	}

	recordCleanupAndRefresh(deps, env, cleanupStatus, childExitCode)
}

func cleanupServices(ctx context.Context, postgresSvc postgresService, minioSvc minioService) []serviceCleanupResult {
	results := make(chan serviceCleanupResult, 2)
	var wg sync.WaitGroup
	if postgresSvc.close != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			start := time.Now().UTC()
			err := postgresSvc.close(ctx)
			results <- serviceCleanupResult{
				service: suiteservices.ServicePostgres,
				stage:   stageCleanupPostgres,
				label:   "test-services terminate postgres",
				start:   start,
				end:     time.Now().UTC(),
				err:     err,
			}
		}()
	}
	if minioSvc.close != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			start := time.Now().UTC()
			err := minioSvc.close(ctx)
			results <- serviceCleanupResult{
				service: suiteservices.ServiceMinIO,
				stage:   stageCleanupMinIO,
				label:   "test-services terminate minio",
				start:   start,
				end:     time.Now().UTC(),
				err:     err,
			}
		}()
	}
	wg.Wait()
	close(results)

	cleanups := make([]serviceCleanupResult, 0, len(results))
	for result := range results {
		cleanups = append(cleanups, result)
	}
	slices.SortFunc(cleanups, func(left, right serviceCleanupResult) int {
		return strings.Compare(left.label, right.label)
	})
	return cleanups
}

func serviceBackedCleanupEnv(env map[string]string, postgresSvc postgresService, minioSvc minioService) map[string]string {
	cleanupEnv := cloneEnv(env)
	if postgresSvc.adminDSN != "" {
		cleanupEnv[suiteservices.PGAdminDSNEnv] = postgresSvc.adminDSN
	}
	if minioSvc.endpoint != "" {
		cleanupEnv[suiteservices.S3EndpointEnv] = minioSvc.endpoint
	}
	if minioSvc.accessKey != "" {
		cleanupEnv[suiteservices.S3AccessKeyEnv] = minioSvc.accessKey
	}
	if minioSvc.secretKey != "" {
		cleanupEnv[suiteservices.S3SecretKeyEnv] = minioSvc.secretKey
	}
	cleanupEnv[suiteservices.S3SecureEnv] = fmt.Sprintf("%t", minioSvc.secure)
	return cleanupEnv
}

func failureSummary(service string, stage string, operation string, err error) suiteservices.FailureSummary {
	failure := suiteservices.FailureSummary{
		Service:   service,
		Stage:     stage,
		Operation: operation,
	}

	var startFailure *testcontainersx.StartFailure
	if errors.As(err, &startFailure) {
		failure.Operation = startFailure.Operation
		failure.Message = startFailure.DiagnosticMessage()
		failure.AttemptsStarted = startFailure.AttemptsStarted
		failure.MaxAttempts = startFailure.MaxAttempts
		failure.Retryable = startFailure.Retryable
		failure.RetryBlockedByContext = startFailure.RetryBlockedByContext
		failure.DockerEndpoint = startFailure.DockerEndpoint
	} else if err != nil {
		failure.Message = err.Error()
	}

	failure.Message = suiteservices.SanitizeDiagnosticText(failure.Message)
	failure.DockerEndpoint = suiteservices.SanitizeDiagnosticText(failure.DockerEndpoint)
	return failure
}

func recordFailureAndRefresh(deps dependencies, env map[string]string, failure suiteservices.FailureSummary) {
	deps.recordEvent(env, suiteservices.Event{
		Type:    suiteservices.EventFailureRecorded,
		Service: failure.Service,
		Details: map[string]any{
			"stage":                    failure.Stage,
			"operation":                failure.Operation,
			"message":                  failure.Message,
			"attempts_started":         failure.AttemptsStarted,
			"max_attempts":             failure.MaxAttempts,
			"retryable":                failure.Retryable,
			"retry_blocked_by_context": failure.RetryBlockedByContext,
			"docker_endpoint":          failure.DockerEndpoint,
		},
	})
	deps.refreshSummary(env)
	printSuiteFailure(env, failure)
}

func printSuiteFailure(env map[string]string, failure suiteservices.FailureSummary) {
	label := failure.Stage
	if strings.TrimSpace(failure.Service) != "" {
		label = fmt.Sprintf("%s (%s)", label, failure.Service)
	}

	artifactDir, ok, err := suiteservices.ResolveSuiteArtifactDir(env)
	if err != nil || !ok {
		artifactDir = "unresolved"
	}

	message := failure.Message
	if strings.TrimSpace(message) == "" {
		message = "unknown failure"
	}
	fmt.Fprintf(os.Stderr, "suite failure %s: %s [artifacts: %s]\n", label, message, artifactDir)
}

func waitForChild(command []string, child childProcess, deps dependencies) int {
	signals := make(chan os.Signal, 1)
	deps.notifySignals(signals, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer deps.stopSignals(signals)

	waitResult := make(chan error, 1)
	go func() {
		waitResult <- child.Wait()
	}()

	signaled := false
	var killTimer <-chan time.Time
	for {
		select {
		case err := <-waitResult:
			exitCode := exitCode(err)
			if signaled && exitCode == 0 {
				return 1
			}
			return exitCode
		case sig := <-signals:
			signaled = true
			if err := child.Signal(sig); err != nil {
				fmt.Fprintf(os.Stderr, "forward signal %s to %s: %v\n", sig, strings.Join(command, " "), err)
				_ = child.Kill()
			}
			killTimer = time.After(signalWaitTimeout)
		case <-killTimer:
			fmt.Fprintf(os.Stderr, "child process %d did not exit after signal, forcing termination\n", child.PID())
			_ = child.Kill()
			killTimer = nil
		}
	}
}

func exitCode(err error) int {
	if err == nil {
		return 0
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return exitErr.ExitCode()
	}
	return 1
}

func generateSuiteID() (string, error) {
	var data [12]byte
	if _, err := rand.Read(data[:]); err != nil {
		return "", fmt.Errorf("read random suite id bytes: %w", err)
	}
	return hex.EncodeToString(data[:]), nil
}

func templateDatabaseName(suiteID string) string {
	return fmt.Sprintf("ct_tpl_%s", suiteservices.ShortHash(suiteID, 12))
}

func cloneEnv(env map[string]string) map[string]string {
	cloned := make(map[string]string, len(env))
	for key, value := range env {
		cloned[key] = value
	}
	return cloned
}

func envMap(entries []string) map[string]string {
	values := make(map[string]string, len(entries))
	for _, entry := range entries {
		key, value, ok := strings.Cut(entry, "=")
		if !ok {
			values[entry] = ""
			continue
		}
		values[key] = value
	}
	return values
}

func envSlice(values map[string]string) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	slices.Sort(keys)

	entries := make([]string, 0, len(keys))
	for _, key := range keys {
		entries = append(entries, fmt.Sprintf("%s=%s", key, values[key]))
	}
	return entries
}

type execChild struct {
	cmd *exec.Cmd
}

func startChildProcess(argv []string, env map[string]string) (childProcess, error) {
	cmd := exec.Command(argv[0], argv[1:]...)
	cmd.Env = envSlice(env)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	return &execChild{cmd: cmd}, nil
}

func (c *execChild) Wait() error {
	return c.cmd.Wait()
}

func (c *execChild) Signal(sig os.Signal) error {
	if c.cmd.Process == nil {
		return nil
	}
	return syscall.Kill(-c.cmd.Process.Pid, sig.(syscall.Signal))
}

func (c *execChild) Kill() error {
	if c.cmd.Process == nil {
		return nil
	}
	return syscall.Kill(-c.cmd.Process.Pid, syscall.SIGKILL)
}

func (c *execChild) PID() int {
	if c.cmd.Process == nil {
		return 0
	}
	return c.cmd.Process.Pid
}
