package pgtest

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	testcontainers "github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	dbmigrations "github.com/JochiRaider/cartulary/db/migrations"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/testutil/suiteservices"
	"github.com/JochiRaider/cartulary/internal/testutil/testcontainersx"
)

const postgresImage = "postgres:16-alpine"

type Harness struct {
	Container testcontainers.Container
	Host      string
	Port      string
	User      string
	Password  string

	adminDSN    string
	dsnTemplate string
	templateDB  string
	suiteHash   string
	processHash string
	counter     uint64
	shared      bool
	attached    bool
}

type TestDatabase struct {
	Name string
	DSN  string
}

var (
	sharedHarnessMu   sync.Mutex
	sharedHarness     *Harness
	startOwnedHarness = StartOwned
	pingAdminDSNFn    = pingAdminDSN
	migrateDatabaseFn = postgres.Migrate
	createDatabaseFn  = createDatabase
)

func Start(t testing.TB) *Harness {
	t.Helper()

	harness, err := StartShared(context.Background())
	if err != nil {
		t.Fatalf("%v", err)
	}

	return harness
}

func StartShared(ctx context.Context) (*Harness, error) {
	sharedHarnessMu.Lock()
	defer sharedHarnessMu.Unlock()

	if sharedHarness != nil {
		return sharedHarness, nil
	}

	harness, attached, err := startAttachedHarness(ctx)
	if err != nil {
		return nil, err
	}
	if !attached {
		harness, err = startOwnedHarness(ctx)
	}
	if err != nil {
		return nil, err
	}
	harness.shared = true
	sharedHarness = harness
	return harness, nil
}

func StartOwned(ctx context.Context) (*Harness, error) {
	return startHarness(ctx)
}

func StopShared(ctx context.Context) error {
	sharedHarnessMu.Lock()
	defer sharedHarnessMu.Unlock()

	if sharedHarness == nil || sharedHarness.Container == nil {
		sharedHarness = nil
		return nil
	}

	container := sharedHarness.Container
	sharedHarness = nil
	return container.Terminate(ctx)
}

func startHarness(ctx context.Context) (*Harness, error) {
	req := testcontainers.ContainerRequest{
		Image:        postgresImage,
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_DB":       "postgres",
			"POSTGRES_USER":     "cartulary",
			"POSTGRES_PASSWORD": "cartulary",
		},
		WaitingFor: wait.ForListeningPort("5432/tcp").WithStartupTimeout(60 * time.Second),
	}

	container, err := testcontainersx.StartWithRetry(ctx, testcontainersx.StartConfig{
		Service: "postgres testcontainer",
		Image:   postgresImage,
	}, func(ctx context.Context) (testcontainers.Container, error) {
		return testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
			ContainerRequest: req,
			Started:          true,
		})
	})
	if err != nil {
		return nil, err
	}

	host, err := container.Host(ctx)
	if err != nil {
		_ = container.Terminate(ctx)
		return nil, fmt.Errorf("resolve postgres host: %w", err)
	}

	port, err := container.MappedPort(ctx, "5432/tcp")
	if err != nil {
		_ = container.Terminate(ctx)
		return nil, fmt.Errorf("resolve postgres mapped port: %w", err)
	}

	harness := &Harness{
		Container:   container,
		Host:        host,
		Port:        port.Port(),
		User:        "cartulary",
		Password:    "cartulary",
		suiteHash:   resolveSuiteHash(),
		processHash: suiteservices.ProcessHash(),
	}
	harness.adminDSN = harness.dsnFor("postgres")

	if err := harness.WaitReady(ctx); err != nil {
		_ = harness.Close(ctx)
		return nil, fmt.Errorf("wait for postgres readiness: %w", err)
	}

	return harness, nil
}

func startAttachedHarness(ctx context.Context) (*Harness, bool, error) {
	adminDSN := strings.TrimSpace(suiteservices.LookupEnvValue(nil, suiteservices.PGAdminDSNEnv))
	dsnTemplate := strings.TrimSpace(suiteservices.LookupEnvValue(nil, suiteservices.PGDSNTemplateEnv))
	if adminDSN == "" && dsnTemplate == "" {
		return nil, false, nil
	}
	if adminDSN == "" || dsnTemplate == "" {
		return nil, false, fmt.Errorf("attach postgres harness: %s and %s must both be set", suiteservices.PGAdminDSNEnv, suiteservices.PGDSNTemplateEnv)
	}
	if !strings.Contains(dsnTemplate, "{database}") {
		return nil, false, fmt.Errorf("attach postgres harness: %s must contain {database}", suiteservices.PGDSNTemplateEnv)
	}

	harness := &Harness{
		adminDSN:    adminDSN,
		dsnTemplate: dsnTemplate,
		templateDB:  strings.TrimSpace(suiteservices.LookupEnvValue(nil, suiteservices.PGTemplateDBEnv)),
		suiteHash:   resolveSuiteHash(),
		processHash: suiteservices.ProcessHash(),
		attached:    true,
	}
	if err := pingAdminDSNFn(ctx, adminDSN); err != nil {
		return nil, false, fmt.Errorf("attach postgres harness: ping admin dsn: %w", err)
	}
	recordSuiteEvent(suiteservices.Event{Type: suiteservices.EventPostgresAttach})
	return harness, true, nil
}

func (h *Harness) WaitReady(ctx context.Context) error {
	deadline := time.Now().Add(30 * time.Second)
	var lastErr error
	for {
		err := pingAdminDSNFn(ctx, h.adminDSN)
		if err == nil {
			return nil
		}
		lastErr = err
		if time.Now().After(deadline) {
			return fmt.Errorf("postgres did not become ready: %w", lastErr)
		}
		time.Sleep(250 * time.Millisecond)
	}
}

func (h *Harness) AdminDSN() string {
	return h.adminDSN
}

func (h *Harness) NewDatabase(ctx context.Context, prefix string) (*TestDatabase, error) {
	name := h.nextDatabaseName(prefix)
	if err := h.createDatabase(ctx, name, ""); err != nil {
		return nil, err
	}
	recordSuiteEvent(suiteservices.Event{
		Type: suiteservices.EventPostgresDBCreated,
		Name: name,
		Kind: "scratch",
	})

	return &TestDatabase{
		Name: name,
		DSN:  h.dsnFor(name),
	}, nil
}

func (h *Harness) PrepareDatabase(ctx context.Context, prefix string) (*TestDatabase, postgres.MigrationStatus, error) {
	if h.templateDB != "" {
		name := h.nextDatabaseName(prefix)
		if err := h.createDatabase(ctx, name, h.templateDB); err != nil {
			return nil, postgres.MigrationStatus{}, err
		}
		recordSuiteEvent(suiteservices.Event{
			Type: suiteservices.EventPostgresDBCreated,
			Name: name,
			Kind: "template-clone",
		})
		recordSuiteEvent(suiteservices.Event{
			Type: suiteservices.EventPostgresTemplateUse,
			Name: name,
			Details: map[string]any{
				"template_database": h.templateDB,
			},
		})

		return &TestDatabase{
				Name: name,
				DSN:  h.dsnFor(name),
			}, postgres.MigrationStatus{
				Command:          "template-clone",
				Directory:        dbmigrations.RepositoryPath,
				TemplateClone:    true,
				TemplateDatabase: h.templateDB,
			}, nil
	}

	testDB, err := h.NewDatabase(ctx, prefix)
	if err != nil {
		return nil, postgres.MigrationStatus{}, err
	}

	db, err := sql.Open("pgx", testDB.DSN)
	if err != nil {
		return nil, postgres.MigrationStatus{}, fmt.Errorf("open test database handle: %w", err)
	}
	defer db.Close()

	status, err := migrateDatabaseFn(db, dbmigrations.Source(), "up")
	if err != nil {
		return nil, postgres.MigrationStatus{}, err
	}
	recordSuiteEvent(suiteservices.Event{
		Type: suiteservices.EventPostgresDBMigrated,
		Name: testDB.Name,
		Kind: "scratch",
	})

	return testDB, status, nil
}

func (h *Harness) PrepareDatabaseT(t testing.TB, prefix string) *TestDatabase {
	t.Helper()

	testDB, _, err := h.PrepareDatabase(context.Background(), prefix)
	if err != nil {
		t.Fatalf("prepare postgres database: %v", err)
	}
	t.Cleanup(func() {
		if err := h.DropDatabase(context.Background(), testDB.Name); err != nil {
			t.Fatalf("drop postgres database: %v", err)
		}
	})
	return testDB
}

func (h *Harness) DropDatabase(ctx context.Context, name string) error {
	admin, err := sql.Open("pgx", h.adminDSN)
	if err != nil {
		return fmt.Errorf("open postgres admin handle: %w", err)
	}
	defer admin.Close()

	if _, err := admin.ExecContext(ctx, fmt.Sprintf(`DROP DATABASE IF EXISTS "%s" WITH (FORCE)`, name)); err != nil {
		return fmt.Errorf("drop test database %s: %w", name, err)
	}

	return nil
}

func (h *Harness) Close(ctx context.Context) error {
	if h == nil || h.Container == nil {
		return nil
	}
	// Shared harnesses stay alive until StopShared or test-process teardown.
	if h.shared {
		return nil
	}

	return h.Container.Terminate(ctx)
}

func (db *TestDatabase) Env() map[string]string {
	return map[string]string{
		postgres.PostgresDSNEnv: db.DSN,
	}
}

func (h *Harness) dsnFor(database string) string {
	if h.dsnTemplate != "" {
		return strings.ReplaceAll(h.dsnTemplate, "{database}", database)
	}
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", h.User, h.Password, h.Host, h.Port, database)
}

func sanitizeIdentifier(value string) string {
	value = strings.ToLower(value)
	value = strings.ReplaceAll(value, "-", "_")
	value = strings.ReplaceAll(value, " ", "_")

	var builder strings.Builder
	for _, r := range value {
		switch {
		case r >= 'a' && r <= 'z':
			builder.WriteRune(r)
		case r >= '0' && r <= '9':
			builder.WriteRune(r)
		case r == '_':
			builder.WriteRune(r)
		}
	}

	result := strings.Trim(builder.String(), "_")
	if result == "" {
		return "testdb"
	}
	return result
}

func resolveSuiteHash() string {
	if hash := suiteservices.SuiteHash(nil); hash != "" {
		return hash
	}
	return suiteservices.ShortHash("local-suite", 8)
}

func (h *Harness) nextDatabaseName(prefix string) string {
	suiteHash := h.suiteHash
	if suiteHash == "" {
		suiteHash = resolveSuiteHash()
	}
	processHash := h.processHash
	if processHash == "" {
		processHash = suiteservices.ProcessHash()
	}

	base := fmt.Sprintf("ct_%s_%s_%06d", suiteHash, processHash, atomic.AddUint64(&h.counter, 1))
	suffix := sanitizeIdentifier(prefix)
	if suffix == "" {
		return truncateIdentifier(base, 63)
	}

	available := 63 - len(base) - 1
	if available <= 0 {
		return truncateIdentifier(base, 63)
	}
	return truncateIdentifier(base+"_"+truncateIdentifier(suffix, available), 63)
}

func truncateIdentifier(value string, max int) string {
	if len(value) <= max {
		return value
	}
	return strings.TrimRight(value[:max], "_")
}

func (h *Harness) createDatabase(ctx context.Context, name string, templateDB string) error {
	return createDatabaseFn(ctx, h.adminDSN, name, templateDB)
}

func pingAdminDSN(ctx context.Context, dsn string) error {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return err
	}
	defer db.Close()
	return db.PingContext(ctx)
}

func createDatabase(ctx context.Context, adminDSN string, name string, templateDB string) error {
	admin, err := sql.Open("pgx", adminDSN)
	if err != nil {
		return fmt.Errorf("open postgres admin handle: %w", err)
	}
	defer admin.Close()

	statement := fmt.Sprintf(`CREATE DATABASE "%s"`, name)
	if templateDB != "" {
		statement += fmt.Sprintf(` TEMPLATE "%s"`, templateDB)
	}
	if _, err := admin.ExecContext(ctx, statement); err != nil {
		if templateDB != "" {
			return fmt.Errorf("create test database %s from template %s: %w", name, templateDB, err)
		}
		return fmt.Errorf("create test database %s: %w", name, err)
	}
	return nil
}

func recordSuiteEvent(event suiteservices.Event) {
	_ = suiteservices.RecordEvent(nil, event)
}
