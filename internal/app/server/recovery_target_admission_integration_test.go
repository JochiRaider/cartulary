package server

import (
	"errors"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/extensions"
	"github.com/JochiRaider/cartulary/internal/platform/objectstore"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/testutil/browserbuild"
	"github.com/JochiRaider/cartulary/internal/testutil/configtest"
	"github.com/JochiRaider/cartulary/internal/testutil/fixtures"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
	"github.com/JochiRaider/cartulary/internal/testutil/reportingfixture"
	"github.com/JochiRaider/cartulary/internal/testutil/s3test"
)

func TestIncompleteRecoveryTargetRejectsStartup_Integration(t *testing.T) {
	ctx := t.Context()
	h := pgtest.Start(t)
	target := h.PrepareIsolatedDatabaseT(t, "recovery-target-admission")
	admitted := openAdmittedRuntimePool(t, ctx, target)
	defer admitted.Close()
	control, err := pgtest.OpenPurposeDatabase(ctx, target.DSN, postgres.PurposeRecovery)
	if err != nil {
		t.Fatal(err)
	}
	defer control.Close()
	if _, err := control.ExecContext(ctx, `UPDATE public.recovery_target_state SET state='unproven',revision=revision+1`); err != nil {
		t.Fatal(err)
	}
	s3 := s3test.Start(t)
	env := s3.Env(s3.BootstrapBucketT(t, "recovery-target-admission"))
	store, err := objectstore.Setup(ctx, objectstore.Settings{
		BindingKind: "managed_service", Endpoint: env[objectstore.EndpointEnv],
		AccessKey: env[objectstore.AccessKeyEnv], SecretKey: env[objectstore.SecretKeyEnv],
		Secure: env[objectstore.SecureEnv] == "true", Bucket: env[objectstore.BucketEnv],
	}, objectstore.Instrumentation{})
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	var before int64
	const stateRows = `SELECT (SELECT count(*) FROM public.extension_state_metadata)
 + (SELECT count(*) FROM public.deployment_bootstrap_state)`
	if err := admitted.Pool().QueryRow(ctx, stateRows).Scan(&before); err != nil {
		t.Fatal(err)
	}
	roots := configtest.SetupTempRoots(t)
	for _, name := range []string{"initial attempt", "changed backup root"} {
		t.Run(name, func(t *testing.T) {
			overlays := make(map[string]string, len(roots.Paths)+1)
			for key, value := range roots.Paths {
				overlays[key] = value
			}
			overlays["CARTULARY__ROOTS__BACKUP_STORAGE__PATH"] = t.TempDir()
			overlays["CARTULARY__BOOTSTRAP__FIRST_ADMIN_MANIFEST_PATH"] = fixtures.Path("bootstrap-admin", "canonical.json")
			cfg := BindPostgres(t, RuntimeConfigWithOverlays(t, overlays), target.Env())
			runtime, err := NewRuntime(t.Context(), cfg, Options{
				BrowserBuild: browserbuild.New(t), ReportingRuntime: reportingfixture.Runtime{},
				Postgres: admitted, ObjectStore: store,
			})
			if runtime != nil {
				runtime.Close()
			}
			var startup *extensions.StartupValidationError
			if !errors.As(err, &startup) || startup.StartupCode() != "recovery_target_incomplete" {
				t.Fatalf("incomplete target reached initialization instead of closed Recovery admission: %v", err)
			}
			if len(startup.Findings) != 1 {
				t.Fatalf("unexpected startup findings: %+v", startup.Findings)
			}
			finding := startup.Findings[0]
			if finding.Path != "$" || finding.ReasonCode != "recovery_target_incomplete" || len(finding.Details) != 0 || finding.Message != "Recovery target completion is not established; application startup is closed." {
				t.Fatalf("unsafe or incorrect Recovery startup diagnostic: %+v", finding)
			}
			var after int64
			if err := admitted.Pool().QueryRow(t.Context(), stateRows).Scan(&after); err != nil || after != before {
				t.Fatalf("failed startup changed initialization state: before=%d after=%d error=%v", before, after, err)
			}
		})
	}
}
