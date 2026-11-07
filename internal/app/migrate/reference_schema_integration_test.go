package migrate

import (
	"bytes"
	"testing"

	"github.com/JochiRaider/cartulary/internal/app/configassembly"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/testutil/configtest"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
)

func TestMigrateReferencePackSchemaBoundary_Integration(t *testing.T) {
	harness := pgtest.Start(t)
	database := harness.PrepareIsolatedDatabaseT(t, "migration-reference-schema")
	db, err := pgtest.OpenPurposeDatabase(t.Context(), database.DSN, postgres.PurposeMigration)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	// Pack content admission belongs to runtime and recovery. The schema command
	// must not parse, rewrite or convert content while applying the current DDL.
	input := []byte(`{"pack_contract_version":"unsupported-synthetic-input"}`)
	if _, err := db.Exec(`INSERT INTO reference_pack_key_state(pack_key) VALUES('enrichment.entra_app_ids')`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO reference_pack_candidates(pack_key,pack_version,distribution_kind,health,admitted_at) VALUES('enrichment.entra_app_ids','synthetic','operator_imported','staged',now())`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO reference_pack_versions(pack_key,pack_version,pack_release_sequence,manifest_sha256,payload_sha256,manifest_bytes) VALUES('enrichment.entra_app_ids','synthetic',1,repeat('a',64),repeat('b',64),$1)`, input); err != nil {
		t.Fatal(err)
	}
	roots := configtest.SetupTempRoots(t)
	configtest.BindPostgresDSNToDatabaseRoot(t, roots.Paths["CARTULARY__ROOTS__DATABASE_STORAGE__PATH"], database.DSN, postgres.PurposeMigration)
	loaded := configtest.LoadFixture(t, []string{"config", "valid.toml"}, roots.Paths)
	runner := newMigrateRunner(&bytes.Buffer{})
	runner.loadConfig = func() (configassembly.Loaded, error) { return loaded, nil }
	if err := runner.run(t.Context()); err != nil {
		t.Fatal("schema command inspected pack content", err)
	}
	var retained []byte
	if err := db.QueryRow(`SELECT manifest_bytes FROM reference_pack_versions WHERE pack_key='enrichment.entra_app_ids' AND pack_version='synthetic'`).Scan(&retained); err != nil || !bytes.Equal(input, retained) {
		t.Fatal("schema command changed pack content", err)
	}
}
