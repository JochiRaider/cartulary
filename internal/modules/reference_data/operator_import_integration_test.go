package reference_data_test

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/app/operator"
	"github.com/JochiRaider/cartulary/internal/modules/reference_data"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/configtest"
)

func TestLocalOperatorImportUsesDeploymentWorker_Integration(t *testing.T) {
	runtime := appsupport.StartRuntime(t)
	harness := startReferencePackServer(t, runtime, "reference-pack-operator-import")
	cfg := harness.Server.Config
	env := harness.Database.Env()
	env["CARTULARY__ROOTS__DATABASE_STORAGE__PATH"] = cfg.Roots.DatabaseStorage.Path
	env["CARTULARY__ROOTS__OBJECT_STORAGE__PATH"] = cfg.Roots.ObjectStorage.Path
	env["CARTULARY__ROOTS__BACKUP_STORAGE__PATH"] = cfg.Roots.BackupStorage.Path
	env["CARTULARY__ROOTS__REFERENCE_PACK_STORAGE__PATH"] = cfg.Roots.ReferencePackStorage.Path
	env["CARTULARY__ROOTS__TEMPORARY_WORK__PATH"] = cfg.Roots.TemporaryWork.Path
	env["CARTULARY__ROOTS__EXPORT_OUTPUTS__PATH"] = cfg.Roots.ExportOutputs.Path
	env["CARTULARY__REFERENCE_PACKS__TRUST_BOOTSTRAP_PATH"] = cfg.ReferencePacks.TrustBootstrapPath
	env["CARTULARY__REFERENCE_PACKS__CLOCK_TRUSTED"] = "true"
	env["CARTULARY__BOOTSTRAP__FIRST_ADMIN_MANIFEST_PATH"] = cfg.Bootstrap.FirstAdminManifestPath
	for key, value := range configtest.EffectiveConfigEnv([]string{"config", "valid.toml"}, env) {
		t.Setenv(key, value)
	}
	incoming := filepath.Join(cfg.Roots.ReferencePackStorage.Path, "incoming")
	if err := os.MkdirAll(incoming, 0700); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name string
		bad  bool
		exit int
	}{{"valid.zip", false, 0}, {"rejected.zip", true, 3}} {
		t.Run(tc.name, func(t *testing.T) {
			container := referencePackBundle(t, bundleOptions{PackKey: "type_registry.host", PackVersion: "1", BadSignature: tc.bad})
			if err := os.WriteFile(filepath.Join(incoming, tc.name), container, 0600); err != nil {
				t.Fatal(err)
			}
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			var stdout, stderr bytes.Buffer
			exit := operator.RunOperatorCLIContext(ctx, []string{"reference-pack", "import", tc.name}, &stdout, &stderr)
			if exit != tc.exit || stderr.Len() != 0 {
				t.Fatalf("operator exit=%d stdout=%s stderr=%s", exit, stdout.String(), stderr.String())
			}
			var result struct {
				OperationID string  `json:"operation_id"`
				JobID       string  `json:"job_id"`
				Result      string  `json:"result"`
				SHA         string  `json:"container_sha256"`
				Code        *string `json:"error_code"`
			}
			if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
				t.Fatal(err)
			}
			if result.SHA != integrationDigest(container) || result.JobID == "" || result.OperationID == "" {
				t.Fatal("incomplete attribution", stdout.String())
			}
			if tc.bad {
				if result.Result != "failed" || result.Code == nil || *result.Code != "reference_pack_verification_failed" {
					t.Fatal("wrong rejection", stdout.String())
				}
			} else if result.Result != "succeeded" || result.Code != nil {
				t.Fatal("wrong success", stdout.String())
			}
			if queryCount(t, harness.DB, `SELECT count(*) FROM jobs j JOIN reference_pack_operations o ON o.job_id=j.job_id WHERE j.job_id=$1 AND j.submitted_by_user_id IS NULL AND o.operation_id=$2 AND o.actor_kind='local_operator' AND o.actor_user_id IS NULL AND o.terminal_at IS NOT NULL`, result.JobID, result.OperationID) != 1 {
				t.Fatal("operator job not attributed")
			}
			if queryCount(t, harness.DB, `SELECT count(*) FROM reference_pack_candidates c WHERE pack_key='type_registry.host' AND pack_version='1' AND health='verified_available' AND NOT EXISTS(SELECT 1 FROM reference_pack_current_set s JOIN reference_pack_set_members m USING(pack_set_id) WHERE m.pack_key=c.pack_key AND m.pack_version=c.pack_version)`) != 1 {
				t.Fatal("renewal rejection changed retained content or import activated")
			}
			if !tc.bad {
				requireReferencePackProof(t, harness.DB, result.JobID, reference_data.ImportJobKind, reference_data.ImportOperation)
			}
		})
	}
}
