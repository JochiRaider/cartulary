package jobs

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"time"
)

type AdmissionReader interface {
	Query(context.Context, string, ...any) (pgx.Rows, error)
	QueryRow(context.Context, string, ...any) pgx.Row
}

// RetainedExtensionJob is a private admission projection, never a public job
// response. Expired tombstones remain available without restoring erased data.
type RetainedExtensionJob struct {
	Resource            Resource
	JobKind             string
	ProgressUnitID      string
	HandlerName         string
	OwnerProfileID      string
	Payload             json.RawMessage
	IdempotencyIdentity json.RawMessage
	RouteKey            string
	ScopeKey            string
	RequestSHA256       string
	ExpiredAt           *time.Time
}

func ReadRetainedExtensionJob(ctx context.Context, reader AdmissionReader, id uuid.UUID) (RetainedExtensionJob, error) {
	stored, err := scanStoredJob(reader.QueryRow(ctx, `SELECT job_id, scope_kind, incident_id, status, cancelable, submitted_by_user_id, auth_policy, submitted_at, updated_at, progress_completed, progress_total, started_at, finished_at, retained_until, result_summary_json, error_summary_json, message, job_kind, progress_unit_id FROM jobs WHERE job_id=$1`, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return RetainedExtensionJob{}, ErrNotFound
	}
	if err != nil {
		return RetainedExtensionJob{}, err
	}
	result := RetainedExtensionJob{Resource: stored.resource, JobKind: stored.jobKind, ProgressUnitID: stored.progressUnitID}
	var resultJSON, errorJSON []byte
	err = reader.QueryRow(ctx, `SELECT handler_name, COALESCE(extension_owner_profile_id,''), handler_payload_json, extension_idempotency_identity, COALESCE(extension_idempotency_route_key,''), COALESCE(extension_idempotency_scope_key,''), COALESCE(extension_normalized_request_sha256,''), expired_at, result_summary_json, error_summary_json FROM jobs WHERE job_id=$1`, id).Scan(&result.HandlerName, &result.OwnerProfileID, &result.Payload, &result.IdempotencyIdentity, &result.RouteKey, &result.ScopeKey, &result.RequestSHA256, &result.ExpiredAt, &resultJSON, &errorJSON)
	if err != nil {
		return RetainedExtensionJob{}, err
	}
	if err := validateRetainedAdmissionResource(result, resultJSON, errorJSON); err != nil {
		return RetainedExtensionJob{}, err
	}
	return result, nil
}

// The OR predicates include contradictions in retained registration identities.
func ReadRetainedExtensionJobPage(ctx context.Context, reader AdmissionReader, profileID, jobKind, handler string, after uuid.UUID) ([]uuid.UUID, error) {
	rows, err := reader.Query(ctx, `SELECT job_id FROM jobs WHERE (extension_owner_profile_id=$1 OR job_kind=$2 OR handler_name=$3) AND job_id>$4 ORDER BY job_id LIMIT 128`, profileID, jobKind, handler, after)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	ids := make([]uuid.UUID, 0, 128)
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

// Validate the retained shell without confusing a private expired tombstone
// with a malformed public terminal resource. Owner-specific success proof
// semantics remain with the initiating owner.
func validateRetainedAdmissionResource(retained RetainedExtensionJob, resultJSON, errorJSON []byte) error {
	resource := retained.Resource
	invalid := func() error { return fmt.Errorf("%w: retained job resource is invalid", ErrInvalidJobDefinition) }
	if validateScope(resource.Scope) != nil || validateProgress(resource.Progress) != nil || resource.SubmittedAt.IsZero() || resource.UpdatedAt.Before(resource.SubmittedAt) {
		return invalid()
	}
	if resource.StartedAt != nil && (resource.StartedAt.Before(resource.SubmittedAt) || resource.UpdatedAt.Before(*resource.StartedAt)) {
		return invalid()
	}
	terminal := resource.Status == StatusSucceeded || resource.Status == StatusFailed || resource.Status == StatusCanceled
	if !terminal {
		if resource.Status != StatusQueued && resource.Status != StatusRunning && resource.Status != StatusCancelRequested {
			return invalid()
		}
		if resource.FinishedAt != nil || resource.RetainedUntil != nil || retained.ExpiredAt != nil || len(resultJSON) != 0 || len(errorJSON) != 0 {
			return invalid()
		}
		if resource.Status == StatusQueued && resource.StartedAt != nil || resource.Status == StatusRunning && resource.StartedAt == nil || resource.Status == StatusCancelRequested && resource.Cancelable {
			return invalid()
		}
		return nil
	}
	if resource.Cancelable || resource.FinishedAt == nil || resource.RetainedUntil == nil || resource.FinishedAt.Before(resource.SubmittedAt) || resource.UpdatedAt.Before(*resource.FinishedAt) || resource.RetainedUntil.Before(*resource.FinishedAt) {
		return invalid()
	}
	if retained.ExpiredAt != nil {
		if retained.ExpiredAt.Before(*resource.RetainedUntil) || len(resultJSON) != 0 || len(errorJSON) != 0 {
			return invalid()
		}
		return nil
	}
	if resource.Status == StatusFailed {
		if len(resultJSON) != 0 || !validRetainedSummary(errorJSON, true) {
			return invalid()
		}
		return nil
	}
	if len(errorJSON) != 0 || !validRetainedSummary(resultJSON, false) || resource.ResultSummary == nil {
		return invalid()
	}
	if resource.Status == StatusCanceled && resource.ResultSummary.Code != "job_canceled" {
		return invalid()
	}
	if resource.Status == StatusSucceeded && resource.Progress.Total != nil && resource.Progress.Completed != *resource.Progress.Total {
		return invalid()
	}
	return nil
}

func validRetainedSummary(raw []byte, failure bool) bool {
	var object map[string]any
	if json.Unmarshal(raw, &object) != nil {
		return false
	}
	code, codeOK := object["code"].(string)
	message, messageOK := object["message"].(string)
	if !codeOK || code == "" || !messageOK || message == "" {
		return false
	}
	delete(object, "code")
	delete(object, "message")
	if failure {
		if _, ok := object["retryable"].(bool); !ok {
			return false
		}
		delete(object, "retryable")
		if details, present := object["details"]; present {
			if _, ok := details.(map[string]any); !ok {
				return false
			}
			delete(object, "details")
		}
	} else if refs, present := object["resource_refs"]; present {
		values, ok := refs.([]any)
		if !ok {
			return false
		}
		for _, value := range values {
			ref, ok := value.(map[string]any)
			if !ok || len(ref) != 3 {
				return false
			}
			for _, key := range []string{"kind", "id", "route"} {
				if text, ok := ref[key].(string); !ok || text == "" {
					return false
				}
			}
		}
		delete(object, "resource_refs")
	}
	return len(object) == 0
}
