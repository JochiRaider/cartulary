package deleterestore

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/revisions/deleterestorecontract"
)

const ActiveIncomingPartyReferenceReason = "active_incoming_party_reference"

// Source implements the Party-owned delete/restore semantics behind the root contribution.
type Source struct{}

var _ deleterestorecontract.DeleteRestoreSource = Source{}

func NewSource() Source {
	return Source{}
}

func (Source) SnapshotTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) (map[string]any, error) {
	return deleterestorecontract.ScanSnapshot(tx.QueryRow(ctx, `
SELECT jsonb_build_object('record', to_jsonb(r), 'source', to_jsonb(p))
  FROM records r
  JOIN parties p
    ON p.record_id = r.record_id
 WHERE r.record_id = $1
`, recordID))
}

func (Source) UpdateSourceDeleteStateTx(context.Context, pgx.Tx, uuid.UUID, uuid.UUID, time.Time, bool) error {
	return nil
}

func (Source) ViewSchemaID(context.Context, pgx.Tx, uuid.UUID) (string, error) {
	return "cartulary.view.parties.v1", nil
}

func (Source) PrepareStateTransitionTx(ctx context.Context, tx pgx.Tx, request deleterestorecontract.StateTransitionRequest) (deleterestorecontract.StateTransitionPreparation, error) {
	if request.Kind == deleterestorecontract.StateTransitionRestore {
		return preparePartyRestoreClaimsTx(ctx, tx, request)
	}
	if request.Kind != deleterestorecontract.StateTransitionDelete {
		return deleterestorecontract.StateTransitionPreparation{}, nil
	}
	var exists bool
	if err := tx.QueryRow(ctx, `
SELECT EXISTS (
    SELECT 1
      FROM evidence e
      JOIN records r
        ON r.incident_id = e.incident_id
       AND r.record_id = e.record_id
       AND r.deleted_at IS NULL
     WHERE e.incident_id = $1
       AND (e.collector_party_id = $2 OR e.source_party_id = $2)
    UNION ALL
    SELECT 1
      FROM task_requests t
      JOIN records r
        ON r.incident_id = t.incident_id
       AND r.record_id = t.record_id
       AND r.deleted_at IS NULL
     WHERE t.incident_id = $1
       AND t.requester_party_id = $2
    UNION ALL
    SELECT 1
      FROM active_record_links_v1 rl
      JOIN records src
        ON src.incident_id = rl.incident_id
       AND src.record_id = rl.src_record_id
       AND src.deleted_at IS NULL
     WHERE rl.incident_id = $1
       AND rl.dst_record_id = $2
       AND rl.link_type = 'references_record'
       AND rl.field_key IN ('comm_log.audience_party_ids', 'comm_log.attendee_party_ids')
)
`, request.IncidentID, request.RecordID).Scan(&exists); err != nil {
		return deleterestorecontract.StateTransitionPreparation{}, fmt.Errorf("validate party delete references: %w", err)
	}
	if exists {
		return deleterestorecontract.StateTransitionPreparation{Blocked: &deleterestorecontract.StateTransitionBlock{
			ReasonCode: ActiveIncomingPartyReferenceReason,
		}}, nil
	}
	return deleterestorecontract.StateTransitionPreparation{}, nil
}

type activeKeyTuple struct {
	keyKind         string
	normalizedValue string
	fieldKey        string
}

func preparePartyRestoreClaimsTx(ctx context.Context, tx pgx.Tx, request deleterestorecontract.StateTransitionRequest) (deleterestorecontract.StateTransitionPreparation, error) {
	rows, err := tx.Query(ctx, `
SELECT candidate.key_kind,
       public.parties_normalize_active_key_v1(candidate.key_kind, candidate.raw_value)
  FROM parties p
  CROSS JOIN LATERAL (VALUES
      ('primary_email'::text, p.primary_email),
      ('external_ref'::text, p.external_ref)
  ) AS candidate(key_kind, raw_value)
 WHERE p.incident_id = $1
   AND p.record_id = $2
   AND candidate.raw_value IS NOT NULL
 ORDER BY candidate.key_kind COLLATE "C",
          public.parties_normalize_active_key_v1(candidate.key_kind, candidate.raw_value) COLLATE "C"
`, request.IncidentID, request.RecordID)
	if err != nil {
		return deleterestorecontract.StateTransitionPreparation{}, fmt.Errorf("load Party restore claims: %w", err)
	}
	defer rows.Close()
	var tuples []activeKeyTuple
	for rows.Next() {
		var tuple activeKeyTuple
		if err := rows.Scan(&tuple.keyKind, &tuple.normalizedValue); err != nil {
			return deleterestorecontract.StateTransitionPreparation{}, fmt.Errorf("scan Party restore claim: %w", err)
		}
		tuple.fieldKey = "party." + tuple.keyKind
		tuples = append(tuples, tuple)
	}
	if err := rows.Err(); err != nil {
		return deleterestorecontract.StateTransitionPreparation{}, fmt.Errorf("iterate Party restore claims: %w", err)
	}
	slices.SortFunc(tuples, func(left, right activeKeyTuple) int {
		if compared := strings.Compare(left.keyKind, right.keyKind); compared != 0 {
			return compared
		}
		return strings.Compare(left.normalizedValue, right.normalizedValue)
	})
	var conflictingFields []string
	for _, tuple := range tuples {
		lockKey := request.IncidentID.String() + "\x1f" + tuple.keyKind + "\x1f" + tuple.normalizedValue
		if _, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock(hashtextextended($1, 0))`, lockKey); err != nil {
			return deleterestorecontract.StateTransitionPreparation{}, fmt.Errorf("lock Party restore claim: %w", err)
		}
		var ownerID uuid.UUID
		err := tx.QueryRow(ctx, `
SELECT party_record_id
  FROM party_active_key_claims
 WHERE incident_id = $1
   AND key_kind = $2
   AND normalized_value = $3
`, request.IncidentID, tuple.keyKind, tuple.normalizedValue).Scan(&ownerID)
		if errors.Is(err, pgx.ErrNoRows) {
			continue
		}
		if err != nil {
			return deleterestorecontract.StateTransitionPreparation{}, fmt.Errorf("validate Party restore claim: %w", err)
		}
		if ownerID != request.RecordID {
			conflictingFields = append(conflictingFields, tuple.fieldKey)
		}
	}
	if len(conflictingFields) == 0 {
		return deleterestorecontract.StateTransitionPreparation{}, nil
	}
	slices.Sort(conflictingFields)
	conflictingFields = slices.Compact(conflictingFields)
	return deleterestorecontract.StateTransitionPreparation{Blocked: &deleterestorecontract.StateTransitionBlock{
		ReasonCode:           "exact_match_key_claimed",
		ConflictingFieldKeys: conflictingFields,
	}}, nil
}
