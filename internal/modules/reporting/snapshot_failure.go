package reporting

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
)

type snapshotWorkError struct {
	stage          string
	cause          error
	canonicalBytes int
}

func (e *snapshotWorkError) Error() string { return "reporting snapshot work failed" }
func (e *snapshotWorkError) Unwrap() error { return e.cause }

// Only closed owner tokens cross the Jobs boundary. Neither database messages,
// source values nor error strings are diagnostic dimensions.
func snapshotFailureDetails(cause error) map[string]any {
	details := map[string]any{}
	var failure *snapshotWorkError
	if !errors.As(cause, &failure) {
		return details
	}
	switch failure.stage {
	case "request_validation", "source_capture", "model_materialization", "model_serialization", "capture_persistence":
	default:
		return details
	}
	kind := "internal"
	var database *pgconn.PgError
	var model *renderValidationError
	switch {
	case errors.Is(cause, context.Canceled):
		kind = "canceled"
	case errors.Is(cause, context.DeadlineExceeded):
		kind = "deadline"
	case errors.As(cause, &database):
		switch database.Code {
		case "54000":
			kind = "storage_limit"
		case "23514", "23502", "23503", "23505":
			kind = "storage_constraint"
		case "40001", "40P01":
			kind = "transaction_conflict"
		default:
			kind = "storage"
		}
	case errors.As(cause, &model):
		classified := renderFailureDetails(cause)
		if validateRenderFailure(classified) == nil {
			kind = "model_validation"
			details["failure_code"], details["reason_code"] = classified.FailureCode, classified.ReasonCode
		}
	}
	details["snapshot_stage"], details["failure_kind"] = failure.stage, kind
	if failure.canonicalBytes > 0 {
		details["canonical_input_bytes"] = failure.canonicalBytes
	}
	return details
}
