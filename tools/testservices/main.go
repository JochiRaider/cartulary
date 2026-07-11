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
	dockercontainer "github.com/moby/moby/api/types/container"
	dockerclient "github.com/moby/moby/client"

	dbmigrations "github.com/JochiRaider/cartulary/db/migrations"
	"github.com/JochiRaider/cartulary/internal/platform/objectstore"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/testutil/pgschema"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
	"github.com/JochiRaider/cartulary/internal/testutil/s3test"
	"github.com/JochiRaider/cartulary/internal/testutil/suiteservices"
	"github.com/JochiRaider/cartulary/internal/testutil/testcontainersx"
)

const (
	postgresStartupTimeout       = 3 * time.Minute
	suitePreflightTimeout        = 3 * time.Second
	suitePostgresAttemptLimit    = 35 * time.Second
	suiteObjectStoreAttemptLimit = 2 * time.Minute
	staleSuiteContainerAge       = 10 * time.Minute
	staleSuiteContainerRecheck   = 2 * time.Second
	templateStartupTimeout       = 2 * time.Minute
	objectStoreStartupTimeout    = 2 * time.Minute
	cleanupTimeout               = 2 * time.Minute
	webE2ELeakCheckTimeout       = 5 * time.Second
	signalWaitTimeout            = 15 * time.Second
	webE2ECleanupWorkers         = 4
	webE2ECleanupMaxWorkers      = 16
	staleFixtureMaxCandidates    = 32
	staleFixtureJanitorTimeout   = 10 * time.Second

	webE2ECleanupWorkersEnv = "CARTULARY_TEST_SERVICES_WEB_E2E_CLEANUP_WORKERS"

	testServiceLabelManaged = "cartulary.test-services.managed"
	testServiceLabelSuiteID = "cartulary.test-services.suite-id"
	testServiceLabelRunID   = "cartulary.test-services.run-id"
	testServiceLabelTarget  = "cartulary.test-services.target"
	testServiceLabelService = "cartulary.test-services.service"

	testServiceManagedValue = "true"

	stagePostgresStart      = "postgres-start"
	stageStartupPreflight   = "startup-preflight"
	stagePostgresTemplate   = "postgres-template"
	stageObjectStoreStart   = "object-store-start"
	stageChildStart         = "child-start"
	stageCleanupLease       = "cleanup-lease"
	stageCleanupReaper      = "cleanup-reaper"
	stageCleanupWebE2E      = "cleanup-web-e2e"
	stageCleanupJanitor     = "cleanup-janitor"
	stageCleanupPostgres    = "cleanup-postgres"
	stageCleanupObjectStore = "cleanup-object-store"

	webE2EReclaimStrategyOwnedStack = "owned_stack_termination"

	bucketSetup       = "setup"
	bucketServiceWait = "service_wait"
	bucketMigration   = "migration"
	bucketTeardown    = "teardown"
)

var (
	startPostgresHarnessWithOptions    = pgtest.StartOwnedWithOptions
	startObjectStoreHarnessWithOptions = s3test.StartOwnedWithOptions
)

type postgresService struct {
	adminDSN    string
	dsnTemplate string
	host        string
	port        string
	user        string
	close       func(context.Context) error
	containerID string
	name        string
	image       string
	labels      map[string]string
}

type objectStoreService struct {
	endpoint    string
	accessKey   string
	secretKey   string
	secure      bool
	close       func(context.Context) error
	containerID string
	name        string
	image       string
	labels      map[string]string
}

type postgresStartResult struct {
	service postgresService
	start   time.Time
	end     time.Time
	err     error
}

type objectStoreStartResult struct {
	service objectStoreService
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

type suitePreflightResult struct {
	DockerEndpoint              string
	DockerOK                    bool
	ReaperReady                 bool
	StaleContainersScanned      int
	StaleContainersRemoved      int
	StaleContainersDeferred     int
	RyukDisabledForSuiteStartup bool
}

type staleSuiteContainerCleanupSummary struct {
	Scanned  int
	Removed  int
	Deferred int
}

type suiteContainerClient interface {
	ContainerList(context.Context, dockerclient.ContainerListOptions) (dockerclient.ContainerListResult, error)
	ContainerRemove(context.Context, string, dockerclient.ContainerRemoveOptions) (dockerclient.ContainerRemoveResult, error)
	ContainerInspect(context.Context, string, dockerclient.ContainerInspectOptions) (dockerclient.ContainerInspectResult, error)
}

type serviceLease struct {
	SchemaID      string                 `json:"schema_id"`
	LeaseID       string                 `json:"lease_id"`
	SuiteID       string                 `json:"suite_id"`
	RunID         string                 `json:"run_id"`
	ResultRoot    string                 `json:"result_root"`
	RunRoot       string                 `json:"run_root"`
	Target        string                 `json:"target"`
	Mode          string                 `json:"mode"`
	OwnershipMode string                 `json:"ownership_mode"`
	OwnerPID      int                    `json:"owner_pid"`
	CreatedAt     string                 `json:"created_at"`
	ExpiresAt     string                 `json:"expires_at,omitempty"`
	Resources     []serviceLeaseResource `json:"resources"`
	ProofLabels   map[string]string      `json:"proof_labels"`
	ProofPrefixes map[string]string      `json:"proof_prefixes"`
	CleanupState  string                 `json:"cleanup_state"`
}

type serviceLeaseResource struct {
	Kind        string            `json:"kind"`
	Service     string            `json:"service"`
	ContainerID string            `json:"container_id"`
	Name        string            `json:"name,omitempty"`
	Image       string            `json:"image"`
	Labels      map[string]string `json:"labels"`
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
	startPostgres       func(context.Context, map[string]string) (postgresService, error)
	startObjectStore    func(context.Context, map[string]string) (objectStoreService, error)
	startChild          func(argv []string, env map[string]string) (childProcess, error)
	startReaper         func(leasePath string, env map[string]string) error
	preflightSuite      func(context.Context, map[string]string) (suitePreflightResult, error)
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
		fmt.Fprintln(os.Stderr, "usage: testservices run -- <command> [args...] | start-suite --env-file <path> --lease-file <path> | record-lifecycle --env-file <path> --event <event> [--child-key <key>] | prepare-web-e2e --env-file <path> --metadata-file <path> | cleanup-web-e2e --metadata-file <path> | terminate-suite --lease <path> | images | warm-images")
		return 2
	}

	switch args[0] {
	case "run":
		return runWrappedCommand(args, env, deps)
	case "start-suite":
		return runStartSuite(args[1:], env, deps)
	case "record-lifecycle":
		return runRecordLifecycle(args[1:], env)
	case "prepare-web-e2e":
		return runPrepareWebE2E(args[1:], env, deps)
	case "cleanup-web-e2e":
		return runCleanupWebE2E(args[1:], env, deps)
	case "terminate-suite":
		return runTerminateSuite(args[1:], env, deps)
	case "images":
		return runImages(args[1:])
	case "warm-images":
		return runWarmImages(args[1:], deps)
	default:
		fmt.Fprintf(os.Stderr, "usage: unknown testservices command %q\n", args[0])
		return 2
	}
}

func runImages(args []string) int {
	if len(args) != 0 {
		fmt.Fprintln(os.Stderr, "usage: testservices images")
		return 2
	}
	for _, image := range serviceImages() {
		fmt.Println(image)
	}
	return 0
}

func startPostgresAsync(parent context.Context, deps dependencies, env map[string]string) <-chan postgresStartResult {
	resultCh := make(chan postgresStartResult, 1)
	go func() {
		ctx, cancel := context.WithTimeout(parent, postgresStartupTimeout)
		defer cancel()

		result := postgresStartResult{start: time.Now().UTC()}
		result.service, result.err = deps.startPostgres(ctx, env)
		result.end = time.Now().UTC()
		resultCh <- result
	}()
	return resultCh
}

func startObjectStoreAsync(parent context.Context, deps dependencies, env map[string]string) <-chan objectStoreStartResult {
	resultCh := make(chan objectStoreStartResult, 1)
	go func() {
		ctx, cancel := context.WithTimeout(parent, objectStoreStartupTimeout)
		defer cancel()

		result := objectStoreStartResult{start: time.Now().UTC()}
		result.service, result.err = deps.startObjectStore(ctx, env)
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
	ownedEnv[suiteservices.LifecycleModeEnv] = "owned"

	deps.recordEvent(ownedEnv, suiteservices.Event{Type: suiteservices.EventWrapperOwnedStart})
	if err := suiteservices.RecordLifecycleEvent(ownedEnv, suiteservices.LifecycleEventStartServices, ""); err != nil {
		fmt.Fprintf(os.Stderr, "record service lifecycle start: %v\n", err)
		return 1
	}
	deps.refreshSummary(ownedEnv)
	recordTimingSpan(deps, ownedEnv, bucketSetup, "test-services wrapper setup", wrapperStart, time.Now().UTC(), "pass")

	restoreRyuk := disableTestcontainersRyukForSuiteStartup()
	defer restoreRyuk()

	preflightStart := time.Now().UTC()
	preflightCtx, cancelPreflight := context.WithTimeout(context.Background(), suitePreflightTimeout)
	preflightResult, err := runConfiguredSuitePreflight(preflightCtx, deps, ownedEnv)
	cancelPreflight()
	preflightResult.RyukDisabledForSuiteStartup = true
	recordSuitePreflight(deps, ownedEnv, preflightResult, err)
	recordTimingSpanStatus(deps, ownedEnv, bucketSetup, "test-services suite startup preflight", preflightStart, err)
	deps.refreshSummary(ownedEnv)
	if err != nil {
		recordStartupFailureAndRefresh(deps, ownedEnv, failureSummary("", stageStartupPreflight, "suite startup preflight", err))
		recordCleanupAndRefresh(deps, ownedEnv, "startup_failed", 1)
		return 1
	}

	var postgresSvc postgresService
	var objectStoreSvc objectStoreService
	childExitCode := 1
	cleanupStatus := "startup_failed"
	leasePath := ""

	startupCtx, cancelStartup := context.WithCancel(context.Background())
	defer cancelStartup()
	postgresResultCh := startPostgresAsync(startupCtx, deps, ownedEnv)
	objectStoreResultCh := startObjectStoreAsync(startupCtx, deps, ownedEnv)

	postgresResult := <-postgresResultCh
	postgresSvc = postgresResult.service
	recordTimingSpanStatusAt(deps, ownedEnv, bucketServiceWait, "test-services start postgres", postgresResult.start, postgresResult.end, postgresResult.err)
	if postgresResult.err != nil {
		cancelStartup()
		objectStoreResult := <-objectStoreResultCh
		objectStoreSvc = objectStoreResult.service
		recordTimingSpanStatusAt(deps, ownedEnv, bucketServiceWait, "test-services start object-store", objectStoreResult.start, objectStoreResult.end, objectStoreResult.err)
		recordStartupFailureAndRefresh(deps, ownedEnv, failureSummary(suiteservices.ServicePostgres, stagePostgresStart, "start suite postgres", postgresResult.err))
		cleanupOwnedServices(deps, ownedEnv, postgresSvc, objectStoreSvc, "", "startup_failed", childExitCode)
		return 1
	}
	defer func() {
		cleanupOwnedServices(deps, ownedEnv, postgresSvc, objectStoreSvc, leasePath, cleanupStatus, childExitCode)
	}()

	schemaHash, err := pgschema.Hash()
	if err != nil {
		recordStartupFailureAndRefresh(deps, ownedEnv, failureSummary(suiteservices.ServicePostgres, stagePostgresTemplate, "hash postgres schema", err))
		return 1
	}
	ownedEnv[suiteservices.PGSchemaHashEnv] = schemaHash
	templateDB := templateDatabaseName(suiteID, schemaHash)
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
	err = withSuiteGooseLog(ownedEnv, func() error {
		return deps.createTemplate(templateCtx, postgresSvc.adminDSN, templateDB)
	})
	recordTimingSpanStatus(deps, ownedEnv, bucketMigration, "test-services prepare postgres template database", templateStart, err)
	cancelTemplate()
	if err != nil {
		cancelStartup()
		objectStoreResult := <-objectStoreResultCh
		objectStoreSvc = objectStoreResult.service
		recordTimingSpanStatusAt(deps, ownedEnv, bucketServiceWait, "test-services start object-store", objectStoreResult.start, objectStoreResult.end, objectStoreResult.err)
		recordStartupFailureAndRefresh(deps, ownedEnv, failureSummary(suiteservices.ServicePostgres, stagePostgresTemplate, "prepare postgres template database", err))
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

	objectStoreResult := <-objectStoreResultCh
	objectStoreSvc = objectStoreResult.service
	recordTimingSpanStatusAt(deps, ownedEnv, bucketServiceWait, "test-services start object-store", objectStoreResult.start, objectStoreResult.end, objectStoreResult.err)
	if objectStoreResult.err != nil {
		recordStartupFailureAndRefresh(deps, ownedEnv, failureSummary(suiteservices.ServiceObjectStore, stageObjectStoreStart, "start suite object-store", objectStoreResult.err))
		return 1
	}
	deps.recordEvent(ownedEnv, suiteservices.Event{
		Type:    suiteservices.EventServiceStarted,
		Service: suiteservices.ServiceObjectStore,
		Details: map[string]any{
			"endpoint": objectStoreSvc.endpoint,
			"secure":   objectStoreSvc.secure,
		},
	})
	deps.refreshSummary(ownedEnv)

	leasePath, err = writeServiceLease(ownedEnv, postgresSvc, objectStoreSvc)
	if err != nil {
		recordStartupFailureAndRefresh(deps, ownedEnv, failureSummary("", stageCleanupLease, "write suite service lease", err))
		cleanupStatus = "startup_failed"
		return 1
	}

	janitorCtx, cancelJanitor := context.WithTimeout(context.Background(), staleFixtureJanitorTimeout)
	janitorStart := time.Now().UTC()
	err = cleanupStaleWebE2EFixtures(janitorCtx, deps, serviceBackedCleanupEnv(ownedEnv, postgresSvc, objectStoreSvc))
	recordTimingSpanStatus(deps, ownedEnv, bucketSetup, "test-services janitor stale browser fixtures", janitorStart, err)
	cancelJanitor()
	if err != nil {
		recordStartupFailureAndRefresh(deps, ownedEnv, failureSummary("", stageCleanupJanitor, "janitor stale browser e2e fixtures", err))
		cleanupStatus = "startup_failed"
		return 1
	}
	if err := suiteservices.RecordLifecycleEvent(ownedEnv, suiteservices.LifecycleEventReadinessPassed, ""); err != nil {
		recordFailureAndRefresh(deps, ownedEnv, failureSummary("", stageChildStart, "record service readiness lifecycle", err))
		cleanupStatus = "startup_failed"
		return 1
	}

	childEnv := cloneEnv(ownedEnv)
	childEnv[suiteservices.PGAdminDSNEnv] = postgresSvc.adminDSN
	childEnv[suiteservices.PGDSNTemplateEnv] = postgresSvc.dsnTemplate
	childEnv[suiteservices.PGTemplateDBEnv] = templateDB
	childEnv[suiteservices.PGSchemaHashEnv] = schemaHash
	childEnv[suiteservices.S3EndpointEnv] = objectStoreSvc.endpoint
	childEnv[suiteservices.S3AccessKeyEnv] = objectStoreSvc.accessKey
	childEnv[suiteservices.S3SecretKeyEnv] = objectStoreSvc.secretKey
	childEnv[suiteservices.S3SecureEnv] = fmt.Sprintf("%t", objectStoreSvc.secure)

	child, err := deps.startChild(command, childEnv)
	if err != nil {
		recordFailureAndRefresh(deps, ownedEnv, failureSummary("", stageChildStart, "start child command", err))
		cleanupStatus = "child_start_failed"
		return 1
	}

	childKey := fmt.Sprintf("%d:%s", child.PID(), strings.Join(command, " "))
	if err := suiteservices.RecordLifecycleEvent(ownedEnv, suiteservices.LifecycleEventChildStarted, childKey); err != nil {
		recordFailureAndRefresh(deps, ownedEnv, failureSummary("", stageChildStart, "record child lifecycle start", err))
		cleanupStatus = "child_start_failed"
		return 1
	}
	childExitCode = waitForChild(command, child, deps)
	if err := suiteservices.RecordLifecycleEvent(ownedEnv, suiteservices.LifecycleEventChildFinished, childKey); err != nil {
		recordFailureAndRefresh(deps, ownedEnv, failureSummary("", stageChildStart, "record child lifecycle finish", err))
		cleanupStatus = "failed"
		return 1
	}
	if childExitCode == 0 {
		cleanupStatus = "succeeded"
		return 0
	}
	cleanupStatus = "failed"
	return childExitCode
}

func runStartSuite(args []string, env map[string]string, deps dependencies) int {
	wrapperStart := time.Now().UTC()
	envFile, leaseFile, usageErr := parseStartSuiteArgs(args)
	if usageErr != nil {
		fmt.Fprintln(os.Stderr, usageErr)
		return 2
	}
	if suiteservices.SuiteActive(env) {
		fmt.Fprintln(os.Stderr, "start-suite requires ownership of a new suite; nested active suites are not supported")
		return 2
	}

	suiteID, err := deps.suiteID()
	if err != nil {
		fmt.Fprintf(os.Stderr, "generate suite id: %v\n", err)
		return 1
	}

	ownedEnv := cloneEnv(env)
	ownedEnv[suiteservices.ActiveEnv] = "1"
	ownedEnv[suiteservices.SuiteIDEnv] = suiteID
	ownedEnv[suiteservices.LifecycleModeEnv] = "owned"

	deps.recordEvent(ownedEnv, suiteservices.Event{Type: suiteservices.EventWrapperOwnedStart})
	if err := suiteservices.RecordLifecycleEvent(ownedEnv, suiteservices.LifecycleEventStartServices, ""); err != nil {
		fmt.Fprintf(os.Stderr, "record service lifecycle start: %v\n", err)
		return 1
	}
	deps.refreshSummary(ownedEnv)
	recordTimingSpan(deps, ownedEnv, bucketSetup, "test-services wrapper setup", wrapperStart, time.Now().UTC(), "pass")

	restoreRyuk := disableTestcontainersRyukForSuiteStartup()
	defer restoreRyuk()

	preflightStart := time.Now().UTC()
	preflightCtx, cancelPreflight := context.WithTimeout(context.Background(), suitePreflightTimeout)
	preflightResult, err := runConfiguredSuitePreflight(preflightCtx, deps, ownedEnv)
	cancelPreflight()
	preflightResult.RyukDisabledForSuiteStartup = true
	recordSuitePreflight(deps, ownedEnv, preflightResult, err)
	recordTimingSpanStatus(deps, ownedEnv, bucketSetup, "test-services suite startup preflight", preflightStart, err)
	deps.refreshSummary(ownedEnv)
	if err != nil {
		recordStartupFailureAndRefresh(deps, ownedEnv, failureSummary("", stageStartupPreflight, "suite startup preflight", err))
		recordCleanupAndRefresh(deps, ownedEnv, "startup_failed", 1)
		return 1
	}

	var postgresSvc postgresService
	var objectStoreSvc objectStoreService
	startupCtx, cancelStartup := context.WithCancel(context.Background())
	defer cancelStartup()
	postgresResultCh := startPostgresAsync(startupCtx, deps, ownedEnv)
	objectStoreResultCh := startObjectStoreAsync(startupCtx, deps, ownedEnv)

	postgresResult := <-postgresResultCh
	postgresSvc = postgresResult.service
	recordTimingSpanStatusAt(deps, ownedEnv, bucketServiceWait, "test-services start postgres", postgresResult.start, postgresResult.end, postgresResult.err)
	if postgresResult.err != nil {
		cancelStartup()
		objectStoreResult := <-objectStoreResultCh
		objectStoreSvc = objectStoreResult.service
		recordTimingSpanStatusAt(deps, ownedEnv, bucketServiceWait, "test-services start object-store", objectStoreResult.start, objectStoreResult.end, objectStoreResult.err)
		recordStartupFailureAndRefresh(deps, ownedEnv, failureSummary(suiteservices.ServicePostgres, stagePostgresStart, "start suite postgres", postgresResult.err))
		cleanupOwnedServices(deps, ownedEnv, postgresSvc, objectStoreSvc, "", "startup_failed", 1)
		return 1
	}

	schemaHash, err := pgschema.Hash()
	if err != nil {
		recordStartupFailureAndRefresh(deps, ownedEnv, failureSummary(suiteservices.ServicePostgres, stagePostgresTemplate, "hash postgres schema", err))
		cleanupOwnedServices(deps, ownedEnv, postgresSvc, objectStoreSvc, "", "startup_failed", 1)
		return 1
	}
	ownedEnv[suiteservices.PGSchemaHashEnv] = schemaHash
	templateDB := templateDatabaseName(suiteID, schemaHash)
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
	err = withSuiteGooseLog(ownedEnv, func() error {
		return deps.createTemplate(templateCtx, postgresSvc.adminDSN, templateDB)
	})
	recordTimingSpanStatus(deps, ownedEnv, bucketMigration, "test-services prepare postgres template database", templateStart, err)
	cancelTemplate()
	if err != nil {
		cancelStartup()
		objectStoreResult := <-objectStoreResultCh
		objectStoreSvc = objectStoreResult.service
		recordTimingSpanStatusAt(deps, ownedEnv, bucketServiceWait, "test-services start object-store", objectStoreResult.start, objectStoreResult.end, objectStoreResult.err)
		recordStartupFailureAndRefresh(deps, ownedEnv, failureSummary(suiteservices.ServicePostgres, stagePostgresTemplate, "prepare postgres template database", err))
		cleanupOwnedServices(deps, ownedEnv, postgresSvc, objectStoreSvc, "", "startup_failed", 1)
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

	objectStoreResult := <-objectStoreResultCh
	objectStoreSvc = objectStoreResult.service
	recordTimingSpanStatusAt(deps, ownedEnv, bucketServiceWait, "test-services start object-store", objectStoreResult.start, objectStoreResult.end, objectStoreResult.err)
	if objectStoreResult.err != nil {
		recordStartupFailureAndRefresh(deps, ownedEnv, failureSummary(suiteservices.ServiceObjectStore, stageObjectStoreStart, "start suite object-store", objectStoreResult.err))
		cleanupOwnedServices(deps, ownedEnv, postgresSvc, objectStoreSvc, "", "startup_failed", 1)
		return 1
	}
	deps.recordEvent(ownedEnv, suiteservices.Event{
		Type:    suiteservices.EventServiceStarted,
		Service: suiteservices.ServiceObjectStore,
		Details: map[string]any{
			"endpoint": objectStoreSvc.endpoint,
			"secure":   objectStoreSvc.secure,
		},
	})
	deps.refreshSummary(ownedEnv)

	generatedLeasePath, err := writeServiceLease(ownedEnv, postgresSvc, objectStoreSvc)
	if err != nil {
		recordStartupFailureAndRefresh(deps, ownedEnv, failureSummary("", stageCleanupLease, "write suite service lease", err))
		cleanupOwnedServices(deps, ownedEnv, postgresSvc, objectStoreSvc, "", "startup_failed", 1)
		return 1
	}
	if err := copyFile(generatedLeasePath, leaseFile, 0o600); err != nil {
		recordStartupFailureAndRefresh(deps, ownedEnv, failureSummary("", stageCleanupLease, "copy suite service lease", err))
		cleanupOwnedServices(deps, ownedEnv, postgresSvc, objectStoreSvc, generatedLeasePath, "startup_failed", 1)
		return 1
	}

	janitorCtx, cancelJanitor := context.WithTimeout(context.Background(), staleFixtureJanitorTimeout)
	janitorStart := time.Now().UTC()
	err = cleanupStaleWebE2EFixtures(janitorCtx, deps, serviceBackedCleanupEnv(ownedEnv, postgresSvc, objectStoreSvc))
	recordTimingSpanStatus(deps, ownedEnv, bucketSetup, "test-services janitor stale browser fixtures", janitorStart, err)
	cancelJanitor()
	if err != nil {
		recordStartupFailureAndRefresh(deps, ownedEnv, failureSummary("", stageCleanupJanitor, "janitor stale browser e2e fixtures", err))
		cleanupOwnedServices(deps, ownedEnv, postgresSvc, objectStoreSvc, generatedLeasePath, "startup_failed", 1)
		return 1
	}
	if err := suiteservices.RecordLifecycleEvent(ownedEnv, suiteservices.LifecycleEventReadinessPassed, ""); err != nil {
		recordFailureAndRefresh(deps, ownedEnv, failureSummary("", stageChildStart, "record service readiness lifecycle", err))
		cleanupOwnedServices(deps, ownedEnv, postgresSvc, objectStoreSvc, generatedLeasePath, "startup_failed", 1)
		return 1
	}

	childEnv := cloneEnv(ownedEnv)
	childEnv[suiteservices.PGAdminDSNEnv] = postgresSvc.adminDSN
	childEnv[suiteservices.PGDSNTemplateEnv] = postgresSvc.dsnTemplate
	childEnv[suiteservices.PGTemplateDBEnv] = templateDB
	childEnv[suiteservices.PGSchemaHashEnv] = schemaHash
	childEnv[suiteservices.S3EndpointEnv] = objectStoreSvc.endpoint
	childEnv[suiteservices.S3AccessKeyEnv] = objectStoreSvc.accessKey
	childEnv[suiteservices.S3SecretKeyEnv] = objectStoreSvc.secretKey
	childEnv[suiteservices.S3SecureEnv] = fmt.Sprintf("%t", objectStoreSvc.secure)
	if err := writeEnvJSON(envFile, childEnv); err != nil {
		recordFailureAndRefresh(deps, ownedEnv, failureSummary("", stageChildStart, "write suite child environment", err))
		cleanupOwnedServices(deps, ownedEnv, postgresSvc, objectStoreSvc, generatedLeasePath, "startup_failed", 1)
		return 1
	}
	deps.refreshSummary(ownedEnv)
	return 0
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

func runRecordLifecycle(args []string, env map[string]string) int {
	envFile, event, childKey, err := parseRecordLifecycleArgs(args)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 2
	}
	lifecycleEnv := cloneEnv(env)
	if envFile != "" {
		fileEnv, err := readEnvJSON(envFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "read lifecycle env file: %v\n", err)
			return 1
		}
		for key, value := range fileEnv {
			lifecycleEnv[key] = value
		}
	}
	if err := suiteservices.RecordLifecycleEvent(lifecycleEnv, event, childKey); err != nil {
		fmt.Fprintf(os.Stderr, "record lifecycle event: %v\n", err)
		return 1
	}
	return 0
}

func recordLifecycleEventIfPresent(env map[string]string, event string, childKey string) error {
	return recordLifecycleEventIfPresentWithFailure(env, event, childKey, "", "")
}

func recordLifecycleFailureEventIfPresent(env map[string]string, event string, childKey string, failureClass string, failureReason string) error {
	return recordLifecycleEventIfPresentWithFailure(env, event, childKey, failureClass, failureReason)
}

func recordLifecycleEventIfPresentWithFailure(env map[string]string, event string, childKey string, failureClass string, failureReason string) error {
	path, ok, err := suiteservices.LifecycleEventsPath(env)
	if err != nil || !ok {
		return err
	}
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	if strings.TrimSpace(failureClass) != "" || strings.TrimSpace(failureReason) != "" {
		return suiteservices.RecordLifecycleFailureEvent(env, event, childKey, failureClass, failureReason)
	}
	return suiteservices.RecordLifecycleEvent(env, event, childKey)
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

func runTerminateSuite(args []string, env map[string]string, deps dependencies) int {
	leasePath, err := parseTerminateSuiteArgs(args)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 2
	}
	lease, err := readServiceLease(leasePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "read suite service lease: %v\n", err)
		return 1
	}

	leaseEnv := cloneEnv(env)
	leaseEnv[suiteservices.ActiveEnv] = "1"
	leaseEnv[suiteservices.SuiteIDEnv] = lease.SuiteID
	leaseEnv[suiteservices.LifecycleModeEnv] = "owned"
	if lease.Target != "" {
		leaseEnv[suiteservices.TargetEnv] = lease.Target
	}
	if lease.RunID != "" {
		leaseEnv["CARTULARY_TEST_RUN_ID"] = lease.RunID
	}

	if err := recordLifecycleEventIfPresent(leaseEnv, suiteservices.LifecycleEventCleanupStarted, ""); err != nil {
		printSuiteFailure(leaseEnv, failureSummary("", stageCleanupReaper, "record cleanup lifecycle start", err))
	}
	status := 0
	for _, result := range terminateSuiteServices(context.Background(), lease) {
		recordTimingSpanStatusAt(deps, leaseEnv, bucketTeardown, result.label, result.start, result.end, result.err, true)
		if result.err != nil {
			status = 1
			printSuiteFailure(leaseEnv, failureSummary(result.service, result.stage, result.label, result.err))
		}
	}
	deps.refreshSummary(leaseEnv)
	if status == 0 {
		if err := recordLifecycleEventIfPresent(leaseEnv, suiteservices.LifecycleEventCleanupSucceeded, ""); err != nil {
			printSuiteFailure(leaseEnv, failureSummary("", stageCleanupReaper, "record cleanup lifecycle success", err))
			return 1
		}
	} else if err := recordLifecycleFailureEventIfPresent(leaseEnv, suiteservices.LifecycleEventCleanupFailed, "", suiteservices.FailureClassHelper, "cleanup_error"); err != nil {
		printSuiteFailure(leaseEnv, failureSummary("", stageCleanupReaper, "record cleanup lifecycle failure", err))
		return 1
	}
	return status
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

func parseRecordLifecycleArgs(args []string) (string, string, string, error) {
	values, err := parseFlagPairs(args, map[string]struct{}{
		"--env-file":  {},
		"--event":     {},
		"--child-key": {},
	})
	if err != nil {
		return "", "", "", err
	}
	event := strings.TrimSpace(values["--event"])
	if event == "" {
		return "", "", "", errors.New("usage: testservices record-lifecycle --env-file <path> --event <event> [--child-key <key>]")
	}
	return strings.TrimSpace(values["--env-file"]), event, strings.TrimSpace(values["--child-key"]), nil
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

func parseStartSuiteArgs(args []string) (string, string, error) {
	values, err := parseFlagPairs(args, map[string]struct{}{
		"--env-file":   {},
		"--lease-file": {},
	})
	if err != nil {
		return "", "", err
	}
	envFile := strings.TrimSpace(values["--env-file"])
	leaseFile := strings.TrimSpace(values["--lease-file"])
	if envFile == "" || leaseFile == "" {
		return "", "", errors.New("usage: testservices start-suite --env-file <path> --lease-file <path>")
	}
	return envFile, leaseFile, nil
}

func parseTerminateSuiteArgs(args []string) (string, error) {
	values, err := parseFlagPairs(args, map[string]struct{}{
		"--lease": {},
	})
	if err != nil {
		return "", err
	}
	leasePath := strings.TrimSpace(values["--lease"])
	if leasePath == "" {
		return "", errors.New("usage: testservices terminate-suite --lease <path>")
	}
	return leasePath, nil
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
		startObjectStore:    startObjectStoreService,
		startChild:          startChildProcess,
		startReaper:         startDetachedSuiteReaper,
		preflightSuite:      runSuiteStartupPreflight,
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

func disableTestcontainersRyukForSuiteStartup() func() {
	const key = "TESTCONTAINERS_RYUK_DISABLED"
	previous, hadPrevious := os.LookupEnv(key)
	_ = os.Setenv(key, "true")
	return func() {
		if hadPrevious {
			_ = os.Setenv(key, previous)
			return
		}
		_ = os.Unsetenv(key)
	}
}

func runConfiguredSuitePreflight(ctx context.Context, deps dependencies, env map[string]string) (suitePreflightResult, error) {
	if deps.preflightSuite == nil {
		return suitePreflightResult{}, nil
	}
	return deps.preflightSuite(ctx, env)
}

func runSuiteStartupPreflight(ctx context.Context, env map[string]string) (suitePreflightResult, error) {
	result := suitePreflightResult{}
	cli, err := dockerclient.New(dockerclient.FromEnv)
	if err != nil {
		return result, fmt.Errorf("create docker client: %w", err)
	}
	defer cli.Close()

	result.DockerEndpoint = cli.DaemonHost()
	if _, err := cli.Ping(ctx, dockerclient.PingOptions{NegotiateAPIVersion: true}); err != nil {
		return result, fmt.Errorf("ping docker endpoint %s: %w", result.DockerEndpoint, err)
	}
	result.DockerOK = true

	if err := verifySuiteReaperArtifactPath(env); err != nil {
		return result, err
	}
	result.ReaperReady = true

	cleanupSummary, err := cleanupPreviousSuiteServiceContainers(ctx, cli, env, time.Now().UTC())
	result.StaleContainersScanned = cleanupSummary.Scanned
	result.StaleContainersRemoved = cleanupSummary.Removed
	result.StaleContainersDeferred = cleanupSummary.Deferred
	if err != nil {
		return result, err
	}
	return result, nil
}

func verifySuiteReaperArtifactPath(env map[string]string) error {
	suiteDir, ok, err := suiteservices.ResolveSuiteArtifactDir(env)
	if err != nil {
		return err
	}
	if !ok {
		return errors.New("suite reaper preflight requires an active suite artifact directory")
	}
	if err := os.MkdirAll(suiteDir, 0o700); err != nil {
		return fmt.Errorf("create suite service artifact dir: %w", err)
	}
	logPath := filepath.Join(suiteDir, "service-reaper.log")
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600) // #nosec G304 -- the reaper log is colocated with the resolved suite artifact path.
	if err != nil {
		return fmt.Errorf("open service reaper log: %w", err)
	}
	if err := logFile.Close(); err != nil {
		return fmt.Errorf("close service reaper log: %w", err)
	}
	return nil
}

func suiteServiceLabels(env map[string]string, service string) map[string]string {
	return map[string]string{
		testServiceLabelManaged: testServiceManagedValue,
		testServiceLabelSuiteID: suiteservices.SuiteID(env),
		testServiceLabelRunID:   suiteservices.ResolveRunID(env),
		testServiceLabelTarget:  suiteservices.LookupEnvValue(env, suiteservices.TargetEnv),
		testServiceLabelService: service,
	}
}

func writeServiceLease(env map[string]string, postgresSvc postgresService, objectStoreSvc objectStoreService) (string, error) {
	suiteDir, ok, err := suiteservices.ResolveSuiteArtifactDir(env)
	if err != nil {
		return "", err
	}
	if !ok {
		return "", errors.New("suite service lease requires an active suite artifact directory")
	}
	if err := os.MkdirAll(suiteDir, 0o700); err != nil {
		return "", fmt.Errorf("create suite service artifact dir: %w", err)
	}

	resultRoot, err := suiteservices.ResolveResultsRoot(env)
	if err != nil {
		return "", err
	}
	runID := suiteservices.ResolveRunID(env)
	runRoot := filepath.Join(resultRoot, runID)
	leaseID, err := generateLeaseID()
	if err != nil {
		return "", err
	}
	proofLabels := suiteServiceLabels(env, "")
	delete(proofLabels, testServiceLabelService)
	now := time.Now().UTC()
	lease := serviceLease{
		SchemaID:      "cartulary.test_services.lease.v1",
		LeaseID:       leaseID,
		SuiteID:       suiteservices.SuiteID(env),
		RunID:         runID,
		ResultRoot:    resultRoot,
		RunRoot:       runRoot,
		Target:        suiteservices.LookupEnvValue(env, suiteservices.TargetEnv),
		Mode:          "owned",
		OwnershipMode: "owned",
		OwnerPID:      os.Getpid(),
		CreatedAt:     now.Format(time.RFC3339Nano),
		ExpiresAt:     now.Add(24 * time.Hour).Format(time.RFC3339Nano),
		Resources: []serviceLeaseResource{
			{
				Kind:        "container",
				Service:     suiteservices.ServicePostgres,
				ContainerID: postgresSvc.containerID,
				Name:        postgresSvc.name,
				Image:       postgresSvc.image,
				Labels:      cloneStringMap(postgresSvc.labels),
			},
			{
				Kind:        "container",
				Service:     suiteservices.ServiceObjectStore,
				ContainerID: objectStoreSvc.containerID,
				Name:        objectStoreSvc.name,
				Image:       objectStoreSvc.image,
				Labels:      cloneStringMap(objectStoreSvc.labels),
			},
		},
		ProofLabels: proofLabels,
		ProofPrefixes: map[string]string{
			"database": "ct_",
			"bucket":   "ct-",
		},
		CleanupState: "not_started",
	}
	payload, err := json.MarshalIndent(lease, "", "  ")
	if err != nil {
		return "", fmt.Errorf("encode suite service lease: %w", err)
	}
	leasePath := filepath.Join(suiteDir, "service-lease.json")
	if err := writeFileAtomic(leasePath, append(payload, '\n'), 0o600); err != nil {
		return "", fmt.Errorf("write suite service lease: %w", err)
	}
	return leasePath, nil
}

func readServiceLease(path string) (serviceLease, error) {
	raw, err := os.ReadFile(path) // #nosec G304 -- lease paths are produced by ResolveSuiteArtifactDir or explicit terminate-suite CLI input.
	if err != nil {
		return serviceLease{}, err
	}
	var lease serviceLease
	if err := json.Unmarshal(raw, &lease); err != nil {
		return serviceLease{}, err
	}
	if lease.SchemaID != "cartulary.test_services.lease.v1" {
		return serviceLease{}, fmt.Errorf("unsupported suite service lease schema %q", lease.SchemaID)
	}
	if strings.TrimSpace(lease.SuiteID) == "" {
		return serviceLease{}, errors.New("suite service lease missing suite_id")
	}
	if strings.TrimSpace(lease.LeaseID) == "" {
		return serviceLease{}, errors.New("suite service lease missing lease_id")
	}
	if strings.TrimSpace(lease.OwnershipMode) == "" {
		lease.OwnershipMode = lease.Mode
	}
	if strings.TrimSpace(lease.OwnershipMode) == "" {
		return serviceLease{}, errors.New("suite service lease missing ownership_mode")
	}
	if lease.OwnershipMode == "owned" && len(lease.Resources) == 0 {
		return serviceLease{}, errors.New("owned suite service lease missing resources")
	}
	return lease, nil
}

func cloneStringMap(values map[string]string) map[string]string {
	if len(values) == 0 {
		return nil
	}
	cloned := make(map[string]string, len(values))
	for key, value := range values {
		cloned[key] = value
	}
	return cloned
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

func startPostgresService(ctx context.Context, env map[string]string) (postgresService, error) {
	labels := suiteServiceLabels(env, suiteservices.ServicePostgres)
	harness, err := startPostgresHarnessWithOptions(ctx, pgtest.StartOptions{
		Labels:         labels,
		Observer:       suiteServiceStartObserver(env, suiteservices.ServicePostgres),
		AttemptTimeout: suitePostgresAttemptLimit,
	})
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
		host:        harness.Host,
		port:        harness.Port,
		user:        harness.User,
		close:       harness.Close,
		containerID: harness.ContainerID(),
		name:        containerName(ctx, harness.Container),
		image:       pgtest.ContainerImage(),
		labels:      labels,
	}, nil
}

func startObjectStoreService(ctx context.Context, env map[string]string) (objectStoreService, error) {
	labels := suiteServiceLabels(env, suiteservices.ServiceObjectStore)
	harness, err := startObjectStoreHarnessWithOptions(ctx, s3test.StartOptions{
		Labels:         labels,
		Observer:       suiteServiceStartObserver(env, suiteservices.ServiceObjectStore),
		AttemptTimeout: suiteObjectStoreAttemptLimit,
	})
	if err != nil {
		return objectStoreService{}, err
	}
	return objectStoreService{
		endpoint:    harness.Endpoint,
		accessKey:   harness.AccessKey,
		secretKey:   harness.SecretKey,
		secure:      harness.Secure,
		close:       harness.Close,
		containerID: harness.ContainerID(),
		name:        containerName(ctx, harness.Container),
		image:       s3test.ContainerImage(),
		labels:      labels,
	}, nil
}

func suiteServiceStartObserver(env map[string]string, service string) testcontainersx.StartObserver {
	return func(event testcontainersx.StartEvent) {
		if event.Type != testcontainersx.StartEventAttemptEnd {
			return
		}
		recordServiceStartupAttempt(env, service, event)
	}
}

func recordServiceStartupAttempt(env map[string]string, service string, event testcontainersx.StartEvent) {
	start := event.StartTime
	end := event.EndTime
	if start.IsZero() {
		start = time.Now().UTC()
	}
	if end.IsZero() || end.Before(start) {
		end = start
	}
	duration := event.Duration
	if duration <= 0 {
		duration = end.Sub(start)
	}
	if duration < 0 {
		duration = 0
	}

	status := event.Status
	if strings.TrimSpace(status) == "" {
		status = "pass"
		if event.Err != nil {
			status = "fail"
		}
	}
	label := fmt.Sprintf("test-services start %s attempt %d", service, event.Attempt)
	details := map[string]any{
		"target":                   suiteservices.LookupEnvValue(env, suiteservices.TargetEnv),
		"bucket":                   bucketServiceWait,
		"label":                    label,
		"start_time":               start.Format(time.RFC3339Nano),
		"end_time":                 end.Format(time.RFC3339Nano),
		"duration_ms":              duration.Milliseconds(),
		"status":                   status,
		"service":                  service,
		"startup_attempt":          true,
		"attempt":                  event.Attempt,
		"max_attempts":             event.MaxAttempts,
		"retryable":                event.Retryable,
		"retry_scheduled":          status == "fail" && event.Retryable && event.Attempt < event.MaxAttempts && !event.RetryBlockedByContext,
		"retry_blocked_by_context": event.RetryBlockedByContext,
	}
	if event.Err != nil {
		details["message"] = suiteservices.SanitizeDiagnosticText(event.Err.Error())
	}
	_ = suiteservices.RecordEvent(env, suiteservices.Event{
		Type:    suiteservices.EventTimingSpan,
		Service: service,
		Status:  status,
		Details: details,
	})
}

type namedContainer interface {
	Name(context.Context) (string, error)
}

func containerName(ctx context.Context, container namedContainer) string {
	if container == nil {
		return ""
	}
	name, err := container.Name(ctx)
	if err != nil {
		return ""
	}
	return strings.TrimPrefix(strings.TrimSpace(name), "/")
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
	if _, err := postgres.Migrate(ctx, db, dbmigrations.Source(), "up"); err != nil {
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

func withSuiteGooseLog(env map[string]string, run func() error) error {
	suiteDir, ok, err := suiteservices.ResolveSuiteArtifactDir(env)
	if err != nil {
		return err
	}
	if !ok {
		return run()
	}
	if err := os.MkdirAll(suiteDir, 0o700); err != nil {
		return fmt.Errorf("create suite artifact dir: %w", err)
	}
	logPath := filepath.Join(suiteDir, "goose.log")
	previous, hadPrevious := os.LookupEnv(postgres.GooseLogFileEnv)
	if err := os.Setenv(postgres.GooseLogFileEnv, logPath); err != nil {
		return fmt.Errorf("set goose log file env: %w", err)
	}
	defer func() {
		if hadPrevious {
			_ = os.Setenv(postgres.GooseLogFileEnv, previous)
			return
		}
		_ = os.Unsetenv(postgres.GooseLogFileEnv)
	}()
	return run()
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
	if schemaHash := strings.TrimSpace(suiteservices.LookupEnvValue(env, suiteservices.PGSchemaHashEnv)); schemaHash != "" {
		details["schema_hash"] = schemaHash
	}
	if strategy == suiteservices.PostgresPreparationTemplate {
		details["fixture_class"] = "suite_template"
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
		return webE2EFixture{}, fmt.Errorf("attach suite object-store: %w", err)
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
		return fmt.Errorf("attach suite object-store: %w", err)
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
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	lines := []string{
		"# Generated by cartulary-test-services prepare-web-e2e.",
		shellExport(mustPostgresServiceRefEnvKey("postgres_primary"), fixture.DSN),
		shellExport(mustObjectStoreServiceRefEnvKeys("object_primary").Endpoint, fixture.S3Endpoint),
		shellExport(mustObjectStoreServiceRefEnvKeys("object_primary").AccessKey, fixture.S3AccessKey),
		shellExport(mustObjectStoreServiceRefEnvKeys("object_primary").SecretKey, fixture.S3SecretKey),
		shellExport(mustObjectStoreServiceRefEnvKeys("object_primary").Secure, fmt.Sprintf("%t", fixture.S3Secure)),
		shellExport(mustObjectStoreServiceRefEnvKeys("object_primary").Bucket, fixture.Bucket),
		"",
	}
	return os.WriteFile(path, []byte(strings.Join(lines, "\n")), 0o600)
}

func mustPostgresServiceRefEnvKey(serviceRef string) string {
	key, err := postgres.EnvKeyForServiceRef(serviceRef)
	if err != nil {
		panic(err)
	}
	return key
}

func mustObjectStoreServiceRefEnvKeys(serviceRef string) objectstore.ServiceRefEnvKeys {
	keys, err := objectstore.EnvKeysForServiceRef(serviceRef)
	if err != nil {
		panic(err)
	}
	return keys
}

func shellExport(key string, value string) string {
	return fmt.Sprintf("export %s=%s", key, shellQuote(value))
}

func shellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\"'\"'") + "'"
}

func writeWebE2EMetadata(path string, metadata webE2EMetadata) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	payload, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(payload, '\n'), 0o600)
}

func readWebE2EMetadata(path string) (webE2EMetadata, error) {
	raw, err := os.ReadFile(path) // #nosec G304 -- web E2E metadata paths are generated under the suite artifact directory.
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

func recordServiceReaperScheduled(deps dependencies, env map[string]string, leasePath string) {
	deps.recordEvent(env, suiteservices.Event{
		Type:   "service-reaper-scheduled",
		Status: "pass",
		Details: map[string]any{
			"target":     suiteservices.LookupEnvValue(env, suiteservices.TargetEnv),
			"lease_path": leasePath,
		},
	})
}

func recordSuitePreflight(deps dependencies, env map[string]string, result suitePreflightResult, err error) {
	details := map[string]any{
		"target":                          suiteservices.LookupEnvValue(env, suiteservices.TargetEnv),
		"docker_endpoint":                 result.DockerEndpoint,
		"docker_ok":                       result.DockerOK,
		"reaper_ready":                    result.ReaperReady,
		"stale_containers_scanned":        result.StaleContainersScanned,
		"stale_containers_removed":        result.StaleContainersRemoved,
		"stale_containers_deferred":       result.StaleContainersDeferred,
		"ryuk_disabled_for_suite_startup": result.RyukDisabledForSuiteStartup,
	}
	status := "pass"
	if err != nil {
		status = "fail"
		details["failure_class"] = suiteservices.FailureClassInfra
		details["failure_reason"] = "preflight_error"
		details["message"] = suiteservices.SanitizeDiagnosticText(err.Error())
	}
	deps.recordEvent(env, suiteservices.Event{
		Type:    suiteservices.EventSuitePreflight,
		Status:  status,
		Details: details,
	})
}

func recordTimingSpanStatus(deps dependencies, env map[string]string, bucket string, label string, start time.Time, err error) {
	status := "pass"
	if err != nil {
		status = "fail"
	}
	recordTimingSpan(deps, env, bucket, label, start, time.Now().UTC(), status)
}

func recordTimingSpanStatusAt(deps dependencies, env map[string]string, bucket string, label string, start time.Time, end time.Time, err error, janitorial ...bool) {
	status := "pass"
	if err != nil {
		status = "fail"
	}
	recordTimingSpan(deps, env, bucket, label, start, end, status, janitorial...)
}

func recordTimingSpan(deps dependencies, env map[string]string, bucket string, label string, start time.Time, end time.Time, status string, janitorial ...bool) {
	if end.Before(start) {
		end = start
	}
	duration := end.Sub(start)
	if duration < 0 {
		duration = 0
	}
	details := map[string]any{
		"target":      suiteservices.LookupEnvValue(env, suiteservices.TargetEnv),
		"bucket":      bucket,
		"label":       label,
		"start_time":  start.Format(time.RFC3339Nano),
		"end_time":    end.Format(time.RFC3339Nano),
		"duration_ms": duration.Milliseconds(),
		"status":      status,
	}
	if len(janitorial) > 0 && janitorial[0] {
		details["janitorial"] = true
	}
	deps.recordEvent(env, suiteservices.Event{
		Type:    suiteservices.EventTimingSpan,
		Status:  status,
		Details: details,
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
	servicesRoot, err := resolveTestServicesRoot(env)
	if err != nil {
		return nil, err
	}
	servicesRootFS, err := os.OpenRoot(servicesRoot)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	defer servicesRootFS.Close()

	completed := make(map[string]struct{})
	retired := make(map[string]webE2EMetadata)
	err = filepath.WalkDir(servicesRoot, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || entry.Name() != "service-scope.json" {
			return nil
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		if !info.Mode().IsRegular() {
			return nil
		}
		if !pathWithinRoot(servicesRoot, path) {
			return fmt.Errorf("service scope path %s escapes %s", path, servicesRoot)
		}
		relativeScopePath, err := filepath.Rel(servicesRoot, path)
		if err != nil {
			return err
		}
		raw, err := servicesRootFS.ReadFile(relativeScopePath)
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

func resolveTestServicesRoot(env map[string]string) (string, error) {
	resultsRoot, err := suiteservices.ResolveResultsRoot(env)
	if err != nil {
		return "", err
	}
	servicesRoot := filepath.Join(resultsRoot, suiteservices.ResolveRunID(env), "_shared", "test-services")
	if !pathWithinRoot(resultsRoot, servicesRoot) {
		return "", fmt.Errorf("test services root %s escapes results root %s", servicesRoot, resultsRoot)
	}
	return servicesRoot, nil
}

func pathWithinRoot(root string, target string) bool {
	cleanRoot := filepath.Clean(root)
	cleanTarget := filepath.Clean(target)
	relative, err := filepath.Rel(cleanRoot, cleanTarget)
	if err != nil {
		return false
	}
	return relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator))
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

func cleanupOwnedServices(deps dependencies, env map[string]string, postgresSvc postgresService, objectStoreSvc objectStoreService, leasePath string, status string, childExitCode int) {
	cleanupCtx, cancel := context.WithTimeout(context.Background(), cleanupTimeout)
	defer cancel()

	if err := recordLifecycleEventIfPresent(env, suiteservices.LifecycleEventCleanupStarted, ""); err != nil {
		printSuiteFailure(env, failureSummary("", stageCleanupReaper, "record cleanup lifecycle start", err))
	}
	cleanupStatus := status
	if cleanupStatus == "" {
		cleanupStatus = "succeeded"
	}
	browserCleanupFailed := false

	cleanupEnv := serviceBackedCleanupEnv(env, postgresSvc, objectStoreSvc)
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

	if strings.TrimSpace(leasePath) != "" {
		reaperStart := time.Now().UTC()
		err := deps.startReaper(leasePath, env)
		recordTimingSpanStatus(deps, env, bucketTeardown, "test-services schedule service reaper", reaperStart, err)
		if err != nil {
			cleanupStatus = "cleanup_failed"
			printSuiteFailure(env, failureSummary("", stageCleanupReaper, "schedule suite service reaper", err))
			for _, result := range cleanupServices(cleanupCtx, postgresSvc, objectStoreSvc) {
				recordTimingSpanStatusAt(deps, env, bucketTeardown, result.label, result.start, result.end, result.err)
				if result.err != nil {
					printSuiteFailure(env, failureSummary(result.service, result.stage, result.label, result.err))
				}
			}
		} else {
			recordServiceReaperScheduled(deps, env, leasePath)
		}
	} else {
		for _, result := range cleanupServices(cleanupCtx, postgresSvc, objectStoreSvc) {
			recordTimingSpanStatusAt(deps, env, bucketTeardown, result.label, result.start, result.end, result.err)
			if result.err != nil {
				cleanupStatus = "cleanup_failed"
				printSuiteFailure(env, failureSummary(result.service, result.stage, result.label, result.err))
			}
		}
	}

	recordCleanupAndRefresh(deps, env, cleanupStatus, childExitCode)
	if cleanupStatus == "cleanup_failed" {
		if err := recordLifecycleFailureEventIfPresent(env, suiteservices.LifecycleEventCleanupFailed, "", suiteservices.FailureClassHelper, "cleanup_error"); err != nil {
			printSuiteFailure(env, failureSummary("", stageCleanupReaper, "record cleanup lifecycle failure", err))
		}
	} else {
		if err := recordLifecycleEventIfPresent(env, suiteservices.LifecycleEventCleanupSucceeded, ""); err != nil {
			printSuiteFailure(env, failureSummary("", stageCleanupReaper, "record cleanup lifecycle success", err))
		}
	}
}

func cleanupServices(ctx context.Context, postgresSvc postgresService, objectStoreSvc objectStoreService) []serviceCleanupResult {
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
	if objectStoreSvc.close != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			start := time.Now().UTC()
			err := objectStoreSvc.close(ctx)
			results <- serviceCleanupResult{
				service: suiteservices.ServiceObjectStore,
				stage:   stageCleanupObjectStore,
				label:   "test-services terminate object-store",
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

func terminateSuiteServices(parent context.Context, lease serviceLease) []serviceCleanupResult {
	if lease.OwnershipMode == "attach" {
		return nil
	}
	ctx, cancel := context.WithTimeout(parent, cleanupTimeout)
	defer cancel()

	cli, err := dockerclient.New(dockerclient.FromEnv)
	if err != nil {
		results := make([]serviceCleanupResult, 0, len(lease.Resources))
		for _, service := range lease.Resources {
			results = append(results, serviceCleanupResult{
				service: service.Service,
				stage:   cleanupStageForService(service.Service),
				label:   cleanupLabelForService(service.Service),
				start:   time.Now().UTC(),
				end:     time.Now().UTC(),
				err:     fmt.Errorf("create docker client: %w", err),
			})
		}
		return results
	}
	defer cli.Close()

	results := make([]serviceCleanupResult, 0, len(lease.Resources))
	for _, service := range lease.Resources {
		start := time.Now().UTC()
		err := terminateLeasedService(ctx, cli, lease, service)
		results = append(results, serviceCleanupResult{
			service: service.Service,
			stage:   cleanupStageForService(service.Service),
			label:   cleanupLabelForService(service.Service),
			start:   start,
			end:     time.Now().UTC(),
			err:     err,
		})
	}
	slices.SortFunc(results, func(left, right serviceCleanupResult) int {
		return strings.Compare(left.label, right.label)
	})
	return results
}

func terminateLeasedService(ctx context.Context, cli *dockerclient.Client, lease serviceLease, service serviceLeaseResource) error {
	containers, err := leasedServiceContainers(ctx, cli, lease, service)
	if err != nil {
		return err
	}
	if len(containers) == 0 {
		return nil
	}
	var errs []error
	for _, containerID := range containers {
		_, err := cli.ContainerRemove(ctx, containerID, dockerclient.ContainerRemoveOptions{
			RemoveVolumes: true,
			Force:         true,
		})
		if err != nil && !isDockerNotFound(err) {
			errs = append(errs, fmt.Errorf("remove container %s: %w", shortContainerID(containerID), err))
		}
	}
	return errors.Join(errs...)
}

func leasedServiceContainers(ctx context.Context, cli *dockerclient.Client, lease serviceLease, service serviceLeaseResource) ([]string, error) {
	filters := dockerclient.Filters{}
	for key, value := range requiredLeaseLabels(lease, service) {
		if strings.TrimSpace(value) == "" {
			continue
		}
		filters = filters.Add("label", fmt.Sprintf("%s=%s", key, value))
	}
	result, err := cli.ContainerList(ctx, dockerclient.ContainerListOptions{
		All:     true,
		Filters: filters,
	})
	if err != nil {
		return nil, fmt.Errorf("list leased %s containers: %w", service.Service, err)
	}
	ids := make([]string, 0, len(result.Items))
	seen := make(map[string]struct{})
	for _, item := range result.Items {
		if service.ContainerID != "" && item.ID != service.ContainerID && !strings.HasPrefix(item.ID, service.ContainerID) && !strings.HasPrefix(service.ContainerID, item.ID) {
			continue
		}
		if _, ok := seen[item.ID]; ok {
			continue
		}
		seen[item.ID] = struct{}{}
		ids = append(ids, item.ID)
	}
	slices.Sort(ids)
	return ids, nil
}

func cleanupPreviousSuiteServiceContainers(ctx context.Context, cli suiteContainerClient, env map[string]string, now time.Time) (staleSuiteContainerCleanupSummary, error) {
	filters := dockerclient.Filters{}.
		Add("label", fmt.Sprintf("%s=%s", testServiceLabelManaged, testServiceManagedValue))
	result, err := cli.ContainerList(ctx, dockerclient.ContainerListOptions{
		All:     true,
		Filters: filters,
	})
	if err != nil {
		return staleSuiteContainerCleanupSummary{}, fmt.Errorf("list managed suite service containers: %w", err)
	}

	summary := staleSuiteContainerCleanupSummary{Scanned: len(result.Items)}
	for _, item := range result.Items {
		if !previousSuiteContainerCleanupEligible(env, item, now) {
			continue
		}
		_, err := cli.ContainerRemove(ctx, item.ID, dockerclient.ContainerRemoveOptions{
			RemoveVolumes: true,
			Force:         true,
		})
		status, fatalErr := classifyStaleContainerRemove(cli, item.ID, err)
		switch status {
		case staleContainerCleanupRemoved, staleContainerCleanupNotFound:
			summary.Removed++
		case staleContainerCleanupRemovalInProgress:
			summary.Deferred++
		case staleContainerCleanupFatal:
			return summary, fatalErr
		}
	}
	return summary, nil
}

const (
	staleContainerCleanupRemoved           = "removed"
	staleContainerCleanupNotFound          = "not_found"
	staleContainerCleanupRemovalInProgress = "removal_in_progress"
	staleContainerCleanupFatal             = "fatal"
)

func classifyStaleContainerRemove(cli suiteContainerClient, containerID string, err error) (string, error) {
	if err == nil {
		return staleContainerCleanupRemoved, nil
	}
	if isDockerNotFound(err) {
		return staleContainerCleanupNotFound, nil
	}
	if isDockerRemovalInProgress(err) {
		return staleContainerCleanupRemovalInProgress, nil
	}
	if isDockerRemoveTimeout(err) {
		return classifyStaleContainerRemoveTimeout(cli, containerID, err)
	}
	return staleContainerCleanupFatal, fmt.Errorf("remove stale suite service container %s: %w", shortContainerID(containerID), err)
}

func classifyStaleContainerRemoveTimeout(cli suiteContainerClient, containerID string, removeErr error) (string, error) {
	recheckCtx, cancel := context.WithTimeout(context.Background(), staleSuiteContainerRecheck)
	defer cancel()

	result, inspectErr := cli.ContainerInspect(recheckCtx, containerID, dockerclient.ContainerInspectOptions{})
	if inspectErr != nil {
		if isDockerNotFound(inspectErr) {
			return staleContainerCleanupNotFound, nil
		}
		return staleContainerCleanupFatal, fmt.Errorf("remove stale suite service container %s timed out and recheck failed: %w", shortContainerID(containerID), inspectErr)
	}
	if inspectedContainerRemovingOrDead(result.Container) {
		return staleContainerCleanupRemovalInProgress, nil
	}
	state := "unknown"
	running := false
	dead := false
	if result.Container.State != nil {
		state = string(result.Container.State.Status)
		running = result.Container.State.Running
		dead = result.Container.State.Dead
	}
	return staleContainerCleanupFatal, fmt.Errorf("remove stale suite service container %s timed out and recheck found state=%s running=%t dead=%t: %w", shortContainerID(containerID), state, running, dead, removeErr)
}

func inspectedContainerRemovingOrDead(container dockercontainer.InspectResponse) bool {
	if container.State == nil {
		return false
	}
	state := string(container.State.Status)
	return state == "removing" || state == "dead" || container.State.Dead
}

func previousSuiteContainerCleanupEligible(env map[string]string, item dockercontainer.Summary, now time.Time) bool {
	labels := item.Labels
	if labels[testServiceLabelManaged] != testServiceManagedValue {
		return false
	}
	suiteID := strings.TrimSpace(labels[testServiceLabelSuiteID])
	runID := strings.TrimSpace(labels[testServiceLabelRunID])
	if suiteID == "" || runID == "" {
		return false
	}
	if suiteID == suiteservices.SuiteID(env) && runID == suiteservices.ResolveRunID(env) {
		return false
	}
	if suiteServiceCleanupRecorded(env, runID, suiteID) {
		return true
	}
	if item.Created <= 0 {
		return false
	}
	return !now.Before(time.Unix(item.Created, 0).UTC().Add(staleSuiteContainerAge))
}

func suiteServiceCleanupRecorded(env map[string]string, runID string, suiteID string) bool {
	if !safeArtifactComponent(runID) || !safeArtifactComponent(suiteID) {
		return false
	}
	resultsRoot, err := suiteservices.ResolveResultsRoot(env)
	if err != nil {
		return false
	}
	summaryPath := filepath.Join(resultsRoot, runID, "_shared", "test-services", suiteID, "service-scope.json")
	raw, err := os.ReadFile(summaryPath) // #nosec G304 -- run and suite identifiers come from Cartulary-owned Docker labels and stay below the configured results root.
	if err != nil {
		return false
	}
	var scope suiteservices.ServiceScope
	if err := json.Unmarshal(raw, &scope); err != nil {
		return false
	}
	return strings.TrimSpace(scope.Cleanup.CompletedAt) != ""
}

func safeArtifactComponent(value string) bool {
	trimmed := strings.TrimSpace(value)
	return trimmed != "" &&
		trimmed != "." &&
		trimmed != ".." &&
		!filepath.IsAbs(trimmed) &&
		!strings.Contains(trimmed, "/") &&
		!strings.Contains(trimmed, "\\") &&
		!strings.Contains(trimmed, string(filepath.Separator))
}

func requiredLeaseLabels(lease serviceLease, service serviceLeaseResource) map[string]string {
	labels := map[string]string{
		testServiceLabelManaged: testServiceManagedValue,
		testServiceLabelSuiteID: lease.SuiteID,
		testServiceLabelRunID:   lease.RunID,
		testServiceLabelService: service.Service,
	}
	for key, value := range service.Labels {
		if _, required := labels[key]; required {
			labels[key] = value
		}
	}
	return labels
}

func cleanupStageForService(service string) string {
	if service == suiteservices.ServiceObjectStore {
		return stageCleanupObjectStore
	}
	return stageCleanupPostgres
}

func cleanupLabelForService(service string) string {
	if service == suiteservices.ServiceObjectStore {
		return "test-services terminate object-store"
	}
	return "test-services terminate postgres"
}

func shortContainerID(containerID string) string {
	if len(containerID) <= 12 {
		return containerID
	}
	return containerID[:12]
}

func isDockerNotFound(err error) bool {
	if err == nil {
		return false
	}
	lower := strings.ToLower(err.Error())
	return strings.Contains(lower, "no such container") || strings.Contains(lower, "not found") || strings.Contains(lower, "404")
}

func isDockerRemovalInProgress(err error) bool {
	if err == nil {
		return false
	}
	lower := strings.ToLower(err.Error())
	return strings.Contains(lower, "removal of container") && strings.Contains(lower, "already in progress")
}

func isDockerRemoveTimeout(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
		return true
	}
	lower := strings.ToLower(err.Error())
	return strings.Contains(lower, "context deadline exceeded") ||
		strings.Contains(lower, "request canceled") ||
		strings.Contains(lower, "request cancelled") ||
		strings.Contains(lower, "client.timeout exceeded")
}

func serviceBackedCleanupEnv(env map[string]string, postgresSvc postgresService, objectStoreSvc objectStoreService) map[string]string {
	cleanupEnv := cloneEnv(env)
	if postgresSvc.adminDSN != "" {
		cleanupEnv[suiteservices.PGAdminDSNEnv] = postgresSvc.adminDSN
	}
	if objectStoreSvc.endpoint != "" {
		cleanupEnv[suiteservices.S3EndpointEnv] = objectStoreSvc.endpoint
	}
	if objectStoreSvc.accessKey != "" {
		cleanupEnv[suiteservices.S3AccessKeyEnv] = objectStoreSvc.accessKey
	}
	if objectStoreSvc.secretKey != "" {
		cleanupEnv[suiteservices.S3SecretKeyEnv] = objectStoreSvc.secretKey
	}
	cleanupEnv[suiteservices.S3SecureEnv] = fmt.Sprintf("%t", objectStoreSvc.secure)
	return cleanupEnv
}

func failureSummary(service string, stage string, operation string, err error) suiteservices.FailureSummary {
	failureReason := classifyFailureReason(stage, operation, err)
	failure := suiteservices.FailureSummary{
		FailureClass:  classifyFailureStage(service, stage, failureReason),
		FailureReason: failureReason,
		Service:       service,
		Stage:         stage,
		Operation:     operation,
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

func classifyFailureReason(stage string, operation string, err error) string {
	switch stage {
	case stageStartupPreflight:
		return "preflight_error"
	case stagePostgresStart, stageObjectStoreStart:
		if serviceReadinessTimeout(err) {
			return "service_readiness_timeout"
		}
		return "service_start_error"
	case stagePostgresTemplate:
		return "fixture_error"
	case stageChildStart:
		return "child_target_failure"
	case stageCleanupLease, stageCleanupReaper, stageCleanupWebE2E, stageCleanupJanitor, stageCleanupPostgres, stageCleanupObjectStore:
		return "cleanup_error"
	}
	if strings.Contains(strings.ToLower(operation), "fixture") {
		return "fixture_error"
	}
	if err != nil {
		return "unknown_failure"
	}
	return "unknown_failure"
}

func classifyFailureStage(service string, stage string, reason string) string {
	switch reason {
	case "preflight_error", "service_start_error", "service_readiness_timeout":
		return suiteservices.FailureClassInfra
	case "configuration_error":
		return "config"
	case "fixture_error", "child_target_failure", "cleanup_error", "scheduler_accounting_error":
		return suiteservices.FailureClassHelper
	case "artifact_error":
		return suiteservices.FailureClassArtifact
	case "timeout_failure":
		return "timing"
	}
	if strings.TrimSpace(service) != "" {
		return suiteservices.FailureClassInfra
	}
	return suiteservices.FailureClassHelper
}

func serviceReadinessTimeout(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}
	lower := strings.ToLower(err.Error())
	return strings.Contains(lower, "readiness") &&
		(strings.Contains(lower, "deadline") || strings.Contains(lower, "timeout") || strings.Contains(lower, "timed out"))
}

func recordFailureAndRefresh(deps dependencies, env map[string]string, failure suiteservices.FailureSummary) {
	deps.recordEvent(env, suiteservices.Event{
		Type:    suiteservices.EventFailureRecorded,
		Service: failure.Service,
		Details: map[string]any{
			"failure_class":            failure.FailureClass,
			"failure_reason":           failure.FailureReason,
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

func recordStartupFailureAndRefresh(deps dependencies, env map[string]string, failure suiteservices.FailureSummary) {
	recordFailureAndRefresh(deps, env, failure)
	_ = suiteservices.RecordLifecycleFailureEvent(env, suiteservices.LifecycleEventStartupFailed, "", failure.FailureClass, failure.FailureReason)
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
	fmt.Fprintf(os.Stderr, "suite failure failure_class=%s reason=%s %s: %s [artifacts: %s]\n", failure.FailureClass, failure.FailureReason, label, message, artifactDir)
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

func generateLeaseID() (string, error) {
	var data [16]byte
	if _, err := rand.Read(data[:]); err != nil {
		return "", fmt.Errorf("read random lease id bytes: %w", err)
	}
	return hex.EncodeToString(data[:]), nil
}

func templateDatabaseName(suiteID string, schemaHash string) string {
	return fmt.Sprintf("ct_tpl_%s_%s", suiteservices.ShortHash(suiteID, 8), pgschema.ShortHash(schemaHash))
}

func writeFileAtomic(path string, payload []byte, mode os.FileMode) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("create %s parent: %w", path, err)
	}
	temp, err := os.CreateTemp(dir, ".tmp-"+filepath.Base(path)+".*")
	if err != nil {
		return fmt.Errorf("create temp file for %s: %w", path, err)
	}
	tempPath := temp.Name()
	cleanup := true
	defer func() {
		if cleanup {
			_ = os.Remove(tempPath)
		}
	}()
	if err := temp.Chmod(mode); err != nil {
		_ = temp.Close()
		return fmt.Errorf("chmod temp file for %s: %w", path, err)
	}
	if _, err := temp.Write(payload); err != nil {
		_ = temp.Close()
		return fmt.Errorf("write temp file for %s: %w", path, err)
	}
	if err := temp.Close(); err != nil {
		return fmt.Errorf("close temp file for %s: %w", path, err)
	}
	if err := os.Rename(tempPath, path); err != nil {
		return fmt.Errorf("rename temp file for %s: %w", path, err)
	}
	cleanup = false
	return nil
}

func copyFile(source string, destination string, mode os.FileMode) error {
	payload, err := os.ReadFile(source) // #nosec G304 -- source path is produced by the suite-service lease writer.
	if err != nil {
		return fmt.Errorf("read %s: %w", source, err)
	}
	if err := os.MkdirAll(filepath.Dir(destination), 0o700); err != nil {
		return fmt.Errorf("create %s parent: %w", destination, err)
	}
	if err := os.WriteFile(destination, payload, mode); err != nil { // #nosec G306 -- caller supplies the intended private file mode.
		return fmt.Errorf("write %s: %w", destination, err)
	}
	return nil
}

func writeEnvJSON(path string, env map[string]string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("create %s parent: %w", path, err)
	}
	payload, err := json.MarshalIndent(env, "", "  ")
	if err != nil {
		return fmt.Errorf("encode suite env: %w", err)
	}
	if err := os.WriteFile(path, append(payload, '\n'), 0o600); err != nil {
		return fmt.Errorf("write suite env: %w", err)
	}
	return nil
}

func readEnvJSON(path string) (map[string]string, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var env map[string]string
	if err := json.Unmarshal(raw, &env); err != nil {
		return nil, err
	}
	if env == nil {
		env = map[string]string{}
	}
	return env, nil
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
	cmd     *exec.Cmd
	logFile *os.File
	logPath string
}

func startChildProcess(argv []string, env map[string]string) (childProcess, error) {
	cmd := exec.Command(argv[0], argv[1:]...)
	cmd.Env = envSlice(env)
	captureChild := suiteservices.LookupEnvValue(env, "CARTULARY_SUPPRESS_CHILD_SUCCESS") == "1"
	if captureChild {
		suiteDir, ok, err := suiteservices.ResolveSuiteArtifactDir(env)
		if err != nil {
			return nil, err
		}
		if ok {
			if err := os.MkdirAll(suiteDir, 0o700); err != nil {
				return nil, err
			}
			logPath := filepath.Join(suiteDir, "child-process.log")
			logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600) // #nosec G304 -- suite artifact path is resolved from the configured results root.
			if err != nil {
				return nil, err
			}
			cmd.Stdout = logFile
			cmd.Stderr = logFile
			cmd.Stdin = os.Stdin
			cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
			if err := cmd.Start(); err != nil {
				_ = logFile.Close()
				return nil, err
			}
			return &execChild{cmd: cmd, logFile: logFile, logPath: logPath}, nil
		}
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	return &execChild{cmd: cmd}, nil
}

func startDetachedSuiteReaper(leasePath string, env map[string]string) error {
	executable, err := os.Executable()
	if err != nil {
		return fmt.Errorf("resolve testservices executable: %w", err)
	}
	cmd := exec.Command(executable, "terminate-suite", "--lease", leasePath)
	cmd.Env = envSlice(env)
	cmd.Stdin = nil
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	logPath := filepath.Join(filepath.Dir(leasePath), "service-reaper.log")
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600) // #nosec G304 -- the reaper log is colocated with an already resolved suite lease path.
	if err != nil {
		return fmt.Errorf("open service reaper log: %w", err)
	}
	defer logFile.Close()
	cmd.Stdout = logFile
	cmd.Stderr = logFile
	if err := cmd.Start(); err != nil {
		return err
	}
	return nil
}

func (c *execChild) Wait() error {
	err := c.cmd.Wait()
	if c.logFile != nil {
		_ = c.logFile.Close()
	}
	return err
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
