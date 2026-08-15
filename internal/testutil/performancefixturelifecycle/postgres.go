package performancefixturelifecycle

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/JochiRaider/cartulary/internal/testutil/suiteservices"
)

func replaceDatabaseInDSN(adminDSN string, database string) (string, error) {
	parsed, err := url.Parse(adminDSN)
	if err != nil {
		return "", fmt.Errorf("parse postgres dsn: %w", err)
	}
	parsed.Path = "/" + database
	return parsed.String(), nil
}

func cloneDatabase(ctx context.Context, adminDSN string, source string, target string) error {
	if !safePostgresIdentifier(source) || !safePostgresIdentifier(target) {
		return errors.New("performance fixture database identity is unsafe")
	}
	admin, err := sql.Open("pgx", adminDSN)
	if err != nil {
		return err
	}
	defer admin.Close()
	if _, err := admin.ExecContext(ctx, fmt.Sprintf(`CREATE DATABASE "%s" TEMPLATE "%s"`, target, source)); err != nil {
		return fmt.Errorf("clone performance fixture database: %w", err)
	}
	return nil
}

func sealTemplate(ctx context.Context, adminDSN string, name string, owner string) error {
	admin, err := sql.Open("pgx", adminDSN)
	if err != nil {
		return err
	}
	defer admin.Close()
	for _, statement := range []string{
		fmt.Sprintf(`ALTER DATABASE "%s" WITH ALLOW_CONNECTIONS false`, name),
		fmt.Sprintf(`ALTER DATABASE "%s" WITH IS_TEMPLATE true`, name),
		fmt.Sprintf(`COMMENT ON DATABASE "%s" IS '%s'`, name, owner),
	} {
		if _, err := admin.ExecContext(ctx, statement); err != nil {
			return fmt.Errorf("seal performance fixture template: %w", err)
		}
	}
	return validateTemplate(ctx, adminDSN, name, owner)
}

func markTemplateOwned(ctx context.Context, adminDSN string, name string, owner string) error {
	if !safePostgresIdentifier(name) || strings.Contains(owner, "'") {
		return errors.New("performance fixture template ownership identity is unsafe")
	}
	admin, err := sql.Open("pgx", adminDSN)
	if err != nil {
		return err
	}
	defer admin.Close()
	if _, err := admin.ExecContext(ctx, fmt.Sprintf(`COMMENT ON DATABASE "%s" IS '%s'`, name, owner)); err != nil {
		return fmt.Errorf("mark performance fixture template ownership: %w", err)
	}
	return nil
}

func validateTemplate(ctx context.Context, adminDSN string, name string, owner string) error {
	admin, err := sql.Open("pgx", adminDSN)
	if err != nil {
		return err
	}
	defer admin.Close()
	var isTemplate bool
	var allowConnections bool
	var comment sql.NullString
	if err := admin.QueryRowContext(ctx, `SELECT datistemplate, datallowconn, shobj_description(oid, 'pg_database') FROM pg_database WHERE datname = $1`, name).Scan(&isTemplate, &allowConnections, &comment); err != nil {
		return fmt.Errorf("inspect performance fixture template: %w", err)
	}
	if !isTemplate || allowConnections || !comment.Valid || comment.String != owner {
		return errors.New("performance fixture template is not sealed for this suite and snapshot key")
	}
	return nil
}

func requireNoConnections(ctx context.Context, adminDSN string, name string) error {
	admin, err := sql.Open("pgx", adminDSN)
	if err != nil {
		return err
	}
	defer admin.Close()
	var count int
	if err := admin.QueryRowContext(ctx, `SELECT COUNT(*) FROM pg_stat_activity WHERE datname = $1`, name).Scan(&count); err != nil {
		return err
	}
	if count != 0 {
		return fmt.Errorf("performance fixture template has %d unknown open connection(s)", count)
	}
	return nil
}

func dropTemplate(ctx context.Context, adminDSN string, name string) error {
	if adminDSN == "" || !safePostgresIdentifier(name) {
		return errors.New("performance fixture template cleanup identity is incomplete")
	}
	admin, err := sql.Open("pgx", adminDSN)
	if err != nil {
		return err
	}
	defer admin.Close()
	if _, err := admin.ExecContext(ctx, fmt.Sprintf(`ALTER DATABASE "%s" WITH IS_TEMPLATE false`, name)); err != nil {
		return err
	}
	if _, err := admin.ExecContext(ctx, fmt.Sprintf(`DROP DATABASE IF EXISTS "%s" WITH (FORCE)`, name)); err != nil {
		return err
	}
	return nil
}

func dropDatabase(ctx context.Context, adminDSN string, name string) error {
	if !safePostgresIdentifier(name) {
		return errors.New("performance fixture clone cleanup identity is unsafe")
	}
	admin, err := sql.Open("pgx", adminDSN)
	if err != nil {
		return err
	}
	defer admin.Close()
	if _, err := admin.ExecContext(ctx, fmt.Sprintf(`DROP DATABASE IF EXISTS "%s" WITH (FORCE)`, name)); err != nil {
		return err
	}
	return nil
}

func CleanupSuite(ctx context.Context, env map[string]string) error {
	adminDSN := strings.TrimSpace(suiteservices.LookupEnvValue(env, suiteservices.PGAdminDSNEnv))
	suiteID := strings.TrimSpace(suiteservices.LookupEnvValue(env, suiteservices.SuiteIDEnv))
	if adminDSN == "" || suiteID == "" {
		return nil
	}
	privateRoot, ok, err := suiteservices.ResolveSuiteRuntimeDir(env)
	if err != nil {
		return err
	}
	if !ok {
		return errors.New("performance fixture cleanup requires a private suite runtime directory")
	}
	suiteRuntimeRoot := filepath.Join(privateRoot, "performance-fixtures", "templates", suiteservices.ShortHash(suiteID, 16))
	cloneRuntimeRoot := filepath.Join(privateRoot, "performance-fixtures", "clones", suiteservices.ShortHash(suiteID, 16))
	foundRuntimeRoot := false
	for _, runtimeRoot := range []string{suiteRuntimeRoot, cloneRuntimeRoot} {
		if _, err := os.Lstat(runtimeRoot); err == nil {
			foundRuntimeRoot = true
		} else if !errors.Is(err, os.ErrNotExist) {
			return err
		}
	}
	if !foundRuntimeRoot {
		return nil
	}
	admin, err := sql.Open("pgx", adminDSN)
	if err != nil {
		return err
	}
	prefix := "cartulary.performance-fixture:" + suiteservices.ShortHash(suiteID, 16) + ":"
	rows, err := admin.QueryContext(ctx, `SELECT datname, shobj_description(oid, 'pg_database') FROM pg_database WHERE shobj_description(oid, 'pg_database') LIKE $1 ORDER BY datname LIMIT 9`, prefix+"%")
	if err != nil {
		_ = admin.Close()
		return err
	}
	type ownedTemplate struct{ name, owner string }
	var owned []ownedTemplate
	for rows.Next() {
		var candidate ownedTemplate
		if err := rows.Scan(&candidate.name, &candidate.owner); err != nil {
			_ = rows.Close()
			_ = admin.Close()
			return err
		}
		owned = append(owned, candidate)
	}
	if err := rows.Close(); err != nil {
		_ = admin.Close()
		return err
	}
	if len(owned) > 8 {
		_ = admin.Close()
		return errors.New("performance fixture suite cleanup exceeded its bounded template scope")
	}
	wantNamePrefix := "ct_pfs_" + suiteservices.ShortHash(suiteID, 8) + "_"
	for _, candidate := range owned {
		key := strings.TrimPrefix(candidate.owner, prefix)
		if !lowerHexDigest(key) || candidate.name != wantNamePrefix+key[:12] {
			_ = admin.Close()
			return errors.New("performance fixture suite cleanup rejected an unowned template")
		}
		if _, err := admin.ExecContext(ctx, fmt.Sprintf(`ALTER DATABASE "%s" WITH IS_TEMPLATE false`, candidate.name)); err != nil {
			_ = admin.Close()
			return err
		}
		if _, err := admin.ExecContext(ctx, fmt.Sprintf(`DROP DATABASE "%s" WITH (FORCE)`, candidate.name)); err != nil {
			_ = admin.Close()
			return err
		}
	}
	if err := admin.Close(); err != nil {
		return err
	}
	return errors.Join(os.RemoveAll(suiteRuntimeRoot), os.RemoveAll(cloneRuntimeRoot))
}
