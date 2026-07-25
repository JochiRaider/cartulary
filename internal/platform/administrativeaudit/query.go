package administrativeaudit

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

var ErrInvalidListFilter = errors.New("administrative audit list filter is invalid")

type ListPosition struct {
	OccurredAt   time.Time
	AuditEventID uuid.UUID
}

type ListFilter struct {
	ScopeKind          string
	ScopeID            *uuid.UUID
	AllowedActionCodes []string
	ActorUserID        *uuid.UUID
	ActionCode         *string
	TargetKind         *string
	TargetID           *string
	OccurredAtGTE      *time.Time
	OccurredAtLT       *time.Time
	After              *ListPosition
	Limit              int
}

type Record struct {
	AuditEventID uuid.UUID
	ScopeKind    string
	ScopeID      *uuid.UUID
	OccurredAt   time.Time
	ActorKind    string
	ActorUserID  *uuid.UUID
	Source       string
	ActionCode   string
	TargetKind   string
	TargetID     *string
	Changes      []Change
	ReasonCode   *string
}

func List(ctx context.Context, db postgres.DB, filter ListFilter) ([]Record, error) {
	if err := validateListFilter(filter); err != nil {
		return nil, err
	}

	var scopeID any
	if filter.ScopeID != nil {
		scopeID = *filter.ScopeID
	}
	var actorUserID any
	if filter.ActorUserID != nil {
		actorUserID = *filter.ActorUserID
	}
	var actionCode any
	if filter.ActionCode != nil {
		actionCode = *filter.ActionCode
	}
	var targetKind any
	if filter.TargetKind != nil {
		targetKind = *filter.TargetKind
	}
	var targetID any
	if filter.TargetID != nil {
		targetID = *filter.TargetID
	}
	var occurredAtGTE any
	if filter.OccurredAtGTE != nil {
		occurredAtGTE = filter.OccurredAtGTE.UTC()
	}
	var occurredAtLT any
	if filter.OccurredAtLT != nil {
		occurredAtLT = filter.OccurredAtLT.UTC()
	}
	var afterOccurredAt any
	var afterAuditEventID any
	if filter.After != nil {
		afterOccurredAt = filter.After.OccurredAt.UTC()
		afterAuditEventID = filter.After.AuditEventID
	}

	rows, err := db.Query(ctx, `
SELECT audit_event_id, scope_kind, scope_id, occurred_at, actor_kind, actor_user_id,
       source, action_code, target_kind, target_id, changes, reason_code
  FROM administrative_audit_projections
 WHERE scope_kind = $1
   AND (($2::uuid IS NULL AND scope_id IS NULL) OR scope_id = $2)
   AND ($3::uuid IS NULL OR actor_user_id = $3)
   AND ($4::text IS NULL OR action_code = $4)
   AND ($5::text IS NULL OR target_kind = $5)
   AND ($6::text IS NULL OR target_id = $6)
   AND ($7::timestamptz IS NULL OR occurred_at >= $7)
   AND ($8::timestamptz IS NULL OR occurred_at < $8)
   AND (
       $9::timestamptz IS NULL OR
       (occurred_at, audit_event_id) < ($9::timestamptz, $10::uuid)
   )
   AND (
       COALESCE(cardinality($11::text[]), 0) = 0 OR
       action_code = ANY($11::text[])
   )
 ORDER BY occurred_at DESC, audit_event_id DESC
 LIMIT $12
`,
		filter.ScopeKind,
		scopeID,
		actorUserID,
		actionCode,
		targetKind,
		targetID,
		occurredAtGTE,
		occurredAtLT,
		afterOccurredAt,
		afterAuditEventID,
		filter.AllowedActionCodes,
		filter.Limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	records := make([]Record, 0, filter.Limit)
	for rows.Next() {
		var (
			record      Record
			changesJSON []byte
		)
		if err := rows.Scan(
			&record.AuditEventID,
			&record.ScopeKind,
			&record.ScopeID,
			&record.OccurredAt,
			&record.ActorKind,
			&record.ActorUserID,
			&record.Source,
			&record.ActionCode,
			&record.TargetKind,
			&record.TargetID,
			&changesJSON,
			&record.ReasonCode,
		); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(changesJSON, &record.Changes); err != nil {
			return nil, fmt.Errorf("decode administrative audit changes: %w", err)
		}
		normalizedChanges, err := validateAndNormalizeEvent(Event{
			ScopeKind:   record.ScopeKind,
			ScopeID:     record.ScopeID,
			OccurredAt:  record.OccurredAt,
			ActorKind:   record.ActorKind,
			ActorUserID: record.ActorUserID,
			Source:      record.Source,
			ActionCode:  record.ActionCode,
			TargetKind:  record.TargetKind,
			TargetID:    record.TargetID,
			Changes:     record.Changes,
			ReasonCode:  record.ReasonCode,
		})
		if err != nil {
			return nil, fmt.Errorf("validate stored administrative audit projection %s: %w", record.AuditEventID, err)
		}
		record.OccurredAt = record.OccurredAt.UTC()
		record.Changes = normalizedChanges
		records = append(records, record)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return records, nil
}

func validateListFilter(filter ListFilter) error {
	if filter.Limit < 1 {
		return fmt.Errorf("%w: limit must be positive", ErrInvalidListFilter)
	}
	switch filter.ScopeKind {
	case ScopeDeployment:
		if filter.ScopeID != nil {
			return fmt.Errorf("%w: deployment scope_id must be null", ErrInvalidListFilter)
		}
	case ScopeIncident:
		if filter.ScopeID == nil {
			return fmt.Errorf("%w: incident scope_id is required", ErrInvalidListFilter)
		}
	default:
		return fmt.Errorf("%w: unknown scope_kind %q", ErrInvalidListFilter, filter.ScopeKind)
	}
	if filter.TargetID != nil && filter.TargetKind == nil {
		return fmt.Errorf("%w: target_id requires target_kind", ErrInvalidListFilter)
	}
	if filter.OccurredAtGTE != nil && filter.OccurredAtLT != nil && !filter.OccurredAtGTE.Before(*filter.OccurredAtLT) {
		return fmt.Errorf("%w: occurred_at lower bound must precede upper bound", ErrInvalidListFilter)
	}
	if filter.After != nil && (filter.After.OccurredAt.IsZero() || filter.After.AuditEventID == uuid.Nil) {
		return fmt.Errorf("%w: incomplete keyset position", ErrInvalidListFilter)
	}
	return nil
}
