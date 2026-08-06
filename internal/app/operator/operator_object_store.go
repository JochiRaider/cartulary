package operator

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"strings"

	"github.com/JochiRaider/cartulary/internal/app/configassembly"
	"github.com/JochiRaider/cartulary/internal/platform/config"
	"github.com/JochiRaider/cartulary/internal/platform/objectstore"
)

const OperatorObjectStoreInitResultSchemaID = "cartulary.operator.object_store_init_result.v1"

type OperatorObjectStoreInitResult struct {
	SchemaID      string `json:"schema_id"`
	Result        string `json:"result"`
	Created       bool   `json:"created"`
	AlreadyExists bool   `json:"already_exists"`
}

type objectStoreExecutor struct {
	transport               operatorTransport
	loadConfig              func(string) (configassembly.Loaded, error)
	ensureObjectStoreBucket func(context.Context, objectstore.Settings) (objectstore.EnsureBucketResult, error)
}

func (executor objectStoreExecutor) runCommand(ctx context.Context, args []string) int {
	parsed := parseObjectStoreInitArgs(args[2:], executor.transport.stderr)
	if parsed.stop {
		return parsed.exitCode
	}
	if err := executor.initialize(ctx, parsed); err != nil {
		executor.transport.logger().Error("operator command failed", "error", err)
		return 1
	}
	return 0
}

func (executor objectStoreExecutor) initialize(ctx context.Context, parsed operatorCLIResult) error {
	loaded, err := executor.loadConfig(parsed.sourceConfigPath)
	if err != nil {
		return sanitizeObjectStoreInitError(err)
	}
	cfg := loaded.Deployment()
	settings, err := objectstore.ResolveSettings(configassembly.ObjectStoreBinding(cfg), nil)
	if err != nil {
		return sanitizeObjectStoreInitError(err)
	}
	result, err := executor.ensureObjectStoreBucket(ctx, settings)
	if err != nil {
		return sanitizeObjectStoreInitError(err)
	}
	payload := OperatorObjectStoreInitResult{
		SchemaID:      OperatorObjectStoreInitResultSchemaID,
		Result:        operatorObjectStoreInitResultCode(result),
		Created:       result.Created,
		AlreadyExists: result.AlreadyExists,
	}
	return executor.transport.encodeJSON(payload)
}

func parseObjectStoreInitArgs(args []string, stderr io.Writer) operatorCLIResult {
	flags := flag.NewFlagSet("operator object-store init", flag.ContinueOnError)
	flags.SetOutput(normalizeOperatorWriter(stderr))
	sourceConfig := flags.String("config", "", "optional deployment config path")
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return operatorCLIResult{stop: true, exitCode: 0}
		}
		return operatorCLIResult{stop: true, exitCode: 2}
	}
	return operatorCLIResult{
		command:          "object-store init",
		sourceConfigPath: strings.TrimSpace(*sourceConfig),
	}
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
