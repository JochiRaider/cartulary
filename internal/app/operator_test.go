package app

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/platform/config"
	"github.com/JochiRaider/cartulary/internal/platform/objectstore"
)

func TestOperatorObjectStoreInitCommand_U_DeploymentLocalResult(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	var gotConfigPath string
	var ensureCalls int
	runner := operatorRunner{
		stdout: &stdout,
		stderr: &stderr,
		loadConfig: func(path string) (config.Config, error) {
			gotConfigPath = path
			return config.Config{ConfigSchemaID: "cartulary.deployment_config.v1"}, nil
		},
		ensureObjectStoreBucket: func(context.Context, config.Config) (objectstore.EnsureBucketResult, error) {
			ensureCalls++
			return objectstore.EnsureBucketResult{Created: true}, nil
		},
	}

	exitCode := runner.runCLI(context.Background(), []string{"object-store", "init", "-config", "/etc/cartulary/config.toml"})
	if exitCode != 0 {
		t.Fatalf("object-store init failed: exit=%d stdout=%s stderr=%s", exitCode, stdout.String(), stderr.String())
	}
	if gotConfigPath != "/etc/cartulary/config.toml" {
		t.Fatalf("unexpected config path: got %q", gotConfigPath)
	}
	if ensureCalls != 1 {
		t.Fatalf("expected exactly one ensure call, got %d", ensureCalls)
	}

	var payload OperatorObjectStoreInitResult
	if err := json.Unmarshal(stdout.Bytes(), &payload); err != nil {
		t.Fatalf("decode object-store init payload: %v\nstdout=%s", err, stdout.String())
	}
	if payload.SchemaID != OperatorObjectStoreInitResultSchemaID || payload.Result != "created" || !payload.Created || payload.AlreadyExists {
		t.Fatalf("unexpected object-store init payload: %#v", payload)
	}
	if strings.Contains(stdout.String(), "bucket") || strings.Contains(stdout.String(), "endpoint") {
		t.Fatalf("object-store init result exposed storage details: %s", stdout.String())
	}
}

func TestOperatorObjectStoreInitCommand_U_RedactsFailure(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	forbidden := []string{
		"http://127.0.0.1:9000",
		"secret-bucket",
		"AKIA-SECRET",
		"object/key.txt",
		"storage://unsafe/ref",
		"postgres://user:pass@db.example.test/cartulary",
	}
	runner := operatorRunner{
		stdout: &stdout,
		stderr: &stderr,
		loadConfig: func(string) (config.Config, error) {
			return config.Config{ConfigSchemaID: "cartulary.deployment_config.v1"}, nil
		},
		ensureObjectStoreBucket: func(context.Context, config.Config) (objectstore.EnsureBucketResult, error) {
			return objectstore.EnsureBucketResult{}, errors.New(strings.Join(forbidden, " "))
		},
	}

	exitCode := runner.runCLI(context.Background(), []string{"object-store", "init"})
	if exitCode == 0 {
		t.Fatalf("expected object-store init failure, stdout=%s stderr=%s", stdout.String(), stderr.String())
	}
	if stdout.Len() != 0 {
		t.Fatalf("expected no success payload on failure, got %s", stdout.String())
	}
	for _, value := range forbidden {
		if strings.Contains(stderr.String(), value) {
			t.Fatalf("object-store init failure leaked forbidden value %q in stderr %q", value, stderr.String())
		}
	}
	if !strings.Contains(stderr.String(), "object-store init failed: reason_code=dependency_unavailable") {
		t.Fatalf("stderr did not include redacted failure reason: %s", stderr.String())
	}
}

func TestOperatorObjectStoreMigrationSupportCommand_U_DoesNotRequireDeploymentAdmin(t *testing.T) {
	var stderr bytes.Buffer
	result := parseOperatorCLIArgs([]string{
		"object-store-migration",
		"run",
		"-source-config",
		"/etc/cartulary/source.toml",
		"-target-config",
		"/etc/cartulary/target.toml",
		"-confirm-backup-set-id",
		"00000000-0000-0000-0000-000000130111",
		"-quiescence-proof",
		"/tmp/quiescence-proof.json",
		"-artifacts-dir",
		"/tmp/object-store-migration",
		"-as-of",
		"2026-06-04T12:00:00Z",
	}, &stderr)
	if result.stop {
		t.Fatalf("parse stopped: exit=%d stderr=%s", result.exitCode, stderr.String())
	}
	if result.command != "object-store-migration run" {
		t.Fatalf("unexpected command: %q", result.command)
	}
	if result.email != "" {
		t.Fatalf("support command unexpectedly required admin identity: %q", result.email)
	}
	if result.confirmBackupSetID.String() != "00000000-0000-0000-0000-000000130111" {
		t.Fatalf("unexpected confirm backup id: %s", result.confirmBackupSetID)
	}
	if got := result.asOf.Format(time.RFC3339); got != "2026-06-04T12:00:00Z" {
		t.Fatalf("unexpected as-of: %s", got)
	}

	stderr.Reset()
	invalid := parseOperatorCLIArgs([]string{
		"object-store-migration",
		"run",
		"-deployment-admin-email",
		"not-an-email",
	}, &stderr)
	if !invalid.stop || invalid.exitCode != 2 {
		t.Fatalf("expected deprecated invalid email flag to fail parse, got stop=%v exit=%d stderr=%s", invalid.stop, invalid.exitCode, stderr.String())
	}
	if !strings.Contains(stderr.String(), "deployment-admin-email must be an email address") {
		t.Fatalf("invalid deprecated email flag was not reported clearly: %s", stderr.String())
	}
}
