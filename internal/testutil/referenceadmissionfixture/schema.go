package referenceadmissionfixture

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/database_migrations"
	"github.com/JochiRaider/cartulary/internal/platform/config"
)

var SchemaCases = []string{"crypto state", "schema behind", "schema ahead", "schema lineage"}

// SeedSchemaFailure combines a schema/crypto failure with an unavailable content
// table. The schema diagnostic must win before any content query is attempted.
func SeedSchemaFailure(t testing.TB, db *sql.DB, mode string) {
	t.Helper()
	SeedMissingTable(t, db, "reference table missing")
	query := ""
	switch mode {
	case "crypto state":
		query = `UPDATE application_crypto_format SET format_id='unsupported'`
	case "schema behind":
		query = `DELETE FROM goose_db_version WHERE version_id=(SELECT max(version_id) FROM goose_db_version)`
	case "schema ahead":
		query = `INSERT INTO goose_db_version(version_id,is_applied) SELECT max(version_id)+1,true FROM goose_db_version`
	case "schema lineage":
		query = `UPDATE schema_migration_lineage SET lineage_id='fixture.unsupported.lineage'`
	default:
		t.Fatalf("unknown schema fixture: %s", mode)
	}
	if _, err := db.ExecContext(t.Context(), query); err != nil {
		t.Fatal(err)
	}
}

func RequireSchemaFailure(t testing.TB, err error, mode string) {
	t.Helper()
	if mode == "crypto state" {
		if !errors.Is(err, database_migrations.ErrIncompatibleCryptoState) {
			t.Fatal("crypto admission ordering", err)
		}
		return
	}
	want := map[string]string{"schema behind": "schema_migration_required", "schema ahead": "schema_version_ahead", "schema lineage": "historical_migration_lineage"}[mode]
	var diagnostics *config.DiagnosticsError
	if errors.As(err, &diagnostics) {
		if diagnostics.Code != config.InvalidDeploymentConfigCode || len(diagnostics.Diagnostics) != 1 || diagnostics.Diagnostics[0].ReasonCode != want {
			t.Fatal("schema diagnostics", err)
		}
		return
	}
	var reason interface{ ReasonCode() string }
	if !errors.As(err, &reason) || reason.ReasonCode() != want {
		t.Fatal("schema admission ordering", mode, err)
	}
}
