package phase1test

import (
	"context"
	"database/sql"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/JochiRaider/cartulary/internal/modules/auth"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/testutil/fixtures"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
	"github.com/JochiRaider/cartulary/internal/testutil/s3test"
)

type RuntimeHarness struct {
	Postgres *pgtest.Harness
	S3       *s3test.Harness
}

type ServerHarness struct {
	Server *httptestx.Server
	DB     *sql.DB
}

func StartRuntime(t testing.TB) *RuntimeHarness {
	t.Helper()

	return &RuntimeHarness{
		Postgres: pgtest.Start(t),
		S3:       s3test.Start(t),
	}
}

func (h *RuntimeHarness) StartServer(t testing.TB, prefix string) *ServerHarness {
	t.Helper()

	testDB, _, err := h.Postgres.PrepareDatabase(context.Background(), prefix)
	if err != nil {
		t.Fatalf("prepare postgres database: %v", err)
	}
	t.Cleanup(func() {
		if err := h.Postgres.DropDatabase(context.Background(), testDB.Name); err != nil {
			t.Fatalf("drop postgres database: %v", err)
		}
	})

	bucket, err := h.S3.BootstrapBucket(context.Background(), prefix)
	if err != nil {
		t.Fatalf("bootstrap bucket: %v", err)
	}
	t.Cleanup(func() {
		if err := h.S3.CleanupBucket(context.Background(), bucket); err != nil {
			t.Logf("cleanup bucket: %v", err)
		}
	})

	env := testDB.Env()
	for key, value := range h.S3.Env(bucket) {
		env[key] = value
	}
	env["CARTULARY__BOOTSTRAP__FIRST_ADMIN_MANIFEST_PATH"] = fixtures.Path("bootstrap-admin", "canonical.json")

	server := httptestx.StartServer(t, httptestx.ServerOptions{
		Env:              env,
		AdditionalRoutes: []httpapi.RouteRegistrar{auth.RegisterTestRoutes()},
	})

	db, err := sql.Open("pgx", testDB.DSN)
	if err != nil {
		t.Fatalf("open postgres sql handle: %v", err)
	}
	t.Cleanup(func() {
		_ = db.Close()
	})

	return &ServerHarness{
		Server: server,
		DB:     db,
	}
}
