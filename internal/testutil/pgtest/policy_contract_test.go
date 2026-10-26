package pgtest

import (
	"context"
	"errors"
	"github.com/JochiRaider/cartulary/internal/testutil/suiteservices"
	"testing"
)

func TestPostgresFixturePolicyResolutionUsesTopLevelTestAndPackage(t *testing.T) {
	t.Setenv(postgresFixturePolicyTestsEnv, t.Name()+"=template_clone")
	t.Setenv(postgresFixturePolicyPackagesEnv, "./internal/testutil/pgtest=transaction")
	t.Setenv(postgresFixturePolicyDefaultEnv, "template_clone")
	t.Run("test_priority", func(t *testing.T) {
		if got := ExplicitPostgresFixturePolicyT(t); got != postgresFixturePolicyTemplateClone {
			t.Fatal(got)
		}
	})
	t.Setenv(postgresFixturePolicyTestsEnv, "")
	t.Run("package_priority", func(t *testing.T) {
		if got := ExplicitPostgresFixturePolicyT(t); got != postgresFixturePolicyTransaction {
			t.Fatal(got)
		}
	})
	t.Setenv(postgresFixturePolicyPackagesEnv, "")
	t.Run("explicit_default", func(t *testing.T) {
		if got := ExplicitPostgresFixturePolicyT(t); got != postgresFixturePolicyTemplateClone {
			t.Fatal(got)
		}
	})
	t.Setenv(postgresFixturePolicyDefaultEnv, "")
	t.Run("no_implicit_default", func(t *testing.T) {
		if got := ExplicitPostgresFixturePolicyT(t); got != "" {
			t.Fatal(got)
		}
	})
}

func TestValidateSelectedPostgresFixturePolicy(t *testing.T) {
	t.Run("migration_issuance_requires_capability", testMigrationIssuanceAdmission)
	dbOps := defaultDatabaseOperations()
	dbOps.create = func(_ context.Context, _, _, _ string, confirmed func() error) (bool, error) {
		t.Fatal("rejected policy acquired database")
		return false, nil
	}
	h := &Harness{binding: bindFixture(processFixtureInputs()), operations: &dbOps, templateDB: "suite_template"}
	for _, policy := range []string{"", postgresFixturePolicyTransaction} {
		t.Run("clone_rejects_"+policy, func(t *testing.T) {
			t.Setenv(postgresFixturePolicyTestsEnv, "")
			t.Setenv(postgresFixturePolicyPackagesEnv, "")
			t.Setenv(postgresFixturePolicyDefaultEnv, policy)
			h.binding = bindFixture(processFixtureInputs())
			failed := false
			func() {
				defer func() {
					if r := recover(); r != nil {
						if _, ok := r.(capturedPreparationFailure); !ok {
							panic(r)
						}
						failed = true
					}
				}()
				h.PrepareIsolatedDatabaseT(preparationFailureProbe{t}, "rejected")
			}()
			if !failed {
				t.Fatal("missing or mismatched policy accepted")
			}
		})
	}
	if err := validateSelectedPostgresFixturePolicy(postgresFixturePolicyTemplateClone, postgresFixturePolicyTemplateClone); err != nil {
		t.Fatal(err)
	}
}

func testMigrationIssuanceAdmission(t *testing.T) {
	for _, capability := range []string{"", "unknown", "postgres_transaction", "postgres_dedicated", "postgres_migration"} {
		for _, scope := range []string{"", "shared", "exclusive", "unknown"} {
			if capability == "postgres_migration" && scope == "exclusive" {
				continue
			}
			for _, targeted := range []bool{false, true} {
				t.Run(capability+"_"+scope+map[bool]string{false: "_head", true: "_targeted"}[targeted], func(t *testing.T) {
					inputs := map[string]string{
						"CARTULARY_FIXTURE_CAPABILITY":              capability,
						suiteservices.PGIsolationEnv:                scope,
						suiteservices.HarnessServiceDependenciesEnv: "postgres",
					}
					ops := defaultDatabaseOperations()
					acquisitions := 0
					ops.create = func(_ context.Context, _, _, _ string, _ func() error) (bool, error) {
						acquisitions++
						return false, errors.New("unexpected acquisition")
					}
					h := &Harness{binding: bindFixture(inputs), operations: &ops}
					func() {
						defer func() {
							if r := recover(); r != nil {
								if _, ok := r.(capturedPreparationFailure); !ok {
									panic(r)
								}
							}
						}()
						if targeted {
							h.MigrationDatabaseThroughT(preparationFailureProbe{t}, 1)
						} else {
							h.MigrationDatabaseT(preparationFailureProbe{t})
						}
					}()
					if acquisitions != 0 {
						t.Error("undeclared migration acquired a database")
					}
				})
			}
		}
	}
}
