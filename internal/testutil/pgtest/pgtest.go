package pgtest

import (
	"context"
	"database/sql"
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	testcontainers "github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

type Harness struct {
	Container testcontainers.Container
	Host      string
	Port      string
	User      string
	Password  string

	adminDSN string
	counter  uint64
}

type TestDatabase struct {
	Name string
	DSN  string
}

func Start(t testing.TB) *Harness {
	t.Helper()

	harness, err := StartShared(context.Background())
	if err != nil {
		t.Fatalf("%v", err)
	}

	t.Cleanup(func() {
		if err := harness.Close(context.Background()); err != nil {
			t.Fatalf("terminate postgres testcontainer: %v", err)
		}
	})

	return harness
}

func StartShared(ctx context.Context) (*Harness, error) {
	req := testcontainers.ContainerRequest{
		Image:        "postgres:16-alpine",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_DB":       "postgres",
			"POSTGRES_USER":     "cartulary",
			"POSTGRES_PASSWORD": "cartulary",
		},
		WaitingFor: wait.ForListeningPort("5432/tcp").WithStartupTimeout(60 * time.Second),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, fmt.Errorf("start postgres testcontainer: %w", err)
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
		Container: container,
		Host:      host,
		Port:      port.Port(),
		User:      "cartulary",
		Password:  "cartulary",
	}
	harness.adminDSN = harness.dsnFor("postgres")

	if err := harness.WaitReady(ctx); err != nil {
		_ = harness.Close(ctx)
		return nil, fmt.Errorf("wait for postgres readiness: %w", err)
	}

	return harness, nil
}

func (h *Harness) WaitReady(ctx context.Context) error {
	deadline := time.Now().Add(30 * time.Second)
	for {
		db, err := sql.Open("pgx", h.adminDSN)
		if err == nil {
			err = db.PingContext(ctx)
			_ = db.Close()
		}
		if err == nil {
			return nil
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("postgres did not become ready: %w", err)
		}
		time.Sleep(250 * time.Millisecond)
	}
}

func (h *Harness) AdminDSN() string {
	return h.adminDSN
}

func (h *Harness) NewDatabase(ctx context.Context, prefix string) (*TestDatabase, error) {
	name := fmt.Sprintf("%s_%06d", sanitizeIdentifier(prefix), atomic.AddUint64(&h.counter, 1))

	admin, err := sql.Open("pgx", h.adminDSN)
	if err != nil {
		return nil, fmt.Errorf("open postgres admin handle: %w", err)
	}
	defer admin.Close()

	if _, err := admin.ExecContext(ctx, fmt.Sprintf(`CREATE DATABASE "%s"`, name)); err != nil {
		return nil, fmt.Errorf("create test database %s: %w", name, err)
	}

	return &TestDatabase{
		Name: name,
		DSN:  h.dsnFor(name),
	}, nil
}

func (h *Harness) PrepareDatabase(ctx context.Context, prefix string) (*TestDatabase, postgres.MigrationStatus, error) {
	testDB, err := h.NewDatabase(ctx, prefix)
	if err != nil {
		return nil, postgres.MigrationStatus{}, err
	}

	db, err := sql.Open("pgx", testDB.DSN)
	if err != nil {
		return nil, postgres.MigrationStatus{}, fmt.Errorf("open test database handle: %w", err)
	}
	defer db.Close()

	status, err := postgres.Migrate(db, migrationsDir(), "up")
	if err != nil {
		return nil, postgres.MigrationStatus{}, err
	}

	return testDB, status, nil
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

	return h.Container.Terminate(ctx)
}

func (db *TestDatabase) Env() map[string]string {
	return map[string]string{
		postgres.PostgresDSNEnv: db.DSN,
	}
}

func (h *Harness) dsnFor(database string) string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", h.User, h.Password, h.Host, h.Port, database)
}

func migrationsDir() string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(file), "..", "..", "..", "db", "migrations")
}

func sanitizeIdentifier(value string) string {
	value = strings.ToLower(value)
	value = strings.ReplaceAll(value, "-", "_")
	value = strings.ReplaceAll(value, " ", "_")
	if value == "" {
		return "testdb"
	}
	return value
}
