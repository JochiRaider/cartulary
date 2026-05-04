package main

import (
	"context"
	"database/sql"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
	"github.com/JochiRaider/cartulary/internal/testutil/suiteservices"
)

func TestMigrateBinaryRunsFromNonRepoWorkingDirectory(t *testing.T) {
	repoRoot, err := suiteservices.FindRepoRoot()
	if err != nil {
		t.Fatalf("resolve repo root: %v", err)
	}

	postgresHarness := pgtest.Start(t)
	testDB, err := postgresHarness.NewDatabase(context.Background(), "cmd-migrate-non-root")
	if err != nil {
		t.Fatalf("prepare postgres database: %v", err)
	}
	defer func() {
		if err := postgresHarness.DropDatabase(context.Background(), testDB.Name); err != nil {
			t.Fatalf("drop postgres database: %v", err)
		}
	}()

	binaryPath := filepath.Join(t.TempDir(), "migrate")
	build := exec.Command("go", "build", "-o", binaryPath, "./cmd/migrate")
	build.Dir = repoRoot
	build.Env = os.Environ()
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build migrate binary: %v\n%s", err, output)
	}

	command := exec.Command(binaryPath, "up")
	command.Dir = t.TempDir()
	command.Env = append(os.Environ(),
		"CARTULARY_CONFIG_FILE="+filepath.Join(repoRoot, "configs", "dev", "config.toml"),
		"CARTULARY_POSTGRES_POSTGRES_PRIMARY_DSN="+testDB.DSN,
	)
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("run migrate binary from non-repo cwd: %v\n%s", err, output)
	}

	db, err := sql.Open("pgx", testDB.DSN)
	if err != nil {
		t.Fatalf("open migrated database: %v", err)
	}
	defer db.Close()

	var exists bool
	if err := db.QueryRowContext(context.Background(), `
SELECT EXISTS (
    SELECT 1
    FROM information_schema.tables
    WHERE table_schema = 'public' AND table_name = 'users'
)`).Scan(&exists); err != nil {
		t.Fatalf("query migrated schema: %v", err)
	}
	if !exists {
		t.Fatal("expected migrate binary to apply schema from a non-repo working directory")
	}
}
