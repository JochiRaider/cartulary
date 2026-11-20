package blob

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/evidence/internal/policy"
)

// ErrEvidenceLifecycleViolation is a REQ-02-196a storage invariant failure:
// a persisted Evidence lifecycle_state outside the closed vocabulary. Every
// route maps it to internal_error, never to illegal_transition.
var ErrEvidenceLifecycleViolation = errors.New("evidence: lifecycle outside the closed vocabulary")

// Outcome classifies a link admission for the caller's own error mapping.
type Outcome uint8

const (
	// Admitted: the resulting Evidence/Blob combination satisfies REQ-02-190
	// and the linked Blob, when present, stays locked through commit.
	Admitted Outcome = iota + 1
	// Rejected: the association rule or the lifecycle bridge is violated.
	Rejected
	// Contended: a non-waiting Blob lock found the row held.
	Contended
)

// Rejection names the violated rule of a Rejected admission.
type Rejection string

const (
	RejectMissing         Rejection = "missing"
	RejectForeignIncident Rejection = "foreign_incident"
	RejectNotFinalized    Rejection = "not_finalized"
	RejectAssociated      Rejection = "associated"
	RejectCleanupClaimed  Rejection = "cleanup_claimed"
	RejectBridge          Rejection = "bridge"
)

// LinkRequest describes the prospective live Evidence state. BlobID nil means
// the Evidence has no link. EvidenceID is excluded from the association
// check, so a record may keep or regain its own Blob.
type LinkRequest struct {
	EvidenceID uuid.UUID
	IncidentID uuid.UUID
	Lifecycle  string
	BlobID     *uuid.UUID
	Mode       LockMode
}

// Admission is the decided link. Blob is the verified, locked Blob when one
// was loaded.
type Admission struct {
	Outcome   Outcome
	Rejection Rejection
	Blob      *Locked
}

// AdmitLinkTx decides the REQ-02-190 association rule and the REQ-02-196a
// lifecycle bridge for a prospective Evidence state against the linked Blob's
// live state, held under a row lock in the caller-selected mode. A Blob that
// violates its lifecycle tuple, or an Evidence lifecycle outside the closed
// vocabulary, is a storage invariant failure returned as an error.
func AdmitLinkTx(ctx context.Context, tx pgx.Tx, request LinkRequest) (Admission, error) {
	if !policy.ValidEvidenceLifecycle(request.Lifecycle) {
		return Admission{}, fmt.Errorf("%w: %s", ErrEvidenceLifecycleViolation, request.EvidenceID)
	}
	if request.BlobID == nil {
		if policy.ViolatesEvidenceBlobBridge(request.Lifecycle, false, "") {
			return Admission{Outcome: Rejected, Rejection: RejectBridge}, nil
		}
		return Admission{Outcome: Admitted}, nil
	}
	locked, err := LockTx(ctx, tx, *request.BlobID, request.Mode)
	switch {
	case errors.Is(err, ErrContended):
		return Admission{Outcome: Contended}, nil
	case errors.Is(err, ErrNotFound):
		return Admission{Outcome: Rejected, Rejection: RejectMissing}, nil
	case err != nil:
		return Admission{}, err
	}
	rejected := func(reason Rejection) (Admission, error) {
		return Admission{Outcome: Rejected, Rejection: reason, Blob: &locked}, nil
	}
	if locked.IncidentID != request.IncidentID {
		return rejected(RejectForeignIncident)
	}
	if !locked.Finalized() {
		return rejected(RejectNotFinalized)
	}
	var associated, claimed bool
	if err := tx.QueryRow(ctx, `
SELECT EXISTS (
           SELECT 1
             FROM evidence e
            WHERE e.object_blob_id = $1
              AND e.record_id <> $2
       ),
       EXISTS (
           SELECT 1
             FROM evidence_blob_cleanup_claims c
            WHERE c.object_blob_id = $1
       )
`, *request.BlobID, request.EvidenceID).Scan(&associated, &claimed); err != nil {
		return Admission{}, err
	}
	if associated {
		return rejected(RejectAssociated)
	}
	if claimed {
		return rejected(RejectCleanupClaimed)
	}
	if policy.ViolatesEvidenceBlobBridge(request.Lifecycle, true, locked.UploadState) {
		return rejected(RejectBridge)
	}
	return Admission{Outcome: Admitted, Blob: &locked}, nil
}
