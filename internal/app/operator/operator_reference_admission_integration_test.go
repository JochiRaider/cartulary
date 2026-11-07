package operator

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/app/configassembly"
	"github.com/JochiRaider/cartulary/internal/modules/reference_data/contractadmission"
	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/testutil/configtest"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
	admissionfixture "github.com/JochiRaider/cartulary/internal/testutil/referenceadmissionfixture"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
)

func TestOperatorReferencePackRetainedAdmission_Integration(t *testing.T) {
	harness := pgtest.Start(t)
	for _, kind := range admissionfixture.Cases {
		t.Run(kind, func(t *testing.T) {
			database := harness.PrepareIsolatedDatabaseT(t, "operator-reference-admission")
			db, err := pgtest.OpenPurposeDatabase(t.Context(), database.DSN, postgres.PurposeMigration)
			if err != nil {
				t.Fatal(err)
			}
			defer db.Close()
			admissionfixture.Seed(t, db, kind)
			roots := configtest.SetupTempRoots(t)
			configtest.BindPostgresDSNToDatabaseRoot(t, roots.Paths["CARTULARY__ROOTS__DATABASE_STORAGE__PATH"], database.DSN, postgres.PurposeRuntime)
			loaded := configtest.LoadFixture(t, []string{"config", "valid.toml"}, roots.Paths)
			cfg := loaded.Deployment()
			admissionfixture.Sentinels(t, cfg.Roots.TemporaryWork.Path, cfg.Roots.ReferencePackStorage.Path)
			state, files := admissionfixture.State(t, db), admissionfixture.Files(t, roots.Base)
			port, close, err := openReferencePackLocalRuntime(t.Context(), loaded)
			if close != nil {
				close()
			}
			if port != nil || !contractadmission.IsIncompatible(err) {
				t.Fatalf("retained %s reached mutable runtime: %v", kind, err)
			}
			var stdout, stderr bytes.Buffer
			executor := referencePackExecutor{transport: operatorTransport{stdout: &stdout, stderr: &stderr}, newOperationID: uuid.New, loadConfig: func(string) (configassembly.Loaded, error) { return loaded, nil }, open: openReferencePackLocalRuntime}
			if exit := executor.runCommand(t.Context(), []string{"reference-pack", "import", "retained.pack"}); exit != 3 {
				t.Fatal("exit", exit)
			}
			var result referencePackOperatorResult
			if err := json.Unmarshal(stdout.Bytes(), &result); err != nil || result.ErrorCode == nil || *result.ErrorCode != "reference_pack_operation_rejected" || result.ReasonCode == nil || *result.ReasonCode != "contract_incompatible" || result.JobID != nil || result.PackKey != nil || result.PackVersion != nil || result.ContainerSHA256 != nil || stderr.Len() != 0 || bytes.Count(stdout.Bytes(), []byte("\n")) != 1 {
				t.Fatal("unsafe or misattributed rejection", stdout.String(), stderr.String(), err)
			}
			encoded, err := canonicaljson.Marshal(result)
			if err != nil || !bytes.Equal(stdout.Bytes(), append(encoded, '\n')) {
				t.Fatal("operator result is not canonical JSON plus LF")
			}
			if after := admissionfixture.State(t, db); after != state {
				t.Fatal("rejection changed database, leases, Jobs or migration ledger")
			}
			if after := admissionfixture.Files(t, roots.Base); after != files {
				t.Fatal("rejection changed retained storage")
			}
		})
	}
	for _, mode := range append([]string{"current", "unavailable portable", "cancelled", "query failure", "binding query failure", "reference table missing", "reporting table missing"}, admissionfixture.SchemaCases...) {
		t.Run(mode, func(t *testing.T) {
			database := harness.PrepareIsolatedDatabaseT(t, "operator-admission-outcome")
			db, err := pgtest.OpenPurposeDatabase(t.Context(), database.DSN, postgres.PurposeMigration)
			if err != nil {
				t.Fatal(err)
			}
			defer db.Close()
			if mode == "unavailable portable" {
				raw, err := os.ReadFile("../../../contracts/reference-pack-fixtures/fixtures/portable-references.v1.json")
				if err != nil {
					t.Fatal(err)
				}
				admissionfixture.SeedPortable(t, db, raw)
			}
			if slices.Contains(admissionfixture.SchemaCases, mode) {
				admissionfixture.SeedSchemaFailure(t, db, mode)
			}
			if mode == "query failure" {
				if _, err := db.ExecContext(t.Context(), `ALTER TABLE reference_pack_sets RENAME COLUMN canonical_set TO unavailable_column`); err != nil {
					t.Fatal(err)
				}
			}
			if mode == "binding query failure" {
				admissionfixture.SeedBindingQueryFailure(t, db)
			}
			if mode == "reference table missing" || mode == "reporting table missing" {
				admissionfixture.SeedMissingTable(t, db, mode)
			}
			roots := configtest.SetupTempRoots(t)
			configtest.BindPostgresDSNToDatabaseRoot(t, roots.Paths["CARTULARY__ROOTS__DATABASE_STORAGE__PATH"], database.DSN, postgres.PurposeRuntime)
			loaded := configtest.LoadFixture(t, []string{"config", "valid.toml"}, roots.Paths)
			cfg := loaded.Deployment()
			admissionfixture.Sentinels(t, cfg.Roots.TemporaryWork.Path, cfg.Roots.ReferencePackStorage.Path)
			state, files := admissionfixture.State(t, db), admissionfixture.Files(t, roots.Base)
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()
			if mode == "cancelled" {
				cancel()
			}
			port, close, err := openReferencePackLocalRuntime(ctx, loaded)
			if close != nil {
				close()
			}
			if mode == "current" || mode == "unavailable portable" {
				if err != nil || port == nil || close == nil {
					t.Fatal("current state rejected", err)
				}
				collected := filepath.Join(cfg.Roots.ReferencePackStorage.Path, "reference-packs/objects", strings.Repeat("a", 64)+"-00000000-0000-4000-8000-000000000001")
				if _, err := os.Stat(collected); !os.IsNotExist(err) {
					t.Fatal("published sentinel was not collectible", err)
				}
				return
			}
			if err == nil || port != nil || contractadmission.IsIncompatible(err) {
				t.Fatal("operational failure classification", err)
			}
			if slices.Contains(admissionfixture.SchemaCases, mode) {
				admissionfixture.RequireSchemaFailure(t, err, mode)
			}
			if mode == "query failure" || mode == "binding query failure" || mode == "reference table missing" || mode == "reporting table missing" {
				code := "42703"
				if mode == "reference table missing" || mode == "reporting table missing" {
					code = "42P01"
				}
				var databaseError *pgconn.PgError
				if !errors.As(err, &databaseError) || databaseError.Code != code {
					t.Fatal("did not reach the query fault", err)
				}
			}
			var stdout, stderr bytes.Buffer
			executor := referencePackExecutor{transport: operatorTransport{stdout: &stdout, stderr: &stderr}, newOperationID: uuid.New, loadConfig: func(string) (configassembly.Loaded, error) { return loaded, nil }, open: openReferencePackLocalRuntime}
			if exit := executor.runCommand(ctx, []string{"reference-pack", "import", "retained.pack"}); exit != 3 {
				t.Fatal("operational failure exit", exit)
			}
			var result referencePackOperatorResult
			if err := json.Unmarshal(stdout.Bytes(), &result); err != nil || result.ErrorCode == nil || *result.ErrorCode != "internal_error" || result.ReasonCode != nil || result.JobID != nil || result.PackKey != nil || result.PackVersion != nil || result.ContainerSHA256 != nil || stderr.Len() != 0 {
				t.Fatal("unsafe or misclassified operational result", stdout.String(), stderr.String(), err)
			}
			encoded, encodeErr := canonicaljson.Marshal(result)
			if encodeErr != nil || !bytes.Equal(stdout.Bytes(), append(encoded, '\n')) {
				t.Fatal("operational result is not one canonical JSON object plus LF")
			}
			if admissionfixture.State(t, db) != state || admissionfixture.Files(t, roots.Base) != files {
				t.Fatal("operational rejection mutated state")
			}
		})
	}
}
