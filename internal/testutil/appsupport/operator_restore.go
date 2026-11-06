package appsupport

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/BurntSushi/toml"
	"github.com/JochiRaider/cartulary/internal/app/configassembly"
	"github.com/JochiRaider/cartulary/internal/modules/recovery"
	"github.com/JochiRaider/cartulary/internal/platform/config"
	"github.com/JochiRaider/cartulary/internal/platform/objectstore"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/testutil/configtest"
	"github.com/JochiRaider/cartulary/internal/testutil/fixtures"
	"github.com/JochiRaider/cartulary/internal/testutil/s3test"
	"github.com/google/uuid"
)

// BackupRestoreServerWithOperator uses the harness-built production operator
// against the source's actual database, retained roots and object namespace.
// The target is newly allocated and empty; only the operator mutates it before
// ordinary server startup performs serving admission. No SQL copy or reseed is
// a substitute for either operation.
func (runtime *Runtime) BackupRestoreServerWithOperator(t testing.TB, source *ServerHarness, options ServerOptions) *ServerHarness {
	t.Helper()
	bin := os.Getenv("CARTULARY_OPERATOR_BIN")
	if !filepath.IsAbs(bin) {
		t.Fatal("an absolute CARTULARY_OPERATOR_BIN is required; select the operator runtime binary through Make")
	}
	if source == nil || options.Database != nil || options.ObjectStore != nil || options.loaded != nil {
		t.Fatal("operator recovery requires an owned source and a newly allocated target")
	}
	source.Server.Close()
	targetDB := runtime.PrepareIsolatedDatabase(t, options.Prefix+"-restore")
	targetObjects, targetObjectEnv := prepareApplicationObjectStore(t, options.Prefix+"-restore")
	service := s3test.Start(t)
	const sourceDBRef, targetDBRef = "ir_backup_source", "ir_restore_target"
	const sourceObjectRef, targetObjectRef = "ir_backup_objects", "ir_restored_objects"
	sourceConfig := source.Server.Config
	sourceConfig.Roots.DatabaseStorage = config.RootBinding{BindingKind: "managed_service", ServiceRef: sourceDBRef}
	sourceConfig.Roots.ObjectStorage = config.RootBinding{BindingKind: "managed_service", ServiceRef: sourceObjectRef}
	sourcePath := operatorRecoveryConfiguration(t, sourceConfig)
	targetRoots := configtest.SetupTempRoots(t).Paths
	targetEnv := map[string]string{}
	for k, v := range targetRoots {
		targetEnv[k] = v
	}
	for k, v := range options.Env {
		targetEnv[k] = v
	}
	targetEnv["CARTULARY__REFERENCE_PACKS__TRUST_BOOTSTRAP_PATH"] = fixtures.Path("reference-packs", "trust-bootstrap.json")
	targetEnv["CARTULARY__BOOTSTRAP__FIRST_ADMIN_MANIFEST_PATH"] = fixtures.Path("bootstrap-admin", "canonical.json")
	targetConfig := configtest.LoadFixture(t, []string{"config", "valid.toml"}, targetEnv).Deployment()
	targetConfig.Roots.DatabaseStorage = config.RootBinding{BindingKind: "managed_service", ServiceRef: targetDBRef}
	targetConfig.Roots.ObjectStorage = config.RootBinding{BindingKind: "managed_service", ServiceRef: targetObjectRef}
	targetPath := operatorRecoveryConfiguration(t, targetConfig)

	// Explicit files own configuration. Only credential references enter the
	// subprocess environment, so source and target overlays cannot bleed across.
	operatorEnv := map[string]string{recovery.RecoveryMasterKeyEnv: "MDEyMzQ1Njc4OWFiY2RlZjAxMjM0NTY3ODlhYmNkZWY="}
	restoredCredentials := map[string]string{}
	for k, v := range source.configurationEnv {
		if strings.HasPrefix(k, "CARTULARY_SECRET_") || k == "CARTULARY_AUTH_MASTER_KEY" {
			operatorEnv[k] = v
			restoredCredentials[k] = v
		}
	}
	for _, values := range []map[string]string{
		source.Database.EnvForServiceRef(sourceDBRef, postgres.PurposeRecovery),
		targetDB.EnvForServiceRef(targetDBRef, postgres.PurposeRecovery),
		service.EnvForServiceRef(sourceObjectRef, source.configurationEnv[objectstore.BucketEnv]),
		service.EnvForServiceRef(targetObjectRef, targetObjectEnv[objectstore.BucketEnv]),
	} {
		for k, v := range values {
			operatorEnv[k] = v
		}
	}
	backup := runRecoveryOperator(t, bin, operatorEnv, "backup", "create", "--source-config-file", sourcePath)
	if backup.BackupSetID == nil || *backup.BackupSetID == "" {
		t.Fatal("operator backup omitted the committed backup identity")
	}
	restored := runRecoveryOperator(t, bin, operatorEnv, "restore", "latest", "--source-config-file", sourcePath, "--target-config-file", targetPath, "--confirm-backup-set-id", *backup.BackupSetID, "--operation-id", uuid.NewString())
	if restored.BackupSetID == nil || *restored.BackupSetID != *backup.BackupSetID {
		t.Fatal("operator restored a different backup")
	}

	// Recovery retains deployment authentication state. The fresh server uses
	// the authorized restored encryption key, never a source incident's login.
	for _, values := range []map[string]string{
		restoredCredentials,
		targetDB.EnvForServiceRef(targetDBRef, postgres.PurposeRuntime),
		service.EnvForServiceRef(targetObjectRef, targetObjectEnv[objectstore.BucketEnv]),
	} {
		for k, v := range values {
			targetEnv[k] = v
		}
	}
	targetEnv[config.ConfigFileEnv] = targetPath
	targetEnv["CARTULARY__DEPLOYMENT_PROFILE"] = "on_prem"
	for kind, ref := range map[string]string{"DATABASE_STORAGE": targetDBRef, "OBJECT_STORAGE": targetObjectRef} {
		targetEnv["CARTULARY__ROOTS__"+kind+"__BINDING_KIND"] = "managed_service"
		targetEnv["CARTULARY__ROOTS__"+kind+"__PATH"] = ""
		targetEnv["CARTULARY__ROOTS__"+kind+"__SERVICE_REF"] = ref
	}
	loaded := configtest.LoadPath(t, targetPath, targetEnv)
	// Prove the operator completed the same target that the serving process
	// will admit. Keep diagnostic output to contract state and digest equality;
	// connection coordinates and credentials never enter retained results.
	bindings, err := configassembly.RecoveryTargetBindings(loaded.Deployment(), targetEnv)
	if err != nil {
		t.Fatalf("resolve restored serving bindings: %v", err)
	}
	inspection := openServerPool(t, targetDB)
	control, err := recovery.ReadTargetControl(t.Context(), inspection)
	inspection.Close()
	if err != nil {
		t.Fatalf("read operator completion: %v", err)
	}
	if err := control.RequireServing(bindings); err != nil {
		if control.BindingDigests == nil {
			t.Fatalf("operator completion has no target binding: state=%s", control.State)
		}
		t.Fatalf("operator completion cannot serve: state=%s database_match=%t objects_match=%t reference_packs_match=%t exports_match=%t control_valid=%t: %v", control.State,
			control.BindingDigests.DatabaseSHA256 == bindings.DatabaseSHA256,
			control.BindingDigests.ObjectStoreSHA256 == bindings.ObjectStoreSHA256,
			control.BindingDigests.ReferencePackStorageSHA256 == bindings.ReferencePackStorageSHA256,
			control.BindingDigests.ExportOutputsSHA256 == bindings.ExportOutputsSHA256,
			control.Validate() == nil, err)
	}
	options.Database, options.ObjectStore, options.Env, options.loaded = targetDB, targetObjects, targetEnv, &loaded
	return runtime.StartServer(t, options)
}

func operatorRecoveryConfiguration(t testing.TB, deployment configassembly.Deployment) string {
	t.Helper()
	var document map[string]any
	if err := toml.Unmarshal(fixtures.MustRead("config", "valid.toml"), &document); err != nil {
		t.Fatal(err)
	}
	document["deployment_profile"] = "on_prem"
	document["roots"] = deployment.Roots
	document["bootstrap"] = deployment.Bootstrap
	document["network_flow_activity"] = deployment.NetworkFlowActivity
	document["reference_packs"] = deployment.ReferencePacks
	document["revisions"] = map[string]any{"conflict_token_key_ring_manifest_path": fixtures.Path("revisions", "conflict-token-key-ring.json")}
	var encoded bytes.Buffer
	if err := toml.NewEncoder(&encoded).Encode(document); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "operator.toml")
	if err := os.WriteFile(path, encoded.Bytes(), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

type recoveryOperatorResult struct {
	SchemaID    string  `json:"schema_id"`
	Result      string  `json:"result"`
	BackupSetID *string `json:"backup_set_id"`
}

func runRecoveryOperator(t testing.TB, bin string, env map[string]string, args ...string) recoveryOperatorResult {
	t.Helper()
	command := exec.CommandContext(t.Context(), bin, args...)
	for _, pair := range os.Environ() {
		key, _, _ := strings.Cut(pair, "=")
		if !strings.HasPrefix(key, "CARTULARY_") {
			command.Env = append(command.Env, pair)
		}
	}
	for k, v := range env {
		command.Env = append(command.Env, k+"="+v)
	}
	var stdout, stderr bytes.Buffer
	command.Stdout, command.Stderr = &stdout, &stderr
	err := command.Run()
	var result recoveryOperatorResult
	if decodeErr := json.Unmarshal(stdout.Bytes(), &result); decodeErr != nil || err != nil || stderr.Len() != 0 || result.SchemaID != "cartulary.operator_recovery_result.v2" || result.Result != "succeeded" {
		t.Fatalf("operator %s %s: process=%v decode=%v stdout=%s stderr=%s", args[0], args[1], err, decodeErr, stdout.String(), stderr.String())
	}
	t.Logf("operator %s %s: exit=0 result=%s", args[0], args[1], stdout.String())
	return result
}
