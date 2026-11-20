package rollback

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	evidenceblob "github.com/JochiRaider/cartulary/internal/modules/evidence/internal/blob"
	"github.com/JochiRaider/cartulary/internal/modules/evidence/internal/policy"
	"github.com/JochiRaider/cartulary/internal/modules/evidence/internal/sourcevalue"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	"github.com/JochiRaider/cartulary/internal/modules/revisions/rollbackcontract"
)

type provider struct{}

var _ rollbackcontract.RowSourceProvider = provider{}

func NewProvider() rollbackcontract.RowSourceProvider { return provider{} }

func (provider) ValidateRollbackValue(value map[string]any) error {
	if _, ok := sourceForRollbackValue(value); !ok {
		return rollbackcontract.ErrTargetNotReversible
	}
	return nil
}

func (provider) RestoreTx(ctx context.Context, tx pgx.Tx, request rollbackcontract.RestoreRequest) error {
	source, ok := sourceForRollbackValue(request.RetainedValue)
	if !ok {
		return rollbackcontract.ErrTargetNotReversible
	}
	values := make([]any, 0, 13)
	for _, key := range []string{
		"title", "lifecycle_state", "requested_at", "received_at", "storage_ref", "blob_hash",
		"collector_party_text", "collector_party_id", "source_party_text", "source_party_id",
		"upload_state", "object_blob_id", "owner_user_id",
	} {
		value := source[key]
		if key == "collector_party_id" || key == "source_party_id" || key == "object_blob_id" || key == "owner_user_id" {
			parsed, _, err := nullableUUID(source, key)
			if err != nil {
				return rollbackcontract.ErrTargetNotReversible
			}
			value = parsed
		}
		values = append(values, value)
	}
	// REQ-02-196a: a restored lifecycle must satisfy the bridge against the
	// linked Blob's live state, not the retained scalar.
	lifecycle, _ := source["lifecycle_state"].(string)
	blobID, _, err := nullableUUID(source, "object_blob_id")
	if err != nil {
		return rollbackcontract.ErrTargetNotReversible
	}
	var incidentID uuid.UUID
	if err := tx.QueryRow(ctx, `SELECT incident_id FROM evidence WHERE record_id = $1`, request.RecordID).Scan(&incidentID); err != nil {
		return err
	}
	var linkedBlob *uuid.UUID
	if id, ok := blobID.(uuid.UUID); ok {
		linkedBlob = &id
	}
	// REQ-02-190 and REQ-02-196a: a re-linked Blob must be same-incident,
	// finalized, unassociated and unclaimed, and the restored lifecycle must
	// satisfy the bridge against its live state, decided before any write.
	admission, err := evidenceblob.AdmitLinkTx(ctx, tx, evidenceblob.LinkRequest{
		EvidenceID: request.RecordID, IncidentID: incidentID, Lifecycle: lifecycle, BlobID: linkedBlob, Mode: evidenceblob.ShareNoWait,
	})
	if err != nil {
		return err
	}
	switch admission.Outcome {
	case evidenceblob.Admitted:
	case evidenceblob.Contended:
		return &revisions.RecordLockedError{RecordID: request.RecordID}
	default:
		return rollbackcontract.ErrTargetNotReversible
	}
	_, err = tx.Exec(ctx, `
UPDATE evidence
   SET title = $2::text,
       lifecycle_state = $3::text,
       requested_at = $4::timestamptz,
       received_at = $5::timestamptz,
       storage_ref = $6::text,
       blob_hash = $7::text,
       collector_party_text = $8::text,
       collector_party_id = $9::uuid,
       source_party_text = $10::text,
       source_party_id = $11::uuid,
       upload_state = $12::text,
       object_blob_id = $13::uuid,
       owner_user_id = $14::uuid,
       updated_at = $15
 WHERE record_id = $1
`, append([]any{request.RecordID}, append(values, request.Now.UTC())...)...)
	if associationGuardRejected(err) {
		return rollbackcontract.ErrTargetNotReversible
	}
	return err
}

// associationGuardRejected recognizes the database guards that remain the
// final concurrent-write defence behind the admission above: the at-most-one
// association index, the same-incident Blob reference, and the durable
// cleanup-claim association trigger.
func associationGuardRejected(err error) bool {
	var postgresError *pgconn.PgError
	if !errors.As(err, &postgresError) {
		return false
	}
	switch postgresError.Code {
	case "23505":
		return postgresError.ConstraintName == "evidence_object_blob_unique_idx"
	case "23503":
		return postgresError.ConstraintName == "evidence_object_blob_fkey"
	case "55000":
		return postgresError.Message == "object blob has an active durable cleanup claim"
	default:
		return false
	}
}

func sourceForRollbackValue(value map[string]any) (map[string]any, bool) {
	source, sourceOK := objectMap(value, "source")
	record, recordOK := objectMap(value, "record")
	if !sourceOK || !recordOK || len(value) != 3 || value["snapshot_schema_id"] != "cartulary.revisions.snapshot.evidence.v2" ||
		sourcevalue.ValidateSnapshotSource(record, source) != nil {
		return nil, false
	}
	// Accessible states need a Blob link whatever the Blob's state; the rest of
	// the bridge is decided against the live Blob in RestoreTx (REQ-02-196a).
	if lifecycle, _ := source["lifecycle_state"].(string); (lifecycle == policy.EvidenceAvailable || lifecycle == policy.EvidenceReleased) && source["object_blob_id"] == nil {
		return nil, false
	}
	return source, true
}

func objectMap(value map[string]any, key string) (map[string]any, bool) {
	raw, ok := value[key]
	if !ok || raw == nil {
		return nil, false
	}
	typed, ok := raw.(map[string]any)
	return typed, ok
}

func nullableUUID(value map[string]any, key string) (any, bool, error) {
	raw, present := value[key]
	if !present || raw == nil {
		return nil, present, nil
	}
	text, ok := raw.(string)
	if !ok || strings.TrimSpace(text) == "" {
		return nil, true, rollbackcontract.ErrTargetNotReversible
	}
	parsed, err := uuid.Parse(text)
	if err != nil {
		return nil, true, err
	}
	if parsed == uuid.Nil {
		return nil, true, rollbackcontract.ErrTargetNotReversible
	}
	return parsed, true, nil
}
