package operator

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/app/configassembly"
	"github.com/JochiRaider/cartulary/internal/platform/objectstore"
	"github.com/JochiRaider/cartulary/internal/testutil/configtest"
)

func TestOperatorObjectStoreInitCommand_U_DeploymentLocalResult(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	var gotConfigPath string
	var ensureCalls int
	objectStorageRoot := t.TempDir()
	runner := operatorRunner{
		stdout: &stdout,
		stderr: &stderr,
		loadConfig: func(path string) (configassembly.Loaded, error) {
			gotConfigPath = path
			return objectStoreTestConfig(t, objectStorageRoot), nil
		},
		ensureObjectStoreBucket: func(context.Context, objectstore.Settings) (objectstore.EnsureBucketResult, error) {
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

func TestOperatorCollaborationRequeueArgs_U_RequireExactIncident(t *testing.T) {
	var stderr bytes.Buffer
	incidentID := uuid.New()
	parsed := parseCollaborationRequeueArgs(
		[]string{"--incident-id", incidentID.String(), "-config", "/etc/cartulary/config.toml"},
		&stderr,
	)
	if parsed.stop || parsed.incidentID != incidentID || parsed.configPath != "/etc/cartulary/config.toml" {
		t.Fatalf("unexpected collaboration requeue arguments: %#v stderr=%s", parsed, stderr.String())
	}

	for _, args := range [][]string{
		nil,
		{"--incident-id", uuid.Nil.String()},
		{"--incident-id", "invalid"},
		{"--incident-id", incidentID.String(), "extra"},
	} {
		stderr.Reset()
		invalid := parseCollaborationRequeueArgs(args, &stderr)
		if !invalid.stop || invalid.exitCode != 2 || !strings.Contains(stderr.String(), "--incident-id") {
			t.Fatalf("invalid collaboration requeue arguments were admitted: args=%v parsed=%#v stderr=%s", args, invalid, stderr.String())
		}
	}
}

func TestOperatorObjectStoreInitCommand_U_RedactsFailure(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	objectStorageRoot := t.TempDir()
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
		loadConfig: func(string) (configassembly.Loaded, error) {
			return objectStoreTestConfig(t, objectStorageRoot), nil
		},
		ensureObjectStoreBucket: func(context.Context, objectstore.Settings) (objectstore.EnsureBucketResult, error) {
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

func objectStoreTestConfig(t testing.TB, rootPath string) configassembly.Loaded {
	t.Helper()
	deployment := configtest.LoadEffectiveFixture(t, []string{"config", "valid.toml"}, nil)
	deployment.Roots.ObjectStorage.BindingKind = "filesystem_root"
	deployment.Roots.ObjectStorage.Path = rootPath
	deployment.Roots.ObjectStorage.ServiceRef = ""
	loaded, err := configassembly.Admit(deployment)
	if err != nil {
		t.Fatalf("admit object-store operator test deployment: %v", err)
	}
	return loaded
}

func TestOperatorCommandRegistry_U_RejectsUnregisteredSixthCommand(t *testing.T) {
	var stderr bytes.Buffer
	runner := operatorRunner{stderr: &stderr}
	registry, err := runner.commandRegistry()
	if err != nil {
		t.Fatalf("build operator registry: %v", err)
	}
	exitCode := registry.run(context.Background(), []string{
		"retired-sixth-command",
		"run",
	})
	if exitCode != 2 {
		t.Fatalf("expected removed command to stop with usage error, got exit=%d stderr=%s", exitCode, stderr.String())
	}
	if strings.Contains(registry.usage(), "retired-sixth-command") {
		t.Fatalf("operator usage advertises an unregistered sixth command")
	}
	if !strings.Contains(stderr.String(), "operator backup inspect latest") || strings.Contains(stderr.String(), "retired-sixth-command") {
		t.Fatalf("removed command usage was not clear: %s", stderr.String())
	}
}
