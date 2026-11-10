package recovery

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/JochiRaider/cartulary/internal/modules/extensions/validation"
	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
	"github.com/JochiRaider/cartulary/internal/platform/extensioncodec"
	"github.com/JochiRaider/cartulary/internal/platform/sqlshape"
)

const bindingRestoreResultSchemaID = "cartulary.extension_binding_restore_validation_result.v1"

type bindingRestoreFinding struct {
	Path       string         `json:"path"`
	ReasonCode string         `json:"reason_code"`
	Message    string         `json:"message"`
	Details    map[string]any `json:"details"`
}

type bindingRestoreResult struct {
	SchemaID  string                  `json:"schema_id"`
	BindingID string                  `json:"binding_id"`
	Status    string                  `json:"status"`
	Findings  []bindingRestoreFinding `json:"findings"`
}

// Core's existing restore invariant condition is the sole structural finding.
// Its fixed formatter never incorporates storage errors, paths or row values.
var bindingRestoreConditions = map[string]struct{ path, message string }{
	"invariant_check_failed": {"$.binding_id", "restore invariant check failed"},
}

func invalidBindingRestoreResult(bindingID string) bindingRestoreResult {
	condition := bindingRestoreConditions["invariant_check_failed"]
	return bindingRestoreResult{SchemaID: bindingRestoreResultSchemaID, BindingID: bindingID, Status: "invalid", Findings: []bindingRestoreFinding{{
		Path: condition.path, ReasonCode: "invariant_check_failed", Message: condition.message, Details: map[string]any{},
	}}}
}

// validatePostgresBinding receives only logical entries from one admitted
// binding. It has neither a transaction/query port nor a profile callback.
func validatePostgresBinding(ctx context.Context, binding ExtensionBackupBinding, expected PostgresBindingProof, entries extensioncodec.Entries, scratch *os.Root) bindingRestoreResult {
	if admitPostgresBindingProof(expected, binding) != nil || binding.CurrentCodec.Postgres == nil {
		return invalidBindingRestoreResult(binding.BindingID)
	}
	actual, err := extensioncodec.ComputeProof(ctx, binding.BindingID, binding.StorageKind, binding.CurrentCodec.Postgres.Binding().Limits, scratch, entries)
	if err != nil || actual.ItemCount != expected.ItemCount || actual.ContentBytes != expected.ContentBytes || actual.BindingDigest != expected.BindingDigest {
		return invalidBindingRestoreResult(binding.BindingID)
	}
	return bindingRestoreResult{SchemaID: bindingRestoreResultSchemaID, BindingID: binding.BindingID, Status: "valid", Findings: []bindingRestoreFinding{}}
}

func classifyBindingRestoreResult(bindingID string, result bindingRestoreResult, invocationErr error) error {
	disposition := validation.Classify(invocationErr, result.Findings != nil, len(result.Findings), 256, func() bool {
		if result.SchemaID != bindingRestoreResultSchemaID || result.BindingID != bindingID ||
			(result.Status != "valid" && result.Status != "invalid") || (result.Status == "valid") != (len(result.Findings) == 0) {
			return false
		}
		// The closed Core condition admits one unique finding. Checking this
		// after Classify's count preserves overflow precedence over duplicates.
		if len(result.Findings) > 1 {
			return false
		}
		for _, finding := range result.Findings {
			condition, ok := bindingRestoreConditions[finding.ReasonCode]
			if !ok || finding.Path != condition.path || finding.Message != condition.message || finding.Details == nil || len(finding.Details) != 0 {
				return false
			}
		}
		body, err := canonicaljson.MarshalExtension(result)
		return err == nil && len(body) <= 67108864
	})
	switch disposition {
	case validation.Success:
		return nil
	case validation.InvocationFailed:
		if errors.Is(invocationErr, context.Canceled) {
			return context.Canceled
		}
		if errors.Is(invocationErr, context.DeadlineExceeded) {
			return context.DeadlineExceeded
		}
		return ErrExtensionBindingInvalid
	case validation.Findings:
		return ErrExtensionBindingInvalid
	case validation.Overflow:
		return fmt.Errorf("%w: extension_diagnostic_overflow limit=4096 actual=4097", ErrExtensionBindingInvalid)
	default:
		return fmt.Errorf("%w: extension_validation_result_invalid", ErrExtensionBindingInvalid)
	}
}

func (mutation *vNextRestoreMutation) ValidatePostgresBinding(ctx context.Context, binding ExtensionBackupBinding, expected PostgresBindingProof) error {
	entries := logicalBindingEntries(ctx, binding.CurrentCodec.Postgres, func(ctx context.Context, query sqlshape.Fragment, args ...any) (VNextRows, error) {
		return mutation.tx.Query(ctx, sqlshape.Text(query), args...)
	})
	result := validatePostgresBinding(ctx, binding, expected, entries, mutation.proofScratch)
	return classifyBindingRestoreResult(binding.BindingID, result, ctx.Err())
}
