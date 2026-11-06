package recoverybrowsertest

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/app/configassembly"
	"github.com/JochiRaider/cartulary/internal/modules/recovery"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/testutil/fixtures"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
)

// The fixture consumes the declared production operator. It never builds a
// command, copies a database or invokes the restore implementation as a shortcut.
type operatorResult struct {
	SchemaID           string     `json:"schema_id"`
	Result             string     `json:"result"`
	BackupSetID        *uuid.UUID `json:"backup_set_id"`
	ConsistencyPointAt *time.Time `json:"consistency_point_at"`
}

func operatorConfiguration(cfg configassembly.Loaded, dsn string) (string, error) {
	deployment := cfg.Deployment()
	root := deployment.Roots.DatabaseStorage.Path
	if err := os.MkdirAll(root, 0700); err != nil {
		return "", err
	}
	for purpose, filename := range map[postgres.Purpose]string{
		postgres.PurposeRecovery: postgres.FilesystemRecoveryDSNFile,
	} {
		selected, err := (&pgtest.TestDatabase{DSN: dsn}).DSNForPurpose(purpose)
		if err != nil {
			return "", err
		}
		if err := os.WriteFile(filepath.Join(root, filename), []byte(selected+"\n"), 0600); err != nil {
			return "", err
		}
	}
	var document map[string]any
	if err := toml.Unmarshal(fixtures.MustRead("config", "valid.toml"), &document); err != nil {
		return "", err
	}
	document["deployment_profile"] = "on_prem"
	document["roots"] = deployment.Roots
	document["bootstrap"] = deployment.Bootstrap
	document["network_flow_activity"] = deployment.NetworkFlowActivity
	document["reference_packs"] = deployment.ReferencePacks
	document["revisions"] = map[string]any{"conflict_token_key_ring_manifest_path": filepath.Join(filepath.Dir(root), "revisions-conflict-token-key-ring.json")}
	var encoded bytes.Buffer
	if err := toml.NewEncoder(&encoded).Encode(document); err != nil {
		return "", err
	}
	path := filepath.Join(filepath.Dir(root), "operator.toml")
	return path, os.WriteFile(path, encoded.Bytes(), 0600)
}

// Filesystem credentials are purpose-exclusive. The operator owns Recovery
// admission first; only a completed restore is rebound for normal serving.
func bindRestoredRuntime(cfg configassembly.Loaded, dsn string) error {
	root := cfg.Deployment().Roots.DatabaseStorage.Path
	selected, err := (&pgtest.TestDatabase{DSN: dsn}).DSNForPurpose(postgres.PurposeRuntime)
	if err != nil {
		return err
	}
	if err := os.Remove(filepath.Join(root, postgres.FilesystemRecoveryDSNFile)); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(root, postgres.FilesystemRuntimeDSNFile), []byte(selected+"\n"), 0600)
}

func runOperator(ctx context.Context, runtimeEnv map[string]string, args ...string) (operatorResult, error) {
	var result operatorResult
	bin := os.Getenv("CARTULARY_OPERATOR_BIN")
	if !filepath.IsAbs(bin) {
		return result, errors.New("browser operator fixture requires a declared absolute CARTULARY_OPERATOR_BIN")
	}
	info, err := os.Lstat(bin)
	if err != nil || !info.Mode().IsRegular() || info.Mode().Perm()&0111 == 0 {
		return result, errors.New("browser operator executable is unavailable")
	}
	command := exec.CommandContext(ctx, bin, args...)
	for _, pair := range os.Environ() {
		key, _, _ := strings.Cut(pair, "=")
		if !strings.HasPrefix(key, "CARTULARY_") {
			command.Env = append(command.Env, pair)
		}
	}
	for key, value := range runtimeEnv {
		if strings.HasPrefix(key, "CARTULARY_SECRET_") || key == authn.AuthMasterKeyEnv {
			command.Env = append(command.Env, key+"="+value)
		}
	}
	command.Env = append(command.Env, recovery.RecoveryMasterKeyEnv+"="+restoreBrowserRecoveryMasterKey)
	var stdout, stderr bytes.Buffer
	command.Stdout, command.Stderr = &stdout, &stderr
	err = command.Run()
	decodeErr := json.Unmarshal(stdout.Bytes(), &result)
	if err != nil || decodeErr != nil || stderr.Len() != 0 || result.SchemaID != "cartulary.operator_recovery_result.v2" || result.Result != "succeeded" || result.BackupSetID == nil || result.ConsistencyPointAt == nil {
		// The owning browser process redacts this private diagnostic pipe.
		return result, fmt.Errorf("operator %s %s failed: process=%v decode=%v stdout=%s stderr=%s", args[0], args[1], err, decodeErr, stdout.String(), stderr.String())
	}
	return result, nil
}
