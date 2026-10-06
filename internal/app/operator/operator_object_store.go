package operator

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"strings"

	"github.com/JochiRaider/cartulary/internal/app/configassembly"
	database_migrations "github.com/JochiRaider/cartulary/internal/modules/database_migrations"
	"github.com/JochiRaider/cartulary/internal/platform/config"
	"github.com/JochiRaider/cartulary/internal/platform/objectstore"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

const operatorObjectStoreInitResultSchemaID = "cartulary.operator.object_store_init_result.v1"

type operatorObjectStoreInitResult struct {
	SchemaID      string `json:"schema_id"`
	Result        string `json:"result"`
	Created       bool   `json:"created"`
	AlreadyExists bool   `json:"already_exists"`
}

type objectStoreExecutor struct {
	setupPostgres           func(context.Context, postgres.Settings) (operatorPostgresPool, error)
	transport               operatorTransport
	loadConfig              func(string) (configassembly.Loaded, error)
	ensureObjectStoreBucket func(context.Context, objectstore.Settings) (objectstore.EnsureBucketResult, error)
}

type objectStoreInitArgs struct {
	sourceConfigPath string
}

func (executor objectStoreExecutor) runCommand(ctx context.Context, args []string) int {
	parsed, stop, exitCode := parseObjectStoreInitArgs(args[2:], executor.transport.stderr)
	if stop {
		return exitCode
	}
	if err := executor.initialize(ctx, parsed); err != nil {
		executor.transport.logger().Error("operator command failed", "error", err)
		return 1
	}
	return 0
}

func (executor objectStoreExecutor) initialize(ctx context.Context, parsed objectStoreInitArgs) error {
	loaded, err := executor.loadConfig(parsed.sourceConfigPath)
	if err != nil {
		return sanitizeObjectStoreInitError(err)
	}
	cfg := loaded.Deployment()
	settings, err := objectstore.ResolveSettings(configassembly.ObjectStoreBinding(cfg), nil)
	if err != nil {
		return sanitizeObjectStoreInitError(err)
	}
	dbSettings, err := postgres.ResolveSettings(configassembly.PostgresBinding(cfg), postgres.PurposeRecovery, nil)
	if err != nil {
		return sanitizeObjectStoreInitError(err)
	}
	if executor.setupPostgres == nil {
		return database_migrations.ErrIncompatibleCryptoState
	}
	pool, err := executor.setupPostgres(ctx, dbSettings)
	if err != nil {
		return sanitizeObjectStoreInitError(err)
	}
	if pool != nil {
		defer pool.Close()
	}
	result, err := executor.ensureObjectStoreBucket(ctx, settings)
	if err != nil {
		return sanitizeObjectStoreInitError(err)
	}
	payload := operatorObjectStoreInitResult{
		SchemaID:      operatorObjectStoreInitResultSchemaID,
		Result:        operatorObjectStoreInitResultCode(result),
		Created:       result.Created,
		AlreadyExists: result.AlreadyExists,
	}
	return executor.transport.encodeJSON(payload)
}

func parseObjectStoreInitArgs(args []string, stderr io.Writer) (objectStoreInitArgs, bool, int) {
	flags := flag.NewFlagSet("operator object-store init", flag.ContinueOnError)
	flags.SetOutput(normalizeOperatorWriter(stderr))
	sourceConfig := flags.String("config", "", "optional deployment config path")
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return objectStoreInitArgs{}, true, 0
		}
		return objectStoreInitArgs{}, true, 2
	}
	return objectStoreInitArgs{
		sourceConfigPath: strings.TrimSpace(*sourceConfig),
	}, false, 0
}

func operatorObjectStoreInitResultCode(result objectstore.EnsureBucketResult) string {
	if result.Created {
		return "created"
	}
	return "already_exists"
}

func sanitizeObjectStoreInitError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, database_migrations.ErrIncompatibleCryptoState) {
		return database_migrations.ErrIncompatibleCryptoState
	}
	reasonCode := "dependency_unavailable"
	var diagnosticsErr *config.DiagnosticsError
	if errors.As(err, &diagnosticsErr) && diagnosticsErr.Code == config.InvalidDeploymentConfigCode {
		reasonCode = config.InvalidDeploymentConfigCode
	} else if adapterErr, ok := objectstore.AsAdapterError(err); ok && knownObjectStoreReason(adapterErr.Reason) {
		reasonCode = string(adapterErr.Reason)
	}
	return fmt.Errorf("object-store init failed: reason_code=%s", reasonCode)
}

func knownObjectStoreReason(reason objectstore.ReasonCode) bool {
	switch reason {
	case objectstore.ReasonEndpointUnreachable,
		objectstore.ReasonBucketMissing,
		objectstore.ReasonCredentialDenied,
		objectstore.ReasonCapabilityMissing,
		objectstore.ReasonCORSRejected,
		objectstore.ReasonDeadlineExceeded,
		objectstore.ReasonRetryExhausted,
		objectstore.ReasonInvalidRequest,
		objectstore.ReasonObjectMissing,
		objectstore.ReasonRangeInvalid,
		objectstore.ReasonIntegrityMismatch,
		objectstore.ReasonCleanupFailed:
		return true
	default:
		return false
	}
}
