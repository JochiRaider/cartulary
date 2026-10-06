package recovery_test

import (
	"context"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/recovery"
	"github.com/JochiRaider/cartulary/internal/platform/objectstore"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
)

func TestVerificationResetPreservesMigrationIdentity_Integration(t *testing.T) {
	ctx := context.Background()
	targetDB := pgtest.Start(t).PrepareIsolatedDatabaseT(t, "recovery-reset-metadata")
	admin, err := pgtest.OpenPurposeDatabase(targetDB.DSN, postgres.PurposeMigration)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = admin.Close() })
	recoveryDSN, err := targetDB.DSNForPurpose(postgres.PurposeRecovery)
	if err != nil {
		t.Fatal(err)
	}
	pool, err := postgres.Setup(ctx, postgres.Settings{BindingKind: "managed_service", DSN: recoveryDSN, Purpose: postgres.PurposeRecovery, ExpectedRole: "cartulary_recovery"})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	objects, err := objectstore.NewFilesystemStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	target := recovery.RestoreTarget{Postgres: pool.Pool(), ObjectStore: objects}
	catalog := currentStateCatalog(t)
	extensions := testExtensionBackupCatalog(t)
	var ledgerBefore, lineageBefore int
	if err := admin.QueryRow(`SELECT (SELECT count(*) FROM goose_db_version), (SELECT count(*) FROM schema_migration_lineage)`).Scan(&ledgerBefore, &lineageBefore); err != nil {
		t.Fatal(err)
	}
	// Ordinary fixtures have completed admitted initialization. Only this
	// migration-purpose fixture preparation may create the missing-marker case.
	var initialFormat string
	if err := admin.QueryRow(`SELECT format_id FROM application_crypto_format WHERE singleton`).Scan(&initialFormat); err != nil || initialFormat != recovery.ApplicationCryptoFormatID {
		t.Fatalf("fixture was not admitted: %q %v", initialFormat, err)
	}
	if _, err := admin.Exec(`DELETE FROM application_crypto_format`); err != nil {
		t.Fatal(err)
	}
	for _, present := range []bool{false, true} {
		if present {
			if _, err := admin.Exec(`INSERT INTO application_crypto_format(singleton,format_id) VALUES (true,$1)`, recovery.ApplicationCryptoFormatID); err != nil {
				t.Fatal(err)
			}
		}
		if _, err := admin.Exec(`INSERT INTO users(email,display_name,password_hash,mfa_required,is_active,is_deployment_admin) VALUES ('reset@example.test','Reset','fixture',false,true,false)`); err != nil {
			t.Fatal(err)
		}
		if err := recovery.ResetRestoreVerificationTarget(ctx, target, extensions, catalog); err != nil {
			t.Fatal(err)
		}
		var markerCount, users, ledger, lineage int
		var format string
		if err := admin.QueryRow(`SELECT (SELECT count(*) FROM application_crypto_format), COALESCE((SELECT format_id FROM application_crypto_format),''), (SELECT count(*) FROM users), (SELECT count(*) FROM goose_db_version), (SELECT count(*) FROM schema_migration_lineage)`).Scan(&markerCount, &format, &users, &ledger, &lineage); err != nil {
			t.Fatal(err)
		}
		if (markerCount == 1) != present || (present && format != recovery.ApplicationCryptoFormatID) || users != 0 || ledger != ledgerBefore || lineage != lineageBefore {
			t.Fatalf("reset changed metadata or retained mutable state: present=%t marker=%d/%q users=%d ledger=%d lineage=%d", present, markerCount, format, users, ledger, lineage)
		}
	}
	var canWrite bool
	if err := pool.Pool().QueryRow(ctx, `SELECT has_table_privilege(current_user,'application_crypto_format','INSERT,UPDATE,DELETE,TRUNCATE')`).Scan(&canWrite); err != nil || canWrite {
		t.Fatalf("recovery identity privileges: %t %v", canWrite, err)
	}
	if _, err := admin.Exec(`INSERT INTO users(email,display_name,password_hash,mfa_required,is_active,is_deployment_admin) VALUES ('keep@example.test','Keep','fixture',false,true,false)`); err != nil {
		t.Fatal(err)
	}
	if _, err := admin.Exec(`CREATE TABLE public.unowned_recovery_state(id integer)`); err != nil {
		t.Fatal(err)
	}
	if err := recovery.ResetRestoreVerificationTarget(ctx, target, extensions, catalog); err == nil {
		t.Fatal("reset accepted unknown table")
	}
	var count int
	if err := admin.QueryRow(`SELECT count(*) FROM users`).Scan(&count); err != nil || count != 1 {
		t.Fatalf("unknown-table rejection mutated target: %d %v", count, err)
	}
}
